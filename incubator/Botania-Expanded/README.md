# Botania Expanded

Forge 1.20.1 addon and high-fidelity visual expansion for Botania magical flora.

## Endoflame family — 0.1.0-dev21 approval candidate

The first completed visual slice rebuilds the Endoflame family as Minecraft-native faceted voxel flowers while keeping the stock flower's gameplay intact and adding two optional generating-flower variants.

- **Endoflame (Default)** — purple/magenta flame-flower renderer applied directly to Botania's existing Endoflame block entity; fuel, mana, recipes, progression, and placement remain Botania-owned.
- **Warped Endoflame** — cyan/teal warped interpretation using the same renderer architecture and Endoflame-style generating behavior.
- **Forgotten Endoflame** — ash/ivory/gold forgotten-ruins interpretation using the same renderer architecture and Endoflame-style generating behavior.

The custom variants include ground, floating, and potted forms, Botania generating-flower tags, Petal Apothecary recipes, loot tables, creative-tab exposure, Wand HUD compatibility, and Botania-style floating conversion recipes.

## dev21 visual system

The renderer was rebuilt through repeated real Forge-client screenshot passes against the approved concept sheets. The current geometry uses:

- six unequal, plan-view-curled outer flame blades
- allocation-free tapered/faceted petal segments rather than thin rectangular ribbons
- cup/lift/roll/twist shaping per blade for a readable top, side, and front silhouette
- a taller dominant hooked inner tongue plus shorter supporting flame tongues
- a compact full-bright faceted core/chamber
- a clean dark stem with two side branches instead of a bulky leaf rosette
- shared geometry with variant-specific purple, warped-cyan, and forgotten ash/gold palettes
- normally lit structural surfaces plus selective emissive accents and subtle active particles
- reusable `VoxelRenderUtil` primitives intended for later flower families

No generated concept art is used as runtime output. The QA evidence in this checkpoint is captured from a real packaged Forge 47.4.23 client under Xvfb/Mesa.

## Gameplay parity

The renderer override is client-only for Botania's original Endoflame. It does not replace the original Endoflame gameplay block entity.

The two custom variants intentionally follow Botania Endoflame behavior:

- consume one valid fuel item at a time
- cap accepted fuel burn time at 32000 before halving
- generate 3 mana every 2 ticks while burning
- cap internal mana at 300
- retain a 3-block fuel pickup range
- retain Botania-style activation/smoke behavior

## QA helper

In a creative/cheats-enabled test world:

```mcfunction
/botaniaexpanded showcase
/botaniaexpanded showcase day
/botaniaexpanded showcase night
```

The command stages Default / Warped / Forgotten on a deterministic grass plinth, fuels them, and normalizes weather/time for repeatable visual QA.

## Verified build target

- Minecraft 1.20.1
- Forge 47.4.23
- Java / Eclipse Temurin 17.0.20.1
- Botania 1.20.1-455
- Patchouli 1.20.1-85
- Curios 5.14.1+1.20.1

`tools/validate_release.py` passes all 32 parsed JSON resources and the required integration-resource checks. `compileJava` and `reobfJar` both pass for dev21.

See `NOTICE.md` for Botania attribution and development/license notes.
