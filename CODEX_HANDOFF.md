# CODEX START HERE — Minecraft Mod Vault / Minecraft Catalog Companion — FULL BASALT INTEGRATION

## Non-negotiable continuity rule

Continue the existing Minecraft Mod Vault / Minecraft Catalog Companion project from the 2.9.5 state. Do **not** restart it, throw away the Companion, or build a separate launcher beside it. The goal is one unified product.

**Codex does not have access to this ChatGPT session's `/mnt/data` or any other chat-local filesystem path. Ignore every old handoff reference to `/mnt/data`. The canonical handoff and all project-owned inputs Codex needs are on the user's Google Drive under:**

`/Google Drive/Minecraft Mod Vault/`

If an older note disagrees with this document, this document wins.

## Canonical Drive inputs

Start from these Drive files, not from chat-local paths:

- `/Google Drive/Minecraft Mod Vault/Minecraft Catalog Companion 2.9.5 - Source.zip`
  - Drive file id: `external-gdrive:file:1XfzDd-l510jQXpID1-v4JoyGbPH1q5y0`
  - This is the editable 2.9.5 Companion source snapshot and is the primary project source to restore first.
- `/Google Drive/Minecraft Mod Vault/Minecraft Catalog Companion 2.9.5 - QA Results.json`
  - Drive file id: `external-gdrive:file:1dyJmS0kxOI1nNyxMmQWjrYCWyfruBixT`
- `/Google Drive/Minecraft Mod Vault/Minecraft Catalog Companion 2.9.5 - Project Checkpoint.json`
  - Drive file id: `external-gdrive:file:1J8fpb5e8l8XmmKzGRgOAiymjGF6Vugdr`
- `/Google Drive/Minecraft Mod Vault/Minecraft Catalog Companion 2.9.5 - SHA256.txt`
- `/Google Drive/Minecraft Mod Vault/Minecraft Catalog Companion 2.9.5 - Windows x64.zip.part01`
- `/Google Drive/Minecraft Mod Vault/Minecraft Catalog Companion 2.9.5 - Windows x64.zip.part02`
- `/Google Drive/Minecraft Mod Vault/Minecraft Catalog Companion 2.9.5 - Windows x64.zip.part03`
- `/Google Drive/Minecraft Mod Vault/Minecraft Catalog Companion 2.9.5 - Reassemble.cmd`
- `/Google Drive/Minecraft Mod Vault/Minecraft Mob Variety - Explorer 2.9.5.html`
- `/Google Drive/Minecraft Mod Vault/Minecraft Mob Girl & Female Mob Vault - Explorer 2.9.5.html`
- `/Google Drive/Minecraft Mod Vault/CurseForge parity reference - modpack library.png`
- `/Google Drive/Minecraft Mod Vault/CODEX HANDOFF - Minecraft Mod Vault.md` — this handoff after replacement/update.

The GitHub mirror is `Herbertofury/Minecraft-Mod-Vault`. The 2.9.5 release checkpoint is documented at `companion/releases/2.9.5/README.md`. Use GitHub as a mirror/history/CI surface, but Google Drive is the canonical external handoff store for this transfer.

## User authorization for Basalt — direct source integration is explicitly intended

The project owner states that they have **full permission/rights to use, copy, modify, merge, adapt, and directly integrate the Basalt Launcher source into this project**. This is not a request for a clean-room reimplementation, a behavioral clone, or merely borrowing "useful capabilities". For this project, treat direct source adoption and modification as authorized by the user.

Upstream Basalt:

- Repository: `https://github.com/MegalithOfficial/basalt-launcher`
- Pinned current `main` commit for this handoff: `00b4f9915def3d201e0a411c3b5453940849c683`
- Main tree at that commit: `5103b7dae640f92e7bc05f583f437b1043e0eaaa`
- Upstream stack: **Rust backend + Tauri 2 desktop runtime + React/TypeScript frontend**.

Do not choose the previous "clean-room rewrite" route. The user's instruction is to **fully integrate Basalt itself**, including its Rust backend, React/TypeScript launcher/mod-manager frontend, data models, native services, and relevant Tauri-era architecture/source, into the Vault project. **However, do not replace the existing Minecraft Catalog Companion Electron shell.** The Catalog remains Electron-based and is the primary desktop shell; the fully integrated Mod Manager/Launcher appears as a first-class top-level tab/workspace inside that same Electron application. Preserve any copyright/attribution/license notices or other terms still required by the permission actually granted to the user, but do not treat ordinary GPL compatibility uncertainty as a reason to fall back to a non-integrated reimplementation; if a specific distribution term truly remains unknown, flag that exact term without blocking engineering work unnecessarily.

Because Basalt is an external public upstream, Codex may fetch the pinned repository directly from GitHub when beginning the integration. All **project-owned continuity/source/checkpoint/reference material** from this chat is on Drive; no chat-local filesystem is required.

## What "FULL integration" means

This is an architectural merger, not a plugin bridge and not two apps glued together.

### Final technical foundation — Electron Catalog stays, Mod Manager becomes a first-class tab

The intended final desktop application is a **single Electron-based Minecraft Catalog Companion shell** with the existing Catalog/Browser/Source experience preserved, plus a fully integrated Launcher / Mod Manager workspace that the user can tab into instantly. Basalt is merged deeply into that application rather than replacing the shell:

- **Electron remains the primary desktop runtime/shell. This is non-negotiable.** Preserve the existing Catalog windowing, `WebContentsView` behavior, persistent Chromium sessions/cookies, live-provider browsing, Source/Browser/Full flows, media/gallery pipeline, and the premium responsive feel already built into the Companion.
- Add a first-class top-level **Mod Manager / Launcher** tab/workspace alongside the Catalog surfaces. Switching between Catalog and Mod Manager should feel native and immediate, with shared navigation/state where useful. It must not launch a disconnected second app.
- **Rust remains the authoritative launcher/mod-organizer backend**, directly integrating/adapting Basalt's Rust source for filesystem, database/migrations, tasks, networking, credentials, downloads, launch planning/process management, recovery/snapshots, content providers, Java/loaders, accounts, worlds, servers, logging, etc.
- **React + TypeScript** from Basalt may be directly merged/adapted for the Mod Manager/Launcher UI inside the Electron renderer. Preserve useful Basalt components/state patterns, but make them conform to the unified Companion design system and navigation.
- Basalt's **Tauri-specific shell/IPC glue is migration input, not permission to replace Electron**. Reuse/adapt Tauri-era Rust modules and application architecture, but expose them to Electron through a robust native IPC boundary (for example a managed Rust companion process/service, native module, or other production-grade transport). The user wants Basalt's real Rust implementation and full feature set, not necessarily a second Tauri window.
- If any Basalt functionality is tightly coupled to Tauri APIs, port that adapter layer while preserving the underlying Rust logic. Do not downgrade the feature or rewrite it in JavaScript merely to avoid Rust integration.
- The final experience must look like the existing Minecraft Catalog Companion gained a complete Mod Manager/Launcher tab, **not** like the Catalog was migrated into or replaced by Basalt.

### Basalt subsystems to integrate, not merely imitate

Audit Basalt source end-to-end and directly integrate/adapt all relevant code and architecture, including at minimum:

- `src-tauri` Rust backend and command structure;
- application state/database/migrations;
- instance model and storage model;
- filesystem/path validation and managed/external roots;
- network/download manager, retries/resume/checksums/proxy/TLS handling;
- credential/keychain handling and Microsoft account auth/refresh;
- Mojang/game metadata and launch planning;
- Java discovery/selection/runtime handling;
- Fabric, Quilt, Forge and NeoForge installation/management;
- Modrinth and CurseForge providers, compatibility filtering, dependencies, changelogs and updates;
- modpack creation/install/upgrade/import/export behavior;
- mods/resource packs/shaders/data packs/content management;
- enabled/disabled state and batch operations;
- worlds/saves and instance file management;
- snapshots, backups, rollback, repair and recovery journals;
- tasks/progress/cancel/retry/recovery UI and backend;
- logs, game output, diagnostics and crash analysis;
- accounts, skins and capes;
- playtime/stats where useful;
- server management, including the newer Basalt server work present before the pinned main state;
- updater/packaging/build/release workflows where appropriate;
- Basalt frontend views/components/store patterns that are genuinely part of the launcher experience.

Do not cherry-pick a small subset and call that "Basalt integration". Produce a subsystem migration/parity checklist and keep incomplete Basalt functionality visible as blockers until it is either integrated or deliberately superseded by something stronger.

## Preserve and port the Companion — no regression deal

The merged product must also retain every useful existing Companion/Vault capability, including the research/catalog side that Basalt does not have:

- rich Minecraft mod catalog and recommendation browsing;
- Source / Browser / Full details flows;
- live provider media and galleries;
- CurseForge/Modrinth/other site adapters;
- persistent browsing/auth sessions where needed;
- provider identity/role safety and media isolation;
- no arbitrary gallery/card caps;
- catalog ingest/sync;
- favorites, notes and research workflows;
- creator/reference catalogs and existing project data;
- existing QA behavior, including the 2.9.5 CurseForge media fixes.

Port this behavior into the new architecture rather than deleting it. The final app should feel like the Companion grew into a full launcher/mod organizer, not like the Companion was replaced by stock Basalt.

## CurseForge + Modrinth parity is a release contract

Inspect the **actually installed CurseForge and Modrinth launchers on the user's Windows development machine**, including their current versions, configured directories, metadata formats, every screen/tab/menu/context action and observable behavior. Do not rely only on documentation or screenshots.

Create and maintain `docs/launcher-parity-matrix.md` with one row per capability and columns such as:

`Feature | CurseForge installed | Modrinth installed | Basalt upstream | Vault target | Implemented | Runtime proof | Notes/blocker`

If the installed CurseForge or Modrinth launcher can do something relevant to Minecraft launcher/modpack/mod-organizer operation, the Vault must support it too or explicitly mark it as an unresolved release blocker.

The supplied CurseForge screenshot already proves several non-negotiable UI concepts:

- game selector rail on the left;
- global search;
- `My Modpacks`;
- `Discover`;
- `Browse`;
- `Servers`;
- `Skins` / account-oriented surfaces;
- `Create`;
- `Import`;
- `Create Group`;
- custom user groups such as Favorites and Old;
- rich instance cards;
- per-instance context menus;
- game/version badges;
- complete modpack organization rather than a flat list.

Inspect the installed apps for everything else and reproduce or improve it.

## External CurseForge + Modrinth libraries must work IN PLACE

This is another non-negotiable goal. Do not force users to duplicate gigabytes of packs just so Vault can manage them.

Support at least three ownership modes:

1. **External live / mounted in place** — preferred when discovering installed CurseForge or Modrinth instances.
2. **Vault-managed** — instances created directly by Vault.
3. **Explicit clone/copy** — optional user action when they intentionally want an independent copy.

For an externally mounted instance, Vault must be able to:

- automatically discover default and custom CurseForge/Modrinth library roots;
- enumerate the original instances without copying them;
- retain source-launcher identity and original path;
- launch Minecraft from the original instance folder;
- read and manage its mods, configs, resource packs, shaders, data packs, worlds, logs, screenshots and runtime settings in that same location;
- pick up changes made later by CurseForge/Modrinth without another import;
- make deliberate Vault changes visible to the source launcher when their metadata model safely permits it;
- use provider-specific metadata adapters rather than rewriting foreign metadata blindly;
- watch/reconcile external changes;
- detect source launcher/game activity and prevent conflicting writes;
- snapshot/back up metadata before mutations;
- journal risky operations and recover interrupted writes;
- fall back to read-only if a foreign schema is unknown/new rather than corrupting it;
- offer "Clone/Copy into Vault" separately, never as a prerequisite.

The UX should make ownership/source obvious but not punish the user for using multiple launchers.

## Full mod organizer scope

The target is no longer merely a launcher. Build a serious mod organizer around the merged Basalt/Vault foundation:

- folders/groups/collections/favorites/tags and user-defined organization;
- grid/list layouts, search, filters and sorting;
- drag/drop where safe;
- bulk selection and batch enable/disable/update/delete/move operations;
- installed mod identification from filenames + hashes/fingerprints/provider metadata;
- provenance/source tracking and links;
- update availability and changelog review;
- dependency and optional-dependency intelligence;
- duplicate/mod-conflict detection;
- loader/game-version compatibility checks;
- missing dependency detection and repair;
- enabled/disabled mod state without destructive deletion;
- profile comparisons/diffs;
- profile cloning;
- resource packs, shaders, data packs and worlds managed alongside mods;
- config discovery/edit/open-folder workflows;
- snapshots/backups/rollback;
- import/export/share/pack formats such as Modrinth/CurseForge/packwiz where applicable;
- launch, kill, restart, safe-mode and diagnostic workflows;
- server management parity where Basalt/CurseForge provide it;
- crash/log analysis and actionable repair suggestions;
- Java/JVM/loader/game argument management globally and per instance;
- multiple Microsoft accounts, skins and capes;
- task/download queue with pause/cancel/retry/resume and useful error recovery;
- storage management and movable roots;
- no fake UI: every visible control must be backed by working behavior.

## Multi-game shell

CurseForge's left rail is not decoration. Model games as data/capabilities, not hard-coded Minecraft-only navigation. Minecraft remains the first fully implemented game, but the side rail/navigation architecture should be capable of multiple games with per-game tabs/actions and should not require a frontend rewrite to add another supported game later.

## First execution plan for Codex

1. Open `/Google Drive/Minecraft Mod Vault/CODEX HANDOFF - Minecraft Mod Vault.md` and this Drive folder.
2. Restore `/Google Drive/Minecraft Mod Vault/Minecraft Catalog Companion 2.9.5 - Source.zip` into Codex's own workspace.
3. Read the 2.9.5 QA/checkpoint/checksum files and run the existing baseline before migration.
4. Fetch/pin `MegalithOfficial/basalt-launcher` at `00b4f9915def3d201e0a411c3b5453940849c683` into Codex's own workspace. Direct source integration is authorized by the project owner.
5. Inventory Basalt source by subsystem and generate `docs/basalt-full-integration-matrix.md`; this is an exhaustive merge checklist, not a feature wish list.
6. Inspect the installed CurseForge and Modrinth applications and their configured data/library roots on the Windows machine; generate `docs/launcher-parity-matrix.md`.
7. Decide the repository layout for the unified **Electron shell + Rust backend + React/TypeScript Mod Manager** application. Preserve the current Electron Catalog as the primary shell, define the native Rust IPC/service boundary, and document a staged integration that keeps the existing Companion testable at every checkpoint.
8. Port/integrate Basalt's backend foundation first, then wire one real vertical slice end-to-end: discover one existing CurseForge instance and one existing Modrinth instance **in place with zero copying**, display full metadata/source/path/group/content state, and launch each from its original folder.
9. Immediately continue beyond that slice through the complete Basalt + CurseForge + Modrinth parity matrices. Do not stop at a demo and call the project complete.
10. **Do not port the Companion Catalog away from Electron.** Keep its catalog/research/media/browser system in place, preserve its regression tests and behavior, and wire the new Rust-backed Mod Manager/Launcher into it as a first-class tab/workspace. Fix the pending real Windows Bok gallery acceptance if it is still reproducible.
11. Build native Windows release candidates and run real workflows against throwaway/test instances before modifying valuable user packs.
12. Publish every meaningful checkpoint to both GitHub and `/Google Drive/Minecraft Mod Vault/`, including source snapshot, built artifact, QA report, checksums, parity matrices and exact next action.

## Definition of done

Do not call this integration complete merely because the new app starts or launches Minecraft. Completion requires:

- Basalt source/architecture substantially merged into the actual project, with **Rust as the authoritative launcher/mod-manager backend while Electron remains the primary application shell/runtime**;
- Companion catalog/research/browser/provider functionality preserved in the merged application;
- full installed CurseForge + installed Modrinth parity matrix closed or explicitly accepted by the user row-by-row;
- existing CurseForge and Modrinth libraries usable in place without mandatory copying;
- optional clone/copy path working;
- full mod-organizer workflows working;
- real native Windows launch/content/update/recovery tests passing;
- no data corruption of external launcher profiles;
- project source/build/QA/checksums/checkpoint published to Google Drive and GitHub.

## One-paragraph directive

Fully merge `MegalithOfficial/basalt-launcher` into Minecraft Mod Vault / Minecraft Catalog Companion — **direct source integration is authorized by the project owner, and "full integration" means adopting and evolving Basalt's actual Rust backend, React/TypeScript launcher/mod-manager frontend, database/state/task/network/credential/download/launch/Java/loader/account/content/world/server/recovery systems and relevant Tauri-era source/architecture, not merely copying features** — while **keeping the existing Minecraft Catalog Companion Electron application as the primary shell and preserving its current Catalog/Browser/Source/provider/media experience exactly as a first-class part of the product**. Add the fully integrated Launcher / Mod Manager as a top-level tab/workspace the user can switch to inside that same Electron app; do not replace the Catalog with Tauri and do not open a disconnected second launcher. Adapt Basalt's Tauri-specific glue so its real Rust core is exposed cleanly to Electron through production-grade native IPC while retaining Basalt's real implementation and complete feature set. Inspect the actually installed CurseForge and Modrinth launchers on the Windows machine and treat complete feature/UI/workflow parity as a release contract, including the left game rail, all tabs, create/import/group/favorite organization, instance actions and every other observable capability. Discover CurseForge and Modrinth libraries and use their existing instances **in place with zero mandatory copying**, launch/manage them where they already live, safely reconcile external changes, and keep explicit clone/copy as an option. The end result is the same premium Electron-based Minecraft Catalog Companion, now with a seamlessly tabbed full Rust-powered launcher + mod organizer, complete modpack/content/account/server/world/backup/repair/update/diagnostic management, data-driven multi-game navigation, strong recovery/data-safety guarantees, native Windows acceptance tests, and every checkpoint/source/build/QA/checksum published to `/Google Drive/Minecraft Mod Vault/` and `Herbertofury/Minecraft-Mod-Vault`.