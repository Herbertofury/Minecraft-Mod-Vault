# Video-to-mod discovery and download pillar

Owner requirement: pasted videos, Shorts and entire creator video archives should reveal every mentioned mod and let the user obtain the identified projects without manually hunting each one. This is a shared behavior contract for modding tools and databases, not a replacement of existing applications.

## Existing implementation to build on

Minecraft-Mod-Vault's Creator Vault already has YouTube/TikTok channel enumeration, captions/local speech transcription, sampled-frame OCR, provider resolution, unresolved mentions, persistent creator catalogs and a recommendation browser. See `creators.go`, `creator_channels.go`, `creator_ocr.go`, `creator_catalogs.go`, issue #5 and release gate #40. Tests/fixtures do not establish complete live-source coverage.

This source-identity checkpoint integrates directly into that extractor: reject lookalike provider hosts, retain original source links/timestamps, identify Modrinth/CurseForge projects rather than version/file IDs, and deduplicate by project identity instead of display name. It does not add X/Bilibili ingestion or bulk downloading.

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
