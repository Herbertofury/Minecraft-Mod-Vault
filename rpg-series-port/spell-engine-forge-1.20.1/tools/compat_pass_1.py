#!/usr/bin/env python3
from pathlib import Path
import re, sys

root = Path(sys.argv[1]).resolve()
java = root / 'common/src/main/java'

# Compile Spell Engine against the already-verified dependency common JARs, never their source trees.
# This keeps Spell Power and Ranged Weapon API as real separate mods and prevents their classes from
# being shadowed into the Spell Engine release artifact.
build = root / 'common/build.gradle'
bs = build.read_text()
bs = re.sub(r"\s*compileOnly fileTree\(dir: rootProject\.file\('../rpg-series-port/spell_power-forge-1\.20\.1/common/build/libs'\), include: \['\*\.jar'\], exclude: \['\*sources\*'\]\)", '', bs)
bs = re.sub(r"\s*compileOnly fileTree\(dir: rootProject\.file\('../rpg-series-port/ranged-weapon-api-forge-1\.20\.1/common/build/libs'\), include: \['\*\.jar'\], exclude: \['\*sources\*'\]\)", '', bs)
bs = re.sub(r"\n\ndef addExternalCompileSources = \{ String envName ->.*?addExternalCompileSources\('RANGED_SOURCE_DIRS'\)\n", '\n', bs, flags=re.S)
bs += r'''

def requireExternalJar = { String envName ->
 def raw = System.getenv(envName)
 if (raw == null || raw.isBlank()) {
  throw new GradleException("Missing required external dependency JAR environment variable: ${envName}")
 }
 def jarFile = file(raw)
 if (!jarFile.isFile()) {
  throw new GradleException("External dependency JAR does not exist for ${envName}: ${jarFile}")
 }
 return jarFile
}

dependencies {
 compileOnly files(requireExternalJar('SPELL_POWER_COMMON_JAR'))
 compileOnly files(requireExternalJar('RANGED_COMMON_JAR'))
}
'''
build.write_text(bs)

for f in java.rglob('*.java'):
    s = f.read_text()
    s = s.replace('Identifier.of(', 'new Identifier(')
    f.write_text(s)

p = java / 'net/spell_engine/network/Packets.java'
s = p.read_text()
s = s.replace('import net.minecraft.network.RegistryByteBuf;\n', '')
s = s.replace('import net.minecraft.network.codec.PacketCodec;\n', '')
s = s.replace('import net.minecraft.network.packet.CustomPayload;\n', '')
s = s.replace('RegistryByteBuf', 'PacketByteBuf')
s = s.replace(' implements CustomPayload', '')
s = re.sub(r'^\s*public static final CustomPayload\.Id<[^\n]+\n', '', s, flags=re.M)
s = re.sub(r'^\s*public static final PacketCodec<[^\n]+\n', '', s, flags=re.M)
s = re.sub(r'\n\s*@Override\s*\n\s*public (?:CustomPayload\.)?Id<\? extends CustomPayload> getId\(\) \{\s*\n\s*return PACKET_ID;\s*\n\s*\}\s*', '\n', s, flags=re.M)
p.write_text(s)

p = java / 'net/spell_engine/Platform.java'
s = p.read_text().replace('import net.minecraft.network.packet.CustomPayload;\n', '').replace('CustomPayload payload', 'Object payload')
p.write_text(s)

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
        @Override public ParticleGroupType read(ParticleType<ParticleGroupType> type, StringReader reader) throws CommandSyntaxException { return ((ParticleGroupType) type).getType(); }
        @Override public ParticleGroupType read(ParticleType<ParticleGroupType> type, PacketByteBuf buffer) { return ((ParticleGroupType) type).getType(); }
    };
    private final Codec<ParticleGroupType> codec = Codec.unit(this::getType);
    private ParticleGroupType root = this;
    private SpellEngineParticles.Entry entry;
    @Nullable private ParticleGroup.Appearance payload;
    @Nullable private Entity sourceEntity;
    public ParticleGroupType() { super(true, FACTORY); }
    public SpellEngineParticles.Entry entry() { return entry; }
    @Nullable public ParticleGroup.Appearance payload() { return payload; }
    @Nullable public Entity sourceEntity() { return sourceEntity; }
    public ParticleGroupType spawnable(@Nullable ParticleGroup.Appearance payload, @Nullable Entity sourceEntity) {
        var copy = new ParticleGroupType(); copy.root=this.root; copy.entry=this.entry; copy.payload=payload; copy.sourceEntity=sourceEntity; return copy;
    }
    void bind(SpellEngineParticles.Entry entry) { this.entry=entry; }
    @Override public ParticleGroupType getType() { return root; }
    @Override public Codec<ParticleGroupType> getCodec() { return codec; }
    @Override public void write(PacketByteBuf buffer) { }
    @Override public String asString() { var id=Registries.PARTICLE_TYPE.getId(getType()); return id != null ? id.toString() : "spell_engine:unregistered"; }
}
''')

for needle in ('Identifier.of(', 'RegistryByteBuf', 'network.codec.PacketCodec', 'network.packet.CustomPayload'):
    found=[str(f.relative_to(java)) for f in java.rglob('*.java') if needle in f.read_text()]
    if found: raise SystemExit(f'compat pass failed; {needle!r} remains in {found[:12]}')
for stale in ('SPELL_POWER_SOURCE_DIRS', 'RANGED_SOURCE_DIRS', 'addExternalCompileSources'):
    if stale in build.read_text(): raise SystemExit(f'compat pass left dependency source injection enabled: {stale}')
for required in ('SPELL_POWER_COMMON_JAR', 'RANGED_COMMON_JAR'):
    if required not in build.read_text(): raise SystemExit(f'compat pass missing dependency common JAR gate: {required}')
print('Spell Engine compatibility pass 1 applied: external dependency JAR compile gate + identifiers + 1.20.1 packet envelope + particle type')
