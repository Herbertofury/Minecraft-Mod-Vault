#!/usr/bin/env python3
from __future__ import annotations
import json, pathlib, sys

if len(sys.argv) != 3:
    raise SystemExit("usage: apply_1201_compat_batch5.py <java-root> <resources-root>")
java_root = pathlib.Path(sys.argv[1]).resolve()
resources_root = pathlib.Path(sys.argv[2]).resolve()
path = java_root / "net/rogues/mixin/ActiveTargetGoalStealth.java"
if path.exists():
    raise SystemExit(f"[Rogues 1.20.1 compat batch5] acquisition mixin unexpectedly already exists: {path}")
path.parent.mkdir(parents=True, exist_ok=True)
text = '''package net.rogues.mixin;

import net.minecraft.entity.LivingEntity;
import net.minecraft.entity.ai.goal.ActiveTargetGoal;
import net.minecraft.entity.ai.goal.TrackTargetGoal;
import net.minecraft.entity.mob.MobEntity;
import net.rogues.RoguesMod;
import net.rogues.effect.RogueEffects;
import org.jetbrains.annotations.Nullable;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.Shadow;
import org.spongepowered.asm.mixin.injection.At;
import org.spongepowered.asm.mixin.injection.Inject;
import org.spongepowered.asm.mixin.injection.callback.CallbackInfoReturnable;

/**
 * Forge 1.20.1 acquisition guard for the upstream Stealth follow-range contract.
 * ActiveTargetGoal selects targetEntity before TrackTargetGoal.mob#getTarget is set,
 * so the upstream getFollowRange hook cannot reduce the initial search radius.
 */
@Mixin(ActiveTargetGoal.class)
public abstract class ActiveTargetGoalStealth extends TrackTargetGoal {
    @Shadow protected @Nullable LivingEntity targetEntity;

    protected ActiveTargetGoalStealth(MobEntity mob, boolean checkVisibility) {
        super(mob, checkVisibility);
    }

    @Inject(method = "canStart", at = @At("RETURN"), cancellable = true)
    private void canStart_RETURN(CallbackInfoReturnable<Boolean> cir) {
        if (!cir.getReturnValue()) {
            return;
        }
        var candidate = targetEntity;
        if (candidate == null
                || !(candidate.hasStatusEffect(RogueEffects.STEALTH.effect)
                    || candidate.hasStatusEffect(RogueEffects.SHADOW_STEP.effect))) {
            return;
        }
        double stealthRange = RoguesMod.tweaksConfig.value.stealth_follow_range;
        if (mob.squaredDistanceTo(candidate) > stealthRange * stealthRange) {
            cir.setReturnValue(false);
        }
    }
}
'''
path.write_text(text, encoding="utf-8")
for required in (
    '@Mixin(ActiveTargetGoal.class)',
    '@Shadow protected @Nullable LivingEntity targetEntity;',
    '@Inject(method = "canStart", at = @At("RETURN"), cancellable = true)',
    'candidate.hasStatusEffect(RogueEffects.STEALTH.effect)',
    'candidate.hasStatusEffect(RogueEffects.SHADOW_STEP.effect)',
    'RoguesMod.tweaksConfig.value.stealth_follow_range',
    'mob.squaredDistanceTo(candidate) > stealthRange * stealthRange',
    'cir.setReturnValue(false);',
):
    if required not in text:
        raise SystemExit(f"[Rogues 1.20.1 compat batch5] required acquisition semantic missing: {required}")

mixins_path = resources_root / "rogues.mixins.json"
data = json.loads(mixins_path.read_text(encoding="utf-8"))
mixins = data.get("mixins")
if mixins != ["LivingEntityStealth", "TrackTargetGoalStealth"]:
    raise SystemExit(f"[Rogues 1.20.1 compat batch5] unexpected upstream mixin frontier: {mixins!r}")
mixins.append("ActiveTargetGoalStealth")
mixins_path.write_text(json.dumps(data, indent=2) + "\n", encoding="utf-8")
check = json.loads(mixins_path.read_text(encoding="utf-8"))
if check.get("required") is not True or check.get("injectors", {}).get("defaultRequire") != 1:
    raise SystemExit("[Rogues 1.20.1 compat batch5] fail-closed mixin contract changed")
if check.get("mixins", []).count("ActiveTargetGoalStealth") != 1:
    raise SystemExit("[Rogues 1.20.1 compat batch5] acquisition mixin registration failed")
print("[Rogues 1.20.1 compat batch5] guarded ActiveTargetGoal acquisition with the configured Stealth/Shadow Step follow range")
