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

# 1.20.1 bundle representation: preserve nested-container behavior through BundleItem NBT.
write('net/spell_engine/compat/container/ContainerCompat.java', r'''package net.spell_engine.compat.container;
import net.spell_engine.Platform;
import net.minecraft.entity.player.PlayerEntity;
import net.minecraft.item.BundleItem;
import net.minecraft.item.ItemStack;
import net.minecraft.nbt.NbtCompound;
import net.minecraft.nbt.NbtList;
import net.minecraft.nbt.NbtElement;
import org.jetbrains.annotations.Nullable;
import java.util.*;
import java.util.function.Function;
public class ContainerCompat {
 public static final ArrayList<Function<PlayerEntity,List<ItemStack>>> providers=new ArrayList<>();
 public static final List<Resolver> resolvers=new ArrayList<>();
 public static void addProvider(Function<PlayerEntity,List<ItemStack>> provider){providers.add(provider);}
 public static void init(){resolvers.add(stack->stack.getItem() instanceof BundleItem?VanillaBundleAdapter.from(stack):null);if(Platform.util().isModLoaded("bundleapi"))CustomBundleCompat.init();}
 @Nullable public static Adapter getContainerComponent(ItemStack stack){for(var r:resolvers){var a=r.getContainerAdapter(stack);if(a!=null)return a;}return null;}
 public interface Resolver{@Nullable Adapter getContainerAdapter(ItemStack stack);}
 public interface Adapter{int size();ItemStack get(int i);Adapter createNewWithContents(List<ItemStack> contents);void attachTo(ItemStack stack);}
 public record VanillaBundleAdapter(List<ItemStack> contents) implements Adapter{
  static VanillaBundleAdapter from(ItemStack stack){var out=new ArrayList<ItemStack>();var nbt=stack.getNbt();if(nbt!=null&&nbt.contains("Items",NbtElement.LIST_TYPE)){var list=nbt.getList("Items",NbtElement.COMPOUND_TYPE);for(int i=0;i<list.size();i++)out.add(ItemStack.fromNbt(list.getCompound(i)));}return new VanillaBundleAdapter(out);}
  public int size(){return contents.size();}public ItemStack get(int i){return contents.get(i);}public Adapter createNewWithContents(List<ItemStack> c){return new VanillaBundleAdapter(List.copyOf(c));}
  public void attachTo(ItemStack stack){var list=new NbtList();for(var item:contents){var n=new NbtCompound();item.writeNbt(n);list.add(n);}stack.getOrCreateNbt().put("Items",list);}
 }
}
''')

# 1.20.1 attributes use UUID/name modifiers and slot-specific item maps rather than data components.
write('net/spell_engine/compat/item/AttributeCompat.java', r'''package net.spell_engine.compat.item;
import net.minecraft.entity.attribute.EntityAttributeModifier;
import net.minecraft.util.Identifier;
import java.nio.charset.StandardCharsets;
import java.util.UUID;
public final class AttributeCompat {
 private AttributeCompat(){}
 public static UUID uuid(Object id){if(id instanceof UUID u)return u;return UUID.nameUUIDFromBytes(String.valueOf(id).getBytes(StandardCharsets.UTF_8));}
 public static EntityAttributeModifier modifier(Object id,double value,EntityAttributeModifier.Operation operation){return new EntityAttributeModifier(uuid(id),String.valueOf(id),value,operation);}
}
''')
write('net/spell_engine/compat/item/AttributeModifierSet.java', r'''package net.spell_engine.compat.item;
import com.google.common.collect.ImmutableMultimap;
import com.google.common.collect.Multimap;
import com.mojang.serialization.Codec;
import com.mojang.serialization.codecs.RecordCodecBuilder;
import net.minecraft.entity.EquipmentSlot;
import net.minecraft.entity.attribute.*;
import net.minecraft.registry.Registries;
import net.minecraft.registry.entry.RegistryEntry;
import net.minecraft.util.Identifier;
import java.util.*;
public record AttributeModifierSet(List<Entry> modifiers){
 public static final AttributeModifierSet EMPTY=new AttributeModifierSet(List.of());
 public record Entry(Identifier attributeId,String id,double value,String operation,String slot){
  static final Codec<Entry> CODEC=RecordCodecBuilder.create(i->i.group(Identifier.CODEC.fieldOf("type").forGetter(Entry::attributeId),Codec.STRING.fieldOf("id").forGetter(Entry::id),Codec.DOUBLE.fieldOf("amount").forGetter(Entry::value),Codec.STRING.fieldOf("operation").forGetter(Entry::operation),Codec.STRING.optionalFieldOf("slot","any").forGetter(Entry::slot)).apply(i,Entry::new));
  public RegistryEntry<EntityAttribute> attribute(){return Registries.ATTRIBUTE.getEntry(attributeId).orElseThrow();}
  public EntityAttributeModifier modifier(){return AttributeCompat.modifier(id,value,switch(operation){case "multiply_base"->EntityAttributeModifier.Operation.MULTIPLY_BASE;case "multiply_total"->EntityAttributeModifier.Operation.MULTIPLY_TOTAL;default->EntityAttributeModifier.Operation.ADDITION;});}
  public boolean matches(EquipmentSlot s){return slot.equals("any")||slot.equals(s.getName());}
 }
 public static final Codec<AttributeModifierSet> CODEC=Entry.CODEC.listOf().xmap(AttributeModifierSet::new,AttributeModifierSet::modifiers);
 public List<Entry> entries(){return modifiers;}
 public Multimap<EntityAttribute,EntityAttributeModifier> forSlot(EquipmentSlot slot){var b=ImmutableMultimap.<EntityAttribute,EntityAttributeModifier>builder();for(var e:modifiers){if(e.matches(slot))b.put(e.attribute().value(),e.modifier());}return b.build();}
 public static Builder builder(){return new Builder();}
 public static class Builder{private final List<Entry> list=new ArrayList<>();
  public Builder add(RegistryEntry<EntityAttribute> attr,EntityAttributeModifier modifier,EquipmentSlot slot){return add(attr.value(),modifier,slot);}
  public Builder add(EntityAttribute attr,EntityAttributeModifier modifier,EquipmentSlot slot){var aid=Registries.ATTRIBUTE.getId(attr);list.add(new Entry(aid,modifier.getName(),modifier.getValue(),switch(modifier.getOperation()){case MULTIPLY_BASE->"multiply_base";case MULTIPLY_TOTAL->"multiply_total";default->"add_value";},slot==null?"any":slot.getName()));return this;}
  public AttributeModifierSet build(){return new AttributeModifierSet(List.copyOf(list));}
 }
}
''')
write('net/spell_engine/compat/item/ItemAttributeCompat.java', r'''package net.spell_engine.compat.item;
import net.minecraft.item.Item;import java.util.*;
public final class ItemAttributeCompat{private static final Map<Item,AttributeModifierSet> VALUES=Collections.synchronizedMap(new WeakHashMap<>());private ItemAttributeCompat(){}public static void set(Item i,AttributeModifierSet s){VALUES.put(i,s);}public static AttributeModifierSet get(Item i){return VALUES.get(i);}}
''')
write('net/spell_engine/mixin/item/ItemAttributeCompatMixin.java', r'''package net.spell_engine.mixin.item;
import com.google.common.collect.Multimap;import net.minecraft.entity.EquipmentSlot;import net.minecraft.entity.attribute.*;import net.minecraft.item.Item;import net.spell_engine.compat.item.ItemAttributeCompat;import org.spongepowered.asm.mixin.Mixin;import org.spongepowered.asm.mixin.injection.*;import org.spongepowered.asm.mixin.injection.callback.CallbackInfoReturnable;
@Mixin(Item.class) public abstract class ItemAttributeCompatMixin{@Inject(method="getAttributeModifiers",at=@At("RETURN"),cancellable=true)private void spellEngine$configured(EquipmentSlot slot,CallbackInfoReturnable<Multimap<EntityAttribute,EntityAttributeModifier>> cir){var set=ItemAttributeCompat.get((Item)(Object)this);if(set!=null){var b=com.google.common.collect.ImmutableMultimap.<EntityAttribute,EntityAttributeModifier>builder();b.putAll(cir.getReturnValue());b.putAll(set.forSlot(slot));cir.setReturnValue(b.build());}}}
''')
