package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestYouTubeChannelAPIEnumeratesEveryPlaylistPageAndClassifiesShorts(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/channels", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("forHandle"); got != "@AsianHalfSquat" {
			t.Fatalf("forHandle=%q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"items": []any{map[string]any{
			"id":             "UC_ARCHIVE",
			"snippet":        map[string]any{"title": "AsianHalfSquat", "thumbnails": map[string]any{"high": map[string]any{"url": "https://img/channel.jpg"}}},
			"contentDetails": map[string]any{"relatedPlaylists": map[string]any{"uploads": "UU_ARCHIVE"}},
			"statistics":     map[string]any{"videoCount": "3"},
		}}})
	})
	mux.HandleFunc("/playlistItems", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("playlistId") != "UU_ARCHIVE" || r.URL.Query().Get("maxResults") != "50" {
			t.Fatalf("unexpected playlist query: %s", r.URL.RawQuery)
		}
		if r.URL.Query().Get("pageToken") == "" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"nextPageToken": "page2", "pageInfo": map[string]any{"totalResults": 3},
				"items": []any{
					playlistFixture("v-new", "Newest mods", "latest recs", "2026-08-19T12:00:00Z"),
					playlistFixture("v-short", "Tiny mod #shorts", "short rec", "2026-08-18T12:00:00Z"),
				},
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"pageInfo": map[string]any{"totalResults": 3},
			"items":    []any{playlistFixture("v-old", "Oldest mods", "old recs", "2018-01-01T12:00:00Z")},
		})
	})
	mux.HandleFunc("/videos", func(w http.ResponseWriter, r *http.Request) {
		ids := strings.Split(r.URL.Query().Get("id"), ",")
		items := make([]any, 0, len(ids))
		for _, id := range ids {
			duration := "PT12M"
			if id == "v-short" {
				duration = "PT45S"
			}
			items = append(items, map[string]any{
				"id":             id,
				"snippet":        map[string]any{"title": id, "description": "description " + id, "publishedAt": "2026-08-19T12:00:00Z"},
				"contentDetails": map[string]any{"duration": duration, "caption": "true"},
			})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"items": items})
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	oldBase := youtubeDataAPIBase
	youtubeDataAPIBase = server.URL
	defer func() { youtubeDataAPIBase = oldBase }()

	app := &App{httpClient: server.Client()}
	ch, videos, err := app.enumerateYouTubeChannelAPI(context.Background(), "test-key", defaultCreatorChannels[0], true)
	if err != nil {
		t.Fatal(err)
	}
	if ch.ChannelID != "UC_ARCHIVE" || ch.UploadsPlaylist != "UU_ARCHIVE" || ch.TotalVideos != 3 {
		t.Fatalf("channel metadata incomplete: %#v", ch)
	}
	if len(videos) != 3 {
		t.Fatalf("expected every playlist page to be indexed, got %d videos: %#v", len(videos), videos)
	}
	seenShort := false
	for _, v := range videos {
		if v.ID == "v-short" {
			seenShort = v.VideoKind == "short" && v.DurationSeconds == 45 && v.CaptionHint
		}
	}
	if !seenShort {
		t.Fatalf("short metadata was not preserved: %#v", videos)
	}
}

func playlistFixture(id, title, description, published string) map[string]any {
	return map[string]any{
		"snippet": map[string]any{
			"title": title, "description": description, "channelTitle": "AsianHalfSquat",
			"resourceId": map[string]any{"videoId": id},
			"thumbnails": map[string]any{"high": map[string]any{"url": "https://img/" + id + ".jpg"}},
		},
		"contentDetails": map[string]any{"videoId": id, "videoPublishedAt": published},
		"status":         map[string]any{"privacyStatus": "public"},
	}
}

func TestCreatorArchiveHasNoHistoricalVideoCap(t *testing.T) {
	app := &App{}
	for i := 0; i < 325; i++ {
		app.upsertCreatorVideo(CreatorVideo{ID: fmt.Sprintf("video-%03d", i), Platform: "youtube", URL: fmt.Sprintf("https://youtube.test/watch?v=%03d", i), ChannelHandle: "@AsianHalfSquat"})
	}
	if len(app.creatorVideos) != 325 {
		t.Fatalf("historical archive was capped: got %d videos", len(app.creatorVideos))
	}
}

func TestCreatorProviderLinksCoverIntegratedSourceEcosystem(t *testing.T) {
	video := CreatorVideo{ID: "v", Platform: "youtube"}
	cases := map[string]string{
		"https://modrinth.com/mod/sodium":                         "modrinth",
		"https://www.curseforge.com/minecraft/mc-mods/create":     "curseforge",
		"https://github.com/FabricMC/fabric":                      "github",
		"https://hangar.papermc.io/PaperMC/Folia":                 "hangar",
		"https://www.spigotmc.org/resources/example.12345/":       "spigot",
		"https://dev.bukkit.org/projects/worldedit":               "bukkitdev",
		"https://builtbybit.com/resources/example.1234/":          "builtbybit",
		"https://polymart.org/resource/example.1234":              "polymart",
		"https://ore.spongepowered.org/Example":                   "spongeore",
		"https://smithed.dev/packs/example":                       "smithed",
		"https://www.moddb.com/mods/example":                      "moddb",
		"https://www.technicpack.net/modpack/example.123":         "technic",
		"https://www.feed-the-beast.com/modpacks/123-example":     "ftb",
		"https://www.nexusmods.com/minecraft/mods/123":            "nexusmods",
		"https://vanillatweaks.net/picker/resource-packs/":        "vanillatweaks",
		"https://www.minecraftmaps.com/survival-maps/example":     "minecraftmaps",
		"https://resourcepack.net/example-resource-pack/":         "resourcepacknet",
		"https://texture-packs.com/resourcepack/example/":         "texturepacks",
		"https://mcreator.net/modification/example":               "mcreator",
		"https://shaderpacks.com/example-shaders/":                "shaderpackscom",
		"https://shaderpacks.net/example-shaders/":                "shaderpacksnet",
		"https://minecraftshader.com/example-shaders/":            "minecraftshader",
		"https://www.minecraftskins.com/skin/123/example/":        "skindex",
		"https://minecrafthub.io/mod/example":                     "minecrafthub",
		"https://www.planetminecraft.com/mod/example/":            "planetminecraft",
		"https://mcpedl.com/example-addon/":                       "mcpedl",
		"https://www.minecraft.net/en-us/marketplace/pdp/example": "marketplace",
	}
	for raw, want := range cases {
		mod, ok := creatorModFromURL(raw, 12, video)
		if !ok || mod.Provider != want {
			t.Errorf("creatorModFromURL(%q) provider=%q ok=%v, want %q", raw, mod.Provider, ok, want)
		}
	}
}

func TestCreatorRecommendationSortFilterAndTranscriptPersistence(t *testing.T) {
	root := t.TempDir()
	app := &App{cfgDir: root, creatorVideos: []CreatorVideo{
		{ID: "new", Platform: "youtube", ChannelHandle: "@AsianHalfSquat", Creator: "AsianHalfSquat", CreatorURL: "https://youtube.test/@AsianHalfSquat", URL: "https://youtube.test/watch?v=new", Title: "New mods", PublishedAt: "2026-08-18T00:00:00Z", VideoKind: "video", AnalyzedAt: "2026-08-18T01:00:00Z", Mods: []CreatorMod{{Name: "Newest Mod", Provider: "modrinth", TimestampS: 25, Confidence: .95, ProjectSummary: "A modern mod"}}},
		{ID: "old", Platform: "youtube", ChannelHandle: "@EnderVerseMC", Creator: "EnderVerseMC", CreatorURL: "https://youtube.test/@EnderVerseMC", URL: "https://youtube.test/watch?v=old", Title: "Old short", PublishedAt: "2020-01-01T00:00:00Z", VideoKind: "short", AnalyzedAt: "2020-01-01T01:00:00Z", Mods: []CreatorMod{{Name: "Old Mod", Provider: "curseforge", TimestampS: 4, Confidence: .80, DescriptionContext: "Old Mod - a classic"}}},
	}}

	req := httptest.NewRequest(http.MethodGet, "/api/creators/recommendations?sort=recommended-asc&kind=short&channel=%40EnderVerseMC&q=classic", nil)
	rec := httptest.NewRecorder()
	app.handleCreatorRecommendations(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("recommendations status=%d body=%s", rec.Code, rec.Body.String())
	}
	var payload struct {
		Count           int                     `json:"count"`
		Recommendations []CreatorRecommendation `json:"recommendations"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Count != 1 || payload.Recommendations[0].Mod.Name != "Old Mod" {
		t.Fatalf("recommendation filters failed: %#v", payload)
	}

	segments := []TranscriptSegment{{StartMS: 1000, EndMS: 3000, Text: "First mod is Wakes"}, {StartMS: 4000, EndMS: 7000, Text: "It changes sleeping."}}
	if err := app.saveCreatorTranscript("old", "youtube-caption", segments); err != nil {
		t.Fatal(err)
	}
	loaded, err := app.loadCreatorTranscript("old")
	if err != nil {
		t.Fatal(err)
	}
	transcriptText := ""
	for _, seg := range loaded.Segments {
		transcriptText += " " + seg.Text
	}
	if loaded.VideoID != "old" || loaded.Source != "youtube-caption" || len(loaded.Segments) != 2 || !strings.Contains(transcriptText, "changes sleeping") {
		t.Fatalf("transcript persistence lost evidence: %#v", loaded)
	}
}

func TestCreatorTranscriptPatternsCaptureRecommendationLanguage(t *testing.T) {
	text := "First mod is Better Combat. Check out Ambient Sounds. Camera Overhaul improves third person movement. Finally Distant Horizons"
	got := strings.ToLower(strings.Join(transcriptModNames(text), " | "))
	for _, want := range []string{"better combat", "ambient sounds", "camera overhaul", "distant horizons"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q from %#v", want, transcriptModNames(text))
		}
	}
}

func TestCreatorRetryCooldown(t *testing.T) {
	if !creatorAnalysisRetryDue("", time.Hour) {
		t.Fatal("empty last-attempt should be immediately retryable")
	}
	if creatorAnalysisRetryDue(time.Now().UTC().Format(time.RFC3339), time.Hour) {
		t.Fatal("fresh failure should honor cooldown")
	}
	if !creatorAnalysisRetryDue(time.Now().Add(-2*time.Hour).UTC().Format(time.RFC3339), time.Hour) {
		t.Fatal("old failure should be retryable")
	}
}

func TestCanonicalCreatorHandleFromYouTubeURL(t *testing.T) {
	for _, raw := range []string{"@EnderVerseMC", "EnderVerseMC", "https://www.youtube.com/@EnderVerseMC", "https://youtube.com/@EnderVerseMC/shorts"} {
		if got := canonicalCreatorHandle(raw); got != "@EnderVerseMC" {
			t.Fatalf("canonicalCreatorHandle(%q)=%q", raw, got)
		}
	}
	if _, err := url.Parse("https://www.youtube.com/@EnderVerseMC"); err != nil {
		t.Fatal(err)
	}
}

func TestNormalizeCreatorChannelInputAcceptsQoLForms(t *testing.T) {
	cases := []struct {
		raw       string
		platform  string
		handle    string
		channelID string
		url       string
	}{
		{"@NoxusMods", "youtube", "@NoxusMods", "", "https://www.youtube.com/@NoxusMods"},
		{"NoxusMods", "youtube", "@NoxusMods", "", "https://www.youtube.com/@NoxusMods"},
		{"https://www.youtube.com/@NoxusMods/videos", "youtube", "@NoxusMods", "", "https://www.youtube.com/@NoxusMods"},
		{"UCYyRRWyMLSMCfKiIvfWYvkw", "youtube", "", "UCYyRRWyMLSMCfKiIvfWYvkw", "https://www.youtube.com/channel/UCYyRRWyMLSMCfKiIvfWYvkw"},
		{"https://youtube.com/channel/UCYyRRWyMLSMCfKiIvfWYvkw", "youtube", "", "UCYyRRWyMLSMCfKiIvfWYvkw", "https://www.youtube.com/channel/UCYyRRWyMLSMCfKiIvfWYvkw"},
		{"https://youtube.com/user/LegacyCreator", "youtube", "", "", "https://www.youtube.com/user/LegacyCreator"},
		{"youtube.com/c/LegacyCreator", "youtube", "", "", "https://www.youtube.com/c/LegacyCreator"},
		{"https://www.tiktok.com/@kizamiringo?is_from_webapp=1&sender_device=pc", "tiktok", "@kizamiringo", "", "https://www.tiktok.com/@kizamiringo"},
		{"https://www.tiktok.com/@its_katsumi", "tiktok", "@its_katsumi", "", "https://www.tiktok.com/@its_katsumi"},
		{"https://www.tiktok.com/@speedychunks?lang=en", "tiktok", "@speedychunks", "", "https://www.tiktok.com/@speedychunks"},
		{"tiktok:@curseforge", "tiktok", "@curseforge", "", "https://www.tiktok.com/@curseforge"},
	}
	for _, tc := range cases {
		got, err := normalizeCreatorChannelInput(tc.raw)
		if err != nil {
			t.Fatalf("normalizeCreatorChannelInput(%q): %v", tc.raw, err)
		}
		if got.Platform != tc.platform || got.Handle != tc.handle || got.ChannelID != tc.channelID || got.URL != tc.url {
			t.Errorf("normalizeCreatorChannelInput(%q)=%#v, want platform=%q handle=%q channelID=%q url=%q", tc.raw, got, tc.platform, tc.handle, tc.channelID, tc.url)
		}
	}
	for _, bad := range []string{"", "https://example.com/@NoxusMods", "https://youtube.com/watch?v=abc", "https://youtube.com/shorts/abc", "https://www.tiktok.com/@kizamiringo/video/12345"} {
		if _, err := normalizeCreatorChannelInput(bad); err == nil {
			t.Errorf("normalizeCreatorChannelInput(%q) unexpectedly succeeded", bad)
		}
	}
}

func TestCreatorChannelIdentityIsPlatformScoped(t *testing.T) {
	youtube := CreatorChannel{Platform: "youtube", Handle: "@same", URL: "https://www.youtube.com/@same"}
	tiktok := CreatorChannel{Platform: "tiktok", Handle: "@same", URL: "https://www.tiktok.com/@same"}
	if creatorChannelsEquivalent(youtube, tiktok) {
		t.Fatal("same handle on YouTube and TikTok must not collapse into one creator")
	}
	if creatorChannelKey(youtube) == creatorChannelKey(tiktok) {
		t.Fatalf("platform-scoped keys collided: %q", creatorChannelKey(youtube))
	}
}

func TestMergeTranscriptSegmentsKeepsSpeechAndVisualEvidence(t *testing.T) {
	speech := []TranscriptSegment{{StartMS: 0, EndMS: 1200, Text: "This one adds better combat"}}
	visual := []TranscriptSegment{{StartMS: 2000, EndMS: 4000, Text: "MOD: Better Combat\nMOD: Combat Roll"}, {StartMS: 4000, EndMS: 6000, Text: "MOD: Better Combat\nMOD: Combat Roll"}}
	got := mergeTranscriptSegments(speech, visual)
	if len(got) != 2 {
		t.Fatalf("merged segments=%d want 2: %#v", len(got), got)
	}
	if !strings.Contains(got[1].Text, "Combat Roll") {
		t.Fatalf("visual evidence missing: %#v", got)
	}
}

func TestDefaultCreatorChannelsIncludeFullCuratedCorpus(t *testing.T) {
	app := &App{cfgDir: t.TempDir(), creatorSyncRunning: map[string]bool{}}
	app.ensureDefaultCreatorChannels()
	if len(app.creatorChannels) != 18 {
		t.Fatalf("default creator count=%d want 18: %#v", len(app.creatorChannels), app.creatorChannels)
	}
	wanted := map[string]bool{
		"AsianHalfSquat": false, "EnderVerseMC": false, "Kizamiringo": false, "Katsumi": false, "SpeedyChunks": false, "CurseForge": false, "HendyVideos": false, "Noxus": false, "ChosenArchitect": false,
		"direwolf20": false, "Gaming On Caffeine": false, "SystemCollapse": false, "Lashmak": false,
		"PwrDown": false, "Mischief of Mice": false, "PopularMMOs": false, "DanTDM": false, "The Breakdown": false,
	}
	for _, ch := range app.creatorChannels {
		if _, ok := wanted[ch.Title]; ok {
			wanted[ch.Title] = ch.Required && !ch.Paused
		}
	}
	for title, ok := range wanted {
		if !ok {
			t.Errorf("built-in creator %q missing or not active/required", title)
		}
	}
}

func TestDefaultCreatorMigrationPreservesStateAndUpgradesRecommended(t *testing.T) {
	app := &App{cfgDir: t.TempDir(), creatorSyncRunning: map[string]bool{}, creatorChannels: []CreatorChannel{
		{Platform: "youtube", Handle: "@NoxusMods", URL: "https://www.youtube.com/@NoxusMods", Title: "Noxus", Source: "recommended", Paused: true, IndexedVideos: 42, Recommendations: 99},
	}}
	app.ensureDefaultCreatorChannels()
	count := 0
	for _, ch := range app.creatorChannels {
		if strings.EqualFold(ch.Handle, "@NoxusMods") {
			count++
			if !ch.Required || ch.Source != "curated-core" || !ch.Paused || ch.IndexedVideos != 42 || ch.Recommendations != 99 {
				t.Fatalf("migration lost state or metadata: %#v", ch)
			}
		}
	}
	if count != 1 {
		t.Fatalf("Noxus duplicate count=%d want 1", count)
	}
	if len(app.creatorChannels) != 18 {
		t.Fatalf("migrated creator count=%d want 18", len(app.creatorChannels))
	}
}

func TestCreatorSyncSlotsBoundFreshBackfill(t *testing.T) {
	app := &App{creatorSyncRunning: map[string]bool{"a": true, "b": true}}
	if got := app.creatorSyncSlots(4); got != 2 {
		t.Fatalf("slots=%d want 2", got)
	}
	if got := app.creatorSyncSlots(2); got != 0 {
		t.Fatalf("slots=%d want 0", got)
	}
	if got := app.creatorSyncSlots(99); got != 2 {
		t.Fatalf("clamped slots=%d want 2", got)
	}
}

func TestCreatorChannelManageAPIAddPausePersistAndUnfollow(t *testing.T) {
	cfg := t.TempDir()
	app := &App{cfgDir: cfg, creatorSyncRunning: map[string]bool{}}
	app.ensureDefaultCreatorChannels()

	postBody := `{"url":"@ExampleCreator","sync":false,"analyze":false}`
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/creators/channels", strings.NewReader(postBody))
	app.handleCreatorChannels(rr, req)
	if rr.Code != http.StatusAccepted {
		t.Fatalf("POST status=%d body=%s", rr.Code, rr.Body.String())
	}
	var post map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &post); err != nil {
		t.Fatal(err)
	}
	added, _ := post["added"].([]any)
	if len(added) != 1 {
		t.Fatalf("expected one added creator: %s", rr.Body.String())
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPatch, "/api/creators/channels", strings.NewReader(`{"url":"@ExampleCreator","paused":true}`))
	app.handleCreatorChannels(rr, req)
	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), `"paused":true`) {
		t.Fatalf("PATCH pause status=%d body=%s", rr.Code, rr.Body.String())
	}

	reloaded := &App{cfgDir: cfg, creatorSyncRunning: map[string]bool{}}
	reloaded.loadPersistentCaches()
	foundPaused := false
	for _, ch := range reloaded.creatorChannels {
		if strings.EqualFold(ch.Handle, "@ExampleCreator") {
			foundPaused = ch.Paused && ch.Source == "custom" && ch.AddedAt != ""
		}
	}
	if !foundPaused {
		t.Fatalf("followed creator did not persist with metadata: %#v", reloaded.creatorChannels)
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/creators/channels?url="+url.QueryEscape("@ExampleCreator"), nil)
	reloaded.handleCreatorChannels(rr, req)
	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), `"preservedArchive":true`) {
		t.Fatalf("DELETE custom status=%d body=%s", rr.Code, rr.Body.String())
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/creators/channels?url="+url.QueryEscape("@AsianHalfSquat"), nil)
	reloaded.handleCreatorChannels(rr, req)
	if rr.Code != http.StatusConflict {
		t.Fatalf("DELETE core status=%d body=%s", rr.Code, rr.Body.String())
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/creators/channels?url="+url.QueryEscape("@NoxusMods"), nil)
	reloaded.handleCreatorChannels(rr, req)
	if rr.Code != http.StatusConflict {
		t.Fatalf("DELETE curated core status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestCreatorChannelsGETIncludesRecommendationsAndTrackedState(t *testing.T) {
	app := &App{cfgDir: t.TempDir(), creatorSyncRunning: map[string]bool{}}
	app.ensureDefaultCreatorChannels()
	if _, _, err := app.addCreatorChannel("@NoxusMods"); err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/creators/channels", nil)
	app.handleCreatorChannels(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET status=%d body=%s", rr.Code, rr.Body.String())
	}
	var payload struct {
		Channels    []CreatorChannel           `json:"channels"`
		Suggestions []CreatorChannelSuggestion `json:"suggestions"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Suggestions) < 10 {
		t.Fatalf("expected a rich recommended creator catalog, got %d", len(payload.Suggestions))
	}
	var noxus bool
	for _, s := range payload.Suggestions {
		if s.Title == "Noxus" {
			noxus = s.Tracked && s.Priority == 100
		}
	}
	if !noxus {
		t.Fatalf("Noxus recommendation did not reflect followed state: %#v", payload.Suggestions)
	}
}

func TestYouTubeChannelAPIEnumeratesByChannelID(t *testing.T) {
	const channelID = "UCYyRRWyMLSMCfKiIvfWYvkw"
	mux := http.NewServeMux()
	mux.HandleFunc("/channels", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("id"); got != channelID {
			t.Fatalf("id=%q want %q", got, channelID)
		}
		if got := r.URL.Query().Get("forHandle"); got != "" {
			t.Fatalf("unexpected forHandle=%q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"items": []any{map[string]any{
			"id": channelID, "snippet": map[string]any{"title": "PwrDown"},
			"contentDetails": map[string]any{"relatedPlaylists": map[string]any{"uploads": "UU_PWRDOWN"}},
			"statistics":     map[string]any{"videoCount": "1"},
		}}})
	})
	mux.HandleFunc("/playlistItems", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"pageInfo": map[string]any{"totalResults": 1}, "items": []any{playlistFixture("pwr-1", "Mod list", "mods", "2021-01-01T00:00:00Z")}})
	})
	mux.HandleFunc("/videos", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"items": []any{map[string]any{"id": "pwr-1", "snippet": map[string]any{"title": "Mod list"}, "contentDetails": map[string]any{"duration": "PT5M", "caption": "true"}}}})
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	oldBase := youtubeDataAPIBase
	youtubeDataAPIBase = server.URL
	defer func() { youtubeDataAPIBase = oldBase }()

	app := &App{httpClient: server.Client()}
	ch, videos, err := app.enumerateYouTubeChannelAPI(context.Background(), "test-key", CreatorChannel{Platform: "youtube", ChannelID: channelID, URL: "https://www.youtube.com/channel/" + channelID}, true)
	if err != nil {
		t.Fatal(err)
	}
	if ch.Title != "PwrDown" || ch.UploadsPlaylist != "UU_PWRDOWN" || len(videos) != 1 || videos[0].ChannelID != channelID {
		t.Fatalf("channel-ID enumeration incomplete: ch=%#v videos=%#v", ch, videos)
	}
}

func TestBulkCreatorSyncSkipsPausedChannels(t *testing.T) {
	app := &App{cfgDir: t.TempDir(), creatorSyncRunning: map[string]bool{}, creatorChannels: []CreatorChannel{{Platform: "youtube", Handle: "@PausedCreator", URL: "https://www.youtube.com/@PausedCreator", Paused: true}}}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/creators/channels/sync", strings.NewReader(`{"full":true,"analyze":true}`))
	app.handleCreatorChannelSync(rr, req)
	if rr.Code != http.StatusAccepted {
		t.Fatalf("bulk sync status=%d body=%s", rr.Code, rr.Body.String())
	}
	var payload struct {
		Started   int `json:"started"`
		Requested int `json:"requested"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Started != 0 || payload.Requested != 0 {
		t.Fatalf("paused creator was included in bulk sync: %#v", payload)
	}
}

func TestCreatorArchiveRefreshUsesFullBackfillThenIncrementalUpdates(t *testing.T) {
	now := time.Date(2026, time.August, 23, 18, 0, 0, 0, time.UTC)
	if due, full := creatorChannelRefreshDecision(CreatorChannel{}, 60, now); !due || !full {
		t.Fatalf("fresh creator refresh decision=(due=%v full=%v), want full backfill", due, full)
	}
	tracked := CreatorChannel{LastFullSyncAt: now.Add(-24 * time.Hour).Format(time.RFC3339), LastSyncedAt: now.Add(-2 * time.Hour).Format(time.RFC3339)}
	if due, full := creatorChannelRefreshDecision(tracked, 60, now); !due || full {
		t.Fatalf("stale tracked creator decision=(due=%v full=%v), want incremental refresh", due, full)
	}
	tracked.LastSyncedAt = now.Add(-10 * time.Minute).Format(time.RFC3339)
	if due, full := creatorChannelRefreshDecision(tracked, 60, now); due || full {
		t.Fatalf("recent creator decision=(due=%v full=%v), want no refresh", due, full)
	}

	failed := CreatorChannel{SyncStatus: "error", SyncFailures: 3, LastAttemptAt: now.Add(-5 * time.Minute).Format(time.RFC3339)}
	if due, full := creatorChannelRefreshDecision(failed, 60, now); due || !full {
		t.Fatalf("failed fresh creator decision=(due=%v full=%v), want full backfill deferred by retry backoff", due, full)
	}
	failed.LastAttemptAt = now.Add(-25 * time.Minute).Format(time.RFC3339)
	if due, full := creatorChannelRefreshDecision(failed, 60, now); !due || !full {
		t.Fatalf("expired retry backoff decision=(due=%v full=%v), want full retry", due, full)
	}
}

func TestVisualTextModNamesCapturesTextOnlyRecommendationSlides(t *testing.T) {
	got := visualTextModNames("5 Minecraft Mods You Need\n1. Better Combat\n2. Combat Roll\nMOD: EMI\n@kizamiringo")
	joined := strings.Join(got, "|")
	for _, want := range []string{"Better Combat", "Combat Roll", "EMI"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("visual candidates=%q missing %q", joined, want)
		}
	}
	if strings.Contains(strings.ToLower(joined), "minecraft mods you need") || strings.Contains(joined, "@kizamiringo") {
		t.Fatalf("visual candidates retained UI/header noise: %q", joined)
	}
}
