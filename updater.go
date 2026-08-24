package main

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type JarMetadata struct {
	ModID        string   `json:"modId,omitempty"`
	Name         string   `json:"name,omitempty"`
	Version      string   `json:"version,omitempty"`
	Loaders      []string `json:"loaders,omitempty"`
	SourceURL    string   `json:"sourceUrl,omitempty"`
	Homepage     string   `json:"homepage,omitempty"`
	Minecraft    string   `json:"minecraft,omitempty"`
	MetadataBy   string   `json:"metadataBy,omitempty"`
	Authors      []string `json:"authors,omitempty"`
	Description  string   `json:"description,omitempty"`
	IconPath     string   `json:"iconPath,omitempty"`
	License      string   `json:"license,omitempty"`
	Credits      string   `json:"credits,omitempty"`
	Environment  string   `json:"environment,omitempty"`
	Dependencies []string `json:"dependencies,omitempty"`
}

type LocalModFile struct {
	Path             string      `json:"path"`
	Filename         string      `json:"filename"`
	Enabled          bool        `json:"enabled"`
	Size             int64       `json:"size"`
	SHA1             string      `json:"sha1"`
	SHA512           string      `json:"sha512"`
	CurseFingerprint uint32      `json:"curseFingerprint"`
	Metadata         JarMetadata `json:"metadata"`
}

type UpdateCandidate struct {
	ID             string            `json:"id"`
	Provider       string            `json:"provider"`
	ProjectID      string            `json:"projectId,omitempty"`
	ProjectTitle   string            `json:"projectTitle"`
	VersionID      string            `json:"versionId,omitempty"`
	Version        string            `json:"version"`
	Filename       string            `json:"filename"`
	URL            string            `json:"url,omitempty"`
	PageURL        string            `json:"pageUrl,omitempty"`
	Hashes         map[string]string `json:"hashes,omitempty"`
	Size           int64             `json:"size,omitempty"`
	GameVersions   []string          `json:"gameVersions,omitempty"`
	Loaders        []string          `json:"loaders,omitempty"`
	Confidence     float64           `json:"confidence"`
	Reason         string            `json:"reason"`
	Safe           bool              `json:"safe"`
	DependencyRisk bool              `json:"dependencyRisk,omitempty"`
}

type UpdateItem struct {
	Local            LocalModFile      `json:"local"`
	Status           string            `json:"status"`
	InstalledProject string            `json:"installedProject,omitempty"`
	InstalledVersion string            `json:"installedVersion,omitempty"`
	SafeUpdate       *UpdateCandidate  `json:"safeUpdate,omitempty"`
	Alternatives     []UpdateCandidate `json:"alternatives,omitempty"`
	Message          string            `json:"message"`
}

type UpdatePlan struct {
	ID            string       `json:"id"`
	CreatedAt     string       `json:"createdAt"`
	GameVersion   string       `json:"gameVersion"`
	Loader        string       `json:"loader"`
	ModsDirectory string       `json:"modsDirectory"`
	Items         []UpdateItem `json:"items"`
	SafeCount     int          `json:"safeCount"`
	CurrentCount  int          `json:"currentCount"`
	ReviewCount   int          `json:"reviewCount"`
	UnknownCount  int          `json:"unknownCount"`
}

type updaterScanRequest struct {
	GameVersion string `json:"gameVersion"`
	Loader      string `json:"loader"`
}

type updaterApplyRequest struct {
	PlanID        string   `json:"planId"`
	ItemFilenames []string `json:"itemFilenames,omitempty"`
}

func (a *App) handleUpdaterScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var in updaterScanRequest
	if r.Body != nil && r.ContentLength != 0 {
		if err := decodeJSON(r, &in); err != nil {
			writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
			return
		}
	}
	a.mu.RLock()
	if in.GameVersion == "" {
		in.GameVersion = a.settings.GameVersion
	}
	if in.Loader == "" {
		in.Loader = a.settings.Loader
	}
	a.mu.RUnlock()
	ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
	defer cancel()
	plan, err := a.buildUpdatePlan(ctx, in.GameVersion, in.Loader)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, APIError{Error: err.Error()})
		return
	}
	a.dataMu.Lock()
	if a.updatePlans == nil {
		a.updatePlans = map[string]UpdatePlan{}
	}
	a.updatePlans[plan.ID] = plan
	for id, old := range a.updatePlans {
		t, _ := time.Parse(time.RFC3339, old.CreatedAt)
		if !t.IsZero() && time.Since(t) > 2*time.Hour {
			delete(a.updatePlans, id)
		}
	}
	a.dataMu.Unlock()
	writeJSON(w, http.StatusOK, plan)
}

func (a *App) handleUpdaterApply(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var in updaterApplyRequest
	if err := decodeJSON(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	a.dataMu.RLock()
	plan, ok := a.updatePlans[in.PlanID]
	a.dataMu.RUnlock()
	if !ok {
		writeJSON(w, http.StatusNotFound, APIError{Error: "update plan expired; scan again"})
		return
	}
	selected := map[string]bool{}
	for _, name := range in.ItemFilenames {
		selected[name] = true
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()
	result, err := a.applyUpdatePlan(ctx, plan, selected)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (a *App) buildUpdatePlan(ctx context.Context, game, loader string) (UpdatePlan, error) {
	modsDir := a.javaTargetDir("mods")
	locals, err := scanLocalModJars(modsDir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return UpdatePlan{}, err
	}
	plan := UpdatePlan{ID: randomToken(12), CreatedAt: time.Now().UTC().Format(time.RFC3339), GameVersion: game, Loader: loader, ModsDirectory: modsDir, Items: make([]UpdateItem, len(locals))}
	for i := range locals {
		plan.Items[i] = UpdateItem{Local: locals[i], Status: "unknown", Message: "Identifying installed project..."}
	}
	if len(locals) == 0 {
		return plan, nil
	}

	// Modrinth can identify local files by cryptographic hash and resolve the newest
	// compatible file in bulk. This is the preferred exact-match path.
	exactMR, updateMR := a.lookupModrinthUpdates(ctx, locals, game, loader)
	projectCache := map[string]ModrinthProject{}
	for i := range plan.Items {
		local := plan.Items[i].Local
		exact, identified := exactMR[local.SHA512]
		updated, hasUpdate := updateMR[local.SHA512]
		if identified {
			project := projectCache[exact.ProjectID]
			if project.ID == "" {
				if p, e := a.fetchModrinthProject(ctx, exact.ProjectID); e == nil {
					project = p
					projectCache[exact.ProjectID] = p
				}
			}
			plan.Items[i].InstalledProject = firstNonEmpty(project.Title, local.Metadata.Name, exact.ProjectID)
			plan.Items[i].InstalledVersion = exact.VersionNumber
			if hasUpdate && updated.ID != "" && updated.ID != exact.ID {
				cand := modrinthVersionCandidate(updated, project)
				cand.Safe = true
				cand.Confidence = 1
				cand.Reason = "Exact SHA-512 identity match + provider-declared target-version compatibility"
				plan.Items[i].SafeUpdate = &cand
				plan.Items[i].Status = "update"
				plan.Items[i].Message = "Exact compatible update found on Modrinth."
				continue
			}
			if hasUpdate && updated.ID == exact.ID {
				plan.Items[i].Status = "current"
				plan.Items[i].Message = "Already on the newest compatible Modrinth version."
				continue
			}
			plan.Items[i].Status = "review"
			plan.Items[i].Message = "This exact mod is known, but its project has no release for the target version. Looking for ports and continuations."
			plan.Items[i].Alternatives = a.findUpdateAlternatives(ctx, local, game, loader, plan.Items[i].InstalledProject)
			continue
		}
	}

	// Use CurseForge fingerprints for files Modrinth did not identify. If a project is
	// identified exactly, only a dependency-safe compatible file is offered as automatic.
	a.lookupCurseForgeUpdates(ctx, &plan, game, loader)

	// Developer-declared project URLs inside mod metadata are a strong identity signal even
	// when the installed JAR is too old or repacked to match provider fingerprints. Resolve
	// canonical Modrinth and CurseForge project URLs before falling back to repository or
	// fuzzy-name discovery. Any automatic replacement still goes through the downloaded-JAR
	// mod-id/loader preflight plus transactional backup/rollback in applyUpdatePlan.
	a.lookupDeclaredProviderMetadataUpdates(ctx, &plan, game, loader)

	// A JAR may explicitly name its canonical GitHub source repository even when it is
	// absent from Modrinth/CurseForge. Treat that source URL as an exact project identity,
	// then require release/asset text to prove the requested Minecraft version and loader
	// before a GitHub replacement can become an automatic update.
	a.lookupGitHubMetadataUpdates(ctx, &plan, game, loader)

	// Unknown local JARs are never silently guessed. Metadata and filename are used to
	// discover likely continuations, ports, or replacement projects for review.
	for i := range plan.Items {
		if plan.Items[i].Status == "unknown" {
			name := firstNonEmpty(plan.Items[i].Local.Metadata.Name, trimJarVersion(plan.Items[i].Local.Filename))
			plan.Items[i].Message = "No exact provider fingerprint match. Review intelligently ranked matches before replacing this file."
			plan.Items[i].Alternatives = a.findUpdateAlternatives(ctx, plan.Items[i].Local, game, loader, name)
			if len(plan.Items[i].Alternatives) > 0 {
				plan.Items[i].Status = "review"
			}
		}
	}
	for _, item := range plan.Items {
		switch item.Status {
		case "update":
			plan.SafeCount++
		case "current":
			plan.CurrentCount++
		case "review":
			plan.ReviewCount++
		default:
			plan.UnknownCount++
		}
	}
	return plan, nil
}

func scanLocalModJars(dir string) ([]LocalModFile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	out := []LocalModFile{}
	for _, de := range entries {
		if de.IsDir() {
			continue
		}
		name := de.Name()
		low := strings.ToLower(name)
		if !strings.HasSuffix(low, ".jar") && !strings.HasSuffix(low, ".jar.disabled") {
			continue
		}
		p := filepath.Join(dir, name)
		m, err := inspectLocalJar(p)
		if err != nil {
			continue
		}
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool { return strings.ToLower(out[i].Filename) < strings.ToLower(out[j].Filename) })
	return out, nil
}

func inspectLocalJar(path string) (LocalModFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return LocalModFile{}, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return LocalModFile{}, err
	}
	h1, h512 := sha1.New(), sha512.New()
	data, err := io.ReadAll(io.TeeReader(f, io.MultiWriter(h1, h512)))
	if err != nil {
		return LocalModFile{}, err
	}
	meta, _ := parseJarMetadataBytes(data)
	return LocalModFile{Path: path, Filename: filepath.Base(path), Enabled: !strings.HasSuffix(strings.ToLower(path), ".disabled"), Size: info.Size(), SHA1: hex.EncodeToString(h1.Sum(nil)), SHA512: hex.EncodeToString(h512.Sum(nil)), CurseFingerprint: curseForgeFingerprint(data), Metadata: meta}, nil
}

func parseJarMetadataBytes(data []byte) (JarMetadata, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return JarMetadata{}, err
	}
	files := map[string]*zip.File{}
	for _, f := range zr.File {
		files[strings.ToLower(strings.TrimPrefix(filepath.ToSlash(f.Name), "./"))] = f
	}
	manifestVersion := ""
	if f := files["meta-inf/manifest.mf"]; f != nil {
		if b, readErr := readZipFile(f, 2<<20); readErr == nil {
			manifestVersion = firstNonEmpty(manifestAttribute(string(b), "Implementation-Version"), manifestAttribute(string(b), "Specification-Version"))
		}
	}
	if f := files["fabric.mod.json"]; f != nil {
		b, _ := readZipFile(f, 2<<20)
		var v struct {
			ID           string         `json:"id"`
			Name         string         `json:"name"`
			Version      string         `json:"version"`
			Description  string         `json:"description"`
			Environment  string         `json:"environment"`
			License      any            `json:"license"`
			Icon         any            `json:"icon"`
			Authors      []any          `json:"authors"`
			Contributors []any          `json:"contributors"`
			Contact      map[string]any `json:"contact"`
			Depends      map[string]any `json:"depends"`
		}
		if json.Unmarshal(b, &v) == nil {
			meta := JarMetadata{
				ModID: v.ID, Name: v.Name, Version: firstNonEmpty(v.Version, manifestVersion), Description: v.Description,
				Environment: v.Environment, License: flexibleString(v.License), IconPath: fabricIconPath(v.Icon), Loaders: []string{"fabric"},
				SourceURL: flexibleString(v.Contact["sources"]), Homepage: flexibleString(v.Contact["homepage"]),
				Minecraft: flexibleString(v.Depends["minecraft"]), MetadataBy: "fabric.mod.json",
			}
			meta.Authors = uniqueStringsPreserve(append(flexiblePeople(v.Authors), flexiblePeople(v.Contributors)...))
			meta.Dependencies = dependencyKeys(v.Depends)
			return meta, nil
		}
	}
	if f := files["quilt.mod.json"]; f != nil {
		b, _ := readZipFile(f, 2<<20)
		var v struct {
			QuiltLoader struct {
				ID       string `json:"id"`
				Version  string `json:"version"`
				Metadata struct {
					Name         string            `json:"name"`
					Description  string            `json:"description"`
					Contact      map[string]string `json:"contact"`
					License      any               `json:"license"`
					Icon         any               `json:"icon"`
					Contributors any               `json:"contributors"`
				} `json:"metadata"`
				Depends []any `json:"depends"`
			} `json:"quilt_loader"`
		}
		if json.Unmarshal(b, &v) == nil {
			meta := JarMetadata{
				ModID: v.QuiltLoader.ID, Name: v.QuiltLoader.Metadata.Name,
				Version: firstNonEmpty(v.QuiltLoader.Version, manifestVersion), Description: v.QuiltLoader.Metadata.Description,
				License: flexibleString(v.QuiltLoader.Metadata.License), IconPath: quiltIconPath(v.QuiltLoader.Metadata.Icon),
				Loaders: []string{"quilt"}, SourceURL: v.QuiltLoader.Metadata.Contact["sources"],
				Homepage: v.QuiltLoader.Metadata.Contact["homepage"], MetadataBy: "quilt.mod.json",
			}
			meta.Authors = flexibleContributorMap(v.QuiltLoader.Metadata.Contributors)
			meta.Dependencies = quiltDependencyIDs(v.QuiltLoader.Depends)
			return meta, nil
		}
	}
	for _, name := range []string{"meta-inf/neoforge.mods.toml", "meta-inf/mods.toml"} {
		if f := files[name]; f != nil {
			b, _ := readZipFile(f, 2<<20)
			t := string(b)
			loader := "forge"
			if strings.Contains(name, "neoforge") || strings.Contains(strings.ToLower(t), "neoforge") {
				loader = "neoforge"
			}
			version := tomlString(t, "version")
			if strings.Contains(version, "${") || version == "" {
				version = manifestVersion
			}
			meta := JarMetadata{
				ModID: tomlString(t, "modId"), Name: tomlString(t, "displayName"), Version: version,
				Homepage: tomlString(t, "displayURL"), SourceURL: firstNonEmpty(tomlString(t, "issueTrackerURL"), tomlString(t, "sourceURL")),
				Description: tomlMultilineString(t, "description"), IconPath: tomlString(t, "logoFile"), License: tomlString(t, "license"),
				Credits: tomlString(t, "credits"), Authors: splitPeople(tomlString(t, "authors")), Loaders: []string{loader},
				Minecraft: forgeMinecraftRange(t), Dependencies: forgeDependencyIDs(t), MetadataBy: filepath.Base(name),
			}
			return meta, nil
		}
	}
	for _, name := range []string{"paper-plugin.yml", "plugin.yml", "bungee.yml", "velocity-plugin.json"} {
		if f := files[name]; f != nil {
			b, _ := readZipFile(f, 2<<20)
			meta := parsePluginMetadata(name, b)
			if meta.Version == "" {
				meta.Version = manifestVersion
			}
			return meta, nil
		}
	}
	return JarMetadata{}, errors.New("no supported mod or plugin metadata found")
}

func manifestAttribute(text, key string) string {
	key = strings.ToLower(strings.TrimSpace(key))
	var currentKey string
	var currentValue strings.Builder
	flush := func() string {
		if strings.EqualFold(currentKey, key) {
			return strings.TrimSpace(currentValue.String())
		}
		return ""
	}
	for _, line := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		if strings.HasPrefix(line, " ") && currentKey != "" {
			currentValue.WriteString(strings.TrimPrefix(line, " "))
			continue
		}
		if value := flush(); value != "" {
			return value
		}
		currentKey, currentValue = "", strings.Builder{}
		if i := strings.Index(line, ":"); i > 0 {
			currentKey = strings.TrimSpace(line[:i])
			currentValue.WriteString(strings.TrimSpace(line[i+1:]))
		}
	}
	return flush()
}

func flexibleString(v any) string {
	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x)
	case []any:
		parts := make([]string, 0, len(x))
		for _, item := range x {
			if value := flexibleString(item); value != "" {
				parts = append(parts, value)
			}
		}
		return strings.Join(uniqueStringsPreserve(parts), ", ")
	case map[string]any:
		for _, key := range []string{"name", "value", "url"} {
			if value := flexibleString(x[key]); value != "" {
				return value
			}
		}
	}
	return ""
}

func flexiblePeople(values []any) []string {
	out := []string{}
	for _, value := range values {
		if person := flexibleString(value); person != "" {
			out = append(out, person)
		}
	}
	return uniqueStringsPreserve(out)
}

func flexibleContributorMap(value any) []string {
	out := []string{}
	switch x := value.(type) {
	case map[string]any:
		for name := range x {
			if strings.TrimSpace(name) != "" {
				out = append(out, strings.TrimSpace(name))
			}
		}
	case []any:
		out = flexiblePeople(x)
	}
	sort.Strings(out)
	return uniqueStringsPreserve(out)
}

func fabricIconPath(value any) string {
	switch x := value.(type) {
	case string:
		return strings.TrimSpace(x)
	case map[string]any:
		bestSize := -1
		best := ""
		for key, raw := range x {
			size, _ := strconv.Atoi(key)
			if path := flexibleString(raw); path != "" && size >= bestSize {
				bestSize, best = size, path
			}
		}
		return best
	}
	return ""
}

func quiltIconPath(value any) string {
	return fabricIconPath(value)
}

func dependencyKeys(values map[string]any) []string {
	out := make([]string, 0, len(values))
	for key := range values {
		if key != "minecraft" && key != "java" && strings.TrimSpace(key) != "" {
			out = append(out, key)
		}
	}
	sort.Strings(out)
	return out
}

func quiltDependencyIDs(values []any) []string {
	out := []string{}
	for _, raw := range values {
		if m, ok := raw.(map[string]any); ok {
			if id := flexibleString(m["id"]); id != "" && id != "minecraft" && id != "java" {
				out = append(out, id)
			}
		}
	}
	return uniqueStringsPreserve(out)
}

func tomlMultilineString(text, key string) string {
	patterns := []string{
		`(?ms)^\s*` + regexp.QuoteMeta(key) + `\s*=\s*"""(.*?)"""`,
		`(?ms)^\s*` + regexp.QuoteMeta(key) + `\s*=\s*'''(.*?)'''`,
		`(?mi)^\s*` + regexp.QuoteMeta(key) + `\s*=\s*["']([^"']*)["']`,
	}
	for _, pattern := range patterns {
		match := regexp.MustCompile(pattern).FindStringSubmatch(text)
		if len(match) > 1 {
			return strings.TrimSpace(match[1])
		}
	}
	return ""
}

func splitPeople(text string) []string {
	text = strings.NewReplacer(";", ",", " and ", ",", "&", ",").Replace(text)
	out := []string{}
	for _, part := range strings.Split(text, ",") {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return uniqueStringsPreserve(out)
}

func forgeMinecraftRange(text string) string {
	for _, block := range forgeDependencyBlocks(text) {
		if strings.EqualFold(tomlString(block, "modId"), "minecraft") {
			return tomlString(block, "versionRange")
		}
	}
	return ""
}

func forgeDependencyIDs(text string) []string {
	out := []string{}
	for _, block := range forgeDependencyBlocks(text) {
		id := strings.TrimSpace(tomlString(block, "modId"))
		if id != "" && id != "minecraft" && id != "forge" && id != "neoforge" && id != "java" {
			out = append(out, id)
		}
	}
	return uniqueStringsPreserve(out)
}

// Go's regexp engine intentionally does not support look-ahead. Split the
// TOML dependency tables by their real section boundaries instead of relying
// on a PCRE-only expression; this also prevents the primary [[mods]] modId
// from being mistaken for a dependency.
func forgeDependencyBlocks(text string) []string {
	header := regexp.MustCompile(`(?mi)^\s*\[\[dependencies\.[^\]]+\]\]\s*$`)
	locations := header.FindAllStringIndex(text, -1)
	blocks := make([]string, 0, len(locations))
	for i, location := range locations {
		start := location[1]
		end := len(text)
		if i+1 < len(locations) {
			end = locations[i+1][0]
		}
		blocks = append(blocks, text[start:end])
	}
	return blocks
}

func parsePluginMetadata(name string, data []byte) JarMetadata {
	loader := "plugin"
	if strings.EqualFold(name, "velocity-plugin.json") {
		var value struct {
			ID           string   `json:"id"`
			Name         string   `json:"name"`
			Version      string   `json:"version"`
			Description  string   `json:"description"`
			Authors      []string `json:"authors"`
			URL          string   `json:"url"`
			Dependencies []struct {
				ID string `json:"id"`
			} `json:"dependencies"`
		}
		if json.Unmarshal(data, &value) == nil {
			deps := []string{}
			for _, dep := range value.Dependencies {
				deps = append(deps, dep.ID)
			}
			return JarMetadata{ModID: value.ID, Name: value.Name, Version: value.Version, Description: value.Description, Authors: value.Authors, Homepage: value.URL, Loaders: []string{"velocity"}, Dependencies: uniqueStringsPreserve(deps), MetadataBy: name}
		}
	}
	text := string(data)
	value := func(key string) string {
		re := regexp.MustCompile(`(?mi)^\s*` + regexp.QuoteMeta(key) + `\s*:\s*["']?([^\r\n#"']+)`)
		m := re.FindStringSubmatch(text)
		if len(m) > 1 {
			return strings.TrimSpace(m[1])
		}
		return ""
	}
	if name == "paper-plugin.yml" {
		loader = "paper"
	} else if name == "bungee.yml" {
		loader = "bungeecord"
	} else {
		loader = "bukkit"
	}
	authors := splitPeople(firstNonEmpty(value("authors"), value("author")))
	return JarMetadata{ModID: firstNonEmpty(value("name"), value("id")), Name: value("name"), Version: value("version"), Description: value("description"), Homepage: value("website"), Authors: authors, Loaders: []string{loader}, MetadataBy: name}
}

func readZipFile(f *zip.File, limit int64) ([]byte, error) {
	r, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(io.LimitReader(r, limit))
}

func tomlString(text, key string) string {
	re := regexp.MustCompile(`(?mi)^\s*` + regexp.QuoteMeta(key) + `\s*=\s*["']([^"']+)["']`)
	m := re.FindStringSubmatch(text)
	if len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	return ""
}

// curseForgeFingerprint implements CurseForge's whitespace-stripped MurmurHash2
// fingerprint. The behavior is intentionally compatible with packwiz's MIT-licensed
// curseforge/murmur2 implementation; see THIRD-PARTY-NOTICES.md.
func curseForgeFingerprint(data []byte) uint32 {
	filtered := make([]byte, 0, len(data))
	for _, b := range data {
		if b != 9 && b != 10 && b != 13 && b != 32 {
			filtered = append(filtered, b)
		}
	}
	const m uint32 = 0x5bd1e995
	const seed uint32 = 1
	h := seed ^ uint32(len(filtered))
	buf := filtered
	for len(buf) >= 4 {
		k := binary.LittleEndian.Uint32(buf[:4])
		k *= m
		k ^= k >> 24
		k *= m
		h *= m
		h ^= k
		buf = buf[4:]
	}
	switch len(buf) {
	case 3:
		h ^= uint32(buf[2]) << 16
		fallthrough
	case 2:
		h ^= uint32(buf[1]) << 8
		fallthrough
	case 1:
		h ^= uint32(buf[0])
		h *= m
	}
	h ^= h >> 13
	h *= m
	h ^= h >> 15
	return h
}

func (a *App) lookupModrinthUpdates(ctx context.Context, locals []LocalModFile, game, loader string) (map[string]ModrinthVersion, map[string]ModrinthVersion) {
	hashes := make([]string, 0, len(locals))
	for _, f := range locals {
		if f.SHA512 != "" {
			hashes = append(hashes, f.SHA512)
		}
	}
	exact := map[string]ModrinthVersion{}
	updates := map[string]ModrinthVersion{}
	if len(hashes) == 0 {
		return exact, updates
	}
	_ = a.postJSON(ctx, modrinthAPIBase()+"/version_files", nil, map[string]any{"hashes": hashes, "algorithm": "sha512"}, &exact)
	body := map[string]any{"hashes": hashes, "algorithm": "sha512", "game_versions": []string{game}}
	if loader != "" && loader != "any" && loader != "vanilla" {
		body["loaders"] = []string{loader}
	}
	_ = a.postJSON(ctx, modrinthAPIBase()+"/version_files/update", nil, body, &updates)
	return exact, updates
}

func modrinthVersionCandidate(v ModrinthVersion, p ModrinthProject) UpdateCandidate {
	var f ModrinthVersionFile
	if len(v.Files) > 0 {
		f = v.Files[0]
		for _, x := range v.Files {
			if x.Primary {
				f = x
				break
			}
		}
	}
	return UpdateCandidate{ID: "modrinth:" + v.ID, Provider: "modrinth", ProjectID: v.ProjectID, ProjectTitle: firstNonEmpty(p.Title, v.ProjectID), VersionID: v.ID, Version: v.VersionNumber, Filename: f.Filename, URL: f.URL, PageURL: "https://modrinth.com/mod/" + firstNonEmpty(p.Slug, v.ProjectID), Hashes: f.Hashes, Size: f.Size, GameVersions: v.GameVersions, Loaders: v.Loaders}
}

func (a *App) lookupCurseForgeUpdates(ctx context.Context, plan *UpdatePlan, game, loader string) {
	a.mu.RLock()
	key := strings.TrimSpace(a.settings.CurseForgeAPIKey)
	a.mu.RUnlock()
	if key == "" {
		return
	}
	indices := []int{}
	fps := []uint32{}
	for i := range plan.Items {
		if plan.Items[i].Status == "unknown" || plan.Items[i].Status == "review" && plan.Items[i].InstalledProject == "" {
			indices = append(indices, i)
			fps = append(fps, plan.Items[i].Local.CurseFingerprint)
		}
	}
	if len(fps) == 0 {
		return
	}
	var matches struct {
		Data struct {
			ExactMatches []struct {
				ID   uint32 `json:"id"`
				File struct {
					ID          int64  `json:"id"`
					ModID       int64  `json:"modId"`
					FileName    string `json:"fileName"`
					DisplayName string `json:"displayName"`
				} `json:"file"`
			} `json:"exactMatches"`
		} `json:"data"`
	}
	if err := a.postJSON(ctx, curseForgeAPIBase()+"/v1/fingerprints", map[string]string{"x-api-key": key}, map[string]any{"fingerprints": fps}, &matches); err != nil {
		return
	}
	byFP := map[uint32]struct {
		ModID                 int64
		FileID                int64
		FileName, DisplayName string
	}{}
	for _, x := range matches.Data.ExactMatches {
		byFP[x.ID] = struct {
			ModID                 int64
			FileID                int64
			FileName, DisplayName string
		}{x.File.ModID, x.File.ID, x.File.FileName, x.File.DisplayName}
	}
	for _, idx := range indices {
		item := &plan.Items[idx]
		match, ok := byFP[item.Local.CurseFingerprint]
		if !ok {
			continue
		}
		item.InstalledProject = firstNonEmpty(match.DisplayName, item.Local.Metadata.Name, trimJarVersion(item.Local.Filename))
		item.InstalledVersion = strconv.FormatInt(match.FileID, 10)
		cand, found := a.curseForgeCompatibleCandidate(ctx, key, match.ModID, match.FileID, game, loader)
		if found {
			if cand.ID == fmt.Sprintf("curseforge:%d", match.FileID) {
				item.Status = "current"
				item.Message = "Already on the newest compatible CurseForge file."
			} else if !cand.DependencyRisk {
				cand.Safe = true
				cand.Confidence = 1
				cand.Reason = "Exact CurseForge MurmurHash2 fingerprint + target-version/loader file match"
				item.SafeUpdate = &cand
				item.Status = "update"
				item.Message = "Exact compatible update found on CurseForge."
			} else {
				cand.Safe = false
				cand.Confidence = .95
				cand.Reason = "Exact CurseForge project match, but the target file adds required dependencies; review before applying"
				item.Alternatives = append([]UpdateCandidate{cand}, item.Alternatives...)
				item.Status = "review"
			}
		}
	}
}

type declaredProviderProject struct {
	Provider  string
	ProjectID string
	PageURL   string
}

func declaredProviderProjects(meta JarMetadata) []declaredProviderProject {
	out := []declaredProviderProject{}
	seen := map[string]bool{}
	for _, raw := range []string{meta.SourceURL, meta.Homepage} {
		provider, id, canonical, ok := integratedProviderFromURL(strings.TrimSpace(raw))
		if !ok || (provider != "modrinth" && provider != "curseforge") {
			continue
		}
		if provider == "curseforge" {
			u, err := url.Parse(canonical)
			if err != nil {
				continue
			}
			parts := strings.Split(strings.Trim(u.Path, "/"), "/")
			for i, part := range parts {
				if i+1 < len(parts) && containsFold([]string{"mc-mods", "mc-mods-server", "bukkit-plugins", "mc-addons"}, part) {
					id = parts[i+1]
					break
				}
			}
		}
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		key := provider + ":" + strings.ToLower(id)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, declaredProviderProject{Provider: provider, ProjectID: id, PageURL: canonical})
	}
	return out
}

func declaredVersionMatchesCandidate(version string, cand UpdateCandidate) bool {
	version = normalizeDeclaredVersion(version)
	if version == "" {
		return false
	}
	for _, value := range []string{cand.Version, cand.VersionID, cand.Filename} {
		normalized := normalizeDeclaredVersion(value)
		if normalized == version {
			return true
		}
		if normalized != "" {
			re := regexp.MustCompile(`(^|[^0-9a-z])` + regexp.QuoteMeta(version) + `([^0-9a-z]|$)`)
			if re.MatchString(normalized) {
				return true
			}
		}
	}
	return false
}

func (a *App) lookupDeclaredProviderMetadataUpdates(ctx context.Context, plan *UpdatePlan, game, loader string) {
	a.mu.RLock()
	curseKey := strings.TrimSpace(a.settings.CurseForgeAPIKey)
	a.mu.RUnlock()
	for i := range plan.Items {
		item := &plan.Items[i]
		if item.SafeUpdate != nil || (item.Status != "unknown" && item.Status != "review") {
			continue
		}
		for _, ref := range declaredProviderProjects(item.Local.Metadata) {
			switch ref.Provider {
			case "modrinth":
				project, err := a.fetchModrinthProject(ctx, ref.ProjectID)
				if err != nil || project.ID == "" || (project.ProjectType != "" && project.ProjectType != "mod") {
					continue
				}
				versions, err := a.fetchModrinthVersions(ctx, project.ID, game, loader, "mod")
				item.InstalledProject = firstNonEmpty(item.InstalledProject, project.Title, item.Local.Metadata.Name, project.Slug)
				item.InstalledVersion = firstNonEmpty(item.InstalledVersion, item.Local.Metadata.Version)
				if err != nil || len(versions) == 0 {
					item.Status = "review"
					item.Message = "Developer-declared Modrinth project identified, but it has no provider-declared release for this Minecraft version and loader. Looking for ports and continuations."
					item.Alternatives = a.findUpdateAlternatives(ctx, item.Local, game, loader, item.InstalledProject)
					break
				}
				cand := modrinthVersionCandidate(versions[0], project)
				if cand.URL == "" || cand.Filename == "" {
					continue
				}
				if declaredVersionMatchesCandidate(item.Local.Metadata.Version, cand) {
					item.Status = "current"
					item.Message = "Developer-declared Modrinth project is already on the newest provider-compatible release."
					break
				}
				cand.Safe = true
				cand.Confidence = .995
				cand.Reason = "Developer-declared canonical Modrinth project URL + provider-declared target-version/loader compatibility; apply revalidates downloaded mod ID and loader before atomic replacement"
				item.SafeUpdate = &cand
				item.Status = "update"
				item.Message = "Canonical Modrinth metadata found a compatible provider release; downloaded JAR identity is revalidated before replacement."
			case "curseforge":
				if curseKey == "" {
					continue
				}
				v := url.Values{"gameId": {"432"}, "slug": {ref.ProjectID}, "pageSize": {"20"}}
				var search struct {
					Data []struct {
						ID   int64  `json:"id"`
						Name string `json:"name"`
						Slug string `json:"slug"`
					} `json:"data"`
				}
				if err := a.getJSON(ctx, curseForgeAPIBase()+"/v1/mods/search?"+v.Encode(), map[string]string{"x-api-key": curseKey}, &search); err != nil {
					continue
				}
				var modID int64
				var title string
				for _, p := range search.Data {
					if strings.EqualFold(strings.TrimSpace(p.Slug), strings.TrimSpace(ref.ProjectID)) {
						modID, title = p.ID, p.Name
						break
					}
				}
				if modID == 0 {
					continue
				}
				item.InstalledProject = firstNonEmpty(item.InstalledProject, title, item.Local.Metadata.Name, ref.ProjectID)
				item.InstalledVersion = firstNonEmpty(item.InstalledVersion, item.Local.Metadata.Version)
				cand, found := a.curseForgeCompatibleCandidate(ctx, curseKey, modID, 0, game, loader)
				if !found {
					item.Status = "review"
					item.Message = "Developer-declared CurseForge project identified, but no API-declared compatible file exists for this Minecraft version and loader."
					item.Alternatives = a.findUpdateAlternatives(ctx, item.Local, game, loader, item.InstalledProject)
					break
				}
				cand.ProjectTitle = firstNonEmpty(title, cand.ProjectTitle, item.Local.Metadata.Name)
				cand.PageURL = ref.PageURL
				if declaredVersionMatchesCandidate(item.Local.Metadata.Version, cand) {
					item.Status = "current"
					item.Message = "Developer-declared CurseForge project is already on the newest provider-compatible file."
					break
				}
				if cand.DependencyRisk {
					cand.Safe = false
					cand.Confidence = .97
					cand.Reason = "Developer-declared canonical CurseForge project URL + target-version/loader match, but the target file introduces required dependencies; review before applying"
					item.Alternatives = append([]UpdateCandidate{cand}, item.Alternatives...)
					item.Status = "review"
					item.Message = "Canonical CurseForge project found a compatible file with required dependency changes; review before applying."
					break
				}
				cand.Safe = true
				cand.Confidence = .995
				cand.Reason = "Developer-declared canonical CurseForge project URL + provider-declared target-version/loader file match; apply revalidates downloaded mod ID and loader before atomic replacement"
				item.SafeUpdate = &cand
				item.Status = "update"
				item.Message = "Canonical CurseForge metadata found a compatible file; downloaded JAR identity is revalidated before replacement."
			}
			if item.Status != "unknown" {
				break
			}
		}
	}
}

type curseForgeFileDependency struct {
	ModID        int64 `json:"modId"`
	RelationType int   `json:"relationType"`
}

type curseForgeFileRecord struct {
	ID           int64    `json:"id"`
	FileName     string   `json:"fileName"`
	DisplayName  string   `json:"displayName"`
	DownloadURL  string   `json:"downloadUrl"`
	FileLength   int64    `json:"fileLength"`
	FileDate     string   `json:"fileDate"`
	ReleaseType  int      `json:"releaseType"`
	GameVersions []string `json:"gameVersions"`
	Hashes       []struct {
		Value string `json:"value"`
		Algo  int    `json:"algo"`
	} `json:"hashes"`
	Dependencies []curseForgeFileDependency `json:"dependencies"`
}

func (a *App) fetchCurseForgeCompatibleFile(ctx context.Context, key string, modID int64, game, loader string) (curseForgeFileRecord, bool) {
	v := url.Values{"pageSize": {"50"}}
	if game != "" {
		v.Set("gameVersion", game)
	}
	if lt := curseLoaderType(loader); lt != 0 {
		v.Set("modLoaderType", strconv.Itoa(lt))
	}
	var resp struct {
		Data []curseForgeFileRecord `json:"data"`
	}
	if err := a.getJSON(ctx, fmt.Sprintf("%s/v1/mods/%d/files?%s", curseForgeAPIBase(), modID, v.Encode()), map[string]string{"x-api-key": key}, &resp); err != nil || len(resp.Data) == 0 {
		return curseForgeFileRecord{}, false
	}
	return resp.Data[0], true
}

func curseForgeFileHashes(f curseForgeFileRecord) map[string]string {
	hashes := map[string]string{}
	for _, h := range f.Hashes {
		switch h.Algo {
		case 1:
			hashes["sha1"] = h.Value
		case 2:
			hashes["md5"] = h.Value
		}
	}
	return hashes
}

func curseForgeRequiredDependencies(f curseForgeFileRecord) []int64 {
	out := []int64{}
	seen := map[int64]bool{}
	for _, d := range f.Dependencies {
		if d.RelationType != 3 || d.ModID <= 0 || seen[d.ModID] {
			continue
		}
		seen[d.ModID] = true
		out = append(out, d.ModID)
	}
	return out
}

func (a *App) curseForgeCompatibleCandidate(ctx context.Context, key string, modID, currentFileID int64, game, loader string) (UpdateCandidate, bool) {
	f, ok := a.fetchCurseForgeCompatibleFile(ctx, key, modID, game, loader)
	if !ok {
		return UpdateCandidate{}, false
	}
	dependencyRisk := len(curseForgeRequiredDependencies(f)) > 0
	return UpdateCandidate{ID: fmt.Sprintf("curseforge:%d", f.ID), Provider: "curseforge", ProjectID: strconv.FormatInt(modID, 10), ProjectTitle: f.DisplayName, VersionID: strconv.FormatInt(f.ID, 10), Version: f.DisplayName, Filename: f.FileName, URL: f.DownloadURL, PageURL: fmt.Sprintf("https://www.curseforge.com/minecraft/mc-mods/x/files/%d", f.ID), Hashes: curseForgeFileHashes(f), Size: f.FileLength, GameVersions: f.GameVersions, Loaders: []string{loader}, Safe: f.ID != currentFileID && !dependencyRisk, DependencyRisk: dependencyRisk}, true
}

type githubUpdaterRelease struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	Body        string `json:"body"`
	PublishedAt string `json:"published_at"`
	Draft       bool   `json:"draft"`
	Prerelease  bool   `json:"prerelease"`
	Assets      []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int64  `json:"size"`
		Digest             string `json:"digest"`
	} `json:"assets"`
}

func githubRepositoryFromMetadata(meta JarMetadata) string {
	for _, raw := range []string{meta.SourceURL, meta.Homepage} {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		u, err := url.Parse(raw)
		if err != nil || !strings.EqualFold(u.Hostname(), "github.com") {
			continue
		}
		parts := strings.Split(strings.Trim(u.Path, "/"), "/")
		if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
			continue
		}
		repo := strings.TrimSuffix(parts[1], ".git")
		if repo != "" {
			return parts[0] + "/" + repo
		}
	}
	return ""
}

func githubCompatibilityText(rel githubUpdaterRelease, assetName string) string {
	return strings.ToLower(strings.Join([]string{rel.TagName, rel.Name, rel.Body, assetName}, " "))
}

func githubTextProvesGame(text, game string) bool {
	game = strings.ToLower(strings.TrimSpace(game))
	if game == "" || game == "any" {
		return true
	}
	return strings.Contains(strings.ToLower(text), game)
}

func githubTextProvesLoader(text, loader string) bool {
	loader = strings.ToLower(strings.TrimSpace(loader))
	if loader == "" || loader == "any" || loader == "vanilla" {
		return true
	}
	normalized := regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(strings.ToLower(text), "")
	loaderNormalized := regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(loader, "")
	return loaderNormalized != "" && strings.Contains(normalized, loaderNormalized)
}

func githubUpdaterAssetAllowed(name string) bool {
	low := strings.ToLower(strings.TrimSpace(name))
	if !strings.HasSuffix(low, ".jar") {
		return false
	}
	for _, bad := range []string{"sources", "source-", "javadoc", "dev", "deobf", "api-", "installer"} {
		if strings.Contains(low, bad) {
			return false
		}
	}
	return true
}

func githubDigestHashes(digest string) map[string]string {
	digest = strings.TrimSpace(digest)
	parts := strings.SplitN(digest, ":", 2)
	if len(parts) != 2 {
		return nil
	}
	algo, value := strings.ToLower(strings.TrimSpace(parts[0])), strings.ToLower(strings.TrimSpace(parts[1]))
	if algo != "sha256" || len(value) != 64 {
		return nil
	}
	if _, err := hex.DecodeString(value); err != nil {
		return nil
	}
	return map[string]string{"sha256": value}
}

func normalizeDeclaredVersion(v string) string {
	v = strings.TrimSpace(strings.ToLower(v))
	v = strings.TrimPrefix(v, "version-")
	v = strings.TrimPrefix(v, "release-")
	v = strings.TrimPrefix(v, "v")
	return strings.TrimSpace(v)
}

func (a *App) lookupGitHubMetadataUpdates(ctx context.Context, plan *UpdatePlan, game, loader string) {
	a.mu.RLock()
	token := strings.TrimSpace(a.settings.GitHubToken)
	a.mu.RUnlock()
	headers := map[string]string{"Accept": "application/vnd.github+json", "X-GitHub-Api-Version": "2022-11-28"}
	if token != "" {
		headers["Authorization"] = "Bearer " + token
	}
	base := providerBase("MMV_GITHUB_API_BASE", "https://api.github.com")
	for i := range plan.Items {
		item := &plan.Items[i]
		if item.SafeUpdate != nil || (item.Status != "unknown" && item.Status != "review") {
			continue
		}
		repo := githubRepositoryFromMetadata(item.Local.Metadata)
		if repo == "" {
			continue
		}
		var releases []githubUpdaterRelease
		endpoint := base + "/repos/" + repo + "/releases?per_page=30"
		if err := a.getJSON(ctx, endpoint, headers, &releases); err != nil {
			continue
		}
		var strict *UpdateCandidate
		var review *UpdateCandidate
		for _, rel := range releases {
			if rel.Draft {
				continue
			}
			for _, asset := range rel.Assets {
				if !githubUpdaterAssetAllowed(asset.Name) || strings.TrimSpace(asset.BrowserDownloadURL) == "" {
					continue
				}
				text := githubCompatibilityText(rel, asset.Name)
				gameOK := githubTextProvesGame(text, game)
				loaderOK := githubTextProvesLoader(text, loader)
				cand := UpdateCandidate{
					ID: "github:" + repo + ":" + firstNonEmpty(rel.TagName, asset.Name), Provider: "github", ProjectID: repo,
					ProjectTitle: firstNonEmpty(item.Local.Metadata.Name, trimJarVersion(item.Local.Filename), repo),
					VersionID:    rel.TagName, Version: firstNonEmpty(rel.TagName, rel.Name), Filename: asset.Name,
					URL: asset.BrowserDownloadURL, PageURL: "https://github.com/" + repo + "/releases", Hashes: githubDigestHashes(asset.Digest), Size: asset.Size,
					GameVersions: nonEmptyStrings(game), Loaders: nonEmptyStrings(loader), Confidence: .96,
					Reason: "Exact GitHub repository from embedded JAR metadata; replacement is hash/size verified when GitHub publishes a digest",
				}
				if gameOK && loaderOK {
					cand.Safe = true
					cand.Confidence = 1
					cand.Reason += "; release metadata explicitly matches the requested Minecraft version and loader"
					strict = &cand
					break
				}
				if review == nil {
					cand.Safe = false
					cand.Reason += "; target-version/loader compatibility is not explicit enough for automatic replacement"
					review = &cand
				}
			}
			if strict != nil {
				break
			}
		}
		item.InstalledProject = firstNonEmpty(item.InstalledProject, item.Local.Metadata.Name, repo)
		item.InstalledVersion = firstNonEmpty(item.InstalledVersion, item.Local.Metadata.Version)
		if strict != nil {
			if normalizeDeclaredVersion(strict.Version) == normalizeDeclaredVersion(item.Local.Metadata.Version) && normalizeDeclaredVersion(item.Local.Metadata.Version) != "" {
				item.Status = "current"
				item.Message = "Exact GitHub source is already on the matching declared release for this target."
				continue
			}
			item.SafeUpdate = strict
			item.Status = "update"
			item.Message = "Exact GitHub source metadata found a release explicitly matching the target Minecraft version and loader."
			continue
		}
		if review != nil {
			item.Alternatives = append([]UpdateCandidate{*review}, item.Alternatives...)
			item.Status = "review"
			item.Message = "Exact GitHub source repository found, but the release metadata does not prove target compatibility; review instead of guessing."
		}
	}
}

func (a *App) findUpdateAlternatives(ctx context.Context, local LocalModFile, game, loader, name string) []UpdateCandidate {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}
	queries := []string{name, name + " continuation", name + " port", name + " unofficial"}
	seen := map[string]bool{}
	var out []UpdateCandidate
	for qi, q := range queries {
		resp := a.searchProviders(ctx, providerSearchOptions{Query: q, GameVersion: game, Loader: loader, ProjectType: "mod", Limit: 8, Sources: a.enabledProvidersForType("mod")})
		for _, p := range resp.Results {
			key := p.Provider + ":" + p.ID
			if seen[key] {
				continue
			}
			seen[key] = true
			sim := titleSimilarity(name, p.Title)
			if sim < .22 && qi == 0 {
				continue
			}
			confidence := mathMin(.92, .45+sim*.45)
			reason := "Name/metadata similarity"
			low := strings.ToLower(p.Title + " " + p.Summary)
			if strings.Contains(low, "continuation") || strings.Contains(low, "port") || strings.Contains(low, "fork") || strings.Contains(low, "unofficial") {
				confidence = mathMin(.97, confidence+.12)
				reason += " + project explicitly looks like a continuation/port/fork"
			}
			out = append(out, UpdateCandidate{ID: key, Provider: p.Provider, ProjectID: p.ID, ProjectTitle: p.Title, Version: "review provider releases", PageURL: p.PageURL, GameVersions: p.Versions, Loaders: p.Loaders, Confidence: confidence, Reason: reason + "; replacement is never auto-applied", Safe: false})
		}
		if len(out) >= 8 {
			break
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Confidence > out[j].Confidence })
	if len(out) > 8 {
		out = out[:8]
	}
	return out
}

func titleSimilarity(a, b string) float64 {
	aTokens := tokenSet(a)
	bTokens := tokenSet(b)
	if len(aTokens) == 0 || len(bTokens) == 0 {
		return 0
	}
	inter := 0
	union := map[string]bool{}
	for k := range aTokens {
		union[k] = true
		if bTokens[k] {
			inter++
		}
	}
	for k := range bTokens {
		union[k] = true
	}
	return float64(inter) / float64(len(union))
}

func tokenSet(s string) map[string]bool {
	s = strings.ToLower(regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(s, " "))
	out := map[string]bool{}
	for _, w := range strings.Fields(s) {
		if len(w) > 1 && w != "mod" && w != "minecraft" {
			out[w] = true
		}
	}
	return out
}

func mathMin(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func trimJarVersion(name string) string {
	s := strings.TrimSuffix(strings.TrimSuffix(name, ".disabled"), ".jar")
	s = regexp.MustCompile(`(?i)[-_ ]v?\d+(?:[._-]\d+)*(?:[-_].*)?$`).ReplaceAllString(s, "")
	s = strings.ReplaceAll(strings.ReplaceAll(s, "_", " "), "-", " ")
	return strings.Join(strings.Fields(s), " ")
}

func (a *App) applyUpdatePlan(ctx context.Context, plan UpdatePlan, selected map[string]bool) (map[string]any, error) {
	backupRoot := filepath.Join(a.cfgDir, "update-backups", time.Now().UTC().Format("20060102-150405")+"-"+plan.ID[:8])
	stageRoot := filepath.Join(a.cfgDir, "update-staging", plan.ID)
	if err := os.MkdirAll(backupRoot, 0o755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(stageRoot, 0o755); err != nil {
		return nil, err
	}
	type appliedOp struct{ oldPath, backupPath, newPath string }
	applied := []appliedOp{}
	installedDeps := []string{}
	rollback := func() {
		for i := len(applied) - 1; i >= 0; i-- {
			op := applied[i]
			_ = os.Remove(op.newPath)
			_ = os.Rename(op.backupPath, op.oldPath)
		}
		for _, p := range installedDeps {
			_ = os.Remove(p)
		}
	}
	updated := []map[string]string{}
	for _, item := range plan.Items {
		if item.SafeUpdate == nil || item.Status != "update" {
			continue
		}
		if len(selected) > 0 && !selected[item.Local.Filename] {
			continue
		}
		cand := *item.SafeUpdate
		if !cand.Safe || cand.URL == "" {
			continue
		}
		stage := filepath.Join(stageRoot, safeFilename(cand.Filename))
		if err := a.downloadURLVerified(ctx, cand.URL, stage, cand.Size, cand.Hashes); err != nil {
			rollback()
			return nil, fmt.Errorf("%s download/verification failed: %w", item.Local.Filename, err)
		}
		meta, err := parseJarMetadataPath(stage)
		if err != nil {
			rollback()
			return nil, fmt.Errorf("%s update is not a readable mod JAR: %w", item.Local.Filename, err)
		}
		if item.Local.Metadata.ModID != "" && meta.ModID != "" && !strings.EqualFold(item.Local.Metadata.ModID, meta.ModID) {
			rollback()
			return nil, fmt.Errorf("%s update failed identity validation: expected mod id %s, downloaded %s", item.Local.Filename, item.Local.Metadata.ModID, meta.ModID)
		}
		if plan.Loader != "" && plan.Loader != "any" && plan.Loader != "vanilla" && len(meta.Loaders) > 0 && !containsFold(meta.Loaders, plan.Loader) {
			rollback()
			return nil, fmt.Errorf("%s update failed loader validation: target is %s, downloaded JAR declares %s", item.Local.Filename, plan.Loader, strings.Join(meta.Loaders, ", "))
		}
		// For Modrinth, stage any newly required dependencies before touching the original.
		if cand.Provider == "modrinth" && cand.VersionID != "" {
			version, err := a.fetchModrinthVersion(ctx, cand.VersionID)
			if err != nil {
				rollback()
				return nil, err
			}
			for _, dep := range version.Dependencies {
				if dep.DependencyType != "required" || dep.ProjectID == "" {
					continue
				}
				installed := []InstalledProject{}
				if err := a.installModrinthProject(ctx, dep.ProjectID, dep.VersionID, plan.GameVersion, plan.Loader, "mods", true, map[string]bool{}, &installed); err != nil {
					rollback()
					return nil, fmt.Errorf("dependency update for %s failed: %w", item.Local.Filename, err)
				}
				for _, ins := range installed {
					installedDeps = append(installedDeps, ins.Path)
				}
			}
		}
		oldPath := item.Local.Path
		backupPath := filepath.Join(backupRoot, item.Local.Filename)
		if err := os.Rename(oldPath, backupPath); err != nil {
			rollback()
			return nil, fmt.Errorf("backup failed for %s: %w", item.Local.Filename, err)
		}
		newName := safeFilename(cand.Filename)
		if !item.Local.Enabled {
			newName += ".disabled"
		}
		newPath := filepath.Join(plan.ModsDirectory, newName)
		if pathExists(newPath) {
			_ = os.Rename(backupPath, oldPath)
			rollback()
			return nil, fmt.Errorf("update destination already exists: %s", newName)
		}
		if err := os.Rename(stage, newPath); err != nil {
			_ = os.Rename(backupPath, oldPath)
			rollback()
			return nil, fmt.Errorf("atomic install failed for %s: %w", item.Local.Filename, err)
		}
		applied = append(applied, appliedOp{oldPath: oldPath, backupPath: backupPath, newPath: newPath})
		updated = append(updated, map[string]string{"from": item.Local.Filename, "to": filepath.Base(newPath), "provider": cand.Provider, "version": cand.Version})
	}
	_ = os.RemoveAll(stageRoot)
	return map[string]any{"ok": true, "updated": updated, "count": len(updated), "backupDir": backupRoot, "dependenciesInstalled": len(installedDeps)}, nil
}

func parseJarMetadataPath(path string) (JarMetadata, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return JarMetadata{}, err
	}
	return parseJarMetadataBytes(b)
}

func (a *App) downloadURLVerified(ctx context.Context, rawURL, dst string, expectedSize int64, hashes map[string]string) error {
	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme != "https" {
		return errors.New("download URL is not HTTPS")
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	req.Header.Set("User-Agent", userAgent)
	res, err := a.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("download returned %s", res.Status)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	hMD5, h1, h256, h512 := md5.New(), sha1.New(), sha256.New(), sha512.New()
	n, copyErr := io.Copy(io.MultiWriter(out, hMD5, h1, h256, h512), io.LimitReader(res.Body, 2<<30))
	closeErr := out.Close()
	if copyErr != nil {
		_ = os.Remove(dst)
		return copyErr
	}
	if closeErr != nil {
		_ = os.Remove(dst)
		return closeErr
	}
	if expectedSize > 0 && n != expectedSize {
		_ = os.Remove(dst)
		return fmt.Errorf("download size mismatch: expected %d got %d", expectedSize, n)
	}
	checks := map[string]string{"md5": hex.EncodeToString(hMD5.Sum(nil)), "sha1": hex.EncodeToString(h1.Sum(nil)), "sha256": hex.EncodeToString(h256.Sum(nil)), "sha512": hex.EncodeToString(h512.Sum(nil))}
	verified := false
	for _, algo := range []string{"sha512", "sha256", "sha1", "md5"} {
		want := strings.ToLower(strings.TrimSpace(hashes[algo]))
		if want == "" {
			continue
		}
		verified = true
		if checks[algo] != want {
			_ = os.Remove(dst)
			return fmt.Errorf("%s verification failed", strings.ToUpper(algo))
		}
		break
	}
	if len(hashes) > 0 && !verified {
		_ = os.Remove(dst)
		return errors.New("provider supplied hashes, but none use a supported algorithm")
	}
	return nil
}
