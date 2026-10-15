#!/usr/bin/env bash
set -euo pipefail

ROOT="${GITHUB_WORKSPACE:?}"
PORT="$ROOT/rpg-series-port/gazebos-forge-1.20.1"
UPROOT="$ROOT/.rpg-upstream"
GAZEBO_UP="$UPROOT/gazebo-2.2.0"
TINY_UP="$UPROOT/tiny-config-3.1.0"
GEN="$PORT/generated"
TARGET_SHA="bc5e9f49e16d2ff31fb6d3aa31bab55ba0a634ee"
TINY_SHA="e20fc8ac72fde8274f0df72de2ebb81ffe6f8727"
ENV_FILE="$PORT/GAZEBOS_GRADUATION.env"
ACTIVE_PID=""
source "$ENV_FILE"

stop_tree() {
  local root="${1:-}" child kids
  [[ -n "$root" ]] || return 0
  kids="$(pgrep -P "$root" 2>/dev/null || true)"
  for child in $kids; do stop_tree "$child"; done
  kill -TERM "$root" 2>/dev/null || true; sleep 1
  kill -KILL "$root" 2>/dev/null || true; wait "$root" 2>/dev/null || true
}
cleanup() { [[ -z "${ACTIVE_PID:-}" ]] || stop_tree "$ACTIVE_PID"; }
trap cleanup EXIT INT TERM

clone_exact() {
  local repo="$1" sha="$2" dest="$3"
  rm -rf "$dest"; mkdir -p "$(dirname "$dest")"; git init -q "$dest"
  git -C "$dest" remote add origin "https://github.com/${repo}.git"
  git -C "$dest" fetch -q --depth=1 origin "$sha"
  git -C "$dest" checkout -q --detach FETCH_HEAD
  [[ "$(git -C "$dest" rev-parse HEAD)" = "$sha" ]]
}
pick_release_jar() {
  find "$1" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' ! -name '*dev-shadow*' ! -name '*javadoc*' | sort | head -n1
}

rm -rf "$UPROOT"
mkdir -p "$UPROOT"
clone_exact ZsoltMolnarrr/Gazebo "$TARGET_SHA" "$GAZEBO_UP"
clone_exact ZsoltMolnarrr/TinyConfig "$TINY_SHA" "$TINY_UP"

# Build the exact graduated dependency sources from the canonical workspace.
echo '[Gazebos] Build graduated Structure Pool API foundation'
gradle --no-daemon --stacktrace -p "$ROOT/rpg-series-port/structure_pool_api-forge-1.20.1" clean :forge:remapJar
SPA_JAR="$(pick_release_jar "$ROOT/rpg-series-port/structure_pool_api-forge-1.20.1/forge/build/libs")"
[[ -f "$SPA_JAR" ]]

echo '[Gazebos] Reconstruct graduated TinyConfig foundation'
TINY_GEN="$PORT/.tiny-config-generated"
python3 "$ROOT/rpg-series-port/tiny-config-forge-1.20.1/tools/prepare_port.py" "$TINY_UP" "$TINY_GEN"
gradle --no-daemon --stacktrace -p "$TINY_GEN" clean :forge:remapJar
TINY_JAR="$(pick_release_jar "$TINY_GEN/forge/build/libs")"
[[ -f "$TINY_JAR" ]]
[[ "$(sha256sum "$TINY_JAR" | awk '{print $1}')" = '0182a492d6c59d7d5f491a39bb2f6634ba5dd38083295305c4769fdb6539db18' ]]

prepare() {
  python3 "$PORT/tools/prepare_port.py" "$GAZEBO_UP" "$GEN"
  mkdir -p "$GEN/libs"
  cp -f "$SPA_JAR" "$GEN/libs/structure_pool_api-forge.jar"
  cp -f "$TINY_JAR" "$GEN/libs/tiny_config-forge.jar"
}
prepare

grep -Fx "feature_authority=$TARGET_SHA" "$GEN/PORT-PINS.txt" >/dev/null
grep -Fx 'target_version=2.2.0' "$GEN/PORT-PINS.txt" >/dev/null
grep -Fx 'minecraft=1.20.1' "$GEN/PORT-PINS.txt" >/dev/null
if grep -R -E 'net\.neoforged|NeoForge' "$GEN/forge/src/main/java"; then
  echo '[Gazebos] NeoForge leakage in Forge source' >&2; exit 1
fi
[[ ! -e "$GEN/common/src/main/resources/data/gazebo/structure" ]]
[[ ! -e "$GEN/common/src/main/resources/data/gazebo/loot_table" ]]
[[ "$(find "$GEN/common/src/main/resources/data/gazebo/structures" -type f -name '*.nbt' | wc -l)" -eq 17 ]]
[[ "$(find "$GEN/common/src/main/resources/data/gazebo/rs_pieces_spawn_counts" -type f -name '*.json' | wc -l)" -eq 12 ]]
[[ "$(find "$GEN/common/src/main/resources/data/gazebo/rs_pool_additions" -type f -name '*.json' | wc -l)" -eq 12 ]]
[[ "$(find "$GEN/common/src/main/resources/data/gazebo/lithostitched/worldgen_modifier/village" -type f -name '*.json' | wc -l)" -eq 5 ]]

echo '[Gazebos] Deterministic source package gate'
SOURCE_ZIP="$ROOT/gazebos-2.2.0-forge-1.20.1-source-ci.zip"
rm -f "$SOURCE_ZIP"
python3 - "$GEN" "$SOURCE_ZIP" <<'PY'
from pathlib import Path
import stat,sys,zipfile
src=Path(sys.argv[1]).resolve(); out=Path(sys.argv[2]).resolve(); skip={'.gradle','build','run','runs','.git','libs'}
files=[]
for p in src.rglob('*'):
    rel=p.relative_to(src)
    if any(part in skip for part in rel.parts): continue
    if p.is_file(): files.append((rel.as_posix(),p))
with zipfile.ZipFile(out,'w',zipfile.ZIP_DEFLATED,compresslevel=9) as z:
    for name,p in sorted(files):
        i=zipfile.ZipInfo(name,(1980,1,1,0,0,0)); i.compress_type=zipfile.ZIP_DEFLATED; i.external_attr=(stat.S_IFREG|0o644)<<16; i.create_system=3
        z.writestr(i,p.read_bytes(),compress_type=zipfile.ZIP_DEFLATED,compresslevel=9)
PY
unzip -tq "$SOURCE_ZIP" >/dev/null
SOURCE_SHA="$(sha256sum "$SOURCE_ZIP" | awk '{print $1}')"
printf '%s  %s\n' "$SOURCE_SHA" "$SOURCE_ZIP" | tee "$PORT/gazebos-source.sha256"

echo '[Gazebos] Build/package gate'
gradle --no-daemon --stacktrace -p "$GEN" clean :forge:remapJar
JAR="$(pick_release_jar "$GEN/forge/build/libs")"; [[ -f "$JAR" ]]; unzip -tq "$JAR" >/dev/null
unzip -p "$JAR" META-INF/mods.toml | grep -F 'modId="gazebo"' >/dev/null
unzip -p "$JAR" META-INF/mods.toml | grep -F 'modId="structure_pool_api"' >/dev/null
unzip -p "$JAR" META-INF/mods.toml | grep -F 'modId="tiny_config"' >/dev/null
[[ "$(unzip -Z1 "$JAR" | grep -E '^data/gazebo/structures/.+\.nbt$' | wc -l)" -eq 17 ]]
[[ "$(unzip -Z1 "$JAR" | grep -E '^data/gazebo/rs_pieces_spawn_counts/.+\.json$' | wc -l)" -eq 12 ]]
[[ "$(unzip -Z1 "$JAR" | grep -E '^data/gazebo/rs_pool_additions/.+\.json$' | wc -l)" -eq 12 ]]
[[ "$(unzip -Z1 "$JAR" | grep -E '^data/gazebo/lithostitched/worldgen_modifier/village/.+\.json$' | wc -l)" -eq 5 ]]
if unzip -Z1 "$JAR" | grep -E '^data/gazebo/(structure|loot_table)/' >/dev/null; then echo '[Gazebos] 1.21 singular resource folder leaked' >&2; exit 1; fi
python3 - "$JAR" <<'PY'
import struct,sys,zipfile
owned=0
with zipfile.ZipFile(sys.argv[1]) as z:
    for n in z.namelist():
        if n.startswith('net/gazebo/') and n.endswith('.class'):
            d=z.read(n); m=struct.unpack('>H',d[6:8])[0]; owned+=1
            if m!=61: raise SystemExit(f'[Gazebos] non-Java17 class {n} major={m}')
if owned < 5: raise SystemExit(f'[Gazebos] incomplete owned class inventory {owned}')
print(f'[Gazebos] JAVA17_PACKAGE_PASS owned={owned}')
PY
FIRST_SHA="$(sha256sum "$JAR" | awk '{print $1}')"
RELEASE_JAR="$PORT/gazebos-forge-2.2.0+1.20.1-release.jar"; cp -f "$JAR" "$RELEASE_JAR"
printf '%s  %s\n' "$FIRST_SHA" "$RELEASE_JAR" | tee "$PORT/gazebos-first-build.sha256"

echo '[Gazebos] Clean reproducibility gate'
gradle --no-daemon --stacktrace -p "$GEN" clean :forge:remapJar
JAR2="$(pick_release_jar "$GEN/forge/build/libs")"; SECOND_SHA="$(sha256sum "$JAR2" | awk '{print $1}')"
[[ "$FIRST_SHA" = "$SECOND_SHA" ]]; cmp -s "$JAR2" "$RELEASE_JAR"
printf '%s  %s\n' "$SECOND_SHA" "$RELEASE_JAR" | tee "$PORT/gazebos-forge.sha256"

# QA-only probe after product sealing: modern GazeboMod must queue five vanilla pools when
# Lithostitched is absent. This code never enters RELEASE_JAR.
ENTRY="$GEN/forge/src/main/java/net/gazebo/forge/ForgeMod.java"
python3 - "$ENTRY" <<'PY'
from pathlib import Path
import sys
p=Path(sys.argv[1]); s=p.read_text(); old='        GazeboMod.init();\n'
new='''        GazeboMod.init();\n        if ("1".equals(System.getenv("GAZEBO_SELF_TEST"))) {\n            try {\n                var f = net.fabric_extras.structure_pool.api.StructurePoolAPI.class.getDeclaredField("pendingInjections");\n                f.setAccessible(true);\n                var q = (java.util.List<?>) f.get(null);\n                int expected = Integer.parseInt(System.getenv().getOrDefault("GAZEBO_EXPECT_PENDING", "5"));\n                if (q.size() != expected) throw new IllegalStateException("[Gazebos CI] pending injections=" + q.size() + " expected=" + expected);\n                System.out.println("GAZEBO_INJECTION_SEMANTICS_PASS pending=" + q.size());\n            } catch (ReflectiveOperationException e) { throw new RuntimeException(e); }\n        }\n'''
if s.count(old)!=1: raise SystemExit('[Gazebos] expected exactly one init hook')
p.write_text(s.replace(old,new))
PY

gradle --no-daemon --stacktrace -p "$GEN" :forge:classes
SERVER_SMOKE="$PORT/gazebos-server-semantics.log"; : > "$SERVER_SMOKE"
rm -rf "$GEN/forge/run"; mkdir -p "$GEN/forge/run"; printf 'eula=true\n' > "$GEN/forge/run/eula.txt"
env GAZEBO_SELF_TEST=1 GAZEBO_EXPECT_PENDING=5 gradle --no-daemon -p "$GEN" :forge:runServer > "$SERVER_SMOKE" 2>&1 &
ACTIVE_PID=$!; DEADLINE=$((SECONDS+150)); PASS=0
FATAL='\[Gazebos CI\]|MixinApplyError|InvalidMixinException|Failed to create mod instance|NoClassDefFoundError|ClassNotFoundException|Exception in server tick loop|The game crashed'
while ((SECONDS<DEADLINE)); do
  LOG=$(find "$GEN/forge/run" -type f -path '*/logs/latest.log' | head -n1 || true); FILES=("$SERVER_SMOKE"); [[ -n "$LOG" ]] && FILES+=("$LOG")
  if grep -Eiq "$FATAL" "${FILES[@]}"; then stop_tree "$ACTIVE_PID"; ACTIVE_PID=""; cat "${FILES[@]}"; exit 1; fi
  if grep -Fq 'GAZEBO_INJECTION_SEMANTICS_PASS pending=5' "${FILES[@]}" && [[ -n "$LOG" ]] && grep -Eq 'Done \([0-9.]+s\)!' "$LOG"; then PASS=1; break; fi
  if ! kill -0 "$ACTIVE_PID" 2>/dev/null; then wait "$ACTIVE_PID" || true; ACTIVE_PID=""; cat "${FILES[@]}"; exit 1; fi
  sleep 1
done
[[ "$PASS" -eq 1 ]]; stop_tree "$ACTIVE_PID"; ACTIVE_PID=""
echo '[Gazebos] VANILLA_INJECTION_SEMANTICS_PASS pending=5'

# Reconstruct pristine product for client and packaged-server lanes.
prepare
rm -rf "$GEN/forge/run"; mkdir -p "$GEN/forge/run/config"; printf 'earlyWindowControl = false\n' > "$GEN/forge/run/config/fml.toml"
CLIENT_SMOKE="$PORT/gazebos-client-smoke.log"; : > "$CLIENT_SMOKE"
env LIBGL_ALWAYS_SOFTWARE=1 MESA_LOADER_DRIVER_OVERRIDE=llvmpipe xvfb-run -a -s '-screen 0 1280x720x24 +extension GLX +extension RENDER -noreset' gradle --no-daemon -p "$GEN" :forge:runClient </dev/null > "$CLIENT_SMOKE" 2>&1 &
ACTIVE_PID=$!; DEADLINE=$((SECONDS+180)); READY=0; PASS=0
while ((SECONDS<DEADLINE)); do
  LOG=$(find "$GEN/forge/run" -type f -path '*/logs/latest.log' | head -n1 || true); FILES=("$CLIENT_SMOKE"); [[ -n "$LOG" ]] && FILES+=("$LOG")
  if grep -Eiq 'MixinApplyError|InvalidMixinException|Failed to create mod instance|NoClassDefFoundError.*gazebo|ClassNotFoundException.*gazebo|The game crashed whilst initializing game|Could not initialize GLFW' "${FILES[@]}"; then stop_tree "$ACTIVE_PID"; ACTIVE_PID=""; cat "${FILES[@]}"; exit 1; fi
  if [[ -n "$LOG" ]] && grep -Fq 'Backend library: LWJGL' "$LOG" && grep -Fq 'Reloading ResourceManager' "$LOG"; then [[ "$READY" -ne 0 ]] || READY=$SECONDS; if ((SECONDS-READY>=5)); then PASS=1; break; fi; fi
  if ! kill -0 "$ACTIVE_PID" 2>/dev/null; then wait "$ACTIVE_PID" || true; ACTIVE_PID=""; cat "${FILES[@]}"; exit 1; fi
  sleep 1
done
[[ "$PASS" -eq 1 ]]; stop_tree "$ACTIVE_PID"; ACTIVE_PID=""
echo '[Gazebos] NATIVE_FORGE_CLIENT_BOOTSTRAP_PASS'

# Fresh official packaged server using untouched certified product + both graduated foundations.
FRESH="$PORT/.fresh-gazebo-server"; rm -rf "$FRESH"; mkdir -p "$FRESH/mods"
curl -fsSL 'https://maven.minecraftforge.net/net/minecraftforge/forge/1.20.1-47.4.23/forge-1.20.1-47.4.23-installer.jar' -o "$FRESH/forge-installer.jar"
( cd "$FRESH"; java -jar forge-installer.jar --installServer >/dev/null; printf 'eula=true\n' > eula.txt; printf '%s\n' '-Xmx2G' > user_jvm_args.txt; cp "$RELEASE_JAR" mods/; cp "$SPA_JAR" mods/structure_pool_api-forge.jar; cp "$TINY_JAR" mods/tiny_config-forge.jar )
[[ "$(sha256sum "$FRESH/mods/$(basename "$RELEASE_JAR")" | awk '{print $1}')" = "$FIRST_SHA" ]]
PACKAGE_SMOKE="$PORT/gazebos-package-server-smoke.log"; : > "$PACKAGE_SMOKE"
( cd "$FRESH" && exec ./run.sh nogui ) > "$PACKAGE_SMOKE" 2>&1 & ACTIVE_PID=$!; DEADLINE=$((SECONDS+150)); PASS=0
while ((SECONDS<DEADLINE)); do
  LOG="$FRESH/logs/latest.log"; FILES=("$PACKAGE_SMOKE"); [[ -f "$LOG" ]] && FILES+=("$LOG")
  if grep -Eiq 'ModLoadingException|Failed to create mod instance|NoClassDefFoundError|ClassNotFoundException|MixinApplyError|InvalidMixinException|Exception in server tick loop|The game crashed' "${FILES[@]}"; then stop_tree "$ACTIVE_PID"; ACTIVE_PID=""; cat "${FILES[@]}"; exit 1; fi
  if [[ -f "$LOG" ]] && grep -Eq 'Done \([0-9.]+s\)!' "$LOG"; then PASS=1; break; fi
  if ! kill -0 "$ACTIVE_PID" 2>/dev/null; then wait "$ACTIVE_PID" || true; ACTIVE_PID=""; cat "${FILES[@]}"; exit 1; fi
  sleep 1
done
[[ "$PASS" -eq 1 ]]; stop_tree "$ACTIVE_PID"; ACTIVE_PID=""
echo '[Gazebos] CANONICAL_PACKAGED_SERVER_PASS'

if [[ "$GAZEBOS_EXPECTED_JAR_SHA" = '__CAPTURE_AFTER_FIRST_GREEN__' || "$GAZEBOS_EXPECTED_SOURCE_SHA" = '__CAPTURE_AFTER_FIRST_GREEN__' ]]; then
  echo "[Gazebos] GAZEBOS_FIRST_GREEN_CAPTURE jar=$FIRST_SHA source=$SOURCE_SHA"
else
  [[ "$FIRST_SHA" = "$GAZEBOS_EXPECTED_JAR_SHA" ]]
  [[ "$SOURCE_SHA" = "$GAZEBOS_EXPECTED_SOURCE_SHA" ]]
  echo "[Gazebos] CROSS_RUN_RELEASE_IDENTITY_PASS jar=$FIRST_SHA source=$SOURCE_SHA"
  echo '[Gazebos] GAZEBOS_GRADUATION_PASS'
fi
