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

## Artifacts

- Replacement 2 JAR — SHA-256 `d826b4b1c0e2b82255efed70e88c844614fdf6401754f9641520859376c64caa` — https://drive.google.com/file/d/14wsXdqjFd0dTLf8v1gbYZihtg-8Qch7O/view
- Reproducible source ZIP — SHA-256 `ade8b0303e5b120438a373eff219f560fedaea0eb97d43133f2da2eb9234b610` — https://drive.google.com/file/d/1m8GNHVsEleRcqw3qWudydiSNpQYEuwf0/view
- Verification report — SHA-256 `5cc6f466313daf8233b26239aef1254ae94ecd50f6e381d39c51dd6830d8cb1b` — https://drive.google.com/file/d/1_uVD_h1uni6HIuLPoMMOEioezW6KdFuu/view
- Canonical Repair Brain — https://drive.google.com/file/d/17nwD2w1_q2suLAMyLQgVLL1ZYPPqFAfw/view

The Drive artifacts were downloaded/materialized after publication and matched the local bytes exactly.

## Remaining gate

Install Replacement 2 in the exact Noxviola instance in place of Replacement 1 and confirm normal startup, Inventory open, hidden Eye launcher default, ALT+E Eye Inventory, ALT+M editor, and neighboring inventory UI behavior.

GitHub's connected app in this session exposes repository text writes but no binary release-asset upload action, so the verified JAR and source ZIP remain on the canonical Drive publication while this repository records the immutable hashes and recovery links.
