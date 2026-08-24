# Minecraft Mod Vault 0.10.0 — OmniManager implementation checkpoint 00

Recorded: 2026-08-21 UTC
Baseline: verified 0.9.0 source archive, local baseline commit `ead1c8c`
Target: **0.10.0 OmniManager**

## User-visible failure being corrected

Existing Manager renders little more than filenames. That repeats the core failure visible in major launchers: a file hosted on CurseForge but absent from Modrinth can appear as an anonymous uploaded cube even though another provider and the JAR itself know its real title, author, artwork, version, loader, game versions, and latest compatible release.

## Acceptance contract

1. Replace the bare folder listing with a premium unified Java + Bedrock library.
2. Resolve identity from immutable file hashes first, then embedded metadata, provider project URLs, provider searches, normalized name/author/slug evidence, and user-confirmed aliases.
3. Merge provider records instead of treating Modrinth, CurseForge, GitHub, MCPEDL, Marketplace, and other supported sources as mutually exclusive identities.
4. Display the best truthful name, icon/artwork, author, installed version, latest compatible version, loader/game compatibility, source badges, update state, provenance confidence, and original filename.
5. Never replace patched or locally modified builds merely because a public version shares a human version string.
6. Detect updates by file identity and target compatibility, not only the displayed mod version.
7. Support bulk selection, filter/search/sort, enable/disable, update, quarantine/trash, inspect/repair/port, open folder/page, rescan, and exact-item detail views.
8. Add complete Bedrock package inspection for `.mcpack`, `.mcaddon`, `.mcworld`, `.mctemplate`, and installed behavior/resource/skin/world/template packs.
9. Parse Bedrock `manifest.json`, UUID/version, modules, dependencies, minimum engine version, scripts/capabilities, and `pack_icon.png`; model linked behavior/resource packs as one add-on family while preserving each pack.
10. Support Minecraft for Windows stable, Preview/Beta, custom `com.mojang`, and configured Bedrock roots without silently writing outside managed paths.
11. Preserve originals; updates and removals remain transactional and recoverable.
12. No fake controls, no decorative-only panels, no image generation, and no unsupported provider capability presented as working.

## Evidence-backed implementation decisions

- Modrinth files are identified by SHA-1/SHA-512 and its API exposes exact-hash and compatible-update endpoints, so cryptographic identity is the first remote resolver.
- CurseForge exposes exact MurmurHash2 fingerprint matching and latest-file indexes when an API key is configured.
- Embedded Fabric/Quilt/Forge/NeoForge metadata remains usable even when no provider recognizes the file.
- Bedrock manifests are authoritative for pack name, UUID, version, modules, dependencies, and minimum engine version; pack icons are read from the package itself.
- Provider outages or missing credentials degrade to local metadata and cached evidence rather than erasing names/art.

## Planned implementation slices

1. `omnimanager_types.go`: unified identity, source, artwork, update, dependency, Bedrock, and action models.
2. `omnimanager_scan.go`: safe Java/ZIP/Bedrock inspection, artwork extraction/cache, folder/profile discovery, duplicate/family graphing.
3. `omnimanager_resolve.go`: Modrinth hash, CurseForge fingerprint, declared URL, GitHub/repository, cross-provider search, scoring, merge, cache.
4. `omnimanager_actions.go`: bulk update/toggle/trash, details, rescan, Bedrock install/import/world activation with receipts.
5. `/api/library/*`: summary, scan, item, enrich, actions, update, Bedrock targets/world assignments.
6. Premium Manager UI: rich rows/cards, artwork, source badges, update/provenance state, facets, bulk command bar, details drawer, responsive layout.
7. Regression and hostile-package tests, exact Cataclysm Dimensions cross-source fixture, Bedrock linked-pack fixture, fresh extracted runtime proof.

## Current primary sources reviewed

- Modrinth API overview and version-file hash/update operations: https://docs.modrinth.com/api/
- CurseForge Core API fingerprints/files/mods endpoints: https://docs.curseforge.com/rest-api/
- Minecraft Bedrock add-on manifest reference: https://learn.microsoft.com/en-us/minecraft/creator/reference/content/addonsreference/packmanifest
- Current manager designs and known failure reports: Prism Launcher, Ferium/libium, packwiz, Pakku, Modpack Inspector, and mod-downloader repositories/issues.
