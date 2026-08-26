#!/usr/bin/env bash
set -euo pipefail

ROOT="$(pwd)"
PORT="$ROOT/rpg-series-port/jewelry-forge-1.20.1/generated"
TOOLS="$ROOT/rpg-series-port/jewelry-forge-1.20.1/tools"
WORK="${RUNNER_TEMP:-/tmp}/jewelry-port"
LOG="$ROOT/rpg-series-port/jewelry-forge-1.20.1/jewelry-smoke.log"
SOURCE_ZIP="$ROOT/jewelry-2.4.0-forge-1.20.1-source-ci.zip"
BASE_SHA="f20b7d94c4c6cdd5a4ed26e4066374b64654fb96"
TARGET_SHA="572cb8759d13075b97e7a1acd969a6203db594cb"
SPELL_POWER_BASE="681993d5f823aa96b1b24e21b145e89f46147f2d"
SPELL_POWER_TARGET="6fed879e796cbe82c43684d914a8fa99a99e8b12"
RANGED_BASE="d95ba51c2f5c35bc8d397057092ba6043b00b705"
RANGED_TARGET="c834f2699faefbdfcefa84f7f45708cd1a6bc55a"

rm -rf "$WORK" "$PORT" "$SOURCE_ZIP"
mkdir -p "$WORK" "$(dirname "$PORT")"
: > "$LOG"
exec > >(tee -a "$LOG") 2>&1

echo "[Jewelry] Native Forge 1.20.1 port lane"
echo "[Jewelry] substrate=$BASE_SHA"
echo "[Jewelry] target=$TARGET_SHA"

clone_exact() {
    local repo="$1" sha="$2" dst="$3"
    git init -q "$dst"
    git -C "$dst" remote add origin "https://github.com/$repo.git"
    git -C "$dst" fetch -q --depth=1 origin "$sha"
    git -C "$dst" checkout -q --detach FETCH_HEAD
    test "$(git -C "$dst" rev-parse HEAD)" = "$sha"
}

clone_exact ZsoltMolnarrr/Jewelry "$BASE_SHA" "$WORK/base" & p1=$!
clone_exact ZsoltMolnarrr/Jewelry "$TARGET_SHA" "$WORK/target" & p2=$!
clone_exact ZsoltMolnarrr/SpellPower "$SPELL_POWER_BASE" "$WORK/spell-power-base" & p3=$!
clone_exact ZsoltMolnarrr/SpellPower "$SPELL_POWER_TARGET" "$WORK/spell-power-target" & p4=$!
clone_exact FabricExtras/RangedWeaponAPI "$RANGED_BASE" "$WORK/ranged-base" & p5=$!
clone_exact FabricExtras/RangedWeaponAPI "$RANGED_TARGET" "$WORK/ranged-target" & p6=$!
wait "$p1" "$p2" "$p3" "$p4" "$p5" "$p6"

echo "[Jewelry] Building verified foundation: Structure Pool API 1.2.1"
STRUCTURE="$ROOT/rpg-series-port/structure_pool_api-forge-1.20.1"
gradle --no-daemon --stacktrace -p "$STRUCTURE" :forge:build
STRUCTURE_JAR="$(find "$STRUCTURE/forge/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' ! -name '*dev-shadow*' ! -name '*javadoc*' | sort | head -n1)"
STRUCTURE_COMMON_JAR="$(find "$STRUCTURE/common/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' ! -name '*javadoc*' | sort | head -n1)"
test -n "$STRUCTURE_JAR" && test -f "$STRUCTURE_JAR"
test -n "$STRUCTURE_COMMON_JAR" && test -f "$STRUCTURE_COMMON_JAR"

SPELL_POWER="$ROOT/rpg-series-port/spell_power-forge-1.20.1"
echo "[Jewelry] Preparing/building verified foundation: Spell Power 1.6.0"
python3 "$SPELL_POWER/tools/prepare_upstream_source.py" "$WORK/spell-power-base" "$WORK/spell-power-target" "$SPELL_POWER/common"
test -f "$SPELL_POWER/common/src/generatedUpstream/java/net/spell_power/api/SpellSchool.java"
gradle --no-daemon --stacktrace -p "$SPELL_POWER" :forge:build
SPELL_POWER_JAR="$(find "$SPELL_POWER/forge/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' ! -name '*dev-shadow*' ! -name '*javadoc*' | sort | head -n1)"
SPELL_POWER_COMMON_JAR="$(find "$SPELL_POWER/common/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' ! -name '*javadoc*' | sort | head -n1)"
test -n "$SPELL_POWER_JAR" && test -f "$SPELL_POWER_JAR"
test -n "$SPELL_POWER_COMMON_JAR" && test -f "$SPELL_POWER_COMMON_JAR"
unzip -tq "$SPELL_POWER_JAR"
unzip -p "$SPELL_POWER_JAR" META-INF/mods.toml | grep -F 'modId="spell_power"' >/dev/null

RANGED="$ROOT/rpg-series-port/ranged-weapon-api-forge-1.20.1"
echo "[Jewelry] Preparing/building verified foundation: Ranged Weapon API 2.3.4"
python3 "$RANGED/tools/prepare_upstream_source.py" "$WORK/ranged-base" "$WORK/ranged-target" "$RANGED/common"
test -f "$RANGED/common/src/main/java/net/fabric_extras/ranged_weapon/api/RangedConfig.java"
gradle --no-daemon --stacktrace -p "$RANGED" :forge:build
RANGED_JAR="$(find "$RANGED/forge/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' ! -name '*dev-shadow*' ! -name '*javadoc*' | sort | head -n1)"
RANGED_COMMON_JAR="$(find "$RANGED/common/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' ! -name '*javadoc*' | sort | head -n1)"
test -n "$RANGED_JAR" && test -f "$RANGED_JAR"
test -n "$RANGED_COMMON_JAR" && test -f "$RANGED_COMMON_JAR"
unzip -tq "$RANGED_JAR"
unzip -p "$RANGED_JAR" META-INF/mods.toml | grep -F 'modId="ranged_weapon_api"' >/dev/null

python3 "$TOOLS/prepare_port.py" "$WORK/base" "$WORK/target" "$PORT"
mkdir -p "$PORT/libs"
# Loader/runtime JARs stay distinct from named common compilation JARs. Never source-inject or
# shadow these foundations into Jewelry.
cp "$STRUCTURE_JAR" "$PORT/libs/structure-pool-api.jar"
cp "$SPELL_POWER_JAR" "$PORT/libs/spell-power.jar"
cp "$RANGED_JAR" "$PORT/libs/ranged-weapon-api.jar"
cp "$STRUCTURE_COMMON_JAR" "$PORT/libs/structure-pool-api-common.jar"
cp "$SPELL_POWER_COMMON_JAR" "$PORT/libs/spell-power-common.jar"
cp "$RANGED_COMMON_JAR" "$PORT/libs/ranged-weapon-api-common.jar"

python3 "$TOOLS/compat_pass_1.py" "$PORT"
python3 "$TOOLS/compat_pass_2.py" "$PORT"
python3 "$TOOLS/compat_pass_3.py" "$PORT"

test -f "$PORT/common/src/main/generated/assets/jewelry/models/item/diamond_ring.json"
test -f "$PORT/common/src/main/java/net/jewelry/items/JewelryItems.java"
python3 - "$PORT" <<'PY'
from pathlib import Path
import sys
root = Path(sys.argv[1])
items = (root / 'common/src/main/java/net/jewelry/items/JewelryItems.java').read_text().lower()
required = [
    'diamond_ring', 'unique_attack_ring', 'unique_crit_ring', 'unique_dex_ring',
    'unique_arcane_ring', 'unique_fire_ring', 'unique_frost_ring', 'unique_healing_ring',
    'unique_spell_ring', 'unique_tank_ring'
]
missing = [x for x in required if x not in items]
if missing:
    raise SystemExit('Current 2.4.0 Jewelry catalog gate missing: ' + ', '.join(missing))
lang_roots = [root / 'common/src/main/resources/assets/jewelry/lang', root / 'common/src/main/generated/assets/jewelry/lang']
langs = sorted({p.name for r in lang_roots if r.exists() for p in r.glob('*.json')})
print(f'[Jewelry] current catalog anchors present; language files={len(langs)}')
if len(langs) < 15:
    raise SystemExit(f'Expected current translation set, found only {len(langs)} language files')
PY

(
    cd "$PORT"
    zip -qr "$SOURCE_ZIP" . -x '.gradle/*' '*/build/*'
)
unzip -tq "$SOURCE_ZIP"

echo "[Jewelry] Compiling full current 2.4.0 source against Minecraft 1.20.1 mappings and native Forge foundations"
gradle --no-daemon --stacktrace -p "$PORT" :forge:remapJar

OUT_JAR="$(find "$PORT/forge/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' ! -name '*dev-shadow*' ! -name '*javadoc*' | sort | head -n1)"
test -f "$OUT_JAR"

for prefix in 'net/spell_power/' 'net/fabric_extras/ranged_weapon/' 'net/fabric_extras/structure_pool/'; do
    if unzip -Z1 "$OUT_JAR" | grep -q "^$prefix"; then
        echo "[Jewelry] ERROR: packaged dependency classes under $prefix" >&2
        exit 3
    fi
done

unzip -tq "$OUT_JAR"
sha256sum "$OUT_JAR" | tee "$ROOT/rpg-series-port/jewelry-forge-1.20.1/jewelry.sha256"
sha256sum "$SOURCE_ZIP" | tee "$ROOT/rpg-series-port/jewelry-forge-1.20.1/jewelry-source.sha256"
echo "[Jewelry] build/package gate passed: $OUT_JAR"
