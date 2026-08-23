package net.runes.forge.client;

import net.minecraft.client.gui.screen.ingame.HandledScreens;
import net.minecraft.client.item.ModelPredicateProviderRegistry;
import net.minecraft.client.render.RenderLayer;
import net.minecraft.client.render.RenderLayers;
import net.minecraft.util.Identifier;
import net.minecraftforge.api.distmarker.Dist;
import net.minecraftforge.eventbus.api.SubscribeEvent;
import net.minecraftforge.fml.common.Mod;
import net.minecraftforge.fml.event.lifecycle.FMLClientSetupEvent;
import net.runes.RunesMod;
import net.runes.client.RuneCraftingScreen;
import net.runes.crafting.RuneCraftingBlock;
import net.runes.crafting.RuneCraftingScreenHandler;
import net.runes.crafting.RunePouchItem;
import net.runes.crafting.RunePouches;

@Mod.EventBusSubscriber(modid=RunesMod.ID,bus=Mod.EventBusSubscriber.Bus.MOD,value=Dist.CLIENT)
public final class ForgeClientMod {
    @SubscribeEvent
    public static void setup(FMLClientSetupEvent event){
        event.enqueueWork(()->{
            HandledScreens.register(RuneCraftingScreenHandler.HANDLER_TYPE,RuneCraftingScreen::new);
            RenderLayers.setRenderLayer(RuneCraftingBlock.INSTANCE,RenderLayer.getCutout());
            for(var entry:RunePouches.all()){
                ModelPredicateProviderRegistry.register(entry.item(),new Identifier("filled"),(stack,world,entity,seed)-> RunePouchItem.hasContents(stack)?1.0F:0.0F);
            }
        });
    }
}
