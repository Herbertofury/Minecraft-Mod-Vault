# Spellbrook Forge 1.20.1 — PRETTY FINAL v3

Status: **PASS — native broom vehicle / rider mount repair**

This checkpoint supersedes PRETTY FINAL v2 for Spellbrook's native broom path.

## Final artifact

- JAR: `Spellbrook-Forge-1.20.1-0.2.0-dev-PRETTY-FINAL-v3.jar`
- Size: **6,933,330 bytes**
- SHA-256: `948bab2296f4698a877569b3d57100ac367c7f3f9f3335e93731500b8cf436bc`
- Drive file ID: `18mFTCaQm1MzEBmecJ4TjNxZChaJ7VdiY`

## Source / evidence

- Clean source ZIP SHA-256: `5d240db1bdb815e9c070a012ef3a449dd21a59508349d2884ca8cc299a55c602`
  - Drive file ID: `1Ds9k7GnyGnB8c4tT761_C-vABoJDXVmm`
- Evidence ZIP SHA-256: `2d0ce8ea1e665c73461d376017b884771e6b57e6f17432a60cb1d092dd8aa8e8`
  - Drive file ID: `1BEBtVqsmlBCtV5pGG-symPTRhkOmU9YH`
- Verification report SHA-256: `6c65cc91b28013f644c4973fe739cc6ae8f3be8c7c580e4534001ed2702d1963`
  - Drive file ID: `1RCktT7KEFzoQ9xG3NC1Oq94qfRMXtUkL`
- SHA manifest Drive file ID: `1SnAVw2HdehBGj8i9N768Vn4XxhnYJxvl`
- Canonical Drive folder: `1xA5xcdiMVD3kDo0UD5l4ouI_goWMsZFg`

Native media:
- Flight GIF Drive file ID: `1uKtfj_UrQrkD-OodJPEzKCh6Dp4Za4Bs`
- Flight MP4 Drive file ID: `1WaAEql32nmfv95BOItdI1Wrpg8Y_wBpf`
- Centered mount still Drive file ID: `1KeoOvA49vgokwxFsxmQRuBrQ0Dn4Iw4O`

## What was actually fixed

This is not the old server's invisible-mini-ghast carrier trick. The Spellbrook broom itself is the ridden entity and owns flight.

1. Removed item-frame `ItemDisplayContext.FIXED` behavior from the entity renderer. The captured FIXED transform rotated the visible broom 90 degrees and translated it away from the real vehicle/seat.
2. Removed double X/Y baked-model centering. Minecraft's ItemRenderer already centers the model; the custom renderer now compensates only the broom-specific Z seat anchor.
3. Corrected handle-forward / rider / brush-trailing orientation and a proper shaft seat pivot.
4. Reduced broom-only visual bank/pitch so the upright vanilla rider stays visibly attached during turns.
5. Replaced sideways A/D strafing with native steering/yaw, inertia, climb/descent, boost, hover/gravity behavior, collision response, and variant/core/mastery handling.
6. Fixed the major network fight: vanilla `LocalPlayer` was sending stale `ServerboundMoveVehiclePacket` positions because the broom inherited `isControlledByLocalInstance() == true`. Spellbrook brooms now return false there; custom broom input goes to the server, the server owns the physics, and ordinary entity tracking returns authoritative movement.

There is no ghast/invisible carrier reference in the broom entity or renderer.

## Native movement proof

Start:
- player `[0.5, 150.02, 0.5]`
- broom `[0.5, 150.0, 0.5]`

After climb + turn:
- player `[-5.698001137355398, 164.11380939276881, 16.66971541392481]`
- broom one command later while freewheeling `[-6.004866888530266, 164.09380942514485, 16.14111292422455]`

This proves roughly 14 blocks vertical and 18 blocks horizontal real translation.

Final recording seat/sync sample:
- player Pos `[3.644532069127696, 160.2474473174165, 0.9600400732332777]`
- broom Pos `[3.644532069127696, 160.22744731741648, 0.9600400732332777]`

Player and broom X/Z are exactly identical; Y differs by the intended 0.02 seat offset.

Server-observed final input sequence covered forward, climb, right turn, boost, left turn, release/freewheel.

## Gates

- changed-path compile: PASS
- real Forge integrated client: PASS (`Dev joined the game`)
- real world translation: PASS
- centered mount / no lateral seat drift: PASS
- early/mid/late native GIF frame audit: PASS
- clean client Save & Quit: PASS (`All dimensions are saved`)
- native task-owned error scan: 0 Spellbrook/model/texture/network failures
- dedicated Forge server: PASS, reached `Done (9.871s)!`
- dedicated Minecraft shutdown/save hooks: PASS, `All dimensions are saved`
- normal product `clean build`: PASS in 28s including `reobfJar`
- QA-only dependency alias omitted from shipping build
- final JAR archive integrity: PASS

The detached Gradle runServer wrapper records a nonzero task result only because the already-ready QA server JVM was deliberately SIGTERM'd after readiness; Minecraft itself ran its shutdown hooks and saved all dimensions cleanly.

Known harness-only warnings remain limited to offline Mojang/Yggdrasil/update-check and external vanilla sound-asset availability. No image generation was used; the published broom GIF/MP4 are real Minecraft runtime captures.
