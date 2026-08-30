package net.fabric_extras.shield_api.forge;

import net.fabric_extras.shield_api.ShieldAPI;
import net.minecraftforge.common.MinecraftForge;
import net.minecraftforge.fml.common.Mod;
import net.minecraftforge.fml.javafmlmod.FMLJavaModLoadingContext;

@Mod(ShieldAPI.MOD_ID)
public final class ShieldAPIForge {
    public ShieldAPIForge() {
        ShieldAPI.init();
        if (ShieldAPISelfTest.enabled()) {
            ShieldAPISelfTest.TEST_ITEMS.register(FMLJavaModLoadingContext.get().getModEventBus());
            MinecraftForge.EVENT_BUS.addListener(ShieldAPISelfTest::onServerStarted);
        }
    }
}
