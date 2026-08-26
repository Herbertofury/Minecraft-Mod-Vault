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
grep -F 'architectury_loom=1.7.435' "$GEN/PORT-PINS.txt"
grep -F 'architectury_plugin=3.4.164' "$GEN/PORT-PINS.txt"
grep -F 'architectury_common_project_id=0674af327c11f602eea7defcf5d514dc' "$GEN/PORT-PINS.txt"
grep -F 'architectury_forge_project_id=baae45cdf3cd30ac01447a62ccd0232e' "$GEN/PORT-PINS.txt"
test "$(cat "$GEN/common/.gradle/architectury-cache/projectID")" = '0674af327c11f602eea7defcf5d514dc'
test "$(cat "$GEN/forge/.gradle/architectury-cache/projectID")" = 'baae45cdf3cd30ac01447a62ccd0232e'
test ! -d "$GEN/fabric"
test ! -d "$GEN/neoforge"
if grep -R 'net.neoforged' "$GEN/common/src/main/java" "$GEN/forge/src/main/java"; then
  echo '[Armor Model API] NeoForge symbol leaked into native Forge source' >&2; exit 1
fi
MIXIN="$GEN/forge/src/main/resources/armor_model_api.mixins.json"
grep -F '"package": "net.rpg_foundation.armor_api.forge.mixin"' "$MIXIN"
grep -F '"compatibilityLevel": "JAVA_17"' "$MIXIN"
ENTRY="$GEN/forge/src/main/java/net/rpg_foundation/armor_api/forge/ArmorModelApiForge.java"
test -f "$ENTRY"
if grep -E 'net\.minecraft\.(client|resource)' "$ENTRY"; then
  echo '[Armor Model API] client/resource class leaked into top-level Forge mod constructor' >&2; exit 1
fi

echo '[Armor Model API] Native Forge 1.20.1 clean build + reobf probe'
gradle --no-daemon -p "$GEN" --stacktrace clean :forge:remapJar
JAR=$(find "$GEN/forge/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' ! -name '*dev-shadow*' | head -1)
test -n "$JAR"
unzip -t "$JAR"
unzip -p "$JAR" META-INF/mods.toml | grep -F 'modId="armor_model_api"'
unzip -p "$JAR" META-INF/MANIFEST.MF | tr -d '\r' | grep -F 'MixinConfigs: armor_model_api.mixins.json'
unzip -p "$JAR" armor_model_api.mixins.json | grep -F '"package": "net.rpg_foundation.armor_api.forge.mixin"'
unzip -p "$JAR" armor_model_api.mixins.json | grep -F '"compatibilityLevel": "JAVA_17"'
unzip -l "$JAR" | grep -F 'net/rpg_foundation/armor_api/client/GeoArmorRenderer.class'
unzip -l "$JAR" | grep -F 'net/rpg_foundation/armor_api/forge/ArmorModelApiForge.class'
unzip -l "$JAR" | grep -F 'net/rpg_foundation/armor_api/forge/client/ArmorModelApiForgeClient.class'
if unzip -l "$JAR" | grep -q 'net/rpg_foundation/armor_api/neoforge/\|META-INF/neoforge.mods.toml'; then
  echo '[Armor Model API] NeoForge package/metadata leaked into release JAR' >&2; exit 1
fi
javap -verbose -classpath "$JAR" net.rpg_foundation.armor_api.forge.ArmorModelApiForge | grep -F 'major version: 61'
FIRST_JAR_SHA=$(sha256sum "$JAR" | awk '{print $1}')
printf '%s  %s\n' "$FIRST_JAR_SHA" "$JAR" | tee "$PORT/armor-model-api-first-build.sha256"

echo '[Armor Model API] Reproducibility gate: clean rebuild must be byte-identical'
gradle --no-daemon -p "$GEN" --stacktrace clean :forge:remapJar
JAR=$(find "$GEN/forge/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' ! -name '*dev-shadow*' | head -1)
test -n "$JAR"
SECOND_JAR_SHA=$(sha256sum "$JAR" | awk '{print $1}')
printf '%s  %s\n' "$SECOND_JAR_SHA" "$JAR" | tee "$PORT/armor-model-api.sha256"
if [[ "$FIRST_JAR_SHA" != "$SECOND_JAR_SHA" ]]; then
  echo "[Armor Model API] non-reproducible release JAR: first=$FIRST_JAR_SHA second=$SECOND_JAR_SHA" >&2
  exit 1
fi
echo "[Armor Model API] Reproducible release JAR SHA-256: $SECOND_JAR_SHA"

SOURCE_ZIP="$ROOT/armor-model-api-1.0.0-forge-1.20.1-source-ci.zip"
rm -f "$SOURCE_ZIP"
python - "$GEN" "$SOURCE_ZIP" <<'PY'
from pathlib import Path
import stat, sys, zipfile
src = Path(sys.argv[1]).resolve()
out = Path(sys.argv[2]).resolve()
skip = {'.gradle', 'build', 'run', 'runs', '.git'}
files = []
for path in src.rglob('*'):
    rel = path.relative_to(src)
    if any(part in skip for part in rel.parts):
        continue
    if path.is_file():
        files.append((rel.as_posix(), path))
with zipfile.ZipFile(out, 'w', compression=zipfile.ZIP_DEFLATED, compresslevel=9) as zf:
    for arcname, path in sorted(files):
        info = zipfile.ZipInfo(arcname, date_time=(1980, 1, 1, 0, 0, 0))
        info.compress_type = zipfile.ZIP_DEFLATED
        info.external_attr = (stat.S_IFREG | 0o644) << 16
        info.create_system = 3
        zf.writestr(info, path.read_bytes(), compress_type=zipfile.ZIP_DEFLATED, compresslevel=9)
PY
unzip -t "$SOURCE_ZIP" >/dev/null
if unzip -l "$SOURCE_ZIP" | grep -E '/\.gradle/|/build/|/run/|/runs/' >/dev/null; then
  echo '[Armor Model API] generated cache/build/runtime state leaked into source archive' >&2; exit 1
fi
sha256sum "$SOURCE_ZIP" | tee "$PORT/armor-model-api-source.sha256"

echo '[Armor Model API] Dedicated-server side-safety gate'
rm -rf "$GEN/forge/run/logs"
mkdir -p "$GEN/forge/run"
printf 'eula=true\n' > "$GEN/forge/run/eula.txt"
: > "$PORT/armor-model-api-server-smoke.log"
set +e
timeout --signal=TERM --kill-after=10s 210s gradle --no-daemon -p "$GEN" :forge:runServer > "$PORT/armor-model-api-server-smoke.log" 2>&1
SERVER_STATUS=$?
set -e
SERVER_LOG=$(find "$GEN/forge/run" -type f -path '*/logs/latest.log' | head -1 || true)
SERVER_FILES=("$PORT/armor-model-api-server-smoke.log")
[[ -n "$SERVER_LOG" ]] && SERVER_FILES+=("$SERVER_LOG")
if grep -Eiq 'MixinApplyError|InvalidMixinException|MixinTransformerError|Failed to create mod instance|NoClassDefFoundError|Exception in server tick loop|The game crashed' "${SERVER_FILES[@]}"; then
  cat "${SERVER_FILES[@]}"; exit 1
fi
if [[ -n "$SERVER_LOG" ]] && grep -Eq 'Done \([0-9.]+s\)!' "$SERVER_LOG"; then
  echo '[Armor Model API] Dedicated server reached ready state.'
elif [[ "$SERVER_STATUS" -ne 124 && "$SERVER_STATUS" -ne 143 ]]; then
  cat "${SERVER_FILES[@]}"; echo "server exited before ready state: $SERVER_STATUS" >&2; exit 1
else
  cat "${SERVER_FILES[@]}"; echo '[Armor Model API] server runtime timed out before a proven ready state' >&2; exit 1
fi

echo '[Armor Model API] Headless Forge client bootstrap gate'
rm -rf "$GEN/forge/run/logs"
: > "$PORT/armor-model-api-client-smoke.log"
set +e
timeout --signal=TERM --kill-after=10s 210s xvfb-run -a gradle --no-daemon -p "$GEN" :forge:runClient > "$PORT/armor-model-api-client-smoke.log" 2>&1
CLIENT_STATUS=$?
set -e
CLIENT_LOG=$(find "$GEN/forge/run" -type f -path '*/logs/latest.log' | head -1 || true)
CLIENT_FILES=("$PORT/armor-model-api-client-smoke.log")
[[ -n "$CLIENT_LOG" ]] && CLIENT_FILES+=("$CLIENT_LOG")
if grep -Eiq 'MixinApplyError|InvalidMixinException|MixinTransformerError|Failed to create mod instance|NoClassDefFoundError.*armor_model_api|The game crashed whilst initializing game|Using missing texture.*armor_model_api' "${CLIENT_FILES[@]}"; then
  cat "${CLIENT_FILES[@]}"; exit 1
fi
if [[ -n "$CLIENT_LOG" ]] && grep -Eiq 'OpenAL initialized|Created: [0-9]+x[0-9]+|Reloading ResourceManager|Sound engine started|Narrator library' "$CLIENT_LOG"; then
  echo '[Armor Model API] Headless client reached post-bootstrap runtime.'
elif [[ "$CLIENT_STATUS" -ne 124 && "$CLIENT_STATUS" -ne 143 ]]; then
  cat "${CLIENT_FILES[@]}"; echo "client exited before bootstrap evidence: $CLIENT_STATUS" >&2; exit 1
else
  cat "${CLIENT_FILES[@]}"; echo '[Armor Model API] client runtime timed out before a proven bootstrap state' >&2; exit 1
fi

echo '[Armor Model API] Fresh packaged Forge server gate'
bash "$ROOT/rpg-series-port/ci/run-armor-model-api-package-smoke.sh" "$JAR" "$GEN" "$PORT"

echo '[Armor Model API] Build, reproducibility, package, Java 17 bytecode, dev server/client, and fresh packaged-server gates passed.'
