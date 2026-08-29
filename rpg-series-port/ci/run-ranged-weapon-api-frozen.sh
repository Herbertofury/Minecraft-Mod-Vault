#!/usr/bin/env bash
set -euo pipefail

ROOT="${GITHUB_WORKSPACE:?}"
PORT="$ROOT/rpg-series-port/ranged-weapon-api-forge-1.20.1"
CERTIFIER="$ROOT/rpg-series-port/ci/certify-ranged-weapon-api-run268.py"
EXPECTED_JAR_SHA="96c81b8187eea072fca39e809c1a339679ec5a2adf5b6ba93ccb8dbbf32ec6ae"
EXPECTED_SOURCE_SHA="9dca4b31dda65ef7cf86219db6c7c81ad64c518ea866c4f292437237de66de45"
FRESH="$PORT/.fresh-forge-server"

stop_tree() {
  local root="$1" child kids
  kids="$(pgrep -P "$root" 2>/dev/null || true)"
  for child in $kids; do stop_tree "$child"; done
  kill -TERM "$root" 2>/dev/null || true
  sleep 1
  kill -KILL "$root" 2>/dev/null || true
  wait "$root" 2>/dev/null || true
}

# Re-run the complete build + native userdev client + fresh packaged-server semantic suite first.
bash "$ROOT/rpg-series-port/ci/run-ranged-weapon-api.sh"

SOURCE_SHA="$(awk '{print $1}' "$PORT/ranged-weapon-api-source.sha256")"
[[ "$SOURCE_SHA" = "$EXPECTED_SOURCE_SHA" ]] || { echo "[Ranged Weapon API certification] source drifted: $SOURCE_SHA != $EXPECTED_SOURCE_SHA" >&2; exit 1; }

# Architectury emits a random 32-hex injection namespace into the archive path and PlatformMethods
# self-name. Canonicalize that build-only token only after proving the normalized payload manifest.
JAR="$(find "$PORT/forge/build/libs" -maxdepth 1 -type f -name 'ranged_weapon_api-forge-2.3.4+1.20.1.jar' -print -quit)"
[[ -f "$JAR" ]] || { echo '[Ranged Weapon API certification] release JAR missing' >&2; exit 1; }
TMP_JAR="$PORT/forge/build/libs/.ranged-weapon-api-certified.jar"
python3 "$CERTIFIER" "$JAR" "$TMP_JAR"
mv -f "$TMP_JAR" "$JAR"
unzip -tq "$JAR" >/dev/null
JAR_SHA="$(sha256sum "$JAR" | awk '{print $1}')"
[[ "$JAR_SHA" = "$EXPECTED_JAR_SHA" ]] || { echo "[Ranged Weapon API certification] canonical release drifted: $JAR_SHA != $EXPECTED_JAR_SHA" >&2; exit 1; }
printf '%s  %s\n' "$JAR_SHA" "$JAR" > "$PORT/ranged-weapon-api-forge.sha256"

# Fail closed on the exact production resource + Mixin/refmap surfaces.
unzip -p "$JAR" META-INF/MANIFEST.MF | grep -F 'MixinConfigs: ranged_weapon_api.mixins.json' >/dev/null
unzip -p "$JAR" ranged_weapon_api.mixins.json | grep -F '"refmap": "ranged_weapon_api-common-common-refmap.json"' >/dev/null
unzip -p "$JAR" ranged_weapon_api-common-common-refmap.json | grep -F 'net/minecraft/world/item/BowItem;m_40661_(I)F' >/dev/null
unzip -p "$JAR" pack.mcmeta | grep -F '"pack_format": 15' >/dev/null

grep -Fq 'Backend library: LWJGL' "$PORT/forge-client-smoke.log"
grep -Fq 'Reloading ResourceManager' "$PORT/forge-client-smoke.log"
if grep -Eiq 'MixinApplyError|InvalidMixinException|MixinTransformerError|Using missing texture.*ranged_weapon|Failed to load model.*ranged_weapon|Unable to load model.*ranged_weapon|The game crashed whilst initializing game' "$PORT/forge-client-smoke.log"; then
  echo '[Ranged Weapon API certification] native client regression detected' >&2
  exit 1
fi

# The original suite ran its packaged server with the pre-canonical random token. Replace only
# RWA with the exact canonical JAR and re-run the real Forge 47.4.23 server/self-test so the bytes
# we graduate are the bytes actually exercised in production packaging.
[[ -x "$FRESH/run.sh" ]] || { echo '[Ranged Weapon API certification] fresh packaged Forge server missing' >&2; exit 1; }
find "$FRESH/mods" -maxdepth 1 -type f -name 'ranged_weapon_api-forge-*.jar' -delete
cp -f "$JAR" "$FRESH/mods/"
INSTALLED="$FRESH/mods/$(basename "$JAR")"
[[ "$(sha256sum "$INSTALLED" | awk '{print $1}')" = "$EXPECTED_JAR_SHA" ]] || { echo '[Ranged Weapon API certification] installed canonical JAR identity mismatch' >&2; exit 1; }
rm -rf "$FRESH/logs"
CERT_LOG="$PORT/forge-package-canonical-smoke.log"
: > "$CERT_LOG"
(
  cd "$FRESH"
  exec ./run.sh nogui
) > "$CERT_LOG" 2>&1 &
PID=$!
DEADLINE=$((SECONDS+180))
PASS=0
FATAL='MixinApplyError|InvalidMixinException|MixinTransformerError|Failed to start the minecraft server|Failed to create mod instance|NoClassDefFoundError|ClassNotFoundException|Registry is already frozen|IllegalStateException: \[Ranged Weapon API CI\]|Exception in server tick loop'
while (( SECONDS < DEADLINE )); do
  LOG="$FRESH/logs/latest.log"
  files=("$CERT_LOG"); [[ -f "$LOG" ]] && files+=("$LOG")
  if grep -Eiq "$FATAL" "${files[@]}" 2>/dev/null; then
    stop_tree "$PID"; cat "${files[@]}"; exit 1
  fi
  if grep -Fq '[Ranged Weapon API CI] Runtime self-test passed: registries + item/equipment attributes + 2.3.3 mob hook + scaling' "$CERT_LOG" \
     && [[ -f "$LOG" ]] && grep -Eq 'Done \([0-9.]+s\)!' "$LOG"; then
    PASS=1
    break
  fi
  if ! kill -0 "$PID" 2>/dev/null; then
    wait "$PID" || true
    cat "${files[@]}"
    exit 1
  fi
  sleep 1
done
[[ "$PASS" -eq 1 ]] || { stop_tree "$PID"; cat "$CERT_LOG"; echo '[Ranged Weapon API certification] canonical packaged server timed out' >&2; exit 1; }
stop_tree "$PID"

echo "[Ranged Weapon API certification] CROSS_RUN_RELEASE_IDENTITY_PASS jar=$JAR_SHA source=$SOURCE_SHA"
echo '[Ranged Weapon API certification] CANONICAL_PACKAGED_SERVER_SEMANTICS_PASS'
echo '[Ranged Weapon API certification] RANGED_WEAPON_API_GRADUATION_PASS'
