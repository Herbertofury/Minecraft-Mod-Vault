package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

func providerBase(envKey, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(envKey)); v != "" {
		return strings.TrimRight(v, "/")
	}
	return strings.TrimRight(fallback, "/")
}

func (a *App) searchHangar(ctx context.Context, query string, opts providerSearchOptions) ([]UnifiedProject, error) {
	base := providerBase("MMV_HANGAR_API_BASE", "https://hangar.papermc.io/api/v1")
	limit := opts.Limit
	if limit <= 0 || limit > 25 {
		limit = 25
	}
	values := url.Values{}
	values.Set("limit", strconv.Itoa(limit))
	values.Set("offset", strconv.Itoa(maxInt(opts.Offset, 0)))
	if strings.TrimSpace(query) != "" && !strings.EqualFold(query, "minecraft") {
		values.Set("query", query)
	}
	type hangarProject struct {
		ID          int64  `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Category    string `json:"category"`
		LastUpdated string `json:"lastUpdated"`
		AvatarURL   string `json:"avatarUrl"`
		Namespace   struct {
			Owner string `json:"owner"`
			Slug  string `json:"slug"`
		} `json:"namespace"`
		Stats struct {
			Downloads int64 `json:"downloads"`
			Stars     int64 `json:"stars"`
			Watchers  int64 `json:"watchers"`
		} `json:"stats"`
	}
	var resp struct {
		Result []hangarProject `json:"result"`
	}
	if err := a.getJSON(ctx, base+"/projects?"+values.Encode(), nil, &resp); err != nil {
		return nil, err
	}
	out := make([]UnifiedProject, 0, len(resp.Result))
	for _, p := range resp.Result {
		slug := p.Namespace.Slug
		if slug == "" {
			slug = strings.ToLower(strings.ReplaceAll(p.Name, " ", "-"))
		}
		page := "https://hangar.papermc.io/" + url.PathEscape(p.Namespace.Owner) + "/" + url.PathEscape(slug)
		out = append(out, UnifiedProject{
			ID: p.Namespace.Owner + "/" + slug, Provider: "hangar", ProjectType: "plugin", Slug: slug,
			Title: p.Name, Summary: p.Description, Author: p.Namespace.Owner, IconURL: p.AvatarURL,
			Downloads: p.Stats.Downloads, Followers: p.Stats.Stars + p.Stats.Watchers, DateUpdated: p.LastUpdated,
			Categories: uniqueStringsPreserve([]string{p.Category, "plugin", "paper"}), PageURL: page,
			Installable: true, InstallMode: "hangar-version", Reason: "Native Hangar API project",
		})
	}
	return out, nil
}

func (a *App) searchSpigot(ctx context.Context, query string, opts providerSearchOptions) ([]UnifiedProject, error) {
	base := providerBase("MMV_SPIGET_API_BASE", "https://api.spiget.org/v2")
	limit := opts.Limit
	if limit <= 0 || limit > 50 {
		limit = 30
	}
	values := url.Values{}
	values.Set("size", strconv.Itoa(limit))
	values.Set("page", strconv.Itoa(maxInt(opts.Offset/limit, 0)))
	values.Set("sort", "-downloads")
	target := base + "/search/resources/" + url.PathEscape(query) + "?" + values.Encode()
	var raw []map[string]any
	if err := a.getJSON(ctx, target, nil, &raw); err != nil {
		return nil, err
	}
	out := make([]UnifiedProject, 0, len(raw))
	for _, m := range raw {
		id := int64FromAny(m["id"])
		name := stringFromAny(m["name"])
		if name == "" || id == 0 {
			continue
		}
		author := nestedString(m, "author", "name")
		if author == "" {
			author = nestedString(m, "author", "id")
		}
		iconURL := ""
		if icon := nestedString(m, "icon", "url"); icon != "" {
			iconURL = absoluteURL("https://www.spigotmc.org", icon)
		}
		updated := timeFromAny(m["updateDate"])
		page := "https://www.spigotmc.org/resources/" + strconv.FormatInt(id, 10) + "/"
		out = append(out, UnifiedProject{
			ID: strconv.FormatInt(id, 10), Provider: "spigot", ProjectType: "plugin", Slug: strconv.FormatInt(id, 10),
			Title: name, Summary: firstNonEmpty(stringFromAny(m["tag"]), stringFromAny(m["description"])), Author: author,
			IconURL: iconURL, Downloads: int64FromAny(m["downloads"]), Followers: int64FromAny(m["likes"]), DateUpdated: updated,
			Categories: []string{"plugin", "spigot"}, PageURL: page,
			Installable: boolFromAny(m["external"]) == false, InstallMode: "spiget-resource", Reason: "Spigot resource via Spiget index",
		})
	}
	return out, nil
}

func (a *App) searchBuiltByBit(ctx context.Context, query string, opts providerSearchOptions) ([]UnifiedProject, error) {
	a.mu.RLock()
	key := strings.TrimSpace(a.settings.BuiltByBitAPIKey)
	oauthReady := strings.TrimSpace(a.settings.BuiltByBitOAuthToken) != ""
	a.mu.RUnlock()
	if key == "" {
		return nil, errors.New("BuiltByBit API token is not configured in Settings")
	}
	base := providerBase("MMV_BUILTBYBIT_API_BASE", "https://api.builtbybit.com")
	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 40
	}
	values := url.Values{}
	values.Set("page", strconv.Itoa(maxInt(opts.Offset/limit, 0)+1))
	values.Set("per_page", strconv.Itoa(limit))
	values.Set("with", "Creator,Description,Category,LatestVersion")
	if strings.TrimSpace(query) != "" && !strings.EqualFold(query, "minecraft") {
		values.Set("filters[__keywords]", query)
	}
	var resp struct {
		Data struct {
			Resources []map[string]any `json:"resources"`
		} `json:"data"`
	}
	if err := a.getJSON(ctx, base+"/v2/resources/discover/resources?"+values.Encode(), map[string]string{"Authorization": key}, &resp); err != nil {
		return nil, err
	}
	out := make([]UnifiedProject, 0, len(resp.Data.Resources))
	for _, m := range resp.Data.Resources {
		id := int64FromAny(m["resource_id"])
		title := stringFromAny(m["title"])
		if id == 0 || title == "" {
			continue
		}
		creator := nestedString(m, "Creator", "username")
		avatar := nestedString(m, "Creator", "avatar_url")
		category := nestedString(m, "Category", "title")
		page := stringFromAny(m["url"])
		if page == "" {
			page = "https://builtbybit.com/resources/" + strconv.FormatInt(id, 10) + "/"
		} else {
			page = absoluteURL("https://builtbybit.com", page)
		}
		gallery := stringSliceFromAny(m["carousel_image_urls"])
		if cover := stringFromAny(m["cover_image_url"]); cover != "" {
			gallery = uniqueStringsPreserve(append([]string{cover}, gallery...))
		}
		out = append(out, UnifiedProject{
			ID: strconv.FormatInt(id, 10), Provider: "builtbybit", ProjectType: inferProjectType(title+" "+stringFromAny(m["summary"]), "plugin"), Slug: strconv.FormatInt(id, 10),
			Title: title, Summary: stringFromAny(m["summary"]), Author: creator, AuthorAvatarURL: avatar,
			IconURL: firstString(gallery), Gallery: gallery, Downloads: int64FromAny(m["downloads"]), Followers: int64FromAny(m["purchases"]),
			DateUpdated: timeFromAny(m["last_updated_at"]), Categories: uniqueStringsPreserve([]string{category, "marketplace"}), PageURL: page,
			Installable: oauthReady, InstallMode: "builtbybit-download-api", Reason: "BuiltByBit official Discovery API",
		})
	}
	return out, nil
}

func (a *App) searchATLauncher(ctx context.Context, query string, opts providerSearchOptions) ([]UnifiedProject, error) {
	endpoint := providerBase("MMV_ATLAUNCHER_GRAPHQL", "https://api.atlauncher.com/v2/graphql")
	limit := opts.Limit
	if limit <= 0 || limit > 40 {
		limit = 25
	}
	q := `query SearchForPackByName($first: Int!, $query: String!) { searchPacks(first: $first, query: $query, field: NAME) { id position name safeName latestVersion { id version minecraftVersion changelog isRecommended canUpdate createdAt updatedAt publishedAt } } }`
	var resp struct {
		Data struct {
			Packs []struct {
				ID       int64  `json:"id"`
				Position int64  `json:"position"`
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
			} `json:"searchPacks"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := a.postJSON(ctx, endpoint, nil, map[string]any{"query": q, "variables": map[string]any{"first": limit, "query": query}}, &resp); err != nil {
		return nil, err
	}
	if len(resp.Errors) > 0 {
		return nil, errors.New(resp.Errors[0].Message)
	}
	out := make([]UnifiedProject, 0, len(resp.Data.Packs))
	for _, p := range resp.Data.Packs {
		versions := []string{}
		if p.Latest.MinecraftVersion != "" {
			versions = append(versions, p.Latest.MinecraftVersion)
		}
		out = append(out, UnifiedProject{
			ID: strconv.FormatInt(p.ID, 10), Provider: "atlauncher", ProjectType: "modpack", Slug: p.SafeName,
			Title: p.Name, Summary: cleanHTMLText(p.Latest.Changelog), DateUpdated: firstNonEmpty(p.Latest.UpdatedAt, p.Latest.PublishedAt), Versions: versions,
			Categories: []string{"modpack", "launcher"}, PageURL: "https://atlauncher.com/pack/" + url.PathEscape(p.SafeName),
			Installable: false, InstallMode: "integrated-modpack-manifest", Reason: "ATLauncher public GraphQL catalog",
		})
	}
	return out, nil
}

func (a *App) searchFTB(ctx context.Context, query string, opts providerSearchOptions) ([]UnifiedProject, error) {
	base := providerBase("MMV_FTB_API_BASE", "https://api.modpacks.ch/public")
	limit := opts.Limit
	if limit <= 0 || limit > 20 {
		limit = 12
	}
	var searchResp struct {
		Packs []int64 `json:"packs"`
	}
	if err := a.getJSON(ctx, fmt.Sprintf("%s/modpack/search/%d?term=%s", base, limit, url.QueryEscape(query)), nil, &searchResp); err != nil {
		return nil, err
	}
	ids := searchResp.Packs
	if len(ids) > limit {
		ids = ids[:limit]
	}
	type result struct {
		p   UnifiedProject
		err error
	}
	ch := make(chan result, len(ids))
	var wg sync.WaitGroup
	for _, id := range ids {
		id := id
		wg.Add(1)
		go func() {
			defer wg.Done()
			var m map[string]any
			if err := a.getJSON(ctx, fmt.Sprintf("%s/modpack/%d", base, id), nil, &m); err != nil {
				ch <- result{err: err}
				return
			}
			name := firstNonEmpty(stringFromAny(m["name"]), stringFromAny(m["title"]))
			if name == "" {
				name = "FTB Pack " + strconv.FormatInt(id, 10)
			}
			art := firstNonEmpty(stringFromAny(m["art"]), stringFromAny(m["icon"]), stringFromAny(m["thumbnail"]))
			mc := firstNonEmpty(stringFromAny(m["mcVersion"]), stringFromAny(m["minecraftVersion"]))
			version := ""
			updated := ""
			if vs, ok := m["versions"].([]any); ok && len(vs) > 0 {
				if vm, ok := vs[0].(map[string]any); ok {
					version = stringFromAny(vm["name"])
					mc = firstNonEmpty(mc, stringFromAny(vm["mcVersion"]), stringFromAny(vm["minecraftVersion"]))
					updated = timeFromAny(vm["updated"])
				}
			}
			ch <- result{p: UnifiedProject{ID: strconv.FormatInt(id, 10), Provider: "ftb", ProjectType: "modpack", Slug: strconv.FormatInt(id, 10), Title: name, Summary: firstNonEmpty(stringFromAny(m["synopsis"]), stringFromAny(m["description"])), IconURL: art, Gallery: nonEmptyStrings(art), DateUpdated: updated, Versions: nonEmptyStrings(mc, version), Categories: []string{"modpack", "ftb"}, PageURL: "https://www.feed-the-beast.com/modpacks/" + strconv.FormatInt(id, 10), Installable: false, InstallMode: "integrated-modpack-manifest", Reason: "FTB public modpack catalog"}}
		}()
	}
	go func() { wg.Wait(); close(ch) }()
	out := []UnifiedProject{}
	var firstErr error
	for r := range ch {
		if r.err != nil {
			if firstErr == nil {
				firstErr = r.err
			}
			continue
		}
		out = append(out, r.p)
	}
	if len(out) == 0 && firstErr != nil {
		return nil, firstErr
	}
	return out, nil
}

func (a *App) searchTechnic(ctx context.Context, query string, opts providerSearchOptions) ([]UnifiedProject, error) {
	base := providerBase("MMV_TECHNIC_API_BASE", "https://api.technicpack.net")
	var raw any
	target := base + "/search/modpacks/" + url.PathEscape(query) + "?build=822"
	if err := a.getJSON(ctx, target, nil, &raw); err != nil {
		return nil, err
	}
	items := mapsFromAny(raw)
	out := []UnifiedProject{}
	for _, m := range items {
		slug := firstNonEmpty(stringFromAny(m["slug"]), stringFromAny(m["name"]), stringFromAny(m["id"]))
		name := firstNonEmpty(stringFromAny(m["displayName"]), stringFromAny(m["name"]), titleFromSlug(slug))
		if slug == "" || name == "" {
			continue
		}
		iconURL := firstNonEmpty(stringFromAny(m["icon"]), stringFromAny(m["logo"]), stringFromAny(m["background"]), stringFromAny(m["iconUrl"]))
		out = append(out, UnifiedProject{ID: slug, Provider: "technic", ProjectType: "modpack", Slug: slug, Title: name, Summary: firstNonEmpty(stringFromAny(m["description"]), stringFromAny(m["tag"])), Author: stringFromAny(m["user"]), IconURL: iconURL, Gallery: nonEmptyStrings(iconURL, stringFromAny(m["background"])), Followers: int64FromAny(m["rating"]), Categories: []string{"modpack", "technic"}, PageURL: "https://www.technicpack.net/modpack/" + url.PathEscape(slug), Installable: false, InstallMode: "integrated-modpack-manifest", Reason: "Technic Platform API"})
		if len(out) >= opts.Limit && opts.Limit > 0 {
			break
		}
	}
	return out, nil
}

func (a *App) searchBukkitDev(ctx context.Context, query string, opts providerSearchOptions) ([]UnifiedProject, error) {
	base := providerBase("MMV_BUKKITDEV_BASE", "https://dev.bukkit.org")
	u := base + "/bukkit-plugins?filter-search=" + url.QueryEscape(query)
	body, err := a.getText(ctx, u, nil)
	if err != nil {
		return nil, err
	}
	re := regexp.MustCompile(`(?is)<a[^>]+href=["'](/projects/[a-z0-9][^"'?#]*)["'][^>]*>(.*?)</a>`)
	items := parseGenericCards(body, "bukkitdev", base, re, "plugin", minPositive(opts.Limit, 30))
	for i := range items {
		items[i].Categories = uniqueStringsPreserve(append(items[i].Categories, "plugin", "bukkit"))
		items[i].InstallMode = "detected-downloads"
	}
	return items, nil
}

func (a *App) searchModDB(ctx context.Context, query string, opts providerSearchOptions) ([]UnifiedProject, error) {
	base := providerBase("MMV_MODDB_BASE", "https://www.moddb.com")
	u := base + "/games/minecraft/mods?filter=t&kw=" + url.QueryEscape(query)
	body, err := a.getText(ctx, u, nil)
	if err != nil {
		return nil, err
	}
	re := regexp.MustCompile(`(?is)<a[^>]+href=["'](/mods/[a-z0-9][^"'?#]*)["'][^>]*>(.*?)</a>`)
	items := parseGenericCards(body, "moddb", base, re, "mod", minPositive(opts.Limit, 30))
	for i := range items {
		items[i].Categories = uniqueStringsPreserve(append(items[i].Categories, "moddb"))
		items[i].InstallMode = "detected-downloads"
	}
	return items, nil
}

func (a *App) searchNexusMods(ctx context.Context, query string, opts providerSearchOptions) ([]UnifiedProject, error) {
	a.mu.RLock()
	key := strings.TrimSpace(a.settings.NexusAPIKey)
	a.mu.RUnlock()
	if key == "" {
		return nil, errors.New("Nexus Mods API key is not configured in Settings")
	}
	base := providerBase("MMV_NEXUS_API_BASE", "https://api.nexusmods.com/v1")
	headers := map[string]string{"apikey": key, "Application-Name": appName, "Application-Version": appVersion}
	endpoints := []string{"latest_updated.json", "latest_added.json", "trending.json"}
	all := []map[string]any{}
	seen := map[int64]bool{}
	for _, ep := range endpoints {
		var rows []map[string]any
		if err := a.getJSON(ctx, base+"/games/minecraft/mods/"+ep, headers, &rows); err != nil {
			continue
		}
		for _, m := range rows {
			id := int64FromAny(m["mod_id"])
			if id > 0 && !seen[id] {
				seen[id] = true
				all = append(all, m)
			}
		}
	}
	if numeric, err := strconv.ParseInt(strings.TrimSpace(query), 10, 64); err == nil && numeric > 0 && !seen[numeric] {
		var m map[string]any
		if err := a.getJSON(ctx, fmt.Sprintf("%s/games/minecraft/mods/%d.json", base, numeric), headers, &m); err == nil {
			all = append(all, m)
		}
	}
	q := strings.ToLower(strings.TrimSpace(query))
	out := []UnifiedProject{}
	for _, m := range all {
		name := stringFromAny(m["name"])
		summary := firstNonEmpty(stringFromAny(m["summary"]), stringFromAny(m["description"]))
		if q != "" && q != "minecraft" && !strings.Contains(strings.ToLower(name+" "+summary), q) {
			continue
		}
		id := int64FromAny(m["mod_id"])
		if id == 0 {
			continue
		}
		img := firstNonEmpty(stringFromAny(m["picture_url"]), stringFromAny(m["mod_picture_url"]))
		out = append(out, UnifiedProject{ID: strconv.FormatInt(id, 10), Provider: "nexusmods", ProjectType: "mod", Slug: strconv.FormatInt(id, 10), Title: name, Summary: summary, Author: stringFromAny(m["author"]), IconURL: img, Gallery: nonEmptyStrings(img), Downloads: int64FromAny(m["endorsement_count"]), DateUpdated: timeFromAny(firstNonNil(m["updated_timestamp"], m["updated_time"])), Categories: []string{"mod", "nexusmods"}, PageURL: fmt.Sprintf("https://www.nexusmods.com/minecraft/mods/%d", id), Installable: false, InstallMode: "integrated-metadata", Reason: "Nexus Mods API metadata"})
		if len(out) >= minPositive(opts.Limit, 30) {
			break
		}
	}
	return out, nil
}

func minPositive(v, fallback int) int {
	if v <= 0 || v > fallback {
		return fallback
	}
	return v
}

func firstString(in []string) string {
	for _, s := range in {
		if strings.TrimSpace(s) != "" {
			return s
		}
	}
	return ""
}

func nonEmptyStrings(xs ...string) []string {
	out := []string{}
	for _, x := range xs {
		if strings.TrimSpace(x) != "" {
			out = append(out, x)
		}
	}
	return uniqueStringsPreserve(out)
}

func absoluteURL(base, raw string) string {
	raw = html.UnescapeString(strings.TrimSpace(raw))
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "//") {
		return "https:" + raw
	}
	if u, err := url.Parse(raw); err == nil && u.IsAbs() {
		return raw
	}
	b, err := url.Parse(base)
	if err != nil {
		return raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	return b.ResolveReference(u).String()
}

func stringFromAny(v any) string {
	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x)
	case json.Number:
		return x.String()
	case float64:
		if x == float64(int64(x)) {
			return strconv.FormatInt(int64(x), 10)
		}
		return strconv.FormatFloat(x, 'f', -1, 64)
	case int:
		return strconv.Itoa(x)
	case int64:
		return strconv.FormatInt(x, 10)
	}
	return ""
}

func int64FromAny(v any) int64 {
	switch x := v.(type) {
	case float64:
		return int64(x)
	case int:
		return int64(x)
	case int64:
		return x
	case json.Number:
		n, _ := x.Int64()
		return n
	case string:
		n, _ := strconv.ParseInt(strings.TrimSpace(x), 10, 64)
		return n
	}
	return 0
}

func boolFromAny(v any) bool {
	switch x := v.(type) {
	case bool:
		return x
	case string:
		b, _ := strconv.ParseBool(x)
		return b
	case float64:
		return x != 0
	}
	return false
}

func nestedString(m map[string]any, keys ...string) string {
	var cur any = m
	for _, key := range keys {
		obj, ok := cur.(map[string]any)
		if !ok {
			return ""
		}
		cur = obj[key]
	}
	return stringFromAny(cur)
}

func stringSliceFromAny(v any) []string {
	out := []string{}
	switch xs := v.(type) {
	case []any:
		for _, x := range xs {
			if s := stringFromAny(x); s != "" {
				out = append(out, s)
			}
		}
	case []string:
		out = append(out, xs...)
	}
	return out
}

func timeFromAny(v any) string {
	switch x := v.(type) {
	case string:
		x = strings.TrimSpace(x)
		if x == "" {
			return ""
		}
		if _, err := time.Parse(time.RFC3339, x); err == nil {
			return x
		}
		if n, err := strconv.ParseInt(x, 10, 64); err == nil {
			return unixFlexible(n)
		}
		return x
	case float64:
		return unixFlexible(int64(x))
	case int64:
		return unixFlexible(x)
	case int:
		return unixFlexible(int64(x))
	}
	return ""
}

func unixFlexible(n int64) string {
	if n <= 0 {
		return ""
	}
	if n > 1_000_000_000_000 {
		n /= 1000
	}
	return time.Unix(n, 0).UTC().Format(time.RFC3339)
}

func mapsFromAny(v any) []map[string]any {
	out := []map[string]any{}
	var visit func(any)
	visit = func(x any) {
		switch t := x.(type) {
		case []any:
			for _, item := range t {
				visit(item)
			}
		case map[string]any:
			if stringFromAny(t["slug"]) != "" || stringFromAny(t["name"]) != "" || stringFromAny(t["displayName"]) != "" {
				out = append(out, t)
				return
			}
			for _, key := range []string{"results", "modpacks", "packs", "data"} {
				if child, ok := t[key]; ok {
					visit(child)
				}
			}
		}
	}
	visit(v)
	return out
}

func inferProjectType(s, fallback string) string {
	s = strings.ToLower(s)
	for _, pair := range [][2]string{{"modpack", "modpack"}, {"plugin", "plugin"}, {"resource pack", "resourcepack"}, {"texture pack", "resourcepack"}, {"shader", "shader"}, {"data pack", "datapack"}, {"datapack", "datapack"}, {"addon", "addon"}, {"add-on", "addon"}, {"map", "world"}} {
		if strings.Contains(s, pair[0]) {
			return pair[1]
		}
	}
	return fallback
}

func firstNonNil(xs ...any) any {
	for _, x := range xs {
		if x != nil {
			return x
		}
	}
	return nil
}

// orderProjectsByQuery provides deterministic local relevance for providers whose
// public APIs do not expose a native relevance sort.
func orderProjectsByQuery(items []UnifiedProject, query string) {
	q := strings.ToLower(strings.TrimSpace(query))
	sort.SliceStable(items, func(i, j int) bool {
		iTitle := strings.ToLower(items[i].Title)
		jTitle := strings.ToLower(items[j].Title)
		iExact, jExact := iTitle == q, jTitle == q
		if iExact != jExact {
			return iExact
		}
		iHas, jHas := strings.Contains(iTitle, q), strings.Contains(jTitle, q)
		if iHas != jHas {
			return iHas
		}
		return items[i].Downloads > items[j].Downloads
	})
}
