package net.fabric_extras.ranged_weapon.internal;

import net.minecraft.item.BowItem;
import net.minecraft.item.CrossbowItem;
import net.minecraft.item.Item;

public final class ScalingUtil {
    public static final float STANDARD_BOW_VELOCITY = 3F;
    public static final float STANDARD_BOW_DAMAGE = 6F;
    public static final float STANDARD_CROSSBOW_VELOCITY = 3.15F;
    public static final float STANDARD_CROSSBOW_DAMAGE = 9F;
    public static final Scaling BOW_BASELINE = new Scaling(STANDARD_BOW_VELOCITY, STANDARD_BOW_DAMAGE);
    public static final Scaling CROSSBOW_BASELINE = new Scaling(STANDARD_CROSSBOW_VELOCITY, STANDARD_CROSSBOW_DAMAGE);
    public record Scaling(double velocity, double damage) {}
    public static Scaling baselineFor(Item item) {
        if (item instanceof BowItem) return BOW_BASELINE;
        if (item instanceof CrossbowItem) return CROSSBOW_BASELINE;
        return new Scaling(1, 1);
    }
    public static double arrowVelocityMultiplier(double standardVelocity, double customVelocity) { return customVelocity / standardVelocity; }
    public static double arrowVelocityMultiplier(Item item, double bonusVelocity) { var b=baselineFor(item); return (b.velocity()+bonusVelocity)/b.velocity(); }
    public static double arrowDamageMultiplier(double standardDamage, double attributeDamage, double velocityMultiplier) {
        var multiplier = attributeDamage / standardDamage;
        if (velocityMultiplier != 1) multiplier /= velocityMultiplier;
        return multiplier;
    }
    private ScalingUtil() {}
}
