package main

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha512"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	appName    = "Minecraft Mod Vault"
	appVersion = "0.11.0"
	userAgent  = "MinecraftModVault/0.11.0 (desktop app; universal Minecraft content browser/updater/converter)"
)

//go:embed web/* assets/* knowledge/* repair-brain/*
var embeddedFiles embed.FS

type Settings struct {
	JavaRoot                  string            `json:"javaRoot"`
	BedrockRoot               string            `json:"bedrockRoot"`
	BedrockPreviewRoot        string            `json:"bedrockPreviewRoot"`
	BedrockCustomRoots        []string          `json:"bedrockCustomRoots"`
	WorldRoot                 string            `json:"worldRoot"`
	ServerRoot                string            `json:"serverRoot"`
	ServerPlatform            string            `json:"serverPlatform"`
	GameVersion               string            `json:"gameVersion"`
	Loader                    string            `json:"loader"`
	ActiveProfile             string            `json:"activeProfile"`
	CurseForgeAPIKey          string            `json:"curseForgeApiKey"`
	GitHubToken               string            `json:"githubToken"`
	YouTubeAPIKey             string            `json:"youtubeApiKey"`
	BuiltByBitAPIKey          string            `json:"builtByBitApiKey"`
	BuiltByBitOAuthToken      string            `json:"builtByBitOAuthToken"`
	NexusAPIKey               string            `json:"nexusApiKey"`
	AutoRefreshMinutes        int               `json:"autoRefreshMinutes"`
	CreatorRefreshMinutes     int               `json:"creatorRefreshMinutes"`
	CreatorTranscriptModel    string            `json:"creatorTranscriptModel"`
	CreatorArchiveConcurrency int               `json:"creatorArchiveConcurrency"`
	PreferredTags             []string          `json:"preferredTags"`
	EnabledSources            []string          `json:"enabledSources"`
	ProviderSchemaVersion     int               `json:"providerSchemaVersion"`
	ConversionToolPaths       map[string]string `json:"conversionToolPaths,omitempty"`
}

type App struct {
	cfgDir                 string
	stateFile              string
	settings               Settings
	mu                     sync.RWMutex
	token                  string
	lastHeartbeat          time.Time
	seenHeartbeat          bool
	httpClient             *http.Client
	dataMu                 sync.RWMutex
	avatarCacheMu          sync.RWMutex
	avatarCache            map[string]string
	providerHealthMu       sync.RWMutex
	providerHealth         map[string]ProviderHealth
	providerCacheMu        sync.RWMutex
	providerCache          map[string]providerCacheEntry
	recommendations        []UnifiedProject
	recommendationsUpdated time.Time
	creatorVideos          []CreatorVideo
	creatorChannels        []CreatorChannel
	creatorSyncMu          sync.Mutex
	creatorSyncRunning     map[string]bool
	updatePlans            map[string]UpdatePlan
	portingPlans           map[string]PortingPlan
	doctorRepairPlans      map[string]DoctorRepairPlan
	brain                  *CompatibilityBrain
	brainInitError         string
	portingMu              sync.RWMutex
	portingRuns            map[string]*PortingBuildRun
	portingCancels         map[string]context.CancelFunc
}

type APIError struct {
	Error string `json:"error"`
}

type ManagerEntry struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Kind     string `json:"kind"`
	Size     int64  `json:"size"`
	Modified string `json:"modified"`
	Enabled  bool   `json:"enabled"`
	IsDir    bool   `json:"isDir"`
}

type ModrinthSearchResponse struct {
	Hits      []ModrinthHit `json:"hits"`
	Offset    int           `json:"offset"`
	Limit     int           `json:"limit"`
	TotalHits int           `json:"total_hits"`
}

type ModrinthHit struct {
	ProjectID         string   `json:"project_id"`
	ProjectType       string   `json:"project_type"`
	Slug              string   `json:"slug"`
	Author            string   `json:"author"`
	Title             string   `json:"title"`
	Description       string   `json:"description"`
	Categories        []string `json:"categories"`
	DisplayCategories []string `json:"display_categories"`
	Versions          []string `json:"versions"`
	Downloads         int64    `json:"downloads"`
	IconURL           string   `json:"icon_url"`
	DateModified      string   `json:"date_modified"`
}

type ModrinthProject struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	ProjectType string `json:"project_type"`
	Description string `json:"description"`
}

type ModrinthVersion struct {
	ID            string                `json:"id"`
	ProjectID     string                `json:"project_id"`
	Name          string                `json:"name"`
	VersionNumber string                `json:"version_number"`
	DatePublished string                `json:"date_published"`
	GameVersions  []string              `json:"game_versions"`
	Loaders       []string              `json:"loaders"`
	Files         []ModrinthVersionFile `json:"files"`
	Dependencies  []ModrinthDependency  `json:"dependencies"`
}

type ModrinthDependency struct {
	VersionID      string `json:"version_id"`
	ProjectID      string `json:"project_id"`
	FileName       string `json:"file_name"`
	DependencyType string `json:"dependency_type"`
}

type InstalledProject struct {
	Project     string `json:"project"`
	ProjectID   string `json:"projectId"`
	ProjectType string `json:"projectType"`
	Version     string `json:"version"`
	File        string `json:"file"`
	Path        string `json:"path"`
	Target      string `json:"target"`
	Dependency  bool   `json:"dependency"`
}

type ModrinthVersionFile struct {
	URL      string            `json:"url"`
	Filename string            `json:"filename"`
	Primary  bool              `json:"primary"`
	Size     int64             `json:"size"`
	Hashes   map[string]string `json:"hashes"`
}

type ModrinthInstallRequest struct {
	ProjectID   string `json:"projectId"`
	GameVersion string `json:"gameVersion"`
	Loader      string `json:"loader"`
	Target      string `json:"target"`
}

type ToggleRequest struct {
	Path string `json:"path"`
}

type OpenRequest struct {
	URL  string `json:"url"`
	Path string `json:"path"`
}

func main() {
	cfgDir, err := configDir()
	if err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		log.Fatal(err)
	}

	a := &App{
		cfgDir:         cfgDir,
		stateFile:      filepath.Join(cfgDir, "settings.json"),
		token:          randomToken(24),
		avatarCache:    map[string]string{},
		providerHealth: map[string]ProviderHealth{},
		providerCache:  map[string]providerCacheEntry{},
		portingRuns:    map[string]*PortingBuildRun{},
		portingCancels: map[string]context.CancelFunc{},
		httpClient: &http.Client{
			Timeout: 45 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 8 {
					return errors.New("too many redirects")
				}
				return nil
			},
		},
	}
	a.loadSettings()
	a.loadPersistentCaches()
	brain, brainErr := openCompatibilityBrain(cfgDir)
	if brainErr != nil {
		a.brainInitError = brainErr.Error()
		log.Printf("compatibility brain: %v", brainErr)
	} else {
		a.brain = brain
		defer brain.Close()
	}

	mux := http.NewServeMux()
	a.registerRoutes(mux)

	addr := "127.0.0.1:0"
	if p := strings.TrimSpace(os.Getenv("MMV_PORT")); p != "" {
		addr = "127.0.0.1:" + p
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	appURL := fmt.Sprintf("http://127.0.0.1:%d/?token=%s", port, url.QueryEscape(a.token))

	server := &http.Server{
		Handler:           withSecurityHeaders(mux),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       90 * time.Second,
	}

	go func() {
		if err := server.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("server: %v", err)
		}
	}()

	go a.backgroundIntelligenceLoop()
	go a.backgroundCreatorArchiveLoop()

	if os.Getenv("MMV_NO_LAUNCH") != "1" {
		if err := launchAppWindow(appURL); err != nil {
			_ = openExternal(appURL)
		}
	} else {
		fmt.Println(appURL)
	}

	// Exit after the UI has disappeared. The browser sends a heartbeat while the app is open.
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		a.mu.RLock()
		seen := a.seenHeartbeat
		last := a.lastHeartbeat
		a.mu.RUnlock()
		if seen && time.Since(last) > 75*time.Second {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			_ = server.Shutdown(ctx)
			cancel()
			return
		}
	}
}

func configDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "MinecraftModVault"), nil
}

func randomToken(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

func defaultJavaRoot() string {
	home, _ := os.UserHomeDir()
	switch runtime.GOOS {
	case "windows":
		if appdata := os.Getenv("APPDATA"); appdata != "" {
			return filepath.Join(appdata, ".minecraft")
		}
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "minecraft")
	}
	return filepath.Join(home, ".minecraft")
}

func defaultBedrockRoot() string {
	if runtime.GOOS == "windows" {
		if local := os.Getenv("LOCALAPPDATA"); local != "" {
			return filepath.Join(local, "Packages", "Microsoft.MinecraftUWP_8wekyb3d8bbwe", "LocalState", "games", "com.mojang")
		}
	}
	return ""
}

func defaultBedrockPreviewRoot() string {
	if runtime.GOOS == "windows" {
		if local := os.Getenv("LOCALAPPDATA"); local != "" {
			return filepath.Join(local, "Packages", "Microsoft.MinecraftWindowsBeta_8wekyb3d8bbwe", "LocalState", "games", "com.mojang")
		}
	}
	return ""
}

func cleanRootList(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		value = filepath.Clean(value)
		key := strings.ToLower(value)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, value)
	}
	return out
}

func (a *App) loadSettings() {
	a.settings = Settings{
		JavaRoot:                  defaultJavaRoot(),
		BedrockRoot:               defaultBedrockRoot(),
		BedrockPreviewRoot:        defaultBedrockPreviewRoot(),
		GameVersion:               "1.21.1",
		Loader:                    "fabric",
		ServerPlatform:            "PAPER",
		ActiveProfile:             "Default",
		AutoRefreshMinutes:        15,
		CreatorRefreshMinutes:     60,
		CreatorTranscriptModel:    defaultWhisperModelName,
		CreatorArchiveConcurrency: 2,
		PreferredTags:             []string{"furniture", "cute", "cozy", "pets", "railroads", "vehicles", "technology", "magic", "particles", "water", "foliage", "cards", "creature collecting"},
		EnabledSources:            allProviderIDs(true),
		ProviderSchemaVersion:     providerSchemaVersion,
	}
	b, err := os.ReadFile(a.stateFile)
	if err == nil {
		_ = json.Unmarshal(b, &a.settings)
	}
	if a.settings.JavaRoot == "" {
		a.settings.JavaRoot = defaultJavaRoot()
	}
	if a.settings.BedrockPreviewRoot == "" {
		a.settings.BedrockPreviewRoot = defaultBedrockPreviewRoot()
	}
	a.settings.BedrockCustomRoots = cleanRootList(a.settings.BedrockCustomRoots)
	if a.settings.ConversionToolPaths == nil {
		a.settings.ConversionToolPaths = map[string]string{}
	}
	if a.settings.GameVersion == "" {
		a.settings.GameVersion = "1.21.1"
	}
	if a.settings.Loader == "" {
		a.settings.Loader = "fabric"
	}
	if a.settings.ServerPlatform == "" {
		a.settings.ServerPlatform = "PAPER"
	}
	if a.settings.AutoRefreshMinutes <= 0 {
		a.settings.AutoRefreshMinutes = 15
	}
	if a.settings.CreatorRefreshMinutes <= 0 {
		a.settings.CreatorRefreshMinutes = 60
	}
	if strings.TrimSpace(a.settings.CreatorTranscriptModel) == "" {
		a.settings.CreatorTranscriptModel = defaultWhisperModelName
	}
	if a.settings.CreatorArchiveConcurrency <= 0 {
		a.settings.CreatorArchiveConcurrency = 1
	}
	if len(a.settings.PreferredTags) == 0 {
		a.settings.PreferredTags = []string{"furniture", "cute", "cozy", "pets", "railroads", "vehicles", "technology", "magic", "particles", "water", "foliage", "cards", "creature collecting"}
	}
	if len(a.settings.EnabledSources) == 0 {
		a.settings.EnabledSources = allProviderIDs(true)
	}
	// v0.7 expands the source registry beyond the six v0.6 lanes. Migrate an
	// existing profile once so newly integrated public providers are not hidden
	// just because an older settings file could not name them yet.
	if a.settings.ProviderSchemaVersion < providerSchemaVersion {
		a.settings.EnabledSources = uniqueStringsPreserve(append(a.settings.EnabledSources, allProviderIDs(true)...))
		a.settings.ProviderSchemaVersion = providerSchemaVersion
	}
}

func (a *App) saveSettings() error {
	a.mu.RLock()
	b, err := json.MarshalIndent(a.settings, "", "  ")
	a.mu.RUnlock()
	if err != nil {
		return err
	}
	return os.WriteFile(a.stateFile, b, 0o644)
}

func (a *App) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/heartbeat", a.auth(a.handleHeartbeat))
	mux.HandleFunc("/api/system", a.auth(a.handleSystem))
	mux.HandleFunc("/api/settings", a.auth(a.handleSettings))
	mux.HandleFunc("/api/open", a.auth(a.handleOpen))
	mux.HandleFunc("/api/manager", a.auth(a.handleManager))
	mux.HandleFunc("/api/library", a.auth(a.handleLibrary))
	mux.HandleFunc("/api/library/art", a.auth(a.handleLibraryArt))
	mux.HandleFunc("/api/library/action", a.auth(a.handleLibraryAction))
	mux.HandleFunc("/api/library/history", a.auth(a.handleLibraryHistory))
	mux.HandleFunc("/api/library/undo", a.auth(a.handleLibraryUndo))
	mux.HandleFunc("/api/library/bedrock/install", a.auth(a.handleBedrockInstall))
	mux.HandleFunc("/api/library/bedrock/activate", a.auth(a.handleBedrockActivate))
	mux.HandleFunc("/api/file/toggle", a.auth(a.handleToggle))
	mux.HandleFunc("/api/file/trash", a.auth(a.handleTrash))
	mux.HandleFunc("/api/import", a.auth(a.handleImport))
	mux.HandleFunc("/api/modrinth/search", a.auth(a.handleModrinthSearch))
	mux.HandleFunc("/api/modrinth/install", a.auth(a.handleModrinthInstall))
	mux.HandleFunc("/api/providers", a.auth(a.handleProviders))
	mux.HandleFunc("/api/providers/search", a.auth(a.handleProviderSearch))
	mux.HandleFunc("/api/providers/detail", a.auth(a.handleProviderDetail))
	mux.HandleFunc("/api/providers/install", a.auth(a.handleProviderInstall))
	mux.HandleFunc("/api/taxonomy", a.auth(a.handleTaxonomy))
	mux.HandleFunc("/api/recommendations", a.auth(a.handleRecommendations))
	mux.HandleFunc("/api/recommendations/refresh", a.auth(a.handleRecommendationRefresh))
	mux.HandleFunc("/api/updater/scan", a.auth(a.handleUpdaterScan))
	mux.HandleFunc("/api/updater/apply", a.auth(a.handleUpdaterApply))
	mux.HandleFunc("/api/doctor/sources", a.auth(a.handleDoctorSources))
	mux.HandleFunc("/api/doctor/scan", a.auth(a.handleDoctorScan))
	mux.HandleFunc("/api/doctor/tools", a.auth(a.handleDoctorTools))
	mux.HandleFunc("/api/doctor/analyze", a.auth(a.handleDoctorAnalyze))
	mux.HandleFunc("/api/doctor/log", a.auth(a.handleDoctorLog))
	mux.HandleFunc("/api/doctor/repair/plan", a.auth(a.handleDoctorRepairPlan))
	mux.HandleFunc("/api/doctor/repair/apply", a.auth(a.handleDoctorRepairApply))
	mux.HandleFunc("/api/doctor/repair/restore", a.auth(a.handleDoctorRepairRestore))
	mux.HandleFunc("/api/brain/status", a.auth(a.handleBrainStatus))
	mux.HandleFunc("/api/brain/search", a.auth(a.handleBrainSearch))
	mux.HandleFunc("/api/brain/version", a.auth(a.handleBrainVersion))
	mux.HandleFunc("/api/atlas/summary", a.auth(a.handleAtlasSummary))
	mux.HandleFunc("/api/atlas/versions", a.auth(a.handleAtlasVersions))
	mux.HandleFunc("/api/atlas/toolchains", a.auth(a.handleAtlasToolchains))
	mux.HandleFunc("/api/atlas/resolve", a.auth(a.handleAtlasResolve))
	mux.HandleFunc("/api/repair-lab/status", a.auth(a.handleRepairLabStatus))
	mux.HandleFunc("/api/repair-lab/import", a.auth(a.handleRepairLabImport))
	mux.HandleFunc("/api/repair-lab/session", a.auth(a.handleRepairLabSession))
	mux.HandleFunc("/api/repair-lab/prepare", a.auth(a.handleRepairLabPrepare))
	mux.HandleFunc("/api/repair-lab/run", a.auth(a.handleRepairLabRun))
	mux.HandleFunc("/api/repair-lab/cancel", a.auth(a.handleRepairLabCancel))
	mux.HandleFunc("/api/repair-lab/reset", a.auth(a.handleRepairLabReset))
	mux.HandleFunc("/api/repair-lab/export", a.auth(a.handleRepairLabExport))
	mux.HandleFunc("/api/repair-lab/log", a.auth(a.handleRepairLabLog))
	mux.HandleFunc("/api/repair-lab/download", a.auth(a.handleRepairLabDownload))
	mux.HandleFunc("/api/conversion/status", a.auth(a.handleConversionStatus))
	mux.HandleFunc("/api/conversion/import", a.auth(a.handleConversionImport))
	mux.HandleFunc("/api/conversion/import-path", a.auth(a.handleConversionImportPath))
	mux.HandleFunc("/api/conversion/session", a.auth(a.handleConversionSession))
	mux.HandleFunc("/api/conversion/plan", a.auth(a.handleConversionPlan))
	mux.HandleFunc("/api/conversion/run", a.auth(a.handleConversionRun))
	mux.HandleFunc("/api/conversion/reset", a.auth(a.handleConversionReset))
	mux.HandleFunc("/api/conversion/download", a.auth(a.handleConversionDownload))
	mux.HandleFunc("/api/conversion/tool", a.auth(a.handleConversionToolConfig))
	mux.HandleFunc("/api/conversion/adapter/run", a.auth(a.handleConversionAdapterRun))
	mux.HandleFunc("/api/porting/atlas", a.auth(a.handlePortingAtlas))
	mux.HandleFunc("/api/porting/plan", a.auth(a.handlePortingPlan))
	mux.HandleFunc("/api/porting/environment", a.auth(a.handlePortingEnvironment))
	mux.HandleFunc("/api/porting/workspaces", a.auth(a.handlePortingWorkspaces))
	mux.HandleFunc("/api/porting/workspace", a.auth(a.handlePortingWorkspace))
	mux.HandleFunc("/api/creators/discover", a.auth(a.handleCreatorDiscover))
	mux.HandleFunc("/api/creators/analyze", a.auth(a.handleCreatorAnalyze))
	mux.HandleFunc("/api/creators/videos", a.auth(a.handleCreatorVideos))
	mux.HandleFunc("/api/creators/channels", a.auth(a.handleCreatorChannels))
	mux.HandleFunc("/api/creators/channels/sync", a.auth(a.handleCreatorChannelSync))
	mux.HandleFunc("/api/creators/channels/links/refresh", a.auth(a.handleCreatorProfileLinksRefresh))
	mux.HandleFunc("/api/creators/channels/modpacks/refresh", a.auth(a.handleCreatorModpacksRefresh))
	mux.HandleFunc("/api/creators/catalogs", a.auth(a.handleCreatorCatalogs))
	mux.HandleFunc("/api/creators/catalogs/reload", a.auth(a.handleCreatorCatalogReload))
	mux.HandleFunc("/api/creators/recommendations", a.auth(a.handleCreatorRecommendations))
	mux.HandleFunc("/api/creators/transcript", a.auth(a.handleCreatorTranscript))
	mux.HandleFunc("/api/transcription/status", a.auth(a.handleTranscriptionStatus))
	mux.HandleFunc("/api/transcription/prepare", a.auth(a.handleTranscriptionPrepare))

	sub, _ := fs.Sub(embeddedFiles, "web")
	indexHTML, indexErr := fs.ReadFile(sub, "index.html")
	fileServer := http.FileServer(http.FS(sub))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		if r.URL.Path == "/" {
			if indexErr != nil {
				http.Error(w, "embedded UI is unavailable", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write(indexHTML)
			return
		}
		if ext := filepath.Ext(r.URL.Path); ext != "" {
			if typ := mime.TypeByExtension(ext); typ != "" {
				w.Header().Set("Content-Type", typ)
			}
		}
		fileServer.ServeHTTP(w, r)
	})
}

func withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' https: data:; style-src 'self'; script-src 'self'; connect-src 'self'; font-src 'self'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}

func (a *App) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-MMV-Token")
		if token == "" {
			token = r.URL.Query().Get("token")
		}
		if token != a.token {
			writeJSON(w, http.StatusForbidden, APIError{Error: "invalid local app token"})
			return
		}
		next(w, r)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, 2<<20))
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

func (a *App) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	a.seenHeartbeat = true
	a.lastHeartbeat = time.Now()
	a.mu.Unlock()
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (a *App) handleSystem(w http.ResponseWriter, r *http.Request) {
	a.mu.RLock()
	s := a.settings
	a.mu.RUnlock()
	writeJSON(w, http.StatusOK, map[string]any{
		"name":                 appName,
		"version":              appVersion,
		"os":                   runtime.GOOS,
		"arch":                 runtime.GOARCH,
		"configDir":            a.cfgDir,
		"settings":             s,
		"javaExists":           pathExists(s.JavaRoot),
		"bedrockExists":        s.BedrockRoot != "" && pathExists(s.BedrockRoot),
		"bedrockPreviewExists": s.BedrockPreviewRoot != "" && pathExists(s.BedrockPreviewRoot),
		"serverExists":         s.ServerRoot != "" && pathExists(s.ServerRoot),
		"worldExists":          s.WorldRoot != "" && pathExists(s.WorldRoot),
	})
}

func (a *App) handleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.mu.RLock()
		s := a.settings
		a.mu.RUnlock()
		writeJSON(w, http.StatusOK, s)
	case http.MethodPost:
		var s Settings
		if err := decodeJSON(r, &s); err != nil {
			writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
			return
		}
		if s.JavaRoot == "" {
			s.JavaRoot = defaultJavaRoot()
		}
		s.JavaRoot = filepath.Clean(s.JavaRoot)
		if s.BedrockRoot != "" {
			s.BedrockRoot = filepath.Clean(s.BedrockRoot)
		}
		if s.BedrockPreviewRoot != "" {
			s.BedrockPreviewRoot = filepath.Clean(s.BedrockPreviewRoot)
		}
		s.BedrockCustomRoots = cleanRootList(s.BedrockCustomRoots)
		if s.ServerRoot != "" {
			s.ServerRoot = filepath.Clean(s.ServerRoot)
		}
		if s.WorldRoot != "" {
			s.WorldRoot = filepath.Clean(s.WorldRoot)
		}
		if s.ServerPlatform == "" {
			s.ServerPlatform = "PAPER"
		}
		s.ServerPlatform = strings.ToUpper(strings.TrimSpace(s.ServerPlatform))
		if s.GameVersion == "" {
			s.GameVersion = "1.21.1"
		}
		if s.Loader == "" {
			s.Loader = "fabric"
		}
		if s.ActiveProfile == "" {
			s.ActiveProfile = "Default"
		}
		if s.AutoRefreshMinutes <= 0 {
			s.AutoRefreshMinutes = 15
		}
		if s.CreatorRefreshMinutes <= 0 {
			s.CreatorRefreshMinutes = 60
		}
		if strings.TrimSpace(s.CreatorTranscriptModel) == "" {
			s.CreatorTranscriptModel = defaultWhisperModelName
		}
		if s.CreatorArchiveConcurrency <= 0 {
			s.CreatorArchiveConcurrency = 1
		}
		if s.CreatorArchiveConcurrency > 4 {
			s.CreatorArchiveConcurrency = 4
		}
		if len(s.EnabledSources) == 0 {
			s.EnabledSources = allProviderIDs(true)
		}
		if s.ProviderSchemaVersion < providerSchemaVersion {
			s.ProviderSchemaVersion = providerSchemaVersion
		}
		a.mu.Lock()
		a.settings = s
		a.mu.Unlock()
		if err := a.saveSettings(); err != nil {
			writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, s)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (a *App) handleOpen(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req OpenRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	var err error
	if req.URL != "" {
		u, parseErr := url.Parse(req.URL)
		if parseErr != nil || (u.Scheme != "https" && u.Scheme != "http") {
			writeJSON(w, http.StatusBadRequest, APIError{Error: "only http/https URLs are supported"})
			return
		}
		err = openExternal(req.URL)
	} else if req.Path != "" {
		err = openExternal(filepath.Clean(req.Path))
	} else {
		err = errors.New("missing url or path")
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (a *App) javaTargetDir(kind string) string {
	a.mu.RLock()
	root := a.settings.JavaRoot
	a.mu.RUnlock()
	switch kind {
	case "mods":
		return filepath.Join(root, "mods")
	case "resourcepacks":
		return filepath.Join(root, "resourcepacks")
	case "shaderpacks":
		return filepath.Join(root, "shaderpacks")
	case "plugins":
		a.mu.RLock()
		serverRoot := a.settings.ServerRoot
		a.mu.RUnlock()
		if serverRoot != "" {
			return filepath.Join(serverRoot, "plugins")
		}
		return filepath.Join(root, "plugins")
	case "worlds":
		return filepath.Join(root, "saves")
	case "datapacks":
		a.mu.RLock()
		worldRoot := a.settings.WorldRoot
		a.mu.RUnlock()
		if worldRoot != "" {
			return filepath.Join(worldRoot, "datapacks")
		}
		return filepath.Join(a.cfgDir, "downloads")
	default:
		return filepath.Join(a.cfgDir, "downloads")
	}
}

func (a *App) handleManager(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	kinds := []string{"mods", "resourcepacks", "shaderpacks", "plugins"}
	a.mu.RLock()
	worldRoot := a.settings.WorldRoot
	a.mu.RUnlock()
	if worldRoot != "" {
		kinds = append(kinds, "datapacks")
	}
	out := map[string][]ManagerEntry{}
	for _, kind := range kinds {
		entries, err := scanDir(a.javaTargetDir(kind), kind)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
			return
		}
		out[kind] = entries
	}
	out["downloads"], _ = scanDir(a.javaTargetDir("downloads"), "downloads")
	writeJSON(w, http.StatusOK, out)
}

func scanDir(dir, kind string) ([]ManagerEntry, error) {
	list, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	out := make([]ManagerEntry, 0, len(list))
	for _, de := range list {
		info, err := de.Info()
		if err != nil {
			continue
		}
		name := de.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		out = append(out, ManagerEntry{
			Name:     name,
			Path:     filepath.Join(dir, name),
			Kind:     kind,
			Size:     info.Size(),
			Modified: info.ModTime().UTC().Format(time.RFC3339),
			Enabled:  !strings.HasSuffix(strings.ToLower(name), ".disabled"),
			IsDir:    de.IsDir(),
		})
	}
	sort.Slice(out, func(i, j int) bool { return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name) })
	return out, nil
}

func (a *App) handleToggle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req ToggleRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	p := filepath.Clean(req.Path)
	if !a.allowedManagedPath(p) {
		writeJSON(w, http.StatusForbidden, APIError{Error: "path is outside managed Minecraft directories"})
		return
	}
	info, err := os.Stat(p)
	if err != nil {
		writeJSON(w, http.StatusNotFound, APIError{Error: err.Error()})
		return
	}
	if info.IsDir() {
		writeJSON(w, http.StatusBadRequest, APIError{Error: "directory resource packs are not toggled automatically"})
		return
	}
	var next string
	if strings.HasSuffix(strings.ToLower(p), ".disabled") {
		next = strings.TrimSuffix(p, ".disabled")
	} else {
		next = p + ".disabled"
	}
	if pathExists(next) {
		writeJSON(w, http.StatusConflict, APIError{Error: "destination already exists"})
		return
	}
	if err := os.Rename(p, next); err != nil {
		writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "path": next})
}

func (a *App) handleTrash(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req ToggleRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	p := filepath.Clean(req.Path)
	if !a.allowedManagedPath(p) {
		writeJSON(w, http.StatusForbidden, APIError{Error: "path is outside managed Minecraft directories"})
		return
	}
	trash := filepath.Join(a.cfgDir, "trash")
	if err := os.MkdirAll(trash, 0o755); err != nil {
		writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
		return
	}
	dst := filepath.Join(trash, fmt.Sprintf("%d-%s", time.Now().Unix(), filepath.Base(p)))
	if err := os.Rename(p, dst); err != nil {
		if err := copyThenRemove(p, dst); err != nil {
			writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "trashPath": dst})
}

func copyThenRemove(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		if err := copyDir(src, dst); err != nil {
			return err
		}
		return os.RemoveAll(src)
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	if err := os.Chmod(dst, info.Mode().Perm()); err != nil {
		return err
	}
	return os.Remove(src)
}

func copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		from := filepath.Join(src, entry.Name())
		to := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err := copyDir(from, to); err != nil {
				return err
			}
			continue
		}
		in, err := os.Open(from)
		if err != nil {
			return err
		}
		info, infoErr := entry.Info()
		mode := os.FileMode(0o644)
		if infoErr == nil {
			mode = info.Mode().Perm()
		}
		out, err := os.OpenFile(to, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
		if err != nil {
			in.Close()
			return err
		}
		_, copyErr := io.Copy(out, in)
		in.Close()
		closeErr := out.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}

func (a *App) allowedManagedPath(p string) bool {
	roots := []string{a.javaTargetDir("mods"), a.javaTargetDir("resourcepacks"), a.javaTargetDir("shaderpacks"), a.javaTargetDir("plugins"), a.javaTargetDir("datapacks"), a.javaTargetDir("downloads")}
	absP, _ := filepath.Abs(p)
	for _, root := range roots {
		absRoot, _ := filepath.Abs(root)
		rel, err := filepath.Rel(absRoot, absP)
		if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
			return true
		}
	}
	return false
}

func (a *App) handleImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 2<<30)
	if err := r.ParseMultipartForm(64 << 20); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	files := r.MultipartForm.File["files"]
	target := strings.ToLower(strings.TrimSpace(r.FormValue("target")))
	if target == "" {
		target = "auto"
	}
	results := make([]map[string]any, 0, len(files))
	for _, fh := range files {
		src, err := fh.Open()
		if err != nil {
			results = append(results, map[string]any{"name": fh.Filename, "ok": false, "error": err.Error()})
			continue
		}
		chosen := target
		if chosen == "auto" {
			chosen = targetForFilename(fh.Filename)
		}
		if chosen == "bedrock" {
			chosen = "downloads"
		}
		if !validImportTarget(chosen) {
			_ = src.Close()
			results = append(results, map[string]any{"name": fh.Filename, "ok": false, "error": "unsupported import target: " + chosen})
			continue
		}
		dir := a.javaTargetDir(chosen)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			src.Close()
			results = append(results, map[string]any{"name": fh.Filename, "ok": false, "error": err.Error()})
			continue
		}
		name := safeFilename(fh.Filename)
		dst := uniquePath(filepath.Join(dir, name))
		out, err := os.Create(dst)
		if err == nil {
			_, err = io.Copy(out, src)
			_ = out.Close()
		}
		_ = src.Close()
		if err != nil {
			results = append(results, map[string]any{"name": fh.Filename, "ok": false, "error": err.Error()})
			continue
		}
		ext := strings.ToLower(filepath.Ext(name))
		opened := false
		if ext == ".mcpack" || ext == ".mcaddon" || ext == ".mcworld" {
			if openExternal(dst) == nil {
				opened = true
			}
		}
		results = append(results, map[string]any{"name": name, "ok": true, "path": dst, "target": chosen, "opened": opened})
	}
	writeJSON(w, http.StatusOK, map[string]any{"results": results})
}

func validImportTarget(target string) bool {
	switch strings.ToLower(strings.TrimSpace(target)) {
	case "mods", "resourcepacks", "shaderpacks", "plugins", "datapacks", "worlds", "downloads":
		return true
	default:
		return false
	}
}

func targetForFilename(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".jar":
		return "mods"
	case ".mcpack", ".mcaddon", ".mcworld", ".mrpack":
		return "downloads"
	case ".zip":
		return "resourcepacks"
	default:
		return "downloads"
	}
}

func safeFilename(name string) string {
	name = filepath.Base(name)
	name = strings.ReplaceAll(name, "\x00", "")
	if name == "." || name == string(os.PathSeparator) || name == "" {
		return "package.bin"
	}
	return name
}

func uniquePath(p string) string {
	if !pathExists(p) {
		return p
	}
	ext := filepath.Ext(p)
	base := strings.TrimSuffix(filepath.Base(p), ext)
	dir := filepath.Dir(p)
	for i := 2; i < 10000; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s (%d)%s", base, i, ext))
		if !pathExists(candidate) {
			return candidate
		}
	}
	return filepath.Join(dir, fmt.Sprintf("%s-%d%s", base, time.Now().UnixNano(), ext))
}

func modrinthAPIBase() string {
	if base := strings.TrimSpace(os.Getenv("MMV_MODRINTH_API_BASE")); base != "" {
		return strings.TrimRight(base, "/")
	}
	return "https://api.modrinth.com/v2"
}

func curseForgeAPIBase() string {
	if base := strings.TrimSpace(os.Getenv("MMV_CURSEFORGE_API_BASE")); base != "" {
		return strings.TrimRight(base, "/")
	}
	return "https://api.curseforge.com"
}

func curseForgeWebBase() string {
	if base := strings.TrimSpace(os.Getenv("MMV_CURSEFORGE_WEB_BASE")); base != "" {
		return strings.TrimRight(base, "/")
	}
	return "https://www.curseforge.com"
}

func (a *App) handleModrinthSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	projectType := strings.TrimSpace(r.URL.Query().Get("type"))
	game := strings.TrimSpace(r.URL.Query().Get("gameVersion"))
	loader := strings.TrimSpace(r.URL.Query().Get("loader"))
	limit := r.URL.Query().Get("limit")
	if limit == "" {
		limit = "40"
	}

	facets := make([][]string, 0, 3)
	if projectType != "" && projectType != "all" {
		facets = append(facets, []string{"project_type:" + projectType})
	}
	if game != "" {
		facets = append(facets, []string{"versions:" + game})
	}
	if loader != "" && loader != "any" && loader != "vanilla" && projectType == "mod" {
		facets = append(facets, []string{"categories:" + loader})
	}
	fb, _ := json.Marshal(facets)
	u := modrinthAPIBase() + "/search"
	values := url.Values{}
	values.Set("query", q)
	values.Set("limit", limit)
	values.Set("index", "relevance")
	if len(facets) > 0 {
		values.Set("facets", string(fb))
	}
	req, _ := http.NewRequestWithContext(r.Context(), http.MethodGet, u+"?"+values.Encode(), nil)
	req.Header.Set("User-Agent", userAgent)
	res, err := a.httpClient.Do(req)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, APIError{Error: "Modrinth is unreachable: " + err.Error()})
		return
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 16<<10))
		writeJSON(w, http.StatusBadGateway, APIError{Error: fmt.Sprintf("Modrinth returned %s: %s", res.Status, strings.TrimSpace(string(body)))})
		return
	}
	var out ModrinthSearchResponse
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		writeJSON(w, http.StatusBadGateway, APIError{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (a *App) handleModrinthInstall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var in ModrinthInstallRequest
	if err := decodeJSON(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	if in.ProjectID == "" {
		writeJSON(w, http.StatusBadRequest, APIError{Error: "projectId is required"})
		return
	}
	a.mu.RLock()
	if in.GameVersion == "" {
		in.GameVersion = a.settings.GameVersion
	}
	if in.Loader == "" {
		in.Loader = a.settings.Loader
	}
	a.mu.RUnlock()

	visited := map[string]bool{}
	installed := make([]InstalledProject, 0, 4)
	if err := a.installModrinthProject(r.Context(), in.ProjectID, "", in.GameVersion, in.Loader, in.Target, false, visited, &installed); err != nil {
		writeJSON(w, http.StatusBadGateway, APIError{Error: err.Error()})
		return
	}
	if len(installed) == 0 {
		writeJSON(w, http.StatusInternalServerError, APIError{Error: "installer completed without a downloaded file"})
		return
	}
	main := installed[len(installed)-1]
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":           true,
		"project":      main.Project,
		"projectType":  main.ProjectType,
		"version":      main.Version,
		"file":         main.File,
		"path":         main.Path,
		"target":       main.Target,
		"installed":    installed,
		"dependencies": len(installed) - 1,
	})
}

func (a *App) installModrinthProject(ctx context.Context, projectID, exactVersionID, gameVersion, loader, requestedTarget string, dependency bool, visited map[string]bool, installed *[]InstalledProject) error {
	visitKey := projectID + "|" + exactVersionID
	if visited[visitKey] {
		return nil
	}
	visited[visitKey] = true

	project, err := a.fetchModrinthProject(ctx, projectID)
	if err != nil {
		return err
	}
	var version ModrinthVersion
	if exactVersionID != "" {
		version, err = a.fetchModrinthVersion(ctx, exactVersionID)
		if err != nil {
			return err
		}
	} else {
		versions, lookupErr := a.fetchModrinthVersions(ctx, projectID, gameVersion, loader, project.ProjectType)
		if lookupErr != nil {
			return lookupErr
		}
		if len(versions) == 0 {
			return fmt.Errorf("no compatible %s version was found for Minecraft %s / %s", project.Title, gameVersion, loader)
		}
		version = versions[0]
	}

	// Resolve required Modrinth dependencies before the requested project. Optional,
	// incompatible, and embedded dependencies are deliberately left untouched.
	for _, dep := range version.Dependencies {
		if dep.DependencyType != "required" || dep.ProjectID == "" {
			continue
		}
		if err := a.installModrinthProject(ctx, dep.ProjectID, dep.VersionID, gameVersion, loader, "auto", true, visited, installed); err != nil {
			return fmt.Errorf("required dependency for %s failed: %w", project.Title, err)
		}
	}

	if len(version.Files) == 0 {
		return fmt.Errorf("compatible version of %s has no downloadable files", project.Title)
	}
	file := version.Files[0]
	for _, f := range version.Files {
		if f.Primary {
			file = f
			break
		}
	}
	target := requestedTarget
	if target == "" || target == "auto" {
		switch project.ProjectType {
		case "mod":
			target = "mods"
		case "resourcepack":
			target = "resourcepacks"
		case "shader":
			target = "shaderpacks"
		default:
			target = "downloads"
		}
	}
	dir := a.javaTargetDir(target)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	dst := uniquePath(filepath.Join(dir, safeFilename(file.Filename)))
	if err := a.downloadVerified(ctx, file, dst); err != nil {
		_ = os.Remove(dst)
		return err
	}
	*installed = append(*installed, InstalledProject{
		Project: project.Title, ProjectID: project.ID, ProjectType: project.ProjectType,
		Version: version.VersionNumber, File: filepath.Base(dst), Path: dst, Target: target, Dependency: dependency,
	})
	return nil
}

func (a *App) fetchModrinthVersion(ctx context.Context, versionID string) (ModrinthVersion, error) {
	var out ModrinthVersion
	u := modrinthAPIBase() + "/version/" + url.PathEscape(versionID)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	req.Header.Set("User-Agent", userAgent)
	res, err := a.httpClient.Do(req)
	if err != nil {
		return out, fmt.Errorf("Modrinth dependency version lookup failed: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return out, fmt.Errorf("Modrinth dependency version lookup returned %s", res.Status)
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return out, err
	}
	return out, nil
}

func (a *App) fetchModrinthProject(ctx context.Context, id string) (ModrinthProject, error) {
	var out ModrinthProject
	u := modrinthAPIBase() + "/project/" + url.PathEscape(id)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	req.Header.Set("User-Agent", userAgent)
	res, err := a.httpClient.Do(req)
	if err != nil {
		return out, fmt.Errorf("Modrinth project lookup failed: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return out, fmt.Errorf("Modrinth project lookup returned %s", res.Status)
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return out, err
	}
	return out, nil
}

func (a *App) fetchModrinthVersions(ctx context.Context, id, game, loader, projectType string) ([]ModrinthVersion, error) {
	u := modrinthAPIBase() + "/project/" + url.PathEscape(id) + "/version"
	values := url.Values{}
	if game != "" {
		b, _ := json.Marshal([]string{game})
		values.Set("game_versions", string(b))
	}
	if projectType == "mod" && loader != "" && loader != "any" && loader != "vanilla" {
		b, _ := json.Marshal([]string{loader})
		values.Set("loaders", string(b))
	}
	if len(values) > 0 {
		u += "?" + values.Encode()
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	req.Header.Set("User-Agent", userAgent)
	res, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Modrinth version lookup failed: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("Modrinth version lookup returned %s", res.Status)
	}
	var out []ModrinthVersion
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, err
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].DatePublished > out[j].DatePublished })
	return out, nil
}

func (a *App) downloadVerified(ctx context.Context, f ModrinthVersionFile, dst string) error {
	u, err := url.Parse(f.URL)
	if err != nil || u.Scheme != "https" {
		return errors.New("download URL is not HTTPS")
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, f.URL, nil)
	req.Header.Set("User-Agent", userAgent)
	res, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("download returned %s", res.Status)
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	h512 := sha512.New()
	h1 := sha1.New()
	n, copyErr := io.Copy(io.MultiWriter(out, h512, h1), res.Body)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	if f.Size > 0 && n != f.Size {
		return fmt.Errorf("download size mismatch: expected %d bytes, got %d", f.Size, n)
	}
	if want := strings.ToLower(f.Hashes["sha512"]); want != "" {
		got := hex.EncodeToString(h512.Sum(nil))
		if got != want {
			return errors.New("SHA-512 verification failed")
		}
	} else if want := strings.ToLower(f.Hashes["sha1"]); want != "" {
		got := hex.EncodeToString(h1.Sum(nil))
		if got != want {
			return errors.New("SHA-1 verification failed")
		}
	}
	return nil
}

func pathExists(p string) bool {
	if p == "" {
		return false
	}
	_, err := os.Stat(p)
	return err == nil
}

func launchAppWindow(appURL string) error {
	switch runtime.GOOS {
	case "windows":
		candidates := []string{
			filepath.Join(os.Getenv("ProgramFiles(x86)"), "Microsoft", "Edge", "Application", "msedge.exe"),
			filepath.Join(os.Getenv("ProgramFiles"), "Microsoft", "Edge", "Application", "msedge.exe"),
			filepath.Join(os.Getenv("LOCALAPPDATA"), "Microsoft", "Edge", "Application", "msedge.exe"),
			filepath.Join(os.Getenv("ProgramFiles"), "Google", "Chrome", "Application", "chrome.exe"),
			filepath.Join(os.Getenv("ProgramFiles(x86)"), "Google", "Chrome", "Application", "chrome.exe"),
		}
		for _, c := range candidates {
			if c != "" && pathExists(c) {
				return exec.Command(c, "--app="+appURL, "--start-maximized", "--no-first-run").Start()
			}
		}
	case "darwin":
		if err := exec.Command("open", "-na", "Google Chrome", "--args", "--app="+appURL).Start(); err == nil {
			return nil
		}
	default:
		for _, name := range []string{"google-chrome", "chromium", "chromium-browser", "microsoft-edge", "brave-browser"} {
			if p, err := exec.LookPath(name); err == nil {
				return exec.Command(p, "--app="+appURL, "--start-maximized", "--no-first-run").Start()
			}
		}
	}
	return errors.New("no Chromium app-mode browser found")
}

func openExternal(target string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", target).Start()
	case "darwin":
		return exec.Command("open", target).Start()
	default:
		return exec.Command("xdg-open", target).Start()
	}
}
