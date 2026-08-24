# Forge 1.20.1 Ragdoll Port - Premium.2 Parity Checkpoint

Date: 2026-08-23
Target: Minecraft 1.20.1 / Forge 47.x / Java 17

## Artifacts
- `sable_player_ragdoll-1.20.1-forge-0.7.2-premium.2.jar`
  - SHA-256: `9c3475587ff63bcc974418a43461b907cbbae6b243380e5a9eeece92dae2b922`
  - Drive file ID: `19wj0RyU6TotAkFsJa7Nl6vb5HjIWclN8`
- `ragdoll_reactions-1.20.1-forge-0.7.0-premium.2.jar`
  - SHA-256: `5c5f7ef99b8bfc0fcd4654b53a7fbb480fbc202877db203adf1db1d95b32d5ae`
  - Drive file ID: `1Ia22lqQktTewB6TTkMr1lQ47UaAXYIuS`
- `ragdoll-forge-1.20.1-premium.2-BUNDLE.zip`
  - SHA-256: `cf49e569ca06cc9ba2b74231fbc8a89f46d5177062ee9a2812404069740c3de2`
  - Drive file ID: `1-fxslv8RgG3IsWLYdgKqLnQwkpGPGlQJ`
- `ragdoll-forge-1.20.1-premium.2-FULL-SOURCE.zip`
  - SHA-256: `c8a572e3d2630c417b1965f1c4d363c18bfb7ab4d61799d7a9782056be759b42`
  - Drive file ID: `19lqOJNGJ9LjFSFUhCOdy2wH3DqT2qos0`

Canonical Drive folder: `Ragdoll Forge 1.20.1 Backport` (`1T7YP1nKcJtWJjsNXvDmRfMpA4asfJIQt`). The full source/workstate is persisted privately on Drive; the public GitHub branch intentionally keeps only the verification/checkpoint state rather than republishing ARR-derived Reactions source.

## Premium parity definition
Behavior-first port of the supplied NeoForge 1.21.1 binaries. Preserve their user-facing resources, addon-facing API/defaults and Ragdoll Reactions behavior while replacing only the Minecraft/loader/physics seams needed for Forge 1.20.1.

## Verification completed
- Full ZIP integrity PASS for both JARs and bundle.
- Core: 59/59 production classes are Java 17 classfile major 61.
- Reactions: 48/48 production classes are Java 17 classfile major 61.
- Proper Forge `META-INF/mods.toml` present after release packaging gate.
- Forge `@Mod` annotations verified for `sable_player_ragdoll` and `ragdoll_reactions`.
- No bundled `net.minecraft`/Forge compile stubs.
- No NeoForge entries or bytecode/string references found.
- Core supplied resource files: 16/16 byte-identical, with only required 1.21 `tags/item` -> 1.20.1 `tags/items` path mapping.
- Reactions supplied resource files: 4/4 byte-identical.
- Tracked core addon-facing public ABI: zero missing public members after loader-event normalization.
- Reactions public behavior ABI: only raw mismatch is expected NeoForge IEventBus -> Forge IEventBus.
- Packaged-artifact parity harness PASS: player launch/release, mob launch/release, playerless/corpse session, velocity bridge, dismount lock, wailing call path, original defaults, 128 m/s clamp, heavy-hit reaction routing and 70% client-motion telemetry threshold.

## Behavior restored in Premium.2
- Player, mob and playerless ragdoll sessions.
- Limb pose/stiffness options, wailing and dismount locking.
- Tagged-item/manual trigger compatibility surface.
- Original reaction defaults and cooldowns.
- Hit, fall, elytra wall crash, lightning, vanilla/Create Big Cannons explosions and sudden-impact reactions.
- Mob hit/fall/explosion/impact reactions, original 30% heavy-hit chance, random angular tumble and settle/release timing.
- Suppressions for bounce/riptide/elytra/creative flight/Create conveyor/wind-charge cases.
- Forge-native client motion telemetry to preserve the original impact anti-false-positive behavior.

## Runtime backend
The Forge port adapts the original Sable physics seam to Ragdollified for Minecraft 1.20.1. Current acceptance target is Ragdollified 1.0.0.

## Remaining acceptance gate
Do not mark user-runtime-confirmed until these exact hashes have launched and had the ragdoll trigger matrix exercised inside the user's real Forge 1.20.1 instance with Ragdollified 1.0.0. Static/package/harness verification is green; exact real-instance gameplay validation is the remaining gate.
