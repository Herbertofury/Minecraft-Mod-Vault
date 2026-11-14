package net.fabric_extras.ranged_weapon.forge;

import net.fabric_extras.ranged_weapon.RangedWeaponMod;
import net.fabric_extras.ranged_weapon.api.*;
import net.minecraft.entity.LivingEntity;
import net.minecraft.entity.attribute.EntityAttribute;
import net.minecraft.entity.effect.StatusEffectInstance;
import net.minecraft.potion.Potion;
import net.minecraft.registry.RegistryKeys;
import net.minecraftforge.event.entity.EntityAttributeModificationEvent;
import net.minecraftforge.eventbus.api.IEventBus;
import net.minecraftforge.fml.common.Mod;
import net.minecraftforge.fml.event.lifecycle.FMLCommonSetupEvent;
import net.minecraftforge.fml.javafmlmod.FMLJavaModLoadingContext;
import net.minecraftforge.registries.RegisterEvent;

@Mod(RangedWeaponMod.ID)
public final class ForgeMod {
    public ForgeMod() {
        RangedWeaponMod.init();
        if (Boolean.getBoolean("rangedWeapon.ci.selftest")) RangedWeaponMod.registerPotions();
        IEventBus bus=FMLJavaModLoadingContext.get().getModEventBus();
        bus.addListener(ForgeMod::register); bus.addListener(ForgeMod::attributes); bus.addListener(ForgeMod::setup);
        if (Boolean.getBoolean("rangedWeapon.ci.selftest")) net.minecraftforge.common.MinecraftForge.EVENT_BUS.addListener(CiSelfTest::onServerStarted);
    }
    private static void register(RegisterEvent event) {
        event.register(RegistryKeys.ATTRIBUTE, helper -> { for (var e:EntityAttributes_RangedWeapon.all) helper.register(e.id,e.attribute); });
        event.register(RegistryKeys.STATUS_EFFECT, helper -> { RangedWeaponMod.prepareStatusEffects(); for (var e:StatusEffects_RangedWeapon.all) helper.register(e.id,e.effect); });
        event.register(RegistryKeys.POTION, helper -> {
            if (!RangedWeaponMod.potionsRequested()) return;
            for (var e:StatusEffects_RangedWeapon.all) helper.register(RangedWeaponMod.potionIdFrom(e.id),new Potion(new StatusEffectInstance(e.effect,3600)));
        });
    }
    private static void attributes(EntityAttributeModificationEvent event) {
        for (var type:event.getTypes()) for (var entry:EntityAttributes_RangedWeapon.all) add(event,type,entry.attribute);
    }
    private static void add(EntityAttributeModificationEvent event, net.minecraft.entity.EntityType<? extends LivingEntity> type, EntityAttribute attr) { if (!event.has(type,attr)) event.add(type,attr); }
    private static void setup(FMLCommonSetupEvent event) { event.enqueueWork(RangedWeaponMod::bindRegistryEntries); }
}
