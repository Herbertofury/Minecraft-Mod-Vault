#!/usr/bin/env python3
import pathlib,re,sys
root=pathlib.Path(sys.argv[1]).resolve()
# Broad, deterministic Yarn 1.21.1 -> 1.20.1 syntax seams.
count=0
for p in sorted(root.rglob('*.java')):
    t=p.read_text(encoding='utf-8'); u=t.replace('Identifier.of(', 'new Identifier(').replace('EntityAttributeModifier.Operation.ADD_MULTIPLIED_BASE','EntityAttributeModifier.Operation.MULTIPLY_BASE').replace('EntityAttributeModifier.Operation.ADD_MULTIPLIED_TOTAL','EntityAttributeModifier.Operation.MULTIPLY_TOTAL').replace('EntityAttributeModifier.Operation.ADD_VALUE','EntityAttributeModifier.Operation.ADDITION').replace('.dimensions(', '.setDimensions(')
    # ofVanilla is not present in target Yarn.
    u=re.sub(r'Identifier\.ofVanilla\("([^"]+)"\)', r'new Identifier("minecraft", "\1")', u)
    if u!=t: count+=1; p.write_text(u,encoding='utf-8')
# 1.20.1 attribute API exposes IDs through the registry, not Attribute#getIdAsString.
for p in sorted(root.rglob('*.java')):
    t=p.read_text(encoding='utf-8')
    if '.getIdAsString()' in t:
        if 'import net.minecraft.registry.Registries;' not in t:
            marker='package '+t.split('package ',1)[1].split(';',1)[0]+';\n'
            t=t.replace(marker,marker+'\nimport net.minecraft.registry.Registries;\n',1)
        t=re.sub(r'EntityAttributes\.([A-Z0-9_]+)\.getIdAsString\(\)', r'Registries.ATTRIBUTE.getId(EntityAttributes.\1).toString()', t)
        p.write_text(t,encoding='utf-8')
# Root effects preserve jump blocking through ActionImpairing; GENERIC_JUMP_STRENGTH does not exist in 1.20.1.
effects=root/'net/rogues/effect/RogueEffects.java'; t=effects.read_text(encoding='utf-8')
t=re.sub(r',\s*new AttributeModifier\(\s*Registries\.ATTRIBUTE\.getId\(EntityAttributes\.GENERIC_JUMP_STRENGTH\)\.toString\(\),\s*-?[12]F,\s*EntityAttributeModifier\.Operation\.MULTIPLY_BASE\s*\)', '', t, flags=re.S)
effects.write_text(t,encoding='utf-8')
# Strength in 1.20.1 is a raw StatusEffect, not a registry holder. Keep the current tuning with target signature.
mod=root/'net/rogues/RoguesMod.java'; t=mod.read_text(encoding='utf-8')
t=t.replace('StatusEffects.STRENGTH.value().addAttributeModifier(\n                    EntityAttributes.GENERIC_ATTACK_DAMAGE,\n                    new Identifier("minecraft", "strength"),\n                    tweaksConfig.value.rebalance_strength_attack_damage_multiplier,\n                    EntityAttributeModifier.Operation.MULTIPLY_BASE\n            );','StatusEffects.STRENGTH.addAttributeModifier(\n                    EntityAttributes.GENERIC_ATTACK_DAMAGE,\n                    "648D7064-6A60-4F59-8ABE-C2C23A6DD7A9",\n                    tweaksConfig.value.rebalance_strength_attack_damage_multiplier,\n                    EntityAttributeModifier.Operation.MULTIPLY_BASE\n            );')
mod.write_text(t,encoding='utf-8')
# ArmorMaterial became a registry-backed record after 1.20.1. Preserve all six current materials as direct interface entries.
arm=root/'net/rogues/item/armor/RogueArmors.java'; t=arm.read_text(encoding='utf-8')
pattern=re.compile(r'''    public static RegistryEntry<ArmorMaterial> material\(\n            String name, int protectionHead, int protectionChest, int protectionLegs, int protectionFeet,\n            int enchantability, RegistryEntry<SoundEvent> equipSound, Supplier<Ingredient> repairIngredient\) \{\n\n        var material = new ArmorMaterial\(.*?        return net\.spell_engine\.compat\.registry\.RegistrationBridge\.registerReference\(Registries\.ARMOR_MATERIAL, new Identifier\(RoguesMod\.NAMESPACE, name\), material\);\n    \}\n''',re.S)
replacement='''    private static final Map<ArmorItem.Type,Integer> BASE_DURABILITY=Map.of(ArmorItem.Type.HELMET,11,ArmorItem.Type.CHESTPLATE,16,ArmorItem.Type.LEGGINGS,15,ArmorItem.Type.BOOTS,13);\n    private static final class RoguesArmorMaterial implements ArmorMaterial {\n        private final String name; private final int durability; private final Map<ArmorItem.Type,Integer> protection; private final int enchantability; private final SoundEvent sound; private final Supplier<Ingredient> repair;\n        RoguesArmorMaterial(String name,int durability,int h,int c,int l,int f,int ench,SoundEvent sound,Supplier<Ingredient> repair){this.name=RoguesMod.ID+":"+name;this.durability=durability;this.protection=Map.of(ArmorItem.Type.HELMET,h,ArmorItem.Type.CHESTPLATE,c,ArmorItem.Type.LEGGINGS,l,ArmorItem.Type.BOOTS,f);this.enchantability=ench;this.sound=sound;this.repair=repair;}\n        public int getDurability(ArmorItem.Type type){return BASE_DURABILITY.get(type)*durability;} public int getProtection(ArmorItem.Type type){return protection.get(type);} public int getEnchantability(){return enchantability;} public SoundEvent getEquipSound(){return sound;} public Ingredient getRepairIngredient(){return repair.get();} public String getName(){return name;} public float getToughness(){return 0F;} public float getKnockbackResistance(){return 0F;}\n    }\n    public static RegistryEntry<ArmorMaterial> material(String name,int durability,int h,int c,int l,int f,int ench,SoundEvent sound,Supplier<Ingredient> repair){return new RegistryEntry.Direct<>(new RoguesArmorMaterial(name,durability,h,c,l,f,ench,sound,repair));}\n'''
t,n=pattern.subn(replacement,t,count=1)
if n!=1: raise SystemExit(f'armor material current shape mismatch: {n}')
for name,dur in [('rogue_armor',15),('assassin_armor',25),('netherite_assassin_armor',37),('warrior_armor',15),('berserker_armor',25),('netherite_berserker_armor',37)]:
    old=f'            "{name}",\n'; new=f'            "{name}", {dur},\n';
    if t.count(old)!=1: raise SystemExit(f'armor material declaration mismatch: {name}')
    t=t.replace(old,new,1)
t=t.replace('RogueSounds.ROGUE_ARMOR_EQUIP.entry()','RogueSounds.ROGUE_ARMOR_EQUIP.soundEvent()').replace('RogueSounds.WARRIOR_ARMOR_EQUIP.entry()','RogueSounds.WARRIOR_ARMOR_EQUIP.soundEvent()')
arm.write_text(t,encoding='utf-8')
forbidden=('Identifier.of(', 'Identifier.ofVanilla(', 'Registries.ARMOR_MATERIAL', 'GENERIC_JUMP_STRENGTH', 'ADD_MULTIPLIED_BASE', 'ADD_MULTIPLIED_TOTAL', 'ADD_VALUE')
s=[]
for p in sorted(root.rglob('*.java')):
    for i,line in enumerate(p.read_text(encoding='utf-8').splitlines(),1):
        code=line.split('//',1)[0]
        if any(x in code for x in forbidden): s.append(f'{p.relative_to(root)}:{i}:{line.strip()}')
if s: raise SystemExit('initial 1.21 API survivors:\n'+'\n'.join(s[:80]))
print(f'[Rogues 1.20.1 compat] initial API translation touched {count} Java files; armor/root/attribute seams preserved')
