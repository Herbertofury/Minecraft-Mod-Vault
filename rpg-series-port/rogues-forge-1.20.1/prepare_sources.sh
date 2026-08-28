#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
PORT="$ROOT/rpg-series-port/rogues-forge-1.20.1"
UPSTREAM="$PORT/.upstream"
CURRENT_REF="$UPSTREAM/rogues-current-3.1.1"
HIST_REF="$UPSTREAM/rogues-1.20.1-reference"
MATERIALIZED="$PORT/materialized-current"
MANIFEST="$PORT/materialized-source.sha256"
PROVENANCE="$PORT/materialized-provenance.txt"
TMP="${RUNNER_TEMP:-/tmp}/rogues-source-materialize.$$"

REPO="ZsoltMolnarrr/Rogues"
CURRENT_COMMIT="d4a7af565559dcff4384eabb2481f63eb5f97d55"
CURRENT_TREE="c6fac5d7c80807b41668274843c9732762d7cb03"
CURRENT_GRADLE_BLOB="5f264bbbaea8f43b53badb4eb6af24549b6211cf"
CURRENT_VERSION="3.1.1"
CURRENT_MC="1.21.1"
CURRENT_YARN="1.21.1+build.3"
HIST_COMMIT="bdfe6447b90758129e12430b497d97c181222b12"
HIST_TREE="0a4f7f94e77031732843f8a40a8460184bd3577a"
HIST_GRADLE_BLOB="1d1e4007bf335129cc6cabb115a0f78d3805fca0"
HIST_VERSION="1.2.0"
HIST_MC="1.20.1"
HIST_YARN="1.20.1+build.10"

cleanup(){ rm -rf "$TMP"; }
trap cleanup EXIT
fail(){ echo "[Rogues source prep] ERROR: $*" >&2; exit 1; }

git_blob_sha1(){
  local file="$1" size
  size="$(wc -c < "$file" | tr -d ' ')"
  { printf 'blob %s\0' "$size"; cat "$file"; } | sha1sum | awk '{print $1}'
}

assert_prop(){
  local file="$1" key="$2" expected="$3" actual
  actual="$(awk -F= -v wanted="$key" '
    /^[[:space:]]*#/ { next }
    {
      k=$1
      gsub(/^[[:space:]]+|[[:space:]]+$/, "", k)
      if (k == wanted) {
        $1=""
        sub(/^=/, "")
        v=$0
        gsub(/^[[:space:]]+|[[:space:]]+$/, "", v)
        print v
        exit
      }
    }' "$file")"
  [[ "$actual" = "$expected" ]] || fail "expected $key=$expected in $file, got ${actual:-<missing>}"
}

fetch_archive(){
  local commit="$1" dest="$2" label="$3"
  local archive="$TMP/$label.tar.gz" extract="$TMP/$label-extract" root
  mkdir -p "$extract"
  curl --fail --location --silent --show-error --retry 4 --retry-delay 2 \
    "https://codeload.github.com/$REPO/tar.gz/$commit" -o "$archive"
  tar -xzf "$archive" -C "$extract"
  root="$(find "$extract" -mindepth 1 -maxdepth 1 -type d -print -quit)"
  [[ -n "$root" && -f "$root/gradle.properties" ]] || fail "$label archive did not contain expected repository root"
  rm -rf "$dest"
  mkdir -p "$(dirname "$dest")"
  cp -a "$root" "$dest"
}

verify_tree_via_api(){
  local commit="$1" expected_tree="$2" label="$3" json="$TMP/$label-commit.json"
  curl --fail --location --silent --show-error --retry 4 --retry-delay 2 \
    -H 'Accept: application/vnd.github+json' \
    "https://api.github.com/repos/$REPO/git/commits/$commit" -o "$json"
  python3 - "$json" "$expected_tree" "$label" <<'PY'
import json, sys
p, expected, label = sys.argv[1:]
obj = json.load(open(p, encoding='utf-8'))
actual = obj.get('tree', {}).get('sha')
if actual != expected:
    raise SystemExit(f"[Rogues source prep] ERROR: {label} root tree {actual!r} != {expected}")
PY
}

verify_snapshot(){
  local dir="$1" blob="$2" version="$3" mc="$4" yarn="$5" label="$6"
  local gradle="$dir/gradle.properties" actual_blob
  [[ -f "$gradle" ]] || fail "$label gradle.properties missing"
  actual_blob="$(git_blob_sha1 "$gradle")"
  [[ "$actual_blob" = "$blob" ]] || fail "$label gradle.properties blob $actual_blob != $blob"
  assert_prop "$gradle" mod_version "$version"
  assert_prop "$gradle" minecraft_version "$mc"
  assert_prop "$gradle" yarn_mappings "$yarn"
  [[ -d "$dir/common/src/main/java" ]] || fail "$label common Java source missing"
  [[ -d "$dir/common/src/main/resources" ]] || fail "$label common resources missing"
}

deterministic_manifest(){
  local dir="$1" output="$2"
  ( cd "$dir" && find . -type f \
      ! -path './.git/*' ! -path './.gradle/*' ! -path '*/build/*' ! -path '*/run/*' ! -path '*/runs/*' \
      -print0 | LC_ALL=C sort -z | xargs -0 sha256sum ) > "$output"
}

rm -rf "$TMP"; mkdir -p "$TMP" "$UPSTREAM"
echo '[Rogues source prep] fetching immutable current 3.1.1 and historical 1.20.1 authorities'
verify_tree_via_api "$CURRENT_COMMIT" "$CURRENT_TREE" current
verify_tree_via_api "$HIST_COMMIT" "$HIST_TREE" historical
fetch_archive "$CURRENT_COMMIT" "$CURRENT_REF" current
fetch_archive "$HIST_COMMIT" "$HIST_REF" historical
verify_snapshot "$CURRENT_REF" "$CURRENT_GRADLE_BLOB" "$CURRENT_VERSION" "$CURRENT_MC" "$CURRENT_YARN" current
verify_snapshot "$HIST_REF" "$HIST_GRADLE_BLOB" "$HIST_VERSION" "$HIST_MC" "$HIST_YARN" historical
[[ -d "$CURRENT_REF/common/src/main/generated" ]] || fail 'current 3.1.1 generated data missing'
[[ -f "$CURRENT_REF/common/src/main/generated/data/rogues/spell/last_stand.json" ]] || fail 'current 3.1.1 Last Stand generated spell data missing'

rm -rf "$MATERIALIZED"
mkdir -p "$MATERIALIZED/common/src/main/java" "$MATERIALIZED/common/src/main/resources" "$MATERIALIZED/common/src/main/generated"
cp -a "$CURRENT_REF/common/src/main/java/." "$MATERIALIZED/common/src/main/java/"
cp -a "$CURRENT_REF/common/src/main/resources/." "$MATERIALIZED/common/src/main/resources/"
cp -a "$CURRENT_REF/common/src/main/generated/." "$MATERIALIZED/common/src/main/generated/"
for path in common/build.gradle common/gradle.properties build.gradle gradle.properties settings.gradle LICENSE; do
  if [[ -f "$CURRENT_REF/$path" ]]; then
    mkdir -p "$MATERIALIZED/$(dirname "$path")"
    cp -a "$CURRENT_REF/$path" "$MATERIALIZED/$path"
  fi
done
[[ -f "$MATERIALIZED/common/src/main/generated/data/rogues/spell/last_stand.json" ]] || fail 'materialized-current silently dropped generated spell data'

if find "$MATERIALIZED" -type f -exec grep -Il '' {} + 2>/dev/null | xargs -r grep -Fl "$HIST_COMMIT" >/dev/null; then
  fail 'historical commit material leaked into current materialization'
fi

deterministic_manifest "$MATERIALIZED" "$MANIFEST"
CURRENT_MANIFEST_SHA="$(sha256sum "$MANIFEST" | awk '{print $1}')"
CURRENT_FILE_COUNT="$(wc -l < "$MANIFEST" | tr -d ' ')"
CURRENT_REF_MANIFEST="$TMP/current-ref.sha256"
HIST_REF_MANIFEST="$TMP/historical-ref.sha256"
deterministic_manifest "$CURRENT_REF" "$CURRENT_REF_MANIFEST"
deterministic_manifest "$HIST_REF" "$HIST_REF_MANIFEST"
CURRENT_REF_SHA="$(sha256sum "$CURRENT_REF_MANIFEST" | awk '{print $1}')"
HIST_REF_SHA="$(sha256sum "$HIST_REF_MANIFEST" | awk '{print $1}')"

cat > "$PROVENANCE" <<EOF
repository=$REPO
current_commit=$CURRENT_COMMIT
current_tree=$CURRENT_TREE
current_gradle_blob=$CURRENT_GRADLE_BLOB
current_version=$CURRENT_VERSION
current_minecraft=$CURRENT_MC
current_yarn=$CURRENT_YARN
current_reference_manifest_sha256=$CURRENT_REF_SHA
historical_commit=$HIST_COMMIT
historical_tree=$HIST_TREE
historical_gradle_blob=$HIST_GRADLE_BLOB
historical_version=$HIST_VERSION
historical_minecraft=$HIST_MC
historical_yarn=$HIST_YARN
historical_reference_manifest_sha256=$HIST_REF_SHA
materialized_current_manifest_sha256=$CURRENT_MANIFEST_SHA
materialized_current_file_count=$CURRENT_FILE_COUNT
materialized_current_includes_generated=true
historical_overlay=false
EOF

SECOND="$TMP/materialized-second.sha256"
deterministic_manifest "$MATERIALIZED" "$SECOND"
cmp -s "$MANIFEST" "$SECOND" || fail 'materialized-current manifest changed within one preparation run'

echo "[Rogues source prep] CURRENT_AUTHORITY_PASS commit=$CURRENT_COMMIT tree=$CURRENT_TREE blob=$CURRENT_GRADLE_BLOB"
echo "[Rogues source prep] HISTORICAL_SUBSTRATE_PASS commit=$HIST_COMMIT tree=$HIST_TREE blob=$HIST_GRADLE_BLOB"
echo "[Rogues source prep] MATERIALIZED_CURRENT_PASS files=$CURRENT_FILE_COUNT manifest=$CURRENT_MANIFEST_SHA generated=true"
echo '[Rogues source prep] HISTORICAL_SEPARATION_PASS: historical 1.20.1 reference remains isolated under .upstream and never overlays current 3.1.1 source.'
