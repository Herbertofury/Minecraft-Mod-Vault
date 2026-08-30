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
# MoreSpellSchools. Preserve the original early/pre-registration intent but execute common init from
# FMLConstructModEvent deferred work, after parallel constructors converge on Forge's serial work queue.
import_anchor = 'import net.minecraftforge.fml.loading.FMLLoader;\n'
imports = ('import net.minecraftforge.fml.event.lifecycle.FMLConstructModEvent;\n'
           'import net.minecraftforge.fml.javafmlmod.FMLJavaModLoadingContext;\n'
           'import net.minecraftforge.eventbus.api.IEventBus;\n')
if s.count(import_anchor) != 1:
    raise SystemExit(f'Forge construct-deferral import seam drifted: found={s.count(import_anchor)}')
if 'FMLConstructModEvent' in s or 'FMLJavaModLoadingContext' in s:
    raise SystemExit('Forge construct deferral unexpectedly pre-exists')
s = s.replace(import_anchor, import_anchor + imports, 1)

old = '        MRPGCMod.init();\n'
new = ('        IEventBus modBus = FMLJavaModLoadingContext.get().getModEventBus();\n'
       '        modBus.addListener(this::construct);\n')
if s.count(old) != 1:
    raise SystemExit(f'Forge direct MRPG init seam drifted: found={s.count(old)}')
s = s.replace(old, new, 1)

class_end = '\n}\n'
if not s.endswith(class_end):
    raise SystemExit('ForgeMod class terminator drifted')
method = '''\n    private void construct(FMLConstructModEvent event) {\n        event.enqueueWork(() -> MRPGCMod.init());\n    }\n'''
s = s[:-len(class_end)] + method + class_end

if s.count('modBus.addListener(this::construct);') != 1:
    raise SystemExit('Forge construct listener missing or duplicated')
if s.count('event.enqueueWork(() -> MRPGCMod.init());') != 1:
    raise SystemExit('Forge deferred MRPG init work missing or duplicated')
if s.count('MRPGCMod.init();') != 1:
    raise SystemExit('Forge MRPG init call cardinality changed')
if '[More RPG Runtime Trace]' in s:
    raise SystemExit('diagnostic runtime trace leaked into clean lifecycle repair')
path.write_text(s, encoding='utf-8')
print('[More RPG 2.7.2] FORGE_CONSTRUCT_DEFERRED_INIT_1201_PASS source=run-352-deadlock timing=construct_event_enqueue_work')
