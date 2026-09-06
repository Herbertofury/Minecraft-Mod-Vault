# SLR Shared Mana — v1.1.0 checkpoint

## Objective restored and completed

Add a second mana behavior as an explicit opt-in while preserving v1.0.0 as the default:

- **Default ON:** classic shared SLR-authoritative mana (`enabled=true`).
- **Default OFF:** separate native pools with SLR emergency borrowing (`separatePoolsBorrowFromIron=false`).

## Separate-pools behavior

When `separatePoolsBorrowFromIron=true`:

- SLR MP remains native, including native regen, rewards, potions, resets and HUD.
- Iron mana remains native, including regen, events, caps and HUD.
- SLR spend paths consume SLR MP first.
- Only the actual shortage is debited from Iron, using `slrMpPerIronMana`.
- No idle transfer, mirror, player-tick polling or nearby scan exists.
- Positive SLR gains remain SLR-only.
- Iron debit uses Iron's native mana setter and native client mana sync.

## Coremod scope

- 43 verified SLR mana-spend classes.
- 1 read-only Spirit Bow pre-use affordability gate.
- 44 unique transformer definitions total.
- Regen/reward/reset/potion/HUD and `GuildBuffManager` are intentionally excluded.
- Non-spending reserve checks such as Goliath manifestation are intentionally excluded.

## Final validation

- Acceptance matrix: GREEN.
- Java 17 compile: GREEN.
- Forge 47.4.23 production reobf: GREEN, 0 missing-class errors.
- JAR integrity: GREEN.
- Production SRG linkage: GREEN.
- Packaged Nashorn + ASM coremod harness: GREEN (43 full + 1 read-only).
- Default/classic regression gates: GREEN.
- Final challenge pass fixed native Iron HUD synchronization after fallback debits.

## Final artifacts

Canonical Drive project folder: `Minecraft Dev Kit/Projects/SLR Shared Mana - Forge 1.20.1`

- Binary: `SLR-Shared-Mana-Forge-1.20.1-1.1.0.jar`
  - Drive file ID: `1dYU2ZQMGOBh64-aRe9iU7DlHWjG_uKjU`
  - SHA-256: `d8baab5056e01bd88dd4f7d3b306317a98476cf4d937faecf9252d9a5e4b08bb`
- Release bundle: `SLR-Shared-Mana-Forge-1.20.1-1.1.0-RELEASE.zip`
  - Drive file ID: `1-6SoZZrtPjs3eEl8qf_TEfNDad_sjkUg`
  - SHA-256: `058034849ea3f15089e0f24a45c55056e5ab10ed374bc9022d21797eeaf82045`
- Complete source ZIP: `SLR-Shared-Mana-Forge-1.20.1-1.1.0-SOURCE.zip`
  - Drive file ID: `1jIL3H8lg5qSFjYicS5vy_b6hZ_GYpV3m`
  - SHA-256: `e63d5899d9e82876bcd43e771aaabf34010debdafb342c8e5a09cb752f087956`
- Sources JAR: `SLR-Shared-Mana-Forge-1.20.1-1.1.0-sources.jar`
  - Drive file ID: `1DOy-JZIFhLt7Udh4mLz2vm8tv9-S_q8V`
- SHA manifest: Drive file ID `1a8W-f90TCZQQ9RUEhr-3N01EUaQW4Qar`
- Drive checkpoint: Drive file ID `1oyvqkh58c9VTRxnSk8kKzvcPBPmUwhlZ`

GitHub's connected contents API supports text source/checkpoints but not release-binary asset upload in this environment, so Drive is the canonical home for the exact binary/source ZIP/release bundle bytes.

## Known validation boundary

The exact production SLR/ISS runtime JAR pair was unavailable in this environment for a fresh client/server launch, so no fresh full-game runtime claim is made for this checkpoint. The release bundle contains the QA report and deterministic evidence used for acceptance.
