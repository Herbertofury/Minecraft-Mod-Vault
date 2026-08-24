# Creator Link-Hub Intelligence — P0.2 checkpoint

Recorded: 2026-08-23
Status: isolated integration checkpoint; merge into the newest concurrent Minecraft Mod Vault source rather than replacing it.

## Goal

Make Creator Vault profiles feel complete rather than video-only. When a followed YouTube/TikTok creator exposes a Linktree-style hub, modpack, mod list, resource pack, shader, social profile, support page, wishlist, or download link, the Vault should discover it automatically, keep it alongside the creator archive, preserve where it was discovered, and refresh it without requiring a full video rescan.

## Katsumi seed

Creator: `https://www.tiktok.com/@its_katsumi`

Current public web-index evidence on 2026-08-23 exposes both of these creator-controlled hubs:

- `https://lnk.bio/itskatsumii` — indexed as `@its_katsumi` / link-in-bio.
- `https://linktr.ee/Itskatsumii` — indexed as `@Itskatsumii` and currently surfaces Discord Server, Tips/Donos, Wishlist, Twitch, YouTube, Instagram, plus TikTok/Discord/Instagram/Twitch/YouTube social icons.

The search index did **not** expose a trustworthy exact current modpack destination URL. The Vault therefore seeds only the verified hub URLs and lets the live hub crawler discover any current modpack/download destination from the creator-controlled page. This avoids fabricating or freezing a stale modpack link.

## Automatic discovery contract

The crawler accepts creator-controlled evidence from:

1. the followed creator profile URL;
2. profile/channel bio text available from provider metadata;
3. up to 24 recent archived upload descriptions for the same creator;
4. known public link-in-bio hubs discovered from those sources;
5. nested known link hubs, bounded to six crawled profile/hub pages per refresh.

Recognized link-hub hosts include Linktree, Lnk.Bio, Beacons, Carrd, Campsite, Solo.to, Bio.site, Taplink, AllMyLinks, Milkshake, Hoo.be, Bio.link, Stan Store, Pillar, Snipfeed, Koji, and Flowpage.

## Safety boundary

The Vault fetches only public YouTube/TikTok profile pages and recognized public link-hub hosts. It **does not** recursively fetch arbitrary outbound destinations.

Outbound URLs are still discovered and stored. If a hub button is implemented as a redirect, the HTTP redirect destination is captured without issuing the outbound request. This lets a CurseForge, Modrinth, Google Drive, Discord, store, or other creator link become first-class metadata without turning the profile crawler into an unrestricted web fetcher.

## Link intelligence

Every discovered link stores:

- canonical URL with tracking parameters removed;
- human label when the source exposes one;
- normalized kind;
- provider/display source;
- evidence page URL;
- evidence type (`seed`, `profile`, `bio`, `video-description`, `link-hub`);
- first-seen and last-seen timestamps.

Current kinds:

- `modpack`
- `mod`
- `mod-list`
- `resource-pack`
- `shader`
- `datapack`
- `link-hub`
- `social`
- `download`
- `support`
- `wishlist`
- `website`

Classification uses both destination and creator label. For example, a generic Google Drive URL labeled `My Minecraft Modpack` is promoted to `modpack`; a CurseForge `/minecraft/modpacks/...` or Modrinth `/modpack/...` URL is recognized directly.

## Refresh behavior

- Creator link refresh runs automatically as part of the normal Creator Vault sync.
- Profile metadata refresh is attempted even when video enumeration fails, so a temporary TikTok/YouTube archive failure does not block creator-link updates.
- A lightweight `POST /api/creators/channels/links/refresh` refreshes creator links independently of the expensive video history workflow.
- Successful refresh replaces stale auto-discovered links while preserving curated seed hubs.
- If every public profile/hub fetch is blocked or rate-limited, the last-known-good links remain visible with `cached` status and a truthful warning.

## Premium UI behavior

Creator cards now render a dedicated Creator Links panel:

- modpacks sort first;
- mod/resource-pack/shader/datapack links receive stronger project styling;
- link hub, social, download, support, wishlist, and website links are typed separately;
- buttons open the exact discovered destination;
- status distinguishes curated seed, live refreshed, cached fallback, and pending discovery;
- the source/evidence URL is preserved in the control tooltip;
- `Refresh links` updates only the creator-link metadata.

## Verification in this checkpoint

Automated tests cover:

- Linktree/CurseForge/Modrinth/social/support/wishlist classification;
- label-aware generic download classification as a modpack;
- anchor and embedded-JSON extraction;
- tracking-parameter removal;
- YouTube redirect-wrapper unwrapping;
- profile -> link hub -> outbound target traversal without fetching outbound targets;
- hub redirect target capture without fetching the destination;
- recent upload-description hub discovery;
- stale-link removal after a successful refresh;
- last-known-good preservation when providers block the refresh;
- Katsumi's two verified hub seeds;
- persistence through the dedicated refresh API.

Remaining live gate: on a networked runtime, refresh Katsumi's current hubs and record the exact live child destinations (including a modpack if one is currently published), then verify a later hub edit is reflected by automatic refresh without losing provenance.
