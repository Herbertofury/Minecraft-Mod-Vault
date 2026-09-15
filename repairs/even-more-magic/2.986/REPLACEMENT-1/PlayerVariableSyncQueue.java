package net.mcreator.evenmoremagic.network;

import java.util.IdentityHashMap;
import java.util.LinkedHashMap;
import java.util.Map;
import net.minecraft.server.MinecraftServer;
import net.minecraft.server.level.ServerPlayer;
import net.minecraft.world.entity.Entity;

/** Coalesces bursty capability sync requests into one queued send per player. */
public final class PlayerVariableSyncQueue {
    private static final IdentityHashMap<MinecraftServer, LinkedHashMap<ServerPlayer, EvenMoreMagicModVariables.PlayerVariables>> PENDING = new IdentityHashMap<>();
    private PlayerVariableSyncQueue() {}

    public static void request(EvenMoreMagicModVariables.PlayerVariables variables, Entity entity) {
        if (!(entity instanceof ServerPlayer player) || variables.syncQueued) return;
        MinecraftServer server = entity.m_20194_();
        if (server == null) return;
        variables.syncQueued = true;
        boolean schedule = false;
        synchronized (PENDING) {
            LinkedHashMap<ServerPlayer, EvenMoreMagicModVariables.PlayerVariables> players = PENDING.get(server);
            if (players == null) {
                players = new LinkedHashMap<>();
                PENDING.put(server, players);
                schedule = true;
            }
            players.put(player, variables);
        }
        if (schedule) server.m_6937_(server.m_6681_(() -> flush(server)));
    }

    private static void flush(MinecraftServer server) {
        LinkedHashMap<ServerPlayer, EvenMoreMagicModVariables.PlayerVariables> players;
        synchronized (PENDING) {
            players = PENDING.remove(server);
            if (players != null) for (EvenMoreMagicModVariables.PlayerVariables variables : players.values()) variables.syncQueued = false;
        }
        if (players == null) return;
        for (Map.Entry<ServerPlayer, EvenMoreMagicModVariables.PlayerVariables> entry : players.entrySet()) {
            ServerPlayer player = entry.getKey();
            if (player != null && !player.m_213877_()) entry.getValue().syncPlayerVariablesNow(player);
        }
    }
}
