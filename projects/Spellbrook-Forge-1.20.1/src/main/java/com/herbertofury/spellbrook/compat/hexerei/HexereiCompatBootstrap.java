package com.herbertofury.spellbrook.compat.hexerei;

import com.herbertofury.spellbrook.Spellbrook;
import com.herbertofury.spellbrook.broom.BroomVariant;
import com.herbertofury.spellbrook.client.BroomRenderUtil;
import com.herbertofury.spellbrook.compat.CompatHooks;
import com.herbertofury.spellbrook.registry.ModItems;
import net.joefoxe.hexerei.client.renderer.entity.BroomType;
import net.minecraft.core.BlockPos;
import net.minecraft.world.entity.EntityType;
import net.minecraft.world.entity.MobCategory;
import net.minecraft.world.entity.player.Player;
import net.minecraft.world.item.ItemStack;
import net.minecraft.world.level.Level;
import net.minecraftforge.api.distmarker.Dist;
import net.minecraftforge.eventbus.api.IEventBus;
import net.minecraftforge.fml.DistExecutor;
import net.minecraftforge.fml.event.lifecycle.FMLCommonSetupEvent;
import net.minecraftforge.registries.DeferredRegister;
import net.minecraftforge.registries.ForgeRegistries;
import net.minecraftforge.registries.RegistryObject;

public final class HexereiCompatBootstrap {
    private static final DeferredRegister<EntityType<?>> ENTITIES=DeferredRegister.create(ForgeRegistries.ENTITY_TYPES,Spellbrook.MOD_ID);
    public static final RegistryObject<EntityType<HexereiSpellbrookBroomEntity>> BROOM=ENTITIES.register("hexerei_broom",()->EntityType.Builder.<HexereiSpellbrookBroomEntity>of(HexereiSpellbrookBroomEntity::new, MobCategory.MISC).sized(1.75f,0.7f).clientTrackingRange(10).updateInterval(1).build("spellbrook:hexerei_broom"));
    private HexereiCompatBootstrap(){}
    public static void register(IEventBus modBus){ENTITIES.register(modBus);modBus.addListener(HexereiCompatBootstrap::commonSetup);CompatHooks.installHexerei(HexereiCompatBootstrap::spawn);DistExecutor.unsafeRunWhenOn(Dist.CLIENT,()->()->HexereiCompatClient.register(modBus));}
    private static void commonSetup(FMLCommonSetupEvent event){event.enqueueWork(()->{for(BroomVariant v:BroomVariant.values())BroomType.create("spellbrook_"+v.id(),ModItems.BROOMSTICK.get(),0.65f*v.speed());});}
    private static boolean spawn(Level level, BlockPos pos, ItemStack stack, Player player,float yaw){if(level.isClientSide)return true;HexereiSpellbrookBroomEntity broom=BROOM.get().create(level);if(broom==null)return false;broom.absMoveTo(pos.getX()+0.5,pos.getY()+0.1,pos.getZ()+0.5,yaw,0);broom.loadFromSpellbrookItem(stack,player);if(!level.noCollision(broom,broom.getBoundingBox()))return false;level.addFreshEntity(broom);return true;}
}
