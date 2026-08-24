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
)

func TestCreatorProviderProfileAndModpackURLClassification(t *testing.T) {
	profiles := []struct {
		raw, provider, user string
	}{
		{"https://www.curseforge.com/members/asianhalfsquat/projects", "CurseForge", "asianhalfsquat"},
		{"https://modrinth.com/user/example", "Modrinth", "example"},
	}
	for _, tc := range profiles {
		got, ok := creatorProviderProfileFromURL(tc.raw)
		if !ok || got.Provider != tc.provider || got.Username != tc.user {
			t.Fatalf("profile parse %q = %#v ok=%v", tc.raw, got, ok)
		}
		if kind := creatorProfileLinkKind(tc.raw, ""); kind != "creator-profile" {
			t.Fatalf("profile kind %q = %q", tc.raw, kind)
		}
	}
	packs := []struct {
		raw, provider, slug string
	}{
		{"https://www.curseforge.com/minecraft/modpacks/fresh-and-smooth", "CurseForge", "fresh-and-smooth"},
		{"https://modrinth.com/modpack/my-pack", "Modrinth", "my-pack"},
	}
	for _, tc := range packs {
		got, ok := creatorModpackFromURL(tc.raw)
		if !ok || got.Provider != tc.provider || got.Slug != tc.slug {
			t.Fatalf("modpack parse %q = %#v ok=%v", tc.raw, got, ok)
		}
	}
}

func TestRefreshCreatorModpacksModrinthEnumeratesEveryPackAndRole(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v2/user/TestCreator/projects":
			json.NewEncoder(w).Encode([]map[string]any{
				{"id": "p1", "slug": "pack-one", "title": "Pack One", "description": "First", "project_type": "modpack", "downloads": 1200, "updated": "2026-08-20T00:00:00Z", "game_versions": []string{"1.21.1"}, "loaders": []string{"fabric"}},
				{"id": "p2", "slug": "pack-two", "title": "Pack Two", "description": "Second", "project_type": "modpack", "downloads": 500, "updated": "2026-08-21T00:00:00Z", "game_versions": []string{"1.20.1"}, "loaders": []string{"forge"}},
				{"id": "m1", "slug": "not-a-pack", "title": "Not Pack", "project_type": "mod"},
			})
		case "/v2/project/p1/members":
			fmt.Fprint(w, `[ {"role":"Owner","user":{"username":"TestCreator"}} ]`)
		case "/v2/project/p2/members":
			fmt.Fprint(w, `[ {"role":"Developer","user":{"username":"TestCreator"}} ]`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	t.Setenv("MMV_MODRINTH_API_BASE", server.URL+"/v2")
	app := &App{httpClient: server.Client()}
	ch := CreatorChannel{Platform: "youtube", Handle: "@TestCreator", Title: "TestCreator", URL: "https://www.youtube.com/@TestCreator", ProfileLinks: []CreatorProfileLink{{URL: "https://modrinth.com/user/TestCreator", Kind: "creator-profile", Provider: "Modrinth", EvidenceType: "bio"}}}
	updated, err := app.refreshCreatorModpacks(context.Background(), ch)
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.CreatorModpacks) != 2 || updated.CreatorModpacksStatus != "ready" {
		t.Fatalf("unexpected modpack library: status=%s packs=%#v", updated.CreatorModpacksStatus, updated.CreatorModpacks)
	}
	bySlug := map[string]CreatorReleasedModpack{}
	for _, p := range updated.CreatorModpacks {
		bySlug[p.Slug] = p
	}
	if bySlug["pack-one"].Relationship != "owner" || bySlug["pack-two"].Relationship != "member" {
		t.Fatalf("relationships not retained: %#v", bySlug)
	}
}

func TestRefreshCreatorModpacksCurseForgeAPIEnumeratesOwnerAndCollaboratorPacks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "test-key" {
			http.Error(w, "missing key", http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		q := r.URL.Query()
		if strings.EqualFold(q.Get("searchFilter"), "AsianHalfSquat") {
			fmt.Fprint(w, `{"data":[{"id":1,"name":"Seed","slug":"seed","authors":[{"id":77,"name":"AsianHalfSquat","url":"https://www.curseforge.com/members/asianhalfsquat"}]}],"pagination":{"totalCount":1}}`)
			return
		}
		if q.Get("authorId") == "77" {
			fmt.Fprint(w, `{"data":[
{"id":101,"name":"Satisfaction Guaranteed","slug":"satisfaction-guaranteed","summary":"Visual overhaul","downloadCount":1000,"dateModified":"2026-06-19T00:00:00Z","authors":[{"id":77,"name":"AsianHalfSquat"}]},
{"id":102,"name":"Shattered Ring","slug":"shattered-ring","summary":"Adventure","downloadCount":5000,"dateModified":"2024-08-08T00:00:00Z","authors":[{"id":77,"name":"AsianHalfSquat"}]},
{"id":103,"name":"AHS Zombie Apocalypse","slug":"asianhalfsquat-zombie-apocalypse-modpack","summary":"Collaboration","downloadCount":400,"dateModified":"2022-10-05T00:00:00Z","authors":[{"id":88,"name":"splatchoot"},{"id":77,"name":"AsianHalfSquat"}]}
],"pagination":{"totalCount":3}}`)
			return
		}
		if q.Get("primaryAuthorId") == "77" {
			fmt.Fprint(w, `{"data":[{"id":101,"name":"Satisfaction Guaranteed","slug":"satisfaction-guaranteed"},{"id":102,"name":"Shattered Ring","slug":"shattered-ring"}],"pagination":{"totalCount":2}}`)
			return
		}
		http.Error(w, "unexpected query: "+q.Encode(), http.StatusBadRequest)
	}))
	defer server.Close()
	t.Setenv("MMV_CURSEFORGE_API_BASE", server.URL)
	app := &App{httpClient: server.Client()}
	app.settings.CurseForgeAPIKey = "test-key"
	ch := CreatorChannel{Platform: "youtube", Handle: "@AsianHalfSquat", Title: "AsianHalfSquat", URL: "https://www.youtube.com/@AsianHalfSquat", ProfileLinks: []CreatorProfileLink{{URL: "https://www.curseforge.com/members/asianhalfsquat/projects", Kind: "creator-profile", Provider: "CurseForge", EvidenceType: "seed"}}}
	updated, err := app.refreshCreatorModpacks(context.Background(), ch)
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.CreatorModpacks) != 3 {
		t.Fatalf("expected every CurseForge profile pack, got %#v", updated.CreatorModpacks)
	}
	owners, members := 0, 0
	for _, p := range updated.CreatorModpacks {
		switch p.Relationship {
		case "owner":
			owners++
		case "member":
			members++
		}
	}
	if owners != 2 || members != 1 {
		t.Fatalf("owner/member split wrong: owners=%d members=%d packs=%#v", owners, members, updated.CreatorModpacks)
	}
}

func TestRefreshCreatorModpacksCurseForgePublicProfileFallbackListsAllPackLinks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/members/enderverse/projects" {
			http.NotFound(w, r)
			return
		}
		fmt.Fprint(w, `<html><body><h1>EnderVerse</h1><div>7 Projects</div>
<a href="/minecraft/modpacks/mc-dungeons-reforged">MC Dungeons - Reforged</a>
<a href="/minecraft/modpacks/breathful">Breathful - An Atmospheric Journey</a>
<a href="/minecraft/modpacks/fresh-and-smooth">Fresh & Smooth</a>
<a href="/minecraft/modpacks/realism-java-modpack">Into The Wilds</a>
<a href="/minecraft/modpacks/deja-vu-ultimate-vanilla-modpack">Deja Vu</a>
<a href="/minecraft/modpacks/ultimate-rpg-season-2">Ultimate RPG Season 2</a>
<a href="/minecraft/modpacks/ultimate-rpg-season-1">Ultimate RPG Season 1</a>
<a href="/minecraft/mc-mods/not-a-pack">Not a pack</a></body></html>`)
	}))
	defer server.Close()
	t.Setenv("MMV_CURSEFORGE_WEB_BASE", server.URL)
	app := &App{httpClient: server.Client()}
	ch := CreatorChannel{Platform: "youtube", Handle: "@EnderVerseMC", Title: "EnderVerseMC", URL: "https://www.youtube.com/@EnderVerseMC", ProfileLinks: []CreatorProfileLink{{URL: "https://www.curseforge.com/members/enderverse/projects", Kind: "creator-profile", Provider: "CurseForge", EvidenceType: "seed"}}}
	updated, err := app.refreshCreatorModpacks(context.Background(), ch)
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.CreatorModpacks) != 7 {
		t.Fatalf("expected all seven public profile modpacks, got %d: %#v", len(updated.CreatorModpacks), updated.CreatorModpacks)
	}
	for _, p := range updated.CreatorModpacks {
		if p.Provider != "CurseForge" || p.Relationship != "profile" {
			t.Fatalf("fallback provenance misleading: %#v", p)
		}
	}
}

func TestDirectCreatorLinkedModrinthPackPromotesVerifiedProfileAndEnumeratesRest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v2/project/first-pack":
			fmt.Fprint(w, `{"id":"p1","slug":"first-pack","title":"First Pack","description":"Direct","project_type":"modpack","downloads":10}`)
		case "/v2/project/p1/members", "/v2/project/p2/members":
			fmt.Fprint(w, `[ {"role":"Owner","user":{"username":"ProviderAlias"}} ]`)
		case "/v2/user/ProviderAlias/projects":
			fmt.Fprint(w, `[
{"id":"p1","slug":"first-pack","title":"First Pack","description":"Direct","project_type":"modpack","downloads":10},
{"id":"p2","slug":"second-pack","title":"Second Pack","description":"Also theirs","project_type":"modpack","downloads":20}
]`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	t.Setenv("MMV_MODRINTH_API_BASE", server.URL+"/v2")
	app := &App{httpClient: server.Client()}
	ch := CreatorChannel{Platform: "youtube", Handle: "@CreatorName", Title: "CreatorName", URL: "https://www.youtube.com/@CreatorName", ProfileLinks: []CreatorProfileLink{{URL: "https://modrinth.com/modpack/first-pack", Label: "My Modpack", Kind: "modpack", Provider: "Modrinth", EvidenceURL: "https://linktr.ee/CreatorName", EvidenceType: "link-hub"}}}
	updated, err := app.refreshCreatorModpacks(context.Background(), ch)
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.CreatorModpacks) != 2 {
		t.Fatalf("direct pack should unlock creator's full provider library: %#v", updated.CreatorModpacks)
	}
	foundProfile := false
	for _, link := range updated.ProfileLinks {
		if link.Kind == "creator-profile" && strings.EqualFold(link.URL, "https://modrinth.com/user/ProviderAlias") {
			foundProfile = true
		}
	}
	if !foundProfile {
		t.Fatalf("verified Modrinth creator profile was not promoted: %#v", updated.ProfileLinks)
	}
}

func TestDirectCreatorLinkedCurseForgePackPromotesSinglePublicAuthorWithoutAPIKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		switch r.URL.Path {
		case "/minecraft/modpacks/creator-pack":
			fmt.Fprint(w, `<html><body><a href="/members/ProviderAlias/projects">ProviderAlias</a></body></html>`)
		case "/members/ProviderAlias/projects":
			fmt.Fprint(w, `<html><body>
				<a href="/minecraft/modpacks/creator-pack">Creator Pack</a>
				<a href="/minecraft/modpacks/second-pack">Second Pack</a>
			</body></html>`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	t.Setenv("MMV_CURSEFORGE_WEB_BASE", server.URL)
	app := &App{httpClient: server.Client()}
	ch := CreatorChannel{
		Platform: "youtube",
		Handle:   "@CreatorName",
		Title:    "CreatorName",
		URL:      "https://www.youtube.com/@CreatorName",
		ProfileLinks: []CreatorProfileLink{{
			URL:          "https://www.curseforge.com/minecraft/modpacks/creator-pack",
			Label:        "My Modpack",
			Kind:         "modpack",
			Provider:     "CurseForge",
			EvidenceURL:  "https://linktr.ee/CreatorName",
			EvidenceType: "link-hub",
		}},
	}
	updated, err := app.refreshCreatorModpacks(context.Background(), ch)
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.CreatorModpacks) != 2 {
		t.Fatalf("single public CurseForge author should unlock all profile packs: %#v", updated.CreatorModpacks)
	}
	foundProfile := false
	for _, link := range updated.ProfileLinks {
		if link.Kind == "creator-profile" && strings.EqualFold(link.URL, "https://www.curseforge.com/members/ProviderAlias/projects") {
			foundProfile = true
		}
	}
	if !foundProfile {
		t.Fatalf("verified CurseForge creator profile was not promoted: %#v", updated.ProfileLinks)
	}
}

func TestCreatorModpackRefreshNeverFabricatesWithoutEvidence(t *testing.T) {
	app := &App{}
	ch := CreatorChannel{Platform: "youtube", Handle: "@SomeCreator", Title: "SomeCreator", URL: "https://www.youtube.com/@SomeCreator"}
	updated, err := app.refreshCreatorModpacks(context.Background(), ch)
	if err != nil {
		t.Fatal(err)
	}
	if updated.CreatorModpacksStatus != "none" || len(updated.CreatorModpacks) != 0 {
		t.Fatalf("unverified packs were fabricated: %#v", updated)
	}
}

func TestCreatorModpackRefreshPreservesLastKnownGoodWhenProviderFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "provider unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()
	t.Setenv("MMV_MODRINTH_API_BASE", server.URL+"/v2")
	app := &App{httpClient: server.Client()}
	old := CreatorReleasedModpack{Provider: "Modrinth", ProjectID: "old", Slug: "known", Title: "Known Pack", URL: "https://modrinth.com/modpack/known", Author: "CacheCreator", ProviderProfileURL: "https://modrinth.com/user/CacheCreator", Relationship: "owner", FirstSeenAt: "2026-01-01T00:00:00Z"}
	ch := CreatorChannel{Platform: "youtube", Handle: "@CacheCreator", Title: "CacheCreator", URL: "https://www.youtube.com/@CacheCreator", ProfileLinks: []CreatorProfileLink{{URL: "https://modrinth.com/user/CacheCreator", Kind: "creator-profile", Provider: "Modrinth"}}, CreatorModpacks: []CreatorReleasedModpack{old}}
	updated, err := app.refreshCreatorModpacks(context.Background(), ch)
	if err != nil {
		t.Fatalf("last-known-good cache should degrade gracefully: %v", err)
	}
	if updated.CreatorModpacksStatus != "partial" || len(updated.CreatorModpacks) != 1 || updated.CreatorModpacks[0].Slug != "known" {
		t.Fatalf("last-known-good creator pack lost: %#v", updated)
	}
}

func TestBuiltInAsianHalfSquatAndEnderVerseHaveVerifiedCurseForgeProfileSeeds(t *testing.T) {
	app := &App{cfgDir: t.TempDir(), creatorSyncRunning: map[string]bool{}}
	app.ensureDefaultCreatorChannels()
	want := map[string]string{"@AsianHalfSquat": "asianhalfsquat", "@EnderVerseMC": "enderverse"}
	for handle, username := range want {
		found := false
		for _, ch := range app.creatorChannels {
			if !strings.EqualFold(ch.Handle, handle) {
				continue
			}
			for _, link := range ch.ProfileLinks {
				profile, ok := creatorProviderProfileFromURL(link.URL)
				if ok && profile.Provider == "CurseForge" && strings.EqualFold(profile.Username, username) && link.EvidenceType == "seed" {
					found = true
				}
			}
		}
		if !found {
			t.Fatalf("%s missing verified CurseForge profile seed", handle)
		}
	}
}

func TestCreatorModpackRefreshEndpointPersistsAllPacks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/v2/user/PersistPacks/projects" {
			fmt.Fprint(w, `[{"id":"p1","slug":"one","title":"One","project_type":"modpack"},{"id":"p2","slug":"two","title":"Two","project_type":"modpack"}]`)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/members") {
			fmt.Fprint(w, `[ {"role":"Owner","user":{"username":"PersistPacks"}} ]`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()
	t.Setenv("MMV_MODRINTH_API_BASE", server.URL+"/v2")
	cfg := t.TempDir()
	app := &App{cfgDir: cfg, httpClient: server.Client(), creatorSyncRunning: map[string]bool{}, creatorChannels: []CreatorChannel{{Platform: "youtube", Handle: "@PersistPacks", Title: "PersistPacks", URL: "https://www.youtube.com/@PersistPacks", ProfileLinks: []CreatorProfileLink{{URL: "https://modrinth.com/user/PersistPacks", Kind: "creator-profile", Provider: "Modrinth"}}}}}
	if err := app.saveCreatorChannels(); err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/creators/channels/modpacks/refresh", strings.NewReader(`{"url":"https://www.youtube.com/@PersistPacks"}`))
	app.handleCreatorModpacksRefresh(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("modpack refresh status=%d body=%s", rr.Code, rr.Body.String())
	}
	var payload struct {
		Count int `json:"count"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil || payload.Count != 2 {
		t.Fatalf("bad refresh payload: %v %s", err, rr.Body.String())
	}
	reloaded := &App{cfgDir: cfg, creatorSyncRunning: map[string]bool{}}
	reloaded.loadPersistentCaches()
	found := false
	for _, ch := range reloaded.creatorChannels {
		if strings.EqualFold(ch.Handle, "@PersistPacks") {
			found = len(ch.CreatorModpacks) == 2
			break
		}
	}
	if !found {
		t.Fatalf("creator modpack library did not persist: %#v", reloaded.creatorChannels)
	}
}

func TestCurseForgeCanonicalProfileFetchURLUsesConfiguredWebBase(t *testing.T) {
	t.Setenv("MMV_CURSEFORGE_WEB_BASE", "http://127.0.0.1:12345")
	u := creatorCurseForgeFetchURL("https://www.curseforge.com/members/example/projects?page=2")
	parsed, err := url.Parse(u)
	if err != nil || parsed.Host != "127.0.0.1:12345" || parsed.Path != "/members/example/projects" || parsed.Query().Get("page") != "2" {
		t.Fatalf("configured CurseForge fetch URL mismatch: %q err=%v", u, err)
	}
}
