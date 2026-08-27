#!/usr/bin/env python3
from __future__ import annotations

import json
import pathlib
import sys

java_root = pathlib.Path(sys.argv[1]).resolve()
resources_root = pathlib.Path(sys.argv[2]).resolve()
if not java_root.is_dir():
    raise SystemExit(f"missing generated Java root: {java_root}")
if not resources_root.is_dir():
    raise SystemExit(f"missing generated resources root: {resources_root}")

identifier_rewrites = 0
for path in sorted(java_root.rglob("*.java")):
    text = path.read_text(encoding="utf-8")
    updated = text.replace("Identifier.of(", "new Identifier(")
    if updated != text:
        identifier_rewrites += text.count("Identifier.of(")
        path.write_text(updated, encoding="utf-8")

quivers = java_root / "net/archers/item/Quivers.java"
if not quivers.is_file():
    raise SystemExit(f"current Archers Quivers.java missing: {quivers}")
quivers.write_text('''package net.archers.item;

import com.github.theredbrain.bundleapi.item.CustomBundleItem;
import net.archers.ArchersMod;
import net.minecraft.client.item.TooltipContext;
import net.minecraft.item.Item;
import net.minecraft.item.ItemStack;
import net.minecraft.registry.Registries;
import net.minecraft.registry.Registry;
import net.minecraft.registry.tag.ItemTags;
import net.minecraft.registry.tag.TagKey;
import net.minecraft.text.Text;
import net.minecraft.util.Formatting;
import net.minecraft.util.Identifier;
import net.minecraft.util.Rarity;
import net.minecraft.world.World;
import org.jetbrains.annotations.Nullable;

import java.util.ArrayList;
import java.util.List;
import java.util.function.Function;

public class Quivers {
    public static final List<Entry> entries = new ArrayList<>();
    public record Entry(Identifier id, int capacity, Item item) { }
    public record Args(TagKey<Item> tag, int capacity, Item.Settings settings) { }

    public static class QuiverItem extends CustomBundleItem {
        public QuiverItem(@Nullable TagKey<Item> tag, int capacity, Settings settings) {
            super(tag, capacity, settings);
        }

        @Override
        public void appendTooltip(ItemStack stack, @Nullable World world, List<Text> tooltip, TooltipContext context) {
            tooltip.add(Text.translatable("item.archers.quiver.hint").formatted(Formatting.GRAY));
            super.appendTooltip(stack, world, tooltip, context);
        }
    }

    public static Function<Args, Item> factory = args -> new QuiverItem(args.tag(), args.capacity(), args.settings());

    public static Entry entry(String name, int capacity, @Nullable Rarity rarity) {
        var settings = new Item.Settings().maxCount(1);
        if (rarity != null) settings.rarity(rarity);
        var bundle = factory.apply(new Args(ItemTags.ARROWS, capacity, settings));
        var id = new Identifier(ArchersMod.ID, name);
        var entry = new Entry(id, capacity, bundle);
        entries.add(entry);
        return entry;
    }

    public static void register() {
        entry("small_quiver", 4, null);
        entry("medium_quiver", 8, null);
        entry("large_quiver", 12, Rarity.UNCOMMON);
        for (var entry : entries) Registry.register(Registries.ITEM, entry.id(), entry.item());
    }
}
''', encoding="utf-8")

# 1.21 ComponentType<Boolean> AUTO_FIRE -> 1.20.1 stack NBT boolean. The feature semantics are
# identical (persistent per-stack true/false state); only the storage API changes.
auto_fire = java_root / "net/archers/item/misc/AutoFireHook.java"
if not auto_fire.is_file():
    raise SystemExit(f"current AutoFireHook.java missing: {auto_fire}")
auto_fire.write_text('''package net.archers.item.misc;

import net.archers.ArchersMod;
import net.minecraft.item.CrossbowItem;
import net.minecraft.item.Item;
import net.minecraft.item.ItemStack;
import net.minecraft.registry.RegistryKeys;
import net.minecraft.registry.tag.TagKey;
import net.minecraft.util.Identifier;

public class AutoFireHook {
    private static final String NBT_KEY = "ArchersAutoFire";
    public static final Identifier id = new Identifier(ArchersMod.ID, "auto_fire_hook");
    public static final Item item = new AutoFireHookItem((new Item.Settings()).maxCount(1));
    public static final TagKey<Item> AFH_ATTACHABLE = TagKey.of(RegistryKeys.ITEM, new Identifier(ArchersMod.ID, "auto_fire_hook_attachables"));

    public static boolean isApplied(ItemStack itemStack) {
        return itemStack != null && !itemStack.isEmpty() && itemStack.hasNbt() && itemStack.getNbt().getBoolean(NBT_KEY);
    }

    public static void apply(ItemStack itemStack) {
        if (itemStack != null && !itemStack.isEmpty()) itemStack.getOrCreateNbt().putBoolean(NBT_KEY, true);
    }

    public static void remove(ItemStack itemStack) {
        if (itemStack == null || !itemStack.hasNbt()) return;
        itemStack.getNbt().remove(NBT_KEY);
        if (itemStack.getNbt().isEmpty()) itemStack.setNbt(null);
    }

    public static boolean isApplicable(ItemStack itemStack) {
        if (itemStack == null || itemStack.isEmpty()) return false;
        return (itemStack.getItem() instanceof CrossbowItem || itemStack.isIn(AFH_ATTACHABLE)) && !isApplied(itemStack);
    }
}
''', encoding="utf-8")

for obsolete in [
    java_root / "net/archers/component/ArcherComponents.java",
    java_root / "net/archers/mixin/component/DataComponentTypesMixin.java",
]:
    if obsolete.exists(): obsolete.unlink()

mixins = resources_root / "archers.mixins.json"
if not mixins.is_file():
    raise SystemExit(f"current archers.mixins.json missing: {mixins}")
data = json.loads(mixins.read_text(encoding="utf-8"))
data["mixins"] = [m for m in data.get("mixins", []) if m != "component.DataComponentTypesMixin"]
mixins.write_text(json.dumps(data, indent=2) + "\n", encoding="utf-8")

print(f"[Archers transforms] Identifier.of rewrites: {identifier_rewrites}")
print("[Archers transforms] Quivers: 1.21 components -> Bundle API 1.20.1 capacity + tooltip override")
print("[Archers transforms] Auto Fire: 1.21 boolean component -> persistent 1.20.1 ItemStack NBT; bootstrap mixin removed")
