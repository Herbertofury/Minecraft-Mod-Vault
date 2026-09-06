# Mob Charms 0.4.1 - BroomFit V2 native verification

## Acceptance result
PASS for the reported broom-keychain size/attachment defect.

## Root cause
Hexerei's real Forge 1.20.1 `BroomRenderer` applies a 0.25 payload scale to the normal keychain-contained item before its `renderItem` call. Mob Charms' mascot geometry is already authored as a compact charm, so the bridge inherited a second shrink and produced the near-pixel mascot shown in the user report.

## Production fix
Only the broom presentation bridge changed. `HexereiBroomKeychainBridge` now compensates the upstream 0.25 payload shrink and applies per-mascot normalization. Net native visual envelope is 0.26-0.34 block across all nine mascots.

Hexerei continues to own and render the gold loop/collar, grey dangling tether, broom attachment point, and broom-space swing transforms. Mob Charms does not add a second chain and does not replace the broom model.

## Native runtime
- Minecraft: 1.20.1
- Forge: 47.4.23
- Java: Eclipse Temurin 17.0.20.1+1
- Hexerei v4 SHA-256: 7107bdcba1a4d6884bab240e12f7b2ecc47b41c976feb0496b367ccf0be785c2
- Punchy 2.7e SHA-256: 459e6889c11ecfbd1faf13749dafa8eea8df39e4d569e8b2bb0cb80aae94dcf2
- Candidate JAR: 1514566 bytes
- Candidate SHA-256: ec40642d386ca88db166a414a534faa945e07d18f66ebebda9354bd72129d341
- Source archive SHA-256: 21c57641b01c94e82929c60f8dbc30c421de91f606f5d1b21f19564e4d075830

## Native visual gates
- Willow + Allay close-up: PASS; readable charm with Hexerei loop and tether visibly attached.
- Willow / Mahogany / Witch Hazel matrix: PASS.
- All nine mascots: PASS (Allay, Axolotl, Bee, Camel, Chicken, Creeper, Pufferfish, Warden, Zombie).
- Broom motion capture: PASS; mascot remains attached through real Hexerei broom movement/tether path.
- Native fatal regression scan: 0 mixin-apply failures, NoClassDefFoundError, OOM, crash report, failed-mod-load, or mod-loading-error matches.
- Native model/texture error scan: 0 task-related matches.
- Clean integrated-server save: PASS; overworld, Nether, End all saved.
- Clean client stop marker: PASS.

## Frozen regression boundary
- GripFitEngine.java SHA-256 remains: a5f5d7bdcd1722eab5a80fa0109161a56f2b7e1577677af36f01467b9dbbff0b
- Expected frozen hash: a5f5d7bdcd1722eab5a80fa0109161a56f2b7e1577677af36f01467b9dbbff0b
- Original Charmsy sword baseline audit: PASS (182 immutable files, 60 sword definitions).
- Source delta versus the previous native-PASS checkpoint: exactly `HexereiBroomKeychainBridge.java` plus its broom-specific audit script.

## Static regression
Relevant unchanged/frozen-path audits pass. The two historical GripFit image-fixture scripts (`audit_gripfit.py`, `audit_full_family_fit.py`) depend on the external ephemeral `/mnt/data/fit-analysis` image corpus and were not used as a fresh gate in this worker. This is non-blocking for BroomFit V2 because GripFit source is byte-identical to the already-native-verified baseline and the immutable/sword/provider/deferred-overlay contracts all pass.
