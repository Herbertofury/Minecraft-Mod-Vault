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
api_pass = Path(__file__).with_name('prepare_more_rpg_library_1201_api_pass.py')
for tool in (loader_neutral, api_pass):
    if not tool.is_file():
        raise SystemExit(f'missing More RPG preparer stage: {tool}')

# First preserve all existing full-source and loader-neutral preparation invariants.
subprocess.run([sys.executable, str(loader_neutral), str(modern), str(old), str(out)], check=True)

# Then adapt only proven 1.21 -> 1.20.1 API seams. This stage is fail-closed and owns its own
# surviving-symbol checks, so upstream drift cannot silently weaken the target contract.
subprocess.run([sys.executable, str(api_pass), str(out)], check=True)

common_gradle = out / 'common/build.gradle'
if not common_gradle.is_file():
    raise SystemExit('generated common/build.gradle missing')
s = common_gradle.read_text()

# The common project uses Yarn named mappings. Feeding final Forge/SRG production artifacts into this
# classpath poisons dependency signatures and produces hundreds of false downstream source errors.
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

# Modern 2.7.2 stealth mixins use MixinExtras annotations/types in common. Forge bundles runtime
# separately; common only needs the loader-neutral API + annotation processor at compile time.
anchor = '''    modImplementation("dev.kosmx.player-anim:player-animation-lib-fabric:${rootProject.player_anim_version}")\n'''
insert = '''    implementation("io.github.llamalad7:mixinextras-common:${rootProject.mixinextras_version}")\n    annotationProcessor("io.github.llamalad7:mixinextras-common:${rootProject.mixinextras_version}")\n'''
if s.count(anchor) != 1:
    raise SystemExit(f'common MixinExtras insertion seam drifted: found {s.count(anchor)}')
s = s.replace(anchor, insert + anchor, 1)

# Keep the 2.7.2 generated-resource authority in the release resources, while excluding datagen Java
# from production compilation. Datagen remains a separately-portable build-time lane instead of forcing
# Fabric datagen classes into a native Forge runtime JAR.
source_anchor = '''loom {\n    accessWidenerPath = file('src/main/resources/more-rpg-classes.accesswidener')\n}\n'''
source_block = '''sourceSets {\n    main {\n        java { exclude 'net/more_rpg_classes/datagen/**' }\n        resources { srcDir 'src/main/generated' }\n    }\n}\n'''
if s.count(source_anchor) != 1:
    raise SystemExit(f'common source-set insertion seam drifted: found {s.count(source_anchor)}')
s = s.replace(source_anchor, source_block + source_anchor, 1)

for old_prop, _ in pairs:
    if old_prop in s:
        raise SystemExit(f'production Forge dependency still leaks into common compile classpath: {old_prop}')
for _, new_prop in pairs:
    if s.count(new_prop) != 1:
        raise SystemExit(f'named common dependency missing or duplicated: {new_prop}')
if s.count('mixinextras-common') != 2:
    raise SystemExit('MixinExtras common compile/annotation contract missing or duplicated')
if s.count("exclude 'net/more_rpg_classes/datagen/**'") != 1 or s.count("srcDir 'src/main/generated'") != 1:
    raise SystemExit('runtime/datagen source separation contract missing or duplicated')
common_gradle.write_text(s)

print('[More RPG 2.7.2] NAMED_COMMON_FOUNDATION_CLASSPATH_PASS '
      'spell_engine+spell_power+ranged+tiny_config=named-common forge-runtime=certified-production mixinextras=common-0.4.1')
print('[More RPG 2.7.2] GENERATED_RESOURCE_AUTHORITY_PRESERVED runtime_java_excludes=datagen')
