# 🔄 Conversion Workflows

> **Status:** 📋 OmniBridge roadmap unless otherwise marked.

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
| Java world ↔ Bedrock world | WorldForge semantic world migration |

## Workflow rule

A successful workflow ends with:

1. usable target artifact;
2. fidelity report;
3. provenance;
4. unresolved items surfaced;
5. validation;
6. real runtime verification where behavior matters.
