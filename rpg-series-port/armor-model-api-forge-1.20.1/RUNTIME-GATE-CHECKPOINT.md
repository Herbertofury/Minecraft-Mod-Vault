# Armor Model API 1.0.0 — Forge 1.20.1 Runtime Gate

Checkpoint: 2026-08-26

- Upstream pin: `FabricExtras/ArmorModelAPI@a664155a0aab3161cd7e4bf0c1f72512b4ec4949`
- Target: Minecraft 1.20.1 / Forge 47.4.23 / Java 17
- Generator repair: `99f206cf4751ed1677e1e58ae42acccc0ef52e14`
- Runtime-verifier expansion: `1c5a476824c3b6de4a542391ff40f1c76706f956`
- Runs #106 and #107 were source-hygiene failures before Gradle, caused by leaked NeoForge `FMLLoader` / `LoadingModList` imports.
- Run #110 reached the real common-source compile and exposed the first 1.21 API batch: `DyedColorComponent`, `DataComponentTypes.TRIM`, and the optional `IrisApi` compile surface.
- Compatibility pass 1: `a1bf23705661e6bcc78dfa698f3596f65c486bb5`.
  - Dye color now uses native 1.20.1 `DyeableItem#getColor` semantics.
  - Armor trims now use `ArmorTrim.getTrim(registryManager, stack)` and the 1.20.1 armor-trim render layer.
  - Shader-pack awareness is optional and guarded for Iris/Oculus without creating a mandatory shader dependency.
  - A hard preparation check rejects any remaining `net.minecraft.component` references.
- Generator rewrites the NeoForge loader imports, mixin package, and `JAVA_21` -> `JAVA_17`; release packaging advertises `MixinConfigs` and keeps client initialization behind a side-safe helper.
- Active verifier requires clean build/reobf, release-JAR integrity and loader hygiene, dedicated-server ready state, and headless Forge client bootstrap.
- Armor Model API remains **ungraduated** until those runtime gates plus fresh packaged-runtime and artifact/persistence checks pass.

## Offline fallback

The canonical Minecraft Dev Kit on Drive contains JDK 17, Gradle 8.8/8.14, and the OmniPorter Forge 1.20.1 / Forge 47.4.23 Gradle cache (full and split copies). If a restricted local runtime is blocked only by DNS/network access, use that cached substrate instead of weakening the acceptance gate.

## Next progression

Armor Model API -> Wizards current -> Archers / Paladins / Rogues -> full packaged-suite client/server integration.
