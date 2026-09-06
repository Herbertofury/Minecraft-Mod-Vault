package com.herbertofury.spellbrook.compat;

import net.minecraft.core.BlockPos;
import net.minecraft.world.entity.player.Player;
import net.minecraft.world.item.ItemStack;
import net.minecraft.world.level.Level;

public final class CompatHooks {
    @FunctionalInterface public interface HexereiSpawner { boolean spawn(Level level, BlockPos pos, ItemStack stack, Player player, float yaw); }
    private static HexereiSpawner hexereiSpawner;
    private CompatHooks() {}
    public static void installHexerei(HexereiSpawner spawner) { hexereiSpawner = spawner; }
    public static void disableHexerei() { hexereiSpawner = null; }
    public static boolean trySpawnHexerei(Level level, BlockPos pos, ItemStack stack, Player player, float yaw) { return hexereiSpawner != null && hexereiSpawner.spawn(level, pos, stack, player, yaw); }
}
