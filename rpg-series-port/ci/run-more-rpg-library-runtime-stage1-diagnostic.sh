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
CLIENT_LOG="$ROOT/rpg-series-port/more-rpg-library-forge-1.20.1/more-rpg-native-client-integrated.log"
mkdir -p "$CLIENT_RUN"
if [[ -f "$CLIENT_OPTIONS" ]] && grep -q '^onboardAccessibility:' "$CLIENT_OPTIONS"; then
  sed -i 's/^onboardAccessibility:.*/onboardAccessibility:false/' "$CLIENT_OPTIONS"
else
  printf 'onboardAccessibility:false\n' >> "$CLIENT_OPTIONS"
fi
[[ "$(grep -Ec '^onboardAccessibility:false$' "$CLIENT_OPTIONS")" -eq 1 ]]
echo '[More RPG 2.7.2] CLIENT_FIRST_RUN_ACCESSIBILITY_GATE_DISABLED_FOR_QA onboardAccessibility=false'

# Prepare the direct HUD transformation proof without weakening the existing world-entry gate.
# A successful fatal-poison marker proves a real integrated player; these Mixin debug settings
# additionally preserve the post-transform GUI bytecode so DrawHeartsMixin itself can be verified.
rm -rf "$CLIENT_RUN/.mixin.out"
MIXIN_JAVA_TOOL_OPTIONS='-Dmixin.debug.export=true -Dmixin.debug.export.filter=net.minecraft.client.gui.** -Dmixin.debug.export.decompile=false -Dmixin.debug.verbose=true'
export JAVA_TOOL_OPTIONS="$MIXIN_JAVA_TOOL_OPTIONS${JAVA_TOOL_OPTIONS:+ $JAVA_TOOL_OPTIONS}"
echo '[More RPG 2.7.2] DRAW_HEARTS_MIXIN_EXPORT_ARMED filter=net.minecraft.client.gui.**'

prove_draw_hearts_transform() {
  local latest="$CLIENT_RUN/logs/latest.log"
  local mixin_out="$CLIENT_RUN/.mixin.out"
  local -a heart_targets=()

  [[ -f "$CLIENT_LOG" && -f "$latest" ]] || {
    echo '[More RPG 2.7.2] direct HUD proof missing client runtime logs' >&2
    return 1
  }
  if ! grep -Eiq 'DrawHeartsMixin.*more-rpg-classes\.mixins\.json|more-rpg-classes\.mixins\.json.*DrawHeartsMixin|DrawHeartsMixin' "$CLIENT_LOG" "$latest"; then
    echo '[More RPG 2.7.2] DrawHeartsMixin verbose application record missing' >&2
    return 1
  fi
  [[ -d "$mixin_out" ]] || {
    echo '[More RPG 2.7.2] Mixin post-transform export directory missing' >&2
    return 1
  }
  mapfile -t heart_targets < <(grep -aRl 'net/more_rpg_classes/client/heart/HeartRegistry' "$mixin_out" --include='*.class' 2>/dev/null | sort || true)
  if ((${#heart_targets[@]} == 0)); then
    echo '[More RPG 2.7.2] No exported transformed GUI class contains HeartRegistry reference' >&2
    find "$mixin_out" -type f -name '*.class' -print | sort | head -n 100 || true
    return 1
  fi
  printf '[More RPG 2.7.2] DRAW_HEARTS_MIXIN_TRANSFORMED_CLIENT_PASS targets=%s first=%s\n' \
    "${#heart_targets[@]}" "${heart_targets[0]#$mixin_out/}"
}

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
    [[ "$rc" -eq 0 ]] || exit "$rc"
    prove_draw_hearts_transform
    echo '[More RPG 2.7.2] RUNTIME_STAGE1_HARDENED_PASS world_entry=true transformed_hud=true production_client_pending=true'
    exit 0
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
[[ "$rc" -eq 0 ]] || exit "$rc"
prove_draw_hearts_transform
echo '[More RPG 2.7.2] RUNTIME_STAGE1_HARDENED_PASS world_entry=true transformed_hud=true production_client_pending=true'
exit 0
