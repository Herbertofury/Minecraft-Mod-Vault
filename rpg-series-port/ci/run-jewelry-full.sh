#!/usr/bin/env bash
set -euo pipefail

ROOT="$(pwd)"
BASE_RUNNER="$ROOT/rpg-series-port/ci/run-jewelry.sh"

# Keep the normal Jewelry runner readable while making the active acceptance lane generate its
# canonical source/package with the downstream Ranged/MixinExtras dev bridge already applied.
# The edit is CI-worktree-local; the authoritative transformation itself lives in compat_pass_6.py.
if ! grep -Fq 'compat_pass_6.py' "$BASE_RUNNER"; then
    python3 - "$BASE_RUNNER" <<'PY'
from pathlib import Path
import sys
p = Path(sys.argv[1])
s = p.read_text()
anchor = 'python3 "$TOOLS/compat_pass_5.py" "$PORT"\n'
addition = anchor + 'python3 "$TOOLS/compat_pass_6.py" "$PORT"\n'
if anchor not in s:
    raise SystemExit('Jewelry full gate could not find compatibility-pass insertion anchor')
p.write_text(s.replace(anchor, addition, 1))
PY
fi

bash "$BASE_RUNNER"

PORT="$ROOT/rpg-series-port/jewelry-forge-1.20.1/generated"
OUT_JAR="$(find "$PORT/forge/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' ! -name '*dev-shadow*' ! -name '*javadoc*' | sort | head -n1)"
SOURCE_ZIP="$ROOT/jewelry-2.4.0-forge-1.20.1-source-ci.zip"
test -f "$OUT_JAR"
test -f "$SOURCE_ZIP"

# Jewelry consumes Ranged's embedded MixinExtras at installation time. Its own release must not
# carry another nested copy; the explicit dependency added by pass 6 is development-classpath only.
if unzip -Z1 "$OUT_JAR" | grep -Eq '^META-INF/jars/mixinextras-(forge|common)-'; then
    echo '[Jewelry CI] ERROR: Jewelry release duplicated Ranged Weapon API MixinExtras.' >&2
    exit 4
fi

python3 - "$SOURCE_ZIP" <<'PY'
import sys, zipfile
with zipfile.ZipFile(sys.argv[1]) as z:
    props = z.read('gradle.properties').decode()
    build = z.read('forge/build.gradle').decode()
if 'mixinextras_version=0.4.1' not in props:
    raise SystemExit('Canonical Jewelry source ZIP missing MixinExtras dev-runtime version pin')
needle = 'implementation "io.github.llamalad7:mixinextras-forge:$rootProject.mixinextras_version"'
if needle not in build:
    raise SystemExit('Canonical Jewelry source ZIP missing downstream Ranged/MixinExtras dev bridge')
if 'include("io.github.llamalad7:mixinextras-forge' in build:
    raise SystemExit('Canonical Jewelry source ZIP would duplicate MixinExtras in release output')
print('[Jewelry CI] Canonical source ZIP carries the dev-runtime bridge without release duplication.')
PY

bash "$ROOT/rpg-series-port/ci/run-jewelry-client-smoke.sh"
