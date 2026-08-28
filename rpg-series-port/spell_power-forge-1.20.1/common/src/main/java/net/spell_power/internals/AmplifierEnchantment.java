package net.spell_power.internals;

import net.minecraft.enchantment.Enchantment;
import net.minecraft.enchantment.EnchantmentTarget;
import net.minecraft.entity.EquipmentSlot;
import net.minecraft.item.ItemStack;
import net.minecraft.registry.RegistryKeys;
import net.minecraft.registry.tag.TagKey;
import net.minecraft.util.Identifier;
import net.spell_power.config.EnchantmentsConfig;
import net.tiny_config.models.EnchantmentConfig;
import org.jetbrains.annotations.Nullable;

public class AmplifierEnchantment extends Enchantment {
    public Operation operation;
    @Nullable protected Identifier tagId;

    public enum Operation { ADD, MULTIPLY }

    public EnchantmentConfig config;

    public double amplified(double value, int level) {
        return switch (operation) {
            case ADD -> value + ((float) level) * config.bonus_per_level;
            case MULTIPLY -> value * (1F + ((float) level) * config.bonus_per_level);
        };
    }

    public AmplifierEnchantment(Rarity weight, Operation operation, EnchantmentConfig config,
                                EnchantmentTarget type, EquipmentSlot[] slotTypes) {
        super(weight, type, slotTypes);
        this.operation = operation;
        this.config = config;
    }

    public AmplifierEnchantment requireTag(Identifier tagId) {
        this.tagId = tagId;
        return this;
    }

    @Override
    public boolean isAvailableForEnchantedBookOffer() {
        return super.isAvailableForEnchantedBookOffer() && config.enabled;
    }

    @Override
    public boolean isAvailableForRandomSelection() {
        return super.isAvailableForRandomSelection() && config.enabled;
    }

    @Override
    public int getMaxLevel() {
        return config.enabled ? config.max_level : 0;
    }

    @Override
    public int getMinPower(int level) {
        return config.min_cost + (level - 1) * config.step_cost;
    }

    @Override
    public int getMaxPower(int level) {
        if (config instanceof EnchantmentsConfig.ModernEnchantmentConfig modern) {
            return modern.max_cost + (level - 1) * modern.max_step_cost;
        }
        return super.getMinPower(level) + 50;
    }

    // 1.6 no longer carries the old blanket bow/crossbow incompatibility here.
    @Override
    protected boolean canAccept(Enchantment other) {
        return super.canAccept(other);
    }

    public boolean matchesRequiredTag(ItemStack stack) {
        return tagId == null || stack.isIn(TagKey.of(RegistryKeys.ITEM, tagId));
    }
}
