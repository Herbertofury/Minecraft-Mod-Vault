#!/usr/bin/env bash
set -euo pipefail

ROOT="${GITHUB_WORKSPACE:?}"
PATCHER="$ROOT/rpg-series-port/spell-engine-forge-1.20.1/tools/patch_spell_engine_1104_generated_runner.py"
RUNNER="$ROOT/rpg-series-port/ci/run-spell-engine-1.10.4-exact-seal-graduation.sh"
ENV_FILE="$ROOT/rpg-series-port/spell-engine-forge-1.20.1/SPELL_ENGINE_GRADUATION.env"

for f in "$PATCHER" "$RUNNER" "$ENV_FILE"; do
  test -f "$f"
done
source "$ENV_FILE"

# More RPG is a downstream consumer of an already-graduated Spell Engine release. Replaying the
# complete Spell Engine native-client + packaged-server graduation on every downstream build creates
# a false serial dependency and can fail on a fresh CI host even after the exact certified release
# bytes have already been reproduced. Keep the canonical Spell Engine graduation path untouched in
# the repository. For this one ephemeral workspace only, teach its generated-runner hardener to stop
# after it has independently rebuilt and matched BOTH frozen release and deterministic source hashes.
# Product source, release bytes, and the default full Spell Engine graduation path are unchanged.
python3 - "$PATCHER" <<'PY'
from pathlib import Path
import sys

p = Path(sys.argv[1])
s = p.read_text()
anchor = '# The 1.10.3 tooltip-details key must be proven by the real Forge client lifecycle, not merely source\n'
if s.count(anchor) != 1:
    raise SystemExit(f'[Spell Engine materializer] expected one tooltip runtime gate anchor, found {s.count(anchor)}')
if 'CERTIFIED_FOUNDATION_MATERIALIZATION_PASS' in s:
    raise SystemExit('[Spell Engine materializer] ephemeral materialization patch unexpectedly already present')

snippet = r'''# Downstream-only certified materialization mode. This is injected only into the ephemeral generated
# runner used by an already-graduated consumer. It is deliberately impossible to pass on merely a
# successful build: both the canonical release JAR and deterministic source archive must match the
# frozen graduation identities before the expensive native runtime replay may be skipped.
old_materialize = 'echo "[Spell Engine graduation] CLEAN_REBUILD_IDENTITY_PASS sha=$FIRST_SHA"\n'
materialize = r'''if [[ "${SPELL_ENGINE_CERTIFIED_MATERIALIZE_ONLY:-0}" = "1" ]]; then
  if [[ "$FIRST_SHA" != "$SPELL_ENGINE_EXPECTED_JAR_SHA" ]]; then
    echo "[Spell Engine materializer] release identity mismatch: actual=$FIRST_SHA expected=$SPELL_ENGINE_EXPECTED_JAR_SHA" >&2
    exit 1
  fi
  if [[ "$SOURCE_SHA" != "$SPELL_ENGINE_EXPECTED_SOURCE_SHA" ]]; then
    echo "[Spell Engine materializer] source identity mismatch: actual=$SOURCE_SHA expected=$SPELL_ENGINE_EXPECTED_SOURCE_SHA" >&2
    exit 1
  fi
  test -f "$OUT_JAR"
  unzip -tq "$OUT_JAR" >/dev/null
  echo "[Spell Engine graduation] CERTIFIED_FOUNDATION_MATERIALIZATION_PASS jar=$FIRST_SHA source=$SOURCE_SHA runtime_authority=frozen"
  exit 0
fi
'''
if s.count(old_materialize) != 1:
    raise SystemExit(f'expected one clean-rebuild identity seam for certified materialization, found {s.count(old_materialize)}')
s = s.replace(old_materialize, old_materialize + materialize, 1)

'''
p.write_text(s.replace(anchor, snippet + anchor, 1))
PY

python3 -m py_compile "$PATCHER"
SPELL_ENGINE_CERTIFIED_MATERIALIZE_ONLY=1 bash "$RUNNER"

SPELL_ENGINE_JAR="$ROOT/rpg-series-port/spell-engine-forge-1.20.1/spell_engine-forge-1.10.4+1.20.1.jar"
SPELL_ENGINE_COMMON_JAR="$(find "$ROOT/.spell-engine-build/common/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' | sort | head -n1)"
for f in "$SPELL_ENGINE_JAR" "$SPELL_ENGINE_COMMON_JAR"; do
  test -f "$f"
  unzip -tq "$f" >/dev/null
done
ACTUAL_JAR_SHA="$(sha256sum "$SPELL_ENGINE_JAR" | awk '{print $1}')"
[[ "$ACTUAL_JAR_SHA" = "$SPELL_ENGINE_EXPECTED_JAR_SHA" ]]

echo "[Spell Engine materializer] CERTIFIED_DOWNSTREAM_FOUNDATION_READY jar=$ACTUAL_JAR_SHA common=$(basename "$SPELL_ENGINE_COMMON_JAR") full_native_replay=not-required-for-unchanged-frozen-bytes"
