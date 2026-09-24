#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
PORT="$ROOT/rpg-series-port/archers-forge-1.20.1"
TMP="${RUNNER_TEMP:-/tmp}/archers-acceptance"
rm -rf "$TMP"; mkdir -p "$TMP/ext"

pick_one() {
  local dir="$1" glob="$2" label="$3"
  local -a found=()
  mapfile -t found < <(find "$dir" -maxdepth 1 -type f -name "$glob" ! -name '*sources*' ! -name '*dev-shadow*' ! -name '*javadoc*' -print | sort)
  if (( ${#found[@]} != 1 )); then
    echo "[Archers acceptance] $label expected exactly one $glob, found ${#found[@]}" >&2
    printf '%s\n' "${found[@]:-}" >&2
    exit 2
  fi
  printf '%s\n' "${found[0]}"
}

validate_jar() {
  local label="$1" jar="$2"
  test -f "$jar"
  unzip -tq "$jar" >/dev/null
  echo "[Archers acceptance] $label sha256=$(sha256sum "$jar" | awk '{print $1}')"
}

download_jar() {
  local label="$1" url="$2" out="$3"
  curl --retry 2 --retry-delay 1 --retry-connrefused -fsSL "$url" -o "$out"
  validate_jar "$label" "$out"
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

collect_run_files() {
  local smoke="$1" run_dir="$2"
  printf '%s\n' "$smoke"
  find "$run_dir" -type f -path '*/logs/latest.log' -print 2>/dev/null | head -n1 || true
}

echo '[Archers acceptance] Rebuild exact package boundary first'
bash "$ROOT/rpg-series-port/ci/run-archers-bootstrap.sh"

RANGED="$ROOT/rpg-series-port/ranged-weapon-api-forge-1.20.1"
SPELL_POWER="$ROOT/rpg-series-port/spell_power-forge-1.20.1"
TINY="$ROOT/rpg-series-port/tiny-config-forge-1.20.1/generated"
SPELL_ENGINE_BUILD="$ROOT/.spell-engine-build"
STRUCTURE="$ROOT/rpg-series-port/structure_pool_api-forge-1.20.1"
BUNDLE="$ROOT/rpg-series-port/bundle-api-forge-1.20.1"
ARMOR="$ROOT/rpg-series-port/armor-model-api-forge-1.20.1/generated"

RANGED_COMMON="$(pick_one "$RANGED/common/build/libs" '*-common-*.jar' 'Ranged Weapon API common')"
RANGED_FORGE="$(pick_one "$RANGED/forge/build/libs" '*-forge-*.jar' 'Ranged Weapon API Forge')"
SPELL_POWER_COMMON="$(pick_one "$SPELL_POWER/common/build/libs" '*-common-*.jar' 'Spell Power common')"
SPELL_POWER_FORGE="$(pick_one "$SPELL_POWER/forge/build/libs" '*-forge-*.jar' 'Spell Power Forge')"
TINY_COMMON="$(pick_one "$TINY/common/build/libs" '*-common-*.jar' 'TinyConfig common')"
TINY_FORGE="$(pick_one "$TINY/forge/build/libs" '*-forge-*.jar' 'TinyConfig Forge')"
SPELL_ENGINE_COMMON="$(pick_one "$SPELL_ENGINE_BUILD/common/build/libs" '*-common-*.jar' 'Spell Engine common')"
SPELL_ENGINE_FORGE="$ROOT/rpg-series-port/spell-engine-forge-1.20.1/spell_engine-forge-1.10.2+1.20.1.jar"
STRUCTURE_COMMON="$(pick_one "$STRUCTURE/common/build/libs" '*-common-*.jar' 'Structure Pool API common')"
STRUCTURE_FORGE="$(pick_one "$STRUCTURE/forge/build/libs" '*-forge-*.jar' 'Structure Pool API Forge')"
BUNDLE_COMMON="$(pick_one "$BUNDLE/common/build/libs" '*-common-*.jar' 'Bundle API common')"
BUNDLE_FORGE="$(pick_one "$BUNDLE/forge/build/libs" '*-forge-*.jar' 'Bundle API Forge')"
ARMOR_COMMON="$(pick_one "$ARMOR/common/build/libs" '*-common-*.jar' 'Armor Model API common')"
ARMOR_FORGE="$(pick_one "$ARMOR/forge/build/libs" '*-forge-*.jar' 'Armor Model API Forge')"
ARCHERS_JAR="$(pick_one "$PORT/forge/build/libs" '*-forge-*.jar' 'Archers Forge release')"

CLOTH_FORGE="$TMP/ext/cloth-config-forge-11.1.136.jar"
PLAYER_FORGE="$TMP/ext/player-animation-lib-forge-1.0.2+1.19.4.jar"
CURIOS_FORGE="$TMP/ext/curios-forge-5.14.1+1.20.1.jar"
download_jar 'Cloth Config Forge 11.1.136' 'https://maven.shedaniel.me/me/shedaniel/cloth/cloth-config-forge/11.1.136/cloth-config-forge-11.1.136.jar' "$CLOTH_FORGE"
download_jar 'Player Animator Forge 1.0.2+1.19.4' 'https://maven.kosmx.dev/dev/kosmx/player-anim/player-animation-lib-forge/1.0.2+1.19.4/player-animation-lib-forge-1.0.2+1.19.4.jar' "$PLAYER_FORGE"
download_jar 'Curios Forge 5.14.1+1.20.1' 'https://maven.theillusivec4.top/top/theillusivec4/curios/curios-forge/5.14.1+1.20.1/curios-forge-5.14.1+1.20.1.jar' "$CURIOS_FORGE"

for pair in \
  "Ranged Weapon API common|$RANGED_COMMON" "Ranged Weapon API Forge|$RANGED_FORGE" \
  "Spell Power common|$SPELL_POWER_COMMON" "Spell Power Forge|$SPELL_POWER_FORGE" \
  "TinyConfig common|$TINY_COMMON" "TinyConfig Forge|$TINY_FORGE" \
  "Spell Engine common|$SPELL_ENGINE_COMMON" "Spell Engine Forge|$SPELL_ENGINE_FORGE" \
  "Structure Pool API common|$STRUCTURE_COMMON" "Structure Pool API Forge|$STRUCTURE_FORGE" \
  "Bundle API common|$BUNDLE_COMMON" "Bundle API Forge|$BUNDLE_FORGE" \
  "Armor Model API common|$ARMOR_COMMON" "Armor Model API Forge|$ARMOR_FORGE" \
  "Archers release|$ARCHERS_JAR"; do
  validate_jar "${pair%%|*}" "${pair#*|}"
done

ARGS=(
  "-Pbundle_api_common_jar=$BUNDLE_COMMON"
  "-Parmor_model_api_common_jar=$ARMOR_COMMON"
  "-Pranged_weapon_api_common_jar=$RANGED_COMMON"
  "-Pstructure_pool_api_common_jar=$STRUCTURE_COMMON"
  "-Pspell_power_common_jar=$SPELL_POWER_COMMON"
  "-Pspell_engine_common_jar=$SPELL_ENGINE_COMMON"
  "-Ptiny_config_common_jar=$TINY_COMMON"
  "-Pbundle_api_forge_jar=$BUNDLE_FORGE"
  "-Parmor_model_api_forge_jar=$ARMOR_FORGE"
  "-Pranged_weapon_api_forge_jar=$RANGED_FORGE"
  "-Pstructure_pool_api_forge_jar=$STRUCTURE_FORGE"
  "-Pspell_power_forge_jar=$SPELL_POWER_FORGE"
  "-Pspell_engine_forge_jar=$SPELL_ENGINE_FORGE"
  "-Ptiny_config_forge_jar=$TINY_FORGE"
  "-Pcloth_config_forge_jar=$CLOTH_FORGE"
  "-Pplayer_animator_forge_jar=$PLAYER_FORGE"
  "-Pcurios_jar=$CURIOS_FORGE"
)

FIRST_SHA="$(sha256sum "$ARCHERS_JAR" | awk '{print $1}')"
echo "[Archers acceptance] First packaged SHA256=$FIRST_SHA"

echo '[Archers acceptance] Deterministic clean remap gate'
gradle --no-daemon --stacktrace -p "$PORT" clean :forge:remapJar "${ARGS[@]}"
ARCHERS_JAR="$(pick_one "$PORT/forge/build/libs" '*-forge-*.jar' 'Archers Forge deterministic rebuild')"
SECOND_SHA="$(sha256sum "$ARCHERS_JAR" | awk '{print $1}')"
[[ "$FIRST_SHA" = "$SECOND_SHA" ]] || { echo "[Archers acceptance] non-deterministic release: $FIRST_SHA != $SECOND_SHA" >&2; exit 3; }
printf '%s  %s\n' "$SECOND_SHA" "$ARCHERS_JAR" | tee "$PORT/archers-deterministic.sha256"

FATAL='ARCHERS_SELF_TEST_FAILED|MixinApplyError|InvalidMixinException|MixinTransformerError|Failed to create mod instance|NoClassDefFoundError|ClassNotFoundException|Missing or unsupported mandatory dependencies|Exception in server tick loop|The game crashed|Attempted to load class .* for invalid dist DEDICATED_SERVER'

echo '[Archers acceptance] Real Forge dev-server + Archers semantic self-test'
rm -rf "$PORT/forge/run/logs"; mkdir -p "$PORT/forge/run"; printf 'eula=true\n' > "$PORT/forge/run/eula.txt"
SERVER_SMOKE="$PORT/archers-server-smoke.log"; : > "$SERVER_SMOKE"
set +e
env ARCHERS_SELF_TEST=1 gradle --no-daemon -p "$PORT" :forge:runServer "${ARGS[@]}" > "$SERVER_SMOKE" 2>&1 & SERVER_PID=$!
set -e
SERVER_DEADLINE=$((SECONDS+180)); SERVER_READY=0; SERVER_SELFTEST=0
while kill -0 "$SERVER_PID" 2>/dev/null && (( SECONDS < SERVER_DEADLINE )); do
  mapfile -t SERVER_FILES < <(collect_run_files "$SERVER_SMOKE" "$PORT/forge/run")
  if grep -Eiq "$FATAL" "${SERVER_FILES[@]}"; then stop_tree "$SERVER_PID"; cat "${SERVER_FILES[@]}"; exit 4; fi
  grep -Ehq 'Done \([0-9.]+s\)!' "${SERVER_FILES[@]}" && SERVER_READY=1 || true
  grep -Fhq 'ARCHERS_SELF_TEST_PASS' "${SERVER_FILES[@]}" && SERVER_SELFTEST=1 || true
  [[ "$SERVER_READY" = 1 && "$SERVER_SELFTEST" = 1 ]] && break
  sleep 2
done
if kill -0 "$SERVER_PID" 2>/dev/null; then stop_tree "$SERVER_PID"; fi
mapfile -t SERVER_FILES < <(collect_run_files "$SERVER_SMOKE" "$PORT/forge/run")
if grep -Eiq "$FATAL" "${SERVER_FILES[@]}"; then cat "${SERVER_FILES[@]}"; exit 4; fi
if [[ "$SERVER_READY" != 1 || "$SERVER_SELFTEST" != 1 ]]; then
  cat "${SERVER_FILES[@]}"; echo "[Archers acceptance] dev server proof incomplete: ready=$SERVER_READY selftest=$SERVER_SELFTEST" >&2; exit 4
fi
echo '[Archers acceptance] Dev server reached ready state and semantic self-test passed.'

echo '[Archers acceptance] Headless Forge client/resource/render/mixin bootstrap'
rm -rf "$PORT/forge/run/logs"; mkdir -p "$PORT/forge/run/config"
printf 'earlyWindowControl = false\n' > "$PORT/forge/run/config/fml.toml"
CLIENT_SMOKE="$PORT/archers-client-smoke.log"; : > "$CLIENT_SMOKE"
set +e
env LIBGL_ALWAYS_SOFTWARE=1 MESA_LOADER_DRIVER_OVERRIDE=llvmpipe \
  xvfb-run -a -s '-screen 0 1280x720x24 +extension GLX +extension RENDER -noreset' \
  gradle --no-daemon -p "$PORT" :forge:runClient "${ARGS[@]}" </dev/null > "$CLIENT_SMOKE" 2>&1 & CLIENT_PID=$!
set -e
CLIENT_DEADLINE=$((SECONDS+180)); CLIENT_RESOURCE=0; CLIENT_LWJGL=0
FATAL_CLIENT='MixinApplyError|InvalidMixinException|MixinTransformerError|Failed to create mod instance|NoClassDefFoundError|ClassNotFoundException|The game crashed whilst initializing game|Exception in thread "Render thread"|Failed to initialize graphics window|Timed out trying to setup the Game Window|Could not initialize GLFW|Missing or unsupported mandatory dependencies|archers_spirit.*(Exception|Error)|Direwolf.*(Exception|Error)'
while kill -0 "$CLIENT_PID" 2>/dev/null && (( SECONDS < CLIENT_DEADLINE )); do
  mapfile -t CLIENT_FILES < <(collect_run_files "$CLIENT_SMOKE" "$PORT/forge/run")
  if grep -Eiq "$FATAL_CLIENT" "${CLIENT_FILES[@]}"; then stop_tree "$CLIENT_PID"; cat "${CLIENT_FILES[@]}"; exit 5; fi
  grep -Fhq 'Reloading ResourceManager' "${CLIENT_FILES[@]}" && CLIENT_RESOURCE=1 || true
  grep -Fhq 'Backend library: LWJGL' "${CLIENT_FILES[@]}" && CLIENT_LWJGL=1 || true
  [[ "$CLIENT_RESOURCE" = 1 && "$CLIENT_LWJGL" = 1 ]] && break
  sleep 2
done
if kill -0 "$CLIENT_PID" 2>/dev/null; then stop_tree "$CLIENT_PID"; fi
mapfile -t CLIENT_FILES < <(collect_run_files "$CLIENT_SMOKE" "$PORT/forge/run")
if grep -Eiq "$FATAL_CLIENT" "${CLIENT_FILES[@]}"; then cat "${CLIENT_FILES[@]}"; exit 5; fi
if [[ "$CLIENT_RESOURCE" != 1 || "$CLIENT_LWJGL" != 1 ]]; then
  cat "${CLIENT_FILES[@]}"; echo "[Archers acceptance] client proof incomplete: resource=$CLIENT_RESOURCE lwjgl=$CLIENT_LWJGL" >&2; exit 5
fi
echo '[Archers acceptance] Headless client reached post-bootstrap resource/render runtime.'

echo '[Archers acceptance] Fresh official Forge 47.4.23 packaged-server gate'
FRESH="$PORT/.fresh-archers-forge-server"
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
  cp "$ARCHERS_JAR" "$ARMOR_FORGE" "$BUNDLE_FORGE" "$RANGED_FORGE" "$STRUCTURE_FORGE" \
     "$SPELL_POWER_FORGE" "$SPELL_ENGINE_FORGE" "$TINY_FORGE" "$CLOTH_FORGE" "$PLAYER_FORGE" "$CURIOS_FORGE" mods/
)
INSTALLED="$FRESH/mods/$(basename "$ARCHERS_JAR")"
cmp -s "$ARCHERS_JAR" "$INSTALLED"
INSTALLED_SHA="$(sha256sum "$INSTALLED" | awk '{print $1}')"
[[ "$INSTALLED_SHA" = "$SECOND_SHA" ]] || { echo '[Archers acceptance] installed Archers JAR bytes changed' >&2; exit 6; }
printf '%s  %s\n' "$INSTALLED_SHA" "$INSTALLED" | tee "$PORT/archers-package-installed.sha256"
PACKAGE_LOG="$PORT/archers-package-server-smoke.log"; : > "$PACKAGE_LOG"
set +e
( cd "$FRESH" && exec env ARCHERS_SELF_TEST=1 ./run.sh nogui ) > "$PACKAGE_LOG" 2>&1 & PID=$!
set -e
DEADLINE=$((SECONDS+180)); READY=0; SELFTEST=0
while kill -0 "$PID" 2>/dev/null && (( SECONDS < DEADLINE )); do
  if grep -Eiq "$FATAL" "$PACKAGE_LOG"; then stop_tree "$PID"; cat "$PACKAGE_LOG"; exit 6; fi
  grep -Eq 'Done \([0-9.]+s\)!' "$PACKAGE_LOG" && READY=1 || true
  grep -Fq 'ARCHERS_SELF_TEST_PASS' "$PACKAGE_LOG" && SELFTEST=1 || true
  [[ "$READY" = 1 && "$SELFTEST" = 1 ]] && break
  sleep 2
done
if kill -0 "$PID" 2>/dev/null; then stop_tree "$PID"; fi
if [[ "$READY" != 1 || "$SELFTEST" != 1 ]]; then cat "$PACKAGE_LOG"; echo "[Archers acceptance] fresh server proof incomplete: ready=$READY selftest=$SELFTEST" >&2; exit 6; fi
if grep -Eiq "$FATAL" "$PACKAGE_LOG"; then cat "$PACKAGE_LOG"; exit 6; fi
echo '[Archers acceptance] Fresh packaged Forge server reached ready state with byte-identical Archers JAR and semantic self-test green.'

echo '[Archers acceptance] PASS: deterministic package + semantic dev server + headless client + fresh packaged server.'
