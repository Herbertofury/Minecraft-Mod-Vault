#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
source "$ROOT/rpg-series-port/ci/run-spell-engine.sh"
bash "$ROOT/rpg-series-port/ci/run-spell-engine-package-smoke.sh" \
  "$OUT_JAR" "$SPELL_POWER_FORGE_JAR" "$RANGED_FORGE_JAR" "$WORK" "$PORT"
echo '[Spell Engine CI] Full modular build/client/fresh-packaged-server verification passed.'
