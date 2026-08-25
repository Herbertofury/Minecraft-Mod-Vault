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
loot_entries = (
    'accessible    field    net/minecraft/loot/LootTable    type    Lnet/minecraft/loot/context/LootContextType;\n',
    'accessible    field    net/minecraft/loot/LootTable    randomSequenceId    Ljava/util/Optional;\n',
    'accessible    field    net/minecraft/loot/LootTable    pools    Ljava/util/List;\n',
    'accessible    field    net/minecraft/loot/LootTable    functions    Ljava/util/List;\n',
    'accessible    method    net/minecraft/loot/LootTable    <init>    (Lnet/minecraft/loot/context/LootContextType;Ljava/util/Optional;Ljava/util/List;Ljava/util/List;)V\n',
)

if poison not in s:
    raise SystemExit('expected obsolete PoisonStatusEffect access-widener entry missing')
for entry in tracker_entries:
    if entry not in s:
        raise SystemExit(f'expected obsolete tracker access-widener entry missing: {entry.strip()}')
for entry in loot_entries:
    if entry not in s:
        raise SystemExit(f'expected NeoForge loot access-widener entry missing: {entry.strip()}')

# PoisonStatusEffect is no longer referenced/subclassed by the 1.10.2 common source.
# The old vanilla entity-tracker widening is no longer live after pass 5c moved tracking behind
# Platform.Util.tracking(Entity), where Forge can provide exact recipient semantics without private
# ServerChunkLoadingManager fields.
# The LootTable private-field/constructor widenings exist for the modern NeoForge implementation,
# which reconstructs a loaded table. Forge 1.20.1 natively exposes LootTableLoadEvent#setTable, so
# the target loader bridge does not need any of those private internals. Remove only these proven
# loader-inapplicable widenings; preserve every remaining AW entry.
source_hits = []
for p in (common / 'src/main/java').rglob('*.java'):
    text = p.read_text()
    for needle in (
        'PoisonStatusEffect', 'ServerChunkLoadingManager', 'entityTrackers', 'tracker.listeners',
        '.randomSequenceId', '.pools', '.functions', 'new LootTable('
    ):
        if needle in text:
            source_hits.append((str(p.relative_to(common)), needle))
if source_hits:
    raise SystemExit(f'refusing to remove live access widenings; source hits: {source_hits[:20]}')

s = s.replace(poison, '')
for entry in tracker_entries + loot_entries:
    s = s.replace(entry, '')
aw.write_text(s)

final = aw.read_text()
for needle in ('PoisonStatusEffect', 'ServerChunkLoadingManager', 'entityTrackers', 'randomSequenceId'):
    if needle in final:
        raise SystemExit(f'access-widener cleanup incomplete: {needle}')
print('Spell Engine compatibility pass 6c applied: removed stale poison/tracker + NeoForge-only loot widenings')
