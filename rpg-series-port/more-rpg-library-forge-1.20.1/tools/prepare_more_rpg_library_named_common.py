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
resource_paths = Path(__file__).with_name('prepare_more_rpg_library_1201_resource_paths.py')
api_pass = Path(__file__).with_name('prepare_more_rpg_library_1201_api_pass.py')
status_consumer_prepass = Path(__file__).with_name('prepare_more_rpg_library_1201_status_consumer_prepass.py')
api_wave2 = Path(__file__).with_name('prepare_more_rpg_library_1201_api_wave2.py')
loot_wave3 = Path(__file__).with_name('prepare_more_rpg_library_1201_loot_wave3.py')
loot_wave3_import_fix = Path(__file__).with_name('prepare_more_rpg_library_1201_loot_wave3_import_fix.py')
registry_wave4 = Path(__file__).with_name('prepare_more_rpg_library_1201_registry_wave4.py')
status_wave5a = Path(__file__).with_name('prepare_more_rpg_library_1201_status_consumers_wave5a.py')
api_wave5b = Path(__file__).with_name('prepare_more_rpg_library_1201_api_wave5b.py')
api_wave5c = Path(__file__).with_name('prepare_more_rpg_library_1201_api_wave5c.py')
runtime_trace = Path(__file__).with_name('prepare_more_rpg_library_runtime_trace.py')
for tool in (loader_neutral, resource_paths, api_pass, status_consumer_prepass, api_wave2, loot_wave3, loot_wave3_import_fix, registry_wave4, status_wave5a, api_wave5b, api_wave5c, runtime_trace):
    if not tool.is_file():
        raise SystemExit(f'missing More RPG preparer stage: {tool}')

subprocess.run([sys.executable, str(loader_neutral), str(modern), str(old), str(out)], check=True)
# Translate modern standard datapack/resource registry paths immediately after staging, before any
# API or runtime compatibility transforms. Both the first build and independent replay therefore
# consume the exact same target-native 1.20.1 data layout.
subprocess.run([sys.executable, str(resource_paths), str(out)], check=True)
subprocess.run([sys.executable, str(api_pass), str(out)], check=True)
subprocess.run([sys.executable, str(status_consumer_prepass), str(out)], check=True)
subprocess.run([sys.executable, str(api_wave2), str(out)], check=True)
subprocess.run([sys.executable, str(loot_wave3), str(out)], check=True)
subprocess.run([sys.executable, str(loot_wave3_import_fix), str(out)], check=True)
subprocess.run([sys.executable, str(registry_wave4), str(out)], check=True)
subprocess.run([sys.executable, str(status_wave5a), str(out)], check=True)
subprocess.run([sys.executable, str(api_wave5b), str(out)], check=True)
subprocess.run([sys.executable, str(api_wave5c), str(out)], check=True)
subprocess.run([sys.executable, str(runtime_trace), str(out)], check=True)

# Modern 2.7.2 intentionally carries datagen output in src/main/generated, and a subset of those paths
# also exists under src/main/resources. Gradle 8 rejects those overlapping Copy inputs. The generated
# tree is the newer release authority (for example tier_5_armors adds virtuoso_upgrade_crystal), so
# remove only static resource files shadowed by an existing generated file. Never hide collisions with
# a broad duplicatesStrategy: make the authority explicit and fail closed if the pinned known overlap
# disappears or any overlapping static path survives.
resources_root = out / 'common/src/main/resources'
generated_root = out / 'common/src/main/generated'
if not resources_root.is_dir() or not generated_root.is_dir():
    raise SystemExit('generated/static More RPG resource roots missing')
overlays = []
for generated in sorted(p for p in generated_root.rglob('*') if p.is_file()):
    rel = generated.relative_to(generated_root)
    static = resources_root / rel
    if static.is_file():
        static.unlink()
        overlays.append(rel.as_posix())
known_overlay = 'data/rpg_series/tags/items/loot_tier/tier_5_armors.json'
if known_overlay not in overlays:
    raise SystemExit(f'known generated/static resource overlap missing: {known_overlay}')
for rel in overlays:
    if (resources_root / rel).exists():
        raise SystemExit(f'static resource still shadows generated authority: {rel}')
if len(overlays) < 1:
    raise SystemExit('expected at least one generated/static resource overlay')

common_gradle = out / 'common/build.gradle'
if not common_gradle.is_file():
    raise SystemExit('generated common/build.gradle missing')
s = common_gradle.read_text()
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
anchor = '''    modImplementation("dev.kosmx.player-anim:player-animation-lib-fabric:${rootProject.player_anim_version}")\n'''
insert = '''    implementation("io.github.llamalad7:mixinextras-common:${rootProject.mixinextras_version}")\n    annotationProcessor("io.github.llamalad7:mixinextras-common:${rootProject.mixinextras_version}")\n'''
if s.count(anchor) != 1:
    raise SystemExit(f'common MixinExtras insertion seam drifted: found {s.count(anchor)}')
s = s.replace(anchor, insert + anchor, 1)
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

# Forge 47 discovers mod Mixins through the packaged JAR manifest. The common resource contains
# more-rpg-classes.mixins.json, but without MixinConfigs every More RPG mixin is inert in production.
forge_gradle = out / 'forge/build.gradle'
if not forge_gradle.is_file():
    raise SystemExit('generated forge/build.gradle missing')
f = forge_gradle.read_text()
manifest_anchor = "processResources { inputs.property 'version', project.version; filesMatching('META-INF/mods.toml') { expand(project.properties) } }\n"
manifest_block = """tasks.withType(Jar).configureEach {\n    manifest { attributes 'MixinConfigs': 'more-rpg-classes.mixins.json' }\n}\n"""
if f.count(manifest_anchor) != 1:
    raise SystemExit(f'Forge Mixin manifest insertion seam drifted: found {f.count(manifest_anchor)}')
if 'MixinConfigs' in f:
    raise SystemExit('Forge Mixin manifest contract unexpectedly pre-exists; inspect authority drift')
f = f.replace(manifest_anchor, manifest_block + manifest_anchor, 1)
if f.count("attributes 'MixinConfigs': 'more-rpg-classes.mixins.json'") != 1:
    raise SystemExit('Forge Mixin manifest contract missing or duplicated after insertion')
forge_gradle.write_text(f)

print('[More RPG 2.7.2] GENERATED_RESOURCE_OVERLAY_1201_PASS '
      f'overlaps={len(overlays)} authority=src/main/generated known={known_overlay}')
print('[More RPG 2.7.2] NAMED_COMMON_FOUNDATION_CLASSPATH_PASS spell_engine+spell_power+ranged+tiny_config=named-common forge-runtime=certified-production mixinextras=common-0.4.1')
print('[More RPG 2.7.2] GENERATED_RESOURCE_AUTHORITY_PRESERVED runtime_java_excludes=datagen')
print('[More RPG 2.7.2] FORGE_MIXIN_MANIFEST_REGISTRATION_PASS config=more-rpg-classes.mixins.json')