#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
bash "$ROOT/rpg-series-port/ci/run-paladins-acceptance.sh"
bash "$ROOT/rpg-series-port/ci/run-paladins-behavior-acceptance.sh"
echo '[Paladins deep acceptance] BASELINE_PLUS_SERVER_BEHAVIOR_PASS; integrated-player Judgement STUN remains the final behavior-specific gate before graduation.'
