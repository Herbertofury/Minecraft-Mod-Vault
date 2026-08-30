#!/usr/bin/env python3
from __future__ import annotations

from pathlib import Path
import re
import sys

if len(sys.argv) != 2:
    raise SystemExit('usage: prepare_more_rpg_library_1201_api_wave2.py <generated-port-root>')
root = Path(sys.argv[1]).resolve()
java = root / 'common/src/main/java'
if not java.is_dir():
    raise SystemExit(f'missing common Java root: {java}')


def p(rel: str) -> Path:
    q = java / rel
    if not q.is_file():
        raise SystemExit(f'missing wave2 seam: {rel}')
    return q


def edit(rel: str, fn, label: str) -> None:
    q = p(rel)
    old = q.read_text(encoding='utf-8')
    new = fn(old)
    if new == old:
        raise SystemExit(f'{label}: transform did not match')
    q.write_text(new, encoding='utf-8')


def transform_named_method(text: str, signature_rx: re.Pattern[str], transform, label: str) -> tuple[str, int]:
    out = []
    cursor = 0
    count = 0
    while True:
        m = signature_rx.search(text, cursor)
        if not m:
            out.append(text[cursor:])
            break
        brace = text.find('{', m.end() - 1)
        if brace < 0:
            raise SystemExit(f'{label}: opening brace missing')
        depth = 0
        end = None
        for i in range(brace, len(text)):
            if text[i] == '{': depth += 1
            elif text[i] == '}':
                depth -= 1
                if depth == 0:
                    end = i + 1
                    break
        if end is None:
            raise SystemExit(f'{label}: unterminated method')
        out.append(text[cursor:m.start()])
        out.append(transform(text[m.start():end], m))
        cursor = end
        count += 1
    return ''.join(out), count


# Wave 1 deliberately used a simple Identifier factory rewrite. Fully-qualified factory calls need a
# distinct constructor spelling; repair that exact generated seam and fail if it can recur.
fq_fixed = 0
for q in java.rglob('*.java'):
    s = q.read_text(encoding='utf-8')
    fq_fixed += s.count('net.minecraft.util.new Identifier(')
    s = s.replace('net.minecraft.util.new Identifier(', 'new net.minecraft.util.Identifier(')
    q.write_text(s, encoding='utf-8')
if fq_fixed < 1:
    raise SystemExit('wave2 expected the #329 fully-qualified Identifier regression but found none')

# Genuine 2.7.2 ConditionalJigsaw backport. 1.20.1 StructurePoolBasedGenerator ends at
# maxDistanceFromCenter; alias lookup, dimension padding and liquid settings did not exist yet. Keep
# mod-presence gating, start pool/jigsaw, height, size, projection and distance semantics intact.
conditional = p('net/more_rpg_classes/worldgen/structure/ConditionalJigsawStructure.java')
conditional.write_text(r'''package net.more_rpg_classes.worldgen.structure;

import com.mojang.serialization.Codec;
import com.mojang.serialization.codecs.RecordCodecBuilder;
import net.minecraft.registry.entry.RegistryEntry;
import net.minecraft.structure.pool.StructurePool;
import net.minecraft.structure.pool.StructurePoolBasedGenerator;
import net.minecraft.util.Identifier;
import net.minecraft.util.math.BlockPos;
import net.minecraft.util.math.ChunkPos;
import net.minecraft.world.Heightmap;
import net.minecraft.world.gen.HeightContext;
import net.minecraft.world.gen.heightprovider.HeightProvider;
import net.minecraft.world.gen.structure.Structure;
import net.minecraft.world.gen.structure.StructureType;
import net.more_rpg_classes.compat.MoreRpgPlatform;
import net.more_rpg_classes.worldgen.ModStructureTypes;

import java.util.Optional;

public class ConditionalJigsawStructure extends Structure {
    public static final int MAX_SIZE = 128;

    public static final Codec<ConditionalJigsawStructure> CODEC = RecordCodecBuilder.create(instance ->
            instance.group(
                    configCodecBuilder(instance),
                    Codec.STRING.fieldOf("mod_id").forGetter(s -> s.modId),
                    StructurePool.REGISTRY_CODEC.fieldOf("start_pool").forGetter(s -> s.startPool),
                    Identifier.CODEC.optionalFieldOf("start_jigsaw_name").forGetter(s -> s.startJigsawName),
                    Codec.intRange(0, MAX_SIZE).fieldOf("size").forGetter(s -> s.size),
                    HeightProvider.CODEC.fieldOf("start_height").forGetter(s -> s.startHeight),
                    Codec.BOOL.optionalFieldOf("use_expansion_hack", false).forGetter(s -> s.useExpansionHack),
                    Heightmap.Type.CODEC.optionalFieldOf("project_start_to_heightmap").forGetter(s -> s.projectStartToHeightmap),
                    Codec.intRange(1, MAX_SIZE).fieldOf("max_distance_from_center").forGetter(s -> s.maxDistanceFromCenter)
            ).apply(instance, ConditionalJigsawStructure::new)
    );

    private final String modId;
    private final RegistryEntry<StructurePool> startPool;
    private final Optional<Identifier> startJigsawName;
    private final int size;
    private final HeightProvider startHeight;
    private final boolean useExpansionHack;
    private final Optional<Heightmap.Type> projectStartToHeightmap;
    private final int maxDistanceFromCenter;

    public ConditionalJigsawStructure(Config config, String modId, RegistryEntry<StructurePool> startPool,
            Optional<Identifier> startJigsawName, int size, HeightProvider startHeight,
            boolean useExpansionHack, Optional<Heightmap.Type> projectStartToHeightmap,
            int maxDistanceFromCenter) {
        super(config);
        this.modId = modId;
        this.startPool = startPool;
        this.startJigsawName = startJigsawName;
        this.size = size;
        this.startHeight = startHeight;
        this.useExpansionHack = useExpansionHack;
        this.projectStartToHeightmap = projectStartToHeightmap;
        this.maxDistanceFromCenter = maxDistanceFromCenter;
    }

    @Override
    public Optional<StructurePosition> getStructurePosition(Context context) {
        if (!MoreRpgPlatform.isModLoaded.test(modId)) return Optional.empty();
        ChunkPos chunkPos = context.chunkPos();
        HeightContext heightContext = new HeightContext(context.chunkGenerator(), context.world());
        int y = this.startHeight.get(context.random(), heightContext);
        BlockPos blockPos = new BlockPos(chunkPos.getStartX(), y, chunkPos.getStartZ());
        return StructurePoolBasedGenerator.generate(context, this.startPool, this.startJigsawName, this.size,
                blockPos, this.useExpansionHack, this.projectStartToHeightmap, this.maxDistanceFromCenter);
    }

    @Override public StructureType<?> getType() { return ModStructureTypes.CONDITIONAL_JIGSAW; }
    public String getModId() { return modId; }
}
''', encoding='utf-8')

# 1.20.1 SpellSchool stores raw EntityAttribute. Keep the modern ranged/melee schools and sources while
# selecting the certified Ranged Weapon API raw attributes instead of 1.21 RegistryEntry wrappers.
def school_patch(s: str) -> str:
    s = s.replace('import net.minecraft.registry.entry.RegistryEntry;\n', '')
    s = s.replace('private static RegistryEntry<EntityAttribute> rangedDamageAttribute()',
                  'private static EntityAttribute rangedDamageAttribute()')
    s = s.replace('EntityAttributes_RangedWeapon.DAMAGE.entry', 'EntityAttributes_RangedWeapon.DAMAGE.attribute')
    s = s.replace('EntityAttributes_RangedWeapon.HASTE.entry', 'EntityAttributes_RangedWeapon.HASTE.attribute')
    s = s.replace('SpellSchools.FROST.attributeEntry', 'SpellSchools.FROST.attribute')
    s = s.replace('SpellSchools.FIRE.attributeEntry', 'SpellSchools.FIRE.attribute')
    return s
edit('net/more_rpg_classes/custom/MoreSpellSchools.java', school_patch, 'MoreSpellSchools raw attributes')

# 1.20.1 status effects consume raw StatusEffect values and old callback shapes. Effects.Entry already
# keeps the raw effect as .effect, so retain all 2.7.2 behavior and translate only API representation.
effect_files = sorted((java / 'net/more_rpg_classes/effect').glob('*.java'))
raw_entry_rewrites = 0
update_methods = 0
on_applied_methods = 0
for q in effect_files:
    s = q.read_text(encoding='utf-8')
    raw_entry_rewrites += len(re.findall(r'MRPGCEffects\.([A-Z0-9_]+)\.entry', s))
    s = re.sub(r'MRPGCEffects\.([A-Z0-9_]+)\.entry', r'MRPGCEffects.\1.effect', s)
    s = s.replace('.getEffectType().value()', '.getEffectType()')

    rx = re.compile(r'public\s+boolean\s+applyUpdateEffect\s*\(\s*LivingEntity\s+(\w+)\s*,\s*int\s+(\w+)\s*\)\s*\{')
    def update_transform(method: str, m: re.Match[str]) -> str:
        method = re.sub(r'public\s+boolean\s+applyUpdateEffect', 'public void applyUpdateEffect', method, count=1)
        # Boolean removal return only signaled 1.21 continuation; execute the removal and return from void.
        method = re.sub(r'return\s+([^;]*\.removeStatusEffect\([^;]+\));', r'\1;\n            return;', method)
        method = re.sub(r'\breturn\s+(?:true|false)\s*;', 'return;', method)
        return method
    s, n = transform_named_method(s, rx, update_transform, f'{q.name} applyUpdateEffect')
    update_methods += n

    rx_applied = re.compile(r'public\s+void\s+onApplied\s*\(\s*LivingEntity\s+(\w+)\s*,\s*int\s+(\w+)\s*\)\s*\{')
    def applied_transform(method: str, m: re.Match[str]) -> str:
        ent, amp = m.group(1), m.group(2)
        method = re.sub(r'public\s+void\s+onApplied\s*\(\s*LivingEntity\s+\w+\s*,\s*int\s+\w+\s*\)',
                        f'public void onApplied(LivingEntity {ent}, net.minecraft.entity.attribute.AttributeContainer attributes, int {amp})',
                        method, count=1)
        method = method.replace(f'super.onApplied({ent}, {amp});', f'super.onApplied({ent}, attributes, {amp});')
        return method
    s, n = transform_named_method(s, rx_applied, applied_transform, f'{q.name} onApplied')
    on_applied_methods += n
    q.write_text(s, encoding='utf-8')
if raw_entry_rewrites < 20:
    raise SystemExit(f'status effect raw-entry wave unexpectedly small: {raw_entry_rewrites}')
if update_methods < 6:
    raise SystemExit(f'status applyUpdateEffect wave unexpectedly small: {update_methods}')

# Attribute operation enum names and registry IDs moved after 1.20.1. Preserve config identities by
# resolving vanilla attribute IDs through Registries rather than hard-coding translation strings.
for q in java.rglob('*.java'):
    s = q.read_text(encoding='utf-8')
    s = s.replace('EntityAttributeModifier.Operation.ADD_MULTIPLIED_TOTAL', 'EntityAttributeModifier.Operation.MULTIPLY_TOTAL')
    s = s.replace('EntityAttributeModifier.Operation.ADD_MULTIPLIED_BASE', 'EntityAttributeModifier.Operation.MULTIPLY_BASE')
    s = s.replace('EntityAttributeModifier.Operation.ADD_VALUE', 'EntityAttributeModifier.Operation.ADDITION')
    s = s.replace('.attributeEntry', '.attribute')
    s = re.sub(r'EntityAttributes\.([A-Z0-9_]+)\.getIdAsString\(\)',
               r'net.minecraft.registry.Registries.ATTRIBUTE.getId(EntityAttributes.\1).toString()', s)
    s = re.sub(r'SpellSchools\.([A-Z0-9_]+)\.attribute\.getIdAsString\(\)',
               r'net.minecraft.registry.Registries.ATTRIBUTE.getId(SpellSchools.\1.attribute).toString()', s)
    s = s.replace('livingEntity.getScale()', 'net.spell_engine.compat.EntityScaleCompat.scale(livingEntity)')
    q.write_text(s, encoding='utf-8')

# Popup particle status sprites and immediate tessellation use the 1.20.1 registry/buffer APIs. Preserve
# custom atlas UVs and exact billboard geometry.
def popup_patch(s: str) -> str:
    s = s.replace('import net.minecraft.registry.RegistryKeys;\n', 'import net.minecraft.registry.Registries;\n')
    s = s.replace('import net.minecraft.registry.entry.RegistryEntry;\n', '')
    old = '''            var maybeEntry = world.getRegistryManager()\n                .get(RegistryKeys.STATUS_EFFECT)\n                .getEntry(effect.iconId);\n            if (maybeEntry.isPresent()) {\n                RegistryEntry.Reference<StatusEffect> entry = maybeEntry.get();\n                var sprite = MinecraftClient.getInstance().getStatusEffectSpriteManager().getSprite(entry);'''
    new = '''            StatusEffect statusEffect = Registries.STATUS_EFFECT.get(effect.iconId);\n            if (statusEffect != null) {\n                var sprite = MinecraftClient.getInstance().getStatusEffectSpriteManager().getSprite(statusEffect);'''
    if old not in s:
        raise SystemExit('PopupParticle status sprite seam drifted')
    s = s.replace(old, new)
    s = s.replace('BufferBuilder builder = Tessellator.getInstance().begin(VertexFormat.DrawMode.QUADS, VertexFormats.POSITION_TEXTURE_COLOR_LIGHT);',
                  'BufferBuilder builder = Tessellator.getInstance().getBuffer();\n        builder.begin(VertexFormat.DrawMode.QUADS, VertexFormats.POSITION_TEXTURE_COLOR_LIGHT);')
    s = s.replace('.light(light);\n        }\n        BuiltBuffer built = builder.endNullable();\n        if (built != null) {\n            try {\n                BufferRenderer.drawWithGlobalProgram(built);\n            } finally {\n                built.close();\n            }\n        }',
                  '.light(light).next();\n        }\n        Tessellator.getInstance().draw();')
    return s
edit('net/more_rpg_classes/client/particle/PopupParticle.java', popup_patch, 'PopupParticle 1.20.1 render API')

# Fail closed for every API family this wave claims to own.
for forbidden in (
    'net.minecraft.util.new Identifier(', 'DimensionPadding', 'StructureLiquidSettings', 'StructurePoolAliasLookup',
    '.getEffectType().value()', 'EntityAttributeModifier.Operation.ADD_MULTIPLIED_TOTAL',
    'EntityAttributeModifier.Operation.ADD_MULTIPLIED_BASE', 'EntityAttributeModifier.Operation.ADD_VALUE',
    '.attributeEntry', '.getIdAsString()', 'Tessellator.getInstance().begin(', 'BuiltBuffer built = builder.endNullable()'
):
    hits = [str(q.relative_to(java)) for q in java.rglob('*.java') if forbidden in q.read_text(encoding='utf-8')]
    if hits:
        raise SystemExit(f'wave2 owned API survived {forbidden}: {hits[:30]}')

# applyUpdateEffect must be void in 1.20.1, but canApplyUpdateEffect intentionally remains boolean.
hits = [str(q.relative_to(java)) for q in java.rglob('*.java')
        if re.search(r'boolean\s+applyUpdateEffect\s*\(', q.read_text(encoding='utf-8'))]
if hits:
    raise SystemExit(f'boolean applyUpdateEffect survived wave2: {hits[:30]}')

print('[More RPG 2.7.2] TARGET_1201_API_WAVE2_PASS '
      f'fq_identifier={fq_fixed} effect_entry={raw_entry_rewrites} update_methods={update_methods} on_applied={on_applied_methods}')
print('[More RPG 2.7.2] CONDITIONAL_JIGSAW_1201_SEMANTIC_BACKPORT_PASS')
print('[More RPG 2.7.2] STATUS_EFFECT_1201_CALLBACK_PASS')
print('[More RPG 2.7.2] ATTRIBUTE_OPERATION_AND_ID_1201_PASS')
print('[More RPG 2.7.2] POPUP_PARTICLE_1201_RENDER_PASS')
