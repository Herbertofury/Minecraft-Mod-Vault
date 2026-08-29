#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
bash "$ROOT/rpg-series-port/ci/run-rogues-acceptance.sh"
bash "$ROOT/rpg-series-port/ci/run-rogues-release-certification.sh"
bash "$ROOT/rpg-series-port/ci/run-rogues-server-behavior.sh"
bash "$ROOT/rpg-series-port/ci/run-rogues-player-behavior-ci-preflight.sh"
echo '[Rogues deep acceptance] FULL_DEEP_BEHAVIOR_PASS: exact release identity, packaged Forge server, native LWJGL client + real join, current spell/equipment data, Charge, Stealth, Bear Trap, ROOT-vs-SHOCK real-player semantics, and positive controls all passed.'
