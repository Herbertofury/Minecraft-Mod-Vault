#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 2:
    raise SystemExit('usage: compat_tiny_config.py <generated-wizards-root>')
root = Path(sys.argv[1]).resolve()

props = root / 'gradle.properties'
s = props.read_text()
lines = s.splitlines()
old_version = 'tiny_config_version=2.3.2'
new_version = 'tiny_config_version=3.1.0+1.20.1'
if old_version in lines:
    lines[lines.index(old_version)] = new_version
elif new_version not in lines:
    raise SystemExit('expected Wizards TinyConfig version property not found')
props.write_text('\n'.join(lines) + '\n')

common = root / 'common/build.gradle'
s = common.read_text()
old = '    modImplementation "maven.modrinth:tiny-config:$rootProject.tiny_config_version-${platform}"'
new = "    modImplementation files('../libs/tiny-config-common.jar')"
if old in s:
    s = s.replace(old, new, 1)
elif new not in s:
    raise SystemExit('expected Wizards common TinyConfig dependency coordinate not found')
common.write_text(s)

forge = root / 'forge/build.gradle'
s = forge.read_text()
old = '    modImplementation "maven.modrinth:tiny-config:$rootProject.tiny_config_version-forge"'
new = "    modImplementation files('../libs/tiny-config-forge.jar')"
if old in s:
    s = s.replace(old, new, 1)
elif new not in s:
    raise SystemExit('expected Wizards Forge TinyConfig dependency coordinate not found')

# Spell Engine is intentionally supplied as a local Forge mod JAR. Packaged Forge expands that mod's
# META-INF/jars dependencies, but downstream Architectury Loom dev launches do not reliably recreate
# the nested runtime classpath for files(...). Recreate only that dev-runtime visibility from the exact
# local Spell Engine release bytes. MixinExtras remains a runtime mod so its own JIJ metadata activates;
# legacy TinyConfig is a Forge runtime library so its classes are visible across mod module classloaders.
marker = '// LOCAL_SPELL_ENGINE_JIJ_DEV_RUNTIME_BRIDGE'
if marker not in s:
    bridge = r'''

// LOCAL_SPELL_ENGINE_JIJ_DEV_RUNTIME_BRIDGE
// Mirrors packaged Forge JAR-in-JAR visibility for a local Spell Engine dependency in Loom dev runs.
def spellEngineLocalJar = file('../libs/spell-engine-forge.jar')
def spellEngineDevRuntime = file('../libs/dev-runtime')
def stageSpellEngineNestedRuntime = tasks.register('stageSpellEngineNestedRuntime') {
    inputs.file spellEngineLocalJar
    outputs.dir spellEngineDevRuntime
    doLast {
        delete spellEngineDevRuntime
        spellEngineDevRuntime.mkdirs()
        copy {
            from zipTree(spellEngineLocalJar)
            include 'META-INF/jars/*.jar'
            into spellEngineDevRuntime
            eachFile { details -> details.path = details.name }
            includeEmptyDirs = false
        }
        def firstLevel = fileTree(spellEngineDevRuntime).matching { include '*.jar' }.files.toList()
        firstLevel.each { nestedJar ->
            copy {
                from zipTree(nestedJar)
                include 'META-INF/jars/*.jar'
                into spellEngineDevRuntime
                eachFile { details -> details.path = details.name }
                includeEmptyDirs = false
            }
        }
        def operationPath = 'com/llamalad7/mixinextras/injector/wrapoperation/Operation.class'
        def operationVisible = fileTree(spellEngineDevRuntime).matching { include '*.jar' }.files.any { nestedJar ->
            !zipTree(nestedJar).matching { include operationPath }.isEmpty()
        }
        if (!operationVisible) {
            throw new GradleException("Spell Engine nested MixinExtras runtime missing ${operationPath}")
        }
        def legacyTinyConfig = file("${spellEngineDevRuntime}/TinyConfig-2.3.2.jar")
        def legacyConfigManager = 'net/tinyconfig/ConfigManager.class'
        if (!legacyTinyConfig.exists() || zipTree(legacyTinyConfig).matching { include legacyConfigManager }.isEmpty()) {
            throw new GradleException("Spell Engine nested TinyConfig 2.3.2 missing ${legacyConfigManager}")
        }
        println '[Wizards] Spell Engine local JIJ dev-runtime gate green: MixinExtras Operation + legacy TinyConfig ConfigManager staged from exact release bytes.'
    }
}

dependencies {
    // Keep the Forge wrapper as a runtime mod; Forge then activates its nested MixinExtras service JAR.
    runtimeOnly fileTree(dir: '../libs/dev-runtime', include: ['mixinextras-forge-*.jar'])
    // TinyConfig 2.3.2 is an implementation library used by Spell Power/Spell Engine, not a downstream mod.
    forgeRuntimeLibrary files('../libs/dev-runtime/TinyConfig-2.3.2.jar')
}

tasks.matching { it.name in ['runServer', 'runClient'] }.configureEach {
    dependsOn stageSpellEngineNestedRuntime
}
'''
    s = s.rstrip() + bridge + '\n'
forge.write_text(s)

mods = root / 'forge/src/main/resources/META-INF/mods.toml'
s = mods.read_text()
if 'modId="tiny_config"' not in s:
    marker_dep = '[[dependencies.wizards]]\nmodId="spell_engine"'
    dependency = '''[[dependencies.wizards]]\nmodId="tiny_config"\nmandatory=true\nversionRange="[3.1.0,)"\nordering="NONE"\nside="BOTH"\n'''
    if marker_dep not in s:
        raise SystemExit('Wizards mods.toml dependency insertion point not found')
    s = s.replace(marker_dep, dependency + marker_dep)
    mods.write_text(s)

final_forge = forge.read_text()
for required in (
    '// LOCAL_SPELL_ENGINE_JIJ_DEV_RUNTIME_BRIDGE',
    "runtimeOnly fileTree(dir: '../libs/dev-runtime', include: ['mixinextras-forge-*.jar'])",
    "forgeRuntimeLibrary files('../libs/dev-runtime/TinyConfig-2.3.2.jar')",
    "tasks.matching { it.name in ['runServer', 'runClient'] }",
    'stageSpellEngineNestedRuntime',
    'com/llamalad7/mixinextras/injector/wrapoperation/Operation.class',
    'net/tinyconfig/ConfigManager.class',
):
    if required not in final_forge:
        raise SystemExit(f'Wizards local Spell Engine JIJ bridge missing: {required}')
for forbidden in (
    "runtimeOnly fileTree(dir: '../libs/dev-runtime', include: ['*.jar'])",
    'com.github.ZsoltMolnarrr:TinyConfig:2.3.2',
    'mixinextras-forge:$rootProject.mixinextras_version',
    'forgeRuntimeLibrary spellEngineTinyConfig',
):
    if forbidden in final_forge:
        raise SystemExit(f'unresolvable/redundant dev-runtime bridge survived: {forbidden}')

print('Wizards TinyConfig compatibility layer applied: TinyConfig 3.1.0 foundation + classloader-correct local Spell Engine JIJ dev-runtime parity')
