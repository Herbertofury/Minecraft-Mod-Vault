#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
PORT="$ROOT/rpg-series-port/more-rpg-library-forge-1.20.1"
WORK="$ROOT/.more-rpg-library-build"
FOUNDATION="$PORT/.foundation"
RUN="$WORK/forge/run"; SAVE='MRPG-QA'; WORLD="$RUN/saves/$SAVE"; QA_PACK="$WORLD/datapacks/more-rpg-runtime-qa"
OUT="$PORT/more_rpg_library-forge-2.7.2+1.20.1.jar"
SPELL_ENGINE_JAR="$ROOT/rpg-series-port/spell-engine-forge-1.20.1/spell_engine-forge-1.10.4+1.20.1.jar"; SPELL_POWER_JAR="$FOUNDATION/spell_power-forge-1.6.0+1.20.1-certified.jar"; RANGED_JAR="$FOUNDATION/ranged_weapon_api-forge-2.3.4+1.20.1-certified.jar"
TINY_JAR="$(find "$ROOT/rpg-series-port/tiny-config-forge-1.20.1/generated/forge/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' ! -name '*dev-shadow*' | sort | head -n1)"
SPELL_ENGINE_COMMON_JAR="$(find "$ROOT/.spell-engine-build/common/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' | sort | head -n1)"; SPELL_POWER_COMMON_JAR="$(find "$ROOT/rpg-series-port/spell_power-forge-1.20.1/common/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' | sort | head -n1)"; RANGED_COMMON_JAR="$(find "$ROOT/rpg-series-port/ranged-weapon-api-forge-1.20.1/common/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' | sort | head -n1)"; TINY_COMMON_JAR="$(find "$ROOT/rpg-series-port/tiny-config-forge-1.20.1/generated/common/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' | sort | head -n1)"
for f in "$OUT" "$SPELL_ENGINE_JAR" "$SPELL_POWER_JAR" "$RANGED_JAR" "$TINY_JAR" "$SPELL_ENGINE_COMMON_JAR" "$SPELL_POWER_COMMON_JAR" "$RANGED_COMMON_JAR" "$TINY_COMMON_JAR"; do test -f "$f"; unzip -tq "$f" >/dev/null; done
[[ -d "$WORLD" && -d "$QA_PACK" ]] || { echo '[More RPG 2.7.2] Stage-1 integrated world missing for persistence gate' >&2; exit 1; }
mapfile -t PLAYERDATA < <(find "$WORLD/playerdata" -maxdepth 1 -type f -name '*.dat' 2>/dev/null | sort || true); ((${#PLAYERDATA[@]} > 0)) || { echo '[More RPG 2.7.2] Stage-1 produced no persisted playerdata file' >&2; exit 1; }
BEFORE_SHA="$(sha256sum "${PLAYERDATA[0]}" | awk '{print $1}')"; printf '[More RPG 2.7.2] RESTART_PERSISTENCE_PLAYERDATA_PRESENT_PASS file=%s sha256=%s\n' "${PLAYERDATA[0]#$WORLD/}" "$BEFORE_SHA"
cat > "$QA_PACK/data/more_rpg_qa/functions/tick.mcfunction" <<'MCFUNCTION'
execute as @a[tag=mrpg_qa_fatal_poison,tag=!mrpg_qa_restart_seen] run tellraw @a {"text":"[More RPG QA] PERSISTED_PLAYER_TAG_SEEN"}
execute as @a[tag=mrpg_qa_fatal_poison,tag=!mrpg_qa_restart_seen] run tag @s add mrpg_qa_restart_seen
MCFUNCTION
! grep -Fq 'effect give' "$QA_PACK/data/more_rpg_qa/functions/tick.mcfunction"
ARGS=("-Pspell_engine_common_jar=$SPELL_ENGINE_COMMON_JAR" "-Pspell_power_common_jar=$SPELL_POWER_COMMON_JAR" "-Pranged_weapon_api_common_jar=$RANGED_COMMON_JAR" "-Ptiny_config_common_jar=$TINY_COMMON_JAR" "-Pspell_engine_forge_jar=$SPELL_ENGINE_JAR" "-Pspell_power_forge_jar=$SPELL_POWER_JAR" "-Pranged_weapon_api_forge_jar=$RANGED_JAR" "-Ptiny_config_forge_jar=$TINY_JAR")
rm -rf "$RUN/logs"; printf 'earlyWindowControl = false\n' > "$RUN/config/fml.toml"; LOG="$PORT/more-rpg-restart-persistence-client.log"; : > "$LOG"; ACTIVE_PID=''
descendants() { local parent="$1" child; while IFS= read -r child; do [[ -z "$child" ]] && continue; descendants "$child"; printf '%s\n' "$child"; done < <(pgrep -P "$parent" 2>/dev/null || true); }
stop_tree() { local root="${1:-}"; [[ -n "$root" ]] || return 0; local -a children=(); mapfile -t children < <(descendants "$root"); ((${#children[@]})) && kill -TERM "${children[@]}" 2>/dev/null || true; kill -TERM "$root" 2>/dev/null || true; for _ in {1..40}; do kill -0 "$root" 2>/dev/null || break; sleep 0.1; done; mapfile -t children < <(descendants "$root"); ((${#children[@]})) && kill -KILL "${children[@]}" 2>/dev/null || true; kill -KILL "$root" 2>/dev/null || true; wait "$root" 2>/dev/null || true; }
cleanup() { [[ -z "${ACTIVE_PID:-}" ]] || stop_tree "$ACTIVE_PID"; }; trap cleanup EXIT INT TERM
env LIBGL_ALWAYS_SOFTWARE=1 MESA_LOADER_DRIVER_OVERRIDE=llvmpipe ALSOFT_DRIVERS=null xvfb-run -a -s '-screen 0 1280x720x24 +extension GLX +extension RENDER -noreset' gradle --no-daemon -p "$WORK" :forge:runClient --args="--width 1280 --height 720 --quickPlaySingleplayer $SAVE" "${ARGS[@]}" </dev/null > "$LOG" 2>&1 &
ACTIVE_PID=$!; PID=$ACTIVE_PID; DEADLINE=$((SECONDS+240)); PASS=0
FATAL='ModLoadingException|Failed to create mod instance|NoClassDefFoundError|ClassNotFoundException|MixinApplyError|InvalidMixinException|MixinTransformerError|Exception in thread "Render thread"|The game crashed|Missing or unsupported mandatory dependencies|Could not initialize GLFW|Failed to initialize graphics window|Timed out trying to setup the Game Window'
while ((SECONDS<DEADLINE)); do LATEST="$RUN/logs/latest.log"; FILES=("$LOG"); [[ -f "$LATEST" ]] && FILES+=("$LATEST"); if grep -Eiq "$FATAL" "${FILES[@]}"; then tail -n 700 "${FILES[@]}" 2>/dev/null || true; exit 1; fi; if [[ -f "$LATEST" ]] && grep -Fq 'Reloading ResourceManager' "$LATEST" && grep -Fq 'Backend library: LWJGL' "$LATEST" && grep -Fq '[More RPG QA] PERSISTED_PLAYER_TAG_SEEN' "${FILES[@]}"; then PASS=1; break; fi; if ! kill -0 "$PID" 2>/dev/null; then wait "$PID" || true; ACTIVE_PID=''; tail -n 700 "${FILES[@]}" 2>/dev/null || true; exit 1; fi; sleep 1; done
[[ "$PASS" -eq 1 ]] || { tail -n 700 "$LOG" "$RUN/logs/latest.log" 2>/dev/null || true; exit 1; }; sleep 3; stop_tree "$PID"; ACTIVE_PID=''; LATEST="$RUN/logs/latest.log"; ! grep -Eiq "$FATAL" "$LOG" "$LATEST"
mapfile -t AFTER_PLAYERDATA < <(find "$WORLD/playerdata" -maxdepth 1 -type f -name '*.dat' 2>/dev/null | sort || true); ((${#AFTER_PLAYERDATA[@]} > 0)); AFTER_SHA="$(sha256sum "${AFTER_PLAYERDATA[0]}" | awk '{print $1}')"; printf '[More RPG 2.7.2] SAME_SAVE_RESTART_PERSISTENCE_PASS marker=PERSISTED_PLAYER_TAG_SEEN playerdata_before=%s playerdata_after=%s\n' "$BEFORE_SHA" "$AFTER_SHA"
