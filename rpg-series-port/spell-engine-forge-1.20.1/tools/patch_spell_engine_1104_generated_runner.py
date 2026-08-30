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
):
    if required not in final:
        raise SystemExit(f'generated 1.10.4 graduation runner hardening missing: {required}')
print('[Spell Engine 1.10.4] EXACT_NESTED_TINYCONFIG_RECURSIVE_SEAL_PAYLOAD_DIFF_AND_TOOLTIP_RUNTIME_GATE_PATCHED')
