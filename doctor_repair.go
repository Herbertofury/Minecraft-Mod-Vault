package main

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type DoctorRepairPlan struct {
	ID              string               `json:"id"`
	CreatedAt       string               `json:"createdAt"`
	ModsDirectory   string               `json:"modsDirectory"`
	Status          string               `json:"status"`
	Actions         []DoctorRepairAction `json:"actions"`
	SafeActionCount int                  `json:"safeActionCount"`
	TotalBytes      int64                `json:"totalBytes"`
	Warnings        []string             `json:"warnings"`
}

type DoctorRepairAction struct {
	ID             string             `json:"id"`
	Kind           string             `json:"kind"`
	Title          string             `json:"title"`
	SHA512         string             `json:"sha512"`
	Keep           DoctorRepairFile   `json:"keep"`
	Quarantine     []DoctorRepairFile `json:"quarantine"`
	Recoverable    bool               `json:"recoverable"`
	Confidence     string             `json:"confidence"`
	Rationale      string             `json:"rationale"`
	Verification   []string           `json:"verification"`
	RequiresReview bool               `json:"requiresReview"`
}

type DoctorRepairFile struct {
	Path     string `json:"path"`
	Filename string `json:"filename"`
	Enabled  bool   `json:"enabled"`
	Size     int64  `json:"size"`
}

type DoctorRepairApplyRequest struct {
	PlanID    string   `json:"planId"`
	ActionIDs []string `json:"actionIds,omitempty"`
}

type DoctorRepairReceipt struct {
	SchemaVersion  int                         `json:"schemaVersion"`
	ID             string                      `json:"id"`
	PlanID         string                      `json:"planId"`
	CreatedAt      string                      `json:"createdAt"`
	QuarantineRoot string                      `json:"quarantineRoot"`
	ReceiptPath    string                      `json:"receiptPath"`
	Status         string                      `json:"status"`
	Moves          []DoctorRepairReceiptMove   `json:"moves"`
	Verification   []string                    `json:"verification"`
	Restore        DoctorRepairRestoreContract `json:"restore"`
}

type DoctorRepairReceiptMove struct {
	ActionID       string `json:"actionId"`
	SHA512         string `json:"sha512"`
	OriginalPath   string `json:"originalPath"`
	QuarantinePath string `json:"quarantinePath"`
	Size           int64  `json:"size"`
}

type DoctorRepairRestoreContract struct {
	Endpoint    string `json:"endpoint"`
	ReceiptPath string `json:"receiptPath"`
	Rule        string `json:"rule"`
}

type DoctorRepairRestoreRequest struct {
	ReceiptPath string `json:"receiptPath"`
}

type DoctorRepairRestoreResult struct {
	ReceiptPath string                    `json:"receiptPath"`
	RestoredAt  string                    `json:"restoredAt"`
	Status      string                    `json:"status"`
	Moves       []DoctorRepairReceiptMove `json:"moves"`
}

func (a *App) handleDoctorRepairPlan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	plan, err := a.buildDoctorRepairPlan()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			writeJSON(w, http.StatusOK, plan)
			return
		}
		writeJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
		return
	}
	a.dataMu.Lock()
	if a.doctorRepairPlans == nil {
		a.doctorRepairPlans = map[string]DoctorRepairPlan{}
	}
	a.doctorRepairPlans[plan.ID] = plan
	for id, old := range a.doctorRepairPlans {
		created, _ := time.Parse(time.RFC3339, old.CreatedAt)
		if !created.IsZero() && time.Since(created) > 2*time.Hour {
			delete(a.doctorRepairPlans, id)
		}
	}
	a.dataMu.Unlock()
	writeJSON(w, http.StatusOK, plan)
}

func (a *App) buildDoctorRepairPlan() (DoctorRepairPlan, error) {
	modsDir := a.javaTargetDir("mods")
	plan := DoctorRepairPlan{
		ID: randomToken(12), CreatedAt: time.Now().UTC().Format(time.RFC3339), ModsDirectory: modsDir,
		Status: "planned-no-mutations", Warnings: []string{
			"Only byte-identical duplicate JARs are eligible for automatic quarantine in this release.",
			"Duplicate mod IDs with different bytes, conflicting classes, dependency changes, and version replacements remain review-only.",
			"Every file is re-hashed immediately before mutation; any drift aborts the whole operation.",
		},
	}
	locals, err := scanLocalModJars(modsDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return plan, nil
		}
		return plan, err
	}
	groups := map[string][]LocalModFile{}
	for _, local := range locals {
		info, lstatErr := os.Lstat(local.Path)
		if lstatErr != nil || info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
			plan.Warnings = append(plan.Warnings, "Skipped non-regular or symbolic-link entry: "+local.Filename)
			continue
		}
		hash := strings.ToLower(strings.TrimSpace(local.SHA512))
		if hash != "" {
			groups[hash] = append(groups[hash], local)
		}
	}
	hashes := make([]string, 0, len(groups))
	for hash, files := range groups {
		if len(files) > 1 {
			hashes = append(hashes, hash)
		}
	}
	sort.Strings(hashes)
	for _, hash := range hashes {
		files := append([]LocalModFile{}, groups[hash]...)
		sort.SliceStable(files, func(i, j int) bool {
			if files[i].Enabled != files[j].Enabled {
				return files[i].Enabled
			}
			if len(files[i].Filename) != len(files[j].Filename) {
				return len(files[i].Filename) < len(files[j].Filename)
			}
			return strings.ToLower(files[i].Filename) < strings.ToLower(files[j].Filename)
		})
		keep := repairFile(files[0])
		action := DoctorRepairAction{
			ID: "exact-duplicate-" + hash[:16], Kind: "quarantine-exact-duplicate", Title: "Quarantine byte-identical duplicate JARs",
			SHA512: hash, Keep: keep, Recoverable: true, Confidence: "cryptographic-exact",
			Rationale:    "All candidate files have the same SHA-512 digest. The deterministic keeper prefers an enabled copy, then the shortest stable filename. Other copies are moved to the Vault quarantine, never deleted.",
			Verification: []string{"Re-hash keeper and every candidate", "Move only non-keeper copies", "Write receipt with original and quarantine paths", "Re-scan mods directory after apply"},
		}
		for _, local := range files[1:] {
			action.Quarantine = append(action.Quarantine, repairFile(local))
			plan.TotalBytes += local.Size
		}
		plan.Actions = append(plan.Actions, action)
	}
	plan.SafeActionCount = len(plan.Actions)
	return plan, nil
}

func repairFile(local LocalModFile) DoctorRepairFile {
	return DoctorRepairFile{Path: local.Path, Filename: local.Filename, Enabled: local.Enabled, Size: local.Size}
}

func (a *App) handleDoctorRepairApply(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request DoctorRepairApplyRequest
	if err := decodeJSON(r, &request); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	a.dataMu.RLock()
	plan, ok := a.doctorRepairPlans[request.PlanID]
	a.dataMu.RUnlock()
	if !ok {
		writeJSON(w, http.StatusNotFound, APIError{Error: "repair plan expired; generate it again"})
		return
	}
	selected := map[string]bool{}
	for _, id := range request.ActionIDs {
		selected[strings.TrimSpace(id)] = true
	}
	receipt, err := a.applyDoctorRepairPlan(plan, selected)
	if err != nil {
		writeJSON(w, http.StatusConflict, APIError{Error: err.Error()})
		return
	}
	a.dataMu.Lock()
	delete(a.doctorRepairPlans, request.PlanID)
	a.dataMu.Unlock()
	writeJSON(w, http.StatusOK, receipt)
}

func (a *App) applyDoctorRepairPlan(plan DoctorRepairPlan, selected map[string]bool) (DoctorRepairReceipt, error) {
	actions := []DoctorRepairAction{}
	for _, action := range plan.Actions {
		if len(selected) == 0 || selected[action.ID] {
			actions = append(actions, action)
		}
	}
	if len(actions) == 0 {
		return DoctorRepairReceipt{}, errors.New("no repair actions selected")
	}
	// Validate the entire transaction before moving one byte.
	for _, action := range actions {
		if action.Kind != "quarantine-exact-duplicate" || len(action.Quarantine) == 0 {
			return DoctorRepairReceipt{}, fmt.Errorf("unsupported or empty action %s", action.ID)
		}
		if err := a.validateRepairFile(action.Keep, action.SHA512); err != nil {
			return DoctorRepairReceipt{}, fmt.Errorf("keeper changed for %s: %w", action.ID, err)
		}
		for _, file := range action.Quarantine {
			if file.Path == action.Keep.Path {
				return DoctorRepairReceipt{}, fmt.Errorf("action %s attempts to quarantine its keeper", action.ID)
			}
			if err := a.validateRepairFile(file, action.SHA512); err != nil {
				return DoctorRepairReceipt{}, fmt.Errorf("candidate changed for %s: %w", action.ID, err)
			}
		}
	}

	receiptID := time.Now().UTC().Format("20060102T150405Z") + "-" + randomToken(4)
	root := filepath.Join(a.cfgDir, "quarantine", "doctor", receiptID)
	if err := os.MkdirAll(root, 0o755); err != nil {
		return DoctorRepairReceipt{}, err
	}
	receipt := DoctorRepairReceipt{
		SchemaVersion: 1, ID: receiptID, PlanID: plan.ID, CreatedAt: time.Now().UTC().Format(time.RFC3339), QuarantineRoot: root,
		ReceiptPath: filepath.Join(root, "receipt.json"), Status: "applied-recoverable",
		Verification: []string{"All inputs passed immediate SHA-512 revalidation", "No installed JAR was deleted", "Every move has an original path, quarantine path, size, and digest"},
	}
	moved := []DoctorRepairReceiptMove{}
	rollback := func() error {
		var failures []string
		for index := len(moved) - 1; index >= 0; index-- {
			move := moved[index]
			if pathExists(move.OriginalPath) {
				failures = append(failures, "original path already occupied: "+move.OriginalPath)
				continue
			}
			if err := moveFileRecoverable(move.QuarantinePath, move.OriginalPath); err != nil {
				failures = append(failures, err.Error())
			}
		}
		if len(failures) > 0 {
			return errors.New(strings.Join(failures, "; "))
		}
		return nil
	}
	for _, action := range actions {
		actionDir := filepath.Join(root, safePortingSlug(action.ID))
		if err := os.MkdirAll(actionDir, 0o755); err != nil {
			_ = rollback()
			return DoctorRepairReceipt{}, err
		}
		for _, file := range action.Quarantine {
			destination := uniquePath(filepath.Join(actionDir, safeFilename(file.Filename)))
			if err := moveFileRecoverable(file.Path, destination); err != nil {
				rollbackErr := rollback()
				if rollbackErr != nil {
					return DoctorRepairReceipt{}, fmt.Errorf("repair failed: %v; rollback also failed: %v", err, rollbackErr)
				}
				return DoctorRepairReceipt{}, fmt.Errorf("repair failed and was rolled back: %w", err)
			}
			move := DoctorRepairReceiptMove{ActionID: action.ID, SHA512: action.SHA512, OriginalPath: file.Path, QuarantinePath: destination, Size: file.Size}
			moved = append(moved, move)
			digest, digestErr := fileSHA512(destination)
			if digestErr != nil || !strings.EqualFold(digest, action.SHA512) {
				rollbackErr := rollback()
				failure := fmt.Errorf("quarantine verification failed for %s", file.Filename)
				if digestErr != nil {
					failure = fmt.Errorf("quarantine verification failed for %s: %w", file.Filename, digestErr)
				}
				if rollbackErr != nil {
					return DoctorRepairReceipt{}, fmt.Errorf("%v; rollback also failed: %v", failure, rollbackErr)
				}
				_ = os.RemoveAll(root)
				return DoctorRepairReceipt{}, fmt.Errorf("%v; transaction rolled back", failure)
			}
		}
		if err := a.validateRepairFile(action.Keep, action.SHA512); err != nil {
			rollbackErr := rollback()
			if rollbackErr != nil {
				return DoctorRepairReceipt{}, fmt.Errorf("keeper verification failed after quarantine: %v; rollback also failed: %v", err, rollbackErr)
			}
			_ = os.RemoveAll(root)
			return DoctorRepairReceipt{}, fmt.Errorf("keeper verification failed after quarantine; transaction rolled back: %w", err)
		}
	}
	receipt.Moves = moved
	receipt.Restore = DoctorRepairRestoreContract{Endpoint: "/api/doctor/repair/restore", ReceiptPath: receipt.ReceiptPath, Rule: "Restore refuses to overwrite an occupied original path and revalidates every quarantined SHA-512 digest."}
	receipt.Verification = append(receipt.Verification,
		"Every quarantine destination passed post-move SHA-512 verification",
		"Every deterministic keeper passed post-transaction SHA-512 verification",
	)
	readmePath := filepath.Join(root, "RESTORE.md")
	readme := renderDoctorRepairReadme(receipt)
	if err := writeFileAtomic(readmePath, []byte(readme), 0o644); err != nil {
		rollbackErr := rollback()
		if rollbackErr != nil {
			return DoctorRepairReceipt{}, fmt.Errorf("could not write restore instructions: %v; rollback also failed: %v", err, rollbackErr)
		}
		_ = os.RemoveAll(root)
		return DoctorRepairReceipt{}, fmt.Errorf("could not write restore instructions; repair rolled back: %w", err)
	}
	if err := writeJSONFileAtomic(receipt.ReceiptPath, receipt); err != nil {
		rollbackErr := rollback()
		if rollbackErr != nil {
			return DoctorRepairReceipt{}, fmt.Errorf("could not write receipt: %v; rollback also failed: %v", err, rollbackErr)
		}
		_ = os.RemoveAll(root)
		return DoctorRepairReceipt{}, fmt.Errorf("could not write receipt; repair rolled back: %w", err)
	}
	return receipt, nil
}

func (a *App) validateRepairFile(file DoctorRepairFile, expectedSHA512 string) error {
	if !a.allowedManagedPath(file.Path) {
		return errors.New("path is outside managed Minecraft directories")
	}
	info, err := os.Lstat(file.Path)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return errors.New("symbolic links are not eligible for automatic repair")
	}
	if !info.Mode().IsRegular() {
		return errors.New("not a regular file")
	}
	if info.Size() != file.Size {
		return fmt.Errorf("size changed: expected %d, got %d", file.Size, info.Size())
	}
	digest, err := fileSHA512(file.Path)
	if err != nil {
		return err
	}
	if !strings.EqualFold(digest, expectedSHA512) {
		return fmt.Errorf("SHA-512 changed: expected %s, got %s", expectedSHA512, digest)
	}
	return nil
}

func fileSHA512(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha512.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func moveFileRecoverable(source, destination string) error {
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}
	if pathExists(destination) {
		return fmt.Errorf("destination already exists: %s", destination)
	}
	if err := os.Rename(source, destination); err == nil {
		return nil
	}
	return copyThenRemove(source, destination)
}

func renderDoctorRepairReadme(receipt DoctorRepairReceipt) string {
	var out strings.Builder
	fmt.Fprintf(&out, "# Minecraft Mod Vault Repair Receipt %s\n\n", receipt.ID)
	out.WriteString("This repair quarantined only byte-identical duplicate JARs. No file was deleted.\n\n")
	out.WriteString("## Restore through the Vault\n\n")
	fmt.Fprintf(&out, "POST `%s` with:\n\n```json\n{\"receiptPath\":%q}\n```\n\n", receipt.Restore.Endpoint, receipt.ReceiptPath)
	out.WriteString("Restore will re-hash each quarantined JAR and refuse to overwrite any occupied original path.\n\n## Moves\n\n")
	for _, move := range receipt.Moves {
		fmt.Fprintf(&out, "- `%s` -> `%s`  \n  SHA-512: `%s`\n", move.OriginalPath, move.QuarantinePath, move.SHA512)
	}
	return out.String()
}

func (a *App) handleDoctorRepairRestore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request DoctorRepairRestoreRequest
	if err := decodeJSON(r, &request); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		return
	}
	result, err := a.restoreDoctorRepair(request.ReceiptPath)
	if err != nil {
		writeJSON(w, http.StatusConflict, APIError{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (a *App) restoreDoctorRepair(receiptPath string) (DoctorRepairRestoreResult, error) {
	receiptPath = filepath.Clean(strings.TrimSpace(receiptPath))
	root := filepath.Join(a.cfgDir, "quarantine", "doctor")
	if !pathWithin(root, receiptPath) {
		return DoctorRepairRestoreResult{}, errors.New("receipt is outside the Doctor quarantine")
	}
	data, err := os.ReadFile(receiptPath)
	if err != nil {
		return DoctorRepairRestoreResult{}, err
	}
	var receipt DoctorRepairReceipt
	if err := json.Unmarshal(data, &receipt); err != nil {
		return DoctorRepairRestoreResult{}, err
	}
	if receipt.SchemaVersion != 1 || len(receipt.Moves) == 0 {
		return DoctorRepairRestoreResult{}, errors.New("receipt is empty or unsupported")
	}
	for _, move := range receipt.Moves {
		if !pathWithin(root, move.QuarantinePath) {
			return DoctorRepairRestoreResult{}, errors.New("receipt contains a quarantine path outside the Doctor quarantine")
		}
		if !a.allowedManagedPath(move.OriginalPath) {
			return DoctorRepairRestoreResult{}, errors.New("receipt contains an original path outside managed Minecraft directories")
		}
		if pathExists(move.OriginalPath) {
			return DoctorRepairRestoreResult{}, fmt.Errorf("restore refused because original path is occupied: %s", move.OriginalPath)
		}
		digest, digestErr := fileSHA512(move.QuarantinePath)
		if digestErr != nil {
			return DoctorRepairRestoreResult{}, digestErr
		}
		if !strings.EqualFold(digest, move.SHA512) {
			return DoctorRepairRestoreResult{}, fmt.Errorf("quarantined file hash changed: %s", move.QuarantinePath)
		}
	}
	restored := []DoctorRepairReceiptMove{}
	rollback := func() error {
		var failures []string
		for index := len(restored) - 1; index >= 0; index-- {
			move := restored[index]
			if err := moveFileRecoverable(move.OriginalPath, move.QuarantinePath); err != nil {
				failures = append(failures, err.Error())
			}
		}
		if len(failures) > 0 {
			return errors.New(strings.Join(failures, "; "))
		}
		return nil
	}
	for _, move := range receipt.Moves {
		if err := moveFileRecoverable(move.QuarantinePath, move.OriginalPath); err != nil {
			rollbackErr := rollback()
			if rollbackErr != nil {
				return DoctorRepairRestoreResult{}, fmt.Errorf("restore failed: %v; rollback also failed: %v", err, rollbackErr)
			}
			return DoctorRepairRestoreResult{}, fmt.Errorf("restore failed and was rolled back: %w", err)
		}
		restored = append(restored, move)
		digest, digestErr := fileSHA512(move.OriginalPath)
		if digestErr != nil || !strings.EqualFold(digest, move.SHA512) {
			rollbackErr := rollback()
			failure := fmt.Errorf("restored file verification failed: %s", move.OriginalPath)
			if digestErr != nil {
				failure = fmt.Errorf("restored file verification failed: %s: %w", move.OriginalPath, digestErr)
			}
			if rollbackErr != nil {
				return DoctorRepairRestoreResult{}, fmt.Errorf("%v; rollback also failed: %v", failure, rollbackErr)
			}
			return DoctorRepairRestoreResult{}, fmt.Errorf("%v; restore rolled back", failure)
		}
	}
	result := DoctorRepairRestoreResult{ReceiptPath: receiptPath, RestoredAt: time.Now().UTC().Format(time.RFC3339), Status: "restored", Moves: restored}
	_ = writeJSONFile(filepath.Join(filepath.Dir(receiptPath), "restore-result.json"), result)
	return result, nil
}

func writeJSONFileAtomic(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return writeFileAtomic(path, append(data, '\n'), 0o644)
}

func pathWithin(root, candidate string) bool {
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return false
	}
	absoluteCandidate, err := filepath.Abs(candidate)
	if err != nil {
		return false
	}
	relative, err := filepath.Rel(absoluteRoot, absoluteCandidate)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(os.PathSeparator))
}
