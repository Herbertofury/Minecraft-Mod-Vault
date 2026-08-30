package net.fabric_extras.ranged_weapon.api;

/** 2.x type-baseline API plus the proven 1.20.1 mutable config bridge. */
public interface CustomRangedWeapon {
    void setTypeBaseline(RangedConfig config);
    RangedConfig getTypeBaseline();
    RangedConfig getRangedWeaponConfig();
    void setRangedWeaponConfig(RangedConfig config);
    default void configure(RangedConfig config) { setRangedWeaponConfig(config); }
}
