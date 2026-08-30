#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
PORT="$ROOT/rpg-series-port/paladins-forge-1.20.1"
FRESH="$PORT/.fresh-paladins-forge-server"
LOG="$PORT/paladins-behavior-server.log"
LATEST="$FRESH/logs/latest.log"
FIFO="$FRESH/paladins-qa-console.fifo"
PID=''
cleanup(){ exec 9>&- 2>/dev/null || true; if [[ -n "$PID" ]] && kill -0 "$PID" 2>/dev/null; then kill -TERM "$PID" 2>/dev/null || true; sleep 1; kill -KILL "$PID" 2>/dev/null || true; wait "$PID" 2>/dev/null || true; fi; rm -f "$FIFO"; }
trap cleanup EXIT
wait_marker(){
  local marker="$1"; local timeout_seconds="$2"; local label="$3"; local deadline=$((SECONDS + timeout_seconds))
  local fatal='ModLoadingException|Failed to create mod instance|Failed to start the minecraft server|NoClassDefFoundError|ClassNotFoundException|MixinApplyError|InvalidMixinException|MixinTransformerError|Registry is already frozen|Can not register to a locked registry|Missing or unsupported mandatory dependencies|Exception in server tick loop|The game crashed'
  while (( SECONDS < deadline )); do
    if grep -Eiq "$fatal" "$LOG" "$LATEST" 2>/dev/null; then echo "[Paladins behavior] fatal runtime signature during $label" >&2; tail -n 500 "$LOG" "$LATEST" 2>/dev/null || true; return 1; fi
    if grep -Fq "$marker" "$LATEST" 2>/dev/null || grep -Fq "$marker" "$LOG" 2>/dev/null; then return 0; fi
    if [[ -n "$PID" ]] && ! kill -0 "$PID" 2>/dev/null; then wait "$PID" 2>/dev/null || true; tail -n 500 "$LOG" "$LATEST" 2>/dev/null || true; echo "[Paladins behavior] server exited before $label marker: $marker" >&2; return 1; fi
    sleep 1
  done
  tail -n 500 "$LOG" "$LATEST" 2>/dev/null || true; echo "[Paladins behavior] timed out waiting for $label marker: $marker" >&2; return 1
}
send_cmd(){ printf '%s\n' "$1" >&9; }
[[ -x "$FRESH/run.sh" ]] || { echo '[Paladins behavior] certified packaged server is missing' >&2; exit 1; }
[[ "$(find "$FRESH/mods" -maxdepth 1 -type f -name '*.jar' | wc -l | tr -d ' ')" = 11 ]] || { echo '[Paladins behavior] expected exact 11-mod packaged runtime' >&2; exit 1; }
PAL_JAR="$(find "$FRESH/mods" -maxdepth 1 -type f -name 'paladins-forge-3.1.1+1.20.1.jar' -print -quit)"
EXPECTED_SHA="$(awk '{print $1}' "$PORT/paladins.sha256")"; ACTUAL_SHA="$(sha256sum "$PAL_JAR" | awk '{print $1}')"
[[ "$EXPECTED_SHA" = '95e8f9e074dd1c432f9486c922173b76a3b3bc57667b28fe154e47dba5c374ee' && "$ACTUAL_SHA" = "$EXPECTED_SHA" ]] || { echo '[Paladins behavior] certified release identity mismatch' >&2; exit 1; }
grep -F 'ActionImpairing.configure(JUDGEMENT.effect, EntityActionsAllowed.STUN);' "$PORT/generated/common/java/net/paladins/effect/PaladinEffects.java" >/dev/null
echo '[Paladins behavior] JUDGEMENT_STUN_WIRING_STATIC_PASS (real-player gate follows)'
rm -rf "$FRESH/logs"; : > "$LOG"; rm -f "$FIFO"; mkfifo "$FIFO"; exec 9<> "$FIFO"
( cd "$FRESH" && exec ./run.sh nogui < "$FIFO" ) > "$LOG" 2>&1 & PID=$!
wait_marker 'Done (' 180 'server readiness'
echo '[Paladins behavior] exact certified packaged server ready; executing deterministic game-thread assertions.'
send_cmd 'gamerule doMobSpawning false'; send_cmd 'gamerule doMobLoot false'; send_cmd 'scoreboard objectives add palqa dummy'; send_cmd 'kill @e[tag=palqa]'
send_cmd 'summon minecraft:cow 0 100 0 {Tags:["palqa","palqa_priest"],NoAI:1b,Silent:1b,PersistenceRequired:1b}'
send_cmd 'effect give @e[tag=palqa_priest,limit=1] paladins:priest_absorption 30 1 true'
send_cmd 'execute store result score priest_on palqa run data get entity @e[tag=palqa_priest,limit=1] AbsorptionAmount 10'
send_cmd 'execute if score priest_on palqa matches 40 run say PALADINS_BEHAVIOR_PRIEST_APPLY_PASS'
send_cmd 'effect clear @e[tag=palqa_priest,limit=1] paladins:priest_absorption'
send_cmd 'execute store result score priest_off palqa run data get entity @e[tag=palqa_priest,limit=1] AbsorptionAmount 10'
send_cmd 'execute if score priest_off palqa matches 0 run say PALADINS_BEHAVIOR_PRIEST_REMOVE_PASS'
send_cmd 'fill 9 80 -1 13 80 1 minecraft:stone'
send_cmd 'fill 9 81 -1 13 104 1 minecraft:air'
send_cmd 'summon minecraft:cow 10 100 0 {Tags:["palqa","palqa_control"],Silent:1b,PersistenceRequired:1b}'
send_cmd 'summon minecraft:cow 12 100 0 {Tags:["palqa","palqa_lev"],Silent:1b,PersistenceRequired:1b}'
send_cmd 'effect give @e[tag=palqa_lev,limit=1] paladins:levitate 2 0 true'
sleep 1
send_cmd 'execute store result score control_y palqa run data get entity @e[tag=palqa_control,limit=1] Pos[1] 1000'
send_cmd 'execute store result score levitate_y palqa run data get entity @e[tag=palqa_lev,limit=1] Pos[1] 1000'
send_cmd 'scoreboard players operation levitate_delta palqa = levitate_y palqa'; send_cmd 'scoreboard players operation levitate_delta palqa -= control_y palqa'
send_cmd 'execute if score levitate_delta palqa matches 1000.. run say PALADINS_BEHAVIOR_LEVITATE_MOTION_PASS'
sleep 2
send_cmd 'execute store success score slowfall palqa run effect clear @e[tag=palqa_lev,limit=1] minecraft:slow_falling'
send_cmd 'execute if score slowfall palqa matches 1 run say PALADINS_BEHAVIOR_LEVITATE_SLOW_FALLING_PASS'
for marker in PALADINS_BEHAVIOR_PRIEST_APPLY_PASS PALADINS_BEHAVIOR_PRIEST_REMOVE_PASS PALADINS_BEHAVIOR_LEVITATE_MOTION_PASS PALADINS_BEHAVIOR_LEVITATE_SLOW_FALLING_PASS; do wait_marker "$marker" 30 "$marker"; done
send_cmd 'kill @e[tag=palqa]'; send_cmd 'stop'; exec 9>&-
for _ in $(seq 1 30); do kill -0 "$PID" 2>/dev/null || break; sleep 1; done
if kill -0 "$PID" 2>/dev/null; then echo '[Paladins behavior] packaged server did not stop cleanly after QA' >&2; exit 1; fi
wait "$PID"; PID=''
echo '[Paladins behavior] SERVER_BEHAVIOR_ACCEPTANCE_PASS: Priest Absorption lifecycle and Levitate gravity/Slow Falling semantics passed on the exact certified packaged release.'
