#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
UP="$ROOT/.rpg-upstream"
PORT="$ROOT/rpg-series-port/spell-engine-forge-1.20.1"
WORK="$ROOT/.spell-engine-build"
BASE_SHA=8721120169ddefd230fc73fc7c332318a92f6c7c
TARGET_SHA=bc02f7a49da950503010020da491f6bdc5871df7

rm -rf "$UP" "$WORK"
mkdir -p "$UP"
clone_exact() {
  local repo="$1" sha="$2" dest="$3"
  git init -q "$dest"
  git -C "$dest" remote add origin "https://github.com/${repo}.git"
  git -C "$dest" fetch -q --depth=1 origin "$sha"
  git -C "$dest" checkout -q --detach FETCH_HEAD
  test "$(git -C "$dest" rev-parse HEAD)" = "$sha"
}
clone_exact ZsoltMolnarrr/SpellEngine "$BASE_SHA" "$UP/spell-engine-1201"
clone_exact ZsoltMolnarrr/SpellEngine "$TARGET_SHA" "$UP/spell-engine-1102"

# Build the already-verified 1.20.1 API foundations as named/common compile dependencies.
gradle --no-daemon -p "$ROOT/rpg-series-port/spell_power-forge-1.20.1" :common:jar
gradle --no-daemon -p "$ROOT/rpg-series-port/ranged-weapon-api-forge-1.20.1" :common:jar

python "$PORT/tools/prepare_spell_engine.py" "$UP/spell-engine-1201" "$UP/spell-engine-1102" "$WORK"

rm -f "$ROOT/spell-engine-1.10.2-forge-1.20.1-source-ci.zip"
(
  cd "$WORK"
  zip -qr "$ROOT/spell-engine-1.10.2-forge-1.20.1-source-ci.zip" . -x '*/build/*' '*/run/*' '.gradle/*'
)
unzip -t "$ROOT/spell-engine-1.10.2-forge-1.20.1-source-ci.zip" >/dev/null

# First gate deliberately targets the full 1.10.2 common source surface. Once this is green,
# the same runner advances to the native Forge platform/package/runtime gates.
gradle --no-daemon --stacktrace -p "$WORK" :common:compileJava

echo '[Spell Engine CI] 1.10.2 common source compiles against Minecraft/Forge 1.20.1 foundations.'
