package com.github.theredbrain.bundleapi.forge;

import com.github.theredbrain.bundleapi.BundleAPI;
import net.minecraftforge.eventbus.api.IEventBus;
import net.minecraftforge.fml.common.Mod;
import net.minecraftforge.fml.event.lifecycle.FMLCommonSetupEvent;
import net.minecraftforge.fml.javafmlmod.FMLJavaModLoadingContext;

@Mod(BundleAPI.MOD_ID)
public final class BundleAPIForge {
    public BundleAPIForge() {
        BundleAPI.init();
        IEventBus modBus = FMLJavaModLoadingContext.get().getModEventBus();
        modBus.addListener(this::commonSetup);
    }

    private void commonSetup(FMLCommonSetupEvent event) {
        if ("1".equals(System.getenv("BUNDLE_API_SELF_TEST"))) {
            event.enqueueWork(BundleAPISelfTest::run);
        }
    }
}
