#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 2:
    raise SystemExit('usage: patch_spell_engine_1104_generated_runner.py <generated-graduation-runner>')

runner = Path(sys.argv[1]).resolve()
s = runner.read_text()

# Seal the first release candidate immediately after the historical runner copies remapJar output.
# TinyConfig remains byte-identical to the certified Forge JAR. Other nested JarJar payloads are
# recursively canonicalized so Loom's nested ZIP bookkeeping cannot masquerade as product drift.
# Preserve a per-entry payload manifest for exact first-vs-clean-rebuild diagnostics.
old = '''cp "$JAR" "$OUT_JAR"\nsha256sum "$OUT_JAR" | tee "$PORT/spell-engine-forge-1.20.1.sha256"\n'''
new = '''cp "$JAR" "$OUT_JAR"\npython3 "$PORT/tools/seal_certified_tinyconfig_nested.py" "$OUT_JAR" "$TINY_CONFIG_FORGE_JAR"\ncp "$OUT_JAR.payload.sha256" "$PORT/spell-engine-first-payload.sha256"\nsha256sum "$OUT_JAR" | tee "$PORT/spell-engine-forge-1.20.1.sha256"\n'''
if s.count(old) != 1:
    raise SystemExit(f'expected one first-build Spell Engine output seam, found {s.count(old)}')
s = s.replace(old, new, 1)

# Seal the independent clean rebuild by the identical deterministic operation before cross-build
# comparison. Compare full per-entry payload manifests first, printing the exact changed paths on any
# real content drift; only then compare canonical outer JAR bytes.
old = '''JAR2="$(find "$WORK/forge/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*dev-shadow*' ! -name '*sources*' | head -n 1)"\nSECOND_SHA="$(sha256sum "$JAR2" | awk '{print $1}')"\n'''
new = '''JAR2="$(find "$WORK/forge/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*dev-shadow*' ! -name '*sources*' | head -n 1)"\npython3 "$PORT/tools/seal_certified_tinyconfig_nested.py" "$JAR2" "$TINY_CONFIG_FORGE_JAR"\ncp "$JAR2.payload.sha256" "$PORT/spell-engine-second-payload.sha256"\nif ! cmp -s "$PORT/spell-engine-first-payload.sha256" "$PORT/spell-engine-second-payload.sha256"; then\n  echo '[Spell Engine graduation] CLEAN_REBUILD_PAYLOAD_DRIFT' >&2\n  diff -u "$PORT/spell-engine-first-payload.sha256" "$PORT/spell-engine-second-payload.sha256" || true\n  exit 1\nfi\necho '[Spell Engine graduation] CLEAN_REBUILD_PAYLOAD_IDENTITY_PASS'\nSECOND_SHA="$(sha256sum "$JAR2" | awk '{print $1}')"\n'''
if s.count(old) != 1:
    raise SystemExit(f'expected one clean-rebuild Spell Engine output seam, found {s.count(old)}')
s = s.replace(old, new, 1)

# The 1.10.3 tooltip-details key must be proven by the real Forge client lifecycle, not merely source
# inspection. Require the CI marker from ForgeClientMod before the existing client bootstrap can pass.
old = '''  if [[ -f "$LOG" ]] && grep -Fq 'Reloading ResourceManager' "$LOG" && grep -Fq 'Backend library: LWJGL' "$LOG"; then\n'''
new = '''  if [[ -f "$LOG" ]] && grep -Fq 'Reloading ResourceManager' "$LOG" && grep -Fq 'Backend library: LWJGL' "$LOG" && grep -Fq '[Spell Engine CI] TOOLTIP_DETAILS_KEY_REGISTERED' "${FILES[@]}"; then\n'''
if s.count(old) != 1:
    raise SystemExit(f'expected one native-client readiness seam, found {s.count(old)}')
s = s.replace(old, new, 1)

# The generic packaged-server harness tries to rediscover cached runtime mods by parsing mods.toml with
# an over-escaped regex. #315 proved the native client itself resolves the exact Cloth Config and Player
# Animator versions, then the server harness exits before launch at those cache lookups. Resolve the same
# immutable production artifacts from the exact Maven repositories/coordinates already declared by the build,
# validate ZIP integrity and literal mod IDs, and never feed Loom's named/remapped dev copies to a packaged server.
old = '''CLOTH_FORGE_JAR="$(find_runtime_mod 'cloth-config-forge-*.jar' cloth_config)"\nPLAYER_ANIM_FORGE_JAR="$(find_runtime_mod 'player-animation-lib-forge-*.jar' playeranimator)"\ntest -f "$CLOTH_FORGE_JAR" -a -f "$PLAYER_ANIM_FORGE_JAR"\n'''
new = '''SERVER_DEPS="$PORT/.packaged-server-deps"\nrm -rf "$SERVER_DEPS"; mkdir -p "$SERVER_DEPS"\nCLOTH_FORGE_JAR="$SERVER_DEPS/cloth-config-forge-11.1.106.jar"\nPLAYER_ANIM_FORGE_JAR="$SERVER_DEPS/player-animation-lib-forge-1.0.2+1.19.4.jar"\ncurl -fsSL 'https://maven.shedaniel.me/me/shedaniel/cloth/cloth-config-forge/11.1.106/cloth-config-forge-11.1.106.jar' -o "$CLOTH_FORGE_JAR"\ncurl -fsSL 'https://maven.kosmx.dev/dev/kosmx/player-anim/player-animation-lib-forge/1.0.2+1.19.4/player-animation-lib-forge-1.0.2+1.19.4.jar' -o "$PLAYER_ANIM_FORGE_JAR"\nunzip -tq "$CLOTH_FORGE_JAR" >/dev/null\nunzip -tq "$PLAYER_ANIM_FORGE_JAR" >/dev/null\nunzip -p "$CLOTH_FORGE_JAR" META-INF/mods.toml | grep -F 'modId="cloth_config"' >/dev/null\nunzip -p "$PLAYER_ANIM_FORGE_JAR" META-INF/mods.toml | grep -F 'modId="playeranimator"' >/dev/null\necho "[Spell Engine graduation] packaged dependency SHA cloth=$(sha256sum "$CLOTH_FORGE_JAR" | awk '{print $1}') playeranim=$(sha256sum "$PLAYER_ANIM_FORGE_JAR" | awk '{print $1}')"\necho '[Spell Engine graduation] PACKAGED_SERVER_RUNTIME_DEPENDENCIES_RESOLVED_EXACT'\n'''
if s.count(old) != 1:
    raise SystemExit(f'expected one packaged-server runtime dependency lookup seam, found {s.count(old)}')
s = s.replace(old, new, 1)

runner.write_text(s)
final = runner.read_text()
for required in (
    'seal_certified_tinyconfig_nested.py" "$OUT_JAR" "$TINY_CONFIG_FORGE_JAR"',
    'seal_certified_tinyconfig_nested.py" "$JAR2" "$TINY_CONFIG_FORGE_JAR"',
    'spell-engine-first-payload.sha256',
    'spell-engine-second-payload.sha256',
    'CLEAN_REBUILD_PAYLOAD_IDENTITY_PASS',
    'CLEAN_REBUILD_PAYLOAD_DRIFT',
    "[Spell Engine CI] TOOLTIP_DETAILS_KEY_REGISTERED",
    "https://maven.shedaniel.me/me/shedaniel/cloth/cloth-config-forge/11.1.106/cloth-config-forge-11.1.106.jar",
    "https://maven.kosmx.dev/dev/kosmx/player-anim/player-animation-lib-forge/1.0.2+1.19.4/player-animation-lib-forge-1.0.2+1.19.4.jar",
    'PACKAGED_SERVER_RUNTIME_DEPENDENCIES_RESOLVED_EXACT',
):
    if required not in final:
        raise SystemExit(f'generated 1.10.4 graduation runner hardening missing: {required}')
print('[Spell Engine 1.10.4] EXACT_SEAL_NATIVE_TOOLTIP_AND_PACKAGED_SERVER_DEPENDENCY_GATE_PATCHED')
