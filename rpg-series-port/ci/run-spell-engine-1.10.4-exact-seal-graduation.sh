#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
INNER="$ROOT/rpg-series-port/ci/run-spell-engine-1.10.4-graduation.sh"
GRAD="$ROOT/rpg-series-port/ci/run-spell-engine-graduation.sh"
PATCHER="$ROOT/rpg-series-port/spell-engine-forge-1.20.1/tools/patch_spell_engine_1104_generated_runner.py"
SEALER="$ROOT/rpg-series-port/spell-engine-forge-1.20.1/tools/seal_certified_tinyconfig_nested.py"

test -f "$INNER"
test -f "$GRAD"
test -f "$PATCHER"
test -f "$SEALER"

# Keep the previously-audited generic graduation generator intact in Git history. In this exact CI
# workspace, add one 1.10.4-specific generated-runner hardening hook before it executes the generated
# script. The inner uplift still owns authority/port transforms; this outer layer owns only final
# distributable sealing and runtime proof of the 1.10.3 tooltip key registration.
python3 - "$GRAD" <<'PY'
from pathlib import Path
import sys
p = Path(sys.argv[1])
s = p.read_text()
old = '''chmod +x "$PATCHED"\nbash -n "$PATCHED"\necho '[Spell Engine graduation] TINY_CONFIG_310_PACKAGED_SERVER_AND_FREEZE_WRAPPER_READY'\nexec bash "$PATCHED"\n'''
new = '''python3 "$ROOT/rpg-series-port/spell-engine-forge-1.20.1/tools/patch_spell_engine_1104_generated_runner.py" "$PATCHED"\nchmod +x "$PATCHED"\nbash -n "$PATCHED"\necho '[Spell Engine graduation] TINY_CONFIG_310_PACKAGED_SERVER_EXACT_SEAL_TOOLTIP_RUNTIME_AND_FREEZE_WRAPPER_READY'\nexec bash "$PATCHED"\n'''
if s.count(old) != 1:
    raise SystemExit(f'[Spell Engine 1.10.4] expected one graduation generated-runner execution seam, found {s.count(old)}')
p.write_text(s.replace(old, new, 1))
PY

bash -n "$INNER"
python3 -m py_compile "$PATCHER" "$SEALER"
echo '[Spell Engine 1.10.4] EXACT_CERTIFIED_NESTED_BYTES_AND_NATIVE_TOOLTIP_RUNTIME_WRAPPER_READY'
exec bash "$INNER"
