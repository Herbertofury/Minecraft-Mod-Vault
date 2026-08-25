# 🔁 Mod Conversion Stack — Java ↔ Bedrock, Versions, Loaders & Bytecode

> **Status:** 📋 OmniBridge / Porting Lab architecture contract. This page covers **mods and add-ons in general**, not only saved worlds.

A Minecraft mod conversion is not a file-format rename. OmniBridge must preserve behavior across Minecraft versions, loaders, mapping namespaces and editions.

## Conversion lanes

| Lane | Preferred current stack | Main hazards |
|---|---|---|
| Java version → Java version | Modstitch + Stonecutter; Loom/Ravel/ModDevGradle; Mapping-IO/Tiny Remapper/Matcher | API moves/splits, registries, components, rendering, networking, Mixins, toolchain/JDK changes |
| Forge ↔ NeoForge | target-native APIs + ModDevGradle; common-source abstraction where useful | lifecycle/event differences, capabilities→attachments/components, networking/rendering/API drift |
| Forge/NeoForge ↔ Fabric | native semantic port first; Architectury/Porting Lib/FFAPI where appropriate; Connector/Kilt as compatibility references | APIs are not symmetric; AT/AW, registries, events, networking and Mixins require semantic migration |
| Fabric/Quilt era → modern Fabric | Loom/Ravel + official mappings boundary handling | Yarn/Intermediary-era source, refmaps/access wideners, 26.1+ unobfuscated transition |
| source-less Java JAR → target Java | remap/deobfuscate → Vineflower + CFR/Recaf → reconstruct project → semantic port | decompiler artifacts, lambdas/bridges, bytecode transforms, missing source metadata |
| Java mod → Bedrock add-on | Behavior IR + Geyser mappings/mappings-generator + Bedrock Creator schemas; Rainbow/Blockbench for assets | Java bytecode/APIs/Mixins have no direct Bedrock equivalent; Script API/components must recreate behavior |
| Bedrock add-on → Java mod | parse BP/RP/Script API/Molang → Behavior IR → target-native Java APIs/Mixins; Blockbench/GeckoLib for animated assets | component groups/events/controllers/scripts require semantic Java reimplementation |
| Java RP/data → Bedrock RP/BP | Geyser/Rainbow + specialist pack translators + Creator schemas | models/materials/custom items/commands/data features differ by edition |
| Bedrock RP/BP → Java assets/data | Bedrock schema parser + canonical asset/data IR | Molang/controllers/material semantics and Bedrock-only components need target-specific implementations |

## Preferred new multi-version / multi-loader workspace

### [Modstitch](https://github.com/isXander/modstitch)
Use as a primary current build abstraction when one project needs Fabric Loom, NeoForge ModDevGradle and legacy Forge ModDevGradle lanes.

### [Modstitch Toolkit](https://github.com/isXander/modstitch-toolkit)
Particularly valuable:

- `modstitch-accessx` for Access Transformer ↔ Access Widener / Class Tweaker conversion;
- common metadata/manifests;
- multiloader source-set/universal-JAR conventions.

### [Stonecutter](https://github.com/stonecutter-versioning/stonecutter)
Use source preprocessing when APIs genuinely diverge by Minecraft version/platform. Do **not** hide version-specific semantics behind brittle reflection merely to force one source body.

### [Architectury API](https://github.com/architectury/architectury-api) + [Architectury Loom](https://github.com/architectury/architectury-loom)
Strong mature common-code alternative where the abstraction matches the mod's domains.

### [Unimined](https://github.com/unimined/unimined)
Keep as a high-value broad/legacy toolchain reference, especially for unusual or older loader generations.

## Mapping & remapping stack

### Hard mapping-era branch

The converter records the exact source mapping namespace and Minecraft era. In particular, the Java 26.1+ unobfuscated era is handled differently from older obfuscated/remapped releases. Old Yarn/Intermediary/SRG/TSRG artifacts remain first-class archaeology inputs.

### Primary tools

- [Fabric Loom](https://github.com/FabricMC/fabric-loom) — project setup/remapping and mapping migration.
- [Ravel](https://github.com/FabricMC/fabric-tooling) / Fabric's current IDE migration workflow — source-aware mapping migration, especially useful for Kotlin and Mixins.
- [Matcher](https://github.com/FabricMC/Matcher) — structural class/member correspondence across game versions.
- [Tiny Remapper](https://github.com/FabricMC/tiny-remapper) — binary namespace remapping.
- [Mapping-IO](https://github.com/FabricMC/mapping-io) — mapping interchange / canonical Mapping IR foundation.
- [Enigma](https://github.com/FabricMC/Enigma) — interactive unresolved mapping/deobfuscation archaeology.
- [Parchment](https://github.com/ParchmentMC/Parchment) — Mojang mapping augmentation for mapped eras.
- [ForgeGradle](https://github.com/MinecraftForge/ForgeGradle), [MCPConfig](https://github.com/MinecraftForge/MCPConfig), [SrgUtils](https://github.com/MinecraftForge/SrgUtils) — Forge/SRG/TSRG/reobf archaeology.
- [NeoForge ModDevGradle](https://github.com/neoforged/ModDevGradle) + [NeoForm Runtime](https://github.com/NeoForged/NeoFormRuntime) — modern NeoForge/legacy Forge development and Minecraft artifact reconstruction.

## Mixin conversion contract

A Mixin is not migrated by changing one method name.

For each injection retain:

- target owner;
- method/field name;
- JVM descriptor;
- source mapping namespace;
- injector type;
- injection point / opcode / referenced member;
- ordinal/slice when present;
- nearby instruction/context fingerprint;
- semantic intent and expected application count.

Port process:

1. remap owner/name/descriptor into target namespace;
2. structurally match the target with Matcher/source diffs when direct mapping fails;
3. inspect target control-flow changes;
4. redesign the injection if the method split, moved, inlined or changed behavior;
5. prefer [MixinExtras](https://github.com/LlamaLad7/MixinExtras) composable injectors such as `@WrapOperation` / expression-level modification where they preserve intent better than brittle redirects;
6. verify actual injector application at runtime and exercise the affected behavior.

Refmap success alone is **not** Mixin-port success.

## Access modification contract

Forge/NeoForge Access Transformers, Fabric Access Wideners and Class Tweakers encode related but non-identical semantics.

Use a canonical Access IR:

`class/member + mapped descriptor + visibility intent + extendability intent + mutability/final intent + provenance`

Then emit AT/AW/ClassTweaker through target-aware adapters. [Modstitch Toolkit accessx](https://github.com/isXander/modstitch-toolkit) is the preferred current implementation reference.

Never perform plain syntax substitution before mapping the member into the target namespace.

## ASM / direct bytecode

Use [ASM](https://asm.ow2.io/) only where a native API or Mixin cannot faithfully express the required behavior.

Required checks:

- owner/name/descriptor remapped first;
- frame/max-stack correctness;
- verifier pass;
- transformed production JAR inspected;
- clean-instance runtime test;
- original bytecode and transform provenance retained.

## JAR-only reconstruction

Preferred pipeline:

`hash original → identify loader/version/mappings → remap/deobfuscate → Vineflower + independent CFR/Recaf decompile → compare disagreement → reconstruct build → compile → target semantic port → package → runtime differential tests`

Tools:

- [Vineflower](https://github.com/Vineflower/vineflower) — primary modern decompiler.
- [CFR](https://github.com/leibnitz27/cfr) — independent second opinion.
- [Recaf](https://github.com/Col-E/Recaf) — bytecode workspace/editing/mapping companion.

Decompiled code is evidence, **not original source**.

## Runtime compatibility systems — use as evidence, not native-port proof

### [Sinytra Connector](https://github.com/Sinytra/Connector)
Fabric-mod compatibility/transformation on NeoForge. Valuable for API/mapping transform research and differential testing; offline transform success is not enough.

### [Sinytra Launchpad](https://github.com/Sinytra/Launchpad)
Fabric Loader conventions/entrypoints/access widening on NeoForge without full code transformation. Great for mods already close to NeoForge-compatible.

### [Forgified Fabric API](https://github.com/Sinytra/ForgifiedFabricAPI)
Important Fabric API implementation/substitution source on NeoForge.

### [Kilt](https://github.com/KiltMC/Kilt)
Experimental Forge/NeoForge-mod compatibility on Fabric. Strong opposite-direction engineering reference, not the default production conversion path.

### [Porting Lib](https://github.com/Fabricators-of-Create/Porting-Lib)
Key library/reference for implementing Forge-like concepts on Fabric.

### [Forgix](https://github.com/PacifistMC/Forgix)
Useful final packaging/merged-JAR layer. Packaging must never be mistaken for semantic conversion.

## Java ↔ Bedrock semantic stack

### [Geyser mappings](https://github.com/GeyserMC/mappings)
Promote to first-class cross-edition mapping evidence. It contains living Java↔Bedrock mappings for blocks/items/effects/particles/sounds and related protocol/game semantics.

### [Geyser mappings-generator](https://github.com/GeyserMC/mappings-generator)
Architectural model: automatically generate deterministic mappings from Java/Bedrock data and **surface manual gaps explicitly** rather than guessing.

### [Rainbow](https://github.com/GeyserMC/Rainbow)
Experimental modern Java custom block/item/resource-pack → Geyser/Bedrock content generator. Strong reference for custom block/item assets, transforms, sounds, animated textures and related mappings.

### [Blockbench](https://www.blockbench.net/) + [GeckoLib](https://github.com/bernie-g/geckolib)
Preferred practical lane for Bedrock-style animated geometry/animations into Java animated entities where applicable. Preserve `.bbmodel` source and animation-controller/Molang semantics separately.

### [bridge.](https://github.com/bridge-core/editor)
Strong Bedrock BP/RP/schema/compiler/editor reference for generated add-ons.

### [Minecraft Creator docs](https://learn.microsoft.com/minecraft/creator/) + [bedrock-samples](https://github.com/Mojang/bedrock-samples)
Canonical target schemas and Script API/component evidence.

## Behavior IR

Java bytecode, loader events and Mixins cannot be mechanically translated into Bedrock JSON/Script API, and Bedrock component/event/controller/script graphs cannot be mechanically translated into Java classes.

Use a common Behavior IR carrying:

- trigger/lifecycle event;
- conditions;
- state and variables;
- actions;
- timing/tick semantics;
- authoritative logical side;
- networking/synchronization;
- persistence;
- referenced blocks/items/entities/registries;
- rendering/animation/particle/sound hooks;
- source-code/component provenance;
- exact / translated / generated / review / blocked verdict.

Target emitters then generate native Java APIs/Mixins or Bedrock components/controllers/Script API code.

## Data/version evidence

Do not rely only on Java symbol mappings. Use:

- [ViaVersion Mappings](https://github.com/ViaVersion/Mappings) / [ViaBackwards](https://github.com/ViaVersion/ViaBackwards) for protocol/registry/version semantic changes;
- [misode/mcmeta](https://github.com/misode/mcmeta) for processed versioned game data/assets/schema differences;
- [PrismarineJS minecraft-data](https://github.com/PrismarineJS/minecraft-data) as an independent Java/Bedrock data cross-check.

These feed the Version Atlas for IDs, components, commands, tags, recipes, loot, protocols and data-driven schema changes.

## Semantic Port IR

Each conversion stores at least:

- source artifact hash/repository commit;
- source/target Minecraft versions, loaders, Java versions and Gradle plugins;
- source/target mapping namespaces;
- class/method/field mapping provenance;
- lifecycle/event intent;
- registry identities;
- Mixin target fingerprints;
- access mutation intent;
- dependency contracts and target substitutes;
- packet schema/direction/phase/logical side;
- persistence/components/capabilities/attachments;
- config schema;
- rendering/model/animation contracts;
- Bedrock component/Molang/Script API contracts;
- generated asset/datagen provenance;
- unresolved semantics and confidence.

## Preferred migration order

1. target-native API;
2. proven common abstraction when semantics remain exact;
3. compatibility bridge as a temporary/differential lane;
4. narrow Mixin/MixinExtras interception;
5. ASM only when necessary;
6. explicit target-specific reimplementation or `blocked` when no equivalent exists.

## Acceptance tests

Compilation is intermediate evidence only. Test fixtures must include:

- Forge 1.20.1 → newer NeoForge;
- Forge 1.20.1 ↔ Fabric;
- pre-26.1 Fabric Yarn/Intermediary → Mojang mappings → 26.1+;
- Kotlin mapping migration;
- Inject/Accessor/Invoker/Redirect/MixinExtras cases;
- AT ↔ AW/ClassTweaker semantics;
- JAR-only dual-decompiler reconstruction;
- client/server networking and persistence;
- registries/data components/capabilities/attachments;
- rendering on client plus dedicated-server startup;
- Java custom blocks/items/entities → Bedrock where representable;
- Bedrock geometry/animations/Molang/scripts → Java reimplementation;
- production/reobfuscated JAR in a clean target instance.

Compare source vs target **behavior**: IDs, recipes, loot, spawning, persistence, networking, config, logical sides, render transforms, animations, sounds/particles, dependency interactions and dedicated-server behavior.

Full research ledger: [`research/minecraft-mod-conversion-stack-2026.md`](https://github.com/Herbertofury/Minecraft-Mod-Vault/blob/main/research/minecraft-mod-conversion-stack-2026.md).