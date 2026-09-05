package com.bert.slrsharedmana;

import net.minecraftforge.fml.ModList;

import java.lang.reflect.Field;
import java.lang.reflect.Method;

public final class BotanyCompat {
    private BotanyCompat() {}

    public static String detectedMode() {
        if (!ModList.get().isLoaded("ironsbotany")) return "NOT_INSTALLED";
        try {
            Class<?> config = Class.forName("com.ironsbotany.common.config.CommonConfig");
            Field f = config.getField("MANA_UNIFICATION_MODE");
            Object configValue = f.get(null);
            Method get = configValue.getClass().getMethod("get");
            Object mode = get.invoke(configValue);
            return String.valueOf(mode);
        } catch (ReflectiveOperationException e) {
            return "INSTALLED_API_UNRESOLVED";
        }
    }
}
