# 🔄 WorldForge — World & Version Conversion Matrix

> **Status:** 📋 Conversion architecture / parity contract. This page is additive to [Conversion Workflows](Conversion-Workflows), [Java ↔ Bedrock](Java-Bedrock), [Repair & Forensics](WorldForge-Repair-and-Forensics), and the root `CONVERSION-CAPABILITY-MATRIX.md`.

WorldForge does **not** define “world conversion” as one boolean capability. Java ↔ Bedrock, forward upgrades, true downgrades, legacy-console migration, modded save repair, server-player migration, and structure conversion are separate pipelines with separate fidelity verdicts.

## Conversion lanes

| Lane | Preferred current evidence / tool lane | WorldForge rule |
|---|---|---|
| Java → Bedrock | Chunker / je2be / UMT + targeted post-fixers | Track terrain, entities, players, items, maps and metadata separately; never imply Chunker currently converts cross-edition entities/player inventories |
| Bedrock → Java | Chunker / je2be / UMT + targeted post-fixers | Preserve Bedrock actor/XUID provenance and validate Java UUID/player mapping |
| Java older → newer | **real target Minecraft WorldUpgrader / DataFixerUpper** first | Run target-native schema upgrades before semantic/modded repair; inspect chunk-level `DataVersion` |
| Java newer → older | target-aware WorldForge downgrade planner / compatible specialist converter | Never change `DataVersion` alone; explicitly map/drop/quarantine unsupported target features |
| Bedrock older → newer | target Bedrock/world adapter + storage-aware translator | Detect storage/chunk era, not just `level.dat` version fields |
| Bedrock newer → older | target-aware downgrade planner | Never change `StorageVersion` alone; regenerate target-valid records and derived data |
| Legacy Console → Java/Bedrock | LegacyEditor + je2be + UMT evidence | Preserve platform/save-container and title-update/era limits; conversion support is not uniform |
| Modded save migration | target runtime upgrade → mc-world-migrator-style semantic repair | Source save remains truth; match block entities by position, entities by UUID and preserve unknown registry data |
| Server → singleplayer | selected Java UUID / Bedrock XUID identity migration | Preserve the selected player's inventory/state; do not arbitrarily promote the first player record |
| Structures/schematics | common IR with Bloxelizer/Litematica-viewer/Structure2Schematic references | Structure conversion is **not** full-world conversion; report unmapped blocks/entities separately |

## Current tool truth table

| Tool / project | Main strength | Important current boundary |
|---|---|---|
| [Chunker](https://www.chunker.app/) | best free/open broad Java ↔ Bedrock baseline; current Java/Bedrock version coverage | official docs currently exclude cross-edition entity and player-inventory conversion |
| [Universal Minecraft Tool](https://www.universalminecrafttool.com/) | broad Java/Bedrock/Legacy Console commercial conversion and target-version controls | vendor capability is a benchmark, not proof; keep fixture verification for invisible metadata |
| [je2be](https://je2be.app/) | independent open-source Java ↔ Bedrock plus Xbox 360/PS3 routes | use round-trip fixtures; not every edition-specific behavior can be represented exactly |
| [Amulet](https://www.amuletmc.com/) | multi-version Java/Bedrock editing and translation architecture | current entity support remains incomplete; item translation is not universal cross-platform fidelity |
| [Mojang DataFixerUpper](https://github.com/Mojang/DataFixerUpper) | canonical Java schema evolution | forward-data migration primitive, **not** a downgrade engine |
| [PaperMC DataConverter](https://github.com/PaperMC/DataConverter) | fast Java data-converter architecture | standalone generic modded conversion can miss other mods' datafixers; do not use blindly |
| [mc-world-migrator](https://github.com/hkniberg/mc-world-migrator) | source-vs-target semantic modded-save repair | specialist post-upgrade repair, not a general cross-edition converter |
| [LegacyEditor](https://github.com/zugebot/LegacyEditor) | current Legacy Console editing/conversion | R/W support varies by Wii U/Vita/PS3/RPCS3/Switch/PS4/Xenia/Xbox 360 and game era |
| [bedrock-world](https://github.com/BE-Community-Dev/bedrock-world) | deep Bedrock semantic/storage façade | library/reference; WorldForge still owns target-version policy and verification |
| [bedrock-leveldb](https://github.com/BE-Community-Dev/bedrock-leveldb) | Bedrock-specific LevelDB storage correctness | storage primitive only; does not understand gameplay semantics by itself |
| [NBT Studio](https://github.com/tryashtar/nbt-studio) | Java region/NBT and little-endian Bedrock NBT forensics | not a full Bedrock LevelDB converter/editor |
| [Server-to-Singleplayer World Converter](https://github.com/imSirr/world-converter) | selected Java UUID / Bedrock XUID → local-player preservation | narrow identity migration specialist, not a general world converter |
| [Bloxelizer Converter](https://bloxelizer.com/converter) | broad structure/build format bridge | structure/build conversion only |
| [Litematica Viewer](https://github.com/albertchen857/Litematica-viewer) | clean common-IR `.litematic` / Java NBT / `.mcstructure` architecture | block mapping still needs target/version policy |
| [JE2BE Resource Pack Converter](https://github.com/Seraphic-Studio/JE2BE-Resource-Pack-Converter) | Java RP → Bedrock RTX/PBR translation | resource packs only, not world conversion |

## Canonical World IR

WorldForge should avoid N×N pairwise converters. The pipeline is:

`source adapter → source version/edition semantic IDs → canonical World IR → capability/fallback policy → target adapter → derived-data repair → target-runtime verifier`

Every IR node carries provenance, original identifier/value, source edition/version and conversion confidence. Unsupported data is preserved in a sidecar/quarantine record instead of silently disappearing.

## Fidelity dimensions

Every conversion report scores these independently:

- terrain, sections, block states and height bounds;
- biomes;
- block entities;
- entities;
- players, inventory and ender chest;
- items/components/enchantments/books;
- maps;
- dimensions and portals;
- villages, trades, reputation and POIs;
- structures/references/jigsaws;
- raids and scheduled ticks;
- heightmaps and lighting;
- gamerules, difficulty and spawn;
- scoreboard/teams;
- advancements/achievements;
- command blocks/functions/datapacks;
- redstone, fluids and waterlogging semantics;
- Java UUID ↔ Bedrock actor ID/XUID identity;
- resource/behavior-pack dependencies;
- modded/custom registry identifiers.

**Terrain success never upgrades an entity/player/map result to success.**

## Java forward-upgrade workflow

1. keep the original immutable;
2. clone to a versioned workspace;
3. inventory `DataVersion` per region/chunk and required mods/datapacks;
4. launch the real target Minecraft upgrader/server with `--forceUpgrade`;
5. use `--eraseCache` only when the specific migration requires regenerated caches;
6. apply mod-specific/source-vs-target semantic repair after canonical target conversion;
7. rebuild lighting/heightmaps/POIs/structure references only through target-aware logic;
8. scan every dimension/region;
9. launch the actual target runtime and exercise representative chunks/content.

## True downgrade workflow

> [!CAUTION]
> **Changing `DataVersion` or Bedrock `StorageVersion` is not a downgrade.** It can make an incompatible world look superficially older while leaving newer schemas/content behind.

A downgrade must:

1. decode the source with its real source schema;
2. build a target capability matrix;
3. enumerate unsupported blocks/items/entities/components/biomes/dimensions/commands/worldgen;
4. choose explicit preserve/remap/replace/quarantine/drop policies;
5. handle target height bounds and negative-Y/depth loss explicitly;
6. translate into fresh target-native records;
7. rebuild target properties, lighting, heightmaps, POIs and structure references;
8. write to a **new destination world**;
9. parse it with the target adapter;
10. launch the actual target Minecraft version.

## Bedrock format eras

WorldForge must recognize at least:

- current Bedrock LevelDB worlds;
- older LevelDB worlds including `StorageVersion <= 4` terrain semantics;
- pre-LevelDB Pocket-era `chunks.dat` worlds where detected;
- `.mcworld` / `.mctemplate` packaging around those stores.

The [bedrock-world](https://github.com/BE-Community-Dev/bedrock-world) and [bedrock-leveldb](https://github.com/BE-Community-Dev/bedrock-leveldb) projects are strong low-level references for these adapters.

## Legacy Console

Do not flatten Legacy Console into a single format. The adapter records platform, title-update era, container/encryption/signing requirements, chunk format and read/write support. [LegacyEditor](https://github.com/zugebot/LegacyEditor), [je2be](https://je2be.app/) and UMT provide complementary evidence.

## Structures and builds

Common structure IR should cover:

- palettes and namespaced block states;
- block entities;
- entities when the format supports them;
- origin/offset/pivot;
- dimensions and bounds;
- transforms;
- target-version block remapping;
- unmapped/unsupported report.

Target formats include `.schematic`, `.schem`, `.litematic`, Java structure `.nbt`, `.mcstructure` and WorldForge build selections.

## Mandatory torture fixtures

- Java 1.12.2 → 1.20.1 → current modern Java with mixed old/new chunks;
- modern Java → 1.20.1 true downgrade with negative-Y and modern item-component cases;
- current Bedrock → older target version;
- Java ↔ Bedrock world containing villagers/trades, named/tamed mobs, containers, item frames, maps, portals, redstone, command blocks, POIs and structure references;
- Xbox 360 / PS3 / Wii U → Java and Bedrock where supported;
- Java and Bedrock server → singleplayer selected-player migration;
- `.litematic ↔ .schem ↔ Java structure NBT ↔ .mcstructure` round trips;
- modded source → target runtime upgrade → semantic repair.

Final acceptance requires parser checks **and a real target-runtime load**. See the durable research ledger at [`research/minecraft-world-conversion-landscape-2026.md`](https://github.com/Herbertofury/Minecraft-Mod-Vault/blob/main/research/minecraft-world-conversion-landscape-2026.md).