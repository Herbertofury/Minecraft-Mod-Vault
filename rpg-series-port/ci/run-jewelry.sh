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

# Fetch the immutable 1.20.1 compatibility substrate and complete 2.4.0 behavior/content target in parallel.
clone_exact ZsoltMolnarrr/Jewelry "$BASE_SHA" "$WORK/base" & p1=$!
clone_exact ZsoltMolnarrr/Jewelry "$TARGET_SHA" "$WORK/target" & p2=$!
wait "$p1" "$p2"

build_foundation() {
    local project="$1" label="$2"
    echo "[Jewelry] Building verified foundation: $label"
    gradle --no-daemon -p "$project" :forge:remapJar
    local jar
    jar="$(find "$project/forge/build/libs" -maxdepth 1 -type f -name '*.jar' \
        ! -name '*sources*' ! -name '*dev-shadow*' ! -name '*javadoc*' -printf '%p\n' | sort | tail -n1)"
    test -n "$jar" && test -f "$jar"
    printf '%s' "$jar"
}

STRUCTURE_JAR="$(build_foundation "$ROOT/rpg-series-port/structure_pool_api-forge-1.20.1" 'Structure Pool API 1.2.1' | tail -n1)"
SPELL_POWER_JAR="$(build_foundation "$ROOT/rpg-series-port/spell_power-forge-1.20.1" 'Spell Power 1.6.0' | tail -n1)"
RANGED_JAR="$(build_foundation "$ROOT/rpg-series-port/ranged-weapon-api-forge-1.20.1" 'Ranged Weapon API 2.3.4' | tail -n1)"

python3 "$TOOLS/prepare_port.py" "$WORK/base" "$WORK/target" "$PORT"
mkdir -p "$PORT/libs"
cp "$STRUCTURE_JAR" "$PORT/libs/structure-pool-api.jar"
cp "$SPELL_POWER_JAR" "$PORT/libs/spell-power.jar"
cp "$RANGED_JAR" "$PORT/libs/ranged-weapon-api.jar"

# Static whole-target gates before compilation: these catch accidental fallback to the tiny 1.3.7 catalog.
test -f "$PORT/common/src/main/generated/assets/jewelry/models/item/diamond_ring.json"
test -f "$PORT/common/src/main/java/net/jewelry/items/JewelryItems.java"
python3 - "$PORT" <<'PY'
from pathlib import Path
import sys
root = Path(sys.argv[1])
items = (root / 'common/src/main/java/net/jewelry/items/JewelryItems.java').read_text()
required = [
    'diamond_ring', 'attack_ring', 'critical_strike_ring', 'dexterity_ring',
    'arcane_ring', 'fire_ring', 'frost_ring', 'healing_ring', 'spell_ring', 'tank_ring'
]
missing = [x for x in required if x not in items.lower()]
if missing:
    raise SystemExit('Current 2.4.0 Jewelry catalog gate missing: ' + ', '.join(missing))
langs = list((root / 'common/src/main/resources/assets/jewelry/lang').glob('*.json'))
print(f'[Jewelry] current catalog anchors present; language files={len(langs)}')
if len(langs) < 15:
    raise SystemExit(f'Expected current translation set, found only {len(langs)} language files')
PY

# Preserve the exact generated state even when the first compatibility compile reveals an API delta.
(
    cd "$PORT"
    zip -qr "$SOURCE_ZIP" . -x '.gradle/*' '*/build/*'
)

echo "[Jewelry] Compiling full current 2.4.0 source against Minecraft 1.20.1 mappings and native Forge foundations"
gradle --no-daemon -p "$PORT" :forge:remapJar

OUT_JAR="$(find "$PORT/forge/build/libs" -maxdepth 1 -type f -name '*.jar' \
    ! -name '*sources*' ! -name '*dev-shadow*' ! -name '*javadoc*' -printf '%p\n' | sort | tail -n1)"
test -f "$OUT_JAR"

# Architecture gates: Jewelry must not vendor or shadow foundation classes.
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
