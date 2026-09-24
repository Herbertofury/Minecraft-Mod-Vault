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

# 1.20.1's nested vanilla SellItemFactory/SellEnchantedToolFactory implementations are not public
# under the target mappings, and its buy helper hardcodes a one-emerald payout. Preserve current
# Archers 3.1.1 economics with tiny target-native factories instead of reducing the trade table.
villagers = java_root / "net/archers/village/ArcherVillagers.java"
require_replace(
    villagers,
    "import net.minecraft.item.Items;\n",
    "import net.minecraft.enchantment.EnchantmentHelper;\nimport net.minecraft.item.Item;\nimport net.minecraft.item.ItemStack;\nimport net.minecraft.item.Items;\n",
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
    "    private static TradeOffers.Factory sell(Item item, int emeralds, int count, int maxUses, int experience) {\n"
    "        return sell(item, emeralds, count, maxUses, experience, 0.05F);\n"
    "    }\n\n"
    "    private static TradeOffers.Factory sell(Item item, int emeralds, int count, int maxUses, int experience, float multiplier) {\n"
    "        return (entity, random) -> new TradeOffer(\n"
    "                new ItemStack(Items.EMERALD, emeralds),\n"
    "                new ItemStack(item, count),\n"
    "                maxUses, experience, multiplier);\n"
    "    }\n\n"
    "    private static TradeOffers.Factory sellEnchanted(Item item, int baseEmeralds, int maxUses, int experience, float multiplier) {\n"
    "        return (entity, random) -> {\n"
    "            int level = 5 + random.nextInt(15);\n"
    "            ItemStack enchanted = EnchantmentHelper.enchant(random, new ItemStack(item), level, false);\n"
    "            int emeralds = Math.min(baseEmeralds + level, 64);\n"
    "            return new TradeOffer(new ItemStack(Items.EMERALD, emeralds), enchanted, maxUses, experience, multiplier);\n"
    "        };\n"
    "    }\n\n"
    "    public static void registerVillagers() {\n",
    "ArcherVillagers target-native factories",
)

trade_rewrites = [
    ("new TradeOffers.SellItemFactory(Items.ARROW, 2, 8, 128, 3, 0.01f)", "sell(Items.ARROW, 2, 8, 128, 3, 0.01F)", "arrow sell"),
    ("new TradeOffers.BuyItemFactory(Items.LEATHER, 8, 12, 6, 5)", "buyForEmeralds(Items.LEATHER, 8, 12, 6, 5)", "leather buy"),
    ("new TradeOffers.SellItemFactory(ArcherWeapons.composite_longbow.item(), 6, 1, 16)", "sell(ArcherWeapons.composite_longbow.item(), 6, 1, 1, 16)", "composite longbow sell"),
    ("new TradeOffers.SellItemFactory(ArcherArmors.archerArmorSet_T1.head, 15, 1, 18)", "sell(ArcherArmors.archerArmorSet_T1.head, 15, 1, 1, 18)", "archer hood sell"),
    ("new TradeOffers.BuyItemFactory(Items.STRING, 6, 12, 8, 3)", "buyForEmeralds(Items.STRING, 6, 12, 8, 3)", "string buy"),
    ("new TradeOffers.SellItemFactory(ArcherArmors.archerArmorSet_T1.feet, 15, 1, 18)", "sell(ArcherArmors.archerArmorSet_T1.feet, 15, 1, 1, 18)", "archer boots sell"),
    ("new TradeOffers.BuyItemFactory(Items.REDSTONE, 12, 12, 5, 8)", "buyForEmeralds(Items.REDSTONE, 12, 12, 5, 8)", "redstone buy"),
    ("new TradeOffers.SellItemFactory(ArcherArmors.archerArmorSet_T1.legs, 15, 1, 18)", "sell(ArcherArmors.archerArmorSet_T1.legs, 15, 1, 1, 18)", "archer leggings sell"),
    ("new TradeOffers.SellItemFactory(ArcherArmors.archerArmorSet_T1.chest, 15, 1, 18)", "sell(ArcherArmors.archerArmorSet_T1.chest, 15, 1, 1, 18)", "archer tunic sell"),
    ("new TradeOffers.SellItemFactory(Items.TURTLE_SCUTE, 20, 12, 10)", "sell(Items.SCUTE, 20, 1, 12, 10)", "scute sell"),
    ("(entity, random) -> new TradeOffers.SellEnchantedToolFactory(\n                        ArcherWeapons.royal_longbow.item(), 40, 3, 30, 0F).create(entity, random)", "sellEnchanted(ArcherWeapons.royal_longbow.item(), 40, 3, 30, 0F)", "royal longbow enchanted sell"),
    ("(entity, random) -> new TradeOffers.SellEnchantedToolFactory(\n                        ArcherWeapons.mechanic_shortbow.item(), 40, 3, 30, 0F).create(entity, random)", "sellEnchanted(ArcherWeapons.mechanic_shortbow.item(), 40, 3, 30, 0F)", "mechanic shortbow enchanted sell"),
    ("(entity, random) -> new TradeOffers.SellEnchantedToolFactory(\n                        ArcherWeapons.rapid_crossbow.item(), 40, 3, 30, 0F).create(entity, random)", "sellEnchanted(ArcherWeapons.rapid_crossbow.item(), 40, 3, 30, 0F)", "rapid crossbow enchanted sell"),
    ("(entity, random) -> new TradeOffers.SellEnchantedToolFactory(\n                        ArcherWeapons.heavy_crossbow.item(), 40, 3, 30, 0F).create(entity, random)", "sellEnchanted(ArcherWeapons.heavy_crossbow.item(), 40, 3, 30, 0F)", "heavy crossbow enchanted sell"),
]
for old, new, label in trade_rewrites:
    require_replace(villagers, old, new, f"ArcherVillagers {label}")

print("[Archers API transforms] workbench note instrument: NoteBlockInstrument.BASS -> Instrument.BASS")
print("[Archers API transforms] workbench + auto-fire-hook tooltips: current text on native 1.20.1 callbacks")
print("[Archers API transforms] villager trades: full 3.1.1 buy/sell/enchant economy preserved with target-native factories")
