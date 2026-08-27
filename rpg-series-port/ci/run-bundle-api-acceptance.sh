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
if grep -R -E 'net\.neoforged|net\.fabricmc\.api|DataComponentTypes|CUSTOM_BUNDLE_CONTENTS_COMPONENT' \
  "$PORT/common/src/main/java" "$PORT/forge/src/main/java"; then
  echo '[Bundle API] forbidden 1.21/loader-only API survived the native 1.20.1 port' >&2
  exit 1
fi
for required in \
  'com/github/theredbrain/bundleapi/BundleAPI.java' \
  'com/github/theredbrain/bundleapi/component/type/CustomBundleContentsComponent.java' \
  'com/github/theredbrain/bundleapi/item/CustomBundleItem.java' \
  'com/github/theredbrain/bundleapi/item/tooltip/CustomBundleTooltipData.java'; do
  test -f "$PORT/common/src/main/java/$required"
done
for required in \
  'com/github/theredbrain/bundleapi/forge/BundleAPISelfTest.java' \
  'com/github/theredbrain/bundleapi/forge/client/BundleAPIForgeClient.java' \
  'com/github/theredbrain/bundleapi/forge/client/CustomBundleTooltipComponent.java'; do
  test -f "$PORT/forge/src/main/java/$required"
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
for class in \
  'com/github/theredbrain/bundleapi/item/tooltip/CustomBundleTooltipData.class' \
  'com/github/theredbrain/bundleapi/forge/client/BundleAPIForgeClient.class' \
  'com/github/theredbrain/bundleapi/forge/client/CustomBundleTooltipComponent.class'; do
  unzip -Z1 "$JAR" | grep -Fx "$class" >/dev/null || { echo "[Bundle API] tooltip parity class missing from release JAR: $class" >&2; exit 1; }
done
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
if owned < 7: raise SystemExit(f'[Bundle API] incomplete owned class inventory after tooltip parity: {owned}')
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

# Run semantic storage/occupancy tests inside a real initialized Forge runtime.
echo '[Bundle API] Dedicated Forge dev-server + semantic self-test gate'
rm -rf "$PORT/forge/run/logs"; mkdir -p "$PORT/forge/run"; printf 'eula=true\n' > "$PORT/forge/run/eula.txt"
SERVER_SMOKE="$PORT/bundle-api-server-smoke.log"; : > "$SERVER_SMOKE"
set +e
timeout --signal=TERM --kill-after=10s 150s env BUNDLE_API_SELF_TEST=1 gradle --no-daemon -p "$PORT" :forge:runServer > "$SERVER_SMOKE" 2>&1
STATUS=$?
set -e
SERVER_LOG=$(find "$PORT/forge/run" -type f -path '*/logs/latest.log' | head -n1 || true)
FILES=("$SERVER_SMOKE"); [[ -n "$SERVER_LOG" ]] && FILES+=("$SERVER_LOG")
if grep -Eiq 'BUNDLE_API_SELF_TEST_FAILED|MixinApplyError|InvalidMixinException|Failed to create mod instance|NoClassDefFoundError|ClassNotFoundException|Missing or unsupported mandatory dependencies|Exception in server tick loop|The game crashed' "${FILES[@]}"; then cat "${FILES[@]}"; exit 1; fi
if ! grep -Fq 'BUNDLE_API_SELF_TEST_PASS' "${FILES[@]}"; then cat "${FILES[@]}"; echo '[Bundle API] semantic self-test did not report success' >&2; exit 1; fi
if [[ -n "$SERVER_LOG" ]] && grep -Eq 'Done \([0-9.]+s\)!' "$SERVER_LOG"; then
  echo '[Bundle API] Forge dev server reached ready state with semantic self-test green.'
elif [[ "$STATUS" -ne 124 && "$STATUS" -ne 143 ]]; then
  cat "${FILES[@]}"; exit "$STATUS"
else
  cat "${FILES[@]}"; echo '[Bundle API] server did not prove ready state' >&2; exit 1
fi

# Tooltip parity is client-side behavior: exercise Forge's real client bootstrap and resource stack.
echo '[Bundle API] Headless Forge client/resource + tooltip-registration bootstrap gate'
rm -rf "$PORT/forge/run/logs"
mkdir -p "$PORT/forge/run/config"
printf 'earlyWindowControl = false\n' > "$PORT/forge/run/config/fml.toml"
CLIENT_SMOKE="$PORT/bundle-api-client-smoke.log"; : > "$CLIENT_SMOKE"
set +e
timeout --signal=TERM --kill-after=10s 180s env LIBGL_ALWAYS_SOFTWARE=1 MESA_LOADER_DRIVER_OVERRIDE=llvmpipe \
  xvfb-run -a -s '-screen 0 1280x720x24 +extension GLX +extension RENDER -noreset' \
  gradle --no-daemon -p "$PORT" :forge:runClient </dev/null > "$CLIENT_SMOKE" 2>&1
CLIENT_STATUS=$?
set -e
CLIENT_LOG=$(find "$PORT/forge/run" -type f -path '*/logs/latest.log' | head -n1 || true)
CLIENT_FILES=("$CLIENT_SMOKE"); [[ -n "$CLIENT_LOG" ]] && CLIENT_FILES+=("$CLIENT_LOG")
FATAL_CLIENT='MixinApplyError|InvalidMixinException|MixinTransformerError|Failed to create mod instance|NoClassDefFoundError|ClassNotFoundException|The game crashed whilst initializing game|Exception in thread "Render thread"|Failed to initialize graphics window|Timed out trying to setup the Game Window|Could not initialize GLFW|Missing or unsupported mandatory dependencies|RegisterClientTooltipComponentFactoriesEvent.*[Ee]xception|CustomBundleTooltip(Component|Data).*(Exception|Error)'
if grep -Eiq "$FATAL_CLIENT" "${CLIENT_FILES[@]}"; then cat "${CLIENT_FILES[@]}"; exit 1; fi
if [[ -n "$CLIENT_LOG" ]] && grep -Fq 'Reloading ResourceManager' "$CLIENT_LOG" && grep -Fq 'Backend library: LWJGL' "$CLIENT_LOG"; then
  echo '[Bundle API] Headless client reached post-bootstrap resource/render runtime with client tooltip registration loaded.'
elif [[ "$CLIENT_STATUS" -ne 124 && "$CLIENT_STATUS" -ne 143 ]]; then
  cat "${CLIENT_FILES[@]}"; echo "[Bundle API] client exited before bootstrap evidence: $CLIENT_STATUS" >&2; exit 1
else
  cat "${CLIENT_FILES[@]}"; echo '[Bundle API] client timed out before proven post-bootstrap state' >&2; exit 1
fi

# Fresh packaged Forge server: prove the exact release artifact is server-safe outside Loom dev runtime.
echo '[Bundle API] Fresh packaged Forge 47.4.23 server gate'
FRESH="$PORT/.fresh-bundle-api-forge-server"
rm -rf "$FRESH"; mkdir -p "$FRESH"
curl -fsSL 'https://maven.minecraftforge.net/net/minecraftforge/forge/1.20.1-47.4.23/forge-1.20.1-47.4.23-installer.jar' -o "$FRESH/forge-installer.jar"
(
  cd "$FRESH"
  java -jar forge-installer.jar --installServer >/dev/null
  printf 'eula=true\n' > eula.txt
  printf '%s\n' '-Xmx2G' > user_jvm_args.txt
  cat > server.properties <<'PROPS'
level-type=minecraft:flat
generate-structures=false
view-distance=3
simulation-distance=3
spawn-protection=0
online-mode=false
PROPS
  mkdir -p mods
  cp "$JAR" mods/
)
INSTALLED="$FRESH/mods/$(basename "$JAR")"
cmp -s "$JAR" "$INSTALLED"
sha256sum "$INSTALLED" | tee "$PORT/bundle-api-package-installed.sha256"
[[ "$(sha256sum "$INSTALLED" | awk '{print $1}')" = "$SECOND_SHA" ]] || { echo '[Bundle API] packaged server JAR bytes differ from release JAR' >&2; exit 1; }

stop_tree() {
  local root="$1" child kids
  kids="$(pgrep -P "$root" 2>/dev/null || true)"
  for child in $kids; do kill -TERM "$child" 2>/dev/null || true; done
  kill -TERM "$root" 2>/dev/null || true; sleep 1
  kids="$(pgrep -P "$root" 2>/dev/null || true)"
  for child in $kids; do kill -KILL "$child" 2>/dev/null || true; done
  kill -KILL "$root" 2>/dev/null || true; wait "$root" 2>/dev/null || true
}
PACKAGE_LOG="$PORT/bundle-api-package-server-smoke.log"; : > "$PACKAGE_LOG"
( cd "$FRESH" && exec ./run.sh nogui ) > "$PACKAGE_LOG" 2>&1 & PID=$!
DEADLINE=$((SECONDS+150)); PASS=0
FATAL_PACKAGE='ModLoadingException|Failed to create mod instance|NoClassDefFoundError|ClassNotFoundException|MixinApplyError|InvalidMixinException|Exception in server tick loop|The game crashed|Attempted to load class .* for invalid dist DEDICATED_SERVER'
while kill -0 "$PID" 2>/dev/null && (( SECONDS < DEADLINE )); do
  if grep -Eiq "$FATAL_PACKAGE" "$PACKAGE_LOG"; then stop_tree "$PID"; cat "$PACKAGE_LOG"; exit 1; fi
  if grep -Eq 'Done \([0-9.]+s\)!' "$PACKAGE_LOG"; then PASS=1; break; fi
  sleep 2
done
stop_tree "$PID"
if [[ "$PASS" != 1 ]]; then cat "$PACKAGE_LOG"; echo '[Bundle API] fresh packaged server did not reach ready state' >&2; exit 1; fi
if grep -Eiq "$FATAL_PACKAGE" "$PACKAGE_LOG"; then cat "$PACKAGE_LOG"; exit 1; fi
echo '[Bundle API] Fresh packaged Forge server reached ready state using byte-identical release JAR.'

echo '[Bundle API] Acceptance boundary passed: compile/package/Java17/determinism/storage semantics/dev-server/client-bootstrap/fresh-packaged-server.'
