#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 3:
    raise SystemExit('usage: compat_pass_6p.py <generated-port-root> <spell-engine-1.20.1-baseline>')

root = Path(sys.argv[1]).resolve()
forge_build = root / 'forge/build.gradle'
s = forge_build.read_text()

# Loom's include(...) requires a module component with capabilities; a raw files(...) dependency has
# neither. Keep the exact certified TinyConfig JAR bytes, but stage them into an artifact-only local
# Maven layout during Gradle configuration and resolve that coordinate as a real module component.
old = '''    def tinyConfig = implementation(files(tinyConfigJar))
    include tinyConfig
    forgeRuntimeLibrary files(tinyConfigJar)
'''
new = '''    def tinyConfigVersion = "3.1.0+1.20.1"
    def tinyConfigGroup = "net.tiny_config.certified"
    def tinyConfigArtifact = "tiny_config-forge"
    def tinyConfigRepo = file("${buildDir}/certified-tiny-config-maven")
    def tinyConfigModuleDir = new File(tinyConfigRepo,
            "${tinyConfigGroup.replace('.', '/')}/${tinyConfigArtifact}/${tinyConfigVersion}")
    tinyConfigModuleDir.mkdirs()
    def tinyConfigModuleJar = new File(tinyConfigModuleDir,
            "${tinyConfigArtifact}-${tinyConfigVersion}.jar")
    if (!tinyConfigModuleJar.exists() || !java.util.Arrays.equals(
            java.nio.file.Files.readAllBytes(tinyConfigModuleJar.toPath()),
            java.nio.file.Files.readAllBytes(tinyConfigJar.toPath()))) {
        java.nio.file.Files.copy(tinyConfigJar.toPath(), tinyConfigModuleJar.toPath(),
                java.nio.file.StandardCopyOption.REPLACE_EXISTING)
    }
    project.repositories.maven {
        url = uri(tinyConfigRepo)
        metadataSources { artifact() }
    }
    def tinyConfigCoordinate = "${tinyConfigGroup}:${tinyConfigArtifact}:${tinyConfigVersion}"
    def tinyConfig = implementation(tinyConfigCoordinate)
    include tinyConfig
    forgeRuntimeLibrary tinyConfigCoordinate
'''
if s.count(old) != 1:
    raise SystemExit(f'expected exactly one raw-file TinyConfig nesting seam, found {s.count(old)}')
s = s.replace(old, new, 1)

# Update pass-6g's own fail-closed generated-build assertions so they describe the module-backed
# solution rather than the rejected raw files(...) attempt.
old_required = '''    'def tinyConfig = implementation(files(tinyConfigJar))',
    'include tinyConfig',
    'forgeRuntimeLibrary files(tinyConfigJar)',
'''
new_required = '''    'def tinyConfigCoordinate = "${tinyConfigGroup}:${tinyConfigArtifact}:${tinyConfigVersion}"',
    'def tinyConfig = implementation(tinyConfigCoordinate)',
    'include tinyConfig',
    'forgeRuntimeLibrary tinyConfigCoordinate',
    'metadataSources { artifact() }',
'''
if s.count(old_required) != 1:
    raise SystemExit(f'expected one pass6g TinyConfig assertion seam, found {s.count(old_required)}')
s = s.replace(old_required, new_required, 1)

forge_build.write_text(s)
final = forge_build.read_text()
for required in (
    'net.tiny_config.certified',
    'certified-tiny-config-maven',
    'metadataSources { artifact() }',
    'implementation(tinyConfigCoordinate)',
    'include tinyConfig',
    'forgeRuntimeLibrary tinyConfigCoordinate',
    'StandardCopyOption.REPLACE_EXISTING',
):
    if required not in final:
        raise SystemExit(f'certified TinyConfig module staging missing: {required}')
for forbidden in (
    'implementation(files(tinyConfigJar))',
    'forgeRuntimeLibrary files(tinyConfigJar)',
    'com.github.ZsoltMolnarrr:TinyConfig',
):
    if forbidden in final:
        raise SystemExit(f'rejected TinyConfig dependency form survived module staging: {forbidden}')

print('Spell Engine compatibility pass 6p applied: certified TinyConfig staged as byte-identical artifact-only local Maven module for Loom nesting')
