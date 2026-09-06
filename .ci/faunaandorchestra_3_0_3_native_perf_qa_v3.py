#!/usr/bin/env python3
from __future__ import annotations

import subprocess
import sys
from pathlib import Path

root = Path(sys.argv[1]).resolve() if len(sys.argv) > 1 else Path.cwd()
base = Path(__file__).with_name("faunaandorchestra_3_0_3_native_perf_qa_v2.py")
subprocess.run([sys.executable, str(base), str(root)], check=True)

path = root / "src/main/java/net/migueel26/faunaandorchestra/event/CITestHandler.java"
text = path.read_text(encoding="utf-8")
old = '        if (System.getenv("CI") != null) {\n'
new = '        if (System.getenv("CI") != null && !Boolean.getBoolean("fauna.perfQa")) {\n'
if text.count(old) != 1:
    raise SystemExit("CITestHandler CI guard shape changed unexpectedly")
text = text.replace(old, new, 1)
if text.count('!Boolean.getBoolean("fauna.perfQa")') != 1:
    raise SystemExit("native perf QA CI bypass was not applied exactly once")
path.write_text(text, encoding="utf-8", newline="\n")
print(path)
