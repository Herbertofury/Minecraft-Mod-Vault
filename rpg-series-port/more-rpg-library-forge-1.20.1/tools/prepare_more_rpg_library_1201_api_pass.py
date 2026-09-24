#!/usr/bin/env python3
from __future__ import annotations

import pathlib
import re
import sys

if len(sys.argv) != 2:
    raise SystemExit('usage: prepare_more_rpg_library_1201_api_pass.py <generated-port-root>')
root = pathlib.Path(sys.argv[1]).resolve()
java_root = root / 'common/src/main/java'
generated_root = root / 'common/src/main/generated'
if not java_root.is_dir():
    raise SystemExit(f'missing generated common Java root: {java_root}')


def require_file(rel: str) -> pathlib.Path:
    path = java_root / rel
    if not path.is_file():
        raise SystemExit(f'missing required 2.7.2 source seam: {rel}')
    return path


def require_replace(path: pathlib.Path, old: str, new: str, label: str, expected: int = 1) -> None:
    text = path.read_text(encoding='utf-8')
    actual = text.count(old)
    if actual != expected:
        raise SystemExit(f'{label}: expected {expected} occurrence(s), found {actual}')
    path.write_text(text.replace(old, new), encoding='utf-8')


def remove_balanced_if(path: pathlib.Path, marker: str, label: str) -> None:
    text = path.read_text(encoding='utf-8')
    start = text.find(marker)
    if start < 0:
        raise SystemExit(f'{label}: guarded block marker missing')
    brace = text.find('{', start)
    if brace < 0:
        raise SystemExit(f'{label}: opening brace missing')
    depth = 0
    end = None
    for i in range(brace, len(text)):
        if text[i] == '{':
            depth += 1
        elif text[i] == '}':
            depth -= 1
            if depth == 0:
                end = i + 1
                break
    if end is None:
        raise SystemExit(f'{label}: unterminated guarded block')
    path.write_text(text[:start] + text[end:], encoding='utf-8')


# 1.21 made Identifier factories static; 1.20.1 still uses public constructors. This is a pure
# namespace-era syntax translation and is already proven by the graduated Archers target pass.
identifier_rewrites = 0
simple_particle_rewrites = 0
item_name_rewrites = 0
chest_equipment_rewrites = 0
for path in sorted(java_root.rglob('*.java')):
    text = path.read_text(encoding='utf-8')
    identifier_rewrites += text.count('Identifier.of(') + text.count('Identifier.ofVanilla(')
    simple_particle_rewrites += text.count('SimpleParticleType')
    item_name_rewrites += text.count('.getItemName()')
    chest_equipment_rewrites += text.count('.chestEquipment()')
    updated = text.replace('Identifier.ofVanilla(', 'new Identifier(')
    updated = updated.replace('Identifier.of(', 'new Identifier(')
    updated = updated.replace('SimpleParticleType', 'DefaultParticleType')
    updated = updated.replace('.getItemName()', '.getName()')
    updated = updated.replace('.chestEquipment()', '.getEquippedStack(net.minecraft.entity.EquipmentSlot.CHEST)')
    if updated != text:
        path.write_text(updated, encoding='utf-8')
if identifier_rewrites < 20:
    raise SystemExit(f'Identifier target-era pass unexpectedly small: {identifier_rewrites}')
if simple_particle_rewrites < 4:
    raise SystemExit(f'particle target-era pass unexpectedly small: {simple_particle_rewrites}')

# Preserve the modern popup payload fields and codec while translating the 1.21 packet-codec API back
# onto 1.20.1 ParticleEffect.Factory + PacketByteBuf, which is the same wire/data contract for this type.
popup = require_file('net/more_rpg_classes/client/particle/PopupParticleEffect.java')
popup.write_text(r'''package net.more_rpg_classes.client.particle;

import com.mojang.brigadier.StringReader;
import com.mojang.brigadier.exceptions.CommandSyntaxException;
import com.mojang.serialization.Codec;
import com.mojang.serialization.codecs.RecordCodecBuilder;
import net.minecraft.network.PacketByteBuf;
import net.minecraft.particle.ParticleEffect;
import net.minecraft.particle.ParticleType;
import net.minecraft.registry.Registries;
import net.minecraft.util.Identifier;

import java.util.Locale;

public class PopupParticleEffect implements ParticleEffect {
    private final ParticleType<PopupParticleEffect> type;
    public final Identifier iconId;
    public final boolean isSpell;
    public final int entityId;

    public PopupParticleEffect(ParticleType<PopupParticleEffect> type, Identifier iconId, boolean isSpell, int entityId) {
        this.type = type;
        this.iconId = iconId;
        this.isSpell = isSpell;
        this.entityId = entityId;
    }

    public static final ParticleEffect.Factory<PopupParticleEffect> FACTORY = new ParticleEffect.Factory<>() {
        @Override
        public PopupParticleEffect read(ParticleType<PopupParticleEffect> type, StringReader reader) throws CommandSyntaxException {
            reader.expect(' ');
            Identifier iconId = Identifier.fromCommandInput(reader);
            reader.expect(' ');
            boolean isSpell = reader.readBoolean();
            reader.expect(' ');
            int entityId = reader.readInt();
            return new PopupParticleEffect(type, iconId, isSpell, entityId);
        }

        @Override
        public PopupParticleEffect read(ParticleType<PopupParticleEffect> type, PacketByteBuf buf) {
            return new PopupParticleEffect(type, buf.readIdentifier(), buf.readBoolean(), buf.readVarInt());
        }
    };

    @Override
    public ParticleType<PopupParticleEffect> getType() {
        return type;
    }

    @Override
    public void write(PacketByteBuf buf) {
        buf.writeIdentifier(iconId);
        buf.writeBoolean(isSpell);
        buf.writeVarInt(entityId);
    }

    @Override
    public String asString() {
        return String.format(Locale.ROOT, "%s %s %s %d", Registries.PARTICLE_TYPE.getId(type), iconId, isSpell, entityId);
    }

    public static Codec<PopupParticleEffect> createCodec(ParticleType<PopupParticleEffect> type) {
        return RecordCodecBuilder.create(instance -> instance.group(
                Identifier.CODEC.fieldOf("icon").forGetter(e -> e.iconId),
                Codec.BOOL.fieldOf("is_spell").forGetter(e -> e.isSpell),
                Codec.INT.fieldOf("entity").forGetter(e -> e.entityId)
        ).apply(instance, (iconId, isSpell, entityId) -> new PopupParticleEffect(type, iconId, isSpell, entityId)));
    }
}
''', encoding='utf-8')

more_particles = require_file('net/more_rpg_classes/client/particle/MoreParticles.java')
text = more_particles.read_text(encoding='utf-8')
for imp in (
        'import com.mojang.serialization.MapCodec;\n',
        'import net.minecraft.network.RegistryByteBuf;\n',
        'import net.minecraft.network.codec.PacketCodec;\n'):
    if text.count(imp) != 1:
        raise SystemExit(f'MoreParticles import seam drifted: {imp.strip()} count={text.count(imp)}')
    text = text.replace(imp, '')
ctor = 'new ParticleType<PopupParticleEffect>(false) {'
if text.count(ctor) != 2:
    raise SystemExit(f'MoreParticles custom particle ctor seam drifted: found {text.count(ctor)}')
text = text.replace(ctor, 'new ParticleType<PopupParticleEffect>(false, PopupParticleEffect.FACTORY) {')
text = text.replace('public MapCodec<PopupParticleEffect> getCodec()', 'public com.mojang.serialization.Codec<PopupParticleEffect> getCodec()')
packet_method = re.compile(
    r'\n\s*@Override\s*\n\s*public PacketCodec<\? super RegistryByteBuf, PopupParticleEffect> getPacketCodec\(\) \{\s*\n'
    r'\s*return PopupParticleEffect\.createPacketCodec\(this\);\s*\n\s*\}', re.MULTILINE)
text, packet_methods = packet_method.subn('', text)
if packet_methods != 2:
    raise SystemExit(f'MoreParticles packet-codec seam drifted: expected 2, removed {packet_methods}')
more_particles.write_text(text, encoding='utf-8')

# 1.20.1 has pre-CustomPayload networking. Keep the 2.7.2 packet fields and deterministic byte order;
# the Forge SimpleChannel binding is platform-owned and can consume this loader-neutral read/write type.
mob_beam = require_file('net/more_rpg_classes/network/MobBeamPacket.java')
mob_beam.write_text(r'''package net.more_rpg_classes.network;

import net.minecraft.network.PacketByteBuf;
import net.minecraft.util.Identifier;
import org.jetbrains.annotations.Nullable;

/** S2C packet sent when a mob starts or stops casting a BEAM spell. */
public record MobBeamPacket(int casterId, int targetId, @Nullable Identifier spellId) {
    public static final Identifier ID = new Identifier("more_rpg_classes", "mob_beam");

    public void write(PacketByteBuf buf) {
        buf.writeInt(casterId);
        buf.writeInt(targetId);
        buf.writeBoolean(spellId != null);
        if (spellId != null) buf.writeIdentifier(spellId);
    }

    public static MobBeamPacket read(PacketByteBuf buf) {
        int casterId = buf.readInt();
        int targetId = buf.readInt();
        Identifier spellId = buf.readBoolean() ? buf.readIdentifier() : null;
        return new MobBeamPacket(casterId, targetId, spellId);
    }
}
''', encoding='utf-8')

networking = require_file('net/more_rpg_classes/network/MRPGCNetworking.java')
networking.write_text(r'''package net.more_rpg_classes.network;

import net.minecraft.server.network.ServerPlayerEntity;

public class MRPGCNetworking {
    public interface Sender {
        void sendToPlayer(ServerPlayerEntity player, MobBeamPacket payload);
    }

    private static Sender sender;

    public static void setSender(Sender sender) {
        MRPGCNetworking.sender = sender;
    }

    public static void sendToPlayer(ServerPlayerEntity player, MobBeamPacket payload) {
        if (sender == null) {
            throw new IllegalStateException("More RPG networking sender not installed");
        }
        sender.sendToPlayer(player, payload);
    }
}
''', encoding='utf-8')

# Entity data construction changed in 1.20.5+. This entity tracks no fields, so the exact target-era
# equivalent is the ordinary no-arg hook.
lightning = require_file('net/more_rpg_classes/entity/FriendlyLightningEntity.java')
require_replace(lightning,
                'protected void initDataTracker(DataTracker.Builder builder) {',
                'protected void initDataTracker() {',
                'FriendlyLightningEntity 1.20.1 data tracker')

# Critical Strike is optional and already has a provider-backed compat owner in this port. The modern
# direct API block would make the optional dependency mandatory at compile/runtime, so remove that
# duplicate direct owner and let CriticalStrikeCompat.init() attach the same sources when present.
schools = require_file('net/more_rpg_classes/custom/MoreSpellSchools.java')
require_replace(schools, 'import net.critical_strike.api.CriticalStrikeAttributes;\n', '',
                'MoreSpellSchools direct Critical Strike import')
remove_balanced_if(
    schools,
    'if (net.more_rpg_classes.compat.MoreRpgPlatform.isModLoaded.test("critical_strike")) {',
    'MoreSpellSchools direct Critical Strike block')
if 'CriticalStrikeAttributes' in schools.read_text(encoding='utf-8'):
    raise SystemExit('direct Critical Strike API reference survived MoreSpellSchools decoupling')

# FabricItemGroup is only a convenience builder. Vanilla 1.20.1 exposes the same builder directly;
# preserve both conditional groups and their icon/name behavior without shipping Fabric API.
item_groups = require_file('net/more_rpg_classes/item/MRPGCItemGroups.java')
require_replace(item_groups, 'import net.fabricmc.fabric.api.itemgroup.v1.FabricItemGroup;\n', '',
                'MRPGCItemGroups Fabric builder import')
text = item_groups.read_text(encoding='utf-8')
if text.count('FabricItemGroup.builder()') != 2:
    raise SystemExit(f'MRPGCItemGroups builder seam drifted: found {text.count("FabricItemGroup.builder()")})')
text = text.replace('FabricItemGroup.builder()', 'new ItemGroup.Builder(ItemGroup.Row.BOTTOM, 6)', 1)
text = text.replace('FabricItemGroup.builder()', 'new ItemGroup.Builder(ItemGroup.Row.BOTTOM, 7)', 1)
item_groups.write_text(text, encoding='utf-8')

# Translate the one modern item tooltip callback already exposed by #328 to the native 1.20.1 API.
smithing = require_file('net/more_rpg_classes/compat/armory_rpgs/SmithingIngredients.java')
require_replace(smithing,
                'import net.minecraft.item.tooltip.TooltipType;\n',
                'import net.minecraft.client.item.TooltipContext;\nimport net.minecraft.world.World;\nimport org.jetbrains.annotations.Nullable;\n',
                'SmithingIngredients 1.20.1 tooltip imports')
require_replace(smithing,
                'public void appendTooltip(ItemStack stack, Item.TooltipContext context, List<Text> tooltip, TooltipType type) {\n            super.appendTooltip(stack, context, tooltip, type);',
                'public void appendTooltip(ItemStack stack, @Nullable World world, List<Text> tooltip, TooltipContext context) {\n            super.appendTooltip(stack, world, tooltip, context);',
                'SmithingIngredients 1.20.1 tooltip callback')

# Datagen is a build-time authority, not a production runtime package. Preserve the generated 2.7.2
# resource tree and let Gradle compile only runtime/common Java; the separate Forge datagen port remains
# auditable instead of forcing Fabric datagen classes into the shipping JAR.
generated_json = sorted(generated_root.rglob('*.json')) if generated_root.is_dir() else []
if len(generated_json) < 10:
    raise SystemExit(f'2.7.2 generated resource authority unexpectedly small: {len(generated_json)} JSON files')

# Fail closed on the API families this wave owns.
for forbidden in (
        'Identifier.of(', 'Identifier.ofVanilla(', 'SimpleParticleType', 'RegistryByteBuf',
        'net.minecraft.network.codec.PacketCodec', 'net.minecraft.network.packet.CustomPayload',
        'DataTracker.Builder', 'FabricItemGroup', 'CriticalStrikeAttributes'):
    owners = []
    for path in sorted(java_root.rglob('*.java')):
        if forbidden in path.read_text(encoding='utf-8'):
            owners.append(path.relative_to(java_root).as_posix())
    if owners:
        raise SystemExit(f'owned 1.21/Fabric API survived ({forbidden}): ' + ', '.join(owners[:40]))

print('[More RPG 2.7.2] TARGET_1201_API_WAVE1_PASS '
      f'identifier={identifier_rewrites} simple_particle={simple_particle_rewrites} '
      f'item_name={item_name_rewrites} chest_equipment={chest_equipment_rewrites} '
      f'generated_json={len(generated_json)}')
print('[More RPG 2.7.2] POPUP_PARTICLE_1201_FACTORY_CODEC_PASS')
print('[More RPG 2.7.2] MOB_BEAM_PRE_CUSTOM_PAYLOAD_CONTRACT_PASS')
print('[More RPG 2.7.2] OPTIONAL_CRITICAL_STRIKE_OWNER_DECOUPLED')
print('[More RPG 2.7.2] PRODUCTION_DATAGEN_SEPARATION_READY')
