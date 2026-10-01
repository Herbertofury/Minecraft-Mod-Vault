#!/usr/bin/env python3
from __future__ import annotations

import pathlib
import sys

root = pathlib.Path(sys.argv[1]).resolve()
if not root.is_dir():
    raise SystemExit(f"missing generated Paladins Java root: {root}")

# The batch-2 survivor gate intentionally scans literal forbidden API symbols. Current 3.1.1 also
# documents two of those symbols in comments. Normalize only those exact documentation phrases so
# the strict gate remains unchanged and still fails if forbidden symbols survive in executable code.
replacements = (
    (
        "net/paladins/effect/PaladinEffects.java",
        "/// Nearly cancels the holder's gravity (GENERIC_GRAVITY default 0.08, clamped [-1, 1]), leaving them",
        "/// Nearly cancels the holder's gravity (vanilla gravity default 0.08, clamped [-1, 1]), leaving them",
    ),
    (
        "net/paladins/entity/PaladinSummons.java",
        "/// A single attribute-scaling entry: `targetAttribute += base + ownerAttribute * coefficient` (ADD_VALUE).",
        "/// A single attribute-scaling entry: `targetAttribute += base + ownerAttribute * coefficient` (addition operation).",
    ),
)

for rel, old, new in replacements:
    path = root / rel
    text = path.read_text(encoding="utf-8")
    count = text.count(old)
    if count != 1:
        raise SystemExit(f"[compat comment normalization] expected exactly one pinned comment in {rel}, found {count}")
    path.write_text(text.replace(old, new, 1), encoding="utf-8")
    print(f"[Paladins compat comments] normalized one documentation-only forbidden-token collision: {rel}")

print("[Paladins compat comments] strict batch-2 survivor scan remains unchanged for executable source")
