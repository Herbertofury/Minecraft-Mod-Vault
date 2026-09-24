package net.rogues.forge.client;

import net.minecraftforge.api.distmarker.Dist;
import net.minecraftforge.client.ConfigScreenHandler;
import net.minecraftforge.client.event.EntityRenderersEvent;
import net.minecraftforge.eventbus.api.SubscribeEvent;
import net.minecraftforge.fml.ModLoadingContext;
import net.minecraftforge.fml.common.Mod;
import net.minecraftforge.fml.event.lifecycle.FMLClientSetupEvent;
import net.rogues.RoguesMod;
import net.rogues.client.RoguesClientMod;
import net.rogues.client.entity.BearTrapEntityModel;
import net.rogues.client.entity.BearTrapEntityRenderer;
import net.rogues.entity.RogueEntities;
import net.spell_engine.client.gui.ConfigMenuScreen;

@Mod.EventBusSubscriber(modid=RoguesMod.ID,bus=Mod.EventBusSubscriber.Bus.MOD,value=Dist.CLIENT)
public final class ForgeClientMod {
    private ForgeClientMod(){}
    @SubscribeEvent public static void onClientSetup(FMLClientSetupEvent e){RoguesClientMod.init();ModLoadingContext.get().registerExtensionPoint(ConfigScreenHandler.ConfigScreenFactory.class,()->new ConfigScreenHandler.ConfigScreenFactory((minecraft,parent)->new ConfigMenuScreen(parent)));}
    @SubscribeEvent public static void onLayers(EntityRenderersEvent.RegisterLayerDefinitions e){e.registerLayerDefinition(BearTrapEntityModel.LAYER,BearTrapEntityModel::getTexturedModelData);}
    @SubscribeEvent public static void onRenderers(EntityRenderersEvent.RegisterRenderers e){e.registerEntityRenderer(RogueEntities.BEAR_TRAP.type,BearTrapEntityRenderer::new);}
}
