package com.github.theredbrain.bundleapi.forge;

import com.github.theredbrain.bundleapi.BundleAPI;
import net.minecraftforge.eventbus.api.IEventBus;
import net.minecraftforge.fml.common.Mod;
import net.minecraftforge.fml.event.lifecycle.FMLCommonSetupEvent;
import net.minecraftforge.fml.javafmlmod.FMLJavaModLoadingContext;

@Mod(BundleAPI.MOD_ID)
public final class BundleAPIForge {
    private static final boolean SELF_TEST_ENABLED = "1".equals(System.getenv("BUNDLE_API_SELF_TEST"));

    public BundleAPIForge() {
        BundleAPI.init();
        IEventBus modBus = FMLJavaModLoadingContext.get().getModEventBus();
        if (SELF_TEST_ENABLED) {
            BundleAPISelfTest.TEST_ITEMS.register(modBus);
        }
        modBus.addListener(this::commonSetup);
    }

    private void commonSetup(FMLCommonSetupEvent event) {
        if (SELF_TEST_ENABLED) {
            event.enqueueWork(BundleAPISelfTest::run);
        }
    }
}
