package com.bert.slrsharedmana.mixin;

import com.bert.slrsharedmana.BridgeConfig;
import com.bert.slrsharedmana.IssManaEventBridge;
import com.bert.slrsharedmana.SharedManaMod;
import com.bert.slrsharedmana.SlrAccess;
import net.minecraft.server.level.ServerPlayer;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.Shadow;
import org.spongepowered.asm.mixin.Unique;
import org.spongepowered.asm.mixin.injection.At;
import org.spongepowered.asm.mixin.injection.Inject;
import org.spongepowered.asm.mixin.injection.callback.CallbackInfo;
import org.spongepowered.asm.mixin.injection.callback.CallbackInfoReturnable;

@Mixin(targets = "io.redspace.ironsspellbooks.api.magic.MagicData", remap = false)
public abstract class MagicDataMixin {
    @Shadow(remap = false) private float mana;
    @Shadow(remap = false) private ServerPlayer serverPlayer;
    @Unique private static final float SLR$EPS = 0.0005F;

    @Inject(method = "getMana()F", at = @At("HEAD"), cancellable = true, remap = false)
    private void slr$authoritativeMana(CallbackInfoReturnable<Float> cir) {
        if (!BridgeConfig.classicSharedMode() || serverPlayer == null) return;
        float value = (float) SlrAccess.ironEquivalent(serverPlayer);
        this.mana = value;
        cir.setReturnValue(value);
    }

    @Inject(method = "setMana(F)V", at = @At("HEAD"), cancellable = true, remap = false)
    private void slr$routeManaChange(float requestedMana, CallbackInfo ci) {
        if (!BridgeConfig.classicSharedMode() || serverPlayer == null) return;
        try {
            float before = (float) SlrAccess.ironEquivalent(serverPlayer);
            float accepted = IssManaEventBridge.resolve(serverPlayer, this, before, requestedMana);
            if (!Float.isFinite(accepted)) accepted = before;

            float delta = accepted - before;
            if (serverPlayer.isCreative() && delta < 0.0F) delta = 0.0F;
            if (Math.abs(delta) > SLR$EPS) {
                SlrAccess.applyIronDelta(serverPlayer, delta);
                if (delta < 0.0F) SlrAccess.triggerManaRefresh(serverPlayer);
            }
            this.mana = (float) SlrAccess.ironEquivalent(serverPlayer);
            ci.cancel();
        } catch (RuntimeException e) {
            SharedManaMod.LOGGER.error("Shared-mana routing failed; allowing Iron's native setMana as a fail-safe", e);
        }
    }
}
