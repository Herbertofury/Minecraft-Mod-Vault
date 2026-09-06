package com.bert.slrsharedmana;

import com.mojang.logging.LogUtils;
import net.minecraft.server.level.ServerPlayer;
import net.minecraftforge.common.MinecraftForge;
import net.minecraftforge.event.RegisterCommandsEvent;
import net.minecraftforge.event.entity.player.PlayerEvent;
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
        MinecraftForge.EVENT_BUS.addListener(this::playerLoggedIn);
        MinecraftForge.EVENT_BUS.addListener(this::playerRespawned);
        MinecraftForge.EVENT_BUS.addListener(this::playerChangedDimension);
    }

    private void registerCommands(RegisterCommandsEvent event) {
        Commands.register(event.getDispatcher());
    }

    private void playerLoggedIn(PlayerEvent.PlayerLoggedInEvent event) {
        bind(event);
    }

    private void playerRespawned(PlayerEvent.PlayerRespawnEvent event) {
        bind(event);
    }

    private void playerChangedDimension(PlayerEvent.PlayerChangedDimensionEvent event) {
        bind(event);
    }

    private void bind(PlayerEvent event) {
        if (event.getEntity() instanceof ServerPlayer player) {
            SlrAccess.bindOwner(player);
        }
    }

    private void serverStarted(ServerStartedEvent event) {
        if (BridgeConfig.LOG_COMPAT_SUMMARY.get()) {
            LOGGER.info("SLR Shared Mana mode={} ratio={} SLR-MP/Iron-mana; Iron's Botany mode={}; passive ISS regen suppressed={}",
                    BridgeConfig.modeName(), BridgeConfig.SLR_MP_PER_IRON_MANA.get(), BotanyCompat.detectedMode(),
                    BridgeConfig.classicSharedMode() && BridgeConfig.SUPPRESS_ISS_PASSIVE_REGEN.get());
        }
    }
}
