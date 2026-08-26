# Minecraft Dev Kit Tool Catalog

Generated from the orchestrator registry. The registry is the machine-readable source of truth; this page is the human-readable mirror.

## Core toolchains, loaders, and porting

| Tool | Role |
|---|---|
| [Eclipse Temurin / OpenJDK 17](https://adoptium.net/temurin/releases/?version=17) | Java for Forge/Minecraft through 1.20.x |
| [Eclipse Temurin / OpenJDK 21](https://adoptium.net/temurin/releases/?version=21) | Exact Java target for Minecraft/NeoForge 1.21.x; registry-supported but currently missing from the offline Drive kit |
| [Eclipse Temurin / OpenJDK 25](https://adoptium.net/temurin/releases/?version=25) | Java target for current 26.x development |
| [Eclipse Temurin / OpenJDK 26](https://adoptium.net/temurin/releases/?version=26) | Newer JDK reference/toolchain |
| [Gradle](https://gradle.org/releases/) 8.8 + 8.14 | Build toolchains |
| Forge 1.20.1 47.4.23 MDK | Canonical Forge 1.20.1 workspace |
| [NeoForge 1.21.1 ModDevGradle MDK](https://github.com/NeoForgeMDKs/MDK-1.21.1-ModDevGradle) | Preferred streamlined NeoForge 1.21.1 template |
| [NeoForge 1.21.1 NeoGradle MDK](https://github.com/NeoForgeMDKs/MDK-1.21.1-NeoGradle) | NeoGradle/multi-version reference |
| [NeoForge 26.1.2 ModDevGradle MDK](https://github.com/NeoForgeMDKs/MDK-26.1.2-ModDevGradle) | Current-generation NeoForge reference |
| [Fabric Example Mod](https://github.com/FabricMC/fabric-example-mod) 26.2 | Canonical Fabric template |
| [Quilt Template Mod](https://github.com/QuiltMC/quilt-template-mod) 1.20.6 | Canonical Quilt template |
| [Architectury Loom](https://github.com/architectury/architectury-loom) 1.17 | Multi-loader build/reference layer |
| [Architectury API](https://github.com/architectury/architectury-api) | Shared cross-loader API reference |
| [Architectury Templates](https://github.com/architectury/architectury-templates) | Multi-loader starter layouts |
| [Modstitch](https://github.com/isXander/Modstitch) | Cross-loader/multi-version source structure |
| [Stonecutter](https://stonecutter.kikugie.dev/) 0.10 | Multi-version build/source preprocessing |
| [Parchment](https://parchmentmc.org/) 1.21.x | Forge/NeoForge mappings |
| [Quilt Mappings](https://github.com/QuiltMC/quilt-mappings) | Quilt/Fabric mapping reference |
| [Enigma](https://github.com/FabricMC/Enigma) | Mapping editor/reverse engineering |
| [MixinExtras](https://github.com/LlamaLad7/MixinExtras) 0.5.4 | Mixin compatibility/reference tooling |
| [Mojang DataFixerUpper](https://github.com/Mojang/DataFixerUpper) | Serialized/world data migration reference |
| [Quilt Loader](https://github.com/QuiltMC/quilt-loader) | Loader implementation reference |
| [Quilt Config](https://github.com/QuiltMC/quilt-config) | Quilt config reference |
| [Quilted Fabric API](https://github.com/QuiltMC/quilted-fabric-api) 1.21.1 | Older-version Quilt API compatibility reference |
| [packwiz](https://github.com/packwiz/packwiz) | Modpack management |
| [Ferium](https://github.com/gorilla-devs/ferium) | Mod update/modpack management |
| Omniporter Forge 1.20.1 Gradle cache | Offline Forge 1.20.1 build acceleration |

## IDE, authoring, assets, and Bedrock

| Tool | Role |
|---|---|
| [IntelliJ IDEA](https://www.jetbrains.com/idea/download/) 2026.2.1 | Primary Java IDE |
| [Minecraft Development IntelliJ Plugin](https://plugins.jetbrains.com/plugin/8327-minecraft-development) | Minecraft project/loader support |
| [MCreator](https://mcreator.net/) 2026.2 + 1.20.1 generator | Generator/reference workflows |
| [Blockbench](https://www.blockbench.net/downloads) 5.1.6 portable | Geometry, pivots, textures, animation |
| [GeckoLib](https://github.com/bernie-g/geckolib) source | Animation API/runtime reference |
| [GeckoLib](https://modrinth.com/mod/geckolib) 4.8.3 Forge 1.20.1 | Forge animation runtime reference |
| [GeckoLib](https://modrinth.com/mod/geckolib) 4.9.2 NeoForge 1.21.1 | NeoForge animation runtime reference |
| [Entity Model Features](https://modrinth.com/mod/entity-model-features) 3.2.4 | Entity-model compatibility reference |
| [Entity Texture Features](https://modrinth.com/mod/entitytexturefeatures) 7.1 | Entity-texture compatibility reference |
| [Embeddium](https://modrinth.com/mod/embeddium) 0.3.31 | Forge rendering/performance reference |
| [Oculus](https://modrinth.com/mod/oculus) 1.8.0 | Forge shader/rendering reference |
| [bridge.](https://bridge-core.app/) 3.0.4 | Bedrock behavior/resource-pack IDE |
| [Amulet](https://www.amuletmc.com/) 0.10.62 | Cross-edition world tooling |

## Launch and test harnesses

| Tool | Role |
|---|---|
| [Prism Launcher](https://prismlauncher.org/download/) 11.0.3 | Preferred isolated test-instance launcher |
| [CurseForge App](https://www.curseforge.com/download/app) | Modpack/launcher compatibility |
| [GDLauncher](https://gdlauncher.com/) 2.0.40 | Alternate launcher/modpack reference |
| [LabyMod Launcher](https://www.labymod.net/download) | Client/launcher reference |
| Minecraft Mod Vault TestGrid | Verified execution, log assertions, network/RCON probes, hashes, reports, evidence capture |

## Reverse engineering

| Tool | Role |
|---|---|
| [Vineflower](https://github.com/Vineflower/vineflower) 1.12.0 + source | Preferred deterministic Java decompiler |
| [Recaf](https://github.com/Col-E/Recaf) 4 alpha + launchers/source | Interactive bytecode/decompile/patch workbench |
| [Enigma](https://github.com/FabricMC/Enigma) | Mapping/reverse-engineering workbench |

## Profiling and rendering

| Tool | Role |
|---|---|
| [spark](https://spark.lucko.me/) Forge 1.10.53 | In-game Forge profiling |
| [spark](https://spark.lucko.me/) NeoForge 1.10.173 | In-game NeoForge profiling |
| [async-profiler](https://github.com/async-profiler/async-profiler) 4.5 | Native CPU/allocation/lock/JFR profiling |
| [VisualVM](https://visualvm.github.io/) 2.2.1 | Heap/thread/JVM/JFR inspection |
| VisualVM JFR Streaming Plugin | Live JFR workflow |
| JFR Converter | JFR conversion workflow |
| [RenderDoc](https://renderdoc.org/) | GPU frame capture/render debugging |

## Worlds and NBT

| Tool | Role |
|---|---|
| [NBT Studio](https://github.com/tryashtar/nbt-studio) 1.15.3 | Java/Bedrock NBT, SNBT and region inspection |
| [MCA Selector](https://github.com/Querz/mcaselector) 2.8 | Region/chunk selection and world surgery |
| [Amulet](https://www.amuletmc.com/) 0.10.62 | Cross-edition world editing/conversion |
| [Mojang DataFixerUpper](https://github.com/Mojang/DataFixerUpper) | Data-version migration reference |

## Orchestration rule

`mmv-devkit` is the single resolver for selecting/preparing these tools. Minecraft Mod Vault TestGrid remains the execution/evidence engine when a workflow needs real builds, launches, assertions, logs, hashes, RCON, or runtime proof. This avoids two competing automation systems while giving both humans and agents one consistent tool registry.
