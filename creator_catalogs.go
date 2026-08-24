package main

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const creatorCatalogMaxBytes int64 = 64 << 20

//go:embed catalogs/creators/*.json
var embeddedCreatorCatalogs embed.FS

type CreatorCatalogCoverage struct {
	ExpectedVideos int    `json:"expectedVideos,omitempty"`
	Complete       bool   `json:"complete"`
	Notes          string `json:"notes,omitempty"`
}

type CreatorCatalog struct {
	SchemaVersion int                    `json:"schemaVersion"`
	ID            string                 `json:"id"`
	Title         string                 `json:"title,omitempty"`
	UpdatedAt     string                 `json:"updatedAt,omitempty"`
	Source        string                 `json:"source,omitempty"`
	SourceURL     string                 `json:"sourceUrl,omitempty"`
	Creator       CreatorChannel         `json:"creator"`
	Coverage      CreatorCatalogCoverage `json:"coverage"`
	Videos        []CreatorVideo         `json:"videos,omitempty"`
}

type creatorCatalogCandidate struct {
	Catalog CreatorCatalog
	Path    string
	Local   bool
	Bytes   []byte
}

type CreatorCatalogState struct {
	Catalogs         int      `json:"catalogs"`
	Videos           int      `json:"videos"`
	Recommendations  int      `json:"recommendations"`
	ExpectedVideos   int      `json:"expectedVideos"`
	CompleteCatalogs int      `json:"completeCatalogs"`
	Directory        string   `json:"directory"`
	Digest           string   `json:"digest,omitempty"`
	LastReloadedAt   string   `json:"lastReloadedAt,omitempty"`
	Errors           []string `json:"errors,omitempty"`
}

var creatorCatalogRuntime struct {
	sync.Mutex
	digest string
	state  CreatorCatalogState
}

func (a *App) creatorCatalogDir() string {
	return filepath.Join(a.cfgDir, "creator-catalogs")
}

func decodeCreatorCatalog(raw []byte, sourcePath string) (CreatorCatalog, error) {
	var c CreatorCatalog
	if err := json.Unmarshal(raw, &c); err != nil {
		return c, fmt.Errorf("%s: %w", sourcePath, err)
	}
	c.ID = strings.TrimSpace(c.ID)
	if c.SchemaVersion != 1 {
		return c, fmt.Errorf("%s: unsupported schemaVersion %d", sourcePath, c.SchemaVersion)
	}
	if c.ID == "" {
		return c, fmt.Errorf("%s: catalog id is required", sourcePath)
	}
	platform := creatorPlatform(c.Creator.Platform, c.Creator.URL)
	if platform != "youtube" && platform != "tiktok" {
		return c, fmt.Errorf("%s: creator platform must be youtube or tiktok", sourcePath)
	}
	c.Creator.Platform = platform
	if strings.TrimSpace(c.Creator.URL) == "" && strings.TrimSpace(c.Creator.Handle) == "" && strings.TrimSpace(c.Creator.ChannelID) == "" {
		return c, fmt.Errorf("%s: creator url, handle, or channelId is required", sourcePath)
	}
	if len(c.Videos) > 10000 {
		return c, fmt.Errorf("%s: video count %d exceeds safety limit", sourcePath, len(c.Videos))
	}
	for i := range c.Videos {
		if strings.TrimSpace(c.Videos[i].ID) == "" {
			return c, fmt.Errorf("%s: videos[%d].id is required", sourcePath, i)
		}
		if len(c.Videos[i].Mods) > 2000 {
			return c, fmt.Errorf("%s: videos[%d] recommendation count exceeds safety limit", sourcePath, i)
		}
	}
	return c, nil
}

func readEmbeddedCreatorCatalogs() ([]creatorCatalogCandidate, []string) {
	entries, err := fs.ReadDir(embeddedCreatorCatalogs, "catalogs/creators")
	if err != nil {
		return nil, []string{err.Error()}
	}
	out := make([]creatorCatalogCandidate, 0, len(entries))
	errs := []string{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".json") || strings.HasSuffix(strings.ToLower(entry.Name()), ".disabled.json") {
			continue
		}
		path := "catalogs/creators/" + entry.Name()
		raw, err := embeddedCreatorCatalogs.ReadFile(path)
		if err != nil {
			errs = append(errs, path+": "+err.Error())
			continue
		}
		cat, err := decodeCreatorCatalog(raw, "embedded:"+path)
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}
		out = append(out, creatorCatalogCandidate{Catalog: cat, Path: "embedded:" + path, Bytes: raw})
	}
	return out, errs
}

func readLocalCreatorCatalogs(root string) ([]creatorCatalogCandidate, []string) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, []string{err.Error()}
	}
	out := []creatorCatalogCandidate{}
	errs := []string{}
	_ = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			errs = append(errs, path+": "+walkErr.Error())
			return nil
		}
		if path == root {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		name := strings.ToLower(entry.Name())
		if !strings.HasSuffix(name, ".json") || strings.HasSuffix(name, ".disabled.json") {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			errs = append(errs, path+": "+err.Error())
			return nil
		}
		if info.Size() > creatorCatalogMaxBytes {
			errs = append(errs, fmt.Sprintf("%s: %d bytes exceeds 64 MiB catalog limit", path, info.Size()))
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			errs = append(errs, path+": "+err.Error())
			return nil
		}
		cat, err := decodeCreatorCatalog(raw, path)
		if err != nil {
			errs = append(errs, err.Error())
			return nil
		}
		out = append(out, creatorCatalogCandidate{Catalog: cat, Path: path, Local: true, Bytes: raw})
		return nil
	})
	return out, errs
}

func catalogUpdatedAt(c CreatorCatalog) time.Time {
	t, _ := time.Parse(time.RFC3339, strings.TrimSpace(c.UpdatedAt))
	return t
}

func selectCreatorCatalogs(candidates []creatorCatalogCandidate) []creatorCatalogCandidate {
	best := map[string]creatorCatalogCandidate{}
	for _, cand := range candidates {
		id := strings.ToLower(cand.Catalog.ID)
		old, ok := best[id]
		if !ok {
			best[id] = cand
			continue
		}
		ot, nt := catalogUpdatedAt(old.Catalog), catalogUpdatedAt(cand.Catalog)
		if nt.After(ot) || (nt.Equal(ot) && cand.Local && !old.Local) {
			best[id] = cand
		}
	}
	out := make([]creatorCatalogCandidate, 0, len(best))
	for _, cand := range best {
		out = append(out, cand)
	}
	sort.Slice(out, func(i, j int) bool { return strings.ToLower(out[i].Catalog.ID) < strings.ToLower(out[j].Catalog.ID) })
	return out
}

func creatorCatalogDigest(candidates []creatorCatalogCandidate) string {
	h := sha256.New()
	for _, cand := range candidates {
		_, _ = h.Write([]byte(strings.ToLower(cand.Catalog.ID)))
		_, _ = h.Write([]byte{0})
		_, _ = h.Write(cand.Bytes)
		_, _ = h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}

func catalogKind(raw string) string {
	s := strings.ToLower(strings.TrimSpace(raw))
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, "_", "")
	switch s {
	case "modpack":
		return "modpack"
	case "resourcepack", "texturepack":
		return "resourcepack"
	case "shader", "shaderpack":
		return "shader"
	case "datapack":
		return "datapack"
	default:
		return "mod"
	}
}

func catalogSourceKind(id string) string { return "catalog:" + strings.ToLower(strings.TrimSpace(id)) }

func mergeStringSliceStable(a, b []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(a)+len(b))
	for _, list := range [][]string{a, b} {
		for _, value := range list {
			value = strings.TrimSpace(value)
			if value == "" || seen[strings.ToLower(value)] {
				continue
			}
			seen[strings.ToLower(value)] = true
			out = append(out, value)
		}
	}
	return out
}

func sameCreatorMod(a, b CreatorMod) bool {
	if a.ProjectID != "" && b.ProjectID != "" && strings.EqualFold(a.Provider, b.Provider) && strings.EqualFold(a.ProjectID, b.ProjectID) {
		return true
	}
	if a.URL != "" && b.URL != "" && strings.EqualFold(strings.TrimSpace(a.URL), strings.TrimSpace(b.URL)) {
		return true
	}
	return a.Name != "" && b.Name != "" && strings.EqualFold(strings.TrimSpace(a.Name), strings.TrimSpace(b.Name))
}

func mergeCatalogMod(dst CreatorMod, src CreatorMod, catalogID, videoURL string) CreatorMod {
	if dst.Name == "" {
		dst.Name = src.Name
	}
	if dst.Provider == "" {
		dst.Provider = src.Provider
	}
	if dst.ProjectID == "" {
		dst.ProjectID = src.ProjectID
	}
	if dst.ProjectType == "" {
		dst.ProjectType = catalogKind(src.ProjectType)
	}
	if dst.Author == "" {
		dst.Author = src.Author
	}
	if dst.IconURL == "" {
		dst.IconURL = src.IconURL
	}
	if dst.URL == "" {
		dst.URL = src.URL
	}
	if dst.Timestamp == "" {
		dst.Timestamp = src.Timestamp
	}
	if dst.TimestampS == 0 {
		dst.TimestampS = src.TimestampS
	}
	if dst.VideoLink == "" {
		dst.VideoLink = firstNonEmpty(src.VideoLink, videoURL)
	}
	if dst.Evidence == "" {
		dst.Evidence = src.Evidence
	}
	if dst.DescriptionContext == "" {
		dst.DescriptionContext = src.DescriptionContext
	}
	if dst.TranscriptContext == "" {
		dst.TranscriptContext = src.TranscriptContext
	}
	if dst.ProjectSummary == "" {
		dst.ProjectSummary = src.ProjectSummary
	}
	if dst.Confidence == 0 {
		dst.Confidence = src.Confidence
	}
	dst.SourceKinds = mergeStringSliceStable(dst.SourceKinds, append(src.SourceKinds, "catalog", catalogSourceKind(catalogID)))
	return dst
}

func mergeCatalogMods(existing, incoming []CreatorMod, catalogID, videoURL string) []CreatorMod {
	out := append([]CreatorMod(nil), existing...)
	for _, src := range incoming {
		src.ProjectType = catalogKind(src.ProjectType)
		found := -1
		for i := range out {
			if sameCreatorMod(out[i], src) {
				found = i
				break
			}
		}
		if found >= 0 {
			out[found] = mergeCatalogMod(out[found], src, catalogID, videoURL)
		} else {
			out = append(out, mergeCatalogMod(CreatorMod{}, src, catalogID, videoURL))
		}
	}
	return out
}

func (a *App) applyCreatorCatalog(c CreatorCatalog) (videos, recommendations int) {
	seed := c.Creator
	if normalized, err := normalizeCreatorChannelInput(firstNonEmpty(seed.URL, seed.Handle, seed.ChannelID)); err == nil {
		if seed.Platform == "" {
			seed.Platform = normalized.Platform
		}
		if seed.Handle == "" {
			seed.Handle = normalized.Handle
		}
		if seed.ChannelID == "" {
			seed.ChannelID = normalized.ChannelID
		}
		if seed.URL == "" {
			seed.URL = normalized.URL
		}
	}
	seed.Source = firstNonEmpty(seed.Source, "catalog:"+c.ID)
	if c.Coverage.ExpectedVideos > seed.TotalVideos {
		seed.TotalVideos = c.Coverage.ExpectedVideos
	}
	if seed.AddedAt == "" {
		seed.AddedAt = firstNonEmpty(c.UpdatedAt, time.Now().UTC().Format(time.RFC3339))
	}

	a.dataMu.Lock()
	foundChannel := -1
	for i := range a.creatorChannels {
		if creatorChannelsEquivalent(a.creatorChannels[i], seed) {
			foundChannel = i
			break
		}
	}
	if foundChannel >= 0 {
		mergeCreatorChannelSeed(&a.creatorChannels[foundChannel], seed)
		if c.Coverage.ExpectedVideos > a.creatorChannels[foundChannel].TotalVideos {
			a.creatorChannels[foundChannel].TotalVideos = c.Coverage.ExpectedVideos
		}
		seed = a.creatorChannels[foundChannel]
	} else {
		a.creatorChannels = append(a.creatorChannels, seed)
	}

	stamp := firstNonEmpty(c.UpdatedAt, time.Now().UTC().Format(time.RFC3339))
	for _, incoming := range c.Videos {
		incoming.Platform = firstNonEmpty(incoming.Platform, seed.Platform)
		incoming.Creator = firstNonEmpty(incoming.Creator, seed.Title, seed.Handle)
		incoming.CreatorURL = firstNonEmpty(incoming.CreatorURL, seed.URL)
		incoming.ChannelID = firstNonEmpty(incoming.ChannelID, seed.ChannelID)
		incoming.ChannelHandle = firstNonEmpty(incoming.ChannelHandle, seed.Handle)
		incoming.URL = firstNonEmpty(incoming.URL, creatorVideoURL(incoming.Platform, incoming.ID))
		incoming.DiscoveredAt = firstNonEmpty(incoming.DiscoveredAt, stamp)
		if len(incoming.Mods) > 0 && incoming.AnalyzedAt == "" {
			incoming.AnalyzedAt = stamp
		}
		for i := range incoming.Mods {
			incoming.Mods[i].VideoLink = firstNonEmpty(incoming.Mods[i].VideoLink, incoming.URL)
		}
		idx := -1
		for i := range a.creatorVideos {
			if strings.EqualFold(a.creatorVideos[i].Platform, incoming.Platform) && a.creatorVideos[i].ID == incoming.ID {
				idx = i
				break
			}
		}
		if idx < 0 {
			incoming.Mods = mergeCatalogMods(nil, incoming.Mods, c.ID, incoming.URL)
			a.creatorVideos = append(a.creatorVideos, incoming)
		} else {
			old := &a.creatorVideos[idx]
			if old.Title == "" {
				old.Title = incoming.Title
			}
			if old.URL == "" {
				old.URL = incoming.URL
			}
			if old.Creator == "" {
				old.Creator = incoming.Creator
			}
			if old.CreatorURL == "" {
				old.CreatorURL = incoming.CreatorURL
			}
			if old.ChannelID == "" {
				old.ChannelID = incoming.ChannelID
			}
			if old.ChannelHandle == "" {
				old.ChannelHandle = incoming.ChannelHandle
			}
			if old.ThumbnailURL == "" {
				old.ThumbnailURL = incoming.ThumbnailURL
			}
			if old.Description == "" {
				old.Description = incoming.Description
			}
			if old.PublishedAt == "" {
				old.PublishedAt = incoming.PublishedAt
			}
			if old.VideoKind == "" {
				old.VideoKind = incoming.VideoKind
			}
			if old.DurationSeconds == 0 {
				old.DurationSeconds = incoming.DurationSeconds
			}
			if old.DiscoveredAt == "" {
				old.DiscoveredAt = incoming.DiscoveredAt
			}
			old.Mods = mergeCatalogMods(old.Mods, incoming.Mods, c.ID, firstNonEmpty(old.URL, incoming.URL))
			if old.AnalyzedAt == "" && len(old.Mods) > 0 {
				old.AnalyzedAt = incoming.AnalyzedAt
			}
		}
	}
	a.refreshCreatorChannelStatsLocked()
	for _, v := range a.creatorVideos {
		if creatorVideoBelongsToChannel(v, seed) || creatorVideoChannelKey(v) == creatorChannelKey(seed) {
			videos++
			recommendations += len(v.Mods)
		}
	}
	a.dataMu.Unlock()
	return videos, recommendations
}

func creatorVideoURL(platform, id string) string {
	if strings.EqualFold(platform, "tiktok") {
		return "https://www.tiktok.com/video/" + id
	}
	return "https://www.youtube.com/watch?v=" + id
}

func (a *App) reloadCreatorCatalogs(force bool) (CreatorCatalogState, error) {
	creatorCatalogRuntime.Lock()
	defer creatorCatalogRuntime.Unlock()
	embedded, errs := readEmbeddedCreatorCatalogs()
	local, localErrs := readLocalCreatorCatalogs(a.creatorCatalogDir())
	errs = append(errs, localErrs...)
	selected := selectCreatorCatalogs(append(embedded, local...))
	digest := creatorCatalogDigest(selected)
	if !force && digest != "" && digest == creatorCatalogRuntime.digest {
		state := creatorCatalogRuntime.state
		state.Errors = append([]string(nil), errs...)
		creatorCatalogRuntime.state = state
		return state, nil
	}
	state := CreatorCatalogState{Directory: a.creatorCatalogDir(), Digest: digest, LastReloadedAt: time.Now().UTC().Format(time.RFC3339), Errors: errs}
	for _, cand := range selected {
		state.Catalogs++
		state.Videos += len(cand.Catalog.Videos)
		for _, v := range cand.Catalog.Videos {
			state.Recommendations += len(v.Mods)
		}
		if cand.Catalog.Coverage.ExpectedVideos > state.ExpectedVideos {
			state.ExpectedVideos = cand.Catalog.Coverage.ExpectedVideos
		}
		if cand.Catalog.Coverage.Complete {
			state.CompleteCatalogs++
		}
		a.applyCreatorCatalog(cand.Catalog)
	}
	if err := a.saveCreatorVideos(); err != nil {
		errs = append(errs, err.Error())
	}
	if err := a.saveCreatorChannels(); err != nil {
		errs = append(errs, err.Error())
	}
	state.Errors = errs
	creatorCatalogRuntime.digest = digest
	creatorCatalogRuntime.state = state
	if len(errs) > 0 {
		return state, fmt.Errorf("creator catalog reload completed with %d warning(s)", len(errs))
	}
	return state, nil
}

func currentCreatorCatalogState(a *App) CreatorCatalogState {
	creatorCatalogRuntime.Lock()
	state := creatorCatalogRuntime.state
	creatorCatalogRuntime.Unlock()
	if state.Directory == "" {
		state.Directory = a.creatorCatalogDir()
	}
	return state
}

func (a *App) handleCreatorCatalogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, currentCreatorCatalogState(a))
}

func (a *App) handleCreatorCatalogReload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	state, err := a.reloadCreatorCatalogs(true)
	status := http.StatusOK
	if err != nil {
		status = http.StatusMultiStatus
	}
	writeJSON(w, status, state)
}
