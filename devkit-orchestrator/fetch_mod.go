package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type FetchReceipt struct {
	Schema       int       `json:"schema"`
	Provider     string    `json:"provider"`
	ProjectID    string    `json:"projectId,omitempty"`
	VersionID    string    `json:"versionId,omitempty"`
	Version      string    `json:"version"`
	Filename     string    `json:"filename"`
	Output       string    `json:"output"`
	Size         int64     `json:"size"`
	SHA256       string    `json:"sha256"`
	SHA512       string    `json:"sha512,omitempty"`
	SHA1         string    `json:"sha1,omitempty"`
	SourceURL    string    `json:"sourceUrl"`
	ResolvedURL  string    `json:"resolvedUrl"`
	PageURL      string    `json:"pageUrl,omitempty"`
	Attempts     int       `json:"attempts"`
	Resumed      bool      `json:"resumed"`
	FallbackUsed bool      `json:"fallbackUsed"`
	VerifiedAt   time.Time `json:"verifiedAt"`
}

func (p *providerClient) resolveExactModrinth(ctx context.Context, project, version string) (Candidate, error) {
	project = strings.TrimSpace(project)
	version = strings.TrimSpace(version)
	if project == "" || version == "" {
		return Candidate{}, fmt.Errorf("exact Modrinth fetch requires project and version")
	}
	var proj struct {
		ID        string `json:"id"`
		Slug      string `json:"slug"`
		SourceURL string `json:"source_url"`
	}
	if err := p.request(ctx, "GET", p.modrinth+"/project/"+url.PathEscape(project), nil, &proj); err != nil {
		return Candidate{}, err
	}
	if proj.ID == "" {
		return Candidate{}, fmt.Errorf("Modrinth project %s resolved without canonical id", project)
	}

	var v mrVersion
	// Version IDs are globally unique. If this fails (for a human version number), use the
	// project-scoped endpoint so exact semantic versions work too.
	err := p.request(ctx, "GET", p.modrinth+"/version/"+url.PathEscape(version), nil, &v)
	if err != nil {
		err = p.request(ctx, "GET", p.modrinth+"/project/"+url.PathEscape(proj.ID)+"/version/"+url.PathEscape(version), nil, &v)
	}
	if err != nil {
		return Candidate{}, err
	}
	if v.ProjectID == "" || !strings.EqualFold(v.ProjectID, proj.ID) {
		return Candidate{}, fmt.Errorf("Modrinth provenance mismatch: requested project %s resolved to %s but version %s belongs to %s", project, proj.ID, version, v.ProjectID)
	}
	if len(v.Files) == 0 {
		return Candidate{}, fmt.Errorf("Modrinth version %s has no downloadable files", version)
	}
	f := v.Files[0]
	for _, x := range v.Files {
		if x.Primary {
			f = x
			break
		}
	}
	deps := []Dependency{}
	for _, d := range v.Dependencies {
		if d.DependencyType == "required" {
			deps = append(deps, Dependency{Provider: "modrinth", ProjectID: d.ProjectID, VersionID: d.VersionID, Required: true})
		}
	}
	c := Candidate{
		Provider: "modrinth", ProjectID: proj.ID, VersionID: v.ID, Version: v.VersionNumber,
		Filename: f.Filename, URL: f.URL, PageURL: "https://modrinth.com/project/" + proj.Slug + "/version/" + v.ID,
		Published: v.DatePublished, SHA256: f.Hashes["sha256"], SHA512: f.Hashes["sha512"], SHA1: f.Hashes["sha1"],
		Size: f.Size, GameVersions: v.GameVersions, Loaders: v.Loaders, Dependencies: deps,
		SourceURL: proj.SourceURL, ReleaseChannel: v.VersionType,
	}
	c.AlternateURLs = appendUniqueURL(c.AlternateURLs, modrinthCanonicalCDNURL(c))
	c.AlternateURLs = appendUniqueURL(c.AlternateURLs, modrinthMavenURL(c))
	return c, nil
}

func (p *providerClient) resolveExactCurseForge(ctx context.Context, project, fileID string) (Candidate, error) {
	if p.cfKey == "" {
		return Candidate{}, fmt.Errorf("CurseForge API key unavailable")
	}
	project = strings.TrimSpace(project)
	fileID = strings.TrimSpace(fileID)
	if project == "" || fileID == "" {
		return Candidate{}, fmt.Errorf("exact CurseForge fetch requires project and file id")
	}
	var out struct {
		Data cfFile `json:"data"`
	}
	raw := p.curseforge + "/mods/" + url.PathEscape(project) + "/files/" + url.PathEscape(fileID)
	if err := p.request(ctx, "GET", raw, map[string]string{"x-api-key": p.cfKey}, &out); err != nil {
		return Candidate{}, err
	}
	f := out.Data
	if strconv.FormatInt(f.ID, 10) != fileID {
		return Candidate{}, fmt.Errorf("CurseForge provenance mismatch: requested file %s but provider returned %d", fileID, f.ID)
	}
	if f.DownloadURL == "" {
		var dl struct {
			Data string `json:"data"`
		}
		if err := p.request(ctx, "GET", raw+"/download-url", map[string]string{"x-api-key": p.cfKey}, &dl); err == nil {
			f.DownloadURL = strings.TrimSpace(dl.Data)
		}
	}
	if f.DownloadURL == "" {
		f.DownloadURL = curseForgeCDNURL(fileID, f.FileName)
	}
	if f.DownloadURL == "" {
		return Candidate{}, fmt.Errorf("CurseForge file %s has no resolvable download URL", fileID)
	}
	hashes := map[int]string{}
	for _, h := range f.Hashes {
		hashes[h.Algo] = h.Value
	}
	c := Candidate{Provider: "curseforge", ProjectID: project, VersionID: fileID, Version: first(f.DisplayName, f.FileName), Filename: f.FileName, URL: f.DownloadURL, Published: f.FileDate, SHA1: hashes[1], Size: f.FileLength, GameVersions: f.GameVersions, ReleaseChannel: cfReleaseName(f.ReleaseType)}
	c.AlternateURLs = appendUniqueURL(c.AlternateURLs, curseForgeCDNURL(fileID, f.FileName))
	return c, nil
}

func atomicWriteBytes(path string, data []byte, mode os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".mmv-fetch-*.tmp")
	if err != nil {
		return err
	}
	tmp := f.Name()
	ok := false
	defer func() {
		_ = f.Close()
		if !ok {
			_ = os.Remove(tmp)
		}
	}()
	if _, err = f.Write(data); err != nil {
		return err
	}
	if err = f.Sync(); err != nil {
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err = os.Chmod(tmp, mode); err != nil {
		return err
	}
	if err = os.Rename(tmp, path); err != nil {
		return err
	}
	ok = true
	return nil
}

func fetchCandidateToPath(ctx context.Context, eng *engine, c Candidate, out string) (FetchReceipt, error) {
	var err error
	if strings.TrimSpace(out) == "" {
		out = c.Filename
	}
	if info, statErr := os.Stat(out); statErr == nil && info.IsDir() {
		out = filepath.Join(out, c.Filename)
	}
	out, err = filepath.Abs(out)
	if err != nil {
		return FetchReceipt{}, err
	}
	result, err := eng.fetchVerifiedToPath(ctx, c, out)
	if err != nil {
		return FetchReceipt{}, err
	}
	return FetchReceipt{
		Schema: 2, Provider: c.Provider, ProjectID: c.ProjectID, VersionID: c.VersionID,
		Version: c.Version, Filename: c.Filename, Output: out, Size: result.Size,
		SHA256: result.SHA256, SHA512: result.SHA512, SHA1: result.SHA1,
		SourceURL: c.URL, ResolvedURL: result.ResolvedURL, PageURL: c.PageURL,
		Attempts: result.Attempts, Resumed: result.Resumed, FallbackUsed: result.FallbackUsed,
		VerifiedAt: time.Now().UTC(),
	}, nil
}

func writeFetchReceipt(path string, r FetchReceipt) error {
	if strings.TrimSpace(path) == "" {
		return nil
	}
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return atomicWriteBytes(path, append(b, '\n'), 0644)
}
