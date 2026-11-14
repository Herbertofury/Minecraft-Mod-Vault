#!/usr/bin/env python3
from pathlib import Path
import re
import shutil
import sys

if len(sys.argv) != 4:
    raise SystemExit('usage: prepare_more_rpg_library.py <modern-2.7.2-root> <old-1.20.1-root> <output-root>')
modern = Path(sys.argv[1]).resolve()
old = Path(sys.argv[2]).resolve()
out = Path(sys.argv[3]).resolve()
for p in (modern, old):
    if not p.is_dir():
        raise SystemExit(f'missing authority tree: {p}')
if out.exists():
    shutil.rmtree(out)
shutil.copytree(modern, out)
for loader in ('fabric', 'neoforge'):
    shutil.rmtree(out / loader, ignore_errors=True)

(out / 'settings.gradle').write_text(r'''pluginManagement {
    repositories {
        maven { url = 'https://maven.fabricmc.net/' }
        maven { url = 'https://maven.architectury.dev/' }
        maven { url = 'https://maven.minecraftforge.net/' }
        gradlePluginPortal()
    }
}
rootProject.name = 'more_rpg_library_forge_1201'
include 'common'
include 'forge'
''')

(out / 'build.gradle').write_text(r'''plugins {
    id 'dev.architectury.loom' version '1.7-SNAPSHOT' apply false
    id 'architectury-plugin' version '3.4-SNAPSHOT'
    id 'com.github.johnrengelman.shadow' version '8.1.1' apply false
}
architectury { minecraft = project.minecraft_version }
allprojects {
    group = rootProject.maven_group
    version = rootProject.mod_version + "+${rootProject.minecraft_version}"
}
subprojects {
    apply plugin: 'dev.architectury.loom'
    apply plugin: 'architectury-plugin'
    apply plugin: 'maven-publish'
    base { archivesName = "$rootProject.archives_name-$project.name" }
    repositories {
        maven { url 'https://maven.kosmx.dev/' }
        maven { url 'https://maven.shedaniel.me/' }
        maven { url 'https://maven.architectury.dev/' }
        maven { url 'https://maven.minecraftforge.net/' }
        mavenCentral()
    }
    dependencies {
        minecraft "com.mojang:minecraft:$rootProject.minecraft_version"
        mappings "net.fabricmc:yarn:$rootProject.yarn_mappings:v2"
    }
    java {
        withSourcesJar()
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }
    tasks.withType(JavaCompile).configureEach { it.options.release = 17 }
}
''')

(out / 'gradle.properties').write_text(r'''org.gradle.jvmargs=-Xmx4G
org.gradle.parallel=true
mod_version=2.7.2
maven_group=net.more_rpg_classes
archives_name=more_rpg_library
minecraft_version=1.20.1
yarn_mappings=1.20.1+build.10
forge_version=1.20.1-47.4.23
player_anim_version=1.0.2+1.19.4
cloth_config_version=11.1.106
mixinextras_version=0.4.1
''')

common = out / 'common'
common.joinpath('build.gradle').write_text(r'''architectury { common 'forge' }
repositories {
    maven { url 'https://maven.kosmx.dev/' }
    maven { url 'https://maven.shedaniel.me/' }
}
dependencies {
    // Exact graduated RPG foundation JARs are supplied by the active runner.
    modImplementation files(rootProject.property('spell_engine_forge_jar'))
    modImplementation files(rootProject.property('spell_power_forge_jar'))
    modImplementation files(rootProject.property('ranged_weapon_api_forge_jar'))
    modImplementation files(rootProject.property('tiny_config_forge_jar'))
    modImplementation("dev.kosmx.player-anim:player-animation-lib-fabric:${rootProject.player_anim_version}")
    modImplementation("me.shedaniel.cloth:cloth-config-fabric:${rootProject.cloth_config_version}") {
        exclude group: 'net.fabricmc.fabric-api'
    }
}
loom {
    accessWidenerPath = file('src/main/resources/more-rpg-classes.accesswidener')
}
''')

java = common / 'src/main/java'
if not java.is_dir():
    raise SystemExit('modern common Java tree missing')

# Loader-neutral provider seam. Forge binds these before common initialization; no Fabric Loader runtime.
platform = java / 'net/more_rpg_classes/compat/MoreRpgPlatform.java'
platform.parent.mkdir(parents=True, exist_ok=True)
platform.write_text(r'''package net.more_rpg_classes.compat;

import java.util.function.BooleanSupplier;
import java.util.function.Predicate;

public final class MoreRpgPlatform {
    public static Predicate<String> isModLoaded = id -> false;
    public static BooleanSupplier isDevelopmentEnvironment = () -> false;
    private MoreRpgPlatform() { }
}
''')

main = java / 'net/more_rpg_classes/MRPGCMod.java'
text = main.read_text()
text = text.replace('import net.fabricmc.loader.api.FabricLoader;\n', '')
if 'import net.more_rpg_classes.compat.MoreRpgPlatform;' not in text:
    text = text.replace('import net.more_rpg_classes.compat.CriticalStrikeCompat;\n', 'import net.more_rpg_classes.compat.CriticalStrikeCompat;\nimport net.more_rpg_classes.compat.MoreRpgPlatform;\n')
text = text.replace('FabricLoader.getInstance().isDevelopmentEnvironment() ||FabricLoader.getInstance().isModLoaded("armory_rpgs")',
                    'MoreRpgPlatform.isDevelopmentEnvironment.getAsBoolean() || MoreRpgPlatform.isModLoaded.test("armory_rpgs")')
main.write_text(text)

# Keep Critical Strike optional. Preserve More RPG's school-source behavior behind externally bindable
# entity stat providers so the future native Critical Strike port can attach without a hard class link.
critical = java / 'net/more_rpg_classes/compat/CriticalStrikeCompat.java'
critical.write_text(r'''package net.more_rpg_classes.compat;

import net.minecraft.entity.Entity;
import net.spell_power.api.SpellSchool;
import net.spell_power.api.SpellSchools;
import java.util.function.ToDoubleFunction;
import static net.more_rpg_classes.custom.MoreSpellSchools.*;

public final class CriticalStrikeCompat {
    public static ToDoubleFunction<Entity> criticalChance = entity -> 0.0;
    public static ToDoubleFunction<Entity> criticalDamageMultiplier = entity -> 1.0;
    private CriticalStrikeCompat() { }

    public static void init() {
        if (!MoreRpgPlatform.isModLoaded.test("critical_strike")) return;
        add(FROST_RANGED);
        add(FIRE_RANGED);
        add(RAGE_MELEE);
    }
    private static void add(SpellSchool school) {
        school.addSource(SpellSchool.Trait.CRIT_CHANCE, SpellSchool.Apply.ADD,
                query -> criticalChance.applyAsDouble(query.entity()));
        school.addSource(SpellSchool.Trait.CRIT_DAMAGE, SpellSchool.Apply.ADD,
                query -> criticalDamageMultiplier.applyAsDouble(query.entity()) - 1.0);
        SpellSchools.configureSpellCritDamage(school);
        SpellSchools.configureSpellCritChance(school);
    }
}
''')

# Modern source must not retain loader hard-links in common after the provider adaptation.
leaks = []
for f in java.rglob('*.java'):
    t = f.read_text(errors='replace')
    for needle in ('net.fabricmc.loader.api.FabricLoader', 'net.neoforged.', 'net.minecraftforge.'):
        if needle in t:
            leaks.append(f'{f.relative_to(out)}: {needle}')
if leaks:
    raise SystemExit('loader hard-links remain in modern common source:\n' + '\n'.join(leaks[:80]))

forge = out / 'forge'
forge_java = forge / 'src/main/java/net/more_rpg_classes/forge'
forge_res = forge / 'src/main/resources/META-INF'
forge_java.mkdir(parents=True, exist_ok=True)
forge_res.mkdir(parents=True, exist_ok=True)
forge.joinpath('gradle.properties').write_text('loom.platform=forge\n')
forge.joinpath('build.gradle').write_text(r'''plugins { id 'com.github.johnrengelman.shadow' }
architectury { platformSetupLoomIde(); forge() }
loom {
    accessWidenerPath = project(':common').loom.accessWidenerPath
    forge {
        convertAccessWideners = true
        extraAccessWideners.add loom.accessWidenerPath.get().asFile.name
    }
}
configurations {
    common { canBeResolved = true; canBeConsumed = false }
    compileClasspath.extendsFrom common
    runtimeClasspath.extendsFrom common
    developmentForge.extendsFrom common
    shadowCommon { canBeResolved = true; canBeConsumed = false }
}
dependencies {
    forge "net.minecraftforge:forge:$rootProject.forge_version"
    common(project(path: ':common', configuration: 'namedElements')) { transitive = false }
    shadowCommon project(path: ':common', configuration: 'transformProductionForge')
    modImplementation files(rootProject.property('spell_engine_forge_jar'))
    modImplementation files(rootProject.property('spell_power_forge_jar'))
    modImplementation files(rootProject.property('ranged_weapon_api_forge_jar'))
    modImplementation files(rootProject.property('tiny_config_forge_jar'))
    modImplementation "dev.kosmx.player-anim:player-animation-lib-forge:$rootProject.player_anim_version"
    modImplementation "me.shedaniel.cloth:cloth-config-forge:$rootProject.cloth_config_version"
    implementation(include("io.github.llamalad7:mixinextras-forge:$rootProject.mixinextras_version"))
}
processResources { inputs.property 'version', project.version; filesMatching('META-INF/mods.toml') { expand(project.properties) } }
shadowJar { configurations = [project.configurations.shadowCommon]; archiveClassifier = 'dev-shadow' }
remapJar { inputFile.set shadowJar.archiveFile; dependsOn shadowJar }
''')

forge_java.joinpath('ForgeMod.java').write_text(r'''package net.more_rpg_classes.forge;

import net.minecraftforge.fml.ModList;
import net.minecraftforge.fml.common.Mod;
import net.minecraftforge.fml.loading.FMLLoader;
import net.more_rpg_classes.MRPGCMod;
import net.more_rpg_classes.compat.MoreRpgPlatform;

@Mod(MRPGCMod.MOD_ID)
public final class ForgeMod {
    public ForgeMod() {
        MoreRpgPlatform.isModLoaded = id -> ModList.get().isLoaded(id);
        MoreRpgPlatform.isDevelopmentEnvironment = () -> !FMLLoader.isProduction();
        // Loader-neutral config/custom spell/entity-relation initialization. Registry-event adaptation
        // is installed by the next compatibility pass after the full modern common source compiles.
        MRPGCMod.init();
    }
}
''')

forge_res.joinpath('mods.toml').write_text(r'''modLoader="javafml"
loaderVersion="[47,)"
license="MIT"
[[mods]]
modId="more_rpg_classes"
version="${version}"
displayName="More RPG Library"
description="Native Forge 1.20.1 forward-port of More RPG Library 2.7.2."
[[dependencies.more_rpg_classes]]
modId="forge"
mandatory=true
versionRange="[47.4.23,)"
ordering="NONE"
side="BOTH"
[[dependencies.more_rpg_classes]]
modId="minecraft"
mandatory=true
versionRange="[1.20.1,1.20.2)"
ordering="NONE"
side="BOTH"
[[dependencies.more_rpg_classes]]
modId="spell_engine"
mandatory=true
versionRange="[1.10.4,)"
ordering="AFTER"
side="BOTH"
[[dependencies.more_rpg_classes]]
modId="spell_power"
mandatory=true
versionRange="[1.6.0,)"
ordering="AFTER"
side="BOTH"
[[dependencies.more_rpg_classes]]
modId="tiny_config"
mandatory=true
versionRange="[3.1.0,)"
ordering="AFTER"
side="BOTH"
''')
forge.joinpath('src/main/resources/pack.mcmeta').write_text('{"pack":{"pack_format":15,"description":"More RPG Library Forge 1.20.1 resources"}}\n')

# Do not let the modern NeoForge/Fabric manifests hitchhike into the native Forge release resources.
for rel in ('common/src/main/resources/fabric.mod.json', 'common/src/main/resources/META-INF/neoforge.mods.toml'):
    p = out / rel
    if p.exists(): p.unlink()

# Preserve old source only as a target-era API/reference snapshot; never compile it into the release.
ref = out / '.authority'
ref.mkdir(exist_ok=True)
shutil.copy2(old / 'src/main/java/net/more_rpg_classes/MRPGCMod.java', ref / 'MRPGCMod-1.20.1.java')

# Fail closed on Java 21 class-level syntax we already know Java 17 cannot accept.
syntax = []
for f in java.rglob('*.java'):
    t = f.read_text(errors='replace')
    if re.search(r'\bwhen\s+[^;]+->', t): syntax.append(str(f.relative_to(out)))
if syntax:
    raise SystemExit('Java 21-only syntax requires compatibility pass: ' + ', '.join(syntax[:40]))

print(f'[More RPG 2.7.2] NATIVE_FORGE_1201_FULL_COMMON_SCAFFOLD_READY java_files={sum(1 for _ in java.rglob("*.java"))}')
