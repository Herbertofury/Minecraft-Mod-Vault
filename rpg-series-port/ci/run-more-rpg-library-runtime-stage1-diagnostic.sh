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
test -f "$INSPECTOR_SRC/mrpg/qa/StateAgent.java"
test -f "$INSPECTOR_SRC/mrpg/qa/AttachMain.java"
test -f "$INSPECTOR_SRC/mrpg/qa/QuickPlayDriverAgent.java"

# Compile the CI-only JDK Attach probes before Minecraft starts. These classes never enter a mod JAR.
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

descendants() {
  local parent="$1" child
  while IFS= read -r child; do
    [[ -z "$child" ]] && continue
    printf '%s\n' "$child"
    descendants "$child"
  done < <(pgrep -P "$parent" 2>/dev/null || true)
}

mapped_client_candidates() {
  local pid cmdline jcmd_line
  local -A seen=()

  # First prefer JVMs inside the Stage-1 process tree. The actual mapped client is launched through
  # xvfb-run -> Gradle -> Architectury Transformer, so it is not guaranteed to expose the complete
  # ModLauncher argument list in /proc/PID/cmdline.
  while IFS= read -r pid; do
    [[ "$pid" =~ ^[0-9]+$ ]] || continue
    [[ -r "/proc/$pid/cmdline" ]] || continue
    cmdline="$(tr '\0' ' ' < "/proc/$pid/cmdline" 2>/dev/null || true)"
    [[ "$cmdline" == *java* ]] || continue
    if [[ "$cmdline" == *'dev.architectury.transformer.TransformerRuntime'* \
       || "$cmdline" == *'net.fabricmc.devlaunchinjector.Main'* \
       || "$cmdline" == *'forgeclientuserdev'* \
       || "$cmdline" == *'--quickPlaySingleplayer MRPG-QA'* ]]; then
      seen["$pid"]=1
      printf '%s\n' "$pid"
    fi
  done < <(descendants "$BASE_PID")

  # jcmd -l sees the JVM main class even when Forge/Architectury keeps launch-target arguments in
  # generated configuration rather than the OS command line. This is the reliable fallback that
  # run #363 was missing.
  while IFS= read -r jcmd_line; do
    pid="${jcmd_line%% *}"
    [[ "$pid" =~ ^[0-9]+$ ]] || continue
    [[ -n "${seen[$pid]:-}" ]] && continue
    if [[ "$jcmd_line" == *'dev.architectury.transformer.TransformerRuntime'* \
       || "$jcmd_line" == *'net.fabricmc.devlaunchinjector.Main'* \
       || "$jcmd_line" == *'forgeclientuserdev'* \
       || "$jcmd_line" == *'--quickPlaySingleplayer MRPG-QA'* ]]; then
      seen["$pid"]=1
      printf '%s\n' "$pid"
    fi
  done < <(jcmd -l 2>/dev/null || true)
}

dump_client_probe_processes() {
  echo '[More RPG 2.7.2] MAPPED_CLIENT_STATE_PROBE_PROCESS_DIAGNOSTICS_BEGIN'
  echo '[More RPG QA] Stage-1 descendant process tree:'
  while IFS= read -r pid; do
    [[ "$pid" =~ ^[0-9]+$ ]] || continue
    ps -o pid=,ppid=,comm=,args= -p "$pid" 2>/dev/null || true
  done < <(descendants "$BASE_PID")
  echo '[More RPG QA] Live JVM list:'
  jcmd -l 2>/dev/null | sed 's/^/[More RPG QA jcmd-l] /' || true
  echo '[More RPG 2.7.2] MAPPED_CLIENT_STATE_PROBE_PROCESS_DIAGNOSTICS_END'
}

echo '[More RPG 2.7.2] RUNTIME_STAGE1_DIAGNOSTIC_WRAPPER_BEGIN source=post-run-364-gated-redrive'
bash "$BASE" &
BASE_PID=$!

RESOURCE_MARKER_SEEN=0
PROBE_ATTEMPTED=0
PROBE_CAPTURED=0
for _ in $(seq 1 210); do
  if ! kill -0 "$BASE_PID" 2>/dev/null; then
    break
  fi
  if [[ -f "$CLIENT_LOG" ]] && grep -Fq 'Reloading ResourceManager' "$CLIENT_LOG"; then
    RESOURCE_MARKER_SEEN=1
    echo '[More RPG 2.7.2] MAPPED_CLIENT_RESOURCE_MARKER_SEEN'

    # The reload-begin marker precedes atlas/model finalization. Give the client a bounded settle
    # interval so the snapshot reports the stable screen/world state rather than LoadingOverlay.
    for _settle in $(seq 1 15); do
      kill -0 "$BASE_PID" 2>/dev/null || break
      sleep 1
    done
    kill -0 "$BASE_PID" 2>/dev/null || break

    mapfile -t CLIENT_CANDIDATES < <(mapped_client_candidates)
    if ((${#CLIENT_CANDIDATES[@]} == 0)); then
      echo '[More RPG 2.7.2] MAPPED_CLIENT_STATE_PROBE_NO_JVM_CANDIDATE_AFTER_RESOURCE_MARKER'
      dump_client_probe_processes
      break
    fi

    for CLIENT_PID in "${CLIENT_CANDIDATES[@]}"; do
      [[ "$CLIENT_PID" =~ ^[0-9]+$ ]] || continue
      kill -0 "$CLIENT_PID" 2>/dev/null || continue
      PROBE_ATTEMPTED=1
      CANDIDATE_STATE="$STATE_LOG.$CLIENT_PID"
      : > "$CANDIDATE_STATE"
      echo "[More RPG 2.7.2] MAPPED_CLIENT_STATE_PROBE_ATTACH_BEGIN pid=$CLIENT_PID"
      jcmd "$CLIENT_PID" VM.command_line 2>/dev/null | sed 's/^/[More RPG QA jcmd] /' || true
      set +e
      java --add-modules jdk.attach -cp "$INSPECTOR_CLASSES" \
        mrpg.qa.AttachMain "$CLIENT_PID" "$STATE_AGENT_JAR" "$CANDIDATE_STATE"
      probe_rc=$?
      set -e
      if [[ "$probe_rc" -ne 0 ]]; then
        echo "[More RPG 2.7.2] MAPPED_CLIENT_STATE_PROBE_ATTACH_FAILED pid=$CLIENT_PID rc=$probe_rc"
        continue
      fi
      if [[ ! -s "$CANDIDATE_STATE" ]]; then
        echo "[More RPG 2.7.2] MAPPED_CLIENT_STATE_PROBE_EMPTY pid=$CLIENT_PID"
        continue
      fi
      cat "$CANDIDATE_STATE"
      if grep -Fq 'error=MinecraftClient_not_loaded' "$CANDIDATE_STATE" \
        || grep -Fq 'error=MinecraftClient_instance_null' "$CANDIDATE_STATE"; then
        echo "[More RPG 2.7.2] MAPPED_CLIENT_STATE_PROBE_WRONG_JVM pid=$CLIENT_PID"
        continue
      fi
      cp "$CANDIDATE_STATE" "$STATE_LOG"
      PROBE_CAPTURED=1
      echo "[More RPG 2.7.2] MAPPED_CLIENT_STATE_PROBE_CAPTURED pid=$CLIENT_PID"

      # Only the exact, already-observed stable title-screen stall is eligible for a one-shot
      # vanilla QuickPlay.startSingleplayer redrive. Every other state remains diagnostic-only.
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
          echo "[More RPG 2.7.2] VANILLA_QUICKPLAY_REDRIVE_ATTACH_FAILED pid=$CLIENT_PID rc=$driver_rc"
          kill -TERM "$CLIENT_PID" 2>/dev/null || true
        else
          for _driver_wait in $(seq 1 10); do
            grep -Fq 'VANILLA_QUICKPLAY_SINGLEPLAYER_REDRIVE invoked=true level=MRPG-QA' "$DRIVER_LOG" && break
            grep -Fq 'VANILLA_QUICKPLAY_SINGLEPLAYER_REDRIVE setup_error=' "$DRIVER_LOG" && break
            grep -Fq 'VANILLA_QUICKPLAY_SINGLEPLAYER_REDRIVE error=' "$DRIVER_LOG" && break
            sleep 1
          done
          cat "$DRIVER_LOG"
          if grep -Fq 'VANILLA_QUICKPLAY_SINGLEPLAYER_REDRIVE scheduled=true level=MRPG-QA' "$DRIVER_LOG" \
            && grep -Fq 'VANILLA_QUICKPLAY_SINGLEPLAYER_REDRIVE invoked=true level=MRPG-QA' "$DRIVER_LOG"; then
            echo '[More RPG 2.7.2] VANILLA_QUICKPLAY_REDRIVE_INVOKED'
          else
            echo '[More RPG 2.7.2] VANILLA_QUICKPLAY_REDRIVE_FAILED stopping_failed_client=true'
            kill -TERM "$CLIENT_PID" 2>/dev/null || true
          fi
        fi
      else
        echo '[More RPG 2.7.2] QUICKPLAY_REDRIVE_NOT_ELIGIBLE state_is_not_exact_title_screen_stall'
      fi
      break
    done
    [[ "$PROBE_CAPTURED" -eq 1 ]] || dump_client_probe_processes
    break
  fi
  sleep 1
done

if [[ "$RESOURCE_MARKER_SEEN" -eq 0 ]]; then
  echo '[More RPG 2.7.2] MAPPED_CLIENT_STATE_PROBE_NOT_REACHED resource_marker_missing'
elif [[ "$PROBE_ATTEMPTED" -eq 0 ]]; then
  echo '[More RPG 2.7.2] MAPPED_CLIENT_STATE_PROBE_NOT_REACHED jvm_candidate_missing'
elif [[ "$PROBE_CAPTURED" -eq 0 ]]; then
  echo '[More RPG 2.7.2] MAPPED_CLIENT_STATE_PROBE_NOT_REACHED attach_or_candidate_validation_failed'
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
