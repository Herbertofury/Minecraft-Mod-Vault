package main

import (
	"net/http"
	"strconv"
	"strings"
)

func (a *App) handleBrainStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if a.brain == nil {
		message := firstNonEmpty(a.brainInitError, "compatibility brain is unavailable")
		writeJSON(w, http.StatusServiceUnavailable, APIError{Error: message})
		return
	}
	status, err := a.brain.Status()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, status)
}

func (a *App) handleBrainSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if a.brain == nil {
		writeJSON(w, http.StatusServiceUnavailable, APIError{Error: firstNonEmpty(a.brainInitError, "compatibility brain is unavailable")})
		return
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	kind := strings.TrimSpace(r.URL.Query().Get("kind"))
	limit := 40
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = clampInt(parsed, 1, 100)
		}
	}
	results, err := a.brain.Search(query, kind, limit)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"query": query, "kind": kind, "total": len(results), "results": results})
}

func (a *App) handleBrainVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if a.brain == nil {
		writeJSON(w, http.StatusServiceUnavailable, APIError{Error: firstNonEmpty(a.brainInitError, "compatibility brain is unavailable")})
		return
	}
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	if id == "" {
		writeJSON(w, http.StatusBadRequest, APIError{Error: "id is required"})
		return
	}
	version, err := a.brain.MinecraftVersion(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
		return
	}
	if !version.Available {
		writeJSON(w, http.StatusNotFound, APIError{Error: "Minecraft version is not present in the compatibility brain"})
		return
	}
	writeJSON(w, http.StatusOK, version)
}
