package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type repairTargetValues struct {
	GameVersion        string
	Loader             string
	LoaderVersion      string
	ForgeVersion       string
	NeoForgeVersion    string
	JavaMajor          int
	LoomVersion        string
	ArchitecturyLoom   string
	ForgeGradleVersion string
	ModDevVersion      string
	NeoGradleVersion   string
	FabricAPIVersion   string
	YarnVersion        string
	Intermediary       string
	QuiltMappings      string
	MCPConfig          string
	NeoForm            string
	ResourcePackFormat int
	DataPackFormat     int
}

func prepareRepairSession(run *PortingBuildRun, request RepairPrepareRequest) error {
	if run == nil {
		return errors.New("nil repair session")
	}
	if run.State == "running" {
		return errRepairSessionBusy
	}
	request.TargetGameVersion = strings.TrimSpace(request.TargetGameVersion)
	request.TargetLoader = normalizeDoctorLoader(request.TargetLoader)
	if request.TargetGameVersion == "" || request.TargetLoader == "" {
		return errors.New("targetGameVersion and targetLoader are required")
	}
	if err := verifyImmutableSource(run); err != nil {
		return err
	}
	atlas, err := loadRepairVersionAtlas()
	if err != nil {
		return err
	}
	resolution := atlas.Resolve(request.TargetGameVersion, request.TargetLoader)
	if !resolution.Exists {
		return fmt.Errorf("target Minecraft version %q is not present in the embedded authoritative atlas", request.TargetGameVersion)
	}
	if err := resetWorkingCopy(run); err != nil {
		return fmt.Errorf("reset working copy: %w", err)
	}
	values := repairValuesFromResolution(resolution)
	projectRoot := filepath.Join(run.Paths.WorkingCopy, filepath.FromSlash(run.Project.ProjectRoot))
	if !pathContainedBy(run.Paths.WorkingCopy, projectRoot) {
		return errors.New("detected project root escapes the working copy")
	}
	changes, warnings, err := applyRecognizedMigrationEdits(projectRoot, values)
	if err != nil {
		return err
	}
	updatedProfile, err := detectRepairProject(run.Paths.WorkingCopy)
	if err != nil {
		return fmt.Errorf("profile patched working copy: %w", err)
	}
	run.Target = &AtlasResolveRequest{GameVersion: request.TargetGameVersion, Loader: request.TargetLoader}
	run.Resolution = &resolution
	run.Changes = changes
	run.Project = updatedProfile
	run.State = "prepared"
	run.Phase = "migration-staged"
	run.LastError = ""
	run.Artifacts = nil
	run.Warnings = uniqueStringsPreserve(append(append(run.Warnings, resolution.Warnings...), warnings...))
	if len(changes) == 0 {
		run.Warnings = uniqueStringsPreserve(append(run.Warnings, "No recognized version field required an automatic edit. The target atlas and build plan were staged without rewriting ambiguous source."))
	}
	return writeRepairReceipt(run)
}

func repairValuesFromResolution(resolution AtlasResolution) repairTargetValues {
	values := repairTargetValues{
		GameVersion: resolution.GameVersion,
		Loader:      resolution.Loader,
		JavaMajor:   resolution.JavaMajor,
	}
	if resolution.Loader == "forge" {
		values.ForgeVersion = resolution.LoaderVersion
	} else if resolution.Loader == "neoforge" {
		values.NeoForgeVersion = resolution.LoaderVersion
	} else {
		values.LoaderVersion = resolution.LoaderVersion
	}
	for _, choice := range resolution.BuildToolchains {
		switch choice.ID {
		case "fabric-loom":
			values.LoomVersion = choice.Version
		case "architectury-loom":
			values.ArchitecturyLoom = choice.Version
		case "forgegradle":
			values.ForgeGradleVersion = choice.Version
		case "moddevgradle":
			values.ModDevVersion = choice.Version
		case "neogradle":
			values.NeoGradleVersion = choice.Version
		}
	}
	for _, choice := range resolution.GameArtifacts {
		switch choice.ID {
		case "fabric-api":
			values.FabricAPIVersion = choice.Version
		case "forge":
			values.ForgeVersion = choice.Version
		case "neoforge":
			values.NeoForgeVersion = choice.Version
		}
	}
	for _, choice := range resolution.Mappings {
		switch choice.ID {
		case "yarn":
			values.YarnVersion = choice.Version
		case "intermediary":
			values.Intermediary = choice.Version
		case "quilt-mappings":
			values.QuiltMappings = choice.Version
		case "mcpconfig":
			values.MCPConfig = choice.Version
		case "neoform":
			values.NeoForm = choice.Version
		}
	}
	if resolution.Version != nil && resolution.Version.MCMeta != nil {
		values.ResourcePackFormat = resolution.Version.MCMeta.ResourcePackVersion
		values.DataPackFormat = resolution.Version.MCMeta.DataPackVersion
	}
	return values
}

func applyRecognizedMigrationEdits(projectRoot string, values repairTargetValues) ([]RepairChange, []string, error) {
	var changes []RepairChange
	var warnings []string
	err := filepath.WalkDir(projectRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == projectRoot {
			return nil
		}
		if entry.IsDir() {
			if repairIgnoredDirectory(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(projectRoot, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		lower := strings.ToLower(entry.Name())
		var fileChanges []RepairChange
		switch {
		case lower == "gradle.properties":
			fileChanges, err = patchRepairProperties(path, rel, values)
		case lower == "build.gradle" || lower == "build.gradle.kts" || lower == "settings.gradle" || lower == "settings.gradle.kts" || lower == "pom.xml":
			fileChanges, err = patchRepairBuildText(path, rel, values)
		case lower == "fabric.mod.json":
			fileChanges, err = patchFabricMetadata(path, rel, values)
		case lower == "quilt.mod.json":
			fileChanges, err = patchQuiltMetadata(path, rel, values)
		case lower == "mods.toml" || lower == "neoforge.mods.toml":
			fileChanges, err = patchModTOML(path, rel, values)
		case lower == "pack.mcmeta":
			fileChanges, err = patchPackMCMeta(path, rel, values)
		}
		if err != nil {
			return fmt.Errorf("patch %s: %w", rel, err)
		}
		changes = append(changes, fileChanges...)
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	if values.Loader == "forge" && values.ForgeVersion == "" {
		warnings = append(warnings, "No exact Forge coordinate was available, so Forge dependency coordinates were not rewritten.")
	}
	if values.Loader == "neoforge" && values.NeoForgeVersion == "" {
		warnings = append(warnings, "No exact NeoForge coordinate was available, so NeoForge dependency coordinates were not rewritten.")
	}
	sort.SliceStable(changes, func(i, j int) bool {
		if changes[i].File != changes[j].File {
			return changes[i].File < changes[j].File
		}
		return changes[i].Field < changes[j].Field
	})
	return changes, warnings, nil
}

func patchRepairProperties(path, rel string, values repairTargetValues) ([]RepairChange, error) {
	data, err := readRepairSmallTextFile(path, 8<<20)
	if err != nil {
		return nil, err
	}
	lineEnding := "\n"
	if strings.Contains(string(data), "\r\n") {
		lineEnding = "\r\n"
	}
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	lines := strings.Split(text, "\n")
	var changes []RepairChange
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "!") {
			continue
		}
		idx := strings.IndexAny(line, "=:")
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])
		target, reason := repairPropertyTarget(strings.ToLower(key), values)
		if target == "" || target == value {
			continue
		}
		prefix := line[:idx+1]
		spacing := ""
		for _, r := range line[idx+1:] {
			if r == ' ' || r == '\t' {
				spacing += string(r)
				continue
			}
			break
		}
		lines[i] = prefix + spacing + target
		changes = append(changes, RepairChange{File: rel, Field: key, Before: limitRepairValue(value), After: limitRepairValue(target), Applied: true, Confidence: "high", Reason: reason})
	}
	if len(changes) == 0 {
		return nil, nil
	}
	return changes, writeFileAtomic(path, []byte(strings.Join(lines, lineEnding)), 0o644)
}

func repairPropertyTarget(key string, values repairTargetValues) (string, string) {
	switch key {
	case "minecraft_version", "minecraftversion", "minecraft.version", "mc_version", "mcversion":
		return values.GameVersion, "Exact target game version from the embedded Mojang/loader atlas."
	case "java_version", "javaversion", "java.version", "jvm_version", "jvmversion":
		if values.JavaMajor > 0 {
			return strconv.Itoa(values.JavaMajor), "Target Java major from the exact Mojang version manifest."
		}
	case "loader_version", "fabric_loader_version", "fabricloader_version":
		if values.Loader == "fabric" && values.LoaderVersion != "" {
			return values.LoaderVersion, "Latest stable Fabric Loader captured by Fabric Meta."
		}
	case "quilt_loader_version", "quiltloader_version":
		if values.Loader == "quilt" && values.LoaderVersion != "" {
			return values.LoaderVersion, "Latest stable Quilt Loader captured by Quilt Meta."
		}
	case "forge_version", "forgeversion":
		if values.ForgeVersion != "" {
			return forgeVersionSuffix(values.GameVersion, values.ForgeVersion), "Exact official Forge coordinate for the target Minecraft version."
		}
	case "neo_version", "neoforge_version", "neoforgeversion":
		if values.NeoForgeVersion != "" {
			return values.NeoForgeVersion, "Exact official NeoForge coordinate for the target release line."
		}
	case "loom_version", "loomversion":
		if values.LoomVersion != "" {
			return values.LoomVersion, "Current stable Fabric Loom release from official Maven metadata."
		}
	case "architectury_loom_version", "architecturyloom_version":
		if values.ArchitecturyLoom != "" {
			return values.ArchitecturyLoom, "Current stable Architectury Loom release from official Maven metadata."
		}
	case "forgegradle_version", "forge_gradle_version":
		if values.ForgeGradleVersion != "" {
			return values.ForgeGradleVersion, "ForgeGradle major selected for the target Minecraft era."
		}
	case "moddevgradle_version", "moddev_version":
		if values.ModDevVersion != "" {
			return values.ModDevVersion, "Current first-party ModDevGradle release."
		}
	case "neogradle_version":
		if values.NeoGradleVersion != "" {
			return values.NeoGradleVersion, "Current NeoGradle userdev release for existing NeoGradle workspaces."
		}
	case "fabric_version", "fabric_api_version", "fabricapi_version":
		if values.FabricAPIVersion != "" {
			return values.FabricAPIVersion, "Highest Fabric API build published for the exact target game version."
		}
	case "yarn_mappings", "yarn_version", "yarn_mappings_version":
		if values.YarnVersion != "" {
			return values.YarnVersion, "Highest official Yarn mapping build for the target game version."
		}
	case "intermediary_version":
		if values.Intermediary != "" {
			return values.Intermediary, "Exact official intermediary mapping coordinate."
		}
	case "quilt_mappings", "quilt_mappings_version":
		if values.QuiltMappings != "" {
			return values.QuiltMappings, "Highest official Quilt mappings build for the target game version."
		}
	case "mcp_config", "mcpconfig_version", "mappings_version":
		if values.Loader == "forge" && values.MCPConfig != "" {
			return values.MCPConfig, "Highest MCPConfig coordinate matching the target game version."
		}
	case "neoform_version":
		if values.NeoForm != "" {
			return values.NeoForm, "Highest NeoForm coordinate matching the target game line."
		}
	}
	return "", ""
}

func forgeVersionSuffix(gameVersion, coordinate string) string {
	prefix := gameVersion + "-"
	if strings.HasPrefix(coordinate, prefix) {
		return strings.TrimPrefix(coordinate, prefix)
	}
	return coordinate
}

func patchRepairBuildText(path, rel string, values repairTargetValues) ([]RepairChange, error) {
	data, err := readRepairSmallTextFile(path, 12<<20)
	if err != nil {
		return nil, err
	}
	original := string(data)
	updated := original
	var changes []RepairChange
	apply := func(field string, pattern *regexp.Regexp, replacement func(string) string, reason string) {
		matches := pattern.FindAllString(updated, -1)
		if len(matches) == 0 {
			return
		}
		before := matches[0]
		updated = pattern.ReplaceAllStringFunc(updated, replacement)
		afterMatch := pattern.FindString(updated)
		if before == afterMatch && updated == original {
			return
		}
		changes = append(changes, RepairChange{File: rel, Field: field, Before: limitRepairValue(before), After: limitRepairValue(replacement(before)), Applied: true, Confidence: "high", Reason: reason})
	}
	if values.ForgeVersion != "" {
		pattern := regexp.MustCompile(`(?i)net\.minecraftforge:forge:[A-Za-z0-9_.+\-]+`)
		apply("Forge coordinate", pattern, func(string) string { return "net.minecraftforge:forge:" + values.ForgeVersion }, "Exact official Forge coordinate matching the target Minecraft version.")
	}
	if values.NeoForgeVersion != "" {
		pattern := regexp.MustCompile(`(?i)net\.neoforged:neoforge:[A-Za-z0-9_.+\-]+`)
		apply("NeoForge coordinate", pattern, func(string) string { return "net.neoforged:neoforge:" + values.NeoForgeVersion }, "Exact official NeoForge coordinate matching the target release line.")
	}
	if values.Loader == "fabric" && values.LoaderVersion != "" {
		pattern := regexp.MustCompile(`(?i)net\.fabricmc:fabric-loader:[A-Za-z0-9_.+\-]+`)
		apply("Fabric Loader coordinate", pattern, func(string) string { return "net.fabricmc:fabric-loader:" + values.LoaderVersion }, "Latest stable Fabric Loader from Fabric Meta.")
	}
	if values.Loader == "quilt" && values.LoaderVersion != "" {
		pattern := regexp.MustCompile(`(?i)org\.quiltmc:quilt-loader:[A-Za-z0-9_.+\-]+`)
		apply("Quilt Loader coordinate", pattern, func(string) string { return "org.quiltmc:quilt-loader:" + values.LoaderVersion }, "Latest stable Quilt Loader from Quilt Meta.")
	}
	if values.FabricAPIVersion != "" {
		pattern := regexp.MustCompile(`(?i)net\.fabricmc\.fabric-api:fabric-api:[A-Za-z0-9_.+\-]+`)
		apply("Fabric API coordinate", pattern, func(string) string { return "net.fabricmc.fabric-api:fabric-api:" + values.FabricAPIVersion }, "Highest Fabric API coordinate published for the target game version.")
	}
	if values.YarnVersion != "" {
		pattern := regexp.MustCompile(`(?i)net\.fabricmc:yarn:[A-Za-z0-9_.+\-]+(?::v2)?`)
		apply("Yarn mapping coordinate", pattern, func(match string) string {
			suffix := ""
			if strings.HasSuffix(strings.ToLower(match), ":v2") {
				suffix = ":v2"
			}
			return "net.fabricmc:yarn:" + values.YarnVersion + suffix
		}, "Highest Yarn build matching the target game version.")
	}
	if values.JavaMajor > 0 {
		java := strconv.Itoa(values.JavaMajor)
		pattern := regexp.MustCompile(`(?i)JavaLanguageVersion\s*\.\s*of\s*\(\s*\d{1,2}\s*\)`)
		apply("Java toolchain", pattern, func(string) string { return "JavaLanguageVersion.of(" + java + ")" }, "Java major required by the exact target Mojang manifest.")
		pattern = regexp.MustCompile(`(?i)JavaVersion\.VERSION_\d{1,2}`)
		apply("Java compatibility", pattern, func(string) string { return "JavaVersion.VERSION_" + java }, "Java major required by the exact target Mojang manifest.")
	}
	if updated == original {
		return nil, nil
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	return changes, writeFileAtomic(path, []byte(updated), info.Mode().Perm())
}

func patchFabricMetadata(path, rel string, values repairTargetValues) ([]RepairChange, error) {
	data, err := readRepairSmallTextFile(path, 4<<20)
	if err != nil {
		return nil, err
	}
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, err
	}
	depends, ok := root["depends"].(map[string]any)
	if !ok {
		return nil, nil
	}
	var changes []RepairChange
	set := func(key, value, reason string) {
		if value == "" {
			return
		}
		before, exists := depends[key]
		if !exists {
			return
		}
		after := value
		if key == "minecraft" {
			after = "=" + value
		} else if key == "java" {
			after = ">=" + value
		}
		if fmt.Sprint(before) == after {
			return
		}
		depends[key] = after
		changes = append(changes, RepairChange{File: rel, Field: "depends." + key, Before: limitRepairValue(fmt.Sprint(before)), After: after, Applied: true, Confidence: "high", Reason: reason})
	}
	set("minecraft", values.GameVersion, "Exact target Minecraft constraint.")
	if values.JavaMajor > 0 {
		set("java", strconv.Itoa(values.JavaMajor), "Minimum Java constraint required by the target game manifest.")
	}
	if values.Loader == "fabric" {
		set("fabricloader", values.LoaderVersion, "Target Fabric Loader constraint from Fabric Meta.")
	}
	if len(changes) == 0 {
		return nil, nil
	}
	root["depends"] = depends
	updated, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return nil, err
	}
	updated = append(updated, '\n')
	return changes, writeFileAtomic(path, updated, 0o644)
}

func patchQuiltMetadata(path, rel string, values repairTargetValues) ([]RepairChange, error) {
	data, err := readRepairSmallTextFile(path, 4<<20)
	if err != nil {
		return nil, err
	}
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, err
	}
	loader, ok := root["quilt_loader"].(map[string]any)
	if !ok {
		return nil, nil
	}
	var changes []RepairChange
	var patchDepends func(value any) any
	patchDepends = func(value any) any {
		switch typed := value.(type) {
		case []any:
			for i, row := range typed {
				typed[i] = patchDepends(row)
			}
			return typed
		case map[string]any:
			id := strings.ToLower(strings.TrimSpace(fmt.Sprint(typed["id"])))
			target := ""
			reason := ""
			switch id {
			case "minecraft":
				target = "=" + values.GameVersion
				reason = "Exact target Minecraft constraint."
			case "quilt_loader":
				if values.Loader == "quilt" && values.LoaderVersion != "" {
					target = ">=" + values.LoaderVersion
					reason = "Target Quilt Loader constraint from Quilt Meta."
				}
			}
			if target != "" {
				before := fmt.Sprint(typed["versions"])
				if before != target {
					typed["versions"] = target
					changes = append(changes, RepairChange{File: rel, Field: "quilt_loader.depends." + id, Before: limitRepairValue(before), After: target, Applied: true, Confidence: "high", Reason: reason})
				}
			}
			return typed
		default:
			return value
		}
	}
	if depends, exists := loader["depends"]; exists {
		loader["depends"] = patchDepends(depends)
	}
	if len(changes) == 0 {
		return nil, nil
	}
	root["quilt_loader"] = loader
	updated, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return nil, err
	}
	updated = append(updated, '\n')
	return changes, writeFileAtomic(path, updated, 0o644)
}

func patchModTOML(path, rel string, values repairTargetValues) ([]RepairChange, error) {
	data, err := readRepairSmallTextFile(path, 4<<20)
	if err != nil {
		return nil, err
	}
	lineEnding := "\n"
	if strings.Contains(string(data), "\r\n") {
		lineEnding = "\r\n"
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	currentDependency := ""
	var changes []RepairChange
	modIDPattern := regexp.MustCompile(`(?i)^\s*modId\s*=\s*["']([^"']+)["']`)
	versionPattern := regexp.MustCompile(`(?i)^(\s*versionRange\s*=\s*["'])([^"']*)(["'].*)$`)
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[[") {
			currentDependency = ""
		}
		if match := modIDPattern.FindStringSubmatch(line); len(match) > 1 {
			currentDependency = strings.ToLower(strings.TrimSpace(match[1]))
			continue
		}
		match := versionPattern.FindStringSubmatch(line)
		if len(match) == 0 || currentDependency == "" {
			continue
		}
		target := ""
		reason := ""
		switch currentDependency {
		case "minecraft":
			target = "[" + values.GameVersion + "]"
			reason = "Exact target Minecraft dependency range."
		case "forge":
			if values.Loader == "forge" && values.ForgeVersion != "" {
				target = "[" + forgeVersionSuffix(values.GameVersion, values.ForgeVersion) + "]"
				reason = "Exact target Forge dependency range."
			}
		case "neoforge":
			if values.Loader == "neoforge" && values.NeoForgeVersion != "" {
				target = "[" + values.NeoForgeVersion + "]"
				reason = "Exact target NeoForge dependency range."
			}
		}
		if target == "" || match[2] == target {
			continue
		}
		lines[i] = match[1] + target + match[3]
		changes = append(changes, RepairChange{File: rel, Field: "dependency." + currentDependency + ".versionRange", Before: match[2], After: target, Applied: true, Confidence: "high", Reason: reason})
	}
	if len(changes) == 0 {
		return nil, nil
	}
	return changes, writeFileAtomic(path, []byte(strings.Join(lines, lineEnding)), 0o644)
}

func patchPackMCMeta(path, rel string, values repairTargetValues) ([]RepairChange, error) {
	if values.ResourcePackFormat <= 0 && values.DataPackFormat <= 0 {
		return nil, nil
	}
	data, err := readRepairSmallTextFile(path, 2<<20)
	if err != nil {
		return nil, err
	}
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, err
	}
	pack, ok := root["pack"].(map[string]any)
	if !ok {
		return nil, nil
	}
	format := values.ResourcePackFormat
	parent := filepath.Dir(path)
	hasAssets := pathExists(filepath.Join(parent, "assets"))
	hasData := pathExists(filepath.Join(parent, "data"))
	reason := "Resource-pack format from the target mcmeta version catalog."
	if hasData && !hasAssets && values.DataPackFormat > 0 {
		format = values.DataPackFormat
		reason = "Data-pack format from the target mcmeta version catalog."
	} else if hasData && hasAssets && values.DataPackFormat > format {
		format = values.DataPackFormat
		reason = "Combined mod resource/data pack uses the higher target pack format; runtime validation remains required."
	}
	if format <= 0 {
		return nil, nil
	}
	before := fmt.Sprint(pack["pack_format"])
	after := strconv.Itoa(format)
	if before == after {
		return nil, nil
	}
	pack["pack_format"] = format
	root["pack"] = pack
	updated, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return nil, err
	}
	updated = append(updated, '\n')
	change := RepairChange{File: rel, Field: "pack.pack_format", Before: before, After: after, Applied: true, Confidence: "medium", Reason: reason}
	return []RepairChange{change}, writeFileAtomic(path, updated, 0o644)
}

func limitRepairValue(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= 280 {
		return value
	}
	return value[:277] + "..."
}

func writeRepairReceipt(run *PortingBuildRun) error {
	if run == nil {
		return errors.New("nil repair session")
	}
	receipt := map[string]any{
		"schemaVersion": repairLabSchemaVersion,
		"generatedAt":   time.Now().UTC().Format(time.RFC3339),
		"sessionId":     run.ID,
		"state":         run.State,
		"phase":         run.Phase,
		"source":        run.Source,
		"project":       run.Project,
		"target":        run.Target,
		"resolution":    run.Resolution,
		"changes":       run.Changes,
		"runs":          run.Runs,
		"artifacts":     run.Artifacts,
		"exports":       run.Exports,
		"security":      run.Security,
		"warnings":      run.Warnings,
		"lastError":     run.LastError,
	}
	data, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return err
	}
	if err := writeFileAtomic(run.Paths.ReceiptJSON, append(data, '\n'), 0o600); err != nil {
		return err
	}
	var md strings.Builder
	fmt.Fprintf(&md, "# Minecraft Mod Vault Repair Receipt\n\n")
	fmt.Fprintf(&md, "- Session: `%s`\n- Generated: `%s`\n- State: **%s**\n- Source: `%s`\n- Source SHA-256: `%s`\n- Immutable tree SHA-256: `%s`\n", run.ID, time.Now().UTC().Format(time.RFC3339), run.State, run.Source.Filename, run.Source.SHA256, run.Source.TreeSHA256)
	if run.Target != nil {
		fmt.Fprintf(&md, "- Target: `%s` / `%s`\n", run.Target.GameVersion, run.Target.Loader)
	}
	fmt.Fprintf(&md, "\n## Detected project\n\n- Build: `%s`\n- Wrapper: `%s`\n- Loader: `%s`\n- Source game: `%s`\n- Java: `%d`\n- Fingerprint: `%s`\n", run.Project.BuildSystem, run.Project.Wrapper, run.Project.Loader, run.Project.GameVersion, run.Project.JavaMajor, run.Project.Fingerprint)
	fmt.Fprintf(&md, "\n## Applied recognized changes (%d)\n\n", len(run.Changes))
	if len(run.Changes) == 0 {
		md.WriteString("No automatic source field was changed.\n")
	}
	for _, change := range run.Changes {
		fmt.Fprintf(&md, "- `%s` — **%s**: `%s` -> `%s` (%s)\n", change.File, change.Field, change.Before, change.After, change.Reason)
	}
	fmt.Fprintf(&md, "\n## Build runs (%d)\n\n", len(run.Runs))
	for _, command := range run.Runs {
		fmt.Fprintf(&md, "- `%s`: **%s**, exit `%d`, started `%s`, finished `%s`\n", strings.Join(command.Command, " "), command.State, command.ExitCode, command.StartedAt, command.FinishedAt)
	}
	fmt.Fprintf(&md, "\n## Produced artifacts (%d)\n\n", len(run.Artifacts))
	for _, artifact := range run.Artifacts {
		fmt.Fprintf(&md, "- `%s` — %d bytes — SHA-256 `%s`\n", artifact.RelativePath, artifact.Size, artifact.SHA256)
	}
	if len(run.Warnings) > 0 {
		md.WriteString("\n## Warnings and remaining proof\n\n")
		for _, warning := range run.Warnings {
			fmt.Fprintf(&md, "- %s\n", warning)
		}
	}
	md.WriteString("\n## Safety statement\n\nThe original source archive and extracted source tree are immutable inputs. All automatic edits and builds occur in a separate working copy. Build scripts execute project code and are not claimed to be OS- or VM-sandboxed. A successful compilation is evidence, not proof of in-game compatibility; loader startup and representative gameplay/runtime testing remain required.\n")
	return writeFileAtomic(run.Paths.ReceiptMarkdown, []byte(md.String()), 0o600)
}
