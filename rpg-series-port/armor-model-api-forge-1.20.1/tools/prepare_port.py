#!/usr/bin/env python3
from pathlib import Path
import shutil, subprocess, sys

if len(sys.argv) != 3:
    raise SystemExit('usage: prepare_port.py <armor-model-api-target> <output>')

src = Path(sys.argv[1]).resolve()
out = Path(sys.argv[2]).resolve()
TARGET_SHA = 'a664155a0aab3161cd7e4bf0c1f72512b4ec4949'
head = subprocess.check_output(['git','-C',str(src),'rev-parse','HEAD'], text=True).strip()
if head != TARGET_SHA:
    raise SystemExit(f'Armor Model API target pin mismatch: expected {TARGET_SHA}, got {head}')

if out.exists(): shutil.rmtree(out)
shutil.copytree(src, out, ignore=shutil.ignore_patterns('.git','.gradle','build','run','runs'))
for name in ('fabric','example'):
    shutil.rmtree(out / name, ignore_errors=True)
for name in ('example-common','example-fabric','example-neoforge'):
    shutil.rmtree(out / name, ignore_errors=True)
if (out / 'forge').exists(): shutil.rmtree(out / 'forge')
(out / 'neoforge').rename(out / 'forge')

(out/'settings.gradle').write_text('''pluginManagement {\n    repositories {\n        maven { url = 'https://maven.fabricmc.net/' }\n        maven { url = 'https://maven.architectury.dev/' }\n        maven { url = 'https://maven.minecraftforge.net/' }\n        gradlePluginPortal()\n    }\n}\nrootProject.name = 'armor-model-api-forge-1.20.1'\ninclude 'common'\ninclude 'forge'\n''')
(out/'gradle.properties').write_text('''org.gradle.jvmargs=-Xmx3G\norg.gradle.parallel=true\nminecraft_version=1.20.1\nyarn_mappings=1.20.1+build.10\nforge_version=1.20.1-47.4.23\nfabric_loader_version=0.16.14\nmod_version=1.0.0+1.20.1\nmaven_group=net.rpg_foundation\narchives_name=armor_model_api\nenabled_platforms=forge\n''')
(out/'build.gradle').write_text('''plugins {\n    id 'dev.architectury.loom' version '1.7.+' apply false\n    id 'architectury-plugin' version '3.4.+'\n    id 'com.github.johnrengelman.shadow' version '8.1.1' apply false\n}\narchitectury { minecraft = project.minecraft_version }\nallprojects { group = rootProject.maven_group; version = rootProject.mod_version }\nsubprojects {\n    apply plugin: 'dev.architectury.loom'\n    apply plugin: 'architectury-plugin'\n    apply plugin: 'maven-publish'\n    base { archivesName = "$rootProject.archives_name-$project.name" }\n    repositories {\n        mavenCentral()\n        maven { url = 'https://maven.fabricmc.net/' }\n        maven { url = 'https://maven.architectury.dev/' }\n        maven { url = 'https://maven.minecraftforge.net/' }\n    }\n    dependencies {\n        minecraft "com.mojang:minecraft:$rootProject.minecraft_version"\n        mappings "net.fabricmc:yarn:$rootProject.yarn_mappings:v2"\n    }\n    java {\n        toolchain { languageVersion = JavaLanguageVersion.of(17) }\n        withSourcesJar()\n        sourceCompatibility = JavaVersion.VERSION_17\n        targetCompatibility = JavaVersion.VERSION_17\n    }\n    tasks.withType(JavaCompile).configureEach { options.encoding='UTF-8'; options.release=17; options.compilerArgs += ['-Xmaxerrs','2000'] }\n}\n''')
(out/'common/build.gradle').write_text('''architectury { common 'forge' }\nloom { accessWidenerPath = file('src/main/resources/armor_model_api.accesswidener') }\ndependencies {\n    modImplementation "net.fabricmc:fabric-loader:$rootProject.fabric_loader_version"\n}\n''')
(out/'forge/gradle.properties').write_text('loom.platform = forge\n')
(out/'forge/build.gradle').write_text('''plugins { id 'com.github.johnrengelman.shadow' }\narchitectury { platformSetupLoomIde(); forge() }\nloom {\n    accessWidenerPath = project(':common').loom.accessWidenerPath\n    forge { mixinConfig 'armor_model_api.mixins.json' }\n}\nconfigurations {\n    common { canBeResolved=true; canBeConsumed=false }\n    compileClasspath.extendsFrom common\n    runtimeClasspath.extendsFrom common\n    developmentForge.extendsFrom common\n    shadowBundle { canBeResolved=true; canBeConsumed=false }\n}\ndependencies {\n    forge "net.minecraftforge:forge:$rootProject.forge_version"\n    common(project(path: ':common', configuration: 'namedElements')) { transitive=false }\n    shadowBundle project(path: ':common', configuration: 'transformProductionForge')\n}\nprocessResources {\n    inputs.property 'version', project.version\n    filesMatching('META-INF/mods.toml') { expand(version: project.version) }\n}\nshadowJar { configurations=[project.configurations.shadowBundle]; archiveClassifier='dev-shadow' }\nremapJar { inputFile.set(shadowJar.archiveFile); dependsOn(shadowJar); archiveClassifier='' }\n''')

java_root = out/'forge/src/main/java'
for p in java_root.rglob('*.java'):
    s=p.read_text()
    s=s.replace('net.rpg_foundation.armor_api.neoforge','net.rpg_foundation.armor_api.forge')
    s=s.replace('net.neoforged.api.distmarker.Dist','net.minecraftforge.api.distmarker.Dist')
    s=s.replace('net.neoforged.bus.api.IEventBus','net.minecraftforge.eventbus.api.IEventBus')
    s=s.replace('net.neoforged.fml.common.Mod','net.minecraftforge.fml.common.Mod')
    s=s.replace('net.neoforged.fml.loading.FMLEnvironment','net.minecraftforge.fml.loading.FMLEnvironment')
    s=s.replace('net.neoforged.neoforge.client.event.RegisterClientReloadListenersEvent','net.minecraftforge.client.event.RegisterClientReloadListenersEvent')
    p.write_text(s)
old=java_root/'net/rpg_foundation/armor_api/neoforge'
new=java_root/'net/rpg_foundation/armor_api/forge'
if old.exists():
    new.parent.mkdir(parents=True,exist_ok=True)
    old.rename(new)
entry=new/'ArmorModelApiNeoForge.java'
if entry.exists():
    s=entry.read_text().replace('public final class ArmorModelApiNeoForge','public final class ArmorModelApiForge')
    s=s.replace('public ArmorModelApiNeoForge(IEventBus modBus) {','public ArmorModelApiForge() {\n        IEventBus modBus = net.minecraftforge.fml.javafmlmod.FMLJavaModLoadingContext.get().getModEventBus();')
    (new/'ArmorModelApiForge.java').write_text(s)
    entry.unlink()

meta=out/'forge/src/main/resources/META-INF'
meta.mkdir(parents=True,exist_ok=True)
for p in meta.glob('*neoforge*'): p.unlink()
(meta/'mods.toml').write_text('''modLoader="javafml"\nloaderVersion="[47,)"\nlicense="MIT"\n[[mods]]\nmodId="armor_model_api"\nversion="${version}"\ndisplayName="Armor Model API"\nauthors="Daedelus"\nlogoFile="assets/armor_model_api/icon.png"\ndescription="Renders Bedrock/GeckoLib geo armor models through the vanilla armor pipeline."\n[[dependencies.armor_model_api]]\nmodId="forge"\nmandatory=true\nversionRange="[47.4,48)"\nordering="NONE"\nside="BOTH"\n[[dependencies.armor_model_api]]\nmodId="minecraft"\nmandatory=true\nversionRange="[1.20.1,1.20.2)"\nordering="NONE"\nside="BOTH"\n''')
(out/'PORT-PINS.txt').write_text(f'target={TARGET_SHA}\nsource_branch=1.21.1\nsource_version=1.0.0\n')

required=[
 out/'common/src/main/java/net/rpg_foundation/armor_api/client/GeoArmorRenderer.java',
 out/'common/src/main/java/net/rpg_foundation/armor_api/client/geo/GeoBaker.java',
 out/'common/src/main/java/net/rpg_foundation/armor_api/client/layer/TrimLayer.java',
 out/'forge/src/main/java/net/rpg_foundation/armor_api/forge/ArmorModelApiForge.java',
 out/'forge/src/main/resources/META-INF/mods.toml',
]
missing=[str(p) for p in required if not p.exists()]
if missing: raise SystemExit('Armor Model API preparation lost required source: '+', '.join(missing))
print('Armor Model API 1.0.0 target prepared for native Forge 1.20.1 compatibility work')
print('target='+TARGET_SHA)
