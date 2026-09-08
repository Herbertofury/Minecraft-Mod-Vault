# Backpacks 2.0.2 Java Port — Final QA Checkpoint

Date: 2026-09-08
Target: Minecraft Forge 1.20.1 / Forge 47.4.23
Mod id: `sqst_bkpk`
Candidate: `Backpacks-2.0.2-Java-2.0.2.1-FINAL-Forge-1.20.1.jar`
SHA-256: `fc2d6221cc219757ef3945191c7722312d42fd76b01567b013b78e058b69b84c`
Size: 1,605,932 bytes

## Durable artifacts

Authoritative binary/source/evidence storage is Google Drive folder `Backpacks 2.0.2 Java Port`:
https://drive.google.com/drive/folders/1qe5T_3Pr864C2yMcmngyJZi1FqxcVI5k

Drive file IDs:
- Final JAR: `1bybO71tC8ZohNMzcWoC8cfxg2GXCTDAK`
- Final source checkpoint: `1yf0REGM_36She7ovZMTatAk_QdhvUrNf`
- Final evidence bundle: `1xsGXc6iBiKMTsfzriSEDJNyFRaVn_nBW`
- Final QA checkpoint: `1hQEfdnRE7_u3UEibLGExnXu9trHxzQwb`
- SHA-256 manifest: `11b-zl1LRl3ztpclSUoeN1v9CtvSwIbQj`

This repository stores the durable GitHub provenance/checkpoint. The current connector does not expose release-asset upload or repository creation, so large binary/source archives remain in the canonical Drive folder instead of being forced into unrelated Git history.

## Release lineage

The final runnable JAR is byte-identical to the previously native-green Camping candidate. No gameplay/source mutation was made after that candidate; this closeout added stronger production-runtime QA for placement, storage, Curios, dye/wash, and hammer behavior.

Key reconstructed source hashes:
- `ScaiCurioRenderer.java`: `2e5a23ee702e1c19b6afe0d2c95227859a7700f878e6b4de5f505d762d751d0e`
- `BedrockModelRenderer.java`: `1071cef7f71189b459e953328cd2ef32ed97590078209abcccdfee9270bcdca9`

Direct source compile cross-check: 51 class files, Java 17 bytecode target (`javac --release 17`), 0 errors, 7 deprecation/removal warnings.

## Production runtime

QA ran in the production Forge `forgeclient` namespace, not mapped userdev. The harness reused the exact historical production SRG client and Forge patch payloads from the verified offline cache.

Known exact payload hashes:
- Minecraft 1.20.1 production SRG client: `e7ed6e44b4181da5770b5f4aa06e908e5008068cddc84446913ccdc2f435c903`
- client-extra: `f43d1ac034a8907edab190a17501fd4ca2b0af107d4231da364eeca8ab5ee317`
- Forge 47.4.23 client patch/binpatched payload: `ad9f1da4f4ae6121c3fd4ec6a67b0e89bb7466d7039879b75d7cb9592c0e5605`
- Forge 47.4.23 universal: `98a6c0a2318c1788db6f592357eb00d09b15316f054942ba432f71027f074fef`

The closeout host had OpenJDK 21 only. The exact release JAR already had earlier native production Java 17 proof for startup/client/render/Camping. The remaining gameplay gates below were rerun on the exact same production Forge namespace/JAR under Java 21; source was also compiled to Java 17 bytecode target. This host-JDK difference is recorded as a harness limitation, not hidden.

## Native QA acceptance matrix

- PASS — Production Forge main menu / mod loading.
- PASS — Integrated `Backpacks QA` world start and player join.
- PASS — Camping visual/animation matrix from prior checkpoint: walk, turning inertia, riding, water/swim, Elytra/glide.
- PASS — Correct Sophisticated placement behavior: crouch + use places the backpack block.
- PASS — Placed backpack opens the real Sophisticated 27-slot UI.
- PASS — Loaded deterministic storage: 7 diamonds, 13 gold ingots, 21 emeralds.
- PASS — Immediate close/reopen persistence.
- PASS — Full integrated-server stop/start persistence; exact 7/13/21 stacks remained.
- PASS — Normal pickup preserved full backpack NBT; reopening the picked item retained exact 7/13/21.
- PASS — Curios `back` slot swap with a distinct Duck backpack: placed Simple -> Duck and original Simple moved to the worn Curios slot.
- PASS — Native third-person showed the original Simple backpack worn on QAPlayer's back.
- PASS — Curios round-trip preserved full storage NBT: swapping back restored Simple as the placed block and reopening still showed exact 7/13/21.
- PASS — Blue dye visibly recolored the placed backpack.
- PASS — Water bucket visibly reset the placed backpack to source/default brown.
- PASS — Dye/wash mutations preserved exact 7/13/21 storage.
- PASS — Hammer destructive reset removed the block and returned/spilled the exact 7 diamonds + 13 gold + 21 emeralds + backpack.
- PASS — Colored hammer reset: blue placed backpack -> hammer -> block removed -> default brown backpack returned.
- PASS — Final graceful world save: integrated server stopped and logged `All dimensions are saved`.

## Log audit

No `sqst_bkpk` / Backpacks renderer, storage, Curios-swap, dye/wash, or hammer exception occurred during the successful gameplay gates.

Classified harness/non-task warnings:
- Optional VanillaBackport compatibility mixins reference absent Moonlight/JER/MouseTweaks classes; they remained warnings and did not affect the tested mod path.
- Mojang authentication/version-check DNS is unavailable in the sandbox.
- No OpenAL audio device is exposed by the headless Xvfb environment, so sound was disabled.
- Minimal external asset objects caused missing sound/icon warnings; mod/client textures used for the visual gates rendered correctly.
- `sophisticatedcore-common.toml` produced one early duplicate `[common]` watcher parse exception during first-time config correction. Forge/Sophisticated self-corrected it at 18:10:24; it did not recur during subsequent world starts, storage restart, Curios, dye/wash, or hammer tests.
- Java 21 produced the expected LWJGL old-JNI compatibility warning; the production client remained stable through the complete gameplay sequence. Target runtime remains Java 17.

## Release decision

GREEN for the tested Forge 1.20.1 / Sophisticated Backpacks / Curios integration scope. The canonical release artifact remains the exact `fc2d6221...` JAR above; no content, storage, visual, or behavior downgrade was introduced during closeout.
