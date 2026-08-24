# Minecraft Mod Vault 0.9.0 release notes

## One release, two repair laboratories, one evidence brain

0.9.0 combines two previously separate development lines into one product instead of sacrificing either workflow:

- **Porting Lab** starts from a managed installed JAR and performs live artifact forensics, exact version/toolchain planning, isolated workspace generation, and cryptographically reversible duplicate repair.
- **Repair Lab** starts from a source-project ZIP and performs hostile-safe immutable intake, project profiling, conservative migration staging, controlled wrapper execution, artifact discovery, rollback, receipts, and proof-bundle export.
- **Compatibility Brain** is the shared offline SQLite evidence layer for official Minecraft/loader/toolchain versions, mappings, current repair research, source conversion, build eras, crash signatures, and durable Repair Brain history.

The distinction is intentional. Binary inspection, source migration, runtime compatibility bridges, metadata edits, semantic API rewrites, data migration, and final client/server verification are different engineering problems. The UI and backend now keep them separate while sharing evidence and exact target resolution.

## SQLite Compatibility Brain

The compiled application now embeds the reviewed seed corpus and initializes a pure-Go SQLite 3.53.3 database with WAL and FTS5 from any working directory. In the verified release runtime it indexed:

- **919** Minecraft versions;
- **949** loader/game relationships;
- **550** loader releases;
- **15,133** toolchain releases;
- **342** searchable knowledge documents;
- durable Repair Brain records that remain provenance-scoped and revalidated per target.

The database covers Mojang manifests, mcmeta formats, Fabric and Quilt metadata, Forge and NeoForge/toolchain Maven metadata, Modrinth version tags, the curated Doctor tool catalog, repair patterns, source references, and prior verified fixes. Search is exposed both as a real Repair Lab workspace and authenticated APIs. The database is reproducible from embedded source identities and seed hashes; it is not an opaque AI claim or a remote service dependency.

## Secure source Repair Lab

Repair Lab turns an uploaded source ZIP into a reversible engineering session:

- ZIP traversal, symbolic links, duplicate paths, archive bombs, suspicious compression, and extraction escapes are rejected;
- the original archive and extracted source tree receive immutable SHA-256 identities;
- build system, wrapper, project root, loader, Minecraft version, mappings, Java target, metadata, and pack formats are detected;
- upgrades and downgrades resolve through the embedded Version Atlas;
- only recognized version, loader, mapping, Java, metadata, and pack-format fields are automatically staged;
- every proposed edit records its file, old value, new value, reason, and evidence;
- build/test/clean actions use only a detected project wrapper and fixed action set;
- execution requires the exact build-script code acknowledgement, uses dedicated caches and a sanitized environment, supports timeout/cancel, and captures full logs;
- discovered artifacts are hashed and inspected; the immutable source is verified again after execution;
- rollback restores the working copy from the immutable source;
- prepared-source and proof-bundle exports preserve the plan, edits, logs, receipts, hashes, artifacts, and source identity.

This is deliberately honest containment. Version 0.9.0 does not claim that arbitrary third-party build scripts are operating-system/VM sandboxed, and it does not label a metadata-only migration as a completed semantic port.

## Runtime-discovered UI race corrected

Real Chromium verification found a race that unit and API tests did not: a forced status refresh after source upload could arrive while the initial Repair Lab status request was still loading, causing the new session to be temporarily hidden behind the prior session. The frontend now queues the forced refresh and immediately renders the import response. The repaired production flow was rerun through source upload, migration, acknowledgement refusal, controlled build, artifact/proof download, brain search, fresh-UI persistence, and rollback with zero page, console, or request errors.

## Porting Lab: from diagnosis to controlled migration

Minecraft Mod Vault 0.9.0 adds a real **Porting Lab** for upgrading, downgrading, reconstructing, and cross-loader porting Minecraft mods without mutating the installed original.

The new workflow is evidence-first:

1. select exact source and target Minecraft versions and loaders;
2. optionally attach an installed managed JAR;
3. build a version-pinned migration plan;
4. probe the local build toolchain;
5. generate an isolated workspace with copied inputs and immutable hashes;
6. execute source reconstruction, mapping, API/data migration, build, and runtime verification as separate gated phases.

A remapped or metadata-edited JAR is never reported as a completed port. Completion requires reproducible source or source-equivalent evidence, exact dependency and loader contracts, client and dedicated-server proof where supported, fresh log inspection, persistence/restart testing, and a rollback-preserved artifact.

## 907-version official Version Atlas

The release embeds a generated atlas spanning 907 Mojang release, snapshot, beta, and alpha manifests. Each version record can expose:

- required Java major version;
- client and dedicated-server artifact availability;
- client/server mappings availability;
- protocol version;
- world/data version;
- data-pack and resource-pack format;
- release time and version type;
- Fabric Loader, Intermediary, Yarn, Quilt Loader, Quilt Mappings, and Modrinth coverage where available.

The atlas also carries reviewed build-tool metadata for Forge, NeoForge, Fabric Loom, Architectury Loom, ForgeGradle, ModDevGradle, Tiny Remapper/Auto Renaming Tool, and related planning inputs. Every embedded upstream snapshot is listed with byte size and SHA-256 provenance.

## Live managed-JAR forensics

When a Porting Lab plan references a managed JAR, the backend reads the actual artifact and records:

- SHA-256, SHA-512, size, and CurseForge-compatible fingerprint;
- loader metadata, mod IDs, versions, dependencies, and source URLs;
- class count, maximum Java class-file level, and mapping-namespace clues;
- Mixins, refmaps, plugins, access wideners, access transformers, coremods, and transformation services;
- nested JARs, native libraries, signatures, data and asset counts;
- reflection, `Unsafe`, `MethodHandles`, MixinExtras, Kotlin, and Scala signals;
- client/server references and packaged metadata;
- explicit risk signals and truncation disclosure.

The analysis feeds the plan’s risk, warnings, boundaries, verification matrix, and completion gates. It is not a promise that static inspection alone can prove compatibility.

## Isolated, hash-locked workspaces

Workspace generation is deliberately non-destructive:

- inputs must be regular managed files; symbolic links and paths outside managed Minecraft directories are rejected;
- the input path cannot be swapped after the plan is created;
- the planned SHA-256 is rechecked before copying;
- the original is copied into the workspace and never moved;
- plan JSON/Markdown, exact coordinates, tool recommendations, generated build scaffolding, evidence directories, scripts, and every workspace-file hash are retained in a manifest;
- the workspace has an explicit rollback contract: delete the disposable workspace, not the original.

## First executable Doctor repair: cryptographic duplicate quarantine

0.9.0 introduces the first Doctor action allowed to mutate managed files. Its scope is intentionally narrow: **exact byte-identical duplicate JARs only**.

The transaction:

- groups only files with the same full SHA-512 digest;
- chooses a deterministic keeper, preferring enabled and stable filenames;
- skips non-regular files and symbolic links;
- re-stats and re-hashes every candidate immediately before moving it;
- moves duplicates into Vault quarantine rather than deleting them;
- re-hashes each quarantined destination and the surviving keeper;
- writes an atomic machine-readable receipt and human restore guide;
- rolls back the whole transaction if validation, movement, hashing, or receipt creation fails;
- restores through a second validated transaction and verifies every restored hash.

Changed bytes, missing files, destination conflicts, paths outside managed directories, forged receipt locations, and ambiguous artifacts abort safely.

## Current research brain expansion

The embedded Doctor/Porting catalog now contains:

- **158** source records;
- **96** detailed execution records;
- **167** unique runtime tool cards;
- **9** reusable repair patterns.

Newly integrated research includes InterMed, modcrawl, Modpack Inspector, ModLens MCP, and ModpackResolver. InterMed is treated as an alpha read-only evidence generator; modcrawl as a young independent JAR-analysis cross-check; Modpack Inspector as a human-facing pack forensics comparison; ModLens MCP as a local evidence index; and ModpackResolver as coarse version-availability triage rather than proof of runtime compatibility.

Incorrect provisional repository links discovered during implementation were corrected before release. Retromod points to Bownlux/Retromod, InterMed to jarettr/intermed, and modcrawl to SirCesarium/modcrawl.

## Toolchain selection and safety posture

Porting plans combine official loader-native toolchains with independent analyzers:

- Fabric Loom, Architectury Loom, ForgeGradle, and ModDevGradle for native target builds;
- RetroFuturaGradle, RetroMCP-era tooling, Unimined, Cleanroom/Fugue/MixinBooter, Ornithe/Feather/Ploceus, Legacy Fabric, StationAPI, and LegacyFix for historical eras;
- Vineflower and CFR for independent binary-source reconstruction;
- Tiny Remapper/Mapping-IO class tooling for namespace work;
- japicmp for binary API drift;
- Modstitch, Stonecutter, and Architectury Loom for deliberate multi-loader/multi-version source layouts;
- Retromod and runtime bridges only as controlled candidates, never as automatic proof;
- packwiz and Ferium for reproducible disposable integration packs and acquisition cross-checks.

## Preserved product capability

The federated 28-lane Mods browser, provider-specific installation, Universal Updater, Creator Archive, Creator Picks, transcription, recommendations, CIT, Furniture, Bedrock, Manager, and all prior Mod Doctor analysis remain intact. Porting Lab is a distinct real workspace and Manager JAR actions deep-link the exact selected file into it.

## Verification contract

The release gate covers Go formatting/tests/vet, JavaScript syntax, JSON/JSONL integrity, atlas invariants, plan and workspace hash-lock tests, repair rollback/restore tests, authenticated loopback API smoke, real workspace generation, exact-duplicate quarantine and restore, UI control wiring, platform builds, deterministic packaging, archive extraction, source rebuild, binary comparison, and final extracted-package runtime verification.

Exact commands, observed responses, hashes, and package checks are recorded in `Minecraft-Mod-Vault-0.9.0-BUILD-VERIFICATION.txt` and the `verification/` directory shipped with every release package.
