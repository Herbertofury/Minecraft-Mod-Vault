#!/usr/bin/env python3
from __future__ import annotations

import json
import sys
from pathlib import Path

root = Path(sys.argv[1]).resolve() if len(sys.argv) > 1 else Path.cwd()
path = root / "src/main/java/net/migueel26/faunaandorchestra/entity/goals/BeaverBuildsDamGoal.java"
text = path.read_text(encoding="utf-8")

old_decl = '''        Optional<BlockPos> water = BlockPos.findClosestMatch(beaver.blockPosition(), 20, 3, this::isWaterApt);\n        Optional<BlockPos> dam = BlockPos.findClosestMatch(beaver.blockPosition(), 20, 3, pred -> level.getBlockState(pred).getBlock() == ModBlocks.DAM_BLOCK.get());\n        BlockPos waterPos = null;\n'''
new_decl = '''        Optional<BlockPos> water = BlockPos.findClosestMatch(beaver.blockPosition(), 20, 3, this::isWaterApt);\n        BlockPos waterPos = null;\n'''
if text.count(old_decl) != 1:
    raise SystemExit("Beaver dam search declaration shape changed unexpectedly")
text = text.replace(old_decl, new_decl, 1)

old_dead = '''\n            if (dam.isPresent()\n                //&& level.getBiome(dam.get()).getKey() == Biomes.RIVER\n            ) {\n                //isDam = level.getRandom().nextFloat() <= 0.25;\n                if (isDam) {\n                    // WE (TRY TO) PLACE ON TOP OF DAM\n                    waterPos = dam.get().immutable();\n                    x = waterPos.getX();\n                    y = waterPos.getY();\n                    z = waterPos.getZ();\n                }\n            }\n'''
if text.count(old_dead) != 1:
    raise SystemExit("Beaver dead dam branch shape changed unexpectedly")
text = text.replace(old_dead, "\n", 1)

# The removed search was provably dead in this shipped source: isDam is assigned false
# immediately before the dam branch and the only assignment that could make it true is commented out.
# Keep all timing, range, water-selection order, pathing, particles, sounds and placement behavior unchanged.
if "Optional<BlockPos> dam = BlockPos.findClosestMatch" in text:
    raise SystemExit("dead Beaver dam scan remains")
if text.count("BlockPos.findClosestMatch(beaver.blockPosition(), 20, 3") != 1:
    raise SystemExit("expected exactly one Beaver closest-match scan after patch")
if "isDam = false;" not in text:
    raise SystemExit("Beaver isDam behavior changed unexpectedly")

path.write_text(text, encoding="utf-8", newline="\n")
report = {
    "patch": "runtime-frontier-beaver-dead-dam-scan",
    "target": str(path.relative_to(root)),
    "behavioral_invariants": [
        "DEFAULT_COOLDOWN remains 200",
        "water search radius remains horizontal 20 / vertical 3",
        "water predicate and closest-match ordering unchanged",
        "path selection unchanged",
        "dam placement behavior unchanged because removed dam result was unreachable while isDam=false",
        "build animation, particles and sounds unchanged",
    ],
    "removed_cost": "one full 20x3 BlockPos.findClosestMatch dam scan plus its predicate/lambda allocation per Beaver build attempt",
}
(root / "RUNTIME-FRONTIER-PATCH-REPORT.json").write_text(json.dumps(report, indent=2) + "\n", encoding="utf-8")
print(json.dumps(report, indent=2))
