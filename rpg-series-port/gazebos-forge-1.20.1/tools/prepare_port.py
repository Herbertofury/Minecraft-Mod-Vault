#!/usr/bin/env python3
from pathlib import Path
import shutil, subprocess, sys

if len(sys.argv) != 3:
    raise SystemExit('usage: prepare_port.py <gazebo-2.2.0-source> <output>')

src = Path(sys.argv[1]).resolve()
out = Path(sys.argv[2]).resolve()
TARGET_SHA = 'bc5e9f49e16d2ff31fb6d3aa31bab55ba0a634ee'
REFERENCE_1201_SHA = '4409bb838e8e5689b2892c25f1a9351592751a0d'

def head(path: Path) -> str:
    return subprocess.check_output(['git', '-C', str(path), 'rev-parse', 'HEAD'], text=True).strip()

actual = head(src)
if actual != TARGET_SHA:
    raise SystemExit(f'Gazebos target pin mismatch: expected {TARGET_SHA}, got {actual}')

if out.exists():
    shutil.rmtree(out)
shutil.copytree(src, out, ignore=shutil.ignore_patterns('.git', '.gradle', 'build', 'run', 'runs'))
shutil.rmtree(out / 'fabric', ignore_errors=True)
if (out / 'forge').exists():
    shutil.rmtree(out / 'forge')
(out / 'neoforge').rename(out / 'forge')

architectury_ids = {
    'common': '42f67fe939a2b8a3fe086c77a7f1075c',
    'forge': 'cb91e22794508958a3cc1fae12f20d9f',
}
for project_name, project_id in architectury_ids.items():
    cache = out / project_name / '.gradle' / 'architectury-cache'
    cache.mkdir(parents=True, exist_ok=True)
    (cache / 'projectID').write_text(project_id)

(out / 'settings.gradle').write_text('''pluginManagement {\n    repositories {\n        maven { url = 'https://maven.fabricmc.net/' }\n        maven { url = 'https://maven.architectury.dev/' }\n        maven { url = 'https://maven.minecraftforge.net/' }\n        gradlePluginPortal()\n    }\n}\nrootProject.name = 'gazebos-forge-1.20.1'\ninclude 'common'\ninclude 'forge'\n''')

(out / 'gradle.properties').write_text('''org.gradle.jvmargs=-Xmx3G\norg.gradle.parallel=true\nminecraft_version=1.20.1\nyarn_mappings=1.20.1+build.10\nforge_version=1.20.1-47.4.23\nfabric_loader_version=0.16.14\nmod_version=2.2.0+1.20.1\nmaven_group=net.rpg_series\narchives_name=gazebo\nenabled_platforms=forge\n''')

(out / 'build.gradle').write_text('''plugins {\n    id 'dev.architectury.loom' version '1.7.435' apply false\n    id 'architectury-plugin' version '3.4.164'\n    id 'com.github.johnrengelman.shadow' version '8.1.1' apply false\n}\narchitectury { minecraft = project.minecraft_version }\nallprojects { group = rootProject.maven_group; version = rootProject.mod_version }\nsubprojects {\n    apply plugin: 'dev.architectury.loom'\n    apply plugin: 'architectury-plugin'\n    apply plugin: 'maven-publish'\n    base { archivesName = "$rootProject.archives_name-$project.name" }\n    repositories {\n        mavenCentral()\n        maven { url = 'https://maven.fabricmc.net/' }\n        maven { url = 'https://maven.architectury.dev/' }\n        maven { url = 'https://maven.minecraftforge.net/' }\n    }\n    dependencies {\n        minecraft "com.mojang:minecraft:$rootProject.minecraft_version"\n        mappings "net.fabricmc:yarn:$rootProject.yarn_mappings:v2"\n    }\n    java {\n        toolchain { languageVersion = JavaLanguageVersion.of(17) }\n        withSourcesJar()\n        sourceCompatibility = JavaVersion.VERSION_17\n        targetCompatibility = JavaVersion.VERSION_17\n    }\n    tasks.withType(JavaCompile).configureEach { options.encoding='UTF-8'; options.release=17 }\n}\n''')

(out / 'common/build.gradle').write_text('''architectury { common 'forge' }\ndependencies {\n    modImplementation "net.fabricmc:fabric-loader:$rootProject.fabric_loader_version"\n    modImplementation files("$rootDir/libs/structure_pool_api-forge.jar")\n    modImplementation files("$rootDir/libs/tiny_config-forge.jar")\n}\n''')

(out / 'forge/gradle.properties').write_text('loom.platform = forge\n')
(out / 'forge/build.gradle').write_text('''plugins { id 'com.github.johnrengelman.shadow' }\narchitectury { platformSetupLoomIde(); forge() }\nconfigurations {\n    common { canBeResolved=true; canBeConsumed=false }\n    compileClasspath.extendsFrom common\n    runtimeClasspath.extendsFrom common\n    developmentForge.extendsFrom common\n    shadowBundle { canBeResolved=true; canBeConsumed=false }\n}\ndependencies {\n    forge "net.minecraftforge:forge:$rootProject.forge_version"\n    common(project(path: ':common', configuration: 'namedElements')) { transitive=false }\n    shadowBundle project(path: ':common', configuration: 'transformProductionForge')\n    modImplementation files("$rootDir/libs/structure_pool_api-forge.jar")\n    modImplementation files("$rootDir/libs/tiny_config-forge.jar")\n}\nprocessResources {\n    inputs.property 'version', project.version\n    filesMatching('META-INF/mods.toml') { expand(version: project.version) }\n}\nshadowJar { configurations=[project.configurations.shadowBundle]; archiveClassifier='dev-shadow' }\nremapJar { inputFile.set(shadowJar.archiveFile); dependsOn(shadowJar); archiveClassifier='' }\n''')

# Translate the tiny loader-specific surface; common source remains modern 2.2.0 authority.
java_root = out / 'forge/src/main/java'
replacements = {
    'net.gazebo.neoforge': 'net.gazebo.forge',
    'net.neoforged.fml.common.Mod': 'net.minecraftforge.fml.common.Mod',
    'net.neoforged.fml.loading.LoadingModList': 'net.minecraftforge.fml.loading.LoadingModList',
    'NeoForgeMod': 'ForgeMod',
    'NeoForgeUtil': 'ForgeUtil',
}
for p in java_root.rglob('*.java'):
    s = p.read_text()
    for old, new in replacements.items():
        s = s.replace(old, new)
    p.write_text(s)
old_pkg = java_root / 'net/gazebo/neoforge'
new_pkg = java_root / 'net/gazebo/forge'
if old_pkg.exists():
    new_pkg.parent.mkdir(parents=True, exist_ok=True)
    old_pkg.rename(new_pkg)
entry = new_pkg / 'NeoForgeMod.java'
if entry.exists():
    entry.rename(new_pkg / 'ForgeMod.java')

# 1.21.1 registry folder names are singular; 1.20.1 still resolves plural paths.
resource_root = out / 'common/src/main/resources/data/gazebo'
for old_name, new_name in [('loot_table', 'loot_tables'), ('structure', 'structures')]:
    old = resource_root / old_name
    new = resource_root / new_name
    if old.exists():
        if new.exists():
            raise SystemExit(f'Gazebos resource remap collision: {old_name} -> {new_name}')
        old.rename(new)

meta = out / 'forge/src/main/resources/META-INF'
meta.mkdir(parents=True, exist_ok=True)
for p in meta.glob('*neoforge*'):
    p.unlink()
(meta / 'mods.toml').write_text('''modLoader="javafml"\nloaderVersion="[47,)"\nlicense="ARR"\n[[mods]]\nmodId="gazebo"\nversion="${version}"\ndisplayName="Gazebos (RPG Series)"\nauthors="Daedelus"\ndescription="Adds gazebos to villages while preserving the modern 2.2.0 content set."\n[[dependencies.gazebo]]\nmodId="forge"\nmandatory=true\nversionRange="[47.4,48)"\nordering="NONE"\nside="BOTH"\n[[dependencies.gazebo]]\nmodId="minecraft"\nmandatory=true\nversionRange="[1.20.1,1.20.2)"\nordering="NONE"\nside="BOTH"\n[[dependencies.gazebo]]\nmodId="structure_pool_api"\nmandatory=true\nversionRange="[1.2.1,)"\nordering="AFTER"\nside="BOTH"\n[[dependencies.gazebo]]\nmodId="tiny_config"\nmandatory=true\nversionRange="[3.1.0,)"\nordering="AFTER"\nside="BOTH"\n''')

(out / 'PORT-PINS.txt').write_text(
    f'feature_authority={TARGET_SHA}\n'
    f'mappings_resource_reference_1_20_1={REFERENCE_1201_SHA}\n'
    'target_version=2.2.0\n'
    'minecraft=1.20.1\nforge=47.4.23\nyarn=1.20.1+build.10\njava=17\n'
    'structure_pool_api=1.2.1+1.20.1\ntiny_config=3.1.0+1.20.1\n'
    'architectury_loom=1.7.435\narchitectury_plugin=3.4.164\n'
)

required = [
    out / 'common/src/main/java/net/gazebo/GazeboMod.java',
    out / 'common/src/main/java/net/gazebo/config/Default.java',
    out / 'forge/src/main/java/net/gazebo/forge/PlatformImpl.java',
    out / 'forge/src/main/java/net/gazebo/forge/ForgeMod.java',
    out / 'common/src/main/resources/data/gazebo/loot_tables/chests/gazebo.json',
    out / 'forge/src/main/resources/META-INF/mods.toml',
]
missing = [str(p) for p in required if not p.exists()]
if missing:
    raise SystemExit('Gazebos preparation lost required modern source/content: ' + ', '.join(missing))
if (out / 'fabric').exists() or (out / 'neoforge').exists():
    raise SystemExit('non-Forge platform directory leaked into generated Gazebos port')
for p in (out / 'forge/src/main/java').rglob('*.java'):
    text = p.read_text()
    if 'net.neoforged' in text or 'NeoForge' in text:
        raise SystemExit(f'NeoForge symbol leaked into Gazebos Forge source: {p}')

structures = list((out / 'common/src/main/resources/data/gazebo/structures').rglob('*.nbt'))
if len(structures) != 17:
    raise SystemExit(f'Gazebos modern structure inventory regression: expected 17 NBTs, got {len(structures)}')
rs_counts = list((out / 'common/src/main/resources/data/gazebo/rs_pieces_spawn_counts').glob('*.json'))
rs_additions = list((out / 'common/src/main/resources/data/gazebo/rs_pool_additions').glob('*.json'))
litho = list((out / 'common/src/main/resources/data/gazebo/lithostitched/worldgen_modifier/village').glob('*.json'))
if (len(rs_counts), len(rs_additions), len(litho)) != (12, 12, 5):
    raise SystemExit(f'Gazebos modern compat inventory regression: counts={len(rs_counts)} additions={len(rs_additions)} litho={len(litho)}')
print('Gazebos 2.2.0 prepared for Forge 1.20.1')
print('feature_authority=' + TARGET_SHA)
print(f'modern_content=17_structures/12_rs_counts/12_rs_additions/5_lithostitched')
