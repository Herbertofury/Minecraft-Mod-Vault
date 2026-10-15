package net.fabric_extras.ranged_weapon.api;

import net.minecraft.item.BowItem;
import net.minecraft.item.ItemStack;
import net.minecraft.recipe.Ingredient;
import java.util.HashSet;
import java.util.function.Supplier;

public class CustomBow extends BowItem {
    public static final HashSet<CustomBow> instances = new HashSet<>();
    private final Supplier<Ingredient> repairIngredientSupplier;
    public CustomBow(Settings settings, RangedConfig config, Supplier<Ingredient> repairIngredientSupplier) {
        super(settings);
        this.repairIngredientSupplier = repairIngredientSupplier;
        ((CustomRangedWeapon)(Object)this).setTypeBaseline(RangedConfig.BOW);
        ((CustomRangedWeapon)(Object)this).setRangedWeaponConfig(config);
        instances.add(this);
    }
    /** Legacy 1.20.1 constructor retained for downstream source compatibility. */
    public CustomBow(Settings settings, Supplier<Ingredient> repairIngredientSupplier) { this(settings, RangedConfig.BOW, repairIngredientSupplier); }
    @Deprecated public void config(RangedConfig config) { ((CustomRangedWeapon)(Object)this).setRangedWeaponConfig(config); }
    public void configure(RangedConfig config) { ((CustomRangedWeapon)(Object)this).setRangedWeaponConfig(config); }
    @Override public boolean canRepair(ItemStack stack, ItemStack ingredient) { return repairIngredientSupplier.get().test(ingredient) || super.canRepair(stack, ingredient); }
    public Supplier<Ingredient> getRepairIngredientSupplier() { return repairIngredientSupplier; }
}
