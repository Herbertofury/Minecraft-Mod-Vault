package com.bert.slrsharedmana.mixin;

import com.bert.slrsharedmana.BridgeConfig;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.injection.At;
import org.spongepowered.asm.mixin.injection.Inject;
import org.spongepowered.asm.mixin.injection.callback.CallbackInfoReturnable;

@Mixin(targets = "io.redspace.ironsspellbooks.capabilities.magic.MagicManager", remap = false)
public abstract class MagicManagerMixin {
    @Inject(method = "regenPlayerMana", at = @At("HEAD"), cancellable = true, remap = false)
    private void slr$disablePassiveIronRegen(CallbackInfoReturnable<Boolean> cir) {
        if (BridgeConfig.ENABLED.get() && BridgeConfig.SUPPRESS_ISS_PASSIVE_REGEN.get()) {
            cir.setReturnValue(false);
        }
    }
}
