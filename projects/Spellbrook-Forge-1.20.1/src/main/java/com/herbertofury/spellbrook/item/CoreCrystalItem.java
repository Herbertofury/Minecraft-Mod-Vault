package com.herbertofury.spellbrook.item;

import com.herbertofury.spellbrook.broom.CoreVariant;
import net.minecraft.nbt.CompoundTag;
import net.minecraft.network.chat.Component;
import net.minecraft.world.item.Item;
import net.minecraft.world.item.ItemStack;
import net.minecraft.world.item.TooltipFlag;
import net.minecraft.world.level.Level;

import javax.annotation.Nullable;
import java.util.List;

public class CoreCrystalItem extends Item {
    public static final String CORE_ID = "SpellbrookCoreVariant";
    public CoreCrystalItem(Properties properties) { super(properties); }
    public static CoreVariant getVariant(ItemStack stack) { String id=stack.getOrCreateTag().getString(CORE_ID); return id.isBlank()?CoreVariant.CRYSTAL:CoreVariant.byId(id); }
    public static void setVariant(ItemStack stack, CoreVariant variant) { CompoundTag tag=stack.getOrCreateTag(); tag.putString(CORE_ID, variant.id()); tag.putInt("CustomModelData", variant.customModelData()); }
    public static ItemStack stackFor(Item item, CoreVariant variant) { ItemStack stack=new ItemStack(item); setVariant(stack,variant); return stack; }
    @Override public void appendHoverText(ItemStack stack, @Nullable Level level, List<Component> tooltip, TooltipFlag flag) {
        CoreVariant v=getVariant(stack); tooltip.add(Component.literal("Core: "+v.id().replace('_',' '))); tooltip.add(Component.literal(String.format("Speed x%.2f | Handling x%.2f",v.speedMultiplier(),v.handlingMultiplier()))); super.appendHoverText(stack,level,tooltip,flag);
    }
}
