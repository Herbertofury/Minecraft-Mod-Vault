# More RPG Library 2.7.2 -> native Forge 1.20.1 loader/dependency translation

Exact source authorities:
- Fabric 1.20.1 substrate `7da3c766ef5aebd850a0eb2f6a26bde2409f626f`
- 2.7.2 modern common/NeoForge authority `bfa35e55133bd795676b6beb40c334d2904bf0ea`

## What can be reused structurally

Modern 2.7.2 already moved most gameplay ownership into the `common` module. Treat that common module as the behavior authority, but compile it against Minecraft 1.20.1 mappings/APIs and provide native Forge lifecycle/registry/network/event glue. Do not copy NeoForge classes or metadata into the final JAR.

Modern NeoForge already consumes `dev.kosmx.player-anim:player-animation-lib-forge`, so the player-animation provider family has a direct Forge variant and should remain an external production dependency unless graduation proves bundling is required.

Cloth Config also has a Forge family and should use the exact Forge 1.20.1-compatible artifact, not a Fabric or NeoForge remap.

## Certified RPG foundations

During graduation, do not resolve moving latest Modrinth artifacts for these foundations. Use the certified exact Forge 1.20.1 outputs:
- Spell Engine 1.10.4: wait for frozen replay identity before activation
- Spell Power 1.6.0: `640f8de69e60b57c18ac232938bdbde67e5579d1d3021847bf699e79dc1cda45`
- Ranged Weapon API 2.3.4: `e387df1f42473864e687715c2495e66d57f64b53daae463b5d2e2157c2da6894`
- TinyConfig 3.1.0: `0182a492d6c59d7d5f491a39bb2f6634ba5dd38083295305c4769fdb6539db18`
- Structure Pool API 1.2.1: reuse certified project output only where current More RPG behavior still needs it.

Critical Strike remains compile/compat optional. The base More RPG JAR must start without it; a present-provider lane must prove the integration.

## Old 1.20.1 dependencies that are references, not final architecture

The historical 1.20.1 build used:
- Fabric API
- Fabric Loader
- Trinkets
- AzureLib Armor
- Fabric Player Animator
- Fabric Cloth Config
- Fabric Spell Engine / Spell Power / Ranged Weapon API
- embedded historical TinyConfig and Structure Pool API

These identify target-era behavior and data contracts. They do not justify Fabric/Connector/FFAPI leakage in the final native Forge build.

## Modern loader-specific dependencies requiring adaptation

### Accessories / owo

Modern common source compiles against Accessories and modern NeoForge uses `accessories-neoforge` plus owo/endec runtime libraries. This is not automatically a valid Forge 1.20.1 dependency graph.

Rules:
1. Inventory every actual Accessories symbol used by common source.
2. Separate public behavior contract (equipment/accessory slot semantics) from modern loader implementation.
3. Adapt to a Forge 1.20.1-compatible equipment/accessory provider only where behavior requires it.
4. Keep base library startup independent from Armory and other downstream content mods.
5. Do not add owo/endec merely because modern NeoForge lists them; add only if the selected Forge 1.20.1 provider actually requires them and prove packaged runtime.

### Fabric annotations/API remnants

Modern common intentionally uses Fabric Loader for `@Environment` remapping and currently declares Fabric API in common. The native Forge port must remove runtime dependence on Fabric/FFAPI/Connector. Client-only separation must be expressed through Forge/dist-safe organization rather than shipping Fabric loader requirements.

## Native Forge 1.20.1 target shape

Target:
- Minecraft 1.20.1
- Forge 47.4.23
- Java 17
- mod id remains `more_rpg_classes` unless upstream compatibility proves otherwise
- native `META-INF/mods.toml`
- native Forge registries/events/networking/config lifecycle
- clean dedicated-server classloading with no client-only symbols
- clean client resource/render/animation bootstrap

The final release must contain no Fabric/NeoForge/Connector metadata and no hard runtime symbols from those loaders.

## Build/graduation discipline

1. Assemble exact dual-authority source deterministically.
2. Port common behavior by feature family before loader glue.
3. Compile against certified exact local Maven/module artifacts for graduated RPG foundations.
4. Run Java 17 bytecode and metadata/dependency gates.
5. Run dedicated server semantic tests for relations, status effects, AI, loot and worldgen.
6. Run native Forge client tests for render/particle/HUD/animation behavior.
7. Run integrated-world gameplay and restart/config persistence.
8. Exercise optional-provider matrix (Critical Strike absent/present; Armory absent/present when available; chosen accessory provider).
9. Build untouched release twice and require deterministic payload/JAR identity.
10. Run fresh packaged Forge server using untouched release JAR.
11. Freeze first-green JAR/source identities and independently replay.
12. Publish evidence to GitHub and canonical Drive; promote only by non-force fast-forward.
