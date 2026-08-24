package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type recommendationCache struct {
	Updated string           `json:"updated"`
	Items   []UnifiedProject `json:"items"`
}

func (a *App) loadPersistentCaches() {
	a.dataMu.Lock()
	a.updatePlans = map[string]UpdatePlan{}
	var rc recommendationCache
	if b, err := os.ReadFile(filepath.Join(a.cfgDir, "recommendations.json")); err == nil && json.Unmarshal(b, &rc) == nil {
		a.recommendations = rc.Items
		a.recommendationsUpdated, _ = time.Parse(time.RFC3339, rc.Updated)
	}
	var cc struct {
		Videos []CreatorVideo `json:"videos"`
	}
	if b, err := os.ReadFile(filepath.Join(a.cfgDir, "creator-videos.json")); err == nil && json.Unmarshal(b, &cc) == nil {
		a.creatorVideos = cc.Videos
	}
	var channels struct {
		Channels []CreatorChannel `json:"channels"`
	}
	if b, err := os.ReadFile(filepath.Join(a.cfgDir, "creator-channels.json")); err == nil && json.Unmarshal(b, &channels) == nil {
		a.creatorChannels = channels.Channels
	}
	for i := range a.creatorChannels {
		if a.creatorChannels[i].Required && a.creatorChannels[i].Source == "" {
			a.creatorChannels[i].Source = "core"
		}
	}
	if a.creatorSyncRunning == nil {
		a.creatorSyncRunning = map[string]bool{}
	}
	changed := a.ensureDefaultCreatorChannelsLocked()
	a.refreshCreatorChannelStatsLocked()
	a.dataMu.Unlock()
	if changed {
		_ = a.saveCreatorChannels()
	}
	// Creator catalog bundles are a hot-drop data layer. Embedded release catalogs
	// and user-supplied config/creator-catalogs JSON are merged after the durable
	// caches so catalog evidence fills gaps without replacing stronger live data.
	_, _ = a.reloadCreatorCatalogs(true)
}

func (a *App) saveRecommendations() error {
	a.dataMu.RLock()
	payload := recommendationCache{Updated: a.recommendationsUpdated.UTC().Format(time.RFC3339), Items: append([]UnifiedProject(nil), a.recommendations...)}
	a.dataMu.RUnlock()
	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return writeFileAtomic(filepath.Join(a.cfgDir, "recommendations.json"), b, 0o644)
}

func (a *App) saveCreatorVideos() error {
	a.dataMu.RLock()
	payload := struct {
		Videos []CreatorVideo `json:"videos"`
	}{Videos: append([]CreatorVideo(nil), a.creatorVideos...)}
	a.dataMu.RUnlock()
	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return writeFileAtomic(filepath.Join(a.cfgDir, "creator-videos.json"), b, 0o644)
}

func (a *App) saveCreatorChannels() error {
	a.dataMu.RLock()
	payload := struct {
		Channels []CreatorChannel `json:"channels"`
	}{Channels: append([]CreatorChannel(nil), a.creatorChannels...)}
	a.dataMu.RUnlock()
	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return writeFileAtomic(filepath.Join(a.cfgDir, "creator-channels.json"), b, 0o644)
}

func writeFileAtomic(path string, data []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".mmv-atomic-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	cleanup := true
	defer func() {
		_ = temporary.Close()
		if cleanup {
			_ = os.Remove(temporaryPath)
		}
	}()
	if err := temporary.Chmod(mode); err != nil {
		return err
	}
	if _, err := temporary.Write(data); err != nil {
		return err
	}
	if err := temporary.Sync(); err != nil {
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return err
	}
	cleanup = false
	return nil
}

func (a *App) backgroundIntelligenceLoop() {
	// Warm the cache immediately if it is missing or stale, then keep it current.
	time.Sleep(2 * time.Second)
	for {
		// Cheap digest-gated rescan: dropping/replacing a catalog file while the
		// app is running becomes visible without a restart or rebuild.
		_, _ = a.reloadCreatorCatalogs(false)
		a.mu.RLock()
		refreshMinutes := a.settings.AutoRefreshMinutes
		creatorMinutes := a.settings.CreatorRefreshMinutes
		a.mu.RUnlock()
		if refreshMinutes < 5 {
			refreshMinutes = 5
		}
		if creatorMinutes < 30 {
			creatorMinutes = 30
		}
		a.dataMu.RLock()
		lastRecommendations := a.recommendationsUpdated
		lastCreator := latestCreatorRefresh(a.creatorVideos)
		a.dataMu.RUnlock()

		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
		if lastRecommendations.IsZero() || time.Since(lastRecommendations) >= time.Duration(refreshMinutes)*time.Minute {
			_ = a.refreshRecommendations(ctx)
		}
		if lastCreator.IsZero() || time.Since(lastCreator) >= time.Duration(creatorMinutes)*time.Minute {
			_ = a.refreshCreatorDiscovery(ctx)
		}
		// Queue draining is owned by backgroundCreatorArchiveLoop. Keeping it in one
		// place prevents the generic discovery loop from racing archival channel jobs.
		cancel()
		time.Sleep(time.Minute)
	}
}

func latestCreatorRefresh(videos []CreatorVideo) time.Time {
	var latest time.Time
	for _, v := range videos {
		if t, err := time.Parse(time.RFC3339, v.DiscoveredAt); err == nil && t.After(latest) {
			latest = t
		}
	}
	return latest
}

func (a *App) handleRecommendations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	a.dataMu.RLock()
	items := append([]UnifiedProject(nil), a.recommendations...)
	updated := a.recommendationsUpdated
	a.dataMu.RUnlock()
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "updated": updated.UTC().Format(time.RFC3339), "count": len(items), "dynamic": true})
}

func (a *App) handleRecommendationRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
	defer cancel()
	if err := a.refreshRecommendations(ctx); err != nil {
		writeJSON(w, http.StatusBadGateway, APIError{Error: err.Error()})
		return
	}
	a.dataMu.RLock()
	count := len(a.recommendations)
	updated := a.recommendationsUpdated
	a.dataMu.RUnlock()
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "count": count, "updated": updated.UTC().Format(time.RFC3339)})
}

func (a *App) refreshRecommendations(ctx context.Context) error {
	a.mu.RLock()
	s := a.settings
	a.mu.RUnlock()
	if len(s.PreferredTags) == 0 {
		return errors.New("no recommendation interests are configured")
	}

	// Live feeds deliberately rotate so the cache evolves instead of replaying a static list.
	// Every cycle contains broad popularity + freshness, a rotating slice of personal interests,
	// and rotating taxonomy deep-dives. Over successive refreshes the whole taste profile and
	// category graph are revisited without hammering providers with hundreds of requests.
	type feed struct {
		query       string
		category    string
		projectType string
		sort        string
		reason      string
	}
	feeds := []feed{
		{query: "minecraft mods", projectType: "mod", sort: "downloads", reason: "Highly used mods right now"},
		{query: "minecraft mods", projectType: "mod", sort: "updated", reason: "Freshly updated mods"},
		{query: "minecraft shaders", projectType: "shader", sort: "updated", reason: "Current shaders and lighting"},
		{query: "minecraft resource packs", projectType: "resourcepack", sort: "downloads", reason: "Popular visual packs"},
		{query: "minecraft data packs", projectType: "datapack", sort: "updated", reason: "Fresh data-pack gameplay"},
	}
	bucket := int(time.Now().Unix() / int64(maxInt(s.AutoRefreshMinutes, 5)*60))
	interestCount := minInt(5, len(s.PreferredTags))
	for i := 0; i < interestCount; i++ {
		tag := s.PreferredTags[(bucket+i)%len(s.PreferredTags)]
		feeds = append(feeds, feed{query: tag, sort: "relevance", reason: "Matches " + tag})
	}
	categories := universalTaxonomy[1:]
	for i := 0; i < 3 && i < len(categories); i++ {
		cat := categories[(bucket*3+i)%len(categories)]
		feeds = append(feeds, feed{category: cat.ID, sort: "downloads", reason: "Top in " + cat.Name})
	}

	creatorSignals := a.creatorRecommendationSignals()
	items := make([]UnifiedProject, 0, 260)
	var errs []string
	for _, f := range feeds {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		resp := a.searchProviders(ctx, providerSearchOptions{Query: f.query, Category: f.category, GameVersion: s.GameVersion, Loader: s.Loader, ProjectType: f.projectType, Limit: 24, Sort: f.sort, Sources: s.EnabledSources})
		for source, msg := range resp.Errors {
			errs = append(errs, source+": "+msg)
		}
		for _, project := range resp.Results {
			project.Reason = recommendationReason(project, f.reason, s)
			project.Score += preferenceBoost(project, s.PreferredTags)
			project.Score += creatorSignalBoost(project, creatorSignals)
			items = append(items, project)
		}
	}
	items = mergeProviderDuplicates(items)
	sort.SliceStable(items, func(i, j int) bool { return items[i].Score > items[j].Score })
	if len(items) > 180 {
		items = items[:180]
	}
	if len(items) == 0 {
		if len(errs) > 0 {
			return fmt.Errorf("recommendation refresh returned no projects: %s", strings.Join(uniqueStringsPreserve(errs), "; "))
		}
		return errors.New("recommendation refresh returned no projects")
	}
	a.dataMu.Lock()
	a.recommendations = items
	a.recommendationsUpdated = time.Now().UTC()
	a.dataMu.Unlock()
	return a.saveRecommendations()
}

func (a *App) creatorRecommendationSignals() []string {
	a.dataMu.RLock()
	defer a.dataMu.RUnlock()
	seen := map[string]bool{}
	out := []string{}
	for i := len(a.creatorVideos) - 1; i >= 0 && len(out) < 80; i-- {
		for _, mod := range a.creatorVideos[i].Mods {
			k := normalizeProjectTitle(mod.Name)
			if k != "" && !seen[k] {
				seen[k] = true
				out = append(out, k)
			}
		}
	}
	return out
}

func creatorSignalBoost(project UnifiedProject, signals []string) float64 {
	name := normalizeProjectTitle(project.Title)
	if name == "" {
		return 0
	}
	for _, signal := range signals {
		if name == signal {
			return 18
		}
		if len(signal) >= 5 && (strings.Contains(name, signal) || strings.Contains(signal, name)) {
			return 8
		}
	}
	return 0
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func recommendationReason(p UnifiedProject, lead string, s Settings) string {
	parts := []string{lead}
	if containsFold(p.Versions, s.GameVersion) {
		parts = append(parts, "compatible with "+s.GameVersion)
	}
	if containsFold(p.Loaders, s.Loader) {
		parts = append(parts, s.Loader+" build")
	}
	if p.Downloads > 1000000 {
		parts = append(parts, "widely used")
	}
	if t, err := time.Parse(time.RFC3339, p.DateUpdated); err == nil && time.Since(t) < 60*24*time.Hour {
		parts = append(parts, "recently updated")
	}
	return strings.Join(parts, " · ")
}

func preferenceBoost(p UnifiedProject, tags []string) float64 {
	hay := strings.ToLower(p.Title + " " + p.Summary + " " + strings.Join(p.Categories, " "))
	boost := 0.0
	for _, tag := range tags {
		for _, word := range strings.Fields(strings.ToLower(tag)) {
			if len(word) >= 3 && strings.Contains(hay, word) {
				boost += 2.5
			}
		}
	}
	return boost
}
