package com.bert.slrsharedmana;

import net.minecraft.server.level.ServerPlayer;

import java.lang.ref.WeakReference;
import java.util.Collections;
import java.util.Map;
import java.util.WeakHashMap;

/**
 * Demand-only SLR -> Iron fallback used by the SLR bytecode coremod.
 * No tick loop, no mirror, and no transfer while idle.
 */
public final class SlrBorrowBridge {
    private static final double EPS = 1.0e-7D;
    private static final Map<Object, WeakReference<ServerPlayer>> OWNERS =
            Collections.synchronizedMap(new WeakHashMap<>());

    private SlrBorrowBridge() {}

    public static void bind(Object variables, ServerPlayer player) {
        if (variables == null || player == null) return;
        OWNERS.put(variables, new WeakReference<>(player));
    }

    private static ServerPlayer owner(Object variables) {
        if (variables == null) return null;
        WeakReference<ServerPlayer> ref = OWNERS.get(variables);
        ServerPlayer player = ref == null ? null : ref.get();
        if (ref != null && player == null) OWNERS.remove(variables);
        return player;
    }

    /** Replacement for selected SLR PlayerVariables.MP GETFIELD instructions. */
    public static double readMp(Object variables) {
        double nativeSlr = sane(SlrAccess.rawCurrent(variables));
        if (!BridgeConfig.separateBorrowMode()) return nativeSlr;
        ServerPlayer player = owner(variables);
        if (player == null) return nativeSlr; // Fail closed: never invent access to another player's pool.
        double ratio = ratio();
        if (!(ratio > 0.0D)) return nativeSlr;
        return nativeSlr + sane(IronManaAccess.current(player)) * ratio;
    }

    /** Replacement for selected SLR PlayerVariables.MP PUTFIELD instructions. */
    public static void writeMp(Object variables, double requestedVirtualMp) {
        if (variables == null || !Double.isFinite(requestedVirtualMp)) return;

        double nativeSlr = sane(SlrAccess.rawCurrent(variables));
        if (!BridgeConfig.separateBorrowMode()) {
            SlrAccess.writeRawCurrent(variables, requestedVirtualMp);
            return;
        }

        ServerPlayer player = owner(variables);
        double ratio = ratio();
        if (player == null || !(ratio > 0.0D)) {
            SlrAccess.writeRawCurrent(variables, requestedVirtualMp);
            return;
        }

        double nativeIron = sane(IronManaAccess.current(player));
        double virtualBefore = nativeSlr + nativeIron * ratio;
        double delta = requestedVirtualMp - virtualBefore;

        if (delta < -EPS) {
            // Real SLR spend: consume SLR first, then debit only the shortage from Iron.
            double spend = -delta;
            double slrDebit = Math.min(nativeSlr, spend);
            double slrAfter = nativeSlr - slrDebit;
            double remaining = spend - slrDebit;
            SlrAccess.writeRawCurrent(variables, slrAfter);
            if (remaining > EPS) {
                IronManaAccess.set(player, nativeIron - remaining / ratio);
            }
            return;
        }

        if (delta > EPS) {
            // Positive SLR changes remain SLR-native. Iron is never credited by SLR regen/rewards.
            double max = sane(SlrAccess.rawMax(variables));
            double after = nativeSlr + delta;
            if (max > 0.0D) after = Math.min(max, after);
            SlrAccess.writeRawCurrent(variables, Math.max(0.0D, after));
            return;
        }

        // Exact/no-op writes keep the native pools untouched.
        SlrAccess.writeRawCurrent(variables, nativeSlr);
    }

    private static double ratio() {
        double ratio = BridgeConfig.SLR_MP_PER_IRON_MANA.get();
        return Double.isFinite(ratio) && ratio > 0.0D ? ratio : 0.0D;
    }

    private static double sane(double value) {
        return Double.isFinite(value) ? Math.max(0.0D, value) : 0.0D;
    }
}
