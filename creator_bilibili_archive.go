package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var bilibiliCreatorID = regexp.MustCompile(`^[1-9][0-9]{0,19}$`)

const bilibiliArchivePageSize = 30
const bilibiliArchiveMaxPages = 1000

func normalizeBilibiliCreator(raw string) (CreatorChannel, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || !creatorPublicWebURL(u) || !strings.EqualFold(u.Hostname(), "space.bilibili.com") {
		return CreatorChannel{}, errors.New("paste a Bilibili creator URL such as https://space.bilibili.com/123")
	}
	p := splitPath(u.Path)
	if len(p) < 1 || !bilibiliCreatorID.MatchString(p[0]) || (len(p) > 1 && !(len(p) == 2 && (p[1] == "video" || p[1] == "upload"))) {
		return CreatorChannel{}, errors.New("unsupported Bilibili creator page; use its numeric space.bilibili.com profile")
	}
	return CreatorChannel{Platform: "bilibili", ChannelID: p[0], URL: "https://space.bilibili.com/" + p[0]}, nil
}

type bilibiliArchiveResponse struct {
	Code int `json:"code"`
	Data struct {
		Page struct {
			PN    int  `json:"pn"`
			PS    int  `json:"ps"`
			Count *int `json:"count"`
		} `json:"page"`
		List struct {
			Videos *[]struct {
				BVID        string `json:"bvid"`
				MID         int64  `json:"mid"`
				Author      string `json:"author"`
				Title       string `json:"title"`
				Description string `json:"description"`
				Pic         string `json:"pic"`
				Created     int64  `json:"created"`
			} `json:"vlist"`
		} `json:"list"`
	} `json:"data"`
}

// The listing is a public-upload inventory, not a promise of private/deleted
// items, all multipart pages, captions or extracted mods. A partial/error pass
// must not advance LastFullSyncAt or replace last-good persisted state.
func (a *App) enumerateBilibiliCreator(ctx context.Context, seed CreatorChannel, full bool) (CreatorChannel, []CreatorVideo, error) {
	normalized, err := normalizeBilibiliCreator(seed.URL)
	if err != nil {
		return seed, nil, err
	}
	if seed.ChannelID != "" && seed.ChannelID != normalized.ChannelID {
		return seed, nil, errors.New("Bilibili creator identity mismatch")
	}
	seed.Platform, seed.ChannelID, seed.URL = normalized.Platform, normalized.ChannelID, normalized.URL
	seed.ArchiveCoverage = "incomplete"
	seed.ArchiveWarning = "Public uploads only; private/deleted items and additional multipart pages are not enumerated. Listing discovery is not completed video analysis."
	out := []CreatorVideo{}
	seen := map[string]bool{}
	expected := -1
	retainedBytes := 0
	maxPages := bilibiliArchiveMaxPages
	if !full {
		maxPages = 3
	}
	for page := 1; page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return seed, out, err
		}
		q := url.Values{"mid": {seed.ChannelID}, "pn": {strconv.Itoa(page)}, "ps": {strconv.Itoa(bilibiliArchivePageSize)}, "order": {"pubdate"}}
		var response bilibiliArchiveResponse
		if err := a.getBilibiliPublicJSON(ctx, bilibiliCreatorAPI+"/x/space/arc/search?"+q.Encode(), nil, &response); err != nil {
			return seed, out, fmt.Errorf("Bilibili archive page %d: %w", page, err)
		}
		if response.Code != 0 {
			return seed, out, bilibiliAccessError(response.Code)
		}
		d := response.Data
		if d.Page.PN != page || d.Page.PS != bilibiliArchivePageSize || d.Page.Count == nil || *d.Page.Count < 0 || d.List.Videos == nil {
			return seed, out, errors.New("Bilibili archive pagination metadata missing or inconsistent")
		}
		if expected < 0 {
			expected = *d.Page.Count
			seed.ArchiveExpected = expected
		} else if expected != *d.Page.Count {
			return seed, out, errors.New("Bilibili archive changed during pagination; retry to establish consistent coverage")
		}
		if len(*d.List.Videos) > bilibiliArchivePageSize {
			return seed, out, errors.New("Bilibili archive exceeded requested page size")
		}
		for _, item := range *d.List.Videos {
			if !bilibiliVideoID.MatchString(item.BVID) || strconv.FormatInt(item.MID, 10) != seed.ChannelID || seen[item.BVID] {
				return seed, out, errors.New("Bilibili archive contains invalid, duplicate or mismatched creator items; coverage incomplete")
			}
			itemBytes := len(item.Title) + len(item.Description) + len(item.Pic) + len(item.Author)
			if len(item.Title) > 4096 || len(item.Description) > 65536 || len(item.Pic) > 8192 || len(item.Author) > 1024 || retainedBytes+itemBytes > 16<<20 {
				return seed, out, errors.New("Bilibili archive text budget exceeded; coverage incomplete")
			}
			retainedBytes += itemBytes
			seen[item.BVID] = true
			if seed.Title == "" {
				seed.Title = item.Author
			}
			v := CreatorVideo{ID: "bilibili_" + item.BVID + "_p1", Platform: "bilibili", URL: "https://www.bilibili.com/video/" + item.BVID + "/?p=1", Title: item.Title, Description: item.Description, Creator: item.Author, CreatorURL: seed.URL, ThumbnailURL: item.Pic, ChannelID: seed.ChannelID, VideoKind: "video", DiscoveredAt: time.Now().UTC().Format(time.RFC3339), Mods: []CreatorMod{}}
			if item.Created > 0 {
				v.PublishedAt = time.Unix(item.Created, 0).UTC().Format(time.RFC3339)
			}
			out = append(out, v)
		}
		seed.ArchiveEnumerated = len(out)
		if len(out) > expected {
			return seed, out, errors.New("Bilibili archive count exceeded provider total")
		}
		if len(out) == expected {
			seed.ArchiveCoverage = "public-upload-list-complete"
			seed.TotalVideos = expected
			return seed, out, nil
		}
		if len(*d.List.Videos) < bilibiliArchivePageSize {
			return seed, out, errors.New("Bilibili archive ended before declared total; coverage incomplete")
		}
	}
	if !full {
		seed.ArchiveCoverage = "recent-window-only"
		seed.TotalVideos = expected
		return seed, out, nil
	}
	return seed, out, fmt.Errorf("Bilibili archive exceeds bounded %d-page pass; coverage incomplete", bilibiliArchiveMaxPages)
}
