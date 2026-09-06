#!/usr/bin/env python3
import re, sys
from pathlib import Path

if len(sys.argv) != 2:
    raise SystemExit('usage: apply-recipe120-foundation.py <java-root>')
root = Path(sys.argv[1]).resolve()
if not root.is_dir():
    raise SystemExit(f'missing Java root: {root}')

def write(rel, text):
    p = root / rel
    p.parent.mkdir(parents=True, exist_ok=True)
    p.write_text(text.strip() + '\n', encoding='utf-8')

base = 'com/sammy/malum/core/systems/recipe'
write(f'{base}/Forge120RecipeInput.java', r'''
package com.sammy.malum.core.systems.recipe;

import net.minecraft.world.item.ItemStack;

/** Malum custom machine input contract for Forge/Minecraft 1.20.1. */
public interface Forge120RecipeInput {
    ItemStack getItem(int index);
    int size();
}
''')
write(f'{base}/Forge120SingleRecipeInput.java', r'''
package com.sammy.malum.core.systems.recipe;

import net.minecraft.world.item.ItemStack;

public record Forge120SingleRecipeInput(ItemStack item) implements Forge120RecipeInput {
    @Override
    public ItemStack getItem(int index) {
        if (index != 0) throw new IndexOutOfBoundsException("single recipe input index " + index);
        return item;
    }

    @Override
    public int size() {
        return 1;
    }
}
''')
write(f'{base}/Forge120InWorldRecipe.java', r'''
package com.sammy.malum.core.systems.recipe;

import net.minecraft.core.RegistryAccess;
import net.minecraft.resources.ResourceLocation;
import net.minecraft.world.Container;
import net.minecraft.world.item.ItemStack;
import net.minecraft.world.item.crafting.Recipe;
import net.minecraft.world.item.crafting.RecipeSerializer;
import net.minecraft.world.item.crafting.RecipeType;
import net.minecraft.world.level.Level;

/**
 * Forge 1.20.1 owner for Malum's newer custom-machine recipes.
 * Vanilla Recipe<Container> methods let RecipeManager own and sync instances; Malum
 * machines call the strongly typed matches(T, Level) overload as in released 1.8.2.
 */
public abstract class Forge120InWorldRecipe<T extends Forge120RecipeInput> implements Recipe<Container> {
    private ResourceLocation id;
    protected final RecipeSerializer<?> serializer;
    protected final RecipeType<?> type;
    protected final ItemStack output;

    protected Forge120InWorldRecipe(RecipeSerializer<?> serializer, RecipeType<?> type) {
        this(serializer, type, ItemStack.EMPTY);
    }

    protected Forge120InWorldRecipe(RecipeSerializer<?> serializer, RecipeType<?> type, ItemStack output) {
        this.serializer = serializer;
        this.type = type;
        this.output = output == null ? ItemStack.EMPTY : output;
    }

    public final void malum$setRecipeId(ResourceLocation id) {
        if (this.id != null && !this.id.equals(id)) {
            throw new IllegalStateException("Recipe id already bound: " + this.id + " -> " + id);
        }
        this.id = id;
    }

    public abstract boolean matches(T input, Level level);

    @Override
    public final boolean matches(Container container, Level level) {
        return false;
    }

    @Override
    public ItemStack assemble(Container container, RegistryAccess registryAccess) {
        return getResultItem(registryAccess).copy();
    }

    @Override
    public boolean canCraftInDimensions(int width, int height) {
        return false;
    }

    @Override
    public ItemStack getResultItem(RegistryAccess registryAccess) {
        return output.copy();
    }

    @Override
    public ResourceLocation getId() {
        if (id == null) throw new IllegalStateException("Malum recipe id was not bound by its serializer");
        return id;
    }

    @Override
    public RecipeSerializer<?> getSerializer() {
        return serializer;
    }

    @Override
    public RecipeType<?> getType() {
        return type;
    }
}
''')

changed = 0
for p in root.rglob('*.java'):
    if p.name in {'Forge120RecipeInput.java', 'Forge120SingleRecipeInput.java', 'Forge120InWorldRecipe.java'}:
        continue
    text = p.read_text(encoding='utf-8')
    new = text
    new = re.sub(r'\bLodestoneInWorldRecipe<', 'com.sammy.malum.core.systems.recipe.Forge120InWorldRecipe<', new)
    new = re.sub(r'\bSingleRecipeInput\b', 'com.sammy.malum.core.systems.recipe.Forge120SingleRecipeInput', new)
    new = re.sub(r'\bRecipeInput\b', 'com.sammy.malum.core.systems.recipe.Forge120RecipeInput', new)
    if new != text:
        p.write_text(new, encoding='utf-8')
        changed += 1

if changed == 0:
    raise SystemExit('recipe foundation found no 1.21 recipe-input surfaces')

leftovers = []
for p in root.rglob('*.java'):
    if p.name.startswith('Forge120'):
        continue
    t = p.read_text(encoding='utf-8')
    if re.search(r'\bLodestoneInWorldRecipe\b', t): leftovers.append(f'{p}: LodestoneInWorldRecipe')
    if re.search(r'\bRecipeInput\b', t): leftovers.append(f'{p}: RecipeInput')
    if re.search(r'\bSingleRecipeInput\b', t): leftovers.append(f'{p}: SingleRecipeInput')
if leftovers:
    raise SystemExit('recipe foundation leftovers:\n' + '\n'.join(leftovers))

print(f'Forge 1.20 recipe foundation applied; source files translated: {changed}')
