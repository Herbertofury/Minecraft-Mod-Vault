#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 3:
    raise SystemExit('usage: compat_pass_6a.py <generated-port-root> <spell-engine-1.20.1-baseline>')
root = Path(sys.argv[1]).resolve()
base = Path(sys.argv[2]).resolve()

# Forge 1.20.1 has no upstream Spell Engine loader module. Create a native Architectury-Loom
# Forge module while keeping common code in Yarn names, then fill loader semantics in follow-up passes.
settings = root / 'settings.gradle'
s = settings.read_text()
if "include 'forge'" not in s:
    s = s.rstrip() + "\ninclude 'forge'\n"
settings.write_text(s)

build = root / 'build.gradle'
s = build.read_text()
needle = "id 'architectury-plugin' version '3.4.+'"
if "com.github.johnrengelman.shadow" not in s:
    if needle not in s:
        raise SystemExit('root architectury plugin marker missing')
    s = s.replace(needle, needle + "\n id 'com.github.johnrengelman.shadow' version '8.1.1' apply false")
build.write_text(s)

forge = root / 'forge'
java = forge / 'src/main/java/net/spell_engine/forge'
res = forge / 'src/main/resources'
(java / 'network').mkdir(parents=True, exist_ok=True)
res.joinpath('META-INF').mkdir(parents=True, exist_ok=True)

forge.joinpath('build.gradle').write_text(r'''plugins {
    id 'com.github.johnrengelman.shadow'
}

architectury {
    platformSetupLoomIde()
    forge()
}

loom {
    accessWidenerPath = project(':common').loom.accessWidenerPath
    forge {
        convertAccessWideners = true
        extraAccessWideners.add loom.accessWidenerPath.get().asFile.name
    }
}

configurations {
    common { canBeResolved = true; canBeConsumed = false }
    compileClasspath.extendsFrom common
    runtimeClasspath.extendsFrom common
    developmentForge.extendsFrom common
    shadowCommon { canBeResolved = true; canBeConsumed = false }
}

dependencies {
    forge "net.minecraftforge:forge:$rootProject.forge_version"
    common(project(path: ':common', configuration: 'namedElements')) { transitive = false }
    shadowCommon project(path: ':common', configuration: 'transformProductionForge')

    modImplementation "dev.kosmx.player-anim:player-animation-lib-forge:$rootProject.player_anim_version"
    modImplementation "me.shedaniel.cloth:cloth-config-forge:$rootProject.cloth_config_version"
}

processResources {
    inputs.property 'version', project.version
    filesMatching('META-INF/mods.toml') { expand(project.properties) }
}

shadowJar {
    configurations = [project.configurations.shadowCommon]
    archiveClassifier = 'dev-shadow'
}

remapJar {
    inputFile.set shadowJar.archiveFile
    dependsOn shadowJar
}
''')

java.joinpath('PlatformImpl.java').write_text(r'''package net.spell_engine.forge;

import com.mojang.serialization.Codec;
import net.minecraft.entity.Entity;
import net.minecraft.entity.EntityType;
import net.minecraft.entity.LivingEntity;
import net.minecraft.entity.attribute.DefaultAttributeContainer;
import net.minecraft.network.packet.Packet;
import net.minecraft.registry.Registry;
import net.minecraft.registry.RegistryKey;
import net.minecraft.server.network.ServerPlayerEntity;
import net.minecraft.util.Identifier;
import net.minecraftforge.fml.ModList;
import net.minecraftforge.fml.loading.FMLLoader;
import net.spell_engine.Platform;

import java.util.Collection;

public final class PlatformImpl {
    public static Platform.Type getPlatformType() { return Platform.Type.FORGE; }

    public static final class ForgeUtil implements Platform.Util {
        @Override public boolean isModLoaded(String modid) { return ModList.get().isLoaded(modid); }
        @Override public boolean isDevelopmentEnvironment() { return !FMLLoader.isProduction(); }
        @Override public void awakeSlotModCompat() { }
        @Override public void sendVanillaPacket_S2C(ServerPlayerEntity player, Packet<?> packet) { player.networkHandler.send(packet); }
        @Override public void registerSummonedEntityAttributes(EntityType<? extends LivingEntity> type, DefaultAttributeContainer.Builder builder) {
            SummonedEntityAttributeRegistrar.buffer(type, builder);
        }
        @Override public boolean networkS2C_CanSend(ServerPlayerEntity player, Identifier packetId) { return ForgeNetwork.isReady(player); }
        @Override public void networkS2C_Send(ServerPlayerEntity player, Object payload) { ForgeNetwork.sendToPlayer(player, payload); }
        @Override public void networkC2S_Send(Object payload) { ForgeNetwork.sendToServer(payload); }
        @Override public <T> void registerSyncedDataRegistry(RegistryKey<Registry<T>> key, Codec<T> localCodec, Codec<T> networkCodec) {
            SyncedDataRegistrar.buffer(key, localCodec, networkCodec);
        }
        @Override public Collection<ServerPlayerEntity> tracking(Entity entity) { return ForgeTracking.players(entity); }
    }

    private static final Platform.Util UTIL = new ForgeUtil();
    public static Platform.Util util() { return UTIL; }
}
''')

java.joinpath('PlatformClientImpl.java').write_text(r'''package net.spell_engine.forge;

import net.minecraft.client.network.ClientPlayerEntity;
import net.minecraft.network.packet.Packet;
import net.spell_engine.PlatformClient;

public final class PlatformClientImpl {
    private static final PlatformClient.Util UTIL = new PlatformClient.Util() {
        @Override public void sendVanillaPacket_C2S(ClientPlayerEntity player, Packet<?> packet) { player.networkHandler.send(packet); }
    };
    public static PlatformClient.Util util() { return UTIL; }
}
''')

java.joinpath('ForgeTracking.java').write_text(r'''package net.spell_engine.forge;

import net.minecraft.entity.Entity;
import net.minecraft.server.network.ServerPlayerEntity;
import java.util.ArrayList;
import java.util.Collection;

/** Exact tracker access is isolated here so it can be validated against Forge 47's mapped server internals. */
final class ForgeTracking {
    private ForgeTracking() { }
    static Collection<ServerPlayerEntity> players(Entity entity) {
        // Pass 6b replaces this compile scaffold with the proven Forge 47 tracker path.
        return new ArrayList<>();
    }
}
''')

java.joinpath('SummonedEntityAttributeRegistrar.java').write_text(r'''package net.spell_engine.forge;

import net.minecraft.entity.EntityType;
import net.minecraft.entity.LivingEntity;
import net.minecraft.entity.attribute.DefaultAttributeContainer;
import net.minecraftforge.event.entity.EntityAttributeCreationEvent;
import java.util.LinkedHashMap;
import java.util.Map;

final class SummonedEntityAttributeRegistrar {
    private static final Map<EntityType<? extends LivingEntity>, DefaultAttributeContainer.Builder> PENDING = new LinkedHashMap<>();
    static void buffer(EntityType<? extends LivingEntity> type, DefaultAttributeContainer.Builder builder) { PENDING.put(type, builder); }
    static void apply(EntityAttributeCreationEvent event) { PENDING.forEach((type, builder) -> event.put(type, builder.build())); }
}
''')

java.joinpath('SyncedDataRegistrar.java').write_text(r'''package net.spell_engine.forge;

import com.mojang.serialization.Codec;
import net.minecraft.registry.Registry;
import net.minecraft.registry.RegistryKey;
import net.minecraftforge.registries.DataPackRegistryEvent;
import java.util.ArrayList;
import java.util.List;

final class SyncedDataRegistrar {
    private record Pending<T>(RegistryKey<Registry<T>> key, Codec<T> local, Codec<T> network) { }
    private static final List<Pending<?>> PENDING = new ArrayList<>();
    static <T> void buffer(RegistryKey<Registry<T>> key, Codec<T> local, Codec<T> network) { PENDING.add(new Pending<>(key, local, network)); }
    static void apply(DataPackRegistryEvent.NewRegistry event) { for (var pending : PENDING) register(event, pending); }
    private static <T> void register(DataPackRegistryEvent.NewRegistry event, Pending<T> pending) {
        event.dataPackRegistry(pending.key(), pending.local(), pending.network());
    }
}
''')

java.joinpath('ForgeNetwork.java').write_text(r'''package net.spell_engine.forge;

import net.minecraft.server.network.ServerPlayerEntity;

/** Forge 47 transport seam. Pass 6b installs all packet codecs/handlers after the loader module compiles. */
final class ForgeNetwork {
    private ForgeNetwork() { }
    static void init() { }
    static boolean isReady(ServerPlayerEntity player) { return true; }
    static void sendToPlayer(ServerPlayerEntity player, Object payload) {
        throw new IllegalStateException("Spell Engine Forge network transport has not been initialized");
    }
    static void sendToServer(Object payload) {
        throw new IllegalStateException("Spell Engine Forge network transport has not been initialized");
    }
}
''')

java.joinpath('PlatformEventsImpl.java').write_text(r'''package net.spell_engine.forge;

import net.minecraft.item.ItemGroup;
import net.minecraft.registry.RegistryKey;
import net.minecraft.server.MinecraftServer;
import net.minecraft.server.network.ServerPlayerEntity;
import net.spell_engine.PlatformEvents;
import java.util.function.Consumer;

/** Forge event-bus seam; compile scaffold first, then pass 6b binds every callback to Forge 47 events. */
public final class PlatformEventsImpl {
    public static void onServerStarting(Consumer<MinecraftServer> callback) { ForgeEventBridge.serverStarting.add(callback); }
    public static void onServerStarted(Consumer<MinecraftServer> callback) { ForgeEventBridge.serverStarted.add(callback); }
    public static void onDataPackReloadComplete(Runnable callback) { ForgeEventBridge.dataReload.add(callback); }
    public static void onPlayerJoin(Consumer<ServerPlayerEntity> callback) { ForgeEventBridge.playerJoin.add(callback); }
    public static void onPlayerChangedWorld(Consumer<ServerPlayerEntity> callback) { ForgeEventBridge.playerChangedWorld.add(callback); }
    public static void onIncomingDamage(PlatformEvents.IncomingDamage callback) { ForgeEventBridge.incomingDamage.add(callback); }
    public static void onCommandRegistration(PlatformEvents.CommandRegistration callback) { ForgeEventBridge.commands.add(callback); }
    public static void onLootTableModify(Consumer<PlatformEvents.LootTableModifyContext> callback) { ForgeEventBridge.loot.add(callback); }
    public static void onItemGroupModify(RegistryKey<ItemGroup> group, PlatformEvents.ItemGroupModifier callback) { ForgeEventBridge.addItemGroup(group, callback); }
    public static void onAllowEnchanting(PlatformEvents.AllowEnchanting callback) { ForgeEventBridge.enchanting.add(callback); }
}
''')

java.joinpath('ForgeEventBridge.java').write_text(r'''package net.spell_engine.forge;

import net.minecraft.item.ItemGroup;
import net.minecraft.registry.RegistryKey;
import net.minecraft.server.MinecraftServer;
import net.minecraft.server.network.ServerPlayerEntity;
import net.spell_engine.PlatformEvents;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.function.Consumer;

final class ForgeEventBridge {
    static final List<Consumer<MinecraftServer>> serverStarting = new ArrayList<>();
    static final List<Consumer<MinecraftServer>> serverStarted = new ArrayList<>();
    static final List<Runnable> dataReload = new ArrayList<>();
    static final List<Consumer<ServerPlayerEntity>> playerJoin = new ArrayList<>();
    static final List<Consumer<ServerPlayerEntity>> playerChangedWorld = new ArrayList<>();
    static final List<PlatformEvents.IncomingDamage> incomingDamage = new ArrayList<>();
    static final List<PlatformEvents.CommandRegistration> commands = new ArrayList<>();
    static final List<Consumer<PlatformEvents.LootTableModifyContext>> loot = new ArrayList<>();
    static final List<PlatformEvents.AllowEnchanting> enchanting = new ArrayList<>();
    static final Map<RegistryKey<ItemGroup>, List<PlatformEvents.ItemGroupModifier>> itemGroups = new HashMap<>();
    static void addItemGroup(RegistryKey<ItemGroup> key, PlatformEvents.ItemGroupModifier callback) {
        itemGroups.computeIfAbsent(key, ignored -> new ArrayList<>()).add(callback);
    }
    private ForgeEventBridge() { }
}
''')

java.joinpath('ForgeMod.java').write_text(r'''package net.spell_engine.forge;

import net.minecraft.registry.RegistryKeys;
import net.minecraftforge.event.entity.EntityAttributeCreationEvent;
import net.minecraftforge.eventbus.api.IEventBus;
import net.minecraftforge.fml.common.Mod;
import net.minecraftforge.fml.javafmlmod.FMLJavaModLoadingContext;
import net.minecraftforge.registries.DataPackRegistryEvent;
import net.minecraftforge.registries.RegisterEvent;
import net.spell_engine.SpellEngineMod;
import net.spell_engine.api.effect.SpellEngineEffects;
import net.spell_engine.fx.SpellEngineParticles;
import net.spell_engine.fx.SpellEngineSounds;
import net.spell_engine.item.SpellEngineItems;

@Mod(SpellEngineMod.ID)
public final class ForgeMod {
    public ForgeMod() {
        IEventBus modBus = FMLJavaModLoadingContext.get().getModEventBus();
        SpellEngineMod.init();
        ForgeNetwork.init();
        modBus.addListener(ForgeMod::register);
        modBus.addListener((EntityAttributeCreationEvent event) -> SummonedEntityAttributeRegistrar.apply(event));
        modBus.addListener((DataPackRegistryEvent.NewRegistry event) -> SyncedDataRegistrar.apply(event));
    }

    private static void register(RegisterEvent event) {
        event.register(RegistryKeys.ENTITY_TYPE, helper -> SpellEngineMod.registerEntityTypes());
        event.register(RegistryKeys.PARTICLE_TYPE, helper -> SpellEngineParticles.register());
        event.register(RegistryKeys.STATUS_EFFECT, helper -> SpellEngineEffects.register());
        event.register(RegistryKeys.ITEM, helper -> SpellEngineItems.register());
        event.register(RegistryKeys.SOUND_EVENT, helper -> SpellEngineSounds.register());
        event.register(RegistryKeys.BLOCK, helper -> SpellEngineMod.registerSpellBinding());
        event.register(RegistryKeys.CRITERION, helper -> SpellEngineMod.registerCriteria());
    }
}
''')

res.joinpath('META-INF/mods.toml').write_text(r'''modLoader="javafml"
loaderVersion="[47,)"
license="GPL-3.0-or-later"

[[mods]]
modId="spell_engine"
version="${version}"
displayName="Spell Engine"
description="Native Forge 1.20.1 port of Spell Engine 1.10.2."

[[dependencies.spell_engine]]
modId="forge"
mandatory=true
versionRange="[47.4.23,)"
ordering="NONE"
side="BOTH"

[[dependencies.spell_engine]]
modId="minecraft"
mandatory=true
versionRange="[1.20.1,1.20.2)"
ordering="NONE"
side="BOTH"
''')
res.joinpath('pack.mcmeta').write_text('{"pack":{"pack_format":15,"description":"Spell Engine Forge 1.20.1 resources"}}\n')

assert "include 'forge'" in settings.read_text()
assert forge.joinpath('build.gradle').exists()
assert java.joinpath('ForgeMod.java').exists()
assert res.joinpath('META-INF/mods.toml').exists()
print('Spell Engine compatibility pass 6a applied: native Forge 47 loader-module compile scaffold')
