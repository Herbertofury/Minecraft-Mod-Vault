package net.spell_power.config;

import net.spell_power.api.enchantment.Enchantments_SpellPower;
import net.spell_power.api.enchantment.Enchantments_SpellPowerMechanics;
import net.tinyconfig.models.EnchantmentConfig;
import net.tinyconfig.versioning.VersionableConfig;

public class EnchantmentsConfig extends VersionableConfig {
    /** Kept for source/config compatibility; 1.6 slot behavior is enforced explicitly. */
    @Deprecated public boolean allow_stacking = true;

    public PowerEnchantmentConfig spell_power = new PowerEnchantmentConfig(false, 5, 1, 11, 12, 11, 0.05F);
    public PowerEnchantmentConfig soulfrost = new PowerEnchantmentConfig(true, 5, 1, 11, 12, 11, 0.03F);
    public PowerEnchantmentConfig sunfire = new PowerEnchantmentConfig(true, 5, 1, 11, 12, 11, 0.03F);
    public PowerEnchantmentConfig energize = new PowerEnchantmentConfig(true, 5, 1, 11, 12, 11, 0.03F);

    // 1.6 balance: weapon-only, max level 3. Spell Volatility is +4%/level;
    // Amplify Spell is +10%/level; Haste is +4%/level.
    public ModernEnchantmentConfig critical_chance = new ModernEnchantmentConfig(3, 5, 12, 15, 15, 0.04F);
    public ModernEnchantmentConfig critical_damage = new ModernEnchantmentConfig(3, 5, 12, 15, 15, 0.10F);
    public ModernEnchantmentConfig haste = new ModernEnchantmentConfig(3, 5, 12, 15, 15, 0.04F);
    public EnchantmentConfig magic_protection = new EnchantmentConfig(4, 3, 6, 2);

    public void apply() {
        Enchantments_SpellPowerMechanics.CRITICAL_CHANCE.config = critical_chance;
        Enchantments_SpellPowerMechanics.CRITICAL_DAMAGE.config = critical_damage;
        Enchantments_SpellPowerMechanics.HASTE.config = haste;
        Enchantments_SpellPowerMechanics.MAGIC_PROTECTION.config = magic_protection;
        Enchantments_SpellPower.SPELL_POWER.config = spell_power;
        Enchantments_SpellPower.SOULFROST.config = soulfrost;
        Enchantments_SpellPower.SUNFIRE.config = sunfire;
        Enchantments_SpellPower.ENERGIZE.config = energize;
    }

    public static class ModernEnchantmentConfig extends EnchantmentConfig {
        public int max_cost;
        public int max_step_cost;

        public ModernEnchantmentConfig(int maxLevel, int minCost, int minStepCost,
                                       int maxCost, int maxStepCost, float bonusPerLevel) {
            super(maxLevel, minCost, minStepCost, bonusPerLevel);
            this.max_cost = maxCost;
            this.max_step_cost = maxStepCost;
        }
    }

    public static class PowerEnchantmentConfig extends ModernEnchantmentConfig {
        public boolean requires_related_attributes = false;
        /** Modern exclusive-set semantics: specialized schools exclude each other, not generic Spell Power. */
        public boolean exclusive_with_specialized = false;

        public PowerEnchantmentConfig(boolean requiresRelatedAttributes, int maxLevel,
                                      int minCost, int minStepCost, int maxCost,
                                      int maxStepCost, float bonusPerLevel) {
            super(maxLevel, minCost, minStepCost, maxCost, maxStepCost, bonusPerLevel);
            this.requires_related_attributes = requiresRelatedAttributes;
            this.exclusive_with_specialized = requiresRelatedAttributes;
        }

        /** Legacy constructor retained for downstream 1.20.1 source compatibility. */
        public PowerEnchantmentConfig(boolean requiresRelatedAttributes, int maxLevel,
                                      int minCost, int stepCost, float bonusPerLevel) {
            this(requiresRelatedAttributes, maxLevel, minCost, stepCost,
                    minCost + stepCost, stepCost, bonusPerLevel);
        }
    }
}
