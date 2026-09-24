#!/usr/bin/env python3
from pathlib import Path
import re, sys
if len(sys.argv)!=3: raise SystemExit('usage: compat_pass_4a.py <generated-port-root> <baseline>')
root=Path(sys.argv[1]).resolve(); java=root/'common/src/main/java'
def p(rel): return java/rel
def write(rel,text): f=p(rel); f.parent.mkdir(parents=True,exist_ok=True); f.write_text(text)
def patch(rel,fn):
 f=p(rel)
 if f.exists(): f.write_text(fn(f.read_text()))
def rep(rel, pairs):
 def fn(s):
  for a,b in pairs: s=s.replace(a,b)
  return s
 patch(rel,fn)

write('net/spell_engine/compat/registry/RegistryCompat.java', r'''package net.spell_engine.compat.registry;
import net.minecraft.registry.Registry;
import net.minecraft.registry.RegistryKey;
import net.minecraft.registry.entry.RegistryEntry;
import net.minecraft.util.Identifier;
import java.util.Optional;
public final class RegistryCompat {
    private RegistryCompat() { }
    public static <T> Optional<RegistryEntry.Reference<T>> entry(Registry<T> registry, Identifier id) {
        if (registry == null || id == null) return Optional.empty();
        return registry.getEntry(RegistryKey.of(registry.getKey(), id));
    }
}
''')
write('net/spell_engine/compat/entity/ReachCompat.java', r'''package net.spell_engine.compat.entity;
import net.minecraft.entity.attribute.EntityAttribute;
import net.minecraft.entity.player.PlayerEntity;
import java.lang.reflect.Method;
public final class ReachCompat {
    private static volatile Object forgeEntityReach;
    private static volatile boolean resolved;
    private ReachCompat() { }
    public static double entityInteractionRange(PlayerEntity player) {
        if (!resolved) resolveForgeReach();
        Object attr = forgeEntityReach;
        if (attr instanceof EntityAttribute entityAttribute) {
            try { return player.getAttributeValue(entityAttribute); } catch (Throwable ignored) { }
        }
        return 3.0D;
    }
    private static synchronized void resolveForgeReach() {
        if (resolved) return;
        resolved = true;
        try {
            Class<?> forgeMod = Class.forName("net.minecraftforge.common.ForgeMod");
            Object registryObject = forgeMod.getField("ENTITY_REACH").get(null);
            Method get = registryObject.getClass().getMethod("get");
            forgeEntityReach = get.invoke(registryObject);
        } catch (Throwable ignored) { forgeEntityReach = null; }
    }
}
''')
write('net/spell_engine/compat/util/CollectionsCompat.java', r'''package net.spell_engine.compat.util;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
public final class CollectionsCompat {
    private CollectionsCompat() { }
    public static <T> List<T> reversedCopy(List<T> input) {
        var copy = new ArrayList<T>(input);
        Collections.reverse(copy);
        return copy;
    }
}
''')

entry_pairs = {
'net/spell_engine/internals/container/SpellContainerSource.java': [('registry.getEntry(id)', 'net.spell_engine.compat.registry.RegistryCompat.entry(registry, id)')],
'net/spell_engine/api/spell/container/SpellContainerHelper.java': [('SpellRegistry.from(world).getEntry(id)', 'net.spell_engine.compat.registry.RegistryCompat.entry(SpellRegistry.from(world), id)')],
'net/spell_engine/internals/delivery/SpellDelivery.java': [('Registries.STATUS_EFFECT.getEntry(id)', 'net.spell_engine.compat.registry.RegistryCompat.entry(Registries.STATUS_EFFECT, id)')],
'net/spell_engine/internals/delivery/melee/Melee.java': [('SpellRegistry.from(caster.getWorld()).getEntry(spellId)', 'net.spell_engine.compat.registry.RegistryCompat.entry(SpellRegistry.from(caster.getWorld()), spellId)'), ('SpellRegistry.from(world).getEntry(spellId)', 'net.spell_engine.compat.registry.RegistryCompat.entry(SpellRegistry.from(world), spellId)')],
'net/spell_engine/internals/casting/MobCastController.java': [('SpellRegistry.from(entity.getWorld()).getEntry(id)', 'net.spell_engine.compat.registry.RegistryCompat.entry(SpellRegistry.from(entity.getWorld()), id)')],
'net/spell_engine/internals/casting/SpellCasting.java': [('SpellRegistry.from(player.getWorld()).getEntry(spellId)', 'net.spell_engine.compat.registry.RegistryCompat.entry(SpellRegistry.from(player.getWorld()), spellId)')],
'net/spell_engine/internals/casting/SpellCastInteractor.java': [('SpellRegistry.from(player.getWorld()).getEntry(spellId)', 'net.spell_engine.compat.registry.RegistryCompat.entry(SpellRegistry.from(player.getWorld()), spellId)')],
'net/spell_engine/internals/casting/SpellCast.java': [('SpellRegistry.from(world).getEntry(id)', 'net.spell_engine.compat.registry.RegistryCompat.entry(SpellRegistry.from(world), id)'), ('SpellRegistry.from(world).getEntry(new Identifier(sync.i()))', 'net.spell_engine.compat.registry.RegistryCompat.entry(SpellRegistry.from(world), new Identifier(sync.i()))')],
'net/spell_engine/internals/SpellExecution.java': [('Registries.ATTRIBUTE.getEntry(new Identifier(impact.attribute))', 'net.spell_engine.compat.registry.RegistryCompat.entry(Registries.ATTRIBUTE, new Identifier(impact.attribute))')],
'net/spell_engine/internals/impact/SpellImpacts.java': [('Registries.STATUS_EFFECT.getEntry(id)', 'net.spell_engine.compat.registry.RegistryCompat.entry(Registries.STATUS_EFFECT, id)')],
'net/spell_engine/internals/impact/SpellEstimation.java': [('Registries.ATTRIBUTE.getEntry(new Identifier(impact.attribute))', 'net.spell_engine.compat.registry.RegistryCompat.entry(Registries.ATTRIBUTE, new Identifier(impact.attribute))')],
'net/spell_engine/rpg_series/item/Weapon.java': [('Registries.ATTRIBUTE.getEntry(attributeId)', 'net.spell_engine.compat.registry.RegistryCompat.entry(Registries.ATTRIBUTE, attributeId)')],
'net/spell_engine/rpg_series/item/Armor.java': [('Registries.ATTRIBUTE.getEntry(new Identifier(attribute.attribute))', 'net.spell_engine.compat.registry.RegistryCompat.entry(Registries.ATTRIBUTE, new Identifier(attribute.attribute))')],
'net/spell_engine/api/entity/SpellEntityPredicates.java': [('Registries.STATUS_EFFECT.getEntry(effectId)', 'net.spell_engine.compat.registry.RegistryCompat.entry(Registries.STATUS_EFFECT, effectId)'), ('Registries.STATUS_EFFECT.getEntry(id)', 'net.spell_engine.compat.registry.RegistryCompat.entry(Registries.STATUS_EFFECT, id)')],
'net/spell_engine/rpg_series/config/ConfigUtil.java': [('Registries.ATTRIBUTE.getEntry(attributeId)', 'net.spell_engine.compat.registry.RegistryCompat.entry(Registries.ATTRIBUTE, attributeId)')],
'net/spell_engine/client/gui/HudMessages.java': [('SpellRegistry.from(world).getEntry(spellId)', 'net.spell_engine.compat.registry.RegistryCompat.entry(SpellRegistry.from(world), spellId)')],
'net/spell_engine/compat/item/AttributeModifierSet.java': [('Registries.ATTRIBUTE.getEntry(attributeId)', 'net.spell_engine.compat.registry.RegistryCompat.entry(Registries.ATTRIBUTE, attributeId)')],
'net/spell_engine/client/gui/SpellTooltip.java': [('registry.getEntry(id)', 'net.spell_engine.compat.registry.RegistryCompat.entry(registry, id)')],
'net/spell_engine/mixin/item/GrindstoneSlotOutputMixin.java': [('registry.getEntry(id)', 'net.spell_engine.compat.registry.RegistryCompat.entry(registry, id)')],
'net/spell_engine/misc/SpellEngineCommands.java': [('SpellRegistry.from(context.getSource().getWorld()).getEntry(spellId)', 'net.spell_engine.compat.registry.RegistryCompat.entry(SpellRegistry.from(context.getSource().getWorld()), spellId)')],
'net/spell_engine/spellbinding/SpellBinding.java': [('registry.getEntry(spellId)', 'net.spell_engine.compat.registry.RegistryCompat.entry(registry, spellId)'), ('registry.getEntry(existingSpellId)', 'net.spell_engine.compat.registry.RegistryCompat.entry(registry, existingSpellId)')],
'net/spell_engine/entity/SummonedEntity.java': [('Registries.ATTRIBUTE.getEntry(new Identifier(custom.id))', 'net.spell_engine.compat.registry.RegistryCompat.entry(Registries.ATTRIBUTE, new Identifier(custom.id))'), ('Registries.ATTRIBUTE.getEntry(new Identifier(entry.attribute_id))', 'net.spell_engine.compat.registry.RegistryCompat.entry(Registries.ATTRIBUTE, new Identifier(entry.attribute_id))'), ('Registries.ATTRIBUTE.getEntry(new Identifier(modifier.attribute_id))', 'net.spell_engine.compat.registry.RegistryCompat.entry(Registries.ATTRIBUTE, new Identifier(modifier.attribute_id))'), ('SpellRegistry.from(getWorld()).getEntry(new Identifier(spell.spell_id))', 'net.spell_engine.compat.registry.RegistryCompat.entry(SpellRegistry.from(getWorld()), new Identifier(spell.spell_id))')],
'net/spell_engine/entity/SpellCloud.java': [('SpellRegistry.from(this.getWorld()).getEntry(this.spellId)', 'net.spell_engine.compat.registry.RegistryCompat.entry(SpellRegistry.from(this.getWorld()), this.spellId)')],
'net/spell_engine/mixin/arrow/PersistentProjectileEntityMixin.java': [('SpellRegistry.from(arrow().getWorld()).getEntry(id)', 'net.spell_engine.compat.registry.RegistryCompat.entry(SpellRegistry.from(arrow().getWorld()), id)')],
'net/spell_engine/entity/SpellProjectile.java': [('SpellRegistry.from(this.getWorld()).getEntry(spellId)', 'net.spell_engine.compat.registry.RegistryCompat.entry(SpellRegistry.from(this.getWorld()), spellId)'), ('SpellRegistry.from(this.getWorld()).getEntry(new Identifier(spellId))', 'net.spell_engine.compat.registry.RegistryCompat.entry(SpellRegistry.from(this.getWorld()), new Identifier(spellId))')],
'net/spell_engine/mixin/client/ItemRendererMixin.java': [('Registries.ITEM.getEntry(modelId)', 'net.spell_engine.compat.registry.RegistryCompat.entry(Registries.ITEM, modelId)')],
'net/spell_engine/api/item/set/EquipmentSet.java': [('registry.getEntry(setId)', 'net.spell_engine.compat.registry.RegistryCompat.entry(registry, setId)')],
'net/spell_engine/api/item/set/EquipmentSetTooltip.java': [('EquipmentSetRegistry.from(player.getWorld()).getEntry(component)', 'net.spell_engine.compat.registry.RegistryCompat.entry(EquipmentSetRegistry.from(player.getWorld()), component)')],
'net/spell_engine/internals/cost/SpellCost.java': [('Registries.STATUS_EFFECT.getEntry(new Identifier(spell.cost.effect_id))', 'net.spell_engine.compat.registry.RegistryCompat.entry(Registries.STATUS_EFFECT, new Identifier(spell.cost.effect_id))')],
}
for rel,pairs in entry_pairs.items(): rep(rel,pairs)

rep('net/spell_engine/internals/casting/SpellCast.java', [('.removeFirst()', '.remove(0)')])
rep('net/spell_engine/internals/impact/SpellImpacts.java', [('mergedTargetModifiers.addFirst(', 'mergedTargetModifiers.add(0, ')])
rep('net/spell_engine/client/gui/SpellTooltip.java', [('for (var line : lines.reversed())', 'for (var line : net.spell_engine.compat.util.CollectionsCompat.reversedCopy(lines))'), ('spellTextLines.addFirst(', 'spellTextLines.add(0, ')])
rep('net/spell_engine/client/input/SpellHotbar.java', [('slots.addFirst(onUseKey);', 'slots.add(0, onUseKey);'), ('slots.addLast(onUseKey);', 'slots.add(onUseKey);')])
rep('net/spell_engine/internals/cost/Ammo.java', [('bundle.createNewWithContents(putBack.reversed())', 'bundle.createNewWithContents(net.spell_engine.compat.util.CollectionsCompat.reversedCopy(putBack))'), ('return tag.getTranslationKey();', 'return tag.id().toTranslationKey("tag.item");')])
rep('net/spell_engine/rpg_series/item/Armor.java', [('case ArmorItem.Type.BOOTS ->', 'case BOOTS ->'), ('case ArmorItem.Type.LEGGINGS ->', 'case LEGGINGS ->'), ('case ArmorItem.Type.CHESTPLATE ->', 'case CHESTPLATE ->'), ('case ArmorItem.Type.HELMET ->', 'case HELMET ->')])
for rel in ['net/spell_engine/internals/delivery/melee/Melee.java']:
 rep(rel, [('getLengthX()', 'getXLength()'), ('getLengthY()', 'getYLength()'), ('getLengthZ()', 'getZLength()')])
rep('net/spell_engine/internals/container/SpellContainerSource.java', [('EquipmentSlot.OFFHAND.asString()', 'EquipmentSlot.OFFHAND.getName()')])
rep('net/spell_engine/utils/WeaponCompatibility.java', [('SpellEngineMod.fallbackConfig.safeValue()', 'SpellEngineMod.fallbackConfig.value')])
rep('net/spell_engine/item/UniversalSpellBookItem.java', [('SpellContainerTemplates.config.safeValue()', 'SpellContainerTemplates.config.value')])
for rel in ['net/spell_engine/client/render/ModelEffectOperations.java','net/spell_engine/api/render/ModelFxEffectRenderer.java']:
 patch(rel, lambda s: s.replace('Math.clamp(', 'MathHelper.clamp(').replace('import net.minecraft.util.math.MathHelper;\n','import net.minecraft.util.math.MathHelper;\n') if 'Math.clamp(' in s else s)
 f=p(rel)
 if f.exists():
  s=f.read_text()
  if 'MathHelper.clamp(' in s and 'import net.minecraft.util.math.MathHelper;' not in s:
   idx=s.find('\n', s.find('package ')); s=s[:idx+1]+'import net.minecraft.util.math.MathHelper;\n'+s[idx+1:]; f.write_text(s)
for rel in ['net/spell_engine/internals/SpellParameters.java','net/spell_engine/internals/delivery/melee/Melee.java']:
 patch(rel, lambda s: re.sub(r'([A-Za-z_][A-Za-z0-9_]*)\.getEntityInteractionRange\(\)', r'net.spell_engine.compat.entity.ReachCompat.entityInteractionRange(\1)', s))
def ammo(s):
 old='''            var enchantmentQuery = needsArrow
                    ? player.getWorld().getRegistryManager().get(RegistryKeys.ENCHANTMENT).getEntry(Enchantments.INFINITY)
                    : player.getWorld().getRegistryManager().get(RegistryKeys.ENCHANTMENT).getEntry(SPELL_INFINITY);
            if (enchantmentQuery.isPresent() &&
                    EnchantmentHelper.getLevel(enchantmentQuery.get(), casterStack) > 0) { // Has infinity
                return new Result(satisfied, ammo, consume, sources);
            }'''
 new='''            var enchantmentRegistry = player.getWorld().getRegistryManager().get(RegistryKeys.ENCHANTMENT);
            var infinity = needsArrow ? Enchantments.INFINITY : enchantmentRegistry.get(SPELL_INFINITY);
            if (infinity != null && EnchantmentHelper.getLevel(infinity, casterStack) > 0) { // Has infinity
                return new Result(satisfied, ammo, consume, sources);
            }'''
 if old not in s: raise SystemExit('Ammo infinity block drifted')
 return s.replace(old,new)
patch('net/spell_engine/internals/cost/Ammo.java', ammo)
for needle in ('.removeFirst()', '.addFirst(', '.addLast(', '.reversed()', 'Math.clamp(', 'getLengthX()', 'getLengthY()', 'getLengthZ()', '.safeValue()', 'EquipmentSlot.OFFHAND.asString()', 'getEntityInteractionRange()', 'tag.getTranslationKey()'):
 found=[str(f.relative_to(java)) for f in java.rglob('*.java') if needle in f.read_text()]
 if found: raise SystemExit(f'compat pass 4a incomplete: {needle} remains in {found[:20]}')
print('Spell Engine compatibility pass 4a applied: registry identifiers + Java17 collections + target API aliases + Forge-aware reach')
