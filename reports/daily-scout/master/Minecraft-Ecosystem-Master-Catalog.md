# Minecraft Ecosystem Master Discovery Catalog

**Last audited:** 2026-08-24  
**Permanent categories:** 33  
**Tracked serious projects/finds in canonical Google Sheet:** 78

This GitHub mirror is the durable category/ranking index for the recurring Minecraft ecosystem scouting service. The canonical sortable registry is maintained in the paired Google Sheet and contains exact versions, supported Minecraft versions, loader/channel, release date, status, dependencies, direct project/version/changelog links, client/server side, save impact, removability/update-friendliness, first/last tracked dates, compatibility notes, and ranking history.

## Latest material delta — 2026-08-24

- **#7 Exploration/Worldgen/Dimension — NEW LEADER:** Jaden's Nether Expansion 2.4.1. Active modern line is NeoForge 1.21.1; Forge 1.20.1 remains on legacy 2.3.5 and receives bug fixes only. World-affecting; use backups/test worlds. https://modrinth.com/mod/jadens-nether-expansion/version/2.4.1
- **#7/#8 — NEW RUNNER-UP:** Aquamirae 7.2.1. Stable modern 1.21.1 line; Forge/Fabric 1.20.1 has 7.1.13 Beta, while 6.4.0 is the older stable 1.20.1 line. World-affecting. https://www.curseforge.com/minecraft/mc-mods/aquamirae
- **#22 Server/Admin/Diagnostics — CATEGORY FILLED:** spark 1.10.173 becomes the mature overall profiler leader. https://modrinth.com/mod/spark
- **#22 specialist / #4 watch:** Shinoyuki-BetterAutoSave 0.20.1 is tracked for Forge 1.20.1 / NeoForge 1.21.1 autosave stalls and diagnostics. It is server-side, pre-1.0, touches the save pipeline, and must not be stacked with overlapping async-save systems. Version 0.20.1 fixes the 0.20.0 startup crash with known Lithium-family ports by gating only the new sync-load detector. https://modrinth.com/mod/shinoyuki-betterautosave
- **#16 Bedrock Add-On — NEW WATCH:** Sweet Dreams Bug Fixes 26. Canopy 1.6.1 remains overall; Sweet Dreams is a fast-moving horror/survival runner-up whose surfaced current file declares Bedrock 26.30. World-affecting; validate version and back up. https://www.curseforge.com/minecraft-bedrock/addons/sweet-dreams
- **#31 Animated / #32 CIT / #33 3D — NO MATERIAL LEADER CHANGE.** Fresh Animations 1.10.4 remains the 1.20.x Java path; 1.10.5 is for 26.1.x/26.2. 3D Default 1.15.0 keeps an exact mc1.20–1.21.1 file. Better 3D Blocks still carries the maintainer's Distant Horizons LOD/FPS warning; its teased 2026-09-01 update is not treated as released.
- **Planet Minecraft #19/#20 sweep — NO MATERIAL CHANGE.** No fresh project had enough verified quality + usability evidence to justify replacing Trident Cliffs City.

Full checkpoint: `reports/daily-scout/2026-08-24-category-leaderboard.md`

## Prior retained late-cycle delta — 2026-08-23

- **#24 Experimental/Bleeding-Edge / #25 Future-Version:** Retromod 1.3.0-snapshot.8 remains the current technical experiment/migration leader; beta, cloned-instance testing only.
- **#26 Client-Side / #29 QoL-UI:** Outline N' Stuff remains a future runner-up for Fabric/NeoForge 1.21.1 + 26.2; no Forge 1.20.1 build.
- **#14 Shader:** PathMax remains the Iris-only experimental path-tracing runner-up; Complementary Reimagined stays mature overall.
- **#17 Bedrock visuals:** Cinematic Visuals 1.2 and Unbound Visuals 2.3.1 remain tracked runners-up.

Full prior checkpoint: `reports/daily-scout/2026-08-23-late-cycle-delta.md`

## Permanent visual categories

### 31. Best Animated Resource Pack
- **Overall / Bedrock:** Actions & Stuff 1.11.
- **Java:** Fresh Animations 1.10.5 for 26.1.x–26.2; use Fresh Animations 1.10.4 for Java 1.20.x/1.21.x, including 1.20.1.
- Recommended modern Java companions: Entity Model Features + Entity Texture Features.
- Links: https://www.minecraft.net/en-us/article/summer-sale-2026 • https://modrinth.com/resourcepack/fresh-animations

### 32. Best CIT Pack
- **Best single classic 1.20.1 pack:** Kaydicraft CIT 1.45 — OptiFine CIT or CIT Resewn-compatible renderer.
- **Best active classic suite:** FAYE CIT collection.
- **Modern CIT-style successor track:** ItemBound: ReBound / modern item-model-definition systems remain separate from classic CIT.
- Links: https://modrinth.com/resourcepack/kaydicraft-cit • https://www.planetminecraft.com/texture-pack/the-faye-wallpaper-amp-flooring-cit-pack/

### 33. Best 3D Resource Pack
- **Best current overall / safest first test:** 3D Default 1.15.0 — exact `mc1.20-1.21.1` build, lightweight true geometry, vanilla-look textures, no save dependency.
- **Best higher-detail contender:** Better 3D Blocks 2.9.0 — broad project coverage, but maintainer warns of a Distant Horizons LOD interaction that can cause very large FPS loss. Treat as TEST BRANCH with DH.
- **Best modern item-only specialist:** Items 3D 5.0.
- **Best subtle modern vanilla extension:** Vanilla 3D Extension 1.19.
- Links: https://modrinth.com/resourcepack/3d-default/version/1.15.0 • https://modrinth.com/resourcepack/better-3d-blocks • https://www.curseforge.com/minecraft/texture-packs/items-3d • https://www.curseforge.com/minecraft/texture-packs/vanilla-3d-extension

## 33-category scoreboard

1. Best Overall Java Mod — Create: Storage 1.2.7
2. Best Forge/NeoForge Mod — Create: Storage 1.2.7 / Easy NPC modern
3. Best Fabric/Quilt Mod — Easy NPC 7.9.0
4. Best Performance/Optimization Mod — Legendary Block Entities 0.11.0 for Forge 1.20.1; ImmediatelyFast runner-up; Sodium modern line
5. Best Compatibility/Fix Mod — Collections Of Optimizations 2.4; Retromod experimental migration track
6. Best Gameplay/QoL Mod — TargetsIndicate 1.2.0
7. Best Exploration/Worldgen/Dimension Mod — **Jaden's Nether Expansion 2.4.1**; Sculk Horde 0.12.7 Forge 1.20.1 + Aquamirae 7.2.1 runners-up
8. Best Mob/Creature Mod — Psychopath 1.4.0; Aquamirae 7.2.1 runner-up
9. Best Combat/Animation Mod or Compatibility Layer — Epic Fight Guard Fix 1.0.9
10. Best Movement/Skating/Grinding/Vehicle Mod — I Wanna Skate 1.2.0
11. Best Create/Technology Mod or Addon — Create: Storage 1.2.7
12. Best Building/Decoration Mod — Create: Copycats+ 3.0.8
13. Best Resource Pack — Optimum Realism 64x
14. Best Shader Pack / Shader-Adjacent Visual Tech — Complementary Reimagined; PathMax experimental runner-up
15. Best Datapack — NO MATERIAL CHANGE
16. Best Bedrock Add-On — Canopy 1.6.1; Sweet Dreams new watch
17. Best Bedrock Resource/Behavior Pack — Newb X Supplementary 8.0; Cinematic Visuals 1.2 and Unbound Visuals 2.3.1 runners-up
18. Best Bedrock Marketplace Find — Auto Factory
19. Best Minecraft Map/Adventure Map — Trident Cliffs City
20. Best Artistic/Showcase Build or Map — Trident Cliffs City
21. Best Tool/Utility/Launcher/Modpack Management — Modrinth App; Prism Launcher power-user runner-up
22. Best Server/Admin/Diagnostics Tool — **spark 1.10.173**; Shinoyuki-BetterAutoSave 0.20.1 specialist runner-up
23. Best Modding/Development Tool or Library — Easy Model Entities 2.3.0
24. Best Experimental/Bleeding-Edge Project — Retromod 1.3.0-snapshot.8
25. Best Future-Version Project Worth Tracking — Retromod 1.3.0-snapshot.8; Easy Model Entities + Easy NPC runner-up
26. Best Client-Side-Only Mod — TargetsIndicate 1.2.0; Outline N' Stuff future runner-up
27. Best Update-Friendly / World-Safe Mod — TargetsIndicate / BetterF3 / Client Tweaks / Legendary Block Entities
28. Best Client-Side Performance Mod — Legendary Block Entities 0.11.0 / ImmediatelyFast
29. Best Client-Side QoL/UI Mod — TargetsIndicate 1.2.0; Outline N' Stuff future runner-up
30. Best Removable-With-Minimal-Save-Risk Mod — TargetsIndicate 1.2.0 / BetterF3
31. Best Animated Resource Pack — Actions & Stuff 1.11 overall; Fresh Animations Java
32. Best CIT Pack — Kaydicraft CIT 1.45; FAYE suite runner-up
33. Best 3D Resource Pack — 3D Default 1.15.0; Better 3D Blocks 2.9.0 runner-up

## Evergreen client stack

**Legendary Block Entities + ImmediatelyFast + Client Tweaks + BetterF3 + Mouse Tweaks + TargetsIndicate**, with resource packs/shaders kept as removable presentation layers.

BetterAutoSave is deliberately **not** included in this client stack because it is server-side and intentionally modifies the world-save pipeline.

## Persistence rules

- Update the same canonical Drive Doc + Google Sheet in place; do not create a new master tracker every run.
- Maintain dedicated Animated Packs, CIT Packs, and 3D Packs views/tabs.
- Keep exact version/download/changelog links and last-checked dates current.
- Preserve historical/superseded entries rather than silently deleting them.
- A no-change run is a successful no-op; do not invent churn.
- Resource packs are normally save-neutral, but renderer dependencies, Marketplace entitlement, modern item-model data, and load-order conflicts must be documented explicitly.
- Every material run publishes a dated checkpoint to both Drive and GitHub and verifies both copies before completion.
