#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
CI="$ROOT/rpg-series-port/ci"
source "$CI/ROGUES_GRADUATION.env"
OLD_JAR_SHA='9e8c880f55ab57d91148c0be702a431bad6e312900b25f65c9dbec266e3ca401'
for script in run-rogues-acceptance.sh run-rogues-release-certification.sh run-rogues-server-behavior.sh run-rogues-player-behavior-acceptance.sh; do
  sed -i "s/$OLD_JAR_SHA/$ROGUES_EXPECTED_JAR_SHA/g" "$CI/$script"
done
if [[ "$ROGUES_EXPECTED_SOURCE_SHA" != '__CAPTURE_AFTER_FIRST_DEEP__' ]]; then
  sed -i "s/__CAPTURE_AFTER_FIRST_DEEP__/$ROGUES_EXPECTED_SOURCE_SHA/g" "$CI/run-rogues-release-certification.sh"
fi
# Bash evaluates all RHS expressions in a `local` command before assigning the
# earlier names. With `set -u`, the original one-line wait_marker declarations
# therefore referenced timeout_seconds while it was still unset. Normalize both
# semantic harnesses before execution and fail closed if the pinned shape drifts.
OLD_WAIT='wait_marker(){ local marker="$1" timeout_seconds="$2" label="$3" deadline=$((SECONDS+timeout_seconds));'
NEW_WAIT='wait_marker(){ local marker="$1" timeout_seconds="$2" label="$3"; local deadline=$((SECONDS+timeout_seconds));'
for script in run-rogues-server-behavior.sh run-rogues-player-behavior-acceptance.sh; do
  grep -Fq "$OLD_WAIT" "$CI/$script" || { echo "[Rogues deep acceptance] expected wait_marker harness shape missing: $script" >&2; exit 2; }
  python3 - "$CI/$script" "$OLD_WAIT" "$NEW_WAIT" <<'PY'
from pathlib import Path
import sys
path=Path(sys.argv[1]); old=sys.argv[2]; new=sys.argv[3]
text=path.read_text()
if text.count(old) != 1:
    raise SystemExit(f'[Rogues deep acceptance] expected exactly one wait_marker seam in {path.name}, found {text.count(old)}')
path.write_text(text.replace(old,new,1))
PY
  grep -Fq "$NEW_WAIT" "$CI/$script"
  bash -n "$CI/$script"
done
# Real-player Stealth visibility must observe hostile target acquisition directly,
# without modifying the hostile AI being measured. Run #235 showed
# movement_speed=0 suppresses the ordinary hostile control, while run #236 showed
# mounting the hostile in a boat also suppresses the ordinary control. Both are
# invalid discriminators rather than product-defect evidence. Java Edition 1.19.4+
# exposes the current mob attack target natively through `execute on target`, so
# the 1.20.1 packaged server can interrogate the exact live target relation. The
# control husk must target the real player while they are still >=5 blocks apart;
# the identically spawned Stealth husk must have no player target at the same
# distance and Stealth itself must remain active. This directly tests the target
# seam without movement, bow timing, mounts, helper mods, or product instrumentation.
PLAYER_QA="$CI/run-rogues-player-behavior-acceptance.sh"
python3 - "$PLAYER_QA" <<'PY'
from pathlib import Path
import sys
path=Path(sys.argv[1])
text=path.read_text()
old="""send_cmd 'tp @a 20.5 100 0.5 0 0'; send_cmd 'effect give @a minecraft:instant_health 1 10 true'; send_cmd 'summon minecraft:skeleton 20.5 100 6.5 {Tags:[\"rogp\",\"rogp_control_skeleton\"],Silent:1b,PersistenceRequired:1b}'; sleep 6; send_cmd 'execute store result score stealth_control_hp rogp run data get entity @a[limit=1] Health 10'; send_cmd 'execute if score stealth_control_hp rogp matches ..199 run say ROGUES_STEALTH_VISIBILITY_CONTROL_PASS'; wait_marker 'ROGUES_STEALTH_VISIBILITY_CONTROL_PASS' 20 'non-stealth hostile acquisition control'; send_cmd 'kill @e[tag=rogp_control_skeleton]'; send_cmd 'effect give @a minecraft:instant_health 1 10 true'; send_cmd 'tp @a 30.5 100 0.5 0 0'; send_cmd 'effect give @a rogues:stealth 30 0 true'; send_cmd 'summon minecraft:skeleton 30.5 100 6.5 {Tags:[\"rogp\",\"rogp_stealth_skeleton\"],Silent:1b,PersistenceRequired:1b}'; sleep 6; send_cmd 'execute store result score stealth_hp rogp run data get entity @a[limit=1] Health 10'; send_cmd 'execute store success score stealth_present rogp run effect clear @a[limit=1] rogues:stealth'; send_cmd 'execute if score stealth_hp rogp matches 200 if score stealth_present rogp matches 1 run say ROGUES_STEALTH_VISIBILITY_RUNTIME_PASS'; wait_marker 'ROGUES_STEALTH_VISIBILITY_RUNTIME_PASS' 20 'stealth hostile visibility reduction'; send_cmd 'kill @e[tag=rogp_stealth_skeleton]'"""
new="""send_cmd 'tp @a 20.5 100 0.5 0 0'; sleep 2; send_cmd 'execute positioned 20.5 100 0.5 if entity @a[distance=..1.0,limit=1] run say ROGUES_CONTROL_ARENA_POSITION_READY'; wait_marker 'ROGUES_CONTROL_ARENA_POSITION_READY' 20 'control arena position'; send_cmd 'execute at @a[limit=1] run summon minecraft:husk ~ ~ ~6 {Tags:[\"rogp\",\"rogp_control_husk\"],Silent:1b,PersistenceRequired:1b}'; sleep 0.5; send_cmd 'scoreboard players set control_target rogp 0'; send_cmd 'execute as @e[type=minecraft:husk,tag=rogp_control_husk,limit=1,sort=nearest] at @s on target if entity @s[type=minecraft:player,distance=5..] run scoreboard players set control_target rogp 1'; send_cmd 'execute if score control_target rogp matches 1 run say ROGUES_STEALTH_VISIBILITY_CONTROL_PASS'; wait_marker 'ROGUES_STEALTH_VISIBILITY_CONTROL_PASS' 20 'non-stealth hostile direct target control'; send_cmd 'kill @e[tag=rogp_control_husk]'; send_cmd 'tp @a 30.5 100 0.5 0 0'; sleep 2; send_cmd 'execute positioned 30.5 100 0.5 if entity @a[distance=..1.0,limit=1] run say ROGUES_STEALTH_ARENA_POSITION_READY'; wait_marker 'ROGUES_STEALTH_ARENA_POSITION_READY' 20 'Stealth arena position'; send_cmd 'effect give @a rogues:stealth 30 0 true'; send_cmd 'execute at @a[limit=1] run summon minecraft:husk ~ ~ ~6 {Tags:[\"rogp\",\"rogp_stealth_husk\"],Silent:1b,PersistenceRequired:1b}'; sleep 0.5; send_cmd 'scoreboard players set stealth_target rogp 0'; send_cmd 'scoreboard players set stealth_distance rogp 0'; send_cmd 'execute as @e[type=minecraft:husk,tag=rogp_stealth_husk,limit=1,sort=nearest] on target if entity @s[type=minecraft:player] run scoreboard players set stealth_target rogp 1'; send_cmd 'execute as @e[type=minecraft:husk,tag=rogp_stealth_husk,limit=1,sort=nearest] at @s if entity @a[distance=5..,limit=1] run scoreboard players set stealth_distance rogp 1'; send_cmd 'execute store success score stealth_present rogp run effect clear @a[limit=1] rogues:stealth'; send_cmd 'execute if score stealth_target rogp matches 0 if score stealth_distance rogp matches 1 if score stealth_present rogp matches 1 run say ROGUES_STEALTH_VISIBILITY_RUNTIME_PASS'; wait_marker 'ROGUES_STEALTH_VISIBILITY_RUNTIME_PASS' 20 'stealth direct hostile target suppression'"""
if text.count(old) != 1:
    raise SystemExit(f'[Rogues deep acceptance] expected exactly one original hostile visibility QA block, found {text.count(old)}')
path.write_text(text.replace(old,new,1))
PY
grep -Fq 'on target if entity @s[type=minecraft:player,distance=5..]' "$PLAYER_QA" || { echo '[Rogues deep acceptance] direct control target relation assertion missing' >&2; exit 2; }
grep -Fq 'scoreboard players set stealth_target rogp 0' "$PLAYER_QA" || { echo '[Rogues deep acceptance] direct Stealth target relation sentinel missing' >&2; exit 2; }
grep -Fq 'on target if entity @s[type=minecraft:player] run scoreboard players set stealth_target rogp 1' "$PLAYER_QA" || { echo '[Rogues deep acceptance] direct Stealth target relation probe missing' >&2; exit 2; }
grep -Fq 'stealth_distance rogp matches 1' "$PLAYER_QA" || { echo '[Rogues deep acceptance] Stealth safe-distance assertion missing' >&2; exit 2; }
# Keep the Stealth visibility husk alive long enough to make the hit-break check
# a real entity-caused hit, then resolve the source through `execute as ... by @s`.
# This avoids a source-less damage path and proves RemoveOnHit against gameplay
# semantics while reusing the exact hostile that just exercised visibility.
OLD_DAMAGE="send_cmd 'damage @a[limit=1] 1 minecraft:generic'; sleep 1;"
NEW_DAMAGE="send_cmd 'execute as @e[type=minecraft:husk,tag=rogp_stealth_husk,limit=1,sort=nearest] run damage @p 1 minecraft:generic by @s'; sleep 1; send_cmd 'kill @e[tag=rogp_stealth_husk]';"
python3 - "$PLAYER_QA" "$OLD_DAMAGE" "$NEW_DAMAGE" <<'PY'
from pathlib import Path
import sys
path=Path(sys.argv[1]); old_damage,new_damage=sys.argv[2:]
text=path.read_text()
if text.count(old_damage) != 1:
    raise SystemExit(f'[Rogues deep acceptance] expected exactly one entity-source damage QA seam, found {text.count(old_damage)}')
path.write_text(text.replace(old_damage,new_damage,1))
PY
grep -Fq "$NEW_DAMAGE" "$PLAYER_QA" || { echo '[Rogues deep acceptance] husk entity-source damage QA patch missing' >&2; exit 2; }
bash -n "$PLAYER_QA"
bash "$CI/run-rogues-acceptance.sh"
bash "$CI/run-rogues-release-certification.sh"
bash "$CI/run-rogues-server-behavior.sh"
# The player QA deliberately teleports a real survival client between isolated
# semantic arenas. Vanilla's airborne watchdog can interpret those scripted
# teleports as sustained floating before the client has acknowledged the new
# ground position. This is test-world infrastructure, not a gameplay setting:
# allow flight only on the ephemeral packaged QA server so the native client is
# not disconnected before Rogues' input/effect semantics can be observed.
PLAYER_PROPERTIES="$ROOT/rpg-series-port/rogues-forge-1.20.1/.fresh-rogues-forge-server/server.properties"
[[ -f "$PLAYER_PROPERTIES" ]] || { echo '[Rogues deep acceptance] packaged player-QA server.properties missing' >&2; exit 2; }
if grep -q '^allow-flight=' "$PLAYER_PROPERTIES"; then
  sed -i 's/^allow-flight=.*/allow-flight=true/' "$PLAYER_PROPERTIES"
else
  printf '\nallow-flight=true\n' >> "$PLAYER_PROPERTIES"
fi
grep -Fqx 'allow-flight=true' "$PLAYER_PROPERTIES" || { echo '[Rogues deep acceptance] failed to harden scripted player arena against vanilla floating kick' >&2; exit 2; }
bash "$CI/run-rogues-player-behavior-ci-preflight.sh"
echo '[Rogues deep acceptance] FULL_DEEP_BEHAVIOR_PASS: exact release identity, packaged Forge server, native LWJGL client + real join, current spell/equipment data, Charge, Stealth, Bear Trap, ROOT-vs-SHOCK real-player semantics, and positive controls all passed.'
