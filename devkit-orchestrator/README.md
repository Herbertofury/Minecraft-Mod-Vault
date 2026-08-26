# Minecraft Dev Kit Orchestrator 2.2.0 — Live Source, Tool & Port-Heritage Manager

`mmv-devkit` keeps the Minecraft development laboratory current and makes mod conversions defendable. It tracks canonical upstream identities for tools, mods, APIs and libraries; resolves target-compatible releases; independently resolves the newest upstream feature authority; follows required dependencies; mirrors source; preserves stable Google Drive file IDs; and blocks conversion completion when original or newest-upstream content was lost or corrupted.

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
