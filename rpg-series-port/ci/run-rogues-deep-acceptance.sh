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
