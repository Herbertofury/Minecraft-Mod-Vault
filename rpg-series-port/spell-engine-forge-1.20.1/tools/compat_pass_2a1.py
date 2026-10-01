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

# Mechanical 1.21 -> 1.20.1 / Java 17 naming differences.
for f in java.rglob('*.java'):
    s = f.read_text()
    s = s.replace('Identifier.ofVanilla(', 'new Identifier("minecraft", ')
    s = s.replace('EntityAttributeModifier.Operation.ADD_VALUE', 'EntityAttributeModifier.Operation.ADDITION')
    s = s.replace('EntityAttributeModifier.Operation.ADD_MULTIPLIED_BASE', 'EntityAttributeModifier.Operation.MULTIPLY_BASE')
    s = s.replace('EntityAttributeModifier.Operation.ADD_MULTIPLIED_TOTAL', 'EntityAttributeModifier.Operation.MULTIPLY_TOTAL')
    s = s.replace('Identifier::of', 'Identifier::new')
    f.write_text(s)

# Preserve the 1.10.2 item metadata contract on 1.20.1 using namespaced NBT + the same codecs.
write('net/spell_engine/api/spell/SpellDataComponents.java', r'''package net.spell_engine.api.spell;

import com.mojang.serialization.Codec;
import net.minecraft.item.Item;
import net.minecraft.item.ItemStack;
import net.minecraft.nbt.NbtCompound;
import net.minecraft.nbt.NbtOps;
import net.minecraft.util.Identifier;
import net.spell_engine.api.spell.container.SpellChoice;
import net.spell_engine.api.spell.container.SpellContainer;
import java.util.*;
import java.util.concurrent.ConcurrentHashMap;

/** 1.20.1 NBT-backed equivalent of Spell Engine 1.10's item data components. */
public final class SpellDataComponents {
    public record Key<T>(String name, Codec<T> codec) { }
    public static final Key<SpellContainer> SPELL_CONTAINER = new Key<>("spell_container", SpellContainer.CODEC);
    public static final Key<SpellChoice> SPELL_CHOICE = new Key<>("spell_choice", SpellChoice.CODEC);
    public static final Key<Identifier> EQUIPMENT_SET = new Key<>("equipment_set", Identifier.CODEC);
    public static final Key<Identifier> ITEM_MODEL = new Key<>("item_model", Identifier.CODEC);
    private static final String ROOT = "spell_engine_components";
    private static final Map<Item, Map<Key<?>, Object>> DEFAULTS = new ConcurrentHashMap<>();
    private SpellDataComponents() { }
    public static void init() { }
    public static <T> void setDefault(Item item, Key<T> key, T value) {
        DEFAULTS.computeIfAbsent(item, k -> new ConcurrentHashMap<>()).put(key, value);
    }
    @SuppressWarnings("unchecked")
    public static <T> T get(ItemStack stack, Key<T> key) {
        if (stack == null || stack.isEmpty()) return null;
        var root = stack.getNbt();
        if (root != null && root.contains(ROOT)) {
            var components = root.getCompound(ROOT);
            if (components.contains(key.name())) {
                var value = key.codec().parse(NbtOps.INSTANCE, components.get(key.name())).result();
                if (value.isPresent()) return value.get();
            }
        }
        var defaults = DEFAULTS.get(stack.getItem());
        return defaults == null ? null : (T) defaults.get(key);
    }
    public static <T> void set(ItemStack stack, Key<T> key, T value) {
        if (stack == null || stack.isEmpty()) return;
        var root = stack.getOrCreateNbt();
        var components = root.contains(ROOT) ? root.getCompound(ROOT) : new NbtCompound();
        if (value == null) components.remove(key.name());
        else key.codec().encodeStart(NbtOps.INSTANCE, value).result().ifPresent(nbt -> components.put(key.name(), nbt));
        root.put(ROOT, components);
    }
    public static <T> boolean contains(ItemStack stack, Key<T> key) { return get(stack, key) != null; }
    public static <T> void remove(ItemStack stack, Key<T> key) { set(stack, key, null); }
}
''')
