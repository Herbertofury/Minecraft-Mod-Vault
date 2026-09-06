package net.spell_power.internals;

import net.minecraft.enchantment.Enchantment;
import net.minecraft.enchantment.EnchantmentTarget;
import net.minecraft.entity.EquipmentSlot;
import net.minecraft.item.ItemStack;
import net.minecraft.util.Identifier;
import net.spell_power.api.SpellSchool;
import net.spell_power.api.enchantment.SpellPowerEnchanting;
import net.spell_power.config.EnchantmentsConfig;

import java.util.Set;

public class SchoolFilteredEnchantment extends AmplifierEnchantment {
    private Set<SpellSchool> schools;

    public Set<SpellSchool> poweredSchools() { return schools; }
    public void setPoweredSchools(Set<SpellSchool> schools) { this.schools = schools; }

    public SchoolFilteredEnchantment(Rarity weight, Operation operation,
                                     EnchantmentsConfig.PowerEnchantmentConfig config,
                                     Set<SpellSchool> schools, EnchantmentTarget type,
                                     EquipmentSlot[] slotTypes) {
        super(weight, operation, config, type, slotTypes);
        this.schools = schools;
    }

    @Override
    public SchoolFilteredEnchantment requireTag(Identifier tagId) {
        this.tagId = tagId;
        return this;
    }

    @Override
    protected boolean canAccept(Enchantment other) {
        if (other instanceof SchoolFilteredEnchantment filtered
                && isSpecialized() && filtered.isSpecialized()) {
            return false;
        }
        return super.canAccept(other);
    }

    public boolean requiresRelatedAttributes() {
        return ((EnchantmentsConfig.PowerEnchantmentConfig) config).requires_related_attributes;
    }

    public boolean isSpecialized() {
        return ((EnchantmentsConfig.PowerEnchantmentConfig) config).exclusive_with_specialized;
    }

    public static boolean schoolsIntersect(Set<SpellSchool> schools, ItemStack stack) {
        var itemStackSchools = SpellPowerEnchanting.relevantSchools(stack);
        for (var school : itemStackSchools) {
            if (schools.contains(school)) return true;
        }
        return false;
    }
}
