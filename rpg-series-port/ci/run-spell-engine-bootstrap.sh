#!/usr/bin/env bash
set -euo pipefail

ROOT="${GITHUB_WORKSPACE:?}"
UP="$ROOT/.rpg-upstream"
OUT="$ROOT/rpg-series-port/spell-engine-forge-1.20.1"
BASE_SHA=8721120169ddefd230fc73fc7c332318a92f6c7c
TARGET_SHA=bc02f7a49da950503010020da491f6bdc5871df7

rm -rf "$UP"
mkdir -p "$UP" "$OUT"

clone_exact() {
  local repo="$1" sha="$2" dest="$3"
  git init -q "$dest"
  git -C "$dest" remote add origin "https://github.com/${repo}.git"
  git -C "$dest" fetch -q --depth=1 origin "$sha"
  git -C "$dest" checkout -q --detach FETCH_HEAD
  test "$(git -C "$dest" rev-parse HEAD)" = "$sha"
}

clone_exact ZsoltMolnarrr/SpellEngine "$BASE_SHA" "$UP/spell-engine-1201"
clone_exact ZsoltMolnarrr/SpellEngine "$TARGET_SHA" "$UP/spell-engine-1102"

rm -f "$ROOT/spell-engine-0.15.12-1.20.1-source-ci.zip" "$ROOT/spell-engine-1.10.2-target-source-ci.zip"
(
  cd "$UP/spell-engine-1201"
  zip -qr "$ROOT/spell-engine-0.15.12-1.20.1-source-ci.zip" . -x '.git/*' 'build/*' '.gradle/*' '*.blend'
)
(
  cd "$UP/spell-engine-1102"
  zip -qr "$ROOT/spell-engine-1.10.2-target-source-ci.zip" . -x '.git/*' '*/build/*' '.gradle/*' '*.blend'
)

BASE_JAVA=$(find "$UP/spell-engine-1201" -type f -name '*.java' | wc -l)
TARGET_JAVA=$(find "$UP/spell-engine-1102" -type f -name '*.java' | wc -l)
BASE_FILES=$(find "$UP/spell-engine-1201" -type f ! -path '*/.git/*' | wc -l)
TARGET_FILES=$(find "$UP/spell-engine-1102" -type f ! -path '*/.git/*' | wc -l)
POST120_IMPORTS=$(grep -RhoE '^import net\.minecraft\.(component|network\.packet\.CustomPayload|registry\.entry)[^;]*;' "$UP/spell-engine-1102/common/src/main/java" 2>/dev/null | sort -u | wc -l || true)

cat > "$OUT/bootstrap-report.txt" <<EOF
Spell Engine native Forge 1.20.1 bootstrap
base_sha=$BASE_SHA
target_sha=$TARGET_SHA
base_java_files=$BASE_JAVA
target_java_files=$TARGET_JAVA
base_total_files=$BASE_FILES
target_total_files=$TARGET_FILES
post_1_20_style_imports=$POST120_IMPORTS
EOF
cat "$OUT/bootstrap-report.txt"

unzip -t "$ROOT/spell-engine-0.15.12-1.20.1-source-ci.zip" >/dev/null
unzip -t "$ROOT/spell-engine-1.10.2-target-source-ci.zip" >/dev/null
sha256sum "$ROOT/spell-engine-0.15.12-1.20.1-source-ci.zip" "$ROOT/spell-engine-1.10.2-target-source-ci.zip"

echo '[Spell Engine bootstrap] Exact upstream source snapshots prepared and verified.'
