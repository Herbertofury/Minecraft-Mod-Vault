package net.spell_power.api.enchantment;

import net.minecraft.enchantment.Enchantment;
import net.minecraft.entity.EquipmentSlot;
import net.minecraft.util.Identifier;
import net.spell_power.SpellPowerMod;
import net.spell_power.api.SpellPowerMechanics;
import net.spell_power.config.EnchantmentsConfig;
import net.spell_power.internals.AmplifierEnchantment;
import net.spell_power.internals.CriticalAmplifierEnchantment;
import net.spell_power.internals.MagicProtectionEnchantment;

import java.util.HashMap;
import java.util.Map;

import static net.minecraft.enchantment.EnchantmentTarget.BREAKABLE;
import static net.spell_power.internals.AmplifierEnchantment.Operation.ADD;

public class Enchantments_SpellPowerMechanics {
    private static final EquipmentSlot[] MAIN_HAND = { EquipmentSlot.MAINHAND };
    private static final EquipmentSlot[] ARMOR = {
            EquipmentSlot.HEAD, EquipmentSlot.CHEST, EquipmentSlot.LEGS, EquipmentSlot.FEET
    };

    public static final Identifier criticalChanceId = new Identifier(SpellPowerMod.ID, SpellPowerMechanics.CRITICAL_CHANCE.name);
    public static final AmplifierEnchantment CRITICAL_CHANCE = new CriticalAmplifierEnchantment(
            Enchantment.Rarity.UNCOMMON, ADD, config().critical_chance, BREAKABLE, MAIN_HAND)
            .requireTag(new Identifier(SpellPowerMod.ID, "enchant_critical_chance"));

    public static final Identifier criticalDamageId = new Identifier(SpellPowerMod.ID, SpellPowerMechanics.CRITICAL_DAMAGE.name);
    public static final AmplifierEnchantment CRITICAL_DAMAGE = new CriticalAmplifierEnchantment(
            Enchantment.Rarity.UNCOMMON, ADD, config().critical_damage, BREAKABLE, MAIN_HAND)
            .requireTag(new Identifier(SpellPowerMod.ID, "enchant_critical_damage"));

    public static final Identifier hasteId = new Identifier(SpellPowerMod.ID, SpellPowerMechanics.HASTE.name);
    public static final AmplifierEnchantment HASTE = new AmplifierEnchantment(
            Enchantment.Rarity.UNCOMMON, ADD, config().haste, BREAKABLE, MAIN_HAND)
            .requireTag(new Identifier(SpellPowerMod.ID, "enchant_haste"));

    public static final String magicProtectionName = "magic_protection";
    public static final Identifier magicProtectionId = new Identifier(SpellPowerMod.ID, magicProtectionName);
    public static final MagicProtectionEnchantment MAGIC_PROTECTION = new MagicProtectionEnchantment(
            Enchantment.Rarity.UNCOMMON, config().magic_protection, ARMOR);

    public static final Map<Identifier, Enchantment> all;
    static {
        Map<Identifier, AmplifierEnchantment> secondaries = new HashMap<>();
        secondaries.put(criticalChanceId, CRITICAL_CHANCE);
        secondaries.put(criticalDamageId, CRITICAL_DAMAGE);
        secondaries.put(hasteId, HASTE);

        all = new HashMap<>();
        all.putAll(secondaries);
        all.put(magicProtectionId, MAGIC_PROTECTION);

        for (var entry : secondaries.entrySet()) {
            var enchantment = entry.getValue();
            EnchantmentRestriction.prohibit(enchantment,
                    itemStack -> !enchantment.matchesRequiredTag(itemStack));
        }
    }

    private static EnchantmentsConfig config() { return SpellPowerMod.enchantmentConfig.value; }
}
