#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"; RUN="$ROOT/rpg-series-port/rogues-forge-1.20.1/forge/run"; OPTIONS="$RUN/options.txt"
mkdir -p "$RUN"; touch "$OPTIONS"
ensure_option(){ local key="$1" value="$2"; if grep -q "^${key}:" "$OPTIONS"; then sed -i "s/^${key}:.*/${key}:${value}/" "$OPTIONS"; else printf '%s:%s\n' "$key" "$value" >> "$OPTIONS"; fi; }
ensure_option onboardAccessibility false; ensure_option skipMultiplayerWarning true; ensure_option joinedFirstServer true; ensure_option pauseOnLostFocus false
for expected in onboardAccessibility:false skipMultiplayerWarning:true joinedFirstServer:true pauseOnLostFocus:false; do grep -Fxq "$expected" "$OPTIONS"; done
echo '[Rogues player QA] FIRST_RUN_CLIENT_PREFLIGHT_PASS: accessibility/multiplayer first-run blockers disabled.'
exec bash "$ROOT/rpg-series-port/ci/run-rogues-player-behavior-acceptance.sh"
