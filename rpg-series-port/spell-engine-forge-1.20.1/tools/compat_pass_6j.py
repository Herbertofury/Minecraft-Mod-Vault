#!/usr/bin/env python3
from pathlib import Path
import sys
import zipfile

if len(sys.argv) != 3:
    raise SystemExit('usage: compat_pass_6j.py <generated-port-root> <spell-engine-1.20.1-baseline>')

root = Path(sys.argv[1]).resolve()
common_java = root / 'common/src/main/java'
common_build = root / 'common/build.gradle'
forge_build = root / 'forge/build.gradle'
mods = root / 'forge/src/main/resources/META-INF/mods.toml'
common_tiny = root / 'libs/tiny-config-common.jar'
forge_tiny = root / 'libs/tiny-config-forge.jar'

for jar in (common_tiny, forge_tiny):
    if not jar.is_file():
        raise SystemExit(f'TinyConfig 3.1.0 foundation JAR missing: {jar}')
    with zipfile.ZipFile(jar) as zf:
        bad = zf.testzip()
        if bad:
            raise SystemExit(f'corrupt TinyConfig foundation JAR {jar}: {bad}')

with zipfile.ZipFile(common_tiny) as zf:
    required_class = 'net/tiny_config/versioning/VersionableConfig.class'
    if required_class not in zf.namelist():
        raise SystemExit(f'TinyConfig common foundation missing {required_class}')
with zipfile.ZipFile(forge_tiny) as zf:
    try:
        metadata = zf.read('META-INF/mods.toml').decode('utf-8')
    except KeyError as exc:
        raise SystemExit('TinyConfig Forge foundation missing META-INF/mods.toml') from exc
    if 'modId="tiny_config"' not in metadata:
        raise SystemExit('TinyConfig Forge foundation has wrong mod identity')

# The 1.10.2 source model intentionally extends TinyConfig 3.x VersionableConfig so downstream
# content mods can independently version their own summoned-entity JSON files. The early 1.20.1
# compatibility bootstrap stripped this inheritance only because TinyConfig 3.x had not yet been
# ported. Restore the authoritative API now that the native 3.1.0 foundation exists.
config = common_java / 'net/spell_engine/api/spell/summon/SummonedEntityConfig.java'
s = config.read_text()
if 'import net.tiny_config.versioning.VersionableConfig;' not in s:
    anchor = 'import net.minecraft.util.Identifier;\n'
    if anchor not in s:
        raise SystemExit('SummonedEntityConfig Identifier import anchor missing')
    s = s.replace(anchor, anchor + 'import net.tiny_config.versioning.VersionableConfig;\n', 1)
if 'public class SummonedEntityConfig {' in s:
    s = s.replace('public class SummonedEntityConfig {',
                  'public class SummonedEntityConfig extends VersionableConfig {', 1)
elif 'public class SummonedEntityConfig extends VersionableConfig {' not in s:
    raise SystemExit('SummonedEntityConfig class declaration no longer matches expected source')
config.write_text(s)

# Common code still uses the legacy net.tinyconfig package for older internal configs. Compile the
# restored public VersionableConfig surface against the real TinyConfig 3.1.0 common foundation without
# injecting its classes into Spell Engine.
cb = common_build.read_text()
modern_common_dep = "    compileOnly files('../libs/tiny-config-common.jar')\n"
if modern_common_dep not in cb:
    anchor = '    implementation "com.github.ZsoltMolnarrr:TinyConfig:$rootProject.tiny_config_version"\n'
    if anchor not in cb:
        raise SystemExit('legacy TinyConfig common dependency anchor missing')
    cb = cb.replace(anchor, anchor + modern_common_dep, 1)
common_build.write_text(cb)

# Dev/runtime must expose the same separate TinyConfig 3.1.0 Forge mod. Do not include/shadow it into
# Spell Engine; packaged users receive it as an explicit dependency just like modern upstream.
fb = forge_build.read_text()
modern_forge_dep = "    modImplementation files('../libs/tiny-config-forge.jar')\n"
if modern_forge_dep not in fb:
    anchor = '    modImplementation "me.shedaniel.cloth:cloth-config-forge:$rootProject.cloth_config_version"\n'
    if anchor not in fb:
        raise SystemExit('Forge dependency anchor missing for TinyConfig 3.1.0 API')
    fb = fb.replace(anchor, anchor + modern_forge_dep, 1)
forge_build.write_text(fb)

mt = mods.read_text()
if 'modId="tiny_config"' not in mt:
    anchor = '''[[dependencies.spell_engine]]\nmodId="minecraft"\nmandatory=true\nversionRange="[1.20.1,1.20.2)"\nordering="NONE"\nside="BOTH"\n'''
    dep = '''\n[[dependencies.spell_engine]]\nmodId="tiny_config"\nmandatory=true\nversionRange="[3.1.0,)"\nordering="NONE"\nside="BOTH"\n'''
    if anchor not in mt:
        raise SystemExit('Spell Engine minecraft dependency anchor missing')
    mt = mt.replace(anchor, anchor + dep, 1)
mods.write_text(mt)

# Hard guards: preserve API parity, keep TinyConfig 3.x external, and retain the legacy implementation
# only for the older internal ConfigManager callsites until that migration is proven independently.
final_config = config.read_text()
for required in (
    'import net.tiny_config.versioning.VersionableConfig;',
    'public class SummonedEntityConfig extends VersionableConfig {',
):
    if required not in final_config:
        raise SystemExit(f'Spell Engine summoned-config versioning parity missing: {required}')
final_common = common_build.read_text()
if modern_common_dep.strip() not in final_common:
    raise SystemExit('TinyConfig 3.1.0 common API dependency missing')
if 'com.github.ZsoltMolnarrr:TinyConfig:$rootProject.tiny_config_version' not in final_common:
    raise SystemExit('legacy TinyConfig internal dependency unexpectedly removed')
final_forge = forge_build.read_text()
if modern_forge_dep.strip() not in final_forge:
    raise SystemExit('TinyConfig 3.1.0 Forge runtime dependency missing')
if 'include files(\'../libs/tiny-config-forge.jar\')' in final_forge or 'shadowBundle files(\'../libs/tiny-config-forge.jar\')' in final_forge:
    raise SystemExit('TinyConfig 3.1.0 foundation must remain a separate mod dependency')
if 'modId="tiny_config"' not in mods.read_text() or 'versionRange="[3.1.0,)"' not in mods.read_text():
    raise SystemExit('Spell Engine TinyConfig 3.1.0 dependency metadata missing')

print('Spell Engine compatibility pass 6j applied: restored TinyConfig 3.1.0 VersionableConfig API parity for summoned configs')
