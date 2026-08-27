package com.github.theredbrain.bundleapi.forge;

import com.github.theredbrain.bundleapi.BundleAPI;
import net.minecraftforge.fml.common.Mod;

@Mod(BundleAPI.MOD_ID)
public final class BundleAPIForge {
    public BundleAPIForge() {
        BundleAPI.init();
    }
}
