# 🧰 Tool Catalogue & Reference Ecosystem

These are **research/integration targets**, not automatic dependencies or permission to reuse code without checking licenses.

> For player-facing Minecraft launchers, clients, and mod managers, see **[Launchers & Mod Managers](Launchers-and-Managers)**.

| Tool / project | Best for | OmniBridge / WorldForge interest | Link |
|---|---|---|---|
| Chunker | Java ↔ Bedrock world conversion | mappings, adapters, packaging, custom identifiers | https://github.com/HiveGamesOSS/Chunker |
| Amulet | multi-version world abstraction/editing | world model, Anvil + LevelDB workflows | https://github.com/Amulet-Team/Amulet-Core |
| Universal Minecraft Tool | converter/pruner/NBT benchmark | parity + known-gap benchmark | https://www.universalminecrafttool.com/ |
| je2be | Java/Bedrock/legacy conversion | conversion architecture/reference | https://github.com/kbinani/je2be-core |
| Axiom | polished in-game editing | UX, selection, sculpting, preview benchmark | https://axiom.moulberry.com/ |
| WorldEdit | mature edit semantics | selections, masks, patterns, schematics | https://worldedit.enginehub.org/ |
| FAWE | huge async edits | bounded resource use, history, large operations | https://github.com/IntellectualSites/FastAsyncWorldEdit |
| MCA Selector | chunk map/pruning | chunk filters, overlays, pruning workflows | https://github.com/Querz/mcaselector |
| WorldPainter | terrain creation | painting, heightmaps, objects, layers | https://www.worldpainter.net/ |
| Minecraft Bedrock Editor | official Bedrock world authoring | living parity benchmark + extension bridge | https://github.com/Mojang/minecraft-editor |
| Chunk Editor - MCA-Selector | in-game region inspection/pruning | LOD world map, criteria selection, custom dimensions, coordinated terrain/entity/POI deletion | https://modrinth.com/mod/mca-selector |
| Misanthropy's World Corruption Fixer | Forge 1.20.1 removed-worldgen/dimension recovery | guided backup-first repair planning and stale-dimension cleanup | https://modrinth.com/mod/world-corruption-fixer |
| Datapack Load Error Fix | invalid removed-content references | detection/cleanup of custom dimension/entity/block references that block world loading | https://modrinth.com/mod/datapack-load-error-fix |
| mc-world-migrator | semantic migration repair | source-vs-target diffs, UUID/position matching, manifests, dry-run, resume, idempotent fixers | https://github.com/hkniberg/mc-world-migrator |
| Neruina | ticking entity/block-entity crash containment | quarantine and diagnose a bad object without sacrificing the entire save | https://modrinth.com/mod/neruina |
| YapiFix | worldgen feature-order conflicts | deterministic feature graph diagnosis/resolution after biome modifiers | https://modrinth.com/mod/yapifix |
| Chunky | chunk pregeneration | mature shape/world-border/multiworld pregen baseline | https://modrinth.com/plugin/chunky |
| Chunky Extension | load-aware pregeneration | persistent pause/resume and player-aware scheduling for long world jobs | https://modrinth.com/mod/chunky-extension |
| Regionerator | gradual server world maintenance | incremental unused-region regeneration instead of giant destructive maintenance windows | https://github.com/Jikoo/Regionerator |
| ChunkCleaner | offline chunk pruning | dry-run + quarantine-before-delete and cautious InhabitedTime use | https://github.com/zeroBzeroT/ChunkCleaner |
| uNmINeD | huge-world 2D rendering | fast inspection, Java/Bedrock/modded visualization and map indexing | https://unmined.net/ |
| BlueMap | browser 3D world maps | streamed 3D inspection and standalone/server map pipeline | https://github.com/BlueMap-Minecraft/BlueMap |
| Seed Atlas | biome/structure/seed analysis | multithreaded local searches, density analysis, scripting and resumable sessions | https://github.com/DUzzL/Seed-Atlas |
| Cubiomes Viewer | seed finding/map analysis | established cubiomes visualization/search lineage | https://github.com/Cubitect/cubiomes-viewer |
| WorldBinder | authorized client-observed preservation | capture queue, recovery cache, loaded-chunk inspection and export workflow research | https://modrinth.com/mod/worldbinder |
| WorldVersionBackport | version-layout/backport research | source/target world-layout migration lessons; backup-first experimental reference | https://modrinth.com/mod/worldversionbackport |
| Blockbench | Minecraft modeling/animation | deep round-trip target | https://www.blockbench.net/ |
| Blender | 3D/rigging/automation | mesh/rig/animation/voxel bridge | https://www.blender.org/ |
| Porting Lib | Forge-like systems on Fabric | compatibility/substitution research | https://github.com/Fabricators-of-Create/Porting-Lib |
| Sinytra ecosystem | cross-loader compatibility | adapter/reference research | https://github.com/Sinytra |
| Azalea | Rust Minecraft client/bot | fast non-rendering TestGrid worker | https://github.com/azalea-rs/azalea |
| Mineflayer | Java protocol/player bot | broad interaction test worker | https://github.com/PrismarineJS/mineflayer |
| Baritone | pathfinding/navigation | real-client agent navigation option | https://github.com/cabaletta/baritone |

## World-tooling research notes

- The dedicated [WorldForge](WorldForge) roadmap now incorporates repair/forensics, resumable pregeneration and inspection lessons from the specialist tools above.
- The [Repair & Forensics](WorldForge-Repair-and-Forensics) contract requires read-only diagnosis, dry runs, source-vs-target comparison, quarantine and real-game verification before calling a recovery successful.
- `InhabitedTime` and similar metadata are useful **signals**, never proof that a chunk is disposable or player-built.
- WorldBinder-style client capture is only an authorized preservation reference; a client cannot recover server data it was never sent.
- Full research ledger: https://github.com/Herbertofury/Minecraft-Mod-Vault/blob/main/research/worldforge-tooling-landscape-2026.md

## Official docs worth bookmarking

- Fabric: https://docs.fabricmc.net/
- NeoForge: https://docs.neoforged.net/
- Minecraft Creator: https://learn.microsoft.com/minecraft/creator/
- Bedrock Editor docs: https://learn.microsoft.com/en-us/minecraft/creator/documents/bedrockeditor/
- Modrinth API: https://docs.modrinth.com/api/
- YouTube Data API: https://developers.google.com/youtube/v3
