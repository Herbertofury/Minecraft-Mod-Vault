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
    # Malum 1.8.2 keeps SpiritLike beside SpiritArcanaType under spirit.type. The
    # first Forge owner shim accidentally imported the later 1.9 package instead.
    fixed = fixed.replace(
        'import com.sammy.malum.core.systems.spirit.SpiritLike;',
        'import com.sammy.malum.core.systems.spirit.type.SpiritLike;'
    )
    # The owner bridge previously replaced the prefix of DeferredHolder::getId as if it
    # were DeferredHolder::get, producing the invalid token `holder -> holder.get()Id`.
    fixed = fixed.replace('holder -> holder.get()Id', 'holder -> holder.getId()')
    # Keep these exact method-reference translations idempotent for future source changes.
    fixed = fixed.replace('DeferredHolder::getId', 'holder -> holder.getId()')
    fixed = re.sub(r'DeferredHolder::get\b', 'holder -> holder.get()', fixed)
    if fixed != text:
        path.write_text(fixed, encoding='utf-8')
        changed += 1

# A generated source tree containing either known owner regression is unacceptable.
remaining = []
for path in root.rglob('*.java'):
    generated = path.read_text(encoding='utf-8')
    if 'holder -> holder.get()Id' in generated:
        remaining.append(f'{path}: broken holder method reference')
    if path.name == 'SpiritHolder.java' and 'core.systems.spirit.SpiritLike' in generated:
        remaining.append(f'{path}: wrong 1.9 SpiritLike package')
if remaining:
    raise SystemExit('unrepaired Forge owner bridge regression: ' + ', '.join(remaining))

print(f'holder owner repair complete; files changed: {changed}')
