#!/usr/bin/env python3
from __future__ import annotations

import pathlib
import sys

root = pathlib.Path(sys.argv[1]).resolve()
if not root.is_dir():
    raise SystemExit(f"missing generated Paladins Java root: {root}")


def replace_exact(rel: str, old: str, new: str, label: str) -> None:
    path = root / rel
    text = path.read_text(encoding="utf-8")
    count = text.count(old)
    if count != 1:
        raise SystemExit(f"[{label}] expected exactly one pinned source shape in {rel}, found {count}")
    path.write_text(text.replace(old, new, 1), encoding="utf-8")
    print(f"[Paladins 1.20.1 API batch2] {label}: {rel}")


def replace_all(rel: str, old: str, new: str, expected: int, label: str) -> None:
    path = root / rel
    text = path.read_text(encoding="utf-8")
    count = text.count(old)
    if count != expected:
        raise SystemExit(f"[{label}] expected {expected} pinned occurrences in {rel}, found {count}")
    path.write_text(text.replace(old, new), encoding="utf-8")
    print(f"[Paladins 1.20.1 API batch2] {label}: {count} replacements in {rel}")


# Entity renderer setupTransforms lost the extra render-scale argument in 1.20.1.
replace_exact(
    "net/paladins/client/entity/LightwellEntityRenderer.java",
    "protected void setupTransforms(LightwellEntity entity, MatrixStack matrices, float animationProgress, float bodyYaw, float tickDelta, float scale) {\n        super.setupTransforms(entity, matrices, animationProgress, bodyYaw, tickDelta, scale);",
    "protected void setupTransforms(LightwellEntity entity, MatrixStack matrices, float animationProgress, float bodyYaw, float tickDelta) {\n        super.setupTransforms(entity, matrices, animationProgress, bodyYaw, tickDelta);",
    "Lightwell renderer 1.20.1 setupTransforms signature",
)

# 1.20.1 entity models still use separate RGBA float render parameters.
replace_exact(
    "net/paladins/client/entity/LightwellEntityModel.java",
    "public void render(MatrixStack matrices, VertexConsumer vertices, int light, int overlay, int color) {\n        root.render(matrices, vertices, light, overlay, color);",
    "public void render(MatrixStack matrices, VertexConsumer vertices, int light, int overlay, float red, float green, float blue, float alpha) {\n        root.render(matrices, vertices, light, overlay, red, green, blue, alpha);",
    "Lightwell model 1.20.1 RGBA render signature",
)
replace_exact(
    "net/paladins/client/entity/LightwellGlowFeatureRenderer.java",
    "this.getContextModel().render(matrices, vertexConsumer, 15728640, OverlayTexture.DEFAULT_UV);",
    "this.getContextModel().render(matrices, vertexConsumer, 15728640, OverlayTexture.DEFAULT_UV, 1F, 1F, 1F, 1F);",
    "Lightwell glow 1.20.1 RGBA render call",
)
replace_exact(
    "net/paladins/client/entity/BannerEntityRenderer.java",
    "model.render(matrices, vertices, light, OverlayTexture.DEFAULT_UV, -1);",
    "model.render(matrices, vertices, light, OverlayTexture.DEFAULT_UV, 1F, 1F, 1F, 1F);",
    "Battle Banner 1.20.1 RGBA render call",
)

# SummonedEntity implements Yarn's 1.20.1 Tameable contract. The extra interface method is the
# entity-view owner lookup context; World is the exact EntityView for this entity.
replace_exact(
    "net/paladins/entity/LightwellEntity.java",
    "import net.minecraft.world.World;",
    "import net.minecraft.world.EntityView;\nimport net.minecraft.world.World;",
    "Lightwell import 1.20.1 EntityView",
)
replace_exact(
    "net/paladins/entity/LightwellEntity.java",
    "    public LightwellEntity(EntityType<? extends LightwellEntity> entityType, World world) {\n        super(entityType, world);\n    }\n",
    "    public LightwellEntity(EntityType<? extends LightwellEntity> entityType, World world) {\n        super(entityType, world);\n    }\n\n    @Override\n    public EntityView method_48926() {\n        return this.getWorld();\n    }\n",
    "Lightwell satisfy Yarn 1.20.1 Tameable EntityView contract",
)

# Barrier networking and dynamic-registry lookup drift.
replace_exact(
    "net/paladins/entity/BarrierEntity.java",
    "import net.minecraft.registry.RegistryKeys;",
    "import net.minecraft.registry.RegistryKey;\nimport net.minecraft.registry.RegistryKeys;",
    "Barrier import RegistryKey",
)
replace_exact(
    "net/paladins/entity/BarrierEntity.java",
    "serverPlayer.networkHandler.send(\n                                        new EntityVelocityUpdateS2CPacket(serverPlayer.getId(), serverPlayer.getVelocity()),\n                                        null\n                                );",
    "serverPlayer.networkHandler.send(\n                                        new EntityVelocityUpdateS2CPacket(serverPlayer.getId(), serverPlayer.getVelocity())\n                                );",
    "Barrier 1.20.1 packet send signature",
)
replace_exact(
    "net/paladins/entity/BarrierEntity.java",
    "return SpellRegistry.from(this.getWorld()).getEntry(this.spellId).orElse(null);",
    "return SpellRegistry.from(this.getWorld()).getEntry(RegistryKey.of(SpellRegistry.KEY, this.spellId)).orElse(null);",
    "Barrier 1.20.1 dynamic registry entry lookup",
)

# 1.20.1 VertexConsumer normal() takes Matrix3f and vertices must be terminated with next().
replace_exact(
    "net/paladins/client/entity/BarrierEntityRenderer.java",
    "import org.joml.Matrix4f;",
    "import org.joml.Matrix3f;\nimport org.joml.Matrix4f;",
    "Barrier renderer import Matrix3f",
)
replace_exact(
    "net/paladins/client/entity/BarrierEntityRenderer.java",
    "                var matrixEntry = matrices.peek();\n                // Matrix3f normalMatrix = matrixEntry.getNormalMatrix();",
    "                var matrixEntry = matrices.peek();\n                Matrix3f normalMatrix = matrixEntry.getNormalMatrix();",
    "Barrier renderer 1.20.1 normal matrix",
)
barrier_renderer = root / "net/paladins/client/entity/BarrierEntityRenderer.java"
text = barrier_renderer.read_text(encoding="utf-8")
normal_count = text.count(".normal(matrixEntry, 0, 0, 0);")
if normal_count != 16:
    raise SystemExit(f"[Barrier vertex completion] expected 16 current vertex endings, found {normal_count}")
text = text.replace(".normal(matrixEntry, 0, 0, 0);", ".normal(normalMatrix, 0, 0, 0).next();")
barrier_renderer.write_text(text, encoding="utf-8")
print(f"[Paladins 1.20.1 API batch2] Barrier 1.20.1 normal/next vertex chain: {normal_count} replacements")

# RenderPhase fields are protected in 1.20.1. Keep the exact current layer composition by making
# the access happen inside a tiny RenderPhase subclass; do not downgrade the shader/transparency setup.
replace_exact(
    "net/paladins/client/entity/BarrierEntityRenderer.java",
    "    private record Config(\n",
    "    private static final class RenderPhaseAccess extends RenderPhase {\n        private RenderPhaseAccess() { super(\"paladins_render_phase_access\", () -> {}, () -> {}); }\n\n        static RenderLayer vanillaBarrierLayer() {\n            return CustomLayers.create(\n                    SpriteAtlasTexture.BLOCK_ATLAS_TEXTURE,\n                    BEACON_BEAM_PROGRAM,\n                    TRANSLUCENT_TRANSPARENCY,\n                    DISABLE_CULLING,\n                    COLOR_MASK,\n                    ENABLE_OVERLAY_COLOR,\n                    MAIN_TARGET,\n                    true);\n        }\n\n        static RenderLayer irisBarrierLayer() {\n            return CustomLayers.create(\n                    SpriteAtlasTexture.BLOCK_ATLAS_TEXTURE,\n                    LIGHTNING_PROGRAM,\n                    LIGHTNING_TRANSPARENCY,\n                    DISABLE_CULLING,\n                    COLOR_MASK,\n                    ENABLE_OVERLAY_COLOR,\n                    MAIN_TARGET,\n                    false);\n        }\n    }\n\n    private record Config(\n",
    "Barrier protected RenderPhase access bridge",
)
replace_exact(
    "net/paladins/client/entity/BarrierEntityRenderer.java",
    "                CustomLayers.create(\n                        SpriteAtlasTexture.BLOCK_ATLAS_TEXTURE,\n                        BEACON_BEAM_PROGRAM,\n                        TRANSLUCENT_TRANSPARENCY,\n                        DISABLE_CULLING,\n                        COLOR_MASK,\n                        ENABLE_OVERLAY_COLOR,\n                        MAIN_TARGET,\n                        true),",
    "                RenderPhaseAccess.vanillaBarrierLayer(),",
    "Barrier vanilla layer retains current shader composition",
)
replace_exact(
    "net/paladins/client/entity/BarrierEntityRenderer.java",
    "                CustomLayers.create(\n                        SpriteAtlasTexture.BLOCK_ATLAS_TEXTURE,\n                        LIGHTNING_PROGRAM,\n                        LIGHTNING_TRANSPARENCY,\n                        DISABLE_CULLING,\n                        COLOR_MASK,\n                        ENABLE_OVERLAY_COLOR,\n                        MAIN_TARGET,\n                        false),",
    "                RenderPhaseAccess.irisBarrierLayer(),",
    "Barrier Iris layer retains current shader composition",
)

# EntityType.Builder naming drift is deterministic in Yarn 1.21 -> 1.20.1.
replace_all(
    "net/paladins/entity/PaladinEntities.java",
    ".dimensions(",
    ".setDimensions(",
    3,
    "EntityType.Builder dimensions -> setDimensions",
)

# 1.20.1 attribute operation enum names.
operation_map = {
    "EntityAttributeModifier.Operation.ADD_MULTIPLIED_BASE": "EntityAttributeModifier.Operation.MULTIPLY_BASE",
    "EntityAttributeModifier.Operation.ADD_MULTIPLIED_TOTAL": "EntityAttributeModifier.Operation.MULTIPLY_TOTAL",
    "EntityAttributeModifier.Operation.ADD_VALUE": "EntityAttributeModifier.Operation.ADDITION",
}
operation_total = 0
for path in sorted(root.rglob("*.java")):
    text = path.read_text(encoding="utf-8")
    original = text
    for old, new in operation_map.items():
        c = text.count(old)
        operation_total += c
        text = text.replace(old, new)
    if text != original:
        path.write_text(text, encoding="utf-8")
if operation_total < 8:
    raise SystemExit(f"expected broad operation-enum drift, translated only {operation_total} occurrences")
print(f"[Paladins 1.20.1 API batch2] attribute operation enum translations: {operation_total}")

# Attribute#getIdAsString is newer. Resolve IDs through the 1.20.1 attribute registry.
replace_exact(
    "net/paladins/effect/PaladinEffects.java",
    "import net.minecraft.entity.effect.StatusEffectCategory;",
    "import net.minecraft.entity.effect.StatusEffectCategory;\nimport net.minecraft.registry.Registries;",
    "PaladinEffects import attribute registry",
)
for attr in ("GENERIC_ATTACK_SPEED", "GENERIC_KNOCKBACK_RESISTANCE"):
    replace_exact(
        "net/paladins/effect/PaladinEffects.java",
        f"EntityAttributes.{attr}.getIdAsString()",
        f"Registries.ATTRIBUTE.getId(EntityAttributes.{attr}).toString()",
        f"PaladinEffects 1.20.1 ID lookup for {attr}",
    )

# Later vanilla attributes do not exist in 1.20.1. Preserve behavior through target-native effect
# logic instead of deleting the features or registering fake attributes the engine would never read.
effects = root / "net/paladins/effect/PaladinEffects.java"
text = effects.read_text(encoding="utf-8")
for marker, replacement, label in (
    (
        '''            new JudgementStatusEffect(StatusEffectCategory.HARMFUL, 0xffffcc),\n            new EffectConfig(List.of(\n                    new AttributeModifier(\n                            EntityAttributes.GENERIC_JUMP_STRENGTH.getIdAsString(),\n                            0,\n                            EntityAttributeModifier.Operation.MULTIPLY_TOTAL\n                    )\n                )\n            )''',
        '''            new JudgementStatusEffect(StatusEffectCategory.HARMFUL, 0xffffcc)''',
        "Judgement absent 1.20.1 jump-strength attribute delegated to STUN action impairment",
    ),
    (
        '''            new LevitateStatusEffect(StatusEffectCategory.BENEFICIAL, 0xffffcc),\n            new EffectConfig(List.of(\n                    new AttributeModifier(\n                            EntityAttributes.GENERIC_GRAVITY.getIdAsString(),\n                            -0.99F,\n                            EntityAttributeModifier.Operation.MULTIPLY_TOTAL\n                    )\n            ))''',
        '''            new LevitateStatusEffect(StatusEffectCategory.BENEFICIAL, 0xffffcc)''',
        "Levitate absent 1.20.1 gravity attribute moved to effect tick emulation",
    ),
    (
        '''            new PriestAbsorptionStatusEffect(StatusEffectCategory.BENEFICIAL, 0xffffcc),\n            new EffectConfig(List.of(\n                    new AttributeModifier(\n                            EntityAttributes.GENERIC_MAX_ABSORPTION.getIdAsString(),\n                            2,\n                            EntityAttributeModifier.Operation.ADDITION\n                    )\n            ))''',
        '''            new PriestAbsorptionStatusEffect(StatusEffectCategory.BENEFICIAL, 0xffffcc)''',
        "Absorption absent 1.20.1 max-absorption attribute moved to effect lifecycle",
    ),
):
    count = text.count(marker)
    if count != 1:
        raise SystemExit(f"[{label}] expected one pinned current shape, found {count}")
    text = text.replace(marker, replacement, 1)
    print(f"[Paladins 1.20.1 API batch2] {label}")
effects.write_text(text, encoding="utf-8")

# Historical 1.20.1 Paladins proves absorption is implemented directly on the entity lifecycle.
priest = root / "net/paladins/effect/PriestAbsorptionStatusEffect.java"
priest.write_text('''package net.paladins.effect;\n\nimport net.minecraft.entity.LivingEntity;\nimport net.minecraft.entity.attribute.AttributeContainer;\nimport net.minecraft.entity.effect.StatusEffect;\nimport net.minecraft.entity.effect.StatusEffectCategory;\n\npublic class PriestAbsorptionStatusEffect extends StatusEffect {\n    private final int healthPerStack;\n\n    public PriestAbsorptionStatusEffect(StatusEffectCategory category, int color) {\n        super(category, color);\n        this.healthPerStack = 2;\n    }\n\n    @Override\n    public void onApplied(LivingEntity entity, AttributeContainer attributes, int amplifier) {\n        entity.setAbsorptionAmount(Math.max(entity.getAbsorptionAmount(), (float)(healthPerStack * (1 + amplifier))));\n        super.onApplied(entity, attributes, amplifier);\n    }\n\n    @Override\n    public void onRemoved(LivingEntity entity, AttributeContainer attributes, int amplifier) {\n        entity.setAbsorptionAmount(Math.max(0F, entity.getAbsorptionAmount() - (float)(healthPerStack * (1 + amplifier))));\n        super.onRemoved(entity, attributes, amplifier);\n    }\n}\n''', encoding="utf-8")
print("[Paladins 1.20.1 API batch2] Priest absorption uses historical 1.20.1 lifecycle API with current max-stack semantics")

# 1.20.1 has no generic gravity attribute. Offset vanilla living-entity gravity each effect tick while
# leaving spell-applied upward impulses intact; hand off to Slow Falling on the final tick as current code does.
levitate = root / "net/paladins/effect/LevitateStatusEffect.java"
levitate.write_text('''package net.paladins.effect;\n\nimport net.minecraft.entity.LivingEntity;\nimport net.minecraft.entity.effect.StatusEffectCategory;\nimport net.minecraft.entity.effect.StatusEffectInstance;\nimport net.minecraft.entity.effect.StatusEffects;\nimport net.spell_engine.api.effect.CustomStatusEffect;\n\npublic class LevitateStatusEffect extends CustomStatusEffect {\n    private static final int SLOW_FALLING_TICKS = 3 * 20;\n    private static final double GRAVITY_COMPENSATION = 0.076D; // 0.08 vanilla -> about 0.004 downward\n\n    public LevitateStatusEffect(StatusEffectCategory category, int color) {\n        super(category, color);\n    }\n\n    @Override\n    public boolean canApplyUpdateEffect(int duration, int amplifier) {\n        return true;\n    }\n\n    @Override\n    public void applyUpdateEffect(LivingEntity entity, int amplifier) {\n        if (!entity.isOnGround() && !entity.hasNoGravity()) {\n            var velocity = entity.getVelocity();\n            entity.setVelocity(velocity.x, velocity.y + GRAVITY_COMPENSATION, velocity.z);\n        }\n        var instance = entity.getStatusEffect(this);\n        if (!entity.getWorld().isClient() && instance != null && instance.getDuration() <= 1) {\n            entity.addStatusEffect(new StatusEffectInstance(\n                    StatusEffects.SLOW_FALLING, SLOW_FALLING_TICKS, 0, false, true, true));\n        }\n    }\n}\n''', encoding="utf-8")
print("[Paladins 1.20.1 API batch2] Levitate current gravity/landing semantics emulated without nonexistent 1.21 attributes")

# Fail closed on every symbol this batch claims to translate.
forbidden = (
    ".dimensions(",
    "ADD_MULTIPLIED_BASE",
    "ADD_MULTIPLIED_TOTAL",
    "ADD_VALUE",
    ".getIdAsString()",
    "GENERIC_JUMP_STRENGTH",
    "GENERIC_GRAVITY",
    "GENERIC_MAX_ABSORPTION",
    ".normal(matrixEntry, 0, 0, 0)",
    "setupTransforms(entity, matrices, animationProgress, bodyYaw, tickDelta, scale)",
)
survivors: list[str] = []
for path in sorted(root.rglob("*.java")):
    for line_no, line in enumerate(path.read_text(encoding="utf-8").splitlines(), 1):
        if any(token in line for token in forbidden):
            survivors.append(f"{path.relative_to(root)}:{line_no}:{line.strip()}")
if survivors:
    raise SystemExit("second-frontier API survived compatibility pass:\n" + "\n".join(survivors))

print("[Paladins 1.20.1 API batch2] second javac compatibility frontier translated fail-closed")
