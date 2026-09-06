package com.herbertofury.spellbrook.client;

import com.herbertofury.spellbrook.Spellbrook;
import com.herbertofury.spellbrook.registry.ModEntities;
import net.minecraftforge.api.distmarker.Dist;
import net.minecraftforge.client.event.EntityRenderersEvent;
import net.minecraftforge.eventbus.api.SubscribeEvent;
import net.minecraftforge.fml.common.Mod;

@Mod.EventBusSubscriber(modid=Spellbrook.MOD_ID,bus=Mod.EventBusSubscriber.Bus.MOD,value=Dist.CLIENT)
public final class ClientModEvents {
    @SubscribeEvent public static void registerRenderers(EntityRenderersEvent.RegisterRenderers event){event.registerEntityRenderer(ModEntities.BROOM.get(),SpellbrookBroomRenderer::new);}
}
