#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 3:
    raise SystemExit('usage: compat_pass_6g.py <generated-port-root> <spell-engine-1.20.1-baseline>')
root = Path(sys.argv[1]).resolve()
forge_build = root / 'forge/build.gradle'
s = forge_build.read_text()
anchor = '    modImplementation "me.shedaniel.cloth:cloth-config-forge:$rootProject.cloth_config_version"\n'
addition = '''    // Spell Engine common Mixins use MixinExtras at runtime. Embed the Forge service JAR exactly
    // like the already-runtime-proven Ranged Weapon API port; compileOnly keeps AP symbols explicit.
    compileOnly(annotationProcessor("io.github.llamalad7:mixinextras-common:$rootProject.mixinextras_version"))
    implementation(include("io.github.llamalad7:mixinextras-forge:$rootProject.mixinextras_version"))
'''
if 'mixinextras-forge' not in s:
    if anchor not in s:
        raise SystemExit('Forge dependency anchor missing')
    s = s.replace(anchor, anchor + addition)
forge_build.write_text(s)

# 1.10.2 targets the newer StatusEffect.onRemoved(AttributeContainer) call. Minecraft 1.20.1 uses
# StatusEffect.onRemoved(LivingEntity, AttributeContainer, int). Keep the WrapOperation at the same
# call site and preserve ordering: vanilla removal first, then Spell Engine's OnRemoval handler.
effect_removal = root / 'common/src/main/java/net/spell_engine/mixin/effect/LivingEntityEffectRemoval.java'
e = effect_removal.read_text()
old_target = 'target = "Lnet/minecraft/entity/effect/StatusEffect;onRemoved(Lnet/minecraft/entity/attribute/AttributeContainer;)V"'
new_target = 'target = "Lnet/minecraft/entity/effect/StatusEffect;onRemoved(Lnet/minecraft/entity/LivingEntity;Lnet/minecraft/entity/attribute/AttributeContainer;I)V"'
old_handler = '''    private void onStatusEffectRemoved_Wrap_onRemoved(StatusEffect instance, AttributeContainer attributeContainer, Operation<Void> original) {\n        original.call(instance, attributeContainer);'''
new_handler = '''    private void onStatusEffectRemoved_Wrap_onRemoved(StatusEffect instance, LivingEntity removedFrom, AttributeContainer attributeContainer, int amplifier, Operation<Void> original) {\n        original.call(instance, removedFrom, attributeContainer, amplifier);'''
if old_target not in e:
    raise SystemExit('modern StatusEffect.onRemoved(AttributeContainer) target missing')
if old_handler not in e:
    raise SystemExit('modern LivingEntityEffectRemoval wrapper signature missing')
e = e.replace(old_target, new_target, 1)
e = e.replace(old_handler, new_handler, 1)
effect_removal.write_text(e)

final = forge_build.read_text()
for required in ('mixinextras-common', 'implementation(include("io.github.llamalad7:mixinextras-forge:'):
    if required not in final:
        raise SystemExit(f'MixinExtras Forge runtime dependency missing: {required}')
final_effect = effect_removal.read_text()
for stale in (
    'onRemoved(Lnet/minecraft/entity/attribute/AttributeContainer;)V',
    'original.call(instance, attributeContainer);',
    'LivingEntity entity, AttributeContainer attributeContainer, int amplifier, Operation<Void> original',
):
    if stale in final_effect:
        raise SystemExit(f'pass6g left stale/ambiguous removal hook: {stale}')
for required in (
    'onRemoved(Lnet/minecraft/entity/LivingEntity;Lnet/minecraft/entity/attribute/AttributeContainer;I)V',
    'StatusEffect instance, LivingEntity removedFrom, AttributeContainer attributeContainer, int amplifier, Operation<Void> original',
    'original.call(instance, removedFrom, attributeContainer, amplifier);',
    'var entity = (LivingEntity) (Object) this;',
):
    if required not in final_effect:
        raise SystemExit(f'pass6g missing 1.20.1 removal hook: {required}')
print('Spell Engine compatibility pass 6g applied: embedded Forge MixinExtras runtime service + exact 1.20.1 effect-removal hook')
