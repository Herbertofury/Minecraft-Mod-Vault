package net.spell_power.forge;

import net.minecraft.entity.attribute.EntityAttribute;
import net.minecraft.entity.effect.StatusEffectInstance;
import net.minecraft.potion.Potion;
import net.minecraft.registry.RegistryKeys;
import net.minecraft.server.network.ServerPlayerEntity;
import net.minecraftforge.common.MinecraftForge;
import net.minecraftforge.event.entity.EntityAttributeModificationEvent;
import net.minecraftforge.event.entity.player.PlayerEvent;
import net.minecraftforge.eventbus.api.IEventBus;
import net.minecraftforge.fml.common.Mod;
import net.minecraftforge.fml.event.lifecycle.FMLCommonSetupEvent;
import net.minecraftforge.fml.javafmlmod.FMLJavaModLoadingContext;
import net.minecraftforge.registries.RegisterEvent;
import net.spell_power.SpellPowerMod;
import net.spell_power.api.*;
import net.spell_power.api.enchantment.Enchantments_SpellPower;
import net.spell_power.api.enchantment.Enchantments_SpellPowerMechanics;

@Mod(SpellPowerMod.ID)
public final class ForgeMod {
    public ForgeMod() {
        SpellPowerMod.refreshConfigs();
        IEventBus bus = FMLJavaModLoadingContext.get().getModEventBus();
        bus.addListener(ForgeMod::register); bus.addListener(ForgeMod::modifyAttributes); bus.addListener(ForgeMod::commonSetup);
        MinecraftForge.EVENT_BUS.addListener(ForgeMod::playerLogin);
        if (Boolean.getBoolean("spellPower.ci.selftest")) MinecraftForge.EVENT_BUS.addListener(CiSelfTest::onServerStarted);
    }
    private static void register(RegisterEvent event) {
        event.register(RegistryKeys.ATTRIBUTE, helper -> {
            SpellPowerMod.touchRegistries();
            for (var entry : SpellPowerMechanics.all.values()) helper.register(entry.id, entry.attribute);
            for (var school : SpellSchools.all()) if (school.attributeManagement.isInternal()) helper.register(school.id, school.attribute);
            for (var resistance : SpellResistance.Attributes.all) helper.register(resistance.id, resistance.attribute);
        });
        event.register(RegistryKeys.STATUS_EFFECT, helper -> {
            SpellPowerMod.prepareStatusEffects();
            for (var school : SpellSchools.all()) if (school.powerEffectManagement.isInternal() && school.boostEffect != null) helper.register(school.id, school.boostEffect);
            for (var entry : SpellPowerMechanics.all.values()) helper.register(entry.id, entry.boostEffect);
        });
        event.register(RegistryKeys.POTION, helper -> {
            if (!SpellPowerMod.attributesConfig.value.register_potions) return;
            for (var school : SpellSchools.all()) {
                if (school.archetype == SpellSchool.Archetype.MAGIC && !school.id.getPath().contains("generic") && school.boostEffect != null) {
                    helper.register(SpellPowerMod.potionIdFrom(school.id), new Potion(new StatusEffectInstance(school.boostEffect, 3600)));
                }
            }
            for (var entry : SpellPowerMechanics.all.values()) helper.register(SpellPowerMod.potionIdFrom(entry.id), new Potion(new StatusEffectInstance(entry.boostEffect, 3600)));
        });
        event.register(RegistryKeys.ENCHANTMENT, helper -> {
            for (var entry : Enchantments_SpellPowerMechanics.all.entrySet()) helper.register(entry.getKey(), entry.getValue());
            for (var entry : Enchantments_SpellPower.all.entrySet()) helper.register(entry.getKey(), entry.getValue());
        });
    }
    private static void modifyAttributes(EntityAttributeModificationEvent event) {
        for (var type : event.getTypes()) {
            for (var entry : SpellPowerMechanics.all.values()) add(event, type, entry.attribute);
            for (var school : SpellSchools.all()) if (school.attributeManagement.isInternal()) add(event, type, school.attribute);
            for (var resistance : SpellResistance.Attributes.all) add(event, type, resistance.attribute);
        }
    }
    private static void add(EntityAttributeModificationEvent event, net.minecraft.entity.EntityType<? extends net.minecraft.entity.LivingEntity> type, EntityAttribute attribute) {
        if (!event.has(type, attribute)) event.add(type, attribute);
    }
    private static void commonSetup(FMLCommonSetupEvent event) { event.enqueueWork(SpellPowerMod::applyEnchantments); }
    private static void playerLogin(PlayerEvent.PlayerLoggedInEvent event) {
        if (event.getEntity() instanceof ServerPlayerEntity player) SpellPowerMod.onPlayerJoin(player);
    }
}
