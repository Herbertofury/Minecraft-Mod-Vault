# Minecraft world conversion landscape — 2026-08-25

This is the source-backed research ledger for OmniBridge / WorldForge world conversion. It is additive to the existing conversion capability matrix and WorldForge research.

## Bottom line

There is no single converter that proves complete semantic fidelity across Java ↔ Bedrock ↔ Legacy Console ↔ arbitrary older/newer versions. WorldForge therefore needs a **multi-lane conversion architecture** with explicit capability gaps, source preservation, target-aware translation, post-fixers, and real target-runtime verification.

The major conversion classes are distinct and must never be collapsed into one generic “convert world” operation:

1. Java → Bedrock;
2. Bedrock → Java;
3. Java forward version upgrade;
4. Java true downgrade;
5. Bedrock forward version upgrade;
6. Bedrock true downgrade;
7. Legacy Console → Java/Bedrock;
8. modded save migration/repair;
9. server → singleplayer/player-identity migration;
10. structure/schematic conversion;
11. resource-pack/PBR cross-edition conversion.

## Current leaders and specialists

| Project / method | Best role | Current evidence / important limitation | Primary source |
|---|---|---|---|
| Chunker | best free/open broad Java ↔ Bedrock baseline | release 1.19.1 (2026-08-03); broad Java/Bedrock version list; converts terrain/settings/biomes/block entities/containers/maps, but current official docs explicitly do **not** cross-edition-convert entities or player inventories | https://github.com/HiveGamesOSS/Chunker ; https://www.chunker.app/ |
| Universal Minecraft Tool | highest-capability commercial full-world benchmark | current site advertises Java 1.3.2–26.2, Bedrock 1.0.0–26.30, Xbox 360/PS3/Wii U legacy input and target-version output; advanced older-chunk/depth/world-center/property repair. Vendor TODO/limitations and performance claims still require fixture verification | https://www.universalminecrafttool.com/ |
| je2be | strongest independent open-source full-world / legacy comparator | Java ↔ Bedrock plus Xbox 360/PS3 → Java/Bedrock; desktop release 5.6.1 surfaced; excellent round-trip/reference tests | https://github.com/kbinani/je2be-core ; https://github.com/kbinani/je2be-desktop ; https://je2be.app/ |
| Amulet Map Editor / Core / PyMCTranslate | best open multi-version editor/translation architecture reference | Map Editor 0.10.62 released 2026-08-24; supports broad Java/Bedrock editing. Current documented limitations include unsupported entity editing and same-platform item translation, so it is not a complete semantic converter | https://github.com/Amulet-Team/Amulet-Map-Editor ; https://github.com/Amulet-Team/PyMCTranslate |
| Mojang DataFixerUpper + real target WorldUpgrader | canonical Java **forward-upgrade** mechanism | DFU is the official versioned-data transformation library. A safe user workflow runs the real target Minecraft upgrader/server `--forceUpgrade`, with `--eraseCache` only when justified, rather than manually rewriting `DataVersion` | https://github.com/Mojang/DataFixerUpper |
| PaperMC DataConverter | high-performance Java data-converter architecture reference | rewrite of the Minecraft data converter; standalone Fabric use is explicitly unsafe for generic modded-world conversion because other mods’ datafixers may be ignored. Treat as integrated/reference path, not a universal standalone fixer | https://github.com/PaperMC/DataConverter |
| mc-world-migrator | semantic modded-save migration/repair | source-vs-target truth, block entities matched by position, entities by UUID, playerdata/.dat repair, old NBT→components, dry-run, resumable manifest, idempotent fixers | https://github.com/hkniberg/mc-world-migrator |
| LegacyEditor | current 2026 Legacy Console specialist | 1.4.5 BatchConverter (2026-01-15); Wii U/Vita/RPCS3/Switch/PS4 R/W paths, PS3 partial write, Xenia/Xbox 360 read-only paths, era-specific support | https://github.com/zugebot/LegacyEditor |
| bedrock-world | unusually complete Bedrock semantic/storage façade | level.dat, chunks/subchunks, players, entities, biomes, maps, villages/global records, scans, `.mcstructure`; detects current LevelDB, old `StorageVersion <= 4`, and pre-LevelDB Pocket `chunks.dat` | https://github.com/BE-Community-Dev/bedrock-world |
| bedrock-leveldb | Bedrock LevelDB storage adapter | Rust LevelDB implementation aware of Bedrock comparator/Snappy/table details, unknown-key preservation and streaming/snapshots; storage primitive, not a gameplay semantic converter | https://github.com/BE-Community-Dev/bedrock-leveldb |
| PrismarineJS minecraft-data | per-version mapping evidence | Java + Bedrock blocks/items/biomes/entities/recipes/protocol/legacy mappings; excellent Version Atlas evidence but not a stored-world converter | https://github.com/PrismarineJS/minecraft-data |
| PrismarineJS prismarine-chunk | neutral chunk abstraction reference | Java/Bedrock chunk abstraction useful for adapter tests; not complete world/player/entity conversion | https://github.com/PrismarineJS/prismarine-chunk |
| NBT Studio | manual NBT/region forensic companion | Java NBT, `.mca`, `.mcr`, little-endian Bedrock NBT / `.mcstructure`, Save As between NBT encodings; not a full Bedrock LevelDB editor | https://github.com/tryashtar/nbt-studio |
| Server-to-Singleplayer World Converter | player identity preservation specialist | browser/local tool handles Java chosen UUID including 26.1+ world layout and Bedrock `player_server_*`/XUID → `~local_player`; narrow but excellent conversion rule reference | https://github.com/imSirr/world-converter |
| Bloxelizer Converter | strongest broad structure/build format bridge found | local/browser conversion across `.schem`, `.schematic`, `.litematic`, `.nbt`, `.mcstructure`, Java world ZIP and several 3D/model formats with target-version / cross-edition block mapping | https://bloxelizer.com/converter |
| Litematica Viewer | clean common-IR structure architecture reference | reads/writes `.litematic`, Java structure NBT and `.mcstructure` through a common structure object; reports unmapped blocks and supports transforms/analysis | https://github.com/albertchen857/Litematica-viewer |
| Structure2Schematic | Bedrock/Java structure → WorldEdit schematic specialist | handles blocks, block entities, entities and palettes; web/CLI/API; still development | https://github.com/Chaoscaot/Structure2Schematic |
| JE2BE Resource Pack Converter | Java resource pack → Bedrock RTX specialist | Java 1.8+ RP → Bedrock RTX, 900+ texture mappings, LabPBR normal/specular → Bedrock MER and texture_set generation | https://github.com/Seraphic-Studio/JE2BE-Resource-Pack-Converter |
| WorldVersionBackport | experimental downgrade/version-layout research | useful case study for modern layout/backport handling; backup-first, experimental and not a generic safe downgrade or server solution | https://modrinth.com/mod/worldversionbackport |

## Hard architectural rules

### 1. Upgrade and downgrade are different algorithms

**Forward Java upgrade:**

`immutable source → clone → real target Minecraft WorldUpgrader/DFU → mod-specific semantic repair → derived-data repair → target runtime verification`

WorldForge must inspect chunk-level `DataVersion`; `level.dat` alone does not prove a mixed-age world was fully upgraded.

**True downgrade:**

Never lower `DataVersion` or Bedrock `StorageVersion` and call the result converted. A downgrade is a target-aware lossy translation:

`source decode → target capability matrix → unsupported-content policy → semantic remap → target schema encode → rebuild derived data → fresh target world → actual target runtime test`

The downgrade planner must explicitly account for features the target cannot represent: blocks/items/components, dimensions/biomes, entity metadata, negative-Y/deeper terrain, height expansion, structure systems, POIs, commands, item components, registries and pack dependencies.

### 2. Use a canonical intermediate representation

Do not build N×N pairwise translators. Use:

`source adapter → version/edition semantic IDs → canonical World IR → policy/fallback pass → target adapter → repair/rebuild pass → verifier`

IR nodes need provenance and confidence so an unsupported source field is never silently dropped.

### 3. Conversion is layered

Every job tracks independently:

- terrain/chunk sections/block states;
- biomes and height bounds;
- block entities;
- entities;
- players/inventory/ender chest;
- items/components/enchantments/books;
- maps;
- dimensions/portals and Nether coordinate semantics;
- villages, villager trades/reputation and POIs;
- structures/references/jigsaws;
- raids;
- scheduled ticks;
- heightmaps and lighting;
- gamerules/difficulty/spawn;
- scoreboard/teams;
- advancements/achievements;
- command blocks/functions/datapacks;
- redstone/fluid/waterlogging semantics;
- Java UUID ↔ Bedrock actor ID/XUID identity;
- resource/behavior-pack dependencies;
- custom/modded registry identifiers.

A green terrain conversion cannot be reported as a green player/entity conversion.

### 4. Chain tools by capability, not brand

A valid pipeline may use a broad converter for terrain, a semantic fixer for player/entity fields, an NBT forensic tool for exceptional records and WorldForge for post-conversion verification. The source remains immutable and the output is always separate.

### 5. Preserve unknown data

Unknown/custom records must be copied into provenance/quarantine sidecars or preserved in-place when the target storage model permits. Never discard merely because the current adapter does not understand them.

### 6. Real target runtime is the final authority

Static parsing is insufficient. The target Minecraft edition/version must actually load the fresh output, visit representative chunks/dimensions and exercise converted systems.

## Explicit current-tool blind spots

- **Chunker:** current official limitation on cross-edition entities and player inventories.
- **Amulet:** entity support remains incomplete/unsupported in current editor documentation; items translate only within the same platform lane.
- **UMT:** commercial capability is a benchmark, not proof; maintain fixture tests for maps, villagers/trades/reputation and other invisible metadata.
- **Paper DataConverter:** do not use standalone as a generic modded conversion route where mod datafixers are required.
- **LegacyEditor:** read/write capability differs by platform/save container and game era; never flatten this into “all Legacy Console supported.”
- **Structure converters:** a successful `.litematic ↔ .mcstructure` conversion does not imply world/player/entity/POI fidelity.

## Mandatory acceptance corpus

1. Java 1.12.2 → 1.20.1 → current modern Java with both old and post-1.18 chunks.
2. Modern Java → 1.20.1 true downgrade including negative-Y terrain and modern item-component loss/replacement policy.
3. Current Bedrock → older target version with unsupported-block/entity policy.
4. Java ↔ Bedrock semantic torture world containing villagers/trades, named/tamed mobs, containers, item frames, maps, portals, redstone, command blocks, POIs and structure references.
5. Xbox 360 / PS3 / Wii U legacy fixtures → Java and Bedrock where supported.
6. Java and Bedrock server → singleplayer with selected-player identity/inventory preservation.
7. `.litematic ↔ .schem ↔ Java structure NBT ↔ .mcstructure` round trips.
8. Modded source → upgraded runtime → semantic repair fixture with removed/renamed registry content.

Assertions include entity/player counts, UUID/XUID mappings, inventory/container hashes, map pixels/IDs, POI and structure-reference integrity, chunk-version normalization, seam checks, output parser success and real target-runtime launch.

## WorldForge priority

1. canonical World IR + provenance;
2. source/target capability planner;
3. Java DFU/WorldUpgrader adapter;
4. Bedrock storage-version adapter using LevelDB-aware semantics;
5. Chunker/je2be/UMT/Amulet adapters with capability-specific verdicts;
6. legacy-console adapter led by LegacyEditor/je2be evidence;
7. semantic player/entity/map/POI fixers;
8. structure common-IR bridge;
9. differential and real-runtime TestGrid corpus.

Last verified: 2026-08-25.