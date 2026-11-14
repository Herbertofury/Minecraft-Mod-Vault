package net.runes.crafting;

import com.google.gson.JsonObject;
import net.minecraft.inventory.Inventory;
import net.minecraft.item.ItemStack;
import net.minecraft.nbt.NbtCompound;
import net.minecraft.network.PacketByteBuf;
import net.minecraft.recipe.Ingredient;
import net.minecraft.recipe.Recipe;
import net.minecraft.recipe.RecipeSerializer;
import net.minecraft.recipe.RecipeType;
import net.minecraft.recipe.ShapedRecipe;
import net.minecraft.registry.DynamicRegistryManager;
import net.minecraft.util.Identifier;
import net.minecraft.util.JsonHelper;
import net.minecraft.world.World;

import java.util.stream.Stream;

public class RuneCraftingRecipe implements Recipe<Inventory> {
    final Ingredient base;
    final Ingredient addition;
    final ItemStack result;
    private final Identifier id;
    public RuneCraftingRecipe(Identifier id, Ingredient base, Ingredient addition, ItemStack result) {
        this.id=id; this.base=base; this.addition=addition; this.result=result;
    }
    public Ingredient base(){ return base; }
    public Ingredient addition(){ return addition; }
    public ItemStack result(){ return result; }
    public boolean matches(Inventory inv, World world){ return base.test(inv.getStack(0)) && addition.test(inv.getStack(1)); }
    public ItemStack craft(Inventory inv, DynamicRegistryManager registryManager){
        ItemStack out=result.copy();
        NbtCompound nbt=inv.getStack(0).getNbt();
        if(nbt!=null) out.setNbt(nbt.copy());
        return out;
    }
    public boolean fits(int width,int height){ return width*height>=2; }
    public ItemStack getOutput(DynamicRegistryManager registryManager){ return result; }
    public boolean testAddition(ItemStack stack){ return addition.test(stack); }
    public ItemStack createIcon(){ return new ItemStack(RuneCraftingBlock.INSTANCE); }
    public Identifier getId(){ return id; }
    public RecipeSerializer<?> getSerializer(){ return RuneCrafting.RECIPE_SERIALIZER; }
    public RecipeType<?> getType(){ return RuneCrafting.RECIPE_TYPE; }
    public boolean isEmpty(){ return Stream.of(base,addition).anyMatch(i -> i.getMatchingStacks().length==0); }
    public static final String NAME="crafting";
    public static class Serializer implements RecipeSerializer<RuneCraftingRecipe> {
        public RuneCraftingRecipe read(Identifier id, JsonObject json){
            Ingredient base=Ingredient.fromJson(JsonHelper.getObject(json,"base"));
            Ingredient addition=Ingredient.fromJson(JsonHelper.getObject(json,"addition"));
            ItemStack result=ShapedRecipe.outputFromJson(JsonHelper.getObject(json,"result"));
            return new RuneCraftingRecipe(id,base,addition,result);
        }
        public RuneCraftingRecipe read(Identifier id, PacketByteBuf buf){
            return new RuneCraftingRecipe(id,Ingredient.fromPacket(buf),Ingredient.fromPacket(buf),buf.readItemStack());
        }
        public void write(PacketByteBuf buf,RuneCraftingRecipe recipe){
            recipe.base.write(buf); recipe.addition.write(buf); buf.writeItemStack(recipe.result);
        }
    }
}
