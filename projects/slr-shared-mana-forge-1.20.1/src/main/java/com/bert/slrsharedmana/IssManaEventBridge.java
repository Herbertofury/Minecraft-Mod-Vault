package com.bert.slrsharedmana;

import net.minecraft.server.level.ServerPlayer;
import net.minecraft.world.entity.player.Player;
import net.minecraftforge.common.MinecraftForge;
import net.minecraftforge.eventbus.api.Event;

import java.lang.reflect.Constructor;
import java.lang.reflect.Method;

/** Re-posts ISS's own cancellable ChangeManaEvent so third-party listeners keep their normal semantics. */
public final class IssManaEventBridge {
    private static volatile State state;
    private record State(Constructor<?> ctor, Method getNewMana) {}

    private IssManaEventBridge() {}

    private static State state(Object magicData) {
        State s = state;
        if (s != null) return s;
        synchronized (IssManaEventBridge.class) {
            s = state;
            if (s != null) return s;
            try {
                Class<?> eventClass = Class.forName("io.redspace.ironsspellbooks.api.events.ChangeManaEvent");
                Class<?> magicDataClass = Class.forName("io.redspace.ironsspellbooks.api.magic.MagicData");
                Constructor<?> ctor = eventClass.getConstructor(Player.class, magicDataClass, float.class, float.class);
                Method getNew = eventClass.getMethod("getNewMana");
                ctor.setAccessible(true);
                getNew.setAccessible(true);
                state = s = new State(ctor, getNew);
                return s;
            } catch (ReflectiveOperationException e) {
                throw new IllegalStateException("Iron's Spells ChangeManaEvent API could not be resolved", e);
            }
        }
    }

    public static float resolve(ServerPlayer player, Object magicData, float oldMana, float requestedMana) {
        try {
            State s = state(magicData);
            Event event = (Event) s.ctor().newInstance(player, magicData, oldMana, requestedMana);
            boolean cancelled = MinecraftForge.EVENT_BUS.post(event);
            if (cancelled) return oldMana;
            Object value = s.getNewMana().invoke(event);
            return value instanceof Number n ? n.floatValue() : requestedMana;
        } catch (ReflectiveOperationException e) {
            throw new IllegalStateException("Iron's Spells mana event dispatch failed", e);
        }
    }
}
