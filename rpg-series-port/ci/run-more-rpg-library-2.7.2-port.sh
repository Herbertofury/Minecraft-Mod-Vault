#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
PORT="$ROOT/rpg-series-port/more-rpg-library-forge-1.20.1"
ENV="$PORT/MORE_RPG_LIBRARY_PORT.env"
WORK="$ROOT/.more-rpg-library-build"
UP="$ROOT/.more-rpg-library-upstream"
FOUNDATION="$PORT/.foundation"
source "$ENV"
echo '[More RPG 2.7.2] ACTIVE_NATIVE_FORGE_1201_GRADUATION_LANE'

rm -rf "$WORK" "$UP" "$FOUNDATION"
mkdir -p "$UP" "$FOUNDATION"
clone_exact() {
  local repo="$1" sha="$2" dest="$3"
  git init -q "$dest"
  git -C "$dest" remote add origin "https://github.com/${repo}.git"
  git -C "$dest" fetch -q --depth=1 origin "$sha"
  git -C "$dest" checkout -q --detach FETCH_HEAD
  test "$(git -C "$dest" rev-parse HEAD)" = "$sha"
}
clone_exact ProfessorFichte/More-RPG-Library "$MORE_RPG_LIBRARY_1201_SUBSTRATE_SHA" "$UP/old-1201" & p1=$!
clone_exact ProfessorFichte/More-RPG-Library "$MORE_RPG_LIBRARY_272_TARGET_SHA" "$UP/modern-272" & p2=$!
wait "$p1" "$p2"
echo '[More RPG 2.7.2] EXACT_UPSTREAM_AUTHORITIES_READY'

# Replay the already-graduated Spell Engine lane once as the fail-closed foundation producer. Its
# dependent Spell Power/Ranged Gradle outputs contain Architectury-generated nonces, so downstream
# runtime consumers pass those raw bytes through their owning exact certifiers before hash comparison.
bash "$ROOT/rpg-series-port/ci/run-spell-engine-1.10.4-exact-seal-graduation.sh"

SPELL_ENGINE_JAR="$ROOT/rpg-series-port/spell-engine-forge-1.20.1/spell_engine-forge-1.10.4+1.20.1.jar"
SPELL_POWER_RAW="$(find "$ROOT/rpg-series-port/spell_power-forge-1.20.1/forge/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' ! -name '*dev-shadow*' | sort | head -n1)"
RANGED_RAW="$(find "$ROOT/rpg-series-port/ranged-weapon-api-forge-1.20.1/forge/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' ! -name '*dev-shadow*' | sort | head -n1)"
TINY_JAR="$(find "$ROOT/rpg-series-port/tiny-config-forge-1.20.1/generated/forge/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' ! -name '*dev-shadow*' | sort | head -n1)"
SPELL_POWER_JAR="$FOUNDATION/spell_power-forge-1.6.0+1.20.1-certified.jar"
RANGED_JAR="$FOUNDATION/ranged_weapon_api-forge-2.3.4+1.20.1-certified.jar"
for f in "$SPELL_ENGINE_JAR" "$SPELL_POWER_RAW" "$RANGED_RAW" "$TINY_JAR"; do test -f "$f"; unzip -tq "$f" >/dev/null; done

python3 "$ROOT/rpg-series-port/ci/certify-spell-power-current.py" "$SPELL_POWER_RAW" "$SPELL_POWER_JAR"
python3 "$ROOT/rpg-series-port/ci/certify-ranged-weapon-api-run268.py" "$RANGED_RAW" "$RANGED_JAR"
for f in "$SPELL_POWER_JAR" "$RANGED_JAR"; do test -f "$f"; unzip -tq "$f" >/dev/null; done

test "$(sha256sum "$SPELL_ENGINE_JAR" | awk '{print $1}')" = "$SPELL_ENGINE_1104_EXPECTED_JAR_SHA"
test "$(sha256sum "$SPELL_POWER_JAR" | awk '{print $1}')" = "$SPELL_POWER_160_EXPECTED_JAR_SHA"
test "$(sha256sum "$RANGED_JAR" | awk '{print $1}')" = "$RANGED_WEAPON_API_234_EXPECTED_JAR_SHA"
test "$(sha256sum "$TINY_JAR" | awk '{print $1}')" = "$TINY_CONFIG_310_EXPECTED_JAR_SHA"
echo '[More RPG 2.7.2] CERTIFIED_RPG_FOUNDATIONS_READY spell_engine=1.10.4 spell_power=current-tinyconfig-3.1 ranged=2.3.4 tiny_config=3.1.0'

# Forge packaging does not guarantee that each Architectury common project's Jar task has run. Build
# the named common artifacts explicitly from the exact replay workspace before discovering them. These
# are compile-only namespace providers; certified production Forge jars above remain release/runtime authority.
gradle --no-daemon -p "$ROOT/.spell-engine-build" :common:jar
gradle --no-daemon -p "$ROOT/rpg-series-port/spell_power-forge-1.20.1" :common:jar
gradle --no-daemon -p "$ROOT/rpg-series-port/ranged-weapon-api-forge-1.20.1" :common:jar
gradle --no-daemon -p "$ROOT/rpg-series-port/tiny-config-forge-1.20.1/generated" :common:jar
echo '[More RPG 2.7.2] NAMED_COMMON_FOUNDATIONS_MATERIALIZED'

# Common/Yarn source must compile against named common artifacts, never final Forge/SRG production
# jars. They are produced by the exact same foundation replay above; the certified production JARs
# remain the only runtime/release inputs on the Forge side.
SPELL_ENGINE_COMMON_JAR="$(find "$ROOT/.spell-engine-build/common/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' | sort | head -n1)"
SPELL_POWER_COMMON_JAR="$(find "$ROOT/rpg-series-port/spell_power-forge-1.20.1/common/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' | sort | head -n1)"
RANGED_COMMON_JAR="$(find "$ROOT/rpg-series-port/ranged-weapon-api-forge-1.20.1/common/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' | sort | head -n1)"
TINY_COMMON_JAR="$(find "$ROOT/rpg-series-port/tiny-config-forge-1.20.1/generated/common/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' | sort | head -n1)"
for f in "$SPELL_ENGINE_COMMON_JAR" "$SPELL_POWER_COMMON_JAR" "$RANGED_COMMON_JAR" "$TINY_COMMON_JAR"; do test -f "$f"; unzip -tq "$f" >/dev/null; done

# Fail closed on the exact namespace poison that owned run #325. Named Spell Power's vulnerability
# effect must reference Yarn's StatusEffect and must not retain Mojmap's MobEffect path.
SPELL_POWER_EFFECT='net/spell_power/api/effect/SpellVulnerabilityStatusEffect.class'
unzip -Z1 "$SPELL_POWER_COMMON_JAR" | grep -Fx "$SPELL_POWER_EFFECT" >/dev/null
unzip -p "$SPELL_POWER_COMMON_JAR" "$SPELL_POWER_EFFECT" | strings | grep -F 'net/minecraft/entity/effect/StatusEffect' >/dev/null
if unzip -p "$SPELL_POWER_COMMON_JAR" "$SPELL_POWER_EFFECT" | strings | grep -Fq 'net/minecraft/world/effect/MobEffect'; then
  echo '[More RPG 2.7.2] Mojmap MobEffect leaked into named Spell Power common artifact' >&2; exit 1
fi
printf '[More RPG 2.7.2] NAMED_FOUNDATION_HASH spell_engine=%s spell_power=%s ranged=%s tiny_config=%s\n' \
  "$(sha256sum "$SPELL_ENGINE_COMMON_JAR" | awk '{print $1}')" \
  "$(sha256sum "$SPELL_POWER_COMMON_JAR" | awk '{print $1}')" \
  "$(sha256sum "$RANGED_COMMON_JAR" | awk '{print $1}')" \
  "$(sha256sum "$TINY_COMMON_JAR" | awk '{print $1}')"

python3 "$PORT/tools/prepare_more_rpg_library_named_common.py" "$UP/modern-272" "$UP/old-1201" "$WORK"

(
  cd "$WORK"
  find common forge -type f -print0 | sort -z | xargs -0 sha256sum
) > "$PORT/more-rpg-library-source-manifest.sha256"
SOURCE_SHA="$(sha256sum "$PORT/more-rpg-library-source-manifest.sha256" | awk '{print $1}')"
echo "$SOURCE_SHA  more-rpg-library-source-manifest.sha256" | tee "$PORT/more-rpg-library-source.sha256"

ARGS=(
  "-Pspell_engine_common_jar=$SPELL_ENGINE_COMMON_JAR"
  "-Pspell_power_common_jar=$SPELL_POWER_COMMON_JAR"
  "-Pranged_weapon_api_common_jar=$RANGED_COMMON_JAR"
  "-Ptiny_config_common_jar=$TINY_COMMON_JAR"
  "-Pspell_engine_forge_jar=$SPELL_ENGINE_JAR"
  "-Pspell_power_forge_jar=$SPELL_POWER_JAR"
  "-Pranged_weapon_api_forge_jar=$RANGED_JAR"
  "-Ptiny_config_forge_jar=$TINY_JAR"
)

gradle --no-daemon --stacktrace -p "$WORK" :forge:compileJava "${ARGS[@]}"
echo '[More RPG 2.7.2] FULL_MODERN_COMMON_NATIVE_FORGE_COMPILE_PASS'

gradle --no-daemon --stacktrace -p "$WORK" :forge:build "${ARGS[@]}"
JAR="$(find "$WORK/forge/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' ! -name '*dev-shadow*' | sort | head -n1)"
test -f "$JAR"; unzip -tq "$JAR" >/dev/null
if unzip -Z1 "$JAR" | grep -E '(^fabric\.mod\.json$|META-INF/neoforge\.mods\.toml|^net/fabricmc/|^net/neoforged/)'; then
  echo '[More RPG 2.7.2] forbidden loader leakage in packaged Forge JAR' >&2
  exit 1
fi
unzip -p "$JAR" META-INF/mods.toml | grep -F 'modId="more_rpg_classes"' >/dev/null
OUT="$PORT/more_rpg_library-forge-2.7.2+1.20.1.jar"
cp "$JAR" "$OUT"
sha256sum "$OUT" | tee "$PORT/more-rpg-library-forge-1.20.1.sha256"
echo '[More RPG 2.7.2] FIRST_NATIVE_FORGE_PACKAGE_PASS runtime-graduation-pending'
