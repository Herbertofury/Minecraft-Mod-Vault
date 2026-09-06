#!/usr/bin/env python3
from __future__ import annotations

import subprocess
import sys
from pathlib import Path

root = Path(sys.argv[1]).resolve() if len(sys.argv) > 1 else Path.cwd()
base = Path(__file__).with_name("faunaandorchestra_3_0_3_native_perf_qa.py")
subprocess.run([sys.executable, str(base), str(root)], check=True)

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
path.write_text(text, encoding="utf-8", newline="\n")
print(path)
