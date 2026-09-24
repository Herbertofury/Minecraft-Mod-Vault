#!/usr/bin/env python3
"""Canonicalize the build-ephemeral Architectury injection namespace in Shield API.

The accepted Shield API 2.1.0 Forge 1.20.1 run-188 JAR is semantically and
byte-for-byte authoritative. Fresh Architectury builds can vary only in a
32-hex generated injection namespace. This script is intentionally fail-closed:
it verifies the complete normalized entry inventory/payload manifest, rewrites
only that namespace plus PlatformMethods.class's matching self-name/CRC, and
requires the final bytes to equal the certified run-188 SHA-256.
"""
from __future__ import annotations

import hashlib
from pathlib import Path
import re
import shutil
import struct
import sys
import zipfile
import zlib

CERT_TOKEN = b"fc15e3154a7d45e38b79c2ed9df5e8cf"
CERT_SHA256 = "bd6a2fbeb357c25953abfb14ba18d2c5344e5351c29d2cb082244bc48e8da48a"
INJECT_RE = re.compile(
    r"^architectury_inject_shieldapiforge1201_common_([0-9a-f]{32})_"
    r"19b20811e4e5b7b4a0a52e900207cb93587ff93f50ed5b1dc6d9238e08b92318"
    r"shield_apicommon2101201devjar/$"
)
EMPTY = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
EXPECTED = {
    "META-INF/MANIFEST.MF": "021a7228a47c2da61c0ee6c4f27dfdd12ba96115c4e8ca5675a738873423f829",
    "META-INF/": EMPTY,
    "META-INF/mods.toml": "1319cd7911ccee1e27a3ef79485c3e46f2b603a5967c3e0b88c2c374a68e3bf2",
    "architectury_inject_shieldapiforge1201_common_{TOKEN}_19b20811e4e5b7b4a0a52e900207cb93587ff93f50ed5b1dc6d9238e08b92318shield_apicommon2101201devjar/": EMPTY,
    "architectury_inject_shieldapiforge1201_common_{TOKEN}_19b20811e4e5b7b4a0a52e900207cb93587ff93f50ed5b1dc6d9238e08b92318shield_apicommon2101201devjar/PlatformMethods.class": "5fc7e312908c753d822fe8e4f808262781ef34c1c44d2acf4c5fcd572e795723",
    "icon.png": "605c9d154bcb817cef675a4458a45324ed93d12d386d889a2965f1026969d91f",
    "net/": EMPTY,
    "net/fabric_extras/": EMPTY,
    "net/fabric_extras/shield_api/": EMPTY,
    "net/fabric_extras/shield_api/ShieldAPI.class": "b856b9b9cfb9009a4c7a17fa01f57ed5e8f5546d98d74c24f038852002927d51",
    "net/fabric_extras/shield_api/client/": EMPTY,
    "net/fabric_extras/shield_api/client/ShieldAPIClient.class": "4bbe64a087da7459e71ff26cb1012ae09e6b458aa66337badcf32eb1ffa7e730",
    "net/fabric_extras/shield_api/forge/": EMPTY,
    "net/fabric_extras/shield_api/forge/ShieldAPIForge.class": "6ec5e409b1395837fae35c788164abd2a4cc887bd519063a34560699e49b134e",
    "net/fabric_extras/shield_api/forge/ShieldAPISelfTest$TestServerPlayer.class": "06fc599692a1dd5242fd521fdd4f920530337703f44235ce089575fe7f183d78",
    "net/fabric_extras/shield_api/forge/ShieldAPISelfTest$TrackingCooldownManager.class": "302f8a7888735433227acde8424da6a37a6759dafb86c72f70e7d7e9e217f861",
    "net/fabric_extras/shield_api/forge/ShieldAPISelfTest.class": "ff3f19aa5ed01476331ced9127c9b41508a103e03fccb464b1865908bfecc50d",
    "net/fabric_extras/shield_api/item/": EMPTY,
    "net/fabric_extras/shield_api/item/CustomShieldItem.class": "bf63b74b19b995a0edbda7e49911af443b56736ddddbb62c83c5cd65e30ac22b",
    "net/fabric_extras/shield_api/mixin/": EMPTY,
    "net/fabric_extras/shield_api/mixin/client/": EMPTY,
    "net/fabric_extras/shield_api/mixin/client/MinecraftClientMixin.class": "adc6bba5b2e5733d62e8d2e6744f45dc9a4db0a1e30f3cbe69240b125559eed0",
    "net/fabric_extras/shield_api/mixin/client/ModelPredicateProviderRegistryInvoker.class": "e2324ff6eef0bedcda8cbd3db1df54145a16ea158b6bfed907328beb48288c42",
    "net/fabric_extras/shield_api/mixin/entity/": EMPTY,
    "net/fabric_extras/shield_api/mixin/entity/player/": EMPTY,
    "net/fabric_extras/shield_api/mixin/entity/player/PlayerEntityMixin.class": "e99559e81537c3cd3bc3eddf9097e17769c1c886bc0ed8f6b6a9907f53a5cc24",
    "shield_api-common-common-refmap.json": "0d1fcd410016665b059a49a475cfabbf24587486c4b2c5659ec2b0a39c66786b",
    "shield_api.mixins.json": "4aec87fa910a43fb40685d67350680084386980f2e3ac9986483b8b4384fe5ac",
}


def sha(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def die(msg: str) -> "NoReturn":
    raise SystemExit(f"[Shield certification] {msg}")


def main() -> None:
    if len(sys.argv) != 3:
        die("usage: certify-shield-api-run188.py INPUT.jar OUTPUT.jar")
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

        observed: dict[str, str] = {}
        platform_infos = []
        payload_token_hits = []
        for info in infos:
            data = zf.read(info.filename)
            name = info.filename.replace(token.decode(), "{TOKEN}")
            if token in data:
                payload_token_hits.append(info.filename)
                data = data.replace(token, b"{TOKEN}")
            observed[name] = sha(data)
            if info.filename.endswith("/PlatformMethods.class"):
                platform_infos.append(info)
        if observed != EXPECTED:
            missing = sorted(set(EXPECTED) - set(observed))
            extra = sorted(set(observed) - set(EXPECTED))
            changed = sorted(k for k in set(observed) & set(EXPECTED) if observed[k] != EXPECTED[k])
            die(f"accepted payload manifest mismatch; missing={missing} extra={extra} changed={changed}")
        if len(platform_infos) != 1 or payload_token_hits != [platform_infos[0].filename]:
            die(f"generated token leaked outside PlatformMethods.class: {payload_token_hits}")
        platform = platform_infos[0]
        platform_data = zf.read(platform.filename)

    current_sha = sha(raw)
    if token == CERT_TOKEN:
        if current_sha != CERT_SHA256:
            die(f"certified namespace already present but archive SHA is {current_sha}, expected {CERT_SHA256}")
        dst.parent.mkdir(parents=True, exist_ok=True)
        shutil.copyfile(src, dst)
        print(f"[Shield certification] already exact run-188 bytes: {CERT_SHA256}")
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

    # Replace the raw DEFLATE stream. The certified token may compress to a different length
    # than this run's ephemeral token, so all following ZIP offsets are adjusted deliberately.
    buf[data_start:data_start + platform.compress_size] = compressed
    descriptor = old_descriptor + delta
    if bytes(buf[descriptor:descriptor + 4]) != b"PK\x07\x08":
        die("PlatformMethods descriptor did not shift with compressed stream")
    struct.pack_into("<III", buf, descriptor + 4, crc, len(compressed), len(normalized))

    eocd = buf.rfind(b"PK\x05\x06")
    if eocd < 0:
        die("ZIP end-of-central-directory record missing")
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
            die(f"invalid central-directory signature at offset {pos}")
        name_len, extra_len, record_comment_len = struct.unpack_from("<HHH", buf, pos + 28)
        name = bytes(buf[pos + 46:pos + 46 + name_len])
        local_offset = struct.unpack_from("<I", buf, pos + 42)[0]
        if name == old_name:
            struct.pack_into("<III", buf, pos + 16, crc, len(compressed), len(normalized))
            central_hits += 1
        if local_offset > platform.header_offset:
            struct.pack_into("<I", buf, pos + 42, local_offset + delta)
        pos += 46 + name_len + extra_len + record_comment_len
        record_count += 1
    if pos != central_end or record_count != len(infos) or central_hits != 1:
        die(
            f"central-directory walk mismatch: end={pos == central_end} records={record_count}/{len(infos)} "
            f"platform={central_hits}"
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
        die(f"canonicalized bytes are not certified run-188 JAR: {result_sha} != {CERT_SHA256}")
    print(
        f"[Shield certification] canonicalized generated namespace {token.decode()} -> {CERT_TOKEN.decode()} "
        f"(PlatformMethods deflate delta {delta:+d})"
    )
    print(f"[Shield certification] exact run-188 bytes proven: {CERT_SHA256}")


if __name__ == "__main__":
    main()
