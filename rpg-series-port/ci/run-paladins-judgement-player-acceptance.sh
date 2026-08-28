#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
PORT="$ROOT/rpg-series-port/paladins-forge-1.20.1"
RUN="$PORT/forge/run"
BASE_WORLD="$RUN/world"
QA_WORLD="$RUN/saves/Paladins-Judgement-QA"
CLIENT_LOG="$PORT/paladins-judgement-player.log"
LATEST="$RUN/logs/latest.log"
ROBOT_DIR="${RUNNER_TEMP:-/tmp}/paladins-input-robot"
PID=''
XVFB_PID=''

cleanup() {
  if [[ -n "$PID" ]] && kill -0 "$PID" 2>/dev/null; then
    kill -TERM "$PID" 2>/dev/null || true
    sleep 1
    kill -KILL "$PID" 2>/dev/null || true
    wait "$PID" 2>/dev/null || true
  fi
  if [[ -n "$XVFB_PID" ]] && kill -0 "$XVFB_PID" 2>/dev/null; then
    kill -TERM "$XVFB_PID" 2>/dev/null || true
    wait "$XVFB_PID" 2>/dev/null || true
  fi
}
trap cleanup EXIT

wait_marker() {
  local marker="$1"
  local timeout_seconds="$2"
  local label="$3"
  local deadline=$((SECONDS + timeout_seconds))
  local fatal='MixinApplyError|InvalidMixinException|MixinTransformerError|Failed to create mod instance|NoClassDefFoundError|ClassNotFoundException|Exception in thread "Render thread"|The game crashed|Failed to load datapacks'
  while (( SECONDS < deadline )); do
    if grep -Eiq "$fatal" "$CLIENT_LOG" "$LATEST" 2>/dev/null; then
      echo "[Paladins player QA] fatal runtime signature during $label" >&2
      tail -n 500 "$CLIENT_LOG" "$LATEST" 2>/dev/null || true
      return 1
    fi
    if grep -Fq "$marker" "$LATEST" 2>/dev/null || grep -Fq "$marker" "$CLIENT_LOG" 2>/dev/null; then
      return 0
    fi
    if [[ -n "$PID" ]] && ! kill -0 "$PID" 2>/dev/null; then
      wait "$PID" 2>/dev/null || true
      tail -n 500 "$CLIENT_LOG" "$LATEST" 2>/dev/null || true
      echo "[Paladins player QA] client exited before $label marker: $marker" >&2
      return 1
    fi
    sleep 1
  done
  tail -n 500 "$CLIENT_LOG" "$LATEST" 2>/dev/null || true
  echo "[Paladins player QA] timed out waiting for $label marker: $marker" >&2
  return 1
}

[[ -f "$BASE_WORLD/level.dat" ]] || {
  echo '[Paladins player QA] baseline acceptance dev-server world is missing; run baseline acceptance first' >&2
  exit 1
}
rm -rf "$QA_WORLD"
mkdir -p "$RUN/saves"
cp -a "$BASE_WORLD" "$QA_WORLD"
rm -f "$QA_WORLD/session.lock"
PACK="$QA_WORLD/datapacks/paladins_judgement_qa"
mkdir -p "$PACK/data/paladins_qa/functions" "$PACK/data/minecraft/tags/functions"
cat > "$PACK/pack.mcmeta" <<'JSON'
{"pack":{"pack_format":15,"description":"Paladins Judgement integrated-player acceptance"}}
JSON
cat > "$PACK/data/minecraft/tags/functions/load.json" <<'JSON'
{"values":["paladins_qa:load"]}
JSON
cat > "$PACK/data/paladins_qa/functions/load.mcfunction" <<'MCF'
scoreboard objectives add palqa dummy
schedule function paladins_qa:await_player 1t replace
MCF
cat > "$PACK/data/paladins_qa/functions/await_player.mcfunction" <<'MCF'
execute unless entity @a run schedule function paladins_qa:await_player 1t replace
execute if entity @a run function paladins_qa:stun_setup
MCF
cat > "$PACK/data/paladins_qa/functions/stun_setup.mcfunction" <<'MCF'
gamerule doMobSpawning false
gamerule doDaylightCycle false
time set noon
weather clear
gamemode survival @a
kill @e[tag=palqa_judgement_target]
effect clear @a
clear @a
fill -2 99 -2 2 99 6 minecraft:stone
fill -2 100 -2 2 103 6 minecraft:air
tp @a 0.5 100 0.5 0 0
give @a minecraft:stone_sword 1
summon minecraft:husk 0.5 100 3.5 {Tags:["palqa_judgement_target"],NoAI:1b,Silent:1b,PersistenceRequired:1b}
effect give @a paladins:judgement 20 0 true
say PALADINS_JUDGEMENT_STUN_INPUT_READY
schedule function paladins_qa:stun_verify 80t replace
MCF
cat > "$PACK/data/paladins_qa/functions/stun_verify.mcfunction" <<'MCF'
execute positioned 0.5 100 0.5 if entity @a[distance=..0.25] run say PALADINS_JUDGEMENT_STUN_MOVE_BLOCK_PASS
execute store result score #stun_hp palqa run data get entity @e[tag=palqa_judgement_target,limit=1] Health 10
execute if score #stun_hp palqa matches 200 run say PALADINS_JUDGEMENT_STUN_ATTACK_BLOCK_PASS
effect clear @a paladins:judgement
kill @e[tag=palqa_judgement_target]
tp @a 0.5 100 0.5 0 0
summon minecraft:husk 0.5 100 3.5 {Tags:["palqa_judgement_target"],NoAI:1b,Silent:1b,PersistenceRequired:1b}
say PALADINS_JUDGEMENT_CONTROL_INPUT_READY
schedule function paladins_qa:control_verify 80t replace
MCF
cat > "$PACK/data/paladins_qa/functions/control_verify.mcfunction" <<'MCF'
execute positioned 0.5 100 0.5 unless entity @a[distance=..0.50] run say PALADINS_JUDGEMENT_CONTROL_MOVE_PASS
execute store result score #control_hp palqa run data get entity @e[tag=palqa_judgement_target,limit=1] Health 10
execute if score #control_hp palqa matches ..199 run say PALADINS_JUDGEMENT_CONTROL_ATTACK_PASS
say PALADINS_JUDGEMENT_PLAYER_QA_FINISHED
MCF

mkdir -p "$ROBOT_DIR"
cat > "$ROBOT_DIR/PaladinsInputRobot.java" <<'JAVA'
import java.awt.Robot;
import java.awt.event.InputEvent;
import java.awt.event.KeyEvent;
public final class PaladinsInputRobot {
  public static void main(String[] args) throws Exception {
    Robot r = new Robot();
    r.setAutoDelay(60);
    r.mouseMove(640, 360);
    r.mousePress(InputEvent.BUTTON1_DOWN_MASK);
    r.mouseRelease(InputEvent.BUTTON1_DOWN_MASK);
    Thread.sleep(350);
    r.keyPress(KeyEvent.VK_W);
    Thread.sleep(1200);
    for (int i = 0; i < 3; i++) {
      r.mousePress(InputEvent.BUTTON1_DOWN_MASK);
      r.mouseRelease(InputEvent.BUTTON1_DOWN_MASK);
      Thread.sleep(220);
    }
    r.keyPress(KeyEvent.VK_SPACE);
    Thread.sleep(160);
    r.keyRelease(KeyEvent.VK_SPACE);
    r.keyRelease(KeyEvent.VK_W);
  }
}
JAVA
javac "$ROBOT_DIR/PaladinsInputRobot.java"

rm -rf "$RUN/logs"
mkdir -p "$RUN/logs" "$RUN/config"
printf 'earlyWindowControl = false\n' > "$RUN/config/fml.toml"
: > "$CLIENT_LOG"
export DISPLAY=:99
export LIBGL_ALWAYS_SOFTWARE=1
export MESA_LOADER_DRIVER_OVERRIDE=llvmpipe
export ALSOFT_DRIVERS=null
Xvfb "$DISPLAY" -screen 0 1280x720x24 -nolisten tcp > "$PORT/paladins-judgement-xvfb.log" 2>&1 & XVFB_PID=$!
sleep 1
kill -0 "$XVFB_PID" 2>/dev/null || {
  echo '[Paladins player QA] Xvfb failed to remain alive' >&2
  cat "$PORT/paladins-judgement-xvfb.log" >&2 || true
  exit 1
}
( gradle --no-daemon -p "$PORT" :forge:runClient --args='--width 1280 --height 720 --quickPlaySingleplayer Paladins-Judgement-QA' </dev/null ) > "$CLIENT_LOG" 2>&1 & PID=$!

# This marker is emitted only by the integrated server after @a resolves to the real Quick Play player.
wait_marker 'PALADINS_JUDGEMENT_STUN_INPUT_READY' 210 'real-player stun input readiness'
java -cp "$ROBOT_DIR" PaladinsInputRobot
wait_marker 'PALADINS_JUDGEMENT_STUN_MOVE_BLOCK_PASS' 20 'stun movement block'
wait_marker 'PALADINS_JUDGEMENT_STUN_ATTACK_BLOCK_PASS' 20 'stun attack block'
wait_marker 'PALADINS_JUDGEMENT_CONTROL_INPUT_READY' 20 'control input readiness'
java -cp "$ROBOT_DIR" PaladinsInputRobot
wait_marker 'PALADINS_JUDGEMENT_CONTROL_MOVE_PASS' 20 'control movement proof'
wait_marker 'PALADINS_JUDGEMENT_CONTROL_ATTACK_PASS' 20 'control attack proof'
wait_marker 'PALADINS_JUDGEMENT_PLAYER_QA_FINISHED' 20 'player QA completion'
grep -Fq 'Backend library: LWJGL' "$LATEST"

echo '[Paladins player QA] JUDGEMENT_REAL_PLAYER_PASS: a real Quick Play player could neither move nor attack while Judgement/STUN was active, while the identical input moved and damaged the target immediately after the effect was cleared.'
