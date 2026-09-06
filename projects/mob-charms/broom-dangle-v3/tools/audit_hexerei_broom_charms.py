#!/usr/bin/env python3
"""Static contract audit for the Hexerei broom-keychain Charmsy lane."""
from __future__ import annotations
import json
import re
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
BASE = ROOT / "src/main/resources/assets/mobcharms/punchy/model_parts_items"
GEO = BASE / "geo/broom_keychain"
DEFS = BASE / "definitions/broom_keychain"
EXPECTED = {"allay", "axol", "bee", "camel", "chicken", "creeper", "puffer", "warden", "zombie"}
DANGLE_BONE = "broom_dangle"
FORBIDDEN = {"charm", "chain", "chainnn", "sword", "sharp", "cable", "gourp_chain", "gourp_chain2"}
ENUM_FOR_GEO = {"allay":"ALLAY", "axol":"AXOLOTL", "bee":"BEE", "camel":"CAMEL", "chicken":"CHICKEN", "creeper":"CREEPER", "puffer":"PUFFERFISH", "warden":"WARDEN", "zombie":"ZOMBIE"}
MASCOT_ROOT = {"allay":"allay", "axol":"axol", "bee":"bee", "camel":"camel", "chicken":"chicken", "creeper":"creeper2", "puffer":"puffer", "warden":"warden", "zombie":"zombie"}
SOURCE_DEF = {name: name + ".json" for name in EXPECTED}


def physics_map(raw):
    result={}
    if isinstance(raw, list):
        for x in raw:
            if isinstance(x, dict): result.update(x)
    elif isinstance(raw, dict):
        result.update(raw)
    return result


def physics_keys(raw):
    return set(physics_map(raw))


def main():
    errors=[]
    geo_names={p.name.removesuffix('.geo.json') for p in GEO.glob('*.geo.json')}
    def_names={p.stem for p in DEFS.glob('*.json')}
    if geo_names != EXPECTED: errors.append(f"geo set mismatch: {sorted(geo_names ^ EXPECTED)}")
    if def_names != EXPECTED: errors.append(f"definition set mismatch: {sorted(def_names ^ EXPECTED)}")

    for name in sorted(EXPECTED):
        gp=GEO/f'{name}.geo.json'; dp=DEFS/f'{name}.json'
        if not gp.exists() or not dp.exists(): continue
        g=json.loads(gp.read_text())['minecraft:geometry'][0]
        d=json.loads(dp.read_text())
        if d.get('items') != ['#mobcharms:broom_keychain']:
            errors.append(f'{name}: wrong selector {d.get("items")}')
        if d.get('geo') != f'mobcharms:punchy/model_parts_items/geo/broom_keychain/{name}.geo.json':
            errors.append(f'{name}: wrong geo path')
        bones=g.get('bones',[])
        roots=[b for b in bones if not b.get('parent')]
        if len(roots) != 1:
            errors.append(f'{name}: root count {len(roots)}')
        else:
            root=roots[0]
            if root.get('name') != DANGLE_BONE:
                errors.append(f'{name}: root is {root.get("name")}, expected invisible {DANGLE_BONE}')
            if [round(float(x),8) for x in root.get('pivot',[0,0,0])] != [0.0,0.0,0.0]:
                errors.append(f'{name}: dangle pivot is not local origin: {root.get("pivot")}')
            if root.get('cubes'):
                errors.append(f'{name}: invisible dangle root contains visible cubes')
        names={b.get('name') for b in bones}
        if DANGLE_BONE not in names:
            errors.append(f'{name}: missing invisible dangle root')
        mascot=next((b for b in bones if b.get('name') == MASCOT_ROOT[name]), None)
        if mascot is None:
            errors.append(f'{name}: missing mascot root {MASCOT_ROOT[name]}')
        elif mascot.get('parent') != DANGLE_BONE:
            errors.append(f'{name}: mascot root parent {mascot.get("parent")} != {DANGLE_BONE}')
        bad=sorted(n for n in names if n in FORBIDDEN or (n and n.startswith('iron_chain_')))
        if bad: errors.append(f'{name}: forbidden visible sword/chain bones {bad}')
        if {'chain','chainnn',DANGLE_BONE} & set((d.get('bone_textures') or {})):
            errors.append(f'{name}: visible/physics-only chain texture leaked')
        pmap=physics_map(d.get('bone_physics'))
        bad_physics=set(pmap) & {'chain','chainnn'}
        if bad_physics: errors.append(f'{name}: visible-chain physics leaked {sorted(bad_physics)}')
        if DANGLE_BONE not in pmap:
            errors.append(f'{name}: missing whole-mascot dangle physics')
        else:
            src_def=ROOT/'src/main/resources/assets/mobcharms/punchy/model_parts_items/definitions/overlay/sword'/SOURCE_DEF[name]
            src=json.loads(src_def.read_text())
            authored=physics_map(src.get('bone_physics')).get('chain')
            if pmap[DANGLE_BONE] != authored:
                errors.append(f'{name}: dangle physics drifted from immutable Charmsy chain tuning')

    # Runtime presentation calibration. Hexerei applies a 0.25x payload scale
    # before renderItem; the bridge must compensate that while keeping each
    # mascot in a readable ~0.30-block keychain envelope.
    bridge=(ROOT/'src/main/java/dev/mobcharms/client/HexereiBroomKeychainBridge.java').read_text()
    scale_pairs=dict(re.findall(r'case\s+([A-Z_]+)\s*->\s*([0-9.]+)F', bridge))
    if set(scale_pairs) != set(ENUM_FOR_GEO.values()):
        errors.append(f'broom scale profile mismatch: {sorted(set(scale_pairs) ^ set(ENUM_FOR_GEO.values()))}')
    if 'poseStack.scale(presentationScale, presentationScale, presentationScale)' not in bridge:
        errors.append('broom presentation scale is not applied in render bridge')
    for name, enum_name in sorted(ENUM_FOR_GEO.items()):
        gp=GEO/f'{name}.geo.json'
        if not gp.exists() or enum_name not in scale_pairs:
            continue
        g=json.loads(gp.read_text())['minecraft:geometry'][0]
        mins=[float('inf')]*3; maxs=[float('-inf')]*3
        for bone in g.get('bones',[]):
            for cube in bone.get('cubes',[]):
                origin=cube.get('origin',[0,0,0]); size=cube.get('size',[0,0,0])
                for i in range(3):
                    mins[i]=min(mins[i], float(origin[i]))
                    maxs[i]=max(maxs[i], float(origin[i])+float(size[i]))
        if any(x == float('inf') for x in mins):
            errors.append(f'{name}: no cubes for presentation bound audit')
            continue
        max_px=max(maxs[i]-mins[i] for i in range(3))
        visible_blocks=max_px / 16.0 * 0.25 * float(scale_pairs[enum_name])
        if not (0.26 <= visible_blocks <= 0.34):
            errors.append(f'{name}: normalized keychain envelope {visible_blocks:.3f} blocks outside 0.26..0.34')

    mixins=json.loads((ROOT/'src/main/resources/mobcharms.mixins.json').read_text())
    if 'HexereiBroomRendererMixin' not in mixins.get('client',[]):
        errors.append('HexereiBroomRendererMixin not registered')

    if errors:
        print('HEXEREI BROOM CONTRACT: FAIL')
        for e in errors: print(' -',e)
        raise SystemExit(1)
    print('HEXEREI BROOM CONTRACT: PASS')
    print(' definitions: 9/9')
    print(' mascot geos: 9/9')
    print(' visible Charmsy weapon/chain geometry: 0')
    print(' invisible whole-mascot dangle pivot: 9/9 exact Charmsy chain tuning')
    print(' Hexerei soft client mixin: registered')
    print(' normalized mascot envelope: 0.26..0.34 blocks')

if __name__ == '__main__': main()
