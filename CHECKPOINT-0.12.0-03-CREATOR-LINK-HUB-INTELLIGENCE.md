# Minecraft Mod Vault 0.12.0 — P0.2 Creator Link-Hub Intelligence verification

Recorded: 2026-08-23
Branch: `checkpoint/creator-linkhub-p02`
Baseline P0.1 commit: `8cd2002fcb806b45f641f79f85ae9cf9a46cec8c`
P0.2 implementation commit: `d968787a067f40b205a0e35693d0bff380ceba86`
Checkpoint identity commit before this receipt: `6533da876482d97a729489f507a451284a0cf57a`

## Requested behavior

- Katsumi (`https://www.tiktok.com/@its_katsumi`) keeps her built-in follow.
- Her currently verified link hubs are part of her Creator Vault record.
- Creator Vault automatically discovers Linktree-style hubs and creator modpacks/related destinations for followed YouTube/TikTok creators.
- Discovery is premium, persistent, provenance-backed, refreshable, and safe rather than a one-off hardcode.

## Current Katsumi public evidence

Current web-index evidence on 2026-08-23 identifies:

- `https://lnk.bio/itskatsumii`
- `https://linktr.ee/Itskatsumii`

The indexed Linktree result currently exposes labels for Discord Server, Tips/Donos, Wishlist, Twitch, YouTube, Instagram, and social icons. The web index did not expose a trustworthy exact current modpack child URL. This checkpoint seeds only the verified hubs and delegates current child discovery to the live crawler; it does not invent a modpack URL.

## Implementation verified

- `creator_profile_links.go` adds public creator-profile/link-hub discovery and persistence.
- Supported link hubs include Linktree, Lnk.Bio, Beacons, Carrd, Campsite, Solo.to, Bio.site, Taplink, AllMyLinks, Milkshake, Hoo.be, Bio.link, Stan Store, Pillar, Snipfeed, Koji, and Flowpage.
- Evidence comes from creator profile pages, provider bio text, up to 24 recent archived upload descriptions, and nested known hubs.
- Crawl is bounded to six public profile/hub pages per refresh.
- Arbitrary outbound targets are recorded but never recursively fetched by this crawler.
- Same-origin platform redirect wrappers are unwrapped; outbound hub redirects are captured without issuing the outbound request.
- Link kinds include modpack, mod, mod-list, resource-pack, shader, datapack, link-hub, social, download, support, wishlist, and website.
- Labels participate in classification, so a generic Drive URL labeled `My Minecraft Modpack` becomes a modpack.
- Source page/evidence type plus first/last-seen timestamps are retained.
- Successful refresh removes stale auto-discovered links while preserving curated seed hubs.
- Complete provider blocking keeps last-known-good links and exposes `cached` state plus a truthful warning.
- Link refresh runs with normal Creator Vault sync and has a dedicated lightweight refresh endpoint.
- Creator Archive cards render typed creator links with modpacks first and a functional Refresh links control.

## Automated verification

Executed from the P0.2 source with the source metadata restored to `go 1.27.0` after compatibility testing:

- `go test -mod=vendor ./...` under host Go 1.23 compatibility metadata: **PASS**
- `go vet -mod=vendor ./...` under the same compatibility metadata: **PASS**
- `node --check web/app.js`: **PASS**
- `node --check web/catalog.js`: **PASS**
- `git diff --check`: **PASS**
- Source `go.mod` after verification: `go 1.27.0`

Regression coverage includes classification, escaped/embedded URL extraction, tracking cleanup, platform redirect unwrapping, profile -> hub -> target traversal, no arbitrary outbound fetch, redirect target capture, upload-description hub discovery, stale-link eviction, blocked-provider cache preservation, Katsumi seed migration, and refresh-API persistence.

## Fresh compatibility builds

Built from the current implementation source using host Go 1.23.2 after a temporary local-only metadata compatibility adjustment; source metadata was restored immediately afterward.

- Linux amd64 raw binary SHA-256: `79639b03f32347cc4aeed8868ff6ea65c68734b41f8635aa3a935e0c658b39f2`
- Windows amd64 raw executable SHA-256: `000d05b469f65ccc2fe5eefd9b304cb8bf394f96a3bb8ef12c4e55450725a07c`

These are checkpoint compatibility builds, not a substitute for the final official Go 1.27 merged release build.

## Fresh packaged-runtime behavior exercised

Fresh Linux config/runtime:

1. Started the freshly built P0.2 Linux binary on loopback.
2. `/api/creators/channels` returned exactly **17** built-in creators.
3. Katsumi loaded as `platform=tiktok`, `required=true`, `source=curated-core`.
4. Katsumi loaded with `profileLinksStatus=seeded` and exactly the two verified hub URLs above.
5. Called the real `POST /api/creators/channels/links/refresh` endpoint against Katsumi.
6. This sandbox blocks outbound DNS. TikTok, Lnk.Bio, and Linktree fetches therefore failed in the real runtime.
7. The endpoint truthfully returned HTTP 502 and stated that the last-known-good cache was kept.
8. A subsequent channel read showed `profileLinksStatus=cached`, both verified hubs still present, and the provider error preserved.
9. Restarted the same built binary against the same config directory.
10. Katsumi still had `cached` status, both hub URLs, and the failure metadata after restart.

This proves the requested persistence/failure QoL path in the real packaged runtime. It does **not** prove a live Linktree child crawl because the container has no outbound DNS.

## UI verification boundary

The backend served the embedded current UI and JavaScript syntax checks pass. Chromium is installed, but this environment applies an organization policy that replaces loopback and `file:` navigations with “Your organization doesn’t allow you to view this site.” Direct production UI screenshot/click verification is therefore externally blocked in this run. The Refresh links control is wired to the exercised real endpoint in `loadCreatorChannels`, but a browser click cannot honestly be claimed here.

## Remaining gates before canonical release completion

- Merge this P0.2 delta into the newest concurrent Minecraft Mod Vault source without overwriting newer work.
- On a networked runtime, refresh Katsumi's current Lnk.Bio/Linktree children and record the exact modpack destination if one is currently published.
- Verify a later creator-hub change is picked up by automatic refresh.
- Complete the existing live Kizamiringo current-yt-dlp history/text-only/incremental gates.
- Exercise Windows.Media.Ocr on Windows.
- Run the final merged source with the official Go 1.27 toolchain and repeat full build/runtime verification.
