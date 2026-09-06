#!/usr/bin/env python3
import re
import sys
from pathlib import Path

if len(sys.argv) != 2:
    raise SystemExit('usage: fix-holder-method-references.py <java-root>')

root = Path(sys.argv[1]).resolve()
if not root.is_dir():
    raise SystemExit(f'missing Java root: {root}')

changed = 0
for path in root.rglob('*.java'):
    text = path.read_text(encoding='utf-8')
    fixed = text
    # The owner bridge previously replaced the prefix of DeferredHolder::getId as if it
    # were DeferredHolder::get, producing the invalid token `holder -> holder.get()Id`.
    fixed = fixed.replace('holder -> holder.get()Id', 'holder -> holder.getId()')
    # Keep these exact method-reference translations idempotent for future source changes.
    fixed = fixed.replace('DeferredHolder::getId', 'holder -> holder.getId()')
    fixed = re.sub(r'DeferredHolder::get\b', 'holder -> holder.get()', fixed)
    if fixed != text:
        path.write_text(fixed, encoding='utf-8')
        changed += 1

# A generated source tree containing the broken token is never acceptable.
remaining = []
for path in root.rglob('*.java'):
    if 'holder -> holder.get()Id' in path.read_text(encoding='utf-8'):
        remaining.append(str(path))
if remaining:
    raise SystemExit('unrepaired holder method-reference syntax: ' + ', '.join(remaining))

print(f'holder method-reference repair complete; files changed: {changed}')
