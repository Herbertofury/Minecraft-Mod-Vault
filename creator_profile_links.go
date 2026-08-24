package main

import (
	"context"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
)

// CreatorProfileLink is creator-controlled profile metadata discovered from a
// public creator profile, a public link-in-bio hub, or a curated seed. The Vault
// stores the evidence page that exposed each URL so a user can tell why a link
// is associated with a creator without pretending the destination itself was
// independently endorsed or security-audited.
type CreatorProfileLink struct {
	URL          string `json:"url"`
	Label        string `json:"label,omitempty"`
	Kind         string `json:"kind"`
	Provider     string `json:"provider,omitempty"`
	EvidenceURL  string `json:"evidenceUrl,omitempty"`
	EvidenceType string `json:"evidenceType,omitempty"`
	FirstSeenAt  string `json:"firstSeenAt,omitempty"`
	LastSeenAt   string `json:"lastSeenAt,omitempty"`
}

type creatorPageRef struct {
	URL          string
	EvidenceType string
}

type creatorRawLink struct {
	URL   string
	Label string
}

var creatorAnchorRE = regexp.MustCompile(`(?is)<a\b[^>]*?\bhref\s*=\s*(?:"([^"]+)"|'([^']+)'|([^\s>]+))[^>]*>(.*?)</a>`)
var creatorTagRE = regexp.MustCompile(`(?is)<[^>]+>`)
var creatorURLRE = regexp.MustCompile(`https?://[^\s"'<>\\]+`)

var creatorLinkHubDomains = []string{
	"linktr.ee", "lnk.bio", "beacons.ai", "carrd.co", "campsite.bio", "solo.to",
	"bio.site", "taplink.cc", "allmylinks.com", "msha.ke", "hoo.be", "bio.link",
	"stan.store", "pillar.io", "snipfeed.co", "withkoji.com", "flow.page",
}

func creatorHostMatches(host string, domains []string) bool {
	host = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(host), "."))
	for _, domain := range domains {
		domain = strings.ToLower(strings.TrimPrefix(domain, "."))
		if host == domain || strings.HasSuffix(host, "."+domain) {
			return true
		}
	}
	return false
}

func isCreatorLinkHubURL(raw string) bool {
	u, err := url.Parse(strings.TrimSpace(raw))
	return err == nil && creatorHostMatches(u.Hostname(), creatorLinkHubDomains)
}

func creatorPublicPageFetchAllowed(raw string) bool {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.User != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Hostname() == "" {
		return false
	}
	host := strings.ToLower(u.Hostname())
	if creatorHostMatches(host, creatorLinkHubDomains) {
		return true
	}
	return creatorHostMatches(host, []string{"youtube.com", "tiktok.com"})
}

func cleanCreatorPageEscapes(raw string) string {
	r := strings.NewReplacer(
		`\/`, `/`,
		`\u002F`, `/`, `\u002f`, `/`,
		`\u003A`, `:`, `\u003a`, `:`,
		`\u0026`, `&`,
		`\u003D`, `=`, `\u003d`, `=`,
		`\u0025`, `%`,
	)
	return html.UnescapeString(r.Replace(raw))
}

func normalizeCreatorProfileLinkURL(raw string, base *url.URL) string {
	raw = strings.TrimSpace(cleanCreatorPageEscapes(raw))
	raw = strings.Trim(raw, "\"'()[]{}<>,.;")
	if raw == "" || strings.HasPrefix(strings.ToLower(raw), "javascript:") || strings.HasPrefix(strings.ToLower(raw), "data:") {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	if !u.IsAbs() && base != nil {
		u = base.ResolveReference(u)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return ""
	}
	if u.User != nil || u.Hostname() == "" {
		return ""
	}
	// Creator/profile platforms commonly wrap external bio destinations in a
	// same-origin redirect URL. Unwrap an explicit target parameter without
	// fetching that target so Linktree/Modrinth/etc. still become first-class
	// creator links while the crawler remains profile/hub-only.
	if creatorPublicPageFetchAllowed(u.String()) {
		q := u.Query()
		for _, key := range []string{"q", "url", "target", "destination", "redirect", "redirect_url"} {
			if target := strings.TrimSpace(q.Get(key)); strings.HasPrefix(target, "https://") || strings.HasPrefix(target, "http://") {
				if nested := normalizeCreatorProfileLinkURL(target, nil); nested != "" {
					return nested
				}
			}
		}
	}
	u.Fragment = ""
	q := u.Query()
	for key := range q {
		lk := strings.ToLower(key)
		if strings.HasPrefix(lk, "utm_") || lk == "fbclid" || lk == "gclid" || lk == "mc_cid" || lk == "mc_eid" {
			q.Del(key)
		}
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func creatorLinkLooksLikeAsset(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return true
	}
	path := strings.ToLower(u.Path)
	for _, ext := range []string{".js", ".css", ".png", ".jpg", ".jpeg", ".gif", ".webp", ".svg", ".ico", ".woff", ".woff2", ".ttf", ".map", ".json", ".xml"} {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	host := strings.ToLower(u.Hostname())
	return creatorHostMatches(host, []string{"google-analytics.com", "googletagmanager.com", "gstatic.com", "doubleclick.net", "cloudfront.net", "sentry.io"})
}

func creatorProfileLinkKind(rawURL, label string) string {
	u, _ := url.Parse(rawURL)
	host, path := strings.ToLower(u.Hostname()), strings.ToLower(u.Path)
	// URL structure wins over button copy. A Linktree button named "My Modpacks"
	// is still a creator hub, while the child Drive/CurseForge destination can be
	// classified as the actual modpack. This prevents profile/library links from
	// masquerading as projects simply because of their label.
	if (strings.Contains(host, "curseforge.com") && strings.HasPrefix(path, "/members/")) || (strings.Contains(host, "modrinth.com") && strings.HasPrefix(path, "/user/")) {
		return "creator-profile"
	}
	if creatorHostMatches(host, creatorLinkHubDomains) {
		return "link-hub"
	}
	if strings.Contains(host, "curseforge.com") {
		switch {
		case strings.Contains(path, "/minecraft/modpacks"):
			return "modpack"
		case strings.Contains(path, "/minecraft/mc-mods"):
			return "mod"
		case strings.Contains(path, "/minecraft/texture-packs") || strings.Contains(path, "/minecraft/resource-packs"):
			return "resource-pack"
		case strings.Contains(path, "/minecraft/shaders"):
			return "shader"
		}
	}
	if strings.Contains(host, "modrinth.com") {
		switch {
		case strings.HasPrefix(path, "/modpack/"):
			return "modpack"
		case strings.HasPrefix(path, "/mod/"):
			return "mod"
		case strings.HasPrefix(path, "/resourcepack/"):
			return "resource-pack"
		case strings.HasPrefix(path, "/shader/"):
			return "shader"
		case strings.HasPrefix(path, "/datapack/"):
			return "datapack"
		}
	}

	lowLabel := strings.ToLower(strings.TrimSpace(label))
	if strings.Contains(lowLabel, "modpack") || strings.Contains(lowLabel, "mod pack") || strings.Contains(lowLabel, "minecraft pack") {
		return "modpack"
	}
	if strings.Contains(lowLabel, "resource pack") || strings.Contains(lowLabel, "texture pack") {
		return "resource-pack"
	}
	if strings.Contains(lowLabel, "shader") {
		return "shader"
	}
	if strings.Contains(lowLabel, "data pack") || strings.Contains(lowLabel, "datapack") {
		return "datapack"
	}
	if strings.Contains(lowLabel, "mod list") || strings.Contains(lowLabel, "mods list") {
		return "mod-list"
	}
	if strings.Contains(lowLabel, "wishlist") || strings.Contains(lowLabel, "wish list") {
		return "wishlist"
	}
	if strings.Contains(lowLabel, "tip") || strings.Contains(lowLabel, "dono") || strings.Contains(lowLabel, "donat") || strings.Contains(lowLabel, "support me") {
		return "support"
	}
	if creatorHostMatches(host, []string{"technicpack.net", "atlauncher.com", "feed-the-beast.com", "ftb.com"}) && strings.Contains(path, "pack") {
		return "modpack"
	}
	if creatorHostMatches(host, []string{"discord.gg", "discord.com", "twitch.tv", "youtube.com", "youtu.be", "instagram.com", "tiktok.com", "twitter.com", "x.com", "bsky.app", "threads.net", "facebook.com"}) {
		return "social"
	}
	if creatorHostMatches(host, []string{"ko-fi.com", "patreon.com", "paypal.me", "paypal.com", "cash.app", "venmo.com", "streamlabs.com", "streamelements.com", "buymeacoffee.com"}) {
		return "support"
	}
	if strings.Contains(host, "amazon.") && strings.Contains(path, "wishlist") || creatorHostMatches(host, []string{"throne.com"}) {
		return "wishlist"
	}
	if creatorHostMatches(host, []string{"mediafire.com", "drive.google.com", "dropbox.com", "mega.nz"}) {
		return "download"
	}
	return "website"
}

func creatorProfileLinkProvider(rawURL string) string {
	u, _ := url.Parse(rawURL)
	host := strings.ToLower(strings.TrimPrefix(u.Hostname(), "www."))
	providers := []struct {
		Domain string
		Name   string
	}{
		{"linktr.ee", "Linktree"}, {"lnk.bio", "Lnk.Bio"}, {"beacons.ai", "Beacons"}, {"carrd.co", "Carrd"},
		{"curseforge.com", "CurseForge"}, {"modrinth.com", "Modrinth"}, {"discord.gg", "Discord"}, {"discord.com", "Discord"},
		{"twitch.tv", "Twitch"}, {"youtube.com", "YouTube"}, {"youtu.be", "YouTube"}, {"instagram.com", "Instagram"},
		{"tiktok.com", "TikTok"}, {"twitter.com", "X / Twitter"}, {"x.com", "X"}, {"patreon.com", "Patreon"},
		{"ko-fi.com", "Ko-fi"}, {"throne.com", "Throne"}, {"github.com", "GitHub"}, {"drive.google.com", "Google Drive"},
	}
	for _, p := range providers {
		if host == p.Domain || strings.HasSuffix(host, "."+p.Domain) {
			return p.Name
		}
	}
	if host == "" {
		return ""
	}
	parts := strings.Split(host, ".")
	if len(parts) >= 2 {
		return creatorDisplayHostPart(parts[len(parts)-2])
	}
	return creatorDisplayHostPart(host)
}

func creatorDisplayHostPart(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	runes := []rune(raw)
	if runes[0] >= 'a' && runes[0] <= 'z' {
		runes[0] -= 'a' - 'A'
	}
	return string(runes)
}

func creatorLinkLabel(raw string) string {
	raw = html.UnescapeString(creatorTagRE.ReplaceAllString(raw, " "))
	return strings.Join(strings.Fields(raw), " ")
}

func extractCreatorRawLinks(content string, baseRaw string) []creatorRawLink {
	content = cleanCreatorPageEscapes(content)
	base, _ := url.Parse(baseRaw)
	seen := map[string]int{}
	out := []creatorRawLink{}
	add := func(raw, label string) {
		normalized := normalizeCreatorProfileLinkURL(raw, base)
		if normalized == "" || creatorLinkLooksLikeAsset(normalized) {
			return
		}
		if base != nil {
			u, _ := url.Parse(normalized)
			if strings.EqualFold(u.Hostname(), base.Hostname()) && !isCreatorLinkHubURL(normalized) {
				return
			}
		}
		label = creatorLinkLabel(label)
		if i, ok := seen[normalized]; ok {
			if out[i].Label == "" && label != "" {
				out[i].Label = label
			}
			return
		}
		seen[normalized] = len(out)
		out = append(out, creatorRawLink{URL: normalized, Label: label})
	}
	for _, m := range creatorAnchorRE.FindAllStringSubmatch(content, -1) {
		href := firstNonEmpty(m[1], m[2], m[3])
		add(href, m[4])
	}
	for _, raw := range creatorURLRE.FindAllString(content, -1) {
		add(raw, "")
	}
	return out
}

func creatorLinkSortWeight(kind string) int {
	switch kind {
	case "modpack":
		return 0
	case "mod", "mod-list":
		return 1
	case "resource-pack", "shader", "datapack":
		return 2
	case "creator-profile":
		return 3
	case "link-hub":
		return 4
	case "social":
		return 5
	case "download":
		return 6
	case "support", "wishlist":
		return 7
	default:
		return 8
	}
}

func (a *App) fetchCreatorPublicPage(ctx context.Context, raw string) ([]byte, string, error) {
	if !creatorPublicPageFetchAllowed(raw) {
		return nil, "", fmt.Errorf("creator profile crawler will not fetch unsupported host: %s", raw)
	}
	base := a.httpClient
	if base == nil {
		base = &http.Client{Timeout: 20 * time.Second}
	}
	client := *base
	if client.Timeout <= 0 || client.Timeout > 30*time.Second {
		client.Timeout = 30 * time.Second
	}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return errors.New("too many creator-profile redirects")
		}
		if !creatorPublicPageFetchAllowed(req.URL.String()) {
			// Keep the redirect response but never issue the outbound request. The
			// Location is useful creator metadata even when its host is not a page
			// the crawler is allowed to fetch.
			return http.ErrUseLastResponse
		}
		return nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml;q=0.9,*/*;q=0.5")
	res, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 && res.StatusCode < 400 {
		if location := normalizeCreatorProfileLinkURL(res.Header.Get("Location"), res.Request.URL); location != "" {
			return []byte(`<a href="` + html.EscapeString(location) + `">Redirect target</a>`), res.Request.URL.String(), nil
		}
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, "", fmt.Errorf("%s returned %s", hostLabel(raw), res.Status)
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, 2<<20))
	if err != nil {
		return nil, "", err
	}
	if len(body) == 0 {
		return nil, "", errors.New("creator profile page was empty")
	}
	return body, res.Request.URL.String(), nil
}

func mergeCreatorProfileSeed(dst *CreatorChannel, seed CreatorChannel) bool {
	changed := false
	hubs := append([]string(nil), dst.ProfileHubURLs...)
	for _, raw := range seed.ProfileHubURLs {
		n := normalizeCreatorProfileLinkURL(raw, nil)
		if n == "" {
			continue
		}
		found := false
		for _, h := range hubs {
			if strings.EqualFold(h, n) {
				found = true
				break
			}
		}
		if !found {
			hubs = append(hubs, n)
			changed = true
		}
	}
	if changed {
		dst.ProfileHubURLs = hubs
	}
	if dst.Bio == "" && seed.Bio != "" {
		dst.Bio = seed.Bio
		changed = true
	}
	if dst.ProfileLinksStatus == "" && seed.ProfileLinksStatus != "" {
		dst.ProfileLinksStatus = seed.ProfileLinksStatus
		changed = true
	}
	byURL := map[string]int{}
	for i := range dst.ProfileLinks {
		byURL[strings.ToLower(dst.ProfileLinks[i].URL)] = i
	}
	for _, link := range seed.ProfileLinks {
		link.URL = normalizeCreatorProfileLinkURL(link.URL, nil)
		if link.URL == "" {
			continue
		}
		key := strings.ToLower(link.URL)
		if _, ok := byURL[key]; ok {
			continue
		}
		if link.Kind == "" {
			link.Kind = creatorProfileLinkKind(link.URL, link.Label)
		}
		if link.Provider == "" {
			link.Provider = creatorProfileLinkProvider(link.URL)
		}
		if link.EvidenceType == "" {
			link.EvidenceType = "seed"
		}
		dst.ProfileLinks = append(dst.ProfileLinks, link)
		byURL[key] = len(dst.ProfileLinks) - 1
		changed = true
	}
	return changed
}

func (a *App) refreshCreatorProfileLinks(ctx context.Context, ch CreatorChannel) (CreatorChannel, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	oldFirstSeen := map[string]string{}
	for _, link := range ch.ProfileLinks {
		oldFirstSeen[strings.ToLower(link.URL)] = link.FirstSeenAt
	}
	seedLinks := []CreatorProfileLink{}
	for _, link := range ch.ProfileLinks {
		if link.EvidenceType == "seed" {
			seedLinks = append(seedLinks, link)
		}
	}
	for _, hub := range ch.ProfileHubURLs {
		seedLinks = append(seedLinks, CreatorProfileLink{URL: hub, Label: "Creator link hub", Kind: "link-hub", EvidenceURL: ch.URL, EvidenceType: "seed"})
	}

	links := map[string]CreatorProfileLink{}
	add := func(link CreatorProfileLink) {
		link.URL = normalizeCreatorProfileLinkURL(link.URL, nil)
		if link.URL == "" || strings.EqualFold(strings.TrimRight(link.URL, "/"), strings.TrimRight(ch.URL, "/")) {
			return
		}
		if link.Kind == "" {
			link.Kind = creatorProfileLinkKind(link.URL, link.Label)
		}
		if link.Provider == "" {
			link.Provider = creatorProfileLinkProvider(link.URL)
		}
		link.LastSeenAt = now
		if link.FirstSeenAt == "" {
			link.FirstSeenAt = firstNonEmpty(oldFirstSeen[strings.ToLower(link.URL)], now)
		}
		key := strings.ToLower(link.URL)
		if existing, ok := links[key]; ok {
			if existing.Label == "" && link.Label != "" {
				existing.Label = link.Label
			}
			if existing.EvidenceType == "seed" && link.EvidenceType != "seed" {
				existing.EvidenceType = link.EvidenceType
				existing.EvidenceURL = link.EvidenceURL
			}
			links[key] = existing
			return
		}
		links[key] = link
	}
	for _, link := range seedLinks {
		add(link)
	}

	queue := []creatorPageRef{}
	if creatorPublicPageFetchAllowed(ch.URL) {
		queue = append(queue, creatorPageRef{URL: ch.URL, EvidenceType: "profile"})
	}
	for _, hub := range ch.ProfileHubURLs {
		if isCreatorLinkHubURL(hub) {
			queue = append(queue, creatorPageRef{URL: hub, EvidenceType: "link-hub"})
		}
	}
	for _, raw := range extractCreatorRawLinks(ch.Bio, ch.URL) {
		add(CreatorProfileLink{URL: raw.URL, Label: raw.Label, EvidenceURL: ch.URL, EvidenceType: "bio"})
		if isCreatorLinkHubURL(raw.URL) {
			queue = append(queue, creatorPageRef{URL: raw.URL, EvidenceType: "link-hub"})
		}
	}

	// Creators often put their Linktree/Beacons/modpack URL in upload
	// descriptions even when the platform profile HTML is sparse or protected.
	// Sample recent archived uploads as an additional creator-controlled signal.
	a.dataMu.RLock()
	evidenceVideos := make([]CreatorVideo, 0, 24)
	for i := len(a.creatorVideos) - 1; i >= 0 && len(evidenceVideos) < 24; i-- {
		v := a.creatorVideos[i]
		if creatorVideoBelongsToChannel(v, ch) && strings.TrimSpace(v.Description) != "" {
			evidenceVideos = append(evidenceVideos, v)
		}
	}
	a.dataMu.RUnlock()
	for _, v := range evidenceVideos {
		for _, raw := range extractCreatorRawLinks(v.Description, v.URL) {
			add(CreatorProfileLink{URL: raw.URL, Label: raw.Label, EvidenceURL: v.URL, EvidenceType: "video-description"})
			if isCreatorLinkHubURL(raw.URL) {
				queue = append(queue, creatorPageRef{URL: raw.URL, EvidenceType: "link-hub"})
			}
		}
	}

	visited := map[string]bool{}
	fetched := 0
	var fetchErrors []string
	for len(queue) > 0 && len(visited) < 6 {
		ref := queue[0]
		queue = queue[1:]
		norm := normalizeCreatorProfileLinkURL(ref.URL, nil)
		if norm == "" || visited[strings.ToLower(norm)] || !creatorPublicPageFetchAllowed(norm) {
			continue
		}
		visited[strings.ToLower(norm)] = true
		pageCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
		body, finalURL, err := a.fetchCreatorPublicPage(pageCtx, norm)
		cancel()
		if err != nil {
			fetchErrors = append(fetchErrors, hostLabel(norm)+": "+err.Error())
			continue
		}
		fetched++
		for _, raw := range extractCreatorRawLinks(string(body), finalURL) {
			link := CreatorProfileLink{URL: raw.URL, Label: raw.Label, EvidenceURL: finalURL, EvidenceType: ref.EvidenceType}
			add(link)
			if isCreatorLinkHubURL(raw.URL) && !visited[strings.ToLower(raw.URL)] {
				queue = append(queue, creatorPageRef{URL: raw.URL, EvidenceType: "link-hub"})
			}
		}
	}

	if fetched == 0 && len(fetchErrors) > 0 {
		ch.ProfileLinksStatus = "cached"
		ch.ProfileLinksError = strings.Join(fetchErrors, "; ")
		// Keep the last known-good links on a blocked/rate-limited refresh; do not
		// make a transient provider failure look like the creator removed them.
		if len(ch.ProfileLinks) > 0 {
			return ch, errors.New(ch.ProfileLinksError)
		}
	}

	out := make([]CreatorProfileLink, 0, len(links))
	hubs := []string{}
	for _, link := range links {
		out = append(out, link)
		if link.Kind == "link-hub" {
			hubs = append(hubs, link.URL)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		wi, wj := creatorLinkSortWeight(out[i].Kind), creatorLinkSortWeight(out[j].Kind)
		if wi != wj {
			return wi < wj
		}
		return strings.ToLower(firstNonEmpty(out[i].Label, out[i].Provider, out[i].URL)) < strings.ToLower(firstNonEmpty(out[j].Label, out[j].Provider, out[j].URL))
	})
	sort.Strings(hubs)
	ch.ProfileLinks = out
	ch.ProfileHubURLs = hubs
	ch.ProfileLinksRefreshedAt = now
	ch.ProfileLinksStatus = "ready"
	ch.ProfileLinksError = ""
	return ch, nil
}

func (a *App) persistCreatorProfileMetadata(updated CreatorChannel) error {
	a.dataMu.Lock()
	for i := range a.creatorChannels {
		if creatorChannelsEquivalent(a.creatorChannels[i], updated) {
			a.creatorChannels[i].Bio = updated.Bio
			a.creatorChannels[i].ProfileHubURLs = append([]string(nil), updated.ProfileHubURLs...)
			a.creatorChannels[i].ProfileLinks = append([]CreatorProfileLink(nil), updated.ProfileLinks...)
			a.creatorChannels[i].ProfileLinksStatus = updated.ProfileLinksStatus
			a.creatorChannels[i].ProfileLinksError = updated.ProfileLinksError
			a.creatorChannels[i].ProfileLinksRefreshedAt = updated.ProfileLinksRefreshedAt
			a.creatorChannels[i].CreatorModpacks = append([]CreatorReleasedModpack(nil), updated.CreatorModpacks...)
			a.creatorChannels[i].CreatorModpacksStatus = updated.CreatorModpacksStatus
			a.creatorChannels[i].CreatorModpacksError = updated.CreatorModpacksError
			a.creatorChannels[i].CreatorModpacksRefreshedAt = updated.CreatorModpacksRefreshedAt
			break
		}
	}
	a.dataMu.Unlock()
	return a.saveCreatorChannels()
}

func (a *App) handleCreatorProfileLinksRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeJSON(w, http.StatusMethodNotAllowed, APIError{Error: "POST required"})
		return
	}
	var in struct {
		URL string `json:"url"`
	}
	if err := decodeJSON(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	seed, err := normalizeCreatorChannelInput(in.URL)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	a.dataMu.RLock()
	var target CreatorChannel
	for _, ch := range a.creatorChannels {
		if creatorChannelsEquivalent(ch, seed) {
			target = ch
			break
		}
	}
	a.dataMu.RUnlock()
	if target.URL == "" {
		writeJSON(w, http.StatusNotFound, APIError{Error: "creator is not followed"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 45*time.Second)
	defer cancel()
	updated, linksErr := a.refreshCreatorProfileLinks(ctx, target)
	updated, modpacksErr := a.refreshCreatorModpacks(ctx, updated)
	if err := a.persistCreatorProfileMetadata(updated); err != nil {
		writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
		return
	}
	if linksErr != nil || modpacksErr != nil {
		parts := []string{}
		if linksErr != nil {
			parts = append(parts, "profile links: "+linksErr.Error())
		}
		if modpacksErr != nil {
			parts = append(parts, "creator modpacks: "+modpacksErr.Error())
		}
		writeJSON(w, http.StatusBadGateway, APIError{Error: "Creator metadata refresh kept last known-good data where available: " + strings.Join(parts, "; ")})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"creator": updated, "links": updated.ProfileLinks, "count": len(updated.ProfileLinks), "modpacks": updated.CreatorModpacks, "modpackCount": len(updated.CreatorModpacks), "refreshed": true})
}
