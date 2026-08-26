#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 2:
    raise SystemExit('usage: compat_1_20_1.py <generated-port-root>')

root = Path(sys.argv[1]).resolve()
common = root / 'common/src/main/java/net/rpg_foundation/armor_api/client'


def rewrite(path: Path, transform):
    if not path.is_file():
        raise SystemExit(f'missing compatibility target: {path}')
    before = path.read_text()
    after = transform(before)
    if after == before:
        raise SystemExit(f'compatibility transform made no change: {path}')
    path.write_text(after)


def armor_dispatcher(source: str) -> str:
    source = source.replace('import net.minecraft.component.type.DyedColorComponent;\n', '')
    source = source.replace('import net.minecraft.registry.tag.ItemTags;\n', '')
    source = source.replace('import net.minecraft.util.math.ColorHelper;\n', '')
    source = source.replace('import net.minecraft.item.ItemStack;\n', 'import net.minecraft.item.ItemStack;\nimport net.minecraft.item.DyeableItem;\n')
    old = '''        int color = stack.isIn(ItemTags.DYEABLE)\n                ? ColorHelper.Argb.fullAlpha(DyedColorComponent.getColor(stack, DyedColorComponent.DEFAULT_COLOR))\n                : -1;'''
    new = '''        int color = stack.getItem() instanceof DyeableItem dyeable\n                ? 0xFF000000 | dyeable.getColor(stack)\n                : -1;'''
    if old not in source:
        raise SystemExit('ArmorRenderDispatcher dye component block drifted upstream')
    return source.replace(old, new)


def trim_layer(source: str) -> str:
    source = source.replace('import net.minecraft.component.DataComponentTypes;\n', '')
    old = '''        var trim = context.stack().get(DataComponentTypes.TRIM);\n        if (trim == null) {\n            return;\n        }'''
    new = '''        var trim = ArmorTrim.getTrim(\n                context.entity().getWorld().getRegistryManager(), context.stack()).orElse(null);\n        if (trim == null) {\n            return;\n        }'''
    if old not in source:
        raise SystemExit('TrimLayer data-component block drifted upstream')
    source = source.replace(old, new)
    old_layer = 'TexturedRenderLayers.getArmorTrims(trim.getPattern().value().decal())'
    if old_layer not in source:
        raise SystemExit('TrimLayer 1.21 render-layer call drifted upstream')
    return source.replace(old_layer, 'TexturedRenderLayers.getArmorTrims()')


rewrite(common / 'ArmorRenderDispatcher.java', armor_dispatcher)
rewrite(common / 'layer/TrimLayer.java', trim_layer)

shader = common / 'compatibility/ShaderCompat.java'
if not shader.is_file():
    raise SystemExit(f'missing compatibility target: {shader}')
shader.write_text('''package net.rpg_foundation.armor_api.client.compatibility;\n\nimport net.fabricmc.api.EnvType;\nimport net.fabricmc.api.Environment;\nimport net.rpg_foundation.armor_api.Platform;\n\nimport java.lang.reflect.Method;\nimport java.util.function.Supplier;\n\n/// Shader-mod awareness for the 1.20.1 backport. Iris is compile-only upstream; Forge 1.20.1\n/// commonly exposes the same API through Oculus. Resolve the API reflectively so neither shader\n/// mod becomes a mandatory dependency while preserving live shader-pack state checks.\n@Environment(EnvType.CLIENT)\npublic final class ShaderCompat {\n\n    private static Supplier<Boolean> shaderPackInUse = () -> false;\n    private static boolean vanillaRenderSystem = true;\n\n    private ShaderCompat() { }\n\n    public static void initialize() {\n        if (!Platform.util().isModLoaded("iris") && !Platform.util().isModLoaded("oculus")) {\n            return;\n        }\n        try {\n            Class<?> apiClass = Class.forName("net.irisshaders.iris.api.v0.IrisApi");\n            Object api = apiClass.getMethod("getInstance").invoke(null);\n            Method query = apiClass.getMethod("isShaderPackInUse");\n            vanillaRenderSystem = false;\n            shaderPackInUse = () -> {\n                try {\n                    return Boolean.TRUE.equals(query.invoke(api));\n                } catch (ReflectiveOperationException ignored) {\n                    return false;\n                }\n            };\n        } catch (ReflectiveOperationException ignored) {\n            shaderPackInUse = () -> false;\n            vanillaRenderSystem = true;\n        }\n    }\n\n    public static boolean isShaderPackInUse() {\n        return shaderPackInUse.get();\n    }\n\n    public static boolean isVanillaRenderSystem() {\n        return vanillaRenderSystem;\n    }\n}\n''')

# Make stale 1.21 component imports a hard preparation failure instead of another CI surprise.
for path in common.rglob('*.java'):
    text = path.read_text()
    if 'net.minecraft.component' in text:
        raise SystemExit(f'1.21 data-component API remains after compatibility pass: {path}')

print('Armor Model API 1.20.1 common-source compatibility pass applied')
