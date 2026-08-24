package main

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func writeDoctorTestJar(t *testing.T, path string, entries map[string][]byte) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	for name, data := range entries {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}

func doctorClassBytes(major uint16, markers ...string) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint32(b[:4], 0xCAFEBABE)
	binary.BigEndian.PutUint16(b[4:6], 0)
	binary.BigEndian.PutUint16(b[6:8], major)
	for _, marker := range markers {
		b = append(b, []byte(marker)...)
		b = append(b, 0)
	}
	return b
}

func doctorFindingIDs(findings []DoctorFinding) map[string]bool {
	out := map[string]bool{}
	for _, finding := range findings {
		out[finding.ID] = true
	}
	return out
}

func TestDoctorScanDetectsPortingRisks(t *testing.T) {
	root := t.TempDir()
	mods := filepath.Join(root, "mods")
	if err := os.MkdirAll(mods, 0o755); err != nil {
		t.Fatal(err)
	}
	fabric := []byte(`{
  "schemaVersion": 1,
  "id": "alpha",
  "name": "Alpha",
  "version": "1.0.0",
  "contact": {"sources": "https://github.com/example/alpha"},
  "depends": {"fabricloader": ">=0.15", "minecraft": "1.21.1", "beta": ">=2"},
  "breaks": {"gamma": "*"},
  "mixins": ["alpha.mixins.json"],
  "accessWidener": "alpha.accesswidener",
  "jars": [{"file":"META-INF/jars/lib.jar"}]
}`)
	writeDoctorTestJar(t, filepath.Join(mods, "alpha.jar"), map[string][]byte{
		"fabric.mod.json":              fabric,
		"alpha.mixins.json":            []byte(`{"required":true,"package":"example.mixin","plugin":"example.Plugin","refmap":"alpha.refmap.json","mixins":["AlphaMixin"]}`),
		"alpha.accesswidener":          []byte("accessWidener v2 named\naccessible class net/minecraft/Test\n"),
		"META-INF/ALPHA.SF":            []byte("Signature-Version: 1.0\n"),
		"META-INF/jars/lib.jar":        []byte("nested"),
		"natives/example.dll":          []byte("native"),
		"data/alpha/tags/test.json":    []byte(`{"values":[]}`),
		"assets/alpha/lang/en_us.json": []byte(`{"alpha.name":"Alpha"}`),
		"com/example/Alpha.class": doctorClassBytes(65,
			"org/spongepowered/asm/mixin/", "net/fabricmc/fabric/api/", "java/lang/reflect", "sun/misc/Unsafe", "net/minecraft/client/"),
	})

	app := &App{cfgDir: t.TempDir(), settings: Settings{JavaRoot: root, GameVersion: "1.21.1", Loader: "fabric"}}
	scan, err := app.buildDoctorScan(DoctorScanRequest{SourceGameVersion: "1.21.1", SourceLoader: "fabric", TargetGameVersion: "1.20.1", TargetLoader: "neoforge"})
	if err != nil {
		t.Fatal(err)
	}
	if scan.Summary.TotalMods != 1 {
		t.Fatalf("total mods=%d want 1", scan.Summary.TotalMods)
	}
	if scan.TargetJava != 17 {
		t.Fatalf("target java=%d want 17", scan.TargetJava)
	}
	if scan.KnowledgeReviewedAt != "2026-08-20" {
		t.Fatalf("knowledge reviewed=%q", scan.KnowledgeReviewedAt)
	}
	if len(scan.Mods) != 1 {
		t.Fatalf("mods=%d want 1", len(scan.Mods))
	}
	report := scan.Mods[0]
	if report.RiskLevel != "critical" {
		t.Fatalf("risk=%s score=%d want critical", report.RiskLevel, report.RiskScore)
	}
	if report.Signals.MaxJava != 21 {
		t.Fatalf("max java=%d want 21", report.Signals.MaxJava)
	}
	if !report.Signals.UsesReflection || !report.Signals.UsesUnsafe || !report.Signals.HasClientReferences {
		t.Fatalf("signals=%+v", report.Signals)
	}
	if len(report.Signals.MixinConfigs) == 0 || len(report.Signals.AccessWideners) == 0 || len(report.Signals.SignatureFiles) == 0 || len(report.Signals.NestedJars) == 0 || len(report.Signals.NativeLibraries) == 0 {
		t.Fatalf("missing archive signals: %+v", report.Signals)
	}
	ids := doctorFindingIDs(report.Findings)
	for _, id := range []string{"java-bytecode-too-new", "mixin-retarget-required", "mixin-plugin", "access-rules", "signed-jar", "native-code", "nested-jars", "unsafe-api", "loader-migration", "resource-migration"} {
		if !ids[id] {
			t.Fatalf("missing finding %q; ids=%v", id, ids)
		}
	}
	if len(report.Plan) < 10 {
		t.Fatalf("plan steps=%d want comprehensive plan", len(report.Plan))
	}
	if len(scan.RecommendedSourceIDs) < 15 {
		t.Fatalf("recommended source IDs=%d want broad recommendations", len(scan.RecommendedSourceIDs))
	}
}

func TestDoctorTargetJavaMatrix(t *testing.T) {
	cases := map[string]int{
		"1.12.2": 8,
		"1.16.5": 8,
		"1.17.1": 16,
		"1.18.2": 17,
		"1.20.1": 17,
		"1.20.4": 17,
		"1.20.5": 21,
		"1.21.1": 21,
		"26.1":   25,
		"26.2":   25,
	}
	for version, want := range cases {
		if got := targetJavaForMinecraft(version); got != want {
			t.Errorf("targetJavaForMinecraft(%q)=%d want %d", version, got, want)
		}
	}
}

func TestDoctorSourcesEndpoint(t *testing.T) {
	app := &App{}
	req := httptest.NewRequest(http.MethodGet, "/api/doctor/sources", nil)
	rec := httptest.NewRecorder()
	app.handleDoctorSources(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var knowledge DoctorKnowledge
	if err := json.Unmarshal(rec.Body.Bytes(), &knowledge); err != nil {
		t.Fatal(err)
	}
	if knowledge.Total < 100 || len(knowledge.Sources) < 100 {
		t.Fatalf("knowledge total=%d sources=%d", knowledge.Total, len(knowledge.Sources))
	}
	if len(knowledge.Categories) < 10 {
		t.Fatalf("categories=%d want broad catalog", len(knowledge.Categories))
	}

	req = httptest.NewRequest(http.MethodGet, "/api/doctor/sources?q=vineflower", nil)
	rec = httptest.NewRecorder()
	app.handleDoctorSources(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("filtered status=%d body=%s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &knowledge); err != nil {
		t.Fatal(err)
	}
	if knowledge.Total == 0 {
		t.Fatal("expected Vineflower search result")
	}
	found := false
	for _, source := range knowledge.Sources {
		if strings.Contains(strings.ToLower(source.Name), "vineflower") {
			found = true
		}
	}
	if !found {
		t.Fatalf("filtered sources=%v", knowledge.Sources)
	}
}

func TestDoctorScanHTTPValidationAndResponse(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "mods"), 0o755); err != nil {
		t.Fatal(err)
	}
	app := &App{settings: Settings{JavaRoot: root, GameVersion: "1.20.1", Loader: "forge"}}
	bad := httptest.NewRequest(http.MethodPost, "/api/doctor/scan", strings.NewReader(`{"sourceGameVersion":"not-a-version"}`))
	bad.Header.Set("Content-Type", "application/json")
	badRec := httptest.NewRecorder()
	app.handleDoctorScan(badRec, bad)
	if badRec.Code != http.StatusBadRequest {
		t.Fatalf("bad status=%d body=%s", badRec.Code, badRec.Body.String())
	}

	goodBody := []byte(`{"sourceGameVersion":"1.20.1","sourceLoader":"forge","targetGameVersion":"26.1","targetLoader":"neoforge"}`)
	good := httptest.NewRequest(http.MethodPost, "/api/doctor/scan", bytes.NewReader(goodBody))
	good.Header.Set("Content-Type", "application/json")
	goodRec := httptest.NewRecorder()
	app.handleDoctorScan(goodRec, good)
	if goodRec.Code != http.StatusOK {
		t.Fatalf("good status=%d body=%s", goodRec.Code, goodRec.Body.String())
	}
	var scan DoctorScan
	if err := json.Unmarshal(goodRec.Body.Bytes(), &scan); err != nil {
		t.Fatal(err)
	}
	if scan.TargetJava != 25 || !crossesDeobfuscationBoundary(scan.SourceGameVersion, scan.TargetGameVersion) {
		t.Fatalf("scan=%+v", scan)
	}
	ids := doctorFindingIDs(scan.GlobalFindings)
	if !ids["global-2026-boundary"] || !ids["global-loader-change"] {
		t.Fatalf("global findings=%v", ids)
	}
}

func TestDoctorToolsEndpointUsesCanonicalKnowledge(t *testing.T) {
	app := &App{}
	req := httptest.NewRequest(http.MethodGet, "/api/doctor/tools", nil)
	rec := httptest.NewRecorder()
	app.handleDoctorTools(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var payload DoctorToolsPayload
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Total < 100 || len(payload.Tools) != payload.Total {
		t.Fatalf("total=%d tools=%d", payload.Total, len(payload.Tools))
	}
	if len(payload.RepairPatterns) < 8 {
		t.Fatalf("repair patterns=%d", len(payload.RepairPatterns))
	}
	seenTool := false
	for _, tool := range payload.Tools {
		if tool.ID == "vineflower" && tool.OfficialURL != "" {
			seenTool = true
		}
	}
	if !seenTool {
		t.Fatal("canonical Vineflower tool missing")
	}
	seenPattern := false
	for _, pattern := range payload.RepairPatterns {
		if pattern.ID == "binary-owner-descriptor-drift" {
			seenPattern = true
		}
	}
	if !seenPattern {
		t.Fatal("repair-brain binary drift pattern missing")
	}
}

func TestDoctorAllSourceReferencesResolveToCatalog(t *testing.T) {
	knowledge, err := loadDoctorKnowledge()
	if err != nil {
		t.Fatal(err)
	}
	known := map[string]bool{}
	for _, source := range knowledge.Sources {
		if known[source.ID] {
			t.Fatalf("duplicate source ID %q", source.ID)
		}
		known[source.ID] = true
	}
	blockPattern := regexp.MustCompile(`(?s)SourceIDs\s*(?::|=)\s*\[\]string\{([^}]*)\}`)
	idPattern := regexp.MustCompile(`"([a-z0-9._-]+)"`)
	for _, filename := range []string{"doctor.go", "doctor_log.go", "doctor_support.go"} {
		data, err := os.ReadFile(filename)
		if err != nil {
			t.Fatal(err)
		}
		for _, block := range blockPattern.FindAllSubmatch(data, -1) {
			for _, match := range idPattern.FindAllSubmatch(block[1], -1) {
				id := string(match[1])
				if !known[id] {
					t.Errorf("%s references missing Doctor source ID %q", filename, id)
				}
			}
		}
	}
}

func TestEmbeddedWebHonorsStrictStyleCSP(t *testing.T) {
	for _, name := range []string{"web/index.html", "web/app.js", "web/catalog.js"} {
		data, err := embeddedFiles.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		if strings.Contains(text, `style="`) {
			t.Fatalf("%s contains an inline style attribute blocked by style-src 'self'", name)
		}
		if strings.Contains(text, ".style.") {
			t.Fatalf("%s mutates element.style and will violate style-src 'self'", name)
		}
	}
}

func TestDoctorUIContractIsEmbeddedAndStyled(t *testing.T) {
	indexBytes, err := embeddedFiles.ReadFile("web/index.html")
	if err != nil {
		t.Fatal(err)
	}
	appBytes, err := embeddedFiles.ReadFile("web/app.js")
	if err != nil {
		t.Fatal(err)
	}
	styleBytes, err := embeddedFiles.ReadFile("web/styles.css")
	if err != nil {
		t.Fatal(err)
	}
	index, appJS, styles := string(indexBytes), string(appBytes), string(styleBytes)
	for _, marker := range []string{`data-view="doctor"`, `id="doctorScan"`, `id="doctorDependencyGraph"`, `id="doctorLogText"`, `id="doctorToolGrid"`, `id="creatorTranscriptModal"`, `id="creatorTranscriptClose"`, `id="creatorTranscriptBody"`, `href="favicon.svg"`} {
		if !strings.Contains(index, marker) {
			t.Errorf("embedded Doctor UI missing %q", marker)
		}
	}
	for _, marker := range []string{"/api/doctor/scan", "/api/doctor/log", "/api/doctor/tools", "renderDoctorDependencyGraph", "renderDoctorScan"} {
		if !strings.Contains(appJS, marker) {
			t.Errorf("embedded Doctor JavaScript missing %q", marker)
		}
	}
	for _, marker := range []string{".doctor-artifact", ".doctor-graph", ".doctor-tool", ".severity-tag.critical"} {
		if !strings.Contains(styles, marker) {
			t.Errorf("embedded Doctor CSS missing %q", marker)
		}
	}
}

func TestDoctorLegacyVersionAndToolchainRouting(t *testing.T) {
	legacyCases := map[string]bool{
		"b1.7.3": true,
		"1.7.10": true,
		"1.12.2": true,
		"1.13.2": true,
		"1.14.4": false,
		"1.21.1": false,
		"26.1":   false,
	}
	for version, want := range legacyCases {
		if got := isLegacyMinecraftVersion(version); got != want {
			t.Errorf("isLegacyMinecraftVersion(%q)=%t want %t", version, got, want)
		}
	}

	cases := []struct {
		name     string
		loader   string
		version  string
		required []string
	}{
		{name: "forge-1710", loader: "forge", version: "1.7.10", required: []string{"retrofuturagradle", "fpgradle", "retromcp-java"}},
		{name: "forge-1122", loader: "forge", version: "1.12.2", required: []string{"forgegradle", "cleanroom", "mixinbooter", "fugue"}},
		{name: "legacy-fabric", loader: "fabric", version: "1.13.2", required: []string{"legacy-fabric-looming", "legacy-fabric-api", "ornithe-feather"}},
		{name: "babric-beta", loader: "babric", version: "b1.7.3", required: []string{"stationapi", "ornithe-ploceus", "ornithe-feather"}},
		{name: "modern-neoforge", loader: "neoforge", version: "1.21.1", required: []string{"moddevgradle", "neoforge-mod-generator", "neoform-runtime", "neoform"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := map[string]bool{}
			for _, id := range doctorBuildSources(tc.loader, tc.version) {
				got[id] = true
			}
			for _, id := range tc.required {
				if !got[id] {
					t.Errorf("doctorBuildSources(%q,%q) missing %q; got=%v", tc.loader, tc.version, id, got)
				}
			}
		})
	}
}

func TestDoctorLegacyPlanSeparatesPortingFromRuntimeCompatibility(t *testing.T) {
	request := DoctorScanRequest{
		SourceGameVersion: "1.7.10",
		SourceLoader:      "forge",
		TargetGameVersion: "1.21.1",
		TargetLoader:      "neoforge",
	}
	steps := buildDoctorModPlan(LocalModFile{}, JarSignals{MixinConfigs: []string{"legacy.mixins.json"}, DataFileCount: 1}, request, DoctorModReport{})
	byID := map[string]DoctorStep{}
	for _, step := range steps {
		byID[step.ID] = step
	}
	legacy, ok := byID["legacy-toolchain-track"]
	if !ok {
		t.Fatalf("legacy-toolchain-track missing; steps=%v", byID)
	}
	if !strings.Contains(strings.ToLower(legacy.Action), "never report a runtime bridge as a completed source port") {
		t.Fatalf("legacy action does not preserve source/runtime separation: %s", legacy.Action)
	}
	for _, id := range []string{"retrofuturagradle", "retromcp-java", "retrofuturabootstrap", "unimixins", "lwjgl3ify", "moddevgradle", "neoform-runtime"} {
		found := false
		for _, sourceID := range legacy.SourceIDs {
			if sourceID == id {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("legacy step missing source %q: %v", id, legacy.SourceIDs)
		}
	}
	packaging := byID["package-and-prove"]
	packagingIDs := map[string]bool{}
	for _, id := range packaging.SourceIDs {
		packagingIDs[id] = true
	}
	for _, id := range []string{"mod-publish-plugin", "minotaur", "forgix", "modshade"} {
		if !packagingIDs[id] {
			t.Errorf("package-and-prove missing %q: %v", id, packaging.SourceIDs)
		}
	}
	if packagingIDs["modrinth-minotaur"] {
		t.Fatal("package-and-prove retained stale non-catalog source id modrinth-minotaur")
	}
}

func TestDoctorGeneratedPlanSourceIDsResolveToCatalog(t *testing.T) {
	knowledge, err := loadDoctorKnowledge()
	if err != nil {
		t.Fatal(err)
	}
	known := map[string]bool{}
	for _, source := range knowledge.Sources {
		known[source.ID] = true
	}
	requests := []DoctorScanRequest{
		{SourceGameVersion: "b1.7.3", SourceLoader: "babric", TargetGameVersion: "1.7.10", TargetLoader: "forge"},
		{SourceGameVersion: "1.7.10", SourceLoader: "forge", TargetGameVersion: "1.12.2", TargetLoader: "forge"},
		{SourceGameVersion: "1.12.2", SourceLoader: "forge", TargetGameVersion: "1.21.1", TargetLoader: "neoforge"},
		{SourceGameVersion: "1.13.2", SourceLoader: "fabric", TargetGameVersion: "1.20.1", TargetLoader: "fabric"},
		{SourceGameVersion: "1.21.1", SourceLoader: "fabric", TargetGameVersion: "26.1", TargetLoader: "neoforge"},
	}
	for _, request := range requests {
		steps := buildDoctorModPlan(LocalModFile{}, JarSignals{
			MixinConfigs:           []string{"example.mixins.json"},
			AccessWideners:         []string{"example.accesswidener"},
			DataFileCount:          1,
			AssetFileCount:         1,
			HasPackMetadata:        true,
			TransformationServices: []string{"example.Service"},
		}, request, DoctorModReport{})
		scan := DoctorScan{
			SourceGameVersion: request.SourceGameVersion,
			SourceLoader:      request.SourceLoader,
			TargetGameVersion: request.TargetGameVersion,
			TargetLoader:      request.TargetLoader,
			Summary:           DoctorScanSummary{BinaryOnly: 1, CriticalRisk: 1},
		}
		steps = append(steps, buildGlobalDoctorPlan(scan)...)
		findings := buildGlobalDoctorFindings(scan)
		for _, step := range steps {
			for _, id := range step.SourceIDs {
				if !known[id] {
					t.Errorf("%s/%s -> %s/%s step %q references missing source %q", request.SourceGameVersion, request.SourceLoader, request.TargetGameVersion, request.TargetLoader, step.ID, id)
				}
			}
		}
		for _, finding := range findings {
			for _, id := range finding.SourceIDs {
				if !known[id] {
					t.Errorf("%s/%s -> %s/%s finding %q references missing source %q", request.SourceGameVersion, request.SourceLoader, request.TargetGameVersion, request.TargetLoader, finding.ID, id)
				}
			}
		}
	}
}

func TestDoctorCatalogContainsCurrentLegacyAndModernLeaders(t *testing.T) {
	knowledge, err := loadDoctorKnowledge()
	if err != nil {
		t.Fatal(err)
	}
	if knowledge.Total < 150 {
		t.Fatalf("knowledge total=%d want at least 150 after legacy/current expansion", knowledge.Total)
	}
	seen := map[string]bool{}
	for _, source := range knowledge.Sources {
		seen[source.ID] = true
	}
	for _, id := range []string{
		"neoform", "neoform-runtime", "retromcp-java", "retrofuturagradle", "retrofuturabootstrap",
		"cleanroom", "mixinbooter", "unimixins", "lwjgl3ify", "fugue", "fabric-meta",
		"ornithe-feather", "ornithe-ploceus", "legacy-fabric-api", "stationapi", "legacyfix",
		"mod-publish-plugin", "forgix", "modshade", "papermc-dataconverter",
	} {
		if !seen[id] {
			t.Errorf("expanded catalog missing %q", id)
		}
	}
}
