# Minecraft Mod Vault 0.11.0 — OmniBridge

OmniBridge is a premium universal Minecraft conversion studio built into the same application as OmniManager, Mod Doctor, Porting Lab, Repair Lab, and the Compatibility Brain.

## Universal inputs

- Java mod JARs and source ZIPs
- Fabric, Quilt, Forge, NeoForge, Bukkit, Paper, Velocity and Bungee artifacts
- Java data packs, resource packs, modpacks and worlds
- Bedrock `.mcpack`, `.mcaddon`, `.mcworld`, `.mctemplate`, behavior packs, resource packs, script projects and worlds
- NBT, Sponge schematic, Litematica and blueprint-style structure files

Every input is copied into an immutable conversion session, archive-extracted with hostile-path and bomb limits, hashed at both archive and tree level, and normalized into the Universal Minecraft Content Graph.

## Universal targets

- Bedrock add-on, behavior pack, resource pack, world, world template, editable source project, and complete world product
- Java data pack, resource pack, paired vanilla add-on family, world, Fabric project, NeoForge project, Forge project, multi-loader project, and standalone world-template mod source
- Universal bundle with both edition lanes, source workspaces, graph, contracts and proof

## Full-stack Bedrock → Java support

Generated Java projects do not hide Bedrock-only systems behind one generic placeholder. The release emits a target-native feature matrix mapping blocks, items, entities, scripts, worldgen, dimensions, spawn rules, animations, particles, render controllers, cameras, dialogue, trades, volumes, fog, UI, materials and other add-on surfaces to concrete Java registry, data-generation, event, networking, persistence, renderer, menu, command and world-system implementation points. Original Bedrock JSON/JavaScript is preserved alongside those contracts.

## Java → Bedrock support

Java assets, localization, common recipes, packaged data and loader/source metadata are translated or preserved where deterministic. Bedrock manifests, UUIDs, pack relationships, texture indexes, behavior/resource layouts and Script API scaffolds are generated. JVM bytecode, Mixins, ASM, custom networking, native code, complex GUIs and renderer hooks remain explicit source-recovery or reimplementation contracts—never silently discarded or mislabeled as converted.

## Worlds and templates

- Same-edition worlds and templates can be packaged directly.
- Cross-edition worlds use an allowlisted Chunker lane and retain warnings/evidence.
- Bedrock world products combine `.mctemplate`, companion `.mcaddon`, editable BP/RP/script source, conversion contracts and acceptance gates.
- Java world products combine an immutable embedded template, safe exporter, loader project, translated packs and integration contracts.

## Version and loader targeting

The embedded Version Atlas supplies Java versions, Java toolchains, mappings, loaders, build plugins and exact data/resource pack metadata when known. Bedrock outputs pin the selected minimum engine version and current stable Script API module contract. Large version jumps remain ordered migration programs rather than one unsafe search-and-replace pass.

## Safe specialist adapters

The first executable adapter set is Chunker, JE2BE Resource Pack Converter, Geyser PackConverter and Regolith. Tool paths are user-configured, validated as regular non-symlink files, and run without a shell using fixed commands, isolated working/cache directories, timeouts, complete logs, output hashing, structural validation and immutable-source rechecks.

Regolith projects use exact BP/RP export paths inside the adapter workspace; the converter never writes a generated project into a live `com.mojang` directory.

## Proof-first UX

The new OmniBridge workspace includes immutable-source status, session vault, adapter radar, target/version/loader architect, content graph statistics, fidelity coverage, ordered pipeline, semantic review queue, generated outputs, validation results and downloadable proof. Every visible output separates exact, translated, generated, tool-assisted, review and blocked content.

## Important boundary

There is no mathematically universal automatic equivalence between arbitrary Java JVM mods and arbitrary Bedrock add-ons. 0.11.0 provides the strongest practical route: deterministic conversion where representation is real, specialist adapters where a proven converter exists, target-native source products for code, complete source/evidence preservation, and a precise implementation/test queue for everything whose semantics cannot be proven automatically.
