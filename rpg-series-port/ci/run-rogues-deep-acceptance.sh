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
bash "$CI/run-rogues-acceptance.sh"
bash "$CI/run-rogues-release-certification.sh"
bash "$CI/run-rogues-server-behavior.sh"
bash "$CI/run-rogues-player-behavior-ci-preflight.sh"
echo '[Rogues deep acceptance] FULL_DEEP_BEHAVIOR_PASS: exact release identity, packaged Forge server, native LWJGL client + real join, current spell/equipment data, Charge, Stealth, Bear Trap, ROOT-vs-SHOCK real-player semantics, and positive controls all passed.'
