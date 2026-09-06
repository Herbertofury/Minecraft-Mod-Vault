package com.bert.slrsharedmana;

import net.minecraft.server.level.ServerPlayer;
import net.minecraft.world.entity.Entity;
import net.minecraftforge.common.capabilities.Capability;

import java.lang.reflect.Field;
import java.lang.reflect.Method;
import java.util.Optional;

/** Cached, bounded access to SLR's public player capability. No classpath copy of SLR is required. */
public final class SlrAccess {
    private static final String VARIABLES = "net.solocraft.network.SololevelingModVariables";
    private static volatile State state;

    private record State(Capability<Object> capability, Field currentMp, Field maxMp, Method sync, Method cooldownSet) {}

    private SlrAccess() {}

    @SuppressWarnings("unchecked")
    private static State state() {
        State s = state;
        if (s != null) return s;
        synchronized (SlrAccess.class) {
            s = state;
            if (s != null) return s;
            try {
                Class<?> vars = Class.forName(VARIABLES);
                Field capabilityField = vars.getField("PLAYER_VARIABLES_CAPABILITY");
                Capability<Object> cap = (Capability<Object>) capabilityField.get(null);
                Class<?> pv = Class.forName(VARIABLES + "$PlayerVariables");
                Field current = pv.getField("MP");
                Field max = pv.getField("Mana");
                Method sync = pv.getMethod("syncPlayerVariables", Entity.class);
                Class<?> cooldown = Class.forName("net.solocraft.util.CooldownManager");
                Method cooldownSet = cooldown.getMethod("set", Entity.class, String.class, int.class);
                current.setAccessible(true);
                max.setAccessible(true);
                sync.setAccessible(true);
                cooldownSet.setAccessible(true);
                state = s = new State(cap, current, max, sync, cooldownSet);
                return s;
            } catch (ReflectiveOperationException e) {
                throw new IllegalStateException("SLR 1.20.1 mana API could not be resolved", e);
            }
        }
    }

    private static Optional<Object> variables(Entity player) {
        State s = state();
        Optional<Object> resolved = player.getCapability(s.capability(), null).resolve();
        if (player instanceof ServerPlayer serverPlayer) {
            resolved.ifPresent(v -> SlrBorrowBridge.bind(v, serverPlayer));
        }
        return resolved;
    }

    /** Bind the SLR PlayerVariables object to its server-side owner without polling. */
    public static void bindOwner(ServerPlayer player) {
        if (player != null) variables(player);
    }

    public static double current(Entity player) {
        return variables(player).map(SlrAccess::rawCurrent).orElse(0.0D);
    }

    public static double max(Entity player) {
        return variables(player).map(v -> Math.max(0.0D, rawMax(v))).orElse(0.0D);
    }

    public static double ironEquivalent(Entity player) {
        double ratio = BridgeConfig.SLR_MP_PER_IRON_MANA.get();
        return ratio <= 0.0D ? 0.0D : Math.max(0.0D, current(player) / ratio);
    }

    /** Apply a delta in Iron-mana units and return the resulting SLR MP. Classic shared-mode path. */
    public static double applyIronDelta(ServerPlayer player, double ironDelta) {
        if (!Double.isFinite(ironDelta) || Math.abs(ironDelta) < 1.0e-7D) return current(player);
        double ratio = BridgeConfig.SLR_MP_PER_IRON_MANA.get();
        if (!(ratio > 0.0D) || !Double.isFinite(ratio)) return current(player);
        if (ironDelta > 0.0D && !BridgeConfig.ACCEPT_POSITIVE_ISS_CREDITS.get()) return current(player);

        Optional<Object> opt = variables(player);
        if (opt.isEmpty()) return 0.0D;
        Object v = opt.get();
        double before = Math.max(0.0D, rawCurrent(v));
        double max = Math.max(0.0D, rawMax(v));
        double requested = before + ironDelta * ratio;
        double after = Math.max(0.0D, Math.min(max, requested));
        if (Math.abs(after - before) > 1.0e-7D) {
            writeRawCurrent(v, after);
            syncRaw(v, player);
        }
        return after;
    }

    public static void triggerManaRefresh(ServerPlayer player) {
        int ticks = BridgeConfig.MANA_REFRESH_TICKS.get();
        if (ticks <= 0) return;
        try {
            state().cooldownSet().invoke(null, player, "mana_refresh", ticks);
        } catch (ReflectiveOperationException e) {
            throw new IllegalStateException("SLR mana_refresh cooldown could not be applied", e);
        }
    }

    static double rawCurrent(Object variables) {
        if (variables == null) return 0.0D;
        return read(state().currentMp(), variables);
    }

    static double rawMax(Object variables) {
        if (variables == null) return 0.0D;
        return read(state().maxMp(), variables);
    }

    static void writeRawCurrent(Object variables, double value) {
        if (variables == null || !Double.isFinite(value)) return;
        write(state().currentMp(), variables, value);
    }

    static void syncRaw(Object variables, ServerPlayer player) {
        if (variables == null || player == null) return;
        try {
            state().sync().invoke(variables, player);
        } catch (ReflectiveOperationException e) {
            throw new IllegalStateException("SLR player mana sync failed", e);
        }
    }

    private static double read(Field f, Object target) {
        try {
            return f.getDouble(target);
        } catch (IllegalAccessException e) {
            throw new IllegalStateException(e);
        }
    }

    private static void write(Field f, Object target, double value) {
        try {
            f.setDouble(target, value);
        } catch (IllegalAccessException e) {
            throw new IllegalStateException(e);
        }
    }
}
