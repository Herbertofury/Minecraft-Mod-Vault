#!/usr/bin/env python3
from __future__ import annotations

import argparse
import concurrent.futures
import hashlib
import json
import os
import platform
import re
import shlex
import shutil
import urllib.request
import zipfile
from pathlib import Path
from typing import Any, Iterable

MANIFEST_V2 = "https://piston-meta.mojang.com/mc/game/version_manifest_v2.json"
ASSET_BASE = "https://resources.download.minecraft.net"
LIBRARIES_BASE = "https://libraries.minecraft.net"
USER_AGENT = "Minecraft-Mod-Vault-More-RPG-Production-QA/1"


def die(msg: str) -> "NoReturn":
    raise SystemExit(msg)


def sha1_file(path: Path) -> str:
    h = hashlib.sha1()
    with path.open("rb") as f:
        for chunk in iter(lambda: f.read(1024 * 1024), b""):
            h.update(chunk)
    return h.hexdigest()


def sha256_file(path: Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as f:
        for chunk in iter(lambda: f.read(1024 * 1024), b""):
            h.update(chunk)
    return h.hexdigest()


def fetch_bytes(url: str) -> bytes:
    req = urllib.request.Request(url, headers={"User-Agent": USER_AGENT})
    with urllib.request.urlopen(req, timeout=90) as resp:
        return resp.read()


def install_bytes(url: str, path: Path, expected_sha1: str | None = None, expected_size: int | None = None) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    if path.is_file():
        size_ok = expected_size is None or path.stat().st_size == expected_size
        hash_ok = expected_sha1 is None or sha1_file(path) == expected_sha1
        if size_ok and hash_ok:
            return
    data = fetch_bytes(url)
    if expected_size is not None and len(data) != expected_size:
        die(f"size mismatch for {url}: {len(data)} != {expected_size}")
    if expected_sha1 is not None and hashlib.sha1(data).hexdigest() != expected_sha1:
        die(f"SHA1 mismatch for {url}")
    tmp = path.with_suffix(path.suffix + ".part")
    tmp.write_bytes(data)
    os.replace(tmp, path)


def load_json(path: Path) -> dict[str, Any]:
    try:
        data = json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        die(f"invalid JSON {path}: {exc}")
    if not isinstance(data, dict):
        die(f"JSON root must be object: {path}")
    return data


def rule_matches(rule: dict[str, Any], features: dict[str, bool]) -> bool:
    os_rule = rule.get("os") or {}
    if os_rule:
        name = os_rule.get("name")
        if name and name != "linux":
            return False
        arch = os_rule.get("arch")
        machine = platform.machine().lower()
        canonical_arch = "x86_64" if machine in {"x86_64", "amd64"} else machine
        if arch and arch != canonical_arch:
            return False
        version_rx = os_rule.get("version")
        if version_rx:
            try:
                if not re.search(version_rx, platform.release()):
                    return False
            except re.error as exc:
                die(f"invalid OS version rule regex {version_rx!r}: {exc}")
    for key, wanted in (rule.get("features") or {}).items():
        if bool(features.get(key, False)) != bool(wanted):
            return False
    return True


def allowed(entry: dict[str, Any], features: dict[str, bool]) -> bool:
    rules = entry.get("rules")
    if not rules:
        return True
    result = False
    for rule in rules:
        if rule_matches(rule, features):
            result = rule.get("action", "allow") == "allow"
    return result


def flatten_args(raw: Iterable[Any], features: dict[str, bool]) -> list[str]:
    out: list[str] = []
    for item in raw:
        if isinstance(item, str):
            out.append(item)
        elif isinstance(item, dict) and allowed(item, features):
            value = item.get("value")
            if isinstance(value, str):
                out.append(value)
            elif isinstance(value, list) and all(isinstance(v, str) for v in value):
                out.extend(value)
            else:
                die(f"unsupported launcher argument value: {value!r}")
    return out


def maven_path(name: str) -> str:
    parts = name.split(":")
    if len(parts) < 3:
        die(f"bad Maven coordinate: {name}")
    group, artifact, version = parts[:3]
    classifier = parts[3] if len(parts) > 3 else None
    filename = f"{artifact}-{version}" + (f"-{classifier}" if classifier else "") + ".jar"
    return f"{group.replace('.', '/')}/{artifact}/{version}/{filename}"


def library_key(lib: dict[str, Any]) -> tuple[str, str, str | None]:
    parts = str(lib.get("name", "")).split(":")
    if len(parts) < 3:
        die(f"bad launcher library coordinate: {lib.get('name')!r}")
    classifier = parts[3] if len(parts) > 3 else None
    return parts[0], parts[1], classifier


def merge_libraries(parent: list[dict[str, Any]], child: list[dict[str, Any]]) -> list[dict[str, Any]]:
    merged: dict[tuple[str, str, str | None], dict[str, Any]] = {}
    order: list[tuple[str, str, str | None]] = []
    for lib in parent + child:
        key = library_key(lib)
        if key not in merged:
            order.append(key)
        merged[key] = lib
    return [merged[k] for k in order]


def artifact_spec(lib: dict[str, Any]) -> tuple[str, str | None, int | None, str]:
    downloads = lib.get("downloads") or {}
    artifact = downloads.get("artifact")
    if artifact:
        path = artifact.get("path") or maven_path(lib["name"])
        url = artifact.get("url") or (str(lib.get("url") or LIBRARIES_BASE).rstrip("/") + "/" + path)
        return url, artifact.get("sha1"), artifact.get("size"), path
    path = maven_path(lib["name"])
    url = str(lib.get("url") or LIBRARIES_BASE).rstrip("/") + "/" + path
    return url, None, None, path


def native_spec(lib: dict[str, Any]) -> tuple[str, str | None, int | None, str] | None:
    natives = lib.get("natives") or {}
    classifier = natives.get("linux")
    if not classifier:
        return None
    arch = "64" if platform.machine().lower() in {"x86_64", "amd64"} else "32"
    classifier = classifier.replace("${arch}", arch)
    classifiers = (lib.get("downloads") or {}).get("classifiers") or {}
    data = classifiers.get(classifier)
    if not data:
        return None
    path = data.get("path")
    url = data.get("url")
    if not path or not url:
        die(f"native classifier lacks path/url: {lib.get('name')} {classifier}")
    return url, data.get("sha1"), data.get("size"), path


def download_library(lib: dict[str, Any], libraries: Path) -> Path:
    url, sha1, size, rel = artifact_spec(lib)
    path = libraries / rel
    install_bytes(url, path, sha1, size)
    if sha1 and sha1_file(path) != sha1:
        die(f"library SHA1 mismatch after install: {path}")
    return path


def download_native(lib: dict[str, Any], libraries: Path) -> Path | None:
    spec = native_spec(lib)
    if not spec:
        return None
    url, sha1, size, rel = spec
    path = libraries / rel
    install_bytes(url, path, sha1, size)
    if sha1 and sha1_file(path) != sha1:
        die(f"native SHA1 mismatch after install: {path}")
    return path


def extract_native(jar: Path, target: Path, excludes: list[str]) -> None:
    target.mkdir(parents=True, exist_ok=True)
    with zipfile.ZipFile(jar) as zf:
        for info in zf.infolist():
            name = info.filename
            if name.endswith("/") or name.startswith("META-INF/") or any(name.startswith(x) for x in excludes):
                continue
            leaf = Path(name).name
            if not leaf:
                continue
            (target / leaf).write_bytes(zf.read(info))


def ensure_vanilla(mc_home: Path, minecraft: str) -> tuple[Path, dict[str, Any]]:
    manifest_path = mc_home / "version_manifest_v2.json"
    install_bytes(MANIFEST_V2, manifest_path)
    manifest = load_json(manifest_path)
    entry = next((v for v in manifest.get("versions", []) if v.get("id") == minecraft), None)
    if not entry:
        die(f"Minecraft {minecraft} not found in official manifest")
    version_dir = mc_home / "versions" / minecraft
    version_json = version_dir / f"{minecraft}.json"
    install_bytes(entry["url"], version_json, entry.get("sha1"))
    version = load_json(version_json)
    client = version.get("downloads", {}).get("client") or {}
    client_jar = version_dir / f"{minecraft}.jar"
    install_bytes(client["url"], client_jar, client.get("sha1"), client.get("size"))
    if client.get("sha1") and sha1_file(client_jar) != client["sha1"]:
        die("official Minecraft client SHA1 mismatch")
    return version_json, version


def ensure_assets(mc_home: Path, version: dict[str, Any], workers: int) -> tuple[Path, int, int]:
    index = version.get("assetIndex") or {}
    if not index.get("url") or not index.get("id"):
        die("Minecraft version metadata has no asset index")
    assets = mc_home / "assets"
    index_path = assets / "indexes" / f"{index['id']}.json"
    install_bytes(index["url"], index_path, index.get("sha1"), index.get("size"))
    index_json = load_json(index_path)
    objects = index_json.get("objects") or {}
    unique: dict[str, tuple[int | None, list[str]]] = {}
    for logical, data in objects.items():
        digest = data.get("hash")
        if not digest or len(digest) != 40:
            die(f"bad asset hash for {logical}")
        size = data.get("size")
        prev = unique.get(digest)
        if prev and prev[0] != size:
            die(f"same asset hash has conflicting sizes: {digest}")
        unique.setdefault(digest, (size, []))[1].append(logical)

    def ensure_one(item: tuple[str, tuple[int | None, list[str]]]) -> str:
        digest, (size, _names) = item
        path = assets / "objects" / digest[:2] / digest
        install_bytes(f"{ASSET_BASE}/{digest[:2]}/{digest}", path, digest, size)
        if sha1_file(path) != digest:
            die(f"asset SHA1 mismatch: {digest}")
        return digest

    with concurrent.futures.ThreadPoolExecutor(max_workers=max(1, workers)) as pool:
        list(pool.map(ensure_one, sorted(unique.items())))

    manifest = assets / "more-rpg-production-assets.sha1"
    manifest.parent.mkdir(parents=True, exist_ok=True)
    manifest.write_text(
        "".join(f"{digest} {size if size is not None else '-'} {','.join(sorted(names))}\n" for digest, (size, names) in sorted(unique.items())),
        encoding="utf-8",
    )
    return index_path, len(objects), len(unique)


def substitute(args: list[str], values: dict[str, str]) -> list[str]:
    out: list[str] = []
    pattern = re.compile(r"\$\{([^}]+)\}")
    for arg in args:
        def repl(match: re.Match[str]) -> str:
            key = match.group(1)
            if key not in values:
                die(f"unresolved launcher placeholder ${{{key}}} in {arg!r}")
            return values[key]
        out.append(pattern.sub(repl, arg))
    return out


def write_sha256_manifest(paths: Iterable[Path], root: Path, out: Path) -> None:
    rows = []
    seen: set[Path] = set()
    for p in paths:
        p = p.resolve()
        if p in seen:
            continue
        seen.add(p)
        if not p.is_file():
            die(f"manifest path missing: {p}")
        try:
            rel = p.relative_to(root.resolve()).as_posix()
        except ValueError:
            rel = str(p)
        rows.append((rel, sha256_file(p)))
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text("".join(f"{digest}  {rel}\n" for rel, digest in sorted(rows)), encoding="utf-8")


def bootstrap(args: argparse.Namespace) -> None:
    mc_home = Path(args.mc_home).resolve()
    mc_home.mkdir(parents=True, exist_ok=True)
    version_json, version = ensure_vanilla(mc_home, args.minecraft)
    libraries = mc_home / "libraries"
    features: dict[str, bool] = {}
    for lib in version.get("libraries", []):
        if allowed(lib, features):
            download_library(lib, libraries)
            download_native(lib, libraries)
    index_path, logical_count, unique_count = ensure_assets(mc_home, version, args.asset_workers)
    launcher_profiles = mc_home / "launcher_profiles.json"
    if not launcher_profiles.exists():
        launcher_profiles.write_text('{"profiles":{},"settings":{},"version":3}\n', encoding="utf-8")
    write_sha256_manifest(
        [version_json, mc_home / "versions" / args.minecraft / f"{args.minecraft}.jar", index_path],
        mc_home,
        mc_home / "more-rpg-production-bootstrap.sha256",
    )
    print(f"[More RPG 2.7.2] OFFICIAL_MOJANG_CLIENT_BOOTSTRAP_PASS minecraft={args.minecraft} assets_logical={logical_count} assets_unique={unique_count}")


def prepare(args: argparse.Namespace) -> None:
    mc_home = Path(args.mc_home).resolve()
    game_dir = Path(args.game_dir).resolve()
    natives = game_dir / ".production-natives"
    shutil.rmtree(natives, ignore_errors=True)
    natives.mkdir(parents=True, exist_ok=True)
    version_json, parent = ensure_vanilla(mc_home, args.minecraft)
    child_path = mc_home / "versions" / args.forge_version_id / f"{args.forge_version_id}.json"
    if not child_path.is_file():
        die(f"Forge client profile missing after installer: {child_path}")
    child = load_json(child_path)
    if child.get("inheritsFrom") != args.minecraft:
        die(f"Forge profile inheritsFrom mismatch: {child.get('inheritsFrom')!r}")
    main_class = child.get("mainClass")
    if main_class != "cpw.mods.bootstraplauncher.BootstrapLauncher":
        die(f"unexpected Forge production mainClass: {main_class!r}")

    features = {
        "is_demo_user": False,
        "has_custom_resolution": False,
        "has_quick_plays_support": False,
        "is_quick_play_singleplayer": False,
        "is_quick_play_multiplayer": False,
        "is_quick_play_realms": False,
    }
    libraries = mc_home / "libraries"
    merged = merge_libraries(parent.get("libraries", []), child.get("libraries", []))
    classpath: list[Path] = []
    native_jars: list[Path] = []
    for lib in merged:
        if not allowed(lib, features):
            continue
        classpath.append(download_library(lib, libraries))
        n = download_native(lib, libraries)
        if n:
            native_jars.append(n)
            excludes = ((lib.get("extract") or {}).get("exclude") or [])
            extract_native(n, natives, excludes)

    vanilla_client = mc_home / "versions" / args.minecraft / f"{args.minecraft}.jar"
    if not vanilla_client.is_file():
        die("vanilla client JAR missing")
    classpath.append(vanilla_client)

    index_path, logical_count, unique_count = ensure_assets(mc_home, parent, args.asset_workers)
    asset_index = parent["assetIndex"]["id"]
    cp = os.pathsep.join(str(p) for p in classpath)
    values = {
        "auth_player_name": "MRPG-QA",
        "version_name": args.forge_version_id,
        "game_directory": str(game_dir),
        "assets_root": str(mc_home / "assets"),
        "assets_index_name": str(asset_index),
        "auth_uuid": "00000000000000000000000000000001",
        "auth_access_token": "0",
        "clientid": "0",
        "auth_xuid": "0",
        "user_type": "legacy",
        "version_type": "release",
        "natives_directory": str(natives),
        "launcher_name": "minecraft-mod-vault-more-rpg-ci",
        "launcher_version": "1",
        "classpath": cp,
        "classpath_separator": os.pathsep,
        "resolution_width": "1280",
        "resolution_height": "720",
    }

    parent_args = parent.get("arguments") or {}
    child_args = child.get("arguments") or {}
    jvm = flatten_args(parent_args.get("jvm", []), features) + flatten_args(child_args.get("jvm", []), features)
    game = flatten_args(parent_args.get("game", []), features) + flatten_args(child_args.get("game", []), features)
    jvm = substitute(jvm, values)
    game = substitute(game, values)
    game += ["--width", "1280", "--height", "720", "--quickPlaySingleplayer", args.quick_play]

    joined_game = "\n".join(game)
    if "forgeclientuserdev" in joined_game:
        die("Forge profile resolved to forbidden forgeclientuserdev target")
    if "forgeclient" not in game:
        die("Forge production profile did not provide exact forgeclient launch target")
    if "-Djava.library.path=" not in "\n".join(jvm):
        jvm.append(f"-Djava.library.path={natives}")
    jvm.append(f"-Dorg.lwjgl.librarypath={natives}")

    game_dir.mkdir(parents=True, exist_ok=True)
    launch_script = Path(args.launch_script).resolve()
    cmd = [args.java] + jvm + [main_class] + game
    launch_script.write_text("#!/usr/bin/env bash\nset -euo pipefail\nexec " + " ".join(shlex.quote(x) for x in cmd) + "\n", encoding="utf-8")
    launch_script.chmod(0o755)

    runtime_manifest = game_dir / "more-rpg-production-runtime.sha256"
    write_sha256_manifest(classpath + native_jars + [version_json, child_path, index_path], mc_home, runtime_manifest)
    launch_meta = game_dir / "more-rpg-production-launch.json"
    launch_meta.write_text(json.dumps({
        "minecraft": args.minecraft,
        "forge_version_id": args.forge_version_id,
        "main_class": main_class,
        "launch_target": "forgeclient",
        "classpath_entries": len(classpath),
        "native_jars": len(native_jars),
        "assets_logical": logical_count,
        "assets_unique": unique_count,
        "quick_play": args.quick_play,
        "java": args.java,
    }, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    print(f"[More RPG 2.7.2] PRODUCTION_FORGECLIENT_RUNTIME_PREPARED classpath={len(classpath)} native_jars={len(native_jars)} assets_unique={unique_count} target=forgeclient")


def main() -> None:
    parser = argparse.ArgumentParser()
    sub = parser.add_subparsers(dest="command", required=True)
    p_boot = sub.add_parser("bootstrap")
    p_boot.add_argument("--mc-home", required=True)
    p_boot.add_argument("--minecraft", default="1.20.1")
    p_boot.add_argument("--asset-workers", type=int, default=16)
    p_boot.set_defaults(func=bootstrap)

    p_prep = sub.add_parser("prepare")
    p_prep.add_argument("--mc-home", required=True)
    p_prep.add_argument("--game-dir", required=True)
    p_prep.add_argument("--minecraft", default="1.20.1")
    p_prep.add_argument("--forge-version-id", default="1.20.1-forge-47.4.23")
    p_prep.add_argument("--quick-play", default="MRPG-QA")
    p_prep.add_argument("--launch-script", required=True)
    p_prep.add_argument("--java", default="java")
    p_prep.add_argument("--asset-workers", type=int, default=16)
    p_prep.set_defaults(func=prepare)

    args = parser.parse_args()
    args.func(args)


if __name__ == "__main__":
    main()
