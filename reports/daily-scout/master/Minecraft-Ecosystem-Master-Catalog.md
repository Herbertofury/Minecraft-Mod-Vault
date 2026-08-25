# Minecraft Ecosystem Master Discovery Catalog

**Last audited:** 2026-08-25  
**Permanent categories:** 34  
**Tracked serious projects/finds in canonical Google Sheet:** 124  
**Female-mob companion vault:** 99 curated entries / 50 explicit 1.20.1 routes / 19 maid-ecosystem entries / 12 source hubs / 2 Bedrock routes

This GitHub mirror is the durable change/ranking index for the permanent Minecraft ecosystem scouting service. The canonical sortable database and complete board live in the paired Google Sheet; the curated Google Docs are the readable long-form layers. Historical entries are retained here, in dated checkpoints, in the canonical Sheets, and in Git history rather than silently deleted.

## Latest material delta — 2026-08-25 fourth late-cycle sweep

- **#26 / #27 / #29 client-safe translation runner:** [BabelCraft 1.2.5.3](https://www.curseforge.com/minecraft/mc-mods/babelcraft) is the stable Forge 1.20.1–1.20.6 route. The Aug 24 **2.0.0 beta** expands into Fabric 1.20.x and Forge/Fabric 1.21.1. It translates mod screens, buttons, tooltips, guidebooks and HUD while leaving vanilla/player chat alone. It is client-only, writes no save-critical world data, adds no gameplay registries and is highly removable. Network/privacy caveat: uncached mod text can be sent to the translation service. [Stable Forge file](https://www.curseforge.com/minecraft/mc-mods/babelcraft/files/8397918).
- **#5 / #23 CustomNPCs creator/compatibility specialist:** [CNPC-plus-Fix 3.2.2](https://www.curseforge.com/minecraft/mc-mods/cnpc-plus-fix) has exact Forge 1.20.1 and NeoForge 1.21.1 routes, plus legacy 1.12.2 support. It focuses on CustomNPCs fixes, performance, compatibility and creator workflow without adding new blocks/items. Back up CustomNPCs-heavy maps/projects before changing creator-stack dependencies. [Forge 1.20.1 file](https://www.curseforge.com/minecraft/mc-mods/cnpc-plus-fix/files/8708175).
- **#21 / #22 Bedrock admin specialist and #16 technical runner:** [GeoGen — A Bedrock World Pre-Generator 1.2.1](https://www.curseforge.com/minecraft-bedrock/addons/geogen) uses native generation and stable scripting to pre-generate Overworld, Nether, End and custom dimensions, with queue, pause/resume and recovery. No direct LevelDB editing, Beta APIs or experimental gameplay are required. Generated chunks intentionally persist after removal, so this is an admin/world-preparation tool rather than a client-safe visual layer. [1.2.1 release](https://www.curseforge.com/minecraft-bedrock/addons/geogen/files/8708960).
- **#34 mermaid/player-transformation specialist:** [Mermod](https://www.curseforge.com/minecraft/mc-mods/mermod-neoforge) closes a major taxonomy gap. The maintained direct 1.20.1 route is **3.0.3** for Forge/NeoForge, with a separate Fabric 3.0.3 project line; the modern line is **4.0.2** for 1.21.11. It is a necklace-driven player-form/gameplay system, **not** a standalone female mob. Architectury API is required from 3.0.0+. Revert transformation/equipment and back up before backend removal.
- **#34 mermaid transformation extension:** [Mermaid origins 1.1.4.1](https://www.curseforge.com/minecraft/mc-mods/mermaid-origins) is a Forge 1.20.1 Mermod extension. World-specific configuration lives inside the save's `serverconfig` area, so it is tracked as transformation companion infrastructure rather than a native mob/model pack. [Exact 1.20.1 file](https://www.curseforge.com/minecraft/mc-mods/mermaid-origins/files/7250220).
- Overall category leaders are unchanged. **#5, #21, #22, #23, #26, #27, #29 and #34** gained qualified depth. Animated, CIT, 3D, Bedrock Marketplace, Planet Minecraft artistic/map, Afdian/BOOTH creator-hub and PitonixRex sweeps produced **NO MATERIAL LEADER CHANGE**.

## Earlier 2026-08-25 deltas retained

- **#8 / #6 livestock:** [DragN's Livestock Overhaul 3.9.4](https://www.curseforge.com/minecraft/mc-mods/dragns-livestock-overhaul/files/8728287) remains a substantive Forge 1.20.1 livestock/genetics/farming runner. It adds persistent entities/items/genetics/farming state; **TEST BRANCH**.
- **#7 / #8 exploration:** [Aquamirae 7.2.3](https://www.curseforge.com/minecraft/mc-mods/aquamirae/files/8728350) is the refreshed NeoForge 1.21.1 route; Forge 1.20.1 remains on its separate line.
- **#8 horror mobs:** [Sons Of Sins 2.2.1b](https://www.curseforge.com/minecraft/mc-mods/sons-of-sins/files/8728502) is current on NeoForge 1.21.1; Forge 1.20.1 remains 2.2.1.
- **#34 companion layer:** [NFF: Girls 0.2.33.1](https://www.curseforge.com/minecraft/mc-mods/nff-girls/files/8726332) remains the refreshed Forge 1.20.1 HMaG friendship/companion route with persistent relationship/custom save data.
- **#9 hybrid combat:** [Better Combat × Epic Fight Compat 1.0.2](https://www.curseforge.com/minecraft/mc-mods/better-combat-epic-fight-compat) remains a direct Forge 1.20.1 bridge. **TEST BRANCH.**
- **#9 / #26 / #27 / #29 firearm animation:** [Epic Fight × TacZ First-Person Compat 0.5.2](https://www.curseforge.com/minecraft/mc-mods/epic-fight-tacz-first-person-compat) remains a client-only Forge 1.20.1 / NeoForge 1.21.1 bridge with no save dependency.
- **#13 visual compatibility:** [Glowing Emissive Ores 1.110.0](https://www.curseforge.com/minecraft/texture-packs/glowing-emissive-ores), [Glowing Emissive Ores Definitive Edition 0.3.5](https://www.curseforge.com/minecraft/mc-mods/glowing-emissive-ores-definitive-edition), and [Stay True Compats Reforged](https://www.curseforge.com/minecraft/texture-packs/stay-true-compats) remain qualified world-safe/client visual runners.
- **#16 Bedrock technical/worldgen:** [Bedrock Energistics 0.15.0](https://www.curseforge.com/minecraft-bedrock/addons/bedrock-energistics) and [Biomes O' Discovery 4.6](https://www.curseforge.com/minecraft-bedrock/addons/bod) remain persistent-content runners; [Canopy](https://www.curseforge.com/minecraft-bedrock/addons/canopy) remains #16 overall.
- **#34 maid automation/cooking/combat/spells:** [MaidUseHandCrank](https://www.curseforge.com/minecraft/mc-mods/maidusehandcrank), [Maid's Bakeries 1.0.3](https://www.curseforge.com/minecraft/mc-mods/maid-s-bakeries), [MaidSwordSoaring 1.0.7](https://www.curseforge.com/minecraft/mc-mods/maidswordsoaring), [Touhou Little Maid: Spell](https://www.curseforge.com/minecraft/mc-mods/touhou-little-maid-spell), and [Maid Form Shift 2.0.0](https://www.curseforge.com/minecraft/mc-mods/maid-form-shift-lmrb-tlm-backend-showcase) remain distinct automation/cooking/movement-combat/spell/transformation tracks.
- **#34 leader:** [Monster Girl 2.7.0](https://www.curseforge.com/minecraft/mc-mods/monster-girl) remains the practical Forge 1.20.1 overall leader. Its exact 1.20.1 file carries a critical warning that deletion after loading a save can make that save unplayable. **BACKUP REQUIRED.**
- **#4 / #11 Create performance:** [StellarCreateOptimization 1.0.2](https://www.curseforge.com/minecraft/mc-mods/stellarcreateoptimization) remains the fresh Create-heavy benchmark-first optimization candidate.
- **#9 advanced Epic Fight:** [Epic Fight Indestructible - Unofficially Enhanced!! 20.14.17](https://www.curseforge.com/minecraft/mc-mods/epic-fight-indestructible-unofficially-enhanced) remains the advanced dynamic mob-patching route.
- **#34 maid/social/worker/logistics:** [Touhou Little Maid: Love & Loathe 2.0.6](https://www.curseforge.com/minecraft/mc-mods/touhou-little-maid-love-loathe), [Sable: MaidRagdoll 0.11-beta](https://www.curseforge.com/minecraft/mc-mods/sable-maidragdoll), [HMaG Spells 1.1.0](https://www.curseforge.com/minecraft/mc-mods/hmag-spells), [Maid Storage Manager 1.15.6](https://www.curseforge.com/minecraft/mc-mods/maid-storage-manager), [Maid Useful Tasks 1.4.2](https://www.curseforge.com/minecraft/mc-mods/maid-useful-tasks), [TLM: True POWER 1.2.3](https://www.curseforge.com/minecraft/mc-mods/tlm-true-power), and [Maid Logistics Network 1.3.4](https://www.curseforge.com/minecraft/mc-mods/maid-logistics-network) remain qualified specialist tracks.
- **#33 geometry:** [Hellim's 3D Blocks v1.0](https://www.curseforge.com/minecraft/texture-packs/hellims-3d-blocks) remains the dependency-free direct-1.20.1 challenger; [3D Default 1.15.0](https://modrinth.com/resourcepack/3d-default/version/1.15.0) remains overall #33.

## Category-board change ledger

- **#5 Compatibility / Fix:** [Collections Of Optimizations 2.4](https://www.curseforge.com/minecraft/mc-mods/collections-of-optimizations) remains leader; [CNPC-plus-Fix 3.2.2](https://www.curseforge.com/minecraft/mc-mods/cnpc-plus-fix) is the new CustomNPCs creator/compatibility runner.
- **#6 Gameplay / QoL:** [TargetsIndicate 1.2.0](https://www.curseforge.com/minecraft/mc-mods/targetsindicate) remains leader; [DragN's Livestock Overhaul 3.9.4](https://www.curseforge.com/minecraft/mc-mods/dragns-livestock-overhaul/files/8728287) is a farming/livestock systems runner.
- **#7 Exploration / Worldgen / Dimension:** [Jaden's Nether Expansion 2.4.1](https://modrinth.com/mod/jadens-nether-expansion) remains leader; [Aquamirae 7.2.3](https://www.curseforge.com/minecraft/mc-mods/aquamirae/files/8728350) is the refreshed modern ocean-adventure runner.
- **#8 Mob / Creature:** [Psychopath 1.4.0](https://www.curseforge.com/minecraft/mc-mods/psychopath) remains leader; [DragN's Livestock Overhaul 3.9.4](https://www.curseforge.com/minecraft/mc-mods/dragns-livestock-overhaul/files/8728287) and [Sons Of Sins 2.2.1b](https://www.curseforge.com/minecraft/mc-mods/sons-of-sins/files/8728502) remain qualified runners.
- **#9 Combat / Animation Compatibility:** [Epic Fight Guard Fix 1.0.9](https://www.curseforge.com/minecraft/mc-mods/epic-fight-guard-fix) remains baseline leader; hybrid/firearm specialists remain tracked.
- **#11 Create / Technology:** [Create: Storage 1.2.7](https://www.curseforge.com/minecraft/mc-mods/create-storage-neo-forge) remains leader; maid-factory and optimization specialists remain tracked.
- **#13 Resource Pack:** [Optimum Realism](https://www.curseforge.com/minecraft/texture-packs/optimum-realism) remains leader; emissive/compatibility packs remain qualified runners.
- **#16 Bedrock Add-On:** [Canopy](https://www.curseforge.com/minecraft-bedrock/addons/canopy) remains leader; [GeoGen 1.2.1](https://www.curseforge.com/minecraft-bedrock/addons/geogen) joins as a technical admin runner alongside current content/worldgen candidates.
- **#21 Tool / Utility / Launcher / Modpack Management:** [Modrinth App](https://modrinth.com/app) / [Prism Launcher](https://prismlauncher.org/) remain overall leaders; [GeoGen 1.2.1](https://www.curseforge.com/minecraft-bedrock/addons/geogen) joins the Bedrock admin/tool specialist track.
- **#22 Server / Admin / Diagnostics:** [spark 1.10.173](https://modrinth.com/mod/spark) remains leader; [GeoGen 1.2.1](https://www.curseforge.com/minecraft-bedrock/addons/geogen) joins as a Bedrock world-preparation specialist.
- **#23 Modding / Development Tool or Library:** [Easy Model Entities 2.3.0](https://modrinth.com/mod/easy-model-entities) remains leader; [CNPC-plus-Fix 3.2.2](https://www.curseforge.com/minecraft/mc-mods/cnpc-plus-fix) joins the CustomNPCs creator-stack track.
- **#26 / #27 / #29 client-safe tracks:** [BabelCraft 1.2.5.3](https://www.curseforge.com/minecraft/mc-mods/babelcraft) joins the translation/accessibility track; the broad leaders remain [TargetsIndicate](https://www.curseforge.com/minecraft/mc-mods/targetsindicate), [BetterF3](https://modrinth.com/mod/betterf3), [Client Tweaks](https://www.curseforge.com/minecraft/mc-mods/client-tweaks), [Legendary Block Entities](https://www.curseforge.com/minecraft/mc-mods/legendary-block-entities), and [ImmediatelyFast](https://modrinth.com/mod/immediatelyfast) as applicable.
- **#31 Animated:** [Actions & Stuff 1.11](https://www.minecraft.net/en-us/marketplace/pdp?id=53c7af69-2425-490e-8e9a-8ad2c2e7cbfe) overall / [Fresh Animations](https://modrinth.com/resourcepack/fresh-animations) Java — **NO MATERIAL LEADER CHANGE**.
- **#32 CIT:** [Kaydicraft CIT 1.45](https://modrinth.com/resourcepack/kaydicraft) / [FAYE](https://www.planetminecraft.com/texture-pack/the-faye-wallpaper-amp-flooring-set-cit/) — **NO MATERIAL LEADER CHANGE**.
- **#33 3D:** [3D Default 1.15.0](https://modrinth.com/resourcepack/3d-default/version/1.15.0) — **NO MATERIAL LEADER CHANGE**.
- **#34 Monster Girl / Female-Coded Mob:** [Monster Girl 2.7.0](https://www.curseforge.com/minecraft/mc-mods/monster-girl) remains overall leader. [Mermod](https://www.curseforge.com/minecraft/mc-mods/mermod-neoforge) is the new best mermaid/player-transformation specialist; [Mermaid origins](https://www.curseforge.com/minecraft/mc-mods/mermaid-origins) is a 1.20.1 transformation extension. [HMaG](https://www.curseforge.com/minecraft/mc-mods/hostile-mobs-and-girls) + [NFF: Girls](https://www.curseforge.com/minecraft/mc-mods/nff-girls) retain bestiary/companion strength.
- **All remaining categories:** independently rechecked; **NO MATERIAL LEADER CHANGE** this sweep.

## Afdian / BOOTH / Asian creator-source status

- [omomomomomo](https://afdian.com/a/omomomomomomo) remains a permanent **MANUAL-AUDIT** discovery source where public product indexing is incomplete.
- [你个人机cc](https://afdian.com/a/ccnie), [Feather_aya](https://afdian.com/a/FliegeSA), and [映白](https://afdian.com/a/ehaku) remain verified official source hubs; no defensible new product-level row was invented from sparse indexing this cycle.
- [AllTheYSM model index](https://alltheysm.top/models) remains discovery-only; original creator/source provenance is required before promotion.
- The PitonixRex author collection was rechecked with no material new portfolio-level promotion; current and legacy individual entries remain preserved in the companion tracker.

## Safety notes — newest additions

| Project | Client only? | Server required? | Save/world impact | Removal | Update friendliness |
|---|---|---|---|---|---|
| [BabelCraft](https://www.curseforge.com/minecraft/mc-mods/babelcraft) | Yes | No | Local config/translation cache only | Very high | Excellent |
| [CNPC-plus-Fix](https://www.curseforge.com/minecraft/mc-mods/cnpc-plus-fix) | No | Yes for matched multiplayer creator behavior | Existing CustomNPCs project/NPC data; no new blocks/items/worldgen surfaced | Medium-high with project backup | Good with dependency caution |
| [GeoGen](https://www.curseforge.com/minecraft-bedrock/addons/geogen) | No | World/operator-side tool | Intentionally generates and saves vanilla terrain/chunks | Medium-high; generated chunks remain | Good |
| [Mermod](https://www.curseforge.com/minecraft/mc-mods/mermod-neoforge) | No | Yes for shared gameplay | Registered necklace/item + player transformation state; no native mob/worldgen | Medium; revert form/equipment and back up | Good with dependency matching |
| [Mermaid origins](https://www.curseforge.com/minecraft/mc-mods/mermaid-origins) | No | Yes | World-specific server config + Mermod-linked gameplay state | Medium; back up before backend changes | Good with caution |

## Persistence state

- Canonical Google Sheet: **124 retained serious projects/finds / 34 categories**.
- Separate female-mob companion set: **99 curated projects/source hubs / 50 explicit 1.20.1 routes / 19 maid-ecosystem entries / 12 source hubs / 2 Bedrock routes**.
- Every specifically named project/creator newly written in this revision is directly linked to its official/source page.
- Full current checkpoint: [`reports/daily-scout/2026-08-25-category-leaderboard.md`](../2026-08-25-category-leaderboard.md).
- Historical findings remain preserved in dated reports, the canonical Sheets/Docs, this master index, and Git history.

## Recurring-service rules

- A no-change category is a valid result; do not invent churn.
- Maintain the canonical Google Sheet + curated Doc in place and keep the separate female-mob companion artifacts synchronized when that track materially changes.
- Preserve exact versions, direct source/version links, loader/channel, side requirements, save impact, removability, compatibility notes, creator provenance, and ranking history.
- Every material run publishes a dated checkpoint to both Drive and GitHub and verifies both copies.
- The recurring scout remains active unless the user explicitly retires it.
