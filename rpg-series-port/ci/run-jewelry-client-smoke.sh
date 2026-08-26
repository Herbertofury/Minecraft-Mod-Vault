#!/usr/bin/env bash
set -euo pipefail

ROOT="$(pwd)"
PORT="$ROOT/rpg-series-port/jewelry-forge-1.20.1/generated"
EVIDENCE="$ROOT/rpg-series-port/jewelry-forge-1.20.1"

# Exercise the Forge client lifecycle, client-only event subscribers, resource reload and LWJGL
# bootstrap with the same generated source/dependency graph. This is explicitly a dev-client gate;
# run-jewelry.sh remains the packaged release-JAR proof on fresh Forge servers.
rm -rf "$PORT/forge/run/logs"
mkdir -p "$PORT/forge/run/config"
printf 'earlyWindowControl = false\n' > "$PORT/forge/run/config/fml.toml"
CLIENT_LOG="$EVIDENCE/forge-client-smoke.log"
: > "$CLIENT_LOG"

stop_tree() {
    local root="$1" child kids
    kids="$(pgrep -P "$root" 2>/dev/null || true)"
    for child in $kids; do kill -TERM "$child" 2>/dev/null || true; done
    kill -TERM "$root" 2>/dev/null || true
    sleep 1
    kids="$(pgrep -P "$root" 2>/dev/null || true)"
    for child in $kids; do kill -KILL "$child" 2>/dev/null || true; done
    kill -KILL "$root" 2>/dev/null || true
    wait "$root" 2>/dev/null || true
}

dump_logs() {
    local file
    for file in "$@"; do
        [[ -f "$file" ]] || continue
        echo "===== tail: $file ====="
        tail -n 400 "$file" || true
    done
}

env LIBGL_ALWAYS_SOFTWARE=1 MESA_LOADER_DRIVER_OVERRIDE=llvmpipe \
    xvfb-run -a -s '-screen 0 1280x720x24 +extension GLX +extension RENDER -noreset' \
    gradle --no-daemon -p "$PORT" :forge:runClient </dev/null > "$CLIENT_LOG" 2>&1 &
PID=$!
DEADLINE=$((SECONDS+180))
READY=0
PASS=0
FATAL='MixinApplyError|InvalidMixinException|MixinTransformerError|Failed to create mod instance|NoClassDefFoundError|ClassNotFoundException|The game crashed whilst initializing game|Exception in thread "Render thread"|Failed to initialize graphics window|Timed out trying to setup the Game Window|Could not initialize GLFW|Missing or unsupported mandatory dependencies|Registry is already frozen|Can not register to a locked registry'

while ((SECONDS<DEADLINE)); do
    LATEST="$PORT/forge/run/logs/latest.log"
    FILES=("$CLIENT_LOG"); [[ -f "$LATEST" ]] && FILES+=("$LATEST")
    if grep -Eiq "$FATAL" "${FILES[@]}"; then
        stop_tree "$PID"
        dump_logs "${FILES[@]}"
        exit 1
    fi
    if [[ -f "$LATEST" ]] \
        && grep -Fq 'Reloading ResourceManager' "$LATEST" \
        && grep -Fq 'Backend library: LWJGL' "$LATEST"; then
        [[ "$READY" -eq 0 ]] && READY=$SECONDS
        if ((SECONDS-READY>=8)); then PASS=1; break; fi
    fi
    if ! kill -0 "$PID" 2>/dev/null; then
        wait "$PID" || true
        dump_logs "${FILES[@]}"
        exit 1
    fi
    sleep 1
done

if [[ "$PASS" -ne 1 ]]; then
    stop_tree "$PID"
    dump_logs "$CLIENT_LOG" "$PORT/forge/run/logs/latest.log"
    exit 1
fi
stop_tree "$PID"
echo "[Jewelry CI] Forge dev-client bootstrap passed with current RPG dependencies + Curios."
