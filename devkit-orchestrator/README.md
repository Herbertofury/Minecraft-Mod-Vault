# Minecraft Dev Kit Orchestrator 2.0.0 — Live Source Manager

The Dev Kit is no longer a static folder of downloads. The 1.0 inventory of 59 tools is preserved in `tool-provenance.json`; 47 currently have canonical GitHub or Modrinth upstream identities, while vendor-only entries remain explicitly marked for manual/provider-specific refresh rather than guessed. `mmv-devkit` tracks the canonical upstream identity of each tool, mod, API and library, resolves the newest **compatible** release for its Minecraft version/loader/platform lane, follows required dependencies recursively, mirrors matching source, and can replace the existing Google Drive object **in place** so the file ID/link remains stable.

## Provider order and truth

- **Modrinth**: project/version API, exact Minecraft + loader filters, provider hashes, required dependency graph, project source repository.
- **CurseForge**: project/file API and dependency relations when `CURSEFORGE_API_KEY` is available. Provider SHA-1 is verified when supplied.
- **GitHub**: release assets for tools and repositories. `GITHUB_TOKEN` is optional but recommended for higher rate limits.
- **Direct/vendor URL**: explicit fallback only; never silently outranks an exact provider identity.

The registry stores provider IDs, not search guesses. A newer file for the wrong Minecraft version, loader, operating system or architecture is not an update.

## Source matching

For Modrinth projects that declare GitHub source, the updater attempts to mirror the GitHub release/tag matching the selected runtime version. `main` is only a fallback when no matching release/tag can be resolved. Runtime and source Drive IDs are tracked separately.

## Dependency behavior

Required dependencies are resolved recursively with cycle protection and a 12-level safety ceiling. An unresolved required dependency blocks the parent transaction. Dependencies that are already represented by another managed artifact are reused; previously unknown required dependencies are automatically enrolled into the registry and mirrored to the dependency library lane during a Drive-backed sync, including their matching source archive when the provider exposes a source repository.

## Drive in-place updates

The registry records stable Drive file IDs. With `MMV_GOOGLE_DRIVE_TOKEN` present, `sync --apply --drive` uses Drive `files.update` media upload semantics to replace bytes without creating a random duplicate. When the upstream filename changes, metadata is renamed on the same file ID. New dependencies/sources are created only when there is no existing tracked Drive ID.

Secrets are never written to the registry or lockfile.

## Commands

```text
mmv-devkit validate --registry devkit-registry.json
mmv-devkit check --registry devkit-registry.json --json
mmv-devkit sync --registry devkit-registry.json --drive        # dry-run plan
mmv-devkit sync --registry devkit-registry.json --apply --drive
mmv-devkit watch --registry devkit-registry.json --drive --interval 15m
mmv-devkit sources --catalog tool-provenance.json --id geckolib-forge
```

`watch` repeats the same compatibility/hash/dependency transaction. It does not bypass the resolver. If a runtime is already current but its tracked source mirror is missing, the planner emits a `source-refresh` transaction instead of needlessly replacing the runtime.

## Environment credentials

```text
CURSEFORGE_API_KEY
GITHUB_TOKEN
MMV_GOOGLE_DRIVE_TOKEN
```

All are optional for read-only operation except the provider that needs them. Drive mutation requires its token.

## Current seeded dependency lanes

The included registry tracks the exact Drive IDs for the current dependency/API batch: Fabric API, Architectury API, Cloth Config, Balm, Forge Config API Port, YACL, Curios, Fabric Language Kotlin, Kotlin for Forge, Resourceful Lib, Moonlight Lib, TerraBlender, Citadel, Bookshelf, Puzzles Lib, Iceberg, Cardinal Components API, Trinkets, CreativeCore and Placebo.

## Safety gates

- release-channel filtering
- Minecraft/loader-aware provider selection
- downgrade block by default
- Modrinth SHA-256 verification
- CurseForge SHA-1 verification when available
- required dependency failure blocks parent update
- recursion cycle/depth protection
- stable Drive file IDs
- atomic registry and lockfile writes
- exact provenance lock entries with provider/project/version/source refs and hashes
- source release/tag matching
- credentials excluded from durable state

## Relationship with Minecraft Mod Vault

Minecraft Mod Vault already has exact Modrinth hash identity, CurseForge fingerprint resolution, GitHub metadata fallback, transactional mod replacement, Doctor/Compatibility Brain and TestGrid. This orchestrator is the **Dev Kit/source/tool/library control plane**. The two use the same provider-first philosophy: Vault manages active game/project artifacts; the orchestrator keeps the development laboratory and canonical source mirrors fresh.
