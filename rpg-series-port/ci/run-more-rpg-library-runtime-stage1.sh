#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
PORT="$ROOT/rpg-series-port/more-rpg-library-forge-1.20.1"
WORK="$ROOT/.more-rpg-library-build"
FOUNDATION="$PORT/.foundation"
OUT="$PORT/more_rpg_library-forge-2.7.2+1.20.1.jar"
SPELL_ENGINE_JAR="$ROOT/rpg-series-port/spell-engine-forge-1.20.1/spell_engine-forge-1.10.4+1.20.1.jar"
SPELL_POWER_JAR="$FOUNDATION/spell_power-forge-1.6.0+1.20.1-certified.jar"
RANGED_JAR="$FOUNDATION/ranged_weapon_api-forge-2.3.4+1.20.1-certified.jar"
TINY_JAR="$(find "$ROOT/rpg-series-port/tiny-config-forge-1.20.1/generated/forge/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' ! -name '*dev-shadow*' | sort | head -n1)"
SPELL_ENGINE_COMMON_JAR="$(find "$ROOT/.spell-engine-build/common/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' | sort | head -n1)"
SPELL_POWER_COMMON_JAR="$(find "$ROOT/rpg-series-port/spell_power-forge-1.20.1/common/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' | sort | head -n1)"
RANGED_COMMON_JAR="$(find "$ROOT/rpg-series-port/ranged-weapon-api-forge-1.20.1/common/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' | sort | head -n1)"
TINY_COMMON_JAR="$(find "$ROOT/rpg-series-port/tiny-config-forge-1.20.1/generated/common/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' | sort | head -n1)"
for f in "$OUT" "$SPELL_ENGINE_JAR" "$SPELL_POWER_JAR" "$RANGED_JAR" "$TINY_JAR" "$SPELL_ENGINE_COMMON_JAR" "$SPELL_POWER_COMMON_JAR" "$RANGED_COMMON_JAR" "$TINY_COMMON_JAR"; do
  test -f "$f"; unzip -tq "$f" >/dev/null
done
ARGS=(
  "-Pspell_engine_common_jar=$SPELL_ENGINE_COMMON_JAR"
  "-Pspell_power_common_jar=$SPELL_POWER_COMMON_JAR"
  "-Pranged_weapon_api_common_jar=$RANGED_COMMON_JAR"
  "-Ptiny_config_common_jar=$TINY_COMMON_JAR"
  "-Pspell_engine_forge_jar=$SPELL_ENGINE_JAR"
  "-Pspell_power_forge_jar=$SPELL_POWER_JAR"
  "-Pranged_weapon_api_forge_jar=$RANGED_JAR"
  "-Ptiny_config_forge_jar=$TINY_JAR"
)
echo '[More RPG 2.7.2] RUNTIME_STAGE1_BEGIN'

ACTIVE_PID=""
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
cleanup_more_rpg_runtime() { [[ -n "${ACTIVE_PID:-}" ]] && stop_tree "$ACTIVE_PID" || true; }
trap cleanup_more_rpg_runtime EXIT INT TERM

dump_runtime_logs() {
  local f
  for f in "$@"; do
    [[ -f "$f" ]] || continue
    echo "===== tail: $f ====="
    tail -n 500 "$f" || true
  done
}

RUNTIME_DEPS="$PORT/.more-rpg-runtime-deps"; rm -rf "$RUNTIME_DEPS"; mkdir -p "$RUNTIME_DEPS"
CLOTH_FORGE_JAR="$RUNTIME_DEPS/cloth-config-forge-11.1.106.jar"
PLAYER_ANIM_FORGE_JAR="$RUNTIME_DEPS/player-animation-lib-forge-1.0.2+1.19.4.jar"
curl -fsSL 'https://maven.shedaniel.me/me/shedaniel/cloth/cloth-config-forge/11.1.106/cloth-config-forge-11.1.106.jar' -o "$CLOTH_FORGE_JAR"
curl -fsSL 'https://maven.kosmx.dev/dev/kosmx/player-anim/player-animation-lib-forge/1.0.2+1.19.4/player-animation-lib-forge-1.0.2+1.19.4.jar' -o "$PLAYER_ANIM_FORGE_JAR"
for f in "$CLOTH_FORGE_JAR" "$PLAYER_ANIM_FORGE_JAR"; do unzip -tq "$f" >/dev/null; done
unzip -p "$CLOTH_FORGE_JAR" META-INF/mods.toml | grep -Eq '^[[:space:]]*modId[[:space:]]*=[[:space:]]*"cloth_config"[[:space:]]*$'
unzip -p "$PLAYER_ANIM_FORGE_JAR" META-INF/mods.toml | grep -Eq '^[[:space:]]*modId[[:space:]]*=[[:space:]]*"playeranimator"[[:space:]]*$'
echo '[More RPG 2.7.2] PACKAGED_RUNTIME_DEPENDENCIES_RESOLVED_EXACT cloth=11.1.106 player_anim=1.0.2+1.19.4'

FRESH_SERVER="$PORT/.fresh-more-rpg-server"; rm -rf "$FRESH_SERVER"; mkdir -p "$FRESH_SERVER/mods"
curl -fsSL 'https://maven.minecraftforge.net/net/minecraftforge/forge/1.20.1-47.4.23/forge-1.20.1-47.4.23-installer.jar' -o "$FRESH_SERVER/forge-installer.jar"
(
  cd "$FRESH_SERVER"
  java -jar forge-installer.jar --installServer >/dev/null
  printf 'eula=true\n' > eula.txt
  printf '%s\n' '-Xmx3G' > user_jvm_args.txt
  cp "$OUT" mods/more-rpg-library-forge.jar
  cp "$SPELL_ENGINE_JAR" mods/spell-engine-forge.jar
  cp "$SPELL_POWER_JAR" mods/spell-power-forge.jar
  cp "$RANGED_JAR" mods/ranged-weapon-api-forge.jar
  cp "$TINY_JAR" mods/tiny-config-forge.jar
  cp "$CLOTH_FORGE_JAR" mods/cloth-config-forge.jar
  cp "$PLAYER_ANIM_FORGE_JAR" mods/player-animation-lib-forge.jar
)
SERVER_LOG="$PORT/more-rpg-packaged-server-smoke.log"; : > "$SERVER_LOG"
( cd "$FRESH_SERVER" && exec ./run.sh nogui ) > "$SERVER_LOG" 2>&1 & ACTIVE_PID=$!
SERVER_PID=$ACTIVE_PID; SERVER_DEADLINE=$((SECONDS+180)); SERVER_PASS=0
FATAL='ModLoadingException|Failed to create mod instance|NoClassDefFoundError|ClassNotFoundException|MixinApplyError|InvalidMixinException|MixinTransformerError|Exception in server tick loop|The game crashed|Missing or unsupported mandatory dependencies'
while ((SECONDS<SERVER_DEADLINE)); do
  LATEST="$FRESH_SERVER/logs/latest.log"; FILES=("$SERVER_LOG"); [[ -f "$LATEST" ]] && FILES+=("$LATEST")
  if grep -Eiq "$FATAL" "${FILES[@]}"; then stop_tree "$SERVER_PID"; ACTIVE_PID=""; dump_runtime_logs "${FILES[@]}"; exit 1; fi
  if [[ -f "$LATEST" ]] && grep -Eq 'Done \([0-9.]+s\)!' "$LATEST" && grep -Fq 'more_rpg_classes' "$LATEST"; then SERVER_PASS=1; break; fi
  if ! kill -0 "$SERVER_PID" 2>/dev/null; then wait "$SERVER_PID" || true; ACTIVE_PID=""; dump_runtime_logs "${FILES[@]}"; exit 1; fi
  sleep 1
done
[[ "$SERVER_PASS" -eq 1 ]] || { stop_tree "$SERVER_PID"; ACTIVE_PID=""; dump_runtime_logs "$SERVER_LOG" "$FRESH_SERVER/logs/latest.log"; exit 1; }
stop_tree "$SERVER_PID"; ACTIVE_PID=""
echo '[More RPG 2.7.2] PACKAGED_FORGE_SERVER_BOOT_PASS'

QA_PACK="$FRESH_SERVER/world/datapacks/more-rpg-runtime-qa"
mkdir -p "$QA_PACK/data/minecraft/tags/functions" "$QA_PACK/data/more_rpg_qa/functions"
cat > "$QA_PACK/pack.mcmeta" <<'JSON'
{"pack":{"pack_format":15,"description":"More RPG deterministic runtime QA"}}
JSON
cat > "$QA_PACK/data/minecraft/tags/functions/tick.json" <<'JSON'
{"values":["more_rpg_qa:tick"]}
JSON
cat > "$QA_PACK/data/more_rpg_qa/functions/tick.mcfunction" <<'MCFUNCTION'
execute as @a[tag=!mrpg_qa_fatal_poison] run effect give @s more_rpg_classes:fatal_poison 120 0 true
execute as @a[tag=!mrpg_qa_fatal_poison] run tellraw @a {"text":"[More RPG QA] FATAL_POISON_APPLIED"}
tag @a[tag=!mrpg_qa_fatal_poison] add mrpg_qa_fatal_poison
MCFUNCTION

CLIENT_RUN="$WORK/forge/run"; rm -rf "$CLIENT_RUN/logs" "$CLIENT_RUN/saves/MRPG-QA" "$CLIENT_RUN/.mixin.out"
mkdir -p "$CLIENT_RUN/config" "$CLIENT_RUN/saves"
cp -a "$FRESH_SERVER/world" "$CLIENT_RUN/saves/MRPG-QA"
printf 'earlyWindowControl = false\n' > "$CLIENT_RUN/config/fml.toml"
CLIENT_LOG="$PORT/more-rpg-native-client-integrated.log"; : > "$CLIENT_LOG"
MIXIN_JAVA_TOOL_OPTIONS='-Dmixin.debug.export=true -Dmixin.debug.export.filter=net.minecraft.client.gui.** -Dmixin.debug.export.decompile=false -Dmixin.debug.verbose=true'
env LIBGL_ALWAYS_SOFTWARE=1 MESA_LOADER_DRIVER_OVERRIDE=llvmpipe ALSOFT_DRIVERS=null JAVA_TOOL_OPTIONS="$MIXIN_JAVA_TOOL_OPTIONS ${JAVA_TOOL_OPTIONS:-}" \
  xvfb-run -a -s '-screen 0 1280x720x24 +extension GLX +extension RENDER -noreset' \
  gradle --no-daemon -p "$WORK" :forge:runClient --args='--width 1280 --height 720 --quickPlaySingleplayer MRPG-QA' "${ARGS[@]}" </dev/null > "$CLIENT_LOG" 2>&1 & ACTIVE_PID=$!
CLIENT_PID=$ACTIVE_PID; CLIENT_DEADLINE=$((SECONDS+240)); CLIENT_PASS=0
while ((SECONDS<CLIENT_DEADLINE)); do
  LATEST="$CLIENT_RUN/logs/latest.log"; FILES=("$CLIENT_LOG"); [[ -f "$LATEST" ]] && FILES+=("$LATEST")
  if grep -Eiq "$FATAL|Exception in thread \"Render thread\"|Could not initialize GLFW|Failed to initialize graphics window|Timed out trying to setup the Game Window" "${FILES[@]}"; then
    stop_tree "$CLIENT_PID"; ACTIVE_PID=""; dump_runtime_logs "${FILES[@]}"; exit 1
  fi
  if [[ -f "$LATEST" ]] && grep -Fq 'Reloading ResourceManager' "$LATEST" && grep -Fq 'Backend library: LWJGL' "$LATEST" && grep -Fq '[More RPG QA] FATAL_POISON_APPLIED' "${FILES[@]}"; then CLIENT_PASS=1; break; fi
  if ! kill -0 "$CLIENT_PID" 2>/dev/null; then wait "$CLIENT_PID" || true; ACTIVE_PID=""; dump_runtime_logs "${FILES[@]}"; exit 1; fi
  sleep 1
done
[[ "$CLIENT_PASS" -eq 1 ]] || { stop_tree "$CLIENT_PID"; ACTIVE_PID=""; dump_runtime_logs "$CLIENT_LOG" "$CLIENT_RUN/logs/latest.log"; exit 1; }

# Do not infer HUD execution from the poison marker. Mixin's debug exporter writes the post-transform
# target bytecode. Require both a verbose application record for DrawHeartsMixin and a transformed GUI
# class containing More RPG's HeartRegistry reference while the fatal-poison player is live.
LATEST="$CLIENT_RUN/logs/latest.log"
if ! grep -Eiq 'DrawHeartsMixin.*more-rpg-classes\.mixins\.json|more-rpg-classes\.mixins\.json.*DrawHeartsMixin' "$CLIENT_LOG" "$LATEST"; then
  stop_tree "$CLIENT_PID"; ACTIVE_PID=""; dump_runtime_logs "$CLIENT_LOG" "$LATEST"; echo '[More RPG 2.7.2] DrawHeartsMixin verbose application record missing' >&2; exit 1
fi
MIXIN_OUT="$CLIENT_RUN/.mixin.out"
[[ -d "$MIXIN_OUT" ]] || { stop_tree "$CLIENT_PID"; ACTIVE_PID=""; echo '[More RPG 2.7.2] Mixin post-transform export directory missing' >&2; exit 1; }
mapfile -t HEART_TARGETS < <(grep -aRl 'net/more_rpg_classes/client/heart/HeartRegistry' "$MIXIN_OUT" --include='*.class' 2>/dev/null || true)
if ((${#HEART_TARGETS[@]} == 0)); then
  stop_tree "$CLIENT_PID"; ACTIVE_PID=""; echo '[More RPG 2.7.2] No exported transformed GUI class contains HeartRegistry reference' >&2; find "$MIXIN_OUT" -type f -name '*.class' -print | sort | head -n 100 || true; exit 1
fi
printf '[More RPG 2.7.2] DRAW_HEARTS_MIXIN_TRANSFORMED_CLIENT_PASS targets=%s first=%s\n' "${#HEART_TARGETS[@]}" "${HEART_TARGETS[0]#$MIXIN_OUT/}"
stop_tree "$CLIENT_PID"; ACTIVE_PID=""
echo '[More RPG 2.7.2] NATIVE_FORGE_CLIENT_FATAL_POISON_HEART_PATH_PASS effect=more_rpg_classes:fatal_poison assets=6 transformed_hud=true'
echo '[More RPG 2.7.2] RUNTIME_STAGE1_PASS packaged_server=true mapped_client_integrated_world=true fatal_poison_gameplay=true heart_renderer_transformed=true production_client_pending=true'
