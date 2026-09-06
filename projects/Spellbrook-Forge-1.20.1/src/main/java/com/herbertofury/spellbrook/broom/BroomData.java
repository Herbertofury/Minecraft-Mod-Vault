package com.herbertofury.spellbrook.broom;

import net.minecraft.nbt.CompoundTag;
import net.minecraft.world.item.ItemStack;

import java.util.UUID;

public final class BroomData {
    public static final String VARIANT = "SpellbrookVariant";
    public static final String CORE = "SpellbrookCore";
    public static final String INVENTORY = "Inventory";
    public static final String OWNER = "SpellbrookOwner";
    public static final String BROOM_UUID = "broomUUID";
    private BroomData() {}

    public static BroomVariant getVariant(ItemStack stack) { return BroomVariant.byId(stack.getOrCreateTag().getString(VARIANT)); }
    public static void setVariant(ItemStack stack, BroomVariant variant) { CompoundTag tag=stack.getOrCreateTag(); tag.putString(VARIANT, variant.id()); tag.putInt("CustomModelData", variant.customModelData()); }
    public static CoreVariant getCore(ItemStack stack) { String id=stack.getOrCreateTag().getString(CORE); return id.isBlank() ? CoreVariant.CRYSTAL : CoreVariant.byId(id); }
    public static void setCore(ItemStack stack, CoreVariant core) { stack.getOrCreateTag().putString(CORE, core.id()); }
    public static UUID getOrCreateBroomUuid(ItemStack stack) { CompoundTag tag=stack.getOrCreateTag(); if (!tag.hasUUID(BROOM_UUID)) tag.putUUID(BROOM_UUID, UUID.randomUUID()); return tag.getUUID(BROOM_UUID); }
}
