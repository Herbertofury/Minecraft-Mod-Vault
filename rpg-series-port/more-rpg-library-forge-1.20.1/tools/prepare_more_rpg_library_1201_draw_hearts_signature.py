#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 2:
    raise SystemExit('usage: prepare_more_rpg_library_1201_draw_hearts_signature.py <prepared-root>')
root = Path(sys.argv[1]).resolve()
path = root / 'common/src/main/java/net/more_rpg_classes/mixin/DrawHeartsMixin.java'
if not path.is_file():
    raise SystemExit(f'missing DrawHeartsMixin source: {path}')
s = path.read_text(encoding='utf-8')

# Run #355 reached a green packaged Forge server and then failed on the mapped client because the
# modern callback descriptor no longer matches Minecraft 1.20.1. Modern drawHeart receives a hardcore
# boolean; target 1.20.1 receives the texture-V integer instead, followed by blinking and halfHeart.
# More RPG never reads the modern hardcore argument, so restore only the target descriptor and keep the
# fatal-poison rendering logic unchanged.
modern = ('    private void drawHeart(DrawContext context, InGameHud.HeartType type, int x, int y, '
          'boolean hardcore, boolean blinking, boolean half, CallbackInfo ci) {')
target = ('    private void drawHeart(DrawContext context, InGameHud.HeartType type, int x, int y, '
          'int v, boolean blinking, boolean half, CallbackInfo ci) {')
if s.count(modern) != 1:
    raise SystemExit(f'DrawHeartsMixin modern callback descriptor seam drifted: found={s.count(modern)}')
if s.count(target) != 0:
    raise SystemExit('DrawHeartsMixin target callback descriptor unexpectedly pre-exists')
s = s.replace(modern, target, 1)

if s.count(target) != 1 or 'boolean hardcore' in s:
    raise SystemExit('DrawHeartsMixin 1.20.1 callback descriptor repair incomplete')
if s.count('@Inject(method = "drawHeart"') != 1:
    raise SystemExit('DrawHeartsMixin drawHeart injection selector drifted')
path.write_text(s, encoding='utf-8')
print('[More RPG 2.7.2] DRAW_HEARTS_CALLBACK_DESCRIPTOR_1201_PASS target=IIIZZ source=run-355-mapped-client')
