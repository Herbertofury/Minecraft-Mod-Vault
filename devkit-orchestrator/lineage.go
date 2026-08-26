package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type heritageMRVersion struct {
	ID            string    `json:"id"`
	ProjectID     string    `json:"project_id"`
	VersionNumber string    `json:"version_number"`
	VersionType   string    `json:"version_type"`
	DatePublished time.Time `json:"date_published"`
	GameVersions  []string  `json:"game_versions"`
	Loaders       []string  `json:"loaders"`
	Changelog     string    `json:"changelog"`
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

type heritageGHRelease struct {
	TagName     string    `json:"tag_name"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	PublishedAt time.Time `json:"published_at"`
	HTMLURL     string    `json:"html_url"`
	Body        string    `json:"body"`
	ZipballURL  string    `json:"zipball_url"`
	Assets      []ghAsset `json:"assets"`
}

func (p *providerClient) resolveHeritage(ctx context.Context, ref ProviderRef, target Target) (Candidate, error) {
	switch strings.ToLower(ref.Type) {
	case "modrinth":
		return p.resolveHeritageModrinth(ctx, ref, target)
	case "curseforge":
		return p.resolveHeritageCurseForge(ctx, ref, target)
	case "github":
		return p.resolveHeritageGitHub(ctx, ref, target)
	default:
		return p.resolve(ctx, ref, target)
	}
}

func (p *providerClient) resolveHeritageModrinth(ctx context.Context, ref ProviderRef, target Target) (Candidate, error) {
	id := first(ref.ProjectID, ref.Project)
	if id == "" {
		return Candidate{}, fmt.Errorf("modrinth provider missing project")
	}
	var proj struct {
		ID        string `json:"id"`
		Slug      string `json:"slug"`
		SourceURL string `json:"source_url"`
	}
	if err := p.request(ctx, "GET", p.modrinth+"/project/"+url.PathEscape(id), nil, &proj); err != nil {
		return Candidate{}, err
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
	var vv []heritageMRVersion
	if err := p.request(ctx, "GET", raw, nil, &vv); err != nil {
		return Candidate{}, err
	}
	sort.SliceStable(vv, func(i, j int) bool { return vv[i].DatePublished.After(vv[j].DatePublished) })
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
		return Candidate{Provider: "modrinth", ProjectID: proj.ID, VersionID: v.ID, Version: v.VersionNumber, Filename: f.Filename, URL: f.URL, PageURL: "https://modrinth.com/project/" + proj.Slug + "/version/" + v.ID, Published: v.DatePublished, SHA256: f.Hashes["sha256"], SHA512: f.Hashes["sha512"], SHA1: f.Hashes["sha1"], Size: f.Size, GameVersions: v.GameVersions, Loaders: v.Loaders, Dependencies: deps, SourceURL: proj.SourceURL, SourceArchive: srcArch, SourceRef: srcRef, ReleaseChannel: v.VersionType}, nil
	}
	return Candidate{}, fmt.Errorf("no heritage Modrinth release for %s (%s/%s)", id, target.Minecraft, target.Loader)
}

func curseForgeLoaderType(loader string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(loader)) {
	case "forge":
		return "1", true
	case "fabric":
		return "4", true
	case "quilt":
		return "5", true
	case "neoforge":
		return "6", true
	default:
		return "", false
	}
}

func (p *providerClient) resolveHeritageCurseForge(ctx context.Context, ref ProviderRef, target Target) (Candidate, error) {
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
		if lt, ok := curseForgeLoaderType(target.Loader); ok {
			q.Set("modLoaderType", lt)
		}
	}
	var out struct {
		Data []cfFile `json:"data"`
	}
	if err := p.request(ctx, "GET", p.curseforge+"/mods/"+url.PathEscape(id)+"/files?"+q.Encode(), map[string]string{"x-api-key": p.cfKey}, &out); err != nil {
		return Candidate{}, err
	}
	sort.SliceStable(out.Data, func(i, j int) bool { return out.Data[i].FileDate.After(out.Data[j].FileDate) })
	for _, f := range out.Data {
		if f.DownloadURL == "" || !cfChannelAllowed(f.ReleaseType, target.Channel) {
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
		return Candidate{Provider: "curseforge", ProjectID: id, VersionID: strconv.FormatInt(f.ID, 10), Version: first(f.DisplayName, f.FileName), Filename: f.FileName, URL: f.DownloadURL, Published: f.FileDate, SHA1: hashes[1], Size: f.FileLength, GameVersions: f.GameVersions, Loaders: []string{target.Loader}, Dependencies: deps, ReleaseChannel: cfReleaseName(f.ReleaseType)}, nil
	}
	return Candidate{}, fmt.Errorf("no heritage CurseForge release for %s", id)
}

func (p *providerClient) resolveHeritageGitHub(ctx context.Context, ref ProviderRef, target Target) (Candidate, error) {
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
	var rr []heritageGHRelease
	if err := p.request(ctx, "GET", p.github+"/repos/"+repo+"/releases?per_page=100", hdr, &rr); err != nil {
		return Candidate{}, err
	}
	sort.SliceStable(rr, func(i, j int) bool { return rr[i].PublishedAt.After(rr[j].PublishedAt) })
	var rx *regexp.Regexp
	if ref.AssetRegex != "" {
		var err error
		rx, err = regexp.Compile(ref.AssetRegex)
		if err != nil {
			return Candidate{}, err
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
	return Candidate{}, fmt.Errorf("no heritage GitHub release asset for %s", repo)
}

func (p *providerClient) releaseLineage(ctx context.Context, ref ProviderRef, compatible, latest Candidate, target Target) ([]ReleaseNote, error) {
	switch strings.ToLower(ref.Type) {
	case "modrinth":
		id := first(ref.ProjectID, ref.Project, latest.ProjectID)
		var vv []heritageMRVersion
		if err := p.request(ctx, "GET", p.modrinth+"/project/"+url.PathEscape(id)+"/version", nil, &vv); err != nil {
			return nil, err
		}
		out := []ReleaseNote{}
		for _, v := range vv {
			if withinLineage(v.DatePublished, compatible, latest) || v.ID == compatible.VersionID || v.ID == latest.VersionID {
				out = append(out, ReleaseNote{Provider: "modrinth", VersionID: v.ID, Version: v.VersionNumber, Published: v.DatePublished, GameVersions: v.GameVersions, Loaders: v.Loaders, Changelog: v.Changelog})
			}
		}
		sort.Slice(out, func(i, j int) bool { return out[i].Published.Before(out[j].Published) })
		return out, nil
	case "curseforge":
		return p.curseForgeLineage(ctx, ref, compatible, latest, target)
	case "github":
		return p.githubLineage(ctx, ref, compatible, latest, target)
	case "github-branch":
		return []ReleaseNote{{Provider: latest.Provider, VersionID: latest.VersionID, Version: latest.Version, Published: latest.Published, PageURL: latest.PageURL}}, nil
	default:
		return nil, nil
	}
}

func withinLineage(t time.Time, compatible, latest Candidate) bool {
	if t.IsZero() || compatible.Published.IsZero() || latest.Published.IsZero() {
		return true
	}
	lo, hi := compatible.Published, latest.Published
	if hi.Before(lo) {
		lo, hi = hi, lo
	}
	return !t.Before(lo) && !t.After(hi)
}

func (p *providerClient) curseForgeLineage(ctx context.Context, ref ProviderRef, compatible, latest Candidate, target Target) ([]ReleaseNote, error) {
	if p.cfKey == "" {
		return nil, fmt.Errorf("CurseForge API key unavailable")
	}
	id := first(ref.ProjectID, ref.Project, latest.ProjectID)
	out := []ReleaseNote{}
	for index := 0; index < 10000 && len(out) < 500; index += 50 {
		q := url.Values{"pageSize": {"50"}, "index": {strconv.Itoa(index)}}
		var page struct {
			Data       []cfFile `json:"data"`
			Pagination struct {
				ResultCount int `json:"resultCount"`
				TotalCount  int `json:"totalCount"`
			} `json:"pagination"`
		}
		if err := p.request(ctx, "GET", p.curseforge+"/mods/"+url.PathEscape(id)+"/files?"+q.Encode(), map[string]string{"x-api-key": p.cfKey}, &page); err != nil {
			return nil, err
		}
		for _, f := range page.Data {
			fid := strconv.FormatInt(f.ID, 10)
			if withinLineage(f.FileDate, compatible, latest) || fid == compatible.VersionID || fid == latest.VersionID {
				out = append(out, ReleaseNote{Provider: "curseforge", VersionID: fid, Version: first(f.DisplayName, f.FileName), Published: f.FileDate, GameVersions: f.GameVersions})
			}
		}
		if len(page.Data) < 50 || page.Pagination.ResultCount == 0 || (page.Pagination.TotalCount > 0 && index+50 >= page.Pagination.TotalCount) {
			break
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Published.Before(out[j].Published) })
	return out, nil
}

func (p *providerClient) githubLineage(ctx context.Context, ref ProviderRef, compatible, latest Candidate, target Target) ([]ReleaseNote, error) {
	repo := first(ref.Repo, ref.Project, latest.ProjectID)
	if r, ok := githubRepo(repo); ok {
		repo = r
	}
	hdr := map[string]string{"Accept": "application/vnd.github+json"}
	if p.ghToken != "" {
		hdr["Authorization"] = "Bearer " + p.ghToken
	}
	var rr []heritageGHRelease
	if err := p.request(ctx, "GET", p.github+"/repos/"+repo+"/releases?per_page=100", hdr, &rr); err != nil {
		return nil, err
	}
	out := []ReleaseNote{}
	for _, r := range rr {
		if r.Draft || (!targetAllowsPrerelease(target) && r.Prerelease) {
			continue
		}
		if withinLineage(r.PublishedAt, compatible, latest) || r.TagName == compatible.VersionID || r.TagName == latest.VersionID {
			out = append(out, ReleaseNote{Provider: "github", VersionID: r.TagName, Version: r.TagName, Published: r.PublishedAt, PageURL: r.HTMLURL, Changelog: r.Body})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Published.Before(out[j].Published) })
	return out, nil
}
