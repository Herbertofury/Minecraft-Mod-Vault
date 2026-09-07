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
import net.minecraftforge.event.server.ServerStartedEvent;
import net.minecraftforge.event.village.VillagerTradesEvent;
import net.minecraftforge.eventbus.api.IEventBus;
import net.minecraftforge.fml.ModList;
import net.minecraftforge.fml.common.Mod;
import net.minecraftforge.fml.event.lifecycle.FMLCommonSetupEvent;
import net.minecraftforge.fml.javafmlmod.FMLJavaModLoadingContext;
import net.minecraftforge.registries.RegisterEvent;
import net.spell_engine.compat.registry.RegistrationBridge;

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
        MinecraftForge.EVENT_BUS.addListener(ForgeMod::onServerStarted);
    }

    private static void commonSetup(FMLCommonSetupEvent event) {
        if (ModList.get().isLoaded("curios")) {
            event.enqueueWork(QuiverCurios::register);
        }
    }

    /**
     * Forge 47 owns registry mutation. Current Archers/Spell Engine common code still owns object
     * construction, config application and side effects; RegistrationBridge routes only the actual
     * insertion through the matching RegisterHelper and rejects cross-registry phase drift.
     */
    public static void register(RegisterEvent event) {
        event.register(RegistryKeys.SOUND_EVENT, helper ->
                withNative(Registries.SOUND_EVENT, helper, ArchersMod::registerSounds));

        event.register(RegistryKeys.ENTITY_TYPE, helper ->
                withNative(Registries.ENTITY_TYPE, helper, ArchersMod::registerEntities));

        event.register(RegistryKeys.BLOCK, helper ->
                withNative(Registries.BLOCK, helper, ArchersMod::registerBlocks));

        event.register(Registries.ITEM_GROUP.getKey(), helper ->
                withNative(Registries.ITEM_GROUP, helper, ArchersMod::registerItemGroup));

        event.register(RegistryKeys.ITEM, helper ->
                withNative(Registries.ITEM, helper, () -> {
                    ArcherBlocks.registerItems();
                    ArchersMod.registerItems();
                }));

        event.register(RegistryKeys.STATUS_EFFECT, helper ->
                withNative(Registries.STATUS_EFFECT, helper, ArchersMod::registerEffects));

        event.register(RegistryKeys.POINT_OF_INTEREST_TYPE, helper ->
                helper.register(ArcherVillagers.POI_ID,
                        new PointOfInterestType(ArcherVillagers.poiBlockStates(),
                                ArcherVillagers.POI_TICKET_COUNT, ArcherVillagers.POI_SEARCH_DISTANCE)));

        event.register(RegistryKeys.VILLAGER_PROFESSION, helper ->
                withNative(Registries.VILLAGER_PROFESSION, helper, ArchersMod::registerVillagers));
    }

    @SuppressWarnings("unchecked")
    private static <T> void withNative(Registry<T> registry, RegisterEvent.RegisterHelper<T> helper, Runnable action) {
        RegistrationBridge.withRegistrar(registry,
                (actualRegistry, id, value) -> helper.register(id, (T) value),
                action);
    }

    private static void buildTabContents(BuildCreativeModeTabContentsEvent event) {
        if (!event.getTabKey().equals(Group.KEY)) return;
        for (var entry : Misc.ENTRIES) event.accept(() -> entry.item());
        for (var entry : ArcherBlocks.all) event.accept(() -> entry.item());
        if (ModList.get().isLoaded("bundleapi")) {
            for (var entry : Quivers.entries) event.accept(() -> entry.item());
        }
    }

    private static void onVillagerTrades(VillagerTradesEvent event) {
        if (event.getType() != ArcherVillagers.PROFESSION) return;
        ArcherVillagers.TRADES.forEach((tier, factories) -> {
            var tierList = event.getTrades().get(tier.intValue());
            if (tierList != null) tierList.addAll(factories);
        });
    }

    private static void onServerStarted(ServerStartedEvent event) {
        ArchersCiSelfTest.runIfRequested();
    }
}