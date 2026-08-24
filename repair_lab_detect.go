package main

import (
	"crypto/sha256"
	"encoding/hex"
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
)

var (
	repairGameCoordinatePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)net\.minecraftforge:forge:([A-Za-z0-9_.+\-]+)`),
		regexp.MustCompile(`(?i)net\.neoforged:neoforge:([A-Za-z0-9_.+\-]+)`),
		regexp.MustCompile(`(?i)com\.mojang:minecraft:([A-Za-z0-9_.+\-]+)`),
	}
	repairJavaPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)JavaLanguageVersion\s*\.\s*of\s*\(\s*(\d{1,2})\s*\)`),
		regexp.MustCompile(`(?i)JavaVersion\.VERSION_(\d{1,2})`),
		regexp.MustCompile(`(?im)^\s*(?:sourceCompatibility|targetCompatibility)\s*=\s*['\"]?(\d{1,2})`),
		regexp.MustCompile(`(?im)^\s*(?:java_version|javaVersion|java\.version)\s*[=:]\s*(\d{1,2})`),
		regexp.MustCompile(`(?is)<maven\.compiler\.(?:source|target|release)>\s*(\d{1,2})\s*</maven\.compiler\.(?:source|target|release)>`),
	}
)

func detectRepairProject(sourceRoot string) (RepairProjectProfile, error) {
	profile := RepairProjectProfile{BuildSystem: "unknown", Loader: "unknown", Confidence: "low"}
	projectRoot, err := findRepairProjectRoot(sourceRoot)
	if err != nil {
		return profile, err
	}
	relProject, err := filepath.Rel(sourceRoot, projectRoot)
	if err != nil {
		return profile, err
	}
	if relProject == "." {
		relProject = ""
	}
	profile.ProjectRoot = filepath.ToSlash(relProject)

	var textCorpus strings.Builder
	propertyValues := map[string]string{}
	maxFiles := 50000
	seenFiles := 0
	err = filepath.WalkDir(projectRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == projectRoot {
			return nil
		}
		rel, err := filepath.Rel(projectRoot, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if entry.IsDir() {
			if repairIgnoredDirectory(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		seenFiles++
		if seenFiles > maxFiles {
			return errors.New("project contains too many files for deterministic profile detection")
		}
		lower := strings.ToLower(entry.Name())
		switch lower {
		case "build.gradle", "build.gradle.kts", "settings.gradle", "settings.gradle.kts", "gradle.properties", "pom.xml", "fabric.mod.json", "quilt.mod.json", "mods.toml", "neoforge.mods.toml", "pack.mcmeta":
			data, err := readRepairSmallTextFile(path, 4<<20)
			if err != nil {
				profile.Warnings = append(profile.Warnings, fmt.Sprintf("Could not read %s: %v", rel, err))
				return nil
			}
			text := string(data)
			textCorpus.WriteString("\n--- " + rel + " ---\n")
			textCorpus.WriteString(text)
			switch lower {
			case "build.gradle", "build.gradle.kts", "settings.gradle", "settings.gradle.kts", "gradle.properties", "pom.xml":
				profile.BuildFiles = append(profile.BuildFiles, rel)
			default:
				profile.MetadataFiles = append(profile.MetadataFiles, rel)
			}
			if lower == "gradle.properties" {
				for key, value := range parseSimpleProperties(text) {
					if _, exists := propertyValues[key]; !exists {
						propertyValues[key] = value
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		return profile, err
	}

	sort.Strings(profile.BuildFiles)
	sort.Strings(profile.MetadataFiles)
	corpus := textCorpus.String()
	profile.BuildSystem, profile.Wrapper = detectRepairBuildSystem(projectRoot, profile.BuildFiles)
	if profile.Wrapper != "" {
		wrapperPath := filepath.Join(projectRoot, filepath.FromSlash(profile.Wrapper))
		profile.WrapperSHA256, _, _ = hashFileSHA256(wrapperPath)
	}
	profile.Loader, profile.Signals = detectRepairLoader(profile.MetadataFiles, corpus)
	profile.GameVersion = detectRepairGameVersion(profile.Loader, propertyValues, corpus)
	profile.JavaMajor = detectRepairJavaMajor(propertyValues, corpus)
	if profile.JavaMajor == 0 && profile.GameVersion != "" {
		profile.JavaMajor = targetJavaForMinecraft(profile.GameVersion)
	}
	profile.Modules = detectRepairModules(projectRoot, profile.BuildFiles)
	profile.AvailableCommands = repairCommandsForProfile(profile)
	profile.Confidence = repairProfileConfidence(profile)
	profile.Warnings = append(profile.Warnings, repairProfileWarnings(profile)...)
	profile.Warnings = uniqueStringsPreserve(profile.Warnings)
	profile.Fingerprint = repairProjectFingerprint(profile, corpus)
	return profile, nil
}

func findRepairProjectRoot(root string) (string, error) {
	info, err := os.Stat(root)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", errors.New("repair source root is not a directory")
	}
	type candidate struct {
		path     string
		priority int
		depth    int
	}
	var candidates []candidate
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path != root && repairIgnoredDirectory(entry.Name()) {
				return filepath.SkipDir
			}
			rel, _ := filepath.Rel(root, path)
			if strings.Count(filepath.ToSlash(rel), "/") > 5 {
				return filepath.SkipDir
			}
			return nil
		}
		name := strings.ToLower(entry.Name())
		priority := 0
		switch name {
		case "settings.gradle", "settings.gradle.kts":
			priority = 4
		case "gradlew", "gradlew.bat", "mvnw", "mvnw.cmd":
			priority = 3
		case "build.gradle", "build.gradle.kts", "pom.xml":
			priority = 2
		case "fabric.mod.json", "quilt.mod.json", "mods.toml", "neoforge.mods.toml":
			priority = 1
		}
		if priority == 0 {
			return nil
		}
		dir := filepath.Dir(path)
		rel, _ := filepath.Rel(root, dir)
		depth := 0
		if rel != "." {
			depth = len(strings.Split(filepath.ToSlash(rel), "/"))
		}
		candidates = append(candidates, candidate{path: dir, priority: priority, depth: depth})
		return nil
	})
	if err != nil {
		return "", err
	}
	if len(candidates) == 0 {
		return root, nil
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].depth != candidates[j].depth {
			return candidates[i].depth < candidates[j].depth
		}
		if candidates[i].priority != candidates[j].priority {
			return candidates[i].priority > candidates[j].priority
		}
		return candidates[i].path < candidates[j].path
	})
	best := candidates[0]
	// A settings file or wrapper is the strongest root signal at the shallowest depth.
	for _, candidate := range candidates {
		if candidate.depth > best.depth {
			break
		}
		if candidate.priority > best.priority {
			best = candidate
		}
	}
	return best.path, nil
}

func repairIgnoredDirectory(name string) bool {
	switch strings.ToLower(name) {
	case ".git", ".gradle", ".idea", ".vscode", "build", "target", "out", "bin", "node_modules", ".mvn-cache", ".mmv-cache":
		return true
	default:
		return false
	}
}

func readRepairSmallTextFile(path string, limit int64) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return readAllLimited(file, limit)
}

func parseSimpleProperties(text string) map[string]string {
	out := map[string]string{}
	for _, line := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "!") {
			continue
		}
		idx := strings.IndexAny(trimmed, "=:")
		if idx <= 0 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(trimmed[:idx]))
		value := strings.TrimSpace(trimmed[idx+1:])
		if key != "" && value != "" {
			out[key] = value
		}
	}
	return out
}

func detectRepairBuildSystem(projectRoot string, buildFiles []string) (string, string) {
	for _, candidate := range []string{"gradlew", "gradlew.bat"} {
		if pathExists(filepath.Join(projectRoot, candidate)) {
			return "gradle", candidate
		}
	}
	for _, candidate := range []string{"mvnw", "mvnw.cmd"} {
		if pathExists(filepath.Join(projectRoot, candidate)) {
			return "maven", candidate
		}
	}
	for _, file := range buildFiles {
		base := strings.ToLower(filepath.Base(file))
		if strings.HasPrefix(base, "build.gradle") || strings.HasPrefix(base, "settings.gradle") || base == "gradle.properties" {
			return "gradle", ""
		}
		if base == "pom.xml" {
			return "maven", ""
		}
	}
	return "unknown", ""
}

func detectRepairLoader(metadataFiles []string, corpus string) (string, []string) {
	lower := strings.ToLower(corpus)
	set := map[string]bool{}
	for _, file := range metadataFiles {
		base := strings.ToLower(filepath.Base(file))
		switch base {
		case "fabric.mod.json":
			set["fabric"] = true
		case "quilt.mod.json":
			set["quilt"] = true
		case "neoforge.mods.toml":
			set["neoforge"] = true
		case "mods.toml":
			if strings.Contains(lower, "javafml") || strings.Contains(lower, "net.minecraftforge") {
				set["forge"] = true
			}
		}
	}
	if strings.Contains(lower, "net.neoforged") || strings.Contains(lower, "net.neoforge") || strings.Contains(lower, "moddevgradle") {
		set["neoforge"] = true
	}
	if strings.Contains(lower, "net.minecraftforge") || strings.Contains(lower, "forgegradle") {
		set["forge"] = true
	}
	if strings.Contains(lower, "fabric-loom") || strings.Contains(lower, "net.fabricmc") {
		set["fabric"] = true
	}
	if strings.Contains(lower, "org.quiltmc") || strings.Contains(lower, "quilt-loom") {
		set["quilt"] = true
	}
	order := []string{"neoforge", "forge", "quilt", "fabric"}
	var signals []string
	for _, loader := range order {
		if set[loader] {
			signals = append(signals, loader)
		}
	}
	if len(signals) == 0 {
		return "unknown", nil
	}
	if len(signals) == 1 {
		return signals[0], signals
	}
	// Multi-loader projects are identified explicitly while preserving every lane signal.
	return "multiloader", signals
}

func detectRepairGameVersion(loader string, properties map[string]string, corpus string) string {
	keys := []string{"minecraft_version", "minecraftversion", "minecraft.version", "mc_version", "mcversion", "minecraft_version_range"}
	for _, key := range keys {
		if value := cleanDetectedVersion(properties[key]); value != "" {
			return value
		}
	}
	for _, pattern := range repairGameCoordinatePatterns {
		match := pattern.FindStringSubmatch(corpus)
		if len(match) < 2 {
			continue
		}
		value := match[1]
		if strings.Contains(strings.ToLower(match[0]), "minecraftforge:forge") {
			if index := strings.Index(value, "-"); index > 0 {
				value = value[:index]
			}
		} else if loader == "neoforge" && !strings.HasPrefix(value, "1.") && !strings.HasPrefix(value, "2") {
			// NeoForge coordinates use the post-1.20 game line. Do not invent the leading 1.
			continue
		}
		if cleaned := cleanDetectedVersion(value); cleaned != "" {
			return cleaned
		}
	}
	// Read standard Fabric/Quilt metadata constraints without treating ranges as proof when ambiguous.
	for _, marker := range []string{`"minecraft"`, `'minecraft'`} {
		idx := strings.Index(corpus, marker)
		if idx < 0 {
			continue
		}
		window := corpus[idx:]
		if len(window) > 300 {
			window = window[:300]
		}
		versionPattern := regexp.MustCompile(`(?i)[~^=><\[\(\s\"]+((?:1\.)?\d+(?:\.\d+){0,2}|\d{2}\.\d+(?:[-A-Za-z0-9.]*)?)`)
		if match := versionPattern.FindStringSubmatch(window); len(match) > 1 {
			if cleaned := cleanDetectedVersion(match[1]); cleaned != "" {
				return cleaned
			}
		}
	}
	return ""
}

func cleanDetectedVersion(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"'[](){}~^=<>* ,`)
	if value == "" || strings.Contains(value, "$") || strings.Contains(strings.ToLower(value), "version") {
		return ""
	}
	if strings.HasPrefix(value, "1.") || regexp.MustCompile(`^\d{2}\.\d+`).MatchString(value) || strings.HasPrefix(value, "rd-") || strings.HasPrefix(value, "a") || strings.HasPrefix(value, "b") {
		return value
	}
	return ""
}

func detectRepairJavaMajor(properties map[string]string, corpus string) int {
	for _, key := range []string{"java_version", "javaversion", "java.version", "jvm_version", "jvmversion"} {
		if value := strings.TrimSpace(properties[key]); value != "" {
			if parsed, err := strconv.Atoi(strings.Trim(value, `"' `)); err == nil && parsed >= 6 && parsed <= 99 {
				return parsed
			}
		}
	}
	for _, pattern := range repairJavaPatterns {
		if match := pattern.FindStringSubmatch(corpus); len(match) > 1 {
			if parsed, err := strconv.Atoi(match[1]); err == nil && parsed >= 6 && parsed <= 99 {
				return parsed
			}
		}
	}
	return 0
}

func detectRepairModules(projectRoot string, buildFiles []string) []string {
	set := map[string]bool{".": true}
	for _, file := range buildFiles {
		base := strings.ToLower(filepath.Base(file))
		if base != "build.gradle" && base != "build.gradle.kts" && base != "pom.xml" {
			continue
		}
		dir := filepath.ToSlash(filepath.Dir(file))
		if dir == "." {
			dir = "."
		}
		set[dir] = true
	}
	modules := make([]string, 0, len(set))
	for module := range set {
		modules = append(modules, module)
	}
	sort.Strings(modules)
	return modules
}

func repairCommandsForProfile(profile RepairProjectProfile) []RepairCommandChoice {
	if profile.Wrapper == "" {
		return nil
	}
	switch profile.BuildSystem {
	case "gradle":
		return []RepairCommandChoice{
			{ID: "build", Label: "Build", Description: "Run the project's Gradle wrapper with a clean, non-daemon build and full stack traces.", Arguments: []string{"--no-daemon", "--stacktrace", "build"}},
			{ID: "test", Label: "Test", Description: "Run the project's declared Gradle test tasks.", Arguments: []string{"--no-daemon", "--stacktrace", "test"}},
			{ID: "clean", Label: "Clean", Description: "Remove project build outputs through the declared Gradle clean task.", Arguments: []string{"--no-daemon", "--stacktrace", "clean"}},
		}
	case "maven":
		return []RepairCommandChoice{
			{ID: "build", Label: "Package", Description: "Run the Maven wrapper through verification and packaging.", Arguments: []string{"-B", "-ntp", "verify"}},
			{ID: "test", Label: "Test", Description: "Run the Maven wrapper test lifecycle.", Arguments: []string{"-B", "-ntp", "test"}},
			{ID: "clean", Label: "Clean", Description: "Run the Maven wrapper clean lifecycle.", Arguments: []string{"-B", "-ntp", "clean"}},
		}
	}
	return nil
}

func repairProfileConfidence(profile RepairProjectProfile) string {
	score := 0
	if profile.BuildSystem != "unknown" {
		score++
	}
	if profile.Wrapper != "" {
		score += 2
	}
	if profile.Loader != "unknown" {
		score += 2
	}
	if profile.GameVersion != "" {
		score += 2
	}
	if len(profile.MetadataFiles) > 0 {
		score++
	}
	switch {
	case score >= 7:
		return "high"
	case score >= 4:
		return "medium"
	default:
		return "low"
	}
}

func repairProfileWarnings(profile RepairProjectProfile) []string {
	var warnings []string
	if profile.BuildSystem == "unknown" {
		warnings = append(warnings, "No supported Gradle or Maven project root was detected. Repair Lab can preserve and inspect the source but cannot execute a deterministic build command yet.")
	}
	if profile.Wrapper == "" && profile.BuildSystem != "unknown" {
		warnings = append(warnings, "The project has no checked-in build wrapper. Repair Lab will not silently substitute a host Gradle or Maven installation because that would weaken reproducibility.")
	}
	if profile.Loader == "unknown" {
		warnings = append(warnings, "Loader metadata was not proven from the source tree.")
	}
	if profile.Loader == "multiloader" {
		warnings = append(warnings, "Multiple loader lanes were detected. Automatic migration edits are conservative and each module still requires its own build/runtime proof.")
	}
	if profile.GameVersion == "" {
		warnings = append(warnings, "The source Minecraft version could not be proven from recognized metadata or build coordinates.")
	}
	return warnings
}

func repairProjectFingerprint(profile RepairProjectProfile, corpus string) string {
	payload, _ := json.Marshal(struct {
		Profile RepairProjectProfile `json:"profile"`
		Corpus  string               `json:"corpus"`
	}{Profile: profile, Corpus: corpus})
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:])
}
