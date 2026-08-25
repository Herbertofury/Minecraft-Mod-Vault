# OmniPorter

OmniPorter is the Minecraft Mod Vault conversion engine for source and JAR-only Minecraft mods/add-ons.

## Goal

Upgrade, downgrade, backport, cross-loader port, and eventually Java ↔ Bedrock semantic conversion while preserving the original mod's IDs, assets, gameplay behavior, data, persistence, networking, rendering, Mixins/access intent, and dependencies wherever the target can represent them.

## Architecture

- **Rust control plane (planned):** fast hashing, archive indexing, dependency graph, mapping graph, rule scheduling, parallel analysis, cache/CAS, report generation and deterministic orchestration.
- **JVM workers:** ASM, Mapping-IO, Tiny Remapper, Mixin-aware analysis, Java source transforms, Gradle/Minecraft build integration.
- **Target adapters:** Forge, NeoForge, Fabric, Quilt/legacy; Bedrock BP/RP/Script API later through Behavior IR.
- **Knowledge base:** version/API transforms with provenance, confidence, source/target constraints and regression fixtures.
- **Verifier:** clean build + production JAR + client/server/runtime behavior comparison. Compile success is never the final verdict.

## Pipeline

`intake → fingerprint → source/target environment model → mapping graph → semantic port plan → source/bytecode transforms → dependency substitution → build-error feedback loop → package → differential verification → learned recipe`

## v0 foundation

`workers/jvm-inspector` is the first verified worker. It reads a mod JAR without executing it and reports SHA-256, loader metadata signals, class-file versions, Mixin configs, access-transform formats, nested JARs and signing metadata. The output becomes evidence for the Port Manifest.

## First torture-test family

RPG Series / Spell Engine ecosystem. This is intentionally difficult because the modern repositories are multi-platform and heavily data-driven, with API evolution, item components, registries, networking, animations, equipment abstractions and content mods built around Spell Engine.
