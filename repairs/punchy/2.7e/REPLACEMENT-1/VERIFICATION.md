# Punchy 2.7e — Hot-Path Performance Fix — REPLACEMENT-1 Verification

Target: Minecraft 1.20.1 / Forge 47.x / Java 17  
Install mode: **REPLACEMENT**  
Internal identity: `punchy` / `2.7e` / `Punchy`

## Final artifact

- `Punchy-2.7e-Hot-Path-Performance-Fix-REPLACEMENT-1.jar`
- SHA-256: `58b2ac2261047a04bc6ebd66e606d6616322f2a5355f2d998bf773666f300698`
- Size: 1,865,607 bytes
- 547 entries / 448 Java-17 classes
- Strict package integrity: Info-ZIP `unzip -t` PASS, Python `ZipFile.testzip()` PASS, Java/JAR parsing PASS
- Runtime status: strong static/package/bytecode/hot-path-allocation verification; exact target runtime/Spark re-profile still pending

## Root cause and repair

Punchy repeatedly rediscovered stable Better Combat/player-animation/TaCZ reflection metadata at client-tick cadence, scanned the same Better Combat animation-layer map twice, allocated an iterator for a stable tick-handler list each tick, and allocated a fresh nine-part `HashMap` every first-person rendered frame.

`punchy.repair.HotPathCache` now caches stable class/field/method linkage with `MethodHandle`/`VarHandle`, caches identifiers, consolidates Better Combat layer suppression to one semantic map pass while preserving original stop/pose side-effect counts, caches TaCZ gun linkage, dispatches tick handlers by index, and provides a fixed-capacity reusable first-person pose map with preallocated views/entries/iterators. The map is explicitly released after its final transform use so `ModelPart` references are not retained between normal frames.

Actual animation state-machine ticks and per-frame first-person rendering remain intact. No content, visual quality, animation quality, combat behavior, render distance, simulation, target counts, configuration behavior, or compatibility descriptor was reduced.

## Exact package diff

Stock: `punchy-2.7e-forge-1.20.1.jar` — SHA-256 `459e6889c11ecfbd1faf13749dafa8eea8df39e4d569e8b2bb0cb80aae94dcf2`.

Unmarked repaired baseline: SHA-256 `4332d16fc48e4acaa892d0bae23b6ec51910748fd74a543fc176638e6687ffda`, size 1,866,394 bytes.

Stock -> unmarked:

- add exactly 28 neutral `punchy/repair/HotPathCache*.class` helper classes;
- remove 0;
- change exactly `PunchyArmRenderer`, `BetterCombatCompat`, `TaczBlacklistCompat`, and `ForgeClientPlatform`;
- metadata-only changes 0;
- 515 original entry contents unchanged;
- untouched original ZIP-entry metadata preserved byte-identically.

`javap -p -s` descriptor-surface diff is empty for all four changed upstream classes.

## Verification

- ASM `BasicVerifier`: **448 classes / 5,371 methods / 0 failures**.
- Functional helper harness: **38 checks PASS**.
- Hot-path bytecode audit: **27 checks PASS** — no direct Better Combat class-discovery calls in repaired hot paths, no tick iterator, no renderer `HashMap` construction, and no steady fixed-map array/HashMap allocation opcodes.
- Warmed fixed pose-map probe: **1,000,000 cycles / 0 allocated bytes** for clear + nine puts + key/value/entry traversals + lookup.
- Loader metadata, manifest, `punchy.mixins.json`, and `punchy.compat.mixins.json` remain byte-identical to stock.
- Neutral/private-label release scan: PASS / 0 hits.

## Repair Mark v2

Exact upstream declared art: `assets/punchy/icon.png`, 146x146 RGBA, SHA-256 `8533db84f11e30bc00161c5a58e521efefca1f5a10956b9ef0ea8d5da45340ce`.

Marked art SHA-256: `d9e707e00da1f06ac7264deaeb756129dbe8edfc1507096493f232c200ef38f3`; 48px QA SHA-256: `0a3504fdfcc2a1abefbb3ff72480c751a0d6050f237e3383837fe746f8f959b1`.

Unmarked -> final changes exactly `assets/punchy/icon.png`; 546 other contents unchanged; Repair Mark gate PASS; no image generation used.

## Persistence

Drive folder: `https://drive.google.com/drive/folders/1QTfuR6GCV16fgduLjqucwJbTVhMOdkIM`

- JAR ID `1bPzSddLhoMn481tbttbGdcdoZINvJAZO`
- Verification ID `12j2fS1o38fZOjUMQdzjMqcmv6H5VTbax`
- Source/evidence ZIP ID `1G8G9s_d0t7U0dZ8-EGC2gDldOqSrFp8A`
- Repair record ID `1A8thcmjJbCMVBsZahRAyGiGpJsX3ToQH`
- Repair Mark ID `17Bgs9eBKDlo2JDVQPwVqo0pqYzTJIenb`
- Repair Mark 48px ID `1noP4s8y4f-DmLO4OOMB2vHRexHVq_vkP`

All six objects were re-downloaded from Drive and byte-for-byte hash verified. The downloaded remote JAR and source ZIP both pass strict `unzip -t`.

Repair Brain record: `mc-1.20.1-forge-punchy-2.7e-hot-path-cache-allocation-replacement-1-2026-09-15`. Canonical Repair Brain after append: 149 records, SHA-256 `cf3e846767c528dd3ab0e1f538eba45b9befca62b1030a39810c2efcfac54ba4`; the entire previous 745,942-byte history is preserved exactly as its prefix.

## Runtime gate

Install only the replacement as Punchy, keep any separate compatible TaCZ/Punchy companion unless runtime evidence proves a conflict, exercise first-person animations + Better Combat + TaCZ + Hand Editor/Mixpacks + `F3+T` + Save/Quit/reload, inspect fresh `latest.log`, then run `/sparkc profiler start --timeout 90 --thread *` under an equivalent workload. No measured FPS/frame-time claim is made until that passes.
