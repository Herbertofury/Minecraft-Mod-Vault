#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
BASE="$ROOT/rpg-series-port/ci/run-spell-engine.sh"
GRAD="$ROOT/rpg-series-port/ci/run-spell-engine-graduation.sh"
PASS6C="$ROOT/rpg-series-port/spell-engine-forge-1.20.1/tools/compat_pass_6c.py"
PASS6F="$ROOT/rpg-series-port/spell-engine-forge-1.20.1/tools/compat_pass_6f.py"
PASS6I="$ROOT/rpg-series-port/spell-engine-forge-1.20.1/tools/compat_pass_6i.py"
ENV_FILE="$ROOT/rpg-series-port/spell-engine-forge-1.20.1/SPELL_ENGINE_GRADUATION.env"

test -f "$BASE"
test -f "$GRAD"
test -f "$PASS6C"
test -f "$PASS6F"
test -f "$PASS6I"
test -f "$ENV_FILE"
source "$ENV_FILE"

[[ "$SPELL_ENGINE_TARGET_VERSION" = '1.10.4' ]]
[[ "$SPELL_ENGINE_1104_TARGET_SHA" = '8843cea8974afffc7ec42c096cac33327a3af3d8' ]]

# Uplift only target authority/path/release labeling in this CI workspace and repair the
# historical Spell Power dependency-build invocation. The historical source-package seam
# intentionally remains untouched here because the audited graduation harness owns its
# deterministic replacement. This ordering prevents the two fail-closed wrappers from
# competing over the same literal.
python3 - "$BASE" <<'PY'
from pathlib import Path
import sys
p=Path(sys.argv[1]); s=p.read_text()

repls=[
    ('TARGET_SHA=bc02f7a49da950503010020da491f6bdc5871df7',
     'TARGET_SHA=8843cea8974afffc7ec42c096cac33327a3af3d8', 1),
    ('spell-engine-1102', 'spell-engine-1104', 2),
    ('spell_engine-forge-1.10.2+1.20.1.jar',
     'spell_engine-forge-1.10.4+1.20.1.jar', 1),
    ('gradle --no-daemon --stacktrace -p "$SPELL_POWER" :forge:build',
     'gradle --no-daemon --stacktrace -p "$SPELL_POWER" -Ptiny_config_common_jar="$TINY_CONFIG_COMMON_JAR" -Ptiny_config_forge_jar="$TINY_CONFIG_FORGE_JAR" :forge:build', 1),
]
for old,new,expected in repls:
    actual=s.count(old)
    if actual != expected:
        raise SystemExit(f'[Spell Engine 1.10.4] expected {expected} occurrences of {old!r}, found {actual}')
    s=s.replace(old,new)

source_name='spell-engine-1.10.2-forge-1.20.1-source-ci.zip'
if s.count(source_name) != 3:
    raise SystemExit(f'[Spell Engine 1.10.4] graduation-owned source seam drifted: expected 3, found {s.count(source_name)}')
if 'bc02f7a49da950503010020da491f6bdc5871df7' in s:
    raise SystemExit('[Spell Engine 1.10.4] stale 1.10.2 target SHA remains in active base runner')
if 'spell-engine-1102' in s:
    raise SystemExit('[Spell Engine 1.10.4] stale 1.10.2 target path remains in active base runner')
if 'gradle --no-daemon --stacktrace -p "$SPELL_POWER" :forge:build' in s:
    raise SystemExit('[Spell Engine 1.10.4] stale Spell Power build invocation still omits TinyConfig 3.1 inputs')
p.write_text(s)
PY

# Spell Engine 1.10.4's loot-inspection crash fix widens LootPool.entries. In modern 1.21.1
# that field is a List, while Yarn/MC 1.20.1 exposes the same named field as LootPoolEntry[].
# Preserve the widening and adapt only its JVM descriptor; deleting it would silently regress the
# exact 1.10.4 behavior this graduation is meant to carry back.
python3 - "$PASS6C" <<'PY'
from pathlib import Path
import sys
p=Path(sys.argv[1]); s=p.read_text()
anchor="raw = aw.read_text()\nlines = raw.splitlines()\n"
insert="""raw = aw.read_text()\nmodern_loot_pool_entries = 'accessible    field    net/minecraft/loot/LootPool    entries    Ljava/util/List;'\ntarget_loot_pool_entries = 'accessible    field    net/minecraft/loot/LootPool    entries    [Lnet/minecraft/loot/entry/LootPoolEntry;'\nif raw.count(modern_loot_pool_entries) != 1:\n    raise SystemExit(f'expected exactly one Spell Engine 1.10.4 LootPool.entries widening, found {raw.count(modern_loot_pool_entries)}')\nraw = raw.replace(modern_loot_pool_entries, target_loot_pool_entries, 1)\naw.write_text(raw)\nlines = raw.splitlines()\n"""
if s.count(anchor) != 1:
    raise SystemExit(f'[Spell Engine 1.10.4] pass-6c AW read seam drifted: found {s.count(anchor)}')
s=s.replace(anchor,insert,1)
old_tail="if remaining:\n    raise SystemExit(f'access-widener cleanup incomplete: {remaining}')\nprint('Spell Engine compatibility pass 6c applied: removed stale poison/tracker + NeoForge-only loot widenings')\n"
new_tail="""if remaining:\n    raise SystemExit(f'access-widener cleanup incomplete: {remaining}')\nfinal_aw = aw.read_text()\nif modern_loot_pool_entries in final_aw:\n    raise SystemExit('Spell Engine 1.10.4 LootPool.entries List descriptor survived 1.20.1 adaptation')\nif final_aw.count(target_loot_pool_entries) != 1:\n    raise SystemExit('Spell Engine 1.10.4 LootPool.entries array widening missing or duplicated')\nprint('Spell Engine compatibility pass 6c applied: removed stale poison/tracker + NeoForge-only loot widenings; adapted LootPool.entries widening to 1.20.1 array descriptor')\n"""
if s.count(old_tail) != 1:
    raise SystemExit('[Spell Engine 1.10.4] pass-6c tail seam drifted')
s=s.replace(old_tail,new_tail,1)
p.write_text(s)
PY

# Spell Engine 1.10.4 widened its loot builder calls: buildPool now receives the registry lookup
# for entry-list variants and EnchantWithLevelsLootFunction.builder also receives that lookup.
# Forge/Yarn 1.20.1 exposes neither registry-bearing builder contract here. Teach the existing
# pass-6f adapter the generalized 1.10.4 shapes before it runs, and keep fail-closed assertions.
python3 - "$PASS6F" <<'PY'
from pathlib import Path
import sys
p=Path(sys.argv[1]); s=p.read_text()
repls=[
    ("s = s.replace('buildPool(registries, pool,', 'buildPool(pool,')",
     "s = s.replace('buildPool(registries, ', 'buildPool(')", 1),
    ("s = s.replace('private static LootPool buildPool(RegistryWrapper.WrapperLookup registries, LootConfig.Pool pool,', 'private static LootPool buildPool(LootConfig.Pool pool,')",
     "s = s.replace('private static LootPool buildPool(RegistryWrapper.WrapperLookup registries, ', 'private static LootPool buildPool(')", 1),
]
for old,new,expected in repls:
    actual=s.count(old)
    if actual != expected:
        raise SystemExit(f'[Spell Engine 1.10.4] pass-6f loot seam drift: expected {expected} occurrence of {old!r}, found {actual}')
    s=s.replace(old,new)
needle="s = s.replace('private static LootPool buildPool(RegistryWrapper.WrapperLookup registries, ', 'private static LootPool buildPool(')\n"
insert="s = s.replace('EnchantWithLevelsLootFunction.builder(registries, ', 'EnchantWithLevelsLootFunction.builder(')\n"
if insert not in s:
    s=s.replace(needle, needle + insert)
old_guard="if 'RegistryWrapper.WrapperLookup registries' in s or 'buildPool(registries' in s or 'configureFallback(registries' in s:\n"
new_guard="if 'RegistryWrapper.WrapperLookup registries' in s or 'buildPool(registries' in s or 'configureFallback(registries' in s or 'builder(registries' in s:\n"
if s.count(old_guard) != 1:
    raise SystemExit('[Spell Engine 1.10.4] pass-6f registry guard seam drifted')
s=s.replace(old_guard,new_guard)
p.write_text(s)
PY

# Pass 6i's original invariants described 1.10.2's direct pool.entries loops. 1.10.4 intentionally
# broadened tag caching to exact injectors + regex injectors + fallback references through entryLists,
# and buildPool now receives List<Pool.Entry>. Preserve and assert that newer structure rather than
# forcing it back to the historical shape.
python3 - "$PASS6I" <<'PY'
from pathlib import Path
import sys
p=Path(sys.argv[1]); s=p.read_text()
old1="if loot_helper.read_text().count('for (var itemInjectorEntry: pool.entries)') != 1: raise SystemExit('pass6i changed LootConfig tag-cache entries')"
new1="lh_final = loot_helper.read_text()\nfor required in ('lootConfig.injectors.values().forEach(pool -> entryLists.add(pool.entries));', 'lootConfig.regex_injectors.values().forEach(pool -> entryLists.add(pool.entries));', 'for (var entries: entryLists)', 'for (var itemInjectorEntry: entries)'):\n    if required not in lh_final: raise SystemExit(f'pass6i lost Spell Engine 1.10.4 tag-cache parity: {required}')"
old2="if loot_helper.read_text().count('for (var entry: pool.entries)') != 1: raise SystemExit('pass6i changed LootConfig buildPool entries')"
new2="if lh_final.count('for (var entry: entries)') != 1: raise SystemExit('pass6i changed Spell Engine 1.10.4 buildPool entries')"
for old,new,label in ((old1,new1,'tag-cache'),(old2,new2,'buildPool')):
    if s.count(old) != 1:
        raise SystemExit(f'[Spell Engine 1.10.4] pass-6i {label} guard seam drifted: found {s.count(old)}')
    s=s.replace(old,new)
p.write_text(s)
PY

bash -n "$BASE"
python3 -m py_compile "$PASS6C" "$PASS6F" "$PASS6I"
echo '[Spell Engine 1.10.4] TARGET_AUTHORITY_SPELL_POWER_LOOT_API_AND_ACCESS_ADAPTATION_PASS sha=8843cea8974afffc7ec42c096cac33327a3af3d8'
exec bash "$GRAD"
