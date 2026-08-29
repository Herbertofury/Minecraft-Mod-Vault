#!/usr/bin/env python3
"""Build the Minecraft Dev Kit Vanilla Feature Atlas.

Authority model:
  1. Mojang version manifest/version JSON/asset indexes for official lineage + hashes.
  2. misode/mcmeta for generated registries, sounds.json, data/resource JSON trees (1.14+).
  3. PrismarineJS/minecraft-data for normalized historical sound/registry-adjacent coverage.
  4. Backport-mod inventories are compatibility-provider evidence only, never vanilla authority.

The output is metadata-complete and bytes-on-demand. It intentionally does not duplicate every
client/server JAR or every OGG across every version; exact official hashes and asset-index sidecars
are retained so bytes can be hydrated and verified when a backport plan actually needs them.
"""
from __future__ import annotations

import argparse
import concurrent.futures as cf
import datetime as dt
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import sys
import time
import urllib.request
from collections import defaultdict

SCHEMA = 2
UA = "Minecraft-Dev-Kit-Vanilla-Feature-Atlas/2.6.0"
MOJANG_MANIFEST = "https://piston-meta.mojang.com/mc/game/version_manifest_v2.json"


def run(cmd, cwd=None, capture=True):
    r = subprocess.run(cmd, cwd=cwd, check=True, text=True,
                       stdout=subprocess.PIPE if capture else None,
                       stderr=subprocess.PIPE if capture else None)
    return r.stdout if capture else ""


def get(url: str, timeout=90, tries=5) -> bytes:
    last = None
    for i in range(tries):
        try:
            req = urllib.request.Request(url, headers={"User-Agent": UA})
            with urllib.request.urlopen(req, timeout=timeout) as r:
                return r.read()
        except Exception as e:
            last = e
            time.sleep(1.5 * (i + 1))
    raise RuntimeError(f"failed GET {url}: {last}")


def sha1(data: bytes) -> str:
    return hashlib.sha1(data).hexdigest()


def sha256_file(path: Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as f:
        for b in iter(lambda: f.read(1024 * 1024), b""):
            h.update(b)
    return h.hexdigest()


def clone_repo(url: str, path: Path, branches=None, depth=None):
    if path.exists():
        shutil.rmtree(path)
    cmd = ["git", "clone", "--filter=blob:none", "--no-checkout"]
    if depth:
        cmd += ["--depth", str(depth)]
    cmd += [url, str(path)]
    run(cmd, capture=False)
    if branches:
        for b in branches:
            run(["git", "fetch", "--filter=blob:none", "origin", f"{b}:refs/remotes/origin/{b}"], cwd=path, capture=False)


def git_show(repo: Path, commit: str, path: str) -> bytes | None:
    p = subprocess.run(["git", "show", f"{commit}:{path}"], cwd=repo,
                       stdout=subprocess.PIPE, stderr=subprocess.DEVNULL)
    return p.stdout if p.returncode == 0 else None


def branch_version_commits(repo: Path, branch: str) -> list[tuple[str, str]]:
    commits = run(["git", "rev-list", "--reverse", f"origin/{branch}"], cwd=repo).splitlines()
    out = []
    by_version = {}
    for c in commits:
        raw = git_show(repo, c, "version.txt")
        if not raw:
            continue
        v = raw.decode("utf-8", "replace").strip()
        if not v:
            continue
        by_version[v] = c
    for c in commits:
        raw = git_show(repo, c, "version.txt")
        if not raw:
            continue
        v = raw.decode("utf-8", "replace").strip()
        if v and by_version.get(v) == c:
            out.append((v, c))
    return out


def ns(x: str) -> str:
    x = str(x).strip()
    return x if ":" in x else "minecraft:" + x


def registry_entries(value):
    if isinstance(value, list):
        for x in value:
            if isinstance(x, str):
                yield x
            elif isinstance(x, dict):
                for key in ("id", "name", "key"):
                    if isinstance(x.get(key), str):
                        yield x[key]
                        break
        return
    if not isinstance(value, dict):
        return
    if "entries" in value:
        yield from registry_entries(value["entries"])
        return
    for k, v in value.items():
        if k in {"default", "protocol_id", "default_id", "intrusive", "type"}:
            continue
        if isinstance(k, str) and (":" in k or re.match(r"^[a-z0-9_.\-/]+$", k)):
            if isinstance(v, (dict, int, str, type(None), bool)):
                yield k


def sound_refs(defn) -> list[str]:
    if not isinstance(defn, dict):
        return []
    vals = defn.get("sounds", [])
    out = []
    if not isinstance(vals, list):
        return out
    for x in vals:
        name = x if isinstance(x, str) else x.get("name") if isinstance(x, dict) else None
        if not isinstance(name, str):
            continue
        namespace = "minecraft"
        path = name
        if ":" in name:
            namespace, path = name.split(":", 1)
        out.append(f"{namespace}/sounds/{path}.ogg")
    return out


class Presence:
    def __init__(self):
        self.present = defaultdict(lambda: defaultdict(set))
        self.sources = defaultdict(lambda: defaultdict(set))
        self.evidence = defaultdict(lambda: defaultdict(set))

    def add(self, kind, ident, version, source, evidence=None):
        ident = ident if kind in {"external_asset", "resource_path", "data_path"} else ns(ident)
        self.present[kind][ident].add(version)
        self.sources[kind][ident].add(source)
        if evidence and len(self.evidence[kind][ident]) < 10:
            self.evidence[kind][ident].add(evidence)


def compress_intervals(present_versions: set[str], order: list[str]):
    indexes = [i for i, v in enumerate(order) if v in present_versions]
    if not indexes:
        return []
    intervals = []
    start = prev = indexes[0]
    for i in indexes[1:]:
        if i == prev + 1:
            prev = i
            continue
        intervals.append({"from": order[start], "to": order[prev]})
        start = prev = i
    intervals.append({"from": order[start], "to": order[prev]})
    return intervals


def load_json_bytes(raw):
    return json.loads(raw.decode("utf-8"))


def build(args):
    out = Path(args.out).resolve()
    work = Path(args.work).resolve()
    out.mkdir(parents=True, exist_ok=True)
    work.mkdir(parents=True, exist_ok=True)
    asset_dir = out / "official-asset-indexes"
    asset_dir.mkdir(exist_ok=True)

    manifest_raw = get(MOJANG_MANIFEST)
    manifest = load_json_bytes(manifest_raw)
    (out / "version_manifest_v2.json").write_bytes(manifest_raw)
    official_entries = manifest["versions"]
    official_by_id = {v["id"]: v for v in official_entries}

    def fetch_version(entry):
        raw = get(entry["url"])
        if entry.get("sha1") and sha1(raw) != entry["sha1"]:
            raise RuntimeError(f"version JSON hash mismatch {entry['id']}")
        return entry["id"], raw, load_json_bytes(raw)

    version_json = {}
    with cf.ThreadPoolExecutor(max_workers=args.workers) as ex:
        futs = [ex.submit(fetch_version, e) for e in official_entries]
        for i, fut in enumerate(cf.as_completed(futs), 1):
            vid, raw, obj = fut.result()
            version_json[vid] = obj
            if i % 100 == 0 or i == len(futs):
                print(f"official version JSONs {i}/{len(futs)}", flush=True)

    asset_meta = {}
    for vid, obj in version_json.items():
        a = obj.get("assetIndex")
        if a and a.get("sha1") and a.get("url"):
            asset_meta[a["sha1"]] = a

    def fetch_asset_index(item):
        h, meta = item
        raw = get(meta["url"], timeout=120)
        if sha1(raw) != h:
            raise RuntimeError(f"asset index hash mismatch {h}")
        return h, raw, load_json_bytes(raw), meta

    asset_indexes = {}
    with cf.ThreadPoolExecutor(max_workers=args.workers) as ex:
        futs = [ex.submit(fetch_asset_index, x) for x in asset_meta.items()]
        for i, fut in enumerate(cf.as_completed(futs), 1):
            h, raw, obj, meta = fut.result()
            (asset_dir / f"{h}.json").write_bytes(raw)
            asset_indexes[h] = obj
            if i % 25 == 0 or i == len(futs):
                print(f"official asset indexes {i}/{len(futs)}", flush=True)

    mcmeta = work / "mcmeta"
    clone_repo("https://github.com/misode/mcmeta.git", mcmeta, branches=["summary", "assets-json", "data-json"])
    minecraft_data = work / "minecraft-data"
    clone_repo("https://github.com/PrismarineJS/minecraft-data.git", minecraft_data, depth=1)
    vanilla_backport = work / "VanillaBackport"
    if args.with_provider_inventory:
        clone_repo("https://github.com/ItsBlackGear/VanillaBackport.git", vanilla_backport, depth=1)
        run(["git", "fetch", "origin", "1.20.1:refs/remotes/origin/1.20.1"], cwd=vanilla_backport, capture=False)

    revisions = {
        "misode-mcmeta-summary": run(["git", "rev-parse", "origin/summary"], cwd=mcmeta).strip(),
        "misode-mcmeta-assets-json": run(["git", "rev-parse", "origin/assets-json"], cwd=mcmeta).strip(),
        "misode-mcmeta-data-json": run(["git", "rev-parse", "origin/data-json"], cwd=mcmeta).strip(),
        "prismarine-minecraft-data": run(["git", "rev-parse", "HEAD"], cwd=minecraft_data).strip(),
    }
    if args.with_provider_inventory:
        revisions["vanillabackport-1.20.1"] = run(["git", "rev-parse", "origin/1.20.1"], cwd=vanilla_backport).strip()

    official_sorted = sorted(official_entries, key=lambda x: x.get("releaseTime") or x.get("time") or "")
    official_ids = [x["id"] for x in official_sorted]
    official_ids_set = set(official_ids)
    prism_versions_file = minecraft_data / "data/pc/common/versions.json"
    prism_ids = json.loads(prism_versions_file.read_text()) if prism_versions_file.exists() else []
    prism_only = [x for x in prism_ids if x not in official_ids_set]
    order = prism_only + official_ids
    seen = set()
    order = [x for x in order if not (x in seen or seen.add(x))]
    order_set = set(order)

    presence = Presence()
    for vid in official_ids:
        obj = version_json[vid]
        am = obj.get("assetIndex")
        if not am or am.get("sha1") not in asset_indexes:
            continue
        idx = asset_indexes[am["sha1"]]
        for logical in idx.get("objects", {}):
            presence.add("external_asset", logical, vid, "mojang-assets", f"asset-index:{am['sha1']}")

    summary_commits = branch_version_commits(mcmeta, "summary")
    for n, (vid, commit) in enumerate(summary_commits, 1):
        if vid not in order_set:
            continue
        raw_sounds = git_show(mcmeta, commit, "sounds/data.json")
        if raw_sounds:
            try:
                sounds = load_json_bytes(raw_sounds)
                if isinstance(sounds, dict):
                    for sid, defn in sounds.items():
                        ev = f"mcmeta:{vid}:sounds/data.json"
                        refs = sound_refs(defn)
                        if refs:
                            ev += ":refs=" + ",".join(refs[:5]) + ("..." if len(refs) > 5 else "")
                        presence.add("sound_definition", sid, vid, "misode-mcmeta", ev)
            except Exception as e:
                print(f"WARN sounds parse {vid}: {e}", file=sys.stderr)
        raw_regs = git_show(mcmeta, commit, "registries/data.json")
        if raw_regs:
            try:
                regs = load_json_bytes(raw_regs)
                if isinstance(regs, dict):
                    for reg, value in regs.items():
                        regid = ns(reg)
                        entries = list(registry_entries(value))
                        for ent in entries:
                            eid = ns(ent)
                            presence.add("registry_entry", f"{regid}|{eid}", vid, "misode-mcmeta", f"mcmeta:{vid}:registry={regid}")
                            if regid.endswith(":sound_event") or regid == "minecraft:sound_event":
                                presence.add("sound_event", eid, vid, "misode-mcmeta", f"mcmeta:{vid}:registry={regid}")
            except Exception as e:
                print(f"WARN registries parse {vid}: {e}", file=sys.stderr)
        if n % 50 == 0:
            print(f"mcmeta summary {n}/{len(summary_commits)}", flush=True)

    for branch, kind in [("assets-json", "resource_path"), ("data-json", "data_path")]:
        commits = branch_version_commits(mcmeta, branch)
        for n, (vid, commit) in enumerate(commits, 1):
            if vid not in order_set:
                continue
            paths = run(["git", "ls-tree", "-r", "--name-only", commit], cwd=mcmeta).splitlines()
            for p in paths:
                if p == "version.txt" or not p:
                    continue
                presence.add(kind, p, vid, f"misode-mcmeta-{branch}", f"mcmeta:{vid}:{branch}")
            if n % 50 == 0:
                print(f"mcmeta {branch} {n}/{len(commits)}", flush=True)

    prism_observed = 0
    for vid in prism_ids:
        p = minecraft_data / "data/pc" / vid / "sounds.json"
        if not p.exists() or vid not in order_set:
            continue
        try:
            arr = json.loads(p.read_text())
        except Exception:
            continue
        if isinstance(arr, list):
            for row in arr:
                if isinstance(row, dict) and isinstance(row.get("name"), str):
                    presence.add("sound_event", row["name"], vid, "prismarine-minecraft-data", f"minecraft-data:{vid}:sounds.json")
            prism_observed += 1

    versions = []
    for vid in order:
        if vid in official_by_id:
            m = official_by_id[vid]
            vj = version_json.get(vid, {})
            ai = vj.get("assetIndex", {})
            versions.append({
                "id": vid, "type": m.get("type", ""), "releaseTime": m.get("releaseTime", ""),
                "assetIndexId": ai.get("id", ""), "assetIndexSha1": ai.get("sha1", ""),
                "dataVersion": int(vj.get("worldVersion") or 0),
            })
        else:
            vjpath = minecraft_data / "data/pc" / vid / "version.json"
            pv = {}
            if vjpath.exists():
                try:
                    pv = json.loads(vjpath.read_text())
                except Exception:
                    pass
            versions.append({"id": vid, "type": "historical-normalized", "protocol": int(pv.get("version") or 0) if isinstance(pv, dict) else 0})

    idx_of = {v: i for i, v in enumerate(order)}
    features = {}
    for kind, items in presence.present.items():
        fam = {}
        for ident, pv in items.items():
            pv = {v for v in pv if v in idx_of}
            if not pv:
                continue
            ordered = sorted(pv, key=idx_of.get)
            fam[ident] = {
                "id": ident, "kind": kind, "firstSeen": ordered[0], "lastSeen": ordered[-1],
                "intervals": compress_intervals(pv, order),
                "sources": sorted(presence.sources[kind][ident]),
                "evidence": sorted(presence.evidence[kind][ident]),
            }
        features[kind] = dict(sorted(fam.items()))

    asset_index_meta = {}
    for h, meta in asset_meta.items():
        obj = asset_indexes[h]
        asset_index_meta[h] = {
            "id": str(meta.get("id", "")), "sha1": h, "size": int(meta.get("size") or 0), "url": meta.get("url", ""),
            "objectCount": len(obj.get("objects", {})),
        }

    providers = [{
        "id": "vanillabackport", "name": "Vanilla Backport", "modIds": ["vanillabackport"],
        "packagePrefixes": ["com.blackgear.vanillabackport"], "supportedLoaders": ["fabric", "forge", "neoforge", "quilt"],
        "homepage": "https://www.curseforge.com/minecraft/mc-mods/vanillabackport",
        "sourceRepo": "https://github.com/ItsBlackGear/VanillaBackport",
        "notes": [
            "Compatibility/provider inventory only; never used as authority for vanilla lineage.",
            "External provider wins only after feature ownership is proven for the exact installed lane; otherwise fail closed and audit before suppressing embedded fallback.",
        ],
    }]

    latest_release = manifest.get("latest", {}).get("release", "")
    latest_snapshot = manifest.get("latest", {}).get("snapshot", "")
    atlas = {
        "schema": SCHEMA,
        "generatedAt": dt.datetime.now(dt.timezone.utc).isoformat().replace("+00:00", "Z"),
        "latestRelease": latest_release, "latestSnapshot": latest_snapshot,
        "sources": [
            {"id": "mojang-manifest", "kind": "official", "url": MOJANG_MANIFEST, "revision": sha1(manifest_raw), "role": "authority for official version lineage/version JSONs"},
            {"id": "mojang-assets", "kind": "official", "url": "https://resources.download.minecraft.net/", "role": "authority for exact external asset hashes/objects via asset indexes"},
            {"id": "misode-mcmeta", "kind": "derived", "url": "https://github.com/misode/mcmeta", "revision": revisions["misode-mcmeta-summary"], "role": "generated vanilla registries, sounds definitions, data/resource JSON history from 1.14+"},
            {"id": "prismarine-minecraft-data", "kind": "derived", "url": "https://github.com/PrismarineJS/minecraft-data", "revision": revisions["prismarine-minecraft-data"], "role": "normalized historical sound/protocol coverage including pre-1.14 versions"},
        ],
        "versions": versions,
        "features": features,
        "assetIndexes": asset_index_meta,
        "providers": providers,
        "coverage": {
            "oldestVersion": order[0] if order else "", "newestVersion": order[-1] if order else "",
            "officialVersions": len(official_ids), "historicalNormalizedVersions": len(prism_only),
            "soundEvents": len(features.get("sound_event", {})),
            "soundDefinitions": len(features.get("sound_definition", {})),
            "registryEntries": len(features.get("registry_entry", {})),
            "externalAssetPaths": len(features.get("external_asset", {})),
            "resourcePaths": len(features.get("resource_path", {})),
            "dataPaths": len(features.get("data_path", {})),
            "assetIndexes": len(asset_index_meta),
            "additionalKinds": sorted(k for k in features if k not in {"sound_event", "sound_definition", "registry_entry", "external_asset", "resource_path", "data_path"}),
            "completenessContract": "Metadata/provenance complete for indexed sources; exact official asset index sidecars retained; large Mojang client/server/OGG bytes hydrate and hash-verify on demand instead of being duplicated for every version. Coverage before mcmeta's 1.14 floor uses normalized Prismarine observations and is labeled derived/historical rather than fabricated as official lineage.",
        },
    }

    atlas_path = out / "VANILLA-ATLAS.json"
    atlas_path.write_text(json.dumps(atlas, indent=2, sort_keys=False) + "\n")

    ovm = {}
    for vid, vj in version_json.items():
        ovm[vid] = {
            "worldVersion": vj.get("worldVersion"), "javaVersion": vj.get("javaVersion"), "assetIndex": vj.get("assetIndex"),
            "downloads": {k: {x: val.get(x) for x in ("sha1", "size", "url")} for k, val in vj.get("downloads", {}).items()},
            "logging": vj.get("logging", {}),
        }
    (out / "OFFICIAL-VERSION-METADATA.json").write_text(json.dumps(ovm, indent=2) + "\n")

    if args.with_provider_inventory:
        inv = build_provider_inventory(vanilla_backport, revisions["vanillabackport-1.20.1"])
        pdir = out / "providers"
        pdir.mkdir(exist_ok=True)
        (pdir / "vanillabackport-1.20.1-source-inventory.json").write_text(json.dumps(inv, indent=2) + "\n")

    receipt = {
        "schema": 1,
        "generatedAt": atlas["generatedAt"],
        "atlasSchema": SCHEMA,
        "latestRelease": latest_release,
        "latestSnapshot": latest_snapshot,
        "sourceRevisions": revisions,
        "counts": atlas["coverage"],
        "goatHornSanity": sound_sanity(atlas, "minecraft:item.goat_horn.play", "1.20.1"),
        "screamingGoatSanity": sound_sanity(atlas, "minecraft:entity.goat.screaming.horn_break", "1.20.1"),
    }
    (out / "BUILD-RECEIPT.json").write_text(json.dumps(receipt, indent=2) + "\n")
    (out / "SOURCES.md").write_text(sources_md(atlas, revisions))

    files = [p for p in out.rglob("*") if p.is_file() and p.name != "SHA256SUMS.txt"]
    lines = [f"{sha256_file(p)}  {p.relative_to(out).as_posix()}" for p in sorted(files)]
    (out / "SHA256SUMS.txt").write_text("\n".join(lines) + "\n")
    print(json.dumps(receipt, indent=2))


def build_provider_inventory(repo: Path, revision: str):
    subprocess.run(["git", "checkout", "--detach", "origin/1.20.1"], cwd=repo, check=True, stdout=subprocess.DEVNULL)
    ids = defaultdict(set)
    files = []
    rx_ns = re.compile(r"(?:minecraft:)?[a-z0-9_.-]+(?:[/:][a-z0-9_.\-/]+)+")
    rx_sound = re.compile(r"(?:minecraft:)?(?:ambient|block|entity|item|music|ui|weather)\.[a-z0-9_.-]+")
    for p in repo.rglob("*"):
        if not p.is_file() or ".git" in p.parts:
            continue
        rel = p.relative_to(repo).as_posix()
        files.append(rel)
        if p.suffix.lower() not in {".java", ".json", ".toml", ".mcmeta", ".properties", ".mixins", ".accesswidener"}:
            continue
        try:
            text = p.read_text(errors="ignore")
        except Exception:
            continue
        for x in rx_ns.findall(text):
            if len(x) <= 180:
                ids["resource_or_registry_id"].add(x)
        for x in rx_sound.findall(text):
            if len(x) <= 180:
                ids["sound_like_id"].add(ns(x))
    return {
        "schema": 1, "provider": "vanillabackport", "branch": "1.20.1", "revision": revision,
        "sourceRepo": "https://github.com/ItsBlackGear/VanillaBackport",
        "fileCount": len(files), "identifiers": {k: sorted(v) for k, v in ids.items()},
        "warning": "Conservative static candidate inventory. Presence is ownership evidence; absence is not proof of non-ownership. Runtime/provider-present QA remains mandatory.",
    }


def feature_present(atlas, kind, ident, version):
    f = atlas.get("features", {}).get(kind, {}).get(ident)
    if not f:
        return False
    order = [v["id"] for v in atlas["versions"]]
    pos = {v: i for i, v in enumerate(order)}
    if version not in pos:
        return False
    n = pos[version]
    for iv in f.get("intervals", []):
        if iv["from"] in pos and iv["to"] in pos and pos[iv["from"]] <= n <= pos[iv["to"]]:
            return True
    return False


def sound_sanity(atlas, ident, version):
    return {
        "id": ident, "version": version,
        "registered": feature_present(atlas, "sound_event", ident, version),
        "defined": feature_present(atlas, "sound_definition", ident, version),
    }


def sources_md(atlas, revisions):
    return f"""# Vanilla Feature Atlas source policy\n\nGenerated: {atlas['generatedAt']}\n\n## Authority order\n\n1. **Mojang version manifest, version JSONs and asset indexes** are the authority for official version lineage and exact external asset hashes.\n2. **misode/mcmeta** is a derived, version-controlled mirror of Mojang-generated registries/data/assets from 1.14 onward. Revisions are pinned in `BUILD-RECEIPT.json`.\n3. **PrismarineJS/minecraft-data** supplies normalized historical observations, especially before mcmeta's 1.14 floor. It is labeled derived/historical and does not overwrite contradictory official evidence.\n4. **Vanilla Backport and other backport mods** are compatibility providers only. They never define what vanilla historically was.\n\n## Sound identity contract\n\nA sound is three separate things: **SoundEvent registration**, **`sounds.json` definition**, and **referenced external OGG object(s)**. The atlas tracks these independently. A log line saying `Missing sound for event` does not prove an OGG is missing; a real registered vanilla event can intentionally or historically lack a `sounds.json` definition.\n\n## Completeness contract\n\nThe atlas is metadata/provenance complete for its indexed sources. Exact Mojang asset indexes are bundled by SHA-1. Heavy client/server JARs and OGG objects are hydrated on demand and must verify against Mojang's published SHA-1 before use; they are not wastefully duplicated for every Minecraft version.\n\n## Pinned revisions\n\n```json\n{json.dumps(revisions, indent=2)}\n```\n"""


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--out", default="out/vanilla-atlas")
    ap.add_argument("--work", default=".atlas-work")
    ap.add_argument("--workers", type=int, default=24)
    ap.add_argument("--with-provider-inventory", action="store_true")
    args = ap.parse_args()
    build(args)


if __name__ == "__main__":
    main()
