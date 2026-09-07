package net.spell_power.api;

import net.minecraft.entity.LivingEntity;
import net.minecraft.entity.attribute.ClampedEntityAttribute;
import net.minecraft.entity.attribute.EntityAttribute;
import net.minecraft.entity.damage.DamageSource;
import net.minecraft.entity.damage.DamageType;
import net.minecraft.registry.RegistryKeys;
import net.minecraft.registry.tag.TagKey;
import net.minecraft.util.Identifier;
import net.spell_power.SpellPowerMod;
import java.util.ArrayList;

public class SpellResistance {
    public static class Attributes {
        public static final ArrayList<Entry> all=new ArrayList<>();
        public static Entry entry(String name,String tagName,double max,boolean tracked){ return entry("resistance."+name,new Identifier(SpellPowerMod.ID,tagName),max,tracked); }
        public static Entry entry(String name,Identifier tagId,double max,boolean tracked){ var e=new Entry(name,TagKey.of(RegistryKeys.DAMAGE_TYPE,tagId),max,tracked); all.add(e); return e; }
        public static class Entry {
            public final Identifier id; public final String translationKey; public final EntityAttribute attribute; public final double baseValue; public final TagKey<DamageType> damageTypes; public final double maxValue;
            public Entry(String name,TagKey<DamageType> tag,double max,boolean tracked){ id=new Identifier(SpellPowerMod.ID,name); translationKey="attribute.name." + SpellPowerMod.ID + "." + name; baseValue=0; maxValue=max; damageTypes=tag; attribute=new ClampedEntityAttribute(translationKey,0,0,max).setTracked(tracked); }
        }
        public static final Entry GENERIC=entry("generic","resistable",1024,true);
    }
    public static double resist(LivingEntity target,double damage,DamageSource source){
        double modifier=1; var config=SpellPowerMod.attributesConfig.value;
        for(var type:Attributes.all){ if(target.getAttributes().hasAttribute(type.attribute) && source.isIn(type.damageTypes)){
            float r=(float)target.getAttributeValue(type.attribute); float reduction=0;
            switch(config.resistance_curve){
                case LINEAR -> reduction=r/config.resistance_tuning_constant;
                case QUADRATIC -> reduction=(float)Math.sqrt(r*config.resistance_tuning_constant)/config.resistance_tuning_constant;
                case HYPERBOLIC -> reduction=r/(r+config.resistance_tuning_constant);
            }
            reduction=Math.min(reduction*config.resistance_multiplier,config.resistance_reduction_cap); modifier*=1-reduction;
        }}
        return damage*modifier;
    }
}
