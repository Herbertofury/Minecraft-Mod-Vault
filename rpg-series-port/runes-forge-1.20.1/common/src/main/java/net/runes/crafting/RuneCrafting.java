package net.runes.crafting;

import net.minecraft.recipe.RecipeSerializer;
import net.minecraft.recipe.RecipeType;
import net.minecraft.util.Identifier;
import net.runes.RunesMod;

public final class RuneCrafting {
    public static final String NAME = "crafting";
    public static final Identifier ID = new Identifier(RunesMod.ID, NAME);
    public static final int SOUND_DELAY = 20;

    public static RecipeType<RuneCraftingRecipe> RECIPE_TYPE;
    public static RecipeSerializer<RuneCraftingRecipe> RECIPE_SERIALIZER;

    private RuneCrafting() { }

    /** Called only while Forge has the RECIPE_TYPE registry open. */
    public static void bootstrapType() {
        if (RECIPE_TYPE == null) {
            RECIPE_TYPE = new RecipeType<>() {
                @Override public String toString() { return NAME; }
            };
        }
    }

    /** Called only while Forge has the RECIPE_SERIALIZER registry open. */
    public static void bootstrapSerializer() {
        if (RECIPE_SERIALIZER == null) {
            RECIPE_SERIALIZER = new RuneCraftingRecipe.Serializer();
        }
    }
}
