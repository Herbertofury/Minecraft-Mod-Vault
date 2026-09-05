#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 2:
    raise SystemExit('usage: patch_forge_production_client_library_directory.py <production-client-helper>')

helper = Path(sys.argv[1]).resolve()
if not helper.is_file():
    raise SystemExit(f'production client helper missing: {helper}')

s = helper.read_text()
old = '''        "assets_root": str(mc_home / "assets"),
        "assets_index_name": str(asset_index),
        "auth_uuid": "00000000000000000000000000000001",
'''
new = '''        "assets_root": str(mc_home / "assets"),
        "assets_index_name": str(asset_index),
        # Forge 47's installed version profile uses this in both -DlibraryDirectory and its
        # module-path expression. Point it at the exact Minecraft home populated by --installClient.
        "library_directory": str(libraries),
        "auth_uuid": "00000000000000000000000000000001",
'''
if s.count(old) != 1:
    raise SystemExit(f'[More RPG 2.7.2] expected one production launcher values seam, found {s.count(old)}')
if '"library_directory": str(libraries),' in s:
    raise SystemExit('[More RPG 2.7.2] production launcher library_directory patch unexpectedly already present')
s = s.replace(old, new, 1)

contracts = {
    '"library_directory": str(libraries),': 1,
    'libraries = mc_home / "libraries"': 2,
    'jvm = substitute(jvm, values)': 1,
    'game = substitute(game, values)': 1,
    'die(f"unresolved launcher placeholder ${{{key}}} in {arg!r}")': 1,
}
for needle, expected in contracts.items():
    actual = s.count(needle)
    if actual != expected:
        raise SystemExit(f'[More RPG 2.7.2] production launcher library-directory contract drifted for {needle!r}: expected {expected}, found {actual}')
helper.write_text(s)
print('[More RPG 2.7.2] FORGE_PRODUCTION_LIBRARY_DIRECTORY_PLACEHOLDER_PATCHED source=installed-forge-profile value=mc_home/libraries')
