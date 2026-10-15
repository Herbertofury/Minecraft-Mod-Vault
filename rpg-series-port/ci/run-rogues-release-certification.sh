#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
PORT="$ROOT/rpg-series-port/rogues-forge-1.20.1"
FRESH="$PORT/.fresh-rogues-forge-server"
EXPECTED_JAR_SHA='9e8c880f55ab57d91148c0be702a431bad6e312900b25f65c9dbec266e3ca401'
# First deep run captures the deterministic source identity. It is then frozen here before promotion.
EXPECTED_SOURCE_SHA='__CAPTURE_AFTER_FIRST_DEEP__'
pick_one(){ local dir="$1" pattern="$2" label="$3"; local -a files=(); mapfile -t files < <(find "$dir" -maxdepth 1 -type f -name "$pattern" ! -name '*sources*' ! -name '*dev-shadow*' ! -name '*javadoc*' -print | sort); (( ${#files[@]} == 1 )) || { echo "[Rogues certification] $label expected one $pattern in $dir, found ${#files[@]}" >&2; exit 1; }; printf '%s\n' "${files[0]}"; }
JAR="$(pick_one "$PORT/forge/build/libs" '*-forge-*.jar' 'release')"; unzip -tq "$JAR" >/dev/null
JAR_SHA="$(sha256sum "$JAR" | awk '{print $1}')"; [[ "$JAR_SHA" = "$EXPECTED_JAR_SHA" ]] || { echo "[Rogues certification] exact release mismatch: $JAR_SHA" >&2; exit 1; }
printf '%s  %s\n' "$JAR_SHA" "$JAR" > "$PORT/rogues.sha256"
INSTALLED="$(pick_one "$FRESH/mods" 'rogues-forge-3.1.1+1.20.1.jar' 'packaged-server Rogues')"; cmp -s "$JAR" "$INSTALLED" || { echo '[Rogues certification] packaged release bytes differ' >&2; exit 1; }
[[ "$(find "$FRESH/mods" -maxdepth 1 -type f -name '*.jar' | wc -l | tr -d ' ')" = 9 ]] || { echo '[Rogues certification] packaged runtime dependency count drifted' >&2; exit 1; }
PREVIOUS_SOURCE_SHA="$(awk '{print $1}' "$PORT/rogues-source.sha256")"
SOURCE_ZIP="$ROOT/rogues-3.1.1-forge-1.20.1-source-ci.zip"; rm -f "$SOURCE_ZIP"
python3 - "$PORT" "$SOURCE_ZIP" <<'PY'
from pathlib import Path
import stat,sys,zipfile
src=Path(sys.argv[1]).resolve(); out=Path(sys.argv[2]).resolve(); skip={'.gradle','build','run','runs','.git','.upstream','.fresh-rogues-forge-server'}; files=[]
for p in src.rglob('*'):
    rel=p.relative_to(src)
    if any(x in skip for x in rel.parts): continue
    if len(rel.parts)==1 and rel.name.startswith('rogues') and rel.suffix=='.sha256': continue
    if len(rel.parts)==1 and (rel.name.endswith('-smoke.log') or rel.name.endswith('-behavior.log') or rel.name.endswith('-server.log') or rel.name.endswith('-xvfb.log')): continue
    if p.is_file(): files.append((rel.as_posix(),p))
with zipfile.ZipFile(out,'w',compression=zipfile.ZIP_DEFLATED,compresslevel=9) as zf:
    for arc,p in sorted(files):
        info=zipfile.ZipInfo(arc,date_time=(1980,1,1,0,0,0)); info.compress_type=zipfile.ZIP_DEFLATED; info.external_attr=(stat.S_IFREG|0o644)<<16; info.create_system=3
        zf.writestr(info,p.read_bytes(),compress_type=zipfile.ZIP_DEFLATED,compresslevel=9)
PY
unzip -tq "$SOURCE_ZIP" >/dev/null
SOURCE_SHA="$(sha256sum "$SOURCE_ZIP" | awk '{print $1}')"; [[ "$SOURCE_SHA" = "$PREVIOUS_SOURCE_SHA" ]] || { echo "[Rogues certification] source archive not deterministic across acceptance/certification stages: $PREVIOUS_SOURCE_SHA -> $SOURCE_SHA" >&2; exit 1; }
printf '%s  %s\n' "$SOURCE_SHA" "$SOURCE_ZIP" > "$PORT/rogues-source.sha256"
if [[ "$EXPECTED_SOURCE_SHA" = '__CAPTURE_AFTER_FIRST_DEEP__' ]]; then
  echo "[Rogues certification] SOURCE_IDENTITY_BASELINE_CAPTURE source=$SOURCE_SHA"
else
  [[ "$SOURCE_SHA" = "$EXPECTED_SOURCE_SHA" ]] || { echo "[Rogues certification] source identity drifted: $SOURCE_SHA != $EXPECTED_SOURCE_SHA" >&2; exit 1; }
  echo "[Rogues certification] CROSS_RUN_RELEASE_IDENTITY_PASS jar=$JAR_SHA source=$SOURCE_SHA"
fi
