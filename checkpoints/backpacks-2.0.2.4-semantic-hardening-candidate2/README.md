# Backpacks 2.0.2.4 Semantic Hardening Candidate 2

Date: 2026-09-10
Target: Minecraft 1.20.1 / Forge 47.4.23 / Java 17
Mod id: `sqst_bkpk`
Status: **BUILD + STATIC/SEMANTIC + PACKAGED PRODUCTION SERVER GREEN; CURRENT PRODUCTION CLIENT QA PENDING**

## Exact candidate

- JAR: `Backpacks-2.0.2.4-SEMANTIC-HARDENING-CANDIDATE2-Forge-1.20.1.jar`
- SHA-256: `5e0c1c7d5441755e7b67ba11b6515ef9a3e3ded19a57cfbfc231d0767728a3b7`
- Size: 2,906,093 bytes
- Source ZIP SHA-256: `eb21e8173d5c763f51ec1a8a086a4e58e24e172fc06f4a79c7fd7212c9e7cc92`
- Evidence ZIP SHA-256: `418f88783323a0234f59e45d4c2297590dbdf979233c01a41bca72322f5b1b4c`

The last fully native-certified release remains 2.0.2.3 (`89fb25992e4cc90e4afa6aab402619144dfd68aed8ff268af9e88a7fcab73eca`). Candidate 2 must not be promoted to FINAL until the current packaged production-client runtime gates complete.

## Fresh build proof

Exact restored toolchain:
- Eclipse Temurin Java 17.0.20.x
- Gradle 8.8
- Forge 47.4.23 populated offline cache

`clean build` completed with exit code 0 and `BUILD SUCCESSFUL`, including `reobfJar`. The 11 compiler warnings are deprecated/removal API notices only; there are no compile errors.

## Semantic closure

Current fail-closed gates all pass:
- 1,014 explicit source-to-target semantic obligations; zero uncovered/unresolved/proof failures.
- 694 declared Bedrock identifiers explicitly covered.
- 37/37 source recipes mapped to 40 required Java crafting lanes with zero payload mismatches.
- 527/527 item textures and 527/527 entity textures preserved, plus all 7 sidecar item contracts.
- 31 designs x 17 colors = 527 visual states complete.
- Behavior-family split preserved: designs 0-14 legacy parrot/rider generation; designs 15-30 modern generation.
- Legacy-wear regression: 7/7 checks pass, including the source `swim+glide=3 -> riding` fallback and no Parrot-only transform leakage into normal wear.
- Geometry/material semantic audit: 153/153 checks pass; zero warnings/errors.
- Emissive-alpha state audit: 323/527 states contain authored semi-alpha emissive data, 43,707 semi-alpha pixels total.

## Material repairs in this hardening line

- Correct full Bedrock X-basis conversion for pivots, origins and Euler rotations; asymmetric Bee/Warden geometry is no longer falsely mirrored.
- Preserve per-state `entity_emissive_alpha` semantics using derived fullbright masks rather than cutout-only rendering.
- Keep source highlight geometry conditional on hammer proximity instead of permanently emissive.
- Preserve/discharge Bedrock `binding` semantics explicitly at Java render-context adapters.
- Restore authored Backpack Crafter geometry, item/block renderer, four cardinal transforms, material/sound/mining semantics and explicit explosion immunity.
- Restore Guide version/lore/page behavior and Locator click-toggle/cross-dimension source semantics.
- Restore source-aware third-person held swing context rather than a static pose approximation.
- Override Creative placement so the placed backpack is a fresh empty storage object while the Creative-held original remains intact, preventing dependency-level content cloning from violating the Bedrock anti-duplication intent.
- Preserve destroy-tool `hand_equipped` presentation.
- Make recipe repair explicit for the three malformed source shaped recipes; distinguish edition aliases from actual semantic broadening; design 30 is narrowed from a wool tag to source-equivalent white wool.
- Separate legacy Parrot Behaviour from normal 2.0 Curios wear; source-relative legacy poses load from preserved Bedrock animation files only when legacy mode is active.
- Add Platform 1.3.4 to the real Vanilla Backport runtime dependency closure.

## Reusable converter / skill proof

Updated Minecraft Dev Kit package SHA-256:
`ca763d45009a2dd24e916102ffda740716ff4c50ce213cdbd3a0be72bd331297`

Updated Minecraft Repair package SHA-256:
`448e13ffcc86a65de755565c983bfd20bb89479db7470f0fb9782f8b815c88dd`

The Dev Kit now includes a semantic IR extractor, fail-closed coverage verifier, self-test, and conversion reference. The tools passed their synthetic self-test and then passed against the real Backpacks source corpus: 2,363 source files, 1,389 semantic source objects, 694 declared identifiers, 3 explicitly classified source anomalies, and all 1,014 conversion obligations covered.

These ZIPs are validated distributable skill updates. They do not imply that ChatGPT's installed immutable global skill registry was modified in place.

## Production runtime state

Mapped Forge userdev `runServer` is not accepted as authoritative for this candidate because Platform ships production-name Mixin/refmap targets and fails in that mapped harness. The authoritative server gate was therefore rerun in an official installer-produced Forge 47.4.23 packaged runtime.

- packaged-server workflow commit: `b4a1ce1fb10e7ad7e7551015d08401f6dc0dca80`
- workflow run: `34519951229` — **SUCCESS**
- local packaged `forgeserver` with exact Candidate 2: **PASS**
- server ready marker: `Done (3.614s)!`
- graceful shutdown: `All dimensions are saved`
- candidate-owned severe/linkage failures: **0**

Current client-runtime reconstruction branch: `qa/backpacks-2.0.2.4-production-client`.
The production-client exporter must reproduce the previously certified client SRG/client-extra/Forge client-patch/Forge universal hashes plus the complete Minecraft 1.20.1 asset index before it is accepted.

Remaining runtime acceptance:
1. Run targeted packaged `forgeclient` + integrated-server QA for the invalidated paths: Bee/Warden geometry, emissive states/exceptions, highlight OFF/ON, crafter cardinal model + UI, Locator toggle/cross-dimension arrow, Guide behavior, Creative fresh-empty placement, moving third-person hand swing, normal Curios wear, and the 0-14 vs 15-30 legacy boundary.
2. Only after those gates pass may 2.0.2.4 be called FINAL/native-certified.
