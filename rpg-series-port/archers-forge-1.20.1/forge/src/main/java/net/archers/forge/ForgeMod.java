package net.archers.forge;

import net.archers.ArchersMod;
import net.archers.block.ArcherBlocks;
import net.archers.forge.compat.curios.QuiverCurios;
import net.archers.item.Group;
import net.archers.item.Quivers;
import net.archers.item.misc.Misc;
import net.archers.village.ArcherVillagers;
import net.minecraft.registry.Registries;
import net.minecraft.registry.Registry;
import net.minecraft.registry.RegistryKeys;
import net.minecraft.world.poi.PointOfInterestType;
import net.minecraftforge.common.MinecraftForge;
import net.minecraftforge.event.BuildCreativeModeTabContentsEvent;
import net.minecraftforge.event.village.VillagerTradesEvent;
import net.minecraftforge.eventbus.api.IEventBus;
import net.minecraftforge.fml.ModList;
import net.minecraftforge.fml.common.Mod;
import net.minecraftforge.fml.event.lifecycle.FMLCommonSetupEvent;
import net.minecraftforge.fml.javafmlmod.FMLJavaModLoadingContext;
import net.minecraftforge.registries.RegisterEvent;

/** Native Forge 47 translation of the current Archers 3.1.1 NeoForge entrypoint. */
@Mod(ArchersMod.ID)
public final class ForgeMod {
    public ForgeMod() {
        IEventBus modBus = FMLJavaModLoadingContext.get().getModEventBus();
        ArchersMod.init();
        modBus.addListener(ForgeMod::register);
        modBus.addListener(ForgeMod::buildTabContents);
        modBus.addListener(ForgeMod::commonSetup);
        MinecraftForge.EVENT_BUS.addListener(ForgeMod::onVillagerTrades);
    }

    private static void commonSetup(FMLCommonSetupEvent event) {
        if (ModList.get().isLoaded("curios")) {
            event.enqueueWork(QuiverCurios::register);
        }
    }

    public static void register(RegisterEvent event) {
        event.register(RegistryKeys.SOUND_EVENT, helper -> ArchersMod.registerSounds());
        event.register(RegistryKeys.ITEM, helper -> {
            ArchersMod.registerEntities();
            ArchersMod.registerItems();
        });
        event.register(RegistryKeys.BLOCK, helper -> ArchersMod.registerBlocks());
        event.register(RegistryKeys.STATUS_EFFECT, helper -> ArchersMod.registerEffects());
        event.register(RegistryKeys.POINT_OF_INTEREST_TYPE, helper -> {
            Registry.register(Registries.POINT_OF_INTEREST_TYPE, ArcherVillagers.POI_ID,
                    new PointOfInterestType(ArcherVillagers.poiBlockStates(),
                            ArcherVillagers.POI_TICKET_COUNT, ArcherVillagers.POI_SEARCH_DISTANCE));
        });
        event.register(RegistryKeys.VILLAGER_PROFESSION, helper -> ArchersMod.registerVillagers());
    }

    private static void buildTabContents(BuildCreativeModeTabContentsEvent event) {
        if (!event.getTabKey().equals(Group.KEY)) return;
        for (var entry : Misc.ENTRIES) event.accept(entry.item());
        for (var entry : ArcherBlocks.all) event.accept(entry.item());
        if (ModList.get().isLoaded("bundleapi")) {
            for (var entry : Quivers.entries) event.accept(entry.item());
        }
    }

    private static void onVillagerTrades(VillagerTradesEvent event) {
        if (event.getType() != ArcherVillagers.PROFESSION) return;
        ArcherVillagers.TRADES.forEach((tier, factories) -> {
            var tierList = event.getTrades().get(tier.intValue());
            if (tierList != null) tierList.addAll(factories);
        });
    }
}
