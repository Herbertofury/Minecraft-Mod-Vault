#!/usr/bin/env bash
set -euo pipefail
bash rpg-series-port/ci/run-jewelry.sh
bash rpg-series-port/ci/run-jewelry-client-smoke.sh
