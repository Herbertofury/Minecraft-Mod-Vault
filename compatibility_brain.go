package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	_ "modernc.org/sqlite"
)

const compatibilityBrainSchemaVersion = 1

type CompatibilityBrain struct {
	db   *sql.DB
	path string
}

type BrainStatus struct {
	Ready                 bool   `json:"ready"`
	DatabasePath          string `json:"databasePath"`
	SchemaVersion         int    `json:"schemaVersion"`
	SQLiteVersion         string `json:"sqliteVersion"`
	JournalMode           string `json:"journalMode"`
	SeedDigest            string `json:"seedDigest"`
	ImportedAt            string `json:"importedAt"`
	MinecraftVersions     int    `json:"minecraftVersions"`
	LoaderGameRows        int    `json:"loaderGameRows"`
	LoaderReleases        int    `json:"loaderReleases"`
	ToolchainReleases     int    `json:"toolchainReleases"`
	KnowledgeDocuments    int    `json:"knowledgeDocuments"`
	RepairRecords         int    `json:"repairRecords"`
	OldestMinecraft       string `json:"oldestMinecraft,omitempty"`
	NewestMinecraft       string `json:"newestMinecraft,omitempty"`
	LatestStableMinecraft string `json:"latestStableMinecraft,omitempty"`
}

type BrainSearchResult struct {
	ID         string  `json:"id"`
	Kind       string  `json:"kind"`
	Name       string  `json:"name"`
	Category   string  `json:"category,omitempty"`
	Summary    string  `json:"summary,omitempty"`
	URL        string  `json:"url,omitempty"`
	Repository string  `json:"repository,omitempty"`
	Maturity   string  `json:"maturity,omitempty"`
	Status     string  `json:"status,omitempty"`
	Snippet    string  `json:"snippet,omitempty"`
	Rank       float64 `json:"rank"`
}

type MinecraftVersionIntelligence struct {
	ID                    string `json:"id"`
	VersionType           string `json:"versionType,omitempty"`
	Stable                bool   `json:"stable"`
	ReleaseTime           string `json:"releaseTime,omitempty"`
	ReleaseTarget         string `json:"releaseTarget,omitempty"`
	DataVersion           int    `json:"dataVersion,omitempty"`
	ProtocolVersion       int    `json:"protocolVersion,omitempty"`
	DataPackVersion       int    `json:"dataPackVersion,omitempty"`
	DataPackVersionMinor  int    `json:"dataPackVersionMinor,omitempty"`
	ResourcePackVersion   int    `json:"resourcePackVersion,omitempty"`
	ResourcePackMinor     int    `json:"resourcePackVersionMinor,omitempty"`
	JavaMajor             int    `json:"javaMajor,omitempty"`
	JavaComponent         string `json:"javaComponent,omitempty"`
	HasClient             bool   `json:"hasClient"`
	HasServer             bool   `json:"hasServer"`
	HasClientMappings     bool   `json:"hasClientMappings"`
	HasServerMappings     bool   `json:"hasServerMappings"`
	ClientSHA1            string `json:"clientSha1,omitempty"`
	ServerSHA1            string `json:"serverSha1,omitempty"`
	ManifestSHA1          string `json:"manifestSha1,omitempty"`
	RuntimeLibraryCount   int    `json:"runtimeLibraryCount,omitempty"`
	Available             bool   `json:"available"`
	OfficialManifestKnown bool   `json:"officialManifestKnown"`
}

type BrainRelease struct {
	Tool     string `json:"tool"`
	Version  string `json:"version"`
	Stable   bool   `json:"stable"`
	Maven    string `json:"maven,omitempty"`
	SHA256   string `json:"sha256,omitempty"`
	FileSize int64  `json:"fileSize,omitempty"`
	Source   string `json:"source,omitempty"`
	Latest   bool   `json:"latest"`
	Release  bool   `json:"release"`
}

type knowledgeDocument struct {
	ID         string
	Kind       string
	Name       string
	Category   string
	Summary    string
	Body       string
	URL        string
	Repository string
	Maturity   string
	Status     string
	ReviewedAt string
}

type mavenMetadata struct {
	GroupID    string `xml:"groupId"`
	ArtifactID string `xml:"artifactId"`
	Versioning struct {
		Latest      string   `xml:"latest"`
		Release     string   `xml:"release"`
		Versions    []string `xml:"versions>version"`
		LastUpdated string   `xml:"lastUpdated"`
	} `xml:"versioning"`
}

type deepMinecraftRecord struct {
	ID       string `json:"id"`
	Manifest struct {
		ID          string `json:"id"`
		Type        string `json:"type"`
		Time        string `json:"time"`
		ReleaseTime string `json:"releaseTime"`
		JavaVersion struct {
			Component    string `json:"component"`
			MajorVersion int    `json:"majorVersion"`
		} `json:"javaVersion"`
		Downloads map[string]struct {
			SHA1 string `json:"sha1"`
			Size int64  `json:"size"`
			URL  string `json:"url"`
		} `json:"downloads"`
		Libraries []json.RawMessage `json:"libraries"`
	} `json:"manifest"`
}

type mcmetaVersion struct {
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	ReleaseTarget        string `json:"release_target"`
	Type                 string `json:"type"`
	Stable               bool   `json:"stable"`
	DataVersion          int    `json:"data_version"`
	ProtocolVersion      int    `json:"protocol_version"`
	DataPackVersion      int    `json:"data_pack_version"`
	DataPackVersionMinor int    `json:"data_pack_version_minor"`
	ResourcePackVersion  int    `json:"resource_pack_version"`
	ResourcePackMinor    int    `json:"resource_pack_version_minor"`
	BuildTime            string `json:"build_time"`
	ReleaseTime          string `json:"release_time"`
	SHA1                 string `json:"sha1"`
}

type loaderGameVersion struct {
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
}

type loaderReleaseRecord struct {
	Maven    string            `json:"maven"`
	Version  string            `json:"version"`
	Build    int               `json:"build"`
	Stable   bool              `json:"stable"`
	FileSize int64             `json:"file_size"`
	Hashes   map[string]string `json:"hashes"`
}

type modrinthGameVersion struct {
	Version     string `json:"version"`
	VersionType string `json:"version_type"`
	Date        string `json:"date"`
	Major       bool   `json:"major"`
}

type modrinthLoader struct {
	Name                  string   `json:"name"`
	SupportedProjectTypes []string `json:"supported_project_types"`
}

func openCompatibilityBrain(cfgDir string) (*CompatibilityBrain, error) {
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		return nil, err
	}
	path := filepath.Join(cfgDir, "compatibility-brain.sqlite3")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(2)
	brain := &CompatibilityBrain{db: db, path: path}
	if err := brain.initialize(); err != nil {
		db.Close()
		return nil, err
	}
	return brain, nil
}

func (b *CompatibilityBrain) Close() error {
	if b == nil || b.db == nil {
		return nil
	}
	return b.db.Close()
}

func (b *CompatibilityBrain) initialize() error {
	for _, statement := range []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA foreign_keys=ON",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA busy_timeout=5000",
	} {
		if _, err := b.db.Exec(statement); err != nil {
			return fmt.Errorf("compatibility brain pragma %q: %w", statement, err)
		}
	}
	if err := b.ensureSchema(); err != nil {
		return err
	}
	digest, err := compatibilitySeedDigest()
	if err != nil {
		return err
	}
	stored, _ := b.meta("seed_digest")
	schema, _ := b.meta("schema_version")
	if stored != digest || schema != strconv.Itoa(compatibilityBrainSchemaVersion) {
		if err := b.importEmbeddedKnowledge(digest); err != nil {
			return err
		}
	}
	return nil
}

func (b *CompatibilityBrain) ensureSchema() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS brain_meta (key TEXT PRIMARY KEY, value TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS minecraft_versions (
			id TEXT PRIMARY KEY,
			version_type TEXT,
			stable INTEGER NOT NULL DEFAULT 0,
			release_time TEXT,
			release_target TEXT,
			data_version INTEGER NOT NULL DEFAULT 0,
			protocol_version INTEGER NOT NULL DEFAULT 0,
			data_pack_version INTEGER NOT NULL DEFAULT 0,
			data_pack_minor INTEGER NOT NULL DEFAULT 0,
			resource_pack_version INTEGER NOT NULL DEFAULT 0,
			resource_pack_minor INTEGER NOT NULL DEFAULT 0,
			java_major INTEGER NOT NULL DEFAULT 0,
			java_component TEXT,
			has_client INTEGER NOT NULL DEFAULT 0,
			has_server INTEGER NOT NULL DEFAULT 0,
			has_client_mappings INTEGER NOT NULL DEFAULT 0,
			has_server_mappings INTEGER NOT NULL DEFAULT 0,
			client_sha1 TEXT,
			server_sha1 TEXT,
			manifest_sha1 TEXT,
			library_count INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS loader_game_support (
			loader TEXT NOT NULL,
			game_version TEXT NOT NULL,
			stable INTEGER NOT NULL DEFAULT 0,
			ordinal INTEGER NOT NULL DEFAULT 0,
			PRIMARY KEY(loader, game_version)
		)`,
		`CREATE TABLE IF NOT EXISTS loader_releases (
			loader TEXT NOT NULL,
			version TEXT NOT NULL,
			stable INTEGER NOT NULL DEFAULT 0,
			maven TEXT,
			file_size INTEGER NOT NULL DEFAULT 0,
			sha256 TEXT,
			ordinal INTEGER NOT NULL DEFAULT 0,
			PRIMARY KEY(loader, version)
		)`,
		`CREATE TABLE IF NOT EXISTS loader_catalog (
			name TEXT PRIMARY KEY,
			supported_project_types TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS toolchain_releases (
			tool TEXT NOT NULL,
			version TEXT NOT NULL,
			is_latest INTEGER NOT NULL DEFAULT 0,
			is_release INTEGER NOT NULL DEFAULT 0,
			ordinal INTEGER NOT NULL DEFAULT 0,
			source_file TEXT NOT NULL,
			last_updated TEXT,
			PRIMARY KEY(tool, version)
		)`,
		`CREATE TABLE IF NOT EXISTS knowledge_documents (
			id TEXT PRIMARY KEY,
			kind TEXT NOT NULL,
			name TEXT NOT NULL,
			category TEXT,
			summary TEXT,
			body TEXT,
			url TEXT,
			repository TEXT,
			maturity TEXT,
			status TEXT,
			reviewed_at TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS repair_records (
			id TEXT PRIMARY KEY,
			recorded_at TEXT,
			outcome TEXT,
			confidence TEXT,
			root_cause TEXT,
			body TEXT NOT NULL
		)`,
		`CREATE VIRTUAL TABLE IF NOT EXISTS knowledge_fts USING fts5(
			id UNINDEXED,
			kind,
			name,
			category,
			summary,
			body,
			tokenize='unicode61 remove_diacritics 2'
		)`,
		`CREATE INDEX IF NOT EXISTS idx_minecraft_release_time ON minecraft_versions(release_time DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_loader_game ON loader_game_support(loader, ordinal)`,
		`CREATE INDEX IF NOT EXISTS idx_loader_release ON loader_releases(loader, stable DESC, ordinal)`,
		`CREATE INDEX IF NOT EXISTS idx_tool_release ON toolchain_releases(tool, is_latest DESC, is_release DESC, ordinal)`,
		`CREATE INDEX IF NOT EXISTS idx_knowledge_kind ON knowledge_documents(kind, category)`,
	}
	for _, statement := range statements {
		if _, err := b.db.Exec(statement); err != nil {
			return fmt.Errorf("compatibility brain schema: %w", err)
		}
	}
	return nil
}

func compatibilitySeedDigest() (string, error) {
	paths := []string{
		"assets/mod-doctor-knowledge.json",
		"assets/mod-doctor-repair-patterns.json",
		"knowledge/doctor-tools.json",
		"repair-brain/repair-history.jsonl",
	}
	entries, err := embeddedFiles.ReadDir("assets/seeds")
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			paths = append(paths, "assets/seeds/"+entry.Name())
		}
	}
	sort.Strings(paths)
	h := sha256.New()
	for _, path := range paths {
		data, err := embeddedFiles.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("seed digest %s: %w", path, err)
		}
		_, _ = io.WriteString(h, path)
		_, _ = h.Write([]byte{0})
		_, _ = h.Write(data)
		_, _ = h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func (b *CompatibilityBrain) importEmbeddedKnowledge(digest string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	tx, err := b.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, table := range []string{"minecraft_versions", "loader_game_support", "loader_releases", "loader_catalog", "toolchain_releases", "knowledge_documents", "repair_records", "knowledge_fts"} {
		if _, err := tx.ExecContext(ctx, "DELETE FROM "+table); err != nil {
			return fmt.Errorf("clear %s: %w", table, err)
		}
	}
	if err := importDeepMinecraftSeed(ctx, tx); err != nil {
		return err
	}
	if err := importMcmetaSeed(ctx, tx); err != nil {
		return err
	}
	if err := importModrinthGameVersions(ctx, tx); err != nil {
		return err
	}
	if err := importLoaderSeeds(ctx, tx); err != nil {
		return err
	}
	if err := importToolchainSeeds(ctx, tx); err != nil {
		return err
	}
	if err := importKnowledgeDocuments(ctx, tx); err != nil {
		return err
	}
	if err := importRepairHistory(ctx, tx); err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	for key, value := range map[string]string{
		"schema_version": strconv.Itoa(compatibilityBrainSchemaVersion),
		"seed_digest":    digest,
		"imported_at":    now,
	} {
		if _, err := tx.ExecContext(ctx, `INSERT INTO brain_meta(key,value) VALUES(?,?) ON CONFLICT(key) DO UPDATE SET value=excluded.value`, key, value); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	if _, err := b.db.Exec("PRAGMA optimize"); err != nil {
		return err
	}
	return nil
}

func importDeepMinecraftSeed(ctx context.Context, tx *sql.Tx) error {
	data, err := embeddedFiles.ReadFile("assets/seeds/mojang-version-details-v2.jsonl.gz")
	if err != nil {
		return err
	}
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer gz.Close()
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO minecraft_versions(
		id, version_type, release_time, java_major, java_component,
		has_client, has_server, has_client_mappings, has_server_mappings,
		client_sha1, server_sha1, library_count
	) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)
	ON CONFLICT(id) DO UPDATE SET
		version_type=excluded.version_type,
		release_time=excluded.release_time,
		java_major=excluded.java_major,
		java_component=excluded.java_component,
		has_client=excluded.has_client,
		has_server=excluded.has_server,
		has_client_mappings=excluded.has_client_mappings,
		has_server_mappings=excluded.has_server_mappings,
		client_sha1=excluded.client_sha1,
		server_sha1=excluded.server_sha1,
		library_count=excluded.library_count`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	scanner := bufio.NewScanner(gz)
	scanner.Buffer(make([]byte, 128*1024), 16<<20)
	for scanner.Scan() {
		var record deepMinecraftRecord
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			return fmt.Errorf("parse Mojang deep seed: %w", err)
		}
		id := firstNonEmpty(record.ID, record.Manifest.ID)
		if id == "" {
			continue
		}
		client, hasClient := record.Manifest.Downloads["client"]
		server, hasServer := record.Manifest.Downloads["server"]
		_, hasClientMappings := record.Manifest.Downloads["client_mappings"]
		_, hasServerMappings := record.Manifest.Downloads["server_mappings"]
		if _, err := stmt.ExecContext(ctx,
			id,
			record.Manifest.Type,
			record.Manifest.ReleaseTime,
			record.Manifest.JavaVersion.MajorVersion,
			record.Manifest.JavaVersion.Component,
			boolInt(hasClient), boolInt(hasServer), boolInt(hasClientMappings), boolInt(hasServerMappings),
			client.SHA1, server.SHA1, len(record.Manifest.Libraries),
		); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func importMcmetaSeed(ctx context.Context, tx *sql.Tx) error {
	data, err := embeddedFiles.ReadFile("assets/seeds/mcmeta-versions.json")
	if err != nil {
		return err
	}
	var versions []mcmetaVersion
	if err := json.Unmarshal(data, &versions); err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO minecraft_versions(
		id, version_type, stable, release_time, release_target,
		data_version, protocol_version, data_pack_version, data_pack_minor,
		resource_pack_version, resource_pack_minor, manifest_sha1
	) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)
	ON CONFLICT(id) DO UPDATE SET
		version_type=CASE WHEN excluded.version_type<>'' THEN excluded.version_type ELSE minecraft_versions.version_type END,
		stable=excluded.stable,
		release_time=CASE WHEN excluded.release_time<>'' THEN excluded.release_time ELSE minecraft_versions.release_time END,
		release_target=excluded.release_target,
		data_version=excluded.data_version,
		protocol_version=excluded.protocol_version,
		data_pack_version=excluded.data_pack_version,
		data_pack_minor=excluded.data_pack_minor,
		resource_pack_version=excluded.resource_pack_version,
		resource_pack_minor=excluded.resource_pack_minor,
		manifest_sha1=excluded.manifest_sha1`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, version := range versions {
		if _, err := stmt.ExecContext(ctx,
			version.ID, version.Type, boolInt(version.Stable), version.ReleaseTime, version.ReleaseTarget,
			version.DataVersion, version.ProtocolVersion, version.DataPackVersion, version.DataPackVersionMinor,
			version.ResourcePackVersion, version.ResourcePackMinor, version.SHA1,
		); err != nil {
			return err
		}
	}
	return nil
}

func importModrinthGameVersions(ctx context.Context, tx *sql.Tx) error {
	data, err := embeddedFiles.ReadFile("assets/seeds/modrinth-game-versions.json")
	if err != nil {
		return err
	}
	var versions []modrinthGameVersion
	if err := json.Unmarshal(data, &versions); err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO minecraft_versions(id,version_type,stable,release_time)
		VALUES(?,?,?,?) ON CONFLICT(id) DO UPDATE SET
		version_type=CASE WHEN minecraft_versions.version_type='' OR minecraft_versions.version_type IS NULL THEN excluded.version_type ELSE minecraft_versions.version_type END,
		stable=CASE WHEN excluded.version_type='release' THEN 1 ELSE minecraft_versions.stable END,
		release_time=CASE WHEN minecraft_versions.release_time='' OR minecraft_versions.release_time IS NULL THEN excluded.release_time ELSE minecraft_versions.release_time END`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, version := range versions {
		if _, err := stmt.ExecContext(ctx, version.Version, version.VersionType, boolInt(version.VersionType == "release"), version.Date); err != nil {
			return err
		}
	}
	return nil
}

func importLoaderSeeds(ctx context.Context, tx *sql.Tx) error {
	for _, spec := range []struct {
		Loader string
		Path   string
	}{
		{"fabric", "assets/seeds/fabric-game-versions.json"},
		{"quilt", "assets/seeds/quilt-game-versions.json"},
	} {
		data, err := embeddedFiles.ReadFile(spec.Path)
		if err != nil {
			return err
		}
		var versions []loaderGameVersion
		if err := json.Unmarshal(data, &versions); err != nil {
			return err
		}
		for ordinal, version := range versions {
			if _, err := tx.ExecContext(ctx, `INSERT INTO loader_game_support(loader,game_version,stable,ordinal) VALUES(?,?,?,?)`, spec.Loader, version.Version, boolInt(version.Stable), ordinal); err != nil {
				return err
			}
		}
	}
	for _, spec := range []struct {
		Loader string
		Path   string
	}{
		{"fabric", "assets/seeds/fabric-loader-versions.json"},
		{"quilt", "assets/seeds/quilt-loader-versions.json"},
	} {
		data, err := embeddedFiles.ReadFile(spec.Path)
		if err != nil {
			return err
		}
		var releases []loaderReleaseRecord
		if err := json.Unmarshal(data, &releases); err != nil {
			return err
		}
		for ordinal, release := range releases {
			if _, err := tx.ExecContext(ctx, `INSERT INTO loader_releases(loader,version,stable,maven,file_size,sha256,ordinal) VALUES(?,?,?,?,?,?,?)`,
				spec.Loader, release.Version, boolInt(release.Stable), release.Maven, release.FileSize, release.Hashes["sha256"], ordinal); err != nil {
				return err
			}
		}
	}
	data, err := embeddedFiles.ReadFile("assets/seeds/modrinth-loaders.json")
	if err != nil {
		return err
	}
	var loaders []modrinthLoader
	if err := json.Unmarshal(data, &loaders); err != nil {
		return err
	}
	for _, loader := range loaders {
		joined := strings.Join(loader.SupportedProjectTypes, ",")
		if _, err := tx.ExecContext(ctx, `INSERT INTO loader_catalog(name,supported_project_types) VALUES(?,?)`, loader.Name, joined); err != nil {
			return err
		}
	}
	return nil
}

func importToolchainSeeds(ctx context.Context, tx *sql.Tx) error {
	entries, err := embeddedFiles.ReadDir("assets/seeds")
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), "-maven-metadata.xml") {
			continue
		}
		path := "assets/seeds/" + entry.Name()
		data, err := embeddedFiles.ReadFile(path)
		if err != nil {
			return err
		}
		var metadata mavenMetadata
		if err := xml.Unmarshal(data, &metadata); err != nil {
			return fmt.Errorf("parse %s: %w", entry.Name(), err)
		}
		tool := strings.TrimSuffix(entry.Name(), "-maven-metadata.xml")
		for ordinal, version := range metadata.Versioning.Versions {
			if _, err := tx.ExecContext(ctx, `INSERT INTO toolchain_releases(tool,version,is_latest,is_release,ordinal,source_file,last_updated) VALUES(?,?,?,?,?,?,?)`,
				tool, version, boolInt(version == metadata.Versioning.Latest), boolInt(version == metadata.Versioning.Release), ordinal, entry.Name(), metadata.Versioning.LastUpdated); err != nil {
				return err
			}
		}
	}
	return nil
}

func importKnowledgeDocuments(ctx context.Context, tx *sql.Tx) error {
	knowledge, err := loadDoctorKnowledge()
	if err != nil {
		return err
	}
	payload, err := loadDoctorToolsPayload()
	if err != nil {
		return err
	}
	docs := make([]knowledgeDocument, 0, len(knowledge.Sources)+len(payload.Tools)+len(payload.RepairPatterns))
	for _, source := range knowledge.Sources {
		docs = append(docs, knowledgeDocument{
			ID: source.ID, Kind: "source", Name: source.Name, Category: source.Category,
			Summary: source.BestFor, Body: strings.TrimSpace(source.Notes + "\n" + source.Integration),
			URL: source.URL, Repository: source.Repository, Maturity: source.Maturity, Status: source.Status, ReviewedAt: knowledge.ReviewedAt,
		})
	}
	for _, tool := range payload.Tools {
		body := strings.TrimSpace(strings.Join([]string{tool.Capability, tool.BestFor, tool.Integration, tool.CurrentEvidence, strings.Join(tool.Risks, "\n")}, "\n"))
		docs = append(docs, knowledgeDocument{
			ID: "tool:" + tool.ID, Kind: "tool", Name: tool.Name, Category: tool.Category,
			Summary: tool.BestFor, Body: body, URL: tool.OfficialURL, Repository: tool.RepositoryURL,
			Maturity: tool.Maturity, Status: "active", ReviewedAt: payload.ReviewedAt,
		})
	}
	for _, pattern := range payload.RepairPatterns {
		docs = append(docs, knowledgeDocument{
			ID: "pattern:" + pattern.ID, Kind: "repair-pattern", Name: pattern.Name, Category: pattern.FailureClass,
			Summary: pattern.Trigger, Body: strings.TrimSpace(pattern.Repair + "\n" + strings.Join(pattern.Verification, "\n")),
			Maturity: pattern.Confidence, Status: "verified", ReviewedAt: payload.ReviewedAt,
		})
	}
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO knowledge_documents(id,kind,name,category,summary,body,url,repository,maturity,status,reviewed_at) VALUES(?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	fts, err := tx.PrepareContext(ctx, `INSERT INTO knowledge_fts(id,kind,name,category,summary,body) VALUES(?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer fts.Close()
	seen := map[string]bool{}
	for _, doc := range docs {
		if doc.ID == "" || seen[doc.ID] {
			continue
		}
		seen[doc.ID] = true
		if _, err := stmt.ExecContext(ctx, doc.ID, doc.Kind, doc.Name, doc.Category, doc.Summary, doc.Body, doc.URL, doc.Repository, doc.Maturity, doc.Status, doc.ReviewedAt); err != nil {
			return err
		}
		if _, err := fts.ExecContext(ctx, doc.ID, doc.Kind, doc.Name, doc.Category, doc.Summary, doc.Body); err != nil {
			return err
		}
	}
	return nil
}

func importRepairHistory(ctx context.Context, tx *sql.Tx) error {
	data, err := embeddedFiles.ReadFile("repair-brain/repair-history.jsonl")
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 64*1024), 8<<20)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var record struct {
			ID         string `json:"id"`
			RecordedAt string `json:"recorded_at"`
			Outcome    string `json:"outcome"`
			Confidence string `json:"confidence"`
			RootCause  string `json:"root_cause"`
		}
		if err := json.Unmarshal(line, &record); err != nil {
			return err
		}
		if record.ID == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO repair_records(id,recorded_at,outcome,confidence,root_cause,body) VALUES(?,?,?,?,?,?)`,
			record.ID, record.RecordedAt, record.Outcome, record.Confidence, record.RootCause, string(line)); err != nil {
			return err
		}
		doc := knowledgeDocument{
			ID: "repair:" + record.ID, Kind: "repair-history", Name: record.ID, Category: record.Outcome,
			Summary: record.RootCause, Body: string(line), Maturity: record.Confidence, Status: record.Outcome, ReviewedAt: record.RecordedAt,
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO knowledge_documents(id,kind,name,category,summary,body,url,repository,maturity,status,reviewed_at) VALUES(?,?,?,?,?,?,?,?,?,?,?)`,
			doc.ID, doc.Kind, doc.Name, doc.Category, doc.Summary, doc.Body, "", "", doc.Maturity, doc.Status, doc.ReviewedAt); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO knowledge_fts(id,kind,name,category,summary,body) VALUES(?,?,?,?,?,?)`, doc.ID, doc.Kind, doc.Name, doc.Category, doc.Summary, doc.Body); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func (b *CompatibilityBrain) meta(key string) (string, error) {
	var value string
	err := b.db.QueryRow(`SELECT value FROM brain_meta WHERE key=?`, key).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return value, err
}

func (b *CompatibilityBrain) Status() (BrainStatus, error) {
	status := BrainStatus{Ready: b != nil && b.db != nil, DatabasePath: b.path, SchemaVersion: compatibilityBrainSchemaVersion}
	if !status.Ready {
		return status, errors.New("compatibility brain is unavailable")
	}
	if err := b.db.QueryRow(`SELECT sqlite_version()`).Scan(&status.SQLiteVersion); err != nil {
		return status, err
	}
	if err := b.db.QueryRow(`PRAGMA journal_mode`).Scan(&status.JournalMode); err != nil {
		return status, err
	}
	status.SeedDigest, _ = b.meta("seed_digest")
	status.ImportedAt, _ = b.meta("imported_at")
	for query, dst := range map[string]*int{
		`SELECT COUNT(*) FROM minecraft_versions`:  &status.MinecraftVersions,
		`SELECT COUNT(*) FROM loader_game_support`: &status.LoaderGameRows,
		`SELECT COUNT(*) FROM loader_releases`:     &status.LoaderReleases,
		`SELECT COUNT(*) FROM toolchain_releases`:  &status.ToolchainReleases,
		`SELECT COUNT(*) FROM knowledge_documents`: &status.KnowledgeDocuments,
		`SELECT COUNT(*) FROM repair_records`:      &status.RepairRecords,
	} {
		if err := b.db.QueryRow(query).Scan(dst); err != nil {
			return status, err
		}
	}
	_ = b.db.QueryRow(`SELECT id FROM minecraft_versions WHERE release_time IS NOT NULL AND release_time<>'' ORDER BY release_time ASC LIMIT 1`).Scan(&status.OldestMinecraft)
	_ = b.db.QueryRow(`SELECT id FROM minecraft_versions WHERE release_time IS NOT NULL AND release_time<>'' ORDER BY release_time DESC LIMIT 1`).Scan(&status.NewestMinecraft)
	_ = b.db.QueryRow(`SELECT id FROM minecraft_versions WHERE version_type='release' ORDER BY release_time DESC LIMIT 1`).Scan(&status.LatestStableMinecraft)
	return status, nil
}

func (b *CompatibilityBrain) Search(query, kind string, limit int) ([]BrainSearchResult, error) {
	query = strings.TrimSpace(query)
	kind = strings.TrimSpace(strings.ToLower(kind))
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	if query == "" {
		base := `SELECT id,kind,name,category,summary,url,repository,maturity,status,summary,0.0 FROM knowledge_documents`
		args := []any{}
		if kind != "" && kind != "all" {
			base += ` WHERE kind=?`
			args = append(args, kind)
		}
		base += ` ORDER BY kind,name LIMIT ?`
		args = append(args, limit)
		return scanBrainSearchRows(b.db.Query(base, args...))
	}
	match := sanitizeFTSQuery(query)
	if match == "" {
		return nil, nil
	}
	statement := `SELECT d.id,d.kind,d.name,d.category,d.summary,d.url,d.repository,d.maturity,d.status,
		snippet(knowledge_fts,5,'<mark>','</mark>','…',24),bm25(knowledge_fts)
		FROM knowledge_fts JOIN knowledge_documents d ON d.id=knowledge_fts.id
		WHERE knowledge_fts MATCH ?`
	args := []any{match}
	if kind != "" && kind != "all" {
		statement += ` AND d.kind=?`
		args = append(args, kind)
	}
	statement += ` ORDER BY bm25(knowledge_fts), d.name LIMIT ?`
	args = append(args, limit)
	rows, err := b.db.Query(statement, args...)
	if err == nil {
		return scanBrainSearchRows(rows, nil)
	}
	// FTS syntax should already be sanitized, but a LIKE fallback keeps the
	// workbench useful even if a future SQLite tokenizer rejects an edge case.
	like := "%" + strings.ToLower(query) + "%"
	fallback := `SELECT id,kind,name,category,summary,url,repository,maturity,status,body,0.0 FROM knowledge_documents
		WHERE lower(name||' '||category||' '||summary||' '||body) LIKE ?`
	args = []any{like}
	if kind != "" && kind != "all" {
		fallback += ` AND kind=?`
		args = append(args, kind)
	}
	fallback += ` ORDER BY name LIMIT ?`
	args = append(args, limit)
	return scanBrainSearchRows(b.db.Query(fallback, args...))
}

func scanBrainSearchRows(rows *sql.Rows, err error) ([]BrainSearchResult, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []BrainSearchResult
	for rows.Next() {
		var item BrainSearchResult
		if err := rows.Scan(&item.ID, &item.Kind, &item.Name, &item.Category, &item.Summary, &item.URL, &item.Repository, &item.Maturity, &item.Status, &item.Snippet, &item.Rank); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func sanitizeFTSQuery(query string) string {
	var tokens []string
	for _, raw := range strings.Fields(query) {
		var b strings.Builder
		for _, r := range raw {
			if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
				b.WriteRune(unicode.ToLower(r))
			}
		}
		if b.Len() > 0 {
			tokens = append(tokens, b.String()+"*")
		}
	}
	return strings.Join(tokens, " AND ")
}

func (b *CompatibilityBrain) MinecraftVersion(id string) (MinecraftVersionIntelligence, error) {
	var out MinecraftVersionIntelligence
	var stable, hasClient, hasServer, hasClientMappings, hasServerMappings int
	err := b.db.QueryRow(`SELECT id,version_type,stable,release_time,release_target,data_version,protocol_version,
		data_pack_version,data_pack_minor,resource_pack_version,resource_pack_minor,java_major,java_component,
		has_client,has_server,has_client_mappings,has_server_mappings,client_sha1,server_sha1,manifest_sha1,library_count
		FROM minecraft_versions WHERE id=?`, strings.TrimSpace(id)).Scan(
		&out.ID, &out.VersionType, &stable, &out.ReleaseTime, &out.ReleaseTarget, &out.DataVersion, &out.ProtocolVersion,
		&out.DataPackVersion, &out.DataPackVersionMinor, &out.ResourcePackVersion, &out.ResourcePackMinor, &out.JavaMajor, &out.JavaComponent,
		&hasClient, &hasServer, &hasClientMappings, &hasServerMappings, &out.ClientSHA1, &out.ServerSHA1, &out.ManifestSHA1, &out.RuntimeLibraryCount,
	)
	if errors.Is(err, sql.ErrNoRows) {
		out.ID = strings.TrimSpace(id)
		out.JavaMajor = targetJavaForMinecraft(out.ID)
		return out, nil
	}
	if err != nil {
		return out, err
	}
	out.Available = true
	out.OfficialManifestKnown = hasClient != 0 || hasServer != 0
	out.Stable = stable != 0
	out.HasClient = hasClient != 0
	out.HasServer = hasServer != 0
	out.HasClientMappings = hasClientMappings != 0
	out.HasServerMappings = hasServerMappings != 0
	return out, nil
}

func (b *CompatibilityBrain) LoaderSupports(loader, gameVersion string) (bool, bool, error) {
	loader = normalizeDoctorLoader(loader)
	var stable int
	err := b.db.QueryRow(`SELECT stable FROM loader_game_support WHERE loader=? AND game_version=?`, loader, strings.TrimSpace(gameVersion)).Scan(&stable)
	if errors.Is(err, sql.ErrNoRows) {
		return false, false, nil
	}
	return true, stable != 0, err
}

func (b *CompatibilityBrain) LatestLoaderRelease(loader string, stableOnly bool) (BrainRelease, error) {
	loader = normalizeDoctorLoader(loader)
	query := `SELECT loader,version,stable,maven,sha256,file_size,ordinal FROM loader_releases WHERE loader=?`
	args := []any{loader}
	if stableOnly {
		query += ` AND stable=1`
	}
	query += ` ORDER BY ordinal ASC LIMIT 1`
	var out BrainRelease
	var stable int
	var ordinal int
	err := b.db.QueryRow(query, args...).Scan(&out.Tool, &out.Version, &stable, &out.Maven, &out.SHA256, &out.FileSize, &ordinal)
	if err != nil {
		return out, err
	}
	out.Stable = stable != 0
	out.Latest = ordinal == 0
	out.Source = loader + "-loader-versions.json"
	return out, nil
}

func (b *CompatibilityBrain) LatestToolRelease(tool string) (BrainRelease, error) {
	var out BrainRelease
	var latest, release int
	err := b.db.QueryRow(`SELECT tool,version,is_latest,is_release,source_file FROM toolchain_releases
		WHERE tool=? ORDER BY is_latest DESC,is_release DESC,ordinal DESC LIMIT 1`, strings.TrimSpace(tool)).Scan(
		&out.Tool, &out.Version, &latest, &release, &out.Source,
	)
	if err != nil {
		return out, err
	}
	out.Latest = latest != 0
	out.Release = release != 0
	return out, nil
}

func (b *CompatibilityBrain) ToolReleaseByPrefix(tool, prefix string) (BrainRelease, error) {
	var out BrainRelease
	var latest, release int
	pattern := strings.TrimSpace(prefix) + "%"
	err := b.db.QueryRow(`SELECT tool,version,is_latest,is_release,source_file FROM toolchain_releases
		WHERE tool=? AND version LIKE ? ORDER BY ordinal DESC LIMIT 1`, strings.TrimSpace(tool), pattern).Scan(
		&out.Tool, &out.Version, &latest, &release, &out.Source,
	)
	if err != nil {
		return out, err
	}
	out.Latest = latest != 0
	out.Release = release != 0
	return out, nil
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
