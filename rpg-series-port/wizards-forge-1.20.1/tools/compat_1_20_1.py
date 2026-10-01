#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 2:
    raise SystemExit('usage: compat_1_20_1.py <generated-wizards-root>')

root = Path(sys.argv[1]).resolve()
forge_java = root / 'forge/src/main/java/net/wizards/forge'
client_dir = forge_java / 'client'
client_dir.mkdir(parents=True, exist_ok=True)


def replace_exact(path: Path, old: str, new: str, label: str) -> None:
    if not path.is_file():
        raise SystemExit(f'Wizards compatibility source missing for {label}: {path.relative_to(root)}')
    text = path.read_text()
    count = text.count(old)
    if count != 1:
        raise SystemExit(f'Wizards compatibility stale transform for {label}: expected 1 occurrence, found {count}')
    path.write_text(text.replace(old, new, 1))


# Forge 47 only opens the registry matching the active RegisterEvent. The current 3.1.1 NeoForge
# entrypoint registers entities while the ITEM registry is active and registerItems() also touches the
# ITEM_GROUP registry. NeoForge tolerates that path; Forge 47 correctly rejects it as a locked-registry
# mutation. Split the common operations so every vanilla registry mutation runs in its own matching event.
common_mod = root / 'common/src/main/java/net/wizards/WizardsMod.java'
old_items = '''    public static void registerItems() {
        Group.WIZARDS = new ItemGroup.Builder(ItemGroup.Row.TOP, 0)
                .icon(() -> new ItemStack(WizardArmors.wizardRobeSet.head))
                .displayName(Text.translatable("itemGroup.wizards.general"))
                .build();
        Registry.register(Registries.ITEM_GROUP, Group.KEY, Group.WIZARDS);
        WizardBooks.register();
        WizardWeapons.register(equipmentConfig.value.weapons);
        WizardArmors.register(equipmentConfig.value.armor_sets);
        equipmentConfig.save();
    }
'''
new_items = '''    public static void registerItemGroup() {
        Group.WIZARDS = new ItemGroup.Builder(ItemGroup.Row.TOP, 0)
                .icon(() -> new ItemStack(WizardArmors.wizardRobeSet.head))
                .displayName(Text.translatable("itemGroup.wizards.general"))
                .build();
        Registry.register(Registries.ITEM_GROUP, Group.KEY, Group.WIZARDS);
    }

    public static void registerItems() {
        WizardBooks.register();
        WizardWeapons.register(equipmentConfig.value.weapons);
        WizardArmors.register(equipmentConfig.value.armor_sets);
        equipmentConfig.save();
    }
'''
replace_exact(common_mod, old_items, new_items, 'Forge registry phase split in WizardsMod.registerItems')

forge_entry = forge_java / 'ForgeMod.java'
old_registry_phase = '''        event.register(RegistryKeys.ITEM, reg -> {
            WizardsMod.registerEntities();
            WizardsMod.registerItems();
        });
'''
new_registry_phase = '''        event.register(RegistryKeys.ENTITY_TYPE, reg -> {
            WizardsMod.registerEntities();
        });
        event.register(RegistryKeys.ITEM, reg -> {
            WizardsMod.registerItems();
        });
        event.register(RegistryKeys.ITEM_GROUP, reg -> {
            WizardsMod.registerItemGroup();
        });
'''
replace_exact(forge_entry, old_registry_phase, new_registry_phase, 'Forge ENTITY_TYPE/ITEM/ITEM_GROUP RegisterEvent split')

old_client = client_dir / 'NeoForgeClientMod.java'
new_client = client_dir / 'ForgeClientMod.java'
if old_client.exists():
    old_client.unlink()

# Forge 47.4.x uses Mod.EventBusSubscriber on the MOD bus for lifecycle/entity renderer events,
# MinecraftForge.EVENT_BUS for RenderLevelStageEvent, and ConfigScreenHandler.ConfigScreenFactory
# as its config-screen extension point. On the exact Forge 47.4.23/Yarn 1.20.1 mapping used by this
# lane, RenderLevelStageEvent#getPoseStack already remaps to MatrixStack. Reuse that event stack
# directly for the current Fire Hydra AFTER_PARTICLES replay; do not reconstruct a second stack.
new_client.write_text('''package net.wizards.forge.client;\n\nimport net.minecraft.client.util.math.MatrixStack;\nimport net.minecraftforge.api.distmarker.Dist;\nimport net.minecraftforge.client.ConfigScreenHandler.ConfigScreenFactory;\nimport net.minecraftforge.client.event.EntityRenderersEvent;\nimport net.minecraftforge.client.event.RenderLevelStageEvent;\nimport net.minecraftforge.common.MinecraftForge;\nimport net.minecraftforge.eventbus.api.SubscribeEvent;\nimport net.minecraftforge.fml.ModLoadingContext;\nimport net.minecraftforge.fml.common.Mod.EventBusSubscriber;\nimport net.minecraftforge.fml.event.lifecycle.FMLClientSetupEvent;\nimport net.spell_engine.client.gui.ConfigMenuScreen;\nimport net.wizards.WizardsMod;\nimport net.wizards.client.WizardsClientMod;\nimport net.wizards.client.entity.ArcaneEmitterModel;\nimport net.wizards.client.entity.ArcaneEmitterRenderer;\nimport net.wizards.client.entity.FireHydraModel;\nimport net.wizards.client.entity.FireHydraRenderer;\nimport net.wizards.client.entity.FrostElementalModel;\nimport net.wizards.client.entity.FrostElementalRenderer;\nimport net.wizards.entity.WizardEntities;\n\n@EventBusSubscriber(modid = WizardsMod.ID, value = Dist.CLIENT, bus = EventBusSubscriber.Bus.MOD)\npublic final class ForgeClientMod {\n    private ForgeClientMod() {}\n\n    @SubscribeEvent\n    public static void onClientSetup(FMLClientSetupEvent event) {\n        WizardsClientMod.init();\n        ModLoadingContext.get().registerExtensionPoint(\n                ConfigScreenFactory.class,\n                () -> new ConfigScreenFactory(parent -> new ConfigMenuScreen(parent)));\n\n        // Preserve current 3.1.1 Fire Hydra ordering: replay after particles on Forge's game bus.\n        MinecraftForge.EVENT_BUS.addListener((RenderLevelStageEvent render) -> {\n            if (render.getStage() == RenderLevelStageEvent.Stage.AFTER_PARTICLES) {\n                MatrixStack matrices = render.getPoseStack();\n                FireHydraRenderer.renderAfterTranslucent(\n                        matrices, render.getCamera(), render.getPartialTick());\n            }\n        });\n    }\n\n    @SubscribeEvent\n    public static void onRegisterLayerDefinitions(EntityRenderersEvent.RegisterLayerDefinitions event) {\n        event.registerLayerDefinition(FrostElementalModel.TEXTURE, FrostElementalModel::getTexturedModelData);\n        event.registerLayerDefinition(ArcaneEmitterModel.LAYER, ArcaneEmitterModel::getTexturedModelData);\n        event.registerLayerDefinition(FireHydraModel.LAYER, FireHydraModel::getTexturedModelData);\n    }\n\n    @SubscribeEvent\n    public static void onRegisterRenderers(EntityRenderersEvent.RegisterRenderers event) {\n        event.registerEntityRenderer(WizardEntities.FROST_ELEMENTAL.type, FrostElementalRenderer::new);\n        event.registerEntityRenderer(WizardEntities.ARCANE_EMITTER.type, ArcaneEmitterRenderer::new);\n        event.registerEntityRenderer(WizardEntities.FIRE_HYDRA.type, FireHydraRenderer::new);\n    }\n}\n''')

# Fail before Gradle if a platform-only namespace survives in loader code. Common source may still
# legitimately need later Minecraft/API compatibility passes, but generated Forge loader code must be native.
leaks = []
for path in (root / 'forge/src/main/java').rglob('*.java'):
    text = path.read_text()
    if 'net.neoforged' in text or 'NeoForge.' in text or 'NeoForgeClientMod' in text:
        leaks.append(str(path.relative_to(root)))
if leaks:
    raise SystemExit('NeoForge loader symbols survived Wizards Forge compatibility pass: ' + ', '.join(leaks))

print('Wizards compatibility pass 1 applied: Forge 47 registry-phase split + native client lifecycle/render/config wiring')
