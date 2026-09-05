#!/usr/bin/env python3
from __future__ import annotations

import pathlib
import sys

java_root = pathlib.Path(sys.argv[1]).resolve()
if not java_root.is_dir():
    raise SystemExit(f"missing generated Archers Java root: {java_root}")

# Route every Archers-owned vanilla registration call through Spell Engine's loader-neutral seam.
# Outside an installed Forge event scope it behaves exactly like Registry.register; inside Forge 47 it
# delegates to RegisterEvent.RegisterHelper and rejects accidental cross-registry phase drift.
changed = 0
for path in sorted(java_root.rglob("*.java")):
    text = path.read_text(encoding="utf-8")
    updated = text.replace("Registry.registerReference(", "net.spell_engine.compat.registry.RegistrationBridge.registerReference(")
    updated = updated.replace("Registry.register(", "net.spell_engine.compat.registry.RegistrationBridge.register(")
    if updated != text:
        changed += 1
        path.write_text(updated, encoding="utf-8")

if changed < 4:
    raise SystemExit(f"expected multiple Archers registry owners to be bridged, changed only {changed} source files")

# The upstream block registrar intentionally registers both its Block and BlockItem together. Forge 47
# emits separate BLOCK and ITEM events, so expose phase-specific methods while preserving the original
# combined register() fallback for non-Forge/common callers.
blocks = java_root / "net/archers/block/ArcherBlocks.java"
text = blocks.read_text(encoding="utf-8")
old = '''    public static void register() {
        for (var entry : all) {
            net.spell_engine.compat.registry.RegistrationBridge.register(Registries.BLOCK, new Identifier(ArchersMod.ID, entry.name), entry.block);
            net.spell_engine.compat.registry.RegistrationBridge.register(Registries.ITEM, new Identifier(ArchersMod.ID, entry.name), entry.item());
        }
        // Creative-tab placement (into the Archers group) is registered per-platform from each loader's
        // entrypoint, iterating ArcherBlocks.all — no Fabric API ItemGroupEvents in common.
    }
'''
if old not in text:
    raise SystemExit("ArcherBlocks registration body changed; refusing unsafe phase split")
new = '''    public static void registerBlocks() {
        for (var entry : all) {
            net.spell_engine.compat.registry.RegistrationBridge.register(
                    Registries.BLOCK, new Identifier(ArchersMod.ID, entry.name), entry.block);
        }
    }

    public static void registerItems() {
        for (var entry : all) {
            net.spell_engine.compat.registry.RegistrationBridge.register(
                    Registries.ITEM, new Identifier(ArchersMod.ID, entry.name), entry.item());
        }
    }

    public static void register() {
        registerBlocks();
        registerItems();
        // Creative-tab placement (into the Archers group) is registered per-platform from each loader's
        // entrypoint, iterating ArcherBlocks.all — no Fabric API ItemGroupEvents in common.
    }
'''
blocks.write_text(text.replace(old, new, 1), encoding="utf-8")

# The current Archers common entrypoint builds/registers its creative tab inside registerItems(). Split
# only the registry insertion; construction, icon/name, equipment registration and config-save semantics
# remain byte-for-byte equivalent in intent.
mod = java_root / "net/archers/ArchersMod.java"
text = mod.read_text(encoding="utf-8")
old = '''    public static void registerBlocks() {
        ArcherBlocks.register();
    }

    public static void registerItems() {
        Group.ARCHERS = new ItemGroup.Builder(ItemGroup.Row.TOP, 0)
                .icon(() -> new ItemStack(ArcherArmors.archerArmorSet_T2.head))
                .displayName(Text.translatable("itemGroup." + ID + ".general"))
                .build();
        net.spell_engine.compat.registry.RegistrationBridge.register(Registries.ITEM_GROUP, Group.KEY, Group.ARCHERS);
        Misc.register();
        ArcherWeapons.register(itemConfig.value.ranged_weapons, itemConfig.value.melee_weapons);
        ArcherArmors.register(itemConfig.value.armor_sets);
        itemConfig.save();
    }
'''
if old not in text:
    raise SystemExit("ArchersMod registration body changed; refusing unsafe item-group split")
new = '''    public static void registerBlocks() {
        ArcherBlocks.registerBlocks();
    }

    public static void registerItemGroup() {
        Group.ARCHERS = new ItemGroup.Builder(ItemGroup.Row.TOP, 0)
                .icon(() -> new ItemStack(ArcherArmors.archerArmorSet_T2.head))
                .displayName(Text.translatable("itemGroup." + ID + ".general"))
                .build();
        net.spell_engine.compat.registry.RegistrationBridge.register(Registries.ITEM_GROUP, Group.KEY, Group.ARCHERS);
    }

    public static void registerItems() {
        Misc.register();
        ArcherWeapons.register(itemConfig.value.ranged_weapons, itemConfig.value.melee_weapons);
        ArcherArmors.register(itemConfig.value.armor_sets);
        itemConfig.save();
    }
'''
mod.write_text(text.replace(old, new, 1), encoding="utf-8")

print(f"[Archers Forge registration] bridged {changed} common registry-owner source files")
print("[Archers Forge registration] split BLOCK/ITEM workbench registration and ITEM_GROUP/ITEM lifecycle")
