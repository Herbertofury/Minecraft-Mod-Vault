#!/usr/bin/env python3
import re
import sys
from pathlib import Path

if len(sys.argv) != 2:
    raise SystemExit('usage: apply-synched-entity-data-bridge.py <java-root>')
root = Path(sys.argv[1]).resolve()
if not root.is_dir():
    raise SystemExit(f'missing Java root: {root}')

changed = 0
methods = 0
for path in root.rglob('*.java'):
    text = path.read_text(encoding='utf-8')
    if 'defineSynchedData(SynchedEntityData.Builder builder)' not in text:
        continue
    fixed = text
    fixed, n = re.subn(
        r'protected\s+void\s+defineSynchedData\(SynchedEntityData\.Builder\s+builder\)',
        'protected void defineSynchedData()',
        fixed)
    methods += n
    fixed = fixed.replace('super.defineSynchedData(builder);', 'super.defineSynchedData();')
    fixed = fixed.replace('builder.define(', 'this.getEntityData().define(')
    if fixed != text:
        path.write_text(fixed, encoding='utf-8')
        changed += 1

remaining = []
for path in root.rglob('*.java'):
    text = path.read_text(encoding='utf-8')
    if 'SynchedEntityData.Builder builder' in text or 'builder.define(' in text or 'super.defineSynchedData(builder)' in text:
        remaining.append(str(path))
if remaining:
    raise SystemExit('untranslated 1.21 SynchedEntityData builder API remains: ' + ', '.join(remaining[:20]))

print(f'Forge 1.20 SynchedEntityData owner staged; methods={methods}, files={changed}')
