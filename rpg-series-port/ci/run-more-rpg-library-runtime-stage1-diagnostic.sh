#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
BASE="$ROOT/rpg-series-port/ci/run-more-rpg-library-runtime-stage1.sh"
test -f "$BASE"
bash -n "$BASE"

# Minecraft 1.20.1 puts a first-launch accessibility onboarding screen in front of
# Quick Play when options.txt is absent. CI uses a fresh disposable run directory,
# so seed only that onboarding flag before the base Stage-1 harness launches the
# mapped client. The gameplay marker in the copied MRPG-QA world remains the
# authoritative proof that Quick Play actually reached the integrated player.
CLIENT_RUN="$ROOT/.more-rpg-library-build/forge/run"
CLIENT_OPTIONS="$CLIENT_RUN/options.txt"
mkdir -p "$CLIENT_RUN"
if [[ -f "$CLIENT_OPTIONS" ]] && grep -q '^onboardAccessibility:' "$CLIENT_OPTIONS"; then
  sed -i 's/^onboardAccessibility:.*/onboardAccessibility:false/' "$CLIENT_OPTIONS"
else
  printf 'onboardAccessibility:false\n' >> "$CLIENT_OPTIONS"
fi
[[ "$(grep -Ec '^onboardAccessibility:false$' "$CLIENT_OPTIONS")" -eq 1 ]]
echo '[More RPG 2.7.2] CLIENT_FIRST_RUN_ACCESSIBILITY_GATE_DISABLED_FOR_QA onboardAccessibility=false'

echo '[More RPG 2.7.2] RUNTIME_STAGE1_DIAGNOSTIC_WRAPPER_BEGIN source=run-359'
bash "$BASE" &
BASE_PID=$!

# #351 stayed alive without progressing to server readiness. If the same condition lasts two minutes,
# ask the Forge JVM for its standard SIGQUIT thread dump. The JVM writes that dump to the already
# preserved Stage-1 stdout/stderr stream; no extra process state or environment data is collected.
for _ in $(seq 1 120); do
  if ! kill -0 "$BASE_PID" 2>/dev/null; then
    set +e
    wait "$BASE_PID"
    rc=$?
    set -e
    echo "[More RPG 2.7.2] RUNTIME_STAGE1_DIAGNOSTIC_WRAPPER_EARLY_EXIT rc=$rc"
    exit "$rc"
  fi
  sleep 1
done

SERVER_PID="$(pgrep -f 'cpw\.mods\.bootstraplauncher\.BootstrapLauncher.*forgeserver' | head -n1 || true)"
if [[ -n "$SERVER_PID" ]]; then
  echo "[More RPG 2.7.2] PACKAGED_SERVER_STALL_SIGQUIT_BEGIN pid=$SERVER_PID"
  kill -QUIT "$SERVER_PID" || true
  sleep 3
  echo '[More RPG 2.7.2] PACKAGED_SERVER_STALL_SIGQUIT_END'
else
  echo '[More RPG 2.7.2] PACKAGED_SERVER_STALL_SIGQUIT_MISSING no_forgeserver_pid'
fi

set +e
wait "$BASE_PID"
rc=$?
set -e
echo "[More RPG 2.7.2] RUNTIME_STAGE1_DIAGNOSTIC_WRAPPER_END rc=$rc"
exit "$rc"
