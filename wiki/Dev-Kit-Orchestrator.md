# Dev Kit Orchestrator

`mmv-devkit` 2.0 is the live source-of-truth manager for the Minecraft Dev Kit. It complements Minecraft Mod Vault TestGrid: the orchestrator keeps tools, APIs, dependency libraries, source repositories, and their Drive mirrors fresh; TestGrid remains the execution/evidence engine for real builds, launches, logs, hashes, probes, and runtime proof.

## What 2.0 manages

1. Resolve the newest **compatible** artifact rather than blindly taking the newest upload.
2. Match Minecraft version, loader, release channel, OS, and architecture where those dimensions apply.
3. Track canonical provider/project/repository identities instead of relying on stale filenames or fuzzy search.
4. Resolve required dependencies recursively with cycle protection and a bounded recursion depth.
5. Mirror matching upstream source alongside runtime artifacts when a source repository is published.
6. Verify provider hashes before accepting downloaded bytes.
7. Block downgrades by default.
8. Preserve stable Google Drive file IDs by replacing existing tracked objects in place.
9. Auto-enroll previously unknown required dependencies into the managed dependency lane during Drive-backed synchronization.
10. Record exact provider, project/version IDs, source refs, hashes, and Drive IDs in durable registry/lock state.

## Providers

- **Modrinth** — project/version API, exact Minecraft + loader filtering, release-channel selection, provider hashes, required dependency graph, and project source repository metadata.
- **CurseForge** — project/file API, release metadata, required dependency relations, and SHA-1 verification when available. Requires `CURSEFORGE_API_KEY` for authenticated API resolution.
- **GitHub** — release assets and source release/tag archives. `GITHUB_TOKEN` is optional but recommended for higher API limits.
- **Direct/vendor URLs** — explicit fallback only. A generic URL never silently outranks a strong provider identity.

A newer file for the wrong Minecraft version, loader, OS, architecture, or release channel is **not** treated as an update.

## Commands

```text
mmv-devkit version
mmv-devkit validate --registry devkit-registry.json
mmv-devkit check --registry devkit-registry.json --json
mmv-devkit sync --registry devkit-registry.json --drive
mmv-devkit sync --registry devkit-registry.json --apply --drive
mmv-devkit watch --registry devkit-registry.json --drive --interval 15m
mmv-devkit sources --catalog tool-provenance.json [--id TOOL_ID] [--json]
```

`sync` without `--apply` is a dry-run plan. Applying changes requires `--drive`; the updater will not download verified bytes and then leave them only in ephemeral local state. `watch` repeats the same compatibility/dependency/hash gates rather than bypassing them.

## Java policy

| Minecraft target | Resolver Java |
|---|---:|
| 1.16.x and older legacy targets | 8 where required |
| 1.17 through 1.20.x | 17 |
| 1.21.x | 21 |
| 26.x | 25 |

The offline Dev Kit now includes JDK 17, **JDK 21**, JDK 25, and JDK 26 archives. The former JDK-21 gap is closed.

## Dependency and source behavior

Required dependencies are resolved recursively. An unresolved required dependency blocks its parent update instead of producing a half-valid environment. Dependencies already represented in the registry are reused; new required dependencies can be enrolled automatically into `04 Reference Runtime Mods/Dependency & API Libraries` during a Drive-backed sync.

When Modrinth exposes a GitHub source repository, the source manager attempts to match the selected runtime version to the corresponding GitHub release/tag. `main` is only a fallback when no matching release/tag is available. Runtime and source archives have separate stable Drive IDs, with source mirrors stored under `02 Build & Port Frameworks/Dependency Source Archives`.

If a runtime is already current but its tracked source mirror is missing, the planner emits a source-refresh operation instead of needlessly replacing the runtime.

## Google Drive in-place updates

The registry records Drive file IDs. With `MMV_GOOGLE_DRIVE_TOKEN` available, an existing tracked artifact is updated with Drive `files.update` media semantics, preserving its file ID/link. If upstream changes the filename, the object is renamed on that same ID. A new Drive object is created only when the registry has no tracked object for that artifact.

Credentials are read only from environment variables and are never written into the registry or lockfile:

```text
CURSEFORGE_API_KEY
GITHUB_TOKEN
MMV_GOOGLE_DRIVE_TOKEN
```

## Provenance catalog

The original Dev Kit 1.0 inventory contained 59 stable tool identities. The 2.0 release bundle preserves that inventory in `tool-provenance.json`; 47 currently have a canonical GitHub or Modrinth upstream mapping. Vendor-only entries stay explicit/manual until a trustworthy machine-readable update route exists rather than being updated from a guess.

The live registry is seeded with the current dependency/API batch, including Fabric API, Architectury API, Cloth Config, Balm, Forge Config API Port, YACL, Curios, Fabric Language Kotlin, Kotlin for Forge, Resourceful Lib, Moonlight Lib, TerraBlender, Citadel, Bookshelf, Puzzles Lib, Iceberg, Cardinal Components API, Trinkets, CreativeCore, and Placebo.

## Safety gates

- release-channel filtering
- Minecraft/loader/platform-aware selection
- downgrade block by default
- Modrinth SHA-256 verification
- CurseForge SHA-1 verification when supplied
- required-dependency failure blocks parent update
- recursion cycle/depth protection
- stable Drive file IDs
- atomic local registry/lockfile writes
- exact provenance records
- source release/tag matching
- credentials excluded from durable state

## TestGrid handoff

The orchestrator decides **what** is current and compatible and maintains the development laboratory. Minecraft Mod Vault/TestGrid proves **whether it actually works** by performing real builds/runs and collecting evidence. Keeping those roles separate avoids duplicated automation while preserving a single source of truth for tool and dependency selection.