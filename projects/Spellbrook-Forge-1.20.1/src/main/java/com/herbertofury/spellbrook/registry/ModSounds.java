package com.herbertofury.spellbrook.registry;

import com.herbertofury.spellbrook.Spellbrook;
import net.minecraft.resources.ResourceLocation;
import net.minecraft.sounds.SoundEvent;
import net.minecraftforge.eventbus.api.IEventBus;
import net.minecraftforge.registries.DeferredRegister;
import net.minecraftforge.registries.ForgeRegistries;
import net.minecraftforge.registries.RegistryObject;

public final class ModSounds {
    public static final DeferredRegister<SoundEvent> SOUNDS=DeferredRegister.create(ForgeRegistries.SOUND_EVENTS, Spellbrook.MOD_ID);
    public static final RegistryObject<SoundEvent> LAUNCH=register("broom.launch");
    public static final RegistryObject<SoundEvent> LAND=register("broom.land");
    public static final RegistryObject<SoundEvent> BOOST=register("broom.boost");
    public static final RegistryObject<SoundEvent> MOUNT=register("broom.mount");
    public static final RegistryObject<SoundEvent> DISMOUNT=register("broom.dismount");
    private static RegistryObject<SoundEvent> register(String id){ return SOUNDS.register(id.replace('.','_'),()->SoundEvent.createVariableRangeEvent(new ResourceLocation(Spellbrook.MOD_ID,id))); }
    private ModSounds(){}
    public static void register(IEventBus bus){ SOUNDS.register(bus); }
}
