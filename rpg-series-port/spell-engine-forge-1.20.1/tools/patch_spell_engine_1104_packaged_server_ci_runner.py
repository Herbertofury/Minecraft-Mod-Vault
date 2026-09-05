#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 2:
    raise SystemExit('usage: patch_spell_engine_1104_packaged_server_ci_runner.py <generated-graduation-runner>')

runner = Path(sys.argv[1]).resolve()
s = runner.read_text()
old = '''PACKAGE_LOG="$PORT/forge-packaged-server-smoke.log"; : > "$PACKAGE_LOG"\n( cd "$FRESH" && exec ./run.sh nogui ) > "$PACKAGE_LOG" 2>&1 & ACTIVE_PID=$!\n'''
new = '''FORGE_MOD_SOURCE="$WORK/forge/src/main/java/net/spell_engine/forge/ForgeMod.java"\ntest -f "$FORGE_MOD_SOURCE"\n[[ "$(grep -Fc 'System.getenv(\"CI\")' "$FORGE_MOD_SOURCE")" -eq 1 ]]\n[[ "$(grep -Fc 'SpellEngineCiSelfTest::run' "$FORGE_MOD_SOURCE")" -eq 1 ]]\necho '[Spell Engine graduation] PACKAGED_SERVER_CI_SELFTEST_HOOK_PASS callback=SpellEngineCiSelfTest::run'\nPACKAGE_LOG="$PORT/forge-packaged-server-smoke.log"; : > "$PACKAGE_LOG"\necho '[Spell Engine graduation] PACKAGED_SERVER_CI_ENV_EXPLICIT CI=true'\n( cd "$FRESH" && exec env CI=true ./run.sh nogui ) > "$PACKAGE_LOG" 2>&1 & ACTIVE_PID=$!\n'''
if s.count(old) != 1:
    raise SystemExit(f'[Spell Engine 1.10.4] expected one packaged-server launch seam, found {s.count(old)}')
if 'PACKAGED_SERVER_CI_ENV_EXPLICIT' in s or 'exec env CI=true ./run.sh nogui' in s:
    raise SystemExit('[Spell Engine 1.10.4] packaged-server CI env patch unexpectedly already present')
s = s.replace(old, new, 1)
for required in (
    "grep -Fc 'System.getenv(\"CI\")'",
    "grep -Fc 'SpellEngineCiSelfTest::run'",
    'PACKAGED_SERVER_CI_SELFTEST_HOOK_PASS',
    'PACKAGED_SERVER_CI_ENV_EXPLICIT CI=true',
    'exec env CI=true ./run.sh nogui',
):
    if s.count(required) != 1:
        raise SystemExit(f'[Spell Engine 1.10.4] packaged-server CI contract failed for {required!r}: {s.count(required)}')
runner.write_text(s)
print('[Spell Engine 1.10.4] PACKAGED_SERVER_EXPLICIT_CI_ENV_RUNNER_PATCHED')
