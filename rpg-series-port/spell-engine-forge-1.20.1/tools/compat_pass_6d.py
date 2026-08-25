#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 3:
    raise SystemExit('usage: compat_pass_6d.py <generated-port-root> <spell-engine-1.20.1-baseline>')
root = Path(sys.argv[1]).resolve()
aw = root / 'common/src/main/resources/spell_engine.accesswidener'
s = aw.read_text()

# Spell Engine 1.10.2 targets the post-1.20.1 LootTable storage shape (Optional + Lists).
# Minecraft 1.20.1 uses nullable Identifier + arrays. Retarget only those descriptors so the
# later Forge loot-table bridge can reconstruct an existing table without losing pools/functions.
pairs = {
    'accessible    field    net/minecraft/loot/LootTable    randomSequenceId    Ljava/util/Optional;':
        'accessible    field    net/minecraft/loot/LootTable    randomSequenceId    Lnet/minecraft/util/Identifier;',
    'accessible    field    net/minecraft/loot/LootTable    pools    Ljava/util/List;':
        'accessible    field    net/minecraft/loot/LootTable    pools    [Lnet/minecraft/loot/LootPool;',
    'accessible    field    net/minecraft/loot/LootTable    functions    Ljava/util/List;':
        'accessible    field    net/minecraft/loot/LootTable    functions    [Lnet/minecraft/loot/function/LootFunction;',
    'accessible    method    net/minecraft/loot/LootTable    <init>    (Lnet/minecraft/loot/context/LootContextType;Ljava/util/Optional;Ljava/util/List;Ljava/util/List;)V':
        'accessible    method    net/minecraft/loot/LootTable    <init>    (Lnet/minecraft/loot/context/LootContextType;Lnet/minecraft/util/Identifier;[Lnet/minecraft/loot/LootPool;[Lnet/minecraft/loot/function/LootFunction;)V',
}

for old, new in pairs.items():
    if old not in s:
        raise SystemExit(f'expected modern LootTable access-widener descriptor missing: {old}')
    s = s.replace(old, new)
aw.write_text(s)

final = aw.read_text()
for old in pairs:
    if old in final:
        raise SystemExit(f'LootTable access-widener retarget incomplete: {old}')
for new in pairs.values():
    if new not in final:
        raise SystemExit(f'expected 1.20.1 LootTable access-widener descriptor missing: {new}')

print('Spell Engine compatibility pass 6d applied: retargeted LootTable access widening to 1.20.1 Identifier/array storage')
