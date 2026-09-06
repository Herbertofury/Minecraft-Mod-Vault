# Mob Charms 0.4.1 - BroomFit V2 native PASS

## Result
PASS for the reported Hexerei broom-keychain defect where the mascot rendered as a near-pixel speck.

## Root cause and fix
Hexerei 1.20.1 `BroomRenderer` applies a 0.25 scale to its normal keychain payload before `renderItem`. Mob Charms' mascot-only geometry was already compact, so it inherited an unintended second shrink. `HexereiBroomKeychainBridge` now compensates the Hexerei payload scale and normalizes each of the nine authored mascots to a 0.26-0.34 block visual envelope.

Hexerei remains authoritative for the gold wrap/collar, grey tether, broom attachment point and swing. Mob Charms adds no duplicate chain and does not replace the broom model.

## Native proof
- Minecraft 1.20.1 / Forge 47.4.23 / Temurin 17.0.20.1+1
- Hexerei v4 SHA-256 `7107bdcba1a4d6884bab240e12f7b2ecc47b41c976feb0496b367ccf0be785c2`
- Punchy 2.7e SHA-256 `459e6889c11ecfbd1faf13749dafa8eea8df39e4d569e8b2bb0cb80aae94dcf2`
- Final JAR: 1,514,566 bytes, SHA-256 `ec40642d386ca88db166a414a534faa945e07d18f66ebebda9354bd72129d341`
- Final bundle: 8,760,977 bytes, SHA-256 `b10ec31af895ec97c76e2d07b2ebfb6e524fcb4158a36ae6f34623319421466e`
- Willow / Mahogany / Witch Hazel Allay matrix: PASS
- Nine mascots: PASS
- Real broom-motion/tether capture: PASS
- Fatal runtime regression matches: 0
- Task-related model/texture error matches: 0
- Integrated world saved all dimensions and client stopped cleanly

## Frozen regression boundary
`GripFitEngine.java` remains byte-identical at SHA-256 `a5f5d7bdcd1722eab5a80fa0109161a56f2b7e1577677af36f01467b9dbbff0b`. The original immutable Charmsy sword baseline, overlays, deferred renderer, provider contract, Punchy ToolKind contract and 36 wrap variants pass. Source delta from the prior native-PASS checkpoint is exactly the broom bridge plus its broom-specific audit.

The two old `/mnt/data/fit-analysis` image-corpus scripts were not re-run because that ephemeral corpus is absent in this worker; they are outside this broom-only change and the covered GripFit source is byte-identical to the prior native-verified baseline.

## Durable artifacts
Google Drive project: `Minecraft Dev Kit/Projects/Mob Charms - Standalone`

- Final native-PASS ZIP Drive file id: `1S7FtQpFTrjOJlULu6aoLQtW2T84G6umd`
- Final standalone JAR Drive file id: `1YVIpV_RVxXyJZSCvtkQqYWMQ_xAi12tr`
- Project Brain receipt Drive file id: `1-a308YvroKKewtOPjG3SJZ0rOSn_fanz`

The final ZIP includes exact source, source delta, build/runtime logs, before/after evidence, three-broom matrix, nine-mascot contact sheet, motion proof and checksums.
