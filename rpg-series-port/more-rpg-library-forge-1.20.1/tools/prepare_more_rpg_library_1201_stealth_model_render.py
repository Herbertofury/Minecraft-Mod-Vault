#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 2:
    raise SystemExit('usage: prepare_more_rpg_library_1201_stealth_model_render.py <prepared-root>')
root = Path(sys.argv[1]).resolve()
path = root / 'common/src/main/java/net/more_rpg_classes/mixin/LivingEntityRenderStealth.java'
if not path.is_file():
    raise SystemExit(f'missing LivingEntityRenderStealth source: {path}')
s = path.read_text(encoding='utf-8')

# The modern renderer wraps EntityModel.render(..., int color). Minecraft 1.20.1 still uses the
# legacy RGBA-float render contract. Preserve the exact stealth visual semantics by forwarding RGB
# unchanged and multiplying only alpha by 0.15 when the entity is stealth-visible.
modern_target = 'target = "Lnet/minecraft/client/render/entity/model/EntityModel;render(Lnet/minecraft/client/util/math/MatrixStack;Lnet/minecraft/client/render/VertexConsumer;III)V"'
target_1201 = 'target = "Lnet/minecraft/client/render/entity/model/EntityModel;render(Lnet/minecraft/client/util/math/MatrixStack;Lnet/minecraft/client/render/VertexConsumer;IIFFFF)V"'
if s.count(modern_target) != 1:
    raise SystemExit(f'Stealth model render target seam drifted: found={s.count(modern_target)}')
if s.count(target_1201) != 0:
    raise SystemExit('Stealth target 1.20.1 RGBA descriptor unexpectedly pre-exists')
s = s.replace(modern_target, target_1201, 1)

modern_params = '            EntityModel instance, MatrixStack matrices, VertexConsumer vertices, int light, int overlay, int color, Operation<Void> original,\n'
target_params = '            EntityModel instance, MatrixStack matrices, VertexConsumer vertices, int light, int overlay, float red, float green, float blue, float alpha, Operation<Void> original,\n'
if s.count(modern_params) != 1:
    raise SystemExit(f'Stealth packed-color handler parameter seam drifted: found={s.count(modern_params)}')
s = s.replace(modern_params, target_params, 1)

modern_body = '''        if (hasStealthEffect(entity) && visibleForLocalPlayer(entity)) {\n            var alpha = ColorHelper.Argb.getAlpha(color);\n            var red = ColorHelper.Argb.getRed(color);\n            var green = ColorHelper.Argb.getGreen(color);\n            var blue = ColorHelper.Argb.getBlue(color);\n            var newColor = ColorHelper.Argb.getArgb((int) (alpha * 0.15F), red, green, blue);\n            original.call(instance, matrices, vertices, light, overlay, newColor);\n        } else {\n            original.call(instance, matrices, vertices, light, overlay, color);\n        }'''
target_body = '''        if (hasStealthEffect(entity) && visibleForLocalPlayer(entity)) {\n            original.call(instance, matrices, vertices, light, overlay, red, green, blue, alpha * 0.15F);\n        } else {\n            original.call(instance, matrices, vertices, light, overlay, red, green, blue, alpha);\n        }'''
if s.count(modern_body) != 1:
    raise SystemExit(f'Stealth packed-color body seam drifted: found={s.count(modern_body)}')
s = s.replace(modern_body, target_body, 1)

color_import = 'import net.minecraft.util.math.ColorHelper;\n'
if s.count(color_import) != 1:
    raise SystemExit(f'Stealth ColorHelper import seam drifted: found={s.count(color_import)}')
s = s.replace(color_import, '', 1)

for forbidden in (modern_target, 'int color, Operation<Void> original', 'ColorHelper.Argb', 'newColor'):
    if forbidden in s:
        raise SystemExit(f'modern packed-color stealth render seam survived: {forbidden}')
if s.count(target_1201) != 1:
    raise SystemExit('target 1.20.1 stealth RGBA render descriptor missing or duplicated')
if s.count('alpha * 0.15F') != 1:
    raise SystemExit('stealth alpha attenuation semantic missing or duplicated')
path.write_text(s, encoding='utf-8')
print('[More RPG 2.7.2] STEALTH_MODEL_RENDER_RGBA_1201_PASS target=IIFFFF alpha_multiplier=0.15 source=client-preaudit')
