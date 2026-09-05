#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
CONSUMER="$ROOT/rpg-series-port/archers-forge-1.20.1/quiver-consumer"
BUNDLE_JAR="${1:?usage: run-bundle-api-archers-consumer.sh /absolute/path/to/bundle-api.jar}"

[[ -f "$BUNDLE_JAR" ]]
[[ -f "$CONSUMER/PORT_CONTRACT_DOES_NOT_EXIST" ]] && { echo '[Archers consumer] impossible sentinel exists' >&2; exit 1; } || true
BUNDLE_SHA_BEFORE="$(sha256sum "$BUNDLE_JAR" | awk '{print $1}')"

echo '[Archers consumer] Verify fixture is external-JAR only'
grep -F 'new CustomBundleItem(ItemTags.ARROWS, 4,' "$CONSUMER/src/main/java/net/archers/validation/ArchersQuiverConsumerMod.java" >/dev/null
grep -F 'new CustomBundleItem(ItemTags.ARROWS, 8,' "$CONSUMER/src/main/java/net/archers/validation/ArchersQuiverConsumerMod.java" >/dev/null
grep -F 'new CustomBundleItem(ItemTags.ARROWS, 12,' "$CONSUMER/src/main/java/net/archers/validation/ArchersQuiverConsumerMod.java" >/dev/null
if grep -R -E 'project\(|bundle-api-forge-1\.20\.1/(common|forge)|sourceSets.*bundle' "$CONSUMER" --include='*.gradle' --include='*.java'; then
  echo '[Archers consumer] source/project coupling detected; packaged Bundle API JAR must be the only implementation dependency' >&2
  exit 1
fi

echo '[Archers consumer] Compile/package against separate Bundle API release JAR'
gradle --no-daemon --stacktrace -p "$CONSUMER" clean build -Pbundle_api_jar="$BUNDLE_JAR"
CONSUMER_JAR="$(find "$CONSUMER/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' ! -name '*javadoc*' | sort | head -n1)"
[[ -n "$CONSUMER_JAR" && -f "$CONSUMER_JAR" ]]
unzip -tq "$CONSUMER_JAR"
unzip -p "$CONSUMER_JAR" META-INF/mods.toml | grep -F 'modId="bundleapi"' >/dev/null
unzip -p "$CONSUMER_JAR" META-INF/mods.toml | grep -F 'modId="archersquiverconsumer"' >/dev/null

python3 - "$CONSUMER_JAR" <<'PY'
import struct,sys,zipfile
jar=sys.argv[1]; owned=[]
with zipfile.ZipFile(jar) as zf:
    for name in zf.namelist():
        if name.startswith('net/archers/validation/') and name.endswith('.class'):
            data=zf.read(name)
            if len(data)<8 or data[:4] != b'\xca\xfe\xba\xbe':
                raise SystemExit('[Archers consumer] malformed class '+name)
            major=struct.unpack('>H',data[6:8])[0]
            if major != 61:
                raise SystemExit(f'[Archers consumer] {name} is class major {major}, expected Java17 major61')
            owned.append(name)
if not owned:
    raise SystemExit('[Archers consumer] no owned consumer classes packaged')
print(f'[Archers consumer] Java17 gate passed for {len(owned)} owned class(es).')
PY

BUNDLE_SHA_AFTER="$(sha256sum "$BUNDLE_JAR" | awk '{print $1}')"
[[ "$BUNDLE_SHA_BEFORE" = "$BUNDLE_SHA_AFTER" ]] || { echo '[Archers consumer] Bundle API release JAR mutated during consumer build' >&2; exit 1; }
printf '%s  %s\n' "$BUNDLE_SHA_AFTER" "$BUNDLE_JAR" | tee "$CONSUMER/bundle-api-consumed.sha256"
sha256sum "$CONSUMER_JAR" | tee "$CONSUMER/archers-quiver-consumer.sha256"

# Real initialized Forge server: tags are loaded by ServerStartedEvent before the matrix executes.
echo '[Archers consumer] Real Forge server + loaded-tag/capacity matrix'
rm -rf "$CONSUMER/run/logs"
mkdir -p "$CONSUMER/run"
printf 'eula=true\n' > "$CONSUMER/run/eula.txt"
SERVER_SMOKE="$CONSUMER/archers-quiver-consumer-server.log"
: > "$SERVER_SMOKE"
set +e
timeout --signal=TERM --kill-after=10s 180s gradle --no-daemon -p "$CONSUMER" runServer \
  -Pbundle_api_jar="$BUNDLE_JAR" > "$SERVER_SMOKE" 2>&1
SERVER_STATUS=$?
set -e
SERVER_LOG=$(find "$CONSUMER/run" -type f -path '*/logs/latest.log' | head -n1 || true)
SERVER_FILES=("$SERVER_SMOKE"); [[ -n "$SERVER_LOG" ]] && SERVER_FILES+=("$SERVER_LOG")
FATAL_SERVER='ARCHERS_QUIVER_CONSUMER_FAILED|ModLoadingException|Failed to create mod instance|NoClassDefFoundError|ClassNotFoundException|MixinApplyError|InvalidMixinException|Exception in server tick loop|The game crashed|Attempted to load class .* for invalid dist DEDICATED_SERVER'
if grep -Eiq "$FATAL_SERVER" "${SERVER_FILES[@]}"; then cat "${SERVER_FILES[@]}"; exit 1; fi
if ! grep -Fq 'ARCHERS_QUIVER_CONSUMER_PASS' "${SERVER_FILES[@]}"; then
  cat "${SERVER_FILES[@]}"; echo '[Archers consumer] live quiver matrix did not report PASS' >&2; exit 1
fi
if [[ -n "$SERVER_LOG" ]] && grep -Eq 'Done \([0-9.]+s\)!' "$SERVER_LOG"; then
  echo '[Archers consumer] Forge server ready + exact current quiver capacity/tag matrix PASS.'
elif [[ "$SERVER_STATUS" -ne 124 && "$SERVER_STATUS" -ne 143 ]]; then
  cat "${SERVER_FILES[@]}"; exit "$SERVER_STATUS"
else
  cat "${SERVER_FILES[@]}"; echo '[Archers consumer] server matrix passed but ready-state evidence is missing' >&2; exit 1
fi

# Both external artifacts must also coexist through real client resource/render bootstrap.
echo '[Archers consumer] Forge client coexistence gate'
rm -rf "$CONSUMER/run/logs"
mkdir -p "$CONSUMER/run/config"
printf 'earlyWindowControl = false\n' > "$CONSUMER/run/config/fml.toml"
CLIENT_SMOKE="$CONSUMER/archers-quiver-consumer-client.log"
: > "$CLIENT_SMOKE"
set +e
timeout --signal=TERM --kill-after=10s 180s env LIBGL_ALWAYS_SOFTWARE=1 MESA_LOADER_DRIVER_OVERRIDE=llvmpipe \
  xvfb-run -a -s '-screen 0 1280x720x24 +extension GLX +extension RENDER -noreset' \
  gradle --no-daemon -p "$CONSUMER" runClient -Pbundle_api_jar="$BUNDLE_JAR" </dev/null > "$CLIENT_SMOKE" 2>&1
CLIENT_STATUS=$?
set -e
CLIENT_LOG=$(find "$CONSUMER/run" -type f -path '*/logs/latest.log' | head -n1 || true)
CLIENT_FILES=("$CLIENT_SMOKE"); [[ -n "$CLIENT_LOG" ]] && CLIENT_FILES+=("$CLIENT_LOG")
FATAL_CLIENT='MixinApplyError|InvalidMixinException|MixinTransformerError|Failed to create mod instance|NoClassDefFoundError|ClassNotFoundException|The game crashed whilst initializing game|Exception in thread "Render thread"|Failed to initialize graphics window|Timed out trying to setup the Game Window|Could not initialize GLFW|Missing or unsupported mandatory dependencies|RegisterClientTooltipComponentFactoriesEvent.*[Ee]xception|CustomBundleTooltip(Component|Data).*(Exception|Error)'
if grep -Eiq "$FATAL_CLIENT" "${CLIENT_FILES[@]}"; then cat "${CLIENT_FILES[@]}"; exit 1; fi
if [[ -n "$CLIENT_LOG" ]] && grep -Fq 'Reloading ResourceManager' "$CLIENT_LOG" && grep -Fq 'Backend library: LWJGL' "$CLIENT_LOG"; then
  echo '[Archers consumer] Bundle API + Archers consumer coexist through post-bootstrap client resource/render runtime.'
elif [[ "$CLIENT_STATUS" -ne 124 && "$CLIENT_STATUS" -ne 143 ]]; then
  cat "${CLIENT_FILES[@]}"; echo "[Archers consumer] client exited before bootstrap evidence: $CLIENT_STATUS" >&2; exit 1
else
  cat "${CLIENT_FILES[@]}"; echo '[Archers consumer] client timed out before proven post-bootstrap state' >&2; exit 1
fi

echo '[Archers consumer] Acceptance passed: separate-JAR compile/package + live tags + 256/512/768 capacities + overflow/removal + Forge server/client coexistence.'
