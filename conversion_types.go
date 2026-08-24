package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	conversionSchemaVersion     = 1
	conversionGraphVersion      = "umcg-1.0"
	conversionMaxUploadBytes    = int64(4 << 30)
	conversionMaxExtractedBytes = int64(8 << 30)
)

var errConversionSessionNotFound = errors.New("conversion session not found")

type ConversionLevel string

const (
	ConversionExact        ConversionLevel = "exact"
	ConversionTranslated   ConversionLevel = "translated"
	ConversionGenerated    ConversionLevel = "generated"
	ConversionToolAssisted ConversionLevel = "tool-assisted"
	ConversionReview       ConversionLevel = "review"
	ConversionBlocked      ConversionLevel = "blocked"
)

type ConversionSourceProfile struct {
	Filename       string            `json:"filename"`
	Path           string            `json:"path"`
	Size           int64             `json:"size"`
	SHA256         string            `json:"sha256"`
	TreeSHA256     string            `json:"treeSha256"`
	FileCount      int               `json:"fileCount"`
	ExtractedBytes int64             `json:"extractedBytes"`
	Edition        string            `json:"edition"`
	Format         string            `json:"format"`
	Kind           string            `json:"kind"`
	Name           string            `json:"name"`
	Description    string            `json:"description,omitempty"`
	Namespace      string            `json:"namespace"`
	Version        string            `json:"version,omitempty"`
	GameVersion    string            `json:"gameVersion,omitempty"`
	Loader         string            `json:"loader,omitempty"`
	UUID           string            `json:"uuid,omitempty"`
	MinimumEngine  string            `json:"minimumEngineVersion,omitempty"`
	ManifestPath   string            `json:"manifestPath,omitempty"`
	PackFormat     int               `json:"packFormat,omitempty"`
	Signals        []string          `json:"signals,omitempty"`
	Warnings       []string          `json:"warnings,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

type UniversalContentGraph struct {
	SchemaVersion int              `json:"schemaVersion"`
	GraphVersion  string           `json:"graphVersion"`
	GeneratedAt   string           `json:"generatedAt"`
	SourceSHA256  string           `json:"sourceSha256"`
	Namespace     string           `json:"namespace"`
	Nodes         []UniversalNode  `json:"nodes"`
	Relationships []UniversalEdge  `json:"relationships,omitempty"`
	Summary       UniversalSummary `json:"summary"`
	Warnings      []string         `json:"warnings,omitempty"`
}

type UniversalNode struct {
	ID             string            `json:"id"`
	Kind           string            `json:"kind"`
	Namespace      string            `json:"namespace"`
	Name           string            `json:"name"`
	SourcePath     string            `json:"sourcePath"`
	SourceFormat   string            `json:"sourceFormat,omitempty"`
	Level          ConversionLevel   `json:"level"`
	Confidence     float64           `json:"confidence"`
	TargetSupport  []string          `json:"targetSupport,omitempty"`
	Dependencies   []string          `json:"dependencies,omitempty"`
	Assets         []string          `json:"assets,omitempty"`
	Properties     map[string]string `json:"properties,omitempty"`
	Data           map[string]any    `json:"data,omitempty"`
	Notes          []string          `json:"notes,omitempty"`
	RequiresReview bool              `json:"requiresReview,omitempty"`
}

type UniversalEdge struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Relation string `json:"relation"`
	Evidence string `json:"evidence,omitempty"`
}

type UniversalSummary struct {
	Total          int            `json:"total"`
	ByKind         map[string]int `json:"byKind"`
	Assets         int            `json:"assets"`
	Data           int            `json:"data"`
	Logic          int            `json:"logic"`
	World          int            `json:"world"`
	Exact          int            `json:"exact"`
	Translated     int            `json:"translated"`
	Generated      int            `json:"generated"`
	ToolAssisted   int            `json:"toolAssisted"`
	ReviewRequired int            `json:"reviewRequired"`
	Blocked        int            `json:"blocked"`
}

type ConversionTargetSpec struct {
	Format              string   `json:"format"`
	Edition             string   `json:"edition"`
	GameVersion         string   `json:"gameVersion,omitempty"`
	Loader              string   `json:"loader,omitempty"`
	Name                string   `json:"name,omitempty"`
	Namespace           string   `json:"namespace,omitempty"`
	Description         string   `json:"description,omitempty"`
	Strategy            string   `json:"strategy,omitempty"`
	IncludeResourcePack bool     `json:"includeResourcePack,omitempty"`
	IncludeDataPack     bool     `json:"includeDataPack,omitempty"`
	IncludeSource       bool     `json:"includeSource,omitempty"`
	SelectedNodes       []string `json:"selectedNodes,omitempty"`
}

type ConversionCoverage struct {
	Total             int     `json:"total"`
	Exact             int     `json:"exact"`
	Translated        int     `json:"translated"`
	Generated         int     `json:"generated"`
	ToolAssisted      int     `json:"toolAssisted"`
	Review            int     `json:"review"`
	Blocked           int     `json:"blocked"`
	AutomatedPercent  float64 `json:"automatedPercent"`
	CompletenessState string  `json:"completenessState"`
}

type ConversionPlanStep struct {
	ID          string          `json:"id"`
	Order       int             `json:"order"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	State       string          `json:"state"`
	Level       ConversionLevel `json:"level"`
	NodeIDs     []string        `json:"nodeIds,omitempty"`
	ToolID      string          `json:"toolId,omitempty"`
	Evidence    []string        `json:"evidence,omitempty"`
}

type ConversionReviewItem struct {
	ID             string          `json:"id"`
	NodeID         string          `json:"nodeId,omitempty"`
	Severity       string          `json:"severity"`
	Category       string          `json:"category"`
	Title          string          `json:"title"`
	Reason         string          `json:"reason"`
	SuggestedRoute string          `json:"suggestedRoute"`
	Level          ConversionLevel `json:"level"`
	Resolved       bool            `json:"resolved"`
}

type ConversionPlan struct {
	SchemaVersion int                    `json:"schemaVersion"`
	CreatedAt     string                 `json:"createdAt"`
	SourceSHA256  string                 `json:"sourceSha256"`
	Target        ConversionTargetSpec   `json:"target"`
	Coverage      ConversionCoverage     `json:"coverage"`
	Steps         []ConversionPlanStep   `json:"steps"`
	ReviewQueue   []ConversionReviewItem `json:"reviewQueue,omitempty"`
	Losses        []string               `json:"losses,omitempty"`
	Warnings      []string               `json:"warnings,omitempty"`
	ToolAdapters  []string               `json:"toolAdapters,omitempty"`
	PlanSHA256    string                 `json:"planSha256,omitempty"`
}

type ConversionOutput struct {
	Name          string `json:"name"`
	RelativePath  string `json:"relativePath"`
	Kind          string `json:"kind"`
	Size          int64  `json:"size"`
	SHA256        string `json:"sha256"`
	CreatedAt     string `json:"createdAt"`
	DownloadIndex int    `json:"downloadIndex"`
	Validated     bool   `json:"validated"`
	Validation    string `json:"validation"`
}

type ConversionToolAdapter struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Role          string   `json:"role"`
	Formats       []string `json:"formats"`
	Directions    []string `json:"directions"`
	Maturity      string   `json:"maturity"`
	License       string   `json:"license,omitempty"`
	OfficialURL   string   `json:"officialUrl"`
	RepositoryURL string   `json:"repositoryUrl,omitempty"`
	Executable    string   `json:"executable,omitempty"`
	DetectedPath  string   `json:"detectedPath,omitempty"`
	Ready         bool     `json:"ready"`
	Notes         []string `json:"notes,omitempty"`
	Configured    bool     `json:"configured,omitempty"`
	CanExecute    bool     `json:"canExecute,omitempty"`
}

type ConversionSessionPaths struct {
	Root           string `json:"root"`
	Original       string `json:"original"`
	Extracted      string `json:"extracted"`
	Workspace      string `json:"workspace"`
	Outputs        string `json:"outputs"`
	ReceiptJSON    string `json:"receiptJson"`
	GraphJSON      string `json:"graphJson"`
	PlanJSON       string `json:"planJson"`
	ReportMarkdown string `json:"reportMarkdown"`
}

type ConversionAdapterRun struct {
	ID             string             `json:"id"`
	ToolID         string             `json:"toolId"`
	ToolName       string             `json:"toolName"`
	State          string             `json:"state"`
	StartedAt      string             `json:"startedAt"`
	CompletedAt    string             `json:"completedAt,omitempty"`
	Command        []string           `json:"command,omitempty"`
	WorkingDir     string             `json:"workingDir,omitempty"`
	LogPath        string             `json:"logPath,omitempty"`
	LogSHA256      string             `json:"logSha256,omitempty"`
	ExitCode       int                `json:"exitCode,omitempty"`
	SourceVerified bool               `json:"sourceVerified"`
	Outputs        []ConversionOutput `json:"outputs,omitempty"`
	Warnings       []string           `json:"warnings,omitempty"`
	Error          string             `json:"error,omitempty"`
}

type ConversionAdapterRunRequest struct {
	SessionID string            `json:"sessionId"`
	ToolID    string            `json:"toolId"`
	Options   map[string]string `json:"options,omitempty"`
}

type ConversionToolConfigRequest struct {
	ToolID string `json:"toolId"`
	Path   string `json:"path"`
}

type ConversionSession struct {
	SchemaVersion int                     `json:"schemaVersion"`
	ID            string                  `json:"id"`
	Name          string                  `json:"name"`
	State         string                  `json:"state"`
	Phase         string                  `json:"phase"`
	CreatedAt     string                  `json:"createdAt"`
	UpdatedAt     string                  `json:"updatedAt"`
	Source        ConversionSourceProfile `json:"source"`
	Graph         UniversalContentGraph   `json:"graph"`
	Plan          *ConversionPlan         `json:"plan,omitempty"`
	Outputs       []ConversionOutput      `json:"outputs,omitempty"`
	AdapterRuns   []ConversionAdapterRun  `json:"adapterRuns,omitempty"`
	Warnings      []string                `json:"warnings,omitempty"`
	LastError     string                  `json:"lastError,omitempty"`
	Paths         ConversionSessionPaths  `json:"paths"`
}

type ConversionSessionSummary struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	State        string  `json:"state"`
	UpdatedAt    string  `json:"updatedAt"`
	SourceFormat string  `json:"sourceFormat"`
	TargetFormat string  `json:"targetFormat,omitempty"`
	Automated    float64 `json:"automatedPercent,omitempty"`
	ReviewCount  int     `json:"reviewCount,omitempty"`
	OutputCount  int     `json:"outputCount"`
	LastError    string  `json:"lastError,omitempty"`
}

type ConversionStatus struct {
	SchemaVersion int                        `json:"schemaVersion"`
	Ready         bool                       `json:"ready"`
	GraphVersion  string                     `json:"graphVersion"`
	Root          string                     `json:"root"`
	Sessions      []ConversionSessionSummary `json:"sessions"`
	Targets       []ConversionTargetOption   `json:"targets"`
	Tools         []ConversionToolAdapter    `json:"tools"`
	Capabilities  []ConversionCapability     `json:"capabilities"`
	Safety        ConversionSafetyBoundary   `json:"safety"`
}

type ConversionTargetOption struct {
	ID          string `json:"id"`
	Edition     string `json:"edition"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Extension   string `json:"extension"`
	Maturity    string `json:"maturity"`
}

type ConversionCapability struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	State       string `json:"state"`
}

type ConversionSafetyBoundary struct {
	ImmutableOriginal bool     `json:"immutableOriginal"`
	PathProtection    bool     `json:"pathProtection"`
	ArchiveBombLimits bool     `json:"archiveBombLimits"`
	NoSilentLoss      bool     `json:"noSilentLoss"`
	Deterministic     bool     `json:"deterministicOutputs"`
	Limitations       []string `json:"limitations"`
}

type ConversionPlanRequest struct {
	SessionID string               `json:"sessionId"`
	Target    ConversionTargetSpec `json:"target"`
}

type ConversionRunRequest struct {
	SessionID string `json:"sessionId"`
}

type ConversionPathImportRequest struct {
	Path string `json:"path"`
}

type ConversionSessionRequest struct {
	SessionID string `json:"sessionId"`
}

func (a *App) conversionRoot() string {
	return filepath.Join(a.cfgDir, "conversion-studio")
}

func (a *App) conversionSessionsRoot() string {
	return filepath.Join(a.conversionRoot(), "sessions")
}

func (a *App) newConversionSession(name string) (*ConversionSession, error) {
	now := time.Now().UTC()
	id := fmt.Sprintf("convert-%s-%s", now.Format("20060102T150405Z"), randomToken(6))
	root := filepath.Join(a.conversionSessionsRoot(), id)
	paths := ConversionSessionPaths{
		Root: root, Original: filepath.Join(root, "original"), Extracted: filepath.Join(root, "extracted"),
		Workspace: filepath.Join(root, "workspace"), Outputs: filepath.Join(root, "outputs"),
		ReceiptJSON: filepath.Join(root, "session.json"), GraphJSON: filepath.Join(root, "universal-content-graph.json"),
		PlanJSON: filepath.Join(root, "conversion-plan.json"), ReportMarkdown: filepath.Join(root, "CONVERSION-REPORT.md"),
	}
	for _, dir := range []string{paths.Root, paths.Original, paths.Extracted, paths.Workspace, paths.Outputs} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}
	session := &ConversionSession{SchemaVersion: conversionSchemaVersion, ID: id, Name: cleanConversionName(name), State: "new", Phase: "intake", CreatedAt: now.Format(time.RFC3339), UpdatedAt: now.Format(time.RFC3339), Paths: paths}
	return session, nil
}

func cleanConversionName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "Untitled conversion"
	}
	value = strings.Map(func(r rune) rune {
		if r < 32 || strings.ContainsRune(`<>:"/\\|?*`, r) {
			return '-'
		}
		return r
	}, value)
	value = strings.Trim(value, " .-")
	if len(value) > 100 {
		value = value[:100]
	}
	if value == "" {
		return "Untitled conversion"
	}
	return value
}

func (a *App) saveConversionSession(session *ConversionSession) error {
	if session == nil || session.ID == "" {
		return errors.New("invalid conversion session")
	}
	session.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := writeJSONFileAtomic(session.Paths.ReceiptJSON, session); err != nil {
		return err
	}
	if session.Graph.GraphVersion != "" {
		if err := writeJSONFileAtomic(session.Paths.GraphJSON, session.Graph); err != nil {
			return err
		}
	}
	if session.Plan != nil {
		if err := writeJSONFileAtomic(session.Paths.PlanJSON, session.Plan); err != nil {
			return err
		}
	}
	return writeConversionReport(session)
}

func (a *App) loadConversionSession(id string) (*ConversionSession, error) {
	id = filepath.Base(strings.TrimSpace(id))
	if id == "" || id == "." {
		return nil, errConversionSessionNotFound
	}
	path := filepath.Join(a.conversionSessionsRoot(), id, "session.json")
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, errConversionSessionNotFound
	}
	if err != nil {
		return nil, err
	}
	var session ConversionSession
	if err := jsonUnmarshalStrict(data, &session); err != nil {
		return nil, err
	}
	if session.ID != id || !pathContainedBy(a.conversionSessionsRoot(), session.Paths.Root) {
		return nil, errors.New("conversion session identity is invalid")
	}
	return &session, nil
}

func jsonUnmarshalStrict(data []byte, value any) error {
	return json.Unmarshal(data, value)
}

func (a *App) listConversionSessions() ([]ConversionSessionSummary, error) {
	root := a.conversionSessionsRoot()
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	out := make([]ConversionSessionSummary, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		session, err := a.loadConversionSession(entry.Name())
		if err != nil {
			continue
		}
		summary := ConversionSessionSummary{ID: session.ID, Name: session.Name, State: session.State, UpdatedAt: session.UpdatedAt, SourceFormat: session.Source.Format, OutputCount: len(session.Outputs), LastError: session.LastError}
		if session.Plan != nil {
			summary.TargetFormat = session.Plan.Target.Format
			summary.Automated = session.Plan.Coverage.AutomatedPercent
			summary.ReviewCount = session.Plan.Coverage.Review + session.Plan.Coverage.Blocked
		}
		out = append(out, summary)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt > out[j].UpdatedAt })
	return out, nil
}

func conversionTargetOptions() []ConversionTargetOption {
	return []ConversionTargetOption{
		{ID: "bedrock-addon", Edition: "bedrock", Name: "Bedrock add-on family", Description: "Behavior pack + resource pack packaged as .mcaddon, with scripts and review queue when required.", Extension: ".mcaddon", Maturity: "stable-core"},
		{ID: "bedrock-behavior", Edition: "bedrock", Name: "Bedrock behavior pack", Description: "Recipes, loot, functions, entities, items, blocks, scripts and world rules as .mcpack.", Extension: ".mcpack", Maturity: "stable-core"},
		{ID: "bedrock-resource", Edition: "bedrock", Name: "Bedrock resource pack", Description: "Textures, models, sounds, languages, particles, animations and UI as .mcpack.", Extension: ".mcpack", Maturity: "stable-core"},
		{ID: "bedrock-world", Edition: "bedrock", Name: "Bedrock world", Description: "World package as .mcworld; delegates chunk conversion to a verified installed adapter when editions differ.", Extension: ".mcworld", Maturity: "adapter-backed"},
		{ID: "bedrock-template", Edition: "bedrock", Name: "Bedrock world template", Description: "Reusable .mctemplate with manifest, world data, packs and template metadata.", Extension: ".mctemplate", Maturity: "stable-core"},
		{ID: "bedrock-project", Edition: "bedrock", Name: "Bedrock add-on source project", Description: "Editable behavior/resource/script project with Regolith and bridge.-compatible structure, GameTest harness, contracts and packaged outputs.", Extension: ".zip", Maturity: "stable-core"},
		{ID: "bedrock-world-product", Edition: "bedrock", Name: "Bedrock world product bundle", Description: "World template, companion add-on, editable source, adapter evidence and product manifest in one release-ready bundle.", Extension: ".zip", Maturity: "stable-core"},
		{ID: "java-datapack", Edition: "java", Name: "Java data pack", Description: "Functions, recipes, loot, tags, advancements, predicates, structures and worldgen as a data pack ZIP.", Extension: ".zip", Maturity: "stable-core"},
		{ID: "java-resourcepack", Edition: "java", Name: "Java resource pack", Description: "Textures, models, sounds, languages, fonts, particles and shaders as a resource pack ZIP.", Extension: ".zip", Maturity: "stable-core"},
		{ID: "java-pack-bundle", Edition: "java", Name: "Java vanilla add-on family", Description: "Paired data pack + resource pack product with conversion contracts, optional source world/template and one install manifest.", Extension: ".zip", Maturity: "stable-core"},
		{ID: "java-world", Edition: "java", Name: "Java world", Description: "Standalone Java world ZIP; delegates cross-edition chunks to a verified installed adapter.", Extension: ".zip", Maturity: "adapter-backed"},
		{ID: "java-world-mod", Edition: "java", Name: "Java world-template mod source", Description: "Loader project with the world/template embedded as immutable resources, an exporter utility, translated packs and explicit installer/menu integration contracts.", Extension: ".zip", Maturity: "generated-scaffold"},
		{ID: "java-fabric", Edition: "java", Name: "Fabric source project", Description: "Generated source workspace and build metadata for Fabric, with unresolved behavior contracts surfaced for implementation.", Extension: ".zip", Maturity: "generated-scaffold"},
		{ID: "java-neoforge", Edition: "java", Name: "NeoForge source project", Description: "Generated source workspace and build metadata for NeoForge, with exact review tasks.", Extension: ".zip", Maturity: "generated-scaffold"},
		{ID: "java-forge", Edition: "java", Name: "Forge source project", Description: "Generated source workspace for Forge targets, including legacy-risk reporting.", Extension: ".zip", Maturity: "generated-scaffold"},
		{ID: "java-multiloader", Edition: "java", Name: "Multi-loader source project", Description: "Shared/common project layout for Fabric, NeoForge and Forge using current adapter recommendations.", Extension: ".zip", Maturity: "generated-scaffold"},
		{ID: "universal-bundle", Edition: "universal", Name: "Universal conversion bundle", Description: "UMCG graph, Bedrock family, Java data/resource packs, source scaffold and full conversion proof in one archive.", Extension: ".zip", Maturity: "stable-core"},
	}
}

func conversionSafetyBoundary() ConversionSafetyBoundary {
	return ConversionSafetyBoundary{ImmutableOriginal: true, PathProtection: true, ArchiveBombLimits: true, NoSilentLoss: true, Deterministic: true, Limitations: []string{
		"Java bytecode, Mixins, ASM, custom networking, native code, renderer internals and complex GUIs do not have universal Bedrock equivalents; they are emitted as explicit review contracts or Script API scaffolds rather than silently discarded.",
		"Cross-edition terrain/chunk conversion requires a supported installed adapter such as Chunker, je2be or Amulet; OmniBridge never pretends an unexecuted adapter step is complete.",
		"Generated Java/Bedrock source projects are real editable workspaces, not proof of semantic parity. Successful target builds plus client, server, GameTest, persistence and gameplay verification remain separate completion gates.",
	}}
}
