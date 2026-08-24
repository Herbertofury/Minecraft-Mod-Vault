package main

import (
	"context"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (a *App) getTestGrid() (*TestGrid, error) {
	a.testGridMu.Lock()
	defer a.testGridMu.Unlock()
	if a.testGrid != nil {
		return a.testGrid, nil
	}
	grid, err := newTestGrid(filepath.Join(a.cfgDir, "testgrid"))
	if err != nil {
		return nil, err
	}
	a.testGrid = grid
	return grid, nil
}

func (a *App) shutdownTestGrid() {
	a.testGridMu.Lock()
	grid := a.testGrid
	a.testGridMu.Unlock()
	if grid != nil {
		grid.CancelAll()
	}
}

func (a *App) handleTestGridCapabilities(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, APIError{Error: "GET required"})
		return
	}
	writeJSON(w, http.StatusOK, testGridCapabilities())
}

func (a *App) handleTestGridRuns(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, APIError{Error: "GET required"})
		return
	}
	grid, err := a.getTestGrid()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
		return
	}
	if id := strings.TrimSpace(r.URL.Query().Get("id")); id != "" {
		run, ok := grid.Get(id)
		if !ok {
			writeJSON(w, http.StatusNotFound, APIError{Error: "TestGrid run not found"})
			return
		}
		writeJSON(w, http.StatusOK, run)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"runs": grid.List()})
}

func (a *App) handleTestGridRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, APIError{Error: "POST required"})
		return
	}
	var manifest TestGridManifest
	if err := decodeJSON(r, &manifest); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	grid, err := a.getTestGrid()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
		return
	}
	run, err := grid.Start(context.Background(), manifest)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusAccepted, run)
}

func (a *App) handleTestGridValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, APIError{Error: "POST required"})
		return
	}
	var manifest TestGridManifest
	if err := decodeJSON(r, &manifest); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	manifest = normalizeTestGridManifest(manifest)
	if err := validateTestGridManifest(manifest); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"valid": true, "schemaVersion": manifest.SchemaVersion, "manifest": redactTestGridManifest(manifest)})
}

func (a *App) handleTestGridCancel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, APIError{Error: "POST required"})
		return
	}
	var request struct {
		ID string `json:"id"`
	}
	if err := decodeJSON(r, &request); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	grid, err := a.getTestGrid()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
		return
	}
	if !grid.Cancel(strings.TrimSpace(request.ID)) {
		writeJSON(w, http.StatusNotFound, APIError{Error: "active TestGrid run not found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "id": request.ID})
}

func (a *App) handleTestGridFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, APIError{Error: "GET required"})
		return
	}
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	kind := strings.TrimSpace(r.URL.Query().Get("kind"))
	grid, err := a.getTestGrid()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
		return
	}
	run, ok := grid.Get(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, APIError{Error: "TestGrid run not found"})
		return
	}
	path := ""
	switch kind {
	case "report":
		path = run.ReportPath
	case "junit":
		path = run.JUnitPath
	case "html":
		path = run.HTMLPath
	case "log":
		path = run.LogPath
	case "artifact":
		name := strings.TrimSpace(r.URL.Query().Get("name"))
		for _, artifact := range run.Artifacts {
			if name == artifact.Name || name == filepath.Base(artifact.Path) {
				path = artifact.Path
				break
			}
		}
	default:
		writeJSON(w, http.StatusBadRequest, APIError{Error: "kind must be report, junit, html, log, or artifact"})
		return
	}
	if path == "" {
		writeJSON(w, http.StatusNotFound, APIError{Error: "TestGrid file not found"})
		return
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	runRoot, err := filepath.Abs(run.RunDirectory)
	if err != nil || !pathWithinTestGridRoot(runRoot, absPath) {
		writeJSON(w, http.StatusForbidden, APIError{Error: "file is outside the TestGrid run directory"})
		return
	}
	info, err := os.Stat(absPath)
	if err != nil || info.IsDir() {
		writeJSON(w, http.StatusNotFound, APIError{Error: "TestGrid file not found"})
		return
	}
	if contentType := mime.TypeByExtension(filepath.Ext(absPath)); contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filepath.Base(absPath)))
	http.ServeFile(w, r, absPath)
}

func pathWithinTestGridRoot(root, target string) bool {
	relative, err := filepath.Rel(filepath.Clean(root), filepath.Clean(target))
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) && !filepath.IsAbs(relative)
}
