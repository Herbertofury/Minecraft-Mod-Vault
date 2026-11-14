#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
RUN="$ROOT/rpg-series-port/paladins-forge-1.20.1/forge/run"
OPTIONS="$RUN/options.txt"
PLAYER_QA="$ROOT/rpg-series-port/ci/run-paladins-player-behavior-acceptance.sh"
PATCHED_QA="${RUNNER_TEMP:-/tmp}/run-paladins-player-behavior-acceptance.single-target.sh"

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

# Minecraft 1.20.x puts the first-run accessibility onboarding screen in front of
# Quick Play when options.txt is absent/uninitialized. Under headless CI that leaves
# a healthy LWJGL client parked forever before --quickPlayMultiplayer executes.
ensure_option onboardAccessibility false
ensure_option skipMultiplayerWarning true
ensure_option joinedFirstServer true
ensure_option pauseOnLostFocus false

grep -Fxq 'onboardAccessibility:false' "$OPTIONS"
grep -Fxq 'skipMultiplayerWarning:true' "$OPTIONS"
grep -Fxq 'joinedFirstServer:true' "$OPTIONS"
grep -Fxq 'pauseOnLostFocus:false' "$OPTIONS"
echo '[Paladins player QA] FIRST_RUN_QUICKPLAY_PREFLIGHT_PASS: accessibility onboarding and multiplayer first-run blockers disabled in Forge dev-client options.'

# /damage requires an entity argument whose selector is statically single-target.
# Run #210 proved the real client/server path, but the three historical `@a` damage
# commands were rejected by Brigadier before Divine Protection could be exercised.
# Keep the product/release bytes immutable and repair only a disposable QA copy.
cp "$PLAYER_QA" "$PATCHED_QA"
BAD_DAMAGE_COUNT="$(grep -Fc "send_cmd 'damage @a 5 minecraft:generic'" "$PATCHED_QA" || true)"
[[ "$BAD_DAMAGE_COUNT" = 3 ]] || {
  echo "[Paladins player QA] expected exactly three legacy multi-target damage commands, found $BAD_DAMAGE_COUNT" >&2
  exit 1
}
sed -i "s/send_cmd 'damage @a 5 minecraft:generic'/send_cmd 'damage @a[limit=1] 5 minecraft:generic'/g" "$PATCHED_QA"
[[ "$(grep -Fc "send_cmd 'damage @a 5 minecraft:generic'" "$PATCHED_QA" || true)" = 0 ]]
[[ "$(grep -Fc "send_cmd 'damage @a[limit=1] 5 minecraft:generic'" "$PATCHED_QA" || true)" = 3 ]]
echo '[Paladins player QA] SINGLE_TARGET_DAMAGE_HARNESS_PASS: all Divine Protection damage probes use a statically single-target selector.'

exec bash "$PATCHED_QA"
