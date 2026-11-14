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

# 1.21 ComponentChanges does not exist in 1.20.1. Preserve the 1.10 feature used by spell choices:
# per-spell custom_model_data and custom_name changes, stored in the same JSON object.
write('net/spell_engine/compat/item/StackChanges.java', r'''package net.spell_engine.compat.item;

import com.mojang.serialization.Codec;
import com.mojang.serialization.codecs.RecordCodecBuilder;
import net.minecraft.item.ItemStack;
import net.minecraft.text.Text;
import java.util.Optional;

public record StackChanges(Optional<Integer> custom_model_data, Optional<String> custom_name) {
    public static final StackChanges EMPTY = new StackChanges(Optional.empty(), Optional.empty());
    public static final Codec<StackChanges> CODEC = RecordCodecBuilder.create(i -> i.group(
        Codec.INT.optionalFieldOf("minecraft:custom_model_data").forGetter(StackChanges::custom_model_data),
        Codec.STRING.optionalFieldOf("minecraft:custom_name").forGetter(StackChanges::custom_name)
    ).apply(i, StackChanges::new));
    public boolean isEmpty() { return custom_model_data.isEmpty() && custom_name.isEmpty(); }
    public void apply(ItemStack stack) {
        custom_model_data.ifPresent(v -> stack.getOrCreateNbt().putInt("CustomModelData", v));
        custom_name.ifPresent(v -> {
            var parsed = Text.Serializer.fromJson(v);
            stack.setCustomName(parsed != null ? parsed : Text.literal(v));
        });
    }
}
''')

def port_choice(s):
    return s.replace('import net.minecraft.component.ComponentChanges;', 'import net.spell_engine.compat.item.StackChanges;').replace('ComponentChanges', 'StackChanges')
patch('net/spell_engine/api/spell/container/SpellChoice.java', port_choice)
for rel in ['net/spell_engine/rpg_series/item/Weapon.java','net/spell_engine/rpg_series/item/Shield.java','net/spell_engine/rpg_series/item/RangedWeapon.java']:
    patch(rel, port_choice)

# Central stack-component call sites.
for f in java.rglob('*.java'):
    s = f.read_text()
    s = re.sub(r'(\b\w+)\.get\(SpellDataComponents\.(\w+)\)', r'SpellDataComponents.get(\1, SpellDataComponents.\2)', s)
    s = re.sub(r'(\b\w+)\.set\(SpellDataComponents\.(\w+),\s*([^;]+)\);', r'SpellDataComponents.set(\1, SpellDataComponents.\2, \3);', s)
    s = re.sub(r'(\b\w+)\.contains\(SpellDataComponents\.(\w+)\)', r'SpellDataComponents.contains(\1, SpellDataComponents.\2)', s)
    s = s.replace('itemStack.applyChanges(changes);', 'changes.apply(itemStack);')
    f.write_text(s)
patch('net/spell_engine/spellbinding/spellchoice/SpellChoiceScreenHandler.java', lambda s: s.replace(
    'SpellDataComponents.set(itemStack, SpellDataComponents.SPELL_CHOICE, SpellChoice.EMPTY);',
    'SpellDataComponents.remove(itemStack, SpellDataComponents.SPELL_CHOICE);'))

# Vanilla name/rarity equivalents for 1.20.1.
for rel in ['net/spell_engine/item/ScrollItem.java','net/spell_engine/item/UniversalSpellBookItem.java']:
    def metadata(s):
        s = s.replace('import net.minecraft.component.DataComponentTypes;\n', '')
        s = re.sub(r'(\w+)\.set\(DataComponentTypes\.ITEM_NAME,\s*([^;]+)\);', r'\1.setCustomName(\2);', s)
        s = re.sub(r'(\w+)\.set\(DataComponentTypes\.RARITY,\s*([^;]+)\);', r'\1.getOrCreateNbt().putString("spell_engine_rarity", \2.name());', s)
        return s
    patch(rel, metadata)

# 1.20.1 tooltip method contracts.
for rel in ['net/spell_engine/client/SpellEngineClient.java','net/spell_engine/client/gui/SpellTooltip.java']:
    def tip_common(s):
        s = s.replace('import net.minecraft.item.tooltip.TooltipType;\n', '')
        s = s.replace('(ItemStack itemStack, TooltipType tooltipType, List<Text> lines)', '(ItemStack itemStack, List<Text> lines)')
        s = s.replace('(itemStack, tooltipType, lines)', '(itemStack, lines)')
        return s
    patch(rel, tip_common)

def scroll_tip(s):
    s = s.replace('import net.minecraft.item.tooltip.TooltipType;\n', 'import net.minecraft.client.item.TooltipContext;\nimport org.jetbrains.annotations.Nullable;\n')
    s = s.replace('public void appendTooltip(ItemStack stack, Item.TooltipContext context, List<Text> tooltip, TooltipType type) {',
                  'public void appendTooltip(ItemStack stack, @Nullable World world, List<Text> tooltip, TooltipContext context) {')
    return s
patch('net/spell_engine/item/ScrollItem.java', scroll_tip)

def block_tip(s):
    s = s.replace('import net.minecraft.item.tooltip.TooltipType;\n', 'import net.minecraft.client.item.TooltipContext;\nimport net.minecraft.world.BlockView;\n')
    s = s.replace('public void appendTooltip(ItemStack stack, Item.TooltipContext context, List<Text> tooltip, TooltipType options) {',
                  'public void appendTooltip(ItemStack stack, @org.jetbrains.annotations.Nullable BlockView world, List<Text> tooltip, TooltipContext options) {')
    s = s.replace('super.appendTooltip(stack, context, tooltip, options);', 'super.appendTooltip(stack, world, tooltip, options);')
    s = s.replace('import com.mojang.serialization.MapCodec;\n', '')
    s = s.replace('    public static final MapCodec<SpellBindingBlock> CODEC = createCodec(SpellBindingBlock::new);\n    public MapCodec<SpellBindingBlock> getCodec() {\n        return CODEC;\n    }\n\n', '')
    return s
patch('net/spell_engine/spellbinding/SpellBindingBlock.java', block_tip)
