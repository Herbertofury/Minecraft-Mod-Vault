#!/usr/bin/env python3
from __future__ import annotations

import pathlib
import sys

root = pathlib.Path(sys.argv[1]).resolve()
if not root.is_dir():
    raise SystemExit(f"missing generated Paladins Java root: {root}")


def replace_exact(path: pathlib.Path, old: str, new: str, label: str) -> None:
    text = path.read_text(encoding="utf-8")
    count = text.count(old)
    if count != 1:
        raise SystemExit(f"[{label}] expected exactly one pinned source shape in {path}, found {count}")
    path.write_text(text.replace(old, new, 1), encoding="utf-8")
    print(f"[Paladins 1.20.1 API] {label}: {path.relative_to(root)}")


# Exact historical mapping proven by the pinned Paladins 1.20.1 substrate:
#   current 1.21.1: Identifier.of(...)
#   target  1.20.1: new Identifier(...)
# This is API syntax translation only; identifiers and behavior remain unchanged.
identifier_needle = "Identifier.of("
identifier_replacement = "new Identifier("
identifier_files: list[pathlib.Path] = []
identifier_replacements = 0
for path in sorted(root.rglob("*.java")):
    text = path.read_text(encoding="utf-8")
    count = text.count(identifier_needle)
    if not count:
        continue
    path.write_text(text.replace(identifier_needle, identifier_replacement), encoding="utf-8")
    identifier_files.append(path)
    identifier_replacements += count

# Current Paladins 3.1.1 uses this API broadly; a tiny count means the pinned-source
# assumption changed and should be reviewed instead of silently succeeding.
if identifier_replacements < 10:
    raise SystemExit(
        f"expected broad current Identifier.of frontier, translated only {identifier_replacements} occurrences"
    )

# 1.21 introduced DataTracker.Builder entity initialization. The pinned 1.20.1
# Paladins BarrierEntity proves the exact equivalent is no-arg initDataTracker()
# followed by startTracking on the entity tracker. Preserve the same three keys/defaults.
barrier = root / "net/paladins/entity/BarrierEntity.java"
replace_exact(
    barrier,
    '''    @Override\n    protected void initDataTracker(DataTracker.Builder builder) {\n        builder.add(SPELL_ID_TRACKER, "");\n        builder.add(OWNER_ID_TRACKER, 0);\n        builder.add(TIME_TO_LIVE_TRACKER, 0);\n    }\n''',
    '''    @Override\n    protected void initDataTracker() {\n        this.getDataTracker().startTracking(SPELL_ID_TRACKER, "");\n        this.getDataTracker().startTracking(OWNER_ID_TRACKER, 0);\n        this.getDataTracker().startTracking(TIME_TO_LIVE_TRACKER, 0);\n    }\n''',
    "BarrierEntity DataTracker.Builder -> 1.20.1 startTracking",
)

# 1.21 moved/reworked block tooltip context. The pinned historical Paladins block
# gives the exact 1.20.1 override shape; only API plumbing changes, the hint text stays current.
workbench = root / "net/paladins/block/MonkWorkbenchBlock.java"
replace_exact(
    workbench,
    "import net.minecraft.item.Item;\n",
    "import net.minecraft.client.item.TooltipContext;\n",
    "MonkWorkbench tooltip context import",
)
replace_exact(
    workbench,
    "import net.minecraft.item.tooltip.TooltipType;\n",
    "",
    "MonkWorkbench remove post-1.20.1 TooltipType",
)
replace_exact(
    workbench,
    '''    public void appendTooltip(ItemStack stack, Item.TooltipContext context, List<Text> tooltip, TooltipType options) {\n        super.appendTooltip(stack, context, tooltip, options);\n''',
    '''    public void appendTooltip(ItemStack stack, @Nullable BlockView world, List<Text> tooltip, TooltipContext options) {\n        super.appendTooltip(stack, world, tooltip, options);\n''',
    "MonkWorkbench 1.20.1 appendTooltip signature",
)

# Fail closed on the exact first-frontier symbols so this transform cannot silently regress.
survivors: list[str] = []
for path in sorted(root.rglob("*.java")):
    for line_no, line in enumerate(path.read_text(encoding="utf-8").splitlines(), 1):
        if identifier_needle in line or "DataTracker.Builder" in line or "net.minecraft.item.tooltip.TooltipType" in line or "Item.TooltipContext" in line:
            survivors.append(f"{path.relative_to(root)}:{line_no}:{line.strip()}")
if survivors:
    raise SystemExit("post-1.20.1 API survived compatibility pass:\n" + "\n".join(survivors))

barrier_text = barrier.read_text(encoding="utf-8")
for required in (
    "protected void initDataTracker()",
    'startTracking(SPELL_ID_TRACKER, "")',
    "startTracking(OWNER_ID_TRACKER, 0)",
    "startTracking(TIME_TO_LIVE_TRACKER, 0)",
):
    if required not in barrier_text:
        raise SystemExit(f"BarrierEntity compatibility invariant missing: {required}")
workbench_text = workbench.read_text(encoding="utf-8")
for required in (
    "import net.minecraft.client.item.TooltipContext;",
    "appendTooltip(ItemStack stack, @Nullable BlockView world, List<Text> tooltip, TooltipContext options)",
    "super.appendTooltip(stack, world, tooltip, options);",
):
    if required not in workbench_text:
        raise SystemExit(f"MonkWorkbench compatibility invariant missing: {required}")

print(
    f"[Paladins 1.20.1 API] Identifier.of -> constructor: "
    f"{identifier_replacements} occurrences across {len(identifier_files)} files"
)
print("[Paladins 1.20.1 API] first javac frontier translated with pinned historical API shapes")
