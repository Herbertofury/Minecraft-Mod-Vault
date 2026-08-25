#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 3:
    raise SystemExit('usage: compat_pass_6c.py <generated-port-root> <spell-engine-1.20.1-baseline>')
root = Path(sys.argv[1]).resolve()
common = root / 'common'
aw = common / 'src/main/resources/spell_engine.accesswidener'
needle = 'accessible    class    net/minecraft/entity/effect/PoisonStatusEffect\n'
s = aw.read_text()
if needle not in s:
    raise SystemExit('expected obsolete PoisonStatusEffect access-widener entry missing')

# The 1.10.2 source no longer references/subclasses PoisonStatusEffect. This target-only widening is a
# stale carry-over from older Spell Engine code and Forge 47 cannot validate that Yarn class owner.
hits = []
for p in (common / 'src/main/java').rglob('*.java'):
    if 'PoisonStatusEffect' in p.read_text():
        hits.append(str(p.relative_to(common)))
if hits:
    raise SystemExit(f'refusing to remove live PoisonStatusEffect widening; source hits: {hits}')

aw.write_text(s.replace(needle, ''))
assert 'PoisonStatusEffect' not in aw.read_text()
print('Spell Engine compatibility pass 6c applied: removed proven-unused PoisonStatusEffect widening')
