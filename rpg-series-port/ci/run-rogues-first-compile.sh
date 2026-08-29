#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"; PORT="$ROOT/rpg-series-port/rogues-forge-1.20.1"; TMP="${RUNNER_TEMP:-/tmp}/paladins-first-compile"
# Reuse the already-proven foundation reconstruction lane; this also fail-closes on graduated identities.
bash "$ROOT/rpg-series-port/ci/run-paladins-first-compile.sh"
resolve(){ local dir="$1" glob="$2"; find "$dir" -maxdepth 1 -type f -name "$glob" ! -name '*sources*' ! -name '*dev-shadow*' ! -name '*javadoc*' -print | sort | head -n1; }
SPELL_POWER="$ROOT/rpg-series-port/spell_power-forge-1.20.1"; TINY="$ROOT/rpg-series-port/tiny-config-forge-1.20.1/generated"; SPELL_ENGINE="$ROOT/.spell-engine-build"; STRUCT="$ROOT/rpg-series-port/structure_pool_api-forge-1.20.1"; ARMOR="$ROOT/rpg-series-port/armor-model-api-forge-1.20.1/generated"
SPC="$(resolve "$SPELL_POWER/common/build/libs" '*-common-*.jar')"; SPF="$(resolve "$SPELL_POWER/forge/build/libs" '*-forge-*.jar')"; TC="$(resolve "$TINY/common/build/libs" '*-common-*.jar')"; TF="$(resolve "$TINY/forge/build/libs" '*-forge-*.jar')"; SEC="$(resolve "$SPELL_ENGINE/common/build/libs" '*-common-*.jar')"; SEF="$ROOT/rpg-series-port/spell-engine-forge-1.20.1/spell_engine-forge-1.10.2+1.20.1.jar"; STC="$(resolve "$STRUCT/common/build/libs" '*-common-*.jar')"; STF="$(resolve "$STRUCT/forge/build/libs" '*-forge-*.jar')"; AMC="$(resolve "$ARMOR/common/build/libs" '*-common-*.jar')"; AMF="$(resolve "$ARMOR/forge/build/libs" '*-forge-*.jar')"
CLOTH="$TMP/ext/cloth-config-forge-11.1.136.jar"; PLAYER="$TMP/ext/player-animation-lib-forge-1.0.2+1.19.4.jar"; CURIOS="$TMP/ext/curios-forge-5.14.1+1.20.1.jar"
for f in "$SPC" "$SPF" "$TC" "$TF" "$SEC" "$SEF" "$STC" "$STF" "$AMC" "$AMF" "$CLOTH" "$PLAYER" "$CURIOS"; do [[ -f "$f" ]] || { echo "[Rogues] missing dependency $f" >&2; exit 3; }; unzip -tq "$f" >/dev/null; done
ARGS=("-Parmor_model_api_common_jar=$AMC" "-Pstructure_pool_api_common_jar=$STC" "-Pspell_power_common_jar=$SPC" "-Pspell_engine_common_jar=$SEC" "-Ptiny_config_common_jar=$TC" "-Parmor_model_api_forge_jar=$AMF" "-Pstructure_pool_api_forge_jar=$STF" "-Pspell_power_forge_jar=$SPF" "-Pspell_engine_forge_jar=$SEF" "-Ptiny_config_forge_jar=$TF" "-Pcloth_config_forge_jar=$CLOTH" "-Pplayer_animator_forge_jar=$PLAYER" "-Pcurios_jar=$CURIOS")
bash "$PORT/materialize_port.sh"; test -f "$PORT/generated/common/java/net/rogues/RoguesMod.java"; test -f "$PORT/generated/common/java/net/rogues/entity/BearTrapEntity.java"
if grep -R -nE 'net\.neoforged|net\.fabricmc\.fabric' "$PORT/generated/common/java" "$PORT/forge/src/main/java"; then echo '[Rogues] loader API leak' >&2; exit 4; fi
echo '[Rogues] common compile'; gradle --no-daemon --stacktrace -p "$PORT" clean :common:compileJava "${ARGS[@]}"
echo '[Rogues] Forge compile'; gradle --no-daemon --stacktrace -p "$PORT" :forge:compileJava "${ARGS[@]}"
echo '[Rogues] package'; gradle --no-daemon --stacktrace -p "$PORT" :forge:remapJar "${ARGS[@]}"
JAR="$(resolve "$PORT/forge/build/libs" '*-forge-*.jar')"; [[ -f "$JAR" ]]; unzip -tq "$JAR" >/dev/null; unzip -p "$JAR" META-INF/mods.toml | grep -F 'modId="rogues"' >/dev/null
python3 - "$JAR" <<'PY'
import struct,sys,zipfile
owned=0
with zipfile.ZipFile(sys.argv[1]) as z:
  for n in z.namelist():
    if not n.endswith('.class'): continue
    d=z.read(n); major=struct.unpack('>H',d[6:8])[0]
    if major>61: raise SystemExit(f'[Rogues] class newer than Java17: {n}={major}')
    if n.startswith('net/rogues/'): owned+=1
if not owned: raise SystemExit('[Rogues] no owned classes packaged')
print(f'[Rogues] Java17 package gate passed with {owned} owned classes')
PY
sha256sum "$JAR" | tee "$PORT/rogues-forge-1.20.1.sha256"; echo '[Rogues] FIRST_COMPILE_PACKAGE_PASS'
