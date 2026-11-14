package net.fabric_extras.ranged_weapon.api;

import net.minecraft.util.Identifier;

public final class AttributeModifierIDs {
    public static final String NAMESPACE = "ranged_weapon";
    public static final Identifier WEAPON_DAMAGE_ID = new Identifier(NAMESPACE, "base_damage");
    public static final Identifier WEAPON_PULL_TIME_ID = new Identifier(NAMESPACE, "base_pull_time");
    public static final Identifier WEAPON_VELOCITY_ID = new Identifier(NAMESPACE, "base_velocity");
    public static final Identifier OTHER_BONUS_ID = new Identifier(NAMESPACE, "ranged_weapon");
    private AttributeModifierIDs() {}
}
