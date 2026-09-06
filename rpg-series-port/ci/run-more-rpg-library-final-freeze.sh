#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
PORT="$ROOT/rpg-series-port/more-rpg-library-forge-1.20.1"
ENV="$PORT/MORE_RPG_LIBRARY_GRADUATION.env"
JAR="$PORT/more_rpg_library-forge-2.7.2+1.20.1.jar"
SOURCE_SHA_FILE="$PORT/more-rpg-library-source.sha256"
SOURCE_PACKAGE_SHA_FILE="$PORT/more-rpg-library-source-package.sha256"
SOURCE_ZIP="$PORT/more-rpg-library-2.7.2-forge-1.20.1-source-ci.zip"
ACTIVE_LOG="$ROOT/rpg-series-port/active-port-verifier.log"
PACKAGE_LOG="$ROOT/rpg-series-port/more-rpg-package-linkage.log"
STAGE1_LOG="$ROOT/rpg-series-port/more-rpg-runtime-stage1.log"
STAGE2_LOG="$ROOT/rpg-series-port/more-rpg-production-stage2.log"
STAGE3_LOG="$ROOT/rpg-series-port/more-rpg-production-stage3-restart.log"
REPRO_LOG="$ROOT/rpg-series-port/more-rpg-reproducibility.log"

for f in "$ENV" "$JAR" "$SOURCE_SHA_FILE" "$SOURCE_PACKAGE_SHA_FILE" "$SOURCE_ZIP" \
         "$ACTIVE_LOG" "$PACKAGE_LOG" "$STAGE1_LOG" "$STAGE2_LOG" "$STAGE3_LOG" "$REPRO_LOG"; do
  test -f "$f"
done
source "$ENV"
unzip -tq "$JAR" >/dev/null
unzip -tq "$SOURCE_ZIP" >/dev/null

# Every preceding graduation family must have emitted its own positive marker on this exact head.
grep -Fq '[More RPG 2.7.2] CERTIFIED_RPG_FOUNDATIONS_READY' "$ACTIVE_LOG"
grep -Fq '[More RPG 2.7.2] ENCHANTMENT_DATA_1201_DOWNGRADE_PASS' "$ACTIVE_LOG"
grep -Fq '[More RPG 2.7.2] ENCHANTMENT_DATA_1201_RUNTIME_BRIDGE_READY' "$ACTIVE_LOG"
grep -Fq '[More RPG 2.7.2] FULL_MODERN_COMMON_NATIVE_FORGE_COMPILE_PASS' "$ACTIVE_LOG"
grep -Fq '[More RPG 2.7.2] FIRST_NATIVE_FORGE_PACKAGE_PASS' "$ACTIVE_LOG"
grep -Fq '[More RPG 2.7.2] PACKAGED_EXTERNAL_SCHOOL_ATTRIBUTE_OWNERSHIP_PASS' "$PACKAGE_LOG"
grep -Fq '[More RPG 2.7.2] PRODUCTION_SYMBOLIC_LINKAGE_AUDIT_PASS' "$PACKAGE_LOG"
grep -Fq '[More RPG 2.7.2] PACKAGED_MIXIN_MANIFEST_REGISTRATION_PASS' "$PACKAGE_LOG"
grep -Fq '[More RPG 2.7.2] RUNTIME_STAGE1_HARDENED_PASS' "$STAGE1_LOG"
grep -Fq '[More RPG 2.7.2] PRODUCTION_FORGECLIENT_STAGE2_PASS' "$STAGE2_LOG"
grep -Fq '[More RPG 2.7.2] PRODUCTION_RESTART_PERSISTENCE_PASS' "$STAGE3_LOG"
grep -Fq '[More RPG 2.7.2] INDEPENDENT_SOURCE_REPLAY_IDENTITY_PASS' "$REPRO_LOG"
grep -Fq '[More RPG 2.7.2] INDEPENDENT_JAR_REPLAY_IDENTITY_PASS' "$REPRO_LOG"
grep -Fq '[More RPG 2.7.2] DETERMINISTIC_SOURCE_PACKAGE_IDENTITY_PASS' "$REPRO_LOG"
grep -Fq '[More RPG 2.7.2] INDEPENDENT_REPLAY_IDENTITY_PASS' "$REPRO_LOG"

JAR_SHA="$(sha256sum "$JAR" | awk '{print $1}')"
SOURCE_SHA="$(awk 'NR==1{print $1}' "$SOURCE_SHA_FILE")"
SOURCE_ZIP_SHA="$(awk 'NR==1{print $1}' "$SOURCE_PACKAGE_SHA_FILE")"
[[ "$JAR_SHA" =~ ^[0-9a-f]{64}$ ]]
[[ "$SOURCE_SHA" =~ ^[0-9a-f]{64}$ ]]
[[ "$SOURCE_ZIP_SHA" =~ ^[0-9a-f]{64}$ ]]
[[ "$(sha256sum "$SOURCE_ZIP" | awk '{print $1}')" = "$SOURCE_ZIP_SHA" ]]

# Java 17 release boundary: every More RPG-owned class must be exactly major 61 and no packaged
# class may require a newer JVM. This inspects the actual final JAR, not Gradle settings.
python3 - "$JAR" <<'PY'
import struct, sys, zipfile
jar = sys.argv[1]
owned = 0
all_classes = 0
with zipfile.ZipFile(jar) as zf:
    for name in zf.namelist():
        if not name.endswith('.class'):
            continue
        raw = zf.read(name)
        if len(raw) < 8 or raw[:4] != b'\xca\xfe\xba\xbe':
            raise SystemExit(f'[More RPG 2.7.2] invalid class header: {name}')
        major = struct.unpack('>H', raw[6:8])[0]
        all_classes += 1
        if major > 61:
            raise SystemExit(f'[More RPG 2.7.2] Java-newer-than-17 class: {name} major={major}')
        if name.startswith('net/more_rpg_classes/'):
            owned += 1
            if major != 61:
                raise SystemExit(f'[More RPG 2.7.2] owned class not Java 17: {name} major={major}')
if owned == 0:
    raise SystemExit('[More RPG 2.7.2] no owned classes found in final JAR')
print(f'[More RPG 2.7.2] JAVA17_FINAL_PACKAGE_PASS owned={owned} packaged={all_classes} major=61')
PY

# Verify the exact already-graduated dependency locks still match this lane's graduation contract.
PORT_ENV="$PORT/MORE_RPG_LIBRARY_PORT.env"
test -f "$PORT_ENV"
source "$PORT_ENV"
[[ "$SPELL_ENGINE_1104_EXPECTED_JAR_SHA" = "7222241a6208f0bede8f3971238ba0efbc8eefb3d299174ae65438d913be5ff3" ]]
[[ "$SPELL_POWER_160_EXPECTED_JAR_SHA" = "bb085d90f5196b08ef9ddae1f1faa8c5631a88a112450d3768511386e65fa4f3" ]]
[[ "$RANGED_WEAPON_API_234_EXPECTED_JAR_SHA" = "e387df1f42473864e687715c2495e66d57f64b53daae463b5d2e2157c2da6894" ]]
[[ "$TINY_CONFIG_310_EXPECTED_JAR_SHA" = "0182a492d6c59d7d5f491a39bb2f6634ba5dd38083295305c4769fdb6539db18" ]]

capture='__CAPTURE_AFTER_FIRST_FULL_GREEN__'
if [[ "$MORE_RPG_LIBRARY_EXPECTED_JAR_SHA" = "$capture" \
   && "$MORE_RPG_LIBRARY_EXPECTED_SOURCE_SHA" = "$capture" \
   && "$MORE_RPG_LIBRARY_EXPECTED_SOURCE_ZIP_SHA" = "$capture" ]]; then
  echo "[More RPG 2.7.2] MORE_RPG_FIRST_FULL_GREEN_CAPTURE jar=$JAR_SHA source=$SOURCE_SHA source_zip=$SOURCE_ZIP_SHA"
  exit 0
fi

for value in "$MORE_RPG_LIBRARY_EXPECTED_JAR_SHA" "$MORE_RPG_LIBRARY_EXPECTED_SOURCE_SHA" "$MORE_RPG_LIBRARY_EXPECTED_SOURCE_ZIP_SHA"; do
  [[ "$value" =~ ^[0-9a-f]{64}$ ]] || { echo '[More RPG 2.7.2] frozen graduation identity is not a SHA-256' >&2; exit 1; }
done
[[ "$JAR_SHA" = "$MORE_RPG_LIBRARY_EXPECTED_JAR_SHA" ]] || { echo "[More RPG 2.7.2] frozen JAR mismatch expected=$MORE_RPG_LIBRARY_EXPECTED_JAR_SHA actual=$JAR_SHA" >&2; exit 1; }
[[ "$SOURCE_SHA" = "$MORE_RPG_LIBRARY_EXPECTED_SOURCE_SHA" ]] || { echo "[More RPG 2.7.2] frozen source mismatch expected=$MORE_RPG_LIBRARY_EXPECTED_SOURCE_SHA actual=$SOURCE_SHA" >&2; exit 1; }
[[ "$SOURCE_ZIP_SHA" = "$MORE_RPG_LIBRARY_EXPECTED_SOURCE_ZIP_SHA" ]] || { echo "[More RPG 2.7.2] frozen source ZIP mismatch expected=$MORE_RPG_LIBRARY_EXPECTED_SOURCE_ZIP_SHA actual=$SOURCE_ZIP_SHA" >&2; exit 1; }

echo "[More RPG 2.7.2] CROSS_RUN_RELEASE_IDENTITY_PASS jar=$JAR_SHA source=$SOURCE_SHA source_zip=$SOURCE_ZIP_SHA"
echo '[More RPG 2.7.2] MORE_RPG_LIBRARY_GRADUATION_PASS forge=47.4.23 minecraft=1.20.1 target=2.7.2 production_client=true restart_persistence=true enchantments=true reproducible=true'
