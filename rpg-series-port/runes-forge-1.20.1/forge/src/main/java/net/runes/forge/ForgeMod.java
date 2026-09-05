package net.runes.forge;

import net.minecraft.advancement.criterion.Criteria;
import net.minecraft.item.ItemGroups;
import net.minecraft.registry.RegistryKeys;
import net.minecraft.util.Identifier;
import net.minecraftforge.common.MinecraftForge;
import net.minecraftforge.common.extensions.IForgeMenuType;
import net.minecraftforge.event.BuildCreativeModeTabContentsEvent;
import net.minecraftforge.eventbus.api.IEventBus;
import net.minecraftforge.fml.common.Mod;
import net.minecraftforge.fml.javafmlmod.FMLJavaModLoadingContext;
import net.minecraftforge.registries.RegisterEvent;
import net.runes.RunesMod;
import net.runes.api.RuneItems;
import net.runes.crafting.*;

@Mod(RunesMod.ID)
public final class ForgeMod {
    private static final Identifier ALTAR_ID = new Identifier(RunesMod.ID, RuneCraftingBlock.NAME);

    public ForgeMod() {
        // Do not construct Item/Block/MenuType instances here. Forge's intrusive registries are
        // frozen during mod construction and are opened only for their matching RegisterEvent.
        Criteria.register(RuneCraftingCriteria.INSTANCE);

        IEventBus bus = FMLJavaModLoadingContext.get().getModEventBus();
        bus.addListener(ForgeMod::register);
        bus.addListener(ForgeMod::buildCreativeTab);

        if (Boolean.getBoolean("runes.ci.selftest")) {
            MinecraftForge.EVENT_BUS.addListener(CiSelfTest::onServerStarted);
        }
    }

    private static void register(RegisterEvent event) {
        event.register(RegistryKeys.SOUND_EVENT, helper ->
                helper.register(RunesMod.CRAFTING_ID, RunesMod.CRAFTING_SOUND));

        event.register(RegistryKeys.RECIPE_TYPE, helper -> {
            RuneCrafting.bootstrapType();
            helper.register(RunesMod.CRAFTING_ID, RuneCrafting.RECIPE_TYPE);
        });

        event.register(RegistryKeys.RECIPE_SERIALIZER, helper -> {
            RuneCrafting.bootstrapSerializer();
            helper.register(RunesMod.CRAFTING_ID, RuneCrafting.RECIPE_SERIALIZER);
        });

        event.register(RegistryKeys.BLOCK, helper -> {
            RuneCraftingBlock.bootstrapBlock();
            helper.register(ALTAR_ID, RuneCraftingBlock.INSTANCE);
        });

        event.register(RegistryKeys.SCREEN_HANDLER, helper -> {
            RuneCraftingScreenHandler.bindHandlerType(IForgeMenuType.create(RuneCraftingScreenHandler::new));
            helper.register(RunesMod.CRAFTING_ID, RuneCraftingScreenHandler.handlerType());
        });

        event.register(RegistryKeys.ITEM, helper -> {
            RuneCraftingBlock.bootstrapItem();
            RuneItems.bootstrap();
            RunePouches.bootstrap();

            helper.register(ALTAR_ID, RuneCraftingBlock.ITEM);
            for (var entry : RuneItems.all()) helper.register(entry.id(), entry.item());
            for (var entry : RunePouches.all()) helper.register(entry.id(), entry.item());
        });
    }

    private static void buildCreativeTab(BuildCreativeModeTabContentsEvent event) {
        if (event.getTabKey().equals(ItemGroups.FUNCTIONAL)) {
            event.accept(() -> RuneCraftingBlock.ITEM);
        }
        if (event.getTabKey().equals(ItemGroups.COMBAT)) {
            for (var entry : RuneItems.all()) event.accept(entry::item);
            for (var entry : RunePouches.all()) event.accept(entry::item);
        }
    }
}
