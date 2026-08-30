#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 2:
    raise SystemExit('usage: patch_spell_engine_1104_generated_runner.py <generated-graduation-runner>')

runner = Path(sys.argv[1]).resolve()
s = runner.read_text()

# Seal the first release candidate immediately after the historical runner copies remapJar output.
# This removes only Loom's injected nested bookkeeping by replacing the nested TinyConfig entry with
# the exact already-certified Forge JAR while preserving the outer Forge JarJar metadata/path.
old = '''cp "$JAR" "$OUT_JAR"\nsha256sum "$OUT_JAR" | tee "$PORT/spell-engine-forge-1.20.1.sha256"\n'''
new = '''cp "$JAR" "$OUT_JAR"\npython3 "$PORT/tools/seal_certified_tinyconfig_nested.py" "$OUT_JAR" "$TINY_CONFIG_FORGE_JAR"\nsha256sum "$OUT_JAR" | tee "$PORT/spell-engine-forge-1.20.1.sha256"\n'''
if s.count(old) != 1:
    raise SystemExit(f'expected one first-build Spell Engine output seam, found {s.count(old)}')
s = s.replace(old, new, 1)

# Seal the independent clean rebuild by the identical deterministic operation before cross-build
# comparison, so reproducibility proves the final distributable bytes rather than pre-seal Loom bytes.
old = '''JAR2="$(find "$WORK/forge/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*dev-shadow*' ! -name '*sources*' | head -n 1)"\nSECOND_SHA="$(sha256sum "$JAR2" | awk '{print $1}')"\n'''
new = '''JAR2="$(find "$WORK/forge/build/libs" -maxdepth 1 -type f -name '*.jar' ! -name '*dev-shadow*' ! -name '*sources*' | head -n 1)"\npython3 "$PORT/tools/seal_certified_tinyconfig_nested.py" "$JAR2" "$TINY_CONFIG_FORGE_JAR"\nSECOND_SHA="$(sha256sum "$JAR2" | awk '{print $1}')"\n'''
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
    "[Spell Engine CI] TOOLTIP_DETAILS_KEY_REGISTERED",
):
    if required not in final:
        raise SystemExit(f'generated 1.10.4 graduation runner hardening missing: {required}')
print('[Spell Engine 1.10.4] EXACT_NESTED_TINYCONFIG_AND_TOOLTIP_RUNTIME_GATE_PATCHED')
