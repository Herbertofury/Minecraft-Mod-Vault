#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
RUN="$ROOT/rpg-series-port/paladins-forge-1.20.1/forge/run"
OPTIONS="$RUN/options.txt"

mkdir -p "$RUN"
touch "$OPTIONS"
ensure_option(){
  local key="$1" value="$2"
  if grep -q "^${key}:" "$OPTIONS"; then
    sed -i "s/^${key}:.*/${key}:${value}/" "$OPTIONS"
  else
    printf '%s:%s\n' "$key" "$value" >> "$OPTIONS"
  fi
}

# Keep vanilla first-run blockers disabled even though the deterministic QA wrapper
# now dispatches Minecraft's own ConnectScreen directly after client initialization.
ensure_option onboardAccessibility false
ensure_option skipMultiplayerWarning true
ensure_option joinedFirstServer true
ensure_option pauseOnLostFocus false

grep -Fxq 'onboardAccessibility:false' "$OPTIONS"
grep -Fxq 'skipMultiplayerWarning:true' "$OPTIONS"
grep -Fxq 'joinedFirstServer:true' "$OPTIONS"
grep -Fxq 'pauseOnLostFocus:false' "$OPTIONS"
echo '[Paladins player QA] FIRST_RUN_QUICKPLAY_PREFLIGHT_PASS: accessibility onboarding and multiplayer first-run blockers disabled in Forge dev-client options.'

exec bash "$ROOT/rpg-series-port/ci/run-paladins-player-autoconnect-wrapper.sh"
