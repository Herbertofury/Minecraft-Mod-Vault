# Noxviola FPS / Error Cleanup Wave 1

Status: **partial / runtime acceptance pending for fresh binary patches**.

Fresh log basis: 2026-09-06 Noxviola Forge 1.20.1 run after C2ME 9.1 + DimThread/Midnight lock fixes.

## Fresh-run findings
- The previous 40s+ watchdog hard lock did **not** reproduce in this log.
- Current Curios is stock 5.14.1 and the accepted ApothicCurios v17 bridge is absent.
- Sentry Mechanical Arm repeatedly reflects `RecipeManager.recipes`, which does not exist under the production SRG runtime name.
- LookinSharp contains malformed `weakening_artifact.json` loot-modifier data.
- Alex's Mobs and Scorched Guns both use global mutable level-tick state that races under DimThread parallel dimensions and throws `ConcurrentModificationException`.
- Physics Mod 3.0.20 probes a missing `PhysXGpu_64.dll`; this remains a performance investigation, not patched in this wave.
- The log reports Entity Culling can hide Immersive Vehicles unless three MTS entity IDs are whitelisted.

## Recommended JAR set

### Previously accepted performance/compatibility artifacts
- `curios-forge-5.14.1+1.20.1-Noxviola-PerfLeakFix-v3.jar`
  - SHA-256 `4c4b668d4ddca09504f1aa80341e9a90e4405f2969178a5fa5e4542cbbff9803`
  - Prior real packaged client/server QA PASS.
  - Optimizes Curios render/alloc hot paths without removing slots/content/behavior.
- `ApothicCurios-1.20.1-1.0.3e-Noxviola-ZeroLagFix-v17-ClientPolish.jar`
  - SHA-256 `c4420a71033266af42e80bd5d5bcea0154994e5fa12c714bf1b8016ec6d7b8c3`
  - Prior production client/server QA PASS.
  - Restores the accepted dynamic `curios:*` Apotheosis loot-category bridge.

### Fresh patches built from exact installed/upstream binaries
- `sentrymechanicalarm-0.3.6-1.20.1-Forge-Noxviola-RecipeMapFix-v1.jar`
  - SHA-256 `e6a2a72475f6c07dab43620a45b4e762491f6d472135995014be6383b0fc6f5d`
  - Only content-changed JAR entry: `euphy/upo/sentrymechanicalarm/recipe/DynamicRecipeManager.class`.
  - Production field lookup changed from Mojmap `recipes` to SRG `f_44007_`.
- `lookinsharp-forge-1.20.1-1.0.3-Noxviola-LootJsonFix-v1.jar`
  - SHA-256 `c653c91344659fbe57f92671798c79c823f79fd75ac5abc6908b25f3ed473bdd`
  - Only content-changed entry: `data/lookinsharp/loot_modifiers/weakening_artifact.json`.
- `alexsmobs-1.22.9-Noxviola-DimThread-Interop-v1.jar`
  - SHA-256 `35e40b0f10cad0c5f67427dbe7de8aa40c136313de201545ebe6a9a1ee284473`
  - Only content-changed entry: `com/github/alexthe666/alexsmobs/event/ServerEvents.class`.
  - Observed shared level-tick handler made synchronized to preserve serial access to its global tick state across DimThread dimension workers.
- `ScorchedGuns-0.5.5-1.20.1-Noxviola-DimThread-Interop-v1.jar`
  - SHA-256 `3dcb0c84a6a7d5f8f7fedfeaaaeb33bae2c6e74290b2a754d92f997e5519dd34`
  - Only content-changed entry: `top/ribs/scguns/entity/raid/RaidManager.class`.
  - Observed shared raid level-tick handler made synchronized to preserve upstream serial semantics across parallel dimensions.

All six recommended artifacts pass ZIP/JAR integrity. The four fresh patches preserve mod metadata and add/remove no JAR entries.

## Held back intentionally
`TCC CuriosSlotFix v1` was built during diagnosis but is **not** recommended for this wave. The current log is missing the already-accepted ApothicCurios v17 category bridge, so restoring that known-good baseline is safer than stacking another new TCC patch first.

## Next acceptance run
Replace stock originals rather than co-installing duplicate mod IDs. Run the same world and inspect:
- no Sentry `NoSuchFieldException: recipes` flood;
- no LookinSharp weakening-artifact decode error;
- no Alex's Mobs / Scorched Guns `ConcurrentModificationException` under DimThread;
- TCC `curios:tcc_*` affix/category errors after ApothicCurios v17 is restored;
- actual FPS/frametime in the same scene.
