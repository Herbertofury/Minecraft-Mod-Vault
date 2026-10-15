# Archers 3.1.1 — Native Forge 1.20.1 Port Contract

## Objective
Port **Archers 3.1.1** to native **Minecraft 1.20.1 / Forge 47.4.23 / Java 17** while preserving the current upstream feature/content contract. Historical 1.20.1 source is mapping/API substrate only and must never silently replace current 3.1.1 behavior or content.

## Immutable upstream pins

### Current feature/content authority
- Repository: `ZsoltMolnarrr/Archers`
- Commit: `82a5329474548f5bcc7bddabd3cbabc2ab2cbda0`
- Upstream version: `3.1.1`
- Upstream Minecraft line: `1.21.1`
- Rule: current-source gameplay, content, assets, registrations, data, recipes, loot, balance, integrations, and quiver behavior are authoritative unless an exact 1.20.1 platform translation is required.

### Historical 1.20.1 mapping/API substrate
- Repository: `ZsoltMolnarrr/Archers`
- Branch provenance: `1.20.1`
- Commit: `795b1c4e7583adb26dc1a6959776817d25cbdd98`
- Historical version: `1.3.0`
- Rule: use only to recover valid 1.20.1 Yarn names, signatures, lifecycle patterns, and semantic anchors. It is **not** the feature authority and must not regress the current 3.1.1 content surface.

## Target platform
- Minecraft: `1.20.1`
- Forge: `47.4.23`
- Java bytecode: major `61` / Java 17
- Yarn mappings: `1.20.1+build.10`
- Native Forge runtime/package acceptance required; compile-only confidence is insufficient.

## Graduated/shared dependency baseline
Use the already graduated Forge 1.20.1 RPG-Series dependency artifacts rather than vendoring/stubbing their implementation:
- Armor Model API `1.0.0`
- Structure Pool API `1.2.1`
- Runes `1.3.2`
- Spell Power `1.6.0`
- Spell Engine `1.10.2 CORE`
- TinyConfig `3.1.0`
- Ranged Weapon API `2.3.4`
- Jewelry `2.4.0`
- Cloth Config
- Player Animator
- Curios
- Bundle API `1.1.0` native Forge 1.20.1 candidate, consumed as a **separate real JAR** for Archers integration/graduation evidence.

## Current 3.1.1 quiver contract
Current source: `common/src/main/java/net/archers/item/Quivers.java` at commit `82a5329474548f5bcc7bddabd3cbabc2ab2cbda0`.

The historical 1.20.1 Archers branch predates quivers. Therefore quivers must be translated from current 3.1.1 source, not copied from old Archers.

### Small quiver
- accepted tag: `ItemTags.ARROWS`
- capacity multiplier: `4`
- ordinary-arrow capacity: `256` (`4 * 64`)
- tier: `T3`
- material match: leather
- equip sound: leather armor equip
- max stack count: `1`
- lore: `item.archers.quiver.hint`, gray
- compatibility: wizard robe / rogue armor / crusader armor categories preserved from current source.

### Medium quiver
- accepted tag: `ItemTags.ARROWS`
- capacity multiplier: `8`
- ordinary-arrow capacity: `512`
- tier: `T4`
- material match: ender pearl
- equip sound: leather armor equip
- max stack count: `1`
- lore: `item.archers.quiver.hint`, gray
- same current-source compatibility categories.

### Large quiver
- accepted tag: `ItemTags.ARROWS`
- capacity multiplier: `12`
- ordinary-arrow capacity: `768`
- tier: `T5`
- material match: ender eye
- equip sound: leather armor equip
- max stack count: `1`
- lore: `item.archers.quiver.hint`, gray
- rarity: `UNCOMMON`
- same current-source compatibility categories.

## 1.20.1 Bundle API translation
Current 1.21.1 Archers creates quivers with `new CustomBundleItem(args.tag, args.settings)` and supplies the capacity through a data component default.

Minecraft 1.20.1 has no equivalent item data-component system. The native Bundle API 1.20.1 compatibility facade intentionally carries the immutable default capacity on the item instance instead:

```java
new CustomBundleItem(args.tag, key.capacity, args.settings)
```

This is a platform translation only. It must preserve the exact current 3.1.1 tag, 4/8/12 capacity multipliers, max-count/lore/rarity/equipment metadata, and runtime behavior.

Bundle API owns Forge-native tooltip registration in the 1.20.1 port; Archers must not add a duplicate tooltip registration shim merely to mimic the 1.21 component registration call.

## Bundle API downstream graduation matrix
Before Bundle API is called graduated, an Archers consumer fixture or the real Archers port must prove against a **separate packaged Bundle API JAR** in initialized Forge 1.20.1:

1. Exact translated current `Quivers` consumer code compiles against the public Bundle API facade.
2. Quiver items are real registered items, not unregistered synthetic instances.
3. Loaded Minecraft tags prove `ItemTags.ARROWS` behavior: ordinary arrows accepted; spectral arrows accepted; a non-arrow item rejected.
4. Small quiver accepts exactly `256` ordinary arrows and stores only legal `<=64` ItemStack chunks.
5. Medium quiver accepts exactly `512` ordinary arrows.
6. Large quiver accepts exactly `768` ordinary arrows.
7. Insert past full capacity is rejected without item loss/duplication.
8. Removal/front-stack semantics remain valid.
9. Each quiver has max stack count `1`.
10. Large quiver rarity is `UNCOMMON`.
11. Gray `item.archers.quiver.hint` lore contract remains present.
12. Dedicated Forge server reaches `Done` with Archers + Bundle API as separate artifacts and no loader/client leakage.
13. Forge client reaches LWJGL + ResourceManager bootstrap with both artifacts and no tooltip/client-registration fatal errors.

Only after the strong Bundle API standalone runtime gates and this real downstream consumer matrix pass may Bundle API be rotated out of `ACTIVE_RUNNER` and marked graduated.

## Archers full-port acceptance after Bundle API graduation
The Archers lane then expands from the quiver boundary to the entire current 3.1.1 surface:
- preserve all current owned Java behavior and registrations;
- preserve current resources/data/assets/models/animations/configs/loot/recipes/tags;
- use real graduated dependencies as separate artifacts;
- Java 17 class-major gate;
- remapped native Forge release JAR inspection;
- deterministic clean rebuild and source archive;
- dedicated Forge dev server runtime;
- headless Forge client/resource runtime;
- fresh packaged Forge 47.4.23 server using the exact release JAR;
- whole-suite integration with the already graduated RPG-Series foundations.

No historical 1.3.0 omission is permission to omit a current 3.1.1 feature.

## Graduation receipt — 2026-08-27

Status: **GRADUATED — FULL NATIVE FORGE ACCEPTANCE PASSED**

### Frozen runtime/product authority
- Accepted canonical head: `d47274c88d26661a11cea25ba4805fc4393fe195`
- Authoritative workflow: RPG Series active-port run `#179` / run ID `33141697191`
- Verifier job: `98753736940`
- Release artifact: `archers-forge-3.1.1+1.20.1.jar`
- Release size: `901621` bytes
- Release SHA-256: `38991c26b9de3e5ac51ce72757b40f1eb2aef3ace3e66f844de44d2a4cbde8b1`
- Java gate: all `51` Archers-owned classes are Java 17 major `61`; all `52` packaged classes are `<=61`.

The runtime/product authority above is frozen. Any later documentation, ecosystem bookkeeping, dependency graduation, or QA-only commit does not supersede these accepted Archers product bytes.

### Native acceptance evidence
The exact accepted head passed the full graduation matrix rather than compile-only confidence:
- exact current Archers 3.1.1 materialization with `545` resource/data files retained;
- common and native Forge compilation/package gates;
- dependency-leakage boundaries with RPG foundations kept as separate JARs;
- byte-identical clean remap/rebuild of the release JAR;
- real Forge development server readiness plus `ARCHERS_SELF_TEST_PASS`;
- real Forge client LWJGL, ResourceManager, mixin, and render/bootstrap acceptance;
- fresh official Forge `47.4.23` packaged server using the exact release JAR bytes plus separate runtime dependencies;
- packaged-server semantic self-test green.

Terminal acceptance marker:

`PASS: deterministic package + semantic dev server + headless client + fresh packaged server.`

### Preserved evidence
- GitHub Actions evidence artifact ID: `9674283430`
- Evidence ZIP SHA-256: `924004f19679b0f851b5e4471f684023d19fd4ce3a7cc03c7c1bdd94f0c2324e`
- Evidence ZIP size: `7227248` bytes
- Canonical Drive evidence ID: `1miGN0xv7TIXVSeBXFd52V3Fl3ER4VRch`
- Canonical Drive release JAR ID: `1n9vMSpPg6kGIGg1yQvZcctjydVWauj--`

The Drive evidence/release copies were size-verified when graduation was sealed, preserving the acceptance record independently of GitHub Actions artifact retention.
