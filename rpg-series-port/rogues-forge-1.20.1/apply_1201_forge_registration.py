#!/usr/bin/env python3
import pathlib,sys
root=pathlib.Path(sys.argv[1]).resolve()
changed=[]
for p in sorted(root.rglob('*.java')):
    t=p.read_text(encoding='utf-8'); u=t.replace('Registry.registerReference(', 'net.spell_engine.compat.registry.RegistrationBridge.registerReference(').replace('Registry.register(', 'net.spell_engine.compat.registry.RegistrationBridge.register(')
    if u!=t: p.write_text(u,encoding='utf-8'); changed.append(p)
if len(changed)<4: raise SystemExit(f'expected >=4 registry owners, changed {len(changed)}')
blocks=root/'net/rogues/block/CustomBlocks.java'; t=blocks.read_text(encoding='utf-8')
old='''    public static void register() {\n        for (var entry : all) {\n            net.spell_engine.compat.registry.RegistrationBridge.register(Registries.BLOCK, Identifier.of(RoguesMod.NAMESPACE, entry.name), entry.block);\n            net.spell_engine.compat.registry.RegistrationBridge.register(Registries.ITEM, Identifier.of(RoguesMod.NAMESPACE, entry.name), entry.item());\n        }\n        // Creative-tab placement (into the Rogues group) is registered per-platform from each loader's\n        // entrypoint, iterating CustomBlocks.all — no Fabric API ItemGroupEvents in common.\n    }\n'''
if old not in t: raise SystemExit('CustomBlocks current registration shape changed')
new='''    public static void registerBlocks() { for (var entry : all) net.spell_engine.compat.registry.RegistrationBridge.register(Registries.BLOCK, Identifier.of(RoguesMod.NAMESPACE, entry.name), entry.block); }\n    public static void registerItems() { for (var entry : all) net.spell_engine.compat.registry.RegistrationBridge.register(Registries.ITEM, Identifier.of(RoguesMod.NAMESPACE, entry.name), entry.item()); }\n    public static void register() { registerBlocks(); registerItems(); }\n'''
blocks.write_text(t.replace(old,new,1),encoding='utf-8')
mod=root/'net/rogues/RoguesMod.java'; t=mod.read_text(encoding='utf-8')
old='''    public static void registerItems() {\n        Group.ROGUES = new ItemGroup.Builder(ItemGroup.Row.TOP, 0)\n                .icon(() -> new ItemStack(RogueArmors.RogueArmorSet_t2.head))\n                .displayName(Text.translatable("itemGroup." + NAMESPACE + ".general"))\n                .build();\n        CustomBlocks.register();\n        net.spell_engine.compat.registry.RegistrationBridge.register(Registries.ITEM_GROUP, Group.KEY, Group.ROGUES);\n        RogueWeapons.register(itemConfig.value.weapons);\n        RogueArmors.register(itemConfig.value.armor_sets);\n        itemConfig.save();\n    }\n'''
if old not in t: raise SystemExit('RoguesMod registerItems current shape changed')
new='''    public static void registerBlocks() { CustomBlocks.registerBlocks(); }\n    public static void registerItemGroup() {\n        Group.ROGUES = new ItemGroup.Builder(ItemGroup.Row.TOP, 0).icon(() -> new ItemStack(RogueArmors.RogueArmorSet_t2.head)).displayName(Text.translatable("itemGroup." + NAMESPACE + ".general")).build();\n        net.spell_engine.compat.registry.RegistrationBridge.register(Registries.ITEM_GROUP, Group.KEY, Group.ROGUES);\n    }\n    public static void registerItems() { CustomBlocks.registerItems(); RogueWeapons.register(itemConfig.value.weapons); RogueArmors.register(itemConfig.value.armor_sets); itemConfig.save(); }\n'''
mod.write_text(t.replace(old,new,1),encoding='utf-8')
left=[]
for p in sorted(root.rglob('*.java')):
    for n,line in enumerate(p.read_text(encoding='utf-8').splitlines(),1):
        if 'Registry.register(' in line or 'Registry.registerReference(' in line: left.append(f'{p.relative_to(root)}:{n}:{line.strip()}')
if left: raise SystemExit('unbridged registry calls:\n'+'\n'.join(left))
print(f'[Rogues Forge registration] bridged {len(changed)} registry owners and split block/item-group/item phases')
