# Protection Pixel 2.2.1 — Event-Driven Armor Render Index — REPLACEMENT-3 Verification

Verified: 2026-09-14 / 2026-09-15 closeout  
Target: Minecraft 1.20.1 / Forge 47.x / Java 17  
Install mode: **REPLACEMENT**  
Internal identity: `protection_pixel` / `2.2.1` / `Protection Pixel`

## Result

REPLACEMENT-3 supersedes the previous per-tick target cache with a lifecycle-maintained sparse index. The render callback no longer discovers armor-preview targets by scanning chunks or unrelated block entities at any frame/tick cadence.

Final artifact:

- filename: `Protection-Pixel-2.2.1-Event-Driven-Armor-Render-Index-REPLACEMENT-3.jar`
- SHA-256: `2a750dcfda368e2e205587aae26d81a20ba66c73e57f8b4ab705a68956bd3613`
- size: 3,251,907 bytes
- entries: 1,407
- classes: 713
- class major: 61 / Java 17 for all classes
- ZIP integrity: PASS
- archive signatures: none

This exact SHA is **strongly static/package/bytecode/lifecycle-stress verified but not yet native-runtime verified** in the user's complete instance.

## Why R3 replaces R2

The original mod/R1 performed render-distance-wide armor-preview target discovery every rendered frame. R2 correctly reduced that to once per game tick, but that remained polling. R3 removes steady-state discovery polling entirely.

The new architecture is:

- **O(changes)** index maintenance when the two actual Protection Pixel target block entities load/remove;
- **O(actual indexed targets in render range)** render lookup;
- **O(0 loaded chunks / unrelated block entities)** steady-state discovery cost;
- no timer, tick subscriber, scheduler, executor, background watchdog, or off-thread world access.

This implements the reusable Repair Brain policy `minecraft-global-hot-path-event-driven-indexing-policy-2026-09-14`: repeated hot-path rediscovery is replaced by exact lifecycle invalidation whenever the platform exposes trustworthy lifecycle hooks.

## Event-driven lifecycle index

New helper:

`net/mcreator/protectionpixel/repair/ArmorRenderTargetIndex`

The two actual target block entities now maintain the sparse index themselves:

- `ArmorhangerBlockEntity.onLoad()` -> register position;
- `ArmorhangerBlockEntity.setRemoved()` -> unregister after its existing cleanup;
- `ArmorloadplatformBlockEntity.onLoad()` -> register position;
- `ArmorloadplatformBlockEntity.setRemoved()` -> unregister after its existing cleanup.

Forge's block-entity load/removal lifecycle is therefore the source of truth rather than a render/tick scan. The helper references common Minecraft types only and immediately no-ops on a server level, so the added lifecycle calls are dedicated-server safe.

The index stores immutable `BlockPos` + precomputed chunk coordinates. It does **not** hold a static strong reference to a world or BlockEntity. Active-level identity is a `WeakReference<Level>` that is replaced only when the active world changes; normal render binds allocate no new weak references.

## Steady render path

`ArmorrenderProcedure.execute(...)` keeps the original render-distance radius and still renders every live armor hanger/load platform every rendered frame. It now receives a reusable map facade over the sparse lifecycle index rather than rebuilding a map from every loaded chunk.

The view/entry-set/cursor objects are reused. The cursor:

1. walks only indexed Protection Pixel positions;
2. checks the precomputed target chunk coordinates against the unchanged render-distance radius;
3. resolves the live block entity for an in-range indexed position;
4. self-heals a stale position if another mod bypassed normal lifecycle removal.

No full chunk grid, `LevelChunk`, `getChunk`, unrelated block-entity map, or registry scan is referenced by the new helper.

## Render/gameplay behavior preserved

R3 does not rewrite the actual armor/load-platform render body. It preserves:

- armor hanger/load-platform behavior;
- equipment slot lookup and live equipment changes;
- armor/entity preview creation and rendering;
- block direction/yaw transforms;
- folded steam ectoskeleton preview;
- load-platform cog rendering;
- lighting/render-buffer behavior;
- the original render-distance radius;
- per-frame rendering/animation cadence;
- content, recipes, progression, entity/block-entity counts and simulation.

The previously accepted R1 UI/QoL repair also remains present: ALT+U editor, overlay transforms/visibility, reactor timer same-tick memoization, hidden-overlay HEAD cancellation, bounded config reload and cached reflection handles.

## Exact R2 -> R3 package diff

`jar_diff.py` reports:

Added exactly five helper classes:

- `ArmorRenderTargetIndex.class`
- `ArmorRenderTargetIndex$Target.class`
- `ArmorRenderTargetIndex$TargetCursor.class`
- `ArmorRenderTargetIndex$TargetEntrySet.class`
- `ArmorRenderTargetIndex$TargetMapView.class`

Removed exactly the old polling helper:

- `ArmorRenderTargetCache.class`

Content-changed exactly three existing classes:

- `ArmorhangerBlockEntity.class`
- `ArmorloadplatformBlockEntity.class`
- `ArmorrenderProcedure.class`

Other results:

- metadata-only changed entries: **0**
- unchanged existing entry contents: **1,399**
- complete ZIP metadata parity for every intersecting R2/R3 entry: PASS

The Python ZIP-writer metadata normalization detected during development was corrected before release; no untouched-entry ZIP metadata drift remains in the final artifact.

## Public/private descriptor compatibility

`javap -p -s` R2 -> R3:

- `ArmorrenderProcedure`: **empty descriptor-surface diff**;
- `ArmorhangerBlockEntity`: only one method added — `public void onLoad()V`;
- `ArmorloadplatformBlockEntity`: only one method added — `public void onLoad()V`;
- each existing `setRemoved()` descriptor remains unchanged.

Bytecode inspection confirms each `setRemoved()` retains its original superclass/capability cleanup and performs exactly one `ArmorRenderTargetIndex.unregister(...)` immediately before return. Each added `onLoad()` performs exactly one `register(...)`.

## Whole-JAR bytecode verification

Final R3 ASM `BasicVerifier`:

- classes: **713**
- methods: **3,202**
- failures: **0**

All 713 classes are Java 17 class-major 61. The modified control-flow classes retain verifier-valid frames.

## Lifecycle/index torture test

The exact hardened helper implementation was compiled against functional 1.20.1-shaped stubs and exercised through **50,000 randomized lifecycle operations**.

Observed result:

`PASS randomized_steps=50000 queries=17786 stale_self_heals=1373 unrelated_block_entities=20000 indexed_lookup_bound=332 reusable_map_view=true reusable_iterator=true server_noop=true world_transition=true`

This proves the tested contract for:

- repeated register/unregister;
- duplicate registration;
- live replacement at an indexed position;
- deliberately missing lifecycle removal with stale-entry self-heal;
- render-radius filtering;
- world transition clearing;
- delayed old-world unregister safety;
- dedicated-server no-op behavior;
- reuse of the same map view and iterator;
- a world containing 20,000 unrelated block entities without discovery work scaling to them.

After the challenge pass, the helper was additionally hardened so normal `targets(...)` render binds do not allocate a `WeakReference`; weak-reference construction occurs only at static initialization/world identity change.

This is structural correctness/performance evidence, **not** a measured FPS claim.

## Compatibility identity

The final JAR preserves the accepted Forge identity/metadata:

- `modId="protection_pixel"`
- Forge version `2.2.1`
- display name `Protection Pixel`
- `logoFile="logo.png"`
- loader `javafml`, range `[47,)`
- Minecraft `[1.20.1]`
- JEI optional dependency unchanged

Unchanged metadata hashes:

- `META-INF/mods.toml`: `ae038f9bf727630227978f984e1aa8b6e87d011f17677471ed441b0332abd8fa`
- `META-INF/neoforge.mods.toml`: `efd70c32727ddf4c59efa0fc4bac68a16e07d470d30b6187e5cf8ef74068ebf3`
- `META-INF/MANIFEST.MF`: `5d50ea40a9bf7157b12afe053cf8e1ccefaf9794bed8c92c5253d43bfb71f22c`

No archive signatures are present.

## Repair Mark v2

The accepted exact upstream Protection Pixel artwork is carried forward without image generation:

- upstream unmarked `logo.png`: 1000x1000 RGBA
- unmarked art SHA-256: `d90b38ebea3da4ca2eef94a2e6b668e749a789706210b3884f7ee5ee6b676bef`
- marked art SHA-256: `e27243ad79a36646f53f0bd37b4b8b27a1ba57ebcf0011ff4eba5a87498ff9fc`
- 48px QA SHA-256: `14eb18197f696a371c1e046fc3e69cee781ed3caed083068b91437b10b7f9c23`
- integration: embedded `logo.png`
- image generation: not used

Verified unmarked R3:

- SHA-256: `936134b624f6afd1178f3fa45959b758d3b4af4cabd416b27e6e3dcfa9cc5950`
- size: 3,266,617 bytes

Unmarked R3 -> final R3:

- same entry set: PASS
- content-changed: exactly `logo.png`
- added: 0
- removed: 0
- metadata-only changed: 0
- 1,406 other entry contents unchanged

## Neutral/private-label scan

Final filename, archive entry names, raw/decompressed JAR contents and repair source names were scanned case-insensitively for private/user/modpack/project labels. Result: **PASS / 0 hits**.

## Runtime gate still required

The exact full Forge client instance is not available in this repair container, so this exact SHA is not being called native-runtime verified yet.

Install/test acceptance:

1. close Minecraft;
2. remove/move stock Protection Pixel, REPLACEMENT-1 and REPLACEMENT-2;
3. install **only** `Protection-Pixel-2.2.1-Event-Driven-Armor-Render-Index-REPLACEMENT-3.jar`;
4. launch the same world/settings;
5. confirm existing armor hangers/load platforms appear immediately after chunk load;
6. place and remove both target blocks and verify previews appear/disappear immediately;
7. change equipment and verify the live preview changes normally;
8. cross chunk/render-distance boundaries and return;
9. change dimension/world, then return;
10. confirm ALT+U and the accepted UI repair still work;
11. Save & Quit, reload, and repeat the target-render checks;
12. run `/sparkc profiler start --timeout 90 --thread *` and provide the fresh `.sparkprofile` plus `latest.log`.

Acceptance is: no Protection Pixel/ASM/Mixin fatal, no missing/stale previews, R1 UI behavior preserved, and `ArmorrenderProcedure.execute` discovery self-cost collapses versus the prior equivalent profile while actual per-target rendering remains intact.
