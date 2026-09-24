#!/usr/bin/env python3
from pathlib import Path
import sys
import zipfile

if len(sys.argv) != 2:
    raise SystemExit('usage: compat_pass_6.py <generated-port-root>')

root = Path(sys.argv[1]).resolve()
props = root / 'gradle.properties'
forge_build = root / 'forge/build.gradle'
ranged_jar = root / 'libs/ranged-weapon-api.jar'

if not ranged_jar.is_file():
    raise SystemExit(f'Ranged Weapon API release JAR missing: {ranged_jar}')

# Ranged Weapon API 2.3.4 deliberately embeds MixinExtras in its distributable JAR.
# Architectury/Loom local file dependencies do not expand nested JarJar entries into a
# downstream development run classpath, so Jewelry's dev client needs the same library
# exposed directly. This is a DEVELOPMENT CLASSPATH bridge only; Jewelry must not package
# another copy of MixinExtras.
expected_nested = 'META-INF/jars/mixinextras-forge-0.4.1.jar'
with zipfile.ZipFile(ranged_jar) as zf:
    names = set(zf.namelist())
if expected_nested not in names:
    raise SystemExit(
        'Ranged Weapon API dependency contract changed: expected embedded '
        f'{expected_nested} in {ranged_jar}'
    )

prop_text = props.read_text()
if 'mixinextras_version=' not in prop_text:
    if not prop_text.endswith('\n'):
        prop_text += '\n'
    prop_text += 'mixinextras_version=0.4.1\n'
props.write_text(prop_text)

build = forge_build.read_text()
anchor = "    modImplementation files(rootProject.file('libs/ranged-weapon-api.jar'))\n"
bridge = '''    modImplementation files(rootProject.file('libs/ranged-weapon-api.jar'))

    // Ranged Weapon API embeds MixinExtras for real installations, but Loom does not
    // expand nested JarJar libraries from this local-file mod dependency into a downstream
    // development run. Mirror Ranged's exact MixinExtras version on the dev runtime only;
    // do NOT include() it in Jewelry's release JAR.
    implementation "io.github.llamalad7:mixinextras-forge:$rootProject.mixinextras_version"
'''
if bridge not in build:
    if anchor not in build:
        raise SystemExit('Jewelry Forge build lost Ranged Weapon API dependency anchor')
    build = build.replace(anchor, bridge, 1)
forge_build.write_text(build)

final_build = forge_build.read_text()
required = 'implementation "io.github.llamalad7:mixinextras-forge:$rootProject.mixinextras_version"'
if required not in final_build:
    raise SystemExit('Jewelry downstream MixinExtras dev-runtime bridge missing')
for forbidden in (
    'include("io.github.llamalad7:mixinextras-forge',
    'include \'io.github.llamalad7:mixinextras-forge',
    'shadowBundle "io.github.llamalad7:mixinextras-forge',
):
    if forbidden in final_build:
        raise SystemExit(f'Jewelry must not package its own MixinExtras copy: {forbidden}')
if 'mixinextras_version=0.4.1' not in props.read_text():
    raise SystemExit('Jewelry MixinExtras version pin missing')

print('Jewelry compatibility pass 6 applied: downstream Ranged/MixinExtras dev-runtime bridge (no release duplication)')
