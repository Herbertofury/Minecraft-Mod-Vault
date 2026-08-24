# Minecraft Ecosystem Scout — 2026-08-23 Late-Cycle Delta

This checkpoint records only material changes from the prior 2026-08-23 leaderboard. Unlisted categories are **NO MATERIAL CHANGE**.

## Material leaderboard changes

### #24 Best Experimental / Bleeding-Edge Project — NEW LEADER
**Retromod 1.3.0-snapshot.8** becomes the current experimental leader. Retromod transforms older Minecraft mods forward across game versions rather than merely translating loaders. The verified NeoForge 26.2 snapshot is beta and was re-tested by its maintainer on 26.2 Fabric and NeoForge after its bytecode-library/toolchain maintenance update.

- Status: **TEST BRANCH / FUTURE-HIGH-VALUE**
- Verified line: `1.3.0-snapshot.8+26.2-neoforge`
- Current project coverage: 1.20.x, 1.21.x, 26.1.x, 26.2; Fabric/Forge/NeoForge/Quilt ecosystem
- Stable line also exists: Retromod 1.2.0
- Safety: run in a cloned instance and preserve original jars/world backups; transformed mods inherit their own save/world risks
- Current-pack note: this does **not** replace Sinytra Connector in an established Forge 1.20.1 pack. Its strongest value is future migration testing.
- Project: https://modrinth.com/mod/retromod
- Exact verified snapshot: https://modrinth.com/mod/retromod/version/1.3.0-snapshot.8%2B26.2-neoforge
- Source: https://github.com/Bownlux/Retromod

### #25 Best Future-Version Project Worth Tracking — NEW LEADER
**Retromod 1.3.0-snapshot.8** also moves into the future-version lead because it directly attacks the problem of carrying old mods onto newer Minecraft hosts. Easy Model Entities 2.3.0 + Easy NPC 7.9.0 remain strong runner-ups for future creator tooling.

## Fresh runner-ups worth keeping

### Outline N' Stuff — #26 Client-Side / #29 QoL-UI runner-up
A polished client-only visual/QoL mod for 1.21.1 and 26.2 on Fabric + NeoForge. It adds geometric or post-process outlines, enchant runes, durability/hit feedback and rarity/chameleon color modes. The project explicitly documents compatibility with Punchy, ImmediatelyFast, GeckoLib/AzureLib, Create, Twilight Forest and many custom item renderers.

- Client only: **Yes**
- Server required: **No**
- Save/world data: **client config only**
- Safe to remove: **High confidence**
- Update friendliness: **Excellent**
- Current-pack caveat: **no Forge 1.20.1 build**, so it is a future-version runner-up rather than a current install recommendation
- Known limitation: Draconic Evolution's custom GPU item renderer is not intercepted
- Project: https://modrinth.com/mod/outlinestuff

### PathMax — #14 Shader experimental runner-up
A new Iris-only voxel path-tracing shader covering Java 1.20.1–1.20.6, 1.21.x, 26.1.x, 26.2 and a 26.3 snapshot. It provides path-traced GI, colored bounce light, voxelized entity shadows/reflections, labPBR fallback, optional POM and five performance profiles.

- Client only / save-safe: **Yes**
- Required renderer: **Iris only**
- GPU/API: **OpenGL 4.6 compatibility profile**
- Current-pack caveat: no verified Distant Horizons support was found in its current project documentation, and Iris-only makes it a poor direct fit for Forge 1.20.1 + Oculus
- Overall shader leader remains **Complementary Reimagined**
- Project: https://modrinth.com/shader/pam

### Bedrock #17 fresh visual runner-ups
**Cinematic Visuals Shader 1.2** and **Unbound Visuals 2.3.1** both received August 23 updates and are retained as meaningful visual runner-ups behind Newb X Supplementary 8.0.

- Cinematic Visuals 1.2: high-resolution texture additions, new water/fog JSON, identifier fixes and smaller package size
- Unbound Visuals 2.3.1: stylized rose-hour Vibrant Visuals overhaul with pink volumetric haze, god rays, aurora and meteor effects
- Both currently declare Bedrock **26.40**. Stable Bedrock is **26.45**, so validate before promotion/default use.
- Cinematic Visuals exact file: https://www.curseforge.com/minecraft-bedrock/texture-packs/cinematic-visuals-shader-vibrant-visuals-pack/files/8714077
- Unbound Visuals: https://www.curseforge.com/minecraft-bedrock/texture-packs/unbound-visuals

## Rechecked specialty visual categories

- #31 Best Animated Resource Pack — **NO MATERIAL CHANGE**: Actions & Stuff 1.11 overall / Bedrock; Fresh Animations remains Java leader.
- #32 Best CIT Pack — **NO MATERIAL CHANGE**: Kaydicraft CIT 1.45 remains best single classic 1.20.1 CIT pack; FAYE remains the active suite runner-up.
- #33 Best 3D Resource Pack — **NO MATERIAL CHANGE**: 3D Default 1.15.0 remains the safer overall leader; Better 3D Blocks remains high-detail runner-up with its Distant Horizons caution.

## Other scoreboard results

All categories not named above are **NO MATERIAL CHANGE** for this cycle. The Planet Minecraft artistic sweep did not surface evidence strong enough to replace Trident Cliffs City. No new stable performance, combat, movement, Create, map, Marketplace, server/admin or datapack candidate cleared the quality bar strongly enough to change its current leader.

## Persistence / safety

The canonical Google Sheet was updated with Retromod, Outline N' Stuff, PathMax, Cinematic Visuals and Unbound Visuals; the leaderboard rows and delta ledger were updated in place. Resource/shader-only entries are world/save neutral aside from client configuration and resource-pack ordering. Retromod is explicitly **not** world-neutral merely because the transformer itself adds no gameplay content: transformed mods retain their own registries/worldgen/save-data behavior and must be tested accordingly.
