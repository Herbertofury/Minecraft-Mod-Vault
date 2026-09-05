#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 2:
    raise SystemExit('usage: prepare_more_rpg_library_1201_loot_wave3.py <generated-port-root>')
J = Path(sys.argv[1]).resolve() / 'common/src/main/java'
if not J.is_dir():
    raise SystemExit(f'missing common Java root: {J}')

GSON = '''import com.google.gson.JsonDeserializationContext;\nimport com.google.gson.JsonObject;\nimport com.google.gson.JsonSerializationContext;\n'''


def path(rel):
    p = J / rel
    if not p.is_file():
        raise SystemExit(f'missing loot wave3 seam: {rel}')
    return p


def replace_once(text, old, new, label):
    n = text.count(old)
    if n != 1:
        raise SystemExit(f'{label}: expected 1 occurrence, found {n}')
    return text.replace(old, new, 1)


def strip_codec_block(text, class_name, field_marker):
    start = text.find(f'    public static final MapCodec<{class_name}> CODEC =')
    end = text.find(field_marker, start)
    if start < 0 or end < 0:
        raise SystemExit(f'{class_name}: codec block seam drifted')
    return text[:start] + f'    public static final LootFunctionType TYPE = new LootFunctionType(new Serializer());\n' + text[end:]


def common_imports(text):
    text = text.replace('import com.mojang.serialization.Codec;\n', '')
    text = text.replace('import com.mojang.serialization.MapCodec;\n', '')
    text = text.replace('import com.mojang.serialization.codecs.RecordCodecBuilder;\n', '')
    text = text.replace('import net.minecraft.loot.provider.number.LootNumberProviderTypes;\n', '')
    if 'import com.google.gson.JsonObject;' not in text:
        text = GSON + text
    return text


def patch_specific():
    p = path('net/more_rpg_classes/util/loot/SpecificSpellScrollPoolLootFunction.java')
    s = common_imports(p.read_text())
    s = strip_codec_block(s, 'SpecificSpellScrollPoolLootFunction', '\n\n\n    @Nullable private final List<String> pools;')
    s = replace_once(s, '            List<LootCondition> conditions,', '            LootCondition[] conditions,', 'specific conditions')
    s = s.replace('public LootFunctionType<SpecificSpellScrollPoolLootFunction> getType()', 'public LootFunctionType getType()')
    marker = '    public static Builder<?> builder(\n'
    serializer = '''    public static class Serializer extends ConditionalLootFunction.Serializer<SpecificSpellScrollPoolLootFunction> {\n        @Override\n        public void toJson(JsonObject json, SpecificSpellScrollPoolLootFunction function, JsonSerializationContext context) {\n            super.toJson(json, function, context);\n            json.add("spell_pools", context.serialize(function.pools != null ? function.pools : List.of()));\n            json.add("spell_tier_min", context.serialize(function.tierMin));\n            json.add("spell_tier_max", context.serialize(function.tierMax));\n            if (function.count != null) json.add("count", context.serialize(function.count));\n            json.add("blacklist_spells", context.serialize(function.blacklist != null ? function.blacklist : List.of()));\n        }\n\n        @Override\n        public SpecificSpellScrollPoolLootFunction fromJson(JsonObject json, JsonDeserializationContext context, LootCondition[] conditions) {\n            List<String> pools = json.has("spell_pools") ? java.util.Arrays.asList(context.deserialize(json.get("spell_pools"), String[].class)) : List.of();\n            LootNumberProvider tierMin = context.deserialize(json.get("spell_tier_min"), LootNumberProvider.class);\n            LootNumberProvider tierMax = context.deserialize(json.get("spell_tier_max"), LootNumberProvider.class);\n            LootNumberProvider count = json.has("count") ? context.deserialize(json.get("count"), LootNumberProvider.class) : null;\n            List<String> blacklist = json.has("blacklist_spells") ? java.util.Arrays.asList(context.deserialize(json.get("blacklist_spells"), String[].class)) : List.of();\n            return new SpecificSpellScrollPoolLootFunction(conditions, pools, tierMin, tierMax, count, blacklist);\n        }\n    }\n\n'''
    s = replace_once(s, marker, serializer + marker, 'specific serializer insertion')
    p.write_text(s)


def patch_conditional():
    p = path('net/more_rpg_classes/util/loot/ConditionalItemLootFunction.java')
    s = common_imports(p.read_text())
    s = strip_codec_block(s, 'ConditionalItemLootFunction', '\n\n    private final Identifier conditionalItemId;')
    s = replace_once(s, '            List<LootCondition> conditions,', '            LootCondition[] conditions,', 'conditional conditions')
    s = s.replace('public LootFunctionType<ConditionalItemLootFunction> getType()', 'public LootFunctionType getType()')
    marker = '    public static Builder<?> builder(Identifier conditionalItem) {'
    serializer = '''    public static class Serializer extends ConditionalLootFunction.Serializer<ConditionalItemLootFunction> {\n        @Override\n        public void toJson(JsonObject json, ConditionalItemLootFunction function, JsonSerializationContext context) {\n            super.toJson(json, function, context);\n            json.addProperty("conditional_item", function.conditionalItemId.toString());\n        }\n\n        @Override\n        public ConditionalItemLootFunction fromJson(JsonObject json, JsonDeserializationContext context, LootCondition[] conditions) {\n            return new ConditionalItemLootFunction(conditions, new Identifier(json.get("conditional_item").getAsString()));\n        }\n    }\n\n'''
    s = replace_once(s, marker, serializer + marker, 'conditional serializer insertion')
    p.write_text(s)


def patch_bind():
    p = path('net/more_rpg_classes/util/loot/BindSpellFromPoolsLootFunction.java')
    s = common_imports(p.read_text())
    s = strip_codec_block(s, 'BindSpellFromPoolsLootFunction', '\n\n    private final List<String> pools;')
    s = replace_once(s, '            List<LootCondition> conditions,', '            LootCondition[] conditions,', 'bind conditions')
    s = s.replace('public LootFunctionType<BindSpellFromPoolsLootFunction> getType()', 'public LootFunctionType getType()')
    marker = '    public static ConditionalLootFunction.Builder<?> builder(\n'
    serializer = '''    public static class Serializer extends ConditionalLootFunction.Serializer<BindSpellFromPoolsLootFunction> {\n        @Override\n        public void toJson(JsonObject json, BindSpellFromPoolsLootFunction function, JsonSerializationContext context) {\n            super.toJson(json, function, context);\n            json.add("spell_pools", context.serialize(function.pools));\n            json.add("count", context.serialize(function.count));\n            if (function.chance != null) json.add("chance", context.serialize(function.chance));\n        }\n\n        @Override\n        public BindSpellFromPoolsLootFunction fromJson(JsonObject json, JsonDeserializationContext context, LootCondition[] conditions) {\n            List<String> pools = json.has("spell_pools") ? java.util.Arrays.asList(context.deserialize(json.get("spell_pools"), String[].class)) : List.of();\n            LootNumberProvider count = json.has("count") ? context.deserialize(json.get("count"), LootNumberProvider.class) : ConstantLootNumberProvider.create(1);\n            LootNumberProvider chance = json.has("chance") ? context.deserialize(json.get("chance"), LootNumberProvider.class) : null;\n            return new BindSpellFromPoolsLootFunction(conditions, pools, count, Optional.ofNullable(chance));\n        }\n    }\n\n'''
    s = replace_once(s, marker, serializer + marker, 'bind serializer insertion')
    p.write_text(s)


def patch_item_tag():
    p = path('net/more_rpg_classes/util/loot/ItemTagPickerLootFunction.java')
    s = common_imports(p.read_text())
    s = strip_codec_block(s, 'ItemTagPickerLootFunction', '\n\n    private final List<String> itemTags;')
    s = replace_once(s, '            List<LootCondition> conditions,', '            LootCondition[] conditions,', 'itemtag conditions')
    s = s.replace('public LootFunctionType<ItemTagPickerLootFunction> getType()', 'public LootFunctionType getType()')
    marker = '    public static Builder<?> builder(\n'
    serializer = '''    public static class Serializer extends ConditionalLootFunction.Serializer<ItemTagPickerLootFunction> {\n        @Override\n        public void toJson(JsonObject json, ItemTagPickerLootFunction function, JsonSerializationContext context) {\n            super.toJson(json, function, context);\n            json.add("item_tags", context.serialize(function.itemTags));\n            json.add("count", context.serialize(function.count));\n            json.addProperty("enchantment_probability", function.enchantmentProbability);\n            json.add("enchantment_level_min", context.serialize(function.enchantmentLevelMin));\n            json.add("enchantment_level_max", context.serialize(function.enchantmentLevelMax));\n        }\n\n        @Override\n        public ItemTagPickerLootFunction fromJson(JsonObject json, JsonDeserializationContext context, LootCondition[] conditions) {\n            List<String> tags = json.has("item_tags") ? java.util.Arrays.asList(context.deserialize(json.get("item_tags"), String[].class)) : List.of();\n            LootNumberProvider count = json.has("count") ? context.deserialize(json.get("count"), LootNumberProvider.class) : ConstantLootNumberProvider.create(1);\n            float probability = json.has("enchantment_probability") ? json.get("enchantment_probability").getAsFloat() : 0.0F;\n            LootNumberProvider min = json.has("enchantment_level_min") ? context.deserialize(json.get("enchantment_level_min"), LootNumberProvider.class) : ConstantLootNumberProvider.create(1);\n            LootNumberProvider max = json.has("enchantment_level_max") ? context.deserialize(json.get("enchantment_level_max"), LootNumberProvider.class) : ConstantLootNumberProvider.create(3);\n            return new ItemTagPickerLootFunction(conditions, tags, count, probability, min, max);\n        }\n    }\n\n'''
    s = replace_once(s, marker, serializer + marker, 'itemtag serializer insertion')
    p.write_text(s)


patch_specific()
patch_conditional()
patch_bind()
patch_item_tag()

# Custom leaf entry uses 1.20.1 LeafEntry.Serializer rather than MapCodec. Preserve item/count fields
# plus the inherited weight, quality, conditions and functions arrays supplied by the base serializer.
p = path('net/more_rpg_classes/util/loot/ConditionalItemEntry.java')
s = p.read_text()
for imp in ('import com.mojang.serialization.Codec;\n', 'import com.mojang.serialization.MapCodec;\n',
            'import com.mojang.serialization.codecs.RecordCodecBuilder;\n', 'import net.minecraft.loot.function.LootFunctionTypes;\n',
            'import net.minecraft.loot.provider.number.LootNumberProviderTypes;\n'):
    s = s.replace(imp, '')
if 'import com.google.gson.JsonObject;' not in s:
    s = GSON + s
s = replace_once(s, '            List<LootCondition> conditions,\n            List<LootFunction> functions',
                 '            LootCondition[] conditions,\n            LootFunction[] functions', 'entry ctor arrays')
start = s.find('    public static final MapCodec<ConditionalItemEntry> CODEC =')
if start < 0:
    raise SystemExit('ConditionalItemEntry codec seam drifted')
end = s.find('\n\n}', start)
if end < 0:
    raise SystemExit('ConditionalItemEntry class tail seam drifted')
serializer = '''    public static class Serializer extends LeafEntry.Serializer<ConditionalItemEntry> {\n        @Override\n        public void addEntryFields(JsonObject json, ConditionalItemEntry entry, JsonSerializationContext context) {\n            super.addEntryFields(json, entry, context);\n            json.addProperty("item", entry.itemId.toString());\n            if (entry.count.isPresent()) json.add("count", context.serialize(entry.count.get()));\n        }\n\n        @Override\n        protected ConditionalItemEntry fromJson(JsonObject json, JsonDeserializationContext context, int weight, int quality, LootCondition[] conditions, LootFunction[] functions) {\n            Identifier item = new Identifier(json.get("item").getAsString());\n            Optional<LootNumberProvider> count = json.has("count")\n                    ? Optional.of(context.deserialize(json.get("count"), LootNumberProvider.class)) : Optional.empty();\n            return new ConditionalItemEntry(item, count, weight, quality, conditions, functions);\n        }\n    }\n'''
s = s[:start] + serializer + s[end:]
# Public modern accessors are retained where target fields remain directly available; list-shaped
# accessors adapt target arrays without changing their external return type.
s = s.replace('public List<LootCondition> getConditionsList() { return this.conditions; }',
              'public List<LootCondition> getConditionsList() { return java.util.Arrays.asList(this.conditions); }')
s = s.replace('public List<LootFunction> getFunctionsList() { return this.functions; }',
              'public List<LootFunction> getFunctionsList() { return java.util.Arrays.asList(this.functions); }')
p.write_text(s)

# Registration must pass the 1.20.1 JsonSerializer adapter, not a MapCodec.
p = path('net/more_rpg_classes/MRPGCMod.java')
s = p.read_text()
s = replace_once(s, 'new LootPoolEntryType(ConditionalItemEntry.CODEC)',
                 'new LootPoolEntryType(new ConditionalItemEntry.Serializer())', 'entry type registration')
p.write_text(s)

# Fail closed on the persistence-era APIs owned by this wave.
for needle in ('MapCodec<SpecificSpellScrollPoolLootFunction>', 'MapCodec<ConditionalItemLootFunction>',
               'MapCodec<BindSpellFromPoolsLootFunction>', 'MapCodec<ItemTagPickerLootFunction>',
               'MapCodec<ConditionalItemEntry>', 'LootFunctionType<SpecificSpellScrollPoolLootFunction>',
               'LootFunctionType<ConditionalItemLootFunction>', 'LootFunctionType<BindSpellFromPoolsLootFunction>',
               'LootFunctionType<ItemTagPickerLootFunction>', 'List<LootCondition> conditions'):
    hits = [str(q.relative_to(J)) for q in J.rglob('*.java') if needle in q.read_text()]
    if hits:
        raise SystemExit(f'loot wave3 incomplete {needle}: {hits[:30]}')
print('[More RPG 2.7.2] LOOT_SERIALIZER_1201_WAVE3_PASS functions=4 leaf_entries=1')
print('[More RPG 2.7.2] MODERN_LOOT_PROCESS_LOGIC_PRESERVED serializer_only=true')
