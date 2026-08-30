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
# preserved Stage-1 stdout/stderr stream; no extra environment or user data is collected.
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

# Forge's generated run.sh passes most launch arguments through @.../unix_args.txt, so /proc command
# lines do not reliably contain BootstrapLauncher/forgeserver. jps resolves the live Java main class;
# the argfile pattern is only a fallback for environments where jps is unavailable.
SERVER_PID="$(jps -l 2>/dev/null | awk '/cpw\.mods\.bootstraplauncher\.BootstrapLauncher/ {print $1; exit}')"
if [[ -z "$SERVER_PID" ]]; then
  SERVER_PID="$(pgrep -f 'java.*forge-1\.20\.1-47\.4\.23.*unix_args\.txt.*nogui' | head -n1 || true)"
fi
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
