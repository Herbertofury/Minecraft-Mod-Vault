package main

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// TestStaticUIControlContract prevents the embedded HTML and JavaScript from
// drifting apart. Every control wired unconditionally by wireStatic must exist
// in the shipped document, otherwise the first missing control aborts all later
// UI wiring in the browser.
func TestStaticUIControlContract(t *testing.T) {
	htmlBytes, err := os.ReadFile("web/index.html")
	if err != nil {
		t.Fatal(err)
	}
	jsBytes, err := os.ReadFile("web/app.js")
	if err != nil {
		t.Fatal(err)
	}
	html := string(htmlBytes)
	js := string(jsBytes)

	start := strings.Index(js, "function wireStatic()")
	if start < 0 {
		t.Fatal("wireStatic function is missing")
	}
	end := strings.Index(js[start:], "\nfunction injectIcons()")
	if end < 0 {
		t.Fatal("wireStatic/injectIcons boundary is missing")
	}
	wire := js[start : start+end]

	wiredID := regexp.MustCompile(`document\.getElementById\(['"]([^'"]+)['"]\)\.(?:onclick|onchange|oninput|onkeydown|ondrop)`)
	htmlID := regexp.MustCompile(`\bid=["']([^"']+)["']`)
	present := map[string]bool{}
	for _, match := range htmlID.FindAllStringSubmatch(html, -1) {
		present[match[1]] = true
	}

	seen := map[string]bool{}
	for _, match := range wiredID.FindAllStringSubmatch(wire, -1) {
		id := match[1]
		if seen[id] {
			continue
		}
		seen[id] = true
		if !present[id] {
			t.Errorf("wireStatic unconditionally wires missing HTML control %q", id)
		}
	}

	for _, id := range []string{
		"doctorScan", "doctorAnalyzeLog", "doctorToolGrid",
		"doctorBuildRepairPlan", "doctorApplyRepairPlan", "doctorRepairPlan",
		"portingBuildPlan", "portingProbeEnvironment", "portingGenerateWorkspace",
		"portingVersionGrid", "portingWorkspaceGrid", "portingPlanResult",
		"projectModalClose", "creatorTranscriptModal",
		"creatorTranscriptClose", "creatorTranscriptBody",
	} {
		if !present[id] {
			t.Errorf("required embedded UI element %q is missing", id)
		}
	}

	for _, required := range []string{
		`['porting','Porting Lab','compass']`,
		`if(id==='porting'){syncPortingTargets();loadPortingAtlas();loadPortingEnvironment();loadPortingWorkspaces()}`,
		`wirePortingVersionCards();`,
		`/api/doctor/repair/apply`,
		`/api/porting/workspace`,
	} {
		if !strings.Contains(js, required) {
			t.Errorf("required v0.9 UI wiring missing %q", required)
		}
	}

	omniBytes, err := os.ReadFile("web/omnimanager.js")
	if err != nil {
		t.Fatal(err)
	}
	omni := string(omniBytes)
	for _, required := range []string{
		`data-manager-port`,
		`selectPortingFile(item.path)`,
		`/api/library`,
		`/api/library/action`,
		`/api/library/bedrock/install`,
		`/api/library/bedrock/activate`,
	} {
		if !strings.Contains(omni, required) {
			t.Errorf("required OmniManager UI wiring missing %q", required)
		}
	}
}

func TestRepairLabUIControlContract(t *testing.T) {
	htmlBytes, err := os.ReadFile("web/index.html")
	if err != nil {
		t.Fatal(err)
	}
	jsBytes, err := os.ReadFile("web/repair-lab.js")
	if err != nil {
		t.Fatal(err)
	}
	html := string(htmlBytes)
	js := string(jsBytes)
	if !strings.Contains(html, `data-view="repair"`) {
		t.Fatal("Repair Lab top-level workspace is missing")
	}
	if !strings.Contains(html, `<script src="repair-lab.js"></script><script src="app.js"></script>`) {
		t.Fatal("Repair Lab script must load before app.js initializes routing")
	}
	if !strings.Contains(string(mustReadTestFile(t, "web/app.js")), "['repair','Repair Lab','shield']") {
		t.Fatal("Repair Lab navigation entry is missing")
	}

	htmlID := regexp.MustCompile(`\bid=["']([^"']+)["']`)
	present := map[string]bool{}
	for _, match := range htmlID.FindAllStringSubmatch(html, -1) {
		present[match[1]] = true
	}
	usedID := regexp.MustCompile(`document\.getElementById\(['"]([^'"]+)['"]\)`)
	for _, match := range usedID.FindAllStringSubmatch(js, -1) {
		if !present[match[1]] {
			t.Errorf("Repair Lab JavaScript references missing HTML element %q", match[1])
		}
	}
	for _, endpoint := range []string{
		"/api/repair-lab/status", "/api/repair-lab/import", "/api/repair-lab/session",
		"/api/repair-lab/prepare", "/api/repair-lab/run", "/api/repair-lab/cancel",
		"/api/repair-lab/reset", "/api/repair-lab/export", "/api/repair-lab/download",
		"/api/brain/search",
	} {
		if !strings.Contains(js, endpoint) {
			t.Errorf("Repair Lab UI is not wired to %s", endpoint)
		}
	}
}

func TestOmniManagerUIControlContract(t *testing.T) {
	htmlBytes, err := os.ReadFile("web/index.html")
	if err != nil {
		t.Fatal(err)
	}
	jsBytes, err := os.ReadFile("web/omnimanager.js")
	if err != nil {
		t.Fatal(err)
	}
	html := string(htmlBytes)
	js := string(jsBytes)
	if !strings.Contains(html, `data-view="manager"`) {
		t.Fatal("OmniManager top-level workspace is missing")
	}
	if !strings.Contains(html, `<script src="omnimanager.js"></script><script src="repair-lab.js"></script><script src="app.js"></script>`) {
		t.Fatal("OmniManager must load before app.js while preserving Repair Lab initialization order")
	}
	for _, endpoint := range []string{
		"/api/library", "/api/library/action", "/api/library/history", "/api/library/undo",
		"/api/library/bedrock/install", "/api/library/bedrock/activate",
	} {
		if !strings.Contains(js, endpoint) {
			t.Errorf("OmniManager UI is not wired to %s", endpoint)
		}
	}
	htmlID := regexp.MustCompile(`\bid=["']([^"']+)["']`)
	present := map[string]bool{}
	for _, source := range []string{html, js} {
		for _, match := range htmlID.FindAllStringSubmatch(source, -1) {
			present[match[1]] = true
		}
	}
	usedID := regexp.MustCompile(`\$\(['"]([^'"]+)['"]\)`)
	for _, match := range usedID.FindAllStringSubmatch(js, -1) {
		if !present[match[1]] {
			t.Errorf("OmniManager JavaScript references missing HTML element %q", match[1])
		}
	}
}

func TestTestGridUIControlContract(t *testing.T) {
	html := string(mustReadTestFile(t, "web/index.html"))
	js := string(mustReadTestFile(t, "web/testgrid.js"))
	appJS := string(mustReadTestFile(t, "web/app.js"))
	if !strings.Contains(html, `data-view="testgrid"`) {
		t.Fatal("TestGrid top-level workspace is missing")
	}
	if !strings.Contains(html, `<script src="testgrid.js"></script><script src="omnimanager.js"></script><script src="repair-lab.js"></script><script src="app.js"></script>`) {
		t.Fatal("TestGrid must load before app.js while preserving OmniManager and Repair Lab initialization order")
	}
	if !strings.Contains(appJS, "['testgrid','TestGrid','gauge']") || !strings.Contains(appJS, "window.TestGridStudio?.activate()") {
		t.Fatal("TestGrid navigation or activation hook is missing")
	}
	htmlID := regexp.MustCompile(`\bid=["']([^"']+)["']`)
	present := map[string]bool{}
	for _, match := range htmlID.FindAllStringSubmatch(html, -1) {
		present[match[1]] = true
	}
	usedID := regexp.MustCompile(`document\.getElementById\(['"]([^'"]+)['"]\)`)
	for _, match := range usedID.FindAllStringSubmatch(js, -1) {
		if !present[match[1]] {
			t.Errorf("TestGrid JavaScript references missing HTML element %q", match[1])
		}
	}
	for _, endpoint := range []string{
		"/api/testgrid/capabilities", "/api/testgrid/validate", "/api/testgrid/run",
		"/api/testgrid/runs", "/api/testgrid/cancel", "/api/testgrid/file",
	} {
		if !strings.Contains(js, endpoint) {
			t.Errorf("TestGrid UI is not wired to %s", endpoint)
		}
	}
}

func mustReadTestFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestOmniBridgeUIControlContract(t *testing.T) {
	html := string(mustReadTestFile(t, "web/index.html"))
	js := string(mustReadTestFile(t, "web/conversion.js"))
	app := string(mustReadTestFile(t, "web/app.js"))
	if !strings.Contains(html, `data-view="conversion"`) {
		t.Fatal("OmniBridge top-level workspace is missing")
	}
	if !strings.Contains(html, `<script src="conversion.js"></script>`) {
		t.Fatal("OmniBridge script is not loaded")
	}
	if !strings.Contains(app, "['conversion','OmniBridge','wand']") || !strings.Contains(app, "loadConversionStudio()") {
		t.Fatal("OmniBridge navigation or activation is missing")
	}
	htmlID := regexp.MustCompile(`\bid=["']([^"']+)["']`)
	present := map[string]bool{}
	for _, match := range htmlID.FindAllStringSubmatch(html, -1) {
		present[match[1]] = true
	}
	usedID := regexp.MustCompile(`document\.getElementById\(['"]([^'"]+)['"]\)`)
	for _, match := range usedID.FindAllStringSubmatch(js, -1) {
		if !present[match[1]] {
			t.Errorf("OmniBridge JavaScript references missing HTML element %q", match[1])
		}
	}
	for _, endpoint := range []string{
		"/api/conversion/status", "/api/conversion/import", "/api/conversion/import-path",
		"/api/conversion/session", "/api/conversion/plan", "/api/conversion/run",
		"/api/conversion/reset", "/api/conversion/download", "/api/conversion/tool",
		"/api/conversion/adapter/run",
	} {
		if !strings.Contains(js, endpoint) {
			t.Errorf("OmniBridge UI is not wired to %s", endpoint)
		}
	}
	for _, control := range []string{
		"conversionChooseSource", "conversionImportPath", "conversionBuildPlan", "conversionRun",
		"conversionOpenSession", "conversionDeleteSession", "conversionDownloadReport",
	} {
		if !strings.Contains(js, control) {
			t.Errorf("OmniBridge control %q has no JavaScript wiring", control)
		}
	}
}
