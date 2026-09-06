#!/usr/bin/env python3
import re, sys
from pathlib import Path

if len(sys.argv) != 2:
    raise SystemExit('usage: apply-core120-api-bridges.py <java-root>')
root = Path(sys.argv[1]).resolve()
if not root.is_dir():
    raise SystemExit(f'missing Java root: {root}')

def edit(rel, transform):
    p = root / rel
    if not p.is_file():
        raise SystemExit(f'missing expected source: {rel}')
    old = p.read_text(encoding='utf-8')
    new = transform(old)
    if new == old:
        raise SystemExit(f'core 1.20 bridge made no change in {rel}')
    p.write_text(new, encoding='utf-8')

# 1.21 Item.TooltipContext does not exist in 1.20.1. Preserve the tooltip body;
# only restore the Forge/Minecraft 1.20 override signature.
def tooltip(text):
    old = 'public void appendHoverText(ItemStack stack, TooltipContext context, List<Component> tooltipComponents, TooltipFlag tooltipFlag)'
    new = 'public void appendHoverText(ItemStack stack, @Nullable Level level, List<Component> tooltipComponents, TooltipFlag tooltipFlag)'
    if old not in text:
        raise SystemExit('SpiritShardItem tooltip signature changed upstream')
    return text.replace(old, new, 1)
edit('com/sammy/malum/common/item/spirit/SpiritShardItem.java', tooltip)

# Rite registry networking is already handled by the Forge registry/NBT owner on 1.20.
# Remove the 1.21-only convenience StreamCodec from SpiritRiteType, while preserving
# the Mojang Codec, NBT save/load, rite matching and all behavior.
def rite_type(text):
    text = re.sub(r'^import io\.netty\.buffer\.ByteBuf;\n', '', text, flags=re.M)
    text = re.sub(r'^import net\.minecraft\.network\.codec\.ByteBufCodecs;\n', '', text, flags=re.M)
    text = re.sub(r'^import net\.minecraft\.network\.codec\.StreamCodec;\n', '', text, flags=re.M)
    text = re.sub(r'\n\s*public static StreamCodec<ByteBuf, SpiritRiteType> STREAM_CODEC = ByteBufCodecs\.fromCodec\(CODEC\);\n', '\n', text, count=1)
    text = text.replace('getSpirits().getLast()', 'getSpirits().get(getSpirits().size() - 1)')
    text = text.replace('tags.addFirst(', 'tags.add(0, ')
    return text
edit('com/sammy/malum/core/systems/rite/SpiritRiteType.java', rite_type)

# These imports are unused in released 1.8.2 SpiritRiteEffect; the class persists via
# RegistryCodecBuddy and NBT on 1.20.
def rite_effect(text):
    text = re.sub(r'^import io\.netty\.buffer\.\*;\n', '', text, flags=re.M)
    text = re.sub(r'^import net\.minecraft\.network\.codec\.\*;\n', '', text, flags=re.M)
    return text
p = root / 'com/sammy/malum/core/systems/rite/effect/SpiritRiteEffect.java'
if p.is_file():
    old = p.read_text(encoding='utf-8')
    new = rite_effect(old)
    if new != old:
        p.write_text(new, encoding='utf-8')

# Lodestone 1.20 block entities expose InteractionResult. Released Malum's old Forge
# owners prove the direct semantic mapping for CONSUME and FAIL as well as SUCCESS/PASS:
# client-side consumption remains CONSUME and rejected inventory actions remain FAIL.
# Convert only BlockEntity source files; ordinary Block.useItemOn implementations are
# deliberately untouched.
changed = 0
unknown = []
for p in (root / 'com/sammy/malum/common/block').rglob('*BlockEntity.java'):
    text = p.read_text(encoding='utf-8')
    if 'ItemInteractionResult' not in text and 'ItemAbilities.AXE_STRIP' not in text:
        continue
    old = text
    for bad in ('ItemInteractionResult.SKIP_DEFAULT_BLOCK_INTERACTION',):
        if bad in text:
            unknown.append(f'{p}: {bad}')
    pattern = re.compile(
        r'public ItemInteractionResult onUseWithItem\(Player\s+(\w+),\s*ItemStack\s+(\w+),\s*InteractionHand\s+(\w+)\)\s*\{'
    )
    def repl(m):
        player, stack, hand = m.groups()
        return (f'public InteractionResult onUse(Player {player}, InteractionHand {hand}) {{\n'
                f'        ItemStack {stack} = {player}.getItemInHand({hand});')
    text = pattern.sub(repl, text)
    text = text.replace('ItemInteractionResult.PASS_TO_DEFAULT_BLOCK_INTERACTION', 'InteractionResult.PASS')
    text = text.replace('ItemInteractionResult.SUCCESS', 'InteractionResult.SUCCESS')
    text = text.replace('ItemInteractionResult.CONSUME', 'InteractionResult.CONSUME')
    text = text.replace('ItemInteractionResult.FAIL', 'InteractionResult.FAIL')
    text = re.sub(r'\bItemInteractionResult\b', 'InteractionResult', text)
    text = re.sub(
        r'super\.onUseWithItem\((\w+),\s*\w+,\s*(\w+)\)',
        r'super.onUse(\1, \2)', text
    )
    text = text.replace('ItemAbilities.AXE_STRIP', 'ToolActions.AXE_STRIP')
    text = text.replace('import net.minecraftforge.common.ItemAbilities;', 'import net.minecraftforge.common.ToolActions;')
    text = text.replace('import net.minecraft.world.ItemInteractionResult;', 'import net.minecraft.world.InteractionResult;')
    if 'InteractionResult' in text and 'import net.minecraft.world.*;' not in text and 'import net.minecraft.world.InteractionResult;' not in text:
        marker = re.search(r'^import net\.minecraft\.world\.[^;]+;\n', text, flags=re.M)
        if marker:
            text = text[:marker.start()] + 'import net.minecraft.world.InteractionResult;\n' + text[marker.start():]
        else:
            pkg_end = text.find('\n', text.find('package ')) + 1
            text = text[:pkg_end] + '\nimport net.minecraft.world.InteractionResult;\n' + text[pkg_end:]
    if 'ToolActions.' in text and 'import net.minecraftforge.common.ToolActions;' not in text:
        pkg_end = text.find('\n', text.find('package ')) + 1
        text = text[:pkg_end] + '\nimport net.minecraftforge.common.ToolActions;\n' + text[pkg_end:]
    if text != old:
        p.write_text(text, encoding='utf-8')
        changed += 1

if unknown:
    raise SystemExit('unmapped ItemInteractionResult semantics:\n' + '\n'.join(unknown))
if changed == 0:
    raise SystemExit('expected at least one BlockEntity interaction bridge')

leftovers = []
for p in root.rglob('*.java'):
    t = p.read_text(encoding='utf-8')
    if p.name.endswith('BlockEntity.java') and 'ItemInteractionResult' in t:
        leftovers.append(f'{p}: ItemInteractionResult')
    if p.name.endswith('BlockEntity.java') and 'ItemAbilities.AXE_STRIP' in t:
        leftovers.append(f'{p}: ItemAbilities.AXE_STRIP')
if leftovers:
    raise SystemExit('core API bridge leftovers:\n' + '\n'.join(leftovers))

print(f'Forge 1.20 core API bridge applied; BlockEntity interaction files changed: {changed}')
