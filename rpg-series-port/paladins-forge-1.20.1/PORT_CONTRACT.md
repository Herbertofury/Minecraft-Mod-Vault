# Paladins 3.1.1 -> Native Forge 1.20.1 Port Contract

Status: PREP ONLY. Do not activate until Archers 3.1.1 has passed full runtime graduation.

## Feature authority

- Upstream: `ZsoltMolnarrr/Paladins`
- Current 1.21.1 commit: `9d5611d3799c56951255fc3e3e61aee4233f3d28`
- Current tree: `be72e53347fa0bbae33515370d374906d192c2d8`
- Current version: `3.1.1`
- Current Minecraft/Yarn: `1.21.1` / `1.21.1+build.3`
- Current loader architecture: Fabric + NeoForge / Architectury
- Java source authority: current 3.1.1. Do not replace current content with historical 1.20.1 content.

## Target mapping substrate

- Historical Paladins 1.20.1 branch commit: `f310c0b12eb3791c6e83f6bda0accdf032aa8a17`
- Historical version: `1.4.0`
- Historical Minecraft/Yarn: `1.20.1` / `1.20.1+build.10`
- Use only for target-era mappings, signatures, assets/data comparisons, and behavioral archaeology where current APIs moved.

## Target runtime

- Minecraft `1.20.1`
- Forge `47.4.23`
- Yarn `1.20.1+build.10`
- Java `17`
- Native Forge release artifact. No Fabric runtime, no NeoForge runtime, no Architectury/Fabric API leakage in the packaged boundary unless an already-graduated shared common ABI requires compile-time annotations only.

## Known current dependency surface

Current 3.1.1 declares:

- Shield API `2.1.0`
- Armor Model API `1.0.0+1.21.1`
- TinyConfig `3.1.0`
- Runes `1.3.1+1.21.1`
- Structure Pool API `1.2.0+1.21.1`
- Spell Power `1.6.0+1.21.1`
- Spell Engine `1.10.0+1.21.1`
- Cloth Config `15.0.130`
- Player Animator `2.0.1+1.21.1`
- Curios `9.5.1+1.21.1` on NeoForge.

Prefer the already-graduated 1.20.1 foundations in this project where their public contract satisfies current Paladins behavior. Port/graduate any genuinely missing foundation before weakening Paladins behavior.

## Shield API substrate

- Source project: `FabricExtras/ShieldAPI`
- Current repository default branch: `1.21.1`.
- Current branch head observed: `3e1f38fe1be03e21a45075cc9fe39bfff7a41296` (repository currently reports Shield API 2.2.0; Paladins itself pins 2.1.0, so exact 2.1.0 release source must be pinned before implementation).
- Exact target-era 1.20.1 branch: `cdcf7ffdcffb31a1dd8c36ba7a27cf312b0e8e71`, Shield API `1.0.1`, Yarn `1.20.1+build.10`.
- Treat current API behavior as authority and 1.20.1 branch as mapping substrate. Do not substitute the historical API wholesale if Paladins 3.1.1 uses later features.

## Acceptance policy

Mirror the Archers discipline:

1. Freeze exact immutable current + historical pins and source manifests.
2. Materialize every current Java/resource/generated asset/data file.
3. Use an explicit deterministic 1.21.1 -> 1.20.1 compatibility transform pipeline.
4. Common compile against separately built named 1.20.1 ABI JARs; Forge loader compile against separate exact Forge artifacts.
5. No dependency shading that changes runtime ownership.
6. Java 17, metadata, leakage, content-count, and deterministic-byte gates.
7. Real Forge dedicated server + headless/native client acceptance with separate dependency JARs.
8. Semantic Paladins runtime self-tests for shields, armor, weapons/spells, effects/entities if present, villagers/POIs/trades if present, configs, Curios integration, recipes/tags/data, and current-only features.
9. Fresh official Forge 47.4.23 packaged-server replay with the exact release JAR.
10. Persist source/evidence/release artifacts to both GitHub and canonical Google Drive before graduation.

Archers graduation remains the gate before this lane becomes canonical.
