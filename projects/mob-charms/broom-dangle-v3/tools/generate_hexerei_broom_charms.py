#!/usr/bin/env python3
"""Generate mascot-only Punchy definitions for Hexerei broom keychains.

The immutable Charmsy sword source is read-only input. Each output keeps only
one mascot subtree, rebases it to a local origin, preserves its authored local
rotations/cubes, and deliberately omits all visible sword/Charmsy chain geometry.

The original Charmsy `chain` bone is also the authored whole-mascot pendulum.
Broom outputs therefore remap that physics onto an empty local `broom_dangle`
parent: Hexerei remains the sole visible tether, while Punchy still drives the
mascot with the same pendulum mode/limit/force/speed/gravity/motion tuning used
on the vanilla sword.
"""
from __future__ import annotations

import copy
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
ASSETS = ROOT / "src/main/resources/assets/mobcharms/punchy/model_parts_items"
GEO_ROOT = ASSETS / "geo"
DEF_ROOT = ASSETS / "definitions"
OUT_GEO = GEO_ROOT / "broom_keychain"
OUT_DEF = DEF_ROOT / "broom_keychain"

DANGLE_BONE = "broom_dangle"

CHARMS = {
    "allay": {"source_geo": "sword_allay.geo.json", "source_def": "allay.json", "root": "allay"},
    "axol": {"source_geo": "sword_axo.geo.json", "source_def": "axol.json", "root": "axol"},
    "bee": {"source_geo": "sword_bee.geo.json", "source_def": "bee.json", "root": "bee"},
    "camel": {"source_geo": "sword_camel.geo.json", "source_def": "camel.json", "root": "camel"},
    "chicken": {"source_geo": "sword_chicken.geo.json", "source_def": "chicken.json", "root": "chicken"},
    "creeper": {"source_geo": "sword_creeper.geo.json", "source_def": "creeper.json", "root": "creeper2"},
    "puffer": {"source_geo": "sword_puffer.geo.json", "source_def": "puffer.json", "root": "puffer"},
    "warden": {"source_geo": "sword_warden.geo.json", "source_def": "warden.json", "root": "warden"},
    "zombie": {"source_geo": "sword_zombie.geo.json", "source_def": "zombie.json", "root": "zombie"},
}


def sub3(values, origin):
    return [round(float(values[i]) - float(origin[i]), 8) for i in range(3)]


def collect_subtree(bones, root_name):
    by_parent = {}
    for bone in bones:
        by_parent.setdefault(bone.get("parent"), []).append(bone["name"])
    keep = set()
    stack = [root_name]
    while stack:
        name = stack.pop()
        if name in keep:
            continue
        keep.add(name)
        stack.extend(by_parent.get(name, ()))
    return keep


def rebase_bone(bone, anchor, root_name):
    out = copy.deepcopy(bone)
    if out["name"] == root_name:
        # Keep the mascot's authored local pose, but parent it to an invisible
        # pendulum proxy.  The proxy has no cubes/textures, so Hexerei remains
        # the only visible chain/tether.
        out["parent"] = DANGLE_BONE
    if "pivot" in out:
        out["pivot"] = sub3(out["pivot"], anchor)
    for cube in out.get("cubes", []):
        if "origin" in cube:
            cube["origin"] = sub3(cube["origin"], anchor)
        if "pivot" in cube:
            cube["pivot"] = sub3(cube["pivot"], anchor)
    return out


def remap_chain_physics(raw):
    """Move the authored visible-chain pendulum onto the invisible broom root.

    We intentionally preserve the exact Charmsy chain physics payload instead
    of inventing new constants.  Only the bone key changes.  `chainnn` is the
    decorative secondary chain branch and remains removed.
    """
    out = copy.deepcopy(raw)
    found = None
    if isinstance(out, list):
        for entry in out:
            if not isinstance(entry, dict):
                continue
            chain = entry.pop("chain", None)
            entry.pop("chainnn", None)
            if chain is not None:
                if found is not None:
                    raise RuntimeError("multiple authored chain physics entries")
                found = copy.deepcopy(chain)
        if found is not None:
            if not out:
                out.append({})
            target = next((entry for entry in out if isinstance(entry, dict)), None)
            if target is None:
                target = {}
                out.append(target)
            target[DANGLE_BONE] = found
    elif isinstance(out, dict):
        found = out.pop("chain", None)
        out.pop("chainnn", None)
        if found is not None:
            out[DANGLE_BONE] = copy.deepcopy(found)
    else:
        raise RuntimeError(f"unsupported bone_physics shape: {type(raw).__name__}")
    if found is None:
        raise RuntimeError("source Charmsy definition is missing authored chain pendulum physics")
    return out


def main():
    OUT_GEO.mkdir(parents=True, exist_ok=True)
    OUT_DEF.mkdir(parents=True, exist_ok=True)

    for out_name, spec in CHARMS.items():
        src_geo_path = GEO_ROOT / spec["source_geo"]
        src_def_path = DEF_ROOT / "overlay/sword" / spec["source_def"]
        src_geo = json.loads(src_geo_path.read_text())
        src_def = json.loads(src_def_path.read_text())

        geometry = copy.deepcopy(src_geo["minecraft:geometry"][0])
        bones = geometry["bones"]
        root_name = spec["root"]
        root = next((b for b in bones if b["name"] == root_name), None)
        if root is None:
            raise RuntimeError(f"missing mascot root {root_name} in {src_geo_path}")
        anchor = root.get("pivot", [0, 0, 0])
        keep = collect_subtree(bones, root_name)
        mascot_bones = [rebase_bone(b, anchor, root_name) for b in bones if b["name"] in keep]
        geometry["bones"] = [
            {"name": DANGLE_BONE, "pivot": [0.0, 0.0, 0.0]},
            *mascot_bones,
        ]
        desc = geometry.setdefault("description", {})
        desc["identifier"] = f"geometry.mobcharms.broom_keychain.{out_name}"
        # These are compact mascots now, not full swords.
        desc["visible_bounds_width"] = 3
        desc["visible_bounds_height"] = 3
        desc["visible_bounds_offset"] = [0, 0, 0]
        out_geo = {"format_version": src_geo.get("format_version", "1.12.0"), "minecraft:geometry": [geometry]}

        bone_textures = {
            key: value for key, value in (src_def.get("bone_textures") or {}).items()
            if key not in {"chain", "chainnn"}
        }
        out_def = {
            "items": ["#mobcharms:broom_keychain"],
            "customName": src_def["customName"],
            "geo": f"mobcharms:punchy/model_parts_items/geo/broom_keychain/{out_name}.geo.json",
            "texture": src_def["texture"],
            "bone_textures": bone_textures,
            "bone_physics": remap_chain_physics(src_def.get("bone_physics", {})),
        }

        (OUT_GEO / f"{out_name}.geo.json").write_text(json.dumps(out_geo, indent=2) + "\n")
        (OUT_DEF / f"{out_name}.json").write_text(json.dumps(out_def, indent=2) + "\n")

    print(f"generated {len(CHARMS)} broom keychain geos + {len(CHARMS)} definitions")


if __name__ == "__main__":
    main()
