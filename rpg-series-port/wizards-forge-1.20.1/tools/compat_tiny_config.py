#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 2:
    raise SystemExit('usage: compat_tiny_config.py <generated-wizards-root>')
root = Path(sys.argv[1]).resolve()

props = root / 'gradle.properties'
s = props.read_text()
s = s.replace('tiny_config_version=2.3.2', 'tiny_config_version=3.1.0+1.20.1')
props.write_text(s)

common = root / 'common/build.gradle'
s = common.read_text()
old = '    modImplementation "maven.modrinth:tiny-config:$rootProject.tiny_config_version-${platform}"'
new = "    modImplementation files('../libs/tiny-config-common.jar')"
if old not in s:
    raise SystemExit('expected Wizards common TinyConfig dependency coordinate not found')
common.write_text(s.replace(old, new))

forge = root / 'forge/build.gradle'
s = forge.read_text()
old = '    modImplementation "maven.modrinth:tiny-config:$rootProject.tiny_config_version-forge"'
new = "    modImplementation files('../libs/tiny-config-forge.jar')"
if old not in s:
    raise SystemExit('expected Wizards Forge TinyConfig dependency coordinate not found')
forge.write_text(s.replace(old, new))

mods = root / 'forge/src/main/resources/META-INF/mods.toml'
s = mods.read_text()
if 'modId="tiny_config"' not in s:
    marker = '[[dependencies.wizards]]\nmodId="spell_engine"'
    dependency = '''[[dependencies.wizards]]\nmodId="tiny_config"\nmandatory=true\nversionRange="[3.1.0,)"\nordering="NONE"\nside="BOTH"\n'''
    if marker not in s:
        raise SystemExit('Wizards mods.toml dependency insertion point not found')
    s = s.replace(marker, dependency + marker)
    mods.write_text(s)

print('Wizards TinyConfig compatibility layer applied: exact 3.1.0 API staged as separate native Forge 1.20.1 JAR')
