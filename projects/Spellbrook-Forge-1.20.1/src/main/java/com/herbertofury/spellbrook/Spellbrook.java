package com.herbertofury.spellbrook;

import com.herbertofury.spellbrook.compat.CompatHooks;
import com.herbertofury.spellbrook.network.SpellbrookNetwork;
import com.herbertofury.spellbrook.registry.ModEntities;
import com.herbertofury.spellbrook.registry.ModItems;
import com.herbertofury.spellbrook.registry.ModSounds;
import com.mojang.logging.LogUtils;
import net.minecraftforge.eventbus.api.IEventBus;
import net.minecraftforge.fml.ModList;
import net.minecraftforge.fml.common.Mod;
import net.minecraftforge.fml.javafmlmod.FMLJavaModLoadingContext;
import org.slf4j.Logger;

import java.lang.reflect.Method;

@Mod(Spellbrook.MOD_ID)
public final class Spellbrook {
    public static final String MOD_ID = "spellbrook";
    public static final Logger LOGGER = LogUtils.getLogger();

    public Spellbrook() {
        IEventBus modBus = FMLJavaModLoadingContext.get().getModEventBus();
        ModItems.register(modBus);
        ModEntities.register(modBus);
        ModSounds.register(modBus);
        modBus.addListener(ModItems::buildCreativeTab);
        modBus.addListener(event -> event.enqueueWork(SpellbrookNetwork::init));

        if (ModList.get().isLoaded("hexerei")) {
            bootstrapHexerei(modBus);
        }
    }

    private static void bootstrapHexerei(IEventBus modBus) {
        try {
            Class<?> bootstrap = Class.forName("com.herbertofury.spellbrook.compat.hexerei.HexereiCompatBootstrap");
            Method method = bootstrap.getMethod("register", IEventBus.class);
            method.invoke(null, modBus);
            LOGGER.info("Spellbrook: Hexerei integration enabled");
        } catch (Throwable t) {
            CompatHooks.disableHexerei();
            LOGGER.error("Spellbrook: Hexerei is installed but the compatibility bridge could not initialize", t);
        }
    }
}
