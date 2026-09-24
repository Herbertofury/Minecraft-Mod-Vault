#!/usr/bin/env python3
from __future__ import annotations

import hashlib
import io
import json
from pathlib import Path
import re
import stat
import sys
import zipfile

CERT_TOKEN = b"3a61cf85e59945b7b3a0e6dad54adb3f"
CERT_SHA256 = "bb085d90f5196b08ef9ddae1f1faa8c5631a88a112450d3768511386e65fa4f3"
CERT_SIZE = 211711
CERT_ENTRY_COUNT = 146
FIXED_TIME = (1980, 1, 1, 0, 0, 0)
INJECT_RE = re.compile(
    r"^architectury_inject_spellpowerforge1201_common_([0-9a-f]{32})_"
    r"460693cbcf13be93e7cb9e72b3955da031f9af9667864a4c652277a5db09fd83"
    r"spell_powercommon1601201devjar/(.*)$"
)
OLD_NESTED_TINY_CONFIG = "META-INF/jars/TinyConfig-2.3.2.jar"


def die(msg: str) -> "NoReturn":
    raise SystemExit(f"[Spell Power current certification] {msg}")


def deterministic_zip(entries: list[tuple[str, bytes, bool]]) -> bytes:
    out = io.BytesIO()
    with zipfile.ZipFile(out, "w", compression=zipfile.ZIP_DEFLATED, compresslevel=9) as zf:
        for name, data, is_dir in sorted(entries, key=lambda e: e[0]):
            info = zipfile.ZipInfo(name, date_time=FIXED_TIME)
            info.create_system = 3
            info.extra = b""
            info.comment = b""
            info.compress_type = zipfile.ZIP_STORED if is_dir else zipfile.ZIP_DEFLATED
            info.external_attr = ((stat.S_IFDIR | 0o755) if is_dir else (stat.S_IFREG | 0o644)) << 16
            if is_dir:
                info.external_attr |= 0x10
            zf.writestr(
                info,
                b"" if is_dir else data,
                compress_type=info.compress_type,
                compresslevel=None if is_dir else 9,
            )
    return out.getvalue()


def validate_contract(zf: zipfile.ZipFile) -> None:
    names = set(zf.namelist())
    if "META-INF/mods.toml" not in names:
        die("META-INF/mods.toml missing")
    if OLD_NESTED_TINY_CONFIG in names:
        die("obsolete embedded TinyConfig 2.3.2 survived current 3.1.0 foundation")
    if any(n.startswith("META-INF/jarjar/") for n in names):
        die("unexpected legacy JarJar metadata survived external TinyConfig 3.1.0 foundation")
    mods = zf.read("META-INF/mods.toml").decode("utf-8", "strict")
    required = (
        'modId="spell_power"',
        'version="1.6.0+1.20.1"',
        'modId="tiny_config"',
        'versionRange="[3.1.0,)"',
        'versionRange="[1.20.1,1.20.2)"',
    )
    for marker in required:
        if marker not in mods:
            die(f"current foundation metadata marker missing: {marker}")
    if 'TinyConfig-2.3.2' in mods:
        die("legacy TinyConfig 2.3.2 metadata survived")


def main() -> None:
    if len(sys.argv) != 3:
        die("usage: certify-spell-power-current.py INPUT.jar OUTPUT.jar")
    src, dst = map(Path, sys.argv[1:])
    if not src.is_file():
        die(f"input JAR missing: {src}")

    with zipfile.ZipFile(src) as zf:
        bad = zf.testzip()
        if bad:
            die(f"input ZIP integrity failure at {bad}")
        infos = zf.infolist()
        if len(infos) != CERT_ENTRY_COUNT:
            die(f"entry count changed: {len(infos)} != {CERT_ENTRY_COUNT}")
        validate_contract(zf)

        inject = [(i.filename, INJECT_RE.match(i.filename)) for i in infos]
        inject = [(n, m) for n, m in inject if m]
        tokens = {m.group(1) for _, m in inject}
        if len(tokens) != 1:
            die(f"expected exactly one Architectury nonce, found {sorted(tokens)}")
        token_text = next(iter(tokens))
        token = token_text.encode()
        expected_inject_names = {
            f"architectury_inject_spellpowerforge1201_common_{token_text}_460693cbcf13be93e7cb9e72b3955da031f9af9667864a4c652277a5db09fd83spell_powercommon1601201devjar/",
            f"architectury_inject_spellpowerforge1201_common_{token_text}_460693cbcf13be93e7cb9e72b3955da031f9af9667864a4c652277a5db09fd83spell_powercommon1601201devjar/PlatformMethods.class",
        }
        actual_inject_names = {n for n, _ in inject}
        if actual_inject_names != expected_inject_names:
            die(f"Architectury injection namespace drifted: {sorted(actual_inject_names)}")

        payload_hits = []
        for info in infos:
            if info.is_dir():
                continue
            data = zf.read(info.filename)
            if token in data:
                payload_hits.append(info.filename)
        platform_path = next(n for n in expected_inject_names if n.endswith("/PlatformMethods.class"))
        if payload_hits != [platform_path]:
            die(f"Architectury nonce leaked outside PlatformMethods.class: {payload_hits}")

        entries: list[tuple[str, bytes, bool]] = []
        for info in infos:
            name = info.filename.replace(token_text, CERT_TOKEN.decode())
            data = zf.read(info.filename)
            if token in data:
                data = data.replace(token, CERT_TOKEN)
            entries.append((name, data, info.is_dir()))

    result = deterministic_zip(entries)
    dst.parent.mkdir(parents=True, exist_ok=True)
    dst.write_bytes(result)

    with zipfile.ZipFile(dst) as zf:
        bad = zf.testzip()
        if bad:
            die(f"canonical ZIP integrity failure at {bad}")
        if len(zf.infolist()) != CERT_ENTRY_COUNT:
            die(f"canonical entry count changed: {len(zf.infolist())} != {CERT_ENTRY_COUNT}")
        validate_contract(zf)
        canonical_names = [n for n in zf.namelist() if INJECT_RE.match(n)]
        if len(canonical_names) != 2 or not all(CERT_TOKEN.decode() in n for n in canonical_names):
            die(f"canonical Architectury namespace missing/drifted: {canonical_names}")

    if len(result) != CERT_SIZE:
        die(f"canonical size drifted: {len(result)} != {CERT_SIZE}")
    result_sha = hashlib.sha256(result).hexdigest()
    if result_sha != CERT_SHA256:
        die(f"canonical bytes drifted: {result_sha} != {CERT_SHA256}")
    print(
        "[Spell Power current certification] CURRENT_TINYCONFIG_310_FOUNDATION_PASS "
        f"architectury_token={token_text}->{CERT_TOKEN.decode()} entries={CERT_ENTRY_COUNT}"
    )
    print(
        "[Spell Power current certification] exact canonical authority bytes proven: "
        f"sha256={CERT_SHA256} size={CERT_SIZE}"
    )


if __name__ == "__main__":
    main()
