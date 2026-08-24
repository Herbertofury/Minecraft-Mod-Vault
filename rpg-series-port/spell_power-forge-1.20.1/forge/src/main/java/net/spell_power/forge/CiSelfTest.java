package net.spell_power.forge;

import net.minecraft.entity.EntityType;
import net.minecraft.entity.EquipmentSlot;
import net.minecraft.entity.attribute.EntityAttributeModifier;
import net.minecraft.item.ItemStack;
import net.minecraft.item.Items;
import net.minecraft.registry.Registries;
import net.minecraftforge.event.server.ServerStartedEvent;
import net.spell_power.SpellPowerMod;
import net.spell_power.api.*;
import net.spell_power.api.enchantment.Enchantments_SpellPower;
import net.spell_power.api.enchantment.Enchantments_SpellPowerMechanics;
import java.util.UUID;

public final class CiSelfTest {
    private CiSelfTest() {}
    public static void onServerStarted(ServerStartedEvent event) {
        var server = event.getServer();
        require("spell_power".equals(Registries.ATTRIBUTE.getId(SpellSchools.ARCANE.attribute).getNamespace()), "arcane attribute registry");
        require(Registries.ATTRIBUTE.getId(SpellSchools.GENERIC.attribute).getPath().equals("generic"), "generic attribute registry");
        require(SpellPowerMechanics.HASTE.min == 10F, "sub-1 haste floor");
        require(SpellPowerMod.attributesConfig.value.resistance_curve == DamageCurve.HYPERBOLIC, "hyperbolic resistance default");
        require(SpellPowerMod.attributesConfig.value.enchantments_require_matching_attribute, "matching-attribute default");
        require(Enchantments_SpellPowerMechanics.CRITICAL_CHANCE.getMaxLevel() == 3, "Spell Volatility max level");
        require(close(Enchantments_SpellPowerMechanics.CRITICAL_CHANCE.config.bonus_per_level, 0.04), "Spell Volatility bonus");
        require(Enchantments_SpellPowerMechanics.CRITICAL_DAMAGE.getMaxLevel() == 3, "Amplify Spell max level");
        require(close(Enchantments_SpellPowerMechanics.CRITICAL_DAMAGE.config.bonus_per_level, 0.10), "Amplify Spell bonus");
        require(Enchantments_SpellPowerMechanics.HASTE.getMaxLevel() == 3, "Spell Haste max level");
        var zombie = EntityType.ZOMBIE.create(server.getOverworld());
        require(zombie != null && zombie.getAttributeInstance(SpellSchools.ARCANE.attribute) != null, "living entity spell attributes");
        require(zombie.getAttributeInstance(SpellResistance.Attributes.GENERIC.attribute) != null, "living entity resistance attribute");
        var arcane = zombie.getAttributeInstance(SpellSchools.ARCANE.attribute);
        arcane.addTemporaryModifier(new EntityAttributeModifier(UUID.fromString("2ea3db9e-6bd4-4a4b-b267-97bdbaf29f61"), "spell_power.ci.flat", 10, EntityAttributeModifier.Operation.ADDITION));
        var sword = new ItemStack(Items.IRON_SWORD);
        sword.addEnchantment(Enchantments_SpellPower.SPELL_POWER, 1);
        sword.addEnchantment(Enchantments_SpellPowerMechanics.CRITICAL_CHANCE, 1);
        zombie.equipStack(EquipmentSlot.MAINHAND, sword);
        var chest = new ItemStack(Items.IRON_CHESTPLATE);
        chest.addEnchantment(Enchantments_SpellPower.SUNFIRE, 1);
        chest.addEnchantment(Enchantments_SpellPowerMechanics.CRITICAL_CHANCE, 3);
        zombie.equipStack(EquipmentSlot.CHEST, chest);
        var result = SpellPower.getSpellPower(SpellSchools.ARCANE, zombie);
        require(close(result.baseValue(), 11.53), "generic multiplier + armor specialized power");
        require(close(result.criticalChance(), 0.09), "weapon-only Spell Volatility + innate crit");
        require(result.forcedCritical().isCritical(), "Result exposes isCritical");
        System.out.println("[Spell Power CI] Runtime self-test passed: registries + living attributes + 1.6 enchant/multiplier mechanics");
    }
    private static boolean close(double actual, double expected) { return Math.abs(actual - expected) < 0.0001; }
    private static void require(boolean ok, String label) { if (!ok) throw new IllegalStateException("[Spell Power CI] Failed: " + label); }
}
