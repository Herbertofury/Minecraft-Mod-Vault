package com.herbertofury.spellbrook.item;

import com.herbertofury.spellbrook.broom.BroomData;
import com.herbertofury.spellbrook.broom.BroomVariant;
import com.herbertofury.spellbrook.compat.CompatHooks;
import com.herbertofury.spellbrook.entity.SpellbrookBroomEntity;
import com.herbertofury.spellbrook.registry.ModEntities;
import net.minecraft.core.BlockPos;
import net.minecraft.network.chat.Component;
import net.minecraft.world.InteractionResult;
import net.minecraft.world.entity.player.Player;
import net.minecraft.world.item.Item;
import net.minecraft.world.item.ItemStack;
import net.minecraft.world.item.TooltipFlag;
import net.minecraft.world.item.context.UseOnContext;
import net.minecraft.world.level.Level;

import javax.annotation.Nullable;
import java.util.List;

public class SpellbrookBroomItem extends Item {
    public SpellbrookBroomItem(Properties properties) { super(properties); }
    public static ItemStack stackFor(Item item, BroomVariant variant) { ItemStack stack=new ItemStack(item); BroomData.setVariant(stack,variant); BroomData.getOrCreateBroomUuid(stack); return stack; }

    @Override public InteractionResult useOn(UseOnContext context) {
        Level level=context.getLevel(); Player player=context.getPlayer(); ItemStack stack=context.getItemInHand();
        BlockPos pos=context.getClickedPos().relative(context.getClickedFace());
        if (level.isClientSide) return InteractionResult.SUCCESS;
        float yaw=player==null?0f:player.getYRot();
        if (CompatHooks.trySpawnHexerei(level,pos,stack,player,yaw)) { if (player==null || !player.getAbilities().instabuild) stack.shrink(1); return InteractionResult.CONSUME; }
        SpellbrookBroomEntity broom=ModEntities.BROOM.get().create(level);
        if (broom==null) return InteractionResult.FAIL;
        broom.absMoveTo(pos.getX()+0.5,pos.getY()+0.1,pos.getZ()+0.5,yaw,0f);
        broom.loadFromItem(stack,player);
        if (!level.noCollision(broom,broom.getBoundingBox())) return InteractionResult.FAIL;
        level.addFreshEntity(broom);
        if (player==null || !player.getAbilities().instabuild) stack.shrink(1);
        return InteractionResult.CONSUME;
    }

    @Override public void appendHoverText(ItemStack stack, @Nullable Level level, List<Component> tooltip, TooltipFlag flag) {
        BroomVariant v=BroomData.getVariant(stack); tooltip.add(Component.literal("Finish: "+v.id().replace('_',' '))); tooltip.add(Component.literal("Core: "+BroomData.getCore(stack).id().replace('_',' '))); tooltip.add(Component.literal("Place to ride. Sneak-right-click the placed broom for storage.")); super.appendHoverText(stack,level,tooltip,flag);
    }
}
