# Spellbrook Forge 1.20.1 — 0.2.0-dev PRETTY FINAL v4

Status: **PASS — unrestricted camera QoL + dual broom steering + 29/29 spell runtime verification**

This checkpoint supersedes PRETTY FINAL v3 for the current Spellbrook release.

## Final artifact

- JAR: `Spellbrook-Forge-1.20.1-0.2.0-dev-PRETTY-FINAL-v4.jar`
- Size: **6,937,409 bytes**
- SHA-256: `d4be9338e759e5d364da6e53225f5a0472919fd4ff8aeb47f52a63a44c213466`
- Drive file ID: `1fDNwAWuh14FdmJ9y6NSMknwWyVxNiaZm`

Source:
- `Spellbrook-Forge-1.20.1-0.2.0-dev-PRETTY-FINAL-v4-SOURCE.zip`
- SHA-256: `41ceb7bd2298c5bc424137aed983f1c0b0f8f5bfd9d2b1f0d3a4d11ffff90c2e`
- Drive file ID: `1q0kYGEbIlXO7htSXldnPNDg1tGFADXwU`

Evidence:
- `Spellbrook-Forge-1.20.1-0.2.0-dev-PRETTY-FINAL-v4-EVIDENCE.zip`
- SHA-256: `219e6bdd4168c1047ea9691c5ecf6f8ca0087708ea952061c2d692f1828a4ac1`
- Drive file ID: `1-BS5-yLnfSYWX4Wr6Y7On7-afvK_16Kd`

Verification report Drive file ID: `1IGP-p3x44bVShN_LQsriDWUEQKs8fZiD`
SHA manifest Drive file ID: `1MG80fkbV5IbxCMUYVkJam7CuWd7b3v5Q`
Canonical Drive folder ID: `1xA5xcdiMVD3kDo0UD5l4ouI_goWMsZFg`

## Broom QoL

Two persistent steering modes now coexist:

- **FREE_LOOK** (default): normal independent Minecraft camera. W/S drive, A/D steer, Space/Shift climb/descend, Sprint boost. F5 first-person / rear third-person / front third-person cycles normally while mounted.
- **LOOK_TO_STEER** (optional): while moving forward with A/D neutral, the broom smoothly follows camera yaw. A/D remains a manual override.

The steering mode is a Forge CLIENT config and can be toggled mid-flight with the configurable `Toggle Broom Steering Mode` keybind (default V, remappable under Controls > Spellbrook).

Native proof:

FREE_LOOK with camera/player yaw deliberately set to 90 degrees and broom yaw 0:
- start `[0.500,150.020,0.500]`, broom yaw `0`, player yaw `90`
- after forward input: Z `9.725`, broom yaw still `0`, player yaw still `90`

LOOK_TO_STEER from the same start:
- after forward input: player/broom `[-3.596,150,4.058]`
- broom yaw `90`, matching the camera heading

Protocol bumped to `3` so stale packet layouts cannot silently connect.

## Spell verification

All 29 Spellbrook spells were exercised in a real Forge integrated client/server:

- 29/29 PASS
- 25 offensive/targeted spells
- 25/25 offensive spells had server-observed real `hurtTime`
- heal/barrier/mobility spells validated against their real player/self state
- offensive fixture used a real 500-HP Iron Golem and authoritative health/hurt/fire/effect/projectile/manifest state
- native GIFs show real hurt flashes, knockback/status manifestations, barriers, healing, and mobility rather than particle-only placeholders

Floral Stairway received a reliability repair so the server-authoritative takeoff/motion cannot be immediately erased by the client movement path.

## Final gates

- clean `compileJava`: PASS in 16s
- 1,382 Spellbrook JSON resources parsed, 0 errors
- temporary QA classes in shipping JAR: 0
- production QA log spam in shipping source: 0
- dedicated server: PASS, `Done (8.033s)!`, then `All dimensions are saved`
- production-only integrated client after QA removal: PASS, `Dev joined the game`
- final client task-owned scan: 0 Spellbrook/model/texture/network failures
- final client Save & Quit: `All dimensions are saved`
- shipping config at final smoke: `steeringMode = "FREE_LOOK"`
- normal product `gradle --offline --no-daemon clean build`: PASS in 23s including `reobfJar`
- QA-only srgutils alias used only by the restored runtime harness, never by the shipping build/JAR
- final JAR archive integrity: PASS

No image generation was used for runtime evidence.
