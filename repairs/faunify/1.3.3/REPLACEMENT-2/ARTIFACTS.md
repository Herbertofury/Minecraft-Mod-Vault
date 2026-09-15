# Faunify 1.3.3 - REPLACEMENT-2

Target: Minecraft 1.20.1 / Forge 47.x / Java 17

This checkpoint records the safety-hardened Faunify performance replacement that preserves the previously accepted Beefly entity-random repair and optimizes Silk Moth light-source search without changing its selection behavior.

## Final install artifact

- `faunify-forge-1.20.1-1.3.3-Entity-Random-and-Silk-Moth-Light-Search-Fix-REPLACEMENT-2.jar`
- SHA-256: `653961530090b0067e8aaf2f7a746d2bd3a0230ec9a03b76c29bcb99ab2ea6af`
- Size: 5,508,247 bytes
- Install mode: **REPLACEMENT**
- Drive file ID: `1Yt19-VzKtvQIF7NbOzeCj9J6Ow1bCF6P`
- Drive folder ID: `1Qwt0E9AkqGsdS_iyD4y3t0BFchYZkzhl`

## Exact static verification

- ZIP integrity: PASS
- Whole-JAR ASM BasicVerifier: 275 classes / 2,401 methods / 0 failures
- Target class signature surface unchanged
- 10,000 randomized selection/tie/threshold/blacklist equivalence cases: PASS
- Baseline -> unmarked repair changes exactly `com/pepper/faunify/entity/SilkMothEntity$LightAttractionGoal.class`
- Unmarked -> final changes exactly `icon.png`
- Added entries: 0
- Removed entries: 0
- Metadata-only changes: 0
- Prior Beefly repair retained exactly, class SHA-256 `cd45ae4699aa7c2e0a76a4404f51173ca745eadb5c318db24cbcef9efbb4dad2`
- Repair Mark v2 gate: PASS
- Private-label scan: PASS

## Behavior preserved

The optimized Silk Moth search keeps the exact +/-15 search cube, blacklist, emission threshold 8, highest block-light priority, nearest squared-distance tie-break, encounter-order behavior for exact ties, and null result semantics. It adds no cache, delayed refresh, thread, scheduler, off-thread world access, content reduction, or radius reduction.

## Persistence

- Verification report Drive ID: `1uw6rxVvsy3xm6mbmLq3rlhWOtU5_Ssqs`
- Repair record Drive ID: `1fkXdzq09QZhwxRAhylgnry3yYPn3H6kQ`
- Source/evidence Drive ID: `1I9xHl9ORxriUS22n4Ll6DNdkV0ukDcBQ`
- Repair Mark Drive ID: `1nKZU4NofjaCtHFXKyxR24WtSTFOYK-ap`
- 48px QA mark Drive ID: `1XSkuzdviqVx7Nvi3Qh-9hUDzY0TYSF7i`
- All six Drive deliverables were redownloaded and SHA-256 verified after upload.
- Canonical Repair Brain record ID: `mc-1.20.1-forge-faunify-1.3.3-silk-moth-light-search-performance-replacement-2-2026-09-14`

## Runtime gate

This exact JAR is not yet called runtime-verified. The full target instance was unavailable in the repair environment. Install only this Faunify replacement, exercise Silk Moths around qualifying lights plus Beefly chunk-generation paths, Save & Quit/reload, then run the same 90-second Spark profile.
