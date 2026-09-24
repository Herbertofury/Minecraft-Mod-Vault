#!/usr/bin/env python3
from pathlib import Path
import re,sys
if len(sys.argv)!=3: raise SystemExit('usage: compat_pass_4b.py <generated-port-root> <baseline>')
J=Path(sys.argv[1]).resolve()/'common/src/main/java'
def f(r): return J/r
def wr(r,s): p=f(r);p.parent.mkdir(parents=True,exist_ok=True);p.write_text(s)
def ed(r,fn): p=f(r);p.exists() and p.write_text(fn(p.read_text()))
def rp(r,*pairs):
 def x(s):
  for a,b in pairs:s=s.replace(a,b)
  return s
 ed(r,x)

wr('net/spell_engine/utils/AttributeModifierUtil.java',r'''package net.spell_engine.utils;
import com.google.common.collect.*;import net.minecraft.entity.*;import net.minecraft.entity.attribute.*;import net.minecraft.entity.player.PlayerEntity;import net.minecraft.item.ItemStack;
public final class AttributeModifierUtil{private AttributeModifierUtil(){}public static Multimap<EntityAttribute,EntityAttributeModifier> modifierMultimap(ItemStack s){var o=HashMultimap.<EntityAttribute,EntityAttributeModifier>create();for(var x:EquipmentSlot.values())o.putAll(s.getAttributeModifiers(x));return o;}public static boolean hasModifier(ItemStack s,EntityAttribute a){return modifierMultimap(s).containsKey(a);}public static double flatBonusFrom(ItemStack s,EntityAttribute a){double v=0;for(var m:modifierMultimap(s).get(a))if(m.getOperation()==EntityAttributeModifier.Operation.ADDITION)v+=m.getValue();return v;}public static double multipliersOf(EntityAttribute a,LivingEntity e){double b=1,t=1;var i=e.getAttributeInstance(a);if(i!=null)for(var m:i.getModifiers())switch(m.getOperation()){case ADDITION->{}case MULTIPLY_BASE->b+=m.getValue();case MULTIPLY_TOTAL->t+=m.getValue();}return b*t;}public static boolean isItemStackEquipped(ItemStack s,PlayerEntity p){if(p.getMainHandStack().equals(s))return true;for(var x:p.getInventory().armor)if(x.equals(s))return true;for(var x:p.getInventory().offHand)if(x.equals(s))return true;return false;}}
''')
rp('net/spell_engine/compat/item/AttributeCompat.java',('return net.spell_engine.compat.item.AttributeCompat.modifier(uuid(id),String.valueOf(id),value,operation);','return new EntityAttributeModifier(uuid(id),String.valueOf(id),value,operation);'))
def ext(s):return s.replace('import net.minecraft.registry.entry.RegistryEntry;\n','').replace('private static RegistryEntry<EntityAttribute> rangedDamageAttribute()','private static EntityAttribute rangedDamageAttribute()').replace('EntityAttributes_RangedWeapon.DAMAGE.entry','EntityAttributes_RangedWeapon.DAMAGE.attribute').replace('EntityAttributes_RangedWeapon.HASTE.entry','EntityAttributes_RangedWeapon.HASTE.attribute').replace('EntityAttributes.GENERIC_ATTACK_DAMAGE.value().setTracked(true);','EntityAttributes.GENERIC_ATTACK_DAMAGE.setTracked(true);')
ed('net/spell_engine/api/spell/ExternalSpellSchools.java',ext)
def est(s):return s.replace('var attribute = school.attributeEntry;','var attribute = school.attribute;').replace('attribute = optionalAttribute.get();','attribute = optionalAttribute.get().value();')
ed('net/spell_engine/internals/impact/SpellEstimation.java',est)
wr('net/spell_engine/compat/effect/StatusEffectCompat.java',r'''package net.spell_engine.compat.effect;
import net.minecraft.entity.attribute.*;import net.minecraft.entity.effect.StatusEffect;import net.minecraft.registry.*;import net.minecraft.registry.entry.RegistryEntry;import net.minecraft.util.Identifier;import java.util.*;import java.util.function.BiConsumer;
public final class StatusEffectCompat{private StatusEffectCompat(){}public static Optional<RegistryKey<StatusEffect>> key(StatusEffect e){return Registries.STATUS_EFFECT.getKey(e);}public static RegistryEntry<StatusEffect> entry(StatusEffect e){return Registries.STATUS_EFFECT.getEntry(e);}public static void forEachAttributeModifier(StatusEffect e,int amp,BiConsumer<EntityAttribute,EntityAttributeModifier> c){for(var x:e.getAttributeModifiers().entrySet()){var m=x.getValue();c.accept(x.getKey(),new EntityAttributeModifier(m.getId(),m.getName(),e.adjustModifierAmount(amp,m),m.getOperation()));}}}
''')
def effects(s):return s.replace('.addAttributeModifier(modifier.attribute(),','.addAttributeModifier(modifier.attribute().value(),').replace('modifier.modifier().id(),','modifier.modifier().getName(),').replace('modifier.modifier().value(),','modifier.modifier().getValue(),').replace('modifier.modifier().operation());','modifier.modifier().getOperation());')
ed('net/spell_engine/api/effect/Effects.java',effects)
def enginefx(s):
 s=re.sub(r'\s*new AttributeModifier\(\s*EntityAttributes\.GENERIC_JUMP_STRENGTH\.getIdAsString\(\),\s*[-0-9.]+,\s*EntityAttributeModifier\.Operation\.MULTIPLY_TOTAL\s*\),?','',s)
 s=s.replace('EntityAttributes.GENERIC_MOVEMENT_SPEED.getIdAsString()','Registries.ATTRIBUTE.getId(EntityAttributes.GENERIC_MOVEMENT_SPEED).toString()')
 return s.replace('import net.minecraft.entity.effect.StatusEffectCategory;\n','import net.minecraft.entity.effect.StatusEffectCategory;\nimport net.minecraft.registry.Registries;\n') if 'Registries.ATTRIBUTE' in s and 'import net.minecraft.registry.Registries;' not in s else s
ed('net/spell_engine/api/effect/SpellEngineEffects.java',enginefx)
wr('net/spell_engine/api/effect/StatusEffectClassification.java',r'''package net.spell_engine.api.effect;
import net.spell_engine.PlatformEvents;import net.spell_engine.compat.effect.StatusEffectCompat;import net.minecraft.entity.attribute.*;import net.minecraft.entity.effect.StatusEffect;import net.minecraft.registry.*;import net.minecraft.registry.entry.RegistryEntry;import java.util.*;
public class StatusEffectClassification{private static final Set<EntityAttribute>A=new HashSet<>();private static final Set<RegistryKey<StatusEffect>>E=new HashSet<>();public static void init(){A.add(EntityAttributes.GENERIC_MOVEMENT_SPEED);A.add(EntityAttributes.GENERIC_FLYING_SPEED);PlatformEvents.onServerStarted(s->parse(Registries.STATUS_EFFECT));}private static void parse(Registry<StatusEffect>r){r.streamEntries().forEach(e->StatusEffectCompat.forEachAttributeModifier(e.value(),0,(a,m)->{if(A.contains(a)&&m.getValue()<0)e.getKey().ifPresent(E::add);}));}public static boolean isMovementImpairing(StatusEffect e){return StatusEffectCompat.key(e).map(E::contains).orElse(false);}public static boolean isMovementImpairing(RegistryEntry<StatusEffect>e){return isMovementImpairing(e.value());}public static boolean disablesMobAI(StatusEffect e){var a=((ActionImpairing)e).actionsAllowed();return a!=null&&!a.mobs().canUseAI();}public static boolean disablesMobAI(RegistryEntry<StatusEffect>e){return disablesMobAI(e.value());}}
''')
def actions(s):return s.replace('Collection<RegistryEntry<StatusEffect>> effects','Collection<StatusEffect> effects').replace('.map(effect -> ((ActionImpairing)effect.value()).actionsAllowed())','.map(effect -> ((ActionImpairing)effect).actionsAllowed())')
ed('net/spell_engine/api/effect/EntityActionsAllowed.java',actions)
rp('net/spell_engine/api/effect/KnockbackImmunity.java',('Collection<RegistryEntry<StatusEffect>> effects','Collection<StatusEffect> effects'))
for r in ['net/spell_engine/mixin/effect/LivingEntityKnockbackImmunity.java','net/spell_engine/mixin/effect/LivingEntityStatusEffectSync.java','net/spell_engine/mixin/action_impair/LivingEntityActionImpairing.java']:rp(r,('Map<RegistryEntry<StatusEffect>, StatusEffectInstance>','Map<StatusEffect, StatusEffectInstance>'))
for p in J.rglob('*.java'):
 s=p.read_text();s=re.sub(r'net\.spell_engine\.compat\.registry\.RegistryCompat\.entry\(Registries\.STATUS_EFFECT,\s*([^\n;]+?)\)',r'java.util.Optional.ofNullable(Registries.STATUS_EFFECT.get(\1))',s);s=s.replace('.getEffectType().value()','.getEffectType()');p.write_text(s)
rp('net/spell_engine/api/effect/InstantCast.java',('var effectKey = entry.getValue().getEffectType().getKey();','var effectKey = net.spell_engine.compat.effect.StatusEffectCompat.key(entry.getValue().getEffectType());'))
rp('net/spell_engine/api/effect/Protection.java',('var optionalKey = entry.getKey().getKey();','var optionalKey = net.spell_engine.compat.effect.StatusEffectCompat.key(entry.getKey());'))
rp('net/spell_engine/api/entity/SpellEntityPredicates.java',('StatusEffects.POISON.getKey().get().getValue()','Registries.STATUS_EFFECT.getId(StatusEffects.POISON)'))
rp('net/spell_engine/client/gui/SpellTooltip.java',('effect.forEachAttributeModifier(','net.spell_engine.compat.effect.StatusEffectCompat.forEachAttributeModifier(effect, '))
for p in J.rglob('*.java'):
 s=p.read_text().replace('.modifier().id()','.modifier().getId()').replace('.modifier().value()','.modifier().getValue()').replace('.modifier().operation()','.modifier().getOperation()').replace('EntityAttributeModifier.Operation.ADD_VALUE','EntityAttributeModifier.Operation.ADDITION').replace('EntityAttributeModifier.Operation.ADD_MULTIPLIED_BASE','EntityAttributeModifier.Operation.MULTIPLY_BASE').replace('EntityAttributeModifier.Operation.ADD_MULTIPLIED_TOTAL','EntityAttributeModifier.Operation.MULTIPLY_TOTAL').replace('Operation.ADD_VALUE','Operation.ADDITION').replace('Operation.ADD_MULTIPLIED_BASE','Operation.MULTIPLY_BASE').replace('Operation.ADD_MULTIPLIED_TOTAL','Operation.MULTIPLY_TOTAL').replace('case ADD_VALUE ->','case ADDITION ->').replace('case ADD_MULTIPLIED_BASE ->','case MULTIPLY_BASE ->').replace('case ADD_MULTIPLIED_TOTAL ->','case MULTIPLY_TOTAL ->')
 s=re.sub(r'\bchosen\.value\(\)','chosen.getValue()',s);s=re.sub(r'\bchosen\.operation\(\)','chosen.getOperation()',s);p.write_text(s)
def tick(s):return s.replace('public boolean applyUpdateEffect(LivingEntity entity, int amplifier)','public void applyUpdateEffect(LivingEntity entity, int amplifier)').replace('        return true;\n    }\n\n    @Override\n    public boolean canApplyUpdateEffect','    }\n\n    @Override\n    public boolean canApplyUpdateEffect')
ed('net/spell_engine/api/effect/TickingStatusEffect.java',tick)
def bleed(s):return s.replace('public boolean applyUpdateEffect(LivingEntity entity, int amplifier)','public void applyUpdateEffect(LivingEntity entity, int amplifier)').replace('            return true; // Damage is server-authoritative','            return; // Damage is server-authoritative').replace('        return true;\n    }','    }')
ed('net/spell_engine/api/effect/BleedStatusEffect.java',bleed)
rp('net/spell_engine/internals/impact/SpellImpacts.java',('PatternMatching.matches(instance.getEffectType(), RegistryKeys.STATUS_EFFECT, data.remove.id)','PatternMatching.matches(Registries.STATUS_EFFECT.getEntry(instance.getEffectType()), RegistryKeys.STATUS_EFFECT, data.remove.id)'),('Optional<RegistryEntry<StatusEffect>> optionalEffect = Optional.empty();','Optional<StatusEffect> optionalEffect = Optional.empty();'),('optionalEffect = Optional.of(java.util.Optional.ofNullable(Registries.STATUS_EFFECT.get(id)).get());','optionalEffect = java.util.Optional.ofNullable(Registries.STATUS_EFFECT.get(id));'),('target.setOnFireFor(data.duration);','target.setOnFireFor((int)Math.ceil(data.duration));'),('playerTarget.disableShield();','playerTarget.disableShield(true);'))
rp('net/spell_engine/mixin/registry/LivingEntityAttributesMixin.java',('info.getReturnValue().add(entry.entry);','info.getReturnValue().add(entry.attribute);'))
rp('net/spell_engine/mixin/entity/PlayerEquipmentSetMixin.java',('removeModifier(modifier.modifier().id())','removeModifier(modifier.modifier().getId())'))
rp('net/spell_engine/client/ClientNetwork.java',('registry.getEntry(packet.spellId())','net.spell_engine.compat.registry.RegistryCompat.entry(registry, packet.spellId())'))
def binding(s):return s.replace('.map(spellRegistry::getEntry)','.map(id -> net.spell_engine.compat.registry.RegistryCompat.entry(spellRegistry, id)).flatMap(Optional::stream)').replace('net.spell_engine.compat.registry.RegistryCompat.entry(registry, spellId);','registry.getEntry(spellId);')
ed('net/spell_engine/spellbinding/SpellBinding.java',binding)
for n in ('GENERIC_JUMP_STRENGTH','GENERIC_GRAVITY','.getEffectType().value()','Operation.ADD_VALUE','Operation.ADD_MULTIPLIED','school.attributeEntry'):
 q=[str(p.relative_to(J)) for p in J.rglob('*.java') if n in p.read_text()]
 if q:raise SystemExit(f'pass4b incomplete {n}: {q[:12]}')
print('Spell Engine compatibility pass 4b applied')
