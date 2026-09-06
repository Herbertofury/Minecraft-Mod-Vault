#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
PORT="$ROOT/rpg-series-port/paladins-forge-1.20.1"
TMP="${RUNNER_TEMP:-/tmp}/paladins-first-compile"

# Reuse the proven exact-pin reconstruction, compatibility transforms, dependency builds,
# Java-17 gate, and first remapped Paladins release boundary.
bash "$ROOT/rpg-series-port/ci/run-paladins-first-compile.sh"

pick_jar() {
  local dir="$1" glob="$2"
  local -a jars=()
  mapfile -t jars < <(find "$dir" -maxdepth 1 -type f -name "$glob" ! -name '*sources*' ! -name '*dev-shadow*' ! -name '*javadoc*' -print | sort)
  (( ${#jars[@]} == 1 )) || { echo "[Paladins acceptance] expected exactly one $glob in $dir, found ${#jars[@]}" >&2; return 1; }
  printf '%s\n' "${jars[0]}"
}

stop_tree() {
  local root="$1" child kids
  kids="$(pgrep -P "$root" 2>/dev/null || true)"
  for child in $kids; do stop_tree "$child"; done
  kill -TERM "$root" 2>/dev/null || true
  sleep 1
  kill -KILL "$root" 2>/dev/null || true
  wait "$root" 2>/dev/null || true
}

wait_for_log_marker() {
  local pid="$1" deadline="$2" marker="$3" fatal="$4" primary="$5" latest="$6" label="$7"
  local pass=0
  while (( SECONDS < deadline )); do
    local -a files=("$primary")
    [[ -f "$latest" ]] && files+=("$latest")
    if grep -Eiq "$fatal" "${files[@]}" 2>/dev/null; then
      echo "[Paladins acceptance] $label fatal signature detected" >&2
      tail -n 500 "${files[@]}" || true
      stop_tree "$pid"
      return 1
    fi
    if [[ -f "$latest" ]] && grep -Eq "$marker" "$latest"; then
      pass=1
      break
    fi
    if ! kill -0 "$pid" 2>/dev/null; then
      wait "$pid" || true
      tail -n 500 "${files[@]}" || true
      echo "[Paladins acceptance] $label exited before marker: $marker" >&2
      return 1
    fi
    sleep 1
  done
  if [[ "$pass" -ne 1 ]]; then
    local -a files=("$primary")
    [[ -f "$latest" ]] && files+=("$latest")
    tail -n 500 "${files[@]}" || true
    stop_tree "$pid"
    echo "[Paladins acceptance] $label timed out before marker: $marker" >&2
    return 1
  fi
  stop_tree "$pid"
}

JAR="$(pick_jar "$PORT/forge/build/libs" '*-forge-*.jar')"
FIRST_SHA="$(sha256sum "$JAR" | awk '{print $1}')"
printf '%s  %s\n' "$FIRST_SHA" "$JAR" | tee "$PORT/paladins-first-build.sha256"

# Resolve the exact dependency graph produced by the first-compile runner. Every foundation
# remains a separate mod at runtime; no dependency classes are shaded into Paladins.
SHIELD_COMMON="$(pick_jar "$ROOT/rpg-series-port/shield-api-forge-1.20.1/common/build/libs" '*-common-*.jar')"
SHIELD_FORGE="$(pick_jar "$ROOT/rpg-series-port/shield-api-forge-1.20.1/forge/build/libs" '*.jar')"
SPELL_POWER_COMMON="$(pick_jar "$ROOT/rpg-series-port/spell_power-forge-1.20.1/common/build/libs" '*-common-*.jar')"
SPELL_POWER_FORGE="$(pick_jar "$ROOT/rpg-series-port/spell_power-forge-1.20.1/forge/build/libs" '*-forge-*.jar')"
TINY_COMMON="$(pick_jar "$ROOT/rpg-series-port/tiny-config-forge-1.20.1/generated/common/build/libs" '*-common-*.jar')"
TINY_FORGE="$(pick_jar "$ROOT/rpg-series-port/tiny-config-forge-1.20.1/generated/forge/build/libs" '*-forge-*.jar')"
SPELL_ENGINE_COMMON="$(pick_jar "$ROOT/.spell-engine-build/common/build/libs" '*-common-*.jar')"
SPELL_ENGINE_FORGE="$ROOT/rpg-series-port/spell-engine-forge-1.20.1/spell_engine-forge-1.10.2+1.20.1.jar"
RUNES_COMMON="$(pick_jar "$ROOT/rpg-series-port/runes-forge-1.20.1/common/build/libs" '*-common-*.jar')"
RUNES_FORGE="$(pick_jar "$ROOT/rpg-series-port/runes-forge-1.20.1/forge/build/libs" '*-forge-*.jar')"
STRUCTURE_COMMON="$(pick_jar "$ROOT/rpg-series-port/structure_pool_api-forge-1.20.1/common/build/libs" '*-common-*.jar')"
STRUCTURE_FORGE="$(pick_jar "$ROOT/rpg-series-port/structure_pool_api-forge-1.20.1/forge/build/libs" '*-forge-*.jar')"
ARMOR_COMMON="$(pick_jar "$ROOT/rpg-series-port/armor-model-api-forge-1.20.1/generated/common/build/libs" '*-common-*.jar')"
ARMOR_FORGE="$(pick_jar "$ROOT/rpg-series-port/armor-model-api-forge-1.20.1/generated/forge/build/libs" '*-forge-*.jar')"
CLOTH_FORGE="$TMP/ext/cloth-config-forge-11.1.136.jar"
PLAYER_FORGE="$TMP/ext/player-animation-lib-forge-1.0.2+1.19.4.jar"
CURIOS_FORGE="$TMP/ext/curios-forge-5.14.1+1.20.1.jar"

for dep in "$SHIELD_COMMON" "$SHIELD_FORGE" "$SPELL_POWER_COMMON" "$SPELL_POWER_FORGE" "$TINY_COMMON" "$TINY_FORGE" \
  "$SPELL_ENGINE_COMMON" "$SPELL_ENGINE_FORGE" "$RUNES_COMMON" "$RUNES_FORGE" "$STRUCTURE_COMMON" "$STRUCTURE_FORGE" \
  "$ARMOR_COMMON" "$ARMOR_FORGE" "$CLOTH_FORGE" "$PLAYER_FORGE" "$CURIOS_FORGE"; do
  [[ -f "$dep" ]] || { echo "[Paladins acceptance] missing dependency: $dep" >&2; exit 1; }
  unzip -tq "$dep" >/dev/null
 done

# The graduated Shield API identity is immutable and must survive every consumer acceptance run.
SHIELD_SHA="$(sha256sum "$SHIELD_FORGE" | awk '{print $1}')"
[[ "$SHIELD_SHA" = 'bd6a2fbeb357c25953abfb14ba18d2c5344e5351c29d2cb082244bc48e8da48a' ]] || {
  echo "[Paladins acceptance] graduated Shield API identity drifted: $SHIELD_SHA" >&2; exit 1; }

ARGS=(
  "-Pshield_api_common_jar=$SHIELD_COMMON"
  "-Parmor_model_api_common_jar=$ARMOR_COMMON"
  "-Prunes_common_jar=$RUNES_COMMON"
  "-Pstructure_pool_api_common_jar=$STRUCTURE_COMMON"
  "-Pspell_power_common_jar=$SPELL_POWER_COMMON"
  "-Pspell_engine_common_jar=$SPELL_ENGINE_COMMON"
  "-Ptiny_config_common_jar=$TINY_COMMON"
  "-Pshield_api_forge_jar=$SHIELD_FORGE"
  "-Parmor_model_api_forge_jar=$ARMOR_FORGE"
  "-Prunes_forge_jar=$RUNES_FORGE"
  "-Pstructure_pool_api_forge_jar=$STRUCTURE_FORGE"
  "-Pspell_power_forge_jar=$SPELL_POWER_FORGE"
  "-Pspell_engine_forge_jar=$SPELL_ENGINE_FORGE"
  "-Ptiny_config_forge_jar=$TINY_FORGE"
  "-Pcloth_config_forge_jar=$CLOTH_FORGE"
  "-Pplayer_animator_forge_jar=$PLAYER_FORGE"
  "-Pcurios_jar=$CURIOS_FORGE"
)

# Static content/resource gate before expensive runtime. These are representative current-source
# systems touched by this backport; all generated JSON must remain syntactically valid.
for rel in \
  common/java/net/paladins/entity/BarrierEntity.java \
  common/java/net/paladins/client/entity/BarrierEntityRenderer.java \
  common/java/net/paladins/entity/LightwellEntity.java \
  common/java/net/paladins/client/entity/LightwellEntityRenderer.java \
  common/java/net/paladins/item/PaladinShields.java \
  common/java/net/paladins/item/armor/Armors.java \
  common/java/net/paladins/village/PaladinVillagers.java; do
  test -f "$PORT/generated/$rel"
done
python3 - "$PORT/generated" <<'PY'
from pathlib import Path
import json, sys
root = Path(sys.argv[1])
files = sorted(set(root.glob('common/resources/**/*.json')) | set(root.glob('forge/resources/**/*.json')) | set(root.glob('common/generated/**/*.json')))
if not files:
    raise SystemExit('[Paladins acceptance] no generated JSON resources found')
for path in files:
    try:
        json.loads(path.read_text())
    except Exception as exc:
        raise SystemExit(f'[Paladins acceptance] invalid JSON {path.relative_to(root)}: {exc}')
print(f'[Paladins acceptance] parsed {len(files)} generated JSON resources; representative current-source systems are present.')
PY

# Reproducibility gate: clean rebuild with the exact same ABI inputs must reproduce the release bit-for-bit.
echo '[Paladins acceptance] Clean reproducibility gate'
gradle --no-daemon --stacktrace -p "$PORT" clean :forge:remapJar "${ARGS[@]}"
JAR="$(pick_jar "$PORT/forge/build/libs" '*-forge-*.jar')"
SECOND_SHA="$(sha256sum "$JAR" | awk '{print $1}')"
printf '%s  %s\n' "$SECOND_SHA" "$JAR" | tee "$PORT/paladins.sha256"
[[ "$FIRST_SHA" = "$SECOND_SHA" ]] || {
  echo "[Paladins acceptance] non-reproducible release: first=$FIRST_SHA second=$SECOND_SHA" >&2; exit 1; }
unzip -tq "$JAR" >/dev/null
unzip -p "$JAR" META-INF/mods.toml | grep -F 'modId="paladins"' >/dev/null
if unzip -Z1 "$JAR" | grep -E '^(net/fabric_extras/shield_api/|net/rpg_foundation/armor_api/|net/runes/|net/fabric_extras/structure_pool/|net/spell_power/|net/spell_engine/|net/tiny_config/)'; then
  echo '[Paladins acceptance] external foundation classes leaked into release during clean rebuild' >&2; exit 1
fi
echo "[Paladins acceptance] reproducible release SHA-256: $SECOND_SHA"

# Deterministic source archive for this exact generated/current-authority candidate.
SOURCE_ZIP="$ROOT/paladins-3.1.1-forge-1.20.1-source-ci.zip"
rm -f "$SOURCE_ZIP"
python3 - "$PORT" "$SOURCE_ZIP" <<'PY'
from pathlib import Path
import stat, sys, zipfile
src = Path(sys.argv[1]).resolve(); out = Path(sys.argv[2]).resolve()
skip = {'.gradle', 'build', 'run', 'runs', '.git', '.upstream'}
files=[]
for path in src.rglob('*'):
    rel=path.relative_to(src)
    if any(part in skip for part in rel.parts):
        continue
    if path.is_file(): files.append((rel.as_posix(), path))
with zipfile.ZipFile(out,'w',compression=zipfile.ZIP_DEFLATED,compresslevel=9) as zf:
    for arcname,path in sorted(files):
        info=zipfile.ZipInfo(arcname,date_time=(1980,1,1,0,0,0)); info.compress_type=zipfile.ZIP_DEFLATED
        info.external_attr=(stat.S_IFREG|0o644)<<16; info.create_system=3
        zf.writestr(info,path.read_bytes(),compress_type=zipfile.ZIP_DEFLATED,compresslevel=9)
PY
unzip -tq "$SOURCE_ZIP" >/dev/null
if unzip -Z1 "$SOURCE_ZIP" | grep -E '(^|/)(\.gradle|build|run|runs|\.upstream)/' >/dev/null; then
  echo '[Paladins acceptance] cache/build/runtime/upstream checkout leaked into source archive' >&2; exit 1
fi
sha256sum "$SOURCE_ZIP" | tee "$PORT/paladins-source.sha256"

# Real Forge dev-server gate. Launch in the background and stop immediately after the real readiness marker
# instead of burning the full timeout after success.
echo '[Paladins acceptance] Dedicated Forge dev-server gate'
rm -rf "$PORT/forge/run/logs"
mkdir -p "$PORT/forge/run"
printf 'eula=true\n' > "$PORT/forge/run/eula.txt"
SERVER_SMOKE="$PORT/paladins-server-smoke.log"; : > "$SERVER_SMOKE"
( gradle --no-daemon -p "$PORT" :forge:runServer "${ARGS[@]}" ) > "$SERVER_SMOKE" 2>&1 & SERVER_PID=$!
SERVER_LATEST="$PORT/forge/run/logs/latest.log"
FATAL_SERVER='MixinApplyError|InvalidMixinException|MixinTransformerError|Failed to create mod instance|NoClassDefFoundError|ClassNotFoundException|Missing or unsupported mandatory dependencies|Registry is already frozen|Can not register to a locked registry|Exception in server tick loop|The game crashed'
wait_for_log_marker "$SERVER_PID" "$((SECONDS+210))" 'Done \([0-9.]+s\)!' "$FATAL_SERVER" "$SERVER_SMOKE" "$SERVER_LATEST" 'dev server'
echo '[Paladins acceptance] Dedicated Forge dev server reached ready state.'

# Real Minecraft client/resource/render bootstrap under Xvfb. This exercises the renderer/model/resource path
# touched by Barrier and Lightwell API translation; missing Paladins resources/models are release blockers.
echo '[Paladins acceptance] Headless Forge client/resource/render bootstrap gate'
rm -rf "$PORT/forge/run/logs"
mkdir -p "$PORT/forge/run/config"
printf 'earlyWindowControl = false\n' > "$PORT/forge/run/config/fml.toml"
CLIENT_SMOKE="$PORT/paladins-client-smoke.log"; : > "$CLIENT_SMOKE"
( env LIBGL_ALWAYS_SOFTWARE=1 MESA_LOADER_DRIVER_OVERRIDE=llvmpipe \
  xvfb-run -a -s '-screen 0 1280x720x24 +extension GLX +extension RENDER -noreset' \
  gradle --no-daemon -p "$PORT" :forge:runClient "${ARGS[@]}" </dev/null ) > "$CLIENT_SMOKE" 2>&1 & CLIENT_PID=$!
CLIENT_LATEST="$PORT/forge/run/logs/latest.log"
FATAL_CLIENT='MixinApplyError|InvalidMixinException|MixinTransformerError|Failed to create mod instance|NoClassDefFoundError|ClassNotFoundException|The game crashed whilst initializing game|Exception in thread "Render thread"|Failed to initialize graphics window|Timed out trying to setup the Game Window|Could not initialize GLFW|Missing or unsupported mandatory dependencies|Using missing texture.*paladins|Failed to load model.*paladins|Unable to load model.*paladins'
wait_for_log_marker "$CLIENT_PID" "$((SECONDS+210))" 'Reloading ResourceManager' "$FATAL_CLIENT" "$CLIENT_SMOKE" "$CLIENT_LATEST" 'headless client'
grep -Fq 'Backend library: LWJGL' "$CLIENT_LATEST" || { tail -n 500 "$CLIENT_LATEST"; echo '[Paladins acceptance] client resource marker occurred without LWJGL backend evidence' >&2; exit 1; }
echo '[Paladins acceptance] Headless client reached post-bootstrap resource/render runtime with LWJGL.'

# Fresh official packaged Forge 47.4.23 server. Install the exact release candidate plus every real runtime
# dependency as separate mods and prove the installed Paladins bytes are unchanged.
FRESH="$PORT/.fresh-paladins-forge-server"
rm -rf "$FRESH"; mkdir -p "$FRESH"
curl --retry 2 --retry-delay 1 -fsSL 'https://maven.minecraftforge.net/net/minecraftforge/forge/1.20.1-47.4.23/forge-1.20.1-47.4.23-installer.jar' -o "$FRESH/forge-installer.jar"
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
  cp "$JAR" "$SHIELD_FORGE" "$ARMOR_FORGE" "$RUNES_FORGE" "$STRUCTURE_FORGE" "$SPELL_POWER_FORGE" \
    "$SPELL_ENGINE_FORGE" "$TINY_FORGE" "$CLOTH_FORGE" "$PLAYER_FORGE" "$CURIOS_FORGE" mods/
)
[[ "$(find "$FRESH/mods" -maxdepth 1 -type f -name '*.jar' | wc -l | tr -d ' ')" = 11 ]] || {
  find "$FRESH/mods" -maxdepth 1 -type f -name '*.jar' -printf '%f\n' | sort >&2
  echo '[Paladins acceptance] packaged server runtime mod count mismatch' >&2; exit 1; }
INSTALLED="$FRESH/mods/$(basename "$JAR")"
cmp -s "$JAR" "$INSTALLED"
sha256sum "$INSTALLED" | tee "$PORT/paladins-package-installed.sha256"
PACKAGE_LOG="$PORT/paladins-package-server-smoke.log"; : > "$PACKAGE_LOG"
( cd "$FRESH" && exec ./run.sh nogui ) > "$PACKAGE_LOG" 2>&1 & PACKAGE_PID=$!
PACKAGE_LATEST="$FRESH/logs/latest.log"
FATAL_PACKAGE='ModLoadingException|Failed to create mod instance|Failed to start the minecraft server|NoClassDefFoundError|ClassNotFoundException|MixinApplyError|InvalidMixinException|MixinTransformerError|Registry is already frozen|Can not register to a locked registry|Missing or unsupported mandatory dependencies|Exception in server tick loop|The game crashed'
wait_for_log_marker "$PACKAGE_PID" "$((SECONDS+180))" 'Done \([0-9.]+s\)!' "$FATAL_PACKAGE" "$PACKAGE_LOG" "$PACKAGE_LATEST" 'fresh packaged server'
echo '[Paladins acceptance] Fresh packaged Forge 47.4.23 server reached ready state with exact Paladins release + separate runtime dependencies.'

echo '[Paladins acceptance] FULL_ACCEPTANCE_PASS: reproducible release, deterministic source, resource integrity, dev server, headless client, and fresh packaged server.'
