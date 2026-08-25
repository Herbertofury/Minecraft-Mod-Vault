# 🩺 WorldForge — Repair & Forensics

> **Status:** 📋 WorldForge roadmap. This is an additive safety/repair contract, not a claim that automatic recovery is always possible.

WorldForge should treat a damaged Minecraft world like evidence: **inspect first, explain the failure, preserve the original, propose the smallest repair, then verify the repaired copy in real Minecraft.**

## Recovery philosophy

1. **Open read-only first.** Never mutate the only copy while diagnosis is still uncertain.
2. **Snapshot before repair.** Record hashes, format/version metadata, dimensions, region inventory and important world files.
3. **Classify the fault.** Distinguish registry/reference failures, malformed NBT, damaged region records, ticking entities, removed dimensions, worldgen graph failures, partial upgrades and semantic migration loss.
4. **Show a repair plan.** List exactly what would change and why, with confidence and risk.
5. **Dry run.** Parse and validate the proposed target without committing the destructive step.
6. **Quarantine instead of deleting.** Preserve suspect chunks/entities/files in a recoverable sidecar whenever practical.
7. **Commit transactionally.** Journal each operation so a crash or interruption can resume or roll back.
8. **Validate structurally and semantically.** A world that opens but silently loses inventories, POIs, maps or modded state is not a successful repair.
9. **Launch the repaired copy.** TestGrid/Agent Driver should exercise the actual target Minecraft build.

## Failure classes WorldForge should recognize

| Failure family | Examples | Preferred first action |
|---|---|---|
| **Registry / removed-content references** | removed custom dimensions, stale biome/dimension keys, invalid datapack references | identify missing IDs and affected data; compare against a source/backup if available |
| **Ticking object crashes** | one entity or block entity throws every tick | quarantine/suspend the exact object and preserve its NBT for inspection |
| **Region / chunk corruption** | malformed region entry, unreadable chunk payload, mismatched entity/POI layer | isolate the smallest affected record; avoid deleting the whole region when one chunk is bad |
| **NBT / saved-data damage** | malformed playerdata, level metadata, map/scoreboard data | schema-aware parse, raw/hex fallback, source-vs-target comparison |
| **Worldgen ordering conflicts** | feature-order cycles after biome modifiers | graph the feature constraints and explain the cycle before applying a deterministic resolution |
| **Partial version migration** | mixed DataVersion, height transition artifacts, folder/schema changes | identify source/target semantics and migrate only the incompatible layers |
| **Modded semantic loss** | block entity kept but inventory/config NBT vanished after a mod upgrade | compare original and target by stable identity/position and restore only fields proven missing |
| **Interrupted long operation** | half-finished conversion/pruning/pregen | resume from an operation manifest; never blindly restart completed work |

## Read-only World Doctor

The first repair surface should be a **World Doctor** that can scan without modifying anything.

Suggested reports:

- world edition, storage format and detected version history;
- dimensions and dimension folders/databases;
- region/chunk inventory and unreadable records;
- per-chunk `DataVersion`, `LastUpdate`, `InhabitedTime` and generation status where available;
- entities, block entities and POIs with parse/reference errors;
- missing registry IDs and custom-content references;
- data packs / behavior packs / resource packs referenced by the save;
- playerdata, inventories, ender chests, advancements/statistics and spawn points;
- maps, scoreboards, raids, villages, portals and structure references;
- duplicate/missing UUIDs or identity inconsistencies;
- mixed-version and partially migrated areas;
- suspiciously orphaned data after mod/add-on removal;
- confidence-ranked repair candidates.

Every issue should be clickable from the report into a 2D/3D location or raw semantic/NBT view when a world coordinate exists.

## Source-vs-target semantic recovery

A particularly strong 2026 research pattern comes from **[mc-world-migrator](https://github.com/hkniberg/mc-world-migrator)**: when both the original/pre-upgrade world and the broken target exist, compare them semantically rather than guessing.

WorldForge should support matching by stable identity:

- chunks by dimension + coordinates;
- block entities by dimension + block position + expected type;
- entities by UUID, with position/type as supporting evidence;
- players by UUID and resolved identity;
- maps/POIs/structures by their native identifiers and references.

The recovery engine can then distinguish **expected schema evolution** from **actual dropped data**, producing field-level diffs before restoration. Repairs must be idempotent so rerunning the same plan does not duplicate inventories, entities or metadata.

## Quarantine model

Borrow the safe idea seen in **[Neruina](https://modrinth.com/mod/neruina)** and **[ChunkCleaner](https://github.com/zeroBzeroT/ChunkCleaner)**: containment is usually safer than immediate destruction.

Quarantine should retain:

- original raw bytes/NBT;
- source file/key and coordinates/UUID;
- detected error;
- repair operation that moved it;
- timestamp and tool/version;
- restore command/action.

Examples:

- move a crashing entity out of the active world while keeping its complete NBT;
- move candidate stale chunks to a quarantine set rather than deleting them;
- retain the original `level.dat` / playerdata before schema repair;
- keep removed-dimension metadata separately when stripping stale references.

## Specialist references

| Project | What it teaches WorldForge | Link |
|---|---|---|
| Misanthropy's World Corruption Fixer | narrow, guided Forge 1.20.1 recovery after removed dimension/worldgen content; explicit backup-first scope | https://modrinth.com/mod/world-corruption-fixer |
| Datapack Load Error Fix | invalid custom dimension/entity/block-reference cleanup to get otherwise blocked worlds loading | https://modrinth.com/mod/datapack-load-error-fix |
| Neruina | ticking entity/block-entity fault containment and diagnostic workflow | https://modrinth.com/mod/neruina |
| YapiFix | deterministic worldgen feature-order conflict diagnosis/repair | https://modrinth.com/mod/yapifix |
| mc-world-migrator | source-vs-target semantic comparison, manifests, dry-run, resume and idempotent fixers | https://github.com/hkniberg/mc-world-migrator |
| Amulet | multi-version Java/Bedrock data abstraction and low-level world manipulation | https://github.com/Amulet-Team/Amulet-Core |
| MCA Selector | chunk filters, version/age metadata, visual isolation and targeted NBT operations | https://github.com/Querz/mcaselector |

## Repair acceptance tests

A repair is not complete merely because the title screen stops crashing.

For representative fixtures, verify:

- world opens in the intended exact Minecraft/loader/modpack version;
- player location, inventory, ender chest, XP and spawn survive;
- containers retain exact contents and slot ordering;
- block entities retain orientation/mode/settings;
- pets/named entities/villagers retain identity and important state;
- portals, maps, structures and POIs still resolve correctly;
- repaired chunks pass save → close → reopen cycles;
- no fresh ticking crash appears after simulation;
- no duplicate UUID/item/entity was introduced;
- source data remains untouched unless the user explicitly chose in-place repair;
- transaction journal can explain every mutation.

## Hard rules

- Never advertise a generic “fix corruption” button that silently deletes whatever fails to parse.
- Never treat `InhabitedTime` as proof that a chunk contains or does not contain a player build.
- Never erase unknown modded/custom data simply because WorldForge lacks a schema; preserve opaque data when structurally possible.
- Never downgrade or migrate the only copy of a world.
- Never claim data can be reconstructed when the client/tool never possessed that data and no backup/source exists.

Related: [WorldForge](WorldForge) · [Pruning & Retrogen](WorldForge-Pruning-and-Retrogen) · [Validation & Fidelity](Validation-and-Fidelity) · [2026 research ledger](https://github.com/Herbertofury/Minecraft-Mod-Vault/blob/main/research/worldforge-tooling-landscape-2026.md)
