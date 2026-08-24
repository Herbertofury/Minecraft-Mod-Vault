package main

import (
	"context"
	"errors"
	"fmt"
	"html"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"sync"
)

// MinecraftHub is used as a curated discovery/enrichment lane. Vault never
// treats the directory as a download mirror: installs are resolved back to the
// original creator/provider and then delegated to the corresponding first-class
// provider integration inside Vault.

type minecraftHubLane struct {
	Path        string
	ProjectType string
}

var minecraftHubLanes = []minecraftHubLane{
	{Path: "/mods", ProjectType: "mod"},
	{Path: "/resource-packs", ProjectType: "resourcepack"},
	{Path: "/shaders", ProjectType: "shader"},
	{Path: "/addons", ProjectType: "addon"},
	{Path: "/modpacks", ProjectType: "modpack"},
	{Path: "/datapacks", ProjectType: "datapack"},
	{Path: "/plugins", ProjectType: "plugin"},
	{Path: "/maps", ProjectType: "world"},
	{Path: "/skins", ProjectType: "skin"},
}

func minecraftHubLanesForType(projectType string) []minecraftHubLane {
	projectType = strings.ToLower(strings.TrimSpace(projectType))
	if projectType == "" || projectType == "all" {
		return append([]minecraftHubLane(nil), minecraftHubLanes...)
	}
	out := make([]minecraftHubLane, 0, 2)
	for _, lane := range minecraftHubLanes {
		if projectTypeMatches(lane.ProjectType, projectType) || projectTypeMatches(projectType, lane.ProjectType) {
			out = append(out, lane)
		}
	}
	return out
}

func minecrafthubPathType(path string) string {
	path = "/" + strings.Trim(strings.ToLower(strings.TrimSpace(path)), "/") + "/"
	switch {
	case strings.HasPrefix(path, "/mods/"):
		return "mod"
	case strings.HasPrefix(path, "/resource-packs/"):
		return "resourcepack"
	case strings.HasPrefix(path, "/shaders/"):
		return "shader"
	case strings.HasPrefix(path, "/addons/"):
		return "addon"
	case strings.HasPrefix(path, "/modpacks/"):
		return "modpack"
	case strings.HasPrefix(path, "/datapacks/"):
		return "datapack"
	case strings.HasPrefix(path, "/plugins/"):
		return "plugin"
	case strings.HasPrefix(path, "/maps/"):
		return "world"
	case strings.HasPrefix(path, "/skins/"):
		return "skin"
	}
	return ""
}

func minecrafthubResourceLinkRE() *regexp.Regexp {
	return regexp.MustCompile(`(?is)<a[^>]+href=["']((?:https?://[^"']*minecrafthub\.io)?/(?:mods|resource-packs|shaders|addons|modpacks|datapacks|plugins|maps|skins)/[a-z0-9][a-z0-9-]*/?)["'][^>]*>(.*?)</a>`)
}

func (a *App) searchMinecraftHub(ctx context.Context, query string, opts providerSearchOptions) ([]UnifiedProject, error) {
	base := providerBase("MMV_MINECRAFTHUB_BASE", "https://minecrafthub.io")
	lanes := minecraftHubLanesForType(opts.ProjectType)
	if len(lanes) == 0 {
		return nil, nil
	}

	type laneResult struct {
		items []UnifiedProject
		err   error
	}
	ch := make(chan laneResult, len(lanes))
	var wg sync.WaitGroup
	for _, lane := range lanes {
		lane := lane
		wg.Add(1)
		go func() {
			defer wg.Done()
			items, err := a.searchMinecraftHubLane(ctx, base, lane, query, opts)
			ch <- laneResult{items: items, err: err}
		}()
	}
	go func() { wg.Wait(); close(ch) }()

	all := []UnifiedProject{}
	var errs []string
	for result := range ch {
		if result.err != nil {
			errs = append(errs, result.err.Error())
			continue
		}
		all = append(all, result.items...)
	}
	if len(all) == 0 && len(errs) == len(lanes) {
		return nil, errors.New(strings.Join(uniqueStringsPreserve(errs), "; "))
	}
	all = mergeProviderDuplicates(all)
	orderProjectsByQuery(all, query)
	limit := providerResultLimit(opts, 80)
	if limit > 0 && len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

func (a *App) searchMinecraftHubLane(ctx context.Context, base string, lane minecraftHubLane, query string, opts providerSearchOptions) ([]UnifiedProject, error) {
	// The largest lane currently needs multiple pages. Keep the scan bounded and
	// stop as soon as a page adds no new resource URLs. This gives federated search
	// a much broader window than the site's first page while avoiding an unbounded crawl.
	maxPages := 1
	if lane.ProjectType == "mod" {
		maxPages = 4
	} else if lane.ProjectType == "all" {
		maxPages = 3
	}
	seen := map[string]bool{}
	items := []UnifiedProject{}
	var lastErr error
	for page := 1; page <= maxPages; page++ {
		endpoint := strings.TrimRight(base, "/") + lane.Path
		if page > 1 {
			endpoint += "?page=" + fmt.Sprint(page)
		}
		body, err := a.getText(ctx, endpoint, nil)
		if err != nil {
			lastErr = err
			if len(items) > 0 {
				break
			}
			continue
		}
		pageItems := parseGenericCards(body, "minecrafthub", base, minecrafthubResourceLinkRE(), lane.ProjectType, 120)
		added := 0
		for _, item := range pageItems {
			if seen[item.PageURL] {
				continue
			}
			seen[item.PageURL] = true
			if ptype := inferProjectTypeFromPageURL("minecrafthub", item.PageURL); ptype != "" {
				item.ProjectType = ptype
			}
			if !projectTypeMatches(item.ProjectType, lane.ProjectType) {
				continue
			}
			item.Categories = uniqueStringsPreserve(append(item.Categories, "minecrafthub", "curated", item.ProjectType))
			item.Reason = "Curated source-linked MinecraftHub index; original provider resolved inside Vault"
			// Search cards do not promise a one-click install until the rich project
			// page has resolved an original provider URL. The detail view promotes
			// the button only after that route is proven.
			item.Installable = false
			item.InstallMode = "canonical-provider-resolution"
			if v := matchMinecraftVersion(item.Title + " " + item.Summary); v != "" {
				item.Versions = uniqueStringsPreserve(append(item.Versions, v))
			}
			for _, loader := range detectedLoaderNames(item.Title + " " + item.Summary) {
				item.Loaders = append(item.Loaders, loader)
			}
			item.Loaders = uniqueStrings(item.Loaders)
			items = append(items, item)
			added++
		}
		if added == 0 {
			break
		}
	}
	if len(items) == 0 && lastErr != nil {
		return nil, lastErr
	}
	return trimProjectsToQuery(items, query, maxInt(providerResultLimit(opts, 80), 48)), nil
}

func detectedLoaderNames(text string) []string {
	low := strings.ToLower(text)
	out := []string{}
	for _, loader := range []string{"fabric", "neoforge", "forge", "quilt", "iris", "optifine", "paper", "spigot", "bukkit", "purpur", "folia", "velocity", "bungeecord", "sponge"} {
		if strings.Contains(low, loader) {
			out = append(out, loader)
		}
	}
	return uniqueStrings(out)
}

func minecraftHubOriginalSource(body, pageURL string) (provider, id, canonical string, ok bool) {
	anchorRE := regexp.MustCompile(`(?is)<a[^>]+href=["']([^"']+)["'][^>]*>(.*?)</a>`)
	type candidate struct {
		provider  string
		id        string
		canonical string
		score     int
	}
	candidates := []candidate{}
	for _, m := range anchorRE.FindAllStringSubmatch(body, -1) {
		if len(m) < 3 {
			continue
		}
		raw := absoluteURL(pageURL, html.UnescapeString(strings.TrimSpace(m[1])))
		p, pid, normalized, found := integratedProviderFromURL(raw)
		if !found || p == "minecrafthub" {
			continue
		}
		text := strings.ToLower(cleanHTMLText(m[2]))
		score := 10
		if strings.Contains(text, "visit") || strings.Contains(text, "original") || strings.Contains(text, "source") {
			score += 30
		}
		if strings.Contains(text, p) {
			score += 15
		}
		if strings.Contains(strings.ToLower(body), "source platform") {
			score += 2
		}
		candidates = append(candidates, candidate{provider: p, id: pid, canonical: normalized, score: score})
	}
	if len(candidates) == 0 {
		return "", "", "", false
	}
	sort.SliceStable(candidates, func(i, j int) bool { return candidates[i].score > candidates[j].score })
	c := candidates[0]
	return c.provider, c.id, c.canonical, true
}

func integratedProviderFromURL(raw string) (provider, id, canonical string, ok bool) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme != "https" || u.Hostname() == "" {
		return "", "", "", false
	}
	host := strings.TrimPrefix(strings.ToLower(u.Hostname()), "www.")
	path := strings.Trim(u.EscapedPath(), "/")
	parts := strings.Split(path, "/")
	last := func() string {
		if len(parts) == 0 {
			return ""
		}
		v, _ := url.PathUnescape(parts[len(parts)-1])
		return strings.TrimSpace(v)
	}
	canonical = u.Scheme + "://" + u.Host + u.EscapedPath()
	if u.RawQuery != "" && (host == "github.com" || host == "mcpedl.com") {
		canonical += "?" + u.RawQuery
	}
	switch {
	case host == "minecrafthub.io":
		return "minecrafthub", last(), canonical, true
	case host == "modrinth.com":
		if len(parts) >= 2 && containsFold([]string{"mod", "shader", "resourcepack", "datapack", "modpack", "plugin"}, parts[0]) {
			return "modrinth", last(), canonical, true
		}
	case host == "curseforge.com":
		if strings.Contains("/"+path+"/", "/minecraft/") || strings.Contains("/"+path+"/", "/minecraft-bedrock/") {
			return "curseforge", last(), canonical, true
		}
	case host == "github.com":
		if len(parts) >= 2 && parts[0] != "topics" && parts[0] != "search" {
			owner, _ := url.PathUnescape(parts[0])
			repo, _ := url.PathUnescape(parts[1])
			return "github", owner + "/" + strings.TrimSuffix(repo, ".git"), canonical, true
		}
	case host == "mcpedl.com":
		return "mcpedl", last(), canonical, last() != ""
	case host == "planetminecraft.com":
		return "planetminecraft", last(), canonical, last() != ""
	case host == "hangar.papermc.io":
		return "hangar", last(), canonical, last() != ""
	case host == "spigotmc.org":
		if m := regexp.MustCompile(`(?:^|/)resources/(?:[^/]*\.)?(\d+)(?:/|$)`).FindStringSubmatch("/" + path + "/"); len(m) > 1 {
			return "spigot", m[1], canonical, true
		}
	case host == "dev.bukkit.org":
		return "bukkitdev", last(), canonical, last() != ""
	case host == "ore.spongepowered.org":
		return "spongeore", last(), canonical, last() != ""
	case host == "smithed.dev":
		return "smithed", last(), canonical, last() != ""
	case host == "builtbybit.com":
		if m := regexp.MustCompile(`(?:^|/)(?:resources/)?[^/.]*\.(\d+)(?:/|$)`).FindStringSubmatch("/" + path + "/"); len(m) > 1 {
			return "builtbybit", m[1], canonical, true
		}
	case host == "polymart.org":
		return "polymart", last(), canonical, last() != ""
	case host == "nexusmods.com":
		if m := regexp.MustCompile(`(?:^|/)minecraft/mods/(\d+)(?:/|$)`).FindStringSubmatch("/" + path + "/"); len(m) > 1 {
			return "nexusmods", m[1], canonical, true
		}
	case host == "mcreator.net":
		return "mcreator", last(), canonical, last() != ""
	case host == "resourcepack.net":
		return "resourcepacknet", last(), canonical, last() != ""
	case host == "texture-packs.com":
		return "texturepacks", last(), canonical, last() != ""
	case host == "shaderpacks.com":
		return "shaderpackscom", last(), canonical, last() != ""
	case host == "shaderpacks.net":
		return "shaderpacksnet", last(), canonical, last() != ""
	case host == "minecraftshader.com":
		return "minecraftshader", last(), canonical, last() != ""
	case host == "minecraftmaps.com":
		return "minecraftmaps", last(), canonical, last() != ""
	case host == "minecraftskins.com":
		if m := regexp.MustCompile(`(?:^|/)skin/(\d+)(?:/|$)`).FindStringSubmatch("/" + path + "/"); len(m) > 1 {
			return "skindex", m[1], canonical, true
		}
	}
	return "", "", "", false
}

func minecraftHubVersionsFromText(text string) []string {
	re := regexp.MustCompile(`\b(?:1\.(?:[0-9]{1,2})(?:\.[0-9]{1,2})?|[2-9][0-9]\.[0-9]+(?:\.[0-9]+)?)\b`)
	return uniqueStringsPreserve(re.FindAllString(text, -1))
}

func (a *App) minecraftHubDetails(ctx context.Context, id, pageURL string) (ProjectDetails, error) {
	if strings.TrimSpace(pageURL) == "" {
		return ProjectDetails{}, errors.New("MinecraftHub project URL is required")
	}
	body, err := a.getText(ctx, pageURL, nil)
	if err != nil {
		return ProjectDetails{}, err
	}
	genericTitle := firstNonEmpty(metaContent(body, "og:title"), matchText(body, `(?is)<h1[^>]*>(.*?)</h1>`), titleFromSlug(id))
	summary := firstNonEmpty(metaContent(body, "og:description"), metaContent(body, "description"), matchText(body, `(?is)<h1[^>]*>.*?</h1>\s*<p[^>]*>(.*?)</p>`))
	description := firstNonEmpty(articleText(body), summary)
	gallery := imagesFromChunkWithBase(body, pageURL, 24)
	iconURL := firstNonEmpty(metaContent(body, "og:image"), firstString(gallery))
	ptype := inferProjectTypeFromPageURL("minecrafthub", pageURL)
	if ptype == "" {
		ptype = inferProjectType(genericTitle+" "+summary, "mod")
	}

	author := firstNonEmpty(
		matchText(body, `(?is)\bBy\s+([^<\n]{2,100}?)\s+via\s+[A-Za-z0-9 ._-]+\.?\s*<`),
		matchText(body, `(?is)\bCreator\s*</[^>]+>\s*<[^>]+>\s*([^<\n]{2,100})`),
		metaContent(body, "author"),
	)
	plain := cleanHTMLText(body)
	versions := minecraftHubVersionsFromText(firstNonEmpty(matchText(body, `(?is)Versions\s*</[^>]+>\s*([^<]{2,500})`), plain))
	loaders := detectedLoaderNames(plain)
	provider, originalID, originalURL, hasOriginal := minecraftHubOriginalSource(body, pageURL)

	d := ProjectDetails{
		ID: firstNonEmpty(id, strings.Trim(filepathBaseFromURL(pageURL), "/")), Slug: firstNonEmpty(id, strings.Trim(filepathBaseFromURL(pageURL), "/")),
		ProjectType: ptype, Title: genericTitle, Summary: summary, Description: description,
		IconURL: iconURL, Gallery: gallery, Categories: []string{"minecrafthub", "curated", ptype},
		GameVersions: versions, Loaders: loaders, PageURL: pageURL,
		Installable: hasOriginal && providerInstallDelegationSupported(provider), InstallMode: "canonical-provider-resolution", Links: map[string]string{},
	}
	if author != "" {
		d.Authors = append(d.Authors, ProjectAuthor{Name: strings.TrimSuffix(strings.TrimSpace(author), "."), Role: "Creator"})
	}
	if hasOriginal {
		d.Links["Original source"] = originalURL
		d.Categories = uniqueStringsPreserve(append(d.Categories, provider, "source-linked"))
		if originalID != "" {
			d.Links["Resolved project"] = provider + ":" + originalID
		}
	}
	if checked := matchText(body, `(?is)Last\s+checked\s*</[^>]+>\s*<[^>]+>\s*([^<]+)`); checked != "" {
		d.Updated = cleanHTMLText(checked)
	}
	return d, nil
}

func filepathBaseFromURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

func providerInstallDelegationSupported(provider string) bool {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "modrinth", "curseforge", "github", "hangar", "spigot", "builtbybit", "smithed", "spongeore", "planetminecraft", "mcpedl", "bukkitdev", "moddb", "minecraftmaps", "resourcepacknet", "texturepacks", "mcreator", "shaderpackscom", "shaderpacksnet", "minecraftshader", "skindex":
		return true
	default:
		return false
	}
}

func (a *App) installMinecraftHubResolved(ctx context.Context, pageURL, id, gameVersion, loader, target string) (any, error) {
	if strings.TrimSpace(pageURL) == "" {
		return nil, errors.New("MinecraftHub project page is required")
	}
	body, err := a.getText(ctx, pageURL, nil)
	if err != nil {
		return nil, err
	}
	provider, projectID, canonical, ok := minecraftHubOriginalSource(body, pageURL)
	if !ok {
		return nil, errors.New("MinecraftHub entry did not expose an original creator/provider URL that Vault can resolve")
	}
	if !providerInstallDelegationSupported(provider) {
		return nil, fmt.Errorf("original provider %q is indexed but does not expose a trustworthy in-app install route", provider)
	}
	request := ProviderInstallRequest{Provider: provider, ID: projectID, Slug: projectID, GameVersion: gameVersion, Loader: loader, Target: target, PageURL: canonical}
	return a.installProviderRequest(ctx, request)
}
