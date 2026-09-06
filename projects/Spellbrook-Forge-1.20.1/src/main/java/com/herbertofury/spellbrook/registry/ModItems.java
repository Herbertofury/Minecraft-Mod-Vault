package com.herbertofury.spellbrook.registry;

import com.herbertofury.spellbrook.Spellbrook;
import com.herbertofury.spellbrook.broom.BroomVariant;
import com.herbertofury.spellbrook.broom.CoreVariant;
import com.herbertofury.spellbrook.item.CoreCrystalItem;
import com.herbertofury.spellbrook.item.SpellbrookBroomItem;
import net.minecraft.world.item.CreativeModeTabs;
import net.minecraft.world.item.Item;
import net.minecraftforge.event.BuildCreativeModeTabContentsEvent;
import net.minecraftforge.eventbus.api.IEventBus;
import net.minecraftforge.registries.DeferredRegister;
import net.minecraftforge.registries.ForgeRegistries;
import net.minecraftforge.registries.RegistryObject;

public final class ModItems {
    public static final DeferredRegister<Item> ITEMS=DeferredRegister.create(ForgeRegistries.ITEMS, Spellbrook.MOD_ID);
    public static final RegistryObject<Item> BROOMSTICK=ITEMS.register("broomstick",()->new SpellbrookBroomItem(new Item.Properties().stacksTo(1)));
    public static final RegistryObject<Item> CORE_CRYSTAL=ITEMS.register("core_crystal",()->new CoreCrystalItem(new Item.Properties().stacksTo(16)));
    private ModItems(){}
    public static void register(IEventBus bus){ ITEMS.register(bus); }
    public static void buildCreativeTab(BuildCreativeModeTabContentsEvent event){
        if(event.getTabKey()!=CreativeModeTabs.TOOLS_AND_UTILITIES) return;
        for(BroomVariant v:BroomVariant.values()) if(v.publicVariant()) event.accept(SpellbrookBroomItem.stackFor(BROOMSTICK.get(),v));
        for(CoreVariant v:CoreVariant.values()) if(v.publicVariant()) event.accept(CoreCrystalItem.stackFor(CORE_CRYSTAL.get(),v));
    }
}
