#!/usr/bin/env python3
from __future__ import annotations

import pathlib
import sys

java_root = pathlib.Path(sys.argv[1]).resolve()
if not java_root.is_dir():
    raise SystemExit(f"missing generated Java root: {java_root}")

path = java_root / "net/archers/mixin/screen/GrindstoneScreenHandlerMixin.java"
if not path.is_file():
    raise SystemExit(f"Grindstone runtime transform: missing source: {path}")

original = path.read_text(encoding="utf-8")
expected = '''package net.archers.mixin.screen;

import net.archers.item.misc.AutoFireHook;
import net.minecraft.item.ItemStack;
import net.minecraft.screen.GrindstoneScreenHandler;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.injection.At;
import org.spongepowered.asm.mixin.injection.Inject;
import org.spongepowered.asm.mixin.injection.callback.CallbackInfoReturnable;

@Mixin(GrindstoneScreenHandler.class)
public class GrindstoneScreenHandlerMixin {
    @Inject(method = "getOutputStack", at = @At("HEAD"), cancellable = true)
    private void getOutputStack_Archers(ItemStack firstInput, ItemStack secondInput, CallbackInfoReturnable<ItemStack> cir) {
        if (firstInput.isEmpty() && secondInput.isEmpty() ) {
            return;
        }
        var output = cir.getReturnValue();
        if (firstInput.isEmpty() || secondInput.isEmpty() ) {
            if (output == null || output.isEmpty()) {
                output = firstInput.isEmpty() ? secondInput : firstInput;
            }
            if (output != null && AutoFireHook.isApplied(output)) {
                var newStack = output.copy();
                AutoFireHook.remove(newStack);
                cir.setReturnValue(newStack);
                cir.cancel();
            }
        }
    }
}
'''
if original != expected:
    raise SystemExit("Grindstone runtime transform: exact current-upstream mixin body changed")

# 1.20.1 has no getOutputStack(first, second). Its private updateResult() owns the same two input
# stacks and result inventory. Mirror the current 3.1.1 semantics at updateResult TAIL: only remove
# Auto Fire when exactly one input is present; leave the two-input repair path untouched.
path.write_text('''package net.archers.mixin.screen;

import net.archers.item.misc.AutoFireHook;
import net.minecraft.inventory.Inventory;
import net.minecraft.screen.GrindstoneScreenHandler;
import org.spongepowered.asm.mixin.Final;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.Shadow;
import org.spongepowered.asm.mixin.injection.At;
import org.spongepowered.asm.mixin.injection.Inject;
import org.spongepowered.asm.mixin.injection.callback.CallbackInfo;

@Mixin(GrindstoneScreenHandler.class)
public class GrindstoneScreenHandlerMixin {
    @Shadow @Final Inventory input;
    @Shadow @Final private Inventory result;

    @Inject(method = "updateResult", at = @At("TAIL"))
    private void updateResult_Archers(CallbackInfo ci) {
        var firstInput = input.getStack(0);
        var secondInput = input.getStack(1);
        if (firstInput.isEmpty() == secondInput.isEmpty()) {
            return;
        }

        var output = result.getStack(0);
        if (!output.isEmpty() && AutoFireHook.isApplied(output)) {
            var newStack = output.copy();
            AutoFireHook.remove(newStack);
            result.setStack(0, newStack);
        }
    }
}
''', encoding="utf-8")

print("[Archers runtime transforms] grindstone Auto Fire removal: current one-input semantics -> 1.20.1 updateResult")
