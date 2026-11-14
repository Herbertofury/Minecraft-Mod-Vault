#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"; PORT="$ROOT/rpg-series-port/rogues-forge-1.20.1"; FRESH="$PORT/.fresh-rogues-forge-server"; LOG="$PORT/rogues-behavior-server.log"; LATEST="$FRESH/logs/latest.log"; FIFO="$FRESH/rogues-qa-console.fifo"; PID=''
cleanup(){ exec 9>&- 2>/dev/null || true; if [[ -n "$PID" ]] && kill -0 "$PID" 2>/dev/null; then kill -TERM "$PID" 2>/dev/null || true; sleep 1; kill -KILL "$PID" 2>/dev/null || true; wait "$PID" 2>/dev/null || true; fi; rm -f "$FIFO"; }; trap cleanup EXIT
wait_marker(){ local marker="$1" timeout_seconds="$2" label="$3" deadline=$((SECONDS+timeout_seconds)); local fatal='ModLoadingException|Failed to create mod instance|Failed to start the minecraft server|NoClassDefFoundError|ClassNotFoundException|MixinApplyError|InvalidMixinException|MixinTransformerError|Registry is already frozen|Can not register to a locked registry|Missing or unsupported mandatory dependencies|Exception in server tick loop|The game crashed'; while (( SECONDS < deadline )); do if grep -Eiq "$fatal" "$LOG" "$LATEST" 2>/dev/null; then echo "[Rogues behavior] fatal runtime signature during $label" >&2; tail -n 500 "$LOG" "$LATEST" 2>/dev/null || true; return 1; fi; if grep -Fq "$marker" "$LATEST" 2>/dev/null || grep -Fq "$marker" "$LOG" 2>/dev/null; then return 0; fi; if [[ -n "$PID" ]] && ! kill -0 "$PID" 2>/dev/null; then wait "$PID" 2>/dev/null || true; tail -n 500 "$LOG" "$LATEST" 2>/dev/null || true; echo "[Rogues behavior] server exited before $label marker: $marker" >&2; return 1; fi; sleep 1; done; tail -n 500 "$LOG" "$LATEST" 2>/dev/null || true; echo "[Rogues behavior] timed out waiting for $label marker: $marker" >&2; return 1; }
send_cmd(){ printf '%s\n' "$1" >&9; }
[[ -x "$FRESH/run.sh" ]] || { echo '[Rogues behavior] certified packaged server missing' >&2; exit 1; }
[[ "$(find "$FRESH/mods" -maxdepth 1 -type f -name '*.jar' | wc -l | tr -d ' ')" = 9 ]] || { echo '[Rogues behavior] expected exact nine-mod packaged runtime' >&2; exit 1; }
ROG_JAR="$(find "$FRESH/mods" -maxdepth 1 -type f -name 'rogues-forge-3.1.1+1.20.1.jar' -print -quit)"; EXPECTED_SHA="$(awk '{print $1}' "$PORT/rogues.sha256")"; ACTUAL_SHA="$(sha256sum "$ROG_JAR" | awk '{print $1}')"; [[ "$EXPECTED_SHA" = '9e8c880f55ab57d91148c0be702a431bad6e312900b25f65c9dbec266e3ca401' && "$ACTUAL_SHA" = "$EXPECTED_SHA" ]] || { echo '[Rogues behavior] certified release identity mismatch' >&2; exit 1; }
grep -Fq 'ActionImpairing.configure(SHOCK.effect, EntityActionsAllowed.STUN);' "$PORT/generated/common/java/net/rogues/effect/RogueEffects.java"
grep -Fq 'ActionImpairing.configure(BEAR_TRAP.effect, ROOT);' "$PORT/generated/common/java/net/rogues/effect/RogueEffects.java"
grep -Fq 'beginDespawn(ATTACK_TICKS);' "$PORT/generated/common/java/net/rogues/entity/BearTrapEntity.java"
echo '[Rogues behavior] ROGUES_ROOT_SHOCK_BEAR_TRAP_WIRING_STATIC_PASS'
rm -rf "$FRESH/logs"; : > "$LOG"; rm -f "$FIFO"; mkfifo "$FIFO"; exec 9<> "$FIFO"; ( cd "$FRESH" && exec ./run.sh nogui < "$FIFO" ) > "$LOG" 2>&1 & PID=$!; wait_marker 'Done (' 180 'server readiness'
send_cmd 'gamerule doMobSpawning false'; send_cmd 'scoreboard objectives add rogqa dummy'; send_cmd 'kill @e[tag=rogqa]'
send_cmd 'summon minecraft:cow 0 100 0 {Tags:["rogqa","rogqa_charge"],NoAI:1b,Silent:1b,PersistenceRequired:1b}'
send_cmd 'execute store result score charge_base rogqa run attribute @e[tag=rogqa_charge,limit=1] minecraft:generic.movement_speed get 100000'
send_cmd 'effect give @e[tag=rogqa_charge,limit=1] rogues:charge 30 0 true'
send_cmd 'execute store result score charge_on rogqa run attribute @e[tag=rogqa_charge,limit=1] minecraft:generic.movement_speed get 100000'
send_cmd 'execute if score charge_on rogqa > charge_base rogqa run say ROGUES_CHARGE_SPEED_APPLY_PASS'
send_cmd 'effect clear @e[tag=rogqa_charge,limit=1] rogues:charge'
send_cmd 'execute store result score charge_off rogqa run attribute @e[tag=rogqa_charge,limit=1] minecraft:generic.movement_speed get 100000'
send_cmd 'execute if score charge_off rogqa = charge_base rogqa run say ROGUES_CHARGE_SPEED_REMOVE_PASS'
send_cmd 'effect give @e[tag=rogqa_charge,limit=1] rogues:stealth_speed 30 0 true'; send_cmd 'effect give @e[tag=rogqa_charge,limit=1] rogues:stealth 30 0 true'; send_cmd 'effect clear @e[tag=rogqa_charge,limit=1] rogues:stealth'; send_cmd 'execute store success score stealth_speed_left rogqa run effect clear @e[tag=rogqa_charge,limit=1] rogues:stealth_speed'; send_cmd 'execute if score stealth_speed_left rogqa matches 0 run say ROGUES_STEALTH_CLEANUP_PASS'
send_cmd 'summon rogues:bear_trap 4 100 0'; sleep 1; send_cmd 'execute if entity @e[type=rogues:bear_trap,limit=1] run say ROGUES_BEAR_TRAP_ENTITY_RUNTIME_PASS'
send_cmd 'summon minecraft:armor_stand 8 100 0 {Tags:["rogqa","rogqa_items"],Invisible:1b,Marker:1b,PersistenceRequired:1b}'; send_cmd 'item replace entity @e[tag=rogqa_items,limit=1] weapon.mainhand with rogues:iron_dagger 1'; send_cmd 'execute if data entity @e[tag=rogqa_items,limit=1] {HandItems:[{id:"rogues:iron_dagger"}]} run say ROGUES_WEAPON_REGISTRY_RUNTIME_PASS'
for marker in ROGUES_CHARGE_SPEED_APPLY_PASS ROGUES_CHARGE_SPEED_REMOVE_PASS ROGUES_STEALTH_CLEANUP_PASS ROGUES_BEAR_TRAP_ENTITY_RUNTIME_PASS ROGUES_WEAPON_REGISTRY_RUNTIME_PASS; do wait_marker "$marker" 30 "$marker"; done
send_cmd 'kill @e[tag=rogqa]'; send_cmd 'kill @e[type=rogues:bear_trap]'; send_cmd 'stop'; exec 9>&-
for _ in $(seq 1 30); do kill -0 "$PID" 2>/dev/null || break; sleep 1; done; if kill -0 "$PID" 2>/dev/null; then echo '[Rogues behavior] packaged server did not stop cleanly' >&2; exit 1; fi; wait "$PID"; PID=''
[[ "$(sha256sum "$ROG_JAR" | awk '{print $1}')" = "$EXPECTED_SHA" ]] || { echo '[Rogues behavior] release identity drifted during behavior QA' >&2; exit 1; }
echo '[Rogues behavior] SERVER_BEHAVIOR_ACCEPTANCE_PASS: exact packaged release proved Charge lifecycle, Stealth cleanup, Bear Trap native entity registration, and representative weapon registration.'
