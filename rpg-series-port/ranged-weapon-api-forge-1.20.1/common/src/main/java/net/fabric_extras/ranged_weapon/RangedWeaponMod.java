package net.fabric_extras.ranged_weapon;

import net.fabric_extras.ranged_weapon.api.*;
import net.minecraft.entity.attribute.EntityAttributeModifier;
import net.minecraft.item.Items;
import net.minecraft.util.Identifier;

public final class RangedWeaponMod {
    public static final String NAMESPACE="ranged_weapon";
    public static final String ID=NAMESPACE+"_api";
    private static boolean initialized;
    private static boolean potionsRequested;
    private RangedWeaponMod() {}

    public static synchronized void init() {
        if (initialized) return;
        initialized=true;
        configureVanilla((CustomRangedWeapon)(Object)Items.BOW,RangedConfig.BOW);
        configureVanilla((CustomRangedWeapon)(Object)Items.CROSSBOW,RangedConfig.CROSSBOW);
    }
    private static void configureVanilla(CustomRangedWeapon weapon,RangedConfig config) { weapon.setTypeBaseline(config); weapon.setRangedWeaponConfig(config); }
    public static void prepareStatusEffects() {
        StatusEffects_RangedWeapon.DAMAGE.effect.addAttributeModifier(EntityAttributes_RangedWeapon.DAMAGE.attribute,"42f4bc42-260f-11ed-a261-0242ac120002",0.1,EntityAttributeModifier.Operation.MULTIPLY_BASE);
        StatusEffects_RangedWeapon.HASTE.effect.addAttributeModifier(EntityAttributes_RangedWeapon.HASTE.attribute,"607c251c-260f-11ed-a261-0242ac120002",0.1,EntityAttributeModifier.Operation.MULTIPLY_BASE);
    }
    /** Exposed 2.x helper. Forge fulfills the request in the potion registry phase. */
    public static void registerPotions() { potionsRequested=true; }
    public static boolean potionsRequested() { return potionsRequested; }
    public static Identifier potionIdFrom(Identifier id) { return new Identifier(id.getNamespace(),id.getNamespace()+"."+id.getPath()); }
    public static void bindRegistryEntries() {
        for (var e:EntityAttributes_RangedWeapon.all) e.bindRegistryEntry();
        for (var e:StatusEffects_RangedWeapon.all) e.bindRegistryEntry();
    }
}
