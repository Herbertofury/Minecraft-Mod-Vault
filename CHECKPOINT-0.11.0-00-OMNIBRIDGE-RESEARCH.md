# Minecraft Mod Vault 0.11.0 — OmniBridge Conversion Studio checkpoint 00

Recorded: 2026-08-21 UTC
Canonical source: verified Minecraft Mod Vault 0.10.0 release source
Local baseline commit: `0e00f73a4340900a905984c56661b41d2d3a9e37`
Target release: **0.11.0 OmniBridge Conversion Studio**

## User outcome

Build a premium, truthful conversion workbench that can accept Java mods/source projects, Java data/resource packs, Bedrock add-ons, behavior/resource packs, scripts, worlds, world templates and structure formats; translate everything that has a deterministic representation; generate a complete target workspace and proof bundle for everything else; preserve originals; and connect difficult semantic work to Repair Lab, Porting Lab and reviewed external tool adapters instead of pretending that a metadata rewrite is a complete conversion.

## Architectural decision

The feature uses a versioned **Universal Minecraft Content Graph (UMCG)** between import and export. Importers normalize identity, assets, recipes, loot, functions, tags, blocks, items, entities, scripts, structures, world links and pack metadata. Target emitters consume that graph for Java datapacks/resource packs/mod scaffolds and Bedrock behavior/resource packs/add-ons/world templates. Every node carries source path, source hash, target mapping, confidence, loss notes and verification state.

Conversion levels are explicit:

1. `exact` — copied or deterministically represented without semantic loss.
2. `translated` — schema/command/path translation with validation.
3. `generated` — target-native scaffold or compatibility adapter generated from evidence.
4. `tool-assisted` — delegated to a detected, version-pinned external converter with captured logs/hashes.
5. `review` — human semantic work remains; generated workspace contains the exact unresolved node and evidence.
6. `blocked` — a safety, rights or target-engine constraint prevents automatic emission.

The application will never label a conversion complete when unresolved review or blocked nodes remain.

## Primary sources and tools reviewed

- Microsoft Minecraft Creator documentation and Mojang `bedrock-samples`: authoritative Bedrock manifests, behavior/resource packs, Script API, world-template layout and current sample schemas.
- HiveGamesOSS Chunker: Java/Bedrock world conversion plus game-version upgrade/downgrade.
- Amulet Map Editor/Core: cross-edition world read/write and structured translation.
- `kbinani/je2be-core`: Java/Bedrock/console world conversion core.
- GeyserMC PackConverter: Java resource-pack to Bedrock conversion.
- GeyserMC Rainbow: generated Geyser block/item/sound mappings and Bedrock resource packs for custom Java content.
- Mojang DataFixerUpper: incremental Java game-data transformations.
- Fabric Loom and Tiny Remapper: decompilation/remapping and mapping-aware Java mod workspaces.
- Stonecutter, Modstitch, Unimined and Architectury: multi-version/multi-loader Java project generation.
- Bedrock-OSS Regolith and bridge.: Bedrock project compilation/editing ecosystems.
- mcbeet/beet and mecha: Java data/resource-pack construction and function validation.
- PortKit and ModMorpher: experimental Java-mod-to-Bedrock pipelines used as optional evidence/adapters, never as unquestioned authority.
- PiTheGuy/SchemConvert and related structure tools: `.nbt`, `.schem`, `.litematic` and blueprint translation.

## Release acceptance contract

- Import `.jar`, source `.zip`, Java pack `.zip`, `.mcpack`, `.mcaddon`, `.mcworld`, `.mctemplate`, unpacked-compatible archives and common structure files.
- Detect edition, package type, loader, versions, manifests, scripts, world/template structure and content inventory.
- Build a persistent conversion session with immutable source hash and reversible working directory.
- Offer target formats: Bedrock add-on/resource/behavior/world/template; Java datapack/resource pack/world; Fabric/NeoForge/Forge/multiloader source workspace.
- Convert common assets, localization, functions, recipes, loot, tags, sounds, pack metadata and world-template packaging through real emitters.
- Generate Java mod and Bedrock script scaffolds for content that cannot live in a vanilla pack.
- Produce explicit coverage, loss analysis, unresolved-node review queue and target-version compatibility plan.
- Integrate detected world/resource/pack converters as optional adapters with fixed commands, timeout, logs and output hashing.
- Generate a deterministic proof bundle containing source identity, graph, plan, mappings, outputs, logs, warnings and hashes.
- Premium production UI with real import, analyze, plan, convert, download, open, reset and handoff controls.
- No image generation, no fake controls, no unsupported “100% converted” claim.
- Full Go/JavaScript tests, fresh package build, real browser flow, restart persistence, archive verification and Drive/GitHub publication before release.
