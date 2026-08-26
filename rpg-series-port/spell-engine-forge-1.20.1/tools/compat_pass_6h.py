#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 3:
    raise SystemExit('usage: compat_pass_6h.py <generated-port-root> <spell-engine-1.20.1-baseline>')

root = Path(sys.argv[1]).resolve()
forge_build = root / 'forge/build.gradle'
s = forge_build.read_text()

# Spell Power and Ranged Weapon API are separately installed RPG Series mods. Compile the common
# module against their common JARs (pass 1), and put their real Forge JARs on this loader module's
# compile/runtime classpath. Do NOT `include` or shadow either dependency into Spell Engine.
if 'SPELL_POWER_FORGE_JAR' not in s:
    s += r'''

def requireExternalModJar = { String envName ->
 def raw = System.getenv(envName)
 if (raw == null || raw.isBlank()) {
  throw new GradleException("Missing required external Forge mod JAR environment variable: ${envName}")
 }
 def jarFile = file(raw)
 if (!jarFile.isFile()) {
  throw new GradleException("External Forge mod JAR does not exist for ${envName}: ${jarFile}")
 }
 return jarFile
}

dependencies {
 modImplementation files(requireExternalModJar('SPELL_POWER_FORGE_JAR'))
 modImplementation files(requireExternalModJar('RANGED_FORGE_JAR'))
}
'''
forge_build.write_text(s)

final = forge_build.read_text()
for required in ('SPELL_POWER_FORGE_JAR', 'RANGED_FORGE_JAR', 'modImplementation files(requireExternalModJar'):
    if required not in final:
        raise SystemExit(f'pass6h missing external Forge dependency wiring: {required}')
for forbidden in (
    "include files(requireExternalModJar('SPELL_POWER_FORGE_JAR'))",
    "include files(requireExternalModJar('RANGED_FORGE_JAR'))",
    'SPELL_POWER_SOURCE_DIRS',
    'RANGED_SOURCE_DIRS',
):
    if forbidden in final:
        raise SystemExit(f'pass6h would embed or source-inject a separate dependency: {forbidden}')

print('Spell Engine compatibility pass 6h applied: real external Spell Power + Ranged Forge mod JARs')
