#!/usr/bin/env python3
from pathlib import Path
import subprocess
import sys

if len(sys.argv) != 2:
    raise SystemExit('usage: prepare_more_rpg_library_runtime_trace.py <generated-port-root>')
root = Path(sys.argv[1]).resolve()
common = root / 'common/src/main/java/net/more_rpg_classes/MRPGCMod.java'
forge = root / 'forge/src/main/java/net/more_rpg_classes/forge/ForgeMod.java'
for p in (common, forge):
    if not p.is_file():
        raise SystemExit(f'missing runtime trace owner: {p}')

# #351 proved a live-process startup stall after Forge initialization with no exception and no Done marker.
# Instrument only ownership boundaries; do not change registration/config semantics. System.err is used so
# phase evidence survives even if the normal logging pipeline is the blocked subsystem.
s = common.read_text(encoding='utf-8')
phase_seams = [
    ('\tpublic static ConfigManager<TweaksConfig> tweaksConfig =',
     '\tstatic { System.err.println("[More RPG Runtime Trace] CLINIT_EFFECTS_READY"); }\n\tpublic static ConfigManager<TweaksConfig> tweaksConfig ='),
    ('\tpublic static ConfigManager<WeaknessConfig> weaknessConfig =',
     '\tstatic { System.err.println("[More RPG Runtime Trace] CLINIT_TWEAKS_READY"); }\n\tpublic static ConfigManager<WeaknessConfig> weaknessConfig ='),
    ('\tpublic static final ConfigManager<LootConfig> lootConfig =',
     '\tstatic { System.err.println("[More RPG Runtime Trace] CLINIT_WEAKNESS_READY"); }\n\tpublic static final ConfigManager<LootConfig> lootConfig ='),
    ('\n\n\tpublic static void init() {',
     '\n\tstatic { System.err.println("[More RPG Runtime Trace] CLINIT_LOOT_READY"); }\n\n\tpublic static void init() {\n\t\tSystem.err.println("[More RPG Runtime Trace] INIT_ENTER");'),
]
for old, new in phase_seams:
    if s.count(old) != 1:
        raise SystemExit(f'#351 runtime trace class-init seam drifted: {old!r} found={s.count(old)}')
    s = s.replace(old, new, 1)

calls = [
    'effectsConfig.refresh();',
    'tweaksConfig.refresh();',
    'weaknessConfig.refresh();',
    'lootConfig.refresh();',
    'CustomSpellImpacts.registerCustomImpacts();',
    'MrpgEntityRelationMatcher.register();',
    'MoreSpellSchools.initialize();',
    'CustomSpellEntityPredicate.registerCustomPredicates();',
    'CriticalStrikeCompat.init();',
]
for i, call in enumerate(calls, 1):
    if s.count(call) != 1:
        raise SystemExit(f'#351 runtime trace init seam drifted: {call} found={s.count(call)}')
    label = call.split('(')[0].replace('.', '_').replace(';', '')
    replacement = (f'System.err.println("[More RPG Runtime Trace] INIT_{i:02d}_{label}_BEGIN");\n\t\t\t{call}\n'
                   f'\t\t\tSystem.err.println("[More RPG Runtime Trace] INIT_{i:02d}_{label}_END");')
    s = s.replace(call, replacement, 1)
common.write_text(s, encoding='utf-8')

f = forge.read_text(encoding='utf-8')
needle = '        MRPGCMod.init();\n'
if f.count(needle) != 1:
    raise SystemExit(f'#351 ForgeMod init seam drifted: found={f.count(needle)}')
f = f.replace(needle,
    '        System.err.println("[More RPG Runtime Trace] FORGE_CONSTRUCTOR_BEFORE_MRPG_CLASS_INIT");\n'
    '        MRPGCMod.init();\n'
    '        System.err.println("[More RPG Runtime Trace] FORGE_CONSTRUCTOR_AFTER_MRPG_INIT");\n', 1)
forge.write_text(f, encoding='utf-8')
print('[More RPG 2.7.2] RUNTIME_STALL_PHASE_TRACE_1201_PASS source=run-351 boundaries=class_init+9_init_calls+forge_constructor')

# Candidate branches may place the independently reviewed Forge construction deferral beside this trace
# stage. Keep the validation branch behavior unchanged when that optional file is absent.
deferral = Path(__file__).with_name('prepare_more_rpg_library_forge_construct_deferral.py')
if deferral.is_file():
    subprocess.run([sys.executable, str(deferral), str(root)], check=True)
