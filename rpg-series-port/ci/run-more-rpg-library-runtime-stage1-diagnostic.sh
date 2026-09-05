#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
BASE="$ROOT/rpg-series-port/ci/run-more-rpg-library-runtime-stage1.sh"
INSPECTOR_SRC="$ROOT/rpg-series-port/ci/mapped-client-inspector"
INSPECTOR_WORK="$ROOT/.more-rpg-client-inspector"
INSPECTOR_CLASSES="$INSPECTOR_WORK/classes"
AGENT_JAR="$INSPECTOR_WORK/more-rpg-mapped-client-state-agent.jar"
AGENT_MANIFEST="$INSPECTOR_WORK/AGENT.MF"
STATE_LOG="$ROOT/rpg-series-port/more-rpg-library-forge-1.20.1/more-rpg-mapped-client-state.log"
CLIENT_RUN="$ROOT/.more-rpg-library-build/forge/run"
CLIENT_OPTIONS="$CLIENT_RUN/options.txt"
CLIENT_LOG="$ROOT/rpg-series-port/more-rpg-library-forge-1.20.1/more-rpg-native-client-integrated.log"

test -f "$BASE"
bash -n "$BASE"
test -f "$INSPECTOR_SRC/mrpg/qa/StateAgent.java"
test -f "$INSPECTOR_SRC/mrpg/qa/AttachMain.java"

# Compile the CI-only JDK Attach probe before Minecraft starts. These classes never enter a mod JAR.
rm -rf "$INSPECTOR_WORK"
mkdir -p "$INSPECTOR_CLASSES"
javac --release 17 --add-modules jdk.attach \
  -d "$INSPECTOR_CLASSES" \
  "$INSPECTOR_SRC/mrpg/qa/StateAgent.java" \
  "$INSPECTOR_SRC/mrpg/qa/AttachMain.java"
cat > "$AGENT_MANIFEST" <<'MANIFEST'
Manifest-Version: 1.0
Agent-Class: mrpg.qa.StateAgent
Can-Redefine-Classes: false
Can-Retransform-Classes: false
MANIFEST
jar cfm "$AGENT_JAR" "$AGENT_MANIFEST" -C "$INSPECTOR_CLASSES" mrpg/qa/StateAgent.class
jar tf "$AGENT_JAR" | grep -Fx 'mrpg/qa/StateAgent.class' >/dev/null
: > "$STATE_LOG"
echo '[More RPG 2.7.2] MAPPED_CLIENT_STATE_PROBE_READY mode=jdk-attach ci_only=true'

# Minecraft 1.20.1 puts its accessibility onboarding screen in front of normal startup when this
# option is true. The fresh disposable run directory must explicitly mark onboarding complete.
mkdir -p "$CLIENT_RUN"
if [[ -f "$CLIENT_OPTIONS" ]] && grep -q '^onboardAccessibility:' "$CLIENT_OPTIONS"; then
  sed -i 's/^onboardAccessibility:.*/onboardAccessibility:false/' "$CLIENT_OPTIONS"
else
  printf 'onboardAccessibility:false\n' >> "$CLIENT_OPTIONS"
fi
[[ "$(grep -Ec '^onboardAccessibility:false$' "$CLIENT_OPTIONS")" -eq 1 ]]
echo '[More RPG 2.7.2] CLIENT_FIRST_RUN_ACCESSIBILITY_GATE_DISABLED_FOR_QA onboardAccessibility=false'

echo '[More RPG 2.7.2] RUNTIME_STAGE1_DIAGNOSTIC_WRAPPER_BEGIN source=run-362'
bash "$BASE" &
BASE_PID=$!

# Wait only until the mapped client has finished core resource initialization, then capture one
# authoritative in-process state snapshot. Do not modify screens or world state from the probe.
PROBE_ATTEMPTED=0
for _ in $(seq 1 210); do
  if ! kill -0 "$BASE_PID" 2>/dev/null; then
    break
  fi
  if [[ -f "$CLIENT_LOG" ]] && grep -Fq 'Reloading ResourceManager' "$CLIENT_LOG"; then
    # The reload-begin marker occurs before atlas/model finalization. Give Quick Play a bounded
    # post-reload window so the snapshot reflects the stable blocker rather than LoadingOverlay.
    for _settle in $(seq 1 15); do
      kill -0 "$BASE_PID" 2>/dev/null || break
      sleep 1
    done
    kill -0 "$BASE_PID" 2>/dev/null || break
    CLIENT_PID="$(pgrep -f 'forgeclientuserdev.*--quickPlaySingleplayer.*MRPG-QA' | head -n1 || true)"
    if [[ -n "$CLIENT_PID" ]]; then
      PROBE_ATTEMPTED=1
      echo "[More RPG 2.7.2] MAPPED_CLIENT_STATE_PROBE_ATTACH_BEGIN pid=$CLIENT_PID"
      jcmd "$CLIENT_PID" VM.command_line 2>/dev/null | sed 's/^/[More RPG QA jcmd] /' || true
      set +e
      java --add-modules jdk.attach -cp "$INSPECTOR_CLASSES" \
        mrpg.qa.AttachMain "$CLIENT_PID" "$AGENT_JAR" "$STATE_LOG"
      probe_rc=$?
      set -e
      if [[ "$probe_rc" -ne 0 ]]; then
        echo "[More RPG 2.7.2] MAPPED_CLIENT_STATE_PROBE_ATTACH_FAILED rc=$probe_rc"
      elif [[ -s "$STATE_LOG" ]]; then
        cat "$STATE_LOG"
        echo '[More RPG 2.7.2] MAPPED_CLIENT_STATE_PROBE_CAPTURED'
        if grep -Fq 'world=<null>' "$STATE_LOG" \
          && grep -Fq 'player=<null>' "$STATE_LOG" \
          && grep -Fq 'integratedServer=false' "$STATE_LOG"; then
          echo '[More RPG 2.7.2] STABLE_MENU_BLOCKER_CAPTURED stopping_stalled_client=true'
          kill -TERM "$CLIENT_PID" 2>/dev/null || true
        fi
      else
        echo '[More RPG 2.7.2] MAPPED_CLIENT_STATE_PROBE_EMPTY'
      fi
      break
    fi
  fi
  sleep 1
done

if [[ "$PROBE_ATTEMPTED" -eq 0 ]]; then
  echo '[More RPG 2.7.2] MAPPED_CLIENT_STATE_PROBE_NOT_REACHED client_resource_marker_or_pid_missing'
fi

set +e
wait "$BASE_PID"
rc=$?
set -e
if [[ -s "$STATE_LOG" ]]; then
  echo '[More RPG 2.7.2] MAPPED_CLIENT_STATE_PROBE_FINAL'
  cat "$STATE_LOG"
fi
echo "[More RPG 2.7.2] RUNTIME_STAGE1_DIAGNOSTIC_WRAPPER_END rc=$rc"
exit "$rc"
