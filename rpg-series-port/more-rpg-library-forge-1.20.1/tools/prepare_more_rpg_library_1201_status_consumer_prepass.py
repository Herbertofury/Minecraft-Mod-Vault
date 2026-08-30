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
# Pin the exact modern 2.7.2 owners and require one seam in each. The /effect package remains owned by
# Wave 2 proper, so this prepass cannot broaden into an unchecked global substitution.
owners = (
    'net/more_rpg_classes/mixin/LivingEntityStealth.java',
    'net/more_rpg_classes/mixin/ControlOwnerMixin.java',
    'net/more_rpg_classes/mixin/LivingEntityMixin.java',
    'net/more_rpg_classes/mixin/TrackTargetGoalStealth.java',
    'net/more_rpg_classes/custom/spell_impacts/SpellthiefImpact.java',
)
needle = '.getEffectType().value()'
replacement = '.getEffectType()'
rewritten = 0
for rel in owners:
    path = java / rel
    if not path.is_file():
        raise SystemExit(f'missing pinned #331 status consumer: {rel}')
    text = path.read_text(encoding='utf-8')
    count = text.count(needle)
    if count != 1:
        raise SystemExit(f'#331 status consumer seam drifted: {rel} expected=1 found={count}')
    path.write_text(text.replace(needle, replacement, 1), encoding='utf-8')
    rewritten += 1

if rewritten != 5:
    raise SystemExit(f'#331 status consumer repair incomplete: expected=5 rewritten={rewritten}')
print('[More RPG 2.7.2] STATUS_EFFECT_NON_EFFECT_CONSUMERS_1201_PASS owners=5')
