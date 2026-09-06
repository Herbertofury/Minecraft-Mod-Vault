package com.herbertofury.spellbrook.compat.hexerei;

import net.minecraftforge.client.event.EntityRenderersEvent;
import net.minecraftforge.eventbus.api.IEventBus;

public final class HexereiCompatClient {
    private HexereiCompatClient(){}
    public static void register(IEventBus modBus){modBus.addListener(HexereiCompatClient::renderers);}
    private static void renderers(EntityRenderersEvent.RegisterRenderers event){event.registerEntityRenderer(HexereiCompatBootstrap.BROOM.get(),HexereiSpellbrookBroomRenderer::new);}
}
