#!/usr/bin/env python3
from pathlib import Path
import hashlib
import io
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
fixed_time = (1980, 1, 1, 0, 0, 0)


def is_nested_jar(name: str) -> bool:
    return name.lower().endswith('.jar')


def canonicalize_refmap_json(raw: bytes, label: str) -> bytes:
    try:
        payload = json.loads(raw)
    except (UnicodeDecodeError, json.JSONDecodeError) as exc:
        raise SystemExit(f'generated Mixin refmap is not valid UTF-8 JSON: {label}: {exc}')
    if not isinstance(payload, dict):
        raise SystemExit(f'generated Mixin refmap root is not an object: {label}')
    # Loom/Mixin can emit semantically identical refmap objects in different insertion orders across
    # otherwise frozen CI runs. JSON object order has no runtime meaning, so make only the outer
    # generated refmap bytes canonical before release hashing. Nested dependency JAR bytes stay under
    # their existing identity rules and are not rewritten by this normalization.
    return (json.dumps(payload, sort_keys=True, separators=(',', ':'), ensure_ascii=False) + '\n').encode('utf-8')


def canonicalize_zip_bytes(raw: bytes, label: str, depth: int = 0) -> bytes:
    if depth > 8:
        raise SystemExit(f'nested JAR depth exceeded while canonicalizing {label}')
    try:
        with zipfile.ZipFile(io.BytesIO(raw), 'r') as zin:
            infos = zin.infolist()
            names = [info.filename for info in infos]
            if len(set(names)) != len(names):
                raise SystemExit(f'nested JAR contains duplicate ZIP entry names: {label}')
            entries = []
            for info in sorted(infos, key=lambda i: i.filename):
                data = zin.read(info.filename)
                if is_nested_jar(info.filename):
                    data = canonicalize_zip_bytes(data, f'{label}!/{info.filename}', depth + 1)
                entries.append((info.filename, info.compress_type, data))
    except zipfile.BadZipFile as exc:
        raise SystemExit(f'nested JAR is not a valid ZIP: {label}: {exc}')

    out = io.BytesIO()
    with zipfile.ZipFile(out, 'w', allowZip64=True) as zout:
        for name, compress_type, data in entries:
            clone = zipfile.ZipInfo(name, fixed_time)
            clone.compress_type = compress_type
            clone.create_system = 3
            clone.external_attr = ((0o40755 if name.endswith('/') else 0o100644) << 16)
            compression_level = 9 if compress_type == zipfile.ZIP_DEFLATED else None
            zout.writestr(clone, data, compress_type=compress_type, compresslevel=compression_level)
    result = out.getvalue()
    with zipfile.ZipFile(io.BytesIO(result), 'r') as check:
        bad = check.testzip()
        if bad is not None:
            raise SystemExit(f'canonical nested JAR failed ZIP integrity: {label}: {bad}')
        check_names = check.namelist()
        if check_names != sorted(check_names):
            raise SystemExit(f'canonical nested JAR entry order is not sorted: {label}')
        for info in check.infolist():
            if info.date_time != fixed_time or info.extra or info.comment:
                raise SystemExit(f'canonical nested JAR retained ZIP metadata drift: {label}!/{info.filename}')
    return result


with zipfile.ZipFile(outer, 'r') as zin:
    infos = zin.infolist()
    names = {info.filename for info in infos}
    if len(names) != len(infos):
        raise SystemExit('Spell Engine release contains duplicate ZIP entry names; refusing ambiguous canonical seal')
    metadata_path = 'META-INF/jarjar/metadata.json'
    if metadata_path not in names:
        raise SystemExit('Spell Engine release lost Forge JarJar metadata')
    jarjar_metadata = json.loads(zin.read(metadata_path))
    candidates = []
    jarjar_paths = []
    for jar in jarjar_metadata.get('jars', []):
        identifier = jar.get('identifier', {})
        artifact = str(identifier.get('artifact', '')).lower()
        path = jar.get('path')
        if path:
            jarjar_paths.append(path)
        if 'tiny' in artifact and 'config' in artifact:
            candidates.append(path)
    if len(candidates) != 1 or not candidates[0]:
        raise SystemExit(f'expected exactly one TinyConfig JarJar entry, found {candidates}')
    nested_path = candidates[0]
    if nested_path not in names:
        raise SystemExit(f'JarJar metadata TinyConfig path missing from outer JAR: {nested_path}')
    for path in jarjar_paths:
        if path not in names:
            raise SystemExit(f'JarJar metadata path missing from outer JAR: {path}')

    entries = []
    payload_tree = hashlib.sha256()
    manifest_lines = []
    canonicalized_nested = 0
    canonicalized_refmaps = 0
    for info in sorted(infos, key=lambda i: i.filename):
        if info.filename == nested_path:
            data = certified_bytes
        else:
            data = zin.read(info.filename)
            if info.filename in jarjar_paths and is_nested_jar(info.filename):
                data = canonicalize_zip_bytes(data, info.filename)
                canonicalized_nested += 1
            elif info.filename.lower().endswith('refmap.json'):
                data = canonicalize_refmap_json(data, info.filename)
                canonicalized_refmaps += 1
        digest = hashlib.sha256(data).hexdigest()
        entries.append((info.filename, info.compress_type, data))
        manifest_lines.append(f'{digest}  {info.filename}')
        payload_tree.update(info.filename.encode('utf-8'))
        payload_tree.update(b'\0')
        payload_tree.update(bytes.fromhex(digest))
        payload_tree.update(b'\n')
    if canonicalized_refmaps < 1:
        raise SystemExit('Spell Engine release has no generated Mixin refmap to canonicalize')
    payload_tree_sha = payload_tree.hexdigest()

    fd, temp_name = tempfile.mkstemp(prefix=outer.name + '.certified-seal-', suffix='.jar', dir=outer.parent)
    os.close(fd)
    temp = Path(temp_name)
    try:
        with zipfile.ZipFile(temp, 'w', allowZip64=True) as zout:
            for name, compress_type, data in entries:
                clone = zipfile.ZipInfo(name, fixed_time)
                clone.compress_type = compress_type
                clone.create_system = 3
                clone.external_attr = ((0o40755 if name.endswith('/') else 0o100644) << 16)
                compression_level = 9 if compress_type == zipfile.ZIP_DEFLATED else None
                zout.writestr(clone, data, compress_type=compress_type, compresslevel=compression_level)

        with zipfile.ZipFile(temp, 'r') as check:
            if check.testzip() is not None:
                raise SystemExit('sealed Spell Engine JAR failed ZIP integrity')
            if json.loads(check.read(metadata_path)) != jarjar_metadata:
                raise SystemExit('Forge JarJar metadata payload changed while sealing certified TinyConfig')
            nested_sha = hashlib.sha256(check.read(nested_path)).hexdigest()
            if nested_sha != certified_sha:
                raise SystemExit(f'certified nested TinyConfig identity mismatch after seal: {nested_sha} != {certified_sha}')
            check_names = check.namelist()
            if check_names != sorted(check_names):
                raise SystemExit('canonical Spell Engine JAR entry order is not sorted')
            for info in check.infolist():
                if info.date_time != fixed_time:
                    raise SystemExit(f'canonical Spell Engine JAR retained timestamp drift: {info.filename} {info.date_time}')
                if info.extra or info.comment:
                    raise SystemExit(f'canonical Spell Engine JAR retained ZIP metadata drift: {info.filename}')
                if info.filename.lower().endswith('refmap.json'):
                    raw = check.read(info.filename)
                    if raw != canonicalize_refmap_json(raw, info.filename):
                        raise SystemExit(f'canonical Spell Engine JAR retained refmap ordering drift: {info.filename}')
        os.replace(temp, outer)
    finally:
        temp.unlink(missing_ok=True)

manifest_path = Path(str(outer) + '.payload.sha256')
manifest_path.write_text('\n'.join(manifest_lines) + '\n', encoding='utf-8')
outer_sha = hashlib.sha256(outer.read_bytes()).hexdigest()
print(f'[Spell Engine graduation] CERTIFIED_TINY_CONFIG_EXACT_NESTED_SEAL_PASS nested={certified_sha} canonicalized_nested={canonicalized_nested} canonicalized_refmaps={canonicalized_refmaps} payload_tree={payload_tree_sha} outer={outer_sha}')
print(f'[Spell Engine graduation] PAYLOAD_MANIFEST path={manifest_path}')
