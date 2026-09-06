#!/usr/bin/env python3
from __future__ import annotations

import json
import sys
from pathlib import Path

root = Path(sys.argv[1]).resolve() if len(sys.argv) > 1 else Path.cwd()
path = root / "src/main/java/net/migueel26/faunaandorchestra/entity/goals/BeaverBuildsDamGoal.java"
text = path.read_text(encoding="utf-8")

# Wave 3B sits strictly on top of the already runtime-proven Wave 3A shape.
imports_old = '''import net.minecraft.world.level.Level;
import net.minecraft.world.level.biome.Biomes;
import net.minecraft.world.level.block.Block;
import net.minecraft.world.level.block.Blocks;
'''
imports_new = '''import net.minecraft.world.level.Level;
import net.minecraft.world.level.biome.Biomes;
import net.minecraft.world.level.block.Block;
import net.minecraft.world.level.block.Blocks;
import net.minecraft.world.level.block.state.BlockState;
import net.minecraft.world.level.chunk.LevelChunk;
'''
if text.count(imports_old) != 1:
    raise SystemExit("Wave 3B import anchor changed unexpectedly")
text = text.replace(imports_old, imports_new, 1)

fields_old = '''    protected boolean finish = false;
    protected boolean isDam = false;
    public BeaverBuildsDamGoal(BeaverEntity beaver, double speedModifier) {
'''
fields_new = '''    protected boolean finish = false;
    protected boolean isDam = false;

    // A 20-block horizontal scan can intersect at most 4x4 chunks. Reuse this
    // tiny cache only for one synchronous lookup, then clear all references.
    private static final int MAX_SCAN_CHUNKS = 16;
    private final int[] scanChunkXs = new int[MAX_SCAN_CHUNKS];
    private final int[] scanChunkZs = new int[MAX_SCAN_CHUNKS];
    private final LevelChunk[] scanChunks = new LevelChunk[MAX_SCAN_CHUNKS];
    private final BlockPos.MutableBlockPos scanAbove = new BlockPos.MutableBlockPos();
    private int scanChunkCount;

    public BeaverBuildsDamGoal(BeaverEntity beaver, double speedModifier) {
'''
if text.count(fields_old) != 1:
    raise SystemExit("Wave 3B field anchor changed unexpectedly")
text = text.replace(fields_old, fields_new, 1)

method_old = '''    private BlockPos findClosestAptWater() {
        for (BlockPos candidate : BlockPos.withinManhattan(beaver.blockPosition(), 20, 3, 20)) {
            if (isWaterApt(candidate)) {
                return candidate.immutable();
            }
        }
        return null;
    }

    private boolean isWaterApt(BlockPos pred) {
        return level.getBlockState(pred).is(Blocks.WATER)
                && level.getBlockState(pred.above()).is(Blocks.AIR);
    }
'''
method_new = '''    private BlockPos findClosestAptWater() {
        scanChunkCount = 0;
        try {
            for (BlockPos candidate : BlockPos.withinManhattan(beaver.blockPosition(), 20, 3, 20)) {
                if (isWaterApt(candidate)) {
                    return candidate.immutable();
                }
            }
            return null;
        } finally {
            clearScanChunks();
        }
    }

    private boolean isWaterApt(BlockPos pred) {
        if (!getScanBlockState(pred).is(Blocks.WATER)) {
            return false;
        }
        scanAbove.set(pred.getX(), pred.getY() + 1, pred.getZ());
        return getScanBlockState(scanAbove).is(Blocks.AIR);
    }

    /**
     * Same state as Level#getBlockState for in-height positions, but avoids
     * resolving the same LevelChunk thousands of times during one search.
     * Cache lifetime is exactly one synchronous findClosestAptWater call.
     */
    private BlockState getScanBlockState(BlockPos pos) {
        if (level.isOutsideBuildHeight(pos)) {
            return Blocks.VOID_AIR.defaultBlockState();
        }

        int chunkX = pos.getX() >> 4;
        int chunkZ = pos.getZ() >> 4;
        for (int i = 0; i < scanChunkCount; i++) {
            if (scanChunkXs[i] == chunkX && scanChunkZs[i] == chunkZ) {
                return scanChunks[i].getBlockState(pos);
            }
        }

        LevelChunk chunk = level.getChunk(chunkX, chunkZ);
        if (scanChunkCount < MAX_SCAN_CHUNKS) {
            scanChunkXs[scanChunkCount] = chunkX;
            scanChunkZs[scanChunkCount] = chunkZ;
            scanChunks[scanChunkCount] = chunk;
            scanChunkCount++;
        }
        return chunk.getBlockState(pos);
    }

    private void clearScanChunks() {
        for (int i = 0; i < scanChunkCount; i++) {
            scanChunks[i] = null;
        }
        scanChunkCount = 0;
    }
'''
if text.count(method_old) != 1:
    raise SystemExit("Wave 3B Wave-3A method shape changed unexpectedly")
text = text.replace(method_old, method_new, 1)

checks = {
    "wave3a_preserved_no_closest_stream": "findClosestMatch" not in text,
    "range_preserved": "BlockPos.withinManhattan(beaver.blockPosition(), 20, 3, 20)" in text,
    "cooldown_preserved": "public static final int DEFAULT_COOLDOWN = 200;" in text and "public static final int INITIAL_DEFAULT_COOLDOWN = 200;" in text,
    "per_call_cache_reset": "scanChunkCount = 0;\n        try" in text and "finally {\n            clearScanChunks();" in text,
    "no_cross_tick_chunk_refs": "scanChunks[i] = null;" in text,
    "chunk_cache_bound_16": "private static final int MAX_SCAN_CHUNKS = 16;" in text,
    "water_air_predicate_preserved": "getScanBlockState(pred).is(Blocks.WATER)" in text and "getScanBlockState(scanAbove).is(Blocks.AIR)" in text,
    "outside_height_semantics": "Blocks.VOID_AIR.defaultBlockState()" in text,
}
if not all(checks.values()):
    raise SystemExit(f"Wave 3B invariant failed: {checks}")

path.write_text(text, encoding="utf-8", newline="\n")
report = {
    "target": "Fauna & Orchestra 3.0.3 native-profiled frontier Wave 3B",
    "changed_files": [str(path.relative_to(root))],
    "baseline_wave": "Wave 3A",
    "optimizations": [
        "Cache LevelChunk resolution only within one Beaver water search; a 20-block horizontal search intersects at most 16 chunks.",
        "Reuse one MutableBlockPos for the above-water AIR check instead of creating BlockPos.above objects.",
        "Clear all cached LevelChunk references in finally immediately after each search, including early-return searches.",
    ],
    "preserved": [
        "20-block horizontal radius",
        "3-block vertical radius",
        "BlockPos.withinManhattan nearest-first iteration order",
        "every-tick retry behavior after cooldown expiry",
        "water + air predicate",
        "200-tick build cooldown",
        "all particles/build timing/content",
    ],
    "invariants": checks,
}
(root / "FRONTIER-WAVE3B-REPORT.json").write_text(json.dumps(report, indent=2) + "\n", encoding="utf-8")
print(json.dumps(report, indent=2))
