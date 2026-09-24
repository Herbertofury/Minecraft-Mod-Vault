#!/usr/bin/env bash
set -euo pipefail

OUT_JAR="${1:?Armor Model API release jar}"
WORK="${2:?generated Armor Model API root}"
PORT="${3:?Armor Model API evidence dir}"

[[ -f "$OUT_JAR" ]] || { echo "Missing release JAR: $OUT_JAR" >&2; exit 1; }

SERVER="$WORK/.fresh-armor-model-api-forge-server"
rm -rf "$SERVER"
mkdir -p "$SERVER"

curl -fsSL "https://maven.minecraftforge.net/net/minecraftforge/forge/1.20.1-47.4.23/forge-1.20.1-47.4.23-installer.jar" -o "$SERVER/forge-installer.jar"
(
  cd "$SERVER"
  java -jar forge-installer.jar --installServer >/dev/null
  printf 'eula=true\n' > eula.txt
  cat > user_jvm_args.txt <<'JVM'
-Xmx2G
JVM
  cat > server.properties <<'PROPS'
level-type=minecraft:flat
generate-structures=false
view-distance=3
simulation-distance=3
spawn-protection=0
online-mode=false
PROPS
  mkdir -p mods
  cp "$OUT_JAR" mods/
)

INSTALLED_JAR="$SERVER/mods/$(basename "$OUT_JAR")"
[[ -f "$INSTALLED_JAR" ]]
[[ "$(find "$SERVER/mods" -maxdepth 1 -type f -name '*.jar' | wc -l | tr -d ' ')" = 1 ]]
sha256sum "$INSTALLED_JAR" | tee "$PORT/armor-model-api-package-installed.sha256"
cmp -s "$OUT_JAR" "$INSTALLED_JAR"
unzip -t "$INSTALLED_JAR" >/dev/null
unzip -p "$INSTALLED_JAR" META-INF/mods.toml | grep -F 'modId="armor_model_api"' >/dev/null
unzip -p "$INSTALLED_JAR" META-INF/MANIFEST.MF | tr -d '\r' | grep -F 'MixinConfigs: armor_model_api.mixins.json' >/dev/null

stop_tree() {
  local root="$1" child
  local kids
  kids="$(pgrep -P "$root" 2>/dev/null || true)"
  for child in $kids; do kill -TERM "$child" 2>/dev/null || true; done
  kill -TERM "$root" 2>/dev/null || true
  sleep 1
  kids="$(pgrep -P "$root" 2>/dev/null || true)"
  for child in $kids; do kill -KILL "$child" 2>/dev/null || true; done
  kill -KILL "$root" 2>/dev/null || true
  wait "$root" 2>/dev/null || true
}

LOG="$PORT/armor-model-api-package-server-smoke.log"
: > "$LOG"
(
  cd "$SERVER"
  exec ./run.sh nogui
) > "$LOG" 2>&1 &
PID=$!
DEADLINE=$((SECONDS+180))
PASS=0
FATAL='ModLoadingException|Failed to create mod instance|Failed to start the minecraft server|NoClassDefFoundError|ClassNotFoundException|MixinApplyError|InvalidMixinException|MixinTransformerError|Registry is already frozen|Can not register to a locked registry|Missing or unsupported mandatory dependencies|Exception in server tick loop|The game crashed'

while ((SECONDS<DEADLINE)); do
  LATEST="$SERVER/logs/latest.log"
  FILES=("$LOG")
  [[ -f "$LATEST" ]] && FILES+=("$LATEST")
  if grep -Eiq "$FATAL" "${FILES[@]}"; then
    tail -n 500 "${FILES[@]}" || true
    stop_tree "$PID"
    exit 1
  fi
  if [[ -f "$LATEST" ]] && grep -Eq 'Done \([0-9.]+s\)!' "$LATEST"; then
    PASS=1
    break
  fi
  if ! kill -0 "$PID" 2>/dev/null; then
    wait "$PID" || true
    tail -n 500 "${FILES[@]}" || true
    echo '[Armor Model API] fresh packaged server exited before ready state' >&2
    exit 1
  fi
  sleep 1
done

if [[ "$PASS" -ne 1 ]]; then
  tail -n 500 "$LOG" || true
  [[ -f "$SERVER/logs/latest.log" ]] && tail -n 500 "$SERVER/logs/latest.log" || true
  stop_tree "$PID"
  echo '[Armor Model API] fresh packaged server timed out before ready state' >&2
  exit 1
fi

stop_tree "$PID"
printf '[Armor Model API] Fresh packaged Forge 47.4.23 server reached ready state with only the release JAR installed.\n'
