#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 2:
    raise SystemExit('usage: patch_spell_engine_1104_nonce_runner.py <generated-graduation-runner>')
runner = Path(sys.argv[1]).resolve()
s = runner.read_text()

first = 'python3 "$PORT/tools/seal_certified_tinyconfig_nested.py" "$OUT_JAR" "$TINY_CONFIG_FORGE_JAR"'
first_new = 'python3 "$PORT/tools/canonicalize_architectury_inject_nonce.py" "$OUT_JAR"\n' + first
if s.count(first) != 1:
    raise SystemExit(f'expected one first-build exact-seal call, found {s.count(first)}')
s = s.replace(first, first_new, 1)

second = 'python3 "$PORT/tools/seal_certified_tinyconfig_nested.py" "$JAR2" "$TINY_CONFIG_FORGE_JAR"'
second_new = 'python3 "$PORT/tools/canonicalize_architectury_inject_nonce.py" "$JAR2"\n' + second
if s.count(second) != 1:
    raise SystemExit(f'expected one clean-rebuild exact-seal call, found {s.count(second)}')
s = s.replace(second, second_new, 1)

runner.write_text(s)
final = runner.read_text()
for required in (
    'canonicalize_architectury_inject_nonce.py" "$OUT_JAR"',
    'canonicalize_architectury_inject_nonce.py" "$JAR2"',
    'seal_certified_tinyconfig_nested.py" "$OUT_JAR"',
    'seal_certified_tinyconfig_nested.py" "$JAR2"',
):
    if required not in final:
        raise SystemExit(f'nonce hardening missing generated-runner requirement: {required}')
print('[Spell Engine 1.10.4] ARCHITECTURY_INJECT_NONCE_GENERATED_RUNNER_PATCHED')
