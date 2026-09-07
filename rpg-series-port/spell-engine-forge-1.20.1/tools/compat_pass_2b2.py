#!/usr/bin/env python3
from pathlib import Path
import json, re, sys

if len(sys.argv) != 3:
    raise SystemExit('usage: compat_pass_2.py <generated-port-root> <spell-engine-1.20.1-baseline>')
root = Path(sys.argv[1]).resolve()
base = Path(sys.argv[2]).resolve()
java = root / 'common/src/main/java'
resources = root / 'common/src/main/resources'


def path(rel): return java / rel
def read(rel): return path(rel).read_text()
def write(rel, text):
    f = path(rel); f.parent.mkdir(parents=True, exist_ok=True); f.write_text(text)
def patch(rel, fn):
    f = path(rel)
    if f.exists(): f.write_text(fn(f.read_text()))

# Rewrite component attribute types and 1.21 modifier construction.
for f in java.rglob('*.java'):
    s = f.read_text()
    if 'AttributeModifiersComponent' in s or 'AttributeModifierSlot' in s:
        s = s.replace('import net.minecraft.component.type.AttributeModifiersComponent;', 'import net.spell_engine.compat.item.AttributeModifierSet;')
        s = s.replace('import net.minecraft.component.type.AttributeModifierSlot;', 'import net.minecraft.entity.EquipmentSlot;')
        s = s.replace('AttributeModifiersComponent', 'AttributeModifierSet')
        s = s.replace('AttributeModifierSlot.MAINHAND', 'EquipmentSlot.MAINHAND').replace('AttributeModifierSlot.ANY', 'null')
        s = s.replace('AttributeModifierSlot ', 'EquipmentSlot ')
        s = s.replace('AttributeModifierSlot.forEquipmentSlot(slot.getEquipmentSlot())', 'slot.getEquipmentSlot()')
        s = s.replace('AttributeModifierSet.DEFAULT', 'AttributeModifierSet.EMPTY').replace('.modifiers()', '.entries()')
    # All target 1.10.2 constructors at this point use a UUID-or-Identifier identity plus amount + operation.
    s = s.replace('new EntityAttributeModifier(', 'net.spell_engine.compat.item.AttributeCompat.modifier(')
    s = s.replace('instance.removeModifier(modifierId);', 'instance.removeModifier(net.spell_engine.compat.item.AttributeCompat.uuid(modifierId));')
    f.write_text(s)

write('net/spell_engine/utils/AttributeModifierUtil.java', r'''package net.spell_engine.utils;
import net.minecraft.entity.EquipmentSlot;import net.minecraft.entity.attribute.*;import net.minecraft.item.ItemStack;
public final class AttributeModifierUtil{private AttributeModifierUtil(){}public static double flat(ItemStack s,EntityAttribute a,EquipmentSlot slot){double v=0;for(var m:s.getAttributeModifiers(slot).get(a))if(m.getOperation()==EntityAttributeModifier.Operation.ADDITION)v+=m.getValue();return v;}public static double multiplyBase(ItemStack s,EntityAttribute a,EquipmentSlot slot){double v=0;for(var m:s.getAttributeModifiers(slot).get(a))if(m.getOperation()==EntityAttributeModifier.Operation.MULTIPLY_BASE)v+=m.getValue();return v;}public static double multiplyTotal(ItemStack s,EntityAttribute a,EquipmentSlot slot){double v=1;for(var m:s.getAttributeModifiers(slot).get(a))if(m.getOperation()==EntityAttributeModifier.Operation.MULTIPLY_TOTAL)v*=1+m.getValue();return v-1;}}
''')

# 1.20.1 ConfigUtil returns our slot-aware builder.
def config_util(s):
    s = s.replace('AttributeModifierSet.Builder componentBuilder = AttributeModifierSet.builder();', 'AttributeModifierSet.Builder componentBuilder = AttributeModifierSet.builder();')
    s = s.replace('componentBuilder.add(modifier.attribute(), modifier.modifier(), null);', 'componentBuilder.add(modifier.attribute(), modifier.modifier(), null);')
    return s
patch('net/spell_engine/rpg_series/config/ConfigUtil.java', config_util)

for rel in ['net/spell_engine/rpg_series/item/Weapon.java','net/spell_engine/rpg_series/item/Shield.java','net/spell_engine/rpg_series/item/RangedWeapon.java']:
    def registration(s):
        if 'import net.spell_engine.compat.item.ItemAttributeCompat;' not in s:
            s = s.replace('import net.spell_engine.Platform;', 'import net.spell_engine.Platform;\nimport net.spell_engine.compat.item.ItemAttributeCompat;')
        s = re.sub(r'\n\s*if \(entry\.spellChoice != null\) \{\s*settings\.component\(SpellDataComponents\.SPELL_CHOICE, entry\.spellChoice\);\s*\}', '', s)
        s = re.sub(r'\n\s*if \(entry\.spellContainer != null\) \{\s*settings\.component\(SpellDataComponents\.SPELL_CONTAINER, entry\.spellContainer\);\s*\}', '', s)
        s = s.replace('var settings = new Item.Settings()\n                    .attributeModifiers(attributesFrom(config));', 'var settings = new Item.Settings();')
        s = s.replace('Registry.register(Registries.ITEM, entry.id(), item);', 'Registry.register(Registries.ITEM, entry.id(), item);\n            ItemAttributeCompat.set(item, attributesFrom(config));\n            if (entry.spellChoice != null) SpellDataComponents.setDefault(item, SpellDataComponents.SPELL_CHOICE, entry.spellChoice);\n            if (entry.spellContainer != null) SpellDataComponents.setDefault(item, SpellDataComponents.SPELL_CONTAINER, entry.spellContainer);')
        s = s.replace('Registry.register(Registries.ITEM, entry.id, shield);', 'Registry.register(Registries.ITEM, entry.id, shield);\n            if (entry.spellChoice != null) SpellDataComponents.setDefault(shield, SpellDataComponents.SPELL_CHOICE, entry.spellChoice);\n            if (entry.spellContainer != null) SpellDataComponents.setDefault(shield, SpellDataComponents.SPELL_CONTAINER, entry.spellContainer);')
        s = s.replace('Registry.register(Registries.ITEM, entry.id, item);', 'Registry.register(Registries.ITEM, entry.id, item);\n            if (entry.spellChoice != null) SpellDataComponents.setDefault(item, SpellDataComponents.SPELL_CHOICE, entry.spellChoice);\n            if (entry.spellContainer != null) SpellDataComponents.setDefault(item, SpellDataComponents.SPELL_CONTAINER, entry.spellContainer);')
        s = re.sub(r'\n\s*@Override\s*\n\s*public TagKey<Block> getInverseTag\(\) \{.*?\n\s*\}', '', s, flags=re.S)
        s = re.sub(r'\n\s*@Override\s*\n\s*public ToolComponent createComponent\(TagKey<Block> tag\) \{.*?\n\s*\}', '', s, flags=re.S)
        s = s.replace('import net.minecraft.component.type.ToolComponent;\n', '')
        s = s.replace('private TagKey<Block> inverseTag;', 'private int miningLevel = 0;')
        s = s.replace('material.inverseTag = vanillaMaterial.getInverseTag();', 'material.miningLevel = vanillaMaterial.getMiningLevel();')
        if 'class CustomMaterial implements ToolMaterial' in s and 'public int getMiningLevel()' not in s:
            marker = '        @Override\n        public int getDurability() {'
            s = s.replace(marker, '        @Override\n        public int getMiningLevel() { return miningLevel; }\n\n' + marker)
        return s
    patch(rel, registration)

# Armor item attributes and 1.20 material texture identity.
def armor(s):
    s = s.replace('private AttributeModifierSet attributeModifiers = AttributeModifierSet.builder().build();', 'private AttributeModifierSet attributeModifiers = AttributeModifierSet.EMPTY;')
    s = s.replace('public AttributeModifierSet getAttributeModifiers() {\n            return this.attributeModifiers;\n        }', 'public AttributeModifierSet getConfiguredAttributes() { return this.attributeModifiers; }\n\n        @Override public com.google.common.collect.Multimap<net.minecraft.entity.attribute.EntityAttribute, net.minecraft.entity.attribute.EntityAttributeModifier> getAttributeModifiers(net.minecraft.entity.EquipmentSlot slot) {\n            var b=com.google.common.collect.ImmutableMultimap.<net.minecraft.entity.attribute.EntityAttribute,net.minecraft.entity.attribute.EntityAttributeModifier>builder(); b.putAll(super.getAttributeModifiers(slot)); b.putAll(attributeModifiers.forSlot(slot)); return b.build();\n        }')
    s = re.sub(r'public Identifier getFirstLayerId\(\) \{.*?\n        \}', 'public Identifier getFirstLayerId() { return new Identifier(customMaterial.value().getName());\n        }', s, flags=re.S)
    s = s.replace('import net.spell_engine.mixin.item.ArmorMaterialLayerAccessor;\n', '')
    return s
patch('net/spell_engine/rpg_series/item/Armor.java', armor)
write('net/spell_engine/rpg_series/item/ConfigurableAttributes.java', 'package net.spell_engine.rpg_series.item; import net.spell_engine.compat.item.AttributeModifierSet; public interface ConfigurableAttributes { void setAttributes(AttributeModifierSet attributes); }\n')
write('net/spell_engine/mixin/item/ArmorMaterialLayerAccessor.java', 'package net.spell_engine.mixin.item; public interface ArmorMaterialLayerAccessor { }\n')

# Held-equipment glow: on 1.20.1 ask the item for its real slot-specific maps.
def glow(s):
    s = s.replace('import net.minecraft.component.DataComponentTypes;\n', '').replace('import net.minecraft.component.type.AttributeModifiersComponent;\n', '')
    old = re.compile(r'    public static boolean isHeldEquipment\(ItemStack stack\) \{.*?\n    \}', re.S)
    new = '''    public static boolean isHeldEquipment(ItemStack stack) {
        return !stack.getAttributeModifiers(EquipmentSlot.MAINHAND).isEmpty()
                || !stack.getAttributeModifiers(EquipmentSlot.OFFHAND).isEmpty();
    }'''
    return old.sub(new, s, count=1)
patch('net/spell_engine/api/effect/GlowingItemStatusEffect.java', glow)
