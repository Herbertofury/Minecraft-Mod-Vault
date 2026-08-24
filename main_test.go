package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestTargetForFilename(t *testing.T) {
	cases := map[string]string{
		"example.jar":    "mods",
		"pack.zip":       "resourcepacks",
		"world.mcworld":  "downloads",
		"addon.mcaddon":  "downloads",
		"pack.mcpack":    "downloads",
		"profile.mrpack": "downloads",
		"unknown.txt":    "downloads",
		"UPPER.JAR":      "mods",
	}
	for name, want := range cases {
		if got := targetForFilename(name); got != want {
			t.Fatalf("targetForFilename(%q)=%q want %q", name, got, want)
		}
	}
}

func TestValidImportTargetIncludesWorldDatapacks(t *testing.T) {
	for _, target := range []string{"mods", "resourcepacks", "shaderpacks", "plugins", "datapacks", "worlds", "downloads"} {
		if !validImportTarget(target) {
			t.Fatalf("valid import target rejected: %s", target)
		}
	}
	for _, target := range []string{"", "../mods", "world", "arbitrary"} {
		if validImportTarget(target) {
			t.Fatalf("invalid import target accepted: %q", target)
		}
	}
}

func TestDuplicateMergeAvoidsSameTitleDifferentAuthors(t *testing.T) {
	items := []UnifiedProject{
		{ID: "1", Provider: "modrinth", ProjectType: "mod", Slug: "chairs", Title: "Chairs", Author: "Alice", PageURL: "https://example.invalid/a", Score: 20},
		{ID: "2", Provider: "curseforge", ProjectType: "mod", Slug: "chairs-reborn", Title: "Chairs", Author: "Bob", PageURL: "https://example.invalid/b", Score: 30},
		{ID: "3", Provider: "github", ProjectType: "mod", Slug: "chairs", Title: "Chairs", Author: "Alice", PageURL: "https://example.invalid/c", Score: 10},
	}
	merged := mergeProviderDuplicates(items)
	if len(merged) != 2 {
		t.Fatalf("same-title unrelated authors were over-merged: %#v", merged)
	}
	var alice UnifiedProject
	for _, item := range merged {
		if item.Author == "Alice" {
			alice = item
		}
	}
	if len(alice.Providers) != 2 || !containsFold(alice.Providers, "modrinth") || !containsFold(alice.Providers, "github") {
		t.Fatalf("true cross-provider duplicate failed to merge: %#v", alice)
	}
}

func TestProviderSearchSkipsSitesIrrelevantToSelectedContentType(t *testing.T) {
	app := &App{providerHealth: map[string]ProviderHealth{}, providerCache: map[string]providerCacheEntry{}}
	resp := app.searchProviders(context.Background(), providerSearchOptions{Query: "create", ProjectType: "mod", Limit: 10, Sources: []string{"spongeore"}})
	if len(resp.Results) != 0 || len(resp.Errors) != 0 {
		t.Fatalf("irrelevant provider should be skipped without querying: results=%#v errors=%#v", resp.Results, resp.Errors)
	}
	if resp.Skipped["spongeore"] == "" {
		t.Fatalf("expected explicit skipped reason, got %#v", resp.Skipped)
	}
}

func TestProviderSearchUsesStaleCacheWhenLiveProviderFails(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v2/packs", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]map[string]any{{
			"id": "cached-pack",
			"data": map[string]any{
				"display":  map[string]any{"name": "Cached Pack", "description": "survives a provider outage"},
				"versions": []map[string]any{{"name": "1.0.0", "supports": []string{"1.21.1"}, "downloads": map[string]any{"datapack": "https://cdn.example/cached.zip"}}},
			},
			"meta": map[string]any{"owner": "maker", "stats": map[string]any{"downloads": map[string]any{"total": 42}}},
		}})
	})
	server := httptest.NewTLSServer(mux)
	t.Setenv("MMV_SMITHED_API_BASE", server.URL+"/v2")
	app := &App{httpClient: server.Client(), providerHealth: map[string]ProviderHealth{}, providerCache: map[string]providerCacheEntry{}}
	opts := providerSearchOptions{Query: "cached", ProjectType: "datapack", GameVersion: "1.21.1", Limit: 10, Sources: []string{"smithed"}}
	first := app.searchProviders(context.Background(), opts)
	if len(first.Results) != 1 || first.Results[0].Title != "Cached Pack" || !first.Live {
		server.Close()
		t.Fatalf("initial live result did not populate cache: %#v", first)
	}
	key := providerSearchCacheKey("smithed", "cached", opts)
	app.providerCacheMu.Lock()
	entry := app.providerCache[key]
	entry.At = time.Now().Add(-2 * providerFreshCacheTTL)
	app.providerCache[key] = entry
	app.providerCacheMu.Unlock()
	server.Close()

	second := app.searchProviders(context.Background(), opts)
	if len(second.Results) != 1 || second.Results[0].Title != "Cached Pack" {
		t.Fatalf("cached provider result was lost after outage: %#v", second)
	}
	if second.Live || second.Warnings["smithed"] == "" || len(second.Errors) != 0 {
		t.Fatalf("stale-cache fallback state incorrect: live=%v warnings=%#v errors=%#v", second.Live, second.Warnings, second.Errors)
	}
}

func TestSafeFilename(t *testing.T) {
	cases := map[string]string{
		"../evil.jar":     "evil.jar",
		"folder/pack.zip": "pack.zip",
		"":                "package.bin",
		".":               "package.bin",
		"ok.mcpack":       "ok.mcpack",
	}
	for input, want := range cases {
		if got := safeFilename(input); got != want {
			t.Fatalf("safeFilename(%q)=%q want %q", input, got, want)
		}
	}
}

func TestAllowedManagedPath(t *testing.T) {
	root := t.TempDir()
	cfg := t.TempDir()
	app := &App{cfgDir: cfg, settings: Settings{JavaRoot: root}}
	inside := filepath.Join(root, "mods", "example.jar")
	outside := filepath.Join(t.TempDir(), "example.jar")
	if !app.allowedManagedPath(inside) {
		t.Fatalf("expected path inside Java root to be allowed: %s", inside)
	}
	if !app.allowedManagedPath(filepath.Join(cfg, "downloads", "pack.zip")) {
		t.Fatal("expected vault download path to be allowed")
	}
	if app.allowedManagedPath(outside) {
		t.Fatalf("expected outside path to be rejected: %s", outside)
	}
}

func TestCopyThenRemoveDirectory(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src")
	dst := filepath.Join(root, "dst")
	if err := os.MkdirAll(filepath.Join(src, "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	payload := []byte("minecraft-mod-vault")
	if err := os.WriteFile(filepath.Join(src, "nested", "file.txt"), payload, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := copyThenRemove(src, dst); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Fatalf("source directory still exists or unexpected stat error: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dst, "nested", "file.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(payload) {
		t.Fatalf("copied data mismatch: %q", got)
	}
}

func TestUniquePathPreservesExisting(t *testing.T) {
	dir := t.TempDir()
	first := filepath.Join(dir, "pack.zip")
	if err := os.WriteFile(first, []byte("one"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := uniquePath(first)
	want := filepath.Join(dir, "pack (2).zip")
	if got != want {
		t.Fatalf("uniquePath=%q want %q", got, want)
	}
}

func TestInstallModrinthProjectResolvesRequiredDependency(t *testing.T) {
	depBytes := []byte("dependency-jar")
	mainBytes := []byte("main-mod-jar")
	sha512hex := func(b []byte) string {
		sum := sha512.Sum512(b)
		return hex.EncodeToString(sum[:])
	}

	mux := http.NewServeMux()
	var server *httptest.Server
	mux.HandleFunc("/v2/project/main-mod", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(ModrinthProject{ID: "main-id", Slug: "main-mod", Title: "Main Furniture Mod", ProjectType: "mod"})
	})
	mux.HandleFunc("/v2/project/dep-mod", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(ModrinthProject{ID: "dep-id", Slug: "dep-mod", Title: "Furniture API", ProjectType: "mod"})
	})
	mux.HandleFunc("/v2/project/main-mod/version", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]ModrinthVersion{{
			ID: "main-v1", ProjectID: "main-id", VersionNumber: "1.0.0", DatePublished: "2026-08-01T00:00:00Z",
			GameVersions: []string{"1.21.1"}, Loaders: []string{"fabric"},
			Dependencies: []ModrinthDependency{{ProjectID: "dep-mod", DependencyType: "required"}},
			Files:        []ModrinthVersionFile{{URL: server.URL + "/files/main.jar", Filename: "main.jar", Primary: true, Size: int64(len(mainBytes)), Hashes: map[string]string{"sha512": sha512hex(mainBytes)}}},
		}})
	})
	mux.HandleFunc("/v2/project/dep-mod/version", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]ModrinthVersion{{
			ID: "dep-v1", ProjectID: "dep-id", VersionNumber: "2.0.0", DatePublished: "2026-07-01T00:00:00Z",
			GameVersions: []string{"1.21.1"}, Loaders: []string{"fabric"},
			Files: []ModrinthVersionFile{{URL: server.URL + "/files/dep.jar", Filename: "dep.jar", Primary: true, Size: int64(len(depBytes)), Hashes: map[string]string{"sha512": sha512hex(depBytes)}}},
		}})
	})
	mux.HandleFunc("/files/main.jar", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(mainBytes) })
	mux.HandleFunc("/files/dep.jar", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(depBytes) })
	server = httptest.NewTLSServer(mux)
	defer server.Close()

	t.Setenv("MMV_MODRINTH_API_BASE", server.URL+"/v2")
	root := t.TempDir()
	app := &App{settings: Settings{JavaRoot: root, GameVersion: "1.21.1", Loader: "fabric"}, httpClient: server.Client()}
	installed := []InstalledProject{}
	if err := app.installModrinthProject(context.Background(), "main-mod", "", "1.21.1", "fabric", "mods", false, map[string]bool{}, &installed); err != nil {
		t.Fatal(err)
	}
	if len(installed) != 2 {
		t.Fatalf("installed %d projects, want 2", len(installed))
	}
	if !installed[0].Dependency || installed[0].Project != "Furniture API" {
		t.Fatalf("dependency not installed first: %#v", installed)
	}
	if installed[1].Dependency || installed[1].Project != "Main Furniture Mod" {
		t.Fatalf("main project missing: %#v", installed)
	}
	for _, name := range []string{"dep.jar", "main.jar"} {
		if _, err := os.Stat(filepath.Join(root, "mods", name)); err != nil {
			t.Fatalf("expected installed file %s: %v", name, err)
		}
	}
}

func TestEmbeddedRootServesIndexWithoutRedirect(t *testing.T) {
	app := &App{token: "test-token"}
	mux := http.NewServeMux()
	app.registerRoutes(mux)
	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/?token=test-token", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("root status=%d want %d; Location=%q", rec.Code, http.StatusOK, rec.Header().Get("Location"))
	}
	if !strings.Contains(rec.Body.String(), "Minecraft Mod Vault") {
		t.Fatal("embedded root did not serve the application shell")
	}
}

func TestUniversalModsEmbeddedAssets(t *testing.T) {
	app := &App{token: "test-token"}
	mux := http.NewServeMux()
	app.registerRoutes(mux)

	get := func(path string) string {
		req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1"+path, nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("GET %s status=%d want %d", path, rec.Code, http.StatusOK)
		}
		return rec.Body.String()
	}

	index := get("/?token=test-token")
	for _, marker := range []string{`data-view="mods"`, `id="providerStrip"`, `id="modCategoryStrip"`, `data-view="updater"`, `data-view="creators"`, `id="recommendationGrid"`} {
		if !strings.Contains(index, marker) {
			t.Fatalf("embedded index missing %q", marker)
		}
	}
	appJS := get("/app.js?token=test-token")
	for _, marker := range []string{"Universal multi-source mod discovery", "author-avatar", "/api/providers/search", "/api/updater/scan", "/api/creators/analyze"} {
		if !strings.Contains(appJS, marker) {
			t.Fatalf("embedded app.js missing %q", marker)
		}
	}

	catalog := get("/catalog.js?token=test-token")
	for _, marker := range []string{
		"Railroads & Trains", "Vehicles & Ships", "Tech & Automation", "Magic & Spellcraft",
		"Living Foliage", "Particles & Water FX", "Cards & Collectibles", "Pets & Companions",
		"Creature Collecting", "Farming & Cooking", "Exploration & Worldgen", "Building & Decor",
		"Storage & Quality of Life", "Mobs & Bosses", "Performance & Visuals",
	} {
		if !strings.Contains(catalog, marker) {
			t.Fatalf("embedded catalog missing category %q", marker)
		}
	}
}

func TestAppVersion0110(t *testing.T) {
	if appVersion != "0.13.0" {
		t.Fatalf("appVersion=%q want 0.13.0", appVersion)
	}
}

func TestCurseForgeFingerprintIgnoresWhitespace(t *testing.T) {
	plain := curseForgeFingerprint([]byte("abc123"))
	spaced := curseForgeFingerprint([]byte("a b\nc\t1\r2 3"))
	if plain != spaced {
		t.Fatalf("fingerprint whitespace compatibility failed: %d != %d", plain, spaced)
	}
	if plain == 0 {
		t.Fatal("fingerprint unexpectedly zero")
	}
}

func fabricJarBytes(t *testing.T, id, name, version string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("fabric.mod.json")
	if err != nil {
		t.Fatal(err)
	}
	meta := map[string]any{"schemaVersion": 1, "id": id, "name": name, "version": version, "contact": map[string]string{"sources": "https://github.com/example/" + id}, "depends": map[string]string{"minecraft": ">=1.21"}}
	if err := json.NewEncoder(w).Encode(meta); err != nil {
		t.Fatal(err)
	}
	payload, err := zw.Create("data/example.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = payload.Write([]byte("payload-" + version))
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestParseFabricJarMetadata(t *testing.T) {
	data := fabricJarBytes(t, "cozy_mod", "Cozy Mod", "1.2.3")
	meta, err := parseJarMetadataBytes(data)
	if err != nil {
		t.Fatal(err)
	}
	if meta.ModID != "cozy_mod" || meta.Name != "Cozy Mod" || meta.Version != "1.2.3" {
		t.Fatalf("unexpected metadata: %#v", meta)
	}
	if len(meta.Loaders) != 1 || meta.Loaders[0] != "fabric" {
		t.Fatalf("unexpected loaders: %#v", meta.Loaders)
	}
}

func TestUnifiedModrinthSearchIncludesImagesAndAuthorAvatar(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v2/search", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"hits": []map[string]any{{
			"project_id": "p1", "project_type": "mod", "slug": "cozy-mod", "author": "builder",
			"title": "Cozy Mod", "description": "Furniture and cozy decor", "categories": []string{"fabric", "decoration"},
			"versions": []string{"1.21.1"}, "downloads": 123456, "follows": 500, "icon_url": "https://cdn.example/icon.png",
			"gallery": []string{"https://cdn.example/gallery.png"}, "date_modified": "2026-08-18T00:00:00Z",
		}}, "total_hits": 1})
	})
	mux.HandleFunc("/v2/project/p1/members", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]map[string]any{{"user": map[string]any{"username": "builder", "avatar_url": "https://cdn.example/avatar.png"}, "role": "Owner"}})
	})
	server := httptest.NewTLSServer(mux)
	defer server.Close()
	t.Setenv("MMV_MODRINTH_API_BASE", server.URL+"/v2")
	app := &App{httpClient: server.Client()}
	resp := app.searchProviders(context.Background(), providerSearchOptions{Query: "cozy", GameVersion: "1.21.1", Loader: "fabric", ProjectType: "mod", Limit: 20, Sources: []string{"modrinth"}})
	if len(resp.Results) != 1 {
		t.Fatalf("got %d results; errors=%v", len(resp.Results), resp.Errors)
	}
	got := resp.Results[0]
	if got.IconURL == "" || len(got.Gallery) == 0 || got.AuthorAvatarURL == "" {
		t.Fatalf("rich media regression: %#v", got)
	}
	if got.Author != "builder" || got.Provider != "modrinth" {
		t.Fatalf("metadata mismatch: %#v", got)
	}
}

func TestUpdaterExactHashPlanAndTransactionalApply(t *testing.T) {
	oldJar := fabricJarBytes(t, "cozy_mod", "Cozy Mod", "1.0.0")
	newJar := fabricJarBytes(t, "cozy_mod", "Cozy Mod", "2.0.0")
	oldSum := sha512.Sum512(oldJar)
	newSum := sha512.Sum512(newJar)
	oldHash := hex.EncodeToString(oldSum[:])
	newHash := hex.EncodeToString(newSum[:])

	mux := http.NewServeMux()
	var server *httptest.Server
	installed := ModrinthVersion{ID: "old-v", ProjectID: "p1", VersionNumber: "1.0.0", GameVersions: []string{"1.20.1"}, Loaders: []string{"fabric"}}
	updated := ModrinthVersion{ID: "new-v", ProjectID: "p1", VersionNumber: "2.0.0", GameVersions: []string{"1.21.1"}, Loaders: []string{"fabric"}, Files: []ModrinthVersionFile{{Filename: "cozy-mod-2.0.0.jar", Primary: true, Size: int64(len(newJar)), Hashes: map[string]string{"sha512": newHash}}}}
	mux.HandleFunc("/v2/version_files", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]ModrinthVersion{oldHash: installed})
	})
	mux.HandleFunc("/v2/version_files/update", func(w http.ResponseWriter, r *http.Request) {
		copy := updated
		copy.Files[0].URL = server.URL + "/files/new.jar"
		json.NewEncoder(w).Encode(map[string]ModrinthVersion{oldHash: copy})
	})
	mux.HandleFunc("/v2/project/p1", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(ModrinthProject{ID: "p1", Slug: "cozy-mod", Title: "Cozy Mod", ProjectType: "mod"})
	})
	mux.HandleFunc("/v2/version/new-v", func(w http.ResponseWriter, r *http.Request) {
		copy := updated
		copy.Files[0].URL = server.URL + "/files/new.jar"
		json.NewEncoder(w).Encode(copy)
	})
	mux.HandleFunc("/files/new.jar", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(newJar) })
	server = httptest.NewTLSServer(mux)
	defer server.Close()
	t.Setenv("MMV_MODRINTH_API_BASE", server.URL+"/v2")

	root := t.TempDir()
	mods := filepath.Join(root, "mods")
	if err := os.MkdirAll(mods, 0o755); err != nil {
		t.Fatal(err)
	}
	oldPath := filepath.Join(mods, "cozy-mod-1.0.0.jar")
	if err := os.WriteFile(oldPath, oldJar, 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := t.TempDir()
	app := &App{cfgDir: cfg, settings: Settings{JavaRoot: root, GameVersion: "1.21.1", Loader: "fabric"}, httpClient: server.Client(), updatePlans: map[string]UpdatePlan{}}
	plan, err := app.buildUpdatePlan(context.Background(), "1.21.1", "fabric")
	if err != nil {
		t.Fatal(err)
	}
	if plan.SafeCount != 1 || len(plan.Items) != 1 || plan.Items[0].SafeUpdate == nil {
		t.Fatalf("unexpected plan: %#v", plan)
	}
	result, err := app.applyUpdatePlan(context.Background(), plan, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result["count"].(int) != 1 {
		t.Fatalf("unexpected apply result: %#v", result)
	}
	newPath := filepath.Join(mods, "cozy-mod-2.0.0.jar")
	if _, err := os.Stat(newPath); err != nil {
		t.Fatalf("new mod missing: %v", err)
	}
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Fatalf("old path still present: %v", err)
	}
	meta, err := parseJarMetadataPath(newPath)
	if err != nil || meta.Version != "2.0.0" {
		t.Fatalf("installed jar validation failed: %#v %v", meta, err)
	}
	backupDir := result["backupDir"].(string)
	if _, err := os.Stat(filepath.Join(backupDir, "cozy-mod-1.0.0.jar")); err != nil {
		t.Fatalf("backup missing: %v", err)
	}
}

func TestDeclaredProviderMetadataUpdater(t *testing.T) {
	t.Run("Modrinth canonical metadata yields safe compatible update", func(t *testing.T) {
		mux := http.NewServeMux()
		var server *httptest.Server
		mux.HandleFunc("/v2/project/cozy-mod", func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(ModrinthProject{ID: "p1", Slug: "cozy-mod", Title: "Cozy Mod", ProjectType: "mod"})
		})
		mux.HandleFunc("/v2/project/p1/version", func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.URL.Query().Get("game_versions"), "1.21.1") || !strings.Contains(r.URL.Query().Get("loaders"), "fabric") {
				t.Fatalf("missing compatibility filters: %s", r.URL.RawQuery)
			}
			_ = json.NewEncoder(w).Encode([]ModrinthVersion{{
				ID: "v2", ProjectID: "p1", VersionNumber: "2.0.0", DatePublished: "2026-08-19T00:00:00Z",
				GameVersions: []string{"1.21.1"}, Loaders: []string{"fabric"},
				Files: []ModrinthVersionFile{{Filename: "cozy-mod-2.0.0.jar", URL: server.URL + "/cozy.jar", Primary: true, Size: 123, Hashes: map[string]string{"sha512": strings.Repeat("ab", 64)}}},
			}})
		})
		server = httptest.NewTLSServer(mux)
		defer server.Close()
		t.Setenv("MMV_MODRINTH_API_BASE", server.URL+"/v2")
		app := &App{httpClient: server.Client()}
		plan := UpdatePlan{Items: []UpdateItem{{Local: LocalModFile{Filename: "cozy-mod-1.0.0.jar", Metadata: JarMetadata{ModID: "cozy_mod", Name: "Cozy Mod", Version: "1.0.0", Homepage: "https://modrinth.com/mod/cozy-mod"}}, Status: "unknown"}}}
		app.lookupDeclaredProviderMetadataUpdates(context.Background(), &plan, "1.21.1", "fabric")
		item := plan.Items[0]
		if item.Status != "update" || item.SafeUpdate == nil || !item.SafeUpdate.Safe || item.SafeUpdate.Provider != "modrinth" {
			t.Fatalf("expected safe canonical Modrinth update, got %#v", item)
		}
		if item.SafeUpdate.Confidence < .99 || !strings.Contains(item.SafeUpdate.Reason, "downloaded mod ID") {
			t.Fatalf("missing metadata-update safeguards: %#v", item.SafeUpdate)
		}
	})

	t.Run("Modrinth canonical metadata recognizes current declared version", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/v2/project/cozy-mod", func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(ModrinthProject{ID: "p1", Slug: "cozy-mod", Title: "Cozy Mod", ProjectType: "mod"})
		})
		mux.HandleFunc("/v2/project/p1/version", func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode([]ModrinthVersion{{ID: "v2", ProjectID: "p1", VersionNumber: "2.0.0", DatePublished: "2026-08-19T00:00:00Z", Files: []ModrinthVersionFile{{Filename: "cozy-mod-2.0.0.jar", URL: "https://cdn.example/cozy.jar", Primary: true}}}})
		})
		server := httptest.NewTLSServer(mux)
		defer server.Close()
		t.Setenv("MMV_MODRINTH_API_BASE", server.URL+"/v2")
		app := &App{httpClient: server.Client()}
		plan := UpdatePlan{Items: []UpdateItem{{Local: LocalModFile{Filename: "cozy-mod-2.0.0.jar", Metadata: JarMetadata{ModID: "cozy_mod", Name: "Cozy Mod", Version: "2.0.0", SourceURL: "https://modrinth.com/mod/cozy-mod"}}, Status: "unknown"}}}
		app.lookupDeclaredProviderMetadataUpdates(context.Background(), &plan, "1.21.1", "fabric")
		if plan.Items[0].Status != "current" || plan.Items[0].SafeUpdate != nil {
			t.Fatalf("expected current canonical Modrinth version, got %#v", plan.Items[0])
		}
	})

	t.Run("CurseForge canonical metadata yields safe dependency-free update", func(t *testing.T) {
		mux := http.NewServeMux()
		var server *httptest.Server
		mux.HandleFunc("/v1/mods/search", func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("x-api-key") != "test-key" || r.URL.Query().Get("slug") != "cozy-mod" {
				t.Fatalf("unexpected CurseForge identity request: key=%q query=%q", r.Header.Get("x-api-key"), r.URL.RawQuery)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{"id": 777, "name": "Cozy Mod", "slug": "cozy-mod"}}})
		})
		mux.HandleFunc("/v1/mods/777/files", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("gameVersion") != "1.21.1" || r.URL.Query().Get("modLoaderType") != "4" {
				t.Fatalf("missing CurseForge compatibility filters: %s", r.URL.RawQuery)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{
				"id": 9002, "fileName": "cozy-mod-2.0.0.jar", "displayName": "Cozy Mod 2.0.0", "downloadUrl": server.URL + "/cozy.jar", "fileLength": 123,
				"gameVersions": []string{"1.21.1", "Fabric"}, "hashes": []map[string]any{{"value": strings.Repeat("ab", 20), "algo": 1}}, "dependencies": []any{},
			}}})
		})
		server = httptest.NewTLSServer(mux)
		defer server.Close()
		t.Setenv("MMV_CURSEFORGE_API_BASE", server.URL)
		app := &App{httpClient: server.Client(), settings: Settings{CurseForgeAPIKey: "test-key"}}
		plan := UpdatePlan{Items: []UpdateItem{{Local: LocalModFile{Filename: "cozy-mod-1.0.0.jar", Metadata: JarMetadata{ModID: "cozy_mod", Name: "Cozy Mod", Version: "1.0.0", Homepage: "https://www.curseforge.com/minecraft/mc-mods/cozy-mod"}}, Status: "unknown"}}}
		app.lookupDeclaredProviderMetadataUpdates(context.Background(), &plan, "1.21.1", "fabric")
		item := plan.Items[0]
		if item.Status != "update" || item.SafeUpdate == nil || !item.SafeUpdate.Safe || item.SafeUpdate.Provider != "curseforge" {
			t.Fatalf("expected safe canonical CurseForge update, got %#v", item)
		}
	})

	t.Run("CurseForge canonical metadata keeps new required dependencies review-only", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/v1/mods/search", func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{"id": 777, "name": "Cozy Mod", "slug": "cozy-mod"}}})
		})
		mux.HandleFunc("/v1/mods/777/files", func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{
				"id": 9002, "fileName": "cozy-mod-2.0.0.jar", "displayName": "Cozy Mod 2.0.0", "downloadUrl": "https://cdn.example/cozy.jar", "fileLength": 123,
				"gameVersions": []string{"1.21.1", "Fabric"}, "dependencies": []map[string]any{{"modId": 42, "relationType": 3}},
			}}})
		})
		server := httptest.NewTLSServer(mux)
		defer server.Close()
		t.Setenv("MMV_CURSEFORGE_API_BASE", server.URL)
		app := &App{httpClient: server.Client(), settings: Settings{CurseForgeAPIKey: "test-key"}}
		plan := UpdatePlan{Items: []UpdateItem{{Local: LocalModFile{Filename: "cozy-mod-1.0.0.jar", Metadata: JarMetadata{ModID: "cozy_mod", Name: "Cozy Mod", Version: "1.0.0", SourceURL: "https://www.curseforge.com/minecraft/mc-mods/cozy-mod"}}, Status: "unknown"}}}
		app.lookupDeclaredProviderMetadataUpdates(context.Background(), &plan, "1.21.1", "fabric")
		item := plan.Items[0]
		if item.Status != "review" || item.SafeUpdate != nil || len(item.Alternatives) == 0 || item.Alternatives[0].Safe || !item.Alternatives[0].DependencyRisk {
			t.Fatalf("dependency-changing CurseForge update must remain review-only: %#v", item)
		}
	})
}

func TestCurseForgeNativeInstallResolvesExactSlugAndRequiredDependencies(t *testing.T) {
	mainJar := fabricJarBytes(t, "cozy_mod", "Cozy Mod", "2.0.0")
	depJar := fabricJarBytes(t, "cozy_lib", "Cozy Lib", "1.0.0")
	mux := http.NewServeMux()
	var server *httptest.Server
	mux.HandleFunc("/v1/mods/search", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("slug") != "cozy-mod" {
			t.Fatalf("installer did not use exact slug lookup: %s", r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{
			{"id": 999, "name": "Wrong fuzzy project", "slug": "cozy-mod-extra"},
			{"id": 777, "name": "Cozy Mod", "slug": "cozy-mod"},
		}})
	})
	mux.HandleFunc("/v1/mods/777/files", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{
			"id": 7001, "fileName": "cozy-mod-2.0.0.jar", "displayName": "Cozy Mod 2.0.0", "downloadUrl": server.URL + "/main.jar", "fileLength": len(mainJar),
			"dependencies": []map[string]any{{"modId": 42, "relationType": 3}},
		}}})
	})
	mux.HandleFunc("/v1/mods/42/files", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{
			"id": 4201, "fileName": "cozy-lib-1.0.0.jar", "displayName": "Cozy Lib 1.0.0", "downloadUrl": server.URL + "/dep.jar", "fileLength": len(depJar),
		}}})
	})
	mux.HandleFunc("/main.jar", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(mainJar) })
	mux.HandleFunc("/dep.jar", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(depJar) })
	server = httptest.NewTLSServer(mux)
	defer server.Close()
	t.Setenv("MMV_CURSEFORGE_API_BASE", server.URL)
	root := t.TempDir()
	app := &App{httpClient: server.Client(), settings: Settings{CurseForgeAPIKey: "test-key", JavaRoot: root}}
	result, err := app.installCurseForge(context.Background(), "cozy-mod", "1.21.1", "fabric", "mods")
	if err != nil {
		t.Fatal(err)
	}
	installed, ok := result["installed"].([]map[string]any)
	if !ok || len(installed) != 2 {
		t.Fatalf("expected dependency + main install, got %#v", result)
	}
	if installed[0]["dependency"] != true || installed[0]["projectId"] != int64(42) || installed[1]["dependency"] != false || installed[1]["projectId"] != int64(777) {
		t.Fatalf("unexpected dependency install ordering: %#v", installed)
	}
	for _, name := range []string{"cozy-lib-1.0.0.jar", "cozy-mod-2.0.0.jar"} {
		if _, err := os.Stat(filepath.Join(root, "mods", name)); err != nil {
			t.Fatalf("installed CurseForge file %s missing: %v", name, err)
		}
	}
}

func TestCurseForgeNativeInstallRollsBackNewDependenciesWhenMainFails(t *testing.T) {
	depJar := fabricJarBytes(t, "cozy_lib", "Cozy Lib", "1.0.0")
	mux := http.NewServeMux()
	var server *httptest.Server
	mux.HandleFunc("/v1/mods/search", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{"id": 777, "name": "Cozy Mod", "slug": "cozy-mod"}}})
	})
	mux.HandleFunc("/v1/mods/777/files", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{
			"id": 7001, "fileName": "cozy-mod-bad.jar", "displayName": "Cozy Mod bad", "downloadUrl": server.URL + "/bad.jar", "fileLength": 16,
			"dependencies": []map[string]any{{"modId": 42, "relationType": 3}},
		}}})
	})
	mux.HandleFunc("/v1/mods/42/files", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{
			"id": 4201, "fileName": "cozy-lib-1.0.0.jar", "displayName": "Cozy Lib 1.0.0", "downloadUrl": server.URL + "/dep.jar", "fileLength": len(depJar),
		}}})
	})
	mux.HandleFunc("/dep.jar", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(depJar) })
	mux.HandleFunc("/bad.jar", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("not a jar at all")) })
	server = httptest.NewTLSServer(mux)
	defer server.Close()
	t.Setenv("MMV_CURSEFORGE_API_BASE", server.URL)
	root := t.TempDir()
	app := &App{httpClient: server.Client(), settings: Settings{CurseForgeAPIKey: "test-key", JavaRoot: root}}
	if _, err := app.installCurseForge(context.Background(), "cozy-mod", "1.21.1", "fabric", "mods"); err == nil {
		t.Fatal("invalid main CurseForge package unexpectedly installed")
	}
	if _, err := os.Stat(filepath.Join(root, "mods", "cozy-lib-1.0.0.jar")); !os.IsNotExist(err) {
		t.Fatalf("new dependency was not rolled back after main install failure: %v", err)
	}
}

func TestGitHubMetadataUpdaterRequiresExplicitTargetCompatibility(t *testing.T) {
	t.Run("safe exact metadata match", func(t *testing.T) {
		mux := http.NewServeMux()
		var server *httptest.Server
		mux.HandleFunc("/repos/example/cozy-mod/releases", func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode([]map[string]any{{
				"tag_name": "v2.0.0", "name": "Cozy Mod 2.0.0 for Minecraft 1.21.1 Fabric", "body": "Minecraft 1.21.1 Fabric release", "draft": false,
				"assets": []map[string]any{{"name": "cozy-mod-2.0.0-mc1.21.1-fabric.jar", "browser_download_url": server.URL + "/files/cozy.jar", "size": 123, "digest": "sha256:" + strings.Repeat("ab", 32)}},
			}})
		})
		server = httptest.NewTLSServer(mux)
		defer server.Close()
		t.Setenv("MMV_GITHUB_API_BASE", server.URL)
		app := &App{httpClient: server.Client(), settings: Settings{}}
		plan := UpdatePlan{Items: []UpdateItem{{Local: LocalModFile{Filename: "cozy-mod-1.0.0.jar", Metadata: JarMetadata{ModID: "cozy_mod", Name: "Cozy Mod", Version: "1.0.0", SourceURL: "https://github.com/example/cozy-mod"}}, Status: "unknown"}}}
		app.lookupGitHubMetadataUpdates(context.Background(), &plan, "1.21.1", "fabric")
		item := plan.Items[0]
		if item.Status != "update" || item.SafeUpdate == nil || !item.SafeUpdate.Safe || item.SafeUpdate.Provider != "github" {
			t.Fatalf("expected safe exact GitHub update, got %#v", item)
		}
		if item.SafeUpdate.Hashes["sha256"] == "" || item.SafeUpdate.Confidence != 1 {
			t.Fatalf("GitHub digest/confidence missing: %#v", item.SafeUpdate)
		}
	})

	t.Run("review when loader is not proven", func(t *testing.T) {
		mux := http.NewServeMux()
		var server *httptest.Server
		mux.HandleFunc("/repos/example/cozy-mod/releases", func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode([]map[string]any{{
				"tag_name": "v2.0.0", "name": "Cozy Mod for Minecraft 1.21.1", "body": "Minecraft 1.21.1 release", "draft": false,
				"assets": []map[string]any{{"name": "cozy-mod-2.0.0.jar", "browser_download_url": server.URL + "/files/cozy.jar", "size": 123}},
			}})
		})
		server = httptest.NewTLSServer(mux)
		defer server.Close()
		t.Setenv("MMV_GITHUB_API_BASE", server.URL)
		app := &App{httpClient: server.Client(), settings: Settings{}}
		plan := UpdatePlan{Items: []UpdateItem{{Local: LocalModFile{Filename: "cozy-mod-1.0.0.jar", Metadata: JarMetadata{Name: "Cozy Mod", Version: "1.0.0", SourceURL: "https://github.com/example/cozy-mod"}}, Status: "unknown"}}}
		app.lookupGitHubMetadataUpdates(context.Background(), &plan, "1.21.1", "neoforge")
		item := plan.Items[0]
		if item.Status != "review" || item.SafeUpdate != nil || len(item.Alternatives) == 0 || item.Alternatives[0].Safe {
			t.Fatalf("ambiguous GitHub release must stay review-only: %#v", item)
		}
	})
}

func TestTransactionalUpdaterRejectsDownloadedJarForWrongLoader(t *testing.T) {
	oldJar := fabricJarBytes(t, "cozy_mod", "Cozy Mod", "1.0.0")
	candidateJar := fabricJarBytes(t, "cozy_mod", "Cozy Mod", "2.0.0")
	mux := http.NewServeMux()
	mux.HandleFunc("/candidate.jar", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(candidateJar) })
	server := httptest.NewTLSServer(mux)
	defer server.Close()
	root := t.TempDir()
	mods := filepath.Join(root, "mods")
	if err := os.MkdirAll(mods, 0o755); err != nil {
		t.Fatal(err)
	}
	oldPath := filepath.Join(mods, "cozy-mod-1.0.0.jar")
	if err := os.WriteFile(oldPath, oldJar, 0o644); err != nil {
		t.Fatal(err)
	}
	app := &App{cfgDir: t.TempDir(), httpClient: server.Client()}
	plan := UpdatePlan{ID: "abcdefghijkl", Loader: "forge", ModsDirectory: mods, Items: []UpdateItem{{
		Local: LocalModFile{Path: oldPath, Filename: filepath.Base(oldPath), Enabled: true, Metadata: JarMetadata{ModID: "cozy_mod", Name: "Cozy Mod", Version: "1.0.0"}}, Status: "update",
		SafeUpdate: &UpdateCandidate{Provider: "github", ProjectID: "example/cozy-mod", Version: "2.0.0", Filename: "cozy-mod-2.0.0.jar", URL: server.URL + "/candidate.jar", Safe: true},
	}}}
	if _, err := app.applyUpdatePlan(context.Background(), plan, nil); err == nil || !strings.Contains(err.Error(), "loader validation") {
		t.Fatalf("wrong-loader candidate was not rejected: %v", err)
	}
	if _, err := os.Stat(oldPath); err != nil {
		t.Fatalf("original mod was not preserved after rejected update: %v", err)
	}
	if _, err := os.Stat(filepath.Join(mods, "cozy-mod-2.0.0.jar")); !os.IsNotExist(err) {
		t.Fatalf("wrong-loader update unexpectedly reached mods directory: %v", err)
	}
}

func TestArchiveTargetRejectsTraversal(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"../evil.exe", "../../outside/file", "folder/../../../escape"} {
		if got, err := archiveTarget(root, name); err == nil {
			t.Fatalf("archiveTarget(%q)=%q; expected traversal rejection", name, got)
		}
	}
	got, err := archiveTarget(root, "tools/bin/whisper-cli")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(got, root) {
		t.Fatalf("safe archive target escaped root: %q", got)
	}
}

func TestReleaseAssetDigestSHA256(t *testing.T) {
	valid := strings.Repeat("ab", 32)
	if got := releaseAssetDigestSHA256(githubReleaseAsset{Digest: "sha256:" + valid}); got != valid {
		t.Fatalf("valid digest=%q want %q", got, valid)
	}
	for _, digest := range []string{"", "sha1:" + valid, "sha256:not-hex", "sha256:abcd"} {
		if got := releaseAssetDigestSHA256(githubReleaseAsset{Digest: digest}); got != "" {
			t.Fatalf("invalid digest %q accepted as %q", digest, got)
		}
	}
}

func TestUniversalTaxonomyHasLivingModsBrowser(t *testing.T) {
	want := map[string]bool{"furniture": false, "cit": false, "railroads": false, "vehicles": false, "technology": false, "magic": false, "foliage": false, "particles": false, "cards": false, "pets": false, "creature-collecting": false}
	for _, cat := range universalTaxonomy {
		if strings.Contains(strings.ToLower(cat.Name), "amazing mods") {
			t.Fatalf("legacy Amazing Mods category still present: %q", cat.Name)
		}
		if _, ok := want[cat.ID]; ok {
			want[cat.ID] = true
		}
	}
	for id, found := range want {
		if !found {
			t.Errorf("taxonomy missing custom category %q", id)
		}
	}
}

func TestCreatorVideosOnlyExposeCompletedAnalysis(t *testing.T) {
	app := &App{creatorVideos: []CreatorVideo{
		{ID: "pending", Platform: "youtube", Title: "Pending", AnalyzedAt: ""},
		{ID: "ready", Platform: "youtube", Title: "Ready", AnalyzedAt: "2026-08-18T00:00:00Z", Mods: []CreatorMod{{Name: "Wakes", Timestamp: "1:23"}}},
	}}
	req := httptest.NewRequest(http.MethodGet, "/api/creators/videos", nil)
	rec := httptest.NewRecorder()
	app.handleCreatorVideos(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
	var body struct {
		Count   int            `json:"count"`
		Pending int            `json:"pending"`
		Videos  []CreatorVideo `json:"videos"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Count != 1 || body.Pending != 1 || len(body.Videos) != 1 || body.Videos[0].ID != "ready" {
		t.Fatalf("unexpected creator exposure: %#v", body)
	}
}

func TestLiveTaxonomyMergesProviderCategoriesWithoutDuplicates(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/tag/category", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{"name": "technology", "project_type": "mod", "header": "Technology"},
			{"name": "social-and-fun", "project_type": "mod", "header": "Social"},
		})
	})
	mux.HandleFunc("/minecraft/mc-mods", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `<a href="/minecraft/search?categories=magic&amp;class=mc-mods">Magic</a><a href="/minecraft/search?categories=bug-fixes&amp;class=mc-mods">Bug Fixes</a>`)
	})
	server := httptest.NewTLSServer(mux)
	defer server.Close()
	t.Setenv("MMV_MODRINTH_API_BASE", server.URL)
	t.Setenv("MMV_CURSEFORGE_WEB_BASE", server.URL)
	app := &App{httpClient: server.Client(), settings: Settings{}}
	cats, errs := app.liveTaxonomy(context.Background())
	if len(errs) != 0 {
		t.Fatalf("live taxonomy errors: %#v", errs)
	}
	technology := 0
	foundDynamic := false
	foundCurseForge := false
	for _, cat := range cats {
		if normalizeTaxonomyName(cat.Name) == normalizeTaxonomyName("Technology & Automation") || normalizeTaxonomyName(cat.Name) == "technology" {
			technology++
		}
		if cat.ID == "mr:social-and-fun" && cat.Name == "Social And Fun" {
			foundDynamic = true
		}
		if cat.ID == "cfweb:bug-fixes" && cat.Name == "Bug Fixes" {
			foundCurseForge = true
		}
	}
	if technology > 1 {
		t.Fatalf("provider taxonomy duplicated an existing technology category: %d", technology)
	}
	if !foundDynamic || !foundCurseForge {
		t.Fatalf("missing live provider categories modrinth=%v curseforge=%v in %#v", foundDynamic, foundCurseForge, cats)
	}
	if got := dynamicCategoryQuery("mr:social-and-fun"); got != "Social And Fun" {
		t.Fatalf("dynamic category query=%q", got)
	}
	if got := dynamicCategoryQuery("cfweb:bug-fixes"); got != "Bug Fixes" {
		t.Fatalf("CurseForge public dynamic category query=%q", got)
	}
}

func TestModrinthPluginProjectTypeIsSentAsNativeFacet(t *testing.T) {
	var gotFacets string
	mux := http.NewServeMux()
	mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		gotFacets = r.URL.Query().Get("facets")
		_ = json.NewEncoder(w).Encode(map[string]any{"hits": []any{}})
	})
	server := httptest.NewTLSServer(mux)
	defer server.Close()
	t.Setenv("MMV_MODRINTH_API_BASE", server.URL)
	app := &App{httpClient: server.Client(), settings: Settings{}, avatarCache: map[string]string{}}
	_, err := app.searchModrinthUnified(context.Background(), "", providerSearchOptions{
		ProjectType: "plugin", GameVersion: "1.21.1", Loader: "paper", Limit: 12,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotFacets, "project_type:plugin") {
		t.Fatalf("native Modrinth plugin facet missing from %q", gotFacets)
	}
	if !strings.Contains(gotFacets, "versions:1.21.1") || !strings.Contains(gotFacets, "categories:paper") {
		t.Fatalf("plugin compatibility facets missing from %q", gotFacets)
	}
}

func TestDynamicModrinthCategoryBecomesLiveSearchFacet(t *testing.T) {
	var gotFacets string
	mux := http.NewServeMux()
	mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		gotFacets = r.URL.Query().Get("facets")
		_ = json.NewEncoder(w).Encode(map[string]any{"hits": []any{}})
	})
	server := httptest.NewTLSServer(mux)
	defer server.Close()
	t.Setenv("MMV_MODRINTH_API_BASE", server.URL)
	app := &App{httpClient: server.Client(), settings: Settings{}, avatarCache: map[string]string{}}
	_, err := app.searchModrinthUnified(context.Background(), "", providerSearchOptions{
		Category: "mr:social-and-fun", ProjectType: "mod", GameVersion: "1.21.1", Loader: "fabric", Limit: 12,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotFacets, "categories:social-and-fun") {
		t.Fatalf("live category facet missing from %q", gotFacets)
	}
	if !strings.Contains(gotFacets, "versions:1.21.1") || !strings.Contains(gotFacets, "categories:fabric") {
		t.Fatalf("compatibility facets missing from %q", gotFacets)
	}
}

func TestPublicCurseForgeCategoryFiltersIntegratedSearch(t *testing.T) {
	var gotCategory, gotClass string
	mux := http.NewServeMux()
	mux.HandleFunc("/minecraft/search", func(w http.ResponseWriter, r *http.Request) {
		gotCategory = r.URL.Query().Get("categories")
		gotClass = r.URL.Query().Get("class")
		_, _ = io.WriteString(w, `<a href="/minecraft/mc-mods/test-mod"><img src="https://media.example/test.png" alt="Test Mod">Test Mod</a>`)
	})
	server := httptest.NewTLSServer(mux)
	defer server.Close()
	t.Setenv("MMV_CURSEFORGE_WEB_BASE", server.URL)
	app := &App{httpClient: server.Client(), settings: Settings{}}
	items, err := app.searchCurseForgeHTML(context.Background(), "minecraft", providerSearchOptions{Category: "cfweb:bug-fixes", ProjectType: "mod"})
	if err != nil {
		t.Fatal(err)
	}
	if gotCategory != "bug-fixes" || gotClass != "mc-mods" {
		t.Fatalf("CurseForge live category filter lost: category=%q class=%q", gotCategory, gotClass)
	}
	if len(items) != 1 || items[0].Title != "Test Mod" || items[0].Provider != "curseforge" {
		t.Fatalf("integrated CurseForge parser returned %#v", items)
	}
}

func TestCurseForgeContentTypeRoutesToNativeWebLanes(t *testing.T) {
	tests := []struct {
		projectType string
		wantRoot    string
		wantClass   string
		wantPath    string
		wantType    string
	}{
		{"resourcepack", "/minecraft/search", "texture-packs", "/minecraft/texture-packs/cozy-pack", "resourcepack"},
		{"shader", "/minecraft/search", "shaders", "/minecraft/shaders/cozy-shader", "shader"},
		{"world", "/minecraft/search", "worlds", "/minecraft/worlds/cozy-world", "world"},
		{"datapack", "/minecraft/search", "data-packs", "/minecraft/data-packs/cozy-datapack", "datapack"},
		{"plugin", "/minecraft/search", "bukkit-plugins", "/minecraft/bukkit-plugins/cozy-plugin", "plugin"},
		{"addon", "/minecraft-bedrock/search", "addons", "/minecraft-bedrock/addons/cozy-addon", "addon"},
		{"skin", "/minecraft-bedrock/search", "skins", "/minecraft-bedrock/skins/cozy-skin", "skin"},
	}
	for _, tc := range tests {
		t.Run(tc.projectType, func(t *testing.T) {
			var mu sync.Mutex
			seen := map[string]bool{}
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				key := r.URL.Path + "?class=" + r.URL.Query().Get("class")
				mu.Lock()
				seen[key] = true
				mu.Unlock()
				if r.URL.Path == tc.wantRoot && r.URL.Query().Get("class") == tc.wantClass {
					_, _ = io.WriteString(w, `<a href="`+tc.wantPath+`"><img src="https://media.example/card.png" alt="Cozy">Cozy</a>`)
					return
				}
				_, _ = io.WriteString(w, `<html><body>No matching cards</body></html>`)
			}))
			defer server.Close()
			t.Setenv("MMV_CURSEFORGE_WEB_BASE", server.URL)
			app := &App{httpClient: server.Client(), settings: Settings{}}
			items, err := app.searchCurseForgeHTML(context.Background(), "cozy", providerSearchOptions{ProjectType: tc.projectType})
			if err != nil {
				t.Fatal(err)
			}
			wantKey := tc.wantRoot + "?class=" + tc.wantClass
			mu.Lock()
			wasSeen := seen[wantKey]
			mu.Unlock()
			if !wasSeen {
				t.Fatalf("project type %q never queried expected lane %q; saw %#v", tc.projectType, wantKey, seen)
			}
			found := false
			for _, item := range items {
				if item.ProjectType == tc.wantType && strings.Contains(item.PageURL, strings.TrimRight(tc.wantPath, "/")) {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("expected %s result not found: %#v", tc.wantType, items)
			}
		})
	}
}

func TestPlanetMinecraftContentTypeRoutesToMatchingIndex(t *testing.T) {
	tests := []struct {
		projectType string
		indexPath   string
		projectPath string
	}{
		{"datapack", "/data-packs/", "/data-pack/cozy-data/"},
		{"resourcepack", "/texture-packs/", "/texture-pack/cozy-pack/"},
		{"world", "/projects/", "/project/cozy-world/"},
		{"skin", "/skins/", "/skin/cozy-skin/"},
	}
	for _, tc := range tests {
		t.Run(tc.projectType, func(t *testing.T) {
			var gotPath string
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				_, _ = io.WriteString(w, `<a href="`+tc.projectPath+`"><img src="/media/card.png" alt="Cozy Project">Cozy Project</a>`)
			}))
			defer server.Close()
			t.Setenv("MMV_PLANETMINECRAFT_BASE", server.URL)
			app := &App{httpClient: server.Client(), settings: Settings{}}
			items, err := app.searchPlanetMinecraft(context.Background(), "cozy", providerSearchOptions{ProjectType: tc.projectType})
			if err != nil {
				t.Fatal(err)
			}
			if gotPath != tc.indexPath {
				t.Fatalf("project type %q routed to %q, want %q", tc.projectType, gotPath, tc.indexPath)
			}
			if len(items) != 1 || items[0].ProjectType != tc.projectType {
				t.Fatalf("unexpected results: %#v", items)
			}
			if tc.projectType == "skin" && items[0].Installable {
				t.Fatal("skins should remain integrated browse items, not archive installs")
			}
		})
	}
}

func TestGitHubSearchIsContentTypeAware(t *testing.T) {
	var gotQuery string
	mux := http.NewServeMux()
	mux.HandleFunc("/search/repositories", func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query().Get("q")
		_ = json.NewEncoder(w).Encode(map[string]any{"items": []map[string]any{{
			"id": 1, "name": "cozy-data", "full_name": "maker/cozy-data", "description": "A Minecraft datapack", "html_url": "https://github.com/maker/cozy-data", "updated_at": "2026-08-19T00:00:00Z", "stargazers_count": 42, "forks_count": 3,
			"owner": map[string]any{"login": "maker", "avatar_url": "https://avatars.example/maker.png"}, "topics": []string{"minecraft", "datapack"},
		}}})
	})
	server := httptest.NewTLSServer(mux)
	defer server.Close()
	t.Setenv("MMV_GITHUB_API_BASE", server.URL)
	app := &App{httpClient: server.Client(), settings: Settings{}}
	items, err := app.searchGitHub(context.Background(), "cozy", providerSearchOptions{ProjectType: "datapack"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.ToLower(gotQuery), "minecraft datapack") {
		t.Fatalf("GitHub query was not type-aware: %q", gotQuery)
	}
	if len(items) != 1 || items[0].ProjectType != "datapack" || !items[0].Installable {
		t.Fatalf("unexpected GitHub result: %#v", items)
	}
}

func TestGitHubReleaseInstallsVerifiedZipIntoRequestedPackTarget(t *testing.T) {
	var zipBuf bytes.Buffer
	zw := zip.NewWriter(&zipBuf)
	f, err := zw.Create("pack.mcmeta")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.Write([]byte(`{"pack":{"pack_format":34,"description":"test"}}`))
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	server := httptest.NewTLSServer(mux)
	defer server.Close()
	mux.HandleFunc("/repos/maker/cozy-pack/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"tag_name": "v1", "assets": []map[string]any{{"name": "cozy-pack-1.21.zip", "browser_download_url": server.URL + "/asset.zip", "size": zipBuf.Len()}}})
	})
	mux.HandleFunc("/asset.zip", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(zipBuf.Bytes()) })
	t.Setenv("MMV_GITHUB_API_BASE", server.URL)
	root := t.TempDir()
	app := &App{httpClient: server.Client(), cfgDir: filepath.Join(root, "cfg"), settings: Settings{JavaRoot: filepath.Join(root, ".minecraft")}}
	result, err := app.installGitHubRelease(context.Background(), "maker/cozy-pack", "1.21", "fabric", "resourcepacks")
	if err != nil {
		t.Fatal(err)
	}
	path, _ := result["path"].(string)
	if !strings.Contains(filepath.Clean(path), filepath.Join(".minecraft", "resourcepacks")) {
		t.Fatalf("installed to wrong target: %#v", result)
	}
	if err := validateZipContainer(path); err != nil {
		t.Fatalf("installed asset invalid: %v", err)
	}
}

func TestCreatorDescriptionAndTranscriptNameExtraction(t *testing.T) {
	if got := timestampedDescriptionCandidate("0:32 Another Furniture"); got != "Another Furniture" {
		t.Fatalf("timestampedDescriptionCandidate=%q", got)
	}
	if got := timestampedDescriptionCandidate("0:00 Intro"); got != "" {
		t.Fatalf("expected chapter noise to be ignored, got %q", got)
	}
	if ts, ok := timestampFromLineOK("12:34 - Wakes"); !ok || ts != 754 {
		t.Fatalf("timestamp parse got %d ok=%v", ts, ok)
	}
	names := transcriptModNames("I'm using Wakes mod today, and next is Another Furniture")
	joined := strings.ToLower(strings.Join(names, " | "))
	if !strings.Contains(joined, "wakes") || !strings.Contains(joined, "another furniture") {
		t.Fatalf("transcript candidates missing expected names: %#v", names)
	}
}

func TestLocalCreatorTranscriptBootstrapsVerifiedToolchain(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("deterministic fake toolchain test currently targets Linux release asset names")
	}
	modelBytes := []byte("fake-whisper-model-for-integration-test")
	modelSHA1 := sha1.Sum(modelBytes)
	ytdlp := []byte(`#!/bin/sh
set -eu
out=""
while [ "$#" -gt 0 ]; do
  if [ "$1" = "-o" ]; then shift; out="$1"; fi
  shift || true
done
out=$(printf '%s' "$out" | sed 's/%(ext)s/wav/g')
printf 'RIFFfakewav' > "$out"
`)
	whisperScript := []byte(`#!/bin/sh
set -eu
out=""
while [ "$#" -gt 0 ]; do
  if [ "$1" = "-of" ]; then shift; out="$1"; fi
  shift || true
done
cat > "${out}.srt" <<'SRT'
1
00:00:03,000 --> 00:00:05,000
The mod called Wakes adds water splashes.

2
00:00:08,500 --> 00:00:10,000
Another Furniture mod is cozy.
SRT
`)
	var tarBuf bytes.Buffer
	gz := gzip.NewWriter(&tarBuf)
	tw := tar.NewWriter(gz)
	if err := tw.WriteHeader(&tar.Header{Name: "whisper-bin/whisper-cli", Mode: 0o755, Size: int64(len(whisperScript)), Typeflag: tar.TypeReg}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(whisperScript); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	whisperArchive := tarBuf.Bytes()
	sha256Hex := func(b []byte) string { sum := sha256.Sum256(b); return hex.EncodeToString(sum[:]) }

	mux := http.NewServeMux()
	var server *httptest.Server
	mux.HandleFunc("/repos/yt-dlp/yt-dlp/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"tag_name": "test", "assets": []map[string]any{{"name": "yt-dlp_linux", "browser_download_url": server.URL + "/assets/yt-dlp_linux", "size": len(ytdlp), "digest": "sha256:" + sha256Hex(ytdlp)}}})
	})
	mux.HandleFunc("/repos/ggml-org/whisper.cpp/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"tag_name": "test", "assets": []map[string]any{{"name": "whisper-bin-ubuntu-x64.tar.gz", "browser_download_url": server.URL + "/assets/whisper.tar.gz", "size": len(whisperArchive), "digest": "sha256:" + sha256Hex(whisperArchive)}}})
	})
	mux.HandleFunc("/assets/yt-dlp_linux", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(ytdlp) })
	mux.HandleFunc("/assets/whisper.tar.gz", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(whisperArchive) })
	mux.HandleFunc("/model", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(modelBytes) })
	server = httptest.NewTLSServer(mux)
	defer server.Close()

	oldAPI, oldModelURL, oldModelSHA := githubReleaseAPIBase, whisperModelDownloadURL, whisperModelExpectedSHA1
	githubReleaseAPIBase = server.URL
	whisperModelDownloadURL = server.URL + "/model"
	whisperModelExpectedSHA1 = hex.EncodeToString(modelSHA1[:])
	defer func() {
		githubReleaseAPIBase, whisperModelDownloadURL, whisperModelExpectedSHA1 = oldAPI, oldModelURL, oldModelSHA
	}()

	cfg := t.TempDir()
	app := &App{cfgDir: cfg, settings: Settings{}, httpClient: server.Client()}
	segments, source, err := app.localCreatorTranscript(context.Background(), "https://example.invalid/video")
	if err != nil {
		t.Fatal(err)
	}
	if source != "local Whisper.cpp ASR" || len(segments) != 2 {
		t.Fatalf("unexpected transcript result source=%q segments=%#v", source, segments)
	}
	if segments[0].StartMS != 3000 || !strings.Contains(segments[0].Text, "Wakes") {
		t.Fatalf("unexpected first segment: %#v", segments[0])
	}
	for _, path := range []string{
		filepath.Join(cfg, "creator-transcription", "tools", "yt-dlp_linux"),
		filepath.Join(cfg, "creator-transcription", "models", "ggml-large-v3-turbo-q5_0.bin"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("bootstrapped artifact missing %s: %v", path, err)
		}
	}
}

func TestDuplicateMergePreservesRichMediaAndSourceVariants(t *testing.T) {
	items := []UnifiedProject{
		{ID: "one", Provider: "modrinth", ProjectType: "mod", Slug: "cozy-furniture", Title: "Cozy Furniture", Summary: "Short summary", Author: "Maker", IconURL: "https://img.test/icon.png", Downloads: 50, Categories: []string{"decoration"}, Versions: []string{"1.21.1"}, Loaders: []string{"fabric"}, PageURL: "https://modrinth.com/mod/cozy-furniture", Score: 100},
		{ID: "two", Provider: "curseforge", ProjectType: "mod", Slug: "cozy-furniture", Title: "Cozy Furniture", Summary: "A much richer project description with cottagecore furniture details.", Author: "Maker", AuthorAvatarURL: "https://img.test/avatar.png", Gallery: []string{"https://img.test/hero.jpg"}, Downloads: 5000, Followers: 300, Categories: []string{"furniture"}, Versions: []string{"1.20.1"}, Loaders: []string{"forge"}, PageURL: "https://curseforge.com/minecraft/mc-mods/cozy-furniture", Installable: true, InstallMode: "native", Score: 80},
	}
	merged := mergeProviderDuplicates(items)
	if len(merged) != 1 {
		t.Fatalf("expected one merged project, got %#v", merged)
	}
	got := merged[0]
	if got.IconURL == "" || got.AuthorAvatarURL == "" || len(got.Gallery) != 1 {
		t.Fatalf("merged project lost rich media: %#v", got)
	}
	if !containsFold(got.Providers, "modrinth") || !containsFold(got.Providers, "curseforge") || len(got.Links) != 2 {
		t.Fatalf("merged project lost provider variants: %#v", got)
	}
	if !containsFold(got.Versions, "1.21.1") || !containsFold(got.Versions, "1.20.1") || !containsFold(got.Loaders, "fabric") || !containsFold(got.Loaders, "forge") {
		t.Fatalf("merged project lost compatibility metadata: %#v", got)
	}
	if got.Downloads != 5000 || got.Followers != 300 || !got.Installable || !strings.Contains(got.Summary, "richer") {
		t.Fatalf("merged project did not preserve strongest metadata: %#v", got)
	}
}

func TestProviderRegistryIsServerOwnedAndFederated(t *testing.T) {
	app := &App{settings: Settings{CurseForgeAPIKey: "cf", GitHubToken: "gh", BuiltByBitAPIKey: "bbb", NexusAPIKey: "nx"}}
	infos := app.providerInfos()
	if len(infos) < 28 {
		t.Fatalf("provider registry shrank to %d entries", len(infos))
	}
	want := map[string]bool{"modrinth": false, "curseforge": false, "github": false, "planetminecraft": false, "mcpedl": false, "marketplace": false, "hangar": false, "spigot": false, "bukkitdev": false, "builtbybit": false, "moddb": false, "atlauncher": false, "technic": false, "ftb": false, "nexusmods": false, "smithed": false, "polymart": false, "spongeore": false, "vanillatweaks": false, "minecraftmaps": false, "resourcepacknet": false, "texturepacks": false, "mcreator": false, "shaderpackscom": false, "shaderpacksnet": false, "minecraftshader": false, "skindex": false, "minecrafthub": false}
	for _, p := range infos {
		if _, ok := want[p.ID]; ok {
			want[p.ID] = true
		}
		if p.SearchMode == "" || p.DetailMode == "" {
			t.Fatalf("provider %s has an incomplete integration contract: %#v", p.ID, p)
		}
		if p.Credential != "" && !p.CredentialConfigured {
			t.Fatalf("configured credential not reflected for %s", p.ID)
		}
	}
	for id, found := range want {
		if !found {
			t.Errorf("provider registry missing %s", id)
		}
	}
}

func TestAdditionalIntegratedIndexesSearchInsideVault(t *testing.T) {
	t.Run("minecraft maps", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.WriteString(w, `<html><body><article><a href="/12345-cute-town"><img src="/img/town.jpg" alt="Cute Town"></a><p>Cozy world map MC 1.21.1 by Builder</p></article></body></html>`)
		})
		server := httptest.NewTLSServer(mux)
		defer server.Close()
		t.Setenv("MMV_MINECRAFTMAPS_BASE", server.URL)
		app := &App{httpClient: server.Client(), providerHealth: map[string]ProviderHealth{}, providerCache: map[string]providerCacheEntry{}}
		items, err := app.searchMinecraftMaps(context.Background(), "cute", providerSearchOptions{ProjectType: "world", Limit: 10})
		if err != nil || len(items) != 1 || items[0].Title != "Cute Town" || items[0].ProjectType != "world" {
			t.Fatalf("Minecraft Maps integration failed: items=%#v err=%v", items, err)
		}
	})

	t.Run("resourcepack net", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.WriteString(w, `<html><body><article><a href="/cozy-resource-pack/"><img src="/img/cozy.jpg" alt="Cozy Resource Pack"></a><p>Warm cottage textures for Minecraft 1.21.1</p></article></body></html>`)
		})
		server := httptest.NewTLSServer(mux)
		defer server.Close()
		t.Setenv("MMV_RESOURCEPACKNET_BASE", server.URL)
		app := &App{httpClient: server.Client(), providerHealth: map[string]ProviderHealth{}, providerCache: map[string]providerCacheEntry{}}
		items, err := app.searchResourcePackNet(context.Background(), "cozy", providerSearchOptions{ProjectType: "resourcepack", Limit: 10})
		if err != nil || len(items) != 1 || items[0].Title != "Cozy Resource Pack" || items[0].ProjectType != "resourcepack" {
			t.Fatalf("ResourcePack.net integration failed: items=%#v err=%v", items, err)
		}
	})

	t.Run("texture packs", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.WriteString(w, `<html><body><article><a href="/resourcepack/pastel/"><img src="/img/pastel.jpg" alt="Pastel Texture Pack"></a><p>Cute pastel textures 1.21.1</p></article></body></html>`)
		})
		server := httptest.NewTLSServer(mux)
		defer server.Close()
		t.Setenv("MMV_TEXTUREPACKS_BASE", server.URL)
		app := &App{httpClient: server.Client(), providerHealth: map[string]ProviderHealth{}, providerCache: map[string]providerCacheEntry{}}
		items, err := app.searchTexturePacks(context.Background(), "pastel", providerSearchOptions{ProjectType: "resourcepack", Limit: 10})
		if err != nil || len(items) != 1 || items[0].Title != "Pastel Texture Pack" || items[0].ProjectType != "resourcepack" {
			t.Fatalf("Texture-Packs.com integration failed: items=%#v err=%v", items, err)
		}
	})
}

func TestSpecialistDirectoriesSearchInsideVault(t *testing.T) {
	t.Run("mcreator", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/modifications", func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.WriteString(w, `<html><body><article><a href="/modification/122247/archers-plushies"><img src="/img/plushies.png" alt="Archer's Plushies"></a><p>Cute plush furniture mod for Minecraft 1.21.1 by Archer</p></article></body></html>`)
		})
		mux.HandleFunc("/img/plushies.png", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })
		server := httptest.NewTLSServer(mux)
		defer server.Close()
		t.Setenv("MMV_MCREATOR_BASE", server.URL)
		app := &App{httpClient: server.Client(), providerHealth: map[string]ProviderHealth{}, providerCache: map[string]providerCacheEntry{}}
		items, err := app.searchMCreator(context.Background(), "plushies", providerSearchOptions{ProjectType: "mod", Limit: 10})
		if err != nil || len(items) != 1 || items[0].Provider != "mcreator" || items[0].ProjectType != "mod" || !items[0].Installable {
			t.Fatalf("MCreator integration failed: items=%#v err=%v", items, err)
		}
		if !strings.Contains(items[0].PageURL, "/modification/122247/archers-plushies") || len(items[0].Gallery) == 0 {
			t.Fatalf("MCreator result lost project URL/media: %#v", items[0])
		}
	})

	shaderTests := []struct {
		name     string
		env      string
		path     string
		fn       func(*App, context.Context, string, providerSearchOptions) ([]UnifiedProject, error)
		provider string
	}{
		{name: "shaderpacks com", env: "MMV_SHADERPACKS_COM_BASE", path: "/browse/shaders", provider: "shaderpackscom", fn: func(a *App, c context.Context, q string, o providerSearchOptions) ([]UnifiedProject, error) {
			return a.searchShaderPacksCom(c, q, o)
		}},
		{name: "shaderpacks net", env: "MMV_SHADERPACKS_NET_BASE", path: "/", provider: "shaderpacksnet", fn: func(a *App, c context.Context, q string, o providerSearchOptions) ([]UnifiedProject, error) {
			return a.searchShaderPacksNet(c, q, o)
		}},
		{name: "minecraft shader", env: "MMV_MINECRAFTSHADER_BASE", path: "/", provider: "minecraftshader", fn: func(a *App, c context.Context, q string, o providerSearchOptions) ([]UnifiedProject, error) {
			return a.searchMinecraftShader(c, q, o)
		}},
	}
	for _, tc := range shaderTests {
		t.Run(tc.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc(tc.path, func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.WriteString(w, `<html><body><article><a href="/bsl-shaders/"><img src="/images/bsl.jpg" alt="BSL Shaders"></a><p>Beautiful cinematic shader pack for Minecraft 1.21.1</p></article></body></html>`)
			})
			server := httptest.NewTLSServer(mux)
			defer server.Close()
			t.Setenv(tc.env, server.URL)
			app := &App{httpClient: server.Client(), providerHealth: map[string]ProviderHealth{}, providerCache: map[string]providerCacheEntry{}}
			items, err := tc.fn(app, context.Background(), "bsl", providerSearchOptions{ProjectType: "shader", Limit: 10})
			if err != nil || len(items) != 1 || items[0].Provider != tc.provider || items[0].ProjectType != "shader" || !items[0].Installable {
				t.Fatalf("%s integration failed: items=%#v err=%v", tc.provider, items, err)
			}
			if len(items[0].Gallery) == 0 {
				t.Fatalf("%s result lost provider artwork: %#v", tc.provider, items[0])
			}
		})
	}

	t.Run("skindex", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/search/skin/cute/1/", func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.WriteString(w, `<html><body><div><a href="/skin/24131580/cottage-girl/"><img src="/uploads/skins/cottage-preview.png" alt="Cute Cottage Girl"></a><p>soft cottagecore outfit</p></div></body></html>`)
		})
		server := httptest.NewTLSServer(mux)
		defer server.Close()
		t.Setenv("MMV_SKINDEX_BASE", server.URL)
		app := &App{httpClient: server.Client(), providerHealth: map[string]ProviderHealth{}, providerCache: map[string]providerCacheEntry{}}
		items, err := app.searchSkindex(context.Background(), "cute", providerSearchOptions{ProjectType: "skin", Limit: 10})
		if err != nil || len(items) != 1 || items[0].ID != "24131580" || items[0].ProjectType != "skin" || items[0].InstallMode != "skin-png" {
			t.Fatalf("Skindex integration failed: items=%#v err=%v", items, err)
		}
	})
}

func TestMCreatorVerifiedDownloadChainInstallsRealJar(t *testing.T) {
	var jar bytes.Buffer
	zw := zip.NewWriter(&jar)
	mf, _ := zw.Create("META-INF/MANIFEST.MF")
	_, _ = mf.Write([]byte("Manifest-Version: 1.0\n"))
	mod, _ := zw.Create("fabric.mod.json")
	_, _ = mod.Write([]byte(`{"schemaVersion":1,"id":"archers_plushies","version":"1.0.0"}`))
	_ = zw.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/modification/122247/archers-plushies", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `<html><head><meta property="og:title" content="Archer's Plushies"><meta property="og:description" content="Cute furniture mod"></head><body><main><a href="/modificationdl/122247/file/98765">Get file</a></main></body></html>`)
	})
	mux.HandleFunc("/modificationdl/122247/file/98765", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `<html><body><a href="/files/archers-plushies.jar">Download JAR</a></body></html>`)
	})
	mux.HandleFunc("/files/archers-plushies.jar", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/java-archive")
		_, _ = w.Write(jar.Bytes())
	})
	server := httptest.NewTLSServer(mux)
	defer server.Close()
	root := t.TempDir()
	app := &App{cfgDir: filepath.Join(root, "cfg"), httpClient: server.Client(), settings: Settings{JavaRoot: filepath.Join(root, ".minecraft")}}
	result, err := app.installDetectedWebPackage(context.Background(), "mcreator", server.URL+"/modification/122247/archers-plushies", "122247", "mods")
	if err != nil {
		t.Fatal(err)
	}
	path, _ := result["path"].(string)
	if !strings.Contains(filepath.Clean(path), filepath.Join(".minecraft", "mods")) {
		t.Fatalf("MCreator package installed to wrong target: %#v", result)
	}
	if err := validateZipContainer(path); err != nil {
		t.Fatalf("installed MCreator JAR invalid: %v", err)
	}
}

func TestSkindexSkinSaveValidatesOriginalPNG(t *testing.T) {
	var img bytes.Buffer
	im := image.NewNRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			im.Set(x, y, color.NRGBA{R: uint8(x * 3), G: uint8(y * 3), B: 180, A: 255})
		}
	}
	if err := png.Encode(&img, im); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/skin/24131580/cottage-girl/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `<html><head><meta property="og:title" content="Cute Cottage Girl"></head><body><img src="/uploads/skins/cottage-girl.png" alt="original skin"></body></html>`)
	})
	mux.HandleFunc("/uploads/skins/cottage-girl.png", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(img.Bytes())
	})
	server := httptest.NewTLSServer(mux)
	defer server.Close()
	cfg := t.TempDir()
	app := &App{cfgDir: cfg, httpClient: server.Client()}
	result, err := app.installSkindexSkin(context.Background(), server.URL+"/skin/24131580/cottage-girl/", "24131580")
	if err != nil {
		t.Fatal(err)
	}
	path, _ := result["path"].(string)
	if filepath.Dir(path) != filepath.Join(cfg, "skins") {
		t.Fatalf("skin saved to wrong directory: %#v", result)
	}
	if err := validateMinecraftSkinImage(path); err != nil {
		t.Fatalf("saved skin failed validation: %v", err)
	}
}

func TestVaultTaxonomyIncludesSpecialistDiscoveryLanes(t *testing.T) {
	want := map[string]bool{"shaders": false, "cinematic": false, "physics": false, "weather": false, "skins": false, "schematics": false, "plushies": false, "kitchen": false, "dragons": false, "hud": false}
	for _, category := range universalTaxonomy {
		if _, ok := want[category.ID]; ok {
			want[category.ID] = true
			if len(category.Queries) == 0 {
				t.Errorf("taxonomy category %s has no live discovery query", category.ID)
			}
		}
	}
	for id, found := range want {
		if !found {
			t.Errorf("taxonomy missing specialist lane %s", id)
		}
	}
}

func TestMinecraftMapDetectedInstallExtractsVerifiedWorldIntoSaves(t *testing.T) {
	var zipBuf bytes.Buffer
	zw := zip.NewWriter(&zipBuf)
	level, _ := zw.Create("Cute Town/level.dat")
	_, _ = level.Write([]byte("test-level"))
	region, _ := zw.Create("Cute Town/region/r.0.0.mca")
	_, _ = region.Write([]byte("region"))
	_ = zw.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/12345-cute-town", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `<html><head><meta property="og:title" content="Cute Town"><meta property="og:description" content="A cozy Java world"></head><body><main><p>Java map by Builder.</p><a href="/files/cute-town.zip">Download Map</a></main></body></html>`)
	})
	mux.HandleFunc("/files/cute-town.zip", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		_, _ = w.Write(zipBuf.Bytes())
	})
	server := httptest.NewTLSServer(mux)
	defer server.Close()
	javaRoot := t.TempDir()
	app := &App{cfgDir: t.TempDir(), httpClient: server.Client(), settings: Settings{JavaRoot: javaRoot}}
	result, err := app.installDetectedWebPackage(context.Background(), "minecraftmaps", server.URL+"/12345-cute-town", "12345-cute-town", "worlds")
	if err != nil {
		t.Fatal(err)
	}
	path := stringFromAny(result["path"])
	if !strings.HasPrefix(path, filepath.Join(javaRoot, "saves")) {
		t.Fatalf("world installed outside saves directory: %s", path)
	}
	if info, err := os.Stat(path); err != nil || !info.IsDir() {
		t.Fatalf("world install did not create a directory: path=%s err=%v", path, err)
	}
	if _, err := os.Stat(filepath.Join(path, "level.dat")); err != nil {
		t.Fatalf("installed world missing level.dat: %v", err)
	}
}

func TestWorldArchiveExtractionRejectsTraversal(t *testing.T) {
	var zipBuf bytes.Buffer
	zw := zip.NewWriter(&zipBuf)
	bad, _ := zw.Create("../escape.txt")
	_, _ = bad.Write([]byte("escape"))
	_ = zw.Close()
	archive := filepath.Join(t.TempDir(), "bad.zip")
	if err := os.WriteFile(archive, zipBuf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	destination := t.TempDir()
	if err := extractZipPathSafe(archive, destination); err == nil {
		t.Fatal("path traversal archive was accepted")
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(destination), "escape.txt")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("path traversal wrote outside extraction root: %v", err)
	}
}

func TestVerifiedDetectedWebPackageInstallNeverTrustsDownloadButtonBlindly(t *testing.T) {
	var zipBuf bytes.Buffer
	zw := zip.NewWriter(&zipBuf)
	f, _ := zw.Create("pack.mcmeta")
	_, _ = f.Write([]byte(`{"pack":{"pack_format":48,"description":"detected"}}`))
	_ = zw.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/cozy-resource-pack/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `<html><head><meta property="og:title" content="Cozy Resource Pack"><meta property="og:description" content="Cottage textures"><meta property="og:image" content="https://cdn.example/cozy.png"></head><body><main><p>A cozy resource pack by Maker.</p><a href="/fake-download">Download mirror</a><a href="/files/cozy.zip">Download ZIP</a></main></body></html>`)
	})
	mux.HandleFunc("/fake-download", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = io.WriteString(w, `<html>not an archive</html>`)
	})
	mux.HandleFunc("/files/cozy.zip", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		_, _ = w.Write(zipBuf.Bytes())
	})
	server := httptest.NewTLSServer(mux)
	defer server.Close()
	javaRoot := t.TempDir()
	app := &App{cfgDir: t.TempDir(), httpClient: server.Client(), settings: Settings{JavaRoot: javaRoot}}
	page := server.URL + "/cozy-resource-pack/"
	d, err := app.genericWebDetails(context.Background(), "resourcepacknet", page, "cozy")
	if err != nil || !d.Installable || len(d.Links) != 2 {
		t.Fatalf("detected detail did not expose verified-install mode: %#v err=%v", d, err)
	}
	result, err := app.installDetectedWebPackage(context.Background(), "resourcepacknet", page, "cozy", "resourcepacks")
	if err != nil {
		t.Fatal(err)
	}
	path := stringFromAny(result["path"])
	if !strings.HasPrefix(path, filepath.Join(javaRoot, "resourcepacks")) {
		t.Fatalf("detected package installed outside resourcepacks: %s", path)
	}
	if err := validateZipContainer(path); err != nil {
		t.Fatalf("detected installer accepted invalid archive: %v", err)
	}
}

func TestSmithedNativeSearchDetailsAndWorldDatapackInstall(t *testing.T) {
	var serverURL string
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	f, _ := zw.Create("pack.mcmeta")
	_, _ = f.Write([]byte(`{"pack":{"pack_format":48,"description":"test"}}`))
	_ = zw.Close()
	payload := buf.Bytes()

	mux := http.NewServeMux()
	mux.HandleFunc("/v2/packs", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("search"); got != "cute" {
			t.Errorf("Smithed search=%q want cute", got)
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{{
			"id": "cute", "displayName": "Cute Pack",
			"data": map[string]any{
				"display":    map[string]any{"name": "Cute Pack", "description": "Cute furniture datapack", "icon": "https://cdn.example/icon.png", "webPage": "https://smithed.dev/packs/cute", "gallery": []string{"https://cdn.example/hero.png"}},
				"versions":   []map[string]any{{"name": "1.2.0", "supports": []string{"1.21.1"}, "downloads": map[string]any{"datapack": serverURL + "/files/cute.zip"}}},
				"categories": []string{"QoL"},
			},
			"meta": map[string]any{"owner": "maker", "stats": map[string]any{"updated": 1700000000, "score": 20, "downloads": map[string]any{"total": 1000}}},
		}})
	})
	mux.HandleFunc("/v2/packs/cute", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": "cute", "display": map[string]any{"name": "Cute Pack", "description": "Cute furniture datapack", "icon": "https://cdn.example/icon.png", "webPage": "https://smithed.dev/packs/cute", "gallery": []string{"https://cdn.example/hero.png"}},
			"versions":   []map[string]any{{"name": "1.2.0", "supports": []string{"1.21.1"}, "downloads": map[string]any{"datapack": serverURL + "/files/cute.zip"}, "dependencies": []map[string]any{}}},
			"categories": []string{"QoL"},
		})
	})
	mux.HandleFunc("/v2/packs/cute/meta", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"owner": "maker", "contributors": []string{"helper"}, "stats": map[string]any{"updated": 1700000000, "added": 1600000000, "score": 20, "downloads": map[string]any{"total": 1000}}})
	})
	mux.HandleFunc("/v2/packs/cute/versions/latest", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"name": "1.2.0", "supports": []string{"1.21.1"}, "downloads": map[string]any{"datapack": serverURL + "/files/cute.zip"}, "dependencies": []map[string]any{}})
	})
	mux.HandleFunc("/files/cute.zip", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(payload) })
	server := httptest.NewTLSServer(mux)
	defer server.Close()
	serverURL = server.URL
	t.Setenv("MMV_SMITHED_API_BASE", server.URL+"/v2")
	world := t.TempDir()
	app := &App{cfgDir: t.TempDir(), httpClient: server.Client(), settings: Settings{WorldRoot: world}, providerHealth: map[string]ProviderHealth{}}

	resp := app.searchProviders(context.Background(), providerSearchOptions{Query: "cute", GameVersion: "1.21.1", ProjectType: "datapack", Limit: 10, Sources: []string{"smithed"}})
	if len(resp.Results) != 1 || resp.Results[0].Title != "Cute Pack" || resp.Results[0].Author != "maker" || !resp.Results[0].Installable {
		t.Fatalf("Smithed search incomplete: %#v errors=%#v", resp.Results, resp.Errors)
	}
	d, err := app.fetchProjectDetails(context.Background(), "smithed", "cute", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Authors) != 2 || len(d.Versions) != 1 || d.Versions[0].Files[0].URL == "" {
		t.Fatalf("Smithed details incomplete: %#v", d)
	}
	result, err := app.installSmithed(context.Background(), "cute", "1.21.1", "datapacks")
	if err != nil {
		t.Fatal(err)
	}
	path := stringFromAny(result["path"])
	if !strings.HasPrefix(path, filepath.Join(world, "datapacks")) {
		t.Fatalf("Smithed datapack installed outside active world: %s", path)
	}
	if err := validateZipContainer(path); err != nil {
		t.Fatalf("installed Smithed zip invalid: %v", err)
	}
}

func TestSpongeOreNativeSearchDetailsAndMD5VerifiedInstall(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	f, _ := zw.Create("META-INF/MANIFEST.MF")
	_, _ = f.Write([]byte("Manifest-Version: 1.0\n"))
	_ = zw.Close()
	payload := buf.Bytes()
	sum := md5.Sum(payload)
	md5hex := hex.EncodeToString(sum[:])

	mux := http.NewServeMux()
	project := map[string]any{"pluginId": "chairs", "name": "Chairs", "owner": "Maker", "description": "Sit anywhere", "href": "/Maker/Chairs", "createdAt": "2025-01-01T00:00:00Z", "downloads": 50, "stars": 7, "category": map[string]any{"title": "Gameplay"}, "recommended": map[string]any{"name": "2.0"}}
	mux.HandleFunc("/api/v1/projects", func(w http.ResponseWriter, r *http.Request) { _ = json.NewEncoder(w).Encode([]map[string]any{project}) })
	mux.HandleFunc("/api/v1/projects/chairs", func(w http.ResponseWriter, r *http.Request) { _ = json.NewEncoder(w).Encode(project) })
	mux.HandleFunc("/api/v1/projects/chairs/versions", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]map[string]any{{"id": 2, "name": "2.0", "createdAt": "2026-01-01T00:00:00Z", "fileSize": len(payload), "md5": md5hex, "downloads": 20, "tags": []map[string]any{{"name": "Sponge", "data": "12.0"}}}})
	})
	mux.HandleFunc("/api/v1/projects/chairs/versions/2.0/download", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(payload) })
	server := httptest.NewTLSServer(mux)
	defer server.Close()
	t.Setenv("MMV_SPONGE_ORE_BASE", server.URL+"/api/v1")
	serverRoot := t.TempDir()
	app := &App{cfgDir: t.TempDir(), httpClient: server.Client(), settings: Settings{ServerRoot: serverRoot}, providerHealth: map[string]ProviderHealth{}}
	resp := app.searchProviders(context.Background(), providerSearchOptions{Query: "chairs", ProjectType: "plugin", Limit: 10, Sources: []string{"spongeore"}})
	if len(resp.Results) != 1 || !resp.Results[0].Installable {
		t.Fatalf("Sponge Ore search incomplete: %#v errors=%#v", resp.Results, resp.Errors)
	}
	d, err := app.fetchProjectDetails(context.Background(), "spongeore", "chairs", "")
	if err != nil || len(d.Versions) != 1 || d.Versions[0].Files[0].Hashes["md5"] != md5hex {
		t.Fatalf("Sponge Ore details missing verified version: %#v err=%v", d, err)
	}
	result, err := app.installSpongeOre(context.Background(), "chairs", "plugins")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(stringFromAny(result["path"]), filepath.Join(serverRoot, "plugins")) {
		t.Fatalf("Sponge plugin installed outside plugin dir: %#v", result)
	}
}

func TestHangarIntegratedSearchDetailsAndVerifiedFileMetadata(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/projects", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("query"); got != "spark" {
			t.Errorf("Hangar query=%q want spark", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"result": []map[string]any{{
			"id": 42, "name": "Spark", "description": "Profiler plugin", "category": "ADMIN_TOOLS", "lastUpdated": "2026-08-01T00:00:00Z", "avatarUrl": "https://cdn.example/spark.png",
			"namespace": map[string]any{"owner": "lucko", "slug": "spark"}, "stats": map[string]any{"downloads": 1234, "stars": 50, "watchers": 5},
		}}})
	})
	mux.HandleFunc("/api/v1/projects/spark", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": 42, "name": "Spark", "description": "Profiler plugin", "mainPageContent": "# Spark\nFast profiler", "category": "ADMIN_TOOLS", "lastUpdated": "2026-08-01T00:00:00Z", "publishedAt": "2025-01-01T00:00:00Z", "avatarUrl": "https://cdn.example/spark.png",
			"namespace": map[string]any{"owner": "lucko", "slug": "spark"}, "stats": map[string]any{"downloads": 1234, "stars": 50, "watchers": 5},
			"supportedPlatforms": map[string]any{"PAPER": []string{"1.21.1", "1.21.4"}}, "memberNames": []string{"lucko"},
		})
	})
	mux.HandleFunc("/api/v1/projects/spark/versions", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"result": []map[string]any{{
			"id": 99, "name": "1.10.100", "createdAt": "2026-08-02T00:00:00Z",
			"platformDependencies": map[string]any{"PAPER": []string{"1.21.1"}},
			"downloads":            map[string]any{"PAPER": map[string]any{"downloadUrl": "https://cdn.example/spark.jar", "fileInfo": map[string]any{"name": "spark.jar", "sizeBytes": 321, "sha256Hash": strings.Repeat("a", 64)}}},
		}}})
	})
	server := httptest.NewTLSServer(mux)
	defer server.Close()
	t.Setenv("MMV_HANGAR_API_BASE", server.URL+"/api/v1")
	app := &App{httpClient: server.Client(), providerHealth: map[string]ProviderHealth{}}

	resp := app.searchProviders(context.Background(), providerSearchOptions{Query: "spark", ProjectType: "plugin", Limit: 10, Sources: []string{"hangar"}})
	if len(resp.Results) != 1 || resp.Results[0].Title != "Spark" || resp.Results[0].IconURL == "" || !resp.Results[0].Installable {
		t.Fatalf("Hangar integrated search lost metadata: %#v errors=%#v", resp.Results, resp.Errors)
	}
	d, err := app.fetchProjectDetails(context.Background(), "hangar", "lucko/spark", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Authors) != 1 || d.Authors[0].Name != "lucko" || len(d.Versions) != 1 || len(d.Versions[0].Files) != 1 {
		t.Fatalf("Hangar rich detail incomplete: %#v", d)
	}
	f := d.Versions[0].Files[0]
	if f.Name != "spark.jar" || f.Size != 321 || f.Hashes["sha256"] != strings.Repeat("a", 64) {
		t.Fatalf("Hangar verified file metadata missing: %#v", f)
	}
}

func TestSpigotIntegratedSearchAndExternalInstallTruthfulness(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v2/search/resources/chairs", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]map[string]any{{"id": 7, "name": "Chairs", "tag": "Sit on chairs", "downloads": 100, "likes": 10, "external": false, "author": map[string]any{"name": "Builder"}, "icon": map[string]any{"url": "/icons/chairs.png"}}})
	})
	mux.HandleFunc("/v2/resources/7", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"id": 7, "name": "Chairs", "tag": "Sit on chairs", "downloads": 100, "likes": 10, "external": true, "author": map[string]any{"name": "Builder"}, "icon": map[string]any{"url": "/icons/chairs.png"}})
	})
	mux.HandleFunc("/v2/resources/7/versions", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]map[string]any{{"id": 2, "name": "2.0", "releaseDate": 1700000000}})
	})
	server := httptest.NewTLSServer(mux)
	defer server.Close()
	t.Setenv("MMV_SPIGET_API_BASE", server.URL+"/v2")
	app := &App{httpClient: server.Client(), providerHealth: map[string]ProviderHealth{}}
	resp := app.searchProviders(context.Background(), providerSearchOptions{Query: "chairs", ProjectType: "plugin", Limit: 10, Sources: []string{"spigot"}})
	if len(resp.Results) != 1 || resp.Results[0].Author != "Builder" || resp.Results[0].IconURL == "" || !resp.Results[0].Installable {
		t.Fatalf("Spigot search integration incomplete: %#v", resp.Results)
	}
	d, err := app.fetchProjectDetails(context.Background(), "spigot", "7", "")
	if err != nil {
		t.Fatal(err)
	}
	if d.Installable {
		t.Fatal("external Spigot resource was incorrectly advertised as one-click installable")
	}
	if len(d.Versions) != 1 || d.Versions[0].Version != "2.0" {
		t.Fatalf("Spigot versions missing: %#v", d.Versions)
	}
}

func TestBuiltByBitSearchNeedsOAuthOnlyForInstall(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v2/resources/discover/resources", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "api-token" {
			t.Errorf("BuiltByBit API Authorization=%q", r.Header.Get("Authorization"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"resources": []map[string]any{{
			"resource_id": 55, "title": "Furniture Plus", "summary": "Furniture plugin", "downloads": 500, "purchases": 30, "cover_image_url": "https://cdn.example/furniture.png", "url": "/resources/55/",
			"Creator": map[string]any{"username": "Maker", "avatar_url": "https://cdn.example/maker.png"}, "Category": map[string]any{"title": "Plugins"}, "LatestVersion": map[string]any{"version_id": 9, "version_string": "1.2.0"},
		}}}})
	})
	server := httptest.NewTLSServer(mux)
	defer server.Close()
	t.Setenv("MMV_BUILTBYBIT_API_BASE", server.URL)
	app := &App{httpClient: server.Client(), settings: Settings{BuiltByBitAPIKey: "api-token"}}
	items, err := app.searchBuiltByBit(context.Background(), "furniture", providerSearchOptions{Limit: 10})
	if err != nil || len(items) != 1 {
		t.Fatalf("BuiltByBit search failed items=%#v err=%v", items, err)
	}
	if items[0].Installable {
		t.Fatal("BuiltByBit search advertised licensed install without OAuth token")
	}
	d, err := app.builtByBitDetails(context.Background(), "55")
	if err != nil {
		t.Fatal(err)
	}
	if d.IconURL == "" || len(d.Authors) != 1 || d.Authors[0].AvatarURL == "" || d.Installable {
		t.Fatalf("BuiltByBit rich details/credential gating wrong: %#v", d)
	}
	app.settings.BuiltByBitOAuthToken = "oauth-token"
	items, err = app.searchBuiltByBit(context.Background(), "furniture", providerSearchOptions{Limit: 10})
	if err != nil || !items[0].Installable {
		t.Fatalf("OAuth-ready BuiltByBit resource not installable: %#v err=%v", items, err)
	}
}

func TestV070FrontendUsesInAppProviderBrowserControls(t *testing.T) {
	indexBytes, err := embeddedFiles.ReadFile("web/index.html")
	if err != nil {
		t.Fatal(err)
	}
	jsBytes, err := embeddedFiles.ReadFile("web/app.js")
	if err != nil {
		t.Fatal(err)
	}
	index, js := string(indexBytes), string(jsBytes)
	for _, marker := range []string{`id="providerPresets"`, `id="providerStrip"`, `id="modSourceFacets"`, `id="sourceSettings"`, `id="serverRoot"`, `id="worldRoot"`, `id="builtByBitOAuthToken"`, `<option value="plugin">Server plugins</option>`, `<option value="skin">Skins</option>`} {
		if !strings.Contains(index, marker) {
			t.Fatalf("v0.7 index missing %q", marker)
		}
	}
	for _, marker := range []string{"/api/providers/detail", "Search ${esc(p.name)} inside Vault", "Every site", "Shaders", "Skins", "Community", "Plugins", "Modpacks", "Bedrock", "installActionLabel", "Live source coverage"} {
		if !strings.Contains(js, marker) {
			t.Fatalf("v0.7 app.js missing integrated browser marker %q", marker)
		}
	}
	if strings.Contains(js, "Provider site") {
		t.Fatal("primary Sources UI regressed to an external provider-site button")
	}
	if strings.Contains(js, "const PROVIDER_NAMES={modrinth") {
		t.Fatal("provider list regressed to a hard-coded client-side catalog")
	}
	if !strings.Contains(js, "for(const p of providerCatalog)") || !strings.Contains(js, "new URL(p.homeUrl||'')") {
		t.Fatal("curated catalog provider resolution must use the live server-owned provider registry")
	}
	if strings.Contains(js, "u.includes('modrinth.com')") || strings.Contains(js, "u.includes('curseforge.com')") {
		t.Fatal("curated catalog provider resolution regressed to a hard-coded provider/domain list")
	}
}

func TestDetectedWebInstallFollowsBoundedDownloadChain(t *testing.T) {
	var zipBuf bytes.Buffer
	zw := zip.NewWriter(&zipBuf)
	f, err := zw.Create("pack.mcmeta")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.Write([]byte(`{"pack":{"pack_format":34,"description":"nested download test"}}`))
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	server := httptest.NewTLSServer(mux)
	defer server.Close()
	mux.HandleFunc("/minecraft/texture-packs/cozy-pack", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `<html><head><meta property="og:title" content="Cozy Pack"></head><body><a href="/download">Download</a></body></html>`)
	})
	mux.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `<html><body><a href="/cdn/cozy-pack.zip">Download file</a></body></html>`)
	})
	mux.HandleFunc("/cdn/cozy-pack.zip", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		_, _ = w.Write(zipBuf.Bytes())
	})

	root := t.TempDir()
	app := &App{
		httpClient:  server.Client(),
		cfgDir:      filepath.Join(root, "cfg"),
		settings:    Settings{JavaRoot: filepath.Join(root, ".minecraft")},
		avatarCache: map[string]string{},
	}
	result, err := app.installDetectedWebPackage(context.Background(), "curseforge", server.URL+"/minecraft/texture-packs/cozy-pack", "cozy-pack", "resourcepacks")
	if err != nil {
		t.Fatal(err)
	}
	path, _ := result["path"].(string)
	if !strings.Contains(filepath.Clean(path), filepath.Join(".minecraft", "resourcepacks")) {
		t.Fatalf("installed to wrong target: %#v", result)
	}
	if err := validateZipContainer(path); err != nil {
		t.Fatalf("installed archive invalid: %v", err)
	}
	if got, _ := result["sourceCandidate"].(string); !strings.Contains(got, "link") {
		t.Fatalf("expected nested download-chain evidence, got %#v", result)
	}
}

func TestCurseForgeProjectTypeInferenceIncludesBedrockAndJavaLanes(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"https://www.curseforge.com/minecraft-bedrock/addons/cozy", "addon"},
		{"https://www.curseforge.com/minecraft-bedrock/scripts/cozy", "addon"},
		{"https://www.curseforge.com/minecraft-bedrock/maps/cozy", "world"},
		{"https://www.curseforge.com/minecraft-bedrock/texture-packs/cozy", "resourcepack"},
		{"https://www.curseforge.com/minecraft-bedrock/skins/cozy", "skin"},
		{"https://www.curseforge.com/minecraft/mc-mods/cozy", "mod"},
		{"https://www.curseforge.com/minecraft/modpacks/cozy", "modpack"},
		{"https://www.curseforge.com/minecraft/shaders/cozy", "shader"},
		{"https://www.curseforge.com/minecraft/data-packs/cozy", "datapack"},
		{"https://www.curseforge.com/minecraft/bukkit-plugins/cozy", "plugin"},
		{"https://www.curseforge.com/minecraft/mc-addons/cozy", "addon"},
		{"https://www.curseforge.com/minecraft/customization/cozy", "tool"},
		{"https://www.curseforge.com/minecraft/worlds/cozy", "world"},
	}
	for _, tc := range tests {
		if got := inferProjectTypeFromPageURL("curseforge", tc.url); got != tc.want {
			t.Errorf("inferProjectTypeFromPageURL(%q)=%q, want %q", tc.url, got, tc.want)
		}
	}
}

func TestFederatedSearchPaginationRanksStableProviderWindowBeforeSlicing(t *testing.T) {
	var mu sync.Mutex
	searchOffsets := []string{}
	searchLimits := []string{}
	mux := http.NewServeMux()
	mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		searchOffsets = append(searchOffsets, r.URL.Query().Get("offset"))
		searchLimits = append(searchLimits, r.URL.Query().Get("limit"))
		mu.Unlock()
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit <= 0 {
			limit = 40
		}
		hits := make([]map[string]any, 0, limit)
		for i := offset; i < offset+limit && i < 75; i++ {
			hits = append(hits, map[string]any{
				"project_id": fmt.Sprintf("p-%03d", i), "project_type": "mod", "slug": fmt.Sprintf("cozy-%03d", i),
				"author": "maker", "title": fmt.Sprintf("Cozy Project %03d", i), "description": "cozy minecraft mod",
				"categories": []string{"fabric"}, "versions": []string{"1.21.1"}, "downloads": 100, "follows": 10,
				"icon_url": "", "date_modified": "2026-08-19T00:00:00Z",
			})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"hits": hits})
	})
	mux.HandleFunc("/project/", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]any{})
	})
	server := httptest.NewTLSServer(mux)
	defer server.Close()
	t.Setenv("MMV_MODRINTH_API_BASE", server.URL)
	app := &App{httpClient: server.Client(), providerCache: map[string]providerCacheEntry{}, providerHealth: map[string]ProviderHealth{}, avatarCache: map[string]string{}}

	page := app.searchProviders(context.Background(), providerSearchOptions{
		Query: "cozy", ProjectType: "mod", GameVersion: "1.21.1", Loader: "fabric", Sort: "relevance",
		Sources: []string{"modrinth"}, Limit: 10, Offset: 10,
	})
	if len(page.Results) != 10 {
		t.Fatalf("expected 10 results, got %d: %#v", len(page.Results), page)
	}
	if page.Results[0].Title != "Cozy Project 010" || page.Results[9].Title != "Cozy Project 019" {
		t.Fatalf("federated page skipped or reordered the provider window: first=%q last=%q", page.Results[0].Title, page.Results[9].Title)
	}
	if !page.HasMore || page.NextOffset != 20 {
		t.Fatalf("expected more results at offset 20, got hasMore=%v next=%d", page.HasMore, page.NextOffset)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(searchOffsets) == 0 || searchOffsets[0] != "0" {
		t.Fatalf("provider window did not start at zero: offsets=%#v", searchOffsets)
	}
	if searchLimits[0] != "40" {
		t.Fatalf("expected bounded Modrinth page window of 40, got limits=%#v", searchLimits)
	}
}

func TestMinecraftHubCuratedSourceSearchAndDetails(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/mods", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") != "" {
			_, _ = io.WriteString(w, `<html><body><p>No more results</p></body></html>`)
			return
		}
		_, _ = io.WriteString(w, `<html><body>
			<article class="resource-card">
				<a href="/mods/create"><img src="/media/create.png" alt="Create"><h3>Create</h3></a>
				<p>Create adds rotational machines, trains and automation for Fabric, Forge and NeoForge on Minecraft 1.21.1.</p>
				<span>by simibubi</span>
			</article>
		</body></html>`)
	})
	mux.HandleFunc("/mods/create", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `<html><head>
			<meta property="og:title" content="Create">
			<meta property="og:description" content="Rotational engineering, trains, factories and contraptions.">
			<meta property="og:image" content="/media/create-hero.jpg">
		</head><body><main>
			<h1>Create</h1><p>By simibubi via Modrinth.</p>
			<section><h2>Versions</h2><p>v 1.20.1 / 1.21.1</p></section>
			<section><h2>Compatibility</h2><p>Fabric / Forge / NeoForge</p></section>
			<img src="/media/create-1.jpg" alt="Create factory">
			<p>Create is a large engineering mod with trains, moving contraptions and automation systems.</p>
			<a href="https://modrinth.com/mod/create">Visit modrinth.com</a>
		</main></body></html>`)
	})
	server := httptest.NewTLSServer(mux)
	defer server.Close()
	t.Setenv("MMV_MINECRAFTHUB_BASE", server.URL)
	app := &App{httpClient: server.Client(), providerHealth: map[string]ProviderHealth{}, providerCache: map[string]providerCacheEntry{}}

	items, err := app.searchMinecraftHub(context.Background(), "create trains", providerSearchOptions{ProjectType: "mod", Limit: 20})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one curated project, got %#v", items)
	}
	got := items[0]
	if got.Provider != "minecrafthub" || got.ProjectType != "mod" || got.Installable || got.InstallMode != "canonical-provider-resolution" {
		t.Fatalf("unexpected MinecraftHub project contract: %#v", got)
	}
	if got.IconURL == "" || !containsFold(got.Loaders, "fabric") || !containsFold(got.Loaders, "neoforge") || !containsFold(got.Versions, "1.21.1") {
		t.Fatalf("curated search lost media/compatibility metadata: %#v", got)
	}

	detail, err := app.minecraftHubDetails(context.Background(), "create", server.URL+"/mods/create")
	if err != nil {
		t.Fatal(err)
	}
	if detail.Title != "Create" || detail.ProjectType != "mod" || !detail.Installable || len(detail.Authors) == 0 || detail.Authors[0].Name != "simibubi" {
		t.Fatalf("unexpected curated detail: %#v", detail)
	}
	if !containsFold(detail.GameVersions, "1.20.1") || !containsFold(detail.GameVersions, "1.21.1") || !containsFold(detail.Loaders, "forge") || !containsFold(detail.Loaders, "fabric") {
		t.Fatalf("detail compatibility not extracted: %#v", detail)
	}
	if detail.Links["Original source"] != "https://modrinth.com/mod/create" {
		t.Fatalf("original provider source was not resolved: %#v", detail.Links)
	}
}

func TestMinecraftHubDelegatesInstallToOriginalProvider(t *testing.T) {
	jar := fabricJarBytes(t, "create", "Create", "1.0.0")
	jarSum := sha512.Sum512(jar)
	jarSHA512 := hex.EncodeToString(jarSum[:])
	mux := http.NewServeMux()
	mux.HandleFunc("/hub/mods/create", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `<html><body><h1>Create</h1><p>By simibubi via Modrinth.</p><a href="https://modrinth.com/mod/create">Visit modrinth.com</a></body></html>`)
	})
	mux.HandleFunc("/v2/project/create", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(ModrinthProject{ID: "create-id", Slug: "create", Title: "Create", ProjectType: "mod"})
	})
	mux.HandleFunc("/v2/project/create/version", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]ModrinthVersion{{
			ID: "create-v1", ProjectID: "create-id", VersionNumber: "1.0.0", DatePublished: "2026-08-19T00:00:00Z",
			GameVersions: []string{"1.21.1"}, Loaders: []string{"fabric"},
			Files: []ModrinthVersionFile{{URL: serverURLForRequest(r) + "/files/create.jar", Filename: "create.jar", Primary: true, Size: int64(len(jar)), Hashes: map[string]string{"sha512": jarSHA512}}},
		}})
	})
	mux.HandleFunc("/files/create.jar", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(jar) })
	server := httptest.NewTLSServer(mux)
	defer server.Close()
	t.Setenv("MMV_MODRINTH_API_BASE", server.URL+"/v2")

	root := t.TempDir()
	app := &App{cfgDir: t.TempDir(), settings: Settings{JavaRoot: root, GameVersion: "1.21.1", Loader: "fabric"}, httpClient: server.Client()}
	result, err := app.installMinecraftHubResolved(context.Background(), server.URL+"/hub/mods/create", "create", "1.21.1", "fabric", "mods")
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("delegated install returned nil result")
	}
	if _, err := os.Stat(filepath.Join(root, "mods", "create.jar")); err != nil {
		t.Fatalf("canonical provider delegation did not install the verified Modrinth JAR: %v", err)
	}
}

func serverURLForRequest(r *http.Request) string {
	scheme := "https"
	if r.TLS == nil {
		scheme = "http"
	}
	return scheme + "://" + r.Host
}

func TestIntegratedProviderURLRecognition(t *testing.T) {
	cases := []struct {
		raw      string
		provider string
		id       string
	}{
		{"https://modrinth.com/shader/kappa-shader", "modrinth", "kappa-shader"},
		{"https://www.curseforge.com/minecraft/mc-mods/create", "curseforge", "create"},
		{"https://github.com/FabricMC/fabric", "github", "FabricMC/fabric"},
		{"https://www.spigotmc.org/resources/luckperms.28140/", "spigot", "28140"},
		{"https://www.minecraftskins.com/skin/12345/cute-frog/", "skindex", "12345"},
	}
	for _, tc := range cases {
		provider, id, canonical, ok := integratedProviderFromURL(tc.raw)
		if !ok || provider != tc.provider || id != tc.id || canonical == "" {
			t.Errorf("integratedProviderFromURL(%q)=(%q,%q,%q,%v), want provider=%q id=%q", tc.raw, provider, id, canonical, ok, tc.provider, tc.id)
		}
	}
}

func TestMinecraftHubFrontendUsesInAppSourceVariantsAndCuratedPreset(t *testing.T) {
	js, err := os.ReadFile(filepath.Join("web", "app.js"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(js)
	for _, marker := range []string{"Curated discovery", "data-source-variant", "source-variant-button"} {
		if !strings.Contains(text, marker) {
			t.Fatalf("frontend missing integrated curated-browser marker %q", marker)
		}
	}
}

func TestCreatorArchiveFrontendContract(t *testing.T) {
	html, err := os.ReadFile(filepath.Join("web", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	js, err := os.ReadFile(filepath.Join("web", "app.js"))
	if err != nil {
		t.Fatal(err)
	}
	combined := string(html) + "\n" + string(js)
	for _, marker := range []string{
		"Every mod recommended by the creators you follow",
		"Sync all followed",
		"Follow + archive",
		"creatorChannelInput",
		"creatorSuggestionGrid",
		"addCreatorTopPicks",
		"data-pause-channel",
		"data-remove-channel",
		"creatorArchiveQuery",
		"creatorArchiveChannel",
		"creatorArchiveKind",
		"creatorArchiveSort",
		"Latest recommended first",
		"Oldest recommended first",
		"Newest videos first",
		"Oldest videos first",
		"Videos with most mods",
		"creatorTranscriptModel",
		"creatorArchiveConcurrency",
		"/api/creators/channels",
		"/api/creators/transcript",
		"configureCreatorSort",
	} {
		if !strings.Contains(combined, marker) {
			t.Fatalf("creator archive UI missing contract marker %q", marker)
		}
	}
	if strings.Contains(string(js), "if(document.querySelector('[data-view=creators]')?.classList.contains('active'))loadCreatorChannels();loadCreatorArchive()") {
		t.Fatal("creator archive periodic refresh escaped active-view guard")
	}
}
