# Punchy 2.7e — Hot-Path Performance Fix — REPLACEMENT-1

Target: Minecraft 1.20.1 / Forge 47.x / Java 17

## Final install artifact

- `Punchy-2.7e-Hot-Path-Performance-Fix-REPLACEMENT-1.jar`
- SHA-256: `58b2ac2261047a04bc6ebd66e606d6616322f2a5355f2d998bf773666f300698`
- Size: 1,865,607 bytes
- Drive file ID: `1bPzSddLhoMn481tbttbGdcdoZINvJAZO`
- Drive folder: `https://drive.google.com/drive/folders/1QTfuR6GCV16fgduLjqucwJbTVhMOdkIM`
- Install mode: **REPLACEMENT** — remove the stock Punchy JAR and install only this Punchy replacement. A separate compatible TaCZ/Punchy companion may remain installed.

## What changed

- Cache stable Better Combat/player-animation/TaCZ reflective metadata and linkage instead of rediscovering it at client-tick cadence.
- Consolidate Better Combat whole-layer suppression to one semantic map traversal while preserving original stop/pose side-effect counts.
- Replace stable tick-handler `Iterator` traversal with indexed dispatch preserving order.
- Replace per-frame first-person `HashMap` construction with a fixed-capacity reusable map with preallocated views/entries/iterators, then explicitly release it after the final pose-transform use.
- Preserve real animation-state ticks, per-frame rendering cadence, visuals, combat behavior, config behavior, content and public compatibility descriptors.

## Verification

- Strict package integrity: Info-ZIP `unzip -t` PASS, Python `ZipFile.testzip()` PASS, Java/JAR parsing PASS.
- Whole-JAR ASM BasicVerifier: **448 classes / 5,371 methods / 0 failures**.
- Helper behavior harness: **38 checks PASS**.
- Hot-path bytecode audit: **27 checks PASS**.
- Warmed reusable render-map allocation probe: **1,000,000 cycles / 0 allocated bytes** for the tested clear + nine puts + view traversals + lookup path.
- Stock -> unmarked: add exactly 28 neutral helper classes, remove 0, change exactly `PunchyArmRenderer`, `BetterCombatCompat`, `TaczBlacklistCompat`, `ForgeClientPlatform`; metadata-only changes 0; 515 original contents unchanged.
- `javap -p -s` descriptor-surface diff is empty for all four changed upstream classes.
- Unmarked -> final changes exactly `assets/punchy/icon.png`; 546 other contents unchanged; Repair Mark v2 gate PASS.
- Neutral/private-label release scan PASS / 0 hits.

## Durable evidence

- Verification report Drive ID: `12j2fS1o38fZOjUMQdzjMqcmv6H5VTbax`
- Source/evidence ZIP Drive ID: `1G8G9s_d0t7U0dZ8-EGC2gDldOqSrFp8A`
- Repair record Drive ID: `1A8thcmjJbCMVBsZahRAyGiGpJsX3ToQH`
- Repair Mark master Drive ID: `17Bgs9eBKDlo2JDVQPwVqo0pqYzTJIenb`
- Repair Mark 48px Drive ID: `1noP4s8y4f-DmLO4OOMB2vHRexHVq_vkP`
- Repair Brain record: `mc-1.20.1-forge-punchy-2.7e-hot-path-cache-allocation-replacement-1-2026-09-15`
- Repair Brain SHA-256 after append: `cf3e846767c528dd3ab0e1f538eba45b9befca62b1030a39810c2efcfac54ba4`
- Repair Brain: 149 records; previous 745,942-byte history preserved exactly as prefix.

## Runtime gate

Native full-instance runtime/Spark proof is still pending. Run `/sparkc profiler start --timeout 90 --thread *` under an equivalent workload after exercising Punchy first-person animations, Better Combat, TaCZ integration, resource reload, and Save/Quit/reload. No measured FPS/frame-time claim is made until that passes.
