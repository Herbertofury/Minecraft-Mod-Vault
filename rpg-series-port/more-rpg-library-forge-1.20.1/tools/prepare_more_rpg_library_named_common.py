#!/usr/bin/env python3
from pathlib import Path
import subprocess
import sys

if len(sys.argv) != 4:
    raise SystemExit('usage: prepare_more_rpg_library_named_common.py <modern-2.7.2-root> <old-1.20.1-root> <output-root>')
modern = Path(sys.argv[1]).resolve()
old = Path(sys.argv[2]).resolve()
out = Path(sys.argv[3]).resolve()
loader_neutral = Path(__file__).with_name('prepare_more_rpg_library_loader_neutral.py')
if not loader_neutral.is_file():
    raise SystemExit(f'missing loader-neutral preparer: {loader_neutral}')

# First preserve all existing full-source and loader-neutral preparation invariants.
subprocess.run([sys.executable, str(loader_neutral), str(modern), str(old), str(out)], check=True)

common_gradle = out / 'common/build.gradle'
if not common_gradle.is_file():
    raise SystemExit('generated common/build.gradle missing')
s = common_gradle.read_text()

# The common project uses Yarn named mappings. Feeding final Forge/SRG production artifacts into this
# classpath poisons dependency signatures (e.g. MobEffect/ResourceLocation/ParticleType Mojmap names)
# and produces hundreds of false downstream source errors. Compile common strictly against the named
# common artifacts produced from the SAME exact-source foundation replay. Certified Forge artifacts
# stay on forge/runtime classpaths and retain their exact release-hash gates.
pairs = [
    ('spell_engine_forge_jar', 'spell_engine_common_jar'),
    ('spell_power_forge_jar', 'spell_power_common_jar'),
    ('ranged_weapon_api_forge_jar', 'ranged_weapon_api_common_jar'),
    ('tiny_config_forge_jar', 'tiny_config_common_jar'),
]
for old_prop, new_prop in pairs:
    needle = f"modImplementation files(rootProject.property('{old_prop}'))"
    repl = f"modImplementation files(rootProject.property('{new_prop}'))"
    if s.count(needle) != 1:
        raise SystemExit(f'common dependency seam drifted for {old_prop}: found {s.count(needle)}')
    s = s.replace(needle, repl, 1)

# Modern 2.7.2 stealth mixins use MixinExtras annotations/types in common. Forge bundles the runtime
# artifact separately; common only needs the loader-neutral API + annotation processor at compile time.
anchor = '''    modImplementation("dev.kosmx.player-anim:player-animation-lib-fabric:${rootProject.player_anim_version}")\n'''
insert = '''    implementation("io.github.llamalad7:mixinextras-common:${rootProject.mixinextras_version}")\n    annotationProcessor("io.github.llamalad7:mixinextras-common:${rootProject.mixinextras_version}")\n'''
if s.count(anchor) != 1:
    raise SystemExit(f'common MixinExtras insertion seam drifted: found {s.count(anchor)}')
s = s.replace(anchor, insert + anchor, 1)

for old_prop, _ in pairs:
    if old_prop in s:
        raise SystemExit(f'production Forge dependency still leaks into common compile classpath: {old_prop}')
for _, new_prop in pairs:
    if s.count(new_prop) != 1:
        raise SystemExit(f'named common dependency missing or duplicated: {new_prop}')
if s.count('mixinextras-common') != 2:
    raise SystemExit('MixinExtras common compile/annotation contract missing or duplicated')
common_gradle.write_text(s)

print('[More RPG 2.7.2] NAMED_COMMON_FOUNDATION_CLASSPATH_PASS '
      'spell_engine+spell_power+ranged+tiny_config=named-common forge-runtime=certified-production mixinextras=common-0.4.1')
