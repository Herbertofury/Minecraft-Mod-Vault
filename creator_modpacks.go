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
	"strconv"
	"strings"
	"time"
)

// CreatorReleasedModpack is a provider-verified pack associated with a followed
// creator. The association is intentionally provenance-heavy: Creator Vault
// distinguishes an owner, collaborator/member, creator-profile listing, and a
// direct creator-controlled link instead of pretending every matching pack is
// owned by the creator.
type CreatorReleasedModpack struct {
	Provider           string   `json:"provider"`
	ProjectID          string   `json:"projectId,omitempty"`
	Slug               string   `json:"slug,omitempty"`
	Title              string   `json:"title"`
	Summary            string   `json:"summary,omitempty"`
	URL                string   `json:"url"`
	IconURL            string   `json:"iconUrl,omitempty"`
	Author             string   `json:"author,omitempty"`
	ProviderProfileURL string   `json:"providerProfileUrl,omitempty"`
	Relationship       string   `json:"relationship"`
	Downloads          int64    `json:"downloads,omitempty"`
	Updated            string   `json:"updated,omitempty"`
	GameVersions       []string `json:"gameVersions,omitempty"`
	Loaders            []string `json:"loaders,omitempty"`
	EvidenceURL        string   `json:"evidenceUrl,omitempty"`
	EvidenceType       string   `json:"evidenceType,omitempty"`
	VerifiedAt         string   `json:"verifiedAt,omitempty"`
	FirstSeenAt        string   `json:"firstSeenAt,omitempty"`
	LastSeenAt         string   `json:"lastSeenAt,omitempty"`
}

type creatorProviderProfileRef struct {
	Provider     string
	Username     string
	URL          string
	EvidenceURL  string
	EvidenceType string
}

type creatorModpackRef struct {
	Provider     string
	Slug         string
	URL          string
	Label        string
	EvidenceURL  string
	EvidenceType string
}

var creatorCurseForgeProjectCountRE = regexp.MustCompile(`(?i)([0-9][0-9,]*)\s+Projects?`)

func normalizeCreatorIdentityToken(raw string) string {
	raw = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(raw, "@")))
	var b strings.Builder
	for _, r := range raw {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func creatorIdentityAliases(ch CreatorChannel) map[string]bool {
	out := map[string]bool{}
	for _, raw := range []string{ch.Handle, ch.Title} {
		if v := normalizeCreatorIdentityToken(raw); v != "" {
			out[v] = true
		}
	}
	return out
}

func creatorIdentityMatches(ch CreatorChannel, raw string) bool {
	return creatorIdentityAliases(ch)[normalizeCreatorIdentityToken(raw)]
}

func creatorLinkClaimsOwnership(label string) bool {
	label = strings.ToLower(strings.Join(strings.Fields(label), " "))
	for _, phrase := range []string{"my modpack", "my mod pack", "my minecraft pack", "our modpack", "our mod pack", "official modpack", "official mod pack", "my pack"} {
		if strings.Contains(label, phrase) {
			return true
		}
	}
	return false
}

func creatorProviderProfileFromURL(raw string) (creatorProviderProfileRef, bool) {
	normalized := normalizeCreatorProfileLinkURL(raw, nil)
	u, err := url.Parse(normalized)
	if err != nil {
		return creatorProviderProfileRef{}, false
	}
	host := strings.ToLower(strings.TrimPrefix(u.Hostname(), "www."))
	parts := splitPath(u.Path)
	if (host == "curseforge.com" || strings.HasSuffix(host, ".curseforge.com")) && len(parts) >= 2 && strings.EqualFold(parts[0], "members") {
		username := strings.TrimSpace(parts[1])
		if username == "" {
			return creatorProviderProfileRef{}, false
		}
		return creatorProviderProfileRef{Provider: "CurseForge", Username: username, URL: "https://www.curseforge.com/members/" + url.PathEscape(username) + "/projects"}, true
	}
	if host == "modrinth.com" && len(parts) >= 2 && strings.EqualFold(parts[0], "user") {
		username := strings.TrimSpace(parts[1])
		if username == "" {
			return creatorProviderProfileRef{}, false
		}
		return creatorProviderProfileRef{Provider: "Modrinth", Username: username, URL: "https://modrinth.com/user/" + url.PathEscape(username)}, true
	}
	return creatorProviderProfileRef{}, false
}

func creatorModpackFromURL(raw string) (creatorModpackRef, bool) {
	normalized := normalizeCreatorProfileLinkURL(raw, nil)
	u, err := url.Parse(normalized)
	if err != nil {
		return creatorModpackRef{}, false
	}
	host := strings.ToLower(strings.TrimPrefix(u.Hostname(), "www."))
	parts := splitPath(u.Path)
	if (host == "curseforge.com" || strings.HasSuffix(host, ".curseforge.com")) && len(parts) >= 3 && strings.EqualFold(parts[0], "minecraft") && strings.EqualFold(parts[1], "modpacks") {
		slug := strings.TrimSpace(parts[2])
		if slug != "" {
			return creatorModpackRef{Provider: "CurseForge", Slug: slug, URL: "https://www.curseforge.com/minecraft/modpacks/" + url.PathEscape(slug)}, true
		}
	}
	if host == "modrinth.com" && len(parts) >= 2 && strings.EqualFold(parts[0], "modpack") {
		slug := strings.TrimSpace(parts[1])
		if slug != "" {
			return creatorModpackRef{Provider: "Modrinth", Slug: slug, URL: "https://modrinth.com/modpack/" + url.PathEscape(slug)}, true
		}
	}
	return creatorModpackRef{}, false
}

func creatorModpackKey(p CreatorReleasedModpack) string {
	id := firstNonEmpty(p.ProjectID, p.Slug, p.URL)
	return strings.ToLower(strings.TrimSpace(p.Provider)) + ":" + strings.ToLower(strings.TrimSpace(id))
}

func creatorRelationshipWeight(raw string) int {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "owner":
		return 0
	case "member", "collaborator":
		return 1
	case "profile":
		return 2
	case "linked":
		return 3
	default:
		return 4
	}
}

func creatorModpackSort(out []CreatorReleasedModpack) {
	sort.SliceStable(out, func(i, j int) bool {
		pi, pj := strings.ToLower(out[i].Provider), strings.ToLower(out[j].Provider)
		if pi != pj {
			return pi < pj
		}
		wi, wj := creatorRelationshipWeight(out[i].Relationship), creatorRelationshipWeight(out[j].Relationship)
		if wi != wj {
			return wi < wj
		}
		if out[i].Downloads != out[j].Downloads {
			return out[i].Downloads > out[j].Downloads
		}
		return strings.ToLower(out[i].Title) < strings.ToLower(out[j].Title)
	})
}

func creatorLinkAsProviderProfile(link CreatorProfileLink) (creatorProviderProfileRef, bool) {
	p, ok := creatorProviderProfileFromURL(link.URL)
	if !ok {
		return creatorProviderProfileRef{}, false
	}
	p.EvidenceURL = firstNonEmpty(link.EvidenceURL, link.URL)
	p.EvidenceType = firstNonEmpty(link.EvidenceType, "creator-link")
	return p, true
}

func creatorLinkAsDirectModpack(link CreatorProfileLink) (creatorModpackRef, bool) {
	m, ok := creatorModpackFromURL(link.URL)
	if !ok {
		return creatorModpackRef{}, false
	}
	m.Label = link.Label
	m.EvidenceURL = firstNonEmpty(link.EvidenceURL, link.URL)
	m.EvidenceType = firstNonEmpty(link.EvidenceType, "creator-link")
	return m, true
}

func appendCreatorProviderProfileLink(ch *CreatorChannel, p creatorProviderProfileRef) {
	if ch == nil || strings.TrimSpace(p.URL) == "" {
		return
	}
	n := normalizeCreatorProfileLinkURL(p.URL, nil)
	for i := range ch.ProfileLinks {
		if strings.EqualFold(ch.ProfileLinks[i].URL, n) {
			if ch.ProfileLinks[i].Kind == "" || ch.ProfileLinks[i].Kind == "website" {
				ch.ProfileLinks[i].Kind = "creator-profile"
			}
			return
		}
	}
	ch.ProfileLinks = append(ch.ProfileLinks, CreatorProfileLink{
		URL:          n,
		Label:        p.Provider + " creator profile",
		Kind:         "creator-profile",
		Provider:     p.Provider,
		EvidenceURL:  p.EvidenceURL,
		EvidenceType: p.EvidenceType,
	})
}

func (a *App) enrichDirectCreatorModpack(ctx context.Context, ch CreatorChannel, ref creatorModpackRef) (CreatorReleasedModpack, *creatorProviderProfileRef) {
	now := time.Now().UTC().Format(time.RFC3339)
	fallbackTitle := strings.TrimSpace(ref.Label)
	if fallbackTitle == "" || strings.EqualFold(fallbackTitle, "download") || strings.EqualFold(fallbackTitle, "install") {
		fallbackTitle = titleFromSlug(ref.Slug)
	}
	base := CreatorReleasedModpack{Provider: ref.Provider, Slug: ref.Slug, Title: fallbackTitle, URL: ref.URL, Relationship: "linked", EvidenceURL: ref.EvidenceURL, EvidenceType: ref.EvidenceType, VerifiedAt: now}
	aliases := creatorIdentityAliases(ch)
	if ref.Provider == "Modrinth" {
		var p struct {
			ID           string   `json:"id"`
			Slug         string   `json:"slug"`
			Title        string   `json:"title"`
			Description  string   `json:"description"`
			ProjectType  string   `json:"project_type"`
			Downloads    int64    `json:"downloads"`
			Updated      string   `json:"updated"`
			IconURL      string   `json:"icon_url"`
			GameVersions []string `json:"game_versions"`
			Loaders      []string `json:"loaders"`
		}
		if err := a.getJSON(ctx, modrinthAPIBase()+"/project/"+url.PathEscape(ref.Slug), nil, &p); err == nil && p.ProjectType == "modpack" {
			base.ProjectID, base.Slug, base.Title, base.Summary = p.ID, p.Slug, p.Title, p.Description
			base.Downloads, base.Updated, base.IconURL = p.Downloads, p.Updated, p.IconURL
			base.GameVersions, base.Loaders = uniqueStringsPreserve(p.GameVersions), uniqueStrings(p.Loaders)
			base.URL = "https://modrinth.com/modpack/" + url.PathEscape(firstNonEmpty(p.Slug, ref.Slug))
		}
		var members []struct {
			Role string `json:"role"`
			User struct {
				Username string `json:"username"`
			} `json:"user"`
		}
		if err := a.getJSON(ctx, modrinthAPIBase()+"/project/"+url.PathEscape(firstNonEmpty(base.ProjectID, ref.Slug))+"/members", nil, &members); err == nil {
			selectMember := func(m struct {
				Role string `json:"role"`
				User struct {
					Username string `json:"username"`
				} `json:"user"`
			}) (CreatorReleasedModpack, *creatorProviderProfileRef) {
				profile := creatorProviderProfileRef{Provider: "Modrinth", Username: m.User.Username, URL: "https://modrinth.com/user/" + url.PathEscape(m.User.Username), EvidenceURL: ref.URL, EvidenceType: "provider-project-author"}
				base.Author = m.User.Username
				base.ProviderProfileURL = profile.URL
				if strings.EqualFold(strings.TrimSpace(m.Role), "owner") {
					base.Relationship = "owner"
				} else {
					base.Relationship = "member"
				}
				return base, &profile
			}
			for _, m := range members {
				if aliases[normalizeCreatorIdentityToken(m.User.Username)] {
					return selectMember(m)
				}
			}
			if creatorLinkClaimsOwnership(ref.Label) {
				owners := make([]int, 0, 1)
				for i, m := range members {
					if strings.EqualFold(strings.TrimSpace(m.Role), "owner") {
						owners = append(owners, i)
					}
				}
				if len(owners) == 1 {
					return selectMember(members[owners[0]])
				}
			}
		}
		return base, nil
	}

	if ref.Provider == "CurseForge" {
		a.mu.RLock()
		key := strings.TrimSpace(a.settings.CurseForgeAPIKey)
		a.mu.RUnlock()
		if key != "" {
			projects, err := a.searchCurseForgeCreatorProjectBySlug(ctx, key, ref.Slug)
			if err == nil && len(projects) > 0 {
				p := projects[0]
				base = p.Modpack("linked", ref.EvidenceURL, ref.EvidenceType)
				selectAuthor := func(author creatorCurseForgeAuthor) (CreatorReleasedModpack, *creatorProviderProfileRef) {
					profileURL := firstNonEmpty(author.URL, "https://www.curseforge.com/members/"+url.PathEscape(author.Name)+"/projects")
					profile, ok := creatorProviderProfileFromURL(profileURL)
					if !ok {
						profile = creatorProviderProfileRef{Provider: "CurseForge", Username: author.Name, URL: "https://www.curseforge.com/members/" + url.PathEscape(author.Name) + "/projects"}
					}
					profile.EvidenceURL, profile.EvidenceType = ref.URL, "provider-project-author"
					base.Author, base.ProviderProfileURL = author.Name, profile.URL
					base.Relationship = "member"
					return base, &profile
				}
				for _, author := range p.Authors {
					if aliases[normalizeCreatorIdentityToken(author.Name)] {
						return selectAuthor(author)
					}
				}
				if creatorLinkClaimsOwnership(ref.Label) && len(p.Authors) == 1 {
					return selectAuthor(p.Authors[0])
				}
			}
		}
		if profile := a.discoverCurseForgeProfileFromProjectHTML(ctx, ch, ref); profile != nil {
			base.Author, base.ProviderProfileURL = profile.Username, profile.URL
			return base, profile
		}
	}
	return base, nil
}

type creatorCurseForgeAuthor struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type creatorCurseForgeProject struct {
	ID              int64                     `json:"id"`
	Name            string                    `json:"name"`
	Slug            string                    `json:"slug"`
	Summary         string                    `json:"summary"`
	DownloadCount   int64                     `json:"downloadCount"`
	DateModified    string                    `json:"dateModified"`
	ClassID         int64                     `json:"classId"`
	PrimaryAuthorID int64                     `json:"primaryAuthorId,omitempty"`
	Authors         []creatorCurseForgeAuthor `json:"authors"`
	Logo            *struct {
		ThumbnailURL string `json:"thumbnailUrl"`
		URL          string `json:"url"`
	} `json:"logo"`
	LatestFilesIndexes []struct {
		GameVersion string `json:"gameVersion"`
		ModLoader   int    `json:"modLoader"`
	} `json:"latestFilesIndexes"`
}

func (p creatorCurseForgeProject) Modpack(relationship, evidenceURL, evidenceType string) CreatorReleasedModpack {
	out := CreatorReleasedModpack{Provider: "CurseForge", ProjectID: strconv.FormatInt(p.ID, 10), Slug: p.Slug, Title: p.Name, Summary: p.Summary, URL: "https://www.curseforge.com/minecraft/modpacks/" + url.PathEscape(p.Slug), Relationship: relationship, Downloads: p.DownloadCount, Updated: p.DateModified, EvidenceURL: evidenceURL, EvidenceType: evidenceType, VerifiedAt: time.Now().UTC().Format(time.RFC3339)}
	if p.Logo != nil {
		out.IconURL = firstNonEmpty(p.Logo.ThumbnailURL, p.Logo.URL)
	}
	if len(p.Authors) > 0 {
		out.Author = p.Authors[0].Name
	}
	for _, idx := range p.LatestFilesIndexes {
		if idx.GameVersion != "" {
			out.GameVersions = append(out.GameVersions, idx.GameVersion)
		}
		if l := curseLoaderName(idx.ModLoader); l != "" {
			out.Loaders = append(out.Loaders, l)
		}
	}
	out.GameVersions = uniqueStringsPreserve(out.GameVersions)
	out.Loaders = uniqueStrings(out.Loaders)
	return out
}

func (a *App) curseForgeCreatorSearch(ctx context.Context, key string, values url.Values) ([]creatorCurseForgeProject, int64, error) {
	var resp struct {
		Data       []creatorCurseForgeProject `json:"data"`
		Pagination struct {
			TotalCount int64 `json:"totalCount"`
		} `json:"pagination"`
	}
	if err := a.getJSON(ctx, curseForgeAPIBase()+"/v1/mods/search?"+values.Encode(), map[string]string{"x-api-key": key}, &resp); err != nil {
		return nil, 0, err
	}
	return resp.Data, resp.Pagination.TotalCount, nil
}

func (a *App) searchCurseForgeCreatorProjectBySlug(ctx context.Context, key, slug string) ([]creatorCurseForgeProject, error) {
	values := url.Values{"gameId": {"432"}, "classId": {"4471"}, "slug": {slug}, "pageSize": {"10"}}
	out, _, err := a.curseForgeCreatorSearch(ctx, key, values)
	if err != nil {
		return nil, err
	}
	filtered := out[:0]
	for _, p := range out {
		if strings.EqualFold(p.Slug, slug) {
			filtered = append(filtered, p)
		}
	}
	return filtered, nil
}

func (a *App) resolveCurseForgeCreatorAuthorID(ctx context.Context, key, username string) (int64, error) {
	search := func(classID bool) (int64, error) {
		values := url.Values{"gameId": {"432"}, "searchFilter": {username}, "pageSize": {"50"}, "sortField": {"5"}, "sortOrder": {"asc"}}
		if classID {
			values.Set("classId", "4471")
		}
		projects, _, err := a.curseForgeCreatorSearch(ctx, key, values)
		if err != nil {
			return 0, err
		}
		for _, p := range projects {
			for _, author := range p.Authors {
				if author.ID != 0 && normalizeCreatorIdentityToken(author.Name) == normalizeCreatorIdentityToken(username) {
					return author.ID, nil
				}
			}
		}
		return 0, nil
	}
	if id, err := search(true); err != nil {
		return 0, err
	} else if id != 0 {
		return id, nil
	}
	if id, err := search(false); err != nil {
		return 0, err
	} else if id != 0 {
		return id, nil
	}
	return 0, errors.New("CurseForge author id could not be verified")
}

func (a *App) fetchCurseForgeCreatorProjectsByAuthor(ctx context.Context, key string, authorID int64, primary bool) ([]creatorCurseForgeProject, error) {
	if authorID == 0 {
		return nil, errors.New("CurseForge author id is required")
	}
	all := []creatorCurseForgeProject{}
	for index := 0; index < 10000 && len(all) < 500; index += 50 {
		values := url.Values{"gameId": {"432"}, "classId": {"4471"}, "pageSize": {"50"}, "index": {strconv.Itoa(index)}, "sortField": {"3"}, "sortOrder": {"desc"}}
		if primary {
			values.Set("primaryAuthorId", strconv.FormatInt(authorID, 10))
		} else {
			values.Set("authorId", strconv.FormatInt(authorID, 10))
		}
		page, total, err := a.curseForgeCreatorSearch(ctx, key, values)
		if err != nil {
			return nil, err
		}
		all = append(all, page...)
		if len(page) == 0 || total > 0 && int64(len(all)) >= total || len(page) < 50 {
			break
		}
	}
	return all, nil
}

func (a *App) fetchCurseForgeCreatorModpacksAPI(ctx context.Context, profile creatorProviderProfileRef) ([]CreatorReleasedModpack, error) {
	a.mu.RLock()
	key := strings.TrimSpace(a.settings.CurseForgeAPIKey)
	a.mu.RUnlock()
	if key == "" {
		return nil, errors.New("CurseForge API key is not configured")
	}
	authorID, err := a.resolveCurseForgeCreatorAuthorID(ctx, key, profile.Username)
	if err != nil {
		return nil, err
	}
	all, err := a.fetchCurseForgeCreatorProjectsByAuthor(ctx, key, authorID, false)
	if err != nil {
		return nil, err
	}
	owned, ownerErr := a.fetchCurseForgeCreatorProjectsByAuthor(ctx, key, authorID, true)
	ownerIDs := map[int64]bool{}
	if ownerErr == nil {
		for _, p := range owned {
			ownerIDs[p.ID] = true
		}
	}
	out := make([]CreatorReleasedModpack, 0, len(all))
	for _, p := range all {
		rel := "member"
		if ownerIDs[p.ID] {
			rel = "owner"
		}
		m := p.Modpack(rel, profile.URL, "provider-profile")
		m.Author = profile.Username
		m.ProviderProfileURL = profile.URL
		out = append(out, m)
	}
	return out, nil
}

func creatorCurseForgeFetchURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	base, err := url.Parse(curseForgeWebBase())
	if err != nil || base.Host == "" {
		return raw
	}
	u.Scheme, u.Host = base.Scheme, base.Host
	return u.String()
}

func creatorCurseForgePageFetchAllowed(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil {
		return false
	}
	base, err := url.Parse(curseForgeWebBase())
	if err != nil || !strings.EqualFold(u.Hostname(), base.Hostname()) {
		return false
	}
	parts := splitPath(u.Path)
	if len(parts) >= 2 && strings.EqualFold(parts[0], "members") {
		return true
	}
	return len(parts) >= 3 && strings.EqualFold(parts[0], "minecraft") && strings.EqualFold(parts[1], "modpacks")
}

func (a *App) fetchCreatorCurseForgePage(ctx context.Context, raw string) ([]byte, string, error) {
	fetchURL := creatorCurseForgeFetchURL(raw)
	if !creatorCurseForgePageFetchAllowed(fetchURL) {
		return nil, "", fmt.Errorf("creator modpack crawler will not fetch unsupported CurseForge page: %s", raw)
	}
	client := a.httpClient
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fetchURL, nil)
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
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, "", fmt.Errorf("CurseForge returned %s", res.Status)
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, 3<<20))
	if err != nil {
		return nil, "", err
	}
	if len(body) == 0 {
		return nil, "", errors.New("CurseForge creator page was empty")
	}
	return body, raw, nil
}

func creatorProfileProjectLinksFromHTML(body []byte, canonicalBase string) []creatorRawLink {
	base, _ := url.Parse(canonicalBase)
	content := cleanCreatorPageEscapes(string(body))
	seen := map[string]int{}
	out := []creatorRawLink{}
	add := func(raw, label string) {
		n := normalizeCreatorProfileLinkURL(html.UnescapeString(raw), base)
		if n == "" {
			return
		}
		if _, ok := creatorModpackFromURL(n); !ok {
			return
		}
		label = creatorLinkLabel(label)
		key := strings.ToLower(n)
		if idx, ok := seen[key]; ok {
			if len(label) > len(out[idx].Label) && !strings.EqualFold(label, "download") && !strings.EqualFold(label, "install") {
				out[idx].Label = label
			}
			return
		}
		seen[key] = len(out)
		out = append(out, creatorRawLink{URL: n, Label: label})
	}
	for _, m := range creatorAnchorRE.FindAllStringSubmatch(content, -1) {
		add(firstNonEmpty(m[1], m[2], m[3]), m[4])
	}
	// CurseForge's modern frontend can place project URLs inside serialized page
	// state even when a crawl receives few rendered anchors. Harvest those exact
	// canonical URLs as a fallback, but still admit only /minecraft/modpacks/.
	for _, raw := range creatorURLRE.FindAllString(content, -1) {
		add(raw, "")
	}
	return out
}

func parseCreatorProjectCount(body []byte) int {
	m := creatorCurseForgeProjectCountRE.FindSubmatch(body)
	if len(m) < 2 {
		return 0
	}
	n, _ := strconv.Atoi(strings.ReplaceAll(string(m[1]), ",", ""))
	return n
}

func (a *App) fetchCurseForgeCreatorModpacksHTML(ctx context.Context, profile creatorProviderProfileRef) ([]CreatorReleasedModpack, error) {
	baseProfile := profile.URL
	if !strings.HasSuffix(strings.TrimRight(baseProfile, "/"), "/projects") {
		baseProfile = strings.TrimRight(baseProfile, "/") + "/projects"
	}
	byURL := map[string]CreatorReleasedModpack{}
	totalExpected := 0
	for page := 1; page <= 6; page++ {
		pageURL := baseProfile
		if page > 1 {
			u, _ := url.Parse(baseProfile)
			q := u.Query()
			q.Set("page", strconv.Itoa(page))
			q.Set("pageSize", "50")
			u.RawQuery = q.Encode()
			pageURL = u.String()
		}
		body, _, err := a.fetchCreatorCurseForgePage(ctx, pageURL)
		if err != nil {
			if page == 1 {
				return nil, err
			}
			break
		}
		if page == 1 {
			totalExpected = parseCreatorProjectCount(body)
		}
		before := len(byURL)
		for _, raw := range creatorProfileProjectLinksFromHTML(body, baseProfile) {
			ref, ok := creatorModpackFromURL(raw.URL)
			if !ok {
				continue
			}
			title := strings.TrimSpace(raw.Label)
			if title == "" || strings.EqualFold(title, "download") || strings.EqualFold(title, "install") || strings.EqualFold(title, "view") {
				title = titleFromSlug(ref.Slug)
			}
			byURL[strings.ToLower(raw.URL)] = CreatorReleasedModpack{Provider: "CurseForge", Slug: ref.Slug, Title: title, URL: raw.URL, Author: profile.Username, ProviderProfileURL: profile.URL, Relationship: "profile", EvidenceURL: profile.URL, EvidenceType: "provider-profile-html", VerifiedAt: time.Now().UTC().Format(time.RFC3339)}
		}
		if len(byURL) == before || totalExpected > 0 && len(byURL) >= totalExpected || len(byURL) >= 250 {
			break
		}
	}
	out := make([]CreatorReleasedModpack, 0, len(byURL))
	for _, p := range byURL {
		out = append(out, p)
	}
	return out, nil
}

func (a *App) discoverCurseForgeProfileFromProjectHTML(ctx context.Context, ch CreatorChannel, ref creatorModpackRef) *creatorProviderProfileRef {
	body, _, err := a.fetchCreatorCurseForgePage(ctx, ref.URL)
	if err != nil {
		return nil
	}
	base, _ := url.Parse(ref.URL)
	profiles := map[string]creatorProviderProfileRef{}
	for _, m := range creatorAnchorRE.FindAllStringSubmatch(string(body), -1) {
		href := normalizeCreatorProfileLinkURL(firstNonEmpty(m[1], m[2], m[3]), base)
		profile, ok := creatorProviderProfileFromURL(href)
		if !ok || profile.Provider != "CurseForge" {
			continue
		}
		profile.EvidenceURL, profile.EvidenceType = ref.URL, "provider-project-author"
		profiles[strings.ToLower(profile.Username)] = profile
		if !creatorIdentityMatches(ch, profile.Username) {
			continue
		}
		return &profile
	}
	// A creator-controlled link explicitly labelled as their own modpack is strong
	// ownership evidence. If the public CurseForge project page exposes exactly one
	// unique member profile, that identity can be promoted without guessing that a
	// differently-named provider account matches the YouTube/TikTok handle. Multiple
	// member identities remain ambiguous and are intentionally not promoted.
	if creatorLinkClaimsOwnership(ref.Label) && len(profiles) == 1 {
		for _, profile := range profiles {
			p := profile
			return &p
		}
	}
	return nil
}

func (a *App) fetchModrinthCreatorModpacks(ctx context.Context, profile creatorProviderProfileRef) ([]CreatorReleasedModpack, error) {
	var projects []struct {
		ID           string   `json:"id"`
		Slug         string   `json:"slug"`
		Title        string   `json:"title"`
		Description  string   `json:"description"`
		ProjectType  string   `json:"project_type"`
		Status       string   `json:"status"`
		Downloads    int64    `json:"downloads"`
		Updated      string   `json:"updated"`
		IconURL      string   `json:"icon_url"`
		GameVersions []string `json:"game_versions"`
		Loaders      []string `json:"loaders"`
	}
	if err := a.getJSON(ctx, modrinthAPIBase()+"/user/"+url.PathEscape(profile.Username)+"/projects", nil, &projects); err != nil {
		return nil, err
	}
	out := []CreatorReleasedModpack{}
	for _, p := range projects {
		if p.ProjectType != "modpack" {
			continue
		}
		relationship := "member"
		var members []struct {
			Role string `json:"role"`
			User struct {
				Username string `json:"username"`
			} `json:"user"`
		}
		if err := a.getJSON(ctx, modrinthAPIBase()+"/project/"+url.PathEscape(p.ID)+"/members", nil, &members); err == nil {
			for _, member := range members {
				if normalizeCreatorIdentityToken(member.User.Username) != normalizeCreatorIdentityToken(profile.Username) {
					continue
				}
				if strings.EqualFold(strings.TrimSpace(member.Role), "owner") {
					relationship = "owner"
				} else if strings.TrimSpace(member.Role) != "" {
					relationship = "member"
				}
				break
			}
		}
		out = append(out, CreatorReleasedModpack{Provider: "Modrinth", ProjectID: p.ID, Slug: p.Slug, Title: p.Title, Summary: p.Description, URL: "https://modrinth.com/modpack/" + url.PathEscape(p.Slug), IconURL: p.IconURL, Author: profile.Username, ProviderProfileURL: profile.URL, Relationship: relationship, Downloads: p.Downloads, Updated: p.Updated, GameVersions: uniqueStringsPreserve(p.GameVersions), Loaders: uniqueStrings(p.Loaders), EvidenceURL: profile.URL, EvidenceType: "provider-profile", VerifiedAt: time.Now().UTC().Format(time.RFC3339)})
	}
	return out, nil
}

func (a *App) enumerateCreatorProviderProfile(ctx context.Context, profile creatorProviderProfileRef) ([]CreatorReleasedModpack, error) {
	switch profile.Provider {
	case "Modrinth":
		return a.fetchModrinthCreatorModpacks(ctx, profile)
	case "CurseForge":
		if packs, err := a.fetchCurseForgeCreatorModpacksAPI(ctx, profile); err == nil {
			return packs, nil
		}
		return a.fetchCurseForgeCreatorModpacksHTML(ctx, profile)
	default:
		return nil, fmt.Errorf("unsupported creator modpack provider %q", profile.Provider)
	}
}

func (a *App) refreshCreatorModpacks(ctx context.Context, ch CreatorChannel) (CreatorChannel, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	oldByKey := map[string]CreatorReleasedModpack{}
	for _, p := range ch.CreatorModpacks {
		oldByKey[creatorModpackKey(p)] = p
	}

	packs := map[string]CreatorReleasedModpack{}
	addPack := func(p CreatorReleasedModpack) {
		if p.Provider == "" || p.URL == "" || p.Title == "" {
			return
		}
		p.LastSeenAt = now
		if p.VerifiedAt == "" {
			p.VerifiedAt = now
		}
		if old := oldByKey[creatorModpackKey(p)]; p.FirstSeenAt == "" && old.FirstSeenAt != "" {
			p.FirstSeenAt = old.FirstSeenAt
		}
		if p.FirstSeenAt == "" {
			p.FirstSeenAt = now
		}
		key := creatorModpackKey(p)
		if existing, ok := packs[key]; ok {
			// Prefer richer provider-profile/API records over a bare direct link.
			if existing.ProjectID == "" && p.ProjectID != "" || existing.Downloads == 0 && p.Downloads > 0 || creatorRelationshipWeight(p.Relationship) < creatorRelationshipWeight(existing.Relationship) {
				p.FirstSeenAt = firstNonEmpty(existing.FirstSeenAt, p.FirstSeenAt)
				packs[key] = p
			}
			return
		}
		packs[key] = p
	}

	profileMap := map[string]creatorProviderProfileRef{}
	addProfile := func(p creatorProviderProfileRef) {
		if p.Provider == "" || p.URL == "" {
			return
		}
		key := strings.ToLower(p.Provider + ":" + p.URL)
		if _, ok := profileMap[key]; !ok {
			profileMap[key] = p
		}
	}

	directRefs := []creatorModpackRef{}
	for _, link := range ch.ProfileLinks {
		if p, ok := creatorLinkAsProviderProfile(link); ok {
			addProfile(p)
		}
		if m, ok := creatorLinkAsDirectModpack(link); ok {
			directRefs = append(directRefs, m)
		}
	}

	for _, ref := range directRefs {
		pack, discoveredProfile := a.enrichDirectCreatorModpack(ctx, ch, ref)
		addPack(pack)
		if discoveredProfile != nil {
			addProfile(*discoveredProfile)
			appendCreatorProviderProfileLink(&ch, *discoveredProfile)
		}
	}

	profileErrors := []string{}
	succeededProfiles := map[string]bool{}
	failedProfiles := map[string]bool{}
	profiles := make([]creatorProviderProfileRef, 0, len(profileMap))
	for _, p := range profileMap {
		profiles = append(profiles, p)
	}
	sort.SliceStable(profiles, func(i, j int) bool {
		if profiles[i].Provider != profiles[j].Provider {
			return profiles[i].Provider < profiles[j].Provider
		}
		return strings.ToLower(profiles[i].URL) < strings.ToLower(profiles[j].URL)
	})
	for _, profile := range profiles {
		profileCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		found, err := a.enumerateCreatorProviderProfile(profileCtx, profile)
		cancel()
		key := strings.ToLower(profile.Provider + ":" + profile.URL)
		if err != nil {
			failedProfiles[key] = true
			profileErrors = append(profileErrors, profile.Provider+" "+profile.Username+": "+err.Error())
			continue
		}
		succeededProfiles[key] = true
		appendCreatorProviderProfileLink(&ch, profile)
		for _, p := range found {
			addPack(p)
		}
	}

	// A failed provider refresh must never erase a previously verified library.
	// Preserve only records tied to the failed profile; successful profile scans
	// are authoritative and therefore naturally retire stale entries.
	for _, old := range ch.CreatorModpacks {
		profile, ok := creatorProviderProfileFromURL(old.ProviderProfileURL)
		if !ok {
			continue
		}
		key := strings.ToLower(profile.Provider + ":" + profile.URL)
		if failedProfiles[key] && !succeededProfiles[key] {
			addPack(old)
		}
	}

	out := make([]CreatorReleasedModpack, 0, len(packs))
	for _, p := range packs {
		out = append(out, p)
	}
	creatorModpackSort(out)
	ch.CreatorModpacks = out
	ch.CreatorModpacksRefreshedAt = now
	ch.CreatorModpacksError = strings.Join(profileErrors, "; ")
	switch {
	case len(profileErrors) > 0 && len(out) > 0:
		ch.CreatorModpacksStatus = "partial"
	case len(profileErrors) > 0:
		ch.CreatorModpacksStatus = "error"
	case len(out) == 0:
		ch.CreatorModpacksStatus = "none"
	default:
		ch.CreatorModpacksStatus = "ready"
	}
	if len(profileErrors) > 0 && len(out) == 0 {
		return ch, errors.New(ch.CreatorModpacksError)
	}
	return ch, nil
}

func creatorModpackStatusCounts(packs []CreatorReleasedModpack) map[string]int {
	out := map[string]int{"total": len(packs)}
	for _, p := range packs {
		out[strings.ToLower(p.Provider)]++
		out[strings.ToLower(p.Relationship)]++
	}
	return out
}

func (a *App) handleCreatorModpacksRefresh(w http.ResponseWriter, r *http.Request) {
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
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	updated, refreshErr := a.refreshCreatorModpacks(ctx, target)
	if err := a.persistCreatorProfileMetadata(updated); err != nil {
		writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
		return
	}
	if refreshErr != nil {
		writeJSON(w, http.StatusBadGateway, APIError{Error: "Creator modpack refresh kept any last known-good records: " + refreshErr.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"creator": updated, "modpacks": updated.CreatorModpacks, "count": len(updated.CreatorModpacks), "counts": creatorModpackStatusCounts(updated.CreatorModpacks), "refreshed": true})
}
