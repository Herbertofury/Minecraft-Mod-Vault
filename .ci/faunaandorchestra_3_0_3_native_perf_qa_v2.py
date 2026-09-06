#!/usr/bin/env python3
from __future__ import annotations

import subprocess
import sys
from pathlib import Path

# Frontier Wave 3C A/B driver: accepted Wave 3A + candidate Wave 3C.
# Rejected Wave 3B is intentionally NOT applied.
root = Path(sys.argv[1]).resolve() if len(sys.argv) > 1 else Path.cwd()
base = Path(__file__).with_name("faunaandorchestra_3_0_3_native_perf_qa.py")
subprocess.run([sys.executable, str(base), str(root)], check=True)
frontier = Path(__file__).with_name("faunaandorchestra_3_0_3_frontier_patch.py")
subprocess.run([sys.executable, str(frontier), str(root)], check=True)
frontier_wave3c = Path(__file__).with_name("faunaandorchestra_3_0_3_frontier_wave3c.py")
subprocess.run([sys.executable, str(frontier_wave3c), str(root)], check=True)

path = root / "src/main/java/net/migueel26/faunaandorchestra/qa/FaunaPerfQaHarness.java"
text = path.read_text(encoding="utf-8")
text = text.replace("    private static boolean serverProfilerAttempted;\n", "", 1)
block = '''        if (!serverProfilerAttempted && age >= 200) {
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

'''
if text.count(block) != 1:
    raise SystemExit("native QA Spark server-command block changed unexpectedly")
text = text.replace(block, "", 1)
if "serverProfilerAttempted" in text or "SERVER_SPARK_COMMAND" in text:
    raise SystemExit("server Spark profiling remnants remain")

prepare_anchor = '''    private static void prepareScene(ServerLevel level, ServerPlayer player) {
        level.getGameRules().getRule(GameRules.RULE_DOMOBSPAWNING).set(false, level.getServer());
'''
verify_block = '''    private static void verifyManhattanOrder() {
        BlockPos center = new BlockPos(123, 64, -456);
        java.util.Iterator<BlockPos> vanilla = BlockPos.withinManhattan(center, 20, 3, 20).iterator();
        int index = 0;

        for (int depth = 0; depth <= 43; depth++) {
            int maxX = Math.min(20, depth);
            for (int dx = -maxX; dx <= maxX; dx++) {
                int maxY = Math.min(3, depth - Math.abs(dx));
                for (int dy = -maxY; dy <= maxY; dy++) {
                    int dz = depth - Math.abs(dx) - Math.abs(dy);
                    if (dz > 20) {
                        continue;
                    }

                    assertVanillaPosition(vanilla, center.getX() + dx, center.getY() + dy, center.getZ() + dz, index++);
                    if (dz != 0) {
                        assertVanillaPosition(vanilla, center.getX() + dx, center.getY() + dy, center.getZ() - dz, index++);
                    }
                }
            }
        }

        if (vanilla.hasNext()) {
            BlockPos extra = vanilla.next();
            throw new IllegalStateException("Wave 3C order verifier ended early at index " + index + "; vanilla next=" + extra);
        }
        if (index != 11767) {
            throw new IllegalStateException("Wave 3C order verifier count mismatch: " + index + " != 11767");
        }
        FaunaAndOrchestra.LOGGER.info(PREFIX + "MANHATTAN_ORDER_EQUIVALENCE_PASS positions={}", index);
    }

    private static void assertVanillaPosition(java.util.Iterator<BlockPos> vanilla, int x, int y, int z, int index) {
        if (!vanilla.hasNext()) {
            throw new IllegalStateException("Wave 3C order verifier exhausted vanilla at index " + index
                    + " expected=" + x + "," + y + "," + z);
        }
        BlockPos actual = vanilla.next();
        if (actual.getX() != x || actual.getY() != y || actual.getZ() != z) {
            throw new IllegalStateException("Wave 3C order mismatch at index " + index
                    + " expected=" + x + "," + y + "," + z + " actual=" + actual);
        }
    }

    private static void prepareScene(ServerLevel level, ServerPlayer player) {
        verifyManhattanOrder();
        level.getGameRules().getRule(GameRules.RULE_DOMOBSPAWNING).set(false, level.getServer());
'''
if text.count(prepare_anchor) != 1:
    raise SystemExit("native QA prepareScene anchor changed unexpectedly")
text = text.replace(prepare_anchor, verify_block, 1)
if text.count("MANHATTAN_ORDER_EQUIVALENCE_PASS") != 1 or text.count("verifyManhattanOrder();") != 1:
    raise SystemExit("Wave 3C order verifier injection failed")
path.write_text(text, encoding="utf-8", newline="\n")
print(path)

ci_path = root / "src/main/java/net/migueel26/faunaandorchestra/event/CITestHandler.java"
ci_text = ci_path.read_text(encoding="utf-8")
ci_old = '        if (System.getenv("CI") != null) {\n'
ci_new = '        if (System.getenv("CI") != null && !Boolean.getBoolean("fauna.perfQa")) {\n'
if ci_text.count(ci_old) != 1:
    raise SystemExit("CITestHandler CI guard shape changed unexpectedly")
ci_text = ci_text.replace(ci_old, ci_new, 1)
if ci_text.count('!Boolean.getBoolean("fauna.perfQa")') != 1:
    raise SystemExit("native perf QA CI bypass was not applied exactly once")
ci_path.write_text(ci_text, encoding="utf-8", newline="\n")
print(ci_path)
