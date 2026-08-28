# Rogues 3.1.1 -> Native Forge 1.20.1 Port Contract

Status: **ISOLATED PREP ONLY**. This lane must not replace the active Paladins authority until Paladins graduates.

## Immutable source authority

Current feature/content authority:
- Repository: `ZsoltMolnarrr/Rogues`
- Version: `3.1.1`
- Minecraft: `1.21.1`
- Commit: `d4a7af565559dcff4384eabb2481f63eb5f97d55`
- Tree: `c6fac5d7c80807b41668274843c9732762d7cb03`
- Yarn: `1.21.1+build.3`

Historical target-version substrate only:
- Repository: `ZsoltMolnarrr/Rogues`
- Version: `1.2.0`
- Minecraft: `1.20.1`
- Commit: `bdfe6447b90758129e12430b497d97c181222b12`
- Tree: `0a4f7f94e77031732843f8a40a8460184bd3577a`
- Yarn: `1.20.1+build.10`

The historical branch is an API/mapping/semantic anchor only. Never substitute its reduced 1.2.0 feature set for current 3.1.1 content or behavior simply to compile.

## Target runtime

- Minecraft `1.20.1`
- Forge `47.4.23`
- Yarn compatibility anchor `1.20.1+build.10`
- Java 17 / owned classes major 61
- Native Forge lifecycle/registries; no NeoForge or Fabric runtime dependency leakage

## Current 3.1.1 dependency authority

Pinned current upstream declarations include:
- TinyConfig `3.1.0`
- Cloth Config `15.0.130`
- Player Animator `2.0.1+1.21.1`
- Armor Model API `1.0.0+1.21.1`
- Structure Pool API `1.2.0+1.21.1`
- Spell Power `1.6.0+1.21.1`
- Spell Engine `1.10.0+1.21.1`
- Curios `9.5.1+1.21.1` on NeoForge

The port must consume the project's already-graduated/current Forge 1.20.1 foundations as separate packaged JARs where applicable. Do not shade, source-inject, fabricate Maven coordinates, or silently revive the historical TinyConfig/Spell Engine/Spell Power stack.

## Port rules

1. Materialize exact current 3.1.1 Java, resources, and generated data from the current authority pin.
2. Use exact historical 1.20.1 source only to resolve target signatures, mappings, or behavior that still applies.
3. Translate registration to native Forge `RegisterEvent` ownership using the proven RPG-Series `RegistrationBridge` pattern when common code must remain loader-neutral.
4. Preserve modern equipment, weapons, skills/effects, entities, professions/trades, models, animations, sounds, loot/data, and optional integrations wherever technically applicable.
5. Treat post-1.20.1 data components/registry holders/render APIs as semantic translations, not reasons to delete features.
6. Require separate real foundation JARs at common/Forge ABI boundaries; reject dependency class leakage from the Rogues release.
7. Any target-native Forge behavior replacing a Fabric/NeoForge-origin mixin must be proven equivalent and must not duplicate Forge behavior.

## Graduation contract

Rogues cannot graduate on compilation alone. The final lane must prove, on one exact head:
- immutable two-pin provenance and deterministic materialization;
- current-source common + native Forge compilation;
- remapped release packaging and Java-17 bytecode gates;
- dependency anti-shading/leakage checks;
- cross-run deterministic/certified release and deterministic source archive;
- real Forge dedicated dev-server readiness;
- real Forge client LWJGL/resource/render bootstrap;
- fresh official Forge `47.4.23` packaged-server readiness with the exact release JAR and separate dependencies;
- representative game-thread semantic tests for Rogues-specific current behavior;
- real-player tests for behavior that cannot honestly be proven with synthetic entities;
- GitHub evidence plus Google Drive persistence/round-trip verification before `GRADUATED`.

## Promotion rule

This clean prep branch starts from Paladins canonical `038ff97ebe069133152ba9da552ae4fb509646d4`. Paladins has since advanced only through QA-owned commits while its deep acceptance remains authoritative. Rogues work stays isolated until Paladins graduation; then compare against the final graduated Paladins head and transplant/fast-forward only the Rogues-owned coherent delta. Never force over newer canonical evidence.
