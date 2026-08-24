#!/usr/bin/env python3
from pathlib import Path
import json, shutil, sys

if len(sys.argv) != 4:
    raise SystemExit('usage: prepare_upstream_source.py <upstream-1.20.1> <upstream-1.21.1> <common-dir>')
old = Path(sys.argv[1])
new = Path(sys.argv[2])
common = Path(sys.argv[3])
gen_java = common / 'src/generatedUpstream/java'
gen_res = common / 'src/generatedUpstream/resources'
for p in (gen_java, gen_res):
    shutil.rmtree(p, ignore_errors=True)
    p.mkdir(parents=True, exist_ok=True)

shutil.copytree(old / 'src/main/java', gen_java, dirs_exist_ok=True)
shutil.copytree(old / 'src/main/resources', gen_res, dirs_exist_ok=True)

for rel in [
    'net/spell_power/SpellPowerMod.java',
    'net/spell_power/mixin/EntityAttributesMixin.java',
    'net/spell_power/mixin/StatusEffectsMixin.java',
    'net/spell_power/mixin/PlayerEntityMixin.java',
    'net/spell_power/mixin/LivingEntityMixin.java',
    'net/spell_power/api/SpellPower.java',
    'net/spell_power/api/SpellPowerMechanics.java',
    'net/spell_power/api/SpellResistance.java',
    'net/spell_power/api/SpellSchools.java',
    'net/spell_power/config/AttributesConfig.java',
]:
    (gen_java / rel).unlink(missing_ok=True)
(gen_res / 'fabric.mod.json').unlink(missing_ok=True)
(gen_res / 'spell_power.mixins.json').unlink(missing_ok=True)

new_assets = new / 'common/src/main/resources/assets/spell_power'
if new_assets.exists():
    for sub in ('lang', 'textures'):
        src = new_assets / sub
        if src.exists():
            shutil.copytree(src, gen_res / 'assets/spell_power' / sub, dirs_exist_ok=True)

for rel in [
    'common/src/main/resources/data/c/tags/damage_type/is_magic.json',
    'common/src/main/resources/data/minecraft/tags/damage_type/panic_causes.json',
    'common/src/main/resources/data/spell_power/tags/damage_type/resistable.json',
]:
    src = new / rel
    if src.exists():
        dst = gen_res / Path(rel).relative_to('common/src/main/resources')
        dst.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(src, dst)

count = 0
for path in gen_res.rglob('*.json'):
    with path.open('r', encoding='utf-8') as fh:
        json.load(fh)
    count += 1
print(f'Prepared upstream Spell Power source; validated {count} JSON resources')
