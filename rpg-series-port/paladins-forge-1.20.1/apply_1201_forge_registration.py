#!/usr/bin/env python3
from __future__ import annotations
import pathlib, sys

root = pathlib.Path(sys.argv[1]).resolve()
if not root.is_dir():
    raise SystemExit(f'missing generated Paladins Java root: {root}')

# Preserve common definition/config/gameplay ownership but route the actual insertion through the
# proven Spell Engine RegistrationBridge whenever Forge has installed a RegisterEvent registrar.
changed=[]
for path in sorted(root.rglob('*.java')):
    text=path.read_text(encoding='utf-8')
    updated=text.replace('Registry.registerReference(', 'net.spell_engine.compat.registry.RegistrationBridge.registerReference(')
    updated=updated.replace('Registry.register(', 'net.spell_engine.compat.registry.RegistrationBridge.register(')
    if updated != text:
        path.write_text(updated,encoding='utf-8'); changed.append(path)
if len(changed) < 4:
    raise SystemExit(f'expected at least four Paladins registry-owner files, changed {len(changed)}')

blocks=root/'net/paladins/block/PaladinBlocks.java'
text=blocks.read_text(encoding='utf-8')
old='''    public static void register() {\n        net.spell_engine.compat.registry.RegistrationBridge.register(Registries.BLOCK, Identifier.of(PaladinsMod.ID, MonkWorkbenchBlock.NAME), MONK_WORKBENCH);\n        net.spell_engine.compat.registry.RegistrationBridge.register(Registries.ITEM, Identifier.of(PaladinsMod.ID, MonkWorkbenchBlock.NAME), MONK_WORKBENCH_BLOCK);\n        // Creative-tab placement of the monk workbench (into the Paladins group) is registered per-platform\n        // from each loader's entrypoint (Fabric: ItemGroupEvents; NeoForge: BuildCreativeModeTabContentsEvent).\n    }\n'''
if old not in text:
    raise SystemExit('PaladinBlocks registration body changed; refusing unsafe BLOCK/ITEM split')
new='''    public static void registerBlocks() {\n        net.spell_engine.compat.registry.RegistrationBridge.register(\n                Registries.BLOCK, Identifier.of(PaladinsMod.ID, MonkWorkbenchBlock.NAME), MONK_WORKBENCH);\n    }\n\n    public static void registerItems() {\n        net.spell_engine.compat.registry.RegistrationBridge.register(\n                Registries.ITEM, Identifier.of(PaladinsMod.ID, MonkWorkbenchBlock.NAME), MONK_WORKBENCH_BLOCK);\n    }\n\n    public static void register() {\n        registerBlocks();\n        registerItems();\n        // Creative-tab placement of the monk workbench (into the Paladins group) is registered per-platform.\n    }\n'''
blocks.write_text(text.replace(old,new,1),encoding='utf-8')

mod=root/'net/paladins/PaladinsMod.java'
text=mod.read_text(encoding='utf-8')
text=text.replace('''    public static void registerBlocks() {\n        PaladinBlocks.register();\n    }\n''','''    public static void registerBlocks() {\n        PaladinBlocks.registerBlocks();\n    }\n\n    public static void registerBlockItems() {\n        PaladinBlocks.registerItems();\n    }\n''',1)
old='''    public static void registerItems() {\n        Group.PALADINS = new ItemGroup.Builder(ItemGroup.Row.TOP, 0)\n                .icon(() -> new ItemStack(Armors.paladinArmorSet_t2.head))\n                .displayName(Text.translatable("itemGroup.paladins.general"))\n                .build();\n        net.spell_engine.compat.registry.RegistrationBridge.register(Registries.ITEM_GROUP, Group.KEY, Group.PALADINS);\n        PaladinBooks.register();\n\n        PaladinWeapons.register(itemConfig.value.weapons);\n        PaladinShields.register(shieldConfig.value.shields);\n        Armors.register(itemConfig.value.armor_sets);\n        shieldConfig.save();\n        itemConfig.save();\n\n        PaladinEntities.register();\n    }\n'''
if old not in text:
    raise SystemExit('PaladinsMod item registration body changed; refusing unsafe phase split')
new='''    public static void registerItemGroup() {\n        Group.PALADINS = new ItemGroup.Builder(ItemGroup.Row.TOP, 0)\n                .icon(() -> new ItemStack(Armors.paladinArmorSet_t2.head))\n                .displayName(Text.translatable("itemGroup.paladins.general"))\n                .build();\n        net.spell_engine.compat.registry.RegistrationBridge.register(Registries.ITEM_GROUP, Group.KEY, Group.PALADINS);\n    }\n\n    public static void registerItems() {\n        PaladinBooks.register();\n        PaladinWeapons.register(itemConfig.value.weapons);\n        PaladinShields.register(shieldConfig.value.shields);\n        Armors.register(itemConfig.value.armor_sets);\n        shieldConfig.save();\n        itemConfig.save();\n    }\n\n    public static void registerEntities() {\n        PaladinEntities.register();\n    }\n'''
mod.write_text(text.replace(old,new,1),encoding='utf-8')

# Every direct vanilla mutation must now be bridge-owned; phase safety is enforced by Forge's registrar.
left=[]
for p in sorted(root.rglob('*.java')):
    for n,line in enumerate(p.read_text(encoding='utf-8').splitlines(),1):
        if 'Registry.register(' in line or 'Registry.registerReference(' in line:
            left.append(f'{p.relative_to(root)}:{n}:{line.strip()}')
if left:
    raise SystemExit('unbridged vanilla registry mutations survived:\n'+'\n'.join(left))

print(f'[Paladins Forge registration] bridged {len(changed)} registry-owner source files')
print('[Paladins Forge registration] split BLOCK/BLOCK_ITEM, ITEM_GROUP/ITEM, and ENTITY_TYPE lifecycle ownership')
