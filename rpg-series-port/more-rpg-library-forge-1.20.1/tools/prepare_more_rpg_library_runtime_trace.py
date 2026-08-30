#!/usr/bin/env python3
from pathlib import Path
import subprocess
import sys

if len(sys.argv) != 2:
    raise SystemExit('usage: prepare_more_rpg_library_runtime_trace.py <generated-port-root>')
root = Path(sys.argv[1]).resolve()
deferral = Path(__file__).with_name('prepare_more_rpg_library_forge_construct_deferral.py')
if not deferral.is_file():
    raise SystemExit(f'missing proven Forge construction deferral stage: {deferral}')
subprocess.run([sys.executable, str(deferral), str(root)], check=True)
print('[More RPG 2.7.2] RUNTIME_TRACE_REMOVED_CLEAN_DEFERRAL_PASS source=run-352-deadlock')
