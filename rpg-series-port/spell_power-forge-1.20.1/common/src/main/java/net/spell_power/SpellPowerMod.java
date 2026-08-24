package net.spell_power;

import net.minecraft.enchantment.Enchantment;
import net.minecraft.enchantment.EnchantmentHelper;
import net.minecraft.entity.LivingEntity;
import net.minecraft.entity.attribute.EntityAttributeModifier;
import net.minecraft.server.network.ServerPlayerEntity;
import net.minecraft.util.Identifier;
import net.spell_power.api.*;
import net.spell_power.api.enchantment.Enchantments_SpellPower;
import net.spell_power.api.enchantment.Enchantments_SpellPowerMechanics;
import net.spell_power.config.AttributesConfig;
import net.spell_power.config.EnchantmentsConfig;
import net.tinyconfig.ConfigManager;

public final class SpellPowerMod {
    public static final String ID = "spell_power";

    public static final ConfigManager<AttributesConfig> attributesConfig =
            new ConfigManager<AttributesConfig>("attributes", AttributesConfig.defaults())
                    .builder().setDirectory(ID).sanitize(true).validate(AttributesConfig::isValid).build();

    public static final ConfigManager<EnchantmentsConfig> enchantmentConfig =
            new ConfigManager<EnchantmentsConfig>("enchantments", new EnchantmentsConfig())
                    .builder().setDirectory(ID).sanitize(true).schemaVersion(5).build();

    private SpellPowerMod() {}

    public static void refreshConfigs() { attributesConfig.refresh(); enchantmentConfig.refresh(); }

    public static void prepareStatusEffects() {
        var powerCfg = attributesConfig.value.spell_power_effect;
        for (var school : SpellSchools.all()) {
            if (school.powerEffectManagement.isInternal() && school.boostEffect != null) {
                school.boostEffect.addAttributeModifier(school.attribute, powerCfg.uuid, powerCfg.bonus_per_stack,
                        EntityAttributeModifier.Operation.MULTIPLY_BASE);
            }
        }
        for (var entry : SpellPowerMechanics.all.values()) {
            var cfg = attributesConfig.value.secondary_effects.get(entry.name);
            if (cfg != null) {
                entry.boostEffect.addAttributeModifier(entry.attribute, cfg.uuid, cfg.bonus_per_stack,
                        EntityAttributeModifier.Operation.MULTIPLY_BASE);
            }
        }
    }

    public static void applyEnchantments() {
        enchantmentConfig.value.apply();
        attachSpecializedEnchantmentsToSchools();
    }

    private static void attachSpecializedEnchantmentsToSchools() {
        for (var school : SpellSchools.all()) {
            var relevant = Enchantments_SpellPower.all.values().stream()
                    .filter(enchantment -> enchantment != Enchantments_SpellPower.SPELL_POWER)
                    .filter(enchantment -> enchantment.poweredSchools().contains(school))
                    .toList();
            school.addSource(SpellSchool.Trait.POWER, new SpellSchool.Source(SpellSchool.Apply.ADD, query -> {
                double value = 0;
                for (var enchantment : relevant) {
                    int level = armorEnchantmentLevel(enchantment, query.entity());
                    value += school.attribute.getDefaultValue() * enchantment.config.bonus_per_level * level;
                }
                return value;
            }));
        }
    }

    public static int mainHandEnchantmentLevel(Enchantment enchantment, LivingEntity entity) {
        var stack = entity.getMainHandStack();
        return stack == null || stack.isEmpty() ? 0 : EnchantmentHelper.getLevel(enchantment, stack);
    }

    public static int armorEnchantmentLevel(Enchantment enchantment, LivingEntity entity) {
        int level = 0;
        for (var stack : entity.getArmorItems()) level += EnchantmentHelper.getLevel(enchantment, stack);
        return level;
    }

    public static double genericSpellPowerEnchantBonus(LivingEntity entity) {
        int level = mainHandEnchantmentLevel(Enchantments_SpellPower.SPELL_POWER, entity);
        return level * Enchantments_SpellPower.SPELL_POWER.config.bonus_per_level;
    }

    public static void onPlayerJoin(ServerPlayerEntity player) {
        if (attributesConfig.value.migrate_attributes_base) migrateAttributes(player);
    }

    public static void migrateAttributes(ServerPlayerEntity player) {
        for (var school : SpellSchools.all()) {
            if (school.archetype != SpellSchool.Archetype.MAGIC || !school.attributeManagement.isInternal()) continue;
            var instance = player.getAttributeInstance(school.attribute);
            if (instance == null) continue;
            double defaultValue = school.attribute.getDefaultValue();
            if (instance.getBaseValue() != defaultValue) instance.setBaseValue(defaultValue);
        }
    }

    public static Identifier potionIdFrom(Identifier id) {
        return new Identifier(id.getNamespace(), id.getNamespace() + "." + id.getPath());
    }

    @Deprecated public static AttributesConfig.AttributeScope attributeScopeOverride = null;
    @Deprecated public static AttributesConfig.AttributeScope attributeScope() { return AttributesConfig.AttributeScope.LIVING_ENTITY; }

    public static void touchRegistries() {
        SpellPowerMechanics.all.size(); SpellSchools.all().size(); SpellResistance.Attributes.all.size();
        Enchantments_SpellPowerMechanics.all.size(); Enchantments_SpellPower.all.size();
    }
}
