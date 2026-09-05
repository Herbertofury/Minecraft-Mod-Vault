#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
BASE="$ROOT/rpg-series-port/ci/run-wizards-spell-engine-1.10.4-gameplay.sh"
PATCHED="$ROOT/rpg-series-port/ci/.run-wizards-spell-engine-1.10.4-gameplay-warning-safe.generated.sh"

test -f "$BASE"

python3 - "$BASE" "$PATCHED" <<'PY'
from pathlib import Path
import sys
src = Path(sys.argv[1]).read_text()
out = Path(sys.argv[2])
old = '''printf 'earlyWindowControl = false\\n' > "$RUN/config/fml.toml"\nif [[ -f "$RUN/options.txt" ]] && grep -q '^onboardAccessibility:' "$RUN/options.txt"; then\n'''
new = '''printf 'earlyWindowControl = false\\n' > "$RUN/config/fml.toml"\ncat > "$RUN/config/forge-client.toml" <<'FORGECLIENT'\n[client]\nshowLoadWarnings = false\nFORGECLIENT\necho '[Wizards QA] FORGE_INTERACTIVE_WARNING_SCREEN_DISABLED_FOR_QA showLoadWarnings=false real_loading_errors_still_fatal=true'\nif [[ -f "$RUN/options.txt" ]] && grep -q '^onboardAccessibility:' "$RUN/options.txt"; then\n'''
if src.count(old) != 1:
    raise SystemExit(f'[Wizards QA] expected one mapped-client config seam, found {src.count(old)}')
if 'showLoadWarnings = false' in src:
    raise SystemExit('[Wizards QA] warning-screen patch unexpectedly already present in base gameplay runner')
patched = src.replace(old, new, 1)
for required in (
    'forge-client.toml',
    '[client]',
    'showLoadWarnings = false',
    'real_loading_errors_still_fatal=true',
):
    if patched.count(required) != 1:
        raise SystemExit(f'[Wizards QA] warning-screen contract failed for {required!r}: {patched.count(required)}')
out.write_text(patched)
PY

chmod +x "$PATCHED"
bash -n "$PATCHED"
grep -Fq 'showLoadWarnings = false' "$PATCHED"
grep -Fq 'real_loading_errors_still_fatal=true' "$PATCHED"
echo '[Wizards QA] WARNING_SAFE_GAMEPLAY_RUNNER_READY'
exec bash "$PATCHED"
