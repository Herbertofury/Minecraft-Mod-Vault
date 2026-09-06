#!/usr/bin/env python3
from pathlib import Path
import json,re,sys
if len(sys.argv)!=3: raise SystemExit('usage: compat_pass_5e.py <generated-port-root> <baseline>')
root=Path(sys.argv[1]).resolve(); J=root/'common/src/main/java'; R=root/'common/src/main/resources'
def p(r): return J/r
def ed(r,fn):
 q=p(r); old=q.read_text(); new=fn(old)
 if new==old: raise SystemExit(f'pass5e transform did not match: {r}')
 q.write_text(new)

# ProjectileEntityRenderer<T> erases T to PersistentProjectileEntity in 1.20.1. Target that exact
# descriptor instead of Entity; this keeps the composite-arrow renderer hook at the same HEAD point.
def projectile(s):
 s=s.replace('import net.minecraft.entity.Entity;','import net.minecraft.entity.Entity;\nimport net.minecraft.entity.projectile.PersistentProjectileEntity;')
 s=s.replace('method = "render(Lnet/minecraft/entity/Entity;FFLnet/minecraft/client/util/math/MatrixStack;Lnet/minecraft/client/render/VertexConsumerProvider;I)V",',
             'method = "render(Lnet/minecraft/entity/projectile/PersistentProjectileEntity;FFLnet/minecraft/client/util/math/MatrixStack;Lnet/minecraft/client/render/VertexConsumerProvider;I)V",')
 s=s.replace('private void render_HEAD_SpellEngine(Entity entity, float yaw, float tickDelta, MatrixStack matrices, VertexConsumerProvider vertexConsumers, int light, CallbackInfo ci)',
             'private void render_HEAD_SpellEngine(PersistentProjectileEntity entity, float yaw, float tickDelta, MatrixStack matrices, VertexConsumerProvider vertexConsumers, int light, CallbackInfo ci)')
 return s
ed('net/spell_engine/mixin/client/render/ProjectileEntityRendererMixin.java',projectile)

# 1.20.1 GrindstoneScreenHandler has updateResult(), not 1.21 getOutputStack(first,second). Shadow the
# genuine input/result inventories and implement the same custom GRINDABLE -> paper output at HEAD.
def grindstone(s):
 return '''package net.spell_engine.mixin.item;\n\nimport net.minecraft.inventory.Inventory;\nimport net.minecraft.item.ItemStack;\nimport net.minecraft.item.Items;\nimport net.minecraft.screen.GrindstoneScreenHandler;\nimport net.spell_engine.api.tags.SpellEngineItemTags;\nimport org.spongepowered.asm.mixin.Final;\nimport org.spongepowered.asm.mixin.Mixin;\nimport org.spongepowered.asm.mixin.Shadow;\nimport org.spongepowered.asm.mixin.injection.At;\nimport org.spongepowered.asm.mixin.injection.Inject;\nimport org.spongepowered.asm.mixin.injection.callback.CallbackInfo;\n\n@Mixin(GrindstoneScreenHandler.class)\npublic abstract class GrindstoneScreenHandlerMixin {\n    @Shadow @Final private Inventory input;\n    @Shadow @Final private Inventory result;\n\n    @Inject(method = "updateResult", at = @At("HEAD"), cancellable = true)\n    private void updateResult_HEAD_SpellEngine(CallbackInfo ci) {\n        ItemStack firstInput = input.getStack(0);\n        ItemStack secondInput = input.getStack(1);\n        if (firstInput.isIn(SpellEngineItemTags.GRINDABLE) && secondInput.isEmpty()) {\n            result.setStack(0, new ItemStack(Items.PAPER, 1));\n            ((GrindstoneScreenHandler)(Object)this).sendContentUpdates();\n            ci.cancel();\n        }\n    }\n}\n'''
ed('net/spell_engine/mixin/item/GrindstoneScreenHandlerMixin.java',grindstone)

# isInvulnerableTo is declared on Entity in 1.20.1, so a LivingEntity mixin cannot legally inject into
# that inherited method. Keep immunity storage/status-effect blocking on LivingEntity, and move damage
# return augmentation to an Entity mixin guarded by LivingEntityImmunity.Owner.
def living(s):
 s=s.replace('import com.llamalad7.mixinextras.injector.ModifyReturnValue;\n','')
 block=re.compile(r'''\n    @ModifyReturnValue\(method = "isInvulnerableTo", at = @At\("RETURN"\)\)\n    private boolean isInvulnerableTo_RETURN_SpellEngine_Immunity\(boolean original, DamageSource damageSource\) \{\n        return original \|\| LivingEntityImmunity\.isDamageProtected\(immunities, damageSource\);\n    \}\n''')
 ns,n=block.subn('\n',s)
 if n!=1: raise SystemExit('LivingEntity immunity inherited-method block not found')
 return ns
ed('net/spell_engine/mixin/entity/LivingEntityImmunityMixin.java',living)
(J/'net/spell_engine/mixin/entity/EntityDamageImmunityMixin.java').write_text('''package net.spell_engine.mixin.entity;\n\nimport com.llamalad7.mixinextras.injector.ModifyReturnValue;\nimport net.minecraft.entity.Entity;\nimport net.minecraft.entity.damage.DamageSource;\nimport net.spell_engine.api.entity.LivingEntityImmunity;\nimport org.spongepowered.asm.mixin.Mixin;\nimport org.spongepowered.asm.mixin.injection.At;\n\n@Mixin(Entity.class)\npublic class EntityDamageImmunityMixin {\n    @ModifyReturnValue(method = "isInvulnerableTo", at = @At("RETURN"))\n    private boolean isInvulnerableTo_RETURN_SpellEngine_Immunity(boolean original, DamageSource source) {\n        if (original) return true;\n        Object self = this;\n        if (self instanceof LivingEntityImmunity.Owner owner) {\n            return LivingEntityImmunity.isDamageProtected(owner.getImmunities(), source);\n        }\n        return false;\n    }\n}\n''')
mix=R/'spell_engine.mixins.json'; data=json.loads(mix.read_text())
if 'entity.EntityDamageImmunityMixin' not in data.get('mixins',[]):
 data.setdefault('mixins',[]).append('entity.EntityDamageImmunityMixin')
mix.write_text(json.dumps(data,indent=2)+'\n')

for needle in ('render(Lnet/minecraft/entity/Entity;FFLnet/minecraft/client/util/math/MatrixStack','@Inject(method = "getOutputStack"','@ModifyReturnValue(method = "isInvulnerableTo"'):
 hits=[]
 for q in J.rglob('*.java'):
  if needle in q.read_text() and q.name!='EntityDamageImmunityMixin.java': hits.append(str(q.relative_to(J)))
 if hits: raise SystemExit(f'pass5e incomplete {needle}: {hits[:20]}')
print('Spell Engine compatibility pass 5e applied: exact projectile/grindstone/damage-immunity Mixin targets')
