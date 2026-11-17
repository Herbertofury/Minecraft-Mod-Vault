package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var bilibiliCreatorAPI = "https://api.bilibili.com"
var bilibiliVideoID = regexp.MustCompile(`^BV[0-9A-Za-z]{10}$`)

type bilibiliVideoEnvelope struct {
	Code int `json:"code"`
	Data struct {
		BVID        string `json:"bvid"`
		AID         int64  `json:"aid"`
		Title       string `json:"title"`
		Description string `json:"desc"`
		Thumbnail   string `json:"pic"`
		Published   int64  `json:"pubdate"`
		State       int    `json:"state"`
		Owner       struct {
			MID  int64  `json:"mid"`
			Name string `json:"name"`
			Face string `json:"face"`
		} `json:"owner"`
		Pages []struct {
			CID      int64  `json:"cid"`
			Page     int    `json:"page"`
			Part     string `json:"part"`
			Duration int64  `json:"duration"`
		} `json:"pages"`
	} `json:"data"`
}

func bilibiliVideoInput(raw string) (string, int, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || !creatorPublicWebURL(u) || !creatorHostMatches(u.Hostname(), []string{"bilibili.com"}) {
		return "", 0, errors.New("invalid Bilibili video URL")
	}
	parts := splitPath(u.Path)
	if len(parts) != 2 || parts[0] != "video" || !bilibiliVideoID.MatchString(parts[1]) {
		return "", 0, errors.New("paste a Bilibili /video/BV… URL; paste creator profiles in the Follow form; short links are not supported yet")
	}
	page := 1
	if rawPage := u.Query().Get("p"); rawPage != "" {
		page, err = strconv.Atoi(rawPage)
		if err != nil || page < 1 {
			return "", 0, errors.New("invalid Bilibili part number")
		}
	}
	return parts[1], page, nil
}

func bilibiliAccessError(code int) error {
	switch code {
	case -101, -403, -10403:
		return fmt.Errorf("Bilibili access restricted or authentication required (code %d); no login or access bypass attempted", code)
	case -404, 62002, 62012:
		return fmt.Errorf("Bilibili video is unavailable or private (code %d)", code)
	case -412, -509, -799:
		return fmt.Errorf("Bilibili rate/access limit (code %d); retry later", code)
	default:
		return fmt.Errorf("Bilibili API rejected request (code %d)", code)
	}
}

func (a *App) fetchBilibiliVideo(ctx context.Context, raw string) (CreatorVideo, int64, int64, error) {
	id, page, err := bilibiliVideoInput(raw)
	if err != nil {
		return CreatorVideo{}, 0, 0, err
	}
	var envelope bilibiliVideoEnvelope
	if err = a.getBilibiliPublicJSON(ctx, bilibiliCreatorAPI+"/x/web-interface/view?bvid="+url.QueryEscape(id), nil, &envelope); err != nil {
		return CreatorVideo{}, 0, 0, fmt.Errorf("Bilibili metadata: %w", err)
	}
	if envelope.Code != 0 {
		return CreatorVideo{}, 0, 0, bilibiliAccessError(envelope.Code)
	}
	d := envelope.Data
	if d.BVID != id || d.AID <= 0 || d.Title == "" || d.State < 0 {
		return CreatorVideo{}, 0, 0, errors.New("Bilibili returned unavailable or mismatched video metadata")
	}
	for _, part := range d.Pages {
		if part.Page != page {
			continue
		}
		if part.CID <= 0 {
			return CreatorVideo{}, 0, 0, errors.New("Bilibili returned an invalid part identity")
		}
		canonical := "https://www.bilibili.com/video/" + id + "/?p=" + strconv.Itoa(page)
		now := time.Now().UTC().Format(time.RFC3339)
		video := CreatorVideo{ID: "bilibili_" + id + "_p" + strconv.Itoa(page), Platform: "bilibili", URL: canonical, Title: d.Title, Description: d.Description, Creator: d.Owner.Name, CreatorURL: "https://space.bilibili.com/" + strconv.FormatInt(d.Owner.MID, 10), CreatorAvatarURL: d.Owner.Face, ThumbnailURL: d.Thumbnail, DiscoveredAt: now, ChannelID: strconv.FormatInt(d.Owner.MID, 10), VideoKind: "video", DurationSeconds: part.Duration, Mods: []CreatorMod{}}
		if len(d.Pages) > 1 {
			video.Title += " · " + part.Part
			video.Warnings = append(video.Warnings, fmt.Sprintf("Only part %d of %d is analyzed; other parts remain unscanned.", page, len(d.Pages)))
		}
		if d.Published > 0 {
			video.PublishedAt = time.Unix(d.Published, 0).UTC().Format(time.RFC3339)
		}
		return video, d.AID, part.CID, nil
	}
	return CreatorVideo{}, 0, 0, errors.New("requested Bilibili part was not returned; no other part was substituted")
}

func (a *App) bilibiliPublicCaptions(ctx context.Context, aid, cid, duration int64) ([]TranscriptSegment, error) {
	var player struct {
		Code int `json:"code"`
		Data struct {
			Subtitle struct {
				Subtitles []struct {
					URL string `json:"subtitle_url"`
				} `json:"subtitles"`
			} `json:"subtitle"`
		} `json:"data"`
	}
	target := fmt.Sprintf("%s/x/player/v2?aid=%d&cid=%d", bilibiliCreatorAPI, aid, cid)
	if err := a.getBilibiliPublicJSON(ctx, target, nil, &player); err != nil {
		return nil, err
	}
	if player.Code != 0 {
		return nil, bilibiliAccessError(player.Code)
	}
	if len(player.Data.Subtitle.Subtitles) == 0 {
		return nil, errors.New("no public subtitle track exposed; speech and on-screen text are not covered")
	}
	// Analyze the first exposed track, without assuming it is original or human-authored.
	// Language/translation selection remains a future user control.
	raw := player.Data.Subtitle.Subtitles[0].URL
	if strings.HasPrefix(raw, "//") {
		raw = "https:" + raw
	}
	u, err := url.Parse(raw)
	if err != nil || !creatorPublicWebURL(u) || u.Scheme != "https" || !creatorHostMatches(u.Hostname(), []string{"hdslb.com", "bilibili.com"}) {
		return nil, errors.New("untrusted Bilibili subtitle destination")
	}
	var track struct {
		Body []struct {
			From    float64 `json:"from"`
			To      float64 `json:"to"`
			Content string  `json:"content"`
		} `json:"body"`
	}
	if err = a.getBilibiliCaptionJSON(ctx, u.String(), &track); err != nil {
		return nil, err
	}
	segments := []TranscriptSegment{}
	for _, item := range track.Body {
		if strings.TrimSpace(item.Content) == "" || math.IsNaN(item.From) || math.IsNaN(item.To) || math.IsInf(item.From, 0) || math.IsInf(item.To, 0) || item.From < 0 || item.To <= item.From || item.To > 7*24*60*60 || (duration > 0 && item.To > float64(duration)+1) {
			continue
		}
		segments = append(segments, TranscriptSegment{StartMS: int64(item.From * 1000), EndMS: int64(item.To * 1000), Text: item.Content, Source: "Bilibili public captions (first exposed track)"})
	}
	if len(segments) == 0 {
		return nil, errors.New("Bilibili subtitle track was empty or invalid")
	}
	return segments, nil
}

func (a *App) analyzeBilibili(ctx context.Context, raw string) (CreatorVideo, error) {
	video, aid, cid, err := a.fetchBilibiliVideo(ctx, raw)
	if err != nil {
		return CreatorVideo{}, err
	}
	segments, captionErr := a.bilibiliPublicCaptions(ctx, aid, cid, video.DurationSeconds)
	if captionErr != nil {
		video.Warnings = append(video.Warnings, "Public captions unavailable: "+captionErr.Error())
	}
	if ctx.Err() != nil {
		return CreatorVideo{}, ctx.Err()
	}
	video.TranscriptAvailable = len(segments) > 0
	video.TranscriptSegments = len(segments)
	if len(segments) > 0 {
		video.TranscriptSource = "Bilibili public captions (first exposed track)"
		for _, segment := range segments {
			video.TranscriptWords += len(strings.Fields(segment.Text))
		}
		if err := a.saveCreatorTranscript(video.ID, video.TranscriptSource, segments); err != nil {
			video.Warnings = append(video.Warnings, "Timed transcript could not be saved: "+err.Error())
		}
	}
	video.Mods, video.UnresolvedMentions = a.extractCreatorMods(ctx, video, segments)
	if ctx.Err() != nil {
		return CreatorVideo{}, ctx.Err()
	}
	video.AnalyzedAt = time.Now().UTC().Format(time.RFC3339)
	video.Warnings = append(video.Warnings, "Bilibili coverage: public description and available captions only. No speech transcription, OCR, comment scan or creator-archive scan was performed; visual-only and non-English name resolution may be incomplete.")
	return video, nil
}

// Subtitle redirects must remain on the same trusted public CDN family. Never
// inherit a cookie jar or silently follow a metadata-supplied arbitrary host.
func (a *App) getBilibiliCaptionJSON(ctx context.Context, target string, dst any) error {
	client := *a.httpClient
	client.Jar = nil
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 || req.URL.Scheme != "https" || !creatorPublicWebURL(req.URL) || !creatorHostMatches(req.URL.Hostname(), []string{"hdslb.com", "bilibili.com"}) {
			return errors.New("untrusted Bilibili caption redirect")
		}
		return nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return err
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Bilibili captions returned HTTP %d", res.StatusCode)
	}
	return json.NewDecoder(io.LimitReader(res.Body, 8<<20)).Decode(dst)
}

// Public metadata is read without any ambient account cookies. API redirects
// are not needed for these fixed endpoints and are rejected, not followed.
func (a *App) getBilibiliPublicJSON(ctx context.Context, target string, headers map[string]string, dst any) error {
	client := *a.httpClient
	client.Jar = nil
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return errors.New("Bilibili metadata redirect refused")
	}
	public := &App{httpClient: &client}
	return public.getJSON(ctx, target, headers, dst)
}
