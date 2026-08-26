# Dev Kit Orchestrator

`mmv-devkit` 2.2 is the live source-of-truth and conversion-heritage manager for the Minecraft Dev Kit. It complements Minecraft Mod Vault TestGrid: the orchestrator keeps tools, APIs, dependency libraries, source repositories, Drive mirrors, and conversion baselines current; TestGrid remains the execution/evidence engine for real builds, launches, logs, hashes, probes, and runtime proof.

## Two different meanings of “latest”

For normal installation/update work, `mmv-devkit` resolves the newest **compatible** artifact for the requested Minecraft version, loader, release channel, OS, and architecture.

For a mod conversion, compatibility is **not** the feature ceiling. The orchestrator also resolves the newest upstream release independently of the target Minecraft version/loader and treats that as the **feature authority**. A Forge 1.20.1 port can therefore use an old 1.20.1 release for API/format guidance while still discovering and carrying forward blocks, items, mobs, recipes, assets, fixes, dependencies, config/data changes, and source work added only in later 26.x releases.

If upstream never shipped the requested target, the compatibility reference is explicitly labelled `minecraft-only`, `loader-only`, or `latest-fallback`; the exact original artifact remains the preservation baseline.

## Mandatory conversion heritage workflow

```text
mmv-devkit heritage \
  --registry devkit-registry.json \
  --id MOD_ID \
  --mc 1.20.1 --loader forge \
  --out .mmv/heritage/MOD_ID \
  --report heritage-report.json
```

`heritage` resolves/downloads the target-compatible runtime and newest upstream runtime, plus matching source snapshots when discoverable. It records release lineage/changelogs where available, hashes, added/changed/removed archive content grouped by type, source deltas, and newly introduced required dependencies.

Before a conversion can be considered complete, run:

```text
mmv-devkit port-guard \
  --registry devkit-registry.json \
  --id MOD_ID \
  --original exact-original.jar \
  --converted build/libs/converted.jar \
  --converted-source src \
  --mc 1.20.1 --loader forge \
  --report port-guard-report.json
```

`port-guard` deliberately **re-resolves current upstream at finish time** instead of trusting the earlier planning snapshot. It compares four states together: the exact original artifact, target-compatible upstream reference, newest upstream feature authority, and converted build/source.

It fails closed when original content that should survive was lost, unchanged original content was unexpectedly modified/corrupted, the conversion silently retained stale bytes after upstream changed them, newest-upstream assets/data are missing, portable newest content such as textures/language/sounds is corrupted, archive entries are unsafe/duplicated, or newly introduced upstream source surface is absent from the converted source.

Version-sensitive code/data that legitimately needs backport adaptation is surfaced for explicit review rather than incorrectly demanding byte identity. Runtime behavior is then verified by TestGrid. The release chain is:

```text
heritage -> implement/backport -> build -> port-guard -> TestGrid runtime proof -> release
```

Passing only one of `port-guard` or TestGrid is insufficient.

## Live update management

1. Resolve the newest compatible artifact rather than blindly taking the newest upload.
2. Track canonical provider/project/repository identities instead of stale filenames or fuzzy search.
3. Resolve required dependencies recursively with cycle/depth protection.
4. Mirror matching upstream source alongside runtime artifacts.
5. Verify provider hashes before accepting downloaded bytes.
6. Block downgrades by default.
7. Preserve stable Google Drive file IDs by replacing tracked objects in place.
8. Auto-enroll previously unknown required dependencies during Drive-backed synchronization.
9. Record provider, project/version IDs, source refs, hashes, and Drive IDs in durable state.
10. Adopt already-present Dev Kit tools non-destructively before allowing them to update.

## Providers

- **Modrinth** — exact target version/loader resolution, unfiltered newest-upstream feature authority, provider hashes, required dependencies, changelogs, and source repository metadata.
- **CurseForge** — game-version/loader-aware file resolution, release lineage, required dependencies, and SHA-1 verification when available. Requires `CURSEFORGE_API_KEY` for authenticated API resolution.
- **GitHub** — release assets, tags/source archives, and branch snapshots pinned to exact commit SHAs.
- **Direct/vendor URLs** — explicit fallback only; generic URLs never silently outrank strong provider identity.

## Commands

```text
mmv-devkit version
mmv-devkit validate --registry devkit-registry.json
mmv-devkit check --registry devkit-registry.json --json
mmv-devkit sync --registry devkit-registry.json --drive
mmv-devkit sync --registry devkit-registry.json --apply --drive
mmv-devkit watch --registry devkit-registry.json --drive --interval 15m
mmv-devkit sources --catalog tool-provenance.json [--id TOOL_ID] [--json]
mmv-devkit adopt-tools --catalog tool-provenance.json --registry devkit-registry.json --drive
mmv-devkit heritage --registry devkit-registry.json --id MOD_ID --mc VERSION --loader LOADER
mmv-devkit port-guard --registry devkit-registry.json --id MOD_ID --original ORIGINAL.jar --converted CONVERTED.jar --converted-source SRC
```

## Java policy

| Minecraft target | Resolver Java |
|---|---:|
| 1.16.x and older legacy targets | 8 where required |
| 1.17 through 1.20.x | 17 |
| 1.21.x | 21 |
| 26.x | 25 |

The offline Dev Kit includes JDK 17, **JDK 21**, JDK 25, and JDK 26 archives.

## Dependency/source behavior

Required dependencies are recursive. An unresolved required dependency blocks its parent update instead of producing a half-valid environment. During conversion heritage, required dependencies introduced after the target-compatible release are separately surfaced because they may represent functionality that also needs backporting or replacement.

When a source repository is exposed, the manager attempts to match selected runtime versions to releases/tags. Branch-managed source snapshots are pinned to exact commit SHAs. `main`, `master`, and `develop` are provenance inputs, not timeless version identifiers.

## Google Drive in-place updates

The registry records Drive file IDs. With `MMV_GOOGLE_DRIVE_TOKEN`, tracked artifacts use Drive `files.update` semantics so the canonical file ID/link survives a version refresh. New objects are created only for genuinely new managed artifacts/dependencies/sources.

Credentials remain environment-only:

```text
CURSEFORGE_API_KEY
GITHUB_TOKEN
MMV_GOOGLE_DRIVE_TOKEN
```

## Provenance catalog

The preserved Dev Kit inventory contains 59 stable tool identities. Eligible existing files can be adopted into live managed state using their real Drive file IDs; ambiguous/vendor-only tools remain explicit/manual until a trustworthy machine-readable update route exists.

## Safety gates

- target-compatible and newest-feature-authority releases are separate concepts
- exact provider/project identities
- release/channel/version/loader/platform selection
- downgrade protection
- provider hash verification
- dependency cycle/depth protection
- stable Drive identities and atomic local state writes
- source tag/commit provenance
- safe archive-path and duplicate-entry rejection
- exact original re-check before conversion completion
- newest upstream re-check before conversion completion
- missing/stale/corrupted content fails closed
- credentials excluded from durable state

## TestGrid handoff

The orchestrator determines **what must be present and current** and protects source fidelity. Minecraft Mod Vault/TestGrid proves **whether the converted result actually works** through real builds/runs and evidence capture. Together they prevent both feature loss and superficially successful but broken ports.
