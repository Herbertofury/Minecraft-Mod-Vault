# Macabre 0.9.2 Eye UI + Performance Fix — Replacement 2

Minecraft 1.20.1 / Forge 47.x / Java 17

Replacement 2 repairs the full-pack startup failure observed on 2026-09-12 with Replacement 1:

`InvalidMixinException: The mixin 'com.curseforge.macabrefix.eyeui.mixin.EyeButtonMixin' is missing an @Mixin annotation`

## Root cause

Replacement 1's compile-only Sponge stubs declared `@Mixin` and `@Pseudo` with `RetentionPolicy.RUNTIME`, placing those class markers in `RuntimeVisibleAnnotations`. Replacement 2 changes only the compile-only retention declarations to `RetentionPolicy.CLASS`, recompiles `EyeButtonMixin.class`, and stores the class markers in `RuntimeInvisibleAnnotations` where the installed Mixin path recognizes them.

## Scope proof

- Replacement 1 -> Replacement 2 changes exactly one JAR entry: `com/curseforge/macabrefix/eyeui/mixin/EyeButtonMixin.class`.
- 7,779 other entries are content-identical.
- No entries added or removed.
- No metadata-only changes.
- `META-INF/mods.toml`, `META-INF/MANIFEST.MF`, and `macabre_eye_ui_fix.mixins.json` are byte-identical.
- `javap -p -c -s` method bytecode is identical before/after.
- All 2,637 classes remain Java 17 class-major 61.
- Internal loader identity remains `macabre` / `0.9.1`.

## Repair Mark v2 closeout

Replacement 2 now carries forward the validated Macabre Repair Mark v2 as a mandatory sidecar/launcher mapping.

- Macabre has no loader `logoFile`; this is now explicitly treated as **sidecar required**, never as a reason to defer the Repair Mark.
- Exact upstream-authored art basis: `macabre_icon.png` from the exact official Forge 1.20.1 Macabre JAR.
- CurseForge project ID: `933001`; file ID: `8362844`.
- Source art: 256x256, SHA-256 `1c9430772b925fa444ef7bfa243d0379a14a1d47c9fca08845655319d4024a65`.
- Repair Mark PNG SHA-256: `36d1eea5713ec7658fb9af4db88cc1032b5d4a7d6ab3d63da808ae3b88abc2fc`.
- 48x48 QA preview SHA-256: `0d924cfd122b3d1d5f6e74d55a442907cc56b51734330b3984daa310b317b498`.
- Deterministic re-render reproduced the Repair Mark PNG byte-for-byte from the official-JAR source art.
- Marker-only diff: PASS — visual identity changes zero bytes in the gameplay JAR.
- Badge registry is rebound to exact Replacement 2 JAR SHA-256 `d826b4b1c0e2b82255efed70e88c844614fdf6401754f9641520859376c64caa`.

## Repair workflow hardening

The Minecraft Repair skill is hardened so this regression cannot silently close again:

1. Every shippable repaired JAR must end in exactly one valid visual state: `embedded` or `sidecar/launcher-mapping`.
2. Missing `logoFile`, signed/unsafe icon mutation, or launcher-external artwork force sidecar integration; they never mean defer/skip.
3. Superseding/rebuilt repairs must carry forward or rebind any previously validated Repair Mark unless official upstream artwork changed.
4. New `scripts/repair_mark_gate.py` blocks closeout if a repaired JAR record lacks Repair Mark provenance/artifacts.
5. Regression proof: the gate FAILS the original Replacement 2 record and PASSES the corrected record.
6. Updated skill bundle validates and packages successfully; SHA-256 `8b3a2430bd3d18242cf227938259a28ae03fad6385d7f9e416127ab4649ba274`.

## Artifacts

- Replacement 2 JAR — SHA-256 `d826b4b1c0e2b82255efed70e88c844614fdf6401754f9641520859376c64caa` — https://drive.google.com/file/d/14wsXdqjFd0dTLf8v1gbYZihtg-8Qch7O/view
- Reproducible source ZIP — SHA-256 `ade8b0303e5b120438a373eff219f560fedaea0eb97d43133f2da2eb9234b610` — https://drive.google.com/file/d/1m8GNHVsEleRcqw3qWudydiSNpQYEuwf0/view
- Final verification report — SHA-256 `54c7d457ab6e9f4b004fdd3b1100266a6c36a38d93ab1b7f85f2d2c68a0d5b51` — https://drive.google.com/file/d/1KBUTJOj7fEkHOcbKII0wSAZsRi40iFV2/view
- Repair Mark v2 sidecar — SHA-256 `36d1eea5713ec7658fb9af4db88cc1032b5d4a7d6ab3d63da808ae3b88abc2fc` — https://drive.google.com/file/d/1KmFKRTxqCU2BkQx7kqYpAR35E646i3g-/view
- Repair Mark 48x48 QA preview — SHA-256 `0d924cfd122b3d1d5f6e74d55a442907cc56b51734330b3984daa310b317b498` — https://drive.google.com/file/d/1dVEvil8h0enhmyQhvbQfhV9_Uyt-Kl4W/view
- Replacement 2 badge registry — SHA-256 `4497a2aa4b2466e1869e4f1a5dfe1b3dfea61590803e163f4ebc8cf926270f74` — https://drive.google.com/file/d/1apJODMluF5TUfhBDkFpwa0sZAeacOcz4/view
- Repair Mark closeout record — https://drive.google.com/file/d/1TKt0ZDA8X11umxKtUr0727Q-Wf7eEFdW/view
- Updated Minecraft Repair `skill.zip` — SHA-256 `8b3a2430bd3d18242cf227938259a28ae03fad6385d7f9e416127ab4649ba274` — https://drive.google.com/file/d/1UGzxqh2isfM4Cjxwm8HVNo2HhIVPH0H2/view
- Canonical Repair Brain — SHA-256 `c89824f514807b7800dfaa3e5cc53e335606a57926d34bb5ffd121c47cbc40d2` — https://drive.google.com/file/d/17nwD2w1_q2suLAMyLQgVLL1ZYPPqFAfw/view

All newly published Drive artifacts and the canonical Repair Brain were downloaded after publication and matched local bytes exactly.

## Remaining runtime gate

Install Replacement 2 in the exact Noxviola instance in place of Replacement 1 and confirm normal startup, Inventory open, hidden Eye launcher default, ALT+E Eye Inventory, ALT+M editor, and neighboring inventory UI behavior.

GitHub's connected app in this session exposes repository text writes but no binary release-asset upload action, so the verified JAR, source, image sidecars, and skill ZIP remain on Drive while this repository records immutable hashes and recovery links.
