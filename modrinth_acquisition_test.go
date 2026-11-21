package main

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type acquisitionTransport func(*http.Request) (*http.Response, error)

func (f acquisitionTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestModrinthAcquisitionGuardPublicHandlers(t *testing.T) {
	type mutation func(map[string]ModrinthVersion)
	fixtures := []struct {
		name   string
		mutate mutation
		want   string
	}{
		{"two explicit pins", func(v map[string]ModrinthVersion) { setGuardDeps(v, "right", guardNeed("lib", "lib-3")) }, "no common exact version for lib: requires both lib-2 and lib-3"},
		{"flexible selection followed by different pin", func(v map[string]ModrinthVersion) {
			setGuardDeps(v, "left", guardNeed("lib", ""))
			setGuardDeps(v, "right", guardNeed("lib", "lib-3"))
		}, "review the exact selection"},
		{"version-only required identity", func(v map[string]ModrinthVersion) { setGuardDeps(v, "right", guardNeed("", "lib-2")) }, ""},
		{"pin owner mismatch", func(v map[string]ModrinthVersion) { setGuardDeps(v, "right", guardNeed("other", "lib-2")) }, "belongs to \"lib\", not requested project \"other\""},
		{"pin version identity mismatch", func(v map[string]ModrinthVersion) { a := v["lib-2"]; a.ID = "not-lib-2"; v["lib-2"] = a }, "unconfirmed version or project identity"},
		{"wrong game", func(v map[string]ModrinthVersion) {
			a := v["lib-2"]
			a.GameVersions = []string{"1.19.2"}
			v["lib-2"] = a
		}, "not confirmed for Minecraft 1.21.1 / fabric"},
		{"wrong loader", func(v map[string]ModrinthVersion) { a := v["lib-2"]; a.Loaders = []string{"forge"}; v["lib-2"] = a }, "not confirmed for Minecraft 1.21.1 / fabric"},
		{"missing required route", func(v map[string]ModrinthVersion) {
			setGuardDeps(v, "right", ModrinthDependency{FileName: "manual-only.jar", DependencyType: "required"})
		}, "manual-only.jar"},
		{"invalid root file", func(v map[string]ModrinthVersion) { a := v["root"]; a.Files = nil; v["root"] = a }, "no downloadable files"},
		{"required cycle", func(v map[string]ModrinthVersion) { setGuardDeps(v, "lib-2", guardNeed("root", "")) }, "required-dependency cycle"},
		{"duplicate aliases for same pin", func(v map[string]ModrinthVersion) { setGuardDeps(v, "right", guardNeed("library-slug", "lib-2")) }, ""},
		{"optional and embedded routes untouched", func(v map[string]ModrinthVersion) {
			setGuardDeps(v, "right", guardNeed("lib", "lib-2"), ModrinthDependency{ProjectID: "must-not-query", DependencyType: "optional"}, ModrinthDependency{ProjectID: "must-not-query", DependencyType: "embedded"}, ModrinthDependency{ProjectID: "must-not-query", DependencyType: "incompatible"})
		}, ""},
	}
	for _, f := range fixtures {
		for _, provider := range []bool{false, true} {
			name := f.name + "/legacy"
			if provider {
				name = f.name + "/provider"
			}
			t.Run(name, func(t *testing.T) {
				t.Setenv("MMV_MODRINTH_API_BASE", "https://metadata.example/v2")
				root := t.TempDir()
				versions := map[string]ModrinthVersion{}
				projectFor := func(key string) string {
					if strings.HasPrefix(key, "lib-") {
						return "lib"
					}
					return key
				}
				for _, key := range []string{"root", "left", "right", "lib-2", "lib-3", "other"} {
					bytes := []byte("test package bytes for " + key)
					sum := sha512.Sum512(bytes)
					versions[key] = ModrinthVersion{ProjectID: projectFor(key), ID: key, VersionNumber: key, GameVersions: []string{"1.21.1"}, Loaders: []string{"fabric"}, Files: []ModrinthVersionFile{{
						URL: "https://package.example/" + key + ".jar", Filename: key + ".jar", Primary: true, Size: int64(len(bytes)), Hashes: map[string]string{"sha512": hex.EncodeToString(sum[:])},
					}}}
				}
				setGuardDeps(versions, "root", guardNeed("left", ""), guardNeed("right", ""))
				setGuardDeps(versions, "left", guardNeed("lib", "lib-2"))
				setGuardDeps(versions, "right", guardNeed("lib", "lib-2"))
				f.mutate(versions)
				requests := []string{}
				firstPackage := false
				fixture := acquisitionTransport(func(req *http.Request) (*http.Response, error) {
					result := func(status int, data string) (*http.Response, error) {
						return &http.Response{StatusCode: status, Status: http.StatusText(status), Body: io.NopCloser(strings.NewReader(data)), Header: http.Header{}}, nil
					}
					if req.URL.Host == "package.example" {
						if f.want != "" {
							t.Errorf("unexpected package request before rejecting metadata: %s", req.URL)
						}
						firstPackage = true
						key := strings.TrimSuffix(strings.TrimPrefix(req.URL.Path, "/"), ".jar")
						requests = append(requests, key)
						return result(200, "test package bytes for "+key)
					}
					if req.URL.Host != "metadata.example" {
						t.Fatalf("unexpected provider route: %s", req.URL)
					}
					if firstPackage {
						t.Error("metadata resolved after acquisition started")
					}
					p := strings.Split(strings.Trim(req.URL.Path, "/"), "/")
					if len(p) < 3 || p[2] == "must-not-query" {
						t.Fatalf("invalid provider metadata route: %s", req.URL)
					}
					var value any
					switch {
					case p[1] == "version":
						row, ok := versions[p[2]]
						if !ok {
							return result(404, "")
						}
						value = row
					case p[1] == "project" && len(p) == 3:
						id := p[2]
						if id == "library-slug" {
							id = "lib"
						}
						if id != "root" && id != "left" && id != "right" && id != "lib" && id != "other" {
							return result(404, "")
						}
						slug := id
						if id == "lib" {
							slug = "library-slug"
						}
						value = ModrinthProject{ID: id, Slug: slug, Title: id, ProjectType: "mod"}
					case p[1] == "project" && len(p) == 4 && p[3] == "version":
						if req.URL.Query().Get("game_versions") != `["1.21.1"]` || req.URL.Query().Get("loaders") != `["fabric"]` {
							t.Fatal("lost exact metadata filters")
						}
						key := p[2]
						if key == "lib" {
							key = "lib-2"
						}
						value = []ModrinthVersion{versions[key]}
					default:
						t.Fatalf("unexpected metadata path: %s", req.URL)
					}
					blob, err := json.Marshal(value)
					if err != nil {
						t.Fatal(err)
					}
					return result(200, string(blob))
				})
				app := &App{settings: Settings{JavaRoot: root, GameVersion: "1.21.1", Loader: "fabric"}, httpClient: &http.Client{Transport: fixture}}
				rec := httptest.NewRecorder()
				if provider {
					app.handleProviderInstall(rec, httptest.NewRequest(http.MethodPost, "/api/providers/install", strings.NewReader(`{"provider":"modrinth","id":"root","target":"mods"}`)))
				} else {
					app.handleModrinthInstall(rec, httptest.NewRequest(http.MethodPost, "/api/modrinth/install", strings.NewReader(`{"projectId":"root","target":"mods"}`)))
				}
				if f.want == "" {
					if rec.Code != http.StatusOK {
						t.Fatalf("%d %s", rec.Code, rec.Body.String())
					}
					if got := strings.Join(requests, ","); got != "lib-2,left,right,root" {
						t.Fatalf("unexpected order or duplicate: %s", got)
					}
					for _, key := range requests {
						data, err := os.ReadFile(filepath.Join(root, "mods", key+".jar"))
						if err != nil || string(data) != "test package bytes for "+key {
							t.Fatalf("wrong fixture bytes for %s: %s %v", key, data, err)
						}
					}
				} else {
					var failure APIError
					if err := json.Unmarshal(rec.Body.Bytes(), &failure); err != nil || rec.Code != http.StatusBadGateway || !strings.Contains(failure.Error, f.want) || !strings.Contains(failure.Error, "needs review before downloading") {
						t.Fatalf("got %d %s; want %q", rec.Code, rec.Body.String(), f.want)
					}
					entries, err := os.ReadDir(root)
					if err != nil || len(entries) != 0 || len(requests) != 0 {
						t.Fatalf("guard failed: dirs %v, requests %v, err %v", entries, requests, err)
					}
				}
			})
		}
	}
}
func guardNeed(project, pin string) ModrinthDependency {
	return ModrinthDependency{ProjectID: project, VersionID: pin, DependencyType: "required"}
}
func setGuardDeps(versions map[string]ModrinthVersion, id string, deps ...ModrinthDependency) {
	v := versions[id]
	v.Dependencies = deps
	versions[id] = v
}
