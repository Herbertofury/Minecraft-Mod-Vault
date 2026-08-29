#!/usr/bin/env python3
import json
from pathlib import Path
import re
import sys

if len(sys.argv) != 4:
    raise SystemExit('usage: audit-dependency-compat.py <ledger.json> <project-key> <upstream-project-root>')

ledger_path = Path(sys.argv[1]).resolve()
project_key = sys.argv[2]
source_root = Path(sys.argv[3]).resolve()
ledger = json.loads(ledger_path.read_text())
project = ledger.get('projects', {}).get(project_key)
if not project:
    raise SystemExit(f'dependency audit has no project ledger entry: {project_key}')

entries = project.get('dependencies', [])
known = []
for entry in entries:
    for token in entry.get('upstream_tokens', []):
        known.append((token, entry))

# Groovy dependency calls in the current Architectury platform build files.
pattern = re.compile(r'^\s*(?:modImplementation|modApi|modCompileOnly|modRuntimeOnly|implementation|compileOnly|runtimeOnly|neoForge)\s*\(?\s*["\']([^"\']+)["\']')
found = []
unknown = []
for rel in project.get('source_files', []):
    path = source_root / rel
    if not path.is_file():
        raise SystemExit(f'dependency audit source file missing: {path}')
    for line_no, raw in enumerate(path.read_text().splitlines(), 1):
        line = raw.strip()
        if not line or line.startswith('//') or line.startswith('#'):
            continue
        match = pattern.match(raw)
        if not match:
            continue
        coordinate = match.group(1)
        matched = next((entry for token, entry in known if token in coordinate), None)
        record = {'file': rel, 'line': line_no, 'coordinate': coordinate}
        if matched is None:
            unknown.append(record)
        else:
            record['id'] = matched['id']
            record['classification'] = matched['classification']
            record['status'] = matched['status']
            found.append(record)

if unknown:
    print('[dependency-audit] unclassified dependency - investigate/download/convert before continuing', file=sys.stderr)
    for item in unknown:
        print(f"  {item['file']}:{item['line']}: {item['coordinate']}", file=sys.stderr)
    raise SystemExit(2)

seen_ids = sorted({item['id'] for item in found})
print(f'[dependency-audit] {project_key}: {len(found)} declarations classified; ids={",".join(seen_ids)}')
for item in found:
    print(f"  {item['id']}: {item['coordinate']} -> {item['status']}")

# Soft/runtime integrations are not Gradle dependencies, but they can still hide a
# compatibility obligation. Scan exact current source for Platform.util().isModLoaded(...)
# calls and require every discovered mod id to be classified in the same ledger.
optional_entries = {entry['id']: entry for entry in project.get('optional_integrations', [])}
optional_found = []
optional_unknown = []
constant_pattern = re.compile(r'\b(?:private\s+)?static\s+final\s+String\s+([A-Z0-9_]+)\s*=\s*"([a-z0-9_.-]+)"\s*;')
call_pattern = re.compile(r'\bisModLoaded\(\s*(?:"([a-z0-9_.-]+)"|([A-Z0-9_]+))\s*\)')

for rel in project.get('optional_source_files', []):
    path = source_root / rel
    if not path.is_file():
        raise SystemExit(f'dependency audit optional source file missing: {path}')
    text = path.read_text()
    constants = {name: value for name, value in constant_pattern.findall(text)}
    for match in call_pattern.finditer(text):
        mod_id = match.group(1) or constants.get(match.group(2))
        line_no = text.count('\n', 0, match.start()) + 1
        if not mod_id:
            optional_unknown.append({'file': rel, 'line': line_no, 'expression': match.group(0)})
            continue
        entry = optional_entries.get(mod_id)
        record = {'file': rel, 'line': line_no, 'id': mod_id}
        if entry is None:
            optional_unknown.append(record)
        else:
            record['status'] = entry['status']
            record['user_download_required'] = bool(entry.get('user_download_required', False))
            optional_found.append(record)

if optional_unknown:
    print('[dependency-audit] unclassified optional runtime integration - investigate/download/convert before continuing', file=sys.stderr)
    for item in optional_unknown:
        detail = item.get('id') or item.get('expression')
        print(f"  {item['file']}:{item['line']}: {detail}", file=sys.stderr)
    raise SystemExit(3)

optional_ids = sorted({item['id'] for item in optional_found})
print(f'[dependency-audit] {project_key}: {len(optional_found)} optional runtime checks classified; ids={",".join(optional_ids)}')
for mod_id in optional_ids:
    entry = optional_entries[mod_id]
    print(f"  optional {mod_id}: {entry['status']}; user_download_required={str(bool(entry.get('user_download_required', False))).lower()}")
