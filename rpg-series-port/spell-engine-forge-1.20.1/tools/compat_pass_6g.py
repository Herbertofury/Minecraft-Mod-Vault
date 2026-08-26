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

    // The common project compiles against TinyConfig, but `common { transitive = false }` intentionally
    // keeps dependency mods/libraries out of the transformed common artifact. Forge dev runs therefore
    // need TinyConfig explicitly on the Forge runtime classpath, and release JARs need it embedded.
    // Mirror the already runtime-proven Spell Power 1.20.1 packaging pattern.
    def tinyConfig = implementation("com.github.ZsoltMolnarrr:TinyConfig:$rootProject.tiny_config_version")
    include tinyConfig
    forgeRuntimeLibrary tinyConfig
'''
if 'mixinextras-forge' not in s:
    if anchor not in s:
        raise SystemExit('Forge dependency anchor missing')
    s = s.replace(anchor, anchor + addition)
elif 'forgeRuntimeLibrary tinyConfig' not in s:
    if anchor not in s:
        raise SystemExit('Forge dependency anchor missing for TinyConfig repair')
    tiny = '''
    // TinyConfig is required by SpellEngineMod static config managers at runtime.
    def tinyConfig = implementation("com.github.ZsoltMolnarrr:TinyConfig:$rootProject.tiny_config_version")
    include tinyConfig
    forgeRuntimeLibrary tinyConfig
'''
    s = s.replace(anchor, anchor + tiny)
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

# Forge 47 patches LootTable.pools from vanilla's LootPool[] to a mutable List<LootPool> so its
# loot-table hooks can add/remove pools. The field name stays `pools`, but the runtime descriptor is
# java.util.List. Match the Forge-patched descriptor exactly or Mixin's accessor fails during APPLY.
loot_accessor = root / 'common/src/main/java/net/spell_engine/mixin/loot/LootTableAccessor.java'
la = loot_accessor.read_text()
import_anchor = 'import net.minecraft.loot.LootTable;\nimport org.spongepowered.asm.mixin.Mixin;\nimport org.spongepowered.asm.mixin.gen.Accessor;'
if import_anchor not in la:
    raise SystemExit('LootTableAccessor import anchor missing')
la = la.replace(import_anchor, import_anchor + '\n\nimport java.util.List;', 1)
if 'LootPool[] spellEngine_pools();' not in la:
    raise SystemExit('vanilla-array LootTableAccessor signature missing')
la = la.replace('LootPool[] spellEngine_pools();', 'List<LootPool> spellEngine_pools();', 1)
loot_accessor.write_text(la)

forge_events = root / 'forge/src/main/java/net/spell_engine/forge/PlatformEventsImpl.java'
fe = forge_events.read_text().replace('import java.util.Arrays;\n', '')
old_pools = 'this.existingPools = List.copyOf(Arrays.asList(((LootTableAccessor) (Object) event.getTable()).spellEngine_pools()));'
new_pools = 'this.existingPools = List.copyOf(((LootTableAccessor) (Object) event.getTable()).spellEngine_pools());'
if old_pools not in fe:
    raise SystemExit('ForgeLootContext array-copy expression missing')
fe = fe.replace(old_pools, new_pools, 1)
forge_events.write_text(fe)

final = forge_build.read_text()
for required in (
    'mixinextras-common',
    'implementation(include("io.github.llamalad7:mixinextras-forge:',
    'def tinyConfig = implementation("com.github.ZsoltMolnarrr:TinyConfig:',
    'include tinyConfig',
    'forgeRuntimeLibrary tinyConfig',
):
    if required not in final:
        raise SystemExit(f'Forge runtime dependency missing: {required}')
final_loot = loot_accessor.read_text()
if 'List<LootPool> spellEngine_pools();' not in final_loot or 'LootPool[] spellEngine_pools();' in final_loot:
    raise SystemExit('pass6g missing Forge-patched LootTable List accessor')
if 'Arrays.asList' in forge_events.read_text():
    raise SystemExit('pass6g left stale array handling in ForgeLootContext')
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
print('Spell Engine compatibility pass 6g applied: embedded MixinExtras + TinyConfig runtimes + exact effect-removal hook + Forge LootTable List accessor')
