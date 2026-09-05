#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
PORT="$ROOT/rpg-series-port/more-rpg-library-forge-1.20.1"
WORK="$ROOT/.more-rpg-library-build"
UP="$ROOT/.more-rpg-library-upstream"
REPLAY="$ROOT/.more-rpg-library-replay"
FOUNDATION="$PORT/.foundation"
OUT="$PORT/more_rpg_library-forge-2.7.2+1.20.1.jar"
FIRST_SOURCE_FILE="$PORT/more-rpg-library-source.sha256"
for f in "$OUT" "$FIRST_SOURCE_FILE" "$UP/modern-272/.git/HEAD" "$UP/old-1201/.git/HEAD"; do test -e "$f"; done
FIRST_JAR_SHA="$(sha256sum "$OUT" | awk '{print $1}')"
FIRST_SOURCE_SHA="$(awk 'NR==1{print $1}' "$FIRST_SOURCE_FILE")"
[[ "$FIRST_JAR_SHA" =~ ^[0-9a-f]{64}$ && "$FIRST_SOURCE_SHA" =~ ^[0-9a-f]{64}$ ]]
echo "[More RPG 2.7.2] INDEPENDENT_REPLAY_BEGIN first_jar=$FIRST_JAR_SHA first_source=$FIRST_SOURCE_SHA"
rm -rf "$REPLAY"
python3 "$PORT/tools/prepare_more_rpg_library_named_common.py" "$UP/modern-272" "$UP/old-1201" "$REPLAY"
(cd "$REPLAY" && find common forge -type f -print0 | sort -z | xargs -0 sha256sum) > "$PORT/more-rpg-library-replay-source-manifest.sha256"
REPLAY_SOURCE_SHA="$(sha256sum "$PORT/more-rpg-library-replay-source-manifest.sha256" | awk '{print $1}')"
[[ "$REPLAY_SOURCE_SHA" = "$FIRST_SOURCE_SHA" ]] || { echo "[More RPG 2.7.2] replay source identity mismatch first=$FIRST_SOURCE_SHA replay=$REPLAY_SOURCE_SHA" >&2; exit 1; }
cmp -s "$PORT/more-rpg-library-source-manifest.sha256" "$PORT/more-rpg-library-replay-source-manifest.sha256"
echo "[More RPG 2.7.2] INDEPENDENT_SOURCE_REPLAY_IDENTITY_PASS sha=$FIRST_SOURCE_SHA"
SPELL_ENGINE_JAR="$ROOT/rpg-series-port/spell-engine-forge-1.20.1/spell_engine-forge-1.10.4+1.20.1.jar"
SPELL_POWER_JAR="$FOUNDATION/spell_power-forge-1.6.0+1.20.1-certified.jar"
RANGED_JAR="$FOUNDATION/ranged_weapon_api-forge-2.3.4+1.20.1-certified.jar"
TINY_JAR="$(find "$ROOT/rpg-series-port/tiny-config-forge-1.20.1/generated/forge/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' ! -name '*dev-shadow*' | sort | head -n1)"
SPELL_ENGINE_COMMON_JAR="$(find "$ROOT/.spell-engine-build/common/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' | sort | head -n1)"
SPELL_POWER_COMMON_JAR="$(find "$ROOT/rpg-series-port/spell_power-forge-1.20.1/common/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' | sort | head -n1)"
RANGED_COMMON_JAR="$(find "$ROOT/rpg-series-port/ranged-weapon-api-forge-1.20.1/common/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' | sort | head -n1)"
TINY_COMMON_JAR="$(find "$ROOT/rpg-series-port/tiny-config-forge-1.20.1/generated/common/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' | sort | head -n1)"
for f in "$SPELL_ENGINE_JAR" "$SPELL_POWER_JAR" "$RANGED_JAR" "$TINY_JAR" "$SPELL_ENGINE_COMMON_JAR" "$SPELL_POWER_COMMON_JAR" "$RANGED_COMMON_JAR" "$TINY_COMMON_JAR"; do test -f "$f"; unzip -tq "$f" >/dev/null; done
ARGS=("-Pspell_engine_common_jar=$SPELL_ENGINE_COMMON_JAR" "-Pspell_power_common_jar=$SPELL_POWER_COMMON_JAR" "-Pranged_weapon_api_common_jar=$RANGED_COMMON_JAR" "-Ptiny_config_common_jar=$TINY_COMMON_JAR" "-Pspell_engine_forge_jar=$SPELL_ENGINE_JAR" "-Pspell_power_forge_jar=$SPELL_POWER_JAR" "-Pranged_weapon_api_forge_jar=$RANGED_JAR" "-Ptiny_config_forge_jar=$TINY_JAR")
gradle --no-daemon --stacktrace -p "$REPLAY" clean :forge:build "${ARGS[@]}"
REPLAY_JAR="$(find "$REPLAY/forge/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' ! -name '*dev-shadow*' | sort | head -n1)"
test -f "$REPLAY_JAR"; unzip -tq "$REPLAY_JAR" >/dev/null
REPLAY_JAR_SHA="$(sha256sum "$REPLAY_JAR" | awk '{print $1}')"
[[ "$REPLAY_JAR_SHA" = "$FIRST_JAR_SHA" ]] || { echo "[More RPG 2.7.2] replay JAR identity mismatch first=$FIRST_JAR_SHA replay=$REPLAY_JAR_SHA" >&2; exit 1; }
cmp -s "$OUT" "$REPLAY_JAR"
echo "[More RPG 2.7.2] INDEPENDENT_JAR_REPLAY_IDENTITY_PASS sha=$FIRST_JAR_SHA"
python3 - "$WORK" "$PORT/more-rpg-library-2.7.2-forge-1.20.1-source-ci.zip" "$REPLAY" "$PORT/more-rpg-library-2.7.2-forge-1.20.1-source-replay-ci.zip" <<'PY'
from pathlib import Path
import sys, zipfile
def pack(root, out):
    root=Path(root); out=Path(out); files=[]
    for p in root.rglob('*'):
        if not p.is_file(): continue
        rel=p.relative_to(root); parts=rel.parts
        if '.gradle' in parts or 'build' in parts or 'run' in parts: continue
        files.append((rel,p))
    with zipfile.ZipFile(out,'w',compression=zipfile.ZIP_DEFLATED,compresslevel=9) as z:
        for rel,p in sorted(files,key=lambda x:x[0].as_posix()):
            info=zipfile.ZipInfo(rel.as_posix(),(1980,1,1,0,0,0)); info.compress_type=zipfile.ZIP_DEFLATED; info.external_attr=(0o100644 << 16); z.writestr(info,p.read_bytes())
pack(sys.argv[1],sys.argv[2]); pack(sys.argv[3],sys.argv[4])
PY
SOURCE_ZIP="$PORT/more-rpg-library-2.7.2-forge-1.20.1-source-ci.zip"; REPLAY_ZIP="$PORT/more-rpg-library-2.7.2-forge-1.20.1-source-replay-ci.zip"
unzip -tq "$SOURCE_ZIP" >/dev/null; unzip -tq "$REPLAY_ZIP" >/dev/null
SOURCE_ZIP_SHA="$(sha256sum "$SOURCE_ZIP" | awk '{print $1}')"; REPLAY_ZIP_SHA="$(sha256sum "$REPLAY_ZIP" | awk '{print $1}')"
[[ "$SOURCE_ZIP_SHA" = "$REPLAY_ZIP_SHA" ]]; cmp -s "$SOURCE_ZIP" "$REPLAY_ZIP"
printf '%s  %s\n' "$SOURCE_ZIP_SHA" "$(basename "$SOURCE_ZIP")" > "$PORT/more-rpg-library-source-package.sha256"
echo "[More RPG 2.7.2] DETERMINISTIC_SOURCE_PACKAGE_IDENTITY_PASS sha=$SOURCE_ZIP_SHA"
echo "[More RPG 2.7.2] INDEPENDENT_REPLAY_IDENTITY_PASS jar=$FIRST_JAR_SHA source=$FIRST_SOURCE_SHA source_zip=$SOURCE_ZIP_SHA"
