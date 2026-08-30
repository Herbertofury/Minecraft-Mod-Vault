#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 2:
    raise SystemExit('usage: prepare_more_rpg_library_1201_status_consumer_prepass.py <generated-port-root>')
root = Path(sys.argv[1]).resolve()
java = root / 'common/src/main/java'
if not java.is_dir():
    raise SystemExit(f'missing common Java root: {java}')

# #331 proved five non-effect consumers still use the 1.21 RegistryEntry-valued status-effect API.
# #332 pinned the exact cardinality: four owners have one seam each and SpellthiefImpact has two.
# Keep those per-owner counts exact. The /effect package remains owned by Wave 2 proper, so this
# prepass cannot broaden into an unchecked global substitution.
owners = {
    'net/more_rpg_classes/mixin/LivingEntityStealth.java': 1,
    'net/more_rpg_classes/mixin/ControlOwnerMixin.java': 1,
    'net/more_rpg_classes/mixin/LivingEntityMixin.java': 1,
    'net/more_rpg_classes/mixin/TrackTargetGoalStealth.java': 1,
    'net/more_rpg_classes/custom/spell_impacts/SpellthiefImpact.java': 2,
}
needle = '.getEffectType().value()'
replacement = '.getEffectType()'
rewritten = 0
for rel, expected in owners.items():
    path = java / rel
    if not path.is_file():
        raise SystemExit(f'missing pinned #331/#332 status consumer: {rel}')
    text = path.read_text(encoding='utf-8')
    count = text.count(needle)
    if count != expected:
        raise SystemExit(f'#331/#332 status consumer seam drifted: {rel} expected={expected} found={count}')
    path.write_text(text.replace(needle, replacement), encoding='utf-8')
    rewritten += count

if rewritten != 6:
    raise SystemExit(f'#331/#332 status consumer repair incomplete: expected=6 rewritten={rewritten}')

# #348 is the first packaged Forge server run with More RPG's Mixin config actually registered.
# It proved LivingEntityStealth still targets the modern RegistryEntry/Holder overload at runtime.
# Minecraft 1.20.1 exposes hasStatusEffect(StatusEffect) instead, so adapt the WrapOperation target
# and handler parameter together. Keep exact cardinality guards so a future upstream change cannot
# silently broaden or weaken this production-Mixin repair.
stealth = java / 'net/more_rpg_classes/mixin/LivingEntityStealth.java'
text = stealth.read_text(encoding='utf-8')
import_old = 'import net.minecraft.registry.entry.RegistryEntry;\n'
target_old = 'target = "Lnet/minecraft/entity/LivingEntity;hasStatusEffect(Lnet/minecraft/registry/entry/RegistryEntry;)Z"'
target_new = 'target = "Lnet/minecraft/entity/LivingEntity;hasStatusEffect(Lnet/minecraft/entity/effect/StatusEffect;)Z"'
param_old = 'RegistryEntry<StatusEffect> effect'
param_new = 'StatusEffect effect'
for label, seam in (
    ('RegistryEntry import', import_old),
    ('WrapOperation target', target_old),
    ('WrapOperation parameter', param_old),
):
    count = text.count(seam)
    if count != 1:
        raise SystemExit(f'#348 LivingEntityStealth {label} seam drifted: expected=1 found={count}')
text = text.replace(import_old, '', 1)
text = text.replace(target_old, target_new, 1)
text = text.replace(param_old, param_new, 1)
if 'RegistryEntry<StatusEffect>' in text or 'minecraft/registry/entry/RegistryEntry' in text:
    raise SystemExit('#348 LivingEntityStealth RegistryEntry status-effect seam survived repair')
if text.count(target_new) != 1 or text.count(param_new) != 1:
    raise SystemExit('#348 LivingEntityStealth raw StatusEffect WrapOperation contract missing or duplicated')
stealth.write_text(text, encoding='utf-8')

print('[More RPG 2.7.2] STATUS_EFFECT_NON_EFFECT_CONSUMERS_1201_PASS owners=5 calls=6')
print('[More RPG 2.7.2] LIVING_ENTITY_STEALTH_1201_WRAP_TARGET_PASS target=hasStatusEffect(StatusEffect) source=run-348-packaged-server')
