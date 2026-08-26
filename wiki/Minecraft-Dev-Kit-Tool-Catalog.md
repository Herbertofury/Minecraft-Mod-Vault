# Minecraft Dev Kit Tool Catalog

The Dev Kit catalog is now backed by two complementary machine-readable sources: the live managed-artifact registry for provider-aware updates and the preserved 1.0 provenance inventory for the wider tool collection. The 2.0 release bundle includes the full 59-tool provenance catalog; 47 entries currently map to a canonical GitHub or Modrinth upstream identity. Vendor-only tools remain explicit/manual until a trustworthy updater is known rather than being refreshed from fuzzy search.

## Core toolchains, loaders, and porting

| Tool | Role / update source |
|---|---|
| [Eclipse Temurin / OpenJDK 17](https://adoptium.net/temurin/releases/?version=17) | Java for Forge/Minecraft through 1.20.x; offline Windows/Linux archives retained |
| [Eclipse Temurin / OpenJDK 21](https://adoptium.net/temurin/releases/?version=21) | Exact Java target for Minecraft/NeoForge 1.21.x; **offline Windows/Linux copies are now present** |
| [Eclipse Temurin / OpenJDK 25](https://adoptium.net/temurin/releases/?version=25) | Java target for current 26.x development |
| [Eclipse Temurin / OpenJDK 26](https://adoptium.net/temurin/releases/?version=26) | Newer JDK reference/toolchain |
| [Gradle](https://gradle.org/releases/) 8.8 + 8.14 | Build toolchains; canonical Gradle release source tracked |
| Forge 1.20.1 47.4.23 MDK | Canonical Forge 1.20.1 workspace |
| [NeoForge 1.21.1 ModDevGradle MDK](https://github.com/NeoForgeMDKs/MDK-1.21.1-ModDevGradle) | Preferred streamlined NeoForge 1.21.1 template |
| [NeoForge 1.21.1 NeoGradle MDK](https://github.com/NeoForgeMDKs/MDK-1.21.1-NeoGradle) | NeoGradle/multi-version reference |
| [NeoForge 26.1.2 ModDevGradle MDK](https://github.com/NeoForgeMDKs/MDK-26.1.2-ModDevGradle) | Current-generation NeoForge reference |
| [Fabric Example Mod](https://github.com/FabricMC/fabric-example-mod) | Canonical Fabric template/source |
| [Quilt Template Mod](https://github.com/QuiltMC/quilt-template-mod) | Canonical Quilt template |
| [Architectury Loom](https://github.com/architectury/architectury-loom) | Multi-loader build/reference layer |
| [Architectury API](https://modrinth.com/mod/architectury-api) | Shared cross-loader API; live provider-backed dependency lane |
| [Architectury Templates](https://github.com/architectury/architectury-templates) | Multi-loader starter layouts |
| [Modstitch](https://github.com/isXander/Modstitch) | Cross-loader/multi-version source structure |
| [Stonecutter](https://stonecutter.kikugie.dev/) | Multi-version build/source preprocessing |
| [Parchment](https://parchmentmc.org/) | Forge/NeoForge mappings |
| [Quilt Mappings](https://github.com/QuiltMC/quilt-mappings) | Quilt/Fabric mapping reference |
| [Enigma](https://github.com/FabricMC/Enigma) | Mapping editor/reverse engineering |
| [MixinExtras](https://github.com/LlamaLad7/MixinExtras) | Mixin compatibility/reference tooling |
| [Mojang DataFixerUpper](https://github.com/Mojang/DataFixerUpper) | Serialized/world data migration reference |
| [Quilt Loader](https://github.com/QuiltMC/quilt-loader) | Loader implementation reference |
| [Quilt Config](https://github.com/QuiltMC/quilt-config) | Quilt config reference |
| [Quilted Fabric API](https://github.com/QuiltMC/quilted-fabric-api) | Quilt API compatibility reference |
| [packwiz](https://github.com/packwiz/packwiz) | Modpack management |
| [Ferium](https://github.com/gorilla-devs/ferium) | Mod update/modpack management |
| Omniporter Forge 1.20.1 Gradle cache | Offline Forge 1.20.1 build acceleration |

## Managed dependency & API libraries

These now live in `04 Reference Runtime Mods/Dependency & API Libraries`, with source archives mirrored separately under `02 Build & Port Frameworks/Dependency Source Archives`. Their registry entries store provider identity, target lane, version, and stable Drive file IDs so newer compatible releases can replace the tracked object in place.

| Library | Canonical update source |
|---|---|
| [Fabric API](https://modrinth.com/mod/fabric-api) | Modrinth |
| [Architectury API](https://modrinth.com/mod/architectury-api) | Modrinth |
| [Cloth Config API](https://modrinth.com/mod/cloth-config) | Modrinth |
| [Balm](https://modrinth.com/mod/balm) | Modrinth |
| [Forge Config API Port](https://modrinth.com/mod/forge-config-api-port) | Modrinth |
| [YetAnotherConfigLib / YACL](https://modrinth.com/mod/yacl) | Modrinth |
| [Curios API](https://modrinth.com/mod/curios) | Modrinth |
| [Fabric Language Kotlin](https://modrinth.com/mod/fabric-language-kotlin) | Modrinth |
| [Kotlin for Forge](https://modrinth.com/mod/kotlin-for-forge) | Modrinth |
| [Resourceful Lib](https://modrinth.com/mod/resourceful-lib) | Modrinth |
| [Moonlight Lib](https://modrinth.com/mod/moonlight) | Modrinth |
| [TerraBlender](https://modrinth.com/mod/terrablender) | Modrinth |
| [Citadel](https://modrinth.com/mod/citadel) | Modrinth |
| [Bookshelf](https://modrinth.com/mod/bookshelf-lib) | Modrinth |
| [Puzzles Lib](https://modrinth.com/mod/puzzles-lib) | Modrinth |
| [Iceberg](https://modrinth.com/mod/iceberg) | Modrinth |
| [Cardinal Components API](https://modrinth.com/mod/cardinal-components-api) | Modrinth |
| [Trinkets](https://modrinth.com/mod/trinkets) | Modrinth |
| [CreativeCore](https://modrinth.com/mod/creativecore) | Modrinth |
| [Placebo](https://modrinth.com/mod/placebo) | Modrinth |

Required dependencies discovered from provider metadata can be recursively resolved and automatically enrolled into this managed lane during a Drive-backed synchronization.

## IDE, authoring, assets, and Bedrock

| Tool | Role |
|---|---|
| [IntelliJ IDEA](https://www.jetbrains.com/idea/download/) | Primary Java IDE |
| [Minecraft Development IntelliJ Plugin](https://plugins.jetbrains.com/plugin/8327-minecraft-development) | Minecraft project/loader support |
| [MCreator](https://mcreator.net/) | Generator/reference workflows |
| [Blockbench](https://www.blockbench.net/downloads) | Geometry, pivots, textures, animation |
| [GeckoLib](https://github.com/bernie-g/geckolib) source | Animation API/source reference; canonical upstream retained in provenance catalog |
| [GeckoLib](https://modrinth.com/mod/geckolib) runtime | Forge/Fabric/NeoForge animation runtime provider |
| [Entity Model Features](https://modrinth.com/mod/entity-model-features) | Entity-model compatibility reference |
| [Entity Texture Features](https://modrinth.com/mod/entitytexturefeatures) | Entity-texture compatibility reference |
| [Embeddium](https://modrinth.com/mod/embeddium) | Forge rendering/performance reference |
| [Oculus](https://modrinth.com/mod/oculus) | Forge shader/rendering reference |
| [bridge.](https://bridge-core.app/) | Bedrock behavior/resource-pack IDE |
| [Amulet](https://www.amuletmc.com/) | Cross-edition world tooling |

## Launch and test harnesses

| Tool | Role |
|---|---|
| [Prism Launcher](https://prismlauncher.org/download/) | Preferred isolated test-instance launcher |
| [CurseForge App](https://www.curseforge.com/download/app) | Modpack/launcher compatibility |
| [GDLauncher](https://gdlauncher.com/) | Alternate launcher/modpack reference |
| [LabyMod Launcher](https://www.labymod.net/download) | Client/launcher reference |
| Minecraft Mod Vault TestGrid | Verified execution, log assertions, network/RCON probes, hashes, reports, evidence capture |

## Reverse engineering

| Tool | Role |
|---|---|
| [Vineflower](https://github.com/Vineflower/vineflower) | Preferred deterministic Java decompiler |
| [Recaf](https://github.com/Col-E/Recaf) | Interactive bytecode/decompile/patch workbench |
| [Enigma](https://github.com/FabricMC/Enigma) | Mapping/reverse-engineering workbench |

## Profiling and rendering

| Tool | Role |
|---|---|
| [spark](https://spark.lucko.me/) | In-game/server profiling |
| [async-profiler](https://github.com/async-profiler/async-profiler) | Native CPU/allocation/lock/JFR profiling |
| [VisualVM](https://visualvm.github.io/) | Heap/thread/JVM/JFR inspection |
| VisualVM JFR Streaming Plugin | Live JFR workflow |
| JFR Converter | JFR conversion workflow |
| [RenderDoc](https://renderdoc.org/) | GPU frame capture/render debugging |

## Worlds and NBT

| Tool | Role |
|---|---|
| [NBT Studio](https://github.com/tryashtar/nbt-studio) | Java/Bedrock NBT, SNBT and region inspection; canonical GitHub source retained |
| [MCA Selector](https://github.com/Querz/mcaselector) | Region/chunk selection and world surgery |
| [Amulet](https://www.amuletmc.com/) | Cross-edition world editing/conversion |
| [Mojang DataFixerUpper](https://github.com/Mojang/DataFixerUpper) | Data-version migration reference |

## Update policy

`mmv-devkit` 2.0 is the source/update control plane. It resolves provider identities and newest compatible releases, recursively resolves required dependencies, mirrors matching source, verifies hashes, blocks downgrades by default, and preserves tracked Google Drive file IDs during replacement.

Minecraft Mod Vault TestGrid remains the execution/evidence engine when a workflow needs real builds, launches, assertions, logs, hashes, RCON, or runtime proof. The split avoids two competing automation systems while giving both humans and agents one consistent live source of truth.