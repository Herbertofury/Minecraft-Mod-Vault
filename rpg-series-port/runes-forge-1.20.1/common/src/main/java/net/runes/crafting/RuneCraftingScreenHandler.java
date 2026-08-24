package net.runes.crafting;

import net.minecraft.block.BlockState;
import net.minecraft.entity.player.PlayerEntity;
import net.minecraft.entity.player.PlayerInventory;
import net.minecraft.item.ItemStack;
import net.minecraft.network.PacketByteBuf;
import net.minecraft.screen.ForgingScreenHandler;
import net.minecraft.screen.ScreenHandlerContext;
import net.minecraft.screen.ScreenHandlerType;
import net.minecraft.screen.slot.ForgingSlotsManager;
import net.minecraft.screen.slot.Slot;
import net.minecraft.server.network.ServerPlayerEntity;
import net.minecraft.sound.SoundCategory;
import net.minecraft.world.World;
import org.jetbrains.annotations.Nullable;

import java.util.List;
import java.util.Objects;

public class RuneCraftingScreenHandler extends ForgingScreenHandler {
    private static ScreenHandlerType<RuneCraftingScreenHandler> HANDLER_TYPE;
    private final World world;
    @Nullable private RuneCraftingRecipe currentRecipe;
    private final List<RuneCraftingRecipe> recipes;

    public static void bindHandlerType(ScreenHandlerType<RuneCraftingScreenHandler> type) {
        Objects.requireNonNull(type, "type");
        if (HANDLER_TYPE != null && HANDLER_TYPE != type) {
            throw new IllegalStateException("Rune crafting screen handler type is already bound");
        }
        HANDLER_TYPE = type;
    }

    public static ScreenHandlerType<RuneCraftingScreenHandler> handlerType() {
        if (HANDLER_TYPE == null) {
            throw new IllegalStateException("Rune crafting screen handler type has not been bound by the active loader");
        }
        return HANDLER_TYPE;
    }

    public RuneCraftingScreenHandler(int syncId,PlayerInventory inventory){ this(syncId,inventory,ScreenHandlerContext.EMPTY); }
    public RuneCraftingScreenHandler(int syncId,PlayerInventory inventory,PacketByteBuf buf){ this(syncId,inventory,ScreenHandlerContext.EMPTY); }
    public RuneCraftingScreenHandler(int syncId,PlayerInventory inventory,ScreenHandlerContext context){
        super(handlerType(),syncId,inventory,context); this.world=inventory.player.getWorld();
        this.recipes=this.world.getRecipeManager().listAllOfType(RuneCrafting.RECIPE_TYPE);
    }
    protected ForgingSlotsManager getForgingSlotsManager(){ return ForgingSlotsManager.create().input(0,27,47,s->true).input(1,76,47,s->true).output(2,134,47).build(); }
    protected boolean canUse(BlockState state){ return state.isOf(RuneCraftingBlock.INSTANCE); }
    protected boolean canTakeOutput(PlayerEntity player,boolean present){ return currentRecipe!=null && currentRecipe.matches(input,world); }
    protected void onTakeOutput(PlayerEntity player,ItemStack stack){
        stack.onCraft(player.getWorld(),player,stack.getCount()); output.unlockLastRecipe(player,getInputStacks()); decrementStack(0); decrementStack(1);
        if(player instanceof ServerPlayerEntity serverPlayer) RuneCraftingCriteria.INSTANCE.trigger(serverPlayer);
        var crafter=(RuneCrafter)player;
        if(crafter.shouldPlayRuneCraftingSound(player.age)){
            world.playSound(player.getX(),player.getY(),player.getZ(),RunesModSound.SOUND,SoundCategory.BLOCKS,world.random.nextFloat()*0.1F+0.9F,1,true);
            crafter.onPlayedRuneCraftingSound(player.age);
        }
    }
    private List<ItemStack> getInputStacks(){ return List.of(input.getStack(0),input.getStack(1)); }
    private void decrementStack(int slot){ ItemStack s=input.getStack(slot); s.decrement(1); input.setStack(slot,s); }
    public void updateResult(){
        List<RuneCraftingRecipe> list=world.getRecipeManager().getAllMatches(RuneCrafting.RECIPE_TYPE,input,world);
        if(list.isEmpty()){ currentRecipe=null; output.setStack(0,ItemStack.EMPTY); }
        else { currentRecipe=list.get(0); ItemStack result=currentRecipe.craft(input,world.getRegistryManager()); output.setLastRecipe(currentRecipe); output.setStack(0,result); }
    }
    protected boolean isUsableAsAddition(ItemStack stack){ return recipes.stream().anyMatch(r->r.testAddition(stack)); }
    public boolean canInsertIntoSlot(ItemStack stack,Slot slot){ return slot.inventory!=output && super.canInsertIntoSlot(stack,slot); }
    private static final class RunesModSound { private static final net.minecraft.sound.SoundEvent SOUND=net.runes.RunesMod.CRAFTING_SOUND; }
}
