#!/usr/bin/env python3
from __future__ import annotations

import json
import sys
from pathlib import Path

ROOT = Path(sys.argv[1]).resolve() if len(sys.argv) > 1 else Path.cwd()
CHANGED: list[str] = []
NOTES: list[str] = []


def read(rel: str) -> str:
    return (ROOT / rel).read_text(encoding="utf-8")


def write(rel: str, text: str) -> None:
    path = ROOT / rel
    old = path.read_text(encoding="utf-8")
    if old == text:
        return
    path.write_text(text, encoding="utf-8", newline="\n")
    CHANGED.append(rel)


def replace_once(rel: str, old: str, new: str, note: str) -> None:
    text = read(rel)
    count = text.count(old)
    if count != 1:
        raise RuntimeError(f"{rel}: expected exactly one match, found {count}")
    write(rel, text.replace(old, new, 1))
    NOTES.append(note)


rel = "src/main/java/net/migueel26/faunaandorchestra/entity/custom/misc/FloatingBlossomEntity.java"

# 1) Position candidates drive authoritative server-side flower placement only. Keep the
# init animation on both sides, but build the 13x13 candidate list on the server only.
replace_once(
    rel,
    """        if (getLifetime() == MAX_LIFETIME) {\n            triggerAnim(\"blossom_controller\", \"init\");\n            initList();\n        } else if (getLifetime() == 40) {\n""",
    """        if (getLifetime() == MAX_LIFETIME) {\n            triggerAnim(\"blossom_controller\", \"init\");\n            if (!level().isClientSide()) {\n                initList();\n            }\n        } else if (getLifetime() == 40) {\n""",
    "Floating Blossom no longer builds its server-only flower-position candidate list on clients; init animation remains unchanged on both sides.",
)

# 2) The circle membership test is integer geometry. Squared distance produces the exact
# same positions as sqrt(pow(dx,2)+pow(dz,2)) <= 6 without pow/sqrt calls.
replace_once(
    rel,
    """        for (int i = 0; i < DIAMETER; i++) {\n            for (int j = 0; j < DIAMETER; j++) {\n                if (Math.sqrt(Math.pow(i - 6, 2) + Math.pow(j - 6, 2)) <= 6) {\n                    BlockPos pos = new BlockPos((int) (getX() + i - 6), (int) getY(), (int) (getZ() + j - 6));\n                    posList.add(pos);\n                }\n            }\n        }\n""",
    """        for (int i = 0; i < DIAMETER; i++) {\n            int dx = i - 6;\n            for (int j = 0; j < DIAMETER; j++) {\n                int dz = j - 6;\n                if (dx * dx + dz * dz <= 36) {\n                    BlockPos pos = new BlockPos((int) (getX() + dx), (int) getY(), (int) (getZ() + dz));\n                    posList.add(pos);\n                }\n            }\n        }\n""",
    "Floating Blossom circular position generation uses exact integer squared distance instead of repeated Math.pow/Math.sqrt calls.",
)

# 3) Preserve tag iteration order and flower choice set while eliminating two temporary
# toList results plus one copy/addAll chain every 10-tick growth step.
replace_once(
    rel,
    """                List<Block> tallFlowers = registry.getTag(BlockTags.TALL_FLOWERS)\n                        .map(namedTag -> namedTag.stream().map(Holder::value).toList())\n                        .orElse(List.of());\n\n                List<Block> smallFlowers = registry.getTag(BlockTags.SMALL_FLOWERS)\n                        .map(namedTag -> namedTag.stream().map(Holder::value).toList())\n                        .orElse(List.of());\n\n                List<Block> flowers = new ArrayList<>(tallFlowers);\n                flowers.addAll(smallFlowers);\n                flowers.remove(Blocks.WITHER_ROSE);\n""",
    """                List<Block> flowers = new ArrayList<>();\n                registry.getTag(BlockTags.TALL_FLOWERS)\n                        .ifPresent(namedTag -> namedTag.stream().map(Holder::value).forEach(flowers::add));\n                registry.getTag(BlockTags.SMALL_FLOWERS)\n                        .ifPresent(namedTag -> namedTag.stream().map(Holder::value).forEach(flowers::add));\n                flowers.remove(Blocks.WITHER_ROSE);\n""",
    "Floating Blossom flower-tag collection preserves the same tall-then-small iteration order while removing intermediate tag lists and copies.",
)

# 4) Vanilla cramming branch counted non-passengers into a local that was never read.
# Preserve the condition and random.nextInt(4) consumption exactly, but drop the dead
# secondary entity traversal.
replace_once(
    rel,
    """                if (i > 0 && list.size() > i - 1 && this.random.nextInt(4) == 0) {\n                    int j = 0;\n\n                    for (Entity entity : list) {\n                        if (!entity.isPassenger()) {\n                            ++j;\n                        }\n                    }\n                }\n\n                for (Entity entity1 : list) {\n""",
    """                if (i > 0 && list.size() > i - 1) {\n                    this.random.nextInt(4);\n                }\n\n                for (Entity entity1 : list) {\n""",
    "Floating Blossom removes a dead cramming-count traversal while preserving the exact conditional RNG consumption and all push behavior.",
)

report = {
    "target": "Fauna & Orchestra Forge 1.20.1 3.0.3 - extreme safe wave",
    "changed_files": sorted(set(CHANGED)),
    "optimizations": NOTES,
}
(ROOT / "EXTREME-SAFE-WAVE-REPORT.json").write_text(json.dumps(report, indent=2) + "\n", encoding="utf-8")
print(json.dumps(report, indent=2))
