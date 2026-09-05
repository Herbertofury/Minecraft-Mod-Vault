#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 2:
    raise SystemExit('usage: patch_forge_production_client_library_directory.py <production-client-helper>')

helper = Path(sys.argv[1]).resolve()
if not helper.is_file():
    raise SystemExit(f'production client helper missing: {helper}')

s = helper.read_text()

# Forge's installed production profile uses ${library_directory} in both the fixed
# BootstrapLauncher module path and -DlibraryDirectory. Preserve the exact installed
# library tree rather than synthesizing a development/userdev path.
old_values = '''        "assets_root": str(mc_home / "assets"),
        "assets_index_name": str(asset_index),
        "auth_uuid": "00000000000000000000000000000001",
'''
new_values = '''        "assets_root": str(mc_home / "assets"),
        "assets_index_name": str(asset_index),
        # Forge 47's installed version profile uses this in both -DlibraryDirectory and its
        # module-path expression. Point it at the exact Minecraft home populated by --installClient.
        "library_directory": str(libraries),
        "auth_uuid": "00000000000000000000000000000001",
'''
if s.count(old_values) != 1:
    raise SystemExit(f'[More RPG 2.7.2] expected one production launcher values seam, found {s.count(old_values)}')
if '"library_directory": str(libraries),' in s:
    raise SystemExit('[More RPG 2.7.2] production launcher library_directory patch unexpectedly already present')
s = s.replace(old_values, new_values, 1)

# A real Forge 47 production profile intentionally ignores ${version_name}.jar in
# BootstrapLauncher. The old helper instead appended versions/1.20.1/1.20.1.jar while
# version_name was 1.20.1-forge-47.4.23, so the ignore rule could not match it. BSL then
# modularized the raw Mojang jar as automatic module _1._20._1 alongside Forge's SRG
# minecraft module, producing the exact split-package ResolutionException seen in run 370.
# Stage the already SHA-verified Mojang client bytes under the Forge profile filename so
# the profile's own ignore contract applies, while keeping those bytes on the legacy
# launcher classpath exactly as a production launcher does.
old_version_jar = '''    vanilla_client = mc_home / "versions" / args.minecraft / f"{args.minecraft}.jar"
    if not vanilla_client.is_file():
        die("vanilla client JAR missing")
    classpath.append(vanilla_client)
'''
new_version_jar = '''    vanilla_client = mc_home / "versions" / args.minecraft / f"{args.minecraft}.jar"
    if not vanilla_client.is_file():
        die("vanilla client JAR missing")
    forge_version_jar = mc_home / "versions" / args.forge_version_id / f"{args.forge_version_id}.jar"
    forge_version_jar.parent.mkdir(parents=True, exist_ok=True)
    vanilla_sha1 = sha1_file(vanilla_client)
    if not forge_version_jar.is_file() or sha1_file(forge_version_jar) != vanilla_sha1:
        shutil.copy2(vanilla_client, forge_version_jar)
    if sha1_file(forge_version_jar) != vanilla_sha1:
        die("Forge profile version JAR is not byte-identical to verified Mojang client")
    classpath.append(forge_version_jar)
'''
if s.count(old_version_jar) != 1:
    raise SystemExit(f'[More RPG 2.7.2] expected one raw vanilla version-jar classpath seam, found {s.count(old_version_jar)}')
if 'forge_version_jar = mc_home / "versions" / args.forge_version_id' in s:
    raise SystemExit('[More RPG 2.7.2] Forge profile version-jar alias patch unexpectedly already present')
s = s.replace(old_version_jar, new_version_jar, 1)

# Fail closed on the Forge profile contract that makes the alias safe. If a future Forge
# profile stops ignoring ${version_name}.jar, do not silently reintroduce the duplicate
# Minecraft module; stop before launching instead.
old_substitute = '''    jvm = substitute(jvm, values)
    game = substitute(game, values)
    game += ["--width", "1280", "--height", "720", "--quickPlaySingleplayer", args.quick_play]
'''
new_substitute = '''    jvm = substitute(jvm, values)
    game = substitute(game, values)
    expected_version_jar_name = f"{args.forge_version_id}.jar"
    ignore_arg = next((arg for arg in jvm if arg.startswith("-DignoreList=")), None)
    if ignore_arg is None:
        die("Forge production profile has no BootstrapLauncher ignoreList")
    ignored_names = ignore_arg.split("=", 1)[1].split(",")
    if expected_version_jar_name not in ignored_names:
        die(f"Forge production profile does not ignore version JAR {expected_version_jar_name}")
    if vanilla_client in classpath:
        die("raw vanilla version JAR leaked onto Forge production classpath")
    if forge_version_jar not in classpath:
        die("Forge profile version JAR alias missing from production classpath")
    print(f"[More RPG 2.7.2] FORGE_PROFILE_VERSION_JAR_ALIAS_PASS name={expected_version_jar_name} sha1={vanilla_sha1} bootstrap_ignored=true")
    game += ["--width", "1280", "--height", "720", "--quickPlaySingleplayer", args.quick_play]
'''
if s.count(old_substitute) != 1:
    raise SystemExit(f'[More RPG 2.7.2] expected one JVM substitution seam, found {s.count(old_substitute)}')
s = s.replace(old_substitute, new_substitute, 1)

contracts = {
    '"library_directory": str(libraries),': 1,
    'libraries = mc_home / "libraries"': 2,
    'forge_version_jar = mc_home / "versions" / args.forge_version_id / f"{args.forge_version_id}.jar"': 1,
    'classpath.append(forge_version_jar)': 1,
    'classpath.append(vanilla_client)': 0,
    'expected_version_jar_name = f"{args.forge_version_id}.jar"': 1,
    'bootstrap_ignored=true': 1,
    'jvm = substitute(jvm, values)': 1,
    'game = substitute(game, values)': 1,
    'die(f"unresolved launcher placeholder ${{{key}}} in {arg!r}")': 1,
}
for needle, expected in contracts.items():
    actual = s.count(needle)
    if actual != expected:
        raise SystemExit(f'[More RPG 2.7.2] production launcher contract drifted for {needle!r}: expected {expected}, found {actual}')
helper.write_text(s)
print('[More RPG 2.7.2] FORGE_PRODUCTION_LIBRARY_DIRECTORY_PLACEHOLDER_PATCHED source=installed-forge-profile value=mc_home/libraries')
print('[More RPG 2.7.2] FORGE_PRODUCTION_VERSION_JAR_CLASSPATH_PATCHED source=run-370-split-package owner=bootstrap-ignore-contract')
