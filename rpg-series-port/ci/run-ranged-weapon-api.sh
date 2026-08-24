#!/usr/bin/env bash
set -euo pipefail

ROOT="${GITHUB_WORKSPACE:?}"
PORT="$ROOT/rpg-series-port/ranged-weapon-api-forge-1.20.1"
UP="$ROOT/.rpg-upstream"
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
test ! -f common/src/generatedUpstream/resources/fabric.mod.json
test "$(find common/src/generatedUpstream/resources/assets/ranged_weapon/lang -name '*.json' | wc -l)" -ge 20
if grep -R 'net.fabricmc' common/src/main/java common/src/generatedUpstream/java; then
  echo 'Fabric loader/API reference leaked into native Forge sources' >&2
  exit 1
fi

rm -f "$ROOT/ranged-weapon-api-forge-1.20.1-source-ci.zip"
zip -qr "$ROOT/ranged-weapon-api-forge-1.20.1-source-ci.zip" . -x '*/build/*' '*/run/*'

gradle --no-daemon --stacktrace :forge:build

JAR=$(find forge/build/libs -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' ! -name '*dev-shadow*' | head -1)
test -n "$JAR"
unzip -t "$JAR"
unzip -p "$JAR" META-INF/mods.toml | grep -F 'modId="ranged_weapon_api"'
unzip -p "$JAR" META-INF/mods.toml | grep -F 'versionRange="[1.20.1,1.20.2)"'
unzip -l "$JAR" | grep -F 'net/fabric_extras/ranged_weapon/forge/ForgeMod.class'
unzip -l "$JAR" | grep -F 'net/fabric_extras/ranged_weapon/api/RangedConfig.class'
unzip -l "$JAR" | grep -F 'net/fabric_extras/ranged_weapon/api/AttributeModifierIDs.class'
unzip -l "$JAR" | grep -F 'assets/ranged_weapon/lang/en_us.json'
if unzip -l "$JAR" | grep -q 'fabric.mod.json\|META-INF/neoforge.mods.toml'; then
  echo 'Non-Forge metadata leaked into final JAR' >&2
  exit 1
fi
sha256sum "$JAR" | tee ranged-weapon-api-forge.sha256

rm -rf forge/run/logs
mkdir -p forge/run
printf 'eula=true\n' > forge/run/eula.txt
: > forge-server-smoke.log
setsid gradle --no-daemon -PrangedWeaponCiSelfTest :forge:runServer > forge-server-smoke.log 2>&1 &
PID=$!
DEADLINE=$((SECONDS+180))
PASS=0
FATAL='MixinApplyError|InvalidMixinException|MixinTransformerError|Failed to start the minecraft server|Failed to create mod instance|NoClassDefFoundError|Registry is already frozen|IllegalStateException: \[Ranged Weapon API CI\]|Exception in server tick loop'
stop_group(){ kill -TERM -- -"$PID" 2>/dev/null||true; sleep 1; kill -KILL -- -"$PID" 2>/dev/null||true; wait "$PID" 2>/dev/null||true; }
while ((SECONDS<DEADLINE)); do
  LOG=$(find forge/run -type f -path '*/logs/latest.log'|head -1||true)
  FILES=(forge-server-smoke.log); [[ -n "$LOG" ]]&&FILES+=("$LOG")
  if grep -Eiq "$FATAL" "${FILES[@]}"; then stop_group; cat "${FILES[@]}"; exit 1; fi
  if grep -Fq '[Ranged Weapon API CI] Runtime self-test passed' forge-server-smoke.log && [[ -n "$LOG" ]] && grep -Eq 'Done \([0-9.]+s\)!' "$LOG"; then PASS=1; break; fi
  if ! kill -0 "$PID" 2>/dev/null; then wait "$PID"||true; cat "${FILES[@]}"; exit 1; fi
  sleep 1
done
[[ "$PASS" -eq 1 ]] || { stop_group; cat forge-server-smoke.log; exit 1; }
stop_group

rm -rf forge/run/logs
: > forge-client-smoke.log
setsid xvfb-run -a gradle --no-daemon :forge:runClient > forge-client-smoke.log 2>&1 &
PID=$!
DEADLINE=$((SECONDS+180))
READY=0
PASS=0
FATAL='MixinApplyError|InvalidMixinException|MixinTransformerError|Failed to create mod instance|NoClassDefFoundError.*ranged_weapon|Using missing texture.*ranged_weapon|The game crashed whilst initializing game'
while ((SECONDS<DEADLINE)); do
  LOG=$(find forge/run -type f -path '*/logs/latest.log'|head -1||true)
  FILES=(forge-client-smoke.log); [[ -n "$LOG" ]]&&FILES+=("$LOG")
  if grep -Eiq "$FATAL" "${FILES[@]}"; then stop_group; cat "${FILES[@]}"; exit 1; fi
  if [[ -n "$LOG" ]]&&grep -Fq 'Reloading ResourceManager' "$LOG"&&grep -Eq 'Backend library: LWJGL|OpenAL initialized' "$LOG"; then [[ "$READY" -eq 0 ]]&&READY=$SECONDS; ((SECONDS-READY>=8))&&{ PASS=1; break; }; fi
  if ! kill -0 "$PID" 2>/dev/null; then wait "$PID"||true; cat "${FILES[@]}"; exit 1; fi
  sleep 1
done
[[ "$PASS" -eq 1 ]] || { stop_group; cat forge-client-smoke.log; exit 1; }
stop_group

echo '[Ranged Weapon API CI] Full build/package/server/client verification passed.'
