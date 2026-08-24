package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVersionAtlasContainsOfficialHistoryAndPins(t *testing.T) {
	atlas, err := loadVersionAtlas()
	if err != nil {
		t.Fatal(err)
	}
	if atlas.SchemaVersion != 1 || len(atlas.Versions) < 900 {
		t.Fatalf("atlas schema=%d versions=%d", atlas.SchemaVersion, len(atlas.Versions))
	}
	if atlas.Summary.LatestRelease == "" || atlas.Summary.LatestSnapshot == "" {
		t.Fatalf("atlas latest pointers missing: %+v", atlas.Summary)
	}
	row, _, ok := findAtlasVersion(atlas, "1.20.1")
	if !ok || javaMajor(row) != 17 || !row.ClientMappings || !row.ServerMappings {
		t.Fatalf("1.20.1 record incomplete: %+v", row)
	}
	if got := atlas.Forge.LatestByMinecraft["1.20.1"]; got != "1.20.1-47.4.23" {
		t.Fatalf("Forge 1.20.1 pin=%q", got)
	}
	if got := atlas.NeoForge.LatestByMinecraft["1.21.1"]; got != "21.1.248" {
		t.Fatalf("NeoForge 1.21.1 pin=%q", got)
	}
	if len(atlas.SourceEvidence) < 20 {
		t.Fatalf("source evidence=%d want broad immutable inputs", len(atlas.SourceEvidence))
	}
	for _, evidence := range atlas.SourceEvidence {
		if evidence.Name == "" || evidence.Bytes <= 0 || len(evidence.SHA256) != 64 {
			t.Fatalf("invalid source evidence: %+v", evidence)
		}
	}
}

func TestPortingAtlasEndpointSearchAndEvidence(t *testing.T) {
	app := &App{}
	req := httptest.NewRequest(http.MethodGet, "/api/porting/atlas?q=1.20.1&type=release&limit=10&evidence=1", nil)
	rec := httptest.NewRecorder()
	app.handlePortingAtlas(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var response VersionAtlasResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.TotalMatches == 0 || len(response.Versions) == 0 || response.Versions[0].ID != "1.20.1" {
		t.Fatalf("unexpected search response: %+v", response)
	}
	if len(response.SourceEvidence) == 0 {
		t.Fatal("evidence=1 did not return immutable source evidence")
	}
}

func TestBuildPortingPlanIncludesLiveInputForensics(t *testing.T) {
	root := t.TempDir()
	mods := filepath.Join(root, "mods")
	if err := os.MkdirAll(mods, 0o755); err != nil {
		t.Fatal(err)
	}
	jar := filepath.Join(mods, "forensic-input.jar")
	writeDoctorTestJar(t, jar, map[string][]byte{
		"fabric.mod.json": []byte(`{
  "schemaVersion":1,
  "id":"forensic_input",
  "name":"Forensic Input",
  "version":"2.4.0",
  "contact":{"sources":"https://github.com/example/forensic-input","homepage":"https://example.invalid/forensic"},
  "depends":{"minecraft":"1.20.1","fabricloader":">=0.15"},
  "mixins":["forensic.mixins.json"],
  "accessWidener":"forensic.accesswidener"
}`),
		"forensic.mixins.json":            []byte(`{"required":true,"package":"example.mixin","refmap":"forensic.refmap.json"}`),
		"forensic.accesswidener":          []byte("accessWidener v2 named\n"),
		"META-INF/FORENSIC.SF":            []byte("Signature-Version: 1.0\n"),
		"META-INF/jars/embedded.jar":      []byte("nested"),
		"natives/libforensic.so":          []byte("native"),
		"data/forensic/recipes/test.json": []byte(`{"type":"minecraft:crafting_shapeless"}`),
		"assets/forensic/lang/en_us.json": []byte(`{"forensic.name":"Forensic"}`),
		"dev/example/Forensic.class": doctorClassBytes(61,
			"net/fabricmc/fabric/api/", "org/spongepowered/asm/mixin/", "java/lang/reflect", "sun/misc/Unsafe", "net/minecraft/client/"),
	})

	app := &App{cfgDir: t.TempDir(), settings: Settings{JavaRoot: root, GameVersion: "1.20.1", Loader: "fabric"}}
	plan, err := app.buildPortingPlan(PortingPlanRequest{
		SourceGameVersion: "1.20.1", SourceLoader: "fabric",
		TargetGameVersion: "1.21.1", TargetLoader: "neoforge",
		SourceMode: "auto", InputJar: jar, ProjectName: "Forensic Port",
	})
	if err != nil {
		t.Fatal(err)
	}
	if plan.SourceMode != "binary" || plan.Direction != "upgrade" || plan.Risk != "critical" {
		t.Fatalf("plan mode=%s direction=%s risk=%s", plan.SourceMode, plan.Direction, plan.Risk)
	}
	if plan.InputAnalysis == nil {
		t.Fatal("input analysis missing")
	}
	input := plan.InputAnalysis
	if input.Metadata.ModID != "forensic_input" || !containsFold(input.DetectedLoaders, "fabric") || input.MaxJava != 17 {
		t.Fatalf("input analysis incomplete: %+v", input)
	}
	if len(input.SHA256) != 64 || len(input.SHA512) != 128 || input.ClassCount != 1 {
		t.Fatalf("input identity incomplete: %+v", input)
	}
	if len(input.MixinConfigs) == 0 || len(input.AccessWideners) == 0 || len(input.SignatureFiles) == 0 || len(input.NativeLibraries) == 0 || len(input.NestedJars) == 0 {
		t.Fatalf("expected forensic signals: %+v", input)
	}
	if !input.UsesUnsafe || !input.UsesReflection || !input.HasClientReferences {
		t.Fatalf("expected bytecode risk signals: %+v", input)
	}
	toolIDs := map[string]bool{}
	for _, tool := range plan.Tools {
		toolIDs[tool.ID] = true
	}
	for _, id := range []string{"intermed", "modcrawl", "vineflower", "cfr", "tiny-remapper", "retromod", "moddevgradle", "modstitch", "japicmp", "packwiz", "ferium"} {
		if !toolIDs[id] {
			t.Fatalf("missing selected tool %q: %v", id, toolIDs)
		}
	}
	joinedWarnings := strings.Join(plan.Warnings, "\n")
	for _, expected := range []string{"signed", "Native libraries", "Reflection", "Cross-loader"} {
		if !strings.Contains(strings.ToLower(joinedWarnings), strings.ToLower(expected)) {
			t.Fatalf("missing warning %q in %s", expected, joinedWarnings)
		}
	}
	if !strings.Contains(plan.Boundaries[0], input.SHA256) {
		t.Fatalf("input boundary does not carry exact identity: %q", plan.Boundaries[0])
	}
	if len(plan.Phases) != 8 || len(plan.VerificationMatrix) < 8 || len(plan.CompletionDefinition) < 5 {
		t.Fatalf("plan is not fully gated: phases=%d verification=%d completion=%d", len(plan.Phases), len(plan.VerificationMatrix), len(plan.CompletionDefinition))
	}
}

func TestPortingWorkspaceIsIsolatedHashedAndHashLocked(t *testing.T) {
	root := t.TempDir()
	mods := filepath.Join(root, "mods")
	if err := os.MkdirAll(mods, 0o755); err != nil {
		t.Fatal(err)
	}
	jar := filepath.Join(mods, "workspace-input.jar")
	writeDoctorTestJar(t, jar, map[string][]byte{
		"fabric.mod.json":         []byte(`{"schemaVersion":1,"id":"workspace_input","name":"Workspace Input","version":"1.0.0","depends":{"minecraft":"1.20.1"}}`),
		"dev/example/Input.class": doctorClassBytes(61, "net/fabricmc/loader/"),
	})
	original, err := os.ReadFile(jar)
	if err != nil {
		t.Fatal(err)
	}
	app := &App{cfgDir: t.TempDir(), settings: Settings{JavaRoot: root, GameVersion: "1.20.1", Loader: "fabric"}}
	plan, err := app.buildPortingPlan(PortingPlanRequest{SourceGameVersion: "1.20.1", SourceLoader: "fabric", TargetGameVersion: "1.21.1", TargetLoader: "neoforge", SourceMode: "auto", InputJar: jar, ProjectName: "Workspace Input"})
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := app.createPortingWorkspace(plan)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Input == nil || manifest.Input.SHA256 != plan.InputAnalysis.SHA256 {
		t.Fatalf("workspace input identity mismatch: manifest=%+v plan=%+v", manifest.Input, plan.InputAnalysis)
	}
	if manifest.Status != "generated-awaiting-source-and-runtime-verification" || len(manifest.Files) < 12 {
		t.Fatalf("workspace manifest incomplete: %+v", manifest)
	}
	for _, name := range []string{"PORTING-PLAN.json", "PORTING-PLAN.md", "README.md", "settings.gradle.kts", "build.gradle.kts", "scripts/verify.sh", "scripts/verify.ps1", "WORKSPACE-MANIFEST.json"} {
		if _, err := os.Stat(filepath.Join(manifest.Path, filepath.FromSlash(name))); err != nil {
			t.Fatalf("missing workspace file %s: %v", name, err)
		}
	}
	build, err := os.ReadFile(filepath.Join(manifest.Path, "build.gradle.kts"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(build, []byte(`id("net.neoforged.moddev")`)) || !bytes.Contains(build, []byte(`version = "21.1.248"`)) {
		t.Fatalf("NeoForge workspace not pinned correctly:\n%s", build)
	}
	after, err := os.ReadFile(jar)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(original, after) {
		t.Fatal("workspace generation mutated the installed input JAR")
	}
	copied, err := os.ReadFile(manifest.Input.WorkspacePath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(original, copied) {
		t.Fatal("workspace input copy differs from original")
	}

	writeDoctorTestJar(t, jar, map[string][]byte{"fabric.mod.json": []byte(`{"schemaVersion":1,"id":"changed"}`)})
	if _, err := app.createPortingWorkspace(plan); err == nil || !strings.Contains(err.Error(), "changed since planning") {
		t.Fatalf("expected hash-lock failure after input drift, got %v", err)
	}
}

func TestPortingWorkspaceHandlerRejectsInputSwapAfterPlanning(t *testing.T) {
	root := t.TempDir()
	mods := filepath.Join(root, "mods")
	if err := os.MkdirAll(mods, 0o755); err != nil {
		t.Fatal(err)
	}
	first := filepath.Join(mods, "first.jar")
	second := filepath.Join(mods, "second.jar")
	writeDoctorTestJar(t, first, map[string][]byte{"fabric.mod.json": []byte(`{"schemaVersion":1,"id":"first"}`)})
	writeDoctorTestJar(t, second, map[string][]byte{"fabric.mod.json": []byte(`{"schemaVersion":1,"id":"second"}`)})
	app := &App{cfgDir: t.TempDir(), settings: Settings{JavaRoot: root, GameVersion: "1.20.1", Loader: "fabric"}, portingPlans: map[string]PortingPlan{}}
	plan, err := app.buildPortingPlan(PortingPlanRequest{SourceGameVersion: "1.20.1", SourceLoader: "fabric", TargetGameVersion: "1.21.1", TargetLoader: "fabric", SourceMode: "auto", InputJar: first})
	if err != nil {
		t.Fatal(err)
	}
	app.portingPlans[plan.ID] = plan
	body, _ := json.Marshal(PortingWorkspaceRequest{PlanID: plan.ID, InputJar: second})
	req := httptest.NewRequest(http.MethodPost, "/api/porting/workspace", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.handlePortingWorkspace(rec, req)
	if rec.Code != http.StatusConflict || !strings.Contains(rec.Body.String(), "changed after planning") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestParseJavaMajorHandlesLegacyAndModernOutput(t *testing.T) {
	cases := map[string]int{
		`java version "1.8.0_402"`:       8,
		`openjdk version "17.0.12" 2024`: 17,
		`openjdk 25.0.1 2026-01-01`:      25,
	}
	for input, want := range cases {
		if got := parseJavaMajor(input); got != want {
			t.Fatalf("parseJavaMajor(%q)=%d want %d", input, got, want)
		}
	}
}
