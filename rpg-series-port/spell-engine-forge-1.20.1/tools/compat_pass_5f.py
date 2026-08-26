#!/usr/bin/env python3
from pathlib import Path
import json,sys
if len(sys.argv)!=3: raise SystemExit('usage: compat_pass_5f.py <generated-port-root> <baseline>')
root=Path(sys.argv[1]).resolve(); J=root/'common/src/main/java'; R=root/'common/src/main/resources'
old=J/'net/spell_engine/mixin/client/render/tint/RenderLayerMixin.java'
if not old.exists(): raise SystemExit('expected old RenderLayerMixin missing')
old.unlink()

# Forge 47.4.23 adds an armor-resource-aware renderModel overload after the vanilla/Yarn
# renderArmorParts(..., String overlay) method. The vanilla method delegates to that Forge overload,
# and the Forge overload owns the actual RenderLayer.getArmorCutoutNoCull(texture) call. Targeting
# renderArmorParts therefore scans a real method but finds zero matching invocations at runtime.
# Hook the exact Forge-added overload instead so custom Forge armor textures/models are preserved while
# translucent Spell Engine entity tints select a blending-capable layer.
q=J/'net/spell_engine/mixin/client/render/tint/ArmorFeatureRendererMixin.java'
q.write_text('''package net.spell_engine.mixin.client.render.tint;\n\nimport com.llamalad7.mixinextras.injector.wrapoperation.Operation;\nimport com.llamalad7.mixinextras.injector.wrapoperation.WrapOperation;\nimport net.minecraft.client.render.RenderLayer;\nimport net.minecraft.client.render.entity.feature.ArmorFeatureRenderer;\nimport net.minecraft.util.Identifier;\nimport net.spell_engine.api.effect.EntityTints;\nimport org.spongepowered.asm.mixin.Mixin;\nimport org.spongepowered.asm.mixin.injection.At;\n\n@Mixin(ArmorFeatureRenderer.class)\npublic class ArmorFeatureRendererMixin {\n    @WrapOperation(\n            method = "renderModel(Lnet/minecraft/client/util/math/MatrixStack;Lnet/minecraft/client/render/VertexConsumerProvider;ILnet/minecraft/item/ArmorItem;Lnet/minecraft/client/model/Model;ZFFFLnet/minecraft/util/Identifier;)V",\n            at = @At(\n                    value = "INVOKE",\n                    target = "Lnet/minecraft/client/render/RenderLayer;getArmorCutoutNoCull(Lnet/minecraft/util/Identifier;)Lnet/minecraft/client/render/RenderLayer;"\n            )\n    )\n    private RenderLayer spellEngine_translucentTintedArmor(Identifier texture, Operation<RenderLayer> original) {\n        if (EntityTints.Current.isTranslucent()) {\n            return RenderLayer.getItemEntityTranslucentCull(texture);\n        }\n        return original.call(texture);\n    }\n}\n''')

mix=R/'spell_engine.mixins.json'; data=json.loads(mix.read_text())
client=data.get('client',[])
client=[x for x in client if x!='client.render.tint.RenderLayerMixin']
if 'client.render.tint.ArmorFeatureRendererMixin' not in client:
    client.append('client.render.tint.ArmorFeatureRendererMixin')
data['client']=client
mix.write_text(json.dumps(data,indent=2)+'\n')

# Own this warning family explicitly: no stale factory target or wrong vanilla overload may survive.
hits=[]
for f in J.rglob('*.java'):
    s=f.read_text()
    if 'createArmorCutoutNoCull' in s or 'class RenderLayerMixin' in s: hits.append(str(f.relative_to(J)))
if hits: raise SystemExit(f'pass5f stale armor render target remains: {hits}')
final=q.read_text()
for required in (
    'method = "renderModel(Lnet/minecraft/client/util/math/MatrixStack;Lnet/minecraft/client/render/VertexConsumerProvider;ILnet/minecraft/item/ArmorItem;Lnet/minecraft/client/model/Model;ZFFFLnet/minecraft/util/Identifier;)V"',
    'RenderLayer;getArmorCutoutNoCull',
    'RenderLayer.getItemEntityTranslucentCull(texture)',
):
    if required not in final: raise SystemExit(f'pass5f missing Forge armor tint hook: {required}')
if 'method = "renderArmorParts"' in final:
    raise SystemExit('pass5f left stale vanilla renderArmorParts armor-layer hook')
print('Spell Engine compatibility pass 5f applied: Forge 47 armor-resource renderModel translucent tint hook')
