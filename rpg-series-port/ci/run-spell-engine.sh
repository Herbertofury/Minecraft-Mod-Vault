#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
UP="$ROOT/.rpg-upstream"
PORT="$ROOT/rpg-series-port/spell-engine-forge-1.20.1"
WORK="$ROOT/.spell-engine-build"
BASE_SHA=8721120169ddefd230fc73fc7c332318a92f6c7c
TARGET_SHA=bc02f7a49da950503010020da491f6bdc5871df7
SPELL_POWER_BASE=681993d5f823aa96b1b24e21b145e89f46147f2d
SPELL_POWER_TARGET=6fed879e796cbe82c43684d914a8fa99a99e8b12
RANGED_BASE=d95ba51c2f5c35bc8d397057092ba6043b00b705
RANGED_TARGET=c834f2699faefbdfcefa84f7f45708cd1a6bc55a
ACTIVE_PID=""

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
  for _ in {1..20}; do kill -0 "$root" 2>/dev/null || break; sleep 0.1; done
  mapfile -t children < <(descendants "$root")
  ((${#children[@]})) && kill -KILL "${children[@]}" 2>/dev/null || true
  kill -KILL "$root" 2>/dev/null || true
  wait "$root" 2>/dev/null || true
}
cleanup() { [[ -n "${ACTIVE_PID:-}" ]] && stop_tree "$ACTIVE_PID" || true; }
trap cleanup EXIT INT TERM

rm -rf "$UP" "$WORK"
mkdir -p "$UP"
clone_exact() {
  local repo="$1" sha="$2" dest="$3"
  git init -q "$dest"; git -C "$dest" remote add origin "https://github.com/${repo}.git"
  git -C "$dest" fetch -q --depth=1 origin "$sha"; git -C "$dest" checkout -q --detach FETCH_HEAD
  test "$(git -C "$dest" rev-parse HEAD)" = "$sha"
}
clone_exact ZsoltMolnarrr/SpellEngine "$BASE_SHA" "$UP/spell-engine-1201" & P1=$!
clone_exact ZsoltMolnarrr/SpellEngine "$TARGET_SHA" "$UP/spell-engine-1102" & P2=$!
clone_exact ZsoltMolnarrr/SpellPower "$SPELL_POWER_BASE" "$UP/spell-power-1201" & P3=$!
clone_exact ZsoltMolnarrr/SpellPower "$SPELL_POWER_TARGET" "$UP/spell-power-160" & P4=$!
clone_exact FabricExtras/RangedWeaponAPI "$RANGED_BASE" "$UP/ranged-1201" & P5=$!
clone_exact FabricExtras/RangedWeaponAPI "$RANGED_TARGET" "$UP/ranged-234" & P6=$!
wait "$P1" "$P2" "$P3" "$P4" "$P5" "$P6"

SPELL_POWER="$ROOT/rpg-series-port/spell_power-forge-1.20.1"
python "$SPELL_POWER/tools/prepare_upstream_source.py" "$UP/spell-power-1201" "$UP/spell-power-160" "$SPELL_POWER/common"
test -f "$SPELL_POWER/common/src/generatedUpstream/java/net/spell_power/api/SpellSchool.java"
SP_SOURCES="$SPELL_POWER/common/src/main/java:$SPELL_POWER/common/src/generatedUpstream/java"

RANGED="$ROOT/rpg-series-port/ranged-weapon-api-forge-1.20.1"
python "$RANGED/tools/prepare_upstream_source.py" "$UP/ranged-1201" "$UP/ranged-234" "$RANGED/common"
test -f "$RANGED/common/src/main/java/net/fabric_extras/ranged_weapon/api/RangedConfig.java"
RW_COMPILE="$UP/ranged-compile"
mkdir -p "$RW_COMPILE/main" "$RW_COMPILE/generated"
cp -a "$RANGED/common/src/main/java/." "$RW_COMPILE/main/"
cp -a "$RANGED/common/src/generatedUpstream/java/." "$RW_COMPILE/generated/"
rm -rf "$RW_COMPILE/main/net/fabric_extras/ranged_weapon/compat/emi"
RW_SOURCES="$RW_COMPILE/main:$RW_COMPILE/generated"
export SPELL_POWER_SOURCE_DIRS="$SP_SOURCES"
export RANGED_SOURCE_DIRS="$RW_SOURCES"

python "$PORT/tools/prepare_spell_engine.py" "$UP/spell-engine-1201" "$UP/spell-engine-1102" "$WORK"
python "$PORT/tools/compat_pass_1.py" "$WORK"
for part in a1 a2 b1 b2 c d; do python "$PORT/tools/compat_pass_2${part}.py" "$WORK" "$UP/spell-engine-1201"; done
python "$PORT/tools/compat_pass_3.py" "$WORK" "$UP/spell-engine-1201"
python "$PORT/tools/compat_pass_4a.py" "$WORK" "$UP/spell-engine-1201"
python "$PORT/tools/compat_pass_4b.py" "$WORK" "$UP/spell-engine-1201"
for part in a b c d e f; do python "$PORT/tools/compat_pass_5${part}.py" "$WORK" "$UP/spell-engine-1201"; done
python "$PORT/tools/compat_pass_6a.py" "$WORK" "$UP/spell-engine-1201"
python "$PORT/tools/compat_pass_6a1.py" "$WORK" "$UP/spell-engine-1201"
for part in b c d e f g; do python "$PORT/tools/compat_pass_6${part}.py" "$WORK" "$UP/spell-engine-1201"; done

test "$(find "$WORK/common/src/main/java" -name '*.java' | wc -l)" -ge 345
test -f "$WORK/forge/src/main/resources/META-INF/mods.toml"
rm -f "$ROOT/spell-engine-1.10.2-forge-1.20.1-source-ci.zip"
(cd "$WORK" && zip -qr "$ROOT/spell-engine-1.10.2-forge-1.20.1-source-ci.zip" . -x '*/build/*' '*/run/*' '.gradle/*')
unzip -t "$ROOT/spell-engine-1.10.2-forge-1.20.1-source-ci.zip" >/dev/null

gradle --no-daemon --stacktrace -p "$WORK" :forge:build
JAR="$(find "$WORK/forge/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*dev-shadow*' ! -name '*sources*' | head -n 1)"
test -n "$JAR"
unzip -t "$JAR" >/dev/null
OUT_JAR="$PORT/spell_engine-forge-1.10.2+1.20.1.jar"
cp "$JAR" "$OUT_JAR"
sha256sum "$OUT_JAR" | tee "$PORT/spell-engine-forge-1.20.1.sha256"
unzip -l "$OUT_JAR" | grep -F 'META-INF/jars/mixinextras-forge-0.4.1.jar'

rm -rf "$WORK/forge/run/logs"
CLIENT_LOG="$PORT/forge-client-smoke.log"
: > "$CLIENT_LOG"
xvfb-run -a gradle --no-daemon -p "$WORK" :forge:runClient > "$CLIENT_LOG" 2>&1 &
ACTIVE_PID=$!
PID=$ACTIVE_PID
DEADLINE=$((SECONDS+180))
READY=0
PASS=0
FATAL='MixinApplyError|InvalidMixinException|MixinTransformerError|Failed to create mod instance|NoClassDefFoundError|ClassNotFoundException|The game crashed whilst initializing game|Exception in thread "Render thread"'
while ((SECONDS<DEADLINE)); do
  LOG="$WORK/forge/run/logs/latest.log"
  FILES=("$CLIENT_LOG"); [[ -f "$LOG" ]] && FILES+=("$LOG")
  if grep -Eiq "$FATAL" "${FILES[@]}"; then stop_tree "$PID"; ACTIVE_PID=""; cat "${FILES[@]}"; exit 1; fi
  if [[ -f "$LOG" ]] && grep -Fq 'Reloading ResourceManager' "$LOG" && grep -Fq 'Backend library: LWJGL' "$LOG"; then
    [[ "$READY" -eq 0 ]] && READY=$SECONDS
    if ((SECONDS-READY>=8)); then PASS=1; break; fi
  fi
  if ! kill -0 "$PID" 2>/dev/null; then wait "$PID" || true; ACTIVE_PID=""; cat "${FILES[@]}"; exit 1; fi
  sleep 1
done
[[ "$PASS" -eq 1 ]] || { stop_tree "$PID"; ACTIVE_PID=""; cat "$CLIENT_LOG"; exit 1; }
stop_tree "$PID"; ACTIVE_PID=""
echo "[Spell Engine CI] Forge client bootstrap passed; JAR: $OUT_JAR"
