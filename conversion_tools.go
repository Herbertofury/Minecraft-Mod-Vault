package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// conversionToolCatalog is deliberately a capability catalog rather than an
// endorsement list.  OmniBridge only marks an adapter executable when it has a
// fixed, reviewed command contract; research-only tools remain visible without
// being run as arbitrary code.
func conversionToolCatalog() []ConversionToolAdapter {
	return []ConversionToolAdapter{
		{ID: "chunker", Name: "Chunker / ChunkerCLI", Role: "Cross-edition Java ↔ Bedrock world conversion and Minecraft-version world upgrades/downgrades.", Formats: []string{"world", "mcworld", "mctemplate"}, Directions: []string{"java-to-bedrock", "bedrock-to-java", "version-to-version"}, Maturity: "production", License: "MIT", OfficialURL: "https://oss.chunker.app/", RepositoryURL: "https://github.com/HiveGamesOSS/Chunker", Executable: "chunker-cli", CanExecute: true},
		{ID: "je2be", Name: "je2be-core", Role: "Local Java ↔ Bedrock world conversion, including legacy console formats.", Formats: []string{"world", "mcworld"}, Directions: []string{"java-to-bedrock", "bedrock-to-java"}, Maturity: "production-library", License: "GPL-3.0", OfficialURL: "https://je2be.com/", RepositoryURL: "https://github.com/kbinani/je2be-core", Executable: "je2be"},
		{ID: "amulet", Name: "Amulet Map Editor", Role: "World editing and conversion across Java 1.12+ and Bedrock 1.7+.", Formats: []string{"world", "structure"}, Directions: []string{"java-to-bedrock", "bedrock-to-java", "version-to-version"}, Maturity: "production", OfficialURL: "https://www.amuletmc.com/", RepositoryURL: "https://github.com/Amulet-Team/Amulet-Map-Editor", Executable: "amulet_app"},
		{ID: "packconverter", Name: "Geyser PackConverter / Thunder", Role: "Java resource-pack → Bedrock resource-pack conversion. Upstream labels it work-in-progress and does not generate all custom-item mappings.", Formats: []string{"java-resourcepack", "bedrock-resource"}, Directions: []string{"java-to-bedrock"}, Maturity: "early-beta", License: "MIT", OfficialURL: "https://geysermc.org/wiki/other/thunder/", RepositoryURL: "https://github.com/GeyserMC/PackConverter", CanExecute: true},
		{ID: "je2be-resource", Name: "JE2BE Resource Pack Converter", Role: "Java resource pack → Bedrock .mcpack conversion with PBR/LabPBR-to-MER and RTX-oriented mappings.", Formats: []string{"java-resourcepack", "bedrock-resource", "pbr", "rtx"}, Directions: []string{"java-to-bedrock"}, Maturity: "active", License: "MIT", OfficialURL: "https://github.com/Seraphic-Studio/JE2BE-Resource-Pack-Converter", RepositoryURL: "https://github.com/Seraphic-Studio/JE2BE-Resource-Pack-Converter", Executable: "je2be", CanExecute: true},
		{ID: "rainbow", Name: "Geyser Rainbow", Role: "Generate Bedrock resource packs and Geyser mappings for custom Java blocks, items, models and sounds.", Formats: []string{"java-resourcepack", "bedrock-resource", "geyser-mappings"}, Directions: []string{"java-to-bedrock-runtime"}, Maturity: "experimental", OfficialURL: "https://geysermc.org/wiki/other/rainbow/", RepositoryURL: "https://github.com/GeyserMC/Rainbow"},
		{ID: "ego-converter-plus", Name: "EgoConverter++", Role: "Modern Java model/resource conversion and Geyser custom-item mapping reference for current item-model formats.", Formats: []string{"java-resourcepack", "bedrock-resource", "geyser-mappings"}, Directions: []string{"java-to-bedrock"}, Maturity: "active-community", OfficialURL: "https://github.com/ego-smp-labs/ego-converter-plus", RepositoryURL: "https://github.com/ego-smp-labs/ego-converter-plus"},
		{ID: "convert-pack", Name: "convert-pack", Role: "Version-aware Java/Bedrock resource-pack conversion reference with multi-version normalization.", Formats: []string{"java-resourcepack", "bedrock-resource"}, Directions: []string{"java-to-bedrock", "bedrock-to-java", "version-to-version"}, Maturity: "community", OfficialURL: "https://github.com/3vorp/convert-pack", RepositoryURL: "https://github.com/3vorp/convert-pack"},
		{ID: "resourcepack-versioner", Name: "ResourcePackConverter", Role: "Java resource-pack migration reference covering legacy and modern pack layouts across Minecraft versions.", Formats: []string{"java-resourcepack"}, Directions: []string{"version-to-version"}, Maturity: "community", OfficialURL: "https://github.com/agentdid127/ResourcePackConverter", RepositoryURL: "https://github.com/agentdid127/ResourcePackConverter"},
		{ID: "schemconvert", Name: "SchemConvert", Role: "Structure conversion among NBT, Sponge schematics, Litematica and supported blueprints.", Formats: []string{"nbt", "schem", "litematic", "bp"}, Directions: []string{"structure-to-structure"}, Maturity: "active", OfficialURL: "https://github.com/SchemConvert/SchemConvert", RepositoryURL: "https://github.com/SchemConvert/SchemConvert", Executable: "schemconvert"},
		{ID: "datafixerupper", Name: "Mojang DataFixerUpper", Role: "Incremental Java game-data schema transforms used by Minecraft itself.", Formats: []string{"java-world-data", "nbt"}, Directions: []string{"version-to-version"}, Maturity: "production-library", License: "LGPL-2.1", OfficialURL: "https://github.com/Mojang/DataFixerUpper", RepositoryURL: "https://github.com/Mojang/DataFixerUpper"},
		{ID: "tiny-remapper", Name: "Fabric Tiny Remapper", Role: "Fast Java JAR namespace remapping for loader and version migration pipelines.", Formats: []string{"jar"}, Directions: []string{"mapping-to-mapping"}, Maturity: "production-library", OfficialURL: "https://github.com/FabricMC/tiny-remapper", RepositoryURL: "https://github.com/FabricMC/tiny-remapper"},
		{ID: "vineflower", Name: "Vineflower", Role: "Primary reproducible Java decompiler for source recovery from binary-only mods; output must be compared with bytecode and an independent decompiler.", Formats: []string{"jar", "java-bytecode"}, Directions: []string{"jar-to-source"}, Maturity: "production", License: "Apache-2.0", OfficialURL: "https://vineflower.org/", RepositoryURL: "https://github.com/Vineflower/vineflower"},
		{ID: "cfr", Name: "CFR", Role: "Independent Java decompiler used as a disagreement detector during closed-source reconstruction.", Formats: []string{"jar", "java-bytecode"}, Directions: []string{"jar-to-source"}, Maturity: "production", License: "MIT", OfficialURL: "https://www.benf.org/other/cfr/", RepositoryURL: "https://github.com/leibnitz27/cfr"},
		{ID: "fabric-loom", Name: "Fabric Loom", Role: "Deobfuscated development, remapping, decompilation and build environment for Fabric projects.", Formats: []string{"java-source", "jar"}, Directions: []string{"source-to-fabric", "version-to-version"}, Maturity: "production", OfficialURL: "https://docs.fabricmc.net/develop/loom/", RepositoryURL: "https://github.com/FabricMC/fabric-loom"},
		{ID: "modstitch", Name: "Modstitch", Role: "Unified Fabric and (Neo)Forge official tooling with metadata and access-rule translation.", Formats: []string{"java-source"}, Directions: []string{"source-to-multiloader"}, Maturity: "unstable-active", OfficialURL: "https://isxander.github.io/modstitch-docs/", RepositoryURL: "https://github.com/isXander/modstitch"},
		{ID: "stonecutter", Name: "Stonecutter", Role: "Multi-version source preprocessing for a shared Minecraft mod codebase.", Formats: []string{"java-source"}, Directions: []string{"version-to-version"}, Maturity: "production", OfficialURL: "https://stonecutter.kikugie.dev/", RepositoryURL: "https://github.com/Kikugie/Stonecutter"},
		{ID: "architectury", Name: "Architectury Loom/API", Role: "Shared Java mod code with Fabric, Forge, NeoForge and Quilt loader targets.", Formats: []string{"java-source"}, Directions: []string{"loader-to-loader", "source-to-multiloader"}, Maturity: "production", OfficialURL: "https://docs.architectury.dev/", RepositoryURL: "https://github.com/architectury/architectury-loom"},
		{ID: "regolith", Name: "Regolith", Role: "Compiler-style pipeline for modular Bedrock add-on projects. OmniBridge uses an exact export path inside the isolated adapter workspace.", Formats: []string{"bedrock-source", "mcaddon"}, Directions: []string{"source-to-bedrock"}, Maturity: "production", License: "MIT", OfficialURL: "https://bedrock-oss.github.io/regolith/", RepositoryURL: "https://github.com/Bedrock-OSS/regolith", Executable: "regolith", CanExecute: true},
		{ID: "bridge", Name: "bridge.", Role: "Bedrock add-on editor with schemas, compiler extensions and project tooling.", Formats: []string{"bedrock-source", "mcaddon"}, Directions: []string{"source-to-bedrock"}, Maturity: "production", OfficialURL: "https://bridge-core.app/", RepositoryURL: "https://github.com/bridge-core/editor"},
		{ID: "beet", Name: "beet + mecha", Role: "Python toolchain for Java data/resource packs and command validation.", Formats: []string{"java-datapack", "java-resourcepack"}, Directions: []string{"source-to-java-pack", "version-to-version"}, Maturity: "production", OfficialURL: "https://mcbeet.dev/", RepositoryURL: "https://github.com/mcbeet/beet"},
		{ID: "portkit", Name: "PortKit", Role: "Experimental AI-assisted Java mod → Bedrock add-on translation reference. Results remain review-required.", Formats: []string{"jar", "java-source", "mcaddon"}, Directions: []string{"java-to-bedrock"}, Maturity: "beta-experimental", License: "MIT", OfficialURL: "https://github.com/anchapin/portkit", RepositoryURL: "https://github.com/anchapin/portkit"},
		{ID: "modmorpher", Name: "ModMorpher", Role: "Experimental decompile-and-scaffold reference for Java mod → Bedrock add-on conversion. Results remain review-required.", Formats: []string{"jar", "mcaddon"}, Directions: []string{"java-to-bedrock"}, Maturity: "experimental", OfficialURL: "https://github.com/Indozilla1234/Modmorpher", RepositoryURL: "https://github.com/Indozilla1234/Modmorpher"},
	}
}

func conversionToolAdapters() []ConversionToolAdapter {
	tools := conversionToolCatalog()
	for index := range tools {
		tools[index] = detectConversionTool(tools[index], "")
	}
	return tools
}

func (a *App) configuredConversionToolAdapters() []ConversionToolAdapter {
	a.mu.RLock()
	configured := make(map[string]string, len(a.settings.ConversionToolPaths))
	for key, value := range a.settings.ConversionToolPaths {
		configured[key] = value
	}
	a.mu.RUnlock()
	tools := conversionToolCatalog()
	for index := range tools {
		tools[index] = detectConversionTool(tools[index], configured[tools[index].ID])
	}
	return tools
}

func detectConversionTool(tool ConversionToolAdapter, explicit string) ConversionToolAdapter {
	if configured := strings.TrimSpace(explicit); configured != "" {
		tool.Configured = true
		if info, err := os.Lstat(configured); err == nil && info.Mode().IsRegular() && info.Mode()&os.ModeSymlink == 0 {
			tool.DetectedPath, tool.Ready = filepath.Clean(configured), true
			return tool
		}
		tool.Notes = append(tool.Notes, "Configured path is unavailable or is not a regular file.")
	}
	key := "MMV_TOOL_" + strings.ToUpper(strings.NewReplacer("-", "_", ".", "_").Replace(tool.ID))
	if configured := strings.TrimSpace(os.Getenv(key)); configured != "" {
		if info, err := os.Lstat(configured); err == nil && info.Mode().IsRegular() && info.Mode()&os.ModeSymlink == 0 {
			tool.DetectedPath, tool.Ready = filepath.Clean(configured), true
			return tool
		}
	}
	if tool.Executable != "" {
		if path, err := exec.LookPath(tool.Executable); err == nil {
			tool.DetectedPath, tool.Ready = path, true
		}
	}
	if !tool.Ready {
		for _, candidate := range conversionToolCandidates(tool.ID) {
			if info, err := os.Lstat(candidate); err == nil && info.Mode().IsRegular() && info.Mode()&os.ModeSymlink == 0 {
				tool.DetectedPath, tool.Ready = candidate, true
				break
			}
		}
	}
	return tool
}

func conversionToolCandidates(id string) []string {
	home, _ := os.UserHomeDir()
	var paths []string
	switch id {
	case "chunker":
		paths = []string{filepath.Join(home, "Chunker", "chunker-cli.jar"), filepath.Join(home, "Chunker", "Chunker.CLI"), filepath.Join(home, "Chunker", "Chunker.CLI.exe")}
	case "je2be":
		paths = []string{filepath.Join(home, "je2be", "je2be"), filepath.Join(home, "je2be", "je2be.exe")}
	case "je2be-resource":
		paths = []string{filepath.Join(home, "JE2BE-Resource-Pack-Converter", "je2be.exe"), filepath.Join(home, "JE2BE-Resource-Pack-Converter", "je2be_converter.py")}
	case "packconverter":
		paths = []string{filepath.Join(home, "PackConverter", "Thunder.jar"), filepath.Join(home, "Thunder.jar")}
	case "amulet":
		paths = []string{filepath.Join(home, "Amulet", "amulet_app.exe"), filepath.Join(home, ".local", "bin", "amulet_app")}
	case "schemconvert":
		paths = []string{filepath.Join(home, "SchemConvert", "schemconvert"), filepath.Join(home, "SchemConvert", "schemconvert.exe")}
	case "regolith":
		paths = []string{filepath.Join(home, "Regolith", "regolith"), filepath.Join(home, "Regolith", "regolith.exe"), filepath.Join(home, ".local", "bin", "regolith")}
	}
	return paths
}

func conversionToolMap() map[string]ConversionToolAdapter {
	out := map[string]ConversionToolAdapter{}
	for _, tool := range conversionToolAdapters() {
		out[tool.ID] = tool
	}
	return out
}
