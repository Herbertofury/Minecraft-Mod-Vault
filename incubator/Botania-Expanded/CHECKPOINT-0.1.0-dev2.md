# Botania Expanded — 0.1.0-dev2 Endoflame checkpoint

## Objective

Build the first actual Botania Expanded production slice from the approved Endoflame concept: a rich sculpted default Endoflame plus Warped and Forgotten variants, with Botania gameplay parity and reusable tooling for the remaining flower catalog.

## Implemented content

- **Default Endoflame visual overhaul** on Botania's existing Endoflame block/entity; gameplay remains Botania-owned.
- **Warped Endoflame** ground, floating, and potted forms.
- **Forgotten Endoflame** ground, floating, and potted forms.
- Variant Petal Apothecary recipes and Botania-style floating conversion recipes.
- Correct generating-special-flower and generating-floating-flower block/item tags.
- Loot tables, item models/icons, blockstates, language entries and creative-tab exposure.
- Botania Wand HUD capability for both custom generating variants.

## Visual implementation

- Renderer-owned 3D flower geometry instead of a crossed-sprite world model.
- Dark sculptural root/stem and four curled leaves.
- Six broad curling outer petals plus five steeper inner flame petals.
- Layered breathing core and sparse orbiting voxel sparks.
- Separate normally-lit base and selective full-bright emissive passes; the entire flower is never forced emissive.
- Active burning state boosts the luminous core/petal accents for both custom variants and the original Botania Endoflame.
- Subtle petal sway and core breathing keep the concept alive without noisy animation.
- Original Botania renderer is delegated for floating-island and radius/HUD behavior.

## dev2 hardening / challenge pass

- Moved the original-Endoflame renderer override onto the same `EntityRenderersEvent.RegisterRenderers` path Botania uses and registered at `EventPriority.LOWEST`, preventing Botania from accidentally re-overwriting the premium renderer later in client initialization.
- Added potted Warped/Forgotten forms and vanilla/Botania-equivalent pot loot behavior so inventory/pot parity is not lost.
- Extracted generic `VoxelRenderUtil` with allocation-free direct cuboid vertex emission and mutable chain-point scratch state for future flowers.
- Removed the first renderer prototype's per-frame primitive-array allocations from articulated petal/leaf chains.
- Extended release validation to guard renderer registration ordering, potted resources, integration tags, JSON integrity, and hot-path allocation regressions.

## Gameplay parity for custom variants

`EndoflameVariantBlockEntity` mirrors Botania's Endoflame behavior:

- one furnace-fuel item consumed at a time;
- excludes items with crafting remainders and mana spreaders;
- burn time capped at 32000 then halved like Botania;
- +3 Mana every 2 ticks while burning;
- 300 internal Mana maximum;
- 3-block square collection radius;
- Botania Endoflame sound/block events and activate/deactivate game events;
- synchronized burn state for client visuals.

## Deterministic QA fixture

Creative/cheats test command:

```mcfunction
/botaniaexpanded showcase
/botaniaexpanded showcase day
/botaniaexpanded showcase night
```

It constructs a grass plinth, places Default / Warped / Forgotten Endoflames, drops coal to activate them, and normalizes time/weather for repeatable screenshot comparison.

## Verification completed

- Java 17 compilation: **PASS**.
- Forge 47.4.23 offline production/reobf build: **PASS**.
- `tools/validate_release.py`: **PASS**, all 32 JSON resources parsed and critical integration payload present.
- Potted/floating/ground resource parity static gate: **PASS**.
- Renderer hot-path primitive-array allocation guard: **PASS**.
- Renderer override ordering guard (`LOWEST` priority): **PASS**.
- Built JAR contains critical renderer, BE, QA command, resources, README and NOTICE: **PASS**.

## Native-client boundary

A native Gradle `runClient` was already attempted for this lineage and stops before Minecraft starts because ForgeGradle needs `piston-meta.mojang.com`, which this restricted environment cannot resolve. That unchanged route is intentionally **not retried**.

The user's Dev Kit contains verified evidence of a reusable Forge 1.20.1 / 47.4.23 packaged-production `forgeclient` harness with complete official 1.20.1 assets, client SRG/client-extra, Forge client patch, libraries and Linux LWJGL natives. Exact Patchouli/Curios runtime binaries and the reusable harness payload still need to be re-materialized before fresh native visual parity can be claimed.

## Dev Kit improvements produced

- `client/render/VoxelRenderUtil.java` — reusable allocation-free runtime cuboid/chain geometry primitive library.
- `tools/voxel_flower_scaffold.py` — explicit-palette resource scaffold for renderer-owned ground/floating flowers.
- `tools/VOXEL_FLOWER_TOOLKIT.md` — visual/performance/QA contract for future Botania Expanded flowers.
- `tools/validate_release.py` — release-resource and hot-path regression validator.
- `/botaniaexpanded showcase` — reusable deterministic native visual QA fixture.

## Exact next action

Recover/reuse the already-proven packaged Forge 1.20.1 client harness plus Botania's real Patchouli and Curios dependencies, stage this dev2 JAR, execute the showcase under Xvfb, and capture Default/Warped/Forgotten active/inactive front/three-quarter/top/night evidence. Tune geometry/palette only from those native captures, then promote the first Endoflame slice.
