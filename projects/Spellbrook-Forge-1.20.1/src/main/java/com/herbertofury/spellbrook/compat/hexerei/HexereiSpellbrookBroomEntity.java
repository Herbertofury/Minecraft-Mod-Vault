package com.herbertofury.spellbrook.compat.hexerei;

import com.herbertofury.spellbrook.broom.BroomData;
import com.herbertofury.spellbrook.broom.BroomVariant;
import com.herbertofury.spellbrook.broom.CoreVariant;
import com.herbertofury.spellbrook.item.CoreCrystalItem;
import com.herbertofury.spellbrook.registry.ModItems;
import net.joefoxe.hexerei.client.renderer.entity.custom.BroomEntity;
import net.minecraft.nbt.CompoundTag;
import net.minecraft.network.syncher.EntityDataAccessor;
import net.minecraft.network.syncher.EntityDataSerializers;
import net.minecraft.network.syncher.SynchedEntityData;
import net.minecraft.world.InteractionHand;
import net.minecraft.world.InteractionResult;
import net.minecraft.world.entity.EntityType;
import net.minecraft.world.entity.player.Player;
import net.minecraft.world.item.ItemStack;
import net.minecraft.world.level.Level;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class HexereiSpellbrookBroomEntity extends BroomEntity {
    private static final EntityDataAccessor<String> SPELLBROOK_VARIANT=SynchedEntityData.defineId(HexereiSpellbrookBroomEntity.class, EntityDataSerializers.STRING);
    private static final EntityDataAccessor<String> SPELLBROOK_CORE=SynchedEntityData.defineId(HexereiSpellbrookBroomEntity.class, EntityDataSerializers.STRING);

    @SuppressWarnings({"unchecked","rawtypes"}) public HexereiSpellbrookBroomEntity(EntityType<? extends HexereiSpellbrookBroomEntity> type,Level level){super((EntityType)type,level);}
    @Override protected void defineSynchedData(){super.defineSynchedData();entityData.define(SPELLBROOK_VARIANT,BroomVariant.DEFAULT.id());entityData.define(SPELLBROOK_CORE,CoreVariant.CRYSTAL.id());}
    public BroomVariant getSpellbrookVariant(){return BroomVariant.byId(entityData.get(SPELLBROOK_VARIANT));} public CoreVariant getSpellbrookCore(){return CoreVariant.byId(entityData.get(SPELLBROOK_CORE));}
    public void loadFromSpellbrookItem(ItemStack stack,@Nullable Player player){BroomVariant v=BroomData.getVariant(stack);CoreVariant c=BroomData.getCore(stack);entityData.set(SPELLBROOK_VARIANT,v.id());entityData.set(SPELLBROOK_CORE,c.id());setBroomType("spellbrook_"+v.id());speedMultiplier=0.65f*v.speed()*c.speedMultiplier();CompoundTag tag=stack.getTag();if(tag!=null&&tag.contains(BroomData.INVENTORY))itemHandler.deserializeNBT(tag.getCompound(BroomData.INVENTORY));if(itemHandler.getStackInSlot(BroomSlot.BRUSH.ordinal()).isEmpty())itemHandler.setStackInSlot(BroomSlot.BRUSH.ordinal(),new ItemStack(net.joefoxe.hexerei.item.ModItems.BROOM_BRUSH.get()));if(tag!=null&&tag.hasUUID(BroomData.BROOM_UUID))broomUUID=tag.getUUID(BroomData.BROOM_UUID);else broomUUID=BroomData.getOrCreateBroomUuid(stack);if(stack.hasCustomHoverName())setCustomName(stack.getHoverName());}
    public ItemStack getRenderStack(){ItemStack stack=new ItemStack(ModItems.BROOMSTICK.get());BroomData.setVariant(stack,getSpellbrookVariant());BroomData.setCore(stack,getSpellbrookCore());return stack;}
    @Override public @NotNull ItemStack getPickResult(){ItemStack stack=super.getPickResult();BroomData.setVariant(stack,getSpellbrookVariant());BroomData.setCore(stack,getSpellbrookCore());return stack;}
    @Override public InteractionResult interact(Player player, InteractionHand hand){ItemStack held=player.getItemInHand(hand);if(held.getItem()==ModItems.CORE_CRYSTAL.get()){if(!level().isClientSide){CoreVariant previous=getSpellbrookCore(),next=CoreCrystalItem.getVariant(held);entityData.set(SPELLBROOK_CORE,next.id());speedMultiplier=0.65f*getSpellbrookVariant().speed()*next.speedMultiplier();if(!player.getAbilities().instabuild)held.shrink(1);ItemStack returned=CoreCrystalItem.stackFor(ModItems.CORE_CRYSTAL.get(),previous);if(!player.getInventory().add(returned))player.drop(returned,false);}return InteractionResult.sidedSuccess(level().isClientSide);}return super.interact(player,hand);}
    @Override protected void addAdditionalSaveData(CompoundTag tag){super.addAdditionalSaveData(tag);tag.putString(BroomData.VARIANT,getSpellbrookVariant().id());tag.putString(BroomData.CORE,getSpellbrookCore().id());}
    @Override protected void readAdditionalSaveData(CompoundTag tag){super.readAdditionalSaveData(tag);BroomVariant v=BroomVariant.byId(tag.getString(BroomData.VARIANT));CoreVariant c=CoreVariant.byId(tag.getString(BroomData.CORE));entityData.set(SPELLBROOK_VARIANT,v.id());entityData.set(SPELLBROOK_CORE,c.id());setBroomType("spellbrook_"+v.id());speedMultiplier=0.65f*v.speed()*c.speedMultiplier();}
}
