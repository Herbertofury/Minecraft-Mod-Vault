package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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
	// The feature authority is the newest release across every canonical hub attached
	// to the artifact. Provider priority breaks ties; publication time/version wins freshness.
	for _, ref := range refs {
		lt := target
		lt.Minecraft = ""
		lt.Loader = ""
		c, err := e.p.resolveHeritage(ctx, ref, lt)
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
	// Prefer an exact target release from any hub. If upstream never shipped one,
	// keep moving with progressively weaker compatibility references rather than
	// losing the latest feature delta entirely.
	for _, ref := range refs {
		c, err := e.p.resolveHeritage(ctx, ref, target)
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
			if c, err := e.p.resolveHeritage(ctx, ref, t); err == nil {
				compatible, compatibilityRef, compatMode = c, ref, "minecraft-only"
				break
			}
		}
	}
	if compatible.URL == "" && target.Loader != "" {
		for _, ref := range refs {
			t := target
			t.Minecraft = ""
			if c, err := e.p.resolveHeritage(ctx, ref, t); err == nil {
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
