package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func emitBedrockProjectTarget(session *ConversionSession) ([]string, error) {
	root := filepath.Join(session.Paths.Workspace, "bedrock-project")
	if err := buildBedrockProjectDirectory(session, root); err != nil {
		return nil, err
	}
	output := filepath.Join(session.Paths.Outputs, cleanConversionName(firstNonEmpty(session.Plan.Target.Name, session.Name))+"-bedrock-project.zip")
	if _, _, err := zipDirectoryDeterministic(root, output, nil); err != nil {
		return nil, err
	}
	return []string{output}, nil
}

func buildBedrockProjectDirectory(session *ConversionSession, root string) error {
	target := session.Plan.Target
	target.Format = "bedrock-addon"
	ids := bedrockPackIDs(session.Source.SHA256, target.Format, target.Namespace)
	bp := filepath.Join(root, "packs", "BP")
	rp := filepath.Join(root, "packs", "RP")
	for _, dir := range []string{bp, rp, filepath.Join(root, "packs", "data"), filepath.Join(root, "dist"), filepath.Join(root, "tests")} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	hasLogic := graphHasKinds(session.Graph, "java-bytecode", "java-source", "bedrock-script", "mixin", "block", "item", "entity", "camera-preset", "dialogue", "trade-table")
	if err := writeBedrockBehaviorManifest(bp, target, ids, true, hasLogic); err != nil {
		return err
	}
	if err := writeBedrockResourceManifest(rp, target, ids); err != nil {
		return err
	}
	if err := emitNodesToBedrock(session, bp, rp, "bedrock-addon"); err != nil {
		return err
	}
	if hasLogic {
		if err := writeBedrockScriptScaffold(session, bp); err != nil {
			return err
		}
	}
	config := map[string]any{
		"$schema": "https://raw.githubusercontent.com/Bedrock-OSS/regolith-schemas/main/config/v1.2.json",
		"author":  "Minecraft Mod Vault OmniBridge",
		"name":    target.Name,
		"packs": map[string]any{
			"behaviorPack": "./packs/BP",
			"resourcePack": "./packs/RP",
		},
		"regolith": map[string]any{
			"dataPath":          "./packs/data",
			"filterDefinitions": map[string]any{},
			"profiles": map[string]any{
				"default": map[string]any{
					"export": map[string]any{
						"readOnly": false,
						"target":   "exact",
						"bpPath":   "./build/export/BP",
						"rpPath":   "./build/export/RP",
					},
					"filters": []any{},
				},
			},
		},
	}
	if err := writeJSONFileAtomic(filepath.Join(root, "config.json"), config); err != nil {
		return err
	}
	acceptance := map[string]any{
		"schemaVersion": 1,
		"sourceSha256":  session.Source.SHA256,
		"targetVersion": target.GameVersion,
		"requiredGates": []string{"manifest-validation", "content-log-clean", "GameTest", "fresh-world", "existing-world", "multiplayer", "restart-persistence", "resource-fallback", "performance"},
		"reviewItems":   session.Plan.ReviewQueue,
	}
	if err := writeJSONFileAtomic(filepath.Join(root, "tests", "OMNIBRIDGE-ACCEPTANCE.json"), acceptance); err != nil {
		return err
	}
	if err := writeTargetContracts(session, filepath.Join(root, "omnibridge")); err != nil {
		return err
	}
	packStage := filepath.Join(root, ".package")
	if err := os.MkdirAll(packStage, 0o755); err != nil {
		return err
	}
	if err := copyDir(bp, filepath.Join(packStage, shortPackFolder(target.Namespace, "BP"))); err != nil {
		return err
	}
	if err := copyDir(rp, filepath.Join(packStage, shortPackFolder(target.Namespace, "RP"))); err != nil {
		return err
	}
	mcaddon := filepath.Join(root, "dist", cleanConversionName(target.Name)+".mcaddon")
	if _, _, err := zipDirectoryDeterministic(packStage, mcaddon, nil); err != nil {
		return err
	}
	_ = os.RemoveAll(packStage)
	readme := fmt.Sprintf("# %s — OmniBridge Bedrock source project\n\nTarget Bedrock: `%s`\nSource SHA-256: `%s`\n\n- `packs/BP` and `packs/RP` are editable source-of-truth packs.\n- `config.json` follows the Regolith v1.2 project schema.\n- `dist/%s.mcaddon` is the deterministic packaged add-on.\n- `omnibridge/contracts` records every semantic item that still requires implementation or gameplay validation.\n- `tests/OMNIBRIDGE-ACCEPTANCE.json` is the release gate, not decorative documentation.\n\nGenerated target files are structurally valid; semantic parity is not claimed until the review queue and test matrix are complete.\n", target.Name, target.GameVersion, session.Source.SHA256, cleanConversionName(target.Name))
	return os.WriteFile(filepath.Join(root, "README.md"), []byte(readme), 0o644)
}

func emitJavaPackBundleTarget(session *ConversionSession) ([]string, error) {
	target := session.Plan.Target
	root := filepath.Join(session.Paths.Workspace, "java-pack-bundle")
	dataRoot := filepath.Join(root, "source", "datapack")
	resourceRoot := filepath.Join(root, "source", "resourcepack")
	packages := filepath.Join(root, "packages")
	for _, dir := range []string{dataRoot, resourceRoot, packages} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}
	if err := writeJavaPackMetadata(dataRoot, target, "data"); err != nil {
		return nil, err
	}
	if err := writeJavaPackMetadata(resourceRoot, target, "resource"); err != nil {
		return nil, err
	}
	if err := emitNodesToJavaPack(session, dataRoot, "data"); err != nil {
		return nil, err
	}
	if err := emitNodesToJavaPack(session, resourceRoot, "resource"); err != nil {
		return nil, err
	}
	if err := writeTargetContracts(session, filepath.Join(root, "omnibridge")); err != nil {
		return nil, err
	}
	base := cleanConversionName(firstNonEmpty(target.Name, session.Name))
	dataArchive := filepath.Join(packages, base+"-datapack.zip")
	resourceArchive := filepath.Join(packages, base+"-resourcepack.zip")
	if _, _, err := zipDirectoryDeterministic(dataRoot, dataArchive, nil); err != nil {
		return nil, err
	}
	if _, _, err := zipDirectoryDeterministic(resourceRoot, resourceArchive, nil); err != nil {
		return nil, err
	}
	components := []map[string]any{}
	for _, component := range []struct {
		kind string
		path string
	}{
		{"java-datapack", dataArchive},
		{"java-resourcepack", resourceArchive},
	} {
		digest, size, err := fileSHA256(component.path)
		if err != nil {
			return nil, err
		}
		components = append(components, map[string]any{"kind": component.kind, "name": filepath.Base(component.path), "sha256": digest, "size": size})
	}
	worldRoot := findConversionWorldRoot(session.Paths.Extracted)
	worldState := "not-present"
	if worldRoot != "" {
		worldState = "source-preserved"
		worldArchive := filepath.Join(root, "source", "world-template-source.zip")
		if _, _, err := zipDirectoryDeterministic(worldRoot, worldArchive, nil); err != nil {
			return nil, err
		}
		digest, size, err := fileSHA256(worldArchive)
		if err != nil {
			return nil, err
		}
		components = append(components, map[string]any{"kind": "world-template-source", "name": filepath.Base(worldArchive), "sha256": digest, "size": size, "requiresAdapter": session.Source.Edition != "java"})
	}
	manifest := map[string]any{
		"schemaVersion": 1,
		"product":       "java-vanilla-addon-family",
		"name":          target.Name,
		"namespace":     target.Namespace,
		"targetVersion": target.GameVersion,
		"sourceSha256":  session.Source.SHA256,
		"worldState":    worldState,
		"components":    components,
		"reviewQueue":   session.Plan.ReviewQueue,
		"installOrder":  []string{"resource-pack", "data-pack", "optional-world-template"},
	}
	if err := writeJSONFileAtomic(filepath.Join(root, "PRODUCT-MANIFEST.json"), manifest); err != nil {
		return nil, err
	}
	readme := fmt.Sprintf("# %s — Java vanilla add-on family\n\nTarget Minecraft: `%s`\nSource SHA-256: `%s`\n\nThis product keeps the data pack and resource pack independently installable while preserving one shared conversion graph, review queue and checksum manifest. Install the resource pack globally or per-world, then place the data pack in the target world's `datapacks` folder. A source world/template archive is included when present; cross-edition chunks still require a verified world adapter.\n", target.Name, target.GameVersion, session.Source.SHA256)
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte(readme), 0o644); err != nil {
		return nil, err
	}
	output := filepath.Join(session.Paths.Outputs, base+"-java-pack-bundle.zip")
	if _, _, err := zipDirectoryDeterministic(root, output, nil); err != nil {
		return nil, err
	}
	return []string{output}, nil
}

func emitBedrockWorldProductTarget(session *ConversionSession) ([]string, error) {
	root := filepath.Join(session.Paths.Workspace, "bedrock-world-product")
	if err := os.MkdirAll(filepath.Join(root, "dist"), 0o755); err != nil {
		return nil, err
	}
	if err := buildBedrockProjectDirectory(session, filepath.Join(root, "companion-addon")); err != nil {
		return nil, err
	}
	worldRoot := findConversionWorldRoot(session.Paths.Extracted)
	worldState := "adapter-required"
	worldArtifact := ""
	if worldRoot != "" && session.Source.Edition == "bedrock" {
		stage := filepath.Join(root, "world-template-source")
		if err := copyDir(worldRoot, stage); err != nil {
			return nil, err
		}
		if err := writeBedrockWorldTemplateMetadata(stage, session); err != nil {
			return nil, err
		}
		worldArtifact = filepath.Join(root, "dist", cleanConversionName(session.Plan.Target.Name)+".mctemplate")
		if _, _, err := zipDirectoryDeterministic(stage, worldArtifact, nil); err != nil {
			return nil, err
		}
		worldState = "packaged"
	} else {
		adapter := filepath.Join(root, "world-adapter")
		if err := copyDir(session.Paths.Extracted, filepath.Join(adapter, "immutable-source")); err != nil {
			return nil, err
		}
		text := "# Cross-edition world lane\n\nThe source world is preserved, but a Bedrock `.mctemplate` is not emitted until Chunker/je2be/Amulet actually converts the chunks. Run the allowlisted adapter from OmniBridge, inspect warnings, and repackage the validated result.\n"
		if err := os.WriteFile(filepath.Join(adapter, "README.md"), []byte(text), 0o644); err != nil {
			return nil, err
		}
	}
	manifest := map[string]any{
		"schemaVersion": 1,
		"name":          session.Plan.Target.Name,
		"sourceSha256":  session.Source.SHA256,
		"targetEdition": "bedrock",
		"targetVersion": session.Plan.Target.GameVersion,
		"worldState":    worldState,
		"worldArtifact": filepath.Base(worldArtifact),
		"components":    []string{"world-template", "companion-addon", "editable-bedrock-project", "conversion-contracts", "verification-matrix"},
		"reviewQueue":   session.Plan.ReviewQueue,
	}
	if err := writeJSONFileAtomic(filepath.Join(root, "PRODUCT-MANIFEST.json"), manifest); err != nil {
		return nil, err
	}
	if err := writeTargetContracts(session, filepath.Join(root, "omnibridge")); err != nil {
		return nil, err
	}
	output := filepath.Join(session.Paths.Outputs, cleanConversionName(session.Plan.Target.Name)+"-bedrock-world-product.zip")
	if _, _, err := zipDirectoryDeterministic(root, output, nil); err != nil {
		return nil, err
	}
	return []string{output}, nil
}

func emitJavaWorldModTarget(session *ConversionSession) ([]string, error) {
	target := session.Plan.Target
	root := filepath.Join(session.Paths.Workspace, "java-world-mod")
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}
	if err := generateJavaProjectScaffold(session, root); err != nil {
		return nil, err
	}
	flavor := javaProjectFlavor(target)
	sourceRoot := filepath.Join(root, "src", "main")
	if flavor == "java-multiloader" {
		sourceRoot = filepath.Join(root, "common", "src", "main")
	}
	resources := filepath.Join(sourceRoot, "resources")
	if err := emitNodesToJavaProjectResources(session, resources); err != nil {
		return nil, err
	}
	worldRoot := findConversionWorldRoot(session.Paths.Extracted)
	archiveName := "world-template.zip"
	usable := session.Source.Edition == "java" && worldRoot != ""
	if session.Source.Edition == "bedrock" {
		archiveName = "source-world.mcworld"
	}
	archivePath := filepath.Join(resources, "omnibridge", archiveName)
	if worldRoot != "" {
		if err := os.MkdirAll(filepath.Dir(archivePath), 0o755); err != nil {
			return nil, err
		}
		if _, _, err := zipDirectoryDeterministic(worldRoot, archivePath, nil); err != nil {
			return nil, err
		}
	}
	packageName := "dev.omnibridge." + strings.ReplaceAll(sanitizeNamespace(target.Namespace), "-", "_")
	javaDir := filepath.Join(sourceRoot, "java", filepath.FromSlash(strings.ReplaceAll(packageName, ".", "/")))
	if err := writeWorldTemplateExporter(javaDir, packageName, "/omnibridge/"+archiveName, usable); err != nil {
		return nil, err
	}
	product := map[string]any{"schemaVersion": 1, "sourceSha256": session.Source.SHA256, "targetVersion": target.GameVersion, "loader": flavor, "embeddedArchive": archiveName, "directlyUsableByJava": usable, "adapterRequired": !usable, "reviewQueue": session.Plan.ReviewQueue}
	if err := writeJSONFileAtomic(filepath.Join(resources, "omnibridge", "world-product.json"), product); err != nil {
		return nil, err
	}
	if err := writeTargetContracts(session, filepath.Join(root, "omnibridge")); err != nil {
		return nil, err
	}
	output := filepath.Join(session.Paths.Outputs, cleanConversionName(target.Name)+"-java-world-mod-source.zip")
	if _, _, err := zipDirectoryDeterministic(root, output, nil); err != nil {
		return nil, err
	}
	return []string{output}, nil
}

func writeGeneratedJavaContentRegistry(javaDir, packageName string, graph UniversalContentGraph) error {
	type entry struct {
		ID, Kind, Name, Path, Level, Surface, Side, Strategy string
	}
	entries := make([]entry, 0, len(graph.Nodes))
	for _, node := range graph.Nodes {
		entries = append(entries, entry{
			ID: node.ID, Kind: node.Kind, Name: node.Name, Path: node.SourcePath,
			Level: string(node.Level), Surface: javaSurfaceForContentKind(node.Kind),
			Side: javaLogicalSideForContentKind(node.Kind), Strategy: conversionStrategyForLevel(node.Level),
		})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].ID < entries[j].ID })
	var builder strings.Builder
	builder.WriteString("package " + packageName + ";\n\n")
	builder.WriteString("import java.util.List;\n\n")
	builder.WriteString("/** Immutable target-native implementation inventory generated from the Universal Minecraft Content Graph. */\n")
	builder.WriteString("public final class GeneratedContentRegistry {\n")
	builder.WriteString("    public record ContentContract(String id, String kind, String name, String sourcePath, String conversionLevel, String targetSurface, String logicalSide, String strategy) {}\n")
	builder.WriteString("    public static final List<ContentContract> ALL = List.of(\n")
	for index, item := range entries {
		comma := ","
		if index == len(entries)-1 {
			comma = ""
		}
		builder.WriteString(fmt.Sprintf("        new ContentContract(%s, %s, %s, %s, %s, %s, %s, %s)%s\n", javaLiteral(item.ID), javaLiteral(item.Kind), javaLiteral(item.Name), javaLiteral(item.Path), javaLiteral(item.Level), javaLiteral(item.Surface), javaLiteral(item.Side), javaLiteral(item.Strategy), comma))
	}
	builder.WriteString("    );\n    private GeneratedContentRegistry() {}\n}\n")
	return os.WriteFile(filepath.Join(javaDir, "GeneratedContentRegistry.java"), []byte(builder.String()), 0o644)
}

func javaSurfaceForContentKind(kind string) string {
	switch kind {
	case "block":
		return "registry/block + block-item + block-entity/data-components"
	case "item":
		return "registry/item + data-components + item-model"
	case "entity":
		return "registry/entity-type + attributes + goals + renderer"
	case "recipe":
		return "data/recipe + serializer when custom"
	case "loot-table":
		return "data/loot_table + loot-function/condition adapters"
	case "function":
		return "server command function + command migration tests"
	case "tag":
		return "data/tag"
	case "advancement":
		return "data/advancement + trigger adapter"
	case "predicate":
		return "data/predicate"
	case "item-modifier":
		return "data/item_modifier"
	case "texture", "pack-icon":
		return "assets/texture-atlas + model references"
	case "model", "attachable", "render-controller":
		return "client model-layer + renderer + item/block model"
	case "sound":
		return "registry/sound-event + sounds.json"
	case "language":
		return "assets/lang"
	case "font":
		return "assets/font-provider"
	case "particle":
		return "registry/particle-type + client particle provider"
	case "animation", "animation-controller":
		return "client animation state machine / animation-library adapter"
	case "spawn-rule":
		return "spawn-placement + biome-modification"
	case "feature", "feature-rule", "worldgen":
		return "configured-feature + placed-feature + biome-modifier + data generation"
	case "biome":
		return "biome registry/data + climate/spawn/effects adapters"
	case "dimension":
		return "dimension-type + level-stem + chunk-generator"
	case "camera-preset":
		return "client camera controller"
	case "dialogue":
		return "menu + screen + dialogue service + localization"
	case "trade-table":
		return "merchant offers + villager/merchant trade registration"
	case "volume":
		return "region trigger + server tick/event service"
	case "fog":
		return "client fog renderer + biome/dimension effects"
	case "ui":
		return "menu + screen + HUD renderer"
	case "shader", "material":
		return "client rendering pipeline / shader resource adapter"
	case "bedrock-script":
		return "loader event bus + tick scheduler + commands + networking + persistence"
	case "java-source":
		return "loader API migration workspace"
	case "java-bytecode", "mixin":
		return "source recovery + remap/decompile + loader API reimplementation"
	case "structure":
		return "structure NBT + processor/placement adapter"
	case "world-data":
		return "world converter adapter + embedded template/exporter"
	default:
		return "target-native review contract"
	}
}

func javaLogicalSideForContentKind(kind string) string {
	switch kind {
	case "texture", "pack-icon", "model", "attachable", "render-controller", "sound", "language", "font", "particle", "animation", "animation-controller", "camera-preset", "fog", "ui", "shader", "material":
		return "client"
	case "recipe", "loot-table", "function", "tag", "advancement", "predicate", "item-modifier", "spawn-rule", "feature", "feature-rule", "worldgen", "biome", "dimension", "trade-table", "volume", "structure", "world-data":
		return "server"
	default:
		return "common"
	}
}

func conversionStrategyForLevel(level ConversionLevel) string {
	switch level {
	case ConversionExact:
		return "preserve-and-reindex"
	case ConversionTranslated:
		return "deterministic-schema-translation"
	case ConversionGenerated:
		return "generated-target-native-adapter"
	case ConversionToolAssisted:
		return "verified-specialist-adapter"
	case ConversionReview:
		return "semantic-implementation-required"
	case ConversionBlocked:
		return "target-specific-reimplementation-required"
	default:
		return "evidence-review"
	}
}

func writeJavaFeatureMatrix(resources, namespace string, graph UniversalContentGraph, target ConversionTargetSpec) error {
	rows := make([]map[string]any, 0, len(graph.Nodes))
	for _, node := range graph.Nodes {
		rows = append(rows, map[string]any{
			"id": node.ID, "kind": node.Kind, "name": node.Name, "sourcePath": node.SourcePath,
			"sourceFormat": node.SourceFormat, "conversionLevel": node.Level,
			"targetSurface":  javaSurfaceForContentKind(node.Kind),
			"logicalSide":    javaLogicalSideForContentKind(node.Kind),
			"strategy":       conversionStrategyForLevel(node.Level),
			"requiresReview": node.RequiresReview,
		})
	}
	sort.Slice(rows, func(i, j int) bool { return fmt.Sprint(rows[i]["id"]) < fmt.Sprint(rows[j]["id"]) })
	value := map[string]any{
		"schemaVersion": 1, "target": target, "nodeCount": len(rows),
		"completionBoundary": "Each row names the concrete Java implementation surface. Generated or review rows are not semantic parity until implemented and runtime-tested.",
		"features":           rows,
	}
	return writeJSONFileAtomic(filepath.Join(resources, "assets", sanitizeNamespace(namespace), "omnibridge", "feature-matrix.json"), value)
}

func writeWorldTemplateExporter(javaDir, packageName, resource string, usable bool) error {
	if err := os.MkdirAll(javaDir, 0o755); err != nil {
		return err
	}
	guard := ""
	if !usable {
		guard = "        throw new IllegalStateException(\"The embedded source world is not Java Edition. Run the recorded world adapter before exporting it as a Java world.\");\n"
	}
	code := fmt.Sprintf(`package %s;

import java.io.IOException;
import java.io.InputStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.StandardCopyOption;
import java.util.zip.ZipEntry;
import java.util.zip.ZipInputStream;

/** Safely exports the immutable world/template bundled by OmniBridge. */
public final class WorldTemplateExporter {
    private static final String RESOURCE = %s;
    private WorldTemplateExporter() {}

    public static void exportTo(Path destination) throws IOException {
%s        Files.createDirectories(destination);
        try (InputStream raw = WorldTemplateExporter.class.getResourceAsStream(RESOURCE)) {
            if (raw == null) throw new IOException("Bundled world archive is missing: " + RESOURCE);
            try (ZipInputStream zip = new ZipInputStream(raw)) {
                ZipEntry entry;
                while ((entry = zip.getNextEntry()) != null) {
                    Path output = destination.resolve(entry.getName()).normalize();
                    if (!output.startsWith(destination.normalize())) throw new IOException("Unsafe world entry: " + entry.getName());
                    if (entry.isDirectory()) Files.createDirectories(output);
                    else {
                        Files.createDirectories(output.getParent());
                        Files.copy(zip, output, StandardCopyOption.REPLACE_EXISTING);
                    }
                }
            }
        }
    }
}
`, packageName, javaLiteral(resource), guard)
	return os.WriteFile(filepath.Join(javaDir, "WorldTemplateExporter.java"), []byte(code), 0o644)
}

func javaLiteral(value string) string {
	encoded, _ := json.Marshal(value)
	return string(encoded)
}

func validateGeneratedProductManifest(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if !json.Valid(data) {
		return errors.New("generated product manifest is invalid JSON")
	}
	return nil
}
