# 🧬 Loaders & Minecraft Versions

## Current repair/porting foundation

Porting Lab + Compatibility Brain already establish the evidence model: detect the actual source loader/version/toolchain before proposing changes.

For the full current toolchain covering mappings, Mixins, access transforms, JAR archaeology, cross-loader workspaces and Java ↔ Bedrock add-on reconstruction, see **[Mod Conversion Stack](Mod-Conversion-Stack)**.

## OmniBridge roadmap

- Forge
- NeoForge
- Fabric
- Quilt-era inputs where practical
- Bukkit/Paper/server ecosystems where relevant to artifact identity or adapters

### Migration domains

| Domain | Examples |
|---|---|
| registration/lifecycle | deferred registers, entry points, event buses |
| mappings/names | Mojang, Yarn/Intermediary, SRG/TSRG, mapping-era → 26.1+ unobfuscated Java |
| build/toolchain | Loom, ModDevGradle, legacy Forge ModDevGradle/ForgeGradle, Modstitch, Stonecutter, Unimined, Java runtime requirements |
| Mixins/bytecode | refmaps, descriptors, injection points, MixinExtras, ASM transforms, target control-flow changes |
| access modification | Forge/NeoForge Access Transformers, Fabric Access Wideners, Class Tweakers |
| networking | packet APIs, codecs, directions, login/play phases, logical sides |
| rendering | model/render layer APIs, client setup, shader hooks |
| data/component systems | capabilities/components/attachments/data components |
| worldgen | registries, codecs, datapack-driven definitions |
| dependencies | API replacements, compatibility shims, optional integrations |
| testing | GameTest, loader JUnit, real client/server smoke tests, source-vs-target differential behavior |

## Preferred current multi-version / multi-loader architecture

- **[Modstitch](https://github.com/isXander/modstitch)** — unified project/build abstraction across Loom, modern NeoForge ModDevGradle and legacy Forge ModDevGradle.
- **[Modstitch Toolkit](https://github.com/isXander/modstitch-toolkit)** — especially `accessx` for AT ↔ AW/ClassTweaker conversion plus common metadata/multiloader conventions.
- **[Stonecutter](https://github.com/stonecutter-versioning/stonecutter)** — source-level version/platform conditional preprocessing when APIs genuinely differ.
- **[Architectury](https://github.com/architectury)** — mature common-code abstraction where its APIs match the mod's domains.
- **[Unimined](https://github.com/unimined/unimined)** — broad/legacy loader and unusual-version toolchain support.

## Mapping stack

- **[Mapping-IO](https://github.com/FabricMC/mapping-io)** as canonical mapping interchange.
- **[Tiny Remapper](https://github.com/FabricMC/tiny-remapper)** for binary namespace transforms.
- **[Matcher](https://github.com/FabricMC/Matcher)** for structural class/member correspondence across versions when direct mappings fail.
- **Fabric Loom `migrateMappings` / Ravel** for source-level migration.
- **ForgeGradle / MCPConfig / SrgUtils** for Forge/SRG/TSRG-era archaeology and reobfuscation.
- **NeoForge ModDevGradle / NeoForm Runtime** for current NeoForge plus legacy Forge-through-1.20.1 bridge and target-source reconstruction.

### 26.1+ mapping boundary

Minecraft Java **26.1+** is treated as a separate tooling era from pre-26.1 obfuscated/remapped releases. The converter must know whether source names are Mojang/Yarn/Intermediary/SRG/TSRG and whether refmaps/reobfuscation are required. Removing the obfuscation step does **not** remove API/behavior migration work.

## Mixin rule

A Mixin target is identified by more than a method name. OmniBridge stores owner, JVM descriptor, source namespace, injector type, injection point/opcode/member, slice/ordinal, nearby instruction fingerprint and semantic intent. When a version jump changes control flow, the injection is reconstructed semantically rather than force-remapped.

Use **[MixinExtras](https://github.com/LlamaLad7/MixinExtras)** where its composable injectors preserve intent more reliably than brittle redirects.

## Access rule

Access Transformers, Access Wideners and Class Tweakers are converted through a canonical **Access IR**, after target members are mapped. They are never converted by blind syntax replacement.

## Compatibility bridges

- **[Sinytra Connector](https://github.com/Sinytra/Connector)** / **Launchpad** / **Forgified Fabric API** are valuable Fabric → NeoForge compatibility/differential-testing lanes.
- **[Kilt](https://github.com/KiltMC/Kilt)** is an important experimental Forge/NeoForge → Fabric reference lane.
- **[Porting Lib](https://github.com/Fabricators-of-Create/Porting-Lib)** is a major Forge-concept-on-Fabric substitution source.

A compatibility bridge working does not prove a native source port is complete.

## Non-negotiable

Loader/version conversion is never just import/package renaming or a successful compile. The final target must preserve the requested behavior in a clean target runtime, including client, dedicated server, persistence, networking, rendering, Mixins, data/registries and dependency interactions.
