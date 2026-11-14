#!/usr/bin/env python3
from __future__ import annotations
import re, stat, sys, tempfile, zipfile
from pathlib import Path

if len(sys.argv) != 2:
    raise SystemExit('usage: canonicalize_release.py <jar>')
jar = Path(sys.argv[1])
fixed = 'c4a9e4f7b7a14a1c9f1f8f0b3d8a2e61'
pat = re.compile(r'^(architectury_inject_roguesforge1201_common_)([0-9a-f]{32})(_[0-9a-f]{64}roguescommon3111201devjar)(/.*)?$')
with zipfile.ZipFile(jar, 'r') as zin:
    rows=[]; tokens=set()
    for info in zin.infolist():
        name=info.filename; data=zin.read(name)
        m=pat.match(name)
        if m:
            tokens.add(m.group(2)); name=f'{m.group(1)}{fixed}{m.group(3)}{m.group(4) or ""}'
        rows.append((name,data,info.is_dir()))
if len(tokens) != 1:
    raise SystemExit(f'[Rogues canonicalize] expected exactly one Architectury inject token, found {sorted(tokens)}')
old=next(iter(tokens)).encode('ascii'); new=fixed.encode('ascii')
rows=[(name, data.replace(old,new), is_dir) for name,data,is_dir in rows]
with tempfile.NamedTemporaryFile(dir=jar.parent, suffix='.jar', delete=False) as tf:
    tmp=Path(tf.name)
try:
    with zipfile.ZipFile(tmp,'w',compression=zipfile.ZIP_DEFLATED,compresslevel=9) as zout:
        for name,data,is_dir in sorted(rows,key=lambda r:r[0]):
            zi=zipfile.ZipInfo(name,date_time=(1980,1,1,0,0,0)); zi.create_system=3; zi.compress_type=zipfile.ZIP_STORED if is_dir else zipfile.ZIP_DEFLATED
            zi.external_attr=((stat.S_IFDIR|0o755) if is_dir else (stat.S_IFREG|0o644))<<16
            if is_dir: zi.external_attr |= 0x10
            zout.writestr(zi,b'' if is_dir else data,compress_type=zi.compress_type,compresslevel=9 if not is_dir else None)
    tmp.replace(jar)
finally:
    if tmp.exists(): tmp.unlink()
print(f'[Rogues canonicalize] Architectury inject namespace {next(iter(tokens))} -> {fixed}; deterministic archive metadata applied')
