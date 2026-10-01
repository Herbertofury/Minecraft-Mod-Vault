#!/usr/bin/env python3
"""Prepare Runes 1.3.2 resources for Minecraft/Forge 1.20.1.

Input is the exact upstream Runes checkout at commit 96a89bb8f51f72a9087bbf1b1fb241d59c335d72.
The script copies common resources, translates 1.21 data-pack directory/schema changes back to
1.20.1, removes loader-only Bundle API recipe gates because this port has native pouches, and fixes
filled pouch model inheritance so model overrides cannot self-reference.
"""
from __future__ import annotations

import argparse
import json
import shutil
from pathlib import Path

UPSTREAM_COMMIT = "96a89bb8f51f72a9087bbf1b1fb241d59c335d72"

DIR_RENAMES = {
    "loot_table": "loot_tables",
    "recipe": "recipes",
    "tags/block": "tags/blocks",
    "tags/item": "tags/items",
}


def mapped_relative(path: Path) -> Path:
    text = path.as_posix()
    for old, new in DIR_RENAMES.items():
        text = text.replace(f"/{old}/", f"/{new}/")
        if text.startswith(old + "/"):
            text = new + text[len(old):]
    return Path(text)


def convert_json(path: Path, data: object) -> object:
    if isinstance(data, dict):
        # 1.3.2 gates pouches on Bundle API. The 1.20.1 Forge port implements pouches natively.
        data.pop("fabric:load_conditions", None)
        data.pop("neoforge:conditions", None)

        # 1.21 recipe results use `id`; 1.20.1 uses `item`.
        if "/recipes/" in path.as_posix() and isinstance(data.get("result"), dict):
            result = data["result"]
            if "id" in result and "item" not in result:
                result["item"] = result.pop("id")

        for key, value in list(data.items()):
            data[key] = convert_json(path, value)
        return data
    if isinstance(data, list):
        return [convert_json(path, value) for value in data]
    return data


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("upstream_root", type=Path)
    parser.add_argument("dest", type=Path)
    args = parser.parse_args()

    upstream_root = args.upstream_root.resolve()
    source = upstream_root / "common" / "src" / "main" / "resources"
    dest = args.dest.resolve()
    if not source.is_dir():
        raise SystemExit(f"Missing upstream resources: {source}")

    if dest.exists():
        shutil.rmtree(dest)
    dest.mkdir(parents=True)

    for src in sorted(source.rglob("*")):
        if not src.is_file():
            continue
        rel = mapped_relative(src.relative_to(source))
        out = dest / rel
        out.parent.mkdir(parents=True, exist_ok=True)
        if src.suffix == ".json":
            data = json.loads(src.read_text(encoding="utf-8"))
            data = convert_json(out, data)
            out.write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
        else:
            shutil.copy2(src, out)

    logo = upstream_root / "logo.png"
    if logo.is_file():
        shutil.copy2(logo, dest / "icon.png")

    # Filled models must be terminal generated models, not children of base models containing overrides.
    for size in ("small", "medium", "large"):
        model = dest / "assets" / "runes" / "models" / "item" / f"{size}_rune_pouch_filled.json"
        model.write_text(json.dumps({
            "parent": "item/generated",
            "textures": {"layer0": f"runes:item/{size}_rune_pouch_filled"},
        }, indent=2) + "\n", encoding="utf-8")

    # Validate every generated JSON file.
    json_files = list(dest.rglob("*.json"))
    for file in json_files:
        json.loads(file.read_text(encoding="utf-8"))

    expected = [
        dest / "assets/runes/textures/item/small_rune_pouch.png",
        dest / "data/runes/recipes/pouch/small_rune_pouch.json",
        dest / "data/runes/tags/items/runes.json",
        dest / "data/spell_engine/tags/items/spell_quiver.json",
        dest / "runes.mixins.json",
        dest / "icon.png",
    ]
    missing = [str(path) for path in expected if not path.is_file()]
    if missing:
        raise SystemExit("Missing expected converted resources:\n" + "\n".join(missing))

    print(f"Prepared {sum(1 for p in dest.rglob('*') if p.is_file())} files; {len(json_files)} JSON files validated")


if __name__ == "__main__":
    main()
