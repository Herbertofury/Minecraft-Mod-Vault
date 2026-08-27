#!/usr/bin/env python3
from __future__ import annotations

import pathlib
import sys

java_root = pathlib.Path(sys.argv[1]).resolve()
if not java_root.is_dir():
    raise SystemExit(f"missing generated Java root: {java_root}")


def require_replace(path: pathlib.Path, old: str, new: str, label: str) -> None:
    if not path.is_file():
        raise SystemExit(f"{label}: missing source: {path}")
    text = path.read_text(encoding="utf-8")
    if old not in text:
        raise SystemExit(f"{label}: expected current-upstream seam not found")
    path.write_text(text.replace(old, new), encoding="utf-8")


# Exact historical Archers 1.20.1 mapping substrate uses Instrument.BASS; current 3.1.1 uses
# the post-1.20.1 NoteBlockInstrument name. Preserve the same BASS instrument behavior.
blocks = java_root / "net/archers/block/ArcherBlocks.java"
require_replace(
    blocks,
    "import net.minecraft.block.enums.NoteBlockInstrument;",
    "import net.minecraft.block.enums.Instrument;",
    "ArcherBlocks instrument import",
)
require_replace(
    blocks,
    ".instrument(NoteBlockInstrument.BASS)",
    ".instrument(Instrument.BASS)",
    "ArcherBlocks workbench instrument",
)

# Current 3.1.1 tooltip content is unchanged; only the Item/Block tooltip callback API changed.
workbench = java_root / "net/archers/block/ArcherWorkbenchBlock.java"
require_replace(
    workbench,
    "import net.minecraft.item.Item;\n",
    "import net.minecraft.client.item.TooltipContext;\n",
    "ArcherWorkbench tooltip import",
)
require_replace(
    workbench,
    "import net.minecraft.item.tooltip.TooltipType;\n",
    "",
    "ArcherWorkbench TooltipType removal",
)
require_replace(
    workbench,
    "public void appendTooltip(ItemStack stack, Item.TooltipContext context, List<Text> tooltip, TooltipType options) {\n        super.appendTooltip(stack, context, tooltip, options);",
    "public void appendTooltip(ItemStack stack, @Nullable BlockView world, List<Text> tooltip, TooltipContext options) {\n        super.appendTooltip(stack, world, tooltip, options);",
    "ArcherWorkbench 1.20.1 tooltip callback",
)

auto_fire_item = java_root / "net/archers/item/misc/AutoFireHookItem.java"
require_replace(
    auto_fire_item,
    "import net.minecraft.item.tooltip.TooltipType;\n",
    "import net.minecraft.client.item.TooltipContext;\nimport net.minecraft.world.World;\nimport org.jetbrains.annotations.Nullable;\n",
    "AutoFireHookItem tooltip imports",
)
require_replace(
    auto_fire_item,
    "public void appendTooltip(ItemStack stack, TooltipContext context, List<Text> tooltip, TooltipType type) {\n        super.appendTooltip(stack, context, tooltip, type);",
    "public void appendTooltip(ItemStack stack, @Nullable World world, List<Text> tooltip, TooltipContext context) {\n        super.appendTooltip(stack, world, tooltip, context);",
    "AutoFireHookItem 1.20.1 tooltip callback",
)

# 1.21 BuyItemFactory gained a configurable emerald payout. 1.20.1's built-in
# BuyForOneEmeraldFactory cannot represent current Archers' 5/3/8 emerald payouts, so preserve
# the exact current economy with an ordinary target-native TradeOffer factory. SellItemFactory's
# 1.20.1 ItemStack overload retains the explicit 0.01 arrow price multiplier.
villagers = java_root / "net/archers/village/ArcherVillagers.java"
require_replace(
    villagers,
    "import net.minecraft.item.Items;\n",
    "import net.minecraft.item.Item;\nimport net.minecraft.item.ItemStack;\nimport net.minecraft.item.Items;\n",
    "ArcherVillagers trade imports",
)
require_replace(
    villagers,
    "import net.minecraft.village.TradeOffers;\n",
    "import net.minecraft.village.TradeOffer;\nimport net.minecraft.village.TradeOffers;\n",
    "ArcherVillagers TradeOffer import",
)
require_replace(
    villagers,
    "    public static void registerVillagers() {\n",
    "    private static TradeOffers.Factory buyForEmeralds(Item item, int count, int maxUses, int experience, int emeralds) {\n"
    "        return (entity, random) -> new TradeOffer(\n"
    "                new ItemStack(item, count),\n"
    "                new ItemStack(Items.EMERALD, emeralds),\n"
    "                maxUses, experience, 0.05F);\n"
    "    }\n\n"
    "    public static void registerVillagers() {\n",
    "ArcherVillagers target-native buy factory",
)
require_replace(
    villagers,
    "new TradeOffers.SellItemFactory(Items.ARROW, 2, 8, 128, 3, 0.01f)",
    "new TradeOffers.SellItemFactory(new ItemStack(Items.ARROW), 2, 8, 128, 3, 0.01f)",
    "ArcherVillagers arrow sell multiplier",
)
require_replace(
    villagers,
    "new TradeOffers.BuyItemFactory(Items.LEATHER, 8, 12, 6, 5)",
    "buyForEmeralds(Items.LEATHER, 8, 12, 6, 5)",
    "ArcherVillagers leather buy",
)
require_replace(
    villagers,
    "new TradeOffers.BuyItemFactory(Items.STRING, 6, 12, 8, 3)",
    "buyForEmeralds(Items.STRING, 6, 12, 8, 3)",
    "ArcherVillagers string buy",
)
require_replace(
    villagers,
    "new TradeOffers.BuyItemFactory(Items.REDSTONE, 12, 12, 5, 8)",
    "buyForEmeralds(Items.REDSTONE, 12, 12, 5, 8)",
    "ArcherVillagers redstone buy",
)

print("[Archers API transforms] workbench note instrument: NoteBlockInstrument.BASS -> Instrument.BASS")
print("[Archers API transforms] workbench + auto-fire-hook tooltips: current text on native 1.20.1 callbacks")
print("[Archers API transforms] villager trades: current counts/maxUses/xp/emerald payouts preserved on 1.20.1")
