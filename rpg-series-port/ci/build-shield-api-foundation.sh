#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
PORT="$ROOT/rpg-series-port/shield-api-forge-1.20.1"
UP="$ROOT/.shield-api-upstream"
# shellcheck disable=SC1091
source "$PORT/UPSTREAM_PINS.env"
rm -rf "$UP" "$PORT/common/src/generatedUpstream"
mkdir -p "$UP"
clone_exact() {
  local sha="$1" dest="$2"
  git init -q "$dest"
  git -C "$dest" remote add origin https://github.com/FabricExtras/ShieldAPI.git
  git -C "$dest" fetch -q --depth=1 origin "$sha"
  git -C "$dest" checkout -q --detach FETCH_HEAD
  test "$(git -C "$dest" rev-parse HEAD)" = "$sha"
}
clone_exact "$SHIELD_API_LEGACY_1201_SHA" "$UP/legacy" & p1=$!
clone_exact "$SHIELD_API_CURRENT_SHA" "$UP/current" & p2=$!
wait "$p1" "$p2"
test "$(git -C "$UP/current" rev-parse HEAD^{tree})" = "$SHIELD_API_CURRENT_TREE"
python3 "$PORT/tools/prepare_upstream_source.py" "$UP/legacy" "$UP/current" "$PORT/common"

# Forge Loom must be selected in the Forge subproject before any expensive Gradle work.
# Keep this as a cheap regression gate because Architectury otherwise fails during project evaluation.
test -f "$PORT/forge/gradle.properties"
grep -Fx 'loom.platform=forge' "$PORT/forge/gradle.properties" >/dev/null

GEN="$PORT/common/src/generatedUpstream"
test -f "$GEN/java/net/fabric_extras/shield_api/item/CustomShieldItem.java"
test -f "$GEN/java/net/fabric_extras/shield_api/mixin/entity/player/PlayerEntityMixin.java"
grep -F 'set(customShieldItem, 100)' "$GEN/java/net/fabric_extras/shield_api/mixin/entity/player/PlayerEntityMixin.java" >/dev/null
if grep -Eq 'EnchantmentHelper|BREAK_SHIELD|nextFloat\(\) <' "$GEN/java/net/fabric_extras/shield_api/mixin/entity/player/PlayerEntityMixin.java"; then
  echo '[Shield API] historical probabilistic disable semantics leaked into current behavior port' >&2
  exit 2
fi
grep -F 'EquipmentSlot.OFFHAND' "$GEN/java/net/fabric_extras/shield_api/item/CustomShieldItem.java" >/dev/null
grep -F 'setAttributeModifiers' "$GEN/java/net/fabric_extras/shield_api/item/CustomShieldItem.java" >/dev/null
grep -F 'instances.add(this)' "$GEN/java/net/fabric_extras/shield_api/item/CustomShieldItem.java" >/dev/null

gradle --no-daemon --stacktrace -p "$PORT" clean :common:compileJava :forge:build
JAR="$(find "$PORT/forge/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*dev-shadow*' ! -name '*sources*' | sort | head -n1)"
test -f "$JAR"
unzip -tq "$JAR"

# Architectury's generated injection namespace contains a build-ephemeral 32-hex token.
# The effective Shield payload is stable, but that token makes fresh archive bytes differ
# across runs. Fail-closed certification verifies every normalized entry/payload against
# the accepted run-188 manifest, canonicalizes only that generated namespace/self-name,
# and requires the resulting JAR to equal the exact graduated run-188 SHA-256.
RAW_SHA="$(sha256sum "$JAR" | awk '{print $1}')"
printf '%s  %s\n' "$RAW_SHA" "$JAR" | tee "$PORT/shield-api-forge-1.20.1.raw.sha256"
CERTIFIED="$PORT/forge/build/libs/.shield-api-run188-certified.jar"
python3 "$ROOT/rpg-series-port/ci/certify-shield-api-run188.py" "$JAR" "$CERTIFIED"
mv "$CERTIFIED" "$JAR"
unzip -tq "$JAR"

grep -F 'modId="shield_api"' < <(unzip -p "$JAR" META-INF/mods.toml) >/dev/null
unzip -Z1 "$JAR" | grep -F 'net/fabric_extras/shield_api/item/CustomShieldItem.class' >/dev/null
unzip -Z1 "$JAR" | grep -F 'net/fabric_extras/shield_api/mixin/entity/player/PlayerEntityMixin.class' >/dev/null
if unzip -Z1 "$JAR" | grep -Eq '^net/fabricmc/|^net/neoforged/'; then
  echo '[Shield API] Fabric/NeoForge loader classes leaked into native Forge release' >&2
  exit 2
fi
MAX_MAJOR="$(unzip -Z1 "$JAR" | grep '\.class$' | while read -r cls; do unzip -p "$JAR" "$cls" | od -An -t u1 -N8 | awk '{print $7*256+$8}'; done | sort -nr | head -1)"
test "$MAX_MAJOR" -le 61
CERT_SHA="$(sha256sum "$JAR" | awk '{print $1}')"
test "$CERT_SHA" = 'bd6a2fbeb357c25953abfb14ba18d2c5344e5351c29d2cb082244bc48e8da48a'
sha256sum "$JAR" | tee "$PORT/shield-api-forge-1.20.1.sha256"
echo "[Shield API] exact graduated run-188 Forge foundation gate green: raw=$RAW_SHA certified=$CERT_SHA jar=$JAR"
