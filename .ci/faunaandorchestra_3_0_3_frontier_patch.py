#!/usr/bin/env python3
from __future__ import annotations

import json
import sys
from pathlib import Path

root = Path(sys.argv[1]).resolve() if len(sys.argv) > 1 else Path.cwd()
path = root / "src/main/java/net/migueel26/faunaandorchestra/entity/goals/BeaverBuildsDamGoal.java"
text = path.read_text(encoding="utf-8")

text = text.replace("import java.util.Optional;\n", "", 1)

old = '''    private Pair<Vec3, BlockPos> getDamPosition() {
        BlockPos pathBlock = null;
        Optional<BlockPos> water = BlockPos.findClosestMatch(beaver.blockPosition(), 20, 3, this::isWaterApt);
        Optional<BlockPos> dam = BlockPos.findClosestMatch(beaver.blockPosition(), 20, 3, pred -> level.getBlockState(pred).getBlock() == ModBlocks.DAM_BLOCK.get());
        BlockPos waterPos = null;
        if (water.isPresent()
            //&& level.getBiome(water.get()).getKey() == Biomes.RIVER
        ) {
            int x, y, z;

            // Default to water
            waterPos = water.get().immutable();
            x = waterPos.getX();
            y = waterPos.getY();
            z = waterPos.getZ();
            isDam = false;

            if (dam.isPresent()
                //&& level.getBiome(dam.get()).getKey() == Biomes.RIVER
            ) {
                //isDam = level.getRandom().nextFloat() <= 0.25;
                if (isDam) {
                    // WE (TRY TO) PLACE ON TOP OF DAM
                    waterPos = dam.get().immutable();
                    x = waterPos.getX();
                    y = waterPos.getY();
                    z = waterPos.getZ();
                }
            }

            Iterator<BlockPos> iterator = BlockPos.betweenClosed(new BlockPos(x + 1, y, z + 1), new BlockPos(x - 1, y, z - 1)).iterator();
            while (iterator.hasNext() && pathBlock == null) {
                BlockPos blockPos = iterator.next();
                if (blockPos != waterPos
                        && level.getBlockState(blockPos).getBlock() != Blocks.WATER
                        && level.getBlockState(blockPos.above()).getBlock() == Blocks.AIR) {
                    pathBlock = blockPos.immutable();
                }
            }
        }
        return pathBlock == null ? null : new Pair<>(pathBlock.getCenter(), waterPos);
    }

    private boolean isWaterApt(BlockPos pred) {
        return (level.getBlockState(pred).is(Blocks.WATER)) && (level.getBlockState(pred.above()).is(Blocks.AIR));
    }
'''

new = '''    private Pair<Vec3, BlockPos> getDamPosition() {
        BlockPos pathBlock = null;
        BlockPos waterPos = findClosestAptWater();
        if (waterPos != null) {
            int x = waterPos.getX();
            int y = waterPos.getY();
            int z = waterPos.getZ();
            isDam = false;

            Iterator<BlockPos> iterator = BlockPos.betweenClosed(new BlockPos(x + 1, y, z + 1), new BlockPos(x - 1, y, z - 1)).iterator();
            while (iterator.hasNext() && pathBlock == null) {
                BlockPos blockPos = iterator.next();
                if (blockPos != waterPos
                        && level.getBlockState(blockPos).getBlock() != Blocks.WATER
                        && level.getBlockState(blockPos.above()).getBlock() == Blocks.AIR) {
                    pathBlock = blockPos.immutable();
                }
            }
        }
        return pathBlock == null ? null : new Pair<>(pathBlock.getCenter(), waterPos);
    }

    /**
     * Equivalent to the vanilla nearest matching-position search at these bounds,
     * but avoids the stream/Optional wrapper. withinManhattan is the iterable
     * backing the vanilla closest-match stream, so nearest-first tie ordering,
     * horizontal radius 20 and vertical radius 3 are preserved exactly.
     */
    private BlockPos findClosestAptWater() {
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

if text.count(old) != 1:
    raise SystemExit("BeaverBuildsDamGoal profiled lookup shape changed unexpectedly")
text = text.replace(old, new, 1)

checks = {
    "no_findClosestMatch": "findClosestMatch" not in text,
    "no_optional_import": "java.util.Optional" not in text,
    "no_dead_dam_lookup": "Optional<BlockPos> dam" not in text and "dam.isPresent()" not in text,
    "radius_20_vertical_3_preserved": "BlockPos.withinManhattan(beaver.blockPosition(), 20, 3, 20)" in text,
    "cooldown_unchanged": "public static final int DEFAULT_COOLDOWN = 200;" in text and "public static final int INITIAL_DEFAULT_COOLDOWN = 200;" in text,
    "water_predicate_preserved": "level.getBlockState(pred).is(Blocks.WATER)" in text and "level.getBlockState(pred.above()).is(Blocks.AIR)" in text,
}
if not all(checks.values()):
    raise SystemExit(f"frontier patch invariant failed: {checks}")

path.write_text(text, encoding="utf-8", newline="\n")
report = {
    "target": "Fauna & Orchestra 3.0.3 native-profiled frontier Wave 3A",
    "changed_files": [str(path.relative_to(root))],
    "optimizations": [
        "Remove the second 20x3 dam closest-match scan, which is dead in shipped logic because isDam is reset false immediately before the only branch that could consume that result.",
        "Replace BlockPos.findClosestMatch stream/Optional plumbing for water with direct iteration over BlockPos.withinManhattan using identical 20 horizontal / 3 vertical bounds and nearest-first iteration order.",
    ],
    "invariants": checks,
}
(root / "FRONTIER-PERFORMANCE-PATCH-REPORT.json").write_text(json.dumps(report, indent=2) + "\n", encoding="utf-8")
print(json.dumps(report, indent=2))
