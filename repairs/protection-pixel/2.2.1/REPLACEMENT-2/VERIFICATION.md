# Protection Pixel 2.2.1 — UI and Armor Render Performance Fix — REPLACEMENT-2

Verified: 2026-09-14 / 2026-09-15 closeout
Target: Minecraft 1.20.1 / Forge 47.x / Java 17
Install mode: **REPLACEMENT**
Internal identity: `protection_pixel` / `2.2.1` / `Protection Pixel`

## Result

This replacement preserves the previously accepted Protection Pixel UI/QoL render repair and adds a targeted repair for the newer Spark hotspot in `ArmorrenderProcedure.execute`.

Final artifact:

- filename: `Protection-Pixel-2.2.1-UI-and-Armor-Render-Performance-Fix-REPLACEMENT-2.jar`
- SHA-256: `8805c7b504d7f7e68c714770cbfdd4551c492a423e722ce970f71dde89297511`
- size: 3,247,645 bytes
- entries: 1,403
- classes: 709
- class major: 61 / Java 17 for all classes
- ZIP integrity: PASS
- archive signatures: none

This exact SHA is **strongly static/package/bytecode/behavior verified but not yet native-runtime verified** in the user's complete instance.

## Lineage

Exact stock target:

- `protection_pixel-2.2.1-forge-1.20.1.jar`
- SHA-256: `d94b944dce3a4cc0d88c167366bf9e95efa5e58212403df7fec463d490f36592`
- CurseForge project: Create: Protection Pixel
- CurseForge project ID: `1023409`
- exact Forge 1.20.1 file ID: `7549820`

Accepted prior UI/QoL repair retained as the R2 baseline:

- `Protection-Pixel-2.2.1-UI-QoL-Performance-Fix-REPLACEMENT-1.jar`
- Drive ID: `1arBKF2naEi8x6e_1hvDC3PYEdkiL7NqX`
- SHA-256: `dc605fccd565ce16239e615189db4b484e862cd72b8b187831626c8e58b79c7b`

R2 supersedes R1; do not install them together.

## Fresh profiler target

The prior 90-second Spark client profile identified:

- `ArmorrenderProcedure.execute`: about **1.7555% inclusive** of the render thread
- method self time: about **1.6755%**

## Root cause

`ArmorrenderProcedure.execute(Event, LevelAccessor)` runs from the client `RenderLevelStageEvent.Stage.AFTER_ENTITIES` path. The original/R1 implementation reads render distance, iterates every chunk in the square around the player, iterates every block entity in each chunk, filters down to `ARMORHANGER` and `ARMORLOADPLATFORM`, then renders the previews. That full discovery scan was repeated every rendered frame.

## R2 repair

R2 adds `net/mcreator/protectionpixel/repair/ArmorRenderTargetCache` and rewrites only the discovery setup inside `ArmorrenderProcedure.execute(Event, LevelAccessor)`.

The helper preserves the original scan radius and traversal order but performs the full discovery scan at most once for a stable client game tick/player chunk/render-distance/world identity. It caches only target `BlockPos` values, holds the cached `ClientLevel` identity through `WeakReference`, retains no static strong world/BlockEntity reference, and re-resolves the small filtered target-position set each rendered frame.

Same-tick removal/replacement at an already-known target position is therefore observed immediately without a render-distance rescan. Newly introduced target positions become discoverable on the next game tick or immediately when the player chunk, render distance, or client-world identity changes.

No thread, executor, timer, scheduler, off-thread world access, content reduction, render-distance reduction, entity/spawn reduction, visual-quality reduction, or artificial cap was added.

## Render behavior preserved

The existing armor/load-platform rendering body remains the prior implementation. R2 does not rewrite armor/load-platform type handling, armor equipment behavior, entity preview rendering, direction/yaw handling, folded steam ectoskeleton preview, cog rendering, lighting/render-buffer calls, or per-frame render cadence.

## Safety hardening

An earlier internal candidate cached `BlockEntity` values directly. It passed functional tests but could retain a previous client world's block entities until a later world render invalidated the cache. That candidate was discarded before release. Final R2 caches positions only and uses a weak world-identity reference.

## Exact change set versus accepted R1

`jar_diff.py` R1 -> final R2:

- added exactly `net/mcreator/protectionpixel/repair/ArmorRenderTargetCache.class`
- content-changed exactly `net/mcreator/protectionpixel/procedures/ArmorrenderProcedure.class`
- removed entries: 0
- metadata-only changed entries: 0
- unchanged existing entry contents: 1,401

Critical class hashes:

- R1 `ArmorrenderProcedure.class`: `96ca28de28eb4812fa570bd94498cd4c27af93466c225eb75738c6cbdedca5c2`
- final R2 `ArmorrenderProcedure.class`: `dd4c03ece38c310f285ce4dcb35bbbd3a987fff421d25747a2badbb777a22148`
- final `ArmorRenderTargetCache.class`: `3fa36f5527330d1cfda059ddaee57823a01c5ef00122c0817197ed8ff1e11d8d`

The complete `javap -p -s` field/method descriptor surface of `ArmorrenderProcedure` is unchanged from R1.

## Previous R1 UI repair retained

All 13 pre-existing repair-layer classes remain byte-for-byte R1. `protection_pixel_ui_repair.mixins.json` is byte-identical. The accepted ALT+U UI editor, overlay positioning/scaling/visibility, reactor timer cache, hidden overlay cancellation, config behavior, PoseStack reflection caching and original reactor toggle are retained.

## Loader / compatibility identity

The final JAR preserves `modId="protection_pixel"`, version `2.2.1`, display name `Protection Pixel`, `logoFile="logo.png"`, Forge `javafml` loader range `[47,)`, Minecraft `[1.20.1]`, and the optional JEI dependency.

Unchanged metadata hashes:

- `META-INF/mods.toml`: `ae038f9bf727630227978f984e1aa8b6e87d011f17677471ed441b0332abd8fa`
- `META-INF/neoforge.mods.toml`: `efd70c32727ddf4c59efa0fc4bac68a16e07d470d30b6187e5cf8ef74068ebf3`
- `META-INF/MANIFEST.MF`: `5d50ea40a9bf7157b12afe053cf8e1ccefaf9794bed8c92c5253d43bfb71f22c`

## Bytecode verification

Whole final JAR ASM `BasicVerifier`: **709 classes / 3,177 methods / 0 failures**.

`ArmorrenderProcedure`: all 19 methods verify, valid frames, descriptor surface unchanged. New helper: all 4 methods verify, Java 17 class-major 61. ZIP/JAR validation passes and no signatures exist.

## Deterministic cache equivalence / invalidation test

The exact final helper implementation was exercised against a functional stub world model for **10,000 randomized client-world cases**. It proves the same target filtering/order; no repeated full chunk-grid scan in a stable tick/key; rebuild on game tick, player chunk, render distance or client-level identity; and same-tick removal/replacement at an already-known target position through live re-resolution.

Result:

`PASS cases=10000 stable-state target order/filter equivalence; same-tick chunk-scan reuse; live target revalidation; tick/chunk/radius/level invalidation`

This is correctness evidence, not a measured FPS claim.

## Repair Mark v2

The validated R1 mark is carried forward byte-for-byte from the exact upstream `logo.png` declared by Protection Pixel 2.2.1.

- unmarked art SHA-256: `d90b38ebea3da4ca2eef94a2e6b668e749a789706210b3884f7ee5ee6b676bef`
- marked art SHA-256: `e27243ad79a36646f53f0bd37b4b8b27a1ba57ebcf0011ff4eba5a87498ff9fc`
- 48x48 QA SHA-256: `14eb18197f696a371c1e046fc3e69cee781ed3caed083068b91437b10b7f9c23`
- unmarked R2 SHA-256: `8c10d63d9a867f2a546be1d8daadf056e43202d55933dfa18b49a5e5b2f8ef68`
- marker-only diff: exactly `logo.png`, zero added/removed/metadata-only changes
- image generation: not used

## Runtime gate

Install only REPLACEMENT-2; test armor hangers/load platforms, equipment changes, chunk/render-distance boundaries, ALT+U, Save & Quit/reload; then run `/sparkc profiler start --timeout 90 --thread *` and provide the fresh `.sparkprofile` plus `latest.log` before promoting this exact SHA to runtime-verified.
