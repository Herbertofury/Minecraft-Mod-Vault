#!/usr/bin/env python3
from __future__ import annotations
import pathlib, sys

root = pathlib.Path(sys.argv[1]).resolve()
path = root / "net/rogues/mixin/TrackTargetGoalStealth.java"
text = path.read_text(encoding="utf-8")
old = '''    @Inject(method = "shouldContinue", at = @At("HEAD"), cancellable = true)
    private void shouldContinue_HEAD(CallbackInfoReturnable<Boolean> cir) {
        var target = mob.getTarget();
        if (target != null
                && (target.hasStatusEffect(RogueEffects.STEALTH.effect) || target.hasStatusEffect(RogueEffects.SHADOW_STEP.effect))) {
            cir.setReturnValue(false);
        }
    }
'''
new = '''    // Forge 1.20.1 target acquisition is guarded by ActiveTargetGoalStealth.
    // Once a target is set, the upstream getFollowRange hook below preserves the
    // configurable stealth_follow_range contract instead of making Stealth absolute.
'''
count = text.count(old)
if count != 1:
    raise SystemExit(f"[Rogues 1.20.1 compat batch6] expected one batch4 continuation fallback, found {count}")
text = text.replace(old, new, 1)
path.write_text(text, encoding="utf-8")
if '@Inject(method = "shouldContinue"' in text:
    raise SystemExit("[Rogues 1.20.1 compat batch6] overbroad shouldContinue guard survived")
for required in (
    '@Inject(method = "getFollowRange", at = @At("HEAD"), cancellable = true)',
    'target.hasStatusEffect(RogueEffects.STEALTH.effect)',
    'target.hasStatusEffect(RogueEffects.SHADOW_STEP.effect)',
    'RoguesMod.tweaksConfig.value.stealth_follow_range',
):
    if required not in text:
        raise SystemExit(f"[Rogues 1.20.1 compat batch6] upstream configurable range semantic missing: {required}")
print("[Rogues 1.20.1 compat batch6] superseded the absolute batch4 continuation fallback; configurable Stealth/Shadow Step range remains authoritative")
