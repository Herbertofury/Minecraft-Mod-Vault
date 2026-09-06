package com.herbertofury.spellbrook.network;

import com.herbertofury.spellbrook.entity.SpellbrookBroomEntity;
import net.minecraft.network.FriendlyByteBuf;
import net.minecraft.server.level.ServerPlayer;
import net.minecraft.world.entity.Entity;
import net.minecraftforge.network.NetworkEvent;

import java.util.function.Supplier;

public record BroomInputPacket(int entityId,int mask) {
    public static void encode(BroomInputPacket msg,FriendlyByteBuf buf){buf.writeVarInt(msg.entityId);buf.writeByte(msg.mask);} public static BroomInputPacket decode(FriendlyByteBuf buf){return new BroomInputPacket(buf.readVarInt(),buf.readUnsignedByte());}
    public static void handle(BroomInputPacket msg,Supplier<NetworkEvent.Context> supplier){NetworkEvent.Context ctx=supplier.get();ctx.enqueueWork(()->{ServerPlayer player=ctx.getSender();if(player==null)return;Entity e=player.level().getEntity(msg.entityId);if(e instanceof SpellbrookBroomEntity broom && player.getVehicle()==broom)broom.setInputs((msg.mask&1)!=0,(msg.mask&2)!=0,(msg.mask&4)!=0,(msg.mask&8)!=0,(msg.mask&16)!=0,(msg.mask&32)!=0,(msg.mask&64)!=0);});ctx.setPacketHandled(true);}
}
