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

# #349 advances past the #348 selector failure and then dies in Forge/ModLauncher while loading
# org.spongepowered.asm.synthetic.args.Args$1. More RPG owns the only ModifyArgs injector class in
# the packaged runtime: LivingEntityMixin, where three handlers all modify only applyDamage argument
# index 1 (the float amount). Mixin's lightweight ModifyArg injector can consume the complete subject
# invocation arguments (DamageSource, float) without generating any synthetic Args subclass. Preserve
# the three handlers and their ordering/side effects; only replace the argument-bundle transport.
living = java / 'net/more_rpg_classes/mixin/LivingEntityMixin.java'
text = living.read_text(encoding='utf-8')
imports_old = ('import org.spongepowered.asm.mixin.injection.ModifyArgs;\n'
               'import org.spongepowered.asm.mixin.injection.invoke.arg.Args;\n')
imports_new = 'import org.spongepowered.asm.mixin.injection.ModifyArg;\n'
if text.count(imports_old) != 1:
    raise SystemExit(f'#349 ModifyArgs import seam drifted: expected=1 found={text.count(imports_old)}')
text = text.replace(imports_old, imports_new, 1)
annotation_old = ('@ModifyArgs(method = "damage", at = @At(value = "INVOKE", '
                  'target = "Lnet/minecraft/entity/LivingEntity;applyDamage(Lnet/minecraft/entity/damage/DamageSource;F)V"))')
annotation_new = ('@ModifyArg(method = "damage", at = @At(value = "INVOKE", '
                  'target = "Lnet/minecraft/entity/LivingEntity;applyDamage(Lnet/minecraft/entity/damage/DamageSource;F)V"), index = 1)')
if text.count(annotation_old) != 3:
    raise SystemExit(f'#349 ModifyArgs annotation cardinality drifted: expected=3 found={text.count(annotation_old)}')
text = text.replace(annotation_old, annotation_new)

def rewrite_modifyarg_handler(source, old_signature, new_signature, args_set_old, amount_update, expected_early_returns):
    if source.count(old_signature) != 1:
        raise SystemExit(f'#349 handler signature drifted: {old_signature} found={source.count(old_signature)}')
    sig_at = source.index(old_signature)
    body_at = source.index('{', sig_at)
    depth = 0
    end = None
    for i in range(body_at, len(source)):
        if source[i] == '{':
            depth += 1
        elif source[i] == '}':
            depth -= 1
            if depth == 0:
                end = i
                break
    if end is None:
        raise SystemExit(f'#349 could not resolve handler body: {old_signature}')
    block = source[sig_at:end + 1]
    if block.count('DamageSource source = args.get(0);') != 1:
        raise SystemExit(f'#349 handler source extraction drifted for {old_signature}')
    if block.count(args_set_old) != 1:
        raise SystemExit(f'#349 handler amount mutation drifted for {old_signature}')
    if block.count('return;') != expected_early_returns:
        raise SystemExit(f'#349 handler early-return cardinality drifted for {old_signature}: expected={expected_early_returns} found={block.count("return;")}')
    block = block.replace(old_signature, new_signature, 1)
    block = block.replace('        DamageSource source = args.get(0);\n', '', 1)
    block = block.replace('return;', 'return amount;')
    block = block.replace(args_set_old, amount_update, 1)
    block = block[:-1] + '        return amount;\n    }'
    return source[:sig_at] + block + source[end + 1:]

text = rewrite_modifyarg_handler(
    text,
    'private void rage$addRageDamage(Args args)',
    'private float rage$addRageDamage(DamageSource source, float amount)',
    'args.set(1, (float) args.get(1) + rageDamage);',
    'amount += rageDamage;',
    6,
)
text = rewrite_modifyarg_handler(
    text,
    'private void duelistsFocus$reduceDamage(Args args)',
    'private float duelistsFocus$reduceDamage(DamageSource source, float amount)',
    'args.set(1, (float) args.get(1) * 0.75F);',
    'amount *= 0.75F;',
    2,
)
text = rewrite_modifyarg_handler(
    text,
    'private void duelistsFocus$increaseDamageToTarget(Args args)',
    'private float duelistsFocus$increaseDamageToTarget(DamageSource source, float amount)',
    'args.set(1, (float) args.get(1) * 1.25F);',
    'amount *= 1.25F;',
    2,
)
if 'ModifyArgs' in text or 'invoke.arg.Args' in text or 'Args args' in text or 'args.get(' in text or 'args.set(' in text:
    raise SystemExit('#349 synthetic Args injector seam survived LivingEntityMixin repair')
if text.count('@ModifyArg(') != 3 or text.count('index = 1)') != 3:
    raise SystemExit('#349 ModifyArg injector contract missing or duplicated')
for name in ('rage$addRageDamage', 'duelistsFocus$reduceDamage', 'duelistsFocus$increaseDamageToTarget'):
    marker = f'private float {name}(DamageSource source, float amount)'
    if text.count(marker) != 1:
        raise SystemExit(f'#349 ModifyArg handler missing after repair: {name}')
living.write_text(text, encoding='utf-8')

print('[More RPG 2.7.2] STATUS_EFFECT_NON_EFFECT_CONSUMERS_1201_PASS owners=5 calls=6')
print('[More RPG 2.7.2] LIVING_ENTITY_STEALTH_1201_WRAP_TARGET_PASS target=hasStatusEffect(StatusEffect) source=run-348-packaged-server')
print('[More RPG 2.7.2] LIVING_ENTITY_MODIFYARG_1201_PASS handlers=3 target=applyDamage amount_index=1 synthetic_args=eliminated source=run-349-packaged-server')
