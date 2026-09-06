package com.herbertofury.spellbrook.client;

import com.herbertofury.spellbrook.Spellbrook;
import com.herbertofury.spellbrook.entity.SpellbrookBroomEntity;
import com.herbertofury.spellbrook.network.BroomInputPacket;
import com.herbertofury.spellbrook.network.SpellbrookNetwork;
import net.minecraft.client.Minecraft;
import net.minecraftforge.api.distmarker.Dist;
import net.minecraftforge.event.TickEvent;
import net.minecraftforge.eventbus.api.SubscribeEvent;
import net.minecraftforge.fml.common.Mod;

@Mod.EventBusSubscriber(modid=Spellbrook.MOD_ID,value=Dist.CLIENT)
public final class ClientForgeEvents {
    private static int lastEntity=-1,lastMask=-1;
    @SubscribeEvent public static void clientTick(TickEvent.ClientTickEvent event){if(event.phase!=TickEvent.Phase.END)return;Minecraft mc=Minecraft.getInstance();if(mc.player==null)return;if(!(mc.player.getVehicle() instanceof SpellbrookBroomEntity broom)){lastEntity=-1;lastMask=-1;return;}int mask=0;if(mc.options.keyUp.isDown())mask|=1;if(mc.options.keyDown.isDown())mask|=2;if(mc.options.keyLeft.isDown())mask|=4;if(mc.options.keyRight.isDown())mask|=8;if(mc.options.keyJump.isDown())mask|=16;if(mc.options.keyShift.isDown())mask|=32;if(mc.options.keySprint.isDown())mask|=64;if(broom.getId()!=lastEntity||mask!=lastMask){SpellbrookNetwork.CHANNEL.sendToServer(new BroomInputPacket(broom.getId(),mask));lastEntity=broom.getId();lastMask=mask;}}
}
