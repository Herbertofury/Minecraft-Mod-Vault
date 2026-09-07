package net.spell_power.config;

import net.spell_power.api.DamageCurve;
import net.spell_power.internals.SpellStatusEffect;
import java.util.Map;

public class AttributesConfig {
    public enum AttributeScope { LIVING_ENTITY, PLAYER_ENTITY }
    @Deprecated public AttributeScope attributes_container_injection_scope = AttributeScope.LIVING_ENTITY;
    public boolean migrate_attributes_base = true;
    public boolean use_vanilla_magic_damage_type = false;
    public float base_spell_power = 1F;
    public double base_spell_critical_chance_percentage = 5;
    public double base_spell_critical_damage_percentage = 50;
    @Deprecated public int status_effect_raw_id_starts_at = 730;
    public SpellStatusEffect.Config spell_power_effect = new SpellStatusEffect.Config("446cf95e-be63-40d9-ad90-6cc388c08460",0.1F);
    public Map<String,SpellStatusEffect.Config> secondary_effects;
    public boolean enchantments_require_matching_attribute = true;
    public DamageCurve resistance_curve = DamageCurve.HYPERBOLIC;
    public float resistance_multiplier = 1F;
    public float resistance_tuning_constant = 20F;
    public float resistance_reduction_cap = 0.9F;
    public boolean register_potions = false;

    public static AttributesConfig defaults(){
        var c=new AttributesConfig();
        c.secondary_effects=Map.of(
            "critical_chance",new SpellStatusEffect.Config("0e0ddd12-0646-42b7-8daf-36b4ccf524df",0.05F),
            "critical_damage",new SpellStatusEffect.Config("0612ed2a-3ce5-11ed-b878-0242ac120002",0.1F),
            "haste",new SpellStatusEffect.Config("092f4f58-3ce5-11ed-b878-0242ac120002",0.05F));
        return c;
    }
    public boolean isValid(){
        if(attributes_container_injection_scope==null || secondary_effects==null || spell_power_effect==null || resistance_curve==null) return false;
        for(var k:defaults().secondary_effects.keySet()) if(!secondary_effects.containsKey(k)) return false;
        return resistance_tuning_constant>0 && resistance_multiplier>=0 && resistance_reduction_cap>=0 && resistance_reduction_cap<=1;
    }
}
