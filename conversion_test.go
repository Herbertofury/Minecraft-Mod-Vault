package main

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func conversionTestArchive(t *testing.T, files map[string][]byte, modes map[string]os.FileMode) []byte {
	t.Helper()
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for name, data := range files {
		header := &zip.FileHeader{Name: name, Method: zip.Deflate}
		if mode := modes[name]; mode != 0 {
			header.SetMode(mode)
		}
		entry, err := writer.CreateHeader(header)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func conversionTestApp(t *testing.T) *App {
	t.Helper()
	root := t.TempDir()
	return &App{cfgDir: root, settings: Settings{JavaRoot: filepath.Join(root, "java"), BedrockRoot: filepath.Join(root, "bedrock"), WorldRoot: filepath.Join(root, "world")}}
}

func TestConversionArchiveSecurity(t *testing.T) {
	cases := []struct {
		name  string
		files map[string][]byte
		modes map[string]os.FileMode
	}{
		{"traversal", map[string][]byte{"../escape.txt": []byte("no")}, nil},
		{"absolute", map[string][]byte{"/escape.txt": []byte("no")}, nil},
		{"case collision", map[string][]byte{"A.txt": []byte("one"), "a.txt": []byte("two")}, nil},
		{"symlink", map[string][]byte{"link": []byte("target")}, map[string]os.FileMode{"link": os.ModeSymlink | 0o777}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			archive := filepath.Join(t.TempDir(), "bad.zip")
			if err := os.WriteFile(archive, conversionTestArchive(t, test.files, test.modes), 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := extractConversionArchive(archive, filepath.Join(t.TempDir(), "out")); err == nil {
				t.Fatalf("unsafe archive %q was accepted", test.name)
			}
		})
	}
}

func TestRawStructureImportCreatesGraph(t *testing.T) {
	app := conversionTestApp(t)
	raw := append([]byte{0x1f, 0x8b, 0x08, 0x00}, bytes.Repeat([]byte{0}, 64)...)
	session, err := app.importConversionFile("castle.schem", bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	if session.Source.Kind != "structure" || session.Graph.Summary.Total == 0 {
		t.Fatalf("unexpected raw structure profile: %#v summary=%#v", session.Source, session.Graph.Summary)
	}
	if !pathExists(filepath.Join(session.Paths.Extracted, "structures", "castle.schem")) {
		t.Fatal("raw structure was not preserved in the immutable extracted source")
	}
}

func TestJavaDataResourcePackToBedrockAddon(t *testing.T) {
	app := conversionTestApp(t)
	archive := conversionTestArchive(t, map[string][]byte{
		"pack.mcmeta":                          []byte(`{"pack":{"pack_format":48,"description":"OmniBridge fixture"}}`),
		"data/demo/recipes/glow.json":          []byte(`{"type":"minecraft:crafting_shapeless","ingredients":[{"item":"minecraft:glowstone_dust"}],"result":{"id":"minecraft:glowstone","count":1}}`),
		"data/demo/functions/hello.mcfunction": []byte("say Hello from Java\n"),
		"assets/demo/lang/en_us.json":          []byte(`{"item.demo.glow":"Glowing Example"}`),
		"assets/demo/textures/item/glow.png":   []byte("not-a-real-png-but-losslessly-preserved"),
	}, nil)
	session, err := app.importConversionFile("demo-pack.zip", bytes.NewReader(archive))
	if err != nil {
		t.Fatal(err)
	}
	if session.Source.Edition != "java" || session.Source.Kind != "datapack" {
		t.Fatalf("unexpected source profile: %#v", session.Source)
	}
	plan, err := buildConversionPlan(session, ConversionTargetSpec{Format: "bedrock-addon", GameVersion: "1.26.30", Name: "Demo Bridge", Namespace: "demo"})
	if err != nil {
		t.Fatal(err)
	}
	session.Plan = plan
	if err := executeConversion(session); err != nil {
		t.Fatal(err)
	}
	if len(session.Outputs) < 2 || !session.Outputs[0].Validated {
		t.Fatalf("expected validated add-on and proof outputs: %#v", session.Outputs)
	}
	entries := zipEntryNames(t, filepath.Join(session.Paths.Outputs, session.Outputs[0].Name))
	for _, required := range []string{"manifest.json", "recipes", "texts/en_US.lang", "item_texture.json"} {
		if !anyContains(entries, required) {
			t.Fatalf("Bedrock add-on missing %q: %v", required, entries)
		}
	}
	if plan.Coverage.Translated == 0 {
		t.Fatal("expected deterministic translation coverage")
	}
	if len(plan.ReviewQueue) == 0 {
		t.Fatal("Java command semantics should remain visible in the review queue")
	}
}

func TestBedrockAddonToJavaPacksAndFabricWorkspace(t *testing.T) {
	app := conversionTestApp(t)
	bpManifest := map[string]any{"format_version": 2, "header": map[string]any{"name": "Bedrock Demo", "description": "fixture", "uuid": "11111111-1111-4111-8111-111111111111", "version": []int{1, 2, 3}, "min_engine_version": []int{1, 21, 0}}, "modules": []map[string]any{{"type": "data", "uuid": "22222222-2222-4222-8222-222222222222", "version": []int{1, 0, 0}}, {"type": "script", "language": "javascript", "entry": "scripts/main.js", "uuid": "33333333-3333-4333-8333-333333333333", "version": []int{1, 0, 0}}}}
	rpManifest := map[string]any{"format_version": 2, "header": map[string]any{"name": "Bedrock Demo Resources", "uuid": "44444444-4444-4444-8444-444444444444", "version": []int{1, 0, 0}, "min_engine_version": []int{1, 21, 0}}, "modules": []map[string]any{{"type": "resources", "uuid": "55555555-5555-4555-8555-555555555555", "version": []int{1, 0, 0}}}}
	bpJSON, _ := json.Marshal(bpManifest)
	rpJSON, _ := json.Marshal(rpManifest)
	archive := conversionTestArchive(t, map[string][]byte{
		"Demo_BP/manifest.json":           bpJSON,
		"Demo_BP/recipes/demo.json":       []byte(`{"format_version":"1.20.10","minecraft:recipe_shapeless":{"description":{"identifier":"demo:test"},"ingredients":[{"item":"minecraft:stone"}],"result":{"item":"minecraft:diamond","count":1}}}`),
		"Demo_BP/scripts/main.js":         []byte(`import { world } from "@minecraft/server"; world.sendMessage("demo");`),
		"Demo_RP/manifest.json":           rpJSON,
		"Demo_RP/texts/en_US.lang":        []byte("item.demo:test.name=Demo Item\n"),
		"Demo_RP/textures/items/demo.png": []byte("texture"),
	}, nil)
	session, err := app.importConversionFile("bedrock-demo.mcaddon", bytes.NewReader(archive))
	if err != nil {
		t.Fatal(err)
	}
	if session.Source.Edition != "bedrock" || session.Source.Kind != "addon-family" {
		t.Fatalf("unexpected Bedrock source profile: %#v", session.Source)
	}
	for _, target := range []ConversionTargetSpec{
		{Format: "java-datapack", GameVersion: "1.21.1", Name: "Bedrock Demo", Namespace: "bedrock_demo"},
		{Format: "java-resourcepack", GameVersion: "1.21.1", Name: "Bedrock Demo", Namespace: "bedrock_demo"},
		{Format: "java-fabric", GameVersion: "1.21.1", Loader: "fabric", Name: "Bedrock Demo", Namespace: "bedrock_demo"},
	} {
		plan, err := buildConversionPlan(session, target)
		if err != nil {
			t.Fatal(err)
		}
		session.Plan = plan
		if err := executeConversion(session); err != nil {
			t.Fatalf("%s conversion failed: %v", target.Format, err)
		}
		if len(session.Outputs) < 2 || !session.Outputs[0].Validated {
			t.Fatalf("%s output was not validated: %#v", target.Format, session.Outputs)
		}
		entries := zipEntryNames(t, filepath.Join(session.Paths.Outputs, session.Outputs[0].Name))
		switch target.Format {
		case "java-datapack", "java-resourcepack":
			if !anyContains(entries, "pack.mcmeta") {
				t.Fatalf("%s output missing pack.mcmeta", target.Format)
			}
		case "java-fabric":
			for _, required := range []string{"build.gradle.kts", "fabric.mod.json", "omnibridge/contracts"} {
				if !anyContains(entries, required) {
					t.Fatalf("Fabric workspace missing %q", required)
				}
			}
		}
	}
}

func TestBedrockWorldTemplateAndCrossEditionWorldWorkspace(t *testing.T) {
	app := conversionTestApp(t)
	bedrockWorld := conversionTestArchive(t, map[string][]byte{"level.dat": []byte("level"), "levelname.txt": []byte("Demo World"), "db/000001.ldb": []byte("chunk")}, nil)
	session, err := app.importConversionFile("demo.mcworld", bytes.NewReader(bedrockWorld))
	if err != nil {
		t.Fatal(err)
	}
	plan, err := buildConversionPlan(session, ConversionTargetSpec{Format: "bedrock-template", GameVersion: "1.26.30", Name: "Demo Template", Namespace: "demo_template"})
	if err != nil {
		t.Fatal(err)
	}
	session.Plan = plan
	if err := executeConversion(session); err != nil {
		t.Fatal(err)
	}
	entries := zipEntryNames(t, filepath.Join(session.Paths.Outputs, session.Outputs[0].Name))
	for _, required := range []string{"manifest.json", "texts/en_US.lang", "level.dat", "db/000001.ldb"} {
		if !anyContains(entries, required) {
			t.Fatalf("world template missing %q", required)
		}
	}

	javaWorld := conversionTestArchive(t, map[string][]byte{"level.dat": []byte("java-level"), "region/r.0.0.mca": []byte("region")}, nil)
	javaSession, err := app.importConversionFile("java-world.zip", bytes.NewReader(javaWorld))
	if err != nil {
		t.Fatal(err)
	}
	plan, err = buildConversionPlan(javaSession, ConversionTargetSpec{Format: "bedrock-world", GameVersion: "1.26.30", Name: "Converted World", Namespace: "converted_world"})
	if err != nil {
		t.Fatal(err)
	}
	javaSession.Plan = plan
	if err := executeConversion(javaSession); err != nil {
		t.Fatal(err)
	}
	if javaSession.State != "review-required" || !strings.Contains(javaSession.Outputs[0].Name, "adapter-workspace") {
		t.Fatalf("cross-edition world was falsely presented as complete: state=%s outputs=%#v", javaSession.State, javaSession.Outputs)
	}
	entries = zipEntryNames(t, filepath.Join(javaSession.Paths.Outputs, javaSession.Outputs[0].Name))
	if !anyContains(entries, "README.md") || !anyContains(entries, "source/region/r.0.0.mca") {
		t.Fatalf("world adapter workspace is incomplete: %v", entries)
	}
}

func TestJavaBytecodeAndMixinNeverClaimAutomaticBedrockParity(t *testing.T) {
	app := conversionTestApp(t)
	archive := conversionTestArchive(t, map[string][]byte{
		"META-INF/mods.toml": []byte("modLoader=\"javafml\"\n[[mods]]\nmodId=\"demo\"\nversion=\"1.0.0\"\ndisplayName=\"Demo\"\n"),
		"demo/Demo.class":    []byte{0xca, 0xfe, 0xba, 0xbe},
		"demo.mixins.json":   []byte(`{"required":true,"package":"demo.mixin","mixins":["GameMixin"]}`),
	}, nil)
	session, err := app.importConversionFile("demo.jar", bytes.NewReader(archive))
	if err != nil {
		t.Fatal(err)
	}
	plan, err := buildConversionPlan(session, ConversionTargetSpec{Format: "bedrock-addon", GameVersion: "1.26.30", Name: "Demo", Namespace: "demo"})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Coverage.Review == 0 || plan.Coverage.Blocked == 0 || plan.Coverage.CompletenessState == "fully-automated-plan" {
		t.Fatalf("bytecode/Mixin conversion was overclaimed: %#v", plan.Coverage)
	}
	joined := ""
	for _, item := range plan.ReviewQueue {
		joined += item.Reason + " " + item.SuggestedRoute + "\n"
	}
	if !strings.Contains(joined, "JVM bytecode") || !strings.Contains(joined, "Reimplement") {
		t.Fatalf("semantic reconstruction evidence missing: %s", joined)
	}
}

func zipEntryNames(t *testing.T, path string) []string {
	t.Helper()
	reader, err := zip.OpenReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	out := make([]string, 0, len(reader.File))
	for _, file := range reader.File {
		out = append(out, filepath.ToSlash(file.Name))
	}
	return out
}

func anyContains(values []string, fragment string) bool {
	for _, value := range values {
		if strings.Contains(value, fragment) {
			return true
		}
	}
	return false
}

func TestWorldProductTargetsAndBedrockFeatureInventory(t *testing.T) {
	app := conversionTestApp(t)
	manifest := map[string]any{"format_version": 2, "header": map[string]any{"name": "Premium Template", "uuid": "11111111-1111-4111-8111-111111111111", "version": []int{1, 0, 0}, "base_game_version": []int{1, 26, 30}}, "modules": []map[string]any{{"type": "world_template", "uuid": "22222222-2222-4222-8222-222222222222", "version": []int{1, 0, 0}}}}
	manifestJSON, _ := json.Marshal(manifest)
	archive := conversionTestArchive(t, map[string][]byte{
		"manifest.json":                     manifestJSON,
		"level.dat":                         []byte("bedrock-level"),
		"db/000001.ldb":                     []byte("chunk"),
		"behavior_packs/Demo/blocks/a.json": []byte(`{"format_version":"1.21.0","minecraft:block":{"description":{"identifier":"demo:a"},"components":{}}}`),
		"behavior_packs/Demo/camera_presets/follow.json": []byte(`{"format_version":"1.21.0","minecraft:camera_preset":{"identifier":"demo:follow"}}`),
		"behavior_packs/Demo/dialogue/guide.json":        []byte(`{"minecraft:npc_dialogue":{"scenes":[]}}`),
		"behavior_packs/Demo/trading/trades.json":        []byte(`{"tiers":[]}`),
		"behavior_packs/Demo/volumes/zone.json":          []byte(`{"format_version":"1.21.0"}`),
		"resource_packs/Demo/textures/blocks/a.png":      []byte("texture"),
	}, nil)
	session, err := app.importConversionFile("premium.mctemplate", bytes.NewReader(archive))
	if err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"camera-preset", "dialogue", "trade-table", "volume"} {
		if session.Graph.Summary.ByKind[kind] == 0 {
			t.Fatalf("Bedrock feature %q was not represented in UMCG: %#v", kind, session.Graph.Summary.ByKind)
		}
	}
	plan, err := buildConversionPlan(session, ConversionTargetSpec{Format: "bedrock-world-product", GameVersion: "1.26.30", Name: "Premium Product", Namespace: "premium"})
	if err != nil {
		t.Fatal(err)
	}
	session.Plan = plan
	if err := executeConversion(session); err != nil {
		t.Fatal(err)
	}
	entries := zipEntryNames(t, filepath.Join(session.Paths.Outputs, session.Outputs[0].Name))
	for _, required := range []string{"PRODUCT-MANIFEST.json", ".mctemplate", ".mcaddon", "companion-addon/config.json", "OMNIBRIDGE-ACCEPTANCE.json"} {
		if !anyContains(entries, required) {
			t.Fatalf("Bedrock world product missing %q: %v", required, entries)
		}
	}

	javaPlan, err := buildConversionPlan(session, ConversionTargetSpec{Format: "java-world-mod", GameVersion: "1.21.11", Loader: "fabric", Name: "Premium Java World", Namespace: "premium_java"})
	if err != nil {
		t.Fatal(err)
	}
	session.Plan = javaPlan
	if err := executeConversion(session); err != nil {
		t.Fatal(err)
	}
	entries = zipEntryNames(t, filepath.Join(session.Paths.Outputs, session.Outputs[0].Name))
	for _, required := range []string{"WorldTemplateExporter.java", "GeneratedContentRegistry.java", "source-world.mcworld", "world-product.json", "omnibridge/contracts"} {
		if !anyContains(entries, required) {
			t.Fatalf("Java world-mod workspace missing %q: %v", required, entries)
		}
	}
	if session.State != "review-required" {
		t.Fatalf("cross-edition world mod falsely reported complete: %s", session.State)
	}
}

func TestChunkerAdapterExecutionIsAllowlistedAndSourceImmutable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake POSIX adapter fixture")
	}
	app := conversionTestApp(t)
	app.settings.ConversionToolPaths = map[string]string{}
	world := conversionTestArchive(t, map[string][]byte{"level.dat": []byte("java-level"), "region/r.0.0.mca": []byte("region")}, nil)
	session, err := app.importConversionFile("java-world.zip", bytes.NewReader(world))
	if err != nil {
		t.Fatal(err)
	}
	plan, err := buildConversionPlan(session, ConversionTargetSpec{Format: "bedrock-world", GameVersion: "1.26.30", Name: "Adapter World", Namespace: "adapter_world"})
	if err != nil {
		t.Fatal(err)
	}
	session.Plan = plan
	helper := filepath.Join(t.TempDir(), "fake-chunker")
	script := `#!/bin/sh
set -eu
out=""
while [ "$#" -gt 0 ]; do
  if [ "$1" = "-o" ]; then out="$2"; shift 2; else shift; fi
done
mkdir -p "$out/db"
printf 'bedrock-level' > "$out/level.dat"
printf 'chunk' > "$out/db/000001.ldb"
`
	if err := os.WriteFile(helper, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	app.settings.ConversionToolPaths["chunker"] = helper
	var tool ConversionToolAdapter
	for _, candidate := range app.configuredConversionToolAdapters() {
		if candidate.ID == "chunker" {
			tool = candidate
		}
	}
	if !tool.Ready || !tool.CanExecute {
		t.Fatalf("fake Chunker was not detected: %#v", tool)
	}
	run, err := app.executeConversionAdapter(context.Background(), session, tool, nil)
	if err != nil {
		t.Fatal(err)
	}
	if run.State != "succeeded" || !run.SourceVerified || len(run.Outputs) != 1 || !run.Outputs[0].Validated {
		t.Fatalf("unexpected adapter result: %#v", run)
	}
	if err := verifyConversionSource(session); err != nil {
		t.Fatalf("adapter changed source: %v", err)
	}
	entries := zipEntryNames(t, filepath.Join(session.Paths.Outputs, run.Outputs[0].Name))
	if !anyContains(entries, "level.dat") || !anyContains(entries, "db/000001.ldb") {
		t.Fatalf("adapter output missing converted world data: %v", entries)
	}
}

func TestConversionAPIEndToEndAndAuth(t *testing.T) {
	root := t.TempDir()
	app := &App{cfgDir: root, stateFile: filepath.Join(root, "settings.json"), token: "test-token", settings: Settings{GameVersion: "1.21.1", Loader: "fabric", ConversionToolPaths: map[string]string{}}, httpClient: &http.Client{}}
	mux := http.NewServeMux()
	app.registerRoutes(mux)
	unauthorized := httptest.NewRequest(http.MethodGet, "/api/conversion/status", nil)
	unauthorizedResult := httptest.NewRecorder()
	mux.ServeHTTP(unauthorizedResult, unauthorized)
	if unauthorizedResult.Code != http.StatusForbidden {
		t.Fatalf("unauthorized conversion API returned %d", unauthorizedResult.Code)
	}

	archive := conversionTestArchive(t, map[string][]byte{"pack.mcmeta": []byte(`{"pack":{"pack_format":48,"description":"API fixture"}}`), "data/api/recipes/test.json": []byte(`{"type":"minecraft:crafting_shapeless","ingredients":[{"item":"minecraft:stone"}],"result":{"id":"minecraft:diamond"}}`)}, nil)
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("source", "api-pack.zip")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write(archive)
	_ = writer.Close()
	upload := httptest.NewRequest(http.MethodPost, "/api/conversion/import", &body)
	upload.Header.Set("Content-Type", writer.FormDataContentType())
	upload.Header.Set("X-MMV-Token", app.token)
	uploadResult := httptest.NewRecorder()
	mux.ServeHTTP(uploadResult, upload)
	if uploadResult.Code != http.StatusCreated {
		t.Fatalf("upload failed: %d %s", uploadResult.Code, uploadResult.Body.String())
	}
	var session ConversionSession
	if err := json.Unmarshal(uploadResult.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}
	planBody, _ := json.Marshal(ConversionPlanRequest{SessionID: session.ID, Target: ConversionTargetSpec{Format: "bedrock-addon", GameVersion: "1.26.30", Name: "API Product", Namespace: "api_product"}})
	planRequest := httptest.NewRequest(http.MethodPost, "/api/conversion/plan", bytes.NewReader(planBody))
	planRequest.Header.Set("X-MMV-Token", app.token)
	planResult := httptest.NewRecorder()
	mux.ServeHTTP(planResult, planRequest)
	if planResult.Code != http.StatusOK {
		t.Fatalf("plan failed: %d %s", planResult.Code, planResult.Body.String())
	}
	runBody, _ := json.Marshal(ConversionRunRequest{SessionID: session.ID})
	runRequest := httptest.NewRequest(http.MethodPost, "/api/conversion/run", bytes.NewReader(runBody))
	runRequest.Header.Set("X-MMV-Token", app.token)
	runResult := httptest.NewRecorder()
	mux.ServeHTTP(runResult, runRequest)
	if runResult.Code != http.StatusOK {
		t.Fatalf("run failed: %d %s", runResult.Code, runResult.Body.String())
	}
	var converted ConversionSession
	if err := json.Unmarshal(runResult.Body.Bytes(), &converted); err != nil {
		t.Fatal(err)
	}
	if len(converted.Outputs) < 2 {
		t.Fatalf("API conversion returned no target/proof outputs: %#v", converted.Outputs)
	}
	download := httptest.NewRequest(http.MethodGet, "/api/conversion/download?id="+converted.ID+"&kind=output&index=0&token="+app.token, nil)
	downloadResult := httptest.NewRecorder()
	mux.ServeHTTP(downloadResult, download)
	if downloadResult.Code != http.StatusOK || downloadResult.Body.Len() == 0 {
		t.Fatalf("download failed: %d %s", downloadResult.Code, downloadResult.Body.String())
	}
}

func TestJavaVanillaBundleAndBedrockFeatureMatrix(t *testing.T) {
	app := conversionTestApp(t)
	bpManifest := []byte(`{"format_version":2,"header":{"name":"Full Stack","description":"All Bedrock lanes","uuid":"31111111-1111-4111-8111-111111111111","version":[1,0,0],"min_engine_version":[1,26,30]},"modules":[{"type":"data","uuid":"32222222-2222-4222-8222-222222222222","version":[1,0,0]},{"type":"script","language":"javascript","uuid":"33333333-3333-4333-8333-333333333333","version":[1,0,0],"entry":"scripts/main.js"}]}`)
	rpManifest := []byte(`{"format_version":2,"header":{"name":"Full Stack RP","description":"Resources","uuid":"34444444-4444-4444-8444-444444444444","version":[1,0,0],"min_engine_version":[1,26,30]},"modules":[{"type":"resources","uuid":"35555555-5555-4555-8555-555555555555","version":[1,0,0]}]}`)
	archive := conversionTestArchive(t, map[string][]byte{
		"behavior_packs/Full/manifest.json":                     bpManifest,
		"behavior_packs/Full/recipes/test.json":                 []byte(`{"format_version":"1.21.0","minecraft:recipe_shapeless":{"description":{"identifier":"full:test"},"tags":["crafting_table"],"ingredients":[{"item":"minecraft:stone"}],"result":{"item":"minecraft:diamond","count":1}}}`),
		"behavior_packs/Full/camera_presets/follow.json":        []byte(`{"format_version":"1.21.0","minecraft:camera_preset":{"identifier":"full:follow","inherit_from":"minecraft:free"}}`),
		"behavior_packs/Full/dialogue/guide.json":               []byte(`{"minecraft:npc_dialogue":{"scenes":[]}}`),
		"behavior_packs/Full/trading/trades.json":               []byte(`{"tiers":[]}`),
		"behavior_packs/Full/volumes/zone.json":                 []byte(`{"format_version":"1.21.0"}`),
		"behavior_packs/Full/scripts/main.js":                   []byte(`import { world, system } from '@minecraft/server'; world.afterEvents.playerSpawn.subscribe(() => system.run(() => {}));`),
		"resource_packs/Full/manifest.json":                     rpManifest,
		"resource_packs/Full/textures/blocks/example.png":       []byte("png"),
		"resource_packs/Full/animations/example.animation.json": []byte(`{"format_version":"1.8.0","animations":{}}`),
	}, nil)
	session, err := app.importConversionFile("full-stack.mcaddon", bytes.NewReader(archive))
	if err != nil {
		t.Fatal(err)
	}
	plan, err := buildConversionPlan(session, ConversionTargetSpec{Format: "java-pack-bundle", GameVersion: "1.21.11", Name: "Full Stack Java", Namespace: "full_stack"})
	if err != nil {
		t.Fatal(err)
	}
	session.Plan = plan
	if err := executeConversion(session); err != nil {
		t.Fatal(err)
	}
	entries := zipEntryNames(t, filepath.Join(session.Paths.Outputs, session.Outputs[0].Name))
	for _, required := range []string{"PRODUCT-MANIFEST.json", "packages/Full Stack Java-datapack.zip", "packages/Full Stack Java-resourcepack.zip", "source/datapack/pack.mcmeta", "source/resourcepack/pack.mcmeta", "omnibridge/conversion-plan.json"} {
		if !anyContains(entries, required) {
			t.Fatalf("Java pack family missing %q: %v", required, entries)
		}
	}

	projectPlan, err := buildConversionPlan(session, ConversionTargetSpec{Format: "java-fabric", GameVersion: "1.21.11", Loader: "fabric", Name: "Full Stack Java Mod", Namespace: "full_stack"})
	if err != nil {
		t.Fatal(err)
	}
	session.Plan = projectPlan
	if err := executeConversion(session); err != nil {
		t.Fatal(err)
	}
	projectPath := filepath.Join(session.Paths.Outputs, session.Outputs[0].Name)
	entries = zipEntryNames(t, projectPath)
	for _, required := range []string{"GeneratedContentRegistry.java", "feature-matrix.json", "omnibridge/source/behavior_packs/Full/scripts/main.js", "omnibridge/contracts"} {
		if !anyContains(entries, required) {
			t.Fatalf("Java target-native feature workspace missing %q: %v", required, entries)
		}
	}
	reader, err := zip.OpenReader(projectPath)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	foundSurface := false
	for _, file := range reader.File {
		if strings.HasSuffix(file.Name, "GeneratedContentRegistry.java") {
			data, err := readZipFile(file, 2<<20)
			if err != nil {
				t.Fatal(err)
			}
			foundSurface = bytes.Contains(data, []byte("client camera controller")) && bytes.Contains(data, []byte("merchant offers")) && bytes.Contains(data, []byte("loader event bus"))
		}
	}
	if !foundSurface {
		t.Fatal("generated Java feature registry did not account for Bedrock camera/trade/script implementation surfaces")
	}
}

func TestRegolithAdapterUsesExactIsolatedExport(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake POSIX adapter fixture")
	}
	app := conversionTestApp(t)
	app.settings.ConversionToolPaths = map[string]string{}
	archive := conversionTestArchive(t, map[string][]byte{
		"behavior_packs/Demo/manifest.json": []byte(`{"format_version":2,"header":{"name":"Demo BP","description":"Demo","uuid":"41111111-1111-4111-8111-111111111111","version":[1,0,0],"min_engine_version":[1,26,30]},"modules":[{"type":"data","uuid":"42222222-2222-4222-8222-222222222222","version":[1,0,0]}]}`),
		"resource_packs/Demo/manifest.json": []byte(`{"format_version":2,"header":{"name":"Demo RP","description":"Demo","uuid":"43333333-3333-4333-8333-333333333333","version":[1,0,0],"min_engine_version":[1,26,30]},"modules":[{"type":"resources","uuid":"44444444-4444-4444-8444-444444444444","version":[1,0,0]}]}`),
	}, nil)
	session, err := app.importConversionFile("demo.mcaddon", bytes.NewReader(archive))
	if err != nil {
		t.Fatal(err)
	}
	plan, err := buildConversionPlan(session, ConversionTargetSpec{Format: "bedrock-project", GameVersion: "1.26.30", Name: "Regolith Product", Namespace: "regolith_product"})
	if err != nil {
		t.Fatal(err)
	}
	session.Plan = plan
	helper := filepath.Join(t.TempDir(), "fake-regolith")
	script := `#!/bin/sh
set -eu
[ "$1" = "run" ]
[ "$2" = "default" ]
grep -q '"target": "exact"' config.json
if grep -q '"target": "development"' config.json; then exit 42; fi
mkdir -p build/export/BP build/export/RP
cp -R packs/BP/. build/export/BP/
cp -R packs/RP/. build/export/RP/
`
	if err := os.WriteFile(helper, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	app.settings.ConversionToolPaths["regolith"] = helper
	var tool ConversionToolAdapter
	for _, candidate := range app.configuredConversionToolAdapters() {
		if candidate.ID == "regolith" {
			tool = candidate
		}
	}
	if !tool.Ready || !tool.CanExecute {
		t.Fatalf("fake Regolith was not detected: %#v", tool)
	}
	run, err := app.executeConversionAdapter(context.Background(), session, tool, nil)
	if err != nil {
		t.Fatal(err)
	}
	if run.State != "succeeded" || !run.SourceVerified || len(run.Outputs) != 1 || !run.Outputs[0].Validated {
		t.Fatalf("unexpected Regolith adapter result: %#v", run)
	}
	configData, err := os.ReadFile(filepath.Join(run.WorkingDir, "project", "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(configData, []byte(`"target": "exact"`)) || bytes.Contains(configData, []byte(`"target": "development"`)) {
		t.Fatalf("Regolith project did not retain isolated exact export: %s", configData)
	}
	if err := verifyConversionSource(session); err != nil {
		t.Fatalf("Regolith adapter changed source: %v", err)
	}
}
