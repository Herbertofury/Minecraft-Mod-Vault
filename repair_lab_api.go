package main

import (
	"context"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func (a *App) handleRepairLabStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if err := os.MkdirAll(filepath.Join(a.repairLabRoot(), "sessions"), 0o755); err != nil {
		writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
		return
	}
	sessions, err := a.listPortingRuns()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
		return
	}
	status := RepairLabStatus{
		SchemaVersion: repairLabSchemaVersion, Ready: true, Root: a.repairLabRoot(), Sessions: sessions,
		Tools: repairHostTools(), ExecutionPhrase: repairLabConfirmationPhrase, SecurityBoundaries: repairRunSecurityBoundary(),
		Capabilities: []RepairCapabilitySummary{
			{ID: "immutable-intake", Name: "Immutable source intake", Description: "ZIP-slip, link, duplicate-path, entry-count, extracted-size, and suspicious-compression checks with SHA-256 source identity.", State: "ready"},
			{ID: "version-atlas", Name: "Offline Version Atlas", Description: "Exact Mojang, mcmeta, Fabric, Quilt, Modrinth, Forge, NeoForge, mappings, and build-plugin resolution from embedded reviewed seeds.", State: "ready"},
			{ID: "recognized-migration", Name: "Recognized migration edits", Description: "Conservative version, loader, mapping, Java, metadata, and pack-format changes in a disposable working copy with per-field evidence.", State: "ready"},
			{ID: "controlled-builds", Name: "Controlled wrapper execution", Description: "Only detected project wrappers and fixed build/test/clean actions, with explicit acknowledgement, dedicated caches, sanitized environment, logs, cancellation, and timeout.", State: "ready"},
			{ID: "artifact-proof", Name: "Artifact proof", Description: "SHA-256, JAR metadata/class inspection, immutable-source re-verification, receipts, deterministic prepared-source ZIPs, and repair bundles.", State: "ready"},
		},
	}
	if atlas, err := loadRepairVersionAtlas(); err == nil {
		copy := atlas.Summary
		status.Atlas = &copy
	}
	if a.brain != nil {
		if brain, err := a.brain.Status(); err == nil {
			status.Brain = &brain
		} else {
			status.BrainError = err.Error()
		}
	} else {
		status.BrainError = firstNonEmpty(a.brainInitError, "compatibility brain unavailable")
	}
	writeJSON(w, http.StatusOK, status)
}

func repairHostTools() []RepairHostTool {
	tools := []struct {
		ID, Name, Binary, Role string
		Args                   []string
	}{
		{"java", "Java runtime", "java", "Execute Gradle/Maven and Minecraft-compatible build tools.", []string{"-version"}},
		{"javac", "Java compiler", "javac", "Compile Java source when the project build requests it.", []string{"-version"}},
		{"git", "Git", "git", "Capture source lineage and support build scripts that read Git metadata.", []string{"--version"}},
	}
	out := make([]RepairHostTool, 0, len(tools))
	for _, tool := range tools {
		entry := RepairHostTool{ID: tool.ID, Name: tool.Name, Role: tool.Role}
		path, err := exec.LookPath(tool.Binary)
		if err == nil {
			entry.Path = path
			entry.Ready = true
			ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
			output, _ := exec.CommandContext(ctx, path, tool.Args...).CombinedOutput()
			cancel()
			entry.Version = firstLine(strings.TrimSpace(string(output)))
		}
		out = append(out, entry)
	}
	return out
}

func firstLine(value string) string {
	if index := strings.IndexByte(value, '\n'); index >= 0 {
		value = value[:index]
	}
	return strings.TrimSpace(strings.TrimSuffix(value, "\r"))
}

func (a *App) handleRepairLabImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, repairLabMaxUploadBytes+(32<<20))
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: "parse source upload: " + err.Error()})
		return
	}
	file, header, err := r.FormFile("source")
	if err != nil {
		file, header, err = r.FormFile("files")
	}
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: "a source ZIP is required"})
		return
	}
	defer file.Close()
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".zip") {
		writeJSON(w, http.StatusBadRequest, APIError{Error: "Repair Lab source intake currently requires a .zip archive"})
		return
	}
	run, err := a.newPortingRun(strings.TrimSuffix(filepath.Base(header.Filename), filepath.Ext(header.Filename)))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
		return
	}
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.RemoveAll(run.Paths.Root)
		}
	}()
	digest, size, err := copyAndHashLimited(run.Paths.OriginalArchive, file, repairLabMaxUploadBytes)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	extracted, err := safeExtractSourceZip(run.Paths.OriginalArchive, run.Paths.ImmutableSource)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	treeDigest, files, extractedBytes, err := hashDirectoryTree(run.Paths.ImmutableSource)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
		return
	}
	if files != extracted.FileCount || extractedBytes != extracted.ExtractedBytes {
		writeJSON(w, http.StatusInternalServerError, APIError{Error: "extracted source identity did not remain stable during intake"})
		return
	}
	run.Source = RepairSourceSnapshot{
		Filename: filepath.Base(header.Filename), Size: size, SHA256: digest, TreeSHA256: treeDigest,
		FileCount: files, ExtractedBytes: extractedBytes, ImportedAt: time.Now().UTC().Format(time.RFC3339),
	}
	if err := resetWorkingCopy(run); err != nil {
		writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
		return
	}
	profile, err := detectRepairProject(run.Paths.ImmutableSource)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: "detect source project: " + err.Error()})
		return
	}
	run.Project = profile
	run.State = "imported"
	run.Phase = "source-profiled"
	run.Warnings = uniqueStringsPreserve(append(run.Warnings, profile.Warnings...))
	if extracted.RootHint != "" && profile.ProjectRoot != "" && !strings.HasPrefix(profile.ProjectRoot, extracted.RootHint) {
		run.Warnings = append(run.Warnings, "The archive has a single top-level directory, but the build root was detected elsewhere; verify the selected project root before execution.")
	}
	if err := writeRepairReceipt(run); err != nil {
		writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
		return
	}
	if err := a.savePortingRun(run); err != nil {
		writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
		return
	}
	cleanup = false
	writeJSON(w, http.StatusCreated, run)
}

func (a *App) handleRepairLabSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	run, err := a.loadPortingRun(id)
	if err != nil {
		writeRepairSessionError(w, err)
		return
	}
	if len(run.Runs) > 0 {
		last := len(run.Runs) - 1
		run.Runs[last].LogTail = tailTextFile(run.Runs[last].LogFile, 96<<10)
	}
	writeJSON(w, http.StatusOK, run)
}

func (a *App) handleRepairLabPrepare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request RepairPrepareRequest
	if err := decodeJSON(r, &request); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	run, err := a.loadPortingRun(strings.TrimSpace(request.SessionID))
	if err != nil {
		writeRepairSessionError(w, err)
		return
	}
	if err := prepareRepairSession(run, request); err != nil {
		writeRepairSessionError(w, err)
		return
	}
	if err := a.savePortingRun(run); err != nil {
		writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, run)
}

func (a *App) handleRepairLabRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request RepairRunRequest
	if err := decodeJSON(r, &request); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	run, err := a.loadPortingRun(strings.TrimSpace(request.SessionID))
	if err != nil {
		writeRepairSessionError(w, err)
		return
	}
	updated, err := a.startRepairCommand(run, request)
	if err != nil {
		writeRepairSessionError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, updated)
}

func (a *App) handleRepairLabCancel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request RepairSessionRequest
	if err := decodeJSON(r, &request); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	if err := a.cancelRepairCommand(strings.TrimSpace(request.SessionID)); err != nil {
		writeRepairSessionError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{"ok": true, "sessionId": request.SessionID})
}

func (a *App) handleRepairLabReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request RepairSessionRequest
	if err := decodeJSON(r, &request); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	run, err := a.loadPortingRun(strings.TrimSpace(request.SessionID))
	if err != nil {
		writeRepairSessionError(w, err)
		return
	}
	if run.State == "running" {
		writeRepairSessionError(w, errRepairSessionBusy)
		return
	}
	if err := verifyImmutableSource(run); err != nil {
		writeRepairSessionError(w, err)
		return
	}
	if err := resetWorkingCopy(run); err != nil {
		writeRepairSessionError(w, err)
		return
	}
	profile, err := detectRepairProject(run.Paths.ImmutableSource)
	if err != nil {
		writeRepairSessionError(w, err)
		return
	}
	run.Project = profile
	run.Target = nil
	run.Resolution = nil
	run.Changes = nil
	run.Artifacts = nil
	run.State = "imported"
	run.Phase = "rolled-back-to-immutable-source"
	run.LastError = ""
	run.Warnings = uniqueStringsPreserve(profile.Warnings)
	if err := writeRepairReceipt(run); err != nil {
		writeRepairSessionError(w, err)
		return
	}
	if err := a.savePortingRun(run); err != nil {
		writeRepairSessionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, run)
}

func (a *App) handleRepairLabExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request RepairExportRequest
	if err := decodeJSON(r, &request); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	run, err := a.loadPortingRun(strings.TrimSpace(request.SessionID))
	if err != nil {
		writeRepairSessionError(w, err)
		return
	}
	exports, err := createRepairExports(run, request.IncludeArtifacts)
	if err != nil {
		writeRepairSessionError(w, err)
		return
	}
	run.Exports = exports
	run.Phase = "exports-ready"
	if err := writeRepairReceipt(run); err != nil {
		writeRepairSessionError(w, err)
		return
	}
	if err := a.savePortingRun(run); err != nil {
		writeRepairSessionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, run)
}

func (a *App) handleRepairLabLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	run, err := a.loadPortingRun(strings.TrimSpace(r.URL.Query().Get("id")))
	if err != nil {
		writeRepairSessionError(w, err)
		return
	}
	commandID := strings.TrimSpace(r.URL.Query().Get("command"))
	var selected *RepairCommandRun
	for i := len(run.Runs) - 1; i >= 0; i-- {
		if commandID == "" || run.Runs[i].ID == commandID {
			copy := run.Runs[i]
			selected = &copy
			break
		}
	}
	if selected == nil {
		writeJSON(w, http.StatusNotFound, APIError{Error: "no build log exists for this session"})
		return
	}
	selected.LogTail = tailTextFile(selected.LogFile, 256<<10)
	writeJSON(w, http.StatusOK, selected)
}

func (a *App) handleRepairLabDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	run, err := a.loadPortingRun(strings.TrimSpace(r.URL.Query().Get("id")))
	if err != nil {
		writeRepairSessionError(w, err)
		return
	}
	kind := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("kind")))
	index, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("index")))
	path := ""
	name := ""
	switch kind {
	case "artifact":
		if index < 0 || index >= len(run.Artifacts) {
			writeJSON(w, http.StatusNotFound, APIError{Error: "artifact index is unavailable"})
			return
		}
		projectRoot := filepath.Join(run.Paths.WorkingCopy, filepath.FromSlash(run.Project.ProjectRoot))
		path = filepath.Join(projectRoot, filepath.FromSlash(run.Artifacts[index].RelativePath))
		name = run.Artifacts[index].Name
	case "export":
		if index < 0 || index >= len(run.Exports) {
			writeJSON(w, http.StatusNotFound, APIError{Error: "export index is unavailable"})
			return
		}
		path = filepath.Join(run.Paths.Root, filepath.FromSlash(run.Exports[index].RelativePath))
		name = run.Exports[index].Name
	case "receipt-json":
		path, name = run.Paths.ReceiptJSON, filepath.Base(run.Paths.ReceiptJSON)
	case "receipt-markdown":
		path, name = run.Paths.ReceiptMarkdown, filepath.Base(run.Paths.ReceiptMarkdown)
	case "log":
		if len(run.Runs) == 0 {
			writeJSON(w, http.StatusNotFound, APIError{Error: "no build log exists"})
			return
		}
		path, name = run.Runs[len(run.Runs)-1].LogFile, filepath.Base(run.Runs[len(run.Runs)-1].LogFile)
	default:
		writeJSON(w, http.StatusBadRequest, APIError{Error: "kind must be artifact, export, receipt-json, receipt-markdown, or log"})
		return
	}
	if !pathContainedBy(run.Paths.Root, path) {
		writeJSON(w, http.StatusForbidden, APIError{Error: "download path escaped the Repair Lab session"})
		return
	}
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		writeJSON(w, http.StatusNotFound, APIError{Error: "requested download is unavailable"})
		return
	}
	if contentType := mime.TypeByExtension(strings.ToLower(filepath.Ext(name))); contentType != "" {
		w.Header().Set("Content-Type", contentType)
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", safeFilename(name)))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeFile(w, r, path)
}

func writeRepairSessionError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, errRepairSessionNotFound):
		writeJSON(w, http.StatusNotFound, APIError{Error: err.Error()})
	case errors.Is(err, errRepairSessionBusy):
		writeJSON(w, http.StatusConflict, APIError{Error: err.Error()})
	default:
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
	}
}
