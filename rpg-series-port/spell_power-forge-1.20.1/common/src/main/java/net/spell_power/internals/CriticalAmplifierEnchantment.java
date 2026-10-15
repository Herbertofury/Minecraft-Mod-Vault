package net.spell_power.internals;

import net.minecraft.enchantment.Enchantment;
import net.minecraft.enchantment.EnchantmentTarget;
import net.minecraft.entity.EquipmentSlot;
import net.tiny_config.models.EnchantmentConfig;

/** 1.6 Spell Volatility / Amplify Spell exclusive-set behavior for 1.20.1. */
public final class CriticalAmplifierEnchantment extends AmplifierEnchantment {
    public CriticalAmplifierEnchantment(Rarity weight, Operation operation, EnchantmentConfig config,
                                        EnchantmentTarget type, EquipmentSlot[] slotTypes) {
        super(weight, operation, config, type, slotTypes);
    }

    @Override
    protected boolean canAccept(Enchantment other) {
        return !(other instanceof CriticalAmplifierEnchantment) && super.canAccept(other);
    }
}
