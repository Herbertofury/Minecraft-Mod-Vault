#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 2:
    raise SystemExit('usage: prepare_more_rpg_library_forge_construct_deferral.py <generated-port-root>')
root = Path(sys.argv[1]).resolve()
path = root / 'forge/src/main/java/net/more_rpg_classes/forge/ForgeMod.java'
if not path.is_file():
    raise SystemExit(f'missing ForgeMod runtime owner: {path}')
s = path.read_text(encoding='utf-8')

# #352 proved a circular class-initialization deadlock during Forge's parallel mod construction:
# More RPG -> MoreSpellSchools -> SpellSchools while SpellSchools' injected static tail waits on
# MoreSpellSchools. Preserve registry listeners on the mod bus, but execute common init from
# FMLConstructModEvent deferred work after parallel constructors converge on Forge's serial queue.
#
# Registry wave6 legitimately introduces FMLJavaModLoadingContext before this stage. Reuse that
# exact bus instead of treating the shared lifecycle import as evidence that deferral already ran.
context_import = 'import net.minecraftforge.fml.javafmlmod.FMLJavaModLoadingContext;\n'
construct_import = 'import net.minecraftforge.fml.event.lifecycle.FMLConstructModEvent;\n'
bus_import = 'import net.minecraftforge.eventbus.api.IEventBus;\n'
if s.count(context_import) != 1:
    raise SystemExit(f'Forge construct-deferral mod-bus import seam drifted: found={s.count(context_import)}')
if 'FMLConstructModEvent' in s or 'IEventBus' in s:
    raise SystemExit('Forge construct deferral unexpectedly pre-exists')
if s.count('ForgeMod::register') != 1:
    raise SystemExit(f'Forge registry-listener seam drifted before deferral: found={s.count("ForgeMod::register")}')
if s.count('event.register(RegistryKeys.') < 7:
    raise SystemExit('Forge registry wave is missing before construct deferral')
s = s.replace(context_import, context_import + construct_import + bus_import, 1)

old = ('        FMLJavaModLoadingContext.get().getModEventBus().addListener(ForgeMod::register);\n'
       '        MRPGCMod.init();\n')
new = ('        IEventBus modBus = FMLJavaModLoadingContext.get().getModEventBus();\n'
       '        modBus.addListener(ForgeMod::register);\n'
       '        modBus.addListener(this::construct);\n')
if s.count(old) != 1:
    raise SystemExit(f'Forge registry-bus/direct-init composition seam drifted: found={s.count(old)}')
s = s.replace(old, new, 1)

class_end = '\n}\n'
if not s.endswith(class_end):
    raise SystemExit('ForgeMod class terminator drifted')
method = '''\n    private void construct(FMLConstructModEvent event) {\n        event.enqueueWork(() -> MRPGCMod.init());\n    }\n'''
s = s[:-len(class_end)] + method + class_end

contracts = {
    'IEventBus modBus = FMLJavaModLoadingContext.get().getModEventBus();': 1,
    'modBus.addListener(ForgeMod::register);': 1,
    'modBus.addListener(this::construct);': 1,
    'event.enqueueWork(() -> MRPGCMod.init());': 1,
    'FMLConstructModEvent': 2,
    'IEventBus': 2,
}
for needle, expected in contracts.items():
    actual = s.count(needle)
    if actual != expected:
        raise SystemExit(f'Forge construct-deferral contract failed for {needle!r}: expected={expected} found={actual}')
if s.count('        MRPGCMod.init();\n') != 0:
    raise SystemExit('Forge direct MRPG init survived construction deferral')
if s.count('MRPGCMod.init()') != 1:
    raise SystemExit('Forge MRPG init invocation cardinality changed')
if '[More RPG Runtime Trace]' in s:
    raise SystemExit('diagnostic runtime trace leaked into clean lifecycle repair')

path.write_text(s, encoding='utf-8')
print('[More RPG 2.7.2] FORGE_CONSTRUCT_DEFERRED_INIT_1201_PASS '
      'source=run-352-deadlock timing=construct_event_enqueue_work registry_bus=reused')
