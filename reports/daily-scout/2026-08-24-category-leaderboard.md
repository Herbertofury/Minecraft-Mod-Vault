# Minecraft Ecosystem Category Scout — 2026-08-24

**Run type:** material update + late-cycle deltas  
**Coverage:** Java + Bedrock • 33 permanent categories • client/world safety • Animated/CIT/3D visual tracks  
**Canonical tracker after this pass:** 87 retained serious projects/finds

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

**Canopy 1.6.1 remains overall.** **Sweet Dreams — Bug Fixes 27.mcaddon** is now the tracked horror/survival watch candidate.

- Exact refreshed file: Bedrock 26.30, updated 2026-08-24.
- No changelog was supplied for Bug Fixes 27, so no behavioral delta is invented.
- Behavior/content add-on, therefore world-affecting; validate on the installed Bedrock build and back up before using on a long-lived world.
- Project: https://www.curseforge.com/minecraft-bedrock/addons/sweet-dreams
- Exact file: https://www.curseforge.com/minecraft-bedrock/addons/sweet-dreams/files/8720676

## Late-cycle material delta — 2026-08-24

### #4 / #28 — Just Enough Threads [JEI Startup Optimize] — NEW SPECIALIST RUNNER-UP

**Just Enough Threads** is a strong current-pack performance specialist.

- Java; client-side only.
- Direct Forge 1.20.1 path plus NeoForge 1.21.1.
- Modern current line: 0.10.1.
- Forge 1.20.1 prerequisites: Forge 47.4.4+ and JEI 15.20.0.120+.
- NeoForge 1.21.1 prerequisites: NeoForge 21.1.x and JEI 19.27.0.340.
- Moves JEI ingredient-search index construction off-thread after world entry; parallel pre-resolution is configurable.
- 0.10.1 hardens startup interactions with FTB Quests + Ixeris and prevents recipe GUI access while JEI is half-initialized.
- **CLIENT ONLY?** Yes. **SERVER REQUIRED?** No. **SAVE/WORLD DATA?** Client config/runtime only. **SAFE TO REMOVE?** High confidence. **UPDATE FRIENDLINESS:** Excellent. **REGISTRIES/WORLDGEN/CONTENT?** None.
- Project: https://www.curseforge.com/minecraft/mc-mods/just-enough-threads

### #5 / #23 — Fusion (Connected Textures) — FRESH INFRASTRUCTURE HOTFIX

**Fusion 1.3.14a** is now current on Forge/NeoForge 1.21.11, while the direct Forge 1.20.1 line remains **1.3.14**.

- 1.3.14a fixes Forge `render_type` invisibility when no other custom render type is used and fixes modded Forge custom-geometry models being baked as regular models.
- The 1.20.1 build remains 1.3.14; do not label it 1.3.14a.
- Fusion adds no gameplay world content itself, but dependent packs/mods can require its model features.
- Project: https://www.curseforge.com/minecraft/mc-mods/fusion-connected-textures

### #32 — Cozy CIT: Furniture Comes Alive — MODERN CIT UTILITY WATCH

**Cozy CIT: Furniture Comes Alive** remains a novel Java 1.21.1 Fabric/NeoForge CIT-ecosystem bridge.

- It is **not a CIT pack and not a CIT renderer**.
- It dynamically registers supported CIT furniture models as real functional blocks.
- Multiplayer pack sets should match; placed furniture is persistent world content.
- **CLIENT ONLY?** No. **SERVER REQUIRED?** Yes in multiplayer. **SAVE/WORLD DATA?** Persistent registered furniture blocks/containers. **SAFE TO REMOVE?** Low confidence. **UPDATE FRIENDLINESS:** Caution.
- Project: https://www.curseforge.com/minecraft/mc-mods/cozy-cit-furniture-comes-alive

## Late-cycle material delta II — 2026-08-24

### #8 — Sons Of Sins 2.2.1 — NEW SAME-DAY RUNNER-UP

- Direct **Forge 1.20.1** and **NeoForge 1.21.1** 2.2.1 releases on 2026-08-24.
- Established, substantive horror/creature project with 15M+ downloads; retained despite an MCreator tag because it is not shovelware.
- Exact 2.2.1 changelog was not surfaced, so no feature change is invented.
- World-affecting: persistent creatures/items/blocks/content. Back up before removal and test spawn density/AI load in mob-heavy packs.
- Status: **TEST BRANCH / OPTIONAL**.
- Project: https://www.curseforge.com/minecraft/mc-mods/sons-of-sins

### #9 — Better Selective Combat 1.0.0 — NEW COMPATIBILITY RUNNER-UP

- Fabric/NeoForge for 1.21.10–1.21.11 and 26.1.2/26.2.
- Lets selected weapons bypass Better Combat and use their original behavior.
- Config-driven; adds no world content. LuckPerms is optional for permission management.
- No Forge 1.20.1 build, so this is **FUTURE VERSION / WATCH** rather than a current-pack solution.
- Project: https://www.curseforge.com/minecraft/mc-mods/better-selective-combat

### #26 / #27 / #29 — Particle Rain — NEW CLIENT-SAFE VISUAL RUNNER-UP

- Fresh **v4-beta.11** landed 2026-08-24 for modern Fabric lines with 26.2 support, rolling block particles, wind/biome-border/mist fixes and configuration hardening.
- Direct **Forge 1.20.1** path remains **v4-beta.10**, uploaded 2026-05-08.
- **CLIENT ONLY?** Yes. **SERVER REQUIRED?** No. **SAVE/WORLD DATA?** Client config/runtime only. **SAFE TO REMOVE?** High. **UPDATE FRIENDLINESS:** Excellent. **REGISTRIES/WORLDGEN/CONTENT?** None.
- Test particle/haze interaction with Oculus/shaders before treating it as baseline.
- Project: https://modrinth.com/mod/particle-rain
- Forge 1.20.1: https://modrinth.com/mod/particle-rain/version/v4-beta.10%2B1.20.1-forge

### #33 / #13 — Nautilus 3D — NEW SERIOUS 3D RUNNER-UP

- Current verified release: **26.2**; project history covers 1.20.x/1.21.x/26.x.
- Vanilla-style true 3D block/item geometry with a broad feature set.
- Current 26.2 work includes chest/bell fixes and Better Block Entities compatibility improvements; the project reports a significant performance benefit for that BBE path.
- Optional feature paths include EMF/ETF, CIT Resewn, Continuity and Respackopts.
- Save-neutral/resource-pack-only; high-confidence removable.
- Distant Horizons behavior was not verified, so **3D Default remains the safer overall 1.20.1/DH-first choice**.
- Project: https://modrinth.com/resourcepack/nautilus3d
- 26.2: https://modrinth.com/resourcepack/nautilus3d/version/a5PFyRlb

### #13 / #33 — "Barely Default" by Mickey Joe V20.21 — NEW MODERN VISUAL RUNNER-UP

- Current file targets 26.x (26.1–26.2 family), not Forge 1.20.1.
- Combines custom entity models, true 3D block/item models, CTM and random textures in a 16x vanilla-inspired style.
- Resource-pack only: no world content, highly removable/update-friendly.
- Project: https://www.curseforge.com/minecraft/texture-packs/mickey-joes-relatively-improved-default

### #21 / #26 — Crash Assistant 1.11.12 — NEW DIAGNOSTICS RUNNER-UP

- Client-side diagnostic mod with Forge/Fabric/NeoForge/Quilt coverage.
- Direct Forge 1.20.1 build exists.
- 1.11.12 adds persistent launch/mod-list history and changes copied-log-link handling/privacy flow.
- **CLIENT ONLY?** Yes. **SERVER REQUIRED?** No. **SAVE/WORLD DATA?** Local config/log analysis/history only. **SAFE TO REMOVE?** High. **UPDATE FRIENDLINESS:** Excellent. **REGISTRIES/WORLDGEN/CONTENT?** None.
- Project: https://www.curseforge.com/minecraft/mc-mods/crash-assistant

## Specialty visual recheck

- **#31 Animated Resource Pack — NO MATERIAL LEADER CHANGE.** Actions & Stuff 1.11 remains overall/Bedrock; Fresh Animations remains Java. Use Fresh Animations 1.10.4 for Java 1.20.x; 1.10.5 targets 26.1.x/26.2 and intentionally removed pre-26.1 support. EMF + ETF remain recommended.
- **#32 CIT Pack — PACK LEADERS UNCHANGED.** Kaydicraft CIT 1.45 remains the classic 1.20.1 single-pack leader; FAYE remains the active classic-CIT suite; ItemBound: ReBound remains in the modern item-model-definition track. Cozy CIT remains a world-affecting utility watch, not a replacement pack.
- **#33 3D Resource Pack — LEADER UNCHANGED, RUNNER DEPTH IMPROVED.** 3D Default 1.15.0 remains overall and has an exact `mc1.20-1.21.1` file. Nautilus 3D is the strongest newly discovered feature-rich challenger. Better 3D Blocks remains the high-detail runner-up but still carries its Distant Horizons performance warning. Barely Default joins as a modern hybrid 3D/CEM alternate.

## Bedrock / Marketplace / Planet Minecraft late-cycle sweep

No new Bedrock add-on, Marketplace item, Bedrock resource/behavior pack or Planet Minecraft map/build surfaced with enough verified quality and freshness to justify a leader replacement. Sweet Dreams did receive the Bug Fixes 27 package update, but still declares Bedrock 26.30 and supplies no changelog. #16 Canopy, #17 Newb X Supplementary, #18 Auto Factory and #19/#20 Trident Cliffs City therefore remain leaders.

## 33-category scoreboard

1. Best Overall Java Mod — Create: Storage 1.2.7 — NO MATERIAL CHANGE
2. Best Forge/NeoForge Mod — Create: Storage 1.2.7 / Easy NPC modern — NO MATERIAL LEADER CHANGE
3. Best Fabric/Quilt Mod — Easy NPC 7.9.0 — NO MATERIAL CHANGE
4. Best Performance/Optimization Mod — **Legendary Block Entities 0.11.0** — Just Enough Threads specialist runner-up
5. Best Compatibility/Fix Mod — **Collections Of Optimizations 2.4** — Fusion 1.3.14a modern / 1.3.14 Forge 1.20.1 runner-up
6. Best Gameplay/QoL Mod — TargetsIndicate 1.2.0 — NO MATERIAL CHANGE
7. Best Exploration/Worldgen/Dimension Mod — **Jaden's Nether Expansion 2.4.1** — NO MATERIAL LEADER CHANGE
8. Best Mob/Creature Mod — **Psychopath 1.4.0** — **NEW RUNNER-UP: Sons Of Sins 2.2.1**; Aquamirae 7.2.1 retained
9. Best Combat/Animation Mod or Compatibility Layer — **Epic Fight Guard Fix 1.0.9** — **NEW RUNNER-UP: Better Selective Combat 1.0.0**
10. Best Movement/Skating/Grinding/Vehicle Mod — I Wanna Skate 1.2.0 — NO MATERIAL CHANGE
11. Best Create/Technology Mod or Addon — Create: Storage 1.2.7 — NO MATERIAL CHANGE
12. Best Building/Decoration Mod — Create: Copycats+ 3.0.8 — NO MATERIAL CHANGE
13. Best Resource Pack — **Optimum Realism 64x** — **NEW RUNNERS: Barely Default V20.21 / Nautilus 3D**
14. Best Shader Pack / Shader-Adjacent Visual Tech — Complementary Reimagined — NO MATERIAL LEADER CHANGE; PathMax remains experimental runner-up
15. Best Datapack — NO MATERIAL CHANGE
16. Best Bedrock Add-On — **Canopy 1.6.1** — Sweet Dreams **Bug Fixes 27** watch update
17. Best Bedrock Resource/Behavior Pack — Newb X Supplementary 8.0 — NO MATERIAL CHANGE
18. Best Bedrock Marketplace Find — Auto Factory — NO MATERIAL CHANGE
19. Best Minecraft Map/Adventure Map — Trident Cliffs City — NO MATERIAL CHANGE
20. Best Artistic/Showcase Build or Map — Trident Cliffs City — NO MATERIAL CHANGE
21. Best Tool/Utility/Launcher/Modpack Management — **Modrinth App / Prism Launcher** — **NEW DIAGNOSTICS RUNNER: Crash Assistant 1.11.12**
22. Best Server/Admin/Diagnostics Tool — spark 1.10.173 — BetterAutoSave 0.20.1 specialist retained
23. Best Modding/Development Tool or Library — **Easy Model Entities 2.3.0** — Fusion 1.3.14a/1.3.14 fresh infrastructure runner-up
24. Best Experimental/Bleeding-Edge Project — Retromod 1.3.0-snapshot.8 — NO MATERIAL CHANGE
25. Best Future-Version Project Worth Tracking — Retromod 1.3.0-snapshot.8 — NO MATERIAL CHANGE
26. Best Client-Side-Only Mod — **TargetsIndicate 1.2.0** — **NEW SAFE RUNNERS: Particle Rain / Crash Assistant**; Just Enough Threads + Outline N' Stuff retained
27. Best Update-Friendly / World-Safe Mod — TargetsIndicate / BetterF3 / Client Tweaks / Legendary Block Entities — Particle Rain + Crash Assistant added as safe runners
28. Best Client-Side Performance Mod — Legendary Block Entities / ImmediatelyFast — Just Enough Threads specialist runner-up
29. Best Client-Side QoL/UI Mod — **TargetsIndicate 1.2.0** — **NEW ATMOSPHERE RUNNER: Particle Rain**
30. Best Removable-With-Minimal-Save-Risk Mod — TargetsIndicate / BetterF3 — NO MATERIAL LEADER CHANGE
31. Best Animated Resource Pack — Actions & Stuff 1.11 overall; Fresh Animations Java — NO MATERIAL LEADER CHANGE
32. Best CIT Pack — Kaydicraft CIT 1.45 / FAYE suite — NO MATERIAL LEADER CHANGE; Cozy CIT utility watch retained
33. Best 3D Resource Pack — **3D Default 1.15.0** — **NEW SERIOUS RUNNER: Nautilus 3D**; Better 3D Blocks / Barely Default / Items 3D alternates

## Evergreen client stack

**Leader stack remains:** Legendary Block Entities + ImmediatelyFast + Client Tweaks + BetterF3 + Mouse Tweaks + TargetsIndicate, with shaders/resource packs kept as removable presentation layers.

**New safe additions worth testing:** Particle Rain is a save-neutral atmospheric layer with a direct Forge 1.20.1 beta.10 path; Crash Assistant is a client-only diagnostics layer with a direct Forge 1.20.1 build. Just Enough Threads remains the targeted JEI startup/index specialist.

## Persistence delta

- Canonical Google Sheet audited and updated 2026-08-24.
- Permanent categories: **33**.
- Retained serious projects/finds: **87**.
- Six new durable project rows this cycle: Better Selective Combat, Sons Of Sins, Particle Rain, Nautilus 3D, Barely Default, Crash Assistant.
- Fusion and Sweet Dreams were updated in place with their same-day version state.
- Dedicated 3D Packs and Client-Safe views were updated.
- Canonical readable Google Doc and dated Drive report updated in place.
- No historical entries deleted; prior deltas remain preserved.
