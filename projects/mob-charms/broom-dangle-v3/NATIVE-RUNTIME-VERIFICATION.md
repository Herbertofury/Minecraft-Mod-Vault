# Mob Charms 0.4.1 - Broom Dangle V3 native verification

## Acceptance result
PASS: Hexerei broom keychain mascots now inherit the same whole-object pendulum behavior authored by Charmsy for the original vanilla-sword charms, while Hexerei remains the only visible tether.

## Runtime matrix
- Minecraft: 1.20.1
- Forge: 47.4.23
- Java: Eclipse Temurin 17.0.20.1+1
- Hexerei v4 SHA-256: `7107bdcba1a4d6884bab240e12f7b2ecc47b41c976feb0496b367ccf0be785c2`
- Punchy 2.7e SHA-256: `459e6889c11ecfbd1faf13749dafa8eea8df39e4d569e8b2bb0cb80aae94dcf2`
- Mob Charms V3 SHA-256: `f03c62f0cf644a36940be5846c57d814141eae999d209931183b078aed9eeb11`

## Physics contract
Every broom mascot gets one `broom_dangle` root with:
- pivot `[0, 0, 0]`
- no cubes
- no texture
- mascot authored root reparented beneath it
- exact pendulum settings copied from the immutable original Charmsy `chain` bone for that mascot

The source `chain`/`chainnn` geometry remains absent from broom geometry. Hexerei's rendered chain/tether is therefore the only visible tether.

The exact original Charmsy modes remain intact: Bee uses `full` / 30 degrees; the other current broom mascots use their source-authored settings, including Allay `forward` / 40 degrees. Internal mascot bone physics are not replaced or flattened.

## Native visual gates
- Real packaged Forge client reached the existing `BroomFitQA` integrated world with exact Punchy 2.7e + Hexerei v4.
- All nine variants loaded and rendered: Allay, Axolotl, Bee, Camel, Chicken, Creeper, Pufferfish, Warden, Zombie.
- Allay rest/forward-pendulum lane: PASS.
- Bee rest + full-pendulum acceleration test: PASS; the entire Bee visibly tilts off vertical and reverses while remaining attached.
- Duplicate Charmsy chain geometry: none visible / static count zero.
- Fresh native log: no task-related Mob Charms Bedrock-item/Hexerei bridge/model errors.
- Client exited through Minecraft `Save and Quit to Title`; integrated server saved Overworld, Nether, and End; final client `Stopping!` marker observed.

## Regression gates
- Immutable Charmsy baseline: 182 files PASS.
- Vanilla-sword isolation: 60 sword definitions PASS.
- Charmsy overlays: 18 checked, 0 failures.
- GripFit source hash unchanged: `a5f5d7bdcd1722eab5a80fa0109161a56f2b7e1577677af36f01467b9dbbff0b`.
- V2 -> V3 changed-path challenge: only 18 broom JSON outputs + generator + broom audit.
- Offline Forge build: PASS.

## Durable artifact verification
- ZIP: 15,194,298 bytes, SHA-256 `f6843e5eed8ba105f08957b002e7b19cee1e267c8dc91d79fe3e5b04d0cc893e`
- JAR: 1,514,901 bytes, SHA-256 `f03c62f0cf644a36940be5846c57d814141eae999d209931183b078aed9eeb11`
- Drive ZIP id: `1UpJOouKRuh662NFy86FrRHjnZUP37Mbs`
- Drive JAR id: `1Qeqg4ZrPlBZGZ35UVdqHMaeCDOHdLPk7`
- Both Drive files were materialized back and compared byte-for-byte to local originals: PASS.
- Remote ZIP integrity: PASS.
