package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func resetCreatorCatalogRuntimeForTest() {
	creatorCatalogRuntime.Lock()
	creatorCatalogRuntime.digest = ""
	creatorCatalogRuntime.state = CreatorCatalogState{}
	creatorCatalogRuntime.Unlock()
}

func TestEmbeddedCreatorCatalogsSeedAsianHalfSquatAndTikTokFollows(t *testing.T) {
	resetCreatorCatalogRuntimeForTest()
	app := &App{cfgDir: t.TempDir(), creatorSyncRunning: map[string]bool{}}
	app.loadPersistentCaches()
	if len(app.creatorChannels) != 20 {
		t.Fatalf("creator count=%d want 20", len(app.creatorChannels))
	}
	wanted := map[string]bool{"@speedychunks": false, "@noxusminecraft": false, "@unyxyt": false}
	for _, ch := range app.creatorChannels {
		if _, ok := wanted[strings.ToLower(ch.Handle)]; ok {
			wanted[strings.ToLower(ch.Handle)] = ch.Platform == "tiktok" && ch.Required && !ch.Paused
		}
		if ch.Handle == "@AsianHalfSquat" && ch.TotalVideos < 349 {
			t.Fatalf("AsianHalfSquat totalVideos=%d want >=349", ch.TotalVideos)
		}
	}
	for handle, ok := range wanted {
		if !ok {
			t.Fatalf("required TikTok creator %s missing", handle)
		}
	}
	videos, recs := 0, 0
	for _, v := range app.creatorVideos {
		if v.ChannelHandle == "@AsianHalfSquat" {
			videos++
			recs += len(v.Mods)
		}
	}
	if videos < 11 {
		t.Fatalf("AsianHalfSquat catalog videos=%d want >=11", videos)
	}
	if recs < 90 {
		t.Fatalf("AsianHalfSquat catalog recommendations=%d want >=90", recs)
	}
}

func TestLocalCreatorCatalogHotDropAndMalformedPreservesLastGood(t *testing.T) {
	resetCreatorCatalogRuntimeForTest()
	cfg := t.TempDir()
	app := &App{cfgDir: cfg, creatorSyncRunning: map[string]bool{}}
	app.loadPersistentCaches()
	dir := filepath.Join(cfg, "creator-catalogs", "community")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "drop.json")
	write := func(videoID, modName string) {
		payload := CreatorCatalog{SchemaVersion: 1, ID: "hot-drop-test", UpdatedAt: "2026-08-24T17:00:00Z", Creator: CreatorChannel{Platform: "youtube", Handle: "@HotDrop", URL: "https://www.youtube.com/@HotDrop", Title: "HotDrop", Required: true}, Videos: []CreatorVideo{{ID: videoID, Platform: "youtube", URL: "https://www.youtube.com/watch?v=" + videoID, Title: videoID, Mods: []CreatorMod{{Name: modName, ProjectType: "mod", Evidence: "fixture", Confidence: .9}}}}}
		raw, _ := json.Marshal(payload)
		if err := os.WriteFile(path, raw, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("drop-1", "Drop Mod")
	state, err := app.reloadCreatorCatalogs(true)
	if err != nil {
		t.Fatal(err)
	}
	if state.Catalogs < 5 {
		t.Fatalf("catalog count=%d want >=5", state.Catalogs)
	}
	found := false
	for _, v := range app.creatorVideos {
		if v.ID == "drop-1" && len(v.Mods) == 1 && v.Mods[0].Name == "Drop Mod" {
			found = true
		}
	}
	if !found {
		t.Fatal("hot-drop video/recommendation was not ingested")
	}
	if err := os.WriteFile(path, []byte(`{"schemaVersion":1,`), 0o644); err != nil {
		t.Fatal(err)
	}
	state, err = app.reloadCreatorCatalogs(true)
	if err == nil || len(state.Errors) == 0 {
		t.Fatal("malformed catalog should surface a warning")
	}
	found = false
	for _, v := range app.creatorVideos {
		if v.ID == "drop-1" && len(v.Mods) > 0 {
			found = true
		}
	}
	if !found {
		t.Fatal("malformed reload wiped last-known-good catalog data")
	}
}

func TestCatalogNeverDowngradesVerifiedLiveRecommendation(t *testing.T) {
	resetCreatorCatalogRuntimeForTest()
	cfg := t.TempDir()
	app := &App{cfgDir: cfg, creatorSyncRunning: map[string]bool{}, creatorVideos: []CreatorVideo{{ID: "same", Platform: "youtube", URL: "https://www.youtube.com/watch?v=same", ChannelHandle: "@HotDrop", AnalyzedAt: "2026-08-24T00:00:00Z", Mods: []CreatorMod{{Name: "Verified Mod", Provider: "modrinth", ProjectID: "abc", URL: "https://modrinth.com/mod/verified", Evidence: "live provider", Confidence: 1}}}}}
	dir := filepath.Join(cfg, "creator-catalogs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	payload := CreatorCatalog{SchemaVersion: 1, ID: "do-not-downgrade", UpdatedAt: "2026-08-24T17:00:00Z", Creator: CreatorChannel{Platform: "youtube", Handle: "@HotDrop", URL: "https://www.youtube.com/@HotDrop"}, Videos: []CreatorVideo{{ID: "same", Platform: "youtube", Mods: []CreatorMod{{Name: "Verified Mod", Evidence: "catalog only", Confidence: .7}}}}}
	raw, _ := json.Marshal(payload)
	if err := os.WriteFile(filepath.Join(dir, "same.json"), raw, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := app.reloadCreatorCatalogs(true); err != nil {
		t.Fatal(err)
	}
	for _, v := range app.creatorVideos {
		if v.ID == "same" {
			m := v.Mods[0]
			if m.Provider != "modrinth" || m.ProjectID != "abc" || m.URL != "https://modrinth.com/mod/verified" || m.Confidence != 1 {
				t.Fatalf("verified live recommendation downgraded: %#v", m)
			}
			return
		}
	}
	t.Fatal("video disappeared")
}

func TestCreatorCatalogReloadAPI(t *testing.T) {
	resetCreatorCatalogRuntimeForTest()
	app := &App{cfgDir: t.TempDir(), creatorSyncRunning: map[string]bool{}}
	app.loadPersistentCaches()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/creators/catalogs/reload", nil)
	app.handleCreatorCatalogReload(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	var state CreatorCatalogState
	if err := json.Unmarshal(rr.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	if state.Catalogs < 4 || state.Videos < 11 || state.Recommendations < 90 {
		t.Fatalf("weak catalog state: %#v", state)
	}
}
