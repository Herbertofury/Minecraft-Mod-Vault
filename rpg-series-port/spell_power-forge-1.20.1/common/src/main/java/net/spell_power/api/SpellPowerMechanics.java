package net.spell_power.api;

import net.minecraft.entity.attribute.EntityAttributeModifier;
import net.minecraft.entity.effect.StatusEffect;
import net.minecraft.entity.effect.StatusEffectCategory;
import net.minecraft.util.Identifier;
import net.spell_power.SpellPowerMod;
import net.spell_power.internals.CustomEntityAttribute;
import net.spell_power.internals.SpellStatusEffect;

import java.util.HashMap;
import java.util.UUID;

public class SpellPowerMechanics {
    public static final float PERCENT_ATTRIBUTE_BASELINE = 100F;
    public static String translationPrefix() { return "attribute.name." + SpellPowerMod.ID + "."; }
    public static class Entry {
        public final String name;
        public final Identifier id;
        public final float defaultValue, min, max;
        public final CustomEntityAttribute attribute;
        public final StatusEffect boostEffect;
        public EntityAttributeModifier innateModifier;
        public Entry(String name, float defaultValue, float min, float max, int color) {
            this.name=name; this.id=new Identifier(SpellPowerMod.ID,name); this.defaultValue=defaultValue; this.min=min; this.max=max;
            this.attribute=(CustomEntityAttribute)new CustomEntityAttribute(translationPrefix()+name,defaultValue,min,max,id).setTracked(true);
            this.boostEffect=new SpellStatusEffect(StatusEffectCategory.BENEFICIAL,color);
        }
        public Entry innate(double value) {
            UUID id = UUID.nameUUIDFromBytes((SpellPowerMod.ID+":innate_bonus/"+name).getBytes(java.nio.charset.StandardCharsets.UTF_8));
            innateModifier = new EntityAttributeModifier(id, SpellPowerMod.ID+".innate_bonus."+name, value, EntityAttributeModifier.Operation.MULTIPLY_BASE);
            return this;
        }
    }
    public static final HashMap<String,Entry> all=new HashMap<>();
    public static Entry entry(String name,float def,float min,float max,int color){ var e=new Entry(name,def,min,max,color); all.put(name,e); return e; }
    public static final Entry CRITICAL_CHANCE=entry("critical_chance",100,100,1000,0x66ccff).innate(SpellPowerMod.attributesConfig.value.base_spell_critical_chance_percentage/100.0);
    public static final Entry CRITICAL_DAMAGE=entry("critical_damage",100,100,1000,0x66ffcc).innate(SpellPowerMod.attributesConfig.value.base_spell_critical_damage_percentage/100.0);
    public static final Entry HASTE=entry("haste",100,10,1000,0xcc99ff);
}
