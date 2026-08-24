# Minecraft Mod Vault 0.7.1 release notes

Version 0.7.1 is the ready-to-run Creator Archive release. It preserves the 0.7.0 federated provider browser, updater, installation and creator transcription systems while making the curated Minecraft YouTube corpus active by default.

## Thirteen built-in creator archives

A fresh install now follows and continually archives all of these channels: AsianHalfSquat, EnderVerseMC, Noxus, ChosenArchitect, direwolf20, Gaming On Caffeine, SystemCollapse, Lashmak, PwrDown, Mischief of Mice, PopularMMOs, DanTDM, and The Breakdown.

Each receives the same exhaustive pipeline: complete upload enumeration, Videos, Shorts and streams, descriptions and outbound project links, timed captions, local whisper.cpp fallback, persistent transcripts, mod/project resolution, evidence, confidence, unresolved mentions, newest/oldest recommendation chronology, per-video browsing, filters and resumable retries.

Existing installations migrate the curated channels into the built-in watchlist without replacing indexed history or pause state. Built-in channels can be paused individually; user-added channels retain safe unfollow-with-archive-preservation behavior.

## First-run and backfill QoL

- Channel enumeration now shares the configured 1-4 worker Creator Archive concurrency budget.
- A fresh 13-channel install therefore drains the complete archive steadily instead of spawning every channel crawl simultaneously.
- The background loop revisits the queue every 20 seconds, so newly available slots are filled automatically.
- The Creator Archive UI clearly identifies the curated set as included while keeping Follow Any YouTuber available for arbitrary additional channels.
- The recommendations panel remains as an explanation and tier browser for the built-in sources rather than an installation checklist.

## Compatibility

All 0.7.0 provider lanes, updater verification, detected-package installation, Creator Archive transcripts, custom YouTuber following, pause/resume, safe custom unfollow, search/filter/sort behavior and Windows/Linux packaging remain intact.
