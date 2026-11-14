package net.fabric_extras.ranged_weapon.api;

import net.fabric_extras.ranged_weapon.internal.ScalingUtil;
import net.minecraft.entity.attribute.EntityAttributeModifier;
import net.minecraft.util.Identifier;
import org.jetbrains.annotations.Nullable;
import java.util.ArrayList;
import java.util.List;

/** 2.3.4 configuration surface, adapted to Minecraft 1.20.1 attribute operations. */
public record RangedConfig(float damage, float pull_time_bonus, float velocity_bonus, @Nullable List<Attribute> attributes) {
    public static final RangedConfig EMPTY = new RangedConfig(0, 0, 0);
    public static final RangedConfig BOW = new RangedConfig((float) ScalingUtil.BOW_BASELINE.damage(), 0, 0);
    public static final RangedConfig CROSSBOW = new RangedConfig((float) ScalingUtil.CROSSBOW_BASELINE.damage(), 0.25F, 0);

    public RangedConfig(float damage, float pullTimeBonus, float velocityBonus) {
        this(damage, pullTimeBonus, velocityBonus, null);
    }
    public record Attribute(String attributeId, Modifier modifier) {}
    public record Modifier(String modifierId, EntityAttributeModifier.Operation operation, double value) {}
    public RangedConfig withAttributes(@Nullable List<Attribute> values) { return new RangedConfig(damage, pull_time_bonus, velocity_bonus, values); }
    public RangedConfig withAttribute(Identifier attributeId, Identifier modifierId, EntityAttributeModifier.Operation operation, double value) {
        var list = new ArrayList<>(attributes != null ? attributes : List.of());
        list.add(new Attribute(attributeId.toString(), new Modifier(modifierId.toString(), operation, value)));
        return withAttributes(list);
    }
    public RangedConfig withAttribute(Identifier attributeId, EntityAttributeModifier.Operation operation, double value) {
        return withAttribute(attributeId, AttributeModifierIDs.OTHER_BONUS_ID, operation, value);
    }
    public float pullTimeSeconds() { return Math.max(0.1F, 1F + pull_time_bonus); }
}
