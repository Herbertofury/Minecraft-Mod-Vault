#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 3:
    raise SystemExit('usage: compat_pass_6k.py <generated-port-root> <spell-engine-1.20.1-baseline>')

root = Path(sys.argv[1]).resolve()
forge_build = root / 'forge/build.gradle'
s = forge_build.read_text()

# Forge 1.20.1 does not provide MixinExtras to mods. Spell Engine owns @WrapOperation usage in
# its common mixins, so carry the already-pinned MixinExtras 0.4.1 Forge library at the Spell Engine
# boundary. Architectury Loom's include() produces the supported jar-in-jar form; do not add this
# dependency to Archers and do not shade/relocate it into Spell Engine classes.
anchor = '    modImplementation "me.shedaniel.cloth:cloth-config-forge:$rootProject.cloth_config_version"\n'
line = '    implementation(include("io.github.llamalad7:mixinextras-forge:$rootProject.mixinextras_version"))\n'
if 'mixinextras-forge' not in s:
    if anchor not in s:
        raise SystemExit('Spell Engine Forge dependency anchor missing')
    s = s.replace(anchor, anchor + line, 1)
    forge_build.write_text(s)

final = forge_build.read_text()
if final.count('mixinextras-forge') != 1:
    raise SystemExit('Spell Engine Forge MixinExtras ownership is missing or duplicated')
if 'implementation(include("io.github.llamalad7:mixinextras-forge:$rootProject.mixinextras_version"))' not in final:
    raise SystemExit('Spell Engine Forge MixinExtras is not using Architectury Loom include()')

print('Spell Engine compatibility pass 6k applied: MixinExtras Forge runtime carried by Spell Engine')
