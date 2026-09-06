# Fauna & Orchestra 3.0.3 — Frontier Wave 4 Acceptance

Date: 2026-09-06

## Lineage

- Upstream repository: `migueel26/Fauna-And-Orchestra`
- Exact upstream source: `131bdcbf76aec07267d68a57185bb103669af83e`
- Performance branch: `perf/faunaandorchestra-3.0.3-native-profiler`
- Rejected Wave 3B remains excluded.

## Accepted baseline — Wave 3C

Native Forge/JFR run: `34054499169`

- Head: `a5cdfa07c253ffdfd498c5c5ce7d2d7f6fe00127`
- Artifact: `9995647325`
- Artifact digest: `sha256:76d68632890d2f567b44816fd78703b312e328dd616283cc0a48d1210be2dbe6`
- `MANHATTAN_ORDER_EQUIVALENCE_PASS positions=11767`
- `SPAWN_SUMMARY requestedEntities=294 placedBlocks=60`
- `SCENE_READY`
- `CrawlingDiscordBlockEntity.performSpreadLogic` allocation weight: `2,095,944`
- `CrawlingDiscordBlockEntity.canGrab` allocation weight: `1,047,968`

## Candidate — Wave 4

Native Forge/JFR run: `34056701618`

- Head: `5524dcd78eed266aff7c83aa62fcdba4e100e0d7`
- Artifact: `9996275217`
- Artifact digest: `sha256:30bd6140276f1a4aa59e8a16381feecc4ab42e89f3e4fa30409c6e82108f8fe8`
- `MANHATTAN_ORDER_EQUIVALENCE_PASS positions=11767`
- `BETWEEN_CLOSED_ORDER_EQUIVALENCE_PASS plane=9 cube=27`
- `SPAWN_SUMMARY requestedEntities=294 placedBlocks=60`
- `SCENE_READY`
- `CrawlingDiscordBlockEntity.performSpreadLogic` allocation weight: `996,152`
- `CrawlingDiscordBlockEntity.canGrab`: absent as a sampled mod allocation hotspot after reusable coordinate/state-read rewrite.

Targeted `performSpreadLogic` sampled allocation weight changed by `-1,099,792`, a **52.47% reduction** versus the accepted Wave 3C native baseline.

## Semantic safety boundary

Wave 4 changes only temporary spread-scan coordinate/iterator/state-read plumbing. It preserves the original 3x1x3 and normalized 3x3x3 candidate order, odd/even candidate parity, EAST→WEST→NORTH→SOUTH `canGrab` short-circuit order, generation limits, child timers, CLIMBER state decisions, child placement calls, sounds/particles/content, and all user-visible quality/quantity.

The runtime order verifier compares the direct loops against Minecraft 1.20.1 `BlockPos.betweenClosed` before scene setup. `spawnChild` consumes the supplied position synchronously through `level.setBlock(...)` and `level.getBlockEntity(...)`; no mutable candidate is retained for later ticks.

## Decision

**Wave 4: ACCEPTED.**

The overall short JFR mod execution sample share is not used as an FPS/TPS claim or as the Wave 4 acceptance metric because it is sampling-sensitive. The acceptance is based on the targeted allocation reduction plus exact-order native equivalence gates and successful full-scene runtime.

## Release

Clean Wave 4 release workflow commit: `18b5492e463ca7533fdad49e2c42e7f0ef23d36f`

The user-facing release build must contain verified base/extreme + accepted Wave 3A + accepted Wave 3C + accepted Wave 4 only. It must exclude rejected Wave 3B and all QA/JFR/profiling-only source.
