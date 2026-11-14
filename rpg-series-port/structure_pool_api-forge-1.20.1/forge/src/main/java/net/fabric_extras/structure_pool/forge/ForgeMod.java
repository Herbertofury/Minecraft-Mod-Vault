package net.fabric_extras.structure_pool.forge;

import net.fabric_extras.structure_pool.StructurePoolMod;
import net.fabric_extras.structure_pool.api.StructurePoolAPI;
import net.minecraftforge.common.MinecraftForge;
import net.minecraftforge.event.server.ServerStartingEvent;
import net.minecraftforge.fml.common.Mod;

@Mod(StructurePoolMod.ID)
public final class ForgeMod {
    public ForgeMod() {
        StructurePoolMod.init();
        MinecraftForge.EVENT_BUS.addListener(this::onServerStarting);
    }

    private void onServerStarting(ServerStartingEvent event) {
        StructurePoolAPI.processInjections(event.getServer());
    }
}
