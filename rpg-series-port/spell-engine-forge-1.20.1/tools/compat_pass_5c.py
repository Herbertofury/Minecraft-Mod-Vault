#!/usr/bin/env python3
from pathlib import Path
import json,re,sys
if len(sys.argv)!=3: raise SystemExit('usage: compat_pass_5c.py <generated-port-root> <baseline>')
root=Path(sys.argv[1]).resolve(); J=root/'common/src/main/java'; R=root/'common/src/main/resources'
def p(r): return J/r
def ed(r,fn):
 q=p(r); old=q.read_text(); new=fn(old)
 if new==old: raise SystemExit(f'pass5c transform did not match: {r}')
 q.write_text(new)
def rp(r,*pairs):
 def fn(s):
  for a,b in pairs:s=s.replace(a,b)
  return s
 ed(r,fn)

# 1.20.1 exposes dynamic dimensions through getDimensions(EntityPose), not 1.21 getBaseDimensions.
# Keep the configured width/height behavior and fall back to vanilla's target-native dimensions path.
def summon(s):
 s=s.replace('public EntityDimensions getBaseDimensions(EntityPose pose) {','public EntityDimensions getDimensions(EntityPose pose) {')
 s=s.replace('return super.getBaseDimensions(pose);','return super.getDimensions(pose);')
 s=s.replace('public boolean isImmuneToExplosion(Explosion explosion) {\n        return !isAttackableSummon() && super.isImmuneToExplosion(explosion);\n    }',
'''public boolean isImmuneToExplosion() {\n        return !isAttackableSummon() || super.isImmuneToExplosion();\n    }''')
 old='''    @Override\n    public void playSound(@Nullable SoundEvent sound) {\n        if (sound != null && behaviour != null) {\n            if (sound == hurtEvent.get())  { playConfiguredSound(behaviour.sounds.hurt);  return; }\n            if (sound == deathEvent.get()) { playConfiguredSound(behaviour.sounds.death); return; }\n        }\n        super.playSound(sound);\n    }'''
 new='''    @Override\n    public void playSound(SoundEvent sound, float volume, float pitch) {\n        if (sound != null && behaviour != null) {\n            if (sound == hurtEvent.get())  { playConfiguredSound(behaviour.sounds.hurt);  return; }\n            if (sound == deathEvent.get()) { playConfiguredSound(behaviour.sounds.death); return; }\n        }\n        if (sound != null) super.playSound(sound, volume, pitch);\n    }'''
 if old not in s: raise SystemExit('summon sound hook block missing')
 s=s.replace(old,new)
 # 1.20.1 has one knockback resistance attribute. It covers the same movement.is_pushable intent.
 s=s.replace('EntityAttributes.GENERIC_EXPLOSION_KNOCKBACK_RESISTANCE','EntityAttributes.GENERIC_KNOCKBACK_RESISTANCE')
 # pass4b accidentally translated a DefaultAttributeContainer builder add into a DataTracker write.
 s=s.replace('net.spell_engine.compat.registry.RegistryCompat.entry(Registries.ATTRIBUTE, new Identifier(custom.id)).ifPresent(e -> this.dataTracker.startTracking(e, custom.value));',
             'net.spell_engine.compat.registry.RegistryCompat.entry(Registries.ATTRIBUTE, new Identifier(custom.id)).ifPresent(e -> builder.add(e.value(), custom.value));')
 s=s.replace('this.getAttributeInstance(targetAttrOpt.get())','this.getAttributeInstance(targetAttrOpt.get().value())')
 s=s.replace('owner.getAttributeInstance(ownerAttrOpt.get())','owner.getAttributeInstance(ownerAttrOpt.get().value())')
 return s
ed('net/spell_engine/entity/SummonedEntity.java',summon)

# Common code delegates tracker lookup to the loader implementation instead of reaching through a
# version-specific ServerChunkManager private field. This is an explicit platform contract now.
def platform(s):
 marker='        <T> void registerSyncedDataRegistry(RegistryKey<Registry<T>> key, Codec<T> localCodec, Codec<T> networkCodec);'
 if marker not in s: raise SystemExit('Platform.Util marker missing')
 s=s.replace(marker, marker+'\n\n        Collection<ServerPlayerEntity> tracking(Entity entity);')
 start=s.find('    /// The server players currently tracking `entity`')
 if start==-1: raise SystemExit('Platform tracking block start missing')
 end=s.find('\n    }\n}', start)
 if end==-1: raise SystemExit('Platform tracking block end missing')
 replacement='''    /// The server players currently tracking `entity`. Loader-specific implementation must use the\n    /// real entity tracker, preserving exact packet-recipient semantics rather than distance guessing.\n    public static Collection<ServerPlayerEntity> tracking(Entity entity) {\n        return util().tracking(entity);\n'''
 s=s[:start]+replacement+s[end:]
 return s
ed('net/spell_engine/Platform.java',platform)

# Preserve modern Trigger.weapon_condition on arrows. 1.20.1 does not retain the firing weapon on
# PersistentProjectileEntity, so carry a copy in our existing ArrowExtension and persist it with NBT.
def arrow_ext(s):
 s=s.replace('import net.minecraft.registry.entry.RegistryEntry;','import net.minecraft.registry.entry.RegistryEntry;\nimport net.minecraft.item.ItemStack;\nimport org.jetbrains.annotations.Nullable;')
 marker='    boolean isInGround_SpellEngine();'
 return s.replace(marker, marker+'\n    void setFiringWeaponStack_SpellEngine(@Nullable ItemStack stack);\n    @Nullable ItemStack getFiringWeaponStack_SpellEngine();')
ed('net/spell_engine/internals/delivery/arrow/ArrowExtension.java',arrow_ext)

def arrow_mixin(s):
 s=s.replace('import net.minecraft.nbt.NbtCompound;','import net.minecraft.nbt.NbtCompound;\nimport net.minecraft.item.ItemStack;')
 s=s.replace('private static final String NBT_KEY_SPELL_ID = "spell_id";','private static final String NBT_KEY_SPELL_ID = "spell_id";\n    private static final String NBT_KEY_WEAPON = "spell_engine_firing_weapon";\n    private ItemStack spellEngine_firingWeapon = ItemStack.EMPTY;')
 s=s.replace('nbt.putString(NBT_KEY_SPELL_ID, json);','''nbt.putString(NBT_KEY_SPELL_ID, json);\n        if (!spellEngine_firingWeapon.isEmpty()) {\n            nbt.put(NBT_KEY_WEAPON, spellEngine_firingWeapon.writeNbt(new NbtCompound()));\n        }''')
 marker='''        if (nbt.contains(NBT_KEY_SPELL_ID)) {'''
 if marker not in s: raise SystemExit('arrow nbt read marker missing')
 # load weapon before existing spell list block
 s=s.replace(marker,'''        if (nbt.contains(NBT_KEY_WEAPON)) {\n            spellEngine_firingWeapon = ItemStack.fromNbt(nbt.getCompound(NBT_KEY_WEAPON));\n        } else {\n            spellEngine_firingWeapon = ItemStack.EMPTY;\n        }\n        if (nbt.contains(NBT_KEY_SPELL_ID)) {''')
 marker2='''    @Override\n    public boolean isInGround_SpellEngine() {'''
 methods='''    @Override\n    public void setFiringWeaponStack_SpellEngine(@Nullable ItemStack stack) {\n        spellEngine_firingWeapon = stack == null ? ItemStack.EMPTY : stack.copy();\n    }\n\n    @Override\n    @Nullable\n    public ItemStack getFiringWeaponStack_SpellEngine() {\n        return spellEngine_firingWeapon.isEmpty() ? null : spellEngine_firingWeapon;\n    }\n\n'''
 if marker2 not in s: raise SystemExit('arrow extension implementation marker missing')
 s=s.replace(marker2,methods+marker2)
 return s
ed('net/spell_engine/mixin/arrow/PersistentProjectileEntityMixin.java',arrow_mixin)

def ranged(s):
 marker='''        applySpellHooks(projectile, shooter);'''
 if marker not in s: raise SystemExit('RangedWeaponCompat spell hook marker missing')
 return s.replace(marker,'''        if (projectile instanceof ArrowExtension arrow) {\n            arrow.setFiringWeaponStack_SpellEngine(weaponStack);\n        }\n        applySpellHooks(projectile, shooter);''')
ed('net/spell_engine/compat/item/RangedWeaponCompat.java',ranged)
rp('net/spell_engine/internals/SpellTriggers.java',('return projectile.getWeaponStack();','return ((ArrowExtension) projectile).getFiringWeaponStack_SpellEngine();'))
# import is already present in modern source through event handling; add only if absent
q=p('net/spell_engine/internals/SpellTriggers.java'); s=q.read_text()
if 'import net.spell_engine.internals.delivery.arrow.ArrowExtension;' not in s:
 s=s.replace('package net.spell_engine.internals;','package net.spell_engine.internals;\n\nimport net.spell_engine.internals.delivery.arrow.ArrowExtension;')
 q.write_text(s)

# 1.20.1 has EntityGroup.UNDEAD rather than a vanilla entity-type tag. Create a compatibility tag
# containing the vanilla 1.20.1 undead roster, so the data-driven vulnerability system remains intact.
def tags(s):
 s=s.replace('EntityTypeTags.UNDEAD','LEGACY_UNDEAD')
 marker='public class SpellEngineEntityTags {'
 return s.replace(marker,marker+'\n    public static final TagKey<EntityType<?>> LEGACY_UNDEAD = TagKey.of(Registries.ENTITY_TYPE.getKey(), new Identifier(SpellEngineMod.ID, "legacy_undead"));')
ed('net/spell_engine/api/tags/SpellEngineEntityTags.java',tags)
tag=R/'data/spell_engine/tags/entity_types/legacy_undead.json'; tag.parent.mkdir(parents=True,exist_ok=True)
tag.write_text(json.dumps({'replace':False,'values':[
'minecraft:drowned','minecraft:giant','minecraft:husk','minecraft:phantom','minecraft:skeleton','minecraft:skeleton_horse','minecraft:stray','minecraft:wither','minecraft:wither_skeleton','minecraft:zoglin','minecraft:zombie','minecraft:zombie_horse','minecraft:zombie_villager','minecraft:zombified_piglin']},indent=2)+'\n')

# Exact 1.20.1 block navigation signature from the historical Spell Engine source.
rp('net/spell_engine/spellbinding/SpellBindingBlock.java',('protected boolean canPathfindThrough(BlockState state, NavigationType type)','public boolean canPathfindThrough(BlockState state, BlockView world, BlockPos pos, NavigationType type)'))

# Fix the exact 1.20.1 enchantment map iteration contract.
def ench(s):
 s=s.replace('import net.minecraft.item.ItemStack;','import net.minecraft.item.ItemStack;\nimport net.minecraft.registry.Registries;')
 s=s.replace('for(var entry: enchants.getEnchantments()) {\n            var id = entry.getKey().get().getValue();','for(var entry: enchants.entrySet()) {\n            var id = Registries.ENCHANTMENT.getId(entry.getKey());')
 return s
ed('net/spell_engine/mixin/criteria/EnchantedItemCriterionMixin.java',ench)

# getConfiguredAttributes is a useful helper but no longer declared by the narrowed compatibility interface.
def armor(s):
 return s.replace('''        @Override\n        public AttributeModifierSet getConfiguredAttributes()''','''        public AttributeModifierSet getConfiguredAttributes()''')
ed('net/spell_engine/rpg_series/item/Armor.java',armor)

# 1.20.1 loot number providers keep their parameters private; expose them with typed Mixins instead of
# reflection so the modern fallback-loot analyzer can preserve source enchant ranges.
(J/'net/spell_engine/mixin/loot/ConstantLootNumberProviderAccessor.java').write_text('''package net.spell_engine.mixin.loot;\nimport net.minecraft.loot.provider.number.ConstantLootNumberProvider;\nimport org.spongepowered.asm.mixin.Mixin;\nimport org.spongepowered.asm.mixin.gen.Accessor;\n@Mixin(ConstantLootNumberProvider.class)\npublic interface ConstantLootNumberProviderAccessor { @Accessor("value") float spellEngine_getValue(); }\n''')
(J/'net/spell_engine/mixin/loot/UniformLootNumberProviderAccessor.java').write_text('''package net.spell_engine.mixin.loot;\nimport net.minecraft.loot.provider.number.LootNumberProvider;\nimport net.minecraft.loot.provider.number.UniformLootNumberProvider;\nimport org.spongepowered.asm.mixin.Mixin;\nimport org.spongepowered.asm.mixin.gen.Accessor;\n@Mixin(UniformLootNumberProvider.class)\npublic interface UniformLootNumberProviderAccessor { @Accessor("min") LootNumberProvider spellEngine_getMin(); @Accessor("max") LootNumberProvider spellEngine_getMax(); }\n''')
def loot(s):
 s=s.replace('constant.value()','((ConstantLootNumberProviderAccessor) constant).spellEngine_getValue()')
 s=s.replace('uniform.min()','((UniformLootNumberProviderAccessor) uniform).spellEngine_getMin()')
 s=s.replace('uniform.max()','((UniformLootNumberProviderAccessor) uniform).spellEngine_getMax()')
 s=s.replace('lo.value()','((ConstantLootNumberProviderAccessor) lo).spellEngine_getValue()')
 s=s.replace('hi.value()','((ConstantLootNumberProviderAccessor) hi).spellEngine_getValue()')
 s=s.replace('EnchantWithLevelsLootFunction.builder(registries, numberProvider(', 'EnchantWithLevelsLootFunction.builder(numberProvider(')
 return s
ed('net/spell_engine/rpg_series/loot/LootHelper.java',loot)
# imports for accessors
q=p('net/spell_engine/rpg_series/loot/LootHelper.java'); s=q.read_text()
for imp in ['import net.spell_engine.mixin.loot.ConstantLootNumberProviderAccessor;','import net.spell_engine.mixin.loot.UniformLootNumberProviderAccessor;']:
 if imp not in s: s=s.replace('package net.spell_engine.rpg_series.loot;','package net.spell_engine.rpg_series.loot;\n'+imp,1)
q.write_text(s)

mix=R/'spell_engine.mixins.json'; data=json.loads(mix.read_text())
for m in ['loot.ConstantLootNumberProviderAccessor','loot.UniformLootNumberProviderAccessor']:
 if m not in data.get('mixins',[]): data.setdefault('mixins',[]).append(m)
mix.write_text(json.dumps(data,indent=2)+'\n')

# Guards for this pass.
for needle in ('getBaseDimensions(','isImmuneToExplosion(Explosion','GENERIC_EXPLOSION_KNOCKBACK_RESISTANCE','this.dataTracker.startTracking(e, custom.value)','getAttributeInstance(targetAttrOpt.get())','chunkLoadingManager','projectile.getWeaponStack()','EntityTypeTags.UNDEAD','enchants.getEnchantments()','constant.value()','uniform.min()','builder(registries, numberProvider('):
 hits=[str(q.relative_to(J)) for q in J.rglob('*.java') if needle in q.read_text()]
 if hits: raise SystemExit(f'pass5c incomplete {needle}: {hits[:20]}')
print('Spell Engine compatibility pass 5c applied: summon semantics + tracker platform contract + arrow weapon persistence + undead tag + loot provider access')
