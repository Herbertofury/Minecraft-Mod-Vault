#!/usr/bin/env python3
import re, sys
from pathlib import Path

if len(sys.argv) != 2:
    raise SystemExit('usage: apply-particle120-network-bridge.py <java-root>')
root = Path(sys.argv[1]).resolve()
if not root.is_dir():
    raise SystemExit(f'missing Java root: {root}')

def write(rel, text):
    p = root / rel
    p.parent.mkdir(parents=True, exist_ok=True)
    p.write_text(text.strip() + '\n', encoding='utf-8')

N = 'team/lodestar/lodestone/systems/network/particle'
write(f'{N}/NetworkedParticleEffectExtraData.java', r'''
package team.lodestar.lodestone.systems.network.particle;
public interface NetworkedParticleEffectExtraData {}
''')
write(f'{N}/NetworkedParticleEffectPositionData.java', r'''
package team.lodestar.lodestone.systems.network.particle;

import com.mojang.serialization.Codec;
import com.mojang.serialization.codecs.RecordCodecBuilder;
import net.minecraft.core.BlockPos;
import net.minecraft.world.entity.Entity;
import net.minecraft.world.phys.Vec3;

public class NetworkedParticleEffectPositionData {
    public static final Codec<NetworkedParticleEffectPositionData> CODEC = RecordCodecBuilder.create(instance -> instance.group(
        Codec.DOUBLE.fieldOf("posX").forGetter(NetworkedParticleEffectPositionData::getPosX),
        Codec.DOUBLE.fieldOf("posY").forGetter(NetworkedParticleEffectPositionData::getPosY),
        Codec.DOUBLE.fieldOf("posZ").forGetter(NetworkedParticleEffectPositionData::getPosZ)
    ).apply(instance, NetworkedParticleEffectPositionData::new));

    protected final double posX, posY, posZ;
    public NetworkedParticleEffectPositionData(BlockPos pos) { this(pos.getX(), pos.getY(), pos.getZ()); }
    public NetworkedParticleEffectPositionData(Entity entity) { this(entity.getX(), entity.getY() + entity.getBbHeight() / 2f, entity.getZ()); }
    public NetworkedParticleEffectPositionData(Vec3 pos) { this(pos.x, pos.y, pos.z); }
    public NetworkedParticleEffectPositionData(double x, double y, double z) { posX=x; posY=y; posZ=z; }
    public BlockPos getAsBlockPos() { return new BlockPos((int)posX, (int)posY, (int)posZ); }
    public Vec3 getAsVector() { return new Vec3(posX, posY, posZ); }
    public double getPosX() { return posX; }
    public double getPosY() { return posY; }
    public double getPosZ() { return posZ; }
}
''')
write(f'{N}/NetworkedParticleEffectColorData.java', r'''
package team.lodestar.lodestone.systems.network.particle;

import com.mojang.serialization.Codec;
import com.mojang.serialization.codecs.RecordCodecBuilder;
import team.lodestar.lodestone.systems.easing.Easing;
import team.lodestar.lodestone.systems.particle.data.color.ColorParticleData;
import java.util.*;

public class NetworkedParticleEffectColorData {
    public static final Codec<ColorParticleData> COLOR_CODEC = RecordCodecBuilder.create(instance -> instance.group(
        Codec.FLOAT.fieldOf("r1").forGetter(d -> d.r1),
        Codec.FLOAT.fieldOf("g1").forGetter(d -> d.g1),
        Codec.FLOAT.fieldOf("b1").forGetter(d -> d.b1),
        Codec.FLOAT.fieldOf("r2").forGetter(d -> d.r2),
        Codec.FLOAT.fieldOf("g2").forGetter(d -> d.g2),
        Codec.FLOAT.fieldOf("b2").forGetter(d -> d.b2),
        Codec.FLOAT.fieldOf("coefficient").forGetter(d -> d.colorCoefficient),
        Codec.STRING.fieldOf("easing").forGetter(d -> d.colorCurveEasing.name)
    ).apply(instance, (r1,g1,b1,r2,g2,b2,coefficient,easing) ->
        ColorParticleData.create(r1,g1,b1,r2,g2,b2)
            .setCoefficient(coefficient).setEasing(Easing.valueOf(easing)).build()));

    public static final Codec<NetworkedParticleEffectColorData> CODEC = RecordCodecBuilder.create(instance -> instance.group(
        COLOR_CODEC.listOf().fieldOf("colors").forGetter(NetworkedParticleEffectColorData::getColors)
    ).apply(instance, NetworkedParticleEffectColorData::new));

    protected final List<ColorParticleData> colors;
    protected int colorCycleCounter;
    public static NetworkedParticleEffectColorData fromColors(List<? extends ColorParticleData> colors) { return new NetworkedParticleEffectColorData(colors); }
    public static NetworkedParticleEffectColorData fromColor(ColorParticleData color) { return new NetworkedParticleEffectColorData(List.of(color)); }
    public NetworkedParticleEffectColorData(List<? extends ColorParticleData> colors) { this.colors = colors.isEmpty() ? Collections.emptyList() : List.copyOf(colors); }
    public NetworkedParticleEffectColorData(ColorParticleData... colors) { this(List.of(colors)); }
    public ColorParticleData getColor() { return colors.size() == 1 ? colors.get(0) : colors.get(colorCycleCounter++ % colors.size()); }
    public List<ColorParticleData> getColors() { return colors; }
}
''')
write(f'{N}/NetworkedParticleEffectPayload.java', r'''
package team.lodestar.lodestone.systems.network.particle;

import javax.annotation.Nullable;

public final class NetworkedParticleEffectPayload {
    private final NetworkedParticleEffectType<?> effect;
    @Nullable private final NetworkedParticleEffectPositionData positionData;
    @Nullable private final NetworkedParticleEffectColorData colorData;
    @Nullable private final NetworkedParticleEffectExtraData extraData;
    public NetworkedParticleEffectPayload(NetworkedParticleEffectType<?> effect, NetworkedParticleEffectPositionData positionData,
            NetworkedParticleEffectColorData colorData, NetworkedParticleEffectExtraData extraData) {
        this.effect=effect; this.positionData=positionData; this.colorData=colorData; this.extraData=extraData;
    }
    public NetworkedParticleEffectType<?> effect() { return effect; }
    public NetworkedParticleEffectPositionData positionData() { return positionData; }
    public NetworkedParticleEffectColorData colorData() { return colorData; }
    public NetworkedParticleEffectExtraData extraData() { return extraData; }
}
''')
write(f'{N}/NetworkedParticleEffectType.java', r'''
package team.lodestar.lodestone.systems.network.particle;

import com.mojang.serialization.Codec;
import com.mojang.serialization.DataResult;
import com.sammy.malum.compat.Forge120ParticleNetwork;
import net.minecraft.core.BlockPos;
import net.minecraft.server.level.ServerLevel;
import net.minecraft.util.RandomSource;
import net.minecraft.world.entity.Entity;
import net.minecraft.world.level.Level;
import net.minecraft.world.phys.Vec3;
import team.lodestar.lodestone.systems.particle.data.color.ColorParticleData;

import java.awt.Color;
import java.util.*;
import java.util.function.Consumer;

public abstract class NetworkedParticleEffectType<T extends NetworkedParticleEffectExtraData> {
    public static final Map<String, NetworkedParticleEffectType<?>> EFFECT_TYPES = new LinkedHashMap<>();
    public static final Codec<NetworkedParticleEffectType<?>> CODEC = Codec.STRING.comapFlatMap(
        id -> EFFECT_TYPES.containsKey(id) ? DataResult.success(EFFECT_TYPES.get(id)) : DataResult.error(() -> "Unknown particle effect " + id),
        NetworkedParticleEffectType::getId);
    protected final String id;
    public NetworkedParticleEffectType(String id) { this.id=id; EFFECT_TYPES.put(id, this); }
    public String getId() { return id; }
    public Optional<Codec<? extends NetworkedParticleEffectPositionData>> getPositionCodec() { return Optional.of(NetworkedParticleEffectPositionData.CODEC); }
    public Optional<Codec<? extends NetworkedParticleEffectColorData>> getColorCodec() { return Optional.of(NetworkedParticleEffectColorData.CODEC); }
    public Optional<Codec<? extends NetworkedParticleEffectExtraData>> getExtraCodec() { return Optional.empty(); }
    public Optional<? extends NetworkedParticleEffectExtraData> getDefaultExtraData() { return Optional.empty(); }
    @SuppressWarnings("unchecked")
    protected void castAndAct(Level level, RandomSource random, NetworkedParticleEffectPositionData positionData,
            NetworkedParticleEffectColorData colorData, NetworkedParticleEffectExtraData extraData) {
        act(level, random, positionData, colorData, (T)extraData);
    }
    public final void dispatch(Level level, NetworkedParticleEffectPositionData positionData,
            NetworkedParticleEffectColorData colorData, NetworkedParticleEffectExtraData extraData) {
        castAndAct(level, level.random, positionData, colorData, extraData);
    }
    public abstract void act(Level level, RandomSource random, NetworkedParticleEffectPositionData positionData,
            NetworkedParticleEffectColorData colorData, T extraData);
    public ParticleEffectBuilder<T> createEffect(BlockPos position) { return createEffect().at(position); }
    public ParticleEffectBuilder<T> createEffect(Vec3 position) { return createEffect().at(position); }
    public ParticleEffectBuilder<T> createEffect(Entity target) { return createEffect().at(target); }
    public ParticleEffectBuilder<T> createEffect() { return new ParticleEffectBuilder<>(this); }

    public static class ParticleEffectBuilder<T extends NetworkedParticleEffectExtraData> {
        protected final NetworkedParticleEffectType<T> type;
        protected NetworkedParticleEffectPositionData position;
        protected NetworkedParticleEffectColorData color;
        protected T extra;
        public ParticleEffectBuilder(NetworkedParticleEffectType<T> type) { this.type=type; }
        public ParticleEffectBuilder<T> at(BlockPos p) { return at(new NetworkedParticleEffectPositionData(p)); }
        public ParticleEffectBuilder<T> at(Vec3 p) { return at(new NetworkedParticleEffectPositionData(p)); }
        public ParticleEffectBuilder<T> at(Entity p) { return at(new NetworkedParticleEffectPositionData(p)); }
        public ParticleEffectBuilder<T> at(NetworkedParticleEffectPositionData p) { position=p; return this; }
        public ParticleEffectBuilder<T> color(Color c) { return color(ColorParticleData.create(c).build()); }
        public ParticleEffectBuilder<T> color(ColorParticleData c) { return color(NetworkedParticleEffectColorData.fromColor(c)); }
        public ParticleEffectBuilder<T> color(List<? extends ColorParticleData> c) { return color(NetworkedParticleEffectColorData.fromColors(c)); }
        public ParticleEffectBuilder<T> color(NetworkedParticleEffectColorData c) { color=c; return this; }
        public ParticleEffectBuilder<T> customData(T e) { extra=e; return this; }
        @SuppressWarnings("unchecked") protected T getCustomData() {
            if (type.getExtraCodec().isEmpty()) return null;
            if (extra == null) {
                var d=type.getDefaultExtraData();
                if (d.isEmpty()) throw new IllegalArgumentException("Particle effect requires custom data: " + type.getId());
                extra=(T)d.get();
            }
            return extra;
        }
        public ParticleEffectBuilder<T> spawn(ServerLevel level) {
            Forge120ParticleNetwork.send(level, new NetworkedParticleEffectPayload(type, position, color, getCustomData()));
            return this;
        }
        public ParticleEffectBuilder<T> spawn(Consumer<NetworkedParticleEffectPayload> sender) {
            sender.accept(new NetworkedParticleEffectPayload(type, position, color, getCustomData())); return this;
        }
    }
}
''')

write('com/sammy/malum/compat/Forge120ParticleNetwork.java', r'''
package com.sammy.malum.compat;

import com.mojang.serialization.Codec;
import com.sammy.malum.MalumMod;
import net.minecraft.client.Minecraft;
import net.minecraft.nbt.*;
import net.minecraft.network.FriendlyByteBuf;
import net.minecraft.server.level.ServerLevel;
import net.minecraftforge.api.distmarker.Dist;
import net.minecraftforge.fml.DistExecutor;
import net.minecraftforge.fml.common.Mod;
import net.minecraftforge.fml.event.lifecycle.FMLCommonSetupEvent;
import net.minecraftforge.eventbus.api.SubscribeEvent;
import net.minecraftforge.network.*;
import net.minecraftforge.network.simple.SimpleChannel;
import team.lodestar.lodestone.systems.network.particle.*;
import java.util.Optional;

@Mod.EventBusSubscriber(modid = MalumMod.MALUM, bus = Mod.EventBusSubscriber.Bus.MOD)
public final class Forge120ParticleNetwork {
    private static final String VERSION = "1";
    private static final SimpleChannel CHANNEL = NetworkRegistry.newSimpleChannel(
        MalumMod.malumPath("particle_bridge"), () -> VERSION, VERSION::equals, VERSION::equals);

    @SubscribeEvent
    public static void setup(FMLCommonSetupEvent event) {
        CHANNEL.registerMessage(0, Packet.class, Packet::encode, Packet::decode, Packet::handle,
            Optional.of(NetworkDirection.PLAY_TO_CLIENT));
    }

    public static void send(ServerLevel level, NetworkedParticleEffectPayload payload) {
        Packet packet = Packet.from(payload);
        if (payload.positionData() == null) {
            CHANNEL.send(PacketDistributor.ALL.noArg(), packet);
        } else {
            CHANNEL.send(PacketDistributor.TRACKING_CHUNK.with(() -> level.getChunkAt(payload.positionData().getAsBlockPos())), packet);
        }
    }

    @SuppressWarnings({"rawtypes","unchecked"})
    private static CompoundTag encodeValue(Codec codec, Object value) {
        if (value == null) return null;
        Tag encoded = (Tag)codec.encodeStart(NbtOps.INSTANCE, value).result().orElseThrow(() -> new IllegalStateException("Could not encode Malum particle payload"));
        CompoundTag wrapper = new CompoundTag();
        wrapper.put("value", encoded);
        return wrapper;
    }

    @SuppressWarnings({"rawtypes","unchecked"})
    private static Object decodeValue(Codec codec, CompoundTag wrapper) {
        if (wrapper == null || !wrapper.contains("value")) return null;
        return codec.parse(NbtOps.INSTANCE, wrapper.get("value")).result().orElseThrow(() -> new IllegalStateException("Could not decode Malum particle payload"));
    }

    public record Packet(String id, CompoundTag position, CompoundTag color, CompoundTag extra) {
        @SuppressWarnings({"rawtypes","unchecked"})
        static Packet from(NetworkedParticleEffectPayload payload) {
            NetworkedParticleEffectType type = payload.effect();
            CompoundTag p = type.getPositionCodec().map(c -> encodeValue((Codec)c, payload.positionData())).orElse(null);
            CompoundTag c = type.getColorCodec().map(x -> encodeValue((Codec)x, payload.colorData())).orElse(null);
            CompoundTag e = type.getExtraCodec().map(x -> encodeValue((Codec)x, payload.extraData())).orElse(null);
            return new Packet(type.getId(), p, c, e);
        }
        static void encode(Packet msg, FriendlyByteBuf buf) {
            buf.writeUtf(msg.id);
            buf.writeBoolean(msg.position != null); if (msg.position != null) buf.writeNbt(msg.position);
            buf.writeBoolean(msg.color != null); if (msg.color != null) buf.writeNbt(msg.color);
            buf.writeBoolean(msg.extra != null); if (msg.extra != null) buf.writeNbt(msg.extra);
        }
        static Packet decode(FriendlyByteBuf buf) {
            String id=buf.readUtf();
            CompoundTag p=buf.readBoolean()?buf.readNbt():null;
            CompoundTag c=buf.readBoolean()?buf.readNbt():null;
            CompoundTag e=buf.readBoolean()?buf.readNbt():null;
            return new Packet(id,p,c,e);
        }
        static void handle(Packet msg, java.util.function.Supplier<NetworkEvent.Context> ctxSupplier) {
            NetworkEvent.Context ctx=ctxSupplier.get();
            ctx.enqueueWork(() -> DistExecutor.unsafeRunWhenOn(Dist.CLIENT, () -> () -> Client.handle(msg)));
            ctx.setPacketHandled(true);
        }
    }

    private static final class Client {
        @SuppressWarnings({"rawtypes","unchecked"})
        static void handle(Packet msg) {
            var level=Minecraft.getInstance().level;
            if (level == null) return;
            NetworkedParticleEffectType type=NetworkedParticleEffectType.EFFECT_TYPES.get(msg.id());
            if (type == null) throw new IllegalStateException("Unknown Malum particle effect: " + msg.id());
            NetworkedParticleEffectPositionData p=(NetworkedParticleEffectPositionData)type.getPositionCodec().map(c -> decodeValue((Codec)c,msg.position())).orElse(null);
            NetworkedParticleEffectColorData c=(NetworkedParticleEffectColorData)type.getColorCodec().map(x -> decodeValue((Codec)x,msg.color())).orElse(null);
            NetworkedParticleEffectExtraData e=(NetworkedParticleEffectExtraData)type.getExtraCodec().map(x -> decodeValue((Codec)x,msg.extra())).orElse(null);
            type.dispatch(level,p,c,e);
        }
    }
}
''')

changed = 0
network_root = root / 'com/sammy/malum/visual_effects/networked'
for p in network_root.rglob('*.java'):
    text=p.read_text(encoding='utf-8')
    old=text
    text=re.sub(r'^import io\.netty\.buffer\.\*?[^;]*;\n', '', text, flags=re.M)
    text=re.sub(r'^import net\.minecraft\.network\.codec\.\*?[^;]*;\n', '', text, flags=re.M)
    text=text.replace('ColorParticleDataWrapper', 'ColorParticleData')
    text=text.replace('ColorParticleData.CODEC', 'NetworkedParticleEffectColorData.COLOR_CODEC')
    text=text.replace('Optional<StreamCodec<ByteBuf, ? extends NetworkedParticleEffectColorData>>', 'Optional<Codec<? extends NetworkedParticleEffectColorData>>')
    text=text.replace('Optional<StreamCodec<ByteBuf, ? extends NetworkedParticleEffectExtraData>>', 'Optional<Codec<? extends NetworkedParticleEffectExtraData>>')
    text=re.sub(r'\b([A-Za-z0-9_$.]+)\.STREAM_CODEC\b', r'\1.CODEC', text)
    text=re.sub(r'\n\s*(?:public|private|protected)\s+static\s+(?:final\s+)?StreamCodec<.*?\bSTREAM_CODEC\s*=.*?;\s*\n', '\n', text, flags=re.S)
    if 'Optional<Codec<' in text and 'com.mojang.serialization' not in text:
        pkg_end=text.find('\n', text.find('package '))+1
        text=text[:pkg_end]+'\nimport com.mojang.serialization.Codec;\n'+text[pkg_end:]
    if text != old:
        p.write_text(text, encoding='utf-8'); changed += 1

left=[]
for p in network_root.rglob('*.java'):
    t=p.read_text(encoding='utf-8')
    if 'net.minecraft.network.codec' in t or 'StreamCodec<' in t or '.STREAM_CODEC' in t:
        left.append(str(p))
if left:
    raise SystemExit('particle StreamCodec surfaces remain:\n'+'\n'.join(left[:100]))
if changed == 0:
    raise SystemExit('particle bridge changed no Malum networked particle sources')
print(f'Forge 1.20 particle network bridge staged; transformed files={changed}')
