package main

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

func executeConversion(session *ConversionSession) error {
	if session == nil || session.Plan == nil {
		return errors.New("conversion plan is required")
	}
	if err := verifyConversionSource(session); err != nil {
		return err
	}
	if err := os.RemoveAll(session.Paths.Workspace); err != nil {
		return err
	}
	if err := os.RemoveAll(session.Paths.Outputs); err != nil {
		return err
	}
	if err := os.MkdirAll(session.Paths.Workspace, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(session.Paths.Outputs, 0o755); err != nil {
		return err
	}
	session.Outputs = nil
	session.State, session.Phase, session.LastError = "converting", "emit-target", ""
	var outputs []string
	var err error
	switch session.Plan.Target.Format {
	case "bedrock-addon", "bedrock-behavior", "bedrock-resource":
		outputs, err = emitBedrockTarget(session)
	case "bedrock-project":
		outputs, err = emitBedrockProjectTarget(session)
	case "bedrock-world-product":
		outputs, err = emitBedrockWorldProductTarget(session)
	case "java-datapack", "java-resourcepack":
		outputs, err = emitJavaPackTarget(session)
	case "java-pack-bundle":
		outputs, err = emitJavaPackBundleTarget(session)
	case "java-fabric", "java-neoforge", "java-forge", "java-multiloader":
		outputs, err = emitJavaProjectTarget(session)
	case "java-world-mod":
		outputs, err = emitJavaWorldModTarget(session)
	case "bedrock-world", "bedrock-template", "java-world":
		outputs, err = emitWorldTarget(session)
	case "universal-bundle":
		outputs, err = emitUniversalBundle(session)
	default:
		err = fmt.Errorf("unsupported conversion target %q", session.Plan.Target.Format)
	}
	if err != nil {
		session.State, session.Phase, session.LastError = "failed", "emit-target", err.Error()
		return err
	}
	for _, path := range outputs {
		output, outputErr := conversionOutputRecord(path, len(session.Outputs))
		if outputErr != nil {
			return outputErr
		}
		output.Validated, output.Validation = validateConversionOutput(path, session.Plan.Target.Format)
		session.Outputs = append(session.Outputs, output)
	}
	if err := writeConversionProofBundle(session); err != nil {
		return err
	}
	proofPath := filepath.Join(session.Paths.Outputs, cleanConversionName(session.Name)+"-conversion-proof.zip")
	if output, outputErr := conversionOutputRecord(proofPath, len(session.Outputs)); outputErr == nil {
		output.Validated, output.Validation = validateConversionOutput(proofPath, "proof")
		session.Outputs = append(session.Outputs, output)
	}
	session.State = "converted"
	session.Phase = "validated"
	for _, output := range session.Outputs {
		if !output.Validated {
			session.State = "review-required"
			break
		}
	}
	if len(session.Plan.ReviewQueue) > 0 {
		session.State = "review-required"
	}
	return nil
}

func verifyConversionSource(session *ConversionSession) error {
	digest, size, err := hashFileSHA256(session.Source.Path)
	if err != nil {
		return fmt.Errorf("verify immutable source: %w", err)
	}
	if digest != session.Source.SHA256 || size != session.Source.Size {
		return errors.New("immutable conversion source hash or size changed")
	}
	tree, files, bytesCount, err := hashDirectoryTree(session.Paths.Extracted)
	if err != nil {
		return fmt.Errorf("verify extracted source: %w", err)
	}
	if tree != session.Source.TreeSHA256 || files != session.Source.FileCount || bytesCount != session.Source.ExtractedBytes {
		return errors.New("extracted conversion source changed after profiling")
	}
	return nil
}

func emitBedrockTarget(session *ConversionSession) ([]string, error) {
	target := session.Plan.Target
	base := cleanConversionName(firstNonEmpty(target.Name, session.Name))
	namespace := sanitizeNamespace(target.Namespace)
	stage := filepath.Join(session.Paths.Workspace, "bedrock")
	bp := filepath.Join(stage, shortPackFolder(namespace, "BP"))
	rp := filepath.Join(stage, shortPackFolder(namespace, "RP"))
	if err := os.MkdirAll(bp, 0o755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(rp, 0o755); err != nil {
		return nil, err
	}
	ids := bedrockPackIDs(session.Source.SHA256, target.Format, namespace)
	hasLogic := graphHasKinds(session.Graph, "java-bytecode", "java-source", "bedrock-script", "mixin")
	if target.Format != "bedrock-resource" {
		if err := writeBedrockBehaviorManifest(bp, target, ids, target.Format != "bedrock-behavior", hasLogic); err != nil {
			return nil, err
		}
	}
	if target.Format != "bedrock-behavior" {
		if err := writeBedrockResourceManifest(rp, target, ids); err != nil {
			return nil, err
		}
	}
	if err := emitNodesToBedrock(session, bp, rp, target.Format); err != nil {
		return nil, err
	}
	if hasLogic && target.Format != "bedrock-resource" {
		if err := writeBedrockScriptScaffold(session, bp); err != nil {
			return nil, err
		}
	}
	if err := writeTargetContracts(session, filepath.Join(stage, "OMNIBRIDGE")); err != nil {
		return nil, err
	}
	outputs := []string{}
	switch target.Format {
	case "bedrock-addon":
		output := filepath.Join(session.Paths.Outputs, base+".mcaddon")
		if _, _, err := zipDirectoryDeterministic(stage, output, nil); err != nil {
			return nil, err
		}
		outputs = append(outputs, output)
	case "bedrock-behavior":
		output := filepath.Join(session.Paths.Outputs, base+"-behavior.mcpack")
		if _, _, err := zipDirectoryDeterministic(bp, output, nil); err != nil {
			return nil, err
		}
		outputs = append(outputs, output)
	case "bedrock-resource":
		output := filepath.Join(session.Paths.Outputs, base+"-resources.mcpack")
		if _, _, err := zipDirectoryDeterministic(rp, output, nil); err != nil {
			return nil, err
		}
		outputs = append(outputs, output)
	}
	return outputs, nil
}

type bedrockIDs struct {
	BPHeader, BPData, BPScript, RPHeader, RPModule uuid.UUID
}

func bedrockPackIDs(sourceSHA, target, namespace string) bedrockIDs {
	ns := uuid.MustParse("f332b10c-6378-4c17-8b9a-893c5c93de42")
	makeID := func(role string) uuid.UUID {
		return uuid.NewSHA1(ns, []byte(sourceSHA+"\x00"+target+"\x00"+namespace+"\x00"+role))
	}
	return bedrockIDs{makeID("bp-header"), makeID("bp-data"), makeID("bp-script"), makeID("rp-header"), makeID("rp-module")}
}

func writeBedrockBehaviorManifest(root string, target ConversionTargetSpec, ids bedrockIDs, linkResource, script bool) error {
	engine := bedrockVersionArray(target.GameVersion)
	modules := []map[string]any{{"description": target.Name + " behavior data", "type": "data", "uuid": ids.BPData.String(), "version": []int{1, 0, 0}}}
	dependencies := []map[string]any{}
	if linkResource {
		dependencies = append(dependencies, map[string]any{"uuid": ids.RPHeader.String(), "version": []int{1, 0, 0}})
	}
	if script {
		modules = append(modules, map[string]any{"description": target.Name + " generated Script API bridge", "type": "script", "language": "javascript", "entry": "scripts/main.js", "uuid": ids.BPScript.String(), "version": []int{1, 0, 0}})
		dependencies = append(dependencies, map[string]any{"module_name": "@minecraft/server", "version": "2.8.0"})
	}
	manifest := map[string]any{"format_version": 2, "header": map[string]any{"name": target.Name + " BP", "description": firstNonEmpty(target.Description, "Generated by Minecraft Mod Vault OmniBridge"), "uuid": ids.BPHeader.String(), "version": []int{1, 0, 0}, "min_engine_version": engine}, "modules": modules, "metadata": map[string]any{"authors": []string{"Minecraft Mod Vault OmniBridge"}, "generated_with": map[string]any{"minecraft_mod_vault": []string{appVersion}}}}
	if len(dependencies) > 0 {
		manifest["dependencies"] = dependencies
	}
	return writeJSONFileAtomic(filepath.Join(root, "manifest.json"), manifest)
}

func writeBedrockResourceManifest(root string, target ConversionTargetSpec, ids bedrockIDs) error {
	manifest := map[string]any{"format_version": 2, "header": map[string]any{"name": target.Name + " RP", "description": firstNonEmpty(target.Description, "Generated by Minecraft Mod Vault OmniBridge"), "uuid": ids.RPHeader.String(), "version": []int{1, 0, 0}, "min_engine_version": bedrockVersionArray(target.GameVersion)}, "modules": []map[string]any{{"description": target.Name + " resources", "type": "resources", "uuid": ids.RPModule.String(), "version": []int{1, 0, 0}}}, "metadata": map[string]any{"authors": []string{"Minecraft Mod Vault OmniBridge"}, "generated_with": map[string]any{"minecraft_mod_vault": []string{appVersion}}}}
	return writeJSONFileAtomic(filepath.Join(root, "manifest.json"), manifest)
}

func emitNodesToBedrock(session *ConversionSession, bp, rp, targetFormat string) error {
	itemTextures := map[string]any{}
	terrainTextures := map[string]any{}
	for _, node := range session.Graph.Nodes {
		source := conversionNodeSourcePath(session, node)
		switch node.Kind {
		case "pack-icon":
			if source != "" {
				if targetFormat != "bedrock-behavior" {
					_ = copyFileReplace(source, filepath.Join(rp, "pack_icon.png"))
				}
				if targetFormat != "bedrock-resource" {
					_ = copyFileReplace(source, filepath.Join(bp, "pack_icon.png"))
				}
			}
		case "texture":
			if source == "" || targetFormat == "bedrock-behavior" {
				continue
			}
			destination, atlas, key := bedrockTextureDestination(node, source, rp)
			if err := copyFileReplace(source, destination); err != nil {
				return err
			}
			texturePath, _ := filepath.Rel(rp, strings.TrimSuffix(destination, filepath.Ext(destination)))
			entry := map[string]any{"textures": filepath.ToSlash(texturePath)}
			if atlas == "items" {
				itemTextures[key] = entry
			} else if atlas == "terrain" {
				terrainTextures[key] = entry
			}
		case "sound":
			if source != "" && targetFormat != "bedrock-behavior" {
				rel := bedrockAssetTail(node.SourcePath, "sounds")
				if rel == "" {
					rel = filepath.Base(source)
				}
				if err := copyFileReplace(source, filepath.Join(rp, "sounds", filepath.FromSlash(rel))); err != nil {
					return err
				}
			}
		case "language":
			if source != "" && targetFormat != "bedrock-behavior" {
				if err := convertLanguageToBedrock(source, rp); err != nil {
					return err
				}
			}
		case "recipe":
			if source != "" && targetFormat != "bedrock-resource" {
				if recipe, ok := translateRecipeToBedrock(node, source, session.Plan.Target.Namespace); ok {
					name := safeNodeFileName(node.Name) + ".json"
					if err := writeJSONFileAtomic(filepath.Join(bp, "recipes", name), recipe); err != nil {
						return err
					}
				} else {
					_ = copyReviewSource(source, bp, node)
				}
			}
		case "function":
			if source != "" && targetFormat != "bedrock-resource" {
				_ = copyReviewSource(source, bp, node)
			}
		case "block", "item", "entity", "animation", "animation-controller", "particle", "attachable", "spawn-rule", "feature", "feature-rule", "biome", "fog":
			if source != "" {
				destRoot := bp
				if node.Kind == "animation" || node.Kind == "animation-controller" || node.Kind == "particle" || node.Kind == "attachable" || node.Kind == "fog" {
					destRoot = rp
				}
				if targetFormat == "bedrock-resource" && destRoot == bp || targetFormat == "bedrock-behavior" && destRoot == rp {
					continue
				}
				rel := bedrockNativeRelative(node)
				if rel == "" {
					_ = copyReviewSource(source, destRoot, node)
				} else if err := copyFileReplace(source, filepath.Join(destRoot, filepath.FromSlash(rel))); err != nil {
					return err
				}
			}
		case "model", "loot-table", "tag", "advancement", "predicate", "item-modifier", "worldgen", "structure", "dimension", "java-source", "java-bytecode", "mixin", "shader", "render-controller", "ui", "material", "world-data":
			if source != "" {
				destRoot := bp
				if node.Kind == "model" || node.Kind == "shader" || node.Kind == "render-controller" || node.Kind == "ui" || node.Kind == "material" {
					destRoot = rp
				}
				_ = copyReviewSource(source, destRoot, node)
			}
		}
	}
	if targetFormat != "bedrock-behavior" {
		if len(itemTextures) > 0 {
			value := map[string]any{"resource_pack_name": session.Plan.Target.Name, "texture_name": "atlas.items", "texture_data": itemTextures}
			if err := writeJSONFileAtomic(filepath.Join(rp, "textures", "item_texture.json"), value); err != nil {
				return err
			}
		}
		if len(terrainTextures) > 0 {
			value := map[string]any{"resource_pack_name": session.Plan.Target.Name, "texture_name": "atlas.terrain", "texture_data": terrainTextures}
			if err := writeJSONFileAtomic(filepath.Join(rp, "textures", "terrain_texture.json"), value); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeBedrockScriptScaffold(session *ConversionSession, bp string) error {
	var builder strings.Builder
	builder.WriteString("// Generated by Minecraft Mod Vault OmniBridge 0.11.0.\n")
	builder.WriteString("// This file is an executable scaffold, not a claim that Java bytecode was automatically reimplemented.\n")
	builder.WriteString("import { world } from '@minecraft/server';\n\n")
	builder.WriteString("const contracts = ")
	contracts := []map[string]string{}
	for _, node := range session.Graph.Nodes {
		if node.Kind == "java-bytecode" || node.Kind == "java-source" || node.Kind == "mixin" || node.Kind == "bedrock-script" {
			contracts = append(contracts, map[string]string{"id": node.ID, "kind": node.Kind, "name": node.Name, "source": node.SourcePath})
		}
	}
	encoded, _ := json.MarshalIndent(contracts, "", "  ")
	builder.Write(encoded)
	builder.WriteString(";\n\nworld.afterEvents.worldLoad.subscribe(() => {\n  console.warn(`[OmniBridge] Loaded ${contracts.length} generated behavior contracts. Review CONVERSION-REPORT.md before distribution.`);\n});\n")
	if err := os.MkdirAll(filepath.Join(bp, "scripts"), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(bp, "scripts", "main.js"), []byte(builder.String()), 0o644); err != nil {
		return err
	}
	return writeJSONFileAtomic(filepath.Join(bp, "omnibridge_behavior_contracts.json"), contracts)
}

func emitJavaPackTarget(session *ConversionSession) ([]string, error) {
	target := session.Plan.Target
	base := cleanConversionName(firstNonEmpty(target.Name, session.Name))
	root := filepath.Join(session.Paths.Workspace, target.Format)
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}
	kind := "data"
	if target.Format == "java-resourcepack" {
		kind = "resource"
	}
	if err := writeJavaPackMetadata(root, target, kind); err != nil {
		return nil, err
	}
	if err := emitNodesToJavaPack(session, root, kind); err != nil {
		return nil, err
	}
	if err := writeTargetContracts(session, filepath.Join(root, "OMNIBRIDGE")); err != nil {
		return nil, err
	}
	name := base + "-datapack.zip"
	if kind == "resource" {
		name = base + "-resourcepack.zip"
	}
	output := filepath.Join(session.Paths.Outputs, name)
	if _, _, err := zipDirectoryDeterministic(root, output, nil); err != nil {
		return nil, err
	}
	return []string{output}, nil
}

func writeJavaPackMetadata(root string, target ConversionTargetSpec, kind string) error {
	format := javaPackFormat(target.GameVersion, kind)
	pack := map[string]any{"description": firstNonEmpty(target.Description, target.Name+" — converted by Minecraft Mod Vault OmniBridge")}
	if isModernDecimalPackFormat(target.GameVersion) {
		pack["min_format"] = format
		pack["max_format"] = format
	} else {
		pack["pack_format"] = int(format)
	}
	return writeJSONFileAtomic(filepath.Join(root, "pack.mcmeta"), map[string]any{"pack": pack})
}

func emitNodesToJavaPack(session *ConversionSession, root, packKind string) error {
	for _, node := range session.Graph.Nodes {
		source := conversionNodeSourcePath(session, node)
		if source == "" {
			continue
		}
		if packKind == "resource" {
			switch node.Kind {
			case "texture", "model", "sound", "particle", "font", "shader", "language", "pack-icon":
			default:
				continue
			}
			if node.Kind == "pack-icon" {
				_ = copyFileReplace(source, filepath.Join(root, "pack.png"))
				continue
			}
			if node.Kind == "language" {
				if err := convertLanguageToJava(source, root, session.Plan.Target.Namespace); err != nil {
					return err
				}
				continue
			}
			rel := javaResourceDestination(node, session.Plan.Target.Namespace)
			if err := copyFileReplace(source, filepath.Join(root, filepath.FromSlash(rel))); err != nil {
				return err
			}
			continue
		}
		switch node.Kind {
		case "recipe":
			if translated, ok := translateRecipeToJava(node, source, session.Plan.Target.Namespace); ok {
				rel := javaDataFolder(session.Plan.Target.GameVersion, "recipe")
				if err := writeJSONFileAtomic(filepath.Join(root, "data", session.Plan.Target.Namespace, rel, safeNodeFileName(node.Name)+".json"), translated); err != nil {
					return err
				}
			} else {
				_ = copyReviewSource(source, root, node)
			}
		case "function":
			rel := javaDataFolder(session.Plan.Target.GameVersion, "function")
			if session.Source.Edition == "java" {
				if err := copyFileReplace(source, filepath.Join(root, "data", firstNonEmpty(node.Namespace, session.Plan.Target.Namespace), rel, safeNodeFileName(node.Name)+".mcfunction")); err != nil {
					return err
				}
			} else {
				_ = copyReviewSource(source, root, node)
			}
		case "loot-table", "tag", "advancement", "predicate", "item-modifier", "worldgen", "structure", "dimension":
			if session.Source.Edition == "java" {
				rel := javaDataFolder(session.Plan.Target.GameVersion, node.Kind)
				if rel != "" {
					if err := copyFileReplace(source, filepath.Join(root, "data", firstNonEmpty(node.Namespace, session.Plan.Target.Namespace), rel, safeNodeFileName(node.Name)+filepath.Ext(source))); err != nil {
						return err
					}
				}
			} else {
				_ = copyReviewSource(source, root, node)
			}
		}
	}
	return nil
}

func emitJavaProjectTarget(session *ConversionSession) ([]string, error) {
	target := session.Plan.Target
	base := cleanConversionName(firstNonEmpty(target.Name, session.Name))
	root := filepath.Join(session.Paths.Workspace, target.Format)
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}
	if err := generateJavaProjectScaffold(session, root); err != nil {
		return nil, err
	}
	resources := filepath.Join(root, "src", "main", "resources")
	if javaProjectFlavor(target) == "java-multiloader" {
		resources = filepath.Join(root, "common", "src", "main", "resources")
	}
	if err := os.MkdirAll(resources, 0o755); err != nil {
		return nil, err
	}
	if err := emitNodesToJavaProjectResources(session, resources); err != nil {
		return nil, err
	}
	if err := writeTargetContracts(session, filepath.Join(root, "omnibridge")); err != nil {
		return nil, err
	}
	output := filepath.Join(session.Paths.Outputs, base+"-"+target.Format+"-source.zip")
	if _, _, err := zipDirectoryDeterministic(root, output, nil); err != nil {
		return nil, err
	}
	return []string{output}, nil
}

func generateJavaProjectScaffold(session *ConversionSession, root string) error {
	target := session.Plan.Target
	atlas, _ := loadRepairVersionAtlas()
	resolution := AtlasResolution{GameVersion: target.GameVersion, Loader: target.Loader, JavaMajor: targetJavaForMinecraft(target.GameVersion)}
	if atlas != nil {
		resolution = atlas.Resolve(target.GameVersion, target.Loader)
	}
	if resolution.JavaMajor == 0 {
		resolution.JavaMajor = 21
	}
	name := cleanConversionName(target.Name)
	modID := sanitizeNamespace(target.Namespace)
	group := "dev.omnibridge." + strings.ReplaceAll(modID, "-", "_")
	packageName := strings.ReplaceAll(group, "-", "_")
	className := javaClassName(name)
	flavor := javaProjectFlavor(target)
	properties := fmt.Sprintf("org.gradle.jvmargs=-Xmx2G\norg.gradle.parallel=true\nminecraft_version=%s\nmod_version=1.0.0\nmaven_group=%s\narchives_base_name=%s\n", target.GameVersion, group, modID)
	if err := os.WriteFile(filepath.Join(root, "gradle.properties"), []byte(properties), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(root, "settings.gradle.kts"), []byte(javaSettingsGradle(name, flavor)), 0o644); err != nil {
		return err
	}
	buildTarget := target
	buildTarget.Format = flavor
	if err := os.WriteFile(filepath.Join(root, "build.gradle.kts"), []byte(javaBuildGradle(buildTarget, resolution, group)), 0o644); err != nil {
		return err
	}
	sourceRoot := filepath.Join(root, "src", "main")
	if flavor == "java-multiloader" {
		sourceRoot = filepath.Join(root, "common", "src", "main")
		for _, module := range []string{"fabric", "neoforge", "forge"} {
			if err := os.MkdirAll(filepath.Join(root, module, "src", "main", "java"), 0o755); err != nil {
				return err
			}
		}
	}
	javaDir := filepath.Join(sourceRoot, "java", filepath.FromSlash(strings.ReplaceAll(packageName, ".", "/")))
	if err := os.MkdirAll(javaDir, 0o755); err != nil {
		return err
	}
	javaSource := generatedJavaEntrypoint(buildTarget, packageName, className, modID, session.Graph)
	if err := os.WriteFile(filepath.Join(javaDir, className+".java"), []byte(javaSource), 0o644); err != nil {
		return err
	}
	resources := filepath.Join(sourceRoot, "resources")
	if err := os.MkdirAll(resources, 0o755); err != nil {
		return err
	}
	if err := writeJavaLoaderMetadata(resources, buildTarget, packageName+"."+className, modID); err != nil {
		return err
	}
	if err := writeGeneratedJavaContentRegistry(javaDir, packageName, session.Graph); err != nil {
		return err
	}
	readme := generatedJavaProjectReadme(session, resolution)
	return os.WriteFile(filepath.Join(root, "README.md"), []byte(readme), 0o644)
}

func javaProjectFlavor(target ConversionTargetSpec) string {
	if target.Format != "java-world-mod" {
		return target.Format
	}
	switch strings.ToLower(strings.TrimSpace(target.Loader)) {
	case "neoforge":
		return "java-neoforge"
	case "forge":
		return "java-forge"
	case "multiloader":
		return "java-multiloader"
	default:
		return "java-fabric"
	}
}

func javaSettingsGradle(name, target string) string {
	pluginRepos := "pluginManagement { repositories { gradlePluginPortal(); mavenCentral(); maven(\"https://maven.fabricmc.net/\"); maven(\"https://maven.neoforged.net/releases\"); maven(\"https://maven.minecraftforge.net/\"); maven(\"https://maven.kikugie.dev/releases\"); maven(\"https://maven.isxander.dev/releases\") } }\n"
	if target == "java-multiloader" {
		return pluginRepos + fmt.Sprintf("rootProject.name = %q\ninclude(\"common\", \"fabric\", \"neoforge\", \"forge\")\n", name)
	}
	return pluginRepos + fmt.Sprintf("rootProject.name = %q\n", name)
}

func javaBuildGradle(target ConversionTargetSpec, resolution AtlasResolution, group string) string {
	javaMajor := resolution.JavaMajor
	if javaMajor == 0 {
		javaMajor = 21
	}
	base := fmt.Sprintf("group = %q\nversion = providers.gradleProperty(\"mod_version\").get()\n\njava { toolchain.languageVersion.set(JavaLanguageVersion.of(%d)) }\n\nrepositories { mavenCentral() }\n", group, javaMajor)
	switch target.Format {
	case "java-fabric":
		loom := atlasChoiceVersion(resolution.BuildToolchains, "fabric-loom", "1.10-SNAPSHOT")
		yarn := atlasChoiceVersion(resolution.Mappings, "yarn", target.GameVersion+"+build.1")
		loader := firstNonEmpty(resolution.LoaderVersion, "0.16.14")
		return fmt.Sprintf("plugins { id(\"fabric-loom\") version %q; `maven-publish`; java }\n%s\ndependencies { minecraft(\"com.mojang:minecraft:%s\"); mappings(\"net.fabricmc:yarn:%s:v2\"); modImplementation(\"net.fabricmc:fabric-loader:%s\") }\n\ntasks.processResources { filesMatching(\"fabric.mod.json\") { expand(mapOf(\"version\" to project.version, \"minecraft\" to %q, \"loader\" to %q)) } }\n", loom, base, target.GameVersion, yarn, loader, target.GameVersion, loader)
	case "java-neoforge":
		plugin := atlasChoiceVersion(resolution.BuildToolchains, "moddevgradle", "2.0.78")
		neo := firstNonEmpty(resolution.LoaderVersion, "21.1.200")
		return fmt.Sprintf("plugins { id(\"net.neoforged.moddev\") version %q; `maven-publish`; java }\n%s\nneoForge { version = %q }\n", plugin, base, neo)
	case "java-forge":
		forge := firstNonEmpty(resolution.LoaderVersion, target.GameVersion+"-47.4.0")
		return fmt.Sprintf("plugins { id(\"net.minecraftforge.gradle\") version \"[6.0,6.2)\"; `maven-publish`; java }\n%s\nminecraft { mappings(\"official\", %q) }\ndependencies { minecraft(\"net.minecraftforge:forge:%s\") }\n", base, target.GameVersion, forge)
	case "java-multiloader":
		return fmt.Sprintf("plugins { id(\"dev.isxander.modstitch.base\") version \"0.6.0-unstable\"; java }\n%s\n// OmniBridge generated a common/fabric/neoforge/forge layout. Review omnibridge/contracts before enabling each platform build.\n", base)
	default:
		return "plugins { java }\n" + base
	}
}

func atlasChoiceVersion(values []AtlasToolchainChoice, id, fallback string) string {
	for _, value := range values {
		if value.ID == id && value.Version != "" {
			return value.Version
		}
	}
	return fallback
}

func generatedJavaEntrypoint(target ConversionTargetSpec, packageName, className, modID string, graph UniversalContentGraph) string {
	contractCount := 0
	for _, node := range graph.Nodes {
		if node.Kind == "java-bytecode" || node.Kind == "java-source" || node.Kind == "bedrock-script" || node.Kind == "block" || node.Kind == "item" || node.Kind == "entity" {
			contractCount++
		}
	}
	header := fmt.Sprintf("package %s;\n\n/**\n * Generated by Minecraft Mod Vault OmniBridge.\n * %d semantic contracts remain visible under omnibridge/contracts.\n */\n", packageName, contractCount)
	switch target.Format {
	case "java-fabric":
		return header + fmt.Sprintf("public final class %s implements net.fabricmc.api.ModInitializer {\n    public static final String MOD_ID = %q;\n    @Override public void onInitialize() {\n        System.out.println(\"[OmniBridge] %s initialized; review generated conversion contracts before distribution.\");\n    }\n}\n", className, modID, modID)
	case "java-neoforge":
		return header + fmt.Sprintf("@net.neoforged.fml.common.Mod(%q)\npublic final class %s {\n    public static final String MOD_ID = %q;\n    public %s() {\n        System.out.println(\"[OmniBridge] %s loaded; review generated conversion contracts before distribution.\");\n    }\n}\n", modID, className, modID, className, modID)
	case "java-forge":
		return header + fmt.Sprintf("@net.minecraftforge.fml.common.Mod(%q)\npublic final class %s {\n    public static final String MOD_ID = %q;\n    public %s() {\n        System.out.println(\"[OmniBridge] %s loaded; review generated conversion contracts before distribution.\");\n    }\n}\n", modID, className, modID, className, modID)
	default:
		return header + fmt.Sprintf("public final class %s {\n    public static final String MOD_ID = %q;\n    private %s() {}\n}\n", className, modID, className)
	}
}

func writeJavaLoaderMetadata(resources string, target ConversionTargetSpec, entrypoint, modID string) error {
	switch target.Format {
	case "java-fabric":
		value := map[string]any{"schemaVersion": 1, "id": modID, "version": "${version}", "name": target.Name, "description": firstNonEmpty(target.Description, "Generated by Minecraft Mod Vault OmniBridge"), "environment": "*", "entrypoints": map[string]any{"main": []string{entrypoint}}, "depends": map[string]any{"fabricloader": ">=${loader}", "minecraft": "~${minecraft}", "java": ">=21"}}
		return writeJSONFileAtomic(filepath.Join(resources, "fabric.mod.json"), value)
	case "java-neoforge":
		text := fmt.Sprintf("modLoader=\"javafml\"\nloaderVersion=\"[1,)\"\nlicense=\"All Rights Reserved\"\n[[mods]]\nmodId=\"%s\"\nversion=\"1.0.0\"\ndisplayName=\"%s\"\ndescription='''%s'''\n", modID, target.Name, firstNonEmpty(target.Description, "Generated by OmniBridge"))
		return writeTextFile(filepath.Join(resources, "META-INF", "neoforge.mods.toml"), text)
	case "java-forge":
		text := fmt.Sprintf("modLoader=\"javafml\"\nloaderVersion=\"[47,)\"\nlicense=\"All Rights Reserved\"\n[[mods]]\nmodId=\"%s\"\nversion=\"1.0.0\"\ndisplayName=\"%s\"\ndescription='''%s'''\n", modID, target.Name, firstNonEmpty(target.Description, "Generated by OmniBridge"))
		return writeTextFile(filepath.Join(resources, "META-INF", "mods.toml"), text)
	default:
		return nil
	}
}

func emitNodesToJavaProjectResources(session *ConversionSession, resources string) error {
	for _, node := range session.Graph.Nodes {
		source := conversionNodeSourcePath(session, node)
		if source == "" {
			continue
		}
		if session.Source.Edition == "bedrock" {
			if err := preserveJavaProjectSourceSnapshot(source, resources, session.Plan.Target.Namespace, node); err != nil {
				return err
			}
		}
		switch node.Kind {
		case "texture", "model", "sound", "particle", "font", "shader", "language", "pack-icon":
			if node.Kind == "language" && session.Source.Edition == "bedrock" {
				if err := convertLanguageToJava(source, resources, session.Plan.Target.Namespace); err != nil {
					return err
				}
				continue
			}
			rel := javaResourceDestination(node, session.Plan.Target.Namespace)
			if node.Kind == "pack-icon" {
				rel = "assets/" + session.Plan.Target.Namespace + "/icon" + filepath.Ext(source)
			}
			if err := copyFileReplace(source, filepath.Join(resources, filepath.FromSlash(rel))); err != nil {
				return err
			}
		case "recipe", "loot-table", "function", "tag", "advancement", "predicate", "item-modifier", "worldgen", "structure", "dimension":
			if session.Source.Edition == "java" {
				rel := javaNativeDataRelative(node, session.Plan.Target.GameVersion, session.Plan.Target.Namespace)
				if rel != "" {
					if err := copyFileReplace(source, filepath.Join(resources, filepath.FromSlash(rel))); err != nil {
						return err
					}
				}
			} else if node.Kind == "recipe" {
				if value, ok := translateRecipeToJava(node, source, session.Plan.Target.Namespace); ok {
					rel := filepath.Join("data", session.Plan.Target.Namespace, javaDataFolder(session.Plan.Target.GameVersion, "recipe"), safeNodeFileName(node.Name)+".json")
					if err := writeJSONFileAtomic(filepath.Join(resources, rel), value); err != nil {
						return err
					}
				}
			}
		}
	}
	return writeJavaFeatureMatrix(resources, session.Plan.Target.Namespace, session.Graph, session.Plan.Target)
}

func preserveJavaProjectSourceSnapshot(source, resources, namespace string, node UniversalNode) error {
	rel, err := safeArchiveEntryName(node.SourcePath)
	if err != nil || rel == "" || rel == "." {
		rel = filepath.ToSlash(filepath.Join(node.Kind, safeNodeFileName(node.Name)+filepath.Ext(source)))
	}
	destination := filepath.Join(resources, "assets", sanitizeNamespace(namespace), "omnibridge", "source", filepath.FromSlash(rel))
	return copyFileReplace(source, destination)
}

func emitWorldTarget(session *ConversionSession) ([]string, error) {
	target := session.Plan.Target
	base := cleanConversionName(firstNonEmpty(target.Name, session.Name))
	worldRoot := findConversionWorldRoot(session.Paths.Extracted)
	if worldRoot == "" {
		return emitWorldAdapterWorkspace(session, "No level.dat/world database was found; source content is preserved for adapter review.")
	}
	if session.Source.Edition != target.Edition {
		return emitWorldAdapterWorkspace(session, "Cross-edition world conversion requires Chunker, je2be or Amulet. OmniBridge generated a hash-locked adapter workspace rather than faking converted chunks.")
	}
	stage := filepath.Join(session.Paths.Workspace, target.Format)
	if err := copyDir(worldRoot, stage); err != nil {
		return nil, err
	}
	var output string
	switch target.Format {
	case "bedrock-world":
		output = filepath.Join(session.Paths.Outputs, base+".mcworld")
	case "java-world":
		output = filepath.Join(session.Paths.Outputs, base+"-java-world.zip")
	case "bedrock-template":
		if err := writeBedrockWorldTemplateMetadata(stage, session); err != nil {
			return nil, err
		}
		output = filepath.Join(session.Paths.Outputs, base+".mctemplate")
	}
	if _, _, err := zipDirectoryDeterministic(stage, output, nil); err != nil {
		return nil, err
	}
	return []string{output}, nil
}

func emitWorldAdapterWorkspace(session *ConversionSession, reason string) ([]string, error) {
	root := filepath.Join(session.Paths.Workspace, "world-adapter-workspace")
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}
	if err := copyDir(session.Paths.Extracted, filepath.Join(root, "source")); err != nil {
		return nil, err
	}
	instructions := "# OmniBridge world conversion adapter workspace\n\n" + reason + "\n\nRecommended adapters:\n\n- Chunker / ChunkerCLI: https://oss.chunker.app/\n- je2be-core: https://github.com/kbinani/je2be-core\n- Amulet: https://github.com/Amulet-Team/Amulet-Map-Editor\n\nThe source SHA-256 is `" + session.Source.SHA256 + "`. Do not call this a converted world until the adapter output passes chunk, entity, inventory, dimension and pack validation.\n"
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte(instructions), 0o644); err != nil {
		return nil, err
	}
	if err := writeTargetContracts(session, filepath.Join(root, "omnibridge")); err != nil {
		return nil, err
	}
	output := filepath.Join(session.Paths.Outputs, cleanConversionName(session.Name)+"-world-adapter-workspace.zip")
	if _, _, err := zipDirectoryDeterministic(root, output, nil); err != nil {
		return nil, err
	}
	return []string{output}, nil
}

func writeBedrockWorldTemplateMetadata(root string, session *ConversionSession) error {
	ids := bedrockPackIDs(session.Source.SHA256, "bedrock-template", session.Plan.Target.Namespace)
	manifest := map[string]any{"format_version": 2, "header": map[string]any{"name": "pack.name", "description": "pack.description", "uuid": ids.BPHeader.String(), "version": []int{1, 0, 0}, "base_game_version": bedrockVersionArray(session.Plan.Target.GameVersion)}, "modules": []map[string]any{{"type": "world_template", "uuid": ids.BPData.String(), "version": []int{1, 0, 0}}}}
	if err := writeJSONFileAtomic(filepath.Join(root, "manifest.json"), manifest); err != nil {
		return err
	}
	if err := writeTextFile(filepath.Join(root, "texts", "en_US.lang"), "pack.name="+session.Plan.Target.Name+"\npack.description="+firstNonEmpty(session.Plan.Target.Description, "Converted by Minecraft Mod Vault OmniBridge")+"\n"); err != nil {
		return err
	}
	return writeJSONFileAtomic(filepath.Join(root, "texts", "languages.json"), []string{"en_US"})
}

func emitUniversalBundle(session *ConversionSession) ([]string, error) {
	root := filepath.Join(session.Paths.Workspace, "universal-bundle")
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}
	originalTarget := session.Plan.Target
	for _, format := range []string{"bedrock-project", "java-pack-bundle", "java-multiloader"} {
		target := originalTarget
		target.Format = format
		option, _ := findConversionTarget(format)
		target.Edition = option.Edition
		clone := *session
		plan, err := buildConversionPlan(&clone, target)
		if err != nil {
			return nil, err
		}
		clone.Plan = plan
		clone.Paths.Workspace = filepath.Join(root, "work", format)
		clone.Paths.Outputs = filepath.Join(root, "packages")
		_ = os.MkdirAll(clone.Paths.Workspace, 0o755)
		_ = os.MkdirAll(clone.Paths.Outputs, 0o755)
		switch format {
		case "bedrock-project":
			if _, err := emitBedrockProjectTarget(&clone); err != nil {
				return nil, err
			}
		case "java-pack-bundle":
			if _, err := emitJavaPackBundleTarget(&clone); err != nil {
				return nil, err
			}
		case "java-multiloader":
			if _, err := emitJavaProjectTarget(&clone); err != nil {
				return nil, err
			}
		}
	}
	if err := writeTargetContracts(session, filepath.Join(root, "evidence")); err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# OmniBridge universal bundle\n\nThis bundle preserves the Universal Minecraft Content Graph and emits both edition lanes. Review `evidence/conversion-plan.json` before distributing any generated target.\n"), 0o644); err != nil {
		return nil, err
	}
	output := filepath.Join(session.Paths.Outputs, cleanConversionName(session.Name)+"-universal-bundle.zip")
	if _, _, err := zipDirectoryDeterministic(root, output, nil); err != nil {
		return nil, err
	}
	return []string{output}, nil
}

func writeTargetContracts(session *ConversionSession, root string) error {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	if err := writeJSONFileAtomic(filepath.Join(root, "universal-content-graph.json"), session.Graph); err != nil {
		return err
	}
	if err := writeJSONFileAtomic(filepath.Join(root, "conversion-plan.json"), session.Plan); err != nil {
		return err
	}
	if session.Plan != nil && len(session.Plan.ReviewQueue) > 0 {
		contractsRoot := filepath.Join(root, "contracts")
		if err := os.MkdirAll(contractsRoot, 0o755); err != nil {
			return err
		}
		nodes := map[string]UniversalNode{}
		for _, node := range session.Graph.Nodes {
			nodes[node.ID] = node
		}
		index := []map[string]any{}
		for order, review := range session.Plan.ReviewQueue {
			node := nodes[review.NodeID]
			base := fmt.Sprintf("%03d-%s-%s", order+1, safeNodeFileName(review.Category), safeNodeFileName(review.Title))
			contract := map[string]any{"schemaVersion": conversionSchemaVersion, "review": review, "sourceNode": node, "sourceSha256": session.Source.SHA256, "target": session.Plan.Target, "completionGate": "Implement or run the stated route, build/package the target, exercise the affected behavior in the target runtime, inspect logs, and record proof before marking resolved."}
			jsonPath := filepath.Join(contractsRoot, base+".json")
			if err := writeJSONFileAtomic(jsonPath, contract); err != nil {
				return err
			}
			markdown := "# OmniBridge semantic conversion contract\n\n- Level: **" + string(review.Level) + "**\n- Category: **" + review.Category + "**\n- Feature: **" + review.Title + "**\n- Source node: `" + review.NodeID + "`\n- Source path: `" + node.SourcePath + "`\n- Target: **" + session.Plan.Target.Format + " / " + session.Plan.Target.GameVersion + "**\n\n## Why this is not automatic\n\n" + review.Reason + "\n\n## Required route\n\n" + review.SuggestedRoute + "\n\n## Completion gate\n\nDo not mark this contract complete until the target builds/packages, the affected behavior is exercised in the real target runtime, logs are inspected, and evidence is attached.\n"
			if err := os.WriteFile(filepath.Join(contractsRoot, base+".md"), []byte(markdown), 0o644); err != nil {
				return err
			}
			index = append(index, map[string]any{"id": review.ID, "nodeId": review.NodeID, "level": review.Level, "json": "contracts/" + base + ".json", "markdown": "contracts/" + base + ".md"})
		}
		if err := writeJSONFileAtomic(filepath.Join(root, "contract-index.json"), index); err != nil {
			return err
		}
	}
	return os.WriteFile(filepath.Join(root, "CONVERSION-REPORT.md"), []byte(conversionReportMarkdown(session)), 0o644)
}

func writeConversionProofBundle(session *ConversionSession) error {
	proofRoot := filepath.Join(session.Paths.Workspace, "proof")
	if err := os.MkdirAll(proofRoot, 0o755); err != nil {
		return err
	}
	if err := writeTargetContracts(session, proofRoot); err != nil {
		return err
	}
	if err := writeJSONFileAtomic(filepath.Join(proofRoot, "session.json"), session); err != nil {
		return err
	}
	checksums := []string{}
	for _, output := range session.Outputs {
		checksums = append(checksums, output.SHA256+"  "+output.Name)
	}
	sort.Strings(checksums)
	if err := os.WriteFile(filepath.Join(proofRoot, "SHA256SUMS.txt"), []byte(strings.Join(checksums, "\n")+"\n"), 0o644); err != nil {
		return err
	}
	output := filepath.Join(session.Paths.Outputs, cleanConversionName(session.Name)+"-conversion-proof.zip")
	_, _, err := zipDirectoryDeterministic(proofRoot, output, nil)
	return err
}

func conversionOutputRecord(path string, index int) (ConversionOutput, error) {
	digest, size, err := fileSHA256(path)
	if err != nil {
		return ConversionOutput{}, err
	}
	return ConversionOutput{Name: filepath.Base(path), RelativePath: filepath.ToSlash(filepath.Join("outputs", filepath.Base(path))), Kind: conversionOutputKind(path), Size: size, SHA256: digest, CreatedAt: time.Now().UTC().Format(time.RFC3339), DownloadIndex: index}, nil
}

func conversionOutputKind(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".mcaddon":
		return "bedrock-addon"
	case ".mcpack":
		return "bedrock-pack"
	case ".mcworld":
		return "bedrock-world"
	case ".mctemplate":
		return "bedrock-template"
	default:
		if strings.Contains(strings.ToLower(filepath.Base(path)), "source") {
			return "source-workspace"
		}
		if strings.Contains(strings.ToLower(filepath.Base(path)), "proof") {
			return "proof"
		}
		return "archive"
	}
}

func validateConversionOutput(path, target string) (bool, string) {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return false, "archive open failed: " + err.Error()
	}
	defer reader.Close()
	if len(reader.File) == 0 {
		return false, "archive is empty"
	}
	seen := map[string]bool{}
	manifestCount, packMetaCount := 0, 0
	for _, entry := range reader.File {
		clean, err := safeArchiveEntryName(entry.Name)
		if err != nil {
			return false, err.Error()
		}
		fold := strings.ToLower(clean)
		if seen[fold] {
			return false, "case-colliding duplicate archive path: " + clean
		}
		seen[fold] = true
		if strings.HasSuffix(fold, "manifest.json") {
			data, err := readZipFile(entry, 4<<20)
			if err != nil || !json.Valid(data) {
				return false, "invalid manifest.json"
			}
			manifestCount++
		}
		if strings.HasSuffix(fold, "pack.mcmeta") {
			data, err := readZipFile(entry, 2<<20)
			if err != nil || !json.Valid(data) {
				return false, "invalid pack.mcmeta"
			}
			packMetaCount++
		}
	}
	if strings.HasPrefix(target, "bedrock-") && target != "bedrock-world" && manifestCount == 0 {
		return false, "Bedrock package has no valid manifest.json"
	}
	if strings.HasPrefix(target, "java-") && (target == "java-datapack" || target == "java-resourcepack") && packMetaCount == 0 {
		return false, "Java pack has no valid pack.mcmeta"
	}
	return true, fmt.Sprintf("structurally valid archive: %d entries, %d manifests, %d pack metadata files", len(reader.File), manifestCount, packMetaCount)
}

func writeConversionReport(session *ConversionSession) error {
	if session == nil || session.Paths.ReportMarkdown == "" {
		return nil
	}
	return os.WriteFile(session.Paths.ReportMarkdown, []byte(conversionReportMarkdown(session)), 0o644)
}

func conversionReportMarkdown(session *ConversionSession) string {
	var builder strings.Builder
	builder.WriteString("# OmniBridge Conversion Report\n\n")
	builder.WriteString("- Session: `" + session.ID + "`\n")
	builder.WriteString("- Source: `" + session.Source.Filename + "`\n")
	builder.WriteString("- Source SHA-256: `" + session.Source.SHA256 + "`\n")
	builder.WriteString("- Source type: **" + session.Source.Edition + " / " + session.Source.Kind + " / " + session.Source.Format + "**\n")
	builder.WriteString("- UMCG: `" + session.Graph.GraphVersion + "` with **" + strconv.Itoa(session.Graph.Summary.Total) + "** nodes\n")
	if session.Plan != nil {
		builder.WriteString("- Target: **" + session.Plan.Target.Format + "** for **" + session.Plan.Target.GameVersion + "**\n")
		builder.WriteString(fmt.Sprintf("- Automated coverage: **%.1f%%** (%d exact, %d translated, %d generated)\n", session.Plan.Coverage.AutomatedPercent, session.Plan.Coverage.Exact, session.Plan.Coverage.Translated, session.Plan.Coverage.Generated))
		builder.WriteString(fmt.Sprintf("- Outstanding: **%d tool-assisted, %d review, %d blocked**\n\n", session.Plan.Coverage.ToolAssisted, session.Plan.Coverage.Review, session.Plan.Coverage.Blocked))
		builder.WriteString("## Conversion levels\n\n- **Exact:** bytes or same-schema content preserved.\n- **Translated:** deterministic schema/path translation.\n- **Generated:** target scaffold generated with explicit contracts.\n- **Tool-assisted:** requires a named specialized adapter and verification.\n- **Review:** semantic parity cannot be proven automatically.\n- **Blocked:** no direct target equivalent; reimplementation required.\n\n")
		if len(session.Plan.ReviewQueue) > 0 {
			builder.WriteString("## Review queue\n\n")
			for _, item := range session.Plan.ReviewQueue {
				builder.WriteString("- **" + item.Category + " / " + item.Title + "** — " + item.Reason + " Route: " + item.SuggestedRoute + "\n")
			}
			builder.WriteString("\n")
		}
		if len(session.Plan.ToolAdapters) > 0 {
			builder.WriteString("## Recommended adapters\n\n")
			toolMap := conversionToolMap()
			for _, id := range session.Plan.ToolAdapters {
				tool := toolMap[id]
				builder.WriteString("- **" + firstNonEmpty(tool.Name, id) + "** — " + tool.Role + " " + firstNonEmpty(tool.OfficialURL, tool.RepositoryURL) + "\n")
			}
			builder.WriteString("\n")
		}
	}
	builder.WriteString("## Completion boundary\n\nA generated archive is structurally validated, but gameplay parity is complete only when every review/tool/blocked item is resolved and the target passes real client, dedicated-server, world-load, persistence, networking and gameplay tests appropriate to the content. OmniBridge never hides unsupported semantics.\n")
	return builder.String()
}

func conversionNodeSourcePath(session *ConversionSession, node UniversalNode) string {
	if session == nil || node.SourcePath == "" || strings.Contains(node.SourcePath, ", ") || node.SourcePath == "." {
		return ""
	}
	path := filepath.Join(session.Paths.Extracted, filepath.FromSlash(node.SourcePath))
	if !pathContainedBy(session.Paths.Extracted, path) {
		return ""
	}
	if info, err := os.Stat(path); err == nil && info.Mode().IsRegular() {
		return path
	}
	return ""
}

func copyFileReplace(source, destination string) error {
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(destination, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func copyReviewSource(source, root string, node UniversalNode) error {
	name := safeNodeFileName(node.Name)
	if ext := filepath.Ext(source); ext != "" && !strings.HasSuffix(name, ext) {
		name += ext
	}
	return copyFileReplace(source, filepath.Join(root, "omnibridge_review", node.Kind, name))
}

func writeTextFile(path, text string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(text), 0o644)
}

func bedrockTextureDestination(node UniversalNode, source, rp string) (string, string, string) {
	low := strings.ToLower(node.SourcePath)
	category := "misc"
	atlas := ""
	if strings.Contains(low, "/textures/item") || strings.Contains(low, "/textures/items") {
		category, atlas = "items", "items"
	} else if strings.Contains(low, "/textures/block") || strings.Contains(low, "/textures/blocks") {
		category, atlas = "blocks", "terrain"
	} else if strings.Contains(low, "/textures/entity") {
		category = "entity"
	} else if strings.Contains(low, "/textures/gui") || strings.Contains(low, "/textures/ui") {
		category = "ui"
	}
	name := safeNodeFileName(strings.TrimSuffix(filepath.Base(source), filepath.Ext(source)))
	return filepath.Join(rp, "textures", category, name+filepath.Ext(source)), atlas, sanitizeNamespace(node.Namespace + "_" + name)
}

func bedrockAssetTail(path, marker string) string {
	parts := strings.Split(filepath.ToSlash(path), "/")
	for i, part := range parts {
		if strings.EqualFold(part, marker) && i+1 < len(parts) {
			return strings.Join(parts[i+1:], "/")
		}
	}
	return ""
}

func bedrockNativeRelative(node UniversalNode) string {
	parts := strings.Split(filepath.ToSlash(node.SourcePath), "/")
	folders := map[string]bool{"blocks": true, "items": true, "entities": true, "recipes": true, "loot_tables": true, "functions": true, "scripts": true, "animations": true, "animation_controllers": true, "render_controllers": true, "particles": true, "textures": true, "models": true, "sounds": true, "texts": true, "ui": true, "structures": true, "biomes": true, "spawn_rules": true, "features": true, "feature_rules": true, "attachables": true, "fog": true, "materials": true}
	for index, part := range parts {
		if folders[strings.ToLower(part)] {
			return strings.Join(parts[index:], "/")
		}
	}
	return ""
}

func javaResourceDestination(node UniversalNode, fallbackNamespace string) string {
	namespace := firstNonEmpty(node.Namespace, fallbackNamespace)
	path := filepath.ToSlash(node.SourcePath)
	if index := strings.Index(strings.ToLower(path), "assets/"); index >= 0 {
		return path[index:]
	}
	category := map[string]string{"texture": "textures", "model": "models", "sound": "sounds", "particle": "particles", "font": "font", "shader": "shaders", "language": "lang"}[node.Kind]
	if category == "" {
		category = node.Kind
	}
	return filepath.ToSlash(filepath.Join("assets", namespace, category, safeNodeFileName(node.Name)+filepath.Ext(path)))
}

func javaNativeDataRelative(node UniversalNode, version, fallbackNamespace string) string {
	path := filepath.ToSlash(node.SourcePath)
	if index := strings.Index(strings.ToLower(path), "data/"); index >= 0 {
		return path[index:]
	}
	folder := javaDataFolder(version, node.Kind)
	if folder == "" {
		return ""
	}
	return filepath.ToSlash(filepath.Join("data", firstNonEmpty(node.Namespace, fallbackNamespace), folder, safeNodeFileName(node.Name)+filepath.Ext(path)))
}

func javaDataFolder(version, kind string) string {
	modern := compareGameVersion(version, "1.21") >= 0
	mapping := map[string][2]string{
		"recipe": {"recipes", "recipe"}, "loot-table": {"loot_tables", "loot_table"}, "function": {"functions", "function"},
		"tag": {"tags", "tags"}, "advancement": {"advancements", "advancement"}, "predicate": {"predicates", "predicate"},
		"item-modifier": {"item_modifiers", "item_modifier"}, "worldgen": {"worldgen", "worldgen"}, "structure": {"structures", "structure"}, "dimension": {"dimension", "dimension"},
	}
	value, ok := mapping[kind]
	if !ok {
		return ""
	}
	if modern {
		return value[1]
	}
	return value[0]
}

func convertLanguageToBedrock(source, rp string) error {
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	locale := strings.TrimSuffix(filepath.Base(source), filepath.Ext(source))
	locale = javaLocaleToBedrock(locale)
	lines := []string{}
	if strings.EqualFold(filepath.Ext(source), ".json") {
		values := map[string]string{}
		if err := json.Unmarshal(data, &values); err != nil {
			return err
		}
		keys := make([]string, 0, len(values))
		for key := range values {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			lines = append(lines, key+"="+strings.ReplaceAll(values[key], "\n", "\\n"))
		}
	} else {
		lines = strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	}
	if err := writeTextFile(filepath.Join(rp, "texts", locale+".lang"), strings.Join(lines, "\n")); err != nil {
		return err
	}
	languagesPath := filepath.Join(rp, "texts", "languages.json")
	languages := []string{}
	if existing, err := os.ReadFile(languagesPath); err == nil {
		_ = json.Unmarshal(existing, &languages)
	}
	languages = uniqueStringsPreserve(append(languages, locale))
	sort.Strings(languages)
	return writeJSONFileAtomic(languagesPath, languages)
}

func convertLanguageToJava(source, root, namespace string) error {
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	locale := strings.ToLower(strings.TrimSuffix(filepath.Base(source), filepath.Ext(source)))
	locale = strings.ReplaceAll(locale, "-", "_")
	values := map[string]string{}
	if strings.EqualFold(filepath.Ext(source), ".json") {
		if err := json.Unmarshal(data, &values); err != nil {
			return err
		}
	} else {
		for _, line := range strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if index := strings.Index(line, "="); index >= 0 {
				values[strings.TrimSpace(line[:index])] = strings.ReplaceAll(strings.TrimSpace(line[index+1:]), "\\n", "\n")
			}
		}
	}
	return writeJSONFileAtomic(filepath.Join(root, "assets", namespace, "lang", locale+".json"), values)
}

func translateRecipeToBedrock(node UniversalNode, source, namespace string) (map[string]any, bool) {
	data, err := os.ReadFile(source)
	if err != nil {
		return nil, false
	}
	var value map[string]any
	if json.Unmarshal(data, &value) != nil {
		return nil, false
	}
	for key := range value {
		if strings.HasPrefix(key, "minecraft:recipe_") {
			return value, true
		}
	}
	typ := strings.TrimPrefix(strings.ToLower(flexibleString(value["type"])), "minecraft:")
	name := namespace + ":" + safeNodeFileName(node.Name)
	result := javaRecipeResultToBedrock(value["result"])
	switch typ {
	case "crafting_shaped":
		return map[string]any{"format_version": "1.20.10", "minecraft:recipe_shaped": map[string]any{"description": map[string]any{"identifier": name}, "tags": []string{"crafting_table"}, "pattern": value["pattern"], "key": javaRecipeKeyToBedrock(value["key"]), "result": result}}, true
	case "crafting_shapeless":
		return map[string]any{"format_version": "1.20.10", "minecraft:recipe_shapeless": map[string]any{"description": map[string]any{"identifier": name}, "tags": []string{"crafting_table"}, "ingredients": javaIngredientsToBedrock(value["ingredients"]), "result": result}}, true
	case "smelting", "blasting", "smoking", "campfire_cooking":
		tags := map[string][]string{"smelting": {"furnace"}, "blasting": {"blast_furnace"}, "smoking": {"smoker"}, "campfire_cooking": {"campfire"}}[typ]
		return map[string]any{"format_version": "1.20.10", "minecraft:recipe_furnace": map[string]any{"description": map[string]any{"identifier": name}, "tags": tags, "input": javaIngredientToBedrock(value["ingredient"]), "output": flexibleString(result["item"])}}, true
	default:
		return nil, false
	}
}

func translateRecipeToJava(node UniversalNode, source, namespace string) (map[string]any, bool) {
	data, err := os.ReadFile(source)
	if err != nil {
		return nil, false
	}
	var value map[string]any
	if json.Unmarshal(data, &value) != nil {
		return nil, false
	}
	if typ := flexibleString(value["type"]); typ != "" {
		return value, true
	}
	for key, raw := range value {
		body, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		switch key {
		case "minecraft:recipe_shaped":
			return map[string]any{"type": "minecraft:crafting_shaped", "pattern": body["pattern"], "key": bedrockRecipeKeyToJava(body["key"]), "result": bedrockRecipeResultToJava(body["result"])}, true
		case "minecraft:recipe_shapeless":
			return map[string]any{"type": "minecraft:crafting_shapeless", "ingredients": bedrockIngredientsToJava(body["ingredients"]), "result": bedrockRecipeResultToJava(body["result"])}, true
		case "minecraft:recipe_furnace":
			return map[string]any{"type": "minecraft:smelting", "ingredient": bedrockIngredientToJava(body["input"]), "result": flexibleString(body["output"]), "experience": 0, "cookingtime": 200}, true
		}
	}
	_ = node
	_ = namespace
	return nil, false
}

func javaRecipeResultToBedrock(raw any) map[string]any {
	switch value := raw.(type) {
	case string:
		return map[string]any{"item": value, "count": 1}
	case map[string]any:
		item := firstNonEmpty(flexibleString(value["id"]), flexibleString(value["item"]))
		count := intFlexible(value["count"])
		if count <= 0 {
			count = 1
		}
		return map[string]any{"item": item, "count": count}
	default:
		return map[string]any{"item": "minecraft:air", "count": 1}
	}
}

func javaIngredientToBedrock(raw any) any {
	switch value := raw.(type) {
	case string:
		return map[string]any{"item": value}
	case map[string]any:
		if item := flexibleString(value["item"]); item != "" {
			return map[string]any{"item": item}
		}
		if tag := flexibleString(value["tag"]); tag != "" {
			return map[string]any{"tag": tag}
		}
	case []any:
		if len(value) > 0 {
			return javaIngredientToBedrock(value[0])
		}
	}
	return map[string]any{"item": "minecraft:air"}
}

func javaIngredientsToBedrock(raw any) []any {
	out := []any{}
	if values, ok := raw.([]any); ok {
		for _, value := range values {
			out = append(out, javaIngredientToBedrock(value))
		}
	}
	return out
}

func javaRecipeKeyToBedrock(raw any) map[string]any {
	out := map[string]any{}
	if values, ok := raw.(map[string]any); ok {
		for key, value := range values {
			out[key] = javaIngredientToBedrock(value)
		}
	}
	return out
}

func bedrockRecipeResultToJava(raw any) map[string]any {
	if value, ok := raw.(map[string]any); ok {
		item := firstNonEmpty(flexibleString(value["item"]), flexibleString(value["id"]))
		count := intFlexible(value["count"])
		if count <= 0 {
			count = 1
		}
		return map[string]any{"id": item, "count": count}
	}
	return map[string]any{"id": flexibleString(raw), "count": 1}
}

func bedrockIngredientToJava(raw any) any {
	if value, ok := raw.(map[string]any); ok {
		if item := flexibleString(value["item"]); item != "" {
			return item
		}
		if tag := flexibleString(value["tag"]); tag != "" {
			return map[string]any{"tag": tag}
		}
	}
	return raw
}

func bedrockIngredientsToJava(raw any) []any {
	out := []any{}
	if values, ok := raw.([]any); ok {
		for _, value := range values {
			out = append(out, bedrockIngredientToJava(value))
		}
	}
	return out
}

func bedrockRecipeKeyToJava(raw any) map[string]any {
	out := map[string]any{}
	if values, ok := raw.(map[string]any); ok {
		for key, value := range values {
			out[key] = bedrockIngredientToJava(value)
		}
	}
	return out
}

func bedrockVersionArray(version string) []int {
	parts := parseNumericVersion(version)
	for len(parts) < 3 {
		parts = append(parts, 0)
	}
	if len(parts) > 3 {
		parts = parts[:3]
	}
	if len(parts) == 0 {
		return []int{1, 26, 30}
	}
	return parts
}

func javaPackFormat(version, kind string) float64 {
	if atlas, err := loadRepairVersionAtlas(); err == nil {
		if row, ok := atlas.MCMeta[strings.TrimSpace(version)]; ok {
			if kind == "resource" && row.ResourcePackVersion > 0 {
				if row.ResourcePackVersionMinor > 0 {
					return float64(row.ResourcePackVersion) + float64(row.ResourcePackVersionMinor)/100
				}
				return float64(row.ResourcePackVersion)
			}
			if kind != "resource" && row.DataPackVersion > 0 {
				if row.DataPackVersionMinor > 0 {
					return float64(row.DataPackVersion) + float64(row.DataPackVersionMinor)/100
				}
				return float64(row.DataPackVersion)
			}
		}
	}
	data := map[string]float64{"1.21": 48, "1.21.1": 48, "1.21.2": 57, "1.21.3": 57, "1.21.4": 61, "1.21.5": 71, "1.21.6": 80, "1.21.7": 81, "1.21.8": 81, "1.21.9": 88, "1.21.10": 88, "1.21.11": 94, "26.1": 101, "26.1.1": 101, "26.1.2": 101, "26.2": 107}
	resource := map[string]float64{"1.21": 34, "1.21.1": 34, "1.21.2": 42, "1.21.3": 42, "1.21.4": 46, "1.21.5": 55, "1.21.6": 63, "1.21.7": 64, "1.21.8": 64, "1.21.9": 69, "1.21.10": 69, "1.21.11": 75, "26.1": 84, "26.1.1": 84, "26.1.2": 84, "26.2": 87}
	if kind == "resource" {
		if value, ok := resource[version]; ok {
			return value
		}
	} else if value, ok := data[version]; ok {
		return value
	}
	legacy := []struct {
		version        string
		data, resource float64
	}{{"1.20.5", 41, 32}, {"1.20.3", 26, 22}, {"1.20.2", 18, 18}, {"1.20", 15, 15}, {"1.19.4", 12, 13}, {"1.19.3", 10, 12}, {"1.19", 10, 9}, {"1.18.2", 9, 8}, {"1.18", 8, 8}, {"1.17", 7, 7}, {"1.16.2", 6, 6}, {"1.15", 5, 5}, {"1.13", 4, 4}, {"1.11", 3, 3}, {"1.9", 2, 2}, {"1.6", 1, 1}}
	for _, item := range legacy {
		if compareGameVersion(version, item.version) >= 0 {
			if kind == "resource" {
				return item.resource
			}
			return item.data
		}
	}
	return 1
}

func isModernDecimalPackFormat(version string) bool {
	return compareGameVersion(version, "1.21.9") >= 0
}

func compareGameVersion(a, b string) int {
	ap, bp := parseNumericVersion(a), parseNumericVersion(b)
	max := len(ap)
	if len(bp) > max {
		max = len(bp)
	}
	for len(ap) < max {
		ap = append(ap, 0)
	}
	for len(bp) < max {
		bp = append(bp, 0)
	}
	for i := 0; i < max; i++ {
		if ap[i] < bp[i] {
			return -1
		}
		if ap[i] > bp[i] {
			return 1
		}
	}
	return 0
}

func parseNumericVersion(value string) []int {
	value = strings.TrimSpace(strings.TrimPrefix(value, "v"))
	parts := []int{}
	for _, part := range strings.Split(value, ".") {
		digits := strings.Builder{}
		for _, r := range part {
			if r >= '0' && r <= '9' {
				digits.WriteRune(r)
			} else {
				break
			}
		}
		if digits.Len() == 0 {
			break
		}
		n, _ := strconv.Atoi(digits.String())
		parts = append(parts, n)
	}
	return parts
}

func javaLocaleToBedrock(locale string) string {
	parts := strings.Split(strings.ReplaceAll(locale, "-", "_"), "_")
	if len(parts) == 2 {
		return strings.ToLower(parts[0]) + "_" + strings.ToUpper(parts[1])
	}
	if strings.EqualFold(locale, "en_us") {
		return "en_US"
	}
	return locale
}

func graphHasKinds(graph UniversalContentGraph, kinds ...string) bool {
	wanted := map[string]bool{}
	for _, kind := range kinds {
		wanted[kind] = true
	}
	for _, node := range graph.Nodes {
		if wanted[node.Kind] {
			return true
		}
	}
	return false
}

func safeNodeFileName(value string) string {
	value = filepath.Base(filepath.ToSlash(strings.TrimSpace(value)))
	value = strings.TrimSuffix(value, filepath.Ext(value))
	return sanitizeNamespace(value)
}

func shortPackFolder(namespace, suffix string) string {
	base := strings.ReplaceAll(sanitizeNamespace(namespace), "_", "")
	if len(base) > 7 {
		base = base[:7]
	}
	return strings.ToUpper(base + suffix)
}

func javaClassName(value string) string {
	parts := strings.FieldsFunc(value, func(r rune) bool { return !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9') })
	var builder strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		builder.WriteString(strings.ToUpper(part[:1]))
		if len(part) > 1 {
			builder.WriteString(part[1:])
		}
	}
	if builder.Len() == 0 {
		return "ConvertedMod"
	}
	if first := builder.String()[0]; first >= '0' && first <= '9' {
		return "Mod" + builder.String()
	}
	return builder.String()
}

func generatedJavaProjectReadme(session *ConversionSession, resolution AtlasResolution) string {
	return fmt.Sprintf("# %s — OmniBridge generated %s workspace\n\nSource SHA-256: `%s`\n\nTarget: Minecraft `%s`, loader `%s`, Java `%d`.\n\nThis workspace contains translated assets/data and executable loader scaffolding. It does **not** claim that edition-specific logic, Mixins, networking, rendering or world generation are complete. Resolve every item in `omnibridge/conversion-plan.json`, then run loader build, client, dedicated-server, persistence and gameplay tests.\n\n## Toolchain evidence\n\nLoader supported in embedded atlas: `%t`\nLoader version: `%s`\n\nGenerated by Minecraft Mod Vault %s.\n", session.Plan.Target.Name, session.Plan.Target.Format, session.Source.SHA256, session.Plan.Target.GameVersion, session.Plan.Target.Loader, resolution.JavaMajor, resolution.LoaderSupported, resolution.LoaderVersion, appVersion)
}

func findConversionWorldRoot(root string) string {
	best := ""
	bestDepth := 1 << 30
	_ = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.EqualFold(entry.Name(), "level.dat") {
			return err
		}
		dir := filepath.Dir(path)
		rel, _ := filepath.Rel(root, dir)
		depth := len(strings.Split(filepath.ToSlash(rel), "/"))
		if depth < bestDepth {
			best, bestDepth = dir, depth
		}
		return nil
	})
	return best
}
