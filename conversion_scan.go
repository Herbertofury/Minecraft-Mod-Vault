package main

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	conversionMaxArchiveEntries = 200000
	conversionMaxEntryBytes     = int64(2 << 30)
	conversionMaxGraphNodes     = 50000
)

type conversionExtractResult struct {
	FileCount      int
	ExtractedBytes int64
	RootHint       string
}

func extractConversionArchive(archivePath, destination string) (conversionExtractResult, error) {
	var result conversionExtractResult
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return result, fmt.Errorf("open conversion archive: %w", err)
	}
	defer reader.Close()
	if len(reader.File) == 0 {
		return result, errors.New("conversion archive is empty")
	}
	if len(reader.File) > conversionMaxArchiveEntries {
		return result, fmt.Errorf("conversion archive has %d entries; limit is %d", len(reader.File), conversionMaxArchiveEntries)
	}
	tmp := destination + ".extracting-" + randomToken(6)
	_ = os.RemoveAll(tmp)
	if err := os.MkdirAll(tmp, 0o755); err != nil {
		return result, err
	}
	defer os.RemoveAll(tmp)
	seen := map[string]string{}
	top := map[string]struct{}{}
	for _, entry := range reader.File {
		clean, err := safeArchiveEntryName(entry.Name)
		if err != nil {
			return result, errors.New(strings.Replace(err.Error(), "source ZIP", "conversion archive", 1))
		}
		if clean == "" {
			continue
		}
		fold := strings.ToLower(clean)
		if prior, ok := seen[fold]; ok {
			return result, fmt.Errorf("conversion archive contains a case-colliding duplicate path: %s and %s", prior, clean)
		}
		seen[fold] = clean
		first := strings.Split(clean, "/")[0]
		if first != "" {
			top[first] = struct{}{}
		}
		mode := entry.Mode()
		if mode&os.ModeSymlink != 0 || mode&os.ModeType != 0 && !mode.IsDir() {
			return result, fmt.Errorf("conversion archive contains an unsupported link or special file: %s", clean)
		}
		if entry.UncompressedSize64 > uint64(conversionMaxEntryBytes) {
			return result, fmt.Errorf("conversion archive entry exceeds the 2 GiB per-file limit: %s", clean)
		}
		if entry.CompressedSize64 > 0 && entry.UncompressedSize64 > entry.CompressedSize64*1000 && entry.UncompressedSize64 > 64<<20 {
			return result, fmt.Errorf("conversion archive entry has a suspicious compression ratio: %s", clean)
		}
		if entry.UncompressedSize64 > uint64(conversionMaxExtractedBytes-result.ExtractedBytes) {
			return result, fmt.Errorf("conversion archive expands beyond the %d-byte safety limit", conversionMaxExtractedBytes)
		}
		target := filepath.Join(tmp, filepath.FromSlash(clean))
		if !pathContainedBy(tmp, target) {
			return result, fmt.Errorf("conversion archive path escapes staging: %s", clean)
		}
		if entry.FileInfo().IsDir() || strings.HasSuffix(clean, "/") {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return result, err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return result, err
		}
		in, err := entry.Open()
		if err != nil {
			return result, err
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err != nil {
			in.Close()
			return result, err
		}
		written, copyErr := io.Copy(out, io.LimitReader(in, int64(entry.UncompressedSize64)+1))
		closeOutErr := out.Close()
		closeInErr := in.Close()
		if copyErr != nil {
			return result, copyErr
		}
		if closeOutErr != nil {
			return result, closeOutErr
		}
		if closeInErr != nil {
			return result, closeInErr
		}
		if written != int64(entry.UncompressedSize64) {
			return result, fmt.Errorf("conversion archive entry size mismatch: %s", clean)
		}
		result.FileCount++
		result.ExtractedBytes += written
	}
	if result.FileCount == 0 {
		return result, errors.New("conversion archive contains no regular files")
	}
	if len(top) == 1 {
		for name := range top {
			if info, err := os.Stat(filepath.Join(tmp, filepath.FromSlash(name))); err == nil && info.IsDir() {
				result.RootHint = name
			}
		}
	}
	if err := os.RemoveAll(destination); err != nil {
		return result, err
	}
	if err := os.Rename(tmp, destination); err != nil {
		return result, err
	}
	return result, nil
}

func (a *App) importConversionFile(name string, source io.Reader) (*ConversionSession, error) {
	session, err := a.newConversionSession(strings.TrimSuffix(filepath.Base(name), filepath.Ext(name)))
	if err != nil {
		return nil, err
	}
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.RemoveAll(session.Paths.Root)
		}
	}()
	ext := strings.ToLower(filepath.Ext(name))
	if !isConversionArchiveExtension(ext) {
		return nil, fmt.Errorf("unsupported conversion input %q; package it as JAR, ZIP, MCPACK, MCADDON, MCWORLD, MCTEMPLATE, MRPACK or a supported structure archive", ext)
	}
	originalPath := filepath.Join(session.Paths.Original, cleanConversionFilename(filepath.Base(name)))
	digest, size, err := copyAndHashLimited(originalPath, source, conversionMaxUploadBytes)
	if err != nil {
		return nil, err
	}
	var extracted conversionExtractResult
	if looksLikeZIPFile(originalPath) {
		extracted, err = extractConversionArchive(originalPath, session.Paths.Extracted)
		if err != nil {
			return nil, err
		}
		_ = expandNestedMinecraftArchives(session.Paths.Extracted)
	} else if isRawStructureExtension(ext) {
		if err := os.MkdirAll(filepath.Join(session.Paths.Extracted, "structures"), 0o755); err != nil {
			return nil, err
		}
		destination := filepath.Join(session.Paths.Extracted, "structures", cleanConversionFilename(filepath.Base(name)))
		if err := copyFileReplace(originalPath, destination); err != nil {
			return nil, err
		}
		extracted = conversionExtractResult{FileCount: 1, ExtractedBytes: size}
	} else {
		return nil, fmt.Errorf("%s is not a valid ZIP-based Minecraft package", filepath.Base(name))
	}
	treeDigest, files, bytesCount, err := hashDirectoryTree(session.Paths.Extracted)
	if err != nil {
		return nil, err
	}
	profile, graph, err := a.profileConversionSource(originalPath, session.Paths.Extracted, filepath.Base(name), digest, size, treeDigest, files, bytesCount)
	if err != nil {
		return nil, err
	}
	if extracted.RootHint != "" {
		profile.Signals = append(profile.Signals, "single archive root: "+extracted.RootHint)
	}
	session.Source = profile
	session.Graph = graph
	session.State = "profiled"
	session.Phase = "universal-graph-ready"
	session.Warnings = uniqueStringsPreserve(append(session.Warnings, profile.Warnings...))
	if err := a.saveConversionSession(session); err != nil {
		return nil, err
	}
	cleanup = false
	return session, nil
}

func (a *App) importConversionPath(path string) (*ConversionSession, error) {
	path = filepath.Clean(path)
	if !a.allowedLibraryPath(path) {
		return nil, errors.New("path is outside configured Minecraft and Vault roots")
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		tmpRoot, err := os.MkdirTemp(a.conversionRoot(), "path-import-")
		if err != nil {
			return nil, err
		}
		defer os.RemoveAll(tmpRoot)
		archive := filepath.Join(tmpRoot, cleanConversionFilename(filepath.Base(path))+".zip")
		if _, _, err := zipDirectoryDeterministic(path, archive, nil); err != nil {
			return nil, err
		}
		file, err := os.Open(archive)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		return a.importConversionFile(filepath.Base(archive), file)
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return a.importConversionFile(filepath.Base(path), file)
}

func looksLikeZIPFile(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()
	var magic [4]byte
	if _, err := io.ReadFull(file, magic[:]); err != nil {
		return false
	}
	return magic == [4]byte{'P', 'K', 3, 4} || magic == [4]byte{'P', 'K', 5, 6} || magic == [4]byte{'P', 'K', 7, 8}
}

func isRawStructureExtension(ext string) bool {
	switch strings.ToLower(ext) {
	case ".schem", ".litematic", ".nbt", ".bp":
		return true
	default:
		return false
	}
}

func isConversionArchiveExtension(ext string) bool {
	switch strings.ToLower(ext) {
	case ".jar", ".zip", ".mrpack", ".mcpack", ".mcaddon", ".mcworld", ".mctemplate", ".schem", ".litematic", ".nbt", ".bp":
		return true
	default:
		return false
	}
}

func cleanConversionFilename(name string) string {
	name = filepath.Base(strings.TrimSpace(name))
	name = strings.Map(func(r rune) rune {
		if r < 32 || strings.ContainsRune(`<>:"/\\|?*`, r) {
			return '-'
		}
		return r
	}, name)
	if name == "" || name == "." {
		return "conversion-input.zip"
	}
	return name
}

func expandNestedMinecraftArchives(root string) error {
	archives := []string{}
	_ = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".mcpack" || ext == ".mcaddon" || ext == ".zip" {
			archives = append(archives, path)
		}
		return nil
	})
	if len(archives) > 32 {
		archives = archives[:32]
	}
	for index, archive := range archives {
		destination := filepath.Join(root, ".omnibridge-nested", fmt.Sprintf("%02d-%s", index, strings.TrimSuffix(cleanConversionFilename(filepath.Base(archive)), filepath.Ext(archive))))
		_, _ = extractConversionArchive(archive, destination)
	}
	return nil
}

func (a *App) profileConversionSource(originalPath, root, filename, sourceSHA string, size int64, treeSHA string, files int, bytesCount int64) (ConversionSourceProfile, UniversalContentGraph, error) {
	profile := ConversionSourceProfile{Filename: filepath.Base(filename), Path: originalPath, Size: size, SHA256: sourceSHA, TreeSHA256: treeSHA, FileCount: files, ExtractedBytes: bytesCount, Name: humanizeMinecraftFilename(filename), Namespace: "omnibridge", Metadata: map[string]string{}}
	ext := strings.ToLower(filepath.Ext(filename))
	profile.Format = strings.TrimPrefix(ext, ".")
	profile.Kind = "archive"
	profile.Edition = "unknown"
	lowName := strings.ToLower(filename)
	if ext == ".jar" {
		profile.Edition, profile.Kind, profile.Format = "java", "mod", "jar"
		if local, err := inspectLocalJar(originalPath); err == nil {
			profile.Name = firstNonEmpty(local.Metadata.Name, profile.Name)
			profile.Description = local.Metadata.Description
			profile.Namespace = firstNonEmpty(local.Metadata.ModID, sanitizeNamespace(profile.Name))
			profile.Version = local.Metadata.Version
			profile.GameVersion = local.Metadata.Minecraft
			profile.Loader = firstString(local.Metadata.Loaders)
			profile.Signals = append(profile.Signals, "embedded "+local.Metadata.MetadataBy+" metadata")
			profile.Metadata["modId"] = local.Metadata.ModID
			profile.Metadata["license"] = local.Metadata.License
			profile.Metadata["homepage"] = local.Metadata.Homepage
		}
	}
	manifestPath, manifest := findBedrockManifest(root)
	if manifestPath != "" {
		relManifest, _ := filepath.Rel(root, manifestPath)
		applyBedrockManifestProfile(&profile, filepath.ToSlash(relManifest), manifest)
	}
	packPath, packMeta := findPackMCMeta(root)
	if packPath != "" && profile.Edition == "unknown" {
		relPack, _ := filepath.Rel(root, packPath)
		packPath = filepath.ToSlash(relPack)
		profile.Edition = "java"
		profile.Format = "zip"
		profile.ManifestPath = filepath.ToSlash(packPath)
		if pack, ok := packMeta["pack"].(map[string]any); ok {
			if format := intFlexible(pack["pack_format"]); format > 0 {
				profile.PackFormat = format
			}
			if desc := minecraftTextComponent(pack["description"]); desc != "" {
				profile.Description = desc
			}
		}
		if pathExists(filepath.Join(root, "data")) || containsPathSegment(root, "data") {
			profile.Kind = "datapack"
		} else {
			profile.Kind = "resourcepack"
		}
		profile.Signals = append(profile.Signals, "pack.mcmeta")
	}
	if hasWorldSignals(root) {
		profile.Kind = "world"
		if profile.Edition == "unknown" {
			if hasFileBelow(root, "db") {
				profile.Edition = "bedrock"
			} else {
				profile.Edition = "java"
			}
		}
	}
	if ext == ".mctemplate" || strings.Contains(strings.Join(profile.Signals, " "), "world_template") {
		profile.Edition, profile.Kind, profile.Format = "bedrock", "world-template", "mctemplate"
	}
	if ext == ".mcworld" {
		profile.Edition, profile.Kind, profile.Format = "bedrock", "world", "mcworld"
	}
	if ext == ".mcaddon" {
		profile.Edition, profile.Kind, profile.Format = "bedrock", "addon-family", "mcaddon"
	}
	if ext == ".mcpack" {
		profile.Edition, profile.Format = "bedrock", "mcpack"
		if profile.Kind == "archive" {
			profile.Kind = "pack"
		}
	}
	if ext == ".mrpack" {
		profile.Edition, profile.Kind, profile.Format = "java", "modpack", "mrpack"
	}
	if ext == ".schem" || ext == ".litematic" || ext == ".nbt" || ext == ".bp" {
		profile.Kind, profile.Format = "structure", strings.TrimPrefix(ext, ".")
	}
	if hasBuildProjectSignals(root) && profile.Kind == "archive" {
		profile.Edition, profile.Kind, profile.Format = "java", "source-project", "source-zip"
	}
	if profile.Namespace == "omnibridge" {
		profile.Namespace = sanitizeNamespace(profile.Name)
	}
	if profile.Namespace == "" {
		profile.Namespace = "converted"
	}
	if profile.Edition == "unknown" {
		profile.Warnings = append(profile.Warnings, "Edition could not be proven from authoritative manifests; conversion will preserve the archive and require target review.")
	}
	if strings.Contains(lowName, "patched") || strings.Contains(lowName, "custom") || strings.Contains(lowName, "fixed") {
		profile.Warnings = append(profile.Warnings, "Patched/custom source detected; generated outputs retain the immutable source hash and will not overwrite public artifacts.")
	}
	graph, err := buildUniversalContentGraph(root, profile)
	return profile, graph, err
}

func findBedrockManifest(root string) (string, map[string]any) {
	paths := []string{}
	_ = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		if strings.EqualFold(entry.Name(), "manifest.json") {
			paths = append(paths, path)
		}
		return nil
	})
	sort.Strings(paths)
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil || len(data) > 4<<20 {
			continue
		}
		var value map[string]any
		if json.Unmarshal(data, &value) == nil {
			if _, ok := value["header"]; ok {
				return path, value
			}
		}
	}
	return "", nil
}

func applyBedrockManifestProfile(profile *ConversionSourceProfile, path string, value map[string]any) {
	profile.Edition = "bedrock"
	profile.ManifestPath = filepath.ToSlash(path)
	profile.Signals = append(profile.Signals, "Bedrock manifest.json")
	if header, ok := value["header"].(map[string]any); ok {
		profile.Name = firstNonEmpty(flexibleString(header["name"]), profile.Name)
		profile.Description = flexibleString(header["description"])
		profile.UUID = flexibleString(header["uuid"])
		profile.Version = versionArrayString(header["version"])
		profile.MinimumEngine = versionArrayString(header["min_engine_version"])
	}
	moduleTypes := []string{}
	if modules, ok := value["modules"].([]any); ok {
		for _, raw := range modules {
			if module, ok := raw.(map[string]any); ok {
				typ := flexibleString(module["type"])
				if typ != "" {
					moduleTypes = append(moduleTypes, typ)
				}
			}
		}
	}
	profile.Metadata["moduleTypes"] = strings.Join(uniqueStrings(moduleTypes), ",")
	switch {
	case containsStringFold(moduleTypes, "world_template"):
		profile.Kind, profile.Format = "world-template", "mctemplate"
		profile.Signals = append(profile.Signals, "world_template module")
	case containsStringFold(moduleTypes, "resources"):
		profile.Kind = "resource-pack"
	case containsStringFold(moduleTypes, "data") || containsStringFold(moduleTypes, "script"):
		profile.Kind = "behavior-pack"
	default:
		profile.Kind = "bedrock-pack"
	}
	profile.Namespace = sanitizeNamespace(profile.Name)
}

func findPackMCMeta(root string) (string, map[string]any) {
	var found string
	_ = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || found != "" {
			return err
		}
		if strings.EqualFold(entry.Name(), "pack.mcmeta") {
			found = path
		}
		return nil
	})
	if found == "" {
		return "", nil
	}
	data, err := os.ReadFile(found)
	if err != nil {
		return "", nil
	}
	var value map[string]any
	if json.Unmarshal(data, &value) != nil {
		return found, nil
	}
	return found, value
}

func buildUniversalContentGraph(root string, profile ConversionSourceProfile) (UniversalContentGraph, error) {
	graph := UniversalContentGraph{SchemaVersion: conversionSchemaVersion, GraphVersion: conversionGraphVersion, GeneratedAt: time.Now().UTC().Format(time.RFC3339), SourceSHA256: profile.SHA256, Namespace: profile.Namespace, Summary: UniversalSummary{ByKind: map[string]int{}}}
	type aggregate struct {
		kind, namespace, name string
		count                 int
		bytes                 int64
		examples              []string
	}
	aggregates := map[string]*aggregate{}
	paths := []string{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		return graph, err
	}
	sort.Strings(paths)
	for _, path := range paths {
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		low := strings.ToLower(rel)
		if strings.HasPrefix(low, ".omnibridge-nested/") && strings.HasSuffix(low, ".mcpack") {
			continue
		}
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		if strings.HasSuffix(low, ".class") || strings.HasSuffix(low, ".java") || strings.HasSuffix(low, ".kt") {
			kind := "java-bytecode"
			if !strings.HasSuffix(low, ".class") {
				kind = "java-source"
			}
			pkg := aggregatePackage(rel)
			key := kind + ":" + pkg
			agg := aggregates[key]
			if agg == nil {
				agg = &aggregate{kind: kind, namespace: profile.Namespace, name: pkg}
				aggregates[key] = agg
			}
			agg.count++
			agg.bytes += info.Size()
			if len(agg.examples) < 8 {
				agg.examples = append(agg.examples, rel)
			}
			continue
		}
		kind, namespace, name, level, support := classifyUniversalPath(rel, profile)
		if kind == "" {
			continue
		}
		if len(graph.Nodes) >= conversionMaxGraphNodes {
			graph.Warnings = append(graph.Warnings, fmt.Sprintf("Universal graph reached the %d-node safety cap; remaining unclassified files are preserved in the immutable source.", conversionMaxGraphNodes))
			break
		}
		node := UniversalNode{ID: universalNodeID(kind, rel), Kind: kind, Namespace: firstNonEmpty(namespace, profile.Namespace), Name: firstNonEmpty(name, strings.TrimSuffix(filepath.Base(rel), filepath.Ext(rel))), SourcePath: rel, SourceFormat: profile.Format, Level: level, Confidence: .86, TargetSupport: support, Properties: map[string]string{"size": strconv.FormatInt(info.Size(), 10)}}
		populateNodeData(path, &node)
		graph.Nodes = append(graph.Nodes, node)
	}
	keys := make([]string, 0, len(aggregates))
	for key := range aggregates {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		agg := aggregates[key]
		level := ConversionReview
		support := []string{"java-fabric", "java-neoforge", "java-forge", "java-multiloader", "universal-bundle"}
		notes := []string{"Executable logic cannot be translated by file copying. OmniBridge emits a target-language contract and review queue."}
		if agg.kind == "java-source" {
			level = ConversionGenerated
			notes = []string{"Source can seed a target project, but loader APIs and edition semantics still require compilation and runtime tests."}
		}
		node := UniversalNode{ID: universalNodeID(agg.kind, agg.name), Kind: agg.kind, Namespace: agg.namespace, Name: agg.name, SourcePath: strings.Join(agg.examples, ", "), SourceFormat: profile.Format, Level: level, Confidence: .92, TargetSupport: support, RequiresReview: true, Properties: map[string]string{"fileCount": strconv.Itoa(agg.count), "bytes": strconv.FormatInt(agg.bytes, 10)}, Notes: notes}
		graph.Nodes = append(graph.Nodes, node)
	}
	addManifestAndWorldNodes(root, profile, &graph)
	sort.Slice(graph.Nodes, func(i, j int) bool {
		if graph.Nodes[i].Kind == graph.Nodes[j].Kind {
			return graph.Nodes[i].SourcePath < graph.Nodes[j].SourcePath
		}
		return graph.Nodes[i].Kind < graph.Nodes[j].Kind
	})
	computeUniversalSummary(&graph)
	return graph, nil
}

func classifyUniversalPath(rel string, profile ConversionSourceProfile) (kind, namespace, name string, level ConversionLevel, support []string) {
	clean := strings.TrimPrefix(filepath.ToSlash(rel), "./")
	parts := strings.Split(clean, "/")
	low := strings.ToLower(clean)
	level = ConversionTranslated
	support = []string{"bedrock-addon", "java-datapack", "java-resourcepack", "universal-bundle"}
	if len(parts) >= 4 && strings.EqualFold(parts[0], "assets") {
		namespace = parts[1]
		category := strings.ToLower(parts[2])
		name = strings.TrimSuffix(strings.Join(parts[3:], "/"), filepath.Ext(clean))
		switch category {
		case "textures":
			kind, support = "texture", []string{"bedrock-addon", "bedrock-resource", "java-resourcepack", "java-fabric", "java-neoforge", "java-forge", "java-multiloader", "universal-bundle"}
		case "models", "items", "blockstates":
			kind, level, support = "model", ConversionReview, []string{"bedrock-addon", "bedrock-resource", "java-resourcepack", "java-fabric", "java-neoforge", "java-forge", "java-multiloader", "universal-bundle"}
		case "lang":
			kind, support = "language", []string{"bedrock-addon", "bedrock-resource", "java-resourcepack", "java-fabric", "java-neoforge", "java-forge", "java-multiloader", "universal-bundle"}
		case "sounds":
			kind, support = "sound", []string{"bedrock-addon", "bedrock-resource", "java-resourcepack", "universal-bundle"}
		case "particles":
			kind, level = "particle", ConversionReview
		case "shaders":
			kind, level = "shader", ConversionBlocked
		case "font":
			kind, level = "font", ConversionReview
		}
		return
	}
	if len(parts) >= 4 && strings.EqualFold(parts[0], "data") {
		namespace = parts[1]
		category := strings.ToLower(parts[2])
		name = strings.TrimSuffix(strings.Join(parts[3:], "/"), filepath.Ext(clean))
		support = []string{"bedrock-addon", "bedrock-behavior", "java-datapack", "java-fabric", "java-neoforge", "java-forge", "java-multiloader", "universal-bundle"}
		switch category {
		case "recipes", "recipe":
			kind = "recipe"
		case "loot_tables", "loot_table":
			kind, level = "loot-table", ConversionReview
		case "functions", "function":
			kind, level = "function", ConversionReview
		case "tags":
			kind = "tag"
		case "advancements", "advancement":
			kind, level = "advancement", ConversionReview
		case "predicates", "predicate":
			kind, level = "predicate", ConversionReview
		case "item_modifiers", "item_modifier":
			kind, level = "item-modifier", ConversionReview
		case "worldgen":
			kind, level = "worldgen", ConversionToolAssisted
		case "structures", "structure":
			kind, level = "structure", ConversionToolAssisted
		case "dimension", "dimension_type":
			kind, level = "dimension", ConversionReview
		}
		return
	}
	bedrockCategories := map[string]string{
		"blocks": "block", "items": "item", "entities": "entity", "recipes": "recipe", "loot_tables": "loot-table", "functions": "function", "scripts": "bedrock-script",
		"animations": "animation", "animation_controllers": "animation-controller", "render_controllers": "render-controller", "particles": "particle", "textures": "texture", "models": "model", "sounds": "sound", "texts": "language", "ui": "ui", "font": "font", "fonts": "font",
		"structures": "structure", "biomes": "biome", "spawn_rules": "spawn-rule", "features": "feature", "feature_rules": "feature-rule", "attachables": "attachable", "fogs": "fog", "fog": "fog", "materials": "material",
		"camera_presets": "camera-preset", "dialogue": "dialogue", "trading": "trade-table", "trading_economy": "trade-table", "trade_tables": "trade-table", "volumes": "volume", "dimensions": "dimension", "worldgen": "worldgen",
	}
	for segment, mapped := range bedrockCategories {
		for index, part := range parts {
			if strings.EqualFold(part, segment) {
				kind = mapped
				name = strings.TrimSuffix(strings.Join(parts[index+1:], "/"), filepath.Ext(clean))
				namespace = profile.Namespace
				level = ConversionTranslated
				if mapped == "bedrock-script" || mapped == "render-controller" || mapped == "ui" || mapped == "material" || mapped == "worldgen" || mapped == "camera-preset" || mapped == "dialogue" || mapped == "trade-table" || mapped == "volume" {
					level = ConversionReview
				}
				support = []string{"bedrock-addon", "bedrock-behavior", "bedrock-resource", "java-datapack", "java-resourcepack", "java-fabric", "java-neoforge", "java-forge", "java-multiloader", "universal-bundle"}
				return
			}
		}
	}
	switch {
	case strings.EqualFold(filepath.Base(clean), "pack.png") || strings.EqualFold(filepath.Base(clean), "pack_icon.png") || strings.HasSuffix(low, "world_icon.jpeg") || strings.HasSuffix(low, "world_icon.png"):
		return "pack-icon", profile.Namespace, filepath.Base(clean), ConversionExact, []string{"bedrock-addon", "bedrock-behavior", "bedrock-resource", "bedrock-template", "java-datapack", "java-resourcepack", "universal-bundle"}
	case strings.HasSuffix(low, "mixins.json") || strings.Contains(low, "mixin") && strings.HasSuffix(low, ".json"):
		return "mixin", profile.Namespace, filepath.Base(clean), ConversionBlocked, []string{"java-fabric", "java-neoforge", "java-forge", "java-multiloader", "universal-bundle"}
	case strings.HasSuffix(low, ".mcfunction"):
		return "function", profile.Namespace, strings.TrimSuffix(filepath.Base(clean), ".mcfunction"), ConversionReview, []string{"bedrock-addon", "bedrock-behavior", "java-datapack", "universal-bundle"}
	case strings.HasSuffix(low, ".ogg") || strings.HasSuffix(low, ".wav"):
		return "sound", profile.Namespace, filepath.Base(clean), ConversionExact, []string{"bedrock-addon", "bedrock-resource", "java-resourcepack", "universal-bundle"}
	case strings.HasSuffix(low, ".png") || strings.HasSuffix(low, ".tga") || strings.HasSuffix(low, ".jpg") || strings.HasSuffix(low, ".jpeg"):
		return "texture", profile.Namespace, filepath.Base(clean), ConversionExact, []string{"bedrock-addon", "bedrock-resource", "java-resourcepack", "universal-bundle"}
	case strings.HasSuffix(low, ".nbt") || strings.HasSuffix(low, ".schem") || strings.HasSuffix(low, ".litematic") || strings.HasSuffix(low, ".mcstructure") || strings.HasSuffix(low, ".bp"):
		return "structure", profile.Namespace, filepath.Base(clean), ConversionToolAssisted, []string{"bedrock-addon", "bedrock-behavior", "java-datapack", "universal-bundle"}
	}
	return "", "", "", "", nil
}

func populateNodeData(path string, node *UniversalNode) {
	if node == nil {
		return
	}
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".json" && ext != ".mcmeta" {
		return
	}
	data, err := os.ReadFile(path)
	if err != nil || len(data) > 8<<20 {
		return
	}
	var value any
	if json.Unmarshal(data, &value) != nil {
		node.Notes = append(node.Notes, "JSON could not be parsed; original bytes remain preserved.")
		node.RequiresReview = true
		return
	}
	if object, ok := value.(map[string]any); ok {
		node.Data = object
		if format := flexibleString(object["format_version"]); format != "" {
			node.Properties["formatVersion"] = format
		}
		collectIdentifierProperties(object, node)
	}
}

func collectIdentifierProperties(value map[string]any, node *UniversalNode) {
	for _, key := range []string{"type", "parent", "identifier", "category"} {
		if text := flexibleString(value[key]); text != "" {
			node.Properties[key] = text
		}
	}
	for key, raw := range value {
		if !strings.HasPrefix(key, "minecraft:") {
			continue
		}
		object, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if description, ok := object["description"].(map[string]any); ok {
			if identifier := flexibleString(description["identifier"]); identifier != "" {
				node.Properties["identifier"] = identifier
			}
		}
	}
}

func addManifestAndWorldNodes(root string, profile ConversionSourceProfile, graph *UniversalContentGraph) {
	if profile.ManifestPath != "" {
		graph.Nodes = append(graph.Nodes, UniversalNode{ID: universalNodeID("manifest", profile.ManifestPath), Kind: "manifest", Namespace: profile.Namespace, Name: filepath.Base(profile.ManifestPath), SourcePath: profile.ManifestPath, SourceFormat: profile.Format, Level: ConversionExact, Confidence: 1, TargetSupport: []string{"bedrock-addon", "bedrock-behavior", "bedrock-resource", "bedrock-template", "java-datapack", "java-resourcepack", "java-fabric", "java-neoforge", "java-forge", "java-multiloader", "universal-bundle"}})
	}
	if hasWorldSignals(root) {
		level := ConversionExact
		if profile.Edition == "unknown" {
			level = ConversionReview
		}
		graph.Nodes = append(graph.Nodes, UniversalNode{ID: universalNodeID("world-data", profile.Name), Kind: "world-data", Namespace: profile.Namespace, Name: profile.Name, SourcePath: ".", SourceFormat: profile.Format, Level: level, Confidence: .95, TargetSupport: []string{"bedrock-world", "bedrock-template", "java-world", "universal-bundle"}, Notes: []string{"Cross-edition chunk and entity translation uses Chunker, je2be or Amulet when an adapter is installed."}})
	}
}

func computeUniversalSummary(graph *UniversalContentGraph) {
	graph.Summary = UniversalSummary{ByKind: map[string]int{}}
	for _, node := range graph.Nodes {
		graph.Summary.Total++
		graph.Summary.ByKind[node.Kind]++
		switch node.Kind {
		case "texture", "model", "sound", "language", "particle", "animation", "animation-controller", "render-controller", "ui", "font", "shader", "pack-icon", "material", "attachable", "fog", "camera-preset":
			graph.Summary.Assets++
		case "recipe", "loot-table", "function", "tag", "advancement", "predicate", "item-modifier", "worldgen", "structure", "dimension", "biome", "spawn-rule", "feature", "feature-rule", "dialogue", "trade-table", "volume":
			graph.Summary.Data++
		case "java-bytecode", "java-source", "bedrock-script", "mixin":
			graph.Summary.Logic++
		case "world-data":
			graph.Summary.World++
		}
		switch node.Level {
		case ConversionExact:
			graph.Summary.Exact++
		case ConversionTranslated:
			graph.Summary.Translated++
		case ConversionGenerated:
			graph.Summary.Generated++
		case ConversionToolAssisted:
			graph.Summary.ToolAssisted++
		case ConversionReview:
			graph.Summary.ReviewRequired++
		case ConversionBlocked:
			graph.Summary.Blocked++
		}
	}
}

func universalNodeID(kind, value string) string {
	hash := sha256.Sum256([]byte(kind + "\x00" + filepath.ToSlash(value)))
	return kind + ":" + hex.EncodeToString(hash[:8])
}

func aggregatePackage(rel string) string {
	parts := strings.Split(filepath.ToSlash(rel), "/")
	if len(parts) > 4 {
		parts = parts[:4]
	}
	name := strings.Join(parts, "/")
	name = strings.TrimSuffix(name, filepath.Ext(name))
	if name == "" {
		return "root"
	}
	return name
}

func sanitizeNamespace(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	for _, r := range value {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_' || r == '-' || r == '.' {
			builder.WriteRune(r)
		} else if r == ' ' {
			builder.WriteByte('_')
		}
	}
	value = strings.Trim(builder.String(), "-._")
	if value == "" {
		return "converted"
	}
	if len(value) > 64 {
		value = value[:64]
	}
	return value
}

func versionArrayString(value any) string {
	list, ok := value.([]any)
	if !ok {
		return flexibleString(value)
	}
	parts := []string{}
	for _, item := range list {
		parts = append(parts, flexibleString(item))
	}
	return strings.Join(parts, ".")
}

func intFlexible(value any) int {
	switch current := value.(type) {
	case float64:
		return int(current)
	case int:
		return current
	case json.Number:
		value, _ := current.Int64()
		return int(value)
	case string:
		value, _ := strconv.Atoi(current)
		return value
	default:
		return 0
	}
}

func containsStringFold(values []string, wanted string) bool {
	for _, value := range values {
		if strings.EqualFold(value, wanted) {
			return true
		}
	}
	return false
}

func hasWorldSignals(root string) bool {
	return hasFileName(root, "level.dat") || hasFileBelow(root, "region") || hasFileBelow(root, "db")
}

func hasFileName(root, name string) bool {
	found := false
	_ = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || found {
			return err
		}
		if !entry.IsDir() && strings.EqualFold(entry.Name(), name) {
			found = true
		}
		return nil
	})
	return found
}

func hasFileBelow(root, directory string) bool {
	found := false
	_ = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || found {
			return err
		}
		if entry.IsDir() && strings.EqualFold(entry.Name(), directory) {
			children, readErr := os.ReadDir(path)
			if readErr == nil && len(children) > 0 {
				found = true
			}
		}
		return nil
	})
	return found
}

func containsPathSegment(root, segment string) bool {
	found := false
	_ = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || found {
			return err
		}
		if entry.IsDir() && strings.EqualFold(entry.Name(), segment) {
			found = true
		}
		return nil
	})
	return found
}

func hasBuildProjectSignals(root string) bool {
	for _, name := range []string{"gradlew", "gradlew.bat", "build.gradle", "build.gradle.kts", "pom.xml", "settings.gradle", "settings.gradle.kts"} {
		if hasFileName(root, name) {
			return true
		}
	}
	return false
}
