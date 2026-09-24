#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 3:
    raise SystemExit('usage: compat_pass_6n.py <generated-port-root> <spell-engine-1.20.1-baseline>')

root = Path(sys.argv[1]).resolve()
forge_java = root / 'forge/src/main/java/net/spell_engine/forge'
client_dir = forge_java / 'client'
client_dir.mkdir(parents=True, exist_ok=True)

# The common 1.10.2 client code is intentionally loader-neutral. The original Forge port created
# server/platform/network bridges but never installed a physical-client entrypoint, leaving
# SpellEngineClient.config null until the first real joined-client HUD/input tick. Restore the
# loader-owned client lifecycle here, using APIs available on Forge 47 / Minecraft 1.20.1.
client = client_dir / 'ForgeClientMod.java'
client.write_text(r'''package net.spell_engine.forge.client;

import net.minecraft.client.gui.screen.ingame.HandledScreens;
import net.minecraftforge.api.distmarker.Dist;
import net.minecraftforge.client.ConfigScreenHandler;
import net.minecraftforge.client.event.EntityRenderersEvent;
import net.minecraftforge.client.event.ModelEvent;
import net.minecraftforge.client.event.RegisterKeyMappingsEvent;
import net.minecraftforge.client.event.RegisterParticleProvidersEvent;
import net.minecraftforge.client.settings.KeyConflictContext;
import net.minecraftforge.eventbus.api.SubscribeEvent;
import net.minecraftforge.fml.ModLoadingContext;
import net.minecraftforge.fml.common.Mod;
import net.minecraftforge.fml.event.lifecycle.FMLClientSetupEvent;
import net.spell_engine.SpellEngineMod;
import net.spell_engine.client.SpellEngineClient;
import net.spell_engine.client.gui.ConfigMenuScreen;
import net.spell_engine.client.input.GuiKeyBinding;
import net.spell_engine.client.input.Keybindings;
import net.spell_engine.client.render.CustomModelRegistry;
import net.spell_engine.client.render.SpellCloudRenderer;
import net.spell_engine.client.render.SpellModelEffectRenderer;
import net.spell_engine.client.render.SpellProjectileRenderer;
import net.spell_engine.entity.SpellCloud;
import net.spell_engine.entity.SpellModelEffect;
import net.spell_engine.entity.SpellProjectile;
import net.spell_engine.spellbinding.SpellBindingScreen;
import net.spell_engine.spellbinding.SpellBindingScreenHandler;
import net.spell_engine.spellbinding.spellchoice.SpellChoiceScreen;
import net.spell_engine.spellbinding.spellchoice.SpellChoiceScreenHandler;

/**
 * Physical-client Forge 47 lifecycle for Spell Engine.
 *
 * Keep this class completely client-gated: dedicated servers must never resolve Minecraft client
 * classes. Common client initialization is invoked exactly once from FMLClientSetupEvent, matching
 * the upstream loader lifecycle while preserving the 1.20.1 HUD mixin and Forge network bridge.
 */
@Mod.EventBusSubscriber(modid = SpellEngineMod.ID, bus = Mod.EventBusSubscriber.Bus.MOD, value = Dist.CLIENT)
public final class ForgeClientMod {
    private ForgeClientMod() { }

    @SubscribeEvent
    public static void onClientSetup(FMLClientSetupEvent event) {
        SpellEngineClient.init();
        event.enqueueWork(() -> {
            SpellEngineClient.onClientStarted();
            HandledScreens.register(SpellBindingScreenHandler.HANDLER_TYPE, SpellBindingScreen::new);
            HandledScreens.register(SpellChoiceScreenHandler.HANDLER_TYPE, SpellChoiceScreen::new);
            ModLoadingContext.get().registerExtensionPoint(
                    ConfigScreenHandler.ConfigScreenFactory.class,
                    () -> new ConfigScreenHandler.ConfigScreenFactory(parent -> new ConfigMenuScreen(parent)));
        });
    }

    @SubscribeEvent
    public static void registerKeys(RegisterKeyMappingsEvent event) {
        for (var keybinding : Keybindings.all()) {
            if (keybinding instanceof GuiKeyBinding) {
                keybinding.setKeyConflictContext(KeyConflictContext.GUI);
            }
            event.register(keybinding);
        }
    }

    @SubscribeEvent
    public static void registerParticleProviders(RegisterParticleProvidersEvent event) {
        SpellEngineClient.registerParticleAppearances(new SpellEngineClient.ParticleAppearanceRegistrar() {
            @Override
            public <T extends net.minecraft.particle.ParticleEffect> void register(
                    net.minecraft.particle.ParticleType<T> type,
                    SpellEngineClient.SpriteFactory<T> factory) {
                event.registerSpriteSet(type, factory::create);
            }
        });
    }

    @SubscribeEvent
    public static void registerEntityRenderers(EntityRenderersEvent.RegisterRenderers event) {
        event.registerEntityRenderer(SpellProjectile.ENTITY_TYPE, SpellProjectileRenderer::new);
        event.registerEntityRenderer(SpellCloud.ENTITY_TYPE, SpellCloudRenderer::new);
        event.registerEntityRenderer(SpellModelEffect.ENTITY_TYPE, SpellModelEffectRenderer::new);
    }

    @SubscribeEvent
    public static void registerAdditionalModels(ModelEvent.RegisterAdditional event) {
        for (var id : CustomModelRegistry.getModelIds()) {
            event.register(id);
        }
    }
}
''')

text = client.read_text()
required = (
    'value = Dist.CLIENT',
    'SpellEngineClient.init();',
    'SpellEngineClient.onClientStarted();',
    'RegisterKeyMappingsEvent',
    'KeyConflictContext.GUI',
    'RegisterParticleProvidersEvent',
    'EntityRenderersEvent.RegisterRenderers',
    'HandledScreens.register(SpellBindingScreenHandler.HANDLER_TYPE',
    'HandledScreens.register(SpellChoiceScreenHandler.HANDLER_TYPE',
    'ConfigScreenHandler.ConfigScreenFactory',
    'ModelEvent.RegisterAdditional',
)
for needle in required:
    if needle not in text:
        raise SystemExit(f'Spell Engine Forge client lifecycle incomplete: {needle}')

print('Spell Engine compatibility pass 6n applied: Forge 47 physical-client lifecycle + registrations restored')
