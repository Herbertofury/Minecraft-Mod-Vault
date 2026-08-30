#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
PORT="$ROOT/rpg-series-port/more-rpg-library-forge-1.20.1"
WORK="$ROOT/.more-rpg-library-build"
FOUNDATION="$PORT/.foundation"
SOURCE_WORLD="$PORT/.fresh-more-rpg-server/world"
OUT="$PORT/more_rpg_library-forge-2.7.2+1.20.1.jar"
SPELL_ENGINE_JAR="$ROOT/rpg-series-port/spell-engine-forge-1.20.1/spell_engine-forge-1.10.4+1.20.1.jar"
SPELL_POWER_JAR="$FOUNDATION/spell_power-forge-1.6.0+1.20.1-certified.jar"
RANGED_JAR="$FOUNDATION/ranged_weapon_api-forge-2.3.4+1.20.1-certified.jar"
TINY_JAR="$(find "$ROOT/rpg-series-port/tiny-config-forge-1.20.1/generated/forge/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' ! -name '*dev-shadow*' | sort | head -n1)"
SPELL_ENGINE_COMMON_JAR="$(find "$ROOT/.spell-engine-build/common/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' | sort | head -n1)"
SPELL_POWER_COMMON_JAR="$(find "$ROOT/rpg-series-port/spell_power-forge-1.20.1/common/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' | sort | head -n1)"
RANGED_COMMON_JAR="$(find "$ROOT/rpg-series-port/ranged-weapon-api-forge-1.20.1/common/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' | sort | head -n1)"
TINY_COMMON_JAR="$(find "$ROOT/rpg-series-port/tiny-config-forge-1.20.1/generated/common/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' | sort | head -n1)"
for f in "$OUT" "$SPELL_ENGINE_JAR" "$SPELL_POWER_JAR" "$RANGED_JAR" "$TINY_JAR" "$SPELL_ENGINE_COMMON_JAR" "$SPELL_POWER_COMMON_JAR" "$RANGED_COMMON_JAR" "$TINY_COMMON_JAR"; do test -f "$f"; unzip -tq "$f" >/dev/null; done
[[ -d "$SOURCE_WORLD" ]] || { echo '[More RPG 2.7.2] Stage-1 source world missing for HUD proof' >&2; exit 1; }

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
RUN="$WORK/forge/run"
SAVE='MRPG-QA-HUD'
rm -rf "$RUN/logs" "$RUN/.mixin.out" "$RUN/saves/$SAVE"
mkdir -p "$RUN/config" "$RUN/saves"
cp -a "$SOURCE_WORLD" "$RUN/saves/$SAVE"
printf 'earlyWindowControl = false\n' > "$RUN/config/fml.toml"
LOG="$PORT/more-rpg-hud-transform-client.log"; : > "$LOG"
ACTIVE_PID=''
descendants() { local parent="$1" child; while IFS= read -r child; do [[ -z "$child" ]] && continue; descendants "$child"; printf '%s\n' "$child"; done < <(pgrep -P "$parent" 2>/dev/null || true); }
stop_tree() { local root="${1:-}"; [[ -n "$root" ]] || return 0; local -a children=(); mapfile -t children < <(descendants "$root"); ((${#children[@]})) && kill -TERM "${children[@]}" 2>/dev/null || true; kill -TERM "$root" 2>/dev/null || true; for _ in {1..20}; do kill -0 "$root" 2>/dev/null || break; sleep 0.1; done; mapfile -t children < <(descendants "$root"); ((${#children[@]})) && kill -KILL "${children[@]}" 2>/dev/null || true; kill -KILL "$root" 2>/dev/null || true; wait "$root" 2>/dev/null || true; }
cleanup() { [[ -z "${ACTIVE_PID:-}" ]] || stop_tree "$ACTIVE_PID"; }
trap cleanup EXIT INT TERM

echo '[More RPG 2.7.2] DRAW_HEARTS_MIXIN_TRANSFORM_PROOF_BEGIN mapped_client=true'
MIXIN_OPTS='-Dmixin.debug.export=true -Dmixin.debug.export.filter=net.minecraft.client.gui.** -Dmixin.debug.export.decompile=false -Dmixin.debug.verbose=true'
env LIBGL_ALWAYS_SOFTWARE=1 MESA_LOADER_DRIVER_OVERRIDE=llvmpipe ALSOFT_DRIVERS=null JAVA_TOOL_OPTIONS="$MIXIN_OPTS ${JAVA_TOOL_OPTIONS:-}" \
  xvfb-run -a -s '-screen 0 1280x720x24 +extension GLX +extension RENDER -noreset' \
  gradle --no-daemon -p "$WORK" :forge:runClient --args="--width 1280 --height 720 --quickPlaySingleplayer $SAVE" "${ARGS[@]}" </dev/null > "$LOG" 2>&1 &
ACTIVE_PID=$!; PID=$ACTIVE_PID; DEADLINE=$((SECONDS+240)); READY=0
FATAL='ModLoadingException|Failed to create mod instance|NoClassDefFoundError|ClassNotFoundException|MixinApplyError|InvalidMixinException|MixinTransformerError|Exception in thread "Render thread"|The game crashed|Missing or unsupported mandatory dependencies|Could not initialize GLFW|Failed to initialize graphics window|Timed out trying to setup the Game Window'
while ((SECONDS<DEADLINE)); do
  LATEST="$RUN/logs/latest.log"; FILES=("$LOG"); [[ -f "$LATEST" ]] && FILES+=("$LATEST")
  if grep -Eiq "$FATAL" "${FILES[@]}"; then tail -n 700 "${FILES[@]}" 2>/dev/null || true; exit 1; fi
  if [[ -f "$LATEST" ]] && grep -Fq 'Reloading ResourceManager' "$LATEST" && grep -Fq 'Backend library: LWJGL' "$LATEST" && grep -Fq '[More RPG QA] FATAL_POISON_APPLIED' "${FILES[@]}"; then READY=1; break; fi
  if ! kill -0 "$PID" 2>/dev/null; then wait "$PID" || true; ACTIVE_PID=''; tail -n 700 "${FILES[@]}" 2>/dev/null || true; exit 1; fi
  sleep 1
done
[[ "$READY" -eq 1 ]] || { tail -n 700 "$LOG" "$RUN/logs/latest.log" 2>/dev/null || true; exit 1; }
LATEST="$RUN/logs/latest.log"
if ! grep -Eiq 'DrawHeartsMixin.*more-rpg-classes\.mixins\.json|more-rpg-classes\.mixins\.json.*DrawHeartsMixin' "$LOG" "$LATEST"; then
  echo '[More RPG 2.7.2] DrawHeartsMixin verbose application record missing' >&2; tail -n 700 "$LOG" "$LATEST" 2>/dev/null || true; exit 1
fi
MIXIN_OUT="$RUN/.mixin.out"
[[ -d "$MIXIN_OUT" ]] || { echo '[More RPG 2.7.2] Mixin post-transform export directory missing' >&2; exit 1; }
mapfile -t TARGETS < <(grep -aRl 'net/more_rpg_classes/client/heart/HeartRegistry' "$MIXIN_OUT" --include='*.class' 2>/dev/null || true)
((${#TARGETS[@]} > 0)) || { echo '[More RPG 2.7.2] no transformed GUI target contains More RPG HeartRegistry bytecode' >&2; find "$MIXIN_OUT" -type f -name '*.class' -print | sort | head -n 100 || true; exit 1; }
printf '[More RPG 2.7.2] DRAW_HEARTS_MIXIN_TRANSFORMED_CLIENT_PASS targets=%s first=%s\n' "${#TARGETS[@]}" "${TARGETS[0]#$MIXIN_OUT/}"
stop_tree "$PID"; ACTIVE_PID=''
echo '[More RPG 2.7.2] MAPPED_CLIENT_FATAL_POISON_HUD_PROOF_PASS effect=more_rpg_classes:fatal_poison transformed_hud=true'
