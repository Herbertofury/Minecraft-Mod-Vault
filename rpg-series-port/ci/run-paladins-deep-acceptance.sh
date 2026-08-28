#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
bash "$ROOT/rpg-series-port/ci/run-paladins-acceptance.sh"
bash "$ROOT/rpg-series-port/ci/run-paladins-behavior-acceptance.sh"
bash "$ROOT/rpg-series-port/ci/run-paladins-judgement-player-acceptance.sh"
echo '[Paladins deep acceptance] FULL_DEEP_BEHAVIOR_PASS: baseline runtime matrix, packaged-server effect semantics, and real integrated-player Judgement/STUN controls all passed.'
