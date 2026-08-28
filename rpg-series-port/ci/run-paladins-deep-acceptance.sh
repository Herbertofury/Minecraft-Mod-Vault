#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
bash "$ROOT/rpg-series-port/ci/run-paladins-acceptance.sh"
bash "$ROOT/rpg-series-port/ci/run-paladins-release-certification.sh"
bash "$ROOT/rpg-series-port/ci/run-paladins-server-behavior-v2.sh"
bash "$ROOT/rpg-series-port/ci/run-paladins-player-behavior-acceptance.sh"
echo '[Paladins deep acceptance] FULL_DEEP_BEHAVIOR_PASS: baseline native runtime matrix, cross-run exact release/source identity, certified packaged-server Priest/Levitate semantics, and exact packaged-server + native-client real-player Divine Protection + Judgement/STUN controls all passed.'
