# Backpacks 2.0.2.4 — Bedrock → Forge Semantic Parity Checkpoint

Date: 2026-09-09  
Target: Minecraft 1.20.1 / Forge 47.4.23 / Java 17  
Mod ID: `sqst_bkpk`

**Status:** static/semantic visual parity gate PASS; 2.0.2.4 native build/reobf/runtime re-certification is still pending.

## Durable artifacts

Google Drive checkpoint folder:
https://drive.google.com/drive/folders/123WbuBqMWfWQYRQmnpD7qtHP6ys9E8Gk

- Source checkpoint: https://drive.google.com/file/d/10MuS4lspQM-bOLdeHJNeJlYTl8XRsZc7/view
  - SHA-256 `3d483542366f78381f9cbad2d6307ee070da640c2e738c9f34c8fec0bd20c024`
- Evidence bundle: https://drive.google.com/file/d/132fkrQDtD8C1EQsdM407jys4RC1DLU_q/view
  - SHA-256 `d9254c5a4a82c2cf8f6fd6eec370f6008bc9485b6716db6e37effd036bacff8e`
- Semantic audit: https://drive.google.com/file/d/1_gFBF5ZM-qhFmV40YdDzR91RAN-3CNOc/view
- Basis before/after preview: https://drive.google.com/file/d/1VLecGNwvuLAF6SkDxZdOX0tiF3aViMAl/view
- Emissive semantic preview: https://drive.google.com/file/d/1yQ_ZBNYM9dutsQbtdBcJAM83ua2e04qf/view
- Fresh-extraction verification: https://drive.google.com/file/d/1hWF3WdIUtilK4ba6kbgUTrnAp5U-eWeG/view
- SHA-256 manifest: https://drive.google.com/file/d/10irWOF5tpFfRXi-Unruuz0YHIil-fSOV/view

Reusable skill packages:
- Minecraft Dev Kit skill: https://drive.google.com/file/d/1BNQSMLl127lGzD6g4HHsgqEqeITq61Q_/view
  - SHA-256 `8ae95dd902f718f3daad2036fd9edc3dd9383cd692d27ae60d179cc0d776c9e6`
- Minecraft Repair skill: https://drive.google.com/file/d/19XUzEvHGdFtDOUBPRm-ohLJqhSz_U4x5/view
  - SHA-256 `ef94afbd938c0efa0e6145bda26128ba57eeba084d8153fca3ab6115a8e4639b`

## What the exhaustive semantic audit found

1. Bedrock model basis must reflect X, not merely negate selected Euler angles:
   - point/pivot `(-x, y, z)`
   - cube X origin `-(origin.x + size.x)`
   - Euler rotation `(-rx, -ry, +rz)`
2. `entity_emissive_alpha` is a material semantic, not ordinary cutout transparency. The Java bridge now uses a translucent/lit base plus a deterministic full-bright mask derived from semi-alpha source pixels.
3. `*.highlight` geometry is a separate conditional full-bright shell driven by the original hammer/destroy-tool proximity condition, not permanent emissive geometry.
4. Converter parity must preserve `client_entity`, attachables, material declarations, behavior properties/controllers, and model/texture bytes together.
5. Bone `binding` is semantic. This source contains 63 binding declarations total; body/item-slot binding families must be explicitly discharged at Java render-adapter boundaries rather than silently ignored.
6. Emissive presence can vary by exact color state: Creeper Black and Kaiju Black are source exceptions, so emissive routing cannot be keyed only by design.

## Exhaustive gate

PASS from working tree and again from a fresh extraction of the packaged source ZIP:

- 31 designs × 17 colors = 527 visual states
- 124 design geometries
- 449 design bones
- 2,560 design cubes
- 9,504 per-face entries
- 1,131 cube pivots
- 1,131 cube rotations
- 1,184 inflate uses
- 426 `uv_rotation` faces
- 3,865 signed `uv_size` faces
- 62 design-model bindings + the held-item binding audited separately
- 323/527 visual states contain semi-alpha emissive pixels
- 43,707 semi-alpha emissive pixels covered
- source-absent feature classes explicitly checked: locators, `poly_mesh`, `texture_meshes`, mirror/reset, box UV, face `material_instance`

## Repair Brain / skill learning

Canonical Repair Brain `repair-history.jsonl` was appended in place with record:
`mc-bedrock-java-render-semantic-parity-backpacks-2.0.2.4-2026-09-09`.

The Minecraft Dev Kit and Minecraft Repair skill packages were updated to encode the semantic-sidecar, basis, asymmetric-sentinel, binding-discharge, material-alpha, conditional-highlight, exhaustive-state, and native-proof rules learned here.

## Explicit open parity gate

The original third-person **main-hand** attachable path uses dynamic `animation.sqst_bkpk.swing` driven by movement distance/speed. The current Java BEWLR path still approximates that with a static transform because it lacks holder/entity motion context. This is intentionally recorded as an open entity-aware render-adapter/native-moving-player gate and is **not** hidden under the static semantic PASS.

The certified 2.0.2.3 JAR remains the last native/reobfuscated green release until 2.0.2.4 is rebuilt with the exact ForgeGradle 8.8 / Forge 47.4.23 toolchain and its invalidated Bee/Warden/emissive visual contexts are retested in the packaged production client.
