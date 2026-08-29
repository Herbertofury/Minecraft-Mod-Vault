#!/usr/bin/env python3
from __future__ import annotations

import hashlib
from pathlib import Path
import re
import shutil
import struct
import sys
import zipfile
import zlib

CERT_TOKEN = b"b4d726d82d294224a02d6350c5479c4d"
CERT_SHA256 = "95e8f9e074dd1c432f9486c922173b76a3b3bc57667b28fe154e47dba5c374ee"
CERT_NORMALIZED_MANIFEST_SHA256 = "050b78f2a762351b366ad85aca99a91b11b9827bd2ce65a6337d55a42a4c4e06"
CERT_ENTRY_COUNT = 879
INJECT_RE = re.compile(
    r"^architectury_inject_paladinsforge1201_common_([0-9a-f]{32})_"
    r"3dfa31edb3884291eabd70b1e4a4e5e673282aa921564420c750866ec12610c6"
    r"paladinscommon3111201devjar/$"
)


def sha(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def die(msg: str) -> "NoReturn":
    raise SystemExit(f"[Paladins certification] {msg}")


def main() -> None:
    if len(sys.argv) != 3:
        die("usage: certify-paladins-run199.py INPUT.jar OUTPUT.jar")
    src, dst = map(Path, sys.argv[1:])
    raw = src.read_bytes()

    with zipfile.ZipFile(src) as zf:
        infos = zf.infolist()
        inject_dirs = [i.filename for i in infos if INJECT_RE.match(i.filename)]
        if len(inject_dirs) != 1:
            die(f"expected one Architectury injection directory, found {len(inject_dirs)}")
        match = INJECT_RE.match(inject_dirs[0])
        assert match is not None
        token = match.group(1).encode()
        if len(token) != len(CERT_TOKEN):
            die("generated token length changed")
        if len(infos) != CERT_ENTRY_COUNT:
            die(f"entry count changed: {len(infos)} != {CERT_ENTRY_COUNT}")

        records = []
        platform_infos = []
        payload_token_hits = []
        for info in infos:
            data = zf.read(info.filename)
            name = info.filename.replace(token.decode(), "{TOKEN}")
            if token in data:
                payload_token_hits.append(info.filename)
                data = data.replace(token, b"{TOKEN}")
            records.append(f"{name}\0{sha(data)}\n".encode())
            if info.filename.endswith("/PlatformMethods.class"):
                platform_infos.append(info)

        manifest_sha = sha(b"".join(sorted(records)))
        if manifest_sha != CERT_NORMALIZED_MANIFEST_SHA256:
            die(
                "normalized payload manifest changed: "
                f"{manifest_sha} != {CERT_NORMALIZED_MANIFEST_SHA256}"
            )
        if len(platform_infos) != 1 or payload_token_hits != [platform_infos[0].filename]:
            die(f"generated token leaked outside PlatformMethods.class: {payload_token_hits}")
        platform = platform_infos[0]
        platform_data = zf.read(platform.filename)

    current_sha = sha(raw)
    if token == CERT_TOKEN:
        if current_sha != CERT_SHA256:
            die(
                f"certified namespace present but archive SHA is {current_sha}, "
                f"expected {CERT_SHA256}"
            )
        dst.parent.mkdir(parents=True, exist_ok=True)
        shutil.copyfile(src, dst)
        print(f"[Paladins certification] already exact run-199 bytes: {CERT_SHA256}")
        return

    normalized = platform_data.replace(token, CERT_TOKEN)
    if normalized == platform_data or token in normalized:
        die("failed to normalize PlatformMethods self-name")
    compressor = zlib.compressobj(6, zlib.DEFLATED, -15)
    compressed = compressor.compress(normalized) + compressor.flush()
    crc = zlib.crc32(normalized) & 0xFFFFFFFF
    delta = len(compressed) - platform.compress_size

    buf = bytearray(raw)
    off = platform.header_offset
    sig, _ver, flag, method, _tm, _dt, lcrc, lcs, lus, name_len, extra_len = struct.unpack_from(
        "<IHHHHHIIIHH", buf, off
    )
    if sig != 0x04034B50 or method != zipfile.ZIP_DEFLATED or not (flag & 8) or any((lcrc, lcs, lus)):
        die("unexpected PlatformMethods local ZIP header layout")
    data_start = off + 30 + name_len + extra_len
    old_descriptor = data_start + platform.compress_size
    if bytes(buf[old_descriptor:old_descriptor + 4]) != b"PK\x07\x08":
        die("PlatformMethods data descriptor signature changed")
    _sig, old_crc, old_comp_size, old_file_size = struct.unpack_from("<IIII", buf, old_descriptor)
    if old_crc != platform.CRC or old_comp_size != platform.compress_size or old_file_size != platform.file_size:
        die("PlatformMethods data descriptor metadata changed")

    buf[data_start:data_start + platform.compress_size] = compressed
    descriptor = old_descriptor + delta
    if bytes(buf[descriptor:descriptor + 4]) != b"PK\x07\x08":
        die("PlatformMethods descriptor did not shift")
    struct.pack_into("<III", buf, descriptor + 4, crc, len(compressed), len(normalized))

    eocd = buf.rfind(b"PK\x05\x06")
    if eocd < 0:
        die("ZIP EOCD missing")
    disk, cd_disk, entries_disk, entries_total, cd_size, old_cd_offset, comment_len = struct.unpack_from(
        "<HHHHIIH", buf, eocd + 4
    )
    if disk or cd_disk or entries_disk != entries_total or entries_total != len(infos) or comment_len:
        die("unexpected multi-disk/commented ZIP layout")
    central_start = old_cd_offset + delta
    if bytes(buf[central_start:central_start + 4]) != b"PK\x01\x02":
        die("central directory did not shift as expected")

    old_name = platform.filename.encode()
    central_hits = 0
    record_count = 0
    pos = central_start
    central_end = central_start + cd_size
    while pos < central_end:
        if bytes(buf[pos:pos + 4]) != b"PK\x01\x02":
            die(f"invalid central-directory signature at {pos}")
        name_len, extra_len, comment_len = struct.unpack_from("<HHH", buf, pos + 28)
        name = bytes(buf[pos + 46:pos + 46 + name_len])
        local_offset = struct.unpack_from("<I", buf, pos + 42)[0]
        if name == old_name:
            struct.pack_into("<III", buf, pos + 16, crc, len(compressed), len(normalized))
            central_hits += 1
        if local_offset > platform.header_offset:
            struct.pack_into("<I", buf, pos + 42, local_offset + delta)
        pos += 46 + name_len + extra_len + comment_len
        record_count += 1
    if pos != central_end or record_count != len(infos) or central_hits != 1:
        die(
            f"central-directory walk mismatch end={pos == central_end} "
            f"records={record_count}/{len(infos)} platform={central_hits}"
        )
    struct.pack_into("<I", buf, eocd + 16, central_start)

    literal_hits = bytes(buf).count(token)
    if literal_hits != 4:
        die(f"expected four generated-token path records after payload rewrite, found {literal_hits}")
    buf = bytearray(bytes(buf).replace(token, CERT_TOKEN))

    dst.parent.mkdir(parents=True, exist_ok=True)
    dst.write_bytes(buf)
    with zipfile.ZipFile(dst) as zf:
        bad = zf.testzip()
        if bad:
            die(f"certified ZIP integrity failure at {bad}")
    result_sha = sha(bytes(buf))
    if result_sha != CERT_SHA256:
        die(f"canonicalized bytes are not certified run-199 JAR: {result_sha} != {CERT_SHA256}")
    print(
        f"[Paladins certification] canonicalized generated namespace {token.decode()} -> {CERT_TOKEN.decode()} "
        f"(PlatformMethods deflate delta {delta:+d})"
    )
    print(f"[Paladins certification] exact run-199 bytes proven: {CERT_SHA256}")


if __name__ == "__main__":
    main()
