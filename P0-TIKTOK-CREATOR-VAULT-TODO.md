# P0 — TikTok Creator Vault parity and always-updated mod discovery

**Priority:** FIRST NEXT TASK when the active Minecraft Mod Vault build chat reaches an integration point.

**Target:** Merge this isolated Creator Vault checkpoint into the newest active build without dropping any newer work from that build.

## Acceptance contract

- [x] Preserve the existing YouTube Creator Vault behavior and protected core creators.
- [x] Treat TikTok creators as persistent subscriptions, not one-off imports.
- [x] Accept TikTok profile URLs (including query strings) and `tiktok:@handle` without colliding with an identical YouTube handle.
- [x] Seed the requested `@kizamiringo`, `@its_katsumi`, `@speedychunks`, `@noxusminecraft`, and `@unyxyt` sources plus high-signal official/dedicated Minecraft recommendation sources.
- [x] Seed Katsumi's currently verified Lnk.Bio + Linktree hubs without inventing an unverified modpack destination.
- [x] Automatically discover creator Linktree/Lnk.Bio/Beacons/Carrd-style hubs from profiles, bios, and recent upload descriptions.
- [x] Follow nested public link hubs with a bounded crawl, classify outbound modpacks/mods/packs/socials/support/downloads, and preserve evidence + first/last-seen timestamps.
- [x] Capture outbound hub redirect destinations without fetching arbitrary external sites.
- [x] Refresh creator links independently of video rescans, remove stale auto-discovered links on success, and retain last-known-good links on provider blocking.
- [x] Render a premium Creator Links panel with modpacks first, exact-destination buttons, typed badges, refresh status, and a lightweight Refresh links action.
- [x] First successful sync performs a full-history archive pass.
- [x] Later automatic refreshes are incremental and preserve the previously indexed archive.
- [x] Automatic provider/network failures back off exponentially instead of retry-spamming; manual Sync remains immediate.
- [x] Preserve pause/resume state across restart and preserve archived results after non-core unfollow.
- [x] Analyze speech through public captions first, then the existing local Whisper fallback when available.
- [x] Analyze text-only videos with sampled-frame OCR and merge visual evidence with captions/speech evidence.
- [x] Resolve visual-text mod names through normal Minecraft project providers before promoting them to verified recommendations.
- [x] Keep unresolved names as evidence instead of silently discarding them.
- [x] Dedupe recommendations while retaining source/provenance and video/timestamp evidence.
- [x] UI truthfully describes YouTube + TikTok, full-first/incremental-after refresh, and speech + visual-text analysis.
- [x] Unit/regression tests cover TikTok normalization, cross-platform identity, full→incremental refresh, retry backoff, OCR merge, and text-only mod-name extraction.
- [x] Local real OCR integration test passes with Tesseract.
- [x] Packaged Linux runtime proves a custom TikTok creator can be followed, paused, restarted, and restored from persistent state.
- [ ] On a networked machine, run current yt-dlp against `https://www.tiktok.com/@kizamiringo` and verify full-history enumeration handles the live profile without login/cookie regressions.
- [ ] Run at least one real Kizamiringo text-only recommendation video end-to-end and compare the Vault's extracted list against every visible recommended mod in the video.
- [ ] Verify an actual new TikTok post appears on the next incremental refresh without forcing a historic re-crawl.
- [ ] Exercise the Windows.Media.Ocr backend on the Windows package and confirm overlay text parity with the Linux Tesseract fixture.
- [ ] On a networked runtime, live-refresh Katsumi's current Lnk.Bio/Linktree child destinations, record the exact current modpack target if one is published, and prove a later hub change is picked up automatically.
- [ ] Rebase/port these changes onto the other chat's newest source/build, resolve any newer Creator Vault edits conservatively, then rerun official Go 1.27 tests/build/runtime verification.
- [ ] Only after that integration proof, mark this P0 complete in the canonical release checklist/wiki.

## Required built-in/followed sources in this checkpoint

- YouTube protected core remains: `@AsianHalfSquat`, `@EnderVerseMC`.
- TikTok protected/core: `@kizamiringo`.
- TikTok curated core/data-driven required follows: `@its_katsumi`, `@speedychunks`, `@noxusminecraft`, `@unyxyt`, `@curseforge`, `@hendyvideos`.

## Additional TikTok discovery recommendations seeded

- `@itsknarfy` — Knarfy.
- `@thebreakdownxyz` — The Breakdown.
- `@thecrimsongaming` — The Crimson Gaming.
- `@ygz207` — laveOrc; recommendation-only because it is a smaller active source.

These are discovery sources, not authoritative project metadata. Every extracted mod still goes through project/provider resolution and keeps provenance.

## Implementation map

- `creator_channels.go` — platform-scoped creator identity, TikTok profile enumeration, full-first/incremental-after scheduling, retry backoff, seeded channels/suggestions.
- `creator_ocr.go` — local visual-text pipeline, frame extraction, Tesseract and Windows.Media.Ocr backends, evidence merge.
- `creators.go` — TikTok metadata/caption/Whisper/OCR analysis and visual-text mod extraction.
- `transcription.go` — shared FFmpeg preparation and OCR capability reporting.
- `creator_channels_test.go`, `creator_ocr_test.go`, `creator_profile_links_test.go` — creator/archive/link-hub regressions and real local OCR tests.
- `creator_profile_links.go` — safe public profile/link-hub crawler, redirect unwrapping, typed creator-link metadata, last-known-good cache, and lightweight refresh API.
- `web/index.html`, `web/app.js`, `web/styles.css` — cross-platform Creator Archive UX, premium creator-link panel, and truthful status/copy.
- `CREATOR-LINK-HUB-INTELLIGENCE.md` — P0.2 architecture, safety, Katsumi evidence, and live gates.

## Continuation rule for the next chat

1. Retrieve the verified source checkpoint and patch from the canonical Minecraft Mod Vault Drive folder.
2. Compare them against the **newest active build/source produced by the other chat**; do not replace newer files wholesale.
3. Port/cherry-pick this feature by behavior and tests, preserving all newer unrelated changes.
4. Run the unresolved live-network and Windows OCR gates above first.
5. Run the full project verification matrix and publish the merged release/checkpoint to Drive and GitHub with hashes/receipts.

This file is deliberately a P0 integration handoff rather than a claim that the concurrently developed release has already been merged.

## P0.3 Creator Modpack Library extension — FIRST NEXT integration scope

- [x] Model a creator as having zero/one/many verified modpacks instead of one guessed pack field.
- [x] Discover CurseForge/Modrinth creator profiles from creator-controlled profile/bio/link-hub/upload-description evidence.
- [x] Enumerate all verifiable modpacks from a verified provider profile.
- [x] Distinguish CurseForge owner vs member projects when API evidence supports it; keep public-profile fallback relationship truthful when it does not.
- [x] Distinguish Modrinth owner vs other project-team membership when available.
- [x] Let an explicit creator-controlled `My Modpack`/`Our Modpack`/`Official Modpack` link promote a differently named provider identity only when provider evidence proves exactly one owner/author.
- [x] Refuse ambiguous provider ownership instead of guessing.
- [x] Preserve last-known-good packs on provider failure and retire stale provider entries only after successful authoritative refresh.
- [x] Seed AsianHalfSquat and EnderVerse with their verified CurseForge **creator profiles**, not hardcoded frozen pack lists.
- [x] Refresh creator modpacks independently and inside the normal always-updated creator refresh loop.
- [x] Render every verified pack in a premium Creator Modpacks section with provider, relationship, exact destination and explicit no-verified-pack/partial-cache states.
- [x] Unit/integration tests cover multi-pack enumeration, ownership/member roles, public fallback, mismatched provider identities, no-fabrication and persistence.
- [x] Fresh compiled Linux runtime proved multi-pack enumeration for AsianHalfSquat/EnderVerse provider-profile fixtures and persistence across restart.
- [ ] On a network-enabled packaged runtime, refresh the real current AsianHalfSquat and EnderVerse provider profiles and compare every returned pack to the provider pages.
- [ ] Run a real creator Linktree/Lnk.Bio direct-modpack example end-to-end through the packaged app and verify provider-profile promotion plus discovery of additional packs where evidence supports it.
- [ ] Port P0.3 into the newest concurrent build without replacing newer files, then rerun official Go 1.27 + production-browser verification.

Implementation additions:

- `creator_modpacks.go` — evidence model, provider-profile/direct-pack discovery, CurseForge API + public fallback, Modrinth enumeration, persistence-safe refresh, API endpoint.
- `creator_modpacks_test.go` — multi-provider enumeration, relationship, no-fabrication, mismatched identity, fallback and persistence regressions.
- `creator_channels.go` — AsianHalfSquat/EnderVerse provider-profile seeds and background modpack refresh.
- `creator_profile_links.go` — creator-profile link type and links+packs refresh integration.
- `web/app.js`, `web/styles.css` — premium multi-pack Creator Modpacks UI.
- `CREATOR-MODPACK-LIBRARY.md` — architecture and no-fabrication contract.


## P0.5 hot-drop creator catalogs — 2026-08-24

- [x] Creator catalogs are schema-versioned JSON data consumed by the real Creator Archive from embedded release bundles and `<config>/creator-catalogs/**/*.json`.
- [x] Runtime catalog replacement is digest-gated and reloadable without an application rebuild/restart; malformed bundles surface warnings and cannot erase last-known-good archive data.
- [x] AsianHalfSquat bootstrap records a verified 349-video channel target, 11 exact current video records, and 93 evidence-backed mod/modpack/resource-pack/shader recommendations. Coverage remains explicitly incomplete until the network-enabled full-history sync fills the remaining uploads.
- [x] `@speedychunks`, `@noxusminecraft`, and `@unyxyt` are protected TikTok follows; NoxusMinecraft and UnyxYT are supplied through the data catalog path so adding future creators does not require creator-specific Go code.
- [x] No provider identity, ownership, project URL, or recommendation is fabricated when evidence is absent.
