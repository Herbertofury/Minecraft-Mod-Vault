#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")" && pwd)"
# shellcheck disable=SC1091
source "$ROOT/UPSTREAM_PINS.env"
UP="$ROOT/.upstream"
OUT="$ROOT/generated"
CURRENT="$UP/current-$PALADINS_CURRENT_SHA"
LEGACY="$UP/legacy-1.20.1-$PALADINS_LEGACY_1201_SHA"

clone_exact() {
  local sha="$1" dest="$2"
  rm -rf "$dest"
  git init -q "$dest"
  git -C "$dest" remote add origin https://github.com/ZsoltMolnarrr/Paladins.git
  git -C "$dest" fetch -q --depth=1 origin "$sha"
  git -C "$dest" checkout -q --detach FETCH_HEAD
  [[ "$(git -C "$dest" rev-parse HEAD)" = "$sha" ]]
}
blob_sha() { git hash-object "$1"; }
require_blob() {
  local file="$1" expected="$2" label="$3" actual
  [[ -f "$file" ]] || { echo "[Paladins source prep] missing $label: $file" >&2; exit 1; }
  actual="$(blob_sha "$file")"
  [[ "$actual" = "$expected" ]] || {
    echo "[Paladins source prep] $label blob mismatch: expected $expected, got $actual" >&2
    exit 1
  }
}

rm -rf "$UP" "$OUT"; mkdir -p "$UP"
clone_exact "$PALADINS_CURRENT_SHA" "$CURRENT" & P1=$!
clone_exact "$PALADINS_LEGACY_1201_SHA" "$LEGACY" & P2=$!
wait "$P1" "$P2"

[[ "$(git -C "$CURRENT" rev-parse HEAD^{tree})" = "$PALADINS_CURRENT_TREE" ]] || {
  echo '[Paladins source prep] current tree hash mismatch' >&2; exit 1; }
[[ "$(git -C "$LEGACY" rev-parse HEAD^{tree})" = "$PALADINS_LEGACY_1201_TREE" ]] || {
  echo '[Paladins source prep] legacy tree hash mismatch' >&2; exit 1; }

require_blob "$CURRENT/common/src/main/java/net/paladins/PaladinsMod.java" "$PALADINS_CURRENT_PALADINSMOD_BLOB" 'current PaladinsMod.java'
require_blob "$CURRENT/gradle.properties" "$PALADINS_CURRENT_GRADLE_PROPERTIES_BLOB" 'current gradle.properties'
require_blob "$LEGACY/src/main/java/net/paladins/PaladinsMod.java" "$PALADINS_LEGACY_1201_PALADINSMOD_BLOB" 'legacy PaladinsMod.java'
require_blob "$LEGACY/gradle.properties" "$PALADINS_LEGACY_1201_GRADLE_PROPERTIES_BLOB" 'legacy gradle.properties'
grep -Fx 'mod_version=3.1.1' "$CURRENT/gradle.properties" >/dev/null
grep -Fx 'minecraft_version=1.21.1' "$CURRENT/gradle.properties" >/dev/null
grep -Fx 'minecraft_version=1.20.1' "$LEGACY/gradle.properties" >/dev/null

bash "$ROOT/prepare_sources.sh" "$CURRENT" "$LEGACY" "$OUT"
python3 "$ROOT/apply_1201_forge_registration.py" "$OUT/common/java"
python3 "$ROOT/apply_1201_api_compat.py" "$OUT/common/java"

# Registration transform acceptance: zero direct vanilla mutations, explicit split lifecycle hooks.
if grep -R -nE '(^|[^A-Za-z0-9_.])Registry\.register(Reference)?\(' "$OUT/common/java"; then
  echo '[Paladins materialize] unbridged vanilla registry mutation survived native Forge transform' >&2
  exit 2
fi
grep -F 'public static void registerBlocks()' "$OUT/common/java/net/paladins/block/PaladinBlocks.java" >/dev/null
grep -F 'public static void registerItems()' "$OUT/common/java/net/paladins/block/PaladinBlocks.java" >/dev/null
grep -F 'public static void registerBlockItems()' "$OUT/common/java/net/paladins/PaladinsMod.java" >/dev/null
grep -F 'public static void registerItemGroup()' "$OUT/common/java/net/paladins/PaladinsMod.java" >/dev/null
grep -F 'public static void registerEntities()' "$OUT/common/java/net/paladins/PaladinsMod.java" >/dev/null
if grep -R -nF 'Identifier.of(' "$OUT/common/java"; then
  echo '[Paladins materialize] 1.21 Identifier.of API survived 1.20.1 compatibility pass' >&2
  exit 2
fi

manifest() {
  local tree="$1" out="$2"
  (cd "$tree" && find . -type f ! -path './.git/*' -print0 | LC_ALL=C sort -z | xargs -0 sha256sum) > "$out"
}
manifest "$CURRENT" "$UP/current-$PALADINS_CURRENT_SHA.files.sha256"
manifest "$LEGACY" "$UP/legacy-$PALADINS_LEGACY_1201_SHA.files.sha256"

(
  cd "$OUT"
  find common -type f -print0 | LC_ALL=C sort -z | xargs -0 sha256sum > CURRENT_PORT_OUTPUTS.sha256
)

printf '[Paladins source prep] current=%s tree=%s files=%s\n' \
  "$PALADINS_CURRENT_SHA" "$PALADINS_CURRENT_TREE" "$(wc -l < "$UP/current-$PALADINS_CURRENT_SHA.files.sha256" | tr -d ' ')"
printf '[Paladins source prep] legacy=%s tree=%s files=%s\n' \
  "$PALADINS_LEGACY_1201_SHA" "$PALADINS_LEGACY_1201_TREE" "$(wc -l < "$UP/legacy-$PALADINS_LEGACY_1201_SHA.files.sha256" | tr -d ' ')"
echo '[Paladins source prep] exact two-pin snapshots verified; current feature authority staged with native Forge registration ownership and exact 1.20.1 API syntax translations; legacy remains API/mapping substrate only.'
