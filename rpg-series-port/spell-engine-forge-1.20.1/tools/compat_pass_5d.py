#!/usr/bin/env python3
from pathlib import Path
import re,sys
if len(sys.argv)!=3: raise SystemExit('usage: compat_pass_5d.py <generated-port-root> <baseline>')
J=Path(sys.argv[1]).resolve()/'common/src/main/java'
def p(r): return J/r
def ed(r,fn):
 q=p(r); old=q.read_text(); new=fn(old)
 if new==old: raise SystemExit(f'pass5d transform did not match: {r}')
 q.write_text(new)

# Minecraft 1.20.1 loot functions use Gson JsonSerializer adapters, not the 1.21 codec-based
# LootFunctionType contract. Keep the complete modern process() behavior; translate only persistence.
def spell_bind(s):
 s=s.replace('import com.mojang.serialization.Codec;\nimport com.mojang.serialization.MapCodec;\nimport com.mojang.serialization.codecs.RecordCodecBuilder;\n',
'''import com.google.gson.JsonDeserializationContext;\nimport com.google.gson.JsonObject;\nimport com.google.gson.JsonSerializationContext;\n''')
 s=s.replace('import net.minecraft.loot.provider.number.LootNumberProviderTypes;\n','')
 start=s.find('    public static final MapCodec<SpellBindRandomlyLootFunction> CODEC =')
 end=s.find('\n\n    private final LootNumberProvider tier;', start)
 if start==-1 or end==-1: raise SystemExit('SpellBind codec block not found')
 replacement='''    public static final LootFunctionType TYPE = new LootFunctionType(new Serializer());'''
 s=s[:start]+replacement+s[end:]
 s=s.replace('private SpellBindRandomlyLootFunction(List<LootCondition> conditions, String pool, LootNumberProvider tier, LootNumberProvider count) {',
             'private SpellBindRandomlyLootFunction(LootCondition[] conditions, String pool, LootNumberProvider tier, LootNumberProvider count) {')
 marker='''    public static ConditionalLootFunction.Builder<?> builder(String pool, LootNumberProvider tier, LootNumberProvider count) {'''
 if marker not in s: raise SystemExit('SpellBind builder marker not found')
 serializer='''    public static class Serializer extends ConditionalLootFunction.Serializer<SpellBindRandomlyLootFunction> {\n        @Override\n        public void toJson(JsonObject json, SpellBindRandomlyLootFunction function, JsonSerializationContext context) {\n            super.toJson(json, function, context);\n            if (function.pool != null) json.addProperty("pool", function.pool);\n            json.add("tier", context.serialize(function.tier));\n            if (function.count != null) json.add("count", context.serialize(function.count));\n        }\n\n        @Override\n        public SpellBindRandomlyLootFunction fromJson(JsonObject json, JsonDeserializationContext context, LootCondition[] conditions) {\n            String pool = json.has("pool") && !json.get("pool").isJsonNull() ? json.get("pool").getAsString() : null;\n            LootNumberProvider tier = context.deserialize(json.get("tier"), LootNumberProvider.class);\n            LootNumberProvider count = json.has("count") && !json.get("count").isJsonNull()\n                    ? context.deserialize(json.get("count"), LootNumberProvider.class) : null;\n            return new SpellBindRandomlyLootFunction(conditions, pool, tier, count);\n        }\n    }\n\n'''
 s=s.replace(marker,serializer+marker)
 return s
ed('net/spell_engine/spellbinding/SpellBindRandomlyLootFunction.java',spell_bind)

# Historical EnchantmentSpecificCriteria takes its stable string ID.
ed('net/spell_engine/mixin/criteria/EnchantedItemCriterionMixin.java',
   lambda s:s.replace('EnchantmentSpecificCriteria.INSTANCE.trigger(player, id);','EnchantmentSpecificCriteria.INSTANCE.trigger(player, id.toString());'))

# Mixin accessor casts pass through Object because javac sees vanilla provider classes as concrete/final;
# Mixin adds the accessor interface at runtime to the transformed class.
def loot(s):
 s=s.replace('((ConstantLootNumberProviderAccessor) constant)', '((ConstantLootNumberProviderAccessor) (Object) constant)')
 s=s.replace('((ConstantLootNumberProviderAccessor) lo)', '((ConstantLootNumberProviderAccessor) (Object) lo)')
 s=s.replace('((ConstantLootNumberProviderAccessor) hi)', '((ConstantLootNumberProviderAccessor) (Object) hi)')
 s=s.replace('((UniformLootNumberProviderAccessor) uniform)', '((UniformLootNumberProviderAccessor) (Object) uniform)')
 return s
ed('net/spell_engine/rpg_series/loot/LootHelper.java',loot)

for needle in ('MapCodec<SpellBindRandomlyLootFunction>','LootNumberProviderTypes.CODEC','List<LootCondition> conditions','trigger(player, id);','((ConstantLootNumberProviderAccessor) constant)','((ConstantLootNumberProviderAccessor) lo)','((UniformLootNumberProviderAccessor) uniform)'):
 hits=[str(q.relative_to(J)) for q in J.rglob('*.java') if needle in q.read_text()]
 if hits: raise SystemExit(f'pass5d incomplete {needle}: {hits[:20]}')
print('Spell Engine compatibility pass 5d applied: 1.20.1 loot JsonSerializer + criterion ID + Mixin accessor casts')
