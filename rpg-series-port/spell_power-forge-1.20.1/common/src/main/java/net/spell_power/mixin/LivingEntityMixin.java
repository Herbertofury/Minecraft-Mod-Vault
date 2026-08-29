package net.spell_power.mixin;

import net.minecraft.entity.Entity;
import net.minecraft.entity.EntityType;
import net.minecraft.entity.LivingEntity;
import net.minecraft.entity.damage.DamageSource;
import net.minecraft.world.World;
import net.spell_power.api.SpellPowerMechanics;
import net.spell_power.api.SpellResistance;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.injection.At;
import org.spongepowered.asm.mixin.injection.Inject;
import org.spongepowered.asm.mixin.injection.ModifyVariable;
import org.spongepowered.asm.mixin.injection.callback.CallbackInfo;

@Mixin(LivingEntity.class)
abstract class LivingEntityMixin extends Entity {
    LivingEntityMixin(EntityType<?> type,World world){ super(type,world); }
    @Inject(method="<init>",at=@At("TAIL"))
    private void spellPower$innate(EntityType<?> type,World world,CallbackInfo ci){
        LivingEntity self=(LivingEntity)(Object)this;
        for(var mechanic:SpellPowerMechanics.all.values()) if(mechanic.innateModifier!=null){
            var instance=self.getAttributeInstance(mechanic.attribute);
            if(instance!=null && instance.getModifier(mechanic.innateModifier.getId())==null) instance.addPersistentModifier(mechanic.innateModifier);
        }
    }
    @ModifyVariable(method="damage",at=@At("HEAD"),ordinal=0)
    private float spellPower$resist(float amount,DamageSource source){ LivingEntity self=(LivingEntity)(Object)this; if(self.isInvulnerableTo(source)||self.isDead()) return amount; return (float)SpellResistance.resist(self,amount,source); }
}
