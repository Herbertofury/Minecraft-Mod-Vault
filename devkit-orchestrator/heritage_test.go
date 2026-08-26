package main

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
)

func zipFixture(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	keys := make([]string, 0, len(files))
	for k := range files {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		w, err := zw.Create(k)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(files[k])); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestHeritageAlwaysUsesNewestUpstreamAsFeatureAuthority(t *testing.T) {
	oldJar := zipFixture(t, map[string]string{"assets/demo/textures/item/old.png": "old", "data/demo/recipes/a.json": "{}"})
	newJar := zipFixture(t, map[string]string{"assets/demo/textures/item/old.png": "old", "assets/demo/textures/item/new.png": "new", "data/demo/recipes/a.json": "{\"new\":true}"})
	var base string
	mux := http.NewServeMux()
	mux.HandleFunc("/project/demo", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "p1", "slug": "demo", "source_url": ""})
	})
	versions := func() []map[string]any {
		oldVersion := map[string]any{"id": "v-old", "project_id": "p1", "name": "Legacy", "version_number": "1.0.0", "version_type": "release", "date_published": "2025-01-01T00:00:00Z", "game_versions": []string{"1.20.1"}, "loaders": []string{"forge"}, "changelog": "Legacy target", "files": []map[string]any{{"hashes": map[string]string{}, "url": base + "/files/old.jar", "filename": "demo-1.20.1.jar", "primary": true, "size": len(oldJar)}}}
		newVersion := map[string]any{"id": "v-new", "project_id": "p1", "name": "Modern", "version_number": "9.0.0", "version_type": "release", "date_published": "2026-08-25T00:00:00Z", "game_versions": []string{"26.2"}, "loaders": []string{"neoforge"}, "changelog": "Added the new item", "files": []map[string]any{{"hashes": map[string]string{}, "url": base + "/files/new.jar", "filename": "demo-26.2.jar", "primary": true, "size": len(newJar)}}}
		return []map[string]any{oldVersion, newVersion}
	}
	mux.HandleFunc("/project/p1/version", func(w http.ResponseWriter, r *http.Request) {
		vv := versions()
		if strings.Contains(r.URL.Query().Get("game_versions"), "1.20.1") {
			vv = vv[:1]
		}
		_ = json.NewEncoder(w).Encode(vv)
	})
	mux.HandleFunc("/project/demo/version", func(w http.ResponseWriter, r *http.Request) { _ = json.NewEncoder(w).Encode(versions()) })
	mux.HandleFunc("/files/old.jar", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(oldJar) })
	mux.HandleFunc("/files/new.jar", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(newJar) })
	s := httptest.NewServer(mux)
	defer s.Close()
	base = s.URL
	reg := Registry{Schema: 2, Defaults: Target{Channel: "release"}, Artifacts: []ManagedArtifact{{ID: "demo", Name: "Demo", Kind: "mod", Target: Target{Minecraft: "1.20.1", Loader: "forge"}, Providers: []ProviderRef{{Type: "modrinth", Project: "demo", Priority: 100}}}}}
	eng := newEngine(reg, s.Client())
	eng.p.modrinth = s.URL
	rep, err := eng.buildHeritage(context.Background(), "demo", Target{}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if rep.TargetCompatible.Version != "1.0.0" {
		t.Fatalf("compatible=%s", rep.TargetCompatible.Version)
	}
	if rep.LatestUpstream.Version != "9.0.0" {
		t.Fatalf("latest=%s", rep.LatestUpstream.Version)
	}
	found := false
	for _, e := range rep.RuntimeDelta.Added {
		if strings.HasSuffix(e.Path, "new.png") {
			found = true
		}
	}
	if !found {
		t.Fatalf("newest content not surfaced: %+v", rep.RuntimeDelta)
	}
	if len(rep.ReleaseLineage) != 2 || !strings.Contains(rep.ReleaseLineage[1].Changelog, "new item") {
		t.Fatalf("bad lineage: %+v", rep.ReleaseLineage)
	}
}

func TestPortGuardFailsWhenLatestFeatureIsMissing(t *testing.T) {
	origBytes := zipFixture(t, map[string]string{"assets/demo/textures/item/a.png": "same"})
	latestBytes := zipFixture(t, map[string]string{"assets/demo/textures/item/a.png": "same", "assets/demo/textures/item/b.png": "new"})
	convBytes := zipFixture(t, map[string]string{"assets/demo/textures/item/a.png": "same"})
	orig, _ := snapshotBytes("original", "original.jar", "original.jar", origBytes)
	latest, _ := snapshotBytes("latest", "latest.jar", "latest.jar", latestBytes)
	conv, _ := snapshotBytes("converted", "converted.jar", "converted.jar", convBytes)
	h := HeritageReport{RuntimeLatest: latest, RuntimeCompatible: orig}
	findings := auditPort(orig, conv, nil, h)
	for _, f := range findings {
		if f.Kind == "latest-feature-missing" && f.Severity == "error" {
			return
		}
	}
	t.Fatalf("expected latest-feature-missing error: %+v", findings)
}

func TestPortGuardFailsWhenOriginalUnchangedContentWasCorrupted(t *testing.T) {
	origBytes := zipFixture(t, map[string]string{"assets/demo/textures/item/a.png": "same"})
	convBytes := zipFixture(t, map[string]string{"assets/demo/textures/item/a.png": "corrupt"})
	orig, _ := snapshotBytes("original", "original.jar", "original.jar", origBytes)
	conv, _ := snapshotBytes("converted", "converted.jar", "converted.jar", convBytes)
	h := HeritageReport{RuntimeLatest: orig, RuntimeCompatible: orig}
	findings := auditPort(orig, conv, nil, h)
	for _, f := range findings {
		if f.Kind == "unchanged-content-corrupted" && f.Severity == "error" {
			return
		}
	}
	t.Fatalf("expected corruption error: %+v", findings)
}

func TestSnapshotRejectsUnsafeArchivePath(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("../escape.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = w.Write([]byte("bad"))
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := snapshotBytes("bad", "bad.jar", "bad.jar", buf.Bytes()); err == nil {
		t.Fatal("expected unsafe archive path rejection")
	}
}
