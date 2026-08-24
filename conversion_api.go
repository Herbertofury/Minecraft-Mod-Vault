package main

import (
	"errors"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func (a *App) handleConversionStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	sessions, err := a.listConversionSessions()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
		return
	}
	status := ConversionStatus{
		SchemaVersion: conversionSchemaVersion,
		Ready:         true,
		GraphVersion:  conversionGraphVersion,
		Root:          a.conversionRoot(),
		Sessions:      sessions,
		Targets:       conversionTargetOptions(),
		Tools:         a.configuredConversionToolAdapters(),
		Safety:        conversionSafetyBoundary(),
		Capabilities: []ConversionCapability{
			{ID: "universal-content-graph", Name: "Universal Minecraft Content Graph", Description: "Normalizes Java and Bedrock identity, assets, data, logic, worlds, scripts, relationships and target support without discarding the original files.", State: "ready"},
			{ID: "java-bedrock", Name: "Java ↔ Bedrock conversion lanes", Description: "Emits real packs, add-ons, source projects, world/template packages and exact semantic review contracts from one persistent graph.", State: "ready"},
			{ID: "version-targeting", Name: "Version and loader targeting", Description: "Uses the embedded Version Atlas, current pack-format rules and target manifests to generate version-pinned Java and Bedrock outputs.", State: "ready"},
			{ID: "world-adapters", Name: "World and structure adapters", Description: "Detects Chunker, je2be, Amulet and SchemConvert and keeps cross-edition chunk work visibly incomplete until an adapter actually runs and validates.", State: "adapter-aware"},
			{ID: "proof-bundles", Name: "Deterministic conversion proof", Description: "Every session preserves source hashes, graph, plan, coverage, review queue, output hashes, validation and reproducible reports.", State: "ready"},
		},
	}
	writeJSON(w, http.StatusOK, status)
}

func (a *App) handleConversionImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, conversionMaxUploadBytes+(32<<20))
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: "parse conversion upload: " + err.Error()})
		return
	}
	file, header, err := r.FormFile("source")
	if err != nil {
		file, header, err = r.FormFile("files")
	}
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: "a Minecraft package, world, structure or project archive is required"})
		return
	}
	defer file.Close()
	if !isConversionArchiveExtension(strings.ToLower(filepath.Ext(header.Filename))) {
		writeJSON(w, http.StatusBadRequest, APIError{Error: "unsupported conversion input extension"})
		return
	}
	session, err := a.importConversionFile(header.Filename, file)
	if err != nil {
		writeConversionError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, session)
}

func (a *App) handleConversionImportPath(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request ConversionPathImportRequest
	if err := decodeJSON(r, &request); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	session, err := a.importConversionPath(request.Path)
	if err != nil {
		writeConversionError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, session)
}

func (a *App) handleConversionSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	session, err := a.loadConversionSession(r.URL.Query().Get("id"))
	if err != nil {
		writeConversionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, session)
}

func (a *App) handleConversionPlan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request ConversionPlanRequest
	if err := decodeJSON(r, &request); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	session, err := a.loadConversionSession(request.SessionID)
	if err != nil {
		writeConversionError(w, err)
		return
	}
	plan, err := buildConversionPlan(session, request.Target)
	if err != nil {
		writeConversionError(w, err)
		return
	}
	session.Plan = plan
	session.Outputs = nil
	session.State = "planned"
	session.Phase = "target-plan-ready"
	session.LastError = ""
	if err := a.saveConversionSession(session); err != nil {
		writeConversionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, session)
}

func (a *App) handleConversionRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request ConversionRunRequest
	if err := decodeJSON(r, &request); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	session, err := a.loadConversionSession(request.SessionID)
	if err != nil {
		writeConversionError(w, err)
		return
	}
	session.State = "converting"
	session.Phase = "emitting-target"
	session.LastError = ""
	_ = a.saveConversionSession(session)
	if err := executeConversion(session); err != nil {
		session.State = "failed"
		session.Phase = "conversion-failed"
		session.LastError = err.Error()
		_ = a.saveConversionSession(session)
		writeConversionError(w, err)
		return
	}
	if err := a.saveConversionSession(session); err != nil {
		writeConversionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, session)
}

func (a *App) handleConversionReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request ConversionSessionRequest
	if err := decodeJSON(r, &request); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	session, err := a.loadConversionSession(request.SessionID)
	if err != nil {
		writeConversionError(w, err)
		return
	}
	if !pathContainedBy(a.conversionSessionsRoot(), session.Paths.Root) {
		writeJSON(w, http.StatusForbidden, APIError{Error: "conversion session path is outside the managed root"})
		return
	}
	if err := os.RemoveAll(session.Paths.Root); err != nil {
		writeConversionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "sessionId": request.SessionID})
}

func (a *App) handleConversionDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	session, err := a.loadConversionSession(r.URL.Query().Get("id"))
	if err != nil {
		writeConversionError(w, err)
		return
	}
	kind := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("kind")))
	index, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("index")))
	path, name := "", ""
	switch kind {
	case "output", "":
		if index < 0 || index >= len(session.Outputs) {
			writeJSON(w, http.StatusNotFound, APIError{Error: "conversion output index is unavailable"})
			return
		}
		path = filepath.Join(session.Paths.Root, filepath.FromSlash(session.Outputs[index].RelativePath))
		name = session.Outputs[index].Name
	case "report":
		path, name = session.Paths.ReportMarkdown, filepath.Base(session.Paths.ReportMarkdown)
	case "graph":
		path, name = session.Paths.GraphJSON, filepath.Base(session.Paths.GraphJSON)
	case "plan":
		path, name = session.Paths.PlanJSON, filepath.Base(session.Paths.PlanJSON)
	case "receipt":
		path, name = session.Paths.ReceiptJSON, filepath.Base(session.Paths.ReceiptJSON)
	case "adapter-log":
		if index < 0 || index >= len(session.AdapterRuns) {
			writeJSON(w, http.StatusNotFound, APIError{Error: "adapter log index is unavailable"})
			return
		}
		path, name = session.AdapterRuns[index].LogPath, filepath.Base(session.AdapterRuns[index].LogPath)
	default:
		writeJSON(w, http.StatusBadRequest, APIError{Error: "kind must be output, report, graph, plan, receipt, or adapter-log"})
		return
	}
	if !pathContainedBy(session.Paths.Root, path) {
		writeJSON(w, http.StatusForbidden, APIError{Error: "download path escaped the conversion session"})
		return
	}
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		writeJSON(w, http.StatusNotFound, APIError{Error: "requested conversion download is unavailable"})
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

func writeConversionError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, errConversionSessionNotFound):
		writeJSON(w, http.StatusNotFound, APIError{Error: err.Error()})
	case errors.Is(err, os.ErrPermission):
		writeJSON(w, http.StatusForbidden, APIError{Error: err.Error()})
	default:
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
	}
}
