#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 2:
    raise SystemExit('usage: patch_spell_engine_1104_warning_screen_runner.py <generated-graduation-runner>')

runner = Path(sys.argv[1]).resolve()
s = runner.read_text()
old = '''rm -rf "$WORK/forge/run/logs"\nmkdir -p "$WORK/forge/run/config"\nprintf 'earlyWindowControl = false\\n' > "$WORK/forge/run/config/fml.toml"\nCLIENT_LOG="$PORT/forge-client-smoke.log"\n'''
new = '''rm -rf "$WORK/forge/run/logs"\nmkdir -p "$WORK/forge/run/config"\nprintf 'earlyWindowControl = false\\n' > "$WORK/forge/run/config/fml.toml"\ncat > "$WORK/forge/run/config/forge-client.toml" <<'FORGECLIENT'\n[client]\nshowLoadWarnings = false\nFORGECLIENT\necho '[Spell Engine 1.10.4] FORGE_INTERACTIVE_WARNING_SCREEN_DISABLED_FOR_QA showLoadWarnings=false real_loading_errors_still_fatal=true'\nCLIENT_LOG="$PORT/forge-client-smoke.log"\n'''
if s.count(old) != 1:
    raise SystemExit(f'[Spell Engine 1.10.4] expected one Forge client config seam, found {s.count(old)}')
if 'showLoadWarnings = false' in s:
    raise SystemExit('[Spell Engine 1.10.4] warning-screen QA patch unexpectedly already present')
s = s.replace(old, new, 1)
for required in (
    'forge-client.toml',
    '[client]',
    'showLoadWarnings = false',
    'real_loading_errors_still_fatal=true',
):
    if s.count(required) != 1:
        raise SystemExit(f'[Spell Engine 1.10.4] warning-screen QA contract failed for {required!r}: {s.count(required)}')
runner.write_text(s)
print('[Spell Engine 1.10.4] FORGE_WARNING_SCREEN_QA_RUNNER_PATCHED')
