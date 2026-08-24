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
	repairLabSchemaVersion      = 1
	repairLabMaxUploadBytes     = int64(768 << 20)
	repairLabMaxExtractedBytes  = int64(4 << 30)
	repairLabMaxEntryBytes      = int64(1 << 30)
	repairLabMaxArchiveEntries  = 75000
	repairLabConfirmationPhrase = "I_UNDERSTAND_BUILD_SCRIPTS_EXECUTE_CODE"
)

var (
	errRepairSessionNotFound = errors.New("repair lab session not found")
	errRepairSessionBusy     = errors.New("repair lab session is running")
)

type PortingBuildRun struct {
	SchemaVersion int                    `json:"schemaVersion"`
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	State         string                 `json:"state"`
	Phase         string                 `json:"phase"`
	CreatedAt     string                 `json:"createdAt"`
	UpdatedAt     string                 `json:"updatedAt"`
	Revision      int                    `json:"revision"`
	Source        RepairSourceSnapshot   `json:"source"`
	Project       RepairProjectProfile   `json:"project"`
	Target        *AtlasResolveRequest   `json:"target,omitempty"`
	Resolution    *AtlasResolution       `json:"resolution,omitempty"`
	Changes       []RepairChange         `json:"changes,omitempty"`
	Runs          []RepairCommandRun     `json:"runs,omitempty"`
	Artifacts     []RepairArtifact       `json:"artifacts,omitempty"`
	Exports       []RepairExport         `json:"exports,omitempty"`
	Security      RepairSecurityBoundary `json:"security"`
	Warnings      []string               `json:"warnings,omitempty"`
	LastError     string                 `json:"lastError,omitempty"`
	Paths         RepairSessionPaths     `json:"paths"`
}

type RepairSourceSnapshot struct {
	Filename       string `json:"filename"`
	Size           int64  `json:"size"`
	SHA256         string `json:"sha256"`
	TreeSHA256     string `json:"treeSha256"`
	FileCount      int    `json:"fileCount"`
	ExtractedBytes int64  `json:"extractedBytes"`
	ImportedAt     string `json:"importedAt"`
}

type RepairProjectProfile struct {
	ProjectRoot       string                `json:"projectRoot"`
	BuildSystem       string                `json:"buildSystem"`
	Wrapper           string                `json:"wrapper,omitempty"`
	WrapperSHA256     string                `json:"wrapperSha256,omitempty"`
	Loader            string                `json:"loader"`
	GameVersion       string                `json:"gameVersion,omitempty"`
	JavaMajor         int                   `json:"javaMajor,omitempty"`
	BuildFiles        []string              `json:"buildFiles,omitempty"`
	MetadataFiles     []string              `json:"metadataFiles,omitempty"`
	Modules           []string              `json:"modules,omitempty"`
	AvailableCommands []RepairCommandChoice `json:"availableCommands,omitempty"`
	Confidence        string                `json:"confidence"`
	Fingerprint       string                `json:"fingerprint"`
	Signals           []string              `json:"signals,omitempty"`
	Warnings          []string              `json:"warnings,omitempty"`
}

type RepairCommandChoice struct {
	ID          string   `json:"id"`
	Label       string   `json:"label"`
	Description string   `json:"description"`
	Arguments   []string `json:"arguments"`
}

type RepairChange struct {
	File       string `json:"file"`
	Field      string `json:"field"`
	Before     string `json:"before,omitempty"`
	After      string `json:"after,omitempty"`
	Applied    bool   `json:"applied"`
	Confidence string `json:"confidence"`
	Reason     string `json:"reason"`
}

type RepairCommandRun struct {
	ID               string   `json:"id"`
	Action           string   `json:"action"`
	State            string   `json:"state"`
	Command          []string `json:"command"`
	WorkingDirectory string   `json:"workingDirectory"`
	StartedAt        string   `json:"startedAt"`
	FinishedAt       string   `json:"finishedAt,omitempty"`
	ExitCode         int      `json:"exitCode"`
	TimedOut         bool     `json:"timedOut"`
	Cancelled        bool     `json:"cancelled"`
	LogFile          string   `json:"logFile"`
	LogTail          string   `json:"logTail,omitempty"`
	Error            string   `json:"error,omitempty"`
}

type RepairArtifact struct {
	Name          string   `json:"name"`
	RelativePath  string   `json:"relativePath"`
	Size          int64    `json:"size"`
	SHA256        string   `json:"sha256"`
	Kind          string   `json:"kind"`
	ModIDs        []string `json:"modIds,omitempty"`
	ClassCount    int      `json:"classCount,omitempty"`
	JavaMajor     int      `json:"javaMajor,omitempty"`
	DownloadIndex int      `json:"downloadIndex"`
}

type RepairExport struct {
	Kind          string `json:"kind"`
	Name          string `json:"name"`
	RelativePath  string `json:"relativePath"`
	Size          int64  `json:"size"`
	SHA256        string `json:"sha256"`
	CreatedAt     string `json:"createdAt"`
	DownloadIndex int    `json:"downloadIndex"`
}

type RepairSecurityBoundary struct {
	ImmutableOriginal        bool     `json:"immutableOriginal"`
	ZipSlipProtection        bool     `json:"zipSlipProtection"`
	SymlinkProtection        bool     `json:"symlinkProtection"`
	ArchiveBombLimits        bool     `json:"archiveBombLimits"`
	SanitizedEnvironment     bool     `json:"sanitizedEnvironment"`
	DedicatedToolCaches      bool     `json:"dedicatedToolCaches"`
	ExplicitCodeExecutionAck bool     `json:"explicitCodeExecutionAck"`
	SourceVerifiedAfterRuns  bool     `json:"sourceVerifiedAfterRuns"`
	Limitations              []string `json:"limitations"`
}

type RepairSessionPaths struct {
	Root            string `json:"root"`
	OriginalArchive string `json:"originalArchive"`
	ImmutableSource string `json:"immutableSource"`
	WorkingCopy     string `json:"workingCopy"`
	Logs            string `json:"logs"`
	Exports         string `json:"exports"`
	ReceiptJSON     string `json:"receiptJson"`
	ReceiptMarkdown string `json:"receiptMarkdown"`
}

type RepairLabStatus struct {
	SchemaVersion      int                       `json:"schemaVersion"`
	Ready              bool                      `json:"ready"`
	Root               string                    `json:"root"`
	Sessions           []RepairSessionSummary    `json:"sessions"`
	Tools              []RepairHostTool          `json:"tools"`
	Atlas              *AtlasSummary             `json:"atlas,omitempty"`
	Brain              *BrainStatus              `json:"brain,omitempty"`
	BrainError         string                    `json:"brainError,omitempty"`
	ExecutionPhrase    string                    `json:"executionPhrase"`
	SecurityBoundaries RepairSecurityBoundary    `json:"securityBoundaries"`
	Capabilities       []RepairCapabilitySummary `json:"capabilities"`
}

type RepairSessionSummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	State       string `json:"state"`
	Phase       string `json:"phase"`
	UpdatedAt   string `json:"updatedAt"`
	Source      string `json:"source"`
	ProjectRoot string `json:"projectRoot"`
	Loader      string `json:"loader"`
	GameVersion string `json:"gameVersion"`
	Target      string `json:"target,omitempty"`
	LastError   string `json:"lastError,omitempty"`
}

type RepairHostTool struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Path    string `json:"path,omitempty"`
	Version string `json:"version,omitempty"`
	Ready   bool   `json:"ready"`
	Role    string `json:"role"`
}

type RepairCapabilitySummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	State       string `json:"state"`
}

type RepairPrepareRequest struct {
	SessionID         string `json:"sessionId"`
	TargetGameVersion string `json:"targetGameVersion"`
	TargetLoader      string `json:"targetLoader"`
}

type RepairRunRequest struct {
	SessionID   string `json:"sessionId"`
	Action      string `json:"action"`
	ConfirmCode string `json:"confirmCode"`
	TimeoutMin  int    `json:"timeoutMinutes"`
}

type RepairSessionRequest struct {
	SessionID string `json:"sessionId"`
}

type RepairExportRequest struct {
	SessionID        string `json:"sessionId"`
	IncludeArtifacts bool   `json:"includeArtifacts"`
}

func (a *App) repairLabRoot() string {
	return filepath.Join(a.cfgDir, "repair-lab")
}

func repairRunSecurityBoundary() RepairSecurityBoundary {
	return RepairSecurityBoundary{
		ImmutableOriginal:        true,
		ZipSlipProtection:        true,
		SymlinkProtection:        true,
		ArchiveBombLimits:        true,
		SanitizedEnvironment:     true,
		DedicatedToolCaches:      true,
		ExplicitCodeExecutionAck: true,
		SourceVerifiedAfterRuns:  true,
		Limitations: []string{
			"Build wrappers and build scripts are project code. They are executed only after explicit acknowledgement, but this release does not claim operating-system or virtual-machine isolation.",
			"Automatic edits are limited to recognized version, loader, mapping, Java, metadata, and pack-format fields. Semantic API rewrites still require source-aware repair work and runtime proof.",
		},
	}
}

func (a *App) newPortingRun(name string) (*PortingBuildRun, error) {
	now := time.Now().UTC()
	id := fmt.Sprintf("repair-%s-%s", now.Format("20060102T150405Z"), randomToken(6))
	root := filepath.Join(a.repairLabRoot(), "sessions", id)
	paths := RepairSessionPaths{
		Root:            root,
		OriginalArchive: filepath.Join(root, "original", "source.zip"),
		ImmutableSource: filepath.Join(root, "source"),
		WorkingCopy:     filepath.Join(root, "work"),
		Logs:            filepath.Join(root, "logs"),
		Exports:         filepath.Join(root, "exports"),
		ReceiptJSON:     filepath.Join(root, "receipts", "repair-receipt.json"),
		ReceiptMarkdown: filepath.Join(root, "receipts", "repair-receipt.md"),
	}
	for _, dir := range []string{filepath.Dir(paths.OriginalArchive), paths.ImmutableSource, paths.WorkingCopy, paths.Logs, paths.Exports, filepath.Dir(paths.ReceiptJSON)} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}
	run := &PortingBuildRun{
		SchemaVersion: repairLabSchemaVersion,
		ID:            id,
		Name:          strings.TrimSpace(name),
		State:         "importing",
		Phase:         "source-intake",
		CreatedAt:     now.Format(time.RFC3339),
		UpdatedAt:     now.Format(time.RFC3339),
		Revision:      1,
		Security:      repairRunSecurityBoundary(),
		Paths:         paths,
	}
	if run.Name == "" {
		run.Name = "Imported mod source"
	}
	return run, nil
}

func (a *App) sessionManifestPath(id string) (string, error) {
	id = strings.TrimSpace(id)
	if !validRepairID(id) {
		return "", errors.New("invalid repair session ID")
	}
	return filepath.Join(a.repairLabRoot(), "sessions", id, "session.json"), nil
}

func validRepairID(id string) bool {
	if len(id) < 8 || len(id) > 100 {
		return false
	}
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			continue
		}
		return false
	}
	return true
}

func clonePortingRun(run *PortingBuildRun) *PortingBuildRun {
	if run == nil {
		return nil
	}
	data, _ := json.Marshal(run)
	var copy PortingBuildRun
	_ = json.Unmarshal(data, &copy)
	return &copy
}

func (a *App) savePortingRun(run *PortingBuildRun) error {
	if run == nil {
		return errors.New("nil repair session")
	}
	manifest, err := a.sessionManifestPath(run.ID)
	if err != nil {
		return err
	}
	run.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	run.Revision++
	data, err := json.MarshalIndent(run, "", "  ")
	if err != nil {
		return err
	}
	if err := writeFileAtomic(manifest, data, 0o600); err != nil {
		return err
	}
	a.portingMu.Lock()
	a.portingRuns[run.ID] = clonePortingRun(run)
	a.portingMu.Unlock()
	return nil
}

func (a *App) loadPortingRun(id string) (*PortingBuildRun, error) {
	if !validRepairID(id) {
		return nil, errRepairSessionNotFound
	}
	a.portingMu.RLock()
	cached := clonePortingRun(a.portingRuns[id])
	a.portingMu.RUnlock()
	if cached != nil {
		return cached, nil
	}
	manifest, err := a.sessionManifestPath(id)
	if err != nil {
		return nil, errRepairSessionNotFound
	}
	data, err := os.ReadFile(manifest)
	if errors.Is(err, os.ErrNotExist) {
		return nil, errRepairSessionNotFound
	}
	if err != nil {
		return nil, err
	}
	var run PortingBuildRun
	if err := json.Unmarshal(data, &run); err != nil {
		return nil, fmt.Errorf("read repair session: %w", err)
	}
	if run.SchemaVersion != repairLabSchemaVersion || run.ID != id {
		return nil, errors.New("repair session schema or identity mismatch")
	}
	a.portingMu.Lock()
	a.portingRuns[id] = clonePortingRun(&run)
	a.portingMu.Unlock()
	return &run, nil
}

func (a *App) mutatePortingRun(id string, mutate func(*PortingBuildRun) error) (*PortingBuildRun, error) {
	a.portingMu.Lock()
	run := clonePortingRun(a.portingRuns[id])
	a.portingMu.Unlock()
	if run == nil {
		loaded, err := a.loadPortingRun(id)
		if err != nil {
			return nil, err
		}
		run = loaded
	}
	if err := mutate(run); err != nil {
		return nil, err
	}
	if err := a.savePortingRun(run); err != nil {
		return nil, err
	}
	return clonePortingRun(run), nil
}

func (a *App) listPortingRuns() ([]RepairSessionSummary, error) {
	root := filepath.Join(a.repairLabRoot(), "sessions")
	entries, err := os.ReadDir(root)
	if errors.Is(err, os.ErrNotExist) {
		return []RepairSessionSummary{}, nil
	}
	if err != nil {
		return nil, err
	}
	out := make([]RepairSessionSummary, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() || !validRepairID(entry.Name()) {
			continue
		}
		run, err := a.loadPortingRun(entry.Name())
		if err != nil {
			continue
		}
		target := ""
		if run.Target != nil {
			target = strings.TrimSpace(run.Target.GameVersion + " / " + run.Target.Loader)
		}
		out = append(out, RepairSessionSummary{
			ID: run.ID, Name: run.Name, State: run.State, Phase: run.Phase, UpdatedAt: run.UpdatedAt,
			Source: run.Source.Filename, ProjectRoot: run.Project.ProjectRoot, Loader: run.Project.Loader,
			GameVersion: run.Project.GameVersion, Target: target, LastError: run.LastError,
		})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].UpdatedAt > out[j].UpdatedAt })
	return out, nil
}

func relativeSessionPath(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil || strings.HasPrefix(rel, "..") {
		return ""
	}
	return filepath.ToSlash(rel)
}
