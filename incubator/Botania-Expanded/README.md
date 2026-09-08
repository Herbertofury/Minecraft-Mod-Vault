# Botania Expanded

Forge 1.20.1 addon/visual overhaul for Botania magical flora.

## Current development slice — Endoflame family

`0.1.0-dev2` starts the mod with a premium voxel reconstruction of the Endoflame plus two optional variants:

- **Endoflame (Default)** — replaces the crossed-sprite world render with a sculpted purple/magenta flame-flower while preserving Botania's original block/entity and gameplay behavior.
- **Warped Endoflame** — cyan/teal warped-biome interpretation with native Botania generating-flower behavior.
- **Forgotten Endoflame** — pale ash/ivory/gold interpretation with native Botania generating-flower behavior.

Both custom variants have ground, floating, and potted forms, Botania generating-flower tags, Petal Apothecary recipes, loot tables, creative-tab exposure, Wand HUD compatibility, and Botania-style floating conversion recipes.

## Visual system

The flower is rendered as many Minecraft-native cuboids instead of a flat cross model:

- layered stem and leaves
- broad curling outer petals
- tall inner flame petals
- separate normally-lit and emissive passes
- animated breathing core and subtle petal sway
- active-burning intensity boost
- sparse orbiting voxel sparks
- variant palettes share one geometry system
- allocation-free reusable `VoxelRenderUtil` primitives for future flower families

The renderer deliberately leaves the stem/leaves normally lit; only selected petal/core accents receive full-bright emissive treatment. The original Endoflame renderer replacement is registered after Botania at `EventPriority.LOWEST` so Botania cannot accidentally overwrite it later in client initialization.

## QA helper

In a creative/cheats-enabled test world:

```mcfunction
/botaniaexpanded showcase
/botaniaexpanded showcase day
/botaniaexpanded showcase night
```

The command stages the default, warped, and forgotten flowers on a deterministic plinth, drops coal to activate them, and normalizes weather/time for repeatable screenshots.

## Build target

- Minecraft 1.20.1
- Forge 47.4.23
- Java 17
- Botania 1.20.1-455+

Current dev2 offline compile, production/reobf build, JSON/resource validation, potted/floating parity checks, renderer-ordering guard, and hot-path allocation guard pass. Fresh native visual parity is still gated on rematerializing the previously proven packaged Forge client harness plus Botania's Patchouli/Curios runtime dependencies.

The full buildable source archive, development JAR, checksums, design references, and detailed checkpoint are stored in the Minecraft Dev Kit project folder on Google Drive. Botania attribution is documented there in `NOTICE.md`.
