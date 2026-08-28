# Ragdoll Forge 1.20.1 — Candidate C checkpoint

Minecraft 1.20.1 / Forge 47.4.23 / Java 17.

## Immutable candidate

- `ragdoll_reactions_future-1.20.1-1.0.0.jar`
- SHA-256: `e0c4ee4e95e12d824b4ccd49b0441a40a95ae83a74d34f628abce219d9598fbc`

Strong base artifacts:

- `sable_player_ragdoll-1.20.1-0.7.2.jar` — `e6d0192a39ddf645078189a6e8ace28a98ffa24feb4739b4cad21faa8fce13bd`
- `ragdoll_reactions-1.20.1-0.7.0.jar` — `afe7d7b2c20577a6136874329311fa2ce2995e683a7b2c0ccb02922b718f4bb2`
- `sable-forge-1.20.1-2.0.0-all.jar` — `3e1b2a1eae26a351d79b965c523454d43510e5111b60288a86c61431f3cc2c8c`

## Fresh green gates

- deterministic bridge build identity (two builds byte-identical)
- production symbolic-linkage validator: 0 problems
- LambdaMetafactory/invokedynamic SAM remap + negative regression fixture
- embedded dedicated server + clean shutdown
- embedded authoritative projectile gameplay: exactly one semantic trigger
- oak trapdoor opens while iron trapdoor stays closed
- prepare-uninstall entity/drop fixture: 3 projectiles + 1 dropped charge removed
- external JTS 1.2.1 dedicated server with exact FTB Library 2001.2.12 + Architectury 9.2.14
- same-world provider removal: expected removed-provider registry warnings, resolver returns to embedded, clean save/shutdown

## Durable Drive checkpoint

`ragdoll-forge-1.20.1-candidate-C-checkpoint-2026-08-27.tar.gz`

- size: `22,300,889` bytes
- SHA-256: `cc6f467e98f47afe15beb8ceed3c27c787f8ccdf058e5780e0c3a350747d7ddd`
- Drive file id: `1sSEKm-toaDKx26WU3sKuFAvWYuMo2-py`
- round-trip re-download SHA-256 matched exactly

## Native-client gate state

Packaged-JAR Forge client harness uses verified Mojang 1.20.1 assets, Linux LWJGL natives, Java 17, and exact production mod bytes. Two harness-only blockers were fixed without changing candidate bytes:

1. ForgeGradle `downloadMCMeta` always performs a network lookup even offline, so verified metadata is staged and only that network task is disabled.
2. Source-less userdev projects lose nonexistent output directories; a non-mod harness anchor class/resource now creates real source-set outputs.

A third userdev-only compatibility condition was proven: production Sable correctly targets SRG `Minecraft.f_91073_`, while `forgeclientuserdev` exposes mapped `Minecraft.level`. The exact generated SRG→MCP map contains `f_91073_ -> level`; the harness now enables Mixin refmap remapping instead of mutating the production Sable JAR.

Remaining acceptance: finish packaged native client/integrated server, open Creative inventory, player inventory + ender cleanup, external JTS Survival 2→1 exactly-once client lane, then final release packaging/publication.