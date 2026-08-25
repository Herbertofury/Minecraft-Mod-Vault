# OmniPorter

OmniPorter is the Minecraft Mod Vault conversion engine for source and JAR-only Minecraft mods/add-ons.

## Goal

Upgrade, downgrade, backport, cross-loader port, and eventually Java ↔ Bedrock semantic conversion while preserving the original mod's IDs, assets, gameplay behavior, data, persistence, networking, rendering, Mixins/access intent, and dependencies wherever the target can represent them.

## Architecture

- **Rust control plane:** fast hashing, archive indexing, provider/source identity, dependency graph, mapping graph, rule scheduling, cache/CAS, report generation and deterministic orchestration.
- **Ferium provider worker:** directory-scale scanning and compatible update discovery across Modrinth, CurseForge and GitHub Releases. Results are candidates until OmniPorter's provenance gate confirms identity.
- **JVM workers:** ASM, Mapping-IO, Tiny Remapper, Mixin-aware analysis, Java source transforms, Gradle/Minecraft build integration.
- **Target adapters:** Forge, NeoForge, Fabric, Quilt/legacy; Bedrock BP/RP/Script API later through Behavior IR.
- **Knowledge base:** version/API transforms with provenance, confidence, source/target constraints and regression fixtures.
- **Verifier:** clean build + production JAR + client/server/runtime behavior comparison. Compile success is never the final verdict.

## Pipeline

`intake → hashes/JAR metadata → cross-provider identity graph → canonical source/tag/commit → source/target environment model → mapping graph → semantic port plan → source/bytecode transforms → dependency substitution → build-error feedback loop → package → differential verification → learned recipe`

## v0.2

`core-rust` is now a verified Rust CLI. It provides:

- Dev Kit diagnostics including Ferium detection.
- SHA-1/SHA-256/SHA-512 artifact fingerprints.
- Exact Modrinth SHA-512 identity lookup.
- Modrinth/GitHub/CurseForge provider resolution.
- An isolated Ferium 4.7.1 scan/update worker that does not pollute a user's normal Ferium config.
- A strict rule that provider matches are evidence, not proof: automatic replacement/porting is blocked until source identity/provenance agrees.

`workers/jvm-inspector` reads a mod JAR without executing it and reports SHA-256, loader metadata signals, class-file versions, Mixin configs, access-transform formats, nested JARs and signing metadata. Its output becomes evidence for the Port Manifest.

## Current verification

Rust v0.2 passes rustfmt, Clippy with warnings denied, and an optimized release build on the Dev Kit nightly toolchain. Ferium 4.7.1 launches and reports its full CLI correctly. The current execution runner blocks Ferium's outbound Modrinth request, so network-provider execution is classified separately from worker/runtime correctness and can fall back to connected web/API routes in chat.

## First torture-test family

RPG Series / Spell Engine ecosystem. This is intentionally difficult because the modern repositories are multi-platform and heavily data-driven, with API evolution, item components, registries, networking, animations, equipment abstractions and content mods built around Spell Engine.
