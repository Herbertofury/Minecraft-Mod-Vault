#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
BASE="$ROOT/rpg-series-port/ci/run-more-rpg-library-runtime-stage1.sh"
INSPECTOR_SRC="$ROOT/rpg-series-port/ci/mapped-client-inspector"
INSPECTOR_WORK="$ROOT/.more-rpg-client-inspector"
INSPECTOR_CLASSES="$INSPECTOR_WORK/classes"
STATE_AGENT_JAR="$INSPECTOR_WORK/more-rpg-mapped-client-state-agent.jar"
DRIVER_AGENT_JAR="$INSPECTOR_WORK/more-rpg-quickplay-driver-agent.jar"
STATE_MANIFEST="$INSPECTOR_WORK/STATE-AGENT.MF"
DRIVER_MANIFEST="$INSPECTOR_WORK/DRIVER-AGENT.MF"
STATE_LOG="$ROOT/rpg-series-port/more-rpg-library-forge-1.20.1/more-rpg-mapped-client-state.log"
DRIVER_LOG="$ROOT/rpg-series-port/more-rpg-library-forge-1.20.1/more-rpg-quickplay-redrive.log"
CLIENT_RUN="$ROOT/.more-rpg-library-build/forge/run"
CLIENT_OPTIONS="$CLIENT_RUN/options.txt"
CLIENT_LOG="$ROOT/rpg-series-port/more-rpg-library-forge-1.20.1/more-rpg-native-client-integrated.log"

test -f "$BASE"
bash -n "$BASE"
for f in StateAgent.java AttachMain.java QuickPlayDriverAgent.java; do
  test -f "$INSPECTOR_SRC/mrpg/qa/$f"
done

# Compile CI-only attach probes before Minecraft starts. None of these classes enter a mod JAR.
rm -rf "$INSPECTOR_WORK"
mkdir -p "$INSPECTOR_CLASSES"
javac --release 17 --add-modules jdk.attach \
  -d "$INSPECTOR_CLASSES" \
  "$INSPECTOR_SRC/mrpg/qa/StateAgent.java" \
  "$INSPECTOR_SRC/mrpg/qa/AttachMain.java" \
  "$INSPECTOR_SRC/mrpg/qa/QuickPlayDriverAgent.java"
cat > "$STATE_MANIFEST" <<'MANIFEST'
Manifest-Version: 1.0
Agent-Class: mrpg.qa.StateAgent
Can-Redefine-Classes: false
Can-Retransform-Classes: false
MANIFEST
cat > "$DRIVER_MANIFEST" <<'MANIFEST'
Manifest-Version: 1.0
Agent-Class: mrpg.qa.QuickPlayDriverAgent
Can-Redefine-Classes: false
Can-Retransform-Classes: false
MANIFEST
jar cfm "$STATE_AGENT_JAR" "$STATE_MANIFEST" -C "$INSPECTOR_CLASSES" mrpg/qa/StateAgent.class
jar cfm "$DRIVER_AGENT_JAR" "$DRIVER_MANIFEST" -C "$INSPECTOR_CLASSES" mrpg/qa/QuickPlayDriverAgent.class
jar tf "$STATE_AGENT_JAR" | grep -Fx 'mrpg/qa/StateAgent.class' >/dev/null
jar tf "$DRIVER_AGENT_JAR" | grep -Fx 'mrpg/qa/QuickPlayDriverAgent.class' >/dev/null
: > "$STATE_LOG"
: > "$DRIVER_LOG"
echo '[More RPG 2.7.2] MAPPED_CLIENT_STATE_AND_QUICKPLAY_PROBES_READY mode=jdk-attach ci_only=true'

# A fresh disposable 1.20.1 run directory must explicitly mark accessibility onboarding complete.
mkdir -p "$CLIENT_RUN"
if [[ -f "$CLIENT_OPTIONS" ]] && grep -q '^onboardAccessibility:' "$CLIENT_OPTIONS"; then
  sed -i 's/^onboardAccessibility:.*/onboardAccessibility:false/' "$CLIENT_OPTIONS"
else
  printf 'onboardAccessibility:false\n' >> "$CLIENT_OPTIONS"
fi
[[ "$(grep -Ec '^onboardAccessibility:false$' "$CLIENT_OPTIONS")" -eq 1 ]]
echo '[More RPG 2.7.2] CLIENT_FIRST_RUN_ACCESSIBILITY_GATE_DISABLED_FOR_QA onboardAccessibility=false'

echo '[More RPG 2.7.2] RUNTIME_STAGE1_DIAGNOSTIC_WRAPPER_BEGIN source=run-363'
bash "$BASE" &
BASE_PID=$!

# Capture one stable post-reload state snapshot. Only the exact proven title-screen stall is eligible
# for a one-shot vanilla QuickPlay.startSingleplayer redrive. Other screens remain diagnostic failures.
PROBE_ATTEMPTED=0
for _ in $(seq 1 210); do
  kill -0 "$BASE_PID" 2>/dev/null || break
  if [[ -f "$CLIENT_LOG" ]] && grep -Fq 'Reloading ResourceManager' "$CLIENT_LOG"; then
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
        mrpg.qa.AttachMain "$CLIENT_PID" "$STATE_AGENT_JAR" "$STATE_LOG"
      probe_rc=$?
      set -e
      if [[ "$probe_rc" -ne 0 ]]; then
        echo "[More RPG 2.7.2] MAPPED_CLIENT_STATE_PROBE_ATTACH_FAILED rc=$probe_rc"
      elif [[ -s "$STATE_LOG" ]]; then
        cat "$STATE_LOG"
        echo '[More RPG 2.7.2] MAPPED_CLIENT_STATE_PROBE_CAPTURED'
        if grep -Fq 'screen=net.minecraft.client.gui.screen.TitleScreen' "$STATE_LOG" \
          && grep -Fq 'world=<null>' "$STATE_LOG" \
          && grep -Fq 'player=<null>' "$STATE_LOG" \
          && grep -Fq 'integratedServer=false' "$STATE_LOG"; then
          echo '[More RPG 2.7.2] TITLE_SCREEN_QUICKPLAY_STALL_CONFIRMED redrive=vanilla_startSingleplayer level=MRPG-QA'
          set +e
          java --add-modules jdk.attach -cp "$INSPECTOR_CLASSES" \
            mrpg.qa.AttachMain "$CLIENT_PID" "$DRIVER_AGENT_JAR" "$DRIVER_LOG|MRPG-QA"
          driver_rc=$?
          set -e
          if [[ "$driver_rc" -ne 0 ]]; then
            echo "[More RPG 2.7.2] VANILLA_QUICKPLAY_REDRIVE_ATTACH_FAILED rc=$driver_rc"
          else
            for _driver_wait in $(seq 1 10); do
              grep -Fq 'VANILLA_QUICKPLAY_SINGLEPLAYER_REDRIVE invoked=true level=MRPG-QA' "$DRIVER_LOG" && break
              sleep 1
            done
            cat "$DRIVER_LOG"
            grep -Fq 'VANILLA_QUICKPLAY_SINGLEPLAYER_REDRIVE scheduled=true level=MRPG-QA' "$DRIVER_LOG"
            grep -Fq 'VANILLA_QUICKPLAY_SINGLEPLAYER_REDRIVE invoked=true level=MRPG-QA' "$DRIVER_LOG"
            echo '[More RPG 2.7.2] VANILLA_QUICKPLAY_REDRIVE_INVOKED'
          fi
        else
          echo '[More RPG 2.7.2] QUICKPLAY_REDRIVE_NOT_ELIGIBLE state_is_not_exact_title_screen_stall'
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
if [[ -s "$DRIVER_LOG" ]]; then
  echo '[More RPG 2.7.2] VANILLA_QUICKPLAY_REDRIVE_FINAL'
  cat "$DRIVER_LOG"
fi
echo "[More RPG 2.7.2] RUNTIME_STAGE1_DIAGNOSTIC_WRAPPER_END rc=$rc"
exit "$rc"
