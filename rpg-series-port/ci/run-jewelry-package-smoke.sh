#!/usr/bin/env bash
set -euo pipefail

OUT_JAR="${1:?jewelry jar}"
STRUCTURE_JAR="${2:?structure pool jar}"
SPELL_POWER_JAR="${3:?spell power jar}"
RANGED_JAR="${4:?ranged weapon api jar}"
WORK="${5:?generated jewelry root}"
PORT="${6:?jewelry port evidence dir}"

find_module_jar() {
  local group="$1"
  local artifact="$2"
  local version="$3"
  local root="$HOME/.gradle/caches/modules-2/files-2.1/$group/$artifact/$version"
  local jar
  jar="$(find "$root" -type f -name "${artifact}-*.jar" ! -name '*sources*' ! -name '*javadoc*' | head -n 1 || true)"
  [[ -n "$jar" && -f "$jar" ]] || { echo "Missing resolved module JAR: $group:$artifact:$version" >&2; exit 1; }
  printf '%s\n' "$jar"
}
prop() { sed -n "s/^${1}=//p" "$WORK/gradle.properties" | tail -n 1 | tr -d '\r'; }
CURIO_VERSION="$(prop curios_version)"
[[ -n "$CURIO_VERSION" ]] || { echo 'Missing Curios version from generated Jewelry properties' >&2; exit 1; }
CURIO_JAR="$(find_module_jar top.theillusivec4.curios curios-forge "$CURIO_VERSION")"

BASE="$WORK/.fresh-jewelry-forge-server-base"
rm -rf "$BASE" "$WORK/.fresh-jewelry-no-curios" "$WORK/.fresh-jewelry-with-curios"
mkdir -p "$BASE"
curl -fsSL "https://maven.minecraftforge.net/net/minecraftforge/forge/1.20.1-47.4.23/forge-1.20.1-47.4.23-installer.jar" -o "$BASE/forge-installer.jar"
(
  cd "$BASE"
  java -jar forge-installer.jar --installServer >/dev/null
  printf 'eula=true\n' > eula.txt
  cat > server.properties <<'PROPS'
level-type=minecraft:flat
generate-structures=false
view-distance=3
simulation-distance=3
spawn-protection=0
PROPS
)

prepare_variant() {
  local dir="$1"
  local expect_curios="$2"
  rm -rf "$dir"
  cp -a "$BASE" "$dir"
  mkdir -p "$dir/mods"
  cp "$OUT_JAR" "$STRUCTURE_JAR" "$SPELL_POWER_JAR" "$RANGED_JAR" "$dir/mods/"
  if [[ "$expect_curios" == true ]]; then
    cp "$CURIO_JAR" "$dir/mods/"
  fi
  cat > "$dir/user_jvm_args.txt" <<EOF
-Xmx2G
-Djewelry.ci.selftest=true
-Djewelry.ci.expectCurios=$expect_curios
EOF
}

stop_tree() {
  local root="$1" child
  local kids="$(pgrep -P "$root" 2>/dev/null || true)"
  for child in $kids; do kill -TERM "$child" 2>/dev/null || true; done
  kill -TERM "$root" 2>/dev/null || true
  sleep 1
  kids="$(pgrep -P "$root" 2>/dev/null || true)"
  for child in $kids; do kill -KILL "$child" 2>/dev/null || true; done
  kill -KILL "$root" 2>/dev/null || true
  wait "$root" 2>/dev/null || true
}

run_variant() {
  local dir="$1"
  local label="$2"
  local expect_curios="$3"
  local log="$PORT/forge-package-${label}-smoke.log"
  : > "$log"
  (
    cd "$dir"
    exec ./run.sh nogui
  ) > "$log" 2>&1 &
  local pid=$!
  local deadline=$((SECONDS+180)) pass=0
  local fatal='Jewelry CI self-test failed|ModLoadingException|Failed to create mod instance|Failed to start the minecraft server|NoClassDefFoundError|ClassNotFoundException|MixinApplyError|InvalidMixinException|MixinTransformerError|Registry is already frozen|Can not register to a locked registry|Missing or unsupported mandatory dependencies|Exception in server tick loop'
  while ((SECONDS<deadline)); do
    local latest="$dir/logs/latest.log"
    local files=("$log"); [[ -f "$latest" ]] && files+=("$latest")
    if grep -Eiq "$fatal" "${files[@]}"; then
      tail -n 500 "${files[@]}" || true
      stop_tree "$pid"
      return 1
    fi
    if [[ -f "$latest" ]] \
      && grep -Eq 'Done \([0-9.]+s\)!' "$latest" \
      && grep -Fq "[Jewelry CI] Packaged runtime self-test passed:" "${files[@]}"; then
      pass=1
      break
    fi
    if ! kill -0 "$pid" 2>/dev/null; then
      wait "$pid" || true
      tail -n 500 "${files[@]}" || true
      return 1
    fi
    sleep 1
  done
  if [[ "$pass" -ne 1 ]]; then
    tail -n 500 "$log" || true
    stop_tree "$pid"
    return 1
  fi
  if [[ "$expect_curios" == true ]]; then
    grep -Fq 'curios=true' "$log" "$dir/logs/latest.log"
  else
    grep -Fq 'curios=false' "$log" "$dir/logs/latest.log"
  fi
  stop_tree "$pid"
  printf '[Jewelry CI] Fresh packaged Forge server bootstrap passed: %s.\n' "$label"
}

NO_CURIOS="$WORK/.fresh-jewelry-no-curios"
WITH_CURIOS="$WORK/.fresh-jewelry-with-curios"
prepare_variant "$NO_CURIOS" false
prepare_variant "$WITH_CURIOS" true
run_variant "$NO_CURIOS" no-curios false
run_variant "$WITH_CURIOS" with-curios true

printf '[Jewelry CI] Packaged server matrix passed with Curios absent and present.\n'
