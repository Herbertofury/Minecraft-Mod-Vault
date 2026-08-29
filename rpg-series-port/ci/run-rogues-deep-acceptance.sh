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
# Keep the real-player Stealth visibility fixture tied to the player's actual
# authoritative server position. A real LWJGL client can briefly reconcile a
# scripted teleport after the server accepts it; fixed absolute skeleton
# coordinates can therefore stop representing six blocks even though the
# gameplay code is correct. Require the player to settle inside each intended
# arena, then summon the hostile exactly six blocks from the player's current
# server-side position. This preserves the semantic distance and fails closed if
# either the fixture shape or server position drifts.
PLAYER_QA="$CI/run-rogues-player-behavior-acceptance.sh"
OLD_CONTROL="send_cmd 'tp @a 20.5 100 0.5 0 0'; send_cmd 'effect give @a minecraft:instant_health 1 10 true'; send_cmd 'summon minecraft:skeleton 20.5 100 6.5 {Tags:[\"rogp\",\"rogp_control_skeleton\"],Silent:1b,PersistenceRequired:1b}';"
NEW_CONTROL="send_cmd 'tp @a 20.5 100 0.5 0 0'; sleep 2; send_cmd 'execute positioned 20.5 100 0.5 if entity @a[distance=..1.0,limit=1] run say ROGUES_CONTROL_ARENA_POSITION_READY'; wait_marker 'ROGUES_CONTROL_ARENA_POSITION_READY' 20 'control arena position'; send_cmd 'effect give @a minecraft:instant_health 1 10 true'; send_cmd 'execute at @a[limit=1] run summon minecraft:skeleton ~ ~ ~6 {Tags:[\"rogp\",\"rogp_control_skeleton\"],Silent:1b,PersistenceRequired:1b}';"
OLD_STEALTH="send_cmd 'tp @a 30.5 100 0.5 0 0'; send_cmd 'effect give @a rogues:stealth 30 0 true'; send_cmd 'summon minecraft:skeleton 30.5 100 6.5 {Tags:[\"rogp\",\"rogp_stealth_skeleton\"],Silent:1b,PersistenceRequired:1b}';"
NEW_STEALTH="send_cmd 'tp @a 30.5 100 0.5 0 0'; sleep 2; send_cmd 'execute positioned 30.5 100 0.5 if entity @a[distance=..1.0,limit=1] run say ROGUES_STEALTH_ARENA_POSITION_READY'; wait_marker 'ROGUES_STEALTH_ARENA_POSITION_READY' 20 'Stealth arena position'; send_cmd 'effect give @a rogues:stealth 30 0 true'; send_cmd 'execute at @a[limit=1] run summon minecraft:skeleton ~ ~ ~6 {Tags:[\"rogp\",\"rogp_stealth_skeleton\"],Silent:1b,PersistenceRequired:1b}';"
python3 - "$PLAYER_QA" "$OLD_CONTROL" "$NEW_CONTROL" "$OLD_STEALTH" "$NEW_STEALTH" <<'PY'
from pathlib import Path
import sys
path=Path(sys.argv[1]); old_control,new_control,old_stealth,new_stealth=sys.argv[2:]
text=path.read_text()
for label, old in (("control six-block arena", old_control), ("Stealth six-block arena", old_stealth)):
    if text.count(old) != 1:
        raise SystemExit(f'[Rogues deep acceptance] expected exactly one {label} QA seam, found {text.count(old)}')
text=text.replace(old_control,new_control,1).replace(old_stealth,new_stealth,1)
path.write_text(text)
PY
grep -Fq 'ROGUES_CONTROL_ARENA_POSITION_READY' "$PLAYER_QA" || { echo '[Rogues deep acceptance] control arena stabilization patch missing' >&2; exit 2; }
grep -Fq 'ROGUES_STEALTH_ARENA_POSITION_READY' "$PLAYER_QA" || { echo '[Rogues deep acceptance] Stealth arena stabilization patch missing' >&2; exit 2; }
grep -Fq 'execute at @a[limit=1] run summon minecraft:skeleton ~ ~ ~6' "$PLAYER_QA" || { echo '[Rogues deep acceptance] relative six-block summon patch missing' >&2; exit 2; }
# The Stealth visibility arena owns exactly one tagged skeleton. Keep that entity
# alive long enough to make the hit-break check a real entity-caused hit, then
# resolve the source through `execute as ... by @s`. This avoids an ambiguous
# source-less damage path and proves RemoveOnHit against gameplay semantics.
OLD_KILL="wait_marker 'ROGUES_STEALTH_VISIBILITY_RUNTIME_PASS' 20 'stealth hostile visibility reduction'; send_cmd 'kill @e[tag=rogp_stealth_skeleton]'"
NEW_KILL="wait_marker 'ROGUES_STEALTH_VISIBILITY_RUNTIME_PASS' 20 'stealth hostile visibility reduction'"
OLD_DAMAGE="send_cmd 'damage @a[limit=1] 1 minecraft:generic'; sleep 1;"
NEW_DAMAGE="send_cmd 'execute as @e[type=minecraft:skeleton,tag=rogp_stealth_skeleton,limit=1,sort=nearest] run damage @p 1 minecraft:generic by @s'; sleep 1; send_cmd 'kill @e[tag=rogp_stealth_skeleton]';"
python3 - "$PLAYER_QA" "$OLD_KILL" "$NEW_KILL" "$OLD_DAMAGE" "$NEW_DAMAGE" <<'PY'
from pathlib import Path
import sys
path=Path(sys.argv[1]); old_kill,new_kill,old_damage,new_damage=sys.argv[2:]
text=path.read_text()
for label, old in (("stealth skeleton lifetime", old_kill), ("entity-source damage", old_damage)):
    if text.count(old) != 1:
        raise SystemExit(f'[Rogues deep acceptance] expected exactly one {label} QA seam, found {text.count(old)}')
text=text.replace(old_kill,new_kill,1).replace(old_damage,new_damage,1)
path.write_text(text)
PY
grep -Fq "$NEW_DAMAGE" "$PLAYER_QA" || { echo '[Rogues deep acceptance] entity-source damage QA patch missing' >&2; exit 2; }
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
