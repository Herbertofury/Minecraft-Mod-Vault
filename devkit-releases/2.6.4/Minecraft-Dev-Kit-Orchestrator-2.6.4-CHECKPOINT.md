# Minecraft Dev Kit Orchestrator 2.6.4 — Backport Ownership Broker + Multi-Version Source

Generated: 2026-08-29 UTC

## Requested behavior

Prevent ports/mods from redundantly backporting the same later-vanilla systems. Use one canonical Future Vanilla Backport provider that yields feature-by-feature to target vanilla or a proven installed external backport. Make one source tree easy to build across Minecraft/loader versions with a Stonecutter-style matrix.

## Implemented

- Added `vanilla-atlas build-backport-catalog`.
- Added `vanilla-atlas resolve-owner`.
- Added `vanilla-atlas validate-capabilities` with duplicate-surface detection and a `--require-complete` full-release gate.
- Ownership precedence is `VANILLA_TARGET > PROVEN_EXTERNAL_PROVIDER > FUTURE_VANILLA_BACKPORT`.
- Multiple proven external owners fail closed as a conflict; load order is never treated as ownership.
- Consumer ports are forbidden by default from privately owning cataloged future-vanilla surfaces.
- Added `multiversion scaffold|validate` around Stonecutter 0.9.7.
- Default example matrix: 1.20.1 Forge/Fabric (Java 17), 1.21.1 NeoForge/Fabric (Java 21), 26.2 NeoForge/Fabric (Java 25).
- Updated Minecraft Dev Kit skill with the single-owner invariant and multi-version-source workflow.

## Real generated 26.2 -> 1.20.1 catalog

- 25,789 target-missing surfaces total
- 14,158 registry entries
- 6,293 data paths
- 2,772 resource paths
- 1,565 external assets
- 500 SoundEvents
- 501 sound definitions
- 584 surfaces have exact static ownership evidence in the imported Vanilla Backport 1.20.1 source inventory

## Verification observed

- `go test ./...`: PASS.
- Fresh-extracted source ZIP: `go test ./...` PASS and executable reports `Minecraft Dev Kit Orchestrator 2.6.4`.
- Windows amd64 cross-build: PASS.
- Six-node multi-version matrix validation: PASS.
- Proven-provider owner resolution: `minecraft:item.wolf_armor.break -> vanillabackport`.
- Provider-absent owner resolution: same feature -> `futurevanillabackport`.
- Empty capability manifest passes structural validation but intentionally fails `--require-complete` with 25,789 unassigned surfaces and exit code 4.
- Source/control-plane/skill ZIP integrity checks: PASS.

## Artifact SHA-256

- `Minecraft-Dev-Kit-Orchestrator-2.6.4-SOURCE.zip`: `126504f4957a061d2699e2b4e40d512e8cda0a3d0a94e1ca73a2e7da1cd217ac`
- `mmv-devkit-linux-amd64-v2.6.4`: `e70233762f58c680f3dd8daab263d4d5217018a931c0b781e1eadd515c39d70e`
- `mmv-devkit-windows-amd64-v2.6.4.exe`: `e13e73ff8a007d8574238715412370fe7bb4b6a044872260fc64fd78b5381cd9`
- `Future-Vanilla-Backport-Control-Plane-2.6.4.zip`: `c475cf900500da66cb5d00d8ed37dd264aa5853e8588602ccbf6131be16cf583`
- `skill.zip`: `10edb3ee7fb327138f3ff89ba391ca3735636317f311b8d8fdcaf3f5c60dba39`

## Important scope truth

2.6.4 builds and verifies the ownership/control-plane and multi-version development system. It does **not** claim that the actual 26.2 gameplay/content backport is finished yet; the full-release gate correctly blocks that claim until every cataloged surface is grouped into a capability module, implemented, and runtime-certified.
