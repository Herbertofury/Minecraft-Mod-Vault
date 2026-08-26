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

# Spell Power 1.6.0's modern API represents these schools with RegistryEntry<EntityAttribute>, which
# means Spell Power does not own their backing attributes/effects. The 1.20.1 compatibility API uses
# SpellSchool.Manage for the same ownership distinction. Preserve it explicitly: without this bridge,
# Forge sees vanilla GENERIC_ATTACK_DAMAGE as INTERNAL and attempts to register that same object again
# as spell_power:physical_melee during the attribute RegisterEvent.
external_schools = root / 'common/src/main/java/net/spell_engine/api/spell/ExternalSpellSchools.java'
es = external_schools.read_text()
anchor = '    private static boolean initialized = false;\n'
ownership_bridge = '''    static {\n        // Backport Spell Power 1.6.0 RegistryEntry ownership semantics. These schools borrow\n        // attributes owned by Minecraft or Ranged Weapon API and must never register them anew.\n        for (var school : new SpellSchool[]{PHYSICAL_MELEE, PHYSICAL_MELEE_DUAL, PHYSICAL_RANGED, DEFENSE, HEALTH}) {\n            school.attributeManagement = SpellSchool.Manage.EXTERNAL;\n            school.powerEffectManagement = SpellSchool.Manage.EXTERNAL;\n        }\n    }\n\n'''
if 'school.attributeManagement = SpellSchool.Manage.EXTERNAL;' not in es:
    if anchor not in es:
        raise SystemExit('ExternalSpellSchools initialization anchor missing')
    es = es.replace(anchor, ownership_bridge + anchor, 1)
external_schools.write_text(es)

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

final_schools = external_schools.read_text()
for school in ('PHYSICAL_MELEE', 'PHYSICAL_MELEE_DUAL', 'PHYSICAL_RANGED', 'DEFENSE', 'HEALTH'):
    if school not in final_schools:
        raise SystemExit(f'pass6h external spell school missing: {school}')
for required in (
    'new SpellSchool[]{PHYSICAL_MELEE, PHYSICAL_MELEE_DUAL, PHYSICAL_RANGED, DEFENSE, HEALTH}',
    'school.attributeManagement = SpellSchool.Manage.EXTERNAL;',
    'school.powerEffectManagement = SpellSchool.Manage.EXTERNAL;',
):
    if required not in final_schools:
        raise SystemExit(f'pass6h missing external Spell Power ownership bridge: {required}')
if final_schools.count('school.attributeManagement = SpellSchool.Manage.EXTERNAL;') != 1:
    raise SystemExit('pass6h produced ambiguous duplicate external-attribute ownership bridge')

print('Spell Engine compatibility pass 6h applied: real external RPG dependency JARs + external Spell Power school ownership')
