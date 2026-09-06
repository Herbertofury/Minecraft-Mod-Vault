#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 2:
    raise SystemExit('usage: prepare_more_rpg_library_1201_client_lifecycle.py <prepared-port-root>')

root = Path(sys.argv[1]).resolve()
common_client = root / 'common/src/main/java/net/more_rpg_classes/client/MoreRPGClassesClient.java'
beam = root / 'common/src/main/java/net/more_rpg_classes/client/render/MobBeamWorldRenderer.java'
forge_mod = root / 'forge/src/main/java/net/more_rpg_classes/forge/ForgeMod.java'
client_dir = root / 'forge/src/main/java/net/more_rpg_classes/forge/client'
mod_client = client_dir / 'ForgeClientMod.java'
game_client = client_dir / 'ForgeClientEvents.java'

for path in (common_client, beam, forge_mod):
    if not path.is_file():
        raise SystemExit(f'More RPG client lifecycle input missing: {path}')
if mod_client.exists() or game_client.exists():
    raise SystemExit('More RPG Forge client lifecycle unexpectedly already present')

client_source = common_client.read_text()
beam_source = beam.read_text()
forge_source = forge_mod.read_text()

# Fail closed on the exact current-2.7.2 behaviors that the loader seam must activate.
client_contracts = {
    'public static void init()': 1,
    'HeartTypes.getHeartTypes().forEach(HeartRegistry::register);': 1,
    'SpellTooltip.addDescriptionMutator': 1,
    'CustomModelRegistry.modelIds.add': 3,
    'public static void registerEntityRenderers': 1,
    'public static void registerParticleAppearances': 1,
}
for needle, expected in client_contracts.items():
    actual = client_source.count(needle)
    if actual != expected:
        raise SystemExit(f'More RPG current client contract drifted for {needle!r}: expected {expected}, found {actual}')
if beam_source.count('public static void render(MatrixStack matrices, Camera camera, float delta)') != 1:
    raise SystemExit('More RPG current MobBeamWorldRenderer.render contract drifted')
if beam_source.count('public static void onDisconnect()') != 1:
    raise SystemExit('More RPG current MobBeamWorldRenderer.onDisconnect contract drifted')
if 'net.more_rpg_classes.client.' in forge_source or 'net.minecraft.client.' in forge_source:
    raise SystemExit('More RPG common ForgeMod already hard-links physical-client classes')

client_dir.mkdir(parents=True, exist_ok=True)
mod_client.write_text(r'''package net.more_rpg_classes.forge.client;

import net.minecraft.client.particle.ParticleFactory;
import net.minecraft.client.render.entity.EntityRendererFactory;
import net.minecraft.entity.Entity;
import net.minecraft.entity.EntityType;
import net.minecraft.particle.ParticleEffect;
import net.minecraft.particle.ParticleType;
import net.minecraftforge.api.distmarker.Dist;
import net.minecraftforge.client.event.EntityRenderersEvent;
import net.minecraftforge.client.event.RegisterParticleProvidersEvent;
import net.minecraftforge.eventbus.api.SubscribeEvent;
import net.minecraftforge.fml.common.Mod;
import net.minecraftforge.fml.event.lifecycle.FMLClientSetupEvent;
import net.more_rpg_classes.MRPGCMod;
import net.more_rpg_classes.client.MoreRPGClassesClient;

/** Physical-client MOD-bus lifecycle for the current More RPG Library 2.7.2 client contract. */
@Mod.EventBusSubscriber(modid = MRPGCMod.MOD_ID, bus = Mod.EventBusSubscriber.Bus.MOD, value = Dist.CLIENT)
public final class ForgeClientMod {
    private ForgeClientMod() { }

    @SubscribeEvent
    public static void onClientSetup(FMLClientSetupEvent event) {
        // Current 2.7.2 NeoForge authority performs this directly during client setup. It registers
        // HeartTypes/HeartRegistry, spell-tooltip mutators, status-effect models and particles.
        MoreRPGClassesClient.init();
    }

    @SubscribeEvent
    public static void registerRenderers(EntityRenderersEvent.RegisterRenderers event) {
        MoreRPGClassesClient.registerEntityRenderers(new MoreRPGClassesClient.EntityRendererRegistrar() {
            @Override
            public <T extends Entity> void register(EntityType<? extends T> type, EntityRendererFactory<T> factory) {
                event.registerEntityRenderer(type, factory);
            }
        });
    }

    @SubscribeEvent
    public static void registerParticleProviders(RegisterParticleProvidersEvent event) {
        MoreRPGClassesClient.registerParticleAppearances(new MoreRPGClassesClient.ParticleFactoryRegistrar() {
            @Override
            public <T extends ParticleEffect> void register(ParticleType<T> type, ParticleFactory<T> factory) {
                event.registerSpecial(type, factory);
            }

            @Override
            public <T extends ParticleEffect> void registerSpriteAware(
                    ParticleType<T> type, MoreRPGClassesClient.SpriteAwareParticleFactory<T> factory) {
                event.registerSpriteSet(type, factory::create);
            }
        });
    }
}
''')

game_client.write_text(r'''package net.more_rpg_classes.forge.client;

import net.minecraftforge.api.distmarker.Dist;
import net.minecraftforge.client.event.ClientPlayerNetworkEvent;
import net.minecraftforge.client.event.RenderLevelStageEvent;
import net.minecraftforge.eventbus.api.SubscribeEvent;
import net.minecraftforge.fml.common.Mod;
import net.more_rpg_classes.MRPGCMod;
import net.more_rpg_classes.client.render.MobBeamWorldRenderer;

/** Physical-client FORGE-bus lifecycle for current More RPG world-render and disconnect behavior. */
@Mod.EventBusSubscriber(modid = MRPGCMod.MOD_ID, bus = Mod.EventBusSubscriber.Bus.FORGE, value = Dist.CLIENT)
public final class ForgeClientEvents {
    private ForgeClientEvents() { }

    @SubscribeEvent
    public static void onRenderLevelStage(RenderLevelStageEvent event) {
        if (event.getStage() != RenderLevelStageEvent.Stage.AFTER_TRANSLUCENT_BLOCKS) return;
        // Forge 47 / Minecraft 1.20.1 exposes the target-native float partial tick directly.
        MobBeamWorldRenderer.render(event.getPoseStack(), event.getCamera(), event.getPartialTick());
    }

    @SubscribeEvent
    public static void onLoggingOut(ClientPlayerNetworkEvent.LoggingOut event) {
        MobBeamWorldRenderer.onDisconnect();
    }
}
''')

mod_after = mod_client.read_text()
game_after = game_client.read_text()
contracts = {
    'MoreRPGClassesClient.init();': (mod_after.count('MoreRPGClassesClient.init();'), 1),
    'MoreRPGClassesClient.registerEntityRenderers': (mod_after.count('MoreRPGClassesClient.registerEntityRenderers'), 1),
    'MoreRPGClassesClient.registerParticleAppearances': (mod_after.count('MoreRPGClassesClient.registerParticleAppearances'), 1),
    'event.registerSpecial(type, factory);': (mod_after.count('event.registerSpecial(type, factory);'), 1),
    'event.registerSpriteSet(type, factory::create);': (mod_after.count('event.registerSpriteSet(type, factory::create);'), 1),
    'Mod.EventBusSubscriber.Bus.MOD': (mod_after.count('Mod.EventBusSubscriber.Bus.MOD'), 1),
    'Dist.CLIENT': (mod_after.count('Dist.CLIENT'), 1),
    'RenderLevelStageEvent.Stage.AFTER_TRANSLUCENT_BLOCKS': (game_after.count('RenderLevelStageEvent.Stage.AFTER_TRANSLUCENT_BLOCKS'), 1),
    'MobBeamWorldRenderer.render(event.getPoseStack(), event.getCamera(), event.getPartialTick());': (game_after.count('MobBeamWorldRenderer.render(event.getPoseStack(), event.getCamera(), event.getPartialTick());'), 1),
    'MobBeamWorldRenderer.onDisconnect();': (game_after.count('MobBeamWorldRenderer.onDisconnect();'), 1),
    'Mod.EventBusSubscriber.Bus.FORGE': (game_after.count('Mod.EventBusSubscriber.Bus.FORGE'), 1),
    'Dist.CLIENT': (game_after.count('Dist.CLIENT'), 1),
}
# Dist.CLIENT appears in each class; validate the duplicate key separately because dict keys are unique.
for needle, (actual, expected) in list(contracts.items()):
    if actual != expected:
        raise SystemExit(f'More RPG Forge client lifecycle contract failed for {needle!r}: expected {expected}, found {actual}')
if mod_after.count('Dist.CLIENT') != 1 or game_after.count('Dist.CLIENT') != 1:
    raise SystemExit('More RPG Forge client lifecycle is not exactly physical-client scoped')
for source, label in ((mod_after, 'mod-bus'), (game_after, 'forge-bus')):
    if 'net.neoforged.' in source or 'net.fabricmc.' in source:
        raise SystemExit(f'More RPG {label} client lifecycle retained non-Forge loader imports')

print('[More RPG 2.7.2] FORGE_CLIENT_LIFECYCLE_1201_PASS init=heart+tooltip+models renderers=true particles=true beam=after_translucent_blocks logout=disconnect physical_client=true')
