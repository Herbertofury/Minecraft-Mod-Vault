package com.bert.slrsharedmana.mixin.client;

import com.bert.slrsharedmana.BridgeConfig;
import com.bert.slrsharedmana.client.ClientManaView;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.injection.At;
import org.spongepowered.asm.mixin.injection.Inject;
import org.spongepowered.asm.mixin.injection.callback.CallbackInfoReturnable;

@Mixin(targets = "io.redspace.ironsspellbooks.api.magic.MagicData", remap = false)
public abstract class MagicDataClientMixin {
    @Inject(method = "getMana()F", at = @At("HEAD"), cancellable = true, remap = false)
    private void slr$clientMana(CallbackInfoReturnable<Float> cir) {
        if (!BridgeConfig.ENABLED.get()) return;
        Float value = ClientManaView.getFor(this);
        if (value != null) cir.setReturnValue(value);
    }
}
