#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")" && pwd)"; source "$ROOT/UPSTREAM_PINS.env"; UP="$ROOT/.upstream"; OUT="$ROOT/generated"; CURRENT="$UP/current-$ROGUES_CURRENT_SHA"; LEGACY="$UP/legacy-$ROGUES_LEGACY_SHA"
clone_exact(){ local sha="$1" dst="$2"; rm -rf "$dst"; git init -q "$dst"; git -C "$dst" remote add origin https://github.com/ZsoltMolnarrr/Rogues.git; git -C "$dst" fetch -q --depth=1 origin "$sha"; git -C "$dst" checkout -q --detach FETCH_HEAD; [[ "$(git -C "$dst" rev-parse HEAD)" = "$sha" ]]; }
blob(){ git hash-object "$1"; }; require(){ [[ "$(blob "$1")" = "$2" ]] || { echo "[Rogues source] blob mismatch $1" >&2; exit 2; }; }
rm -rf "$UP" "$OUT"; mkdir -p "$UP"; clone_exact "$ROGUES_CURRENT_SHA" "$CURRENT" & a=$!; clone_exact "$ROGUES_LEGACY_SHA" "$LEGACY" & b=$!; wait "$a" "$b"
[[ "$(git -C "$CURRENT" rev-parse HEAD^{tree})" = "$ROGUES_CURRENT_TREE" ]]; [[ "$(git -C "$LEGACY" rev-parse HEAD^{tree})" = "$ROGUES_LEGACY_TREE" ]]
require "$CURRENT/common/src/main/java/net/rogues/RoguesMod.java" "$ROGUES_CURRENT_MOD_BLOB"; require "$CURRENT/gradle.properties" "$ROGUES_CURRENT_GRADLE_BLOB"; require "$LEGACY/src/main/java/net/rogues/RoguesMod.java" "$ROGUES_LEGACY_MOD_BLOB"; require "$LEGACY/gradle.properties" "$ROGUES_LEGACY_GRADLE_BLOB"
grep -Fx 'mod_version=3.1.1' "$CURRENT/gradle.properties" >/dev/null; grep -Fx 'minecraft_version=1.20.1' "$LEGACY/gradle.properties" >/dev/null
bash "$ROOT/prepare_sources.sh" "$CURRENT" "$LEGACY" "$OUT"; python3 "$ROOT/apply_1201_forge_registration.py" "$OUT/common/java"; python3 "$ROOT/apply_1201_compat.py" "$OUT/common/java"; python3 "$ROOT/apply_1201_compat_batch2.py" "$OUT/common/java"; python3 "$ROOT/apply_1201_compat_batch3.py" "$OUT/common/java"; python3 "$ROOT/apply_1201_compat_batch4.py" "$OUT/common/java"; python3 "$ROOT/apply_1201_compat_batch5.py" "$OUT/common/java" "$OUT/common/resources"
if grep -R -nE 'Registry\.register(Reference)?\(' "$OUT/common/java"; then echo '[Rogues] unbridged registry mutation' >&2; exit 2; fi
(cd "$OUT"; find common -type f -print0 | LC_ALL=C sort -z | xargs -0 sha256sum > CURRENT_PORT_OUTPUTS.sha256)
echo "[Rogues source] exact current=$ROGUES_CURRENT_SHA legacy=$ROGUES_LEGACY_SHA; current feature authority translated, legacy reference-only"
