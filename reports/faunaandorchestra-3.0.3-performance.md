# Fauna & Orchestra 3.0.3 Forge 1.20.1 — performance optimization receipt

## Lineage

- User input SHA-256: `3328e5ad3b509b517e30564c583e3d538f9931fb1cfed9457183c57c1310a685`
- Exact upstream source: `migueel26/Fauna-And-Orchestra@131bdcbf76aec07267d68a57185bb103669af83e`
- Reproducible patch harness: `.ci/faunaandorchestra_3_0_3_performance_patch.py` + `.ci/faunaandorchestra_3_0_3_performance_patch_v2.py`
- Build workflow: `.github/workflows/faunaandorchestra-3.0.3-performance.yml`
- Verified build commit: `c8f57adf078160aab055f6c6a373241fc997cf83`
- GitHub Actions run: `34027623367`
- Actions artifact: `9987621014`
- Optimized JAR SHA-256: `8cea89050d1cdff677d8394c02605f21f1758395683bb046f8d127d18ade7907`
- Optimized JAR size: `50,002,459` bytes

## Implemented performance fixes

1. Frog choir global `LivingTickEvent` now rejects non-frogs before side/modulo/RNG work and filters frogs in the entity query.
2. Conductor Absolute Hearing player query runs on its existing 30-tick decision cadence instead of every tick.
3. Faust listener polling fixes the missing 60-tick delay reset and eliminates per-tick empty-list allocation.
4. Dan B listener polling fixes the same missing reset/allocation bug.
5. Great Composer music listener tracking iterates server players directly instead of querying a huge entity AABB every tick.
6. Orchestra listener start/stop replaces an approximately 1,030,301-position `BlockPos` sweep per pass with loaded-chunk block-entity iteration while preserving the same 50-block listener area.
7. Orchestra player tracking iterates players directly and centroid computation no longer allocates streams.
8. Client orchestra-presence sound check collapses up to three stream traversals into one allocation-free loop.
9. Parrot conductor check filters inside the spatial query instead of stream/filter/findAny.
10. `AnimalEatGoal` closest-item selection uses one allocation-free pass.
11. Anya HUD's 30-block entity lookup is cached per game tick instead of per rendered frame; dialogue sound emits once per dialogue tick.
12. Great Composer HUD gets the same per-tick cache/typewriter-sound fix.
13. Dialogue UI reuses one client RNG instead of allocating a `RandomSource` per sound.
14. Six GeckoLib block-entity renderers no longer force `shouldRenderOffScreen=true`; normal frustum culling is restored.
15. Those six block entities' render bounds are tightened from +/-16 to a conservative +/-3 blocks after measuring shipped geometry (max model extent about 2.22 blocks; max animated translation about 0.52 block).

## Verification

- Forge Java 17 clean reobfuscated build: PASS.
- JAR `unzip -t`: PASS.
- Class entry count: original 585, optimized 585.
- Total JAR entry names: identical; zero missing and zero added entries.
- `assets/` + `data/` entries: original 1,267, optimized 1,267.
- Missing `assets/`/`data/` entries: 0.
- Byte-different common `assets/`/`data/` files: 0.
- Production bytecode confirms all six renderers return `false` from `shouldRenderOffScreen`.
- Production bytecode confirms Faust and Dan B reset their listener-search delays after polling.
- Production orchestra-goal class contains no `betweenClosedStream` symbol.
- Production sound-engine method uses one iterator loop rather than three stream traversals.
- Drive readback of the final JAR matches SHA-256 `8cea89050d1cdff677d8394c02605f21f1758395683bb046f8d127d18ade7907` and size `50,002,459` bytes.

## Scope note

No mobs, mechanics, animations, music, effects, models, textures, sounds, recipes, loot/data content, or other packaged resources were removed. This run proves build integrity and structural performance fixes. It does not claim a measured in-game FPS/TPS delta because a native Minecraft client/profile was not run in this CI/container environment.
