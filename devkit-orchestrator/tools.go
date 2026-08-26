package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type provenanceCatalog struct {
	Schema int              `json:"schema"`
	Count  int              `json:"count"`
	Tools  []provenanceTool `json:"tools"`
}

type provenanceTool struct {
	ID                 string        `json:"id"`
	Name               string        `json:"name"`
	Category           string        `json:"category"`
	Kind               string        `json:"kind"`
	Patterns           []string      `json:"patterns"`
	Capabilities       []string      `json:"capabilities,omitempty"`
	Loaders            []string      `json:"loaders,omitempty"`
	Minecraft          []string      `json:"minecraft,omitempty"`
	Platforms          []string      `json:"platforms,omitempty"`
	Homepage           string        `json:"homepage,omitempty"`
	UpdateProviders    []ProviderRef `json:"updateProviders"`
	AutoUpdateEligible bool          `json:"autoUpdateEligible"`
}

type driveItem struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	MimeType string   `json:"mimeType"`
	Parents  []string `json:"parents,omitempty"`
	Size     string   `json:"size,omitempty"`
}

func loadProvenanceCatalog(path string) (provenanceCatalog, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return provenanceCatalog{}, err
	}
	var c provenanceCatalog
	if err := json.Unmarshal(b, &c); err != nil {
		return c, err
	}
	return c, nil
}

func (d *driveClient) listChildren(ctx context.Context, parent string) ([]driveItem, error) {
	out := []driveItem{}
	token := ""
	for {
		q := url.Values{}
		q.Set("q", fmt.Sprintf("'%s' in parents and trashed = false", parent))
		q.Set("pageSize", "1000")
		q.Set("fields", "nextPageToken,files(id,name,mimeType,size,parents)")
		if token != "" {
			q.Set("pageToken", token)
		}
		var page struct {
			NextPageToken string      `json:"nextPageToken"`
			Files         []driveItem `json:"files"`
		}
		if err := d.do(ctx, "GET", d.api+"/files?"+q.Encode(), "", nil, &page); err != nil {
			return nil, err
		}
		out = append(out, page.Files...)
		if page.NextPageToken == "" {
			break
		}
		token = page.NextPageToken
	}
	return out, nil
}

func (d *driveClient) walkFiles(ctx context.Context, root string) ([]driveItem, error) {
	queue := []string{root}
	seen := map[string]bool{}
	files := []driveItem{}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if seen[id] {
			continue
		}
		seen[id] = true
		kids, err := d.listChildren(ctx, id)
		if err != nil {
			return nil, err
		}
		for _, x := range kids {
			if x.MimeType == "application/vnd.google-apps.folder" {
				queue = append(queue, x.ID)
				continue
			}
			files = append(files, x)
		}
	}
	sort.Slice(files, func(i, j int) bool { return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name) })
	return files, nil
}

func adoptTools(ctx context.Context, reg *Registry, catalog provenanceCatalog, dc *driveClient) (int, []string, error) {
	files, err := dc.walkFiles(ctx, reg.Drive.RootFolderID)
	if err != nil {
		return 0, nil, err
	}
	existingDrive := map[string]bool{}
	for _, a := range reg.Artifacts {
		if a.DriveFileID != "" {
			existingDrive[a.DriveFileID] = true
		}
	}
	adopted := 0
	skipped := []string{}
	for _, t := range catalog.Tools {
		if !t.AutoUpdateEligible || len(t.UpdateProviders) == 0 {
			continue
		}
		matched := 0
		for _, f := range files {
			if existingDrive[f.ID] || !toolPatternMatch(t.Patterns, f.Name) {
				continue
			}
			matched++
			ref, ok := providerForExistingTool(t, f.Name)
			if !ok {
				skipped = append(skipped, t.ID+":"+f.Name+" (no safe provider matcher)")
				continue
			}
			target := inferToolTarget(t, f.Name)
			id := "tool/" + safeName(t.ID) + "/" + safeName(f.ID)
			reg.Artifacts = append(reg.Artifacts, ManagedArtifact{
				ID: id, Name: t.Name, Kind: "tool", Filename: f.Name, DriveFileID: f.ID,
				Target: target, Providers: []ProviderRef{ref},
				UpdatePolicy: UpdatePolicy{KeepFilename: ref.Type == "github-branch"},
			})
			existingDrive[f.ID] = true
			adopted++
		}
		if matched == 0 {
			skipped = append(skipped, t.ID+" (no matching Drive file)")
		}
	}
	sort.Slice(reg.Artifacts, func(i, j int) bool { return reg.Artifacts[i].ID < reg.Artifacts[j].ID })
	return adopted, skipped, nil
}

func toolPatternMatch(patterns []string, name string) bool {
	low := strings.ToLower(name)
	for _, p := range patterns {
		ok, _ := filepath.Match(strings.ToLower(p), low)
		if ok {
			return true
		}
	}
	return false
}

func inferToolTarget(t provenanceTool, filename string) Target {
	x := Target{}
	if len(t.Minecraft) == 1 {
		x.Minecraft = t.Minecraft[0]
	}
	if len(t.Loaders) == 1 {
		x.Loader = strings.ToLower(t.Loaders[0])
	}
	low := strings.ToLower(filename)
	switch {
	case strings.Contains(low, "windows") || strings.HasSuffix(low, ".exe") || strings.HasSuffix(low, ".msi"):
		x.OS = "windows"
	case strings.Contains(low, "linux") || strings.HasSuffix(low, ".tar.gz"):
		x.OS = "linux"
	case strings.Contains(low, "mac") || strings.Contains(low, "darwin") || strings.HasSuffix(low, ".dmg"):
		x.OS = "darwin"
	}
	if strings.Contains(low, "x64") || strings.Contains(low, "x86_64") || strings.Contains(low, "amd64") || strings.Contains(low, "w64") {
		x.Arch = "amd64"
	}
	if strings.Contains(low, "aarch64") || strings.Contains(low, "arm64") {
		x.Arch = "arm64"
	}
	return x
}

func providerForExistingTool(t provenanceTool, filename string) (ProviderRef, bool) {
	p := t.UpdateProviders[0]
	p.Priority = maxInt(p.Priority, 100)
	low := strings.ToLower(filename)
	if p.Type == "modrinth" {
		return p, true
	}
	if p.Type != "github" && p.Type != "github-branch" {
		return p, p.Type != ""
	}
	if looksLikeSourceSnapshot(low) || branchManagedTool(t.ID) {
		p.Type = "github-branch"
		p.AssetRegex = ""
		p.Branch = branchForFilename(low)
		return p, true
	}
	p.Type = "github"
	p.AssetRegex = releaseAssetRegex(t.ID, filename)
	if p.AssetRegex == "" {
		return ProviderRef{}, false
	}
	if _, err := regexp.Compile(p.AssetRegex); err != nil {
		return ProviderRef{}, false
	}
	return p, true
}

func looksLikeSourceSnapshot(low string) bool {
	return strings.Contains(low, "-main.zip") || strings.Contains(low, "-master.zip") || strings.Contains(low, "-develop.zip") || strings.Contains(low, "source")
}
func branchForFilename(low string) string {
	if strings.Contains(low, "develop") {
		return "develop"
	}
	if strings.Contains(low, "master") {
		return "master"
	}
	return "main"
}
func branchManagedTool(id string) bool {
	switch id {
	case "neoforge-1.21.1-mdg", "neoforge-1.21.1-neogradle", "neoforge-26.1-mdg", "fabric-example", "quilt-template", "architectury-loom", "architectury-api", "architectury-templates", "modstitch", "stonecutter", "packwiz", "parchment", "quilt-mappings", "datafixerupper", "quilt-loader", "quilted-fabric-api", "quilt-config", "geckolib-source", "vineflower-source", "recaf-source", "enigma", "renderdoc":
		return true
	}
	return false
}

func releaseAssetRegex(id, filename string) string {
	switch id {
	case "java-17":
		return `(?i)^OpenJDK17U-jdk_.*\.(?:zip|tar\.gz)$`
	case "java-21":
		return `(?i)^OpenJDK21U-jdk_.*\.(?:zip|tar\.gz)$`
	case "java-25":
		return `(?i)^OpenJDK25U-jdk_.*\.(?:zip|tar\.gz)$`
	case "java-26":
		return `(?i)^OpenJDK26U-jdk_.*\.(?:zip|tar\.gz)$`
	case "gradle-8.8":
		return `(?i)^gradle-8\.8(?:\.\d+)*-bin\.zip$`
	case "gradle-8.14":
		return `(?i)^gradle-8\.14(?:\.\d+)*-bin\.zip$`
	case "mcreator":
		return `(?i)^MCreator\..*(?:Windows\.64bit\.(?:exe|zip)|Linux\.64bit\.tar\.gz)$`
	case "prism":
		return `(?i)^PrismLauncher-.*(?:Portable.*\.zip|Setup.*\.exe)$`
	case "blockbench":
		return `(?i)^Blockbench_.*portable\.exe$`
	case "nbt-studio":
		return `(?i).*nbt.*studio.*\.(?:exe|zip)$`
	case "mca-selector":
		return `(?i)^mcaselector-.*-setup\.exe$`
	case "vineflower":
		return `(?i)^vineflower-[0-9].*\.jar$`
	case "recaf":
		low := strings.ToLower(filename)
		if strings.Contains(low, "launcher-gui") {
			return `(?i).*launcher-gui.*\.jar$`
		}
		if strings.Contains(low, "launcher-cli") {
			return `(?i).*launcher-cli.*\.jar$`
		}
		return `(?i)^recaf-4.*\.jar$`
	case "ferium":
		return `(?i).*ferium.*(?:\.exe|\.zip|\.tar\.gz)?$`
	case "mixinextras":
		return `(?i)^MixinExtras-.*\.jar$|^mixinextras-.*\.zip$`
	case "spark-forge":
		return `(?i)^spark-.*forge.*\.jar$`
	case "spark-neoforge":
		return `(?i)^spark-.*neoforge.*\.jar$`
	case "async-profiler":
		return `(?i)^async-profiler-.*(?:\.zip|\.tar\.gz)$|^async-profiler\.jar$`
	case "visualvm":
		return `(?i)^visualvm[_-].*\.zip$|^VisualVM[_-].*\.dmg$`
	}
	return ""
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
