#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 2:
    raise SystemExit("usage: compat_pass_2.py <generated-port-root>")

root = Path(sys.argv[1]).resolve()
common = root / "common/src/main/java"

# Common code is compiled in Yarn named mappings. Never expose it to the remapped Forge/Mojmap
# distributable JARs: use the separately-built common/named foundation JARs for compilation, while
# :forge continues to consume the real Forge JARs at loader/runtime level.
common_build = root / "common/build.gradle"
build = common_build.read_text()
for old, new in (
    ("libs/structure-pool-api.jar", "libs/structure-pool-api-common.jar"),
    ("libs/spell-power.jar", "libs/spell-power-common.jar"),
    ("libs/ranged-weapon-api.jar", "libs/ranged-weapon-api-common.jar"),
):
    build = build.replace(old, new)
common_build.write_text(build)

# 1.21 exposes BuyItemFactory/SellItemFactory publicly. In 1.20.1 those implementation helpers are
# unavailable/private, but the stable TradeOffers.Factory contract and TradeOffer constructor already
# provide exactly the same behavior. Preserve the CURRENT 2.4.0 counts/prices/max-uses/xp rather than
# resurrecting Jewelry 1.3.7's older economy.
villagers = common / "net/jewelry/village/JewelryVillagers.java"
villagers.write_text(r'''package net.jewelry.village;

import com.google.common.collect.ImmutableSet;
import net.jewelry.JewelryMod;
import net.jewelry.blocks.JewelryBlocks;
import net.jewelry.items.JewelryItems;
import net.jewelry.util.SoundHelper;
import net.minecraft.block.BlockState;
import net.minecraft.item.Item;
import net.minecraft.item.ItemStack;
import net.minecraft.item.Items;
import net.minecraft.registry.Registries;
import net.minecraft.registry.Registry;
import net.minecraft.registry.RegistryKey;
import net.minecraft.text.Text;
import net.minecraft.util.Identifier;
import net.minecraft.village.TradeOffer;
import net.minecraft.village.TradeOffers;
import net.minecraft.village.VillagerProfession;
import net.minecraft.world.poi.PointOfInterestType;

import java.util.LinkedHashMap;
import java.util.List;
import java.util.Set;

public class JewelryVillagers {
    public static final String JEWELER = "jeweler";

    public static VillagerProfession JEWELER_PROFESSION;
    public static Identifier POI_ID = new Identifier(JewelryMod.ID, JEWELER);
    public static final int POI_TICKET_COUNT = 1;
    public static final int POI_SEARCH_DISTANCE = 10;

    public static Set<BlockState> poiBlockStates() {
        return ImmutableSet.copyOf(JewelryBlocks.JEWELERS_KIT.block().getStateManager().getStates());
    }

    public static VillagerProfession createProfession(String name, RegistryKey<PointOfInterestType> workStation) {
        var id = new Identifier(JewelryMod.ID, name);
        return new VillagerProfession(
                id.toString(),
                entry -> entry.matchesKey(workStation),
                entry -> entry.matchesKey(workStation),
                ImmutableSet.of(),
                ImmutableSet.of(),
                SoundHelper.JEWELRY_WORKBENCH
        );
    }

    public static void registerVillagers() {
        var workStation = RegistryKey.of(Registries.POINT_OF_INTEREST_TYPE.getKey(), POI_ID);
        JEWELER_PROFESSION = Registry.register(Registries.VILLAGER_PROFESSION,
                new Identifier(JewelryMod.ID, JEWELER), createProfession(JEWELER, workStation));
    }

    // Mirrors the 1.21 BuyItemFactory(ItemConvertible, count, maxUses, experience, price).
    private static TradeOffers.Factory buy(Item item, int count, int maxUses, int experience, int price) {
        return (entity, random) -> new TradeOffer(
                new ItemStack(item, count),
                new ItemStack(Items.EMERALD, price),
                maxUses, experience, 0.05F);
    }

    // Mirrors the 1.21 SellItemFactory(Item, price, count, maxUses, experience).
    private static TradeOffers.Factory sell(Item item, int price, int count, int maxUses, int experience) {
        return (entity, random) -> new TradeOffer(
                new ItemStack(Items.EMERALD, price),
                new ItemStack(item, count),
                maxUses, experience, 0.05F);
    }

    public static LinkedHashMap<Integer, List<TradeOffers.Factory>> createTrades() {
        LinkedHashMap<Integer, List<TradeOffers.Factory>> trades = new LinkedHashMap<>();

        trades.put(1, List.of(
                buy(Items.COPPER_INGOT, 8, 8, 3, 2),
                buy(Items.STRING, 7, 6, 3, 2),
                sell(JewelryItems.copper_ring.item(), 4, 1, 12, 4)
        ));
        trades.put(2, List.of(
                buy(Items.GOLD_INGOT, 7, 8, 2, 8),
                sell(JewelryItems.iron_ring.item(), 4, 1, 6, 5),
                sell(JewelryItems.gold_ring.item(), 18, 1, 6, 5)
        ));
        trades.put(3, List.of(
                buy(Items.DIAMOND, 1, 12, 10, 10),
                sell(JewelryItems.emerald_necklace.item(), 20, 1, 12, 10),
                sell(JewelryItems.diamond_necklace.item(), 25, 1, 12, 10)
        ));
        trades.put(4, List.of(
                sell(JewelryItems.ruby_ring.item(), 35, 1, 5, 15),
                sell(JewelryItems.topaz_ring.item(), 35, 1, 5, 15),
                sell(JewelryItems.citrine_ring.item(), 35, 1, 5, 15),
                sell(JewelryItems.jade_ring.item(), 35, 1, 5, 15),
                sell(JewelryItems.sapphire_ring.item(), 35, 1, 5, 13),
                sell(JewelryItems.tanzanite_ring.item(), 35, 1, 5, 13)
        ));
        trades.put(5, List.of(
                sell(JewelryItems.ruby_necklace.item(), 45, 1, 3, 15),
                sell(JewelryItems.topaz_necklace.item(), 45, 1, 3, 15),
                sell(JewelryItems.citrine_necklace.item(), 45, 1, 3, 15),
                sell(JewelryItems.jade_necklace.item(), 45, 1, 3, 15),
                sell(JewelryItems.sapphire_necklace.item(), 45, 1, 3, 15),
                sell(JewelryItems.tanzanite_necklace.item(), 45, 1, 3, 15)
        ));

        return trades;
    }
}
''')

# Registry#getEntry(Identifier) is a later convenience overload. The 1.20.1 registry directly
# resolves the Item by Identifier, which is all the creative-tab icon supplier needs.
group = common / "net/jewelry/items/Group.java"
g = group.read_text()
g = g.replace("Registries.ITEM.getEntry(JewelryItems.ruby_ring.id()).get().value()",
              "Registries.ITEM.get(JewelryItems.ruby_ring.id())")
group.write_text(g)

for required in (
    "libs/structure-pool-api-common.jar",
    "libs/spell-power-common.jar",
    "libs/ranged-weapon-api-common.jar",
):
    if required not in common_build.read_text():
        raise SystemExit(f"compat pass 2 missing named common dependency: {required}")

v = villagers.read_text()
for forbidden in ("TradeOffers.BuyItemFactory", "TradeOffers.SellItemFactory"):
    if forbidden in v:
        raise SystemExit(f"compat pass 2 left inaccessible 1.21 trade helper: {forbidden}")
for required in (
    "buy(Items.COPPER_INGOT, 8, 8, 3, 2)",
    "sell(JewelryItems.gold_ring.item(), 18, 1, 6, 5)",
    "sell(JewelryItems.tanzanite_necklace.item(), 45, 1, 3, 15)",
    "new TradeOffer(",
):
    if required not in v:
        raise SystemExit(f"compat pass 2 lost current 2.4.0 villager trade semantics: {required}")
if "getEntry(JewelryItems.ruby_ring.id())" in group.read_text():
    raise SystemExit("compat pass 2 left post-1.20.1 Registry#getEntry(Identifier) call")

print("Jewelry compatibility pass 2 applied: named foundation classpath + 1.20.1 trade factories/registry lookup")
