#!/usr/bin/env python3
from pathlib import Path
import shutil, subprocess, sys

if len(sys.argv) != 3:
    raise SystemExit('usage: prepare_port.py <tiny-config-3.1.0-source> <output>')

src = Path(sys.argv[1]).resolve()
out = Path(sys.argv[2]).resolve()
TARGET_SHA = 'e20fc8ac72fde8274f0df72de2ebb81ffe6f8727'

def head(path: Path) -> str:
    return subprocess.check_output(['git', '-C', str(path), 'rev-parse', 'HEAD'], text=True).strip()

actual = head(src)
if actual != TARGET_SHA:
    raise SystemExit(f'TinyConfig target pin mismatch: expected {TARGET_SHA}, got {actual}')

if out.exists():
    shutil.rmtree(out)
shutil.copytree(src, out, ignore=shutil.ignore_patterns('.git', '.gradle', 'build', 'run', 'runs'))
shutil.rmtree(out / 'fabric', ignore_errors=True)
if (out / 'forge').exists():
    shutil.rmtree(out / 'forge')
(out / 'neoforge').rename(out / 'forge')

architectury_ids = {
    'common': '9fc74092c9d7743dc70145a61b3bcf2e',
    'forge': '685c607f4a2b8df7dd7b39fb42d903de',
}
for project_name, project_id in architectury_ids.items():
    cache = out / project_name / '.gradle' / 'architectury-cache'
    cache.mkdir(parents=True, exist_ok=True)
    (cache / 'projectID').write_text(project_id)

(out / 'settings.gradle').write_text('''pluginManagement {\n    repositories {\n        maven { url = 'https://maven.fabricmc.net/' }\n        maven { url = 'https://maven.architectury.dev/' }\n        maven { url = 'https://maven.minecraftforge.net/' }\n        gradlePluginPortal()\n    }\n}\nrootProject.name = 'tiny-config-forge-1.20.1'\ninclude 'common'\ninclude 'forge'\n''')

(out / 'gradle.properties').write_text('''org.gradle.jvmargs=-Xmx2G\norg.gradle.parallel=true\nminecraft_version=1.20.1\nyarn_mappings=1.20.1+build.10\nforge_version=1.20.1-47.4.23\nfabric_loader_version=0.16.14\nmod_version=3.1.0+1.20.1\nmaven_group=net.tiny_config\narchives_name=tiny_config\nenabled_platforms=forge\n''')

(out / 'build.gradle').write_text('''plugins {\n    id 'dev.architectury.loom' version '1.7.435' apply false\n    id 'architectury-plugin' version '3.4.164'\n    id 'com.github.johnrengelman.shadow' version '8.1.1' apply false\n}\narchitectury { minecraft = project.minecraft_version }\nallprojects { group = rootProject.maven_group; version = rootProject.mod_version }\nsubprojects {\n    apply plugin: 'dev.architectury.loom'\n    apply plugin: 'architectury-plugin'\n    apply plugin: 'maven-publish'\n    base { archivesName = "$rootProject.archives_name-$project.name" }\n    repositories {\n        mavenCentral()\n        maven { url = 'https://maven.fabricmc.net/' }\n        maven { url = 'https://maven.architectury.dev/' }\n        maven { url = 'https://maven.minecraftforge.net/' }\n    }\n    dependencies {\n        minecraft "com.mojang:minecraft:$rootProject.minecraft_version"\n        mappings "net.fabricmc:yarn:$rootProject.yarn_mappings:v2"\n    }\n    java {\n        toolchain { languageVersion = JavaLanguageVersion.of(17) }\n        withSourcesJar()\n        sourceCompatibility = JavaVersion.VERSION_17\n        targetCompatibility = JavaVersion.VERSION_17\n    }\n    tasks.withType(JavaCompile).configureEach {\n        options.encoding = 'UTF-8'\n        options.release = 17\n    }\n}\n''')

(out / 'common/build.gradle').write_text('''architectury { common 'forge' }\ndependencies {\n    modImplementation "net.fabricmc:fabric-loader:$rootProject.fabric_loader_version"\n}\n''')

(out / 'forge/gradle.properties').write_text('loom.platform = forge\n')
(out / 'forge/build.gradle').write_text('''plugins { id 'com.github.johnrengelman.shadow' }\narchitectury { platformSetupLoomIde(); forge() }\nconfigurations {\n    common { canBeResolved=true; canBeConsumed=false }\n    compileClasspath.extendsFrom common\n    runtimeClasspath.extendsFrom common\n    developmentForge.extendsFrom common\n    shadowBundle { canBeResolved=true; canBeConsumed=false }\n}\ndependencies {\n    forge "net.minecraftforge:forge:$rootProject.forge_version"\n    common(project(path: ':common', configuration: 'namedElements')) { transitive=false }\n    shadowBundle project(path: ':common', configuration: 'transformProductionForge')\n}\nprocessResources {\n    inputs.property 'version', project.version\n    filesMatching('META-INF/mods.toml') { expand(version: project.version) }\n}\nshadowJar {\n    configurations=[project.configurations.shadowBundle]\n    archiveClassifier='dev-shadow'\n    manifest { attributes 'MixinConfigs': 'tiny_config.mixins.json' }\n}\nremapJar { inputFile.set(shadowJar.archiveFile); dependsOn(shadowJar); archiveClassifier='' }\n''')

java_root = out / 'forge/src/main/java'
replacements = {
    'net.tiny_config.neoforge': 'net.tiny_config.forge',
    'net.neoforged.fml.ModList': 'net.minecraftforge.fml.ModList',
    'net.neoforged.fml.loading.FMLPaths': 'net.minecraftforge.fml.loading.FMLPaths',
    'net.neoforged.fml.common.Mod': 'net.minecraftforge.fml.common.Mod',
    'Platform.Type.NEOFORGE': 'Platform.Type.FORGE',
    'NeoForgeUtil': 'ForgeUtil',
    'ExampleModNeoForge': 'ExampleModForge',
}
for p in java_root.rglob('*.java'):
    s = p.read_text()
    for old, new in replacements.items():
        s = s.replace(old, new)
    p.write_text(s)
old_pkg = java_root / 'net/tiny_config/neoforge'
new_pkg = java_root / 'net/tiny_config/forge'
if old_pkg.exists():
    new_pkg.parent.mkdir(parents=True, exist_ok=True)
    old_pkg.rename(new_pkg)
entry = new_pkg / 'ExampleModNeoForge.java'
if entry.exists():
    entry.rename(new_pkg / 'ExampleModForge.java')

mixin = out / 'common/src/main/resources/tiny_config.mixins.json'
if mixin.exists():
    mixin.write_text(mixin.read_text().replace('JAVA_21', 'JAVA_17'))

meta = out / 'forge/src/main/resources/META-INF'
meta.mkdir(parents=True, exist_ok=True)
for p in meta.glob('*neoforge*'):
    p.unlink()
(meta / 'mods.toml').write_text('''modLoader="javafml"\nloaderVersion="[47,)"\nlicense="CC0-1.0"\n[[mods]]\nmodId="tiny_config"\nversion="${version}"\ndisplayName="Tiny Config"\nauthors="Daedelus"\nlogoFile="logo.png"\ndescription="A very small config library."\n[[dependencies.tiny_config]]\nmodId="forge"\nmandatory=true\nversionRange="[47.4,48)"\nordering="NONE"\nside="BOTH"\n[[dependencies.tiny_config]]\nmodId="minecraft"\nmandatory=true\nversionRange="[1.20.1,1.20.2)"\nordering="NONE"\nside="BOTH"\n''')

(out / 'PORT-PINS.txt').write_text(
    f'target={TARGET_SHA}\n'
    'target_version=3.1.0\n'
    'minecraft=1.20.1\nforge=47.4.23\nyarn=1.20.1+build.10\njava=17\n'
    'architectury_loom=1.7.435\narchitectury_plugin=3.4.164\n'
    f'architectury_common_project_id={architectury_ids["common"]}\n'
    f'architectury_forge_project_id={architectury_ids["forge"]}\n'
)

required = [
    out / 'common/src/main/java/net/tiny_config/ConfigManager.java',
    out / 'common/src/main/java/net/tiny_config/versioning/Versionable.java',
    out / 'forge/src/main/java/net/tiny_config/forge/PlatformImpl.java',
    out / 'forge/src/main/java/net/tiny_config/forge/ExampleModForge.java',
    out / 'forge/src/main/resources/META-INF/mods.toml',
]
missing = [str(p) for p in required if not p.exists()]
if missing:
    raise SystemExit('TinyConfig preparation lost required source: ' + ', '.join(missing))
if (out / 'fabric').exists() or (out / 'neoforge').exists():
    raise SystemExit('non-Forge platform directory leaked into generated TinyConfig port')
for p in (out / 'forge/src/main/java').rglob('*.java'):
    text = p.read_text()
    if 'net.neoforged' in text or 'NeoForge.' in text:
        raise SystemExit(f'NeoForge symbol leaked into TinyConfig Forge source: {p}')
print('TinyConfig 3.1.0 target prepared for native Forge 1.20.1')
print('target=' + TARGET_SHA)
