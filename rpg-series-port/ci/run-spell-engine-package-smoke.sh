#!/usr/bin/env bash
set -euo pipefail
OUT_JAR="${1:?spell engine jar}"
SPELL_POWER_JAR="${2:?spell power jar}"
RANGED_JAR="${3:?ranged jar}"
WORK="${4:?generated spell engine root}"
PORT="${5:?spell engine port evidence dir}"

find_module_jar() {
  local group="$1"
  local artifact="$2"
  local version="$3"
  local root="$HOME/.gradle/caches/modules-2/files-2.1/$group/$artifact/$version"
  local jar
  jar="$(find "$root" -type f -name "${artifact}-*.jar" ! -name '*sources*' ! -name '*javadoc*' | head -n 1 || true)"
  [[ -n "$jar" && -f "$jar" ]] || { echo "Missing resolved module JAR: $group:$artifact:$version" >&2; exit 1; }
  printf '%s\n' "$jar"
}
prop() { sed -n "s/^${1}=//p" "$WORK/gradle.properties" | tail -n 1 | tr -d '\r'; }
CLOTH_VERSION="$(prop cloth_config_version)"
PLAYER_VERSION="$(prop player_anim_version)"
[[ -n "$CLOTH_VERSION" && -n "$PLAYER_VERSION" ]] || { echo 'Missing resolved dependency versions' >&2; exit 1; }
CLOTH_JAR="$(find_module_jar me.shedaniel.cloth cloth-config-forge "$CLOTH_VERSION")"
PLAYER_JAR="$(find_module_jar dev.kosmx.player-anim player-animation-lib-forge "$PLAYER_VERSION")"

FRESH="$WORK/.fresh-forge-server"
rm -rf "$FRESH"
mkdir -p "$FRESH/mods"
curl -fsSL "https://maven.minecraftforge.net/net/minecraftforge/forge/1.20.1-47.4.23/forge-1.20.1-47.4.23-installer.jar" -o "$FRESH/forge-installer.jar"
(
  cd "$FRESH"
  java -jar forge-installer.jar --installServer >/dev/null
  cp "$OUT_JAR" "$SPELL_POWER_JAR" "$RANGED_JAR" "$CLOTH_JAR" "$PLAYER_JAR" mods/
  printf 'eula=true\n' > eula.txt
  printf '%s\n' '-Xmx2G' > user_jvm_args.txt
  cat > server.properties <<'PROPS'
level-type=minecraft:flat
generate-structures=false
view-distance=3
simulation-distance=3
spawn-protection=0
PROPS
)
LOG="$PORT/forge-package-smoke.log"
: > "$LOG"
(
  cd "$FRESH"
  exec ./run.sh nogui
) > "$LOG" 2>&1 &
PID=$!
stop_tree() {
  local root="$1" child
  local kids="$(pgrep -P "$root" 2>/dev/null || true)"
  for child in $kids; do kill -TERM "$child" 2>/dev/null || true; done
  kill -TERM "$root" 2>/dev/null || true
  sleep 1
  kids="$(pgrep -P "$root" 2>/dev/null || true)"
  for child in $kids; do kill -KILL "$child" 2>/dev/null || true; done
  kill -KILL "$root" 2>/dev/null || true
  wait "$root" 2>/dev/null || true
}
trap 'stop_tree "$PID" 2>/dev/null || true' EXIT INT TERM
DEADLINE=$((SECONDS+180)); PASS=0
FATAL='MixinApplyError|InvalidMixinException|MixinTransformerError|Failed to start the minecraft server|Failed to create mod instance|NoClassDefFoundError|ClassNotFoundException|Registry is already frozen|Can not register to a locked registry|Missing or unsupported mandatory dependencies|Exception in server tick loop'
while ((SECONDS<DEADLINE)); do
  LATEST="$FRESH/logs/latest.log"
  FILES=("$LOG"); [[ -f "$LATEST" ]] && FILES+=("$LATEST")
  if grep -Eiq "$FATAL" "${FILES[@]}"; then tail -n 500 "${FILES[@]}" || true; exit 1; fi
  if [[ -f "$LATEST" ]] && grep -Eq 'Done \([0-9.]+s\)!' "$LATEST"; then PASS=1; break; fi
  if ! kill -0 "$PID" 2>/dev/null; then wait "$PID" || true; tail -n 500 "${FILES[@]}" || true; exit 1; fi
  sleep 1
done
[[ "$PASS" -eq 1 ]] || { tail -n 500 "$LOG" || true; exit 1; }
stop_tree "$PID"; trap - EXIT INT TERM
printf '[Spell Engine CI] Fresh packaged Forge server bootstrap passed.\n'
