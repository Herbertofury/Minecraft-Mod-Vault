#!/usr/bin/env python3
from __future__ import annotations

import sys
from pathlib import Path

root = Path(sys.argv[1]).resolve() if len(sys.argv) > 1 else Path.cwd()
qa_dir = root / "src/main/java/net/migueel26/faunaandorchestra/qa"
qa_dir.mkdir(parents=True, exist_ok=True)

java = r'''package net.migueel26.faunaandorchestra.qa;

import net.migueel26.faunaandorchestra.FaunaAndOrchestra;
import net.minecraft.core.BlockPos;
import net.minecraft.resources.ResourceLocation;
import net.minecraft.server.MinecraftServer;
import net.minecraft.server.level.ServerLevel;
import net.minecraft.server.level.ServerPlayer;
import net.minecraft.world.entity.Entity;
import net.minecraft.world.entity.EntityType;
import net.minecraft.world.entity.Mob;
import net.minecraft.world.entity.MobSpawnType;
import net.minecraft.world.level.GameRules;
import net.minecraft.world.level.block.Block;
import net.minecraft.world.level.block.Blocks;
import net.minecraftforge.event.TickEvent;
import net.minecraftforge.eventbus.api.SubscribeEvent;
import net.minecraftforge.fml.common.Mod;
import net.minecraftforge.registries.ForgeRegistries;

import java.util.LinkedHashMap;
import java.util.Map;

/** CI-only deterministic native performance scene. Never packaged in the release jar. */
@Mod.EventBusSubscriber(modid = FaunaAndOrchestra.MOD_ID)
public final class FaunaPerfQaHarness {
    private static final String PREFIX = "[FAUNA-PERF-QA] ";
    private static final int PLATFORM_Y = 200;
    private static boolean prepared;
    private static long readyGameTime = -1L;
    private static boolean serverProfilerAttempted;

    private FaunaPerfQaHarness() {}

    @SubscribeEvent
    public static void onServerTick(TickEvent.ServerTickEvent event) {
        if (event.phase != TickEvent.Phase.END || !Boolean.getBoolean("fauna.perfQa")) return;

        MinecraftServer server = event.getServer();
        if (server.getPlayerList().getPlayers().isEmpty()) return;
        ServerLevel level = server.overworld();
        ServerPlayer player = server.getPlayerList().getPlayers().get(0);

        if (!prepared) {
            prepareScene(level, player);
            prepared = true;
            readyGameTime = level.getGameTime();
            long count = 0L;
            for (Entity ignored : level.getAllEntities()) count++;
            FaunaAndOrchestra.LOGGER.info(PREFIX + "SCENE_READY gameTime={} entities={}", readyGameTime, count);
            return;
        }

        long age = level.getGameTime() - readyGameTime;
        if (!serverProfilerAttempted && age >= 200) {
            serverProfilerAttempted = true;
            if (server.getCommands().getDispatcher().getRoot().getChild("spark") != null) {
                String command = "spark profiler start --timeout 45 --thread * --save-to-file";
                try {
                    server.getCommands().performPrefixedCommand(server.createCommandSourceStack().withPermission(4), command);
                    FaunaAndOrchestra.LOGGER.info(PREFIX + "SERVER_SPARK_COMMAND_ISSUED {}", command);
                } catch (Throwable t) {
                    FaunaAndOrchestra.LOGGER.error(PREFIX + "SERVER_SPARK_COMMAND_FAILED", t);
                }
            } else {
                FaunaAndOrchestra.LOGGER.warn(PREFIX + "SERVER_SPARK_COMMAND_UNAVAILABLE");
            }
        }

        if (age == 400 || age == 800 || age == 1200 || age == 1600 || age == 2200) {
            long count = 0L;
            for (Entity ignored : level.getAllEntities()) count++;
            FaunaAndOrchestra.LOGGER.info(PREFIX + "HEARTBEAT age={} entities={}", age, count);
        }
    }

    private static void prepareScene(ServerLevel level, ServerPlayer player) {
        level.getGameRules().getRule(GameRules.RULE_DOMOBSPAWNING).set(false, level.getServer());
        level.getGameRules().getRule(GameRules.RULE_DAYLIGHT).set(false, level.getServer());
        level.setDayTime(6000L);
        level.setWeatherParameters(1000000, 0, false, false);

        BlockPos origin = new BlockPos(0, PLATFORM_Y, 0);
        for (int x = -34; x <= 34; x++) {
            for (int z = -34; z <= 34; z++) {
                level.setBlock(origin.offset(x, 0, z), Blocks.SMOOTH_STONE.defaultBlockState(), 2);
            }
        }

        player.setGameMode(net.minecraft.world.level.GameType.CREATIVE);
        player.teleportTo(level, 0.5, PLATFORM_Y + 2.0, 30.5, 180.0F, 12.0F);
        player.setInvulnerable(true);

        Map<String, Integer> entities = new LinkedHashMap<>();
        entities.put("the_great_composer", 8);
        entities.put("anya_ghost", 12);
        entities.put("quirky_frog", 36);
        entities.put("living_music", 36);
        entities.put("wandering_note_entity", 36);
        entities.put("floating_blossom", 12);
        entities.put("singing_sproutling", 24);
        entities.put("wandering_koala", 18);
        entities.put("worker_koala", 12);
        entities.put("tailor_koala", 12);
        entities.put("farmer_koala", 12);
        entities.put("melomancer_koala", 12);
        entities.put("faust", 10);
        entities.put("dan_b", 10);
        entities.put("beaver", 16);
        entities.put("red_panda", 16);
        entities.put("mantis", 12);

        int ordinal = 0;
        for (Map.Entry<String, Integer> entry : entities.entrySet()) {
            for (int i = 0; i < entry.getValue(); i++) {
                int col = ordinal % 18;
                int row = ordinal / 18;
                double x = -17.0 + col * 2.0;
                double z = 20.0 - (row % 18) * 2.0;
                spawn(level, entry.getKey(), x + 0.5, PLATFORM_Y + 1.0, z + 0.5, ordinal);
                ordinal++;
            }
        }

        String[] blockIds = {
                "mailbox", "beaver_statue", "flora_enhancer", "sewing_machine", "composer_gravestone", "mother_statue"
        };
        int placed = 0;
        for (int row = 0; row < 5; row++) {
            for (int col = 0; col < 12; col++) {
                String id = blockIds[(row * 12 + col) % blockIds.length];
                Block block = ForgeRegistries.BLOCKS.getValue(ResourceLocation.fromNamespaceAndPath(FaunaAndOrchestra.MOD_ID, id));
                if (block == null || block == Blocks.AIR) {
                    FaunaAndOrchestra.LOGGER.warn(PREFIX + "MISSING_BLOCK {}", id);
                    continue;
                }
                BlockPos pos = origin.offset(-22 + col * 4, 1 + row * 2, -18);
                level.setBlock(pos, block.defaultBlockState(), 3);
                placed++;
            }
        }

        FaunaAndOrchestra.LOGGER.info(PREFIX + "SPAWN_SUMMARY requestedEntities={} placedBlocks={}", ordinal, placed);
    }

    private static void spawn(ServerLevel level, String id, double x, double y, double z, int ordinal) {
        EntityType<?> type = ForgeRegistries.ENTITY_TYPES.getValue(ResourceLocation.fromNamespaceAndPath(FaunaAndOrchestra.MOD_ID, id));
        if (type == null) {
            FaunaAndOrchestra.LOGGER.warn(PREFIX + "MISSING_ENTITY_TYPE {}", id);
            return;
        }
        Entity entity = type.create(level);
        if (entity == null) {
            FaunaAndOrchestra.LOGGER.warn(PREFIX + "CREATE_FAILED {}", id);
            return;
        }
        entity.moveTo(x, y, z, (ordinal * 29) % 360, 0.0F);
        entity.setInvulnerable(true);
        if (entity instanceof Mob mob) {
            try {
                mob.finalizeSpawn(level, level.getCurrentDifficultyAt(BlockPos.containing(x, y, z)), MobSpawnType.COMMAND, null, null);
            } catch (Throwable t) {
                FaunaAndOrchestra.LOGGER.warn(PREFIX + "FINALIZE_FAILED {}: {}", id, t.toString());
            }
            mob.setPersistenceRequired();
        }
        level.addFreshEntity(entity);
    }
}
'''

path = qa_dir / "FaunaPerfQaHarness.java"
path.write_text(java, encoding="utf-8", newline="\n")
print(path)
