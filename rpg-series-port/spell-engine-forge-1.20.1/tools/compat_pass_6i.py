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

# Minecraft 1.21.1 made enchantments data-driven; Minecraft 1.20.1 still uses the static ENCHANTMENT
# registry. Preserve the current spell_infinity definition as a real 1.20.1 Enchantment object so
# modern tags, Ammo lookup, enchanting/anvils and the non_treasure tag all resolve the same ID.
spell_infinity = java / 'net/spell_engine/compat/enchantment/SpellInfinityEnchantment1201.java'
spell_infinity.parent.mkdir(parents=True, exist_ok=True)
spell_infinity.write_text(r'''package net.spell_engine.compat.enchantment;

import net.minecraft.enchantment.Enchantment;
import net.minecraft.enchantment.EnchantmentTarget;
import net.minecraft.enchantment.MendingEnchantment;
import net.minecraft.entity.EquipmentSlot;
import net.minecraft.item.ItemStack;
import net.minecraft.util.Identifier;
import net.spell_engine.SpellEngineMod;
import net.spell_engine.api.tags.SpellEngineItemTags;

/** Registry-backed 1.20.1 equivalent of Spell Engine 1.10.2's data-driven spell_infinity. */
public final class SpellInfinityEnchantment1201 extends Enchantment {
    public static final Identifier ID = new Identifier(SpellEngineMod.ID, "spell_infinity");
    public static final SpellInfinityEnchantment1201 INSTANCE = new SpellInfinityEnchantment1201();

    private SpellInfinityEnchantment1201() {
        super(Rarity.VERY_RARE, EnchantmentTarget.BREAKABLE, new EquipmentSlot[]{EquipmentSlot.MAINHAND});
    }

    @Override public int getMaxLevel() { return 1; }
    @Override public int getMinPower(int level) { return 20; }
    @Override public int getMaxPower(int level) { return 50; }
    @Override public boolean isAcceptableItem(ItemStack stack) { return stack.isIn(SpellEngineItemTags.ENCHANTABLE_SPELL_INFINITY); }
    @Override public boolean canAccept(Enchantment other) { return !(other instanceof MendingEnchantment) && super.canAccept(other); }
}
''')

forge_mod = root / 'forge/src/main/java/net/spell_engine/forge/ForgeMod.java'
fm = forge_mod.read_text()
if 'import net.spell_engine.compat.enchantment.SpellInfinityEnchantment1201;' not in fm:
    fm = fm.replace('import net.spell_engine.SpellEngineMod;\n',
                    'import net.spell_engine.SpellEngineMod;\nimport net.spell_engine.compat.enchantment.SpellInfinityEnchantment1201;\n', 1)
spell_infinity_registration = 'event.register(RegistryKeys.ENCHANTMENT, helper -> helper.register(SpellInfinityEnchantment1201.ID, SpellInfinityEnchantment1201.INSTANCE));'
if spell_infinity_registration not in fm:
    anchor = '        event.register(RegistryKeys.STATUS_EFFECT, helper -> SpellEngineEffects.register());\n'
    if anchor not in fm:
        raise SystemExit('pass6i Forge ENCHANTMENT registration anchor missing')
    fm = fm.replace(anchor, anchor + '        ' + spell_infinity_registration + '\n', 1)

# CI-only packaged-runtime assertions. GitHub Actions exports CI=true to the child Java process; normal
# user launches do not, so the release JAR carries the verifier but never executes it for players.
self_test = root / 'forge/src/main/java/net/spell_engine/forge/SpellEngineCiSelfTest.java'
self_test.write_text(r'''package net.spell_engine.forge;

import net.minecraft.item.ItemStack;
import net.minecraft.item.Items;
import net.minecraft.nbt.NbtCompound;
import net.minecraft.registry.RegistryKeys;
import net.minecraft.server.MinecraftServer;
import net.minecraft.util.Identifier;
import net.spell_engine.SpellEngineMod;
import net.spell_engine.api.spell.SpellDataComponents;
import net.spell_engine.api.spell.registry.SpellRegistry;
import net.spell_engine.compat.enchantment.SpellInfinityEnchantment1201;

final class SpellEngineCiSelfTest {
    private SpellEngineCiSelfTest() { }

    static void run(MinecraftServer server) {
        var spells = server.getRegistryManager().get(SpellRegistry.KEY);
        if (spells == null || spells.size() <= 0) {
            throw new IllegalStateException("Spell Engine CI self-test: synced SpellRegistry missing or empty");
        }

        var enchantments = server.getRegistryManager().get(RegistryKeys.ENCHANTMENT);
        if (enchantments.get(SpellInfinityEnchantment1201.ID) == null) {
            throw new IllegalStateException("Spell Engine CI self-test: spell_engine:spell_infinity missing");
        }

        var expected = new Identifier(SpellEngineMod.ID, "ci_item_model_roundtrip");
        var stack = new ItemStack(Items.STICK);
        SpellDataComponents.set(stack, SpellDataComponents.ITEM_MODEL, expected);
        var encoded = stack.writeNbt(new NbtCompound());
        var restored = ItemStack.fromNbt(encoded);
        var actual = SpellDataComponents.get(restored, SpellDataComponents.ITEM_MODEL);
        if (!expected.equals(actual)) {
            throw new IllegalStateException("Spell Engine CI self-test: NBT component round-trip failed: " + actual);
        }

        System.out.println("[Spell Engine CI] Packaged runtime self-test passed: SpellRegistry=" + spells.size()
                + ", SpellInfinity=present, NBT-components=roundtrip");
    }
}
''')
if 'SpellEngineCiSelfTest::run' not in fm:
    constructor_anchor = '        modBus.addListener(PlatformEventsImpl::onCreativeTab);\n'
    if constructor_anchor not in fm:
        raise SystemExit('pass6i CI self-test constructor anchor missing')
    fm = fm.replace(constructor_anchor, constructor_anchor + '''        if ("true".equalsIgnoreCase(System.getenv("CI"))) {\n            PlatformEventsImpl.onServerStarting(SpellEngineCiSelfTest::run);\n        }\n''', 1)
forge_mod.write_text(fm)

mixins.write_text(json.dumps(data, indent=2) + '\n')

# Loom dev launches discover mixins from the run configuration; installed Forge discovers them from
# the packaged JAR manifest. Apply this to every Jar task so shadowJar/remapJar preserve the config.
fb = forge_build.read_text()
if "attributes 'MixinConfigs': 'spell_engine.mixins.json'" not in fb:
    fb += r'''

// Real Forge installations do not inherit Loom's dev-run mixin bootstrap.
tasks.withType(Jar).configureEach {
    manifest { attributes 'MixinConfigs': 'spell_engine.mixins.json' }
}
'''
forge_build.write_text(fb)

gate = root / 'COMPAT-GATE-PENDING.md'
if gate.exists():
    lines = gate.read_text().splitlines()
    lines = [line for line in lines if '`effect.PoisonEffectMixin`' not in line and '`client.render.ImmediateItemGlowMixin`' not in line]
    gate.write_text('\n'.join(lines).rstrip() + '\n')

for required in ('if ((Object)this != StatusEffects.POISON) return;', 'Math.min(amplifiedAmount, entity.getHealth() - 1.0F)', 'duration % 25 == 0'):
    if required not in poison.read_text(): raise SystemExit(f'pass6i missing 1.20.1 poison parity: {required}')
for required in ('Map<RenderLayer, BufferBuilder> layerBuffers', 'new BufferBuilder(renderLayer.getExpectedBufferSize())', 'CustomLayers.isItemGlowLayer(renderLayer)'):
    if required not in glow.read_text(): raise SystemExit(f'pass6i missing 1.20.1 item-glow ordering parity: {required}')
final = json.loads(mixins.read_text())
for required in ('effect.StatusEffectActionImpairing', 'effect.StatusEffectSynchronized', 'action_impair.LivingEntityActionImpairing', 'loot.LootPoolAccessor'):
    if required not in final.get('mixins', []): raise SystemExit(f'pass6i required common mixin missing: {required}')
if 'effect.PoisonEffectMixin' not in final.get('mixins', []): raise SystemExit('pass6i did not activate poison mixin')
if 'client.render.ImmediateItemGlowMixin' not in final.get('client', []): raise SystemExit('pass6i did not activate item glow mixin')
if 'LootPoolEntry[] spellEngine_getEntries();' not in loot_pool_accessor.read_text(): raise SystemExit('pass6i LootPool accessor descriptor missing')
if '((LootPoolAccessor) (Object) pool).spellEngine_getEntries()' not in loot_helper.read_text(): raise SystemExit('pass6i LootHelper lacks packaged-safe LootPool accessor')
if loot_helper.read_text().count('for (var itemInjectorEntry: pool.entries)') != 1: raise SystemExit('pass6i changed LootConfig tag-cache entries')
if loot_helper.read_text().count('for (var entry: pool.entries)') != 1: raise SystemExit('pass6i changed LootConfig buildPool entries')
for required in ('new EquipmentSlot[]{EquipmentSlot.MAINHAND}', 'SpellEngineItemTags.ENCHANTABLE_SPELL_INFINITY', 'return 20;', 'return 50;', 'MendingEnchantment'):
    if required not in spell_infinity.read_text(): raise SystemExit(f'pass6i missing Spell Infinity parity: {required}')
if spell_infinity_registration not in forge_mod.read_text(): raise SystemExit('pass6i did not register spell_infinity')
for required in ('SpellEngineCiSelfTest::run', 'SpellDataComponents.ITEM_MODEL', 'ItemStack.fromNbt(encoded)', 'get(SpellRegistry.KEY)', 'get(RegistryKeys.ENCHANTMENT)', 'Packaged runtime self-test passed'):
    if required not in (forge_mod.read_text() + self_test.read_text()): raise SystemExit(f'pass6i missing packaged runtime self-test: {required}')
final_build = forge_build.read_text()
for required in ('tasks.withType(Jar).configureEach', "attributes 'MixinConfigs': 'spell_engine.mixins.json'"):
    if required not in final_build: raise SystemExit(f'pass6i missing packaged Forge mixin bootstrap: {required}')
if gate.exists():
    text = gate.read_text()
    if '`effect.PoisonEffectMixin`' in text or '`client.render.ImmediateItemGlowMixin`' in text: raise SystemExit('pass6i left implemented parity in pending gate')

print('Spell Engine compatibility pass 6i applied: packaged parity + runtime self-test')
