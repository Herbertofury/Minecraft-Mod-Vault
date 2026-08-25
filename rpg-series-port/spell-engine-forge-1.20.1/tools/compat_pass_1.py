#!/usr/bin/env python3
from pathlib import Path
import re, sys

root = Path(sys.argv[1]).resolve()
java = root / 'common/src/main/java'

# Use exact absolute paths supplied by the CI runner for the two already-verified API jars.
# Avoid Loom/fileTree ambiguity while keeping them compile-only (they are separate mods at runtime).
build = root / 'common/build.gradle'
bs = build.read_text()
bs = re.sub(r"\s*compileOnly fileTree\(dir: rootProject\.file\('../rpg-series-port/spell_power-forge-1\.20\.1/common/build/libs'\), include: \['\*\.jar'\], exclude: \['\*sources\*'\]\)",
            "\n compileOnly files(System.getenv('SPELL_POWER_COMMON_JAR'))", bs)
bs = re.sub(r"\s*compileOnly fileTree\(dir: rootProject\.file\('../rpg-series-port/ranged-weapon-api-forge-1\.20\.1/common/build/libs'\), include: \['\*\.jar'\], exclude: \['\*sources\*'\]\)",
            "\n compileOnly files(System.getenv('RANGED_COMMON_JAR'))", bs)
build.write_text(bs)

# 1.21's Identifier.of(...) is constructor syntax on 1.20.1. Both the one- and two-argument
# overloads map directly to the corresponding 1.20.1 constructors.
for f in java.rglob('*.java'):
    s = f.read_text()
    s = s.replace('Identifier.of(', 'new Identifier(')
    f.write_text(s)

# 1.20.1 networking uses PacketByteBuf + loader channels, not 1.20.5+'s CustomPayload/PacketCodec.
# Keep the exact record payloads and their read/write wire format; only remove the newer vanilla
# envelope/codec types. Forge transport is added in the platform stage.
p = java / 'net/spell_engine/network/Packets.java'
s = p.read_text()
s = s.replace('import net.minecraft.network.RegistryByteBuf;\n', '')
s = s.replace('import net.minecraft.network.codec.PacketCodec;\n', '')
s = s.replace('import net.minecraft.network.packet.CustomPayload;\n', '')
s = s.replace('RegistryByteBuf', 'PacketByteBuf')
s = s.replace(' implements CustomPayload', '')
s = re.sub(r'^\s*public static final CustomPayload\.Id<[^\n]+\n', '', s, flags=re.M)
s = re.sub(r'^\s*public static final PacketCodec<[^\n]+\n', '', s, flags=re.M)
s = re.sub(
    r'\n\s*@Override\s*\n\s*public (?:CustomPayload\.)?Id<\? extends CustomPayload> getId\(\) \{\s*\n\s*return PACKET_ID;\s*\n\s*\}\s*',
    '\n', s, flags=re.M
)
p.write_text(s)

# Loader-neutral common transport carries the packet records as Objects on 1.20.1. The Forge
# implementation pattern-matches the concrete Packets.* type and serializes via its unchanged
# write/read methods.
p = java / 'net/spell_engine/Platform.java'
s = p.read_text()
s = s.replace('import net.minecraft.network.packet.CustomPayload;\n', '')
s = s.replace('CustomPayload payload', 'Object payload')
p.write_text(s)

# Backport the 1.10 particle type to the 1.20.1 ParticleType/ParticleEffect contract. Spawn-specific
# appearance still rides Spell Engine's own packet; vanilla particle serialization remains payloadless.
p = java / 'net/spell_engine/fx/ParticleGroupType.java'
p.write_text(r'''package net.spell_engine.fx;

import com.mojang.brigadier.StringReader;
import com.mojang.brigadier.exceptions.CommandSyntaxException;
import com.mojang.serialization.Codec;
import net.minecraft.entity.Entity;
import net.minecraft.network.PacketByteBuf;
import net.minecraft.particle.ParticleEffect;
import net.minecraft.particle.ParticleType;
import net.minecraft.registry.Registries;
import net.spell_engine.api.spell.fx.ParticleGroup;
import org.jetbrains.annotations.Nullable;

public class ParticleGroupType extends ParticleType<ParticleGroupType> implements ParticleEffect {
    private static final ParticleEffect.Factory<ParticleGroupType> FACTORY = new ParticleEffect.Factory<>() {
        @Override
        public ParticleGroupType read(ParticleType<ParticleGroupType> type, StringReader reader) throws CommandSyntaxException {
            return ((ParticleGroupType) type).getType();
        }

        @Override
        public ParticleGroupType read(ParticleType<ParticleGroupType> type, PacketByteBuf buffer) {
            return ((ParticleGroupType) type).getType();
        }
    };

    private final Codec<ParticleGroupType> codec = Codec.unit(this::getType);
    private ParticleGroupType root = this;
    private SpellEngineParticles.Entry entry;
    @Nullable private ParticleGroup.Appearance payload;
    @Nullable private Entity sourceEntity;

    public ParticleGroupType() {
        super(true, FACTORY);
    }

    public SpellEngineParticles.Entry entry() { return entry; }
    @Nullable public ParticleGroup.Appearance payload() { return payload; }
    @Nullable public Entity sourceEntity() { return sourceEntity; }

    public ParticleGroupType spawnable(@Nullable ParticleGroup.Appearance payload, @Nullable Entity sourceEntity) {
        var copy = new ParticleGroupType();
        copy.root = this.root;
        copy.entry = this.entry;
        copy.payload = payload;
        copy.sourceEntity = sourceEntity;
        return copy;
    }

    void bind(SpellEngineParticles.Entry entry) { this.entry = entry; }

    @Override
    public ParticleGroupType getType() { return root; }

    @Override
    public Codec<ParticleGroupType> getCodec() { return codec; }

    @Override
    public void write(PacketByteBuf buffer) { }

    @Override
    public String asString() {
        var id = Registries.PARTICLE_TYPE.getId(getType());
        return id != null ? id.toString() : "spell_engine:unregistered";
    }
}
''')

# Guard the pass: if these survive, the mechanical migration silently missed something.
for needle in ('Identifier.of(', 'RegistryByteBuf', 'network.codec.PacketCodec', 'network.packet.CustomPayload'):
    found=[]
    for f in java.rglob('*.java'):
        if needle in f.read_text(): found.append(str(f.relative_to(java)))
    if found:
        raise SystemExit(f'compat pass failed; {needle!r} remains in {found[:12]}')
print('Spell Engine compatibility pass 1 applied: identifiers + 1.20.1 packet envelope + particle type')
