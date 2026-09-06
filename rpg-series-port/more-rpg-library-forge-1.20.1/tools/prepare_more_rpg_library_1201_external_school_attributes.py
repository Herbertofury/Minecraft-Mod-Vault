#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 2:
    raise SystemExit('usage: prepare_more_rpg_library_1201_external_school_attributes.py <prepared-root>')
root = Path(sys.argv[1]).resolve()
path = root / 'common/src/main/java/net/more_rpg_classes/custom/MoreSpellSchools.java'
if not path.is_file():
    raise SystemExit(f'missing MoreSpellSchools source: {path}')
s = path.read_text(encoding='utf-8')

# Modern Spell Power distinguishes owned attributes from RegistryEntry-backed external attributes by
# constructor shape. The 1.20.1 SpellSchool API uses a raw EntityAttribute constructor whose default is
# Manage.INTERNAL. During the 1.21 -> 1.20.1 representation bridge, FROST_RANGED, FIRE_RANGED and
# RAGE_MELEE keep their existing Ranged Weapon API / vanilla attributes, so their ownership must be
# restored explicitly or Forge will try to register the same external Attribute object under school IDs.
for school in ('FROST_RANGED', 'FIRE_RANGED', 'RAGE_MELEE'):
    if f'{school}.attributeManagement' in s:
        raise SystemExit(f'external school attribute ownership unexpectedly pre-exists: {school}')

anchor = '        FROST_RANGED.addSource(SpellSchool.Trait.POWER'
if s.count(anchor) != 1:
    raise SystemExit(f'MoreSpellSchools first source seam drifted: found={s.count(anchor)}')
ownership = '''        FROST_RANGED.attributeManagement = SpellSchool.Manage.EXTERNAL;\n        FIRE_RANGED.attributeManagement = SpellSchool.Manage.EXTERNAL;\n        RAGE_MELEE.attributeManagement = SpellSchool.Manage.EXTERNAL;\n\n'''
s = s.replace(anchor, ownership + anchor, 1)

for school in ('FROST_RANGED', 'FIRE_RANGED', 'RAGE_MELEE'):
    expected = f'{school}.attributeManagement = SpellSchool.Manage.EXTERNAL;'
    if s.count(expected) != 1:
        raise SystemExit(f'external school attribute ownership missing or duplicated: {school}')
if s.count('attributeManagement = SpellSchool.Manage.EXTERNAL;') != 3:
    raise SystemExit('external school attribute ownership cardinality changed')
path.write_text(s, encoding='utf-8')
print('[More RPG 2.7.2] EXTERNAL_SCHOOL_ATTRIBUTE_OWNERSHIP_1201_PASS schools=frost_ranged,fire_ranged,rage_melee source=run-353-registry-boundary')
