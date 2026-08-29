#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")" && pwd)"
# shellcheck disable=SC1091
source "$ROOT/UPSTREAM_PINS.env"
OUT="${1:-$ROOT/.upstream}"
CURRENT="$OUT/current-$ARCHERS_CURRENT_SHA"
LEGACY="$OUT/legacy-1.20.1-$ARCHERS_LEGACY_1201_SHA"

fetch_commit() {
  local sha="$1" dest="$2" archive="$3" tmp="$4"
  rm -rf "$dest" "$tmp"
  mkdir -p "$tmp" "$OUT"
  curl -fsSL "https://api.github.com/repos/ZsoltMolnarrr/Archers/tarball/$sha" -o "$archive"
  tar -xzf "$archive" -C "$tmp"
  local top
  top="$(find "$tmp" -mindepth 1 -maxdepth 1 -type d | head -n1)"
  [[ -n "$top" ]]
  mv "$top" "$dest"
  rmdir "$tmp"
}

blob_sha() { git hash-object "$1"; }
require_blob() {
  local file="$1" expected="$2" label="$3" actual
  [[ -f "$file" ]] || { echo "[Archers source prep] missing $label: $file" >&2; exit 1; }
  actual="$(blob_sha "$file")"
  [[ "$actual" = "$expected" ]] || {
    echo "[Archers source prep] $label blob mismatch: expected $expected, got $actual" >&2
    exit 1
  }
}

CURRENT_ARCHIVE="$OUT/current-$ARCHERS_CURRENT_SHA.tar.gz"
LEGACY_ARCHIVE="$OUT/legacy-$ARCHERS_LEGACY_1201_SHA.tar.gz"
fetch_commit "$ARCHERS_CURRENT_SHA" "$CURRENT" "$CURRENT_ARCHIVE" "$OUT/.extract-current"
fetch_commit "$ARCHERS_LEGACY_1201_SHA" "$LEGACY" "$LEGACY_ARCHIVE" "$OUT/.extract-legacy"

# Current 3.1.1 is the content/feature authority and uses the modern multi-platform layout.
require_blob "$CURRENT/gradle.properties" "$ARCHERS_CURRENT_GRADLE_PROPERTIES_BLOB" 'current gradle.properties'
require_blob "$CURRENT/common/src/main/java/net/archers/ArchersMod.java" "$ARCHERS_CURRENT_ARCHERSMOD_BLOB" 'current ArchersMod.java'
require_blob "$CURRENT/common/src/main/java/net/archers/item/Quivers.java" "$ARCHERS_CURRENT_QUIVERS_BLOB" 'current Quivers.java'
grep -Fx 'mod_version=3.1.1' "$CURRENT/gradle.properties" >/dev/null
grep -Fx 'minecraft_version=1.21.1' "$CURRENT/gradle.properties" >/dev/null

# Historical 1.20.1 is mapping/API substrate only and intentionally has a single-module src/main layout.
require_blob "$LEGACY/gradle.properties" "$ARCHERS_LEGACY_1201_GRADLE_PROPERTIES_BLOB" 'legacy gradle.properties'
require_blob "$LEGACY/src/main/java/net/archers/ArchersMod.java" "$ARCHERS_LEGACY_1201_ARCHERSMOD_BLOB" 'legacy ArchersMod.java'
grep -Fx 'minecraft_version=1.20.1' "$LEGACY/gradle.properties" >/dev/null
if find "$LEGACY" -path '*/Quivers.java' -print -quit | grep -q .; then
  echo '[Archers source prep] historical 1.20.1 unexpectedly contains Quivers.java; provenance assumption changed' >&2
  exit 1
fi

manifest() {
  local tree="$1" out="$2"
  (cd "$tree" && find . -type f ! -path './.git/*' -print0 | sort -z | xargs -0 sha256sum) > "$out"
}
manifest "$CURRENT" "$OUT/current-$ARCHERS_CURRENT_SHA.files.sha256"
manifest "$LEGACY" "$OUT/legacy-$ARCHERS_LEGACY_1201_SHA.files.sha256"
sha256sum "$CURRENT_ARCHIVE" "$LEGACY_ARCHIVE" > "$OUT/upstream-archives.sha256"

CURRENT_FILES="$(wc -l < "$OUT/current-$ARCHERS_CURRENT_SHA.files.sha256" | tr -d ' ')"
LEGACY_FILES="$(wc -l < "$OUT/legacy-$ARCHERS_LEGACY_1201_SHA.files.sha256" | tr -d ' ')"
printf '[Archers source prep] current=%s tree=%s files=%s\n' "$ARCHERS_CURRENT_SHA" "$ARCHERS_CURRENT_TREE" "$CURRENT_FILES"
printf '[Archers source prep] legacy=%s tree=%s files=%s\n' "$ARCHERS_LEGACY_1201_SHA" "$ARCHERS_LEGACY_1201_TREE" "$LEGACY_FILES"
echo '[Archers source prep] exact upstream snapshots verified; current remains feature authority, legacy remains mapping-only.'
