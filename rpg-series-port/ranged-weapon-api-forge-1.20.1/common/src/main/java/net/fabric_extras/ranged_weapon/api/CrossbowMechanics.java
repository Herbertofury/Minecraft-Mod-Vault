package net.fabric_extras.ranged_weapon.api;

import net.minecraft.enchantment.EnchantmentHelper;
import net.minecraft.enchantment.Enchantments;
import net.minecraft.entity.LivingEntity;
import net.minecraft.item.ItemStack;
import org.jetbrains.annotations.Nullable;

/** 2.3.x provider shape adapted to the 1.20.1 Quick Charge API. */
public final class CrossbowMechanics {
    public static final class PullTime {
        public static final Provider defaultProvider = (originalPullTime, crossbow, user) -> {
            int quickCharge = EnchantmentHelper.getLevel(Enchantments.QUICK_CHARGE, crossbow);
            return Math.max(1, originalPullTime - (int)(originalPullTime * 0.2F) * quickCharge);
        };
        public static Provider modifier = defaultProvider;
        @FunctionalInterface public interface Provider {
            int getPullTime(int originalPullTime, ItemStack crossbow, @Nullable LivingEntity user);
        }
        private PullTime() {}
    }
    private CrossbowMechanics() {}
}
