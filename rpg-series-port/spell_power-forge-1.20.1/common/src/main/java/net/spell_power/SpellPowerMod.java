package net.spell_power;

import net.minecraft.entity.attribute.EntityAttributeModifier;
import net.spell_power.api.*;
import net.spell_power.api.enchantment.Enchantments_SpellPower;
import net.spell_power.api.enchantment.Enchantments_SpellPowerMechanics;
import net.spell_power.api.enchantment.SpellPowerEnchanting;
import net.spell_power.config.AttributesConfig;
import net.spell_power.config.EnchantmentsConfig;
import net.tinyconfig.ConfigManager;
import java.util.Map;

public final class SpellPowerMod {
    public static final String ID="spell_power";
    public static final ConfigManager<AttributesConfig> attributesConfig=new ConfigManager<AttributesConfig>("attributes",AttributesConfig.defaults()).builder().setDirectory(ID).sanitize(true).validate(AttributesConfig::isValid).build();
    public static final ConfigManager<EnchantmentsConfig> enchantmentConfig=new ConfigManager<EnchantmentsConfig>("enchantments",new EnchantmentsConfig()).builder().setDirectory(ID).sanitize(true).schemaVersion(4).build();
    private SpellPowerMod(){}

    public static void refreshConfigs(){ attributesConfig.refresh(); enchantmentConfig.refresh(); }
    public static void prepareStatusEffects(){
        var powerCfg=attributesConfig.value.spell_power_effect;
        for(var school:SpellSchools.all()) if(school.powerEffectManagement.isInternal() && school.boostEffect!=null){ school.boostEffect.addAttributeModifier(school.attribute,powerCfg.uuid,powerCfg.bonus_per_stack,EntityAttributeModifier.Operation.MULTIPLY_TOTAL); }
        for(var entry:SpellPowerMechanics.all.values()){
            var cfg=attributesConfig.value.secondary_effects.get(entry.name);
            if(cfg!=null) entry.boostEffect.addAttributeModifier(entry.attribute,cfg.uuid,cfg.bonus_per_stack,EntityAttributeModifier.Operation.MULTIPLY_TOTAL);
        }
    }
    public static void applyEnchantments(){ enchantmentConfig.value.apply(); attachEnchantmentsToSchools(); }
    private static void attachEnchantmentsToSchools(){
        for(var school:SpellSchools.all()){
            var powering=Enchantments_SpellPower.all.entrySet().stream().filter(e->e.getValue().poweredSchools().contains(school)).map(Map.Entry::getValue).toList();
            school.addSource(SpellSchool.Trait.POWER,new SpellSchool.Source(SpellSchool.Apply.MULTIPLY,q->{ double value=0; for(var ench:powering){ int level=SpellPowerEnchanting.getEnchantmentLevel(ench,q.entity(),null); value=ench.amplified(value,level); } return value; }));
        }
    }
    @Deprecated public static AttributesConfig.AttributeScope attributeScopeOverride=null;
    @Deprecated public static AttributesConfig.AttributeScope attributeScope(){ return AttributesConfig.AttributeScope.LIVING_ENTITY; }
    public static void touchRegistries(){ SpellPowerMechanics.all.size(); SpellSchools.all().size(); SpellResistance.Attributes.all.size(); Enchantments_SpellPowerMechanics.all.size(); Enchantments_SpellPower.all.size(); }
}
