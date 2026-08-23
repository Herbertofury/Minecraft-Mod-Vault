package net.runes.compat.emi;

import dev.emi.emi.api.EmiEntrypoint;
import dev.emi.emi.api.EmiPlugin;
import dev.emi.emi.api.EmiRegistry;
import dev.emi.emi.api.recipe.EmiRecipeCategory;
import dev.emi.emi.api.recipe.EmiRecipeSorting;
import dev.emi.emi.api.stack.EmiStack;
import net.fabricmc.api.EnvType;
import net.fabricmc.api.Environment;
import net.runes.crafting.RuneCrafting;
import net.runes.crafting.RuneCraftingBlock;
import net.runes.crafting.RuneCraftingRecipe;

/** Optional EMI integration for the altar-only custom recipe type. */
@EmiEntrypoint
@Environment(EnvType.CLIENT)
public final class RunesEmiPlugin implements EmiPlugin {
    public static final EmiStack ALTAR = EmiStack.of(RuneCraftingBlock.ITEM);
    public static final EmiRecipeCategory CATEGORY = new EmiRecipeCategory(
            RuneCrafting.ID, ALTAR, ALTAR, EmiRecipeSorting.compareOutputThenInput());

    @Override
    public void register(EmiRegistry registry) {
        registry.addCategory(CATEGORY);
        registry.addWorkstation(CATEGORY, ALTAR);
        for (RuneCraftingRecipe recipe : registry.getRecipeManager().listAllOfType(RuneCrafting.RECIPE_TYPE)) {
            registry.addRecipe(new RuneCraftingEmiRecipe(recipe));
        }
    }
}
