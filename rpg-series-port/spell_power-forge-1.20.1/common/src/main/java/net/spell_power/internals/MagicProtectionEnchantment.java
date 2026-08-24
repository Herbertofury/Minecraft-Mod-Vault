package net.spell_power.internals;

import net.minecraft.enchantment.Enchantment;
import net.minecraft.enchantment.ProtectionEnchantment;
import net.minecraft.entity.EquipmentSlot;
import net.minecraft.entity.damage.DamageSource;
import net.minecraft.entity.damage.DamageType;
import net.minecraft.registry.RegistryKeys;
import net.minecraft.registry.tag.DamageTypeTags;
import net.minecraft.registry.tag.TagKey;
import net.minecraft.util.Identifier;
import net.tinyconfig.models.EnchantmentConfig;

public class MagicProtectionEnchantment extends ProtectionEnchantment {
    private static final TagKey<DamageType> IS_MAGIC =
            TagKey.of(RegistryKeys.DAMAGE_TYPE, new Identifier("c", "is_magic"));

    public EnchantmentConfig config;

    public MagicProtectionEnchantment(Rarity weight, EnchantmentConfig config, EquipmentSlot... slotTypes) {
        super(weight, ProtectionEnchantment.Type.ALL, slotTypes);
        this.config = config;
    }

    @Override
    public int getMinPower(int level) {
        return config.min_cost + (level - 1) * config.step_cost;
    }

    @Override
    public int getMaxPower(int level) {
        return getMinPower(level) + config.step_cost;
    }

    @Override
    public int getMaxLevel() { return config.max_level; }

    @Override
    public int getProtectionAmount(int level, DamageSource source) {
        if (source.isIn(DamageTypeTags.BYPASSES_INVULNERABILITY)) return 0;
        return source.isIn(IS_MAGIC) ? Math.round((float) level * config.bonus_per_level) : 0;
    }

    @Override
    public boolean canAccept(Enchantment other) {
        if (other instanceof MagicProtectionEnchantment) return false;
        if (other instanceof ProtectionEnchantment protection) {
            return this.protectionType == ProtectionEnchantment.Type.FALL
                    || protection.protectionType == ProtectionEnchantment.Type.FALL;
        }
        return super.canAccept(other);
    }
}
