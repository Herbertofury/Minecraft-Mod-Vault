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

# Fan in all immutable sources in parallel-sized groups, then assemble the two already-verified
# API foundations exactly as their release gates did. This produces the real API jars Spell Engine
# compiles against without publishing or rebuilding unrelated ports.
clone_exact ZsoltMolnarrr/SpellEngine "$BASE_SHA" "$UP/spell-engine-1201" & P1=$!
clone_exact ZsoltMolnarrr/SpellEngine "$TARGET_SHA" "$UP/spell-engine-1102" & P2=$!
clone_exact ZsoltMolnarrr/SpellPower "$SPELL_POWER_BASE" "$UP/spell-power-1201" & P3=$!
clone_exact ZsoltMolnarrr/SpellPower "$SPELL_POWER_TARGET" "$UP/spell-power-160" & P4=$!
clone_exact FabricExtras/RangedWeaponAPI "$RANGED_BASE" "$UP/ranged-1201" & P5=$!
clone_exact FabricExtras/RangedWeaponAPI "$RANGED_TARGET" "$UP/ranged-234" & P6=$!
wait "$P1" "$P2" "$P3" "$P4" "$P5" "$P6"

SPELL_POWER="$ROOT/rpg-series-port/spell_power-forge-1.20.1"
python "$SPELL_POWER/tools/prepare_upstream_source.py" "$UP/spell-power-1201" "$UP/spell-power-160" "$SPELL_POWER/common"
gradle --no-daemon -p "$SPELL_POWER" :common:jar

test -f "$SPELL_POWER/common/src/generatedUpstream/java/net/spell_power/api/SpellSchool.java"
RANGED="$ROOT/rpg-series-port/ranged-weapon-api-forge-1.20.1"
python "$RANGED/tools/prepare_upstream_source.py" "$UP/ranged-1201" "$UP/ranged-234" "$RANGED/common"
gradle --no-daemon -p "$RANGED" :common:jar

python "$PORT/tools/prepare_spell_engine.py" "$UP/spell-engine-1201" "$UP/spell-engine-1102" "$WORK"

rm -f "$ROOT/spell-engine-1.10.2-forge-1.20.1-source-ci.zip"
(
  cd "$WORK"
  zip -qr "$ROOT/spell-engine-1.10.2-forge-1.20.1-source-ci.zip" . -x '*/build/*' '*/run/*' '.gradle/*'
)
unzip -t "$ROOT/spell-engine-1.10.2-forge-1.20.1-source-ci.zip" >/dev/null

# Full 1.10.2 common surface. The high javac error cap is intentional: one failure corpus should
# expose whole API families so fixes can be batched instead of one-at-a-time CI whack-a-mole.
gradle --no-daemon --stacktrace -p "$WORK" :common:compileJava

echo '[Spell Engine CI] 1.10.2 common source compiles against Minecraft/Forge 1.20.1 foundations.'
