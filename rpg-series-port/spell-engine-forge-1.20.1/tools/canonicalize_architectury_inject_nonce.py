#!/usr/bin/env python3
from pathlib import Path
import hashlib
import os
import re
import sys
import tempfile
import zipfile

if len(sys.argv) != 2:
    raise SystemExit('usage: canonicalize_architectury_inject_nonce.py <spell-engine-jar>')
jar = Path(sys.argv[1]).resolve()
if not jar.is_file():
    raise SystemExit(f'missing Spell Engine JAR: {jar}')

FIXED = '00000000000000000000000000000000'
PREFIX = 'architectury_inject_spellengineforge1201_common_'
PAT = re.compile(
    r'^(architectury_inject_spellengineforge1201_common_)'
    r'([0-9a-f]{32})'
    r'(_[0-9a-f]+aspell_enginecommon11041201devjar)'
    r'(/PlatformMethods\.class)?/?$'
)
FIXED_TIME = (1980, 1, 1, 0, 0, 0)

with zipfile.ZipFile(jar, 'r') as zin:
    infos = zin.infolist()
    if len({i.filename for i in infos}) != len(infos):
        raise SystemExit('duplicate ZIP entries before Architectury nonce canonicalization')
    entries = []
    matched = []
    class_hits = 0
    for info in infos:
        name = info.filename
        data = zin.read(name)
        m = PAT.match(name)
        if m:
            old = m.group(2)
            matched.append((name, old))
            name = name.replace(old, FIXED, 1)
            if info.filename.endswith('/PlatformMethods.class'):
                old_bytes = old.encode('ascii')
                if len(old_bytes) != len(FIXED):
                    raise SystemExit('Architectury nonce length changed unexpectedly')
                count = data.count(old_bytes)
                if count not in (0, 1):
                    raise SystemExit(f'unexpected Architectury nonce count in PlatformMethods.class: {count}')
                if count == 1:
                    data = data.replace(old_bytes, FIXED.encode('ascii'), 1)
                elif old != FIXED:
                    raise SystemExit('PlatformMethods.class path nonce not present in class constant pool')
                class_hits += 1
        entries.append((name, info.compress_type, data))

# Exactly one directory and one PlatformMethods.class are expected. Re-running after canonicalization is allowed.
if len(matched) != 2 or class_hits != 1:
    raise SystemExit(f'expected exactly Architectury injection directory + PlatformMethods.class, got entries={len(matched)} classes={class_hits}: {matched}')
if len({name for name, _, _ in entries}) != len(entries):
    raise SystemExit('Architectury nonce canonicalization would create duplicate ZIP entries')

fd, temp_name = tempfile.mkstemp(prefix=jar.name + '.architectury-nonce-', suffix='.jar', dir=jar.parent)
os.close(fd)
temp = Path(temp_name)
try:
    with zipfile.ZipFile(temp, 'w', allowZip64=True) as zout:
        for name, compress_type, data in sorted(entries, key=lambda e: e[0]):
            zi = zipfile.ZipInfo(name, FIXED_TIME)
            zi.compress_type = compress_type
            zi.create_system = 3
            zi.external_attr = ((0o40755 if name.endswith('/') else 0o100644) << 16)
            level = 9 if compress_type == zipfile.ZIP_DEFLATED else None
            zout.writestr(zi, data, compress_type=compress_type, compresslevel=level)
    with zipfile.ZipFile(temp, 'r') as check:
        bad = check.testzip()
        if bad is not None:
            raise SystemExit(f'canonicalized Spell Engine JAR failed ZIP integrity: {bad}')
        names = check.namelist()
        if names != sorted(names):
            raise SystemExit('canonicalized Spell Engine JAR entry order is not sorted')
        pm = [n for n in names if n.startswith(PREFIX) and n.endswith('/PlatformMethods.class')]
        if len(pm) != 1 or FIXED not in pm[0]:
            raise SystemExit(f'canonical PlatformMethods path missing: {pm}')
        class_bytes = check.read(pm[0])
        if class_bytes.count(FIXED.encode('ascii')) != 1:
            raise SystemExit('canonical PlatformMethods internal class name not fixed exactly once')
        residual = [n for n in names if n.startswith(PREFIX) and FIXED not in n]
        if residual:
            raise SystemExit(f'residual noncanonical Architectury injection paths: {residual}')
    os.replace(temp, jar)
finally:
    temp.unlink(missing_ok=True)

print(f'[Spell Engine graduation] ARCHITECTURY_INJECT_NONCE_CANONICAL_PASS sha={hashlib.sha256(jar.read_bytes()).hexdigest()}')
