# Minecraft Ecosystem Category Scout — 2026-08-24

**Run type:** material update + late-cycle delta  
**Coverage:** Java + Bedrock • 33 permanent categories • client/world safety • Animated/CIT/3D visual tracks  
**Canonical tracker after this pass:** 81 retained serious projects/finds

## Material changes — earlier Aug 24 pass

### #7 Best Exploration / Worldgen / Dimension — NEW LEADER

**Jaden's Nether Expansion 2.4.1** is the modern leader.

- Java 1.21.1, NeoForge, Release, updated 2026-08-22.
- Required on the 2.4.1 line: Lodestone + Elysium API.
- Forge 1.20.1 remains on legacy **2.3.5** and receives bug fixes only; the maintained forward line is 1.21.1 NeoForge.
- Adds worldgen, biomes, mobs, blocks/items and progression, so treat it as world-affecting and do not casually remove it from an established save.
- Project: https://modrinth.com/mod/jadens-nether-expansion
- Exact 2.4.1: https://modrinth.com/mod/jadens-nether-expansion/version/2.4.1

**Aquamirae 7.2.1** remains a strong #7/#8 runner-up.

- 1.21.1 stable: NeoForge plus Fabric/Quilt line.
- Forge/Fabric 1.20.1 active line: **7.1.13 Beta**, updated 2026-08-18.
- Conservative Forge 1.20.1 stable line: **6.4.0**.
- World-affecting exploration/mob/structure content; back up before adoption/removal.
- Project: https://www.curseforge.com/minecraft/mc-mods/aquamirae
- Exact Forge 1.20.1 beta: https://www.curseforge.com/minecraft/mc-mods/aquamirae/files/8678729

### #22 Best Server / Admin / Diagnostics — CATEGORY FILLED

**spark 1.10.173** remains the mature overall leader.

- Cross-loader profiler for clients, servers and proxies.
- Current 1.10.173 line ships on Forge/Fabric/NeoForge 26.2; the project keeps broad older-version coverage.
- Project: https://modrinth.com/mod/spark
- CurseForge: https://www.curseforge.com/minecraft/mc-mods/spark

**Shinoyuki-BetterAutoSave 0.20.1** remains the Forge 1.20.1 / NeoForge 1.21.1 specialist runner-up.

- Server-side only; clients do not need it.
- Forge 1.20.1 requires Forge 47.3.22+ and Java 17; use the `-all` jar.
- NeoForge 1.21.1 requires NeoForge 21.1 and Java 21.
- Moves save serialization/I/O off the main thread and records sync chunk-load stalls / long inter-tick gaps.
- **0.20.1** fixes the 0.20.0 startup crash with known Lithium-family ports by gating the sync-load detector around conflicting overwrites; saving and async loading remain available.
- Do not combine with overlapping async-save systems such as Smooth Chunk Save / Fast Async World Save or conflicting C2ME save-side functionality.
- Safety: touches the world-save pipeline but adds no gameplay registries/worldgen/items/blocks/entities. Back up before first test. Async loading is opt-in/off by default.
- Project: https://modrinth.com/mod/shinoyuki-betterautosave
- 0.20.1 notes: https://github.com/ShinoyukiMiyako/Shinoyuki-BetterAutoSave/blob/main/release-notes/v0.20.1.en.md

### #16 Bedrock Add-On — WATCH RETAINED

**Canopy 1.6.1 remains overall.** **Sweet Dreams — Bug Fixes 26.mcaddon** remains tracked as a strong horror/survival watch candidate.

- Surfaced current file: Bedrock 26.30, updated 2026-08-23.
- Behavior/content add-on, therefore world-affecting; validate on the installed Bedrock build and back up before using on a long-lived world.
- Project: https://www.curseforge.com/minecraft-bedrock/addons/sweet-dreams

## Late-cycle material delta — 2026-08-24

### #4 / #28 — Just Enough Threads [JEI Startup Optimize] — NEW SPECIALIST RUNNER-UP

**Just Enough Threads** is the strongest new current-pack performance discovery in this pass.

- Java; client-side only.
- Direct Forge 1.20.1 path plus NeoForge 1.21.1.
- Modern current line: 0.10.1; project surfaced current again on 2026-08-24.
- Forge 1.20.1 prerequisites: Forge 47.4.4+ and JEI 15.20.0.120+.
- NeoForge 1.21.1 prerequisites: NeoForge 21.1.x and JEI 19.27.0.340.
- Moves JEI ingredient-search index construction off-thread after world entry; parallel pre-resolution is configurable.
- 0.10.1 hardens startup interactions with FTB Quests + Ixeris and prevents recipe GUI access while JEI is half-initialized.
- Project-published large-pack example reduced JEI `Building` runtime from about 5.18s to 0.55s and `Starting JEI` total from about 10.7s to 6.4s; treat these as project measurements, not guaranteed pack-wide gains.
- **CLIENT ONLY?** Yes. **SERVER REQUIRED?** No. **SAVE/WORLD DATA?** Client config/runtime only. **SAFE TO REMOVE?** High confidence. **UPDATE FRIENDLINESS:** Excellent. **REGISTRIES/WORLDGEN/CONTENT?** None.
- Status: **TEST NOW** once the Forge/JEI minimums are satisfied.
- Project: https://www.curseforge.com/minecraft/mc-mods/just-enough-threads
- Files: https://www.curseforge.com/minecraft/mc-mods/just-enough-threads/files

Legendary Block Entities remains #4/#28 overall because it is broader renderer performance infrastructure; Just Enough Threads is a targeted JEI startup/search-index specialist.

### #5 / #23 — Fusion (Connected Textures) 1.3.14 — FRESH INFRASTRUCTURE RUNNER-UP

**Fusion 1.3.14** landed on 2026-08-23 with a direct Forge 1.20.1 build and broad modern loader/version coverage.

- Adds connected-texture/custom-model infrastructure for resource packs and consuming mods.
- 1.3.14 fixes `DefaultModelTypes#ITEM_MODEL_GENERATOR` registration.
- Recent 1.3.13 fixes cover empty geometry handling, `random` subtexture crashes, item-model face visibility and pane culling.
- 1.3.11 substantially reduced Fusion-model memory usage.
- Modern lines also include fixes around PBR/normal-specular integration, renderer/model lighting and FramedBlocks-related model behavior.
- **CLIENT ONLY?** Partial/build/use dependent. **SERVER REQUIRED?** Depends on the consuming mod/pack. **SAVE/WORLD DATA?** No gameplay world data itself. **SAFE TO REMOVE?** High only when no installed mod/resource pack depends on Fusion features. **UPDATE FRIENDLINESS:** Excellent when dependency-managed. **REGISTRIES/WORLDGEN/CONTENT?** No gameplay-content focus.
- Status: **UPDATE IF USED / STRONG SUPPORT LAYER**.
- Project: https://www.curseforge.com/minecraft/mc-mods/fusion-connected-textures
- Files: https://www.curseforge.com/minecraft/mc-mods/fusion-connected-textures/files

Collections Of Optimizations remains #5 overall and Easy Model Entities remains #23 overall; Fusion is the fresh resource/model-infrastructure runner-up.

### #32 — Cozy CIT: Furniture Comes Alive — NEW MODERN CIT UTILITY WATCH

**Cozy CIT: Furniture Comes Alive** is a genuinely novel CIT-ecosystem bridge for Java 1.21.1 Fabric/NeoForge.

- Current surfaced line: 1.0.0; initial release 2026-08-19, project surfaced updated/current on 2026-08-24.
- This is **not a CIT pack and not a CIT renderer**.
- It reads supported CIT/resource packs from the `resourcepacks` folder at launch and dynamically registers their furniture models as real functional blocks.
- Mizuno's 16 Craft alone yields 820+ furniture blocks; the supported-pack set can reach roughly 1,950+.
- No OptiFine, anvil renaming or item-frame placement workflow is required for the generated furniture.
- Adds functional drawers/cabinets/fridges, lamps, sinks, bed frames and placement behavior.
- Required baseline includes Mizuno's 16 Craft CIT + regular Mizuno 16 Craft resource pack; Fabric API on Fabric.
- Multiplayer requires the mod and matching pack sets on server/clients.
- **CLIENT ONLY?** No. **SERVER REQUIRED?** Yes in multiplayer. **SAVE/WORLD DATA?** Persistent registered furniture blocks/containers. **SAFE TO REMOVE?** Low confidence. **UPDATE FRIENDLINESS:** Caution. **REGISTRIES/CONTENT?** Yes.
- Status: **FUTURE VERSION / WATCH** due very low current adoption and persistent world content.
- Project: https://www.curseforge.com/minecraft/mc-mods/cozy-cit-furniture-comes-alive

Kaydicraft CIT/FAYE remain the #32 pack leaders. Cozy CIT is tracked separately as experimental tooling/content infrastructure.

## Specialty visual recheck

- **#31 Animated Resource Pack — NO MATERIAL LEADER CHANGE.** Actions & Stuff 1.11 remains overall/Bedrock; Fresh Animations remains Java. Use Fresh Animations 1.10.4 for Java 1.20.x; 1.10.5 targets 26.1.x/26.2 and intentionally removed pre-26.1 support. EMF + ETF remain recommended.
  - https://modrinth.com/resourcepack/fresh-animations/version/1.10.4
  - https://modrinth.com/resourcepack/fresh-animations/version/1.10.5
- **#32 CIT Pack — PACK LEADERS UNCHANGED.** Kaydicraft CIT 1.45 remains the classic 1.20.1 single-pack leader; FAYE remains the active classic-CIT suite; ItemBound: ReBound remains in the modern item-model-definition track. Cozy CIT is now tracked as a world-affecting utility watch, not a replacement pack.
- **#33 3D Resource Pack — NO MATERIAL LEADER CHANGE.** 3D Default 1.15.0 remains overall and has an exact `mc1.20-1.21.1` file. Better 3D Blocks remains the high-detail runner-up, but its maintainer still warns of a Distant Horizons interaction that can lose up to roughly 80 FPS. Its teased 2026-09-01 update is not treated as released.
  - https://modrinth.com/resourcepack/3d-default/version/1.15.0
  - https://modrinth.com/resourcepack/better-3d-blocks

## Bedrock / Marketplace / Planet Minecraft late-cycle sweep

No late-cycle Bedrock add-on, Marketplace item, Bedrock resource/behavior pack or Planet Minecraft map/build surfaced with enough verified quality and freshness to justify a leader replacement. #16 Canopy, #17 Newb X Supplementary, #18 Auto Factory and #19/#20 Trident Cliffs City therefore remain unchanged rather than padding the board.

## 33-category scoreboard

1. Best Overall Java Mod — Create: Storage 1.2.7 — NO MATERIAL CHANGE
2. Best Forge/NeoForge Mod — Create: Storage 1.2.7 / Easy NPC modern — NO MATERIAL CHANGE
3. Best Fabric/Quilt Mod — Easy NPC 7.9.0 — NO MATERIAL CHANGE
4. Best Performance/Optimization Mod — **Legendary Block Entities 0.11.0** — NEW SPECIALIST RUNNER-UP: Just Enough Threads
5. Best Compatibility/Fix Mod — **Collections Of Optimizations 2.4** — FRESH RUNNER-UP: Fusion 1.3.14
6. Best Gameplay/QoL Mod — TargetsIndicate 1.2.0 — NO MATERIAL CHANGE
7. Best Exploration/Worldgen/Dimension Mod — **Jaden's Nether Expansion 2.4.1** — leader retained from earlier Aug 24 pass
8. Best Mob/Creature Mod — Psychopath 1.4.0 — Aquamirae 7.2.1 runner-up retained
9. Best Combat/Animation Mod or Compatibility Layer — Epic Fight Guard Fix 1.0.9 — NO MATERIAL CHANGE
10. Best Movement/Skating/Grinding/Vehicle Mod — I Wanna Skate 1.2.0 — NO MATERIAL CHANGE
11. Best Create/Technology Mod or Addon — Create: Storage 1.2.7 — NO MATERIAL CHANGE
12. Best Building/Decoration Mod — Create: Copycats+ 3.0.8 — NO MATERIAL CHANGE
13. Best Resource Pack — Optimum Realism 64x — NO MATERIAL CHANGE
14. Best Shader Pack / Shader-Adjacent Visual Tech — Complementary Reimagined — NO MATERIAL CHANGE; PathMax remains experimental runner-up
15. Best Datapack — NO MATERIAL CHANGE
16. Best Bedrock Add-On — Canopy 1.6.1 — Sweet Dreams watch retained
17. Best Bedrock Resource/Behavior Pack — Newb X Supplementary 8.0 — NO MATERIAL CHANGE
18. Best Bedrock Marketplace Find — Auto Factory — NO MATERIAL CHANGE
19. Best Minecraft Map/Adventure Map — Trident Cliffs City — NO MATERIAL CHANGE
20. Best Artistic/Showcase Build or Map — Trident Cliffs City — NO MATERIAL CHANGE
21. Best Tool/Utility/Launcher/Modpack Management — Modrinth App / Prism Launcher — NO MATERIAL CHANGE
22. Best Server/Admin/Diagnostics Tool — spark 1.10.173 — BetterAutoSave 0.20.1 specialist retained
23. Best Modding/Development Tool or Library — **Easy Model Entities 2.3.0** — FRESH RUNNER-UP: Fusion 1.3.14
24. Best Experimental/Bleeding-Edge Project — Retromod 1.3.0-snapshot.8 — NO MATERIAL CHANGE
25. Best Future-Version Project Worth Tracking — Retromod 1.3.0-snapshot.8 — NO MATERIAL CHANGE
26. Best Client-Side-Only Mod — TargetsIndicate 1.2.0 — NEW SPECIALIST CANDIDATE: Just Enough Threads
27. Best Update-Friendly / World-Safe Mod — TargetsIndicate / BetterF3 / Client Tweaks / Legendary Block Entities — NO LEADER CHANGE
28. Best Client-Side Performance Mod — **Legendary Block Entities / ImmediatelyFast** — NEW SPECIALIST RUNNER-UP: Just Enough Threads
29. Best Client-Side QoL/UI Mod — TargetsIndicate 1.2.0 — NO MATERIAL CHANGE
30. Best Removable-With-Minimal-Save-Risk Mod — TargetsIndicate / BetterF3 — NO MATERIAL CHANGE
31. Best Animated Resource Pack — Actions & Stuff 1.11 overall; Fresh Animations Java — NO MATERIAL LEADER CHANGE
32. Best CIT Pack — Kaydicraft CIT 1.45 / FAYE suite — PACK LEADERS UNCHANGED; NEW UTILITY WATCH: Cozy CIT
33. Best 3D Resource Pack — 3D Default 1.15.0 — NO MATERIAL LEADER CHANGE; Better 3D Blocks remains DH-caution runner-up

## Evergreen client stack

**Leader stack remains:** Legendary Block Entities + ImmediatelyFast + Client Tweaks + BetterF3 + Mouse Tweaks + TargetsIndicate, with shaders/resource packs kept as removable presentation layers.

**New optional specialist:** Just Enough Threads can be layered into a JEI-heavy client once its exact Forge/JEI minimums are met. It is client-only and save-neutral, but because it touches JEI startup/index lifecycle it should be tested with JEI integrations before being treated as universal baseline.

## Persistence delta

- Canonical Google Sheet audited and updated 2026-08-24.
- Permanent categories: **33**.
- Retained serious projects/finds: **81**.
- Three durable late-cycle additions: Just Enough Threads, Fusion (Connected Textures), Cozy CIT: Furniture Comes Alive.
- Earlier Aug 24 additions retained: Jaden's Nether Expansion, Aquamirae, Shinoyuki-BetterAutoSave, Sweet Dreams.
- Canonical readable Google Doc updated in place.
- No historical entries deleted; prior deltas remain in earlier sections/checkpoints.
