package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func (a *App) handleLibraryAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request LibraryActionRequest
	if err := decodeJSON(r, &request); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	request.Action = strings.ToLower(strings.TrimSpace(request.Action))
	paths := uniqueStringsPreserve(request.Paths)
	if len(request.IDs) > 0 {
		response, err := a.scanOmniLibrary(r.Context(), false, false)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
			return
		}
		wanted := map[string]bool{}
		for _, id := range request.IDs {
			wanted[id] = true
		}
		for _, item := range response.Items {
			if wanted[item.ID] {
				paths = append(paths, item.Path)
			}
		}
		paths = uniqueStringsPreserve(paths)
	}
	if request.Action == "update" {
		a.handleLibraryUpdateAction(w, r, paths)
		return
	}
	if len(paths) == 0 {
		writeJSON(w, http.StatusBadRequest, APIError{Error: "at least one managed item is required"})
		return
	}
	result := LibraryActionResult{OK: true, Action: request.Action}
	for _, rawPath := range paths {
		entry := LibraryActionEntry{Path: rawPath}
		path := filepath.Clean(rawPath)
		if !a.allowedLibraryPath(path) {
			entry.Error = "path is outside configured Minecraft and Vault roots"
			result.OK = false
			result.Results = append(result.Results, entry)
			continue
		}
		var receipt LibraryTransaction
		var err error
		switch request.Action {
		case "toggle":
			receipt, err = a.toggleLibraryPath(path)
		case "disable":
			receipt, err = a.disableLibraryPath(path)
		case "trash", "quarantine":
			receipt, err = a.trashLibraryPath(path, request.Action)
		default:
			err = fmt.Errorf("unsupported library action %q", request.Action)
		}
		if err != nil {
			entry.Error = err.Error()
			result.OK = false
		} else {
			entry.OK = true
			entry.Result = firstString(receipt.TargetPaths)
			entry.ReceiptID = receipt.ID
			result.Receipts = append(result.Receipts, receipt.ID)
		}
		result.Results = append(result.Results, entry)
	}
	status := http.StatusOK
	if !result.OK && len(result.Results) == 1 {
		status = http.StatusBadRequest
	}
	writeJSON(w, status, result)
}

func (a *App) handleLibraryUpdateAction(w http.ResponseWriter, r *http.Request, paths []string) {
	if len(paths) == 0 {
		writeJSON(w, http.StatusBadRequest, APIError{Error: "select at least one Java mod to update"})
		return
	}
	selected := map[string]bool{}
	for _, path := range paths {
		if !a.allowedLibraryPath(path) || !strings.HasSuffix(strings.TrimSuffix(strings.ToLower(path), ".disabled"), ".jar") {
			writeJSON(w, http.StatusBadRequest, APIError{Error: "updates are limited to managed Java JAR files"})
			return
		}
		selected[filepath.Base(path)] = true
	}
	a.mu.RLock()
	game, loader := a.settings.GameVersion, a.settings.Loader
	a.mu.RUnlock()
	plan, err := a.buildUpdatePlan(r.Context(), game, loader)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, APIError{Error: err.Error()})
		return
	}
	for name := range selected {
		found := false
		for _, item := range plan.Items {
			if item.Local.Filename == name && item.SafeUpdate != nil {
				found = true
				break
			}
		}
		if !found {
			writeJSON(w, http.StatusConflict, APIError{Error: name + " does not currently have an exact, safe update; review its details instead"})
			return
		}
	}
	response, err := a.applyUpdatePlan(r.Context(), plan, selected)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (a *App) allowedLibraryPath(path string) bool {
	if a.allowedManagedPath(path) {
		return true
	}
	roots := []string{filepath.Join(a.cfgDir, "library-disabled"), filepath.Join(a.cfgDir, "trash", "omnimanager"), filepath.Join(a.cfgDir, "library-originals")}
	for _, profile := range a.libraryProfiles() {
		if profile.Edition == "bedrock" && profile.Root != "" {
			roots = append(roots, profile.Root)
		}
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	for _, root := range roots {
		absRoot, rootErr := filepath.Abs(root)
		if rootErr != nil {
			continue
		}
		rel, relErr := filepath.Rel(absRoot, absPath)
		if relErr == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
			return true
		}
	}
	return false
}

func (a *App) toggleLibraryPath(path string) (LibraryTransaction, error) {
	info, err := os.Stat(path)
	if err != nil {
		return LibraryTransaction{}, err
	}
	if info.IsDir() || a.isBedrockPath(path) {
		return a.disableLibraryPath(path)
	}
	next := ""
	if strings.HasSuffix(strings.ToLower(path), ".disabled") {
		next = strings.TrimSuffix(path, ".disabled")
	} else {
		next = path + ".disabled"
	}
	if pathExists(next) {
		return LibraryTransaction{}, errors.New("toggle destination already exists")
	}
	if err := os.Rename(path, next); err != nil {
		return LibraryTransaction{}, err
	}
	receipt := LibraryTransaction{
		SchemaVersion: librarySchemaVersion, ID: randomToken(12), Action: "toggle", CreatedAt: time.Now().UTC().Format(time.RFC3339),
		SourcePaths: []string{path}, TargetPaths: []string{next}, ItemNames: []string{filepath.Base(path)},
		Metadata: map[string]any{"enabled": strings.HasSuffix(strings.ToLower(path), ".disabled")},
	}
	if err := a.writeLibraryTransaction(receipt); err != nil {
		_ = os.Rename(next, path)
		return LibraryTransaction{}, err
	}
	return receipt, nil
}

func (a *App) disableLibraryPath(path string) (LibraryTransaction, error) {
	if _, err := os.Stat(path); err != nil {
		return LibraryTransaction{}, err
	}
	if !a.isBedrockPath(path) {
		return a.toggleLibraryPath(path)
	}
	disabledRoot := filepath.Join(a.cfgDir, "library-disabled")
	if err := os.MkdirAll(disabledRoot, 0o755); err != nil {
		return LibraryTransaction{}, err
	}
	dst := uniquePath(filepath.Join(disabledRoot, randomToken(5)+"-"+filepath.Base(path)))
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		dst = uniqueDirectoryPath(dst)
	}
	if err := os.Rename(path, dst); err != nil {
		if err := copyThenRemove(path, dst); err != nil {
			return LibraryTransaction{}, err
		}
	}
	receipt := LibraryTransaction{
		SchemaVersion: librarySchemaVersion, ID: randomToken(12), Action: "disable", CreatedAt: time.Now().UTC().Format(time.RFC3339),
		SourcePaths: []string{path}, TargetPaths: []string{dst}, ItemNames: []string{filepath.Base(path)},
		Metadata: map[string]any{"edition": "bedrock", "reason": "moved outside com.mojang so Minecraft cannot load it"},
	}
	if err := a.writeLibraryTransaction(receipt); err != nil {
		_ = os.Rename(dst, path)
		return LibraryTransaction{}, err
	}
	return receipt, nil
}

func (a *App) trashLibraryPath(path, action string) (LibraryTransaction, error) {
	if _, err := os.Stat(path); err != nil {
		return LibraryTransaction{}, err
	}
	root := filepath.Join(a.cfgDir, "trash", "omnimanager")
	if err := os.MkdirAll(root, 0o755); err != nil {
		return LibraryTransaction{}, err
	}
	dst := filepath.Join(root, time.Now().UTC().Format("20060102T150405.000000000Z")+"-"+randomToken(4)+"-"+filepath.Base(path))
	if err := os.Rename(path, dst); err != nil {
		if err := copyThenRemove(path, dst); err != nil {
			return LibraryTransaction{}, err
		}
	}
	receipt := LibraryTransaction{
		SchemaVersion: librarySchemaVersion, ID: randomToken(12), Action: action, CreatedAt: time.Now().UTC().Format(time.RFC3339),
		SourcePaths: []string{path}, TargetPaths: []string{dst}, ItemNames: []string{filepath.Base(path)},
		Metadata: map[string]any{"recoverable": true},
	}
	if err := a.writeLibraryTransaction(receipt); err != nil {
		_ = os.Rename(dst, path)
		return LibraryTransaction{}, err
	}
	return receipt, nil
}

func (a *App) isBedrockPath(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	for _, profile := range a.libraryProfiles() {
		if profile.Edition != "bedrock" || profile.Root == "" {
			continue
		}
		root, _ := filepath.Abs(profile.Root)
		rel, relErr := filepath.Rel(root, absPath)
		if relErr == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
			return true
		}
	}
	return false
}

func (a *App) transactionDir() string {
	return filepath.Join(a.cfgDir, "library-transactions")
}

func (a *App) writeLibraryTransaction(transaction LibraryTransaction) error {
	if transaction.ID == "" {
		transaction.ID = randomToken(12)
	}
	if transaction.SchemaVersion == 0 {
		transaction.SchemaVersion = librarySchemaVersion
	}
	if transaction.CreatedAt == "" {
		transaction.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	if err := os.MkdirAll(a.transactionDir(), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(transaction, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(a.transactionDir(), safeFilename(transaction.ID)+".json")
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func (a *App) readLibraryTransaction(id string) (LibraryTransaction, error) {
	id = strings.TrimSuffix(safeFilename(id), ".json")
	data, err := os.ReadFile(filepath.Join(a.transactionDir(), id+".json"))
	if err != nil {
		return LibraryTransaction{}, err
	}
	var transaction LibraryTransaction
	if err := json.Unmarshal(data, &transaction); err != nil {
		return LibraryTransaction{}, err
	}
	if transaction.ID != id {
		return LibraryTransaction{}, errors.New("transaction identity mismatch")
	}
	return transaction, nil
}

func (a *App) handleLibraryHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	entries, err := os.ReadDir(a.transactionDir())
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
		return
	}
	transactions := []LibraryTransaction{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".json") {
			continue
		}
		data, readErr := os.ReadFile(filepath.Join(a.transactionDir(), entry.Name()))
		if readErr != nil {
			continue
		}
		var transaction LibraryTransaction
		if json.Unmarshal(data, &transaction) == nil {
			transactions = append(transactions, transaction)
		}
	}
	sort.SliceStable(transactions, func(i, j int) bool { return transactions[i].CreatedAt > transactions[j].CreatedAt })
	if len(transactions) > 100 {
		transactions = transactions[:100]
	}
	writeJSON(w, http.StatusOK, LibraryHistoryResponse{Transactions: transactions})
}

type libraryUndoRequest struct {
	ReceiptID string `json:"receiptId"`
}

func (a *App) handleLibraryUndo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request libraryUndoRequest
	if err := decodeJSON(r, &request); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	transaction, err := a.undoLibraryTransaction(request.ReceiptID)
	if err != nil {
		writeJSON(w, http.StatusConflict, APIError{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "transaction": transaction})
}

func (a *App) undoLibraryTransaction(id string) (LibraryTransaction, error) {
	transaction, err := a.readLibraryTransaction(id)
	if err != nil {
		return LibraryTransaction{}, err
	}
	if transaction.UndoneAt != "" {
		return LibraryTransaction{}, errors.New("transaction has already been undone")
	}
	switch transaction.Action {
	case "toggle", "disable", "trash", "quarantine":
		if len(transaction.SourcePaths) != len(transaction.TargetPaths) {
			return LibraryTransaction{}, errors.New("transaction path ledger is incomplete")
		}
		for i := range transaction.SourcePaths {
			src, dst := transaction.TargetPaths[i], transaction.SourcePaths[i]
			if !pathExists(src) {
				return LibraryTransaction{}, fmt.Errorf("recovery source is missing: %s", src)
			}
			if pathExists(dst) {
				return LibraryTransaction{}, fmt.Errorf("restore destination already exists: %s", dst)
			}
			if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
				return LibraryTransaction{}, err
			}
			if err := os.Rename(src, dst); err != nil {
				if err := copyThenRemove(src, dst); err != nil {
					return LibraryTransaction{}, err
				}
			}
		}
	case "bedrock-install":
		for _, path := range transaction.TargetPaths {
			if !a.allowedLibraryPath(path) {
				return LibraryTransaction{}, errors.New("install receipt points outside managed roots")
			}
			if pathExists(path) {
				trash := filepath.Join(a.cfgDir, "trash", "omnimanager", "undo-"+transaction.ID+"-"+filepath.Base(path))
				if err := os.MkdirAll(filepath.Dir(trash), 0o755); err != nil {
					return LibraryTransaction{}, err
				}
				if err := os.Rename(path, trash); err != nil {
					if err := copyThenRemove(path, trash); err != nil {
						return LibraryTransaction{}, err
					}
				}
			}
		}
	case "bedrock-activate":
		if len(transaction.SourcePaths) != 1 {
			return LibraryTransaction{}, errors.New("activation receipt is incomplete")
		}
		encoded := stringFromAny(transaction.Metadata["beforeBase64"])
		before, decodeErr := base64.StdEncoding.DecodeString(encoded)
		if decodeErr != nil {
			return LibraryTransaction{}, decodeErr
		}
		path := transaction.SourcePaths[0]
		if len(before) == 0 {
			if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
				return LibraryTransaction{}, err
			}
		} else if err := os.WriteFile(path, before, 0o644); err != nil {
			return LibraryTransaction{}, err
		}
	default:
		return LibraryTransaction{}, fmt.Errorf("transaction action %q is not reversible", transaction.Action)
	}
	transaction.UndoneAt = time.Now().UTC().Format(time.RFC3339)
	if err := a.writeLibraryTransaction(transaction); err != nil {
		return LibraryTransaction{}, err
	}
	return transaction, nil
}

func (a *App) scanDisabledLibraryItems() []LibraryItem {
	entries, err := os.ReadDir(a.transactionDir())
	if err != nil {
		return nil
	}
	items := []LibraryItem{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".json") {
			continue
		}
		transaction, readErr := a.readLibraryTransaction(strings.TrimSuffix(entry.Name(), ".json"))
		if readErr != nil || transaction.UndoneAt != "" || transaction.Action != "disable" || len(transaction.TargetPaths) != 1 || !pathExists(transaction.TargetPaths[0]) {
			continue
		}
		path := transaction.TargetPaths[0]
		info, statErr := os.Stat(path)
		if statErr != nil {
			continue
		}
		item := LibraryItem{
			ID: "disabled:" + transaction.ID, Path: path, Filename: filepath.Base(transaction.SourcePaths[0]),
			Name:    firstNonEmpty(firstString(transaction.ItemNames), humanizeMinecraftFilename(filepath.Base(path))),
			Edition: firstNonEmpty(stringFromAny(transaction.Metadata["edition"]), "bedrock"), Kind: firstNonEmpty(stringFromAny(transaction.Metadata["kind"]), "addon"),
			Profile: stringFromAny(transaction.Metadata["profile"]), Enabled: false, IsDir: info.IsDir(), Size: info.Size(),
			Modified: info.ModTime().UTC().Format(time.RFC3339), ManagedRoot: filepath.Dir(path), UpdateStatus: "current",
			UpdateMessage: "Disabled through a reversible Vault transaction. Use History > Undo to restore it.", ReceiptID: transaction.ID,
			ProvenanceConfidence: 1, MatchEvidence: []string{"Reversible Vault disable receipt " + transaction.ID},
		}
		items = append(items, item)
	}
	return items
}
