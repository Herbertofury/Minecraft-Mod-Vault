#!/usr/bin/env bash
set -euo pipefail

ROOT="${GITHUB_WORKSPACE:?}"
PORT="$ROOT/rpg-series-port/ranged-weapon-api-forge-1.20.1"
EXPECTED_JAR_SHA="a9671bf4a379c4c6e49d06a8f5140feb124505fc4901c9378aa7e64be5a841da"
EXPECTED_SOURCE_SHA="9dca4b31dda65ef7cf86219db6c7c81ad64c518ea866c4f292437237de66de45"

bash "$ROOT/rpg-series-port/ci/run-ranged-weapon-api.sh"

JAR_SHA="$(awk '{print $1}' "$PORT/ranged-weapon-api-forge.sha256")"
SOURCE_SHA="$(awk '{print $1}' "$PORT/ranged-weapon-api-source.sha256")"
[[ "$JAR_SHA" = "$EXPECTED_JAR_SHA" ]] || { echo "[Ranged Weapon API certification] release drifted: $JAR_SHA != $EXPECTED_JAR_SHA" >&2; exit 1; }
[[ "$SOURCE_SHA" = "$EXPECTED_SOURCE_SHA" ]] || { echo "[Ranged Weapon API certification] source drifted: $SOURCE_SHA != $EXPECTED_SOURCE_SHA" >&2; exit 1; }

JAR="$(find "$PORT/forge/build/libs" -maxdepth 1 -type f -name 'ranged_weapon_api-forge-2.3.4+1.20.1.jar' -print -quit)"
[[ -f "$JAR" ]] || { echo '[Ranged Weapon API certification] exact release JAR missing' >&2; exit 1; }
unzip -tq "$JAR" >/dev/null
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

grep -Fq '[Ranged Weapon API CI] Runtime self-test passed: registries + item/equipment attributes + 2.3.3 mob hook + scaling' "$PORT/forge-package-smoke.log"
grep -Eq 'Done \([0-9.]+s\)!' "$PORT/.fresh-forge-server/logs/latest.log"

echo "[Ranged Weapon API certification] CROSS_RUN_RELEASE_IDENTITY_PASS jar=$JAR_SHA source=$SOURCE_SHA"
echo '[Ranged Weapon API certification] PRODUCTION_MIXIN_RESOURCE_CONTRACT_PASS'
echo '[Ranged Weapon API certification] RANGED_WEAPON_API_GRADUATION_PASS'
