#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 3:
    raise SystemExit('usage: compat_pass_6e.py <generated-port-root> <spell-engine-1.20.1-baseline>')
root = Path(sys.argv[1]).resolve()
common_java = root / 'common/src/main/java'
forge_java = root / 'forge/src/main/java/net/spell_engine/forge'

packets = common_java / 'net/spell_engine/network/Packets.java'
s = packets.read_text()
# Forge SimpleChannel needs callable codecs. 1.10.2 only kept these two helpers private because
# NeoForge PacketCodec can access them from inside the record; exposing them changes no wire format.
s = s.replace('        private void write(PacketByteBuf buffer) {', '        public void write(PacketByteBuf buffer) {')
s = s.replace('        private static SpellRegistrySync read(PacketByteBuf buffer) {', '        public static SpellRegistrySync read(PacketByteBuf buffer) {')
packets.write_text(s)

forge_java.joinpath('ForgeNetwork.java').write_text(r'''package net.spell_engine.forge;

import net.minecraft.server.network.ServerPlayerEntity;
import net.minecraft.util.Identifier;
import net.minecraftforge.api.distmarker.Dist;
import net.minecraftforge.fml.DistExecutor;
import net.minecraftforge.network.NetworkDirection;
import net.minecraftforge.network.NetworkEvent;
import net.minecraftforge.network.NetworkRegistry;
import net.minecraftforge.network.PacketDistributor;
import net.minecraftforge.network.simple.SimpleChannel;
import net.spell_engine.SpellEngineMod;
import net.spell_engine.network.Packets;
import net.spell_engine.network.ServerNetwork;

import java.util.Optional;
import java.util.function.BiConsumer;
import java.util.function.Function;

/** Native Forge 47 play-channel transport for the 1.10.2 packet model. */
final class ForgeNetwork {
    private static final String PROTOCOL = "1";
    private static final SimpleChannel CHANNEL = NetworkRegistry.newSimpleChannel(
            new Identifier(SpellEngineMod.ID, "main"),
            () -> PROTOCOL,
            PROTOCOL::equals,
            PROTOCOL::equals);
    private static int discriminator;

    private ForgeNetwork() { }

    static void init() {
        registerC2S(Packets.CastRequest.class, Packets.CastRequest::write, Packets.CastRequest::read,
                (packet, player) -> ServerNetwork.handleCastRequest(packet, player.server, player));
        registerC2S(Packets.TargetStream.class, Packets.TargetStream::write, Packets.TargetStream::read,
                (packet, player) -> ServerNetwork.handleTargetStream(packet, player.server, player));
        registerC2S(Packets.CastInput.class, Packets.CastInput::write, Packets.CastInput::read,
                (packet, player) -> ServerNetwork.handleCastInput(packet, player.server, player));
        registerC2S(Packets.AttackPerform.class, Packets.AttackPerform::write, Packets.AttackPerform::read,
                (packet, player) -> ServerNetwork.handleAttackPerform(packet, player.server, player));
        registerC2S(Packets.AttackFxBroadcast.class, Packets.AttackFxBroadcast::write, Packets.AttackFxBroadcast::read,
                (packet, player) -> ServerNetwork.handleAttackFxBroadcast(packet, player.server, player));

        // Forge 1.20.1 has no NeoForge-style custom configuration-task API. Config/registry sync
        // therefore travel over this version-locked play channel immediately after login.
        registerS2C(Packets.ConfigSync.class, Packets.ConfigSync::write, Packets.ConfigSync::read);
        registerS2C(Packets.SpellRegistrySync.class, Packets.SpellRegistrySync::write, Packets.SpellRegistrySync::read);
        registerS2C(Packets.SpellCooldown.class, Packets.SpellCooldown::write, Packets.SpellCooldown::read);
        registerS2C(Packets.SpellCooldownSync.class, Packets.SpellCooldownSync::write, Packets.SpellCooldownSync::read);
        registerS2C(Packets.SpellMessage.class, Packets.SpellMessage::write, Packets.SpellMessage::read);
        registerS2C(Packets.ParticleEffects.class, Packets.ParticleEffects::write, Packets.ParticleEffects::read);
        registerS2C(Packets.SpellAnimation.class, Packets.SpellAnimation::write, Packets.SpellAnimation::read);
        registerS2C(Packets.SpellContainerSync.class, Packets.SpellContainerSync::write, Packets.SpellContainerSync::read);
        registerS2C(Packets.AttackAvailable.class, Packets.AttackAvailable::write, Packets.AttackAvailable::read);
    }

    private static <T> void registerC2S(Class<T> type,
                                        BiConsumer<T, net.minecraft.network.PacketByteBuf> encoder,
                                        Function<net.minecraft.network.PacketByteBuf, T> decoder,
                                        BiConsumer<T, ServerPlayerEntity> handler) {
        CHANNEL.registerMessage(discriminator++, type, encoder, decoder, (packet, contextSupplier) -> {
            var context = contextSupplier.get();
            var player = context.getSender();
            if (player != null) {
                context.enqueueWork(() -> handler.accept(packet, player));
            }
            context.setPacketHandled(true);
        }, Optional.of(NetworkDirection.PLAY_TO_SERVER));
    }

    private static <T> void registerS2C(Class<T> type,
                                        BiConsumer<T, net.minecraft.network.PacketByteBuf> encoder,
                                        Function<net.minecraft.network.PacketByteBuf, T> decoder) {
        CHANNEL.registerMessage(discriminator++, type, encoder, decoder, (packet, contextSupplier) -> {
            NetworkEvent.Context context = contextSupplier.get();
            context.enqueueWork(() -> DistExecutor.unsafeRunWhenOn(Dist.CLIENT,
                    () -> () -> ForgeClientNetwork.handle(packet)));
            context.setPacketHandled(true);
        }, Optional.of(NetworkDirection.PLAY_TO_CLIENT));
    }

    static boolean isReady(ServerPlayerEntity player) {
        // ServerPlayNetworkHandler.connection is the target-native 1.20.1 Yarn mapping of Forge's
        // ServerGamePacketListenerImpl.connection field expected by SimpleChannel#isRemotePresent.
        return CHANNEL.isRemotePresent(player.networkHandler.connection);
    }

    static void sendToPlayer(ServerPlayerEntity player, Object payload) {
        CHANNEL.send(PacketDistributor.PLAYER.with(() -> player), payload);
    }

    static void sendToServer(Object payload) {
        CHANNEL.sendToServer(payload);
    }
}
''')

forge_java.joinpath('ForgeClientNetwork.java').write_text(r'''package net.spell_engine.forge;

import net.spell_engine.client.ClientNetwork;
import net.spell_engine.network.Packets;

/** Physical-client-only dispatch target kept out of dedicated-server handler code. */
final class ForgeClientNetwork {
    private ForgeClientNetwork() { }

    static void handle(Object payload) {
        if (payload instanceof Packets.ConfigSync packet) { ClientNetwork.handleConfigSync(packet); return; }
        if (payload instanceof Packets.SpellRegistrySync packet) { ClientNetwork.handleSpellRegistrySync(packet); return; }
        if (payload instanceof Packets.SpellCooldown packet) { ClientNetwork.handleSpellCooldown(packet); return; }
        if (payload instanceof Packets.SpellCooldownSync packet) { ClientNetwork.handleSpellCooldownSync(packet); return; }
        if (payload instanceof Packets.SpellMessage packet) { ClientNetwork.handleSpellMessage(packet); return; }
        if (payload instanceof Packets.ParticleEffects packet) { ClientNetwork.handleParticleEffects(packet); return; }
        if (payload instanceof Packets.SpellAnimation packet) { ClientNetwork.handleSpellAnimation(packet); return; }
        if (payload instanceof Packets.SpellContainerSync packet) { ClientNetwork.handleSpellContainerSync(packet); return; }
        if (payload instanceof Packets.AttackAvailable packet) { ClientNetwork.handleAttackAvailable(packet); return; }
        throw new IllegalArgumentException("Unknown Spell Engine S2C payload: " + payload.getClass().getName());
    }
}
''')

network = forge_java.joinpath('ForgeNetwork.java').read_text()
for forbidden in ('transport has not been initialized', 'static void init() { }', '.getConnection()'):
    if forbidden in network:
        raise SystemExit(f'Forge network scaffold/mapping survived pass 6e: {forbidden}')
for packet in ('CastRequest', 'TargetStream', 'CastInput', 'AttackPerform', 'AttackFxBroadcast',
               'ConfigSync', 'SpellRegistrySync', 'SpellCooldown', 'SpellCooldownSync', 'SpellMessage',
               'ParticleEffects', 'SpellAnimation', 'SpellContainerSync', 'AttackAvailable'):
    if f'Packets.{packet}.class' not in network:
        raise SystemExit(f'missing Forge packet registration: {packet}')
print('Spell Engine compatibility pass 6e applied: native Forge 47 SimpleChannel transport + handlers')
