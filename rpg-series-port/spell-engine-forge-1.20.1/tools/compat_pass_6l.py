#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 3:
    raise SystemExit('usage: compat_pass_6l.py <generated-port-root> <spell-engine-1.20.1-baseline>')

root = Path(sys.argv[1]).resolve()
source_roots = [root / 'common/src/main/java', root / 'forge/src/main/java']
legacy = 'net.tinyconfig'
modern = 'net.tiny_config'
changed = 0
remaining = []

# TinyConfig 3.1.0 is now the proven native Forge 1.20.1 foundation. Earlier compatibility passes
# deliberately retained the pre-3.x Java package for internal ConfigManager callsites while the
# foundation was incomplete. Migrate those callsites at the final transform boundary so all prior
# behavioral transforms remain stable and the generated port has one TinyConfig API era.
for source_root in source_roots:
    if not source_root.is_dir():
        raise SystemExit(f'Spell Engine source root missing: {source_root}')
    for path in sorted(source_root.rglob('*.java')):
        text = path.read_text()
        if legacy in text:
            updated = text.replace(legacy, modern)
            path.write_text(updated)
            changed += text.count(legacy)

for source_root in source_roots:
    for path in sorted(source_root.rglob('*.java')):
        if legacy in path.read_text():
            remaining.append(str(path.relative_to(root)))

if remaining:
    raise SystemExit('Legacy TinyConfig package survived migration: ' + ', '.join(remaining))
if changed == 0:
    raise SystemExit('TinyConfig runtime migration changed zero callsites; expected legacy internal ConfigManager references')

# Guard the concrete runtime class that caused downstream Forge construction to fail in Archers #175.
config_manager_users = []
for source_root in source_roots:
    for path in sorted(source_root.rglob('*.java')):
        text = path.read_text()
        if 'ConfigManager' in text:
            config_manager_users.append(text)
if not config_manager_users or not any('net.tiny_config.ConfigManager' in text for text in config_manager_users):
    raise SystemExit('No migrated TinyConfig 3.1.0 ConfigManager callsite found after pass 6l')

print(f'Spell Engine compatibility pass 6l applied: migrated {changed} legacy TinyConfig package references to TinyConfig 3.1.0')
