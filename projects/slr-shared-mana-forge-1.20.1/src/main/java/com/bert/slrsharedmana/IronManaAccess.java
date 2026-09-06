package com.bert.slrsharedmana;

import net.minecraft.server.level.ServerPlayer;
import net.minecraft.world.entity.LivingEntity;

import java.lang.reflect.Method;

/** Cached reflection access to Iron's native MagicData. Used only by opt-in separate-pools borrowing. */
public final class IronManaAccess {
    private static volatile State state;

    private record State(Method getPlayerMagicData, Method getMana, Method setMana, Method syncMana) {}

    private IronManaAccess() {}

    private static State state() {
        State s = state;
        if (s != null) return s;
        synchronized (IronManaAccess.class) {
            s = state;
            if (s != null) return s;
            try {
                Class<?> magicData = Class.forName("io.redspace.ironsspellbooks.api.magic.MagicData");
                Method lookup = magicData.getMethod("getPlayerMagicData", LivingEntity.class);
                Method get = magicData.getMethod("getMana");
                Method set = magicData.getMethod("setMana", float.class);

                // Iron's native callers pair setMana with this packet helper. Reuse the exact path so
                // separate-pool mode's visible Iron mana HUD updates immediately after an SLR borrow.
                Class<?> updateClient = Class.forName("io.redspace.ironsspellbooks.api.util.UpdateClient");
                Method sync = updateClient.getMethod("SendManaUpdate", ServerPlayer.class, magicData);

                lookup.setAccessible(true);
                get.setAccessible(true);
                set.setAccessible(true);
                sync.setAccessible(true);
                state = s = new State(lookup, get, set, sync);
                return s;
            } catch (ReflectiveOperationException e) {
                throw new IllegalStateException("Iron's Spells 1.20.1 mana API could not be resolved", e);
            }
        }
    }

    private static Object data(ServerPlayer player) {
        try {
            return state().getPlayerMagicData().invoke(null, player);
        } catch (ReflectiveOperationException e) {
            throw new IllegalStateException("Iron's native MagicData lookup failed", e);
        }
    }

    public static double current(ServerPlayer player) {
        if (player == null) return 0.0D;
        try {
            Object data = data(player);
            if (data == null) return 0.0D;
            Object value = state().getMana().invoke(data);
            if (!(value instanceof Number n)) return 0.0D;
            double mana = n.doubleValue();
            return Double.isFinite(mana) ? Math.max(0.0D, mana) : 0.0D;
        } catch (ReflectiveOperationException e) {
            throw new IllegalStateException("Iron's native mana read failed", e);
        }
    }

    /**
     * Uses Iron's own setMana path (including ChangeManaEvent/max-mana handling) and then Iron's own
     * client mana sync helper. This keeps the native separate-pool HUD authoritative and immediate.
     */
    public static double set(ServerPlayer player, double requestedMana) {
        if (player == null || !Double.isFinite(requestedMana)) return current(player);
        try {
            State s = state();
            Object data = data(player);
            if (data == null) return 0.0D;
            s.setMana().invoke(data, (float) Math.max(0.0D, requestedMana));
            s.syncMana().invoke(null, player, data);
            Object value = s.getMana().invoke(data);
            double result = value instanceof Number n ? n.doubleValue() : 0.0D;
            return Double.isFinite(result) ? Math.max(0.0D, result) : 0.0D;
        } catch (ReflectiveOperationException e) {
            throw new IllegalStateException("Iron's native mana write/sync failed", e);
        }
    }
}
