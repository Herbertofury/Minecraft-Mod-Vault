# OmniBridge 0.11.0 conversion capability matrix

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
| JVM bytecode/Mixins/ASM | Decompile/remap/semantic reconstruction route; never text-translated | Extract deterministic data/assets only | Dual-decompiler/remap/loader migration route | Reproduce source build and behavior |
| Camera/dialogue/trades/volumes | Preserve Bedrock definitions | Review source | Concrete camera/menu/trade/region surfaces | Client/server behavior tests |
| Biomes/features/worldgen/dimensions | Preserve Bedrock data or generate review contracts | Related Java data lanes where representable | Datagen/registry/biome-modifier/dimension contracts | Seed and placement verification |
| Worlds | Chunker/je2be/Amulet lane | Source world retained | World-template mod/exporter lane | Chunks, entities, inventories, dimensions, packs |
| World templates | `.mctemplate` and companion add-on product | Optional source world in pack family | Standalone world-template mod project | Fresh world, restart and upgrade tests |
| Structures | SchemConvert/world-adapter route | Java structure data when native | Structure processor/placement contract | Palette/block-state placement test |
| Java version/loader changes | Not applicable to Bedrock runtime | Target pack format from Version Atlas | Target Java/loader/mappings/build-tool scaffold | Compile, client, server, persistence |
| Bedrock engine changes | Target minimum engine version and manifest regeneration | Java target is separately versioned | Java target is separately versioned | Content-log and GameTest |

## World & version conversion lanes

World conversion is **layered semantic migration**, not one generic yes/no capability. The full source-backed 2026 reference is [WorldForge — World & Version Conversion Matrix](wiki/WorldForge-Conversion-Matrix.md).

| Conversion lane | Preferred current path / evidence | Required policy | Verification boundary |
|---|---|---|---|
| Java → Bedrock | Chunker + je2be + UMT comparison; Amulet for inspection/exceptional edits | Treat terrain, block entities, entities, players, items, maps, POIs, structures and packs as independent fidelity layers | Parse target + actual Bedrock load; representative dimensions/chunks and semantic torture fixtures |
| Bedrock → Java | Chunker + je2be + UMT comparison; explicit Bedrock actor/XUID provenance | Deterministically map player identity; never conflate converted terrain with converted entities/player inventory | Parse target + actual Java load; UUID/player/inventory/map/POI assertions |
| Java older → newer | **real target Minecraft WorldUpgrader/DataFixerUpper** first | Immutable source, cloned target, per-chunk `DataVersion` audit, mod-specific repair after canonical target upgrade | Mixed-age regions normalized; actual target Java runtime loads and traverses representative content |
| Java newer → older | WorldForge target-capability planner / compatible specialist converter | Never rewrite `DataVersion` alone; enumerate unsupported blocks/items/components/biomes/dimensions/height/worldgen and apply explicit remap/replace/quarantine/drop policy | Fresh target-native world parses and loads in the exact older Java runtime |
| Bedrock older → newer | Bedrock storage-era adapter + real target-aware translator | Detect current LevelDB vs older `StorageVersion <= 4` terrain vs pre-LevelDB Pocket `chunks.dat` where present | Target Bedrock parser + actual target Bedrock load |
| Bedrock newer → older | Target-aware downgrade planner | Never rewrite `StorageVersion` alone; encode target-valid chunk/entity/player records and rebuild derived data | Exact older target loads; unsupported-content ledger matches policy |
| Legacy Console → Java/Bedrock | LegacyEditor + je2be + UMT evidence | Record exact platform, title-update era, container/encryption/signing requirements and read/write capability | Source decode + target load + terrain/entity/player fixture assertions |
| Modded save → newer/different runtime | target-native upgrade first, then mc-world-migrator-style source-vs-target semantic repair | Original/pre-upgrade save is evidence; match block entities by position, entities by UUID; preserve unknown registry data | Target runtime plus semantic diff and unresolved-registry report |
| Server → singleplayer | selected Java UUID / Bedrock XUID migration | User-selected player becomes local player; inventory/state/position/dimension preserved; never choose arbitrary first record | Fresh singleplayer load as selected identity |
| Structure/build conversion | common Structure IR; Bloxelizer / Litematica Viewer / Structure2Schematic references | Palette/block-state remap, block entities/entities, origin/offset, transforms, unsupported report | Round-trip fixtures across `.schematic`, `.schem`, `.litematic`, Java structure `.nbt`, `.mcstructure` |
| Java RP → Bedrock/RTX | specialist resource-pack translator such as JE2BE Resource Pack Converter | Texture/material mapping separate from world conversion; preserve unsupported feature report | Target pack loads; visual/material fixture comparison |

### Current tool boundaries that must remain explicit

- **Chunker:** best free/open broad Java ↔ Bedrock baseline, but current official scope does not cross-edition-convert entities or player inventories; do not turn terrain success into a blanket world-success claim.
- **Universal Minecraft Tool:** high-capability commercial parity benchmark; vendor claims remain fixture-tested rather than accepted as proof.
- **je2be:** independent open-source Java/Bedrock/legacy comparator and conversion lane; maintain round-trip fixtures.
- **Amulet:** strong multi-version editor/translation architecture, not a complete semantic converter where entities/items remain unsupported or limited.
- **DataFixerUpper:** canonical Java forward migration primitive, never a reverse/downgrade engine.
- **PaperMC DataConverter:** useful high-performance architecture; standalone generic modded conversion can miss other mods' datafixers.
- **LegacyEditor:** Legacy Console read/write support is platform/save-format/title-update specific and must never be flattened to “all Legacy supported.”
- **bedrock-world / bedrock-leveldb:** strong Bedrock semantic/storage foundations, but target gameplay policy and end-to-end conversion still belong to WorldForge.

### Canonical World IR

Avoid N×N pairwise translators:

`source adapter → version/edition semantic IDs → canonical World IR → target capability/fallback policy → target adapter → derived-data repair → verifier`

Every IR node retains source edition/version, original identifier/value, provenance and conversion confidence. Unknown/custom fields are preserved in-place where safe or written to provenance/quarantine sidecars rather than silently discarded.

### Independent world fidelity verdicts

A conversion report tracks at least:

- terrain/chunk sections/block states/height bounds;
- biomes;
- block entities and containers;
- entities;
- players/inventory/ender chest;
- items/components/enchantments/books;
- maps;
- dimensions/portals;
- villages/trades/reputation/POIs;
- structures/references/jigsaws;
- raids/scheduled ticks;
- heightmaps/lighting;
- gamerules/difficulty/spawn;
- scoreboard/teams;
- advancements/achievements;
- command blocks/functions/datapacks;
- redstone/fluids/waterlogging semantics;
- Java UUID ↔ Bedrock actor ID/XUID mapping;
- resource/behavior-pack dependencies;
- custom/modded registry identifiers.

## Conversion levels

- **Exact:** source bytes or same-schema data preserved, with target indexes/metadata regenerated when necessary.
- **Translated:** a deterministic schema/path converter exists and is tested.
- **Generated:** a valid target artifact or source implementation surface is generated, but semantic parity still requires target tests.
- **Tool-assisted:** a reviewed specialist converter must run and produce a validated artifact.
- **Review:** related target concepts exist, but automatic equivalence cannot be proven.
- **Blocked:** no direct target representation exists; the original and a reimplementation contract are preserved.
