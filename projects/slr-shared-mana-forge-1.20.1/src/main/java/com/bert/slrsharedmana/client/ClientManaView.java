package com.bert.slrsharedmana.client;

import com.bert.slrsharedmana.SlrAccess;
import net.minecraft.client.Minecraft;
import net.minecraft.world.entity.LivingEntity;

import java.lang.reflect.Method;

public final class ClientManaView {
    private static volatile Method getPlayerMagicData;

    private ClientManaView() {}

    public static Float getFor(Object magicData) {
        var player = Minecraft.getInstance().player;
        if (player == null) return null;
        try {
            Method m = getPlayerMagicData;
            if (m == null) {
                synchronized (ClientManaView.class) {
                    m = getPlayerMagicData;
                    if (m == null) {
                        Class<?> c = Class.forName("io.redspace.ironsspellbooks.api.magic.MagicData");
                        m = c.getMethod("getPlayerMagicData", LivingEntity.class);
                        m.setAccessible(true);
                        getPlayerMagicData = m;
                    }
                }
            }
            Object localData = m.invoke(null, player);
            if (localData != magicData) return null;
            return (float) SlrAccess.ironEquivalent(player);
        } catch (ReflectiveOperationException | RuntimeException ignored) {
            return null;
        }
    }
}
