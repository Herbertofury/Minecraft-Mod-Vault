# TikTok Creator Source Catalog — Minecraft Mod Vault

Recorded: 2026-08-23

This catalog seeds Creator Vault discovery. A creator recommendation is **evidence input**, not authoritative project metadata. The Vault must resolve extracted names/links against normal project providers and preserve the exact source video/profile.

| Source | TikTok | Role | Seed status | Notes |
| --- | --- | --- | --- | --- |
| Kizamiringo | https://www.tiktok.com/@kizamiringo | Short-form Minecraft mod discovery, including text-first lists | Protected/core follow | User-requested primary TikTok parity target. |
| Katsumi | https://www.tiktok.com/@its_katsumi | Short-form Minecraft mod discovery + creator link hub intelligence | Curated-core follow | Current indexed hubs: https://lnk.bio/itskatsumii and https://linktr.ee/Itskatsumii. Linktree index exposes Discord, tips/donos, wishlist, Twitch, YouTube, Instagram and social icons. Exact current modpack child URL was not exposed by the web index, so the live crawler resolves it from the hub instead of hardcoding a guess. |
| SpeedyChunks | https://www.tiktok.com/@speedychunks | Short-form Minecraft mod discovery + automatic creator link/modpack discovery | Curated-core follow | User-requested built-in follow. Public web search did not expose a trustworthy Linktree/CurseForge/Modrinth identity during this checkpoint, so none is hardcoded; the normal live profile/link-hub/modpack discovery pipeline will attach verified destinations when found. |
| NoxusMinecraft | https://www.tiktok.com/@noxusminecraft | Minecraft short-form discovery | Data-driven required follow | User-requested. No unverified external provider/modpack identity is hardcoded; live profile/link-hub/modpack discovery attaches evidence when available. |
| UnyxYT | https://www.tiktok.com/@unyxyt | Minecraft short-form discovery | Data-driven required follow | User-requested. No unverified external provider/modpack identity is hardcoded; live profile/link-hub/modpack discovery attaches evidence when available. |
| CurseForge | https://www.tiktok.com/@curseforge | Official ecosystem feed; mods, modpacks, resource packs | Curated-core follow | Highest-confidence official discovery source; provider resolution still determines the actual project record. |
| HendyVideos | https://www.tiktok.com/@hendyvideos | Dedicated Minecraft mod lists/showcases | Curated-core follow | High recommendation density. |
| Knarfy | https://www.tiktok.com/@itsknarfy | Mod-driven experiments/showcases | Recommended | Useful active discovery signal; not treated as authoritative metadata. |
| The Breakdown | https://www.tiktok.com/@thebreakdownxyz | Mods/loaders/shaders/modpacks/install ecosystem | Recommended | Complements the existing YouTube creator archive. |
| The Crimson Gaming | https://www.tiktok.com/@thecrimsongaming | Dedicated Minecraft mod showcases | Recommended | Secondary archive/discovery source. |
| laveOrc | https://www.tiktok.com/@ygz207 | Current horror/mod clips with project-name callouts | Recommended | Smaller active source; recommendation-only by design. |

## Discovery policy

- Prefer active dedicated Minecraft mod/showcase accounts and official mod ecosystem accounts.
- Keep the built-in followed set intentionally smaller than the recommendation catalog to avoid noisy first-run crawling.
- Let users follow any additional TikTok profile from a full URL or `tiktok:@handle`.
- Never infer that a visually mentioned name is a valid downloadable mod solely because OCR saw it. Resolve it against known providers first.
- Keep unresolved names visible with the source video and evidence so the resolver can improve later.
- Preserve archived recommendations even if a non-core creator is unfollowed.
- First sync backfills history; later background syncs are incremental and retry with backoff on provider failures.


## Creator link-hub policy

- Follow public profile/link-hub evidence, not arbitrary external pages.
- Discover outbound destinations without recursively crawling them.
- Classify creator-owned labels and destination paths into modpack/mod/mod-list/resource-pack/shader/datapack/link-hub/social/download/support/wishlist/website.
- Preserve evidence URL/type and first/last-seen timestamps.
- Use recent archived upload descriptions as a fallback signal when the platform profile HTML hides external links.
- Keep last-known-good creator links if a provider blocks a refresh; remove stale auto-discovered links only after a successful refresh.

## Verified creator modpack-provider profiles — P0.3

These provider profiles are separate from recommendation-source discovery. They are evidence-backed identities used to enumerate a creator's current modpack library; individual packs are provider-discovered rather than frozen into the built-in seed list.

| Creator | Provider profile | Current web evidence (2026-08-23) | Vault behavior |
| --- | --- | --- | --- |
| AsianHalfSquat | https://www.curseforge.com/members/asianhalfsquat/projects | Public profile reports 11 projects and visibly mixes Owner entries with non-owner/member-associated projects. | Seed the verified profile, enumerate all modpacks, preserve owner/member/profile provenance, refresh over time. |
| EnderVerse | https://www.curseforge.com/members/enderverse/projects | Public profile reports 7 projects; visible entries are modpacks marked Owner. | Seed the verified profile and enumerate all current modpacks instead of hardcoding one pack. |

### Creator modpack evidence policy

- A verified provider creator profile is the strongest reusable enumeration anchor.
- A direct CurseForge/Modrinth modpack link from the creator's own profile, Linktree-style hub, or archived upload description is valid project evidence.
- A creator-controlled `My Modpack`/`Our Modpack`/`Official Modpack` label may bridge a differently named provider account only when provider data proves exactly one owner/author.
- Preserve owner, member/collaborator, provider-profile and direct-link relationships distinctly.
- Do not promote fan packs or projects that merely mention the creator as releases by that creator unless provider/team evidence establishes the relationship.
- No trustworthy evidence means an explicit empty/unknown state, never a fabricated modpack.

### AsianHalfSquat hot-drop bootstrap — 2026-08-24

The shipped `asianhalfsquat` catalog targets channel ID `UC0E_vIe1e1lVeojYOgVg_5Q`, records an expected public-channel size of 349 videos, and currently seeds 11 exact video records with 93 evidence-backed recommendations. It is intentionally `complete=false`; the normal network-enabled full-history Creator Vault sync extends the same record rather than creating a second database.
