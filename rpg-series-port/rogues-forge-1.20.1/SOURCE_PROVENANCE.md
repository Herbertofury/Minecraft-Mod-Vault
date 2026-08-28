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
- `common/src` tree: `ba3110c62b9901f9f6a729a0c660e0ea10b72d91`
- Yarn declaration: `1.21.1+build.3`

All intended Rogues 3.1.1 content and behavior originate here unless a target-native translation is required by Minecraft/Forge 1.20.1.

## Historical Minecraft 1.20.1 substrate

Repository: `ZsoltMolnarrr/Rogues`

- Version: `1.2.0`
- Minecraft: `1.20.1`
- Commit: `bdfe6447b90758129e12430b497d97c181222b12`
- Root tree: `0a4f7f94e77031732843f8a40a8460184bd3577a`
- `gradle.properties` blob: `1d1e4007bf335129cc6cabb115a0f78d3805fca0`
- `common` tree: `294546c95c33e2f97b79325e21d0434c4ce77a02`
- `forge` tree: `6f31c256dc901a247be108488509206079317734b`
- `gradle` tree: `54519e011fbca16ca7415fcea30fedda8b78e0cb`
- `gradlew` blob: `555c706d8842a1d4ed184d7eeb6783a0f770b3f2`
- `gradlew.bat` blob: `f955316af440d3ec5ea49b27d41c0c1257084042`
- `settings.gradle` blob: `ed288563467414e15d9c7e6db8f2dc24b8d2ed96`
- Yarn declaration: `1.20.1+build.10`

The two `gradle.properties` blob IDs above were re-verified directly from the immutable commits before this clean prep was restored. The current file also contains properties with optional whitespace around `=`, so source-prep validation must parse property keys/values rather than require one exact spacing style.

This historical tree is an API, mapping, and target-semantics substrate only. It must never replace the newer 3.1.1 feature set merely because old code compiles more easily.

## Materialization contract

A source-preparation step is valid only when it:

1. resolves the exact immutable commits above, never an unpinned branch or tag;
2. verifies the resolved root tree before consuming source;
3. verifies the exact `gradle.properties` Git blob and required version/Minecraft/Yarn values;
4. tolerates harmless whitespace around Gradle property separators without weakening value checks;
5. keeps the current 3.1.1 tree and historical 1.20.1 tree physically distinct;
6. places historical source under an `.upstream`/reference-only location and never overlays it wholesale onto generated current source;
7. records deterministic provenance for every materialized authority;
8. fails if any immutable pin changes unexpectedly;
9. excludes Git metadata, caches, build outputs, generated runtime worlds/logs, and mutable download metadata from deterministic source packages.

## Semantic authority rule

When current and historical implementations differ, preserve the current 3.1.1 intent and translate it to target-native Forge 1.20.1 semantics. Historical code may answer questions such as target method signatures, old registry ownership, or mapping names, but cannot silently delete or downgrade current features.

High-value semantics already identified for explicit preservation include:

- `SHOCK` as full Spell Engine `STUN`;
- `BEAR_TRAP` and `NET_TRAP` as custom `ROOT`, not STUN: movement/jumping blocked while attacks, item use, casting, and mob actions remain allowed;
- NET_TRAP knockback immunity;
- STEALTH visibility/tint and removal triggers;
- CHARGE cadence;
- LAST_STAND current fortification behavior;
- arms-merchant POI/profession and all current trade tiers/economics;
- configured vanilla Strength modifier rebalance.

## Runtime authority

These hashes establish source identity only. They do not establish a successful port. Graduation still requires exact-head compile/package gates plus real Forge dev-server, native client, fresh official packaged-server, and representative game-thread / real-player behavioral evidence as defined in `PORT_CONTRACT.md`.
