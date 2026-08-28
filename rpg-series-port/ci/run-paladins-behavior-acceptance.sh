#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
PORT="$ROOT/rpg-series-port/paladins-forge-1.20.1"
FRESH="$PORT/.fresh-paladins-forge-server"
LOG="$PORT/paladins-behavior-server.log"
LATEST="$FRESH/logs/latest.log"
FIFO="$FRESH/paladins-qa-console.fifo"
PID=''

cleanup() {
  exec 9>&- 2>/dev/null || true
  if [[ -n "$PID" ]] && kill -0 "$PID" 2>/dev/null; then
    kill -TERM "$PID" 2>/dev/null || true
    sleep 1
    kill -KILL "$PID" 2>/dev/null || true
    wait "$PID" 2>/dev/null || true
  fi
  rm -f "$FIFO"
}
trap cleanup EXIT

wait_marker() {
  local marker="$1"
  local timeout_seconds="$2"
  local label="$3"
  local deadline=$((SECONDS + timeout_seconds))
  local fatal='ModLoadingException|Failed to create mod instance|Failed to start the minecraft server|NoClassDefFoundError|ClassNotFoundException|MixinApplyError|InvalidMixinException|MixinTransformerError|Registry is already frozen|Can not register to a locked registry|Missing or unsupported mandatory dependencies|Exception in server tick loop|The game crashed'
  while (( SECONDS < deadline )); do
    if grep -Eiq "$fatal" "$LOG" "$LATEST" 2>/dev/null; then
      echo "[Paladins behavior] fatal runtime signature during $label" >&2
      tail -n 500 "$LOG" "$LATEST" 2>/dev/null || true
      return 1
    fi
    if grep -Fq "$marker" "$LATEST" 2>/dev/null || grep -Fq "$marker" "$LOG" 2>/dev/null; then
      return 0
    fi
    if [[ -n "$PID" ]] && ! kill -0 "$PID" 2>/dev/null; then
      wait "$PID" 2>/dev/null || true
      tail -n 500 "$LOG" "$LATEST" 2>/dev/null || true
      echo "[Paladins behavior] server exited before $label marker: $marker" >&2
      return 1
    fi
    sleep 1
  done
  tail -n 500 "$LOG" "$LATEST" 2>/dev/null || true
  echo "[Paladins behavior] timed out waiting for $label marker: $marker" >&2
  return 1
}

send_cmd() {
  printf '%s\n' "$1" >&9
}

[[ -x "$FRESH/run.sh" ]] || { echo '[Paladins behavior] fresh packaged server from baseline acceptance is missing' >&2; exit 1; }
[[ "$(find "$FRESH/mods" -maxdepth 1 -type f -name '*.jar' | wc -l | tr -d ' ')" = 11 ]] || {
  echo '[Paladins behavior] expected exact 11-mod packaged runtime from baseline acceptance' >&2; exit 1; }
PAL_JAR="$(find "$FRESH/mods" -maxdepth 1 -type f -name 'paladins-forge-3.1.1+1.20.1.jar' -print -quit)"
[[ -f "$PAL_JAR" ]] || { echo '[Paladins behavior] exact Paladins release is missing from packaged server' >&2; exit 1; }
EXPECTED_SHA="$(awk '{print $1}' "$PORT/paladins.sha256")"
ACTUAL_SHA="$(sha256sum "$PAL_JAR" | awk '{print $1}')"
[[ "$ACTUAL_SHA" = "$EXPECTED_SHA" ]] || {
  echo "[Paladins behavior] packaged release identity drifted: expected=$EXPECTED_SHA actual=$ACTUAL_SHA" >&2; exit 1; }

# Judgement cannot be honestly action-tested without a real player. Keep a fail-closed wiring assertion here;
# integrated-player STUN remains a distinct graduation gate instead of being simulated by a mob command.
grep -F 'ActionImpairing.configure(JUDGEMENT.effect, EntityActionsAllowed.STUN);' \
  "$PORT/generated/common/java/net/paladins/effect/PaladinEffects.java" >/dev/null
echo '[Paladins behavior] JUDGEMENT_STUN_WIRING_STATIC_PASS (player action gate still required)'

rm -rf "$FRESH/logs"
: > "$LOG"
rm -f "$FIFO"
mkfifo "$FIFO"
exec 9<> "$FIFO"
( cd "$FRESH" && exec ./run.sh nogui < "$FIFO" ) > "$LOG" 2>&1 & PID=$!
wait_marker 'Done (' 180 'server readiness'
echo '[Paladins behavior] exact packaged server ready; executing deterministic game-thread assertions.'

# Keep the fixture deterministic and non-destructive.
send_cmd 'gamerule doMobSpawning false'
send_cmd 'gamerule doMobLoot false'
send_cmd 'scoreboard objectives add palqa dummy'
send_cmd 'kill @e[tag=palqa]'

# Priest Absorption: amplifier 1 must grant exactly 4.0 absorption hearts, then remove exactly that grant.
send_cmd 'summon minecraft:cow 0 100 0 {Tags:["palqa","palqa_priest"],NoAI:1b,Silent:1b,PersistenceRequired:1b}'
send_cmd 'effect give @e[tag=palqa_priest,limit=1] paladins:priest_absorption 30 1 true'
send_cmd 'execute store result score priest_on palqa run data get entity @e[tag=palqa_priest,limit=1] AbsorptionAmount 10'
send_cmd 'execute if score priest_on palqa matches 40 run say PALADINS_BEHAVIOR_PRIEST_APPLY_PASS'
send_cmd 'effect clear @e[tag=palqa_priest,limit=1] paladins:priest_absorption'
send_cmd 'execute store result score priest_off palqa run data get entity @e[tag=palqa_priest,limit=1] AbsorptionAmount 10'
send_cmd 'execute if score priest_off palqa matches 0 run say PALADINS_BEHAVIOR_PRIEST_REMOVE_PASS'

# Divine Protection: one stack must fully intercept generic damage and be consumed.
send_cmd 'summon minecraft:cow 5 100 0 {Tags:["palqa","palqa_divine"],NoAI:1b,Silent:1b,PersistenceRequired:1b}'
send_cmd 'effect give @e[tag=palqa_divine,limit=1] paladins:divine_protection 30 0 true'
send_cmd 'damage @e[tag=palqa_divine,limit=1] 5 minecraft:generic'
send_cmd 'execute store result score divine_hp palqa run data get entity @e[tag=palqa_divine,limit=1] Health 10'
send_cmd 'execute store success score divine_left palqa run effect clear @e[tag=palqa_divine,limit=1] paladins:divine_protection'
send_cmd 'execute if score divine_hp palqa matches 100 if score divine_left palqa matches 0 run say PALADINS_BEHAVIOR_DIVINE_BLOCK_CONSUME_PASS'

# A two-charge instance must still block the first hit but remain present at one lower stack.
send_cmd 'summon minecraft:cow 7 100 0 {Tags:["palqa","palqa_divine2"],NoAI:1b,Silent:1b,PersistenceRequired:1b}'
send_cmd 'effect give @e[tag=palqa_divine2,limit=1] paladins:divine_protection 30 1 true'
send_cmd 'damage @e[tag=palqa_divine2,limit=1] 5 minecraft:generic'
send_cmd 'execute store result score divine2_hp palqa run data get entity @e[tag=palqa_divine2,limit=1] Health 10'
send_cmd 'execute store success score divine2_left palqa run effect clear @e[tag=palqa_divine2,limit=1] paladins:divine_protection'
send_cmd 'execute if score divine2_hp palqa matches 100 if score divine2_left palqa matches 1 run say PALADINS_BEHAVIOR_DIVINE_DECREMENT_PASS'

# Levitate: compare identical falling cows, then prove final-tick handoff to vanilla Slow Falling.
send_cmd 'summon minecraft:cow 10 100 0 {Tags:["palqa","palqa_control"],NoAI:1b,Silent:1b,PersistenceRequired:1b}'
send_cmd 'summon minecraft:cow 12 100 0 {Tags:["palqa","palqa_lev"],NoAI:1b,Silent:1b,PersistenceRequired:1b}'
send_cmd 'effect give @e[tag=palqa_lev,limit=1] paladins:levitate 1 0 true'
sleep 2
send_cmd 'execute store result score control_y palqa run data get entity @e[tag=palqa_control,limit=1] Pos[1] 1000'
send_cmd 'execute store result score levitate_y palqa run data get entity @e[tag=palqa_lev,limit=1] Pos[1] 1000'
send_cmd 'scoreboard players operation levitate_delta palqa = levitate_y palqa'
send_cmd 'scoreboard players operation levitate_delta palqa -= control_y palqa'
send_cmd 'execute if score levitate_delta palqa matches 1000.. run say PALADINS_BEHAVIOR_LEVITATE_MOTION_PASS'
send_cmd 'execute store success score slowfall palqa run effect clear @e[tag=palqa_lev,limit=1] minecraft:slow_falling'
send_cmd 'execute if score slowfall palqa matches 1 run say PALADINS_BEHAVIOR_LEVITATE_SLOW_FALLING_PASS'

for marker in \
  PALADINS_BEHAVIOR_PRIEST_APPLY_PASS \
  PALADINS_BEHAVIOR_PRIEST_REMOVE_PASS \
  PALADINS_BEHAVIOR_DIVINE_BLOCK_CONSUME_PASS \
  PALADINS_BEHAVIOR_DIVINE_DECREMENT_PASS \
  PALADINS_BEHAVIOR_LEVITATE_MOTION_PASS \
  PALADINS_BEHAVIOR_LEVITATE_SLOW_FALLING_PASS; do
  wait_marker "$marker" 30 "$marker"
done

send_cmd 'kill @e[tag=palqa]'
send_cmd 'stop'
exec 9>&-
for _ in $(seq 1 30); do
  kill -0 "$PID" 2>/dev/null || break
  sleep 1
done
if kill -0 "$PID" 2>/dev/null; then
  echo '[Paladins behavior] packaged server did not stop cleanly after QA' >&2
  exit 1
fi
wait "$PID"
PID=''

echo '[Paladins behavior] SERVER_BEHAVIOR_ACCEPTANCE_PASS: Priest Absorption lifecycle, Divine Protection intercept/decrement, and Levitate motion/Slow Falling handoff passed on the exact packaged release.'
