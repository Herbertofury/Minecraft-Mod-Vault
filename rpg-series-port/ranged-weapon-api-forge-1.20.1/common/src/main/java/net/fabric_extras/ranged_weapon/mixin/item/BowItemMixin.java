package net.fabric_extras.ranged_weapon.mixin.item;

import com.llamalad7.mixinextras.injector.wrapoperation.Operation;
import com.llamalad7.mixinextras.injector.wrapoperation.WrapOperation;
import net.fabric_extras.ranged_weapon.api.CustomRangedWeapon;
import net.fabric_extras.ranged_weapon.api.EntityAttributes_RangedWeapon;
import net.fabric_extras.ranged_weapon.internal.ScalingUtil;
import net.minecraft.entity.Entity;
import net.minecraft.entity.LivingEntity;
import net.minecraft.entity.projectile.PersistentProjectileEntity;
import net.minecraft.item.BowItem;
import net.minecraft.item.Item;
import net.minecraft.item.ItemStack;
import net.minecraft.world.World;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.injection.At;

@Mixin(BowItem.class)
public class BowItemMixin {
    public float getPullProgress_RWA(int useTicks, LivingEntity user) {
        float ticks=Math.max(2F,(float)user.getAttributeValue(EntityAttributes_RangedWeapon.PULL_TIME.attribute)*20F);
        float f=useTicks/ticks; f=(f*f+f*2F)/3F; return Math.min(f,1F);
    }
    @WrapOperation(method="onStoppedUsing",at=@At(value="INVOKE",target="Lnet/minecraft/item/BowItem;getPullProgress(I)F"))
    private float rwa$pull(int ticks, Operation<Float> original, ItemStack stack, World world, LivingEntity user, int remaining) { return getPullProgress_RWA(ticks,user); }
    @WrapOperation(method="onStoppedUsing",at=@At(value="INVOKE",target="Lnet/minecraft/world/World;spawnEntity(Lnet/minecraft/entity/Entity;)Z"))
    private boolean rwa$projectile(World instance, Entity entity, Operation<Boolean> original, ItemStack stack, World world, LivingEntity user, int remaining) {
        if (entity instanceof PersistentProjectileEntity projectile) {
            var item=(Item)(Object)this;
            double velocityMultiplier=ScalingUtil.arrowVelocityMultiplier(item,user.getAttributeValue(EntityAttributes_RangedWeapon.VELOCITY.attribute));
            if (velocityMultiplier!=1) projectile.setVelocity(projectile.getVelocity().multiply(velocityMultiplier));
            double damage=user.getAttributeValue(EntityAttributes_RangedWeapon.DAMAGE.attribute);
            if (damage>0 && (Object)this instanceof CustomRangedWeapon weapon) projectile.setDamage(projectile.getDamage()*ScalingUtil.arrowDamageMultiplier(weapon.getTypeBaseline().damage(),damage,velocityMultiplier));
        }
        return original.call(instance,entity);
    }
}
