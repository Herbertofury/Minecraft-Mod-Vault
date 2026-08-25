# OmniBridge 0.11.0 conversion capability matrix

This matrix covers **mods/add-ons/content plus saved worlds**. General mod/version/loader conversion is governed by [Mod Conversion Stack](wiki/Mod-Conversion-Stack.md); saved-world conversion is governed by [WorldForge — World & Version Conversion Matrix](wiki/WorldForge-Conversion-Matrix.md).

| Source capability | Bedrock target | Java vanilla pack target | Java loader project target | Verification boundary |
|---|---|---|---|---|
| Textures and icons | Copy, normalize paths, rebuild texture indexes | Copy and rebuild Java asset paths | Copy into `assets/<namespace>` | Visual parity in target client |
| Languages | Convert key/value representation | Convert to Java locale JSON | Convert and preserve original | Formatting/fallback review |
| Common recipes | Translate shaped, shapeless, and furnace-family schemas | Translate to target Java recipe path | Translate into project data resources | Registry IDs and custom serializers |
| Models | Copy simple assets; specialist pack adapters for richer Java models | Preserve/normalize Java models | Preserve plus renderer/model contract | Predicates, transforms, geometry review |
| Sounds | Copy and preserve sound files | Copy Java sound assets | Copy plus sound-event surface | Sound index/event parity |
| Particles/animations/controllers | Preserve Bedrock-native files; generate target manifests | Review contracts | Client particle/animation implementation surfaces | Real client rendering test |
| Blocks/items/entities | Generate Bedrock definitions or preserve native definitions | Related recipes/assets only | Registry, component, attributes, goals and renderer contracts | Target-native gameplay implementation |
| Functions and commands | Preserve native functions or retain migration source | Command migration review source | Server command/function contracts | Every command tested in target edition |
| Loot/tags/advancements/predicates | Preserve or review target schema | Preserve native Java; Bedrock source retained for translation | Data and serializer/trigger contracts | Schema and runtime semantics |
| Bedrock Script API | Preserve for Bedrock and pin Script API manifest | Cannot execute in vanilla Java pack | Event/tick/command/network/persistence contracts and original JS | Implement and test logical sides |
| Java source | Script/data contracts for Bedrock | Extract data/resource layers | Loader/version migration workspace | Compile and runtime test |
| JVM bytecode/Mixins/ASM | Semantic Behavior IR; no direct bytecode translation | Extract deterministic data/assets only | remap/decompile/Mixin/ASM semantic reconstruction | Production-JAR transform + actual behavior |
| Access transformers/wideners/tweakers | Recreate behavior if needed; no direct Bedrock equivalent | N/A | canonical Access IR → target AT/AW/ClassTweaker | target member mapped + runtime behavior |
| Loader lifecycle/events | Behavior IR → Bedrock components/Script API | N/A | target-native loader lifecycle/events or proven abstraction | client + dedicated server behavior |
| Networking | Script/API/state sync contract | N/A | packet schema/direction/phase/logical-side migration | two-ended runtime network test |
| Config/persistence/components/capabilities | target Bedrock property/dynamic-property/script storage | N/A | target loader/game persistence model | save/reload/reconnect regression |
| Camera/dialogue/trades/volumes | Preserve Bedrock definitions | Review source | Concrete camera/menu/trade/region surfaces | Client/server behavior tests |
| Biomes/features/worldgen/dimensions | Preserve Bedrock data or generate review contracts | Related Java data lanes where representable | Datagen/registry/biome-modifier/dimension contracts | Seed and placement verification |
| Worlds | Chunker/je2be/Amulet lane | Source world retained | World-template mod/exporter lane | Chunks, entities, inventories, dimensions, packs |
| World templates | `.mctemplate` and companion add-on product | Optional source world in pack family | Standalone world-template mod project | Fresh world, restart and upgrade tests |
| Structures | common Structure IR / specialist adapters | Java structure data when native | Structure processor/placement contract | Palette/block-state placement test |
| Java version/loader changes | N/A to Bedrock runtime unless cross-edition port | Target pack format from Version Atlas | Modstitch/Stonecutter + mapping/Mixin/access/API migration | compile + client + server + persistence + behavior diff |
| Bedrock engine changes | Target minimum engine version and manifest regeneration | Java target separately versioned | Java target separately versioned | Content log + GameTest/runtime tests |

## General mod / add-on conversion lanes

| Conversion lane | Preferred current path / evidence | Required policy | Verification boundary |
|---|---|---|---|
| Java version → Java version | Modstitch + Stonecutter; Loom/Ravel/ModDevGradle; Mapping-IO/Tiny Remapper/Matcher | map namespaces first, then migrate APIs, Mixins, access rules, registries, networking, rendering, persistence and dependencies | clean target build + production JAR client/server differential tests |
| Forge ↔ NeoForge | target-native APIs / ModDevGradle; shared abstractions only where exact | lifecycle/events/registries/networking/capability/component semantics explicitly migrated | real client + dedicated server + save/reload |
| Forge/NeoForge → Fabric | native semantic port; Porting Lib/Architectury where useful; Kilt only as experimental reference | AT→AW/ClassTweaker through Access IR; loader APIs not regex-renamed | Fabric production JAR + behavior comparison |
| Fabric → Forge/NeoForge | native semantic port; FFAPI/Connector/Launchpad as compatibility references | Yarn/Intermediary/Mojang mappings, Mixins/refmaps, AW/ClassTweaker and Fabric API dependencies handled explicitly | target production JAR + behavior comparison |
| pre-26.1 Fabric → 26.1+ | Fabric mapping migration with Loom/Ravel before API migration | account for Yarn/Intermediary → Mojang and the 26.1 unobfuscated boundary; manually review Mixins | compiler + Mixin application + runtime behavior |
| JAR-only Java mod → target Java | identify/remap → Vineflower + CFR/Recaf independent decompile → reconstruct build → semantic port | decompiled code is evidence, not original source; preserve disagreements and bytecode provenance | reconstructed clean build + runtime differential tests |
| Mixin → new version/loader | owner/name/descriptor + injection fingerprint + Matcher/source diff + semantic redesign | never accept refmap/name remap alone when target control flow changed | injector application count + affected gameplay behavior |
| AT ↔ AW/ClassTweaker | Modstitch accessx / canonical Access IR | map target first, then translate visibility/extendability/mutability intent | target dev + production JAR verification |
| Java mod → Bedrock add-on | Behavior IR + Geyser mappings/generator + Creator schemas; Rainbow/Blockbench for applicable assets | reimplement behavior as Bedrock components/controllers/Script API; no bytecode→JSON fiction | Bedrock content log + GameTest/manual runtime parity fixtures |
| Bedrock add-on → Java mod | parse BP/RP/Script API/Molang → Behavior IR → Java APIs/Mixins; Blockbench/GeckoLib for animations | preserve component/event/script intent and mark unmappable semantics | Java client/server runtime parity fixtures |
| Java custom blocks/items/resource pack → Bedrock | Geyser mappings + Rainbow + specialist pack lanes | generated mappings/assets are separate from behavior conversion | Java/Geyser/Bedrock representation comparison |

### Mapping & bytecode infrastructure

- **Mapping-IO** — canonical mapping interchange/model.
- **Tiny Remapper** — binary namespace transformation.
- **Matcher** — structural cross-version class/member correspondence.
- **Loom/Ravel** — source mapping migration, with Ravel preferred for Kotlin/complex IDE-aware cases.
- **ForgeGradle/MCPConfig/SrgUtils** — Forge/SRG/TSRG-era archaeology/reobfuscation.
- **ModDevGradle/NeoForm Runtime** — modern NeoForge and legacy-Forge bridge/source reconstruction.
- **Mixin + MixinExtras** — narrow behavior interception; target control flow remains part of the semantic contract.
- **Vineflower + CFR/Recaf** — independent JAR-only archaeology/decompilation pipeline.

### Canonical Semantic Port IR

Avoid N×N import/regex translators:

`source/JAR → evidence + namespaces → Semantic Port IR → target capability/fallback policy → deterministic transforms → target-native implementation → package → differential verifier`

The Port IR retains source/target versions/loaders/mappings, symbol provenance, registry IDs, lifecycle intent, Mixin target fingerprints, access intent, dependencies, packet schemas, persistence/data components/capabilities, config schema, rendering/model/animation contracts, Bedrock component/Molang/Script API behavior, generated assets and unresolved-confidence state.

## World & version conversion lanes

World conversion is **layered semantic migration**, not one generic yes/no capability. The full source-backed 2026 reference is [WorldForge — World & Version Conversion Matrix](wiki/WorldForge-Conversion-Matrix.md).

| Conversion lane | Preferred current path / evidence | Required policy | Verification boundary |
|---|---|---|---|
| Java → Bedrock | Chunker + je2be + UMT comparison; Amulet for inspection/exceptional edits | Treat terrain, block entities, entities, players, items, maps, POIs, structures and packs as independent fidelity layers | Parse target + actual Bedrock load; representative dimensions/chunks and semantic torture fixtures |
| Bedrock → Java | Chunker + je2be + UMT comparison; explicit Bedrock actor/XUID provenance | Deterministically map player identity; never conflate converted terrain with converted entities/player inventory | Parse target + actual Java load; UUID/player/inventory/map/POI assertions |
| Java older → newer | real target Minecraft WorldUpgrader/DataFixerUpper first | immutable source, cloned target, per-chunk `DataVersion` audit, mod-specific repair after canonical target upgrade | mixed-age regions normalized; actual target Java runtime loads |
| Java newer → older | WorldForge target-capability planner / compatible specialist converter | never rewrite `DataVersion` alone; enumerate unsupported content and apply explicit policy | fresh target-native world parses and loads in exact older Java runtime |
| Bedrock older → newer | Bedrock storage-era adapter + real target-aware translator | detect current LevelDB vs older storage eras | target parser + actual target Bedrock load |
| Bedrock newer → older | target-aware downgrade planner | never rewrite `StorageVersion` alone; encode target-valid records and rebuild derived data | exact older target load |
| Legacy Console → Java/Bedrock | LegacyEditor + je2be + UMT evidence | record exact platform/title-update/container/read-write support | source decode + target load fixture assertions |
| Modded save → newer/different runtime | target-native upgrade first, then source-vs-target semantic repair | original/pre-upgrade save is evidence | target runtime + semantic diff |
| Server → singleplayer | selected Java UUID / Bedrock XUID migration | preserve selected player state | fresh singleplayer load as selected identity |
| Structure/build conversion | common Structure IR | palette/block-state remap, entities, block entities, transforms, unsupported report | round-trip format fixtures |

## Conversion levels

- **Exact:** source bytes or same-schema semantics preserved, with target indexes/metadata regenerated when necessary.
- **Translated:** deterministic semantic/schema/path converter exists and is tested.
- **Generated:** valid target implementation surface exists, but parity still requires target tests.
- **Tool-assisted:** reviewed specialist converter must run and produce a validated artifact.
- **Review:** related target concepts exist, but automatic equivalence cannot be proven.
- **Blocked:** no direct target representation exists; preserve source and an explicit reimplementation contract.
