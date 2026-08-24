#!/usr/bin/env python3
from pathlib import Path
import json, shutil, sys

if len(sys.argv)!=4: raise SystemExit('usage: prepare_upstream_source.py <upstream-1.20.1> <upstream-2.3.4> <common-dir>')
old=Path(sys.argv[1]); modern=Path(sys.argv[2]); common=Path(sys.argv[3])
gj=common/'src/generatedUpstream/java'; gr=common/'src/generatedUpstream/resources'
for p in (gj,gr): shutil.rmtree(p,ignore_errors=True); p.mkdir(parents=True,exist_ok=True)
shutil.copytree(old/'src/main/java',gj,dirs_exist_ok=True)
shutil.copytree(old/'src/main/resources',gr,dirs_exist_ok=True)
for rel in [
 'net/fabric_extras/ranged_weapon/RangedWeaponMod.java',
 'net/fabric_extras/ranged_weapon/client/RangedWeaponAPIClient.java',
 'net/fabric_extras/ranged_weapon/client/ModelPredicateHelper.java',
 'net/fabric_extras/ranged_weapon/client/TooltipHelper.java',
 'net/fabric_extras/ranged_weapon/api/AttributeModifierIDs.java',
 'net/fabric_extras/ranged_weapon/api/RangedConfig.java',
 'net/fabric_extras/ranged_weapon/api/EntityAttributes_RangedWeapon.java',
 'net/fabric_extras/ranged_weapon/api/StatusEffects_RangedWeapon.java',
 'net/fabric_extras/ranged_weapon/api/CustomRangedWeapon.java',
 'net/fabric_extras/ranged_weapon/api/CustomBow.java',
 'net/fabric_extras/ranged_weapon/api/CustomCrossbow.java',
 'net/fabric_extras/ranged_weapon/api/CrossbowMechanics.java',
 'net/fabric_extras/ranged_weapon/internal/ScalingUtil.java',
 'net/fabric_extras/ranged_weapon/mixin/PersistentProjectileEntityMixin.java',
 'net/fabric_extras/ranged_weapon/mixin/attribute/EntityAttributesMixin.java',
 'net/fabric_extras/ranged_weapon/mixin/attribute/StatusEffectsMixin.java',
 'net/fabric_extras/ranged_weapon/mixin/attribute/LivingEntityMixin.java',
 'net/fabric_extras/ranged_weapon/mixin/item/BowItemMixin.java',
 'net/fabric_extras/ranged_weapon/mixin/item/CrossbowItemMixin.java',
 'net/fabric_extras/ranged_weapon/mixin/item/RangedWeaponItemMixin.java',
 'net/fabric_extras/ranged_weapon/mixin/client/AbstractClientPlayerEntityMixin.java']:
    (gj/rel).unlink(missing_ok=True)
(gr/'fabric.mod.json').unlink(missing_ok=True); (gr/'ranged_weapon_api.mixins.json').unlink(missing_ok=True)
assets=modern/'common/src/main/resources/assets/ranged_weapon'
if assets.exists(): shutil.copytree(assets,gr/'assets/ranged_weapon',dirs_exist_ok=True)
count=0
for p in gr.rglob('*.json'):
    with p.open('r',encoding='utf-8') as fh: json.load(fh)
    count+=1
print(f'Prepared Ranged Weapon API 1.20.1 substrate + 2.3.4 resources; validated {count} JSON resources')
