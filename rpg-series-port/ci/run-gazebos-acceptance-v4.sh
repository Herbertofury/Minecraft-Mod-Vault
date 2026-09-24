#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
BASE="$ROOT/rpg-series-port/ci/run-gazebos-acceptance.sh"
V2="$ROOT/rpg-series-port/ci/run-gazebos-acceptance-v2.sh"
V3="$ROOT/rpg-series-port/ci/run-gazebos-acceptance-v3.sh"

test -f "$BASE"
test -f "$V2"
test -f "$V3"

# Structure Pool API correctly ships MixinExtras through Forge JarJar. Loom's
# userdev remapping of a local file dependency does not preserve the wrapper's
# nested JarJar relationship, and exposing wrapper/core as two local modules is
# still insufficient because ModLauncher isolates their automatic modules.
# For the QA-only userdev semantic lane, flatten the certified wrapper + its exact
# nested core into ONE temporary GAMELIBRARY jar so MixinExtrasBootstrap is in the
# same module as the Forge config plugin. This is injected only AFTER Gazebos
# source/release bytes are sealed. A later prepare() reconstructs pristine product
# sources, and the fresh packaged-server lane still proves the real untouched
# Structure Pool API JarJar relationship.
python3 - "$BASE" <<'PY'
from pathlib import Path
import sys
p=Path(sys.argv[1]); s=p.read_text()
needle='gradle --no-daemon --stacktrace -p "$GEN" :forge:classes\nSERVER_SMOKE="$PORT/gazebos-server-semantics.log"; : > "$SERVER_SMOKE"\n'
bridge='''# QA-only userdev completion for Structure Pool API's certified JarJar dependency.\nunzip -p "$SPA_JAR" META-INF/jars/mixinextras-forge-0.4.1.jar > "$GEN/libs/mixinextras-forge-qa.jar"\nunzip -tq "$GEN/libs/mixinextras-forge-qa.jar" >/dev/null\nunzip -p "$GEN/libs/mixinextras-forge-qa.jar" META-INF/jars/MixinExtras-0.4.1.jar > "$GEN/libs/mixinextras-core-qa.jar"\nunzip -tq "$GEN/libs/mixinextras-core-qa.jar" >/dev/null\npython3 - "$GEN/libs/mixinextras-forge-qa.jar" "$GEN/libs/mixinextras-core-qa.jar" "$GEN/libs/mixinextras-forge-flat-qa.jar" <<'PYFLAT'\nimport sys, zipfile\nfrom pathlib import Path\nouter, core, merged = map(Path, sys.argv[1:])\nwith zipfile.ZipFile(outer, 'r') as zo, zipfile.ZipFile(core, 'r') as zc, zipfile.ZipFile(merged, 'w') as zw:\n    seen=set()\n    for info in zo.infolist():\n        zw.writestr(info, zo.read(info.filename)); seen.add(info.filename)\n    for info in zc.infolist():\n        name=info.filename\n        if name.upper() == 'META-INF/MANIFEST.MF' or name in seen:\n            continue\n        zw.writestr(info, zc.read(name)); seen.add(name)\nwith zipfile.ZipFile(merged, 'r') as z:\n    required='com/llamalad7/mixinextras/MixinExtrasBootstrap.class'\n    if required not in z.namelist():\n        raise SystemExit('[Gazebos v4] flattened MixinExtras QA jar lacks bootstrap class')\nPYFLAT\nunzip -tq "$GEN/libs/mixinextras-forge-flat-qa.jar" >/dev/null\npython3 - "$GEN/forge/build.gradle" <<'PYQA'\nfrom pathlib import Path\nimport sys\np=Path(sys.argv[1]); text=p.read_text()\ndep='    modImplementation files("$rootDir/libs/tiny_config-forge.jar")\\n'\nextra='    modRuntimeOnly files("$rootDir/libs/mixinextras-forge-flat-qa.jar")\\n'\nif text.count(dep) != 1:\n    raise SystemExit(f'[Gazebos v4] expected one Forge TinyConfig dependency seam, found {text.count(dep)}')\np.write_text(text.replace(dep, dep + extra))\nPYQA\necho "[Gazebos] USERDEV_MIXINEXTRAS_FLAT_QA_RUNTIME_PASS merged=$(sha256sum "$GEN/libs/mixinextras-forge-flat-qa.jar" | awk '{print $1}')"\n\ngradle --no-daemon --stacktrace -p "$GEN" :forge:classes\nSERVER_SMOKE="$PORT/gazebos-server-semantics.log"; : > "$SERVER_SMOKE"\n'''
if s.count(needle) != 1:
    raise SystemExit(f'[Gazebos v4] expected exactly one post-seal semantic compile seam, found {s.count(needle)}')
p.write_text(s.replace(needle, bridge))
PY

bash -n "$BASE"
echo '[Gazebos v4] POST_SEAL_CERTIFIED_SPA_MIXINEXTRAS_FLAT_USERDEV_BRIDGE_PASS'
exec bash "$V3"
