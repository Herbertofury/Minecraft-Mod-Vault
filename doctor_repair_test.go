package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func duplicateTestJar(t *testing.T, path, id string) {
	t.Helper()
	writeDoctorTestJar(t, path, map[string][]byte{
		"fabric.mod.json":             []byte(`{"schemaVersion":1,"id":"` + id + `","version":"1.0.0"}`),
		"dev/example/Duplicate.class": doctorClassBytes(61, "net/fabricmc/loader/"),
	})
}

func TestDoctorRepairQuarantinesAndRestoresExactDuplicates(t *testing.T) {
	root := t.TempDir()
	mods := filepath.Join(root, "mods")
	if err := os.MkdirAll(mods, 0o755); err != nil {
		t.Fatal(err)
	}
	keeper := filepath.Join(mods, "alpha.jar")
	duplicate := filepath.Join(mods, "alpha-copy.jar.disabled")
	unrelated := filepath.Join(mods, "other.jar")
	duplicateTestJar(t, keeper, "alpha")
	bytes, err := os.ReadFile(keeper)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(duplicate, bytes, 0o644); err != nil {
		t.Fatal(err)
	}
	duplicateTestJar(t, unrelated, "other")

	app := &App{cfgDir: t.TempDir(), settings: Settings{JavaRoot: root}, doctorRepairPlans: map[string]DoctorRepairPlan{}}
	plan, err := app.buildDoctorRepairPlan()
	if err != nil {
		t.Fatal(err)
	}
	if plan.SafeActionCount != 1 || len(plan.Actions) != 1 {
		t.Fatalf("plan=%+v", plan)
	}
	action := plan.Actions[0]
	if action.Keep.Path != keeper || len(action.Quarantine) != 1 || action.Quarantine[0].Path != duplicate {
		t.Fatalf("deterministic keeper/quarantine incorrect: %+v", action)
	}
	receipt, err := app.applyDoctorRepairPlan(plan, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(receipt.Moves) != 1 || receipt.Status != "applied-recoverable" {
		t.Fatalf("receipt=%+v", receipt)
	}
	if _, err := os.Stat(keeper); err != nil {
		t.Fatalf("keeper missing: %v", err)
	}
	if _, err := os.Stat(duplicate); !os.IsNotExist(err) {
		t.Fatalf("duplicate still installed: %v", err)
	}
	for _, path := range []string{receipt.ReceiptPath, filepath.Join(receipt.QuarantineRoot, "RESTORE.md"), receipt.Moves[0].QuarantinePath} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("receipt artifact missing %s: %v", path, err)
		}
	}
	quarantinedHash, err := fileSHA512(receipt.Moves[0].QuarantinePath)
	if err != nil || !strings.EqualFold(quarantinedHash, action.SHA512) {
		t.Fatalf("quarantine hash=%s err=%v", quarantinedHash, err)
	}
	var reread DoctorRepairReceipt
	receiptBytes, err := os.ReadFile(receipt.ReceiptPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(receiptBytes, &reread); err != nil {
		t.Fatal(err)
	}
	if len(reread.Verification) < 5 {
		t.Fatalf("receipt verification too weak: %+v", reread.Verification)
	}

	result, err := app.restoreDoctorRepair(receipt.ReceiptPath)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "restored" || len(result.Moves) != 1 {
		t.Fatalf("restore result=%+v", result)
	}
	for _, path := range []string{keeper, duplicate, unrelated} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("restored mods missing %s: %v", path, err)
		}
	}
	if _, err := os.Stat(receipt.Moves[0].QuarantinePath); !os.IsNotExist(err) {
		t.Fatalf("quarantine move still exists after restore: %v", err)
	}
	restoredHash, err := fileSHA512(duplicate)
	if err != nil || !strings.EqualFold(restoredHash, action.SHA512) {
		t.Fatalf("restored hash=%s err=%v", restoredHash, err)
	}
}

func TestDoctorRepairAbortsWithoutMutationWhenCandidateDrifts(t *testing.T) {
	root := t.TempDir()
	mods := filepath.Join(root, "mods")
	if err := os.MkdirAll(mods, 0o755); err != nil {
		t.Fatal(err)
	}
	keeper := filepath.Join(mods, "alpha.jar")
	duplicate := filepath.Join(mods, "alpha-copy.jar")
	duplicateTestJar(t, keeper, "alpha")
	payload, err := os.ReadFile(keeper)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(duplicate, payload, 0o644); err != nil {
		t.Fatal(err)
	}
	app := &App{cfgDir: t.TempDir(), settings: Settings{JavaRoot: root}}
	plan, err := app.buildDoctorRepairPlan()
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Actions) != 1 {
		t.Fatalf("expected one action, got %+v", plan)
	}
	if err := os.WriteFile(duplicate, append(payload, 'x'), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := app.applyDoctorRepairPlan(plan, nil); err == nil || !strings.Contains(err.Error(), "candidate changed") {
		t.Fatalf("expected preflight drift rejection, got %v", err)
	}
	for _, path := range []string{keeper, duplicate} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("preflight failure mutated %s: %v", path, err)
		}
	}
	quarantineRoot := filepath.Join(app.cfgDir, "quarantine", "doctor")
	entries, err := os.ReadDir(quarantineRoot)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("preflight failure created quarantine artifacts: %+v", entries)
	}
}

func TestDoctorRepairSkipsSymbolicLinkEntries(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symbolic-link creation is permission-dependent on Windows")
	}
	root := t.TempDir()
	mods := filepath.Join(root, "mods")
	if err := os.MkdirAll(mods, 0o755); err != nil {
		t.Fatal(err)
	}
	real := filepath.Join(mods, "real.jar")
	link := filepath.Join(mods, "linked.jar")
	duplicateTestJar(t, real, "real")
	if err := os.Symlink(real, link); err != nil {
		t.Fatal(err)
	}
	app := &App{cfgDir: t.TempDir(), settings: Settings{JavaRoot: root}}
	plan, err := app.buildDoctorRepairPlan()
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Actions) != 0 {
		t.Fatalf("symbolic link must not become an automatic repair action: %+v", plan.Actions)
	}
	if !strings.Contains(strings.Join(plan.Warnings, "\n"), "symbolic-link") {
		t.Fatalf("symbolic-link skip was not reported: %+v", plan.Warnings)
	}
}

func TestDoctorRestoreRejectsReceiptOutsideQuarantine(t *testing.T) {
	app := &App{cfgDir: t.TempDir(), settings: Settings{JavaRoot: t.TempDir()}}
	outside := filepath.Join(t.TempDir(), "receipt.json")
	if err := os.WriteFile(outside, []byte(`{"schemaVersion":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := app.restoreDoctorRepair(outside); err == nil || !strings.Contains(err.Error(), "outside") {
		t.Fatalf("expected outside-receipt rejection, got %v", err)
	}
}
