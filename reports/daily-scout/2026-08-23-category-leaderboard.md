# Minecraft Daily Category Scout — 2026-08-23

**Run:** daily category leaderboard  
**Scope:** Java + Bedrock, active versions/loaders, current Forge 1.20.1 fit, client-only/update-friendly safety, resource packs, shaders, tools, Bedrock, Marketplace, and Planet Minecraft.  
**Baseline rechecked:** Java stable 26.2; Bedrock stable 26.45 hotfix line; Bedrock Preview/Beta 26.50.26.

## Material changes since the 2026-08-22 report

### 1) Create: Storage 1.2.7 — strongest same-day Forge 1.20.1 update

**Status:** TEST / STRONG OPTIONAL  
**Minecraft:** 1.20.1 Forge  
**Released:** 2026-08-23  
**Environment:** client + server

This is the cleanest mature same-day update for the current Create-heavy 1.20.1 ecosystem. Version 1.2.7 adds Jade upgrade tooltips, clearer Void mode indication, Smart Observer detection for item transfers through Passer blocks, a Smart Passer upgrade recipe, cheaper Jetpack upgrades, and moves Vanilla Backport compatibility into a datapack to stop log errors when that mod is absent. It also fixes an `AbstractMethodError` crash with Construction Wand - KOTS and a free-block-placement bug involving unreadable Storage Boxes.

**World/update safety:** medium. It adds persistent blocks/items/storage content, so it is not in the update-neutral client bucket. Back up before removing from a live world.

- https://www.curseforge.com/minecraft/mc-mods/create-storage-neo-forge
- https://www.curseforge.com/minecraft/mc-mods/create-storage-neo-forge/files/8712044

### 2) TargetsIndicate 1.2.0 — best new client-only Forge 1.20.1 QoL find

**Status:** CLIENT-SAFE / TEST NOW  
**Minecraft:** 1.20.1 Forge; 1.21.1 NeoForge  
**Released:** 2026-08-23  
**Environment:** client only

A lightweight animated combat HUD that tracks recently damaged entities with health bars, entity icons and configurable layouts. Version 1.2.0 adds line-of-sight hover previews for untracked targets and fixes bow attack detection.

**CLIENT ONLY?** Yes.  
**SERVER REQUIRED?** No.  
**SAVE/WORLD DATA TOUCHED?** None expected beyond client config.  
**ADDS REGISTRIES/WORLDGEN/ENTITIES/ITEMS/BLOCKS?** No.  
**SAFE TO REMOVE?** High confidence — removing it removes only the HUD.  
**UPDATE FRIENDLINESS:** Excellent. A newer Minecraft version can normally be tested by replacing/removing the mod without migrating the world.

- https://www.curseforge.com/minecraft/mc-mods/targetsindicate
- https://www.curseforge.com/minecraft/mc-mods/targetsindicate/files/8711568

### 3) Psychopath 1.4.0 — best substantial same-day mob/horror update

**Status:** TEST BRANCH / OPTIONAL  
**Minecraft:** 1.20.1 Forge; 1.21.1 NeoForge; 1.19.2 Forge  
**Released:** 2026-08-23

The project has roughly 260K downloads and 1.4.0 is a real content/engineering update rather than a metadata refresh: the author says the codebase was optimized, multiplayer bugs were fixed, AI was tweaked, config was added, a new chase theme and lore structures were added, and the stalker can now throw knives and crawl through 1x1 gaps. The author also explicitly warns that the 1.20.1 and 1.21.1 ports may still have bugs.

**World/update safety:** low. It adds mob/structure/gameplay state and should be treated as a committed world-content mod, not an evergreen client utility.

- https://www.curseforge.com/minecraft/mc-mods/psychopath
- https://www.curseforge.com/minecraft/mc-mods/psychopath/files/8712347

### 4) Easy Model Entities 2.3.0 — best same-day creator/modding-tool update

**Status:** TOOL / FUTURE-HIGH-VALUE  
**Minecraft:** 26.2 Forge + NeoForge + Fabric/Quilt; project also maintains 1.20.1 and 1.21.x lines  
**Released:** 2026-08-23

This project turns Blockbench models into simple Minecraft entities without requiring Java code. The 2.3.0 line landed today for 26.2 across Forge, NeoForge and Fabric/Quilt. The project also has a native Forge 1.20.1 line, though today's 2.3.0 update is for modern versions.

This is especially useful as reference tooling for custom mob/NPC-heavy projects because it keeps a broad loader matrix and focuses on creator workflow rather than a single pack.

- https://www.curseforge.com/minecraft/mc-mods/easy-model-entities

### 5) Easy NPC 7.9.0 — major mature multi-loader creator ecosystem update

**Status:** TOOL / FUTURE VERSION  
**Minecraft:** 26.2 Forge + NeoForge + Fabric/Quilt  
**Released:** 2026-08-23

Easy NPC is a mature multi-million-download NPC/dialogue framework and today shipped 7.9.0 for 26.2 across the major loaders. This is a better future-version NPC foundation signal than most tiny same-day NPC uploads because it has years of project history and a large install base.

- https://www.curseforge.com/minecraft/mc-mods/easy-npc

### 6) Bedrock visuals: GlowCraft 2.6 and Newb X Supplementary 8.0

**Status:** BEDROCK VISUAL / TEST ON 26.45  
**Declared target:** 26.40  
**Released:** 2026-08-23

**GlowCraft 2.6** is a Vibrant Visuals pack focused on emissive/glowing blocks, items and entities and has roughly 95K project downloads. **Newb X Supplementary 8.0** also released today, with roughly 119K project downloads and a merged 23.8 MB pack. Both are meaningful same-day visual updates, but their current project files declare Bedrock 26.40 while stable Bedrock is 26.45, so treat them as test-on-26.45 rather than guaranteed 26.45 declarations.

As resource/visual packs they are much more update-friendly than behavior-heavy add-ons, though Bedrock rendering-version changes can still break visuals.

- https://www.curseforge.com/minecraft-bedrock/texture-packs/glowcraft
- https://www.curseforge.com/minecraft-bedrock/texture-packs/newb-x-supplementary

## 30-category scoreboard

| # | Category | 2026-08-23 result | Change vs prior run |
|---|---|---|---|
| 1 | Best Overall Java Mod | **Create: Storage 1.2.7** as today's strongest practical mature update | NEW TODAY |
| 2 | Best Forge/NeoForge Mod | **Create: Storage 1.2.7** for Forge 1.20.1; **Easy NPC 7.9.0** for modern multi-loader | NEW TODAY |
| 3 | Best Fabric/Quilt Mod | **Easy NPC 7.9.0**; runner-up **Simple Orbital Strikes 1.0.8** | NEW TODAY |
| 4 | Best Performance/Optimization Mod | **Legendary Block Entities 0.11.0** remains the 1.20.1 client pick; **Async Advancement Toast 1.0.0** is today's experimental watch | NO LEADER CHANGE |
| 5 | Best Compatibility/Fix Mod | **Collections Of Optimizations 2.4** remains broad winner; Create: Storage 1.2.7 adds useful Construction Wand/Vanilla Backport fixes | NO LEADER CHANGE |
| 6 | Best Gameplay/QoL Mod | **TargetsIndicate 1.2.0** | NEW TODAY |
| 7 | Best Exploration/Worldgen/Dimension Mod | **Candyverse** is today's future-version dimension watch; no better current-pack leader surfaced | WATCH / NO 1.20.1 LEADER CHANGE |
| 8 | Best Mob/Creature Mod | **Psychopath 1.4.0** | NEW TODAY |
| 9 | Best Combat/Animation Mod or Compat | **No material leader change.** TargetsIndicate improves combat UI; Deus Chrono-Machina 6.10 is a niche AlienEvo combat update | NO MATERIAL CHANGE |
| 10 | Best Movement/Skating/Grinding/Vehicle Mod | **No material change**; continue tracking native 1.20.1 I Wanna Skate and newer-version skating ports | NO MATERIAL CHANGE |
| 11 | Best Create/Technology Mod or Addon | **Create: Storage 1.2.7** | NEW TODAY |
| 12 | Best Building/Decoration Mod | **Create: Copycats+ 3.0.8** remains current overall; a 1.21.1 NeoForge file landed today while the 1.20.1 3.0.8 build remains Aug 20 | FUTURE COVERAGE UPDATE |
| 13 | Best Resource Pack | **FeelTrue x32 PBR** today's fresh realism pick; **Optimum Realism 64x** remains stronger mature overall realism baseline | FRESH PICK ONLY |
| 14 | Best Shader Pack | Java overall: **Complementary Reimagined** unchanged. Bedrock same-day visual leaders: **Newb X Supplementary 8.0 / GlowCraft 2.6** | JAVA NO CHANGE / BEDROCK UPDATE |
| 15 | Best Datapack | **No material change** | NO MATERIAL CHANGE |
| 16 | Best Bedrock Add-On | **Canopy** today's strongest established technical/QoL update; **CobbleDrock 1.2.3** remains the larger gameplay-content pick | NEW DAILY PICK |
| 17 | Best Bedrock Resource/Behavior Pack | **Newb X Supplementary 8.0**, runner-up **GlowCraft 2.6** | NEW TODAY |
| 18 | Best Bedrock Marketplace Find | **Auto Factory** remains current August utility pick | NO MATERIAL CHANGE |
| 19 | Best Minecraft Map/Adventure Map | **No defensible new winner**; Planet Minecraft's indexed latest-map page still lags behind Aug 23 | NO MATERIAL CHANGE |
| 20 | Best Artistic/Showcase Build or Map | **Trident Cliffs City** remains last verified standout; no same-day PMC winner claimed | NO MATERIAL CHANGE |
| 21 | Best Tool/Utility/Launcher | **Modrinth App** integrated workflow; **Prism Launcher 11.0.3** power-user baseline | NO MATERIAL CHANGE |
| 22 | Best Server/Admin/Diagnostics Tool | **No material change** | NO MATERIAL CHANGE |
| 23 | Best Modding/Development Tool or Library | **Easy Model Entities 2.3.0**; runner-up **Easy NPC 7.9.0** | NEW TODAY |
| 24 | Best Experimental/Bleeding-Edge Project | **Async Advancement Toast 1.0.0** — interesting client-thread optimization idea, but only a few downloads so WATCH only | NEW WATCH |
| 25 | Best Future-Version Project Worth Tracking | **Easy Model Entities 2.3.0 / Easy NPC 7.9.0** modern 26.2 multi-loader creator stack | NEW TODAY |
| 26 | Best Client-Side-Only Mod | **TargetsIndicate 1.2.0** today's fresh pick; **Legendary Block Entities 0.11.0** remains performance-focused overall pick | NEW TODAY |
| 27 | Best Update-Friendly / World-Safe Mod | **TargetsIndicate 1.2.0** fresh; **BetterF3 / Client Tweaks / LBE** remain evergreen stack members | NEW TODAY |
| 28 | Best Client-Side Performance Mod | **Legendary Block Entities 0.11.0** remains winner; **Async Advancement Toast** watch only | NO LEADER CHANGE |
| 29 | Best Client-Side QoL/UI Mod | **TargetsIndicate 1.2.0** | NEW TODAY |
| 30 | Best Removable-With-Minimal-Save-Risk Mod | **TargetsIndicate 1.2.0** fresh; **BetterF3** remains ultra-safe evergreen | NEW TODAY |

## Client-side / update-friendly safety track

### TargetsIndicate 1.2.0
- Client only: **Yes**
- Server required: **No**
- Save/world data touched: **none expected; client config only**
- Adds registries/worldgen/entities/items/blocks: **No**
- Safe to remove: **High confidence**
- Existing-world safety: **High**
- Update friendliness: **Excellent**
- Upgrade behavior: normally remove/replace the mod for a new Minecraft version; world migration should not be required by this mod.

### Legendary Block Entities 0.11.0
- Client only: **Yes**
- Server required: **No**
- Save/world data touched: **none expected; rendering/config only**
- Adds registries/worldgen/entities/items/blocks: **No**
- Safe to remove: **High confidence**
- Existing-world safety: **High**
- Update friendliness: **Excellent**, but this particular release is a Forge 1.20.1 build.

### Async Advancement Toast 1.0.0
- Client only: **Yes**
- Server required: **No**
- Save/world data touched: **none expected**
- Adds registries/worldgen/entities/items/blocks: **No**
- Safe to remove: **High by design, but low field confidence because it is brand-new**
- Existing-world safety: **Likely high**
- Update friendliness: **Excellent in concept**
- Recommendation: **WATCH**, not install-first, until it has more field use because it moves advancement-toast preparation off the main render thread.

### Hold Tab Minimap for JourneyMap 1.1.2 (1.20.1)
- Client only: **Yes**
- Server required: **No**
- Save/world data touched: **client config only**
- Adds content/worldgen: **No**
- Safe to remove: **High confidence**
- Update friendliness: **Excellent**, but it depends on JourneyMap and therefore is not relevant to a Xaero-only setup.

## Best evergreen client stack

For maximum QoL/performance/visual improvement while keeping the world easy to upgrade, the strongest current layer remains:

- **Legendary Block Entities** — client rendering optimization; no world dependency.
- **ImmediatelyFast** — client rendering optimization; broad modern loader/version coverage.
- **Client Tweaks** — client QoL with low world coupling.
- **BetterF3** — client debug-HUD replacement; extremely easy to remove.
- **Mouse Tweaks** — client inventory-input QoL; extremely easy to remove.
- **TargetsIndicate** — new Aug 23 combat HUD candidate; client-only and world-neutral.
- **Resource packs / shaders** — generally world-neutral presentation layer; swap/remove for version upgrades as renderer support changes.

Keep **Async Advancement Toast** out of the evergreen stack for now: its design is exactly the right kind of client-only optimization, but it is too new to elevate above proven options yet.

## Bedrock notes

Current stable Bedrock remains **26.45**, while many Aug 23 CurseForge files still declare **26.40**. Continue treating those as **test candidates on 26.45**, not guaranteed compatibility declarations.

Strong Aug 23 Bedrock updates discovered:
- **Canopy** — established technical/QoL add-on, ~82K downloads in today's index.
- **Newb X Supplementary 8.0** — major visual pack update.
- **GlowCraft 2.6** — Vibrant Visuals emissive/glow update.
- **Simple Auto Totem** and **Block & Entity Details** — established utility-style add-ons updated today.
- **Bedrock Cinematic Studio** — interesting creator-tool direction, but still much smaller/younger than established leaders.

## Planet Minecraft artistic sweep

Planet Minecraft was deliberately checked again. Its indexed latest-map page available to this run still only surfaced items through Aug 19, so **no Aug 23 artistic/map winner is claimed**. The previous verified standout, **Trident Cliffs City**, remains the current placeholder until PMC's index catches up or a fresher project can be verified directly.

## Tool / launcher sweep

No material tool leaderboard change today:
- **Modrinth App** remains the strongest integrated discovery/update workflow after the Aug 18 Play-page/instance-management overhaul.
- **Prism Launcher 11.0.3** remains the current stable power-user launcher baseline; its official news page still lists 11.0.3 as latest stable.

## Current-pack priority order after this run

1. **TargetsIndicate 1.2.0** — easiest same-day client-only QoL test with essentially no world risk.
2. **Create: Storage 1.2.7** — strong if the pack already wants/uses its storage system; today fixes real compatibility/crash issues.
3. **Keep Legendary Block Entities 0.11.0 at the top of the client-performance test list.** No stronger proven same-day replacement surfaced.
4. **Collections Of Optimizations 2.4** remains the broad optimization audit target from the prior run.
5. **Psychopath 1.4.0** is a fun/content test only; do not treat it as update-neutral and note the author's warning that ports may still be buggy.
6. **Do not promote Async Advancement Toast yet.** Track it until it has enough real-world usage/source evidence.
7. **Future-version creator stack:** watch Easy Model Entities 2.3.0 + Easy NPC 7.9.0 on 26.2.
8. **Bedrock visuals:** test Newb X Supplementary 8.0 / GlowCraft 2.6 on 26.45 because their project files declare 26.40.

## Source hubs checked

- Minecraft Java 26.2: https://www.minecraft.net/en-us/article/minecraft-java-edition-26-2
- Bedrock 26.45 hotfix: https://feedback.minecraft.net/hc/en-us/articles/48149564061965-Minecraft-Bedrock-Edition-26-44-45-Hotfix-Changelog
- Bedrock Preview/Beta 26.50.26: https://feedback.minecraft.net/hc/en-us/articles/48228443785101-Minecraft-Beta-Preview-26-50-26
- CurseForge Java latest listings and exact project/file pages
- CurseForge Bedrock add-on/resource-pack latest listings and exact project/file pages
- Modrinth news/changelog
- Prism Launcher official news
- Planet Minecraft latest maps index

This report becomes the delta watermark for the next daily run.