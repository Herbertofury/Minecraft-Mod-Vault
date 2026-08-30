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
# userdev classpath does not reliably expand nested JarJar dependencies from a
# local file dependency. Expose the *certified dependency's own embedded* wrapper
# + core only AFTER Gazebos source and release bytes are sealed and immediately
# before the QA-only semantic server. Both are modRuntimeOnly so Loom hands both
# jars to Forge's userdev mod/game-library discovery instead of leaving the core
# only on Gradle's ordinary runtime classpath behind the ModLauncher module layer.
# A later prepare() reconstructs pristine product sources before native-client
# and fresh packaged-server lanes.
python3 - "$BASE" <<'PY'
from pathlib import Path
import sys
p=Path(sys.argv[1]); s=p.read_text()
needle='gradle --no-daemon --stacktrace -p "$GEN" :forge:classes\nSERVER_SMOKE="$PORT/gazebos-server-semantics.log"; : > "$SERVER_SMOKE"\n'
bridge='''# QA-only userdev completion for Structure Pool API's certified JarJar dependency.\nunzip -p "$SPA_JAR" META-INF/jars/mixinextras-forge-0.4.1.jar > "$GEN/libs/mixinextras-forge-qa.jar"\nunzip -tq "$GEN/libs/mixinextras-forge-qa.jar" >/dev/null\nunzip -p "$GEN/libs/mixinextras-forge-qa.jar" META-INF/jars/MixinExtras-0.4.1.jar > "$GEN/libs/mixinextras-core-qa.jar"\nunzip -tq "$GEN/libs/mixinextras-core-qa.jar" >/dev/null\npython3 - "$GEN/forge/build.gradle" <<'PYQA'\nfrom pathlib import Path\nimport sys\np=Path(sys.argv[1]); text=p.read_text()\ndep='    modImplementation files("$rootDir/libs/tiny_config-forge.jar")\\n'\nextra='    modRuntimeOnly files("$rootDir/libs/mixinextras-forge-qa.jar")\\n    modRuntimeOnly files("$rootDir/libs/mixinextras-core-qa.jar")\\n'\nif text.count(dep) != 1:\n    raise SystemExit(f'[Gazebos v4] expected one Forge TinyConfig dependency seam, found {text.count(dep)}')\np.write_text(text.replace(dep, dep + extra))\nPYQA\necho "[Gazebos] USERDEV_MIXINEXTRAS_QA_RUNTIME_PASS outer=$(sha256sum "$GEN/libs/mixinextras-forge-qa.jar" | awk '{print $1}') core=$(sha256sum "$GEN/libs/mixinextras-core-qa.jar" | awk '{print $1}')"\n\ngradle --no-daemon --stacktrace -p "$GEN" :forge:classes\nSERVER_SMOKE="$PORT/gazebos-server-semantics.log"; : > "$SERVER_SMOKE"\n'''
if s.count(needle) != 1:
    raise SystemExit(f'[Gazebos v4] expected exactly one post-seal semantic compile seam, found {s.count(needle)}')
p.write_text(s.replace(needle, bridge))
PY

bash -n "$BASE"
echo '[Gazebos v4] POST_SEAL_CERTIFIED_SPA_MIXINEXTRAS_USERDEV_MODULE_BRIDGE_PASS'
exec bash "$V3"
