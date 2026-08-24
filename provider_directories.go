package main

import (
	"context"
	"errors"
	"fmt"
	"image"
	_ "image/png"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// This file contains first-class integrations for focused Minecraft directories
// that are valuable enough to deserve their own source lane even though they do
// not expose a stable public JSON API. They are queried and rendered inside Vault;
// their web pages are used only as the provider transport, never as the primary UI.

func fetchFirstProviderPage(ctx context.Context, a *App, endpoints []string) (string, string, error) {
	var lastErr error
	for _, endpoint := range endpoints {
		body, err := a.getText(ctx, endpoint, nil)
		if err != nil {
			lastErr = err
			continue
		}
		if strings.TrimSpace(body) != "" {
			return body, endpoint, nil
		}
	}
	if lastErr != nil {
		return "", "", lastErr
	}
	return "", "", errors.New("provider returned an empty index")
}

func projectTextMatchesQuery(p UnifiedProject, query string) bool {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" || q == "minecraft" || q == "minecraft mods" || q == "minecraft shaders" || q == "minecraft skins" {
		return true
	}
	hay := strings.ToLower(strings.Join([]string{p.Title, p.Summary, p.Author, p.Slug, strings.Join(p.Categories, " "), strings.Join(p.Versions, " ")}, " "))
	meaningful := 0
	matched := 0
	for _, token := range strings.Fields(q) {
		token = strings.Trim(token, "-_.,:;!?()[]{}\"'")
		if len(token) < 2 || token == "minecraft" || token == "mod" || token == "mods" || token == "shader" || token == "shaders" || token == "skin" || token == "skins" {
			continue
		}
		meaningful++
		if strings.Contains(hay, token) {
			matched++
		}
	}
	return meaningful == 0 || matched > 0 || strings.Contains(hay, q)
}

func trimProjectsToQuery(items []UnifiedProject, query string, limit int) []UnifiedProject {
	filtered := make([]UnifiedProject, 0, len(items))
	for _, item := range items {
		if projectTextMatchesQuery(item, query) {
			filtered = append(filtered, item)
		}
	}
	orderProjectsByQuery(filtered, query)
	if limit > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}
	return filtered
}

func (a *App) searchMCreator(ctx context.Context, query string, opts providerSearchOptions) ([]UnifiedProject, error) {
	base := providerBase("MMV_MCREATOR_BASE", "https://mcreator.net")
	endpoints := []string{
		base + "/modifications?title=" + url.QueryEscape(strings.TrimSpace(query)),
		base + "/modifications?combine=" + url.QueryEscape(strings.TrimSpace(query)),
		base + "/modifications?sort_by=changed&sort_order=DESC",
	}
	body, _, err := fetchFirstProviderPage(ctx, a, endpoints)
	if err != nil {
		return nil, err
	}
	re := regexp.MustCompile(`(?is)<a[^>]+href=["']((?:https?://[^"']*mcreator\.net)?/modification/\d+/[^"'?#]+)["'][^>]*>(.*?)</a>`)
	items := parseGenericCards(body, "mcreator", base, re, "mod", maxInt(providerResultLimit(opts, 60), 30))
	for i := range items {
		text := strings.ToLower(items[i].Title + " " + items[i].Summary)
		if strings.Contains(text, "bedrock") || strings.Contains(text, "add-on") || strings.Contains(text, "addon") || regexp.MustCompile(`(?i)(^|\s)BE(?:\s|$)`).MatchString(items[i].Summary) {
			items[i].ProjectType = "addon"
		}
		if opts.ProjectType != "" && opts.ProjectType != "all" && !projectTypeMatches(items[i].ProjectType, opts.ProjectType) {
			continue
		}
		if v := matchMinecraftVersion(items[i].Title + " " + items[i].Summary); v != "" {
			items[i].Versions = []string{v}
		}
		items[i].Categories = uniqueStringsPreserve(append(items[i].Categories, "mcreator", "community-made"))
		items[i].Installable = true
		items[i].InstallMode = "verified-detected-download"
		items[i].Reason = "Live MCreator community modification index"
	}
	filtered := items[:0]
	for _, item := range items {
		if opts.ProjectType != "" && opts.ProjectType != "all" && !projectTypeMatches(item.ProjectType, opts.ProjectType) {
			continue
		}
		filtered = append(filtered, item)
	}
	return trimProjectsToQuery(filtered, query, providerResultLimit(opts, 60)), nil
}

func shaderDirectoryItems(body, provider, base, query string, opts providerSearchOptions, pathPattern *regexp.Regexp) []UnifiedProject {
	items := parseGenericCards(body, provider, base, pathPattern, "shader", maxInt(providerResultLimit(opts, 80), 50))
	out := make([]UnifiedProject, 0, len(items))
	for _, item := range items {
		pathURL, _ := url.Parse(item.PageURL)
		path := strings.Trim(pathURL.Path, "/")
		low := strings.ToLower(item.Title + " " + item.Summary + " " + path)
		if !strings.Contains(low, "shader") || path == "" || strings.Contains(path, "/") {
			continue
		}
		switch path {
		case "about", "contact", "privacy-policy", "terms", "category", "browse", "shaders", "shaderpacks", "downloads", "faq":
			continue
		}
		item.ProjectType = "shader"
		item.Categories = uniqueStringsPreserve(append(item.Categories, "shader", "visuals", provider))
		item.Installable = true
		item.InstallMode = "verified-detected-download"
		item.Reason = "Live dedicated shader directory"
		if v := matchMinecraftVersion(item.Title + " " + item.Summary); v != "" {
			item.Versions = []string{v}
		}
		out = append(out, item)
	}
	return trimProjectsToQuery(out, query, providerResultLimit(opts, 60))
}

func (a *App) searchShaderPacksCom(ctx context.Context, query string, opts providerSearchOptions) ([]UnifiedProject, error) {
	base := providerBase("MMV_SHADERPACKS_COM_BASE", "https://shaderpacks.com")
	body, _, err := fetchFirstProviderPage(ctx, a, []string{
		base + "/browse/shaders?search=" + url.QueryEscape(query),
		base + "/?s=" + url.QueryEscape(query),
		base + "/",
	})
	if err != nil {
		return nil, err
	}
	re := regexp.MustCompile(`(?is)<a[^>]+href=["']((?:https?://[^"']*shaderpacks\.com)?/[a-z0-9][a-z0-9-]*/?)["'][^>]*>(.*?)</a>`)
	return shaderDirectoryItems(body, "shaderpackscom", base, query, opts, re), nil
}

func (a *App) searchShaderPacksNet(ctx context.Context, query string, opts providerSearchOptions) ([]UnifiedProject, error) {
	base := providerBase("MMV_SHADERPACKS_NET_BASE", "https://shaderpacks.net")
	body, _, err := fetchFirstProviderPage(ctx, a, []string{
		base + "/?s=" + url.QueryEscape(query),
		base + "/category/shaderpacks/",
	})
	if err != nil {
		return nil, err
	}
	re := regexp.MustCompile(`(?is)<a[^>]+href=["']((?:https?://[^"']*shaderpacks\.net)?/[a-z0-9][a-z0-9-]*/)["'][^>]*>(.*?)</a>`)
	return shaderDirectoryItems(body, "shaderpacksnet", base, query, opts, re), nil
}

func (a *App) searchMinecraftShader(ctx context.Context, query string, opts providerSearchOptions) ([]UnifiedProject, error) {
	base := providerBase("MMV_MINECRAFTSHADER_BASE", "https://minecraftshader.com")
	body, _, err := fetchFirstProviderPage(ctx, a, []string{
		base + "/?s=" + url.QueryEscape(query),
		base + "/category/minecraft-shaders/",
	})
	if err != nil {
		return nil, err
	}
	re := regexp.MustCompile(`(?is)<a[^>]+href=["']((?:https?://[^"']*minecraftshader\.com)?/[a-z0-9][a-z0-9-]*/)["'][^>]*>(.*?)</a>`)
	return shaderDirectoryItems(body, "minecraftshader", base, query, opts, re), nil
}

func (a *App) searchSkindex(ctx context.Context, query string, opts providerSearchOptions) ([]UnifiedProject, error) {
	base := providerBase("MMV_SKINDEX_BASE", "https://www.minecraftskins.com")
	slug := strings.Trim(strings.ToLower(regexp.MustCompile(`[^a-zA-Z0-9]+`).ReplaceAllString(strings.TrimSpace(query), "-")), "-")
	if slug == "" || slug == "minecraft" || slug == "minecraft-skins" {
		slug = "cute"
	}
	endpoints := []string{
		base + "/search/skin/" + url.PathEscape(slug) + "/1/",
		base + "/top/",
	}
	body, _, err := fetchFirstProviderPage(ctx, a, endpoints)
	if err != nil {
		return nil, err
	}
	re := regexp.MustCompile(`(?is)<a[^>]+href=["']((?:https?://[^"']*minecraftskins\.com)?/skin/\d+/[^"'?#]+/)["'][^>]*>(.*?)</a>`)
	items := parseGenericCards(body, "skindex", base, re, "skin", maxInt(providerResultLimit(opts, 72), 48))
	for i := range items {
		if m := regexp.MustCompile(`/skin/(\d+)/`).FindStringSubmatch(items[i].PageURL); len(m) > 1 {
			items[i].ID = m[1]
			items[i].Slug = m[1]
		}
		items[i].ProjectType = "skin"
		items[i].Categories = uniqueStringsPreserve(append(items[i].Categories, "skin", "cosmetic", "skindex"))
		items[i].Installable = true
		items[i].InstallMode = "skin-png"
		items[i].Reason = "Live Skindex skin catalog"
	}
	return trimProjectsToQuery(items, query, providerResultLimit(opts, 60)), nil
}

func skinPNGFromHTML(body, pageURL string) string {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?is)(https://www\.minecraftskins\.com/uploads/skins/[^"'<>? ]+\.png(?:\?[^"'<> ]*)?)`),
		regexp.MustCompile(`(?is)<img[^>]+(?:src|data-src)=["']([^"']*/uploads/skins/[^"']+\.png(?:\?[^"']*)?)["']`),
	}
	for _, re := range patterns {
		if m := re.FindStringSubmatch(body); len(m) > 1 {
			return absoluteURL(pageURL, m[1])
		}
	}
	return ""
}

func validateMinecraftSkinImage(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	cfg, format, err := image.DecodeConfig(f)
	if err != nil {
		return fmt.Errorf("decode skin image: %w", err)
	}
	if format != "png" {
		return fmt.Errorf("skin must be a PNG, got %s", format)
	}
	if !((cfg.Width == 64 && cfg.Height == 64) || (cfg.Width == 64 && cfg.Height == 32)) {
		return fmt.Errorf("unexpected Minecraft skin dimensions %dx%d", cfg.Width, cfg.Height)
	}
	return nil
}

func (a *App) installSkindexSkin(ctx context.Context, pageURL, id string) (map[string]any, error) {
	if strings.TrimSpace(pageURL) == "" {
		return nil, errors.New("Skindex project page is required")
	}
	body, err := a.getText(ctx, pageURL, nil)
	if err != nil {
		return nil, err
	}
	skinURL := skinPNGFromHTML(body, pageURL)
	if skinURL == "" {
		return nil, errors.New("Skindex page did not expose the original skin PNG")
	}
	dir := filepath.Join(a.cfgDir, "skins")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	name := safeFilename(firstNonEmpty(metaContent(body, "og:title"), id, "minecraft-skin"))
	if name == "" {
		name = "minecraft-skin"
	}
	if !strings.HasSuffix(strings.ToLower(name), ".png") {
		name += ".png"
	}
	dst := uniquePath(filepath.Join(dir, name))
	if err := a.downloadURLVerified(ctx, skinURL, dst, 0, nil); err != nil {
		return nil, err
	}
	if err := validateMinecraftSkinImage(dst); err != nil {
		_ = os.Remove(dst)
		return nil, err
	}
	return map[string]any{
		"ok": true, "provider": "skindex", "project": id, "file": filepath.Base(dst), "path": dst, "target": "skins",
		"verification": "original PNG downloaded inside Vault and validated as a 64x64 or legacy 64x32 Minecraft skin",
	}, nil
}
