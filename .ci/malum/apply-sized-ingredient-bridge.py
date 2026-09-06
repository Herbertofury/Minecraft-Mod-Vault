#!/usr/bin/env python3
import sys
from pathlib import Path

if len(sys.argv) != 2:
    raise SystemExit('usage: apply-sized-ingredient-bridge.py <java-root>')

root = Path(sys.argv[1]).resolve()
if not root.is_dir():
    raise SystemExit(f'missing Java root: {root}')

target = root / 'net/minecraftforge/common/crafting/SizedIngredient.java'
target.parent.mkdir(parents=True, exist_ok=True)
target.write_text(r'''package net.minecraftforge.common.crafting;

import com.google.gson.JsonArray;
import com.google.gson.JsonElement;
import com.google.gson.JsonObject;
import com.google.gson.JsonParseException;
import com.mojang.serialization.Codec;
import com.mojang.serialization.Dynamic;
import com.mojang.serialization.JsonOps;
import net.minecraft.tags.TagKey;
import net.minecraft.world.item.Item;
import net.minecraft.world.item.ItemStack;
import net.minecraft.world.item.crafting.Ingredient;
import net.minecraft.world.level.ItemLike;

import java.util.Arrays;
import java.util.Objects;

/**
 * Forge 1.20 compatibility owner for NeoForge's count-aware SizedIngredient.
 *
 * Malum 1.8.2 uses this as a value object/recipe codec; it does not consume the
 * 1.21 StreamCodec surface. The bridge therefore keeps the shipped recipe
 * semantics without importing the 1.21 networking stack.
 */
public final class SizedIngredient {
    public static final Codec<SizedIngredient> FLAT_CODEC = Codec.PASSTHROUGH.xmap(
            dynamic -> decodeFlat(dynamic.convert(JsonOps.INSTANCE).getValue()),
            value -> new Dynamic<>(JsonOps.INSTANCE, encodeFlat(value)));

    public static final Codec<SizedIngredient> NESTED_CODEC = Codec.PASSTHROUGH.xmap(
            dynamic -> decodeNested(dynamic.convert(JsonOps.INSTANCE).getValue()),
            value -> new Dynamic<>(JsonOps.INSTANCE, encodeNested(value)));

    private final Ingredient ingredient;
    private final int count;
    private ItemStack[] cachedStacks;

    public SizedIngredient(Ingredient ingredient, int count) {
        if (count <= 0) {
            throw new IllegalArgumentException("Size must be positive");
        }
        this.ingredient = Objects.requireNonNull(ingredient, "ingredient");
        this.count = count;
    }

    public static SizedIngredient of(ItemLike item, int count) {
        return new SizedIngredient(Ingredient.of(item), count);
    }

    public static SizedIngredient of(TagKey<Item> tag, int count) {
        return new SizedIngredient(Ingredient.of(tag), count);
    }

    public Ingredient ingredient() {
        return ingredient;
    }

    public int count() {
        return count;
    }

    public boolean test(ItemStack stack) {
        return ingredient.test(stack) && stack.getCount() >= count;
    }

    /** Returns the represented ingredient stacks with the required count applied. */
    public ItemStack[] getItems() {
        if (cachedStacks == null) {
            cachedStacks = Arrays.stream(ingredient.getItems()).map(stack -> {
                ItemStack copy = stack.copy();
                copy.setCount(count);
                return copy;
            }).toArray(ItemStack[]::new);
        }
        return cachedStacks;
    }

    private static SizedIngredient decodeFlat(JsonElement element) {
        if (element == null || element.isJsonNull()) {
            throw new JsonParseException("Sized ingredient cannot be null");
        }
        if (element.isJsonArray()) {
            return new SizedIngredient(Ingredient.fromJson(element), 1);
        }
        if (!element.isJsonObject()) {
            throw new JsonParseException("Sized ingredient must be an object or array");
        }

        JsonObject object = element.getAsJsonObject().deepCopy();
        int count = object.has("count") ? object.remove("count").getAsInt() : 1;
        if (count <= 0) {
            throw new JsonParseException("Sized ingredient count must be positive");
        }

        // Be permissive with NeoForge's release-era flat compound representation.
        if (object.has("ingredient")) {
            return new SizedIngredient(Ingredient.fromJson(normalizeIngredientJson(object.get("ingredient"))), count);
        }
        if (object.has("ingredients") && isCompoundType(object)) {
            return new SizedIngredient(Ingredient.fromJson(normalizeIngredientJson(object.get("ingredients"))), count);
        }
        return new SizedIngredient(Ingredient.fromJson(normalizeIngredientJson(object)), count);
    }

    private static SizedIngredient decodeNested(JsonElement element) {
        if (element == null || !element.isJsonObject()) {
            throw new JsonParseException("Nested sized ingredient must be an object");
        }
        JsonObject object = element.getAsJsonObject();
        if (!object.has("ingredient")) {
            throw new JsonParseException("Nested sized ingredient is missing 'ingredient'");
        }
        int count = object.has("count") ? object.get("count").getAsInt() : 1;
        if (count <= 0) {
            throw new JsonParseException("Sized ingredient count must be positive");
        }
        return new SizedIngredient(Ingredient.fromJson(normalizeIngredientJson(object.get("ingredient"))), count);
    }

    private static JsonElement encodeFlat(SizedIngredient value) {
        JsonElement ingredientJson = value.ingredient.toJson();
        if (ingredientJson.isJsonObject()) {
            JsonObject object = ingredientJson.getAsJsonObject().deepCopy();
            object.addProperty("count", value.count);
            return object;
        }
        // Vanilla/Forge 1.20 represents compound ingredients as arrays. Preserve the
        // ingredient exactly and use the nested shape when an inline count has nowhere
        // to live; our decoder accepts both forms.
        JsonObject object = new JsonObject();
        object.add("ingredient", ingredientJson.deepCopy());
        object.addProperty("count", value.count);
        return object;
    }

    private static JsonElement encodeNested(SizedIngredient value) {
        JsonObject object = new JsonObject();
        object.add("ingredient", value.ingredient.toJson());
        object.addProperty("count", value.count);
        return object;
    }

    private static boolean isCompoundType(JsonObject object) {
        if (!object.has("type")) return false;
        String type = object.get("type").getAsString();
        return type.endsWith(":compound");
    }

    /**
     * 1.20 Forge accepts arrays for OR/compound ingredients. Convert only the
     * NeoForge release namespace inside custom ingredient type IDs; ordinary item/tag
     * JSON is left untouched.
     */
    private static JsonElement normalizeIngredientJson(JsonElement element) {
        if (element == null || element.isJsonNull()) return element;
        if (element.isJsonArray()) {
            JsonArray copy = new JsonArray();
            for (JsonElement child : element.getAsJsonArray()) {
                copy.add(normalizeIngredientJson(child));
            }
            return copy;
        }
        if (!element.isJsonObject()) return element.deepCopy();

        JsonObject object = element.getAsJsonObject().deepCopy();
        if (object.has("type")) {
            String type = object.get("type").getAsString();
            if (type.startsWith("neoforge:")) {
                object.addProperty("type", "forge:" + type.substring("neoforge:".length()));
            }
        }
        if (object.has("ingredients") && isCompoundType(object)) {
            return normalizeIngredientJson(object.get("ingredients"));
        }
        if (object.has("children")) {
            object.add("children", normalizeIngredientJson(object.get("children")));
        }
        return object;
    }

    @Override
    public boolean equals(Object other) {
        if (this == other) return true;
        if (!(other instanceof SizedIngredient that)) return false;
        return count == that.count && ingredient.equals(that.ingredient);
    }

    @Override
    public int hashCode() {
        return Objects.hash(ingredient, count);
    }

    @Override
    public String toString() {
        return count + "x " + ingredient;
    }
}
''', encoding='utf-8')

print(f'staged Forge 1.20 SizedIngredient owner: {target}')
