package com.bert.slrsharedmana;

import com.mojang.logging.LogUtils;
import net.minecraftforge.common.MinecraftForge;
import net.minecraftforge.event.RegisterCommandsEvent;
import net.minecraftforge.event.server.ServerStartedEvent;
import net.minecraftforge.fml.ModLoadingContext;
import net.minecraftforge.fml.common.Mod;
import net.minecraftforge.fml.config.ModConfig;
import org.slf4j.Logger;

@Mod(SharedManaMod.MODID)
public final class SharedManaMod {
    public static final String MODID = "slr_shared_mana";
    public static final Logger LOGGER = LogUtils.getLogger();

    public SharedManaMod() {
        ModLoadingContext.get().registerConfig(ModConfig.Type.COMMON, BridgeConfig.SPEC, "slr-shared-mana.toml");
        MinecraftForge.EVENT_BUS.addListener(this::registerCommands);
        MinecraftForge.EVENT_BUS.addListener(this::serverStarted);
    }

    private void registerCommands(RegisterCommandsEvent event) {
        Commands.register(event.getDispatcher());
    }

    private void serverStarted(ServerStartedEvent event) {
        if (BridgeConfig.LOG_COMPAT_SUMMARY.get()) {
            LOGGER.info("SLR Shared Mana active={} ratio={} SLR-MP/Iron-mana; Iron's Botany mode={}; passive ISS regen suppressed={}",
                    BridgeConfig.ENABLED.get(), BridgeConfig.SLR_MP_PER_IRON_MANA.get(), BotanyCompat.detectedMode(),
                    BridgeConfig.SUPPRESS_ISS_PASSIVE_REGEN.get());
        }
    }
}
