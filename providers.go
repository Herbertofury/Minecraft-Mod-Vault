package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type UnifiedProject struct {
	ID              string            `json:"id"`
	Provider        string            `json:"provider"`
	Providers       []string          `json:"providers,omitempty"`
	ProjectType     string            `json:"projectType"`
	Slug            string            `json:"slug,omitempty"`
	Title           string            `json:"title"`
	Summary         string            `json:"summary"`
	Author          string            `json:"author,omitempty"`
	AuthorAvatarURL string            `json:"authorAvatarUrl,omitempty"`
	IconURL         string            `json:"iconUrl,omitempty"`
	Gallery         []string          `json:"gallery,omitempty"`
	Downloads       int64             `json:"downloads,omitempty"`
	Followers       int64             `json:"followers,omitempty"`
	DateUpdated     string            `json:"dateUpdated,omitempty"`
	Categories      []string          `json:"categories,omitempty"`
	Versions        []string          `json:"versions,omitempty"`
	Loaders         []string          `json:"loaders,omitempty"`
	PageURL         string            `json:"pageUrl"`
	Installable     bool              `json:"installable"`
	InstallMode     string            `json:"installMode,omitempty"`
	Score           float64           `json:"score,omitempty"`
	Reason          string            `json:"reason,omitempty"`
	Links           map[string]string `json:"links,omitempty"`
}

type ProviderSearchResponse struct {
	Results    []UnifiedProject  `json:"results"`
	Sources    map[string]int    `json:"sources"`
	Errors     map[string]string `json:"errors,omitempty"`
	Warnings   map[string]string `json:"warnings,omitempty"`
	Skipped    map[string]string `json:"skipped,omitempty"`
	Total      int               `json:"total"`
	Query      string            `json:"query"`
	Category   string            `json:"category,omitempty"`
	Refreshed  string            `json:"refreshed"`
	Live       bool              `json:"live"`
	Game       string            `json:"gameVersion,omitempty"`
	Loader     string            `json:"loader,omitempty"`
	ProjectTyp string            `json:"projectType,omitempty"`
	HasMore    bool              `json:"hasMore"`
	NextOffset int               `json:"nextOffset,omitempty"`
	PoolTotal  int               `json:"poolTotal,omitempty"`
}

type ProviderInstallRequest struct {
	Provider    string `json:"provider"`
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	GameVersion string `json:"gameVersion"`
	Loader      string `json:"loader"`
	Target      string `json:"target"`
	PageURL     string `json:"pageUrl,omitempty"`
}

type providerCacheEntry struct {
	At    time.Time
	Items []UnifiedProject
}

const (
	providerFreshCacheTTL   = 90 * time.Second
	providerStaleCacheTTL   = 24 * time.Hour
	providerSearchTimeout   = 15 * time.Second
	providerCacheMaxEntries = 512
)

func providerSearchCacheKey(source, query string, opts providerSearchOptions) string {
	return strings.Join([]string{strings.ToLower(source), strings.ToLower(strings.TrimSpace(query)), strings.ToLower(opts.Category), strings.ToLower(opts.GameVersion), strings.ToLower(opts.Loader), strings.ToLower(opts.ProjectType), strings.ToLower(opts.Sort), strconv.Itoa(opts.Limit), strconv.Itoa(opts.Offset)}, "\x1f")
}

func cloneUnifiedProjects(in []UnifiedProject) []UnifiedProject {
	out := make([]UnifiedProject, len(in))
	for i, p := range in {
		p.Providers = append([]string(nil), p.Providers...)
		p.Gallery = append([]string(nil), p.Gallery...)
		p.Categories = append([]string(nil), p.Categories...)
		p.Versions = append([]string(nil), p.Versions...)
		p.Loaders = append([]string(nil), p.Loaders...)
		if p.Links != nil {
			p.Links = map[string]string{}
			for k, v := range in[i].Links {
				p.Links[k] = v
			}
		}
		out[i] = p
	}
	return out
}

func (a *App) getProviderCache(key string, maxAge time.Duration) ([]UnifiedProject, bool) {
	a.providerCacheMu.RLock()
	entry, ok := a.providerCache[key]
	a.providerCacheMu.RUnlock()
	if !ok || entry.At.IsZero() || time.Since(entry.At) > maxAge {
		return nil, false
	}
	return cloneUnifiedProjects(entry.Items), true
}

func (a *App) putProviderCache(key string, items []UnifiedProject) {
	a.providerCacheMu.Lock()
	if a.providerCache == nil {
		a.providerCache = map[string]providerCacheEntry{}
	}
	now := time.Now()
	for cacheKey, entry := range a.providerCache {
		if entry.At.IsZero() || now.Sub(entry.At) > providerStaleCacheTTL {
			delete(a.providerCache, cacheKey)
		}
	}
	if len(a.providerCache) >= providerCacheMaxEntries {
		oldestKey := ""
		var oldest time.Time
		for cacheKey, entry := range a.providerCache {
			if oldestKey == "" || entry.At.Before(oldest) {
				oldestKey, oldest = cacheKey, entry.At
			}
		}
		if oldestKey != "" {
			delete(a.providerCache, oldestKey)
		}
	}
	a.providerCache[key] = providerCacheEntry{At: now, Items: cloneUnifiedProjects(items)}
	a.providerCacheMu.Unlock()
}

type providerSearchOptions struct {
	Query       string
	Category    string
	GameVersion string
	Loader      string
	ProjectType string
	Limit       int
	Offset      int
	Sort        string
	Sources     []string
}

type TaxonomyCategory struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Icon     string   `json:"icon"`
	Queries  []string `json:"queries"`
	Modrinth []string `json:"modrinth,omitempty"`
	Group    string   `json:"group"`
}

var universalTaxonomy = []TaxonomyCategory{
	{ID: "all", Name: "All Mods", Icon: "globe", Queries: []string{"minecraft mods"}, Group: "Browse"},
	{ID: "technology", Name: "Technology & Automation", Icon: "gear", Queries: []string{"technology automation create industrial engineering factories"}, Modrinth: []string{"technology"}, Group: "Gameplay"},
	{ID: "redstone", Name: "Redstone & Engineering", Icon: "gear", Queries: []string{"redstone engineering machines logic circuitry contraptions"}, Modrinth: []string{"technology", "game-mechanics"}, Group: "Gameplay"},
	{ID: "magic", Name: "Magic & Spellcraft", Icon: "wand", Queries: []string{"magic spells sorcery wizard witch mana occult alchemy"}, Modrinth: []string{"magic"}, Group: "Gameplay"},
	{ID: "adventure", Name: "Adventure & RPG", Icon: "compass", Queries: []string{"adventure rpg quests dungeons progression skills classes"}, Modrinth: []string{"adventure"}, Group: "Gameplay"},
	{ID: "combat", Name: "Combat & Weapons", Icon: "hammer", Queries: []string{"combat weapons swords guns bows melee fighting"}, Modrinth: []string{"equipment", "game-mechanics"}, Group: "Gameplay"},
	{ID: "equipment", Name: "Armor, Tools & Equipment", Icon: "backpack", Queries: []string{"armor tools equipment accessories trinkets weapons"}, Modrinth: []string{"equipment"}, Group: "Gameplay"},
	{ID: "enchanting", Name: "Enchanting & Progression", Icon: "sparkle", Queries: []string{"enchanting enchantments progression skills leveling attributes"}, Modrinth: []string{"game-mechanics", "magic"}, Group: "Gameplay"},
	{ID: "quests", Name: "Quests & Story", Icon: "cards", Queries: []string{"quests story missions progression adventure book"}, Modrinth: []string{"adventure", "game-mechanics"}, Group: "Gameplay"},
	{ID: "economy", Name: "Economy & Trading", Icon: "cards", Queries: []string{"economy shops trading currency marketplace villagers"}, Modrinth: []string{"economy"}, Group: "Gameplay"},
	{ID: "worldgen", Name: "Worldgen & Exploration", Icon: "compass", Queries: []string{"world generation terrain biomes dimensions structures exploration"}, Modrinth: []string{"worldgen"}, Group: "World"},
	{ID: "biomes", Name: "Biomes & Dimensions", Icon: "globe", Queries: []string{"biomes dimensions realms worlds nether end terrain"}, Modrinth: []string{"worldgen", "adventure"}, Group: "World"},
	{ID: "structures", Name: "Structures & Dungeons", Icon: "hammer", Queries: []string{"structures dungeons towers villages ruins strongholds"}, Modrinth: []string{"worldgen", "adventure"}, Group: "World"},
	{ID: "villages", Name: "Villages & Civilization", Icon: "home", Queries: []string{"villages villagers colonies civilization towns npcs minecolonies"}, Modrinth: []string{"adventure", "game-mechanics"}, Group: "World"},
	{ID: "mobs", Name: "Mobs & Wildlife", Icon: "skull", Queries: []string{"mobs creatures wildlife animals monsters fauna"}, Modrinth: []string{"mobs"}, Group: "World"},
	{ID: "bosses", Name: "Bosses & Encounters", Icon: "skull", Queries: []string{"bosses boss fights raids encounters monsters"}, Modrinth: []string{"mobs", "adventure"}, Group: "World"},
	{ID: "pets", Name: "Pets & Companions", Icon: "paw", Queries: []string{"pets companions tameable animals dogs cats mounts followers"}, Modrinth: []string{"mobs"}, Group: "World"},
	{ID: "creature-collecting", Name: "Creature Collecting", Icon: "ball", Queries: []string{"pokemon creature collecting cobblemon pixelmon monsters capture battle"}, Group: "Niche"},
	{ID: "railroads", Name: "Railroads & Trains", Icon: "train", Queries: []string{"trains railroads railway locomotives create trains transit"}, Modrinth: []string{"transportation", "technology"}, Group: "Transport"},
	{ID: "vehicles", Name: "Vehicles & Ships", Icon: "car", Queries: []string{"vehicles cars ships planes boats helicopters transport"}, Modrinth: []string{"transportation"}, Group: "Transport"},
	{ID: "building", Name: "Building & Architecture", Icon: "hammer", Queries: []string{"building architecture construction blocks roofs windows doors"}, Modrinth: []string{"decoration"}, Group: "Building"},
	{ID: "decoration", Name: "Decoration & Interiors", Icon: "chair", Queries: []string{"decoration interiors decor lamps paintings clutter blocks"}, Modrinth: []string{"decoration"}, Group: "Building"},
	{ID: "furniture", Name: "Furniture", Icon: "chair", Queries: []string{"furniture decor cozy kitchen bedroom chairs tables sofas"}, Modrinth: []string{"decoration"}, Group: "Building"},
	{ID: "cit", Name: "CIT & Custom Items", Icon: "sparkle", Queries: []string{"CIT custom item textures optifine cit resewn cottagecore furniture"}, Group: "Building"},
	{ID: "farming", Name: "Farming & Agriculture", Icon: "crop", Queries: []string{"farming agriculture crops animals seasons farmer delight"}, Modrinth: []string{"food", "game-mechanics"}, Group: "Survival"},
	{ID: "food", Name: "Food & Cooking", Icon: "crop", Queries: []string{"food cooking kitchen recipes meals baking drinks"}, Modrinth: []string{"food"}, Group: "Survival"},
	{ID: "fishing", Name: "Fishing & Aquatics", Icon: "droplets", Queries: []string{"fishing fish ocean aquatic boats water wildlife"}, Modrinth: []string{"game-mechanics", "mobs"}, Group: "Survival"},
	{ID: "archaeology", Name: "Archaeology & Collecting", Icon: "compass", Queries: []string{"archaeology fossils artifacts relics museums collecting"}, Modrinth: []string{"adventure", "game-mechanics"}, Group: "Niche"},
	{ID: "storage", Name: "Storage & Inventory", Icon: "backpack", Queries: []string{"storage inventory backpacks chests logistics drawers"}, Modrinth: []string{"storage", "management"}, Group: "Utility"},
	{ID: "qol", Name: "Quality of Life", Icon: "sparkle", Queries: []string{"quality of life utility convenience inventory tooltip interface"}, Modrinth: []string{"utility", "management"}, Group: "Utility"},
	{ID: "maps", Name: "Maps, Minimap & Navigation", Icon: "compass", Queries: []string{"minimap world map navigation waypoints compass atlas"}, Modrinth: []string{"utility", "management"}, Group: "Utility"},
	{ID: "library", Name: "Libraries & APIs", Icon: "pack", Queries: []string{"library api dependency framework core mod"}, Modrinth: []string{"library"}, Group: "Utility"},
	{ID: "performance", Name: "Performance & Optimization", Icon: "gauge", Queries: []string{"performance optimization fps memory server rendering sodium lithium"}, Modrinth: []string{"optimization"}, Group: "Utility"},
	{ID: "server", Name: "Server Utility & Administration", Icon: "settings", Queries: []string{"server utility administration permissions claims moderation performance"}, Modrinth: []string{"management", "utility"}, Group: "Multiplayer"},
	{ID: "social", Name: "Multiplayer & Social", Icon: "globe", Queries: []string{"multiplayer social voice chat friends teams parties emotes"}, Modrinth: []string{"social"}, Group: "Multiplayer"},
	{ID: "minigames", Name: "Minigames", Icon: "cards", Queries: []string{"minigames games party pvp arenas racing spleef"}, Modrinth: []string{"minigame"}, Group: "Multiplayer"},
	{ID: "visuals", Name: "Visuals & Immersion", Icon: "sparkle", Queries: []string{"visuals immersion animation shader atmosphere ambience graphics"}, Group: "Visual"},
	{ID: "shaders", Name: "Shaders & Lighting", Icon: "sparkle", Queries: []string{"shaders lighting iris optifine complementary bsl cinematic volumetric shadows reflections"}, Group: "Visual"},
	{ID: "cinematic", Name: "Cinematic & Photography", Icon: "sparkle", Queries: []string{"cinematic photography camera replay depth of field screenshots film lighting"}, Group: "Visual"},
	{ID: "physics", Name: "Physics & Simulation", Icon: "droplets", Queries: []string{"physics ragdoll item physics water physics block physics immersive simulation"}, Group: "Visual"},
	{ID: "weather", Name: "Weather, Seasons & Atmosphere", Icon: "droplets", Queries: []string{"weather seasons storms rain snow wind fog atmosphere climate"}, Group: "World"},
	{ID: "skins", Name: "Skins & Cosmetics", Icon: "sparkle", Queries: []string{"cute pastel cottagecore skins cosmetics outfits character skin"}, Group: "Cosmetics"},
	{ID: "schematics", Name: "Schematics & Litematica Builds", Icon: "hammer", Queries: []string{"schematic litematica builds houses structures blueprints worldedit"}, Group: "Building"},
	{ID: "plushies", Name: "Plushies & Tiny Decor", Icon: "sparkle", Queries: []string{"plushies plush toys cute tiny decor stuffed animals kawaii"}, Group: "Niche"},
	{ID: "kitchen", Name: "Kitchen & Bakery Decor", Icon: "chair", Queries: []string{"kitchen bakery cooking decor cafe food furniture cottagecore"}, Group: "Niche"},
	{ID: "dragons", Name: "Dragons & Mythical Creatures", Icon: "paw", Queries: []string{"dragons mythical creatures griffins fantasy pets tameable mounts"}, Group: "Niche"},
	{ID: "hud", Name: "HUD, UI & Interface", Icon: "settings", Queries: []string{"hud ui interface inventory menu tooltip crosshair minimap accessibility"}, Group: "Utility"},
	{ID: "animations", Name: "Animations & Movement", Icon: "sparkle", Queries: []string{"animations player animation movement first person entity animation"}, Group: "Visual"},
	{ID: "foliage", Name: "Living Grass & Foliage", Icon: "leaf", Queries: []string{"grass animation foliage leaves wind waving plants vegetation"}, Group: "Niche"},
	{ID: "particles", Name: "Particles & Water FX", Icon: "droplets", Queries: []string{"particles water splashes wakes rain ripple effects ambience physics"}, Group: "Niche"},
	{ID: "audio", Name: "Sound, Music & Ambience", Icon: "sparkle", Queries: []string{"sound music ambience environmental audio footsteps reverb"}, Group: "Visual"},
	{ID: "cards", Name: "Cards & Collectibles", Icon: "cards", Queries: []string{"card game trading cards collectibles tcg deck booster"}, Group: "Niche"},
	{ID: "cute", Name: "Cute & Cozy", Icon: "sparkle", Queries: []string{"cute cozy cottagecore pastel plushies decor furniture pets kawaii"}, Group: "Niche"},
	{ID: "horror", Name: "Horror & Dark Fantasy", Icon: "skull", Queries: []string{"horror scary dark fantasy monsters ambience survival"}, Modrinth: []string{"adventure", "mobs"}, Group: "Niche"},
	{ID: "space", Name: "Space & Sci-Fi", Icon: "globe", Queries: []string{"space planets rockets astronomy sci fi technology dimensions"}, Modrinth: []string{"technology", "adventure"}, Group: "Niche"},
	{ID: "computers", Name: "Computers & Programming", Icon: "gear", Queries: []string{"computers programming lua computer craft networking automation"}, Modrinth: []string{"technology"}, Group: "Niche"},
}

func (a *App) handleTaxonomy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()
	categories, errs := a.liveTaxonomy(ctx)
	writeJSON(w, http.StatusOK, map[string]any{
		"categories": categories,
		"providers":  allProviderIDs(false),
		"live":       true,
		"errors":     errs,
		"refreshed":  time.Now().UTC().Format(time.RFC3339),
	})
}

func (a *App) liveTaxonomy(ctx context.Context) ([]TaxonomyCategory, map[string]string) {
	categories := append([]TaxonomyCategory(nil), universalTaxonomy...)
	errs := map[string]string{}
	seen := map[string]bool{}
	seenModrinth := map[string]bool{}
	for _, c := range categories {
		seen[normalizeTaxonomyName(c.Name)] = true
		for _, tag := range c.Modrinth {
			seenModrinth[strings.ToLower(strings.TrimSpace(tag))] = true
		}
	}

	type taxonomyResult struct {
		provider string
		items    []TaxonomyCategory
		err      error
	}
	ch := make(chan taxonomyResult, 2)
	go func() {
		var tags []struct {
			Name        string `json:"name"`
			ProjectType string `json:"project_type"`
			Header      string `json:"header"`
		}
		err := a.getJSON(ctx, modrinthAPIBase()+"/tag/category", nil, &tags)
		items := []TaxonomyCategory{}
		if err == nil {
			for _, tag := range tags {
				name := strings.TrimSpace(tag.Name)
				if name == "" || (tag.ProjectType != "" && tag.ProjectType != "mod" && tag.ProjectType != "modpack" && tag.ProjectType != "resourcepack" && tag.ProjectType != "shader") {
					continue
				}
				label := titleFromSlug(name)
				items = append(items, TaxonomyCategory{ID: "mr:" + name, Name: label, Icon: "pack", Queries: []string{label + " minecraft"}, Modrinth: []string{name}, Group: "Modrinth live"})
			}
		}
		ch <- taxonomyResult{provider: "modrinth", items: items, err: err}
	}()
	go func() {
		a.mu.RLock()
		key := strings.TrimSpace(a.settings.CurseForgeAPIKey)
		a.mu.RUnlock()
		if key == "" {
			body, err := a.getText(ctx, curseForgeWebBase()+"/minecraft/mc-mods", nil)
			items := parseCurseForgePublicTaxonomy(body)
			ch <- taxonomyResult{provider: "curseforge", items: items, err: err}
			return
		}
		var resp struct {
			Data []struct {
				ID       int64  `json:"id"`
				Name     string `json:"name"`
				Slug     string `json:"slug"`
				IsClass  bool   `json:"isClass"`
				ClassID  int64  `json:"classId"`
				ParentID int64  `json:"parentCategoryId"`
			} `json:"data"`
		}
		err := a.getJSON(ctx, curseForgeAPIBase()+"/v1/categories?gameId=432&classesOnly=false", map[string]string{"x-api-key": key}, &resp)
		items := []TaxonomyCategory{}
		if err == nil {
			for _, c := range resp.Data {
				if c.IsClass || strings.TrimSpace(c.Name) == "" {
					continue
				}
				slug := strings.TrimSpace(c.Slug)
				if slug == "" {
					slug = strings.ToLower(strings.ReplaceAll(c.Name, " ", "-"))
				}
				items = append(items, TaxonomyCategory{ID: fmt.Sprintf("cf:%d:%s", c.ID, slug), Name: c.Name, Icon: "pack", Queries: []string{c.Name + " minecraft"}, Group: "CurseForge live"})
			}
		}
		ch <- taxonomyResult{provider: "curseforge", items: items, err: err}
	}()

	for i := 0; i < 2; i++ {
		res := <-ch
		if res.err != nil {
			errs[res.provider] = res.err.Error()
			continue
		}
		for _, c := range res.items {
			key := normalizeTaxonomyName(c.Name)
			if key == "" || seen[key] {
				continue
			}
			if res.provider == "modrinth" && len(c.Modrinth) > 0 && seenModrinth[strings.ToLower(strings.TrimSpace(c.Modrinth[0]))] {
				continue
			}
			seen[key] = true
			for _, tag := range c.Modrinth {
				seenModrinth[strings.ToLower(strings.TrimSpace(tag))] = true
			}
			categories = append(categories, c)
		}
	}
	return categories, errs
}

func normalizeTaxonomyName(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(s, "")
	return s
}

func titleFromSlug(s string) string {
	s = strings.ReplaceAll(strings.ReplaceAll(s, "-", " "), "_", " ")
	parts := strings.Fields(s)
	for i := range parts {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, " ")
}

func dynamicCategoryQuery(id string) string {
	if strings.HasPrefix(id, "mr:") {
		return titleFromSlug(strings.TrimPrefix(id, "mr:"))
	}
	if strings.HasPrefix(id, "cfweb:") {
		return titleFromSlug(strings.TrimPrefix(id, "cfweb:"))
	}
	if strings.HasPrefix(id, "cf:") {
		parts := strings.SplitN(id, ":", 3)
		if len(parts) == 3 {
			return titleFromSlug(parts[2])
		}
	}
	return ""
}

func curseForgeCategoryID(id string) string {
	if !strings.HasPrefix(id, "cf:") {
		return ""
	}
	parts := strings.SplitN(id, ":", 3)
	if len(parts) < 2 {
		return ""
	}
	if _, err := strconv.ParseInt(parts[1], 10, 64); err != nil {
		return ""
	}
	return parts[1]
}

func (a *App) handleProviderSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	a.mu.RLock()
	defaults := a.settings
	a.mu.RUnlock()
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 || limit > 120 {
		limit = 48
	}
	offset, _ := strconv.Atoi(q.Get("offset"))
	sources := splitCSV(q.Get("sources"))
	if len(sources) == 0 {
		sources = append([]string(nil), defaults.EnabledSources...)
	}
	opts := providerSearchOptions{
		Query: strings.TrimSpace(q.Get("q")), Category: strings.TrimSpace(q.Get("category")),
		GameVersion: strings.TrimSpace(q.Get("gameVersion")), Loader: strings.TrimSpace(q.Get("loader")),
		ProjectType: strings.TrimSpace(q.Get("type")), Limit: limit, Offset: offset, Sort: strings.TrimSpace(q.Get("sort")), Sources: sources,
	}
	if opts.GameVersion == "" {
		opts.GameVersion = defaults.GameVersion
	}
	if opts.Loader == "" {
		opts.Loader = defaults.Loader
	}
	out := a.searchProviders(r.Context(), opts)
	writeJSON(w, http.StatusOK, out)
}

func (a *App) searchProviders(ctx context.Context, opts providerSearchOptions) ProviderSearchResponse {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	query := strings.TrimSpace(opts.Query)
	if query == "" && opts.Category != "" && opts.Category != "all" {
		if cat := taxonomyByID(opts.Category); cat != nil && len(cat.Queries) > 0 {
			query = cat.Queries[0]
		} else if dynamic := dynamicCategoryQuery(opts.Category); dynamic != "" {
			query = dynamic
		}
	}
	if query == "" {
		query = "minecraft"
	}
	if opts.Offset < 0 {
		opts.Offset = 0
	}
	if opts.Limit <= 0 {
		opts.Limit = 48
	}
	// Federated pagination must rank a stable prefix of every source, then slice
	// the merged pool. Asking each provider for offset=N and slicing again would
	// skip results and make Load More unstable. Grow the per-provider window from
	// zero instead, with a hard ceiling that keeps a single request bounded.
	wantedWindow := opts.Offset + opts.Limit
	if wantedWindow < opts.Limit { // integer overflow guard
		wantedWindow = opts.Limit
	}
	if wantedWindow > 240 {
		wantedWindow = 240
	}
	out := ProviderSearchResponse{Sources: map[string]int{}, Errors: map[string]string{}, Warnings: map[string]string{}, Skipped: map[string]string{}, Query: query, Category: opts.Category, Refreshed: time.Now().UTC().Format(time.RFC3339), Live: true, Game: opts.GameVersion, Loader: opts.Loader, ProjectTyp: opts.ProjectType}

	type result struct {
		source  string
		items   []UnifiedProject
		err     error
		warning string
		more    bool
	}
	ch := make(chan result, len(opts.Sources))
	var wg sync.WaitGroup
	for _, source := range uniqueStrings(opts.Sources) {
		source = strings.ToLower(source)
		if !providerSupportsProjectType(source, opts.ProjectType) {
			out.Skipped[source] = "provider does not publish " + strings.TrimSpace(opts.ProjectType) + " content"
			continue
		}
		wg.Add(1)
		go func(source string) {
			defer wg.Done()
			providerCtx, providerCancel := context.WithTimeout(ctx, providerSearchTimeout)
			defer providerCancel()
			started := time.Now()
			windowOpts := opts
			windowOpts.Offset = 0
			windowOpts.Limit = wantedWindow
			cacheKey := providerSearchCacheKey(source, query, windowOpts)
			if cached, ok := a.getProviderCache(cacheKey, providerFreshCacheTTL); ok {
				ch <- result{source: source, items: cached, more: providerWindowMayHaveMore(source, cached, wantedWindow)}
				return
			}
			items, more, err := a.searchProviderWindow(providerCtx, source, query, opts, wantedWindow)
			a.noteProviderAttempt(source, started, len(items), err)
			if err == nil {
				a.putProviderCache(cacheKey, items)
				ch <- result{source: source, items: items, more: more}
				return
			}
			if stale, ok := a.getProviderCache(cacheKey, providerStaleCacheTTL); ok {
				ch <- result{source: source, items: stale, warning: "live provider request failed; showing the latest cached result: " + err.Error(), more: providerWindowMayHaveMore(source, stale, wantedWindow)}
				return
			}
			ch <- result{source: source, err: err}
		}(source)
	}
	go func() { wg.Wait(); close(ch) }()

	all := make([]UnifiedProject, 0, maxInt(opts.Limit*2, wantedWindow*2))
	providerMayHaveMore := false
	for res := range ch {
		if res.err != nil {
			out.Errors[res.source] = res.err.Error()
			continue
		}
		if res.warning != "" {
			out.Warnings[res.source] = res.warning
			out.Live = false
		}
		if res.more {
			providerMayHaveMore = true
		}
		out.Sources[res.source] = len(res.items)
		for i := range res.items {
			res.items[i].Score = projectScore(res.items[i], query, opts)
			all = append(all, res.items[i])
		}
	}
	all = mergeProviderDuplicates(all)
	if opts.ProjectType != "" && opts.ProjectType != "all" {
		filtered := all[:0]
		for _, p := range all {
			if projectTypeMatches(p.ProjectType, opts.ProjectType) {
				filtered = append(filtered, p)
			}
		}
		all = filtered
	}
	sort.SliceStable(all, func(i, j int) bool {
		switch opts.Sort {
		case "downloads":
			return all[i].Downloads > all[j].Downloads
		case "updated":
			return all[i].DateUpdated > all[j].DateUpdated
		case "name":
			return strings.ToLower(all[i].Title) < strings.ToLower(all[j].Title)
		default:
			return all[i].Score > all[j].Score
		}
	})
	out.Total = len(all)
	out.PoolTotal = len(all)
	if opts.Offset >= len(all) {
		out.Results = []UnifiedProject{}
		out.HasMore = providerMayHaveMore && wantedWindow < 240
		if out.HasMore {
			out.NextOffset = opts.Offset
		}
		return out
	}
	end := opts.Offset + opts.Limit
	if end > len(all) {
		end = len(all)
	}
	out.Results = all[opts.Offset:end]
	out.HasMore = end < len(all) || (providerMayHaveMore && wantedWindow < 240)
	if out.HasMore {
		out.NextOffset = end
	}
	return out
}

func providerWindowMayHaveMore(source string, items []UnifiedProject, wanted int) bool {
	return providerSupportsWindowPaging(source) && wanted < 240 && len(items) >= wanted
}

func providerSupportsWindowPaging(source string) bool {
	switch source {
	case "modrinth", "curseforge", "github", "hangar", "spigot", "builtbybit", "smithed", "spongeore":
		return true
	default:
		return false
	}
}

func providerPageSize(source string) int {
	switch source {
	case "hangar", "spongeore":
		return 25
	case "github", "spigot":
		return 30
	case "curseforge", "modrinth", "builtbybit", "smithed":
		return 40
	default:
		return 40
	}
}

func (a *App) searchProviderWindow(ctx context.Context, source, query string, opts providerSearchOptions, wanted int) ([]UnifiedProject, bool, error) {
	if wanted <= 0 {
		wanted = maxInt(opts.Limit, 40)
	}
	if wanted > 240 {
		wanted = 240
	}
	if !providerSupportsWindowPaging(source) {
		local := opts
		local.Offset = 0
		local.Limit = wanted
		items, err := a.searchProviderSource(ctx, source, query, local)
		return items, false, err
	}
	chunk := providerPageSize(source)
	accumulated := []UnifiedProject{}
	seen := map[string]bool{}
	mayHaveMore := false
	for offset := 0; offset < wanted; offset += chunk {
		local := opts
		local.Offset = offset
		local.Limit = chunk
		page, err := a.searchProviderSource(ctx, source, query, local)
		if err != nil {
			if len(accumulated) == 0 {
				return nil, false, err
			}
			return accumulated, true, nil
		}
		newCount := 0
		for _, item := range page {
			key := strings.ToLower(strings.TrimSpace(item.Provider + "\x1f" + firstNonEmpty(item.ID, item.PageURL, item.Slug, item.Title)))
			if seen[key] {
				continue
			}
			seen[key] = true
			accumulated = append(accumulated, item)
			newCount++
		}
		if len(page) < chunk || newCount == 0 {
			mayHaveMore = false
			break
		}
		mayHaveMore = true
		if len(accumulated) >= wanted {
			break
		}
	}
	return accumulated, mayHaveMore, nil
}

func (a *App) searchProviderSource(ctx context.Context, source, query string, opts providerSearchOptions) ([]UnifiedProject, error) {
	switch source {
	case "modrinth":
		return a.searchModrinthUnified(ctx, query, opts)
	case "curseforge":
		return a.searchCurseForge(ctx, query, opts)
	case "github":
		return a.searchGitHub(ctx, query, opts)
	case "smithed":
		return a.searchSmithed(ctx, query, opts)
	case "planetminecraft":
		return a.searchPlanetMinecraft(ctx, query, opts)
	case "mcpedl":
		return a.searchMCPEDL(ctx, query, opts)
	case "marketplace":
		return a.searchMarketplace(ctx, query, opts)
	case "hangar":
		return a.searchHangar(ctx, query, opts)
	case "spigot":
		return a.searchSpigot(ctx, query, opts)
	case "bukkitdev":
		return a.searchBukkitDev(ctx, query, opts)
	case "spongeore":
		return a.searchSpongeOre(ctx, query, opts)
	case "builtbybit":
		return a.searchBuiltByBit(ctx, query, opts)
	case "polymart":
		return a.searchPolymart(ctx, query, opts)
	case "moddb":
		return a.searchModDB(ctx, query, opts)
	case "atlauncher":
		return a.searchATLauncher(ctx, query, opts)
	case "technic":
		return a.searchTechnic(ctx, query, opts)
	case "ftb":
		return a.searchFTB(ctx, query, opts)
	case "nexusmods":
		return a.searchNexusMods(ctx, query, opts)
	case "vanillatweaks":
		return a.searchVanillaTweaks(ctx, query, opts)
	case "minecrafthub":
		return a.searchMinecraftHub(ctx, query, opts)
	case "minecraftmaps":
		return a.searchMinecraftMaps(ctx, query, opts)
	case "resourcepacknet":
		return a.searchResourcePackNet(ctx, query, opts)
	case "texturepacks":
		return a.searchTexturePacks(ctx, query, opts)
	case "mcreator":
		return a.searchMCreator(ctx, query, opts)
	case "shaderpackscom":
		return a.searchShaderPacksCom(ctx, query, opts)
	case "shaderpacksnet":
		return a.searchShaderPacksNet(ctx, query, opts)
	case "minecraftshader":
		return a.searchMinecraftShader(ctx, query, opts)
	case "skindex":
		return a.searchSkindex(ctx, query, opts)
	default:
		return nil, fmt.Errorf("unknown provider %q", source)
	}
}

func providerSupportsProjectType(providerID, requested string) bool {
	requested = strings.ToLower(strings.TrimSpace(requested))
	if requested == "" || requested == "all" {
		return true
	}
	p := providerInfoByID(providerID)
	if p == nil || len(p.ProjectTypes) == 0 {
		return true
	}
	for _, actual := range p.ProjectTypes {
		if projectTypeMatches(actual, requested) || strings.EqualFold(actual, requested) {
			return true
		}
	}
	return false
}

func projectTypeMatches(actual, requested string) bool {
	a := strings.ToLower(strings.TrimSpace(actual))
	r := strings.ToLower(strings.TrimSpace(requested))
	if a == r {
		return true
	}
	switch r {
	case "resourcepack":
		return a == "texturepack" || a == "texture-pack" || a == "resource-pack"
	case "shader":
		return a == "shaderpack" || a == "shader-pack"
	case "world":
		return a == "map"
	case "addon":
		return a == "add-on" || a == "bedrock-addon"
	case "plugin":
		return a == "server-plugin" || a == "bukkit-plugin" || a == "paper-plugin"
	}
	return false
}

func taxonomyByID(id string) *TaxonomyCategory {
	for i := range universalTaxonomy {
		if universalTaxonomy[i].ID == id {
			return &universalTaxonomy[i]
		}
	}
	return nil
}

func splitCSV(s string) []string {
	var out []string
	for _, x := range strings.Split(s, ",") {
		if x = strings.TrimSpace(x); x != "" {
			out = append(out, x)
		}
	}
	return out
}

func uniqueStrings(in []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.ToLower(strings.TrimSpace(s))
		if s != "" && !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

func projectScore(p UnifiedProject, query string, opts providerSearchOptions) float64 {
	score := 0.0
	q := strings.ToLower(query)
	title := strings.ToLower(p.Title)
	if q != "minecraft" && q != "minecraft mods" {
		if title == q {
			score += 80
		} else if strings.Contains(title, q) {
			score += 45
		} else {
			for _, token := range strings.Fields(q) {
				if len(token) > 2 && strings.Contains(title+" "+strings.ToLower(p.Summary)+" "+strings.Join(p.Categories, " "), token) {
					score += 5
				}
			}
		}
	}
	if p.Downloads > 0 {
		score += math.Log10(float64(p.Downloads)+1) * 8
	}
	if p.Followers > 0 {
		score += math.Log10(float64(p.Followers)+1) * 4
	}
	if t, err := time.Parse(time.RFC3339, p.DateUpdated); err == nil {
		days := time.Since(t).Hours() / 24
		if days < 30 {
			score += 18
		} else if days < 180 {
			score += 10
		} else if days < 365 {
			score += 4
		}
	}
	if p.IconURL != "" {
		score += 3
	}
	if len(p.Gallery) > 0 {
		score += 2
	}
	if p.Installable {
		score += 5
	}
	if opts.GameVersion != "" && containsFold(p.Versions, opts.GameVersion) {
		score += 12
	}
	if opts.Loader != "" && opts.Loader != "any" && containsFold(p.Loaders, opts.Loader) {
		score += 8
	}
	return score
}

func normalizeProjectTitle(s string) string {
	s = strings.ToLower(s)
	s = regexp.MustCompile(`\[[^\]]+\]|\([^\)]+\)`).ReplaceAllString(s, " ")
	s = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(s, "")
	return s
}

func mergeProviderDuplicates(items []UnifiedProject) []UnifiedProject {
	byTitle := map[string][]int{}
	out := make([]UnifiedProject, 0, len(items))
	for _, p := range items {
		titleKey := normalizeProjectTitle(p.Title)
		if titleKey == "" {
			titleKey = p.Provider + ":" + p.ID
		}
		mergeAt := -1
		for _, idx := range byTitle[titleKey] {
			if projectsLikelySame(out[idx], p) {
				mergeAt = idx
				break
			}
		}
		if mergeAt >= 0 {
			existing := out[mergeAt]
			if p.Score > existing.Score {
				out[mergeAt] = mergeUnifiedProjectMetadata(p, existing)
			} else {
				out[mergeAt] = mergeUnifiedProjectMetadata(existing, p)
			}
			continue
		}
		p.Providers = uniqueStrings(append(p.Providers, p.Provider))
		if p.Links == nil {
			p.Links = map[string]string{}
		}
		if p.Provider != "" && p.PageURL != "" {
			p.Links[p.Provider] = p.PageURL
		}
		idx := len(out)
		byTitle[titleKey] = append(byTitle[titleKey], idx)
		out = append(out, p)
	}
	return out
}

func mergeUnifiedProjectMetadata(primary, secondary UnifiedProject) UnifiedProject {
	primary.Providers = uniqueStrings(append(append(primary.Providers, primary.Provider), append(secondary.Providers, secondary.Provider)...))
	if primary.Links == nil {
		primary.Links = map[string]string{}
	}
	for k, v := range secondary.Links {
		if strings.TrimSpace(v) != "" {
			primary.Links[k] = v
		}
	}
	if secondary.Provider != "" && secondary.PageURL != "" {
		primary.Links[secondary.Provider] = secondary.PageURL
	}
	if primary.Provider != "" && primary.PageURL != "" {
		primary.Links[primary.Provider] = primary.PageURL
	}
	primary.Categories = uniqueStringsPreserve(append(primary.Categories, secondary.Categories...))
	primary.Versions = uniqueStringsPreserve(append(primary.Versions, secondary.Versions...))
	primary.Loaders = uniqueStrings(append(primary.Loaders, secondary.Loaders...))
	primary.Gallery = uniqueStringsPreserve(append(primary.Gallery, secondary.Gallery...))
	if primary.IconURL == "" {
		primary.IconURL = secondary.IconURL
	}
	if primary.Author == "" {
		primary.Author = secondary.Author
	}
	if primary.AuthorAvatarURL == "" {
		primary.AuthorAvatarURL = secondary.AuthorAvatarURL
	}
	if len(strings.TrimSpace(secondary.Summary)) > len(strings.TrimSpace(primary.Summary)) {
		primary.Summary = secondary.Summary
	}
	if secondary.Downloads > primary.Downloads {
		primary.Downloads = secondary.Downloads
	}
	if secondary.Followers > primary.Followers {
		primary.Followers = secondary.Followers
	}
	if secondary.DateUpdated > primary.DateUpdated {
		primary.DateUpdated = secondary.DateUpdated
	}
	if !primary.Installable && secondary.Installable {
		primary.Installable = true
		primary.InstallMode = secondary.InstallMode
	}
	if primary.Reason == "" {
		primary.Reason = secondary.Reason
	}
	return primary
}

func projectsLikelySame(a, b UnifiedProject) bool {
	if normalizeProjectTitle(a.Title) != normalizeProjectTitle(b.Title) {
		return false
	}
	if a.ProjectType != "" && b.ProjectType != "" && !projectTypeMatches(a.ProjectType, b.ProjectType) && !projectTypeMatches(b.ProjectType, a.ProjectType) {
		return false
	}
	sa, sb := normalizeProjectTitle(a.Slug), normalizeProjectTitle(b.Slug)
	if sa != "" && sb != "" && sa == sb {
		return true
	}
	aa, ab := normalizeProjectTitle(a.Author), normalizeProjectTitle(b.Author)
	if aa == "" || ab == "" {
		return true
	}
	return aa == ab || strings.Contains(aa, ab) || strings.Contains(ab, aa)
}

func containsFold(items []string, target string) bool {
	for _, s := range items {
		if strings.EqualFold(s, target) {
			return true
		}
	}
	return false
}

func (a *App) searchModrinthUnified(ctx context.Context, query string, opts providerSearchOptions) ([]UnifiedProject, error) {
	values := url.Values{}
	values.Set("query", query)
	limit := opts.Limit
	if limit > 40 {
		limit = 40
	}
	if limit < 12 {
		limit = 12
	}
	values.Set("limit", strconv.Itoa(limit))
	values.Set("offset", strconv.Itoa(maxInt(opts.Offset, 0)))
	switch strings.ToLower(opts.Sort) {
	case "downloads":
		values.Set("index", "downloads")
	case "updated":
		values.Set("index", "updated")
	case "newest":
		values.Set("index", "newest")
	default:
		values.Set("index", "relevance")
	}
	facets := [][]string{}
	if opts.ProjectType != "" && opts.ProjectType != "all" {
		pt := opts.ProjectType
		if pt == "resourcepack" || pt == "shader" || pt == "mod" || pt == "modpack" || pt == "datapack" || pt == "plugin" {
			facets = append(facets, []string{"project_type:" + pt})
		}
	}
	if opts.GameVersion != "" && opts.GameVersion != "any" {
		facets = append(facets, []string{"versions:" + opts.GameVersion})
	}
	if opts.Loader != "" && opts.Loader != "any" && opts.ProjectType != "resourcepack" && opts.ProjectType != "shader" {
		facets = append(facets, []string{"categories:" + opts.Loader})
	}
	if cat := taxonomyByID(opts.Category); cat != nil && len(cat.Modrinth) > 0 {
		orFacet := make([]string, 0, len(cat.Modrinth))
		for _, c := range cat.Modrinth {
			orFacet = append(orFacet, "categories:"+c)
		}
		facets = append(facets, orFacet)
	} else if strings.HasPrefix(opts.Category, "mr:") {
		if c := strings.TrimSpace(strings.TrimPrefix(opts.Category, "mr:")); c != "" {
			facets = append(facets, []string{"categories:" + c})
		}
	}
	if len(facets) > 0 {
		b, _ := json.Marshal(facets)
		values.Set("facets", string(b))
	}
	u := modrinthAPIBase() + "/search?" + values.Encode()
	var data struct {
		Hits []struct {
			ProjectID    string   `json:"project_id"`
			ProjectType  string   `json:"project_type"`
			Slug         string   `json:"slug"`
			Author       string   `json:"author"`
			Title        string   `json:"title"`
			Description  string   `json:"description"`
			Categories   []string `json:"categories"`
			Versions     []string `json:"versions"`
			Downloads    int64    `json:"downloads"`
			Follows      int64    `json:"follows"`
			IconURL      string   `json:"icon_url"`
			DateModified string   `json:"date_modified"`
			Gallery      []string `json:"gallery"`
		} `json:"hits"`
	}
	if err := a.getJSON(ctx, u, nil, &data); err != nil {
		return nil, err
	}
	out := make([]UnifiedProject, 0, len(data.Hits))
	for _, h := range data.Hits {
		pageType := h.ProjectType
		if pageType == "modpack" {
			pageType = "modpack"
		}
		p := UnifiedProject{ID: h.ProjectID, Provider: "modrinth", ProjectType: h.ProjectType, Slug: h.Slug, Title: h.Title, Summary: h.Description, Author: h.Author, IconURL: h.IconURL, Gallery: h.Gallery, Downloads: h.Downloads, Followers: h.Follows, DateUpdated: h.DateModified, Categories: h.Categories, Versions: h.Versions, PageURL: "https://modrinth.com/" + pageType + "/" + url.PathEscape(h.Slug), Installable: true, InstallMode: "native"}
		p.Loaders = filterLoaders(h.Categories)
		out = append(out, p)
	}
	// Resolve real author avatars for the first visible results. Failures are non-fatal.
	sem := make(chan struct{}, 6)
	var wg sync.WaitGroup
	for i := 0; i < len(out) && i < 24; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			if avatar := a.modrinthAuthorAvatar(ctx, out[i].ID, out[i].Author); avatar != "" {
				out[i].AuthorAvatarURL = avatar
			}
		}(i)
	}
	wg.Wait()
	return out, nil
}

func filterLoaders(categories []string) []string {
	loaders := []string{"fabric", "forge", "neoforge", "quilt", "liteloader", "rift", "bukkit", "spigot", "paper", "purpur"}
	var out []string
	for _, cat := range categories {
		if containsFold(loaders, cat) {
			out = append(out, strings.ToLower(cat))
		}
	}
	return uniqueStrings(out)
}

func (a *App) modrinthAuthorAvatar(ctx context.Context, projectID, author string) string {
	cacheKey := "modrinth:" + projectID
	a.avatarCacheMu.RLock()
	if cached := a.avatarCache[cacheKey]; cached != "" {
		a.avatarCacheMu.RUnlock()
		return cached
	}
	a.avatarCacheMu.RUnlock()
	var members []struct {
		User struct {
			Username  string `json:"username"`
			AvatarURL string `json:"avatar_url"`
		} `json:"user"`
		Role string `json:"role"`
	}
	if err := a.getJSON(ctx, modrinthAPIBase()+"/project/"+url.PathEscape(projectID)+"/members", nil, &members); err != nil {
		return ""
	}
	avatar := ""
	for _, m := range members {
		if strings.EqualFold(m.User.Username, author) && m.User.AvatarURL != "" {
			avatar = m.User.AvatarURL
			break
		}
	}
	if avatar == "" && len(members) > 0 {
		avatar = members[0].User.AvatarURL
	}
	if avatar != "" {
		a.avatarCacheMu.Lock()
		if a.avatarCache == nil {
			a.avatarCache = map[string]string{}
		}
		a.avatarCache[cacheKey] = avatar
		a.avatarCacheMu.Unlock()
	}
	return avatar
}

func (a *App) searchCurseForge(ctx context.Context, query string, opts providerSearchOptions) ([]UnifiedProject, error) {
	a.mu.RLock()
	key := strings.TrimSpace(a.settings.CurseForgeAPIKey)
	a.mu.RUnlock()
	// The CurseForge API path below is class-aware only for the Java mod class today.
	// For every other content lane, and for the combined "all" browser, use the
	// public integrated index so worlds, shaders, packs and customizations are not
	// silently mislabeled as mods.
	if key != "" && strings.EqualFold(strings.TrimSpace(opts.ProjectType), "mod") {
		return a.searchCurseForgeAPI(ctx, key, query, opts)
	}
	return a.searchCurseForgeHTML(ctx, query, opts)
}

func curseLoaderType(loader string) int {
	switch strings.ToLower(loader) {
	case "forge":
		return 1
	case "fabric":
		return 4
	case "quilt":
		return 5
	case "neoforge":
		return 6
	default:
		return 0
	}
}

func (a *App) searchCurseForgeAPI(ctx context.Context, key, query string, opts providerSearchOptions) ([]UnifiedProject, error) {
	sortField := "2" // popularity
	switch strings.ToLower(opts.Sort) {
	case "downloads":
		sortField = "6"
	case "updated", "newest":
		sortField = "3"
	case "name":
		sortField = "4"
	}
	pageSize := opts.Limit
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 40
	}
	v := url.Values{"gameId": {"432"}, "pageSize": {strconv.Itoa(pageSize)}, "index": {strconv.Itoa(maxInt(opts.Offset, 0))}, "searchFilter": {query}, "sortField": {sortField}, "sortOrder": {"desc"}}
	if opts.GameVersion != "" && opts.GameVersion != "any" {
		v.Set("gameVersion", opts.GameVersion)
	}
	if lt := curseLoaderType(opts.Loader); lt != 0 {
		v.Set("modLoaderType", strconv.Itoa(lt))
	}
	if categoryID := curseForgeCategoryID(opts.Category); categoryID != "" {
		v.Set("categoryId", categoryID)
	}
	var resp struct {
		Data []struct {
			ID          int64  `json:"id"`
			Name        string `json:"name"`
			Slug        string `json:"slug"`
			Summary     string `json:"summary"`
			DownloadCnt int64  `json:"downloadCount"`
			DateMod     string `json:"dateModified"`
			Logo        *struct {
				ThumbnailURL string `json:"thumbnailUrl"`
				URL          string `json:"url"`
			} `json:"logo"`
			Screenshots []struct {
				ThumbnailURL string `json:"thumbnailUrl"`
				URL          string `json:"url"`
			} `json:"screenshots"`
			Authors []struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"authors"`
			Categories []struct {
				Name string `json:"name"`
			} `json:"categories"`
			LatestFilesIndexes []struct {
				GameVersion string `json:"gameVersion"`
				ModLoader   int    `json:"modLoader"`
			} `json:"latestFilesIndexes"`
		} `json:"data"`
	}
	headers := map[string]string{"x-api-key": key}
	if err := a.getJSON(ctx, curseForgeAPIBase()+"/v1/mods/search?"+v.Encode(), headers, &resp); err != nil {
		return nil, err
	}
	out := make([]UnifiedProject, 0, len(resp.Data))
	for _, x := range resp.Data {
		p := UnifiedProject{ID: strconv.FormatInt(x.ID, 10), Provider: "curseforge", ProjectType: "mod", Slug: x.Slug, Title: x.Name, Summary: x.Summary, Downloads: x.DownloadCnt, DateUpdated: x.DateMod, PageURL: "https://www.curseforge.com/minecraft/mc-mods/" + url.PathEscape(x.Slug), Installable: true, InstallMode: "native-api"}
		if x.Logo != nil {
			p.IconURL = firstNonEmpty(x.Logo.ThumbnailURL, x.Logo.URL)
		}
		for _, shot := range x.Screenshots {
			if imageURL := firstNonEmpty(shot.URL, shot.ThumbnailURL); imageURL != "" {
				p.Gallery = append(p.Gallery, imageURL)
			}
		}
		if len(x.Authors) > 0 {
			p.Author = x.Authors[0].Name
		}
		for _, c := range x.Categories {
			p.Categories = append(p.Categories, c.Name)
		}
		for _, idx := range x.LatestFilesIndexes {
			p.Versions = append(p.Versions, idx.GameVersion)
			if l := curseLoaderName(idx.ModLoader); l != "" {
				p.Loaders = append(p.Loaders, l)
			}
		}
		p.Versions = uniqueStringsPreserve(p.Versions)
		p.Loaders = uniqueStrings(p.Loaders)
		out = append(out, p)
	}
	return out, nil
}

func curseLoaderName(v int) string {
	switch v {
	case 1:
		return "forge"
	case 4:
		return "fabric"
	case 5:
		return "quilt"
	case 6:
		return "neoforge"
	}
	return ""
}

var cfPublicCategoryRE = regexp.MustCompile(`(?is)<a[^>]+href="([^"]*?/minecraft/search\?[^"]*\bcategories=[^"]+)"[^>]*>(.*?)</a>`)

func parseCurseForgePublicTaxonomy(body string) []TaxonomyCategory {
	out := []TaxonomyCategory{}
	seen := map[string]bool{}
	for _, m := range cfPublicCategoryRE.FindAllStringSubmatch(body, -1) {
		if len(m) < 3 {
			continue
		}
		href := html.UnescapeString(m[1])
		u, err := url.Parse(href)
		if err != nil {
			continue
		}
		if class := u.Query().Get("class"); class != "" && class != "mc-mods" {
			continue
		}
		slug := strings.TrimSpace(u.Query().Get("categories"))
		if slug == "" || strings.Contains(slug, ",") || seen[slug] {
			continue
		}
		name := cleanHTMLText(m[2])
		if name == "" || strings.EqualFold(name, "all") {
			name = titleFromSlug(slug)
		}
		seen[slug] = true
		out = append(out, TaxonomyCategory{ID: "cfweb:" + slug, Name: name, Icon: "pack", Queries: []string{name + " minecraft"}, Group: "CurseForge live"})
	}
	return out
}

type curseForgeBrowseLane struct {
	SearchRoot  string
	Class       string
	ProjectType string
}

func curseForgeBrowseLanes(projectType string) []curseForgeBrowseLane {
	java := func(class, ptype string) curseForgeBrowseLane {
		return curseForgeBrowseLane{SearchRoot: "/minecraft/search", Class: class, ProjectType: ptype}
	}
	bedrock := func(class, ptype string) curseForgeBrowseLane {
		return curseForgeBrowseLane{SearchRoot: "/minecraft-bedrock/search", Class: class, ProjectType: ptype}
	}
	switch strings.ToLower(strings.TrimSpace(projectType)) {
	case "mod":
		return []curseForgeBrowseLane{java("mc-mods", "mod")}
	case "modpack":
		return []curseForgeBrowseLane{java("modpacks", "modpack")}
	case "resourcepack":
		return []curseForgeBrowseLane{java("texture-packs", "resourcepack"), bedrock("texture-packs", "resourcepack")}
	case "shader":
		return []curseForgeBrowseLane{java("shaders", "shader")}
	case "world":
		return []curseForgeBrowseLane{java("worlds", "world"), bedrock("maps", "world")}
	case "addon":
		return []curseForgeBrowseLane{java("mc-addons", "addon"), bedrock("addons", "addon"), bedrock("scripts", "addon")}
	case "skin":
		return []curseForgeBrowseLane{bedrock("skins", "skin")}
	case "tool":
		return []curseForgeBrowseLane{java("customization", "tool")}
	case "datapack":
		return []curseForgeBrowseLane{java("data-packs", "datapack")}
	case "plugin":
		return []curseForgeBrowseLane{java("bukkit-plugins", "plugin")}
	case "", "all":
		return []curseForgeBrowseLane{
			java("mc-mods", "mod"), java("modpacks", "modpack"), java("texture-packs", "resourcepack"),
			java("shaders", "shader"), java("worlds", "world"), java("customization", "tool"),
			java("data-packs", "datapack"), java("bukkit-plugins", "plugin"), java("mc-addons", "addon"),
			bedrock("addons", "addon"), bedrock("maps", "world"), bedrock("texture-packs", "resourcepack"),
			bedrock("scripts", "addon"), bedrock("skins", "skin"),
		}
	default:
		return nil
	}
}

func (a *App) searchCurseForgeHTML(ctx context.Context, query string, opts providerSearchOptions) ([]UnifiedProject, error) {
	lanes := curseForgeBrowseLanes(opts.ProjectType)
	if len(lanes) == 0 {
		return []UnifiedProject{}, nil
	}
	type laneResult struct {
		items []UnifiedProject
		err   error
	}
	ch := make(chan laneResult, len(lanes))
	for _, lane := range lanes {
		lane := lane
		go func() {
			pageSize := 40
			page := maxInt(opts.Offset/pageSize, 0) + 1
			values := url.Values{"class": {lane.Class}, "page": {strconv.Itoa(page)}, "pageSize": {strconv.Itoa(pageSize)}, "search": {query}, "sortBy": {"relevancy"}}
			if strings.HasPrefix(opts.Category, "cfweb:") {
				values.Set("categories", strings.TrimSpace(strings.TrimPrefix(opts.Category, "cfweb:")))
			}
			if opts.GameVersion != "" && !strings.EqualFold(opts.GameVersion, "any") {
				values.Set("version", opts.GameVersion)
			}
			u := curseForgeWebBase() + lane.SearchRoot + "?" + values.Encode()
			body, err := a.getText(ctx, u, nil)
			if err != nil {
				ch <- laneResult{err: err}
				return
			}
			items := parseCurseForgeSearchHTML(body)
			for i := range items {
				// The page class is authoritative when markup doesn't expose a
				// lane-specific project URL (or if the site changes a route).
				if items[i].ProjectType == "" || (items[i].ProjectType == "mod" && lane.ProjectType != "mod") {
					items[i].ProjectType = lane.ProjectType
				}
			}
			ch <- laneResult{items: items}
		}()
	}
	out := []UnifiedProject{}
	var errs []string
	for range lanes {
		r := <-ch
		if r.err != nil {
			errs = append(errs, r.err.Error())
			continue
		}
		out = append(out, r.items...)
	}
	if len(out) == 0 && len(errs) == len(lanes) {
		return nil, errors.New(strings.Join(uniqueStrings(errs), "; "))
	}
	return mergeProviderDuplicates(out), nil
}

var cfLinkRE = regexp.MustCompile(`(?is)<a[^>]+href="(/(?:minecraft/(?:mc-mods|texture-packs|modpacks|customization|worlds|shaders|data-packs|bukkit-plugins|mc-addons)|minecraft-bedrock/(?:addons|maps|texture-packs|scripts|skins))/[^"?#]+)"[^>]*>(.*?)</a>`)

func parseCurseForgeSearchHTML(body string) []UnifiedProject {
	matches := cfLinkRE.FindAllStringSubmatchIndex(body, -1)
	seen := map[string]bool{}
	out := []UnifiedProject{}
	for _, m := range matches {
		if len(m) < 6 {
			continue
		}
		href := body[m[2]:m[3]]
		if strings.Contains(href, "/download") || strings.Contains(href, "/files") || seen[href] {
			continue
		}
		seen[href] = true
		start := m[0] - 2200
		if start < 0 {
			start = 0
		}
		end := m[1] + 3500
		if end > len(body) {
			end = len(body)
		}
		chunk := body[start:end]
		title := cleanHTMLText(body[m[4]:m[5]])
		if title == "" || len(title) > 160 {
			title = attributeFromChunk(chunk, "alt")
		}
		if title == "" {
			continue
		}
		img := imageFromChunk(chunk)
		author := matchText(chunk, `(?is)(?:by|author)[^<]{0,20}<[^>]*>\s*([^<]{2,80})`)
		summary := matchText(chunk, `(?is)<p[^>]*>([^<]{10,500})</p>`)
		downloads := parseCompactNumber(matchText(chunk, `(?is)([0-9][0-9.,]*\s*[KMB]?)\s*(?:Downloads|download)`))
		ptype := "mod"
		switch {
		case strings.Contains(href, "/minecraft-bedrock/addons/"), strings.Contains(href, "/minecraft-bedrock/scripts/"):
			ptype = "addon"
		case strings.Contains(href, "/minecraft-bedrock/maps/"), strings.Contains(href, "/minecraft/worlds/"):
			ptype = "world"
		case strings.Contains(href, "/texture-packs/"):
			ptype = "resourcepack"
		case strings.Contains(href, "/minecraft-bedrock/skins/"):
			ptype = "skin"
		case strings.Contains(href, "/modpacks/"):
			ptype = "modpack"
		case strings.Contains(href, "/shaders/"):
			ptype = "shader"
		case strings.Contains(href, "/data-packs/"):
			ptype = "datapack"
		case strings.Contains(href, "/bukkit-plugins/"):
			ptype = "plugin"
		case strings.Contains(href, "/mc-addons/"):
			ptype = "addon"
		case strings.Contains(href, "/customization/"):
			ptype = "tool"
		}
		slug := strings.Trim(strings.TrimPrefix(filepath.Base(href), "/"), " ")
		out = append(out, UnifiedProject{ID: slug, Provider: "curseforge", ProjectType: ptype, Slug: slug, Title: title, Summary: summary, Author: author, IconURL: img, Downloads: downloads, PageURL: "https://www.curseforge.com" + href, Installable: true, InstallMode: "verified-detected-download", Reason: "Live CurseForge index with verified package detection"})
		if len(out) >= 40 {
			break
		}
	}
	return out
}

func githubSearchDescriptor(projectType string) (suffix, normalizedType string) {
	switch strings.ToLower(strings.TrimSpace(projectType)) {
	case "mod":
		return "minecraft mod", "mod"
	case "plugin":
		return "minecraft plugin", "plugin"
	case "datapack":
		return "minecraft datapack", "datapack"
	case "resourcepack":
		return "minecraft resource pack", "resourcepack"
	case "shader":
		return "minecraft shader", "shader"
	case "tool":
		return "minecraft tool", "tool"
	default:
		return "minecraft", ""
	}
}

func inferGitHubProjectType(text string) string {
	s := strings.ToLower(text)
	switch {
	case strings.Contains(s, "datapack") || strings.Contains(s, "data pack"):
		return "datapack"
	case strings.Contains(s, "resourcepack") || strings.Contains(s, "resource pack") || strings.Contains(s, "texture pack"):
		return "resourcepack"
	case strings.Contains(s, "shaderpack") || strings.Contains(s, "shader pack") || strings.Contains(s, " shader"):
		return "shader"
	case strings.Contains(s, "bukkit") || strings.Contains(s, "spigot") || strings.Contains(s, "paper plugin") || strings.Contains(s, "minecraft plugin"):
		return "plugin"
	case strings.Contains(s, "launcher") || strings.Contains(s, "mod manager") || strings.Contains(s, "minecraft tool"):
		return "tool"
	default:
		return "mod"
	}
}

func githubReleaseInstallableProjectType(projectType string) bool {
	switch strings.ToLower(strings.TrimSpace(projectType)) {
	case "mod", "plugin", "datapack", "resourcepack", "shader":
		return true
	default:
		return false
	}
}

func (a *App) searchGitHub(ctx context.Context, query string, opts providerSearchOptions) ([]UnifiedProject, error) {
	suffix, requestedType := githubSearchDescriptor(opts.ProjectType)
	q := strings.TrimSpace(strings.TrimSpace(query) + " " + suffix)
	apiBase := providerBase("MMV_GITHUB_API_BASE", "https://api.github.com")
	perPage := opts.Limit
	if perPage <= 0 || perPage > 100 {
		perPage = 30
	}
	page := maxInt(opts.Offset/perPage, 0) + 1
	apiURL := apiBase + "/search/repositories?" + url.Values{"q": {q}, "sort": {"stars"}, "order": {"desc"}, "per_page": {strconv.Itoa(perPage)}, "page": {strconv.Itoa(page)}}.Encode()
	a.mu.RLock()
	token := strings.TrimSpace(a.settings.GitHubToken)
	a.mu.RUnlock()
	headers := map[string]string{"Accept": "application/vnd.github+json", "X-GitHub-Api-Version": "2022-11-28"}
	if token != "" {
		headers["Authorization"] = "Bearer " + token
	}
	var resp struct {
		Items []struct {
			ID          int64  `json:"id"`
			Name        string `json:"name"`
			FullName    string `json:"full_name"`
			Description string `json:"description"`
			HTMLURL     string `json:"html_url"`
			UpdatedAt   string `json:"updated_at"`
			Stars       int64  `json:"stargazers_count"`
			Forks       int64  `json:"forks_count"`
			Owner       struct {
				Login     string `json:"login"`
				AvatarURL string `json:"avatar_url"`
			} `json:"owner"`
			Topics []string `json:"topics"`
		} `json:"items"`
	}
	if err := a.getJSON(ctx, apiURL, headers, &resp); err != nil {
		return nil, err
	}
	out := make([]UnifiedProject, 0, len(resp.Items))
	for _, x := range resp.Items {
		classificationText := x.Description + " " + strings.Join(x.Topics, " ") + " " + x.Name
		if !looksMinecraftRepo(classificationText) {
			continue
		}
		ptype := requestedType
		if ptype == "" {
			ptype = inferGitHubProjectType(classificationText)
		}
		installable := githubReleaseInstallableProjectType(ptype)
		out = append(out, UnifiedProject{ID: x.FullName, Provider: "github", ProjectType: ptype, Slug: x.FullName, Title: humanizeRepoName(x.Name), Summary: x.Description, Author: x.Owner.Login, AuthorAvatarURL: x.Owner.AvatarURL, IconURL: x.Owner.AvatarURL, Gallery: []string{"https://opengraph.githubassets.com/mmv/" + x.FullName}, Downloads: x.Stars, Followers: x.Forks, DateUpdated: x.UpdatedAt, Categories: x.Topics, PageURL: x.HTMLURL, Installable: installable, InstallMode: "github-release", Reason: "GitHub repository and verified release assets"})
	}
	return out, nil
}

func looksMinecraftRepo(s string) bool {
	s = strings.ToLower(s)
	return strings.Contains(s, "minecraft") || strings.Contains(s, "fabric") || strings.Contains(s, "forge") || strings.Contains(s, "neoforge") || strings.Contains(s, "modrinth") || strings.Contains(s, "curseforge")
}

func humanizeRepoName(s string) string {
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "_", " ")
	return strings.TrimSpace(s)
}

type planetMinecraftBrowseLane struct {
	IndexPath   string
	ProjectPath string
	ProjectType string
}

func planetMinecraftBrowseLanes(projectType string) []planetMinecraftBrowseLane {
	switch strings.ToLower(strings.TrimSpace(projectType)) {
	case "mod":
		return []planetMinecraftBrowseLane{{IndexPath: "/mods/", ProjectPath: "/mod/", ProjectType: "mod"}}
	case "datapack":
		return []planetMinecraftBrowseLane{{IndexPath: "/data-packs/", ProjectPath: "/data-pack/", ProjectType: "datapack"}}
	case "resourcepack":
		return []planetMinecraftBrowseLane{{IndexPath: "/texture-packs/", ProjectPath: "/texture-pack/", ProjectType: "resourcepack"}}
	case "world":
		return []planetMinecraftBrowseLane{{IndexPath: "/projects/", ProjectPath: "/project/", ProjectType: "world"}}
	case "skin":
		return []planetMinecraftBrowseLane{{IndexPath: "/skins/", ProjectPath: "/skin/", ProjectType: "skin"}}
	case "", "all":
		return []planetMinecraftBrowseLane{
			{IndexPath: "/mods/", ProjectPath: "/mod/", ProjectType: "mod"},
			{IndexPath: "/data-packs/", ProjectPath: "/data-pack/", ProjectType: "datapack"},
			{IndexPath: "/texture-packs/", ProjectPath: "/texture-pack/", ProjectType: "resourcepack"},
			{IndexPath: "/projects/", ProjectPath: "/project/", ProjectType: "world"},
			{IndexPath: "/skins/", ProjectPath: "/skin/", ProjectType: "skin"},
		}
	default:
		return nil
	}
}

func (a *App) searchPlanetMinecraft(ctx context.Context, query string, opts providerSearchOptions) ([]UnifiedProject, error) {
	lanes := planetMinecraftBrowseLanes(opts.ProjectType)
	if len(lanes) == 0 {
		return []UnifiedProject{}, nil
	}
	base := providerBase("MMV_PLANETMINECRAFT_BASE", "https://www.planetminecraft.com")
	type laneResult struct {
		items []UnifiedProject
		err   error
	}
	ch := make(chan laneResult, len(lanes))
	for _, lane := range lanes {
		lane := lane
		go func() {
			u := base + lane.IndexPath + "?keywords=" + url.QueryEscape(query)
			body, err := a.getText(ctx, u, nil)
			if err != nil {
				ch <- laneResult{err: err}
				return
			}
			pathRE := regexp.QuoteMeta(lane.ProjectPath)
			re := regexp.MustCompile(`(?is)<a[^>]+href="(` + pathRE + `[^"?#]+)"[^>]*>(.*?)</a>`)
			items := parseGenericCards(body, "planetminecraft", base, re, lane.ProjectType, 30)
			for i := range items {
				items[i].Installable = providerSupportsVerifiedDetectedInstall("planetminecraft") && lane.ProjectType != "skin"
				if items[i].Installable {
					items[i].InstallMode = "verified-detected-download"
					items[i].Reason = "Planet Minecraft project page with verified package detection"
				} else {
					items[i].InstallMode = "integrated-browse"
				}
			}
			ch <- laneResult{items: items}
		}()
	}
	out := []UnifiedProject{}
	var errs []string
	for range lanes {
		r := <-ch
		if r.err != nil {
			errs = append(errs, r.err.Error())
			continue
		}
		out = append(out, r.items...)
	}
	if len(out) == 0 && len(errs) == len(lanes) {
		return nil, errors.New(strings.Join(uniqueStrings(errs), "; "))
	}
	return mergeProviderDuplicates(out), nil
}

func (a *App) searchMCPEDL(ctx context.Context, query string, opts providerSearchOptions) ([]UnifiedProject, error) {
	u := "https://mcpedl.com/?s=" + url.QueryEscape(query)
	body, err := a.getText(ctx, u, nil)
	if err != nil {
		return nil, err
	}
	// MCPEDL article cards link to top-level slugs; exclude site navigation.
	re := regexp.MustCompile(`(?is)<a[^>]+href="(https://mcpedl\.com/[a-z0-9][^"?#]+/)"[^>]*>(.*?)</a>`)
	items := parseGenericCards(body, "mcpedl", "", re, "addon", 30)
	filtered := items[:0]
	for _, p := range items {
		low := strings.ToLower(p.PageURL)
		if strings.Contains(low, "/category/") || strings.Contains(low, "/tag/") || strings.Contains(low, "/author/") || strings.HasSuffix(low, "mcpedl.com/") {
			continue
		}
		p.ProjectType = "addon"
		p.Installable = false
		p.InstallMode = "details"
		filtered = append(filtered, p)
	}
	return filtered, nil
}

func (a *App) searchMarketplace(ctx context.Context, query string, opts providerSearchOptions) ([]UnifiedProject, error) {
	u := "https://www.minecraft.net/en-us/marketplace/search?search=" + url.QueryEscape(query)
	body, err := a.getText(ctx, u, nil)
	if err != nil {
		return nil, err
	}
	re := regexp.MustCompile(`(?is)<a[^>]+href="(/en-us/marketplace/pdp/[^"?#]+)"[^>]*>(.*?)</a>`)
	items := parseGenericCards(body, "marketplace", "https://www.minecraft.net", re, "addon", 30)
	for i := range items {
		items[i].Installable = false
		items[i].InstallMode = "marketplace"
	}
	return items, nil
}

func parseGenericCards(body, provider, base string, re *regexp.Regexp, projectType string, max int) []UnifiedProject {
	matches := re.FindAllStringSubmatchIndex(body, -1)
	seen := map[string]bool{}
	out := []UnifiedProject{}
	for _, m := range matches {
		if len(m) < 6 {
			continue
		}
		href := html.UnescapeString(body[m[2]:m[3]])
		if seen[href] {
			continue
		}
		seen[href] = true
		start := m[0] - 1800
		if start < 0 {
			start = 0
		}
		end := m[1] + 3000
		if end > len(body) {
			end = len(body)
		}
		chunk := body[start:end]
		title := cleanHTMLText(body[m[4]:m[5]])
		if title == "" || len(title) > 180 {
			title = attributeFromChunk(chunk, "alt")
		}
		if title == "" || strings.EqualFold(title, "download") || strings.EqualFold(title, "read more") {
			continue
		}
		pageURL := href
		if strings.HasPrefix(href, "/") {
			pageURL = strings.TrimRight(base, "/") + href
		}
		author := matchText(chunk, `(?is)(?:by|author)\s*</?[^>]*>?\s*([^<\n]{2,80})`)
		summary := matchText(chunk, `(?is)<p[^>]*>(.*?)</p>`)
		images := imagesFromChunkWithBase(chunk, base, 8)
		img := ""
		if len(images) > 0 {
			img = images[0]
		}
		out = append(out, UnifiedProject{ID: href, Provider: provider, ProjectType: projectType, Slug: strings.Trim(filepath.Base(strings.TrimRight(href, "/")), "/"), Title: title, Summary: summary, Author: author, IconURL: img, Gallery: images, PageURL: pageURL, Installable: false, InstallMode: "integrated-browse", Reason: "Live provider index"})
		if len(out) >= max {
			break
		}
	}
	return out
}

func nonEmptySlice(s string) []string {
	if s == "" {
		return nil
	}
	return []string{s}
}

func cleanHTMLText(s string) string {
	s = regexp.MustCompile(`(?is)<script.*?</script>|<style.*?</style>`).ReplaceAllString(s, " ")
	s = regexp.MustCompile(`(?s)<[^>]+>`).ReplaceAllString(s, " ")
	s = html.UnescapeString(s)
	return strings.Join(strings.Fields(s), " ")
}

func attributeFromChunk(chunk, name string) string {
	re := regexp.MustCompile(`(?is)` + regexp.QuoteMeta(name) + `\s*=\s*["']([^"']+)["']`)
	m := re.FindStringSubmatch(chunk)
	if len(m) > 1 {
		return cleanHTMLText(m[1])
	}
	return ""
}

func imagesFromChunkWithBase(chunk, base string, max int) []string {
	out := imagesFromChunk(chunk, max)
	if len(out) >= max || strings.TrimSpace(base) == "" {
		return out
	}
	seen := map[string]bool{}
	for _, item := range out {
		seen[item] = true
	}
	re := regexp.MustCompile(`(?is)<img[^>]+(?:src|data-src)\s*=\s*["']([^"']+)["']`)
	for _, m := range re.FindAllStringSubmatch(chunk, -1) {
		if len(m) < 2 {
			continue
		}
		u := absoluteURL(base, html.UnescapeString(strings.TrimSpace(m[1])))
		if !strings.HasPrefix(strings.ToLower(u), "https://") || strings.Contains(strings.ToLower(u), "sprite") || seen[u] {
			continue
		}
		seen[u] = true
		out = append(out, u)
		if len(out) >= max {
			break
		}
	}
	return out
}

func imageFromChunk(chunk string) string {
	images := imagesFromChunk(chunk, 1)
	if len(images) == 0 {
		return ""
	}
	return images[0]
}

func imagesFromChunk(chunk string, max int) []string {
	if max <= 0 {
		max = 8
	}
	seen := map[string]bool{}
	out := []string{}
	// Prefer OpenGraph/Twitter media when the nearby card/page exposes it.
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?is)<meta[^>]+(?:property|name)=["'](?:og:image|twitter:image)["'][^>]+content=["']([^"']+)["']`),
		regexp.MustCompile(`(?is)<meta[^>]+content=["']([^"']+)["'][^>]+(?:property|name)=["'](?:og:image|twitter:image)["']`),
		regexp.MustCompile(`(?is)<img[^>]+(?:src|data-src)\s*=\s*["']([^"']+)["']`),
	}
	for _, re := range patterns {
		for _, m := range re.FindAllStringSubmatch(chunk, -1) {
			if len(m) < 2 {
				continue
			}
			u := html.UnescapeString(strings.TrimSpace(m[1]))
			if strings.HasPrefix(u, "//") {
				u = "https:" + u
			}
			if !strings.HasPrefix(u, "https://") || strings.Contains(strings.ToLower(u), "sprite") || seen[u] {
				continue
			}
			seen[u] = true
			out = append(out, u)
			if len(out) >= max {
				return out
			}
		}
	}
	return out
}

func matchText(chunk, pattern string) string {
	re := regexp.MustCompile(pattern)
	m := re.FindStringSubmatch(chunk)
	if len(m) > 1 {
		return cleanHTMLText(m[1])
	}
	return ""
}

func parseCompactNumber(s string) int64 {
	s = strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(s), ",", ""))
	mult := float64(1)
	if strings.HasSuffix(s, "K") {
		mult, s = 1e3, strings.TrimSuffix(s, "K")
	} else if strings.HasSuffix(s, "M") {
		mult, s = 1e6, strings.TrimSuffix(s, "M")
	} else if strings.HasSuffix(s, "B") {
		mult, s = 1e9, strings.TrimSuffix(s, "B")
	}
	v, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return int64(v * mult)
}

func firstNonEmpty(xs ...string) string {
	for _, x := range xs {
		if strings.TrimSpace(x) != "" {
			return x
		}
	}
	return ""
}

func uniqueStringsPreserve(in []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		k := strings.ToLower(strings.TrimSpace(s))
		if k != "" && !seen[k] {
			seen[k] = true
			out = append(out, s)
		}
	}
	return out
}

func retryableProviderStatus(status int) bool {
	switch status {
	case http.StatusTooManyRequests, http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func providerRetryDelay(res *http.Response, attempt int) time.Duration {
	if res != nil {
		if raw := strings.TrimSpace(res.Header.Get("Retry-After")); raw != "" {
			if seconds, err := strconv.Atoi(raw); err == nil && seconds >= 0 {
				d := time.Duration(seconds) * time.Second
				if d > 2*time.Second {
					d = 2 * time.Second
				}
				return d
			}
		}
	}
	if attempt <= 0 {
		return 200 * time.Millisecond
	}
	return time.Duration(attempt+1) * 350 * time.Millisecond
}

func waitProviderRetry(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}
	t := time.NewTimer(delay)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func (a *App) getJSON(ctx context.Context, target string, headers map[string]string, dst any) error {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
		if err != nil {
			return err
		}
		req.Header.Set("User-Agent", userAgent)
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		res, err := a.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if attempt < 2 {
				if waitErr := waitProviderRetry(ctx, providerRetryDelay(nil, attempt)); waitErr != nil {
					return waitErr
				}
				continue
			}
			return err
		}
		if res.StatusCode < 200 || res.StatusCode >= 300 {
			b, _ := io.ReadAll(io.LimitReader(res.Body, 8<<10))
			lastErr = fmt.Errorf("%s returned %s: %s", hostLabel(target), res.Status, strings.TrimSpace(string(b)))
			retry := attempt < 2 && retryableProviderStatus(res.StatusCode)
			delay := providerRetryDelay(res, attempt)
			_ = res.Body.Close()
			if retry {
				if waitErr := waitProviderRetry(ctx, delay); waitErr != nil {
					return waitErr
				}
				continue
			}
			return lastErr
		}
		err = json.NewDecoder(io.LimitReader(res.Body, 8<<20)).Decode(dst)
		_ = res.Body.Close()
		return err
	}
	return lastErr
}

func (a *App) postJSON(ctx context.Context, target string, headers map[string]string, input, dst any) error {
	b, err := json.Marshal(input)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res, err := a.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 16<<10))
		return fmt.Errorf("%s returned %s: %s", hostLabel(target), res.Status, strings.TrimSpace(string(body)))
	}
	if dst == nil {
		return nil
	}
	return json.NewDecoder(io.LimitReader(res.Body, 16<<20)).Decode(dst)
}

func (a *App) getText(ctx context.Context, target string, headers map[string]string) (string, error) {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
		if err != nil {
			return "", err
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/126 Safari/537.36 MinecraftModVault/"+appVersion)
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		res, err := a.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if attempt < 2 {
				if waitErr := waitProviderRetry(ctx, providerRetryDelay(nil, attempt)); waitErr != nil {
					return "", waitErr
				}
				continue
			}
			return "", err
		}
		if res.StatusCode < 200 || res.StatusCode >= 300 {
			lastErr = fmt.Errorf("%s returned %s", hostLabel(target), res.Status)
			retry := attempt < 2 && retryableProviderStatus(res.StatusCode)
			delay := providerRetryDelay(res, attempt)
			_ = res.Body.Close()
			if retry {
				if waitErr := waitProviderRetry(ctx, delay); waitErr != nil {
					return "", waitErr
				}
				continue
			}
			return "", lastErr
		}
		b, err := io.ReadAll(io.LimitReader(res.Body, 8<<20))
		_ = res.Body.Close()
		return string(b), err
	}
	return "", lastErr
}

func hostLabel(target string) string {
	u, _ := url.Parse(target)
	if u != nil && u.Host != "" {
		return u.Host
	}
	return "provider"
}

func (a *App) handleProviderInstall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var in ProviderInstallRequest
	if err := decodeJSON(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	result, err := a.installProviderRequest(r.Context(), in)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, APIError{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (a *App) installProviderRequest(ctx context.Context, in ProviderInstallRequest) (any, error) {
	a.mu.RLock()
	if in.GameVersion == "" {
		in.GameVersion = a.settings.GameVersion
	}
	if in.Loader == "" {
		in.Loader = a.settings.Loader
	}
	a.mu.RUnlock()
	provider := strings.ToLower(strings.TrimSpace(in.Provider))
	var result any
	var err error
	switch provider {
	case "modrinth":
		installed := []InstalledProject{}
		err = a.installModrinthProject(ctx, firstNonEmpty(in.ID, in.Slug), "", in.GameVersion, in.Loader, in.Target, false, map[string]bool{}, &installed)
		result = map[string]any{"ok": err == nil, "installed": installed}
	case "github":
		result, err = a.installGitHubRelease(ctx, firstNonEmpty(in.ID, in.Slug), in.GameVersion, in.Loader, in.Target)
	case "curseforge":
		a.mu.RLock()
		hasCurseKey := strings.TrimSpace(a.settings.CurseForgeAPIKey) != ""
		a.mu.RUnlock()
		isBedrockPage := strings.Contains(strings.ToLower(in.PageURL), "/minecraft-bedrock/")
		// The native API path is strongest for Java mods. All other CurseForge
		// content lanes use the same verified package detector as the integrated
		// browser, which also handles Bedrock .mcaddon/.mcpack/.mcworld files.
		if in.PageURL != "" && (isBedrockPage || !hasCurseKey || (in.Target != "" && in.Target != "auto" && in.Target != "mods")) {
			result, err = a.installDetectedWebPackage(ctx, "curseforge", in.PageURL, firstNonEmpty(in.ID, in.Slug), in.Target)
		} else {
			result, err = a.installCurseForge(ctx, firstNonEmpty(in.ID, in.Slug), in.GameVersion, in.Loader, in.Target)
		}
	case "hangar":
		result, err = a.installHangar(ctx, firstNonEmpty(in.ID, in.Slug), in.GameVersion, in.Target)
	case "spigot":
		result, err = a.installSpigot(ctx, firstNonEmpty(in.ID, in.Slug), in.Target)
	case "builtbybit":
		result, err = a.installBuiltByBit(ctx, firstNonEmpty(in.ID, in.Slug), in.Target)
	case "smithed":
		result, err = a.installSmithed(ctx, firstNonEmpty(in.ID, in.Slug), in.GameVersion, in.Target)
	case "spongeore":
		result, err = a.installSpongeOre(ctx, firstNonEmpty(in.ID, in.Slug), in.Target)
	case "planetminecraft", "mcpedl", "bukkitdev", "moddb", "minecraftmaps", "resourcepacknet", "texturepacks", "mcreator", "shaderpackscom", "shaderpacksnet", "minecraftshader":
		result, err = a.installDetectedWebPackage(ctx, provider, in.PageURL, firstNonEmpty(in.ID, in.Slug), in.Target)
	case "skindex":
		result, err = a.installSkindexSkin(ctx, in.PageURL, firstNonEmpty(in.ID, in.Slug))
	case "minecrafthub":
		result, err = a.installMinecraftHubResolved(ctx, in.PageURL, firstNonEmpty(in.ID, in.Slug), in.GameVersion, in.Loader, in.Target)
	default:
		err = errors.New("this provider can be browsed fully in-app, but it does not expose a trustworthy direct package install route for this item")
	}
	return result, err
}

func githubReleaseAssetKind(target string) (extensions []string, defaultTarget string) {
	switch strings.ToLower(strings.TrimSpace(target)) {
	case "plugins":
		return []string{".jar"}, "plugins"
	case "resourcepacks":
		return []string{".zip"}, "resourcepacks"
	case "shaderpacks":
		return []string{".zip"}, "shaderpacks"
	case "datapacks":
		return []string{".zip"}, "datapacks"
	case "mods", "", "auto":
		return []string{".jar"}, "mods"
	default:
		return []string{".jar", ".zip"}, target
	}
}

func (a *App) installGitHubRelease(ctx context.Context, repo, game, loader, target string) (map[string]any, error) {
	repo = strings.Trim(strings.TrimPrefix(repo, "https://github.com/"), "/")
	parts := strings.Split(repo, "/")
	if len(parts) < 2 {
		return nil, errors.New("GitHub repository must be owner/name")
	}
	repo = parts[0] + "/" + parts[1]
	a.mu.RLock()
	token := strings.TrimSpace(a.settings.GitHubToken)
	a.mu.RUnlock()
	headers := map[string]string{"Accept": "application/vnd.github+json", "X-GitHub-Api-Version": "2022-11-28"}
	if token != "" {
		headers["Authorization"] = "Bearer " + token
	}
	var rel struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
			Size               int64  `json:"size"`
			Digest             string `json:"digest"`
		} `json:"assets"`
	}
	apiBase := providerBase("MMV_GITHUB_API_BASE", "https://api.github.com")
	if err := a.getJSON(ctx, apiBase+"/repos/"+repo+"/releases/latest", headers, &rel); err != nil {
		return nil, err
	}
	extensions, defaultTarget := githubReleaseAssetKind(target)
	if target == "" || target == "auto" {
		target = defaultTarget
	}
	extAllowed := func(name string) bool {
		for _, ext := range extensions {
			if strings.HasSuffix(name, ext) {
				return true
			}
		}
		return false
	}
	best := -1
	bestScore := -999
	for i, asset := range rel.Assets {
		name := strings.ToLower(asset.Name)
		if !extAllowed(name) || strings.Contains(name, "sources") || strings.Contains(name, "source-") || strings.Contains(name, "javadoc") || strings.Contains(name, "dev") || strings.Contains(name, "deobf") || strings.Contains(name, "api-") {
			continue
		}
		s := 0
		if game != "" && strings.Contains(name, strings.ToLower(game)) {
			s += 5
		}
		if loader != "" && loader != "any" && strings.Contains(name, strings.ToLower(loader)) {
			s += 4
		}
		if len(extensions) == 1 && strings.HasSuffix(name, extensions[0]) {
			s += 2
		}
		if strings.Contains(name, "release") || strings.Contains(name, "universal") {
			s++
		}
		if s > bestScore {
			best, bestScore = i, s
		}
	}
	if best < 0 {
		return nil, fmt.Errorf("latest GitHub release has no compatible %s asset for target %s", strings.Join(extensions, "/"), target)
	}
	asset := rel.Assets[best]
	dir := a.javaTargetDir(target)
	if err := osMkdirAll(dir); err != nil {
		return nil, err
	}
	dst := uniquePath(filepath.Join(dir, safeFilename(asset.Name)))
	hashes := map[string]string{}
	if strings.HasPrefix(asset.Digest, "sha256:") {
		hashes["sha256"] = strings.TrimPrefix(asset.Digest, "sha256:")
	}
	if err := a.downloadURLVerified(ctx, asset.BrowserDownloadURL, dst, asset.Size, hashes); err != nil {
		return nil, err
	}
	if err := validateZipContainer(dst); err != nil {
		_ = os.Remove(dst)
		return nil, fmt.Errorf("GitHub release asset was not a valid JAR/ZIP container: %w", err)
	}
	return map[string]any{"ok": true, "provider": "github", "repository": repo, "version": rel.TagName, "file": filepath.Base(dst), "path": dst, "target": target, "verification": "release asset size/hash checked when published and archive container validated"}, nil
}

func (a *App) resolveCurseForgeProjectID(ctx context.Context, key, idOrSlug string) (int64, string, error) {
	if id, err := strconv.ParseInt(strings.TrimSpace(idOrSlug), 10, 64); err == nil && id > 0 {
		return id, "", nil
	}
	slug := strings.TrimSpace(idOrSlug)
	if slug == "" {
		return 0, "", errors.New("CurseForge project id or slug is required")
	}
	v := url.Values{"gameId": {"432"}, "slug": {slug}, "pageSize": {"20"}}
	var resp struct {
		Data []struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
			Slug string `json:"slug"`
		} `json:"data"`
	}
	if err := a.getJSON(ctx, curseForgeAPIBase()+"/v1/mods/search?"+v.Encode(), map[string]string{"x-api-key": key}, &resp); err != nil {
		return 0, "", err
	}
	for _, project := range resp.Data {
		if strings.EqualFold(strings.TrimSpace(project.Slug), slug) && project.ID > 0 {
			return project.ID, project.Name, nil
		}
	}
	return 0, "", fmt.Errorf("could not resolve exact CurseForge project slug %q", slug)
}

func (a *App) installedCurseForgeProjectIDs(ctx context.Context, key, modsDir string) map[int64]bool {
	out := map[int64]bool{}
	locals, err := scanLocalModJars(modsDir)
	if err != nil || len(locals) == 0 {
		return out
	}
	fps := make([]uint32, 0, len(locals))
	for _, local := range locals {
		if local.CurseFingerprint != 0 {
			fps = append(fps, local.CurseFingerprint)
		}
	}
	if len(fps) == 0 {
		return out
	}
	var matches struct {
		Data struct {
			ExactMatches []struct {
				File struct {
					ModID int64 `json:"modId"`
				} `json:"file"`
			} `json:"exactMatches"`
		} `json:"data"`
	}
	if err := a.postJSON(ctx, curseForgeAPIBase()+"/v1/fingerprints", map[string]string{"x-api-key": key}, map[string]any{"fingerprints": fps}, &matches); err != nil {
		return out
	}
	for _, match := range matches.Data.ExactMatches {
		if match.File.ModID > 0 {
			out[match.File.ModID] = true
		}
	}
	return out
}

func (a *App) installCurseForgeProject(ctx context.Context, key string, modID int64, game, loader, target string, dependency bool, knownInstalled map[int64]bool, visiting map[int64]bool, installed *[]map[string]any) error {
	if modID <= 0 || knownInstalled[modID] {
		return nil
	}
	if visiting[modID] {
		return fmt.Errorf("CurseForge dependency cycle detected at project %d", modID)
	}
	visiting[modID] = true
	defer delete(visiting, modID)

	f, ok := a.fetchCurseForgeCompatibleFile(ctx, key, modID, game, loader)
	if !ok {
		return fmt.Errorf("CurseForge project %d has no file compatible with Minecraft %s / %s", modID, game, loader)
	}
	for _, depID := range curseForgeRequiredDependencies(f) {
		if err := a.installCurseForgeProject(ctx, key, depID, game, loader, "mods", true, knownInstalled, visiting, installed); err != nil {
			return fmt.Errorf("required CurseForge dependency %d for project %d failed: %w", depID, modID, err)
		}
	}
	if strings.TrimSpace(f.DownloadURL) == "" {
		return fmt.Errorf("CurseForge file %d does not expose a direct download URL", f.ID)
	}
	if target == "" || target == "auto" {
		target = "mods"
	}
	dir := a.javaTargetDir(target)
	if err := osMkdirAll(dir); err != nil {
		return err
	}
	dst := uniquePath(filepath.Join(dir, safeFilename(f.FileName)))
	if err := a.downloadURLVerified(ctx, f.DownloadURL, dst, f.FileLength, curseForgeFileHashes(f)); err != nil {
		return err
	}
	if err := validateZipContainer(dst); err != nil {
		_ = os.Remove(dst)
		return fmt.Errorf("CurseForge file %d was not a valid JAR/ZIP container: %w", f.ID, err)
	}
	knownInstalled[modID] = true
	*installed = append(*installed, map[string]any{
		"provider": "curseforge", "projectId": modID, "fileId": f.ID, "file": filepath.Base(dst), "path": dst, "target": target, "dependency": dependency,
	})
	return nil
}

func (a *App) installCurseForge(ctx context.Context, idOrSlug, game, loader, target string) (map[string]any, error) {
	a.mu.RLock()
	key := strings.TrimSpace(a.settings.CurseForgeAPIKey)
	a.mu.RUnlock()
	if key == "" {
		return nil, errors.New("CurseForge native installation needs a CurseForge API key in Settings; browsing remains integrated without one")
	}
	id, title, err := a.resolveCurseForgeProjectID(ctx, key, idOrSlug)
	if err != nil {
		return nil, err
	}
	if target == "" || target == "auto" {
		target = "mods"
	}
	knownInstalled := a.installedCurseForgeProjectIDs(ctx, key, a.javaTargetDir("mods"))
	installed := []map[string]any{}
	if err := a.installCurseForgeProject(ctx, key, id, game, loader, target, false, knownInstalled, map[int64]bool{}, &installed); err != nil {
		for i := len(installed) - 1; i >= 0; i-- {
			if path, _ := installed[i]["path"].(string); path != "" {
				_ = os.Remove(path)
			}
		}
		return nil, err
	}
	if len(installed) == 0 {
		return map[string]any{"ok": true, "provider": "curseforge", "projectId": id, "project": title, "alreadyInstalled": true, "installed": installed}, nil
	}
	main := installed[len(installed)-1]
	return map[string]any{
		"ok": true, "provider": "curseforge", "projectId": id, "project": title, "fileId": main["fileId"], "file": main["file"], "path": main["path"], "target": main["target"], "installed": installed,
	}, nil
}

func osMkdirAll(path string) error {
	return ensureDir(path)
}

func ensureDir(path string) error {
	return mkdirAll(path)
}
