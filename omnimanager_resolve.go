package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type libraryModrinthProject struct {
	ID           string   `json:"id"`
	Slug         string   `json:"slug"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	IconURL      string   `json:"icon_url"`
	ProjectType  string   `json:"project_type"`
	GameVersions []string `json:"game_versions"`
	Loaders      []string `json:"loaders"`
	SourceURL    string   `json:"source_url"`
	WikiURL      string   `json:"wiki_url"`
	IssuesURL    string   `json:"issues_url"`
}

type libraryCurseProject struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Slug          string `json:"slug"`
	Summary       string `json:"summary"`
	DownloadCount int64  `json:"downloadCount"`
	DateModified  string `json:"dateModified"`
	Logo          struct {
		ThumbnailURL string `json:"thumbnailUrl"`
		URL          string `json:"url"`
	} `json:"logo"`
	Authors []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"authors"`
	Links struct {
		WebsiteURL string `json:"websiteUrl"`
		WikiURL    string `json:"wikiUrl"`
		SourceURL  string `json:"sourceUrl"`
	} `json:"links"`
}

type libraryCurseMatch struct {
	Fingerprint uint32
	ModID       int64
	FileID      int64
	FileName    string
	DisplayName string
}

func (a *App) enrichLibrary(ctx context.Context, items []LibraryItem, refresh bool, cache *libraryCacheFile) error {
	if cache == nil {
		value := a.loadLibraryCache()
		cache = &value
	}
	var errs []string
	if err := a.enrichLibraryModrinth(ctx, items); err != nil {
		errs = append(errs, "Modrinth: "+err.Error())
	}
	if err := a.enrichLibraryCurseForge(ctx, items); err != nil {
		errs = append(errs, "CurseForge: "+err.Error())
	}
	if err := a.enrichDeclaredLibrarySources(ctx, items); err != nil {
		errs = append(errs, "declared sources: "+err.Error())
	}
	if err := a.enrichLibrarySearchMatches(ctx, items, refresh); err != nil {
		errs = append(errs, "cross-source search: "+err.Error())
	}
	for i := range items {
		finalizeLibraryIdentity(&items[i])
		cache.Records[items[i].ID] = cacheRecordFromItem(items[i])
	}
	if err := a.saveLibraryCache(*cache); err != nil {
		errs = append(errs, "cache: "+err.Error())
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

func (a *App) enrichLibraryModrinth(ctx context.Context, items []LibraryItem) error {
	type group struct {
		indices []int
		locals  []LocalModFile
		game    string
		loader  string
	}
	a.mu.RLock()
	defaultGame, defaultLoader := a.settings.GameVersion, a.settings.Loader
	a.mu.RUnlock()
	groups := map[string]*group{}
	for i := range items {
		item := &items[i]
		if item.Hashes.SHA512 == "" || item.Edition == "bedrock" || (item.Kind != "mod" && item.Kind != "plugin") {
			continue
		}
		loader := defaultLoader
		if item.Kind == "plugin" {
			loader = ""
		} else if len(item.Loaders) > 0 && containsFold([]string{"fabric", "forge", "neoforge", "quilt"}, item.Loaders[0]) {
			loader = item.Loaders[0]
		}
		game := defaultGame
		key := game + "\x1f" + loader
		g := groups[key]
		if g == nil {
			g = &group{game: game, loader: loader}
			groups[key] = g
		}
		g.indices = append(g.indices, i)
		g.locals = append(g.locals, localModFromLibraryItem(*item))
	}
	projectIDs := map[string]bool{}
	type resolved struct {
		index  int
		exact  ModrinthVersion
		update ModrinthVersion
	}
	resolvedItems := []resolved{}
	for _, g := range groups {
		exact, updates := a.lookupModrinthUpdates(ctx, g.locals, g.game, g.loader)
		for localIndex, local := range g.locals {
			version, ok := exact[local.SHA512]
			if !ok {
				continue
			}
			projectIDs[version.ProjectID] = true
			resolvedItems = append(resolvedItems, resolved{index: g.indices[localIndex], exact: version, update: updates[local.SHA512]})
		}
	}
	projects, err := a.fetchLibraryModrinthProjects(ctx, mapKeys(projectIDs))
	if err != nil && len(resolvedItems) == 0 {
		return err
	}
	for _, resolved := range resolvedItems {
		item := &items[resolved.index]
		project := projects[resolved.exact.ProjectID]
		source := LibrarySource{
			Provider: "modrinth", ProviderLabel: "Modrinth", ProjectID: resolved.exact.ProjectID, Slug: project.Slug,
			Title: firstNonEmpty(project.Title, item.Name), IconURL: project.IconURL,
			PageURL:       "https://modrinth.com/" + firstNonEmpty(project.ProjectType, "mod") + "/" + firstNonEmpty(project.Slug, resolved.exact.ProjectID),
			LatestVersion: resolved.update.VersionNumber, GameVersions: uniqueStringsPreserve(append(project.GameVersions, resolved.update.GameVersions...)),
			Loaders: uniqueStrings(append(project.Loaders, resolved.update.Loaders...)), Exact: true, Confidence: 1,
			Evidence: "Exact SHA-512 match returned by Modrinth's version-file identity API",
		}
		item.Sources = mergeLibrarySources(item.Sources, []LibrarySource{source})
		item.MatchEvidence = uniqueStringsPreserve(append(item.MatchEvidence, source.Evidence))
		item.ProvenanceConfidence = 1
		if project.Title != "" {
			item.Name = project.Title
		}
		item.Summary = firstNonEmpty(item.Summary, project.Description)
		item.RemoteArtURL = firstNonEmpty(item.RemoteArtURL, project.IconURL)
		item.InstalledVersion = firstNonEmpty(resolved.exact.VersionNumber, item.InstalledVersion)
		if resolved.update.ID == "" || resolved.update.ID == resolved.exact.ID {
			if item.UpdateStatus != "modified" {
				item.UpdateStatus = "current"
				item.UpdateMessage = "Exact Modrinth file identity is already the newest compatible release for this target."
			}
			item.LatestVersion = firstNonEmpty(resolved.exact.VersionNumber, item.LatestVersion)
			continue
		}
		candidate := modrinthVersionCandidate(resolved.update, ModrinthProject{ID: project.ID, Slug: project.Slug, Title: project.Title, ProjectType: project.ProjectType, Description: project.Description})
		candidate.Safe = true
		candidate.Confidence = 1
		candidate.Reason = "Exact installed SHA-512 identity plus provider-declared target game/loader compatibility"
		item.LatestVersion = candidate.Version
		if item.UpdateStatus == "modified" {
			candidate.Safe = false
			candidate.Reason += "; local filename/metadata indicates a custom or patched build, so replacement requires review"
			item.Alternatives = prependCandidate(item.Alternatives, candidate)
			item.UpdateMessage = "A compatible Modrinth release exists, but this custom build is protected from automatic replacement."
		} else {
			item.SafeUpdate = &candidate
			item.UpdateStatus = "update"
			item.UpdateMessage = "Exact Modrinth identity found a newer file for this Minecraft version and loader."
		}
	}
	return err
}

func localModFromLibraryItem(item LibraryItem) LocalModFile {
	return LocalModFile{
		Path: item.Path, Filename: item.Filename, Enabled: item.Enabled, Size: item.Size,
		SHA1: item.Hashes.SHA1, SHA512: item.Hashes.SHA512, CurseFingerprint: item.Hashes.CurseFingerprint,
		Metadata: JarMetadata{ModID: item.ModID, Name: item.Name, Version: item.InstalledVersion, Loaders: item.Loaders, SourceURL: item.SourceURL, Homepage: item.Homepage, Minecraft: firstString(item.GameVersions), Authors: item.Authors, Description: item.Description, License: item.License, MetadataBy: item.MetadataBy},
	}
}

func mapKeys(values map[string]bool) []string {
	out := make([]string, 0, len(values))
	for key := range values {
		if key != "" {
			out = append(out, key)
		}
	}
	sort.Strings(out)
	return out
}

func (a *App) fetchLibraryModrinthProjects(ctx context.Context, ids []string) (map[string]libraryModrinthProject, error) {
	out := map[string]libraryModrinthProject{}
	if len(ids) == 0 {
		return out, nil
	}
	for start := 0; start < len(ids); start += 100 {
		end := start + 100
		if end > len(ids) {
			end = len(ids)
		}
		encoded, _ := json.Marshal(ids[start:end])
		var projects []libraryModrinthProject
		endpoint := modrinthAPIBase() + "/projects?ids=" + url.QueryEscape(string(encoded))
		if err := a.getJSON(ctx, endpoint, nil, &projects); err != nil {
			return out, err
		}
		for _, project := range projects {
			out[project.ID] = project
		}
	}
	return out, nil
}

func (a *App) enrichLibraryCurseForge(ctx context.Context, items []LibraryItem) error {
	a.mu.RLock()
	key, game, defaultLoader := strings.TrimSpace(a.settings.CurseForgeAPIKey), a.settings.GameVersion, a.settings.Loader
	a.mu.RUnlock()
	if key == "" {
		return nil
	}
	fingerprints := []uint32{}
	indexByFingerprint := map[uint32][]int{}
	for i := range items {
		if items[i].Hashes.CurseFingerprint == 0 || items[i].Edition == "bedrock" || (items[i].Kind != "mod" && items[i].Kind != "plugin") {
			continue
		}
		fingerprint := items[i].Hashes.CurseFingerprint
		if len(indexByFingerprint[fingerprint]) == 0 {
			fingerprints = append(fingerprints, fingerprint)
		}
		indexByFingerprint[fingerprint] = append(indexByFingerprint[fingerprint], i)
	}
	if len(fingerprints) == 0 {
		return nil
	}
	var response struct {
		Data struct {
			ExactMatches []struct {
				ID   uint32 `json:"id"`
				File struct {
					ID          int64  `json:"id"`
					ModID       int64  `json:"modId"`
					FileName    string `json:"fileName"`
					DisplayName string `json:"displayName"`
				} `json:"file"`
			} `json:"exactMatches"`
		} `json:"data"`
	}
	if err := a.postJSON(ctx, curseForgeAPIBase()+"/v1/fingerprints", map[string]string{"x-api-key": key}, map[string]any{"fingerprints": fingerprints}, &response); err != nil {
		return err
	}
	matches := []libraryCurseMatch{}
	modIDs := map[int64]bool{}
	for _, match := range response.Data.ExactMatches {
		matches = append(matches, libraryCurseMatch{Fingerprint: match.ID, ModID: match.File.ModID, FileID: match.File.ID, FileName: match.File.FileName, DisplayName: match.File.DisplayName})
		modIDs[match.File.ModID] = true
	}
	projects, projectErr := a.fetchLibraryCurseProjects(ctx, key, int64Keys(modIDs))
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 6)
	var updateErrors []string
	for _, match := range matches {
		match := match
		for _, index := range indexByFingerprint[match.Fingerprint] {
			index := index
			wg.Add(1)
			go func() {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()
				loader := defaultLoader
				if len(items[index].Loaders) > 0 {
					loader = items[index].Loaders[0]
				}
				candidate, found := a.curseForgeCompatibleCandidate(ctx, key, match.ModID, match.FileID, game, loader)
				mu.Lock()
				defer mu.Unlock()
				item := &items[index]
				project := projects[match.ModID]
				author := ""
				if len(project.Authors) > 0 {
					author = project.Authors[0].Name
				}
				pageURL := firstNonEmpty(project.Links.WebsiteURL, "https://www.curseforge.com/minecraft/mc-mods/"+project.Slug)
				source := LibrarySource{
					Provider: "curseforge", ProviderLabel: "CurseForge", ProjectID: strconv.FormatInt(match.ModID, 10), Slug: project.Slug,
					Title: firstNonEmpty(project.Name, match.DisplayName, item.Name), Author: author,
					IconURL: firstNonEmpty(project.Logo.ThumbnailURL, project.Logo.URL), PageURL: pageURL,
					Exact: true, Confidence: 1, Evidence: "Exact CurseForge MurmurHash2 fingerprint match",
				}
				if found {
					source.LatestVersion = candidate.Version
					source.GameVersions = candidate.GameVersions
					source.Loaders = candidate.Loaders
				}
				item.Sources = mergeLibrarySources(item.Sources, []LibrarySource{source})
				item.MatchEvidence = uniqueStringsPreserve(append(item.MatchEvidence, source.Evidence))
				item.ProvenanceConfidence = 1
				item.Name = firstNonEmpty(project.Name, match.DisplayName, item.Name)
				item.Summary = firstNonEmpty(item.Summary, project.Summary)
				item.Authors = uniqueStringsPreserve(append(item.Authors, author))
				item.RemoteArtURL = firstNonEmpty(item.RemoteArtURL, source.IconURL)
				if found {
					item.LatestVersion = candidate.Version
					if candidate.ID == fmt.Sprintf("curseforge:%d", match.FileID) {
						if item.UpdateStatus != "modified" && item.UpdateStatus != "update" {
							item.UpdateStatus = "current"
							item.UpdateMessage = "Exact CurseForge file identity is already the newest compatible release."
						}
					} else if item.UpdateStatus == "modified" || candidate.DependencyRisk {
						candidate.Safe = false
						candidate.Confidence = 1
						candidate.Reason = "Exact CurseForge project identity; review required because the local build is modified or the target changes required dependencies"
						item.Alternatives = prependCandidate(item.Alternatives, candidate)
						if item.UpdateStatus != "modified" {
							item.UpdateStatus = "review"
						}
						item.UpdateMessage = "CurseForge has a compatible file, but Vault requires review before changing this build."
					} else if item.SafeUpdate == nil {
						candidate.Safe = true
						candidate.Confidence = 1
						candidate.Reason = "Exact CurseForge fingerprint plus provider-declared target game/loader compatibility"
						item.SafeUpdate = &candidate
						item.UpdateStatus = "update"
						item.UpdateMessage = "Exact CurseForge identity found a newer compatible file."
					}
				} else {
					updateErrors = append(updateErrors, item.Name+": no compatible CurseForge file for target")
				}
			}()
		}
	}
	wg.Wait()
	if projectErr != nil {
		return projectErr
	}
	if len(updateErrors) > 0 && len(updateErrors) == len(matches) {
		return errors.New(strings.Join(updateErrors[:minInt(len(updateErrors), 4)], "; "))
	}
	return nil
}

func int64Keys(values map[int64]bool) []int64 {
	out := make([]int64, 0, len(values))
	for key := range values {
		if key > 0 {
			out = append(out, key)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func (a *App) fetchLibraryCurseProjects(ctx context.Context, key string, ids []int64) (map[int64]libraryCurseProject, error) {
	out := map[int64]libraryCurseProject{}
	if len(ids) == 0 {
		return out, nil
	}
	for start := 0; start < len(ids); start += 50 {
		end := start + 50
		if end > len(ids) {
			end = len(ids)
		}
		var response struct {
			Data []libraryCurseProject `json:"data"`
		}
		err := a.postJSON(ctx, curseForgeAPIBase()+"/v1/mods", map[string]string{"x-api-key": key}, map[string]any{"modIds": ids[start:end]}, &response)
		if err != nil {
			for _, id := range ids[start:end] {
				var single struct {
					Data libraryCurseProject `json:"data"`
				}
				if getErr := a.getJSON(ctx, fmt.Sprintf("%s/v1/mods/%d", curseForgeAPIBase(), id), map[string]string{"x-api-key": key}, &single); getErr == nil {
					out[single.Data.ID] = single.Data
				}
			}
			if len(out) == 0 {
				return out, err
			}
			continue
		}
		for _, project := range response.Data {
			out[project.ID] = project
		}
	}
	return out, nil
}

func (a *App) enrichDeclaredLibrarySources(ctx context.Context, items []LibraryItem) error {
	var errs []string
	for i := range items {
		item := &items[i]
		declared := uniqueStringsPreserve(nonEmptyStrings(item.SourceURL, item.Homepage))
		for _, raw := range declared {
			provider, id, canonical, ok := integratedProviderFromURL(raw)
			if !ok {
				continue
			}
			if hasLibraryProvider(item.Sources, provider) {
				continue
			}
			detail, err := a.fetchProjectDetails(ctx, provider, id, canonical)
			if err != nil {
				errs = append(errs, provider+": "+err.Error())
				continue
			}
			author := ""
			if len(detail.Authors) > 0 {
				author = detail.Authors[0].Name
			}
			latest := ""
			if len(detail.Versions) > 0 {
				latest = firstNonEmpty(detail.Versions[0].Version, detail.Versions[0].Name)
			}
			source := LibrarySource{
				Provider: provider, ProviderLabel: detail.ProviderLabel, ProjectID: detail.ID, Slug: detail.Slug,
				Title: detail.Title, Author: author, IconURL: detail.IconURL, PageURL: detail.PageURL,
				LatestVersion: latest, GameVersions: detail.GameVersions, Loaders: detail.Loaders,
				Exact: true, Confidence: .995, Evidence: "Developer-declared canonical provider URL embedded in local metadata",
			}
			item.Sources = mergeLibrarySources(item.Sources, []LibrarySource{source})
			item.MatchEvidence = uniqueStringsPreserve(append(item.MatchEvidence, source.Evidence))
			item.ProvenanceConfidence = maxFloat(item.ProvenanceConfidence, .995)
			item.Name = firstNonEmpty(detail.Title, item.Name)
			item.Summary = firstNonEmpty(item.Summary, detail.Summary)
			item.Authors = uniqueStringsPreserve(append(item.Authors, author))
			item.RemoteArtURL = firstNonEmpty(item.RemoteArtURL, detail.IconURL)
			item.LatestVersion = firstNonEmpty(item.LatestVersion, latest)
		}
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs[:minInt(len(errs), 5)], "; "))
	}
	return nil
}

func (a *App) enrichLibrarySearchMatches(ctx context.Context, items []LibraryItem, refresh bool) error {
	type job struct{ index int }
	jobs := []job{}
	for i := range items {
		item := &items[i]
		if item.Name == "" || item.Kind == "world" || item.Kind == "world-package" || item.Kind == "world-bedrock" {
			continue
		}
		hasExact := false
		for _, source := range item.Sources {
			if source.Exact {
				hasExact = true
				break
			}
		}
		// Search unresolved items and a bounded number of exact items to discover
		// alternate storefronts without turning a 500-mod scan into thousands of calls.
		if !hasExact || (refresh && len(jobs) < 80) {
			jobs = append(jobs, job{index: i})
		}
		if len(jobs) >= 120 {
			break
		}
	}
	if len(jobs) == 0 {
		return nil
	}
	var mu sync.Mutex
	var errs []string
	sem := make(chan struct{}, 4)
	var wg sync.WaitGroup
	for _, current := range jobs {
		current := current
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			item := items[current.index]
			sources := librarySearchSources(item)
			if len(sources) == 0 {
				return
			}
			projectType := libraryProjectType(item)
			a.mu.RLock()
			game, loader := a.settings.GameVersion, a.settings.Loader
			a.mu.RUnlock()
			if item.Edition == "bedrock" {
				game, loader = "", "bedrock"
			}
			response := a.searchProviders(ctx, providerSearchOptions{Query: item.Name, GameVersion: game, Loader: loader, ProjectType: projectType, Limit: 12, Sources: sources})
			if len(response.Results) == 0 {
				if len(response.Errors) > 0 {
					mu.Lock()
					errs = append(errs, item.Name+": provider search unavailable")
					mu.Unlock()
				}
				return
			}
			matches := scoreLibrarySearchMatches(item, response.Results)
			if len(matches) == 0 {
				return
			}
			mu.Lock()
			defer mu.Unlock()
			target := &items[current.index]
			for _, source := range matches {
				if hasLibraryProviderProject(target.Sources, source.Provider, source.ProjectID) {
					continue
				}
				target.Sources = mergeLibrarySources(target.Sources, []LibrarySource{source})
				if target.RemoteArtURL == "" && source.Confidence >= .86 {
					target.RemoteArtURL = source.IconURL
				}
				if len(target.Authors) == 0 && source.Confidence >= .92 && source.Author != "" {
					target.Authors = append(target.Authors, source.Author)
				}
			}
			if target.UpdateStatus == "unknown" && matches[0].Confidence >= .92 {
				target.UpdateStatus = "review"
				target.UpdateMessage = "Vault found a high-confidence cross-site project match, but no exact artifact fingerprint; review before replacing the local file."
				target.ProvenanceConfidence = maxFloat(target.ProvenanceConfidence, matches[0].Confidence)
				target.MatchEvidence = uniqueStringsPreserve(append(target.MatchEvidence, matches[0].Evidence))
			}
		}()
	}
	wg.Wait()
	if len(errs) > 0 && len(errs) == len(jobs) {
		return errors.New(strings.Join(errs[:minInt(len(errs), 4)], "; "))
	}
	return nil
}

func librarySearchSources(item LibraryItem) []string {
	switch item.Edition {
	case "bedrock":
		return []string{"mcpedl", "marketplace", "curseforge", "planetminecraft", "github"}
	case "server":
		return []string{"hangar", "spigot", "polymart", "builtbybit", "github", "modrinth", "curseforge"}
	default:
		switch item.Kind {
		case "resourcepack", "shader", "datapack":
			return []string{"modrinth", "curseforge", "planetminecraft", "github", "resourcepacknet", "texturepacks", "shaderpacksnet", "shaderpackscom"}
		default:
			return []string{"modrinth", "curseforge", "github", "planetminecraft", "moddb", "mcreator"}
		}
	}
}

func libraryProjectType(item LibraryItem) string {
	switch item.Kind {
	case "resourcepack":
		return "resourcepack"
	case "shader":
		return "shader"
	case "datapack":
		return "datapack"
	case "plugin":
		return "plugin"
	case "world", "world-package":
		return "world"
	case "skinpack":
		return "skin"
	case "behaviorpack", "addon-package", "resourcepack-bedrock", "resourcepack-bedrock-dev", "behaviorpack-dev":
		return "addon"
	default:
		return "mod"
	}
}

func scoreLibrarySearchMatches(item LibraryItem, projects []UnifiedProject) []LibrarySource {
	out := []LibrarySource{}
	for _, project := range projects {
		titleScore := titleSimilarity(item.Name, project.Title)
		idScore := 0.0
		if item.ModID != "" && (strings.EqualFold(item.ModID, project.Slug) || strings.Contains(strings.ToLower(project.PageURL), "/"+strings.ToLower(item.ModID))) {
			idScore = 1
		}
		authorScore := 0.0
		for _, author := range item.Authors {
			if author != "" && project.Author != "" && titleSimilarity(author, project.Author) >= .75 {
				authorScore = 1
				break
			}
		}
		confidence := .52 + titleScore*.38 + idScore*.08 + authorScore*.05
		if strings.EqualFold(strings.TrimSpace(item.Name), strings.TrimSpace(project.Title)) {
			confidence = maxFloat(confidence, .95)
		}
		if confidence < .74 {
			continue
		}
		if confidence > .985 {
			confidence = .985
		}
		out = append(out, LibrarySource{
			Provider: project.Provider, ProviderLabel: providerDisplayName(project.Provider), ProjectID: project.ID, Slug: project.Slug,
			Title: project.Title, Author: project.Author, IconURL: project.IconURL, PageURL: project.PageURL,
			GameVersions: project.Versions, Loaders: project.Loaders, Exact: false, Confidence: confidence,
			Evidence: fmt.Sprintf("Cross-source title/mod-id/author match scored %.3f; not treated as exact artifact identity", confidence),
		})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Confidence > out[j].Confidence })
	if len(out) > 6 {
		out = out[:6]
	}
	return out
}

func providerDisplayName(id string) string {
	if info := providerInfoByID(id); info != nil {
		return info.Name
	}
	if id == "" {
		return "Unknown"
	}
	return strings.ToUpper(id[:1]) + id[1:]
}

func hasLibraryProvider(sources []LibrarySource, provider string) bool {
	for _, source := range sources {
		if strings.EqualFold(source.Provider, provider) {
			return true
		}
	}
	return false
}

func hasLibraryProviderProject(sources []LibrarySource, provider, projectID string) bool {
	for _, source := range sources {
		if strings.EqualFold(source.Provider, provider) && strings.EqualFold(source.ProjectID, projectID) {
			return true
		}
	}
	return false
}

func prependCandidate(existing []UpdateCandidate, candidate UpdateCandidate) []UpdateCandidate {
	for _, current := range existing {
		if current.ID == candidate.ID {
			return existing
		}
	}
	return append([]UpdateCandidate{candidate}, existing...)
}

func finalizeLibraryIdentity(item *LibraryItem) {
	if item == nil {
		return
	}
	item.Authors = uniqueStringsPreserve(nonEmptyStrings(item.Authors...))
	item.Loaders = uniqueStrings(item.Loaders)
	item.GameVersions = uniqueStringsPreserve(nonEmptyStrings(item.GameVersions...))
	item.Modules = uniqueStringsPreserve(nonEmptyStrings(item.Modules...))
	item.Capabilities = uniqueStringsPreserve(nonEmptyStrings(item.Capabilities...))
	item.MatchEvidence = uniqueStringsPreserve(nonEmptyStrings(item.MatchEvidence...))
	item.Warnings = uniqueStringsPreserve(nonEmptyStrings(item.Warnings...))
	item.Sources = mergeLibrarySources(nil, item.Sources)
	if item.Name == "" {
		item.Name = humanizeMinecraftFilename(item.Filename)
	}
	for _, source := range item.Sources {
		if source.Exact {
			item.ProvenanceConfidence = maxFloat(item.ProvenanceConfidence, source.Confidence)
		}
		if item.RemoteArtURL == "" && source.IconURL != "" && source.Confidence >= .86 {
			item.RemoteArtURL = source.IconURL
		}
		if item.LatestVersion == "" && source.LatestVersion != "" {
			item.LatestVersion = source.LatestVersion
		}
	}
	if item.UpdateStatus == "" {
		item.UpdateStatus = "unknown"
	}
	if item.UpdateStatus == "unknown" && len(item.Sources) > 0 {
		item.UpdateStatus = "review"
		item.UpdateMessage = "Provider metadata is available, but the local artifact lacks an exact update identity for automatic replacement."
	}
	if item.LocalArtURL != "" {
		item.ArtOrigin = "embedded-local"
	} else if item.RemoteArtURL != "" {
		item.ArtOrigin = "provider"
	}
}

var _ = time.Now
