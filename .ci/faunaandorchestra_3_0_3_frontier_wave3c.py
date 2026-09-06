#!/usr/bin/env python3
from __future__ import annotations

import json
import sys
from pathlib import Path

root = Path(sys.argv[1]).resolve() if len(sys.argv) > 1 else Path.cwd()
path = root / "src/main/java/net/migueel26/faunaandorchestra/entity/goals/BeaverBuildsDamGoal.java"
text = path.read_text(encoding="utf-8")

# Wave 3C sits on accepted Wave 3A only. Wave 3B is deliberately rejected.
old = '''    private BlockPos findClosestAptWater() {
        for (BlockPos candidate : BlockPos.withinManhattan(beaver.blockPosition(), 20, 3, 20)) {
            if (isWaterApt(candidate)) {
                return candidate.immutable();
            }
        }
        return null;
    }
'''

new = '''    private BlockPos findClosestAptWater() {
        BlockPos center = beaver.blockPosition();
        int centerX = center.getX();
        int centerY = center.getY();
        int centerZ = center.getZ();
        BlockPos.MutableBlockPos candidate = new BlockPos.MutableBlockPos();

        // Exact BlockPos.withinManhattan(20, 3, 20) traversal order without
        // the AbstractIterator state machine. Native QA compares every emitted
        // coordinate against vanilla before the stress profile is accepted.
        for (int depth = 0; depth <= 43; depth++) {
            int maxX = Math.min(20, depth);
            for (int dx = -maxX; dx <= maxX; dx++) {
                int maxY = Math.min(3, depth - Math.abs(dx));
                for (int dy = -maxY; dy <= maxY; dy++) {
                    int dz = depth - Math.abs(dx) - Math.abs(dy);
                    if (dz > 20) {
                        continue;
                    }

                    candidate.set(centerX + dx, centerY + dy, centerZ + dz);
                    if (isWaterApt(candidate)) {
                        return candidate.immutable();
                    }

                    if (dz != 0) {
                        candidate.setZ(centerZ - dz);
                        if (isWaterApt(candidate)) {
                            return candidate.immutable();
                        }
                    }
                }
            }
        }
        return null;
    }
'''

if text.count(old) != 1:
    raise SystemExit("Wave 3C accepted Wave-3A method shape changed unexpectedly")
text = text.replace(old, new, 1)

checks = {
    "no_vanilla_iterator_in_production_search": "for (BlockPos candidate : BlockPos.withinManhattan" not in text,
    "max_depth_43_preserved": "for (int depth = 0; depth <= 43; depth++)" in text,
    "x_radius_20_preserved": "int maxX = Math.min(20, depth);" in text,
    "y_radius_3_preserved": "int maxY = Math.min(3, depth - Math.abs(dx));" in text,
    "z_radius_20_preserved": "if (dz > 20)" in text,
    "positive_z_first": "centerZ + dz" in text and "centerZ - dz" in text,
    "mirror_only_nonzero": "if (dz != 0)" in text,
    "cooldown_preserved": "public static final int DEFAULT_COOLDOWN = 200;" in text and "public static final int INITIAL_DEFAULT_COOLDOWN = 200;" in text,
    "water_predicate_preserved": "level.getBlockState(pred).is(Blocks.WATER)" in text and "level.getBlockState(pred.above()).is(Blocks.AIR)" in text,
}
if not all(checks.values()):
    raise SystemExit(f"Wave 3C invariant failed: {checks}")

path.write_text(text, encoding="utf-8", newline="\n")
report = {
    "target": "Fauna & Orchestra 3.0.3 native-profiled frontier Wave 3C",
    "baseline_wave": "accepted Wave 3A; rejected Wave 3B is not applied",
    "changed_files": [str(path.relative_to(root))],
    "optimization": "Replace only BlockPos.withinManhattan/AbstractIterator traversal bookkeeping with direct loops that emit the same Manhattan-depth/y/x/+z/-z sequence.",
    "runtime_gate": "QA harness must compare every emitted coordinate in exact order against BlockPos.withinManhattan(center, 20, 3, 20) before scene setup.",
    "preserved": [
        "horizontal radius 20",
        "vertical radius 3",
        "nearest Manhattan depth ordering",
        "same-y x ordering",
        "positive-z before negative-z mirroring",
        "water + air predicate",
        "every-tick retry behavior after cooldown expiry",
        "200-tick build cooldown",
        "all build timing, navigation, sounds, particles and content",
    ],
    "invariants": checks,
}
(root / "FRONTIER-WAVE3C-REPORT.json").write_text(json.dumps(report, indent=2) + "\n", encoding="utf-8")
print(json.dumps(report, indent=2))
