# Minecraft Daily Category Scout — 2026-08-22

This report covers Java and Bedrock across active versions/loaders, with special attention to Forge 1.20.1, client-side/update-friendly additions, shaders/resource packs, tools, Bedrock, and artistic maps.

## Top recommendations

### INSTALL / TEST NOW — Forge 1.20.1

1. **Legendary Block Entities 0.11.0** — strongest fresh client-side performance find today. Forge 1.20.1 only, 100% client-side, now also optimizes signs in addition to chests, shulker boxes, bells and beds. High confidence world-safe/removable because it changes rendering rather than registries/world data. Test beside Embeddium/Oculus and the existing optimization stack before promoting.
   - https://www.curseforge.com/minecraft/mc-mods/legendary-block-entities
   - https://www.curseforge.com/minecraft/mc-mods/legendary-block-entities/files/8703898

2. **Client Tweaks 11.1.11** — best recent evergreen QoL update for Forge 1.20.1. The project is extremely established, supports Fabric/Forge/NeoForge across many Minecraft versions, and received its 1.20.1 Forge update on Aug 21. This is the kind of low-world-coupling utility that is usually easy to remove when testing a future Minecraft upgrade.
   - https://www.curseforge.com/minecraft/mc-mods/client-tweaks

3. **Epic Fight Guard Fix** — highest relevance combat/compatibility watch for the current Epic Fight-heavy setup. CurseForge’s current listing shows project activity on Aug 22, while the directly surfaced latest file remains `guardfix-1.0.9.jar` for Forge 1.20.1 from Aug 16. It targets shield/weapon guard failures, Weapons of Miracles Perfect Bulwark, Nightfall parry, unarmed combat and other core bugs. Because it uses mixins in an already dense combat stack, TEST BRANCH rather than blind main-pack install.
   - https://www.curseforge.com/minecraft/mc-mods/epic-fight-guard-fix

4. **Collections Of Optimizations** — updated Aug 22 for Forge 1.20.1. This is already part of the known optimization stack, so this is an **audit/update candidate**, not a recommendation to duplicate-install it. Verify the installed build against today’s current file before changing anything.
   - https://www.curseforge.com/minecraft/mc-mods/collections-of-optimizations

5. **MoreInformation 1.0.0** — brand-new configurable information HUD with Forge 1.20.1 and NeoForge 1.21.1 builds. It shows TPS/MSPT/RAM/player/FPS/time and optional integrations. Very low adoption today, so it is a TEST/OPTIONAL pick rather than a default add.
   - https://www.curseforge.com/minecraft/mc-mods/moreinformation

### Gameplay / content

6. **Ars Zero 2.0.2** — substantive Aug 22 release for Forge 1.20.1 and NeoForge 1.21.1, adding cast devices and glyphs for Ars Nouveau. Strong if Ars Nouveau is in the target instance; otherwise skip rather than adding a new magic ecosystem only for the addon.
   - https://www.curseforge.com/minecraft/mc-mods/ars-zero

7. **World Bosses 1.4.0** — today’s notable future-loader expansion: a Forge 26.2 beta migration landed Aug 22. The mod already has Fabric builds for 1.20.1, 1.21.1, 1.21.11, 26.1.2 and 26.2 and provides shrine/ritual-driven world bosses, phases, raid scaling, trophies and endgame gear. Treat it as FUTURE/HIGH-VALUE for the current Forge 1.20.1 pack rather than forcing the Fabric 1.20.1 build through Connector.
   - https://www.curseforge.com/minecraft/mc-mods/world-bosses
   - https://www.curseforge.com/minecraft/mc-mods/world-bosses/files/8706287

## Client-side / update-friendly leaderboard

### Best fresh client-side performance pick
**Legendary Block Entities 0.11.0**
- Client only: **Yes**
- Server required: **No**
- Save/world data touched: **None expected; rendering optimization**
- Adds blocks/items/entities/worldgen: **No**
- Safe to remove: **High confidence**
- Update friendliness: **Excellent for a 1.20.1-only client optimization**

### Best recent evergreen client QoL pick
**Client Tweaks 11.1.11**
- Broad Fabric/Forge/NeoForge and multi-version support
- Very mature project
- Low world/save coupling compared with content/worldgen mods
- Good candidate for a long-lived client profile

### Best evergreen client performance foundation already worth keeping
**ImmediatelyFast** remains a strong lightweight, client-side rendering optimizer across Fabric/Forge/NeoForge/Quilt and 1.20.x through 26.x. If already installed, do not duplicate it; keep it in the evergreen client stack and audit versions rather than replacing it casually.
- https://modrinth.com/mod/immediatelyfast

### Additional world-safe evergreen client utility
**BetterF3** remains a strong cross-loader client-only debug HUD replacement with support from 1.16.x through 26.2.
- https://modrinth.com/mod/betterf3

**Mouse Tweaks** remains a mature client-only inventory input improvement with very broad loader/version coverage. The 1.20.1 Forge build is older, but the project itself is actively maintained for current versions.
- https://www.curseforge.com/minecraft/mc-mods/mouse-tweaks

## Resource pack

### Today’s best fresh resource-pack candidate: N87 PBR Pack 128x — Aug 22 build
The Aug 22 file supports Minecraft 1.20.1 through 26.2 and provides physically based rendering support for vanilla textures. It is still in production, so treat it as a high-end visual test rather than a finished universal replacement. As a resource pack it is world-safe and removable without save migration.
- https://www.curseforge.com/minecraft/texture-packs/n87-pbr-pack
- https://www.curseforge.com/minecraft/texture-packs/n87-pbr-pack/files/8704749

## Shader

### Best current proven lightweight pick: Author87668’s LIGHT Shaders
Not a new Aug 22 file, but still one of the strongest current low-cost visual options surfaced in today’s scan. It supports Iris/Oculus/OptiFine, includes 1.20.1 support, and is designed for strong FPS while retaining water reflections, ripples, shadows, weather and optional TAA.
- https://www.curseforge.com/minecraft/shaders/author87668s-light-shaders

### Today’s experimental shader watch: Revival Shader
CurseForge’s current shader category listing shows Revival Shader as updated Aug 22 and describes additional optimization beyond Mellow Shader, but the latest directly inspectable file metadata in this run still pointed to v0.8 from Aug 16. **WATCH until file metadata settles; do not call it a verified new Aug 22 binary yet.**
- https://www.curseforge.com/minecraft/shaders/revival-shader

## Bedrock — current stable 26.40 ecosystem

### Best fresh Bedrock worldgen add-on: Better Biomes V2
Updated Aug 22 for Bedrock 26.40. Adds new biomes while overhauling vanilla biomes/caves. High visual/gameplay value, but it is intentionally world-affecting: not an update-friendly/removable pick for a forever world.
- https://www.curseforge.com/minecraft-bedrock/addons/better-biomes-addon-mrstrader

### Best fresh Bedrock performance pack: Feather FPS Boost Mod v11
Updated Aug 22 for Bedrock 26.40 and marked achievement-friendly, no experiments required, existing-world compatible and Vibrant Visuals compatible. Important caveat: the behavior-pack component reduces mob spawn density; for the most update-friendly/world-neutral use, prefer its resource-pack-only mode.
- https://www.curseforge.com/minecraft-bedrock/addons/feather-fps-boost-mod

### Best polished action add-on updated today: Spider-Man | EL SANDO
Updated Aug 22 with a redesigned Spider-Man HUD plus bug/stability/performance work. The project has broad Bedrock version coverage and substantially more adoption than most same-day releases.
- https://www.curseforge.com/minecraft-bedrock/addons/spider-man-el-sando

### Best tiny/update-friendly Bedrock QoL experiment: Actual Notes
New Aug 22 for 26.40, achievement-compatible, no experimental toggles, and does not use `player.json`. It displays notes from paper on the action bar. Adoption is effectively zero today, so this is a TEST/WATCH item, not a default install.
- https://www.curseforge.com/minecraft-bedrock/addons/actual-notes

## Marketplace

### Best official August Marketplace category pick: Auto Factory / Warpstones / Uncrafter+ 2.1
Mojang’s official August 2026 Marketplace Pass roundup highlights several unusually useful add-ons. For general survival utility, **Auto Factory** (automation), **Warpstones** (multiplayer-friendly teleportation), and **Uncrafter+ 2.1** (reversible crafting with broad addon compatibility) are the strongest utility-oriented picks from the official August lineup.
- https://www.minecraft.net/en-us/article/marketplace-content-august-2026

## Artistic map / Planet Minecraft

Planet Minecraft’s web-search index available to this run lagged behind Aug 22, so I am **not** inventing a same-day winner. The strongest recent current map surfaced is **Trident Cliffs City** (Aug 5): a large player-scale island city with skyscrapers, historic districts, suburbs, harbor, interiors, terrain-integrated urban design and deliberate entity/block-entity performance restraint. It is still 80% complete.
- https://www.planetminecraft.com/project/trident-cliffs-city/

## Tool / launcher

### Best current integrated mod-discovery/updating tool: Modrinth App 0.18.0
Released Aug 18 with a redesigned Play experience, better instance creation/search, icon creation, onboarding improvements and multiple update/content-management fixes. Modrinth also says content updates are checked immediately when viewing an instance’s content rather than waiting for cache invalidation.
- https://modrinth.com/app
- https://modrinth.com/news/changelog

### Best power-user launcher baseline: Prism Launcher 11.0.3 stable
Prism remains the strongest current power-user multi-instance baseline; a newer 11.1.0-pre1 exists for bleeding-edge testing, but 11.0.3 is the safer stable recommendation.
- https://prismlauncher.org/
- https://prismlauncher.org/news/

## Today’s priority order for the Forge 1.20.1 pack

1. **Test Legendary Block Entities 0.11.0 first.** It is exactly the kind of client-only, world-safe performance gain worth adding if it behaves cleanly beside the current renderer stack.
2. **Audit Client Tweaks 11.1.11** if it is not already present.
3. **Test Epic Fight Guard Fix** in a copied instance, specifically shield + one-handed guard, WOM Perfect Bulwark, Nightfall parry and unarmed combat.
4. **Check Collections Of Optimizations installed version** before doing anything; update only if today’s build is newer than the installed jar.
5. **Keep N87 PBR as a visual test**, not a permanent default until coverage/quality is inspected in-game.
6. **Do not force World Bosses Fabric 1.20.1 through Connector** just because it exists; track the new Forge 26.2 path instead.
7. **Keep Bedrock worldgen add-ons out of forever worlds unless intentionally committed**, while favoring resource-only/client-presentation packs for update-friendly worlds.

## Schedule policy added today

The daily scout now evaluates 30 explicit categories including Java/Bedrock, performance, compatibility, client-only, world-safe/update-friendly, resource packs, shaders, datapacks, Marketplace, Planet Minecraft artistic maps, launchers/tools, diagnostics, modding tools and future-version projects. It also records client-only status, server requirement, save/world-data impact, removability confidence and future-upgrade friendliness instead of treating every mod as equally safe.
