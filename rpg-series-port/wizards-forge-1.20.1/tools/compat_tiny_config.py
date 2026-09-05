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
# JarJar dependencies, but downstream Architectury Loom dev launches do not reliably recreate nested
# service visibility for files(...). Recreate only MixinExtras dev-runtime visibility. TinyConfig 3.1.0
# is already a direct Wizards modImplementation; prove Spell Engine embeds the exact same certified JAR
# instead of loading a duplicate or reviving the obsolete TinyConfig 2.x runtime bridge.
marker = '// LOCAL_SPELL_ENGINE_JIJ_DEV_RUNTIME_BRIDGE'
if marker not in s:
    bridge = r'''

// LOCAL_SPELL_ENGINE_JIJ_DEV_RUNTIME_BRIDGE
// Mirrors packaged Forge JarJar service visibility for a local Spell Engine dependency in Loom dev runs.
def spellEngineLocalJar = file('../libs/spell-engine-forge.jar')
def certifiedTinyConfigLocalJar = file('../libs/tiny-config-forge.jar')
def spellEngineDevRuntime = file('../libs/dev-runtime')
def stageSpellEngineNestedRuntime = tasks.register('stageSpellEngineNestedRuntime') {
    inputs.file spellEngineLocalJar
    inputs.file certifiedTinyConfigLocalJar
    outputs.dir spellEngineDevRuntime
    doLast {
        delete spellEngineDevRuntime
        spellEngineDevRuntime.mkdirs()

        // Stage nested service JARs from either Forge JarJar or the historical nested-JAR directory.
        copy {
            from zipTree(spellEngineLocalJar)
            include 'META-INF/jarjar/*.jar'
            include 'META-INF/jars/*.jar'
            into spellEngineDevRuntime
            eachFile { details -> details.path = details.name }
            includeEmptyDirs = false
        }
        def firstLevel = fileTree(spellEngineDevRuntime).matching { include '*.jar' }.files.toList()
        firstLevel.each { nestedJar ->
            copy {
                from zipTree(nestedJar)
                include 'META-INF/jarjar/*.jar'
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

        // The sealed 1.10.4 release must advertise exactly one TinyConfig JarJar entry. Resolve its
        // payload path from Forge JarJar metadata, then prove those bytes are the same certified 3.1.0
        // JAR Wizards already loads directly. Do not add a second TinyConfig runtime dependency.
        def runtimeTinyConfig = file("${spellEngineDevRuntime}/runtime-tinyconfig.jar")
        def spellZip = new java.util.zip.ZipFile(spellEngineLocalJar)
        try {
            def metadataEntry = spellZip.getEntry('META-INF/jarjar/metadata.json')
            if (metadataEntry == null) {
                throw new GradleException('Spell Engine release lost META-INF/jarjar/metadata.json')
            }
            def metadata = new groovy.json.JsonSlurper().parse(spellZip.getInputStream(metadataEntry))
            def tinyEntries = metadata.jars.findAll { jar ->
                def artifact = jar.identifier?.artifact?.toString()?.toLowerCase(java.util.Locale.ROOT) ?: ''
                artifact.contains('tiny') && artifact.contains('config')
            }
            if (tinyEntries.size() != 1) {
                throw new GradleException("Spell Engine JarJar metadata must contain exactly one TinyConfig entry, found ${tinyEntries.size()}")
            }
            def tinyPath = tinyEntries[0].path?.toString()
            if (tinyPath == null || tinyPath.isBlank()) {
                throw new GradleException('Spell Engine TinyConfig JarJar metadata entry has no payload path')
            }
            def tinyEntry = spellZip.getEntry(tinyPath)
            if (tinyEntry == null) {
                throw new GradleException("Spell Engine JarJar metadata points to missing TinyConfig payload: ${tinyPath}")
            }
            runtimeTinyConfig.bytes = spellZip.getInputStream(tinyEntry).bytes
        } finally {
            spellZip.close()
        }

        def sha256 = { File f ->
            java.security.MessageDigest.getInstance('SHA-256').digest(f.bytes).encodeHex().toString()
        }
        def nestedTinySha = sha256(runtimeTinyConfig)
        def directTinySha = sha256(certifiedTinyConfigLocalJar)
        if (nestedTinySha != directTinySha) {
            throw new GradleException("Spell Engine nested TinyConfig differs from certified Wizards runtime: ${nestedTinySha} != ${directTinySha}")
        }

        def tinyZip = new java.util.zip.ZipFile(runtimeTinyConfig)
        try {
            def managerPath = 'net/tiny_config/ConfigManager.class'
            if (tinyZip.getEntry(managerPath) == null) {
                throw new GradleException("Certified TinyConfig 3.1.0 payload missing ${managerPath}")
            }
            def modsEntry = tinyZip.getEntry('META-INF/mods.toml')
            if (modsEntry == null) {
                throw new GradleException('Certified TinyConfig 3.1.0 payload missing META-INF/mods.toml')
            }
            def modsText = tinyZip.getInputStream(modsEntry).getText('UTF-8')
            if (!modsText.contains('modId="tiny_config"')) {
                throw new GradleException('Certified TinyConfig 3.1.0 payload lost modId="tiny_config"')
            }
        } finally {
            tinyZip.close()
        }
        println "[Wizards] Spell Engine local JIJ dev-runtime gate green: MixinExtras Operation staged; certified TinyConfig 3.1.0 JarJar identity=${nestedTinySha}."
    }
}

dependencies {
    // Keep only the Forge MixinExtras wrapper as a staged runtime service. TinyConfig 3.1.0 is already
    // present through modImplementation files('../libs/tiny-config-forge.jar') and is identity-checked above.
    runtimeOnly fileTree(dir: '../libs/dev-runtime', include: ['mixinextras-forge-*.jar'])
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
    "modImplementation files('../libs/tiny-config-forge.jar')",
    "tasks.matching { it.name in ['runServer', 'runClient'] }",
    'stageSpellEngineNestedRuntime',
    'META-INF/jarjar/metadata.json',
    'Spell Engine TinyConfig JarJar metadata entry has no payload path',
    'runtime-tinyconfig.jar',
    'net/tiny_config/ConfigManager.class',
    'Spell Engine nested TinyConfig differs from certified Wizards runtime',
    'com/llamalad7/mixinextras/injector/wrapoperation/Operation.class',
):
    if required not in final_forge:
        raise SystemExit(f'Wizards local Spell Engine JIJ bridge missing: {required}')
for forbidden in (
    "runtimeOnly fileTree(dir: '../libs/dev-runtime', include: ['*.jar'])",
    'TinyConfig-2.3.2.jar',
    'net/tinyconfig/ConfigManager.class',
    'com.github.ZsoltMolnarrr:TinyConfig:2.3.2',
    'mixinextras-forge:$rootProject.mixinextras_version',
    'forgeRuntimeLibrary spellEngineTinyConfig',
    "forgeRuntimeLibrary files('../libs/dev-runtime/",
):
    if forbidden in final_forge:
        raise SystemExit(f'unresolvable/redundant/stale dev-runtime bridge survived: {forbidden}')

print('Wizards TinyConfig compatibility layer applied: TinyConfig 3.1.0 foundation + certified Spell Engine JarJar identity + MixinExtras-only dev-runtime parity')
