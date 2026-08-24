package net.fabric_extras.ranged_weapon.mixin.item;

import net.fabric_extras.ranged_weapon.api.CustomRangedWeapon;
import net.fabric_extras.ranged_weapon.api.EntityAttributes_RangedWeapon;
import net.fabric_extras.ranged_weapon.internal.ArrowExtension;
import net.minecraft.entity.LivingEntity;
import net.minecraft.entity.projectile.PersistentProjectileEntity;
import net.minecraft.entity.projectile.ProjectileUtil;
import net.minecraft.item.ItemStack;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.injection.At;
import org.spongepowered.asm.mixin.injection.Inject;
import org.spongepowered.asm.mixin.injection.callback.CallbackInfoReturnable;

/**
 * Backports the 2.3.3 bow-using-mob damage fix to Minecraft 1.20.1.
 * Skeleton-family and Illusioner shots are created through ProjectileUtil rather than BowItem#onStoppedUsing.
 */
@Mixin(ProjectileUtil.class)
public final class ProjectileUtilMixin {
    @Inject(method = "createArrowProjectile", at = @At("RETURN"))
    private static void rwa$scaleMobBowDamage(LivingEntity shooter, ItemStack ammunition, float damageModifier,
                                               CallbackInfoReturnable<PersistentProjectileEntity> cir) {
        var projectile = cir.getReturnValue();
        if (projectile == null || ((ArrowExtension)(Object)projectile).rwa_isModified()) return;

        ItemStack weaponStack = shooter.getMainHandStack();
        if (!(weaponStack.getItem() instanceof CustomRangedWeapon)) weaponStack = shooter.getOffHandStack();
        if (!(weaponStack.getItem() instanceof CustomRangedWeapon weapon)) return;

        double baselineDamage = weapon.getTypeBaseline().damage();
        double rangedDamage = shooter.getAttributeValue(EntityAttributes_RangedWeapon.DAMAGE.attribute);
        if (baselineDamage <= 0 || rangedDamage <= 0) return;

        projectile.setDamage(projectile.getDamage() * (rangedDamage / baselineDamage));
        ((ArrowExtension)(Object)projectile).rwa_markModified(true);
    }
}
