#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
CI="$ROOT/rpg-series-port/ci"
TARGET="$CI/run-rpg-series-integration-acceptance.sh"
OLD='  local marker="$1" timeout_seconds="$2" label="$3" deadline=$((SECONDS+timeout_seconds))'
NEW='  local marker="$1" timeout_seconds="$2" label="$3"'
[[ -f "$TARGET" ]] || { echo '[RPG integration deep] integration acceptance runner missing' >&2; exit 2; }
COUNT="$(grep -Fxc "$OLD" "$TARGET" || true)"
[[ "$COUNT" = 1 ]] || { echo "[RPG integration deep] expected exactly one compound-local wait_marker seam, found $COUNT" >&2; exit 2; }
python3 - "$TARGET" "$OLD" "$NEW" <<'PY'
from pathlib import Path
import sys
path=Path(sys.argv[1]); old=sys.argv[2]; new=sys.argv[3]
text=path.read_text()
replacement=new+'\n  local deadline=$((SECONDS+timeout_seconds))'
if text.count(old) != 1:
    raise SystemExit(f'[RPG integration deep] expected one wait_marker seam, found {text.count(old)}')
path.write_text(text.replace(old,replacement,1))
PY
grep -Fq '  local deadline=$((SECONDS+timeout_seconds))' "$TARGET" || { echo '[RPG integration deep] split deadline declaration missing after patch' >&2; exit 2; }
if grep -Fq "$OLD" "$TARGET"; then echo '[RPG integration deep] unsafe compound local declaration survived patch' >&2; exit 2; fi
bash -n "$TARGET"
echo '[RPG integration deep] WAIT_MARKER_HARNESS_HARDENING_PASS'
bash "$TARGET"
