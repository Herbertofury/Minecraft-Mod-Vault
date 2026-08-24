package net.spell_power.forge;

import net.minecraft.entity.attribute.EntityAttribute;
import net.minecraft.registry.Registries;
import net.minecraft.registry.RegistryKeys;
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
    public ForgeMod(){
        SpellPowerMod.refreshConfigs();
        IEventBus bus=FMLJavaModLoadingContext.get().getModEventBus();
        bus.addListener(ForgeMod::register);
        bus.addListener(ForgeMod::modifyAttributes);
        bus.addListener(ForgeMod::commonSetup);
        MinecraftForge.EVENT_BUS.addListener(ForgeMod::playerLogin);
        if(Boolean.getBoolean("spellPower.ci.selftest")) MinecraftForge.EVENT_BUS.addListener(CiSelfTest::onServerStarted);
    }
    private static void register(RegisterEvent event){
        event.register(RegistryKeys.ATTRIBUTE,helper->{
            SpellPowerMod.touchRegistries();
            for(var e:SpellPowerMechanics.all.values()) helper.register(e.id,e.attribute);
            for(var s:SpellSchools.all()) if(s.attributeManagement.isInternal()) helper.register(s.id,s.attribute);
            for(var r:SpellResistance.Attributes.all) helper.register(r.id,r.attribute);
        });
        event.register(RegistryKeys.STATUS_EFFECT,helper->{
            SpellPowerMod.prepareStatusEffects();
            for(var s:SpellSchools.all()) if(s.powerEffectManagement.isInternal() && s.boostEffect!=null) helper.register(s.id,s.boostEffect);
            for(var e:SpellPowerMechanics.all.values()) helper.register(e.id,e.boostEffect);
        });
        event.register(RegistryKeys.ENCHANTMENT,helper->{
            for(var e:Enchantments_SpellPowerMechanics.all.entrySet()) helper.register(e.getKey(),e.getValue());
            for(var e:Enchantments_SpellPower.all.entrySet()) helper.register(e.getKey(),e.getValue());
        });
    }
    private static void modifyAttributes(EntityAttributeModificationEvent event){
        for(var type:event.getTypes()){
            for(var e:SpellPowerMechanics.all.values()) add(event,type,e.attribute);
            for(var s:SpellSchools.all()) if(s.attributeManagement.isInternal()) add(event,type,s.attribute);
            for(var r:SpellResistance.Attributes.all) add(event,type,r.attribute);
        }
    }
    private static void add(EntityAttributeModificationEvent event,net.minecraft.entity.EntityType<? extends net.minecraft.entity.LivingEntity> type,EntityAttribute attribute){
        var entry=Registries.ATTRIBUTE.getEntry(attribute);
        if(!event.has(type,entry)) event.add(type,entry);
    }
    private static void commonSetup(FMLCommonSetupEvent event){ event.enqueueWork(SpellPowerMod::applyEnchantments); }
    private static void playerLogin(PlayerEvent.PlayerLoggedInEvent event){ }
}
