package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type ProjectAuthor struct {
	Name      string `json:"name"`
	AvatarURL string `json:"avatarUrl,omitempty"`
	Role      string `json:"role,omitempty"`
	URL       string `json:"url,omitempty"`
}

type ProjectFileSummary struct {
	Name    string            `json:"name"`
	URL     string            `json:"url,omitempty"`
	Size    int64             `json:"size,omitempty"`
	Primary bool              `json:"primary,omitempty"`
	Hashes  map[string]string `json:"hashes,omitempty"`
}

type ProjectVersionSummary struct {
	ID           string               `json:"id"`
	Name         string               `json:"name,omitempty"`
	Version      string               `json:"version,omitempty"`
	Published    string               `json:"published,omitempty"`
	GameVersions []string             `json:"gameVersions,omitempty"`
	Loaders      []string             `json:"loaders,omitempty"`
	Dependencies []string             `json:"dependencies,omitempty"`
	Changelog    string               `json:"changelog,omitempty"`
	Files        []ProjectFileSummary `json:"files,omitempty"`
	Installable  bool                 `json:"installable,omitempty"`
}

type ProjectDetails struct {
	Provider      string                  `json:"provider"`
	ID            string                  `json:"id"`
	Slug          string                  `json:"slug,omitempty"`
	ProjectType   string                  `json:"projectType,omitempty"`
	Title         string                  `json:"title"`
	Summary       string                  `json:"summary,omitempty"`
	Description   string                  `json:"description,omitempty"`
	IconURL       string                  `json:"iconUrl,omitempty"`
	Gallery       []string                `json:"gallery,omitempty"`
	Authors       []ProjectAuthor         `json:"authors,omitempty"`
	Categories    []string                `json:"categories,omitempty"`
	GameVersions  []string                `json:"gameVersions,omitempty"`
	Loaders       []string                `json:"loaders,omitempty"`
	Versions      []ProjectVersionSummary `json:"versions,omitempty"`
	Downloads     int64                   `json:"downloads,omitempty"`
	Followers     int64                   `json:"followers,omitempty"`
	Updated       string                  `json:"updated,omitempty"`
	Published     string                  `json:"published,omitempty"`
	PageURL       string                  `json:"pageUrl,omitempty"`
	Links         map[string]string       `json:"links,omitempty"`
	Installable   bool                    `json:"installable"`
	InstallMode   string                  `json:"installMode,omitempty"`
	ProviderLabel string                  `json:"providerLabel,omitempty"`
	FetchedAt     string                  `json:"fetchedAt"`
}

func (a *App) handleProviderDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	q := r.URL.Query()
	provider := strings.ToLower(strings.TrimSpace(q.Get("provider")))
	id := firstNonEmpty(q.Get("id"), q.Get("slug"))
	pageURL := strings.TrimSpace(q.Get("url"))
	if provider == "" || (id == "" && pageURL == "") {
		writeJSON(w, http.StatusBadRequest, APIError{Error: "provider and id/slug/url are required"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 35*time.Second)
	defer cancel()
	started := time.Now()
	detail, err := a.fetchProjectDetails(ctx, provider, id, pageURL)
	a.noteProviderAttempt(provider, started, boolToInt(err == nil), err)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, APIError{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

func (a *App) fetchProjectDetails(ctx context.Context, provider, id, pageURL string) (ProjectDetails, error) {
	var d ProjectDetails
	var err error
	switch provider {
	case "modrinth":
		d, err = a.modrinthDetails(ctx, id)
	case "curseforge":
		d, err = a.curseForgeDetails(ctx, id, pageURL)
	case "github":
		d, err = a.githubDetails(ctx, id)
	case "smithed":
		d, err = a.smithedDetails(ctx, id)
	case "hangar":
		d, err = a.hangarDetails(ctx, id)
	case "spigot":
		d, err = a.spigotDetails(ctx, id)
	case "spongeore":
		d, err = a.spongeOreDetails(ctx, id)
	case "builtbybit":
		d, err = a.builtByBitDetails(ctx, id)
	case "atlauncher":
		d, err = a.atLauncherDetails(ctx, id)
	case "ftb":
		d, err = a.ftbDetails(ctx, id)
	case "technic":
		d, err = a.technicDetails(ctx, id)
	case "nexusmods":
		d, err = a.nexusDetails(ctx, id)
	case "vanillatweaks":
		d, err = a.vanillaTweaksDetails(ctx, id, pageURL)
	case "minecrafthub":
		d, err = a.minecraftHubDetails(ctx, id, pageURL)
	case "planetminecraft", "mcpedl", "marketplace", "bukkitdev", "moddb", "polymart", "minecraftmaps", "resourcepacknet", "texturepacks", "mcreator", "shaderpackscom", "shaderpacksnet", "minecraftshader", "skindex":
		d, err = a.genericWebDetails(ctx, provider, pageURL, id)
	default:
		err = fmt.Errorf("unknown provider %q", provider)
	}
	if err != nil {
		return ProjectDetails{}, err
	}
	d.Provider = provider
	d.FetchedAt = time.Now().UTC().Format(time.RFC3339)
	if p := providerInfoByID(provider); p != nil {
		d.ProviderLabel = p.Name
		if d.PageURL == "" {
			d.PageURL = p.HomeURL
		}
	}
	d.Gallery = uniqueStringsPreserve(d.Gallery)
	d.Categories = uniqueStringsPreserve(d.Categories)
	d.GameVersions = uniqueStringsPreserve(d.GameVersions)
	d.Loaders = uniqueStrings(d.Loaders)
	return d, nil
}

func (a *App) modrinthDetails(ctx context.Context, id string) (ProjectDetails, error) {
	var p struct {
		ID                   string   `json:"id"`
		Slug                 string   `json:"slug"`
		ProjectType          string   `json:"project_type"`
		Title                string   `json:"title"`
		Description          string   `json:"description"`
		Body                 string   `json:"body"`
		IconURL              string   `json:"icon_url"`
		Categories           []string `json:"categories"`
		AdditionalCategories []string `json:"additional_categories"`
		GameVersions         []string `json:"game_versions"`
		Loaders              []string `json:"loaders"`
		Downloads            int64    `json:"downloads"`
		Followers            int64    `json:"followers"`
		Published            string   `json:"published"`
		Updated              string   `json:"updated"`
		IssuesURL            string   `json:"issues_url"`
		SourceURL            string   `json:"source_url"`
		WikiURL              string   `json:"wiki_url"`
		DiscordURL           string   `json:"discord_url"`
		Gallery              []struct {
			URL      string `json:"url"`
			Title    string `json:"title"`
			Featured bool   `json:"featured"`
		} `json:"gallery"`
	}
	if err := a.getJSON(ctx, modrinthAPIBase()+"/project/"+url.PathEscape(id), nil, &p); err != nil {
		return ProjectDetails{}, err
	}
	var versions []ModrinthVersion
	_ = a.getJSON(ctx, modrinthAPIBase()+"/project/"+url.PathEscape(p.ID)+"/version", nil, &versions)
	var members []struct {
		User struct {
			Username  string `json:"username"`
			AvatarURL string `json:"avatar_url"`
			ID        string `json:"id"`
		} `json:"user"`
		Role string `json:"role"`
	}
	_ = a.getJSON(ctx, modrinthAPIBase()+"/project/"+url.PathEscape(p.ID)+"/members", nil, &members)
	d := ProjectDetails{ID: p.ID, Slug: p.Slug, ProjectType: p.ProjectType, Title: p.Title, Summary: p.Description, Description: p.Body, IconURL: p.IconURL, Downloads: p.Downloads, Followers: p.Followers, Updated: p.Updated, Published: p.Published, PageURL: "https://modrinth.com/" + p.ProjectType + "/" + p.Slug, Categories: append(append([]string{}, p.Categories...), p.AdditionalCategories...), GameVersions: p.GameVersions, Loaders: p.Loaders, Installable: true, InstallMode: "verified-native", Links: map[string]string{}}
	for _, g := range p.Gallery {
		d.Gallery = append(d.Gallery, g.URL)
	}
	for _, m := range members {
		d.Authors = append(d.Authors, ProjectAuthor{Name: m.User.Username, AvatarURL: m.User.AvatarURL, Role: m.Role, URL: "https://modrinth.com/user/" + url.PathEscape(m.User.Username)})
	}
	for k, v := range map[string]string{"Issues": p.IssuesURL, "Source": p.SourceURL, "Wiki": p.WikiURL, "Discord": p.DiscordURL} {
		if v != "" {
			d.Links[k] = v
		}
	}
	for _, v := range versions {
		pv := ProjectVersionSummary{ID: v.ID, Name: v.Name, Version: v.VersionNumber, Published: v.DatePublished, GameVersions: v.GameVersions, Loaders: v.Loaders, Installable: true}
		for _, f := range v.Files {
			pv.Files = append(pv.Files, ProjectFileSummary{Name: f.Filename, URL: f.URL, Size: f.Size, Primary: f.Primary, Hashes: f.Hashes})
		}
		d.Versions = append(d.Versions, pv)
	}
	return d, nil
}

func (a *App) curseForgeDetails(ctx context.Context, id, pageURL string) (ProjectDetails, error) {
	a.mu.RLock()
	key := strings.TrimSpace(a.settings.CurseForgeAPIKey)
	a.mu.RUnlock()
	if key == "" {
		return a.genericWebDetails(ctx, "curseforge", pageURL, id)
	}
	if _, err := strconv.ParseInt(id, 10, 64); err != nil {
		a.mu.RLock()
		game, loader := a.settings.GameVersion, a.settings.Loader
		a.mu.RUnlock()
		items, searchErr := a.searchCurseForgeAPI(ctx, key, id, providerSearchOptions{GameVersion: game, Loader: loader, Limit: 20})
		if searchErr != nil || len(items) == 0 {
			// The public web result remains a fully integrated detail fallback even
			// when the API cannot resolve a web-index slug.
			return a.genericWebDetails(ctx, "curseforge", pageURL, id)
		}
		best := items[0]
		for _, item := range items {
			if strings.EqualFold(item.Slug, id) || strings.EqualFold(item.Title, id) {
				best = item
				break
			}
		}
		id = best.ID
		if pageURL == "" {
			pageURL = best.PageURL
		}
	}
	var resp struct {
		Data map[string]any `json:"data"`
	}
	if err := a.getJSON(ctx, curseForgeAPIBase()+"/v1/mods/"+url.PathEscape(id), map[string]string{"x-api-key": key}, &resp); err != nil {
		return ProjectDetails{}, err
	}
	m := resp.Data
	title := stringFromAny(m["name"])
	slug := stringFromAny(m["slug"])
	d := ProjectDetails{ID: stringFromAny(m["id"]), Slug: slug, ProjectType: "mod", Title: title, Summary: stringFromAny(m["summary"]), Downloads: int64FromAny(m["downloadCount"]), Updated: timeFromAny(m["dateModified"]), PageURL: firstNonEmpty(stringFromAny(m["links"]), pageURL, "https://www.curseforge.com/minecraft/mc-mods/"+slug), Installable: true, InstallMode: "verified-native-with-key", Links: map[string]string{}}
	if logo, ok := m["logo"].(map[string]any); ok {
		d.IconURL = firstNonEmpty(stringFromAny(logo["thumbnailUrl"]), stringFromAny(logo["url"]))
	}
	for _, item := range anySlice(m["screenshots"]) {
		if mm, ok := item.(map[string]any); ok {
			d.Gallery = append(d.Gallery, firstNonEmpty(stringFromAny(mm["url"]), stringFromAny(mm["thumbnailUrl"])))
		}
	}
	for _, item := range anySlice(m["authors"]) {
		if mm, ok := item.(map[string]any); ok {
			d.Authors = append(d.Authors, ProjectAuthor{Name: stringFromAny(mm["name"]), URL: stringFromAny(mm["url"])})
		}
	}
	for _, item := range anySlice(m["categories"]) {
		if mm, ok := item.(map[string]any); ok {
			d.Categories = append(d.Categories, stringFromAny(mm["name"]))
		}
	}
	values := url.Values{"pageSize": {"30"}}
	if game := strings.TrimSpace(a.settings.GameVersion); game != "" {
		values.Set("gameVersion", game)
	}
	if loader := curseLoaderType(a.settings.Loader); loader != 0 {
		values.Set("modLoaderType", strconv.Itoa(loader))
	}
	var filesResp struct {
		Data []map[string]any `json:"data"`
	}
	if err := a.getJSON(ctx, curseForgeAPIBase()+"/v1/mods/"+url.PathEscape(id)+"/files?"+values.Encode(), map[string]string{"x-api-key": key}, &filesResp); err == nil {
		for _, fm := range filesResp.Data {
			pv := ProjectVersionSummary{ID: stringFromAny(fm["id"]), Name: stringFromAny(fm["displayName"]), Version: stringFromAny(fm["fileName"]), Published: timeFromAny(fm["fileDate"]), GameVersions: stringSliceFromAny(fm["gameVersions"]), Installable: true}
			if dl := stringFromAny(fm["downloadUrl"]); dl != "" {
				pv.Files = append(pv.Files, ProjectFileSummary{Name: stringFromAny(fm["fileName"]), URL: dl, Size: int64FromAny(fm["fileLength"]), Primary: true})
			}
			d.Versions = append(d.Versions, pv)
			d.GameVersions = append(d.GameVersions, pv.GameVersions...)
		}
	}
	return d, nil
}

func (a *App) githubDetails(ctx context.Context, repo string) (ProjectDetails, error) {
	repo = strings.Trim(strings.TrimPrefix(repo, "https://github.com/"), "/")
	parts := strings.Split(repo, "/")
	if len(parts) < 2 {
		return ProjectDetails{}, errors.New("GitHub project must be owner/repository")
	}
	repo = parts[0] + "/" + parts[1]
	headers := map[string]string{"Accept": "application/vnd.github+json", "X-GitHub-Api-Version": "2022-11-28"}
	a.mu.RLock()
	if token := strings.TrimSpace(a.settings.GitHubToken); token != "" {
		headers["Authorization"] = "Bearer " + token
	}
	a.mu.RUnlock()
	var m map[string]any
	if err := a.getJSON(ctx, "https://api.github.com/repos/"+repo, headers, &m); err != nil {
		return ProjectDetails{}, err
	}
	d := ProjectDetails{ID: repo, Slug: repo, ProjectType: inferProjectType(stringFromAny(m["description"])+" "+strings.Join(stringSliceFromAny(m["topics"]), " "), "mod"), Title: firstNonEmpty(stringFromAny(m["name"]), repo), Summary: stringFromAny(m["description"]), Description: stringFromAny(m["description"]), Downloads: int64FromAny(m["stargazers_count"]), Followers: int64FromAny(m["subscribers_count"]), Updated: timeFromAny(m["updated_at"]), Published: timeFromAny(m["created_at"]), PageURL: stringFromAny(m["html_url"]), Categories: stringSliceFromAny(m["topics"]), Installable: true, InstallMode: "release-assets", Links: map[string]string{"Source": stringFromAny(m["html_url"]), "Issues": stringFromAny(m["issues_url"]), "Homepage": stringFromAny(m["homepage"])}}
	owner, _ := m["owner"].(map[string]any)
	if owner != nil {
		d.IconURL = stringFromAny(owner["avatar_url"])
		d.Authors = append(d.Authors, ProjectAuthor{Name: stringFromAny(owner["login"]), AvatarURL: stringFromAny(owner["avatar_url"]), URL: stringFromAny(owner["html_url"]), Role: "owner"})
	}
	d.Gallery = []string{"https://opengraph.githubassets.com/mmv/" + repo}
	var readme map[string]any
	if err := a.getJSON(ctx, "https://api.github.com/repos/"+repo+"/readme", headers, &readme); err == nil {
		if content := strings.ReplaceAll(stringFromAny(readme["content"]), "\n", ""); content != "" {
			if b, err := base64.StdEncoding.DecodeString(content); err == nil {
				d.Description = cleanMarkdownForPreview(string(b))
			}
		}
	}
	var releases []map[string]any
	if err := a.getJSON(ctx, "https://api.github.com/repos/"+repo+"/releases?per_page=30", headers, &releases); err == nil {
		for _, rel := range releases {
			pv := ProjectVersionSummary{ID: stringFromAny(rel["id"]), Name: stringFromAny(rel["name"]), Version: stringFromAny(rel["tag_name"]), Published: timeFromAny(rel["published_at"]), Changelog: cleanMarkdownForPreview(stringFromAny(rel["body"])), Installable: true}
			for _, asset := range anySlice(rel["assets"]) {
				if am, ok := asset.(map[string]any); ok {
					pv.Files = append(pv.Files, ProjectFileSummary{Name: stringFromAny(am["name"]), URL: stringFromAny(am["browser_download_url"]), Size: int64FromAny(am["size"])})
				}
			}
			d.Versions = append(d.Versions, pv)
		}
	}
	return d, nil
}

func (a *App) hangarDetails(ctx context.Context, id string) (ProjectDetails, error) {
	base := providerBase("MMV_HANGAR_API_BASE", "https://hangar.papermc.io/api/v1")
	slug := id
	owner := ""
	if strings.Contains(id, "/") {
		parts := strings.SplitN(id, "/", 2)
		owner, slug = parts[0], parts[1]
	}
	var m map[string]any
	if err := a.getJSON(ctx, base+"/projects/"+url.PathEscape(slug), nil, &m); err != nil {
		return ProjectDetails{}, err
	}
	ns, _ := m["namespace"].(map[string]any)
	owner = firstNonEmpty(stringFromAny(ns["owner"]), owner)
	slug = firstNonEmpty(stringFromAny(ns["slug"]), slug)
	stats, _ := m["stats"].(map[string]any)
	d := ProjectDetails{ID: owner + "/" + slug, Slug: slug, ProjectType: "plugin", Title: stringFromAny(m["name"]), Summary: stringFromAny(m["description"]), Description: cleanMarkdownForPreview(stringFromAny(m["mainPageContent"])), IconURL: stringFromAny(m["avatarUrl"]), Downloads: int64FromAny(stats["downloads"]), Followers: int64FromAny(stats["stars"]) + int64FromAny(stats["watchers"]), Updated: timeFromAny(m["lastUpdated"]), Published: timeFromAny(m["publishedAt"]), Categories: nonEmptyStrings(stringFromAny(m["category"]), "plugin", "paper"), PageURL: "https://hangar.papermc.io/" + url.PathEscape(owner) + "/" + url.PathEscape(slug), Installable: true, InstallMode: "native-version-download"}
	for platform, versions := range mapStringSlices(m["supportedPlatforms"]) {
		d.Loaders = append(d.Loaders, strings.ToLower(platform))
		d.GameVersions = append(d.GameVersions, versions...)
	}
	for _, name := range stringSliceFromAny(m["memberNames"]) {
		d.Authors = append(d.Authors, ProjectAuthor{Name: name})
	}
	var vr struct {
		Result []map[string]any `json:"result"`
	}
	if err := a.getJSON(ctx, base+"/projects/"+url.PathEscape(slug)+"/versions?limit=25&offset=0", nil, &vr); err == nil {
		for _, vm := range vr.Result {
			pv := ProjectVersionSummary{ID: stringFromAny(vm["id"]), Name: firstNonEmpty(stringFromAny(vm["name"]), stringFromAny(vm["version"])), Version: firstNonEmpty(stringFromAny(vm["name"]), stringFromAny(vm["version"])), Published: firstNonEmpty(timeFromAny(vm["createdAt"]), timeFromAny(vm["publishedAt"])), Installable: true}
			for platform, versions := range mapStringSlices(vm["platformDependencies"]) {
				pv.Loaders = append(pv.Loaders, strings.ToLower(platform))
				pv.GameVersions = append(pv.GameVersions, versions...)
			}
			if downloads, ok := vm["downloads"].(map[string]any); ok {
				for platform, raw := range downloads {
					dm, _ := raw.(map[string]any)
					if dm == nil {
						continue
					}
					fi, _ := dm["fileInfo"].(map[string]any)
					name := stringFromAny(fi["name"])
					fileURL := firstNonEmpty(stringFromAny(dm["downloadUrl"]), stringFromAny(dm["externalUrl"]))
					hashes := map[string]string{}
					if h := stringFromAny(fi["sha256Hash"]); h != "" {
						hashes["sha256"] = h
					}
					pv.Files = append(pv.Files, ProjectFileSummary{Name: firstNonEmpty(name, platform+" download"), URL: fileURL, Size: int64FromAny(fi["sizeBytes"]), Hashes: hashes})
				}
			}
			d.Versions = append(d.Versions, pv)
		}
	}
	return d, nil
}

func (a *App) spigotDetails(ctx context.Context, id string) (ProjectDetails, error) {
	base := providerBase("MMV_SPIGET_API_BASE", "https://api.spiget.org/v2")
	var m map[string]any
	if err := a.getJSON(ctx, base+"/resources/"+url.PathEscape(id), nil, &m); err != nil {
		return ProjectDetails{}, err
	}
	d := ProjectDetails{ID: id, Slug: id, ProjectType: "plugin", Title: stringFromAny(m["name"]), Summary: stringFromAny(m["tag"]), Description: firstNonEmpty(stringFromAny(m["tag"]), stringFromAny(m["description"])), Downloads: int64FromAny(m["downloads"]), Followers: int64FromAny(m["likes"]), Updated: timeFromAny(m["updateDate"]), Published: timeFromAny(m["releaseDate"]), Categories: []string{"plugin", "spigot"}, PageURL: "https://www.spigotmc.org/resources/" + id + "/", Installable: !boolFromAny(m["external"]), InstallMode: "resource-download-when-available"}
	if icon, ok := m["icon"].(map[string]any); ok {
		d.IconURL = absoluteURL("https://www.spigotmc.org", stringFromAny(icon["url"]))
	}
	if author, ok := m["author"].(map[string]any); ok {
		d.Authors = append(d.Authors, ProjectAuthor{Name: firstNonEmpty(stringFromAny(author["name"]), stringFromAny(author["id"]))})
	}
	var versions []map[string]any
	if err := a.getJSON(ctx, base+"/resources/"+url.PathEscape(id)+"/versions?size=30&sort=-id", nil, &versions); err == nil {
		for _, vm := range versions {
			d.Versions = append(d.Versions, ProjectVersionSummary{ID: stringFromAny(vm["id"]), Name: stringFromAny(vm["name"]), Version: stringFromAny(vm["name"]), Published: timeFromAny(vm["releaseDate"]), Installable: d.Installable})
		}
	}
	return d, nil
}

func (a *App) builtByBitDetails(ctx context.Context, id string) (ProjectDetails, error) {
	a.mu.RLock()
	key := strings.TrimSpace(a.settings.BuiltByBitAPIKey)
	oauthReady := strings.TrimSpace(a.settings.BuiltByBitOAuthToken) != ""
	a.mu.RUnlock()
	if key == "" {
		return ProjectDetails{}, errors.New("BuiltByBit API token is not configured in Settings")
	}
	base := providerBase("MMV_BUILTBYBIT_API_BASE", "https://api.builtbybit.com")
	values := url.Values{"resource_ids": {id}, "per_page": {"1"}, "with": {"Creator,Description,Category,LatestVersion,LatestReviews"}}
	var resp struct {
		Data struct {
			Resources []map[string]any `json:"resources"`
		} `json:"data"`
	}
	if err := a.getJSON(ctx, base+"/v2/resources/discover/resources?"+values.Encode(), map[string]string{"Authorization": key}, &resp); err != nil {
		return ProjectDetails{}, err
	}
	if len(resp.Data.Resources) == 0 {
		return ProjectDetails{}, errors.New("BuiltByBit resource not found")
	}
	m := resp.Data.Resources[0]
	d := ProjectDetails{ID: id, Slug: id, ProjectType: inferProjectType(stringFromAny(m["title"])+" "+stringFromAny(m["summary"]), "plugin"), Title: stringFromAny(m["title"]), Summary: stringFromAny(m["summary"]), Description: nestedString(m, "Description", "raw"), Downloads: int64FromAny(m["downloads"]), Followers: int64FromAny(m["purchases"]), Updated: timeFromAny(m["last_updated_at"]), Published: timeFromAny(m["published_at"]), PageURL: absoluteURL("https://builtbybit.com", stringFromAny(m["url"])), Installable: oauthReady, InstallMode: "licensed-download-api", Categories: nonEmptyStrings(nestedString(m, "Category", "title"))}
	cover := stringFromAny(m["cover_image_url"])
	d.Gallery = uniqueStringsPreserve(append(nonEmptyStrings(cover), stringSliceFromAny(m["carousel_image_urls"])...))
	d.IconURL = firstString(d.Gallery)
	if creator, ok := m["Creator"].(map[string]any); ok {
		d.Authors = append(d.Authors, ProjectAuthor{Name: stringFromAny(creator["username"]), AvatarURL: stringFromAny(creator["avatar_url"])})
	}
	if latest, ok := m["LatestVersion"].(map[string]any); ok {
		d.Versions = append(d.Versions, ProjectVersionSummary{ID: stringFromAny(latest["version_id"]), Version: stringFromAny(latest["version_string"]), Published: timeFromAny(latest["created_at"]), Installable: oauthReady})
	}
	return d, nil
}

func (a *App) atLauncherDetails(ctx context.Context, id string) (ProjectDetails, error) {
	endpoint := providerBase("MMV_ATLAUNCHER_GRAPHQL", "https://api.atlauncher.com/v2/graphql")
	q := `query GetPackBySafeName($safeName: String!) { pack(safeName: $safeName) { id position name safeName latestVersion { id version minecraftVersion changelog isRecommended canUpdate createdAt updatedAt publishedAt } versions(first: 25) { id version minecraftVersion changelog isRecommended canUpdate createdAt updatedAt publishedAt } } }`
	var resp struct {
		Data struct {
			Pack struct {
				ID       int64  `json:"id"`
				Name     string `json:"name"`
				SafeName string `json:"safeName"`
				Latest   struct {
					ID               int64  `json:"id"`
					Version          string `json:"version"`
					MinecraftVersion string `json:"minecraftVersion"`
					Changelog        string `json:"changelog"`
					UpdatedAt        string `json:"updatedAt"`
					PublishedAt      string `json:"publishedAt"`
				} `json:"latestVersion"`
				Versions []struct {
					ID               int64  `json:"id"`
					Version          string `json:"version"`
					MinecraftVersion string `json:"minecraftVersion"`
					Changelog        string `json:"changelog"`
					UpdatedAt        string `json:"updatedAt"`
					PublishedAt      string `json:"publishedAt"`
				} `json:"versions"`
			} `json:"pack"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := a.postJSON(ctx, endpoint, nil, map[string]any{"query": q, "variables": map[string]any{"safeName": id}}, &resp); err != nil {
		return ProjectDetails{}, err
	}
	if len(resp.Errors) > 0 {
		return ProjectDetails{}, errors.New(resp.Errors[0].Message)
	}
	p := resp.Data.Pack
	if p.SafeName == "" {
		return ProjectDetails{}, errors.New("ATLauncher pack not found")
	}
	d := ProjectDetails{ID: strconv.FormatInt(p.ID, 10), Slug: p.SafeName, ProjectType: "modpack", Title: p.Name, Summary: cleanHTMLText(p.Latest.Changelog), Description: cleanHTMLText(p.Latest.Changelog), Updated: firstNonEmpty(p.Latest.UpdatedAt, p.Latest.PublishedAt), Categories: []string{"modpack", "atlauncher"}, GameVersions: nonEmptyStrings(p.Latest.MinecraftVersion), PageURL: "https://atlauncher.com/pack/" + url.PathEscape(p.SafeName), Installable: false, InstallMode: "integrated-modpack-manifest"}
	for _, v := range p.Versions {
		d.GameVersions = append(d.GameVersions, v.MinecraftVersion)
		d.Versions = append(d.Versions, ProjectVersionSummary{ID: strconv.FormatInt(v.ID, 10), Version: v.Version, Published: firstNonEmpty(v.PublishedAt, v.UpdatedAt), GameVersions: nonEmptyStrings(v.MinecraftVersion), Changelog: cleanHTMLText(v.Changelog)})
	}
	return d, nil
}

func (a *App) ftbDetails(ctx context.Context, id string) (ProjectDetails, error) {
	base := providerBase("MMV_FTB_API_BASE", "https://api.modpacks.ch/public")
	var m map[string]any
	if err := a.getJSON(ctx, base+"/modpack/"+url.PathEscape(id), nil, &m); err != nil {
		return ProjectDetails{}, err
	}
	d := ProjectDetails{ID: id, Slug: id, ProjectType: "modpack", Title: firstNonEmpty(stringFromAny(m["name"]), "FTB Pack "+id), Summary: firstNonEmpty(stringFromAny(m["synopsis"]), stringFromAny(m["description"])), Description: firstNonEmpty(stringFromAny(m["description"]), stringFromAny(m["synopsis"])), IconURL: firstNonEmpty(stringFromAny(m["art"]), stringFromAny(m["icon"])), Categories: []string{"modpack", "ftb"}, PageURL: "https://www.feed-the-beast.com/modpacks/" + id, Installable: false, InstallMode: "integrated-modpack-manifest"}
	for _, x := range anySlice(m["versions"]) {
		if vm, ok := x.(map[string]any); ok {
			pv := ProjectVersionSummary{ID: stringFromAny(vm["id"]), Version: stringFromAny(vm["name"]), Published: timeFromAny(vm["updated"]), GameVersions: nonEmptyStrings(firstNonEmpty(stringFromAny(vm["mcVersion"]), stringFromAny(vm["minecraftVersion"])))}
			d.Versions = append(d.Versions, pv)
			d.GameVersions = append(d.GameVersions, pv.GameVersions...)
		}
	}
	return d, nil
}

func (a *App) technicDetails(ctx context.Context, id string) (ProjectDetails, error) {
	base := providerBase("MMV_TECHNIC_API_BASE", "https://api.technicpack.net")
	var m map[string]any
	if err := a.getJSON(ctx, base+"/modpack/"+url.PathEscape(id)+"?build=822", nil, &m); err != nil {
		return ProjectDetails{}, err
	}
	d := ProjectDetails{ID: id, Slug: id, ProjectType: "modpack", Title: firstNonEmpty(stringFromAny(m["displayName"]), stringFromAny(m["name"]), titleFromSlug(id)), Summary: firstNonEmpty(stringFromAny(m["description"]), stringFromAny(m["tag"])), Description: stringFromAny(m["description"]), IconURL: firstNonEmpty(stringFromAny(m["icon"]), stringFromAny(m["logo"])), Gallery: nonEmptyStrings(stringFromAny(m["background"]), stringFromAny(m["logo"])), Categories: []string{"modpack", "technic"}, PageURL: "https://www.technicpack.net/modpack/" + url.PathEscape(id), Installable: false, InstallMode: "integrated-modpack-manifest"}
	if author := firstNonEmpty(stringFromAny(m["user"]), stringFromAny(m["author"])); author != "" {
		d.Authors = append(d.Authors, ProjectAuthor{Name: author})
	}
	return d, nil
}

func (a *App) nexusDetails(ctx context.Context, id string) (ProjectDetails, error) {
	a.mu.RLock()
	key := strings.TrimSpace(a.settings.NexusAPIKey)
	a.mu.RUnlock()
	if key == "" {
		return ProjectDetails{}, errors.New("Nexus Mods API key is not configured in Settings")
	}
	base := providerBase("MMV_NEXUS_API_BASE", "https://api.nexusmods.com/v1")
	headers := map[string]string{"apikey": key, "Application-Name": appName, "Application-Version": appVersion}
	var m map[string]any
	if err := a.getJSON(ctx, base+"/games/minecraft/mods/"+url.PathEscape(id)+".json", headers, &m); err != nil {
		return ProjectDetails{}, err
	}
	img := firstNonEmpty(stringFromAny(m["picture_url"]), stringFromAny(m["mod_picture_url"]))
	d := ProjectDetails{ID: id, Slug: id, ProjectType: "mod", Title: stringFromAny(m["name"]), Summary: stringFromAny(m["summary"]), Description: cleanHTMLText(stringFromAny(m["description"])), IconURL: img, Gallery: nonEmptyStrings(img), Authors: []ProjectAuthor{{Name: stringFromAny(m["author"])}}, Followers: int64FromAny(m["endorsement_count"]), Updated: timeFromAny(firstNonNil(m["updated_timestamp"], m["updated_time"])), Published: timeFromAny(firstNonNil(m["created_timestamp"], m["created_time"])), Categories: []string{"mod", "nexusmods"}, PageURL: "https://www.nexusmods.com/minecraft/mods/" + id, Installable: false, InstallMode: "integrated-metadata"}
	return d, nil
}

func (a *App) genericWebDetails(ctx context.Context, provider, pageURL, id string) (ProjectDetails, error) {
	if pageURL == "" {
		if p := providerInfoByID(provider); p != nil {
			pageURL = p.HomeURL
		}
	}
	if pageURL == "" {
		return ProjectDetails{}, errors.New("project URL is unavailable")
	}
	body, err := a.getText(ctx, pageURL, nil)
	if err != nil {
		return ProjectDetails{}, err
	}
	title := firstNonEmpty(metaContent(body, "og:title"), matchText(body, `(?is)<h1[^>]*>(.*?)</h1>`), matchText(body, `(?is)<title[^>]*>(.*?)</title>`))
	summary := firstNonEmpty(metaContent(body, "og:description"), metaContent(body, "description"))
	desc := firstNonEmpty(articleText(body), summary)
	gallery := imagesFromChunkWithBase(body, pageURL, 18)
	iconURL := firstNonEmpty(metaContent(body, "og:image"), firstString(gallery))
	author := firstNonEmpty(metaContent(body, "author"), matchText(body, `(?is)(?:by|author)\s*</?[^>]*>?\s*([^<\n]{2,80})`))
	ptype := inferProjectType(title+" "+summary, defaultProviderType(provider))
	if fromURL := inferProjectTypeFromPageURL(provider, pageURL); fromURL != "" {
		ptype = fromURL
	}
	d := ProjectDetails{ID: firstNonEmpty(id, pageURL), Slug: id, ProjectType: ptype, Title: title, Summary: summary, Description: desc, IconURL: iconURL, Gallery: gallery, Categories: []string{provider}, PageURL: pageURL, Installable: false, InstallMode: "detected-downloads", Links: map[string]string{}}
	if author != "" {
		d.Authors = append(d.Authors, ProjectAuthor{Name: author})
	}
	for _, link := range downloadLinks(body, pageURL, 30) {
		d.Links["Download "+strconv.Itoa(len(d.Links)+1)] = link
	}
	if providerSupportsVerifiedDetectedInstall(provider) && len(d.Links) > 0 {
		d.Installable = true
		d.InstallMode = "verified-detected-download"
	}
	return d, nil
}

func metaContent(body, key string) string {
	key = regexp.QuoteMeta(key)
	patterns := []string{
		`(?is)<meta[^>]+(?:property|name)=["']` + key + `["'][^>]+content=["']([^"']+)["']`,
		`(?is)<meta[^>]+content=["']([^"']+)["'][^>]+(?:property|name)=["']` + key + `["']`,
	}
	for _, pattern := range patterns {
		m := regexp.MustCompile(pattern).FindStringSubmatch(body)
		if len(m) > 1 {
			return cleanHTMLText(html.UnescapeString(m[1]))
		}
	}
	return ""
}

func articleText(body string) string {
	for _, re := range []*regexp.Regexp{
		regexp.MustCompile(`(?is)<article[^>]*>(.*?)</article>`),
		regexp.MustCompile(`(?is)<main[^>]*>(.*?)</main>`),
		regexp.MustCompile(`(?is)<div[^>]+class=["'][^"']*(?:description|content|post-body)[^"']*["'][^>]*>(.*?)</div>`),
	} {
		if m := re.FindStringSubmatch(body); len(m) > 1 {
			text := cleanHTMLText(m[1])
			if len(text) > 12000 {
				text = text[:12000]
			}
			if len(text) > 80 {
				return text
			}
		}
	}
	return ""
}

func downloadLinks(body, pageURL string, limit int) []string {
	seen := map[string]bool{}
	out := []string{}
	re := regexp.MustCompile(`(?is)<a[^>]+href=["']([^"']+)["'][^>]*>(.*?)</a>`)
	for _, m := range re.FindAllStringSubmatch(body, -1) {
		if len(m) < 3 {
			continue
		}
		href := absoluteURL(pageURL, m[1])
		text := strings.ToLower(cleanHTMLText(m[2]))
		low := strings.ToLower(href)
		if href == "" || seen[href] || (!strings.Contains(text, "download") && !strings.Contains(low, "/download") && !strings.Contains(low, "/modificationdl/") && !strings.HasSuffix(low, ".jar") && !strings.HasSuffix(low, ".zip") && !strings.HasSuffix(low, ".mcpack") && !strings.HasSuffix(low, ".mcaddon") && !strings.HasSuffix(low, ".mcworld") && !strings.HasSuffix(low, ".mrpack")) {
			continue
		}
		seen[href] = true
		out = append(out, href)
		if len(out) >= limit {
			break
		}
	}
	return out
}

func inferProjectTypeFromPageURL(provider, pageURL string) string {
	low := strings.ToLower(pageURL)
	if provider == "minecrafthub" {
		if u, err := url.Parse(pageURL); err == nil {
			if ptype := minecrafthubPathType(u.Path); ptype != "" {
				return ptype
			}
		}
	}
	if provider == "curseforge" {
		switch {
		case strings.Contains(low, "/minecraft-bedrock/addons/"), strings.Contains(low, "/minecraft-bedrock/scripts/"):
			return "addon"
		case strings.Contains(low, "/minecraft-bedrock/maps/"), strings.Contains(low, "/minecraft/worlds/"):
			return "world"
		case strings.Contains(low, "/texture-packs/"):
			return "resourcepack"
		case strings.Contains(low, "/minecraft-bedrock/skins/"):
			return "skin"
		case strings.Contains(low, "/minecraft/modpacks/"):
			return "modpack"
		case strings.Contains(low, "/minecraft/shaders/"):
			return "shader"
		case strings.Contains(low, "/minecraft/data-packs/"):
			return "datapack"
		case strings.Contains(low, "/minecraft/bukkit-plugins/"):
			return "plugin"
		case strings.Contains(low, "/minecraft/mc-addons/"):
			return "addon"
		case strings.Contains(low, "/minecraft/customization/"):
			return "tool"
		case strings.Contains(low, "/minecraft/mc-mods/"):
			return "mod"
		}
	}
	if provider == "planetminecraft" {
		switch {
		case strings.Contains(low, "/data-pack/"):
			return "datapack"
		case strings.Contains(low, "/texture-pack/"):
			return "resourcepack"
		case strings.Contains(low, "/project/"):
			return "world"
		case strings.Contains(low, "/skin/"):
			return "skin"
		}
	}
	switch provider {
	case "skindex":
		return "skin"
	case "shaderpackscom", "shaderpacksnet", "minecraftshader":
		return "shader"
	case "mcreator", "minecrafthub":
		return "mod"
	}
	return ""
}

func defaultProviderType(provider string) string {
	switch provider {
	case "mcpedl", "marketplace":
		return "addon"
	case "bukkitdev", "spigot", "hangar", "builtbybit", "spongeore", "polymart":
		return "plugin"
	case "minecraftmaps":
		return "world"
	case "resourcepacknet", "texturepacks":
		return "resourcepack"
	case "shaderpackscom", "shaderpacksnet", "minecraftshader":
		return "shader"
	case "skindex":
		return "skin"
	case "mcreator":
		return "mod"
	case "atlauncher", "ftb", "technic":
		return "modpack"
	default:
		return "mod"
	}
}

func cleanMarkdownForPreview(s string) string {
	s = regexp.MustCompile(`(?m)^#{1,6}\s*`).ReplaceAllString(s, "")
	s = regexp.MustCompile(`!\[[^\]]*\]\([^\)]+\)`).ReplaceAllString(s, "")
	s = regexp.MustCompile(`\[([^\]]+)\]\([^\)]+\)`).ReplaceAllString(s, "$1")
	s = strings.ReplaceAll(s, "```", "")
	if len(s) > 18000 {
		s = s[:18000]
	}
	return strings.TrimSpace(s)
}

func anySlice(v any) []any {
	if xs, ok := v.([]any); ok {
		return xs
	}
	return nil
}

func mapStringSlices(v any) map[string][]string {
	out := map[string][]string{}
	if m, ok := v.(map[string]any); ok {
		for k, raw := range m {
			out[k] = stringSliceFromAny(raw)
		}
	}
	return out
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
