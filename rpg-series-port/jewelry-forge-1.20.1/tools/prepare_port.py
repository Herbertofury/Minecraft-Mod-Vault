#!/usr/bin/env python3
from pathlib import Path
import shutil
import sys

if len(sys.argv) != 4:
    raise SystemExit("usage: prepare_port.py <jewelry-1.20.1-substrate> <jewelry-2.4.0-target> <output>")

base = Path(sys.argv[1]).resolve()
target = Path(sys.argv[2]).resolve()
out = Path(sys.argv[3]).resolve()

BASE_SHA = "f20b7d94c4c6cdd5a4ed26e4066374b64654fb96"
TARGET_SHA = "572cb8759d13075b97e7a1acd969a6203db594cb"

for repo, expected, label in ((base, BASE_SHA, "substrate"), (target, TARGET_SHA, "target")):
    head = __import__("subprocess").check_output(["git", "-C", str(repo), "rev-parse", "HEAD"], text=True).strip()
    if head != expected:
        raise SystemExit(f"Jewelry {label} pin mismatch: expected {expected}, got {head}")

if out.exists():
    shutil.rmtree(out)
shutil.copytree(target, out, ignore=shutil.ignore_patterns(".git", ".gradle", "build", "runs"))

shutil.rmtree(out / "fabric", ignore_errors=True)
if (out / "forge").exists():
    shutil.rmtree(out / "forge")
(out / "neoforge").rename(out / "forge")

(out / "settings.gradle").write_text(r'''pluginManagement {
    repositories {
        maven { url = 'https://maven.fabricmc.net/' }
        maven { url = 'https://maven.architectury.dev/' }
        maven { url = 'https://maven.minecraftforge.net/' }
        gradlePluginPortal()
    }
}
rootProject.name = 'jewelry-forge-1.20.1'
include 'common'
include 'forge'
''')

(out / "gradle.properties").write_text(r'''org.gradle.jvmargs=-Xmx3G
org.gradle.parallel=true
minecraft_version=1.20.1
yarn_mappings=1.20.1+build.10
forge_version=1.20.1-47.4.23
fabric_loader_version=0.16.14
mod_version=2.4.0+1.20.1
maven_group=net.jewelry
archives_name=jewelry
curios_version=5.14.1+1.20.1
tiny_config_version=2.3.2
''')

(out / "build.gradle").write_text(r'''plugins {
    id 'dev.architectury.loom' version '1.7.+' apply false
    id 'architectury-plugin' version '3.4.+'
    id 'com.github.johnrengelman.shadow' version '8.1.1' apply false
}
architectury { minecraft = project.minecraft_version }
allprojects { group = rootProject.maven_group; version = rootProject.mod_version }
subprojects {
    apply plugin: 'dev.architectury.loom'
    apply plugin: 'architectury-plugin'
    apply plugin: 'maven-publish'
    base { archivesName = "$rootProject.archives_name-$project.name" }
    repositories {
        mavenCentral()
        maven { url = 'https://maven.architectury.dev/' }
        maven { url = 'https://maven.minecraftforge.net/' }
        maven { url = 'https://maven.fabricmc.net/' }
        maven { url = 'https://jitpack.io' }
        maven { url = 'https://api.modrinth.com/maven' }
        maven { url = 'https://maven.theillusivec4.top/' }
    }
    dependencies {
        minecraft "com.mojang:minecraft:$rootProject.minecraft_version"
        mappings "net.fabricmc:yarn:$rootProject.yarn_mappings:v2"
    }
    java {
        toolchain { languageVersion = JavaLanguageVersion.of(17) }
        withSourcesJar()
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }
    tasks.withType(JavaCompile).configureEach { options.encoding='UTF-8'; options.release=17; options.compilerArgs += ['-Xmaxerrs','2000'] }
}
''')

(out / "common/build.gradle").write_text(r'''architectury { common 'forge' }
def generatedResources = file('src/main/generated')
sourceSets { main { resources.srcDir generatedResources } }
repositories {
    maven { url = 'https://jitpack.io' }
    maven { url = 'https://api.modrinth.com/maven' }
    maven { url = 'https://maven.fabricmc.net/' }
}
dependencies {
    modImplementation "net.fabricmc:fabric-loader:$rootProject.fabric_loader_version"
    implementation "com.github.ZsoltMolnarrr:TinyConfig:$rootProject.tiny_config_version"
    compileOnly files(rootProject.file('libs/structure-pool-api.jar'))
    compileOnly files(rootProject.file('libs/spell-power.jar'))
    compileOnly files(rootProject.file('libs/ranged-weapon-api.jar'))
}
''')

# Match upstream Architectury's multi-loader layout: platform selection belongs to the platform
# subproject. Keeping this out of the root properties leaves :common in common-mode while :forge
# enables Loom's Forge dependency/runtime configurations.
(out / "forge/gradle.properties").write_text("loom.platform = forge\n")

(out / "forge/build.gradle").write_text(r'''plugins { id 'com.github.johnrengelman.shadow' }
architectury { platformSetupLoomIde(); forge() }
loom { forge { } }
configurations {
    common { canBeResolved=true; canBeConsumed=false }
    compileClasspath.extendsFrom common
    runtimeClasspath.extendsFrom common
    developmentForge.extendsFrom common
    shadowBundle { canBeResolved=true; canBeConsumed=false }
}
repositories {
    maven { url = 'https://jitpack.io' }
    maven { url = 'https://api.modrinth.com/maven' }
    maven { url = 'https://maven.theillusivec4.top/' }
}
dependencies {
    forge "net.minecraftforge:forge:$rootProject.forge_version"
    common(project(path: ':common', configuration: 'namedElements')) { transitive=false }
    shadowBundle project(path: ':common', configuration: 'transformProductionForge')

    // Reuse the exact TinyConfig 2.3.2 packaging already proven by the Spell Engine Forge 1.20.1
    // release: compile/runtime library plus JarJar include from JitPack.
    def tinyConfig = implementation("com.github.ZsoltMolnarrr:TinyConfig:$rootProject.tiny_config_version")
    include tinyConfig
    forgeRuntimeLibrary tinyConfig

    modImplementation files(rootProject.file('libs/structure-pool-api.jar'))
    modImplementation files(rootProject.file('libs/spell-power.jar'))
    modImplementation files(rootProject.file('libs/ranged-weapon-api.jar'))
    modImplementation "top.theillusivec4.curios:curios-forge:$rootProject.curios_version"
}
processResources {
    inputs.property 'version', project.version
    filesMatching('META-INF/mods.toml') { expand(version: project.version) }
}
shadowJar { configurations=[project.configurations.shadowBundle]; archiveClassifier='dev-shadow' }
remapJar { inputFile.set(shadowJar.archiveFile); dependsOn(shadowJar); archiveClassifier='' }
''')

java_root = out / "forge/src/main/java"
for path in list(java_root.rglob("*.java")):
    s = path.read_text()
    s = s.replace("net.jewelry.neoforge", "net.jewelry.forge")
    s = s.replace("net.neoforged.bus.api.IEventBus", "net.minecraftforge.eventbus.api.IEventBus")
    s = s.replace("net.neoforged.fml.common.Mod", "net.minecraftforge.fml.common.Mod")
    s = s.replace("net.neoforged.neoforge.common.NeoForge", "net.minecraftforge.common.MinecraftForge")
    s = s.replace("net.neoforged.neoforge.event.BuildCreativeModeTabContentsEvent", "net.minecraftforge.event.BuildCreativeModeTabContentsEvent")
    s = s.replace("net.neoforged.neoforge.event.village.VillagerTradesEvent", "net.minecraftforge.event.village.VillagerTradesEvent")
    s = s.replace("net.neoforged.neoforge.registries.RegisterEvent", "net.minecraftforge.registries.RegisterEvent")
    s = s.replace("NeoForge.EVENT_BUS", "MinecraftForge.EVENT_BUS")
    path.write_text(s)

old_pkg = java_root / "net/jewelry/neoforge"
new_pkg = java_root / "net/jewelry/forge"
if old_pkg.exists():
    new_pkg.parent.mkdir(parents=True, exist_ok=True)
    if new_pkg.exists():
        shutil.rmtree(new_pkg)
    old_pkg.rename(new_pkg)

forge_mod = new_pkg / "NeoForgeMod.java"
if forge_mod.exists():
    s = forge_mod.read_text().replace("public final class NeoForgeMod", "public final class ForgeMod")
    s = s.replace("public NeoForgeMod(IEventBus modBus) {", "public ForgeMod() {\n        IEventBus modBus = net.minecraftforge.fml.javafmlmod.FMLJavaModLoadingContext.get().getModEventBus();")
    s = s.replace("NeoForgeMod::", "ForgeMod::")
    s = s.replace("// Ore world-gen injection is data-driven on NeoForge", "// Ore world-gen injection is data-driven on Forge")
    new_file = new_pkg / "ForgeMod.java"
    new_file.write_text(s)
    forge_mod.unlink()

platform_impl = new_pkg / "PlatformImpl.java"
platform_impl.write_text(r'''package net.jewelry.forge;

import net.jewelry.Platform;
import net.minecraftforge.fml.loading.LoadingModList;

public class PlatformImpl {
    public static class ForgeUtil implements Platform.Util {
        @Override public boolean isModLoaded(String modid) {
            return LoadingModList.get().getModFileById(modid) != null;
        }
    }
    private static final Platform.Util UTIL = new ForgeUtil();
    public static Platform.Util util() { return UTIL; }
}
''')

neo_data = out / "common/src/main/resources/data/jewelry/neoforge"
forge_data = out / "common/src/main/resources/data/jewelry/forge"
if neo_data.exists():
    forge_data.parent.mkdir(parents=True, exist_ok=True)
    if forge_data.exists():
        shutil.rmtree(forge_data)
    neo_data.rename(forge_data)

meta = out / "forge/src/main/resources/META-INF"
meta.mkdir(parents=True, exist_ok=True)
for stale in (meta / "neoforge.mods.toml", meta / "neoforge.mods.toml.bak"):
    if stale.exists():
        stale.unlink()
(meta / "mods.toml").write_text(r'''modLoader="javafml"
loaderVersion="[47,)"
license="All Rights Reserved"
[[mods]]
modId="jewelry"
version="${version}"
displayName="Jewelry"
description="Native Forge 1.20.1 backport of Jewelry 2.4.0."
[[dependencies.jewelry]]
modId="forge"
mandatory=true
versionRange="[47.4,48)"
ordering="NONE"
side="BOTH"
[[dependencies.jewelry]]
modId="minecraft"
mandatory=true
versionRange="[1.20.1,1.20.2)"
ordering="NONE"
side="BOTH"
[[dependencies.jewelry]]
modId="spell_power"
mandatory=true
versionRange="[1.6,2)"
ordering="AFTER"
side="BOTH"
[[dependencies.jewelry]]
modId="ranged_weapon_api"
mandatory=true
versionRange="[2.3,3)"
ordering="AFTER"
side="BOTH"
[[dependencies.jewelry]]
modId="structure_pool_api"
mandatory=true
versionRange="[1.2,2)"
ordering="AFTER"
side="BOTH"
[[dependencies.jewelry]]
modId="curios"
mandatory=true
versionRange="[5.14,6)"
ordering="AFTER"
side="BOTH"
''')

(out / "PORT-PINS.txt").write_text(f"substrate={BASE_SHA}\ntarget={TARGET_SHA}\n")

required = [
    out / "common/src/main/java/net/jewelry/items/JewelryItems.java",
    out / "common/src/main/generated/assets/jewelry/models/item/diamond_ring.json",
    out / "common/src/main/resources/data/jewelry",
    out / "forge/src/main/java/net/jewelry/forge/ForgeMod.java",
    out / "forge/src/main/resources/META-INF/mods.toml",
    out / "forge/gradle.properties",
]
missing = [str(p) for p in required if not p.exists()]
if missing:
    raise SystemExit("Jewelry preparation lost required current content/platform state: " + ", ".join(missing))
if "loom.platform = forge" not in (out / "forge/gradle.properties").read_text():
    raise SystemExit("Jewelry Forge Loom platform flag missing")
for build in (out / "common/build.gradle", out / "forge/build.gradle"):
    text = build.read_text()
    if "com.github.ZsoltMolnarrr:TinyConfig" not in text:
        raise SystemExit(f"Jewelry TinyConfig proven coordinate missing from {build}")
if "forgeRuntimeLibrary tinyConfig" not in (out / "forge/build.gradle").read_text():
    raise SystemExit("Jewelry TinyConfig Forge runtime/embed wiring missing")

print("Jewelry 2.4.0 full-target Forge 1.20.1 preparation complete")
print(f"substrate={BASE_SHA}")
print(f"target={TARGET_SHA}")
