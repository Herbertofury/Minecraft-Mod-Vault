# Future Feature Bridge - Release Evidence

Date: 2026-08-27
Target: Minecraft Forge 1.20.1 / Forge 47.4.23
Project family: Ragdoll Reactions + Sable Player Ragdoll + Future Feature Bridge companion

## Evidence status

This record deliberately separates bytes recovered in the resumed session from exact-final runtime evidence preserved by the dead-chat checkpoint.

### A. Recovered byte evidence in this resumed session

Canonical Ragdoll/Sable checkpoint recovered from Google Drive:

- File: `ragdoll-sable-forge-full-checkpoint-2026-08-27-full-compile-remap-server-green.zip`
- SHA-256: `9be8ed0148daf90c0bd157f34411a473fe6275a8d18c534c3edeb5d413af6d6a`
- Recovered remap validator result: `classes=250 staleMethods=0 staleFields=0`
- Recovered baseline dedicated-server smoke reached `Done (35.058s)!`, then `Stopping server`, `Saving worlds`, and `All dimensions are saved`.

Baseline JAR hashes recorded inside that checkpoint:

- `ragdoll_reactions-1.20.1-0.7.0.jar`: `40526d60b19395c1855939958a6349d60962861f977edeca1af639f0b2e6f211`
- `sable-forge-1.20.1-2.0.0-all.jar`: `3e1b2a1eae26a351d79b965c523454d43510e5111b60288a86c61431f3cc2c8c`
- `sable_player_ragdoll-1.20.1-0.7.2.jar`: `ebd46b613129f0a36dcf35db43a4522f81ec2e148602616bdf9a88f146a79cbe`

Updated Minecraft Dev Kit skill produced in this resumed session:

- File: `minecraft-dev-kit.skill.zip`
- SHA-256: `d417021ea7538ddc6091717dfee17ebad8bed6c8b9aa69b3f5f0eb6148b66c4c`
- Official skill validator: `Skill is valid!`
- ZIP integrity: `No errors detected in compressed data`

### B. Preserved exact-final Future Feature Bridge runtime evidence

The dead-chat checkpoint identifies the final companion only by the SHA-256 prefix `b20acda1e2d...`. That prefix is preserved exactly. This record does not invent the missing suffix.

The exact final candidate was reported byte-for-byte reproducible across two fresh builds and passed these native/runtime lanes before the chat died:

1. Native Forge client, embedded provider
   - Resolver: `minecraft:wind_charge -> ragdoll_reactions_future [EMBEDDED]`
   - Creative inventory opened successfully on the fixed exact candidate.
   - Oak and iron trapdoors were authoritatively closed before the burst.
   - After impact, oak became `open=true` while iron remained `open=false`.
   - Clean save/quit reached `All dimensions are saved`.
   - Zero bridge-owned linkage or mod-loading failures.

2. `prepare-uninstall` cleanup path
   - Removed 3 loaded fallback projectiles.
   - Removed 1 dropped charge.
   - Removed 4 player/ender-chest charges.
   - Counts were observed in one run on the exact final binary.

3. Native Forge client, real JTS external provider
   - JTS Wind Charge was physically consumed in Survival: 2 -> 1.
   - Resolver: `justtrialspawners [EXTERNAL] / PROJECTILE_IMPACT_OR_DISCARD`.
   - One impact produced exactly one Reactions trigger.
   - The same impact produced exactly one zombie ragdoll.
   - Project-specific regression observation: 7 sublevels / 6 joints.
   - Zero bridge-owned error.

4. Same saved world after removing JTS and its provider-side dependencies
   - Forge emitted normal missing-registry warnings for removed JTS/Architectury/FTB content.
   - World continued safely and player rejoined.
   - Both original ragdoll systems activated.
   - Resolver returned to `ragdoll_reactions_future [EMBEDDED]`.
   - Zero bridge-owned failures.
   - This proves no stale external-provider lock-in was retained across restart.

5. Dedicated Forge server, embedded final companion
   - Earlier final-matrix server gate reached `Done (26.931s)!`.
   - Legacy/tolerated client-class mixin warnings were separated from bridge-owned failures rather than hidden.

6. Dedicated Forge server with exact final companion + JTS + Architectury + FTB Library
   - Reached `Done (7.841s)!`.
   - Both original ragdoll systems activated.
   - Clean shutdown: `Stopping server -> Saving worlds -> All dimensions are saved`.
   - Zero bridge-owned failures.

### C. Release blocker caught only by native production-client QA

The previous candidate crashed when Creative inventory opened because code linked to `BuildCreativeModeTabContentsEvent.accept(ItemLike)` through a mapped-development convenience overload whose production Forge 47.4.23 owner/name/descriptor did not exist as compiled.

Important result: mapped-dev compilation and ordinary remap validation both passed before the real production client exposed the `NoSuchMethodError`.

The fix added an exact production symbolic-linkage validator that can reproduce this class of failure offline before Minecraft launches. The fixed exact candidate then opened Creative inventory successfully in native Forge.

This is now a permanent Minecraft Dev Kit false-pass protection:

- mapped compilation is not production-linkage proof;
- ordinary remap success is not production-linkage proof;
- resolve and validate exact production owner/name/descriptor before runtime;
- still exercise client-only entry points such as Creative inventory on the exact release candidate.

### D. Harness error distinguished from product error

A prior oak/iron gameplay check was invalid because the player fell away from the target. It was replaced with a platformed/fixed-position test and authoritative post-action block-state assertions.

The Dev Kit now records this as a general rule: moving/falling actor fixtures must not be allowed to create false gameplay failures when proximity matters.

## Unrecovered exact-final bytes

The resumed runtime, Google Drive search, and GitHub search did not yield the separate final Future Feature Bridge JAR/source snapshot identified by `b20acda1e2d...`, nor the full 64-character final companion SHA-256. The older Ragdoll/Sable checkpoint recovered from Drive predates/separates that companion and contains no `ragdoll_reactions_future`, `justtrialspawners`, `prepare-uninstall`, or Future Feature Bridge source.

Therefore this durability record does not claim that the final companion bytes were recovered or rebuilt in the resumed session. The runtime results above are preserved from the exact dead-chat checkpoint; recovered byte claims are limited to section A.

## Release rule going forward

A Future Feature Bridge release is not considered green unless the same exact candidate hash crosses all applicable gates:

- production symbolic-linkage validation;
- native client boot and in-world join;
- Creative inventory/client-only linkage surfaces;
- embedded provider behavior;
- external provider behavior;
- exactly-once impact semantics;
- same-world provider removal/fallback;
- uninstall cleanup;
- dedicated server without optional provider;
- dedicated server with optional provider stack;
- clean shutdown and preserved logs;
- deterministic rebuild/hash comparison.
