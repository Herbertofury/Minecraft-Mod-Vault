#!/usr/bin/env python3
from __future__ import annotations
import json, pathlib, sys

if len(sys.argv) != 3:
    raise SystemExit("usage: apply_1201_compat_batch7.py <java-root> <resources-root>")
java_root = pathlib.Path(sys.argv[1]).resolve()
resources_root = pathlib.Path(sys.argv[2]).resolve()
path = java_root / "net/rogues/mixin/MobEntityTargetStealth.java"
if path.exists():
    raise SystemExit("[Rogues batch7] universal target mixin unexpectedly exists")
path.parent.mkdir(parents=True, exist_ok=True)
text = '''package net.rogues.mixin;

import net.minecraft.entity.LivingEntity;
import net.minecraft.entity.mob.MobEntity;
import net.rogues.RoguesMod;
import net.rogues.effect.RogueEffects;
import org.jetbrains.annotations.Nullable;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.injection.At;
import org.spongepowered.asm.mixin.injection.Inject;
import org.spongepowered.asm.mixin.injection.callback.CallbackInfo;

@Mixin(MobEntity.class)
public abstract class MobEntityTargetStealth {
    @Inject(method = "setTarget", at = @At("HEAD"), cancellable = true)
    private void setTarget_HEAD(@Nullable LivingEntity target, CallbackInfo ci) {
        if (target == null || !(target.hasStatusEffect(RogueEffects.STEALTH.effect) || target.hasStatusEffect(RogueEffects.SHADOW_STEP.effect))) return;
        var self = (MobEntity) (Object) this;
        double range = RoguesMod.tweaksConfig.value.stealth_follow_range;
        if (self.squaredDistanceTo(target) > range * range) ci.cancel();
    }
}
'''
path.write_text(text, encoding="utf-8")
for required in ('@Mixin(MobEntity.class)', 'method = "setTarget"', 'RogueEffects.STEALTH.effect', 'RogueEffects.SHADOW_STEP.effect', 'stealth_follow_range', 'squaredDistanceTo(target)', 'ci.cancel()'):
    if required not in text:
        raise SystemExit(f"[Rogues batch7] missing semantic: {required}")
mixins_path = resources_root / "rogues.mixins.json"
data = json.loads(mixins_path.read_text(encoding="utf-8"))
expected = ["LivingEntityStealth", "TrackTargetGoalStealth", "ActiveTargetGoalStealth"]
if data.get("mixins") != expected:
    raise SystemExit(f"[Rogues batch7] unexpected mixin frontier: {data.get('mixins')!r}")
data["mixins"].append("MobEntityTargetStealth")
if data.get("required") is not True or data.get("injectors", {}).get("defaultRequire") != 1:
    raise SystemExit("[Rogues batch7] fail-closed mixin contract changed")
mixins_path.write_text(json.dumps(data, indent=2) + "\n", encoding="utf-8")
print("[Rogues 1.20.1 compat batch7] enforced configured Stealth/Shadow Step range at universal MobEntity#setTarget boundary")
