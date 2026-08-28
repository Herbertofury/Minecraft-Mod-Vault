#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
PORT="$ROOT/rpg-series-port/paladins-forge-1.20.1"
RUN="$PORT/forge/run"
BASE_WORLD="$RUN/world"
QA_WORLD="$RUN/saves/Paladins-Player-QA"
CLIENT_LOG="$PORT/paladins-player-behavior.log"
LATEST="$RUN/logs/latest.log"
ROBOT_DIR="${RUNNER_TEMP:-/tmp}/paladins-input-robot"
PID=''
XVFB_PID=''
cleanup(){
  if [[ -n "$PID" ]] && kill -0 "$PID" 2>/dev/null; then kill -TERM "$PID" 2>/dev/null || true; sleep 1; kill -KILL "$PID" 2>/dev/null || true; wait "$PID" 2>/dev/null || true; fi
  if [[ -n "$XVFB_PID" ]] && kill -0 "$XVFB_PID" 2>/dev/null; then kill -TERM "$XVFB_PID" 2>/dev/null || true; wait "$XVFB_PID" 2>/dev/null || true; fi
}
trap cleanup EXIT
wait_marker(){
  local marker="$1"; local timeout_seconds="$2"; local label="$3"; local deadline=$((SECONDS + timeout_seconds))
  local fatal='MixinApplyError|InvalidMixinException|MixinTransformerError|Failed to create mod instance|NoClassDefFoundError|ClassNotFoundException|Exception in thread "Render thread"|The game crashed|Failed to load datapacks'
  while (( SECONDS < deadline )); do
    if grep -Eiq "$fatal" "$CLIENT_LOG" "$LATEST" 2>/dev/null; then echo "[Paladins player QA] fatal runtime signature during $label" >&2; tail -n 500 "$CLIENT_LOG" "$LATEST" 2>/dev/null || true; return 1; fi
    if grep -Fq "$marker" "$LATEST" 2>/dev/null || grep -Fq "$marker" "$CLIENT_LOG" 2>/dev/null; then return 0; fi
    if [[ -n "$PID" ]] && ! kill -0 "$PID" 2>/dev/null; then wait "$PID" 2>/dev/null || true; tail -n 500 "$CLIENT_LOG" "$LATEST" 2>/dev/null || true; echo "[Paladins player QA] client exited before $label marker: $marker" >&2; return 1; fi
    sleep 1
  done
  tail -n 500 "$CLIENT_LOG" "$LATEST" 2>/dev/null || true; echo "[Paladins player QA] timed out waiting for $label marker: $marker" >&2; return 1
}
[[ -f "$BASE_WORLD/level.dat" ]] || { echo '[Paladins player QA] baseline dev-server world missing; run baseline first' >&2; exit 1; }
[[ "$(awk '{print $1}' "$PORT/paladins.sha256")" = '95e8f9e074dd1c432f9486c922173b76a3b3bc57667b28fe154e47dba5c374ee' ]] || { echo '[Paladins player QA] certified release ledger missing/drifted' >&2; exit 1; }
rm -rf "$QA_WORLD"; mkdir -p "$RUN/saves"; cp -a "$BASE_WORLD" "$QA_WORLD"; rm -f "$QA_WORLD/session.lock"
PACK="$QA_WORLD/datapacks/paladins_player_qa"
mkdir -p "$PACK/data/paladins_qa/functions" "$PACK/data/minecraft/tags/functions"
cat > "$PACK/pack.mcmeta" <<'JSON'
{"pack":{"pack_format":15,"description":"Paladins real-player Divine Protection + Judgement acceptance"}}
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
execute if entity @a run function paladins_qa:divine_one_setup
MCF
cat > "$PACK/data/paladins_qa/functions/divine_one_setup.mcfunction" <<'MCF'
gamerule doMobSpawning false
gamerule doDaylightCycle false
time set noon
weather clear
gamemode survival @a
effect clear @a
clear @a
effect give @a minecraft:instant_health 1 10 true
effect give @a paladins:divine_protection 30 0 true
damage @a 5 minecraft:generic
schedule function paladins_qa:divine_one_verify 2t replace
MCF
cat > "$PACK/data/paladins_qa/functions/divine_one_verify.mcfunction" <<'MCF'
execute store result score #divine1_hp palqa run data get entity @a[limit=1] Health 10
execute store success score #divine1_left palqa run effect clear @a paladins:divine_protection
execute if score #divine1_hp palqa matches 200 if score #divine1_left palqa matches 0 run say PALADINS_DIVINE_BLOCK_CONSUME_PASS
schedule function paladins_qa:divine_two_setup 12t replace
MCF
cat > "$PACK/data/paladins_qa/functions/divine_two_setup.mcfunction" <<'MCF'
effect clear @a
effect give @a minecraft:instant_health 1 10 true
effect give @a paladins:divine_protection 30 1 true
damage @a 5 minecraft:generic
schedule function paladins_qa:divine_two_second 12t replace
MCF
cat > "$PACK/data/paladins_qa/functions/divine_two_second.mcfunction" <<'MCF'
execute store result score #divine2_first_hp palqa run data get entity @a[limit=1] Health 10
damage @a 5 minecraft:generic
schedule function paladins_qa:divine_two_verify 2t replace
MCF
cat > "$PACK/data/paladins_qa/functions/divine_two_verify.mcfunction" <<'MCF'
execute store result score #divine2_hp palqa run data get entity @a[limit=1] Health 10
execute store success score #divine2_left palqa run effect clear @a paladins:divine_protection
execute if score #divine2_first_hp palqa matches 200 if score #divine2_hp palqa matches 200 if score #divine2_left palqa matches 0 run say PALADINS_DIVINE_TWO_CHARGE_PASS
schedule function paladins_qa:stun_setup 12t replace
MCF
cat > "$PACK/data/paladins_qa/functions/stun_setup.mcfunction" <<'MCF'
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
say PALADINS_PLAYER_BEHAVIOR_QA_FINISHED
MCF
mkdir -p "$ROBOT_DIR"
cat > "$ROBOT_DIR/PaladinsInputRobot.java" <<'JAVA'
import java.awt.Robot;
import java.awt.event.InputEvent;
import java.awt.event.KeyEvent;
public final class PaladinsInputRobot {
  public static void main(String[] args) throws Exception {
    Robot r = new Robot(); r.setAutoDelay(60); r.mouseMove(640,360);
    r.mousePress(InputEvent.BUTTON1_DOWN_MASK); r.mouseRelease(InputEvent.BUTTON1_DOWN_MASK); Thread.sleep(350);
    r.keyPress(KeyEvent.VK_W); Thread.sleep(1200);
    for(int i=0;i<3;i++){ r.mousePress(InputEvent.BUTTON1_DOWN_MASK); r.mouseRelease(InputEvent.BUTTON1_DOWN_MASK); Thread.sleep(220); }
    r.keyPress(KeyEvent.VK_SPACE); Thread.sleep(160); r.keyRelease(KeyEvent.VK_SPACE); r.keyRelease(KeyEvent.VK_W);
  }
}
JAVA
javac "$ROBOT_DIR/PaladinsInputRobot.java"
rm -rf "$RUN/logs"; mkdir -p "$RUN/logs" "$RUN/config"; printf 'earlyWindowControl = false\n' > "$RUN/config/fml.toml"; : > "$CLIENT_LOG"
export DISPLAY=:99 LIBGL_ALWAYS_SOFTWARE=1 MESA_LOADER_DRIVER_OVERRIDE=llvmpipe ALSOFT_DRIVERS=null
Xvfb "$DISPLAY" -screen 0 1280x720x24 -nolisten tcp > "$PORT/paladins-player-xvfb.log" 2>&1 & XVFB_PID=$!
sleep 1; kill -0 "$XVFB_PID" 2>/dev/null || { echo '[Paladins player QA] Xvfb failed to remain alive' >&2; cat "$PORT/paladins-player-xvfb.log" >&2 || true; exit 1; }
( gradle --no-daemon -p "$PORT" :forge:runClient --args='--width 1280 --height 720 --quickPlaySingleplayer Paladins-Player-QA' </dev/null ) > "$CLIENT_LOG" 2>&1 & PID=$!
wait_marker 'PALADINS_DIVINE_BLOCK_CONSUME_PASS' 210 'Divine Protection one-charge real-player block/consume'
wait_marker 'PALADINS_DIVINE_TWO_CHARGE_PASS' 30 'Divine Protection two-charge real-player decrement/consume'
wait_marker 'PALADINS_JUDGEMENT_STUN_INPUT_READY' 30 'real-player Judgement input readiness'
java -cp "$ROBOT_DIR" PaladinsInputRobot
wait_marker 'PALADINS_JUDGEMENT_STUN_MOVE_BLOCK_PASS' 20 'Judgement movement block'
wait_marker 'PALADINS_JUDGEMENT_STUN_ATTACK_BLOCK_PASS' 20 'Judgement attack block'
wait_marker 'PALADINS_JUDGEMENT_CONTROL_INPUT_READY' 20 'control input readiness'
java -cp "$ROBOT_DIR" PaladinsInputRobot
wait_marker 'PALADINS_JUDGEMENT_CONTROL_MOVE_PASS' 20 'control movement proof'
wait_marker 'PALADINS_JUDGEMENT_CONTROL_ATTACK_PASS' 20 'control attack proof'
wait_marker 'PALADINS_PLAYER_BEHAVIOR_QA_FINISHED' 20 'player QA completion'
grep -Fq 'Backend library: LWJGL' "$LATEST" || { echo '[Paladins player QA] integrated player ran without LWJGL evidence' >&2; exit 1; }
echo '[Paladins player QA] PLAYER_BEHAVIOR_ACCEPTANCE_PASS: exact client/integrated-server runtime proved Divine Protection one/two-charge damage interception and Judgement/STUN input blocking with positive post-clear controls.'
