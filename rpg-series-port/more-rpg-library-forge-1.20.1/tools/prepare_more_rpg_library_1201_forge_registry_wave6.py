#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 2:
    raise SystemExit('usage: prepare_more_rpg_library_1201_forge_registry_wave6.py <prepared-port-root>')

root = Path(sys.argv[1]).resolve()
forge_mod = root / 'forge/src/main/java/net/more_rpg_classes/forge/ForgeMod.java'
if not forge_mod.is_file():
    raise SystemExit(f'generated Forge entrypoint missing: {forge_mod}')

before = forge_mod.read_text()
required_preconditions = {
    '@Mod(MRPGCMod.MOD_ID)': 1,
    'MoreRpgPlatform.isModLoaded = id -> ModList.get().isLoaded(id);': 1,
    'MoreRpgPlatform.isDevelopmentEnvironment = () -> !FMLLoader.isProduction();': 1,
    'MRPGCMod.init();': 1,
}
for needle, expected in required_preconditions.items():
    actual = before.count(needle)
    if actual != expected:
        raise SystemExit(f'Forge registry wave6 precondition drifted for {needle!r}: expected {expected}, found {actual}')
for forbidden in ('RegisterEvent', 'RegistryKeys.STATUS_EFFECT', 'MRPGCMod.registerEffects()'):
    if forbidden in before:
        raise SystemExit(f'Forge registry wave6 unexpectedly already present: {forbidden}')

forge_mod.write_text(r'''package net.more_rpg_classes.forge;

import net.minecraft.registry.RegistryKeys;
import net.more_rpg_classes.MRPGCMod;
import net.more_rpg_classes.client.particle.MoreParticles;
import net.more_rpg_classes.compat.MoreRpgPlatform;
import net.minecraftforge.fml.ModList;
import net.minecraftforge.fml.common.Mod;
import net.minecraftforge.fml.javafmlmod.FMLJavaModLoadingContext;
import net.minecraftforge.fml.loading.FMLLoader;
import net.minecraftforge.registries.RegisterEvent;

@Mod(MRPGCMod.MOD_ID)
public final class ForgeMod {
    public ForgeMod() {
        MoreRpgPlatform.isModLoaded = id -> ModList.get().isLoaded(id);
        MoreRpgPlatform.isDevelopmentEnvironment = () -> !FMLLoader.isProduction();

        // Forge 1.20.1 registry writes must happen while the matching registry is open. Mirror the
        // pinned 2.7.2 NeoForge loader's seven registry families rather than registering common
        // content from the mod constructor. Common non-registry initialization remains unchanged.
        FMLJavaModLoadingContext.get().getModEventBus().addListener(ForgeMod::register);
        MRPGCMod.init();
    }

    private static void register(RegisterEvent event) {
        event.register(RegistryKeys.LOOT_FUNCTION_TYPE, helper -> MRPGCMod.registerLootFunction());
        event.register(RegistryKeys.SOUND_EVENT, helper -> MRPGCMod.registerSounds());
        event.register(RegistryKeys.ITEM, helper -> MRPGCMod.registerItems());
        event.register(RegistryKeys.STATUS_EFFECT, helper -> MRPGCMod.registerEffects());
        event.register(RegistryKeys.PARTICLE_TYPE, helper -> MoreParticles.register());
        event.register(RegistryKeys.ENTITY_TYPE, helper -> MRPGCMod.registerEntities());
        event.register(RegistryKeys.STRUCTURE_TYPE, helper -> MRPGCMod.registerStructures());
    }
}
''')

after = forge_mod.read_text()
contracts = {
    'FMLJavaModLoadingContext.get().getModEventBus().addListener(ForgeMod::register);': 1,
    'event.register(RegistryKeys.LOOT_FUNCTION_TYPE': 1,
    'event.register(RegistryKeys.SOUND_EVENT': 1,
    'event.register(RegistryKeys.ITEM': 1,
    'event.register(RegistryKeys.STATUS_EFFECT': 1,
    'event.register(RegistryKeys.PARTICLE_TYPE': 1,
    'event.register(RegistryKeys.ENTITY_TYPE': 1,
    'event.register(RegistryKeys.STRUCTURE_TYPE': 1,
    'MRPGCMod.registerEffects()': 1,
    'MRPGCMod.init();': 1,
}
for needle, expected in contracts.items():
    actual = after.count(needle)
    if actual != expected:
        raise SystemExit(f'Forge registry wave6 contract failed for {needle!r}: expected {expected}, found {actual}')

families = [
    'LOOT_FUNCTION_TYPE',
    'SOUND_EVENT',
    'ITEM',
    'STATUS_EFFECT',
    'PARTICLE_TYPE',
    'ENTITY_TYPE',
    'STRUCTURE_TYPE',
]
if sum(after.count(f'event.register(RegistryKeys.{name}') for name in families) != 7:
    raise SystemExit('Forge registry wave6 family cardinality is not exactly seven')

print('[More RPG 2.7.2] FORGE_REGISTRY_EVENTS_1201_PASS '
      'families=7 loot_function+sound+item+status_effect+particle+entity+structure fatal_poison_registry_path=enabled')
