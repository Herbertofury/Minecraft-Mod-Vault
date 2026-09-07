#!/usr/bin/env python3
from __future__ import annotations
import json, pathlib, sys

if len(sys.argv) != 3:
    raise SystemExit("usage: apply_1201_compat_batch8.py <java-root> <resources-root>")
java_root = pathlib.Path(sys.argv[1]).resolve()
resources_root = pathlib.Path(sys.argv[2]).resolve()
path = java_root / "net/rogues/mixin/MobEntityTargetReadStealth.java"
if path.exists():
    raise SystemExit("[Rogues batch8] target-read mixin unexpectedly exists")
path.parent.mkdir(parents=True, exist_ok=True)
text = '''package net.rogues.mixin;

import net.minecraft.entity.LivingEntity;
import net.minecraft.entity.mob.MobEntity;
import net.rogues.RoguesMod;
import net.rogues.effect.RogueEffects;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.injection.At;
import org.spongepowered.asm.mixin.injection.Inject;
import org.spongepowered.asm.mixin.injection.callback.CallbackInfoReturnable;

@Mixin(MobEntity.class)
public abstract class MobEntityTargetReadStealth {
    @Inject(method = "getTarget", at = @At("RETURN"), cancellable = true)
    private void getTarget_RETURN(CallbackInfoReturnable<LivingEntity> cir) {
        LivingEntity target = cir.getReturnValue();
        if (target == null) return;
        if (!(target.hasStatusEffect(RogueEffects.STEALTH.effect) || target.hasStatusEffect(RogueEffects.SHADOW_STEP.effect))) return;
        var self = (MobEntity) (Object) this;
        double range = RoguesMod.tweaksConfig.value.stealth_follow_range;
        if (self.squaredDistanceTo(target) > range * range) cir.setReturnValue(null);
    }
}
'''
path.write_text(text, encoding="utf-8")
for required in ('@Mixin(MobEntity.class)', 'method = "getTarget"', '@At("RETURN")', 'CallbackInfoReturnable<LivingEntity>', 'RogueEffects.STEALTH.effect', 'RogueEffects.SHADOW_STEP.effect', 'stealth_follow_range', 'squaredDistanceTo(target)', 'cir.setReturnValue(null)'):
    if required not in text:
        raise SystemExit(f"[Rogues batch8] missing semantic: {required}")
mixins_path = resources_root / "rogues.mixins.json"
data = json.loads(mixins_path.read_text(encoding="utf-8"))
expected = ["LivingEntityStealth", "TrackTargetGoalStealth", "ActiveTargetGoalStealth", "MobEntityTargetStealth"]
if data.get("mixins") != expected:
    raise SystemExit(f"[Rogues batch8] unexpected mixin frontier: {data.get('mixins')!r}")
data["mixins"].append("MobEntityTargetReadStealth")
if data.get("required") is not True or data.get("injectors", {}).get("defaultRequire") != 1:
    raise SystemExit("[Rogues batch8] fail-closed mixin contract changed")
mixins_path.write_text(json.dumps(data, indent=2) + "\n", encoding="utf-8")
print("[Rogues 1.20.1 compat batch8] filtered Stealth/Shadow Step targets at universal MobEntity#getTarget read boundary outside configured reveal range")
