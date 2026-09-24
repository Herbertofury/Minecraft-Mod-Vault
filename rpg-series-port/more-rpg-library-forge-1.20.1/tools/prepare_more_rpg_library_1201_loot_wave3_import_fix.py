#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 2:
    raise SystemExit('usage: prepare_more_rpg_library_1201_loot_wave3_import_fix.py <generated-port-root>')
J = Path(sys.argv[1]).resolve() / 'common/src/main/java'
if not J.is_dir():
    raise SystemExit(f'missing common Java root: {J}')

GSON = '''import com.google.gson.JsonDeserializationContext;\nimport com.google.gson.JsonObject;\nimport com.google.gson.JsonSerializationContext;\n'''
FILES = [
    'net/more_rpg_classes/util/loot/SpecificSpellScrollPoolLootFunction.java',
    'net/more_rpg_classes/util/loot/ConditionalItemLootFunction.java',
    'net/more_rpg_classes/util/loot/BindSpellFromPoolsLootFunction.java',
    'net/more_rpg_classes/util/loot/ItemTagPickerLootFunction.java',
    'net/more_rpg_classes/util/loot/ConditionalItemEntry.java',
]

moved = 0
for rel in FILES:
    p = J / rel
    if not p.is_file():
        raise SystemExit(f'missing Wave 3 loot source: {rel}')
    s = p.read_text()
    if not s.startswith(GSON):
        raise SystemExit(f'Wave 3 Gson import seam drifted before package: {rel}')
    s = s[len(GSON):]
    if not s.startswith('package '):
        raise SystemExit(f'Java package seam missing after Wave 3 import block: {rel}')
    package_end = s.find('\n')
    if package_end < 0:
        raise SystemExit(f'Java package line malformed: {rel}')
    s = s[:package_end + 1] + '\n' + GSON + s[package_end + 1:]
    if s.count('import com.google.gson.JsonDeserializationContext;') != 1 or \
       s.count('import com.google.gson.JsonObject;') != 1 or \
       s.count('import com.google.gson.JsonSerializationContext;') != 1:
        raise SystemExit(f'Gson import cardinality changed while repairing package order: {rel}')
    if not s.startswith('package '):
        raise SystemExit(f'Java source still begins before package declaration: {rel}')
    p.write_text(s)
    moved += 1

if moved != 5:
    raise SystemExit(f'Wave 3 import repair cardinality drifted: moved={moved}')
print('[More RPG 2.7.2] LOOT_SERIALIZER_1201_WAVE3_IMPORT_ORDER_PASS files=5')
