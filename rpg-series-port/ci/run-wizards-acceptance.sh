#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
PORT="$ROOT/rpg-series-port/wizards-forge-1.20.1"
GEN="$PORT/generated"

# Reuse the proven exact-pin reconstruction/compatibility/compile/package bootstrap first.
bash "$ROOT/rpg-series-port/ci/run-wizards.sh"

pick_jar() {
  find "$1" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' ! -name '*dev-shadow*' ! -name '*javadoc*' | sort | head -n1
}
find_module_jar() {
  local group="$1" artifact="$2" version="$3"
  local root="$HOME/.gradle/caches/modules-2/files-2.1/$group/$artifact/$version"
  local jar
  jar="$(find "$root" -type f -name "${artifact}-*.jar" ! -name '*sources*' ! -name '*javadoc*' | sort | head -n1 || true)"
  [[ -n "$jar" && -f "$jar" ]] || { echo "[Wizards] Missing resolved runtime module JAR: $group:$artifact:$version" >&2; exit 1; }
  printf '%s\n' "$jar"
}
prop() { sed -n "s/^${1}=//p" "$GEN/gradle.properties" | tail -n1 | tr -d '\r'; }

JAR="$(pick_jar "$GEN/forge/build/libs")"
[[ -n "$JAR" && -f "$JAR" ]]
FIRST_JAR_SHA="$(sha256sum "$JAR" | awk '{print $1}')"
printf '%s  %s\n' "$FIRST_JAR_SHA" "$JAR" | tee "$PORT/wizards-first-build.sha256"

# Wizards-owned classes must target Java 17 exactly. Architectury may inject helper bytecode
# compiled for an older Java release, which is valid on Java 17; reject anything newer than 17.
python3 - "$JAR" <<'PY'
from pathlib import Path
import struct, sys, zipfile
jar = Path(sys.argv[1])
invalid = []
newer_than_java17 = []
wizards_not_java17 = []
count = owned = 0
with zipfile.ZipFile(jar) as zf:
    for name in zf.namelist():
        if not name.endswith('.class'):
            continue
        data = zf.read(name)
        count += 1
        if len(data) < 8 or data[:4] != b'\xca\xfe\xba\xbe':
            invalid.append(name)
            continue
        major = struct.unpack('>H', data[6:8])[0]
        if major > 61:
            newer_than_java17.append((name, major))
        if name.startswith('net/wizards/'):
            owned += 1
            if major != 61:
                wizards_not_java17.append((name, major))
if count == 0 or owned == 0:
    raise SystemExit(f'[Wizards] invalid release class inventory: total={count}, owned={owned}')
if invalid:
    raise SystemExit('[Wizards] invalid class headers: ' + ', '.join(invalid[:30]))
if newer_than_java17:
    raise SystemExit('[Wizards] classes newer than Java 17: ' + ', '.join(f'{n}={m}' for n,m in newer_than_java17[:30]))
if wizards_not_java17:
    raise SystemExit('[Wizards] Wizards-owned classes not Java 17 major 61: ' + ', '.join(f'{n}={m}' for n,m in wizards_not_java17[:30]))
print(f'[Wizards] Java gate passed: {owned} Wizards-owned classes are major 61; all {count} packaged classes are Java-17-compatible (major <= 61).')
PY

# Current 3.1.1 entity animation/model sources must survive reconstruction; parse every generated JSON resource.
for rel in \
  common/src/main/java/net/wizards/client/entity/ArcaneEmitterAnimations.java \
  common/src/main/java/net/wizards/client/entity/FireHydraAnimations.java \
  common/src/main/java/net/wizards/client/entity/FrostElementalAnimations.java \
  common/src/main/java/net/wizards/client/entity/ArcaneEmitterModel.java \
  common/src/main/java/net/wizards/client/entity/FireHydraModel.java \
  common/src/main/java/net/wizards/client/entity/FrostElementalModel.java; do
  test -f "$GEN/$rel"
done
python3 - "$GEN" <<'PY'
from pathlib import Path
import json, sys
root = Path(sys.argv[1])
files = sorted(set(root.glob('common/src/main/resources/**/*.json')) | set(root.glob('common/src/main/generated/**/*.json')) | set(root.glob('forge/src/main/resources/**/*.json')))
if not files:
    raise SystemExit('[Wizards] no JSON resources found after reconstruction')
for path in files:
    try:
        json.loads(path.read_text())
    except Exception as exc:
        raise SystemExit(f'[Wizards] invalid JSON resource {path.relative_to(root)}: {exc}')
print(f'[Wizards] resource integrity gate parsed {len(files)} JSON resources; current entity animation/model source files are present.')
PY

echo '[Wizards] Reproducibility gate: clean remapped release must be byte-identical'
gradle --no-daemon --stacktrace -p "$GEN" clean :forge:remapJar
JAR="$(pick_jar "$GEN/forge/build/libs")"
[[ -n "$JAR" && -f "$JAR" ]]
SECOND_JAR_SHA="$(sha256sum "$JAR" | awk '{print $1}')"
printf '%s  %s\n' "$SECOND_JAR_SHA" "$JAR" | tee "$PORT/wizards.sha256"
if [[ "$FIRST_JAR_SHA" != "$SECOND_JAR_SHA" ]]; then
  echo "[Wizards] non-reproducible release JAR: first=$FIRST_JAR_SHA second=$SECOND_JAR_SHA" >&2
  exit 1
fi
echo "[Wizards] Reproducible release JAR SHA-256: $SECOND_JAR_SHA"
unzip -tq "$JAR"
unzip -p "$JAR" META-INF/mods.toml | grep -F 'modId="wizards"' >/dev/null
unzip -p "$JAR" META-INF/mods.toml | grep -F 'modId="tiny_config"' >/dev/null
unzip -p "$JAR" META-INF/MANIFEST.MF | tr -d '\r' | grep -F 'MixinConfigs: wizards.mixins.json' >/dev/null

SOURCE_ZIP="$ROOT/wizards-3.1.1-forge-1.20.1-source-ci.zip"
rm -f "$SOURCE_ZIP"
python3 - "$GEN" "$SOURCE_ZIP" <<'PY'
from pathlib import Path
import stat, sys, zipfile
src = Path(sys.argv[1]).resolve(); out = Path(sys.argv[2]).resolve()
skip = {'.gradle', 'build', 'run', 'runs', '.git', 'libs'}
files = []
for path in src.rglob('*'):
    rel = path.relative_to(src)
    if any(part in skip for part in rel.parts):
        continue
    if path.is_file():
        files.append((rel.as_posix(), path))
with zipfile.ZipFile(out, 'w', compression=zipfile.ZIP_DEFLATED, compresslevel=9) as zf:
    for arcname, path in sorted(files):
        info = zipfile.ZipInfo(arcname, date_time=(1980,1,1,0,0,0))
        info.compress_type = zipfile.ZIP_DEFLATED
        info.external_attr = (stat.S_IFREG | 0o644) << 16
        info.create_system = 3
        zf.writestr(info, path.read_bytes(), compress_type=zipfile.ZIP_DEFLATED, compresslevel=9)
PY
unzip -tq "$SOURCE_ZIP"
if unzip -Z1 "$SOURCE_ZIP" | grep -E '(^|/)(\.gradle|build|run|runs|libs)/' >/dev/null; then
  echo '[Wizards] cache/build/runtime/dependency binaries leaked into deterministic source archive' >&2
  exit 1
fi
sha256sum "$SOURCE_ZIP" | tee "$PORT/wizards-source.sha256"

# Forge dev-server gate: require a real ready state and reject loader/mixin/classloading crashes.
echo '[Wizards] Dedicated Forge dev-server gate'
rm -rf "$GEN/forge/run/logs"
mkdir -p "$GEN/forge/run"
printf 'eula=true\n' > "$GEN/forge/run/eula.txt"
SERVER_SMOKE="$PORT/wizards-server-smoke.log"
: > "$SERVER_SMOKE"
set +e
timeout --signal=TERM --kill-after=10s 210s gradle --no-daemon -p "$GEN" :forge:runServer > "$SERVER_SMOKE" 2>&1
SERVER_STATUS=$?
set -e
SERVER_LOG=$(find "$GEN/forge/run" -type f -path '*/logs/latest.log' | head -n1 || true)
SERVER_FILES=("$SERVER_SMOKE"); [[ -n "$SERVER_LOG" ]] && SERVER_FILES+=("$SERVER_LOG")
FATAL_SERVER='MixinApplyError|InvalidMixinException|MixinTransformerError|Failed to create mod instance|NoClassDefFoundError|ClassNotFoundException|Missing or unsupported mandatory dependencies|Exception in server tick loop|The game crashed'
if grep -Eiq "$FATAL_SERVER" "${SERVER_FILES[@]}"; then cat "${SERVER_FILES[@]}"; exit 1; fi
if [[ -n "$SERVER_LOG" ]] && grep -Eq 'Done \([0-9.]+s\)!' "$SERVER_LOG"; then
  echo '[Wizards] Dedicated Forge dev server reached ready state.'
elif [[ "$SERVER_STATUS" -ne 124 && "$SERVER_STATUS" -ne 143 ]]; then
  cat "${SERVER_FILES[@]}"; echo "[Wizards] dev server exited before ready state: $SERVER_STATUS" >&2; exit 1
else
  cat "${SERVER_FILES[@]}"; echo '[Wizards] dev server timed out before proven ready state' >&2; exit 1
fi

# Headless real-client bootstrap. Disable only Forge's early splash controller; Minecraft's real window remains exercised.
echo '[Wizards] Headless Forge client/resource/model bootstrap gate'
rm -rf "$GEN/forge/run/logs"
mkdir -p "$GEN/forge/run/config"
printf 'earlyWindowControl = false\n' > "$GEN/forge/run/config/fml.toml"
CLIENT_SMOKE="$PORT/wizards-client-smoke.log"
: > "$CLIENT_SMOKE"
set +e
timeout --signal=TERM --kill-after=10s 210s env LIBGL_ALWAYS_SOFTWARE=1 MESA_LOADER_DRIVER_OVERRIDE=llvmpipe \
  xvfb-run -a -s '-screen 0 1280x720x24 +extension GLX +extension RENDER -noreset' \
  gradle --no-daemon -p "$GEN" :forge:runClient </dev/null > "$CLIENT_SMOKE" 2>&1
CLIENT_STATUS=$?
set -e
CLIENT_LOG=$(find "$GEN/forge/run" -type f -path '*/logs/latest.log' | head -n1 || true)
CLIENT_FILES=("$CLIENT_SMOKE"); [[ -n "$CLIENT_LOG" ]] && CLIENT_FILES+=("$CLIENT_LOG")
FATAL_CLIENT='MixinApplyError|InvalidMixinException|MixinTransformerError|Failed to create mod instance|NoClassDefFoundError|ClassNotFoundException|The game crashed whilst initializing game|Exception in thread "Render thread"|Failed to initialize graphics window|Timed out trying to setup the Game Window|Could not initialize GLFW|Missing or unsupported mandatory dependencies|Using missing texture.*wizards|Failed to load model.*wizards|Unable to load model.*wizards'
if grep -Eiq "$FATAL_CLIENT" "${CLIENT_FILES[@]}"; then cat "${CLIENT_FILES[@]}"; exit 1; fi
if [[ -n "$CLIENT_LOG" ]] && grep -Fq 'Reloading ResourceManager' "$CLIENT_LOG" && grep -Fq 'Backend library: LWJGL' "$CLIENT_LOG"; then
  echo '[Wizards] Headless client reached post-bootstrap resource/render runtime.'
elif [[ "$CLIENT_STATUS" -ne 124 && "$CLIENT_STATUS" -ne 143 ]]; then
  cat "${CLIENT_FILES[@]}"; echo "[Wizards] client exited before bootstrap evidence: $CLIENT_STATUS" >&2; exit 1
else
  cat "${CLIENT_FILES[@]}"; echo '[Wizards] client timed out before proven post-bootstrap state' >&2; exit 1
fi

# Fresh packaged server with the exact release JAR and every known real runtime dependency as separate mods.
STRUCTURE_FORGE="$(pick_jar "$ROOT/rpg-series-port/structure_pool_api-forge-1.20.1/forge/build/libs")"
RUNES_FORGE="$(pick_jar "$ROOT/rpg-series-port/runes-forge-1.20.1/forge/build/libs")"
SPELL_POWER_FORGE="$(pick_jar "$ROOT/rpg-series-port/spell_power-forge-1.20.1/forge/build/libs")"
RANGED_FORGE="$(pick_jar "$ROOT/rpg-series-port/ranged-weapon-api-forge-1.20.1/forge/build/libs")"
SPELL_ENGINE_FORGE="$ROOT/rpg-series-port/spell-engine-forge-1.20.1/spell_engine-forge-1.10.2+1.20.1.jar"
ARMOR_FORGE="$(pick_jar "$ROOT/rpg-series-port/armor-model-api-forge-1.20.1/generated/forge/build/libs")"
TINY_FORGE="$(pick_jar "$ROOT/rpg-series-port/tiny-config-forge-1.20.1/generated/forge/build/libs")"
for dep in "$STRUCTURE_FORGE" "$RUNES_FORGE" "$SPELL_POWER_FORGE" "$RANGED_FORGE" "$SPELL_ENGINE_FORGE" "$ARMOR_FORGE" "$TINY_FORGE"; do
  [[ -n "$dep" && -f "$dep" ]]; unzip -tq "$dep"
done
CLOTH_VERSION="$(prop cloth_config_version)"; PLAYER_VERSION="$(prop player_anim_version)"; CURIOS_VERSION="$(prop curios_version)"
CLOTH_JAR="$(find_module_jar me.shedaniel.cloth cloth-config-forge "$CLOTH_VERSION")"
PLAYER_JAR="$(find_module_jar dev.kosmx.player-anim player-animation-lib-forge "$PLAYER_VERSION")"
CURIOS_JAR="$(find_module_jar top.theillusivec4.curios curios-forge "$CURIOS_VERSION")"

FRESH="$GEN/.fresh-wizards-forge-server"
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
  cp "$JAR" "$STRUCTURE_FORGE" "$RUNES_FORGE" "$SPELL_POWER_FORGE" "$RANGED_FORGE" "$SPELL_ENGINE_FORGE" "$ARMOR_FORGE" "$TINY_FORGE" "$CLOTH_JAR" "$PLAYER_JAR" "$CURIOS_JAR" mods/
)
[[ "$(find "$FRESH/mods" -maxdepth 1 -type f -name '*.jar' | wc -l | tr -d ' ')" = 11 ]]
INSTALLED="$FRESH/mods/$(basename "$JAR")"
cmp -s "$JAR" "$INSTALLED"
sha256sum "$INSTALLED" | tee "$PORT/wizards-package-installed.sha256"

stop_tree() {
  local root="$1" child kids
  kids="$(pgrep -P "$root" 2>/dev/null || true)"
  for child in $kids; do kill -TERM "$child" 2>/dev/null || true; done
  kill -TERM "$root" 2>/dev/null || true; sleep 1
  kids="$(pgrep -P "$root" 2>/dev/null || true)"
  for child in $kids; do kill -KILL "$child" 2>/dev/null || true; done
  kill -KILL "$root" 2>/dev/null || true; wait "$root" 2>/dev/null || true
}
PACKAGE_LOG="$PORT/wizards-package-server-smoke.log"; : > "$PACKAGE_LOG"
( cd "$FRESH" && exec ./run.sh nogui ) > "$PACKAGE_LOG" 2>&1 & PID=$!
DEADLINE=$((SECONDS+180)); PASS=0
FATAL_PACKAGE='ModLoadingException|Failed to create mod instance|Failed to start the minecraft server|NoClassDefFoundError|ClassNotFoundException|MixinApplyError|InvalidMixinException|MixinTransformerError|Registry is already frozen|Can not register to a locked registry|Missing or unsupported mandatory dependencies|Exception in server tick loop|The game crashed'
while ((SECONDS<DEADLINE)); do
  LATEST="$FRESH/logs/latest.log"; FILES=("$PACKAGE_LOG"); [[ -f "$LATEST" ]] && FILES+=("$LATEST")
  if grep -Eiq "$FATAL_PACKAGE" "${FILES[@]}"; then tail -n 500 "${FILES[@]}" || true; stop_tree "$PID"; exit 1; fi
  if [[ -f "$LATEST" ]] && grep -Eq 'Done \([0-9.]+s\)!' "$LATEST"; then PASS=1; break; fi
  if ! kill -0 "$PID" 2>/dev/null; then wait "$PID" || true; tail -n 500 "${FILES[@]}" || true; exit 1; fi
  sleep 1
done
if [[ "$PASS" -ne 1 ]]; then tail -n 500 "$PACKAGE_LOG" || true; [[ -f "$FRESH/logs/latest.log" ]] && tail -n 500 "$FRESH/logs/latest.log" || true; stop_tree "$PID"; exit 1; fi
stop_tree "$PID"
echo '[Wizards] Fresh packaged Forge 47.4.23 server reached ready state with release Wizards + separate runtime dependencies.'
echo '[Wizards] Full acceptance gates passed: reproducible release, Java 17, resource/model integrity, dev server, headless client, and fresh packaged server.'
