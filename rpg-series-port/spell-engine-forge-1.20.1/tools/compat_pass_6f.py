#!/usr/bin/env python3
from pathlib import Path
import json, sys

if len(sys.argv) != 3:
    raise SystemExit('usage: compat_pass_6f.py <generated-port-root> <spell-engine-1.20.1-baseline>')
root = Path(sys.argv[1]).resolve()
common_java = root / 'common/src/main/java'
common_res = root / 'common/src/main/resources'
forge_java = root / 'forge/src/main/java/net/spell_engine/forge'

# The modern LootTableModifyContext carried a WrapperLookup that LootHelper 1.10.2 no longer uses
# anywhere. Remove that dead parameter from the loader-neutral contract instead of inventing a fake
# registry lookup on Forge 1.20.1's LootTableLoadEvent (which does not provide one).
platform_events = common_java / 'net/spell_engine/PlatformEvents.java'
s = platform_events.read_text()
s = s.replace('import net.minecraft.registry.RegistryWrapper;\n', '')
s = s.replace('        RegistryWrapper.WrapperLookup registries();\n', '')
platform_events.write_text(s)

core = common_java / 'net/spell_engine/rpg_series/RPGSeriesCore.java'
s = core.read_text().replace('LootHelper.configure(context.registries(), context.tableId(),', 'LootHelper.configure(context.tableId(),')
core.write_text(s)

loot_helper = common_java / 'net/spell_engine/rpg_series/loot/LootHelper.java'
s = loot_helper.read_text()
s = s.replace('import net.minecraft.registry.RegistryWrapper;\n', '')
s = s.replace('public static void configure(RegistryWrapper.WrapperLookup registries, Identifier lootTableId,', 'public static void configure(Identifier lootTableId,')
s = s.replace('configureFallback(registries, tableId,', 'configureFallback(tableId,')
s = s.replace('buildPool(registries, pool,', 'buildPool(pool,')
s = s.replace('private static void configureFallback(RegistryWrapper.WrapperLookup registries, String tableId,', 'private static void configureFallback(String tableId,')
s = s.replace('buildPool(registries, fallback,', 'buildPool(fallback,')
s = s.replace('private static LootPool buildPool(RegistryWrapper.WrapperLookup registries, LootConfig.Pool pool,', 'private static LootPool buildPool(LootConfig.Pool pool,')
if 'RegistryWrapper.WrapperLookup registries' in s or 'buildPool(registries' in s or 'configureFallback(registries' in s:
    raise SystemExit('dead loot registry parameter survived pass 6f')
loot_helper.write_text(s)

# Vanilla Yarn 1.20.1 stores built-table pools as LootPool[] (not the List used by newer patched
# implementations). Access exactly that array for fallback inspection; Forge's addPool handles writes.
accessor_dir = common_java / 'net/spell_engine/mixin/loot'
accessor_dir.mkdir(parents=True, exist_ok=True)
accessor_dir.joinpath('LootTableAccessor.java').write_text(r'''package net.spell_engine.mixin.loot;

import net.minecraft.loot.LootPool;
import net.minecraft.loot.LootTable;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.gen.Accessor;

@Mixin(LootTable.class)
public interface LootTableAccessor {
    @Accessor("pools")
    LootPool[] spellEngine_pools();
}
''')
mixins = common_res / 'spell_engine.mixins.json'
data = json.loads(mixins.read_text())
if 'loot.LootTableAccessor' not in data.get('mixins', []):
    data.setdefault('mixins', []).append('loot.LootTableAccessor')
mixins.write_text(json.dumps(data, indent=2) + '\n')

forge_java.joinpath('PlatformEventsImpl.java').write_text(r'''package net.spell_engine.forge;

import net.minecraft.item.ItemGroup;
import net.minecraft.item.ItemStack;
import net.minecraft.loot.LootPool;
import net.minecraft.registry.RegistryKey;
import net.minecraft.registry.entry.RegistryEntry;
import net.minecraft.enchantment.Enchantment;
import net.minecraft.server.MinecraftServer;
import net.minecraft.server.network.ServerPlayerEntity;
import net.minecraft.util.Identifier;
import net.minecraftforge.common.MinecraftForge;
import net.minecraftforge.event.BuildCreativeModeTabContentsEvent;
import net.minecraftforge.event.LootTableLoadEvent;
import net.minecraftforge.event.OnDatapackSyncEvent;
import net.minecraftforge.event.RegisterCommandsEvent;
import net.minecraftforge.event.entity.living.LivingHurtEvent;
import net.minecraftforge.event.entity.player.PlayerEvent;
import net.minecraftforge.event.server.ServerStartedEvent;
import net.minecraftforge.event.server.ServerStartingEvent;
import net.spell_engine.PlatformEvents;
import net.spell_engine.SpellEngineMod;
import net.spell_engine.api.util.TriState;
import net.spell_engine.internals.container.SpellAssignments;
import net.spell_engine.mixin.loot.LootTableAccessor;
import net.spell_engine.network.Packets;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.function.Consumer;

/** Native Forge 47 event bridge; callbacks remain loader-neutral in common code. */
public final class PlatformEventsImpl {
    private PlatformEventsImpl() { }

    public static void onServerStarting(Consumer<MinecraftServer> callback) {
        MinecraftForge.EVENT_BUS.addListener((ServerStartingEvent event) -> callback.accept(event.getServer()));
    }

    public static void onServerStarted(Consumer<MinecraftServer> callback) {
        MinecraftForge.EVENT_BUS.addListener((ServerStartedEvent event) -> callback.accept(event.getServer()));
    }

    public static void onDataPackReloadComplete(Runnable callback) {
        MinecraftForge.EVENT_BUS.addListener((OnDatapackSyncEvent event) -> callback.run());
    }

    public static void onPlayerJoin(Consumer<ServerPlayerEntity> callback) {
        MinecraftForge.EVENT_BUS.addListener((PlayerEvent.PlayerLoggedInEvent event) -> {
            if (event.getEntity() instanceof ServerPlayerEntity player) {
                if (ForgeNetwork.isReady(player)) {
                    ForgeNetwork.sendToPlayer(player, new Packets.ConfigSync(SpellEngineMod.config));
                    ForgeNetwork.sendToPlayer(player, new Packets.SpellRegistrySync(SpellAssignments.encoded));
                }
                callback.accept(player);
            }
        });
    }

    public static void onPlayerChangedWorld(Consumer<ServerPlayerEntity> callback) {
        MinecraftForge.EVENT_BUS.addListener((PlayerEvent.PlayerChangedDimensionEvent event) -> {
            if (event.getEntity() instanceof ServerPlayerEntity player) callback.accept(player);
        });
    }

    public static void onIncomingDamage(PlatformEvents.IncomingDamage callback) {
        MinecraftForge.EVENT_BUS.addListener((LivingHurtEvent event) ->
                callback.accept(event.getEntity(), event.getSource(), event.getAmount()));
    }

    public static void onCommandRegistration(PlatformEvents.CommandRegistration callback) {
        MinecraftForge.EVENT_BUS.addListener((RegisterCommandsEvent event) ->
                callback.register(event.getDispatcher(), event.getBuildContext(), event.getCommandSelection()));
    }

    public static void onLootTableModify(Consumer<PlatformEvents.LootTableModifyContext> callback) {
        MinecraftForge.EVENT_BUS.addListener((LootTableLoadEvent event) -> {
            var context = new ForgeLootContext(event);
            callback.accept(context);
            for (var pool : context.addedPools) event.getTable().addPool(pool);
        });
    }

    private static final Map<RegistryKey<ItemGroup>, List<PlatformEvents.ItemGroupModifier>> ITEM_GROUPS = new HashMap<>();

    public static void onItemGroupModify(RegistryKey<ItemGroup> group, PlatformEvents.ItemGroupModifier callback) {
        ITEM_GROUPS.computeIfAbsent(group, ignored -> new ArrayList<>()).add(callback);
    }

    static void onCreativeTab(BuildCreativeModeTabContentsEvent event) {
        var callbacks = ITEM_GROUPS.get(event.getTabKey());
        if (callbacks == null) return;
        for (var callback : callbacks) {
            callback.modify((ItemGroup.Entries) (Object) event, (ItemGroup.DisplayContext) (Object) event.getParameters());
        }
    }

    private static final List<PlatformEvents.AllowEnchanting> ENCHANTING = new ArrayList<>();

    public static void onAllowEnchanting(PlatformEvents.AllowEnchanting callback) { ENCHANTING.add(callback); }

    static TriState evaluateAllowEnchanting(RegistryEntry<Enchantment> enchantment, ItemStack stack) {
        var result = TriState.PASS;
        for (var callback : ENCHANTING) {
            switch (callback.allow(enchantment, stack)) {
                case DENY -> { return TriState.DENY; }
                case ALLOW -> result = TriState.ALLOW;
                case PASS -> { }
            }
        }
        return result;
    }

    private static final class ForgeLootContext implements PlatformEvents.LootTableModifyContext {
        private final LootTableLoadEvent event;
        private final List<LootPool> existingPools;
        private final List<LootPool> addedPools = new ArrayList<>();

        private ForgeLootContext(LootTableLoadEvent event) {
            this.event = event;
            this.existingPools = List.copyOf(Arrays.asList(((LootTableAccessor) (Object) event.getTable()).spellEngine_pools()));
        }

        @Override public Identifier tableId() { return event.getName(); }
        @Override public List<LootPool> existingPools() { return existingPools; }
        @Override public void addPool(LootPool pool) { addedPools.add(pool); }
    }
}
''')

forge_java.joinpath('ForgeEventBridge.java').write_text(r'''package net.spell_engine.forge;

/** Legacy scaffold intentionally empty; all event wiring lives in PlatformEventsImpl. */
final class ForgeEventBridge {
    private ForgeEventBridge() { }
}
''')

forge_mod = forge_java / 'ForgeMod.java'
s = forge_mod.read_text()
s = s.replace('    public ForgeMod() {\n        IEventBus modBus = FMLJavaModLoadingContext.get().getModEventBus();',
              '    public ForgeMod(FMLJavaModLoadingContext context) {\n        IEventBus modBus = context.getModEventBus();')
anchor = '        modBus.addListener((DataPackRegistryEvent.NewRegistry event) -> SyncedDataRegistrar.apply(event));\n'
if anchor not in s:
    raise SystemExit('ForgeMod event registration anchor missing')
s = s.replace(anchor, anchor + '        modBus.addListener(PlatformEventsImpl::onCreativeTab);\n')
forge_mod.write_text(s)

pe = forge_java.joinpath('PlatformEventsImpl.java').read_text()
for required in ('ServerStartingEvent', 'ServerStartedEvent', 'OnDatapackSyncEvent', 'PlayerLoggedInEvent',
                 'PlayerChangedDimensionEvent', 'LivingHurtEvent', 'RegisterCommandsEvent', 'LootTableLoadEvent',
                 'BuildCreativeModeTabContentsEvent', 'Packets.ConfigSync', 'Packets.SpellRegistrySync'):
    if required not in pe:
        raise SystemExit(f'missing Forge event behavior: {required}')
if 'FMLJavaModLoadingContext.get()' in forge_mod.read_text():
    raise SystemExit('deprecated FMLJavaModLoadingContext.get() survived pass 6f')
if 'registries()' in core.read_text() or 'RegistryWrapper.WrapperLookup' in loot_helper.read_text():
    raise SystemExit('fabricated loot registry lookup survived pass 6f')
print('Spell Engine compatibility pass 6f applied: native Forge lifecycle/player/damage/commands/loot/creative wiring + login sync')
