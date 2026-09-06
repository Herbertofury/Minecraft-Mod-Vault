#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 2:
    raise SystemExit('usage: prepare_more_rpg_library_1201_api_wave5b.py <prepared-root>')
root = Path(sys.argv[1]).resolve()
java = root / 'common/src/main/java'
if not java.is_dir():
    raise SystemExit(f'missing common Java root: {java}')

def owner(rel):
    p = java / rel
    if not p.is_file():
        raise SystemExit(f'missing Wave 5b owner: {p}')
    return p

def replace_exact(s, needle, repl, expected, label):
    found = s.count(needle)
    if found != expected:
        raise SystemExit(f'{label} seam drifted: expected={expected} found={found}')
    return s.replace(needle, repl)

# Certified Spell Engine 1.20.1 RegistryCompat bridge for Identifier lookups.
p = owner('net/more_rpg_classes/client/render/MobBeamTracker.java')
s = p.read_text()
s = replace_exact(s,
    'SpellRegistry.from(world).getEntry(payload.spellId()).orElse(null)',
    'net.spell_engine.compat.registry.RegistryCompat.entry(SpellRegistry.from(world), payload.spellId()).orElse(null)',
    1, 'MobBeamTracker spell lookup')
p.write_text(s)

p = owner('net/more_rpg_classes/util/CustomMethods.java')
s = p.read_text()
s = replace_exact(s,
    'registry.getEntry(spellId).orElse(null)',
    'net.spell_engine.compat.registry.RegistryCompat.entry(registry, spellId).orElse(null)',
    1, 'CustomMethods spell lookup')
p.write_text(s)

p = owner('net/more_rpg_classes/entity/goal/MobSpellCastGoal.java')
s = p.read_text()
s = replace_exact(s,
    'SpellRegistry.from(world).getEntry(new Identifier(spellOrTag)).orElse(null)',
    'net.spell_engine.compat.registry.RegistryCompat.entry(SpellRegistry.from(world), new Identifier(spellOrTag)).orElse(null)',
    1, 'MobSpellCastGoal resolve lookup')
s = replace_exact(s,
    'registry.getEntry(id).orElse(null)',
    'net.spell_engine.compat.registry.RegistryCompat.entry(registry, id).orElse(null)',
    2, 'MobSpellCastGoal intelligent lookups')
s = replace_exact(s,
    'SpellRegistry.from(caster.asMobEntity().getWorld()).getEntry(activeSpellId).orElse(null)',
    'net.spell_engine.compat.registry.RegistryCompat.entry(SpellRegistry.from(caster.asMobEntity().getWorld()), activeSpellId).orElse(null)',
    2, 'MobSpellCastGoal caster active lookup')
s = replace_exact(s,
    'SpellRegistry.from(entity.getWorld()).getEntry(activeSpellId).orElse(null)',
    'net.spell_engine.compat.registry.RegistryCompat.entry(SpellRegistry.from(entity.getWorld()), activeSpellId).orElse(null)',
    2, 'MobSpellCastGoal entity active lookup')
s = replace_exact(s,
    'net.minecraft.registry.Registries.STATUS_EFFECT\n                                .getEntry(new net.minecraft.util.Identifier(stash.id))',
    'java.util.Optional.ofNullable(net.minecraft.registry.Registries.STATUS_EFFECT\n                                .get(new net.minecraft.util.Identifier(stash.id)))',
    1, 'MobSpellCastGoal stash effect lookup')
p.write_text(s)

# RegistryKey-compatible damage type lookup.
p = owner('net/more_rpg_classes/effect/BleedingEffect.java')
s = p.read_text()
s = replace_exact(s,
    'reg.getEntry(new net.minecraft.util.Identifier("more_rpg_classes", "bleeding"))',
    'net.spell_engine.compat.registry.RegistryCompat.entry(reg, new net.minecraft.util.Identifier("more_rpg_classes", "bleeding"))',
    1, 'BleedingEffect damage type lookup')
p.write_text(s)

# Raw target-era EntityAttribute + StatusEffect consumer shapes.
p = owner('net/more_rpg_classes/mixin/PersistentProjectileEntityMixin.java')
s = p.read_text()
s = replace_exact(s, 'RegistryEntry<EntityAttribute> fuseAttribute', 'EntityAttribute fuseAttribute', 1,
                  'PersistentProjectileEntityMixin fuse holder')
effects = {
    'MRPGCEffects.IGNITED.entry': 'MRPGCEffects.IGNITED.effect',
    'MRPGCEffects.STAGGER.entry': 'MRPGCEffects.STAGGER.effect',
    'MRPGCEffects.FROZEN_SOLID.entry': 'MRPGCEffects.FROZEN_SOLID.effect',
    'SpellEngineEffects.STUN.entry': 'SpellEngineEffects.STUN.effect',
    'SpellEngineEffects.BLEED.entry': 'SpellEngineEffects.BLEED.effect',
}
for old, new in effects.items():
    s = replace_exact(s, old, new, 1, f'PersistentProjectileEntityMixin {old}')
if 'RegistryEntry<' not in s:
    s = replace_exact(s, 'import net.minecraft.registry.entry.RegistryEntry;\n', '', 1,
                      'PersistentProjectileEntityMixin RegistryEntry import')
else:
    raise SystemExit('PersistentProjectileEntityMixin unexpected RegistryEntry usage survived Wave 5b')
p.write_text(s)

p = owner('net/more_rpg_classes/mixin/LivingEntityMixin.java')
s = p.read_text()
s = replace_exact(s,
    '@Shadow public abstract boolean hasStatusEffect(RegistryEntry<StatusEffect> effect);',
    '@Shadow public abstract boolean hasStatusEffect(StatusEffect effect);',
    1, 'LivingEntityMixin shadow status holder')
s = replace_exact(s,
    'RegistryEntry<EntityAttribute> fuseAttribute', 'EntityAttribute fuseAttribute', 1,
    'LivingEntityMixin fuse holder')
living_effect_counts = {
    'MRPGCEffects.DUELISTS_FOCUS_OWNER.entry': 2,
    'MRPGCEffects.DUELISTS_FOCUS_TARGET.entry': 2,
    'MRPGCEffects.IGNITED.entry': 1,
    'MRPGCEffects.STAGGER.entry': 1,
    'MRPGCEffects.FROZEN_SOLID.entry': 2,
    'MRPGCEffects.FROSTED.entry': 1,
    'SpellEngineEffects.STUN.entry': 1,
    'SpellEngineEffects.BLEED.entry': 1,
}
for old, count in living_effect_counts.items():
    s = replace_exact(s, old, old.removesuffix('.entry') + '.effect', count, f'LivingEntityMixin {old}')
s = replace_exact(s,
    'RegistryEntry<StatusEffect> effectType = effect.getEffectType();',
    'StatusEffect effectType = effect.getEffectType();',
    1, 'LivingEntityMixin tenacity raw effect')
omen_block = '''        if (effectType.matchesKey(StatusEffects.BAD_OMEN.getKey().get()) ||\n            effectType.matchesKey(StatusEffects.TRIAL_OMEN.getKey().get()) ||\n            effectType.matchesKey(StatusEffects.RAID_OMEN.getKey().get())) return;'''
s = replace_exact(s, omen_block,
    '        if (effectType == StatusEffects.BAD_OMEN) return;',
    1, 'LivingEntityMixin target-era omen set')
if 'RegistryEntry<' not in s:
    s = replace_exact(s, 'import net.minecraft.registry.entry.RegistryEntry;\n', '', 1,
                      'LivingEntityMixin RegistryEntry import')
else:
    raise SystemExit('LivingEntityMixin unexpected RegistryEntry usage survived Wave 5b')
p.write_text(s)

for rel, refs in {
    'net/more_rpg_classes/mixin/DrawHeartsMixin.java': {'MRPGCEffects.FATAL_POISON.entry': 1},
    'net/more_rpg_classes/mixin/HudMessagesMixin.java': {
        'MRPGCEffects.FROZEN_SOLID.entry': 1,
        'MRPGCEffects.IGNITED.entry': 1,
        'MRPGCEffects.FEAR.entry': 1,
        'MRPGCEffects.STAGGER.entry': 1,
    },
}.items():
    p = owner(rel)
    s = p.read_text()
    for old, count in refs.items():
        s = replace_exact(s, old, old.removesuffix('.entry') + '.effect', count, f'{Path(rel).name} {old}')
    p.write_text(s)

# Status effect registry ID on raw 1.20.1 StatusEffect.
p = owner('net/more_rpg_classes/custom/spell_impacts/SpellthiefImpact.java')
s = p.read_text()
s = replace_exact(s,
    'Identifier effectId = effect.getEffectType().getKey().map(k -> k.getValue()).orElse(null);',
    'Identifier effectId = net.minecraft.registry.Registries.STATUS_EFFECT.getId(effect.getEffectType());',
    1, 'SpellthiefImpact raw status effect id')
p.write_text(s)

# Target 1.20.1 entity builder name.
p = owner('net/more_rpg_classes/entity/MRPGCEntities.java')
s = p.read_text()
s = replace_exact(s, '.dimensions(0.0F, 0.0F)', '.setDimensions(0.0F, 0.0F)', 1,
                  'MRPGCEntities dimensions')
p.write_text(s)

# 1.20.1 StructureProcessorType consumes Codec, while modern processors expose MapCodec.
p = owner('net/more_rpg_classes/worldgen/ModStructureProcessorTypes.java')
s = p.read_text()
for name in ('PathAdaptationProcessor', 'WaterPillarProcessor', 'TerrainBlendingProcessor'):
    s = replace_exact(s, f'() -> {name}.CODEC;', f'() -> {name}.CODEC.codec();', 1,
                      f'ModStructureProcessorTypes {name}')
p.write_text(s)

# Target Spell Engine NBT-backed component setter + one-arg Identifier constructor.
p = owner('net/more_rpg_classes/util/loot/SpecificSpellScrollPoolLootFunction.java')
s = p.read_text()
s = replace_exact(s, 'Identifier::of', 'value -> new Identifier(value)', 2,
                  'SpecificSpellScrollPoolLootFunction Identifier parser')
s = replace_exact(s,
    'stack.set(SpellDataComponents.SPELL_CONTAINER, newContainer);',
    'SpellDataComponents.set(stack, SpellDataComponents.SPELL_CONTAINER, newContainer);',
    1, 'SpecificSpellScrollPoolLootFunction component persistence')
p.write_text(s)

p = owner('net/more_rpg_classes/util/loot/BindSpellFromPoolsLootFunction.java')
s = p.read_text()
s = replace_exact(s, 'Identifier::of', 'value -> new Identifier(value)', 1,
                  'BindSpellFromPoolsLootFunction Identifier parser')
s = replace_exact(s,
    'stack.set(SpellDataComponents.SPELL_CONTAINER, container);',
    'SpellDataComponents.set(stack, SpellDataComponents.SPELL_CONTAINER, container);',
    1, 'BindSpellFromPoolsLootFunction component persistence')
p.write_text(s)

# Target registry getEntry(value) is a RegistryEntry, not Optional; unwrap its value for 1.20.1 EnchantmentHelper.
p = owner('net/more_rpg_classes/custom/spell_impacts/KnockbackRangeScaledSpellImpact.java')
s = p.read_text()
power_block = '''            if (power_enchantment.isPresent()) {\n                power_enchantment_value = EnchantmentHelper.getLevel(power_enchantment.get(), attacker.getMainHandStack());\n            }'''
knock_block = '''            if (knockback_enchantment.isPresent()) {\n                knockback_enchantment_value = EnchantmentHelper.getLevel(knockback_enchantment.get(), attacker.getMainHandStack());\n            }'''
s = replace_exact(s, power_block,
    '            power_enchantment_value = EnchantmentHelper.getLevel(power_enchantment.value(), attacker.getMainHandStack());',
    1, 'KnockbackRangeScaledSpellImpact punch holder')
s = replace_exact(s, knock_block,
    '            knockback_enchantment_value = EnchantmentHelper.getLevel(knockback_enchantment.value(), attacker.getMainHandStack());',
    1, 'KnockbackRangeScaledSpellImpact knockback holder')
p.write_text(s)

# Fail closed on the modern APIs this wave owns.
for rel in (
    'net/more_rpg_classes/client/render/MobBeamTracker.java',
    'net/more_rpg_classes/util/CustomMethods.java',
    'net/more_rpg_classes/entity/goal/MobSpellCastGoal.java',
):
    text = owner(rel).read_text()
    if '.getEntry(activeSpellId)' in text or 'registry.getEntry(id)' in text or 'registry.getEntry(spellId)' in text:
        raise SystemExit(f'modern Identifier registry lookup survived Wave 5b: {rel}')
status_forbidden = {
    'net/more_rpg_classes/mixin/PersistentProjectileEntityMixin.java': tuple(effects),
    'net/more_rpg_classes/mixin/LivingEntityMixin.java': tuple(living_effect_counts),
    'net/more_rpg_classes/mixin/DrawHeartsMixin.java': ('MRPGCEffects.FATAL_POISON.entry',),
    'net/more_rpg_classes/mixin/HudMessagesMixin.java': (
        'MRPGCEffects.FROZEN_SOLID.entry',
        'MRPGCEffects.IGNITED.entry',
        'MRPGCEffects.FEAR.entry',
        'MRPGCEffects.STAGGER.entry',
    ),
}
for rel, forbidden_refs in status_forbidden.items():
    text = owner(rel).read_text()
    survived = [ref for ref in forbidden_refs if ref in text]
    if survived:
        raise SystemExit(f'status holder survived Wave 5b: {rel}: {survived}')

print('[More RPG 2.7.2] TARGET_1201_API_WAVE5B_PASS registry_lookups=10 attribute_consumers=2 status_consumers=20')
print('[More RPG 2.7.2] LOOT_COMPONENT_AND_IDENTIFIER_1201_PASS component_sets=2 identifier_refs=3')
print('[More RPG 2.7.2] WORLDGEN_ENTITY_ENCHANTMENT_1201_PASS processors=3 entity_dimensions=1 enchantment_holders=2')
print('[More RPG 2.7.2] MODERN_GAMEPLAY_LOGIC_PRESERVED wave5b=representation_only')
