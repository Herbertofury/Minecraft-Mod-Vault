package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type CreatorVideo struct {
	ID                  string           `json:"id"`
	Platform            string           `json:"platform"`
	URL                 string           `json:"url"`
	Title               string           `json:"title"`
	Creator             string           `json:"creator"`
	CreatorURL          string           `json:"creatorUrl,omitempty"`
	CreatorAvatarURL    string           `json:"creatorAvatarUrl,omitempty"`
	ThumbnailURL        string           `json:"thumbnailUrl,omitempty"`
	Description         string           `json:"description,omitempty"`
	PublishedAt         string           `json:"publishedAt,omitempty"`
	DiscoveredAt        string           `json:"discoveredAt"`
	AnalyzedAt          string           `json:"analyzedAt,omitempty"`
	ChannelID           string           `json:"channelId,omitempty"`
	ChannelHandle       string           `json:"channelHandle,omitempty"`
	VideoKind           string           `json:"videoKind,omitempty"`
	DurationSeconds     int64            `json:"durationSeconds,omitempty"`
	CaptionHint         bool             `json:"captionHint,omitempty"`
	TranscriptSource    string           `json:"transcriptSource,omitempty"`
	TranscriptAvailable bool             `json:"transcriptAvailable"`
	TranscriptSegments  int              `json:"transcriptSegments,omitempty"`
	TranscriptWords     int              `json:"transcriptWords,omitempty"`
	AnalysisAttempts    int              `json:"analysisAttempts,omitempty"`
	AnalysisError       string           `json:"analysisError,omitempty"`
	LastAnalysisAttempt string           `json:"lastAnalysisAttempt,omitempty"`
	Mods                []CreatorMod     `json:"mods"`
	Warnings            []string         `json:"warnings,omitempty"`
	UnresolvedMentions  []CreatorMention `json:"unresolvedMentions,omitempty"`
}

type CreatorMention struct {
	Name       string `json:"name"`
	Timestamp  string `json:"timestamp"`
	TimestampS int64  `json:"timestampSeconds"`
	VideoLink  string `json:"videoLink"`
	Evidence   string `json:"evidence"`
}

type CreatorMod struct {
	Name               string   `json:"name"`
	Provider           string   `json:"provider,omitempty"`
	ProjectID          string   `json:"projectId,omitempty"`
	ProjectType        string   `json:"projectType,omitempty"`
	Author             string   `json:"author,omitempty"`
	IconURL            string   `json:"iconUrl,omitempty"`
	URL                string   `json:"url"`
	Timestamp          string   `json:"timestamp"`
	TimestampS         int64    `json:"timestampSeconds"`
	VideoLink          string   `json:"videoLink"`
	Evidence           string   `json:"evidence"`
	SourceKinds        []string `json:"sourceKinds,omitempty"`
	DescriptionContext string   `json:"descriptionContext,omitempty"`
	TranscriptContext  string   `json:"transcriptContext,omitempty"`
	ProjectSummary     string   `json:"projectSummary,omitempty"`
	Confidence         float64  `json:"confidence"`
}

type TranscriptSegment struct {
	StartMS int64
	EndMS   int64
	Text    string
	Source  string
}

type creatorAnalyzeRequest struct {
	URL string `json:"url"`
}

func (a *App) handleCreatorVideos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	includePending := r.URL.Query().Get("includePending") == "1" || strings.EqualFold(r.URL.Query().Get("includePending"), "true")
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	channel := strings.TrimSpace(r.URL.Query().Get("channel"))
	kind := strings.TrimSpace(r.URL.Query().Get("kind"))
	sortBy := firstNonEmpty(r.URL.Query().Get("sort"), "published-desc")
	a.dataMu.RLock()
	visible := make([]CreatorVideo, 0, len(a.creatorVideos))
	pending := 0
	failed := 0
	for _, video := range a.creatorVideos {
		if video.AnalyzedAt == "" {
			pending++
			if video.AnalysisError != "" {
				failed++
			}
			if !includePending {
				continue
			}
		}
		if channel != "" && channel != creatorVideoChannelKey(video) && channel != video.ChannelID && !strings.EqualFold(channel, video.ChannelHandle) && !strings.EqualFold(channel, video.Creator) {
			continue
		}
		if kind != "" && kind != "all" && video.VideoKind != kind {
			continue
		}
		if q != "" {
			hay := strings.ToLower(strings.Join([]string{video.Title, video.Description, video.Creator, video.ChannelHandle, strings.Join(creatorModNames(video.Mods), " ")}, " "))
			if !strings.Contains(hay, q) {
				continue
			}
		}
		visible = append(visible, video)
	}
	a.dataMu.RUnlock()
	sort.SliceStable(visible, func(i, j int) bool {
		a, b := visible[i], visible[j]
		switch sortBy {
		case "published-asc":
			return firstNonEmpty(a.PublishedAt, a.DiscoveredAt) < firstNonEmpty(b.PublishedAt, b.DiscoveredAt)
		case "title":
			return strings.ToLower(a.Title) < strings.ToLower(b.Title)
		case "mods-desc":
			if len(a.Mods) == len(b.Mods) {
				return firstNonEmpty(a.PublishedAt, a.DiscoveredAt) > firstNonEmpty(b.PublishedAt, b.DiscoveredAt)
			}
			return len(a.Mods) > len(b.Mods)
		case "mods-asc":
			if len(a.Mods) == len(b.Mods) {
				return firstNonEmpty(a.PublishedAt, a.DiscoveredAt) > firstNonEmpty(b.PublishedAt, b.DiscoveredAt)
			}
			return len(a.Mods) < len(b.Mods)
		default:
			return firstNonEmpty(a.PublishedAt, a.DiscoveredAt) > firstNonEmpty(b.PublishedAt, b.DiscoveredAt)
		}
	})
	writeJSON(w, http.StatusOK, map[string]any{"videos": visible, "count": len(visible), "pending": pending, "failed": failed, "dynamic": true, "includePending": includePending, "sort": sortBy})
}

func creatorModNames(mods []CreatorMod) []string {
	out := make([]string, 0, len(mods))
	for _, mod := range mods {
		if strings.TrimSpace(mod.Name) != "" {
			out = append(out, mod.Name)
		}
	}
	return out
}

func (a *App) handleCreatorDiscover(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Minute)
	defer cancel()
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	var found []CreatorVideo
	var err error
	if query != "" {
		found, err = a.discoverCreatorVideos(ctx, query)
	} else {
		err = a.refreshCreatorDiscovery(ctx)
		// Resolve a few queued entries immediately so the button produces usable creator
		// lists now. The background loop keeps draining the rest without user work.
		_ = a.processCreatorQueue(ctx, 3)
		a.dataMu.RLock()
		for _, video := range a.creatorVideos {
			if video.AnalyzedAt != "" {
				found = append(found, video)
			}
		}
		a.dataMu.RUnlock()
	}
	if err != nil && len(found) == 0 {
		writeJSON(w, http.StatusBadGateway, APIError{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"videos": found, "count": len(found), "warning": errorString(err)})
}

func (a *App) handleCreatorAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var in creatorAnalyzeRequest
	if err := decodeJSON(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	if strings.TrimSpace(in.URL) == "" {
		writeJSON(w, http.StatusBadRequest, APIError{Error: "video URL is required"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Minute)
	defer cancel()
	video, err := a.analyzeCreatorVideo(ctx, in.URL)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, APIError{Error: err.Error()})
		return
	}
	a.upsertCreatorVideo(video)
	_ = a.saveCreatorVideos()
	writeJSON(w, http.StatusOK, video)
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func (a *App) refreshCreatorDiscovery(ctx context.Context) error {
	a.mu.RLock()
	s := a.settings
	a.mu.RUnlock()
	queries := []string{
		"best minecraft mods " + s.GameVersion + " " + s.Loader,
		"minecraft mods you need " + s.GameVersion,
		"minecraft cozy furniture mods",
		"minecraft underrated mods",
		"minecraft visual immersion mods",
	}
	for _, tag := range s.PreferredTags {
		if len(queries) >= 10 {
			break
		}
		queries = append(queries, "minecraft "+tag+" mods")
	}
	all := []CreatorVideo{}
	var failures []string
	for _, q := range queries {
		videos, err := a.discoverCreatorVideos(ctx, q)
		if err != nil {
			failures = append(failures, err.Error())
			continue
		}
		all = append(all, videos...)
	}
	seen := map[string]bool{}
	dedup := []CreatorVideo{}
	for _, v := range all {
		if seen[v.Platform+":"+v.ID] {
			continue
		}
		seen[v.Platform+":"+v.ID] = true
		dedup = append(dedup, v)
	}
	if len(dedup) == 0 {
		if len(failures) > 0 {
			return errors.New(strings.Join(uniqueStringsPreserve(failures), "; "))
		}
		return errors.New("creator discovery returned no videos")
	}
	if len(dedup) > 60 {
		dedup = dedup[:60]
	}
	for _, v := range dedup {
		a.upsertCreatorVideo(v)
	}
	return a.saveCreatorVideos()
}

func (a *App) discoverCreatorVideos(ctx context.Context, query string) ([]CreatorVideo, error) {
	// Use both creator ecosystems. YouTube prefers its official Data API when configured,
	// with a public search-page fallback. TikTok is discovered from its public search
	// hydration data and is best-effort because the site may require a challenge/login.
	a.mu.RLock()
	key := strings.TrimSpace(a.settings.YouTubeAPIKey)
	a.mu.RUnlock()
	all := []CreatorVideo{}
	errs := []string{}
	var yt []CreatorVideo
	var err error
	if key != "" {
		yt, err = a.discoverYouTubeAPI(ctx, key, query)
	}
	if len(yt) == 0 {
		yt, err = a.discoverYouTubeHTML(ctx, query)
	}
	if err != nil {
		errs = append(errs, "YouTube: "+err.Error())
	} else {
		all = append(all, yt...)
	}
	if tt, ttErr := a.discoverTikTokHTML(ctx, query); ttErr != nil {
		errs = append(errs, "TikTok: "+ttErr.Error())
	} else {
		all = append(all, tt...)
	}
	seen := map[string]bool{}
	out := make([]CreatorVideo, 0, len(all))
	for _, video := range all {
		k := video.Platform + ":" + video.ID
		if video.ID == "" || seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, video)
		if len(out) >= 36 {
			break
		}
	}
	if len(out) == 0 && len(errs) > 0 {
		return nil, errors.New(strings.Join(uniqueStringsPreserve(errs), "; "))
	}
	return out, nil
}

func (a *App) discoverTikTokHTML(ctx context.Context, query string) ([]CreatorVideo, error) {
	body, err := a.getText(ctx, "https://www.tiktok.com/search/video?q="+url.QueryEscape(query), map[string]string{"Accept-Language": "en-US,en;q=0.9"})
	if err != nil {
		return nil, err
	}
	payload, err := extractScriptJSONByID(body, "__UNIVERSAL_DATA_FOR_REHYDRATION__")
	if err != nil {
		return nil, err
	}
	var root any
	if err := json.Unmarshal([]byte(payload), &root); err != nil {
		return nil, err
	}
	var candidates []map[string]any
	collectTikTokItemStructs(root, &candidates)
	out := []CreatorVideo{}
	seen := map[string]bool{}
	now := time.Now().UTC().Format(time.RFC3339)
	for _, item := range candidates {
		id := asString(item["id"])
		author, _ := item["author"].(map[string]any)
		videoMap, _ := item["video"].(map[string]any)
		handle := asString(author["uniqueId"])
		if id == "" || handle == "" || seen[id] {
			continue
		}
		seen[id] = true
		creator := firstNonEmpty(asString(author["nickname"]), handle)
		thumb := firstNonEmpty(asString(videoMap["cover"]), asString(videoMap["originCover"]), deepFirstURL(videoMap["cover"]), deepFirstURL(videoMap["originCover"]))
		avatar := firstNonEmpty(asString(author["avatarLarger"]), asString(author["avatarMedium"]), deepFirstURL(author["avatarLarger"]), deepFirstURL(author["avatarMedium"]))
		published := ""
		if ts := int64(asFloat(item["createTime"])); ts > 0 {
			published = time.Unix(ts, 0).UTC().Format(time.RFC3339)
		}
		videoURL := "https://www.tiktok.com/@" + handle + "/video/" + id
		out = append(out, CreatorVideo{ID: id, Platform: "tiktok", URL: videoURL, Title: asString(item["desc"]), Description: asString(item["desc"]), Creator: creator, CreatorURL: "https://www.tiktok.com/@" + handle, CreatorAvatarURL: avatar, ThumbnailURL: thumb, PublishedAt: published, DiscoveredAt: now, Mods: []CreatorMod{}})
		if len(out) >= 12 {
			break
		}
	}
	if len(out) == 0 {
		return nil, errors.New("public TikTok search returned no readable video items")
	}
	return out, nil
}

func collectTikTokItemStructs(v any, out *[]map[string]any) {
	switch x := v.(type) {
	case map[string]any:
		_, hasAuthor := x["author"].(map[string]any)
		_, hasVideo := x["video"].(map[string]any)
		if hasAuthor && hasVideo && asString(x["id"]) != "" && (asString(x["desc"]) != "" || x["createTime"] != nil) {
			*out = append(*out, x)
		}
		for _, child := range x {
			collectTikTokItemStructs(child, out)
		}
	case []any:
		for _, child := range x {
			collectTikTokItemStructs(child, out)
		}
	}
}

func (a *App) discoverYouTubeAPI(ctx context.Context, key, query string) ([]CreatorVideo, error) {
	u := "https://www.googleapis.com/youtube/v3/search?" + url.Values{"part": {"snippet"}, "type": {"video"}, "maxResults": {"25"}, "order": {"relevance"}, "q": {query}, "key": {key}}.Encode()
	var resp struct {
		Items []struct {
			ID struct {
				VideoID string `json:"videoId"`
			} `json:"id"`
			Snippet struct {
				Title        string `json:"title"`
				Description  string `json:"description"`
				ChannelTitle string `json:"channelTitle"`
				ChannelID    string `json:"channelId"`
				PublishedAt  string `json:"publishedAt"`
				Thumbnails   map[string]struct {
					URL string `json:"url"`
				} `json:"thumbnails"`
			} `json:"snippet"`
		} `json:"items"`
	}
	if err := a.getJSON(ctx, u, nil, &resp); err != nil {
		return nil, err
	}
	out := []CreatorVideo{}
	for _, x := range resp.Items {
		if x.ID.VideoID == "" {
			continue
		}
		thumb := ""
		for _, k := range []string{"maxres", "standard", "high", "medium", "default"} {
			if t, ok := x.Snippet.Thumbnails[k]; ok && t.URL != "" {
				thumb = t.URL
				break
			}
		}
		out = append(out, CreatorVideo{ID: x.ID.VideoID, Platform: "youtube", URL: "https://www.youtube.com/watch?v=" + x.ID.VideoID, Title: html.UnescapeString(x.Snippet.Title), Creator: x.Snippet.ChannelTitle, CreatorURL: "https://www.youtube.com/channel/" + x.Snippet.ChannelID, ThumbnailURL: thumb, Description: x.Snippet.Description, PublishedAt: x.Snippet.PublishedAt, DiscoveredAt: time.Now().UTC().Format(time.RFC3339), Mods: []CreatorMod{}})
	}
	return out, nil
}

func (a *App) discoverYouTubeHTML(ctx context.Context, query string) ([]CreatorVideo, error) {
	body, err := a.getText(ctx, "https://www.youtube.com/results?search_query="+url.QueryEscape(query), map[string]string{"Accept-Language": "en-US,en;q=0.9"})
	if err != nil {
		return nil, err
	}
	data, err := extractAssignedJSON(body, "ytInitialData")
	if err != nil {
		return nil, fmt.Errorf("YouTube discovery parser: %w", err)
	}
	var root any
	if err := json.Unmarshal([]byte(data), &root); err != nil {
		return nil, err
	}
	var renderers []map[string]any
	collectMapsByKey(root, "videoRenderer", &renderers)
	out := []CreatorVideo{}
	seen := map[string]bool{}
	for _, vr := range renderers {
		id := asString(vr["videoId"])
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		title := runsText(vr["title"])
		if title == "" {
			continue
		}
		creator := runsText(vr["ownerText"])
		thumb := largestThumbnail(vr["thumbnail"])
		out = append(out, CreatorVideo{ID: id, Platform: "youtube", URL: "https://www.youtube.com/watch?v=" + id, Title: title, Creator: creator, ThumbnailURL: thumb, DiscoveredAt: time.Now().UTC().Format(time.RFC3339), Mods: []CreatorMod{}})
		if len(out) >= 25 {
			break
		}
	}
	return out, nil
}

func collectMapsByKey(v any, key string, out *[]map[string]any) {
	switch x := v.(type) {
	case map[string]any:
		if child, ok := x[key].(map[string]any); ok {
			*out = append(*out, child)
		}
		for _, c := range x {
			collectMapsByKey(c, key, out)
		}
	case []any:
		for _, c := range x {
			collectMapsByKey(c, key, out)
		}
	}
}

func runsText(v any) string {
	m, _ := v.(map[string]any)
	if s := asString(m["simpleText"]); s != "" {
		return s
	}
	runs, _ := m["runs"].([]any)
	parts := []string{}
	for _, r := range runs {
		if rm, ok := r.(map[string]any); ok {
			if t := asString(rm["text"]); t != "" {
				parts = append(parts, t)
			}
		}
	}
	return strings.Join(parts, "")
}

func largestThumbnail(v any) string {
	m, _ := v.(map[string]any)
	arr, _ := m["thumbnails"].([]any)
	best := ""
	area := float64(0)
	for _, x := range arr {
		if mm, ok := x.(map[string]any); ok {
			u := asString(mm["url"])
			w := asFloat(mm["width"])
			h := asFloat(mm["height"])
			if u != "" && w*h >= area {
				best = u
				area = w * h
			}
		}
	}
	return best
}

func (a *App) analyzeCreatorVideo(ctx context.Context, rawURL string) (CreatorVideo, error) {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || u.Host == "" {
		return CreatorVideo{}, errors.New("invalid video URL")
	}
	host := strings.ToLower(strings.TrimPrefix(u.Host, "www."))
	switch {
	case strings.Contains(host, "youtube.com") || host == "youtu.be":
		return a.analyzeYouTube(ctx, rawURL)
	case strings.Contains(host, "tiktok.com") || strings.Contains(host, "tiktokv.com"):
		return a.analyzeTikTok(ctx, rawURL)
	default:
		return CreatorVideo{}, errors.New("Creator Picks currently analyzes YouTube and TikTok URLs")
	}
}

func (a *App) analyzeYouTube(ctx context.Context, rawURL string) (CreatorVideo, error) {
	body, err := a.getText(ctx, rawURL, map[string]string{"Accept-Language": "en-US,en;q=0.9"})
	if err != nil {
		return CreatorVideo{}, err
	}
	payload, err := extractAssignedJSON(body, "ytInitialPlayerResponse")
	if err != nil {
		return CreatorVideo{}, fmt.Errorf("YouTube player metadata: %w", err)
	}
	var root map[string]any
	if err := json.Unmarshal([]byte(payload), &root); err != nil {
		return CreatorVideo{}, err
	}
	vd, _ := root["videoDetails"].(map[string]any)
	id := asString(vd["videoId"])
	if id == "" {
		id = youtubeID(rawURL)
	}
	channelID := asString(vd["channelId"])
	creatorURL := ""
	if channelID != "" {
		creatorURL = "https://www.youtube.com/channel/" + channelID
	}
	duration, _ := strconv.ParseInt(asString(vd["lengthSeconds"]), 10, 64)
	video := CreatorVideo{ID: id, Platform: "youtube", URL: "https://www.youtube.com/watch?v=" + id, Title: asString(vd["title"]), Creator: asString(vd["author"]), CreatorURL: creatorURL, Description: asString(vd["shortDescription"]), ThumbnailURL: largestThumbnail(vd["thumbnail"]), DiscoveredAt: time.Now().UTC().Format(time.RFC3339), AnalyzedAt: time.Now().UTC().Format(time.RFC3339), ChannelID: channelID, DurationSeconds: duration, VideoKind: youtubeKindFromMetadata(asString(vd["title"]), asString(vd["shortDescription"]), duration), Mods: []CreatorMod{}}
	if micro := getMap(root, "microformat", "playerMicroformatRenderer"); micro != nil {
		video.PublishedAt = firstNonEmpty(asString(micro["publishDate"]), asString(micro["uploadDate"]))
	}
	if data, initialErr := extractAssignedJSON(body, "ytInitialData"); initialErr == nil {
		var initial any
		if json.Unmarshal([]byte(data), &initial) == nil {
			var owners []map[string]any
			collectMapsByKey(initial, "videoOwnerRenderer", &owners)
			if len(owners) > 0 {
				video.CreatorAvatarURL = largestThumbnail(owners[0]["thumbnail"])
				if video.Creator == "" {
					video.Creator = runsText(owners[0]["title"])
				}
			}
		}
	}
	segments, source := a.youtubeTranscript(ctx, root)
	if len(segments) == 0 {
		if captionSegments, captionSource, captionErr := a.ytDLPTranscript(ctx, video.URL); captionErr == nil && len(captionSegments) > 0 {
			segments, source = captionSegments, captionSource
		} else {
			if captionErr != nil {
				video.Warnings = append(video.Warnings, "Direct YouTube captions were unavailable; yt-dlp caption recovery also failed: "+captionErr.Error())
			}
			if localSegments, localSource, localErr := a.localCreatorTranscript(ctx, video.URL); localErr == nil && len(localSegments) > 0 {
				segments, source = localSegments, localSource
			} else if localErr != nil {
				video.Warnings = append(video.Warnings, "Local archival speech-to-text fallback could not complete: "+localErr.Error())
			}
		}
	}
	if video.VideoKind == "short" || len(segments) == 0 {
		if visualSegments, visualSource, visualErr := a.creatorVisualText(ctx, video.URL); visualErr == nil && len(visualSegments) > 0 {
			segments = mergeTranscriptSegments(segments, visualSegments)
			source = combineCreatorTranscriptSource(source, visualSource)
		} else if visualErr != nil {
			video.Warnings = append(video.Warnings, "Visual-text OCR did not complete for this Short/video: "+visualErr.Error())
		}
	}
	video.TranscriptAvailable = len(segments) > 0
	video.TranscriptSource = source
	video.TranscriptSegments = len(segments)
	for _, seg := range segments {
		video.TranscriptWords += len(strings.Fields(seg.Text))
	}
	if len(segments) > 0 {
		_ = a.saveCreatorTranscript(video.ID, source, segments)
	}
	video.Mods, video.UnresolvedMentions = a.extractCreatorMods(ctx, video, segments)
	if len(segments) == 0 {
		video.Warnings = append(video.Warnings, "Description links and timestamps were still analyzed even though no timed transcript was available.")
	}
	return video, nil
}

func (a *App) youtubeTranscript(ctx context.Context, root map[string]any) ([]TranscriptSegment, string) {
	caps := getMap(root, "captions", "playerCaptionsTracklistRenderer")
	if caps == nil {
		return nil, ""
	}
	tracks, _ := caps["captionTracks"].([]any)
	var chosen map[string]any
	for _, x := range tracks {
		m, _ := x.(map[string]any)
		if chosen == nil {
			chosen = m
		}
		lang := strings.ToLower(asString(m["languageCode"]))
		if strings.HasPrefix(lang, "en") {
			chosen = m
			break
		}
	}
	if chosen == nil {
		return nil, ""
	}
	base := html.UnescapeString(asString(chosen["baseUrl"]))
	if base == "" {
		return nil, ""
	}
	sep := "&"
	if !strings.Contains(base, "?") {
		sep = "?"
	}
	text, err := a.getText(ctx, base+sep+"fmt=json3", nil)
	if err != nil {
		return nil, ""
	}
	var j struct {
		Events []struct {
			TStartMs    int64 `json:"tStartMs"`
			DDurationMs int64 `json:"dDurationMs"`
			Segs        []struct {
				UTF8 string `json:"utf8"`
			} `json:"segs"`
		} `json:"events"`
	}
	if json.Unmarshal([]byte(text), &j) != nil {
		return parseTimedText(text), "youtube captions"
	}
	out := []TranscriptSegment{}
	for _, e := range j.Events {
		parts := []string{}
		for _, s := range e.Segs {
			parts = append(parts, s.UTF8)
		}
		t := strings.TrimSpace(strings.Join(parts, ""))
		if t != "" {
			out = append(out, TranscriptSegment{StartMS: e.TStartMs, EndMS: e.TStartMs + e.DDurationMs, Text: t})
		}
	}
	return out, "youtube captions"
}

func (a *App) analyzeTikTok(ctx context.Context, rawURL string) (CreatorVideo, error) {
	body, err := a.getText(ctx, rawURL, map[string]string{"Accept-Language": "en-US,en;q=0.9"})
	if err != nil {
		return CreatorVideo{}, err
	}
	payload, err := extractScriptJSONByID(body, "__UNIVERSAL_DATA_FOR_REHYDRATION__")
	if err != nil {
		return CreatorVideo{}, fmt.Errorf("TikTok metadata: %w", err)
	}
	var root map[string]any
	if json.Unmarshal([]byte(payload), &root) != nil {
		return CreatorVideo{}, errors.New("TikTok metadata JSON was unreadable")
	}
	def, _ := root["__DEFAULT_SCOPE__"].(map[string]any)
	detail := getMap(def, "webapp.video-detail", "itemInfo", "itemStruct")
	if detail == nil {
		return CreatorVideo{}, errors.New("TikTok video detail not found")
	}
	author, _ := detail["author"].(map[string]any)
	vi, _ := detail["video"].(map[string]any)
	id := asString(detail["id"])
	handle := asString(author["uniqueId"])
	creator := firstNonEmpty(asString(author["nickname"]), handle)
	duration, _ := strconv.ParseInt(asString(vi["duration"]), 10, 64)
	published := ""
	if created, err := strconv.ParseInt(asString(detail["createTime"]), 10, 64); err == nil && created > 0 {
		published = time.Unix(created, 0).UTC().Format(time.RFC3339)
	}
	video := CreatorVideo{ID: id, Platform: "tiktok", URL: rawURL, Title: asString(detail["desc"]), Description: asString(detail["desc"]), Creator: creator, CreatorURL: "https://www.tiktok.com/@" + handle, CreatorAvatarURL: firstNonEmpty(asString(author["avatarLarger"]), asString(author["avatarMedium"])), ThumbnailURL: firstNonEmpty(asString(vi["cover"]), asString(vi["originCover"])), PublishedAt: published, DiscoveredAt: time.Now().UTC().Format(time.RFC3339), AnalyzedAt: time.Now().UTC().Format(time.RFC3339), ChannelHandle: canonicalCreatorHandle(handle), VideoKind: "short", DurationSeconds: duration, Mods: []CreatorMod{}}
	segments := a.tiktokTranscript(ctx, detail)
	source := ""
	if len(segments) > 0 {
		source = "TikTok auto captions"
	} else if localSegments, localSource, localErr := a.localCreatorTranscript(ctx, video.URL); localErr == nil && len(localSegments) > 0 {
		segments, source = localSegments, localSource
	} else if localErr != nil {
		video.Warnings = append(video.Warnings, "No public TikTok caption track was exposed. Local speech-to-text fallback could not complete: "+localErr.Error())
	}
	if visualSegments, visualSource, visualErr := a.creatorVisualText(ctx, video.URL); visualErr == nil && len(visualSegments) > 0 {
		segments = mergeTranscriptSegments(segments, visualSegments)
		source = combineCreatorTranscriptSource(source, visualSource)
	} else if visualErr != nil {
		video.Warnings = append(video.Warnings, "Visual-text OCR did not complete, so text-only recommendations may be incomplete: "+visualErr.Error())
	}
	video.TranscriptAvailable = len(segments) > 0
	video.TranscriptSource = source
	video.TranscriptSegments = len(segments)
	for _, seg := range segments {
		video.TranscriptWords += len(strings.Fields(seg.Text))
	}
	if len(segments) > 0 {
		_ = a.saveCreatorTranscript(video.ID, source, segments)
	}
	video.Mods, video.UnresolvedMentions = a.extractCreatorMods(ctx, video, segments)
	if len(segments) == 0 {
		video.Warnings = append(video.Warnings, "Description/caption metadata was still analyzed even though no timed transcript was available.")
	}
	return video, nil
}

func (a *App) tiktokTranscript(ctx context.Context, detail map[string]any) []TranscriptSegment {
	vi, _ := detail["video"].(map[string]any)
	for _, key := range []string{"subtitleInfos", "subtitle_infos"} {
		arr, _ := vi[key].([]any)
		for _, x := range arr {
			m, _ := x.(map[string]any)
			lang := strings.ToLower(firstNonEmpty(asString(m["LanguageCodeName"]), asString(m["languageCodeName"]), asString(m["lang"])))
			if lang != "" && !strings.Contains(lang, "en") {
				continue
			}
			u := firstNonEmpty(asString(m["Url"]), asString(m["url"]))
			if u == "" {
				continue
			}
			text, err := a.getText(ctx, u, nil)
			if err != nil {
				continue
			}
			if seg := parseCaptionPayload(text); len(seg) > 0 {
				return seg
			}
		}
	}
	// Some TikTok responses expose an auto caption info block under interaction stickers.
	stickers, _ := detail["interactionStickers"].([]any)
	if stickers == nil {
		stickers, _ = detail["interaction_stickers"].([]any)
	}
	for _, x := range stickers {
		m, _ := x.(map[string]any)
		var infos []any
		collectArraysByKey(m, "auto_captions", &infos)
		for _, raw := range infos {
			cm, _ := raw.(map[string]any)
			u := deepFirstURL(cm)
			if u == "" {
				continue
			}
			text, err := a.getText(ctx, u, nil)
			if err == nil {
				if seg := parseCaptionPayload(text); len(seg) > 0 {
					return seg
				}
			}
		}
	}
	return nil
}

func collectArraysByKey(v any, key string, out *[]any) {
	switch x := v.(type) {
	case map[string]any:
		for k, c := range x {
			if k == key {
				if a, ok := c.([]any); ok {
					*out = append(*out, a...)
				}
			}
			collectArraysByKey(c, key, out)
		}
	case []any:
		for _, c := range x {
			collectArraysByKey(c, key, out)
		}
	}
}
func deepFirstURL(v any) string {
	switch x := v.(type) {
	case map[string]any:
		for k, c := range x {
			if strings.EqualFold(k, "url") {
				if s := asString(c); strings.HasPrefix(s, "http") {
					return s
				}
			}
			if s := deepFirstURL(c); s != "" {
				return s
			}
		}
	case []any:
		for _, c := range x {
			if s := deepFirstURL(c); s != "" {
				return s
			}
		}
	}
	return ""
}

func parseCaptionPayload(text string) []TranscriptSegment {
	var j struct {
		Utterances []struct {
			StartTime int64  `json:"start_time"`
			EndTime   int64  `json:"end_time"`
			Text      string `json:"text"`
		} `json:"utterances"`
	}
	if json.Unmarshal([]byte(text), &j) == nil && len(j.Utterances) > 0 {
		out := []TranscriptSegment{}
		for _, u := range j.Utterances {
			if strings.TrimSpace(u.Text) != "" {
				out = append(out, TranscriptSegment{StartMS: u.StartTime, EndMS: u.EndTime, Text: u.Text})
			}
		}
		return out
	}
	return parseTimedText(text)
}

func parseTimedText(text string) []TranscriptSegment {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	out := []TranscriptSegment{}
	timeRE := regexp.MustCompile(`(?i)(\d{1,2}:)?\d{1,2}:\d{2}[.,]\d{3}\s*-->\s*(\d{1,2}:)?\d{1,2}:\d{2}[.,]\d{3}`)
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if !timeRE.MatchString(line) {
			continue
		}
		parts := strings.Split(line, "-->")
		start := parseTimestamp(strings.TrimSpace(parts[0]))
		end := parseTimestamp(strings.Fields(strings.TrimSpace(parts[1]))[0])
		txt := []string{}
		for j := i + 1; j < len(lines) && strings.TrimSpace(lines[j]) != ""; j++ {
			txt = append(txt, cleanHTMLText(lines[j]))
			i = j
		}
		joined := strings.TrimSpace(strings.Join(txt, " "))
		if joined != "" {
			out = append(out, TranscriptSegment{StartMS: start * 1000, EndMS: end * 1000, Text: joined})
		}
	}
	return out
}

func (a *App) extractCreatorMods(ctx context.Context, video CreatorVideo, segments []TranscriptSegment) ([]CreatorMod, []CreatorMention) {
	mods := []CreatorMod{}
	unresolved := []CreatorMention{}
	seenResolved := map[string]int{}
	seenResolvedNames := map[string]bool{}
	seenMentions := map[string]bool{}

	addResolved := func(cm CreatorMod) {
		nameKey := normalizeProjectTitle(cm.Name)
		if nameKey == "" {
			return
		}
		key := nameKey + ":" + strings.ToLower(cm.Provider)
		if idx, ok := seenResolved[key]; ok {
			old := &mods[idx]
			old.SourceKinds = uniqueStringsPreserve(append(old.SourceKinds, cm.SourceKinds...))
			if old.DescriptionContext == "" {
				old.DescriptionContext = cm.DescriptionContext
			}
			if old.TranscriptContext == "" {
				old.TranscriptContext = cm.TranscriptContext
			}
			if old.ProjectSummary == "" {
				old.ProjectSummary = cm.ProjectSummary
			}
			if cm.Confidence > old.Confidence {
				old.Confidence = cm.Confidence
			}
			if old.Evidence == "" {
				old.Evidence = cm.Evidence
			} else if cm.Evidence != "" && !strings.Contains(old.Evidence, cm.Evidence) {
				old.Evidence += " | " + cm.Evidence
			}
			return
		}
		seenResolved[key] = len(mods)
		seenResolvedNames[nameKey] = true
		mods = append(mods, cm)
	}

	// Creator descriptions are a first-class evidence source. Preserve nearby lines so a
	// recommendation keeps the creator's own explanation, not just the project URL.
	lines := strings.Split(strings.ReplaceAll(video.Description, "\r\n", "\n"), "\n")
	var inheritedTS int64
	for i, line := range lines {
		lineTS, hasTS := timestampFromLineOK(line)
		if hasTS {
			inheritedTS = lineTS
		}
		ts := inheritedTS
		contextText := creatorDescriptionContext(lines, i)
		links := extractURLs(line)
		for _, link := range links {
			if cm, ok := creatorModFromURL(link, ts, video); ok {
				cm.DescriptionContext = contextText
				cm.SourceKinds = uniqueStringsPreserve(append(cm.SourceKinds, "description"))
				cm = a.enrichDirectCreatorMod(ctx, cm)
				addResolved(cm)
			}
		}
		candidate := descriptionModName(line)
		if candidate == "" && hasTS {
			candidate = timestampedDescriptionCandidate(line)
		}
		if candidate != "" && len(links) == 0 {
			if cm, ok := a.resolveCreatorCandidate(ctx, candidate, ts, video, "video description"); ok {
				cm.SourceKinds = []string{"description-name"}
				cm.DescriptionContext = contextText
				addResolved(cm)
			} else {
				key := normalizeProjectTitle(candidate)
				if key != "" && !seenMentions[key] {
					seenMentions[key] = true
					unresolved = append(unresolved, CreatorMention{Name: candidate, Timestamp: formatTimestamp(ts), TimestampS: ts, VideoLink: timestampVideoLink(video, ts), Evidence: "unresolved name from video description: " + truncate(contextText, 220)})
				}
			}
		}
	}

	// Single-mod spotlights often put the exact project name only in the title. Resolve title
	// candidates too, but keep them lower confidence than direct creator links.
	titleCandidates := append(transcriptModNames(video.Title), descriptionModName(video.Title))
	for _, candidate := range uniqueStringsPreserve(titleCandidates) {
		key := normalizeProjectTitle(candidate)
		if key == "" || seenResolvedNames[key] {
			continue
		}
		if cm, ok := a.resolveCreatorCandidate(ctx, candidate, 0, video, "video title: "+truncate(video.Title, 180)); ok {
			cm.SourceKinds = []string{"video-title"}
			cm.Confidence = mathMin(cm.Confidence, .88)
			addResolved(cm)
		}
	}

	// Transcripts can contain dozens or hundreds of mods. Search individual cues plus rolling
	// three-cue context so names split across subtitle boundaries are still recoverable. Keep every
	// distinct candidate; there is deliberately no recommendation-count cap.
	type transcriptCandidate struct {
		Name     string
		StartS   int64
		Evidence string
		Context  string
		Source   string
	}
	candidates := []transcriptCandidate{}
	candidateIndex := map[string]int{}
	for i, seg := range segments {
		ctxStart := i - 1
		if ctxStart < 0 {
			ctxStart = 0
		}
		ctxEnd := i + 1
		if ctxEnd >= len(segments) {
			ctxEnd = len(segments) - 1
		}
		parts := []string{}
		for j := ctxStart; j <= ctxEnd; j++ {
			parts = append(parts, segments[j].Text)
		}
		contextText := strings.TrimSpace(strings.Join(parts, " "))
		searchTexts := []string{seg.Text, contextText}
		for _, searchText := range searchTexts {
			for _, candidate := range transcriptModNames(searchText) {
				key := normalizeProjectTitle(candidate)
				if key == "" || seenResolvedNames[key] {
					continue
				}
				if idx, exists := candidateIndex[key]; exists {
					// Preserve the richer context if a later rolling window contains more detail.
					if len(contextText) > len(candidates[idx].Context) {
						candidates[idx].Context = truncate(contextText, 500)
					}
					continue
				}
				candidateIndex[key] = len(candidates)
				candidates = append(candidates, transcriptCandidate{Name: candidate, StartS: seg.StartMS / 1000, Evidence: "transcript: " + truncate(seg.Text, 200), Context: truncate(contextText, 500), Source: "transcript"})
			}
		}
		if seg.Source == "visual-ocr" {
			for _, candidate := range visualTextModNames(seg.Text) {
				key := normalizeProjectTitle(candidate)
				if key == "" || seenResolvedNames[key] {
					continue
				}
				if idx, exists := candidateIndex[key]; exists {
					if candidates[idx].Source != "visual-ocr" {
						candidates[idx].Source = "visual-ocr"
						candidates[idx].Evidence = "visual OCR: " + truncate(seg.Text, 220)
					}
					continue
				}
				candidateIndex[key] = len(candidates)
				candidates = append(candidates, transcriptCandidate{Name: candidate, StartS: seg.StartMS / 1000, Evidence: "visual OCR: " + truncate(seg.Text, 220), Context: truncate(contextText, 500), Source: "visual-ocr"})
			}
		}
	}
	for _, candidate := range candidates {
		select {
		case <-ctx.Done():
			return mods, unresolved
		default:
		}
		key0 := normalizeProjectTitle(candidate.Name)
		if seenResolvedNames[key0] {
			continue
		}
		if cm, ok := a.resolveCreatorCandidate(ctx, candidate.Name, candidate.StartS, video, candidate.Evidence); ok {
			cm.SourceKinds = []string{firstNonEmpty(candidate.Source, "transcript")}
			cm.TranscriptContext = candidate.Context
			addResolved(cm)
		} else if !seenMentions[key0] {
			seenMentions[key0] = true
			unresolved = append(unresolved, CreatorMention{Name: candidate.Name, Timestamp: formatTimestamp(candidate.StartS), TimestampS: candidate.StartS, VideoLink: timestampVideoLink(video, candidate.StartS), Evidence: candidate.Evidence + " | context: " + candidate.Context})
		}
	}
	sort.SliceStable(mods, func(i, j int) bool { return mods[i].TimestampS < mods[j].TimestampS })
	sort.SliceStable(unresolved, func(i, j int) bool { return unresolved[i].TimestampS < unresolved[j].TimestampS })
	return mods, unresolved
}

func creatorDescriptionContext(lines []string, at int) string {
	parts := []string{}
	for i := at - 1; i <= at+1; i++ {
		if i < 0 || i >= len(lines) {
			continue
		}
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		parts = append(parts, line)
	}
	return truncate(strings.Join(parts, " | "), 500)
}

func timestampedDescriptionCandidate(line string) string {
	clean := regexp.MustCompile(`^\s*(?:(?:\d{1,2}:)?\d{1,2}:\d{2})\s*[-–—:|]*\s*`).ReplaceAllString(strings.TrimSpace(line), "")
	clean = regexp.MustCompile(`https?://\S+`).ReplaceAllString(clean, "")
	clean = strings.Trim(clean, " \t-*•#–—:|[]()")
	if clean == "" || len([]rune(clean)) > 90 || len(strings.Fields(clean)) > 10 {
		return ""
	}
	low := strings.ToLower(clean)
	for _, noise := range []string{"intro", "sponsor", "outro", "chapter", "conclusion", "thanks", "subscribe", "settings", "installation", "download links"} {
		if low == noise || strings.HasPrefix(low, noise+" ") {
			return ""
		}
	}
	return clean
}

func creatorModFromURL(raw string, ts int64, video CreatorVideo) (CreatorMod, bool) {
	u, err := url.Parse(strings.TrimRight(raw, ".,);]"))
	if err != nil {
		return CreatorMod{}, false
	}
	host := strings.ToLower(strings.TrimPrefix(u.Host, "www."))
	provider := ""
	id := ""
	name := ""
	switch {
	case strings.Contains(host, "modrinth.com"):
		provider = "modrinth"
		parts := splitPath(u.Path)
		if len(parts) >= 2 {
			id = parts[len(parts)-1]
			name = humanizeRepoName(id)
		}
	case strings.Contains(host, "curseforge.com"):
		provider = "curseforge"
		parts := splitPath(u.Path)
		if len(parts) > 0 {
			id = parts[len(parts)-1]
			name = humanizeRepoName(id)
		}
	case strings.Contains(host, "github.com"):
		provider = "github"
		parts := splitPath(u.Path)
		if len(parts) >= 2 {
			id = parts[0] + "/" + parts[1]
			name = humanizeRepoName(parts[1])
		}
	case strings.Contains(host, "planetminecraft.com"):
		provider = "planetminecraft"
		parts := splitPath(u.Path)
		if len(parts) > 0 {
			name = humanizeRepoName(parts[len(parts)-1])
			id = u.Path
		}
	case strings.Contains(host, "mcpedl.com"):
		provider = "mcpedl"
		parts := splitPath(u.Path)
		if len(parts) > 0 {
			name = humanizeRepoName(parts[len(parts)-1])
			id = u.Path
		}
	case strings.Contains(host, "hangar.papermc.io"):
		provider = "hangar"
		parts := splitPath(u.Path)
		if len(parts) > 0 {
			id = parts[len(parts)-1]
			name = humanizeRepoName(id)
		}
	case strings.Contains(host, "spigotmc.org"):
		provider = "spigot"
		parts := splitPath(u.Path)
		if len(parts) > 0 {
			id = strings.TrimSuffix(parts[len(parts)-1], ".")
			name = humanizeRepoName(strings.TrimSuffix(id, "."))
		}
	case strings.Contains(host, "dev.bukkit.org"):
		provider = "bukkitdev"
		parts := splitPath(u.Path)
		if len(parts) > 0 {
			id = parts[len(parts)-1]
			name = humanizeRepoName(id)
		}
	case strings.Contains(host, "builtbybit.com"):
		provider = "builtbybit"
		parts := splitPath(u.Path)
		if len(parts) > 0 {
			id = strings.TrimSuffix(parts[len(parts)-1], ".")
			name = humanizeRepoName(id)
		}
	case strings.Contains(host, "polymart.org"):
		provider = "polymart"
		parts := splitPath(u.Path)
		if len(parts) > 0 {
			id = parts[len(parts)-1]
			name = humanizeRepoName(id)
		}
	case strings.Contains(host, "ore.spongepowered.org"):
		provider = "spongeore"
		parts := splitPath(u.Path)
		if len(parts) > 0 {
			id = parts[len(parts)-1]
			name = humanizeRepoName(id)
		}
	case strings.Contains(host, "smithed.dev"):
		provider = "smithed"
		parts := splitPath(u.Path)
		if len(parts) > 0 {
			id = parts[len(parts)-1]
			name = humanizeRepoName(id)
		}
	case strings.Contains(host, "moddb.com"):
		provider = "moddb"
		parts := splitPath(u.Path)
		if len(parts) > 0 {
			id = u.Path
			name = humanizeRepoName(parts[len(parts)-1])
		}
	case strings.Contains(host, "atlauncher.com"):
		provider = "atlauncher"
		parts := splitPath(u.Path)
		if len(parts) > 0 {
			id = parts[len(parts)-1]
			name = humanizeRepoName(id)
		}
	case strings.Contains(host, "technicpack.net"):
		provider = "technic"
		parts := splitPath(u.Path)
		if len(parts) > 0 {
			id = parts[len(parts)-1]
			name = humanizeRepoName(id)
		}
	case strings.Contains(host, "feed-the-beast.com"), strings.Contains(host, "modpacks.ch"):
		provider = "ftb"
		parts := splitPath(u.Path)
		if len(parts) > 0 {
			id = parts[len(parts)-1]
			name = humanizeRepoName(id)
		}
	case strings.Contains(host, "nexusmods.com"):
		provider = "nexusmods"
		parts := splitPath(u.Path)
		if len(parts) > 0 {
			id = u.Path
			name = humanizeRepoName(parts[len(parts)-1])
		}
	case strings.Contains(host, "vanillatweaks.net"):
		provider = "vanillatweaks"
		id = strings.Trim(u.Path, "/")
		name = "Vanilla Tweaks"
	case strings.Contains(host, "minecraftmaps.com"):
		provider = "minecraftmaps"
		parts := splitPath(u.Path)
		if len(parts) > 0 {
			id = u.Path
			name = humanizeRepoName(parts[len(parts)-1])
		}
	case strings.Contains(host, "resourcepack.net"):
		provider = "resourcepacknet"
		parts := splitPath(u.Path)
		if len(parts) > 0 {
			id = u.Path
			name = humanizeRepoName(parts[len(parts)-1])
		}
	case strings.Contains(host, "texture-packs.com"):
		provider = "texturepacks"
		parts := splitPath(u.Path)
		if len(parts) > 0 {
			id = u.Path
			name = humanizeRepoName(parts[len(parts)-1])
		}
	case strings.Contains(host, "mcreator.net"):
		provider = "mcreator"
		parts := splitPath(u.Path)
		if len(parts) > 0 {
			id = u.Path
			name = humanizeRepoName(parts[len(parts)-1])
		}
	case strings.Contains(host, "shaderpacks.com"):
		provider = "shaderpackscom"
		parts := splitPath(u.Path)
		if len(parts) > 0 {
			id = u.Path
			name = humanizeRepoName(parts[len(parts)-1])
		}
	case strings.Contains(host, "shaderpacks.net"):
		provider = "shaderpacksnet"
		parts := splitPath(u.Path)
		if len(parts) > 0 {
			id = u.Path
			name = humanizeRepoName(parts[len(parts)-1])
		}
	case strings.Contains(host, "minecraftshader.com"):
		provider = "minecraftshader"
		parts := splitPath(u.Path)
		if len(parts) > 0 {
			id = u.Path
			name = humanizeRepoName(parts[len(parts)-1])
		}
	case strings.Contains(host, "minecraftskins.com"):
		provider = "skindex"
		parts := splitPath(u.Path)
		if len(parts) > 0 {
			id = u.Path
			name = humanizeRepoName(parts[len(parts)-1])
		}
	case strings.Contains(host, "minecrafthub.io"):
		provider = "minecrafthub"
		parts := splitPath(u.Path)
		if len(parts) > 0 {
			id = u.Path
			name = humanizeRepoName(parts[len(parts)-1])
		}
	case strings.Contains(host, "minecraft.net") && strings.Contains(strings.ToLower(u.Path), "marketplace"):
		provider = "marketplace"
		parts := splitPath(u.Path)
		if len(parts) > 0 {
			id = u.Path
			name = humanizeRepoName(parts[len(parts)-1])
		}
	default:
		return CreatorMod{}, false
	}
	if name == "" {
		return CreatorMod{}, false
	}
	return CreatorMod{Name: name, Provider: provider, ProjectID: id, URL: u.String(), Timestamp: formatTimestamp(ts), TimestampS: ts, VideoLink: timestampVideoLink(video, ts), Evidence: "direct link in video description", SourceKinds: []string{"description-link"}, Confidence: 1}, true
}

func (a *App) enrichDirectCreatorMod(ctx context.Context, mod CreatorMod) CreatorMod {
	if mod.Provider == "" || mod.Name == "" {
		return mod
	}
	a.mu.RLock()
	s := a.settings
	a.mu.RUnlock()
	resp := a.searchProviders(ctx, providerSearchOptions{Query: firstNonEmpty(mod.ProjectID, mod.Name), GameVersion: s.GameVersion, Loader: s.Loader, ProjectType: "all", Limit: 5, Sources: []string{mod.Provider}})
	if len(resp.Results) == 0 && mod.ProjectID != mod.Name {
		resp = a.searchProviders(ctx, providerSearchOptions{Query: mod.Name, GameVersion: s.GameVersion, Loader: s.Loader, ProjectType: "all", Limit: 5, Sources: []string{mod.Provider}})
	}
	if len(resp.Results) == 0 {
		return mod
	}
	best := resp.Results[0]
	if titleSimilarity(mod.Name, best.Title) < .18 && !strings.Contains(strings.ToLower(best.PageURL), strings.ToLower(strings.Trim(mod.ProjectID, "/"))) {
		return mod
	}
	mod.Name = firstNonEmpty(best.Title, mod.Name)
	mod.ProjectID = firstNonEmpty(best.ID, mod.ProjectID)
	mod.ProjectType = best.ProjectType
	mod.Author = best.Author
	mod.IconURL = best.IconURL
	mod.ProjectSummary = best.Summary
	if best.PageURL != "" {
		mod.URL = best.PageURL
	}
	return mod
}

func (a *App) resolveCreatorCandidate(ctx context.Context, name string, ts int64, video CreatorVideo, evidence string) (CreatorMod, bool) {
	name = strings.TrimSpace(name)
	if len(name) < 2 || len(name) > 90 {
		return CreatorMod{}, false
	}
	a.mu.RLock()
	s := a.settings
	a.mu.RUnlock()
	resp := a.searchProviders(ctx, providerSearchOptions{Query: name, GameVersion: s.GameVersion, Loader: s.Loader, ProjectType: "mod", Limit: 6, Sources: a.enabledProvidersForType("mod")})
	if len(resp.Results) == 0 {
		return CreatorMod{}, false
	}
	best := resp.Results[0]
	sim := titleSimilarity(name, best.Title)
	if sim < .20 && !strings.Contains(strings.ToLower(best.Title), strings.ToLower(name)) {
		return CreatorMod{}, false
	}
	confidence := mathMin(.96, .55+sim*.4)
	return CreatorMod{Name: best.Title, Provider: best.Provider, ProjectID: best.ID, ProjectType: best.ProjectType, Author: best.Author, IconURL: best.IconURL, URL: best.PageURL, Timestamp: formatTimestamp(ts), TimestampS: ts, VideoLink: timestampVideoLink(video, ts), Evidence: evidence, SourceKinds: []string{"transcript"}, ProjectSummary: best.Summary, Confidence: confidence}, true
}

func descriptionModName(line string) string {
	clean := strings.TrimSpace(regexp.MustCompile(`https?://\S+`).ReplaceAllString(line, ""))
	clean = regexp.MustCompile(`^\s*(?:\d{1,2}:)?\d{1,2}:\d{2}\s*[-–—:|]*\s*`).ReplaceAllString(clean, "")
	clean = regexp.MustCompile(`^\s*[-*•#\d.)]+\s*`).ReplaceAllString(clean, "")
	low := strings.ToLower(clean)
	if strings.Contains(low, " mod") || strings.Contains(low, "addon") || strings.Contains(low, "resource pack") || strings.Contains(low, "shader") {
		clean = regexp.MustCompile(`(?i)\s+(?:mod|addon|resource\s*pack|shader).*$`).ReplaceAllString(clean, "")
		clean = strings.Trim(clean, " -–—:|[]()")
		if len(strings.Fields(clean)) <= 8 {
			return clean
		}
	}
	return ""
}

func transcriptModNames(text string) []string {
	out := []string{}
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:mod|addon|pack)\s+(?:called|named)\s+([A-Za-z0-9][A-Za-z0-9 '&:+._-]{1,55})`),
		regexp.MustCompile(`(?i)([A-Za-z0-9][A-Za-z0-9 '&:+._-]{1,45})\s+(?:mod|addon|resource pack|shader)\b`),
		regexp.MustCompile(`(?i)(?:i(?:'m| am)?\s+(?:using|running|playing with)|we(?:'re| are)?\s+(?:using|running|playing with)|install(?:ed|ing)?|try(?:ing)?|recommend(?:ing|ed)?|suggest(?:ing|ed)?|check(?:ing)?\s+out|look(?:ing)?\s+at|show(?:ing)?\s+off|feature(?:d|ing)?|first is|next is|another is|last is|finally)\s+([A-Za-z0-9][A-Za-z0-9 '&:+._-]{1,55})`),
		regexp.MustCompile(`(?i)(?:called|named)\s+([A-Za-z0-9][A-Za-z0-9 '&:+._-]{1,55})`),
		regexp.MustCompile(`(?i)(?:first|next|another|second|third|last|final)\s+(?:mod|addon|pack|shader)\s+(?:is|is called|is named|we have|i have)?\s*([A-Za-z0-9][A-Za-z0-9 '&:+._-]{1,55})`),
		regexp.MustCompile(`(?i)(?:this|the)\s+(?:mod|addon|pack|shader)\s+([A-Za-z0-9][A-Za-z0-9 '&:+._-]{1,55})\s+(?:adds?|changes?|overhauls?|improves?|lets?|gives?|makes?|brings?)\b`),
		regexp.MustCompile(`(?i)([A-Za-z0-9][A-Za-z0-9 '&:+._-]{1,45})\s+(?:adds?|changes?|overhauls?|improves?|lets you|gives you|makes|brings)\b`),
	}
	for _, re := range patterns {
		for _, m := range re.FindAllStringSubmatch(text, -1) {
			if len(m) <= 1 {
				continue
			}
			if n := cleanTranscriptCandidate(m[1]); n != "" {
				out = append(out, n)
			}
		}
	}
	return uniqueStringsPreserve(out)
}

func visualTextModNames(text string) []string {
	out := []string{}
	for _, raw := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		explicit := regexp.MustCompile(`(?i)^(?:mod|addon|shader|resource\s*pack|texture\s*pack)\s*[:\-–—]\s*`).FindStringIndex(line) != nil
		listed := regexp.MustCompile(`^\s*(?:[#*•-]\s*|\d{1,2}[.)\-:]\s+)`).FindStringIndex(line) != nil
		line = regexp.MustCompile(`(?i)^(?:mod|addon|shader|resource\s*pack|texture\s*pack)\s*[:\-–—]\s*`).ReplaceAllString(line, "")
		line = regexp.MustCompile(`^\s*(?:[#*•-]\s*|\d{1,2}[.)\-:]\s+)`).ReplaceAllString(line, "")
		line = cleanTranscriptCandidate(line)
		if line == "" || strings.Contains(line, "http://") || strings.Contains(line, "https://") || strings.HasPrefix(line, "@") {
			continue
		}
		low := strings.ToLower(line)
		if strings.Contains(low, "minecraft") && strings.Contains(low, "mod") && (strings.Contains(low, "need") || strings.Contains(low, "best") || strings.Contains(low, "top")) {
			continue
		}
		noise := false
		for _, phrase := range []string{
			"minecraft mods", "mods you need", "best mods", "top mods", "link in bio", "links in bio",
			"follow for", "follow me", "part 1", "part 2", "part 3", "original sound", "more videos",
			"comments", "comment", "likes", "share", "save", "subscribe", "download", "curseforge", "modrinth",
		} {
			if low == phrase || strings.HasPrefix(low, phrase+" ") {
				noise = true
				break
			}
		}
		if noise || low == "minecraft" || low == "tiktok" {
			continue
		}
		words := strings.Fields(line)
		if len(words) == 0 || len(words) > 8 || len([]rune(line)) > 80 {
			continue
		}
		// Explicit labels and numbered/bulleted lists are strong visual recommendation
		// evidence. Bare name-shaped overlays are retained as lower-confidence candidates;
		// provider resolution still has to validate them before they become recommendations.
		if explicit || listed || (len(words) <= 5 && regexp.MustCompile(`[A-Za-z]`).MatchString(line) && !strings.ContainsAny(line, "!?")) {
			out = append(out, line)
		}
	}
	return uniqueStringsPreserve(out)
}

func cleanTranscriptCandidate(name string) string {
	name = strings.Trim(strings.TrimSpace(name), " \t\n\r-–—:|,.;!?[](){}\"")
	low := strings.ToLower(name)
	// Greedy phrase captures can include the spoken lead-in when a later "mod" token is
	// what caused the match. Keep the name-shaped suffix rather than the narration.
	for _, marker := range []string{"i'm using ", "i am using ", "we're using ", "we are using ", "playing with ", "running ", "installed ", "install ", "trying ", "recommend ", "recommending ", "first is ", "next is "} {
		if idx := strings.LastIndex(low, marker); idx >= 0 {
			name = name[idx+len(marker):]
			low = strings.ToLower(name)
		}
	}
	for _, marker := range []string{" mod ", " addon ", " resource pack ", " shader ", " and then ", " and next ", " but ", " because ", " which ", " that ", " today ", " here "} {
		if idx := strings.Index(low, marker); idx > 0 {
			name = name[:idx]
			low = strings.ToLower(name)
		}
	}
	name = regexp.MustCompile(`(?i)^(?:this|called|named)\s+`).ReplaceAllString(strings.TrimSpace(name), "")
	name = strings.Trim(name, " \t\n\r-–—:|,.;!?[](){}\"")
	if len(name) < 2 || len([]rune(name)) > 90 || len(strings.Fields(name)) > 8 {
		return ""
	}
	return name
}

func timestampFromLine(line string) int64 {
	sec, _ := timestampFromLineOK(line)
	return sec
}

func timestampFromLineOK(line string) (int64, bool) { // regexp2-style lookaround is not supported by Go, so use a simpler matcher below.
	re := regexp.MustCompile(`(?:^|\s)(?:(\d{1,2}):)?(\d{1,2}):(\d{2})(?:\s|$|[-–—:|])`)
	m := re.FindStringSubmatch(line)
	if len(m) == 0 {
		return 0, false
	}
	h, _ := strconv.ParseInt(m[1], 10, 64)
	min, _ := strconv.ParseInt(m[2], 10, 64)
	sec, _ := strconv.ParseInt(m[3], 10, 64)
	return h*3600 + min*60 + sec, true
}
func parseTimestamp(s string) int64 {
	s = strings.ReplaceAll(s, ",", ".")
	parts := strings.Split(strings.Split(s, ".")[0], ":")
	var v int64
	for _, p := range parts {
		n, _ := strconv.ParseInt(p, 10, 64)
		v = v*60 + n
	}
	return v
}
func formatTimestamp(sec int64) string {
	if sec < 0 {
		sec = 0
	}
	h := sec / 3600
	m := (sec % 3600) / 60
	s := sec % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}
func timestampVideoLink(v CreatorVideo, sec int64) string {
	if v.Platform == "youtube" {
		return fmt.Sprintf("https://www.youtube.com/watch?v=%s&t=%ds", url.QueryEscape(v.ID), sec)
	}
	return v.URL
}
func extractURLs(s string) []string {
	return regexp.MustCompile(`https?://[^\s<>"']+`).FindAllString(s, -1)
}
func splitPath(p string) []string {
	raw := strings.Split(strings.Trim(p, "/"), "/")
	out := []string{}
	for _, x := range raw {
		if x != "" {
			out = append(out, x)
		}
	}
	return out
}
func youtubeID(raw string) string {
	u, _ := url.Parse(raw)
	if u == nil {
		return ""
	}
	if strings.Contains(u.Host, "youtu.be") {
		parts := splitPath(u.Path)
		if len(parts) > 0 {
			return strings.Trim(parts[0], "/")
		}
		return ""
	}
	return u.Query().Get("v")
}
func truncate(s string, n int) string {
	r := []rune(strings.TrimSpace(s))
	if len(r) <= n {
		return string(r)
	}
	return string(r[:n]) + "…"
}

func (a *App) processCreatorQueue(ctx context.Context, max int) error {
	if max <= 0 {
		max = 1
	}
	a.dataMu.RLock()
	tracked := map[string]bool{}
	for _, ch := range a.creatorChannels {
		tracked[creatorChannelKey(ch)] = true
	}
	queued := make([]CreatorVideo, 0, max)
	for _, video := range a.creatorVideos {
		// Complete tracked-channel archives have their own queue with retry/cooldown
		// semantics. Keep generic search discovery from racing the same video.
		if tracked[creatorVideoChannelKey(video)] {
			continue
		}
		if video.AnalyzedAt == "" && video.URL != "" {
			queued = append(queued, video)
			if len(queued) >= max {
				break
			}
		}
	}
	a.dataMu.RUnlock()
	if len(queued) == 0 {
		return nil
	}
	var failures []string
	processed := 0
	for _, queuedVideo := range queued {
		select {
		case <-ctx.Done():
			if processed == 0 {
				return ctx.Err()
			}
			return nil
		default:
		}
		video, err := a.analyzeCreatorVideo(ctx, queuedVideo.URL)
		if err != nil {
			failures = append(failures, queuedVideo.Platform+":"+queuedVideo.ID+": "+err.Error())
			continue
		}
		a.upsertCreatorVideo(video)
		processed++
	}
	if processed > 0 {
		_ = a.saveCreatorVideos()
	}
	if processed == 0 && len(failures) > 0 {
		return errors.New(strings.Join(failures, "; "))
	}
	return nil
}

func (a *App) upsertCreatorVideo(video CreatorVideo) {
	a.dataMu.Lock()
	defer a.dataMu.Unlock()
	if video.DiscoveredAt == "" {
		video.DiscoveredAt = time.Now().UTC().Format(time.RFC3339)
	}
	for i := range a.creatorVideos {
		if a.creatorVideos[i].Platform == video.Platform && a.creatorVideos[i].ID == video.ID {
			old := a.creatorVideos[i]
			if video.Description == "" {
				video.Description = old.Description
			}
			if video.ThumbnailURL == "" {
				video.ThumbnailURL = old.ThumbnailURL
			}
			if video.Creator == "" {
				video.Creator = old.Creator
			}
			if video.ChannelID == "" {
				video.ChannelID = old.ChannelID
			}
			if video.ChannelHandle == "" {
				video.ChannelHandle = old.ChannelHandle
			}
			if video.VideoKind == "" {
				video.VideoKind = old.VideoKind
			}
			if video.DurationSeconds == 0 {
				video.DurationSeconds = old.DurationSeconds
			}
			if video.PublishedAt == "" {
				video.PublishedAt = old.PublishedAt
			}
			if video.AnalyzedAt == "" {
				video.AnalyzedAt = old.AnalyzedAt
				video.Mods = old.Mods
				video.TranscriptAvailable = old.TranscriptAvailable
				video.TranscriptSource = old.TranscriptSource
				video.TranscriptSegments = old.TranscriptSegments
				video.TranscriptWords = old.TranscriptWords
			}
			if video.AnalysisAttempts == 0 {
				video.AnalysisAttempts = old.AnalysisAttempts
			}
			if video.AnalysisError == "" && video.AnalyzedAt == "" {
				video.AnalysisError = old.AnalysisError
			}
			if video.LastAnalysisAttempt == "" {
				video.LastAnalysisAttempt = old.LastAnalysisAttempt
			}
			a.creatorVideos[i] = video
			return
		}
	}
	a.creatorVideos = append(a.creatorVideos, video)
}

func extractAssignedJSON(body, name string) (string, error) {
	markers := []string{"var " + name + " = ", name + " = ", "window[\"" + name + "\"] = "}
	for _, marker := range markers {
		if i := strings.Index(body, marker); i >= 0 {
			start := strings.Index(body[i+len(marker):], "{")
			if start < 0 {
				continue
			}
			start += i + len(marker)
			if raw, ok := balancedJSON(body, start); ok {
				return raw, nil
			}
		}
	}
	return "", fmt.Errorf("%s not found", name)
}
func extractScriptJSONByID(body, id string) (string, error) {
	re := regexp.MustCompile(`(?is)<script[^>]+id=["']` + regexp.QuoteMeta(id) + `["'][^>]*>(.*?)</script>`)
	m := re.FindStringSubmatch(body)
	if len(m) < 2 {
		return "", fmt.Errorf("script %s not found", id)
	}
	return html.UnescapeString(strings.TrimSpace(m[1])), nil
}
func balancedJSON(s string, start int) (string, bool) {
	if start < 0 || start >= len(s) || s[start] != '{' {
		return "", false
	}
	depth := 0
	inString := false
	escp := false
	for i := start; i < len(s); i++ {
		c := s[i]
		if inString {
			if escp {
				escp = false
				continue
			}
			if c == '\\' {
				escp = true
				continue
			}
			if c == '"' {
				inString = false
			}
			continue
		}
		if c == '"' {
			inString = true
			continue
		}
		if c == '{' {
			depth++
		}
		if c == '}' {
			depth--
			if depth == 0 {
				return s[start : i+1], true
			}
		}
	}
	return "", false
}
func getMap(root map[string]any, keys ...string) map[string]any {
	var cur any = root
	for _, key := range keys {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur = m[key]
	}
	m, _ := cur.(map[string]any)
	return m
}
func asString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case json.Number:
		return x.String()
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	}
	return ""
}
func asFloat(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case json.Number:
		f, _ := x.Float64()
		return f
	case string:
		f, _ := strconv.ParseFloat(x, 64)
		return f
	}
	return 0
}
