#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
PORT="$ROOT/rpg-series-port/shield-api-forge-1.20.1"
UP="$ROOT/.shield-api-upstream"

pick_jar() {
  find "$1" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' ! -name '*dev-shadow*' ! -name '*javadoc*' | sort | head -n1
}

materialize() {
  # shellcheck disable=SC1091
  source "$PORT/UPSTREAM_PINS.env"
  rm -rf "$UP" "$PORT/common/src/generatedUpstream"
  mkdir -p "$UP"
  clone_exact() {
    local sha="$1" dest="$2"
    git init -q "$dest"
    git -C "$dest" remote add origin https://github.com/FabricExtras/ShieldAPI.git
    git -C "$dest" fetch -q --depth=1 origin "$sha"
    git -C "$dest" checkout -q --detach FETCH_HEAD
    test "$(git -C "$dest" rev-parse HEAD)" = "$sha"
  }
  clone_exact "$SHIELD_API_LEGACY_1201_SHA" "$UP/legacy" & p1=$!
  clone_exact "$SHIELD_API_CURRENT_SHA" "$UP/current" & p2=$!
  wait "$p1" "$p2"
  test "$(git -C "$UP/current" rev-parse HEAD^{tree})" = "$SHIELD_API_CURRENT_TREE"
  python3 "$PORT/tools/prepare_upstream_source.py" "$UP/legacy" "$UP/current" "$PORT/common"
}

[[ -f "$PORT/gradle.properties" ]]
grep -Fx 'minecraft_version=1.20.1' "$PORT/gradle.properties" >/dev/null
grep -Fx 'forge_version=1.20.1-47.4.23' "$PORT/gradle.properties" >/dev/null
grep -Fx 'yarn_mappings=1.20.1+build.10' "$PORT/gradle.properties" >/dev/null
grep -Fx 'loom.platform=forge' "$PORT/forge/gradle.properties" >/dev/null

materialize
GEN="$PORT/common/src/generatedUpstream"
MIXINS="$GEN/resources/shield_api.mixins.json"
test -f "$GEN/java/net/fabric_extras/shield_api/item/CustomShieldItem.java"
test -f "$GEN/java/net/fabric_extras/shield_api/mixin/entity/player/PlayerEntityMixin.java"
test ! -e "$GEN/java/net/fabric_extras/shield_api/mixin/item/AxeItemMixin.java"
grep -F 'set(customShieldItem, 100)' "$GEN/java/net/fabric_extras/shield_api/mixin/entity/player/PlayerEntityMixin.java" >/dev/null
if grep -Eq 'EnchantmentHelper|BREAK_SHIELD|nextFloat\(\) <' "$GEN/java/net/fabric_extras/shield_api/mixin/entity/player/PlayerEntityMixin.java"; then
  echo '[Shield API] historical probabilistic disable semantics leaked into current behavior port' >&2
  exit 2
fi
python3 - "$MIXINS" <<'PY'
import json,sys
p=sys.argv[1]
d=json.load(open(p,encoding='utf-8'))
assert d.get('compatibilityLevel') == 'JAVA_17', d
assert 'item.AxeItemMixin' not in d.get('mixins',[]), d
assert 'client.MinecraftClientMixin' not in d.get('mixins',[]), d
assert 'client.MinecraftClientMixin' in d.get('client',[]), d
assert 'client.ModelPredicateProviderRegistryInvoker' in d.get('client',[]), d
print('[Shield API] mixin translation gate passed: JAVA_17 + target-era Axe omission + client dist ownership.')
PY

echo '[Shield API] First exact compile/remapped package'
gradle --no-daemon --stacktrace -p "$PORT" clean :common:compileJava :forge:remapJar
JAR="$(pick_jar "$PORT/forge/build/libs")"
[[ -n "$JAR" && -f "$JAR" ]]
unzip -tq "$JAR"
unzip -p "$JAR" META-INF/mods.toml | grep -F 'modId="shield_api"' >/dev/null
for class in \
  'net/fabric_extras/shield_api/item/CustomShieldItem.class' \
  'net/fabric_extras/shield_api/mixin/entity/player/PlayerEntityMixin.class' \
  'net/fabric_extras/shield_api/forge/ShieldAPIForge.class' \
  'net/fabric_extras/shield_api/forge/ShieldAPISelfTest.class'; do
  unzip -Z1 "$JAR" | grep -Fx "$class" >/dev/null || { echo "[Shield API] required class missing: $class" >&2; exit 1; }
done
if unzip -Z1 "$JAR" | grep -F 'net/fabric_extras/shield_api/mixin/item/AxeItemMixin.class' >/dev/null; then
  echo '[Shield API] post-1.20.1 Axe mixin leaked into target JAR' >&2; exit 1
fi
if unzip -Z1 "$JAR" | grep -E '(^|/)(net/neoforged|net/fabricmc)/' >/dev/null; then
  echo '[Shield API] Fabric/NeoForge implementation classes leaked into native Forge release' >&2; exit 1
fi
unzip -p "$JAR" shield_api.mixins.json > "$PORT/shield-api-packaged-mixins.json"
python3 - "$PORT/shield-api-packaged-mixins.json" <<'PY'
import json,sys
d=json.load(open(sys.argv[1],encoding='utf-8'))
assert d.get('compatibilityLevel') == 'JAVA_17'
assert 'item.AxeItemMixin' not in d.get('mixins',[])
assert 'client.MinecraftClientMixin' not in d.get('mixins',[])
assert 'client.MinecraftClientMixin' in d.get('client',[])
print('[Shield API] packaged mixin config is target-native and dedicated-server safe.')
PY
python3 - "$JAR" <<'PY'
import struct,sys,zipfile
jar=sys.argv[1]; owned=total=0; bad=[]; newer=[]
with zipfile.ZipFile(jar) as zf:
    for name in zf.namelist():
        if not name.endswith('.class'): continue
        data=zf.read(name); total += 1
        if len(data) < 8 or data[:4] != b'\xca\xfe\xba\xbe': bad.append(name); continue
        major=struct.unpack('>H',data[6:8])[0]
        if major > 61: newer.append((name,major))
        if name.startswith('net/fabric_extras/shield_api/'):
            owned += 1
            if major != 61: bad.append(f'{name}=major{major}')
if owned < 7: raise SystemExit(f'[Shield API] incomplete owned class inventory: {owned}')
if bad: raise SystemExit('[Shield API] invalid/non-Java17 owned classes: '+', '.join(bad[:20]))
if newer: raise SystemExit('[Shield API] packaged classes newer than Java17: '+', '.join(f'{n}={m}' for n,m in newer[:20]))
print(f'[Shield API] Java gate passed: {owned} owned classes major 61; {total} packaged classes <=61.')
PY
FIRST_SHA="$(sha256sum "$JAR" | awk '{print $1}')"
printf '%s  %s\n' "$FIRST_SHA" "$JAR" | tee "$PORT/shield-api-first-build.sha256"

# Re-materialize from immutable pins and rebuild from clean state to prove reproducible release bytes.
echo '[Shield API] Reproducibility gate from immutable source pins'
materialize
gradle --no-daemon --stacktrace -p "$PORT" clean :common:compileJava :forge:remapJar
JAR="$(pick_jar "$PORT/forge/build/libs")"
SECOND_SHA="$(sha256sum "$JAR" | awk '{print $1}')"
printf '%s  %s\n' "$SECOND_SHA" "$JAR" | tee "$PORT/shield-api.sha256"
[[ "$FIRST_SHA" = "$SECOND_SHA" ]] || { echo "[Shield API] non-deterministic JAR: $FIRST_SHA != $SECOND_SHA" >&2; exit 1; }

SOURCE_ZIP="$ROOT/shield-api-2.1.0-forge-1.20.1-source-ci.zip"
rm -f "$SOURCE_ZIP"
python3 - "$PORT" "$SOURCE_ZIP" <<'PY'
from pathlib import Path
import stat,sys,zipfile
src=Path(sys.argv[1]).resolve(); out=Path(sys.argv[2]).resolve(); skip={'.gradle','build','run','runs','.git','libs','generatedUpstream'}
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
sha256sum "$SOURCE_ZIP" | tee "$PORT/shield-api-source.sha256"

# Execute current 2.1.0 semantics inside a genuinely initialized Forge server.
echo '[Shield API] Forge dev-server semantic gate'
rm -rf "$PORT/forge/run/logs"; mkdir -p "$PORT/forge/run"; printf 'eula=true\n' > "$PORT/forge/run/eula.txt"
SERVER_SMOKE="$PORT/shield-api-server-smoke.log"; : > "$SERVER_SMOKE"
set +e
timeout --signal=TERM --kill-after=10s 150s env SHIELD_API_SELF_TEST=1 gradle --no-daemon -p "$PORT" :forge:runServer > "$SERVER_SMOKE" 2>&1
STATUS=$?
set -e
SERVER_LOG=$(find "$PORT/forge/run" -type f -path '*/logs/latest.log' | head -n1 || true)
FILES=("$SERVER_SMOKE"); [[ -n "$SERVER_LOG" ]] && FILES+=("$SERVER_LOG")
FATAL_SERVER='SHIELD_API_SELF_TEST_FAILED|MixinApplyError|InvalidMixinException|MixinTransformerError|Failed to create mod instance|NoClassDefFoundError|ClassNotFoundException|Missing or unsupported mandatory dependencies|Exception in server tick loop|The game crashed|Attempted to load class .* for invalid dist DEDICATED_SERVER'
if grep -Eiq "$FATAL_SERVER" "${FILES[@]}"; then cat "${FILES[@]}"; exit 1; fi
if ! grep -Fq 'SHIELD_API_SELF_TEST_PASS' "${FILES[@]}"; then cat "${FILES[@]}"; echo '[Shield API] semantic self-test did not report success' >&2; exit 1; fi
if [[ -n "$SERVER_LOG" ]] && grep -Eq 'Done \([0-9.]+s\)!' "$SERVER_LOG"; then
  echo '[Shield API] Forge dev server reached ready state with exact 2.1.0 semantic self-test green.'
elif [[ "$STATUS" -ne 124 && "$STATUS" -ne 143 ]]; then
  cat "${FILES[@]}"; exit "$STATUS"
else
  cat "${FILES[@]}"; echo '[Shield API] server did not prove ready state' >&2; exit 1
fi

# Client-only mixins/model-predicate path must apply with an actual CustomShieldItem instance present.
echo '[Shield API] Headless Forge client + registered custom-shield model-predicate gate'
rm -rf "$PORT/forge/run/logs"
mkdir -p "$PORT/forge/run/config"
printf 'earlyWindowControl = false\n' > "$PORT/forge/run/config/fml.toml"
CLIENT_SMOKE="$PORT/shield-api-client-smoke.log"; : > "$CLIENT_SMOKE"
set +e
timeout --signal=TERM --kill-after=10s 180s env SHIELD_API_SELF_TEST=1 LIBGL_ALWAYS_SOFTWARE=1 MESA_LOADER_DRIVER_OVERRIDE=llvmpipe \
  xvfb-run -a -s '-screen 0 1280x720x24 +extension GLX +extension RENDER -noreset' \
  gradle --no-daemon -p "$PORT" :forge:runClient </dev/null > "$CLIENT_SMOKE" 2>&1
CLIENT_STATUS=$?
set -e
CLIENT_LOG=$(find "$PORT/forge/run" -type f -path '*/logs/latest.log' | head -n1 || true)
CLIENT_FILES=("$CLIENT_SMOKE"); [[ -n "$CLIENT_LOG" ]] && CLIENT_FILES+=("$CLIENT_LOG")
FATAL_CLIENT='MixinApplyError|InvalidMixinException|MixinTransformerError|Failed to create mod instance|NoClassDefFoundError|ClassNotFoundException|The game crashed whilst initializing game|Exception in thread "Render thread"|Failed to initialize graphics window|Timed out trying to setup the Game Window|Could not initialize GLFW|Missing or unsupported mandatory dependencies|AxeItemMixin'
if grep -Eiq "$FATAL_CLIENT" "${CLIENT_FILES[@]}"; then cat "${CLIENT_FILES[@]}"; exit 1; fi
if [[ -n "$CLIENT_LOG" ]] && grep -Fq 'Reloading ResourceManager' "$CLIENT_LOG" && grep -Fq 'Backend library: LWJGL' "$CLIENT_LOG" && grep -Fq 'Shield API initialized!' "$CLIENT_LOG"; then
  echo '[Shield API] Headless client reached post-bootstrap resource/render runtime with Shield API client mixins loaded.'
elif [[ "$CLIENT_STATUS" -ne 124 && "$CLIENT_STATUS" -ne 143 ]]; then
  cat "${CLIENT_FILES[@]}"; echo "[Shield API] client exited before bootstrap evidence: $CLIENT_STATUS" >&2; exit 1
else
  cat "${CLIENT_FILES[@]}"; echo '[Shield API] client timed out before proven post-bootstrap state' >&2; exit 1
fi

# Fresh official Forge install proves the exact release JAR survives outside Loom and repeats semantics.
echo '[Shield API] Fresh packaged Forge 47.4.23 server + semantic gate'
FRESH="$PORT/.fresh-shield-api-forge-server"
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
sha256sum "$INSTALLED" | tee "$PORT/shield-api-package-installed.sha256"
[[ "$(sha256sum "$INSTALLED" | awk '{print $1}')" = "$SECOND_SHA" ]] || { echo '[Shield API] packaged server JAR bytes differ from release JAR' >&2; exit 1; }

stop_tree() {
  local root="$1" child kids
  kids="$(pgrep -P "$root" 2>/dev/null || true)"
  for child in $kids; do kill -TERM "$child" 2>/dev/null || true; done
  kill -TERM "$root" 2>/dev/null || true; sleep 1
  kids="$(pgrep -P "$root" 2>/dev/null || true)"
  for child in $kids; do kill -KILL "$child" 2>/dev/null || true; done
  kill -KILL "$root" 2>/dev/null || true; wait "$root" 2>/dev/null || true
}
PACKAGE_LOG="$PORT/shield-api-package-server-smoke.log"; : > "$PACKAGE_LOG"
( cd "$FRESH" && exec env SHIELD_API_SELF_TEST=1 ./run.sh nogui ) > "$PACKAGE_LOG" 2>&1 & PID=$!
DEADLINE=$((SECONDS+150)); PASS=0; SEMANTIC=0
FATAL_PACKAGE='SHIELD_API_SELF_TEST_FAILED|ModLoadingException|Failed to create mod instance|NoClassDefFoundError|ClassNotFoundException|MixinApplyError|InvalidMixinException|MixinTransformerError|Exception in server tick loop|The game crashed|Attempted to load class .* for invalid dist DEDICATED_SERVER|AxeItemMixin'
while kill -0 "$PID" 2>/dev/null && (( SECONDS < DEADLINE )); do
  if grep -Eiq "$FATAL_PACKAGE" "$PACKAGE_LOG"; then stop_tree "$PID"; cat "$PACKAGE_LOG"; exit 1; fi
  grep -Fq 'SHIELD_API_SELF_TEST_PASS' "$PACKAGE_LOG" && SEMANTIC=1
  if grep -Eq 'Done \([0-9.]+s\)!' "$PACKAGE_LOG" && [[ "$SEMANTIC" = 1 ]]; then PASS=1; break; fi
  sleep 2
done
stop_tree "$PID"
if [[ "$PASS" != 1 ]]; then cat "$PACKAGE_LOG"; echo '[Shield API] fresh packaged server did not prove ready + semantic state' >&2; exit 1; fi
if grep -Eiq "$FATAL_PACKAGE" "$PACKAGE_LOG"; then cat "$PACKAGE_LOG"; exit 1; fi
echo '[Shield API] Fresh packaged Forge server reached ready state and repeated exact semantic tests using byte-identical release JAR.'

echo '[Shield API] ACCEPTANCE PASS: pins/translation/package/Java17/determinism/current semantics/dev-server/client-bootstrap/fresh-packaged-server.'
