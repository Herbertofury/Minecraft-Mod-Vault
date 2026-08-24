package main

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type bedrockManifest struct {
	FormatVersion string
	Name          string
	Description   string
	UUID          string
	Version       string
	MinEngine     string
	Modules       []string
	Dependencies  []LibraryDependency
	Capabilities  []string
	Authors       []string
	License       string
	URL           string
	Kind          string
	MetadataBy    string
}

type bedrockManifestJSON struct {
	FormatVersion any `json:"format_version"`
	Header        struct {
		Name                string `json:"name"`
		Description         string `json:"description"`
		UUID                string `json:"uuid"`
		Version             any    `json:"version"`
		MinEngineVersion    any    `json:"min_engine_version"`
		PackScope           string `json:"pack_scope"`
		LockTemplateOptions bool   `json:"lock_template_options"`
	} `json:"header"`
	Modules []struct {
		Type        string `json:"type"`
		UUID        string `json:"uuid"`
		Version     any    `json:"version"`
		Language    string `json:"language"`
		Entry       string `json:"entry"`
		Description string `json:"description"`
	} `json:"modules"`
	Dependencies []struct {
		UUID       string `json:"uuid"`
		ModuleName string `json:"module_name"`
		Version    any    `json:"version"`
	} `json:"dependencies"`
	Capabilities []string `json:"capabilities"`
	Metadata     struct {
		Authors []string `json:"authors"`
		License string   `json:"license"`
		URL     string   `json:"url"`
	} `json:"metadata"`
}

func (a *App) scanBedrockRoot(ctx context.Context, root, kind, profile string) ([]LibraryItem, []string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, nil, err
	}
	items := []LibraryItem{}
	warnings := []string{}
	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return items, warnings, ctx.Err()
		default:
		}
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		path := filepath.Join(root, entry.Name())
		var item LibraryItem
		var inspectErr error
		if strings.Contains(kind, "world") {
			item, inspectErr = a.inspectBedrockWorld(path, kind, profile)
		} else {
			item, inspectErr = a.inspectBedrockPackDirectory(path, kind, profile)
		}
		if inspectErr != nil {
			warnings = append(warnings, fmt.Sprintf("%s: %v", path, inspectErr))
			continue
		}
		items = append(items, item)
	}
	return items, warnings, nil
}

func (a *App) inspectBedrockPackDirectory(path, kind, profile string) (LibraryItem, error) {
	info, err := os.Stat(path)
	if err != nil {
		return LibraryItem{}, err
	}
	if !info.IsDir() {
		return LibraryItem{}, errors.New("Bedrock pack entry is not a directory")
	}
	manifestPath := filepath.Join(path, "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return LibraryItem{}, errors.New("manifest.json is missing")
	}
	translations := loadBedrockTranslations(path)
	manifest, err := parseBedrockManifest(data, translations)
	if err != nil {
		return LibraryItem{}, err
	}
	digest, _, bytesCount, err := hashDirectoryTree(path)
	if err != nil {
		return LibraryItem{}, err
	}
	resolvedKind := firstNonEmpty(manifest.Kind, normalizeBedrockKind(kind))
	item := LibraryItem{
		ID:   "bedrock:" + firstNonEmpty(strings.ToLower(manifest.UUID), "no-uuid") + ":" + digest,
		Path: path, Filename: filepath.Base(path), Name: firstNonEmpty(manifest.Name, humanizeMinecraftFilename(filepath.Base(path))),
		Summary: manifest.Description, Description: manifest.Description, Authors: manifest.Authors,
		Edition: "bedrock", Kind: resolvedKind, Profile: profile, InstalledVersion: manifest.Version,
		UUID: strings.ToLower(manifest.UUID), MinEngineVersion: manifest.MinEngine, Modules: manifest.Modules,
		Capabilities: manifest.Capabilities, Dependencies: manifest.Dependencies, License: manifest.License,
		Homepage: manifest.URL, Enabled: true, IsDir: true, Size: bytesCount,
		Modified: info.ModTime().UTC().Format(time.RFC3339), Hashes: LibraryHashes{SHA256: digest},
		MetadataBy: "manifest.json", ManagedRoot: filepath.Dir(path), UpdateStatus: "unknown",
		UpdateMessage:        "Bedrock manifest identity loaded. Provider releases require evidence-backed matching.",
		ProvenanceConfidence: .9, MatchEvidence: []string{"Embedded Bedrock manifest UUID and version"},
	}
	if manifest.MinEngine != "" {
		item.GameVersions = []string{"Bedrock engine " + manifest.MinEngine + "+"}
	}
	if hasScriptModule(manifest.Modules) {
		item.Warnings = append(item.Warnings, "This add-on contains script modules; engine and @minecraft/server API compatibility must be verified before updating.")
	}
	if icon := firstExistingPath(filepath.Join(path, "pack_icon.png"), filepath.Join(path, "world_icon.jpeg"), filepath.Join(path, "world_icon.png")); icon != "" {
		if artURL, artErr := a.cacheLocalArtFile(item.ID, icon); artErr == nil {
			item.LocalArtURL, item.ArtOrigin = artURL, "embedded-local"
		}
	}
	return item, nil
}

func (a *App) inspectBedrockWorld(path, kind, profile string) (LibraryItem, error) {
	info, err := os.Stat(path)
	if err != nil {
		return LibraryItem{}, err
	}
	if !info.IsDir() {
		return LibraryItem{}, errors.New("Bedrock world entry is not a directory")
	}
	digest, _, bytesCount, err := hashDirectoryTree(path)
	if err != nil {
		return LibraryItem{}, err
	}
	name := ""
	if data, readErr := os.ReadFile(filepath.Join(path, "levelname.txt")); readErr == nil {
		name = strings.TrimSpace(strings.TrimPrefix(string(data), "\ufeff"))
	}
	item := LibraryItem{
		ID: "bedrock-world:" + digest, Path: path, Filename: filepath.Base(path), Name: firstNonEmpty(name, humanizeMinecraftFilename(filepath.Base(path))),
		Edition: "bedrock", Kind: normalizeBedrockKind(kind), Profile: profile, Enabled: true, IsDir: true, Size: bytesCount,
		Modified: info.ModTime().UTC().Format(time.RFC3339), Hashes: LibraryHashes{SHA256: digest}, MetadataBy: "levelname.txt / level.dat",
		ManagedRoot: filepath.Dir(path), UpdateStatus: "current", UpdateMessage: "Local Bedrock world managed by Vault.",
		ProvenanceConfidence: .9, MatchEvidence: []string{"Bedrock world directory and level metadata"},
	}
	if icon := firstExistingPath(filepath.Join(path, "world_icon.jpeg"), filepath.Join(path, "world_icon.png")); icon != "" {
		if artURL, artErr := a.cacheLocalArtFile(item.ID, icon); artErr == nil {
			item.LocalArtURL, item.ArtOrigin = artURL, "embedded-local"
		}
	}
	return item, nil
}

func normalizeBedrockKind(kind string) string {
	switch strings.ToLower(kind) {
	case "behaviorpack-dev":
		return "behaviorpack"
	case "resourcepack-bedrock", "resourcepack-bedrock-dev":
		return "resourcepack"
	case "world-bedrock":
		return "world"
	default:
		return strings.ToLower(kind)
	}
}

func parseBedrockManifest(data []byte, translations map[string]string) (bedrockManifest, error) {
	var raw bedrockManifestJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return bedrockManifest{}, fmt.Errorf("parse manifest.json: %w", err)
	}
	if strings.TrimSpace(raw.Header.UUID) == "" {
		return bedrockManifest{}, errors.New("manifest header UUID is missing")
	}
	manifest := bedrockManifest{
		FormatVersion: bedrockVersionString(raw.FormatVersion), Name: resolveBedrockText(raw.Header.Name, translations),
		Description: resolveBedrockText(raw.Header.Description, translations), UUID: raw.Header.UUID,
		Version: bedrockVersionString(raw.Header.Version), MinEngine: bedrockVersionString(raw.Header.MinEngineVersion),
		Capabilities: uniqueStringsPreserve(raw.Capabilities), Authors: uniqueStringsPreserve(raw.Metadata.Authors),
		License: raw.Metadata.License, URL: raw.Metadata.URL, MetadataBy: "manifest.json",
	}
	for _, module := range raw.Modules {
		moduleType := strings.ToLower(strings.TrimSpace(module.Type))
		if moduleType == "" {
			continue
		}
		label := moduleType
		if module.Language != "" {
			label += ":" + module.Language
		}
		manifest.Modules = append(manifest.Modules, label)
	}
	manifest.Modules = uniqueStringsPreserve(manifest.Modules)
	manifest.Kind = bedrockKindFromModules(manifest.Modules)
	for _, dependency := range raw.Dependencies {
		id := firstNonEmpty(dependency.UUID, dependency.ModuleName)
		if id == "" {
			continue
		}
		kind := "pack"
		if dependency.ModuleName != "" {
			kind = "module"
		}
		manifest.Dependencies = append(manifest.Dependencies, LibraryDependency{ID: strings.ToLower(id), Version: bedrockVersionString(dependency.Version), Kind: kind, Required: true})
	}
	return manifest, nil
}

func bedrockVersionString(value any) string {
	switch x := value.(type) {
	case string:
		return strings.TrimSpace(x)
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case []any:
		parts := make([]string, 0, len(x))
		for _, part := range x {
			parts = append(parts, bedrockVersionString(part))
		}
		return strings.Join(parts, ".")
	case []int:
		parts := make([]string, len(x))
		for i, part := range x {
			parts[i] = strconv.Itoa(part)
		}
		return strings.Join(parts, ".")
	}
	return ""
}

func bedrockKindFromModules(modules []string) string {
	for _, module := range modules {
		base := strings.SplitN(strings.ToLower(module), ":", 2)[0]
		switch base {
		case "data":
			return "behaviorpack"
		case "resources":
			return "resourcepack"
		case "skin_pack":
			return "skinpack"
		case "world_template":
			return "template"
		}
	}
	return "addon"
}

func isBedrockPackageExtension(name string) bool {
	name = strings.ToLower(strings.TrimSpace(name))
	for _, ext := range []string{".mcpack", ".mcaddon", ".mcworld", ".mctemplate"} {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}
	return false
}

func hasScriptModule(modules []string) bool {
	for _, module := range modules {
		if strings.HasPrefix(strings.ToLower(module), "script") {
			return true
		}
	}
	return false
}

func loadBedrockTranslations(root string) map[string]string {
	translations := map[string]string{}
	for _, name := range []string{"en_US.lang", "en_GB.lang"} {
		data, err := os.ReadFile(filepath.Join(root, "texts", name))
		if err != nil {
			continue
		}
		for _, line := range strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if i := strings.Index(line, "="); i > 0 {
				key := strings.TrimSpace(line[:i])
				value := strings.TrimSpace(strings.SplitN(line[i+1:], "\t", 2)[0])
				if key != "" && value != "" {
					translations[key] = value
				}
			}
		}
		if len(translations) > 0 {
			break
		}
	}
	return translations
}

func resolveBedrockText(value string, translations map[string]string) string {
	value = strings.TrimSpace(value)
	if translated := strings.TrimSpace(translations[value]); translated != "" {
		return translated
	}
	return value
}

func (a *App) inspectBedrockArchivePackage(path, profile string) (LibraryItem, error) {
	info, err := os.Stat(path)
	if err != nil {
		return LibraryItem{}, err
	}
	if err := validateBedrockArchive(path); err != nil {
		return LibraryItem{}, err
	}
	sha1Value, sha256Value, sha512Value, size, err := hashFileSet(path)
	if err != nil {
		return LibraryItem{}, err
	}
	manifests, err := readBedrockManifestsFromArchive(path)
	if err != nil {
		return LibraryItem{}, err
	}
	ext := strings.ToLower(filepath.Ext(path))
	kind := map[string]string{".mcpack": "addon-package", ".mcaddon": "addon-package", ".mcworld": "world-package", ".mctemplate": "template-package"}[ext]
	item := LibraryItem{
		ID: "sha512:" + sha512Value, Path: path, Filename: filepath.Base(path), Name: humanizeMinecraftFilename(filepath.Base(path)),
		Edition: "bedrock", Kind: firstNonEmpty(kind, "download"), Profile: firstNonEmpty(profile, "Vault downloads"), Enabled: true,
		IsDir: false, Size: size, Modified: info.ModTime().UTC().Format(time.RFC3339),
		Hashes: LibraryHashes{SHA1: sha1Value, SHA256: sha256Value, SHA512: sha512Value}, ManagedRoot: filepath.Dir(path),
		UpdateStatus: "unknown", UpdateMessage: "Bedrock package is preserved in Vault downloads and can be installed into Stable, Preview, or a custom profile.",
	}
	if len(manifests) > 0 {
		first := manifests[0]
		item.Name = firstNonEmpty(first.Name, item.Name)
		item.Summary = first.Description
		item.Authors = first.Authors
		item.InstalledVersion = first.Version
		item.UUID = strings.ToLower(first.UUID)
		item.MinEngineVersion = first.MinEngine
		item.Modules = first.Modules
		item.Dependencies = first.Dependencies
		item.Capabilities = first.Capabilities
		item.License = first.License
		item.Homepage = first.URL
		item.MetadataBy = "embedded manifest.json"
		item.ProvenanceConfidence = .9
		item.MatchEvidence = []string{"Embedded Bedrock manifest UUID and version"}
		if len(manifests) > 1 {
			item.Name += fmt.Sprintf(" + %d linked packs", len(manifests)-1)
			item.Modules = uniqueStringsPreserve(item.Modules)
			for _, manifest := range manifests[1:] {
				item.Modules = uniqueStringsPreserve(append(item.Modules, manifest.Modules...))
				item.Dependencies = append(item.Dependencies, manifest.Dependencies...)
			}
		}
	}
	if artURL, artErr := a.cacheBedrockArchiveArt(item.ID, path); artErr == nil {
		item.LocalArtURL, item.ArtOrigin = artURL, "embedded-local"
	}
	return item, nil
}

func readBedrockManifestsFromArchive(path string) ([]bedrockManifest, error) {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	translationsByDir := map[string]map[string]string{}
	for _, entry := range reader.File {
		name := strings.TrimPrefix(filepath.ToSlash(entry.Name), "./")
		low := strings.ToLower(name)
		textIndex := strings.Index(low, "texts/")
		if textIndex < 0 || (textIndex > 0 && low[textIndex-1] != '/') || (!strings.HasSuffix(low, "/en_us.lang") && !strings.HasSuffix(low, "/en_gb.lang")) {
			continue
		}
		data, readErr := readZipFile(entry, 4<<20)
		if readErr != nil {
			continue
		}
		root := strings.Trim(strings.TrimSuffix(name[:textIndex], "/"), "/")
		translations := map[string]string{}
		for _, line := range strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n") {
			if i := strings.Index(line, "="); i > 0 {
				translations[strings.TrimSpace(line[:i])] = strings.TrimSpace(strings.SplitN(line[i+1:], "\t", 2)[0])
			}
		}
		translationsByDir[strings.ToLower(root)] = translations
	}
	manifests := []bedrockManifest{}
	for _, entry := range reader.File {
		name := strings.TrimPrefix(filepath.ToSlash(entry.Name), "./")
		if !strings.EqualFold(filepath.Base(name), "manifest.json") {
			continue
		}
		data, readErr := readZipFile(entry, 4<<20)
		if readErr != nil {
			continue
		}
		root := filepath.ToSlash(filepath.Dir(name))
		if root == "." {
			root = ""
		}
		root = strings.ToLower(strings.Trim(root, "/"))
		manifest, parseErr := parseBedrockManifest(data, translationsByDir[root])
		if parseErr == nil {
			manifests = append(manifests, manifest)
		}
	}
	if len(manifests) == 0 {
		return nil, errors.New("Bedrock package did not contain a valid manifest.json")
	}
	return manifests, nil
}

func (a *App) cacheBedrockArchiveArt(id, path string) (string, error) {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return "", err
	}
	defer reader.Close()
	for _, entry := range reader.File {
		base := strings.ToLower(filepath.Base(entry.Name))
		if base != "pack_icon.png" && base != "world_icon.jpeg" && base != "world_icon.png" {
			continue
		}
		data, readErr := readZipFile(entry, 12<<20)
		if readErr != nil {
			continue
		}
		return a.cacheLocalArtBytes(id, filepath.Ext(entry.Name), data)
	}
	return "", os.ErrNotExist
}

func validateBedrockArchive(path string) error {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return fmt.Errorf("open Bedrock package: %w", err)
	}
	defer reader.Close()
	if len(reader.File) == 0 || len(reader.File) > 100000 {
		return errors.New("Bedrock package has an invalid file count")
	}
	seen := map[string]bool{}
	var total uint64
	for _, entry := range reader.File {
		if entry.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("Bedrock package contains a symbolic link: %s", entry.Name)
		}
		clean := filepath.Clean(filepath.FromSlash(entry.Name))
		if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
			return fmt.Errorf("Bedrock package contains an unsafe path: %s", entry.Name)
		}
		key := strings.ToLower(filepath.ToSlash(clean))
		if seen[key] {
			return fmt.Errorf("Bedrock package contains a duplicate path: %s", entry.Name)
		}
		seen[key] = true
		total += entry.UncompressedSize64
		if total > uint64(8<<30) {
			return errors.New("Bedrock package expands beyond the 8 GiB safety limit")
		}
		if entry.UncompressedSize64 > 16<<20 && entry.CompressedSize64 > 0 && entry.UncompressedSize64/entry.CompressedSize64 > 5000 {
			return fmt.Errorf("Bedrock package contains a suspicious compression ratio: %s", entry.Name)
		}
	}
	return nil
}

func (a *App) handleBedrockInstall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 2<<30)
	if err := r.ParseMultipartForm(64 << 20); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	profileID := strings.TrimSpace(r.FormValue("profile"))
	profile, ok := a.resolveBedrockProfile(profileID)
	if !ok {
		writeJSON(w, http.StatusBadRequest, APIError{Error: "Bedrock profile was not found; configure Stable, Preview, or a custom com.mojang root"})
		return
	}
	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		writeJSON(w, http.StatusBadRequest, APIError{Error: "no Bedrock packages were provided"})
		return
	}
	results := []any{}
	for _, header := range files {
		result, err := a.installUploadedBedrockPackage(r.Context(), header, profile)
		if err != nil {
			results = append(results, map[string]any{"ok": false, "package": header.Filename, "error": err.Error()})
		} else {
			results = append(results, result)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"results": results})
}

func (a *App) resolveBedrockProfile(id string) (LibraryProfile, bool) {
	profiles := a.libraryProfiles()
	for _, profile := range profiles {
		if profile.Edition != "bedrock" {
			continue
		}
		if id == "" || strings.EqualFold(profile.ID, id) || strings.EqualFold(profile.Name, id) {
			return profile, true
		}
	}
	return LibraryProfile{}, false
}

func (a *App) installUploadedBedrockPackage(ctx context.Context, header *multipart.FileHeader, profile LibraryProfile) (BedrockInstallResult, error) {
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".mcpack" && ext != ".mcaddon" && ext != ".mcworld" && ext != ".mctemplate" {
		return BedrockInstallResult{}, fmt.Errorf("unsupported Bedrock package extension %s", ext)
	}
	input, err := header.Open()
	if err != nil {
		return BedrockInstallResult{}, err
	}
	defer input.Close()
	stageDir := filepath.Join(a.cfgDir, "staging", "bedrock")
	if err := os.MkdirAll(stageDir, 0o755); err != nil {
		return BedrockInstallResult{}, err
	}
	stage := uniquePath(filepath.Join(stageDir, safeFilename(header.Filename)))
	out, err := os.OpenFile(stage, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o644)
	if err != nil {
		return BedrockInstallResult{}, err
	}
	_, copyErr := io.Copy(out, input)
	closeErr := out.Close()
	if copyErr != nil {
		_ = os.Remove(stage)
		return BedrockInstallResult{}, copyErr
	}
	if closeErr != nil {
		_ = os.Remove(stage)
		return BedrockInstallResult{}, closeErr
	}
	defer os.Remove(stage)
	if err := validateBedrockArchive(stage); err != nil {
		return BedrockInstallResult{}, err
	}
	return a.installBedrockPackage(ctx, stage, header.Filename, profile)
}

func (a *App) installBedrockPackage(ctx context.Context, packagePath, packageName string, profile LibraryProfile) (BedrockInstallResult, error) {
	select {
	case <-ctx.Done():
		return BedrockInstallResult{}, ctx.Err()
	default:
	}
	digest, _, err := hashFileSHA256(packagePath)
	if err != nil {
		return BedrockInstallResult{}, err
	}
	originalDir := filepath.Join(a.cfgDir, "library-originals", "bedrock")
	if err := os.MkdirAll(originalDir, 0o755); err != nil {
		return BedrockInstallResult{}, err
	}
	original := filepath.Join(originalDir, digest+"-"+safeFilename(packageName))
	if !pathExists(original) {
		if err := copyFileExclusive(packagePath, original); err != nil {
			return BedrockInstallResult{}, fmt.Errorf("preserve original package: %w", err)
		}
	}
	stagingRoot := filepath.Join(a.cfgDir, "staging")
	if err := os.MkdirAll(stagingRoot, 0o755); err != nil {
		return BedrockInstallResult{}, err
	}
	stageRoot, err := os.MkdirTemp(stagingRoot, "bedrock-unpack-")
	if err != nil {
		return BedrockInstallResult{}, err
	}
	defer os.RemoveAll(stageRoot)
	unpacked := filepath.Join(stageRoot, "unpacked")
	if err := os.MkdirAll(unpacked, 0o755); err != nil {
		return BedrockInstallResult{}, err
	}
	if err := extractZipPathSafe(packagePath, unpacked); err != nil {
		return BedrockInstallResult{}, err
	}
	ext := strings.ToLower(filepath.Ext(packageName))
	installed := []string{}
	kinds := []string{}
	if ext == ".mcworld" {
		worldRoot, err := findMinecraftWorldRoot(unpacked)
		if err != nil {
			return BedrockInstallResult{}, err
		}
		name := bedrockWorldName(worldRoot, strings.TrimSuffix(packageName, ext))
		dst := uniqueDirectoryPath(filepath.Join(profile.Root, "minecraftWorlds", safeWorldDirectoryName(name)))
		if err := copyDir(worldRoot, dst); err != nil {
			return BedrockInstallResult{}, err
		}
		installed, kinds = append(installed, dst), append(kinds, "world")
	} else {
		if err := expandNestedBedrockPacks(unpacked); err != nil {
			return BedrockInstallResult{}, err
		}
		roots, err := findBedrockManifestRoots(unpacked)
		if err != nil {
			return BedrockInstallResult{}, err
		}
		for _, root := range roots {
			data, readErr := os.ReadFile(filepath.Join(root, "manifest.json"))
			if readErr != nil {
				continue
			}
			manifest, parseErr := parseBedrockManifest(data, loadBedrockTranslations(root))
			if parseErr != nil {
				continue
			}
			kind := manifest.Kind
			dir := bedrockInstallDirectory(profile.Root, kind)
			if dir == "" {
				continue
			}
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return BedrockInstallResult{}, err
			}
			base := safeWorldDirectoryName(firstNonEmpty(manifest.Name, filepath.Base(root)))
			if manifest.UUID != "" {
				base += "-" + strings.ReplaceAll(shortUUID(manifest.UUID), "-", "")
			}
			dst := uniqueDirectoryPath(filepath.Join(dir, base))
			if err := copyDir(root, dst); err != nil {
				return BedrockInstallResult{}, err
			}
			installed, kinds = append(installed, dst), append(kinds, kind)
		}
		if len(installed) == 0 {
			return BedrockInstallResult{}, errors.New("package contained no installable Bedrock manifests")
		}
	}
	receipt := LibraryTransaction{
		SchemaVersion: librarySchemaVersion, ID: randomToken(12), Action: "bedrock-install", CreatedAt: time.Now().UTC().Format(time.RFC3339),
		SourcePaths: []string{original}, TargetPaths: installed, ItemNames: []string{packageName}, SHA512: []string{digest},
		Metadata: map[string]any{"profileId": profile.ID, "profileRoot": profile.Root, "kinds": kinds, "originalSha256": digest},
	}
	if err := a.writeLibraryTransaction(receipt); err != nil {
		for _, path := range installed {
			_ = os.RemoveAll(path)
		}
		return BedrockInstallResult{}, err
	}
	return BedrockInstallResult{OK: true, Profile: profile.Name, Package: packageName, InstalledPaths: installed, ReceiptID: receipt.ID, Kinds: kinds}, nil
}

func copyFileExclusive(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		_ = os.Remove(dst)
		return copyErr
	}
	if closeErr != nil {
		_ = os.Remove(dst)
		return closeErr
	}
	return nil
}

func expandNestedBedrockPacks(root string) error {
	var nested []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext == ".mcpack" || ext == ".mcaddon" {
			nested = append(nested, path)
		}
		return nil
	})
	if err != nil {
		return err
	}
	if len(nested) > 64 {
		return errors.New("Bedrock package contains too many nested pack archives")
	}
	for i, archive := range nested {
		if err := validateBedrockArchive(archive); err != nil {
			return err
		}
		dst := filepath.Join(root, fmt.Sprintf(".mmv-nested-%02d", i+1))
		if err := os.MkdirAll(dst, 0o755); err != nil {
			return err
		}
		if err := extractZipPathSafe(archive, dst); err != nil {
			return err
		}
	}
	return nil
}

func findBedrockManifestRoots(root string) ([]string, error) {
	roots := []string{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.EqualFold(entry.Name(), "manifest.json") {
			return nil
		}
		roots = append(roots, filepath.Dir(path))
		if len(roots) > 64 {
			return errors.New("Bedrock package contains more than 64 manifest roots")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(roots) == 0 {
		return nil, errors.New("Bedrock package did not contain manifest.json")
	}
	sort.Slice(roots, func(i, j int) bool { return pathDepth(roots[i]) < pathDepth(roots[j]) })
	return uniqueStringsPreserve(roots), nil
}

func pathDepth(path string) int {
	return len(strings.Split(filepath.Clean(path), string(os.PathSeparator)))
}

func bedrockInstallDirectory(root, kind string) string {
	switch kind {
	case "behaviorpack":
		return filepath.Join(root, "behavior_packs")
	case "resourcepack":
		return filepath.Join(root, "resource_packs")
	case "skinpack":
		return filepath.Join(root, "skin_packs")
	case "template":
		return filepath.Join(root, "world_templates")
	default:
		return ""
	}
}

func bedrockWorldName(root, fallback string) string {
	if data, err := os.ReadFile(filepath.Join(root, "levelname.txt")); err == nil {
		if value := strings.TrimSpace(strings.TrimPrefix(string(data), "\ufeff")); value != "" {
			return value
		}
	}
	return fallback
}

func shortUUID(value string) string {
	value = strings.TrimSpace(value)
	if len(value) > 8 {
		return value[:8]
	}
	return value
}

func (a *App) handleBedrockActivate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request BedrockActivationRequest
	if err := decodeJSON(r, &request); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	receipt, err := a.activateBedrockPack(request)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "receiptId": receipt.ID, "worldPath": request.WorldPath, "packUuid": request.PackUUID})
}

func (a *App) activateBedrockPack(request BedrockActivationRequest) (LibraryTransaction, error) {
	world := filepath.Clean(request.WorldPath)
	if !a.allowedBedrockWorldPath(world) {
		return LibraryTransaction{}, errors.New("world path is outside configured Bedrock profiles")
	}
	if strings.TrimSpace(request.PackUUID) == "" {
		return LibraryTransaction{}, errors.New("pack UUID is required")
	}
	fileName := "world_behavior_packs.json"
	if strings.Contains(strings.ToLower(request.PackKind), "resource") {
		fileName = "world_resource_packs.json"
	}
	path := filepath.Join(world, fileName)
	before, _ := os.ReadFile(path)
	var entries []map[string]any
	if len(before) > 0 && json.Unmarshal(before, &entries) != nil {
		return LibraryTransaction{}, fmt.Errorf("%s is not valid JSON", fileName)
	}
	versionParts := []int{1, 0, 0}
	if parsed := parseBedrockVersionInts(request.Version); len(parsed) > 0 {
		versionParts = parsed
	}
	found := false
	for _, entry := range entries {
		if strings.EqualFold(stringFromAny(entry["pack_id"]), request.PackUUID) {
			entry["version"] = versionParts
			found = true
		}
	}
	if !found {
		entries = append(entries, map[string]any{"pack_id": request.PackUUID, "version": versionParts})
	}
	after, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return LibraryTransaction{}, err
	}
	after = append(after, '\n')
	if err := os.WriteFile(path+".tmp", after, 0o644); err != nil {
		return LibraryTransaction{}, err
	}
	if err := os.Rename(path+".tmp", path); err != nil {
		_ = os.Remove(path + ".tmp")
		return LibraryTransaction{}, err
	}
	receipt := LibraryTransaction{
		SchemaVersion: librarySchemaVersion, ID: randomToken(12), Action: "bedrock-activate", CreatedAt: time.Now().UTC().Format(time.RFC3339),
		SourcePaths: []string{path}, TargetPaths: []string{path}, ItemNames: []string{request.PackUUID},
		Metadata: map[string]any{"worldPath": world, "packUuid": request.PackUUID, "packKind": request.PackKind, "beforeBase64": base64.StdEncoding.EncodeToString(before), "afterSha256": sha256Text(after)},
	}
	if err := a.writeLibraryTransaction(receipt); err != nil {
		_ = os.WriteFile(path, before, 0o644)
		return LibraryTransaction{}, err
	}
	return receipt, nil
}

func (a *App) allowedBedrockWorldPath(path string) bool {
	abs, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	for _, profile := range a.libraryProfiles() {
		if profile.Edition != "bedrock" {
			continue
		}
		root, _ := filepath.Abs(filepath.Join(profile.Root, "minecraftWorlds"))
		rel, relErr := filepath.Rel(root, abs)
		if relErr == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
			return true
		}
	}
	return false
}

func parseBedrockVersionInts(value string) []int {
	parts := strings.Split(strings.TrimSpace(value), ".")
	out := []int{}
	for _, part := range parts {
		if part == "" {
			continue
		}
		n, err := strconv.Atoi(part)
		if err != nil {
			return nil
		}
		out = append(out, n)
	}
	return out
}

func sha256Text(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}
