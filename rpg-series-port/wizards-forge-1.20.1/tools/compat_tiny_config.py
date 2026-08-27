#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 2:
    raise SystemExit('usage: compat_tiny_config.py <generated-wizards-root>')
root = Path(sys.argv[1]).resolve()

props = root / 'gradle.properties'
s = props.read_text()
lines = s.splitlines()
old_version = 'tiny_config_version=2.3.2'
new_version = 'tiny_config_version=3.1.0+1.20.1'
if old_version in lines:
    lines[lines.index(old_version)] = new_version
elif new_version not in lines:
    raise SystemExit('expected Wizards TinyConfig version property not found')
s = '\n'.join(lines) + '\n'
# Spell Engine is supplied to Wizards as a local mod JAR. Loom does not automatically restore
# the Gradle runtime metadata behind nested JAR-in-JAR libraries from that local file dependency.
# Keep exact versions pinned here so dev/CI sees the same runtime that packaged Forge sees.
if 'spell_engine_tiny_config_version=' not in s:
    s += 'spell_engine_tiny_config_version=2.3.2\n'
if 'mixinextras_version=' not in s:
    s += 'mixinextras_version=0.4.1\n'
props.write_text(s)

common = root / 'common/build.gradle'
s = common.read_text()
old = '    modImplementation "maven.modrinth:tiny-config:$rootProject.tiny_config_version-${platform}"'
new = "    modImplementation files('../libs/tiny-config-common.jar')"
if old in s:
    s = s.replace(old, new, 1)
elif new not in s:
    raise SystemExit('expected Wizards common TinyConfig dependency coordinate not found')
common.write_text(s)

forge = root / 'forge/build.gradle'
s = forge.read_text()
old = '    modImplementation "maven.modrinth:tiny-config:$rootProject.tiny_config_version-forge"'
new = "    modImplementation files('../libs/tiny-config-forge.jar')"
if old in s:
    s = s.replace(old, new, 1)
elif new not in s:
    raise SystemExit('expected Wizards Forge TinyConfig dependency coordinate not found')

# The Spell Engine release correctly embeds MixinExtras Forge 0.4.1 and TinyConfig 2.3.2. Real Forge
# expands those nested libraries, but Architectury Loom does not expand them when Spell Engine itself
# enters this downstream project through `modImplementation files(...)`. Re-declare the same upstream
# coordinates for the dev runtime only. Nothing here is part of shadowBundle, so Wizards does not
# duplicate or repackage Spell Engine's libraries in its release JAR.
anchor = "    modImplementation files('../libs/spell-engine-forge.jar')\n"
dev_runtime = '''    implementation "io.github.llamalad7:mixinextras-forge:$rootProject.mixinextras_version"\n    def spellEngineTinyConfig = implementation("com.github.ZsoltMolnarrr:TinyConfig:$rootProject.spell_engine_tiny_config_version")\n    forgeRuntimeLibrary spellEngineTinyConfig\n'''
if 'mixinextras-forge:$rootProject.mixinextras_version' not in s:
    if anchor not in s:
        raise SystemExit('expected local Spell Engine Forge dependency not found')
    s = s.replace(anchor, anchor + dev_runtime, 1)
forge.write_text(s)

mods = root / 'forge/src/main/resources/META-INF/mods.toml'
s = mods.read_text()
if 'modId="tiny_config"' not in s:
    marker = '[[dependencies.wizards]]\nmodId="spell_engine"'
    dependency = '''[[dependencies.wizards]]\nmodId="tiny_config"\nmandatory=true\nversionRange="[3.1.0,)"\nordering="NONE"\nside="BOTH"\n'''
    if marker not in s:
        raise SystemExit('Wizards mods.toml dependency insertion point not found')
    s = s.replace(marker, dependency + marker)
    mods.write_text(s)

final_props = props.read_text()
for expected in ('tiny_config_version=3.1.0+1.20.1', 'spell_engine_tiny_config_version=2.3.2', 'mixinextras_version=0.4.1'):
    if expected not in final_props:
        raise SystemExit(f'Wizards runtime property missing: {expected}')
final_forge = forge.read_text()
for expected in (
    'implementation "io.github.llamalad7:mixinextras-forge:$rootProject.mixinextras_version"',
    'def spellEngineTinyConfig = implementation("com.github.ZsoltMolnarrr:TinyConfig:$rootProject.spell_engine_tiny_config_version")',
    'forgeRuntimeLibrary spellEngineTinyConfig',
):
    if expected not in final_forge:
        raise SystemExit(f'Wizards Spell Engine dev-runtime bridge missing: {expected}')

print('Wizards TinyConfig compatibility layer applied: TinyConfig 3.1.0 foundation + exact Spell Engine nested-runtime parity for Loom dev/CI')
