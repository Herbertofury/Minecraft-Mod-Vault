package main

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func tinyPNGBytes(t *testing.T) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.NRGBA{R: 80, G: 190, B: 120, A: 255})
		}
	}
	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

func zipBytes(t *testing.T, files map[string][]byte) []byte {
	t.Helper()
	var out bytes.Buffer
	zw := zip.NewWriter(&out)
	for name, data := range files {
		entry, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

func forgeJarBytes(t *testing.T, modID, name, version, author string) []byte {
	t.Helper()
	manifest := "Manifest-Version: 1.0\nImplementation-Version: " + version + "\n\n"
	modsTOML := `modLoader="javafml"
loaderVersion="[47,)"
license="All Rights Reserved"
[[mods]]
modId="` + modID + `"
version="` + version + `"
displayName="` + name + `"
logoFile="assets/` + modID + `/icon.png"
authors="` + author + `"
displayURL="https://www.curseforge.com/minecraft/mc-mods/` + strings.ReplaceAll(strings.ToLower(name), " ", "-") + `"
description='''A dimension expansion used to prove cross-store identity and artwork.''' 
[[dependencies.` + modID + `]]
modId="minecraft"
mandatory=true
versionRange="[1.20.1,1.21)"
ordering="NONE"
side="BOTH"
`
	return zipBytes(t, map[string][]byte{
		"META-INF/MANIFEST.MF":                 []byte(manifest),
		"META-INF/mods.toml":                   []byte(modsTOML),
		"assets/" + modID + "/icon.png":        tinyPNGBytes(t),
		"assets/" + modID + "/lang/en_us.json": []byte(`{"item.example":"Example"}`),
	})
}

func TestOmniManagerExtractsForgeNameArtworkAuthorAndVersion(t *testing.T) {
	root := t.TempDir()
	mods := filepath.Join(root, "mods")
	if err := os.MkdirAll(mods, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(mods, "cataclysm_dimension-forge1.20.1-1.5.7.jar")
	if err := os.WriteFile(path, forgeJarBytes(t, "cataclysm_dimension", "Cataclysm Dimensions", "1.5.7", "P1nero"), 0o644); err != nil {
		t.Fatal(err)
	}
	app := &App{cfgDir: t.TempDir(), settings: Settings{JavaRoot: root, ActiveProfile: "Test", GameVersion: "1.20.1", Loader: "forge"}}
	item, warnings, err := app.inspectJavaLibraryPath(path, "mod", "java", "Test")
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %#v", warnings)
	}
	if item.Name != "Cataclysm Dimensions" || item.InstalledVersion != "1.5.7" || item.ModID != "cataclysm_dimension" {
		t.Fatalf("embedded Forge identity was not preserved: %#v", item)
	}
	if len(item.Authors) != 1 || item.Authors[0] != "P1nero" {
		t.Fatalf("Forge author missing: %#v", item.Authors)
	}
	if !containsFold(item.Loaders, "forge") || item.LocalArtURL == "" || item.ArtOrigin != "embedded-local" {
		t.Fatalf("loader/art metadata incomplete: loaders=%#v localArt=%q origin=%q", item.Loaders, item.LocalArtURL, item.ArtOrigin)
	}
	if len(item.GameVersions) != 1 || !strings.Contains(item.GameVersions[0], "1.20.1") {
		t.Fatalf("Minecraft compatibility range missing: %#v", item.GameVersions)
	}
}

func TestOmniManagerCurseForgeFingerprintRestoresCrossStoreMetadataAndDetectsSameLabelNewFile(t *testing.T) {
	root := t.TempDir()
	mods := filepath.Join(root, "mods")
	if err := os.MkdirAll(mods, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(mods, "cataclysm_dimension-forge1.20.1-1.5.7.jar")
	if err := os.WriteFile(path, forgeJarBytes(t, "cataclysm_dimension", "Cataclysm Dimensions", "1.5.7", "P1nero"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := t.TempDir()
	app := &App{cfgDir: cfg, settings: Settings{JavaRoot: root, ActiveProfile: "Test", GameVersion: "1.20.1", Loader: "forge", CurseForgeAPIKey: "test-key"}}
	item, _, err := app.inspectJavaLibraryPath(path, "mod", "java", "Test")
	if err != nil {
		t.Fatal(err)
	}
	fingerprint := item.Hashes.CurseFingerprint
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/fingerprints", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "test-key" {
			t.Fatalf("missing CurseForge API key")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"exactMatches": []map[string]any{{
			"id":   fingerprint,
			"file": map[string]any{"id": 1001, "modId": 777, "fileName": filepath.Base(path), "displayName": "1.5.7"},
		}}}})
	})
	mux.HandleFunc("/v1/mods", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{
			"id": 777, "name": "Cataclysm Dimensions", "slug": "cataclysm-dimensions", "summary": "The real provider description.",
			"logo":    map[string]any{"thumbnailUrl": "https://cdn.example/cataclysm.png"},
			"authors": []map[string]any{{"name": "P1nero", "url": "https://example.invalid/author"}},
			"links":   map[string]any{"websiteUrl": "https://www.curseforge.com/minecraft/mc-mods/cataclysm-dimensions"},
		}}})
	})
	mux.HandleFunc("/v1/mods/777/files", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("gameVersion") != "1.20.1" || r.URL.Query().Get("modLoaderType") != "1" {
			t.Fatalf("missing compatibility filters: %s", r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{
			"id": 1002, "fileName": "cataclysm_dimension-forge1.20.1-1.5.7-r2.jar", "displayName": "1.5.7",
			"downloadUrl": "https://cdn.example/cataclysm-r2.jar", "fileLength": 456, "gameVersions": []string{"1.20.1", "Forge"}, "dependencies": []any{},
		}}})
	})
	server := httptest.NewTLSServer(mux)
	defer server.Close()
	t.Setenv("MMV_CURSEFORGE_API_BASE", server.URL)
	app.httpClient = server.Client()
	items := []LibraryItem{item}
	if err := app.enrichLibraryCurseForge(context.Background(), items); err != nil {
		t.Fatal(err)
	}
	got := items[0]
	if got.Name != "Cataclysm Dimensions" || got.RemoteArtURL != "https://cdn.example/cataclysm.png" || !containsFold(got.Authors, "P1nero") {
		t.Fatalf("cross-store metadata was not restored: %#v", got)
	}
	if got.UpdateStatus != "update" || got.SafeUpdate == nil || got.SafeUpdate.ID != "curseforge:1002" {
		t.Fatalf("new compatible file with the same display version was not detected: %#v", got)
	}
	if len(got.Sources) != 1 || !got.Sources[0].Exact || got.Sources[0].Provider != "curseforge" || got.ProvenanceConfidence != 1 {
		t.Fatalf("exact provider provenance missing: %#v", got.Sources)
	}
}

func TestOmniManagerProtectsPatchedBuildFromProviderReplacement(t *testing.T) {
	root := t.TempDir()
	mods := filepath.Join(root, "mods")
	if err := os.MkdirAll(mods, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(mods, "cataclysm_spellbooks-1.2.9-patched.jar")
	if err := os.WriteFile(path, forgeJarBytes(t, "cataclysm_spellbooks", "Cataclysm: Spellbooks", "1.2.9", "Example"), 0o644); err != nil {
		t.Fatal(err)
	}
	app := &App{cfgDir: t.TempDir(), settings: Settings{JavaRoot: root, GameVersion: "1.20.1", Loader: "forge", CurseForgeAPIKey: "test-key"}}
	item, _, err := app.inspectJavaLibraryPath(path, "mod", "java", "Default")
	if err != nil {
		t.Fatal(err)
	}
	if item.UpdateStatus != "modified" {
		t.Fatalf("patched build was not protected at scan time: %#v", item)
	}
	fingerprint := item.Hashes.CurseFingerprint
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/fingerprints", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"exactMatches": []map[string]any{{
					"id": fingerprint,
					"file": map[string]any{
						"id": 11, "modId": 88, "fileName": filepath.Base(path), "displayName": "1.2.9",
					},
				}},
			},
		})
	})
	mux.HandleFunc("/v1/mods", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{{
				"id": 88, "name": "Cataclysm: Spellbooks", "slug": "cataclysm-spellbooks",
			}},
		})
	})
	mux.HandleFunc("/v1/mods/88/files", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{{
				"id": 12, "fileName": "cataclysm-spellbooks-1.3.0.jar", "displayName": "1.3.0",
				"downloadUrl": "https://cdn.example/new.jar", "gameVersions": []string{"1.20.1", "Forge"},
			}},
		})
	})
	server := httptest.NewTLSServer(mux)
	defer server.Close()
	t.Setenv("MMV_CURSEFORGE_API_BASE", server.URL)
	app.httpClient = server.Client()
	items := []LibraryItem{item}
	if err := app.enrichLibraryCurseForge(context.Background(), items); err != nil {
		t.Fatal(err)
	}
	got := items[0]
	if got.UpdateStatus != "modified" || got.SafeUpdate != nil || len(got.Alternatives) == 0 || got.Alternatives[0].Safe {
		t.Fatalf("patched build protection was bypassed: %#v", got)
	}
}

func bedrockPackBytes(t *testing.T, uuid, name, kind string) []byte {
	t.Helper()
	moduleType := "resources"
	if kind == "behavior" {
		moduleType = "data"
	}
	manifest := map[string]any{
		"format_version": 2,
		"header":         map[string]any{"name": "pack.name", "description": "pack.description", "uuid": uuid, "version": []int{1, 2, 3}, "min_engine_version": []int{1, 20, 80}},
		"modules":        []map[string]any{{"type": moduleType, "uuid": "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee", "version": []int{1, 2, 3}}},
		"metadata":       map[string]any{"authors": []string{"Bedrock Creator"}, "license": "MIT", "url": "https://example.invalid/bedrock"},
	}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	return zipBytes(t, map[string][]byte{
		"manifest.json":    data,
		"pack_icon.png":    tinyPNGBytes(t),
		"texts/en_US.lang": []byte("pack.name=" + name + "\npack.description=Beautiful Bedrock test pack\n"),
	})
}

func TestOmniManagerBedrockPackageInstallScanActivateAndUndo(t *testing.T) {
	cfg := t.TempDir()
	stable := filepath.Join(t.TempDir(), "com.mojang")
	if err := os.MkdirAll(stable, 0o755); err != nil {
		t.Fatal(err)
	}
	packagePath := filepath.Join(t.TempDir(), "premium-pack.mcpack")
	uuid := "12345678-1234-5678-9abc-123456789abc"
	if err := os.WriteFile(packagePath, bedrockPackBytes(t, uuid, "Premium Bedrock Pack", "resource"), 0o644); err != nil {
		t.Fatal(err)
	}
	app := &App{cfgDir: cfg, settings: Settings{JavaRoot: filepath.Join(t.TempDir(), "java"), BedrockRoot: stable, ActiveProfile: "Test"}}
	item, err := app.inspectBedrockArchivePackage(packagePath, "Vault downloads")
	if err != nil {
		t.Fatal(err)
	}
	if item.Name != "Premium Bedrock Pack" || item.UUID != uuid || item.InstalledVersion != "1.2.3" || item.LocalArtURL == "" {
		t.Fatalf("Bedrock package metadata incomplete: %#v", item)
	}
	profile, ok := app.resolveBedrockProfile("bedrock:stable")
	if !ok {
		t.Fatal("stable Bedrock profile was not resolved")
	}
	result, err := app.installBedrockPackage(context.Background(), packagePath, filepath.Base(packagePath), profile)
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || len(result.InstalledPaths) != 1 || result.ReceiptID == "" || !pathExists(result.InstalledPaths[0]) {
		t.Fatalf("Bedrock installation incomplete: %#v", result)
	}
	scanned, _, err := app.scanBedrockRoot(context.Background(), filepath.Join(stable, "resource_packs"), "resourcepack-bedrock", profile.Name)
	if err != nil || len(scanned) != 1 || scanned[0].Name != "Premium Bedrock Pack" || scanned[0].UUID != uuid {
		t.Fatalf("installed Bedrock pack was not scanned correctly: err=%v items=%#v", err, scanned)
	}
	world := filepath.Join(stable, "minecraftWorlds", "Premium World")
	if err := os.MkdirAll(world, 0o755); err != nil {
		t.Fatal(err)
	}
	activation, err := app.activateBedrockPack(BedrockActivationRequest{WorldPath: world, PackUUID: uuid, Version: "1.2.3", PackKind: "resourcepack"})
	if err != nil {
		t.Fatal(err)
	}
	activeFile := filepath.Join(world, "world_resource_packs.json")
	if data, err := os.ReadFile(activeFile); err != nil || !strings.Contains(string(data), uuid) {
		t.Fatalf("Bedrock world activation missing: %v %q", err, data)
	}
	if _, err := app.undoLibraryTransaction(activation.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(activeFile); !os.IsNotExist(err) {
		t.Fatalf("activation undo did not restore the absent file: %v", err)
	}
	if _, err := app.undoLibraryTransaction(result.ReceiptID); err != nil {
		t.Fatal(err)
	}
	if pathExists(result.InstalledPaths[0]) {
		t.Fatal("Bedrock install undo left the installed pack active")
	}
}

func TestOmniManagerRejectsUnsafeBedrockArchives(t *testing.T) {
	path := filepath.Join(t.TempDir(), "unsafe.mcpack")
	if err := os.WriteFile(path, zipBytes(t, map[string][]byte{"../escape.txt": []byte("bad"), "manifest.json": []byte(`{"format_version":2}`)}), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := validateBedrockArchive(path); err == nil || !strings.Contains(strings.ToLower(err.Error()), "unsafe path") {
		t.Fatalf("unsafe archive was not rejected: %v", err)
	}
}
