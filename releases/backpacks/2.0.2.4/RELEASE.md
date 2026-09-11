# Backpacks 2.0.2.4 — Forge 1.20.1 Native-Certified Final

Target: Minecraft 1.20.1 / Forge 47.4.23 / Java 17.0.20.1

## Final artifact identity

- `Backpacks-2.0.2.4-FINAL-Forge-1.20.1.jar`
  - SHA-256: `252e83b9f7fbe20fd15a4de047d4b02f1abcf87852dbbe7944beb386a85b1342`
  - Size: 2,906,589 bytes
- `Backpacks-2.0.2.4-FINAL-SOURCE.zip`
  - SHA-256: `d5798b9031fcf69412d1d1c08ce1a0f562b8cdd68465c11b92cd0327549c2052`
- `Backpacks-2.0.2.4-FINAL-EVIDENCE.zip`
  - SHA-256: `419e173a43af3b1a0559e179b021d54032580b51b9b5961b27a1dc0c8556e424`
- `Backpacks-2.0.2.4-FINAL-VERIFICATION.md`
  - SHA-256: `e801ee68f1a3f3d119a8192f8ec1e1fb3a8ef17168d52468fc7ed69241e1a596`

Previous native-certified baseline: 2.0.2.3 (`89fb25992e4cc90e4afa6aab402619144dfd68aed8ff268af9e88a7fcab73eca`).

## Acceptance closure

- 156/156 semantic checks pass.
- 31 designs × 17 colors = 527 visual states complete.
- 323 emissive state files / 43,707 semi-alpha emissive pixels audited.
- Fail-closed semantic coverage retains 1,014 source-to-target obligations with zero uncovered/unresolved/proof failures.
- 2.0.2.3 → 2.0.2.4 baseline-delta policy is versioned and passes; intentional preserved-source/emissive families and the explicit design-30 recipe correction are classified instead of bypassed.
- Exact Candidate5 JAR launches in official Forge 47.4.23 production client, loads with the exact dependency set, crafts the Warden backpack through the real Backpack Crafting Table, and produces zero `Unknown recipe category: sqst_bkpk:backpack_crafting` warnings.
- Exact Candidate5 JAR reaches `Done (...)!` in the official Forge 47.4.23 packaged dedicated server, then exits cleanly with `All dimensions are saved` and zero Candidate5-owned severe errors.
- Candidate4 native visual/gameplay proofs remain applicable because Candidate5 changed only client recipe-book category registration plus QA tooling: old-save crafter relight migration, Creative fresh-empty placement, Locator ON/OFF/cross-dimension/session persistence boundary, held movement animation, Curios/legacy wear, Bee/Warden/Creaking placed/held/worn emissive behavior, and Hammer highlight OFF/ON.

## Durable Google Drive artifacts

Canonical folder: `Backpacks 2.0.2 Java Port/Backpacks 2.0.2.4 Semantic Parity Checkpoint`

- Final JAR: https://drive.google.com/file/d/1_oE_tYL_7cadCI5U4bLQH7HYTAJGw6xK/view
- Final source: https://drive.google.com/file/d/1ACuOKU8qX-8JD3zJKCkpz0c6PndqabS3/view
- Final evidence: https://drive.google.com/file/d/1120OF8HwosYujemFwcrQ_L72SEbUlKlM/view
- Final verification: https://drive.google.com/file/d/1-v5sXce7qw0H9fSg2ChQ75u8zaJckRgO/view
- SHA-256 manifest: https://drive.google.com/file/d/1SNroNUprkvHuegMvlOtFd8RgavxeTNMA/view
- Updated Minecraft Dev Kit skill: https://drive.google.com/file/d/1rDdC50LAevaPyq-I9LILi5WS6Hi8hAUR/view
- Updated Minecraft Repair skill: https://drive.google.com/file/d/1sV0W-u7cGC6lFndH8J8spMUV9lV46Zp9/view

All uploaded Drive bytes were downloaded again and byte-compared/hash-verified against the frozen local release artifacts.

## Reusable Bedrock→Java improvements learned from this release

1. Open Bedrock custom block geometry must carry explicit Java occlusion/light semantics; a correct model/texture is not enough.
2. When an earlier build registered the block as opaque, old saved chunks require upgrade-safe light invalidation; this release proves an untouched Candidate2-saved block self-heals under the fixed code.
3. A custom Java `RecipeType` can craft correctly but still be incomplete on the Forge client recipe-book path; explicit category registration plus zero-warning native validation is required.

The same rules are now in the packaged Minecraft Dev Kit / Repair skill updates and the canonical Repair Brain `repair-history.jsonl`.

## GitHub publication note

This repository record is the GitHub durable index for the release. The connected GitHub tool in this session supports repository/branch/text mutations but does not expose binary release-asset upload. Therefore the complete binary artifacts are hash-verified on Google Drive and linked above; no false claim is made that the binary files were uploaded to GitHub.

## Post-release latest Sophisticated compatibility

The frozen 2.0.2.4 JAR has now also passed a fresh packaged-client + dedicated-server compatibility challenge against **Sophisticated Backpacks 3.26.3.2157** and **Sophisticated Core 1.5.1.2335**, using Forge 47.4.23, Temurin 17.0.20.1+1, Curios 5.14.1, VanillaBackport 1.1.7.10, and Platform 1.3.4.

- No compatibility patch was required; the release JAR remains byte-identical at SHA-256 `252e83b9f7fbe20fd15a4de047d4b02f1abcf87852dbbe7944beb386a85b1342`.
- Native server boots reached `Done (...)!` and shut down with `All dimensions are saved`.
- The production Forge client joined the real server, opened the Guide UI, and rendered the Warden backpack live in Curios at midnight with emissive geometry active.
- Strict scan remained at zero for `NoSuchMethodError`, `NoSuchFieldError`, `AbstractMethodError`, `VerifyError`, `LinkageError`, `ClassCastException`, the custom recipe-category warning, and candidate-owned `sqst_bkpk` errors.
- GitHub certificate: `LATEST-SOPHISTICATED-COMPATIBILITY.md` in this release directory.
- Full compatibility evidence: https://drive.google.com/file/d/1iCEgDhxdYyyOjUPVADrj-q5yKrO_6vM1/view
- Compatibility checksums: https://drive.google.com/file/d/1zuino0t7e-Iq1xHwhNcsqSpH7kUlJe1y/view
