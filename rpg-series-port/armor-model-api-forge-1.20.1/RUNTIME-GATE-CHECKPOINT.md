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
- Run #111 cleared that API batch and exposed the 1.20.1 armor-glint/model-render signatures; passes 2/3 fixed those without changing rendering intent.
- Run #114 proved the common module and production Forge transform, then isolated the Forge reload-listener signature mismatch.
- `913af0d360c31045a7994871e4c71fb252f2ffcc` fixes the Forge 1.20.1 client reload listener API.
- `b6752d6e8cc852f0afa6f21cff70939138b58e4b` retargets the armor takeover to the exact six-argument Forge/Yarn 1.20.1 `renderArmor` method and strips Fabric-only environment annotations.
- Final dev-runtime acceptance run #116 / `32996732628` is green at exact head `b6752d6e8cc852f0afa6f21cff70939138b58e4b`:
  - clean common + Forge compile, transform, remap and package passed;
  - release JAR metadata/mixin/package integrity passed;
  - dedicated Forge development server reached `Done (...)!`;
  - headless Forge client reached post-bootstrap runtime;
  - candidate JAR SHA-256 `3957bb845cbef279a7dda90196de894b81d8a894a44ed14d6dd3a0d74ddff78d`;
  - candidate source ZIP SHA-256 `cc6cadd76a8ea8b65c779c976eb588da346498484f50a12f6d0c2d6f1eb1c5a9`;
  - evidence artifact ID `9617160688`, ZIP SHA-256 `837db8dcadbecda3326c74560d46515ba414f1436c27956ad5331ac7de1c0382`.
- `97e5bd303ee1570e73119c0b67999c60bf88caec` adds the final graduation gate:
  - Java 17 class major version 61 verification;
  - a brand-new Forge 47.4.23 server installation;
  - exactly one mod JAR installed: the built Armor Model API release JAR;
  - byte-for-byte equality between the built and installed JAR;
  - fresh packaged server must reach `Done (...)!` with no loader, classloading, mixin, registry, dependency, or server-tick fatal errors.
- Armor Model API remains **ungraduated** until the exact-head verifier containing that fresh packaged-runtime gate passes and the final artifact is Drive round-trip verified, Repair Brain is updated, and PR evidence is sealed.

## Offline fallback

The canonical Minecraft Dev Kit on Drive contains JDK 17, Gradle 8.8/8.14, and the OmniPorter Forge 1.20.1 / Forge 47.4.23 Gradle cache (full and split copies). If a restricted local runtime is blocked only by DNS/network access, use that cached substrate instead of weakening the acceptance gate.

## Next progression

Armor Model API -> Wizards current -> Archers / Paladins / Rogues -> full packaged-suite client/server integration.
