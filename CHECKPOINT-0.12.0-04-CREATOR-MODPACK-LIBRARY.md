# Minecraft Mod Vault 0.12.0 — P0.3 Creator Modpack Library verification

Recorded: 2026-08-23
Branch: `checkpoint/creator-modpack-library-p03`
Implementation commit: `9dd0d90ff04042d0128b5a27a1accd2b2e597cb2`
Base checkpoint: P0.2 runtime evidence commit `836a8e807f1fd98881c751f263ff6f20be81bafa`

## Implemented

- Persistent multi-pack creator libraries for CurseForge and Modrinth.
- Evidence-backed provider-profile discovery from existing creator link-hub/profile/description evidence.
- Direct modpack links can safely promote a provider identity when provider evidence proves the creator/unique owner.
- Differently named provider accounts are supported without fuzzy-name guessing.
- CurseForge API path distinguishes member (`authorId`) from owner (`primaryAuthorId`).
- CurseForge public-profile fallback enumerates all public modpack project links when no API key exists.
- Modrinth official user-project enumeration plus project-team relationship checks.
- Stale provider entries retire only after successful provider refresh; failed refresh preserves last-known-good packs.
- Built-in AsianHalfSquat and EnderVerse Creator Vault records seed their verified CurseForge creator profiles, not a frozen project list.
- Dedicated `POST /api/creators/channels/modpacks/refresh` endpoint.
- Existing `Refresh links + packs` path refreshes both creator links and modpack library without a historical video rescan.
- Premium Creator Modpacks UI lists all verified packs, provider, relationship, metadata and exact project destination, with explicit no-verified-pack and partial-cache states.

## Regression verification

- `go test -mod=vendor ./...` — PASS using the documented temporary host-Go metadata compatibility procedure.
- `go vet -mod=vendor ./...` — PASS under the same procedure.
- `node --check web/app.js` — PASS.
- `node --check web/catalog.js` — PASS.
- `git diff --check` — PASS.
- Source metadata restored to `go 1.27.0` immediately after each compatibility check/build.

Coverage includes:

- provider/profile URL classification;
- Modrinth multi-pack enumeration and owner/member roles;
- CurseForge API owner/member distinction;
- CurseForge public profile fallback retaining all pack links;
- direct creator-linked Modrinth `My Modpack` promoting a differently named unique owner and discovering the rest of that owner's packs;
- direct creator-linked CurseForge `My Modpack` doing the same from a single public author without an API key;
- no-evidence => zero fabricated packs;
- provider failure => last-known-good packs retained;
- AsianHalfSquat/EnderVerse verified profile seeds;
- API persistence across reload.

## Fresh raw builds

- Linux x64 raw binary SHA-256: `2b350313cf18961b82c4da8ef8a6a1489388c4669edc75ace0f6fe607ed6620a`
- Windows x64 raw executable SHA-256: `98e43a542e2b388ed25123dffba32f7a527752f8f6f1decce1068b2bdb31ecec`

These are compatibility checkpoint builds created by temporarily lowering only the local Go metadata for the available Go 1.23 host. Canonical source remains Go 1.27.0; official Go 1.27 merged-release verification remains a gate.

## Packaged-runtime provider proof

The Linux executable was freshly extracted from the actual deliverable tarball and run with a fresh configuration against a local provider-faithful CurseForge fixture through `MMV_CURSEFORGE_WEB_BASE`:

- AsianHalfSquat fixture exposed 4 distinct profile modpacks; the real compiled API returned all 4 (`dark-and-dangerous`, `oceanum`, `satisfaction-guaranteed`, `shattered-ring`).
- EnderVerse fixture exposed 3 distinct profile modpacks; the compiled API returned all 3 (`breathful`, `fresh-and-smooth`, `mc-dungeons-reforged`).
- Process was stopped and restarted using the same fresh config.
- AsianHalfSquat restored exactly 4 creator modpacks with `creatorModpacksStatus=ready`.
- EnderVerse restored exactly 3 creator modpacks with `creatorModpacksStatus=ready`.
- Built-in creator count remained 17.

The executable extracted from the deliverable package is byte-identical to the raw Linux build (`2b350313...`), so this proves the packaged backend's multi-pack enumeration + persistence path. It is not presented as a live CurseForge network crawl.

## Current external evidence

Web verification on 2026-08-23 independently confirmed:

- AsianHalfSquat's current CurseForge profile reports 11 projects and includes both Owner and non-owner/member-associated projects.
- EnderVerse's current CurseForge profile reports 7 projects; visible entries are modpacks marked Owner.
- CurseForge's current API docs distinguish `authorId` member projects from `primaryAuthorId` owner projects.
- Modrinth's official API documents user-project enumeration.

## Remaining release gates

1. Port this P0.3 delta into the newest concurrent Minecraft Mod Vault source without overwriting newer work.
2. On a network-enabled app runtime, refresh the real AsianHalfSquat and EnderVerse provider profiles and compare the returned library against the current provider pages.
3. Run a real Linktree/Lnk.Bio creator-controlled direct-modpack example through the packaged app and confirm profile promotion + discovery of additional provider packs when applicable.
4. Perform production-browser click/screenshot verification where browser policy permits.
5. Run the merged source under the official Go 1.27 toolchain plus the existing unresolved TikTok/OCR live gates.

Status: **P0.3 implementation checkpoint ready for packaging/integration; not canonical release completion.**
