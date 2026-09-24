#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
CI="$ROOT/rpg-series-port/ci"
TARGET="$CI/run-rpg-series-integration-acceptance.sh"
OLD='  local marker="$1" timeout_seconds="$2" label="$3" deadline=$((SECONDS+timeout_seconds))'
NEW='  local marker="$1" timeout_seconds="$2" label="$3"'
ROG_CERT_OLD=$'bash "$CI/run-rogues-acceptance.sh"\nbash "$CI/run-rogues-release-certification.sh"'
ROG_CERT_NEW=$'ROG_CONTRACT_REL="rpg-series-port/rogues-forge-1.20.1/PORT_CONTRACT.md"\nROG_CONTRACT="$ROOT/$ROG_CONTRACT_REL"\nROG_CONTRACT_SAVED="${RUNNER_TEMP:-/tmp}/rogues-port-contract-current.$$"\ncp -f "$ROG_CONTRACT" "$ROG_CONTRACT_SAVED"\nROG_CONTRACT_CURRENT_BLOB="$(git -C "$ROOT" hash-object "$ROG_CONTRACT")"\ngit -C "$ROOT" fetch --no-tags --depth=1 origin 77af5b416ab0889e770028ad3b93fbf2034d311c\ngit -C "$ROOT" show 77af5b416ab0889e770028ad3b93fbf2034d311c:"$ROG_CONTRACT_REL" > "$ROG_CONTRACT"\n[[ "$(git -C "$ROOT" hash-object "$ROG_CONTRACT")" = "49929e20053d91300d7551c567efc97e93a9d328" ]] || { echo "[RPG integration] frozen Rogues source-authority contract blob mismatch" >&2; cp -f "$ROG_CONTRACT_SAVED" "$ROG_CONTRACT"; rm -f "$ROG_CONTRACT_SAVED"; exit 1; }\necho "[RPG integration] ROGUES_FROZEN_SOURCE_VIEW_PASS authority=77af5b416ab0889e770028ad3b93fbf2034d311c contract=49929e20053d91300d7551c567efc97e93a9d328"\nROG_CERT_STATUS=0\nbash "$CI/run-rogues-acceptance.sh" || ROG_CERT_STATUS=$?\nif (( ROG_CERT_STATUS == 0 )); then bash "$CI/run-rogues-release-certification.sh" || ROG_CERT_STATUS=$?; fi\ncp -f "$ROG_CONTRACT_SAVED" "$ROG_CONTRACT"\nrm -f "$ROG_CONTRACT_SAVED"\n[[ "$(git -C "$ROOT" hash-object "$ROG_CONTRACT")" = "$ROG_CONTRACT_CURRENT_BLOB" ]] || { echo "[RPG integration] current Rogues contract restoration failed" >&2; exit 1; }\necho "[RPG integration] ROGUES_CURRENT_CONTRACT_RESTORE_PASS blob=$ROG_CONTRACT_CURRENT_BLOB"\n(( ROG_CERT_STATUS == 0 )) || exit "$ROG_CERT_STATUS"'
PAL_CERT_OLD=$'bash "$CI/run-paladins-acceptance.sh"\nbash "$CI/run-paladins-release-certification.sh"'
PAL_CERT_NEW=$'bash "$CI/run-paladins-acceptance.sh"\nPAL_CONTRACT_REL="rpg-series-port/paladins-forge-1.20.1/PORT_CONTRACT.md"\nPAL_CONTRACT="$ROOT/$PAL_CONTRACT_REL"\nPAL_CONTRACT_SAVED="${RUNNER_TEMP:-/tmp}/paladins-port-contract-current.$$"\ncp -f "$PAL_CONTRACT" "$PAL_CONTRACT_SAVED"\nPAL_CONTRACT_CURRENT_BLOB="$(git -C "$ROOT" hash-object "$PAL_CONTRACT")"\ngit -C "$ROOT" fetch --no-tags --depth=1 origin 257171c71a363285bb5bdbb58083121f3a7456d3\ngit -C "$ROOT" show 257171c71a363285bb5bdbb58083121f3a7456d3:"$PAL_CONTRACT_REL" > "$PAL_CONTRACT"\n[[ "$(git -C "$ROOT" hash-object "$PAL_CONTRACT")" = "09b523c7d4688961ec992f77dc08835b12a29bb2" ]] || { echo "[RPG integration] frozen Paladins source-authority contract blob mismatch" >&2; cp -f "$PAL_CONTRACT_SAVED" "$PAL_CONTRACT"; rm -f "$PAL_CONTRACT_SAVED"; exit 1; }\necho "[RPG integration] PALADINS_FROZEN_SOURCE_VIEW_PASS authority=257171c71a363285bb5bdbb58083121f3a7456d3 contract=09b523c7d4688961ec992f77dc08835b12a29bb2"\nPAL_CERT_STATUS=0\nbash "$CI/run-paladins-release-certification.sh" || PAL_CERT_STATUS=$?\ncp -f "$PAL_CONTRACT_SAVED" "$PAL_CONTRACT"\nrm -f "$PAL_CONTRACT_SAVED"\n[[ "$(git -C "$ROOT" hash-object "$PAL_CONTRACT")" = "$PAL_CONTRACT_CURRENT_BLOB" ]] || { echo "[RPG integration] current Paladins contract restoration failed" >&2; exit 1; }\necho "[RPG integration] PALADINS_CURRENT_CONTRACT_RESTORE_PASS blob=$PAL_CONTRACT_CURRENT_BLOB"\n(( PAL_CERT_STATUS == 0 )) || exit "$PAL_CERT_STATUS"'
[[ -f "$TARGET" ]] || { echo '[RPG integration deep] integration acceptance runner missing' >&2; exit 2; }
COUNT="$(grep -Fxc "$OLD" "$TARGET" || true)"
[[ "$COUNT" = 1 ]] || { echo "[RPG integration deep] expected exactly one compound-local wait_marker seam, found $COUNT" >&2; exit 2; }
python3 - "$TARGET" "$OLD" "$NEW" "$ROG_CERT_OLD" "$ROG_CERT_NEW" "$PAL_CERT_OLD" "$PAL_CERT_NEW" <<'PY'
from pathlib import Path
import sys
path=Path(sys.argv[1]); old=sys.argv[2]; new=sys.argv[3]; rog_old=sys.argv[4]; rog_new=sys.argv[5]; pal_old=sys.argv[6]; pal_new=sys.argv[7]
text=path.read_text()
replacement=new+'\n  local deadline=$((SECONDS+timeout_seconds))'
parser_old = r'''        ids=re.findall(r'(?m)^\s*modId\s*=\s*["\']([^"\']+)',toml)'''
parser_new = r'''        ids=[]
        section=None
        for raw in toml.splitlines():
            line=raw.strip()
            if not line or line.startswith('#'):
                continue
            if line.startswith('[[') and line.endswith(']]'):
                section=line[2:-2].strip()
                continue
            if line.startswith('[') and line.endswith(']'):
                section=line[1:-1].strip()
                continue
            if section == 'mods':
                match=re.match(r'^modId\s*=\s*["\']([^"\']+)["\']\s*(?:#.*)?$', line)
                if match:
                    ids.append(match.group(1))
        if not ids:
            raise SystemExit(f'[RPG integration] no owned [[mods]] modId found in {jar.name}')
        if 'forge' in ids:
            raise SystemExit(f'[RPG integration] dependency modId forge incorrectly appeared as owned in {jar.name}')'''
if text.count(old) != 1:
    raise SystemExit(f'[RPG integration deep] expected one wait_marker seam, found {text.count(old)}')
if text.count(rog_old) != 1:
    raise SystemExit(f'[RPG integration deep] expected one Rogues certification seam, found {text.count(rog_old)}')
if text.count(pal_old) != 1:
    raise SystemExit(f'[RPG integration deep] expected one Paladins certification seam, found {text.count(pal_old)}')
if text.count(parser_old) != 1:
    raise SystemExit(f'[RPG integration deep] expected one mod ownership parser seam, found {text.count(parser_old)}')
text=text.replace(old,replacement,1)
text=text.replace(rog_old,rog_new,1)
text=text.replace(pal_old,pal_new,1)
text=text.replace(parser_old,parser_new,1)
path.write_text(text)
PY
grep -Fq '  local deadline=$((SECONDS+timeout_seconds))' "$TARGET" || { echo '[RPG integration deep] split deadline declaration missing after patch' >&2; exit 2; }
if grep -Fq "$OLD" "$TARGET"; then echo '[RPG integration deep] unsafe compound local declaration survived patch' >&2; exit 2; fi
grep -Fq 'ROGUES_FROZEN_SOURCE_VIEW_PASS' "$TARGET" || { echo '[RPG integration deep] frozen Rogues source-view seam missing after patch' >&2; exit 2; }
grep -Fq '49929e20053d91300d7551c567efc97e93a9d328' "$TARGET" || { echo '[RPG integration deep] frozen Rogues contract blob pin missing after patch' >&2; exit 2; }
grep -Fq 'PALADINS_FROZEN_SOURCE_VIEW_PASS' "$TARGET" || { echo '[RPG integration deep] frozen Paladins source-view seam missing after patch' >&2; exit 2; }
grep -Fq '09b523c7d4688961ec992f77dc08835b12a29bb2' "$TARGET" || { echo '[RPG integration deep] frozen Paladins contract blob pin missing after patch' >&2; exit 2; }
grep -Fq 'no owned [[mods]] modId found' "$TARGET" || { echo '[RPG integration deep] section-aware mod ownership parser missing after patch' >&2; exit 2; }
grep -Fq "section == 'mods'" "$TARGET" || { echo '[RPG integration deep] mod ownership parser is not scoped to [[mods]]' >&2; exit 2; }
PAL_MIXIN_OLD="    if 'MixinConfigs:' not in manifest: raise SystemExit('[RPG integration] Paladins production mixin activation missing')"
PAL_MIXIN_COUNT="$(grep -Fxc "$PAL_MIXIN_OLD" "$TARGET" || true)"
[[ "$PAL_MIXIN_COUNT" = 1 ]] || { echo "[RPG integration deep] expected exactly one Paladins mixin activation seam, found $PAL_MIXIN_COUNT" >&2; exit 2; }
python3 - "$TARGET" "$PAL_MIXIN_OLD" <<'PY_MIXIN'
from pathlib import Path
import sys
path=Path(sys.argv[1]); old=sys.argv[2]
text=path.read_text()
new=r'''    pal_mixin_cfg=__import__('json').loads(z.read('paladins.mixins.json').decode('utf-8','replace'))
    pal_declared_mixins=[entry for key in ('mixins','client','server') for entry in pal_mixin_cfg.get(key,[])]
    if pal_declared_mixins and 'MixinConfigs: paladins.mixins.json' not in manifest:
        raise SystemExit('[RPG integration] Paladins non-empty production mixin config is not activated')
    if not pal_declared_mixins and any(n.startswith('net/paladins/mixin/') and n.endswith('.class') for n in names):
        raise SystemExit('[RPG integration] Paladins empty mixin config disagrees with packaged mixin classes')'''
if text.count(old) != 1:
    raise SystemExit(f'[RPG integration deep] expected one Paladins mixin activation seam, found {text.count(old)}')
text=text.replace(old,new,1)
path.write_text(text)
PY_MIXIN
grep -Fq 'Paladins non-empty production mixin config is not activated' "$TARGET" || { echo '[RPG integration deep] conditional Paladins mixin activation gate missing' >&2; exit 2; }
grep -Fq 'Paladins empty mixin config disagrees with packaged mixin classes' "$TARGET" || { echo '[RPG integration deep] empty Paladins mixin consistency gate missing' >&2; exit 2; }
bash -n "$TARGET"
echo '[RPG integration deep] WAIT_MARKER_HARNESS_HARDENING_PASS'
echo '[RPG integration deep] ROGUES_SOURCE_VIEW_HARDENING_PASS'
echo '[RPG integration deep] PALADINS_SOURCE_VIEW_HARDENING_PASS'
echo '[RPG integration deep] MOD_OWNERSHIP_PARSER_HARDENING_PASS'
echo '[RPG integration deep] PALADINS_MIXIN_CONTRACT_HARDENING_PASS'
bash "$TARGET"
