#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
CI="$ROOT/rpg-series-port/ci"
BASE_RUNNER="$CI/run-wizards.sh"
BASE_ACCEPTANCE="$CI/run-wizards-acceptance.sh"
PATCHED_RUNNER="$CI/.run-wizards-1104.generated.sh"
PATCHED_ACCEPTANCE="$CI/.run-wizards-1104-acceptance.generated.sh"

for f in "$BASE_RUNNER" "$BASE_ACCEPTANCE" "$CI/run-spell-engine-1.10.4-exact-seal-graduation.sh"; do
  test -f "$f"
done

python3 - "$BASE_RUNNER" "$PATCHED_RUNNER" "$BASE_ACCEPTANCE" "$PATCHED_ACCEPTANCE" <<'PY'
from pathlib import Path
import sys
runner_src = Path(sys.argv[1]).read_text()
runner_out = Path(sys.argv[2])
accept_src = Path(sys.argv[3]).read_text()
accept_out = Path(sys.argv[4])

old_builder = 'bash "$ROOT/rpg-series-port/ci/build-spell-engine-foundation.sh"'
new_builder = 'bash "$ROOT/rpg-series-port/ci/run-spell-engine-1.10.4-exact-seal-graduation.sh"'
old_jar = 'spell_engine-forge-1.10.2+1.20.1.jar'
new_jar = 'spell_engine-forge-1.10.4+1.20.1.jar'

if runner_src.count(old_builder) != 1:
    raise SystemExit(f'[Wizards 1.10.4] expected one legacy Spell Engine builder seam, found {runner_src.count(old_builder)}')
if runner_src.count(old_jar) != 1:
    raise SystemExit(f'[Wizards 1.10.4] expected one legacy Spell Engine JAR seam in runner, found {runner_src.count(old_jar)}')
runner_src = runner_src.replace(old_builder, new_builder, 1).replace(old_jar, new_jar, 1)

old_accept_call = 'bash "$ROOT/rpg-series-port/ci/run-wizards.sh"'
new_accept_call = 'bash "$ROOT/rpg-series-port/ci/.run-wizards-1104.generated.sh"'
if accept_src.count(old_accept_call) != 1:
    raise SystemExit(f'[Wizards 1.10.4] expected one acceptance runner seam, found {accept_src.count(old_accept_call)}')
if accept_src.count(old_jar) != 1:
    raise SystemExit(f'[Wizards 1.10.4] expected one legacy Spell Engine JAR seam in acceptance, found {accept_src.count(old_jar)}')
accept_src = accept_src.replace(old_accept_call, new_accept_call, 1).replace(old_jar, new_jar, 1)

for label, text in [('runner', runner_src), ('acceptance', accept_src)]:
    if old_jar in text:
        raise SystemExit(f'[Wizards 1.10.4] stale Spell Engine 1.10.2 JAR remains in generated {label}')
if 'run-spell-engine-1.10.4-exact-seal-graduation.sh' not in runner_src:
    raise SystemExit('[Wizards 1.10.4] generated runner does not consume exact-sealed 1.10.4 producer')

runner_out.write_text(runner_src)
accept_out.write_text(accept_src)
PY

chmod +x "$PATCHED_RUNNER" "$PATCHED_ACCEPTANCE"
bash -n "$PATCHED_RUNNER"
bash -n "$PATCHED_ACCEPTANCE"
! grep -Fq 'spell_engine-forge-1.10.2+1.20.1.jar' "$PATCHED_RUNNER" "$PATCHED_ACCEPTANCE"
grep -Fq 'spell_engine-forge-1.10.4+1.20.1.jar' "$PATCHED_RUNNER"
grep -Fq 'spell_engine-forge-1.10.4+1.20.1.jar' "$PATCHED_ACCEPTANCE"
grep -Fq 'run-spell-engine-1.10.4-exact-seal-graduation.sh' "$PATCHED_RUNNER"
echo '[Wizards 3.1.1] SPELL_ENGINE_1104_GENERATED_ACCEPTANCE_READY'
exec bash "$PATCHED_ACCEPTANCE"
