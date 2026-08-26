#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
PORT="$ROOT/rpg-series-port/armor-model-api-forge-1.20.1"
UP="$ROOT/.rpg-upstream/armor-model-api"
GEN="$PORT/generated"
TARGET=a664155a0aab3161cd7e4bf0c1f72512b4ec4949
rm -rf "$UP" "$GEN"
mkdir -p "$(dirname "$UP")"
git init -q "$UP"
git -C "$UP" remote add origin https://github.com/FabricExtras/ArmorModelAPI.git
git -C "$UP" fetch -q --depth=1 origin "$TARGET"
git -C "$UP" checkout -q --detach FETCH_HEAD
test "$(git -C "$UP" rev-parse HEAD)" = "$TARGET"
python "$PORT/tools/prepare_port.py" "$UP" "$GEN"

test -f "$GEN/PORT-PINS.txt"
grep -F "target=$TARGET" "$GEN/PORT-PINS.txt"
test ! -d "$GEN/fabric"
test ! -d "$GEN/neoforge"
if grep -R 'net.neoforged' "$GEN/common/src/main/java" "$GEN/forge/src/main/java"; then
  echo '[Armor Model API] NeoForge symbol leaked into native Forge source' >&2; exit 1
fi

echo '[Armor Model API] First native Forge 1.20.1 compile probe'
gradle --no-daemon -p "$GEN" --stacktrace :forge:remapJar
JAR=$(find "$GEN/forge/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' ! -name '*dev-shadow*' | head -1)
test -n "$JAR"
unzip -t "$JAR"
unzip -p "$JAR" META-INF/mods.toml | grep -F 'modId="armor_model_api"'
unzip -l "$JAR" | grep -F 'net/rpg_foundation/armor_api/client/GeoArmorRenderer.class'
unzip -l "$JAR" | grep -F 'net/rpg_foundation/armor_api/forge/ArmorModelApiForge.class'
sha256sum "$JAR" | tee "$PORT/armor-model-api.sha256"
rm -f "$ROOT/armor-model-api-1.0.0-forge-1.20.1-source-ci.zip"
(cd "$GEN" && zip -qr "$ROOT/armor-model-api-1.0.0-forge-1.20.1-source-ci.zip" . -x '*/build/*' '*/run/*' '.gradle/*')
sha256sum "$ROOT/armor-model-api-1.0.0-forge-1.20.1-source-ci.zip" | tee "$PORT/armor-model-api-source.sha256"
echo '[Armor Model API] Initial build/package probe passed.'
