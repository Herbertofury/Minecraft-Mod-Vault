package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

func buildConversionPlan(session *ConversionSession, target ConversionTargetSpec) (*ConversionPlan, error) {
	if session == nil || session.Source.SHA256 == "" || session.Graph.GraphVersion == "" {
		return nil, errors.New("conversion source has not been profiled")
	}
	target.Format = strings.ToLower(strings.TrimSpace(target.Format))
	option, ok := findConversionTarget(target.Format)
	if !ok {
		return nil, fmt.Errorf("unsupported conversion target %q", target.Format)
	}
	target.Edition = option.Edition
	target.Name = firstNonEmpty(strings.TrimSpace(target.Name), session.Source.Name)
	target.Namespace = sanitizeNamespace(firstNonEmpty(target.Namespace, session.Source.Namespace, target.Name))
	if target.GameVersion == "" {
		if target.Edition == "bedrock" {
			target.GameVersion = "1.26.30"
		} else {
			target.GameVersion = firstNonEmpty(session.Source.GameVersion, "1.21.1")
		}
	}
	if strings.HasPrefix(target.Format, "java-") && target.Loader == "" {
		switch target.Format {
		case "java-fabric":
			target.Loader = "fabric"
		case "java-neoforge":
			target.Loader = "neoforge"
		case "java-forge":
			target.Loader = "forge"
		case "java-multiloader":
			target.Loader = "multiloader"
		case "java-world-mod":
			target.Loader = "fabric"
		default:
			target.Loader = "vanilla"
		}
	}
	selected := map[string]bool{}
	for _, id := range target.SelectedNodes {
		selected[id] = true
	}
	plan := &ConversionPlan{SchemaVersion: conversionSchemaVersion, CreatedAt: time.Now().UTC().Format(time.RFC3339), SourceSHA256: session.Source.SHA256, Target: target}
	levels := map[ConversionLevel][]string{}
	toolSet := map[string]bool{}
	for _, node := range session.Graph.Nodes {
		if len(selected) > 0 && !selected[node.ID] {
			continue
		}
		level, reason, suggested, toolID := conversionLevelForTarget(session.Source, node, target)
		levels[level] = append(levels[level], node.ID)
		if toolID != "" {
			toolSet[toolID] = true
		}
		needsSemanticReview := level == ConversionReview || level == ConversionBlocked || level == ConversionToolAssisted || level == ConversionGenerated && generatedNodeRequiresSemanticReview(node.Kind)
		if needsSemanticReview {
			severity := "review"
			if level == ConversionBlocked {
				severity = "blocked"
			} else if level == ConversionToolAssisted {
				severity = "adapter"
			} else if level == ConversionGenerated {
				severity = "implementation"
			}
			plan.ReviewQueue = append(plan.ReviewQueue, ConversionReviewItem{ID: universalNodeID("review", target.Format+node.ID), NodeID: node.ID, Severity: severity, Category: node.Kind, Title: node.Name, Reason: reason, SuggestedRoute: suggested, Level: level})
		}
	}
	plan.Coverage.Total = sumLevelCounts(levels)
	plan.Coverage.Exact = len(levels[ConversionExact])
	plan.Coverage.Translated = len(levels[ConversionTranslated])
	plan.Coverage.Generated = len(levels[ConversionGenerated])
	plan.Coverage.ToolAssisted = len(levels[ConversionToolAssisted])
	plan.Coverage.Review = len(levels[ConversionReview])
	plan.Coverage.Blocked = len(levels[ConversionBlocked])
	if plan.Coverage.Total > 0 {
		automated := float64(plan.Coverage.Exact+plan.Coverage.Translated+plan.Coverage.Generated) / float64(plan.Coverage.Total) * 100
		plan.Coverage.AutomatedPercent = float64(int(automated*10+0.5)) / 10
	}
	switch {
	case plan.Coverage.Blocked > 0:
		plan.Coverage.CompletenessState = "blocked-items-require-reimplementation"
	case len(plan.ReviewQueue) > 0:
		plan.Coverage.CompletenessState = "review-or-adapter-required"
	default:
		plan.Coverage.CompletenessState = "fully-automated-plan"
	}
	plan.Steps = conversionPlanSteps(levels, target)
	for tool := range toolSet {
		plan.ToolAdapters = append(plan.ToolAdapters, tool)
	}
	sort.Strings(plan.ToolAdapters)
	plan.Warnings = uniqueStringsPreserve(append(plan.Warnings, session.Source.Warnings...))
	if target.Edition == "java" {
		if atlas, err := loadRepairVersionAtlas(); err == nil {
			resolution := atlas.Resolve(target.GameVersion, target.Loader)
			if !resolution.Exists {
				plan.Warnings = append(plan.Warnings, "Target Java version is not present in the embedded Mojang/mcmeta Version Atlas; manifests and pack formats require manual verification.")
			}
			plan.Warnings = append(plan.Warnings, resolution.Warnings...)
		}
	}
	if session.Source.Edition != target.Edition && target.Edition != "universal" {
		plan.Warnings = append(plan.Warnings, "Cross-edition conversion translates content contracts, not virtual-machine semantics. Runtime behavior must pass target-edition tests before the result is called complete.")
	}
	if plan.Coverage.Blocked > 0 {
		plan.Losses = append(plan.Losses, "Blocked nodes are preserved in the proof bundle and generated as reimplementation contracts; they are never silently omitted.")
	}
	if plan.Coverage.ToolAssisted > 0 {
		plan.Warnings = append(plan.Warnings, "Tool-assisted nodes remain incomplete until the named adapter runs successfully and its output passes validation.")
	}
	digestValue := struct {
		Source string                       `json:"source"`
		Target ConversionTargetSpec         `json:"target"`
		Levels map[ConversionLevel][]string `json:"levels"`
	}{session.Source.SHA256, target, levels}
	encoded, _ := json.Marshal(digestValue)
	digest := sha256.Sum256(encoded)
	plan.PlanSHA256 = hex.EncodeToString(digest[:])
	return plan, nil
}

func generatedNodeRequiresSemanticReview(kind string) bool {
	switch kind {
	case "bedrock-script", "java-source", "block", "item", "entity", "animation", "animation-controller", "particle", "attachable", "spawn-rule", "feature", "feature-rule", "biome", "fog", "camera-preset", "dialogue", "trade-table", "volume", "dimension", "worldgen", "ui", "render-controller", "material":
		return true
	default:
		return false
	}
}

func findConversionTarget(id string) (ConversionTargetOption, bool) {
	for _, option := range conversionTargetOptions() {
		if option.ID == id {
			return option, true
		}
	}
	return ConversionTargetOption{}, false
}

func conversionLevelForTarget(source ConversionSourceProfile, node UniversalNode, target ConversionTargetSpec) (ConversionLevel, string, string, string) {
	sameEdition := source.Edition == target.Edition
	if target.Format == "universal-bundle" {
		level := node.Level
		if level == ConversionBlocked {
			return ConversionReview, "Universal bundle preserves this feature as a target-specific reimplementation contract.", "Review the generated contract and choose a Java or Bedrock implementation lane.", ""
		}
		return level, "Universal bundle retains this node and all target evidence.", "Inspect the per-target lane.", adapterForNode(node)
	}
	if !targetSupportsNode(node, target.Format) {
		switch node.Kind {
		case "mixin", "shader", "render-controller", "ui", "material":
			return ConversionBlocked, "This feature is tied to renderer, bytecode injection or edition-specific UI/runtime internals and has no direct representation in the selected target.", "Reimplement against target-supported rendering, UI or runtime APIs; OmniBridge preserves the source and an explicit contract.", ""
		case "java-bytecode":
			if target.Edition == "bedrock" {
				return ConversionReview, "JVM bytecode and loader callbacks have no direct Bedrock runtime; decompilation and semantic reconstruction are required.", "Use Repair Lab/Porting Lab plus generated Script API contracts; PortKit/ModMorpher may assist but remain experimental.", "portkit"
			}
		case "world-data":
			tool := "chunker"
			if target.Edition == "java" {
				tool = "je2be"
			}
			return ConversionToolAssisted, "World chunks, entities, block states and metadata require a dedicated cross-edition data converter.", "Run a verified Chunker, je2be or Amulet adapter and inspect every warning before promotion.", tool
		case "structure":
			return ConversionToolAssisted, "Structure NBT/schematic formats need block-state and palette conversion.", "Use SchemConvert or a world adapter, then validate placement in the target edition.", "schemconvert"
		}
		return ConversionReview, "The selected package type cannot represent this content category directly.", "Choose a richer target such as bedrock-addon, java-multiloader or universal-bundle, or implement the generated target contract.", ""
	}
	if sameEdition {
		switch target.Format {
		case "bedrock-addon", "bedrock-behavior", "bedrock-resource", "bedrock-template", "bedrock-world", "bedrock-project", "bedrock-world-product":
			if source.Edition == "bedrock" {
				if node.Kind == "manifest" {
					return ConversionGenerated, "Manifest identity and dependency links must be regenerated for a collision-free output.", "Validate in the target Bedrock version.", ""
				}
				return ConversionExact, "Bedrock-native source is preserved in the matching pack lane.", "Validate references and minimum engine version.", ""
			}
		case "java-datapack", "java-resourcepack", "java-pack-bundle":
			if source.Edition == "java" && node.Kind != "java-bytecode" && node.Kind != "java-source" && node.Kind != "mixin" {
				return ConversionExact, "Java-native pack content is copied without schema translation.", "Run target-version pack validation.", ""
			}
		case "java-fabric", "java-neoforge", "java-forge", "java-multiloader", "java-world-mod":
			if source.Edition == "java" {
				if target.Format == "java-world-mod" && node.Kind == "world-data" {
					return ConversionExact, "Java world data is embedded as an immutable world-template archive with a safe exporter utility.", "Add the desired loader menu/command integration and test world creation, restart and upgrade behavior.", ""
				}
				switch node.Kind {
				case "texture", "sound", "language", "pack-icon", "recipe", "loot-table", "function", "tag", "advancement", "predicate", "item-modifier":
					return ConversionExact, "Java content is preserved in the generated source project.", "Build and run client/server tests.", ""
				case "java-source":
					return ConversionGenerated, "Source is staged into a target-loader workspace with migration contracts.", "Use Repair Lab and loader tooling to resolve APIs, then compile.", toolForJavaTarget(target)
				case "java-bytecode", "mixin":
					return ConversionToolAssisted, "Compiled logic requires decompilation/remapping and API migration; byte copying is not a port.", "Open the generated workspace in Porting Lab and Repair Lab.", "tiny-remapper"
				}
			}
		}
	}

	switch node.Kind {
	case "texture", "pack-icon", "sound":
		return ConversionExact, "Binary media can be preserved while target indexes and paths are regenerated.", "Inspect visual/audio parity in game.", ""
	case "language":
		return ConversionTranslated, "Localization key/value data has a deterministic target representation.", "Review formatting codes and fallback locale names.", ""
	case "recipe":
		if recipeIsCommon(node.Data) {
			return ConversionTranslated, "Shaped, shapeless and furnace-family recipes have deterministic target emitters.", "Validate item identifiers against the target registry.", ""
		}
		return ConversionReview, "Recipe type is not one of the safely translated common schemas.", "Implement or select a recipe adapter after reviewing the source JSON.", ""
	case "function":
		return ConversionReview, "Function files share a text format but Java and Bedrock commands, selectors and execution semantics differ.", "Use the command migration report and test every command in the target edition.", ""
	case "loot-table", "tag", "advancement", "predicate", "item-modifier":
		return ConversionReview, "The target has related data-driven concepts, but schemas and runtime semantics are not identical.", "Use the emitted source contract to complete target-schema translation.", ""
	case "model":
		if simpleJavaCubeModel(node.Data) && target.Edition == "bedrock" {
			return ConversionTranslated, "Simple cube-all Java models can be expressed as Bedrock block texture/material definitions.", "Inspect geometry and display transforms.", "packconverter"
		}
		return ConversionReview, "Model geometry, predicates and display transforms require target-specific interpretation.", "Use PackConverter/Rainbow for supported Java assets or edit the generated geometry contract.", "packconverter"
	case "block", "item", "entity", "animation", "animation-controller", "particle", "attachable", "spawn-rule", "feature", "feature-rule", "biome", "fog":
		return ConversionGenerated, "OmniBridge can generate a target definition/scaffold, but gameplay semantics require validation.", "Review generated components and run target-edition gameplay tests.", ""
	case "bedrock-script":
		if strings.HasPrefix(target.Format, "java-") {
			return ConversionGenerated, "Bedrock Script API events are emitted as Java service/event contracts rather than text-translated JavaScript.", "Implement each contract with the selected loader API and test both logical sides.", toolForJavaTarget(target)
		}
		return ConversionExact, "Bedrock script is preserved for a Bedrock target.", "Type-check against the target Script API module versions.", ""
	case "java-source":
		if target.Edition == "bedrock" {
			return ConversionGenerated, "Java source is summarized into Script API and data-component contracts; direct language translation would be unsafe.", "Implement generated behavior contracts and validate with GameTest.", ""
		}
		return ConversionGenerated, "Java source is staged into a target-loader project.", "Compile and run target-version tests.", toolForJavaTarget(target)
	case "java-bytecode":
		if target.Edition == "bedrock" {
			return ConversionReview, "JVM bytecode and loader callbacks have no direct Bedrock runtime; decompilation and semantic reconstruction are required.", "Use Repair Lab/Porting Lab plus generated Script API contracts; PortKit/ModMorpher may assist but are experimental.", "portkit"
		}
		return ConversionToolAssisted, "Compiled classes require mappings, remapping, decompilation and loader/API migration.", "Use Tiny Remapper/Loom and the generated migration workspace.", "tiny-remapper"
	case "mixin", "shader", "render-controller", "ui", "material":
		return ConversionBlocked, "This feature is tied to renderer, bytecode injection or edition-specific UI/runtime internals.", "Reimplement against target-supported rendering/UI APIs; the source and contract are preserved.", ""
	case "world-data":
		tool := "chunker"
		if target.Edition == "java" {
			tool = "je2be"
		}
		return ConversionToolAssisted, "World chunks, entities, block states and metadata require a dedicated cross-edition data converter.", "Install/configure Chunker, je2be or Amulet and run the adapter, then inspect conversion warnings.", tool
	case "structure":
		return ConversionToolAssisted, "Structure NBT/schematic formats need block-state and palette conversion.", "Use SchemConvert or a world adapter, then validate placement in the target edition.", "schemconvert"
	case "worldgen", "dimension":
		return ConversionReview, "World-generation and dimension systems differ substantially between editions and versions.", "Generate target worldgen contracts and validate seeds, placement and biome behavior.", ""
	case "manifest":
		return ConversionGenerated, "Target manifests and pack metadata must receive new identities and version requirements.", "Review generated UUIDs, dependencies and minimum target version.", ""
	default:
		return node.Level, "Content is retained according to its Universal Content Graph classification.", "Review target validation results.", adapterForNode(node)
	}
}

func targetSupportsNode(node UniversalNode, target string) bool {
	if len(node.TargetSupport) == 0 {
		return true
	}
	aliases := map[string][]string{
		"bedrock-project":       {"bedrock-addon", "bedrock-behavior", "bedrock-resource"},
		"bedrock-world-product": {"bedrock-world", "bedrock-template", "bedrock-addon", "bedrock-behavior", "bedrock-resource"},
		"java-world-mod":        {"java-world", "java-fabric", "java-neoforge", "java-forge", "java-multiloader", "java-datapack", "java-resourcepack"},
		"java-pack-bundle":      {"java-datapack", "java-resourcepack"},
	}
	for _, value := range node.TargetSupport {
		if value == target || value == "universal-bundle" && target == "universal-bundle" {
			return true
		}
		for _, alias := range aliases[target] {
			if value == alias {
				return true
			}
		}
	}
	return false
}

func toolForJavaTarget(target ConversionTargetSpec) string {
	if target.Format == "java-multiloader" {
		return "modstitch"
	}
	if target.Format == "java-world-mod" && target.Loader == "multiloader" {
		return "modstitch"
	}
	return "fabric-loom"
}

func adapterForNode(node UniversalNode) string {
	switch node.Kind {
	case "world-data":
		return "chunker"
	case "structure":
		return "schemconvert"
	case "model":
		return "packconverter"
	case "java-bytecode":
		return "tiny-remapper"
	}
	return ""
}

func conversionPlanSteps(levels map[ConversionLevel][]string, target ConversionTargetSpec) []ConversionPlanStep {
	steps := []ConversionPlanStep{
		{ID: "verify-source", Order: 1, Title: "Verify immutable source", Description: "Re-hash the original archive and extracted tree before mutation.", State: "ready", Level: ConversionExact},
		{ID: "materialize-graph", Order: 2, Title: "Materialize Universal Minecraft Content Graph", Description: "Normalize assets, data, logic, worlds and relationships into a versioned intermediate representation.", State: "ready", Level: ConversionExact},
	}
	order := 3
	for _, item := range []struct {
		level ConversionLevel
		title string
		desc  string
	}{
		{ConversionExact, "Preserve lossless content", "Copy binary media and same-schema content with deterministic path normalization."},
		{ConversionTranslated, "Translate data schemas", "Run built-in localization, recipe, texture-index and pack-schema emitters."},
		{ConversionGenerated, "Generate target scaffolds", "Generate manifests, loaders, Script API/Java contracts and target project structure."},
		{ConversionToolAssisted, "Run specialized adapters", "Hand off worlds, structures, remapping and advanced resources to verified tools when installed."},
		{ConversionReview, "Resolve semantic review queue", "Review commands, logic, worldgen, advanced models and identifiers that cannot be proven automatically."},
		{ConversionBlocked, "Reimplement target-specific internals", "Preserve blocked features as explicit contracts instead of silently dropping them."},
	} {
		if len(levels[item.level]) == 0 {
			continue
		}
		state := "ready"
		if item.level == ConversionToolAssisted {
			state = "adapter-required"
		}
		if item.level == ConversionReview || item.level == ConversionBlocked {
			state = "review-required"
		}
		steps = append(steps, ConversionPlanStep{ID: string(item.level), Order: order, Title: item.title, Description: item.desc, State: state, Level: item.level, NodeIDs: levels[item.level]})
		order++
	}
	steps = append(steps,
		ConversionPlanStep{ID: "package", Order: order, Title: "Package target", Description: "Create deterministic " + target.Format + " output and proof bundle.", State: "ready", Level: ConversionGenerated},
		ConversionPlanStep{ID: "validate", Order: order + 1, Title: "Validate target", Description: "Check archive safety, manifests, references, hashes, review queue and target-specific structural rules.", State: "ready", Level: ConversionExact},
	)
	return steps
}

func sumLevelCounts(levels map[ConversionLevel][]string) int {
	total := 0
	for _, ids := range levels {
		total += len(ids)
	}
	return total
}

func recipeIsCommon(data map[string]any) bool {
	if len(data) == 0 {
		return false
	}
	typ := strings.ToLower(flexibleString(data["type"]))
	if strings.HasPrefix(typ, "minecraft:") {
		typ = strings.TrimPrefix(typ, "minecraft:")
	}
	switch typ {
	case "crafting_shaped", "crafting_shapeless", "smelting", "blasting", "smoking", "campfire_cooking", "stonecutting":
		return true
	}
	for key := range data {
		if strings.HasPrefix(key, "minecraft:recipe_") {
			return true
		}
	}
	return false
}

func simpleJavaCubeModel(data map[string]any) bool {
	parent := strings.ToLower(flexibleString(data["parent"]))
	return parent == "minecraft:block/cube_all" || parent == "block/cube_all" || parent == "minecraft:block/cube" || parent == "block/cube"
}
