#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
PORT="$ROOT/rpg-series-port/wizards-forge-1.20.1"
GEN="$PORT/generated"
WORK="${RUNNER_TEMP:-/tmp}/wizards-port"
BASE_SHA=395ade75b50067c19f9b57a84c409bf962e09224
TARGET_SHA=82fd3a0f48366e6e406b4e7ca4b6d827a3793fb9
ARMOR_SHA=a664155a0aab3161cd7e4bf0c1f72512b4ec4949

rm -rf "$WORK" "$GEN"
mkdir -p "$WORK"

clone_exact() {
  local repo="$1" sha="$2" dst="$3"
  git init -q "$dst"
  git -C "$dst" remote add origin "https://github.com/$repo.git"
  git -C "$dst" fetch -q --depth=1 origin "$sha"
  git -C "$dst" checkout -q --detach FETCH_HEAD
  test "$(git -C "$dst" rev-parse HEAD)" = "$sha"
}

echo '[Wizards] Fetching exact Wizards pins + Armor Model API in parallel'
clone_exact ZsoltMolnarrr/Wizards "$BASE_SHA" "$WORK/wizards-base" & p1=$!
clone_exact ZsoltMolnarrr/Wizards "$TARGET_SHA" "$WORK/wizards-target" & p2=$!
clone_exact FabricExtras/ArmorModelAPI "$ARMOR_SHA" "$WORK/armor-target" & p3=$!
wait "$p1" "$p2" "$p3"

echo '[Wizards] Auditing current upstream dependency declarations against the Forge 1.20.1 compatibility ledger'
python3 "$ROOT/rpg-series-port/ci/audit-dependency-compat.py" \
  "$ROOT/rpg-series-port/dependency-compatibility.json" \
  wizards \
  "$WORK/wizards-target"

echo '[Wizards] Reconstructing graduated Spell Engine foundation and compile dependencies (build-only fast path)'
bash "$ROOT/rpg-series-port/ci/build-spell-engine-foundation.sh"

echo '[Wizards] Building graduated Structure Pool API and Runes foundations'
STRUCTURE="$ROOT/rpg-series-port/structure_pool_api-forge-1.20.1"
RUNES="$ROOT/rpg-series-port/runes-forge-1.20.1"
gradle --no-daemon --stacktrace -p "$STRUCTURE" :forge:build
gradle --no-daemon --stacktrace -p "$RUNES" :forge:build

echo '[Wizards] Reconstructing/building graduated Armor Model API foundation without rerunning its already-sealed runtime suite'
ARMOR="$ROOT/rpg-series-port/armor-model-api-forge-1.20.1"
python3 "$ARMOR/tools/prepare_port.py" "$WORK/armor-target" "$ARMOR/generated"
gradle --no-daemon --stacktrace -p "$ARMOR/generated" :forge:remapJar

echo '[Wizards] Reusing TinyConfig 3.1.0 foundation built with Spell Engine API parity'
TINY="$ROOT/rpg-series-port/tiny-config-forge-1.20.1"

SPELL_POWER="$ROOT/rpg-series-port/spell_power-forge-1.20.1"
SPELL_ENGINE_WORK="$ROOT/.spell-engine-build"
SPELL_ENGINE_PORT="$ROOT/rpg-series-port/spell-engine-forge-1.20.1"

pick_jar() {
  find "$1" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' ! -name '*dev-shadow*' ! -name '*javadoc*' | sort | head -n1
}

STRUCTURE_COMMON="$(pick_jar "$STRUCTURE/common/build/libs")"
STRUCTURE_FORGE="$(pick_jar "$STRUCTURE/forge/build/libs")"
RUNES_COMMON="$(pick_jar "$RUNES/common/build/libs")"
RUNES_FORGE="$(pick_jar "$RUNES/forge/build/libs")"
SPELL_POWER_COMMON="$(pick_jar "$SPELL_POWER/common/build/libs")"
SPELL_POWER_FORGE="$(pick_jar "$SPELL_POWER/forge/build/libs")"
SPELL_ENGINE_COMMON="$(pick_jar "$SPELL_ENGINE_WORK/common/build/libs")"
SPELL_ENGINE_FORGE="$SPELL_ENGINE_PORT/spell_engine-forge-1.10.2+1.20.1.jar"
ARMOR_COMMON="$(pick_jar "$ARMOR/generated/common/build/libs")"
ARMOR_FORGE="$(pick_jar "$ARMOR/generated/forge/build/libs")"
TINY_COMMON="$(pick_jar "$TINY/generated/common/build/libs")"
TINY_FORGE="$(pick_jar "$TINY/generated/forge/build/libs")"
for jar in "$STRUCTURE_COMMON" "$STRUCTURE_FORGE" "$RUNES_COMMON" "$RUNES_FORGE" \
           "$SPELL_POWER_COMMON" "$SPELL_POWER_FORGE" "$SPELL_ENGINE_COMMON" "$SPELL_ENGINE_FORGE" \
           "$ARMOR_COMMON" "$ARMOR_FORGE" "$TINY_COMMON" "$TINY_FORGE"; do
  test -n "$jar" && test -f "$jar"
  unzip -tq "$jar"
done

# Runes is a graduated runtime foundation. Its mixin class, exact config resource and manifest
# declaration must ship together; a compile-only success is not sufficient for downstream mods.
unzip -p "$RUNES_FORGE" runes.mixins.json | grep -F '"package": "net.runes.mixin"' >/dev/null
unzip -p "$RUNES_FORGE" runes.mixins.json | grep -F '"PlayerEntityMixin"' >/dev/null
unzip -Z1 "$RUNES_FORGE" | grep -F 'net/runes/mixin/PlayerEntityMixin.class' >/dev/null
unzip -p "$RUNES_FORGE" META-INF/MANIFEST.MF | tr -d '\r' | grep -F 'MixinConfigs: runes.mixins.json' >/dev/null
echo '[Wizards] Runes runtime package gate green: mixin class/config/manifest are self-consistent.'

python3 "$PORT/tools/prepare_port.py" "$WORK/wizards-base" "$WORK/wizards-target" "$GEN"
python3 "$PORT/tools/compat_1_20_1.py" "$GEN"
python3 "$PORT/tools/compat_api_1_20_1.py" "$GEN"
python3 "$PORT/tools/compat_trade_access.py" "$GEN"
python3 "$PORT/tools/compat_tiny_config.py" "$GEN"
mkdir -p "$GEN/libs"
cp "$ARMOR_COMMON" "$GEN/libs/armor-model-api-common.jar"
cp "$ARMOR_FORGE" "$GEN/libs/armor-model-api-forge.jar"
cp "$STRUCTURE_COMMON" "$GEN/libs/structure-pool-api-common.jar"
cp "$STRUCTURE_FORGE" "$GEN/libs/structure-pool-api-forge.jar"
cp "$RUNES_COMMON" "$GEN/libs/runes-common.jar"
cp "$RUNES_FORGE" "$GEN/libs/runes-forge.jar"
cp "$SPELL_POWER_COMMON" "$GEN/libs/spell-power-common.jar"
cp "$SPELL_POWER_FORGE" "$GEN/libs/spell-power-forge.jar"
cp "$SPELL_ENGINE_COMMON" "$GEN/libs/spell-engine-common.jar"
cp "$SPELL_ENGINE_FORGE" "$GEN/libs/spell-engine-forge.jar"
cp "$TINY_COMMON" "$GEN/libs/tiny-config-common.jar"
cp "$TINY_FORGE" "$GEN/libs/tiny-config-forge.jar"

grep -F "substrate=$BASE_SHA" "$GEN/PORT-PINS.txt"
grep -F "target=$TARGET_SHA" "$GEN/PORT-PINS.txt"
grep -F 'architectury_loom=1.7.435' "$GEN/PORT-PINS.txt"
grep -F 'architectury_plugin=3.4.164' "$GEN/PORT-PINS.txt"
grep -F 'tiny_config_version=3.1.0+1.20.1' "$GEN/gradle.properties"
test ! -d "$GEN/fabric"
test ! -d "$GEN/neoforge"
test -f "$GEN/forge/src/main/java/net/wizards/forge/ForgeMod.java"
test -f "$GEN/forge/src/main/java/net/wizards/forge/client/ForgeClientMod.java"
test -f "$GEN/forge/src/main/resources/META-INF/mods.toml"
unzip -p "$TINY_FORGE" META-INF/mods.toml | grep -F 'modId="tiny_config"'
grep -F 'modId="tiny_config"' "$GEN/forge/src/main/resources/META-INF/mods.toml"
if grep -R -nE 'net\.neoforged|NeoForge\.' "$GEN/forge/src/main/java"; then
  echo '[Wizards] NeoForge loader symbol leaked into generated Forge source' >&2
  exit 2
fi
if grep -R -nE 'sourceSets.*(spell|runes|armor|structure|tiny)|srcDirs.*(spell|runes|armor|structure|tiny)' "$GEN/common/build.gradle" "$GEN/forge/build.gradle"; then
  echo '[Wizards] dependency source injection detected' >&2
  exit 2
fi
if grep -R -n 'maven.modrinth:tiny-config' "$GEN/common/build.gradle" "$GEN/forge/build.gradle"; then
  echo '[Wizards] unresolved external TinyConfig coordinate survived local foundation staging' >&2
  exit 2
fi
if grep -E 'TradeOffers\.(SellItemFactory|SellEnchantedToolFactory)' "$GEN/common/src/main/java/net/wizards/villager/WizardVillagers.java"; then
  echo '[Wizards] runtime-inaccessible vanilla villager trade implementation survived compatibility passes' >&2
  exit 2
fi

echo '[Wizards] Compiling current 3.1.1 behavior/content against Minecraft 1.20.1 + real separate foundation JARs'
gradle --no-daemon --stacktrace -p "$GEN" :common:compileJava :forge:compileJava

echo '[Wizards] Initial compile layer green; probing remapped release package'
gradle --no-daemon --stacktrace -p "$GEN" :forge:remapJar
JAR="$(pick_jar "$GEN/forge/build/libs")"
test -f "$JAR"
unzip -tq "$JAR"
unzip -p "$JAR" META-INF/mods.toml | grep -F 'modId="wizards"'
unzip -p "$JAR" META-INF/MANIFEST.MF | tr -d '\r' | grep -F 'MixinConfigs: wizards.mixins.json'
for prefix in 'net/spell_engine/' 'net/spell_power/' 'net/rpg_series/runes/' 'net/rpg_foundation/armor_api/' 'net/fabric_extras/structure_pool/' 'net/tiny_config/'; do
  if unzip -Z1 "$JAR" | grep -q "^$prefix"; then
    echo "[Wizards] ERROR: packaged foundation classes leaked under $prefix" >&2
    exit 3
  fi
done
sha256sum "$JAR" | tee "$PORT/wizards.sha256"
echo "[Wizards] compile/package bootstrap passed: $JAR"
