package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const versionAtlasPath = "assets/version-atlas.json"

type VersionAtlas struct {
	SchemaVersion      int                         `json:"schemaVersion"`
	GeneratedAt        string                      `json:"generatedAt"`
	EvidenceSnapshotAt string                      `json:"evidenceSnapshotAt"`
	Summary            VersionAtlasSummary         `json:"summary"`
	Versions           []VersionAtlasRecord        `json:"versions"`
	Loaders            VersionAtlasLoaders         `json:"loaders"`
	Toolchains         map[string]VersionToolchain `json:"toolchains"`
	Forge              VersionBuildCatalog         `json:"forge"`
	NeoForge           VersionBuildCatalog         `json:"neoForge"`
	SourceEvidence     []VersionSourceEvidence     `json:"sourceEvidence"`
}

type VersionAtlasSummary struct {
	MinecraftVersions   int            `json:"minecraftVersions"`
	LatestRelease       string         `json:"latestRelease"`
	LatestSnapshot      string         `json:"latestSnapshot"`
	Oldest              string         `json:"oldest"`
	Newest              string         `json:"newest"`
	Types               map[string]int `json:"types"`
	JavaMajors          map[string]int `json:"javaMajors"`
	ClientMappings      int            `json:"clientMappings"`
	ServerMappings      int            `json:"serverMappings"`
	MCMETACoverage      int            `json:"mcmetaCoverage"`
	FabricGameVersions  int            `json:"fabricGameVersions"`
	QuiltGameVersions   int            `json:"quiltGameVersions"`
	ModrinthGameVersion int            `json:"modrinthGameVersions"`
}

type VersionAtlasRecord struct {
	ID                  string `json:"id"`
	Type                string `json:"type"`
	ReleaseTime         string `json:"releaseTime"`
	Time                string `json:"time"`
	JavaMajor           *int   `json:"javaMajor"`
	JavaComponent       string `json:"javaComponent"`
	Client              bool   `json:"client"`
	Server              bool   `json:"server"`
	ClientMappings      bool   `json:"clientMappings"`
	ServerMappings      bool   `json:"serverMappings"`
	AssetIndex          string `json:"assetIndex"`
	LibraryCount        int    `json:"libraryCount"`
	ComplianceLevel     *int   `json:"complianceLevel"`
	DataVersion         *int   `json:"dataVersion"`
	ProtocolVersion     *int   `json:"protocolVersion"`
	DataPackVersion     *int   `json:"dataPackVersion"`
	ResourcePackVersion *int   `json:"resourcePackVersion"`
	ReleaseTarget       string `json:"releaseTarget"`
	Stable              bool   `json:"stable"`
	Fabric              bool   `json:"fabric"`
	Quilt               bool   `json:"quilt"`
	Modrinth            bool   `json:"modrinth"`
}

type VersionAtlasLoaders struct {
	Fabric   VersionLoaderFamily `json:"fabric"`
	Quilt    VersionLoaderFamily `json:"quilt"`
	Modrinth struct {
		IDs              []string `json:"ids"`
		GameVersionCount int      `json:"gameVersionCount"`
	} `json:"modrinth"`
}

type VersionLoaderFamily struct {
	Latest           string   `json:"latest"`
	LatestStable     string   `json:"latestStable"`
	Recent           []string `json:"recent"`
	GameVersionCount int      `json:"gameVersionCount"`
}

type VersionToolchain struct {
	Latest       string   `json:"latest"`
	Release      string   `json:"release"`
	LatestStable string   `json:"latestStable"`
	Count        int      `json:"count"`
	Recent       []string `json:"recent"`
	Source       string   `json:"source"`
	SHA256       string   `json:"sha256"`
}

type VersionBuildCatalog struct {
	Latest            string            `json:"latest"`
	LatestStable      string            `json:"latestStable"`
	LatestByMinecraft map[string]string `json:"latestByMinecraft"`
	Count             int               `json:"count"`
}

type VersionSourceEvidence struct {
	Name   string `json:"name"`
	Bytes  int64  `json:"bytes"`
	SHA256 string `json:"sha256"`
}

type VersionAtlasResponse struct {
	SchemaVersion      int                         `json:"schemaVersion"`
	GeneratedAt        string                      `json:"generatedAt"`
	EvidenceSnapshotAt string                      `json:"evidenceSnapshotAt"`
	Summary            VersionAtlasSummary         `json:"summary"`
	Loaders            VersionAtlasLoaders         `json:"loaders"`
	Toolchains         map[string]VersionToolchain `json:"toolchains"`
	Forge              VersionBuildCatalog         `json:"forge"`
	NeoForge           VersionBuildCatalog         `json:"neoForge"`
	Versions           []VersionAtlasRecord        `json:"versions"`
	TotalMatches       int                         `json:"totalMatches"`
	SourceEvidence     []VersionSourceEvidence     `json:"sourceEvidence,omitempty"`
}

type PortingPlanRequest struct {
	SourceGameVersion string `json:"sourceGameVersion"`
	SourceLoader      string `json:"sourceLoader"`
	TargetGameVersion string `json:"targetGameVersion"`
	TargetLoader      string `json:"targetLoader"`
	SourceMode        string `json:"sourceMode,omitempty"`
	InputJar          string `json:"inputJar,omitempty"`
	ProjectName       string `json:"projectName,omitempty"`
}

type PortingPhase struct {
	Order       int      `json:"order"`
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Goal        string   `json:"goal"`
	Actions     []string `json:"actions"`
	Gates       []string `json:"gates"`
	ToolIDs     []string `json:"toolIds,omitempty"`
	Destructive bool     `json:"destructive"`
}

type PortingToolRecommendation struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Category      string `json:"category"`
	Maturity      string `json:"maturity"`
	OfficialURL   string `json:"officialUrl"`
	PinnedVersion string `json:"pinnedVersion,omitempty"`
	Required      bool   `json:"required"`
	Reason        string `json:"reason"`
	Guardrail     string `json:"guardrail,omitempty"`
}

type PortingPlan struct {
	ID                   string                      `json:"id"`
	CreatedAt            string                      `json:"createdAt"`
	EvidenceSnapshotAt   string                      `json:"evidenceSnapshotAt"`
	ProjectName          string                      `json:"projectName"`
	SourceMode           string                      `json:"sourceMode"`
	Direction            string                      `json:"direction"`
	Risk                 string                      `json:"risk"`
	Source               VersionAtlasRecord          `json:"source"`
	Target               VersionAtlasRecord          `json:"target"`
	SourceLoader         string                      `json:"sourceLoader"`
	TargetLoader         string                      `json:"targetLoader"`
	InputJar             string                      `json:"inputJar,omitempty"`
	InputAnalysis        *PortingInputAnalysis       `json:"inputAnalysis,omitempty"`
	Pins                 map[string]string           `json:"pins"`
	Boundaries           []string                    `json:"boundaries"`
	Warnings             []string                    `json:"warnings"`
	Tools                []PortingToolRecommendation `json:"tools"`
	Phases               []PortingPhase              `json:"phases"`
	VerificationMatrix   []string                    `json:"verificationMatrix"`
	CompletionDefinition []string                    `json:"completionDefinition"`
}

type PortingInputAnalysis struct {
	Path                   string             `json:"path"`
	Filename               string             `json:"filename"`
	Size                   int64              `json:"size"`
	SHA256                 string             `json:"sha256"`
	SHA512                 string             `json:"sha512"`
	CurseFingerprint       uint32             `json:"curseFingerprint"`
	Metadata               JarMetadata        `json:"metadata"`
	DetectedLoaders        []string           `json:"detectedLoaders,omitempty"`
	ModIDs                 []string           `json:"modIds,omitempty"`
	Dependencies           []DoctorDependency `json:"dependencies,omitempty"`
	ClassCount             int                `json:"classCount"`
	MaxJava                int                `json:"maxJava,omitempty"`
	MappingNamespace       string             `json:"mappingNamespace,omitempty"`
	MixinConfigs           []string           `json:"mixinConfigs,omitempty"`
	AccessWideners         []string           `json:"accessWideners,omitempty"`
	AccessTransformers     []string           `json:"accessTransformers,omitempty"`
	Coremods               []string           `json:"coremods,omitempty"`
	TransformationServices []string           `json:"transformationServices,omitempty"`
	NestedJars             []string           `json:"nestedJars,omitempty"`
	NativeLibraries        []string           `json:"nativeLibraries,omitempty"`
	SignatureFiles         []string           `json:"signatureFiles,omitempty"`
	DataFileCount          int                `json:"dataFileCount,omitempty"`
	AssetFileCount         int                `json:"assetFileCount,omitempty"`
	UsesReflection         bool               `json:"usesReflection,omitempty"`
	UsesUnsafe             bool               `json:"usesUnsafe,omitempty"`
	UsesMethodHandles      bool               `json:"usesMethodHandles,omitempty"`
	UsesMixinExtras        bool               `json:"usesMixinExtras,omitempty"`
	UsesKotlin             bool               `json:"usesKotlin,omitempty"`
	UsesScala              bool               `json:"usesScala,omitempty"`
	HasClientReferences    bool               `json:"hasClientReferences,omitempty"`
	HasServerReferences    bool               `json:"hasServerReferences,omitempty"`
	HasPackMetadata        bool               `json:"hasPackMetadata,omitempty"`
	TruncatedAnalysis      bool               `json:"truncatedAnalysis,omitempty"`
	RiskSignals            []string           `json:"riskSignals,omitempty"`
}

type PortingWorkspaceRequest struct {
	PlanID      string `json:"planId"`
	ProjectName string `json:"projectName,omitempty"`
	InputJar    string `json:"inputJar,omitempty"`
}

type PortingWorkspaceManifest struct {
	SchemaVersion int                    `json:"schemaVersion"`
	ID            string                 `json:"id"`
	CreatedAt     string                 `json:"createdAt"`
	Path          string                 `json:"path"`
	PlanID        string                 `json:"planId"`
	ProjectName   string                 `json:"projectName"`
	Source        string                 `json:"source"`
	Target        string                 `json:"target"`
	SourceLoader  string                 `json:"sourceLoader"`
	TargetLoader  string                 `json:"targetLoader"`
	Input         *PortingWorkspaceInput `json:"input,omitempty"`
	Files         []PortingWorkspaceFile `json:"files"`
	Status        string                 `json:"status"`
	NextCommand   string                 `json:"nextCommand"`
	Rollback      string                 `json:"rollback"`
	Evidence      map[string]string      `json:"evidence"`
}

type PortingWorkspaceInput struct {
	OriginalPath  string `json:"originalPath"`
	WorkspacePath string `json:"workspacePath"`
	Filename      string `json:"filename"`
	Size          int64  `json:"size"`
	SHA256        string `json:"sha256"`
}

type PortingWorkspaceFile struct {
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}

type PortingEnvironment struct {
	CreatedAt          string                   `json:"createdAt"`
	OS                 string                   `json:"os"`
	Arch               string                   `json:"arch"`
	TargetGameVersion  string                   `json:"targetGameVersion,omitempty"`
	RequiredJava       int                      `json:"requiredJava,omitempty"`
	Commands           []PortingEnvironmentTool `json:"commands"`
	ReadyForPlanning   bool                     `json:"readyForPlanning"`
	ReadyForLocalBuild bool                     `json:"readyForLocalBuild"`
	Warnings           []string                 `json:"warnings"`
}

type PortingEnvironmentTool struct {
	ID        string `json:"id"`
	Installed bool   `json:"installed"`
	Path      string `json:"path,omitempty"`
	Version   string `json:"version,omitempty"`
	Required  bool   `json:"required"`
	Purpose   string `json:"purpose"`
}

var (
	versionAtlasOnce sync.Once
	versionAtlasData VersionAtlas
	versionAtlasErr  error
)

func loadVersionAtlas() (VersionAtlas, error) {
	versionAtlasOnce.Do(func() {
		data, err := embeddedFiles.ReadFile(versionAtlasPath)
		if err != nil {
			versionAtlasErr = err
			return
		}
		if err := json.Unmarshal(data, &versionAtlasData); err != nil {
			versionAtlasErr = err
			return
		}
		if versionAtlasData.SchemaVersion <= 0 || len(versionAtlasData.Versions) == 0 {
			versionAtlasErr = errors.New("version atlas is empty or invalid")
		}
	})
	return versionAtlasData, versionAtlasErr
}

func (a *App) handlePortingAtlas(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	atlas, err := loadVersionAtlas()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
		return
	}
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	kind := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("type")))
	limit := 120
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, parseErr := strconv.Atoi(raw); parseErr == nil && parsed > 0 {
			limit = parsed
		}
	}
	if limit > 1000 {
		limit = 1000
	}
	matches := make([]VersionAtlasRecord, 0, minInt(limit, len(atlas.Versions)))
	total := 0
	for _, row := range atlas.Versions {
		if q != "" && !strings.Contains(strings.ToLower(row.ID), q) && !strings.Contains(strings.ToLower(row.ReleaseTarget), q) {
			continue
		}
		if kind != "" && kind != "all" && strings.ToLower(row.Type) != kind {
			continue
		}
		total++
		if len(matches) < limit {
			matches = append(matches, row)
		}
	}
	response := VersionAtlasResponse{
		SchemaVersion: atlas.SchemaVersion, GeneratedAt: atlas.GeneratedAt, EvidenceSnapshotAt: atlas.EvidenceSnapshotAt,
		Summary: atlas.Summary, Loaders: atlas.Loaders, Toolchains: atlas.Toolchains, Forge: atlas.Forge,
		NeoForge: atlas.NeoForge, Versions: matches, TotalMatches: total,
	}
	if r.URL.Query().Get("evidence") == "1" {
		response.SourceEvidence = atlas.SourceEvidence
	}
	writeJSON(w, http.StatusOK, response)
}

func (a *App) handlePortingPlan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request PortingPlanRequest
	if r.Body != nil && r.ContentLength != 0 {
		if err := decodeJSON(r, &request); err != nil {
			writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
			return
		}
	}
	plan, err := a.buildPortingPlan(request)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	a.dataMu.Lock()
	if a.portingPlans == nil {
		a.portingPlans = map[string]PortingPlan{}
	}
	a.portingPlans[plan.ID] = plan
	for id, old := range a.portingPlans {
		created, _ := time.Parse(time.RFC3339, old.CreatedAt)
		if !created.IsZero() && time.Since(created) > 24*time.Hour {
			delete(a.portingPlans, id)
		}
	}
	a.dataMu.Unlock()
	writeJSON(w, http.StatusOK, plan)
}

func (a *App) buildPortingPlan(request PortingPlanRequest) (PortingPlan, error) {
	atlas, err := loadVersionAtlas()
	if err != nil {
		return PortingPlan{}, err
	}
	a.mu.RLock()
	settings := a.settings
	a.mu.RUnlock()
	request.SourceGameVersion = firstNonEmpty(strings.TrimSpace(request.SourceGameVersion), settings.GameVersion)
	request.TargetGameVersion = firstNonEmpty(strings.TrimSpace(request.TargetGameVersion), settings.GameVersion)
	request.SourceLoader = normalizePortingLoader(firstNonEmpty(request.SourceLoader, settings.Loader))
	request.TargetLoader = normalizePortingLoader(firstNonEmpty(request.TargetLoader, settings.Loader))
	request.SourceMode = strings.ToLower(strings.TrimSpace(request.SourceMode))
	if request.SourceMode == "" || request.SourceMode == "auto" {
		if strings.TrimSpace(request.InputJar) != "" {
			request.SourceMode = "binary"
		} else {
			request.SourceMode = "source"
		}
	}
	if request.SourceMode != "source" && request.SourceMode != "binary" {
		return PortingPlan{}, errors.New("sourceMode must be source, binary, or auto")
	}
	var inputAnalysis *PortingInputAnalysis
	if strings.TrimSpace(request.InputJar) != "" {
		analysis, analysisErr := a.analyzePortingInput(request.InputJar)
		if analysisErr != nil {
			return PortingPlan{}, fmt.Errorf("input JAR analysis failed: %w", analysisErr)
		}
		inputAnalysis = &analysis
		request.InputJar = analysis.Path
	}
	source, sourceIndex, ok := findAtlasVersion(atlas, request.SourceGameVersion)
	if !ok {
		return PortingPlan{}, fmt.Errorf("source Minecraft version %q is not in the embedded official atlas", request.SourceGameVersion)
	}
	target, targetIndex, ok := findAtlasVersion(atlas, request.TargetGameVersion)
	if !ok {
		return PortingPlan{}, fmt.Errorf("target Minecraft version %q is not in the embedded official atlas", request.TargetGameVersion)
	}
	if !validPortingLoader(request.SourceLoader) || !validPortingLoader(request.TargetLoader) {
		return PortingPlan{}, errors.New("loader must be fabric, quilt, forge, neoforge, vanilla, or multiloader")
	}

	direction := "rebuild"
	switch {
	case sourceIndex > targetIndex:
		direction = "upgrade"
	case sourceIndex < targetIndex:
		direction = "downgrade"
	case request.SourceLoader != request.TargetLoader:
		direction = "loader-migration"
	}
	project := strings.TrimSpace(request.ProjectName)
	if project == "" {
		project = fmt.Sprintf("%s-%s-to-%s-%s", request.SourceLoader, source.ID, request.TargetLoader, target.ID)
	}
	pins := portingPins(atlas, target, request.TargetLoader)
	boundaries := portingBoundaries(source, target, request.SourceLoader, request.TargetLoader)
	warnings := portingWarnings(direction, request.SourceMode, source, target, request.SourceLoader, request.TargetLoader, pins)
	tools := portingTools(atlas, source, target, request.SourceLoader, request.TargetLoader, request.SourceMode, pins)
	phases := portingPhases(direction, request.SourceMode, request.SourceLoader, request.TargetLoader)
	risk := portingRisk(direction, request.SourceMode, source, target, request.SourceLoader, request.TargetLoader)

	plan := PortingPlan{
		ID: randomToken(12), CreatedAt: time.Now().UTC().Format(time.RFC3339), EvidenceSnapshotAt: atlas.EvidenceSnapshotAt,
		ProjectName: project, SourceMode: request.SourceMode, Direction: direction, Risk: risk,
		Source: source, Target: target, SourceLoader: request.SourceLoader, TargetLoader: request.TargetLoader,
		InputJar: strings.TrimSpace(request.InputJar), InputAnalysis: inputAnalysis, Pins: pins, Boundaries: boundaries, Warnings: warnings,
		Tools: tools, Phases: phases,
		VerificationMatrix: []string{
			"Compile from a clean dependency cache with the pinned Java and build toolchain.",
			"Run static linkage, mixin/refmap, access widener/transformer, resource, and data schema checks.",
			"Launch a disposable client and create/load a disposable world with fresh logs.",
			"Launch a disposable dedicated server, join once, exercise networking/config sync, then stop cleanly.",
			"Exercise the mod's highest-risk features, integrations, saved data, recipes, registries, and world generation.",
			"Rebuild twice and compare package contents, hashes, embedded dependencies, signatures, and metadata.",
		},
		CompletionDefinition: []string{
			"No source, mapping, compile, linkage, mixin, resource, registry, data-fixer, network, or side-only failures remain.",
			"Client and dedicated-server tests cover the mod's actual behavior rather than startup alone.",
			"Upgrade and downgrade paths preserve documented data or explicitly report irreversible loss.",
			"The original artifact and every mutation remain recoverable through hashes, receipts, backups, and rollback instructions.",
		},
	}
	if inputAnalysis != nil {
		enrichPortingPlanWithInput(&plan, *inputAnalysis)
	}
	return plan, nil
}

func (a *App) analyzePortingInput(inputPath string) (PortingInputAnalysis, error) {
	inputPath = filepath.Clean(strings.TrimSpace(inputPath))
	if !a.allowedManagedPath(inputPath) {
		return PortingInputAnalysis{}, errors.New("input JAR must be inside a configured Minecraft mods or Vault downloads directory")
	}
	info, err := os.Lstat(inputPath)
	if err != nil {
		return PortingInputAnalysis{}, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return PortingInputAnalysis{}, errors.New("symbolic-link inputs are refused; select the real JAR")
	}
	if !info.Mode().IsRegular() {
		return PortingInputAnalysis{}, errors.New("input is not a regular file")
	}
	lower := strings.ToLower(filepath.Base(inputPath))
	if !strings.HasSuffix(lower, ".jar") && !strings.HasSuffix(lower, ".jar.disabled") {
		return PortingInputAnalysis{}, errors.New("input must be a .jar or .jar.disabled file")
	}
	local, err := inspectLocalJar(inputPath)
	if err != nil {
		return PortingInputAnalysis{}, err
	}
	signals, err := inspectJarSignals(inputPath)
	if err != nil {
		return PortingInputAnalysis{}, err
	}
	sha256Digest, size, err := fileSHA256(inputPath)
	if err != nil {
		return PortingInputAnalysis{}, err
	}
	detected := []string{}
	for _, loader := range local.Metadata.Loaders {
		loader = normalizePortingLoader(loader)
		if validPortingLoader(loader) {
			detected = append(detected, loader)
		}
	}
	for _, api := range signals.LoaderAPIs {
		switch api {
		case "fabric-loader", "fabric-api":
			detected = append(detected, "fabric")
		case "quilt-loader", "quilt-api":
			detected = append(detected, "quilt")
		case "forge":
			detected = append(detected, "forge")
		case "neoforge":
			detected = append(detected, "neoforge")
		}
	}
	modIDs := append([]string{}, signals.ModIDs...)
	if local.Metadata.ModID != "" {
		modIDs = append(modIDs, local.Metadata.ModID)
	}
	riskSignals := []string{}
	if local.Metadata.SourceURL == "" {
		riskSignals = append(riskSignals, "No source repository is declared in the JAR metadata.")
	}
	if len(signals.MixinConfigs) > 0 {
		riskSignals = append(riskSignals, fmt.Sprintf("%d mixin configuration(s) require target/refmap revalidation.", len(signals.MixinConfigs)))
	}
	if len(signals.AccessWideners)+len(signals.AccessTransformers) > 0 {
		riskSignals = append(riskSignals, "Access widening/transformation rules must be translated and reverified for the target namespace.")
	}
	if len(signals.Coremods)+len(signals.TransformationServices) > 0 {
		riskSignals = append(riskSignals, "Coremod or transformation-service hooks require manual semantic reconstruction and startup-order testing.")
	}
	if len(signals.NativeLibraries) > 0 {
		riskSignals = append(riskSignals, "Native libraries require OS/architecture ABI verification and cannot be assumed portable.")
	}
	if len(signals.SignatureFiles) > 0 {
		riskSignals = append(riskSignals, "The input is signed; any bytecode/resource mutation invalidates the original signature and provenance must be re-established.")
	}
	if signals.UsesUnsafe || signals.UsesReflection || signals.UsesMethodHandles {
		riskSignals = append(riskSignals, "Reflection, Unsafe, or MethodHandles usage raises Java/module/runtime compatibility risk.")
	}
	if signals.TruncatedAnalysis {
		riskSignals = append(riskSignals, "Static inspection hit its bounded read ceiling; run independent full-artifact analysis before editing.")
	}
	return PortingInputAnalysis{
		Path: inputPath, Filename: local.Filename, Size: size, SHA256: sha256Digest, SHA512: local.SHA512,
		CurseFingerprint: local.CurseFingerprint, Metadata: local.Metadata, DetectedLoaders: uniqueStringsPreserve(detected),
		ModIDs: uniqueStringsPreserve(modIDs), Dependencies: signals.Dependencies, ClassCount: signals.ClassCount, MaxJava: signals.MaxJava,
		MappingNamespace: signals.MappingNamespace, MixinConfigs: signals.MixinConfigs, AccessWideners: signals.AccessWideners,
		AccessTransformers: signals.AccessTransformers, Coremods: signals.Coremods, TransformationServices: signals.TransformationServices,
		NestedJars: signals.NestedJars, NativeLibraries: signals.NativeLibraries, SignatureFiles: signals.SignatureFiles,
		DataFileCount: signals.DataFileCount, AssetFileCount: signals.AssetFileCount, UsesReflection: signals.UsesReflection,
		UsesUnsafe: signals.UsesUnsafe, UsesMethodHandles: signals.UsesMethodHandles, UsesMixinExtras: signals.UsesMixinExtras,
		UsesKotlin: signals.UsesKotlin, UsesScala: signals.UsesScala, HasClientReferences: signals.HasClientReferences,
		HasServerReferences: signals.HasServerReferences, HasPackMetadata: signals.HasPackMetadata,
		TruncatedAnalysis: signals.TruncatedAnalysis, RiskSignals: uniqueStringsPreserve(riskSignals),
	}, nil
}

func enrichPortingPlanWithInput(plan *PortingPlan, input PortingInputAnalysis) {
	identity := input.Filename
	if len(input.ModIDs) > 0 {
		identity += " [" + strings.Join(input.ModIDs, ", ") + "]"
	}
	plan.Boundaries = append([]string{fmt.Sprintf("Input artifact: %s · %d classes · SHA-256 %s", identity, input.ClassCount, input.SHA256)}, plan.Boundaries...)
	if len(input.DetectedLoaders) > 0 && plan.SourceLoader != "multiloader" {
		matched := false
		for _, loader := range input.DetectedLoaders {
			if loader == plan.SourceLoader {
				matched = true
				break
			}
		}
		if !matched {
			plan.Warnings = append(plan.Warnings, fmt.Sprintf("Selected source loader %s does not match JAR evidence (%s). Reconfirm the exact source artifact before migration.", plan.SourceLoader, strings.Join(input.DetectedLoaders, ", ")))
		}
	}
	if input.Metadata.Minecraft != "" && !strings.Contains(strings.ToLower(input.Metadata.Minecraft), strings.ToLower(plan.Source.ID)) {
		plan.Warnings = append(plan.Warnings, fmt.Sprintf("JAR metadata declares Minecraft %q, which does not explicitly contain selected source %s.", input.Metadata.Minecraft, plan.Source.ID))
	}
	if input.MaxJava > 0 && input.MaxJava > javaMajor(plan.Source) {
		plan.Warnings = append(plan.Warnings, fmt.Sprintf("The input contains Java %d bytecode while the selected source manifest requests Java %d; the source version or artifact may be mislabeled.", input.MaxJava, javaMajor(plan.Source)))
	}
	plan.Warnings = uniqueStringsPreserve(append(plan.Warnings, input.RiskSignals...))
	plan.VerificationMatrix = append([]string{
		"Re-hash the selected input immediately before reconstruction and match SHA-256 " + input.SHA256 + ".",
		"Reconcile detected mod IDs, loader evidence, dependencies, mixins/access rules, nested JARs, signatures, and native libraries against the recovered source tree.",
	}, plan.VerificationMatrix...)
	if input.Metadata.SourceURL != "" {
		plan.CompletionDefinition = append([]string{"The recovered upstream source and exact commit are linked back to declared source " + input.Metadata.SourceURL + "."}, plan.CompletionDefinition...)
	} else {
		plan.CompletionDefinition = append([]string{"Source provenance is resolved independently, or binary-only reconstruction limitations are documented with dual-decompiler evidence."}, plan.CompletionDefinition...)
	}
}

func findAtlasVersion(atlas VersionAtlas, id string) (VersionAtlasRecord, int, bool) {
	id = strings.TrimSpace(id)
	for index, row := range atlas.Versions {
		if strings.EqualFold(row.ID, id) {
			return row, index, true
		}
	}
	return VersionAtlasRecord{}, -1, false
}

func normalizePortingLoader(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "neo-forge", "neo_forge":
		return "neoforge"
	case "multi", "cross-loader", "crossloader":
		return "multiloader"
	default:
		return value
	}
}

func validPortingLoader(value string) bool {
	switch value {
	case "fabric", "quilt", "forge", "neoforge", "vanilla", "multiloader":
		return true
	default:
		return false
	}
}

func javaMajor(row VersionAtlasRecord) int {
	if row.JavaMajor != nil && *row.JavaMajor > 0 {
		return *row.JavaMajor
	}
	return 8
}

func portingPins(atlas VersionAtlas, target VersionAtlasRecord, loader string) map[string]string {
	pins := map[string]string{
		"minecraft": target.ID,
		"java":      strconv.Itoa(javaMajor(target)),
	}
	stable := func(key string) string {
		tool := atlas.Toolchains[key]
		return firstNonEmpty(tool.LatestStable, tool.Release, tool.Latest)
	}
	switch loader {
	case "fabric":
		pins["fabricLoader"] = firstNonEmpty(atlas.Loaders.Fabric.LatestStable, atlas.Loaders.Fabric.Latest)
		pins["fabricLoom"] = stable("fabricLoom")
		pins["mappings"] = "officialMojangMappings"
	case "quilt":
		pins["quiltLoader"] = firstNonEmpty(atlas.Loaders.Quilt.LatestStable, atlas.Loaders.Quilt.Latest)
		pins["architecturyLoom"] = stable("architecturyLoom")
		pins["mappings"] = "officialMojangMappings"
	case "forge":
		pins["forge"] = atlas.Forge.LatestByMinecraft[target.ID]
		pins["forgeGradle"] = stable("forgeGradle")
		pins["mappings"] = "official"
	case "neoforge":
		pins["neoForge"] = atlas.NeoForge.LatestByMinecraft[target.ID]
		pins["modDevGradle"] = stable("modDevGradle")
		pins["neoForm"] = stable("neoForm")
		pins["mappings"] = "official"
	case "multiloader":
		pins["fabricLoader"] = firstNonEmpty(atlas.Loaders.Fabric.LatestStable, atlas.Loaders.Fabric.Latest)
		pins["neoForge"] = atlas.NeoForge.LatestByMinecraft[target.ID]
		pins["architecturyLoom"] = stable("architecturyLoom")
		pins["fabricLoom"] = stable("fabricLoom")
		pins["modDevGradle"] = stable("modDevGradle")
		pins["mappings"] = "officialMojangMappings"
	}
	pins["tinyRemapper"] = stable("autoRenamingTool")
	pins["mixin"] = stable("mixin")
	pins["mixinExtras"] = stable("mixinExtras")
	pins["srgUtils"] = stable("srgUtils")
	return pins
}

func portingBoundaries(source, target VersionAtlasRecord, sourceLoader, targetLoader string) []string {
	out := []string{}
	if source.ID != target.ID {
		out = append(out, fmt.Sprintf("Minecraft runtime: %s -> %s", source.ID, target.ID))
	}
	if sourceLoader != targetLoader {
		out = append(out, fmt.Sprintf("Loader/API surface: %s -> %s", sourceLoader, targetLoader))
	}
	if javaMajor(source) != javaMajor(target) {
		out = append(out, fmt.Sprintf("Java bytecode/runtime: Java %d -> Java %d", javaMajor(source), javaMajor(target)))
	}
	if source.ClientMappings != target.ClientMappings || source.ServerMappings != target.ServerMappings {
		out = append(out, "Official mapping artifact availability changes across the selected versions; namespace assumptions must be re-derived, not copied.")
	}
	if source.DataPackVersion != nil && target.DataPackVersion != nil && *source.DataPackVersion != *target.DataPackVersion {
		out = append(out, fmt.Sprintf("Data pack schema: %d -> %d", *source.DataPackVersion, *target.DataPackVersion))
	}
	if source.ResourcePackVersion != nil && target.ResourcePackVersion != nil && *source.ResourcePackVersion != *target.ResourcePackVersion {
		out = append(out, fmt.Sprintf("Resource pack schema: %d -> %d", *source.ResourcePackVersion, *target.ResourcePackVersion))
	}
	if source.ProtocolVersion != nil && target.ProtocolVersion != nil && *source.ProtocolVersion != *target.ProtocolVersion {
		out = append(out, fmt.Sprintf("Network protocol: %d -> %d", *source.ProtocolVersion, *target.ProtocolVersion))
	}
	if len(out) == 0 {
		out = append(out, "No version or loader boundary was selected; this is a reproducibility and compatibility rebuild.")
	}
	return out
}

func portingWarnings(direction, sourceMode string, source, target VersionAtlasRecord, sourceLoader, targetLoader string, pins map[string]string) []string {
	warnings := []string{
		"A generated plan or successful compile is evidence, not proof of a completed port.",
		"The workspace never mutates the installed mod; changes happen in an isolated copy with hashes and rollback instructions.",
	}
	if direction == "downgrade" {
		warnings = append(warnings, "Downgrades can destroy or reinterpret registries, components/NBT, recipes, world data, networking, and configuration. Test only on disposable copies until loss is measured.")
	}
	if sourceMode == "binary" {
		warnings = append(warnings, "Binary-only reconstruction is lower confidence than an upstream source build. Resolve official source and exact dependencies before editing decompiler output whenever possible.")
	}
	if sourceLoader != targetLoader {
		warnings = append(warnings, "Cross-loader ports require semantic API replacement and side/runtime testing; metadata translation or a compatibility bridge alone is not completion.")
	}
	if targetLoader == "forge" && pins["forge"] == "" {
		warnings = append(warnings, "No exact Forge coordinate was found for the target version in the embedded official Maven snapshot; do not guess a coordinate.")
	}
	if (targetLoader == "neoforge" || targetLoader == "multiloader") && pins["neoForge"] == "" {
		warnings = append(warnings, "No exact NeoForge coordinate was found for the target version in the embedded official Maven snapshot; do not guess a coordinate.")
	}
	if !target.Server {
		warnings = append(warnings, "The official target manifest has no dedicated-server artifact, so server verification must be adapted and explicitly documented.")
	}
	return uniqueStringsPreserve(warnings)
}

func portingRisk(direction, sourceMode string, source, target VersionAtlasRecord, sourceLoader, targetLoader string) string {
	score := 0
	if direction == "downgrade" {
		score += 4
	} else if direction == "upgrade" {
		score += 2
	}
	if sourceMode == "binary" {
		score += 4
	}
	if sourceLoader != targetLoader {
		score += 3
	}
	javaDelta := javaMajor(target) - javaMajor(source)
	if javaDelta < 0 {
		javaDelta = -javaDelta
	}
	if javaDelta >= 8 {
		score += 2
	} else if javaDelta > 0 {
		score++
	}
	if source.ClientMappings != target.ClientMappings {
		score++
	}
	switch {
	case score >= 8:
		return "critical"
	case score >= 5:
		return "high"
	case score >= 2:
		return "moderate"
	default:
		return "low"
	}
}

func portingTools(atlas VersionAtlas, source, target VersionAtlasRecord, sourceLoader, targetLoader, sourceMode string, pins map[string]string) []PortingToolRecommendation {
	tools := []PortingToolRecommendation{
		{ID: "intermed", Name: "InterMed", Category: "static-analysis", Maturity: "alpha", OfficialURL: "https://github.com/jarettr/intermed", Required: true, Reason: "Preflight bytecode, metadata, data/resource, mixin, security, SBOM, log, and performance signals before changing the artifact.", Guardrail: "Treat findings as evidence to verify, not automatic rewrites."},
		{ID: "modcrawl", Name: "modcrawl", Category: "artifact-analysis", Maturity: "active", OfficialURL: "https://github.com/SirCesarium/modcrawl", Required: false, Reason: "Independent fast JAR metadata, dependency, class, mixin, and duplicate inspection for cross-checking the Vault."},
	}
	add := func(tool PortingToolRecommendation) {
		for _, existing := range tools {
			if existing.ID == tool.ID {
				return
			}
		}
		tools = append(tools, tool)
	}
	if sourceMode == "binary" {
		add(PortingToolRecommendation{ID: "vineflower", Name: "Vineflower", Category: "decompiler", Maturity: "stable", OfficialURL: "https://github.com/Vineflower/vineflower", Required: true, Reason: "Primary modern Java decompiler for source reconstruction; compare its output against a second decompiler and bytecode."})
		add(PortingToolRecommendation{ID: "cfr", Name: "CFR", Category: "decompiler", Maturity: "stable", OfficialURL: "https://github.com/leibnitz27/cfr", Required: true, Reason: "Independent decompiler cross-check to detect control-flow or synthetic-code reconstruction errors."})
		add(PortingToolRecommendation{ID: "tiny-remapper", Name: "Tiny Remapper", Category: "mapping", Maturity: "stable", OfficialURL: "https://github.com/FabricMC/tiny-remapper", PinnedVersion: atlas.Toolchains["autoRenamingTool"].LatestStable, Required: true, Reason: "Controlled namespace remapping for compatible mapping graphs.", Guardrail: "Remapping does not replace semantic API migration."})
		add(PortingToolRecommendation{ID: "retromod", Name: "Retromod", Category: "binary-compatibility", Maturity: "experimental", OfficialURL: "https://github.com/Bownlux/Retromod", Required: false, Reason: "Candidate bytecode transforms and compatibility shims for narrow, measured binary gaps.", Guardrail: "Never call a Retromod-transformed JAR ported until source-equivalent behavior passes the full runtime matrix."})
	}
	legacy := javaMajor(source) <= 8 || javaMajor(target) <= 8
	if legacy {
		add(PortingToolRecommendation{ID: "retrofuturagradle", Name: "RetroFuturaGradle", Category: "legacy-build", Maturity: "stable", OfficialURL: "https://github.com/GTNewHorizons/RetroFuturaGradle", Required: sourceLoader == "forge" || targetLoader == "forge", Reason: "Reconstruct old Forge/LaunchWrapper-era builds with maintained Gradle and mapping support."})
		add(PortingToolRecommendation{ID: "unimined", Name: "Unimined", Category: "unified-build", Maturity: "active", OfficialURL: "https://github.com/unimined/unimined", Required: false, Reason: "Broad historical loader coverage makes it a strong fallback when the native era toolchain cannot be reproduced."})
	}
	switch targetLoader {
	case "fabric":
		add(PortingToolRecommendation{ID: "fabric-loom", Name: "Fabric Loom", Category: "build", Maturity: "stable", OfficialURL: "https://github.com/FabricMC/fabric-loom", PinnedVersion: pins["fabricLoom"], Required: true, Reason: "Official Fabric development, mapping, remapping, launch, and packaging toolchain."})
	case "quilt":
		add(PortingToolRecommendation{ID: "architectury-loom", Name: "Architectury Loom", Category: "build", Maturity: "stable", OfficialURL: "https://github.com/architectury/architectury-loom", PinnedVersion: pins["architecturyLoom"], Required: true, Reason: "Maintained Loom fork suitable for Quilt and cross-loader source work."})
	case "forge":
		add(PortingToolRecommendation{ID: "forgegradle", Name: "ForgeGradle", Category: "build", Maturity: "stable", OfficialURL: "https://github.com/MinecraftForge/ForgeGradle", PinnedVersion: pins["forgeGradle"], Required: true, Reason: "Official Forge source build, mappings, run configurations, and reobfuscation pipeline."})
	case "neoforge":
		add(PortingToolRecommendation{ID: "moddevgradle", Name: "ModDevGradle", Category: "build", Maturity: "stable", OfficialURL: "https://github.com/neoforged/ModDevGradle", PinnedVersion: pins["modDevGradle"], Required: true, Reason: "Official NeoForge development plugin with target-version run and test configuration."})
	case "multiloader":
		add(PortingToolRecommendation{ID: "modstitch", Name: "Modstitch", Category: "multi-loader-build", Maturity: "experimental", OfficialURL: "https://github.com/isXander/Modstitch", Required: true, Reason: "High-capability unified Fabric/NeoForge source layout with translation support.", Guardrail: "Pin and harden the selected unstable build; retain a native-loader verification path."})
		add(PortingToolRecommendation{ID: "stonecutter", Name: "Stonecutter", Category: "multi-version-build", Maturity: "stable", OfficialURL: "https://github.com/stonecutter-versioning/stonecutter", Required: true, Reason: "Version-aware source preprocessing and project orchestration for maintained multi-version branches."})
		add(PortingToolRecommendation{ID: "architectury-loom", Name: "Architectury Loom", Category: "multi-loader-build", Maturity: "stable", OfficialURL: "https://github.com/architectury/architectury-loom", PinnedVersion: pins["architecturyLoom"], Required: false, Reason: "Conservative cross-loader fallback when Modstitch translation is not yet proven for the target feature set."})
	}
	if sourceLoader != targetLoader && targetLoader != "multiloader" {
		add(PortingToolRecommendation{ID: "modstitch", Name: "Modstitch", Category: "loader-migration", Maturity: "experimental", OfficialURL: "https://github.com/isXander/Modstitch", Required: false, Reason: "Can reduce duplicated source while translating selected loader declarations and access mechanisms.", Guardrail: "Generated translation must be compared against a native target-loader implementation."})
	}
	add(PortingToolRecommendation{ID: "japicmp", Name: "japicmp", Category: "api-diff", Maturity: "stable", OfficialURL: "https://github.com/siom79/japicmp", Required: true, Reason: "Binary API diffing across dependency and loader versions to turn compile failures into an ordered migration map."})
	add(PortingToolRecommendation{ID: "packwiz", Name: "packwiz", Category: "test-pack", Maturity: "stable", OfficialURL: "https://github.com/packwiz/packwiz", Required: false, Reason: "Reproducible disposable client/server integration packs for matrix testing."})
	add(PortingToolRecommendation{ID: "ferium", Name: "Ferium", Category: "dependency-acquisition", Maturity: "stable", OfficialURL: "https://github.com/gorilla-devs/ferium", Required: false, Reason: "Independent Modrinth/CurseForge/GitHub acquisition path for test fixtures and comparison against provider resolution."})
	return tools
}

func portingPhases(direction, sourceMode, sourceLoader, targetLoader string) []PortingPhase {
	return []PortingPhase{
		{Order: 1, ID: "preserve-identify", Title: "Preserve and identify", Goal: "Freeze exact inputs before any mutation.", Actions: []string{"Hash every input and dependency", "Extract loader metadata, mappings, mixins, access rules, services, native libraries, and source provenance", "Record source and target toolchain coordinates"}, Gates: []string{"Original bytes are unchanged and independently recoverable", "Every input has SHA-256 and SHA-512 identity"}, ToolIDs: []string{"intermed", "modcrawl"}},
		{Order: 2, ID: "reconstruct", Title: "Reconstruct the exact source environment", Goal: "Make the original artifact reproducible before changing versions.", Actions: []string{"Resolve upstream source and commit", "Restore exact Java, Gradle, loader, mappings, libraries, processors, and generated sources", "For binary-only inputs, decompile twice and reconcile against bytecode"}, Gates: []string{"A clean source build reproduces equivalent metadata and behavior", "Unresolved/generated/decompiler artifacts are listed explicitly"}, ToolIDs: []string{"vineflower", "cfr", "retrofuturagradle", "unimined"}},
		{Order: 3, ID: "mapping-namespace", Title: "Migrate mappings and namespaces", Goal: "Translate names without hiding semantic API changes.", Actions: []string{"Select one canonical mapping graph", "Remap source or bytecode in an isolated branch", "Diff APIs before replacing calls"}, Gates: []string{"No unmapped owners/descriptors remain", "Mixin targets, refmaps, access wideners/transformers, and invokedynamic descriptors resolve"}, ToolIDs: []string{"tiny-remapper", "japicmp"}},
		{Order: 4, ID: "loader-api", Title: "Port loader and game APIs", Goal: "Replace lifecycle, registry, networking, rendering, config, data, and side APIs deliberately.", Actions: []string{"Port dependency declarations and entrypoints", "Replace removed or behaviorally changed APIs", "Separate common logic from loader adapters when appropriate"}, Gates: []string{"Compile and static linkage pass on the native target loader", "No compatibility bridge is the sole proof of correctness"}, ToolIDs: []string{"fabric-loom", "forgegradle", "moddevgradle", "modstitch", "architectury-loom", "stonecutter"}},
		{Order: 5, ID: "data-assets", Title: "Transform data and assets", Goal: "Preserve registries, tags, recipes, loot, models, shaders, translations, configs, and saved data.", Actions: []string{"Diff pack/schema versions", "Write explicit forward and reverse transforms", "Measure downgrade loss on disposable copies"}, Gates: []string{"Generated resources validate", "World/data migration and rollback behavior are documented"}},
		{Order: 6, ID: "compile-audit", Title: "Compile and audit", Goal: "Turn every build warning and static finding into an owned result.", Actions: []string{"Run clean build with pinned Java", "Run dependency convergence, duplicate-class, API diff, mixin, security, and SBOM checks", "Reject undeclared or floating dependencies"}, Gates: []string{"Build is clean and repeatable", "No unresolved critical/high finding is silently waived"}, ToolIDs: []string{"intermed", "modcrawl", "japicmp"}},
		{Order: 7, ID: "runtime-matrix", Title: "Exercise real client and server matrices", Goal: "Prove behavior, not merely startup.", Actions: []string{"Launch disposable client and dedicated server", "Exercise networking, configs, registries, world creation/load, rendering, integrations, and highest-risk features", "Inspect fresh logs and crash reports"}, Gates: []string{"Both supported sides pass", "Feature-level acceptance checks pass with no new severe log evidence"}, ToolIDs: []string{"packwiz", "ferium"}},
		{Order: 8, ID: "package-rollback", Title: "Package, compare, and retain rollback", Goal: "Produce a reproducible artifact with complete lineage.", Actions: []string{"Build twice from clean state", "Compare archive contents and hashes", "Write release, migration, loss, and restore receipts"}, Gates: []string{"Rebuilds are byte-identical or differences are explained", "Original and last-known-good artifacts remain restorable"}},
	}
}

func (a *App) handlePortingEnvironment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	target := strings.TrimSpace(r.URL.Query().Get("target"))
	env, err := buildPortingEnvironment(r.Context(), target)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, env)
}

func buildPortingEnvironment(ctx context.Context, target string) (PortingEnvironment, error) {
	atlas, err := loadVersionAtlas()
	if err != nil {
		return PortingEnvironment{}, err
	}
	requiredJava := 0
	if target != "" {
		if row, _, ok := findAtlasVersion(atlas, target); ok {
			requiredJava = javaMajor(row)
		}
	}
	probes := []struct {
		id, executable string
		args           []string
		required       bool
		purpose        string
	}{
		{"java", "java", []string{"-version"}, true, "Minecraft runtime and Gradle toolchain"},
		{"javac", "javac", []string{"-version"}, true, "Source compilation"},
		{"git", "git", []string{"--version"}, true, "Source provenance and reversible patches"},
		{"gradle", "gradle", []string{"--version"}, false, "Fallback when a generated workspace does not yet contain a wrapper"},
		{"node", "node", []string{"--version"}, false, "Optional asset and web tooling"},
		{"python", "python3", []string{"--version"}, false, "Atlas/build helper scripts"},
		{"unzip", "unzip", []string{"-v"}, false, "Archive verification"},
		{"tar", "tar", []string{"--version"}, false, "Source and evidence packaging"},
	}
	environment := PortingEnvironment{CreatedAt: time.Now().UTC().Format(time.RFC3339), OS: runtime.GOOS, Arch: runtime.GOARCH, TargetGameVersion: target, RequiredJava: requiredJava, ReadyForPlanning: true}
	requiredReady := true
	for _, probe := range probes {
		tool := probePortingCommand(ctx, probe.id, probe.executable, probe.args, probe.required, probe.purpose)
		environment.Commands = append(environment.Commands, tool)
		if probe.required && !tool.Installed {
			requiredReady = false
		}
	}
	environment.ReadyForLocalBuild = requiredReady
	if !requiredReady {
		environment.Warnings = append(environment.Warnings, "Java/Javac/Git are required for a local source-build workflow; planning and workspace generation remain available.")
	}
	if requiredJava > 0 {
		javaTool := findEnvironmentTool(environment.Commands, "java")
		if javaTool.Installed {
			major := parseJavaMajor(javaTool.Version)
			if major > 0 && major != requiredJava {
				environment.Warnings = append(environment.Warnings, fmt.Sprintf("Active Java appears to be %d while the official target manifest requests Java %d. Use a Gradle toolchain or an exact JDK installation.", major, requiredJava))
			}
		}
	}
	return environment, nil
}

func probePortingCommand(parent context.Context, id, executable string, args []string, required bool, purpose string) PortingEnvironmentTool {
	tool := PortingEnvironmentTool{ID: id, Required: required, Purpose: purpose}
	path, err := exec.LookPath(executable)
	if err != nil {
		return tool
	}
	tool.Installed = true
	tool.Path = path
	ctx, cancel := context.WithTimeout(parent, 4*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, path, args...)
	output, err := command.CombinedOutput()
	if len(output) > 0 {
		lines := strings.Split(strings.ReplaceAll(string(output), "\r\n", "\n"), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				tool.Version = line
				break
			}
		}
	}
	if err != nil && tool.Version == "" {
		tool.Version = err.Error()
	}
	return tool
}

func findEnvironmentTool(tools []PortingEnvironmentTool, id string) PortingEnvironmentTool {
	for _, tool := range tools {
		if tool.ID == id {
			return tool
		}
	}
	return PortingEnvironmentTool{}
}

var javaVersionPattern = regexp.MustCompile(`(?i)(?:version\s+\")?(\d+)(?:\.(\d+))?`)

func parseJavaMajor(value string) int {
	match := javaVersionPattern.FindStringSubmatch(value)
	if len(match) < 2 {
		return 0
	}
	major, _ := strconv.Atoi(match[1])
	if major == 1 && len(match) > 2 {
		major, _ = strconv.Atoi(match[2])
	}
	return major
}

func (a *App) handlePortingWorkspace(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request PortingWorkspaceRequest
	if err := decodeJSON(r, &request); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	a.dataMu.RLock()
	plan, ok := a.portingPlans[request.PlanID]
	a.dataMu.RUnlock()
	if !ok {
		writeJSON(w, http.StatusNotFound, APIError{Error: "porting plan expired; build the plan again"})
		return
	}
	if strings.TrimSpace(request.ProjectName) != "" {
		plan.ProjectName = request.ProjectName
	}
	if strings.TrimSpace(request.InputJar) != "" {
		requestedInput := filepath.Clean(strings.TrimSpace(request.InputJar))
		plannedInput := filepath.Clean(strings.TrimSpace(plan.InputJar))
		if plannedInput == "." || !strings.EqualFold(requestedInput, plannedInput) {
			writeJSON(w, http.StatusConflict, APIError{Error: "input JAR changed after planning; rebuild the evidence-backed plan before generating a workspace"})
			return
		}
	}
	manifest, err := a.createPortingWorkspace(plan)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, manifest)
}

func (a *App) handlePortingWorkspaces(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	root := filepath.Join(a.cfgDir, "porting-workspaces")
	if err := os.MkdirAll(root, 0o755); err != nil {
		writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
		return
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			writeJSON(w, http.StatusOK, []PortingWorkspaceManifest{})
			return
		}
		writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
		return
	}
	manifests := []PortingWorkspaceManifest{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		data, readErr := os.ReadFile(filepath.Join(root, entry.Name(), "WORKSPACE-MANIFEST.json"))
		if readErr != nil {
			continue
		}
		var manifest PortingWorkspaceManifest
		if json.Unmarshal(data, &manifest) == nil {
			manifests = append(manifests, manifest)
		}
	}
	sort.Slice(manifests, func(i, j int) bool { return manifests[i].CreatedAt > manifests[j].CreatedAt })
	writeJSON(w, http.StatusOK, manifests)
}

func (a *App) createPortingWorkspace(plan PortingPlan) (PortingWorkspaceManifest, error) {
	root := filepath.Join(a.cfgDir, "porting-workspaces")
	if err := os.MkdirAll(root, 0o755); err != nil {
		return PortingWorkspaceManifest{}, err
	}
	project := safePortingSlug(plan.ProjectName)
	if project == "" {
		project = "minecraft-port"
	}
	id := time.Now().UTC().Format("20060102T150405Z") + "-" + randomToken(4)
	workspace := filepath.Join(root, id+"-"+project)
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		return PortingWorkspaceManifest{}, err
	}
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.RemoveAll(workspace)
		}
	}()

	manifest := PortingWorkspaceManifest{
		SchemaVersion: 1, ID: id, CreatedAt: time.Now().UTC().Format(time.RFC3339), Path: workspace, PlanID: plan.ID,
		ProjectName: project, Source: plan.Source.ID, Target: plan.Target.ID, SourceLoader: plan.SourceLoader, TargetLoader: plan.TargetLoader,
		Status: "generated-awaiting-source-and-runtime-verification", Rollback: "Delete this isolated workspace; the installed mod was not changed.",
		Evidence: map[string]string{"versionAtlasSnapshot": plan.EvidenceSnapshotAt, "planRisk": plan.Risk},
	}
	if strings.TrimSpace(plan.InputJar) != "" {
		input, err := a.copyPortingInput(workspace, plan.InputJar)
		if err != nil {
			return PortingWorkspaceManifest{}, err
		}
		if plan.InputAnalysis != nil && !strings.EqualFold(input.SHA256, plan.InputAnalysis.SHA256) {
			return PortingWorkspaceManifest{}, fmt.Errorf("input JAR changed since planning: expected SHA-256 %s, got %s", plan.InputAnalysis.SHA256, input.SHA256)
		}
		manifest.Input = &input
		manifest.Evidence["inputSHA256"] = input.SHA256
		if plan.InputAnalysis != nil {
			manifest.Evidence["inputSHA512"] = plan.InputAnalysis.SHA512
			manifest.Evidence["inputDetectedLoaders"] = strings.Join(plan.InputAnalysis.DetectedLoaders, ",")
			manifest.Evidence["inputModIDs"] = strings.Join(plan.InputAnalysis.ModIDs, ",")
		}
	}

	if err := writeJSONFile(filepath.Join(workspace, "PORTING-PLAN.json"), plan); err != nil {
		return PortingWorkspaceManifest{}, err
	}
	if err := os.WriteFile(filepath.Join(workspace, "PORTING-PLAN.md"), []byte(renderPortingPlanMarkdown(plan)), 0o644); err != nil {
		return PortingWorkspaceManifest{}, err
	}
	if err := os.WriteFile(filepath.Join(workspace, "README.md"), []byte(renderWorkspaceReadme(plan)), 0o644); err != nil {
		return PortingWorkspaceManifest{}, err
	}
	if err := os.WriteFile(filepath.Join(workspace, "gradle.properties"), []byte(renderGradleProperties(plan)), 0o644); err != nil {
		return PortingWorkspaceManifest{}, err
	}
	if err := os.WriteFile(filepath.Join(workspace, ".gitignore"), []byte(".gradle/\nbuild/\nrun/\n.idea/\n*.iml\n.gradle-user-home/\n"), 0o644); err != nil {
		return PortingWorkspaceManifest{}, err
	}
	if err := writeWorkspaceBuildFiles(workspace, plan); err != nil {
		return PortingWorkspaceManifest{}, err
	}
	if err := writeWorkspaceScripts(workspace, plan); err != nil {
		return PortingWorkspaceManifest{}, err
	}
	for _, dir := range []string{"src/main/java", "src/main/resources", "src/test/java", "evidence/logs", "evidence/reports", "patches"} {
		if err := os.MkdirAll(filepath.Join(workspace, filepath.FromSlash(dir)), 0o755); err != nil {
			return PortingWorkspaceManifest{}, err
		}
	}
	for _, keep := range []string{"src/main/java/.gitkeep", "src/main/resources/.gitkeep", "src/test/java/.gitkeep", "evidence/logs/.gitkeep", "evidence/reports/.gitkeep", "patches/.gitkeep"} {
		if err := os.WriteFile(filepath.Join(workspace, filepath.FromSlash(keep)), nil, 0o644); err != nil {
			return PortingWorkspaceManifest{}, err
		}
	}

	files, err := collectWorkspaceFiles(workspace, "WORKSPACE-MANIFEST.json")
	if err != nil {
		return PortingWorkspaceManifest{}, err
	}
	manifest.Files = files
	if runtime.GOOS == "windows" {
		manifest.NextCommand = `.\scripts\verify.ps1`
	} else {
		manifest.NextCommand = `./scripts/verify.sh`
	}
	if err := writeJSONFile(filepath.Join(workspace, "WORKSPACE-MANIFEST.json"), manifest); err != nil {
		return PortingWorkspaceManifest{}, err
	}
	cleanup = false
	return manifest, nil
}

func (a *App) copyPortingInput(workspace, inputPath string) (PortingWorkspaceInput, error) {
	inputPath = filepath.Clean(strings.TrimSpace(inputPath))
	if !a.allowedManagedPath(inputPath) {
		return PortingWorkspaceInput{}, errors.New("input JAR must be inside a configured Minecraft mods or Vault downloads directory")
	}
	info, err := os.Lstat(inputPath)
	if err != nil {
		return PortingWorkspaceInput{}, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return PortingWorkspaceInput{}, errors.New("symbolic-link inputs are refused; select the real JAR")
	}
	if !info.Mode().IsRegular() {
		return PortingWorkspaceInput{}, errors.New("input is not a regular file")
	}
	lower := strings.ToLower(filepath.Base(inputPath))
	if !strings.HasSuffix(lower, ".jar") && !strings.HasSuffix(lower, ".jar.disabled") {
		return PortingWorkspaceInput{}, errors.New("input must be a .jar or .jar.disabled file")
	}
	inputDir := filepath.Join(workspace, "inputs")
	if err := os.MkdirAll(inputDir, 0o755); err != nil {
		return PortingWorkspaceInput{}, err
	}
	destination := filepath.Join(inputDir, safeFilename(filepath.Base(inputPath)))
	if err := copyFilePreserve(inputPath, destination); err != nil {
		return PortingWorkspaceInput{}, err
	}
	digest, size, err := fileSHA256(destination)
	if err != nil {
		return PortingWorkspaceInput{}, err
	}
	return PortingWorkspaceInput{OriginalPath: inputPath, WorkspacePath: destination, Filename: filepath.Base(destination), Size: size, SHA256: digest}, nil
}

func copyFilePreserve(source, destination string) error {
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	info, err := in.Stat()
	if err != nil {
		return err
	}
	out, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY|os.O_EXCL, info.Mode().Perm())
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func writeWorkspaceBuildFiles(workspace string, plan PortingPlan) error {
	settings := renderSettingsGradle(plan)
	if err := os.WriteFile(filepath.Join(workspace, "settings.gradle.kts"), []byte(settings), 0o644); err != nil {
		return err
	}
	if plan.TargetLoader == "forge" {
		return os.WriteFile(filepath.Join(workspace, "build.gradle"), []byte(renderForgeGradle(plan)), 0o644)
	}
	return os.WriteFile(filepath.Join(workspace, "build.gradle.kts"), []byte(renderKotlinGradle(plan)), 0o644)
}

func renderSettingsGradle(plan PortingPlan) string {
	return fmt.Sprintf(`pluginManagement {
    repositories {
        gradlePluginPortal()
        mavenCentral()
        maven("https://maven.fabricmc.net/")
        maven("https://maven.neoforged.net/releases/")
        maven("https://maven.minecraftforge.net/")
        maven("https://maven.architectury.dev/")
    }
}

dependencyResolutionManagement {
    repositoriesMode.set(RepositoriesMode.PREFER_SETTINGS)
    repositories {
        mavenCentral()
        maven("https://libraries.minecraft.net/")
        maven("https://maven.fabricmc.net/")
        maven("https://maven.neoforged.net/releases/")
        maven("https://maven.minecraftforge.net/")
        maven("https://maven.quiltmc.org/repository/release/")
    }
}

rootProject.name = %q
`, safePortingSlug(plan.ProjectName))
}

func renderKotlinGradle(plan PortingPlan) string {
	java := javaMajor(plan.Target)
	common := fmt.Sprintf(`

group = "dev.minecraftmodvault.port"
version = "0.1.0-port"

java {
    toolchain.languageVersion.set(JavaLanguageVersion.of(%d))
    withSourcesJar()
}

tasks.withType<JavaCompile>().configureEach {
    options.release.set(%d)
    options.encoding = "UTF-8"
}

tasks.register("mmvEvidence") {
    doLast {
        println("MMV plan: %s")
        println("Target: Minecraft %s / %s / Java %d")
        println("A successful build is not a completed port; run the client/server verification matrix in PORTING-PLAN.md.")
    }
}
`, java, java, plan.ID, plan.Target.ID, plan.TargetLoader, java)
	switch plan.TargetLoader {
	case "fabric":
		return fmt.Sprintf(`plugins {
    id("fabric-loom") version %q
    java
}

dependencies {
    minecraft("com.mojang:minecraft:%s")
    mappings(loom.officialMojangMappings())
    modImplementation("net.fabricmc:fabric-loader:%s")
}
`, plan.Pins["fabricLoom"], plan.Target.ID, plan.Pins["fabricLoader"]) + common
	case "quilt":
		return fmt.Sprintf(`plugins {
    id("dev.architectury.loom") version %q
    java
}

dependencies {
    minecraft("com.mojang:minecraft:%s")
    mappings(loom.officialMojangMappings())
    modImplementation("org.quiltmc:quilt-loader:%s")
}
`, plan.Pins["architecturyLoom"], plan.Target.ID, plan.Pins["quiltLoader"]) + common
	case "neoforge":
		return fmt.Sprintf(`plugins {
    id("net.neoforged.moddev") version %q
    java
}

neoForge {
    version = %q
}
`, plan.Pins["modDevGradle"], plan.Pins["neoForge"]) + common
	case "multiloader":
		return `plugins {
    java
}

// The plan selected a multi-loader target. Keep this root build deliberately
// minimal until common, Fabric, and NeoForge modules are created from native
// templates. PORTING-PLAN.md records Modstitch/Stonecutter and Architectury
// alternatives; do not guess plugin coordinates or claim translation parity.
` + common
	default:
		return `plugins {
    java
}
` + common
	}
}

func renderForgeGradle(plan PortingPlan) string {
	return fmt.Sprintf(`buildscript {
    repositories {
        mavenCentral()
        maven { url = 'https://maven.minecraftforge.net/' }
    }
    dependencies {
        classpath 'net.minecraftforge.gradle:ForgeGradle:%s'
    }
}

apply plugin: 'net.minecraftforge.gradle'
apply plugin: 'java'

group = 'dev.minecraftmodvault.port'
version = '0.1.0-port'

java {
    toolchain.languageVersion = JavaLanguageVersion.of(%d)
    withSourcesJar()
}

minecraft {
    mappings channel: 'official', version: '%s'
}

dependencies {
    minecraft 'net.minecraftforge:forge:%s'
}

tasks.withType(JavaCompile).configureEach {
    options.release = %d
    options.encoding = 'UTF-8'
}

tasks.register('mmvEvidence') {
    doLast {
        println 'MMV plan: %s'
        println 'A successful build is not a completed port; run the client/server verification matrix in PORTING-PLAN.md.'
    }
}
`, plan.Pins["forgeGradle"], javaMajor(plan.Target), plan.Target.ID, plan.Pins["forge"], javaMajor(plan.Target), plan.ID)
}

func renderGradleProperties(plan PortingPlan) string {
	return fmt.Sprintf("org.gradle.jvmargs=-Xmx4G -Dfile.encoding=UTF-8\norg.gradle.parallel=true\norg.gradle.caching=true\norg.gradle.configuration-cache=false\nmmv.planId=%s\nmmv.source=%s\nmmv.target=%s\n", plan.ID, plan.Source.ID, plan.Target.ID)
}

func writeWorkspaceScripts(workspace string, plan PortingPlan) error {
	dir := filepath.Join(workspace, "scripts")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	sh := fmt.Sprintf(`#!/usr/bin/env sh
set -eu
ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$ROOT"
echo "Minecraft Mod Vault Porting Lab plan %s"
echo "Target Minecraft %s / %s / Java %d"
java -version
javac -version
if [ -x ./gradlew ]; then
  ./gradlew --no-daemon clean build mmvEvidence --stacktrace
elif command -v gradle >/dev/null 2>&1; then
  gradle --no-daemon clean build mmvEvidence --stacktrace
else
  echo "No Gradle wrapper or system Gradle found. Generate a version-pinned wrapper before building." >&2
  exit 2
fi
mkdir -p evidence/reports
find build -type f -maxdepth 5 -print 2>/dev/null | sort > evidence/reports/build-files.txt || true
echo "Build evidence collected. This does not replace client/server feature verification."
`, plan.ID, plan.Target.ID, plan.TargetLoader, javaMajor(plan.Target))
	if err := os.WriteFile(filepath.Join(dir, "verify.sh"), []byte(sh), 0o755); err != nil {
		return err
	}
	ps := fmt.Sprintf(`$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
Set-Location $Root
Write-Host "Minecraft Mod Vault Porting Lab plan %s"
Write-Host "Target Minecraft %s / %s / Java %d"
& java -version
& javac -version
if (Test-Path .\gradlew.bat) {
  & .\gradlew.bat --no-daemon clean build mmvEvidence --stacktrace
} elseif (Get-Command gradle -ErrorAction SilentlyContinue) {
  & gradle --no-daemon clean build mmvEvidence --stacktrace
} else {
  throw "No Gradle wrapper or system Gradle found. Generate a version-pinned wrapper before building."
}
New-Item -ItemType Directory -Force evidence\reports | Out-Null
Get-ChildItem build -Recurse -File -ErrorAction SilentlyContinue | Sort-Object FullName | Select-Object -ExpandProperty FullName | Set-Content evidence\reports\build-files.txt
Write-Host "Build evidence collected. This does not replace client/server feature verification."
`, plan.ID, plan.Target.ID, plan.TargetLoader, javaMajor(plan.Target))
	return os.WriteFile(filepath.Join(dir, "verify.ps1"), []byte(ps), 0o644)
}

func renderWorkspaceReadme(plan PortingPlan) string {
	return fmt.Sprintf(`# Minecraft Mod Vault Porting Lab Workspace

This isolated workspace was generated for plan **%s**.

- Source: Minecraft %s / %s
- Target: Minecraft %s / %s
- Direction: %s
- Risk: %s
- Required Java: %d
- Evidence snapshot: %s

## Start here

1. Read PORTING-PLAN.md.
2. Resolve the exact upstream source and commit, or document why only binary reconstruction is possible.
3. Place reconstructed or upstream source under src/ without altering the original input in inputs/.
4. Generate and pin a Gradle wrapper compatible with the selected native loader toolchain.
5. Run scripts/verify.sh or scripts/verify.ps1.
6. Retain build, client, server, gameplay, data-migration, and rollback evidence under evidence/.

A successful remap, compile, or startup is not a completed port. Completion requires the behavior and rollback gates in PORTING-PLAN.md.
`, plan.ID, plan.Source.ID, plan.SourceLoader, plan.Target.ID, plan.TargetLoader, plan.Direction, plan.Risk, javaMajor(plan.Target), plan.EvidenceSnapshotAt)
}

func renderPortingPlanMarkdown(plan PortingPlan) string {
	var out strings.Builder
	fmt.Fprintf(&out, "# Porting Plan %s\n\n", plan.ID)
	fmt.Fprintf(&out, "**%s / %s -> %s / %s**  \nDirection: **%s**  \nRisk: **%s**  \nSource mode: **%s**  \nEvidence snapshot: **%s**\n\n", plan.Source.ID, plan.SourceLoader, plan.Target.ID, plan.TargetLoader, plan.Direction, plan.Risk, plan.SourceMode, plan.EvidenceSnapshotAt)
	out.WriteString("## Exact pins\n\n")
	keys := make([]string, 0, len(plan.Pins))
	for key := range plan.Pins {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if plan.Pins[key] != "" {
			fmt.Fprintf(&out, "- `%s`: `%s`\n", key, plan.Pins[key])
		}
	}
	out.WriteString("\n## Boundaries\n\n")
	for _, item := range plan.Boundaries {
		fmt.Fprintf(&out, "- %s\n", item)
	}
	out.WriteString("\n## Warnings\n\n")
	for _, item := range plan.Warnings {
		fmt.Fprintf(&out, "- %s\n", item)
	}
	out.WriteString("\n## Ordered phases\n\n")
	for _, phase := range plan.Phases {
		fmt.Fprintf(&out, "### %d. %s\n\n%s\n\nActions:\n", phase.Order, phase.Title, phase.Goal)
		for _, item := range phase.Actions {
			fmt.Fprintf(&out, "- %s\n", item)
		}
		out.WriteString("\nAcceptance gates:\n")
		for _, item := range phase.Gates {
			fmt.Fprintf(&out, "- [ ] %s\n", item)
		}
		out.WriteString("\n")
	}
	out.WriteString("## Runtime verification matrix\n\n")
	for _, item := range plan.VerificationMatrix {
		fmt.Fprintf(&out, "- [ ] %s\n", item)
	}
	out.WriteString("\n## Completion definition\n\n")
	for _, item := range plan.CompletionDefinition {
		fmt.Fprintf(&out, "- [ ] %s\n", item)
	}
	return out.String()
}

func safePortingSlug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var out strings.Builder
	lastDash := false
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			out.WriteRune(r)
			lastDash = false
		} else if !lastDash && out.Len() > 0 {
			out.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(out.String(), "-")
}

func writeJSONFile(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func fileSHA256(path string) (string, int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()
	hash := sha256.New()
	size, err := io.Copy(hash, file)
	if err != nil {
		return "", 0, err
	}
	return hex.EncodeToString(hash.Sum(nil)), size, nil
}

func collectWorkspaceFiles(root, excludedBase string) ([]PortingWorkspaceFile, error) {
	files := []PortingWorkspaceFile{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || entry.Name() == excludedBase {
			return nil
		}
		digest, size, digestErr := fileSHA256(path)
		if digestErr != nil {
			return digestErr
		}
		relative, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		files = append(files, PortingWorkspaceFile{Path: filepath.ToSlash(relative), Size: size, SHA256: digest})
		return nil
	})
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files, err
}

// compile-time use of bytes keeps workspace rendering tests able to compare
// normalized output without pulling in an external diff dependency.
func normalizedWorkspaceBytes(value []byte) []byte {
	return bytes.ReplaceAll(value, []byte("\r\n"), []byte("\n"))
}
