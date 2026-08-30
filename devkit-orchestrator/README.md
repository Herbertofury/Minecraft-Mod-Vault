# Minecraft Dev Kit Orchestrator 2.7.0 — Compatibility Atlas, Live Source, Tool, Cache & Native-Client Manager

`mmv-devkit` keeps the Minecraft development laboratory current and makes mod conversions defendable. It tracks canonical upstream identities for tools, mods, APIs and libraries; resolves target-compatible releases; independently resolves the newest upstream feature authority; follows required dependencies; mirrors source; preserves stable Google Drive file IDs; and blocks conversion completion when original or newest-upstream content was lost or corrupted.

2.5 adds the **Compatibility Atlas**: a plugin-driven, version-aware memory for optional mod APIs. It learns exact class/method descriptors from runtime JARs, keeps source-only evidence separate, compares API lanes across Minecraft/mod/loader versions, generates soft-linked adapter scaffolds, audits accidental hard links, and builds the complete none/any-subset/all-installed matrix for core modpack integrations. The 2.3 portable-build/native-rendering primitives and 2.4 exact-provider fetch remain intact.

## Compatibility Atlas (2.6.3)

Initialize the built-in plugin pack, ingest every runtime/source lane you possess, then plan or verify a target mod:

```text
mmv-devkit compat init --root DEVKIT
mmv-devkit compat ingest --root DEVKIT --plugin tacz --artifact tacz-1.20.1-1.1.8-release.jar
mmv-devkit compat ingest --root DEVKIT --plugin bettercombat --artifact BetterCombat-26.2.zip
mmv-devkit compat diff --root DEVKIT --plugin epicfight --from 20.14.17 --to 21.17.3.1 --json
mmv-devkit compat matrix --root DEVKIT --mods tacz,bettercombat,punchy,epicfight,ragdoll,create,valkyrienskies,clockwork,eureka,ysm
mmv-devkit compat plan --root DEVKIT --target PROJECT --mods epicfight,bettercombat,punchy,ysm --mc 1.20.1 --loader forge
mmv-devkit compat scaffold --root DEVKIT --target PROJECT --mods tacz,bettercombat --package com.example.compat
mmv-devkit compat verify --root DEVKIT --target PROJECT --mods epicfight,bettercombat,punchy,ysm --mc 1.20.1 --loader forge
```

The durable state lives under `.mmv/compat/`:

- `atlas.json` stores learned version lanes, artifact SHA-256, evidence strength, exact runtime class/method descriptors, semantic-anchor resolutions, and API fingerprints.
- `plugins/*.json` defines one independent compatibility plugin per provider mod. Built-ins currently cover **TaCZ, Better Combat, Punchy, Epic Fight, Ragdoll/Reactions, Create, Valkyrien Skies, Clockwork, Eureka, and Yes Steve Model**. Plugin-level `supportedLoaders` describes the known loader family while each learned lane remains exact to its artifact. **Better Combat is explicitly Fabric + Forge + NeoForge capable; the 1.20.1 1.9.0 lane supports Forge and NeoForge.**
- `compat init` never overwrites a customized plugin. `--refresh` writes a changed built-in beside it as `*.upstream.json` for deliberate merging.
- Runtime JAR lanes are `RUNTIME_BYTECODE`; source ZIP/JAR lanes are `SOURCE_EVIDENCE`. Source evidence can guide a port but cannot be promoted to production descriptor proof.
- `compat diff` reports complete added/removed symbol sets, semantic-anchor moves/signature drift, and migration guidance between learned lanes.
- `compat matrix` includes the empty case, every valid subset, and the all-installed case. Dependency-invalid combinations are retained and explained instead of silently skipped.
- `compat plan/verify` rejects direct optional-provider package links in Java/Kotlin/Scala sources while ignoring comments/string literals used by safe reflection/Mixins.
- `compat scaffold` generates `Class.forName`-based adapters and never imports an optional provider class.

This is intentionally a **memory atlas**, not a single-version patcher: learning a new mod/Minecraft/loader lane adds evidence without erasing old mappings. API drift becomes a comparison result rather than a reason to rewrite history.

## Portable cache + native client QA

Large Gradle/Forge caches and Minecraft client assets are now treated as verifiable artifacts rather than opaque folders.

```text
mmv-devkit archive-split --file BIG.zip --out-dir BIG.zip-SPLIT --part-mib 85
mmv-devkit cache-reassemble --manifest BIG.zip-SPLIT/BIG.zip-SPLIT-MANIFEST.json --extract GRADLE_HOME
mmv-devkit cache-doctor --cache GRADLE_HOME --mc 1.20.1 --forge 47.4.23 --forgegradle 6.0.54 --expect-min-files 1600
mmv-devkit client-assets --gradle-home GRADLE_HOME --mc 1.20.1 --workers 16
mmv-devkit client-natives --gradle-home GRADLE_HOME --platform linux --out PROJECT/build/mmv-natives-linux
```

`archive-split` writes an ordered split manifest with SHA-256 and exact byte size for the original plus every part. `cache-reassemble` verifies all of it before use, rejects traversal/symlink/duplicate-normalized entries, normalizes Windows backslash ZIP paths, and resumes extraction by CRC32 so a terminated extraction cannot masquerade as a complete cache. `cache-doctor` verifies the actual cache surfaces needed by a target instead of trusting file count or folder existence alone.

`client-assets` reads the exact Mojang asset index from cached Minecraft version metadata and downloads content-addressed objects from Mojang's official CDN. Every index and asset is size/hash verified, correct existing files are reused, and `--verify-only` is network-free. No Microsoft/Minecraft credentials are required.

`client-natives` extracts the actual native libraries from cached LWJGL classifier JARs for `linux`, `windows`, or `macos`. This is important for portable ForgeGradle caches that were generated on another OS and therefore contain a cached client POM selecting the wrong platform classifiers.

## The two-version rule

For ordinary runtime/update work, the newest **compatible** release is still the correct artifact to install for the requested Minecraft version/loader/platform.

For conversion/port work, compatibility is not the content ceiling. `heritage` and `port-guard` always maintain two independent truths:

1. **Target compatibility reference** — newest release that matches the requested Minecraft + loader when one exists.
2. **Feature authority** — newest upstream release across every canonical provider attached to the artifact, with the Minecraft/loader filter intentionally removed.

That means a Forge 1.20.1 conversion can use the 1.20.1 build for old API/format guidance while still harvesting blocks, items, mobs, recipes, assets, fixes, dependencies and source work that only landed in a later 26.x release.

If upstream never shipped the requested target at all, heritage does not stop. The compatibility reference is labelled `minecraft-only`, `loader-only`, or `latest-fallback`; the exact original artifact remains the preservation baseline.

## Port heritage

```text
mmv-devkit heritage \
  --registry devkit-registry.json \
  --id some-mod \
  --mc 1.20.1 --loader forge \
  --out .mmv/heritage/some-mod \
  --report heritage-report.json
```

The command downloads both the target-compatible reference and the newest upstream runtime. When upstream source is discoverable, it also resolves and downloads matching source snapshots/tags. The report records:

- exact provider/project/version/file identities;
- release lineage and changelogs where the provider exposes them;
- newest-vs-compatible runtime archive delta;
- newest-vs-compatible source delta;
- added/changed/removed assets and data grouped by category;
- new required dependency relationships;
- hashes, file counts and category counts;
- the compatibility-reference strength (`exact`, `minecraft-only`, `loader-only`, `latest-fallback`).

Provider response ordering is never trusted as freshness proof: Modrinth versions, CurseForge files and GitHub releases are sorted by publication time before selection.

## Mandatory conversion completion guard

A conversion is not done because it compiles. Before finishing, run:

```text
mmv-devkit port-guard \
  --registry devkit-registry.json \
  --id some-mod \
  --original path/to/the-exact-original.jar \
  --converted path/to/build/libs/converted.jar \
  --converted-source path/to/project/src \
  --mc 1.20.1 --loader forge \
  --out .mmv/heritage/some-mod \
  --report port-guard-report.json
```

`port-guard` re-downloads/re-resolves current upstream heritage at the finish gate instead of trusting a stale planning report. It then checks four states together:

1. **original artifact** actually being converted/worked on;
2. **target-compatible upstream reference** when available;
3. **newest upstream feature authority** regardless of target version;
4. **converted target build + converted source tree**.

It exits non-zero when a hard fidelity requirement fails. This is designed to be consumed directly by Minecraft Mod Vault TestGrid/CI.

### Hard failures

- content present in the original and still present upstream disappeared from the conversion;
- original content that upstream did not change has different converted bytes (probable corruption/unwanted mutation);
- upstream changed content but the conversion silently kept the stale original bytes;
- a newly added latest-upstream asset/data file is missing;
- portable newest content such as textures/language/sounds is byte-corrupted or stale;
- converted JAR/ZIP is malformed or has duplicate archive entries;
- newest source adds a top-level type that is absent from the converted source surface;
- newest source has additions but the converted source tree was omitted from the completion audit.

Version-sensitive JSON/data/model differences may legitimately require adaptation when backporting; these are surfaced as explicit review warnings rather than hidden. Portable binary/text content that should survive unchanged is strict.

The output is a machine-readable evidence ledger, not a vague similarity score.

## Live tool/dependency management

2.0 made dependencies and source mirrors live. 2.1 closed the tool-side loop: `adopt-tools` walks the configured Dev Kit Drive tree, matches existing files against the preserved 59-tool provenance catalog, binds real Drive file IDs into managed state, infers obvious OS/architecture/Minecraft/loader constraints, and attaches only safe canonical providers.

```text
mmv-devkit validate --registry devkit-registry.json
mmv-devkit sources --catalog tool-provenance.json
mmv-devkit adopt-tools --catalog tool-provenance.json --registry devkit-registry.json --drive
mmv-devkit check --registry devkit-registry.json --json
mmv-devkit sync --registry devkit-registry.json --drive
mmv-devkit sync --registry devkit-registry.json --apply --drive
mmv-devkit watch --registry devkit-registry.json --drive --interval 15m
```

`adopt-tools` is idempotent for already tracked Drive file IDs. It never creates or replaces files; it binds safely matched existing files into managed state. Applying updates uses the same verified in-place Drive path as dependencies.

## Provider truth

- **Modrinth**: project/version API, exact Minecraft + loader filters for compatible artifacts, unfiltered newest release for feature authority, provider hashes, required dependency graph, changelogs and project source repository.
- **CurseForge**: project/file API, game-version + loader filters where applicable, pagination, required dependency relations and SHA-1 verification when `CURSEFORGE_API_KEY` is available.
- **GitHub releases**: release assets for rolling binaries/mod releases, publication-time ordering, platform/architecture scoring and curated asset-family filters.
- **GitHub branches**: exact commit SHA + source archive for templates/source snapshots whose freshness is branch-head based.
- **Direct/vendor URL**: explicit fallback only; never silently outranks an exact provider identity.

The heritage feature authority is selected across all configured canonical hubs, not merely the first provider that happens to answer.

## Dependencies

Required dependencies are recursive with cycle protection and a 12-level ceiling. An unresolved required dependency blocks the parent update transaction. Unknown dependencies discovered from provider metadata can be enrolled automatically into `04 Reference Runtime Mods/Dependency & API Libraries`; matching source is mirrored into `02 Build & Port Frameworks/Dependency Source Archives` when available.

Heritage separately reports **new required dependencies introduced after the target-compatible release**, because those often represent functionality that must also be backported or replaced during a conversion.

## Source matching

For Modrinth projects with GitHub source, the manager tries to match the selected runtime version to a GitHub release/tag. Branch-managed tools resolve branch heads to exact commit SHAs before download. `main`, `master` and `develop` are therefore provenance inputs, not unversioned final states.

Port heritage compares the compatible source snapshot to the newest source snapshot and can compare newly introduced source type surfaces to the converted project source.

## Drive in-place updates

The registry stores Google Drive file IDs. With `MMV_GOOGLE_DRIVE_TOKEN`, tracked artifacts are replaced through Drive `files.update` media semantics so the canonical Drive identity/link survives the update. Upstream filename changes can rename the same object. New Drive objects are created only for genuinely new managed artifacts/dependencies/sources with no existing tracked ID.

## Environment credentials

```text
CURSEFORGE_API_KEY
GITHUB_TOKEN
MMV_GOOGLE_DRIVE_TOKEN
```

Secrets are environment-only and are never written to registries, lockfiles, provenance catalogs or audit reports.

## Safety gates

- canonical provider/project identity instead of fuzzy search;
- compatibility artifact and newest feature authority are separate concepts;
- latest selection is publication-time/version ordered rather than response-order dependent;
- Minecraft/loader/platform/release-channel filtering for installed artifacts;
- downgrade protection by default;
- Modrinth SHA-256 verification;
- CurseForge SHA-1 verification when supplied;
- required-dependency failures block update transactions;
- recursive dependency cycle/depth protection;
- stable Drive file IDs and atomic local state writes;
- exact provider/project/version/source provenance;
- GitHub branch updates pinned to commit SHA;
- tool adoption is non-destructive and skips ambiguous matches;
- conversion completion re-checks the original + current upstream rather than trusting stale planning state;
- malformed/duplicate archive content fails verification;
- credentials excluded from durable state.

## Relationship with Minecraft Mod Vault / TestGrid

`mmv-devkit` is the source/tool/library/heritage control plane. Minecraft Mod Vault TestGrid remains the execution/evidence engine for real builds, launches, assertions, logs, networking and runtime proof.

For ports the intended completion chain is:

```text
heritage -> implement/backport -> build -> port-guard -> TestGrid runtime proof -> release
```

A build that passes TestGrid but fails `port-guard` is incomplete because it lost source content or newest-upstream features. A build that passes `port-guard` but fails TestGrid is incomplete because structural fidelity alone does not prove runtime correctness. Both gates are required.

## Exact mod artifact fetch (2.7.0)

`fetch-mod` retrieves an exact provider release to the local filesystem without requiring Google Drive. It is intended for porting, forensic parity work, reproducible QA, and dependency recovery where "latest compatible" is not precise enough.

```bash
mmv-devkit fetch-mod --provider modrinth --project 9qn2AQBc --version VCCCalGp --out refs/AoA3-1.16.5-3.6.11.jar --receipt refs/aoa-fetch.json
mmv-devkit fetch-mod --provider curseforge --project 311054 --file 4431558 --out refs/AoA3-1.16.5-3.6.11.jar --receipt refs/aoa-fetch.json
```

Safety properties: canonical project/version provenance checking, provider SHA-512/SHA-256/SHA-1 verification when available, provider size verification, streaming to disk instead of whole-artifact buffering, interrupted-transfer resume with HTTP Range, bounded retries/backoff, redirect-safe provider headers, deterministic fallback URLs, atomic verified promotion, and cleanup/no replacement when verification fails. Modrinth accepts either a canonical project ID or slug and verifies the exact version belongs to the resolved canonical project. If the primary Modrinth file URL fails, the resolver can retry the canonical CDN route and the Modrinth Maven artifact for JAR primaries.

CurseForge direct CDN downloads require `CURSEFORGE_API_KEY` (aliases `CF_API_KEY` and `CURSEFORGE_KEY` are also accepted). The key is attached as `x-api-key` to both metadata/download-url requests and the CDN transfer, and is preserved across redirects. If a CurseForge file record omits `downloadUrl`, the Dev Kit requests the dedicated download-url endpoint and can deterministically reconstruct the official `edge.forgecdn.net/files/{id}/{subId}/{filename}` route from exact file metadata.

## Vanilla Feature Atlas + Backport Ownership 2.6.4

The `vanilla-atlas` command family prevents cross-version vanilla dependencies from being stubbed or silently discarded. It models vanilla lineage as versioned feature families and produces fail-closed ownership decisions. Future vanilla may be owned only by target vanilla, one proven external provider, or the canonical `futurevanillabackport` provider; individual ports do not privately embed duplicate vanilla implementations.

**2.6.3 makes the ordering explicit:** finish and certify the full target-native mod port first. `plan-backport` is an always-generated post-port **offer**, default OFF. Planning can happen earlier, but implementation authorization requires explicit `--opt-in` plus a passing base `port-guard` report via `--port-report`. The certified base artifact remains intact. Every offer exposes the exact dependency closure and labels future vanilla as an optional enhancement or a requirement for exact source-version parity where applicable.

Sound identity is explicitly three-layered: registered `SoundEvent`, `sounds.json` definition, and external OGG object(s). A registered event with no definition is preserved as a real historical state rather than misreported as a missing asset.

```text
mmv-devkit vanilla-atlas verify --atlas VANILLA-ATLAS.json
mmv-devkit vanilla-atlas query --atlas VANILLA-ATLAS.json --id minecraft:item.goat_horn.play --target 1.20.1
mmv-devkit vanilla-atlas sound-status --atlas VANILLA-ATLAS.json --id minecraft:item.goat_horn.play --target 1.20.1
mmv-devkit vanilla-atlas diff --atlas VANILLA-ATLAS.json --from 1.20.1 --to 26.2 --kind sound_event
mmv-devkit vanilla-atlas providers --atlas VANILLA-ATLAS.json --mods-dir mods
mmv-devkit vanilla-atlas plan-backport --atlas VANILLA-ATLAS.json --target 1.20.1 --source 26.2 --feature <id> --mods-dir mods
# After base port-guard passes and the user opts in:
mmv-devkit vanilla-atlas plan-backport --atlas VANILLA-ATLAS.json --target 1.20.1 --source 26.2 --feature <id> --port-report port-guard-report.json --opt-in --json
```

The reproducible builder in `tools/build_vanilla_atlas.py` uses Mojang's version metadata/asset indexes as official authority, misode/mcmeta for generated registry/data/resource history, PrismarineJS/minecraft-data for normalized older observations, and Vanilla Backport only as compatibility-provider evidence. Heavy vanilla bytes are hydrated and hash-verified on demand instead of duplicated for every version.
