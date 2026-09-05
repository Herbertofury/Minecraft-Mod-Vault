package com.bert.slrsharedmana.client;

import com.bert.slrsharedmana.BridgeConfig;
import com.bert.slrsharedmana.SharedManaMod;
import net.minecraftforge.api.distmarker.Dist;
import net.minecraftforge.fml.common.Mod;
import net.minecraftforge.fml.event.lifecycle.FMLClientSetupEvent;
import net.minecraftforge.eventbus.api.SubscribeEvent;

import java.lang.reflect.Field;
import java.lang.reflect.Method;

@Mod.EventBusSubscriber(modid = SharedManaMod.MODID, bus = Mod.EventBusSubscriber.Bus.MOD, value = Dist.CLIENT)
public final class ClientIntegration {
    private ClientIntegration() {}

    @SubscribeEvent
    public static void clientSetup(FMLClientSetupEvent event) {
        if (!BridgeConfig.HIDE_ISS_MANA_HUD.get()) return;
        event.enqueueWork(ClientIntegration::hideIronManaHud);
    }

    @SuppressWarnings({"rawtypes", "unchecked"})
    private static void hideIronManaHud() {
        try {
            Class<?> configs = Class.forName("io.redspace.ironsspellbooks.config.ClientConfigs");
            Field displayField = configs.getField("MANA_BAR_DISPLAY");
            Object configValue = displayField.get(null);
            Class<?> display = Class.forName("io.redspace.ironsspellbooks.gui.overlays.ManaBarOverlay$Display");
            Object never = Enum.valueOf((Class<? extends Enum>) display.asSubclass(Enum.class), "Never");
            Method set = configValue.getClass().getMethod("set", Object.class);
            set.invoke(configValue, never);

            try {
                Field textField = configs.getField("MANA_BAR_TEXT_VISIBLE");
                Object textValue = textField.get(null);
                textValue.getClass().getMethod("set", Object.class).invoke(textValue, Boolean.FALSE);
            } catch (ReflectiveOperationException ignored) {
                // Older/newer ISS may not expose this secondary flag; hiding the bar is sufficient.
            }
            SharedManaMod.LOGGER.info("Iron's mana HUD hidden; SLR MP HUD remains authoritative");
        } catch (ReflectiveOperationException e) {
            SharedManaMod.LOGGER.warn("Could not hide Iron's mana HUD through its client config; mana sharing still works", e);
        }
    }
}
