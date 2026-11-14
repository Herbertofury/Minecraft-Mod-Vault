#!/usr/bin/env bash
set -euo pipefail
CI="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
bash "$CI/run-more-rpg-library-2.7.2-port.sh"
bash "$CI/verify-more-rpg-client-lifecycle.sh"
echo '[More RPG 2.7.2] ACTIVE_FINAL_GRADE_PACKAGE_PASS client_lifecycle=true'
