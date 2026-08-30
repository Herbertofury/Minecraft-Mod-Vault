#!/usr/bin/env bash
set -euo pipefail
ROOT="${GITHUB_WORKSPACE:?}"
PORT="$ROOT/rpg-series-port/more-rpg-library-forge-1.20.1"
RUN="$ROOT/.more-rpg-library-build/forge/run"
MIXIN_OUT="$RUN/.mixin.out"
EVIDENCE="$PORT/transformed-hud-evidence"
[[ -d "$MIXIN_OUT" ]] || { echo '[More RPG 2.7.2] Mixin export directory missing for HUD evidence preservation' >&2; exit 1; }
mapfile -t TARGETS < <(grep -aRl 'net/more_rpg_classes/client/heart/HeartRegistry' "$MIXIN_OUT" --include='*.class' 2>/dev/null | sort || true)
((${#TARGETS[@]} > 0)) || { echo '[More RPG 2.7.2] no transformed HUD target contains HeartRegistry' >&2; exit 1; }
rm -rf "$EVIDENCE"
mkdir -p "$EVIDENCE"
TARGET="${TARGETS[0]}"
REL="${TARGET#$MIXIN_OUT/}"
cp "$TARGET" "$EVIDENCE/more-rpg-transformed-hud-target.class"
SHA="$(sha256sum "$EVIDENCE/more-rpg-transformed-hud-target.class" | awk '{print $1}')"
printf '%s  more-rpg-transformed-hud-target.class\n' "$SHA" > "$EVIDENCE/more-rpg-transformed-hud-target.sha256"
printf '%s\n' "$REL" > "$EVIDENCE/more-rpg-transformed-hud-target.source-path.txt"
javap -c -p "$TARGET" > "$EVIDENCE/more-rpg-transformed-hud-target.javap.txt" 2>&1 || true
strings -a "$TARGET" | sort -u > "$EVIDENCE/more-rpg-transformed-hud-target.strings.txt"
grep -F 'net/more_rpg_classes/client/heart/HeartRegistry' "$EVIDENCE/more-rpg-transformed-hud-target.strings.txt" >/dev/null
[[ "$(sha256sum "$EVIDENCE/more-rpg-transformed-hud-target.class" | awk '{print $1}')" = "$SHA" ]]
printf '[More RPG 2.7.2] TRANSFORMED_HUD_EVIDENCE_PRESERVED_PASS sha256=%s source=%s targets=%s\n' "$SHA" "$REL" "${#TARGETS[@]}"
