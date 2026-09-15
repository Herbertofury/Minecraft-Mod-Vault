package net.mcreator.protectionpixel.repair;

import java.lang.ref.WeakReference;
import java.util.LinkedHashMap;
import java.util.LinkedHashSet;
import java.util.Map;
import net.minecraft.client.multiplayer.ClientLevel;
import net.minecraft.core.BlockPos;
import net.minecraft.core.SectionPos;
import net.minecraft.world.level.block.entity.BlockEntity;
import net.minecraft.world.level.chunk.LevelChunk;
import net.mcreator.protectionpixel.init.ProtectionPixelModBlocks;

/**
 * Render-thread-only cache for Protection Pixel's armor preview target discovery.
 *
 * The upstream renderer scans every block entity in every client chunk inside
 * render distance every rendered frame. Block/chunk membership changes on game
 * ticks, while rendering can run many times per tick. This keeps the exact scan
 * radius/order and memoizes only target positions for the current game tick and
 * player chunk. Block entities are re-resolved for each rendered frame so the
 * cache never owns long-lived world/block-entity references and same-tick removals
 * or replacements at an already-known position are observed immediately.
 */
public final class ArmorRenderTargetCache {
    private static final LinkedHashSet<BlockPos> TARGET_POSITIONS = new LinkedHashSet<>();
    private static WeakReference<ClientLevel> cachedLevelRef = new WeakReference<>(null);
    private static long cachedGameTime = Long.MIN_VALUE;
    private static int cachedCenterChunkX = Integer.MIN_VALUE;
    private static int cachedCenterChunkZ = Integer.MIN_VALUE;
    private static int cachedRadius = Integer.MIN_VALUE;

    private ArmorRenderTargetCache() {}

    public static Map<BlockPos, BlockEntity> targets(ClientLevel level, BlockPos center, int radius) {
        int centerChunkX = SectionPos.m_123171_(center.m_123341_());
        int centerChunkZ = SectionPos.m_123171_(center.m_123343_());
        long gameTime = level.m_46467_();
        ClientLevel cachedLevel = cachedLevelRef.get();

        if (level != cachedLevel || gameTime != cachedGameTime ||
                centerChunkX != cachedCenterChunkX || centerChunkZ != cachedCenterChunkZ ||
                radius != cachedRadius) {
            rebuildPositions(level, center, radius);
            cachedLevelRef = new WeakReference<>(level);
            cachedGameTime = gameTime;
            cachedCenterChunkX = centerChunkX;
            cachedCenterChunkZ = centerChunkZ;
            cachedRadius = radius;
        }

        // Resolve only the already-filtered target positions for this frame. This
        // preserves live block-entity state without repeating the render-distance scan.
        LinkedHashMap<BlockPos, BlockEntity> liveTargets = new LinkedHashMap<>(Math.max(4, TARGET_POSITIONS.size() * 2));
        Object armorHanger = ProtectionPixelModBlocks.ARMORHANGER.get();
        Object armorLoadPlatform = ProtectionPixelModBlocks.ARMORLOADPLATFORM.get();
        for (BlockPos pos : TARGET_POSITIONS) {
            BlockEntity blockEntity = level.m_7702_(pos);
            if (blockEntity == null) {
                continue;
            }
            Object block = blockEntity.m_58900_().m_60734_();
            if (block == armorHanger || block == armorLoadPlatform) {
                liveTargets.put(pos, blockEntity);
            }
        }
        return liveTargets;
    }

    private static void rebuildPositions(ClientLevel level, BlockPos center, int radius) {
        TARGET_POSITIONS.clear();
        Object armorHanger = ProtectionPixelModBlocks.ARMORHANGER.get();
        Object armorLoadPlatform = ProtectionPixelModBlocks.ARMORLOADPLATFORM.get();

        // Deliberately preserve upstream's chunk traversal order: Z outer, X inner.
        for (int chunkZ = -radius; chunkZ <= radius; ++chunkZ) {
            for (int chunkX = -radius; chunkX <= radius; ++chunkX) {
                LevelChunk chunk = level.m_6325_(
                        SectionPos.m_123171_(center.m_123341_() + (chunkX << 4)),
                        SectionPos.m_123171_(center.m_123343_() + (chunkZ << 4)));
                if (chunk == null) {
                    continue;
                }
                for (Map.Entry<BlockPos, BlockEntity> entry : chunk.m_62954_().entrySet()) {
                    BlockEntity blockEntity = entry.getValue();
                    Object block = blockEntity.m_58900_().m_60734_();
                    if (block == armorHanger || block == armorLoadPlatform) {
                        TARGET_POSITIONS.add(entry.getKey());
                    }
                }
            }
        }
    }
}
