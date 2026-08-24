package net.fabric_extras.ranged_weapon.mixin.item;

import com.llamalad7.mixinextras.injector.wrapoperation.Operation;
import com.llamalad7.mixinextras.injector.wrapoperation.WrapOperation;
import net.fabric_extras.ranged_weapon.api.CrossbowMechanics;
import net.fabric_extras.ranged_weapon.api.CustomRangedWeapon;
import net.fabric_extras.ranged_weapon.api.EntityAttributes_RangedWeapon;
import net.fabric_extras.ranged_weapon.internal.ScalingUtil;
import net.minecraft.entity.Entity;
import net.minecraft.entity.LivingEntity;
import net.minecraft.entity.projectile.PersistentProjectileEntity;
import net.minecraft.item.CrossbowItem;
import net.minecraft.item.ItemStack;
import net.minecraft.util.Hand;
import net.minecraft.world.World;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.injection.At;
import org.spongepowered.asm.mixin.injection.Inject;
import org.spongepowered.asm.mixin.injection.ModifyVariable;
import org.spongepowered.asm.mixin.injection.callback.CallbackInfoReturnable;

@Mixin(CrossbowItem.class)
public class CrossbowItemMixin {
    @Inject(method="getPullTime",at=@At("HEAD"),cancellable=true)
    private static void rwa$pull(ItemStack stack, CallbackInfoReturnable<Integer> cir) {
        if (stack.getItem() instanceof CustomRangedWeapon weapon) {
            int ticks=Math.max(2,Math.round(weapon.getRangedWeaponConfig().pullTimeSeconds()*20F));
            cir.setReturnValue(CrossbowMechanics.PullTime.modifier.getPullTime(ticks,stack,null));
        }
    }
    @ModifyVariable(method="shoot",at=@At("HEAD"),ordinal=1,argsOnly=true)
    private static float rwa$velocity(float speed, World world, LivingEntity shooter, Hand hand, ItemStack crossbow, ItemStack projectile, float soundPitch, boolean creative, float speed1, float divergence, float simulated) {
        return speed*(float)ScalingUtil.arrowVelocityMultiplier(crossbow.getItem(),shooter.getAttributeValue(EntityAttributes_RangedWeapon.VELOCITY.attribute));
    }
    @WrapOperation(method="shoot",at=@At(value="INVOKE",target="Lnet/minecraft/world/World;spawnEntity(Lnet/minecraft/entity/Entity;)Z"))
    private static boolean rwa$damage(World instance, Entity entity, Operation<Boolean> original, World world, LivingEntity shooter, Hand hand, ItemStack crossbow, ItemStack projectileStack, float soundPitch, boolean creative, float speed, float divergence, float simulated) {
        if (entity instanceof PersistentProjectileEntity projectile && crossbow.getItem() instanceof CustomRangedWeapon weapon) {
            double velocityMultiplier=ScalingUtil.arrowVelocityMultiplier(crossbow.getItem(),shooter.getAttributeValue(EntityAttributes_RangedWeapon.VELOCITY.attribute));
            double damage=shooter.getAttributeValue(EntityAttributes_RangedWeapon.DAMAGE.attribute);
            if (damage>0) projectile.setDamage(projectile.getDamage()*ScalingUtil.arrowDamageMultiplier(weapon.getTypeBaseline().damage(),damage,velocityMultiplier));
        }
        return original.call(instance,entity);
    }
}
