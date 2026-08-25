#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 3:
    raise SystemExit('usage: compat_pass_6c.py <generated-port-root> <spell-engine-1.20.1-baseline>')
root = Path(sys.argv[1]).resolve()
common = root / 'common'
aw = common / 'src/main/resources/spell_engine.accesswidener'
s = aw.read_text()

poison = 'accessible    class    net/minecraft/entity/effect/PoisonStatusEffect\n'
tracker_entries = (
    'accessible    field    net/minecraft/server/world/ServerChunkLoadingManager    entityTrackers    Lit/unimi/dsi/fastutil/ints/Int2ObjectMap;\n',
    'accessible    class    net/minecraft/server/world/ServerChunkLoadingManager$EntityTracker\n',
    'accessible    field    net/minecraft/server/world/ServerChunkLoadingManager$EntityTracker    listeners    Ljava/util/Set;\n',
)

if poison not in s:
    raise SystemExit('expected obsolete PoisonStatusEffect access-widener entry missing')
for entry in tracker_entries:
    if entry not in s:
        raise SystemExit(f'expected obsolete tracker access-widener entry missing: {entry.strip()}')

# PoisonStatusEffect is no longer referenced/subclassed by the 1.10.2 common source.
# The old vanilla entity-tracker widening is also no longer live after pass 5c moved tracking behind
# Platform.Util.tracking(Entity), where Forge can provide exact recipient semantics without private
# ServerChunkLoadingManager fields. Remove only these proven-dead widenings; keep every other AW entry.
source_hits = []
for p in (common / 'src/main/java').rglob('*.java'):
    text = p.read_text()
    for needle in ('PoisonStatusEffect', 'ServerChunkLoadingManager', 'entityTrackers', 'tracker.listeners'):
        if needle in text:
            source_hits.append((str(p.relative_to(common)), needle))
if source_hits:
    raise SystemExit(f'refusing to remove live access widenings; source hits: {source_hits[:20]}')

s = s.replace(poison, '')
for entry in tracker_entries:
    s = s.replace(entry, '')
aw.write_text(s)

final = aw.read_text()
for needle in ('PoisonStatusEffect', 'ServerChunkLoadingManager', 'entityTrackers'):
    if needle in final:
        raise SystemExit(f'access-widener cleanup incomplete: {needle}')
print('Spell Engine compatibility pass 6c applied: removed proven-unused poison + vanilla tracker widenings')
