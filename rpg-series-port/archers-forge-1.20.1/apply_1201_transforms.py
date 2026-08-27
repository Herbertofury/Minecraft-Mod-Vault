#!/usr/bin/env python3
from __future__ import annotations

import pathlib
import sys

root = pathlib.Path(sys.argv[1]).resolve()
if not root.is_dir():
    raise SystemExit(f"missing generated Java root: {root}")

# Mechanical mapping seam valid across current source: Identifier.of(namespace, path) was
# introduced after 1.20.1; the two-String constructor is the exact 1.20.1 equivalent.
identifier_rewrites = 0
for path in sorted(root.rglob("*.java")):
    text = path.read_text(encoding="utf-8")
    updated = text.replace("Identifier.of(", "new Identifier(")
    if updated != text:
        identifier_rewrites += text.count("Identifier.of(")
        path.write_text(updated, encoding="utf-8")

quivers = root / "net/archers/item/Quivers.java"
if not quivers.is_file():
    raise SystemExit(f"current Archers Quivers.java missing: {quivers}")

# Quivers are current 3.1.1 feature authority but their 1.21 data components have no 1.20.1
# equivalent. Translate capacity into the Bundle API compatibility constructor and preserve the
# gray hint via appendTooltip on an Archers-owned subclass. This is intentionally explicit rather
# than a broad data-component regex so compiler failures reveal every other 1.21 seam separately.
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
        for (var entry : entries) {
            Registry.register(Registries.ITEM, entry.id(), entry.item());
        }
    }
}
''', encoding="utf-8")

print(f"[Archers transforms] Identifier.of rewrites: {identifier_rewrites}")
print("[Archers transforms] Quivers.java: current 3.1.1 data-component capacity/lore translated to native 1.20.1 APIs")
