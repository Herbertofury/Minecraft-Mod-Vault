package net.fabric_extras.ranged_weapon.internal;

import net.fabric_extras.ranged_weapon.api.RangedConfig;
import net.minecraft.item.Item;
import org.jetbrains.annotations.Nullable;

/** 1.20.1 adapter for the 2.x Item.Settings ranged-attribute builder concept. */
public interface RangedItemSettings {
    @Nullable RangedConfig getRangedAttributes();
    Item.Settings rangedAttributes(RangedConfig config);
}
