#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 2:
    raise SystemExit('usage: prepare_more_rpg_library_1201_status_consumers_wave5a.py <prepared-root>')
root = Path(sys.argv[1]).resolve()
java = root / 'common/src/main/java'
if not java.is_dir():
    raise SystemExit(f'missing common Java root: {java}')

custom = java / 'net/more_rpg_classes/util/CustomMethods.java'
low = java / 'net/more_rpg_classes/entity/goal/LowHealthFleeGoal.java'
for p in (custom, low):
    if not p.is_file():
        raise SystemExit(f'missing Wave 5a owner: {p}')

s = custom.read_text()
checks = {
    'RegistryEntry import': ('import net.minecraft.registry.entry.RegistryEntry;\n', 1),
    'raw list holder': ('new java.util.ArrayList<RegistryEntry<StatusEffect>>()', 1),
    'holder value beneficial': ('effectEntry.value().isBeneficial()', 1),
    'future trial omen': ('StatusEffects.TRIAL_OMEN', 1),
    'status holder parameter': ('RegistryEntry<StatusEffect> statusEffect', 2),
    'ranged damage holder': ('EntityAttributes_RangedWeapon.DAMAGE.entry', 2),
}
for label, (needle, expected) in checks.items():
    found = s.count(needle)
    if found != expected:
        raise SystemExit(f'CustomMethods {label} seam drifted: expected={expected} found={found}')

s = s.replace('import net.minecraft.registry.entry.RegistryEntry;\n', '', 1)
s = s.replace('new java.util.ArrayList<RegistryEntry<StatusEffect>>()', 'new java.util.ArrayList<StatusEffect>()', 1)
s = s.replace('effectEntry.value().isBeneficial()', 'effectEntry.isBeneficial()', 1)
s = s.replace('StatusEffects.TRIAL_OMEN', 'StatusEffects.BAD_OMEN', 1)
s = s.replace('RegistryEntry<StatusEffect> statusEffect', 'StatusEffect statusEffect')
s = s.replace('EntityAttributes_RangedWeapon.DAMAGE.entry', 'EntityAttributes_RangedWeapon.DAMAGE.attribute')
for forbidden in ('RegistryEntry<StatusEffect>', 'effectEntry.value().isBeneficial()', 'StatusEffects.TRIAL_OMEN', 'EntityAttributes_RangedWeapon.DAMAGE.entry'):
    if forbidden in s:
        raise SystemExit(f'CustomMethods Wave 5a API survived: {forbidden}')
custom.write_text(s)

s = low.read_text()
modern = 'Registries.STATUS_EFFECT.getEntry(effectId).orElse(null)'
if s.count(modern) != 2:
    raise SystemExit(f'LowHealthFleeGoal status lookup seam drifted: expected=2 found={s.count(modern)}')
s = s.replace(modern, 'Registries.STATUS_EFFECT.get(effectId)')
if 'Registries.STATUS_EFFECT.getEntry(effectId)' in s:
    raise SystemExit('LowHealthFleeGoal modern status lookup survived Wave 5a')
if s.count('Registries.STATUS_EFFECT.get(effectId)') != 2:
    raise SystemExit('LowHealthFleeGoal target raw status lookup missing or duplicated')
low.write_text(s)

print('[More RPG 2.7.2] STATUS_CONSUMERS_1201_WAVE5A_PASS custom_methods=1 low_health_lookups=2')
print('[More RPG 2.7.2] TARGET_ERA_OMEN_SEMANTICS_PRESERVED trial_omen=bad_omen')
