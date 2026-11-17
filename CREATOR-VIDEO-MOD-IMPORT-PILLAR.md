# Video-to-mod discovery and download pillar

Owner requirement: pasted videos, Shorts and entire creator video archives should reveal every mentioned mod and let the user obtain the identified projects without manually hunting each one. This is a shared behavior contract for modding tools and databases, not a replacement of existing applications.

## Existing implementation to build on

Minecraft-Mod-Vault's Creator Vault already has YouTube/TikTok channel enumeration, captions/local speech transcription, sampled-frame OCR, provider resolution, unresolved mentions, persistent creator catalogs and a recommendation browser. See `creators.go`, `creator_channels.go`, `creator_ocr.go`, `creator_catalogs.go`, issue #5 and release gate #40. Tests/fixtures do not establish complete live-source coverage.

This source-identity checkpoint integrates directly into that extractor: reject lookalike provider hosts, retain original source links/timestamps, identify Modrinth/CurseForge projects rather than version/file IDs, and deduplicate by project identity instead of display name. The next checkpoint adds Bilibili single-BV-video public metadata/caption ingestion through the existing analyze endpoint. X ingestion and bulk downloading remain pending; Bilibili public-upload listing is implemented with live enumeration still access-limited.

## Required continuation backlog (not shipped)

- [ ] Shared platform adapter contract for YouTube (videos, Shorts, playlists, creator archives), Bilibili, X and additional supported video sources. Each adapter exposes supported URL types, pagination, source identity, captions, descriptions, linked articles, media availability and explicit capability/error states. Never silently route an unsupported source to YouTube.
- [ ] Full archive and incremental discovery with resumable page cursors, per-item retry/backoff, cancellation, failed-item review and last-good persistence. Account access, deleted/private/age-restricted items, rate limits and paywalls remain distinct states, not empty successful archives.
- [ ] Combine description links, pinned creator context when available, captions, speech transcription and on-screen text. Preserve original URL, creator, video ID, timestamp, method and confidence for each occurrence, including repeated mentions; expose ambiguous/unresolved candidates instead of guessing.
- [ ] Resolve names to provider project identity, game, edition, loader and version. Compare source context before accepting search results; matching titles alone are not proof. Keep incompatible and ambiguous matches visible in review.
- [ ] Cross-video review/download manifest: deduplicate by authoritative project identity, retain every referring video, show artwork, confidence, compatibility, installed status, version choice, license and access requirements. Support select all, filtering and individual overrides.
- [ ] Dispatch user-selected compatible releases through provider download integrations; verify files/checksums when supplied, report per-item progress/errors, resume safely and distinguish discovery, download and installation. Do not treat a project page URL as a binary, silently purchase access or auto-install executable content.
- [ ] Share schemas/adapters and regression fixtures with Athena, MO2R, ModHunter, Noxviola and databases without requiring Athena at runtime. New reusable core components should be Rust-first; preserve this existing Go integration until measured migration establishes compatibility and benefit.
- [ ] Validate current real YouTube, Shorts, TikTok, Bilibili and X examples against a human-checked complete mod list, plus deleted media, repeated mentions, title collisions, broken links, login-required sources and interrupted large archives. Report recall/precision and unresolved items rather than promise zero errors.

Visual feedback and organization remain first-class: rich source cards, image carousel/video preview where available, evidence links, sortable/filterable project columns, saved views, meaningful stage/progress counters and cancel/pause/resume controls. Pending work stays visible until verified, not hidden behind a generic success message.

## Bilibili single-video checkpoint

Implemented: strict BV URL/part validation; official-host public metadata and available subtitle retrieval; namespaced per-part IDs; canonical creator/artwork/source attribution; reuse of shared mod extraction; correct platform-aware transcript timestamp links. Errors distinguish unavailable/private content, authentication/access requirements and rate limits. Caption failures retain description analysis with explicit coverage warnings. Subtitle redirects must stay within the trusted CDN family; no cookie import, login or access bypass is performed.

Not implemented: b23 short links/legacy AV URLs, speech/OCR/comment extraction, robust Chinese mod-name matching, X adapters or bulk downloading. Missing captions are not evidence of absent mods. Multi-part videos analyze only the selected part and disclose the rest.

Verification: fixture tests cover multi-part identity, access errors, untrusted subtitle URLs/redirects and the full analyze-to-existing-project-extraction path. An opt-in live metadata test (`MMV_BILIBILI_LIVE=1 go test -run TestBilibiliLiveMetadata -v .`) verified https://www.bilibili.com/video/BV1JV4y1y71n/ and its creator-provided sophisticated-backpacks project link on 2026-11-17. That proves public metadata/link extraction only, not complete video recall or a live subtitle track.

Reference inspected: yt-dlp's Bilibili extractor https://github.com/yt-dlp/yt-dlp/blob/master/yt_dlp/extractor/bilibili.py . No upstream code copied. Existing YouTube channel enumeration uses the uploads playlist and contained video IDs; do not replace it with search results, which are not a complete channel inventory.

## Bilibili creator-listing checkpoint

Follow accepts numeric `https://space.bilibili.com/<mid>` profiles. Sync routes to the existing public `/x/space/arc/search` endpoint using the same cookie-free, redirect-refusing client. It requires explicit count/list fields and validates page number/size, declared total, creator identity and duplicate BV identities before a pass succeeds. Verified zero-upload lists advance sync while retaining historical evidence. Full passes are bounded to 1,000 pages, per-entry text limits and a 16 MiB retained-text budget; incremental passes cover at most three newest pages and explicitly report `recent-window-only`, never full coverage. A pagination gap/change, inaccessible source, rate limit or cancellation fails the pass without advancing the last-full-sync marker or dropping the saved archive. Failed partial passes currently restart rather than resume from a cursor.

Coverage is **public upload listing**, not every video part, private/deleted post, captions or completed mod extraction. Discovered entries queue part 1; additional multipart pages remain explicitly unscanned. Cards expose declared/observed counts and limitations. No account import, challenge solving, signing workaround or access bypass is attempted.

Live check 2026-11-17: `/x/space/arc/search?mid=77257482&ps=30&pn=1&order=pubdate` returned code `-799` (rate limit). The request was stopped; no complete live creator inventory is claimed. Fixture coverage proves two-page enumeration, bounded recent windows, malformed/duplicate/changed listings, source identity, cancellation and failed-sync retention. Real permitted full-channel enumeration remains an acceptance gate.
