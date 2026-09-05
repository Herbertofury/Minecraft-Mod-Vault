#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
PREP="$ROOT/rpg-series-port/gazebos-forge-1.20.1/tools/prepare_port.py"
V2="$ROOT/rpg-series-port/ci/run-gazebos-acceptance-v2.sh"

test -f "$PREP"
test -f "$V2"

python3 - "$PREP" <<'PY'
from pathlib import Path
import sys
s=Path(sys.argv[1]).read_text()
expected="rs_additions = list((out / 'common/src/main/resources/data/gazebo/rs_pool_additions').rglob('*.json'))"
old="rs_additions = list((out / 'common/src/main/resources/data/gazebo/rs_pool_additions').glob('*.json'))"
if s.count(expected) != 1:
    raise SystemExit(f'[Gazebos v3] expected exactly one persistent recursive rs_pool_additions guard, found {s.count(expected)}')
if old in s:
    raise SystemExit('[Gazebos v3] obsolete non-recursive rs_pool_additions guard still present')
PY

python3 -m py_compile "$PREP"
echo '[Gazebos v3] PERSISTENT_NESTED_RS_POOL_ADDITIONS_GUARD_PASS'
exec bash "$V2"
