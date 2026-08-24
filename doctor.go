package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	doctorKnowledgePath  = "assets/mod-doctor-knowledge.json"
	doctorClassReadLimit = int64(8 << 20)
	doctorTotalReadLimit = int64(96 << 20)
)

type DoctorKnowledge struct {
	SchemaVersion int            `json:"schemaVersion"`
	ReviewedAt    string         `json:"reviewedAt"`
	Sources       []DoctorSource `json:"sources"`
	Categories    []string       `json:"categories,omitempty"`
	Total         int            `json:"total"`
}

type DoctorSource struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	URL         string `json:"url"`
	Repository  string `json:"repository,omitempty"`
	Maturity    string `json:"maturity"`
	Status      string `json:"status"`
	BestFor     string `json:"bestFor"`
	Integration string `json:"integration"`
	Priority    int    `json:"priority"`
	Notes       string `json:"notes"`
}

type DoctorTool struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Category        string   `json:"category"`
	Maturity        string   `json:"maturity"`
	Capability      string   `json:"capability"`
	BestFor         string   `json:"bestFor"`
	OfficialURL     string   `json:"officialUrl"`
	RepositoryURL   string   `json:"repositoryUrl,omitempty"`
	Integration     string   `json:"integration"`
	CurrentEvidence string   `json:"currentEvidence,omitempty"`
	Risks           []string `json:"risks,omitempty"`
	Priority        int      `json:"priority,omitempty"`
}

type DoctorRepairPattern struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	FailureClass string   `json:"failureClass"`
	Trigger      string   `json:"trigger"`
	Repair       string   `json:"repair"`
	Verification []string `json:"verification"`
	Confidence   string   `json:"confidence"`
}

type DoctorToolsPayload struct {
	ReviewedAt     string                `json:"reviewedAt"`
	Total          int                   `json:"total"`
	Categories     []string              `json:"categories"`
	Tools          []DoctorTool          `json:"tools"`
	RepairPatterns []DoctorRepairPattern `json:"repairPatterns"`
}

type DoctorDependency struct {
	Owner        string `json:"owner,omitempty"`
	ModID        string `json:"modId"`
	Version      string `json:"version,omitempty"`
	Relationship string `json:"relationship"`
	Side         string `json:"side,omitempty"`
	Source       string `json:"source"`
}

type DoctorDependencyEdge struct {
	ModID        string `json:"modId"`
	DependsOn    string `json:"dependsOn"`
	Version      string `json:"version,omitempty"`
	Relationship string `json:"relationship"`
	Source       string `json:"source,omitempty"`
}

type DoctorPortingDecision struct {
	PrimaryRoute string   `json:"primaryRoute"`
	Confidence   string   `json:"confidence"`
	Reasons      []string `json:"reasons"`
	SourceIDs    []string `json:"sourceIds,omitempty"`
}

type DoctorLogRequest struct {
	Text        string `json:"text"`
	Filename    string `json:"filename,omitempty"`
	GameVersion string `json:"gameVersion,omitempty"`
	Loader      string `json:"loader,omitempty"`
}

type DoctorLogReport struct {
	CreatedAt            string                `json:"createdAt"`
	Filename             string                `json:"filename,omitempty"`
	GameVersion          string                `json:"gameVersion,omitempty"`
	Loader               string                `json:"loader,omitempty"`
	RootCause            DoctorFinding         `json:"rootCause"`
	Findings             []DoctorFinding       `json:"findings"`
	EvidenceLines        []string              `json:"evidenceLines,omitempty"`
	RepairPatterns       []DoctorRepairPattern `json:"repairPatterns,omitempty"`
	RecommendedSourceIDs []string              `json:"recommendedSourceIds,omitempty"`
	Confidence           string                `json:"confidence"`
}

type DoctorScanRequest struct {
	SourceGameVersion string `json:"sourceGameVersion,omitempty"`
	SourceLoader      string `json:"sourceLoader,omitempty"`
	TargetGameVersion string `json:"targetGameVersion,omitempty"`
	TargetLoader      string `json:"targetLoader,omitempty"`
}

type DoctorScan struct {
	ID                   string                 `json:"id"`
	CreatedAt            string                 `json:"createdAt"`
	KnowledgeReviewedAt  string                 `json:"knowledgeReviewedAt"`
	SourceGameVersion    string                 `json:"sourceGameVersion"`
	SourceLoader         string                 `json:"sourceLoader"`
	TargetGameVersion    string                 `json:"targetGameVersion"`
	TargetLoader         string                 `json:"targetLoader"`
	TargetJava           int                    `json:"targetJava"`
	ModsDirectory        string                 `json:"modsDirectory"`
	Summary              DoctorScanSummary      `json:"summary"`
	GlobalFindings       []DoctorFinding        `json:"globalFindings"`
	MigrationPlan        []DoctorStep           `json:"migrationPlan"`
	RecommendedSourceIDs []string               `json:"recommendedSourceIds"`
	DependencyOrder      []string               `json:"dependencyOrder,omitempty"`
	DependencyCycles     [][]string             `json:"dependencyCycles,omitempty"`
	DependencyEdges      []DoctorDependencyEdge `json:"dependencyEdges,omitempty"`
	Mods                 []DoctorModReport      `json:"mods"`
	Errors               []DoctorScanError      `json:"errors,omitempty"`
}

type DoctorScanSummary struct {
	TotalMods                   int `json:"totalMods"`
	LowRisk                     int `json:"lowRisk"`
	ModerateRisk                int `json:"moderateRisk"`
	HighRisk                    int `json:"highRisk"`
	CriticalRisk                int `json:"criticalRisk"`
	SourceAvailable             int `json:"sourceAvailable"`
	BinaryOnly                  int `json:"binaryOnly"`
	MixinMods                   int `json:"mixinMods"`
	Coremods                    int `json:"coremods"`
	SignedMods                  int `json:"signedMods"`
	NativeMods                  int `json:"nativeMods"`
	MaxRiskScore                int `json:"maxRiskScore"`
	DuplicateModIDs             int `json:"duplicateModIds"`
	ExactDuplicateJars          int `json:"exactDuplicateJars"`
	ExactDuplicateClasses       int `json:"exactDuplicateClasses"`
	ConflictingDuplicateClasses int `json:"conflictingDuplicateClasses"`
	MissingRequiredDependencies int `json:"missingRequiredDependencies"`
	DeclaredConflicts           int `json:"declaredConflicts"`
	DependencyCycles            int `json:"dependencyCycles"`
	CriticalFindings            int `json:"criticalFindings"`
	HighFindings                int `json:"highFindings"`
	WarningFindings             int `json:"warningFindings"`
	InfoFindings                int `json:"infoFindings"`
}

type DoctorScanError struct {
	Filename string `json:"filename"`
	Error    string `json:"error"`
}

type DoctorModReport struct {
	Local                LocalModFile          `json:"local"`
	Signals              JarSignals            `json:"signals"`
	RiskScore            int                   `json:"riskScore"`
	RiskLevel            string                `json:"riskLevel"`
	MigrationDirection   string                `json:"migrationDirection"`
	Porting              DoctorPortingDecision `json:"porting"`
	Findings             []DoctorFinding       `json:"findings"`
	Plan                 []DoctorStep          `json:"plan"`
	RecommendedSourceIDs []string              `json:"recommendedSourceIds"`
}

type JarSignals struct {
	EntryCount             int                `json:"entryCount"`
	ModIDs                 []string           `json:"modIds,omitempty"`
	Dependencies           []DoctorDependency `json:"dependencies,omitempty"`
	ClassEntries           []string           `json:"classEntries,omitempty"`
	ClassDigests           map[string]string  `json:"-"`
	ClassCount             int                `json:"classCount"`
	MaxClassMajor          int                `json:"maxClassMajor,omitempty"`
	MaxJava                int                `json:"maxJava,omitempty"`
	MappingNamespace       string             `json:"mappingNamespace,omitempty"`
	ManifestAttributes     []string           `json:"manifestAttributes,omitempty"`
	MixinConfigs           []string           `json:"mixinConfigs,omitempty"`
	MixinPlugins           []string           `json:"mixinPlugins,omitempty"`
	MixinRefmaps           []string           `json:"mixinRefmaps,omitempty"`
	AccessWideners         []string           `json:"accessWideners,omitempty"`
	AccessTransformers     []string           `json:"accessTransformers,omitempty"`
	Coremods               []string           `json:"coremods,omitempty"`
	TransformationServices []string           `json:"transformationServices,omitempty"`
	ServiceProviders       []string           `json:"serviceProviders,omitempty"`
	NestedJars             []string           `json:"nestedJars,omitempty"`
	NativeLibraries        []string           `json:"nativeLibraries,omitempty"`
	SignatureFiles         []string           `json:"signatureFiles,omitempty"`
	LoaderAPIs             []string           `json:"loaderApis,omitempty"`
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
	ReadBytes              int64              `json:"readBytes"`
	TruncatedAnalysis      bool               `json:"truncatedAnalysis,omitempty"`
}

type DoctorFinding struct {
	ID        string   `json:"id"`
	Code      string   `json:"code,omitempty"`
	Severity  string   `json:"severity"`
	Category  string   `json:"category"`
	Title     string   `json:"title"`
	Evidence  string   `json:"evidence"`
	Action    string   `json:"action"`
	SourceIDs []string `json:"sourceIds,omitempty"`
}

type DoctorStep struct {
	Order     int      `json:"order"`
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Action    string   `json:"action"`
	Why       string   `json:"why"`
	Mode      string   `json:"mode"`
	SourceIDs []string `json:"sourceIds,omitempty"`
}

type mixinConfigProbe struct {
	Required   bool   `json:"required"`
	MinVersion string `json:"minVersion"`
	Package    string `json:"package"`
	Plugin     string `json:"plugin"`
	Refmap     string `json:"refmap"`
}

type gameVersionKey struct {
	Era   int
	Parts [4]int
	Valid bool
}

func loadDoctorKnowledge() (DoctorKnowledge, error) {
	b, err := embeddedFiles.ReadFile(doctorKnowledgePath)
	if err != nil {
		return DoctorKnowledge{}, err
	}
	var knowledge DoctorKnowledge
	if err := json.Unmarshal(b, &knowledge); err != nil {
		return DoctorKnowledge{}, err
	}
	knowledge.Total = len(knowledge.Sources)
	cats := map[string]bool{}
	for _, source := range knowledge.Sources {
		if strings.TrimSpace(source.Category) != "" {
			cats[source.Category] = true
		}
	}
	for category := range cats {
		knowledge.Categories = append(knowledge.Categories, category)
	}
	sort.Strings(knowledge.Categories)
	sort.SliceStable(knowledge.Sources, func(i, j int) bool {
		if knowledge.Sources[i].Priority != knowledge.Sources[j].Priority {
			return knowledge.Sources[i].Priority > knowledge.Sources[j].Priority
		}
		return strings.ToLower(knowledge.Sources[i].Name) < strings.ToLower(knowledge.Sources[j].Name)
	})
	return knowledge, nil
}

func (a *App) handleDoctorSources(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	knowledge, err := loadDoctorKnowledge()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
		return
	}
	category := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("category")))
	status := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("status")))
	query := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("q")))
	if category == "" && status == "" && query == "" {
		writeJSON(w, http.StatusOK, knowledge)
		return
	}
	filtered := make([]DoctorSource, 0, len(knowledge.Sources))
	for _, source := range knowledge.Sources {
		if category != "" && !strings.EqualFold(source.Category, category) {
			continue
		}
		if status != "" && !strings.EqualFold(source.Status, status) {
			continue
		}
		if query != "" {
			haystack := strings.ToLower(strings.Join([]string{source.Name, source.Category, source.BestFor, source.Notes, source.Repository, source.URL}, " "))
			if !strings.Contains(haystack, query) {
				continue
			}
		}
		filtered = append(filtered, source)
	}
	knowledge.Sources = filtered
	knowledge.Total = len(filtered)
	writeJSON(w, http.StatusOK, knowledge)
}

func (a *App) handleDoctorScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request DoctorScanRequest
	if r.Body != nil && r.ContentLength != 0 {
		if err := decodeJSON(r, &request); err != nil {
			writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
			return
		}
	}
	a.mu.RLock()
	settings := a.settings
	a.mu.RUnlock()
	request.SourceGameVersion = firstNonEmpty(strings.TrimSpace(request.SourceGameVersion), settings.GameVersion, "1.21.1")
	request.SourceLoader = normalizeDoctorLoader(firstNonEmpty(strings.TrimSpace(request.SourceLoader), settings.Loader, "fabric"))
	request.TargetGameVersion = firstNonEmpty(strings.TrimSpace(request.TargetGameVersion), request.SourceGameVersion)
	request.TargetLoader = normalizeDoctorLoader(firstNonEmpty(strings.TrimSpace(request.TargetLoader), request.SourceLoader))
	if !parseGameVersionKey(request.SourceGameVersion).Valid || !parseGameVersionKey(request.TargetGameVersion).Valid {
		writeJSON(w, http.StatusBadRequest, APIError{Error: "source and target Minecraft versions must begin with numeric version components"})
		return
	}
	if request.SourceLoader == "" || request.TargetLoader == "" {
		writeJSON(w, http.StatusBadRequest, APIError{Error: "source and target loaders are required"})
		return
	}
	scan, err := a.buildDoctorScan(request)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, scan)
}

func (a *App) buildDoctorScan(request DoctorScanRequest) (DoctorScan, error) {
	knowledge, err := loadDoctorKnowledge()
	if err != nil {
		return DoctorScan{}, err
	}
	modsDir := a.javaTargetDir("mods")
	scan := DoctorScan{
		ID:                  randomToken(12),
		CreatedAt:           time.Now().UTC().Format(time.RFC3339),
		KnowledgeReviewedAt: knowledge.ReviewedAt,
		SourceGameVersion:   request.SourceGameVersion,
		SourceLoader:        normalizeDoctorLoader(request.SourceLoader),
		TargetGameVersion:   request.TargetGameVersion,
		TargetLoader:        normalizeDoctorLoader(request.TargetLoader),
		TargetJava:          targetJavaForMinecraft(request.TargetGameVersion),
		ModsDirectory:       modsDir,
	}
	entries, err := os.ReadDir(modsDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			scan.GlobalFindings = append(scan.GlobalFindings, DoctorFinding{
				ID: "mods-directory-missing", Severity: "warning", Category: "instance", Title: "Mods directory does not exist yet",
				Evidence: modsDir, Action: "Configure the Java instance root or create the mods directory before scanning installed mods.",
			})
			scan.MigrationPlan = buildGlobalDoctorPlan(scan)
			scan.RecommendedSourceIDs = collectDoctorSourceIDs(scan.GlobalFindings, scan.MigrationPlan, nil)
			return scan, nil
		}
		return DoctorScan{}, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		low := strings.ToLower(name)
		if !strings.HasSuffix(low, ".jar") && !strings.HasSuffix(low, ".jar.disabled") {
			continue
		}
		path := filepath.Join(modsDir, name)
		local, inspectErr := inspectLocalJar(path)
		if inspectErr != nil {
			info, _ := entry.Info()
			local = LocalModFile{Path: path, Filename: name, Enabled: !strings.HasSuffix(low, ".disabled")}
			if info != nil {
				local.Size = info.Size()
			}
			report := DoctorModReport{
				Local: local, RiskScore: 100, RiskLevel: "critical", MigrationDirection: migrationDirection(request.SourceGameVersion, request.TargetGameVersion),
				Porting:  DoctorPortingDecision{PrimaryRoute: "recover-artifact", Confidence: "high", Reasons: []string{"The input is not a readable JAR."}},
				Findings: []DoctorFinding{{ID: "invalid-jar", Severity: "critical", Category: "archive", Title: "JAR could not be inspected", Evidence: inspectErr.Error(), Action: "Recover a valid copy before attempting any version migration or binary patch."}},
				Plan:     []DoctorStep{{Order: 1, ID: "recover-valid-artifact", Title: "Recover a valid original", Action: "Restore this file from a verified provider release, source build, instance backup, or known-good archive before continuing.", Why: "A corrupt or non-ZIP JAR cannot be safely remapped, decompiled, or patched.", Mode: "manual-review", SourceIDs: []string{"modrinth-api", "curseforge-api", "github-releases"}}},
			}
			report.RecommendedSourceIDs = collectDoctorSourceIDs(report.Findings, report.Plan, nil)
			scan.Mods = append(scan.Mods, report)
			scan.Errors = append(scan.Errors, DoctorScanError{Filename: name, Error: inspectErr.Error()})
			continue
		}
		signals, signalErr := inspectJarSignals(path)
		if signalErr != nil {
			scan.Errors = append(scan.Errors, DoctorScanError{Filename: name, Error: signalErr.Error()})
		}
		report := analyzeDoctorMod(local, signals, request)
		scan.Mods = append(scan.Mods, report)
	}
	sort.Slice(scan.Mods, func(i, j int) bool {
		if scan.Mods[i].RiskScore != scan.Mods[j].RiskScore {
			return scan.Mods[i].RiskScore > scan.Mods[j].RiskScore
		}
		return strings.ToLower(scan.Mods[i].Local.Filename) < strings.ToLower(scan.Mods[j].Local.Filename)
	})
	scan.Summary = summarizeDoctorReports(scan.Mods)
	applyDoctorGraphAnalysis(&scan)
	extraGlobal := buildGlobalDoctorFindings(scan)
	countDoctorScanFindings(&scan.Summary, extraGlobal)
	scan.GlobalFindings = append(scan.GlobalFindings, extraGlobal...)
	sortDoctorFindings(scan.GlobalFindings)
	scan.MigrationPlan = buildGlobalDoctorPlan(scan)
	scan.RecommendedSourceIDs = collectDoctorSourceIDs(scan.GlobalFindings, scan.MigrationPlan, scan.Mods)
	return scan, nil
}

func inspectJarSignals(path string) (JarSignals, error) {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return JarSignals{}, err
	}
	defer reader.Close()
	signals := JarSignals{EntryCount: len(reader.File), ClassDigests: map[string]string{}}
	loaderAPIs := map[string]bool{}
	manifestAttributes := map[string]bool{}
	var readTotal int64
	for _, file := range reader.File {
		name := filepath.ToSlash(file.Name)
		lower := strings.ToLower(name)
		if strings.HasPrefix(lower, "data/") && !strings.HasSuffix(lower, "/") {
			signals.DataFileCount++
		}
		if strings.HasPrefix(lower, "assets/") && !strings.HasSuffix(lower, "/") {
			signals.AssetFileCount++
		}
		if lower == "pack.mcmeta" {
			signals.HasPackMetadata = true
		}
		if isMixinConfigName(lower) {
			signals.MixinConfigs = appendUniqueString(signals.MixinConfigs, name)
			if b, ok := doctorReadZipFile(file, 2<<20, &readTotal); ok {
				var probe mixinConfigProbe
				if json.Unmarshal(b, &probe) == nil {
					if probe.Plugin != "" {
						signals.MixinPlugins = appendUniqueString(signals.MixinPlugins, probe.Plugin)
					}
					if probe.Refmap != "" {
						signals.MixinRefmaps = appendUniqueString(signals.MixinRefmaps, probe.Refmap)
					}
				}
			}
		}
		if strings.HasSuffix(lower, ".accesswidener") || strings.Contains(lower, "access_widener") {
			signals.AccessWideners = appendUniqueString(signals.AccessWideners, name)
		}
		if lower == "meta-inf/accesstransformer.cfg" || lower == "meta-inf/accesstransformers.cfg" || strings.Contains(lower, "accesstransformer") {
			signals.AccessTransformers = appendUniqueString(signals.AccessTransformers, name)
		}
		if lower == "meta-inf/coremods.json" || strings.Contains(lower, "/coremods/") || strings.HasSuffix(lower, ".coremod.js") {
			signals.Coremods = appendUniqueString(signals.Coremods, name)
		}
		if strings.HasPrefix(lower, "meta-inf/services/") {
			signals.ServiceProviders = appendUniqueString(signals.ServiceProviders, name)
			if strings.Contains(lower, "itransformationservice") || strings.Contains(lower, "imodlocator") || strings.Contains(lower, "imodfilecandidatelocator") || strings.Contains(lower, "dependencylocator") {
				signals.TransformationServices = appendUniqueString(signals.TransformationServices, name)
			}
		}
		if strings.HasSuffix(lower, ".jar") {
			signals.NestedJars = appendUniqueString(signals.NestedJars, name)
		}
		if isNativeLibrary(lower) {
			signals.NativeLibraries = appendUniqueString(signals.NativeLibraries, name)
		}
		if isSignatureFile(lower) {
			signals.SignatureFiles = appendUniqueString(signals.SignatureFiles, name)
		}
		if lower == "meta-inf/manifest.mf" {
			if b, ok := doctorReadZipFile(file, 2<<20, &readTotal); ok {
				attrs := parseManifestAttributes(b)
				for key, value := range attrs {
					attribute := key + ": " + value
					manifestAttributes[attribute] = true
					if strings.EqualFold(key, "Mapping-Namespace") || strings.EqualFold(key, "Fabric-Mapping-Namespace") || strings.EqualFold(key, "Fabric-Loom-Mapping-Namespace") {
						signals.MappingNamespace = strings.TrimSpace(value)
					}
				}
			}
		}
		if lower == "fabric.mod.json" || lower == "quilt.mod.json" {
			if b, ok := doctorReadZipFile(file, 2<<20, &readTotal); ok {
				detectMetadataSignals(lower, b, &signals)
			}
		}
		if lower == "meta-inf/mods.toml" || lower == "meta-inf/neoforge.mods.toml" {
			if b, ok := doctorReadZipFile(file, 4<<20, &readTotal); ok {
				ids, dependencies := doctorParseTOMLMetadata(string(b), filepath.Base(lower))
				signals.ModIDs = append(signals.ModIDs, ids...)
				signals.Dependencies = append(signals.Dependencies, dependencies...)
			}
		}
		if strings.HasSuffix(lower, ".class") {
			signals.ClassCount++
			if readTotal >= doctorTotalReadLimit {
				signals.TruncatedAnalysis = true
				continue
			}
			b, ok := doctorReadZipFile(file, doctorClassReadLimit, &readTotal)
			if !ok {
				continue
			}
			signals.ClassEntries = append(signals.ClassEntries, name)
			digest := sha256.Sum256(b)
			signals.ClassDigests[name] = fmt.Sprintf("%x", digest[:])
			if len(b) >= 8 && bytes.Equal(b[:4], []byte{0xca, 0xfe, 0xba, 0xbe}) {
				major := int(binary.BigEndian.Uint16(b[6:8]))
				if major > signals.MaxClassMajor {
					signals.MaxClassMajor = major
					signals.MaxJava = javaForClassMajor(major)
				}
			}
			detectClassSignals(b, &signals, loaderAPIs)
		}
	}
	for attribute := range manifestAttributes {
		signals.ManifestAttributes = append(signals.ManifestAttributes, attribute)
	}
	for api := range loaderAPIs {
		signals.LoaderAPIs = append(signals.LoaderAPIs, api)
	}
	signals.ModIDs = uniqueNonEmpty(signals.ModIDs)
	signals.Dependencies = uniqueDoctorDependencies(signals.Dependencies)
	sort.Strings(signals.ModIDs)
	sort.Strings(signals.ClassEntries)
	sort.Strings(signals.ManifestAttributes)
	sort.Strings(signals.LoaderAPIs)
	sort.Strings(signals.MixinConfigs)
	sort.Strings(signals.MixinPlugins)
	sort.Strings(signals.MixinRefmaps)
	sort.Strings(signals.AccessWideners)
	sort.Strings(signals.AccessTransformers)
	sort.Strings(signals.Coremods)
	sort.Strings(signals.TransformationServices)
	sort.Strings(signals.ServiceProviders)
	sort.Strings(signals.NestedJars)
	sort.Strings(signals.NativeLibraries)
	sort.Strings(signals.SignatureFiles)
	signals.ReadBytes = readTotal
	return signals, nil
}

func doctorReadZipFile(file *zip.File, limit int64, total *int64) ([]byte, bool) {
	if *total >= doctorTotalReadLimit {
		return nil, false
	}
	remaining := doctorTotalReadLimit - *total
	if limit > remaining {
		limit = remaining
	}
	reader, err := file.Open()
	if err != nil {
		return nil, false
	}
	defer reader.Close()
	b, err := io.ReadAll(io.LimitReader(reader, limit))
	if err != nil {
		return nil, false
	}
	*total += int64(len(b))
	return b, true
}

func detectMetadataSignals(name string, b []byte, signals *JarSignals) {
	if name == "fabric.mod.json" {
		var metadata struct {
			AccessWidener string          `json:"accessWidener"`
			Mixins        json.RawMessage `json:"mixins"`
			Jars          []struct {
				File string `json:"file"`
			} `json:"jars"`
		}
		if json.Unmarshal(b, &metadata) == nil {
			if metadata.AccessWidener != "" {
				signals.AccessWideners = appendUniqueString(signals.AccessWideners, metadata.AccessWidener)
			}
			for _, mixin := range rawStringValues(metadata.Mixins) {
				signals.MixinConfigs = appendUniqueString(signals.MixinConfigs, mixin)
			}
			for _, nested := range metadata.Jars {
				if nested.File != "" {
					signals.NestedJars = appendUniqueString(signals.NestedJars, nested.File)
				}
			}
		}
		ids, dependencies := doctorParseFabricMetadata(b)
		signals.ModIDs = append(signals.ModIDs, ids...)
		signals.Dependencies = append(signals.Dependencies, dependencies...)
		return
	}
	var metadata struct {
		Mixin         json.RawMessage `json:"mixin"`
		AccessWidener string          `json:"access_widener"`
	}
	if json.Unmarshal(b, &metadata) == nil {
		if metadata.AccessWidener != "" {
			signals.AccessWideners = appendUniqueString(signals.AccessWideners, metadata.AccessWidener)
		}
		for _, mixin := range rawStringValues(metadata.Mixin) {
			signals.MixinConfigs = appendUniqueString(signals.MixinConfigs, mixin)
		}
	}
	ids, dependencies := doctorParseQuiltMetadata(b)
	signals.ModIDs = append(signals.ModIDs, ids...)
	signals.Dependencies = append(signals.Dependencies, dependencies...)
}

func rawStringValues(raw json.RawMessage) []string {
	if len(raw) == 0 {
		return nil
	}
	var single string
	if json.Unmarshal(raw, &single) == nil && single != "" {
		return []string{single}
	}
	var list []any
	if json.Unmarshal(raw, &list) != nil {
		return nil
	}
	out := []string{}
	for _, item := range list {
		switch value := item.(type) {
		case string:
			if value != "" {
				out = append(out, value)
			}
		case map[string]any:
			if config, ok := value["config"].(string); ok && config != "" {
				out = append(out, config)
			}
		}
	}
	return out
}

func detectClassSignals(b []byte, signals *JarSignals, loaderAPIs map[string]bool) {
	patterns := []struct {
		needle string
		apply  func()
	}{
		{"java/lang/reflect", func() { signals.UsesReflection = true }},
		{"sun/misc/Unsafe", func() { signals.UsesUnsafe = true }},
		{"jdk/internal/misc/Unsafe", func() { signals.UsesUnsafe = true }},
		{"java/lang/invoke/MethodHandles", func() { signals.UsesMethodHandles = true }},
		{"com/llamalad7/mixinextras", func() { signals.UsesMixinExtras = true }},
		{"kotlin/Metadata", func() { signals.UsesKotlin = true }},
		{"scala/", func() { signals.UsesScala = true }},
		{"net/minecraft/client/", func() { signals.HasClientReferences = true }},
		{"net/minecraft/server/", func() { signals.HasServerReferences = true }},
	}
	for _, pattern := range patterns {
		if bytes.Contains(b, []byte(pattern.needle)) {
			pattern.apply()
		}
	}
	apiPatterns := map[string]string{
		"fabric-loader": "net/fabricmc/loader/",
		"fabric-api":    "net/fabricmc/fabric/api/",
		"quilt-loader":  "org/quiltmc/loader/",
		"quilt-api":     "org/quiltmc/qsl/",
		"forge":         "net/minecraftforge/",
		"neoforge":      "net/neoforged/",
		"architectury":  "dev/architectury/",
		"mixin":         "org/spongepowered/asm/mixin/",
		"mixinextras":   "com/llamalad7/mixinextras/",
		"kotlin":        "kotlin/",
		"scala":         "scala/",
	}
	for id, pattern := range apiPatterns {
		if bytes.Contains(b, []byte(pattern)) {
			loaderAPIs[id] = true
		}
	}
}

func analyzeDoctorMod(local LocalModFile, signals JarSignals, request DoctorScanRequest) DoctorModReport {
	report := DoctorModReport{Local: local, Signals: signals, MigrationDirection: migrationDirection(request.SourceGameVersion, request.TargetGameVersion)}
	score := 0
	add := func(points int, finding DoctorFinding) {
		score += points
		report.Findings = append(report.Findings, finding)
	}
	if strings.TrimSpace(local.Metadata.SourceURL) == "" {
		add(8, DoctorFinding{ID: "source-not-declared", Severity: "warning", Category: "provenance", Title: "Source repository is not declared in the JAR", Evidence: local.Metadata.MetadataBy, Action: "Resolve the exact upstream source and release tag before attempting a source port. If source cannot be found, use a reproducible dual-decompiler recovery path.", SourceIDs: []string{"github-releases", "vineflower", "cfr"}})
	} else {
		report.Findings = append(report.Findings, DoctorFinding{ID: "source-declared", Severity: "info", Category: "provenance", Title: "Source location is declared", Evidence: local.Metadata.SourceURL, Action: "Verify that the repository and tag reproduce this exact JAR before editing.", SourceIDs: []string{"github-releases"}})
	}
	if signals.MaxJava > 0 {
		if signals.MaxJava > targetJavaForMinecraft(request.TargetGameVersion) {
			add(25, DoctorFinding{ID: "java-bytecode-too-new", Severity: "critical", Category: "java", Title: "Class files require a newer Java level than the target runtime", Evidence: fmt.Sprintf("JAR max Java %d; target Minecraft runtime Java %d", signals.MaxJava, targetJavaForMinecraft(request.TargetGameVersion)), Action: "Recompile source with the target toolchain and target bytecode. Metadata-only edits cannot make newer bytecode run on an older JVM.", SourceIDs: []string{"jdk-tools", "gradle-versions-plugin"}})
		} else if signals.MaxJava < targetJavaForMinecraft(request.TargetGameVersion) {
			add(4, DoctorFinding{ID: "java-toolchain-transition", Severity: "info", Category: "java", Title: "Target Minecraft uses a newer Java toolchain", Evidence: fmt.Sprintf("JAR max Java %d; target runtime Java %d", signals.MaxJava, targetJavaForMinecraft(request.TargetGameVersion)), Action: "Move the Gradle toolchain and CI matrix to the target Java version, then compile and run tests under that JVM.", SourceIDs: []string{"jdk-tools", "gradle-versions-plugin"}})
		}
	}
	if len(signals.MixinConfigs) > 0 {
		points := 6 + minInt(18, len(signals.MixinConfigs)*2)
		add(points, DoctorFinding{ID: "mixin-retarget-required", Severity: severityForPoints(points), Category: "bytecode", Title: "Mixin injection points must be revalidated", Evidence: fmt.Sprintf("%d mixin config(s), %d plugin(s), %d refmap(s)", len(signals.MixinConfigs), len(signals.MixinPlugins), len(signals.MixinRefmaps)), Action: "Resolve every target class, method descriptor, injection point, slice, ordinal, accessor, invoker, redirect, and overwrite against the target bytecode. Never blanket-disable failing mixins.", SourceIDs: []string{"mixin-wiki", "mixin-source", "mixintrace", "enigma", "vineflower"}})
	}
	if len(signals.MixinPlugins) > 0 {
		add(8, DoctorFinding{ID: "mixin-plugin", Severity: "warning", Category: "bytecode", Title: "Custom mixin plugin controls conditional transformation", Evidence: strings.Join(signals.MixinPlugins, ", "), Action: "Port plugin predicates and target discovery before evaluating individual mixins.", SourceIDs: []string{"mixin-source", "mixintrace"}})
	}
	if len(signals.AccessWideners) > 0 || len(signals.AccessTransformers) > 0 {
		add(8, DoctorFinding{ID: "access-rules", Severity: "warning", Category: "access", Title: "Access rules depend on mappings and target members", Evidence: fmt.Sprintf("%d access widener(s), %d access transformer(s)", len(signals.AccessWideners), len(signals.AccessTransformers)), Action: "Remap and validate every class/member descriptor against the target mappings. Remove entries only after proving the target is publicly accessible or unused.", SourceIDs: []string{"access-widener", "mapping-io", "tiny-remapper", "specialsource"}})
	}
	if len(signals.Coremods) > 0 || len(signals.TransformationServices) > 0 {
		add(32, DoctorFinding{ID: "early-transformer", Severity: "critical", Category: "bytecode", Title: "Early-launch transformation code requires a source-level port", Evidence: fmt.Sprintf("%d coremod artifact(s), %d transformation service(s)", len(signals.Coremods), len(signals.TransformationServices)), Action: "Port the transformation contract against the target loader bootstrap and verify class ordering in a real launch. A compatibility metadata edit is insufficient.", SourceIDs: []string{"asm", "minecraftforge", "neoforge", "mixin-source"}})
	}
	if len(signals.SignatureFiles) > 0 {
		add(18, DoctorFinding{ID: "signed-jar", Severity: "high", Category: "integrity", Title: "JAR contains signing metadata", Evidence: strings.Join(signals.SignatureFiles, ", "), Action: "Treat any mutation as a new artifact, remove stale signature blocks only in the rebuilt output, and publish fresh hashes and provenance. Preserve the untouched signed original.", SourceIDs: []string{"jdk-tools"}})
	}
	if len(signals.NativeLibraries) > 0 {
		add(25, DoctorFinding{ID: "native-code", Severity: "critical", Category: "native", Title: "Native libraries require platform and ABI validation", Evidence: strings.Join(signals.NativeLibraries, ", "), Action: "Verify every supported OS/architecture, JNI signature, bundled runtime, and extraction path. Rebuild native components when the target JVM or LWJGL ABI changes.", SourceIDs: []string{"jdk-tools", "headlessmc"}})
	}
	if len(signals.NestedJars) > 0 {
		add(8, DoctorFinding{ID: "nested-jars", Severity: "warning", Category: "packaging", Title: "Nested dependencies are bundled", Evidence: fmt.Sprintf("%d nested JAR(s)", len(signals.NestedJars)), Action: "Resolve each nested dependency independently, preserve loader-specific jar-in-jar metadata, and reject duplicate or incompatible embedded libraries.", SourceIDs: []string{"jarsplitter", "fabric-loader", "neoforge"}})
	}
	if signals.UsesUnsafe {
		add(12, DoctorFinding{ID: "unsafe-api", Severity: "high", Category: "java", Title: "Unsafe or internal JVM APIs are referenced", Evidence: "sun.misc.Unsafe or jdk.internal.misc.Unsafe symbol detected", Action: "Audit module-access flags and replace internal API use where possible before moving Java runtimes.", SourceIDs: []string{"jdk-tools", "jdk-tools"}})
	} else if signals.UsesReflection || signals.UsesMethodHandles {
		add(6, DoctorFinding{ID: "reflection-api", Severity: "warning", Category: "java", Title: "Reflection or method handles may bind to changed members", Evidence: fmt.Sprintf("reflection=%t, methodHandles=%t", signals.UsesReflection, signals.UsesMethodHandles), Action: "Record each reflective owner/name/descriptor and verify it against target bytecode and module-access behavior.", SourceIDs: []string{"jdk-tools", "classgraph", "vineflower"}})
	}
	if signals.UsesKotlin || signals.UsesScala {
		add(5, DoctorFinding{ID: "language-runtime", Severity: "warning", Category: "toolchain", Title: "Additional JVM language runtime detected", Evidence: fmt.Sprintf("Kotlin=%t, Scala=%t", signals.UsesKotlin, signals.UsesScala), Action: "Update the language plugin and bundled runtime together with the loader and Java toolchain; detect duplicate stdlib copies.", SourceIDs: []string{"fabric-language-kotlin", "gradle-versions-plugin"}})
	}
	if signals.TruncatedAnalysis {
		add(5, DoctorFinding{ID: "analysis-cap", Severity: "warning", Category: "analysis", Title: "Static symbol scan reached its bounded read budget", Evidence: fmt.Sprintf("Read %d bytes", signals.ReadBytes), Action: "Run the external deep-analysis pipeline before patching this unusually large artifact.", SourceIDs: []string{"classgraph", "semgrep", "codeql"}})
	}
	if crossesDeobfuscationBoundary(request.SourceGameVersion, request.TargetGameVersion) {
		add(25, DoctorFinding{ID: "deobfuscation-boundary", Severity: "critical", Category: "mappings", Title: "Migration crosses Minecraft's 2026 deobfuscation boundary", Evidence: request.SourceGameVersion + " -> " + request.TargetGameVersion, Action: "Migrate through adjacent loader primers, replace assumptions about obfuscated/intermediary namespaces, regenerate mappings-dependent access and mixin metadata, and compile against the official deobfuscated target.", SourceIDs: []string{"minecraft-release-notes", "neoforge-primers", "quilt-site", "mapping-io", "tiny-remapper"}})
	}
	if distance := doctorVersionRisk(request.SourceGameVersion, request.TargetGameVersion); distance > 0 {
		add(distance, DoctorFinding{ID: "minecraft-version-distance", Severity: severityForPoints(distance), Category: "minecraft", Title: "Minecraft API and data migration is required", Evidence: request.SourceGameVersion + " -> " + request.TargetGameVersion, Action: "Port one adjacent release boundary at a time, compile and launch at each checkpoint, and preserve a reversible commit for every successful hop.", SourceIDs: []string{"minecraft-release-notes", "neoforge-primers", "fabric-docs", "forge-docs", "misode-changelog"}})
	}
	if loaderRisk := doctorLoaderRisk(request.SourceLoader, request.TargetLoader); loaderRisk > 0 {
		add(loaderRisk, DoctorFinding{ID: "loader-migration", Severity: severityForPoints(loaderRisk), Category: "loader", Title: "Loader APIs and metadata must be migrated", Evidence: normalizeDoctorLoader(request.SourceLoader) + " -> " + normalizeDoctorLoader(request.TargetLoader), Action: loaderMigrationAction(request.SourceLoader, request.TargetLoader), SourceIDs: loaderMigrationSources(request.SourceLoader, request.TargetLoader)})
	}
	if !doctorArtifactSupportsLoader(local.Metadata.Loaders, request.TargetLoader) {
		add(20, DoctorFinding{ID: "loader-mismatch", Severity: "critical", Category: "loader", Title: "Installed artifact targets a different loader", Evidence: strings.Join(local.Metadata.Loaders, ", ") + " artifact; target is " + normalizeDoctorLoader(request.TargetLoader), Action: "Resolve a native target-loader build or perform a source migration. Test runtime bridges only in an isolated compatibility profile.", SourceIDs: loaderMigrationSources(firstNonEmpty(strings.Join(local.Metadata.Loaders, ","), request.SourceLoader), request.TargetLoader)})
	}
	if match, known := doctorMinecraftConstraintLikelyMatches(local.Metadata.Minecraft, request.TargetGameVersion); known && !match {
		add(18, DoctorFinding{ID: "minecraft-constraint-mismatch", Severity: "high", Category: "minecraft", Title: "Embedded Minecraft constraint excludes the target release", Evidence: firstNonEmpty(local.Metadata.Minecraft, "unspecified") + " does not match " + request.TargetGameVersion, Action: "Port and rebuild against the target Minecraft APIs and data formats. Do not broaden metadata until compile and runtime verification pass.", SourceIDs: []string{"minecraft-release-notes", "neoforge-primers", "fabric-docs", "forge-docs"}})
	}
	if signals.HasClientReferences && signals.HasServerReferences {
		add(3, DoctorFinding{ID: "side-boundaries", Severity: "info", Category: "sides", Title: "Both client and server symbols are present", Evidence: "net.minecraft.client and net.minecraft.server references detected", Action: "Revalidate physical/logical side guards and run both dedicated-server and client tests.", SourceIDs: []string{"neoforge-gametest", "fabric-testing", "headlessmc"}})
	}
	if signals.DataFileCount > 0 || signals.AssetFileCount > 0 || signals.HasPackMetadata {
		add(3, DoctorFinding{ID: "resource-migration", Severity: "info", Category: "resources", Title: "Data or asset resources need schema validation", Evidence: fmt.Sprintf("data=%d, assets=%d, pack.mcmeta=%t", signals.DataFileCount, signals.AssetFileCount, signals.HasPackMetadata), Action: "Validate pack formats, registries, tags, recipes, loot, worldgen, codecs, shaders, models, and localization against the target release.", SourceIDs: []string{"mcmeta", "minecraft-data", "misode-changelog", "datafixerupper"}})
	}
	if score > 100 {
		score = 100
	}
	report.RiskScore = score
	report.RiskLevel = doctorRiskLevel(score)
	report.Porting = decideDoctorPorting(local, signals, request)
	report.Plan = buildDoctorModPlan(local, signals, request, report)
	report.RecommendedSourceIDs = collectDoctorSourceIDs(report.Findings, report.Plan, nil)
	sortDoctorFindings(report.Findings)
	return report
}

func buildDoctorModPlan(local LocalModFile, signals JarSignals, request DoctorScanRequest, report DoctorModReport) []DoctorStep {
	steps := []DoctorStep{}
	add := func(id, title, action, why, mode string, sources ...string) {
		steps = append(steps, DoctorStep{Order: len(steps) + 1, ID: id, Title: title, Action: action, Why: why, Mode: mode, SourceIDs: uniqueNonEmpty(sources)})
	}
	add("preserve-original", "Preserve and fingerprint the exact original", "Copy the JAR to an immutable job workspace, retain its enabled/disabled state, record SHA-1/SHA-256/SHA-512, archive entry list, metadata, signatures, and dependency graph.", "Every later edit must be reversible and attributable to this exact input.", "automatic")
	sourceRecoverySources := doctorSourceRecoverySources(request.SourceGameVersion, request.SourceLoader)
	if strings.TrimSpace(local.Metadata.SourceURL) != "" {
		sources := append(append([]string{}, sourceRecoverySources...), "github-releases", "gradle-versions-plugin")
		add("reproduce-source", "Reproduce the upstream source build", "Resolve the declared repository, exact release/tag/commit, mappings, loader, Gradle wrapper, and dependency lock state. Build until the output's metadata and behavior match the installed JAR.", "A source port is safest when the starting build is reproducible.", "guided", sources...)
	} else {
		sources := append(append([]string{}, sourceRecoverySources...), "vineflower", "cfr", "mapping-io", "tiny-remapper", "enigma")
		add("recover-source", "Recover a reviewable source baseline", "Search provider metadata and release provenance first. When source is unavailable, reconstruct with the version-appropriate game/toolchain workspace plus independent Vineflower and CFR output, compare control flow, remap names, and preserve a machine-readable uncertainty ledger.", "A single decompiler or a modern-only toolchain can silently reconstruct incorrect legacy source; independent outputs and era-specific mappings expose ambiguity.", "guided", sources...)
	}
	if isLegacyMinecraftVersion(request.SourceGameVersion) || isLegacyMinecraftVersion(request.TargetGameVersion) {
		sources := append(doctorLegacyRuntimeSources(request.SourceGameVersion, request.SourceLoader), doctorLegacyRuntimeSources(request.TargetGameVersion, request.TargetLoader)...)
		sources = append(sources, doctorBuildSources(request.SourceLoader, request.SourceGameVersion)...)
		sources = append(sources, doctorBuildSources(request.TargetLoader, request.TargetGameVersion)...)
		add("legacy-toolchain-track", "Separate legacy source reconstruction from runtime compatibility", fmt.Sprintf("Treat %s/%s -> %s/%s as three independent tracks: reproduce or reconstruct source with the correct historical mappings and build system; port the code and data contracts; then test optional modern-Java or loader compatibility layers in a copied profile. Never report a runtime bridge as a completed source port.", request.SourceGameVersion, normalizeDoctorLoader(request.SourceLoader), request.TargetGameVersion, normalizeDoctorLoader(request.TargetLoader)), "Minecraft 1.7.10, 1.12.2, beta/classic, Legacy Fabric, and Ornithe ecosystems use materially different launchers, mappings, transformers, Java assumptions, and build pipelines.", "guided", sources...)
	}
	targetSources := append(doctorBuildSources(request.TargetLoader, request.TargetGameVersion), "jdk-tools", "gradle-versions-plugin")
	add("target-toolchain", "Create the exact target toolchain", fmt.Sprintf("Create a clean target branch using Minecraft %s, %s, Java %d, era-correct loader build tooling, and exact mappings/dependencies. Do not edit only the JAR metadata.", request.TargetGameVersion, normalizeDoctorLoader(request.TargetLoader), targetJavaForMinecraft(request.TargetGameVersion)), "The compiler and loader must expose every API break before runtime.", "automatic", targetSources...)
	if request.SourceGameVersion != request.TargetGameVersion {
		sources := []string{"minecraft-release-notes", "neoforge-primers", "fabric-docs", "forge-docs", "misode-changelog", "ornithe-gitcraft"}
		add("adjacent-version-chain", "Port through adjacent Minecraft boundaries", "Generate an ordered migration chain from the source release to the target. Apply one release primer/change set at a time, compile, run static checks, and checkpoint before the next hop. For legacy ranges, generate private mapped/decompiled history with the correct manifest and mapping source rather than extrapolating modern changelogs backwards.", "Large jumps hide which release removed or changed a symbol, registry, codec, network contract, launcher behavior, or data format.", "guided", sources...)
	}
	if normalizeDoctorLoader(request.SourceLoader) != normalizeDoctorLoader(request.TargetLoader) {
		sources := append(loaderMigrationSources(request.SourceLoader, request.TargetLoader), doctorBuildSources(request.SourceLoader, request.SourceGameVersion)...)
		sources = append(sources, doctorBuildSources(request.TargetLoader, request.TargetGameVersion)...)
		add("loader-port", "Port loader contracts", loaderMigrationAction(request.SourceLoader, request.TargetLoader), "Loader metadata, lifecycle, registries, networking, config, capabilities/components, rendering hooks, and distribution rules are not interchangeable.", "guided", sources...)
	}
	if crossesDeobfuscationBoundary(request.SourceGameVersion, request.TargetGameVersion) || signals.MappingNamespace != "" || len(signals.AccessWideners)+len(signals.AccessTransformers) > 0 {
		sources := []string{"mapping-io", "tiny-remapper", "specialsource", "enigma"}
		if isLegacyMinecraftVersion(request.SourceGameVersion) || isLegacyMinecraftVersion(request.TargetGameVersion) {
			sources = append(sources, "ornithe-feather", "retromcp-java", "legacy-fabric-looming")
		}
		add("mapping-migration", "Rebuild mapping-dependent code and metadata", "Remap source and bytecode into the target namespace, regenerate access wideners/transformers and refmaps, then diff owners, member names, and descriptors against target classes. Carry mapping confidence and provenance across every namespace edge.", "Mappings are part of the executable contract for access rules, reflection, mixins, and binary patches.", "automatic", sources...)
	}
	add("automated-source-rewrites", "Apply provable source transformations", "Use compiler diagnostics plus OpenRewrite, JavaParser, Spoon, or focused AST transforms for mechanical API moves. Require a rule test and before/after diff for every transformation family.", "AST-aware changes scale across a codebase without unsafe global text replacement.", "automatic", "openrewrite", "javaparser", "spoon", "gumtree")
	add("api-compatibility-loop", "Drive the API compatibility loop to zero errors", "Compile, classify each missing class/method/field/descriptor, inspect both source and target APIs, apply the smallest semantic replacement, and record the mapping in the conversion knowledge base.", "The compiler and binary API diff are higher-confidence evidence than permissive dependency ranges.", "guided", "japicmp", "revapi", "jdk-tools", "classgraph")
	if len(signals.MixinConfigs)+len(signals.Coremods)+len(signals.TransformationServices) > 0 {
		add("transformer-verification", "Retarget bytecode transformations", "Decompile and inspect exact target bytecode, update descriptors/injection sites, validate plugin predicates and ordering, then launch with mixin diagnostics. Keep every transformation's intended behavior testable.", "Transformers can compile while failing at bootstrap or silently altering the wrong instruction.", "guided", "mixin-source", "mixin-wiki", "mixintrace", "enigma", "vineflower", "asm")
	}
	if signals.DataFileCount > 0 || signals.AssetFileCount > 0 || signals.HasPackMetadata {
		sources := []string{"mcmeta", "minecraft-data", "misode-changelog", "datafixerupper", "amulet", "mcaselector"}
		if isLegacyMinecraftVersion(request.SourceGameVersion) || isLegacyMinecraftVersion(request.TargetGameVersion) {
			sources = append(sources, "papermc-dataconverter")
		}
		add("resource-data-migration", "Migrate resources and persistent data", "Update pack formats and target schemas for registries, tags, recipes, loot, worldgen, codecs, models, shaders, language files, network payloads, saved data, and DataFixer rules. Validate sample worlds on copies. PaperMC DataConverter is architecture/reference evidence only for modded worlds because it does not run other mods' data fixers.", "A clean compile does not prove data packs, resources, or old saves remain valid.", "guided", sources...)
	}
	runtimeSources := []string{"neoforge-gametest", "fabric-testing", "headlessmc", "junit", "mclogs"}
	runtimeSources = append(runtimeSources, doctorLegacyRuntimeSources(request.TargetGameVersion, request.TargetLoader)...)
	add("real-runtime-matrix", "Run the real runtime matrix", "Launch the exact rebuilt artifact on client and dedicated server where applicable, run GameTests/unit tests, exercise original features, inspect fresh logs, restart, and verify persistence and networking. Run runtime bridges and legacy compatibility layers as separate matrix variants, never as hidden dependencies of the source-port result.", "Static analysis and compilation cannot prove loader bootstrap, mixins, rendering, world data, native libraries, or protocol behavior.", "automatic", runtimeSources...)
	add("package-and-prove", "Package, diff, and publish provenance", "Create fresh native-loader JARs, test ZIP integrity, compare entries/resources against the source artifact, scan dependencies, generate checksums/SBOM, install into fresh instances, and retain the complete build receipt. Shade only plain Java libraries; retain native outputs before optionally merging multi-loader artifacts.", "The final artifact must be reproducible, attributable, and proven to load instead of merely compiling.", "automatic", "mc-publish", "mod-publish-plugin", "minotaur", "cursegradle", "packwiz", "forgix", "modshade")
	return steps
}

func buildGlobalDoctorFindings(scan DoctorScan) []DoctorFinding {
	findings := []DoctorFinding{}
	if crossesDeobfuscationBoundary(scan.SourceGameVersion, scan.TargetGameVersion) {
		findings = append(findings, DoctorFinding{ID: "global-2026-boundary", Severity: "critical", Category: "mappings", Title: "The whole migration crosses the 2026 deobfuscation and Java 25 transition", Evidence: scan.SourceGameVersion + " -> " + scan.TargetGameVersion, Action: "Create an adjacent-version conversion program, not a metadata bump. Rebuild every mod with mapping-sensitive code, bytecode transforms, reflection, access rules, or bundled runtimes.", SourceIDs: []string{"minecraft-release-notes", "minecraft-version-manifest", "neoforge-primers", "quilt-site"}})
	}
	if scan.SourceLoader != scan.TargetLoader {
		sources := append(loaderMigrationSources(scan.SourceLoader, scan.TargetLoader), doctorBuildSources(scan.SourceLoader, scan.SourceGameVersion)...)
		sources = append(sources, doctorBuildSources(scan.TargetLoader, scan.TargetGameVersion)...)
		findings = append(findings, DoctorFinding{ID: "global-loader-change", Severity: "high", Category: "loader", Title: "The instance changes loader ecosystems", Evidence: scan.SourceLoader + " -> " + scan.TargetLoader, Action: "Resolve each project to a native target build first. Port source where no maintained build exists, and use runtime bridges only as separately tested compatibility options.", SourceIDs: uniqueNonEmpty(sources)})
	}
	if isLegacyMinecraftVersion(scan.SourceGameVersion) || isLegacyMinecraftVersion(scan.TargetGameVersion) {
		sources := append(doctorSourceRecoverySources(scan.SourceGameVersion, scan.SourceLoader), doctorBuildSources(scan.TargetLoader, scan.TargetGameVersion)...)
		sources = append(sources, doctorLegacyRuntimeSources(scan.SourceGameVersion, scan.SourceLoader)...)
		sources = append(sources, doctorLegacyRuntimeSources(scan.TargetGameVersion, scan.TargetLoader)...)
		findings = append(findings, DoctorFinding{ID: "global-legacy-toolchain", Severity: "high", Category: "legacy-toolchain", Title: "The migration includes a legacy Minecraft toolchain boundary", Evidence: scan.SourceGameVersion + "/" + scan.SourceLoader + " -> " + scan.TargetGameVersion + "/" + scan.TargetLoader, Action: "Use era-specific source reconstruction, mappings, build plugins, Java/runtime bootstraps, and compatibility patches. Keep runtime modernization, source porting, and data migration as separate verified deliverables.", SourceIDs: uniqueNonEmpty(sources)})
	}
	if scan.Summary.CriticalRisk > 0 {
		findings = append(findings, DoctorFinding{ID: "critical-mod-gate", Severity: "critical", Category: "workflow", Title: "Critical-risk mods block unattended conversion", Evidence: fmt.Sprintf("%d critical-risk mod(s)", scan.Summary.CriticalRisk), Action: "Process critical artifacts individually with source recovery, exact API evidence, transformation audits, and real runtime tests before enabling a batch migration."})
	}
	if scan.Summary.BinaryOnly > 0 {
		findings = append(findings, DoctorFinding{ID: "binary-only-corpus", Severity: "warning", Category: "provenance", Title: "Some installed mods do not declare source provenance", Evidence: fmt.Sprintf("%d mod(s) need source resolution or controlled decompilation", scan.Summary.BinaryOnly), Action: "Resolve repository lineage through provider metadata and release records. Keep decompiled-source uncertainty explicit.", SourceIDs: []string{"modrinth-api", "curseforge-api", "github-releases", "vineflower", "cfr"}})
	}
	if len(scan.Mods) == 0 && len(scan.Errors) == 0 {
		findings = append(findings, DoctorFinding{ID: "empty-instance", Severity: "info", Category: "instance", Title: "No installed mod JARs were found", Evidence: scan.ModsDirectory, Action: "Point the Vault at the intended Java instance or import the mod set before generating a compatibility plan."})
	}
	sortDoctorFindings(findings)
	return findings
}

func buildGlobalDoctorPlan(scan DoctorScan) []DoctorStep {
	steps := []DoctorStep{}
	add := func(id, title, action, why, mode string, sources ...string) {
		steps = append(steps, DoctorStep{Order: len(steps) + 1, ID: id, Title: title, Action: action, Why: why, Mode: mode, SourceIDs: uniqueNonEmpty(sources)})
	}
	add("snapshot-instance", "Snapshot the complete instance", "Record the launcher/profile, Minecraft, loader, Java, full mod inventory, configs, scripts, datapacks, resource packs, world copies, hashes, and known runtime baseline before changing anything.", "Compatibility work must be recoverable and measured against a known-good baseline.", "automatic", "prism-launcher", "packwiz")
	add("resolve-identity", "Resolve every mod to canonical project and source identity", "Use cryptographic provider matches, embedded project URLs, release metadata, repository tags, and dependency declarations. Keep forks, ports, and substitutions explicit rather than silently merging identities.", "A correct converter cannot patch the wrong project lineage.", "automatic", "modrinth-api", "curseforge-api", "github-releases")
	add("dependency-graph", "Build the complete dependency and incompatibility graph", "Combine embedded metadata, provider requirements, optional dependencies, runtime-discovered class links, mixin targets, services, nested JARs, configuration contracts, and known repair records.", "Version migration is a graph problem; resolving one JAR in isolation can break its dependents.", "automatic", "classgraph", "jdk-tools", "repair-brain")
	if isLegacyMinecraftVersion(scan.SourceGameVersion) || isLegacyMinecraftVersion(scan.TargetGameVersion) {
		sources := append(doctorSourceRecoverySources(scan.SourceGameVersion, scan.SourceLoader), doctorBuildSources(scan.TargetLoader, scan.TargetGameVersion)...)
		sources = append(sources, doctorLegacyRuntimeSources(scan.SourceGameVersion, scan.SourceLoader)...)
		sources = append(sources, doctorLegacyRuntimeSources(scan.TargetGameVersion, scan.TargetLoader)...)
		add("legacy-reconstruction-program", "Run a dedicated legacy reconstruction and compatibility program", "Reproduce the original historical workspace first, then port through explicit source/mapping/data steps. Test Cleanroom, Fugue, MixinBooter, UniMixins, LWJGL3ify, LegacyFix, or other runtime compatibility layers only in separate copied profiles with exact mod/version evidence.", "Old mappings, launchers, coremods, Java assumptions, natives, and APIs cannot be safely inferred from modern Loom or NeoForge workflows.", "guided", sources...)
	}
	add("ordered-port-program", "Execute the ordered conversion program", "Port libraries and APIs before dependents, then performance/core transformers, content mods, integrations, client-only visuals, and final packs. Preserve a verified artifact at every layer.", "Topological ordering reduces repeated failures and makes root causes attributable.", "guided", "neoforge-primers", "fabric-docs", "forge-docs", "architectury-plugin")
	add("compatibility-sandbox", "Validate bridge-based compatibility separately", "Where relevant, test maintained compatibility layers such as Sinytra Connector, Forgified Fabric API, Kilt, or loader-specific bridges in an isolated profile. Never treat one successful launch as proof of full behavior.", "A bridge can save a source port, but it also adds transformation and API-emulation failure modes.", "guided", "sinytra-connector", "forgified-fabric-api", "connector-extras", "kilt")
	add("runtime-regression-matrix", "Run clean client, server, world, and feature regression matrices", "Use fresh instances, exact rebuilt artifacts, startup and gameplay reproducers, GameTests, log classification, restart persistence, protocol checks, and before/after performance evidence.", "The migration is complete only when the real user workflows work without replacement failures.", "automatic", "headlessmc", "neoforge-gametest", "fabric-testing", "mclogs", "spark")
	add("transactional-deployment", "Deploy transactionally with rollback", "Stage every artifact, verify hashes and embedded metadata, back up the existing mod set, apply in dependency order, launch, and roll back the entire batch if validation fails.", "Batch updates must never leave the instance in a half-migrated state.", "automatic", "packwiz", "ferium", "mrpack-updater")
	return steps
}

func summarizeDoctorReports(reports []DoctorModReport) DoctorScanSummary {
	summary := DoctorScanSummary{TotalMods: len(reports)}
	for _, report := range reports {
		switch report.RiskLevel {
		case "low":
			summary.LowRisk++
		case "moderate":
			summary.ModerateRisk++
		case "high":
			summary.HighRisk++
		case "critical":
			summary.CriticalRisk++
		}
		if report.RiskScore > summary.MaxRiskScore {
			summary.MaxRiskScore = report.RiskScore
		}
		if strings.TrimSpace(report.Local.Metadata.SourceURL) != "" {
			summary.SourceAvailable++
		} else {
			summary.BinaryOnly++
		}
		if len(report.Signals.MixinConfigs) > 0 {
			summary.MixinMods++
		}
		if len(report.Signals.Coremods)+len(report.Signals.TransformationServices) > 0 {
			summary.Coremods++
		}
		if len(report.Signals.SignatureFiles) > 0 {
			summary.SignedMods++
		}
		if len(report.Signals.NativeLibraries) > 0 {
			summary.NativeMods++
		}
	}
	return summary
}

func parseGameVersionKey(version string) gameVersionKey {
	version = strings.TrimSpace(strings.TrimPrefix(strings.ToLower(version), "minecraft "))
	if i := strings.IndexAny(version, "-+ "); i >= 0 {
		version = version[:i]
	}
	parts := strings.Split(version, ".")
	key := gameVersionKey{}
	for i := 0; i < len(parts) && i < len(key.Parts); i++ {
		digits := ""
		for _, r := range parts[i] {
			if r < '0' || r > '9' {
				break
			}
			digits += string(r)
		}
		if digits == "" {
			return gameVersionKey{}
		}
		value, err := strconv.Atoi(digits)
		if err != nil {
			return gameVersionKey{}
		}
		key.Parts[i] = value
	}
	if key.Parts[0] == 0 {
		return gameVersionKey{}
	}
	if key.Parts[0] >= 20 {
		key.Era = 2
	} else {
		key.Era = 1
	}
	key.Valid = true
	return key
}

func compareGameVersions(a, b string) int {
	ak, bk := parseGameVersionKey(a), parseGameVersionKey(b)
	if !ak.Valid || !bk.Valid {
		return strings.Compare(a, b)
	}
	if ak.Era != bk.Era {
		if ak.Era < bk.Era {
			return -1
		}
		return 1
	}
	for i := range ak.Parts {
		if ak.Parts[i] < bk.Parts[i] {
			return -1
		}
		if ak.Parts[i] > bk.Parts[i] {
			return 1
		}
	}
	return 0
}

func targetJavaForMinecraft(version string) int {
	key := parseGameVersionKey(version)
	if !key.Valid {
		return 17
	}
	if key.Era == 2 {
		return 25
	}
	minor := key.Parts[1]
	patch := key.Parts[2]
	switch {
	case minor <= 16:
		return 8
	case minor == 17:
		return 16
	case minor < 20:
		return 17
	case minor == 20 && patch <= 4:
		return 17
	default:
		return 21
	}
}

func crossesDeobfuscationBoundary(source, target string) bool {
	s := parseGameVersionKey(source)
	t := parseGameVersionKey(target)
	return s.Valid && t.Valid && s.Era != t.Era
}

func migrationDirection(source, target string) string {
	switch compareGameVersions(source, target) {
	case -1:
		return "upgrade"
	case 1:
		return "downgrade"
	default:
		return "same-version"
	}
}

func doctorVersionRisk(source, target string) int {
	s, t := parseGameVersionKey(source), parseGameVersionKey(target)
	if !s.Valid || !t.Valid || compareGameVersions(source, target) == 0 {
		return 0
	}
	if s.Era != t.Era {
		return 24
	}
	if s.Parts[0] != t.Parts[0] {
		return 24
	}
	if s.Parts[1] != t.Parts[1] {
		delta := s.Parts[1] - t.Parts[1]
		if delta < 0 {
			delta = -delta
		}
		return minInt(22, 10+delta*2)
	}
	if s.Parts[2] != t.Parts[2] {
		return 6
	}
	return 3
}

func normalizeDoctorLoader(loader string) string {
	loader = strings.ToLower(strings.TrimSpace(loader))
	switch loader {
	case "neo forge", "neo-forge", "neo_forge":
		return "neoforge"
	case "minecraftforge", "fml":
		return "forge"
	case "fabric-loader":
		return "fabric"
	case "quilt-loader":
		return "quilt"
	case "legacy fabric", "legacy_fabric", "legacy-fabric-loader":
		return "legacy-fabric"
	case "babric-loader", "babric fabric":
		return "babric"
	case "ornithe-loader":
		return "ornithe"
	}
	return loader
}

func doctorLoaderRisk(source, target string) int {
	s, t := normalizeDoctorLoader(source), normalizeDoctorLoader(target)
	if s == t {
		return 0
	}
	if (s == "forge" && t == "neoforge") || (s == "neoforge" && t == "forge") {
		return 14
	}
	if (s == "fabric" && t == "quilt") || (s == "quilt" && t == "fabric") {
		return 12
	}
	return 25
}

func loaderMigrationAction(source, target string) string {
	s, t := normalizeDoctorLoader(source), normalizeDoctorLoader(target)
	switch {
	case (s == "forge" && t == "neoforge") || (s == "neoforge" && t == "forge"):
		return "Migrate mod metadata, event bus/lifecycle hooks, registries, networking, config, capabilities/attachments, access transformers, Gradle plugin, and loader-specific imports. Compile against both APIs to expose semantic differences."
	case (s == "fabric" && t == "quilt") || (s == "quilt" && t == "fabric"):
		return "Migrate loader metadata, entrypoints, mappings, access wideners, loader APIs, and retired QSL/QFAPI usage. Prefer currently maintained Fabric APIs where Quilt no longer ships an equivalent library."
	default:
		return "Separate common game logic from loader adapters, create a native target module, migrate lifecycle/registries/networking/config/rendering/data generation, and use bridges only as independently verified fallback paths."
	}
}

func loaderMigrationSources(source, target string) []string {
	s, t := normalizeDoctorLoader(source), normalizeDoctorLoader(target)
	ids := []string{loaderBuildSource(s), loaderBuildSource(t), "architectury-plugin", "multiloader-template"}
	if (s == "fabric" && (t == "forge" || t == "neoforge")) || (t == "fabric" && (s == "forge" || s == "neoforge")) {
		ids = append(ids, "sinytra-connector", "forgified-fabric-api", "connector-extras", "kilt")
	}
	if (s == "forge" && t == "neoforge") || (s == "neoforge" && t == "forge") {
		ids = append(ids, "reforged")
	}
	return uniqueNonEmpty(ids)
}

func loaderBuildSource(loader string) string {
	switch normalizeDoctorLoader(loader) {
	case "fabric":
		return "fabric-loom"
	case "quilt":
		return "quilt-loom"
	case "forge":
		return "forgegradle"
	case "neoforge":
		return "moddevgradle"
	case "legacy-fabric":
		return "legacy-fabric-looming"
	case "ornithe", "babric":
		return "ornithe-ploceus"
	default:
		return "minecraft-release-notes"
	}
}

func isLegacyMinecraftVersion(version string) bool {
	normalized := strings.ToLower(strings.TrimSpace(version))
	normalized = strings.TrimPrefix(normalized, "minecraft ")
	for _, prefix := range []string{"rd-", "c0.", "classic", "indev", "infdev", "alpha", "a1.", "beta", "b1."} {
		if strings.HasPrefix(normalized, prefix) {
			return true
		}
	}
	key := parseGameVersionKey(version)
	return key.Valid && key.Era == 1 && compareGameVersions(version, "1.13.2") <= 0
}

func gameVersionAtMost(version, boundary string) bool {
	versionKey, boundaryKey := parseGameVersionKey(version), parseGameVersionKey(boundary)
	return versionKey.Valid && boundaryKey.Valid && compareGameVersions(version, boundary) <= 0
}

func doctorBuildSources(loader, version string) []string {
	normalized := normalizeDoctorLoader(loader)
	legacy := isLegacyMinecraftVersion(version)
	switch normalized {
	case "forge":
		if legacy && gameVersionAtMost(version, "1.7.10") {
			return []string{"retrofuturagradle", "fpgradle", "forgegradle", "retromcp-java"}
		}
		if legacy {
			return []string{"forgegradle", "cleanroom", "mixinbooter", "fugue"}
		}
		return []string{"forgegradle", "mcpconfig"}
	case "neoforge":
		return []string{"moddevgradle", "neoforge-mod-generator", "neoform-runtime", "neoform"}
	case "fabric":
		if legacy {
			return []string{"legacy-fabric-looming", "legacy-fabric-api", "ornithe-ploceus", "ornithe-feather", "fabric-meta"}
		}
		return []string{"fabric-loom", "fabric-meta"}
	case "legacy-fabric":
		return []string{"legacy-fabric-looming", "legacy-fabric-api", "ornithe-feather"}
	case "ornithe":
		return []string{"ornithe-ploceus", "ornithe-feather", "ornithe-gitcraft"}
	case "babric", "stationapi":
		return []string{"stationapi", "ornithe-ploceus", "ornithe-feather"}
	case "quilt":
		if legacy {
			return []string{"ornithe-ploceus", "ornithe-feather", "quilt-loom"}
		}
		return []string{"quilt-loom"}
	default:
		if legacy {
			return []string{"retromcp-java", "ornithe-feather", "ornithe-gitcraft"}
		}
		return []string{"neoform-runtime", "minecraft-release-notes"}
	}
}

func doctorSourceRecoverySources(version, loader string) []string {
	ids := []string{"modrinth-api", "curseforge-api", "github-api", "github-releases"}
	if isLegacyMinecraftVersion(version) {
		ids = append(ids, "retromcp-java", "ornithe-feather", "ornithe-gitcraft")
		ids = append(ids, doctorBuildSources(loader, version)...)
	} else {
		ids = append(ids, "neoform", "neoform-runtime", "mcp-reborn", "mcpconfig")
	}
	return uniqueNonEmpty(ids)
}

func doctorLegacyRuntimeSources(version, loader string) []string {
	if !isLegacyMinecraftVersion(version) {
		return nil
	}
	normalized := normalizeDoctorLoader(loader)
	ids := []string{"legacyfix"}
	if normalized == "fabric" || normalized == "legacy-fabric" {
		ids = append(ids, "legacy-fabric-api", "legacy-fabric-looming")
	}
	if normalized == "ornithe" {
		ids = append(ids, "ornithe-ploceus", "ornithe-feather")
	}
	if normalized == "babric" || normalized == "stationapi" || strings.HasPrefix(strings.ToLower(strings.TrimSpace(version)), "b1.7.3") {
		ids = append(ids, "stationapi", "ornithe-feather")
	}
	if gameVersionAtMost(version, "1.7.10") {
		ids = append(ids, "retrofuturabootstrap", "unimixins", "lwjgl3ify", "hodgepodge", "retrofuturagradle")
	} else if gameVersionAtMost(version, "1.12.2") {
		ids = append(ids, "cleanroom", "fugue", "mixinbooter")
	}
	return uniqueNonEmpty(ids)
}

func doctorRiskLevel(score int) string {
	switch {
	case score >= 71:
		return "critical"
	case score >= 46:
		return "high"
	case score >= 21:
		return "moderate"
	default:
		return "low"
	}
}

func severityForPoints(points int) string {
	switch {
	case points >= 24:
		return "critical"
	case points >= 12:
		return "high"
	case points >= 5:
		return "warning"
	default:
		return "info"
	}
}

func sortDoctorFindings(findings []DoctorFinding) {
	for i := range findings {
		if findings[i].Code == "" {
			findings[i].Code = findings[i].ID
		}
		if findings[i].ID == "" {
			findings[i].ID = findings[i].Code
		}
	}
	rank := map[string]int{"critical": 4, "high": 3, "warning": 2, "info": 1}
	sort.SliceStable(findings, func(i, j int) bool {
		if rank[findings[i].Severity] != rank[findings[j].Severity] {
			return rank[findings[i].Severity] > rank[findings[j].Severity]
		}
		return findings[i].Title < findings[j].Title
	})
}

func collectDoctorSourceIDs(findings []DoctorFinding, steps []DoctorStep, reports []DoctorModReport) []string {
	ids := []string{}
	for _, finding := range findings {
		ids = append(ids, finding.SourceIDs...)
	}
	for _, step := range steps {
		ids = append(ids, step.SourceIDs...)
	}
	for _, report := range reports {
		ids = append(ids, report.RecommendedSourceIDs...)
	}
	return uniqueNonEmpty(ids)
}

func uniqueNonEmpty(values []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func appendUniqueString(values []string, value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return values
	}
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func parseManifestAttributes(data []byte) map[string]string {
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	attrs := map[string]string{}
	lastKey := ""
	for _, line := range lines {
		if strings.HasPrefix(line, " ") && lastKey != "" {
			attrs[lastKey] += strings.TrimPrefix(line, " ")
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			lastKey = ""
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key != "" {
			attrs[key] = value
			lastKey = key
		}
	}
	return attrs
}

func isMixinConfigName(lower string) bool {
	if !strings.HasSuffix(lower, ".json") {
		return false
	}
	base := filepath.Base(lower)
	return strings.Contains(base, "mixin") && (strings.Contains(base, "mixins") || strings.HasSuffix(base, ".mixin.json"))
}

func isNativeLibrary(lower string) bool {
	return strings.HasSuffix(lower, ".dll") || strings.HasSuffix(lower, ".so") || strings.HasSuffix(lower, ".dylib") || strings.HasSuffix(lower, ".jnilib")
}

func isSignatureFile(lower string) bool {
	if !strings.HasPrefix(lower, "meta-inf/") {
		return false
	}
	return strings.HasSuffix(lower, ".sf") || strings.HasSuffix(lower, ".rsa") || strings.HasSuffix(lower, ".dsa") || strings.HasSuffix(lower, ".ec")
}

func javaForClassMajor(major int) int {
	if major < 45 {
		return 0
	}
	if major <= 48 {
		return major - 44
	}
	return major - 44
}

type DoctorReportSummary struct {
	Artifacts          int `json:"artifacts"`
	Critical           int `json:"critical"`
	High               int `json:"high"`
	Warnings           int `json:"warnings"`
	Info               int `json:"info"`
	DuplicateModIDs    int `json:"duplicateModIds"`
	ExactDuplicateJars int `json:"exactDuplicateJars"`
	ExactDuplicates    int `json:"exactDuplicates"`
	ConflictingClasses int `json:"conflictingClasses"`
	MissingRequired    int `json:"missingRequired"`
	DeclaredConflicts  int `json:"declaredConflicts"`
	SourceAvailable    int `json:"sourceAvailable"`
	BinaryOnly         int `json:"binaryOnly"`
}

type DoctorReport struct {
	ID                   string                `json:"id"`
	CreatedAt            string                `json:"createdAt"`
	KnowledgeReviewedAt  string                `json:"knowledgeReviewedAt"`
	SourceGameVersion    string                `json:"sourceGameVersion"`
	SourceLoader         string                `json:"sourceLoader"`
	TargetGameVersion    string                `json:"targetGameVersion"`
	TargetLoader         string                `json:"targetLoader"`
	TargetJava           int                   `json:"targetJava"`
	ModsDirectory        string                `json:"modsDirectory"`
	Summary              DoctorReportSummary   `json:"summary"`
	GlobalFindings       []DoctorFinding       `json:"globalFindings"`
	Artifacts            []DoctorModReport     `json:"artifacts"`
	MigrationPlan        []DoctorStep          `json:"migrationPlan"`
	Tools                []DoctorTool          `json:"tools"`
	RepairPatterns       []DoctorRepairPattern `json:"repairPatterns"`
	RecommendedSourceIDs []string              `json:"recommendedSourceIds"`
	Errors               []DoctorScanError     `json:"errors,omitempty"`
}

func loadDoctorToolsPayload() (DoctorToolsPayload, error) {
	knowledge, err := loadDoctorKnowledge()
	if err != nil {
		return DoctorToolsPayload{}, err
	}
	detailedBytes, err := embeddedFiles.ReadFile("knowledge/doctor-tools.json")
	if err != nil {
		return DoctorToolsPayload{}, err
	}
	var detailedTools []DoctorTool
	if err := json.Unmarshal(detailedBytes, &detailedTools); err != nil {
		return DoctorToolsPayload{}, err
	}
	patternsBytes, err := embeddedFiles.ReadFile("assets/mod-doctor-repair-patterns.json")
	if err != nil {
		return DoctorToolsPayload{}, err
	}
	var patterns []DoctorRepairPattern
	if err := json.Unmarshal(patternsBytes, &patterns); err != nil {
		return DoctorToolsPayload{}, err
	}

	detailedByID := make(map[string]DoctorTool, len(detailedTools))
	for _, tool := range detailedTools {
		tool.ID = strings.TrimSpace(tool.ID)
		if tool.ID != "" {
			detailedByID[tool.ID] = tool
		}
	}

	payload := DoctorToolsPayload{ReviewedAt: knowledge.ReviewedAt, RepairPatterns: patterns}
	seen := map[string]bool{}
	categories := map[string]bool{}
	for _, source := range knowledge.Sources {
		tool := detailedByID[source.ID]
		tool.ID = source.ID
		if strings.TrimSpace(tool.Name) == "" {
			tool.Name = source.Name
		}
		if strings.TrimSpace(tool.Category) == "" {
			tool.Category = source.Category
		}
		if strings.TrimSpace(tool.Maturity) == "" {
			tool.Maturity = source.Maturity
		}
		if strings.TrimSpace(tool.Capability) == "" {
			tool.Capability = source.BestFor
		}
		if strings.TrimSpace(tool.BestFor) == "" {
			tool.BestFor = source.BestFor
		}
		if strings.TrimSpace(tool.OfficialURL) == "" {
			tool.OfficialURL = source.URL
		}
		if strings.TrimSpace(tool.RepositoryURL) == "" {
			tool.RepositoryURL = source.Repository
		}
		if strings.TrimSpace(tool.Integration) == "" {
			tool.Integration = source.Integration
		}
		if strings.TrimSpace(tool.CurrentEvidence) == "" {
			tool.CurrentEvidence = source.Notes
		}
		if source.Priority > tool.Priority {
			tool.Priority = source.Priority
		}
		if source.Status != "active" && source.Status != "official" {
			tool.Risks = appendUniqueString(tool.Risks, "Status: "+source.Status)
		}
		if source.Maturity == "experimental" || source.Maturity == "research" || source.Maturity == "legacy" {
			tool.Risks = appendUniqueString(tool.Risks, "Maturity: "+source.Maturity)
		}
		payload.Tools = append(payload.Tools, tool)
		seen[tool.ID] = true
		if tool.Category != "" {
			categories[tool.Category] = true
		}
	}
	for _, tool := range detailedTools {
		if tool.ID == "" || seen[tool.ID] {
			continue
		}
		payload.Tools = append(payload.Tools, tool)
		seen[tool.ID] = true
		if tool.Category != "" {
			categories[tool.Category] = true
		}
	}
	payload.Total = len(payload.Tools)
	for category := range categories {
		payload.Categories = append(payload.Categories, category)
	}
	sort.Strings(payload.Categories)
	sort.SliceStable(payload.Tools, func(i, j int) bool {
		if payload.Tools[i].Priority != payload.Tools[j].Priority {
			return payload.Tools[i].Priority > payload.Tools[j].Priority
		}
		if payload.Tools[i].Category != payload.Tools[j].Category {
			return payload.Tools[i].Category < payload.Tools[j].Category
		}
		return strings.ToLower(payload.Tools[i].Name) < strings.ToLower(payload.Tools[j].Name)
	})
	sortDoctorPatterns(payload.RepairPatterns)
	return payload, nil
}

func (a *App) handleDoctorTools(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	payload, err := loadDoctorToolsPayload()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, payload)
}

func (a *App) handleDoctorAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request DoctorScanRequest
	if r.Body != nil && r.ContentLength != 0 {
		if err := decodeJSON(r, &request); err != nil {
			writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
			return
		}
	}
	a.mu.RLock()
	settings := a.settings
	a.mu.RUnlock()
	targetGame := firstNonEmpty(strings.TrimSpace(request.TargetGameVersion), settings.GameVersion, "1.21.1")
	targetLoader := firstNonEmpty(strings.TrimSpace(request.TargetLoader), settings.Loader, "fabric")
	report, err := a.buildDoctorReport(targetGame, targetLoader)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (a *App) buildDoctorReport(targetGameVersion, targetLoader string) (DoctorReport, error) {
	a.mu.RLock()
	settings := a.settings
	a.mu.RUnlock()
	request := DoctorScanRequest{
		SourceGameVersion: firstNonEmpty(strings.TrimSpace(settings.GameVersion), strings.TrimSpace(targetGameVersion), "1.21.1"),
		SourceLoader:      firstNonEmpty(strings.TrimSpace(settings.Loader), strings.TrimSpace(targetLoader), "fabric"),
		TargetGameVersion: firstNonEmpty(strings.TrimSpace(targetGameVersion), strings.TrimSpace(settings.GameVersion), "1.21.1"),
		TargetLoader:      firstNonEmpty(strings.TrimSpace(targetLoader), strings.TrimSpace(settings.Loader), "fabric"),
	}
	scan, err := a.buildDoctorScan(request)
	if err != nil {
		return DoctorReport{}, err
	}
	tools, err := loadDoctorToolsPayload()
	if err != nil {
		return DoctorReport{}, err
	}
	report := DoctorReport{
		ID:                   scan.ID,
		CreatedAt:            scan.CreatedAt,
		KnowledgeReviewedAt:  scan.KnowledgeReviewedAt,
		SourceGameVersion:    scan.SourceGameVersion,
		SourceLoader:         scan.SourceLoader,
		TargetGameVersion:    scan.TargetGameVersion,
		TargetLoader:         scan.TargetLoader,
		TargetJava:           scan.TargetJava,
		ModsDirectory:        scan.ModsDirectory,
		GlobalFindings:       scan.GlobalFindings,
		Artifacts:            scan.Mods,
		MigrationPlan:        scan.MigrationPlan,
		Tools:                tools.Tools,
		RepairPatterns:       tools.RepairPatterns,
		RecommendedSourceIDs: scan.RecommendedSourceIDs,
		Errors:               scan.Errors,
		Summary: DoctorReportSummary{
			Artifacts:          scan.Summary.TotalMods,
			DuplicateModIDs:    scan.Summary.DuplicateModIDs,
			ExactDuplicateJars: scan.Summary.ExactDuplicateJars,
			ExactDuplicates:    scan.Summary.ExactDuplicateClasses,
			ConflictingClasses: scan.Summary.ConflictingDuplicateClasses,
			MissingRequired:    scan.Summary.MissingRequiredDependencies,
			DeclaredConflicts:  scan.Summary.DeclaredConflicts,
			SourceAvailable:    scan.Summary.SourceAvailable,
			BinaryOnly:         scan.Summary.BinaryOnly,
		},
	}
	for _, finding := range report.GlobalFindings {
		countDoctorReportFinding(&report.Summary, finding)
	}
	for _, artifact := range report.Artifacts {
		for _, finding := range artifact.Findings {
			countDoctorReportFinding(&report.Summary, finding)
		}
	}
	return report, nil
}

func countDoctorReportFinding(summary *DoctorReportSummary, finding DoctorFinding) {
	switch finding.Severity {
	case "critical":
		summary.Critical++
	case "high":
		summary.High++
	case "warning":
		summary.Warnings++
	default:
		summary.Info++
	}
}

func doctorParseFabricMetadata(data []byte) ([]string, []DoctorDependency) {
	var metadata struct {
		ID         string         `json:"id"`
		Provides   []string       `json:"provides"`
		Depends    map[string]any `json:"depends"`
		Recommends map[string]any `json:"recommends"`
		Suggests   map[string]any `json:"suggests"`
		Breaks     map[string]any `json:"breaks"`
		Conflicts  map[string]any `json:"conflicts"`
	}
	if json.Unmarshal(data, &metadata) != nil {
		return nil, nil
	}
	ids := append([]string{}, metadata.ID)
	ids = append(ids, metadata.Provides...)
	dependencies := []DoctorDependency{}
	appendRelationship := func(values map[string]any, relationship string) {
		keys := make([]string, 0, len(values))
		for key := range values {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			dependencies = append(dependencies, DoctorDependency{Owner: metadata.ID, ModID: key, Version: fmt.Sprint(values[key]), Relationship: relationship, Source: "fabric.mod.json"})
		}
	}
	appendRelationship(metadata.Depends, "required")
	appendRelationship(metadata.Recommends, "recommended")
	appendRelationship(metadata.Suggests, "optional")
	appendRelationship(metadata.Breaks, "conflict")
	appendRelationship(metadata.Conflicts, "conflict")
	return uniqueNonEmpty(ids), dependencies
}

func doctorParseQuiltMetadata(data []byte) ([]string, []DoctorDependency) {
	var raw map[string]any
	if json.Unmarshal(data, &raw) != nil {
		return nil, nil
	}
	loader, _ := raw["quilt_loader"].(map[string]any)
	id, _ := loader["id"].(string)
	ids := []string{id}
	if values, ok := loader["provides"].([]any); ok {
		for _, value := range values {
			switch item := value.(type) {
			case string:
				ids = append(ids, item)
			case map[string]any:
				provided, _ := item["id"].(string)
				ids = append(ids, provided)
			}
		}
	}
	dependencies := []DoctorDependency{}
	parseList := func(key, relationship string) {
		values, _ := loader[key].([]any)
		for _, value := range values {
			dependency := DoctorDependency{Owner: id, Relationship: relationship, Source: "quilt.mod.json"}
			switch item := value.(type) {
			case string:
				dependency.ModID = item
			case map[string]any:
				dependency.ModID, _ = item["id"].(string)
				if versions, ok := item["versions"]; ok {
					dependency.Version = fmt.Sprint(versions)
				}
				if dependency.ModID == "" {
					for candidate, constraint := range item {
						dependency.ModID = candidate
						dependency.Version = fmt.Sprint(constraint)
						break
					}
				}
			}
			if strings.TrimSpace(dependency.ModID) != "" {
				dependencies = append(dependencies, dependency)
			}
		}
	}
	parseList("depends", "required")
	parseList("recommends", "recommended")
	parseList("suggests", "optional")
	parseList("breaks", "conflict")
	return uniqueNonEmpty(ids), dependencies
}

func doctorParseTOMLMetadata(text, source string) ([]string, []DoctorDependency) {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	ids := []string{}
	dependencies := []DoctorDependency{}
	section := ""
	owner := ""
	current := DoctorDependency{Source: source}
	mandatory := false
	flush := func() {
		if section != "dependency" || strings.TrimSpace(current.ModID) == "" {
			return
		}
		if current.Relationship == "" {
			if mandatory {
				current.Relationship = "required"
			} else {
				current.Relationship = "optional"
			}
		}
		current.Owner = owner
		dependencies = append(dependencies, current)
		current = DoctorDependency{Source: source}
		mandatory = false
	}
	for _, rawLine := range lines {
		line := strings.TrimSpace(strings.SplitN(rawLine, "#", 2)[0])
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "[[") && strings.HasSuffix(line, "]]") {
			flush()
			header := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(line, "[["), "]]"))
			switch {
			case header == "mods":
				section = "mods"
			case strings.HasPrefix(header, "dependencies."):
				section = "dependency"
				owner = strings.Trim(strings.TrimPrefix(header, "dependencies."), "\"'")
				current = DoctorDependency{Owner: owner, Source: source}
				mandatory = false
			default:
				section = header
			}
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), "\"'")
		switch section {
		case "mods":
			if key == "modId" && value != "" {
				ids = append(ids, value)
				if owner == "" {
					owner = value
				}
			}
		case "dependency":
			switch key {
			case "modId":
				current.ModID = value
			case "versionRange":
				current.Version = value
			case "side":
				current.Side = value
			case "mandatory":
				mandatory = strings.EqualFold(value, "true")
			case "type":
				switch strings.ToLower(value) {
				case "required":
					current.Relationship = "required"
				case "optional":
					current.Relationship = "optional"
				case "incompatible", "discouraged":
					current.Relationship = "conflict"
				}
			}
		}
	}
	flush()
	return uniqueNonEmpty(ids), dependencies
}

func uniqueDoctorDependencies(values []DoctorDependency) []DoctorDependency {
	seen := map[string]bool{}
	out := []DoctorDependency{}
	for _, value := range values {
		value.ModID = strings.TrimSpace(value.ModID)
		if value.ModID == "" {
			continue
		}
		key := strings.ToLower(strings.Join([]string{value.Owner, value.ModID, value.Version, value.Relationship, value.Side, value.Source}, "\x00"))
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, value)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].ModID != out[j].ModID {
			return out[i].ModID < out[j].ModID
		}
		return out[i].Relationship < out[j].Relationship
	})
	return out
}

func doctorArtifactSupportsLoader(loaders []string, target string) bool {
	target = normalizeDoctorLoader(target)
	if target == "" || target == "any" || target == "vanilla" || len(loaders) == 0 {
		return true
	}
	for _, loader := range loaders {
		if normalizeDoctorLoader(loader) == target {
			return true
		}
	}
	return false
}

func decideDoctorPorting(local LocalModFile, signals JarSignals, request DoctorScanRequest) DoctorPortingDecision {
	decision := DoctorPortingDecision{}
	sourceRecovery := doctorSourceRecoverySources(request.SourceGameVersion, request.SourceLoader)
	targetBuild := doctorBuildSources(request.TargetLoader, request.TargetGameVersion)
	legacyRuntime := doctorLegacyRuntimeSources(request.SourceGameVersion, request.SourceLoader)
	switch {
	case strings.TrimSpace(local.Metadata.SourceURL) != "":
		decision.PrimaryRoute = "source-migrate"
		decision.Confidence = "high"
		decision.Reasons = []string{"The artifact declares a source repository.", "A reproducible, era-correct source build can expose compiler, mapping, loader, data, and runtime API failures directly."}
		decision.SourceIDs = append(append(sourceRecovery, targetBuild...), "openrewrite", "japicmp")
	case len(signals.Coremods)+len(signals.TransformationServices)+len(signals.MixinConfigs) > 0:
		decision.PrimaryRoute = "binary-reconstruct"
		decision.Confidence = "low"
		decision.Reasons = []string{"Source is not declared.", "The artifact transforms bytecode, so metadata-only conversion cannot preserve behavior.", "Reconstruct against the correct historical mappings and launcher before retargeting transformations."}
		decision.SourceIDs = append(sourceRecovery, "vineflower", "cfr", "enigma", "asm", "mixin-source")
		decision.SourceIDs = append(decision.SourceIDs, legacyRuntime...)
	default:
		decision.PrimaryRoute = "source-resolve-or-reconstruct"
		decision.Confidence = "medium"
		decision.Reasons = []string{"Resolve upstream source through provider and release metadata first.", "Use era-specific game/toolchain reconstruction plus controlled dual-decompiler comparison only when source remains unavailable."}
		decision.SourceIDs = append(sourceRecovery, "vineflower", "cfr")
		decision.SourceIDs = append(decision.SourceIDs, targetBuild...)
	}
	if isLegacyMinecraftVersion(request.SourceGameVersion) || isLegacyMinecraftVersion(request.TargetGameVersion) {
		decision.Reasons = append(decision.Reasons, "Runtime modernization and compatibility layers remain separate candidates until the source/data migration is independently proven.")
		decision.SourceIDs = append(decision.SourceIDs, legacyRuntime...)
	}
	decision.SourceIDs = uniqueNonEmpty(decision.SourceIDs)
	return decision
}

func applyDoctorGraphAnalysis(scan *DoctorScan) {
	modOwners := map[string][]int{}
	jarOwners := map[string][]int{}
	classOwners := map[string]map[string][]int{}
	reportIDs := make([][]string, len(scan.Mods))
	for i := range scan.Mods {
		ids := append([]string{}, scan.Mods[i].Signals.ModIDs...)
		if len(ids) == 0 && scan.Mods[i].Local.Metadata.ModID != "" {
			ids = []string{scan.Mods[i].Local.Metadata.ModID}
			scan.Mods[i].Signals.ModIDs = append([]string{}, ids...)
		}
		for _, id := range ids {
			id = strings.ToLower(strings.TrimSpace(id))
			if id == "" {
				continue
			}
			reportIDs[i] = appendUniqueString(reportIDs[i], id)
			modOwners[id] = append(modOwners[id], i)
		}
		if hash := strings.ToLower(strings.TrimSpace(scan.Mods[i].Local.SHA512)); hash != "" {
			jarOwners[hash] = append(jarOwners[hash], i)
		}
		for className, digest := range scan.Mods[i].Signals.ClassDigests {
			if classOwners[className] == nil {
				classOwners[className] = map[string][]int{}
			}
			classOwners[className][digest] = append(classOwners[className][digest], i)
		}
	}
	for modID, owners := range modOwners {
		owners = uniqueDoctorIndexes(owners)
		if len(owners) < 2 {
			continue
		}
		scan.Summary.DuplicateModIDs++
		finding := DoctorFinding{ID: "duplicate-mod-id", Severity: "critical", Category: "dependency-graph", Title: "Duplicate mod ID is provided by multiple JARs", Evidence: modID + " in " + strings.Join(doctorModFiles(scan.Mods, owners), ", "), Action: "Resolve the canonical artifact and remove or merge only after comparing versions, providers, embedded libraries, and dependent requirements.", SourceIDs: []string{"modrinth-api", "curseforge-api", "github-api"}}
		scan.GlobalFindings = append(scan.GlobalFindings, finding)
		for _, index := range owners {
			scan.Mods[index].Findings = append(scan.Mods[index].Findings, finding)
		}
	}
	for _, owners := range jarOwners {
		owners = uniqueDoctorIndexes(owners)
		if len(owners) < 2 {
			continue
		}
		scan.Summary.ExactDuplicateJars++
		finding := DoctorFinding{ID: "exact-duplicate-jar", Severity: "critical", Category: "classpath", Title: "Byte-identical JAR is installed more than once", Evidence: strings.Join(doctorModFiles(scan.Mods, owners), ", "), Action: "Keep one canonical copy and preserve any disabled-state choice explicitly.", SourceIDs: []string{"jdk-tools", "recaf"}}
		scan.GlobalFindings = append(scan.GlobalFindings, finding)
		for _, index := range owners {
			scan.Mods[index].Findings = append(scan.Mods[index].Findings, finding)
		}
	}

	// These IDs are supplied by the game or loader runtime. Fabric API and QSL are deliberately
	// excluded because they are separately installed mod dependencies and must never be hidden.
	ignoredRequired := map[string]bool{"minecraft": true, "java": true, "fabricloader": true, "fabric-loader": true, "forge": true, "neoforge": true, "quilt_loader": true, "quilt-loader": true, "fml": true, "modlauncher": true}
	requiredDeps := map[string]map[string]bool{}
	edgeSeen := map[string]bool{}
	for i := range scan.Mods {
		owner := ""
		if len(reportIDs[i]) > 0 {
			owner = reportIDs[i][0]
		}
		for _, dependency := range scan.Mods[i].Signals.Dependencies {
			id := strings.ToLower(strings.TrimSpace(dependency.ModID))
			if id == "" {
				continue
			}
			relationship := strings.ToLower(strings.TrimSpace(dependency.Relationship))
			present := ignoredRequired[id] || len(modOwners[id]) > 0
			if owner != "" && !ignoredRequired[id] {
				key := owner + "\x00" + id + "\x00" + relationship + "\x00" + dependency.Version + "\x00" + dependency.Source
				if !edgeSeen[key] {
					edgeSeen[key] = true
					scan.DependencyEdges = append(scan.DependencyEdges, DoctorDependencyEdge{ModID: owner, DependsOn: id, Version: dependency.Version, Relationship: relationship, Source: dependency.Source})
				}
			}
			if relationship == "required" && !present {
				scan.Summary.MissingRequiredDependencies++
				scan.Mods[i].Findings = append(scan.Mods[i].Findings, DoctorFinding{ID: "missing-required-dependency", Severity: "critical", Category: "dependency-graph", Title: "Required dependency is missing", Evidence: strings.TrimSpace(dependency.ModID + " " + dependency.Version), Action: "Resolve an exact compatible dependency build or port the dependency first; never delete the declaration merely to force loading.", SourceIDs: []string{"modrinth-api", "curseforge-api", "github-api"}})
			}
			if relationship == "required" && present && owner != "" && id != owner && !ignoredRequired[id] {
				if requiredDeps[owner] == nil {
					requiredDeps[owner] = map[string]bool{}
				}
				requiredDeps[owner][id] = true
			}
			if relationship == "conflict" && present {
				scan.Summary.DeclaredConflicts++
				scan.Mods[i].Findings = append(scan.Mods[i].Findings, DoctorFinding{ID: "declared-conflict", Severity: "critical", Category: "dependency-graph", Title: "Declared incompatible mod is installed", Evidence: dependency.ModID + " is present", Action: "Resolve the overlap through a version change or a narrow compatibility adapter before launching both mods together.", SourceIDs: []string{"sinytra-probe", "mixin-source", "classgraph"}})
			}
		}
	}
	sort.Slice(scan.DependencyEdges, func(i, j int) bool {
		if scan.DependencyEdges[i].ModID != scan.DependencyEdges[j].ModID {
			return scan.DependencyEdges[i].ModID < scan.DependencyEdges[j].ModID
		}
		if scan.DependencyEdges[i].DependsOn != scan.DependencyEdges[j].DependsOn {
			return scan.DependencyEdges[i].DependsOn < scan.DependencyEdges[j].DependsOn
		}
		return scan.DependencyEdges[i].Relationship < scan.DependencyEdges[j].Relationship
	})

	for className, digests := range classOwners {
		allOwners := []int{}
		for _, owners := range digests {
			allOwners = append(allOwners, owners...)
		}
		allOwners = uniqueDoctorIndexes(allOwners)
		if len(allOwners) < 2 {
			continue
		}
		if len(digests) == 1 {
			scan.Summary.ExactDuplicateClasses++
			finding := DoctorFinding{ID: "duplicate-classes", Severity: "critical", Category: "classpath", Title: "Exact duplicate class is bundled by multiple mods", Evidence: className + " in " + strings.Join(doctorModFiles(scan.Mods, allOwners), ", "), Action: "Identify the shared library owner and keep exactly one compatible copy through supported jar-in-jar or dependency packaging.", SourceIDs: []string{"classgraph", "jarsplitter", "recaf"}}
			scan.GlobalFindings = append(scan.GlobalFindings, finding)
			for _, index := range allOwners {
				scan.Mods[index].Findings = append(scan.Mods[index].Findings, finding)
			}
		} else {
			scan.Summary.ConflictingDuplicateClasses++
			finding := DoctorFinding{ID: "duplicate-class-conflict", Severity: "critical", Category: "classpath", Title: "Different class implementations share one classpath name", Evidence: className + " in " + strings.Join(doctorModFiles(scan.Mods, allOwners), ", "), Action: "Relocate the private library or reconcile the implementations; classpath order is not a safe compatibility strategy.", SourceIDs: []string{"classgraph", "recaf", "asm", "japicmp"}}
			scan.GlobalFindings = append(scan.GlobalFindings, finding)
			for _, index := range allOwners {
				scan.Mods[index].Findings = append(scan.Mods[index].Findings, finding)
			}
		}
	}

	nodes := make([]string, 0, len(modOwners))
	for id := range modOwners {
		nodes = append(nodes, id)
	}
	scan.DependencyOrder, scan.DependencyCycles = doctorDependencyOrderAndCycles(nodes, requiredDeps)
	scan.Summary.DependencyCycles = len(scan.DependencyCycles)
	for _, cycle := range scan.DependencyCycles {
		if len(cycle) == 0 {
			continue
		}
		evidence := strings.Join(cycle, " -> ") + " -> " + cycle[0]
		finding := DoctorFinding{ID: "required-dependency-cycle", Severity: "critical", Category: "dependency-graph", Title: "Required dependencies form a cycle", Evidence: evidence, Action: "Resolve the cycle as one coordinated compatibility unit. Rebuild or replace the participating versions together, then verify initialization order and runtime behavior.", SourceIDs: []string{"classgraph", "gradle-toolchains", "repair-brain"}}
		scan.GlobalFindings = append(scan.GlobalFindings, finding)
		inCycle := map[string]bool{}
		for _, id := range cycle {
			inCycle[id] = true
		}
		for id, owners := range modOwners {
			if !inCycle[id] {
				continue
			}
			for _, index := range uniqueDoctorIndexes(owners) {
				scan.Mods[index].Findings = append(scan.Mods[index].Findings, finding)
			}
		}
	}

	graphCounts := scan.Summary
	for i := range scan.Mods {
		sortDoctorFindings(scan.Mods[i].Findings)
		scan.Mods[i].RiskScore = doctorRiskFromFindings(scan.Mods[i].Findings, scan.Mods[i].RiskScore)
		scan.Mods[i].RiskLevel = doctorRiskLevel(scan.Mods[i].RiskScore)
		scan.Mods[i].RecommendedSourceIDs = collectDoctorSourceIDs(scan.Mods[i].Findings, scan.Mods[i].Plan, nil)
	}
	sort.Slice(scan.Mods, func(i, j int) bool {
		if scan.Mods[i].RiskScore != scan.Mods[j].RiskScore {
			return scan.Mods[i].RiskScore > scan.Mods[j].RiskScore
		}
		return strings.ToLower(scan.Mods[i].Local.Filename) < strings.ToLower(scan.Mods[j].Local.Filename)
	})
	base := summarizeDoctorReports(scan.Mods)
	base.DuplicateModIDs = graphCounts.DuplicateModIDs
	base.ExactDuplicateJars = graphCounts.ExactDuplicateJars
	base.ExactDuplicateClasses = graphCounts.ExactDuplicateClasses
	base.ConflictingDuplicateClasses = graphCounts.ConflictingDuplicateClasses
	base.MissingRequiredDependencies = graphCounts.MissingRequiredDependencies
	base.DeclaredConflicts = graphCounts.DeclaredConflicts
	base.DependencyCycles = graphCounts.DependencyCycles
	for i := range scan.Mods {
		countDoctorScanFindings(&base, scan.Mods[i].Findings)
	}
	for i := range scan.GlobalFindings {
		if scan.GlobalFindings[i].Code == "" {
			scan.GlobalFindings[i].Code = scan.GlobalFindings[i].ID
		}
	}
	countDoctorScanFindings(&base, scan.GlobalFindings)
	scan.Summary = base
	sortDoctorFindings(scan.GlobalFindings)
}

func doctorDependencyOrderAndCycles(nodes []string, dependencies map[string]map[string]bool) ([]string, [][]string) {
	nodeSet := map[string]bool{}
	for _, node := range nodes {
		node = strings.ToLower(strings.TrimSpace(node))
		if node != "" {
			nodeSet[node] = true
		}
	}
	for owner, deps := range dependencies {
		nodeSet[owner] = true
		for dep := range deps {
			nodeSet[dep] = true
		}
	}
	all := make([]string, 0, len(nodeSet))
	for node := range nodeSet {
		all = append(all, node)
	}
	sort.Strings(all)

	index := 0
	indices := map[string]int{}
	lowlink := map[string]int{}
	onStack := map[string]bool{}
	stack := []string{}
	cycles := [][]string{}
	var strongConnect func(string)
	strongConnect = func(node string) {
		indices[node] = index
		lowlink[node] = index
		index++
		stack = append(stack, node)
		onStack[node] = true
		deps := make([]string, 0, len(dependencies[node]))
		for dep := range dependencies[node] {
			deps = append(deps, dep)
		}
		sort.Strings(deps)
		for _, dep := range deps {
			if _, seen := indices[dep]; !seen {
				strongConnect(dep)
				if lowlink[dep] < lowlink[node] {
					lowlink[node] = lowlink[dep]
				}
			} else if onStack[dep] && indices[dep] < lowlink[node] {
				lowlink[node] = indices[dep]
			}
		}
		if lowlink[node] != indices[node] {
			return
		}
		component := []string{}
		for len(stack) > 0 {
			last := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			onStack[last] = false
			component = append(component, last)
			if last == node {
				break
			}
		}
		sort.Strings(component)
		if len(component) > 1 || (len(component) == 1 && dependencies[component[0]][component[0]]) {
			cycles = append(cycles, component)
		}
	}
	for _, node := range all {
		if _, seen := indices[node]; !seen {
			strongConnect(node)
		}
	}
	sort.Slice(cycles, func(i, j int) bool { return strings.Join(cycles[i], "\x00") < strings.Join(cycles[j], "\x00") })

	indegree := map[string]int{}
	dependents := map[string]map[string]bool{}
	for _, node := range all {
		indegree[node] = 0
	}
	for owner, deps := range dependencies {
		for dep := range deps {
			if owner == dep {
				continue
			}
			if dependents[dep] == nil {
				dependents[dep] = map[string]bool{}
			}
			if !dependents[dep][owner] {
				dependents[dep][owner] = true
				indegree[owner]++
			}
		}
	}
	ready := []string{}
	for _, node := range all {
		if indegree[node] == 0 {
			ready = append(ready, node)
		}
	}
	order := []string{}
	seenOrder := map[string]bool{}
	for len(ready) > 0 {
		node := ready[0]
		ready = ready[1:]
		if seenOrder[node] {
			continue
		}
		seenOrder[node] = true
		order = append(order, node)
		next := make([]string, 0, len(dependents[node]))
		for dependent := range dependents[node] {
			next = append(next, dependent)
		}
		sort.Strings(next)
		for _, dependent := range next {
			indegree[dependent]--
			if indegree[dependent] == 0 {
				ready = append(ready, dependent)
				sort.Strings(ready)
			}
		}
	}
	for _, node := range all {
		if !seenOrder[node] {
			order = append(order, node)
		}
	}
	return order, cycles
}

func doctorRiskFromFindings(findings []DoctorFinding, baseline int) int {
	weights := map[string]int{"critical": 24, "high": 12, "warning": 5, "info": 1}
	score := 0
	seen := map[string]bool{}
	for _, finding := range findings {
		key := firstNonEmpty(finding.Code, finding.ID) + "\x00" + finding.Evidence
		if seen[key] {
			continue
		}
		seen[key] = true
		score += weights[finding.Severity]
	}
	if score < baseline {
		score = baseline
	}
	if score > 100 {
		score = 100
	}
	return score
}

func countDoctorScanFindings(summary *DoctorScanSummary, findings []DoctorFinding) {
	for _, finding := range findings {
		switch finding.Severity {
		case "critical":
			summary.CriticalFindings++
		case "high":
			summary.HighFindings++
		case "warning":
			summary.WarningFindings++
		default:
			summary.InfoFindings++
		}
	}
}

func doctorModFiles(reports []DoctorModReport, indexes []int) []string {
	files := []string{}
	for _, index := range uniqueDoctorIndexes(indexes) {
		if index >= 0 && index < len(reports) {
			files = append(files, reports[index].Local.Filename)
		}
	}
	sort.Strings(files)
	return files
}

func uniqueDoctorIndexes(values []int) []int {
	seen := map[int]bool{}
	out := []int{}
	for _, value := range values {
		if seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	sort.Ints(out)
	return out
}
