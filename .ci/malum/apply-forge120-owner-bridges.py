#!/usr/bin/env python3
import re, sys
from pathlib import Path

if len(sys.argv) != 2:
    raise SystemExit('usage: apply-forge120-owner-bridges.py <java-root>')
root = Path(sys.argv[1]).resolve()
if not root.is_dir():
    raise SystemExit(f'missing Java root: {root}')

def write(rel, text):
    p = root / rel
    p.parent.mkdir(parents=True, exist_ok=True)
    p.write_text(text.strip() + '\n', encoding='utf-8')

def replace_once(rel, pattern, replacement):
    p = root / rel
    text = p.read_text(encoding='utf-8')
    text, n = re.subn(pattern, replacement, text, count=1, flags=re.M)
    if n != 1:
        raise SystemExit(f'owner bridge signature mismatch in {rel}: {pattern}')
    p.write_text(text, encoding='utf-8')

B = 'com/sammy/malum/core/systems/registry'
write(f'{B}/ForgeRegistryHolder.java', '''
package com.sammy.malum.core.systems.registry;
import net.minecraft.core.Holder;
import net.minecraft.resources.ResourceKey;
import net.minecraft.resources.ResourceLocation;
import net.minecraftforge.registries.RegistryObject;
import java.util.Objects;
import java.util.Optional;
import java.util.function.Supplier;
public class ForgeRegistryHolder<T> implements Supplier<T> {
    protected final RegistryObject<T> delegate;
    protected ForgeRegistryHolder(RegistryObject<T> delegate) { this.delegate = Objects.requireNonNull(delegate); }
    public T get() { return delegate.get(); }
    public T value() { return delegate.get(); }
    public boolean isBound() { return delegate.isPresent(); }
    public boolean isPresent() { return delegate.isPresent(); }
    public ResourceLocation getId() { return delegate.getId(); }
    public ResourceKey<T> getKey() { return delegate.getKey(); }
    public Optional<Holder<T>> getHolder() { return delegate.getHolder(); }
    public RegistryObject<T> asRegistryObject() { return delegate; }
    public boolean equals(Object other) {
        if (this == other) return true;
        if (other instanceof ForgeRegistryHolder<?> holder) return delegate.equals(holder.delegate);
        return delegate.equals(other);
    }
    public int hashCode() { return delegate.hashCode(); }
    public String toString() { return delegate.toString(); }
}
''')
write(f'{B}/ForgeRegistryBuilderCompat.java', '''
package com.sammy.malum.core.systems.registry;
import net.minecraft.resources.ResourceLocation;
import net.minecraftforge.registries.RegistryBuilder;
public final class ForgeRegistryBuilderCompat<T> {
    private final RegistryBuilder<T> delegate = new RegistryBuilder<>();
    public ForgeRegistryBuilderCompat<T> defaultKey(ResourceLocation key) { delegate.setDefaultKey(key); return this; }
    public ForgeRegistryBuilderCompat<T> sync(boolean value) { if (!value) delegate.disableSync(); return this; }
    RegistryBuilder<T> unwrap() { return delegate; }
}
''')
write(f'{B}/ForgeRegistryProxy.java', '''
package com.sammy.malum.core.systems.registry;
import com.mojang.serialization.Codec;
import com.mojang.serialization.DataResult;
import net.minecraft.core.Holder;
import net.minecraft.resources.ResourceKey;
import net.minecraft.resources.ResourceLocation;
import net.minecraftforge.registries.IForgeRegistry;
import java.util.Collection;
import java.util.Iterator;
import java.util.Objects;
import java.util.Optional;
import java.util.Set;
import java.util.function.Supplier;
public final class ForgeRegistryProxy<T> implements Iterable<T> {
    private final Supplier<IForgeRegistry<T>> supplier;
    public ForgeRegistryProxy(Supplier<IForgeRegistry<T>> supplier) { this.supplier = Objects.requireNonNull(supplier); }
    private IForgeRegistry<T> registry() {
        IForgeRegistry<T> r = supplier.get();
        if (r == null) throw new IllegalStateException("Forge registry is not available yet");
        return r;
    }
    public ResourceLocation getKey(T value) { return registry().getKey(value); }
    public Optional<ResourceKey<T>> getResourceKey(T value) { return registry().getResourceKey(value); }
    public T get(ResourceLocation id) { return registry().getValue(id); }
    public T getValue(ResourceLocation id) { return registry().getValue(id); }
    public boolean containsKey(ResourceLocation id) { return registry().containsKey(id); }
    public Set<ResourceLocation> keySet() { return registry().getKeys(); }
    public Set<ResourceLocation> getKeys() { return registry().getKeys(); }
    public Collection<T> getValues() { return registry().getValues(); }
    public Optional<Holder<T>> getHolder(ResourceLocation id) { return registry().getHolder(id); }
    public Optional<Holder<T>> getHolder(ResourceKey<T> key) { return registry().getHolder(key); }
    public Optional<Holder<T>> getHolder(T value) { return registry().getHolder(value); }
    public Holder<T> getHolderOrThrow(ResourceKey<T> key) { return registry().getDelegateOrThrow(key); }
    public Codec<T> byNameCodec() {
        return ResourceLocation.CODEC.comapFlatMap(id -> {
            T value = registry().getValue(id);
            return value == null ? DataResult.error(() -> "Unknown registry entry " + id) : DataResult.success(value);
        }, value -> {
            ResourceLocation id = registry().getKey(value);
            if (id == null) throw new IllegalStateException("Unregistered registry value: " + value);
            return id;
        });
    }
    public Codec<Holder<T>> holderByNameCodec() {
        return byNameCodec().xmap(value -> registry().getHolder(value).orElseThrow(), Holder::value);
    }
    public IForgeRegistry<T> forgeRegistry() { return registry(); }
    public Iterator<T> iterator() { return registry().iterator(); }
}
''')
write(f'{B}/ForgeDeferredRegisterCompat.java', '''
package com.sammy.malum.core.systems.registry;
import net.minecraft.core.Registry;
import net.minecraft.resources.ResourceKey;
import net.minecraft.resources.ResourceLocation;
import net.minecraftforge.eventbus.api.IEventBus;
import net.minecraftforge.registries.DeferredRegister;
import net.minecraftforge.registries.IForgeRegistry;
import net.minecraftforge.registries.RegistryObject;
import java.util.Collection;
import java.util.function.Function;
import java.util.function.Supplier;
public class ForgeDeferredRegisterCompat<T> {
    private final DeferredRegister<T> delegate;
    private final String namespace;
    protected ForgeDeferredRegisterCompat(ResourceKey<? extends Registry<T>> key, String namespace) {
        this.delegate = DeferredRegister.create(key, namespace); this.namespace = namespace;
    }
    protected <I extends T> RegistryObject<I> registerObject(String name, Supplier<? extends I> supplier) {
        return delegate.register(name, supplier);
    }
    protected <I extends T> RegistryObject<I> registerObject(String name, Function<ResourceLocation, ? extends I> factory) {
        ResourceLocation id = new ResourceLocation(namespace, name);
        return delegate.register(name, () -> factory.apply(id));
    }
    public ForgeRegistryProxy<T> makeRegistry(Function<ForgeRegistryBuilderCompat<T>, ForgeRegistryBuilderCompat<T>> factory) {
        Supplier<IForgeRegistry<T>> r = delegate.makeRegistry(() -> factory.apply(new ForgeRegistryBuilderCompat<>()).unwrap());
        return new ForgeRegistryProxy<>(r);
    }
    public void register(IEventBus bus) { delegate.register(bus); }
    public Collection<RegistryObject<T>> getEntries() { return delegate.getEntries(); }
}
''')
write(f'{B}/SpiritHolder.java', '''
package com.sammy.malum.core.systems.registry;
import com.sammy.malum.MalumMod;
import com.sammy.malum.core.systems.spirit.SpiritLike;
import com.sammy.malum.core.systems.spirit.type.SpiritArcanaType;
import com.sammy.malum.registry.common.magic.MalumSpiritTypes;
import net.minecraft.nbt.CompoundTag;
import net.minecraft.resources.ResourceLocation;
import net.minecraftforge.registries.RegistryObject;
import org.jetbrains.annotations.NotNull;
public class SpiritHolder<T extends SpiritArcanaType> extends ForgeRegistryHolder<T> implements SpiritLike {
    protected SpiritHolder(RegistryObject<T> delegate) { super(delegate); }
    public static SpiritHolder<SpiritArcanaType> getSpiritType(CompoundTag tag) { return getSpiritType(tag.getString("spirit")); }
    public static SpiritHolder<SpiritArcanaType> getSpiritType(String spirit) { return getSpiritType(new ResourceLocation(spirit)); }
    public static SpiritHolder<SpiritArcanaType> getSpiritType(ResourceLocation spirit) {
        if (spirit.getNamespace().equals("minecraft")) spirit = MalumMod.malumPath(spirit.getPath());
        return new SpiritHolder<>(RegistryObject.create(spirit, MalumSpiritTypes.SPIRIT_TYPES_KEY, MalumMod.MALUM));
    }
    public boolean is(SpiritLike spirit) { return getSpirit().equals(spirit.getSpirit()); }
    public @NotNull SpiritArcanaType getSpirit() { return get(); }
    public SpiritArcanaType orElse(SpiritArcanaType fallback) { return isBound() ? value() : fallback; }
}
''')

holder_defs = [
    ('GeasHolder', B, 'com.sammy.malum.core.systems.geas.GeasEffectType', 'GeasEffectType'),
    ('RiteEffectHolder', B + '/rite', 'com.sammy.malum.core.systems.rite.effect.SpiritRiteEffect', 'SpiritRiteEffect'),
    ('RiteHolder', B + '/rite', 'com.sammy.malum.core.systems.rite.SpiritRiteType', 'SpiritRiteType'),
]
for cls, pkg_path, type_import, base_type in holder_defs:
    pkg = pkg_path.replace('/', '.')
    parent_import = '' if pkg_path == B else 'import com.sammy.malum.core.systems.registry.ForgeRegistryHolder;\n'
    write(f'{pkg_path}/{cls}.java', f'''package {pkg};
{parent_import}import {type_import};
import net.minecraftforge.registries.RegistryObject;
public class {cls}<T extends {base_type}> extends ForgeRegistryHolder<T> {{
    protected {cls}(RegistryObject<T> delegate) {{ super(delegate); }}
}}
''')

deferred_defs = [
    ('DeferredSpiritTypes', B, 'SpiritArcanaType', 'com.sammy.malum.core.systems.spirit.type.SpiritArcanaType', 'SpiritHolder', 'MalumSpiritTypes', 'com.sammy.malum.registry.common.magic.MalumSpiritTypes', 'SPIRIT_TYPES_KEY'),
    ('DeferredGeasTypes', B, 'GeasEffectType', 'com.sammy.malum.core.systems.geas.GeasEffectType', 'GeasHolder', 'MalumGeasEffectTypes', 'com.sammy.malum.registry.common.magic.MalumGeasEffectTypes', 'GEAS_TYPES_KEY'),
    ('DeferredRiteEntityEffectTypes', B + '/rite', 'SpiritRiteEffect', 'com.sammy.malum.core.systems.rite.effect.SpiritRiteEffect', 'RiteEffectHolder', 'MalumSpiritRiteEffectTypes', 'com.sammy.malum.registry.common.magic.rite.MalumSpiritRiteEffectTypes', 'EFFECT_KEY'),
    ('DeferredRiteTypes', B + '/rite', 'SpiritRiteType', 'com.sammy.malum.core.systems.rite.SpiritRiteType', 'RiteHolder', 'MalumSpiritRiteTypes', 'com.sammy.malum.registry.common.magic.rite.MalumSpiritRiteTypes', 'RITE_KEY'),
]
for cls, pkg_path, base_type, type_import, holder, registry_cls, registry_import, key in deferred_defs:
    pkg = pkg_path.replace('/', '.')
    parent_import = '' if pkg_path == B else 'import com.sammy.malum.core.systems.registry.ForgeDeferredRegisterCompat;\n'
    write(f'{pkg_path}/{cls}.java', f'''package {pkg};
{parent_import}import {type_import};
import {registry_import};
import net.minecraft.resources.ResourceLocation;
import java.util.function.Function;
import java.util.function.Supplier;
public class {cls} extends ForgeDeferredRegisterCompat<{base_type}> {{
    protected {cls}(String namespace) {{ super({registry_cls}.{key}, namespace); }}
    public static {cls} create(String modid) {{ return new {cls}(modid); }}
    public <I extends {base_type}> {holder}<I> register(String name, Function<ResourceLocation, ? extends I> factory) {{
        return new {holder}<>(registerObject(name, factory));
    }}
    public <I extends {base_type}> {holder}<I> register(String name, Supplier<? extends I> supplier) {{
        return new {holder}<>(registerObject(name, supplier));
    }}
}}
''')

write(f'{B}/RegistryCodecBuddy.java', '''
package com.sammy.malum.core.systems.registry;
import com.mojang.datafixers.util.Pair;
import com.mojang.serialization.Codec;
import net.minecraft.core.Holder;
import net.minecraft.nbt.CompoundTag;
import net.minecraft.nbt.NbtOps;
import java.util.Optional;
import java.util.function.Function;
public class RegistryCodecBuddy<T> {
    protected final Codec<Holder<T>> holderCodec;
    protected final Codec<T> codec;
    protected final String defaultEntryName;
    public RegistryCodecBuddy(ForgeRegistryProxy<T> registry, String name) {
        holderCodec = registry.holderByNameCodec(); codec = registry.byNameCodec(); defaultEntryName = name;
    }
    public Codec<Holder<T>> getHolderCodec() { return holderCodec; }
    public Codec<T> getCodec() { return codec; }
    public void save(T entry, CompoundTag tag) { save(entry, tag, defaultEntryName); }
    public void save(T entry, CompoundTag tag, String name) { tag.put(name, codec.encodeStart(NbtOps.INSTANCE, entry).result().orElseThrow()); }
    public Optional<T> load(CompoundTag tag) { return load(tag, defaultEntryName); }
    public Optional<T> load(CompoundTag tag, String name) { return codec.decode(NbtOps.INSTANCE, tag.get(name)).map(Pair::getFirst).result(); }
    public <K extends T> Optional<K> load(CompoundTag tag, Class<K> type) { return load(tag, type, defaultEntryName); }
    public <K extends T> Optional<K> load(CompoundTag tag, Class<K> type, String name) {
        return load(tag, value -> type.isInstance(value) ? type.cast(value) : null, name);
    }
    public <K extends T> Optional<K> load(CompoundTag tag, Function<T, K> mapper) { return load(tag, mapper, defaultEntryName); }
    public <K extends T> Optional<K> load(CompoundTag tag, Function<T, K> mapper, String name) { return load(tag, name).map(mapper); }
    @SuppressWarnings("unchecked") public interface RegistryCodecBuddyHelper<T> {
        RegistryCodecBuddy<T> getCodec();
        default void save(CompoundTag tag) { getCodec().save((T)this, tag); }
        default void save(CompoundTag tag, String name) { getCodec().save((T)this, tag, name); }
    }
}
''')

for rel, var in [
    ('com/sammy/malum/registry/common/magic/MalumSpiritTypes.java', 'SPIRIT_TYPES_REGISTRY'),
    ('com/sammy/malum/registry/common/magic/MalumGeasEffectTypes.java', 'GEAS_TYPES_REGISTRY'),
    ('com/sammy/malum/registry/common/magic/rite/MalumSpiritRiteEffectTypes.java', 'EFFECT_TYPE_REGISTRY'),
    ('com/sammy/malum/registry/common/magic/rite/MalumSpiritRiteTypes.java', 'RITE_REGISTRY'),
]:
    replace_once(rel, rf'public static final Registry<([^>]+)>\s+{var}\s*=', rf'public static final com.sammy.malum.core.systems.registry.ForgeRegistryProxy<\1> {var} =')

touched = 0
for p in root.rglob('*.java'):
    text = p.read_text(encoding='utf-8')
    new = text.replace('DeferredHolder::get', 'holder -> holder.get()').replace('import net.minecraftforge.registries.DeferredHolder;\n', '')
    if new != text:
        p.write_text(new, encoding='utf-8'); touched += 1
print(f'Forge 1.20 owner bridge applied; holder/import files touched: {touched}')
