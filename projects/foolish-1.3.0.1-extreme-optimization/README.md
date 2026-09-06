# Foolish 1.3.0.1 Extreme Optimization

Production-safe performance patch for the Minecraft 1.20.1 Foolish mod artifact `foolish-1.3.0 patched.jar`.

## Release identity

- Output: `foolish-1.3.0.1-extreme-optimized.jar`
- SHA-256: `d84f0afe46fba4257495d67af478fac7e33220cb5e71757aa50e309083c62d03`
- Size: 133,243,571 bytes
- Internal mod identity intentionally preserved: `foolish` / `1.3.0`
- Production QA: Forge 47.4.23 / Minecraft 1.20.1 / MCP 20230612.114412 / Temurin Java 17.0.20.1

## What was fixed

The five global `LivingTickEvent` AoE response handlers were generated with repeated spatial entity scans and full nearest-target stream sorts for every living entity tick. The patch keeps the original tick cadence and gameplay branches but replaces that waste with per-execution query reuse, direct player-list spatial checks, and linear nearest selection. Ten additional high-density Foolish entity tick procedures received the same direct-player/linear-nearest hardening.

No five-tick throttle, entity cap, render downgrade, feature disable, content deletion, or quality reduction is used.

## Verification

- 15 changed public APIs match the input artifact.
- All 5 AoE Forge event subscriber annotations remain.
- 0 `Stream.sorted` calls remain in the optimized target procedures.
- ZIP/package audit found 0 unintended added, missing, or changed non-target entries.
- `META-INF/mods.toml`, `META-INF/neoforge.mods.toml`, and `pack.mcmeta` are byte-identical to the input.
- Fresh official packaged Forge 47.4.23 `forgeserver` reached `Done (6.408s)!` with the optimized JAR and GeckoLib 4.8.4.
- Same saved world / same production runtime / 100 persisted ArmorStand living entities, two ~5-second whole-JVM CPU samples per build:
  - original: `2.500s`, `2.220s` — mean `2.360s`
  - optimized: `2.080s`, `1.950s` — mean `2.015s`
  - mean JVM CPU reduction: **14.62%**
  - all captured Forge TPS snapshots: **20.000 TPS**

The Foolish data/dependency warnings seen at startup (missing optional/content references such as `foolish_biomes:emberlight_valley`) are pre-existing and reproduce with the original input JAR.

## Durable artifact

Google Drive project folder: https://drive.google.com/drive/folders/1l-CVYjKm_LBgr7wQh3C_OtrYFhabMRLV

The Drive connector rejected the single 133 MB JAR upload, so the exact binary is stored there losslessly as two split parts plus `FINAL-JAR-SHA256.txt` and `REASSEMBLE-WINDOWS.ps1`. A Drive round-trip download + reassembly was verified to produce the release SHA-256 above.

The complete final verification report and source/evidence ZIP are also stored in that folder. The runnable JAR is delivered directly in the originating ChatGPT artifact handoff.
