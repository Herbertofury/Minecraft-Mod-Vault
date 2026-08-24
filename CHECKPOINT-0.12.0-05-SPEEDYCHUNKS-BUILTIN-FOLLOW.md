# Minecraft Mod Vault 0.12.0 P0.4 — SpeedyChunks built-in follow verification

Recorded: 2026-08-23

## Requested change

Add `https://www.tiktok.com/@speedychunks?lang=en` as another built-in Creator Vault follow while preserving the existing TikTok archive, creator link-hub intelligence, and zero/one/many creator modpack library behavior.

## Implemented

- Added `@speedychunks` as a required `curated-core` TikTok creator.
- Canonical profile: `https://www.tiktok.com/@speedychunks`.
- The supplied `?lang=en` URL is regression-tested to normalize to the canonical profile.
- Fresh-install and upgrade migration expectations move from 17 to **18** protected creators.
- SpeedyChunks is included in the recommendation catalogue as a current short-form discovery source.
- Existing full-history-first, incremental-after, captions/Whisper + visual OCR, link-hub discovery, and creator modpack discovery behavior is reused unchanged.
- No Linktree, CurseForge, Modrinth, or modpack destination is hardcoded because current public search did not produce a trustworthy identity. If the live profile or archived descriptions expose one later, the existing evidence-backed discovery pipeline can attach it automatically.

## Verification

Implementation commit: `b7ee57a4698ea3068505d3a974bdb4d4059c4e0c`

- Targeted creator normalization/default/migration tests: PASS.
- Full `go test ./... -count=1`: PASS under the documented temporary host-Go compatibility metadata procedure.
- `go vet ./...`: PASS under the same procedure.
- `node --check web/app.js`: PASS.
- `node --check web/catalog.js`: PASS.
- `git diff --check`: PASS.
- Canonical source restored to `go 1.27.0` after verification.
- Fresh Linux runtime: `/api/creators/channels` returned HTTP 200, exactly **18** creators, exactly one `@speedychunks`, `required=true`, `source=curated-core`, `paused=false`, canonical URL correct.
- Full process restart against the same config: again exactly **18** creators and exactly one required SpeedyChunks entry.

## Preservation / merge rule

This is an isolated merge-safe checkpoint. Port/cherry-pick the delta into the newest concurrent Minecraft Mod Vault source; do not replace newer files wholesale.
