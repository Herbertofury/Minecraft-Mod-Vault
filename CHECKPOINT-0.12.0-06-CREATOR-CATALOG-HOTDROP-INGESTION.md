# Minecraft Mod Vault 0.12.0 — P0.5 Creator Catalog Hot-Drop Ingestion

Date: 2026-08-24
Status: merge-safe integration checkpoint

## Shipped behavior

- Generic schema-v1 creator catalogs load from embedded `catalogs/creators/*.json` and runtime `<config>/creator-catalogs/**/*.json`.
- Runtime reload is digest-gated; local newer/equal catalogs can supersede embedded data without a rebuild; `*.disabled.json`, symlinks, oversized and malformed bundles are handled safely.
- Catalog data merges into the existing Creator Vault channels/videos/recommendations persistence, never a separate demo database. Stronger live/provider identities, URLs, project IDs and confidence are not downgraded by weaker catalog evidence.
- Creator Archive exposes a real **Reload catalogs** action backed by `POST /api/creators/catalogs/reload`; state is inspectable through `GET /api/creators/catalogs`.
- AsianHalfSquat bootstrap: expected channel size 349, 11 exact seeded videos, 93 evidence-backed recommendations across mods/modpacks/resource packs/shaders, `coverage.complete=false`.
- Protected TikTok follows include SpeedyChunks, NoxusMinecraft and UnyxYT. NoxusMinecraft and UnyxYT are data-driven catalog follows, proving normal creator additions do not need creator-specific Go code.
- No provider URL, project ownership, Linktree, modpack or recommendation is invented when evidence is absent.

## Current-run verification

- Targeted catalog regression tests: PASS.
- Full `go test -mod=vendor ./... -count=1`: PASS under the documented local Go 1.23 compatibility metadata shim; canonical source restored to `go 1.27.0`.
- `go vet`: PASS earlier in this P0.5 implementation run under the same shim.
- JavaScript syntax checks and all four shipped catalog JSON parses: PASS.
- Fresh Linux runtime: 20 creator channels; exactly one required/active SpeedyChunks, NoxusMinecraft and UnyxYT; AsianHalfSquat reports 349 total, 11 indexed, 93 recommendations.
- Runtime catalog state: 4 shipped catalogs, 11 videos, 93 recommendations, expectedVideos 349.
- Hot-drop test: a fifth runtime catalog added a new creator/video/recommendation without application restart. Replacing that file with malformed JSON returned HTTP 207 with an error and retained last-known-good data.
- Full process restart with the same config retained the hot-dropped record plus AsianHalfSquat 349/11/93 and the protected TikTok follows.

## Explicit boundary

Container DNS cannot resolve YouTube/TikTok, so full live enumeration of AsianHalfSquat's remaining channel history cannot be truthfully claimed here. P0.5 removes the architectural blocker: a network-enabled full-history sync now extends the same hot-drop-backed Creator Vault state rather than requiring hardcoded code/data migrations.

## Merge rule

Port this delta into the newest concurrent source; do not replace newer files wholesale.
