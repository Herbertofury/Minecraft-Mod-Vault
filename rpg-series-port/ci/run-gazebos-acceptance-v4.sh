#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
V2="$ROOT/rpg-series-port/ci/run-gazebos-acceptance-v2.sh"
V3="$ROOT/rpg-series-port/ci/run-gazebos-acceptance-v3.sh"

test -f "$V2"
test -f "$V3"

# Structure Pool API correctly ships MixinExtras through Forge JarJar. Loom's
# userdev classpath does not reliably expand nested JarJar dependencies from a
# local file dependency, so expose the *certified dependency's own embedded*
# wrapper + core only to the generated QA runtime. This changes no release JAR.
python3 - "$V2" <<'PY'
from pathlib import Path
import sys
p=Path(sys.argv[1]); s=p.read_text()
old='''  cp -f "$SPA_JAR" "$GEN/libs/structure_pool_api-forge.jar"\n  cp -f "$TINY_JAR" "$GEN/libs/tiny_config-forge.jar"\n}\n'''
new='''  cp -f "$SPA_JAR" "$GEN/libs/structure_pool_api-forge.jar"\n  cp -f "$TINY_JAR" "$GEN/libs/tiny_config-forge.jar"\n  unzip -p "$SPA_JAR" META-INF/jars/mixinextras-forge-0.4.1.jar > "$GEN/libs/mixinextras-forge-qa.jar"\n  unzip -tq "$GEN/libs/mixinextras-forge-qa.jar" >/dev/null\n  unzip -p "$GEN/libs/mixinextras-forge-qa.jar" META-INF/jars/MixinExtras-0.4.1.jar > "$GEN/libs/mixinextras-core-qa.jar"\n  unzip -tq "$GEN/libs/mixinextras-core-qa.jar" >/dev/null\n  python3 - "$GEN/forge/build.gradle" <<'PYQA'\nfrom pathlib import Path\nimport sys\np=Path(sys.argv[1]); text=p.read_text()\nneedle='    modImplementation files("$rootDir/libs/tiny_config-forge.jar")\\n'\naddition='    modRuntimeOnly files("$rootDir/libs/mixinextras-forge-qa.jar")\\n    runtimeOnly files("$rootDir/libs/mixinextras-core-qa.jar")\\n'\nif text.count(needle) != 1:\n    raise SystemExit(f'[Gazebos v4] expected one Forge TinyConfig dependency seam, found {text.count(needle)}')\np.write_text(text.replace(needle, needle + addition))\nPYQA\n  echo "[Gazebos] USERDEV_MIXINEXTRAS_QA_RUNTIME_PASS outer=$(sha256sum "$GEN/libs/mixinextras-forge-qa.jar" | awk '{print $1}') core=$(sha256sum "$GEN/libs/mixinextras-core-qa.jar" | awk '{print $1}')"\n}\n'''
if s.count(old) != 1:
    raise SystemExit(f'[Gazebos v4] expected exactly one prepare dependency-copy seam, found {s.count(old)}')
p.write_text(s.replace(old,new))
PY

bash -n "$V2"
echo '[Gazebos v4] CERTIFIED_SPA_MIXINEXTRAS_USERDEV_BRIDGE_PASS'
exec bash "$V3"
