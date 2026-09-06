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

## Final binary

- File: `SLR-Shared-Mana-Forge-1.20.1-1.1.0.jar`
- SHA-256: `d8baab5056e01bd88dd4f7d3b306317a98476cf4d937faecf9252d9a5e4b08bb`

## Known validation boundary

The exact production SLR/ISS runtime JAR pair was unavailable in this environment for a fresh client/server launch, so no fresh full-game runtime claim is made for this checkpoint. See `QA-REPORT.md` and `evidence/` for the completed deterministic gates.
