package com.herbertofury.spellbrook.entity;

import com.herbertofury.spellbrook.broom.BroomData;
import com.herbertofury.spellbrook.broom.BroomVariant;
import com.herbertofury.spellbrook.broom.CoreVariant;
import com.herbertofury.spellbrook.item.CoreCrystalItem;
import com.herbertofury.spellbrook.registry.ModItems;
import com.herbertofury.spellbrook.registry.ModSounds;
import net.minecraft.core.Direction;
import net.minecraft.nbt.CompoundTag;
import net.minecraft.network.chat.Component;
import net.minecraft.network.protocol.Packet;
import net.minecraft.network.protocol.game.ClientGamePacketListener;
import net.minecraft.network.syncher.EntityDataAccessor;
import net.minecraft.network.syncher.EntityDataSerializers;
import net.minecraft.network.syncher.SynchedEntityData;
import net.minecraft.server.level.ServerPlayer;
import net.minecraft.sounds.SoundSource;
import net.minecraft.util.Mth;
import net.minecraft.world.Container;
import net.minecraft.world.InteractionHand;
import net.minecraft.world.InteractionResult;
import net.minecraft.world.SimpleMenuProvider;
import net.minecraft.world.damagesource.DamageSource;
import net.minecraft.world.entity.Entity;
import net.minecraft.world.entity.EntityType;
import net.minecraft.world.entity.MoverType;
import net.minecraft.world.entity.player.Player;
import net.minecraft.world.inventory.ChestMenu;
import net.minecraft.world.item.ItemStack;
import net.minecraft.world.level.Level;
import net.minecraft.world.phys.Vec3;
import net.minecraftforge.common.capabilities.Capability;
import net.minecraftforge.common.capabilities.ForgeCapabilities;
import net.minecraftforge.common.util.LazyOptional;
import net.minecraftforge.items.IItemHandler;
import net.minecraftforge.items.ItemStackHandler;
import net.minecraftforge.network.NetworkHooks;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

import java.util.UUID;

public class SpellbrookBroomEntity extends Entity {
    private static final EntityDataAccessor<String> VARIANT = SynchedEntityData.defineId(SpellbrookBroomEntity.class, EntityDataSerializers.STRING);
    private static final EntityDataAccessor<String> CORE = SynchedEntityData.defineId(SpellbrookBroomEntity.class, EntityDataSerializers.STRING);
    private static final EntityDataAccessor<Float> BANK = SynchedEntityData.defineId(SpellbrookBroomEntity.class, EntityDataSerializers.FLOAT);
    private static final EntityDataAccessor<Float> VISUAL_PITCH = SynchedEntityData.defineId(SpellbrookBroomEntity.class, EntityDataSerializers.FLOAT);
    private static final EntityDataAccessor<Boolean> BOOSTING = SynchedEntityData.defineId(SpellbrookBroomEntity.class, EntityDataSerializers.BOOLEAN);

    private final ItemStackHandler itemHandler = new ItemStackHandler(30) { @Override protected void onContentsChanged(int slot) { setChangedFlag = true; } };
    private LazyOptional<IItemHandler> itemCap = LazyOptional.of(() -> itemHandler);
    private UUID owner;
    private UUID broomUuid = UUID.randomUUID();
    private boolean setChangedFlag;
    private float damage;
    private boolean forward, back, left, right, ascend, descend, boost;
    private boolean boostSoundLatched;

    public SpellbrookBroomEntity(EntityType<? extends SpellbrookBroomEntity> type, Level level) {
        super(type, level);
        setNoGravity(true);
    }

    @Override protected void defineSynchedData() {
        entityData.define(VARIANT, BroomVariant.DEFAULT.id());
        entityData.define(CORE, CoreVariant.CRYSTAL.id());
        entityData.define(BANK, 0f);
        entityData.define(VISUAL_PITCH, 0f);
        entityData.define(BOOSTING, false);
    }

    public BroomVariant getVariant() { return BroomVariant.byId(entityData.get(VARIANT)); }
    public CoreVariant getCore() { return CoreVariant.byId(entityData.get(CORE)); }
    public float getBank(float partialTick) { return entityData.get(BANK); }
    public float getVisualPitch(float partialTick) { return entityData.get(VISUAL_PITCH); }
    public boolean isBoosting() { return entityData.get(BOOSTING); }

    public void setInputs(boolean forward, boolean back, boolean left, boolean right, boolean ascend, boolean descend, boolean boost) {
        this.forward=forward; this.back=back; this.left=left; this.right=right; this.ascend=ascend; this.descend=descend; this.boost=boost;
    }

    public void loadFromItem(ItemStack stack, @Nullable Player player) {
        BroomVariant variant=BroomData.getVariant(stack); CoreVariant core=BroomData.getCore(stack);
        entityData.set(VARIANT, variant.id()); entityData.set(CORE, core.id());
        CompoundTag tag=stack.getTag();
        if(tag!=null && tag.contains(BroomData.INVENTORY)) itemHandler.deserializeNBT(tag.getCompound(BroomData.INVENTORY));
        if(tag!=null && tag.hasUUID(BroomData.OWNER)) owner=tag.getUUID(BroomData.OWNER); else if(player!=null) owner=player.getUUID();
        if(tag!=null && tag.hasUUID(BroomData.BROOM_UUID)) broomUuid=tag.getUUID(BroomData.BROOM_UUID);
        if(stack.hasCustomHoverName()) setCustomName(stack.getHoverName());
    }

    public ItemStack getRenderStack() {
        ItemStack stack=new ItemStack(ModItems.BROOMSTICK.get()); BroomData.setVariant(stack,getVariant()); BroomData.setCore(stack,getCore()); return stack;
    }

    @Override public @NotNull ItemStack getPickResult() {
        ItemStack stack=getRenderStack(); CompoundTag tag=stack.getOrCreateTag(); tag.put(BroomData.INVENTORY,itemHandler.serializeNBT()); tag.putUUID(BroomData.BROOM_UUID,broomUuid); if(owner!=null) tag.putUUID(BroomData.OWNER,owner); if(hasCustomName()) stack.setHoverName(getCustomName()); return stack;
    }

    @Override public void tick() {
        super.tick(); fallDistance=0;
        if(level().isClientSide) return;
        Entity rider=getControllingPassenger();
        if(rider instanceof Player player) {
            BroomVariant variant=getVariant(); CoreVariant core=getCore();
            float handling=variant.handling()*core.handlingMultiplier();
            float yawDelta=Mth.wrapDegrees(player.getYRot()-getYRot()); setYRot(getYRot()+Mth.clamp(yawDelta,-5.5f*handling,5.5f*handling));
            Vec3 fwd=Vec3.directionFromRotation(0,getYRot()).multiply(1,0,1).normalize(); Vec3 side=new Vec3(-fwd.z,0,fwd.x);
            double forwardAxis=(forward?1:0)-(back?0.55:0); double sideAxis=(right?0.70:0)-(left?0.70:0);
            Vec3 wish=fwd.scale(forwardAxis).add(side.scale(sideAxis)); if(wish.lengthSqr()>1) wish=wish.normalize();
            double topSpeed=0.43*variant.speed()*core.speedMultiplier()*(boost?1.52:1.0); double response=Mth.clamp(0.075*handling,0.035,0.18);
            Vec3 old=getDeltaMovement(); double targetX=wish.x*topSpeed, targetZ=wish.z*topSpeed;
            double vx=Mth.lerp(response,old.x,targetX); double vz=Mth.lerp(response,old.z,targetZ);
            double targetY=ascend?0.34*handling:(descend?-0.30:0.0); double vy=Mth.lerp(0.16,old.y,targetY);
            setDeltaMovement(vx,vy,vz); move(MoverType.SELF,getDeltaMovement()); setDeltaMovement(getDeltaMovement().multiply(0.985,0.94,0.985));
            float bankTarget=(left?18f:0f)+(right?-18f:0f); entityData.set(BANK,Mth.lerp(0.22f,entityData.get(BANK),bankTarget));
            float pitchTarget=ascend?-8f:(descend?10f:(forward?-3f:0f)); entityData.set(VISUAL_PITCH,Mth.lerp(0.18f,entityData.get(VISUAL_PITCH),pitchTarget));
            entityData.set(BOOSTING,boost);
            if(boost && !boostSoundLatched){ level().playSound(null,blockPosition(),ModSounds.BOOST.get(),SoundSource.PLAYERS,0.8f,1f); boostSoundLatched=true; }
            if(!boost) boostSoundLatched=false;
        } else {
            Vec3 old=getDeltaMovement(); setDeltaMovement(old.x*0.94,Math.max(old.y-0.035,-0.24),old.z*0.94); move(MoverType.SELF,getDeltaMovement()); entityData.set(BANK,Mth.lerp(0.18f,entityData.get(BANK),0)); entityData.set(VISUAL_PITCH,Mth.lerp(0.18f,entityData.get(VISUAL_PITCH),0)); entityData.set(BOOSTING,false);
        }
    }

    @Override public InteractionResult interact(Player player, InteractionHand hand) {
        ItemStack held=player.getItemInHand(hand);
        if(held.getItem()==ModItems.CORE_CRYSTAL.get()) {
            if(!level().isClientSide) {
                CoreVariant previous=getCore(); CoreVariant next=CoreCrystalItem.getVariant(held); entityData.set(CORE,next.id());
                if(!player.getAbilities().instabuild) held.shrink(1);
                ItemStack returned=CoreCrystalItem.stackFor(ModItems.CORE_CRYSTAL.get(),previous); if(!player.getInventory().add(returned)) player.drop(returned,false);
            }
            return InteractionResult.sidedSuccess(level().isClientSide);
        }
        if(player.isSecondaryUseActive()) {
            if(!level().isClientSide && player instanceof ServerPlayer serverPlayer) serverPlayer.openMenu(new SimpleMenuProvider((id,inv,p)->ChestMenu.threeRows(id,inv,new StorageView(this)),Component.translatable("container.spellbrook.broom_storage")));
            return InteractionResult.sidedSuccess(level().isClientSide);
        }
        if(!level().isClientSide && !player.isPassenger()) { if(owner==null) owner=player.getUUID(); player.startRiding(this); level().playSound(null,blockPosition(),ModSounds.MOUNT.get(),SoundSource.PLAYERS,0.7f,1f); }
        return InteractionResult.sidedSuccess(level().isClientSide);
    }

    @Override protected boolean canAddPassenger(Entity passenger) { return getPassengers().isEmpty(); }
    @Override protected void positionRider(Entity passenger, MoveFunction moveFunction) { if(hasPassenger(passenger)) { Vec3 seat=new Vec3(0,0.45,0).yRot(-getYRot()*Mth.DEG_TO_RAD); moveFunction.accept(passenger,getX()+seat.x,getY()+seat.y,getZ()+seat.z); passenger.setYRot(getYRot()); } }
    @Override public double getPassengersRidingOffset() { return 0.2; }
    @Override public boolean isPickable() { return true; }
    @Override public boolean isPushable() { return true; }

    @Override public boolean hurt(DamageSource source,float amount) {
        if(isInvulnerableTo(source)) return false; if(level().isClientSide) return true; damage+=amount*10f; markHurt(); boolean creative=source.getEntity() instanceof Player p && p.getAbilities().instabuild;
        if(creative || damage>50f){ if(!creative) spawnAtLocation(getPickResult()); discard(); } return true;
    }

    @Override protected void addAdditionalSaveData(CompoundTag tag) { tag.putString(BroomData.VARIANT,getVariant().id()); tag.putString(BroomData.CORE,getCore().id()); tag.put(BroomData.INVENTORY,itemHandler.serializeNBT()); tag.putUUID(BroomData.BROOM_UUID,broomUuid); if(owner!=null) tag.putUUID(BroomData.OWNER,owner); }
    @Override protected void readAdditionalSaveData(CompoundTag tag) { entityData.set(VARIANT,tag.getString(BroomData.VARIANT).isBlank()?BroomVariant.DEFAULT.id():tag.getString(BroomData.VARIANT)); entityData.set(CORE,tag.getString(BroomData.CORE).isBlank()?CoreVariant.CRYSTAL.id():tag.getString(BroomData.CORE)); if(tag.contains(BroomData.INVENTORY)) itemHandler.deserializeNBT(tag.getCompound(BroomData.INVENTORY)); if(tag.hasUUID(BroomData.BROOM_UUID)) broomUuid=tag.getUUID(BroomData.BROOM_UUID); if(tag.hasUUID(BroomData.OWNER)) owner=tag.getUUID(BroomData.OWNER); }
    @Override public Packet<ClientGamePacketListener> getAddEntityPacket() { return NetworkHooks.getEntitySpawningPacket(this); }
    @Override public <T> LazyOptional<T> getCapability(Capability<T> cap,@Nullable Direction side) { if(cap==ForgeCapabilities.ITEM_HANDLER && isAlive()) return itemCap.cast(); return super.getCapability(cap,side); }
    @Override public void invalidateCaps() { super.invalidateCaps(); itemCap.invalidate(); }

    private static final class StorageView implements Container {
        private final SpellbrookBroomEntity broom; StorageView(SpellbrookBroomEntity broom){this.broom=broom;}
        private int map(int slot){return slot+3;} public int getContainerSize(){return 27;} public boolean isEmpty(){for(int i=0;i<27;i++)if(!getItem(i).isEmpty())return false;return true;}
        public ItemStack getItem(int slot){return broom.itemHandler.getStackInSlot(map(slot));} public ItemStack removeItem(int slot,int count){return broom.itemHandler.extractItem(map(slot),count,false);} public ItemStack removeItemNoUpdate(int slot){ItemStack s=getItem(slot); broom.itemHandler.setStackInSlot(map(slot),ItemStack.EMPTY); return s;}
        public void setItem(int slot,ItemStack stack){broom.itemHandler.setStackInSlot(map(slot),stack);} public void setChanged(){broom.setChangedFlag=true;} public boolean stillValid(Player player){return broom.isAlive()&&player.distanceToSqr(broom)<64;}
        public void clearContent(){for(int i=0;i<27;i++)broom.itemHandler.setStackInSlot(map(i),ItemStack.EMPTY);}
    }
}
