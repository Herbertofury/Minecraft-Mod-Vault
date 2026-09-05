package com.bert.slrsharedmana;

import com.mojang.brigadier.CommandDispatcher;
import net.minecraft.ChatFormatting;
import net.minecraft.commands.CommandSourceStack;
import net.minecraft.network.chat.Component;
import net.minecraft.server.level.ServerPlayer;

public final class Commands {
    private Commands() {}

    public static void register(CommandDispatcher<CommandSourceStack> dispatcher) {
        dispatcher.register(net.minecraft.commands.Commands.literal("slrmana")
                .then(net.minecraft.commands.Commands.literal("status").executes(ctx -> status(ctx.getSource()))));
    }

    private static int status(CommandSourceStack source) {
        ServerPlayer player;
        try {
            player = source.getPlayerOrException();
        } catch (Exception e) {
            source.sendFailure(Component.literal("/slrmana status must be run by a player."));
            return 0;
        }
        double current = SlrAccess.current(player);
        double max = SlrAccess.max(player);
        double ratio = BridgeConfig.SLR_MP_PER_IRON_MANA.get();
        String botany = BotanyCompat.detectedMode();
        Component line = Component.literal("SLR Shared Mana: ").withStyle(ChatFormatting.AQUA)
                .append(Component.literal(BridgeConfig.ENABLED.get() ? "ACTIVE" : "DISABLED")
                        .withStyle(BridgeConfig.ENABLED.get() ? ChatFormatting.GREEN : ChatFormatting.RED))
                .append(Component.literal(String.format(" | SLR %.1f/%.1f MP | Iron %.1f mana | ratio %.2f:1 | Botany %s",
                        current, max, ratio > 0 ? current / ratio : 0.0D, ratio, botany)).withStyle(ChatFormatting.GRAY));
        source.sendSuccess(() -> line, false);
        return 1;
    }
}
