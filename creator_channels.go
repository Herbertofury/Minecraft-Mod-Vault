package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var youtubeDataAPIBase = "https://www.googleapis.com/youtube/v3"

var defaultCreatorChannels = []CreatorChannel{
	{Platform: "youtube", Handle: "@AsianHalfSquat", URL: "https://www.youtube.com/@AsianHalfSquat", Title: "AsianHalfSquat", Required: true, Source: "core", ProfileLinksStatus: "seeded", ProfileLinks: []CreatorProfileLink{{URL: "https://www.curseforge.com/members/asianhalfsquat/projects", Label: "AsianHalfSquat CurseForge projects", Kind: "creator-profile", Provider: "CurseForge", EvidenceURL: "https://www.youtube.com/@AsianHalfSquat", EvidenceType: "seed"}}},
	{Platform: "youtube", Handle: "@EnderVerseMC", URL: "https://www.youtube.com/@EnderVerseMC", Title: "EnderVerseMC", Required: true, Source: "core", ProfileLinksStatus: "seeded", ProfileLinks: []CreatorProfileLink{{URL: "https://www.curseforge.com/members/enderverse/projects", Label: "EnderVerse CurseForge projects", Kind: "creator-profile", Provider: "CurseForge", EvidenceURL: "https://www.youtube.com/@EnderVerseMC", EvidenceType: "seed"}}},
	{Platform: "tiktok", Handle: "@kizamiringo", URL: "https://www.tiktok.com/@kizamiringo", Title: "Kizamiringo", Required: true, Source: "core"},
	{Platform: "tiktok", Handle: "@its_katsumi", URL: "https://www.tiktok.com/@its_katsumi", Title: "Katsumi", Required: true, Source: "curated-core", ProfileLinksStatus: "seeded", ProfileHubURLs: []string{"https://lnk.bio/itskatsumii", "https://linktr.ee/Itskatsumii"}, ProfileLinks: []CreatorProfileLink{{URL: "https://lnk.bio/itskatsumii", Label: "Katsumi link in bio", Kind: "link-hub", Provider: "Lnk.Bio", EvidenceURL: "https://www.tiktok.com/@its_katsumi", EvidenceType: "seed"}, {URL: "https://linktr.ee/Itskatsumii", Label: "Katsumi Linktree", Kind: "link-hub", Provider: "Linktree", EvidenceURL: "https://www.tiktok.com/@its_katsumi", EvidenceType: "seed"}}},
	{Platform: "tiktok", Handle: "@speedychunks", URL: "https://www.tiktok.com/@speedychunks", Title: "SpeedyChunks", Required: true, Source: "curated-core"},
	{Platform: "tiktok", Handle: "@curseforge", URL: "https://www.tiktok.com/@curseforge", Title: "CurseForge", Required: true, Source: "curated-core"},
	{Platform: "tiktok", Handle: "@hendyvideos", URL: "https://www.tiktok.com/@hendyvideos", Title: "HendyVideos", Required: true, Source: "curated-core"},
	{Platform: "youtube", Handle: "@NoxusMods", URL: "https://www.youtube.com/@NoxusMods", Title: "Noxus", Required: true, Source: "curated-core"},
	{Platform: "youtube", Handle: "@chosenarchitect", ChannelID: "UClmdJ2bwqHjZONP9rIK7geA", URL: "https://www.youtube.com/@chosenarchitect", Title: "ChosenArchitect", Required: true, Source: "curated-core"},
	{Platform: "youtube", Handle: "@direwolf20", ChannelID: "UC_ViSsVg_3JUDyLS3E2Un5g", URL: "https://www.youtube.com/@direwolf20", Title: "direwolf20", Required: true, Source: "curated-core"},
	{Platform: "youtube", Handle: "@GOC_", ChannelID: "UCdqEs5A8vmq38Yp7Fz-Hz5A", URL: "https://www.youtube.com/@GOC_", Title: "Gaming On Caffeine", Required: true, Source: "curated-core"},
	{Platform: "youtube", Handle: "@Systemcollapse", ChannelID: "UCTWT9LtF8tXmd3qXNEoqYvA", URL: "https://www.youtube.com/@Systemcollapse", Title: "SystemCollapse", Required: true, Source: "curated-core"},
	{Platform: "youtube", ChannelID: "UCE7HSqVe0-fTumbJJn3S5hg", URL: "https://www.youtube.com/channel/UCE7HSqVe0-fTumbJJn3S5hg", Title: "Lashmak", Required: true, Source: "curated-core"},
	{Platform: "youtube", ChannelID: "UCYyRRWyMLSMCfKiIvfWYvkw", URL: "https://www.youtube.com/channel/UCYyRRWyMLSMCfKiIvfWYvkw", Title: "PwrDown", Required: true, Source: "curated-core"},
	{Platform: "youtube", ChannelID: "UCU3gwpclVZSYofj616OQKLQ", URL: "https://www.youtube.com/channel/UCU3gwpclVZSYofj616OQKLQ", Title: "Mischief of Mice", Required: true, Source: "curated-core"},
	{Platform: "youtube", Handle: "@popularmmos", ChannelID: "UCpGdL9Sn3Q5YWUH2DVUW1Ug", URL: "https://www.youtube.com/@popularmmos", Title: "PopularMMOs", Required: true, Source: "curated-core"},
	{Platform: "youtube", ChannelID: "UCS5Oz6CHmeoF7vSad0qqXfw", URL: "https://www.youtube.com/channel/UCS5Oz6CHmeoF7vSad0qqXfw", Title: "DanTDM", Required: true, Source: "curated-core"},
	{Platform: "youtube", ChannelID: "UC6Ec5NXzcESo60F3UgtgQRA", URL: "https://www.youtube.com/channel/UC6Ec5NXzcESo60F3UgtgQRA", Title: "The Breakdown", Required: true, Source: "curated-core"},
}

type CreatorChannelSuggestion struct {
	Platform  string `json:"platform"`
	Title     string `json:"title"`
	Handle    string `json:"handle,omitempty"`
	ChannelID string `json:"channelId,omitempty"`
	URL       string `json:"url"`
	Tier      string `json:"tier"`
	Focus     string `json:"focus"`
	Why       string `json:"why"`
	Priority  int    `json:"priority"`
	Tracked   bool   `json:"tracked"`
}

// Curated for recommendation signal rather than raw subscriber count. Current showcase
// channels rank first, then deep modded playthroughs, then historically important archives.
var recommendedCreatorChannels = []CreatorChannelSuggestion{
	{Platform: "tiktok", Title: "CurseForge", Handle: "@curseforge", URL: "https://www.tiktok.com/@curseforge", Tier: "current-showcases", Focus: "High-volume Minecraft mod, modpack and resource-pack discovery", Why: "Official CurseForge feed with frequent current mod showcases and direct ecosystem provenance.", Priority: 100},
	{Platform: "tiktok", Title: "Kizamiringo", Handle: "@kizamiringo", URL: "https://www.tiktok.com/@kizamiringo", Tier: "current-showcases", Focus: "Short-form Minecraft mod discovery, including visual/text-first recommendations", Why: "Core short-form source requested for full visual-text extraction and always-updated tracking.", Priority: 99},
	{Platform: "tiktok", Title: "Katsumi", Handle: "@its_katsumi", URL: "https://www.tiktok.com/@its_katsumi", Tier: "current-showcases", Focus: "Short-form Minecraft mod discovery", Why: "User-requested built-in TikTok follow; recommendations remain evidence-backed and provider-verified before entering the resolved archive.", Priority: 98},
	{Platform: "tiktok", Title: "SpeedyChunks", Handle: "@speedychunks", URL: "https://www.tiktok.com/@speedychunks", Tier: "current-showcases", Focus: "Short-form Minecraft mod discovery", Why: "User-requested built-in TikTok follow; creator links and modpacks are discovered from public evidence and never fabricated.", Priority: 97},
	{Platform: "tiktok", Title: "HendyVideos", Handle: "@hendyvideos", URL: "https://www.tiktok.com/@hendyvideos", Tier: "current-showcases", Focus: "Dedicated Minecraft mod lists and short-form mod recommendations", Why: "High recommendation density with a large cross-platform Minecraft mod catalogue.", Priority: 96},
	{Platform: "tiktok", Title: "Knarfy", Handle: "@itsknarfy", URL: "https://www.tiktok.com/@itsknarfy", Tier: "current-modded", Focus: "Minecraft experiments and mod-driven showcases", Why: "Useful active short-form signal for unusual and high-impact mods; treated as discovery rather than canonical project metadata.", Priority: 86},
	{Platform: "tiktok", Title: "The Breakdown", Handle: "@thebreakdownxyz", URL: "https://www.tiktok.com/@thebreakdownxyz", Tier: "utility", Focus: "Minecraft mod, loader, shader and modpack installation/discovery", Why: "Cross-platform utility source that complements its built-in YouTube archive with short-form ecosystem updates.", Priority: 78},
	{Platform: "tiktok", Title: "The Crimson Gaming", Handle: "@thecrimsongaming", URL: "https://www.tiktok.com/@thecrimsongaming", Tier: "historical-gold", Focus: "Dedicated Minecraft mod showcases and project-specific walkthroughs", Why: "Long-running mod-showcase catalogue with an explicitly linked TikTok presence; useful as a secondary discovery archive.", Priority: 76},
	{Platform: "tiktok", Title: "laveOrc", Handle: "@ygz207", URL: "https://www.tiktok.com/@ygz207", Tier: "current-showcases", Focus: "Current Minecraft horror-mod clips with project-name callouts", Why: "Active 2026 short-form mod signal; recommendations remain provider-verified before entering the resolved archive.", Priority: 72},
	{Platform: "youtube", Title: "Noxus", Handle: "@NoxusMods", URL: "https://www.youtube.com/@NoxusMods", Tier: "current-showcases", Focus: "Current mod reviews, monthly roundups, shaders, packs and full showcases", Why: "Extremely high recommendation density and current-version coverage.", Priority: 100},
	{Platform: "youtube", Title: "ChosenArchitect", Handle: "@chosenarchitect", ChannelID: "UClmdJ2bwqHjZONP9rIK7geA", URL: "https://www.youtube.com/@chosenarchitect", Tier: "current-modded", Focus: "Daily modded Minecraft, Create, automation, magic and curated modpacks", Why: "Large active archive with strong modern mod and modpack discovery signal.", Priority: 95},
	{Platform: "youtube", Title: "direwolf20", Handle: "@direwolf20", ChannelID: "UC_ViSsVg_3JUDyLS3E2Un5g", URL: "https://www.youtube.com/@direwolf20", Tier: "current-modded", Focus: "Mod spotlights, tutorials, FTB and long-running modded Let's Plays", Why: "One of the deepest historical and current technical mod archives on YouTube.", Priority: 94},
	{Platform: "youtube", Title: "Gaming On Caffeine", Handle: "@GOC_", ChannelID: "UCdqEs5A8vmq38Yp7Fz-Hz5A", URL: "https://www.youtube.com/@GOC_", Tier: "current-modded", Focus: "Modded questing packs, tech progression and automation", Why: "Thousands of modded videos with dense in-context mod usage and progression evidence.", Priority: 90},
	{Platform: "youtube", Title: "SystemCollapse", Handle: "@Systemcollapse", ChannelID: "UCTWT9LtF8tXmd3qXNEoqYvA", URL: "https://www.youtube.com/@Systemcollapse", Tier: "deep-modded", Focus: "Methodical modded Minecraft and questing pack playthroughs", Why: "Excellent long-form source for identifying useful mods inside real pack progression.", Priority: 86},
	{Platform: "youtube", Title: "Lashmak", ChannelID: "UCE7HSqVe0-fTumbJJn3S5hg", URL: "https://www.youtube.com/channel/UCE7HSqVe0-fTumbJJn3S5hg", Tier: "deep-modded", Focus: "Fast-paced expert modpack playthroughs and technical progression", Why: "High-signal expert usage, especially for complex automation and endgame mods.", Priority: 84},
	{Platform: "youtube", Title: "PwrDown", ChannelID: "UCYyRRWyMLSMCfKiIvfWYvkw", URL: "https://www.youtube.com/channel/UCYyRRWyMLSMCfKiIvfWYvkw", Tier: "historical-gold", Focus: "Minecraft Java mod lists, shaders, packs, data packs and lesser-known finds", Why: "A compact historical archive with unusually high recommendations per video.", Priority: 82},
	{Platform: "youtube", Title: "Mischief of Mice", ChannelID: "UCU3gwpclVZSYofj616OQKLQ", URL: "https://www.youtube.com/channel/UCU3gwpclVZSYofj616OQKLQ", Tier: "historical-gold", Focus: "Bit-by-Bit mod tutorials, showcases and long-form modded play", Why: "Strong detailed historical explanations and project-specific tutorial evidence.", Priority: 80},
	{Platform: "youtube", Title: "PopularMMOs", Handle: "@popularmmos", ChannelID: "UCpGdL9Sn3Q5YWUH2DVUW1Ug", URL: "https://www.youtube.com/@popularmmos", Tier: "historical-giant", Focus: "Classic mod showcases, modded Let's Plays and challenge packs", Why: "Massive historical Minecraft mod corpus and one of the genre's most influential archives.", Priority: 76},
	{Platform: "youtube", Title: "DanTDM", ChannelID: "UCS5Oz6CHmeoF7vSad0qqXfw", URL: "https://www.youtube.com/channel/UCS5Oz6CHmeoF7vSad0qqXfw", Tier: "historical-giant", Focus: "Classic Minecraft mod reviews and showcases", Why: "Hundreds of classic mod-showcase videos make this a uniquely valuable historical index.", Priority: 74},
	{Platform: "youtube", Title: "The Breakdown", ChannelID: "UC6Ec5NXzcESo60F3UgtgQRA", URL: "https://www.youtube.com/channel/UC6Ec5NXzcESo60F3UgtgQRA", Tier: "utility", Focus: "Minecraft mod, loader, shader, datapack and installation tutorials", Why: "Useful supporting signal for mod discovery, installation and ecosystem changes.", Priority: 68},
}

type CreatorChannel struct {
	Platform                   string                   `json:"platform"`
	Handle                     string                   `json:"handle"`
	URL                        string                   `json:"url"`
	ChannelID                  string                   `json:"channelId,omitempty"`
	Title                      string                   `json:"title,omitempty"`
	AvatarURL                  string                   `json:"avatarUrl,omitempty"`
	Bio                        string                   `json:"bio,omitempty"`
	UploadsPlaylist            string                   `json:"uploadsPlaylist,omitempty"`
	ProfileHubURLs             []string                 `json:"profileHubUrls,omitempty"`
	ProfileLinks               []CreatorProfileLink     `json:"profileLinks,omitempty"`
	ProfileLinksStatus         string                   `json:"profileLinksStatus,omitempty"`
	ProfileLinksError          string                   `json:"profileLinksError,omitempty"`
	ProfileLinksRefreshedAt    string                   `json:"profileLinksRefreshedAt,omitempty"`
	CreatorModpacks            []CreatorReleasedModpack `json:"creatorModpacks,omitempty"`
	CreatorModpacksStatus      string                   `json:"creatorModpacksStatus,omitempty"`
	CreatorModpacksError       string                   `json:"creatorModpacksError,omitempty"`
	CreatorModpacksRefreshedAt string                   `json:"creatorModpacksRefreshedAt,omitempty"`
	Required                   bool                     `json:"required"`
	TotalVideos                int                      `json:"totalVideos"`
	IndexedVideos              int                      `json:"indexedVideos"`
	AnalyzedVideos             int                      `json:"analyzedVideos"`
	PendingVideos              int                      `json:"pendingVideos"`
	FailedVideos               int                      `json:"failedVideos"`
	Shorts                     int                      `json:"shorts"`
	Recommendations            int                      `json:"recommendations"`
	Unresolved                 int                      `json:"unresolved"`
	SyncStatus                 string                   `json:"syncStatus"`
	SyncError                  string                   `json:"syncError,omitempty"`
	LastSyncedAt               string                   `json:"lastSyncedAt,omitempty"`
	LastFullSyncAt             string                   `json:"lastFullSyncAt,omitempty"`
	LastAttemptAt              string                   `json:"lastAttemptAt,omitempty"`
	SyncFailures               int                      `json:"syncFailures,omitempty"`
	LastAnalyzeAt              string                   `json:"lastAnalyzeAt,omitempty"`
	Source                     string                   `json:"source,omitempty"`
	AddedAt                    string                   `json:"addedAt,omitempty"`
	Paused                     bool                     `json:"paused"`
}

type CreatorRecommendation struct {
	ChannelHandle     string     `json:"channelHandle"`
	ChannelTitle      string     `json:"channelTitle"`
	ChannelURL        string     `json:"channelUrl"`
	VideoID           string     `json:"videoId"`
	VideoTitle        string     `json:"videoTitle"`
	VideoURL          string     `json:"videoUrl"`
	VideoKind         string     `json:"videoKind"`
	VideoPublishedAt  string     `json:"videoPublishedAt"`
	VideoThumbnailURL string     `json:"videoThumbnailUrl,omitempty"`
	Mod               CreatorMod `json:"mod"`
	RecommendationAt  string     `json:"recommendationAt"`
}

type creatorChannelSyncRequest struct {
	URL     string `json:"url"`
	Full    bool   `json:"full"`
	Analyze bool   `json:"analyze"`
}

type creatorChannelManageRequest struct {
	URL     string   `json:"url"`
	URLs    []string `json:"urls,omitempty"`
	Sync    *bool    `json:"sync,omitempty"`
	Analyze *bool    `json:"analyze,omitempty"`
	Paused  *bool    `json:"paused,omitempty"`
}

type ytDLPFlatEntry struct {
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	URL         string           `json:"url"`
	WebpageURL  string           `json:"webpage_url"`
	Description string           `json:"description"`
	Timestamp   int64            `json:"timestamp"`
	UploadDate  string           `json:"upload_date"`
	Duration    float64          `json:"duration"`
	Thumbnail   string           `json:"thumbnail"`
	Channel     string           `json:"channel"`
	ChannelID   string           `json:"channel_id"`
	ChannelURL  string           `json:"channel_url"`
	Uploader    string           `json:"uploader"`
	UploaderID  string           `json:"uploader_id"`
	Entries     []ytDLPFlatEntry `json:"entries"`
}

func canonicalCreatorHandle(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "@") {
		return raw
	}
	if strings.HasPrefix(raw, "UC") && !strings.ContainsAny(raw, "/?#") {
		return ""
	}
	if u, err := url.Parse(raw); err == nil && u.Host != "" {
		for _, p := range splitPath(u.Path) {
			if strings.HasPrefix(p, "@") {
				return p
			}
		}
		return ""
	}
	if strings.Contains(raw, "/@") {
		p := raw[strings.LastIndex(raw, "/@")+1:]
		if i := strings.IndexAny(p, "/?#"); i >= 0 {
			p = p[:i]
		}
		return p
	}
	if !strings.ContainsAny(raw, "/?# ") {
		return "@" + strings.TrimPrefix(raw, "@")
	}
	return ""
}

var youtubeChannelIDPattern = regexp.MustCompile(`^UC[A-Za-z0-9_-]{18,30}$`)

func normalizeCreatorChannelInput(raw string) (CreatorChannel, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return CreatorChannel{}, errors.New("YouTube or TikTok creator URL, @handle or channel ID is required")
	}
	if strings.HasPrefix(strings.ToLower(raw), "tiktok:") {
		h := canonicalCreatorHandle(strings.TrimSpace(raw[len("tiktok:"):]))
		if h == "" {
			return CreatorChannel{}, errors.New("TikTok creator handle is required after tiktok:")
		}
		return CreatorChannel{Platform: "tiktok", Handle: h, URL: "https://www.tiktok.com/" + h}, nil
	}
	if youtubeChannelIDPattern.MatchString(raw) {
		return CreatorChannel{Platform: "youtube", ChannelID: raw, URL: "https://www.youtube.com/channel/" + raw}, nil
	}
	if h := canonicalCreatorHandle(raw); h != "" && !strings.Contains(raw, "://") {
		return CreatorChannel{Platform: "youtube", Handle: h, URL: "https://www.youtube.com/" + h}, nil
	}
	candidate := raw
	if !strings.Contains(candidate, "://") && (strings.Contains(candidate, "youtube.com/") || strings.Contains(candidate, "tiktok.com/")) {
		candidate = "https://" + candidate
	}
	u, err := url.Parse(candidate)
	if err == nil && u.Host != "" {
		host := strings.ToLower(strings.TrimPrefix(strings.TrimPrefix(u.Hostname(), "www."), "m."))
		if host == "tiktok.com" || host == "tiktokv.com" {
			parts := splitPath(u.Path)
			if len(parts) == 0 || !strings.HasPrefix(parts[0], "@") {
				return CreatorChannel{}, errors.New("paste a TikTok creator profile URL such as https://www.tiktok.com/@creator")
			}
			if len(parts) > 1 && strings.EqualFold(parts[1], "video") {
				return CreatorChannel{}, errors.New("paste the TikTok creator profile URL rather than one video")
			}
			h := canonicalCreatorHandle(parts[0])
			return CreatorChannel{Platform: "tiktok", Handle: h, URL: "https://www.tiktok.com/" + h}, nil
		}
		if host != "youtube.com" {
			return CreatorChannel{}, errors.New("only YouTube and TikTok creators can be added to the Creator Archive")
		}
		parts := splitPath(u.Path)
		if len(parts) == 0 {
			return CreatorChannel{}, errors.New("paste a YouTube channel URL, not the YouTube homepage")
		}
		if strings.HasPrefix(parts[0], "@") {
			h := canonicalCreatorHandle(parts[0])
			return CreatorChannel{Platform: "youtube", Handle: h, URL: "https://www.youtube.com/" + h}, nil
		}
		if parts[0] == "channel" && len(parts) > 1 && youtubeChannelIDPattern.MatchString(parts[1]) {
			return CreatorChannel{Platform: "youtube", ChannelID: parts[1], URL: "https://www.youtube.com/channel/" + parts[1]}, nil
		}
		if parts[0] == "watch" || parts[0] == "shorts" || parts[0] == "live" {
			return CreatorChannel{}, errors.New("paste the creator's channel URL or @handle rather than one video")
		}
		// Legacy /user/, /c/ and custom channel paths are still valid inputs. yt-dlp
		// resolves them to the canonical channel ID during the first sync.
		if parts[0] == "user" || parts[0] == "c" || len(parts) == 1 {
			u.RawQuery, u.Fragment = "", ""
			return CreatorChannel{Platform: "youtube", URL: "https://www.youtube.com" + strings.TrimRight(u.EscapedPath(), "/")}, nil
		}
	}
	if h := canonicalCreatorHandle(raw); h != "" {
		return CreatorChannel{Platform: "youtube", Handle: h, URL: "https://www.youtube.com/" + h}, nil
	}
	return CreatorChannel{}, errors.New("could not recognize that YouTube or TikTok creator")
}

func creatorPlatform(values ...string) string {
	for _, raw := range values {
		v := strings.ToLower(strings.TrimSpace(raw))
		if v == "youtube" || v == "tiktok" {
			return v
		}
		if strings.Contains(v, "tiktok.com") || strings.Contains(v, "tiktokv.com") {
			return "tiktok"
		}
		if strings.Contains(v, "youtube.com") || strings.Contains(v, "youtu.be") {
			return "youtube"
		}
	}
	return "youtube"
}

func normalizedCreatorURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return strings.ToLower(strings.TrimRight(raw, "/"))
	}
	u.RawQuery, u.Fragment = "", ""
	u.Scheme = "https"
	u.Host = strings.ToLower(u.Host)
	return strings.ToLower(strings.TrimRight(u.String(), "/"))
}

func creatorChannelsEquivalent(a, b CreatorChannel) bool {
	if creatorPlatform(a.Platform, a.URL) != creatorPlatform(b.Platform, b.URL) {
		return false
	}
	if a.ChannelID != "" && b.ChannelID != "" && a.ChannelID == b.ChannelID {
		return true
	}
	ah, bh := canonicalCreatorHandle(a.Handle), canonicalCreatorHandle(b.Handle)
	if ah != "" && bh != "" && strings.EqualFold(ah, bh) {
		return true
	}
	au, bu := normalizedCreatorURL(a.URL), normalizedCreatorURL(b.URL)
	return au != "" && bu != "" && au == bu
}

func creatorChannelKey(ch CreatorChannel) string {
	platform := creatorPlatform(ch.Platform, ch.URL)
	if ch.ChannelID != "" {
		return platform + ":" + ch.ChannelID
	}
	if h := canonicalCreatorHandle(firstNonEmpty(ch.Handle, ch.URL)); h != "" {
		return platform + ":" + strings.ToLower(h)
	}
	return platform + ":url:" + normalizedCreatorURL(ch.URL)
}

func creatorVideoChannelKey(v CreatorVideo) string {
	platform := creatorPlatform(v.Platform, v.CreatorURL, v.URL)
	if v.ChannelID != "" {
		return platform + ":" + v.ChannelID
	}
	if v.ChannelHandle != "" {
		return platform + ":" + strings.ToLower(v.ChannelHandle)
	}
	if v.CreatorURL != "" {
		if h := canonicalCreatorHandle(v.CreatorURL); h != "" {
			return platform + ":" + strings.ToLower(h)
		}
		return platform + ":url:" + normalizedCreatorURL(v.CreatorURL)
	}
	return platform + ":" + strings.ToLower(v.Creator)
}

func creatorVideoBelongsToChannel(v CreatorVideo, ch CreatorChannel) bool {
	if creatorVideoChannelKey(v) == creatorChannelKey(ch) {
		return true
	}
	if creatorPlatform(v.Platform, v.CreatorURL, v.URL) != creatorPlatform(ch.Platform, ch.URL) {
		return false
	}
	return v.ChannelHandle != "" && ch.Handle != "" && strings.EqualFold(v.ChannelHandle, ch.Handle)
}

func mergeCreatorChannelSeed(dst *CreatorChannel, seed CreatorChannel) bool {
	changed := false
	if dst.Platform == "" && seed.Platform != "" {
		dst.Platform = seed.Platform
		changed = true
	}
	if dst.Handle == "" && seed.Handle != "" {
		dst.Handle = seed.Handle
		changed = true
	}
	if dst.URL == "" && seed.URL != "" {
		dst.URL = seed.URL
		changed = true
	}
	if dst.ChannelID == "" && seed.ChannelID != "" {
		dst.ChannelID = seed.ChannelID
		changed = true
	}
	if dst.Title == "" && seed.Title != "" {
		dst.Title = seed.Title
		changed = true
	}
	if seed.Required && !dst.Required {
		dst.Required = true
		changed = true
	}
	if seed.Required && seed.Source != "" && dst.Source != seed.Source {
		dst.Source = seed.Source
		changed = true
	} else if !seed.Required && dst.Source == "" && seed.Source != "" {
		dst.Source = seed.Source
		changed = true
	}
	if mergeCreatorProfileSeed(dst, seed) {
		changed = true
	}
	return changed
}

func (a *App) ensureDefaultCreatorChannelsLocked() bool {
	changed := false
	for _, seed := range defaultCreatorChannels {
		found := -1
		for i := range a.creatorChannels {
			if creatorChannelsEquivalent(a.creatorChannels[i], seed) {
				found = i
				break
			}
		}
		if found >= 0 {
			if mergeCreatorChannelSeed(&a.creatorChannels[found], seed) {
				changed = true
			}
			continue
		}
		a.creatorChannels = append(a.creatorChannels, seed)
		changed = true
	}
	if a.creatorSyncRunning == nil {
		a.creatorSyncRunning = map[string]bool{}
	}
	return changed
}

func (a *App) ensureDefaultCreatorChannels() {
	a.dataMu.Lock()
	changed := a.ensureDefaultCreatorChannelsLocked()
	a.dataMu.Unlock()
	if changed {
		_ = a.saveCreatorChannels()
	}
}

func (a *App) creatorSyncSlots(max int) int {
	if max < 1 {
		max = 1
	}
	if max > 4 {
		max = 4
	}
	a.creatorSyncMu.Lock()
	running := len(a.creatorSyncRunning)
	a.creatorSyncMu.Unlock()
	if running >= max {
		return 0
	}
	return max - running
}

func (a *App) refreshCreatorChannelStatsLocked() {
	for i := range a.creatorChannels {
		ch := &a.creatorChannels[i]
		indexed, analyzed, pending, failed, shorts, recs, unresolved := 0, 0, 0, 0, 0, 0, 0
		ck := creatorChannelKey(*ch)
		for _, v := range a.creatorVideos {
			if creatorVideoChannelKey(v) != ck && !creatorVideoBelongsToChannel(v, *ch) {
				continue
			}
			indexed++
			if v.VideoKind == "short" {
				shorts++
			}
			if v.AnalyzedAt != "" {
				analyzed++
			} else {
				pending++
			}
			if v.AnalysisError != "" {
				failed++
			}
			recs += len(v.Mods)
			unresolved += len(v.UnresolvedMentions)
		}
		ch.IndexedVideos, ch.AnalyzedVideos, ch.PendingVideos, ch.FailedVideos = indexed, analyzed, pending, failed
		ch.Shorts, ch.Recommendations, ch.Unresolved = shorts, recs, unresolved
		if ch.TotalVideos < indexed {
			ch.TotalVideos = indexed
		}
	}
}

func (a *App) handleCreatorChannels(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.dataMu.Lock()
		a.refreshCreatorChannelStatsLocked()
		channels := append([]CreatorChannel(nil), a.creatorChannels...)
		a.dataMu.Unlock()
		sort.SliceStable(channels, func(i, j int) bool {
			if channels[i].Required != channels[j].Required {
				return channels[i].Required
			}
			return strings.ToLower(firstNonEmpty(channels[i].Title, channels[i].Handle, channels[i].URL)) < strings.ToLower(firstNonEmpty(channels[j].Title, channels[j].Handle, channels[j].URL))
		})
		suggestions := append([]CreatorChannelSuggestion(nil), recommendedCreatorChannels...)
		for i := range suggestions {
			seed := CreatorChannel{Platform: firstNonEmpty(suggestions[i].Platform, "youtube"), Handle: suggestions[i].Handle, ChannelID: suggestions[i].ChannelID, URL: suggestions[i].URL}
			for _, ch := range channels {
				if creatorChannelsEquivalent(ch, seed) {
					suggestions[i].Tracked = true
					break
				}
			}
		}
		sort.SliceStable(suggestions, func(i, j int) bool { return suggestions[i].Priority > suggestions[j].Priority })
		writeJSON(w, http.StatusOK, map[string]any{"channels": channels, "count": len(channels), "suggestions": suggestions, "archival": true})
	case http.MethodPost:
		var in creatorChannelManageRequest
		if err := decodeJSON(r, &in); err != nil {
			writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
			return
		}
		inputs := append([]string(nil), in.URLs...)
		if strings.TrimSpace(in.URL) != "" {
			inputs = append(inputs, in.URL)
		}
		if len(inputs) == 0 {
			writeJSON(w, http.StatusBadRequest, APIError{Error: "YouTube or TikTok creator URL, @handle or channel ID is required"})
			return
		}
		syncNow, analyze := true, true
		if in.Sync != nil {
			syncNow = *in.Sync
		}
		if in.Analyze != nil {
			analyze = *in.Analyze
		}
		added, existing, started := []CreatorChannel{}, []CreatorChannel{}, 0
		for _, raw := range inputs {
			ch, wasAdded, err := a.addCreatorChannel(raw)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
				return
			}
			if wasAdded {
				added = append(added, ch)
			} else {
				existing = append(existing, ch)
			}
			if syncNow && !ch.Paused && a.startCreatorChannelSync(ch, true, analyze) {
				started++
			}
		}
		writeJSON(w, http.StatusAccepted, map[string]any{"added": added, "existing": existing, "started": started, "full": true, "analyze": analyze})
	case http.MethodPatch:
		var in creatorChannelManageRequest
		if err := decodeJSON(r, &in); err != nil || in.Paused == nil {
			writeJSON(w, http.StatusBadRequest, APIError{Error: "channel and paused state are required"})
			return
		}
		seed, err := normalizeCreatorChannelInput(in.URL)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
			return
		}
		a.dataMu.Lock()
		found := false
		var updated CreatorChannel
		for i := range a.creatorChannels {
			if creatorChannelsEquivalent(a.creatorChannels[i], seed) {
				a.creatorChannels[i].Paused = *in.Paused
				updated = a.creatorChannels[i]
				found = true
				break
			}
		}
		a.dataMu.Unlock()
		if !found {
			writeJSON(w, http.StatusNotFound, APIError{Error: "tracked creator channel not found"})
			return
		}
		if err := a.saveCreatorChannels(); err != nil {
			writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"channel": updated, "paused": updated.Paused})
	case http.MethodDelete:
		raw := strings.TrimSpace(r.URL.Query().Get("url"))
		seed, err := normalizeCreatorChannelInput(raw)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
			return
		}
		a.dataMu.Lock()
		idx := -1
		var target CreatorChannel
		for i := range a.creatorChannels {
			if creatorChannelsEquivalent(a.creatorChannels[i], seed) {
				idx, target = i, a.creatorChannels[i]
				break
			}
		}
		if idx < 0 {
			a.dataMu.Unlock()
			writeJSON(w, http.StatusNotFound, APIError{Error: "tracked creator channel not found"})
			return
		}
		if target.Required {
			a.dataMu.Unlock()
			writeJSON(w, http.StatusConflict, APIError{Error: "core Creator Archive channels stay tracked; pause automatic sync instead"})
			return
		}
		key := creatorChannelKey(target)
		a.creatorSyncMu.Lock()
		running := a.creatorSyncRunning[key]
		a.creatorSyncMu.Unlock()
		if running {
			a.dataMu.Unlock()
			writeJSON(w, http.StatusConflict, APIError{Error: "channel sync is currently running; pause it and remove after this pass finishes"})
			return
		}
		a.creatorChannels = append(a.creatorChannels[:idx], a.creatorChannels[idx+1:]...)
		a.dataMu.Unlock()
		if err := a.saveCreatorChannels(); err != nil {
			writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"removed": true, "preservedArchive": true, "channel": target})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (a *App) addCreatorChannel(raw string) (CreatorChannel, bool, error) {
	seed, err := normalizeCreatorChannelInput(raw)
	if err != nil {
		return CreatorChannel{}, false, err
	}
	seed.AddedAt = time.Now().UTC().Format(time.RFC3339)
	seed.Source = "custom"
	for _, s := range recommendedCreatorChannels {
		candidate := CreatorChannel{Platform: firstNonEmpty(s.Platform, "youtube"), Handle: s.Handle, ChannelID: s.ChannelID, URL: s.URL}
		if creatorChannelsEquivalent(seed, candidate) {
			seed.Source = "recommended"
			seed.Title = s.Title
			if seed.Handle == "" {
				seed.Handle = s.Handle
			}
			if seed.ChannelID == "" {
				seed.ChannelID = s.ChannelID
			}
			if seed.URL == "" {
				seed.URL = s.URL
			}
			break
		}
	}
	a.dataMu.Lock()
	for _, ch := range a.creatorChannels {
		if creatorChannelsEquivalent(ch, seed) {
			a.dataMu.Unlock()
			return ch, false, nil
		}
	}
	a.creatorChannels = append(a.creatorChannels, seed)
	a.dataMu.Unlock()
	if err := a.saveCreatorChannels(); err != nil {
		return CreatorChannel{}, false, err
	}
	return seed, true, nil
}

func (a *App) handleCreatorChannelSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var in creatorChannelSyncRequest
	if err := decodeJSON(r, &in); err != nil && !errors.Is(err, context.Canceled) {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	if !in.Full {
		in.Full = true
	}
	if !in.Analyze {
		in.Analyze = true
	}
	targets := []CreatorChannel{}
	var requested CreatorChannel
	if strings.TrimSpace(in.URL) != "" {
		var err error
		requested, err = normalizeCreatorChannelInput(in.URL)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
			return
		}
	}
	a.dataMu.RLock()
	for _, ch := range a.creatorChannels {
		if strings.TrimSpace(in.URL) == "" {
			if ch.Paused {
				continue
			}
			targets = append(targets, ch)
			continue
		}
		if creatorChannelsEquivalent(ch, requested) {
			targets = append(targets, ch)
		}
	}
	a.dataMu.RUnlock()
	if len(targets) == 0 && strings.TrimSpace(in.URL) != "" {
		ch, _, err := a.addCreatorChannel(in.URL)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
			return
		}
		targets = append(targets, ch)
	}
	started := 0
	for _, target := range targets {
		if a.startCreatorChannelSync(target, in.Full, in.Analyze) {
			started++
		}
	}
	writeJSON(w, http.StatusAccepted, map[string]any{"started": started, "requested": len(targets), "full": in.Full, "analyze": in.Analyze})
}

func (a *App) startCreatorChannelSync(ch CreatorChannel, full, analyze bool) bool {
	key := creatorChannelKey(ch)
	a.creatorSyncMu.Lock()
	if a.creatorSyncRunning == nil {
		a.creatorSyncRunning = map[string]bool{}
	}
	if a.creatorSyncRunning[key] {
		a.creatorSyncMu.Unlock()
		return false
	}
	a.creatorSyncRunning[key] = true
	a.creatorSyncMu.Unlock()
	a.setCreatorChannelSyncState(ch, "syncing", "")
	go func() {
		defer func() { a.creatorSyncMu.Lock(); delete(a.creatorSyncRunning, key); a.creatorSyncMu.Unlock() }()
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Minute)
		defer cancel()
		if err := a.syncCreatorChannel(ctx, ch, full); err != nil {
			a.setCreatorChannelSyncState(ch, "error", err.Error())
			return
		}
		a.setCreatorChannelSyncState(ch, "indexed", "")
		if analyze {
			if full {
				a.resetCreatorChannelFailures(ch)
			}
			ctx2, cancel2 := context.WithCancel(context.Background())
			defer cancel2()
			_ = a.processCreatorChannelQueue(ctx2, ch, 0)
		}
		a.setCreatorChannelSyncState(ch, "ready", "")
	}()
	return true
}

func (a *App) setCreatorChannelSyncState(target CreatorChannel, status, errText string) {
	a.dataMu.Lock()
	for i := range a.creatorChannels {
		if creatorChannelsEquivalent(a.creatorChannels[i], target) {
			a.creatorChannels[i].SyncStatus = status
			a.creatorChannels[i].SyncError = errText
			if target.ChannelID != "" {
				a.creatorChannels[i].ChannelID = target.ChannelID
			}
			if target.Title != "" {
				a.creatorChannels[i].Title = target.Title
			}
			if target.AvatarURL != "" {
				a.creatorChannels[i].AvatarURL = target.AvatarURL
			}
			if target.UploadsPlaylist != "" {
				a.creatorChannels[i].UploadsPlaylist = target.UploadsPlaylist
			}
			if target.TotalVideos > 0 {
				a.creatorChannels[i].TotalVideos = target.TotalVideos
			}
			now := time.Now().UTC().Format(time.RFC3339)
			if status == "syncing" {
				a.creatorChannels[i].LastAttemptAt = now
			}
			if status == "error" {
				a.creatorChannels[i].SyncFailures++
			}
			if status == "ready" || status == "indexed" {
				a.creatorChannels[i].LastSyncedAt = now
				a.creatorChannels[i].SyncFailures = 0
			}
			break
		}
	}
	a.refreshCreatorChannelStatsLocked()
	a.dataMu.Unlock()
	_ = a.saveCreatorChannels()
}

func (a *App) syncCreatorChannel(ctx context.Context, ch CreatorChannel, full bool) error {
	var resolved CreatorChannel
	var videos []CreatorVideo
	var err error
	switch creatorPlatform(ch.Platform, ch.URL) {
	case "tiktok":
		resolved, videos, err = a.enumerateTikTokChannelYTDLP(ctx, ch, full)
	default:
		a.mu.RLock()
		key := strings.TrimSpace(a.settings.YouTubeAPIKey)
		a.mu.RUnlock()
		if key != "" {
			resolved, videos, err = a.enumerateYouTubeChannelAPI(ctx, key, ch, full)
		}
		if err != nil || len(videos) == 0 {
			resolved, videos, err = a.enumerateYouTubeChannelYTDLP(ctx, ch, full)
		}
	}
	if err != nil {
		// Creator-controlled profile/link-hub metadata is refreshed independently
		// of video enumeration so a TikTok/YouTube crawl failure does not prevent
		// the Vault from keeping creator links current.
		if linked, _ := a.refreshCreatorProfileLinks(ctx, ch); linked.URL != "" {
			if packed, _ := a.refreshCreatorModpacks(ctx, linked); packed.URL != "" {
				linked = packed
			}
			_ = a.persistCreatorProfileMetadata(linked)
		}
		return err
	}
	if linked, _ := a.refreshCreatorProfileLinks(ctx, resolved); linked.URL != "" {
		resolved = linked
	}
	if packed, _ := a.refreshCreatorModpacks(ctx, resolved); packed.URL != "" {
		resolved = packed
	}
	if len(videos) == 0 {
		return errors.New("channel enumeration returned zero videos")
	}
	for _, v := range videos {
		a.upsertCreatorVideo(v)
	}
	if err := a.saveCreatorVideos(); err != nil {
		return err
	}
	resolved.Required = ch.Required
	resolved.Source = firstNonEmpty(ch.Source, resolved.Source)
	resolved.AddedAt = firstNonEmpty(ch.AddedAt, resolved.AddedAt)
	resolved.Paused = ch.Paused
	resolved.TotalVideos = len(videos)
	now := time.Now().UTC().Format(time.RFC3339)
	resolved.LastSyncedAt = now
	if full {
		resolved.LastFullSyncAt = now
	}
	a.dataMu.Lock()
	found := false
	for i := range a.creatorChannels {
		if creatorChannelsEquivalent(a.creatorChannels[i], ch) || creatorChannelsEquivalent(a.creatorChannels[i], resolved) {
			old := a.creatorChannels[i]
			if resolved.LastFullSyncAt == "" {
				resolved.LastFullSyncAt = old.LastFullSyncAt
			}
			if resolved.AddedAt == "" {
				resolved.AddedAt = old.AddedAt
			}
			if resolved.Source == "" {
				resolved.Source = old.Source
			}
			if resolved.LastAttemptAt == "" {
				resolved.LastAttemptAt = old.LastAttemptAt
			}
			resolved.SyncFailures = old.SyncFailures
			resolved.Paused = old.Paused
			resolved.Required = old.Required
			if !full && resolved.TotalVideos < old.TotalVideos {
				resolved.TotalVideos = old.TotalVideos
			}
			a.creatorChannels[i] = resolved
			found = true
			break
		}
	}
	if !found {
		a.creatorChannels = append(a.creatorChannels, resolved)
	}
	a.refreshCreatorChannelStatsLocked()
	a.dataMu.Unlock()
	return a.saveCreatorChannels()
}

func (a *App) enumerateYouTubeChannelAPI(ctx context.Context, key string, seed CreatorChannel, full bool) (CreatorChannel, []CreatorVideo, error) {
	handle := canonicalCreatorHandle(firstNonEmpty(seed.Handle, seed.URL))
	if handle == "" && seed.ChannelID == "" {
		return seed, nil, errors.New("channel handle or channel ID is required for YouTube Data API enumeration")
	}
	var cr struct {
		Items []struct {
			ID      string `json:"id"`
			Snippet struct {
				Title       string `json:"title"`
				Description string `json:"description"`
				Thumbnails  map[string]struct {
					URL string `json:"url"`
				} `json:"thumbnails"`
			} `json:"snippet"`
			ContentDetails struct {
				RelatedPlaylists struct {
					Uploads string `json:"uploads"`
				} `json:"relatedPlaylists"`
			} `json:"contentDetails"`
			Statistics struct {
				VideoCount string `json:"videoCount"`
			} `json:"statistics"`
		} `json:"items"`
	}
	channelQuery := url.Values{"part": {"snippet,contentDetails,statistics"}, "key": {key}}
	if seed.ChannelID != "" {
		channelQuery.Set("id", seed.ChannelID)
	} else {
		channelQuery.Set("forHandle", handle)
	}
	u := strings.TrimRight(youtubeDataAPIBase, "/") + "/channels?" + channelQuery.Encode()
	if err := a.getJSON(ctx, u, nil, &cr); err != nil || len(cr.Items) == 0 {
		if err == nil {
			err = errors.New("YouTube channel not found")
		}
		return seed, nil, err
	}
	item := cr.Items[0]
	seed.ChannelID = item.ID
	seed.Title = item.Snippet.Title
	seed.Bio = item.Snippet.Description
	seed.UploadsPlaylist = item.ContentDetails.RelatedPlaylists.Uploads
	seed.AvatarURL = bestYTThumb(item.Snippet.Thumbnails)
	seed.TotalVideos, _ = strconv.Atoi(item.Statistics.VideoCount)
	if seed.URL == "" {
		if handle != "" {
			seed.URL = "https://www.youtube.com/" + handle
		} else {
			seed.URL = "https://www.youtube.com/channel/" + seed.ChannelID
		}
	}
	if handle != "" {
		seed.Handle = handle
	}
	if seed.UploadsPlaylist == "" {
		return seed, nil, errors.New("YouTube uploads playlist was not exposed")
	}
	known := map[string]bool{}
	a.dataMu.RLock()
	for _, v := range a.creatorVideos {
		if creatorVideoChannelKey(v) == creatorChannelKey(seed) {
			known[v.ID] = true
		}
	}
	a.dataMu.RUnlock()
	out := []CreatorVideo{}
	pageToken := ""
	pages := 0
	now := time.Now().UTC().Format(time.RFC3339)
	for {
		vals := url.Values{"part": {"snippet,contentDetails,status"}, "playlistId": {seed.UploadsPlaylist}, "maxResults": {"50"}, "key": {key}}
		if pageToken != "" {
			vals.Set("pageToken", pageToken)
		}
		var pr struct {
			NextPageToken string `json:"nextPageToken"`
			PageInfo      struct {
				TotalResults int `json:"totalResults"`
			} `json:"pageInfo"`
			Items []struct {
				Snippet struct {
					Title, Description, ChannelTitle string
					Thumbnails                       map[string]struct {
						URL string `json:"url"`
					} `json:"thumbnails"`
					ResourceID struct {
						VideoID string `json:"videoId"`
					} `json:"resourceId"`
				} `json:"snippet"`
				ContentDetails struct{ VideoID, VideoPublishedAt string } `json:"contentDetails"`
				Status         struct {
					PrivacyStatus string `json:"privacyStatus"`
				} `json:"status"`
			} `json:"items"`
		}
		if err := a.getJSON(ctx, strings.TrimRight(youtubeDataAPIBase, "/")+"/playlistItems?"+vals.Encode(), nil, &pr); err != nil {
			return seed, out, err
		}
		pages++
		allKnown := len(pr.Items) > 0
		for _, x := range pr.Items {
			id := firstNonEmpty(x.ContentDetails.VideoID, x.Snippet.ResourceID.VideoID)
			if id == "" || strings.EqualFold(x.Status.PrivacyStatus, "private") {
				continue
			}
			if !known[id] {
				allKnown = false
			}
			out = append(out, CreatorVideo{ID: id, Platform: "youtube", URL: "https://www.youtube.com/watch?v=" + id, Title: x.Snippet.Title, Creator: firstNonEmpty(seed.Title, x.Snippet.ChannelTitle), CreatorURL: seed.URL, CreatorAvatarURL: seed.AvatarURL, ThumbnailURL: bestYTThumb(x.Snippet.Thumbnails), Description: x.Snippet.Description, PublishedAt: x.ContentDetails.VideoPublishedAt, DiscoveredAt: now, ChannelID: seed.ChannelID, ChannelHandle: seed.Handle, VideoKind: youtubeKindFromMetadata(x.Snippet.Title, x.Snippet.Description, 0), Mods: []CreatorMod{}})
		}
		if pr.PageInfo.TotalResults > 0 {
			seed.TotalVideos = pr.PageInfo.TotalResults
		}
		if pr.NextPageToken == "" {
			break
		}
		if !full && pages >= 3 && allKnown {
			break
		}
		pageToken = pr.NextPageToken
	}
	if err := a.enrichYouTubeVideoDurations(ctx, key, out); err != nil { /* metadata remains sufficient */
	}
	return seed, dedupeCreatorVideos(out), nil
}

func bestYTThumb(m map[string]struct {
	URL string `json:"url"`
}) string {
	for _, k := range []string{"maxres", "standard", "high", "medium", "default"} {
		if x, ok := m[k]; ok && x.URL != "" {
			return x.URL
		}
	}
	return ""
}

func (a *App) enrichYouTubeVideoDurations(ctx context.Context, key string, videos []CreatorVideo) error {
	for start := 0; start < len(videos); start += 50 {
		end := start + 50
		if end > len(videos) {
			end = len(videos)
		}
		ids := []string{}
		index := map[string]int{}
		for i := start; i < end; i++ {
			ids = append(ids, videos[i].ID)
			index[videos[i].ID] = i
		}
		var vr struct {
			Items []struct {
				ID      string `json:"id"`
				Snippet struct {
					Title, Description, PublishedAt string
					Tags                            []string `json:"tags"`
				} `json:"snippet"`
				ContentDetails struct {
					Duration string `json:"duration"`
					Caption  string `json:"caption"`
				} `json:"contentDetails"`
			} `json:"items"`
		}
		vals := url.Values{"part": {"snippet,contentDetails"}, "id": {strings.Join(ids, ",")}, "key": {key}}
		if err := a.getJSON(ctx, strings.TrimRight(youtubeDataAPIBase, "/")+"/videos?"+vals.Encode(), nil, &vr); err != nil {
			return err
		}
		for _, x := range vr.Items {
			i, ok := index[x.ID]
			if !ok {
				continue
			}
			d := parseISODurationSeconds(x.ContentDetails.Duration)
			videos[i].DurationSeconds = d
			videos[i].CaptionHint = strings.EqualFold(x.ContentDetails.Caption, "true")
			videos[i].VideoKind = youtubeKindFromMetadata(x.Snippet.Title, x.Snippet.Description, d)
			if x.Snippet.Description != "" {
				videos[i].Description = x.Snippet.Description
			}
			if x.Snippet.PublishedAt != "" {
				videos[i].PublishedAt = x.Snippet.PublishedAt
			}
		}
	}
	return nil
}

func parseISODurationSeconds(raw string) int64 {
	re := regexp.MustCompile(`^P(?:(\d+)D)?T?(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?$`)
	m := re.FindStringSubmatch(strings.ToUpper(strings.TrimSpace(raw)))
	if len(m) == 0 {
		return 0
	}
	var n [4]int64
	for i := 1; i <= 4; i++ {
		n[i-1], _ = strconv.ParseInt(m[i], 10, 64)
	}
	return n[0]*86400 + n[1]*3600 + n[2]*60 + n[3]
}
func youtubeKindFromMetadata(title, description string, duration int64) string {
	low := strings.ToLower(title + " " + description)
	if strings.Contains(low, "#shorts") || strings.Contains(low, "#short ") || strings.HasSuffix(low, "#short") {
		return "short"
	}
	if duration > 0 && duration <= 180 {
		return "short"
	}
	return "video"
}

func (a *App) enumerateYouTubeChannelYTDLP(ctx context.Context, seed CreatorChannel, full bool) (CreatorChannel, []CreatorVideo, error) {
	ytdlp, err := a.ensureCreatorYTDLP(ctx)
	if err != nil {
		return seed, nil, err
	}
	handle := canonicalCreatorHandle(firstNonEmpty(seed.Handle, seed.URL))
	base := strings.TrimRight(seed.URL, "/")
	if handle != "" {
		base = "https://www.youtube.com/" + handle
		seed.Handle = handle
	} else if seed.ChannelID != "" {
		base = "https://www.youtube.com/channel/" + seed.ChannelID
	}
	if base == "" {
		return seed, nil, errors.New("channel URL, handle or ID is required")
	}
	seed.URL = base
	out := []CreatorVideo{}
	sections := []struct{ suffix, kind string }{{"/videos", "video"}, {"/shorts", "short"}, {"/streams", "stream"}}
	now := time.Now().UTC().Format(time.RFC3339)
	var errs []string
	for _, sec := range sections {
		args := []string{"--flat-playlist", "--dump-single-json", "--no-warnings", "--ignore-errors"}
		if !full {
			args = append(args, "--playlist-end", "80")
		}
		args = append(args, base+sec.suffix)
		cmd := exec.CommandContext(ctx, ytdlp, args...)
		b, e := cmd.Output()
		if e != nil {
			errs = append(errs, sec.kind+": "+e.Error())
			continue
		}
		var root ytDLPFlatEntry
		if json.Unmarshal(b, &root) != nil {
			errs = append(errs, sec.kind+": unreadable yt-dlp JSON")
			continue
		}
		if seed.Title == "" {
			seed.Title = firstNonEmpty(root.Channel, root.Uploader)
		}
		if seed.Bio == "" {
			seed.Bio = root.Description
		}
		if seed.ChannelID == "" {
			seed.ChannelID = firstNonEmpty(root.ChannelID, root.UploaderID)
		}
		if root.ChannelURL != "" {
			seed.URL = root.ChannelURL
		}
		if seed.Handle == "" {
			seed.Handle = canonicalCreatorHandle(firstNonEmpty(root.ChannelURL, root.UploaderID))
			handle = seed.Handle
		}
		for _, x := range root.Entries {
			if x.ID == "" {
				continue
			}
			published := ""
			if x.Timestamp > 0 {
				published = time.Unix(x.Timestamp, 0).UTC().Format(time.RFC3339)
			} else if len(x.UploadDate) == 8 {
				if t, e := time.Parse("20060102", x.UploadDate); e == nil {
					published = t.UTC().Format(time.RFC3339)
				}
			}
			kind := sec.kind
			if kind == "video" {
				kind = youtubeKindFromMetadata(x.Title, x.Description, int64(x.Duration))
			}
			out = append(out, CreatorVideo{ID: x.ID, Platform: "youtube", URL: "https://www.youtube.com/watch?v=" + x.ID, Title: x.Title, Creator: firstNonEmpty(x.Channel, x.Uploader, seed.Title), CreatorURL: seed.URL, ThumbnailURL: x.Thumbnail, Description: x.Description, PublishedAt: published, DiscoveredAt: now, ChannelID: firstNonEmpty(x.ChannelID, seed.ChannelID), ChannelHandle: handle, VideoKind: kind, DurationSeconds: int64(x.Duration), Mods: []CreatorMod{}})
		}
	}
	out = dedupeCreatorVideos(out)
	if full || seed.TotalVideos == 0 {
		seed.TotalVideos = len(out)
	}
	if len(out) == 0 {
		return seed, nil, errors.New(strings.Join(errs, "; "))
	}
	return seed, out, nil
}

func (a *App) enumerateTikTokChannelYTDLP(ctx context.Context, seed CreatorChannel, full bool) (CreatorChannel, []CreatorVideo, error) {
	ytdlp, err := a.ensureCreatorYTDLP(ctx)
	if err != nil {
		return seed, nil, err
	}
	handle := canonicalCreatorHandle(firstNonEmpty(seed.Handle, seed.URL))
	if handle == "" {
		return seed, nil, errors.New("TikTok creator handle is required")
	}
	seed.Platform = "tiktok"
	seed.Handle = handle
	seed.URL = "https://www.tiktok.com/" + handle
	args := []string{"--flat-playlist", "--dump-single-json", "--no-warnings", "--ignore-errors"}
	if !full {
		args = append(args, "--playlist-end", "80")
	}
	args = append(args, seed.URL)
	cmd := exec.CommandContext(ctx, ytdlp, args...)
	cmd.Env = append(cmd.Environ(), "PYTHONUTF8=1")
	b, err := cmd.Output()
	if err != nil {
		return seed, nil, fmt.Errorf("TikTok creator enumeration failed: %w", err)
	}
	var root ytDLPFlatEntry
	if err := json.Unmarshal(b, &root); err != nil {
		return seed, nil, fmt.Errorf("TikTok creator enumeration returned unreadable yt-dlp JSON: %w", err)
	}
	if seed.Title == "" {
		seed.Title = firstNonEmpty(root.Channel, root.Uploader, strings.TrimPrefix(handle, "@"))
	}
	if seed.Bio == "" {
		seed.Bio = root.Description
	}
	if seed.AvatarURL == "" {
		seed.AvatarURL = root.Thumbnail
	}
	now := time.Now().UTC().Format(time.RFC3339)
	out := make([]CreatorVideo, 0, len(root.Entries))
	for _, x := range root.Entries {
		if x.ID == "" {
			continue
		}
		published := ""
		if x.Timestamp > 0 {
			published = time.Unix(x.Timestamp, 0).UTC().Format(time.RFC3339)
		} else if len(x.UploadDate) == 8 {
			if t, e := time.Parse("20060102", x.UploadDate); e == nil {
				published = t.UTC().Format(time.RFC3339)
			}
		}
		videoURL := firstNonEmpty(x.WebpageURL, x.URL)
		if !strings.HasPrefix(videoURL, "http://") && !strings.HasPrefix(videoURL, "https://") {
			videoURL = "https://www.tiktok.com/" + handle + "/video/" + x.ID
		}
		out = append(out, CreatorVideo{
			ID:              x.ID,
			Platform:        "tiktok",
			URL:             videoURL,
			Title:           firstNonEmpty(x.Title, x.Description),
			Creator:         firstNonEmpty(x.Channel, x.Uploader, seed.Title, strings.TrimPrefix(handle, "@")),
			CreatorURL:      seed.URL,
			ThumbnailURL:    x.Thumbnail,
			Description:     x.Description,
			PublishedAt:     published,
			DiscoveredAt:    now,
			ChannelHandle:   handle,
			VideoKind:       "short",
			DurationSeconds: int64(x.Duration),
			Mods:            []CreatorMod{},
		})
	}
	out = dedupeCreatorVideos(out)
	if full || seed.TotalVideos == 0 {
		seed.TotalVideos = len(out)
	}
	if len(out) == 0 {
		return seed, nil, errors.New("TikTok creator enumeration returned zero videos")
	}
	return seed, out, nil
}

func dedupeCreatorVideos(in []CreatorVideo) []CreatorVideo {
	seen := map[string]int{}
	out := []CreatorVideo{}
	for _, v := range in {
		if v.ID == "" {
			continue
		}
		if i, ok := seen[v.ID]; ok {
			if out[i].VideoKind != "short" && v.VideoKind == "short" {
				out[i].VideoKind = "short"
			}
			continue
		}
		seen[v.ID] = len(out)
		out = append(out, v)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return firstNonEmpty(out[i].PublishedAt, out[i].DiscoveredAt) > firstNonEmpty(out[j].PublishedAt, out[j].DiscoveredAt)
	})
	return out
}

func (a *App) processCreatorChannelQueue(ctx context.Context, ch CreatorChannel, max int) error {
	processed := 0
	failures := []string{}
	for {
		a.dataMu.RLock()
		var next *CreatorVideo
		for i := range a.creatorVideos {
			v := a.creatorVideos[i]
			if !creatorVideoBelongsToChannel(v, ch) {
				continue
			}
			if v.AnalyzedAt != "" {
				continue
			}
			if v.AnalysisAttempts >= 5 && v.AnalysisError != "" {
				continue
			}
			if v.AnalysisError != "" && !creatorAnalysisRetryDue(v.LastAnalysisAttempt, 20*time.Minute) {
				continue
			}
			vv := v
			next = &vv
			break
		}
		a.dataMu.RUnlock()
		if next == nil || (max > 0 && processed >= max) {
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		oneCtx, cancel := context.WithTimeout(ctx, 40*time.Minute)
		video, err := a.analyzeCreatorVideo(oneCtx, next.URL)
		cancel()
		if err != nil {
			a.markCreatorAnalysisFailure(*next, err)
			failures = append(failures, next.ID+": "+err.Error())
			continue
		}
		video.ChannelHandle = firstNonEmpty(video.ChannelHandle, next.ChannelHandle, ch.Handle)
		video.ChannelID = firstNonEmpty(video.ChannelID, next.ChannelID, ch.ChannelID)
		video.VideoKind = firstNonEmpty(video.VideoKind, next.VideoKind)
		video.DurationSeconds = firstNonZero(video.DurationSeconds, next.DurationSeconds)
		video.PublishedAt = firstNonEmpty(video.PublishedAt, next.PublishedAt)
		video.AnalysisAttempts = next.AnalysisAttempts + 1
		video.AnalysisError = ""
		video.LastAnalysisAttempt = time.Now().UTC().Format(time.RFC3339)
		a.upsertCreatorVideo(video)
		_ = a.saveCreatorVideos()
		processed++
		a.dataMu.Lock()
		for i := range a.creatorChannels {
			if creatorChannelKey(a.creatorChannels[i]) == creatorChannelKey(ch) {
				a.creatorChannels[i].LastAnalyzeAt = time.Now().UTC().Format(time.RFC3339)
			}
		}
		a.refreshCreatorChannelStatsLocked()
		a.dataMu.Unlock()
		_ = a.saveCreatorChannels()
	}
	if processed == 0 && len(failures) > 0 {
		return errors.New(strings.Join(failures, "; "))
	}
	return nil
}

func creatorAnalysisRetryDue(lastAttempt string, cooldown time.Duration) bool {
	if strings.TrimSpace(lastAttempt) == "" {
		return true
	}
	t, err := time.Parse(time.RFC3339, lastAttempt)
	return err != nil || time.Since(t) >= cooldown
}

func (a *App) resetCreatorChannelFailures(ch CreatorChannel) {
	a.dataMu.Lock()
	changed := false
	for i := range a.creatorVideos {
		v := &a.creatorVideos[i]
		if !creatorVideoBelongsToChannel(*v, ch) {
			continue
		}
		if v.AnalyzedAt == "" && (v.AnalysisAttempts > 0 || v.AnalysisError != "") {
			v.AnalysisAttempts = 0
			v.AnalysisError = ""
			v.LastAnalysisAttempt = ""
			changed = true
		}
	}
	if changed {
		a.refreshCreatorChannelStatsLocked()
	}
	a.dataMu.Unlock()
	if changed {
		_ = a.saveCreatorVideos()
		_ = a.saveCreatorChannels()
	}
}

func firstNonZero(v ...int64) int64 {
	for _, x := range v {
		if x != 0 {
			return x
		}
	}
	return 0
}

func (a *App) markCreatorAnalysisFailure(v CreatorVideo, err error) {
	v.AnalysisAttempts++
	v.AnalysisError = truncate(err.Error(), 700)
	v.LastAnalysisAttempt = time.Now().UTC().Format(time.RFC3339)
	a.upsertCreatorVideo(v)
	_ = a.saveCreatorVideos()
}

func (a *App) backgroundCreatorArchiveLoop() {
	time.Sleep(4 * time.Second)
	for {
		a.ensureDefaultCreatorChannels()
		a.dataMu.RLock()
		channels := append([]CreatorChannel(nil), a.creatorChannels...)
		a.dataMu.RUnlock()
		a.mu.RLock()
		concurrency := a.settings.CreatorArchiveConcurrency
		refreshMinutes := a.settings.CreatorRefreshMinutes
		a.mu.RUnlock()
		if concurrency < 1 {
			concurrency = 1
		}
		if concurrency > 4 {
			concurrency = 4
		}
		if refreshMinutes < 30 {
			refreshMinutes = 30
		}
		slots := a.creatorSyncSlots(concurrency)
		for _, ch := range channels {
			if slots <= 0 {
				break
			}
			if ch.Paused {
				continue
			}
			due, needsFull := creatorChannelRefreshDecision(ch, refreshMinutes, time.Now())
			if due {
				// Channel enumeration shares the configured archive concurrency budget.
				// With the built-in creator corpus this prevents a fresh install from
				// spawning every yt-dlp/API crawl simultaneously while still draining
				// the complete queue automatically on subsequent 20-second passes.
				if a.startCreatorChannelSync(ch, needsFull, false) {
					slots--
				}
			}
		}

		sem := make(chan struct{}, concurrency)
		done := make(chan struct{}, len(channels))
		jobs := 0
		for _, ch := range channels {
			if ch.Paused || ch.LastFullSyncAt == "" {
				continue
			}
			jobs++
			go func(ch CreatorChannel) {
				sem <- struct{}{}
				defer func() { <-sem; done <- struct{}{} }()
				ctx, cancel := context.WithTimeout(context.Background(), 42*time.Minute)
				defer cancel()
				_ = a.processCreatorChannelQueue(ctx, ch, 1)
			}(ch)
		}
		for i := 0; i < jobs; i++ {
			<-done
		}

		// Preserve generic Creator Picks behavior for manually pasted or discovered
		// videos that are not part of a tracked archival channel.
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
		_ = a.processCreatorQueue(ctx, 1)
		cancel()
		time.Sleep(20 * time.Second)
	}
}

func creatorChannelRefreshDecision(ch CreatorChannel, refreshMinutes int, now time.Time) (due, full bool) {
	if refreshMinutes < 30 {
		refreshMinutes = 30
	}
	full = strings.TrimSpace(ch.LastFullSyncAt) == ""
	// Automatic retries back off after provider/network failures so a temporarily blocked
	// TikTok or YouTube endpoint is not hammered every background-loop pass. Manual Sync
	// remains immediate because it bypasses this scheduler decision entirely.
	if ch.SyncFailures > 0 {
		if attempted, err := time.Parse(time.RFC3339, ch.LastAttemptAt); err == nil && !attempted.IsZero() {
			shift := ch.SyncFailures - 1
			if shift > 6 {
				shift = 6
			}
			backoff := 5 * time.Minute * time.Duration(1<<shift)
			if backoff > 6*time.Hour {
				backoff = 6 * time.Hour
			}
			if now.Sub(attempted) < backoff {
				return false, full
			}
		}
	}
	if full {
		return true, true
	}
	lastSync, _ := time.Parse(time.RFC3339, ch.LastSyncedAt)
	if lastSync.IsZero() {
		return true, false
	}
	return now.Sub(lastSync) >= time.Duration(refreshMinutes)*time.Minute, false
}

func (a *App) handleCreatorRecommendations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	channel := strings.TrimSpace(r.URL.Query().Get("channel"))
	kind := strings.TrimSpace(r.URL.Query().Get("kind"))
	sortBy := firstNonEmpty(r.URL.Query().Get("sort"), "recommended-desc")
	a.dataMu.RLock()
	videos := append([]CreatorVideo(nil), a.creatorVideos...)
	a.dataMu.RUnlock()
	out := []CreatorRecommendation{}
	for _, v := range videos {
		if v.AnalyzedAt == "" {
			continue
		}
		if channel != "" && channel != creatorVideoChannelKey(v) && channel != v.ChannelID && !strings.EqualFold(channel, v.ChannelHandle) && !strings.EqualFold(channel, v.Creator) {
			continue
		}
		if kind != "" && kind != "all" && v.VideoKind != kind {
			continue
		}
		for _, m := range v.Mods {
			hay := strings.ToLower(strings.Join([]string{m.Name, m.Author, m.Provider, m.ProjectSummary, m.DescriptionContext, m.TranscriptContext, v.Title, v.Description}, " "))
			if q != "" && !strings.Contains(hay, q) {
				continue
			}
			out = append(out, CreatorRecommendation{ChannelHandle: v.ChannelHandle, ChannelTitle: v.Creator, ChannelURL: v.CreatorURL, VideoID: v.ID, VideoTitle: v.Title, VideoURL: v.URL, VideoKind: v.VideoKind, VideoPublishedAt: v.PublishedAt, VideoThumbnailURL: v.ThumbnailURL, Mod: m, RecommendationAt: v.PublishedAt})
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		switch sortBy {
		case "recommended-asc", "video-asc":
			if a.VideoPublishedAt == b.VideoPublishedAt {
				return a.Mod.TimestampS < b.Mod.TimestampS
			}
			return a.VideoPublishedAt < b.VideoPublishedAt
		case "name":
			return strings.ToLower(a.Mod.Name) < strings.ToLower(b.Mod.Name)
		case "channel":
			if strings.EqualFold(a.ChannelTitle, b.ChannelTitle) {
				return a.VideoPublishedAt > b.VideoPublishedAt
			}
			return strings.ToLower(a.ChannelTitle) < strings.ToLower(b.ChannelTitle)
		case "confidence":
			return a.Mod.Confidence > b.Mod.Confidence
		default:
			if a.VideoPublishedAt == b.VideoPublishedAt {
				return a.Mod.TimestampS > b.Mod.TimestampS
			}
			return a.VideoPublishedAt > b.VideoPublishedAt
		}
	})
	writeJSON(w, http.StatusOK, map[string]any{"recommendations": out, "count": len(out), "sort": sortBy, "archival": true})
}

func (a *App) handleCreatorTranscript(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimSpace(r.URL.Query().Get("videoId"))
	if id == "" {
		writeJSON(w, http.StatusBadRequest, APIError{Error: "videoId is required"})
		return
	}
	t, err := a.loadCreatorTranscript(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, APIError{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, t)
}
