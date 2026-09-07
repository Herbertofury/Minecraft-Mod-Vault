#!/usr/bin/env bash
set -euo pipefail

ROOT="${GITHUB_WORKSPACE:?}"
PORT="$ROOT/rpg-series-port/ranged-weapon-api-forge-1.20.1"
UP="$ROOT/.rpg-upstream"
ACTIVE_PID=""

# Kill a spawned Gradle/Minecraft process tree without depending on process-group
# identity. The previous setsid/group cleanup could wait forever after the actual
# smoke checks had already passed.
descendants() {
  local parent="$1" child
  while IFS= read -r child; do
    [[ -z "$child" ]] && continue
    descendants "$child"
    printf '%s\n' "$child"
  done < <(pgrep -P "$parent" 2>/dev/null || true)
}

stop_tree() {
  local root="${1:-}"
  [[ -n "$root" ]] || return 0
  local -a children=()
  mapfile -t children < <(descendants "$root")
  ((${#children[@]})) && kill -TERM "${children[@]}" 2>/dev/null || true
  kill -TERM "$root" 2>/dev/null || true
  for _ in {1..20}; do
    kill -0 "$root" 2>/dev/null || break
    sleep 0.1
  done
  mapfile -t children < <(descendants "$root")
  ((${#children[@]})) && kill -KILL "${children[@]}" 2>/dev/null || true
  kill -KILL "$root" 2>/dev/null || true
  wait "$root" 2>/dev/null || true
}

cleanup() {
  if [[ -n "${ACTIVE_PID:-}" ]]; then
    stop_tree "$ACTIVE_PID"
    ACTIVE_PID=""
  fi
}
trap cleanup EXIT INT TERM

rm -rf "$UP"
mkdir -p "$UP"

clone_exact() {
  local repo="$1" sha="$2" dest="$3"
  git init -q "$dest"
  git -C "$dest" remote add origin "https://github.com/${repo}.git"
  git -C "$dest" fetch -q --depth=1 origin "$sha"
  git -C "$dest" checkout -q --detach FETCH_HEAD
  test "$(git -C "$dest" rev-parse HEAD)" = "$sha"
}

clone_exact FabricExtras/RangedWeaponAPI d95ba51c2f5c35bc8d397057092ba6043b00b705 "$UP/rwa120"
clone_exact FabricExtras/RangedWeaponAPI c834f2699faefbdfcefa84f7f45708cd1a6bc55a "$UP/rwa234"

cd "$PORT"
python tools/prepare_upstream_source.py "$UP/rwa120" "$UP/rwa234" common

test ! -f common/src/generatedUpstream/java/net/fabric_extras/ranged_weapon/api/CrossbowMechanics.java
test ! -f common/src/generatedUpstream/java/net/fabric_extras/ranged_weapon/RangedWeaponMod.java
test ! -f common/src/generatedUpstream/java/net/fabric_extras/ranged_weapon/mixin/PersistentProjectileEntityMixin.java
test ! -f common/src/generatedUpstream/resources/fabric.mod.json
test "$(find common/src/generatedUpstream/resources/assets/ranged_weapon/lang -name '*.json' | wc -l)" -ge 20
if grep -R 'net.fabricmc' common/src/main/java common/src/generatedUpstream/java; then
  echo 'Fabric loader/API reference leaked into native Forge sources' >&2
  exit 1
fi

SOURCE_ZIP="$ROOT/ranged-weapon-api-forge-1.20.1-source-ci.zip"
rm -f "$SOURCE_ZIP"
python3 - "$PORT" "$SOURCE_ZIP" <<'PY_SOURCE'
from pathlib import Path
import stat, sys, zipfile
src=Path(sys.argv[1]).resolve(); out=Path(sys.argv[2]).resolve()
skip={'.gradle','build','run','runs','.git','.fresh-forge-server'}
files=[]
for path in src.rglob('*'):
    rel=path.relative_to(src)
    if any(part in skip for part in rel.parts):
        continue
    if len(rel.parts)==1 and (rel.name.endswith('.sha256') or rel.name.endswith('-smoke.log')):
        continue
    if path.is_file():
        files.append((rel.as_posix(),path))
with zipfile.ZipFile(out,'w',compression=zipfile.ZIP_DEFLATED,compresslevel=9) as zf:
    for arcname,path in sorted(files):
        info=zipfile.ZipInfo(arcname,date_time=(1980,1,1,0,0,0))
        info.compress_type=zipfile.ZIP_DEFLATED
        info.external_attr=(stat.S_IFREG|0o644)<<16
        info.create_system=3
        zf.writestr(info,path.read_bytes(),compress_type=zipfile.ZIP_DEFLATED,compresslevel=9)
PY_SOURCE
unzip -tq "$SOURCE_ZIP" >/dev/null
SOURCE_SHA="$(sha256sum "$SOURCE_ZIP" | awk '{print $1}')"
printf '%s  %s\n' "$SOURCE_SHA" "$SOURCE_ZIP" | tee ranged-weapon-api-source.sha256
echo "[Ranged Weapon API CI] DETERMINISTIC_SOURCE_PACKAGE_PASS sha256=$SOURCE_SHA"

gradle --no-daemon --stacktrace :forge:build

JAR=$(find forge/build/libs -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' ! -name '*dev-shadow*' | head -1)
test -n "$JAR"
JAR=$(realpath "$JAR")
unzip -t "$JAR"
unzip -p "$JAR" META-INF/mods.toml | grep -F 'modId="ranged_weapon_api"'
unzip -p "$JAR" META-INF/mods.toml | grep -F 'versionRange="[1.20.1,1.20.2)"'
unzip -l "$JAR" | grep -F 'net/fabric_extras/ranged_weapon/forge/ForgeMod.class'
unzip -l "$JAR" | grep -F 'net/fabric_extras/ranged_weapon/api/RangedConfig.class'
unzip -l "$JAR" | grep -F 'net/fabric_extras/ranged_weapon/api/AttributeModifierIDs.class'
unzip -l "$JAR" | grep -F 'net/fabric_extras/ranged_weapon/mixin/PersistentProjectileEntityMixin.class'
unzip -l "$JAR" | grep -F 'net/fabric_extras/ranged_weapon/mixin/item/ProjectileUtilMixin.class'
unzip -l "$JAR" | grep -F 'net/fabric_extras/ranged_weapon/compat/emi/RangedWeaponEmiPlugin.class'
unzip -l "$JAR" | grep -F 'assets/ranged_weapon/lang/en_us.json'
unzip -l "$JAR" | grep -F 'META-INF/jars/mixinextras-forge-0.4.1.jar'
if unzip -l "$JAR" | grep -q 'fabric.mod.json\|META-INF/neoforge.mods.toml'; then
  echo 'Non-Forge metadata leaked into final JAR' >&2
  exit 1
fi
python3 - "$JAR" <<'PY_PACK'
import json,sys,zipfile
jar=sys.argv[1]
with zipfile.ZipFile(jar) as z:
    names=set(z.namelist())
    if 'pack.mcmeta' not in names:
        raise SystemExit('[Ranged Weapon API CI] production pack.mcmeta missing')
    meta=json.loads(z.read('pack.mcmeta'))
    fmt=meta.get('pack',{}).get('pack_format')
    if fmt != 15:
        raise SystemExit(f'[Ranged Weapon API CI] production pack format drifted: {fmt} != 15')
    langs=sorted(n for n in names if n.startswith('assets/ranged_weapon/lang/') and n.endswith('.json'))
    if len(langs) < 20 or 'assets/ranged_weapon/lang/en_us.json' not in names:
        raise SystemExit(f'[Ranged Weapon API CI] production language resource contract failed: {len(langs)} language files')
    print(f'[Ranged Weapon API CI] PRODUCTION_RESOURCE_PACK_CONTRACT_PASS pack_format={fmt} languages={len(langs)}')
PY_PACK
sha256sum "$JAR" | tee ranged-weapon-api-forge.sha256

# Client proof: load the current source in Forge, reach LWJGL + resource reload,
# then keep it alive briefly to surface delayed mixin/resource failures.
rm -rf forge/run/logs
: > forge-client-smoke.log
xvfb-run -a gradle --no-daemon :forge:runClient > forge-client-smoke.log 2>&1 &
ACTIVE_PID=$!
PID=$ACTIVE_PID
DEADLINE=$((SECONDS+180))
READY=0
PASS=0
FATAL='MixinApplyError|InvalidMixinException|MixinTransformerError|Failed to create mod instance|NoClassDefFoundError.*ranged_weapon|ClassNotFoundException.*ranged_weapon|Using missing texture.*ranged_weapon|The game crashed whilst initializing game'
while ((SECONDS<DEADLINE)); do
  LOG=$(find forge/run -type f -path '*/logs/latest.log' | head -1 || true)
  FILES=(forge-client-smoke.log); [[ -n "$LOG" ]] && FILES+=("$LOG")
  if grep -Eiq "$FATAL" "${FILES[@]}"; then
    stop_tree "$PID"; ACTIVE_PID=""; cat "${FILES[@]}"; exit 1
  fi
  if [[ -n "$LOG" ]] && grep -Fq 'Reloading ResourceManager' "$LOG" && grep -Fq 'Backend library: LWJGL' "$LOG"; then
    [[ "$READY" -eq 0 ]] && READY=$SECONDS
    if ((SECONDS-READY>=8)); then PASS=1; break; fi
  fi
  if ! kill -0 "$PID" 2>/dev/null; then
    wait "$PID" || true
    ACTIVE_PID=""
    cat "${FILES[@]}"
    exit 1
  fi
  sleep 1
done
[[ "$PASS" -eq 1 ]] || { stop_tree "$PID"; ACTIVE_PID=""; cat forge-client-smoke.log; exit 1; }
stop_tree "$PID"
ACTIVE_PID=""

# Strongest server proof: install a clean Forge 47.4.23 server and load only the
# packaged release JAR. This supersedes the redundant dev-server boot while still
# executing the full runtime self-test against the distributable artifact.
FRESH="$PORT/.fresh-forge-server"
rm -rf "$FRESH"
mkdir -p "$FRESH/mods"
curl -fsSL "https://maven.minecraftforge.net/net/minecraftforge/forge/1.20.1-47.4.23/forge-1.20.1-47.4.23-installer.jar" -o "$FRESH/forge-installer.jar"
(
  cd "$FRESH"
  java -jar forge-installer.jar --installServer >/dev/null
  cp "$JAR" mods/
  printf 'eula=true\n' > eula.txt
  printf '%s\n' '-Xmx2G' '-DrangedWeapon.ci.selftest=true' > user_jvm_args.txt
  cat > server.properties <<'PROPS'
level-type=minecraft:flat
generate-structures=false
view-distance=3
simulation-distance=3
spawn-protection=0
PROPS
)
: > forge-package-smoke.log
(
  cd "$FRESH"
  exec ./run.sh nogui
) > "$PORT/forge-package-smoke.log" 2>&1 &
ACTIVE_PID=$!
PID=$ACTIVE_PID
DEADLINE=$((SECONDS+180))
PASS=0
FATAL='MixinApplyError|InvalidMixinException|MixinTransformerError|Failed to start the minecraft server|Failed to create mod instance|NoClassDefFoundError|ClassNotFoundException|Registry is already frozen|IllegalStateException: \[Ranged Weapon API CI\]|Exception in server tick loop'
while ((SECONDS<DEADLINE)); do
  LOG="$FRESH/logs/latest.log"
  FILES=(forge-package-smoke.log); [[ -f "$LOG" ]] && FILES+=("$LOG")
  if grep -Eiq "$FATAL" "${FILES[@]}"; then
    stop_tree "$PID"; ACTIVE_PID=""; cat "${FILES[@]}"; exit 1
  fi
  if grep -Fq '[Ranged Weapon API CI] Runtime self-test passed' forge-package-smoke.log && [[ -f "$LOG" ]] && grep -Eq 'Done \([0-9.]+s\)!' "$LOG"; then
    PASS=1
    break
  fi
  if ! kill -0 "$PID" 2>/dev/null; then
    wait "$PID" || true
    ACTIVE_PID=""
    cat "${FILES[@]}"
    exit 1
  fi
  sleep 1
done
[[ "$PASS" -eq 1 ]] || { stop_tree "$PID"; ACTIVE_PID=""; cat forge-package-smoke.log; exit 1; }
stop_tree "$PID"
ACTIVE_PID=""

echo '[Ranged Weapon API CI] Full build/package/client/fresh-packaged-JAR verification passed.'
