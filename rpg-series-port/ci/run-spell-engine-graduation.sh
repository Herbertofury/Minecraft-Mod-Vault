#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
BASE="$ROOT/rpg-series-port/ci/run-spell-engine.sh"
PATCHED="$ROOT/rpg-series-port/ci/.run-spell-engine-graduation.generated.sh"
ENV_FILE="$ROOT/rpg-series-port/spell-engine-forge-1.20.1/SPELL_ENGINE_GRADUATION.env"

test -f "$BASE"
test -f "$ENV_FILE"
source "$ENV_FILE"
export SPELL_ENGINE_EXPECTED_JAR_SHA SPELL_ENGINE_EXPECTED_SOURCE_SHA
export SPELL_POWER_160_EXPECTED_JAR_SHA RANGED_WEAPON_API_234_EXPECTED_JAR_SHA TINY_CONFIG_310_EXPECTED_JAR_SHA

python3 - "$BASE" "$PATCHED" <<'PY'
from pathlib import Path
import sys
src=Path(sys.argv[1]).read_text()
out=Path(sys.argv[2])
needle="clone_exact FabricExtras/RangedWeaponAPI \"$RANGED_TARGET\" \"$UP/ranged-234\" & P6=$!\nwait \"$P1\" \"$P2\" \"$P3\" \"$P4\" \"$P5\" \"$P6\"\n"
replacement="clone_exact FabricExtras/RangedWeaponAPI \"$RANGED_TARGET\" \"$UP/ranged-234\" & P6=$!\nclone_exact ZsoltMolnarrr/TinyConfig e20fc8ac72fde8274f0df72de2ebb81ffe6f8727 \"$UP/tiny-config-310\" & P7=$!\nwait \"$P1\" \"$P2\" \"$P3\" \"$P4\" \"$P5\" \"$P6\" \"$P7\"\n"
if src.count(needle) != 1: raise SystemExit(f'[Spell Engine graduation] expected one exact-source fan-in seam, found {src.count(needle)}')
src=src.replace(needle,replacement)
needle='''# Build the two already-verified foundation dependencies as actual separate mods. Spell Engine compiles\n# against their named common JARs and runs against their Forge JARs; their source trees are never added\n# to Spell Engine sourceSets, so dependency implementation classes cannot leak into its release JAR.\nSPELL_POWER="$ROOT/rpg-series-port/spell_power-forge-1.20.1"\n'''
replacement='''# Reconstruct certified TinyConfig 3.1.0. Spell Engine 1.10.2 compiles against its common API and\n# intentionally embeds the Forge artifact via JarJar, matching upstream 1.10.2 packaging.\nTINY_CONFIG="$ROOT/rpg-series-port/tiny-config-forge-1.20.1"\nTINY_GEN="$TINY_CONFIG/generated"\npython3 "$TINY_CONFIG/tools/prepare_port.py" "$UP/tiny-config-310" "$TINY_GEN"\ngradle --no-daemon --stacktrace -p "$TINY_GEN" clean :common:jar :forge:remapJar\nTINY_CONFIG_COMMON_JAR="$(find "$TINY_GEN/common/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' | head -n 1)"\nTINY_CONFIG_FORGE_JAR="$(find "$TINY_GEN/forge/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*dev-shadow*' ! -name '*sources*' | head -n 1)"\ntest -f "$TINY_CONFIG_COMMON_JAR" -a -f "$TINY_CONFIG_FORGE_JAR"\nTINY_CONFIG_ACTUAL_SHA="$(sha256sum "$TINY_CONFIG_FORGE_JAR" | awk '{print $1}')"\necho "[Spell Engine graduation] TinyConfig SHA=$TINY_CONFIG_ACTUAL_SHA expected=$TINY_CONFIG_310_EXPECTED_JAR_SHA"\n[[ "$TINY_CONFIG_ACTUAL_SHA" = "$TINY_CONFIG_310_EXPECTED_JAR_SHA" ]]\necho '[Spell Engine graduation] CERTIFIED_TINY_CONFIG_310_IDENTITY_PASS'\n\n# Build the already-verified external RPG foundation dependencies as actual separate mods. Spell Engine\n# compiles against their named common JARs and runs against their Forge JARs; their source trees are\n# never added to Spell Engine sourceSets, so dependency implementation classes cannot leak into release.\nSPELL_POWER="$ROOT/rpg-series-port/spell_power-forge-1.20.1"\n'''
if src.count(needle) != 1: raise SystemExit(f'[Spell Engine graduation] expected one foundation-build seam, found {src.count(needle)}')
src=src.replace(needle,replacement)
needle='export SPELL_POWER_COMMON_JAR RANGED_COMMON_JAR SPELL_POWER_FORGE_JAR RANGED_FORGE_JAR\n'
replacement='export TINY_CONFIG_COMMON_JAR TINY_CONFIG_FORGE_JAR SPELL_POWER_COMMON_JAR RANGED_COMMON_JAR SPELL_POWER_FORGE_JAR RANGED_FORGE_JAR\n'
if src.count(needle) != 1: raise SystemExit(f'[Spell Engine graduation] expected one dependency export seam, found {src.count(needle)}')
src=src.replace(needle,replacement)
needle='''rm -f "$ROOT/spell-engine-1.10.2-forge-1.20.1-source-ci.zip"\n(cd "$WORK" && zip -qr "$ROOT/spell-engine-1.10.2-forge-1.20.1-source-ci.zip" . -x '*/build/*' '*/run/*' '.gradle/*')\nunzip -t "$ROOT/spell-engine-1.10.2-forge-1.20.1-source-ci.zip" >/dev/null\n'''
replacement='''SOURCE_ZIP="$ROOT/spell-engine-1.10.2-forge-1.20.1-source-ci.zip"\nrm -f "$SOURCE_ZIP"\npython3 - "$WORK" "$SOURCE_ZIP" <<'PYSOURCE'\nfrom pathlib import Path\nimport sys, zipfile\nroot=Path(sys.argv[1]); out=Path(sys.argv[2]); files=[]\nfor p in root.rglob('*'):\n    if not p.is_file(): continue\n    rel=p.relative_to(root); parts=rel.parts\n    if '.gradle' in parts or 'build' in parts or 'run' in parts: continue\n    files.append((rel,p))\nwith zipfile.ZipFile(out,'w',compression=zipfile.ZIP_DEFLATED,compresslevel=9) as z:\n    for rel,p in sorted(files,key=lambda x:x[0].as_posix()):\n        info=zipfile.ZipInfo(rel.as_posix(),(1980,1,1,0,0,0)); info.compress_type=zipfile.ZIP_DEFLATED; info.external_attr=(0o100644 << 16); z.writestr(info,p.read_bytes())\nPYSOURCE\nunzip -t "$SOURCE_ZIP" >/dev/null\nSOURCE_SHA="$(sha256sum "$SOURCE_ZIP" | awk '{print $1}')"\necho "[Spell Engine graduation] DETERMINISTIC_SOURCE_PACKAGE_PASS sha=$SOURCE_SHA"\n'''
if src.count(needle) != 1: raise SystemExit(f'[Spell Engine graduation] expected one historical source-package seam, found {src.count(needle)}')
src=src.replace(needle,replacement)
needle="unzip -l \"$OUT_JAR\" | grep -F 'META-INF/jars/TinyConfig-'\n"
replacement='''FIRST_SHA="$(sha256sum "$OUT_JAR" | awk '{print $1}')"\nif unzip -Z1 "$OUT_JAR" | grep -Eiq 'META-INF/jars/[^/]*(tiny.?config|TinyConfig)[^/]*2\\.'; then echo 'Obsolete TinyConfig 2.x leaked into Spell Engine 1.10.2 release' >&2; exit 1; fi\nTINY_NESTED="$(unzip -Z1 "$OUT_JAR" | grep -Ei '^META-INF/jars/.*tiny.?config.*\\.jar$' | head -n 1 || true)"\n[[ -n "$TINY_NESTED" ]] || { echo 'Spell Engine release lost upstream TinyConfig JarJar payload' >&2; exit 1; }\nunzip -p "$OUT_JAR" "$TINY_NESTED" > "$PORT/tiny-config-embedded-check.jar"\nunzip -tq "$PORT/tiny-config-embedded-check.jar" >/dev/null\n[[ "$(sha256sum "$PORT/tiny-config-embedded-check.jar" | awk '{print $1}')" = "$TINY_CONFIG_310_EXPECTED_JAR_SHA" ]] || { echo 'Spell Engine embedded TinyConfig is not the certified 3.1.0 Forge artifact' >&2; exit 1; }\necho '[Spell Engine graduation] CERTIFIED_TINY_CONFIG_310_EMBEDDED_PASS'\ngradle --no-daemon --stacktrace -p "$WORK" clean :forge:build\nJAR2="$(find "$WORK/forge/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*dev-shadow*' ! -name '*sources*' | head -n 1)"\nSECOND_SHA="$(sha256sum "$JAR2" | awk '{print $1}')"\n[[ "$FIRST_SHA" = "$SECOND_SHA" ]]; cmp -s "$JAR2" "$OUT_JAR"\necho "[Spell Engine graduation] CLEAN_REBUILD_IDENTITY_PASS sha=$FIRST_SHA"\n'''
if src.count(needle) != 1: raise SystemExit(f'[Spell Engine graduation] expected one legacy TinyConfig package gate, found {src.count(needle)}')
src=src.replace(needle,replacement)
needle='echo "[Spell Engine CI] Forge client bootstrap passed with external RPG dependency mods; JAR: $OUT_JAR"\n'
server=r'''echo "[Spell Engine CI] Forge client bootstrap passed with external RPG dependency mods; JAR: $OUT_JAR"
echo '[Spell Engine graduation] Fresh packaged Forge server gate'
find_runtime_mod() {
  local pattern="$1" modid="$2" jar
  while IFS= read -r jar; do
    [[ -f "$jar" ]] || continue
    if unzip -p "$jar" META-INF/mods.toml 2>/dev/null | grep -Eq "modId[[:space:]]*=[[:space:]]*\\?\"${modid}\\?\""; then printf '%s\n' "$jar"; return 0; fi
  done < <(find "${GRADLE_USER_HOME:-$HOME/.gradle}/caches/modules-2/files-2.1" -type f -name "$pattern" 2>/dev/null | sort)
  return 1
}
CLOTH_FORGE_JAR="$(find_runtime_mod 'cloth-config-forge-*.jar' cloth_config)"
PLAYER_ANIM_FORGE_JAR="$(find_runtime_mod 'player-animation-lib-forge-*.jar' playeranimator)"
test -f "$CLOTH_FORGE_JAR" -a -f "$PLAYER_ANIM_FORGE_JAR"
FRESH="$PORT/.fresh-spell-engine-server"; rm -rf "$FRESH"; mkdir -p "$FRESH/mods"
curl -fsSL 'https://maven.minecraftforge.net/net/minecraftforge/forge/1.20.1-47.4.23/forge-1.20.1-47.4.23-installer.jar' -o "$FRESH/forge-installer.jar"
( cd "$FRESH"; java -jar forge-installer.jar --installServer >/dev/null; printf 'eula=true\n' > eula.txt; printf '%s\n' '-Xmx3G' > user_jvm_args.txt; cp "$OUT_JAR" mods/; cp "$SPELL_POWER_FORGE_JAR" mods/spell_power-forge.jar; cp "$CLOTH_FORGE_JAR" mods/cloth-config-forge.jar; cp "$PLAYER_ANIM_FORGE_JAR" mods/player-animation-lib-forge.jar )
[[ "$(sha256sum "$FRESH/mods/$(basename "$OUT_JAR")" | awk '{print $1}')" = "$FIRST_SHA" ]]
PACKAGE_LOG="$PORT/forge-packaged-server-smoke.log"; : > "$PACKAGE_LOG"
( cd "$FRESH" && exec ./run.sh nogui ) > "$PACKAGE_LOG" 2>&1 & ACTIVE_PID=$!; PID=$ACTIVE_PID; DEADLINE=$((SECONDS+180)); PASS=0
while ((SECONDS<DEADLINE)); do
  LOG="$FRESH/logs/latest.log"; FILES=("$PACKAGE_LOG"); [[ -f "$LOG" ]] && FILES+=("$LOG")
  if grep -Eiq 'ModLoadingException|Failed to create mod instance|NoClassDefFoundError|ClassNotFoundException|MixinApplyError|InvalidMixinException|MixinTransformerError|Exception in server tick loop|The game crashed|Spell Engine CI self-test:' "${FILES[@]}"; then stop_tree "$PID"; ACTIVE_PID=""; dump_logs "${FILES[@]}"; exit 1; fi
  if [[ -f "$LOG" ]] && grep -Fq '[Spell Engine CI] Packaged runtime self-test passed:' "$LOG" && grep -Eq 'Done \([0-9.]+s\)!' "$LOG"; then PASS=1; break; fi
  if ! kill -0 "$PID" 2>/dev/null; then wait "$PID" || true; ACTIVE_PID=""; dump_logs "${FILES[@]}"; exit 1; fi
  sleep 1
done
[[ "$PASS" -eq 1 ]] || { stop_tree "$PID"; ACTIVE_PID=""; dump_logs "$PACKAGE_LOG" "$FRESH/logs/latest.log"; exit 1; }
stop_tree "$PID"; ACTIVE_PID=""
echo '[Spell Engine graduation] CANONICAL_PACKAGED_SERVER_AND_SELF_TEST_PASS'
if [[ "$SPELL_ENGINE_EXPECTED_JAR_SHA" = '__CAPTURE_AFTER_FIRST_FULL_GREEN__' || "$SPELL_ENGINE_EXPECTED_SOURCE_SHA" = '__CAPTURE_AFTER_FIRST_FULL_GREEN__' ]]; then
  echo "[Spell Engine graduation] SPELL_ENGINE_FIRST_GREEN_CAPTURE jar=$FIRST_SHA source=$SOURCE_SHA"
else
  [[ "$FIRST_SHA" = "$SPELL_ENGINE_EXPECTED_JAR_SHA" ]]; [[ "$SOURCE_SHA" = "$SPELL_ENGINE_EXPECTED_SOURCE_SHA" ]]
  echo "[Spell Engine graduation] CROSS_RUN_RELEASE_IDENTITY_PASS jar=$FIRST_SHA source=$SOURCE_SHA"; echo '[Spell Engine graduation] SPELL_ENGINE_GRADUATION_PASS'
fi
'''
if src.count(needle) != 1: raise SystemExit(f'[Spell Engine graduation] expected one historical client finish marker, found {src.count(needle)}')
src=src.replace(needle,server)
out.write_text(src)
PY
chmod +x "$PATCHED"
bash -n "$PATCHED"
echo '[Spell Engine graduation] TINY_CONFIG_310_PACKAGED_SERVER_AND_FREEZE_WRAPPER_READY'
exec bash "$PATCHED"
