package net.fabric_extras.ranged_weapon.api;

import net.minecraft.item.CrossbowItem;
import net.minecraft.item.ItemStack;
import net.minecraft.recipe.Ingredient;
import java.util.HashSet;
import java.util.function.Supplier;

public class CustomCrossbow extends CrossbowItem {
    public static final HashSet<CustomCrossbow> instances = new HashSet<>();
    private final Supplier<Ingredient> repairIngredientSupplier;
    public CustomCrossbow(Settings settings, RangedConfig config, Supplier<Ingredient> repairIngredientSupplier) {
        super(settings);
        this.repairIngredientSupplier = repairIngredientSupplier;
        ((CustomRangedWeapon)(Object)this).setTypeBaseline(RangedConfig.CROSSBOW);
        ((CustomRangedWeapon)(Object)this).setRangedWeaponConfig(config);
        instances.add(this);
    }
    public CustomCrossbow(Settings settings, Supplier<Ingredient> repairIngredientSupplier) { this(settings, RangedConfig.CROSSBOW, repairIngredientSupplier); }
    @Deprecated public void config(RangedConfig config) { ((CustomRangedWeapon)(Object)this).setRangedWeaponConfig(config); }
    public void configure(RangedConfig config) { ((CustomRangedWeapon)(Object)this).setRangedWeaponConfig(config); }
    @Override public boolean canRepair(ItemStack stack, ItemStack ingredient) { return repairIngredientSupplier.get().test(ingredient) || super.canRepair(stack, ingredient); }
    public Supplier<Ingredient> getRepairIngredientSupplier() { return repairIngredientSupplier; }
}
