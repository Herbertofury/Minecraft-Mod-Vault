# OmniBridge 0.11.0 architecture

OmniBridge is Minecraft Mod Vault's conversion control plane for Java Edition and Bedrock Edition content. It does not treat a filename rewrite, metadata edit, archive repack, successful remap, or generated source tree as proof of semantic parity.

## Pipeline

1. **Immutable intake** — copy the source into a session-owned vault, hash the original, extract with traversal/symlink/duplicate/archive-bomb protection, and hash the complete extracted tree.
2. **Universal Minecraft Content Graph (UMCG)** — normalize assets, data, scripts, Java source/bytecode, loader metadata, Bedrock manifests, worlds, templates, structures, dependencies, and relationships without deleting edition-specific information.
3. **Target classification** — every node is classified as exact, translated, generated, tool-assisted, review-required, or blocked for the selected edition/version/loader/package type.
4. **Deterministic emitters** — generate real data/resource packs, `.mcpack`, `.mcaddon`, `.mcworld`, `.mctemplate`, Java loader projects, Java vanilla pack families, editable Bedrock projects, world products, or universal bundles.
5. **Specialist adapters** — run only reviewed allowlisted command contracts in session workspaces. Current executable adapters are Chunker, Geyser PackConverter, JE2BE Resource Pack Converter, and Regolith.
6. **Proof** — hash every output, validate archive paths and package metadata, preserve adapter logs, re-hash the immutable source, and export the graph, plan, review contracts, report, and checksums.

## Universal content graph

The graph records identity and source path for textures, models, sounds, languages, fonts, shaders, particles, recipes, functions, loot, tags, advancements, predicates, item modifiers, blocks, items, entities, animations, animation controllers, render controllers, attachables, spawn rules, features, feature rules, biomes, dimensions, world generation, fog, materials, UI, camera presets, dialogue, trade tables, volumes, Bedrock scripts, Java source, JVM bytecode, Mixins, structures, world data, manifests, and pack metadata.

The graph is an intermediate representation and evidence ledger, not a claim that Java and Bedrock expose identical runtime semantics.

## Product lanes

### Bedrock

- Bedrock add-on family (`.mcaddon`)
- Behavior pack (`.mcpack`)
- Resource pack (`.mcpack`)
- Bedrock world (`.mcworld`)
- Bedrock world template (`.mctemplate`)
- Editable Regolith-compatible behavior/resource/script project
- World product containing template, companion add-on, source project, contracts, and adapter evidence

### Java

- Data pack ZIP
- Resource pack ZIP
- Java vanilla add-on family containing paired data/resource packs and optional world-template source
- Java world ZIP
- Fabric source project
- NeoForge source project
- Forge source project
- Fabric/NeoForge/Forge multi-loader source project
- World-template mod source with immutable embedded world, safe exporter, feature matrix, and loader integration contracts

### Universal

A universal release bundle includes the UMCG, a Bedrock source project/product lane, a Java vanilla pack family, a multi-loader Java source lane, contracts, checksums, and a proof report.

## Target-native Java feature matrix

Bedrock capabilities are not collapsed into a generic TODO. Generated Java projects include a machine-readable and compiled-source feature inventory that maps each source feature to a concrete Java implementation surface and logical side. Examples include:

- blocks/items/entities → registries, data components, attributes, goals, renderers;
- scripts → loader event bus, tick scheduler, commands, networking, persistence;
- spawn rules/biomes/worldgen/dimensions → placements, biome modifiers, configured/placed features, dimension types and generators;
- cameras/dialogue/trades/volumes → camera controllers, menus/screens, merchant offers and region-trigger services;
- animations/render controllers/attachables/UI/fog/materials → client animation and rendering adapters;
- recipes/loot/functions/tags/advancements/predicates → Java data formats and custom serializers/triggers when required.

Original Bedrock source files are preserved inside the generated Java project beside the feature matrix and conversion contracts.

## Version intelligence

Java pack metadata uses the embedded Version Atlas/mcmeta seed when an exact target is known and falls back to reviewed compatibility mappings when it is not. Java loader projects pin the selected Minecraft version, Java toolchain, loader, mappings, and build plugin choices exposed by the Atlas. Bedrock manifests receive a regenerated collision-free identity and the selected minimum engine version.

Version targeting produces a migration plan. It does not imply that custom APIs, commands, data schemas, renderer hooks, Mixins, or game behavior remain compatible without build/runtime verification.

## Adapter execution boundary

- A configured tool path must be an existing regular file, never a symbolic link.
- Command lines are assembled from a fixed adapter-specific allowlist; no shell is used.
- Each run receives an isolated working directory, dedicated HOME/cache locations, a sanitized environment, a timeout, a complete log, output hashing, structural validation, and source re-verification.
- Regolith uses an `exact` export profile whose BP/RP paths remain inside the adapter workspace; OmniBridge never points generated projects at the user's live `com.mojang` tree.
- A successful external process is not accepted without a recognized target artifact.

## Completion boundary

Exact media or same-schema content can be copied and reindexed automatically. Common recipes and localization can be translated deterministically. Manifests, indexes, project structures, and implementation contracts can be generated. Worlds and advanced resources can be delegated to specialist converters.

Arbitrary JVM behavior, Bedrock Script API behavior, renderer internals, complex GUIs, networking, persistence, native code, Mixins, ASM, custom world generation, and edition-specific gameplay systems still require target-native implementation and real client/server/world tests. OmniBridge makes that work organized, attributable, and recoverable instead of pretending it disappeared.
