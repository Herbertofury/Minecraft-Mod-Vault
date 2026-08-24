package net.spell_power.forge;

import net.minecraft.entity.EntityType;
import net.minecraft.registry.Registries;
import net.minecraftforge.event.server.ServerStartedEvent;
import net.spell_power.SpellPowerMod;
import net.spell_power.api.*;

public final class CiSelfTest {
    private CiSelfTest(){}
    public static void onServerStarted(ServerStartedEvent event){
        var server=event.getServer();
        require("spell_power".equals(Registries.ATTRIBUTE.getId(SpellSchools.ARCANE.attribute).getNamespace()),"arcane attribute registry");
        require(Registries.ATTRIBUTE.getId(SpellSchools.GENERIC.attribute).getPath().equals("generic"),"generic attribute registry");
        require(SpellPowerMechanics.HASTE.min==10F,"sub-1 haste floor");
        require(SpellPowerMod.attributesConfig.value.resistance_curve==DamageCurve.HYPERBOLIC,"hyperbolic resistance default");
        var zombie=EntityType.ZOMBIE.create(server.getOverworld());
        require(zombie!=null && zombie.getAttributeInstance(SpellSchools.ARCANE.attribute)!=null,"living entity spell attributes");
        require(zombie.getAttributeInstance(SpellResistance.Attributes.GENERIC.attribute)!=null,"living entity resistance attribute");
        var crit=SpellPower.getSpellPower(SpellSchools.ARCANE,zombie).forcedCritical();
        require(crit.isCritical(),"Result exposes isCritical");
        System.out.println("[Spell Power CI] Runtime self-test passed: registries + living attributes + 1.6 mechanics");
    }
    private static void require(boolean ok,String label){ if(!ok) throw new IllegalStateException("[Spell Power CI] Failed: "+label); }
}
