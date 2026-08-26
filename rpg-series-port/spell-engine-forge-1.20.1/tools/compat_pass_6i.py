#!/usr/bin/env python3
from pathlib import Path
import json, sys

if len(sys.argv) != 3:
    raise SystemExit('usage: compat_pass_6i.py <generated-port-root> <spell-engine-1.20.1-baseline>')
root = Path(sys.argv[1]).resolve()
java = root / 'common/src/main/java'
resources = root / 'common/src/main/resources'
forge_build = root / 'forge/build.gradle'

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

# The modern fallback-loot analyzer reads LootPool entries. A direct field read compiles in Loom's
# widened development classes but the private SRG field remains private in a real Forge install.
# Bridge it with a typed Mixin accessor so the exact same analyzer works in the packaged artifact.
loot_pool_accessor = java / 'net/spell_engine/mixin/loot/LootPoolAccessor.java'
loot_pool_accessor.write_text(r'''package net.spell_engine.mixin.loot;

import net.minecraft.loot.LootPool;
import net.minecraft.loot.entry.LootPoolEntry;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.gen.Accessor;

@Mixin(LootPool.class)
public interface LootPoolAccessor {
    @Accessor("entries")
    LootPoolEntry[] spellEngine_getEntries();
}
''')
if 'loot.LootPoolAccessor' not in mixins_common:
    mixins_common.append('loot.LootPoolAccessor')

loot_helper = java / 'net/spell_engine/rpg_series/loot/LootHelper.java'
lh = loot_helper.read_text()
if 'import net.spell_engine.mixin.loot.LootPoolAccessor;' not in lh:
    lh = lh.replace('package net.spell_engine.rpg_series.loot;',
                    'package net.spell_engine.rpg_series.loot;\nimport net.spell_engine.mixin.loot.LootPoolAccessor;', 1)
inspect_anchor = '''    private static List<PoolContents> inspect(List<LootPool> pools) {\n        var result = new ArrayList<PoolContents>(pools.size());\n        for (var pool: pools) {\n            var items = new LinkedHashMap<String, ItemOccurrence>();\n            int[] total = { 0 };\n            for (var entry: pool.entries) {'''
inspect_replacement = '''    private static List<PoolContents> inspect(List<LootPool> pools) {\n        var result = new ArrayList<PoolContents>(pools.size());\n        for (var pool: pools) {\n            var items = new LinkedHashMap<String, ItemOccurrence>();\n            int[] total = { 0 };\n            for (var entry: ((LootPoolAccessor) (Object) pool).spellEngine_getEntries()) {'''
if inspect_anchor in lh:
    lh = lh.replace(inspect_anchor, inspect_replacement, 1)
elif '((LootPoolAccessor) (Object) pool).spellEngine_getEntries()' not in lh:
    raise SystemExit('pass6i could not locate vanilla LootPool inspection field read')
loot_helper.write_text(lh)

mixins.write_text(json.dumps(data, indent=2) + '\n')

# Loom's development launches discover mixin configs from the generated run configuration. A real
# Forge installation does not inherit that launch configuration: it discovers mod mixins from the
# packaged JAR manifest. Without this declaration the release JAR contains spell_engine.mixins.json
# but applies none of its common/client mixins, which breaks contracts such as StatusEffect ->
# ActionImpairing on a fresh dedicated server. Put the declaration on every Jar task so shadowJar,
# the input to remapJar, carries it into the distributable artifact.
fb = forge_build.read_text()
if "attributes 'MixinConfigs': 'spell_engine.mixins.json'" not in fb:
    fb += r'''

// Real Forge installations do not inherit Loom's dev-run mixin bootstrap. Every release/shadow JAR
// must advertise the Spell Engine mixin config from its manifest.
tasks.withType(Jar).configureEach {
    manifest {
        attributes 'MixinConfigs': 'spell_engine.mixins.json'
    }
}
'''
forge_build.write_text(fb)

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
for required in (
    'effect.StatusEffectActionImpairing',
    'effect.StatusEffectSynchronized',
    'action_impair.LivingEntityActionImpairing',
    'loot.LootPoolAccessor',
):
    if required not in final.get('mixins', []):
        raise SystemExit(f'pass6i required common mixin missing from packaged config: {required}')
if 'LootPoolEntry[] spellEngine_getEntries();' not in loot_pool_accessor.read_text():
    raise SystemExit('pass6i LootPool accessor descriptor missing')
if '((LootPoolAccessor) (Object) pool).spellEngine_getEntries()' not in loot_helper.read_text():
    raise SystemExit('pass6i LootHelper still lacks packaged-safe LootPool accessor use')
# Public LootConfig.Pool.entries is intentionally untouched; only vanilla LootPool private access is bridged.
if loot_helper.read_text().count('for (var itemInjectorEntry: pool.entries)') != 1:
    raise SystemExit('pass6i accidentally changed LootConfig tag-cache pool entries')
if loot_helper.read_text().count('for (var entry: pool.entries)') != 1:
    raise SystemExit('pass6i accidentally changed LootConfig buildPool entries')
final_build = forge_build.read_text()
for required in (
    'tasks.withType(Jar).configureEach',
    "attributes 'MixinConfigs': 'spell_engine.mixins.json'",
):
    if required not in final_build:
        raise SystemExit(f'pass6i missing packaged Forge mixin bootstrap: {required}')
if gate.exists():
    text = gate.read_text()
    if '`effect.PoisonEffectMixin`' in text or '`client.render.ImmediateItemGlowMixin`' in text:
        raise SystemExit('pass6i left implemented parity entries in pending gate')

print('Spell Engine compatibility pass 6i applied: poison/item glow + packaged mixins + LootPool accessor')
