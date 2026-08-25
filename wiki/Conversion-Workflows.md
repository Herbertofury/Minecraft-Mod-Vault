# 🔄 Conversion Workflows

> **Status:** 📋 OmniBridge roadmap unless otherwise marked.

General mod/add-on conversion is documented in **[Mod Conversion Stack](Mod-Conversion-Stack)**. Saved-world/version conversion remains a separate WorldForge pipeline in **[WorldForge — World & Version Conversion Matrix](WorldForge-Conversion-Matrix)**.

## Target workflows

| Workflow | Outcome |
|---|---|
| CIT → Java furniture mod | real blocks/entities, placement, recipes, storage/sitting/animation where inferable |
| CIT → Bedrock furniture | BP/RP + behavior needed for real interaction |
| Resource-pack mob replacement → standalone mob | original vanilla mob remains untouched; converted entity gets independent ID/spawn/loot/behavior |
| Bedrock entity/add-on → Java mod | geometry/animation/controller/Molang plus components/events/Script API reconstructed in target-native Java systems |
| Java mod/entity → Bedrock add-on | BP/RP + Script API/components/controllers where required; no bytecode-to-JSON pretending |
| `.mcaddon` → Java project | dependency-aware Java project, not a JSON dump |
| old Java mod → newer Minecraft | mapping/API/toolchain migration + Mixin/access/registry/network/render/data semantic tests |
| newer Java mod → older Minecraft | target-capability backport with explicit replacements/blocked features; never version-string spoofing |
| Forge ↔ NeoForge | lifecycle/API/data/network/render migration with target-native runtime tests |
| Forge/NeoForge ↔ Fabric | target-native semantic port; common abstractions or compatibility layers used only when behavior remains correct |
| Fabric Yarn/Intermediary era → Mojang / 26.1+ | Loom/Ravel mapping migration + Mixin/access target review across the unobfuscated boundary |
| source-less JAR → target Java mod | deobfuscate/remap + dual decompile + build reconstruction + semantic target port |
| AT ↔ AW/ClassTweaker | canonical Access IR → target-aware emitted access file |
| Mixin → target version/loader | owner/name/descriptor/injection-point/control-flow-aware retargeting; runtime application verification |
| Java custom blocks/items/RP → Bedrock/Geyser | Geyser semantic mappings + Rainbow/pack-generation lane where representable |
| Blockbench/Bedrock animation → Java animated entity | rig/material/animation-aware conversion, GeckoLib/native renderer where appropriate |
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

## Mod conversion architecture

Avoid N×N regex translators:

`source/JAR → archaeology + namespaces → Semantic Port IR → target capability plan → deterministic transforms → target-native implementation → package → differential verifier`

The Semantic Port IR retains mappings, registry identities, lifecycle intent, Mixin fingerprints, access intent, dependency contracts, packet schemas, persistence/components/capabilities, rendering/animation contracts and Java↔Bedrock Behavior IR.

## World conversion architecture

Saved worlds use their separate canonical IR:

`source adapter → version/edition semantic IDs → World IR → capability/fallback policy → target adapter → repair/rebuild → verifier`

Unknown or unsupported source data is retained in provenance/quarantine sidecars instead of silently discarded.

## Workflow rule

A successful workflow ends with:

1. usable target artifact;
2. fidelity report broken down by conversion layer;
3. provenance;
4. unresolved/unsupported items surfaced;
5. validation;
6. real runtime verification where behavior matters.

For mods, **compile/startup success alone is never behavioral parity**. For worlds, **terrain success alone is never entity/player/inventory/map/metadata parity**.
