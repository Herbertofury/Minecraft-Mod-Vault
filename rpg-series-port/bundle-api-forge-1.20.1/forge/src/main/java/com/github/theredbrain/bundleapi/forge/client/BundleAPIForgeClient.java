package com.github.theredbrain.bundleapi.forge.client;

import com.github.theredbrain.bundleapi.BundleAPI;
import com.github.theredbrain.bundleapi.item.tooltip.CustomBundleTooltipData;
import net.minecraftforge.api.distmarker.Dist;
import net.minecraftforge.client.event.RegisterClientTooltipComponentFactoriesEvent;
import net.minecraftforge.eventbus.api.SubscribeEvent;
import net.minecraftforge.fml.common.Mod;

/** Native Forge client registration; kept out of the common entrypoint so dedicated servers never load client classes. */
@Mod.EventBusSubscriber(modid = BundleAPI.MOD_ID, bus = Mod.EventBusSubscriber.Bus.MOD, value = Dist.CLIENT)
public final class BundleAPIForgeClient {
    private BundleAPIForgeClient() {
    }

    @SubscribeEvent
    public static void registerTooltipFactories(RegisterClientTooltipComponentFactoriesEvent event) {
        event.register(CustomBundleTooltipData.class, CustomBundleTooltipComponent::new);
    }
}
