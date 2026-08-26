package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestModrinthCompatibleResolutionAndDeps(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/project/demo", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "p1", "slug": "demo", "source_url": "https://github.com/acme/demo"})
	})
	mux.HandleFunc("/project/p1/version", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]map[string]any{{"id": "v2", "project_id": "p1", "version_number": "2.0.0", "version_type": "release", "date_published": "2026-08-25T00:00:00Z", "game_versions": []string{"1.21.1"}, "loaders": []string{"neoforge"}, "files": []map[string]any{{"hashes": map[string]string{"sha256": "abc"}, "url": "https://cdn.test/demo.jar", "filename": "demo.jar", "primary": true, "size": 10}}, "dependencies": []map[string]any{{"project_id": "dep1", "dependency_type": "required"}}}})
	})
	s := httptest.NewServer(mux)
	defer s.Close()
	p := newProviderClient(s.Client())
	p.modrinth = s.URL
	c, e := p.resolveModrinth(context.Background(), ProviderRef{Type: "modrinth", Project: "demo"}, Target{Minecraft: "1.21.1", Loader: "neoforge", Channel: "release"})
	if e != nil {
		t.Fatal(e)
	}
	if c.Version != "2.0.0" || len(c.Dependencies) != 1 {
		t.Fatalf("bad candidate: %+v", c)
	}
	if !strings.Contains(c.SourceArchive, "github.com/acme/demo") {
		t.Fatalf("source not derived: %s", c.SourceArchive)
	}
}
func TestDriveReplacePreservesID(t *testing.T) {
	var body []byte
	renamed := false
	mux := http.NewServeMux()
	mux.HandleFunc("/upload/files/file123", func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)
		_ = json.NewEncoder(w).Encode(map[string]string{"id": "file123", "name": "old.jar"})
	})
	mux.HandleFunc("/api/files/file123", func(w http.ResponseWriter, r *http.Request) {
		renamed = true
		_ = json.NewEncoder(w).Encode(map[string]string{"id": "file123", "name": "new.jar"})
	})
	s := httptest.NewServer(mux)
	defer s.Close()
	_ = os.Setenv("TEST_DRIVE_TOKEN", "token")
	defer os.Unsetenv("TEST_DRIVE_TOKEN")
	d, e := newDriveClient(DriveConfig{AccessTokenEnv: "TEST_DRIVE_TOKEN", APIBase: s.URL + "/api", UploadBase: s.URL + "/upload"}, s.Client())
	if e != nil {
		t.Fatal(e)
	}
	id, e := d.replace(context.Background(), "file123", "new.jar", []byte("new-bytes"))
	if e != nil {
		t.Fatal(e)
	}
	if id != "file123" || string(body) != "new-bytes" || !renamed {
		t.Fatalf("id=%s body=%q renamed=%v", id, body, renamed)
	}
}
func TestValidateDuplicate(t *testing.T) {
	r := Registry{Schema: 1, Artifacts: []ManagedArtifact{{ID: "x", Providers: []ProviderRef{{Type: "url", URL: "https://x"}}}, {ID: "x", Providers: []ProviderRef{{Type: "url", URL: "https://y"}}}}}
	if validateRegistry(r) == nil {
		t.Fatal("expected duplicate error")
	}
}

func TestCurrentRuntimeMissingSourcePlansRefresh(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/project/demo", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "p1", "slug": "demo", "source_url": "https://github.com/acme/demo"})
	})
	mux.HandleFunc("/project/p1/version", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]map[string]any{{"id": "v2", "project_id": "p1", "version_number": "2.0.0", "version_type": "release", "date_published": "2026-08-25T00:00:00Z", "game_versions": []string{"1.21.1"}, "loaders": []string{"fabric"}, "files": []map[string]any{{"hashes": map[string]string{"sha256": "abc"}, "url": "https://cdn.test/demo.jar", "filename": "demo.jar", "primary": true, "size": 10}}}})
	})
	s := httptest.NewServer(mux)
	defer s.Close()
	reg := Registry{Schema: 2, Defaults: Target{Channel: "release"}, Artifacts: []ManagedArtifact{{ID: "demo", Name: "Demo", Kind: "library", Version: "2.0.0", Filename: "demo.jar", Target: Target{Minecraft: "1.21.1", Loader: "fabric"}, Providers: []ProviderRef{{Type: "modrinth", Project: "demo", Priority: 100}}, UpdatePolicy: UpdatePolicy{MirrorSource: true}}}}
	eng := newEngine(reg, s.Client())
	eng.p.modrinth = s.URL
	plan, err := eng.plan(context.Background(), Target{})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Artifacts) != 1 || plan.Artifacts[0].Status != "source-refresh" {
		t.Fatalf("unexpected plan: %+v", plan.Artifacts)
	}
}

func TestCompareVersionsBlocksOlder(t *testing.T) {
	if compareVersions("1.9.9", "2.0.0") >= 0 {
		t.Fatal("older version not detected")
	}
	if compareVersions("2.0.0", "2.0.0") != 0 {
		t.Fatal("equal versions differ")
	}
	if compareVersions("2.1.0", "2.0.9") <= 0 {
		t.Fatal("newer version not detected")
	}
}
