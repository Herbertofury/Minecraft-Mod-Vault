package main

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	libraryCacheTTL    = 7 * 24 * time.Hour
	libraryScanWorkers = 6
)

func (a *App) handleLibrary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	enrich := queryBool(r.URL.Query(), "enrich", true)
	refresh := queryBool(r.URL.Query(), "refresh", false)
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()
	resp, err := a.scanOmniLibrary(ctx, enrich, refresh)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func queryBool(values url.Values, key string, fallback bool) bool {
	value := strings.ToLower(strings.TrimSpace(values.Get(key)))
	if value == "" {
		return fallback
	}
	return value == "1" || value == "true" || value == "yes" || value == "on"
}

func (a *App) scanOmniLibrary(ctx context.Context, enrich, refresh bool) (LibraryResponse, error) {
	profiles := a.libraryProfiles()
	items, warnings, errorsByRoot := a.scanLocalLibrary(ctx, profiles)
	cache := a.loadLibraryCache()
	oldest := time.Time{}
	for i := range items {
		if record, ok := cache.Records[items[i].ID]; ok && !refresh {
			applyLibraryCacheRecord(&items[i], record)
			if t, err := time.Parse(time.RFC3339, record.UpdatedAt); err == nil && (oldest.IsZero() || t.Before(oldest)) {
				oldest = t
			}
		}
	}
	if enrich {
		if err := a.enrichLibrary(ctx, items, refresh, &cache); err != nil {
			warnings = append(warnings, "Provider enrichment was partially unavailable: "+err.Error())
		}
	}
	linkBedrockFamilies(items)
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Edition != items[j].Edition {
			return items[i].Edition < items[j].Edition
		}
		if items[i].Kind != items[j].Kind {
			return items[i].Kind < items[j].Kind
		}
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})
	resp := newLibraryResponse(items, profiles, enrich)
	resp.Warnings = uniqueStringsPreserve(warnings)
	resp.Errors = errorsByRoot
	if !oldest.IsZero() {
		resp.CacheAge = time.Since(oldest).Round(time.Second).String()
	}
	return resp, nil
}

func (a *App) libraryProfiles() []LibraryProfile {
	a.mu.RLock()
	settings := a.settings
	a.mu.RUnlock()
	profiles := []LibraryProfile{{ID: "java:" + profileSlug(settings.ActiveProfile), Name: firstNonEmpty(settings.ActiveProfile, "Default"), Edition: "java", Root: settings.JavaRoot, Exists: pathExists(settings.JavaRoot), Channel: "Java"}}
	seen := map[string]bool{}
	appendBedrock := func(id, name, channel, root string) {
		root = strings.TrimSpace(root)
		if root == "" {
			return
		}
		abs, _ := filepath.Abs(filepath.Clean(root))
		key := strings.ToLower(abs)
		if seen[key] {
			return
		}
		seen[key] = true
		profiles = append(profiles, LibraryProfile{ID: id, Name: name, Edition: "bedrock", Root: abs, Exists: pathExists(abs), Channel: channel})
	}
	appendBedrock("bedrock:stable", "Bedrock Stable", "Stable", settings.BedrockRoot)
	appendBedrock("bedrock:preview", "Bedrock Preview", "Preview / Beta", settings.BedrockPreviewRoot)
	for i, root := range settings.BedrockCustomRoots {
		appendBedrock(fmt.Sprintf("bedrock:custom-%d", i+1), fmt.Sprintf("Bedrock Custom %d", i+1), "Custom", root)
	}
	return profiles
}

func profileSlug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	if value == "" {
		return "default"
	}
	return value
}

func (a *App) scanLocalLibrary(ctx context.Context, profiles []LibraryProfile) ([]LibraryItem, []string, map[string]string) {
	type target struct {
		Edition string
		Kind    string
		Profile string
		Root    string
	}
	targets := []target{}
	a.mu.RLock()
	settings := a.settings
	a.mu.RUnlock()
	javaProfile := firstNonEmpty(settings.ActiveProfile, "Default")
	for _, pair := range []struct{ kind, root string }{
		{"mod", filepath.Join(settings.JavaRoot, "mods")},
		{"resourcepack", filepath.Join(settings.JavaRoot, "resourcepacks")},
		{"shader", filepath.Join(settings.JavaRoot, "shaderpacks")},
		{"world", filepath.Join(settings.JavaRoot, "saves")},
		{"plugin", a.javaTargetDir("plugins")},
		{"datapack", a.javaTargetDir("datapacks")},
		{"download", a.javaTargetDir("downloads")},
	} {
		edition := "java"
		if pair.kind == "plugin" {
			edition = "server"
		}
		targets = append(targets, target{Edition: edition, Kind: pair.kind, Profile: javaProfile, Root: pair.root})
	}
	for _, profile := range profiles {
		if profile.Edition != "bedrock" || profile.Root == "" {
			continue
		}
		for _, pair := range []struct{ kind, dir string }{
			{"behaviorpack", "behavior_packs"}, {"resourcepack-bedrock", "resource_packs"}, {"skinpack", "skin_packs"},
			{"world-bedrock", "minecraftWorlds"}, {"template", "world_templates"},
			{"behaviorpack-dev", "development_behavior_packs"}, {"resourcepack-bedrock-dev", "development_resource_packs"},
		} {
			targets = append(targets, target{Edition: "bedrock", Kind: pair.kind, Profile: profile.Name, Root: filepath.Join(profile.Root, pair.dir)})
		}
	}
	var mu sync.Mutex
	items := []LibraryItem{}
	warnings := []string{}
	errorsByRoot := map[string]string{}
	sem := make(chan struct{}, libraryScanWorkers)
	var wg sync.WaitGroup
	for _, current := range targets {
		current := current
		if current.Root == "" || !pathExists(current.Root) {
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			var found []LibraryItem
			var scanWarnings []string
			var err error
			if current.Edition == "bedrock" {
				found, scanWarnings, err = a.scanBedrockRoot(ctx, current.Root, current.Kind, current.Profile)
			} else {
				found, scanWarnings, err = a.scanJavaRoot(ctx, current.Root, current.Kind, current.Edition, current.Profile)
			}
			mu.Lock()
			defer mu.Unlock()
			items = append(items, found...)
			warnings = append(warnings, scanWarnings...)
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				errorsByRoot[current.Root] = err.Error()
			}
		}()
	}
	wg.Wait()
	items = append(items, a.scanDisabledLibraryItems()...)
	items = deduplicateLibraryItems(items)
	return items, warnings, errorsByRoot
}

func deduplicateLibraryItems(items []LibraryItem) []LibraryItem {
	out := make([]LibraryItem, 0, len(items))
	seen := map[string]int{}
	for _, item := range items {
		key := strings.ToLower(filepath.Clean(item.Path))
		if idx, ok := seen[key]; ok {
			mergeLibraryItem(&out[idx], item)
			continue
		}
		seen[key] = len(out)
		out = append(out, item)
	}
	return out
}

func (a *App) scanJavaRoot(ctx context.Context, root, kind, edition, profile string) ([]LibraryItem, []string, error) {
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
		item, itemWarnings, inspectErr := a.inspectJavaLibraryPath(path, kind, edition, profile)
		if inspectErr != nil {
			warnings = append(warnings, fmt.Sprintf("%s: %v", entry.Name(), inspectErr))
			continue
		}
		warnings = append(warnings, itemWarnings...)
		items = append(items, item)
	}
	return items, warnings, nil
}

func (a *App) inspectJavaLibraryPath(path, kind, edition, profile string) (LibraryItem, []string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return LibraryItem{}, nil, err
	}
	base := filepath.Base(path)
	item := LibraryItem{
		Path: path, Filename: base, Name: humanizeMinecraftFilename(base), Edition: edition, Kind: kind, Profile: profile,
		Enabled: !strings.HasSuffix(strings.ToLower(base), ".disabled"), IsDir: info.IsDir(), Size: info.Size(),
		Modified: info.ModTime().UTC().Format(time.RFC3339), ManagedRoot: filepath.Dir(path), UpdateStatus: "unknown",
		UpdateMessage: "Local metadata loaded; provider identity has not been resolved yet.",
	}
	warnings := []string{}
	if info.IsDir() {
		digest, _, bytesCount, hashErr := hashDirectoryTree(path)
		if hashErr != nil {
			return LibraryItem{}, nil, hashErr
		}
		item.ID = "dir-sha256:" + digest
		item.Hashes.SHA256 = digest
		item.Size = bytesCount
		if kind == "resourcepack" || kind == "datapack" {
			applyPackMetadataDirectory(&item, path)
		}
		if icon := firstExistingPath(filepath.Join(path, "pack.png"), filepath.Join(path, "icon.png")); icon != "" {
			if artURL, artErr := a.cacheLocalArtFile(item.ID, icon); artErr == nil {
				item.LocalArtURL, item.ArtOrigin = artURL, "embedded-local"
			}
		}
		return item, warnings, nil
	}
	low := strings.ToLower(base)
	if kind == "download" && isBedrockPackageExtension(low) {
		bedrockItem, bedrockErr := a.inspectBedrockArchivePackage(path, "Vault downloads")
		if bedrockErr != nil {
			return LibraryItem{}, nil, bedrockErr
		}
		return bedrockItem, warnings, nil
	}
	if strings.HasSuffix(low, ".jar") || strings.HasSuffix(low, ".jar.disabled") {
		local, inspectErr := inspectLocalJar(path)
		if inspectErr != nil {
			return LibraryItem{}, nil, inspectErr
		}
		item.ID = "sha512:" + local.SHA512
		item.Hashes = LibraryHashes{SHA1: local.SHA1, SHA512: local.SHA512, CurseFingerprint: local.CurseFingerprint}
		item.Name = firstNonEmpty(local.Metadata.Name, item.Name)
		item.Summary = local.Metadata.Description
		item.Description = local.Metadata.Description
		item.Authors = local.Metadata.Authors
		item.InstalledVersion = local.Metadata.Version
		item.ModID = local.Metadata.ModID
		item.Loaders = uniqueStrings(local.Metadata.Loaders)
		if local.Metadata.Minecraft != "" {
			item.GameVersions = []string{local.Metadata.Minecraft}
		}
		item.Dependencies = stringDependencies(local.Metadata.Dependencies)
		item.License = local.Metadata.License
		item.Homepage = local.Metadata.Homepage
		item.SourceURL = local.Metadata.SourceURL
		item.MetadataBy = local.Metadata.MetadataBy
		item.MatchEvidence = append(item.MatchEvidence, "Embedded "+local.Metadata.MetadataBy+" metadata")
		if local.Metadata.Name != "" {
			item.ProvenanceConfidence = .82
		}
		if local.Metadata.IconPath != "" {
			if artURL, artErr := a.cacheArchiveArt(item.ID, path, local.Metadata.IconPath); artErr == nil {
				item.LocalArtURL, item.ArtOrigin = artURL, "embedded-local"
			} else {
				warnings = append(warnings, base+": declared icon could not be loaded: "+artErr.Error())
			}
		}
		if isLikelyModifiedBuild(base, local.Metadata.Version) {
			item.UpdateStatus = "modified"
			item.UpdateMessage = "Custom, patched, or repacked build detected. Vault will not replace it automatically without exact artifact evidence."
			item.Warnings = append(item.Warnings, "Modified builds are protected from blind provider replacement.")
		}
		return item, warnings, nil
	}
	sha1Value, sha256Value, sha512Value, size, hashErr := hashFileSet(path)
	if hashErr != nil {
		return LibraryItem{}, nil, hashErr
	}
	item.ID = "sha512:" + sha512Value
	item.Hashes = LibraryHashes{SHA1: sha1Value, SHA256: sha256Value, SHA512: sha512Value}
	item.Size = size
	if strings.HasSuffix(low, ".zip") || strings.HasSuffix(low, ".mrpack") {
		applyPackMetadataArchive(&item, path)
		if artURL, artErr := a.cacheFirstArchiveArt(item.ID, path, []string{"pack.png", "icon.png", "server-icon.png", "world_icon.jpeg", "world_icon.png"}); artErr == nil {
			item.LocalArtURL, item.ArtOrigin = artURL, "embedded-local"
		}
	}
	return item, warnings, nil
}

func hashFileSet(path string) (string, string, string, int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", "", "", 0, err
	}
	defer file.Close()
	h1, h256, h512 := sha1.New(), sha256.New(), sha512.New()
	n, err := io.Copy(io.MultiWriter(h1, h256, h512), file)
	if err != nil {
		return "", "", "", n, err
	}
	return hex.EncodeToString(h1.Sum(nil)), hex.EncodeToString(h256.Sum(nil)), hex.EncodeToString(h512.Sum(nil)), n, nil
}

func stringDependencies(ids []string) []LibraryDependency {
	out := make([]LibraryDependency, 0, len(ids))
	for _, id := range ids {
		if id = strings.TrimSpace(id); id != "" {
			out = append(out, LibraryDependency{ID: id, Required: true})
		}
	}
	return out
}

func humanizeMinecraftFilename(name string) string {
	name = strings.TrimSuffix(name, ".disabled")
	ext := filepath.Ext(name)
	name = strings.TrimSuffix(name, ext)
	name = regexp.MustCompile(`(?i)(?:[-_.](?:forge|fabric|neoforge|quilt|mc)?\s*v?\d+(?:\.\d+){1,3}.*)$`).ReplaceAllString(name, "")
	name = strings.NewReplacer("_", " ", "-", " ", ".", " ").Replace(name)
	name = strings.Join(strings.Fields(name), " ")
	if name == "" {
		return filepath.Base(strings.TrimSuffix(name, ext))
	}
	words := strings.Fields(name)
	for i, word := range words {
		if len(word) > 1 && word == strings.ToLower(word) {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, " ")
}

func isLikelyModifiedBuild(filename, declaredVersion string) bool {
	low := strings.ToLower(filename)
	for _, marker := range []string{"patched", "fixed", "hotfix", "compat", "custom", "repack", "unofficial", "fork"} {
		if strings.Contains(low, marker) {
			return true
		}
	}
	return declaredVersion != "" && strings.Contains(declaredVersion, "+local")
}

func applyPackMetadataArchive(item *LibraryItem, archivePath string) {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return
	}
	defer reader.Close()
	for _, entry := range reader.File {
		if strings.EqualFold(strings.TrimPrefix(filepath.ToSlash(entry.Name), "./"), "pack.mcmeta") {
			if data, readErr := readZipFile(entry, 2<<20); readErr == nil {
				applyPackMetadataBytes(item, data)
			}
			return
		}
	}
}

func applyPackMetadataDirectory(item *LibraryItem, root string) {
	if data, err := os.ReadFile(filepath.Join(root, "pack.mcmeta")); err == nil {
		applyPackMetadataBytes(item, data)
	}
}

func applyPackMetadataBytes(item *LibraryItem, data []byte) {
	var value struct {
		Pack struct {
			PackFormat  int `json:"pack_format"`
			Description any `json:"description"`
		} `json:"pack"`
	}
	if json.Unmarshal(data, &value) != nil {
		return
	}
	item.Summary = minecraftTextComponent(value.Pack.Description)
	item.Description = item.Summary
	item.MetadataBy = "pack.mcmeta"
	if value.Pack.PackFormat > 0 {
		item.GameVersions = append(item.GameVersions, fmt.Sprintf("pack format %d", value.Pack.PackFormat))
	}
	item.MatchEvidence = append(item.MatchEvidence, "Embedded pack.mcmeta")
	item.ProvenanceConfidence = maxFloat(item.ProvenanceConfidence, .7)
}

func minecraftTextComponent(value any) string {
	switch x := value.(type) {
	case string:
		return strings.TrimSpace(x)
	case map[string]any:
		parts := []string{}
		if text := flexibleString(x["text"]); text != "" {
			parts = append(parts, text)
		}
		if translate := flexibleString(x["translate"]); translate != "" {
			parts = append(parts, translate)
		}
		if extra, ok := x["extra"].([]any); ok {
			for _, part := range extra {
				if text := minecraftTextComponent(part); text != "" {
					parts = append(parts, text)
				}
			}
		}
		return strings.Join(parts, "")
	case []any:
		parts := []string{}
		for _, part := range x {
			if text := minecraftTextComponent(part); text != "" {
				parts = append(parts, text)
			}
		}
		return strings.Join(parts, "")
	}
	return ""
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func firstExistingPath(paths ...string) string {
	for _, path := range paths {
		if info, err := os.Stat(path); err == nil && info.Mode().IsRegular() {
			return path
		}
	}
	return ""
}

func (a *App) libraryArtDir() string {
	return filepath.Join(a.cfgDir, "library-art")
}

func libraryArtKey(id string) string {
	h := sha256.Sum256([]byte(id))
	return hex.EncodeToString(h[:])
}

func (a *App) cacheLocalArtFile(id, source string) (string, error) {
	data, err := os.ReadFile(source)
	if err != nil {
		return "", err
	}
	return a.cacheLocalArtBytes(id, filepath.Ext(source), data)
}

func (a *App) cacheArchiveArt(id, archivePath, wanted string) (string, error) {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", err
	}
	defer reader.Close()
	wanted = strings.ToLower(strings.TrimPrefix(filepath.ToSlash(wanted), "./"))
	for _, entry := range reader.File {
		name := strings.ToLower(strings.TrimPrefix(filepath.ToSlash(entry.Name), "./"))
		if name != wanted {
			continue
		}
		data, err := readZipFile(entry, 12<<20)
		if err != nil {
			return "", err
		}
		return a.cacheLocalArtBytes(id, filepath.Ext(entry.Name), data)
	}
	return "", os.ErrNotExist
}

func (a *App) cacheFirstArchiveArt(id, archivePath string, candidates []string) (string, error) {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", err
	}
	defer reader.Close()
	wanted := map[string]bool{}
	for _, candidate := range candidates {
		wanted[strings.ToLower(strings.TrimPrefix(filepath.ToSlash(candidate), "./"))] = true
	}
	for _, entry := range reader.File {
		name := strings.ToLower(strings.TrimPrefix(filepath.ToSlash(entry.Name), "./"))
		if !wanted[name] {
			continue
		}
		data, err := readZipFile(entry, 12<<20)
		if err != nil {
			return "", err
		}
		return a.cacheLocalArtBytes(id, filepath.Ext(entry.Name), data)
	}
	return "", os.ErrNotExist
}

func (a *App) cacheLocalArtBytes(id, extension string, data []byte) (string, error) {
	if len(data) == 0 || len(data) > 12<<20 {
		return "", errors.New("art asset is empty or exceeds 12 MiB")
	}
	extension = strings.ToLower(extension)
	switch extension {
	case ".png", ".jpg", ".jpeg", ".webp", ".gif", ".svg":
	default:
		extension = ".png"
	}
	dir := a.libraryArtDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	key := libraryArtKey(id)
	for _, match := range libraryArtCandidates(dir, key) {
		if filepath.Ext(match) != extension {
			_ = os.Remove(match)
		}
	}
	path := filepath.Join(dir, key+extension)
	if current, err := os.ReadFile(path); err == nil && bytes.Equal(current, data) {
		return "/api/library/art?id=" + url.QueryEscape(id), nil
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return "", err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return "", err
	}
	return "/api/library/art?id=" + url.QueryEscape(id), nil
}

func libraryArtCandidates(dir, key string) []string {
	matches, _ := filepath.Glob(filepath.Join(dir, key+".*"))
	return matches
}

func (a *App) handleLibraryArt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	if id == "" {
		http.NotFound(w, r)
		return
	}
	matches := libraryArtCandidates(a.libraryArtDir(), libraryArtKey(id))
	if len(matches) == 0 {
		http.NotFound(w, r)
		return
	}
	path := matches[0]
	data, err := os.ReadFile(path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	contentType := mime.TypeByExtension(filepath.Ext(path))
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "private, max-age=31536000, immutable")
	w.Header().Set("ETag", `"`+libraryArtKey(id)+`"`)
	_, _ = w.Write(data)
}

func mergeLibraryItem(dst *LibraryItem, src LibraryItem) {
	if dst == nil {
		return
	}
	dst.Name = firstNonEmpty(dst.Name, src.Name)
	dst.Summary = firstNonEmpty(dst.Summary, src.Summary)
	dst.Description = firstNonEmpty(dst.Description, src.Description)
	dst.Authors = uniqueStringsPreserve(append(dst.Authors, src.Authors...))
	dst.Loaders = uniqueStrings(append(dst.Loaders, src.Loaders...))
	dst.GameVersions = uniqueStringsPreserve(append(dst.GameVersions, src.GameVersions...))
	dst.Dependencies = append(dst.Dependencies, src.Dependencies...)
	dst.Sources = mergeLibrarySources(dst.Sources, src.Sources)
	dst.MatchEvidence = uniqueStringsPreserve(append(dst.MatchEvidence, src.MatchEvidence...))
	dst.Warnings = uniqueStringsPreserve(append(dst.Warnings, src.Warnings...))
	if dst.LocalArtURL == "" {
		dst.LocalArtURL, dst.ArtOrigin = src.LocalArtURL, src.ArtOrigin
	}
	if dst.RemoteArtURL == "" {
		dst.RemoteArtURL = src.RemoteArtURL
	}
	dst.ProvenanceConfidence = maxFloat(dst.ProvenanceConfidence, src.ProvenanceConfidence)
}

func mergeLibrarySources(existing, added []LibrarySource) []LibrarySource {
	out := append([]LibrarySource(nil), existing...)
	index := map[string]int{}
	for i, source := range out {
		index[strings.ToLower(source.Provider+":"+source.ProjectID+":"+source.Slug+":"+source.PageURL)] = i
	}
	for _, source := range added {
		key := strings.ToLower(source.Provider + ":" + source.ProjectID + ":" + source.Slug + ":" + source.PageURL)
		if i, ok := index[key]; ok {
			if source.Confidence > out[i].Confidence {
				out[i] = source
			}
			continue
		}
		index[key] = len(out)
		out = append(out, source)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Exact != out[j].Exact {
			return out[i].Exact
		}
		return out[i].Confidence > out[j].Confidence
	})
	return out
}

func (a *App) loadLibraryCache() libraryCacheFile {
	cache := libraryCacheFile{SchemaVersion: librarySchemaVersion, Records: map[string]libraryCacheRecord{}}
	path := filepath.Join(a.cfgDir, "library-enrichment-cache.json")
	a.dataMu.RLock()
	data, err := os.ReadFile(path)
	a.dataMu.RUnlock()
	if err != nil || json.Unmarshal(data, &cache) != nil || cache.SchemaVersion != librarySchemaVersion {
		cache = libraryCacheFile{SchemaVersion: librarySchemaVersion, Records: map[string]libraryCacheRecord{}}
	}
	if cache.Records == nil {
		cache.Records = map[string]libraryCacheRecord{}
	}
	cutoff := time.Now().Add(-30 * 24 * time.Hour)
	for id, record := range cache.Records {
		when, err := time.Parse(time.RFC3339, record.UpdatedAt)
		if err != nil || when.Before(cutoff) {
			delete(cache.Records, id)
		}
	}
	return cache
}

func (a *App) saveLibraryCache(cache libraryCacheFile) error {
	cache.SchemaVersion = librarySchemaVersion
	if cache.Records == nil {
		cache.Records = map[string]libraryCacheRecord{}
	}
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(a.cfgDir, "library-enrichment-cache.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	a.dataMu.Lock()
	defer a.dataMu.Unlock()
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func applyLibraryCacheRecord(item *LibraryItem, record libraryCacheRecord) {
	if item == nil {
		return
	}
	if t, err := time.Parse(time.RFC3339, record.UpdatedAt); err != nil || time.Since(t) > libraryCacheTTL {
		return
	}
	if item.Name == "" || strings.EqualFold(item.Name, humanizeMinecraftFilename(item.Filename)) {
		item.Name = firstNonEmpty(record.Name, item.Name)
	}
	item.Summary = firstNonEmpty(item.Summary, record.Summary)
	item.Authors = uniqueStringsPreserve(append(item.Authors, record.Authors...))
	item.RemoteArtURL = firstNonEmpty(item.RemoteArtURL, record.RemoteArtURL)
	item.Sources = mergeLibrarySources(item.Sources, record.Sources)
	item.LatestVersion = firstNonEmpty(record.LatestVersion, item.LatestVersion)
	if item.UpdateStatus != "modified" {
		item.UpdateStatus = firstNonEmpty(record.UpdateStatus, item.UpdateStatus)
		item.UpdateMessage = firstNonEmpty(record.UpdateMessage, item.UpdateMessage)
		item.SafeUpdate = record.SafeUpdate
	}
	item.Alternatives = append(item.Alternatives, record.Alternatives...)
	item.ProvenanceConfidence = maxFloat(item.ProvenanceConfidence, record.ProvenanceConfidence)
	item.MatchEvidence = uniqueStringsPreserve(append(item.MatchEvidence, record.MatchEvidence...))
}

func cacheRecordFromItem(item LibraryItem) libraryCacheRecord {
	return libraryCacheRecord{
		ItemID: item.ID, UpdatedAt: time.Now().UTC().Format(time.RFC3339), Name: item.Name, Summary: item.Summary,
		Authors: item.Authors, RemoteArtURL: item.RemoteArtURL, Sources: item.Sources, LatestVersion: item.LatestVersion,
		UpdateStatus: item.UpdateStatus, UpdateMessage: item.UpdateMessage, SafeUpdate: item.SafeUpdate,
		Alternatives: item.Alternatives, ProvenanceConfidence: item.ProvenanceConfidence, MatchEvidence: item.MatchEvidence,
	}
}

func linkBedrockFamilies(items []LibraryItem) {
	byUUID := map[string]int{}
	for i := range items {
		if items[i].Edition == "bedrock" && items[i].UUID != "" {
			byUUID[strings.ToLower(items[i].UUID)] = i
		}
	}
	for i := range items {
		if items[i].Edition != "bedrock" {
			continue
		}
		for _, dep := range items[i].Dependencies {
			if other, ok := byUUID[strings.ToLower(dep.ID)]; ok {
				family := "bedrock-family:" + minString(strings.ToLower(items[i].UUID), strings.ToLower(items[other].UUID))
				items[i].FamilyID = family
				items[other].FamilyID = family
			}
		}
	}
}

func minString(a, b string) string {
	if a < b {
		return a
	}
	return b
}

var _ fs.FileMode
var _ = runtime.GOOS
