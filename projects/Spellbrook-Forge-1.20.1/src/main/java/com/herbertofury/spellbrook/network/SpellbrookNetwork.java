package com.herbertofury.spellbrook.network;

import com.herbertofury.spellbrook.Spellbrook;
import net.minecraft.resources.ResourceLocation;
import net.minecraftforge.network.NetworkRegistry;
import net.minecraftforge.network.simple.SimpleChannel;

public final class SpellbrookNetwork {
    private static final String PROTOCOL="1";
    public static final SimpleChannel CHANNEL= NetworkRegistry.ChannelBuilder.named(new ResourceLocation(Spellbrook.MOD_ID,"main")).networkProtocolVersion(()->PROTOCOL).clientAcceptedVersions(PROTOCOL::equals).serverAcceptedVersions(PROTOCOL::equals).simpleChannel();
    private static boolean initialized;
    private SpellbrookNetwork(){}
    public static synchronized void init(){ if(initialized)return; initialized=true; CHANNEL.registerMessage(0,BroomInputPacket.class,BroomInputPacket::encode,BroomInputPacket::decode,BroomInputPacket::handle); }
}
