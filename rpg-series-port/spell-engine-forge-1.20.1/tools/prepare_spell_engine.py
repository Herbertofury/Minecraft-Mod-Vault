#!/usr/bin/env python3
from pathlib import Path
import argparse, shutil

p=argparse.ArgumentParser()
p.add_argument('base')
p.add_argument('target')
p.add_argument('out')
a=p.parse_args()
base=Path(a.base).resolve(); target=Path(a.target).resolve(); out=Path(a.out).resolve()
if out.exists(): shutil.rmtree(out)
(out/'common/src/main').mkdir(parents=True)
shutil.copytree(target/'common/src/main/java', out/'common/src/main/java')
if (target/'common/src/main/resources').exists(): shutil.copytree(target/'common/src/main/resources', out/'common/src/main/resources')
if (target/'common/src/main/generated').exists(): shutil.copytree(target/'common/src/main/generated', out/'common/src/main/generated')
for name in ['LICENCE','LICENSE','README.md','CHANGELOG.md']:
    src=target/name
    if src.exists(): shutil.copy2(src, out/name)

# TinyConfig 3.x changed the package used by this source line. Keep the established 1.20.1 library.
for f in (out/'common/src/main/java').rglob('*.java'):
    s=f.read_text()
    s=s.replace('import net.tiny_config.ConfigManager;', 'import net.tinyconfig.ConfigManager;')
    f.write_text(s)
sc=out/'common/src/main/java/net/spell_engine/api/spell/summon/SummonedEntityConfig.java'
if sc.exists():
    s=sc.read_text().replace('import net.tiny_config.versioning.VersionableConfig;\n','').replace('public class SummonedEntityConfig extends VersionableConfig {','public class SummonedEntityConfig {')
    sc.write_text(s)

# Optional integrations are isolated until the core 1.10.2 common contract compiles. They are restored
# against exact 1.20.1 APIs afterward so optional mods cannot hide Minecraft-version failures.
stubs={
'net/spell_engine/compat/MeleeCompat.java': '''package net.spell_engine.compat;\nimport net.minecraft.entity.Entity;\nimport net.minecraft.entity.player.PlayerEntity;\nimport java.util.function.Function;\npublic class MeleeCompat {\n public record Attack(boolean isCombo, boolean isOffhand) { public static final Attack EMPTY=new Attack(false,false); }\n public static Function<PlayerEntity,Attack> attackProperties=p->new Attack(p.getLastAttackTime()==(p.getAttackCooldownProgressPerTick()*20),false);\n public static Function<Entity,Boolean> isEntityHostileVehicle=e->false;\n public static void init() {}\n}\n''',
'net/spell_engine/compat/CombatRollCompat.java': '''package net.spell_engine.compat;\nimport net.minecraft.entity.player.PlayerEntity;\nimport java.util.function.Function;\npublic class CombatRollCompat { public static Function<PlayerEntity,Boolean> isRolling=p->false; public static void init() {} }\n''',
'net/spell_engine/compat/FTBTeamsCompat.java': '''package net.spell_engine.compat; public class FTBTeamsCompat { public static void init() {} }\n''',
'net/spell_engine/client/compatibility/ShoulderSurfingCompatibility.java': '''package net.spell_engine.client.compatibility; public class ShoulderSurfingCompatibility { }\n''',
'net/spell_engine/client/compatibility/IrisCompatibility.java': '''package net.spell_engine.client.compatibility; import net.minecraft.client.render.RenderLayer; public class IrisCompatibility { public static void markAsDecal(RenderLayer layer) {} }\n''',
'net/spell_engine/client/compatibility/ShaderCompatibility.java': '''package net.spell_engine.client.compatibility; public class ShaderCompatibility { static void initialize() {} public static boolean isShaderPackInUse(){return false;} public static boolean isVanillaRenderSystem(){return true;} }\n''',
'net/spell_engine/compat/container/CustomBundleCompat.java': '''package net.spell_engine.compat.container; public class CustomBundleCompat { public static void init() {} }\n''',
'net/spell_engine/compat/CriticalStrikeCompat.java': '''package net.spell_engine.compat; import net.minecraft.entity.damage.DamageSource; public class CriticalStrikeCompat { public static void init(){} public static boolean isCriticalStrike(DamageSource s){return false;} public static void setCriticalStrike(DamageSource s,float m){} }\n'''
}
for rel,txt in stubs.items():
    f=out/'common/src/main/java'/rel
    f.parent.mkdir(parents=True,exist_ok=True)
    f.write_text(txt)

(out/'settings.gradle').write_text('''pluginManagement {\n repositories {\n  gradlePluginPortal()\n  maven { url = 'https://maven.fabricmc.net/' }\n  maven { url = 'https://maven.architectury.dev/' }\n  maven { url = 'https://maven.minecraftforge.net/' }\n }\n}\nrootProject.name='spell-engine-forge-1.20.1'\ninclude 'common'\n''')
(out/'gradle.properties').write_text('''org.gradle.jvmargs=-Xmx5G\norg.gradle.parallel=true\norg.gradle.caching=true\nminecraft_version=1.20.1\nyarn_mappings=1.20.1+build.10\nforge_version=1.20.1-47.4.23\nmaven_group=net.spell_engine\narchives_name=spell_engine\nmod_version=1.10.2+1.20.1\nenabled_platforms=forge\ncloth_config_version=11.1.106\nplayer_anim_version=1.0.2+1.20\ntiny_config_version=2.3.2\nmixinextras_version=0.4.1\n''')
(out/'build.gradle').write_text('''plugins {\n id 'dev.architectury.loom' version '1.7.+' apply false\n id 'architectury-plugin' version '3.4.+'\n}\narchitectury { minecraft = project.minecraft_version }\nallprojects { group=rootProject.maven_group; version=rootProject.mod_version }\nsubprojects {\n apply plugin:'dev.architectury.loom'\n apply plugin:'architectury-plugin'\n apply plugin:'maven-publish'\n base { archivesName = "$rootProject.archives_name-$project.name" }\n repositories {\n  mavenCentral()\n  maven { url='https://maven.architectury.dev/' }\n  maven { url='https://maven.minecraftforge.net/' }\n  maven { url='https://maven.fabricmc.net/' }\n  maven { url='https://maven.kosmx.dev/' }\n  maven { url='https://maven.shedaniel.me/' }\n  maven { url='https://jitpack.io' }\n  maven { name='Modrinth'; url='https://api.modrinth.com/maven'; content { includeGroup 'maven.modrinth' } }\n  maven { url='https://repo.spongepowered.org/repository/maven-public/' }\n }\n dependencies {\n  minecraft "com.mojang:minecraft:$rootProject.minecraft_version"\n  mappings "net.fabricmc:yarn:$rootProject.yarn_mappings:v2"\n }\n java { toolchain { languageVersion=JavaLanguageVersion.of(17) }; withSourcesJar(); sourceCompatibility=JavaVersion.VERSION_17; targetCompatibility=JavaVersion.VERSION_17 }\n tasks.withType(JavaCompile).configureEach { options.encoding='UTF-8'; options.release=17 }\n}\n''')
(out/'common/build.gradle').write_text('''architectury { common(rootProject.enabled_platforms.split(',')) }\nsourceSets { main { resources.srcDir 'src/main/generated' } }\nrepositories {\n maven { url='https://maven.kosmx.dev/' }\n maven { url='https://maven.shedaniel.me/' }\n maven { url='https://jitpack.io' }\n}\ndependencies {\n compileOnly 'org.spongepowered:mixin:0.8.5'\n annotationProcessor 'org.spongepowered:mixin:0.8.5:processor'\n implementation "io.github.llamalad7:mixinextras-common:$rootProject.mixinextras_version"\n annotationProcessor "io.github.llamalad7:mixinextras-common:$rootProject.mixinextras_version"\n modCompileOnly 'net.fabricmc:fabric-loader:0.15.7'\n implementation "com.github.ZsoltMolnarrr:TinyConfig:$rootProject.tiny_config_version"\n modImplementation "dev.kosmx.player-anim:player-animation-lib-fabric:$rootProject.player_anim_version"\n modImplementation "me.shedaniel.cloth:cloth-config-fabric:$rootProject.cloth_config_version"\n compileOnly fileTree(dir: rootProject.file('../spell_power-forge-1.20.1/common/build/libs'), include: ['*.jar'], exclude: ['*sources*'])\n compileOnly fileTree(dir: rootProject.file('../ranged-weapon-api-forge-1.20.1/common/build/libs'), include: ['*.jar'], exclude: ['*sources*'])\n}\n''')
(out/'UPSTREAM-PINS.txt').write_text('base=8721120169ddefd230fc73fc7c332318a92f6c7c\ntarget=bc02f7a49da950503010020da491f6bdc5871df7\n')
print(out)
