#!/usr/bin/env python3
from pathlib import Path
import hashlib
import json
import os
import sys
import tempfile
import zipfile

if len(sys.argv) != 3:
    raise SystemExit('usage: seal_certified_tinyconfig_nested.py <outer-spell-engine-jar> <certified-tinyconfig-jar>')

outer = Path(sys.argv[1]).resolve()
certified = Path(sys.argv[2]).resolve()
if not outer.is_file() or not certified.is_file():
    raise SystemExit('outer Spell Engine JAR and certified TinyConfig JAR must both exist')

certified_bytes = certified.read_bytes()
certified_sha = hashlib.sha256(certified_bytes).hexdigest()

with zipfile.ZipFile(outer, 'r') as zin:
    infos = zin.infolist()
    names = {info.filename for info in infos}
    metadata_path = 'META-INF/jarjar/metadata.json'
    if metadata_path not in names:
        raise SystemExit('Spell Engine release lost Forge JarJar metadata')
    jarjar_metadata = json.loads(zin.read(metadata_path))
    candidates = []
    for jar in jarjar_metadata.get('jars', []):
        identifier = jar.get('identifier', {})
        artifact = str(identifier.get('artifact', '')).lower()
        if 'tiny' in artifact and 'config' in artifact:
            candidates.append(jar.get('path'))
    if len(candidates) != 1 or not candidates[0]:
        raise SystemExit(f'expected exactly one TinyConfig JarJar entry, found {candidates}')
    nested_path = candidates[0]
    if nested_path not in names:
        raise SystemExit(f'JarJar metadata TinyConfig path missing from outer JAR: {nested_path}')

    fd, temp_name = tempfile.mkstemp(prefix=outer.name + '.certified-seal-', suffix='.jar', dir=outer.parent)
    os.close(fd)
    temp = Path(temp_name)
    try:
        with zipfile.ZipFile(temp, 'w', allowZip64=True) as zout:
            for info in infos:
                data = certified_bytes if info.filename == nested_path else zin.read(info.filename)
                clone = zipfile.ZipInfo(info.filename, info.date_time)
                clone.compress_type = info.compress_type
                clone.comment = info.comment
                clone.extra = info.extra
                clone.internal_attr = info.internal_attr
                clone.external_attr = info.external_attr
                clone.create_system = info.create_system
                clone.create_version = info.create_version
                clone.extract_version = info.extract_version
                clone.flag_bits = info.flag_bits
                clone.volume = info.volume
                compression_level = 9 if info.compress_type == zipfile.ZIP_DEFLATED else None
                zout.writestr(clone, data, compress_type=info.compress_type, compresslevel=compression_level)

        with zipfile.ZipFile(temp, 'r') as check:
            if check.testzip() is not None:
                raise SystemExit('sealed Spell Engine JAR failed ZIP integrity')
            if json.loads(check.read(metadata_path)) != jarjar_metadata:
                raise SystemExit('Forge JarJar metadata changed while sealing certified TinyConfig')
            nested_sha = hashlib.sha256(check.read(nested_path)).hexdigest()
            if nested_sha != certified_sha:
                raise SystemExit(f'certified nested TinyConfig identity mismatch after seal: {nested_sha} != {certified_sha}')
        os.replace(temp, outer)
    finally:
        temp.unlink(missing_ok=True)

outer_sha = hashlib.sha256(outer.read_bytes()).hexdigest()
print(f'[Spell Engine graduation] CERTIFIED_TINY_CONFIG_EXACT_NESTED_SEAL_PASS nested={certified_sha} outer={outer_sha}')
