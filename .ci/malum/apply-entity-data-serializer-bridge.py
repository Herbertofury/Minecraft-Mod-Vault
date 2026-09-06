#!/usr/bin/env python3
import re
import sys
from pathlib import Path

if len(sys.argv) != 2:
    raise SystemExit('usage: apply-entity-data-serializer-bridge.py <java-root>')
root = Path(sys.argv[1]).resolve()
if not root.is_dir():
    raise SystemExit(f'missing Java root: {root}')

def write(rel, text):
    p = root / rel
    p.parent.mkdir(parents=True, exist_ok=True)
    p.write_text(text.strip() + '\n', encoding='utf-8')

# Forge 1.20 still has a first-class custom EntityDataSerializer registry. Keep 1.8.2's
# typed SpiritArcanaType accessors and only replace NeoForge 1.21's StreamCodec factory.
write('com/sammy/malum/registry/common/MalumEntityDataSerializers.java', r'''
package com.sammy.malum.registry.common;

import com.sammy.malum.MalumMod;
import com.sammy.malum.core.systems.spirit.type.SpiritArcanaType;
import com.sammy.malum.registry.common.magic.MalumSpiritTypes;
import net.minecraft.network.FriendlyByteBuf;
import net.minecraft.network.syncher.EntityDataSerializer;
import net.minecraftforge.registries.DeferredRegister;
import net.minecraftforge.registries.ForgeRegistries;
import net.minecraftforge.registries.RegistryObject;

public final class MalumEntityDataSerializers {
    public static final DeferredRegister<EntityDataSerializer<?>> ENTITY_DATA_SERIALIZERS =
            DeferredRegister.create(ForgeRegistries.Keys.ENTITY_DATA_SERIALIZERS, MalumMod.MALUM);

    public static final RegistryObject<EntityDataSerializer<SpiritArcanaType>> SPIRIT_ARCANA =
            ENTITY_DATA_SERIALIZERS.register("spirit_arcana", () -> new EntityDataSerializer<>() {
                @Override
                public void write(FriendlyByteBuf buffer, SpiritArcanaType value) {
                    buffer.writeResourceLocation(value.getRegistryName());
                }

                @Override
                public SpiritArcanaType read(FriendlyByteBuf buffer) {
                    var id = buffer.readResourceLocation();
                    var value = MalumSpiritTypes.SPIRIT_TYPES_REGISTRY.getValue(id);
                    return value == null ? MalumSpiritTypes.ARCANE_SPIRIT.get() : value;
                }

                @Override
                public SpiritArcanaType copy(SpiritArcanaType value) {
                    return value;
                }
            });

    private MalumEntityDataSerializers() {}
}
''')

arcana = root / 'com/sammy/malum/core/systems/spirit/type/SpiritArcanaType.java'
text = arcana.read_text(encoding='utf-8')
text = text.replace('import io.netty.buffer.*;\n', '')
text = text.replace('import net.minecraft.network.codec.*;\n', '')
text, n = re.subn(
    r'\n\s*public static StreamCodec<ByteBuf, SpiritArcanaType> STREAM_CODEC = ByteBufCodecs\.fromCodec\(SpiritArcanaType\.CODEC\);\n',
    '\n', text, count=1)
if n != 1:
    raise SystemExit('SpiritArcanaType StreamCodec signature changed; refusing ambiguous rewrite')
arcana.write_text(text, encoding='utf-8')

print('Forge 1.20 typed spirit entity-data serializer owner staged')
