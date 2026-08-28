package net.fabric_extras.shield_api.forge;

import net.fabric_extras.shield_api.ShieldAPI;
import net.minecraftforge.fml.common.Mod;

@Mod(ShieldAPI.MOD_ID)
public final class ShieldAPIForge {
    public ShieldAPIForge() {
        ShieldAPI.init();
    }
}
