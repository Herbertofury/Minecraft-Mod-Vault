package main

import (
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type engine struct {
	reg  Registry
	p    *providerClient
	http *http.Client
}

func newEngine(r Registry, h *http.Client) *engine {
	p := newProviderClient(h)
	p.cfKey = os.Getenv("CURSEFORGE_API_KEY")
	p.ghToken = os.Getenv("GITHUB_TOKEN")
	return &engine{reg: r, p: p, http: p.http}
}
func (e *engine) target(a ManagedArtifact, o Target) Target {
	t := e.reg.Defaults
	mergeTarget(&t, a.Target)
	mergeTarget(&t, o)
	return t
}
func mergeTarget(d *Target, s Target) {
	if s.Minecraft != "" {
		d.Minecraft = s.Minecraft
	}
	if s.Loader != "" {
		d.Loader = s.Loader
	}
	if s.OS != "" {
		d.OS = s.OS
	}
	if s.Arch != "" {
		d.Arch = s.Arch
	}
	if s.Channel != "" {
		d.Channel = s.Channel
	}
}
func (e *engine) plan(ctx context.Context, o Target) (UpdatePlan, error) {
	out := UpdatePlan{Schema: 1, CreatedAt: time.Now().UTC(), Target: o}
	for _, a := range e.reg.Artifacts {
		ap := ArtifactPlan{ArtifactID: a.ID, Current: LocalState{Version: a.Version, Filename: a.Filename, SHA256: a.SHA256, DriveFileID: a.DriveFileID}}
		if a.UpdatePolicy.Pinned {
			ap.Status = "pinned"
			ap.Reason = "artifact is pinned"
			out.Artifacts = append(out.Artifacts, ap)
			continue
		}
		refs := append([]ProviderRef(nil), a.Providers...)
		sort.SliceStable(refs, func(i, j int) bool { return refs[i].Priority > refs[j].Priority })
		var c Candidate
		var errs []string
		for _, r := range refs {
			x, er := e.p.resolve(ctx, r, e.target(a, o))
			if er == nil {
				c = x
				break
			}
			errs = append(errs, r.Type+": "+er.Error())
		}
		if c.URL == "" {
			ap.Status = "unresolved"
			ap.Reason = strings.Join(errs, "; ")
			out.Artifacts = append(out.Artifacts, ap)
			continue
		}
		ap.Candidate = &c
		if !a.UpdatePolicy.AllowDowngrade && a.Version != "" && c.Version != "" && compareVersions(c.Version, a.Version) < 0 {
			ap.Status = "ahead"
			ap.Reason = "recorded version is newer than provider candidate; downgrade blocked"
		} else if sameArtifact(a, c) {
			if a.UpdatePolicy.MirrorSource && c.SourceURL != "" && a.SourceDriveID == "" {
				ap.Status = "source-refresh"
				ap.Reason = "runtime is current but tracked source mirror is missing"
			} else {
				ap.Status = "current"
				ap.Reason = "provider candidate matches recorded version/hash"
			}
		} else {
			ap.Status = "update"
			ap.Reason = "new compatible provider candidate"
		}
		seenDeps := map[string]bool{}
		ap.Deps = e.resolveDepTree(ctx, c.Dependencies, e.target(a, o), seenDeps, 0)
		out.Artifacts = append(out.Artifacts, ap)
	}
	return out, nil
}

func (e *engine) resolveDepTree(ctx context.Context, deps []Dependency, target Target, seen map[string]bool, depth int) []DepPlan {
	if depth >= 12 {
		return nil
	}
	out := []DepPlan{}
	for _, d := range deps {
		if !d.Required || d.ProjectID == "" {
			continue
		}
		key := strings.ToLower(d.Provider + ":" + d.ProjectID)
		if seen[key] {
			continue
		}
		seen[key] = true
		dp := DepPlan{Provider: d.Provider, ProjectID: d.ProjectID, Status: "required"}
		dc, err := e.p.resolve(ctx, ProviderRef{Type: d.Provider, ProjectID: d.ProjectID}, target)
		if err != nil {
			dp.Status = "unresolved"
		} else {
			dp.Candidate = &dc
			dp.Children = e.resolveDepTree(ctx, dc.Dependencies, target, seen, depth+1)
		}
		out = append(out, dp)
	}
	return out
}
func sameArtifact(a ManagedArtifact, c Candidate) bool {
	if a.SHA256 != "" && c.SHA256 != "" && strings.EqualFold(a.SHA256, c.SHA256) {
		return true
	}
	return a.Version != "" && c.Version != "" && a.Version == c.Version
}
func (e *engine) download(ctx context.Context, c Candidate) ([]byte, string, error) {
	req, er := http.NewRequestWithContext(ctx, "GET", c.URL, nil)
	if er != nil {
		return nil, "", er
	}
	req.Header.Set("User-Agent", "MinecraftDevKitOrchestrator/2.0")
	resp, er := e.http.Do(req)
	if er != nil {
		return nil, "", er
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("download http %d", resp.StatusCode)
	}
	data, er := io.ReadAll(io.LimitReader(resp.Body, 2<<30))
	if er != nil {
		return nil, "", er
	}
	sum := sha256.Sum256(data)
	got := hex.EncodeToString(sum[:])
	if c.SHA256 != "" && !strings.EqualFold(c.SHA256, got) {
		return nil, "", fmt.Errorf("sha256 mismatch: provider=%s downloaded=%s", c.SHA256, got)
	}
	if c.SHA1 != "" {
		s1 := sha1.Sum(data)
		got1 := hex.EncodeToString(s1[:])
		if !strings.EqualFold(c.SHA1, got1) {
			return nil, "", fmt.Errorf("sha1 mismatch: provider=%s downloaded=%s", c.SHA1, got1)
		}
	}
	return data, got, nil
}
func (e *engine) apply(ctx context.Context, p UpdatePlan, registryPath, lockPath string, useDrive bool) (LockFile, error) {
	r := e.reg
	idx := map[string]int{}
	for i := range r.Artifacts {
		idx[r.Artifacts[i].ID] = i
	}
	lock := LockFile{Schema: 1, UpdatedAt: time.Now().UTC()}
	var dc *driveClient
	var er error
	if useDrive {
		dc, er = newDriveClient(r.Drive, e.http)
		if er != nil {
			return lock, er
		}
	}
	knownProjects := map[string]bool{}
	for _, ap := range p.Artifacts {
		if ap.Candidate != nil && ap.Candidate.ProjectID != "" {
			knownProjects[strings.ToLower(ap.Candidate.Provider+":"+ap.Candidate.ProjectID)] = true
		}
	}
	for _, ap := range p.Artifacts {
		if (ap.Status != "update" && ap.Status != "source-refresh") || ap.Candidate == nil {
			continue
		}
		i, ok := idx[ap.ArtifactID]
		if !ok {
			return lock, fmt.Errorf("plan artifact missing: %s", ap.ArtifactID)
		}
		a := &r.Artifacts[i]
		if useDrive {
			if er = e.mirrorDependencyTree(ctx, ap.Deps, e.target(*a, p.Target), &r, dc, &lock, knownProjects); er != nil {
				return lock, fmt.Errorf("%s dependencies: %w", a.ID, er)
			}
		}
		driveID := a.DriveFileID
		hash := a.SHA256
		if ap.Status == "update" {
			data, newHash, dlErr := e.download(ctx, *ap.Candidate)
			if dlErr != nil {
				return lock, fmt.Errorf("%s: %w", a.ID, dlErr)
			}
			hash = newHash
			if useDrive {
				if driveID != "" {
					_, er = dc.replace(ctx, driveID, ap.Candidate.Filename, data)
				} else {
					parent := r.Drive.RuntimeFolderID
					if a.Kind == "source" {
						parent = r.Drive.SourceFolderID
					}
					driveID, er = dc.uploadFile(ctx, parent, ap.Candidate.Filename, data)
				}
				if er != nil {
					return lock, fmt.Errorf("%s drive: %w", a.ID, er)
				}
			}
			a.Version = ap.Candidate.Version
			a.Filename = ap.Candidate.Filename
			a.SHA256 = hash
			a.DriveFileID = driveID
		}
		le := LockEntry{ArtifactID: a.ID, Provider: ap.Candidate.Provider, ProjectID: ap.Candidate.ProjectID, VersionID: ap.Candidate.VersionID, Version: ap.Candidate.Version, Filename: ap.Candidate.Filename, SHA256: hash, DriveFileID: driveID, SourceDriveID: a.SourceDriveID, SourceURL: ap.Candidate.SourceURL, SourceRef: ap.Candidate.SourceRef, ResolvedAt: time.Now().UTC()}
		if a.UpdatePolicy.MirrorSource && ap.Candidate.SourceURL != "" {
			srcURL, srcRef, srcErr := e.p.sourceArchiveForCandidate(ctx, *ap.Candidate)
			if srcErr != nil {
				return lock, fmt.Errorf("%s source resolution: %w", a.ID, srcErr)
			}
			src := *ap.Candidate
			src.URL = srcURL
			src.SHA256 = ""
			src.Filename = safeName(a.ID) + "-source-" + safeName(srcRef) + ".zip"
			sdata, _, se := e.download(ctx, src)
			if se != nil {
				return lock, fmt.Errorf("%s source mirror: %w", a.ID, se)
			}
			if useDrive {
				if a.SourceDriveID != "" {
					_, se = dc.replace(ctx, a.SourceDriveID, src.Filename, sdata)
				} else {
					a.SourceDriveID, se = dc.uploadFile(ctx, r.Drive.SourceFolderID, src.Filename, sdata)
				}
				if se != nil {
					return lock, fmt.Errorf("%s source drive: %w", a.ID, se)
				}
				le.SourceDriveID = a.SourceDriveID
			}
		}
		lock.Entries = append(lock.Entries, le)
	}
	if er = writeJSONAtomic(registryPath, r); er != nil {
		return lock, er
	}
	if er = writeJSONAtomic(lockPath, lock); er != nil {
		return lock, er
	}
	e.reg = r
	return lock, nil
}

func (e *engine) mirrorDependencyTree(ctx context.Context, deps []DepPlan, target Target, r *Registry, dc *driveClient, lock *LockFile, known map[string]bool) error {
	for _, dp := range deps {
		if dp.Status == "unresolved" {
			return fmt.Errorf("required dependency %s:%s unresolved", dp.Provider, dp.ProjectID)
		}
		if dp.Candidate == nil {
			continue
		}
		c := *dp.Candidate
		key := strings.ToLower(c.Provider + ":" + c.ProjectID)
		if !known[key] {
			data, hash, err := e.download(ctx, c)
			if err != nil {
				return err
			}
			driveID, err := dc.uploadFile(ctx, r.Drive.RuntimeFolderID, c.Filename, data)
			if err != nil {
				return err
			}
			sourceDriveID := ""
			sourceRef := c.SourceRef
			if c.SourceURL != "" {
				srcURL, resolvedRef, srcErr := e.p.sourceArchiveForCandidate(ctx, c)
				if srcErr != nil {
					return srcErr
				}
				sourceRef = resolvedRef
				src := c
				src.URL = srcURL
				src.SHA256 = ""
				src.SHA1 = ""
				src.Filename = "auto-dependency-" + safeName(c.Provider) + "-" + safeName(c.ProjectID) + "-source-" + safeName(sourceRef) + ".zip"
				srcData, _, srcErr := e.download(ctx, src)
				if srcErr != nil {
					return srcErr
				}
				sourceDriveID, srcErr = dc.uploadFile(ctx, r.Drive.SourceFolderID, src.Filename, srcData)
				if srcErr != nil {
					return srcErr
				}
			}
			id := "auto-dependency/" + safeName(c.Provider) + "/" + safeName(c.ProjectID) + "/" + safeName(target.Minecraft) + "/" + safeName(target.Loader)
			ref := ProviderRef{Type: c.Provider, ProjectID: c.ProjectID, Priority: 100}
			r.Artifacts = append(r.Artifacts, ManagedArtifact{ID: id, Name: c.Filename, Kind: "dependency", Version: c.Version, Filename: c.Filename, SHA256: hash, DriveFileID: driveID, SourceDriveID: sourceDriveID, Target: target, Providers: []ProviderRef{ref}, UpdatePolicy: UpdatePolicy{MirrorSource: true}})
			lock.Entries = append(lock.Entries, LockEntry{ArtifactID: id, Provider: c.Provider, ProjectID: c.ProjectID, VersionID: c.VersionID, Version: c.Version, Filename: c.Filename, SHA256: hash, DriveFileID: driveID, SourceDriveID: sourceDriveID, SourceURL: c.SourceURL, SourceRef: sourceRef, ResolvedAt: time.Now().UTC()})
			known[key] = true
		}
		if err := e.mirrorDependencyTree(ctx, dp.Children, target, r, dc, lock, known); err != nil {
			return err
		}
	}
	return nil
}
func writeJSONAtomic(path string, v any) error {
	b, er := json.MarshalIndent(v, "", "  ")
	if er != nil {
		return er
	}
	if er = os.MkdirAll(filepath.Dir(path), 0755); er != nil {
		return er
	}
	tmp := path + ".tmp"
	if er = os.WriteFile(tmp, append(b, '\n'), 0644); er != nil {
		return er
	}
	return os.Rename(tmp, path)
}
func safeName(s string) string {
	if strings.TrimSpace(s) == "" {
		return "current"
	}
	return strings.NewReplacer("/", "-", "\\", "-", " ", "-").Replace(s)
}

func compareVersions(a, b string) int {
	re := regexp.MustCompile(`[0-9]+`)
	ax, bx := re.FindAllString(a, -1), re.FindAllString(b, -1)
	n := len(ax)
	if len(bx) > n {
		n = len(bx)
	}
	for i := 0; i < n; i++ {
		av, bv := 0, 0
		if i < len(ax) {
			av, _ = strconv.Atoi(ax[i])
		}
		if i < len(bx) {
			bv, _ = strconv.Atoi(bx[i])
		}
		if av < bv {
			return -1
		}
		if av > bv {
			return 1
		}
	}
	return 0
}
func validateRegistry(r Registry) error {
	if r.Schema <= 0 {
		return errors.New("registry schema required")
	}
	seen := map[string]bool{}
	for _, a := range r.Artifacts {
		if a.ID == "" {
			return errors.New("artifact id required")
		}
		if seen[a.ID] {
			return fmt.Errorf("duplicate artifact id %s", a.ID)
		}
		seen[a.ID] = true
		if len(a.Providers) == 0 {
			return fmt.Errorf("%s has no providers", a.ID)
		}
	}
	return nil
}
