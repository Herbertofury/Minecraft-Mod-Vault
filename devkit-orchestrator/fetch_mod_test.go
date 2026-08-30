package main

import (
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveExactModrinthAcceptsSlugAndVerifiesCanonicalProject(t *testing.T) {
	jar := []byte("exact-jar")
	s256 := sha256.Sum256(jar)
	s1 := sha1.Sum(jar)
	var ts *httptest.Server
	mux := http.NewServeMux()
	mux.HandleFunc("/project/aoa", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"id":"9qn2AQBc","slug":"aoa","source_url":"https://github.com/example/aoa"}`)
	})
	mux.HandleFunc("/version/VCCCalGp", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"id":"VCCCalGp","project_id":"9qn2AQBc","version_number":"1.16.5-3.6.11","version_type":"release","game_versions":["1.16.5"],"loaders":["forge"],"files":[{"hashes":{"sha256":"%s","sha1":"%s"},"url":"%s/file.jar","filename":"AoA3.jar","primary":true,"size":%d}]}`, hex.EncodeToString(s256[:]), hex.EncodeToString(s1[:]), ts.URL, len(jar))
	})
	mux.HandleFunc("/file.jar", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(jar) })
	ts = httptest.NewServer(mux)
	defer ts.Close()
	p := newProviderClient(ts.Client())
	p.modrinth = ts.URL
	c, err := p.resolveExactModrinth(context.Background(), "aoa", "VCCCalGp")
	if err != nil {
		t.Fatal(err)
	}
	if c.ProjectID != "9qn2AQBc" || c.VersionID != "VCCCalGp" || c.Filename != "AoA3.jar" {
		t.Fatalf("bad candidate: %+v", c)
	}
}

func TestResolveExactModrinthRejectsCrossProjectVersion(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/project/aoa", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, `{"id":"9qn2AQBc","slug":"aoa"}`) })
	mux.HandleFunc("/version/wrong", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"id":"wrong","project_id":"OTHER","version_number":"x","files":[{"url":"https://invalid","filename":"x.jar","primary":true}]}`)
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()
	p := newProviderClient(ts.Client())
	p.modrinth = ts.URL
	_, err := p.resolveExactModrinth(context.Background(), "aoa", "wrong")
	if err == nil || !strings.Contains(err.Error(), "provenance mismatch") {
		t.Fatalf("expected provenance mismatch, got %v", err)
	}
}

func TestFetchCandidateAtomicAndVerified(t *testing.T) {
	jar := []byte("verified-bytes")
	s256 := sha256.Sum256(jar)
	s1 := sha1.Sum(jar)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(jar) }))
	defer ts.Close()
	eng := newEngine(Registry{}, ts.Client())
	out := filepath.Join(t.TempDir(), "nested", "demo.jar")
	c := Candidate{Provider: "modrinth", ProjectID: "p", VersionID: "v", Version: "1", Filename: "demo.jar", URL: ts.URL, Size: int64(len(jar)), SHA256: hex.EncodeToString(s256[:]), SHA1: hex.EncodeToString(s1[:])}
	rec, err := fetchCandidateToPath(context.Background(), eng, c, out)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(jar) || rec.SHA256 != hex.EncodeToString(s256[:]) {
		t.Fatalf("bad output/receipt: %+v", rec)
	}
	leftovers, _ := filepath.Glob(filepath.Join(filepath.Dir(out), ".mmv-fetch-*.tmp"))
	if len(leftovers) != 0 {
		t.Fatalf("temp files left behind: %v", leftovers)
	}
}

func TestFetchCandidateDoesNotReplaceOnHashFailure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "bad") }))
	defer ts.Close()
	eng := newEngine(Registry{}, ts.Client())
	dir := t.TempDir()
	out := filepath.Join(dir, "demo.jar")
	if err := os.WriteFile(out, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}
	c := Candidate{Filename: "demo.jar", URL: ts.URL, SHA256: strings.Repeat("0", 64)}
	if _, err := fetchCandidateToPath(context.Background(), eng, c, out); err == nil {
		t.Fatal("expected hash failure")
	}
	got, _ := os.ReadFile(out)
	if string(got) != "old" {
		t.Fatalf("existing artifact was replaced on failed verification: %q", got)
	}
}

func TestAoAExactModrinthFallbackURLsAreDeterministic(t *testing.T) {
	c := Candidate{Provider: "modrinth", ProjectID: "9qn2AQBc", VersionID: "VCCCalGp", Filename: "AoA3-1.16.5-3.6.11.jar"}
	cdn := modrinthCanonicalCDNURL(c)
	maven := modrinthMavenURL(c)
	if cdn != "https://cdn.modrinth.com/data/9qn2AQBc/versions/VCCCalGp/AoA3-1.16.5-3.6.11.jar" {
		t.Fatalf("unexpected canonical CDN URL: %s", cdn)
	}
	if maven != "https://api.modrinth.com/maven/maven/modrinth/9qn2AQBc/VCCCalGp/9qn2AQBc-VCCCalGp.jar" {
		t.Fatalf("unexpected Maven fallback URL: %s", maven)
	}
}

func TestFetchCandidateFallsBackAfterPrimary404(t *testing.T) {
	jar := []byte("fallback-verified")
	s256 := sha256.Sum256(jar)
	mux := http.NewServeMux()
	mux.HandleFunc("/primary", func(w http.ResponseWriter, r *http.Request) { http.NotFound(w, r) })
	mux.HandleFunc("/fallback", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(jar) })
	ts := httptest.NewServer(mux)
	defer ts.Close()
	eng := newEngine(Registry{}, ts.Client())
	out := filepath.Join(t.TempDir(), "demo.jar")
	c := Candidate{Provider: "modrinth", Filename: "demo.jar", URL: ts.URL + "/primary", AlternateURLs: []string{ts.URL + "/fallback"}, Size: int64(len(jar)), SHA256: hex.EncodeToString(s256[:])}
	rec, err := fetchCandidateToPath(context.Background(), eng, c, out)
	if err != nil {
		t.Fatal(err)
	}
	if !rec.FallbackUsed || rec.ResolvedURL != ts.URL+"/fallback" {
		t.Fatalf("expected fallback URL, got %+v", rec)
	}
}

func TestFetchCandidateResumesInterruptedDownload(t *testing.T) {
	jar := []byte(strings.Repeat("resume-me-", 4096))
	s256 := sha256.Sum256(jar)
	var calls int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("Range") == "" {
			w.Header().Set("Content-Length", fmt.Sprint(len(jar)))
			_, _ = w.Write(jar[:len(jar)/2])
			return
		}
		want := fmt.Sprintf("bytes=%d-", len(jar)/2)
		if r.Header.Get("Range") != want {
			t.Fatalf("unexpected range header %q, want %q", r.Header.Get("Range"), want)
		}
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", len(jar)/2, len(jar)-1, len(jar)))
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write(jar[len(jar)/2:])
	}))
	defer ts.Close()
	eng := newEngine(Registry{}, ts.Client())
	out := filepath.Join(t.TempDir(), "resume.jar")
	c := Candidate{Provider: "modrinth", Filename: "resume.jar", URL: ts.URL, Size: int64(len(jar)), SHA256: hex.EncodeToString(s256[:])}
	rec, err := fetchCandidateToPath(context.Background(), eng, c, out)
	if err != nil {
		t.Fatal(err)
	}
	if !rec.Resumed || calls < 2 {
		t.Fatalf("expected a resumed transfer, receipt=%+v calls=%d", rec, calls)
	}
	got, _ := os.ReadFile(out)
	if string(got) != string(jar) {
		t.Fatal("resumed output did not match source")
	}
}

func TestCurseForgeAPIKeyIsPreservedAcrossRedirects(t *testing.T) {
	jar := []byte("curseforge-authenticated")
	s1 := sha1.Sum(jar)
	var ts *httptest.Server
	mux := http.NewServeMux()
	mux.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "secret" {
			t.Fatalf("missing key on initial request")
		}
		http.Redirect(w, r, ts.URL+"/file", http.StatusFound)
	})
	mux.HandleFunc("/file", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "secret" {
			t.Fatalf("missing key after redirect")
		}
		_, _ = w.Write(jar)
	})
	ts = httptest.NewServer(mux)
	defer ts.Close()
	eng := newEngine(Registry{}, ts.Client())
	eng.p.cfKey = "secret"
	out := filepath.Join(t.TempDir(), "cf.jar")
	c := Candidate{Provider: "curseforge", ProjectID: "123", VersionID: "4567890", Filename: "cf.jar", URL: ts.URL + "/start", Size: int64(len(jar)), SHA1: hex.EncodeToString(s1[:])}
	if _, err := fetchCandidateToPath(context.Background(), eng, c, out); err != nil {
		t.Fatal(err)
	}
}

func TestCurseForgeDownloadFailsActionablyWithoutAPIKey(t *testing.T) {
	eng := newEngine(Registry{}, http.DefaultClient)
	eng.p.cfKey = ""
	c := Candidate{Provider: "curseforge", ProjectID: "1", VersionID: "2", Filename: "x.jar", URL: "https://edge.forgecdn.net/files/0/002/x.jar"}
	_, err := fetchCandidateToPath(context.Background(), eng, c, filepath.Join(t.TempDir(), "x.jar"))
	if err == nil || !strings.Contains(err.Error(), "CURSEFORGE_API_KEY") {
		t.Fatalf("expected actionable missing-key error, got %v", err)
	}
}

func TestResolveExactCurseForgeUsesDownloadURLRouteWhenFileOmitsURL(t *testing.T) {
	jar := []byte("cf-file")
	s1 := sha1.Sum(jar)
	mux := http.NewServeMux()
	mux.HandleFunc("/mods/123/files/456", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "secret" {
			t.Fatalf("missing API key")
		}
		fmt.Fprintf(w, `{"data":{"id":456,"displayName":"Demo","fileName":"demo.jar","fileLength":%d,"releaseType":1,"hashes":[{"value":"%s","algo":1}]}}`, len(jar), hex.EncodeToString(s1[:]))
	})
	mux.HandleFunc("/mods/123/files/456/download-url", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"data":"https://edge.forgecdn.net/files/0/456/demo.jar"}`)
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()
	p := newProviderClient(ts.Client())
	p.curseforge = ts.URL
	p.cfKey = "secret"
	c, err := p.resolveExactCurseForge(context.Background(), "123", "456")
	if err != nil {
		t.Fatal(err)
	}
	if c.URL != "https://edge.forgecdn.net/files/0/456/demo.jar" || c.SHA1 != hex.EncodeToString(s1[:]) {
		t.Fatalf("unexpected candidate: %+v", c)
	}
}
