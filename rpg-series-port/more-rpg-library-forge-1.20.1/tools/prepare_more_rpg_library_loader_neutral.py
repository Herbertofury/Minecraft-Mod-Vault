#!/usr/bin/env python3
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile

if len(sys.argv) != 4:
    raise SystemExit('usage: prepare_more_rpg_library_loader_neutral.py <modern-2.7.2-root> <old-1.20.1-root> <output-root>')
modern = Path(sys.argv[1]).resolve()
old = Path(sys.argv[2]).resolve()
out = Path(sys.argv[3]).resolve()
base = Path(__file__).with_name('prepare_more_rpg_library.py')
for p in (modern, old):
    if not p.is_dir():
        raise SystemExit(f'missing authority tree: {p}')
if not base.is_file():
    raise SystemExit(f'missing base full-source preparer: {base}')

# Keep the exact 2.7.2 authority immutable. Work on a byte-for-byte temporary copy and adapt only
# Fabric Loader environment/mod-presence calls to the provider seam installed by the base preparer.
# This catches loader hard-links anywhere in the modern common tree (including worldgen structures),
# instead of maintaining a brittle per-file allowlist.
with tempfile.TemporaryDirectory(prefix='more-rpg-272-loader-neutral-') as td:
    staged = Path(td) / 'modern-272'
    shutil.copytree(modern, staged)
    java = staged / 'common/src/main/java'
    if not java.is_dir():
        raise SystemExit('modern 2.7.2 common Java tree missing')

    changed = []
    mod_loaded_calls = 0
    dev_env_calls = 0
    for f in sorted(java.rglob('*.java')):
        before = f.read_text(errors='strict')
        after = before.replace('import net.fabricmc.loader.api.FabricLoader;\n', '')
        count = after.count('FabricLoader.getInstance().isModLoaded(')
        if count:
            mod_loaded_calls += count
            after = after.replace(
                'FabricLoader.getInstance().isModLoaded(',
                'net.more_rpg_classes.compat.MoreRpgPlatform.isModLoaded.test('
            )
        count = after.count('FabricLoader.getInstance().isDevelopmentEnvironment()')
        if count:
            dev_env_calls += count
            after = after.replace(
                'FabricLoader.getInstance().isDevelopmentEnvironment()',
                'net.more_rpg_classes.compat.MoreRpgPlatform.isDevelopmentEnvironment.getAsBoolean()'
            )
        if after != before:
            f.write_text(after)
            changed.append(f.relative_to(staged).as_posix())

    lingering = []
    for f in sorted(java.rglob('*.java')):
        text = f.read_text(errors='strict')
        if 'FabricLoader' in text or 'net.fabricmc.loader.api' in text:
            lingering.append(f.relative_to(staged).as_posix())
    if lingering:
        raise SystemExit('Fabric Loader hard-links survived loader-neutral authority staging: ' + ', '.join(lingering[:80]))
    if mod_loaded_calls < 2:
        raise SystemExit(f'expected multiple 2.7.2 mod-presence loader calls, adapted only {mod_loaded_calls}')
    if dev_env_calls < 1:
        raise SystemExit('expected 2.7.2 development-environment loader call was not adapted')

    subprocess.run([sys.executable, str(base), str(staged), str(old), str(out)], check=True)

print('[More RPG 2.7.2] COMMON_FABRIC_LOADER_PROVIDER_ADAPTATION_PASS '
      f'files={len(changed)} mod_loaded_calls={mod_loaded_calls} dev_env_calls={dev_env_calls}')
