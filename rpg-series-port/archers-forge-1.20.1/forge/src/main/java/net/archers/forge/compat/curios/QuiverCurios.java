package net.archers.forge.compat.curios;

import net.archers.item.Quivers;
import net.minecraft.item.ItemStack;
import net.minecraft.sound.SoundEvents;
import top.theillusivec4.curios.api.CuriosApi;
import top.theillusivec4.curios.api.SlotContext;
import top.theillusivec4.curios.api.type.capability.ICurioItem;

/** Forge 1.20.1 equivalent of the current NeoForge quiver capability adapter. */
public final class QuiverCurios {
    private QuiverCurios() { }

    public static void register() {
        for (var entry : Quivers.entries) {
            CuriosApi.registerCurio(entry.item(), new QuiverCurio());
        }
    }

    private static final class QuiverCurio implements ICurioItem {
        @Override
        public void onEquip(SlotContext slotContext, ItemStack prevStack, ItemStack stack) {
            var entity = slotContext.entity();
            if (entity == null) return;
            var world = entity.getWorld();
            if (world.isClient()
                    || entity.age <= 100
                    || prevStack.isOf(stack.getItem())) {
                return;
            }
            world.playSound(null, entity.getBlockPos(), SoundEvents.ITEM_ARMOR_EQUIP_GENERIC,
                    entity.getSoundCategory(), 1.0F, 1.0F);
        }

        @Override
        public void onEquipFromUse(SlotContext slotContext, ItemStack stack) {
            // onEquip above covers every equip path; keep this silent to prevent a duplicate sound.
        }
    }
}
