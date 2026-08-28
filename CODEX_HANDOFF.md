# Codex Handoff — Minecraft Mod Vault / Minecraft Catalog Companion

Continue from Minecraft Catalog Companion **2.9.5**. Do not restart the project or replace the existing research browser/catalog functionality with a new launcher shell. The complete handoff is stored persistently in Google Drive at `/Google Drive/Minecraft Mod Vault/CODEX HANDOFF - Minecraft Mod Vault.md` (Drive file `external-gdrive:file:1LMU6TUyiM28Op5mjYReUrRPgV9lNnDFi`). A CurseForge UI parity screenshot is stored beside it as `/Google Drive/Minecraft Mod Vault/CurseForge parity reference - modpack library.png` (`external-gdrive:file:1a-X2USk1AQz5OjgxmZ2xn0ncDU3a661u`).

## Canonical source and release state

The full editable 2.9.5 working tree in the originating workspace is `/mnt/data/mcc-work-2.9.5`. The persistent source archive is `/Google Drive/Minecraft Mod Vault/Minecraft Catalog Companion 2.9.5 - Source.zip` (`external-gdrive:file:1XfzDd-l510jQXpID1-v4JoyGbPH1q5y0`). The release/checkpoint is `companion/releases/2.9.5/README.md`; the release commit immediately before this handoff was `b5a879af8d36f1fd158f5f1ff69a8978feb86913`.

Important source files include `main.js`, `shell.js`, `preload.js`, `catalog-preload.js`, `src/catalog-store.js`, `src/catalog-renderer.js`, `src/provider-media.js`, `src/curseforge-fastlane.js`, `src/site-adapters.js`, and `scripts/release-qa.js`. The app is currently Electron/Node >=22, package `minecraft-catalog-companion`, version 2.9.5. Preserve the existing browser tabs, persistent sessions/cookies, research split, source/catalog center, live provider media, site adapters, identity/role safety, no-cap galleries, catalog ingest/sync, notes/favorites, and the existing QA suite.

The 2.9.5 checkpoint is `ready_for_native_windows_bok_acceptance`. All 40 existing release suites pass against source, fresh Source ZIP extraction, and final Windows `resources/app`. The pending native Windows acceptance target remains Bok's Banging Butterflies; do not regress its CurseForge gallery repair. Current cache schema is 14 and the gallery recovery has Chromium full-HTML, Node full-HTML, and Chromium DOM fallback lanes.

## Product goal

Turn the Companion into a complete launcher + mod organizer + research/catalog environment by fully integrating the useful launcher/instance/content-management capabilities of `MegalithOfficial/basalt-launcher` while preserving the Companion, then exceed the **actually installed CurseForge and Modrinth launchers** feature-for-feature. Existing CurseForge and Modrinth libraries must be discovered and mounted **in place**: no forced import/copy before Vault can use them. Vault must launch and manage those exact instances where they already live, watch changes made by the original launcher, and optionally clone/copy an instance into Vault-managed storage when the user explicitly chooses. The UI must support the full launcher experience: data-driven left game selector, My Modpacks/library, Discover, Browse, Servers, Skins/accounts, Create, Import, Create Group, user-defined groups, favorites/tags, grid/list instance views, context menus, downloads/tasks, updates, settings, logs, worlds, resource packs, shaders, data packs, configs, snapshots/backups/repair, export/share, accounts, Java/loaders and launch controls. “Full” means every relevant feature observed in installed CurseForge or Modrinth is either implemented or remains an explicit release blocker in a parity matrix.

## Basalt reference and license

Reference upstream: `https://github.com/MegalithOfficial/basalt-launcher`, observed main commit `00b4f9915def3d201e0a411c3b5453940849c683`. Basalt is Tauri 2 + React 19 + Rust and documents instance management, Forge/Fabric/Quilt/NeoForge, Modrinth/CurseForge discovery, dependency resolution, updates/changelogs, Microsoft accounts, skins/capes, Java detection, background tasks, diagnostics, worlds and imports.

Basalt is **GPL-3.0-only** while the current Companion package is private/`UNLICENSED`. Before copying Basalt source, explicitly choose either (1) a GPL-3.0-compatible combined-project licensing path, or (2) a clean-room implementation that uses Basalt as a behavioral/architectural reference without copying its source. Do not silently mix GPL source into an incompatible closed/unlicensed build.

## Architecture direction

Do not bolt two desktop apps together and do not migrate away from Electron merely because Basalt uses Tauri. Preserve the Companion shell unless a replacement proves 100% parity for its WebContentsView/browser/session behavior. Add an authoritative launcher/mod-organizer core behind a narrow API boundary; a Rust native core/sidecar is preferred for filesystem, database, download, archive, credentials, Java, loader, process, snapshot and recovery operations. Build provider abstractions such as `GameProvider`, `LauncherAdapter`, `ExternalInstance`, `ManagedInstance`, `ContentProvider`, `LaunchPlan`, `Task`, and `Snapshot` rather than spreading CurseForge/Modrinth special cases through the UI.

## Installed CurseForge + Modrinth are the parity authority

On the real Windows development machine, inspect the installed CurseForge and Modrinth binaries, versions, configured library roots, metadata schemas and every observable screen/action/context menu. Do not rely only on public docs or screenshots. Build `docs/launcher-parity-matrix.md` with one row per feature and columns for CurseForge, Modrinth, Basalt reference, Vault status, evidence/test, and caveats. The supplied CurseForge reference demonstrates non-negotiable UX concepts including the left game rail, global search, `My Modpacks / Discover / Browse / Servers / Skins`, `Create / Import / Create Group`, custom groups, rich instance cards and group organization. Inspect the installed app for the rest.

## In-place library contract

Support at least three ownership modes: **external live/mounted in place**, **Vault-managed**, and **explicit clone/copy**. For external instances, discover custom/non-default roots; preserve original paths/source identity; launch from the original folder; read/manage mods, packs, worlds, configs, logs, screenshots, loader/runtime settings from that folder; refresh changes made by the source launcher without re-import; ensure intentional Vault file changes are visible to the source launcher; use provider-specific reversible metadata adapters; prevent unsafe concurrent writes while Minecraft/source launcher is mutating the profile; snapshot metadata before writes; and fall back to read-only mode for unknown schemas rather than guessing.

## Quality floor

The final organizer must cover complete instance organization, create/import/adopt, Modrinth + CurseForge browse/install, dependency resolution, enabled/disabled mods, batch actions, updates/changelogs, provenance/hash/fingerprint, duplicate/conflict detection, resource/shader/data packs, worlds/saves, Java/JVM/game args, loaders, accounts/skins/capes, task progress/cancel/retry/resume, snapshots/rollback/repair, export/share, diagnostics/crash logs and multi-game shell capability flags. Use staging + atomic switching for multi-file mutations, keep recovery journals, validate paths/archives, secure credentials, and never corrupt external launcher metadata.

## First action

Restore the exact 2.9.5 source from `/mnt/data/mcc-work-2.9.5` if available or the Drive Source ZIP otherwise; run the existing 40-suite baseline; finish the pending native Windows Bok acceptance when possible; inspect the installed CurseForge and Modrinth apps/data roots and create the parity matrix; pin the Basalt commit/license strategy; then implement the first vertical slice: enumerate one CurseForge instance and one Modrinth instance **in place with zero copying**, display their real metadata/path/group, and launch them from their original folders. Once that works safely, expand through the complete parity surface. Do not stop after the vertical slice and call the project complete.

Publish checkpoints to both Google Drive `/Google Drive/Minecraft Mod Vault/` and this GitHub repository, with source/build/QA/checksums and exact next action. Preserve the existing split-archive + Windows reassembly pattern for large deliverables.