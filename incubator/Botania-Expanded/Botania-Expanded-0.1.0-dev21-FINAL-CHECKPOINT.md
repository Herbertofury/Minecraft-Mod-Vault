# Botania Expanded — Endoflame Family Final Checkpoint

Version: `0.1.0-dev21`  
Target: Minecraft 1.20.1 / Forge 47.4.23 / Java 17  
Status: Endoflame-family approval candidate; implementation, build, packaged-client boot, integrated-world render, and gameplay regression gates passed.

## Restored objective

Reconstruct the supplied Default, Warped, and Forgotten Endoflame concept sheets as faithfully as practical in a live Minecraft renderer, without generated runtime imagery, while retaining Botania gameplay behavior and building a reusable renderer foundation for later flowers.

## dev21 renderer result

The final dev21 pass replaces the early thin/ribbon silhouette with a faceted flame-flower architecture:

- six broad unequal outer blades with independent scale/lift/roll and explicit plan-view curl
- tapered widening-to-narrowing facet chains, giving each blade a shoulder, body, curl, and needle tip
- a tall dominant hooked flame tongue and two shorter supporting tongues
- compact bright inner chamber instead of the earlier oversized white pyramid
- two-side-branch stem silhouette closer to the concept sheets
- shared geometry for Default / Warped / Forgotten with per-variant palettes and emissive behavior
- selective full-bright core/accent rendering and sparse active sparks
- hot-path geometry uses reusable allocation-free renderer primitives

The native front and top captures are the evidence of record for the final visual challenge pass:

- `dev21-three-clean.png`
- `dev21-top-attempt3.png`
- `dev21-showcase-regression.png`

## Build and static verification

`python3 tools/validate_release.py`

- PASS — 32 JSON files parsed
- PASS — required integration resources present

ForgeGradle:

- `compileJava` — PASS
- `reobfJar` — PASS
- final reobfuscated JAR: `botaniaexpanded-0.1.0-dev21.jar`

SHA-256:

`e1aa3b9a6acfe052ad4547cce63120be45746cc39db12e49209aaf95592cfefb`

JAR sanity inspection confirmed packaged classes/resources including:

- `BotaniaExpanded`
- `QaCommands`
- `ClientEvents`
- `EndoflameRenderer`
- `VoxelRenderUtil`
- `EndoflameVariant`
- `EndoflameVariantBlockEntity`
- `META-INF/mods.toml`
- `pack.mcmeta`

## Packaged-client runtime QA

Runtime harness:

- real production-namespace `forgeclient`
- Forge 47.4.23 / Minecraft 1.20.1
- Eclipse Temurin 17.0.20.1
- Xvfb 1280x720 + Mesa software renderer
- Botania 1.20.1-455
- Patchouli 1.20.1-85
- Curios 5.14.1+1.20.1
- Botania Expanded 0.1.0-dev21

Runtime gates:

- Forge client main menu — PASS
- mod discovery / Botania Expanded initialization — PASS
- integrated-world join — PASS
- Default / Warped / Forgotten renderer staging — PASS
- front-view native capture — PASS
- plan/top-view native capture — PASS
- task-owned missing-model / missing-texture / linkage / Mixin-failure audit — PASS

Observed harness-only/non-task failures were restricted-network Mojang/Yggdrasil/version-check requests, missing test-window icon resources, and no ALSA/OpenAL device in the headless container. None prevented resource initialization, world entry, rendering, or gameplay QA.

## Gameplay regression challenge

The final regression restaged `/botaniaexpanded showcase night` and queried the live block entities after fuel consumption.

Observed values included:

- Warped `burnTime = 760`
- Forgotten `burnTime = 684`
- stock Botania Endoflame `burnTime = 350`
- Warped `mana = 300`

Values are sampled at different ticks, so the exact remaining burn values are not expected to match one another. The material result is that the two custom variants consumed fuel and generated mana to their cap while the stock Endoflame continued burning under the client-only visual override.

## Scope / parity

The original Botania Endoflame remains Botania's own gameplay block entity. The client renderer is replaced only for that flower. Other Botania special-flower renderers are not replaced by this slice.

Warped and Forgotten are new optional variants that mirror Endoflame-style fuel/mana behavior and integrate with Botania tags/recipes/floating/potted content.

## Final challenge result

Compared with the early dev2/dev12 renderer, dev21 materially resolves the major concept-readability failures: narrow antler/ribbon petals, flat star plan-view, oversized central white mass, and bulky stem rosette. The remaining differences are the expected discretization of a Minecraft voxel/faceted renderer rather than an untested architectural defect.

This checkpoint therefore closes the Endoflame-family renderer slice as the current approval candidate. The reusable renderer primitives can now be carried forward to flower #2 without discarding this implementation.
