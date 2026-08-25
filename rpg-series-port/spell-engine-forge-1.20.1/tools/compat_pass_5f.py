#!/usr/bin/env python3
from pathlib import Path
import json,sys
if len(sys.argv)!=3: raise SystemExit('usage: compat_pass_5f.py <generated-port-root> <baseline>')
root=Path(sys.argv[1]).resolve(); J=root/'common/src/main/java'; R=root/'common/src/main/resources'
old=J/'net/spell_engine/mixin/client/render/tint/RenderLayerMixin.java'
if not old.exists(): raise SystemExit('expected old RenderLayerMixin missing')
old.unlink()

# 1.20.1 RenderLayer does not have the later createArmorCutoutNoCull factory that the upstream
# mixin targeted. The stable semantic call site is ArmorFeatureRenderer.renderArmorParts(), which
# invokes RenderLayer.getArmorCutoutNoCull(texture). Wrap that exact call and select vanilla's
# blending-capable entity layer only while Spell Engine's active entity tint has alpha < 1.
q=J/'net/spell_engine/mixin/client/render/tint/ArmorFeatureRendererMixin.java'
q.write_text('''package net.spell_engine.mixin.client.render.tint;\n\nimport com.llamalad7.mixinextras.injector.wrapoperation.Operation;\nimport com.llamalad7.mixinextras.injector.wrapoperation.WrapOperation;\nimport net.minecraft.client.render.RenderLayer;\nimport net.minecraft.client.render.entity.feature.ArmorFeatureRenderer;\nimport net.minecraft.util.Identifier;\nimport net.spell_engine.api.effect.EntityTints;\nimport org.spongepowered.asm.mixin.Mixin;\nimport org.spongepowered.asm.mixin.injection.At;\n\n@Mixin(ArmorFeatureRenderer.class)\npublic class ArmorFeatureRendererMixin {\n    @WrapOperation(\n            method = "renderArmorParts",\n            at = @At(\n                    value = "INVOKE",\n                    target = "Lnet/minecraft/client/render/RenderLayer;getArmorCutoutNoCull(Lnet/minecraft/util/Identifier;)Lnet/minecraft/client/render/RenderLayer;"\n            )\n    )\n    private RenderLayer spellEngine_translucentTintedArmor(Identifier texture, Operation<RenderLayer> original) {\n        if (EntityTints.Current.isTranslucent()) {\n            return RenderLayer.getItemEntityTranslucentCull(texture);\n        }\n        return original.call(texture);\n    }\n}\n''')

mix=R/'spell_engine.mixins.json'; data=json.loads(mix.read_text())
client=data.get('client',[])
client=[x for x in client if x!='client.render.tint.RenderLayerMixin']
if 'client.render.tint.ArmorFeatureRendererMixin' not in client:
    client.append('client.render.tint.ArmorFeatureRendererMixin')
data['client']=client
mix.write_text(json.dumps(data,indent=2)+'\n')

# Own this warning family explicitly: no stale factory target may survive.
hits=[]
for f in J.rglob('*.java'):
    s=f.read_text()
    if 'createArmorCutoutNoCull' in s or 'class RenderLayerMixin' in s: hits.append(str(f.relative_to(J)))
if hits: raise SystemExit(f'pass5f stale armor render target remains: {hits}')
print('Spell Engine compatibility pass 5f applied: 1.20.1 ArmorFeatureRenderer translucent tint hook')
