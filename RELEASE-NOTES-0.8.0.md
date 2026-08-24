# Minecraft Mod Vault 0.8.0 release notes

Minecraft Mod Vault 0.8.0 adds **Mod Doctor**, a local-first artifact, dependency, crash, repair, and porting intelligence system built on a current 154-source research catalog, 92 detailed execution records, 163 unique runtime tool cards, and 9 reusable repair patterns.

## Mod Doctor

- Inspects installed JARs without trusting filenames.
- Records SHA-1, SHA-256, SHA-512, CurseForge fingerprints, loader metadata, mod IDs, versions, dependencies, source/homepage provenance, Java class level, mappings, mixins, plugins, refmaps, access wideners/transformers, coremods, transformation services, nested JARs, signatures, native libraries, Kotlin/Scala usage, side references, data/assets, and pack metadata.
- Detects exact duplicate JARs, duplicate mod IDs, exact duplicate classes, conflicting duplicate classes, missing required dependencies, declared conflicts, and required-dependency cycles.
- Produces deterministic dependency-first deployment order.
- Distinguishes source rebuild, namespace remap, source transform, narrow binary repair, compatibility adapter, runtime bridge, metadata/data migration, and review-only substitution routes.
- Models Minecraft-to-Java compatibility, including Java 8, 16, 17, 21, and the Java 25 requirement for the unobfuscated 26.x generation.
- Analyzes pasted crash reports, `latest.log`, `debug.log`, and build failures, prioritizing causal linkage errors, dependency failures, Mixin errors, side violations, registry/data bootstrap failures, native/graphics failures, and resource exhaustion.
- Imports durable Repair Brain patterns for API owner/descriptor drift, NBT namespace/schema repair, Mixin annotation contracts, loader metadata limitations, adapter design, artifact verification, and supersession.
- Routes legacy source recovery and target builds through version-appropriate toolchains instead of applying modern Loom/Forge/NeoForge assumptions to 1.7.10, 1.12.2, beta/classic, Legacy Fabric, or Ornithe projects.
- Separates historical source reconstruction, semantic source/data porting, and runtime compatibility layers such as Cleanroom, Fugue, LWJGL3ify, UniMixins, LegacyFix, or Retromod.
- Validates every generated plan source ID against the embedded catalog across legacy and modern migration fixtures.

## Research and tooling

- Ships the machine-readable catalog in `assets/mod-doctor-knowledge.json`.
- Ships detailed tool records in `knowledge/doctor-tools.json`.
- Ships reusable repair patterns in `assets/mod-doctor-repair-patterns.json` and `knowledge/repair-patterns.json`.
- Ships the complete human-readable guide `ULTIMATE-TOOLS-AND-PORTING-KNOWLEDGE.md` with direct official and repository links.
- Covers current loaders, mappings, remappers, decompilers, API diffing, bytecode libraries/editors, source transformation, multi-loader/version tooling, compatibility bridges, provider APIs, modpack updaters, crash analyzers, test harnesses, profilers, world/NBT repair, protocol tools, publishing, and official specifications.

## Safety and verification model

- Automatic actions require strong identity and target compatibility evidence.
- Ambiguous ports, forks, substitutions, runtime bridges, and dependency-changing candidates remain review-first.
- Every mutation is intended to run through staging, backup, hash verification, JAR inspection, dependency validation, target build, real runtime launch, fresh-log inspection, and rollback on failure.
- A successful remap, patch, build, or startup is intermediate evidence, not completion.

## Preserved systems

The 0.7.1 federated 28-provider browser, universal updater, provider-aware installation, living recommendations, 13-channel Creator Archive, transcription pipeline, CIT/furniture/Bedrock workspaces, and local loopback desktop architecture remain present.
