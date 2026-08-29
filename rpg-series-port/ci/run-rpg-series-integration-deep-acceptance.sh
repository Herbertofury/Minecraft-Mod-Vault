#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
CI="$ROOT/rpg-series-port/ci"
TARGET="$CI/run-rpg-series-integration-acceptance.sh"
OLD='  local marker="$1" timeout_seconds="$2" label="$3" deadline=$((SECONDS+timeout_seconds))'
NEW='  local marker="$1" timeout_seconds="$2" label="$3"'
PAL_CERT_OLD=$'bash "$CI/run-paladins-acceptance.sh"\nbash "$CI/run-paladins-release-certification.sh"'
PAL_CERT_NEW=$'bash "$CI/run-paladins-acceptance.sh"\nPAL_CONTRACT_REL="rpg-series-port/paladins-forge-1.20.1/PORT_CONTRACT.md"\nPAL_CONTRACT="$ROOT/$PAL_CONTRACT_REL"\nPAL_CONTRACT_SAVED="${RUNNER_TEMP:-/tmp}/paladins-port-contract-current.$$"\ncp -f "$PAL_CONTRACT" "$PAL_CONTRACT_SAVED"\nPAL_CONTRACT_CURRENT_BLOB="$(git -C "$ROOT" hash-object "$PAL_CONTRACT")"\ngit -C "$ROOT" fetch --no-tags --depth=1 origin 257171c71a363285bb5bdbb58083121f3a7456d3\ngit -C "$ROOT" show 257171c71a363285bb5bdbb58083121f3a7456d3:"$PAL_CONTRACT_REL" > "$PAL_CONTRACT"\n[[ "$(git -C "$ROOT" hash-object "$PAL_CONTRACT")" = "09b523c7d4688961ec992f77dc08835b12a29bb2" ]] || { echo "[RPG integration] frozen Paladins source-authority contract blob mismatch" >&2; cp -f "$PAL_CONTRACT_SAVED" "$PAL_CONTRACT"; rm -f "$PAL_CONTRACT_SAVED"; exit 1; }\necho "[RPG integration] PALADINS_FROZEN_SOURCE_VIEW_PASS authority=257171c71a363285bb5bdbb58083121f3a7456d3 contract=09b523c7d4688961ec992f77dc08835b12a29bb2"\nPAL_CERT_STATUS=0\nbash "$CI/run-paladins-release-certification.sh" || PAL_CERT_STATUS=$?\ncp -f "$PAL_CONTRACT_SAVED" "$PAL_CONTRACT"\nrm -f "$PAL_CONTRACT_SAVED"\n[[ "$(git -C "$ROOT" hash-object "$PAL_CONTRACT")" = "$PAL_CONTRACT_CURRENT_BLOB" ]] || { echo "[RPG integration] current Paladins contract restoration failed" >&2; exit 1; }\necho "[RPG integration] PALADINS_CURRENT_CONTRACT_RESTORE_PASS blob=$PAL_CONTRACT_CURRENT_BLOB"\n(( PAL_CERT_STATUS == 0 )) || exit "$PAL_CERT_STATUS"'
[[ -f "$TARGET" ]] || { echo '[RPG integration deep] integration acceptance runner missing' >&2; exit 2; }
COUNT="$(grep -Fxc "$OLD" "$TARGET" || true)"
[[ "$COUNT" = 1 ]] || { echo "[RPG integration deep] expected exactly one compound-local wait_marker seam, found $COUNT" >&2; exit 2; }
python3 - "$TARGET" "$OLD" "$NEW" "$PAL_CERT_OLD" "$PAL_CERT_NEW" <<'PY'
from pathlib import Path
import sys
path=Path(sys.argv[1]); old=sys.argv[2]; new=sys.argv[3]; pal_old=sys.argv[4]; pal_new=sys.argv[5]
text=path.read_text()
replacement=new+'\n  local deadline=$((SECONDS+timeout_seconds))'
if text.count(old) != 1:
    raise SystemExit(f'[RPG integration deep] expected one wait_marker seam, found {text.count(old)}')
if text.count(pal_old) != 1:
    raise SystemExit(f'[RPG integration deep] expected one Paladins certification seam, found {text.count(pal_old)}')
text=text.replace(old,replacement,1)
text=text.replace(pal_old,pal_new,1)
path.write_text(text)
PY
grep -Fq '  local deadline=$((SECONDS+timeout_seconds))' "$TARGET" || { echo '[RPG integration deep] split deadline declaration missing after patch' >&2; exit 2; }
if grep -Fq "$OLD" "$TARGET"; then echo '[RPG integration deep] unsafe compound local declaration survived patch' >&2; exit 2; fi
grep -Fq 'PALADINS_FROZEN_SOURCE_VIEW_PASS' "$TARGET" || { echo '[RPG integration deep] frozen Paladins source-view seam missing after patch' >&2; exit 2; }
grep -Fq '09b523c7d4688961ec992f77dc08835b12a29bb2' "$TARGET" || { echo '[RPG integration deep] frozen Paladins contract blob pin missing after patch' >&2; exit 2; }
bash -n "$TARGET"
echo '[RPG integration deep] WAIT_MARKER_HARNESS_HARDENING_PASS'
echo '[RPG integration deep] PALADINS_SOURCE_VIEW_HARDENING_PASS'
bash "$TARGET"
