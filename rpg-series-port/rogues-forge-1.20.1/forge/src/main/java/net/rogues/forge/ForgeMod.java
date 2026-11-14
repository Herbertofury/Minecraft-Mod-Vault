package net.rogues.forge;

import net.minecraft.registry.Registries;
import net.minecraft.registry.Registry;
import net.minecraft.registry.RegistryKeys;
import net.minecraft.world.poi.PointOfInterestType;
import net.minecraftforge.common.MinecraftForge;
import net.minecraftforge.event.BuildCreativeModeTabContentsEvent;
import net.minecraftforge.event.village.VillagerTradesEvent;
import net.minecraftforge.eventbus.api.IEventBus;
import net.minecraftforge.fml.common.Mod;
import net.minecraftforge.fml.javafmlmod.FMLJavaModLoadingContext;
import net.minecraftforge.registries.RegisterEvent;
import net.rogues.RoguesMod;
import net.rogues.block.CustomBlocks;
import net.rogues.item.Group;
import net.rogues.village.RogueVillagers;
import net.spell_engine.compat.registry.RegistrationBridge;

@Mod(RoguesMod.ID)
public final class ForgeMod {
    public ForgeMod() {
        IEventBus bus=FMLJavaModLoadingContext.get().getModEventBus(); RoguesMod.init(); bus.addListener(ForgeMod::register); bus.addListener(ForgeMod::buildTabContents); MinecraftForge.EVENT_BUS.addListener(ForgeMod::onVillagerTrades);
    }
    public static void register(RegisterEvent e) {
        e.register(RegistryKeys.SOUND_EVENT,h->withNative(Registries.SOUND_EVENT,h,RoguesMod::registerSounds));
        e.register(RegistryKeys.BLOCK,h->withNative(Registries.BLOCK,h,RoguesMod::registerBlocks));
        e.register(Registries.ITEM_GROUP.getKey(),h->withNative(Registries.ITEM_GROUP,h,RoguesMod::registerItemGroup));
        e.register(RegistryKeys.ITEM,h->withNative(Registries.ITEM,h,RoguesMod::registerItems));
        e.register(RegistryKeys.STATUS_EFFECT,h->withNative(Registries.STATUS_EFFECT,h,RoguesMod::registerEffects));
        e.register(RegistryKeys.ENTITY_TYPE,h->withNative(Registries.ENTITY_TYPE,h,RoguesMod::registerEntities));
        e.register(RegistryKeys.POINT_OF_INTEREST_TYPE,h->h.register(RogueVillagers.POI_ID,new PointOfInterestType(RogueVillagers.poiBlockStates(),RogueVillagers.POI_TICKET_COUNT,RogueVillagers.POI_SEARCH_DISTANCE)));
        e.register(RegistryKeys.VILLAGER_PROFESSION,h->withNative(Registries.VILLAGER_PROFESSION,h,RoguesMod::registerVillagers));
    }
    @SuppressWarnings("unchecked") private static <T> void withNative(Registry<T> r,RegisterEvent.RegisterHelper<T> h,Runnable a){RegistrationBridge.withRegistrar(r,(actual,id,value)->h.register(id,(T)value),a);}
    private static void buildTabContents(BuildCreativeModeTabContentsEvent e){if(e.getTabKey().equals(Group.KEY)) for(var entry:CustomBlocks.all)e.accept(()->entry.item());}
    private static void onVillagerTrades(VillagerTradesEvent e){if(e.getType()!=RogueVillagers.PROFESSION)return; RogueVillagers.TRADES.forEach((tier,factories)->{var list=e.getTrades().get(tier.intValue());if(list!=null)list.addAll(factories);});}
}
