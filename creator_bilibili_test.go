package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

const bilibiliFixture = `{"code":0,"data":{"bvid":"BV1JV4y1y71n","aid":42,"title":"Mod spotlight","desc":"https://modrinth.com/mod/sodium/version/example","pic":"https://i0.hdslb.com/cover.jpg","pubdate":1700000000,"owner":{"mid":7,"name":"Creator"},"pages":[{"cid":12,"page":1,"part":"First","duration":50},{"cid":13,"page":2,"part":"Second","duration":80}]}}`

func TestBilibiliInputAndPartIdentity(t *testing.T) {
	for _, raw := range []string{"https://bilibili.com.attacker.invalid/video/BV1JV4y1y71n", "https://space.bilibili.com/123", "https://b23.tv/abc", "https://bilibili.com/video/BV1JV4y1y71n?p=0", "https://bilibili.com/video/BV1JV4y1y71n?p=no"} {
		if _, _, err := bilibiliVideoInput(raw); err == nil {
			t.Errorf("accepted %s", raw)
		}
	}
	app := &App{httpClient: &http.Client{Transport: creatorRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		return creatorJSONResponse(r, 200, bilibiliFixture), nil
	})}}
	v, aid, cid, err := app.fetchBilibiliVideo(context.Background(), "https://www.bilibili.com/video/BV1JV4y1y71n?p=2&tracking=discard")
	if err != nil || aid != 42 || cid != 13 || v.ID != "bilibili_BV1JV4y1y71n_p2" || v.DurationSeconds != 80 || len(v.Warnings) != 1 {
		t.Fatalf("%#v %d %d %v", v, aid, cid, err)
	}
	if creatorVideoChannelKey(v) != "bilibili:7" {
		t.Fatal("platform identity lost")
	}
	if got := timestampVideoLink(v, 12); !strings.Contains(got, "p=2&t=12") {
		t.Fatalf("timestamp/part lost: %s", got)
	}
	if _, _, _, err := app.fetchBilibiliVideo(context.Background(), "https://www.bilibili.com/video/BV1JV4y1y71n?p=3"); err == nil {
		t.Fatal("missing part silently substituted")
	}
}

func TestBilibiliAnalysisUsesSharedModExtraction(t *testing.T) {
	app := &App{cfgDir: t.TempDir(), httpClient: &http.Client{Transport: creatorRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/x/web-interface/view":
			return creatorJSONResponse(r, 200, bilibiliFixture), nil
		case "/x/player/v2":
			return creatorJSONResponse(r, 200, `{"code":0,"data":{"subtitle":{"subtitles":[{"subtitle_url":"https://aisubtitle.hdslb.com/track.json"}]}}}`), nil
		case "/track.json":
			return creatorJSONResponse(r, 200, `{"body":[{"from":1.5,"to":3,"content":"测试字幕"}]}`), nil
		default:
			return creatorJSONResponse(r, 404, `{}`), nil
		}
	})}}
	v, err := app.analyzeCreatorVideo(context.Background(), "https://www.bilibili.com/video/BV1JV4y1y71n/")
	if err != nil {
		t.Fatal(err)
	}
	if len(v.Mods) != 1 || v.Mods[0].ProjectID != "sodium" || !v.TranscriptAvailable || v.TranscriptSegments != 1 || v.AnalyzedAt == "" {
		t.Fatalf("incomplete pipeline: %#v", v)
	}
	if len(v.Warnings) < 2 {
		t.Fatal("coverage warnings missing")
	}
}

func TestBilibiliRestrictedAndUntrustedCaptions(t *testing.T) {
	for _, body := range []string{`{"code":-101}`, `{"code":-404}`, `{"code":-412}`} {
		app := &App{httpClient: &http.Client{Transport: creatorRoundTripFunc(func(r *http.Request) (*http.Response, error) { return creatorJSONResponse(r, 200, body), nil })}}
		if _, _, _, err := app.fetchBilibiliVideo(context.Background(), "https://www.bilibili.com/video/BV1JV4y1y71n/"); err == nil {
			t.Fatalf("error code became success: %s", body)
		}
	}
	requests := 0
	app := &App{httpClient: &http.Client{Transport: creatorRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		requests++
		return creatorJSONResponse(r, 200, `{"code":0,"data":{"subtitle":{"subtitles":[{"subtitle_url":"https://evil.invalid/track.json"}]}}}`), nil
	})}}
	if _, err := app.bilibiliPublicCaptions(context.Background(), 42, 12, 50); err == nil || requests != 1 {
		t.Fatal("untrusted caption fetched")
	}
}

func TestBilibiliCaptionRedirectRejected(t *testing.T) {
	requests := 0
	app := &App{httpClient: &http.Client{Transport: creatorRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		requests++
		res := creatorJSONResponse(r, 302, "")
		res.Header.Set("Location", "https://evil.invalid/track")
		return res, nil
	})}}
	var dst any
	if err := app.getBilibiliCaptionJSON(context.Background(), "https://aisubtitle.hdslb.com/track", &dst); err == nil || requests != 1 {
		t.Fatal("unsafe redirect followed")
	}
}

// Opt-in network proof, never required by offline CI. Does not download media,
// log in, run speech models or claim complete extraction from the video.
func TestBilibiliLiveMetadata(t *testing.T) {
	if os.Getenv("MMV_BILIBILI_LIVE") != "1" {
		t.Skip("explicit live-source opt-in required")
	}
	app := &App{httpClient: &http.Client{Timeout: 20 * time.Second}}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	v, _, _, err := app.fetchBilibiliVideo(ctx, "https://www.bilibili.com/video/BV1JV4y1y71n/")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, raw := range extractURLs(v.Description) {
		if mod, ok := creatorModFromURL(raw, 0, v); ok && mod.ProjectID == "sophisticated-backpacks" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected creator-provided backpack project link was absent")
	}
	t.Logf("Live source %s: metadata and creator-provided sophisticated-backpacks project identity verified", v.URL)
}

func TestBilibiliTranscriptEvidenceUsesActualPlatform(t *testing.T) {
	app := &App{cfgDir: t.TempDir(), creatorVideos: []CreatorVideo{{ID: "bilibili_BV1JV4y1y71n_p2", Platform: "bilibili", URL: "https://www.bilibili.com/video/BV1JV4y1y71n/?p=2"}}}
	if err := app.saveCreatorTranscript("bilibili_BV1JV4y1y71n_p2", "fixture", []TranscriptSegment{{StartMS: 12500, EndMS: 14000, Text: "evidence"}}); err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	app.handleCreatorTranscript(rec, httptest.NewRequest("GET", "/api/creators/transcript?videoId=bilibili_BV1JV4y1y71n_p2", nil))
	var got CreatorTranscriptRecord
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Segments) != 1 || got.Segments[0].VideoLink != "https://www.bilibili.com/video/BV1JV4y1y71n/?p=2&t=12" {
		t.Fatalf("incorrect evidence link: %s", rec.Body.String())
	}
}

func TestBilibiliPublicMetadataDoesNotSendAmbientCookies(t *testing.T) {
	jar, _ := cookiejar.New(nil)
	host, _ := url.Parse("https://api.bilibili.com")
	jar.SetCookies(host, []*http.Cookie{{Name: "session", Value: "private"}})
	app := &App{httpClient: &http.Client{Jar: jar, Transport: creatorRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.Header.Get("Cookie") != "" {
			t.Fatal("ambient account cookie transmitted")
		}
		return creatorJSONResponse(r, 200, bilibiliFixture), nil
	})}}
	if _, _, _, err := app.fetchBilibiliVideo(context.Background(), "https://www.bilibili.com/video/BV1JV4y1y71n/"); err != nil {
		t.Fatal(err)
	}
}

func TestBilibiliCaptionRejectsOverflowAndOutOfRangeTimes(t *testing.T) {
	app := &App{httpClient: &http.Client{Transport: creatorRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/x/player/v2" {
			return creatorJSONResponse(r, 200, `{"code":0,"data":{"subtitle":{"subtitles":[{"subtitle_url":"https://aisubtitle.hdslb.com/track.json"}]}}}`), nil
		}
		return creatorJSONResponse(r, 200, `{"body":[{"from":1e100,"to":1e101,"content":"overflow"},{"from":-1,"to":2,"content":"negative"},{"from":4,"to":4,"content":"zero"},{"from":3,"to":100,"content":"past duration"},{"from":1,"to":2,"content":"valid"}]}`), nil
	})}}
	segments, err := app.bilibiliPublicCaptions(context.Background(), 42, 12, 50)
	if err != nil || len(segments) != 1 || segments[0].StartMS != 1000 || segments[0].EndMS != 2000 {
		t.Fatalf("invalid timestamps survived: %#v %v", segments, err)
	}
}
