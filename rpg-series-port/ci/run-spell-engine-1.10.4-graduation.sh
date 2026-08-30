#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
BASE="$ROOT/rpg-series-port/ci/run-spell-engine.sh"
GRAD="$ROOT/rpg-series-port/ci/run-spell-engine-graduation.sh"
ENV_FILE="$ROOT/rpg-series-port/spell-engine-forge-1.20.1/SPELL_ENGINE_GRADUATION.env"

test -f "$BASE"
test -f "$GRAD"
test -f "$ENV_FILE"
source "$ENV_FILE"

[[ "$SPELL_ENGINE_TARGET_VERSION" = '1.10.4' ]]
[[ "$SPELL_ENGINE_1104_TARGET_SHA" = '8843cea8974afffc7ec42c096cac33327a3af3d8' ]]

# Uplift only target authority/path/release labeling in this CI workspace and repair the
# historical Spell Power dependency-build invocation. The historical source-package seam
# intentionally remains untouched here because the audited graduation harness owns its
# deterministic replacement. This ordering prevents the two fail-closed wrappers from
# competing over the same literal.
python3 - "$BASE" <<'PY'
from pathlib import Path
import sys
p=Path(sys.argv[1]); s=p.read_text()

repls=[
    ('TARGET_SHA=bc02f7a49da950503010020da491f6bdc5871df7',
     'TARGET_SHA=8843cea8974afffc7ec42c096cac33327a3af3d8', 1),
    ('spell-engine-1102', 'spell-engine-1104', 2),
    ('spell_engine-forge-1.10.2+1.20.1.jar',
     'spell_engine-forge-1.10.4+1.20.1.jar', 1),
    ('gradle --no-daemon --stacktrace -p "$SPELL_POWER" :forge:build',
     'gradle --no-daemon --stacktrace -p "$SPELL_POWER" -Ptiny_config_common_jar="$TINY_CONFIG_COMMON_JAR" -Ptiny_config_forge_jar="$TINY_CONFIG_FORGE_JAR" :forge:build', 1),
]
for old,new,expected in repls:
    actual=s.count(old)
    if actual != expected:
        raise SystemExit(f'[Spell Engine 1.10.4] expected {expected} occurrences of {old!r}, found {actual}')
    s=s.replace(old,new)

# Graduation harness must still see its historical source-package seam exactly three times.
source_name='spell-engine-1.10.2-forge-1.20.1-source-ci.zip'
if s.count(source_name) != 3:
    raise SystemExit(f'[Spell Engine 1.10.4] graduation-owned source seam drifted: expected 3, found {s.count(source_name)}')
if 'bc02f7a49da950503010020da491f6bdc5871df7' in s:
    raise SystemExit('[Spell Engine 1.10.4] stale 1.10.2 target SHA remains in active base runner')
if 'spell-engine-1102' in s:
    raise SystemExit('[Spell Engine 1.10.4] stale 1.10.2 target path remains in active base runner')
if 'gradle --no-daemon --stacktrace -p "$SPELL_POWER" :forge:build' in s:
    raise SystemExit('[Spell Engine 1.10.4] stale Spell Power build invocation still omits TinyConfig 3.1 inputs')
p.write_text(s)
PY

bash -n "$BASE"
echo '[Spell Engine 1.10.4] TARGET_AUTHORITY_AND_SPELL_POWER_INPUTS_PASS sha=8843cea8974afffc7ec42c096cac33327a3af3d8'
exec bash "$GRAD"
