#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
echo '[Shield API acceptance] Exact 2.1.0 behavior -> native Forge 1.20.1 compile/package gate'
bash "$ROOT/rpg-series-port/ci/build-shield-api-foundation.sh"
echo '[Shield API acceptance] PASS: exact pins + behavior translation guards + Java17 native Forge package.'
