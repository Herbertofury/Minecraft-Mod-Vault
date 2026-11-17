package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

func bilibiliArchiveFixture(page, total, n int) string {
	entries := []map[string]any{}
	for i := 0; i < n; i++ {
		entries = append(entries, map[string]any{"bvid": fmt.Sprintf("BV%010d", (page-1)*30+i+1), "mid": 42, "author": "Creator", "title": "Public upload", "created": 1700000000})
	}
	b, _ := json.Marshal(map[string]any{"code": 0, "data": map[string]any{"page": map[string]any{"pn": page, "ps": 30, "count": total}, "list": map[string]any{"vlist": entries}}})
	return string(b)
}

func TestBilibiliCreatorNormalization(t *testing.T) {
	c, err := normalizeCreatorChannelInput("https://space.bilibili.com/42/video?tracking=ignored")
	if err != nil || c.Platform != "bilibili" || c.ChannelID != "42" || c.URL != "https://space.bilibili.com/42" {
		t.Fatalf("%#v %v", c, err)
	}
	for _, raw := range []string{"https://space.bilibili.com.attacker.invalid/42", "https://space.bilibili.com/42/favorites", "https://user:secret@space.bilibili.com/42", "https://space.bilibili.com/0", "ftp://space.bilibili.com/42"} {
		if _, err := normalizeBilibiliCreator(raw); err == nil {
			t.Errorf("accepted %s", raw)
		}
	}
}

func TestBilibiliCreatorArchivePagesAndCoverage(t *testing.T) {
	calls := 0
	app := &App{httpClient: &http.Client{Transport: creatorRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		pn, _ := strconv.Atoi(r.URL.Query().Get("pn"))
		n := 30
		if pn == 2 {
			n = 1
		}
		return creatorJSONResponse(r, 200, bilibiliArchiveFixture(pn, 31, n)), nil
	})}}
	c, v, err := app.enumerateBilibiliCreator(context.Background(), CreatorChannel{URL: "https://space.bilibili.com/42"}, true)
	if err != nil || calls != 2 || len(v) != 31 || c.ArchiveCoverage != "public-upload-list-complete" || c.ArchiveExpected != 31 || c.ArchiveEnumerated != 31 {
		t.Fatalf("%#v %d %d %v", c, len(v), calls, err)
	}
	if v[30].ID != "bilibili_BV0000000031_p1" || v[0].AnalyzedAt != "" || creatorVideoChannelKey(v[0]) != "bilibili:42" {
		t.Fatal("listing confused with completed analysis/identity")
	}
}

func TestBilibiliCreatorRecentWindowNeverClaimsFull(t *testing.T) {
	calls := 0
	app := &App{httpClient: &http.Client{Transport: creatorRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		pn, _ := strconv.Atoi(r.URL.Query().Get("pn"))
		return creatorJSONResponse(r, 200, bilibiliArchiveFixture(pn, 10000, 30)), nil
	})}}
	c, v, err := app.enumerateBilibiliCreator(context.Background(), CreatorChannel{URL: "https://space.bilibili.com/42"}, false)
	if err != nil || calls != 3 || len(v) != 90 || c.ArchiveCoverage != "recent-window-only" || c.TotalVideos != 10000 {
		t.Fatalf("%#v %d %d %v", c, len(v), calls, err)
	}
}

func TestBilibiliCreatorIncompleteAndRestrictedArchivesFail(t *testing.T) {
	for _, mode := range []string{"rate", "short", "duplicate", "changed", "identity"} {
		t.Run(mode, func(t *testing.T) {
			calls := 0
			app := &App{httpClient: &http.Client{Transport: creatorRoundTripFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				pn, _ := strconv.Atoi(r.URL.Query().Get("pn"))
				body := bilibiliArchiveFixture(pn, 31, 30)
				switch mode {
				case "rate":
					body = `{"code":-799}`
				case "short":
					body = bilibiliArchiveFixture(pn, 31, 1)
				case "duplicate":
					if pn == 2 {
						body = bilibiliArchiveFixture(1, 31, 1)
						body = strings.Replace(body, `"pn":1`, `"pn":2`, 1)
					}
				case "changed":
					if pn == 2 {
						body = bilibiliArchiveFixture(pn, 32, 2)
					}
				case "identity":
					body = strings.ReplaceAll(body, `"mid":42`, `"mid":43`)
				}
				return creatorJSONResponse(r, 200, body), nil
			})}}
			c, _, err := app.enumerateBilibiliCreator(context.Background(), CreatorChannel{URL: "https://space.bilibili.com/42"}, true)
			if err == nil || c.ArchiveCoverage != "incomplete" || calls > 2 {
				t.Fatalf("false success or retry: %#v %d %v", c, calls, err)
			}
		})
	}
}

func TestBilibiliCreatorCancellationAndFailedSyncKeepArchive(t *testing.T) {
	app := &App{cfgDir: t.TempDir(), httpClient: &http.Client{Transport: creatorRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		return creatorJSONResponse(r, 200, `{"code":-799}`), nil
	})}, creatorChannels: []CreatorChannel{{Platform: "bilibili", ChannelID: "42", URL: "https://space.bilibili.com/42", LastFullSyncAt: "previous-success"}}, creatorVideos: []CreatorVideo{{ID: "old", Platform: "bilibili", Title: "retained"}}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := app.enumerateBilibiliCreator(ctx, app.creatorChannels[0], true); err != context.Canceled {
		t.Fatalf("cancellation lost: %v", err)
	}
	if err := app.syncCreatorChannel(context.Background(), app.creatorChannels[0], true); err == nil {
		t.Fatal("restricted listing counted as sync")
	}
	if len(app.creatorVideos) != 1 || app.creatorVideos[0].Title != "retained" || app.creatorChannels[0].LastFullSyncAt != "previous-success" {
		t.Fatal("failed archive refresh discarded old state")
	}
}

func TestBilibiliArchiveRequiresPresentCountAndList(t *testing.T) {
	for _, body := range []string{
		`{"code":0,"data":{"page":{"pn":1,"ps":30},"list":{"vlist":[]}}}`,
		`{"code":0,"data":{"page":{"pn":1,"ps":30,"count":null},"list":{"vlist":[]}}}`,
		`{"code":0,"data":{"page":{"pn":1,"ps":30,"count":0},"list":{"vlist":null}}}`,
		`{"code":0,"data":{"page":{"pn":1,"ps":30,"count":0},"list":{}}}`,
	} {
		app := &App{httpClient: &http.Client{Transport: creatorRoundTripFunc(func(r *http.Request) (*http.Response, error) { return creatorJSONResponse(r, 200, body), nil })}}
		c, _, err := app.enumerateBilibiliCreator(context.Background(), CreatorChannel{URL: "https://space.bilibili.com/42"}, true)
		if err == nil || c.ArchiveCoverage == "public-upload-list-complete" {
			t.Fatalf("missing fields became empty success: %s", body)
		}
	}
}

func TestBilibiliEmptyArchiveSyncPreservesHistoricalEvidence(t *testing.T) {
	seed := CreatorChannel{Platform: "bilibili", ChannelID: "42", URL: "https://space.bilibili.com/42", LastFullSyncAt: "old"}
	app := &App{cfgDir: t.TempDir(), creatorChannels: []CreatorChannel{seed}, creatorVideos: []CreatorVideo{{ID: "historical", Platform: "bilibili", ChannelID: "42", Title: "Keep me"}}, httpClient: &http.Client{Transport: creatorRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/x/space/arc/search" {
			return creatorJSONResponse(r, 200, bilibiliArchiveFixture(1, 0, 0)), nil
		}
		return creatorJSONResponse(r, 404, `{}`), nil
	})}}
	if err := app.syncCreatorChannel(context.Background(), seed, true); err != nil {
		t.Fatal(err)
	}
	if len(app.creatorVideos) != 1 || app.creatorVideos[0].Title != "Keep me" || app.creatorChannels[0].LastFullSyncAt == "old" || app.creatorChannels[0].ArchiveExpected != 0 {
		t.Fatal("empty current listing erased historical evidence or failed to advance verified sync")
	}
}

func TestBilibiliArchiveAggregateTextBudget(t *testing.T) {
	calls := 0
	app := &App{httpClient: &http.Client{Transport: creatorRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		pn, _ := strconv.Atoi(r.URL.Query().Get("pn"))
		body := bilibiliArchiveFixture(pn, 1000, 30)
		body = strings.ReplaceAll(body, `"title":"Public upload"`, `"title":"Public upload","description":"`+strings.Repeat("x", 65000)+`"`)
		return creatorJSONResponse(r, 200, body), nil
	})}}
	c, _, err := app.enumerateBilibiliCreator(context.Background(), CreatorChannel{URL: "https://space.bilibili.com/42"}, true)
	if err == nil || !strings.Contains(err.Error(), "text budget") || calls > 9 || c.ArchiveCoverage != "incomplete" {
		t.Fatalf("unbounded retention: calls=%d coverage=%s err=%v", calls, c.ArchiveCoverage, err)
	}
}
