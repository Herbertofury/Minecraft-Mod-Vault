# Noxviola FPS / Error Cleanup Wave 2

Status: **candidate / real Noxviola runtime acceptance pending**.

Wave 2 is designed to compose with Wave 1 and preserves content, entity counts, render distance, physics features, and visual quality.

## Main FPS target: BadOptimizations 2.4.1

The pack config enables `enable_entity_renderer_caching`, but stock BadOptimizations disables that whole optimization whenever Twilight Forest is installed. The upstream incompatibility exists because specific Twilight Forest renderers such as the Naga/chain-goblin family are unsafe with the one-renderer-per-EntityType cache.

The Noxviola patch removes only Twilight Forest from the global built-in disable and filters renderer-cache population instead: renderer classes in the `twilightforest.*` package are deliberately left uncached, so BadOptimizations' existing null fallback uses the stock dispatcher for those entity types. Every other safe renderer retains the cached fast path. Other BadOptimizations incompatibility guards remain intact.

Output: `BadOptimizations-2.4.1-1.20.1-Noxviola-SelectiveTFRendererCache-v1.jar`

SHA-256: `01aebad5ce4afe919e0c968b0902d5ff4b71581a89cf5010bdc6453249825743`

## Entity Culling / Immersive Vehicles

The installed Entity Culling 1.10.5 already carries the three MTS IDs in its tick-culling whitelist, but not its render-culling whitelist. The Noxviola patch merges these three IDs into the loaded `entityWhitelist` before the internal registry whitelist is built, while preserving all user entries:

- `mts:builder_existing`
- `mts:builder_rendering`
- `mts:builder_seat`

Output: `entityculling-forge-1.10.5-mc1.20.1-Noxviola-MTS-RenderWhitelist-v1.jar`

SHA-256: `553ee648ac2015feffbaf4c0b7dda38dc0718070148d31cbd519241ca10f90ea`

## Physics Mod 3.0.20

The mod JAR contains `assets/physicsmod/cloth/Vanilla Cape.dae`, but its cloth loader reads `cloth_local/Vanilla Cape.dae` and the Noxviola log shows the default cape failing to load. The patch copies the bundled DAE to `cloth_local` only when it is missing. Existing/custom cape bytes are never overwritten.

Output: `physics-mod-3.0.20-mc-1.20.1-forge-Noxviola-BundledCapeFix-v1.jar`

SHA-256: `a10a0c24fb458fd78b058b23b5f55107d3ddd6f93a89e75fb24b3562e8f00058`

The separate `PhysXGpu_64.dll` / CUDA probe is intentionally unchanged; no fake DLL or unsupported CUDA hack is included.

## GTBCS Geomancy Plus 2.0.0

Two concrete packaged-resource defects are repaired:

1. `data/gtbcs_geomancy_plus/loot_tables/chests/test.json` is zero bytes and causes an EOF datapack parse failure. It becomes a valid empty chest loot table, preserving its prior no-drop semantics.
2. Stonecaller's particle texture points at the wrong namespace/path. It now resolves to the PNG that actually ships in the JAR.

Output: `gtbcs_geomancy_plus-2.0.0-1.20.1-Noxviola-DataResourceFix-v1.jar`

SHA-256: `b144705b0e4c526e5bf328964a86eb299573024bd25894eaf759a96831670146`

## Verification

All four outputs pass ZIP/JAR integrity and exact changed-path audits. No input JAR contained signatures. Mod metadata is preserved. Deterministic tests passed for selective renderer filtering, MTS whitelist idempotence, Physics Mod bundled-cape extraction/no-overwrite behavior, and the GTBCS texture target.

The full binary bundle and exact Noxviola patch-source checkpoint are persisted in the Google Drive patched-mods workspace. Real Forge acceptance remains pending the user's same-scene Wave 2 run.

The 37 `Receive capability sync packet from unknown entity.` warnings are deliberately not suppressed here: CapabilitySyncer 4.0.1 and Celestisynth 1.3.4 were traced, but neither exact source contains that log string, so the causal emitter is not yet proven.
