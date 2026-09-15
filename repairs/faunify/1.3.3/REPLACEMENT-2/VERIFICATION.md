# Faunify 1.3.3 - Entity Random and Silk Moth Light Search Fix - REPLACEMENT-2

Verified: 2026-09-14 (America/Denver)
Target: Minecraft 1.20.1 / Forge 47.x / Java 17
Install mode: **REPLACEMENT**
Internal identity: `faunify` / `1.3.3` / `Faunify`

## Result

This repair keeps the previously accepted C2ME-safe Beefly entity-random fix and adds one targeted Silk Moth AI performance repair. It does **not** reduce Silk Moth search radius, light threshold, spawn/content, goal timing, attraction rules, or visual/gameplay behavior.

Final artifact:

- filename: `faunify-forge-1.20.1-1.3.3-Entity-Random-and-Silk-Moth-Light-Search-Fix-REPLACEMENT-2.jar`
- SHA-256: `653961530090b0067e8aaf2f7a746d2bd3a0230ec9a03b76c29bcb99ab2ea6af`
- size: 5,508,247 bytes
- ZIP entries: 1,270
- class files: 275
- class major: 61 / Java 17 for all classes
- ZIP integrity: PASS
- archive signatures: none

Provided repaired baseline:

- SHA-256: `ce816046f34b61f13c57e88b28d40f0197ec3889f46e4c628064c7d4e03ac947`
- size: 5,488,600 bytes

## Performance root cause

`SilkMothEntity$LightAttractionGoal.findHighestPriorityLightSource()` searches a 31 x 31 x 31 cube: **29,791 block positions** per full scan.

The previous implementation used `BlockPos.betweenClosedStream(...)`, converted each yielded mutable position to a new immutable `BlockPos`, filtered through lambdas, built comparator chains, and used `Stream.min(...)`. That creates unnecessary stream/lambda/comparator work and can allocate an immutable position for essentially every scanned coordinate.

The profiler identified this method as a high-cost path even with a single Silk Moth loaded.

## Repair

Only `findHighestPriorityLightSource()` was rewritten. The replacement directly iterates Minecraft's existing `BlockPos.betweenClosed(...)` iterable and tracks the best candidate in primitive locals.

The repair deliberately preserves:

- exact +/-15 X/Y/Z search bounds;
- exact `SILKMOTH_BLACKLIST` rejection;
- exact block-state emission threshold of 8;
- exact priority of highest `LightLayer.BLOCK` brightness;
- exact nearest squared-distance tie-break;
- exact first-encounter behavior when brightness and distance are fully tied;
- `null` when no valid candidate exists;
- the original block-position encounter order;
- all goal timing, movement, orbit, pathing, and random behavior.

Only a candidate that actually becomes the current best is copied with `BlockPos.immutable()`. There is no cache, stale state, extra thread, delayed refresh, scheduler, or off-thread world access.

## Previous Beefly safety repair retained exactly

The prior C2ME-safe Beefly constructor repair remains byte-for-byte unchanged.

`BeeflyEntity$BeeflyGoToKnownFlowerGoal.class` in the exact final JAR:

- SHA-256: `cd45ae4699aa7c2e0a76a4404f51173ca745eadb5c318db24cbcef9efbb4dad2`

This is the same accepted class hash from the previous replacement and keeps the entity-owned random source rather than reading the level's shared random source during the repaired draw.

## Exact bytecode and API verification

Changed performance class:

- original SHA-256: `4d7de674270249c7fd909d5a64fb08bb910aa883bbb73972c362c952d99ab400`
- final SHA-256: `72b82f5cb2678f7ce97725e6c803e56f8606a0746dead69563ddfaa3fbcda6d2`
- `javap -p -s` signature diff: **empty / PASS**
- `StackMapTable`: present
- exact final class ASM `BasicVerifier`: PASS

Whole final JAR ASM verification:

- classes verified: **275**
- methods verified: **2,401**
- failures: **0**

The rewritten method contains direct `BlockPos.betweenClosed(...)` iteration and the original blacklist, emission, block-light, squared-distance, and immutable-best-candidate operations. Its method body no longer invokes the Stream pipeline.

The old synthetic lambda helper methods remain in the class only to preserve the class/member surface; they are no longer called by the repaired search method.

## Behavioral equivalence test

A deterministic selection harness compared the original comparator semantics with the imperative replacement across **10,000 randomized cases**, including:

- blacklisted candidates;
- emission values around the threshold;
- all block-light levels;
- nearest-distance ties;
- exact fully tied candidates;
- empty/no-valid-candidate cases.

Result: **PASS - all 10,000 cases selected the same candidate.**

## Package diff

Provided baseline -> unmarked repaired JAR using the Minecraft Repair `jar_diff.py` contract:

- same entry set: PASS
- added: 0
- removed: 0
- metadata-only changes: 0
- content changes: exactly 1
  - `com/pepper/faunify/entity/SilkMothEntity$LightAttractionGoal.class`
- unchanged content entries: 1,269

Unmarked repair -> final marked JAR:

- same entry set: PASS
- added: 0
- removed: 0
- metadata-only changes: 0
- content changes: exactly 1
  - `icon.png`
- unchanged content entries: 1,269

Provided baseline -> final:

- content changes: exactly the Silk Moth goal class and `icon.png`
- all other 1,268 entry contents unchanged
- metadata-only changes: 0

## Compatibility identity

The exact final JAR preserves:

- `modId="faunify"`
- version `1.3.3`
- display name `Faunify`
- `logoFile="icon.png"`
- loader `javafml`
- Forge dependency `[47,)`
- Minecraft dependency `[1.20.1,1.21)`
- GeckoLib dependency `[4.4,)`
- manifest / mixin identity
- registry/resource IDs
- entities, spawning, loot, recipes, blocks, items, configuration, networking, and world data

Critical unchanged metadata hashes:

- `META-INF/mods.toml`: `4d12de467dae923bdee8ca88180273f29388607435d1f07b688c8ebf94ea5763`
- `META-INF/MANIFEST.MF`: `531952c3cc1de7006b90c8cd51e68282d001248e01b4599a8324a6743a88bf15`

## Repair Mark v2

Current project identity was rechecked against CurseForge: project `Faunify`, project ID `1123041`, author `Peppercorn`, Minecraft 1.20.1, Forge/Fabric, with Forge `faunify-forge-1.20.1-1.3.3.jar` listed as the current 1.20.1 release dated 2026-07-24.

The current storefront image bytes could not be materialized in the repair environment. Repair Mark policy explicitly permits exact upstream-authored art from the exact official release in this case. Therefore the mark is derived from the exact 1.3.3 embedded `icon.png`, not guessed or synthesized artwork.

- official-art basis: exact upstream-authored `icon.png` from Faunify 1.3.3
- source dimensions: 300 x 55 RGBA
- source SHA-256: `b009064c78ae9a31ea7ee7f7896cfb34d177558c0f5b25a2b9f4285690b6a4a8`
- marked SHA-256: `115c7c4ef5a039c6337a1ef9a419e9ff2fe1cc112e1a5afcd3106820a02da91f`
- 48x48 QA SHA-256: `4858b67c82704b98bd42715eadf7e2deb7b490f8be69b9a1288466ba098eb774`
- treatment: deterministic Repair Mark v2 red double frame, corner brackets, and verified check; zero rendered text
- integration: embedded in the already-declared `icon.png`
- marker-only diff: PASS - exactly `icon.png`
- image generation: not used

## Neutral-identity scan

Release filename, archive entry names, raw archive bytes, and decompressed archive contents were scanned case-insensitively for private/user/pack/project labels from the repair environment.

Result: **PASS - no prohibited private labels found.**

## Runtime status / remaining gate

This exact final SHA is **strongly static/package/bytecode verified but not native-runtime verified** because the repair environment does not contain the full Forge 1.20.1 + GeckoLib target runtime or the user's full instance.

Do not treat that limitation as a known defect. The remaining acceptance gate is:

1. replace the previous Faunify JAR with only this REPLACEMENT-2 JAR;
2. launch the same Forge 1.20.1 instance;
3. exercise Silk Moths around qualifying lights, including adding/removing light sources;
4. verify normal approach/orbit behavior and no missing attraction updates;
5. generate/explore chunks that can instantiate Beeflies and confirm the previous C2ME random failure remains absent;
6. Save & Quit, reload, and repeat the behavior check;
7. run the same 90-second Spark profile and compare `LightAttractionGoal#findHighestPriorityLightSource` against the prior profile.

If the runtime gate passes, this replacement can be promoted from static-verified to runtime-verified without any further code change.
