#!/usr/bin/env bash
set -euo pipefail

ROOT="${GITHUB_WORKSPACE:?}"
BASE="$ROOT/rpg-series-port/ci/run-gazebos-acceptance.sh"
PATCHED="$ROOT/rpg-series-port/ci/.run-gazebos-acceptance-v2.generated.sh"
TINY_ENV="$ROOT/rpg-series-port/tiny-config-forge-1.20.1/TINY_CONFIG_GRADUATION.env"

test -f "$BASE"
test -f "$TINY_ENV"

python3 - "$BASE" "$PATCHED" <<'PY'
from pathlib import Path
import sys
src=Path(sys.argv[1]).read_text()
out=Path(sys.argv[2])

old='''echo '[Gazebos] Reconstruct graduated TinyConfig foundation'\nTINY_GEN="$PORT/.tiny-config-generated"\npython3 "$ROOT/rpg-series-port/tiny-config-forge-1.20.1/tools/prepare_port.py" "$TINY_UP" "$TINY_GEN"\ngradle --no-daemon --stacktrace -p "$TINY_GEN" clean :forge:remapJar\nTINY_JAR="$(pick_release_jar "$TINY_GEN/forge/build/libs")"\n[[ -f "$TINY_JAR" ]]\n[[ "$(sha256sum "$TINY_JAR" | awk '{print $1}')" = '0182a492d6c59d7d5f491a39bb2f6634ba5dd38083295305c4769fdb6539db18' ]]\n'''
new='''echo '[Gazebos] Reconstruct graduated TinyConfig foundation at canonical build path'\nsource "$ROOT/rpg-series-port/tiny-config-forge-1.20.1/TINY_CONFIG_GRADUATION.env"\nTINY_GEN="$ROOT/rpg-series-port/tiny-config-forge-1.20.1/generated"\npython3 "$ROOT/rpg-series-port/tiny-config-forge-1.20.1/tools/prepare_port.py" "$TINY_UP" "$TINY_GEN"\ngradle --no-daemon --stacktrace -p "$TINY_GEN" clean :forge:remapJar\nTINY_JAR="$(pick_release_jar "$TINY_GEN/forge/build/libs")"\n[[ -f "$TINY_JAR" ]]\nTINY_ACTUAL_SHA="$(sha256sum "$TINY_JAR" | awk '{print $1}')"\necho "[Gazebos] TinyConfig foundation SHA=$TINY_ACTUAL_SHA expected=$TINY_CONFIG_EXPECTED_JAR_SHA"\nif [[ "$TINY_ACTUAL_SHA" != "$TINY_CONFIG_EXPECTED_JAR_SHA" ]]; then\n  echo '[Gazebos] certified TinyConfig identity mismatch' >&2\n  exit 1\nfi\necho '[Gazebos] CERTIFIED_TINY_CONFIG_IDENTITY_PASS'\n'''
if src.count(old) != 1:
    raise SystemExit(f'[Gazebos v2] expected exactly one TinyConfig reconstruction seam, found {src.count(old)}')
src=src.replace(old,new)

marker='''# Reconstruct pristine product for client and packaged-server lanes.\n'''
insert=r'''# Complementary QA-only provider-presence lane. The sealed release JAR above is untouched.
# Add a real second JavaFML mod identity plus @Mod entrypoint only to the generated QA tree,
# start a fresh JVM/static state, and prove Gazebo suppresses direct StructurePoolAPI injection.
echo '[Gazebos] Lithostitched-present suppression semantic gate'
LITHO_STUB="$GEN/forge/src/main/java/net/gazebo/forge/LithostitchedCiStub.java"
cat > "$LITHO_STUB" <<'JAVA'
package net.gazebo.forge;

import net.minecraftforge.fml.common.Mod;

@Mod("lithostitched")
public final class LithostitchedCiStub {
    public LithostitchedCiStub() {
        System.out.println("GAZEBO_LITHOSTITCHED_STUB_DISCOVERED");
    }
}
JAVA
MODS_TOML="$GEN/forge/src/main/resources/META-INF/mods.toml"
cat >> "$MODS_TOML" <<'TOML'

[[mods]]
modId="lithostitched"
version="1.0.0-ci"
displayName="Lithostitched CI Discovery Stub"
description="QA-only real JavaFML mod identity used to validate Gazebo provider suppression."
TOML

gradle --no-daemon --stacktrace -p "$GEN" :forge:classes
LITHO_SMOKE="$PORT/gazebos-lithostitched-semantics.log"; : > "$LITHO_SMOKE"
rm -rf "$GEN/forge/run"; mkdir -p "$GEN/forge/run"; printf 'eula=true\n' > "$GEN/forge/run/eula.txt"
env GAZEBO_SELF_TEST=1 GAZEBO_EXPECT_PENDING=0 gradle --no-daemon -p "$GEN" :forge:runServer > "$LITHO_SMOKE" 2>&1 &
ACTIVE_PID=$!; DEADLINE=$((SECONDS+150)); PASS=0
while ((SECONDS<DEADLINE)); do
  LOG=$(find "$GEN/forge/run" -type f -path '*/logs/latest.log' | head -n1 || true); FILES=("$LITHO_SMOKE"); [[ -n "$LOG" ]] && FILES+=("$LOG")
  if grep -Eiq "$FATAL" "${FILES[@]}"; then stop_tree "$ACTIVE_PID"; ACTIVE_PID=""; cat "${FILES[@]}"; exit 1; fi
  if grep -Fq 'GAZEBO_LITHOSTITCHED_STUB_DISCOVERED' "${FILES[@]}" && grep -Fq 'GAZEBO_INJECTION_SEMANTICS_PASS pending=0' "${FILES[@]}" && [[ -n "$LOG" ]] && grep -Eq 'Done \([0-9.]+s\)!' "$LOG"; then PASS=1; break; fi
  if ! kill -0 "$ACTIVE_PID" 2>/dev/null; then wait "$ACTIVE_PID" || true; ACTIVE_PID=""; cat "${FILES[@]}"; exit 1; fi
  sleep 1
done
[[ "$PASS" -eq 1 ]]; stop_tree "$ACTIVE_PID"; ACTIVE_PID=""
echo '[Gazebos] LITHOSTITCHED_SUPPRESSION_SEMANTICS_PASS pending=0'

'''
if src.count(marker) != 1:
    raise SystemExit(f'[Gazebos v2] expected exactly one pristine-product marker, found {src.count(marker)}')
src=src.replace(marker,insert+marker)
out.write_text(src)
PY

chmod +x "$PATCHED"
bash -n "$PATCHED"
echo '[Gazebos v2] CANONICAL_TINYCONFIG_AND_LITHOSTITCHED_HARDENING_PASS'
exec bash "$PATCHED"
