package net.fabric_extras.ranged_weapon.mixin.client;

import com.llamalad7.mixinextras.injector.wrapoperation.Operation;
import com.llamalad7.mixinextras.injector.wrapoperation.WrapOperation;
import net.fabric_extras.ranged_weapon.api.CustomBow;
import net.fabric_extras.ranged_weapon.api.CustomRangedWeapon;
import net.minecraft.client.network.AbstractClientPlayerEntity;
import net.minecraft.item.Item;
import net.minecraft.item.ItemStack;
import net.minecraft.item.Items;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.injection.At;
import org.spongepowered.asm.mixin.injection.Constant;
import org.spongepowered.asm.mixin.injection.ModifyConstant;

@Mixin(AbstractClientPlayerEntity.class)
public class AbstractClientPlayerEntityMixin {
    @WrapOperation(method="getFovMultiplier",at=@At(value="INVOKE",target="Lnet/minecraft/item/ItemStack;isOf(Lnet/minecraft/item/Item;)Z"))
    private boolean rwa$customBow(ItemStack stack, Item item, Operation<Boolean> original) {
        if (item==Items.BOW && CustomBow.instances.contains(stack.getItem())) return true;
        return original.call(stack,item);
    }
    @ModifyConstant(method="getFovMultiplier",constant=@Constant(floatValue=20F))
    private float rwa$pullTicks(float original) {
        var item=((AbstractClientPlayerEntity)(Object)this).getActiveItem().getItem();
        if (item instanceof CustomBow && item instanceof CustomRangedWeapon weapon) return weapon.getRangedWeaponConfig().pullTimeSeconds()*20F;
        return original;
    }
}
