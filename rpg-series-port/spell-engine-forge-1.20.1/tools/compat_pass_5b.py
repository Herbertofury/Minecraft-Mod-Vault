#!/usr/bin/env python3
from pathlib import Path
import re,sys
if len(sys.argv)!=3: raise SystemExit('usage: compat_pass_5b.py <generated-port-root> <baseline>')
J=Path(sys.argv[1]).resolve()/'common/src/main/java'
def p(r): return J/r
def ed(r,fn):
 q=p(r); old=q.read_text(); new=fn(old)
 if new==old: raise SystemExit(f'pass5b transform did not match: {r}')
 q.write_text(new)
def rp(r,*pairs):
 def fn(s):
  for a,b in pairs:s=s.replace(a,b)
  return s
 ed(r,fn)

rp('net/spell_engine/client/render/BeamRenderer.java',('.normal(matrix, 0.0F, 1.0F, 0.0F)', '.normal(matrix.getNormalMatrix(), 0.0F, 1.0F, 0.0F)'))
rp('net/spell_engine/utils/TargetHelper.java',('ShapeContext.absent()', 'null'))
rp('net/spell_engine/utils/WeaponCompatibility.java',(' || item instanceof MaceItem', ''))
rp('net/spell_engine/api/datagen/SpellBuilder.java',('cleanse.action.status_effect.remove.id = "!" + StatusEffects.TRIAL_OMEN.getIdAsString();','cleanse.action.status_effect.remove.id = null;'))
rp('net/spell_engine/api/effect/RemoveOnHit.java',('var isInDirect = !damageSource.isDirect() || ((DamageSourceExtension)damageSource).isSpellIndirect();','var isInDirect = damageSource.getSource() != damageSource.getAttacker() || ((DamageSourceExtension)damageSource).isSpellIndirect();'))
rp('net/spell_engine/api/item/weapon/SpellWeaponItem.java',('super(toolMaterial, settings);','super(toolMaterial, 0, 0.0F, settings);'))

def staff(s):
 s=s.replace('    @Override\n    public boolean postHit(ItemStack stack, LivingEntity target, LivingEntity attacker) {\n        return true;\n    }\n\n    @Override\n    public void postDamageEntity(ItemStack stack, LivingEntity target, LivingEntity attacker) {\n        stack.damage(1, attacker, EquipmentSlot.MAINHAND);\n    }','    @Override\n    public boolean postHit(ItemStack stack, LivingEntity target, LivingEntity attacker) {\n        stack.damage(1, attacker, e -> e.sendEquipmentBreakStatus(EquipmentSlot.MAINHAND));\n        return true;\n    }')
 return s
ed('net/spell_engine/api/item/weapon/StaffItem.java',staff)

def binding(s):
 s=s.replace('                    .map(id -> net.spell_engine.compat.registry.RegistryCompat.entry(spellRegistry, id)).flatMap(Optional::stream)\n                    .filter(Optional::isPresent)\n                    .map(Optional::get)','                    .map(id -> net.spell_engine.compat.registry.RegistryCompat.entry(spellRegistry, id))\n                    .flatMap(Optional::stream)')
 s=s.replace('            var spellEntry = registry.getEntry(spellId);\n            if (spellEntry.isPresent()) {\n                var newSpellTier', '            var spellEntry = net.spell_engine.compat.registry.RegistryCompat.entry(registry, spellId);\n            if (spellEntry.isPresent()) {\n                var newSpellTier')
 return s
ed('net/spell_engine/spellbinding/SpellBinding.java',binding)

def block(s):
 s=s.replace('validateTicker(type, SpellBindingBlockEntity.ENTITY_TYPE, SpellBindingBlockEntity::tick)','SpellBindingBlock.checkType(type, SpellBindingBlockEntity.ENTITY_TYPE, SpellBindingBlockEntity::tick)')
 s=s.replace('import net.minecraft.util.ActionResult;','import net.minecraft.util.ActionResult;\nimport net.minecraft.util.Hand;')
 s=s.replace('protected ActionResult onUse(BlockState state, World world, BlockPos pos, PlayerEntity player, BlockHitResult hit)', 'public ActionResult onUse(BlockState state, World world, BlockPos pos, PlayerEntity player, Hand hand, BlockHitResult hit)')
 return s
ed('net/spell_engine/spellbinding/SpellBindingBlock.java',block)
rp('net/spell_engine/internals/delivery/ProjectileLauncher.java',('caster.getRotationVector(directionPitch, directionYaw).normalize()','Vec3d.fromPolar(directionPitch, directionYaw).normalize()'))

def ench(s):
 s=s.replace('var enchants = EnchantmentHelper.getEnchantments(stack);','var enchants = EnchantmentHelper.get(stack);')
 s=s.replace('for(var entry: enchants.getEnchantments()) {\n            var enchantment = entry.value();','for(var entry: enchants.entrySet()) {\n            var enchantment = entry.getKey();')
 return s
ed('net/spell_engine/mixin/criteria/EnchantedItemCriterionMixin.java',ench)

def hudmixin(s):
 s=s.replace('method = "renderHotbarVanilla"','method = "renderHotbar"')
 start=s.find('    @Inject(method = "renderMainHud"')
 if start!=-1:
  end=s.find('\n    }\n}', start)
  if end==-1: raise SystemExit('InGameHud renderMainHud block end missing')
  s=s[:start] + s[end+7:]
 for imp in ['import net.minecraft.client.gui.DrawContext;\n','import net.minecraft.client.render.RenderTickCounter;\n','import net.spell_engine.Platform;\n','import net.spell_engine.client.gui.HudRenderHelper;\n','import org.spongepowered.asm.mixin.injection.Inject;\n','import org.spongepowered.asm.mixin.injection.callback.CallbackInfo;\n']:
  s=s.replace(imp,'')
 return s
ed('net/spell_engine/mixin/client/render/InGameHudMixin.java',hudmixin)

def weapon(s):
 s=s.replace('Item.BASE_ATTACK_DAMAGE_MODIFIER_ID,','ItemAccessor.ATTACK_DAMAGE_MODIFIER_ID(),').replace('Item.BASE_ATTACK_SPEED_MODIFIER_ID,','ItemAccessor.ATTACK_SPEED_MODIFIER_ID(),')
 marker='\n    private static final Identifier equipmentBonusId'
 if marker in s and 'class ItemAccessor extends Item' not in s:
  s=s.replace(marker,'\n    private static abstract class ItemAccessor extends Item {\n        private ItemAccessor(Settings settings) { super(settings); }\n        static java.util.UUID ATTACK_DAMAGE_MODIFIER_ID() { return ATTACK_DAMAGE_MODIFIER_ID; }\n        static java.util.UUID ATTACK_SPEED_MODIFIER_ID() { return ATTACK_SPEED_MODIFIER_ID; }\n    }\n' + marker)
 return s
ed('net/spell_engine/rpg_series/item/Weapon.java',weapon)

def armor(s):
 s=s.replace('super(material, slot, settings);','super(material.value(), slot, settings);')
 s=s.replace('ArmorItem.Type.HELMET.getMaxDamage(durability)','11 * durability')
 s=s.replace('ArmorItem.Type.CHESTPLATE.getMaxDamage(durability)','16 * durability')
 s=s.replace('ArmorItem.Type.LEGGINGS.getMaxDamage(durability)','15 * durability')
 s=s.replace('ArmorItem.Type.BOOTS.getMaxDamage(durability)','13 * durability')
 return s
ed('net/spell_engine/rpg_series/item/Armor.java',armor)

for needle in ('normal(matrix, 0.0F','ShapeContext.absent()','instanceof MaceItem','StatusEffects.TRIAL_OMEN','damageSource.isDirect()','super(toolMaterial, settings)','postDamageEntity(','validateTicker(','getRotationVector(directionPitch, directionYaw)','EnchantmentHelper.getEnchantments(stack)','renderHotbarVanilla','BASE_ATTACK_DAMAGE_MODIFIER_ID,','getMaxDamage(durability)'):
 hits=[str(q.relative_to(J)) for q in J.rglob('*.java') if needle in q.read_text()]
 if hits: raise SystemExit(f'pass5b incomplete {needle}: {hits[:20]}')
print('Spell Engine compatibility pass 5b applied: target raycast/combat/item/block/enchantment/HUD signatures')
