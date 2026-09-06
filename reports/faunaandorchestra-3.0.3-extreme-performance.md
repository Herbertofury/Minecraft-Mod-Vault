# Fauna & Orchestra 3.0.3 — Extreme Performance Verification

## Lineage

- Target: Forge 1.20.1, Fauna & Orchestra 3.0.3
- Exact upstream source: `migueel26/Fauna-And-Orchestra@131bdcbf76aec07267d68a57185bb103669af83e`
- Build branch: `build/faunaandorchestra-3.0.3-extreme-performance`
- Final build commit: `8e3312685dabd947f239904b50613272dbda36c1`
- GitHub Actions run: `34049578279`
- Workflow artifact ID: `9994182448`

## Final artifact

- `faunaandorchestra-forge-1.20.1-3.0.3-extreme-optimized.jar`
- Size: `50,002,183` bytes
- SHA-256: `acb1692f7d76d037a31a5d15507ca1c939993e7d9e50cb5267cb6b48e01340f3`
- Forge clean reobfuscated build: PASS
- ZIP integrity: PASS

## Package/content parity against the untouched uploaded JAR

- Total JAR entries: 1,914 -> 1,914
- Classes: 585 -> 585
- `assets/` + `data/` entry names including directories: 1,267 -> 1,267
- Missing entry names: 0
- Added entry names: 0
- Missing assets/data files: 0
- Added assets/data files: 0
- Byte-different common assets/data files: 0
- `META-INF/mods.toml`: byte-identical

No packaged content, resource, model, texture, animation, sound, recipe, loot/data content, or registry file was removed or changed.

## Round-two extreme optimizations

Wave 2A:

- Living Music's 80-tick death/trap countdown no longer dirties `SynchedEntityData` every server tick; the same persisted NBT key, 80->0 timing, and trapped sentinel are retained.
- Active frog choir keeps the exact 45-block living-player condition but avoids an entity-section query/list allocation every goal tick.
- Great Composer listener tracking reuses scratch/listener lists rather than allocating three player lists every active server tick.
- Orchestra listener tracking reuses player buffers and lazily creates one musician-UUID snapshot only if a membership-change packet actually needs it.
- Musician goal shutdown no longer has a null-conductor crash edge case.
- Anya Ghost skips redundant synchronized UUID writes when the nearest player has not changed.
- Listener Container reads the above block state once per tick instead of twice.

Zero-regression Floating Blossom safe wave:

- Server-authoritative flower-position candidate construction is no longer duplicated on clients; init animation still runs on both sides.
- Circular geometry uses exact integer squared distance instead of repeated `Math.pow`/`Math.sqrt`.
- Flower tag aggregation preserves the original tall-then-small ordering while eliminating intermediate lists/copies.
- A dead entity-count traversal is removed while preserving the exact conditional `random.nextInt(4)` consumption and all entity-push behavior.

## Production-bytecode challenge

Verified in the final reobfuscated JAR:

- `LivingMusicEntity`: synced `TICKS_UNTIL_DEATH` accessor removed; plain `int ticksUntilDeath` is present.
- `QuirkyFrogConductingChoirGoal.tick`: radius entity query count 1 -> 0.
- `TheGreatComposer.tick`: new `ArrayList` allocations 3 -> 0; reusable scratch list present.
- `ConductorEntityConductingOrchestraGoal.tick`: new `ArrayList` allocations 2 -> 0; Java stream references 12 -> 0; reusable player collector and lazy UUID snapshot helpers present.
- `AnyaGhost.tick`: UUID equality guard is present before setter.
- `ListenerContainerBlockEntity.tick`: repeated block-state read count 2 -> 1.
- `FloatingBlossomEntity`: `Math.pow`/`Math.sqrt` removed; tag lambdas no longer return intermediate lists.
- All six first-pass animated block renderers still return `false` from `shouldRenderOffScreen` and all six render AABBs remain conservative +/-3 blocks.
- First-pass Faust/Dan B 60-tick reset, single-pass SoundEngine orchestra check, frog type-first global event check, and no-`betweenClosedStream` orchestra listener walk are preserved.

## Deliberately withheld without native visual proof

The following were not applied because they cross the no-regression boundary without an actual Minecraft client visual/runtime test:

- Floating Blossom lifetime conversion from synced metadata to spawn-data/local countdown.
- Wandering Note lifetime conversion from synced metadata to spawn-data/local countdown.
- Moving server-broadcast particles to client-local generation where exact multiplayer distribution/timing was not proven.

No spawn caps, entity caps, reduced ranges, reduced particle counts, reduced flower counts, reduced animation quality, altered music timing, longer AI cooldowns, or content cuts were introduced.

## Verification boundary

This verifies exact source lineage, successful Forge reobfuscated compilation, package integrity/content parity, and production-bytecode presence of the intended performance changes. A native Minecraft client/modpack benchmark was not executed in this environment, so this receipt does not claim a numeric FPS/TPS uplift. The next performance frontier should be driven by a Spark/client profile from the user's actual pack rather than blind semantic changes.
