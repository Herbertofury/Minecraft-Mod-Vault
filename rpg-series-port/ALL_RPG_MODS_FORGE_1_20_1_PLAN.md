# Full RPG Ecosystem -> Forge 1.20.1

## Finish line

Build, verify, and release one coherent **Forge 1.20.1 / Java 17** RPG ecosystem containing every mandatory project supplied by the user plus worthwhile verified RPG-Series-adjacent companions. A project is not considered complete merely because it compiles: applicable dedicated-server, native-client/integrated-world, restart/persistence, content-parity, dependency, compatibility, and deterministic-release gates must pass.

The user-supplied versions are mandatory coverage/reference points, not a ceiling. For each project, use the newest/current upstream as **feature authority** where legally/source-wise available, and use historical/native 1.20.1 releases as mapping/API/behavior substrates when useful. Do not silently downgrade modern content to obtain compatibility.

## Promotion boundary

- Certified integration branch: `workbench/rpg-series-forge-1.20.1`
- Validation PRs against `main` are CI triggers/history only; never merge them into `main` as the RPG promotion mechanism.
- Promote each graduated module with a guarded, non-force fast-forward only after frozen replay evidence.
- Preserve failure, first-green, and frozen-green evidence to GitHub Actions and the canonical Google Drive RPG folder.

## Already graduated foundations / modules

These remain sealed unless a downstream requirement proves a later upstream feature-authority refresh is required:

- Wizards (RPG Series) 3.1.1
- Archers (RPG Series) 3.1.1
- Paladins & Priests (RPG Series) 3.1.1
- Rogues & Warriors (RPG Series) 3.1.1
- Jewelry (RPG Series) 2.4.0
- Ranged Weapon API 2.3.4
- Runes 1.3.2
- Spell Power Attributes 1.6.0
- Spell Engine 1.10.2 core
- Structure Pool API 1.2.1
- TinyConfig 3.1.0
- supporting graduated APIs already in workbench, including Bundle API, Shield API and armor/model foundations.

Current upstream versions must still be checked before final ecosystem release. Example: Spell Engine has advanced beyond the original 1.10.2 input, so a later-authority delta audit is mandatory before final lock.

## Active graduation

- **Gazebos (RPG Series) 2.2.0** — active.
  - Preserve all 17 modern structures.
  - Preserve 12 Repurposed Structures spawn-count integrations.
  - Preserve 12 nested Repurposed Structures pool-addition integrations.
  - Preserve 5 Lithostitched village modifiers.
  - Prove Lithostitched absent -> five direct vanilla pool injections.
  - Prove real Forge discovery of `lithostitched` -> zero direct injections.
  - Native Forge client + fresh packaged dedicated server + deterministic build/replay required.

## Mandatory ecosystem inventory

Every row below must end in one of two valid terminal states: **GRADUATED FORGE 1.20.1** or **REUSED NATIVE FORGE 1.20.1 + INTEGRATION CERTIFIED**. `REFERENCE` means the supplied newer artifact informs feature parity; it does not remove the project from scope.

| Project / supplied artifact | Current migration status |
|---|---|
| Additional RPG Jewelry 2.2.2+1.21.1 | QUEUED - source/dependency audit |
| Archers Expansion 2.1.0+1.21.1 | QUEUED - source/dependency audit |
| Archers 3.1.1+1.21.1 | GRADUATED core; final latest-authority audit remains |
| Armory 1.5.1+1.21.1 | QUEUED |
| Arsenal 1.5.0+1.21.1 | QUEUED |
| Artificers 1.19.0+1.21.1 | QUEUED - source/dependency audit |
| AzureLibArmor 3.1.3 / 1.21.1 | QUEUED foundation - determine exact downstream ABI need and 1.20.1 substrate |
| Bards RPG 1.1.0 / 1.21.1 | QUEUED |
| Berserker RPG 3.1.0+1.21.1 | QUEUED |
| Better Combat Forge 1.9.0+1.20.1 | NATIVE FORGE 1.20.1 substrate; integration certification required |
| Better Combat NeoForge 3.0.0+1.21.10 | REFERENCE / later feature authority input |
| Better Combat NeoForge 3.2.2+26.2 | REFERENCE / later feature authority input |
| Brimstone Battlemages 1.4.0 | QUEUED |
| Combat Roll NeoForge 3.0.1+26.2 | REFERENCE; find latest 1.20.1-compatible substrate then port required modern behavior |
| Critical Strike 1.0.4+1.21.1 | QUEUED |
| Death Knights Fabric 1.0.0-1.20.1 | QUEUED loader conversion to Forge 1.20.1 |
| Dungeon Difficulty 3.7.0+1.21.1 | QUEUED |
| Elemental Wizards RPG 3.1.0+1.21.1 | QUEUED |
| Extra Spell Attributes 1.8.0+1.21.1 | QUEUED foundation/add-on |
| Forcemaster RPG 3.0.0+1.21.1 | QUEUED |
| Gazebos 2.2.0+1.21.1 | ACTIVE |
| Jewelry 2.4.0+1.21.1 | GRADUATED core; final latest-authority audit remains |
| LNE Archers 1.1.3+1.21.1 | QUEUED |
| LNE Paladins 1.1.2+1.21.1 | QUEUED |
| LNE Rogues 1.1.2+1.21.1 | QUEUED |
| LNE Wizards 1.1.4+1.21.1 | QUEUED |
| Loot N Explore 1.0.21-1.21.1 | QUEUED ecosystem/worldgen/content |
| More Relics 1.2.1+1.21.1 | QUEUED |
| More RPG Library 2.7.1+1.21.1 | QUEUED foundation; dependency graph priority |
| MRPGC Skill Tree 1.1.2+1.21.1 | QUEUED |
| Oathsworn Paladins 1.7.0+1.21.1 | QUEUED |
| Paladins 3.1.1+1.21.1 | GRADUATED core; final latest-authority audit remains |
| Ranged Weapon API 2.3.4+1.21.1 | GRADUATED |
| Relics 1.4.0+1.21.1 | QUEUED |
| Rogues 3.1.1+1.21.1 | GRADUATED core; final latest-authority audit remains |
| RPG SERIES DOWNGRADE.zip | REQUIRED provenance/reference input; inspect and reconcile, never overwrite silently |
| RPG Series Icons 1.4.0+1.21.1.zip | REQUIRED resource-pack backport/asset-format validation |
| RPG Class Selection 1.4.0 | QUEUED |
| RPG Minibosses 2.1.0+1.21.1 | QUEUED |
| Runes 1.3.2+1.21.1 | GRADUATED |
| Skill Tree 1.6.0+1.21.1 | QUEUED; requires class quartet + Pufferfish Skills dependency audit |
| Soulmaster 1.0.3 | QUEUED |
| Soulwarden 1.0.4+1.21.1 | QUEUED |
| Spell Engine 1.10.2+1.21.1 | GRADUATED 1.10.2 core; **latest-authority refresh audit required** |
| Spell Power 1.6.0+1.21.1 | GRADUATED |
| Spellblades 0.1.0+1.21.1 | QUEUED |
| Structure Pool API 1.2.1+1.21.1 | GRADUATED |
| Tavern Brawl 0.1.3+1.21.1 | QUEUED |
| TCOTS Witcher 1.1.0+1.21 | QUEUED ecosystem integration |
| Trial Guardians 2.0.0+1.21.1 | QUEUED |
| Village Taverns 1.2.0+1.21.1 | QUEUED |
| Witcher Class Mod 1.2.5-1.20.1 | NATIVE/TARGET-ERA reference; Forge integration/upgrade audit required |
| Wizard Mobs 1.1.1 | QUEUED |
| Wizards 3.1.1+1.21.1 | GRADUATED core; final latest-authority audit remains |

## Verified candidate additions missing from the supplied list

These are not allowed to distract from mandatory scope, but they are explicitly in the discovery queue because they are current, relevant RPG-Series/RPG-Tweaks companions:

- **Projectile Damage Attribute** — ZsoltMolnarrr; ranged damage library, with Forge lineage.
- **Durability Tweaks (RPG Tweaks)** — equipment durability tuning.
- **Enchant Limiter (RPG Tweaks)** — enchant-count limiting mechanic.
- **Hunger Tweaks (RPG Tweaks)** — RPG-oriented exhaustion/regeneration tuning.
- **Quest Items** — quest-objective item pack by the same creator.
- **Attribute Icons (RPG Series)** — resource-pack/UI companion; keep separate from executable mods.
- **Spell Engine Extension (RPG Series Tweaks)** — active extension by TheRedBrain; has a Fabric 1.20.1 line and much newer 1.21.1 feature authority, so it is a strong loader-conversion/update candidate.

Other candidates may be added only after source/license/dependency/behavior viability is verified.

## Dependency-ordered execution strategy

1. **Foundations first**: More RPG Library, AzureLibArmor and any shared attribute/API/provider dependencies proven to gate several downstream modules.
2. **Current Zsolt RPG core/content**: Gazebos -> Relics -> Arsenal -> Armory -> Village Taverns -> Skill Tree -> Dungeon Difficulty / Critical Strike and other creator-adjacent mechanics, ordered by resolved dependency graph rather than arbitrary list order.
3. **Class expansions**: Archers Expansion, Additional RPG Jewelry, LNE quartet, Forcemaster, Berserker, Elemental Wizards, Bards, Artificers, Death Knights, Oathsworn, Brimstone, Soulwarden/Soulmaster, Spellblades, etc.
4. **World/content integrations**: Loot N Explore, Tavern Brawl, RPG Minibosses, Trial Guardians, Wizard Mobs, Witcher/TCOTS ecosystem and related loot/worldgen compatibility.
5. **UI/resource artifacts**: RPG Series Icons / Attribute Icons and other packs, validated for 1.20.1 resource/model formats.
6. **Latest-authority refresh pass** over previously graduated core foundations when upstream moved during the project (especially Spell Engine), but only where the delta is needed or beneficial and can be independently re-certified.
7. **Full-stack certification** with the complete intended mod set installed together: native client/integrated world, dedicated server, config generation, worldgen, loot, class/spell behavior, restart/persistence, dependency ownership, duplicate-provider protection, and performance/regression smoke.

## Per-module acceptance floor

- exact source/version lineage recorded;
- modern mod-owned source/assets/data inventoried;
- target-native Forge 1.20.1 build with Java 17;
- no accidental Fabric/NeoForge-only symbols or metadata in shipped bytes;
- deterministic clean rebuild and frozen identity replay;
- target-version datapack/resource registry naming audited;
- dedicated server required for common/server/worldgen/loot/network behavior;
- native client/integrated-world required for rendering/UI/input/gameplay presentation;
- config/state restart proof where applicable;
- optional-provider present/absent lanes when behavior changes by mod discovery;
- release JAR/source/evidence/checksums preserved to Drive and GitHub;
- guarded non-force promotion into the canonical workbench.

## Final deliverable

A clean Forge 1.20.1 RPG ecosystem release set with no duplicate superseded JARs, a dependency/install manifest, checksums, source/provenance ledger, compatibility notes, resource packs, runtime evidence, and a tested full-stack profile. Every mandatory project above must be accounted for explicitly; nothing gets silently dropped because it is difficult or originally targeted Fabric/NeoForge/newer Minecraft.
