#!/usr/bin/env python3
"""Fetch compact real-image assets for the current Mob Variety delta manifest.

This deliberately reuses the battle-tested provider adapters in fetch_visuals.py
without mutating the 253-project baseline bundle. Future scours can point INPUT
at a new delta manifest and keep old visual caches stable.
"""
import concurrent.futures as cf
import json
import time
from pathlib import Path

import fetch_visuals as fv

ROOT = Path(__file__).resolve().parent
REPO = ROOT.parents[1]
INPUT = REPO / "research/mob-variety/fourth-scour-2026-08-24-candidates.json"
BUILD = ROOT / "build_delta"
ASSETS = BUILD / "assets"
ASSETS.mkdir(parents=True, exist_ok=True)

# Reuse the existing compact-image functions against an isolated delta output.
fv.BUILD = BUILD
fv.ASSETS = ASSETS


def main():
    data = json.loads(INPUT.read_text(encoding="utf-8"))
    items = data.get("items", [])
    expected = int(data.get("count", len(items)))
    if len(items) != expected:
        raise SystemExit(f"manifest count mismatch: expected {expected}, got {len(items)}")

    enriched = []
    with cf.ThreadPoolExecutor(max_workers=3) as ex:
        for i, item in enumerate(ex.map(fv.enrich, items), 1):
            enriched.append(item)
            if i % 10 == 0:
                print("metadata", i, flush=True)

    built = []
    with cf.ThreadPoolExecutor(max_workers=3) as ex:
        for i, item in enumerate(ex.map(fv.assets, enriched), 1):
            built.append(item)
            if i % 10 == 0:
                print("assets", i, flush=True)

    built.sort(key=lambda x: int(x["row"]))
    cov = {
        "projects": len(built),
        "authors": sum(bool(x.get("author")) for x in built),
        "author_avatars": sum(bool(x.get("author_avatar_url")) for x in built),
        "icons": sum(bool(x.get("icon_url")) for x in built),
        "gallery_sources": sum(bool(x.get("gallery_urls")) for x in built),
        "icon_assets": sum(bool(x.get("icon_asset")) for x in built),
        "author_avatar_assets": sum(bool(x.get("author_avatar_asset")) for x in built),
        "gallery_assets": sum(bool(x.get("gallery_asset")) for x in built),
    }
    BUILD.mkdir(exist_ok=True)
    (BUILD / "visual_manifest.json").write_text(
        json.dumps(
            {
                "generated": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
                "source_manifest": str(INPUT.relative_to(REPO)),
                "watermark_before": data.get("watermark_before"),
                "coverage": cov,
                "items": built,
            },
            ensure_ascii=False,
            indent=2,
        ),
        encoding="utf-8",
    )
    print("COVERAGE", json.dumps(cov), flush=True)
    if len(built) != expected:
        raise SystemExit(f"expected {expected} delta items")


if __name__ == "__main__":
    main()
