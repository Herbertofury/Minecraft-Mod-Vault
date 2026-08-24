package net.spell_power.api;

import net.minecraft.entity.attribute.EntityAttribute;
import net.minecraft.entity.damage.DamageTypes;
import net.minecraft.entity.effect.StatusEffect;
import net.minecraft.entity.effect.StatusEffectCategory;
import net.minecraft.registry.RegistryKey;
import net.minecraft.registry.RegistryKeys;
import net.minecraft.util.Identifier;
import net.spell_power.SpellPowerMod;
import net.spell_power.api.enchantment.Enchantments_SpellPowerMechanics;
import net.spell_power.internals.CustomEntityAttribute;
import net.spell_power.internals.SpellStatusEffect;
import org.jetbrains.annotations.Nullable;
import java.util.*;
import static net.spell_power.api.SpellPowerMechanics.PERCENT_ATTRIBUTE_BASELINE;

public class SpellSchools {
    public static final String DEFAULT_NAMESPACE = SpellPowerMod.ID;
    private static final LinkedHashMap<Identifier, SpellSchool> REGISTRY = new LinkedHashMap<>();
    public static SpellSchool register(SpellSchool school) { REGISTRY.put(school.id, school); return school; }
    public static Set<SpellSchool> all() { return new LinkedHashSet<>(REGISTRY.values()); }

    public static final SpellSchool GENERIC = register(createMagic("generic", true, 0x9999BB, 100F));
    public static final SpellSchool ARCANE = register(createMagic("arcane", true, 0xff66ff, SpellPowerMod.attributesConfig.value.base_spell_power));
    public static final SpellSchool FIRE = register(createMagic("fire", true, 0xff3300, SpellPowerMod.attributesConfig.value.base_spell_power));
    public static final SpellSchool FROST = register(createMagic("frost", true, 0xccffff, SpellPowerMod.attributesConfig.value.base_spell_power));
    public static final SpellSchool HEALING = register(createMagic("healing", true, 0x66ff66, SpellPowerMod.attributesConfig.value.base_spell_power));
    public static final SpellSchool LIGHTNING = register(createMagic("lightning", true, 0xffff99, SpellPowerMod.attributesConfig.value.base_spell_power));
    public static final SpellSchool SOUL = register(createMagic("soul", true, 0x2dd4da, SpellPowerMod.attributesConfig.value.base_spell_power));

    public static SpellSchool createMagic(String name, int color) { return createMagic(name, true, color, SpellPowerMod.attributesConfig.value.base_spell_power); }
    public static SpellSchool createMagic(String name, boolean customDamageType, int color) { return createMagic(name, customDamageType, color, SpellPowerMod.attributesConfig.value.base_spell_power); }
    public static SpellSchool createMagic(String name, boolean customDamageType, int color, float base) { return createMagic(new Identifier(DEFAULT_NAMESPACE, name.toLowerCase(Locale.ROOT)), customDamageType, color, base); }
    public static SpellSchool createMagic(Identifier id, int color) { return createMagic(id, true, color, SpellPowerMod.attributesConfig.value.base_spell_power); }
    public static SpellSchool createMagic(Identifier id, boolean customDamageType, int color) { return createMagic(id, customDamageType, color, SpellPowerMod.attributesConfig.value.base_spell_power); }
    public static SpellSchool createMagic(Identifier id, boolean customDamageType, int color, float base) {
        var effect = new SpellStatusEffect(StatusEffectCategory.BENEFICIAL, color);
        var attr = (EntityAttribute) new CustomEntityAttribute("attribute.name." + id.getNamespace() + "." + id.getPath(), base, 0, 2048, id).setTracked(true);
        return createMagic(id, color, customDamageType, attr, effect);
    }
    @Deprecated public static SpellSchool createMagic(Identifier id, int color, EntityAttribute attr, StatusEffect effect) { return createMagic(id, color, false, attr, effect); }
    public static SpellSchool createMagic(Identifier id, int color, boolean customDamageType, EntityAttribute attr, StatusEffect effect) {
        var school = new SpellSchool(SpellSchool.Archetype.MAGIC, id, color,
                customDamageType ? RegistryKey.of(RegistryKeys.DAMAGE_TYPE, id) : DamageTypes.MAGIC, attr, effect);
        return configureAsMagic(school, attr);
    }
    public static SpellSchool configureAsMagic(SpellSchool school, EntityAttribute powerAttribute) {
        school.addSource(SpellSchool.Trait.POWER, new SpellSchool.Source(SpellSchool.Apply.ADD, query -> query.entity().getAttributeValue(powerAttribute)));
        configureSpellHaste(school); configureSpellCritChance(school); configureSpellCritDamage(school); return school;
    }
    public static SpellSchool configureSpellHaste(SpellSchool school) {
        school.addSource(SpellSchool.Trait.HASTE, new SpellSchool.Source(SpellSchool.Apply.ADD, query -> query.entity().getAttributeValue(SpellPowerMechanics.HASTE.attribute) / PERCENT_ATTRIBUTE_BASELINE - 1));
        school.addSource(SpellSchool.Trait.HASTE, new SpellSchool.Source(SpellSchool.Apply.ADD, query -> {
            var enchantment = Enchantments_SpellPowerMechanics.HASTE;
            return enchantment.amplified(0, SpellPowerMod.mainHandEnchantmentLevel(enchantment, query.entity()));
        })); return school;
    }
    public static SpellSchool configureSpellCritChance(SpellSchool school) {
        school.addSource(SpellSchool.Trait.CRIT_CHANCE, new SpellSchool.Source(SpellSchool.Apply.ADD, query -> query.entity().getAttributeValue(SpellPowerMechanics.CRITICAL_CHANCE.attribute) / PERCENT_ATTRIBUTE_BASELINE - 1));
        school.addSource(SpellSchool.Trait.CRIT_CHANCE, new SpellSchool.Source(SpellSchool.Apply.ADD, query -> {
            var enchantment = Enchantments_SpellPowerMechanics.CRITICAL_CHANCE;
            return enchantment.amplified(0, SpellPowerMod.mainHandEnchantmentLevel(enchantment, query.entity()));
        })); return school;
    }
    public static SpellSchool configureSpellCritDamage(SpellSchool school) {
        school.addSource(SpellSchool.Trait.CRIT_DAMAGE, new SpellSchool.Source(SpellSchool.Apply.ADD, query -> query.entity().getAttributeValue(SpellPowerMechanics.CRITICAL_DAMAGE.attribute) / PERCENT_ATTRIBUTE_BASELINE - 1));
        school.addSource(SpellSchool.Trait.CRIT_DAMAGE, new SpellSchool.Source(SpellSchool.Apply.ADD, query -> {
            var enchantment = Enchantments_SpellPowerMechanics.CRITICAL_DAMAGE;
            return enchantment.amplified(0, SpellPowerMod.mainHandEnchantmentLevel(enchantment, query.entity()));
        })); return school;
    }
    @Nullable public static SpellSchool getSchool(String value) {
        var id = new Identifier(value.toLowerCase(Locale.US));
        if (id.getNamespace().equals(Identifier.DEFAULT_NAMESPACE)) id = new Identifier(DEFAULT_NAMESPACE, id.getPath());
        return REGISTRY.get(id);
    }
}
