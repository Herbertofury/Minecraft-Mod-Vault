package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type creatorRoundTripFunc func(*http.Request) (*http.Response, error)

func (f creatorRoundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func creatorHTMLResponse(req *http.Request, status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}
}

func creatorJSONResponse(req *http.Request, status int, body string) *http.Response {
	res := creatorHTMLResponse(req, status, body)
	res.Header.Set("Content-Type", "application/json")
	return res
}

func TestCreatorProfileLinkClassification(t *testing.T) {
	cases := []struct {
		url, label, want string
	}{
		{"https://drive.google.com/file/d/abc/view", "My Minecraft Modpack", "modpack"},
		{"https://www.curseforge.com/minecraft/modpacks/all-the-mods-10", "", "modpack"},
		{"https://modrinth.com/mod/sodium", "", "mod"},
		{"https://modrinth.com/resourcepack/fresh-animations", "", "resource-pack"},
		{"https://linktr.ee/creator", "", "link-hub"},
		{"https://linktr.ee/creator", "My Modpacks", "link-hub"},
		{"https://www.curseforge.com/members/creator/projects", "My Modpacks", "creator-profile"},
		{"https://discord.gg/example", "Discord", "social"},
		{"https://example.com/tips", "TIPS/DONOS", "support"},
		{"https://www.amazon.com/hz/wishlist/ls/example", "Wishlist", "wishlist"},
	}
	for _, tc := range cases {
		if got := creatorProfileLinkKind(tc.url, tc.label); got != tc.want {
			t.Errorf("creatorProfileLinkKind(%q,%q)=%q want %q", tc.url, tc.label, got, tc.want)
		}
	}
}

func TestExtractCreatorRawLinksCapturesAnchorsEmbeddedJSONAndCleansTracking(t *testing.T) {
	html := `<html><body>
<a href="https://linktr.ee/TestCreator?utm_source=tiktok">My links</a>
<a href="https://drive.google.com/file/d/abc/view?utm_medium=bio">My Modpack</a>
<script>window.__data={"discord":"https:\/\/discord.gg\/abc"}</script>
</body></html>`
	got := extractCreatorRawLinks(html, "https://www.tiktok.com/@testcreator")
	byURL := map[string]creatorRawLink{}
	for _, link := range got {
		byURL[link.URL] = link
	}
	if _, ok := byURL["https://linktr.ee/TestCreator"]; !ok {
		t.Fatalf("clean Linktree URL missing: %#v", got)
	}
	if link, ok := byURL["https://drive.google.com/file/d/abc/view"]; !ok || link.Label != "My Modpack" {
		t.Fatalf("labeled modpack link missing/unclean: %#v", got)
	}
	if _, ok := byURL["https://discord.gg/abc"]; !ok {
		t.Fatalf("embedded escaped JSON URL missing: %#v", got)
	}
}

func TestRefreshCreatorProfileLinksFollowsProfilesAndHubsButNotOutboundTargets(t *testing.T) {
	requested := []string{}
	client := &http.Client{Transport: creatorRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		requested = append(requested, req.URL.Hostname())
		switch req.URL.Hostname() {
		case "www.tiktok.com":
			return creatorHTMLResponse(req, http.StatusOK, `<a href="https://linktr.ee/TestCreator">Link in bio</a>`), nil
		case "linktr.ee":
			return creatorHTMLResponse(req, http.StatusOK, `<a href="https://www.curseforge.com/minecraft/modpacks/example-pack">My Minecraft Modpack</a><a href="https://discord.gg/example">Discord Server</a><a href="https://www.twitch.tv/example">Twitch</a><a href="https://drive.google.com/file/d/pack/view">Backup Modpack Download</a>`), nil
		default:
			t.Fatalf("crawler fetched an outbound target instead of only profile/hub pages: %s", req.URL.String())
			return nil, nil
		}
	})}
	app := &App{httpClient: client}
	ch := CreatorChannel{
		Platform: "tiktok", Handle: "@testcreator", URL: "https://www.tiktok.com/@testcreator",
		ProfileLinks: []CreatorProfileLink{{URL: "https://old.example/stale", Label: "stale", Kind: "website", EvidenceType: "link-hub"}},
	}
	updated, err := app.refreshCreatorProfileLinks(context.Background(), ch)
	if err != nil {
		t.Fatal(err)
	}
	if updated.ProfileLinksStatus != "ready" || updated.ProfileLinksRefreshedAt == "" {
		t.Fatalf("profile refresh status incomplete: %#v", updated)
	}
	if len(requested) != 2 || requested[0] != "www.tiktok.com" || requested[1] != "linktr.ee" {
		t.Fatalf("unexpected crawl graph: %#v", requested)
	}
	if len(updated.ProfileLinks) < 5 {
		t.Fatalf("expected hub + outbound links, got %#v", updated.ProfileLinks)
	}
	if updated.ProfileLinks[0].Kind != "modpack" {
		t.Fatalf("premium sort should prioritize modpacks, got %#v", updated.ProfileLinks)
	}
	seen := map[string]CreatorProfileLink{}
	for _, link := range updated.ProfileLinks {
		seen[link.URL] = link
		if strings.Contains(link.URL, "old.example") {
			t.Fatalf("successful refresh retained stale auto-discovered link: %#v", updated.ProfileLinks)
		}
	}
	for _, want := range []string{
		"https://www.curseforge.com/minecraft/modpacks/example-pack",
		"https://discord.gg/example",
		"https://www.twitch.tv/example",
		"https://drive.google.com/file/d/pack/view",
	} {
		if _, ok := seen[want]; !ok {
			t.Errorf("discovered target %s missing from %#v", want, updated.ProfileLinks)
		}
	}
	if got := seen["https://drive.google.com/file/d/pack/view"].Kind; got != "modpack" {
		t.Fatalf("label-aware generic download classification=%q want modpack", got)
	}
}

func TestRefreshCreatorProfileLinksKeepsLastKnownGoodCacheWhenProviderBlocks(t *testing.T) {
	app := &App{httpClient: &http.Client{Transport: creatorRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return creatorHTMLResponse(req, http.StatusForbidden, "blocked"), nil
	})}}
	old := CreatorProfileLink{URL: "https://modrinth.com/modpack/known-good", Label: "Known Good", Kind: "modpack", EvidenceType: "link-hub", FirstSeenAt: "2026-01-01T00:00:00Z", LastSeenAt: "2026-01-01T00:00:00Z"}
	ch := CreatorChannel{Platform: "tiktok", Handle: "@cached", URL: "https://www.tiktok.com/@cached", ProfileLinks: []CreatorProfileLink{old}}
	updated, err := app.refreshCreatorProfileLinks(context.Background(), ch)
	if err == nil {
		t.Fatal("blocked provider refresh should return a truthful warning error")
	}
	if updated.ProfileLinksStatus != "cached" || updated.ProfileLinksError == "" {
		t.Fatalf("cached status/error missing: %#v", updated)
	}
	if len(updated.ProfileLinks) != 1 || updated.ProfileLinks[0].URL != old.URL {
		t.Fatalf("last-known-good links were lost on transient failure: %#v", updated.ProfileLinks)
	}
}

func TestKatsumiDefaultSeedIncludesCurrentProfileHubs(t *testing.T) {
	app := &App{cfgDir: t.TempDir(), creatorSyncRunning: map[string]bool{}}
	app.ensureDefaultCreatorChannels()
	for _, ch := range app.creatorChannels {
		if strings.EqualFold(ch.Handle, "@its_katsumi") {
			if !ch.Required || ch.Source != "curated-core" {
				t.Fatalf("Katsumi is no longer protected curated-core: %#v", ch)
			}
			joined := strings.ToLower(strings.Join(ch.ProfileHubURLs, " "))
			if !strings.Contains(joined, "lnk.bio/itskatsumii") || !strings.Contains(joined, "linktr.ee/itskatsumii") {
				t.Fatalf("Katsumi current hubs missing: %#v", ch.ProfileHubURLs)
			}
			if ch.ProfileLinksStatus != "seeded" {
				t.Fatalf("Katsumi seeded link status=%q", ch.ProfileLinksStatus)
			}
			if len(ch.ProfileLinks) < 2 {
				t.Fatalf("Katsumi seeded cached profile links missing: %#v", ch.ProfileLinks)
			}
			return
		}
	}
	t.Fatal("Katsumi built-in creator missing")
}

func TestCreatorProfileLinksRefreshAPIPersistsMetadata(t *testing.T) {
	cfg := t.TempDir()
	client := &http.Client{Transport: creatorRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Hostname() {
		case "www.tiktok.com":
			return creatorHTMLResponse(req, http.StatusOK, `<a href="https://linktr.ee/PersistCreator">Links</a>`), nil
		case "linktr.ee":
			return creatorHTMLResponse(req, http.StatusOK, `<a href="https://modrinth.com/modpack/persist-pack">My Modpack</a>`), nil
		case "api.modrinth.com":
			if strings.HasSuffix(req.URL.Path, "/members") {
				return creatorJSONResponse(req, http.StatusOK, `[]`), nil
			}
			return creatorJSONResponse(req, http.StatusOK, `{"id":"persist-id","slug":"persist-pack","title":"Persist Pack","description":"Creator-linked pack","project_type":"modpack","downloads":42,"updated":"2026-08-23T00:00:00Z","game_versions":["1.21.1"],"loaders":["fabric"]}`), nil
		default:
			t.Fatalf("unexpected fetch host: %s", req.URL.Hostname())
			return nil, nil
		}
	})}
	app := &App{cfgDir: cfg, httpClient: client, creatorSyncRunning: map[string]bool{}, creatorChannels: []CreatorChannel{{Platform: "tiktok", Handle: "@persistcreator", URL: "https://www.tiktok.com/@persistcreator", Title: "Persist Creator", Source: "custom"}}}
	if err := app.saveCreatorChannels(); err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/creators/channels/links/refresh", strings.NewReader(`{"url":"https://www.tiktok.com/@persistcreator"}`))
	app.handleCreatorProfileLinksRefresh(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("profile links refresh status=%d body=%s", rr.Code, rr.Body.String())
	}
	var response struct {
		Count int `json:"count"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Count < 2 {
		t.Fatalf("expected hub + modpack in response: %s", rr.Body.String())
	}

	reloaded := &App{cfgDir: cfg, creatorSyncRunning: map[string]bool{}}
	reloaded.loadPersistentCaches()
	for _, ch := range reloaded.creatorChannels {
		if strings.EqualFold(ch.Handle, "@persistcreator") {
			if ch.ProfileLinksStatus != "ready" || ch.ProfileLinksRefreshedAt == "" {
				t.Fatalf("persisted profile metadata incomplete: %#v", ch)
			}
			for _, link := range ch.ProfileLinks {
				if link.URL == "https://modrinth.com/modpack/persist-pack" && link.Kind == "modpack" {
					return
				}
			}
			t.Fatalf("persisted modpack missing: %#v", ch.ProfileLinks)
		}
	}
	t.Fatal("custom creator missing after reload")
}

func TestNormalizeCreatorProfileLinkURLUnwrapsPlatformRedirects(t *testing.T) {
	raw := "https://www.youtube.com/redirect?q=https%3A%2F%2Flinktr.ee%2FCreator%3Futm_source%3Dyoutube"
	if got := normalizeCreatorProfileLinkURL(raw, nil); got != "https://linktr.ee/Creator" {
		t.Fatalf("wrapped creator destination=%q", got)
	}
}

func TestCreatorHubOutboundRedirectIsRecordedWithoutFetchingDestination(t *testing.T) {
	requests := []string{}
	client := &http.Client{Transport: creatorRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		requests = append(requests, req.URL.Hostname()+req.URL.Path)
		if req.URL.Hostname() != "linktr.ee" {
			t.Fatalf("outbound redirect target was fetched: %s", req.URL.String())
		}
		res := creatorHTMLResponse(req, http.StatusFound, "")
		res.Header.Set("Location", "https://www.curseforge.com/minecraft/modpacks/redirected-pack")
		return res, nil
	})}
	app := &App{httpClient: client}
	body, finalURL, err := app.fetchCreatorPublicPage(context.Background(), "https://linktr.ee/redirect/123")
	if err != nil {
		t.Fatal(err)
	}
	if finalURL != "https://linktr.ee/redirect/123" {
		t.Fatalf("final evidence URL=%q", finalURL)
	}
	links := extractCreatorRawLinks(string(body), finalURL)
	if len(links) != 1 || links[0].URL != "https://www.curseforge.com/minecraft/modpacks/redirected-pack" {
		t.Fatalf("redirect target not captured: %#v", links)
	}
	if len(requests) != 1 {
		t.Fatalf("crawler made outbound request(s): %#v", requests)
	}
}

func TestRefreshCreatorProfileLinksUsesArchivedUploadDescriptionsAsHubEvidence(t *testing.T) {
	requested := []string{}
	client := &http.Client{Transport: creatorRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		requested = append(requested, req.URL.Hostname())
		switch req.URL.Hostname() {
		case "www.youtube.com":
			return creatorHTMLResponse(req, http.StatusOK, `<html><body>No public links here</body></html>`), nil
		case "beacons.ai":
			return creatorHTMLResponse(req, http.StatusOK, `<a href="https://modrinth.com/modpack/from-description">Download my modpack</a>`), nil
		default:
			t.Fatalf("unexpected fetched host %s", req.URL.Hostname())
			return nil, nil
		}
	})}
	ch := CreatorChannel{Platform: "youtube", Handle: "@desc", URL: "https://www.youtube.com/@desc"}
	app := &App{httpClient: client, creatorVideos: []CreatorVideo{{ID: "v1", Platform: "youtube", ChannelHandle: "@desc", URL: "https://www.youtube.com/watch?v=v1", Description: "All my links: https://beacons.ai/DescCreator"}}}
	updated, err := app.refreshCreatorProfileLinks(context.Background(), ch)
	if err != nil {
		t.Fatal(err)
	}
	if len(requested) != 2 || requested[0] != "www.youtube.com" || requested[1] != "beacons.ai" {
		t.Fatalf("unexpected profile/description hub crawl: %#v", requested)
	}
	found := false
	for _, link := range updated.ProfileLinks {
		if link.URL == "https://modrinth.com/modpack/from-description" && link.Kind == "modpack" {
			found = true
		}
	}
	if !found {
		t.Fatalf("modpack behind upload-description hub missing: %#v", updated.ProfileLinks)
	}
}
