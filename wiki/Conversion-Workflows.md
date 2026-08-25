# 🔄 Conversion Workflows

> **Status:** 📋 OmniBridge roadmap unless otherwise marked.

World, version and content conversion are separate semantic pipelines. For the full world/version matrix, current specialist tools and downgrade rules, see **[WorldForge — World & Version Conversion Matrix](WorldForge-Conversion-Matrix)**.

## Target workflows

| Workflow | Outcome |
|---|---|
| CIT → Java furniture mod | real blocks/entities, placement, recipes, storage/sitting/animation where inferable |
| CIT → Bedrock furniture | BP/RP + behavior needed for real interaction |
| Resource-pack mob replacement → standalone mob | original vanilla mob remains untouched; converted entity gets independent ID/spawn/loot/behavior |
| Bedrock entity → Java entity | geometry, animation/controllers, behavior, spawn, drops, sounds, particles |
| Java entity → Bedrock add-on | BP/RP + scripts where required |
| `.mcaddon` → Java project | dependency-aware Java project, not a JSON dump |
| old Java mod → newer Minecraft | mappings/API/loader/toolchain migration + tests |
| Forge ↔ NeoForge/Fabric | semantic lifecycle/registry/network/render migration, not text substitution |
| Blockbench/Blender → entity/build | rig/material/animation-aware conversion |
| Java world → Bedrock world | WorldForge terrain + semantic data migration with independent terrain/entity/player/map/metadata verdicts |
| Bedrock world → Java world | WorldForge LevelDB → Java world migration with Bedrock actor/XUID provenance |
| old Java world → newer Java | real target Minecraft WorldUpgrader/DataFixerUpper first, then semantic/modded repair |
| newer Java world → older Java | explicit target-aware lossy downgrade; never `DataVersion` spoofing |
| old Bedrock world → newer Bedrock | storage-era-aware target upgrade |
| newer Bedrock world → older Bedrock | explicit target-aware downgrade; never `StorageVersion` spoofing |
| Legacy Console → Java/Bedrock | platform/title-update-aware extraction + translation; support varies by save format |
| modded source world → target version | target runtime migration followed by source-vs-target semantic repair |
| multiplayer/server world → singleplayer | selected Java UUID / Bedrock XUID becomes the intended local player while preserving inventory/state |
| `.litematic` / `.schem` / Java structure NBT / `.mcstructure` | common structure IR, target-version block mapping, unsupported-block/entity report |
| Java resource pack → Bedrock resource pack/RTX | texture/model/material mapping with explicit unsupported-feature report |

## Conversion architecture

Avoid N×N pairwise converters where possible:

`source adapter → version/edition semantic IDs → canonical IR → capability/fallback policy → target adapter → repair/rebuild → verifier`

Unknown or unsupported source data is retained in provenance/quarantine sidecars instead of silently discarded.

## Workflow rule

A successful workflow ends with:

1. usable target artifact;
2. fidelity report broken down by conversion layer;
3. provenance;
4. unresolved/unsupported items surfaced;
5. validation;
6. real runtime verification where behavior matters.

For world conversions, **terrain success alone is never enough** to claim entity, player, inventory, map, POI or structure-metadata success.