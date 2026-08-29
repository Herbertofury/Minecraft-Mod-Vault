package net.fabric_extras.ranged_weapon.client;

import net.fabric_extras.ranged_weapon.api.CustomRangedWeapon;
import net.minecraft.item.ItemStack;
import net.minecraft.text.Text;
import net.minecraft.util.Formatting;
import java.util.List;
import java.util.Locale;

public final class TooltipUtil {
    public static void addPullTime(ItemStack stack, List<Text> lines) {
        if (!(stack.getItem() instanceof CustomRangedWeapon weapon)) return;
        var seconds = weapon.getRangedWeaponConfig().pullTimeSeconds();
        if (seconds <= 0) return;
        lines.add(Text.literal(" ").append(Text.translatable("item.ranged_weapon.pull_time", String.format(Locale.ROOT, "%.1f", seconds)).formatted(Formatting.DARK_GREEN)));
    }
    private TooltipUtil() {}
}
