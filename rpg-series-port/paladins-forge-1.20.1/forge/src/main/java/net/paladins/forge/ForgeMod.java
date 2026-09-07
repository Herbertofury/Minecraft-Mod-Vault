package net.paladins.forge;

import net.minecraft.registry.Registries;
import net.minecraft.registry.Registry;
import net.minecraft.registry.RegistryKeys;
import net.minecraft.world.poi.PointOfInterestType;
import net.minecraftforge.common.MinecraftForge;
import net.minecraftforge.event.BuildCreativeModeTabContentsEvent;
import net.minecraftforge.event.village.VillagerTradesEvent;
import net.minecraftforge.eventbus.api.IEventBus;
import net.minecraftforge.fml.common.Mod;
import net.minecraftforge.fml.javafmlmod.FMLJavaModLoadingContext;
import net.minecraftforge.registries.RegisterEvent;
import net.paladins.PaladinsMod;
import net.paladins.block.PaladinBlocks;
import net.paladins.item.Group;
import net.paladins.village.PaladinVillagers;
import net.spell_engine.compat.registry.RegistrationBridge;

/** Native Forge 47 translation of the current Paladins 3.1.1 NeoForge entrypoint. */
@Mod(PaladinsMod.ID)
public final class ForgeMod {
    public ForgeMod() {
        IEventBus modBus = FMLJavaModLoadingContext.get().getModEventBus();
        PaladinsMod.init();
        modBus.addListener(ForgeMod::register);
        modBus.addListener(ForgeMod::buildTabContents);
        MinecraftForge.EVENT_BUS.addListener(ForgeMod::onVillagerTrades);
    }

    /** Forge owns insertion and phase ordering; common code retains all definitions/config/side effects. */
    public static void register(RegisterEvent event) {
        event.register(RegistryKeys.SOUND_EVENT, helper ->
                withNative(Registries.SOUND_EVENT, helper, PaladinsMod::registerSounds));

        event.register(RegistryKeys.ENTITY_TYPE, helper ->
                withNative(Registries.ENTITY_TYPE, helper, PaladinsMod::registerEntities));

        event.register(RegistryKeys.BLOCK, helper ->
                withNative(Registries.BLOCK, helper, PaladinsMod::registerBlocks));

        event.register(Registries.ITEM_GROUP.getKey(), helper ->
                withNative(Registries.ITEM_GROUP, helper, PaladinsMod::registerItemGroup));

        event.register(RegistryKeys.ITEM, helper ->
                withNative(Registries.ITEM, helper, () -> {
                    PaladinsMod.registerBlockItems();
                    PaladinsMod.registerItems();
                }));

        event.register(RegistryKeys.STATUS_EFFECT, helper ->
                withNative(Registries.STATUS_EFFECT, helper, PaladinsMod::registerEffects));

        event.register(RegistryKeys.POINT_OF_INTEREST_TYPE, helper ->
                helper.register(PaladinVillagers.POI_ID,
                        new PointOfInterestType(PaladinVillagers.poiBlockStates(),
                                PaladinVillagers.POI_TICKET_COUNT, PaladinVillagers.POI_SEARCH_DISTANCE)));

        event.register(RegistryKeys.VILLAGER_PROFESSION, helper ->
                withNative(Registries.VILLAGER_PROFESSION, helper, PaladinsMod::registerVillagers));
    }

    @SuppressWarnings("unchecked")
    private static <T> void withNative(Registry<T> registry, RegisterEvent.RegisterHelper<T> helper, Runnable action) {
        RegistrationBridge.withRegistrar(registry,
                (actualRegistry, id, value) -> helper.register(id, (T) value),
                action);
    }

    private static void buildTabContents(BuildCreativeModeTabContentsEvent event) {
        if (event.getTabKey().equals(Group.KEY)) event.accept(() -> PaladinBlocks.MONK_WORKBENCH_BLOCK);
    }

    private static void onVillagerTrades(VillagerTradesEvent event) {
        if (event.getType() != PaladinVillagers.PROFESSION) return;
        PaladinVillagers.TRADES.forEach((tier, factories) -> {
            var tierList = event.getTrades().get(tier.intValue());
            if (tierList != null) tierList.addAll(factories);
        });
    }
}
