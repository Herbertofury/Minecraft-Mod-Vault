package com.herbertofury.spellbrook.registry;

import com.herbertofury.spellbrook.Spellbrook;
import com.herbertofury.spellbrook.entity.SpellbrookBroomEntity;
import net.minecraft.world.entity.EntityType;
import net.minecraft.world.entity.MobCategory;
import net.minecraftforge.eventbus.api.IEventBus;
import net.minecraftforge.registries.DeferredRegister;
import net.minecraftforge.registries.ForgeRegistries;
import net.minecraftforge.registries.RegistryObject;

public final class ModEntities {
    public static final DeferredRegister<EntityType<?>> ENTITIES=DeferredRegister.create(ForgeRegistries.ENTITY_TYPES, Spellbrook.MOD_ID);
    public static final RegistryObject<EntityType<SpellbrookBroomEntity>> BROOM=ENTITIES.register("broom",()->EntityType.Builder.<SpellbrookBroomEntity>of(SpellbrookBroomEntity::new, MobCategory.MISC).sized(1.75f,0.7f).clientTrackingRange(10).updateInterval(1).build("spellbrook:broom"));
    private ModEntities(){}
    public static void register(IEventBus bus){ ENTITIES.register(bus); }
}
