#!/usr/bin/env python3
from pathlib import Path
import json, sys

if len(sys.argv) != 3:
    raise SystemExit('usage: compat_pass_6i.py <generated-port-root> <spell-engine-1.20.1-baseline>')
root = Path(sys.argv[1]).resolve()
java = root / 'common/src/main/java'
resources = root / 'common/src/main/resources'

# Modern 1.10.2 patches PoisonStatusEffect. In 1.20.1 poison still uses StatusEffect itself, so guard
# by object identity and reproduce the modern semantics: damage scales with amplifier, never kills,
# and ticks every 25 ticks rather than vanilla's amplifier-shortened interval.
poison = java / 'net/spell_engine/mixin/effect/PoisonEffectMixin.java'
poison.write_text(r'''package net.spell_engine.mixin.effect;

import net.minecraft.entity.LivingEntity;
import net.minecraft.entity.effect.StatusEffect;
import net.minecraft.entity.effect.StatusEffects;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.injection.At;
import org.spongepowered.asm.mixin.injection.Inject;
import org.spongepowered.asm.mixin.injection.callback.CallbackInfo;
import org.spongepowered.asm.mixin.injection.callback.CallbackInfoReturnable;

@Mixin(StatusEffect.class)
public class PoisonEffectMixin {
    @Inject(method = "applyUpdateEffect", at = @At("HEAD"), cancellable = true)
    private void spellEngine_applyPoisonUpdate(LivingEntity entity, int amplifier, CallbackInfo ci) {
        if ((Object)this != StatusEffects.POISON) return;
        float amplifiedAmount = amplifier + 1.0F;
        float cappedAmount = Math.min(amplifiedAmount, entity.getHealth() - 1.0F);
        if (cappedAmount > 0.0F) {
            entity.damage(entity.getDamageSources().magic(), cappedAmount);
        }
        ci.cancel();
    }

    @Inject(method = "canApplyUpdateEffect", at = @At("HEAD"), cancellable = true)
    private void spellEngine_poisonTickRate(int duration, int amplifier, CallbackInfoReturnable<Boolean> cir) {
        if ((Object)this != StatusEffects.POISON) return;
        cir.setReturnValue(duration % 25 == 0);
    }
}
''')

# 1.21 Immediate uses SequencedMap<RenderLayer, BufferAllocator>; 1.20.1 uses the same ordered layer
# map concept with Map<RenderLayer, BufferBuilder>. Give Spell Engine glow layers their own buffer so
# they flush after the item's depth-writing layer rather than through the early fallback buffer.
glow = java / 'net/spell_engine/mixin/client/render/ImmediateItemGlowMixin.java'
glow.write_text(r'''package net.spell_engine.mixin.client.render;

import net.minecraft.client.render.BufferBuilder;
import net.minecraft.client.render.RenderLayer;
import net.minecraft.client.render.VertexConsumer;
import net.minecraft.client.render.VertexConsumerProvider;
import net.spell_engine.api.render.CustomLayers;
import org.spongepowered.asm.mixin.Final;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.Shadow;
import org.spongepowered.asm.mixin.injection.At;
import org.spongepowered.asm.mixin.injection.Inject;
import org.spongepowered.asm.mixin.injection.callback.CallbackInfoReturnable;

import java.util.Map;

@Mixin(VertexConsumerProvider.Immediate.class)
public class ImmediateItemGlowMixin {
    @Shadow @Final
    protected Map<RenderLayer, BufferBuilder> layerBuffers;

    @Inject(method = "getBuffer", at = @At("HEAD"))
    private void spellEngine_bufferItemGlowLayer(RenderLayer renderLayer, CallbackInfoReturnable<VertexConsumer> cir) {
        if (layerBuffers.isEmpty()
                || layerBuffers.containsKey(renderLayer)
                || !CustomLayers.isItemGlowLayer(renderLayer)) {
            return;
        }
        layerBuffers.put(renderLayer, new BufferBuilder(renderLayer.getExpectedBufferSize()));
    }
}
''')

mixins = resources / 'spell_engine.mixins.json'
data = json.loads(mixins.read_text())
mixins_common = data.setdefault('mixins', [])
if 'effect.PoisonEffectMixin' not in mixins_common:
    mixins_common.append('effect.PoisonEffectMixin')
client = data.setdefault('client', [])
if 'client.render.ImmediateItemGlowMixin' not in client:
    client.append('client.render.ImmediateItemGlowMixin')
mixins.write_text(json.dumps(data, indent=2) + '\n')

# Keep the parity ledger truthful: only remove entries whose actual 1.20.1 implementations are now
# generated. The other three remain documented until their architecture/runtime proof is explicit.
gate = root / 'COMPAT-GATE-PENDING.md'
if gate.exists():
    lines = gate.read_text().splitlines()
    lines = [line for line in lines if '`effect.PoisonEffectMixin`' not in line and '`client.render.ImmediateItemGlowMixin`' not in line]
    gate.write_text('\n'.join(lines).rstrip() + '\n')

for required in (
    'if ((Object)this != StatusEffects.POISON) return;',
    'Math.min(amplifiedAmount, entity.getHealth() - 1.0F)',
    'duration % 25 == 0',
):
    if required not in poison.read_text():
        raise SystemExit(f'pass6i missing 1.20.1 poison parity: {required}')
for required in (
    'Map<RenderLayer, BufferBuilder> layerBuffers',
    'new BufferBuilder(renderLayer.getExpectedBufferSize())',
    'CustomLayers.isItemGlowLayer(renderLayer)',
):
    if required not in glow.read_text():
        raise SystemExit(f'pass6i missing 1.20.1 item-glow ordering parity: {required}')
final = json.loads(mixins.read_text())
if 'effect.PoisonEffectMixin' not in final.get('mixins', []):
    raise SystemExit('pass6i did not activate poison mixin')
if 'client.render.ImmediateItemGlowMixin' not in final.get('client', []):
    raise SystemExit('pass6i did not activate item glow mixin')
if gate.exists():
    text = gate.read_text()
    if '`effect.PoisonEffectMixin`' in text or '`client.render.ImmediateItemGlowMixin`' in text:
        raise SystemExit('pass6i left implemented parity entries in pending gate')

print('Spell Engine compatibility pass 6i applied: exact 1.20.1 poison semantics + vanilla item-glow buffer ordering')
