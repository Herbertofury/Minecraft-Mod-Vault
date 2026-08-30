#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
PREP="$ROOT/rpg-series-port/gazebos-forge-1.20.1/tools/prepare_port.py"
V2="$ROOT/rpg-series-port/ci/run-gazebos-acceptance-v2.sh"

test -f "$PREP"
test -f "$V2"

# QA-owned fix for the preparer's modern Repurposed Structures inventory guard.
# Upstream rs_pool_additions are nested under villages/<biome>/houses.json, so
# a root-only glob falsely reports zero even though copytree preserved all files.
python3 - "$PREP" <<'PY'
from pathlib import Path
import sys
p=Path(sys.argv[1])
s=p.read_text()
old="rs_additions = list((out / 'common/src/main/resources/data/gazebo/rs_pool_additions').glob('*.json'))"
new="rs_additions = list((out / 'common/src/main/resources/data/gazebo/rs_pool_additions').rglob('*.json'))"
if s.count(old) != 1:
    raise SystemExit(f'[Gazebos v3] expected exactly one non-recursive rs_pool_additions guard, found {s.count(old)}')
p.write_text(s.replace(old,new))
PY

python3 -m py_compile "$PREP"
echo '[Gazebos v3] NESTED_RS_POOL_ADDITIONS_GUARD_PASS'
exec bash "$V2"
