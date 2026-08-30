#!/usr/bin/env python3
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile

if len(sys.argv) != 4:
    raise SystemExit('usage: prepare_more_rpg_library_mc1201_worldgen_api.py <modern-2.7.2-root> <old-1.20.1-root> <output-root>')
modern = Path(sys.argv[1]).resolve()
old = Path(sys.argv[2]).resolve()
out = Path(sys.argv[3]).resolve()
loader_neutral = Path(__file__).with_name('prepare_more_rpg_library_loader_neutral.py')
for p in (modern, old):
    if not p.is_dir():
        raise SystemExit(f'missing authority tree: {p}')
if not loader_neutral.is_file():
    raise SystemExit(f'missing loader-neutral preparer: {loader_neutral}')

# Quarantined MC 1.20.1 syntax/API bridge for the modern 2.7.2 common tree. Every rewrite here is
# behavior-neutral: it changes only APIs whose 1.20.1 representation differs from the 1.21.1 source.
# Newer worldgen features are never deleted. ConditionalJigsawStructure's newer dimension/liquid
# semantics are deliberately NOT touched here; they need their own compiler/runtime-proven backport.
with tempfile.TemporaryDirectory(prefix='more-rpg-272-mc1201-api-') as td:
    staged = Path(td) / 'modern-272'
    shutil.copytree(modern, staged)
    java = staged / 'common/src/main/java'
    if not java.is_dir():
        raise SystemExit('modern 2.7.2 common Java tree missing')

    identifier_rewrites = 0
    for f in sorted(java.rglob('*.java')):
        s = f.read_text(errors='strict')
        count = s.count('Identifier.of(')
        if count:
            identifier_rewrites += count
            s = s.replace('Identifier.of(', 'new Identifier(')
            f.write_text(s)
    if identifier_rewrites < 4:
        raise SystemExit(f'expected multiple 1.21 Identifier.of calls, adapted only {identifier_rewrites}')

    processors = [
        java / 'net/more_rpg_classes/worldgen/processor/PathAdaptationProcessor.java',
        java / 'net/more_rpg_classes/worldgen/processor/WaterPillarProcessor.java',
        java / 'net/more_rpg_classes/worldgen/processor/TerrainBlendingProcessor.java',
    ]
    for f in processors:
        if not f.is_file():
            raise SystemExit(f'modern worldgen processor missing: {f.relative_to(staged)}')
        s = f.read_text(errors='strict')
        if s.count('import com.mojang.serialization.MapCodec;') != 1:
            raise SystemExit(f'MapCodec import seam drifted in {f.name}')
        s = s.replace('import com.mojang.serialization.MapCodec;\n', '')
        class_name = f.stem
        old_decl = f'public static final MapCodec<{class_name}> CODEC = RecordCodecBuilder.mapCodec(instance ->'
        new_decl = f'public static final Codec<{class_name}> CODEC = RecordCodecBuilder.create(instance ->'
        if s.count(old_decl) != 1:
            raise SystemExit(f'MapCodec declaration seam drifted in {f.name}')
        s = s.replace(old_decl, new_decl, 1)

        # Yarn/MC 1.20.1 StructureBlockInfo is a normal data holder with public fields. 1.21.1 exposes
        # record-style accessors. Limit this rewrite to the three processor files so unrelated records
        # in the modern source are untouched.
        accessor_counts = {
            '.state()': s.count('currentBlockInfo.state()'),
            '.pos()': s.count('currentBlockInfo.pos()'),
            '.nbt()': s.count('currentBlockInfo.nbt()'),
        }
        if accessor_counts['.state()'] < 1 or accessor_counts['.pos()'] < 1:
            raise SystemExit(f'StructureBlockInfo accessor seam drifted in {f.name}: {accessor_counts}')
        s = s.replace('currentBlockInfo.state()', 'currentBlockInfo.state')
        s = s.replace('currentBlockInfo.pos()', 'currentBlockInfo.pos')
        s = s.replace('currentBlockInfo.nbt()', 'currentBlockInfo.nbt')
        f.write_text(s)

    # This pass owns only the known processor MapCodec family; do not accidentally hide the newer
    # ConditionalJigsawStructure contract. Its MapCodec remains as an intentional next-frontier marker.
    remaining_processor_mapcodec = []
    for f in processors:
        text = f.read_text(errors='strict')
        if 'MapCodec<' in text or 'RecordCodecBuilder.mapCodec' in text:
            remaining_processor_mapcodec.append(f.name)
    if remaining_processor_mapcodec:
        raise SystemExit('processor MapCodec conversion incomplete: ' + ', '.join(remaining_processor_mapcodec))

    subprocess.run([sys.executable, str(loader_neutral), str(staged), str(old), str(out)], check=True)

print('[More RPG 2.7.2] MC1201_MECHANICAL_API_BRIDGE_PASS '
      f'identifier_rewrites={identifier_rewrites} processors=3 conditional_jigsaw_semantics_deferred')
