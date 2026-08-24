# Minecraft Mod Vault 0.8.0 Ultimate Mod Repair, Porting, and Tool Knowledge Base

- Reviewed: **2026-08-20**
- Curated sources: **154**
- Detailed execution records: **92**
- Unique runtime tool cards after catalog merge: **163**
- Reusable repair patterns: **9**
- Reusable verified repair patterns: **9**
- Purpose: power exact mod identification, dependency solving, crash diagnosis, API-drift repair, source recovery, upgrade/downgrade planning, multi-loader conversion, verified builds, and world/data repair.
- Maturity labels describe current integration posture. Runtime bridges are never treated as proof of compatibility without actual launch and workload verification.

## Product architecture this research supports

1. **Artifact intelligence:** inspect hashes, loader metadata, dependencies, Java class levels, mappings, mixins, transformers, nested JARs, signatures, native libraries, and source provenance.
2. **Compatibility graph:** detect missing dependencies, duplicate IDs/classes, declared conflicts, dependency cycles, loader/API boundaries, and dependency-first transaction order.
3. **Porting planner:** choose source rebuild, remap, source transform, narrow binary repair, compatibility adapter, runtime bridge, metadata/data migration, or review-only substitution based on evidence.
4. **Repair executor:** stage every change, preserve originals, verify hashes and archive integrity, build in isolated toolchains, launch real client/server targets, inspect fresh logs, and roll back on failure.
5. **Durable Repair Brain:** preserve successful and failed repair evidence, exact hashes, version applicability, symbols, data keys, and supersession so mistakes are not repeated.

## Best overall porting foundation

- [Fabric Loom](https://github.com/FabricMC/fabric-loom): Strongest Fabric workspace, remap, run, and test foundation.
- [NeoForge ModDevGradle](https://github.com/neoforged/ModDevGradle): Current NeoForge development and testing foundation.
- [ForgeGradle](https://github.com/MinecraftForge/ForgeGradle): Canonical Forge build and remapping path.
- [Mapping-IO](https://github.com/FabricMC/mapping-io): Common mapping graph and namespace conversion layer.
- [Tiny Remapper](https://github.com/FabricMC/tiny-remapper): High-performance JAR remapping engine.
- [Vineflower](https://github.com/Vineflower/vineflower): Leading maintained Java decompiler for source recovery and diffing.

## Highest-capability repair stack

- [ASM](https://asm.ow2.io/): Low-level, exact bytecode inspection and patch generation.
- [Recaf](https://github.com/Col-E/Recaf): Interactive bytecode exploration, decompilation, and repair.
- [japicmp](https://github.com/siom79/japicmp): Binary API change detection between dependency versions.
- [Revapi](https://revapi.org/): Extensible API compatibility analysis.
- [OpenRewrite](https://github.com/openrewrite/rewrite): Recipe-driven source migrations at scale.
- [SpongePowered Mixin](https://github.com/SpongePowered/Mixin): Transformer inspection and compatibility work for Mixin-heavy mods.
- [MixinExtras](https://github.com/LlamaLad7/MixinExtras): Safer, more expressive injection primitives for modern ports.

## Strongest bleeding-edge options to harden

- [Sinytra Connector](https://github.com/Sinytra/Connector): Automated Fabric-to-NeoForge runtime compatibility where the target mod set is proven compatible.
- [Sinytra Probe](https://github.com/Sinytra/Probe): Current automated compatibility probing for Connector candidates.
- [Kilt](https://github.com/KiltMC/Kilt): Forge-on-Fabric compatibility research path with meaningful but non-universal coverage.
- [Modstitch](https://github.com/isXander/Modstitch): Modern multi-loader build unification and shared-source orchestration.
- [Stonecutter](https://github.com/stonecutter-versioning/stonecutter): Version-aware preprocessing and multi-version source management.
- [Unimined](https://github.com/unimined/unimined): Loader-agnostic Gradle experimentation and remapping.

## Update, provenance, and modpack backbone

- [Modrinth API](https://docs.modrinth.com/api/): Cryptographic identity, compatible versions, dependencies, and project metadata.
- [CurseForge API](https://docs.curseforge.com/): Project/file metadata and fingerprint-based identification.
- [GitHub Releases API](https://docs.github.com/en/rest/releases/releases): Release assets, metadata, and current SHA-256 asset digests.
- [Packwiz](https://github.com/packwiz/packwiz): Reproducible, provider-aware modpack manifests and updates.
- [Ferium](https://github.com/gorilla-devs/ferium): Fast CLI mod and modpack updates from Modrinth and CurseForge.

## World and data repair backbone

- [DataFixerUpper](https://github.com/Mojang/DataFixerUpper): Schema-driven forward migration model for serialized game data.
- [NBT Studio](https://github.com/tryashtar/nbt-studio): Deep NBT inspection and targeted repair.
- [Amulet](https://github.com/Amulet-Team/Amulet-Core): Cross-version world conversion and structured save access.
- [MCA Selector](https://github.com/Querz/mcaselector): Region-level inspection, filtering, and selective repair.
- [Spyglass](https://github.com/SpyglassMC/Spyglass): Current Minecraft data/resource language tooling and diagnostics.

## Complete curated catalog

### API compatibility and binary diffing

#### [japicmp](https://siom79.github.io/japicmp/)

- **ID:** `japicmp`
- **Maturity/status:** stable / active
- **Priority:** 100
- **Repository:** [https://github.com/siom79/japicmp](https://github.com/siom79/japicmp)
- **Best use:** Proving removed/changed classes, methods, fields, descriptors and inheritance across mod or dependency versions.
- **Capability:** Binary compatibility comparison between Java archives.
- **Vault integration:** Compare exact old/new dependency JARs and feed break records into source or bytecode repair planning.
- **Notes:** Core evidence generator for missing methods, fields, classes, and descriptor changes.

#### [Revapi](https://revapi.org/)

- **ID:** `revapi`
- **Maturity/status:** stable / active
- **Priority:** 90
- **Repository:** [https://github.com/revapi/revapi](https://github.com/revapi/revapi)
- **Best use:** Maintaining compatibility baselines and detecting accidental API regressions in rebuilt mods.
- **Capability:** Extensible API analysis across Java archives with configurable compatibility rules.
- **Vault integration:** Run before and after ports, classify intentional changes and block unexplained public API drift.
- **Notes:** Use alongside japicmp for richer API policy and extension points.

### Archive and source diffing

#### [GumTree](https://github.com/GumTreeDiff/gumtree)

- **ID:** `gumtree`
- **Maturity/status:** stable / active
- **Priority:** 90
- **Best use:** AST-aware source differencing and edit scripts
- **Vault integration:** library-adapter
- **Notes:** Use to mine upstream mod ports and loader API changes into candidate rewrite rules.

### Automated testing and runtime harnesses

#### [Fabric automated testing](https://docs.fabricmc.net/develop/automatic-testing)

- **ID:** `fabric-testing`
- **Maturity/status:** official / active
- **Priority:** 100
- **Repository:** [https://github.com/FabricMC/fabric](https://github.com/FabricMC/fabric)
- **Best use:** Fabric GameTest and automated launch testing
- **Vault integration:** runtime-adapter
- **Notes:** Core real-runtime verification layer for Fabric repairs.

#### [NeoForge Game Tests](https://docs.neoforged.net/docs/misc/gametest/)

- **ID:** `neoforge-gametest`
- **Maturity/status:** official / active
- **Priority:** 100
- **Repository:** [https://github.com/neoforged/NeoForge](https://github.com/neoforged/NeoForge)
- **Best use:** Behavioral regression proof for NeoForge ports and compatibility patches.
- **Capability:** GameTest framework integration and dedicated run targets for NeoForge mods.
- **Vault integration:** Run server/client-relevant tests plus clean dedicated-server startup before packaging.
- **Notes:** Core real-runtime verification layer for NeoForge repairs.

#### [HeadlessMC](https://github.com/3arthqu4ke/HeadlessMC)

- **ID:** `headlessmc`
- **Maturity/status:** stable / active
- **Priority:** 90
- **Best use:** Headless Minecraft launch and automation
- **Vault integration:** runtime-adapter
- **Notes:** High-value isolated client/server smoke-test runner.

#### [JUnit 5](https://github.com/junit-team/junit5)

- **ID:** `junit`
- **Maturity/status:** stable / active
- **Priority:** 90
- **Best use:** Unit and integration test framework
- **Vault integration:** library-adapter
- **Notes:** Use for generated migration recipes, parsers, and deterministic patch tests.

#### [Testcontainers for Java](https://github.com/testcontainers/testcontainers-java)

- **ID:** `testcontainers`
- **Maturity/status:** stable / active
- **Priority:** 75
- **Best use:** Disposable containerized integration environments
- **Vault integration:** library-adapter
- **Notes:** Useful for server management and provider-adapter tests, though not a substitute for real Minecraft runtime tests.

### Binary remappers

#### [Tiny Remapper](https://github.com/FabricMC/tiny-remapper)

- **ID:** `tiny-remapper`
- **Maturity/status:** stable / active
- **Priority:** 100
- **Best use:** Rewriting compiled mod JAR namespaces and class/member references before deeper API repair.
- **Capability:** Fast Java bytecode remapping with hierarchy-aware member resolution and extension points.
- **Vault integration:** Perform deterministic dry-run remaps, capture unresolved references, compare class trees and preserve reproducible inputs.
- **Notes:** Primary Fabric ecosystem remapper and a strong foundation for Vault namespace transforms.

#### [Auto Renaming Tool](https://github.com/MinecraftForge/AutoRenamingTool)

- **ID:** `auto-renaming-tool`
- **Maturity/status:** stable / active
- **Priority:** 90
- **Best use:** Forge-oriented JAR renaming and mapping application
- **Vault integration:** cli-adapter
- **Notes:** Use for SRG and official mapping transformations where applicable.

#### [SpecialSource](https://github.com/md-5/SpecialSource)

- **ID:** `specialsource`
- **Maturity/status:** stable / active
- **Priority:** 80
- **Best use:** Class, member, and inheritance-aware remapping
- **Vault integration:** cli-adapter
- **Notes:** Useful for Bukkit, Forge, and older mapping pipelines.

### Build systems and loader toolchains

#### [Fabric Loom](https://docs.fabricmc.net/develop/loom/)

- **ID:** `fabric-loom`
- **Maturity/status:** stable / active
- **Priority:** 100
- **Repository:** [https://github.com/FabricMC/fabric-loom](https://github.com/FabricMC/fabric-loom)
- **Best use:** Fabric source ports, namespace migration and reproducible target builds.
- **Capability:** Gradle workspace setup, Minecraft dependency provisioning, decompilation, remapping, access widener migration, run configurations, production runs and game-test integration.
- **Vault integration:** Generate version-pinned workspaces and invoke mapping migration, remapJar, runClient, runServer and production test tasks inside isolated build sandboxes.
- **Current evidence:** Loom 1.16 is the latest stable release observed on 2026-08-20; active 1.17 development builds are already used for Minecraft 26.2.
- **Notes:** Primary Fabric source reconstruction and remap tool.
- **Material risks:** Build scripts are version-sensitive; New unobfuscated Minecraft versions use different plugin/remapping paths

#### [ForgeGradle](https://docs.minecraftforge.net/en/latest/gettingstarted/)

- **ID:** `forgegradle`
- **Maturity/status:** stable / active
- **Priority:** 100
- **Repository:** [https://github.com/MinecraftForge/ForgeGradle](https://github.com/MinecraftForge/ForgeGradle)
- **Best use:** Forge source ports and faithful rebuilding of Forge-origin projects.
- **Capability:** MinecraftForge development, mappings, reobfuscation and run-configuration toolchain.
- **Vault integration:** Generate target Forge workspaces, preserve access transformers and test client/server/data/game-test runs.
- **Notes:** Required for many legacy Forge source recoveries and rebuilds.

#### [ModDevGradle](https://docs.neoforged.net/toolchain/docs/plugins/mdg/)

- **ID:** `moddevgradle`
- **Maturity/status:** stable / active
- **Priority:** 100
- **Repository:** [https://github.com/neoforged/ModDevGradle](https://github.com/neoforged/ModDevGradle)
- **Best use:** NeoForge source ports and target runtime matrices.
- **Capability:** Modern NeoForge development plugin with run configurations, configuration-cache support and current Gradle practices.
- **Vault integration:** Generate version-pinned NeoForge workspaces, run client/server/data/game-test targets and capture produced artifacts and logs.
- **Notes:** Preferred current NeoForge development path.

#### [Gradle JVM toolchains](https://docs.gradle.org/current/userguide/toolchains.html)

- **ID:** `gradle-toolchains`
- **Maturity/status:** official / active
- **Priority:** 99
- **Repository:** [https://github.com/gradle/gradle](https://github.com/gradle/gradle)
- **Best use:** Building old and new Minecraft targets with their exact Java requirements.
- **Capability:** Reproducible JDK selection, download and compile/test/runtime separation.
- **Vault integration:** Declare target JDKs per version, compile with release constraints and record vendor/version in build receipts.
- **Notes:** Current Gradle 9.7 documentation recommends toolchains and combining them with --release for strict bytecode/API compatibility.

#### [Architectury Loom](https://docs.architectury.dev/loom/introduction)

- **ID:** `architectury-loom`
- **Maturity/status:** stable / active
- **Priority:** 95
- **Repository:** [https://github.com/architectury/architectury-loom](https://github.com/architectury/architectury-loom)
- **Best use:** Multi-loader projects that benefit from shared source and loader-specific outputs.
- **Capability:** Loom-derived Gradle plugin supporting Fabric, Forge, NeoForge and Quilt development workflows.
- **Vault integration:** Use only when a shared-source architecture reduces duplicated behavior without hiding loader-specific correctness.
- **Notes:** Strong foundation for common source across Fabric, Forge, NeoForge, and Quilt generations.

#### [Architectury Plugin](https://github.com/architectury/architectury-plugin)

- **ID:** `architectury-plugin`
- **Maturity/status:** stable / active
- **Priority:** 90
- **Best use:** Platform transformation and common-module injection for Architectury projects
- **Vault integration:** cli-adapter
- **Notes:** Pair with Architectury Loom when reconstructing a unified multi-loader source tree.

#### [Modstitch](https://github.com/isXander/Modstitch)

- **ID:** `modstitch`
- **Maturity/status:** experimental / active
- **Priority:** 88
- **Best use:** Modern Fabric/NeoForge shared-source ports needing loader-native toolchains.
- **Capability:** Unifies official Loom and ModDevGradle workflows and translates access wideners/access transformers for multi-loader projects.
- **Vault integration:** Evaluate as the high-capability shared-build path; retain generated loader-native outputs and independent runtime proof.
- **Notes:** High-capability unified build layer, currently published through unstable plugin versions. Pin exact plugin/toolchain tuples and run every target build before adopting generated changes.
- **Material risks:** Fast-moving build integration requires version pinning and regression tests

#### [Unimined](https://unimined.github.io/)

- **ID:** `unimined`
- **Maturity/status:** beta / active
- **Priority:** 86
- **Repository:** [https://github.com/unimined/Unimined](https://github.com/unimined/Unimined)
- **Best use:** Reconstructing and testing modern or legacy source workspaces with one extensible toolchain.
- **Capability:** Multi-loader, multi-version and legacy-aware Gradle provisioning, mappings, remapping, runs and custom transformation pipelines.
- **Vault integration:** Generate a clean target workspace from the current LTS branch, pin every loader/mapping version, and compare output against loader-native builds.
- **Current evidence:** Repository default branch observed as lts/1.4 on 2026-08-20.
- **Notes:** Broadest build-tool reach in the catalog, especially valuable for legacy recovery. Use the current LTS branch and validate exact loader/version support instead of assuming every matrix cell is complete.
- **Material risks:** Very broad support matrix has uneven maturity; Custom/legacy combinations need exact fixture projects and runtime tests

#### [Cloche](https://github.com/terrarium-earth/cloche)

- **ID:** `cloche`
- **Maturity/status:** experimental / active
- **Priority:** 82
- **Best use:** Exploring a single-project target model for complex multi-loader/version builds.
- **Capability:** Target-oriented cross-platform Minecraft Gradle plugin with source-set separation, data generation, tests, runs, metadata generation and pre-applied Mixin support.
- **Vault integration:** Benchmark against loader-native builds in isolated generated projects before adopting.
- **Current evidence:** Active repository documented in August 2026.
- **Notes:** Promising high-capability alternative architecture. Its pre-applied Mixin/debug paths include work-in-progress areas, so adopt behind generated-workspace tests.
- **Material risks:** Some debugging/Mixin features are explicitly work in progress; Smaller production evidence base than loader-native tooling

#### [Stonecraft](https://github.com/meza/Stonecraft)

- **ID:** `stonecraft`
- **Maturity/status:** beta / active
- **Priority:** 80
- **Best use:** Reducing repeated build logic in source-owned Architectury projects.
- **Capability:** Versioned Gradle plugin and template that wires Stonecutter and Architectury for multi-version, multi-loader mod projects.
- **Vault integration:** Generate or migrate in a branch, run all target builds and compare published metadata/artifacts.
- **Current evidence:** Current repository documentation advertises a maintained plugin and template.
- **Notes:** Good productivity layer for source-owned projects already aligned with Architectury. Validate generated configuration and publishing tasks for every supported tuple.
- **Material risks:** Adds an abstraction layer over two other toolchains; Generated publishing behavior needs target-by-target verification

#### [NeoGradle](https://docs.neoforged.net/toolchain/docs/plugins/ng/)

- **ID:** `neogradle`
- **Maturity/status:** maintenance / active
- **Priority:** 75
- **Repository:** [https://github.com/neoforged/NeoGradle](https://github.com/neoforged/NeoGradle)
- **Best use:** Projects whose target version or upstream source expects NeoGradle rather than ModDevGradle.
- **Capability:** NeoForge Gradle toolchain used by supported NeoForge development templates.
- **Vault integration:** Select from project/target compatibility evidence and avoid gratuitous build-system migration.
- **Notes:** Keep for existing project migration; prefer ModDevGradle for new reconstructed projects when supported.

#### [Multisource](https://github.com/lukebemishprojects/Multisource)

- **ID:** `multisource`
- **Maturity/status:** experimental / active
- **Priority:** 72
- **Best use:** Low-duplication Fabric/NeoForge source ownership with explicit parent source sets.
- **Capability:** Gradle settings plugin that models common and loader-specific source sets through feature variants and delegated Architectury Loom projects.
- **Vault integration:** Evaluate in a generated sample, verify remap/publishing artifacts and compare against established Architectury/Stonecutter layouts.
- **Current evidence:** Repository and plugin documentation available in 2026; adoption remains small.
- **Notes:** Interesting low-duplication multiloader architecture, but smaller and less proven than Architectury, Stonecutter or Modstitch. Evaluate through a generated sacrificial workspace first.
- **Material risks:** Limited adoption and release history; Not a binary conversion engine

#### [Quilt Loom](https://github.com/QuiltMC/quilt-loom)

- **ID:** `quilt-loom`
- **Maturity/status:** maintenance / active
- **Priority:** 70
- **Best use:** Quilt workspace and mappings tooling for supported legacy and current projects
- **Vault integration:** cli-adapter
- **Notes:** Evaluate project version before choosing over Fabric Loom.

#### [Blahaj](https://github.com/txnimc/Blahaj)

- **ID:** `blahaj`
- **Maturity/status:** research / active
- **Priority:** 60
- **Best use:** Mining automation ideas and rapidly testing version matrices.
- **Capability:** Stonecutter-based Gradle plugin for multiversion project setup, dependency declarations, compatibility metadata and publishing.
- **Vault integration:** Keep as a research adapter until it passes the Vault cross-version fixture matrix.
- **Current evidence:** Current repository exists with a fluent multi-platform DSL but limited adoption evidence.
- **Notes:** Bleeding-edge candidate with limited adoption evidence. Mine ideas and benchmark against Stonecraft/Stonecutter before any direct dependency decision.
- **Material risks:** Research-grade maturity; Must not replace proven builds solely for convenience

### Bytecode editors

#### [Recaf](https://recaf.coley.software/)

- **ID:** `recaf`
- **Maturity/status:** stable / active
- **Priority:** 95
- **Repository:** [https://github.com/Col-E/Recaf](https://github.com/Col-E/Recaf)
- **Best use:** Interactive investigation and scripted narrow repairs of closed-source JARs.
- **Capability:** Modern bytecode/source workspace, decompilers, assemblers, scripting/API and headless operations.
- **Vault integration:** Use a pinned headless/scripted path for reproducibility; preserve originals and emit entry-level diffs and verification reports.
- **Notes:** Excellent interactive repair workbench and potential headless analysis integration.

### Bytecode indexing and classpath analysis

#### [ModLens MCP](https://github.com/CreeperHost/modlens-mcp)

- **ID:** `modlens-mcp`
- **Maturity/status:** beta / active
- **Priority:** 100
- **Best use:** Giving the Vault and coding agents a searchable local corpus of the exact mods and Minecraft sources being repaired.
- **Capability:** Persistent mod/JAR metadata, class, member, dependency, Mixin, AT/AW and decompiled-source indexing with CLI and MCP access, cross-mod queries, diffs and crash analysis.
- **Vault integration:** Use embedded SQLite by default; ingest exact JARs by hash, call scoped CLI/MCP actions from Doctor reports, and keep the Vault as the authority for mutation/rollback.
- **Current evidence:** Active through 2026-08-16; recent work includes Node 26 support, SQLite template repair, crash facts and remote modpack ingestion.
- **Notes:** Strongest current external analysis/indexing candidate. Embedded SQLite is the zero-setup path; Vineflower and its bytecode indexer are acquired on demand. Inspect exact release and schema before enabling write actions.
- **Material risks:** Large MCP tool schemas consume context unless profiles disable unused tools; Optional semantic search adds Ollama and database complexity; External decompiler/indexer downloads must be hash-pinned

#### [ClassGraph](https://github.com/classgraph/classgraph)

- **ID:** `classgraph`
- **Maturity/status:** stable / active
- **Priority:** 85
- **Best use:** Duplicate classes, shaded libraries, entrypoints and cross-JAR ownership maps.
- **Capability:** Fast classpath/module scanning, annotation discovery and dependency metadata inspection.
- **Vault integration:** Build a complete class ownership index and retain collisions as first-class compatibility findings.
- **Notes:** Candidate for deep dependency graph and side-only class analysis.

### Bytecode inspection and transformation tools

#### [JDK javap, jdeps, jar, and jarsigner](https://docs.oracle.com/en/java/javase/25/docs/specs/man/)

- **ID:** `jdk-tools`
- **Maturity/status:** official / active
- **Priority:** 100
- **Best use:** Class descriptors, dependency analysis, archive integrity, signatures, and Java level
- **Vault integration:** cli-adapter
- **Notes:** Always available in a correct JDK toolchain and essential for independent verification.

#### [Access Widener](https://github.com/FabricMC/access-widener)

- **ID:** `access-widener`
- **Maturity/status:** stable / active
- **Priority:** 95
- **Best use:** Fabric access widening format and transformer
- **Vault integration:** library-adapter
- **Notes:** Required for correct Fabric source and binary reconstruction.

#### [MixinExtras](https://github.com/LlamaLad7/MixinExtras/wiki)

- **ID:** `mixinextras`
- **Maturity/status:** stable / active
- **Priority:** 95
- **Repository:** [https://github.com/LlamaLad7/MixinExtras](https://github.com/LlamaLad7/MixinExtras)
- **Best use:** Replacing brittle redirects or local-capture patterns during a source port.
- **Capability:** Additional expressive and safer injection primitives layered onto Mixin.
- **Vault integration:** Detect bundled/runtime versions and only rewrite injections when target semantics and loader packaging are verified.
- **Notes:** Prefer recognized MixinExtras semantics over brittle custom redirects when source can be repaired.

#### [JarSplitter](https://github.com/neoforged/JarSplitter)

- **ID:** `jarsplitter`
- **Maturity/status:** stable / active
- **Priority:** 80
- **Best use:** Splits classes and resources for modern toolchain pipelines
- **Vault integration:** cli-adapter
- **Notes:** Useful in NeoForge and mapping preparation workflows.

#### [MixinSquared](https://github.com/Bawnorton/MixinSquared)

- **ID:** `mixinsquared`
- **Maturity/status:** experimental / active
- **Priority:** 65
- **Best use:** Mixins targeting other mixins for compatibility patches
- **Vault integration:** runtime-adapter
- **Notes:** Powerful last-resort compatibility mechanism; keep scope narrow and test exact competing transforms.

### Bytecode libraries

#### [ObjectWeb ASM](https://asm.ow2.io/)

- **ID:** `asm`
- **Maturity/status:** stable / active
- **Priority:** 100
- **Repository:** [https://gitlab.ow2.org/asm/asm](https://gitlab.ow2.org/asm/asm)
- **Best use:** Deterministic class structure inspection, descriptor changes, reference graphs and narrowly proven bytecode patches.
- **Capability:** Low-level JVM class parsing, verification, transformation and analysis.
- **Vault integration:** Use structured visitors instead of raw string replacement, run CheckClassAdapter and compare intended class-level changes.
- **Notes:** Foundation for deterministic class repair and verification.

#### [SpongePowered Mixin source](https://github.com/SpongePowered/Mixin)

- **ID:** `mixin-source`
- **Maturity/status:** stable / active
- **Priority:** 100
- **Best use:** Exact annotation retention, injection semantics, refmaps, services, and bytecode weaving contracts
- **Vault integration:** source-analysis
- **Notes:** Inspect source and classfiles together when wiki guidance is insufficient; annotation processor and runtime contracts are both migration-critical.

### Bytecode references and guides

#### [SpongePowered Mixin wiki](https://github.com/SpongePowered/Mixin/wiki)

- **ID:** `mixin-wiki`
- **Maturity/status:** official / active
- **Priority:** 100
- **Repository:** [https://github.com/SpongePowered/Mixin](https://github.com/SpongePowered/Mixin)
- **Best use:** Mixin injection, selectors, refmaps, obfuscation, debugging, and compatibility
- **Vault integration:** reference
- **Notes:** Primary reference for mixin target and injection repairs.

### Compatibility libraries

#### [Balm](https://github.com/TwelveIterations/Balm)

- **ID:** `balm`
- **Maturity/status:** stable / active
- **Priority:** 82
- **Best use:** Reference architecture or dependency for maintainable source-level loader convergence.
- **Capability:** Common interfaces, events, configs, networking and third-party integrations for Fabric, Forge and NeoForge source mods using loader-native build plugins.
- **Vault integration:** Use as an explicit library dependency in source ports and test every loader artifact separately.
- **Current evidence:** Battle-tested across a large family of maintained mods.
- **Notes:** Strong reference architecture for loader abstraction and third-party integrations. It is a source library, not a binary cross-loader converter.
- **Material risks:** Not a binary cross-loader bridge; Adds a runtime library dependency

#### [Porting Lib](https://github.com/Fabricators-of-Create/Porting-Lib)

- **ID:** `porting-lib`
- **Maturity/status:** stable / active
- **Priority:** 80
- **Best use:** Reducing repeated Forge-to-Fabric rewrites when the needed subsystem is already implemented.
- **Capability:** Forge-inspired utilities and compatibility abstractions for Fabric.
- **Vault integration:** Select subsystem-by-subsystem and verify API/version support instead of treating it as a universal Forge compatibility layer.
- **Notes:** Useful for source ports that depend on Forge concepts, especially Create-derived ecosystems.

### Crash, log, and runtime diagnostics

#### [Crash Assistant](https://github.com/KostromDan/Crash-Assistant)

- **ID:** `crash-assistant`
- **Maturity/status:** stable / active
- **Priority:** 95
- **Best use:** Reliable end-user crash intake and expanding the Vault signature/remediation corpus.
- **Capability:** Collects game, launcher and JVM crash artifacts, analyzes known causes, routes reports, identifies environment problems and performs selected common fixes.
- **Vault integration:** Import only evidence-backed signatures and fixes into the Repair Brain; keep complete logs and the earliest causal chain.
- **Current evidence:** Actively maintained multi-branch Fabric/Forge/NeoForge project with shared app/config code and version-specific shims.
- **Notes:** Strong end-user crash intake companion. Import its proven signatures and remediation evidence, while retaining the Vault repair brain as the canonical deduplicated history.
- **Material risks:** A matched warning is not necessarily the terminal root cause; Auto-fixes must remain reversible and version-gated

#### [Sinytra Probe](https://github.com/Sinytra/Probe)

- **ID:** `sinytra-probe`
- **Maturity/status:** beta / active
- **Priority:** 94
- **Best use:** Collecting unsupported symbols and translation gaps from runtime-bridge failures.
- **Capability:** Inspection and probing utilities in the Sinytra compatibility ecosystem.
- **Vault integration:** Normalize findings into the Vault's compatibility knowledge graph and link them to exact versions.
- **Notes:** Probe 0.1.64 powered the July 2026 Connector 3.0.0-beta.1 compatibility run. Its maintainers explicitly warn that successful patching does not prove in-game behavior.

#### [MixinTrace](https://github.com/comp500/mixintrace)

- **ID:** `mixintrace`
- **Maturity/status:** stable / active
- **Priority:** 90
- **Best use:** Adds mixin attribution to stack traces
- **Vault integration:** runtime-adapter
- **Notes:** High-value crash diagnosis input for automatic root-cause ranking.

#### [mclo.gs](https://mclo.gs/)

- **ID:** `mclogs`
- **Maturity/status:** stable / active
- **Priority:** 85
- **Repository:** [https://github.com/aternosorg/mclogs](https://github.com/aternosorg/mclogs)
- **Best use:** Turning fresh launch logs into normalized failure signatures and shareable evidence.
- **Capability:** Minecraft log parsing, paste hosting and structured analyzer integrations.
- **Vault integration:** Keep local/private analysis primary; optionally use structured analyzer rules and preserve exact source lines in repair records.
- **Notes:** Use as an optional log normalization and sharing adapter.

#### [StackDeobfuscator](https://github.com/Bawnorton/StackDeobfuscator)

- **ID:** `stackdeobfuscator`
- **Maturity/status:** stable / active
- **Priority:** 85
- **Best use:** Maps obfuscated stack traces at runtime
- **Vault integration:** runtime-adapter
- **Notes:** Useful for older mapped environments and crash ingestion.

#### [Not Enough Crashes](https://github.com/natanfudge/Not-Enough-Crashes)

- **ID:** `not-enough-crashes`
- **Maturity/status:** stable / active
- **Priority:** 75
- **Best use:** Improved crash handling and reports in Fabric environments
- **Vault integration:** runtime-adapter
- **Notes:** Useful supplemental client diagnostics where supported.

### Cross-loader and runtime compatibility bridges

#### [Forgified Fabric API](https://github.com/Sinytra/ForgifiedFabricAPI)

- **ID:** `forgified-fabric-api`
- **Maturity/status:** stable / active
- **Priority:** 100
- **Best use:** Fabric API-dependent mods running through Connector.
- **Capability:** NeoForge-side implementation/port of Fabric API needed by Connector workflows.
- **Vault integration:** Resolve exact Connector-compatible versions and test only the modules required by the target graph.
- **Notes:** Core dependency for Connector-based compatibility.

#### [Sinytra Connector](https://connector.sinytra.org/)

- **ID:** `sinytra-connector`
- **Maturity/status:** stable / active
- **Priority:** 100
- **Repository:** [https://github.com/Sinytra/Connector](https://github.com/Sinytra/Connector)
- **Best use:** A controlled temporary or permanent runtime path where the exact mod graph is supported.
- **Capability:** Runs many Fabric mods on NeoForge through translation and compatibility layers.
- **Vault integration:** Probe exact combinations in isolated profiles, capture incompatibilities and keep source-port plans for unsupported behavior.
- **Current evidence:** 1.21.1 is the primary supported line; 1.20.1 receives long-term critical fixes.
- **Notes:** Best current high-capability Fabric to NeoForge runtime bridge. Treat compatibility as per-mod evidence, not universal proof.
- **Material risks:** Not every Fabric API or mixin pattern is compatible; Bridge success is not equivalent to a converted mod

#### [Retromod](https://bownlux.github.io/Retromod/)

- **ID:** `retromod`
- **Maturity/status:** experimental / active
- **Priority:** 90
- **Repository:** [https://github.com/Bownlux/Retromod](https://github.com/Bownlux/Retromod)
- **Best use:** Binary-only simple content, library and quality-of-life mods when a source port is unavailable.
- **Capability:** Transforms older mod bytecode, mappings, Mixins, access wideners/transformers, metadata, nested JARs and selected runtime contracts for newer Minecraft hosts.
- **Vault integration:** Use its standalone CLI in a sandbox, require a compatibility-database match, compare every changed entry, and then launch a copied client/server profile.
- **Current evidence:** Active snapshot development observed 2026-08-18 with a compatibility database and broad Fabric/Forge/NeoForge host matrix.
- **Notes:** Highest-ambition binary compatibility project found. Treat its compatibility database and exact tuple tests as mandatory gates; complex renderer, loader and save-affecting mods still need source ports or targeted adapters.
- **Material risks:** Renderer, loader-internal and save-affecting mods may require real source ports; Broad runtime transformation claims demand tuple-specific verification; Never apply directly to the only copy of a modpack or world

#### [Connector Extras](https://github.com/Sinytra/ConnectorExtras)

- **ID:** `connector-extras`
- **Maturity/status:** experimental / active
- **Priority:** 80
- **Best use:** Known gaps such as config/menu, recipe viewer or energy/API interoperability covered by a released module.
- **Capability:** Additional compatibility bridges for selected APIs and mod integrations around Sinytra Connector.
- **Vault integration:** Select explicit modules from detected dependencies and keep each as an auditable compatibility decision.
- **Notes:** Use only when exact mod evidence supports the patch set.

#### [Kilt](https://github.com/KiltMC/Kilt)

- **ID:** `kilt`
- **Maturity/status:** experimental / active
- **Priority:** 75
- **Best use:** Research and isolated compatibility probes when a source port is not yet available.
- **Capability:** Attempts to run Forge/NeoForge mods on Fabric.
- **Vault integration:** Never apply automatically to a real profile. Clone the instance, test world safety, collect exact incompatibilities and preserve a source-port fallback.
- **Current evidence:** Current releases remain explicitly experimental.
- **Notes:** High-value opposite-direction bridge, but verify supported Minecraft and Forge API coverage per build.
- **Material risks:** May crash or corrupt worlds; Coverage varies sharply by mod and Forge API surface

#### [ReForged](https://github.com/Arc-Stuido/ReForged)

- **ID:** `reforged`
- **Maturity/status:** experimental / active
- **Priority:** 55
- **Best use:** NeoForge mod compatibility on Forge 1.21.1 through bytecode transforms and shims
- **Vault integration:** research-only
- **Notes:** Bleeding-edge research candidate. Restrictive licensing and narrow target mean no code transplant into Vault.

#### [Insanity Wrapper](https://github.com/InsanityLabs/Wrapper)

- **ID:** `insanity-wrapper`
- **Maturity/status:** experimental / active
- **Priority:** 45
- **Best use:** Fabric and Quilt compatibility layer on NeoForge for newer versions
- **Vault integration:** research-only
- **Notes:** Research candidate with restrictive license and manual mapping workflow.

#### [Patchwork Patcher](https://github.com/PatchworkMC/patchwork-patcher)

- **ID:** `patchwork-patcher`
- **Maturity/status:** legacy / archived
- **Priority:** 30
- **Best use:** Historical Fabric to Forge binary patching research
- **Vault integration:** reference
- **Notes:** Do not adopt as a current engine; mine architecture and failure lessons only.

### Data and NBT libraries

#### [Querz NBT](https://github.com/Querz/NBT)

- **ID:** `querz-nbt`
- **Maturity/status:** stable / active
- **Priority:** 80
- **Best use:** Java NBT and region file library
- **Vault integration:** library-adapter
- **Notes:** Candidate for direct world and structure repair modules.

### Decompilers

#### [Vineflower](https://vineflower.org/)

- **ID:** `vineflower`
- **Maturity/status:** stable / active
- **Priority:** 100
- **Repository:** [https://github.com/Vineflower/vineflower](https://github.com/Vineflower/vineflower)
- **Best use:** Primary reproducible decompilation of closed-source Minecraft mods.
- **Capability:** Modern Java decompiler derived from Fernflower with strong support for current bytecode constructs.
- **Vault integration:** Decompile with pinned options, retain line maps and compare output against a second decompiler where constructs are ambiguous.
- **Current evidence:** Version 1.12.0 was the latest release observed during the 2026-08-20 research pass.
- **Notes:** Best overall current decompiler candidate for source recovery pipelines.

#### [CFR](https://www.benf.org/other/cfr/)

- **ID:** `cfr`
- **Maturity/status:** stable / active
- **Priority:** 90
- **Repository:** [https://github.com/leibnitz27/cfr](https://github.com/leibnitz27/cfr)
- **Best use:** Cross-checking switches, lambdas, generics, exception flow and synthetic constructs.
- **Capability:** Independent Java decompiler with different reconstruction heuristics from Vineflower.
- **Vault integration:** Run as a disagreement detector rather than blindly selecting one decompiler output.
- **Notes:** Use as a second decompiler to cross-check ambiguous Vineflower output.

### Dependency and update automation

#### [Renovate](https://github.com/renovatebot/renovate)

- **ID:** `renovate`
- **Maturity/status:** stable / active
- **Priority:** 90
- **Best use:** Automated Gradle, GitHub Actions, Java, and custom dependency updates
- **Vault integration:** automation-adapter
- **Notes:** Strong foundation for continuous loader, mappings, and library update PRs.

#### [Gradle Versions Plugin](https://github.com/ben-manes/gradle-versions-plugin)

- **ID:** `gradle-versions-plugin`
- **Maturity/status:** stable / active
- **Priority:** 80
- **Best use:** Reports dependency and plugin updates
- **Vault integration:** cli-adapter
- **Notes:** Use as an inventory source, not as the final compatibility decision.

#### [Hopper](https://github.com/oraxen/hopper)

- **ID:** `hopper`
- **Maturity/status:** beta / active
- **Priority:** 72
- **Best use:** Designing the Vault server-plugin dependency and update lane.
- **Capability:** Platform-aware runtime server-plugin dependency resolution from Hangar, Modrinth, Spiget, GitHub Releases and direct URLs with range/update policies.
- **Vault integration:** Borrow provider-resolution and policy ideas; stage downloads and verify hot-load behavior in disposable Paper/Folia/Velocity profiles.
- **Current evidence:** Current repository documentation covers multiple providers and platform detection.
- **Notes:** Useful design source for the server-plugin lane. Hot-loading and third-party dependency resolution require isolated runtime and rollback tests before integration.
- **Material risks:** Runtime hot-loading has platform-specific failure modes; Remote dependency policies require provenance and rollback

### Developer productivity tools

#### [Minecraft Development IntelliJ Plugin](https://github.com/minecraft-dev/MinecraftDev)

- **ID:** `minecraftdev`
- **Maturity/status:** stable / active
- **Priority:** 90
- **Best use:** IDE support for loaders, mappings, mixins, access transformers, and run configs
- **Vault integration:** developer-adapter
- **Notes:** Useful in generated source projects and manual repair review.

### Instance architecture references

#### [Prism Launcher wiki](https://prismlauncher.org/wiki/)

- **ID:** `prism-wiki`
- **Maturity/status:** official / active
- **Priority:** 90
- **Repository:** [https://github.com/PrismLauncher/PrismLauncher](https://github.com/PrismLauncher/PrismLauncher)
- **Best use:** Instance layout, mod management, logs, Java, loaders, and pack import/export
- **Vault integration:** reference
- **Notes:** Excellent real-instance integration target for Vault repair workflows.

### Language runtimes

#### [Fabric Language Kotlin](https://github.com/FabricMC/fabric-language-kotlin)

- **ID:** `fabric-language-kotlin`
- **Maturity/status:** stable / active
- **Priority:** 80
- **Best use:** Kotlin runtime and adapter for Fabric mods
- **Vault integration:** dependency-adapter
- **Notes:** Detect Kotlin metadata before source migration or bytecode rewriting.

### Launchers and instance managers

#### [Prism Launcher](https://github.com/PrismLauncher/PrismLauncher)

- **ID:** `prism-launcher`
- **Maturity/status:** stable / active
- **Priority:** 100
- **Best use:** Open-source multi-instance launcher with loaders, packs, logs, and Java management
- **Vault integration:** instance-adapter
- **Notes:** Primary desktop instance integration target.

#### [Modrinth App](https://github.com/modrinth/code)

- **ID:** `modrinth-app`
- **Maturity/status:** stable / active
- **Priority:** 95
- **Best use:** Open-source launcher and Modrinth instance manager
- **Vault integration:** instance-adapter
- **Notes:** Useful for profile format and update behavior comparisons.

#### [ATLauncher](https://github.com/ATLauncher/ATLauncher)

- **ID:** `atlauncher`
- **Maturity/status:** stable / active
- **Priority:** 85
- **Best use:** Open-source launcher and pack manager
- **Vault integration:** instance-adapter
- **Notes:** Important additional instance format and provider workflow.

#### [GDLauncher Carbon](https://github.com/gorilla-devs/GDLauncher-Carbon)

- **ID:** `gdlauncher`
- **Maturity/status:** stable / active
- **Priority:** 80
- **Best use:** Modern open-source Minecraft launcher
- **Vault integration:** instance-adapter
- **Notes:** Useful for multi-launcher import and update compatibility.

### Loader APIs

#### [Fabric API](https://fabricmc.net/develop/)

- **ID:** `fabric-api`
- **Maturity/status:** stable / active
- **Priority:** 100
- **Repository:** [https://github.com/FabricMC/fabric](https://github.com/FabricMC/fabric)
- **Best use:** Determining whether a Fabric port depends on loader-only APIs or Fabric API modules.
- **Capability:** Core event, registry, networking, rendering, lifecycle and game-test APIs used by a large share of Fabric mods.
- **Vault integration:** Map module-level dependencies, detect API removals and run version-specific compile/runtime tests.
- **Current evidence:** Actively maintained for current Minecraft releases.
- **Notes:** Use source and changelog diffs to map API drift.

### Mappings references

#### [ParchmentMC documentation](https://parchmentmc.org/docs/)

- **ID:** `parchment-docs`
- **Maturity/status:** official / active
- **Priority:** 90
- **Repository:** [https://github.com/ParchmentMC](https://github.com/ParchmentMC)
- **Best use:** Parameter names and Javadocs layered on official mappings
- **Vault integration:** reference
- **Notes:** Particularly valuable after 26.1 because documentation remains useful even when obfuscation ends.

### Mappings tools

#### [mapping-io](https://github.com/FabricMC/mapping-io)

- **ID:** `mapping-io`
- **Maturity/status:** stable / active
- **Priority:** 100
- **Best use:** Building source-to-target mapping graphs across official, intermediary, Yarn and custom mappings.
- **Capability:** Read, write, transform and compose mapping formats and namespaces.
- **Vault integration:** Normalize every mapping source into one graph, retain provenance and reject unresolved or ambiguous symbols.
- **Notes:** Core namespace graph library for legacy and cross-loader work.

#### [MCPConfig](https://github.com/MinecraftForge/MCPConfig)

- **ID:** `mcpconfig`
- **Maturity/status:** stable / active
- **Priority:** 95
- **Best use:** Forge mapping and patch configuration for Minecraft versions
- **Vault integration:** data-adapter
- **Notes:** Core legacy Forge namespace reconstruction source.

#### [Enigma](https://github.com/FabricMC/Enigma)

- **ID:** `enigma`
- **Maturity/status:** stable / active
- **Priority:** 90
- **Best use:** Resolving ambiguous decompiler output and authoring missing mappings during closed-source reconstruction.
- **Capability:** Interactive deobfuscation and mapping editor with source/class views.
- **Vault integration:** Open only unresolved symbol clusters and export reviewed mappings back into the project mapping graph.
- **Notes:** Still valuable for legacy versions and undocumented third-party binaries.

#### [Lorenz](https://github.com/CadixDev/Lorenz)

- **ID:** `lorenz`
- **Maturity/status:** stable / active
- **Priority:** 80
- **Best use:** Mapping model and transformations
- **Vault integration:** library-adapter
- **Notes:** Useful for format-independent mapping operations.

### Minecraft data catalogs

#### [PrismarineJS minecraft-data](https://github.com/PrismarineJS/minecraft-data)

- **ID:** `minecraft-data`
- **Maturity/status:** stable / active
- **Priority:** 90
- **Best use:** Machine-readable protocol, blocks, items, entities, recipes, and version data
- **Vault integration:** data-adapter
- **Notes:** Strong cross-version structured data source, especially outside loader APIs.

#### [Burger](https://github.com/Pokechu22/Burger)

- **ID:** `burger`
- **Maturity/status:** stable / active
- **Priority:** 75
- **Best use:** Extracts structured information from Minecraft jars
- **Vault integration:** data-adapter
- **Notes:** Useful for historical metadata extraction and cross-checking generated catalogs.

### Mod loaders

#### [Fabric Loader](https://fabricmc.net/use/)

- **ID:** `fabric-loader`
- **Maturity/status:** stable / active
- **Priority:** 100
- **Repository:** [https://github.com/FabricMC/fabric-loader](https://github.com/FabricMC/fabric-loader)
- **Best use:** Exact Fabric runtime contract validation and clean target profiles.
- **Capability:** Mostly version-independent Fabric loader, metadata validation, dependency resolution and entrypoint discovery.
- **Vault integration:** Resolve supported loader versions from Fabric Meta, validate fabric.mod.json and launch clean compatibility matrices.
- **Current evidence:** Actively maintained through July 2026.
- **Notes:** Inspect loader internals only when metadata and public APIs are insufficient.

#### [MinecraftForge](https://github.com/MinecraftForge/MinecraftForge)

- **ID:** `minecraftforge`
- **Maturity/status:** stable / active
- **Priority:** 100
- **Best use:** Forge loader and API implementation
- **Vault integration:** source-adapter
- **Notes:** Primary source for Forge symbols and loader behavior.

#### [NeoForge](https://github.com/neoforged/NeoForge)

- **ID:** `neoforge`
- **Maturity/status:** stable / active
- **Priority:** 100
- **Best use:** NeoForge loader and API implementation
- **Vault integration:** source-adapter
- **Notes:** Primary source for NeoForge symbols, behavior, tests, and migration verification.

#### [Quilt Loader](https://quiltmc.org/en/)

- **ID:** `quilt-loader`
- **Maturity/status:** stable / active
- **Priority:** 85
- **Repository:** [https://github.com/QuiltMC/quilt-loader](https://github.com/QuiltMC/quilt-loader)
- **Best use:** Quilt-native target validation and Fabric-on-Quilt compatibility testing.
- **Capability:** Fabric-compatible loader with Quilt metadata and loader-specific dependency semantics.
- **Vault integration:** Validate quilt.mod.json, dependency groups and Fabric compatibility in isolated target profiles.
- **Current evidence:** Quilt announced its non-obfuscated-era direction in February 2026.
- **Notes:** Treat Quilt Loader as active while QSL and new Quilt Mappings are retired.

### Modpack format references

#### [packwiz documentation](https://packwiz.infra.link/)

- **ID:** `packwiz-docs`
- **Maturity/status:** official / active
- **Priority:** 90
- **Repository:** [https://github.com/packwiz/packwiz](https://github.com/packwiz/packwiz)
- **Best use:** Git-friendly pack metadata, updates, side declarations, exports, and installer flows
- **Vault integration:** reference
- **Notes:** Useful as an interchange and regression fixture format.

### Modpack management tools

#### [packwiz](https://packwiz.infra.link/)

- **ID:** `packwiz`
- **Maturity/status:** stable / active
- **Priority:** 100
- **Repository:** [https://github.com/packwiz/packwiz](https://github.com/packwiz/packwiz)
- **Best use:** Reproducible dependency manifests, pack upgrades and rollback-friendly instance composition.
- **Capability:** Declarative, version-controlled modpack management with Modrinth/CurseForge metadata and refresh/update workflows.
- **Vault integration:** Import/export manifests, reuse CurseForge fingerprint logic, solve provider metadata and verify final pack state.
- **Notes:** Best git-friendly pack metadata engine to integrate or interoperate with.

#### [Ferium](https://github.com/gorilla-devs/ferium)

- **ID:** `ferium`
- **Maturity/status:** stable / active
- **Priority:** 95
- **Best use:** A comparative implementation and command-line fallback for straightforward provider updates.
- **Capability:** CLI mod manager for upgrading projects from Modrinth, CurseForge and GitHub releases.
- **Vault integration:** Benchmark identity/update behavior and borrow only proven provider-resolution ideas while retaining the Vault's deeper repair stages.
- **Notes:** Strong updater benchmark and potential provider-resolution adapter.

#### [MCMan](https://github.com/ParadigmMC/mcman)

- **ID:** `mcman`
- **Maturity/status:** stable / active
- **Priority:** 90
- **Best use:** Declarative Minecraft server and modpack management
- **Vault integration:** cli-adapter
- **Notes:** Useful server-side pack definition and reproducibility benchmark.

#### [Modrinth Pack Version Updater](https://github.com/KrisTC/mrpack-updater)

- **ID:** `mrpack-updater`
- **Maturity/status:** experimental / active
- **Priority:** 45
- **Best use:** Checks a Modrinth pack against another Minecraft version
- **Vault integration:** research-only
- **Notes:** Use as a test corpus and UX reference, not the core solver.

### Multi-loader architecture references

#### [Architectury documentation](https://docs.architectury.dev/)

- **ID:** `architectury-docs`
- **Maturity/status:** official / active
- **Priority:** 90
- **Repository:** [https://github.com/architectury](https://github.com/architectury)
- **Best use:** Cross-loader project structure, common code, Architectury Loom, and platform APIs
- **Vault integration:** reference
- **Notes:** Useful for source-level loader convergence, not a magic binary converter.

### Network and protocol compatibility

#### [Geyser](https://github.com/GeyserMC/Geyser)

- **ID:** `geyser`
- **Maturity/status:** stable / active
- **Priority:** 85
- **Best use:** Bedrock to Java protocol translation
- **Vault integration:** runtime-adapter
- **Notes:** Useful for Bedrock compatibility test matrices and server integration.

#### [ViaVersion](https://github.com/ViaVersion/ViaVersion)

- **ID:** `viaversion`
- **Maturity/status:** stable / active
- **Priority:** 85
- **Best use:** Minecraft protocol translation across versions
- **Vault integration:** runtime-adapter
- **Notes:** Useful for protocol compatibility analysis, not mod API conversion.

#### [ViaFabricPlus](https://github.com/ViaVersion/ViaFabricPlus)

- **ID:** `viafabricplus`
- **Maturity/status:** stable / active
- **Priority:** 80
- **Best use:** Fabric client protocol translation and old-version behavior compatibility
- **Vault integration:** runtime-adapter
- **Notes:** Useful for cross-version connection testing, not a substitute for porting mods.

### Official specifications

#### [Minecraft Java release notes](https://www.minecraft.net/en-us/articles)

- **ID:** `minecraft-release-notes`
- **Maturity/status:** official / active
- **Priority:** 100
- **Best use:** Authoritative game changes, pack versions, Java requirements, and release dates
- **Vault integration:** reference
- **Notes:** Use exact release pages when building an adjacent-version migration chain.

#### [Mojang version manifest v2](https://piston-meta.mojang.com/mc/game/version_manifest_v2.json)

- **ID:** `minecraft-version-manifest`
- **Maturity/status:** official / active
- **Priority:** 100
- **Best use:** Machine-readable Java version and asset metadata
- **Vault integration:** api
- **Notes:** Use as the canonical version catalog and never infer latest versions from filenames.

### Parser libraries

#### [Brigadier](https://github.com/Mojang/brigadier)

- **ID:** `brigadier`
- **Maturity/status:** official / active
- **Priority:** 80
- **Best use:** Minecraft command parsing and syntax trees
- **Vault integration:** library-adapter
- **Notes:** Useful for command migration and validation.

### Patch generation and application

#### [BinaryPatcher](https://github.com/MinecraftForge/BinaryPatcher)

- **ID:** `binarypatcher`
- **Maturity/status:** stable / active
- **Priority:** 80
- **Best use:** Forge binary patch creation and application
- **Vault integration:** cli-adapter
- **Notes:** Useful for deterministic distribution of narrow binary deltas.

#### [DiffPatch](https://github.com/neoforged/DiffPatch)

- **ID:** `diffpatch`
- **Maturity/status:** stable / active
- **Priority:** 80
- **Best use:** Source and archive patch generation and application
- **Vault integration:** cli-adapter
- **Notes:** Candidate for reproducible migration patch bundles.

### Porting and migration guides

#### [Fabric developer documentation](https://docs.fabricmc.net/)

- **ID:** `fabric-docs`
- **Maturity/status:** official / active
- **Priority:** 100
- **Repository:** [https://github.com/FabricMC/fabric-docs](https://github.com/FabricMC/fabric-docs)
- **Best use:** Version-correct source transformations and explanation links in repair plans.
- **Capability:** Official build, mappings, data generation, networking, rendering and migration documentation.
- **Vault integration:** Index versioned migration pages and attach the exact relevant documentation to generated actions.
- **Notes:** Pair with Fabric API source and changelogs for exact removed symbols.

#### [NeoForge migration primers](https://docs.neoforged.net/primer/)

- **ID:** `neoforge-primers`
- **Maturity/status:** official / active
- **Priority:** 100
- **Repository:** [https://github.com/neoforged/Documentation](https://github.com/neoforged/Documentation)
- **Best use:** Recipe-backed upgrades and downgrades across major Minecraft API transitions.
- **Capability:** Official change guides describing Java, mappings, registry, rendering, networking, data and API changes between Minecraft versions.
- **Vault integration:** Convert each primer delta into testable migration rules while retaining the original source link and target range.
- **Current evidence:** The 1.21.11 to 26.1 primer documents Java 25 and the return to unobfuscated game executables.
- **Notes:** Best maintained public chain of vanilla and NeoForge migration notes, including 1.21.11 to 26.1 and 26.1.x to 26.2.

#### [MinecraftForge documentation](https://docs.minecraftforge.net/en/latest/)

- **ID:** `forge-docs`
- **Maturity/status:** official / active
- **Priority:** 95
- **Repository:** [https://github.com/MinecraftForge/Documentation](https://github.com/MinecraftForge/Documentation)
- **Best use:** Validating Forge-specific source and metadata migrations.
- **Capability:** Official Forge metadata, registries, events, networking, resources and lifecycle documentation.
- **Vault integration:** Link exact documentation sections to generated findings and code transformations.
- **Notes:** Coverage varies by Minecraft generation; inspect Forge source and MDKs when docs lag.

#### [Misode Minecraft version changelog](https://misode.github.io/changelog/)

- **ID:** `misode-changelog`
- **Maturity/status:** community / active
- **Priority:** 95
- **Repository:** [https://github.com/misode/misode.github.io](https://github.com/misode/misode.github.io)
- **Best use:** Data pack, resource pack, registry, command, and JSON format changes
- **Vault integration:** reference
- **Notes:** Best paired with Mojang release notes and generated reports.

#### [QuiltMC documentation and status](https://quiltmc.org/)

- **ID:** `quilt-site`
- **Maturity/status:** official / active
- **Priority:** 80
- **Repository:** [https://github.com/QuiltMC](https://github.com/QuiltMC)
- **Best use:** Quilt Loader, Enigma, configuration, and current project status
- **Vault integration:** reference
- **Notes:** Quilt Mappings and QSL are retired for new versions; retain for legacy repair only.

### Profilers

#### [spark](https://spark.lucko.me/)

- **ID:** `spark`
- **Maturity/status:** stable / active
- **Priority:** 95
- **Repository:** [https://github.com/lucko/spark](https://github.com/lucko/spark)
- **Best use:** Performance repair after correctness and compatibility are established.
- **Capability:** Minecraft profiler for CPU, memory, tick, heap and server/client performance investigations.
- **Vault integration:** Capture comparable before/after workloads and reject fixes that merely reduce content or fidelity.
- **Notes:** Required evidence source for performance repair without feature reduction.

#### [Observable](https://github.com/tobspr-games/Observable)

- **ID:** `observable`
- **Maturity/status:** stable / active
- **Priority:** 75
- **Best use:** In-game tick-time profiling and visualization
- **Vault integration:** runtime-adapter
- **Notes:** Useful for entity and block-entity hot spot diagnosis.

### Project templates

#### [Jared's MultiLoader Template](https://github.com/jaredlll08/MultiLoader-Template)

- **ID:** `multiloader-template`
- **Maturity/status:** stable / active
- **Priority:** 85
- **Best use:** Bootstrapping source ports that need transparent loader modules and shared logic.
- **Capability:** Practical common-source project template for Fabric, Forge and NeoForge without a required third-party runtime library.
- **Vault integration:** Use as a generated starting topology, then transplant only verified project behavior and version-specific APIs.
- **Notes:** Use as a comparison target, not blindly as a replacement for project-specific build logic.

### Provider APIs and release metadata

#### [Modrinth API documentation](https://docs.modrinth.com/api/)

- **ID:** `modrinth-api`
- **Maturity/status:** official / active
- **Priority:** 100
- **Repository:** [https://github.com/modrinth/code](https://github.com/modrinth/code)
- **Best use:** Exact SHA identity, compatible release resolution and dependency staging.
- **Capability:** Project/version search, exact hash lookup, update lookup, dependency metadata and package downloads.
- **Vault integration:** Use SHA-512 identity first, cache versioned metadata, verify published hashes/size and preserve source/project lineage.
- **Notes:** Use SHA-512 identity and version metadata for exact update and dependency plans.

#### [GitHub REST releases API](https://docs.github.com/en/rest/releases/releases)

- **ID:** `github-releases`
- **Maturity/status:** official / active
- **Priority:** 99
- **Repository:** [https://github.com/github/rest-api-description](https://github.com/github/rest-api-description)
- **Best use:** Release assets, tags, version lineage, checksums, and source/binary recovery
- **Vault integration:** api
- **Notes:** The current versioned API documents release asset SHA-256 digests and latest/tag endpoints; do not equate a tag with a published release.

#### [GitHub REST repository contents API](https://docs.github.com/en/rest)

- **ID:** `github-api`
- **Maturity/status:** official / active
- **Priority:** 98
- **Repository:** [https://github.com/](https://github.com/)
- **Best use:** Resolving canonical source and ports not represented by Modrinth or CurseForge.
- **Capability:** Source discovery, tags, commits, release assets, checksums, issues, pull requests and CI evidence.
- **Vault integration:** Require explicit repository identity and target-version/loader evidence before treating an asset as an update.
- **Notes:** Use versioned REST requests and preserve commit SHA plus content digest for reproducible source recovery.

#### [CurseForge for Studios API](https://docs.curseforge.com/)

- **ID:** `curseforge-api`
- **Maturity/status:** official / active
- **Priority:** 95
- **Repository:** [https://github.com/CurseForgeCommunity](https://github.com/CurseForgeCommunity)
- **Best use:** Exact MurmurHash2 identity and target-file resolution for CurseForge-hosted mods.
- **Capability:** Project/file search, fingerprint matching, dependency metadata and downloads where permitted by project distribution settings.
- **Vault integration:** Use authenticated exact fingerprints, verify file/game/loader relations and retain required-dependency changes for explicit review.
- **Notes:** Requires configured credentials for reliable automated resolution.

### Publishing and release automation

#### [mc-publish](https://github.com/Kir-Antipov/mc-publish)

- **ID:** `mc-publish`
- **Maturity/status:** stable / active
- **Priority:** 90
- **Best use:** Publishes mod artifacts to multiple Minecraft platforms from CI
- **Vault integration:** cli-adapter
- **Notes:** Useful for reconstructed mod release pipelines and proof artifacts.

#### [Minotaur](https://github.com/modrinth/minotaur)

- **ID:** `minotaur`
- **Maturity/status:** stable / active
- **Priority:** 85
- **Best use:** Official Modrinth Gradle publishing plugin
- **Vault integration:** cli-adapter
- **Notes:** Use when rebuilding source projects that publish to Modrinth.

#### [CurseGradle](https://github.com/matthewprenger/CurseGradle)

- **ID:** `cursegradle`
- **Maturity/status:** stable / active
- **Priority:** 80
- **Best use:** Gradle publishing integration for CurseForge
- **Vault integration:** cli-adapter
- **Notes:** Use for compatible legacy and current CurseForge publication workflows.

### Source analysis

#### [JavaParser](https://github.com/javaparser/javaparser)

- **ID:** `javaparser`
- **Maturity/status:** stable / active
- **Priority:** 90
- **Best use:** Java AST parsing, symbol solving, editing, and code generation
- **Vault integration:** library-adapter
- **Notes:** Good for focused transformations where full OpenRewrite recipes are unnecessary.

#### [Spoon](https://github.com/INRIA/spoon)

- **ID:** `spoon`
- **Maturity/status:** stable / active
- **Priority:** 90
- **Best use:** Typed Java metamodel, analysis, transformations, and pretty printing
- **Vault integration:** library-adapter
- **Notes:** Strong alternative for semantic source repairs and mining migration patterns.

### Source conversion

#### [MC Mod Porter](https://github.com/reqsery/mc-mod-porter)

- **ID:** `mc-mod-porter`
- **Maturity/status:** beta / active
- **Priority:** 88
- **Best use:** Bootstrapping adjacent-version source ports and harvesting verified migration recipes.
- **Capability:** Version-chain rules, build.gradle/gradle.properties/metadata rewrites, Java source patches, resource migrations, build attempts, static debugging and visual testing helpers.
- **Vault integration:** Run only on a copied source project, ingest its verified rule corpus, compare every edit, then compile and exercise the actual mod.
- **Current evidence:** Observed beta 1.1.3 with active July/August 2026 commits and version-chain migration work through 26.2.
- **Notes:** Useful rule corpus and source-workspace automation. Use only against extracted source projects; logic redesigns and third-party API changes still require evidence-backed engineering. Observed beta 1.1.3.
- **Material risks:** Known-pattern rewriting cannot prove semantic equivalence; Third-party mod APIs and major logic redesigns remain manual; Input must be a source project, not a compiled JAR

#### [ModMorpher](https://github.com/Indozilla1234/Modmorpher)

- **ID:** `modmorpher`
- **Maturity/status:** experimental / active
- **Priority:** 35
- **Best use:** Compiled-JAR decompile and conversion experiments
- **Vault integration:** research-only
- **Notes:** Useful for ideas and test cases, not a trusted production converter.

### Source remapping

#### [ModForge](https://github.com/champmk/modforge)

- **ID:** `modforge`
- **Maturity/status:** beta / active
- **Priority:** 100
- **Best use:** Source ports where false-positive renames are more dangerous than explicit unresolved findings.
- **Capability:** Deterministic class/member migration across Minecraft versions using real mappings and target JAR verification, API deltas, Gradle migration, Mixin checks and an MCP server.
- **Vault integration:** Run bridge/delta/mixin-check in a copied source workspace; auto-apply only EXACT findings, preserve CANDIDATE and UNRESOLVED evidence, and expose its MCP tools to the porting agent.
- **Current evidence:** Observed v0.1.2 with current 2026 work, Node 24+, deterministic audit chains and a 170-test suite.
- **Notes:** Current high-capability leader for evidence-backed mapping migration. Its EXACT/CANDIDATE/UNRESOLVED model is suitable for guarded automation; observed release 0.1.2 on 2026-08-20.
- **Material risks:** Semantic API redesigns are not automatically solved; Current strongest lane centers on the 1.21.x to 26.x naming boundary; Requires target artifacts and a warmed verified cache for offline operation

#### [Mercury](https://github.com/CadixDev/Mercury)

- **ID:** `mercury`
- **Maturity/status:** stable / active
- **Priority:** 75
- **Best use:** Java source transformation and remapping
- **Vault integration:** library-adapter
- **Notes:** Candidate for source-level namespace conversions after decompilation.

#### [Srg2Source](https://github.com/MinecraftForge/Srg2Source)

- **ID:** `srg2source`
- **Maturity/status:** legacy / maintenance
- **Priority:** 60
- **Best use:** Legacy Forge source range and mapping transformations
- **Vault integration:** cli-adapter
- **Notes:** Keep for older Forge recovery only.

### Source transformation

#### [OpenRewrite](https://docs.openrewrite.org/)

- **ID:** `openrewrite`
- **Maturity/status:** stable / active
- **Priority:** 100
- **Repository:** [https://github.com/openrewrite/rewrite](https://github.com/openrewrite/rewrite)
- **Best use:** Automating known package, symbol, signature and build migrations while keeping diffs reviewable.
- **Capability:** Loss-minimizing Java source transformation recipes with type attribution and repeatable Gradle/Maven integration.
- **Vault integration:** Maintain Minecraft-versioned recipes, apply in dry-run mode, compile after each recipe group and retain provenance for every edit.
- **Notes:** Best overall engine candidate for verified Minecraft API migration recipes.

#### [Error Prone and Refaster](https://github.com/google/error-prone)

- **ID:** `error-prone`
- **Maturity/status:** stable / active
- **Priority:** 80
- **Best use:** Compile-time bug checks and template-based Java rewrites
- **Vault integration:** cli-adapter
- **Notes:** Refaster can encode compact API replacement templates after rules are verified.

### Static analysis

#### [Spyglass](https://github.com/SpyglassMC/Spyglass)

- **ID:** `spyglass`
- **Maturity/status:** beta / active
- **Priority:** 92
- **Best use:** Validating JSON, commands, NBT-like data and resource/data-pack migrations before runtime.
- **Capability:** Minecraft data/resource language server and parser infrastructure using mcdoc schemas for diagnostics, completion and version-aware structured-text analysis.
- **Vault integration:** Run target-version validation across changed resources and feed precise diagnostics into Doctor findings and generated migration tests.
- **Current evidence:** Active organization/repository observed in August 2026.
- **Notes:** Best current structured-text validation candidate for resource/data migration. Use exact target-version schemas and treat semantic world behavior as a separate runtime test.
- **Material risks:** Schema validation cannot prove gameplay semantics; Version schema selection must be exact

#### [CodeQL](https://github.com/github/codeql)

- **ID:** `codeql`
- **Maturity/status:** stable / active
- **Priority:** 85
- **Best use:** Deep semantic code queries and data flow analysis
- **Vault integration:** cli-adapter
- **Notes:** Use for high-risk native, reflection, network, serialization, and injection analysis.

#### [Semgrep](https://github.com/semgrep/semgrep)

- **ID:** `semgrep`
- **Maturity/status:** stable / active
- **Priority:** 85
- **Best use:** Pattern and semantic static analysis across source trees
- **Vault integration:** cli-adapter
- **Notes:** Useful for migration anti-pattern rules and security scanning.

### Version matrix and source-set management

#### [Stonecutter](https://stonecutter.kikugie.dev/)

- **ID:** `stonecutter`
- **Maturity/status:** stable / active
- **Priority:** 95
- **Repository:** [https://github.com/kikugie/stonecutter](https://github.com/kikugie/stonecutter)
- **Best use:** One mod that must build old and new Minecraft targets without cloning the entire codebase.
- **Capability:** Preprocessor and Gradle workflow for maintaining source variants across many Minecraft versions.
- **Vault integration:** Generate explicit version branches/conditions, compile every supported target in CI and reject dead conditional paths.
- **Notes:** One of the strongest tools for maintaining many Minecraft versions from one repository.

### Wikis and knowledge sources

#### [Minecraft Repair Brain](https://drive.google.com/drive/folders/1SEI9cZQHMEalLcKTMqpXGbfNJDZ1zWbS)

- **ID:** `repair-brain`
- **Maturity/status:** internal / active
- **Priority:** 100
- **Repository:** [https://github.com/Herbertofury/Minecraft-Mod-Vault](https://github.com/Herbertofury/Minecraft-Mod-Vault)
- **Best use:** Preventing repeated failed repairs and reusing proven fixes only on matching version tuples.
- **Capability:** Append-only, versioned repair cases covering signatures, exact artifacts, root causes, fixes, failed attempts, supersession, verification and user-confirmed outcomes.
- **Vault integration:** Search before every diagnosis; update the canonical JSONL only after preserving evidence, hashes and applicability.
- **Current evidence:** Five recovered real repair records were loaded into the v0.8.0 Doctor classifier on 2026-08-20.
- **Notes:** Canonical user-owned repair memory. Search before every repair and append only after evidence is preserved; never replace failed or superseded history with a cleaned-up success narrative.
- **Material risks:** Never promote partial static evidence to runtime success; Never overwrite failed or superseded attempts

#### [Minecraft Wiki](https://minecraft.wiki/)

- **ID:** `minecraft-wiki`
- **Maturity/status:** community / active
- **Priority:** 85
- **Best use:** Data-only migrations and cross-checking version-specific schema changes.
- **Capability:** Detailed version history, data/resource pack formats, NBT, registries, commands, world data and technical mechanics.
- **Vault integration:** Use alongside official release notes and loader primers; store exact page/revision access dates for migration rules.
- **Notes:** Strong secondary technical reference; confirm loader-specific claims in loader docs.

### World and serialized-data migration

#### [DataFixerUpper](https://github.com/Mojang/DataFixerUpper)

- **ID:** `datafixerupper`
- **Maturity/status:** official / active
- **Priority:** 100
- **Best use:** Schema-driven serialized game data upgrades
- **Vault integration:** library-adapter
- **Notes:** Primary model for forward world and data migration. Downgrades need separate reversible transforms.

#### [mcmeta](https://github.com/misode/mcmeta)

- **ID:** `mcmeta`
- **Maturity/status:** stable / active
- **Priority:** 100
- **Best use:** Generated vanilla reports, registries, tags, recipes, and version diffs
- **Vault integration:** data-adapter
- **Notes:** Excellent machine-readable source for assets and data-pack migration rules.

### World, region, NBT, and save repair tools

#### [Amulet Map Editor and Core](https://github.com/Amulet-Team/Amulet-Map-Editor)

- **ID:** `amulet`
- **Maturity/status:** stable / active
- **Priority:** 90
- **Repository:** [https://github.com/Amulet-Team/Amulet-Core](https://github.com/Amulet-Team/Amulet-Core)
- **Best use:** Java and Bedrock world editing and format translation
- **Vault integration:** cli-adapter
- **Notes:** Strong base for world-level repair and conversion on copied saves.

#### [NBT Studio](https://github.com/tryashtar/nbt-studio)

- **ID:** `nbt-studio`
- **Maturity/status:** stable / active
- **Priority:** 82
- **Best use:** Human verification of copied worlds, structures and repaired NBT payloads.
- **Capability:** GUI viewer/editor for Java NBT, Java region files, Bedrock little-endian NBT and SNBT with undo/redo, multiselect and format conversion.
- **Vault integration:** Open only copied artifacts after deterministic repair, compare semantic paths, and preserve automated before/after hashes separately.
- **Current evidence:** Current maintained successor to NBTExplorer with Java and Bedrock support.
- **Notes:** Modern manual verification companion to NBTExplorer. Keep automated repairs deterministic and use NBT Studio for human inspection of copied saves and structures.
- **Material risks:** Manual edits are not reproducible unless exported as a patch receipt; Region/world changes still require runtime validation

#### [MCA Selector](https://github.com/Querz/mcaselector)

- **ID:** `mcaselector`
- **Maturity/status:** stable / active
- **Priority:** 80
- **Best use:** Region inspection, selection, filtering, and deletion
- **Vault integration:** cli-adapter
- **Notes:** Useful for targeted region recovery and reproducible world test fixtures.

#### [Chunker](https://chunker.app/)

- **ID:** `chunker`
- **Maturity/status:** stable / active
- **Priority:** 75
- **Best use:** Java and Bedrock world conversion
- **Vault integration:** external-adapter
- **Notes:** Treat as an external conversion lane and verify exact output versions.

#### [NBTExplorer](https://github.com/jaquadro/NBTExplorer)

- **ID:** `nbtexplorer`
- **Maturity/status:** stable / active
- **Priority:** 70
- **Best use:** Interactive NBT inspection and editing
- **Vault integration:** developer-adapter
- **Notes:** Useful manual verification tool for save and config repairs.

<!-- MMV-EXPANDED-TOOLCHAIN-START -->
## Expanded modern, legacy, and classic toolchain intelligence

These records are the execution-oriented additions from the 2026-08-20 audit. They supplement the original complete catalog and are embedded in the Doctor UI/API. Runtime compatibility products are deliberately separated from source reconstruction and semantic porting.

### Data Migration Research

#### [PaperMC DataConverter](https://github.com/PaperMC/DataConverter)

- **ID:** `papermc-dataconverter`
- **Maturity:** research
- **Priority:** 65
- **Capability:** High-performance rewrite of Minecraft data conversion with explicit walkers and debuggable conversion code.
- **Best use:** Researching faster conversion architecture, diffing vanilla schema behavior and constructing focused data fixtures.
- **Vault integration:** Index converter implementations and test vectors as reference only; never dispatch its Fabric mod against a modded world.
- **Current evidence:** The official README explicitly warns that mod-registered data fixers are skipped and modded-world data can be corrupted.
- **Material risks:** Unsafe for modded worlds; Not a general downgrade engine; Architecture findings must be reimplemented with complete mod datafixer coverage

### Dependency Packaging

#### [ModShade](https://github.com/mezz/ModShade)

- **ID:** `modshade`
- **Maturity:** stable
- **Priority:** 90
- **Capability:** Relocates and shades plain Java libraries while preserving unshaded debug artifacts and clean publication metadata.
- **Best use:** Bundling non-mod Java dependencies that would otherwise conflict on a shared mod classpath.
- **Vault integration:** Shade only explicitly classified plain libraries, retain unshaded outputs, scan relocated packages, and test service/resource merges.
- **Current evidence:** Current documentation specifies plugin 0.5.0, Gradle 8.3+, Java 17+, and configuration-cache support.
- **Material risks:** Never embed another Minecraft mod with it; Relocation can break reflection, services or serialized class names

### Legacy Build Remap

#### [RetroFuturaGradle](https://github.com/GTNewHorizons/RetroFuturaGradle)

- **ID:** `retrofuturagradle`
- **Maturity:** stable
- **Priority:** 99
- **Capability:** Modern Gradle replacement for legacy ForgeGradle 1.7.10 development with toolchains, caches, dependency deobfuscation and reliable reobfuscation.
- **Best use:** Rebuilding and porting source-owned Minecraft 1.7.10 Forge mods.
- **Vault integration:** Generate or migrate an isolated 1.7.10 workspace, build the untouched source, run deobfuscation/reobfuscation checks, and compare output JAR structure.
- **Current evidence:** 2.0.2 is the latest verified release; 2.0 added Gradle 9 support and a Java 25 build requirement.
- **Material risks:** Build JVM requirements differ from game runtime requirements; Legacy coremods still require runtime transformer tests

#### [Ornithe Ploceus](https://github.com/OrnitheMC/ploceus)

- **ID:** `ornithe-ploceus`
- **Maturity:** stable
- **Priority:** 94
- **Capability:** Extends Fabric/Quilt Loom with Ornithe mappings, loader support and legacy Minecraft workspace behavior.
- **Best use:** Source-owned Ornithe ports and reproducible ancient-version target builds.
- **Vault integration:** Generate from the official template, pin the full mapping/loader/plugin/JDK tuple, then run client/server/remap tasks.
- **Current evidence:** Pinned official Ornithe tooling and actively maintained organization in August 2026.
- **Material risks:** Legacy versions vary widely in structure and launch behavior; Mappings and loader support must be verified for the exact release

#### [Legacy Looming](https://github.com/Legacy-Fabric/legacy-looming)

- **ID:** `legacy-fabric-looming`
- **Maturity:** stable
- **Priority:** 90
- **Capability:** Legacy Fabric extension of Loom for old Minecraft mappings, setup and remapping.
- **Best use:** Reproducible Legacy Fabric workspaces where normal current Loom does not support the target.
- **Vault integration:** Pin Legacy Looming with Legacy Intermediaries/Yarn/API and prove decompile, remapJar, client and server tasks.
- **Current evidence:** Repository updated in May 2026 within the active Legacy Fabric organization.
- **Material risks:** Toolchain tuples are tightly coupled; Historical Java and launcher behavior vary by version

#### [FPGradle](https://plugins.gradle.org/plugin/com.falsepattern.fpgradle-mc)

- **ID:** `fpgradle`
- **Maturity:** stable
- **Priority:** 84
- **Repository:** [https://github.com/FalsePattern/FPGradle](https://github.com/FalsePattern/FPGradle)
- **Capability:** Declarative Gradle plugin for Minecraft 1.7.10 mod development.
- **Best use:** Alternative modernized source build path for selected 1.7.10 Forge projects.
- **Vault integration:** Benchmark clean builds and reobfuscation against RetroFuturaGradle, then select the path that reproduces the upstream artifact with fewer project-specific patches.
- **Current evidence:** 2.1.1 was published 2025-10-26 and remains the current verified plugin version.
- **Material risks:** Smaller ecosystem than RetroFuturaGradle; Do not churn a working build solely for DSL preference

### Legacy Compatibility Patches

#### [Fugue](https://github.com/CleanroomMC/Fugue)

- **ID:** `fugue`
- **Maturity:** beta
- **Priority:** 90
- **Capability:** Targeted Cleanroom-era compatibility patches for mods that otherwise fail on modern Java or the revised 1.12.2 runtime.
- **Best use:** Known per-mod fixes after a Cleanroom baseline identifies exact failures.
- **Vault integration:** Match exact mod/version tuples, enable only applicable patches, and verify the original reproducer plus regression paths.
- **Current evidence:** Repository was actively pushed on 2026-08-12.
- **Material risks:** Patch coverage is finite and version-sensitive; A compatibility patch can hide a deeper semantic break

#### [LegacyFix](https://github.com/betacraftuk/legacyfix)

- **ID:** `legacyfix`
- **Maturity:** stable
- **Priority:** 88
- **Capability:** Version-aware runtime/coremod patches for legacy Minecraft across Forge, Fabric and direct agent-style installations.
- **Best use:** Reproducing and fixing launcher/runtime failures in very old versions before attempting source conversion.
- **Vault integration:** Choose installation mode by exact game/loader version, retain an unpatched profile, and compare startup/network/resource behavior.
- **Current evidence:** Current repository provides stable and nightly builds with loader-specific installation instructions.
- **Material risks:** Patch behavior differs substantially by game version; Runtime repair is not a source port

#### [Hodgepodge](https://github.com/GTNewHorizons/Hodgepodge)

- **ID:** `hodgepodge`
- **Maturity:** stable
- **Priority:** 82
- **Capability:** Configurable bug, performance and Java-compatibility fixes for large Minecraft 1.7.10 modpacks.
- **Best use:** Pack-level compatibility gaps that LWJGL3ify intentionally leaves to mod-specific patches.
- **Vault integration:** Enable only relevant fixes, record configuration, run the failing feature, then test adjacent mods and restart persistence.
- **Current evidence:** Explicitly recommended by LWJGL3ify for multiple Java 17+ compatibility cases.
- **Material risks:** Large patch surface can mask root causes; Pack-specific configuration needs regression evidence

### Legacy Java Runtime

#### [LWJGL3ify](https://github.com/GTNewHorizons/lwjgl3ify)

- **ID:** `lwjgl3ify`
- **Maturity:** stable
- **Priority:** 93
- **Capability:** LWJGL2-to-LWJGL3 compatibility, modern-Java bootstrap, native/runtime replacement and targeted ASM/Mixin fixes for 1.7.10.
- **Best use:** Modernizing the runtime environment of large 1.7.10 packs without pretending the mods were source-ported.
- **Vault integration:** Build a separate Prism/MultiMC profile, install required UniMixins/Hodgepodge components, validate all OS/native paths, then exercise rendering, input, sound and server startup.
- **Current evidence:** Active current line documents modern Java and bundles RetroFuturaBootstrap.
- **Material risks:** Graphics/native behavior is platform-specific; Old reflection and coremod assumptions still need per-mod patches

### Legacy Loader Api

#### [Legacy Fabric API](https://legacyfabric.net/)

- **ID:** `legacy-fabric-api`
- **Maturity:** stable
- **Priority:** 92
- **Repository:** [https://github.com/Legacy-Fabric/fabric](https://github.com/Legacy-Fabric/fabric)
- **Capability:** Fabric-style hooks and interoperability modules for Minecraft 1.3.x through 1.13.2.
- **Best use:** Native Legacy Fabric source ports and API compatibility analysis.
- **Vault integration:** Read the per-version module matrix, map actual imported APIs, and compile/test only against modules marked for the selected target.
- **Current evidence:** Repository remained active in June 2026 and documents module support by Minecraft version.
- **Material risks:** API coverage is not uniform across versions; Modern Fabric documentation may not apply

#### [StationAPI](https://github.com/ModificationStation/StationAPI)

- **ID:** `stationapi`
- **Maturity:** stable
- **Priority:** 91
- **Capability:** General API ecosystem for Babric/Fabric mods on Minecraft Beta 1.7.3.
- **Best use:** Native beta 1.7.3 mod development and compatibility repairs.
- **Vault integration:** Generate from the documented example path, keep StationAPI as an explicit dependency, and run the Babric profile on the required Java version.
- **Current evidence:** Repository updated August 2026 and documents current Beta 1.7.3 Babric usage.
- **Material risks:** Only applies to its specific legacy ecosystem; Sparse documentation requires source inspection and fixtures

### Legacy Loader Runtime

#### [Cleanroom](https://cleanroommc.com/)

- **ID:** `cleanroom`
- **Maturity:** beta
- **Priority:** 92
- **Repository:** [https://github.com/CleanroomMC/Cleanroom](https://github.com/CleanroomMC/Cleanroom)
- **Capability:** Minecraft 1.12.2 loader/runtime continuation with modern Java, LWJGL3, Mixin and compatibility patches.
- **Best use:** Testing whether 1.12.2 packs can run on modern Java before or alongside source repair.
- **Vault integration:** Create a copied profile, install exact Cleanroom/Fugue versions, run client/server/world/restart matrices, and store mod-by-mod compatibility evidence.
- **Current evidence:** Active CleanroomMC project in July-August 2026; README advertises Java 25+ and LWJGL3.
- **Material risks:** Broad compatibility percentages are maintainer claims, not proof for a specific pack; Coremods and graphics mods need dedicated coverage

### Legacy Mappings

#### [Ornithe Feather](https://github.com/OrnitheMC/feather)

- **ID:** `ornithe-feather`
- **Maturity:** stable
- **Priority:** 96
- **Capability:** CC0 mappings and version graph for Minecraft c0.0.12a_03 through 1.14.4, with mapping and dual-decompiler tasks.
- **Best use:** Ancient game-source comparison and named legacy workspaces outside modern mapping coverage.
- **Vault integration:** Resolve the exact Feather graph node, map to named, compare CFR/Vineflower outputs, and retain unmapped/ambiguous symbols.
- **Current evidence:** Active Ornithe mapping project with coverage from classic alpha through 1.14.4.
- **Material risks:** Decompiler output is not necessarily recompilable; Mapping propagation across a graph must preserve confidence and provenance

### Legacy Mixin Runtime

#### [UniMixins](https://github.com/LegacyModdingMC/UniMixins)

- **ID:** `unimixins`
- **Maturity:** stable
- **Priority:** 95
- **Capability:** Modular compatibility layer combining legacy Mixin loader features without forcing one monolithic provider.
- **Best use:** Minecraft 1.7.10 Mixin compatibility and selected 1.8.9-1.12.2 cases.
- **Vault integration:** Resolve required modules from artifact metadata, reject duplicate providers, and execute the project's published compatibility fixtures plus pack-specific tests.
- **Current evidence:** 0.3.1 released 2026-05-27 with signed release assets and an ASM remapper fix.
- **Material risks:** Module overlap must be solved explicitly; Partial support above 1.7.10 requires tuple-specific proof

#### [MixinBooter](https://github.com/CleanroomMC/MixinBooter)

- **ID:** `mixinbooter`
- **Maturity:** stable
- **Priority:** 92
- **Capability:** Unified Mixin configuration discovery, CleanMix/MixinExtras integration and dedicated diagnostics for 1.8-1.12.2.
- **Best use:** Porting or stabilizing Mixin-heavy legacy Forge mods and resolving competing Mixin loaders.
- **Vault integration:** Detect existing Mixin providers, select one compatible backend, register configs explicitly, and retain mixinbooter.log and trace output.
- **Current evidence:** Current 11.x line uses CleanMix and supports 1.8 through 1.12.2.
- **Material risks:** Multiple embedded Mixin versions can still conflict; Blacklist toggles are diagnostic, not a substitute for fixing broken injections

### Legacy Runtime Bootstrap

#### [RetroFuturaBootstrap](https://github.com/GTNewHorizons/RetroFuturaBootstrap)

- **ID:** `retrofuturabootstrap`
- **Maturity:** stable
- **Priority:** 94
- **Capability:** Backwards-compatible LaunchWrapper replacement with ordered compatibility plugins and class/transformer dump diagnostics.
- **Best use:** Modern-Java legacy runtime bootstrapping and diagnosing early transformer incompatibilities.
- **Vault integration:** Launch disposable profiles with class-dump flags, retain per-transformer outputs, and compare class ordering before and after compatibility plugins.
- **Current evidence:** Active and used by LWJGL3ify as the early-load foundation.
- **Material risks:** Bootstrap success does not prove mod behavior; Compatibility plugins can introduce ordering conflicts

### Legacy Source Recovery

#### [RetroMCP-Java](https://github.com/NeRdTheNed/RetroMCP-Java)

- **ID:** `retromcp-java`
- **Maturity:** stable
- **Priority:** 96
- **Capability:** Version-profile-driven download, deobfuscation, decompilation, patching, recompilation and reobfuscation for many classic Minecraft versions.
- **Best use:** Recovering reviewable source and build workspaces for alpha, beta and old release mods.
- **Vault integration:** Select an exact version profile, record mappings/decompiler/patches, build the untouched baseline, then layer port changes with output diffs.
- **Current evidence:** v1.2 is the current verified release; the project remains the broadest maintained MCP-style legacy reconstruction option found.
- **Material risks:** Coverage and fidelity differ by historical version; Decompiled source requires ambiguity review

### Multi Loader Packaging

#### [Forgix](https://plugins.gradle.org/plugin/io.github.pacifistmc.forgix)

- **ID:** `forgix`
- **Maturity:** stable
- **Priority:** 88
- **Repository:** [https://github.com/PacifistMC/Forgix](https://github.com/PacifistMC/Forgix)
- **Capability:** Merges loader-specific Fabric/Quilt/Forge-family outputs into one distributable JAR.
- **Best use:** Packaging already verified multi-loader builds while retaining native artifacts.
- **Vault integration:** Run after every native loader build passes; inspect merged metadata/resources/services and launch every loader from a fresh profile.
- **Current evidence:** 1.3.4 is the current stable plugin release; 2.0.0 is available only as snapshots.
- **Material risks:** Merged JARs can conceal resource or service collisions; A merged artifact cannot replace native-loader verification

### Publishing

#### [Mod Publish Plugin](https://plugins.gradle.org/plugin/me.modmuss50.mod-publish-plugin)

- **ID:** `mod-publish-plugin`
- **Maturity:** stable
- **Priority:** 96
- **Repository:** [https://github.com/modmuss50/mod-publish-plugin](https://github.com/modmuss50/mod-publish-plugin)
- **Capability:** Gradle artifact publication to common Minecraft destinations with configuration-cache support.
- **Best use:** Source-owned projects that need reproducible multi-destination publication after verification.
- **Vault integration:** Generate publication configuration from verified build metadata, publish only signed/checksummed artifacts, and re-download remote objects for hashing.
- **Current evidence:** 2.1.1 was published 2026-06-25 and listed as current on 2026-08-20.
- **Material risks:** Upload acknowledgement is not byte verification; Credentials and provider metadata must remain external to project memory

### Source Diff Research

#### [Ornithe Gitcraft](https://github.com/OrnitheMC/gitcraft)

- **ID:** `ornithe-gitcraft`
- **Maturity:** research
- **Priority:** 85
- **Capability:** Generates decompiled Minecraft history repositories using multiple mapping and manifest sources.
- **Best use:** Finding exact adjacent-version implementation/API changes across modern and legacy versions.
- **Vault integration:** Generate private repositories with selected mappings, compare adjacent commits, export only evidence/patch recipes, and never publish generated source trees.
- **Current evidence:** Current tool supports Mojmap, Parchment, Yarn, Intermediary, Calamus and Feather mapping modes plus historic manifests.
- **Material risks:** Generated repositories must not be shared; Decompiler differences can create false diffs

### Source Reconstruction

#### [NeoForm](https://github.com/neoforged/NeoForm)

- **ID:** `neoform`
- **Maturity:** official
- **Priority:** 100
- **Capability:** Creates reproducible and recompilable modern Minecraft sources and adjacent-version patch workspaces.
- **Best use:** Exact modern game-source reconstruction and tracking upstream changes across releases.
- **Vault integration:** Create isolated patch workspaces, preserve the exact branch/tag/config, apply updates with reject capture, and hash every generated artifact.
- **Current evidence:** Active NeoForged project; current branches distinguish obfuscated pre-26.1 inputs from unobfuscated Minecraft generations.
- **Material risks:** Generated game source is reference/build input, not automatically publishable mod source; Branch selection is version-sensitive

#### [NeoForm Runtime](https://github.com/neoforged/NeoFormRuntime)

- **ID:** `neoform-runtime`
- **Maturity:** official
- **Priority:** 100
- **Capability:** Standalone execution-graph CLI for deobfuscation, merging, patching, recompilation, assets and artifact resolution.
- **Best use:** Deterministic, auditable Minecraft/NeoForge artifact generation outside a full Gradle plugin.
- **Vault integration:** Invoke with pinned NeoForm or NeoForge coordinates, print the graph, redirect every named result, and retain caches/manifests/logs as provenance.
- **Current evidence:** Actively maintained in the NeoForge toolchain through July 2026.
- **Material risks:** Large cached artifact graph must be isolated per tuple; Custom repositories and manifests can change reproducibility

### Source Reconstruction Research

#### [MCP-Reborn](https://github.com/Hexeption/MCP-Reborn)

- **ID:** `mcp-reborn`
- **Maturity:** research
- **Priority:** 80
- **Capability:** MCPConfig/ForgeGradle-based game-source research workspaces for Minecraft 1.13 through current generations.
- **Best use:** Inspecting exact game changes and validating symbol assumptions where official/loader tooling is insufficient.
- **Vault integration:** Use in an isolated research workspace, record version and JDK requirements, and extract only independently described API findings.
- **Current evidence:** Repository advertises coverage through 26.2 and documents the JDK transition through Java 25.
- **Material risks:** Generated code is explicitly non-publishable; Research output must not be copied into release artifacts

### Version Metadata Api

#### [Fabric Meta](https://meta.fabricmc.net/)

- **ID:** `fabric-meta`
- **Maturity:** official
- **Priority:** 100
- **Repository:** [https://github.com/FabricMC/fabric-meta](https://github.com/FabricMC/fabric-meta)
- **Capability:** Frequently refreshed JSON API for Fabric game, loader, intermediary, mappings, installer and launcher metadata.
- **Best use:** Resolving exact compatible Fabric version tuples instead of scraping download pages.
- **Vault integration:** Consume V2 endpoints, cache response hashes/timestamps, retain stable flags and Maven coordinates, and fail closed on incompatible tuples.
- **Current evidence:** Official service states metadata is refreshed every five minutes; repository was active in July 2026.
- **Material risks:** Newest is not always stable; Cached metadata must carry a freshness timestamp

### Workspace Generator

#### [NeoForge Mod Generator](https://neoforged.net/mod-generator)

- **ID:** `neoforge-mod-generator`
- **Maturity:** official
- **Priority:** 92
- **Repository:** [https://github.com/neoforged/mod-generator](https://github.com/neoforged/mod-generator)
- **Capability:** Official browser-embeddable and CLI-core project generator for clean NeoForge workspaces.
- **Best use:** Creating a known-current target skeleton before migrating recovered source.
- **Vault integration:** Generate into a disposable branch, lock all selected versions, diff output against templates, then transplant source through compiler-driven steps.
- **Current evidence:** Active official NeoForged generator with shared web/CLI generator core.
- **Material risks:** Generated defaults still require target-specific review; A scaffold is not a completed port

<!-- MMV-EXPANDED-TOOLCHAIN-END -->

## Verified repair patterns imported into Mod Doctor

### Binary owner, member, and descriptor drift

- **ID:** `binary-owner-descriptor-drift`
- **Failure family:** NoSuchMethodError, NoSuchFieldError, AbstractMethodError, IncompatibleClassChangeError, or a dependency-owned class moved while behavior remained equivalent
- **Trigger:** The dependent bytecode references an owner/name/descriptor that the installed dependency no longer exports.
- **Repair:** Compare the exact dependent constant-pool reference with old and new provider bytecode. Prefer a source rebuild; otherwise patch only the proven owner/name/descriptor or add a tightly version-gated adapter while preserving the original artifact.
- **Verification:** Disassemble dependent call sites; Inspect old and new provider descriptors; Validate rebuilt class files; Run the exact original reproducer; Record hashes and applicability
- **Confidence:** high

### Data namespace, registry, codec, or NBT schema defect

- **ID:** `data-namespace-schema-defect`
- **Failure family:** Unknown registry key, malformed datapack, jigsaw/template-pool failure, codec error, or invalid NBT/resource namespace
- **Trigger:** A packaged or persisted identifier points to the wrong namespace or no longer matches the target schema.
- **Repair:** Correct only the proven resource identifier, codec payload, registry remap, or NBT path on an immutable copy. Preserve unrelated world and mod data.
- **Verification:** Parse every changed resource; Load a copied world; Exercise the affected registry or structure; Restart and compare persisted data
- **Confidence:** high

### Mixin annotation, target, and injection contract

- **ID:** `mixin-annotation-contract`
- **Failure family:** Mixin PREPARE/APPLY failure, invalid injection, missing target, wrong annotation retention, accessor/invoker mismatch, or competing transformer
- **Trigger:** Mixin metadata or bytecode no longer describes the exact target class and member contract.
- **Repair:** Inspect RuntimeVisibleAnnotations and RuntimeInvisibleAnnotations plus the exact target bytecode. Retarget classes, descriptors, injection points, slices, ordinals, plugin predicates, and priorities without blanket-disabling the feature.
- **Verification:** javap -verbose annotation check; Mixin diagnostic launch; Target bytecode comparison; Feature-level behavior test; Clean restart
- **Confidence:** high

### Loader metadata is not binary compatibility proof

- **ID:** `loader-metadata-not-proof`
- **Failure family:** A JAR declares an acceptable loader or dependency range but fails linkage or bootstrap
- **Trigger:** Metadata permits the combination while class, method, field, transformer, or lifecycle contracts differ.
- **Repair:** Treat metadata as a candidate filter only. Verify actual bytecode, dependency graph, loader services, access rules, mixins, and runtime behavior against the exact versions.
- **Verification:** Dependency graph scan; Binary API diff; Client and dedicated-server launch; Original feature reproducer
- **Confidence:** high

### Separate compatibility adapter

- **ID:** `separate-compatibility-adapter`
- **Failure family:** Two valid upstream mods expose incompatible public APIs or data contracts
- **Trigger:** Each upstream artifact works alone, but their integration layer no longer aligns.
- **Repair:** Build a separately versioned adapter that translates only the proven boundary. Keep both upstream JARs byte-identical and make supported version tuples explicit.
- **Verification:** Both upstream hashes unchanged; Adapter-only archive diff; Exact integration workflow exercised; Adapter removal restores baseline
- **Confidence:** medium-high

### Dependency graph and classpath collision resolution

- **ID:** `dependency-graph-resolution`
- **Failure family:** Duplicate mod IDs, missing required dependencies, declared conflicts, duplicate classes, or dependency cycles
- **Trigger:** The installed artifact graph cannot be topologically or uniquely resolved.
- **Repair:** Resolve every artifact to canonical identity, preserve dependency relationship types, reconcile required versions, remove exact duplicates, relocate conflicting shaded classes, and migrate cyclic components as one unit.
- **Verification:** Rescan all metadata; Hash every JAR; Recompute duplicate classes; Launch in a clean profile; Restart after successful validation
- **Confidence:** high

### Controlled source recovery with independent decompilers

- **ID:** `source-recovery-dual-decompiler`
- **Failure family:** The installed binary has no reproducible source repository or matching tag
- **Trigger:** Provider metadata and release history cannot reproduce the exact artifact.
- **Repair:** Preserve the original, decompile with at least two independent engines, remap with known mappings, compare control flow and resources, record ambiguity, and rebuild before changing semantics.
- **Verification:** Dual-decompiler comparison; Recompiled archive validation; Behavioral baseline comparison; Provenance and uncertainty ledger
- **Confidence:** medium

### Transactional batch migration and rollback

- **ID:** `transactional-batch-repair`
- **Failure family:** A modpack update or loader/version migration can leave a partially applied instance
- **Trigger:** Multiple interdependent artifacts, configs, scripts, datapacks, or worlds change together.
- **Repair:** Stage the complete batch, verify hashes and metadata, snapshot mutable state, apply in dependency order, run the runtime matrix, and restore the whole snapshot on failure.
- **Verification:** Pre/post inventory hashes; Atomic replacement receipt; Client/server/world matrix; Rollback drill
- **Confidence:** high

### Repair supersession and failed-attempt memory

- **ID:** `repair-supersession-memory`
- **Failure family:** A previous patch failed, partially worked, or was superseded
- **Trigger:** The same version tuple or failure signature reappears.
- **Repair:** Retain the earlier attempt and outcome, link the replacement, record exact versions/hashes, and require new evidence before reuse on another tuple.
- **Verification:** History record preserved; Replacement linked; Exact versions and hashes recorded; Observed user result captured
- **Confidence:** high

## Non-negotiable conversion rules

- A metadata edit is not a port. Class files, loader entrypoints, mappings, APIs, resources, data schemas, Java level, and runtime behavior must all agree with the target.
- Runtime bridges are candidate routes, not proof. Successful patching or startup does not establish gameplay correctness.
- Downgrades require explicit reverse transforms and loss analysis. DataFixerUpper primarily models forward migration.
- Mixin repairs require exact target classes, descriptors, injection points, priorities, refmaps, annotations, and competing transformer analysis.
- Binary patching is limited to proven narrow symbol or descriptor changes. Source rebuilds are preferred when available.
- Every replacement or port is staged, hashed, archive-tested, dependency-checked, launched in the target runtime, and retained only when the original reproducer and regression paths pass.

<!-- MMV-PORTING-LAB-090-START -->
## Minecraft Mod Vault 0.9.0 Porting Lab execution layer

The 0.9.0 release turns the catalog into an operational planning layer while keeping execution bounded by evidence.

### Built-in Version Atlas

`assets/version-atlas.json` contains 907 normalized Mojang version records plus loader/build-tool coverage and immutable source hashes. The application can answer source/target Java, mappings, client/server, protocol, data, and pack-format questions offline without inventing missing coordinates. A missing exact coordinate becomes a warning and blocked gate, not a guessed dependency.

### New independent evidence tools

#### InterMed

- **Repository:** https://github.com/jarettr/intermed
- **Maturity:** 0.1.4-alpha at the 2026-08-20 review.
- **Capability:** read-only dependency, resource, Mixin, security/SBOM, log, and performance analysis with findings traceable to exact files.
- **Vault role:** required preflight evidence candidate in Porting Lab.
- **Guardrail:** alpha output contracts can change; findings do not mutate or prove a repair.

#### modcrawl

- **Repository:** https://github.com/SirCesarium/modcrawl
- **Maturity:** early active Rust project.
- **Capability:** loader identification, metadata/dependencies, class Java levels, constant-pool grep, Mixin targets, duplicate classes, JSON, library, and C FFI.
- **Vault role:** fast independent cross-check against the native JAR forensic model.
- **Guardrail:** pin exact revisions and fixture-test parser behavior before using it as an execution dependency.

#### Modpack Inspector

- **Repository:** https://github.com/Rearth/Modpack-Inspector
- **Capability:** desktop pack inventory, dependency graphs, provider enrichment, config ownership clues, and log-oriented debugging.
- **Vault role:** human-facing comparison and pack-maintenance research.
- **Guardrail:** heuristic unused/config findings may not trigger automatic deletion.

#### ModLens MCP

- **Repository:** https://github.com/CreeperHost/modlens-mcp
- **Capability:** local searchable index of metadata, classes, Mixins, access rules, and decompiled source through CLI/MCP.
- **Vault role:** future optional evidence index for difficult multi-JAR investigations.
- **Guardrail:** MCP/AI output is not authoritative; preserve decompilation provenance and verify against bytecode/runtime.

#### ModpackResolver

- **Repository:** https://github.com/iTrooz/ModpackResolver
- **Capability:** compares selected-mod availability across Minecraft versions.
- **Vault role:** coarse candidate-version signal.
- **Guardrail:** availability is not API, dependency, loader, or runtime compatibility.

### Eight-gate migration workflow

1. **Preserve and identify** — hash inputs, extract metadata, and freeze provenance.
2. **Reconstruct** — reproduce the source environment or reconcile two decompilers against bytecode.
3. **Map namespaces** — build an explicit mapping graph and reject semantic misuse of remapping.
4. **Migrate APIs and loader contracts** — replace changed registrations, lifecycle, networking, rendering, config, capabilities/components, and side contracts.
5. **Migrate resources and data** — validate recipes, tags, models, codecs, registries, NBT/world schemas, pack formats, and downgrade loss.
6. **Build reproducibly** — pin Java, Gradle, mappings, loader, processors, dependencies, and archive contents.
7. **Verify the runtime matrix** — clean client, dedicated server, integrated server/world, feature reproducer, persistence/restart, multiplayer/network, and fresh logs as applicable.
8. **Deploy transactionally** — preserve originals, stage the candidate, retain evidence, and make rollback deterministic.

### Executable repair substrate

The Doctor’s first automatic mutation is exact-duplicate quarantine. It intentionally requires complete SHA-512 equality, regular managed files, unchanged bytes at apply time, destination hash verification, keeper re-verification, atomic receipts, and a hash-verified restore transaction. This substrate will be reused for broader repairs only after each new action class has equally strong preconditions and regression proof.
<!-- MMV-PORTING-LAB-090-END -->
