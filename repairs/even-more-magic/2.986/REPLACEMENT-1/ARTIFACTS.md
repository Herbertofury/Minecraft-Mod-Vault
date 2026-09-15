# Even More Magic 2.986 — Hot-Path Performance Fix — REPLACEMENT-1

Target: Minecraft 1.20.1 / Forge 47.x / Java 17

## Final install artifact

- `Even-More-Magic-2.986-Hot-Path-Performance-Fix-REPLACEMENT-1.jar`
- SHA-256: `835a6324292338245eda356dd8ca52eed046901b21a851fcb345f87ac624480d`
- Size: 78,274,539 bytes
- Drive file ID: `1TquQWXgScCtsyPM3ARdq2ry2R7UHeqMG`
- Drive folder: `https://drive.google.com/drive/folders/1ypmXGWbQBsGJtTQbcHlqws73O2rtFJtb`
- Install mode: **REPLACEMENT** — remove stock Even More Magic 2.986 and install only this JAR.

## Architecture changes

- Eliminates all 3,588 exact eager `LazyOptional.orElse(new PlayerVariables())` fallback allocations across 946 generated classes in the normal present-capability path.
- Coalesces bursty full player-variable synchronization with a one-shot server task and per-player `syncQueued` fast path while preserving the original public `syncPlayerVariables(Entity)` descriptor and packet/NBT payload.
- Reduces `PlayerVariables` constructor writes from 173 to the 20 genuinely non-default values.
- Maintains cooldown seconds/type directly at 292 source cooldown writes and removes the derived cooldown PlayerTick subscription.
- Consolidates 22 global menu PlayerTick subscribers into one exact open-menu dispatcher; total mod PlayerTick subscribers drop from 25 to 3 while genuine animated-item and mutable Curios/equipment reconciliation remain tick-driven.
- Repairs seven one-frame animation metadata files, removing 151 invalid frame references without inventing assets.

## Verification

- Strict Info-ZIP `unzip -t`: PASS.
- Whole-JAR ASM BasicVerifier: **8,204 classes / 38,098 methods / 0 failures**.
- Existing API descriptor/hierarchy surface preserved across all 8,202 stock classes; only internal `syncQueued` and `syncPlayerVariablesNow` are added.
- Hot-path audit: eager fallback sites 0; resolver calls 3,588; cooldown source setters 292; total PlayerTick subscribers 3; constructor writes 20; queue TickEvent refs 0.
- Queue harness: 12 checks PASS, including 1,000-request burst coalescing.
- Capability/cooldown harness: 28 checks PASS and zero fallback allocations for present capability.
- Menu dispatcher equivalence: stock 22 subscribers -> one dispatcher with exactly the same 22 procedure targets.
- `.mcmeta` scan: 273 files, 0 invalid frame indices.
- Stock -> unmarked: +2 neutral helper classes, 0 removed, 1,229 original entries changed, 0 metadata-only drift, 16,630 original contents unchanged.
- Unmarked -> final changes exactly `logo.png`; 17,860 other contents unchanged.
- Repair Mark v2 gate PASS.
- Neutral/private-label scan PASS.

## Durable evidence

- Verification report Drive ID: `1Y2VV3dATM5O1W2oatlfy32ZHO6diS3lD`
- Source/evidence ZIP Drive ID: `1bsFEla4-FStCgH8nmsFOjYPKOI6l33mn`
- Repair record Drive ID: `1YkVNFM4_466p6fjY6OXfJdMBqM-Z-D4Z`
- Repair Mark master Drive ID: `1gto6gXePkwuP98LD8SSffYPI2c-UygNG`
- Repair Mark 48px Drive ID: `102g0kDEelx7kq4nIPOaNANQP7m9u77el`
- All six Drive deliverables were redownloaded and SHA-256 verified; remote JAR/source ZIP also passed strict ZIP integrity.
- Repair Brain record: `mc-1.20.1-forge-even-more-magic-2.986-hot-path-player-variable-sync-replacement-1-2026-09-15`
- Repair Brain final SHA-256: `ea5d2c99e631e18cb49a3c15020e3458f0280d46b2fcc1536bb688afc25cea12`
- Repair Brain now has 150 records; previous 754,902-byte history was preserved byte-for-byte as prefix.

## Remaining gate

Exact full-instance runtime/Spark validation is pending. Exercise representative spells/wands, cooldown transitions, quest GUIs/Ring Synthesizer, Curios/equipment including in-place NBT changes, animated items, F3+T, Save/Quit/reload, then inspect `latest.log` and run `/sparkc profiler start --timeout 90 --thread *` under an equivalent workload. No measured FPS/tick-time claim is made until this passes.
