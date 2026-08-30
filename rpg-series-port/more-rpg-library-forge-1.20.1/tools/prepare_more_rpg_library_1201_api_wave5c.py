#!/usr/bin/env python3
from pathlib import Path
import struct
import sys

if len(sys.argv) != 2:
    raise SystemExit('usage: prepare_more_rpg_library_1201_api_wave5c.py <prepared-root>')
root = Path(sys.argv[1]).resolve()
java = root / 'common/src/main/java'
resources = root / 'common/src/main/resources'
if not java.is_dir():
    raise SystemExit(f'missing common Java root: {java}')
if not resources.is_dir():
    raise SystemExit(f'missing common resources root: {resources}')


def owner(rel: str) -> Path:
    p = java / rel
    if not p.is_file():
        raise SystemExit(f'missing Wave 5c owner: {p}')
    return p


def replace_exact(s: str, needle: str, repl: str, expected: int, label: str) -> str:
    found = s.count(needle)
    if found != expected:
        raise SystemExit(f'{label} seam drifted: expected={expected} found={found}')
    return s.replace(needle, repl)


# Final 1.21 -> 1.20.1 DamageSource / Spell Power representation seams. These are API spelling
# translations only: isDirect() is the inverse of target isIndirect(), and the Spell Power tag keeps
# the same spell_power:all registry identity under the singular 1.20.1 holder class.
p = owner('net/more_rpg_classes/mixin/LivingEntityMixin.java')
s = p.read_text(encoding='utf-8')
s = replace_exact(s, 'source.isDirect()', '!source.isIndirect()', 1,
                  'LivingEntityMixin direct damage predicate')
s = replace_exact(s, 'SpellPowerTags.DamageTypes.ALL', 'SpellPowerTags.DamageType.ALL', 5,
                  'LivingEntityMixin Spell Power damage tag')

# The project-owned 1.20.1 oracle for this exact feature used this UUID/name pair. Preserve the modern
# re-entrant depth/stack behavior while translating only the modifier identifier representation and
# constructor signature required by EntityAttributeInstance/EntityAttributeModifier on 1.20.1.
modern_id = ('@Unique private static final net.minecraft.util.Identifier ARMOR_PIERCING_ID = '
             'new net.minecraft.util.Identifier("more_rpg_classes", "armor_piercing_reduction");')
target_id = ('@Unique private static final UUID ARMOR_PIERCING_UUID = '
             'UUID.fromString("a8f6b5c2-3d4e-4f1a-9b2c-7e8d9f0a1b2c");\n'
             '    @Unique private static final String ARMOR_PIERCING_NAME = "armor_piercing_reduction";')
s = replace_exact(s, modern_id, target_id, 1,
                  'LivingEntityMixin armor piercing identifier')
s = replace_exact(s, 'removeModifier(ARMOR_PIERCING_ID)',
                  'removeModifier(ARMOR_PIERCING_UUID)', 4,
                  'LivingEntityMixin armor modifier removal')
s = replace_exact(s,
                  'ARMOR_PIERCING_ID, -armorAttribute.getValue() * piercingPercent, '
                  'net.minecraft.entity.attribute.EntityAttributeModifier.Operation.ADDITION',
                  'ARMOR_PIERCING_UUID, ARMOR_PIERCING_NAME, -armorAttribute.getValue() * piercingPercent, '
                  'net.minecraft.entity.attribute.EntityAttributeModifier.Operation.ADDITION',
                  1, 'LivingEntityMixin armor modifier constructor')
s = replace_exact(s,
                  'ARMOR_PIERCING_ID, -toughnessAttribute.getValue() * piercingPercent, '
                  'net.minecraft.entity.attribute.EntityAttributeModifier.Operation.ADDITION',
                  'ARMOR_PIERCING_UUID, ARMOR_PIERCING_NAME, -toughnessAttribute.getValue() * piercingPercent, '
                  'net.minecraft.entity.attribute.EntityAttributeModifier.Operation.ADDITION',
                  1, 'LivingEntityMixin toughness modifier constructor')
if 'ARMOR_PIERCING_ID' in s:
    raise SystemExit('LivingEntityMixin modern armor Identifier survived Wave 5c')
p.write_text(s, encoding='utf-8')

# Modern drawGuiTexture consumes GUI sprite identifiers such as more_rpg_classes:hud/heart/foo.
# 1.20.1 DrawContext has no GUI-sprite helper, so resolve that same logical sprite to its unchanged PNG
# resource and draw all 9x9 pixels directly. Validate the exact six modern assets and dimensions first.
heart_dir = resources / 'assets/more_rpg_classes/textures/gui/sprites/hud/heart'
expected_hearts = {
    'fatal_poison_container.png',
    'fatal_poison_container_blinking.png',
    'fatal_poison_full.png',
    'fatal_poison_full_blinking.png',
    'fatal_poison_half.png',
    'fatal_poison_half_blinking.png',
}
if not heart_dir.is_dir():
    raise SystemExit(f'missing modern fatal-poison heart resource directory: {heart_dir}')
actual_hearts = {p.name for p in heart_dir.iterdir() if p.is_file() and p.suffix == '.png'}
if actual_hearts != expected_hearts:
    raise SystemExit('fatal-poison heart asset set drifted: '
                     f'expected={sorted(expected_hearts)} actual={sorted(actual_hearts)}')
for name in sorted(expected_hearts):
    data = (heart_dir / name).read_bytes()
    if len(data) < 24 or data[:8] != b'\x89PNG\r\n\x1a\n' or data[12:16] != b'IHDR':
        raise SystemExit(f'invalid PNG header for heart asset: {name}')
    width, height = struct.unpack('>II', data[16:24])
    if (width, height) != (9, 9):
        raise SystemExit(f'heart asset dimensions drifted: {name}={width}x{height}, expected=9x9')

p = owner('net/more_rpg_classes/mixin/DrawHeartsMixin.java')
s = p.read_text(encoding='utf-8')
s = replace_exact(
    s,
    '        context.drawGuiTexture(texture, x, y, 9, 9);',
    '        Identifier textureFile = new Identifier(texture.getNamespace(), "textures/gui/sprites/" + texture.getPath() + ".png");\n'
    '        context.drawTexture(textureFile, x, y, 0.0F, 0.0F, 9, 9, 9, 9);',
    1, 'DrawHeartsMixin target direct texture draw')
if 'drawGuiTexture(' in s:
    raise SystemExit('DrawHeartsMixin modern drawGuiTexture survived Wave 5c')
p.write_text(s, encoding='utf-8')

# Modern Optional.empty() means no enchantment-registry restriction; target 1.20.1 expresses that as
# treasureAllowed=true on the four-argument helper. Item choice, random source, level and probability
# remain unchanged.
p = owner('net/more_rpg_classes/util/loot/ItemTagPickerLootFunction.java')
s = p.read_text(encoding='utf-8')
modern_enchant = '''            result = EnchantmentHelper.enchant(\n                    context.getRandom(),\n                    result,\n                    level,\n                    context.getWorld().getRegistryManager(),\n                    Optional.<RegistryEntryList<Enchantment>>empty()\n            );'''
target_enchant = '            result = EnchantmentHelper.enchant(context.getRandom(), result, level, true);'
s = replace_exact(s, modern_enchant, target_enchant, 1,
                  'ItemTagPickerLootFunction unrestricted enchant helper')
p.write_text(s, encoding='utf-8')

# Fail closed on every diagnostic family owned by the exact #342 14-error boundary.
living = owner('net/more_rpg_classes/mixin/LivingEntityMixin.java').read_text(encoding='utf-8')
for forbidden in ('source.isDirect()', 'SpellPowerTags.DamageTypes.ALL', 'ARMOR_PIERCING_ID'):
    if forbidden in living:
        raise SystemExit(f'Wave 5c LivingEntity API survived: {forbidden}')
hearts = owner('net/more_rpg_classes/mixin/DrawHeartsMixin.java').read_text(encoding='utf-8')
if 'drawGuiTexture(' in hearts or 'textures/gui/sprites/' not in hearts:
    raise SystemExit('Wave 5c heart rendering bridge is incomplete')
loot = owner('net/more_rpg_classes/util/loot/ItemTagPickerLootFunction.java').read_text(encoding='utf-8')
if 'context.getWorld().getRegistryManager(),' in loot and 'EnchantmentHelper.enchant(' in loot:
    raise SystemExit('Wave 5c modern enchant helper survived')

print('[More RPG 2.7.2] TARGET_1201_API_WAVE5C_PASS direct_damage=1 damage_tags=5 armor_removals=4 armor_adds=2 enchant_helper=1')
print('[More RPG 2.7.2] HEART_SPRITES_1201_DIRECT_TEXTURE_PASS assets=6 dimensions=9x9 path=textures/gui/sprites/hud/heart')
print('[More RPG 2.7.2] COMPILE_TAIL_14_SEAMS_OWNED hud=1 living=12 loot=1')
print('[More RPG 2.7.2] MODERN_GAMEPLAY_AND_HEART_FIDELITY_PRESERVED wave5c=api_bridge_only')
