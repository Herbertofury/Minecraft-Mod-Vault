#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 2:
    raise SystemExit('usage: prepare_more_rpg_library_1201_registry_wave4.py <prepared-root>')
root = Path(sys.argv[1]).resolve()
path = root / 'common/src/main/java/net/more_rpg_classes/entity/attribute/MRPGCEntityAttributes.java'
if not path.is_file():
    raise SystemExit(f'missing More RPG attribute registry source: {path}')
s = path.read_text()

holder = 'RegistryEntry<EntityAttribute>'
expected_holders = 20  # 19 current 2.7.2 fields + register(...) return type.
if s.count(holder) != expected_holders:
    raise SystemExit(f'More RPG attribute holder seam drifted: expected={expected_holders} found={s.count(holder)}')
if s.count('import net.minecraft.registry.entry.RegistryEntry;') != 1:
    raise SystemExit('More RPG RegistryEntry attribute import seam drifted')
register_ref = 'return Registry.registerReference(Registries.ATTRIBUTE, new Identifier(MOD_ID, name), attribute);'
if s.count(register_ref) != 1:
    raise SystemExit(f'More RPG attribute registerReference seam drifted: found={s.count(register_ref)}')

s = s.replace('import net.minecraft.registry.entry.RegistryEntry;\n', '', 1)
s = s.replace(holder, 'EntityAttribute')
s = s.replace(register_ref, 'return Registry.register(Registries.ATTRIBUTE, new Identifier(MOD_ID, name), attribute);', 1)

if 'RegistryEntry<EntityAttribute>' in s or 'registerReference(Registries.ATTRIBUTE' in s:
    raise SystemExit('modern 1.21 entity-attribute holder API survived Wave 4')
if s.count('public static final EntityAttribute') != 19:
    raise SystemExit('Wave 4 changed More RPG 2.7.2 attribute cardinality')
if s.count('return Registry.register(Registries.ATTRIBUTE, new Identifier(MOD_ID, name), attribute);') != 1:
    raise SystemExit('target 1.20.1 raw attribute registration missing or duplicated')
path.write_text(s)

print('[More RPG 2.7.2] ATTRIBUTE_REGISTRY_1201_WAVE4_PASS attributes=19 holder=raw registration=Registry.register')
print('[More RPG 2.7.2] MODERN_ATTRIBUTE_SET_PRESERVED removed_legacy_attributes=not_resurrected')
