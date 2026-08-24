package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Smithed exposes a first-class public v2 API for datapacks and resource packs.
// Keep the integration native so search, compatibility, metadata, versions,
// dependencies and installation stay in the Vault instead of degrading to links.
func (a *App) searchSmithed(ctx context.Context, query string, opts providerSearchOptions) ([]UnifiedProject, error) {
	base := providerBase("MMV_SMITHED_API_BASE", "https://api.smithed.dev/v2")
	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 40
	}
	values := url.Values{}
	if strings.TrimSpace(query) != "" && !strings.EqualFold(query, "minecraft") {
		values.Set("search", query)
	}
	values.Set("limit", strconv.Itoa(limit))
	values.Set("start", strconv.Itoa(maxInt(opts.Offset, 0)))
	sortMode := "trending"
	switch strings.ToLower(opts.Sort) {
	case "downloads":
		sortMode = "downloads"
	case "updated":
		sortMode = "newest"
	case "name":
		sortMode = "alphabetically"
	}
	values.Set("sort", sortMode)
	if opts.GameVersion != "" {
		values.Add("version", opts.GameVersion)
	}
	for _, scope := range []string{
		"data.display.name", "data.display.description", "data.display.icon", "data.display.webPage",
		"data.display.urls", "data.display.gallery", "data.versions", "data.categories",
		"meta.owner", "meta.contributors", "meta.stats",
	} {
		values.Add("scope", scope)
	}
	var raw []map[string]any
	if err := a.getJSON(ctx, base+"/packs?"+values.Encode(), nil, &raw); err != nil {
		return nil, err
	}
	out := make([]UnifiedProject, 0, len(raw))
	for _, item := range raw {
		id := stringFromAny(item["id"])
		data, _ := item["data"].(map[string]any)
		meta, _ := item["meta"].(map[string]any)
		display, _ := data["display"].(map[string]any)
		name := firstNonEmpty(stringFromAny(display["name"]), stringFromAny(item["displayName"]), id)
		if id == "" || name == "" {
			continue
		}
		versions := smithedVersions(data["versions"])
		gameVersions := []string{}
		hasData, hasResource := false, false
		for _, v := range versions {
			gameVersions = append(gameVersions, v.Supports...)
			hasData = hasData || v.Datapack != ""
			hasResource = hasResource || v.Resourcepack != ""
		}
		ptype := "datapack"
		if hasResource && !hasData {
			ptype = "resourcepack"
		}
		if opts.ProjectType != "" && opts.ProjectType != "all" && !projectTypeMatches(ptype, opts.ProjectType) {
			// Packs that ship both datapack + resourcepack should appear in either lane.
			if !(hasData && hasResource && (opts.ProjectType == "datapack" || opts.ProjectType == "resourcepack")) {
				continue
			}
			ptype = opts.ProjectType
		}
		stats, _ := meta["stats"].(map[string]any)
		downloads, _ := stats["downloads"].(map[string]any)
		owner := stringFromAny(meta["owner"])
		page := firstNonEmpty(stringFromAny(display["webPage"]), "https://smithed.dev/packs/"+url.PathEscape(id))
		gallery := smithedGallery(display["gallery"])
		iconURL := stringFromAny(display["icon"])
		if iconURL != "" {
			gallery = uniqueStringsPreserve(append([]string{iconURL}, gallery...))
		}
		categories := stringSliceFromAny(data["categories"])
		project := UnifiedProject{
			ID: id, Provider: "smithed", ProjectType: ptype, Slug: id, Title: name,
			Summary: stringFromAny(display["description"]), Author: owner,
			AuthorAvatarURL: smithedUserAvatarURL(base, owner), IconURL: iconURL, Gallery: gallery,
			Downloads: int64FromAny(downloads["total"]), Followers: int64FromAny(stats["score"]),
			DateUpdated: timeFromAny(stats["updated"]), Categories: uniqueStringsPreserve(append(categories, "smithed")),
			Versions: uniqueStringsPreserve(gameVersions), PageURL: page, Installable: hasData || hasResource,
			InstallMode: "smithed-pack", Reason: "Smithed native v2 pack API",
		}
		out = append(out, project)
	}
	orderProjectsByQuery(out, query)
	return out, nil
}

type smithedVersion struct {
	Name         string
	Supports     []string
	Datapack     string
	Resourcepack string
	Dependencies []string
}

func smithedVersions(v any) []smithedVersion {
	out := []smithedVersion{}
	for _, item := range anySlice(v) {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		downloads, _ := m["downloads"].(map[string]any)
		deps := []string{}
		for _, dep := range anySlice(m["dependencies"]) {
			if dm, ok := dep.(map[string]any); ok {
				id := stringFromAny(dm["id"])
				ver := stringFromAny(dm["version"])
				if id != "" {
					if ver != "" {
						id += "@" + ver
					}
					deps = append(deps, id)
				}
			}
		}
		out = append(out, smithedVersion{
			Name: stringFromAny(m["name"]), Supports: stringSliceFromAny(m["supports"]),
			Datapack: stringFromAny(downloads["datapack"]), Resourcepack: stringFromAny(downloads["resourcepack"]),
			Dependencies: deps,
		})
	}
	return out
}

func smithedGallery(v any) []string {
	if direct := stringSliceFromAny(v); len(direct) > 0 {
		return direct
	}
	if m, ok := v.(map[string]any); ok {
		if content := stringFromAny(m["content"]); content != "" && strings.HasPrefix(content, "http") {
			return []string{content}
		}
	}
	return nil
}

func smithedUserAvatarURL(base, user string) string {
	if user == "" {
		return ""
	}
	return strings.TrimRight(base, "/") + "/users/" + url.PathEscape(user) + "/pfp"
}

func (a *App) smithedDetails(ctx context.Context, id string) (ProjectDetails, error) {
	base := providerBase("MMV_SMITHED_API_BASE", "https://api.smithed.dev/v2")
	var data map[string]any
	if err := a.getJSON(ctx, base+"/packs/"+url.PathEscape(id), nil, &data); err != nil {
		return ProjectDetails{}, err
	}
	var meta map[string]any
	_ = a.getJSON(ctx, base+"/packs/"+url.PathEscape(id)+"/meta", nil, &meta)
	display, _ := data["display"].(map[string]any)
	versions := smithedVersions(data["versions"])
	gameVersions := []string{}
	hasData, hasResource := false, false
	for _, v := range versions {
		gameVersions = append(gameVersions, v.Supports...)
		hasData = hasData || v.Datapack != ""
		hasResource = hasResource || v.Resourcepack != ""
	}
	ptype := "datapack"
	if hasResource && !hasData {
		ptype = "resourcepack"
	}
	owner := stringFromAny(meta["owner"])
	if owner == "" {
		owner = stringFromAny(data["owner"])
	}
	page := firstNonEmpty(stringFromAny(display["webPage"]), "https://smithed.dev/packs/"+url.PathEscape(id))
	d := ProjectDetails{
		ID: firstNonEmpty(stringFromAny(data["id"]), id), Slug: id, ProjectType: ptype,
		Title: firstNonEmpty(stringFromAny(display["name"]), id), Summary: stringFromAny(display["description"]),
		Description: stringFromAny(display["description"]), IconURL: stringFromAny(display["icon"]),
		Gallery: smithedGallery(display["gallery"]), Categories: stringSliceFromAny(data["categories"]),
		GameVersions: uniqueStringsPreserve(gameVersions), PageURL: page, Installable: hasData || hasResource,
		InstallMode: "smithed-pack", Links: map[string]string{},
	}
	if d.IconURL != "" {
		d.Gallery = uniqueStringsPreserve(append([]string{d.IconURL}, d.Gallery...))
	}
	if owner != "" {
		d.Authors = append(d.Authors, ProjectAuthor{Name: owner, AvatarURL: smithedUserAvatarURL(base, owner)})
	}
	for _, contributor := range stringSliceFromAny(meta["contributors"]) {
		if !strings.EqualFold(contributor, owner) {
			d.Authors = append(d.Authors, ProjectAuthor{Name: contributor, Role: "Contributor", AvatarURL: smithedUserAvatarURL(base, contributor)})
		}
	}
	if urls, ok := display["urls"].(map[string]any); ok {
		for label, key := range map[string]string{"Homepage": "homepage", "Source": "source", "Discord": "discord"} {
			if v := stringFromAny(urls[key]); v != "" {
				d.Links[label] = v
			}
		}
	}
	if stats, ok := meta["stats"].(map[string]any); ok {
		if downloads, ok := stats["downloads"].(map[string]any); ok {
			d.Downloads = int64FromAny(downloads["total"])
		}
		d.Updated = timeFromAny(stats["updated"])
		d.Published = timeFromAny(stats["added"])
	}
	for _, v := range versions {
		pv := ProjectVersionSummary{ID: v.Name, Name: v.Name, Version: v.Name, GameVersions: v.Supports, Dependencies: v.Dependencies, Installable: v.Datapack != "" || v.Resourcepack != ""}
		if v.Datapack != "" {
			pv.Files = append(pv.Files, ProjectFileSummary{Name: d.Slug + "-" + v.Name + "-datapack.zip", URL: v.Datapack, Primary: true})
		}
		if v.Resourcepack != "" {
			pv.Files = append(pv.Files, ProjectFileSummary{Name: d.Slug + "-" + v.Name + "-resourcepack.zip", URL: v.Resourcepack, Primary: v.Datapack == ""})
		}
		d.Versions = append(d.Versions, pv)
	}
	return d, nil
}

func (a *App) installSmithed(ctx context.Context, id, game, target string) (map[string]any, error) {
	return a.installSmithedRecursive(ctx, id, game, target, map[string]bool{}, false)
}

func (a *App) installSmithedRecursive(ctx context.Context, id, game, target string, seen map[string]bool, dependency bool) (map[string]any, error) {
	base := providerBase("MMV_SMITHED_API_BASE", "https://api.smithed.dev/v2")
	cleanID := id
	if at := strings.Index(cleanID, "@"); at > 0 {
		cleanID = cleanID[:at]
	}
	if seen[cleanID] {
		return map[string]any{"ok": true, "provider": "smithed", "project": cleanID, "skipped": "already installed in dependency transaction"}, nil
	}
	seen[cleanID] = true
	values := url.Values{}
	if game != "" {
		values.Set("version", game)
	}
	endpoint := base + "/packs/" + url.PathEscape(cleanID) + "/versions/latest"
	if len(values) > 0 {
		endpoint += "?" + values.Encode()
	}
	var m map[string]any
	if err := a.getJSON(ctx, endpoint, nil, &m); err != nil {
		return nil, err
	}
	versions := smithedVersions([]any{m})
	if len(versions) == 0 {
		return nil, errors.New("Smithed did not return a compatible pack version")
	}
	v := versions[0]
	for _, dep := range v.Dependencies {
		if _, err := a.installSmithedRecursive(ctx, dep, game, "auto", seen, true); err != nil {
			return nil, fmt.Errorf("Smithed dependency %s: %w", dep, err)
		}
	}
	selectedURL := ""
	selectedTarget := target
	kind := "datapack"
	if selectedTarget == "" || selectedTarget == "auto" {
		if v.Datapack != "" {
			selectedTarget = "datapacks"
			selectedURL = v.Datapack
			kind = "datapack"
		} else {
			selectedTarget = "resourcepacks"
			selectedURL = v.Resourcepack
			kind = "resourcepack"
		}
	} else if selectedTarget == "datapacks" || selectedTarget == "downloads" {
		if v.Datapack != "" {
			selectedURL = v.Datapack
			kind = "datapack"
		} else if selectedTarget == "downloads" && v.Resourcepack != "" {
			selectedURL = v.Resourcepack
			kind = "resourcepack"
		} else {
			return nil, errors.New("this Smithed pack has no datapack payload")
		}
	} else if selectedTarget == "resourcepacks" {
		if v.Resourcepack == "" {
			return nil, errors.New("this Smithed pack has no resource-pack payload")
		}
		selectedURL = v.Resourcepack
		kind = "resourcepack"
	} else {
		return nil, errors.New("Smithed packs install to datapacks, resourcepacks, or Vault downloads")
	}
	if selectedURL == "" {
		return nil, errors.New("Smithed version has no downloadable payload")
	}
	if selectedTarget == "datapacks" {
		a.mu.RLock()
		worldRoot := strings.TrimSpace(a.settings.WorldRoot)
		a.mu.RUnlock()
		if worldRoot == "" {
			return nil, errors.New("set Active world directory in Settings before installing datapacks directly into a world")
		}
	}
	dir := a.javaTargetDir(selectedTarget)
	if err := osMkdirAll(dir); err != nil {
		return nil, err
	}
	version := firstNonEmpty(v.Name, "latest")
	filename := safeFilename(cleanID + "-" + version + "-" + kind + downloadExtension(selectedURL))
	if filepath.Ext(filename) == "" {
		filename += ".zip"
	}
	dst := uniquePath(filepath.Join(dir, filename))
	if err := a.downloadURLVerified(ctx, selectedURL, dst, 0, nil); err != nil {
		return nil, err
	}
	if err := validateZipContainer(dst); err != nil {
		_ = os.Remove(dst)
		return nil, fmt.Errorf("downloaded Smithed pack is not a valid ZIP: %w", err)
	}
	return map[string]any{"ok": true, "provider": "smithed", "project": cleanID, "version": version, "file": filepath.Base(dst), "path": dst, "target": selectedTarget, "kind": kind, "dependency": dependency}, nil
}

func (a *App) searchPolymart(ctx context.Context, query string, opts providerSearchOptions) ([]UnifiedProject, error) {
	base := providerBase("MMV_POLYMART_BASE", "https://polymart.org")
	values := url.Values{}
	if strings.TrimSpace(query) != "" && !strings.EqualFold(query, "minecraft") {
		values.Set("search", query)
	}
	values.Set("sort", "relevant")
	body, err := a.getText(ctx, base+"/resources?"+values.Encode(), nil)
	if err != nil {
		body, err = a.getText(ctx, base+"/products?"+values.Encode(), nil)
		if err != nil {
			return nil, err
		}
	}
	// Modern pages use /product/<id>/<slug>, while legacy aliases may use /resource/<slug>.<id>.
	re := regexp.MustCompile(`(?is)<a[^>]+href=["']((?:https://polymart\.org)?/(?:product/\d+(?:/[^"'?#]*)?|resource/[^"'?#]+))["'][^>]*>(.*?)</a>`)
	items := parseGenericCards(body, "polymart", base, re, "plugin", maxInt(opts.Limit, 30))
	for i := range items {
		items[i].ProjectType = inferProjectType(items[i].Title+" "+items[i].Summary, "plugin")
		items[i].Installable = false
		items[i].InstallMode = "integrated-marketplace"
		items[i].Reason = "Live Polymart product index rendered inside Minecraft Mod Vault"
		if id := polymartID(items[i].PageURL); id != "" {
			items[i].ID = id
			items[i].Slug = id
		}
	}
	orderProjectsByQuery(items, query)
	if opts.Limit > 0 && len(items) > opts.Limit {
		items = items[:opts.Limit]
	}
	return items, nil
}

func polymartID(s string) string {
	if m := regexp.MustCompile(`/product/(\d+)`).FindStringSubmatch(s); len(m) > 1 {
		return m[1]
	}
	if m := regexp.MustCompile(`/resource/[^/?#]*\.(\d+)`).FindStringSubmatch(s); len(m) > 1 {
		return m[1]
	}
	return ""
}

func (a *App) searchSpongeOre(ctx context.Context, query string, opts providerSearchOptions) ([]UnifiedProject, error) {
	base := providerBase("MMV_SPONGE_ORE_BASE", "https://ore.spongepowered.org/api/v1")
	limit := opts.Limit
	if limit <= 0 || limit > 25 {
		limit = 25
	}
	values := url.Values{"limit": {strconv.Itoa(limit)}, "offset": {strconv.Itoa(maxInt(opts.Offset, 0))}, "sort": {"1"}}
	if strings.TrimSpace(query) != "" && !strings.EqualFold(query, "minecraft") {
		values.Set("q", query)
	}
	var raw []map[string]any
	if err := a.getJSON(ctx, base+"/projects?"+values.Encode(), nil, &raw); err != nil {
		return nil, err
	}
	out := make([]UnifiedProject, 0, len(raw))
	for _, m := range raw {
		id := stringFromAny(m["pluginId"])
		name := stringFromAny(m["name"])
		if id == "" || name == "" {
			continue
		}
		owner := stringFromAny(m["owner"])
		category := nestedString(m, "category", "title")
		page := absoluteURL("https://ore.spongepowered.org", stringFromAny(m["href"]))
		if page == "" {
			page = "https://ore.spongepowered.org/" + url.PathEscape(owner) + "/" + url.PathEscape(name)
		}
		updated := timeFromAny(m["updatedAt"])
		if updated == "" {
			updated = timeFromAny(m["createdAt"])
		}
		out = append(out, UnifiedProject{ID: id, Provider: "spongeore", ProjectType: "plugin", Slug: id, Title: name,
			Summary: stringFromAny(m["description"]), Author: owner, Downloads: int64FromAny(m["downloads"]), Followers: int64FromAny(m["stars"]),
			DateUpdated: updated, Categories: uniqueStringsPreserve([]string{category, "sponge", "plugin"}), PageURL: page,
			Installable: m["recommended"] != nil, InstallMode: "sponge-ore-version", Reason: "Sponge Ore native project API"})
	}
	orderProjectsByQuery(out, query)
	return out, nil
}

func (a *App) spongeOreDetails(ctx context.Context, id string) (ProjectDetails, error) {
	base := providerBase("MMV_SPONGE_ORE_BASE", "https://ore.spongepowered.org/api/v1")
	var m map[string]any
	if err := a.getJSON(ctx, base+"/projects/"+url.PathEscape(id), nil, &m); err != nil {
		return ProjectDetails{}, err
	}
	name := firstNonEmpty(stringFromAny(m["name"]), id)
	owner := stringFromAny(m["owner"])
	page := absoluteURL("https://ore.spongepowered.org", stringFromAny(m["href"]))
	d := ProjectDetails{ID: id, Slug: id, ProjectType: "plugin", Title: name, Summary: stringFromAny(m["description"]), Description: stringFromAny(m["description"]),
		Categories: uniqueStringsPreserve([]string{nestedString(m, "category", "title"), "sponge", "plugin"}), Downloads: int64FromAny(m["downloads"]), Followers: int64FromAny(m["stars"]),
		Published: timeFromAny(m["createdAt"]), Updated: timeFromAny(m["updatedAt"]), PageURL: page, Installable: true, InstallMode: "sponge-ore-version", Links: map[string]string{}}
	if owner != "" {
		d.Authors = append(d.Authors, ProjectAuthor{Name: owner})
	}
	var versions []map[string]any
	if err := a.getJSON(ctx, base+"/projects/"+url.PathEscape(id)+"/versions?limit=25&offset=0", nil, &versions); err == nil {
		for _, vm := range versions {
			versionName := stringFromAny(vm["name"])
			deps := []string{}
			for _, dep := range anySlice(vm["dependencies"]) {
				if dm, ok := dep.(map[string]any); ok {
					if depID := stringFromAny(dm["pluginId"]); depID != "" {
						deps = append(deps, depID+"@"+stringFromAny(dm["version"]))
					}
				}
			}
			games := []string{}
			for _, tag := range anySlice(vm["tags"]) {
				if tm, ok := tag.(map[string]any); ok && strings.EqualFold(stringFromAny(tm["name"]), "Sponge") {
					games = append(games, stringFromAny(tm["data"]))
				}
			}
			d.GameVersions = append(d.GameVersions, games...)
			hashes := map[string]string{}
			if md5v := stringFromAny(vm["md5"]); md5v != "" {
				hashes["md5"] = md5v
			}
			download := base + "/projects/" + url.PathEscape(id) + "/versions/" + url.PathEscape(versionName) + "/download"
			d.Versions = append(d.Versions, ProjectVersionSummary{ID: stringFromAny(vm["id"]), Name: versionName, Version: versionName, Published: timeFromAny(vm["createdAt"]), GameVersions: games, Loaders: []string{"sponge"}, Dependencies: deps, Installable: true,
				Files: []ProjectFileSummary{{Name: safeFilename(id + "-" + versionName + ".jar"), URL: download, Size: int64FromAny(vm["fileSize"]), Primary: true, Hashes: hashes}}})
		}
	}
	return d, nil
}

func (a *App) installSpongeOre(ctx context.Context, id, target string) (map[string]any, error) {
	base := providerBase("MMV_SPONGE_ORE_BASE", "https://ore.spongepowered.org/api/v1")
	var versions []map[string]any
	if err := a.getJSON(ctx, base+"/projects/"+url.PathEscape(id)+"/versions?limit=25&offset=0", nil, &versions); err != nil {
		return nil, err
	}
	if len(versions) == 0 {
		return nil, errors.New("Sponge Ore returned no published versions")
	}
	chosen := versions[0]
	name := stringFromAny(chosen["name"])
	if name == "" {
		return nil, errors.New("Sponge Ore version is missing its name")
	}
	if target == "" || target == "auto" {
		target = "plugins"
	}
	if target != "plugins" && target != "downloads" {
		return nil, errors.New("Sponge Ore projects install to server plugins or Vault downloads")
	}
	dir := a.javaTargetDir(target)
	if err := osMkdirAll(dir); err != nil {
		return nil, err
	}
	dst := uniquePath(filepath.Join(dir, safeFilename(id+"-"+name+".jar")))
	hashes := map[string]string{}
	if md5v := stringFromAny(chosen["md5"]); md5v != "" {
		hashes["md5"] = md5v
	}
	download := base + "/projects/" + url.PathEscape(id) + "/versions/" + url.PathEscape(name) + "/download"
	if err := a.downloadURLVerified(ctx, download, dst, int64FromAny(chosen["fileSize"]), hashes); err != nil {
		return nil, err
	}
	if err := validateZipContainer(dst); err != nil {
		_ = os.Remove(dst)
		return nil, fmt.Errorf("downloaded Sponge Ore plugin is not a valid JAR: %w", err)
	}
	return map[string]any{"ok": true, "provider": "spongeore", "project": id, "version": name, "file": filepath.Base(dst), "path": dst, "target": target}, nil
}

func (a *App) searchVanillaTweaks(ctx context.Context, query string, opts providerSearchOptions) ([]UnifiedProject, error) {
	base := providerBase("MMV_VANILLA_TWEAKS_BASE", "https://vanillatweaks.net")
	q := strings.ToLower(strings.TrimSpace(query))
	types := []struct {
		id, name, summary, path, ptype string
	}{
		{"resource-packs", "Vanilla Tweaks Resource Packs", "Official selectable resource-pack tweaks for textures, models, shaders, sounds and visual polish.", "/picker/resource-packs/", "resourcepack"},
		{"datapacks", "Vanilla Tweaks Data Packs", "Official selectable datapacks for quality-of-life mechanics and vanilla-friendly gameplay changes.", "/picker/datapacks/", "datapack"},
		{"crafting-tweaks", "Vanilla Tweaks Crafting Tweaks", "Official recipe-changing crafting tweaks delivered as datapacks.", "/picker/crafting-tweaks/", "datapack"},
	}
	out := []UnifiedProject{}
	for _, t := range types {
		if opts.ProjectType != "" && opts.ProjectType != "all" && !projectTypeMatches(t.ptype, opts.ProjectType) {
			continue
		}
		text := strings.ToLower(t.name + " " + t.summary)
		if q != "" && q != "minecraft" && q != "minecraft mods" {
			matched := true
			for _, token := range strings.Fields(q) {
				if len(token) > 2 && !strings.Contains(text, token) {
					matched = false
					break
				}
			}
			if !matched && !strings.Contains(text, q) {
				continue
			}
		}
		page := strings.TrimRight(base, "/") + t.path
		// Read the current official page so this lane is a live integration rather than
		// a hard-coded hyperlink. Extract current Minecraft versions and preview media.
		body, err := a.getText(ctx, page, nil)
		if err != nil {
			return nil, err
		}
		versions := uniqueStringsPreserve(regexp.MustCompile(`\b(?:26\.[0-9]+|1\.(?:[0-9]{2})(?:\.[0-9]+)?)\b`).FindAllString(cleanHTMLText(body), -1))
		gallery := imagesFromChunk(body, 8)
		out = append(out, UnifiedProject{ID: t.id, Provider: "vanillatweaks", ProjectType: t.ptype, Slug: t.id, Title: t.name, Summary: t.summary,
			Author: "Vanilla Tweaks", IconURL: firstString(gallery), Gallery: gallery, Categories: []string{"vanilla", "tweaks", t.ptype}, Versions: versions,
			PageURL: page, Installable: false, InstallMode: "official-picker", Reason: "Live official Vanilla Tweaks picker rendered as an integrated Vault source"})
	}
	return out, nil
}

func (a *App) vanillaTweaksDetails(ctx context.Context, id, pageURL string) (ProjectDetails, error) {
	if pageURL == "" {
		base := providerBase("MMV_VANILLA_TWEAKS_BASE", "https://vanillatweaks.net")
		switch id {
		case "resource-packs":
			pageURL = base + "/picker/resource-packs/"
		case "crafting-tweaks":
			pageURL = base + "/picker/crafting-tweaks/"
		default:
			pageURL = base + "/picker/datapacks/"
		}
	}
	d, err := a.genericWebDetails(ctx, "vanillatweaks", pageURL, id)
	if err != nil {
		return ProjectDetails{}, err
	}
	d.ID = id
	d.Slug = id
	d.Authors = []ProjectAuthor{{Name: "Vanilla Tweaks"}}
	d.Installable = false
	d.InstallMode = "official-picker"
	d.GameVersions = uniqueStringsPreserve(regexp.MustCompile(`\b(?:26\.[0-9]+|1\.(?:[0-9]{2})(?:\.[0-9]+)?)\b`).FindAllString(d.Description+" "+d.Summary, -1))
	switch id {
	case "resource-packs":
		d.ProjectType = "resourcepack"
	case "crafting-tweaks":
		d.ProjectType = "datapack"
	default:
		d.ProjectType = "datapack"
	}
	return d, nil
}

func (a *App) searchMinecraftMaps(ctx context.Context, query string, opts providerSearchOptions) ([]UnifiedProject, error) {
	base := providerBase("MMV_MINECRAFTMAPS_BASE", "https://www.minecraftmaps.com")
	paths := []string{
		base + "/search?view=search&searchword=" + url.QueryEscape(query),
		base + "/search/" + url.PathEscape(strings.ReplaceAll(strings.TrimSpace(query), " ", "-")),
	}
	var lastErr error
	for _, endpoint := range paths {
		body, err := a.getText(ctx, endpoint, nil)
		if err != nil {
			lastErr = err
			continue
		}
		items := parseGenericCards(body, "minecraftmaps", base, regexp.MustCompile(`(?is)<a[^>]+href=["'](/\d{3,8}-[^"'?#]+)["'][^>]*>(.*?)</a>`), "world", providerResultLimit(opts, 40))
		for i := range items {
			items[i].Categories = uniqueStringsPreserve(append(items[i].Categories, "maps", "worlds"))
			items[i].InstallMode = "verified-detected-download"
			if v := matchMinecraftVersion(items[i].Title + " " + items[i].Summary); v != "" {
				items[i].Versions = []string{v}
			}
		}
		if len(items) > 0 {
			orderProjectsByQuery(items, query)
			return items, nil
		}
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return []UnifiedProject{}, nil
}

func (a *App) searchResourcePackNet(ctx context.Context, query string, opts providerSearchOptions) ([]UnifiedProject, error) {
	base := providerBase("MMV_RESOURCEPACKNET_BASE", "https://resourcepack.net")
	body, err := a.getText(ctx, base+"/?s="+url.QueryEscape(query), nil)
	if err != nil {
		return nil, err
	}
	re := regexp.MustCompile(`(?is)<a[^>]+href=["']((?:https?://[^"']*resourcepack\.net)?/[^"'?#]*(?:resource|texture)[^"'?#]*/)["'][^>]*>(.*?)</a>`)
	items := parseGenericCards(body, "resourcepacknet", base, re, "resourcepack", providerResultLimit(opts, 40))
	for i := range items {
		items[i].Categories = uniqueStringsPreserve(append(items[i].Categories, "resourcepack", "texture-pack"))
		items[i].InstallMode = "verified-detected-download"
		if v := matchMinecraftVersion(items[i].Title + " " + items[i].Summary); v != "" {
			items[i].Versions = []string{v}
		}
	}
	orderProjectsByQuery(items, query)
	return items, nil
}

func (a *App) searchTexturePacks(ctx context.Context, query string, opts providerSearchOptions) ([]UnifiedProject, error) {
	base := providerBase("MMV_TEXTUREPACKS_BASE", "https://texture-packs.com")
	body, err := a.getText(ctx, base+"/?s="+url.QueryEscape(query), nil)
	if err != nil {
		return nil, err
	}
	re := regexp.MustCompile(`(?is)<a[^>]+href=["']((?:https?://[^"']*texture-packs\.com)?/resourcepack/[^"'?#]+/)["'][^>]*>(.*?)</a>`)
	items := parseGenericCards(body, "texturepacks", base, re, "resourcepack", providerResultLimit(opts, 40))
	for i := range items {
		items[i].Categories = uniqueStringsPreserve(append(items[i].Categories, "resourcepack", "texture-pack"))
		items[i].InstallMode = "verified-detected-download"
		if v := matchMinecraftVersion(items[i].Title + " " + items[i].Summary); v != "" {
			items[i].Versions = []string{v}
		}
	}
	orderProjectsByQuery(items, query)
	return items, nil
}

func providerResultLimit(opts providerSearchOptions, fallback int) int {
	if opts.Limit > 0 && opts.Limit < fallback {
		return opts.Limit
	}
	return fallback
}

func matchMinecraftVersion(text string) string {
	m := regexp.MustCompile(`(?i)\b(?:MC\s*)?((?:1\.\d{1,2}(?:\.\d{1,2})?)|(?:2[6-9]\.\d{1,2}(?:\.\d{1,2})?))\b`).FindStringSubmatch(text)
	if len(m) > 1 {
		return m[1]
	}
	return ""
}
