#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 3:
    raise SystemExit('usage: compat_pass_6g.py <generated-port-root> <spell-engine-1.20.1-baseline>')
root = Path(sys.argv[1]).resolve()
forge_build = root / 'forge/build.gradle'
s = forge_build.read_text()
anchor = '    modImplementation "me.shedaniel.cloth:cloth-config-forge:$rootProject.cloth_config_version"\n'
addition = '''    // Spell Engine common Mixins use MixinExtras at runtime. Embed the Forge service JAR exactly
    // like the already-runtime-proven Ranged Weapon API port; compileOnly keeps AP symbols explicit.
    compileOnly(annotationProcessor("io.github.llamalad7:mixinextras-common:$rootProject.mixinextras_version"))
    implementation(include("io.github.llamalad7:mixinextras-forge:$rootProject.mixinextras_version"))
'''
if 'mixinextras-forge' not in s:
    if anchor not in s:
        raise SystemExit('Forge dependency anchor missing')
    s = s.replace(anchor, anchor + addition)
forge_build.write_text(s)

final = forge_build.read_text()
for required in ('mixinextras-common', 'implementation(include("io.github.llamalad7:mixinextras-forge:'):
    if required not in final:
        raise SystemExit(f'MixinExtras Forge runtime dependency missing: {required}')
print('Spell Engine compatibility pass 6g applied: embedded Forge MixinExtras runtime service')
