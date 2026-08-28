#!/usr/bin/env python3
from __future__ import annotations

import pathlib
import sys

root = pathlib.Path(sys.argv[1]).resolve()
if not root.is_dir():
    raise SystemExit(f"missing generated Paladins Java root: {root}")

# Exact historical mapping proven by the pinned Paladins 1.20.1 substrate:
#   current 1.21.1: Identifier.of(namespace, path)
#   target  1.20.1: new Identifier(namespace, path)
# This is API syntax translation only; identifiers and behavior remain unchanged.
needle = "Identifier.of("
replacement = "new Identifier("
changed_files: list[pathlib.Path] = []
replacements = 0
for path in sorted(root.rglob("*.java")):
    text = path.read_text(encoding="utf-8")
    count = text.count(needle)
    if not count:
        continue
    updated = text.replace(needle, replacement)
    path.write_text(updated, encoding="utf-8")
    changed_files.append(path)
    replacements += count

# Current Paladins 3.1.1 uses this API broadly; a tiny count means the pinned-source
# assumption changed and should be reviewed instead of silently succeeding.
if replacements < 10:
    raise SystemExit(
        f"expected broad current Identifier.of frontier, translated only {replacements} occurrences"
    )

survivors = []
for path in sorted(root.rglob("*.java")):
    for line_no, line in enumerate(path.read_text(encoding="utf-8").splitlines(), 1):
        if needle in line:
            survivors.append(f"{path.relative_to(root)}:{line_no}:{line.strip()}")
if survivors:
    raise SystemExit("Identifier.of survived 1.20.1 compatibility pass:\n" + "\n".join(survivors))

print(
    f"[Paladins 1.20.1 API] Identifier.of -> constructor: "
    f"{replacements} occurrences across {len(changed_files)} files"
)
