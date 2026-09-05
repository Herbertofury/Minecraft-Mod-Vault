#!/usr/bin/env python3
from pathlib import Path
import subprocess
import sys

if len(sys.argv) != 2:
    raise SystemExit('usage: prepare_more_rpg_library_runtime_trace.py <generated-port-root>')
root = Path(sys.argv[1]).resolve()
forge_registry_wave6 = Path(__file__).with_name('prepare_more_rpg_library_1201_forge_registry_wave6.py')
draw_hearts_signature = Path(__file__).with_name('prepare_more_rpg_library_1201_draw_hearts_signature.py')
stealth_model_render = Path(__file__).with_name('prepare_more_rpg_library_1201_stealth_model_render.py')
external_school_attributes = Path(__file__).with_name('prepare_more_rpg_library_1201_external_school_attributes.py')
deferral = Path(__file__).with_name('prepare_more_rpg_library_forge_construct_deferral.py')
for tool in (forge_registry_wave6, draw_hearts_signature, stealth_model_render, external_school_attributes, deferral):
    if not tool.is_file():
        raise SystemExit(f'missing proven More RPG runtime repair stage: {tool}')
subprocess.run([sys.executable, str(forge_registry_wave6), str(root)], check=True)
subprocess.run([sys.executable, str(draw_hearts_signature), str(root)], check=True)
subprocess.run([sys.executable, str(stealth_model_render), str(root)], check=True)
subprocess.run([sys.executable, str(external_school_attributes), str(root)], check=True)
subprocess.run([sys.executable, str(deferral), str(root)], check=True)
print('[More RPG 2.7.2] FORGE_REGISTRY_EVENTS_1201_RUNTIME_BRIDGE_READY source=upstream-2.7.2-neoforge-loader')
print('[More RPG 2.7.2] RUNTIME_TRACE_REMOVED_CLEAN_DEFERRAL_PASS source=run-352-deadlock')
print('[More RPG 2.7.2] EXTERNAL_SCHOOL_ATTRIBUTE_RUNTIME_BRIDGE_READY source=run-353-registry-boundary')
print('[More RPG 2.7.2] DRAW_HEARTS_1201_RUNTIME_BRIDGE_READY source=run-355-mapped-client')
print('[More RPG 2.7.2] STEALTH_RENDER_1201_RUNTIME_BRIDGE_READY source=client-preaudit')
