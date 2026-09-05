#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
SOURCE="${1:?usage: build-tiny-config-foundation.sh <exact-tiny-config-source>}"
PORT="$ROOT/rpg-series-port/tiny-config-forge-1.20.1"
GEN="$PORT/generated"
TARGET_SHA=e20fc8ac72fde8274f0df72de2ebb81ffe6f8727

test "$(git -C "$SOURCE" rev-parse HEAD)" = "$TARGET_SHA"
rm -rf "$GEN"
python3 "$PORT/tools/prepare_port.py" "$SOURCE" "$GEN"
grep -F "target=$TARGET_SHA" "$GEN/PORT-PINS.txt"
grep -F 'architectury_loom=1.7.435' "$GEN/PORT-PINS.txt"
grep -F 'architectury_plugin=3.4.164' "$GEN/PORT-PINS.txt"

gradle --no-daemon --stacktrace -p "$GEN" :common:jar :forge:remapJar
pick_jar() {
  find "$1" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' ! -name '*dev-shadow*' ! -name '*javadoc*' | sort | head -n1
}
COMMON_JAR="$(pick_jar "$GEN/common/build/libs")"
FORGE_JAR="$(pick_jar "$GEN/forge/build/libs")"
test -n "$COMMON_JAR" -a -f "$COMMON_JAR"
test -n "$FORGE_JAR" -a -f "$FORGE_JAR"
unzip -tq "$COMMON_JAR"
unzip -tq "$FORGE_JAR"
unzip -p "$FORGE_JAR" META-INF/mods.toml | grep -F 'modId="tiny_config"'
unzip -p "$FORGE_JAR" META-INF/MANIFEST.MF | tr -d '\r' | grep -F 'MixinConfigs: tiny_config.mixins.json'
unzip -Z1 "$FORGE_JAR" | grep -F 'net/tiny_config/forge/PlatformImpl.class'
if unzip -Z1 "$FORGE_JAR" | grep -Eq 'net/tiny_config/neoforge/'; then
  echo '[TinyConfig] NeoForge classes leaked into Forge release JAR' >&2
  exit 2
fi
javap -verbose -classpath "$FORGE_JAR" net.tiny_config.forge.ExampleModForge | grep -F 'major version: 61'
sha256sum "$FORGE_JAR" | tee "$PORT/tiny-config-forge-1.20.1.sha256"
echo "[TinyConfig] native Forge 1.20.1 foundation build passed: $FORGE_JAR"
