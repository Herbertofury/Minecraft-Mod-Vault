# Spellbrook 0.2.0-dev — PRETTY FINAL v9 FIRE WYRM checkpoint

## Lineage
- Parent durable checkpoint: `c003ee0edb5ebe23fc1978c9ca2f8b10327721ea` (v8 spellcasting UX/audio)
- Canonical Google Drive folder: `1xA5xcdiMVD3kDo0UD5l4ouI_goWMsZFg`
- Canonical repository: `Herbertofury/Minecraft-Mod-Vault`
- Target: Minecraft Forge 1.20.1 / Forge 47.4.23 / Java 17

## User correction preserved
v9 does **not** ship Spellbrook's custom spell HUD/feedback overlay. UI/UX ownership is intentionally left to normal Minecraft / Iron's-style interfaces and other magic mods.

Removed from the shipping source/JAR:
- `client/SpellHudOverlay.java`
- `client/SpellHudState.java`
- `network/SpellCastFeedbackPacket.java`

Authored Spellbrook audio remains. Normal vanilla/system spell-selection text remains. The accepted v7 spell VFX baseline is otherwise preserved unless Fire Wyrm required a targeted extension.

## Objective completed
Add Spellbrook's flagship Inferno Dragon / Fire Wyrm as a complete 30th spell and raise its native model/animation/impact presentation to AAA quality without generated imagery or replacement geometry.

Public benchmark used during the pass: Spellbrook Wiki's Mythical Fire spell **Inferno Dragon**, level 30, mana 80, described as conjuring a blazing dragon to wreak scorching destruction in the target direction. Public sources did not expose authoritative hidden animation curves, so animation fidelity is grounded in the shipped authored Spellbrook model/texture plus native Minecraft visual QA rather than claimed exact hidden-source parity.

## Fire Wyrm gameplay identity
`SpellDefinition.FIRE_WYRM("fire_wyrm", "fire", 80, 180, 18, Kind.PROJECTILE, "vfx_me", "vfx_inferno_dragon")`

Alias: `inferno_dragon` -> `FIRE_WYRM`.

Behavior:
- no-gravity flagship dragon projectile
- approximately 0.68 flight speed
- penetrates living targets instead of vanishing on the first hit
- per-target hit cooldown: 10 ticks
- each body pass deals about 52% effective spell power through the existing indirect-magic damage route
- burns struck targets for 6 seconds
- pushes targets along flight direction
- terrain contact or 50-tick travel cap triggers the final detonation
- strong splash + authored impact/aftermath phases
- never edits terrain blocks
- cleans every temporary manifestation/projectile after completion

## Authored Inferno Dragon source
Original model:
`assets/spellbrook/models/vfx_me/vfx_inferno_dragon.json`

Original texture:
`assets/spellbrook/textures/modelengine/entity/models/vfx_dragonshot.png`

Model facts:
- 135 authored elements
- 54 Euler-element transforms
- custom loader: `spellbrook:euler_elements`
- original model SHA-256: `abfa806b53d5c0d26309f942a910b20217309d664a4ca0d2da8abab0ec9d4aad`
- texture SHA-256: `e9a3c71a7d4ee15e8b003f17aecb1bc60db98c3f85bc208c43285c3308e7b52f`

## Lossless runtime articulation
The original authored dragon model was partitioned into six runtime render slices:
- `fire_wyrm_aura.json`
- `fire_wyrm_body_0.json`
- `fire_wyrm_body_1.json`
- `fire_wyrm_body_2.json`
- `fire_wyrm_body_3.json`
- `fire_wyrm_body_4.json`

New Java compositor:
- `client/FireWyrmRenderModels.java`

`ClientModEvents` registers all six additional baked model locations and `SpellProjectileRenderer` composes them in flight.

Lossless model-partition audit:
- original elements: 135
- slice element total: 135
- exact element multiset equality: PASS
- original Euler transforms: 54
- slice Euler transforms: 54
- all slices keep `spellbrook:euler_elements`
- all slices use the original `vfx_dragonshot` texture

No authored geometry was dropped and no generated/replacement art was introduced.

## AAA motion pass
Native QA exposed two important defects in earlier candidates and they were corrected instead of accepted:

1. The first articulated pass had wave amplitude backwards, making the head flex more than the tail. Final motion stabilizes the head and progressively increases delayed motion through neck/body/tail.
2. The authored dragon's head faces local `-Z`, while the generic projectile alignment assumed local `+Z`. That caused the early candidate to fly tail-first. Final Fire Wyrm applies the required local 180-degree Y correction after motion alignment, so the head leads.

Final head-to-tail motion uses progressively stronger yaw/pitch/roll amplitudes with phase delay across five body slices. Global banking remains subtle, birth scale grows in smoothly, and the inferno halo rotates independently.

The halo begins larger as summoning energy, then settles behind/supporting the dragon instead of dominating the silhouette as a permanent flaming wheel.

## Mythical impact finish
The final impact combines only existing authored Spellbrook effects:
- primary: full 12-stage `vfx_phoenix_explosion_1..12`
- secondary: dynamic 9-stage `vfx_phoenix_blaze_wave_1..9`
- bounded flame/smoke/lava choreography around the impact point

The final challenge pass confirmed the stronger finish does not obscure targets or flood the frame.

## Exact production delta from v8
23 `src/main` paths differ.

Modified Java:
- `client/ClientModEvents.java`
- `client/SpellCameraEffects.java`
- `client/SpellProjectileRenderer.java`
- `client/SpellVisualProfile.java`
- `entity/SpellProjectileEntity.java`
- `event/CommonEvents.java`
- `item/VesselItem.java`
- `magic/MagicProgression.java`
- `magic/SpellDefinition.java`
- `magic/SpellVisualParticles.java`
- `magic/SpellVisualPhase.java`
- `magic/Spellcasting.java`
- `network/SpellbrookNetwork.java`

New Java:
- `client/FireWyrmRenderModels.java`

New resources:
- six articulated Fire Wyrm model JSON slices listed above

Removed from v8:
- `client/SpellHudOverlay.java`
- `client/SpellHudState.java`
- `network/SpellCastFeedbackPacket.java`

`SpellAudio` is unchanged from v8. `sounds.json` and language audio metadata are unchanged from v8.

## Native Forge proof
The final r4 candidate ran in the real Forge 47.4.23 client + integrated server using Java 17 and the established Linux/Xvfb/Mesa QA path.

Authoritative final markers:
- `[WYRMQA] PREPARED player=Dev first=true second=true wallZ=24`
- `[WYRMQA] CLIENT_READY camera=FIRST_PERSON gui=false`
- `[WYRMQA] CAST cast=true projectileCount=1 noGravity=true manaAfter=310.0 definitionCount=30`
- `[WYRMQA] PASS FIRST_TARGET health=490.64 fireTicks=120 projectileAlive=true`
- `[WYRMQA] PASS SECOND_TARGET health=490.64 fireTicks=120 projectileAlive=true`
- `[WYRMQA] PASS IMPACT impact=1 aftermath=0`
- `[WYRMQA] PASS AFTERMATH impact=0 aftermath=1`
- `[WYRMQA] PASS CLEANUP projectiles=0 impact=0 aftermath=0 cast=0`
- `[WYRMQA] COMPLETE pass=true projectile=true first=true second=true impact=true aftermath=true cleanup=true finalProjectiles=0 finalManifests=0`
- `[WYRMQA] CLIENT_STOP clean=true`

Integrated server close:
`All dimensions are saved`

Final accepted native run wrapper:
`BUILD SUCCESSFUL in 48s`

Task-owned model/exception scan: 0. Remaining network/auth/icon warnings are the known sandbox-offline Mojang warnings, not Spellbrook runtime failures.

## Shipping gates
Temporary `qa/FireWyrmQaDriver.java` and `qa/FireWyrmQaClient.java` were copied into evidence, then deleted from the shipping source before the clean build.

Final shipping build:
- `clean build`: BUILD SUCCESSFUL in 20s
- `reobfJar`: executed normally
- final JAR size: 6,991,347 bytes
- final JAR SHA-256: `1f12e26819e44d26aafeb174844eee66d14ec4abee89f38ee3e38e2d184acf3c`
- JAR entries: 3,518
- class files: 125, all Java class major 61
- packaged JSON: 1,441 / 0 parse errors
- spell definitions: 30 / 30 unique
- `fire_wyrm`: exactly 1
- `inferno_dragon` alias: present
- QA class entries: 0
- `WYRMQA` marker bytes: 0
- custom HUD/feedback packet class entries: 0
- archive signatures: 0
- JAR ZIP integrity: PASS

Fresh source archive proof:
- extracted files: 3,537
- QA source files: 0
- fresh-copy `compileJava`: BUILD SUCCESSFUL in 13s

## Final artifacts / hashes
- JAR: `1f12e26819e44d26aafeb174844eee66d14ec4abee89f38ee3e38e2d184acf3c` — 6,991,347 bytes
- source ZIP: `2fdbbdd5fc3ce86e344969eb64b998542b3bc0267e4db6a58ca7c1697cc344b0` — 22,534,953 bytes
- native evidence ZIP: `66ef66d1d446ecf8d44644775ce85a6e22b0ebd5b8e6663158fe2c61b06d9684` — 13,289,963 bytes
- release bundle ZIP: `4dc5645a02fe4704614f58aa2f0f15b8b185ee2bba01e3cf418c994ddce1a247` — 49,283,782 bytes
- verification report: `3a2ba5bdc79f5fc7c2402419305b6d401f5052a2d6d3961b77ccef24b7783d98`
- exact v8 -> v9 patch: `1ee708a2cffac29a63c908cf9c758c7cf267a2e79a6bf20be02b615dca700b53`

## Canonical Google Drive objects
- installable JAR: `1uORkALOUo94y31HNFEvRt9FSclR9HSBD`
- source ZIP: `1HauaK4QYzhpgN0c-zOdC0w7kWx398-Sy`
- native evidence ZIP: `1U3g7qdTsvM47c11eo0N5Rp4zwwNCm6UJ`
- release bundle ZIP: `1KVJtN21CgByOtkwNqd6FhqZrJ3y5sVvF`
- verification report: `1-u1GVRrZA7QHTlmqi4ZZSM80NTjj-g3A`
- exact v8 -> v9 patch: `16exEYYId-vi3rxPMzdZUQXYL-ZPqcWpV`
- native contact sheet: `1-EAdsCRQkYvH0LEkQV2SWhRM12ICFCju`
- native highlight GIF: `1p5kIxH7ZmGl8mBJOVf0IdQwMI9CIYAbz`
- native highlight MP4: `15nm2lsZ3Mk7YfrVYWlTEYcYvr2uQY4MO`
- core SHA manifest: `1baHmZlD87UsSeYWvtqa6-foA22bO4Q5g`
- bundle SHA manifest: `1zsxA6ceMWXm5tk4V-FbWsK0VU7fy-BM5`
- content audit: `1_bnZNXeCiAlggiiMhMzsAYAc01acbpyZ`

Post-upload byte verification was performed by materializing the five core Drive objects back into the runtime and comparing them with the frozen local release objects. SHA-256, size, and `cmp` all match for JAR/source/evidence/bundle/report: `REMOTE_BYTE_IDENTITY=PASS`.

## No-image-generation statement
No generated imagery was used. Model work is derived from the real authored Spellbrook Inferno Dragon JSON/texture. All showcase PNG/GIF/MP4 outputs are deterministic derivatives of the accepted real Minecraft capture.

## Exact next action
v9 is the new flagship-spell baseline. Preserve the accepted Fire Wyrm implementation unless a real in-game regression invalidates it. Future AAA work should apply the same standard to other high-value spells: authored-model-preserving articulation, stronger motion language, premium impact staging, and native Forge evidence. Do not reintroduce Spellbrook's custom HUD/UX layer.