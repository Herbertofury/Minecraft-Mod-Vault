# Even More Magic 2.986 — Hot-Path Performance Fix — REPLACEMENT-1 Verification

Target: Minecraft 1.20.1 / Forge 47.x / Java 17  
Install mode: **REPLACEMENT**  
Internal identity: `even_more_magic` / `2.986` / `Even More Magic`

## Final artifact

- `Even-More-Magic-2.986-Hot-Path-Performance-Fix-REPLACEMENT-1.jar`
- SHA-256: `835a6324292338245eda356dd8ca52eed046901b21a851fcb345f87ac624480d`
- Size: 78,274,539 bytes
- Entries: 17,861
- Classes: 8,204
- Strict Info-ZIP `unzip -t`: PASS
- Python ZIP validation: PASS
- Archive signatures: none

This exact SHA is strongly static/package/bytecode/hot-path-behavior verified, but exact target runtime/Spark validation is still pending.

## Root causes and repair

Stock generated code contained **3,588** exact eager `LazyOptional.orElse(new PlayerVariables())` allocations across **946 classes**. REPLACEMENT-1 routes those reads through `PlayerVariableAccess.resolve`, so a normal present capability does not construct the fallback object. The generated `PlayerVariables` constructor was also reduced from 173 explicit writes to the 20 genuinely non-default initializers.

Full player-variable synchronization is coalesced through `PlayerVariableSyncQueue`: the existing public `syncPlayerVariables(Entity)` descriptor is preserved, but repeated same-burst requests produce one queued server task / latest state per player. The original packet/NBT send path remains internally as `syncPlayerVariablesNow(Entity)`. The queue adds no tick subscriber, timer, polling thread, executor, or off-thread world access.

All **292** source writes to the cooldown value now maintain derived seconds/type immediately, allowing the derived cooldown PlayerTick subscriber to be removed. The attribute/Curios reconciliation tick remains because relevant item NBT can change in-place while the same stack stays equipped; only its private scheduler-value network sync was removed.

Stock contained **22** globally subscribed menu PlayerTick wrappers. They are consolidated into one existing menu dispatcher while preserving exactly the same 22 procedure targets and once-per-END-tick-open behavior. Total Even More Magic PlayerTick subscribers therefore drop from **25 to 3**: genuine animated-item state, genuine mutable equipment/Curios reconciliation, and the consolidated menu dispatcher.

Seven one-frame `.png.mcmeta` files contained **151** invalid frame references. They now reference only physical frame 0 while preserving their timing/interpolation settings; final scan reports 273 metadata files, 0 issues, 0 invalid indices.

## Verification

- Whole-JAR ASM BasicVerifier: **8,204 classes / 38,098 methods / 0 failures**.
- API audit across all **8,202 stock classes** preserves every existing class hierarchy, field descriptor, and method descriptor; only internal `syncQueued:Z` and `syncPlayerVariablesNow(Entity):void` are added.
- Hot-path audit: `eager_fallback_sites=0`, `resolve_calls=3588`, `cooldown_source_setters=292`, `player_tick_subscribers=3`, `constructor_putfields=20`, `queue_tick_event_refs=0`.
- Sync queue harness: **12 checks PASS**, including 1,000-request burst coalescing, latest-state-wins and removed-player handling.
- Capability/cooldown helper harness: **28 checks PASS**, including zero fallback construction for a present capability.
- Menu dispatcher audit: **22 stock menu subscribers -> 1 final dispatcher**, exact 22 procedure targets preserved.
- Stock -> unmarked: add exactly 2 neutral helper classes, remove 0, change 1,229 original entries, metadata-only changes 0, 16,630 original contents unchanged, intersecting ZIP metadata drift 0.
- Unmarked -> final changes exactly `logo.png`; 17,860 other contents unchanged.
- Loader metadata/manifest are byte-identical to stock.
- Repair Mark v2 gate PASS.
- Neutral/private-label scan PASS.

## Persistence

Drive folder: `https://drive.google.com/drive/folders/1ypmXGWbQBsGJtTQbcHlqws73O2rtFJtb`

- JAR: `1TquQWXgScCtsyPM3ARdq2ry2R7UHeqMG`
- Verification: `1Y2VV3dATM5O1W2oatlfy32ZHO6diS3lD`
- Source/evidence: `1bsFEla4-FStCgH8nmsFOjYPKOI6l33mn`
- Repair record: `1YkVNFM4_466p6fjY6OXfJdMBqM-Z-D4Z`
- Repair Mark: `1gto6gXePkwuP98LD8SSffYPI2c-UygNG`
- Repair Mark 48px: `102g0kDEelx7kq4nIPOaNANQP7m9u77el`

All six Drive files were downloaded after publication and SHA-256 verified; the remote JAR and source bundle also passed strict ZIP integrity.

Repair Brain after append: 150 records, SHA-256 `ea5d2c99e631e18cb49a3c15020e3458f0280d46b2fcc1536bb688afc25cea12`; the previous 754,902-byte history is preserved byte-for-byte as its prefix.

## Runtime gate

Install only the replacement JAR, exercise representative spells/wands, cooldown transitions, quest GUIs/Ring Synthesizer, Curios/equipment including in-place NBT changes, animated items, `F3+T`, and Save/Quit/reload. Inspect a fresh `latest.log`, then run `/sparkc profiler start --timeout 90 --thread *` under an equivalent workload. No measured FPS/tick-time claim is made until this passes.
