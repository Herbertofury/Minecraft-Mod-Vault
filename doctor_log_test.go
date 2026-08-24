package main

import (
	"archive/zip"
	"encoding/binary"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDoctorLogMatchesKnownRepairBrainSignature(t *testing.T) {
	report, err := doctorAnalyzeLog(DoctorLogRequest{
		GameVersion: "1.20.1",
		Loader:      "forge",
		Text: `java.lang.NoClassDefFoundError: com/github/L_Ender/cataclysm/client/render/entity/Ancient_Remnant_Rework_Renderer
Caused by: java.lang.ClassNotFoundException: com.github.L_Ender.cataclysm.client.render.entity.Ancient_Remnant_Rework_Renderer`,
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.Confidence != "high" {
		t.Fatalf("confidence=%q want high", report.Confidence)
	}
	if report.RootCause.ID != "known-cataclysm-renderer-api-drift" {
		t.Fatalf("root=%q findings=%+v", report.RootCause.ID, report.Findings)
	}
	if len(report.RepairPatterns) == 0 || report.RepairPatterns[0].ID != "binary-owner-descriptor-drift" {
		t.Fatalf("repair patterns=%+v", report.RepairPatterns)
	}
}

func TestDoctorLogEndpointRejectsEmptyAndClassifiesMixin(t *testing.T) {
	app := &App{settings: Settings{GameVersion: "1.21.1", Loader: "fabric"}}
	empty := httptest.NewRequest(http.MethodPost, "/api/doctor/log", strings.NewReader(`{"text":""}`))
	empty.Header.Set("Content-Type", "application/json")
	emptyRec := httptest.NewRecorder()
	app.handleDoctorLog(emptyRec, empty)
	if emptyRec.Code != http.StatusBadRequest {
		t.Fatalf("empty status=%d body=%s", emptyRec.Code, emptyRec.Body.String())
	}

	request := httptest.NewRequest(http.MethodPost, "/api/doctor/log", strings.NewReader(`{"text":"org.spongepowered.asm.mixin.throwables.MixinApplyError: Mixin [alpha.mixins.json:AlphaMixin] failed injection check, (0/1) succeeded. Critical injection failure"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	app.handleDoctorLog(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var report DoctorLogReport
	if err := json.Unmarshal(recorder.Body.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.RootCause.ID != "mixin-application" {
		t.Fatalf("root=%q findings=%+v", report.RootCause.ID, report.Findings)
	}
}

func writeGraphTestJar(t *testing.T, path, modID string, metadata map[string]any, className string, classMarker byte) {
	t.Helper()
	if metadata == nil {
		metadata = map[string]any{}
	}
	metadata["schemaVersion"] = 1
	metadata["id"] = modID
	metadata["name"] = modID
	metadata["version"] = "1"
	b, err := json.Marshal(metadata)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	meta, err := zw.Create("fabric.mod.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := meta.Write(b); err != nil {
		t.Fatal(err)
	}
	class, err := zw.Create(className)
	if err != nil {
		t.Fatal(err)
	}
	classBytes := make([]byte, 9)
	binary.BigEndian.PutUint32(classBytes[:4], 0xCAFEBABE)
	binary.BigEndian.PutUint16(classBytes[6:8], 61)
	classBytes[8] = classMarker
	if _, err := class.Write(classBytes); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestDoctorGraphDetectsDuplicateIDsClassesDependenciesAndConflicts(t *testing.T) {
	root := t.TempDir()
	mods := filepath.Join(root, "mods")
	if err := os.MkdirAll(mods, 0o755); err != nil {
		t.Fatal(err)
	}
	metadata := map[string]any{
		"depends": map[string]any{"minecraft": "1.20.1", "beta": ">=2"},
		"breaks":  map[string]any{"gamma": "*"},
	}
	writeGraphTestJar(t, filepath.Join(mods, "alpha-a.jar"), "alpha", metadata, "com/example/Shared.class", 1)
	writeGraphTestJar(t, filepath.Join(mods, "alpha-b.jar"), "alpha", metadata, "com/example/Shared.class", 1)
	writeGraphTestJar(t, filepath.Join(mods, "gamma.jar"), "gamma", nil, "com/example/Gamma.class", 2)

	app := &App{settings: Settings{JavaRoot: root, GameVersion: "1.20.1", Loader: "fabric"}}
	scan, err := app.buildDoctorScan(DoctorScanRequest{SourceGameVersion: "1.20.1", SourceLoader: "fabric", TargetGameVersion: "1.20.1", TargetLoader: "fabric"})
	if err != nil {
		t.Fatal(err)
	}
	if scan.Summary.DuplicateModIDs != 1 {
		t.Fatalf("duplicate IDs=%d want 1", scan.Summary.DuplicateModIDs)
	}
	if scan.Summary.ExactDuplicateClasses != 1 {
		t.Fatalf("exact duplicate classes=%d want 1", scan.Summary.ExactDuplicateClasses)
	}
	if scan.Summary.MissingRequiredDependencies != 2 {
		t.Fatalf("missing required=%d want 2", scan.Summary.MissingRequiredDependencies)
	}
	if scan.Summary.DeclaredConflicts != 2 {
		t.Fatalf("declared conflicts=%d want 2", scan.Summary.DeclaredConflicts)
	}
	globalIDs := doctorFindingIDs(scan.GlobalFindings)
	if !globalIDs["duplicate-mod-id"] || !globalIDs["duplicate-classes"] {
		t.Fatalf("global findings=%v", globalIDs)
	}
	for _, report := range scan.Mods {
		if strings.HasPrefix(report.Local.Filename, "alpha") {
			ids := doctorFindingIDs(report.Findings)
			if !ids["missing-required-dependency"] || !ids["declared-conflict"] {
				t.Fatalf("%s findings=%v", report.Local.Filename, ids)
			}
		}
	}
}

func TestDoctorGraphBuildsDependencyOrderAndDetectsCycles(t *testing.T) {
	root := t.TempDir()
	mods := filepath.Join(root, "mods")
	if err := os.MkdirAll(mods, 0o755); err != nil {
		t.Fatal(err)
	}
	writeGraphTestJar(t, filepath.Join(mods, "core.jar"), "core", nil, "com/example/Core.class", 1)
	writeGraphTestJar(t, filepath.Join(mods, "addon.jar"), "addon", map[string]any{
		"depends": map[string]any{"minecraft": "1.20.1", "core": ">=1"},
	}, "com/example/Addon.class", 2)
	writeGraphTestJar(t, filepath.Join(mods, "cycle-a.jar"), "cycle-a", map[string]any{
		"depends": map[string]any{"minecraft": "1.20.1", "cycle-b": "*"},
	}, "com/example/CycleA.class", 3)
	writeGraphTestJar(t, filepath.Join(mods, "cycle-b.jar"), "cycle-b", map[string]any{
		"depends": map[string]any{"minecraft": "1.20.1", "cycle-a": "*"},
	}, "com/example/CycleB.class", 4)
	writeGraphTestJar(t, filepath.Join(mods, "needs-api.jar"), "needs-api", map[string]any{
		"depends": map[string]any{"minecraft": "1.20.1", "fabric-api": "*"},
	}, "com/example/NeedsAPI.class", 5)

	app := &App{settings: Settings{JavaRoot: root, GameVersion: "1.20.1", Loader: "fabric"}}
	scan, err := app.buildDoctorScan(DoctorScanRequest{SourceGameVersion: "1.20.1", SourceLoader: "fabric", TargetGameVersion: "1.20.1", TargetLoader: "fabric"})
	if err != nil {
		t.Fatal(err)
	}
	if scan.Summary.DependencyCycles != 1 || len(scan.DependencyCycles) != 1 {
		t.Fatalf("cycles summary=%d cycles=%v", scan.Summary.DependencyCycles, scan.DependencyCycles)
	}
	if got := strings.Join(scan.DependencyCycles[0], ","); got != "cycle-a,cycle-b" {
		t.Fatalf("cycle=%q want cycle-a,cycle-b", got)
	}
	if !doctorFindingIDs(scan.GlobalFindings)["required-dependency-cycle"] {
		t.Fatalf("global findings=%v", doctorFindingIDs(scan.GlobalFindings))
	}
	positions := map[string]int{}
	for i, id := range scan.DependencyOrder {
		positions[id] = i
	}
	if positions["core"] >= positions["addon"] {
		t.Fatalf("dependency order=%v, core must precede addon", scan.DependencyOrder)
	}
	if scan.Summary.MissingRequiredDependencies != 1 {
		t.Fatalf("missing required=%d want fabric-api to be reported", scan.Summary.MissingRequiredDependencies)
	}
	foundFabricAPI := false
	for _, report := range scan.Mods {
		if report.Local.Metadata.ModID == "needs-api" && doctorFindingIDs(report.Findings)["missing-required-dependency"] {
			foundFabricAPI = true
		}
	}
	if !foundFabricAPI {
		t.Fatal("fabric-api missing dependency was hidden")
	}
	if len(scan.DependencyEdges) < 4 {
		t.Fatalf("dependency edges=%v", scan.DependencyEdges)
	}
}
