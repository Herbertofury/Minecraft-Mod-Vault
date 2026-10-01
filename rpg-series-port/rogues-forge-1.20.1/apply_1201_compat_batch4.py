#!/usr/bin/env python3
from __future__ import annotations
import pathlib, sys

root = pathlib.Path(sys.argv[1]).resolve()
path = root / "net/rogues/mixin/TrackTargetGoalStealth.java"
text = path.read_text(encoding="utf-8")
old = '''//    @Inject(method = "shouldContinue", at = @At("HEAD"), cancellable = true)
//    private void shouldContinue_HEAD(CallbackInfoReturnable<Boolean> cir) {
//        var target = mob.getTarget();
//        if (target != null && target.hasStatusEffect(Effects.STEALTH)) {
//            cir.setReturnValue(false);
//        }
//    }
'''
new = '''    @Inject(method = "shouldContinue", at = @At("HEAD"), cancellable = true)
    private void shouldContinue_HEAD(CallbackInfoReturnable<Boolean> cir) {
        var target = mob.getTarget();
        if (target != null
                && (target.hasStatusEffect(RogueEffects.STEALTH.effect) || target.hasStatusEffect(RogueEffects.SHADOW_STEP.effect))) {
            cir.setReturnValue(false);
        }
    }
'''
count = text.count(old)
if count != 1:
    raise SystemExit(f"[Rogues 1.20.1 compat batch4] expected one upstream shouldContinue fallback seam, found {count}")
text = text.replace(old, new, 1)
path.write_text(text, encoding="utf-8")
if '//    @Inject(method = "shouldContinue"' in text:
    raise SystemExit("[Rogues 1.20.1 compat batch4] commented shouldContinue fallback survived")
for required in (
    '@Inject(method = "shouldContinue", at = @At("HEAD"), cancellable = true)',
    'target.hasStatusEffect(RogueEffects.STEALTH.effect)',
    'target.hasStatusEffect(RogueEffects.SHADOW_STEP.effect)',
    'cir.setReturnValue(false);',
):
    if required not in text:
        raise SystemExit(f"[Rogues 1.20.1 compat batch4] required target-loss semantic missing: {required}")
print("[Rogues 1.20.1 compat batch4] activated upstream-authored TrackTargetGoal shouldContinue fallback for Stealth/Shadow Step on Forge 1.20.1")
