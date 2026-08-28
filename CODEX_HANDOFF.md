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

Do not choose the previous "clean-room or sidecar" route. The user's instruction is to **fully integrate Basalt itself**, including its Rust/Tauri architecture and source, into the Vault project. Preserve any copyright/attribution/license notices or other terms still required by the permission actually granted to the user, but do not treat ordinary GPL compatibility uncertainty as a reason to fall back to a non-integrated reimplementation; if a specific distribution term truly remains unknown, flag that exact term without blocking engineering work unnecessarily.

Because Basalt is an external public upstream, Codex may fetch the pinned repository directly from GitHub when beginning the integration. All **project-owned continuity/source/checkpoint/reference material** from this chat is on Drive; no chat-local filesystem is required.

## What "FULL integration" means

This is an architectural merger, not a plugin bridge and not two apps glued together.

### Final technical foundation

The intended final desktop application should use Basalt's modern architecture as the primary application foundation:

- **Rust** owns the launcher/mod-organizer backend and durable application behavior.
- **Tauri 2** becomes the primary desktop runtime/shell.
- **React + TypeScript** remains the frontend layer, adapted/merged from Basalt and the Companion as appropriate.
- Basalt's native application patterns — commands, Rust state, filesystem/path handling, database/migrations, tasks, networking, credentials, downloads, launch planning/process management, recovery/snapshots, content providers, Java/loaders, accounts, worlds, servers, logging, packaging and CI concepts — should be brought over directly and evolved as one codebase.
- Existing Electron/Node code in the 2.9.5 Companion is **migration source**, not a sacred final architecture. Port its unique capabilities into the Rust/Tauri application. Electron may be used temporarily as a migration scaffold only if needed; it should not remain as a second permanent launcher runtime just because the current Companion started there.
- Where the Companion has browser/research behavior that depended on Electron `WebContentsView`, persistent Chromium sessions, provider DOM fallback, etc., reproduce or improve that behavior inside the final Tauri-based application using the strongest practical WebView2/webview/native/browser-service approach. Do not delete those features merely to make the migration easier.

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
7. Decide the new repository layout for the unified **Rust + Tauri 2 + React/TypeScript** application and document a staged migration that keeps the Companion testable at each checkpoint.
8. Port/integrate Basalt's backend foundation first, then wire one real vertical slice end-to-end: discover one existing CurseForge instance and one existing Modrinth instance **in place with zero copying**, display full metadata/source/path/group/content state, and launch each from its original folder.
9. Immediately continue beyond that slice through the complete Basalt + CurseForge + Modrinth parity matrices. Do not stop at a demo and call the project complete.
10. Port the Companion catalog/research/media system into the new Tauri foundation and keep its regression tests; fix the pending real Windows Bok gallery acceptance as part of the migration if it is still reproducible.
11. Build native Windows release candidates and run real workflows against throwaway/test instances before modifying valuable user packs.
12. Publish every meaningful checkpoint to both GitHub and `/Google Drive/Minecraft Mod Vault/`, including source snapshot, built artifact, QA report, checksums, parity matrices and exact next action.

## Definition of done

Do not call this integration complete merely because the new app starts or launches Minecraft. Completion requires:

- Basalt source/architecture substantially merged into the actual project, including Rust/Tauri as the primary backend/runtime;
- Companion catalog/research/browser/provider functionality preserved in the merged application;
- full installed CurseForge + installed Modrinth parity matrix closed or explicitly accepted by the user row-by-row;
- existing CurseForge and Modrinth libraries usable in place without mandatory copying;
- optional clone/copy path working;
- full mod-organizer workflows working;
- real native Windows launch/content/update/recovery tests passing;
- no data corruption of external launcher profiles;
- project source/build/QA/checksums/checkpoint published to Google Drive and GitHub.

## One-paragraph directive

Fully merge `MegalithOfficial/basalt-launcher` into Minecraft Mod Vault / Minecraft Catalog Companion — **direct source integration is authorized by the project owner, and "full integration" means adopting and evolving Basalt's actual Rust backend, Tauri 2 runtime, React/TypeScript frontend architecture, database/state/task/network/credential/download/launch/Java/loader/account/content/world/server/recovery systems, not merely copying features or bolting on a sidecar** — while porting and preserving every existing Companion catalog/research/provider/browser capability into that unified architecture. Inspect the actually installed CurseForge and Modrinth launchers on the Windows machine and treat complete feature/UI/workflow parity as a release contract, including the left game rail, all tabs, create/import/group/favorite organization, instance actions and every other observable capability. Discover CurseForge and Modrinth libraries and use their existing instances **in place with zero mandatory copying**, launch/manage them where they already live, safely reconcile external changes, and keep explicit clone/copy as an option. The end result is one premium "it just works" launcher + mod organizer + research/catalog application, with Rust/Tauri as the final native foundation, full modpack/content/account/server/world/backup/repair/update/diagnostic management, data-driven multi-game navigation, strong recovery/data-safety guarantees, native Windows acceptance tests, and every checkpoint/source/build/QA/checksum published to `/Google Drive/Minecraft Mod Vault/` and `Herbertofury/Minecraft-Mod-Vault`.