#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
BASE="$ROOT/rpg-series-port/ci/run-more-rpg-library-runtime-stage1.sh"
test -f "$BASE"
bash -n "$BASE"

echo '[More RPG 2.7.2] RUNTIME_STAGE1_DIAGNOSTIC_WRAPPER_BEGIN source=run-351'
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
