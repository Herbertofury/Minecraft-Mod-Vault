# Forge 1.20.1 Ragdoll Port Verification

Status: build/package verification green; exact user-instance runtime acceptance pending.

Target:
- Minecraft 1.20.1
- Forge 47.4.23
- Java 17
- Ragdollified 1.0.0-RELEASE (exact backend target)

Artifacts:
- `sable_player_ragdoll-1.20.1-forge-0.7.2.jar`
  - SHA-256 `5e00e3f80892b5e0ca2aa8c13753b663286dc9212142c61e9fd2d88e9cbef5b3`
  - Drive file `1O7a3Utr0nKztAk6nk_RCmG5c9hkMny8c`
- `ragdoll_reactions-1.20.1-forge-0.7.0.jar`
  - SHA-256 `0cd76654bdfc1e34a1fe490c28d24afbbdc7e39baa59aca4359aedb650554bbb`
  - Drive file `1_Pv1iCo8qHjF-bi8202x1AZ2uUdqA88B`
- `ragdoll-forge-1.20.1-BUNDLE.zip`
  - SHA-256 `2bf3e0179d36cf9529352a9045f15ebc082007ce49f5f76b10fde9d6c907754a`
  - Drive file `1fIiNb-joEEfaBl7n-YamWaoRc5DhI6An`
- `ragdoll-forge-1.20.1-SOURCE.zip`
  - SHA-256 `154f333ab8381d0ecf7b2eb11fbebb54da5d0e364c1aa09f1fbca35786b36347`
  - Drive file `1PlHcfdaHr3x_0roYU8jhfSERGx95rWO_`

Verification completed:
1. Official Forge 47.4.23 MDK/toolchain export completed successfully and supplied the exact mapped Forge + Ragdollified compile classpath and MCP-to-SRG mappings.
2. Real mapped Java 17 compile completed with zero errors for the core and Reactions modules.
3. Production remap validation covered 108 classes and reported `staleMethods=0`, `staleFields=0`.
4. Both packaged JARs pass ZIP integrity, use classfile major 61, have Forge metadata, and contain no NeoForge references or signature remnants.
5. Original non-class resources pass the tracked parity audit: core 16/16 and Reactions 4/4, with only the required Minecraft 1.21 `tags/item` to 1.20.1 `tags/items` path translation.
6. Tracked original player/playerless/mob public API audit reports zero missing public members.
7. Fresh bundle extraction and internal SHA-256 verification pass after correcting the bundle checksum paths.
8. Google Drive round-trip downloads of both JARs, bundle, source archive, and checksum manifest are byte-identical to the local final artifacts.

Implementation notes:
- The port uses an explicit adapter for the exact Ragdollified 1.0.0 API rather than an unconstrained runtime reflection scanner.
- Native Forge config, networking, lifecycle/events, manual H-key trigger, release/control flow, player/mob/playerless sessions, equipment refresh, `/ragdoll`, and the Ragdoll Reactions trigger matrix are wired in the Forge 1.20.1 source.
- The Reactions source is not published in this public repository because the upstream project is All Rights Reserved. The private full source checkpoint is stored on Drive.
- Ragdollified does not expose the original Sable server-side dismember physics API. The public compatibility signatures remain present, but exact Sable-internal dismember behavior is not claimed.

Remaining acceptance gate:
- Launch these exact hashes in the user's real Forge 1.20.1 instance with Ragdollified 1.0.0-RELEASE, then exercise client startup/world join, manual H-key ragdoll/release, hit/fall/explosion/lightning reactions, mob ragdolls, multiplayer synchronization, and restart/persistence before promoting the repair record from partial to runtime-confirmed success.
