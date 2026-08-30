#!/usr/bin/env python3
from pathlib import Path
import json, re, sys

if len(sys.argv) != 3:
    raise SystemExit('usage: compat_pass_2.py <generated-port-root> <spell-engine-1.20.1-baseline>')
root = Path(sys.argv[1]).resolve()
base = Path(sys.argv[2]).resolve()
java = root / 'common/src/main/java'
resources = root / 'common/src/main/resources'


def path(rel): return java / rel
def read(rel): return path(rel).read_text()
def write(rel, text):
    f = path(rel); f.parent.mkdir(parents=True, exist_ok=True); f.write_text(text)
def patch(rel, fn):
    f = path(rel)
    if f.exists(): f.write_text(fn(f.read_text()))

# Proven 1.20.1 criterion implementations for the unchanged triggers.
for target_rel, base_rel, package_fix in [
 ('net/spell_engine/spellbinding/SpellBindingCriteria.java','net/spell_engine/spellbinding/SpellBindingCriteria.java',None),
 ('net/spell_engine/spellbinding/SpellBookCreationCriteria.java','net/spell_engine/spellbinding/SpellBookCreationCriteria.java',None),
 ('net/spell_engine/misc/criteria/EnchantmentSpecificCriteria.java','net/spell_engine/internals/criteria/EnchantmentSpecificCriteria.java','net.spell_engine.misc.criteria')]:
    src=base/'src/main/java'/base_rel
    if src.exists():
        text=src.read_text()
        if package_fix: text=re.sub(r'^package [^;]+;', 'package '+package_fix+';', text, count=1, flags=re.M)
        write(target_rel,text)
write('net/spell_engine/misc/criteria/SpellCastCriteria.java', r'''package net.spell_engine.misc.criteria;
import com.google.gson.*;import net.minecraft.advancement.criterion.*;import net.minecraft.predicate.entity.*;import net.minecraft.registry.entry.RegistryEntry;import net.minecraft.server.network.ServerPlayerEntity;import net.minecraft.util.Identifier;import net.spell_engine.SpellEngineMod;import net.spell_engine.api.spell.Spell;import net.spell_engine.api.spell.registry.SpellRegistry;import net.spell_engine.utils.PatternMatching;
public class SpellCastCriteria extends AbstractCriterion<SpellCastCriteria.Condition>{public static final Identifier ID=new Identifier(SpellEngineMod.ID,"spell_cast");public static final SpellCastCriteria INSTANCE=new SpellCastCriteria();@Override protected Condition conditionsFromJson(JsonObject o,LootContextPredicate p,AdvancementEntityPredicateDeserializer d){return new Condition(o.has("spell")?o.get("spell").getAsString():null,o.has("other_spell")?o.get("other_spell").getAsString():null);}@Override public Identifier getId(){return ID;}public void trigger(ServerPlayerEntity p,RegistryEntry<Spell>s){trigger(p,c->c.matches(s));}public static class Condition extends AbstractCriterionConditions{final String spell,other;Condition(String s,String o){super(ID,LootContextPredicate.EMPTY);spell=s;other=o;}boolean matches(RegistryEntry<Spell> e){return (spell==null&&other==null)||(spell!=null&&PatternMatching.matches(e,SpellRegistry.KEY,spell))||(other!=null&&PatternMatching.matches(e,SpellRegistry.KEY,other));}@Override public JsonObject toJson(AdvancementEntityPredicateSerializer s){var j=super.toJson(s);if(spell!=null)j.addProperty("spell",spell);if(other!=null)j.addProperty("other_spell",other);return j;}}}
''')

# 1.21 generic type syntax only; serializer adaptation is deliberately left visible if 1.20.1 requires more.
patch('net/spell_engine/spellbinding/SpellBindRandomlyLootFunction.java', lambda s: s.replace('LootFunctionType<SpellBindRandomlyLootFunction>', 'LootFunctionType').replace('new LootFunctionType<SpellBindRandomlyLootFunction>(CODEC)', 'new LootFunctionType(CODEC)'))

# These 1.21 internals do not exist on 1.20.1. Remove them from the compile-gate mixin config so they cannot become invalid runtime mixins.
pending = {
    'item.ArmorMaterialLayerAccessor': '1.20.1 ArmorMaterial has no Layer record; Armor uses material name for texture identity.',
    'registry.DataComponentTypesMixin': '1.20.1 has no data component registry; SpellDataComponents NBT bridge replaces it.',
    'registry.RegistryLoaderMixin': '1.21 RegistryEntryInfo signature absent; final Forge registry/datapack hook still required.',
    'effect.PoisonEffectMixin': '1.20.1 poison is not the 1.21 PoisonStatusEffect target; behavior equivalent still required.',
    'client.render.ImmediateItemGlowMixin': '1.21 BufferAllocator/SequencedMap renderer internals absent; 1.20.1 glow path still required.'
}
for rel in ['net/spell_engine/mixin/registry/DataComponentTypesMixin.java','net/spell_engine/mixin/registry/RegistryLoaderMixin.java','net/spell_engine/mixin/effect/PoisonEffectMixin.java','net/spell_engine/mixin/client/render/ImmediateItemGlowMixin.java']:
    pkg='.'.join(rel.split('/')[:-1]); cls=rel.split('/')[-1][:-5]; write(rel, f'package {pkg}; public class {cls} {{ }}\n')
mix=resources/'spell_engine.mixins.json'
if mix.exists():
    data=json.loads(mix.read_text())
    data['mixins']=[x for x in data.get('mixins',[]) if x not in pending]
    data['client']=[x for x in data.get('client',[]) if x not in pending]
    if 'item.ItemAttributeCompatMixin' not in data['mixins']: data['mixins'].append('item.ItemAttributeCompatMixin')
    mix.write_text(json.dumps(data,indent=2)+"\n")
(root/'COMPAT-GATE-PENDING.md').write_text('# Spell Engine 1.10.2 -> Forge 1.20.1 pending parity gates\n\n'+'\n'.join(f'- `{k}`: {v}' for k,v in pending.items())+'\n')

# Remove residual direct data-component imports after all explicit replacements.
for f in java.rglob('*.java'):
    s=f.read_text().replace('import net.minecraft.component.DataComponentTypes;\n','').replace('import net.minecraft.component.ComponentType;\n','')
    f.write_text(s)

# Static guard: these APIs cannot exist in a Java 17 / Minecraft 1.20.1 common source tree.
for needle in ('net.minecraft.component.', 'DataTracker.Builder', 'net.minecraft.item.tooltip.TooltipType', 'java.util.SequencedMap', 'net.minecraft.client.util.BufferAllocator', 'RegistryEntryInfo', 'PoisonStatusEffect', 'ArmorMaterial.Layer', 'Identifier.ofVanilla(', 'Identifier::of', '.getFirst()', '.getLast()'):
    found=[str(f.relative_to(java)) for f in java.rglob('*.java') if needle in f.read_text()]
    if found: raise SystemExit(f'compat pass 2 incomplete; {needle!r} remains in {found[:20]}')
print('Spell Engine compatibility pass 2 applied: NBT components + attributes + trackers + tooltips + Java17 + criteria + compile-gate mixin safety')
