#!/usr/bin/env python3
"""Compatibility wrapper for the audited Fauna & Orchestra 3.0.3 performance patch.

The first CI pass proved that the exact 3.0.3 renderer source uses
shouldRenderOffScreen=true instead of the newer renderer-side AABB override. This wrapper
keeps the already-validated performance transformations and swaps only that final renderer
section for the exact 3.0.3 source shape before executing it.
"""
from pathlib import Path
import sys

legacy = Path(__file__).with_name("faunaandorchestra_3_0_3_performance_patch.py")
source = legacy.read_text(encoding="utf-8")

start_marker = "# 11) Animated block entity culling bounds."
end_marker = "# Forge also lets some block entities expose their own render AABB."
start = source.index(start_marker)
end = source.index(end_marker, start)

replacement = r'''# 11) Animated block entity culling. The exact 3.0.3 renderer source forces
# shouldRenderOffScreen() to true, bypassing normal frustum culling. The shipped model
# geometry measured from the user JAR is <=2.22 blocks with <=0.52-block animated
# translation, so normal culling plus the conservative +/-3 block BE bounds below is safe.
renderer_dir = ROOT / "src/main/java/net/migueel26/faunaandorchestra/client/block"
renderers_reenabled = []
for path in renderer_dir.glob("*BlockEntityRenderer.java"):
    text = path.read_text(encoding="utf-8")
    pattern = r"(@Override\s+public boolean shouldRenderOffScreen\([^)]*\)\s*\{\s*)return true;(\s*\})"
    new_text, count = re.subn(pattern, r"\1return false;\2", text)
    if count:
        path.write_text(new_text, encoding="utf-8", newline="\n")
        relpath = str(path.relative_to(ROOT)).replace("\\", "/")
        if relpath not in CHANGED:
            CHANGED.append(relpath)
        renderers_reenabled.append(relpath)
if not renderers_reenabled:
    raise RuntimeError("No forced-offscreen block entity renderers found; upstream source shape changed")
NOTES.append(
    f"Re-enabled frustum culling for {len(renderers_reenabled)} animated block-entity renderers that forced shouldRenderOffScreen=true."
)

'''
source = source[:start] + replacement + source[end:]
source = source.replace('"renderers_tightened": render_changed,', '"renderers_reenabled_for_culling": renderers_reenabled,')

# Avoid Python 3.12+ invalid-escape warnings for the two re.sub replacement strings while
# preserving their \g<body> backreferences.
source = source.replace('    """        if (isPlaying() && !level().isClientSide()) {\\n            if (nearbyPlayersSearchDelay < 60)',
                        '    r"""        if (isPlaying() && !level().isClientSide()) {\\n            if (nearbyPlayersSearchDelay < 60)')

code = compile(source, str(legacy), "exec")
namespace = {"__name__": "__main__", "__file__": str(legacy)}
exec(code, namespace, namespace)
