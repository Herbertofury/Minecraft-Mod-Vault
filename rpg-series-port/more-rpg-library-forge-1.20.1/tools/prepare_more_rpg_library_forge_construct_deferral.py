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

# Forge 1.20.1 constructs mods in parallel. More RPG currently enters MRPGCMod.init() directly from
# that constructor and immediately calls dependency APIs. Preserve its pre-RegisterEvent timing while
# moving the cross-mod work onto the serial deferred-work phase of FMLConstructModEvent.
import_anchor = 'import net.minecraftforge.fml.loading.FMLLoader;\n'
imports = ('import net.minecraftforge.fml.event.lifecycle.FMLConstructModEvent;\n'
           'import net.minecraftforge.fml.javafmlmod.FMLJavaModLoadingContext;\n'
           'import net.minecraftforge.eventbus.api.IEventBus;\n')
if s.count(import_anchor) != 1:
    raise SystemExit(f'Forge construct-deferral import seam drifted: found={s.count(import_anchor)}')
if 'FMLConstructModEvent' in s or 'FMLJavaModLoadingContext' in s:
    raise SystemExit('Forge construct deferral unexpectedly pre-exists')
s = s.replace(import_anchor, import_anchor + imports, 1)

old = ('        System.err.println("[More RPG Runtime Trace] FORGE_CONSTRUCTOR_BEFORE_MRPG_CLASS_INIT");\n'
       '        MRPGCMod.init();\n'
       '        System.err.println("[More RPG Runtime Trace] FORGE_CONSTRUCTOR_AFTER_MRPG_INIT");\n')
new = ('        IEventBus modBus = FMLJavaModLoadingContext.get().getModEventBus();\n'
       '        modBus.addListener(this::construct);\n'
       '        System.err.println("[More RPG Runtime Trace] FORGE_CONSTRUCTOR_DEFERRED_MRPG_INIT");\n')
if s.count(old) != 1:
    raise SystemExit(f'Forge direct MRPG init trace seam drifted: found={s.count(old)}')
s = s.replace(old, new, 1)

class_end = '\n}\n'
if not s.endswith(class_end):
    raise SystemExit('ForgeMod class terminator drifted')
method = '''\n    private void construct(FMLConstructModEvent event) {\n        event.enqueueWork(() -> {\n            System.err.println("[More RPG Runtime Trace] CONSTRUCT_DEFERRED_MRPG_INIT_BEGIN");\n            MRPGCMod.init();\n            System.err.println("[More RPG Runtime Trace] CONSTRUCT_DEFERRED_MRPG_INIT_END");\n        });\n    }\n'''
s = s[:-len(class_end)] + method + class_end
if s.count('modBus.addListener(this::construct);') != 1:
    raise SystemExit('Forge construct listener missing or duplicated')
if s.count('event.enqueueWork(() -> {') != 1 or s.count('MRPGCMod.init();') != 1:
    raise SystemExit('Forge deferred MRPG init contract missing or duplicated')
path.write_text(s, encoding='utf-8')
print('[More RPG 2.7.2] FORGE_CONSTRUCT_DEFERRED_INIT_1201_PASS timing=after_parallel_construction_before_register_events enqueueWork=true')
