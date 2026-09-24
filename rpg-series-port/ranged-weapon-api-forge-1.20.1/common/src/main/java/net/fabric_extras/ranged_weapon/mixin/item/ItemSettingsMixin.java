package net.fabric_extras.ranged_weapon.mixin.item;

import net.fabric_extras.ranged_weapon.api.RangedConfig;
import net.fabric_extras.ranged_weapon.internal.RangedItemSettings;
import net.minecraft.item.Item;
import org.jetbrains.annotations.Nullable;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.Unique;

@Mixin(Item.Settings.class)
public class ItemSettingsMixin implements RangedItemSettings {
    @Unique @Nullable private RangedConfig rwa$config;
    @Override public @Nullable RangedConfig getRangedAttributes() { return rwa$config; }
    @Override public Item.Settings rangedAttributes(RangedConfig config) { rwa$config=config; return (Item.Settings)(Object)this; }
}
