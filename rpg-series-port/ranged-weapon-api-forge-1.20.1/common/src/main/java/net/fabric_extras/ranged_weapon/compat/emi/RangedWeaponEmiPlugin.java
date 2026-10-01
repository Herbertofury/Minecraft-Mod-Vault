package net.fabric_extras.ranged_weapon.compat.emi;

import dev.emi.emi.api.EmiEntrypoint;
import dev.emi.emi.api.EmiPlugin;
import dev.emi.emi.api.EmiRegistry;
import dev.emi.emi.api.stack.EmiIngredient;
import dev.emi.emi.api.stack.EmiStack;
import dev.emi.emi.recipe.EmiAnvilRecipe;
import net.fabric_extras.ranged_weapon.api.CustomBow;
import net.fabric_extras.ranged_weapon.api.CustomCrossbow;
import net.minecraft.item.Item;
import net.minecraft.recipe.Ingredient;
import net.minecraft.registry.Registries;
import net.minecraft.util.Identifier;

/** Optional 2.3.2 EMI repair integration; EMI remains compile-only and non-required at runtime. */
@EmiEntrypoint
public final class RangedWeaponEmiPlugin implements EmiPlugin {
    @Override
    public void register(EmiRegistry registry) {
        for (var bow : CustomBow.instances) registerAnvilRecipe(registry, bow, bow.getRepairIngredientSupplier().get());
        for (var crossbow : CustomCrossbow.instances) registerAnvilRecipe(registry, crossbow, crossbow.getRepairIngredientSupplier().get());
    }

    private static void registerAnvilRecipe(EmiRegistry registry, Item item, Ingredient repairIngredient) {
        Identifier itemId = Registries.ITEM.getId(item);
        if (itemId == null || itemId.equals(Registries.ITEM.getDefaultId())) return;
        Identifier recipeId = new Identifier(itemId.getNamespace(), "anvil_repair_rwa/" + itemId.getPath());
        registry.addRecipe(new EmiAnvilRecipe(EmiStack.of(item), EmiIngredient.of(repairIngredient), recipeId));
    }
}
