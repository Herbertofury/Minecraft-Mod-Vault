#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
PORT="$ROOT/rpg-series-port/more-rpg-library-forge-1.20.1"
ENV="$PORT/MORE_RPG_LIBRARY_PORT.env"
WORK="$ROOT/.more-rpg-library-build"
UP="$ROOT/.more-rpg-library-upstream"
source "$ENV"

rm -rf "$WORK" "$UP"
mkdir -p "$UP"
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

# Rebuild and replay the already-graduated Spell Engine lane to materialize exact dependency jars in
# this clean CI workspace. This is intentionally expensive but fail-closed until a separately verified
# artifact cache is added; More RPG must never compile against a moving Maven substitute.
bash "$ROOT/rpg-series-port/ci/run-spell-engine-1.10.4-exact-seal-graduation.sh"

SPELL_ENGINE_JAR="$ROOT/rpg-series-port/spell-engine-forge-1.20.1/spell_engine-forge-1.10.4+1.20.1.jar"
SPELL_POWER_JAR="$(find "$ROOT/rpg-series-port/spell_power-forge-1.20.1/forge/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' ! -name '*dev-shadow*' | sort | head -n1)"
RANGED_JAR="$(find "$ROOT/rpg-series-port/ranged-weapon-api-forge-1.20.1/forge/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' ! -name '*dev-shadow*' | sort | head -n1)"
TINY_JAR="$(find "$ROOT/rpg-series-port/tiny-config-forge-1.20.1/generated/forge/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' ! -name '*dev-shadow*' | sort | head -n1)"
for f in "$SPELL_ENGINE_JAR" "$SPELL_POWER_JAR" "$RANGED_JAR" "$TINY_JAR"; do test -f "$f"; unzip -tq "$f" >/dev/null; done

test "$(sha256sum "$SPELL_ENGINE_JAR" | awk '{print $1}')" = "$SPELL_ENGINE_1104_EXPECTED_JAR_SHA"
test "$(sha256sum "$SPELL_POWER_JAR" | awk '{print $1}')" = "$SPELL_POWER_160_EXPECTED_JAR_SHA"
test "$(sha256sum "$RANGED_JAR" | awk '{print $1}')" = "$RANGED_WEAPON_API_234_EXPECTED_JAR_SHA"
test "$(sha256sum "$TINY_JAR" | awk '{print $1}')" = "$TINY_CONFIG_310_EXPECTED_JAR_SHA"
echo '[More RPG 2.7.2] CERTIFIED_RPG_FOUNDATIONS_READY'

python3 "$PORT/tools/prepare_more_rpg_library.py" "$UP/modern-272" "$UP/old-1201" "$WORK"

# Deterministic source identity over all source/resource files that can affect the port. Git metadata,
# Gradle caches and generated build output are excluded by construction because WORK was just created.
(
  cd "$WORK"
  find common forge -type f -print0 | sort -z | xargs -0 sha256sum
) > "$PORT/more-rpg-library-source-manifest.sha256"
SOURCE_SHA="$(sha256sum "$PORT/more-rpg-library-source-manifest.sha256" | awk '{print $1}')"
echo "$SOURCE_SHA  more-rpg-library-source-manifest.sha256" | tee "$PORT/more-rpg-library-source.sha256"

ARGS=(
  "-Pspell_engine_forge_jar=$SPELL_ENGINE_JAR"
  "-Pspell_power_forge_jar=$SPELL_POWER_JAR"
  "-Pranged_weapon_api_forge_jar=$RANGED_JAR"
  "-Ptiny_config_forge_jar=$TINY_JAR"
)

# The entire modern common tree compiles here. Any 1.21 -> 1.20.1 API mismatch is a real port failure
# and becomes the next compatibility-pass owner; do not reduce the source set to manufacture green.
gradle --no-daemon --stacktrace -p "$WORK" :forge:compileJava "${ARGS[@]}"
echo '[More RPG 2.7.2] FULL_MODERN_COMMON_NATIVE_FORGE_COMPILE_PASS'

# Once compilation is green, package the untouched native Forge candidate. Runtime gates are added only
# after loader/registry adaptation is complete, but packaging boundaries are enforced immediately.
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
