package net.fabric_extras.ranged_weapon.mixin;

import net.fabric_extras.ranged_weapon.internal.ArrowExtension;
import net.minecraft.entity.projectile.PersistentProjectileEntity;
import net.minecraft.util.math.MathHelper;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.Shadow;
import org.spongepowered.asm.mixin.Unique;
import org.spongepowered.asm.mixin.injection.At;
import org.spongepowered.asm.mixin.injection.ModifyVariable;

import java.util.Random;

/** 2.0.6 critical-hit scaling semantics on the proven 1.20.1 projectile target. */
@Mixin(PersistentProjectileEntity.class)
public abstract class PersistentProjectileEntityMixin implements ArrowExtension {
    @Unique private static final Random RWA_CRIT_RANDOM = new Random();
    @Unique private boolean rwa_modified;

    @Shadow private double damage;
    @Shadow public abstract boolean isCritical();

    @ModifyVariable(method = "onEntityHit", at = @At("STORE"), ordinal = 0)
    private int rwa$scaleCriticalDamage(int value) {
        if (!isCritical()) return value;
        var projectile = (PersistentProjectileEntity)(Object)this;
        var velocity = projectile.getVelocity().length();
        var criticalMultiplier = 1F + (0.1F + RWA_CRIT_RANDOM.nextFloat() * 0.5F);
        return (int)Math.round(MathHelper.clamp(velocity * this.damage * criticalMultiplier, 0.0, 2.147483647E9));
    }

    @Override public void rwa_markModified(boolean modified) { this.rwa_modified = modified; }
    @Override public boolean rwa_isModified() { return this.rwa_modified; }
}
