package net.fabric_extras.ranged_weapon.forge;

import net.fabric_extras.ranged_weapon.RangedWeaponMod;
import net.fabric_extras.ranged_weapon.client.TooltipUtil;
import net.minecraftforge.api.distmarker.Dist;
import net.minecraftforge.event.entity.player.ItemTooltipEvent;
import net.minecraftforge.eventbus.api.SubscribeEvent;
import net.minecraftforge.fml.common.Mod;

@Mod.EventBusSubscriber(modid=RangedWeaponMod.ID,value=Dist.CLIENT,bus=Mod.EventBusSubscriber.Bus.FORGE)
public final class ForgeClientEvents {
    @SubscribeEvent public static void tooltip(ItemTooltipEvent event) { TooltipUtil.addPullTime(event.getItemStack(),event.getToolTip()); }
    private ForgeClientEvents() {}
}
