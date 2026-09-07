#!/usr/bin/env python3
from __future__ import annotations
import pathlib, sys

root = pathlib.Path(sys.argv[1]).resolve()
path = root / "net/rogues/mixin/LivingEntityStealth.java"
text = path.read_text(encoding="utf-8")

old_imports = "import com.llamalad7.mixinextras.injector.wrapoperation.Operation;\nimport com.llamalad7.mixinextras.injector.wrapoperation.WrapOperation;\n"
if text.count(old_imports) != 1:
    raise SystemExit("[Rogues 1.20.1 compat batch9] expected exactly one MixinExtras visibility-wrapper import block")
text = text.replace(old_imports, "import org.spongepowered.asm.mixin.injection.Redirect;\n", 1)

old_handler = '''    @WrapOperation(method = "updatePotionVisibility", at = @At(value = "INVOKE", target = "Lnet/minecraft/entity/LivingEntity;hasStatusEffect(Lnet/minecraft/entity/effect/StatusEffect;)Z"))
    private boolean updatePotionVisibility_WRAP_Stealth(LivingEntity instance, StatusEffect effect, Operation<Boolean> original) {
        return original.call(instance, effect) || instance.hasStatusEffect(RogueEffects.STEALTH.effect);
    }
'''
new_handler = '''    @Redirect(method = "updatePotionVisibility", at = @At(value = "INVOKE", target = "Lnet/minecraft/entity/LivingEntity;hasStatusEffect(Lnet/minecraft/entity/effect/StatusEffect;)Z"))
    private boolean updatePotionVisibility_REDIRECT_Stealth(LivingEntity instance, StatusEffect effect) {
        return instance.hasStatusEffect(effect) || instance.hasStatusEffect(RogueEffects.STEALTH.effect);
    }
'''
if text.count(old_handler) != 1:
    raise SystemExit("[Rogues 1.20.1 compat batch9] expected exactly one updatePotionVisibility MixinExtras wrapper")
text = text.replace(old_handler, new_handler, 1)

for forbidden in ("WrapOperation", "Operation<Boolean>", "updatePotionVisibility_WRAP_Stealth"):
    if forbidden in text:
        raise SystemExit(f"[Rogues 1.20.1 compat batch9] stale visibility wrapper token survived: {forbidden}")
if text.count('@Redirect(method = "updatePotionVisibility"') != 1:
    raise SystemExit("[Rogues 1.20.1 compat batch9] standard redirect was not installed exactly once")

path.write_text(text, encoding="utf-8")
print("[Rogues 1.20.1 compat batch9] replaced unmapped MixinExtras visibility wrapper with standard remappable @Redirect")
