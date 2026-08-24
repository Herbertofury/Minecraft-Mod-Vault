# Minecraft Mod Vault 0.12.0 P0.1 - Katsumi built-in follow checkpoint

Recorded: 2026-08-23

## Requested delta

Add `https://www.tiktok.com/@its_katsumi` as another built-in Creator Vault TikTok follow with the same always-updated archive behavior as the existing YouTube/TikTok subscriptions.

## Implemented

- Added `@its_katsumi` / **Katsumi** to `defaultCreatorChannels` as `Platform: tiktok`, `Required: true`, `Source: curated-core`.
- Added Katsumi to the Creator Vault recommendation catalogue so UI/catalog state stays coherent with the built-in watchlist.
- Fresh-install and upgrade migration expectations now cover **17** protected creators.
- Added normalization coverage for the exact profile URL `https://www.tiktok.com/@its_katsumi`.
- Updated Creator Archive UI copy, README, P0 TODO, creator source catalogue, STATUS, and HANDOFF continuity state.
- The underlying full-history-first / incremental-refresh-after / archive-preserving TikTok behavior is reused unchanged; Katsumi is not a separate one-off importer.

## Verification performed in this checkpoint

- Targeted creator seed/migration/URL-normalization tests: **PASS**.
- Full `go test ./...`: **PASS** under the same documented host-Go metadata compatibility procedure used by the parent P0 checkpoint.
- `go vet ./...`: **PASS** under that procedure.
- `node --check web/app.js`: **PASS**.
- `node --check web/catalog.js`: **PASS**.
- Fresh Linux and Windows x86-64 host-compatibility binaries built.
- Fresh package extraction produced byte-identical binaries.
- Fresh Linux package runtime: `GET /api/creators/channels` returned HTTP 200, exactly **17** creators, exactly one `@its_katsumi`, with `required=true` and `source=curated-core`.
- Restarted against the same clean runtime state: still exactly **17** creators and exactly one Katsumi entry, proving the built-in migration is idempotent.
- Fresh UI HTML contains **Katsumi** in the cross-platform Creator Archive description.

## Host-toolchain note

Canonical source remains `go 1.27.0`. This environment provides Go 1.23.2. For tests/vet/build only, `go.mod` and vendored module `go 1.25.0` metadata were temporarily lowered to 1.23 and restored immediately afterward. The next canonical release still requires official Go 1.27 verification after merging with the other chat's newest build.

## Artifact hashes before source packaging

- Linux package SHA-256: `273f7544d4ff082c091d5a2d6256681c73db2cf0840ab15dda06420981b2beb8`
- Windows package SHA-256: `3e098e6b1b2265ed7ac70d772fecc0d0e2d24aa7018ab4b83807d08f8619d4a5`
- Linux binary SHA-256: `1f31979619f4be0bd5c541baa928290e7ed1eb0456bf3b7f2e49412462956e31`
- Windows binary SHA-256: `84224c337f3bca949b76664f3ebbf573d0ebead076abfacfd3e95bcc5a86f965`

## Truthful remaining gate

This delta verifies Katsumi's **built-in follow and persistence/migration integration**, not a live full-history crawl of her TikTok account. The parent P0 live-TikTok gates remain: networked current-yt-dlp profile/video validation, real text-only completeness comparison, incremental-new-post proof, and Windows.Media.Ocr runtime proof before calling the combined Creator Vault release complete.
