# Minecraft Ecosystem Category Scout — 2026-08-24

**Run type:** material update  
**Coverage:** Java + Bedrock • 33 permanent categories • client/world safety • Animated/CIT/3D visual tracks

## Material changes

### #7 Best Exploration / Worldgen / Dimension — NEW LEADER

**Jaden's Nether Expansion 2.4.1** is the new modern leader.

- Java 1.21.1, NeoForge, Release, updated 2026-08-22.
- Required on the 2.4.1 line: Lodestone + Elysium API.
- Forge 1.20.1 remains on legacy **2.3.5** and receives bug fixes only; the maintained forward line is 1.21.1 NeoForge.
- Adds worldgen, biomes, mobs, blocks/items and progression, so treat it as world-affecting and do not casually remove it from an established save.
- Project: https://modrinth.com/mod/jadens-nether-expansion
- Exact 2.4.1: https://modrinth.com/mod/jadens-nether-expansion/version/2.4.1

**Aquamirae 7.2.1** joins #7/#8 as a strong runner-up.

- 1.21.1 stable: NeoForge plus Fabric/Quilt line.
- Forge/Fabric 1.20.1 current active line: **7.1.13 Beta**, updated 2026-08-18.
- Conservative Forge 1.20.1 stable line: **6.4.0**.
- World-affecting exploration/mob/structure content; back up before adoption/removal.
- Project: https://www.curseforge.com/minecraft/mc-mods/aquamirae
- Exact Forge 1.20.1 beta: https://www.curseforge.com/minecraft/mc-mods/aquamirae/files/8678729

### #22 Best Server / Admin / Diagnostics — CATEGORY FILLED

**spark 1.10.173** becomes the mature overall leader.

- Cross-loader profiler for clients, servers and proxies.
- Current 1.10.173 line ships on Forge/Fabric/NeoForge 26.2; the project keeps broad older-version coverage.
- Project: https://modrinth.com/mod/spark
- CurseForge: https://www.curseforge.com/minecraft/mc-mods/spark

**Shinoyuki-BetterAutoSave 0.20.1** is the fresh Forge 1.20.1 / NeoForge 1.21.1 specialist runner-up.

- Server-side only; clients do not need it.
- Forge 1.20.1 requires Forge 47.3.22+ and Java 17; use the `-all` jar.
- NeoForge 1.21.1 requires NeoForge 21.1 and Java 21.
- Moves save serialization/I/O off the main thread and, since 0.20.0, records sync chunk-load stalls and long inter-tick gaps.
- **0.20.1** fixes the 0.20.0 startup crash with known Lithium-family ports by gating the sync-load detector around conflicting overwrites; saving and async loading remain available.
- Do not combine with overlapping async-save systems such as Smooth Chunk Save / Fast Async World Save or conflicting C2ME save-side functionality.
- Safety: touches the world-save pipeline but adds no gameplay registries/worldgen/items/blocks/entities. Back up before first test. Async loading is opt-in/off by default.
- Project: https://modrinth.com/mod/shinoyuki-betterautosave
- 0.20.1 notes: https://github.com/ShinoyukiMiyako/Shinoyuki-BetterAutoSave/blob/main/release-notes/v0.20.1.en.md

### #16 Bedrock Add-On — NEW WATCH

**Canopy 1.6.1 remains overall.** **Sweet Dreams — Bug Fixes 26.mcaddon** is now tracked as a strong horror/survival watch candidate.

- Current surfaced file: Bedrock 26.30, updated 2026-08-23.
- Behavior/content add-on, therefore world-affecting; validate on the installed Bedrock build and back up before using on a long-lived world.
- Project: https://www.curseforge.com/minecraft-bedrock/addons/sweet-dreams

## Specialty visual recheck

- **#31 Animated Resource Pack — NO MATERIAL LEADER CHANGE.** Actions & Stuff 1.11 remains overall/Bedrock; Fresh Animations remains Java. Use Fresh Animations 1.10.4 for Java 1.20.x; 1.10.5 targets 26.1.x/26.2 and intentionally removed pre-26.1 support. EMF + ETF remain recommended.
  - https://modrinth.com/resourcepack/fresh-animations/version/1.10.4
  - https://modrinth.com/resourcepack/fresh-animations/version/1.10.5
- **#32 CIT Pack — NO MATERIAL LEADER CHANGE.** Kaydicraft CIT 1.45 remains the classic 1.20.1 single-pack leader; FAYE remains the active classic-CIT suite; ItemBound: ReBound remains in the modern item-model-definition track.
- **#33 3D Resource Pack — NO MATERIAL LEADER CHANGE.** 3D Default 1.15.0 remains overall and has an exact `mc1.20-1.21.1` file. Better 3D Blocks remains the high-detail runner-up, but its maintainer still warns of a Distant Horizons interaction that can lose up to roughly 80 FPS. Its teased 2026-09-01 update is not treated as released.
  - https://modrinth.com/resourcepack/3d-default/version/1.15.0
  - https://modrinth.com/resourcepack/better-3d-blocks

## Planet Minecraft / map sweep

No sufficiently strong fresh replacement was verified, so #19 and #20 remain **Trident Cliffs City** rather than forcing leaderboard churn.

## 33-category scoreboard

1. Best Overall Java Mod — Create: Storage 1.2.7 — NO MATERIAL CHANGE
2. Best Forge/NeoForge Mod — Create: Storage 1.2.7 / Easy NPC modern — NO MATERIAL CHANGE
3. Best Fabric/Quilt Mod — Easy NPC 7.9.0 — NO MATERIAL CHANGE
4. Best Performance/Optimization Mod — Legendary Block Entities 0.11.0 — NO MATERIAL CHANGE
5. Best Compatibility/Fix Mod — Collections Of Optimizations 2.4 — NO MATERIAL CHANGE
6. Best Gameplay/QoL Mod — TargetsIndicate 1.2.0 — NO MATERIAL CHANGE
7. Best Exploration/Worldgen/Dimension Mod — **Jaden's Nether Expansion 2.4.1 — NEW LEADER**
8. Best Mob/Creature Mod — Psychopath 1.4.0 — **NEW RUNNER-UP: Aquamirae 7.2.1**
9. Best Combat/Animation Mod or Compatibility Layer — Epic Fight Guard Fix 1.0.9 — NO MATERIAL CHANGE
10. Best Movement/Skating/Grinding/Vehicle Mod — I Wanna Skate 1.2.0 — NO MATERIAL CHANGE
11. Best Create/Technology Mod or Addon — Create: Storage 1.2.7 — NO MATERIAL CHANGE
12. Best Building/Decoration Mod — Create: Copycats+ 3.0.8 — NO MATERIAL CHANGE
13. Best Resource Pack — Optimum Realism 64x — NO MATERIAL CHANGE
14. Best Shader Pack / Shader-Adjacent Visual Tech — Complementary Reimagined — NO MATERIAL CHANGE; PathMax stays experimental runner-up
15. Best Datapack — NO MATERIAL CHANGE
16. Best Bedrock Add-On — Canopy 1.6.1 — **NEW WATCH: Sweet Dreams Bug Fixes 26**
17. Best Bedrock Resource/Behavior Pack — Newb X Supplementary 8.0 — NO MATERIAL CHANGE
18. Best Bedrock Marketplace Find — Auto Factory — NO MATERIAL CHANGE
19. Best Minecraft Map/Adventure Map — Trident Cliffs City — NO MATERIAL CHANGE
20. Best Artistic/Showcase Build or Map — Trident Cliffs City — NO MATERIAL CHANGE
21. Best Tool/Utility/Launcher/Modpack Management — Modrinth App / Prism Launcher — NO MATERIAL CHANGE
22. Best Server/Admin/Diagnostics Tool — **spark 1.10.173 — NEW CATEGORY LEADER; BetterAutoSave 0.20.1 specialist runner-up**
23. Best Modding/Development Tool or Library — Easy Model Entities 2.3.0 — NO MATERIAL CHANGE
24. Best Experimental/Bleeding-Edge Project — Retromod 1.3.0-snapshot.8 — NO MATERIAL CHANGE
25. Best Future-Version Project Worth Tracking — Retromod 1.3.0-snapshot.8 — NO MATERIAL CHANGE
26. Best Client-Side-Only Mod — TargetsIndicate 1.2.0 — NO MATERIAL CHANGE
27. Best Update-Friendly / World-Safe Mod — TargetsIndicate / BetterF3 / Client Tweaks / Legendary Block Entities — NO MATERIAL CHANGE
28. Best Client-Side Performance Mod — Legendary Block Entities / ImmediatelyFast — NO MATERIAL CHANGE
29. Best Client-Side QoL/UI Mod — TargetsIndicate 1.2.0 — NO MATERIAL CHANGE
30. Best Removable-With-Minimal-Save-Risk Mod — TargetsIndicate / BetterF3 — NO MATERIAL CHANGE
31. Best Animated Resource Pack — Actions & Stuff 1.11 overall; Fresh Animations Java — NO MATERIAL LEADER CHANGE
32. Best CIT Pack — Kaydicraft CIT 1.45 / FAYE suite — NO MATERIAL LEADER CHANGE
33. Best 3D Resource Pack — 3D Default 1.15.0 — NO MATERIAL LEADER CHANGE; Better 3D Blocks remains DH-caution runner-up

## Evergreen client stack

**NO MATERIAL CHANGE:** Legendary Block Entities + ImmediatelyFast + Client Tweaks + BetterF3 + Mouse Tweaks + TargetsIndicate, with shaders/resource packs kept as removable presentation layers. BetterAutoSave is deliberately excluded because it is server-side and modifies the save pipeline.

## Persistence delta

- Canonical Google Sheet audited 2026-08-24.
- Permanent categories: **33**.
- Retained serious projects/finds: **78**.
- Four durable catalog additions this run: Jaden's Nether Expansion, Aquamirae, Shinoyuki-BetterAutoSave, Sweet Dreams.
- Canonical readable Google Doc updated in place.
- No historical entries deleted; prior deltas remain in earlier dated checkpoints.
