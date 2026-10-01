#!/usr/bin/env python3
from pathlib import Path
import shutil, subprocess, sys

if len(sys.argv) != 4:
    raise SystemExit('usage: prepare_port.py <wizards-1.20.1-substrate> <wizards-current> <output>')

base = Path(sys.argv[1]).resolve()
target = Path(sys.argv[2]).resolve()
out = Path(sys.argv[3]).resolve()
BASE_SHA = '395ade75b50067c19f9b57a84c409bf962e09224'
TARGET_SHA = '82fd3a0f48366e6e406b4e7ca4b6d827a3793fb9'

def head(path: Path) -> str:
    return subprocess.check_output(['git', '-C', str(path), 'rev-parse', 'HEAD'], text=True).strip()

if head(base) != BASE_SHA:
    raise SystemExit(f'Wizards substrate pin mismatch: expected {BASE_SHA}, got {head(base)}')
if head(target) != TARGET_SHA:
    raise SystemExit(f'Wizards target pin mismatch: expected {TARGET_SHA}, got {head(target)}')

if out.exists():
    shutil.rmtree(out)
shutil.copytree(target, out, ignore=shutil.ignore_patterns('.git', '.gradle', 'build', 'run', 'runs'))
for name in ('fabric',):
    shutil.rmtree(out / name, ignore_errors=True)
if (out / 'forge').exists():
    shutil.rmtree(out / 'forge')
(out / 'neoforge').rename(out / 'forge')

# Same deterministic Architectury strategy proven by the graduated Armor Model API lane.
architectury_ids = {
    'common': '60594421397b7bb7c20be5881585b912',
    'forge': 'eb4f895b95648451330ef1772429cf29',
}
for project_name, project_id in architectury_ids.items():
    cache = out / project_name / '.gradle' / 'architectury-cache'
    cache.mkdir(parents=True, exist_ok=True)
    (cache / 'projectID').write_text(project_id)

(out / 'settings.gradle').write_text('''pluginManagement {\n    repositories {\n        maven { url = 'https://maven.fabricmc.net/' }\n        maven { url = 'https://maven.architectury.dev/' }\n        maven { url = 'https://maven.minecraftforge.net/' }\n        gradlePluginPortal()\n    }\n}\nrootProject.name = 'wizards-forge-1.20.1'\ninclude 'common'\ninclude 'forge'\n''')

(out / 'gradle.properties').write_text('''org.gradle.jvmargs=-Xmx5G\norg.gradle.parallel=true\nminecraft_version=1.20.1\nyarn_mappings=1.20.1+build.10\nforge_version=1.20.1-47.4.23\nfabric_loader_version=0.16.14\nmod_version=3.1.1+1.20.1\nmaven_group=net.rpg_series\narchives_name=wizards\nenabled_platforms=forge\nplayer_anim_version=1.0.2+1.19.4\ncloth_config_version=11.1.136\ntiny_config_version=2.3.2\ncurios_version=5.14.1+1.20.1\n''')

(out / 'build.gradle').write_text('''plugins {\n    id 'dev.architectury.loom' version '1.7.435' apply false\n    id 'architectury-plugin' version '3.4.164'\n    id 'com.github.johnrengelman.shadow' version '8.1.1' apply false\n}\narchitectury { minecraft = project.minecraft_version }\nallprojects { group = rootProject.maven_group; version = rootProject.mod_version }\nsubprojects {\n    apply plugin: 'dev.architectury.loom'\n    apply plugin: 'architectury-plugin'\n    apply plugin: 'maven-publish'\n    base { archivesName = "$rootProject.archives_name-$project.name" }\n    repositories {\n        mavenCentral()\n        maven { url = 'https://maven.fabricmc.net/' }\n        maven { url = 'https://maven.architectury.dev/' }\n        maven { url = 'https://maven.minecraftforge.net/' }\n        maven { url = 'https://maven.kosmx.dev/' }\n        maven { url = 'https://maven.shedaniel.me/' }\n        maven { url = 'https://api.modrinth.com/maven' }\n        maven { url = 'https://maven.theillusivec4.top/' }\n    }\n    dependencies {\n        minecraft "com.mojang:minecraft:$rootProject.minecraft_version"\n        mappings "net.fabricmc:yarn:$rootProject.yarn_mappings:v2"\n    }\n    java {\n        toolchain { languageVersion = JavaLanguageVersion.of(17) }\n        withSourcesJar()\n        sourceCompatibility = JavaVersion.VERSION_17\n        targetCompatibility = JavaVersion.VERSION_17\n    }\n    tasks.withType(JavaCompile).configureEach { options.encoding='UTF-8'; options.release=17; options.compilerArgs += ['-Xmaxerrs','2000'] }\n}\n''')

(out / 'common/build.gradle').write_text('''architectury { common 'forge' }\ndef platform = 'fabric'\nrepositories {\n    maven { url 'https://maven.kosmx.dev/' }\n    maven { url 'https://maven.shedaniel.me/' }\n    maven { url 'https://api.modrinth.com/maven' }\n}\ndependencies {\n    modImplementation "net.fabricmc:fabric-loader:$rootProject.fabric_loader_version"\n    modImplementation "dev.kosmx.player-anim:player-animation-lib-${platform}:$rootProject.player_anim_version"\n    modImplementation "me.shedaniel.cloth:cloth-config-${platform}:$rootProject.cloth_config_version"\n    modImplementation "maven.modrinth:tiny-config:$rootProject.tiny_config_version-${platform}"\n    modImplementation files('../libs/armor-model-api-common.jar')\n    modImplementation files('../libs/structure-pool-api-common.jar')\n    modImplementation files('../libs/runes-common.jar')\n    modImplementation files('../libs/spell-power-common.jar')\n    modImplementation files('../libs/spell-engine-common.jar')\n}\n''')

(out / 'forge/gradle.properties').write_text('loom.platform = forge\n')
(out / 'forge/build.gradle').write_text('''plugins { id 'com.github.johnrengelman.shadow' }\narchitectury { platformSetupLoomIde(); forge() }\ndef generatedResources = project(':common').file('src/main/generated')\nsourceSets { main { resources.srcDir generatedResources } }\nconfigurations {\n    common { canBeResolved=true; canBeConsumed=false }\n    compileClasspath.extendsFrom common\n    runtimeClasspath.extendsFrom common\n    developmentForge.extendsFrom common\n    shadowBundle { canBeResolved=true; canBeConsumed=false }\n}\ndependencies {\n    forge "net.minecraftforge:forge:$rootProject.forge_version"\n    common(project(path: ':common', configuration: 'namedElements')) { transitive=false }\n    shadowBundle project(path: ':common', configuration: 'transformProductionForge')\n    modImplementation "dev.kosmx.player-anim:player-animation-lib-forge:$rootProject.player_anim_version"\n    modImplementation "me.shedaniel.cloth:cloth-config-forge:$rootProject.cloth_config_version"\n    modImplementation "maven.modrinth:tiny-config:$rootProject.tiny_config_version-forge"\n    modImplementation "top.theillusivec4.curios:curios-forge:$rootProject.curios_version"\n    modImplementation files('../libs/armor-model-api-forge.jar')\n    modImplementation files('../libs/structure-pool-api-forge.jar')\n    modImplementation files('../libs/runes-forge.jar')\n    modImplementation files('../libs/spell-power-forge.jar')\n    modImplementation files('../libs/spell-engine-forge.jar')\n}\nprocessResources {\n    inputs.property 'version', project.version\n    filesMatching('META-INF/mods.toml') { expand(version: project.version) }\n}\nshadowJar {\n    configurations=[project.configurations.shadowBundle]\n    archiveClassifier='dev-shadow'\n    manifest { attributes 'MixinConfigs': 'wizards.mixins.json' }\n}\nremapJar { inputFile.set(shadowJar.archiveFile); dependsOn(shadowJar); archiveClassifier='' }\n''')

# NeoForge -> Forge loader translation. Keep current common behavior/content authoritative.
java_root = out / 'forge/src/main/java'
replacements = {
    'net.wizards.neoforge': 'net.wizards.forge',
    'net.neoforged.bus.api.IEventBus': 'net.minecraftforge.eventbus.api.IEventBus',
    'net.neoforged.fml.common.Mod': 'net.minecraftforge.fml.common.Mod',
    'net.neoforged.neoforge.common.NeoForge': 'net.minecraftforge.common.MinecraftForge',
    'net.neoforged.neoforge.event.village.VillagerTradesEvent': 'net.minecraftforge.event.village.VillagerTradesEvent',
    'net.neoforged.neoforge.registries.RegisterEvent': 'net.minecraftforge.registries.RegisterEvent',
    'NeoForge.EVENT_BUS': 'MinecraftForge.EVENT_BUS',
    'NeoForgeMod': 'ForgeMod',
}
for p in java_root.rglob('*.java'):
    s = p.read_text()
    for old, new in replacements.items():
        s = s.replace(old, new)
    p.write_text(s)
old_pkg = java_root / 'net/wizards/neoforge'
new_pkg = java_root / 'net/wizards/forge'
if old_pkg.exists():
    new_pkg.parent.mkdir(parents=True, exist_ok=True)
    old_pkg.rename(new_pkg)
entry = new_pkg / 'NeoForgeMod.java'
if entry.exists():
    entry.rename(new_pkg / 'ForgeMod.java')
entry = new_pkg / 'ForgeMod.java'
if entry.exists():
    s = entry.read_text()
    s = s.replace('import net.minecraftforge.eventbus.api.IEventBus;\n', 'import net.minecraftforge.eventbus.api.IEventBus;\nimport net.minecraftforge.fml.javafmlmod.FMLJavaModLoadingContext;\n')
    s = s.replace('public ForgeMod(IEventBus modBus) {', 'public ForgeMod() {\n        IEventBus modBus = FMLJavaModLoadingContext.get().getModEventBus();')
    s = s.replace('modBus.addListener(RegisterEvent.class, ForgeMod::register);', 'modBus.addListener(ForgeMod::register);')
    s = s.replace('MinecraftForge.EVENT_BUS.addListener(VillagerTradesEvent.class, ForgeMod::onVillagerTrades);', 'MinecraftForge.EVENT_BUS.addListener(ForgeMod::onVillagerTrades);')
    entry.write_text(s)

# Rewrite platform package references in resources/mixins if present.
for p in (out / 'forge/src/main/resources').rglob('*'):
    if p.is_file():
        try:
            s = p.read_text()
        except UnicodeDecodeError:
            continue
        s = s.replace('net.wizards.neoforge', 'net.wizards.forge')
        s = s.replace('JAVA_21', 'JAVA_17')
        p.write_text(s)

meta = out / 'forge/src/main/resources/META-INF'
meta.mkdir(parents=True, exist_ok=True)
for p in meta.glob('*neoforge*'):
    p.unlink()
(meta / 'mods.toml').write_text('''modLoader="javafml"\nloaderVersion="[47,)"\nlicense="All Rights Reserved"\nissueTrackerURL="https://github.com/ZsoltMolnarrr/Wizards/issues"\n[[mods]]\nmodId="wizards"\nversion="${version}"\ndisplayName="Wizards (RPG Series)"\nauthors="Daedelus"\nlogoFile="icon.png"\ndescription="Destroy your enemies with Arcane, Fire and Frost magic."\n[[dependencies.wizards]]\nmodId="forge"\nmandatory=true\nversionRange="[47.4,48)"\nordering="NONE"\nside="BOTH"\n[[dependencies.wizards]]\nmodId="minecraft"\nmandatory=true\nversionRange="[1.20.1,1.20.2)"\nordering="NONE"\nside="BOTH"\n[[dependencies.wizards]]\nmodId="armor_model_api"\nmandatory=true\nversionRange="[1.0.0,)"\nordering="NONE"\nside="BOTH"\n[[dependencies.wizards]]\nmodId="spell_engine"\nmandatory=true\nversionRange="[1.10.2,)"\nordering="BEFORE"\nside="BOTH"\n[[dependencies.wizards]]\nmodId="runes"\nmandatory=true\nversionRange="[1.3.2,)"\nordering="NONE"\nside="BOTH"\n[[dependencies.wizards]]\nmodId="structure_pool_api"\nmandatory=true\nversionRange="[1.2.1,)"\nordering="NONE"\nside="BOTH"\n''')

(out / 'PORT-PINS.txt').write_text(
    f'substrate={BASE_SHA}\n'
    f'target={TARGET_SHA}\n'
    'target_version=3.1.1\n'
    'minecraft=1.20.1\nforge=47.4.23\nyarn=1.20.1+build.10\njava=17\n'
    'architectury_loom=1.7.435\narchitectury_plugin=3.4.164\n'
    f'architectury_common_project_id={architectury_ids["common"]}\n'
    f'architectury_forge_project_id={architectury_ids["forge"]}\n'
)

required = [
    out / 'common/src/main/java/net/wizards/WizardsMod.java',
    out / 'forge/src/main/java/net/wizards/forge/ForgeMod.java',
    out / 'forge/src/main/resources/META-INF/mods.toml',
]
missing = [str(p) for p in required if not p.exists()]
if missing:
    raise SystemExit('Wizards preparation lost required source: ' + ', '.join(missing))
if (out / 'fabric').exists() or (out / 'neoforge').exists():
    raise SystemExit('non-Forge platform directory leaked into generated Wizards port')
print('Wizards 3.1.1 target reconstructed for native Forge 1.20.1 compatibility work')
print(f'substrate={BASE_SHA}')
print(f'target={TARGET_SHA}')
