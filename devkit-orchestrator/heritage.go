package main

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type ReleaseNote struct {
	Provider     string    `json:"provider"`
	VersionID    string    `json:"versionId,omitempty"`
	Version      string    `json:"version"`
	Published    time.Time `json:"published,omitempty"`
	GameVersions []string  `json:"gameVersions,omitempty"`
	Loaders      []string  `json:"loaders,omitempty"`
	PageURL      string    `json:"pageUrl,omitempty"`
	Changelog    string    `json:"changelog,omitempty"`
}

type ContentEntry struct {
	Path     string   `json:"path"`
	Category string   `json:"category"`
	SHA256   string   `json:"sha256"`
	Size     int64    `json:"size"`
	Strict   bool     `json:"strict"`
	Portable bool     `json:"portable"`
	Symbols  []string `json:"symbols,omitempty"`
}

type ContentSnapshot struct {
	Label      string         `json:"label"`
	Source     string         `json:"source,omitempty"`
	SHA256     string         `json:"sha256,omitempty"`
	EntryCount int            `json:"entryCount"`
	Categories map[string]int `json:"categories"`
	Entries    []ContentEntry `json:"entries"`
}

type ContentChange struct {
	Path       string `json:"path"`
	Category   string `json:"category"`
	BeforeHash string `json:"beforeHash,omitempty"`
	AfterHash  string `json:"afterHash,omitempty"`
}

type ContentDelta struct {
	Added   []ContentEntry  `json:"added,omitempty"`
	Changed []ContentChange `json:"changed,omitempty"`
	Removed []ContentEntry  `json:"removed,omitempty"`
}

type DependencyDelta struct {
	AddedRequired   []Dependency `json:"addedRequired,omitempty"`
	RemovedRequired []Dependency `json:"removedRequired,omitempty"`
}

type HeritageReport struct {
	Schema                int              `json:"schema"`
	CreatedAt             time.Time        `json:"createdAt"`
	ArtifactID            string           `json:"artifactId"`
	Target                Target           `json:"target"`
	Provider              string           `json:"provider"`
	CompatibilityProvider string           `json:"compatibilityProvider,omitempty"`
	CompatibilityMode     string           `json:"compatibilityMode"`
	TargetCompatibleExact bool             `json:"targetCompatibleExact"`
	TargetCompatible      Candidate        `json:"targetCompatible"`
	LatestUpstream        Candidate        `json:"latestUpstream"`
	ReleaseLineage        []ReleaseNote    `json:"releaseLineage,omitempty"`
	RuntimeCompatible     ContentSnapshot  `json:"runtimeCompatible"`
	RuntimeLatest         ContentSnapshot  `json:"runtimeLatest"`
	RuntimeDelta          ContentDelta     `json:"runtimeDelta"`
	SourceCompatible      *ContentSnapshot `json:"sourceCompatible,omitempty"`
	SourceLatest          *ContentSnapshot `json:"sourceLatest,omitempty"`
	SourceDelta           *ContentDelta    `json:"sourceDelta,omitempty"`
	DependencyDelta       DependencyDelta  `json:"dependencyDelta"`
	DownloadDirectory     string           `json:"downloadDirectory,omitempty"`
}

type AuditFinding struct {
	Severity string `json:"severity"`
	Kind     string `json:"kind"`
	Path     string `json:"path,omitempty"`
	Category string `json:"category,omitempty"`
	Message  string `json:"message"`
	Expected string `json:"expectedSha256,omitempty"`
	Actual   string `json:"actualSha256,omitempty"`
}

type PortGuardReport struct {
	Schema          int              `json:"schema"`
	CreatedAt       time.Time        `json:"createdAt"`
	ArtifactID      string           `json:"artifactId"`
	Target          Target           `json:"target"`
	Original        ContentSnapshot  `json:"original"`
	Converted       ContentSnapshot  `json:"converted"`
	ConvertedSource *ContentSnapshot `json:"convertedSource,omitempty"`
	Heritage        HeritageReport   `json:"heritage"`
	Findings        []AuditFinding   `json:"findings"`
	Errors          int              `json:"errors"`
	Warnings        int              `json:"warnings"`
	Passed          bool             `json:"passed"`
}

func (e *engine) buildHeritage(ctx context.Context, artifactID string, override Target, outDir string) (HeritageReport, error) {
	var a *ManagedArtifact
	for i := range e.reg.Artifacts {
		if strings.EqualFold(e.reg.Artifacts[i].ID, artifactID) {
			a = &e.reg.Artifacts[i]
			break
		}
	}
	if a == nil {
		return HeritageReport{}, fmt.Errorf("unknown artifact %s", artifactID)
	}
	refs := append([]ProviderRef(nil), a.Providers...)
	sort.SliceStable(refs, func(i, j int) bool { return refs[i].Priority > refs[j].Priority })
	target := e.target(*a, override)
	var compatible, latest Candidate
	var compatibilityRef, latestRef ProviderRef
	compatMode := ""
	var errs []string
	for _, ref := range refs {
		lt := target
		lt.Minecraft = ""
		lt.Loader = ""
		c, err := e.p.resolve(ctx, ref, lt)
		if err != nil {
			errs = append(errs, ref.Type+" latest: "+err.Error())
			continue
		}
		if latest.URL == "" || candidateNewer(c, latest) {
			latest, latestRef = c, ref
		}
	}
	if latest.URL == "" {
		return HeritageReport{}, fmt.Errorf("unable to resolve newest upstream feature authority: %s", strings.Join(errs, "; "))
	}
	for _, ref := range refs {
		c, err := e.p.resolve(ctx, ref, target)
		if err == nil {
			compatible, compatibilityRef, compatMode = c, ref, "exact"
			break
		}
		errs = append(errs, ref.Type+" compatible: "+err.Error())
	}
	if compatible.URL == "" && target.Minecraft != "" {
		for _, ref := range refs {
			t := target
			t.Loader = ""
			if c, err := e.p.resolve(ctx, ref, t); err == nil {
				compatible, compatibilityRef, compatMode = c, ref, "minecraft-only"
				break
			}
		}
	}
	if compatible.URL == "" && target.Loader != "" {
		for _, ref := range refs {
			t := target
			t.Minecraft = ""
			if c, err := e.p.resolve(ctx, ref, t); err == nil {
				compatible, compatibilityRef, compatMode = c, ref, "loader-only"
				break
			}
		}
	}
	if compatible.URL == "" {
		compatible, compatibilityRef, compatMode = latest, latestRef, "latest-fallback"
	}
	if outDir == "" {
		outDir = filepath.Join(".mmv", "heritage", safeName(artifactID))
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return HeritageReport{}, err
	}
	cdata, _, err := e.download(ctx, compatible)
	if err != nil {
		return HeritageReport{}, fmt.Errorf("compatible download: %w", err)
	}
	ldata, _, err := e.download(ctx, latest)
	if err != nil {
		return HeritageReport{}, fmt.Errorf("latest download: %w", err)
	}
	cpath := filepath.Join(outDir, "target-compatible-"+safeName(compatible.Filename))
	lpath := filepath.Join(outDir, "latest-upstream-"+safeName(latest.Filename))
	if err := os.WriteFile(cpath, cdata, 0644); err != nil {
		return HeritageReport{}, err
	}
	if err := os.WriteFile(lpath, ldata, 0644); err != nil {
		return HeritageReport{}, err
	}
	cs, err := snapshotBytes("target-compatible", cpath, compatible.Filename, cdata)
	if err != nil {
		return HeritageReport{}, fmt.Errorf("compatible artifact invalid/corrupt: %w", err)
	}
	ls, err := snapshotBytes("latest-upstream", lpath, latest.Filename, ldata)
	if err != nil {
		return HeritageReport{}, fmt.Errorf("latest artifact invalid/corrupt: %w", err)
	}
	rep := HeritageReport{Schema: 1, CreatedAt: time.Now().UTC(), ArtifactID: a.ID, Target: target, Provider: latestRef.Type, CompatibilityProvider: compatibilityRef.Type, CompatibilityMode: compatMode, TargetCompatibleExact: compatMode == "exact", TargetCompatible: compatible, LatestUpstream: latest, RuntimeCompatible: cs, RuntimeLatest: ls, RuntimeDelta: diffSnapshots(cs, ls), DependencyDelta: diffDependencies(compatible.Dependencies, latest.Dependencies), DownloadDirectory: outDir}
	if history, herr := e.p.releaseLineage(ctx, latestRef, compatible, latest, target); herr == nil {
		rep.ReleaseLineage = history
	}
	csrcURL, csrcRef, cerr := e.p.sourceArchiveForCandidate(ctx, compatible)
	lsrcURL, lsrcRef, lerr := e.p.sourceArchiveForCandidate(ctx, latest)
	if cerr == nil && lerr == nil && csrcURL != "" && lsrcURL != "" {
		csrc := compatible
		csrc.URL = csrcURL
		csrc.SHA256 = ""
		csrc.SHA1 = ""
		csrc.Filename = "source-" + safeName(csrcRef) + ".zip"
		lsrc := latest
		lsrc.URL = lsrcURL
		lsrc.SHA256 = ""
		lsrc.SHA1 = ""
		lsrc.Filename = "source-" + safeName(lsrcRef) + ".zip"
		cb, _, ce := e.download(ctx, csrc)
		lb, _, le := e.download(ctx, lsrc)
		if ce == nil && le == nil {
			cp := filepath.Join(outDir, "source-compatible-"+safeName(csrcRef)+".zip")
			lp := filepath.Join(outDir, "source-latest-"+safeName(lsrcRef)+".zip")
			_ = os.WriteFile(cp, cb, 0644)
			_ = os.WriteFile(lp, lb, 0644)
			ss1, se1 := snapshotBytes("source-compatible", cp, csrc.Filename, cb)
			ss2, se2 := snapshotBytes("source-latest", lp, lsrc.Filename, lb)
			if se1 == nil && se2 == nil {
				rep.SourceCompatible = &ss1
				rep.SourceLatest = &ss2
				d := diffSnapshots(ss1, ss2)
				rep.SourceDelta = &d
			}
		}
	}
	return rep, nil
}

func candidateNewer(a, b Candidate) bool {
	if !a.Published.IsZero() && !b.Published.IsZero() && !a.Published.Equal(b.Published) {
		return a.Published.After(b.Published)
	}
	return compareVersions(a.Version, b.Version) > 0
}

func (e *engine) portGuard(ctx context.Context, artifactID string, override Target, originalPath, convertedPath, convertedSourcePath, outDir string) (PortGuardReport, error) {
	if originalPath == "" || convertedPath == "" {
		return PortGuardReport{}, fmt.Errorf("original and converted paths are required")
	}
	original, err := snapshotPath("original", originalPath)
	if err != nil {
		return PortGuardReport{}, fmt.Errorf("original validation failed: %w", err)
	}
	converted, err := snapshotPath("converted", convertedPath)
	if err != nil {
		return PortGuardReport{}, fmt.Errorf("converted validation failed: %w", err)
	}
	heritage, err := e.buildHeritage(ctx, artifactID, override, outDir)
	if err != nil {
		return PortGuardReport{}, err
	}
	rep := PortGuardReport{Schema: 1, CreatedAt: time.Now().UTC(), ArtifactID: artifactID, Target: heritage.Target, Original: original, Converted: converted, Heritage: heritage}
	if convertedSourcePath != "" {
		src, se := snapshotPath("converted-source", convertedSourcePath)
		if se != nil {
			return rep, fmt.Errorf("converted source validation failed: %w", se)
		}
		rep.ConvertedSource = &src
	}
	rep.Findings = auditPort(original, converted, rep.ConvertedSource, heritage)
	for _, f := range rep.Findings {
		if f.Severity == "error" {
			rep.Errors++
		}
		if f.Severity == "warning" {
			rep.Warnings++
		}
	}
	rep.Passed = rep.Errors == 0
	return rep, nil
}

func snapshotPath(label, path string) (ContentSnapshot, error) {
	st, err := os.Stat(path)
	if err != nil {
		return ContentSnapshot{}, err
	}
	if st.IsDir() {
		return snapshotDirectory(label, path)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return ContentSnapshot{}, err
	}
	return snapshotBytes(label, path, filepath.Base(path), b)
}

func snapshotDirectory(label, root string) (ContentSnapshot, error) {
	out := ContentSnapshot{Label: label, Source: root, Categories: map[string]int{}}
	h := sha256.New()
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		ent := makeEntry(rel, b)
		out.Entries = append(out.Entries, ent)
		out.Categories[ent.Category]++
		_, _ = h.Write([]byte(rel))
		_, _ = h.Write(b)
		return nil
	})
	if err != nil {
		return out, err
	}
	sortEntries(out.Entries)
	out.EntryCount = len(out.Entries)
	out.SHA256 = hex.EncodeToString(h.Sum(nil))
	return out, nil
}

func snapshotBytes(label, source, filename string, data []byte) (ContentSnapshot, error) {
	sum := sha256.Sum256(data)
	out := ContentSnapshot{Label: label, Source: source, SHA256: hex.EncodeToString(sum[:]), Categories: map[string]int{}}
	low := strings.ToLower(filename)
	if !strings.HasSuffix(low, ".jar") && !strings.HasSuffix(low, ".zip") {
		ent := makeEntry(filename, data)
		out.Entries = []ContentEntry{ent}
		out.EntryCount = 1
		out.Categories[ent.Category] = 1
		return out, nil
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return out, err
	}
	names := make([]string, 0, len(zr.File))
	seen := map[string]bool{}
	for _, f := range zr.File {
		if !f.FileInfo().IsDir() {
			names = append(names, filepath.ToSlash(f.Name))
		}
	}
	prefix := commonArchiveRoot(names)
	if strings.HasSuffix(low, ".jar") {
		prefix = ""
	}
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		rawName := filepath.ToSlash(f.Name)
		if unsafeArchivePath(rawName) {
			return out, fmt.Errorf("unsafe archive entry %s", rawName)
		}
		n := normalizeArchivePath(rawName, prefix)
		if n == "" {
			continue
		}
		key := strings.ToLower(n)
		if seen[key] {
			return out, fmt.Errorf("duplicate archive entry %s", n)
		}
		seen[key] = true
		rc, err := f.Open()
		if err != nil {
			return out, err
		}
		b, err := io.ReadAll(io.LimitReader(rc, 256<<20))
		closeErr := rc.Close()
		if err != nil {
			return out, err
		}
		if closeErr != nil {
			return out, closeErr
		}
		ent := makeEntry(n, b)
		out.Entries = append(out.Entries, ent)
		out.Categories[ent.Category]++
	}
	sortEntries(out.Entries)
	out.EntryCount = len(out.Entries)
	return out, nil
}

func commonArchiveRoot(names []string) string {
	if len(names) == 0 {
		return ""
	}
	var root string
	for _, n := range names {
		p := strings.Split(strings.TrimPrefix(n, "/"), "/")
		if len(p) < 2 {
			return ""
		}
		if root == "" {
			root = p[0]
		} else if root != p[0] {
			return ""
		}
	}
	return root + "/"
}
func unsafeArchivePath(n string) bool {
	n = filepath.ToSlash(n)
	clean := filepath.ToSlash(filepath.Clean(n))
	return strings.HasPrefix(n, "/") || clean == ".." || strings.HasPrefix(clean, "../") || strings.Contains(clean, ":/")
}

func normalizeArchivePath(n, prefix string) string {
	n = strings.TrimPrefix(n, "/")
	if prefix != "" {
		n = strings.TrimPrefix(n, prefix)
	}
	return strings.TrimPrefix(filepath.ToSlash(filepath.Clean(n)), "./")
}
func makeEntry(path string, b []byte) ContentEntry {
	s := sha256.Sum256(b)
	c := contentCategory(path)
	return ContentEntry{Path: path, Category: c, SHA256: hex.EncodeToString(s[:]), Size: int64(len(b)), Strict: isStrictContent(path, c), Portable: isPortableContent(path, c), Symbols: sourceSymbols(path, b)}
}
func sortEntries(v []ContentEntry) { sort.Slice(v, func(i, j int) bool { return v[i].Path < v[j].Path }) }

func contentCategory(path string) string {
	p := strings.ToLower(filepath.ToSlash(path))
	switch {
	case strings.HasSuffix(p, ".class"):
		return "code"
	case strings.HasSuffix(p, ".java") || strings.HasSuffix(p, ".kt") || strings.HasSuffix(p, ".scala"):
		return "source"
	case strings.Contains(p, "/textures/"):
		return "textures"
	case strings.Contains(p, "/blockstates/"):
		return "blockstates"
	case strings.Contains(p, "/models/"):
		return "models"
	case strings.Contains(p, "/lang/"):
		return "language"
	case strings.Contains(p, "/sounds/") || strings.HasSuffix(p, "sounds.json"):
		return "sounds"
	case strings.Contains(p, "/recipes/") || strings.Contains(p, "/recipe/"):
		return "recipes"
	case strings.Contains(p, "/loot_tables/") || strings.Contains(p, "/loot_table/"):
		return "loot"
	case strings.Contains(p, "/tags/"):
		return "tags"
	case strings.Contains(p, "/advancements/") || strings.Contains(p, "/advancement/"):
		return "advancements"
	case strings.Contains(p, "/worldgen/"):
		return "worldgen"
	case strings.Contains(p, "/structures/") || strings.Contains(p, "/structure/"):
		return "structures"
	case strings.HasPrefix(p, "assets/"):
		return "assets-other"
	case strings.HasPrefix(p, "data/"):
		return "data-other"
	case p == "fabric.mod.json" || p == "quilt.mod.json" || strings.Contains(p, "mods.toml") || strings.Contains(p, "neoforge.mods.toml"):
		return "metadata"
	case strings.Contains(p, "mixin") && strings.HasSuffix(p, ".json"):
		return "mixins"
	case strings.HasPrefix(p, "meta-inf/"):
		return "metadata"
	default:
		return "other"
	}
}
func isStrictContent(path, cat string) bool {
	p := strings.ToLower(path)
	return strings.HasPrefix(p, "assets/") || strings.HasPrefix(p, "data/") || p == "pack.mcmeta" || cat == "mixins"
}
func isPortableContent(path, cat string) bool {
	switch cat {
	case "textures", "language", "sounds":
		return true
	}
	return false
}

var typeRx = regexp.MustCompile(`(?m)\b(?:class|interface|enum|record|object)\s+([A-Za-z_$][A-Za-z0-9_$]*)`)

func sourceSymbols(path string, b []byte) []string {
	c := contentCategory(path)
	if c != "source" {
		return nil
	}
	m := typeRx.FindAllSubmatch(b, -1)
	seen := map[string]bool{}
	out := []string{}
	for _, x := range m {
		if len(x) > 1 {
			v := string(x[1])
			if !seen[v] {
				seen[v] = true
				out = append(out, v)
			}
		}
	}
	sort.Strings(out)
	return out
}

func entryMap(s ContentSnapshot) map[string]ContentEntry {
	m := map[string]ContentEntry{}
	for _, e := range s.Entries {
		m[strings.ToLower(e.Path)] = e
	}
	return m
}
func diffSnapshots(a, b ContentSnapshot) ContentDelta {
	am, bm := entryMap(a), entryMap(b)
	d := ContentDelta{}
	for k, x := range bm {
		if y, ok := am[k]; !ok {
			d.Added = append(d.Added, x)
		} else if y.SHA256 != x.SHA256 {
			d.Changed = append(d.Changed, ContentChange{Path: x.Path, Category: x.Category, BeforeHash: y.SHA256, AfterHash: x.SHA256})
		}
	}
	for k, x := range am {
		if _, ok := bm[k]; !ok {
			d.Removed = append(d.Removed, x)
		}
	}
	sortEntries(d.Added)
	sortEntries(d.Removed)
	sort.Slice(d.Changed, func(i, j int) bool { return d.Changed[i].Path < d.Changed[j].Path })
	return d
}
func depKey(d Dependency) string { return strings.ToLower(d.Provider + ":" + d.ProjectID) }
func diffDependencies(a, b []Dependency) DependencyDelta {
	am := map[string]Dependency{}
	bm := map[string]Dependency{}
	for _, d := range a {
		if d.Required {
			am[depKey(d)] = d
		}
	}
	for _, d := range b {
		if d.Required {
			bm[depKey(d)] = d
		}
	}
	out := DependencyDelta{}
	for k, d := range bm {
		if _, ok := am[k]; !ok {
			out.AddedRequired = append(out.AddedRequired, d)
		}
	}
	for k, d := range am {
		if _, ok := bm[k]; !ok {
			out.RemovedRequired = append(out.RemovedRequired, d)
		}
	}
	return out
}

func auditPort(original, converted ContentSnapshot, convertedSource *ContentSnapshot, h HeritageReport) []AuditFinding {
	om, cm, lm, tm := entryMap(original), entryMap(converted), entryMap(h.RuntimeLatest), entryMap(h.RuntimeCompatible)
	findings := []AuditFinding{}
	add := func(sev, kind, path, cat, msg, exp, act string) {
		findings = append(findings, AuditFinding{Severity: sev, Kind: kind, Path: path, Category: cat, Message: msg, Expected: exp, Actual: act})
	}
	for k, o := range om {
		if !o.Strict {
			continue
		}
		l, lok := lm[k]
		if !lok {
			add("info", "upstream-removed", o.Path, o.Category, "original content was removed by newest upstream and is not mandatory in the conversion", "", "")
			continue
		}
		c, cok := cm[k]
		if !cok {
			add("error", "original-content-missing", o.Path, o.Category, "content present in the original and still present upstream is missing from the converted target", l.SHA256, "")
			continue
		}
		if o.SHA256 == l.SHA256 && c.SHA256 != o.SHA256 {
			add("error", "unchanged-content-corrupted", o.Path, o.Category, "upstream did not change this original content, but the converted target did", o.SHA256, c.SHA256)
			continue
		}
		if o.SHA256 != l.SHA256 {
			if c.SHA256 == o.SHA256 {
				add("error", "stale-original-content", o.Path, o.Category, "new upstream content exists but the converted target still carries the original bytes", l.SHA256, c.SHA256)
			} else if l.Portable && c.SHA256 != l.SHA256 {
				add("error", "latest-portable-content-mismatch", o.Path, o.Category, "portable upstream content should be copied exactly from the newest release", l.SHA256, c.SHA256)
			} else if c.SHA256 != l.SHA256 {
				add("warning", "adapted-latest-content", o.Path, o.Category, "latest content differs from the conversion; verify this is a required target-version adaptation", l.SHA256, c.SHA256)
			}
		}
	}
	for k, l := range lm {
		if !l.Strict {
			continue
		}
		if _, old := om[k]; old {
			continue
		}
		c, ok := cm[k]
		if !ok {
			add("error", "latest-feature-missing", l.Path, l.Category, "new content exists in the newest upstream release but is absent from the converted target", l.SHA256, "")
			continue
		}
		if l.Portable && c.SHA256 != l.SHA256 {
			add("error", "latest-feature-corrupted", l.Path, l.Category, "new portable content does not match newest upstream bytes", l.SHA256, c.SHA256)
		} else if c.SHA256 != l.SHA256 {
			add("warning", "latest-feature-adapted", l.Path, l.Category, "new upstream content is present but adapted; verify target-version compatibility", l.SHA256, c.SHA256)
		}
	}
	for k, t := range tm {
		if !t.Strict {
			continue
		}
		if l, ok := lm[k]; ok && t.SHA256 != l.SHA256 {
			if c, ok := cm[k]; ok && c.SHA256 == t.SHA256 {
				add("error", "target-version-stale-content", c.Path, c.Category, "converted target matches the older target-compatible release while newer upstream content exists", l.SHA256, c.SHA256)
			}
		}
	}
	if convertedSource != nil && h.SourceCompatible != nil && h.SourceLatest != nil {
		before := entryMap(*h.SourceCompatible)
		latest := entryMap(*h.SourceLatest)
		symbols := map[string]bool{}
		for _, e := range convertedSource.Entries {
			for _, s := range e.Symbols {
				symbols[s] = true
			}
		}
		for k, e := range latest {
			if e.Category != "source" {
				continue
			}
			if _, old := before[k]; old {
				continue
			}
			for _, s := range e.Symbols {
				if !symbols[s] {
					add("error", "latest-source-surface-missing", e.Path, "source", "new upstream source type "+s+" is not represented in the converted source tree", "", "")
				}
			}
		}
	} else if h.SourceDelta != nil && len(h.SourceDelta.Added) > 0 {
		add("error", "converted-source-not-supplied", "", "source", "new upstream source files exist; pass --converted-source so the completion gate can prove their type surface was carried into the conversion", "", "")
	}
	for _, d := range h.DependencyDelta.AddedRequired {
		add("warning", "latest-required-dependency-added", "", "dependency", "new upstream required dependency must be evaluated/backported: "+d.Provider+":"+d.ProjectID, "", "")
	}
	sort.SliceStable(findings, func(i, j int) bool {
		rank := func(s string) int {
			if s == "error" {
				return 0
			}
			if s == "warning" {
				return 1
			}
			return 2
		}
		ri, rj := rank(findings[i].Severity), rank(findings[j].Severity)
		if ri != rj {
			return ri < rj
		}
		return findings[i].Path < findings[j].Path
	})
	return findings
}

func writeReport(path string, v any) error {
	if path == "" {
		return nil
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0644)
}
