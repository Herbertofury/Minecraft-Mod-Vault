package net.spell_power.api.enchantment;

import net.minecraft.enchantment.Enchantment;
import net.minecraft.entity.EquipmentSlot;
import net.minecraft.util.Identifier;
import net.spell_power.SpellPowerMod;
import net.spell_power.api.SpellSchools;
import net.spell_power.config.EnchantmentsConfig;
import net.spell_power.internals.SchoolFilteredEnchantment;

import java.util.HashMap;
import java.util.Map;
import java.util.Set;

import static net.minecraft.enchantment.EnchantmentTarget.BREAKABLE;
import static net.spell_power.internals.AmplifierEnchantment.Operation.ADD;

public class Enchantments_SpellPower {
    private static final EquipmentSlot[] MAIN_HAND = { EquipmentSlot.MAINHAND };
    private static final EquipmentSlot[] ARMOR = {
            EquipmentSlot.HEAD, EquipmentSlot.CHEST, EquipmentSlot.LEGS, EquipmentSlot.FEET
    };

    public static final String spellPowerName = "spell_power";
    public static final Identifier spellPowerId = new Identifier(SpellPowerMod.ID, spellPowerName);
    public static final SchoolFilteredEnchantment SPELL_POWER = new SchoolFilteredEnchantment(
            Enchantment.Rarity.COMMON, ADD, config().spell_power,
            Set.of(SpellSchools.ARCANE, SpellSchools.FIRE, SpellSchools.FROST,
                    SpellSchools.HEALING, SpellSchools.LIGHTNING, SpellSchools.SOUL),
            BREAKABLE, MAIN_HAND)
            .requireTag(new Identifier(SpellPowerMod.ID, "enchant_spell_power_generic"));

    public static final String soulfrostName = "soulfrost";
    public static final Identifier soulfrostId = new Identifier(SpellPowerMod.ID, soulfrostName);
    public static final SchoolFilteredEnchantment SOULFROST = new SchoolFilteredEnchantment(
            Enchantment.Rarity.RARE, ADD, config().soulfrost,
            Set.of(SpellSchools.SOUL, SpellSchools.FROST), BREAKABLE, ARMOR)
            .requireTag(new Identifier(SpellPowerMod.ID, "enchant_spell_power_soulfrost"));

    public static final String sunfireName = "sunfire";
    public static final Identifier sunfireId = new Identifier(SpellPowerMod.ID, sunfireName);
    public static final SchoolFilteredEnchantment SUNFIRE = new SchoolFilteredEnchantment(
            Enchantment.Rarity.RARE, ADD, config().sunfire,
            Set.of(SpellSchools.ARCANE, SpellSchools.FIRE), BREAKABLE, ARMOR)
            .requireTag(new Identifier(SpellPowerMod.ID, "enchant_spell_power_sunfire"));

    public static final String energizeName = "energize";
    public static final Identifier energizeId = new Identifier(SpellPowerMod.ID, energizeName);
    public static final SchoolFilteredEnchantment ENERGIZE = new SchoolFilteredEnchantment(
            Enchantment.Rarity.RARE, ADD, config().energize,
            Set.of(SpellSchools.HEALING, SpellSchools.LIGHTNING), BREAKABLE, ARMOR)
            .requireTag(new Identifier(SpellPowerMod.ID, "enchant_spell_power_energize"));

    public static final Map<Identifier, SchoolFilteredEnchantment> all;
    static {
        all = new HashMap<>();
        all.put(spellPowerId, SPELL_POWER);
        all.put(soulfrostId, SOULFROST);
        all.put(sunfireId, SUNFIRE);
        all.put(energizeId, ENERGIZE);

        for (var entry : all.entrySet()) {
            var enchantment = entry.getValue();
            EnchantmentRestriction.prohibit(enchantment, itemStack -> {
                var typeMatches = enchantment.matchesRequiredTag(itemStack);
                var matchingAttributesEnabled = SpellPowerMod.attributesConfig.value.enchantments_require_matching_attribute;
                var schoolMatches = !matchingAttributesEnabled
                        || !enchantment.requiresRelatedAttributes()
                        || SchoolFilteredEnchantment.schoolsIntersect(enchantment.poweredSchools(), itemStack);
                return !typeMatches || !schoolMatches;
            });
        }
    }

    private static EnchantmentsConfig config() { return SpellPowerMod.enchantmentConfig.value; }
}
