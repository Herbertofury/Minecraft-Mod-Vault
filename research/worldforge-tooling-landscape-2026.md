# WorldForge tooling landscape — 2026-08-25

This is the durable research ledger behind the additive WorldForge expansion. It is a reference/inspiration record, not a dependency list and not permission to reuse code without checking each project's license.

## Strongest new gems

| Project | Role | Current evidence | Why it matters to WorldForge | Primary source |
|---|---|---|---|---|
| Chunk Editor - MCA-Selector | in-game chunk inspection/pruning | 1.0.1, Java 26.2, Fabric/Quilt, 2026-08-23 | Region-file terrain rendering, three LOD layers, custom dimensions, criteria-driven selection, and coordinated region/entities/POI deletion | https://modrinth.com/mod/mca-selector |
| Misanthropy's World Corruption Fixer | recovery after removed worldgen/dimension mods | v0.8, Forge 1.20.1, 2026-07-05 | Proves a narrow guided recovery workflow can rescue saves blocked by stale dimension/worldgen metadata; backup-first and not a generic NBT-corruption fixer | https://modrinth.com/mod/world-corruption-fixer |
| Datapack Load Error Fix | removed-mod/datapack reference repair | broad 1.18.2–1.21.x Forge/NeoForge coverage | Detects and cleans invalid custom dimension/entity/block references so damaged worlds can load again | https://modrinth.com/mod/datapack-load-error-fix |
| mc-world-migrator | semantic modded-world migration | active niche migration tool | Excellent source-vs-target design: original save as truth, block entities by position, entities by UUID, playerdata/.dat repair, NBT-to-components mapping, dry-run, resumable manifest, idempotent fixers | https://github.com/hkniberg/mc-world-migrator |
| Neruina - Ticking Entity Fixer | crash containment | Forge 1.20.1 route; 3.2.2 surfaced | Treats ticking entity/block-entity failures as quarantinable faults instead of letting one bad object brick the whole world | https://modrinth.com/mod/neruina |
| YapiFix | worldgen compatibility | Forge 1.20.1; active 2026 | Deterministically resolves FeatureSorter/feature-order cycles after biome modifiers, a valuable model for diagnosing worldgen graph conflicts | https://modrinth.com/mod/yapifix |
| WorldVersionBackport | version-layout migration research | Fabric; 26.2 transition research | Useful case study for folder/data-format changes and downgrade/backport workflows; explicitly backup-first and not server-ready | https://modrinth.com/mod/worldversionbackport |
| Chunky | chunk pregeneration | active broad server/loader support incl. current versions | Mature long-running pregeneration baseline with shapes, world borders and multiple worlds | https://modrinth.com/plugin/chunky |
| Chunky Extension | load-aware pregeneration | Fabric/Forge/NeoForge/Quilt modern lines | Pause while players are online, resume when empty, scheduler/GUI/commands: direct inspiration for safe background-style long operations | https://modrinth.com/mod/chunky-extension |
| Regionerator | server world maintenance | active Bukkit ecosystem reference | Gradual deletion of old/unused areas is a useful continuous-maintenance model; direct-world edits require backups | https://github.com/Jikoo/Regionerator |
| ChunkCleaner | offline chunk pruning | Go utility | Dry-run plus move-to-quarantine instead of immediate deletion; also highlights that InhabitedTime is evidence, not proof of player builds | https://github.com/zeroBzeroT/ChunkCleaner |
| uNmINeD | world inspection/rendering | 0.19.58 dev line surfaced 2026-03 | Strong modded/Bedrock-aware 2D rendering reference, including texture-derived colors for custom/modded blocks | https://unmined.net/ |
| BlueMap | 3D map/inspection pipeline | 5.23 surfaced 2026-08 | Reads world files into a browser-viewable 3D representation and can run standalone or server-side; useful streamed-inspection benchmark | https://github.com/BlueMap-Minecraft/BlueMap |
| Seed Atlas | seed/world analysis | 1.0 release 2026-07-14; active 2026 fork lineage | Fast local biome/structure search, density analysis, multithreaded large-area searches, Lua filters, resumable sessions | https://github.com/DUzzL/Seed-Atlas |
| WorldBinder | authorized world capture/preservation | active 26.2 Fabric line | Recovery cache, capture queue, loaded-chunk radar, export targeting and crash/disconnect recovery are valuable preservation/streaming references; only use on worlds the operator is authorized to preserve | https://modrinth.com/mod/worldbinder |

## Existing baselines refreshed

| Project | 2026 note | Source |
|---|---|---|
| Amulet Map Editor | 0.10.60 surfaced 2026-07-27; recent work includes CurseForge/Modrinth world paths and a modded block-entity corruption fix; 0.10.57 added Java 26.2 / Bedrock 26.30 support | https://github.com/Amulet-Team/Amulet-Map-Editor |
| MCA Selector | 2.8 surfaced in current install docs; filters include DataVersion, LastUpdate and InhabitedTime; NBT Changer supports targeted chunk metadata changes | https://github.com/Querz/mcaselector |
| Cubiomes Viewer | remains the important seed-finding/map-viewer lineage behind the newer Seed Atlas direction | https://github.com/Cubitect/cubiomes-viewer |

## WorldForge feature debt exposed by the sweep

The research adds several concrete requirements beyond the prior broad 'repair & forensics' line:

1. **Read-only diagnosis first** — classify damage before any write: invalid registry references, missing dimensions, ticking objects, malformed NBT/region data, schema/version mismatch, metadata drift and generation-graph conflicts.
2. **Source-vs-target semantic diff** — when an original/pre-upgrade save exists, use it as authoritative evidence rather than guessing lost modded fields.
3. **Repair plan + dry run** — every destructive or schema-changing operation should show the exact proposed changes, counts, affected dimensions/chunks/entities/files and confidence before commit.
4. **Transactional quarantine** — prefer move/copy/quarantine of suspect chunks/entities/data over irreversible deletion; keep rollback artifacts and journals.
5. **Resumable/idempotent operations** — region-scale conversion, pregeneration and repair should survive interruption and safely resume without replaying completed work.
6. **Long-operation scheduler** — pause/throttle around active players, CPU/IO pressure and configured maintenance windows.
7. **Layer-coherent chunk operations** — Java chunk terrain, entity and POI files must be treated as one logical unit when pruning/recovering; Bedrock requires equivalent LevelDB key-family integrity.
8. **Heuristic honesty** — InhabitedTime and block-pattern heuristics are evidence, not proof of player construction; manual protection always wins.
9. **Inspection before mutation** — streamed 2D/3D maps, LOD overlays, seed/structure views, chunk age/version/error heatmaps and diff overlays should be first-class diagnostic surfaces.
10. **Authorized capture/recovery** — client-observed capture workflows can inform disaster recovery, but must never imply access to server state the client was never sent.

## Suggested implementation priority

1. Repair & Forensics contract + read-only scanner.
2. Transaction/dry-run/quarantine journal shared by repair, pruning and conversion.
3. Chunk-map inspection with version/age/error overlays and coherent layer handling.
4. Source-vs-target semantic migration engine for modded/version transitions.
5. Resumable pregeneration/retrogen scheduler with backpressure.
6. Seed/structure and 2D/3D inspection adapters.
7. Optional authorized capture/import workflows.

Last verified: 2026-08-25.