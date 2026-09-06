# Mob Charms 0.4.1 - Broom Dangle V3 native PASS

## Result
PASS for whole-mascot Charmsy-quality pendulum physics on Hexerei broom keychains.

V2 solved broom keychain size and attachment, but the broom-specific mascot geometry intentionally removed Charmsy's visible `chain` bone and accidentally lost the *physics parent* carried by that bone. Internal wing/limb physics still ran, but the whole mascot no longer swung like the original Charmsy vanilla-sword charm.

## Production fix
The broom generator now creates a zero-geometry, zero-texture root named `broom_dangle`, reparents the mascot root beneath it, and copies the immutable source Charmsy `chain` pendulum settings onto that invisible parent.

This restores the whole-mascot pendulum without restoring Charmsy's visible chain. Hexerei remains the only visible loop/tether/chain and still owns the broom attachment point and broom-space transforms.

Original per-mascot tuning is preserved exactly. Examples: Allay uses Charmsy's `forward` 40-degree pendulum; Bee uses Charmsy's `full` 30-degree pendulum. Existing authored wing/limb/body physics remain nested beneath the new parent.

## Native proof
- Minecraft 1.20.1 / Forge 47.4.23 / Eclipse Temurin 17.0.20.1+1
- Hexerei v4 SHA-256 `7107bdcba1a4d6884bab240e12f7b2ecc47b41c976feb0496b367ccf0be785c2`
- Punchy 2.7e SHA-256 `459e6889c11ecfbd1faf13749dafa8eea8df39e4d569e8b2bb0cb80aae94dcf2`
- Final JAR: 1,514,901 bytes, SHA-256 `f03c62f0cf644a36940be5846c57d814141eae999d209931183b078aed9eeb11`
- Final bundle: 15,194,298 bytes, SHA-256 `f6843e5eed8ba105f08957b002e7b19cee1e267c8dc91d79fe3e5b04d0cc893e`
- 9/9 broom definitions: exact invisible source-chain pendulum tuning PASS
- All nine mascots loaded/rendered in the real packaged client with Punchy 2.7e + Hexerei v4
- Bee full-pendulum motion: visible whole-mascot off-vertical swing and reversal PASS
- Allay forward-pendulum lane: PASS
- Duplicate Charmsy chain geometry: 0
- Task-related render/bridge errors: 0
- Clean integrated-server save for all dimensions and clean client stop: PASS

## Frozen regression boundary
- `GripFitEngine.java` remains byte-identical at SHA-256 `a5f5d7bdcd1722eab5a80fa0109161a56f2b7e1577677af36f01467b9dbbff0b`.
- Original immutable Charmsy baseline: 182 files PASS.
- Original Charmsy vanilla-sword definitions: 60 PASS.
- V2 -> V3 source delta is restricted to 18 broom-only generated JSON files plus `generate_hexerei_broom_charms.py` and `audit_hexerei_broom_charms.py`.
- Weapon/GripFit/bridge source is unchanged by V3.

## Durable artifacts
Google Drive project: `Minecraft Dev Kit/Projects/Mob Charms - Standalone/03 Builds & QA`

- V3 native-PASS ZIP Drive file id: `1UpJOouKRuh662NFy86FrRHjnZUP37Mbs`
- V3 standalone JAR Drive file id: `1Qeqg4ZrPlBZGZ35UVdqHMaeCDOHdLPk7`
- Drive roundtrip byte identity: PASS for both ZIP and JAR.
- Remote ZIP integrity (`unzip -t`): PASS.

The final Drive ZIP is the authoritative complete source/evidence snapshot and includes exact source, build log, native client log, audits, QA datapack, motion MP4/GIF, representative swing frames, checksums, and verification receipt.
