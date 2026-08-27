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
    print('[dependency-audit] unclassified dependency — investigate/download/convert before continuing', file=sys.stderr)
    for item in unknown:
        print(f"  {item['file']}:{item['line']}: {item['coordinate']}", file=sys.stderr)
    raise SystemExit(2)

seen_ids = sorted({item['id'] for item in found})
print(f'[dependency-audit] {project_key}: {len(found)} declarations classified; ids={",".join(seen_ids)}')
for item in found:
    print(f"  {item['id']}: {item['coordinate']} -> {item['status']}")
