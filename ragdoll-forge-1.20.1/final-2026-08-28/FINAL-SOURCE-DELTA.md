# Final source delta — production Forge 1.20.1 fixes

This file records the exact final source/resource delta applied after the previous full compile/remap/server-green checkpoint. The complete source snapshot is also embedded in the SHA-pinned Drive release archive (`373baea9096c2e8f1126e634b8e1183620c364d0dff45d6357105a020f870524`).

## Direct Mixin production targets

Changed these mapped member targets to their exact Forge 1.20.1 production SRG names:

- `LivingEntityRendererAccessor`: `layers` -> `f_115291_`
- `RagdollCameraDistanceMixin`: Camera `entity` -> `f_90551_`
- `RagdollFirstPersonEyeMixin`: Camera `position` -> `f_90552_`
- `RagdollFirstPersonEyeMixin`: Camera `setPosition(double,double,double)` -> `m_90584_`
- `HumanoidModelMixin`: `head` -> `f_102808_`
- `HumanoidModelMixin`: `rightArm` -> `f_102811_`
- `HumanoidModelMixin`: `leftArm` -> `f_102812_`

Affected source files:

- `core/src/main/java/dev/leo/sableplayerragdoll/neoforge/mixin/LivingEntityRendererAccessor.java`
- `core/src/main/java/dev/leo/sableplayerragdoll/neoforge/mixin/RagdollCameraDistanceMixin.java`
- `core/src/main/java/dev/leo/sableplayerragdoll/neoforge/mixin/RagdollFirstPersonEyeMixin.java`
- `core/src/main/java/dev/leo/sableplayerragdoll/neoforge/mixin/HumanoidModelMixin.java`

## Reflection mapped + SRG aliases

Reflection helpers were changed to try development/mapped names and production aliases deliberately. Final aliases:

- ModelPart `cubes` / `f_104212_`
- ModelPart `children` / `f_104213_`
- LivingEntityRenderer `layers` / `f_115291_`
- EntityModel `young` / `f_102610_`
- LivingEntityRenderer `getWhiteOverlayProgress` / `m_6931_`
- LivingEntityRenderer `scale` / `m_7546_`
- AgeableListModel `headParts` / `m_5607_`
- AgeableListModel `bodyParts` / `m_5608_`
- AgeableListModel `scaleHead` / `f_102007_`
- `babyYHeadOffset` / `f_170338_`
- `babyZHeadOffset` / `f_170339_`
- `babyHeadScale` / `f_102010_`
- `babyBodyScale` / `f_102011_`
- `bodyYOffset` / `f_102012_`

Affected files:

- `mob/client/MobRagdollModelParts.java`
- `mob/client/ModelPartMask.java`
- `mob/client/MobRagdollPartBlockEntityRenderer.java`
- `mob/client/MobRagdollLayerRenderer.java`
- `mob/client/MobRagdollClientExtractor.java`
- `mob/model/RenderedModelExtractor.java`

## Descriptor-aware LambdaMetafactory SAM remapping

`tools/ForgeRuntimeRemapper.java` now uses a `SamAwareClassRemapper` / `MethodRemapper`. For `java/lang/invoke/LambdaMetafactory` sites it:

1. reads bootstrap argument 0 as the SAM `Type` descriptor;
2. derives the functional-interface owner from the invokedynamic call-site return type;
3. resolves the method with `maps.findMethod(owner, name, sam.getDescriptor())`;
4. rewrites the invokedynamic method name only when that exact tuple maps;
5. increments the ordinary method-remap hit count.

The generic `mapInvokeDynamicMethodName` path intentionally leaves names unchanged because the descriptor-aware remap occurs at the actual call site.

This moved the core remap result from the old **1743** method remaps to the final **1751**, exactly accounting for the eight previously broken SAM sites. The upgraded production validator reports zero problems on the final JAR.

`tools/ProductionLinkageValidator.java` and `tools/RemapValidator.java` were also upgraded so invokedynamic/SAM linkage and stale production mappings are part of static release validation.

## Resource-pack completion

Added `pack.mcmeta` with Minecraft 1.20.1 `pack_format: 15` to:

- core
- reactions
- future bridge

The last production model warning was `sable_player_ragdoll:ragdoll_seat`. Source inspection proves `RagdollSeatBlock#getRenderShape` returns `RenderShape.INVISIBLE`; therefore the release adds a valid zero-geometry model rather than a visible placeholder:

`assets/sable_player_ragdoll/blockstates/ragdoll_seat.json`

```json
{
  "variants": {
    "": { "model": "sable_player_ragdoll:block/ragdoll_seat" }
  }
}
```

`assets/sable_player_ragdoll/models/block/ragdoll_seat.json`

```json
{
  "textures": {
    "particle": "sable_player_ragdoll:block/empty"
  }
}
```

## Verification identity

Final core JAR: `0e9ee19b3b4b3efe4da07d2f7c47c29f98eeb805f96f95698163975fb71e1b47`.

Its 201 packaged `.class` files are byte-identical to the newest remap output. Reactions and future-bridge final class sets are byte-identical to the previously gameplay-proven Candidate C class sets; only release resources changed. The exact source snapshot, validators, logs, and checksums are preserved in the full Drive release archive.
