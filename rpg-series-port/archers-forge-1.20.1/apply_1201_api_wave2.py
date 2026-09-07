#!/usr/bin/env python3
from __future__ import annotations

import pathlib
import sys

java_root = pathlib.Path(sys.argv[1]).resolve()
resources_root = pathlib.Path(sys.argv[2]).resolve()
if not java_root.is_dir():
    raise SystemExit(f"missing generated Java root: {java_root}")
if not resources_root.is_dir():
    raise SystemExit(f"missing generated resources root: {resources_root}")


def require_replace(path: pathlib.Path, old: str, new: str, label: str) -> None:
    if not path.is_file():
        raise SystemExit(f"{label}: missing source: {path}")
    text = path.read_text(encoding="utf-8")
    if old not in text:
        raise SystemExit(f"{label}: expected current-upstream seam not found")
    path.write_text(text.replace(old, new), encoding="utf-8")


def require_replace_all(path: pathlib.Path, old: str, new: str, expected: int, label: str) -> None:
    if not path.is_file():
        raise SystemExit(f"{label}: missing source: {path}")
    text = path.read_text(encoding="utf-8")
    actual = text.count(old)
    if actual != expected:
        raise SystemExit(f"{label}: expected {expected} occurrences, found {actual}")
    path.write_text(text.replace(old, new), encoding="utf-8")


# 1.20.1 predates GENERIC_SCALE. Preserve the current marker's 75%-of-entity-height placement;
# there is no scale divisor to undo in the target runtime.
hunters_mark = java_root / "net/archers/client/effect/HuntersMarkRenderer.java"
require_replace(
    hunters_mark,
    "var verticalOffset = (livingEntity.getHeight() / livingEntity.getScale()) * 0.75F;",
    "var verticalOffset = livingEntity.getHeight() * 0.75F;",
    "Hunter's Mark target height",
)

# 1.20.1 Model still uses four float tint channels instead of the later packed-color argument.
direwolf_model = java_root / "net/archers/client/entity/DirewolfEntityModel.java"
require_replace(
    direwolf_model,
    "public void render(MatrixStack matrices, VertexConsumer vertices, int light, int overlay, int color) {\n\t\troot.render(matrices, vertices, light, overlay, color);\n\t}",
    "public void render(MatrixStack matrices, VertexConsumer vertices, int light, int overlay, float red, float green, float blue, float alpha) {\n\t\troot.render(matrices, vertices, light, overlay, red, green, blue, alpha);\n\t}",
    "Direwolf 1.20.1 model render signature",
)

# The current Archers spirit layer intentionally differs from vanilla emissive rendering by writing
# depth (ALL_MASK). 1.20.1's seven-argument RenderLayer.of factory is private, but RenderLayer's
# constructor and each RenderPhase start/end action are public. Reproduce the same phases directly,
# preserving translucent emissive rendering, no culling, overlay, and depth writes without reflection.
render_layers = java_root / "net/archers/client/render/ArcherRenderLayers.java"
old_render_factory = '''    private static final Function<Identifier, RenderLayer> SPIRIT = Util.memoize(texture -> {
        var parameters = MultiPhaseParameters.builder()
                .program(ENTITY_TRANSLUCENT_EMISSIVE_PROGRAM)
                .texture(new RenderPhase.Texture(texture, false, false))
                .transparency(TRANSLUCENT_TRANSPARENCY)
                .cull(DISABLE_CULLING)
                .writeMaskState(ALL_MASK)
                .overlay(ENABLE_OVERLAY_COLOR)
                .build(true);
        return of("archers_spirit",
                VertexFormats.POSITION_COLOR_TEXTURE_OVERLAY_LIGHT_NORMAL,
                VertexFormat.DrawMode.QUADS,
                1536,
                true,
                true,
                parameters);
    });'''
new_render_factory = '''    private static final Function<Identifier, RenderLayer> SPIRIT = Util.memoize(texture -> {
        var texturePhase = new RenderPhase.Texture(texture, false, false);
        RenderPhase[] phases = new RenderPhase[] {
                ENTITY_TRANSLUCENT_EMISSIVE_PROGRAM,
                texturePhase,
                TRANSLUCENT_TRANSPARENCY,
                DISABLE_CULLING,
                ALL_MASK,
                ENABLE_OVERLAY_COLOR
        };
        return new ArcherRenderLayers(
                "archers_spirit",
                VertexFormats.POSITION_COLOR_TEXTURE_OVERLAY_LIGHT_NORMAL,
                VertexFormat.DrawMode.QUADS,
                1536,
                true,
                true,
                () -> {
                    for (RenderPhase phase : phases) {
                        phase.startDrawing();
                    }
                },
                () -> {
                    for (RenderPhase phase : phases) {
                        phase.endDrawing();
                    }
                });
    });'''
require_replace(render_layers, old_render_factory, new_render_factory, "Spirit render layer 1.20.1 construction")

# Attribute operation names and attribute identifiers changed after 1.20.1. Registry lookup is the
# target-native equivalent of modern EntityAttribute#getIdAsString(). Generic jump strength does not
# exist on 1.20.1 living entities; its Entangling Roots behavior is retained by the compatibility mixin
# emitted below instead of being silently dropped or incorrectly mapped to HORSE_JUMP_STRENGTH.
effects = java_root / "net/archers/effect/ArcherEffects.java"
require_replace(
    effects,
    "import net.minecraft.entity.effect.StatusEffectCategory;\n",
    "import net.minecraft.entity.effect.StatusEffectCategory;\nimport net.minecraft.registry.Registries;\n",
    "ArcherEffects registry import",
)
require_replace_all(
    effects,
    "EntityAttributeModifier.Operation.ADD_MULTIPLIED_BASE",
    "EntityAttributeModifier.Operation.MULTIPLY_BASE",
    3,
    "ArcherEffects operation names",
)
require_replace(
    effects,
    "EntityAttributes.GENERIC_MOVEMENT_SPEED.getIdAsString()",
    "Registries.ATTRIBUTE.getId(EntityAttributes.GENERIC_MOVEMENT_SPEED).toString()",
    "Entangling Roots movement attribute id",
)
require_replace(
    effects,
    '''                    new AttributeModifier(
                            EntityAttributes.GENERIC_JUMP_STRENGTH.getIdAsString(),
                            -0.5F,
                            EntityAttributeModifier.Operation.MULTIPLY_BASE
                    )''',
    '''                    // 1.20.1 has no generic jump-strength attribute. Jump suppression is applied
                    // by EntanglingRootsJumpMixin at LivingEntity#getJumpVelocity instead.''',
    "Entangling Roots jump compatibility bridge",
)
# Removing the second List.of element leaves a dangling comma after the movement modifier.
require_replace(
    effects,
    "EntityAttributeModifier.Operation.MULTIPLY_BASE\n                    ),\n                    // 1.20.1 has no generic jump-strength attribute.",
    "EntityAttributeModifier.Operation.MULTIPLY_BASE\n                    )\n                    // 1.20.1 has no generic jump-strength attribute.",
    "Entangling Roots single attribute list",
)

summons = java_root / "net/archers/entity/ArcherSummons.java"
require_replace(
    summons,
    "import net.minecraft.entity.attribute.EntityAttributes;\n",
    "import net.minecraft.entity.attribute.EntityAttributes;\nimport net.minecraft.registry.Registries;\n",
    "ArcherSummons registry import",
)
require_replace(
    summons,
    "ExternalSpellSchools.PHYSICAL_RANGED.attributeEntry.getIdAsString()",
    "Registries.ATTRIBUTE.getId(ExternalSpellSchools.PHYSICAL_RANGED.attribute).toString()",
    "ranged school owner attribute id",
)
for attribute in (
    "GENERIC_MAX_HEALTH",
    "GENERIC_ARMOR",
    "GENERIC_ATTACK_DAMAGE",
    "GENERIC_ATTACK_KNOCKBACK",
    "GENERIC_KNOCKBACK_RESISTANCE",
):
    require_replace(
        summons,
        f"EntityAttributes.{attribute}.getIdAsString()",
        f"Registries.ATTRIBUTE.getId(EntityAttributes.{attribute}).toString()",
        f"summon scaling {attribute} id",
    )
require_replace(
    summons,
    "EntityAttributeModifier.Operation.ADD_VALUE",
    "EntityAttributeModifier.Operation.ADDITION",
    "summon owner scaling operation",
)

# EntityType.Builder and living movement APIs have target-native 1.20.1 names. Generic jump/step
# attributes are newer than the target, so retain the exact current defaults as explicit compatibility
# constants used by SpiritWolfEntity rather than substituting semantically unrelated horse attributes.
entities = java_root / "net/archers/entity/ArcherEntities.java"
require_replace(entities, ".dimensions(0.8F, 0.925F)", ".setDimensions(0.8F, 0.925F)", "Spirit Wolf dimensions")
require_replace(
    entities,
    '''        // Flat +50% over the vanilla base values (jump 0.42, step 0.6) — agile like a wolf
        e.custom.add(new SummonedEntityConfig.CustomAttribute(
                EntityAttributes.GENERIC_JUMP_STRENGTH.getIdAsString(), 0.63));
        e.custom.add(new SummonedEntityConfig.CustomAttribute(
                EntityAttributes.GENERIC_STEP_HEIGHT.getIdAsString(), 0.9));
        return e;''',
    '''        // 1.20.1 has no generic living jump/step attributes. SpiritWolfEntity applies these
        // same current defaults through target-native movement hooks instead.
        return e;''',
    "Spirit Wolf unsupported generic movement attributes",
)
require_replace(
    entities,
    "public class ArcherEntities {\n",
    "public class ArcherEntities {\n    public static final float SPIRIT_WOLF_JUMP_VELOCITY = 0.63F;\n    public static final float SPIRIT_WOLF_STEP_HEIGHT = 0.9F;\n",
    "Spirit Wolf target movement constants",
)

spirit_wolf = java_root / "net/archers/entity/SpiritWolfEntity.java"
require_replace(
    spirit_wolf,
    "import net.minecraft.world.World;\n",
    "import net.minecraft.world.EntityView;\nimport net.minecraft.world.World;\n",
    "Spirit Wolf EntityView import",
)
require_replace(
    spirit_wolf,
    '''    public SpiritWolfEntity(EntityType<? extends SpiritWolfEntity> entityType, World world) {
        super(entityType, world);
    }
}''',
    '''    public SpiritWolfEntity(EntityType<? extends SpiritWolfEntity> entityType, World world) {
        super(entityType, world);
        setStepHeight(ArcherEntities.SPIRIT_WOLF_STEP_HEIGHT);
    }

    // Yarn 1.20.1's Tameable bridge has not yet been named getWorld; it is the same intermediary
    // method and returns the lookup view used by the default owner resolver.
    @Override
    public EntityView method_48926() {
        return getWorld();
    }

    @Override
    protected float getJumpVelocity() {
        return ArcherEntities.SPIRIT_WOLF_JUMP_VELOCITY;
    }
}''',
    "Spirit Wolf 1.20.1 tameable and movement hooks",
)

# Current SpellEngine's loader-neutral callback exposes a registry entry on 1.20.1; the vanilla
# enchantment constant itself is still the raw Enchantment object in this target.
archers_mod = java_root / "net/archers/ArchersMod.java"
require_replace(
    archers_mod,
    "enchantment.getKey().get().getValue().equals(Enchantments.INFINITY.getValue())",
    "enchantment.value() == Enchantments.INFINITY",
    "Infinity enchantment identity",
)

living_autofire = java_root / "net/archers/mixin/client/autofire/LivingEntityMixin.java"
require_replace(
    living_autofire,
    "ModelPredicateProviderRegistry.get(mainHandStack, new Identifier(\"pull\"))",
    "ModelPredicateProviderRegistry.get(mainHandStack.getItem(), new Identifier(\"pull\"))",
    "Auto Fire model predicate target item",
)

# Preserve Entangling Roots' modern -50% generic jump behavior on a target that has no corresponding
# attribute. This runs after LivingEntity computes its normal jump velocity, including Spirit Wolf's
# target-native override above, and scales once per effect level exactly like the former multiplier.
jump_mixin = java_root / "net/archers/mixin/effect/EntanglingRootsJumpMixin.java"
jump_mixin.parent.mkdir(parents=True, exist_ok=True)
jump_mixin.write_text('''package net.archers.mixin.effect;

import net.archers.effect.ArcherEffects;
import net.minecraft.entity.LivingEntity;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.injection.At;
import org.spongepowered.asm.mixin.injection.Inject;
import org.spongepowered.asm.mixin.injection.callback.CallbackInfoReturnable;

@Mixin(LivingEntity.class)
public abstract class EntanglingRootsJumpMixin {
    @Inject(method = "getJumpVelocity", at = @At("RETURN"), cancellable = true)
    private void archers$applyEntanglingRootsJumpReduction(CallbackInfoReturnable<Float> cir) {
        var entity = (LivingEntity) (Object) this;
        var effect = entity.getStatusEffect(ArcherEffects.ENTANGLING_ROOTS.effect);
        if (effect == null) {
            return;
        }
        float multiplier = Math.max(0.0F, 1.0F - 0.5F * (effect.getAmplifier() + 1));
        cir.setReturnValue(cir.getReturnValue() * multiplier);
    }
}
''', encoding="utf-8")

mixins = resources_root / "archers.mixins.json"
require_replace(
    mixins,
    '    "screen.GrindstoneScreenHandlerMixin"\n',
    '    "screen.GrindstoneScreenHandlerMixin",\n    "effect.EntanglingRootsJumpMixin"\n',
    "Entangling Roots jump mixin registration",
)

print("[Archers API wave2] current combat/entity ids and operation enums -> target-native 1.20.1 APIs")
print("[Archers API wave2] Spirit Wolf: dimensions + 0.63 jump + 0.9 step preserved without horse-attribute substitution")
print("[Archers API wave2] Entangling Roots: movement attribute + generic jump reduction preserved")
print("[Archers API wave2] current Direwolf render signature + custom emissive depth-writing layer adapted to 1.20.1")
print("[Archers API wave2] Infinity callback + Auto Fire model predicate adapted without feature loss")
