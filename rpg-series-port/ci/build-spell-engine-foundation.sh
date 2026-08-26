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

# Independent exact pins are fetched concurrently. This is the same deterministic reconstruction
# path as the graduated verifier, but intentionally stops after the verified build/package boundary.
clone_exact ZsoltMolnarrr/SpellEngine "$BASE_SHA" "$UP/spell-engine-1201" & p1=$!
clone_exact ZsoltMolnarrr/SpellEngine "$TARGET_SHA" "$UP/spell-engine-1102" & p2=$!
clone_exact ZsoltMolnarrr/SpellPower "$SPELL_POWER_BASE" "$UP/spell-power-1201" & p3=$!
clone_exact ZsoltMolnarrr/SpellPower "$SPELL_POWER_TARGET" "$UP/spell-power-160" & p4=$!
clone_exact FabricExtras/RangedWeaponAPI "$RANGED_BASE" "$UP/ranged-1201" & p5=$!
clone_exact FabricExtras/RangedWeaponAPI "$RANGED_TARGET" "$UP/ranged-234" & p6=$!
wait "$p1" "$p2" "$p3" "$p4" "$p5" "$p6"

SPELL_POWER="$ROOT/rpg-series-port/spell_power-forge-1.20.1"
python3 "$SPELL_POWER/tools/prepare_upstream_source.py" "$UP/spell-power-1201" "$UP/spell-power-160" "$SPELL_POWER/common"
gradle --no-daemon --stacktrace -p "$SPELL_POWER" :forge:build
SPELL_POWER_COMMON_JAR="$(find "$SPELL_POWER/common/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' | sort | head -n1)"
SPELL_POWER_FORGE_JAR="$(find "$SPELL_POWER/forge/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*dev-shadow*' ! -name '*sources*' | sort | head -n1)"
test -f "$SPELL_POWER_COMMON_JAR" -a -f "$SPELL_POWER_FORGE_JAR"
unzip -tq "$SPELL_POWER_FORGE_JAR"

RANGED="$ROOT/rpg-series-port/ranged-weapon-api-forge-1.20.1"
python3 "$RANGED/tools/prepare_upstream_source.py" "$UP/ranged-1201" "$UP/ranged-234" "$RANGED/common"
gradle --no-daemon --stacktrace -p "$RANGED" :forge:build
RANGED_COMMON_JAR="$(find "$RANGED/common/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' | sort | head -n1)"
RANGED_FORGE_JAR="$(find "$RANGED/forge/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*dev-shadow*' ! -name '*sources*' | sort | head -n1)"
test -f "$RANGED_COMMON_JAR" -a -f "$RANGED_FORGE_JAR"
unzip -tq "$RANGED_FORGE_JAR"

export SPELL_POWER_COMMON_JAR RANGED_COMMON_JAR SPELL_POWER_FORGE_JAR RANGED_FORGE_JAR
unset SPELL_POWER_SOURCE_DIRS RANGED_SOURCE_DIRS || true

python3 "$PORT/tools/prepare_spell_engine.py" "$UP/spell-engine-1201" "$UP/spell-engine-1102" "$WORK"
python3 "$PORT/tools/compat_pass_1.py" "$WORK"
for part in a1 a2 b1 b2 c d; do python3 "$PORT/tools/compat_pass_2${part}.py" "$WORK" "$UP/spell-engine-1201"; done
python3 "$PORT/tools/compat_pass_3.py" "$WORK" "$UP/spell-engine-1201"
python3 "$PORT/tools/compat_pass_4a.py" "$WORK" "$UP/spell-engine-1201"
python3 "$PORT/tools/compat_pass_4b.py" "$WORK" "$UP/spell-engine-1201"
for part in a b c d e f; do python3 "$PORT/tools/compat_pass_5${part}.py" "$WORK" "$UP/spell-engine-1201"; done
python3 "$PORT/tools/compat_pass_6a.py" "$WORK" "$UP/spell-engine-1201"
python3 "$PORT/tools/compat_pass_6a1.py" "$WORK" "$UP/spell-engine-1201"
for part in b c d e f g h i; do python3 "$PORT/tools/compat_pass_6${part}.py" "$WORK" "$UP/spell-engine-1201"; done

# Preserve the same anti-source-injection boundary as the graduated verifier.
test "$(find "$WORK/common/src/main/java" -name '*.java' | wc -l)" -ge 345
test -f "$WORK/forge/src/main/resources/META-INF/mods.toml"
if grep -R -nE 'SPELL_POWER_SOURCE_DIRS|RANGED_SOURCE_DIRS|addExternalCompileSources' "$WORK/common/build.gradle" "$WORK/forge/build.gradle"; then
  echo '[Spell Engine foundation] dependency source injection leaked into generated build' >&2
  exit 1
fi

gradle --no-daemon --stacktrace -p "$WORK" :forge:build
JAR="$(find "$WORK/forge/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*dev-shadow*' ! -name '*sources*' | sort | head -n1)"
test -f "$JAR"
unzip -tq "$JAR"
OUT_JAR="$PORT/spell_engine-forge-1.10.2+1.20.1.jar"
cp "$JAR" "$OUT_JAR"
sha256sum "$OUT_JAR" | tee "$PORT/spell-engine-forge-1.20.1.sha256"
if unzip -Z1 "$OUT_JAR" | grep -Eq '^(net/spell_power/|net/fabric_extras/ranged_weapon/)'; then
  echo '[Spell Engine foundation] separate RPG dependency classes leaked into release JAR' >&2
  exit 1
fi
if unzip -Z1 "$OUT_JAR" | grep -Ei '^META-INF/jars/.*(spell.?power|ranged.?weapon)'; then
  echo '[Spell Engine foundation] separate RPG dependency JAR embedded into Spell Engine' >&2
  exit 1
fi
unzip -p "$OUT_JAR" META-INF/mods.toml | grep -F 'modId="spell_power"'
echo '[Spell Engine foundation] deterministic build/package boundary green; runtime replay intentionally skipped for downstream lane.'
