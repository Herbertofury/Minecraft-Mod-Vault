# Minecraft Mod Vault 0.6.0 Release Notes

## Universal Mods browser

- Renamed the old “Amazing Mods” workspace to **Mods**.
- Replaced source-logo external-link shortcuts with provider toggles inside one browser.
- Unified Modrinth, CurseForge, GitHub, Planet Minecraft, MCPEDL and Marketplace results.
- Added 40+ first-class Vault categories plus live Modrinth taxonomy discovery and live CurseForge public category discovery, deduplicated into one browse graph. A configured CurseForge API key upgrades that lane to official category IDs.
- Added project galleries, project icons, author avatars, source variants and an integrated detail panel.
- Preserved rich media through search normalization and cross-provider deduplication.

## Living intelligence

- Added persistent live recommendations with automatic refresh.
- Added rotating popular, recently updated, preference and category feeds so recommendation discovery changes over time instead of cycling a fixed static set.
- Added creator-recommendation signals to live ranking.
- Added configurable refresh intervals and interests; defaults now refresh recommendations every 15 minutes and creator discovery every 60 minutes while the app is running. The creator analysis queue is drained every minute.
- Removed frozen curated mod cards from successful Home/Mods discovery. Primary browsing is live; preserved Vault seeds appear only as a clearly labeled outage fallback.

## Ultimate updater

- Added exact installed-JAR identity using SHA-1, SHA-512, CurseForge Murmur2 and embedded mod metadata.
- Added target Minecraft version/loader migration planning.
- Added exact Modrinth update matching and CurseForge fingerprint matching.
- Added dependency-aware Modrinth update staging.
- Added live port/fork/continuation searches for version gaps, with ambiguous replacements kept review-only.
- Added download hash verification, JAR readability checks, identity checks, backups, atomic replacement and rollback.

## Creator Picks

- Added YouTube and TikTok discovery.
- Added description link/timestamp extraction.
- Added public caption extraction with real timestamps.
- Added a continuously drained background analysis queue so discovery-only videos are not shown as complete lists.
- Added per-mod direct links, timestamp deep-links, confidence/evidence, project icons and authors.
- Removed the arbitrary small transcript-result cap: long recommendation videos retain every distinct plausible mod mention until analysis completes.
- Added timestamp inheritance for creator descriptions and a Needs review lane so unresolved spoken/written mod names are retained instead of silently dropped.
- Added on-demand local speech-to-text for captionless videos using current yt-dlp/FFmpeg/whisper.cpp release assets and a verified multilingual Whisper model.
- Added transcript-engine status and a one-click preparation control.

## Correctness and regressions

- Provider cards no longer claim a native install route when the source only supports in-app discovery/details.
- Source websites remain available as secondary detail actions instead of replacing the integrated browser.
- Existing CIT, furniture, Bedrock, Manager, Import and local file-management features are preserved.
- Legacy `#amazing` and `#browser` routes redirect to `#mods` for old bookmarks without reviving the old UI.
