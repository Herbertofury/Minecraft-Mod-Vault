# Minecraft mod / add-on conversion stack — 2026 research ledger

Verified: **2026-08-25**

Scope: general Minecraft mod/content conversion, **not just world conversion**. This ledger covers Java version ports, loader ports, Java ↔ Bedrock add-on conversion, mappings/remapping, Mixins/ASM, access transforms, source/JAR archaeology, data/resource packs, models/animations, dependency substitution, packaging and behavioral verification.

## Permanent architectural conclusion

A “mod converter” is not one translator. It is a pipeline of independently verifiable semantic migrations:

1. artifact/source archaeology;
2. source Minecraft version + loader + mapping namespace detection;
3. namespace/symbol mapping;
4. Minecraft API/version migration;
5. loader lifecycle/registry/event migration;
6. Mixins/bytecode/access-transform migration;
7. dependency/API substitution;
8. networking serialization and logical-side migration;
9. config/persistence/capabilities/components migration;
10. data/resource-pack and registry-schema migration;
11. rendering/model/animation/particle/sound migration;
12. Java ↔ Bedrock behavior/Script API/Molang semantic reconstruction where applicable;
13. metadata/build/packaging migration;
14. differential source-vs-target runtime verification.

The canonical pipeline should therefore be:

`source/JAR → evidence + namespaces → Semantic Port IR → target capability plan → deterministic transforms → target-native implementation → package → differential verifier`

Never report a port as successful just because it compiles or the JAR loads.

---

## 1. Cross-loader and multi-version development

### Modstitch — first-class recommendation

- Project: https://github.com/isXander/modstitch
- Current source inspected: build declares **0.8.5**.
- Purpose: unified Gradle DSL over Fabric Loom, NeoForge ModDevGradle and legacy Forge ModDevGradle.
- Important capabilities:
  - one build abstraction for Fabric / NeoForge / Forge;
  - automatic Mixin metadata handling;
  - Parchment integration;
  - designed to work with preprocessing systems such as Stonecutter.
- Verdict: **best current build-abstraction gem for a new multi-loader/multi-version porting workspace**.

### Modstitch Toolkit / accessx — major gem

- Project: https://github.com/isXander/modstitch-toolkit
- `modstitch-accessx`: converts between access-modification formats.
- `modstitch-manifests`: generates Fabric/NeoForge metadata from common declarations.
- `modstitch-multiloader`: source-set conventions and a universal-JAR path for Fabric + NeoForge.
- Particularly important for OmniBridge: the toolkit gives us a real implementation reference for **canonical access semantics → target AT/AW/ClassTweaker output**, instead of hand-written string conversion.
- Current surfaced versions on 2026-08-25 include accessx 0.1.1, manifests 0.1.5 and multiloader 0.1.8.

### Stonecutter

- Project: https://github.com/stonecutter-versioning/stonecutter
- Docs: https://stonecutter.kikugie.dev/
- Modern templates were still being updated in July 2026.
- Purpose: one source tree with version/platform conditional preprocessing and version-specific Gradle properties.
- Verified real project examples span `1.20.1` through `1.21.x` and `26.x`.
- Verdict: **best current source-level version-conditional method** when APIs genuinely diverge and cannot be abstracted cleanly.

### Architectury API / Architectury Loom

- API: https://github.com/architectury/architectury-api
- Loom: https://github.com/architectury/architectury-loom
- Organization: https://github.com/architectury
- Architectury API provides events, networking, registry and loader abstractions; Loom can set up Fabric, Forge, NeoForge and Quilt environments.
- Current repositories were active in July 2026.
- Verdict: mature common-code architecture; preserve as one of the primary multi-loader approaches.

### Unimined

- Project: https://github.com/unimined/unimined
- Unified Gradle environment supporting Fabric, Quilt, Forge, NeoForge and numerous legacy loaders/toolchains including LiteLoader, Rift, FoxLoader, ModLoader, jar modding, Bukkit/Spigot/Paper and custom loaders.
- LTS branch policy exists for build stability.
- Verdict: **best broad/legacy Gradle archaeology toolchain reference**; uniquely useful for old mods and unusual loader eras.

### Sinytra Connector

- Project: https://github.com/Sinytra/Connector
- Site: https://connector.sinytra.org/
- Current compatibility tests surfaced `3.0.0-beta.6+26.1.2` on 2026-08-15.
- Direction: **Fabric mods → NeoForge runtime compatibility**, not Forge → Fabric.
- Official testing explicitly warns that a successful offline JAR transform can still fail in-game.
- 1.20.1 remains an important LTS line in Sinytra documentation, but new compatibility work is concentrated on newer lines.
- Verdict: top runtime transformation reference and useful temporary compatibility lane; native source port is still preferred when full fidelity/control is required.

### Sinytra Launchpad

- Project: https://github.com/Sinytra/Launchpad
- Latest surfaced release: **1.9.2+26.1.2**, 2026-08-15.
- Reads `fabric.mod.json`, runs Fabric entrypoints, supports Class Tweakers / Access Wideners, nested JARs and Fabric Loader-facing runtime semantics on NeoForge.
- Explicitly designed for small/medium Fabric mods already close to NeoForge-compatible; it does not transform code like Connector.
- Verdict: **excellent low-invasion Fabric-conventions → NeoForge development bridge**.

### Forgified Fabric API

- Project: https://github.com/Sinytra/ForgifiedFabricAPI
- Direct Fabric API implementation on NeoForge, kept close to upstream.
- It is not a universal abstraction layer; loader-specific code still needs separate treatment.
- Verdict: important API substitution layer for Fabric → NeoForge source ports.

### Connector Extras

- Docs: https://sinytra.org/docs
- Bridges common third-party ecosystems such as Team Reborn Energy, JEI/REI/EMI, Forge Config API Port and KubeJS.
- Verdict: dependency-substitution research source; useful because real ports fail on secondary APIs as often as on the loader itself.

### Kilt

- Project: https://github.com/KiltMC/Kilt
- Direction: **(Neo)Forge mods → Fabric runtime compatibility**.
- Latest surfaced Forge-1.20.1 line: **v20.1.19**, released 2026-07-31.
- Explicitly experimental/unstable and can break worlds.
- Uses Forge API recreation, mappings/remapping and compatibility fixers; credits Porting Lib heavily.
- Verdict: powerful opposite-direction reference and testing lane, **not** production default.

### Porting Lib

- Project: https://github.com/Fabricators-of-Create/Porting-Lib
- Forge-like systems reimplemented on Fabric.
- Verdict: key source for substituting Forge concepts during native Fabric ports.

### Forgix

- Project: https://github.com/PacifistMC/Forgix
- Purpose: combine loader-specific artifacts; current project supports loader JAR merging and multiversion merging.
- Verdict: packaging layer only; never count a merged JAR as semantic loader conversion.

### Octo Loader — research/watch, not baseline

- Project: https://github.com/MilkdromedaStudios/Octo-Loader
- 2026 project attempting a unified Fabric/Quilt/Forge/NeoForge runtime plus cross-version remapping, API migration, access unification and Mixin retargeting.
- Its architecture contains unusually relevant ideas:
  - mapping-route composition across SRG/TSRG/Tiny/official-style namespaces;
  - API migration rule tables beyond symbol renames;
  - shared AT/AW access table;
  - adaptive Mixin targets using descriptor/type/context evidence;
  - compatibility reports that say what was transformed/skipped;
  - offline scan/planning before launch.
- The project itself states that mechanical translation cannot recreate missing behavior and that deep rendering/worldgen mods are difficult.
- Verdict: **bleeding-edge architecture reference/watch only** until independently validated against broad real-mod fixtures.

---

## 2. Mapping, remapping and version correspondence

### The 26.1 boundary is a hard branch in the converter

Fabric's current official docs establish:

- Minecraft Java was obfuscated through **1.21.11**;
- **26.1+ is unobfuscated and includes parameter names**;
- Yarn was deprecated for new post-1.21.11 versions;
- Fabric ports crossing into 26.1+ should migrate to Mojang/official names first;
- Loom uses separate remapping/non-remapping plugin paths across this boundary.

OmniBridge must therefore classify source projects into:

- **pre-26.1 obfuscated/remapped lane** — Mojmap/Yarn/Intermediary/SRG/TSRG/refmaps/reobf matter;
- **26.1+ unobfuscated lane** — ordinary obfuscation remapping disappears, but API/control-flow/loader/version semantic changes remain.

### Fabric Loom mapping migration

- Docs: https://docs.fabricmc.net/develop/porting/mappings/
- `migrateMappings` performs semi-automated source migration.
- Current Fabric docs explicitly note that manual review is still required, especially for Mixins.
- Loom is a strong default for Java source projects; not ideal for Kotlin mapping migration.

### Ravel

- Current Fabric docs recommend Ravel as the IntelliJ-based alternative to Loom migration.
- Uses IDE/PSI resolution and supports Kotlin.
- Handles Java/Kotlin source, Mixins, Access Wideners and Class Tweakers.
- Fabric API itself used Ravel for Yarn → Mojang migration.
- Verdict: **best complex/Kotlin source remap path**.

### Fabric Matcher

- Project: https://github.com/FabricMC/Matcher
- Latest surfaced release: **0.1.0**, 2026-03-12.
- Tracks corresponding classes/methods/fields across different JAR versions using structural evidence and existing mappings.
- Verdict: **major version-port gem** for finding moved/renamed targets when simple name mapping is insufficient.

### Tiny Remapper

- Project: https://github.com/FabricMC/tiny-remapper
- High-performance JAR remapper used throughout Fabric tooling.
- Verdict: canonical binary namespace-transform primitive for Tiny mapping pipelines.

### Mapping-IO

- Project: https://github.com/FabricMC/mapping-io
- Mapping format model/parser/writer used across Fabric tooling.
- Verdict: preferred mapping interchange layer for a canonical Mapping IR.

### Enigma

- Project: https://github.com/FabricMC/Enigma
- Interactive deobfuscation/mapping environment; integrates multiple decompilers.
- Verdict: manual archaeology tool for unresolved or ancient namespaces.

### Yarn + Intermediary

- Yarn: https://github.com/FabricMC/yarn
- Intermediary: https://github.com/FabricMC/intermediary
- Historical/pre-26.1 Fabric naming lanes remain essential when ingesting older source/JARs even though post-26.1 development no longer needs them.

### Mojang mappings + Parchment

- Parchment: https://github.com/ParchmentMC/Parchment
- Parchment remains a loader-neutral augmentation of Mojang names with parameter names/Javadocs for mapped eras and is still active in 2026.
- For **26.1+**, the game itself is unobfuscated and includes parameter names, reducing the old mapping role.

### ForgeGradle / MCPConfig / SrgUtils

- ForgeGradle: https://github.com/MinecraftForge/ForgeGradle
- MCPConfig: https://github.com/MinecraftForge/MCPConfig
- SrgUtils: https://github.com/MinecraftForge/SrgUtils
- Critical for Forge-era and especially 1.20.1/older source/JAR archaeology, reobfuscation and SRG/TSRG mapping data.
- ForgeGradle 7 remains actively developed in 2026, but older version ranges still require era-appropriate toolchains.

### NeoForge ModDevGradle

- Docs: https://docs.neoforged.net/toolchain/docs/plugins/mdg/
- Project: https://github.com/neoforged/ModDevGradle
- Provides current NeoForge dev tooling, Access Transformers, run environments and Parchment integration.
- Critically, its **legacy Forge plugin** can cover Forge through **1.20.1**, making it a strong bridge for this Vault's common 1.20.1 target.

### NeoForm Runtime

- Project: https://github.com/NeoForged/NeoFormRuntime
- Standalone low-level pipeline for acquiring/merging/patching/decompiling/recompiling Minecraft artifacts from NeoForm data.
- Supports access transformers and interface-injection data.
- Verdict: excellent target-source reconstruction and version-diff infrastructure reference.

### Mercury / MercuryMixin / Lorenz

- MercuryMixin: https://github.com/FabricMC/MercuryMixin
- Lorenz: https://github.com/CadixDev/Lorenz
- Useful source-remapping and mapping-object infrastructure, especially for Mixin-aware source transforms.

---

## 3. Mixins, ASM and access transforms

### SpongePowered Mixin

- Project: https://github.com/SpongePowered/Mixin
- Soft references in annotations must be mapped with the correct source/target namespace/refmap. Plain string rename is unsafe.
- A successful symbol remap does **not** prove the target method still has equivalent control flow.

### Fabric Mixin fork

- Fabric Loader docs: https://docs.fabricmc.net/develop/loader/
- Fabric uses an enhanced Mixin fork with constructor/interface injection and other fixes.
- Sinytra documents that NeoForge 1.20.4+ includes Fabric's Mixin fork, eliminating an older compatibility gap.

### MixinExtras

- Project: https://github.com/LlamaLad7/MixinExtras
- Current surfaced release: **0.5.4**, 2026-04-15.
- Fabric Loader 0.15+ and NeoForge 20.2.84+ bundle MixinExtras.
- `@WrapOperation`, `@ModifyExpressionValue`, `@Expression` and related injectors can be safer/more composable than brute-force redirects.
- Verdict: first-class Mixin port/hardening dependency.

### Access Transformers ↔ Access Wideners / Class Tweakers

- NeoForge AT docs: https://docs.neoforged.net/docs/1.20.4/advanced/accesstransformers/
- Fabric access syntax and mappings are handled by Loom/Fabric tooling.
- Modstitch Toolkit's `accessx` is a concrete current converter worth integrating.
- Important rule: these are **semantic permissions**, not simple syntax transforms. `public/protected/default`, `+f/-f`, `accessible/extendable/mutable` do not form a perfect 1:1 text mapping.
- Always remap the target member/class into the target namespace **before** converting access semantics.

### ASM

- Project: https://asm.ow2.io/
- Last-resort bytecode transform layer when a loader API/Mixin cannot express required equivalence.
- Required safeguards:
  - target owner/name/descriptor mapped first;
  - verify frames/max stack;
  - verify transformed bytecode with JVM/ASM verifier;
  - production-JAR test, not only dev workspace;
  - preserve original bytecode and transformation provenance.

### Mixin target fingerprints — recommended OmniBridge addition

For each injection, retain more than a method name:

- owner;
- descriptor;
- mapping namespace;
- injection point/opcode/member;
- nearby instruction/context fingerprint;
- optional source semantic label.

When a target disappears after a version jump, query structural/version evidence (Matcher + source diff + descriptor candidates) before rewriting the Mixin. If a method split, inlined or changed semantics, require human/agent semantic reconstruction instead of force-remapping.

Audit at runtime:

- Mixin config loaded;
- refmap namespace correct;
- injector application counts (`require` / `expect` where used);
- no silent optional injection loss unless intentionally approved;
- exported/transformed target bytecode when diagnosing;
- actual gameplay behavior still changes as intended.

---

## 4. JAR-only archaeology and reconstruction

### Vineflower

- Project: https://github.com/Vineflower/vineflower
- Current surfaced latest release: **1.12.0**, 2026-04-29.
- Modern JVM decompiler with strong bytecode mapping and current language-feature support.
- Verdict: default primary decompiler.

### Recaf

- Project: https://github.com/Col-E/Recaf
- Modern bytecode workspace/editor with multiple decompilers, mapping operations, rename support and class editing.
- Verdict: best interactive binary archaeology companion.

### CFR

- Project: https://github.com/leibnitz27/cfr
- Strong independent decompiler.
- Verdict: use as the second opinion in JAR-only ports.

### Dual-decompiler rule

For source-less mods:

`original JAR → hash/metadata → deobfuscate/remap → Vineflower + CFR/Recaf independent decompile → compare control flow → reconstruct build → compile → bytecode/API diff → runtime differential tests`

Never claim decompiled source is the original source. Preserve synthetic/bridge/lambda evidence and any sections where decompilers disagree.

---

## 5. Java version semantic data

### ViaVersion Mappings / ViaBackwards

- Mappings: https://github.com/ViaVersion/Mappings
- ViaBackwards: https://github.com/ViaVersion/ViaBackwards
- Not source-code converters, but excellent evidence for version-to-version renames/removals/registry/protocol semantics.
- Important for a Version Atlas that understands changes beyond Java symbol names.

### misode/mcmeta

- Project: https://github.com/misode/mcmeta
- Versioned processed Minecraft generated data/assets, with diff-oriented branches/data.
- Useful for blocks, registries, commands, item components, tags, recipes, loot, datapack/resource-pack schema and other data-driven changes.

### PrismarineJS minecraft-data

- Project: https://github.com/PrismarineJS/minecraft-data
- Broad versioned Java/Bedrock protocol/game-data corpus.
- Useful as independent cross-check for registry/protocol/version tables.

---

## 6. Java ↔ Bedrock mod/add-on semantic conversion

### GeyserMC mappings — first-class cross-edition data source

- Project: https://github.com/GeyserMC/mappings
- Current generated data inspected for **Java 26.2 ↔ Bedrock 1.26.30.5**.
- Includes blocks, collisions, effects, interactions, item components, items, particles, sounds and utility data.
- Verdict: **best living Java↔Bedrock semantic mapping corpus**.

### Geyser mappings-generator

- Project: https://github.com/GeyserMC/mappings-generator
- Generates most mapping data from Java datagen plus Bedrock data/samples and explicitly reports manual mapping gaps.
- This is the architecture OmniBridge should emulate: **generate what is deterministic, surface the irreducible manual semantic gaps**.

### Rainbow

- Project: https://github.com/GeyserMC/Rainbow
- Current experimental Fabric **26.2** tool.
- Generates custom Geyser block/item/skull/waypoint mappings and Bedrock resource-pack content from Java custom blocks/items/resource packs.
- Handles simple blocks, 2D/3D items, display transforms, sounds, animated textures, language files and some armor/equipment assets.
- Verdict: **strongest new Java-custom-content → Bedrock asset/mapping gem**; experimental, not a complete Java-mod → Bedrock-addon converter.

### Geyser PackConverter

- Project: https://github.com/GeyserMC/PackConverter
- Java resource-pack → Bedrock library/tool; explicitly WIP and not production-ready.
- Does not fully create custom item mappings; Rainbow fills more of that gap.

### EasyEdit-Data

- Project: https://github.com/platz1de/EasyEdit-Data
- Java/Bedrock block-state mapping and versioned state conversion evidence.
- Useful secondary block-state mapping source.

### bridge.

- Project: https://github.com/bridge-core/editor
- Bedrock add-on IDE/editor with schema validation, completions, compiler plugins and previews.
- Verdict: strong Bedrock-side parser/editor/validation reference for generated BP/RP projects.

### Blockbench

- Official: https://www.blockbench.net/
- Supports Java block/item, modded entity and Bedrock model/animation workflows plus project conversion.
- Verdict: canonical model/rig interchange workbench, but conversion can be lossy and `.bbmodel` source should be preserved.

### GeckoLib — major Bedrock → Java entity bridge

- Project: https://github.com/bernie-g/geckolib
- GeckoLib's Blockbench workflows can convert existing Bedrock/modded entity model projects to GeckoLib animated models.
- GeckoLib 4.x supports a Bedrock-style animation format and Molang-like animation expressions.
- Verdict: **one of the best practical reuse paths for Bedrock geometry/animations in a Java mod**.

### Minecraft Creator schemas / bedrock-samples

- Creator docs: https://learn.microsoft.com/minecraft/creator/
- Samples: https://github.com/Mojang/bedrock-samples
- Canonical source of target Bedrock component/schema/Script API expectations.

### JE2BE Resource Pack Converter

- Project: https://github.com/Seraphic-Studio/JE2BE-Resource-Pack-Converter
- Java resource pack → Bedrock RTX/PBR specialist.
- Pack translation only; not behavior/mod conversion.

### Java behavior ↔ Bedrock behavior rule

There is no safe general bytecode→Script API or JSON-component→Java-bytecode text translator.

Represent behavior in a **Behavior IR**:

- lifecycle/event trigger;
- state/variables;
- conditions;
- actions;
- authoritative side;
- persistence;
- networking/sync;
- timing/ticks;
- entity/block/item/registry references;
- rendering/animation hooks;
- fallback when no equivalent exists.

Then emit target-native Java events/APIs/Mixins or Bedrock components/animation controllers/Script API code. Preserve source script/code and mark each behavior as exact / translated / generated / review / blocked.

---

## 7. Required Semantic Port IR

Every mod port record should retain:

- source artifact hash and optional source repository commit;
- source Minecraft version, loader, Java version and build plugin;
- source mapping namespace(s);
- target Minecraft version/loader/Java version/mapping namespace;
- class/method/field mapping provenance;
- registry identifiers and registry lifecycle intent;
- event/lifecycle intent;
- mixin targets + descriptors + instruction fingerprints;
- access mutation intent;
- dependency/API contracts and target substitutes;
- packet schema, direction, phase and logical side;
- persistence/data component/capability/component contracts;
- config format/schema;
- model/renderer/animation/material contracts;
- Bedrock component/Molang/Script API behavior contracts where applicable;
- generated resources/datagen provenance;
- confidence and unresolved semantics.

This IR prevents conversion from degenerating into import renames and regex patches.

---

## 8. Preferred migration decision order

For each source feature:

1. use the target loader/game's native API when an equivalent exists;
2. use a proven common abstraction (Architectury, FFAPI, Porting Lib, etc.) when it preserves semantics;
3. use a compatibility bridge only when native port cost is not justified or as a differential reference;
4. use Mixin/MixinExtras for narrow behavior interception;
5. use ASM only when the above cannot express the required behavior;
6. if the target edition has no equivalent, generate an explicit target-specific implementation contract or mark blocked — never fake support.

---

## 9. Mandatory verification corpus

A conversion test suite needs real fixtures for:

- Forge 1.20.1 → NeoForge 1.21.x / 26.x;
- Forge 1.20.1 → Fabric and Fabric → Forge/NeoForge;
- Fabric Yarn/Intermediary pre-26.1 → Mojang mappings → 26.1+ unobfuscated;
- a Kotlin Fabric mod through Ravel mapping migration;
- Mixins containing `@Inject`, accessors/invokers, redirects and MixinExtras `@WrapOperation` / expressions;
- AT ↔ AW/ClassTweaker conversion including final/mutable/extendable cases;
- JAR-only mod with no source using dual decompilers;
- networking packets both directions and login/play phases;
- registries, recipes, loot, tags, worldgen, commands and item components;
- capabilities/components/attachments and persistent player/entity data;
- client-only rendering mod and dedicated-server startup;
- config migration;
- Java custom entity/model/animation → Bedrock representation where possible;
- Bedrock geometry/animation/Molang → Java GeckoLib/native implementation;
- Java custom blocks/items/resource pack → Geyser/Rainbow Bedrock content;
- Bedrock Script API behavior → Java semantic reimplementation;
- Java behavior → Bedrock components/Script API semantic reimplementation;
- release/reobfuscated JAR test in a clean production instance, not just `runClient`.

For every fixture compare source and target behavior, not just startup:

- registry IDs;
- spawn/placement;
- recipes/loot;
- save/reload;
- network sync;
- logical sides;
- commands/config;
- rendering/model transforms;
- animations/sounds/particles;
- persistent data;
- dedicated server behavior;
- dependency interactions.

---

## 10. Recommendations promoted to Minecraft Mod Vault

1. Make **Modstitch + Stonecutter** the preferred new multi-loader/multi-version workspace pattern, with Architectury and Unimined retained as major alternatives.
2. Integrate **modstitch-accessx** as the primary AT/AW/ClassTweaker conversion reference.
3. Build Mapping IR around **Mapping-IO + Tiny Remapper**, with **Matcher** for cross-version structural correspondence and **Ravel/Loom** for source migration.
4. Add a permanent **pre-26.1 vs 26.1+ mapping/toolchain branch**.
5. Make Mixin conversion descriptor/control-flow aware; refmap remapping alone is insufficient.
6. Prefer **MixinExtras** over brittle redirects when its injectors express the intended operation.
7. Use **Vineflower + CFR/Recaf** as the default dual-decompiler JAR-only pipeline.
8. Use **Geyser mappings + mappings-generator** as the primary Java↔Bedrock semantic mapping evidence, with **Rainbow** for modern Java custom-content→Bedrock generation.
9. Use **Blockbench + GeckoLib** as the preferred Bedrock-model/animation→Java reuse path when applicable.
10. Treat Sinytra Connector/Kilt/Octo-style runtimes as compatibility/differential research lanes, not proof of a native port.
11. Make compilation, startup and offline transform success intermediate evidence only; final acceptance requires real source-vs-target behavior tests.
