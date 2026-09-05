#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
PORT="$ROOT/rpg-series-port/more-rpg-library-forge-1.20.1"
FOUNDATION="$PORT/.foundation"
HELPER="$PORT/tools/prepare_forge_production_client.py"
OUT="$PORT/more_rpg_library-forge-2.7.2+1.20.1.jar"
SPELL_ENGINE_JAR="$ROOT/rpg-series-port/spell-engine-forge-1.20.1/spell_engine-forge-1.10.4+1.20.1.jar"
SPELL_POWER_JAR="$FOUNDATION/spell_power-forge-1.6.0+1.20.1-certified.jar"
RANGED_JAR="$FOUNDATION/ranged_weapon_api-forge-2.3.4+1.20.1-certified.jar"
TINY_JAR="$(find "$ROOT/rpg-series-port/tiny-config-forge-1.20.1/generated/forge/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' ! -name '*dev-shadow*' | sort | head -n1)"
RUNTIME_DEPS="$PORT/.more-rpg-runtime-deps"
CLOTH_FORGE_JAR="$RUNTIME_DEPS/cloth-config-forge-11.1.106.jar"
PLAYER_ANIM_FORGE_JAR="$RUNTIME_DEPS/player-animation-lib-forge-1.0.2+1.19.4.jar"
FRESH_SERVER="$PORT/.fresh-more-rpg-server"
FORGE_VERSION='1.20.1-47.4.23'
FORGE_PROFILE='1.20.1-forge-47.4.23'
# Official Forge 1.20.1 downloads page identity for 47.4.23 installer (2026-08-19).
FORGE_INSTALLER_SHA1='ed31ce02ac69176f34353235cb2508d5a0f1e088'
MC_HOME="$PORT/.production-minecraft-home"
GAME_DIR="$PORT/.production-client-run"
INSTALLER="$MC_HOME/forge-${FORGE_VERSION}-installer.jar"
LAUNCH_SCRIPT="$GAME_DIR/launch-forgeclient.sh"
CLIENT_LOG="$PORT/more-rpg-production-forgeclient.log"

for f in "$HELPER" "$OUT" "$SPELL_ENGINE_JAR" "$SPELL_POWER_JAR" "$RANGED_JAR" "$TINY_JAR" "$CLOTH_FORGE_JAR" "$PLAYER_ANIM_FORGE_JAR"; do
  test -f "$f"
  case "$f" in *.jar) unzip -tq "$f" >/dev/null ;; esac
done
python3 -m py_compile "$HELPER"
[[ -d "$FRESH_SERVER/world/datapacks/more-rpg-runtime-qa" ]] || {
  echo '[More RPG 2.7.2] Stage-2 requires the Stage-1 frozen QA world/datapack checkpoint' >&2
  exit 1
}

echo '[More RPG 2.7.2] PRODUCTION_CLIENT_STAGE2_BEGIN target=forgeclient namespace=SRG'
rm -rf "$MC_HOME" "$GAME_DIR"
mkdir -p "$MC_HOME" "$GAME_DIR/mods" "$GAME_DIR/config" "$GAME_DIR/saves"

# Build a complete Mojang launcher cache first. The helper verifies the official 1.20.1 client JAR,
# library hashes, asset index and every referenced asset object; this is exhaustive, not sampled.
python3 "$HELPER" bootstrap --mc-home "$MC_HOME" --minecraft 1.20.1 --asset-workers 16
curl -fsSL "https://maven.minecraftforge.net/net/minecraftforge/forge/${FORGE_VERSION}/forge-${FORGE_VERSION}-installer.jar" -o "$INSTALLER"
unzip -tq "$INSTALLER" >/dev/null
[[ "$(sha1sum "$INSTALLER" | awk '{print $1}')" = "$FORGE_INSTALLER_SHA1" ]]
echo "[More RPG 2.7.2] FORGE_INSTALLER_IDENTITY_PASS version=$FORGE_VERSION sha1=$FORGE_INSTALLER_SHA1"

# Use Forge's real client installer/profile. ForgeGradle userdev is forbidden in this gate.
java -jar "$INSTALLER" --installClient "$MC_HOME" > "$PORT/more-rpg-forge-install-client.log" 2>&1
FORGE_JSON="$MC_HOME/versions/$FORGE_PROFILE/$FORGE_PROFILE.json"
test -f "$FORGE_JSON"
grep -Fq 'cpw.mods.bootstraplauncher.BootstrapLauncher' "$FORGE_JSON"
grep -Fq 'forgeclient' "$FORGE_JSON"
! grep -Fq 'forgeclientuserdev' "$FORGE_JSON"
echo '[More RPG 2.7.2] FORGE_PRODUCTION_PROFILE_INSTALL_PASS launchTarget=forgeclient'

# Copy only exact packaged/certified mod JARs. No source-set registration or development mod path.
cp "$OUT" "$GAME_DIR/mods/more-rpg-library-forge.jar"
cp "$SPELL_ENGINE_JAR" "$GAME_DIR/mods/spell-engine-forge.jar"
cp "$SPELL_POWER_JAR" "$GAME_DIR/mods/spell-power-forge.jar"
cp "$RANGED_JAR" "$GAME_DIR/mods/ranged-weapon-api-forge.jar"
cp "$TINY_JAR" "$GAME_DIR/mods/tiny-config-forge.jar"
cp "$CLOTH_FORGE_JAR" "$GAME_DIR/mods/cloth-config-forge.jar"
cp "$PLAYER_ANIM_FORGE_JAR" "$GAME_DIR/mods/player-animation-lib-forge.jar"
(
  cd "$GAME_DIR/mods"
  sha256sum *.jar | sort > "$GAME_DIR/more-rpg-production-mods.sha256"
)
for f in "$GAME_DIR"/mods/*.jar; do unzip -tq "$f" >/dev/null; done
[[ "$(sha256sum "$GAME_DIR/mods/more-rpg-library-forge.jar" | awk '{print $1}')" = "$(sha256sum "$OUT" | awk '{print $1}')" ]]
[[ "$(sha256sum "$GAME_DIR/mods/spell-engine-forge.jar" | awk '{print $1}')" = "$(sha256sum "$SPELL_ENGINE_JAR" | awk '{print $1}')" ]]
echo '[More RPG 2.7.2] PRODUCTION_PACKAGED_MOD_BYTES_LOCKED count=7'

cp -a "$FRESH_SERVER/world" "$GAME_DIR/saves/MRPG-QA"
printf 'earlyWindowControl = false\n' > "$GAME_DIR/config/fml.toml"
printf 'onboardAccessibility:false\n' > "$GAME_DIR/options.txt"

python3 "$HELPER" prepare \
  --mc-home "$MC_HOME" \
  --game-dir "$GAME_DIR" \
  --minecraft 1.20.1 \
  --forge-version-id "$FORGE_PROFILE" \
  --quick-play MRPG-QA \
  --launch-script "$LAUNCH_SCRIPT" \
  --java "$(command -v java)" \
  --asset-workers 16
bash -n "$LAUNCH_SCRIPT"

grep -Fq '"launch_target": "forgeclient"' "$GAME_DIR/more-rpg-production-launch.json"
grep -Fq 'cpw.mods.bootstraplauncher.BootstrapLauncher' "$LAUNCH_SCRIPT"
grep -Fq -- '--launchTarget forgeclient' "$LAUNCH_SCRIPT"
! grep -Fq 'forgeclientuserdev' "$LAUNCH_SCRIPT"
echo '[More RPG 2.7.2] PRODUCTION_FORGECLIENT_COMMAND_AUDIT_PASS'

ACTIVE_PID=''
descendants() {
  local parent="$1" child
  while IFS= read -r child; do
    [[ -z "$child" ]] && continue
    descendants "$child"
    printf '%s\n' "$child"
  done < <(pgrep -P "$parent" 2>/dev/null || true)
}
stop_tree() {
  local root="${1:-}"; [[ -n "$root" ]] || return 0
  local -a children=(); mapfile -t children < <(descendants "$root")
  ((${#children[@]})) && kill -TERM "${children[@]}" 2>/dev/null || true
  kill -TERM "$root" 2>/dev/null || true
  for _ in {1..20}; do kill -0 "$root" 2>/dev/null || break; sleep 0.1; done
  mapfile -t children < <(descendants "$root")
  ((${#children[@]})) && kill -KILL "${children[@]}" 2>/dev/null || true
  kill -KILL "$root" 2>/dev/null || true
  wait "$root" 2>/dev/null || true
}
cleanup() { [[ -n "${ACTIVE_PID:-}" ]] && stop_tree "$ACTIVE_PID" || true; }
trap cleanup EXIT INT TERM

rm -rf "$GAME_DIR/.mixin.out" "$GAME_DIR/logs"
: > "$CLIENT_LOG"
export JAVA_TOOL_OPTIONS="-Dmixin.debug.export=true -Dmixin.debug.export.filter=net.minecraft.client.gui.** -Dmixin.debug.export.decompile=false -Dmixin.debug.verbose=true${JAVA_TOOL_OPTIONS:+ $JAVA_TOOL_OPTIONS}"
env LIBGL_ALWAYS_SOFTWARE=1 MESA_LOADER_DRIVER_OVERRIDE=llvmpipe ALSOFT_DRIVERS=null \
  xvfb-run -a -s '-screen 0 1280x720x24 +extension GLX +extension RENDER -noreset' \
  "$LAUNCH_SCRIPT" </dev/null > "$CLIENT_LOG" 2>&1 &
ACTIVE_PID=$!
CLIENT_PID=$ACTIVE_PID
DEADLINE=$((SECONDS+300))
PASS=0
FATAL='ModLoadingException|Failed to create mod instance|NoClassDefFoundError|ClassNotFoundException|MixinApplyError|InvalidMixinException|MixinTransformerError|Exception in thread "Render thread"|The game crashed|Missing or unsupported mandatory dependencies|Could not initialize GLFW|Failed to initialize graphics window|Timed out trying to setup the Game Window'
while ((SECONDS<DEADLINE)); do
  LATEST="$GAME_DIR/logs/latest.log"
  FILES=("$CLIENT_LOG"); [[ -f "$LATEST" ]] && FILES+=("$LATEST")
  if grep -Eiq "$FATAL" "${FILES[@]}"; then
    stop_tree "$CLIENT_PID"; ACTIVE_PID=''; tail -n 700 "${FILES[@]}" || true; exit 1
  fi
  if grep -Fq 'forgeclientuserdev' "${FILES[@]}"; then
    echo '[More RPG 2.7.2] forbidden development launch target observed in production log' >&2
    stop_tree "$CLIENT_PID"; ACTIVE_PID=''; exit 1
  fi
  if [[ -f "$LATEST" ]] \
    && grep -Fq 'forgeclient' "${FILES[@]}" \
    && grep -Fq 'Reloading ResourceManager' "$LATEST" \
    && grep -Fq 'Backend library: LWJGL' "$LATEST" \
    && grep -Fq 'more_rpg_classes' "${FILES[@]}" \
    && grep -Fq '[More RPG QA] FATAL_POISON_APPLIED' "${FILES[@]}"; then
    PASS=1; break
  fi
  if ! kill -0 "$CLIENT_PID" 2>/dev/null; then
    wait "$CLIENT_PID" || true; ACTIVE_PID=''; tail -n 700 "${FILES[@]}" || true; exit 1
  fi
  sleep 1
done
[[ "$PASS" -eq 1 ]] || {
  stop_tree "$CLIENT_PID"; ACTIVE_PID=''; tail -n 700 "$CLIENT_LOG" "$GAME_DIR/logs/latest.log" 2>/dev/null || true; exit 1
}
stop_tree "$CLIENT_PID"; ACTIVE_PID=''

LATEST="$GAME_DIR/logs/latest.log"
grep -Fq 'DrawHeartsMixin' "$CLIENT_LOG" "$LATEST"
mapfile -t HEART_TARGETS < <(grep -aRl 'net/more_rpg_classes/client/heart/HeartRegistry' "$GAME_DIR/.mixin.out" --include='*.class' 2>/dev/null | sort || true)
((${#HEART_TARGETS[@]} > 0)) || {
  echo '[More RPG 2.7.2] production transformed HUD proof missing HeartRegistry reference' >&2
  exit 1
}
printf '[More RPG 2.7.2] PRODUCTION_DRAW_HEARTS_MIXIN_TRANSFORM_PASS targets=%s first=%s\n' \
  "${#HEART_TARGETS[@]}" "${HEART_TARGETS[0]#$GAME_DIR/.mixin.out/}"
echo '[More RPG 2.7.2] PRODUCTION_FORGECLIENT_STAGE2_PASS forge=47.4.23 minecraft=1.20.1 namespace=SRG integrated_world=true fatal_poison=true transformed_hud=true restart_persistence_pending=true'
