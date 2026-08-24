# Creator Modpack Library — P0.3

Recorded: 2026-08-23
Implementation branch: `checkpoint/creator-modpack-library-p03`
Implementation commit: `9dd0d90ff04042d0128b5a27a1accd2b2e597cb2`

## Goal

Creator Vault should make a best evidence-backed effort to answer a simple question for every followed YouTube/TikTok creator: **what Minecraft modpacks has this creator actually released or contributed to?**

The answer is a library, not a single URL. A creator may have zero, one, or many packs across CurseForge and Modrinth. The Vault must enumerate every verifiable pack it can find, retain provider/profile provenance and relationship, keep the library refreshed, and show an explicit empty/unknown state instead of inventing a project.

## Discovery ladder

1. Use verified creator provider profiles already known to Creator Vault.
2. Discover CurseForge/Modrinth creator profiles from the creator's own public YouTube/TikTok profile, bio, recognized link hub, or recent archived upload descriptions.
3. Accept direct CurseForge/Modrinth modpack links from the same creator-controlled evidence.
4. Enrich a direct project through the provider. If the uploader identity matches the creator, promote the provider profile.
5. When provider usernames differ from the public creator name, an explicit creator-controlled `My Modpack` / `Our Modpack` / `Official Modpack` label may promote a provider identity only when provider evidence resolves exactly one owner/author. Ambiguous ownership is never guessed.
6. Once a verified provider profile exists, enumerate **all** of its modpack projects, not only the originally linked project.
7. Refresh the creator modpack library independently and as part of normal creator background refresh.

## Provider behavior

### CurseForge

Preferred path when an API key is configured:

- Resolve the author identity.
- Query Minecraft modpacks using `authorId` for projects where the creator is a member.
- Query `primaryAuthorId` separately for projects the creator owns.
- Preserve `owner` vs `member` relationship instead of flattening them.
- Paginate up to the provider limits used by the app.

Fallback without an API key:

- Fetch only the public CurseForge creator member/project page or an exact direct modpack page needed for author attribution.
- Enumerate every modpack project URL visible/serialized on the creator profile.
- Report relationship as `profile` when the public page does not prove owner/member status.
- A creator-controlled `My Modpack` link can promote a differently named CurseForge account only if the direct project page exposes exactly one unique CurseForge member identity.

### Modrinth

- Use the official user-project endpoint and filter to `project_type=modpack`.
- Query each project team to distinguish `owner` from other team/member roles when available.
- A direct creator-controlled modpack link may promote a differently named Modrinth profile only when the project team proves exactly one owner.

## Link-hub integration

Creator Link-Hub Intelligence remains the evidence intake layer. Linktree, Lnk.Bio, Beacons, Carrd and other recognized hubs can surface:

- direct modpack links;
- provider creator profiles;
- generic destinations labelled as a modpack;
- later changes to those destinations.

The bounded link-hub crawler still does **not** recursively crawl arbitrary outbound websites. CurseForge and Modrinth are handled by the dedicated provider-aware modpack resolver with their own narrow fetch/API boundaries.

## Persistent record

Each verified creator pack keeps:

- provider and project ID/slug;
- exact project URL and provider creator-profile URL;
- title, summary/icon when provider data exposes them;
- author/provider identity;
- relationship (`owner`, `member`, `profile`, `linked`);
- downloads/update time/game versions/loaders when available;
- evidence URL + evidence type;
- verified/first-seen/last-seen timestamps.

Provider failures preserve last-known-good packs. A successful authoritative provider refresh can retire packs no longer returned by that provider profile.

## UI contract

Each Creator Vault card now exposes **Creator modpacks** above generic creator links:

- every verified pack is listed; no artificial single-pack assumption;
- provider and relationship are visible;
- exact project buttons open the exact CurseForge/Modrinth destination;
- large libraries can collapse as one details section without truncating stored results;
- `none` explicitly says no verified pack is attached yet and that the Vault will not invent one;
- partial provider failures visibly retain cached packs with a warning;
- `Refresh links + packs` refreshes creator metadata without forcing a full historical video rescan.

## Current acceptance fixtures (web-verified 2026-08-23)

- AsianHalfSquat CurseForge profile: `https://www.curseforge.com/members/asianhalfsquat/projects` — current public page reports 11 projects and visibly mixes Owner projects with non-owner/member-associated entries. This is why ownership and membership must remain distinct.
- EnderVerse CurseForge profile: `https://www.curseforge.com/members/enderverse/projects` — current public page reports 7 projects; visible entries are modpacks and marked Owner.
- CurseForge REST API documents `authorId` as projects where an author is a member and `primaryAuthorId` as projects where the author is the owner.
- Modrinth official API exposes `GET /user/{id|username}/projects`; project-team membership is used for relationship evidence.

These profile URLs are seeded as evidence-backed provider profiles. Their individual project lists are intentionally **not hardcoded** so Creator Vault can track changes.

## No-fabrication rule

A name resemblance, search result, fan pack, or pack description mentioning a creator is not sufficient to claim ownership. If provider/profile/direct-link evidence cannot establish a relationship, Creator Vault retains `none`/unknown rather than creating a false creator pack.
