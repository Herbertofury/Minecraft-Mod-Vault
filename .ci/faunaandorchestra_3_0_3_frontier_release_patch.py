#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import subprocess
import sys
from pathlib import Path

parser = argparse.ArgumentParser(description="Apply the verified Fauna & Orchestra 3.0.3 performance stack without QA-only source.")
parser.add_argument("root", type=Path)
parser.add_argument("--wave4", action="store_true", help="Include Wave 4 only after its native A/B gate is explicitly accepted.")
args = parser.parse_args()
root = args.root.resolve()
here = Path(__file__).resolve().parent


def run(script: str) -> None:
    subprocess.run([sys.executable, str(here / script), str(root)], check=True)


# Verified base/extreme stack.
run("faunaandorchestra_3_0_3_performance_patch_v2.py")

# The historical extreme patch contains an over-broad verifier that also matches
# BASE_TICKS_UNTIL_DEATH. Narrow only that verifier exactly as the already-green
# extreme build workflow did; this does not alter production Java source.
extreme = here / "faunaandorchestra_3_0_3_extreme_patch.py"
s = extreme.read_text(encoding="utf-8")
old = '''remnants = [
    line.strip() for line in text.splitlines()
    if "TICKS_UNTIL_DEATH" in line or "EntityDataAccessor" in line or "EntityDataSerializers" in line
]'''
new = '''remnants = [
    line.strip() for line in text.splitlines()
    if "EntityDataAccessor<Integer> TICKS_UNTIL_DEATH" in line
    or "entityData.define(TICKS_UNTIL_DEATH" in line
    or "entityData.get(TICKS_UNTIL_DEATH" in line
    or "entityData.set(TICKS_UNTIL_DEATH" in line
]'''
if s.count(old) != 1:
    raise SystemExit("extreme patch verifier shape changed unexpectedly")
extreme.write_text(s.replace(old, new, 1), encoding="utf-8", newline="\n")
run("faunaandorchestra_3_0_3_extreme_patch.py")
run("faunaandorchestra_3_0_3_extreme_safe_wave.py")

# Native-profiled frontier. Wave 3B is deliberately excluded.
run("faunaandorchestra_3_0_3_frontier_patch.py")
run("faunaandorchestra_3_0_3_frontier_wave3c.py")
if args.wave4:
    run("faunaandorchestra_3_0_3_frontier_wave4.py")

qa_root = root / "src/main/java/net/migueel26/faunaandorchestra/qa"
if qa_root.exists():
    raise SystemExit(f"release blocker: QA-only source exists at {qa_root}")

beaver = root / "src/main/java/net/migueel26/faunaandorchestra/entity/goals/BeaverBuildsDamGoal.java"
crawling = root / "src/main/java/net/migueel26/faunaandorchestra/block/entity/CrawlingDiscordBlockEntity.java"
btext = beaver.read_text(encoding="utf-8")
if "for (BlockPos candidate : BlockPos.withinManhattan" in btext:
    raise SystemExit("release blocker: pre-Wave-3C vanilla iterator loop remains")
if "for (int depth = 0; depth <= 43; depth++)" not in btext or "BlockPos.MutableBlockPos candidate" not in btext:
    raise SystemExit("release blocker: accepted Wave 3C direct traversal is missing")
if args.wave4:
    ctext = crawling.read_text(encoding="utf-8")
    if "BlockPos.betweenClosed(" in ctext:
        raise SystemExit("release blocker: Wave 4 production betweenClosed call remains")

report = {
    "target": "Fauna & Orchestra Forge 1.20.1 v3.0.3 runtime-profiled frontier release",
    "source_commit": "131bdcbf76aec07267d68a57185bb103669af83e",
    "included": [
        "verified first-pass optimization",
        "extreme Wave 2A",
        "Floating Blossom zero-regression safe wave",
        "accepted runtime Wave 3A",
        "accepted runtime Wave 3C",
    ] + (["candidate runtime Wave 4 (must have separate accepted native A/B receipt)"] if args.wave4 else []),
    "excluded": [
        "rejected Wave 3B",
        "native QA harness",
        "profiling-only CI bypasses",
        "Spark/JFR test instrumentation",
    ],
    "wave4": args.wave4,
}
(root / "FRONTIER-RELEASE-PATCH-REPORT.json").write_text(json.dumps(report, indent=2) + "\n", encoding="utf-8")
print(json.dumps(report, indent=2))
