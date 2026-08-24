package net.fabric_extras.ranged_weapon.mixin.attribute;

import net.fabric_extras.ranged_weapon.api.EntityAttributes_RangedWeapon;
import net.minecraft.entity.Entity;
import net.minecraft.entity.EntityType;
import net.minecraft.entity.LivingEntity;
import net.minecraft.item.ItemStack;
import net.minecraft.util.UseAction;
import net.minecraft.world.World;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.Shadow;
import org.spongepowered.asm.mixin.injection.At;
import org.spongepowered.asm.mixin.injection.Inject;
import org.spongepowered.asm.mixin.injection.callback.CallbackInfoReturnable;

@Mixin(LivingEntity.class)
public abstract class LivingEntityMixin extends Entity {
    LivingEntityMixin(EntityType<?> type, World world) { super(type,world); }
    @Shadow protected int itemUseTimeLeft;
    @Shadow protected ItemStack activeItemStack;
    @Inject(method="getItemUseTimeLeft", at=@At("HEAD"), cancellable=true)
    private void rwa$haste(CallbackInfoReturnable<Integer> cir) {
        var entity=(LivingEntity)(Object)this;
        if (!entity.isUsingItem()) return;
        var action=activeItemStack.getUseAction();
        if (action!=UseAction.BOW && action!=UseAction.CROSSBOW) return;
        var progress=activeItemStack.getMaxUseTime()-itemUseTimeLeft;
        var haste=entity.getAttributeValue(EntityAttributes_RangedWeapon.HASTE.attribute);
        var adjusted=(int)(progress*EntityAttributes_RangedWeapon.HASTE.asMultiplier(haste));
        cir.setReturnValue(activeItemStack.getMaxUseTime()-adjusted);
    }
}
