package main

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

const (
	atlasMojangDetailsPath = "assets/seeds/mojang-version-details-v2.jsonl.gz"
	atlasMCMetaPath        = "assets/seeds/mcmeta-versions.json"
	atlasReviewedAt        = "2026-08-20T00:00:00Z"
)

type AtlasSummary struct {
	SchemaVersion       int                    `json:"schemaVersion"`
	ReviewedAt          string                 `json:"reviewedAt"`
	GeneratedAt         string                 `json:"generatedAt"`
	MojangVersions      int                    `json:"mojangVersions"`
	MCMetaVersions      int                    `json:"mcmetaVersions"`
	LatestRelease       string                 `json:"latestRelease"`
	LatestSnapshot      string                 `json:"latestSnapshot"`
	NewestVersion       string                 `json:"newestVersion"`
	OldestVersion       string                 `json:"oldestVersion"`
	VersionsWithJava    int                    `json:"versionsWithJava"`
	VersionsWithServer  int                    `json:"versionsWithServer"`
	VersionsWithMaps    int                    `json:"versionsWithMappings"`
	RuntimeLibraries    int                    `json:"runtimeLibraryRows"`
	Toolchains          int                    `json:"toolchains"`
	EvidenceSources     []AtlasEvidenceSource  `json:"evidenceSources"`
	LoaderCoverage      map[string]int         `json:"loaderCoverage"`
	ToolchainHighlights []AtlasToolchainChoice `json:"toolchainHighlights"`
}

type AtlasEvidenceSource struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	URL       string `json:"url"`
	SeedPath  string `json:"seedPath"`
	SHA256    string `json:"sha256,omitempty"`
	Records   int    `json:"records,omitempty"`
	Retrieved string `json:"retrieved"`
}

type AtlasMCMeta struct {
	ID                       string `json:"id"`
	Name                     string `json:"name"`
	ReleaseTarget            string `json:"releaseTarget,omitempty"`
	Type                     string `json:"type"`
	Stable                   bool   `json:"stable"`
	DataVersion              int    `json:"dataVersion,omitempty"`
	ProtocolVersion          int    `json:"protocolVersion,omitempty"`
	DataPackVersion          int    `json:"dataPackVersion,omitempty"`
	DataPackVersionMinor     int    `json:"dataPackVersionMinor,omitempty"`
	ResourcePackVersion      int    `json:"resourcePackVersion,omitempty"`
	ResourcePackVersionMinor int    `json:"resourcePackVersionMinor,omitempty"`
	BuildTime                string `json:"buildTime,omitempty"`
	ReleaseTime              string `json:"releaseTime,omitempty"`
	SHA1                     string `json:"sha1,omitempty"`
}

type AtlasDownload struct {
	SHA1 string `json:"sha1,omitempty"`
	Size int64  `json:"size,omitempty"`
	URL  string `json:"url,omitempty"`
}

type AtlasVersion struct {
	ID                 string                   `json:"id"`
	Type               string                   `json:"type"`
	Time               string                   `json:"time,omitempty"`
	ReleaseTime        string                   `json:"releaseTime,omitempty"`
	ManifestURL        string                   `json:"manifestUrl,omitempty"`
	ManifestSHA1       string                   `json:"manifestSha1,omitempty"`
	ManifestSHA256     string                   `json:"manifestSha256,omitempty"`
	JavaMajor          int                      `json:"javaMajor,omitempty"`
	JavaComponent      string                   `json:"javaComponent,omitempty"`
	MainClass          string                   `json:"mainClass,omitempty"`
	AssetIndex         string                   `json:"assetIndex,omitempty"`
	LibraryCount       int                      `json:"libraryCount"`
	HasClient          bool                     `json:"hasClient"`
	HasServer          bool                     `json:"hasServer"`
	HasClientMappings  bool                     `json:"hasClientMappings"`
	HasServerMappings  bool                     `json:"hasServerMappings"`
	Downloads          map[string]AtlasDownload `json:"downloads,omitempty"`
	MCMeta             *AtlasMCMeta             `json:"mcmeta,omitempty"`
	FabricSupported    bool                     `json:"fabricSupported"`
	QuiltSupported     bool                     `json:"quiltSupported"`
	ModrinthRecognized bool                     `json:"modrinthRecognized"`
}

type AtlasMavenArtifact struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	GroupID       string   `json:"groupId"`
	ArtifactID    string   `json:"artifactId"`
	Latest        string   `json:"latest,omitempty"`
	Release       string   `json:"release,omitempty"`
	LatestStable  string   `json:"latestStable,omitempty"`
	VersionCount  int      `json:"versionCount"`
	Versions      []string `json:"versions,omitempty"`
	SourcePath    string   `json:"sourcePath"`
	OfficialURL   string   `json:"officialUrl"`
	RepositoryURL string   `json:"repositoryUrl,omitempty"`
}

type AtlasLoaderVersion struct {
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
	Maven   string `json:"maven,omitempty"`
}

type RepairVersionAtlas struct {
	Summary         AtlasSummary
	Versions        []AtlasVersion
	ByID            map[string]AtlasVersion
	MCMeta          map[string]AtlasMCMeta
	Maven           map[string]AtlasMavenArtifact
	FabricGames     map[string]bool
	QuiltGames      map[string]bool
	ModrinthGames   map[string]bool
	FabricLoaders   []AtlasLoaderVersion
	QuiltLoaders    []AtlasLoaderVersion
	ModrinthLoaders []string
}

type AtlasVersionQueryResponse struct {
	Total    int            `json:"total"`
	Offset   int            `json:"offset"`
	Limit    int            `json:"limit"`
	Versions []AtlasVersion `json:"versions"`
}

type AtlasToolchainChoice struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Version     string `json:"version,omitempty"`
	Channel     string `json:"channel"`
	Reason      string `json:"reason"`
	OfficialURL string `json:"officialUrl,omitempty"`
	SourcePath  string `json:"sourcePath,omitempty"`
}

type AtlasResolveRequest struct {
	GameVersion string `json:"gameVersion"`
	Loader      string `json:"loader"`
}

type AtlasResolution struct {
	ResolvedAt          string                 `json:"resolvedAt"`
	ReviewedAt          string                 `json:"reviewedAt"`
	GameVersion         string                 `json:"gameVersion"`
	Loader              string                 `json:"loader"`
	Exists              bool                   `json:"exists"`
	Version             *AtlasVersion          `json:"version,omitempty"`
	JavaMajor           int                    `json:"javaMajor"`
	LoaderSupported     bool                   `json:"loaderSupported"`
	LoaderVersion       string                 `json:"loaderVersion,omitempty"`
	LoaderChannel       string                 `json:"loaderChannel,omitempty"`
	GameArtifacts       []AtlasToolchainChoice `json:"gameArtifacts"`
	BuildToolchains     []AtlasToolchainChoice `json:"buildToolchains"`
	Mappings            []AtlasToolchainChoice `json:"mappings"`
	CompatibilityRoutes []AtlasToolchainChoice `json:"compatibilityRoutes"`
	Warnings            []string               `json:"warnings,omitempty"`
	Evidence            []AtlasEvidenceSource  `json:"evidence"`
}

type mojangAtlasLine struct {
	ID           string `json:"id"`
	SourceSHA1   string `json:"sourceSHA1"`
	SourceSHA256 string `json:"sourceSHA256"`
	SourceURL    string `json:"sourceUrl"`
	Manifest     struct {
		ID          string `json:"id"`
		Type        string `json:"type"`
		Time        string `json:"time"`
		ReleaseTime string `json:"releaseTime"`
		MainClass   string `json:"mainClass"`
		JavaVersion struct {
			Component    string `json:"component"`
			MajorVersion int    `json:"majorVersion"`
		} `json:"javaVersion"`
		AssetIndex struct {
			ID string `json:"id"`
		} `json:"assetIndex"`
		Downloads map[string]AtlasDownload `json:"downloads"`
		Libraries []json.RawMessage        `json:"libraries"`
	} `json:"manifest"`
}

type mavenMetadataXML struct {
	XMLName    xml.Name `xml:"metadata"`
	GroupID    string   `xml:"groupId"`
	ArtifactID string   `xml:"artifactId"`
	Versioning struct {
		Latest   string   `xml:"latest"`
		Release  string   `xml:"release"`
		Versions []string `xml:"versions>version"`
	} `xml:"versioning"`
}

type atlasSeedSummary struct {
	Count               int    `json:"count"`
	GeneratedAt         string `json:"generatedAt"`
	LatestRelease       string `json:"latestRelease"`
	LatestSnapshot      string `json:"latestSnapshot"`
	Newest              string `json:"newest"`
	Oldest              string `json:"oldest"`
	RuntimeLibraryRows  int    `json:"runtimeLibraryRows"`
	VersionsWithJava    int    `json:"versionsWithJavaRuntime"`
	VersionsWithServer  int    `json:"versionsWithServer"`
	VersionsWithMapping int    `json:"versionsWithClientMappings"`
	SourceManifestSHA   string `json:"sourceManifestSHA256"`
	GzipSHA             string `json:"gzipSHA256"`
}

var atlasGlobal struct {
	once sync.Once
	data *RepairVersionAtlas
	err  error
}

var atlasMavenDefinitions = []struct {
	ID, Name, Path, OfficialURL, RepositoryURL string
}{
	{"fabric-loom", "Fabric Loom", "assets/seeds/fabric-loom-maven-metadata.xml", "https://docs.fabricmc.net/develop/loom/", "https://github.com/FabricMC/fabric-loom"},
	{"forge", "MinecraftForge", "assets/seeds/forge-maven-metadata.xml", "https://files.minecraftforge.net/", "https://github.com/MinecraftForge/MinecraftForge"},
	{"forgegradle", "ForgeGradle", "assets/seeds/forgegradle-maven-metadata.xml", "https://docs.minecraftforge.net/en/latest/gettingstarted/", "https://github.com/MinecraftForge/ForgeGradle"},
	{"moddevgradle", "ModDevGradle", "assets/seeds/moddevgradle-plugin-maven-metadata.xml", "https://docs.neoforged.net/toolchain/docs/plugins/mdg/", "https://github.com/neoforged/ModDevGradle"},
	{"neoforge", "NeoForge", "assets/seeds/neoforge-maven-metadata.xml", "https://neoforged.net/", "https://github.com/neoforged/NeoForge"},
	{"neoform", "NeoForm", "assets/seeds/neoform-maven-metadata.xml", "https://projects.neoforged.net/neoforged/neoform", "https://github.com/neoforged/NeoForm"},
	{"neogradle", "NeoGradle UserDev", "assets/seeds/neogradle-userdev-maven-metadata.xml", "https://docs.neoforged.net/toolchain/docs/plugins/ng/", "https://github.com/neoforged/NeoGradle"},
	{"architectury-loom", "Architectury Loom", "assets/seeds/architectury-loom-maven-metadata.xml", "https://docs.architectury.dev/loom/introduction", "https://github.com/architectury/architectury-loom"},
	{"auto-renaming-tool", "Auto Renaming Tool", "assets/seeds/auto-renaming-tool-maven-metadata.xml", "https://github.com/MinecraftForge/AutoRenamingTool", "https://github.com/MinecraftForge/AutoRenamingTool"},
	{"fabric-api", "Fabric API", "assets/seeds/fabric-api-maven-metadata.xml", "https://fabricmc.net/develop/", "https://github.com/FabricMC/fabric"},
	{"mcpconfig", "MCPConfig", "assets/seeds/mcpconfig-maven-metadata.xml", "https://github.com/MinecraftForge/MCPConfig", "https://github.com/MinecraftForge/MCPConfig"},
	{"intermediary", "Fabric Intermediary", "assets/seeds/intermediary-maven-metadata.xml", "https://fabricmc.net/wiki/tutorial:mappings", "https://github.com/FabricMC/intermediary"},
	{"mixin", "SpongePowered Mixin", "assets/seeds/mixin-maven-metadata.xml", "https://github.com/SpongePowered/Mixin/wiki", "https://github.com/SpongePowered/Mixin"},
	{"mixinextras", "MixinExtras", "assets/seeds/mixinextras-maven-metadata.xml", "https://github.com/LlamaLad7/MixinExtras/wiki", "https://github.com/LlamaLad7/MixinExtras"},
	{"quilt-mappings", "Quilt Mappings", "assets/seeds/quilt-mappings-maven-metadata.xml", "https://github.com/QuiltMC/quilt-mappings", "https://github.com/QuiltMC/quilt-mappings"},
	{"srgutils", "SRGUtils", "assets/seeds/srgutils-maven-metadata.xml", "https://github.com/MinecraftForge/SrgUtils", "https://github.com/MinecraftForge/SrgUtils"},
	{"yarn", "Yarn Mappings", "assets/seeds/yarn-maven-metadata.xml", "https://fabricmc.net/wiki/tutorial:mappings", "https://github.com/FabricMC/yarn"},
}

func loadRepairVersionAtlas() (*RepairVersionAtlas, error) {
	atlasGlobal.once.Do(func() {
		atlasGlobal.data, atlasGlobal.err = buildVersionAtlas()
	})
	return atlasGlobal.data, atlasGlobal.err
}

func buildVersionAtlas() (*RepairVersionAtlas, error) {
	atlas := &RepairVersionAtlas{
		ByID:            map[string]AtlasVersion{},
		MCMeta:          map[string]AtlasMCMeta{},
		Maven:           map[string]AtlasMavenArtifact{},
		FabricGames:     map[string]bool{},
		QuiltGames:      map[string]bool{},
		ModrinthGames:   map[string]bool{},
		ModrinthLoaders: []string{},
	}
	if err := loadAtlasMCMeta(atlas); err != nil {
		return nil, fmt.Errorf("load mcmeta atlas: %w", err)
	}
	if err := loadAtlasGameAndLoaderSeeds(atlas); err != nil {
		return nil, err
	}
	if err := loadAtlasMavenMetadata(atlas); err != nil {
		return nil, err
	}
	if err := loadAtlasMojangVersions(atlas); err != nil {
		return nil, err
	}
	if len(atlas.Versions) == 0 {
		return nil, errors.New("version atlas is empty")
	}

	var seed atlasSeedSummary
	if b, err := embeddedFiles.ReadFile("assets/seeds/Mojang-deep-seed-summary.json"); err == nil {
		_ = json.Unmarshal(b, &seed)
	}
	atlas.Summary = AtlasSummary{
		SchemaVersion:      1,
		ReviewedAt:         atlasReviewedAt,
		GeneratedAt:        firstNonEmpty(seed.GeneratedAt, atlasReviewedAt),
		MojangVersions:     len(atlas.Versions),
		MCMetaVersions:     len(atlas.MCMeta),
		LatestRelease:      seed.LatestRelease,
		LatestSnapshot:     seed.LatestSnapshot,
		NewestVersion:      firstNonEmpty(seed.Newest, atlas.Versions[0].ID),
		OldestVersion:      firstNonEmpty(seed.Oldest, atlas.Versions[len(atlas.Versions)-1].ID),
		VersionsWithJava:   seed.VersionsWithJava,
		VersionsWithServer: seed.VersionsWithServer,
		VersionsWithMaps:   seed.VersionsWithMapping,
		RuntimeLibraries:   seed.RuntimeLibraryRows,
		Toolchains:         len(atlas.Maven),
		LoaderCoverage: map[string]int{
			"fabric":   len(atlas.FabricGames),
			"quilt":    len(atlas.QuiltGames),
			"modrinth": len(atlas.ModrinthGames),
		},
		EvidenceSources: atlasEvidence(seed),
	}
	if atlas.Summary.LatestRelease == "" || atlas.Summary.LatestSnapshot == "" {
		for _, version := range atlas.Versions {
			switch version.Type {
			case "release":
				if atlas.Summary.LatestRelease == "" {
					atlas.Summary.LatestRelease = version.ID
				}
			case "snapshot":
				if atlas.Summary.LatestSnapshot == "" {
					atlas.Summary.LatestSnapshot = version.ID
				}
			}
		}
	}
	for _, id := range []string{"fabric-loom", "moddevgradle", "forgegradle", "neoforge", "architectury-loom"} {
		if artifact, ok := atlas.Maven[id]; ok {
			version := firstNonEmpty(artifact.LatestStable, artifact.Release, artifact.Latest)
			channel := "stable"
			if isPrereleaseVersion(version) {
				channel = "preview"
			}
			atlas.Summary.ToolchainHighlights = append(atlas.Summary.ToolchainHighlights, AtlasToolchainChoice{
				ID: id, Name: artifact.Name, Version: version, Channel: channel,
				Reason: "Latest version captured from the official Maven metadata seed.", OfficialURL: artifact.OfficialURL, SourcePath: artifact.SourcePath,
			})
		}
	}
	return atlas, nil
}

func loadAtlasMCMeta(atlas *RepairVersionAtlas) error {
	b, err := embeddedFiles.ReadFile(atlasMCMetaPath)
	if err != nil {
		return err
	}
	var rows []struct {
		ID                       string `json:"id"`
		Name                     string `json:"name"`
		ReleaseTarget            string `json:"release_target"`
		Type                     string `json:"type"`
		Stable                   bool   `json:"stable"`
		DataVersion              int    `json:"data_version"`
		ProtocolVersion          int    `json:"protocol_version"`
		DataPackVersion          int    `json:"data_pack_version"`
		DataPackVersionMinor     int    `json:"data_pack_version_minor"`
		ResourcePackVersion      int    `json:"resource_pack_version"`
		ResourcePackVersionMinor int    `json:"resource_pack_version_minor"`
		BuildTime                string `json:"build_time"`
		ReleaseTime              string `json:"release_time"`
		SHA1                     string `json:"sha1"`
	}
	if err := json.Unmarshal(b, &rows); err != nil {
		return err
	}
	for _, row := range rows {
		atlas.MCMeta[row.ID] = AtlasMCMeta{
			ID: row.ID, Name: row.Name, ReleaseTarget: row.ReleaseTarget, Type: row.Type, Stable: row.Stable,
			DataVersion: row.DataVersion, ProtocolVersion: row.ProtocolVersion, DataPackVersion: row.DataPackVersion,
			DataPackVersionMinor: row.DataPackVersionMinor, ResourcePackVersion: row.ResourcePackVersion,
			ResourcePackVersionMinor: row.ResourcePackVersionMinor, BuildTime: row.BuildTime, ReleaseTime: row.ReleaseTime, SHA1: row.SHA1,
		}
	}
	return nil
}

func loadAtlasGameAndLoaderSeeds(atlas *RepairVersionAtlas) error {
	loadGameVersions := func(path string, target map[string]bool) error {
		b, err := embeddedFiles.ReadFile(path)
		if err != nil {
			return err
		}
		var rows []struct {
			Version string `json:"version"`
			Stable  bool   `json:"stable"`
		}
		if err := json.Unmarshal(b, &rows); err != nil {
			return err
		}
		for _, row := range rows {
			target[row.Version] = row.Stable
		}
		return nil
	}
	if err := loadGameVersions("assets/seeds/fabric-game-versions.json", atlas.FabricGames); err != nil {
		return fmt.Errorf("load Fabric game versions: %w", err)
	}
	if err := loadGameVersions("assets/seeds/quilt-game-versions.json", atlas.QuiltGames); err != nil {
		return fmt.Errorf("load Quilt game versions: %w", err)
	}
	if b, err := embeddedFiles.ReadFile("assets/seeds/modrinth-game-versions.json"); err == nil {
		var rows []struct {
			Version string `json:"version"`
		}
		if err := json.Unmarshal(b, &rows); err != nil {
			return fmt.Errorf("load Modrinth game versions: %w", err)
		}
		for _, row := range rows {
			atlas.ModrinthGames[row.Version] = true
		}
	} else {
		return err
	}
	if b, err := embeddedFiles.ReadFile("assets/seeds/fabric-loader-versions.json"); err == nil {
		var rows []AtlasLoaderVersion
		if err := json.Unmarshal(b, &rows); err != nil {
			return fmt.Errorf("load Fabric loader versions: %w", err)
		}
		atlas.FabricLoaders = rows
	} else {
		return err
	}
	if b, err := embeddedFiles.ReadFile("assets/seeds/quilt-loader-versions.json"); err == nil {
		var rows []AtlasLoaderVersion
		if err := json.Unmarshal(b, &rows); err != nil {
			return fmt.Errorf("load Quilt loader versions: %w", err)
		}
		atlas.QuiltLoaders = rows
	} else {
		return err
	}
	if b, err := embeddedFiles.ReadFile("assets/seeds/modrinth-loaders.json"); err == nil {
		var rows []struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(b, &rows); err != nil {
			return fmt.Errorf("load Modrinth loaders: %w", err)
		}
		for _, row := range rows {
			atlas.ModrinthLoaders = append(atlas.ModrinthLoaders, row.Name)
		}
	} else {
		return err
	}
	return nil
}

func loadAtlasMavenMetadata(atlas *RepairVersionAtlas) error {
	for _, def := range atlasMavenDefinitions {
		b, err := embeddedFiles.ReadFile(def.Path)
		if err != nil {
			return fmt.Errorf("load %s metadata: %w", def.ID, err)
		}
		var metadata mavenMetadataXML
		if err := xml.Unmarshal(b, &metadata); err != nil {
			return fmt.Errorf("parse %s metadata: %w", def.ID, err)
		}
		artifact := AtlasMavenArtifact{
			ID: def.ID, Name: def.Name, GroupID: metadata.GroupID, ArtifactID: metadata.ArtifactID,
			Latest: metadata.Versioning.Latest, Release: metadata.Versioning.Release, VersionCount: len(metadata.Versioning.Versions),
			Versions: metadata.Versioning.Versions, SourcePath: def.Path, OfficialURL: def.OfficialURL, RepositoryURL: def.RepositoryURL,
		}
		artifact.LatestStable = latestMatchingVersion(artifact.Versions, func(string) bool { return true }, false)
		atlas.Maven[def.ID] = artifact
	}
	return nil
}

func loadAtlasMojangVersions(atlas *RepairVersionAtlas) error {
	file, err := embeddedFiles.Open(atlasMojangDetailsPath)
	if err != nil {
		return err
	}
	defer file.Close()
	reader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer reader.Close()

	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64<<10), 16<<20)
	for scanner.Scan() {
		var row mojangAtlasLine
		if err := json.Unmarshal(scanner.Bytes(), &row); err != nil {
			return fmt.Errorf("parse Mojang detail row: %w", err)
		}
		id := firstNonEmpty(row.Manifest.ID, row.ID)
		version := AtlasVersion{
			ID: id, Type: row.Manifest.Type, Time: row.Manifest.Time, ReleaseTime: row.Manifest.ReleaseTime,
			ManifestURL: row.SourceURL, ManifestSHA1: row.SourceSHA1, ManifestSHA256: row.SourceSHA256,
			JavaMajor: row.Manifest.JavaVersion.MajorVersion, JavaComponent: row.Manifest.JavaVersion.Component,
			MainClass: row.Manifest.MainClass, AssetIndex: row.Manifest.AssetIndex.ID,
			LibraryCount: len(row.Manifest.Libraries), Downloads: row.Manifest.Downloads,
			FabricSupported: atlas.FabricGames[id], QuiltSupported: atlas.QuiltGames[id], ModrinthRecognized: atlas.ModrinthGames[id],
		}
		_, version.HasClient = row.Manifest.Downloads["client"]
		_, version.HasServer = row.Manifest.Downloads["server"]
		_, version.HasClientMappings = row.Manifest.Downloads["client_mappings"]
		_, version.HasServerMappings = row.Manifest.Downloads["server_mappings"]
		if meta, ok := atlas.MCMeta[id]; ok {
			copy := meta
			version.MCMeta = &copy
		}
		atlas.Versions = append(atlas.Versions, version)
		atlas.ByID[id] = version
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	sort.SliceStable(atlas.Versions, func(i, j int) bool {
		ti, ei := time.Parse(time.RFC3339, atlas.Versions[i].ReleaseTime)
		tj, ej := time.Parse(time.RFC3339, atlas.Versions[j].ReleaseTime)
		if ei == nil && ej == nil && !ti.Equal(tj) {
			return ti.After(tj)
		}
		return compareVersionStrings(atlas.Versions[i].ID, atlas.Versions[j].ID) > 0
	})
	return nil
}

func atlasEvidence(seed atlasSeedSummary) []AtlasEvidenceSource {
	return []AtlasEvidenceSource{
		{ID: "mojang-version-manifest-v2", Name: "Mojang version manifest v2 and per-version manifests", URL: "https://piston-meta.mojang.com/mc/game/version_manifest_v2.json", SeedPath: atlasMojangDetailsPath, SHA256: seed.GzipSHA, Records: seed.Count, Retrieved: atlasReviewedAt},
		{ID: "mcmeta-versions", Name: "Misode mcmeta version catalog", URL: "https://github.com/misode/mcmeta", SeedPath: atlasMCMetaPath, Records: 450, Retrieved: atlasReviewedAt},
		{ID: "fabric-meta", Name: "Fabric Meta", URL: "https://meta.fabricmc.net/", SeedPath: "assets/seeds/fabric-game-versions.json", Records: 519, Retrieved: atlasReviewedAt},
		{ID: "quilt-meta", Name: "Quilt Meta", URL: "https://meta.quiltmc.org/", SeedPath: "assets/seeds/quilt-game-versions.json", Records: 430, Retrieved: atlasReviewedAt},
		{ID: "modrinth-tags", Name: "Modrinth version and loader tags", URL: "https://docs.modrinth.com/api/operations/gettaggameversion/", SeedPath: "assets/seeds/modrinth-game-versions.json", Records: 907, Retrieved: atlasReviewedAt},
		{ID: "official-maven-metadata", Name: "Official loader, mappings, and build-plugin Maven metadata", URL: "https://maven.fabricmc.net/", SeedPath: "assets/seeds/", Records: len(atlasMavenDefinitions), Retrieved: atlasReviewedAt},
	}
}

func (a *App) handleAtlasSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	atlas, err := loadRepairVersionAtlas()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, atlas.Summary)
}

func (a *App) handleAtlasVersions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	atlas, err := loadRepairVersionAtlas()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
		return
	}
	query := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	kind := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("type")))
	loader := normalizeDoctorLoader(r.URL.Query().Get("loader"))
	limit := clampInt(parseQueryInt(r, "limit", 80), 1, 200)
	offset := atlasMaxInt(parseQueryInt(r, "offset", 0), 0)
	filtered := make([]AtlasVersion, 0, len(atlas.Versions))
	for _, version := range atlas.Versions {
		if query != "" && !strings.Contains(strings.ToLower(version.ID+" "+version.Type), query) {
			continue
		}
		if kind != "" && kind != "all" && !strings.EqualFold(version.Type, kind) {
			continue
		}
		switch loader {
		case "fabric":
			if !version.FabricSupported {
				continue
			}
		case "quilt":
			if !version.QuiltSupported {
				continue
			}
		}
		filtered = append(filtered, version)
	}
	if offset > len(filtered) {
		offset = len(filtered)
	}
	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}
	writeJSON(w, http.StatusOK, AtlasVersionQueryResponse{Total: len(filtered), Offset: offset, Limit: limit, Versions: filtered[offset:end]})
}

func (a *App) handleAtlasToolchains(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	atlas, err := loadRepairVersionAtlas()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
		return
	}
	query := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	items := make([]AtlasMavenArtifact, 0, len(atlas.Maven))
	for _, artifact := range atlas.Maven {
		if query != "" {
			haystack := strings.ToLower(strings.Join([]string{artifact.ID, artifact.Name, artifact.GroupID, artifact.ArtifactID, artifact.Latest, artifact.OfficialURL}, " "))
			if !strings.Contains(haystack, query) {
				continue
			}
		}
		copy := artifact
		copy.Versions = nil
		items = append(items, copy)
	}
	sort.Slice(items, func(i, j int) bool { return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name) })
	writeJSON(w, http.StatusOK, map[string]any{"reviewedAt": atlasReviewedAt, "total": len(items), "toolchains": items})
}

func (a *App) handleAtlasResolve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request AtlasResolveRequest
	if err := decodeJSON(r, &request); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	request.GameVersion = strings.TrimSpace(request.GameVersion)
	request.Loader = normalizeDoctorLoader(request.Loader)
	if request.GameVersion == "" || request.Loader == "" {
		writeJSON(w, http.StatusBadRequest, APIError{Error: "gameVersion and loader are required"})
		return
	}
	atlas, err := loadRepairVersionAtlas()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
		return
	}
	resolution := atlas.Resolve(request.GameVersion, request.Loader)
	writeJSON(w, http.StatusOK, resolution)
}

func (atlas *RepairVersionAtlas) Resolve(gameVersion, loader string) AtlasResolution {
	loader = normalizeDoctorLoader(loader)
	resolution := AtlasResolution{
		ResolvedAt: time.Now().UTC().Format(time.RFC3339), ReviewedAt: atlasReviewedAt,
		GameVersion: gameVersion, Loader: loader, JavaMajor: targetJavaForMinecraft(gameVersion), Evidence: atlas.Summary.EvidenceSources,
	}
	if version, ok := atlas.ByID[gameVersion]; ok {
		copy := version
		resolution.Exists = true
		resolution.Version = &copy
		if version.JavaMajor > 0 {
			resolution.JavaMajor = version.JavaMajor
		}
	} else {
		resolution.Warnings = append(resolution.Warnings, "The exact Minecraft version is not present in the embedded Mojang manifest snapshot; execution must refresh authoritative metadata before downloading or patching anything.")
	}

	addBuild := func(id, reason string, preferStable bool) {
		artifact, ok := atlas.Maven[id]
		if !ok {
			return
		}
		version := artifact.Latest
		channel := "preview"
		if preferStable && artifact.LatestStable != "" {
			version = artifact.LatestStable
			channel = "stable"
		} else if !isPrereleaseVersion(version) {
			channel = "stable"
		}
		resolution.BuildToolchains = append(resolution.BuildToolchains, AtlasToolchainChoice{ID: id, Name: artifact.Name, Version: version, Channel: channel, Reason: reason, OfficialURL: artifact.OfficialURL, SourcePath: artifact.SourcePath})
	}
	addMapping := func(id, version, reason string) {
		artifact, ok := atlas.Maven[id]
		if !ok || version == "" {
			return
		}
		resolution.Mappings = append(resolution.Mappings, AtlasToolchainChoice{ID: id, Name: artifact.Name, Version: version, Channel: channelForVersion(version), Reason: reason, OfficialURL: artifact.OfficialURL, SourcePath: artifact.SourcePath})
	}
	addGameArtifact := func(id, version, reason string) {
		artifact, ok := atlas.Maven[id]
		if !ok || version == "" {
			return
		}
		resolution.GameArtifacts = append(resolution.GameArtifacts, AtlasToolchainChoice{ID: id, Name: artifact.Name, Version: version, Channel: channelForVersion(version), Reason: reason, OfficialURL: artifact.OfficialURL, SourcePath: artifact.SourcePath})
	}

	switch loader {
	case "fabric":
		resolution.LoaderSupported = atlas.FabricGames[gameVersion]
		resolution.LoaderVersion, resolution.LoaderChannel = latestLoader(atlas.FabricLoaders)
		addGameArtifact("fabric-api", latestMatchingVersion(atlas.Maven["fabric-api"].Versions, func(v string) bool { return versionSuffixMatches(v, gameVersion) }, true), "Latest Fabric API build whose published suffix matches the requested Minecraft version.")
		addBuild("fabric-loom", "Primary Fabric Gradle plugin. The stable lane is selected for execution; the newest preview remains visible in the atlas.", true)
		addBuild("architectury-loom", "Optional multi-loader Loom fork for Architectury workspaces.", true)
		addMapping("intermediary", latestMatchingVersion(atlas.Maven["intermediary"].Versions, func(v string) bool { return v == gameVersion }, true), "Exact Fabric intermediary namespace for this Minecraft version.")
		addMapping("yarn", latestMatchingVersion(atlas.Maven["yarn"].Versions, func(v string) bool {
			return strings.HasPrefix(v, gameVersion+"+") || strings.HasPrefix(v, gameVersion+".")
		}, true), "Highest published Yarn build matching the requested Minecraft version.")
		if !resolution.LoaderSupported {
			resolution.Warnings = append(resolution.Warnings, "Fabric Meta does not currently list this exact game version. Keep the plan review-only until refreshed metadata proves support.")
		}
	case "quilt":
		resolution.LoaderSupported = atlas.QuiltGames[gameVersion]
		resolution.LoaderVersion, resolution.LoaderChannel = latestLoader(atlas.QuiltLoaders)
		addBuild("fabric-loom", "Quilt projects commonly use Quilt Loom/Fabric Loom lineage; verify the project's exact plugin before mutation.", true)
		addMapping("quilt-mappings", latestMatchingVersion(atlas.Maven["quilt-mappings"].Versions, func(v string) bool { return strings.HasPrefix(v, gameVersion+"+") }, true), "Highest Quilt mappings build matching the requested Minecraft version.")
		addMapping("intermediary", latestMatchingVersion(atlas.Maven["intermediary"].Versions, func(v string) bool { return v == gameVersion }, true), "Intermediary namespace used by Quilt-compatible tooling where available.")
		if !resolution.LoaderSupported {
			resolution.Warnings = append(resolution.Warnings, "Quilt Meta does not currently list this exact game version. Keep execution review-only until refreshed metadata proves support.")
		}
	case "forge":
		forgeVersion := latestMatchingVersion(atlas.Maven["forge"].Versions, func(v string) bool { return strings.HasPrefix(v, gameVersion+"-") }, true)
		resolution.LoaderSupported = forgeVersion != ""
		resolution.LoaderVersion = forgeVersion
		resolution.LoaderChannel = channelForVersion(forgeVersion)
		addGameArtifact("forge", forgeVersion, "Highest official Forge coordinate matching the exact Minecraft version prefix.")
		key := parseGameVersionKey(gameVersion)
		if key.Valid && key.Era == 1 && key.Parts[1] <= 12 {
			resolution.CompatibilityRoutes = append(resolution.CompatibilityRoutes,
				AtlasToolchainChoice{ID: "retrofuturagradle", Name: "RetroFuturaGradle", Channel: "specialized", Reason: "Legacy Forge source reconstruction and reproducible builds require era-specific tooling rather than modern ForgeGradle assumptions.", OfficialURL: "https://github.com/GTNewHorizons/RetroFuturaGradle"},
				AtlasToolchainChoice{ID: "retromcp-java", Name: "RetroMCP-Java", Channel: "specialized", Reason: "Recover and compare historical Minecraft source/mappings before semantic porting.", OfficialURL: "https://github.com/MCPHackers/RetroMCP-Java"},
			)
			resolution.Warnings = append(resolution.Warnings, "Legacy Forge execution is routed through the dedicated recovery stack. A modern ForgeGradle version must not be written into this workspace automatically.")
		} else {
			major := preferredForgeGradleMajor(gameVersion)
			version := latestMatchingVersion(atlas.Maven["forgegradle"].Versions, func(v string) bool { return strings.HasPrefix(v, major+".") }, true)
			artifact := atlas.Maven["forgegradle"]
			resolution.BuildToolchains = append(resolution.BuildToolchains, AtlasToolchainChoice{ID: "forgegradle", Name: artifact.Name, Version: version, Channel: channelForVersion(version), Reason: "ForgeGradle major selected from the target Minecraft era, not merely the globally newest plugin.", OfficialURL: artifact.OfficialURL, SourcePath: artifact.SourcePath})
			addMapping("mcpconfig", latestMatchingVersion(atlas.Maven["mcpconfig"].Versions, func(v string) bool { return strings.HasPrefix(v, gameVersion+"-") }, true), "Highest MCPConfig mapping snapshot matching the requested Minecraft version.")
		}
		if !resolution.LoaderSupported {
			resolution.Warnings = append(resolution.Warnings, "No official Forge Maven coordinate matched this exact Minecraft version in the embedded snapshot.")
		}
	case "neoforge":
		prefix := neoForgeGamePrefix(gameVersion)
		neoVersion := latestMatchingVersion(atlas.Maven["neoforge"].Versions, func(v string) bool { return strings.HasPrefix(v, prefix+".") || v == prefix }, true)
		resolution.LoaderSupported = neoVersion != ""
		resolution.LoaderVersion = neoVersion
		resolution.LoaderChannel = channelForVersion(neoVersion)
		addGameArtifact("neoforge", neoVersion, "Highest official NeoForge coordinate matching the target game's NeoForge release line.")
		addBuild("moddevgradle", "Current first-party NeoForge development plugin and preferred default for new or migrated source workspaces.", true)
		addBuild("neogradle", "Legacy/alternative NeoForge userdev path retained for existing workspaces that already depend on it.", true)
		addMapping("neoform", latestMatchingVersion(atlas.Maven["neoform"].Versions, func(v string) bool { return strings.HasPrefix(v, gameVersion+"-") || strings.HasPrefix(v, prefix+"-") }, true), "Highest NeoForm coordinate matching the requested game line.")
		if !resolution.LoaderSupported {
			resolution.Warnings = append(resolution.Warnings, "No official NeoForge Maven coordinate matched this target line in the embedded snapshot.")
		}
	case "vanilla":
		resolution.LoaderSupported = resolution.Exists
		resolution.LoaderVersion = gameVersion
		resolution.LoaderChannel = "official"
		resolution.GameArtifacts = append(resolution.GameArtifacts, AtlasToolchainChoice{ID: "mojang-client-server", Name: "Official Mojang client/server artifacts", Version: gameVersion, Channel: "official", Reason: "Exact manifest hashes and download identities are retained in the atlas.", OfficialURL: "https://piston-meta.mojang.com/mc/game/version_manifest_v2.json", SourcePath: atlasMojangDetailsPath})
	default:
		resolution.Warnings = append(resolution.Warnings, "This loader is not yet an executable atlas lane. Mod Doctor can still produce a research-backed manual plan.")
	}

	if resolution.JavaMajor == 0 {
		resolution.JavaMajor = targetJavaForMinecraft(gameVersion)
	}
	if resolution.JavaMajor == 0 {
		resolution.Warnings = append(resolution.Warnings, "Java runtime could not be proven from the embedded Mojang manifest; do not execute a build until a runtime is selected explicitly.")
	}
	return resolution
}

func latestLoader(rows []AtlasLoaderVersion) (string, string) {
	for _, row := range rows {
		if row.Stable && row.Version != "" {
			return row.Version, "stable"
		}
	}
	for _, row := range rows {
		if row.Version != "" && !isPrereleaseVersion(row.Version) {
			return row.Version, "stable"
		}
	}
	if len(rows) > 0 {
		return rows[0].Version, channelForVersion(rows[0].Version)
	}
	return "", "unknown"
}

func preferredForgeGradleMajor(gameVersion string) string {
	key := parseGameVersionKey(gameVersion)
	if !key.Valid {
		return "6"
	}
	minor := key.Parts[1]
	switch {
	case key.Era == 2 || minor >= 20:
		return "6"
	case minor >= 18:
		return "5"
	case minor >= 13:
		return "3"
	default:
		return "2"
	}
}

func neoForgeGamePrefix(gameVersion string) string {
	trimmed := strings.TrimSpace(gameVersion)
	if strings.HasPrefix(trimmed, "1.") {
		parts := strings.Split(trimmed, ".")
		if len(parts) >= 3 {
			return parts[1] + "." + parts[2]
		}
		if len(parts) >= 2 {
			return parts[1] + ".0"
		}
	}
	parts := strings.Split(trimmed, ".")
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	return trimmed
}

func versionSuffixMatches(version, gameVersion string) bool {
	lower := strings.ToLower(version)
	return strings.HasSuffix(lower, "+"+strings.ToLower(gameVersion)) || strings.Contains(lower, "+"+strings.ToLower(gameVersion)+".")
}

func latestMatchingVersion(versions []string, predicate func(string) bool, allowPrerelease bool) string {
	matches := make([]string, 0)
	for _, version := range versions {
		version = strings.TrimSpace(version)
		if version == "" || !predicate(version) {
			continue
		}
		if !allowPrerelease && isPrereleaseVersion(version) {
			continue
		}
		matches = append(matches, version)
	}
	if len(matches) == 0 && !allowPrerelease {
		return latestMatchingVersion(versions, predicate, true)
	}
	sort.SliceStable(matches, func(i, j int) bool { return compareVersionStrings(matches[i], matches[j]) > 0 })
	if len(matches) == 0 {
		return ""
	}
	return matches[0]
}

func channelForVersion(version string) string {
	if version == "" {
		return "unknown"
	}
	if isPrereleaseVersion(version) {
		return "preview"
	}
	return "stable"
}

func isPrereleaseVersion(version string) bool {
	lower := strings.ToLower(version)
	return strings.Contains(lower, "snapshot") || strings.Contains(lower, "alpha") || strings.Contains(lower, "beta") || strings.Contains(lower, "-rc") || strings.Contains(lower, ".rc") || strings.Contains(lower, "pre")
}

func compareVersionStrings(left, right string) int {
	lt := versionTokens(left)
	rt := versionTokens(right)
	max := len(lt)
	if len(rt) > max {
		max = len(rt)
	}
	for i := 0; i < max; i++ {
		if i >= len(lt) {
			return -1
		}
		if i >= len(rt) {
			return 1
		}
		l, r := lt[i], rt[i]
		if l.numeric && r.numeric {
			if l.number > r.number {
				return 1
			}
			if l.number < r.number {
				return -1
			}
			continue
		}
		if l.numeric != r.numeric {
			if l.numeric {
				return 1
			}
			return -1
		}
		ls, rs := strings.ToLower(l.text), strings.ToLower(r.text)
		if ls > rs {
			return 1
		}
		if ls < rs {
			return -1
		}
	}
	return 0
}

type versionToken struct {
	numeric bool
	number  int64
	text    string
}

func versionTokens(value string) []versionToken {
	var out []versionToken
	var current []rune
	currentNumeric := false
	flush := func() {
		if len(current) == 0 {
			return
		}
		text := string(current)
		if currentNumeric {
			n, _ := strconv.ParseInt(text, 10, 64)
			out = append(out, versionToken{numeric: true, number: n, text: text})
		} else {
			out = append(out, versionToken{text: text})
		}
		current = nil
	}
	for _, r := range value {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			flush()
			continue
		}
		numeric := unicode.IsDigit(r)
		if len(current) > 0 && numeric != currentNumeric {
			flush()
		}
		currentNumeric = numeric
		current = append(current, r)
	}
	flush()
	return out
}

func parseQueryInt(r *http.Request, name string, fallback int) int {
	value := strings.TrimSpace(r.URL.Query().Get(name))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func clampInt(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func atlasMaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func readAllLimited(reader io.Reader, limit int64) ([]byte, error) {
	limited := io.LimitReader(reader, limit+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("input exceeds %d bytes", limit)
	}
	return data, nil
}
