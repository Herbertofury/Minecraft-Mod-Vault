#!/usr/bin/env python3
"""Build the compact, embedded Minecraft Mod Vault Version Atlas.

Inputs are immutable snapshots of primary provider metadata. The generated atlas
contains only the fields needed by the offline Porting Lab and records source
hashes so every recommendation can be traced back to its evidence snapshot.
"""
from __future__ import annotations

import argparse
import gzip
import hashlib
import json
import re
import xml.etree.ElementTree as ET
from collections import defaultdict
from datetime import datetime, timezone
from pathlib import Path
from typing import Any

PRERELEASE = re.compile(r"(?i)(?:alpha|beta|snapshot|nightly|dev|rc|pre|experimental|unstable)")


def sha256(path: Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as handle:
        for chunk in iter(lambda: handle.read(1024 * 1024), b""):
            h.update(chunk)
    return h.hexdigest()


def read_json(path: Path) -> Any:
    with path.open("r", encoding="utf-8") as handle:
        return json.load(handle)


def xml_versions(path: Path) -> tuple[list[str], str | None, str | None]:
    root = ET.parse(path).getroot()
    versions = [node.text.strip() for node in root.findall("./versioning/versions/version") if node.text]
    latest_node = root.find("./versioning/latest")
    release_node = root.find("./versioning/release")
    latest = latest_node.text.strip() if latest_node is not None and latest_node.text else (versions[-1] if versions else None)
    release = release_node.text.strip() if release_node is not None and release_node.text else None
    return versions, latest, release


def latest_stable(versions: list[str]) -> str | None:
    for version in reversed(versions):
        if not PRERELEASE.search(version):
            return version
    return versions[-1] if versions else None


def forge_target(version: str) -> str | None:
    if "-" not in version:
        return None
    return version.rsplit("-", 1)[0]


def neoforge_target(version: str) -> str | None:
    core = version.split("-", 1)[0]
    parts = core.split(".")
    if len(parts) < 2 or not all(p.isdigit() for p in parts[:2]):
        return None
    major = int(parts[0])
    if major >= 26:
        return ".".join(parts[:3]) if len(parts) >= 3 else ".".join(parts[:2])
    if major >= 20:
        patch = parts[1]
        return f"1.{major}.{patch}" if patch != "0" else f"1.{major}"
    return None


def newest_by_target(versions: list[str], mapper) -> dict[str, str]:
    out: dict[str, str] = {}
    for version in versions:
        target = mapper(version)
        if target:
            out[target] = version
    return dict(sorted(out.items(), reverse=True))


def compact_version(record: dict[str, Any], mcmeta: dict[str, Any], fabric_games: set[str], quilt_games: set[str], modrinth_games: set[str]) -> dict[str, Any]:
    manifest = record.get("manifest") or {}
    downloads = manifest.get("downloads") or {}
    java = manifest.get("javaVersion") or {}
    mc = mcmeta.get(record["id"]) or {}
    asset = manifest.get("assetIndex") or {}
    return {
        "id": record["id"],
        "type": manifest.get("type", "unknown"),
        "releaseTime": manifest.get("releaseTime", ""),
        "time": manifest.get("time", ""),
        "javaMajor": java.get("majorVersion"),
        "javaComponent": java.get("component", ""),
        "client": "client" in downloads,
        "server": "server" in downloads,
        "clientMappings": "client_mappings" in downloads,
        "serverMappings": "server_mappings" in downloads,
        "assetIndex": asset.get("id", ""),
        "libraryCount": len(manifest.get("libraries") or []),
        "complianceLevel": manifest.get("complianceLevel"),
        "dataVersion": mc.get("data_version"),
        "protocolVersion": mc.get("protocol_version"),
        "dataPackVersion": mc.get("data_pack_version"),
        "resourcePackVersion": mc.get("resource_pack_version"),
        "releaseTarget": mc.get("release_target"),
        "stable": bool(mc.get("stable", False)),
        "fabric": record["id"] in fabric_games,
        "quilt": record["id"] in quilt_games,
        "modrinth": record["id"] in modrinth_games,
    }


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--seed-root", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()

    seed = args.seed_root
    mcmeta_path = seed / "mmv-v090-mcmeta-versions-seed-artifact" / "mcmeta-versions.json"
    deep_path = seed / "mmv-v090-mojang-deep-seed-artifact" / "mojang-version-details-v2.jsonl.gz"
    official = seed / "mmv-v090-official-source-seeds-artifact"
    toolchain = seed / "mmv-v090-toolchain-seeds-artifact"

    required = [mcmeta_path, deep_path, official / "fabric-game-versions.json", official / "quilt-game-versions.json", official / "modrinth-game-versions.json"]
    missing = [str(path) for path in required if not path.exists()]
    if missing:
        raise SystemExit("missing atlas inputs: " + ", ".join(missing))

    mcmeta_records = read_json(mcmeta_path)
    mcmeta = {row["id"]: row for row in mcmeta_records}
    fabric_game_rows = read_json(official / "fabric-game-versions.json")
    quilt_game_rows = read_json(official / "quilt-game-versions.json")
    modrinth_game_rows = read_json(official / "modrinth-game-versions.json")
    fabric_games = {row["version"] for row in fabric_game_rows}
    quilt_games = {row["version"] for row in quilt_game_rows}
    modrinth_games = {row["version"] for row in modrinth_game_rows}

    versions: list[dict[str, Any]] = []
    with gzip.open(deep_path, "rt", encoding="utf-8") as handle:
        for line in handle:
            line = line.strip()
            if not line:
                continue
            versions.append(compact_version(json.loads(line), mcmeta, fabric_games, quilt_games, modrinth_games))

    fabric_loaders = read_json(official / "fabric-loader-versions.json")
    quilt_loaders = read_json(official / "quilt-loader-versions.json")
    modrinth_loaders = read_json(official / "modrinth-loaders.json")

    metadata_files = {
        "fabricLoom": official / "fabric-loom-maven-metadata.xml",
        "forge": official / "forge-maven-metadata.xml",
        "forgeGradle": official / "forgegradle-maven-metadata.xml",
        "modDevGradle": official / "moddevgradle-plugin-maven-metadata.xml",
        "neoForge": official / "neoforge-maven-metadata.xml",
        "neoForm": official / "neoform-maven-metadata.xml",
        "neoGradleUserdev": official / "neogradle-userdev-maven-metadata.xml",
        "architecturyLoom": toolchain / "architectury-loom-maven-metadata.xml",
        "autoRenamingTool": toolchain / "auto-renaming-tool-maven-metadata.xml",
        "fabricApi": toolchain / "fabric-api-maven-metadata.xml",
        "intermediary": toolchain / "intermediary-maven-metadata.xml",
        "mcpConfig": toolchain / "mcpconfig-maven-metadata.xml",
        "mixin": toolchain / "mixin-maven-metadata.xml",
        "mixinExtras": toolchain / "mixinextras-maven-metadata.xml",
        "quiltMappings": toolchain / "quilt-mappings-maven-metadata.xml",
        "srgUtils": toolchain / "srgutils-maven-metadata.xml",
        "yarn": toolchain / "yarn-maven-metadata.xml",
    }
    parsed_meta: dict[str, dict[str, Any]] = {}
    raw_versions: dict[str, list[str]] = {}
    for key, path in metadata_files.items():
        rows, latest, release = xml_versions(path)
        raw_versions[key] = rows
        parsed_meta[key] = {
            "latest": latest,
            "release": release,
            "latestStable": latest_stable(rows),
            "count": len(rows),
            "recent": rows[-12:],
            "source": path.name,
            "sha256": sha256(path),
        }

    forge_versions = raw_versions["forge"]
    neoforge_versions = raw_versions["neoForge"]
    by_type: dict[str, int] = defaultdict(int)
    java_counts: dict[str, int] = defaultdict(int)
    for row in versions:
        by_type[row["type"]] += 1
        if row["javaMajor"] is not None:
            java_counts[str(row["javaMajor"])] += 1

    latest_manifest = read_json(official / "mojang-version-manifest-v2.json").get("latest", {})
    source_paths = [mcmeta_path, deep_path] + [p for p in official.iterdir() if p.is_file()] + [p for p in toolchain.iterdir() if p.is_file()]
    atlas = {
        "schemaVersion": 1,
        "generatedAt": datetime.now(timezone.utc).replace(microsecond=0).isoformat().replace("+00:00", "Z"),
        "evidenceSnapshotAt": "2026-08-20T00:00:00Z",
        "summary": {
            "minecraftVersions": len(versions),
            "latestRelease": latest_manifest.get("release", ""),
            "latestSnapshot": latest_manifest.get("snapshot", ""),
            "oldest": versions[-1]["id"] if versions else "",
            "newest": versions[0]["id"] if versions else "",
            "types": dict(sorted(by_type.items())),
            "javaMajors": dict(sorted(java_counts.items(), key=lambda item: int(item[0]))),
            "clientMappings": sum(1 for row in versions if row["clientMappings"]),
            "serverMappings": sum(1 for row in versions if row["serverMappings"]),
            "mcmetaCoverage": len(mcmeta_records),
            "fabricGameVersions": len(fabric_games),
            "quiltGameVersions": len(quilt_games),
            "modrinthGameVersions": len(modrinth_games),
        },
        "versions": versions,
        "loaders": {
            "fabric": {
                "latest": fabric_loaders[0]["version"] if fabric_loaders else "",
                "latestStable": next((row["version"] for row in fabric_loaders if row.get("stable")), fabric_loaders[0]["version"] if fabric_loaders else ""),
                "recent": [row["version"] for row in fabric_loaders[:16]],
                "gameVersionCount": len(fabric_games),
            },
            "quilt": {
                "latest": quilt_loaders[0]["version"] if quilt_loaders else "",
                "latestStable": next((row["version"] for row in quilt_loaders if not PRERELEASE.search(row["version"])), quilt_loaders[0]["version"] if quilt_loaders else ""),
                "recent": [row["version"] for row in quilt_loaders[:16]],
                "gameVersionCount": len(quilt_games),
            },
            "modrinth": {
                "ids": sorted(row.get("name", "") for row in modrinth_loaders if row.get("name")),
                "gameVersionCount": len(modrinth_games),
            },
        },
        "toolchains": parsed_meta,
        "forge": {
            "latest": parsed_meta["forge"]["latest"],
            "latestStable": parsed_meta["forge"]["latestStable"],
            "latestByMinecraft": newest_by_target(forge_versions, forge_target),
            "count": len(forge_versions),
        },
        "neoForge": {
            "latest": parsed_meta["neoForge"]["latest"],
            "latestStable": parsed_meta["neoForge"]["latestStable"],
            "latestByMinecraft": newest_by_target(neoforge_versions, neoforge_target),
            "count": len(neoforge_versions),
        },
        "sourceEvidence": [
            {"name": path.name, "bytes": path.stat().st_size, "sha256": sha256(path)}
            for path in sorted(source_paths, key=lambda p: p.name)
        ],
    }

    args.output.parent.mkdir(parents=True, exist_ok=True)
    with args.output.open("w", encoding="utf-8", newline="\n") as handle:
        json.dump(atlas, handle, ensure_ascii=False, separators=(",", ":"), sort_keys=True)
        handle.write("\n")
    print(json.dumps({"output": str(args.output), "bytes": args.output.stat().st_size, "sha256": sha256(args.output), "versions": len(versions)}, indent=2))


if __name__ == "__main__":
    main()
