package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type providerClient struct {
	http                                         *http.Client
	modrinth, curseforge, github, cfKey, ghToken string
}

func newProviderClient(h *http.Client) *providerClient {
	if h == nil {
		tr := http.DefaultTransport.(*http.Transport).Clone()
		tr.ResponseHeaderTimeout = 30 * time.Second
		tr.TLSHandshakeTimeout = 15 * time.Second
		tr.ExpectContinueTimeout = 2 * time.Second
		tr.IdleConnTimeout = 90 * time.Second
		tr.MaxIdleConns = 32
		tr.MaxIdleConnsPerHost = 8
		h = &http.Client{Transport: tr}
	}
	return &providerClient{http: h, modrinth: "https://api.modrinth.com/v2", curseforge: "https://api.curseforge.com/v1", github: "https://api.github.com"}
}
func (p *providerClient) request(ctx context.Context, method, raw string, headers map[string]string, out any) error {
	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, raw, nil)
		if err != nil {
			return err
		}
		req.Header.Set("User-Agent", devKitUserAgent())
		req.Header.Set("Accept", "application/json")
		for k, v := range headers {
			if v != "" {
				req.Header.Set(k, v)
			}
		}
		resp, err := p.http.Do(req)
		if err != nil {
			lastErr = err
			if attempt < 3 {
				time.Sleep(retryDelay(nil, attempt))
				continue
			}
			break
		}
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
		_ = resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			if attempt < 3 {
				time.Sleep(retryDelay(resp, attempt))
				continue
			}
			break
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			lastErr = fmt.Errorf("%s: http %d: %s", raw, resp.StatusCode, strings.TrimSpace(string(body)))
			if shouldRetryStatus(resp.StatusCode) && attempt < 3 {
				time.Sleep(retryDelay(resp, attempt))
				continue
			}
			break
		}
		if err := json.Unmarshal(body, out); err != nil {
			return err
		}
		return nil
	}
	return lastErr
}
func (p *providerClient) resolve(ctx context.Context, ref ProviderRef, target Target) (Candidate, error) {
	switch strings.ToLower(ref.Type) {
	case "modrinth":
		return p.resolveModrinth(ctx, ref, target)
	case "curseforge":
		return p.resolveCurseForge(ctx, ref, target)
	case "github":
		return p.resolveGitHub(ctx, ref, target)
	case "github-branch":
		return p.resolveGitHubBranch(ctx, ref, target)
	case "url":
		if ref.URL == "" {
			return Candidate{}, fmt.Errorf("url provider missing url")
		}
		return Candidate{Provider: "url", Version: "unversioned", Filename: filenameFromURL(ref.URL), URL: ref.URL}, nil
	default:
		return Candidate{}, fmt.Errorf("unsupported provider %q", ref.Type)
	}
}

type mrVersion struct {
	ID            string    `json:"id"`
	ProjectID     string    `json:"project_id"`
	VersionNumber string    `json:"version_number"`
	VersionType   string    `json:"version_type"`
	DatePublished time.Time `json:"date_published"`
	GameVersions  []string  `json:"game_versions"`
	Loaders       []string  `json:"loaders"`
	Files         []struct {
		Hashes   map[string]string `json:"hashes"`
		URL      string            `json:"url"`
		Filename string            `json:"filename"`
		Primary  bool              `json:"primary"`
		Size     int64             `json:"size"`
	} `json:"files"`
	Dependencies []struct {
		VersionID      string `json:"version_id"`
		ProjectID      string `json:"project_id"`
		DependencyType string `json:"dependency_type"`
	} `json:"dependencies"`
}

func (p *providerClient) resolveModrinth(ctx context.Context, ref ProviderRef, target Target) (Candidate, error) {
	id := first(ref.ProjectID, ref.Project)
	if id == "" {
		return Candidate{}, fmt.Errorf("modrinth provider missing project")
	}
	var proj struct {
		ID        string `json:"id"`
		Slug      string `json:"slug"`
		SourceURL string `json:"source_url"`
	}
	if e := p.request(ctx, "GET", p.modrinth+"/project/"+url.PathEscape(id), nil, &proj); e != nil {
		return Candidate{}, e
	}
	q := url.Values{}
	if target.Minecraft != "" {
		b, _ := json.Marshal([]string{target.Minecraft})
		q.Set("game_versions", string(b))
	}
	if target.Loader != "" {
		b, _ := json.Marshal([]string{strings.ToLower(target.Loader)})
		q.Set("loaders", string(b))
	}
	raw := p.modrinth + "/project/" + url.PathEscape(proj.ID) + "/version"
	if len(q) > 0 {
		raw += "?" + q.Encode()
	}
	var vv []mrVersion
	if e := p.request(ctx, "GET", raw, nil, &vv); e != nil {
		return Candidate{}, e
	}
	for _, v := range vv {
		if !channelAllowed(v.VersionType, target.Channel) || len(v.Files) == 0 {
			continue
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
		srcArch, srcRef := "", ""
		if repo, ok := githubRepo(proj.SourceURL); ok {
			srcArch = "https://github.com/" + repo + "/archive/refs/heads/main.zip"
			srcRef = "main"
		}
		c := Candidate{Provider: "modrinth", ProjectID: proj.ID, VersionID: v.ID, Version: v.VersionNumber, Filename: f.Filename, URL: f.URL, PageURL: "https://modrinth.com/project/" + proj.Slug + "/version/" + v.ID, Published: v.DatePublished, SHA256: f.Hashes["sha256"], SHA512: f.Hashes["sha512"], SHA1: f.Hashes["sha1"], Size: f.Size, GameVersions: v.GameVersions, Loaders: v.Loaders, Dependencies: deps, SourceURL: proj.SourceURL, SourceArchive: srcArch, SourceRef: srcRef, ReleaseChannel: v.VersionType}
		c.AlternateURLs = appendUniqueURL(c.AlternateURLs, modrinthCanonicalCDNURL(c))
		c.AlternateURLs = appendUniqueURL(c.AlternateURLs, modrinthMavenURL(c))
		return c, nil
	}
	return Candidate{}, fmt.Errorf("no compatible Modrinth release for %s (%s/%s)", id, target.Minecraft, target.Loader)
}
func channelAllowed(got, want string) bool {
	want = strings.ToLower(strings.TrimSpace(want))
	got = strings.ToLower(got)
	if want == "" || want == "release" {
		return got == "release" || got == ""
	}
	if want == "beta" {
		return got == "release" || got == "beta" || got == ""
	}
	return true
}

type cfFile struct {
	ID           int64     `json:"id"`
	DisplayName  string    `json:"displayName"`
	FileName     string    `json:"fileName"`
	DownloadURL  string    `json:"downloadUrl"`
	FileDate     time.Time `json:"fileDate"`
	FileLength   int64     `json:"fileLength"`
	ReleaseType  int       `json:"releaseType"`
	GameVersions []string  `json:"gameVersions"`
	Dependencies []struct {
		ModID        int64 `json:"modId"`
		RelationType int   `json:"relationType"`
	} `json:"dependencies"`
	Hashes []struct {
		Value string `json:"value"`
		Algo  int    `json:"algo"`
	} `json:"hashes"`
}

func (p *providerClient) resolveCurseForge(ctx context.Context, ref ProviderRef, target Target) (Candidate, error) {
	if p.cfKey == "" {
		return Candidate{}, fmt.Errorf("CurseForge API key unavailable")
	}
	id := first(ref.ProjectID, ref.Project)
	if id == "" {
		return Candidate{}, fmt.Errorf("curseforge provider missing project id")
	}
	q := url.Values{"pageSize": {"50"}}
	if target.Minecraft != "" {
		q.Set("gameVersion", target.Minecraft)
	}
	var out struct {
		Data []cfFile `json:"data"`
	}
	if e := p.request(ctx, "GET", p.curseforge+"/mods/"+url.PathEscape(id)+"/files?"+q.Encode(), map[string]string{"x-api-key": p.cfKey}, &out); e != nil {
		return Candidate{}, e
	}
	for _, f := range out.Data {
		if !cfChannelAllowed(f.ReleaseType, target.Channel) {
			continue
		}
		if f.DownloadURL == "" {
			var dl struct {
				Data string `json:"data"`
			}
			raw := p.curseforge + "/mods/" + url.PathEscape(id) + "/files/" + strconv.FormatInt(f.ID, 10) + "/download-url"
			if e := p.request(ctx, "GET", raw, map[string]string{"x-api-key": p.cfKey}, &dl); e == nil {
				f.DownloadURL = strings.TrimSpace(dl.Data)
			}
		}
		if f.DownloadURL == "" {
			f.DownloadURL = curseForgeCDNURL(strconv.FormatInt(f.ID, 10), f.FileName)
		}
		if f.DownloadURL == "" {
			continue
		}
		deps := []Dependency{}
		for _, d := range f.Dependencies {
			if d.RelationType == 3 {
				deps = append(deps, Dependency{Provider: "curseforge", ProjectID: strconv.FormatInt(d.ModID, 10), Required: true})
			}
		}
		hashes := map[int]string{}
		for _, h := range f.Hashes {
			hashes[h.Algo] = h.Value
		}
		c := Candidate{Provider: "curseforge", ProjectID: id, VersionID: strconv.FormatInt(f.ID, 10), Version: first(f.DisplayName, f.FileName), Filename: f.FileName, URL: f.DownloadURL, Published: f.FileDate, SHA1: hashes[1], Size: f.FileLength, GameVersions: f.GameVersions, Loaders: []string{target.Loader}, Dependencies: deps, ReleaseChannel: cfReleaseName(f.ReleaseType)}
		c.AlternateURLs = appendUniqueURL(c.AlternateURLs, curseForgeCDNURL(c.VersionID, c.Filename))
		return c, nil
	}
	return Candidate{}, fmt.Errorf("no compatible CurseForge release for %s", id)
}
func cfChannelAllowed(t int, w string) bool {
	w = strings.ToLower(w)
	if w == "" || w == "release" {
		return t == 1
	}
	if w == "beta" {
		return t == 1 || t == 2
	}
	return true
}
func cfReleaseName(t int) string {
	if t == 1 {
		return "release"
	}
	if t == 2 {
		return "beta"
	}
	return "alpha"
}

type ghAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}
type ghRelease struct {
	TagName     string    `json:"tag_name"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	PublishedAt time.Time `json:"published_at"`
	HTMLURL     string    `json:"html_url"`
	ZipballURL  string    `json:"zipball_url"`
	Assets      []ghAsset `json:"assets"`
}

func (p *providerClient) resolveGitHub(ctx context.Context, ref ProviderRef, target Target) (Candidate, error) {
	repo := first(ref.Repo, ref.Project)
	if r, ok := githubRepo(repo); ok {
		repo = r
	}
	if !strings.Contains(repo, "/") {
		return Candidate{}, fmt.Errorf("github provider missing owner/repo")
	}
	hdr := map[string]string{"Accept": "application/vnd.github+json"}
	if p.ghToken != "" {
		hdr["Authorization"] = "Bearer " + p.ghToken
	}
	var rr []ghRelease
	if e := p.request(ctx, "GET", p.github+"/repos/"+repo+"/releases?per_page=50", hdr, &rr); e != nil {
		return Candidate{}, e
	}
	var rx *regexp.Regexp
	if ref.AssetRegex != "" {
		var e error
		rx, e = regexp.Compile(ref.AssetRegex)
		if e != nil {
			return Candidate{}, e
		}
	}
	for _, r := range rr {
		if r.Draft || (!targetAllowsPrerelease(target) && r.Prerelease) {
			continue
		}
		assets := append([]ghAsset(nil), r.Assets...)
		sort.SliceStable(assets, func(i, j int) bool {
			return scoreAsset(assets[i].Name, target, rx) > scoreAsset(assets[j].Name, target, rx)
		})
		if len(assets) == 0 || scoreAsset(assets[0].Name, target, rx) < 0 {
			continue
		}
		a := assets[0]
		ch := "release"
		if r.Prerelease {
			ch = "prerelease"
		}
		return Candidate{Provider: "github", ProjectID: repo, VersionID: r.TagName, Version: r.TagName, Filename: a.Name, URL: a.BrowserDownloadURL, PageURL: r.HTMLURL, Published: r.PublishedAt, Size: a.Size, SourceURL: "https://github.com/" + repo, SourceArchive: r.ZipballURL, SourceRef: r.TagName, ReleaseChannel: ch}, nil
	}
	return Candidate{}, fmt.Errorf("no compatible GitHub release asset for %s", repo)
}

type ghRepoMeta struct {
	DefaultBranch string `json:"default_branch"`
}
type ghBranchMeta struct {
	Commit struct {
		SHA string `json:"sha"`
	} `json:"commit"`
}

func (p *providerClient) resolveGitHubBranch(ctx context.Context, ref ProviderRef, target Target) (Candidate, error) {
	repo := first(ref.Repo, ref.Project)
	if r, ok := githubRepo(repo); ok {
		repo = r
	}
	if !strings.Contains(repo, "/") {
		return Candidate{}, fmt.Errorf("github-branch provider missing owner/repo")
	}
	hdr := map[string]string{"Accept": "application/vnd.github+json"}
	if p.ghToken != "" {
		hdr["Authorization"] = "Bearer " + p.ghToken
	}
	branch := ref.Branch
	if branch == "" {
		var meta ghRepoMeta
		if err := p.request(ctx, "GET", p.github+"/repos/"+repo, hdr, &meta); err != nil {
			return Candidate{}, err
		}
		branch = first(meta.DefaultBranch, "main")
	}
	var b ghBranchMeta
	if err := p.request(ctx, "GET", p.github+"/repos/"+repo+"/branches/"+url.PathEscape(branch), hdr, &b); err != nil {
		return Candidate{}, err
	}
	if b.Commit.SHA == "" {
		return Candidate{}, fmt.Errorf("github branch %s/%s missing commit sha", repo, branch)
	}
	short := b.Commit.SHA
	if len(short) > 12 {
		short = short[:12]
	}
	name := strings.ReplaceAll(repo, "/", "-") + "-" + branch + "-" + short + ".zip"
	return Candidate{Provider: "github-branch", ProjectID: repo, VersionID: b.Commit.SHA, Version: b.Commit.SHA, Filename: name, URL: "https://github.com/" + repo + "/archive/" + b.Commit.SHA + ".zip", PageURL: "https://github.com/" + repo + "/tree/" + branch, SourceURL: "https://github.com/" + repo, SourceArchive: "https://github.com/" + repo + "/archive/" + b.Commit.SHA + ".zip", SourceRef: b.Commit.SHA, ReleaseChannel: "branch"}, nil
}

func targetAllowsPrerelease(t Target) bool {
	x := strings.ToLower(t.Channel)
	return x == "beta" || x == "alpha" || x == "nightly"
}
func scoreAsset(name string, t Target, rx *regexp.Regexp) int {
	if rx != nil && !rx.MatchString(name) {
		return -100
	}
	low := strings.ToLower(name)
	s := 0
	if t.Minecraft != "" && strings.Contains(low, strings.ToLower(t.Minecraft)) {
		s += 8
	}
	if t.Loader != "" && strings.Contains(low, strings.ToLower(t.Loader)) {
		s += 6
	}
	if t.OS != "" && strings.Contains(low, strings.ToLower(t.OS)) {
		s += 4
	}
	if t.Arch != "" && strings.Contains(low, strings.ToLower(t.Arch)) {
		s += 3
	}
	if strings.Contains(low, "source") || strings.Contains(low, "javadoc") {
		s -= 20
	}
	return s
}
func filenameFromURL(raw string) string {
	u, e := url.Parse(raw)
	if e != nil {
		return "download.bin"
	}
	p := strings.TrimSuffix(u.Path, "/")
	if i := strings.LastIndex(p, "/"); i >= 0 && i+1 < len(p) {
		return p[i+1:]
	}
	return "download.bin"
}
func githubRepo(raw string) (string, bool) {
	raw = strings.TrimSpace(strings.TrimSuffix(raw, ".git"))
	raw = strings.TrimPrefix(raw, "https://github.com/")
	raw = strings.TrimPrefix(raw, "http://github.com/")
	parts := strings.Split(strings.Trim(raw, "/"), "/")
	if len(parts) >= 2 {
		return parts[0] + "/" + parts[1], true
	}
	return "", false
}
func first(v ...string) string {
	for _, s := range v {
		if strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

type ghTag struct {
	Name       string `json:"name"`
	ZipballURL string `json:"zipball_url"`
}

func (p *providerClient) sourceArchiveForCandidate(ctx context.Context, c Candidate) (string, string, error) {
	repo, ok := githubRepo(c.SourceURL)
	if !ok {
		return c.SourceArchive, c.SourceRef, nil
	}
	hdr := map[string]string{"Accept": "application/vnd.github+json"}
	if p.ghToken != "" {
		hdr["Authorization"] = "Bearer " + p.ghToken
	}
	var releases []ghRelease
	if err := p.request(ctx, "GET", p.github+"/repos/"+repo+"/releases?per_page=100", hdr, &releases); err == nil {
		want := normalizeVersion(c.Version)
		for _, r := range releases {
			if normalizeVersion(r.TagName) == want || normalizeVersion(strings.TrimSpace(r.TagName)) == want {
				if r.ZipballURL != "" {
					return r.ZipballURL, r.TagName, nil
				}
			}
		}
	}
	var tags []ghTag
	if err := p.request(ctx, "GET", p.github+"/repos/"+repo+"/tags?per_page=100", hdr, &tags); err == nil {
		want := normalizeVersion(c.Version)
		for _, t := range tags {
			if normalizeVersion(t.Name) == want || strings.Contains(normalizeVersion(t.Name), want) {
				if t.ZipballURL != "" {
					return t.ZipballURL, t.Name, nil
				}
			}
		}
	}
	if c.SourceArchive != "" {
		return c.SourceArchive, c.SourceRef, nil
	}
	return "https://github.com/" + repo + "/archive/refs/heads/main.zip", "main", nil
}
func normalizeVersion(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.TrimPrefix(s, "v")
	s = strings.ReplaceAll(s, "_", ".")
	return s
}
