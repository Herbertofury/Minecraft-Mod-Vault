#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
CI="$ROOT/rpg-series-port/ci"
PORT="$ROOT/rpg-series-port/bundle-api-forge-1.20.1"

bash "$CI/run-bundle-api-acceptance.sh"

BUNDLE_JAR="$(find "$PORT/forge/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' ! -name '*dev-shadow*' ! -name '*javadoc*' | sort | head -n1)"
[[ -n "$BUNDLE_JAR" && -f "$BUNDLE_JAR" ]]

bash "$CI/run-bundle-api-archers-consumer.sh" "$BUNDLE_JAR"

echo '[Bundle API] Standalone strong-runtime + real Archers 3.1.1 quiver-consumer boundary passed.'
