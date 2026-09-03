# Entity Render Guard 1.0.0 - Verified Release Checkpoint

Date: 2026-09-03

## Target baseline

- Minecraft 1.20.1
- Noxviola runtime: Forge 47.4.20, Java 17
- Native QA: Forge 47.4.23, Temurin Java 17.0.20.1
- Accelerated Rendering target: 1.0.14-1.20.1-alpha
- Ars Nouveau target: 4.12.7
- Initial guarded entity: `ars_nouveau:whirlisprig`

## Root cause being guarded

The supplied Noxviola crash reaches the world and fails on the render thread while rendering `ars_nouveau:whirlisprig`. The call passes through Accelerated Rendering's entity hook. Entity Render Guard keeps Accelerated Rendering enabled globally but routes configured incompatible entity types through its vanilla entity/item/text pipelines only for the duration of that entity render.

## Static and package gates

- ForgeGradle 6 offline `clean jar reobfJar` on Java 17 / Forge 47.4.23: PASS.
- Main source Java 17 compilation: PASS.
- Real uploaded `latest.log` parsing -> `ars_nouveau:whirlisprig`: PASS.
- Exact/glob/regex matching and invalid-regex handling: PASS.
- Nested render-scope push/pop: PASS.
- Crash auto-learning, malformed UTF-8 tolerance, config persistence, idempotent rescan: PASS.
- Production SRG `m_109517_` render target and mapped `renderEntity` descriptor compatibility: PASS.
- JAR ZIP integrity and `MixinConfigs: entityrenderguard.mixins.json`: PASS.
- Fresh extraction of delivered source archive rebuilt offline: PASS.

## Native production-JAR QA

A source-less Forge 1.20.1 client harness loaded the final reobfuscated Entity Render Guard JAR from `run/mods` plus a QA-only contract-equivalent Accelerated Rendering provider. The test config guarded `minecraft:player` solely to guarantee a deterministic render subject.

- Real integrated singleplayer world load: PASS.
- Packaged Mixin applied to `LevelRenderer`: PASS.
- Guarded entity render triggered vanilla fallback: PASS.
- Accelerated Rendering bridge connected: PASS.
- ENTITY push/pop: 593 / 593.
- ITEM push/pop: 593 / 593.
- TEXT push/pop: 593 / 593.
- Every observed pop returned to depth 0; underflow count: 0.
- World save + integrated-server shutdown: PASS.
- Client closeout: `BUILD SUCCESSFUL`.

The exact user's Accelerated Rendering + Ars Nouveau binary pair was not available in Drive, so the native branch test uses contract-equivalent AR pipeline methods. The original Noxviola crash evidence remains the basis for the seeded Whirlisprig rule.

## Release artifact

- File: `entity-render-guard-1.0.0-forge-1.20.1.jar`
- Size: 24,263 bytes
- SHA-256: `d52dea39ef3aefb806c1b454d0b0ca98cde40ae0b9e55904fedc5a2d91a76420`
- Source ZIP SHA-256: `fcc8e9b7d8c168810de8e518311b5dca8ff24fd6b1cf95dfa760dc4789ae2a19`
- Native QA log SHA-256: `6fb234e39d3417645c8063c7b04512fa3ae1c7bb94cbc8d8bb27f7a016991ef9`
