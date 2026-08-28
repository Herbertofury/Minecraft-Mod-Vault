#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
PORT="$ROOT/rpg-series-port/paladins-forge-1.20.1"
CERTIFIER="$ROOT/rpg-series-port/ci/certify-paladins-run199.py"
FRESH="$PORT/.fresh-paladins-forge-server"
EXPECTED_JAR_SHA="95e8f9e074dd1c432f9486c922173b76a3b3bc57667b28fe154e47dba5c374ee"
EXPECTED_SOURCE_SHA="fb0e5812857a2fd46de488cd17a80011ef5d18795ff96fa1a3ebed5fd19a4377"

pick_one() {
  local dir="$1" pattern="$2" label="$3"
  local -a files=()
  mapfile -t files < <(find "$dir" -maxdepth 1 -type f -name "$pattern" ! -name '*sources*' ! -name '*dev-shadow*' ! -name '*javadoc*' -print | sort)
  (( ${#files[@]} == 1 )) || { echo "[Paladins certification] $label expected one $pattern in $dir, found ${#files[@]}" >&2; exit 1; }
  printf '%s\n' "${files[0]}"
}

JAR="$(pick_one "$PORT/forge/build/libs" '*-forge-*.jar' 'build release')"
TMP_JAR="$PORT/forge/build/libs/.paladins-certified.jar"
python3 "$CERTIFIER" "$JAR" "$TMP_JAR"
mv -f "$TMP_JAR" "$JAR"
unzip -tq "$JAR" >/dev/null
JAR_SHA="$(sha256sum "$JAR" | awk '{print $1}')"
[[ "$JAR_SHA" = "$EXPECTED_JAR_SHA" ]] || { echo "[Paladins certification] exact release mismatch: $JAR_SHA" >&2; exit 1; }
printf '%s  %s\n' "$JAR_SHA" "$JAR" | tee "$PORT/paladins.sha256"
printf '%s  %s\n' "$JAR_SHA" "$JAR" > "$PORT/paladins-forge-1.20.1.sha256"

INSTALLED="$(pick_one "$FRESH/mods" 'paladins-forge-3.1.1+1.20.1.jar' 'packaged-server Paladins')"
cp -f "$JAR" "$INSTALLED"
unzip -tq "$INSTALLED" >/dev/null
INSTALLED_SHA="$(sha256sum "$INSTALLED" | awk '{print $1}')"
[[ "$INSTALLED_SHA" = "$EXPECTED_JAR_SHA" ]] || { echo "[Paladins certification] packaged-server release mismatch: $INSTALLED_SHA" >&2; exit 1; }
printf '%s  %s\n' "$INSTALLED_SHA" "$INSTALLED" | tee "$PORT/paladins-package-installed.sha256"
[[ "$(find "$FRESH/mods" -maxdepth 1 -type f -name '*.jar' | wc -l | tr -d ' ')" = 11 ]] || { echo '[Paladins certification] packaged runtime dependency count drifted' >&2; exit 1; }

SOURCE_ZIP="$ROOT/paladins-3.1.1-forge-1.20.1-source-ci.zip"
rm -f "$SOURCE_ZIP"
python3 - "$PORT" "$SOURCE_ZIP" <<'PY'
from pathlib import Path
import stat, sys, zipfile
src = Path(sys.argv[1]).resolve(); out = Path(sys.argv[2]).resolve()
skip = {'.gradle', 'build', 'run', 'runs', '.git', '.upstream'}
files=[]
for path in src.rglob('*'):
    rel=path.relative_to(src)
    if any(part in skip for part in rel.parts):
        continue
    if len(rel.parts) == 1 and rel.name.startswith('paladins') and rel.suffix == '.sha256':
        continue
    if path.is_file():
        files.append((rel.as_posix(), path))
with zipfile.ZipFile(out,'w',compression=zipfile.ZIP_DEFLATED,compresslevel=9) as zf:
    for arcname,path in sorted(files):
        info=zipfile.ZipInfo(arcname,date_time=(1980,1,1,0,0,0)); info.compress_type=zipfile.ZIP_DEFLATED
        info.external_attr=(stat.S_IFREG|0o644)<<16; info.create_system=3
        zf.writestr(info,path.read_bytes(),compress_type=zipfile.ZIP_DEFLATED,compresslevel=9)
PY
unzip -tq "$SOURCE_ZIP" >/dev/null
if unzip -Z1 "$SOURCE_ZIP" | grep -E '(^|/)(\.gradle|build|run|runs|\.upstream)/|^paladins.*\.sha256$' >/dev/null; then
  echo '[Paladins certification] generated/cache/runtime material leaked into certified source archive' >&2
  exit 1
fi
SOURCE_SHA="$(sha256sum "$SOURCE_ZIP" | awk '{print $1}')"
[[ "$SOURCE_SHA" = "$EXPECTED_SOURCE_SHA" ]] || { echo "[Paladins certification] source archive drifted: $SOURCE_SHA != $EXPECTED_SOURCE_SHA" >&2; exit 1; }
printf '%s  %s\n' "$SOURCE_SHA" "$SOURCE_ZIP" | tee "$PORT/paladins-source.sha256"
echo "[Paladins certification] CROSS_RUN_RELEASE_IDENTITY_PASS jar=$JAR_SHA source=$SOURCE_SHA"
