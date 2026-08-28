# Rogues Native Forge 1.20.1 - Immutable Source Provenance

This file is a fail-closed provenance ledger for the Rogues port. Moving branch names and tags are never source authority.

## Current feature/content authority

Repository: `ZsoltMolnarrr/Rogues`

- Version: `3.1.1`
- Minecraft: `1.21.1`
- Commit: `d4a7af565559dcff4384eabb2481f63eb5f97d55`
- Root tree: `c6fac5d7c80807b41668274843c9732762d7cb03`
- `gradle.properties` blob: `5f264bbbaea8f43b53badb4eb6af24549b6211cf`
- `common` tree: `6f21ea7f91c0af664fe77b92d9fc136857878d57`
- `common/src` tree: `ab33a47e89a46fc2c7bad26539e6bb3b2e5b95ca`
- `common/src/main` tree: `f889df40388170e9e278fccb8ffbbe0066ee378c`
- generated assets tree: `6d6ef7e3febaf886f308f61805ebb11ab681468b`
- generated data tree: `8260801839ef1e8adcc5642d28acdd1d3421838a`
- current generated Last Stand spell blob: `8c266eeb644b908590b2a72bf9b98cb9cd89e31e`
- Yarn declaration: `1.21.1+build.3`

All intended Rogues 3.1.1 Java, resources, and generated data originate here unless a target-native translation is required by Minecraft/Forge 1.20.1. The current generated tree is source authority, not disposable build output: it contains current spell definitions such as `rogues:last_stand` and must be materialized with the Java/resources tree.

## Historical Minecraft 1.20.1 substrate

Repository: `ZsoltMolnarrr/Rogues`

- Version: `1.2.0`
- Minecraft: `1.20.1`
- Commit: `bdfe6447b90758129e12430b497d97c181222b12`
- Root tree: `0a4f7f94e77031732843f8a40a8460184bd3577a`
- `gradle.properties` blob: `1d1e4007bf335129cc6cabb115a0f78d3805fca0`
- `build.gradle` blob: `b2042ce3a0b8b15418f9c7c4bedef5a1f3d3338e`
- `gradle` tree: `59896c6c521e647462c281c276fb973d5ee6bdc2`
- `gradlew` blob: `aeb74cbb43e3931a2455a838345c3f6b8131aaa2`
- `gradlew.bat` blob: `6689b85beecde676054c39c2408085f41e6be6dc`
- `settings.gradle` blob: `f91a4fe7e1f1240c4ca98d81fd7a3d7cede3efb5`
- `src` tree: `6a6057dfc702860f6ddd2fd49fffee8f47b84469`
- `src/main` tree: `02c0bf29477a997d144d87cec1cb604ebd435241`
- `src/main/java` tree: `ff8de0b53c1fb6dc05b47d052ee8b1cf7fe25c54`
- `src/main/resources` tree: `4d74c108a0f186d547dc1ed0e54b8e58aa81c633`
- Yarn declaration: `1.20.1+build.10`

The historical 1.20.1 authority is a **single-project layout** rooted at `src/main`, not a `common`/`forge` multi-module tree. Earlier prep notes that listed historical `common` and `forge` trees, plus older wrapper/settings hashes, were stale and are superseded by the immutable root-tree enumeration above.

The two `gradle.properties` blob IDs above were re-verified directly from the immutable commits. The current file also contains properties with optional whitespace around `=`, so source-prep validation must parse property keys/values rather than require one exact spacing style.

The current `common/src` subtree was also re-derived directly from the immutable commit after an older prep note was found to contain a stale/non-resolving value. The hashes in this ledger are the corrected values and must supersede that older note.

This historical tree is an API, mapping, and target-semantics substrate only. It must never replace the newer 3.1.1 feature set merely because old code compiles more easily.

## Materialization contract

A source-preparation step is valid only when it:

1. resolves the exact immutable commits above, never an unpinned branch or tag;
2. verifies the resolved root tree before consuming source;
3. verifies the exact `gradle.properties` Git blob and required version/Minecraft/Yarn values;
4. tolerates harmless whitespace around Gradle property separators without weakening value checks;
5. verifies the expected **different layouts**: current `common/src/main/...` and historical `src/main/...`;
6. keeps the current 3.1.1 tree and historical 1.20.1 tree physically distinct;
7. materializes current `common/src/main/java`, `common/src/main/resources`, **and** `common/src/main/generated` data rather than treating current generated spell/data definitions as disposable;
8. places historical source under an `.upstream`/reference-only location and never overlays it wholesale onto generated current source;
9. records deterministic provenance for every materialized authority;
10. fails if any immutable pin changes unexpectedly or if the current generated Last Stand definition disappears;
11. excludes Git metadata, caches, build outputs, generated runtime worlds/logs, and mutable download metadata from deterministic source packages.

## Semantic authority rule

When current and historical implementations differ, preserve the current 3.1.1 intent and translate it to target-native Forge 1.20.1 semantics. Historical code may answer questions such as target method signatures, old registry ownership, or mapping names, but cannot silently delete or downgrade current features.

High-value semantics already identified for explicit preservation include:

- `SHOCK` as full Spell Engine `STUN`;
- `BEAR_TRAP` and `NET_TRAP` as custom `ROOT`, not STUN: movement/jumping blocked while attacks, item use, casting, and mob actions remain allowed;
- NET_TRAP knockback immunity;
- STEALTH visibility/tint and removal triggers;
- CHARGE cadence;
- LAST_STAND current five-channel-tick fortification behavior and generated spell definition;
- arms-merchant POI/profession and all current trade tiers/economics;
- configured vanilla Strength modifier rebalance.

## Runtime authority

These hashes establish source identity only. They do not establish a successful port. Graduation still requires exact-head compile/package gates plus real Forge dev-server, native client, fresh official packaged-server, and representative game-thread / real-player behavioral evidence as defined in `PORT_CONTRACT.md` and `RUNTIME_QA_CONTRACT.md`.
