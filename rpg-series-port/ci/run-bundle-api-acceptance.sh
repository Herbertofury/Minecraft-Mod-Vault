#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
PORT="$ROOT/rpg-series-port/bundle-api-forge-1.20.1"

pick_jar() {
  find "$1" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' ! -name '*dev-shadow*' ! -name '*javadoc*' | sort | head -n1
}

[[ -f "$PORT/gradle.properties" ]]
grep -Fx 'minecraft_version=1.20.1' "$PORT/gradle.properties" >/dev/null
grep -Fx 'forge_version=1.20.1-47.4.23' "$PORT/gradle.properties" >/dev/null
grep -Fx 'yarn_mappings=1.20.1+build.10' "$PORT/gradle.properties" >/dev/null

# Source/API hygiene before paying for a build.
grep -R -E 'net\.neoforged|net\.fabricmc\.api|DataComponentTypes|CUSTOM_BUNDLE_CONTENTS_COMPONENT' \
  "$PORT/common/src/main/java" "$PORT/forge/src/main/java" && {
  echo '[Bundle API] forbidden 1.21/loader-only API survived the native 1.20.1 port' >&2
  exit 1
} || true
for required in \
  'com/github/theredbrain/bundleapi/BundleAPI.java' \
  'com/github/theredbrain/bundleapi/component/type/CustomBundleContentsComponent.java' \
  'com/github/theredbrain/bundleapi/item/CustomBundleItem.java'; do
  test -f "$PORT/common/src/main/java/$required"
done

echo '[Bundle API] Compile + remapped package'
gradle --no-daemon --stacktrace -p "$PORT" clean :forge:remapJar
JAR="$(pick_jar "$PORT/forge/build/libs")"
[[ -n "$JAR" && -f "$JAR" ]]
unzip -tq "$JAR"
unzip -p "$JAR" META-INF/mods.toml | grep -F 'modId="bundleapi"' >/dev/null
if unzip -Z1 "$JAR" | grep -E '(^|/)(net/neoforged|net/fabricmc)/' >/dev/null; then
  echo '[Bundle API] loader implementation classes leaked into release JAR' >&2; exit 1
fi
FIRST_SHA="$(sha256sum "$JAR" | awk '{print $1}')"
printf '%s  %s\n' "$FIRST_SHA" "$JAR" | tee "$PORT/bundle-api-first-build.sha256"

python3 - "$JAR" <<'PY'
import struct, sys, zipfile
jar=sys.argv[1]; owned=total=0; bad=[]; newer=[]
with zipfile.ZipFile(jar) as zf:
    for name in zf.namelist():
        if not name.endswith('.class'): continue
        data=zf.read(name); total += 1
        if len(data) < 8 or data[:4] != b'\xca\xfe\xba\xbe': bad.append(name); continue
        major=struct.unpack('>H', data[6:8])[0]
        if major > 61: newer.append((name,major))
        if name.startswith('com/github/theredbrain/bundleapi/'):
            owned += 1
            if major != 61: bad.append(f'{name}=major{major}')
if owned < 4: raise SystemExit(f'[Bundle API] incomplete owned class inventory: {owned}')
if bad: raise SystemExit('[Bundle API] invalid/non-Java17 owned classes: '+', '.join(bad[:20]))
if newer: raise SystemExit('[Bundle API] packaged classes newer than Java17: '+', '.join(f'{n}={m}' for n,m in newer[:20]))
print(f'[Bundle API] Java gate passed: {owned} owned classes major 61; {total} packaged classes <=61.')
PY

echo '[Bundle API] Reproducibility gate'
gradle --no-daemon --stacktrace -p "$PORT" clean :forge:remapJar
JAR="$(pick_jar "$PORT/forge/build/libs")"
SECOND_SHA="$(sha256sum "$JAR" | awk '{print $1}')"
printf '%s  %s\n' "$SECOND_SHA" "$JAR" | tee "$PORT/bundle-api.sha256"
[[ "$FIRST_SHA" = "$SECOND_SHA" ]] || { echo "[Bundle API] non-deterministic JAR: $FIRST_SHA != $SECOND_SHA" >&2; exit 1; }

SOURCE_ZIP="$ROOT/bundle-api-1.1.0-forge-1.20.1-source-ci.zip"
rm -f "$SOURCE_ZIP"
python3 - "$PORT" "$SOURCE_ZIP" <<'PY'
from pathlib import Path
import stat,sys,zipfile
src=Path(sys.argv[1]).resolve(); out=Path(sys.argv[2]).resolve(); skip={'.gradle','build','run','runs','.git','libs'}
files=[]
for p in src.rglob('*'):
    rel=p.relative_to(src)
    if p.is_file() and not any(x in skip for x in rel.parts): files.append((rel.as_posix(),p))
with zipfile.ZipFile(out,'w',zipfile.ZIP_DEFLATED,compresslevel=9) as z:
    for name,p in sorted(files):
        info=zipfile.ZipInfo(name,(1980,1,1,0,0,0)); info.compress_type=zipfile.ZIP_DEFLATED; info.external_attr=(stat.S_IFREG|0o644)<<16; info.create_system=3
        z.writestr(info,p.read_bytes(),compress_type=zipfile.ZIP_DEFLATED,compresslevel=9)
PY
unzip -tq "$SOURCE_ZIP"
sha256sum "$SOURCE_ZIP" | tee "$PORT/bundle-api-source.sha256"

# Real Forge development runtime smoke. Bundle API is a foundation library, so a clean
# mod bootstrap is required here; behavior/storage tests are promoted immediately after
# this first compile/runtime boundary is known green.
echo '[Bundle API] Dedicated Forge dev-server gate'
rm -rf "$PORT/forge/run/logs"; mkdir -p "$PORT/forge/run"; printf 'eula=true\n' > "$PORT/forge/run/eula.txt"
SERVER_SMOKE="$PORT/bundle-api-server-smoke.log"; : > "$SERVER_SMOKE"
set +e
timeout --signal=TERM --kill-after=10s 150s gradle --no-daemon -p "$PORT" :forge:runServer > "$SERVER_SMOKE" 2>&1
STATUS=$?
set -e
SERVER_LOG=$(find "$PORT/forge/run" -type f -path '*/logs/latest.log' | head -n1 || true)
FILES=("$SERVER_SMOKE"); [[ -n "$SERVER_LOG" ]] && FILES+=("$SERVER_LOG")
if grep -Eiq 'MixinApplyError|InvalidMixinException|Failed to create mod instance|NoClassDefFoundError|ClassNotFoundException|Missing or unsupported mandatory dependencies|Exception in server tick loop|The game crashed' "${FILES[@]}"; then cat "${FILES[@]}"; exit 1; fi
if [[ -n "$SERVER_LOG" ]] && grep -Eq 'Done \([0-9.]+s\)!' "$SERVER_LOG"; then
  echo '[Bundle API] Forge dev server reached ready state.'
elif [[ "$STATUS" -ne 124 && "$STATUS" -ne 143 ]]; then
  cat "${FILES[@]}"; exit "$STATUS"
else
  cat "${FILES[@]}"; echo '[Bundle API] server did not prove ready state' >&2; exit 1
fi

echo '[Bundle API] First acceptance boundary passed: compile/package/Java17/determinism/dev-server.'
