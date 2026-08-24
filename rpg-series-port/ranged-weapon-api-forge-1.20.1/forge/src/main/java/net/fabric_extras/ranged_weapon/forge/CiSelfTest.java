package net.fabric_extras.ranged_weapon.forge;

import net.fabric_extras.ranged_weapon.RangedWeaponMod;
import net.fabric_extras.ranged_weapon.api.*;
import net.fabric_extras.ranged_weapon.internal.ScalingUtil;
import net.minecraft.entity.EntityType;
import net.minecraft.entity.EquipmentSlot;
import net.minecraft.entity.attribute.EntityAttributeModifier;
import net.minecraft.entity.projectile.ProjectileUtil;
import net.minecraft.item.Item;
import net.minecraft.item.ItemStack;
import net.minecraft.item.Items;
import net.minecraft.recipe.Ingredient;
import net.minecraft.registry.Registries;
import net.minecraftforge.event.server.ServerStartedEvent;

public final class CiSelfTest {
    private CiSelfTest() {}
    public static void onServerStarted(ServerStartedEvent event) {
        var world=event.getServer().getOverworld();
        require(new net.minecraft.util.Identifier("ranged_weapon","damage").equals(Registries.ATTRIBUTE.getId(EntityAttributes_RangedWeapon.DAMAGE.attribute)),"damage registry");
        require(new net.minecraft.util.Identifier("ranged_weapon","pull_time").equals(Registries.ATTRIBUTE.getId(EntityAttributes_RangedWeapon.PULL_TIME.attribute)),"pull time registry");
        require(Registries.STATUS_EFFECT.getId(StatusEffects_RangedWeapon.HASTE.effect).getPath().equals("haste"),"haste status effect registry");
        require(Registries.POTION.containsId(RangedWeaponMod.potionIdFrom(StatusEffects_RangedWeapon.DAMAGE.id)),"potion helper registration");

        // Prove the item-side modifier wiring first. LivingEntity applies equipment changes on its
        // normal equipment-check tick, so reading a just-equipped mob before that tick is not a
        // valid gameplay test.
        require(Items.BOW.getAttributeModifiers(EquipmentSlot.MAINHAND)
                        .get(EntityAttributes_RangedWeapon.DAMAGE.attribute).stream()
                        .anyMatch(modifier -> close(modifier.getValue(),6)),
                "vanilla bow item damage modifier");

        var skeleton=EntityType.SKELETON.create(world); require(skeleton!=null,"skeleton creation");
        require(skeleton.getAttributeInstance(EntityAttributes_RangedWeapon.DAMAGE.attribute)!=null,"living damage attribute");
        require(skeleton.getAttributeInstance(EntityAttributes_RangedWeapon.PULL_TIME.attribute)!=null,"living pull-time attribute");
        require(world.spawnEntity(skeleton),"skeleton spawn");

        skeleton.equipStack(EquipmentSlot.MAINHAND,new ItemStack(Items.BOW));
        skeleton.tick(); // normal equipment-diff pass applies the bow's attribute modifiers
        require(close(skeleton.getAttributeValue(EntityAttributes_RangedWeapon.DAMAGE.attribute),6),"vanilla bow baseline damage");
        require(close(skeleton.getAttributeValue(EntityAttributes_RangedWeapon.PULL_TIME.attribute),1),"vanilla bow pull baseline");
        var vanillaMobArrow=ProjectileUtil.createArrowProjectile(skeleton,new ItemStack(Items.ARROW),1F);
        double vanillaMobDamage=vanillaMobArrow.getDamage();

        var config=new RangedConfig(10,0.5F,0.2F).withAttribute(EntityAttributes_RangedWeapon.HASTE.id,EntityAttributeModifier.Operation.ADDITION,20);
        var custom=new CustomBow(new Item.Settings(),config,()->Ingredient.EMPTY);
        require(custom.getAttributeModifiers(EquipmentSlot.MAINHAND)
                        .get(EntityAttributes_RangedWeapon.DAMAGE.attribute).stream()
                        .anyMatch(modifier -> close(modifier.getValue(),10)),
                "custom bow item damage modifier");
        require(custom.getAttributeModifiers(EquipmentSlot.MAINHAND)
                        .get(EntityAttributes_RangedWeapon.PULL_TIME.attribute).stream()
                        .anyMatch(modifier -> close(modifier.getValue(),0.5)),
                "custom bow item pull modifier");

        skeleton.equipStack(EquipmentSlot.MAINHAND,new ItemStack(custom));
        skeleton.tick(); // removes vanilla modifiers and applies the custom stack modifiers
        require(close(skeleton.getAttributeValue(EntityAttributes_RangedWeapon.DAMAGE.attribute),10),"custom bow damage");
        require(close(skeleton.getAttributeValue(EntityAttributes_RangedWeapon.PULL_TIME.attribute),1.5),"custom bow pull-time");
        require(close(skeleton.getAttributeValue(EntityAttributes_RangedWeapon.VELOCITY.attribute),0.2),"custom bow velocity bonus");
        require(close(skeleton.getAttributeValue(EntityAttributes_RangedWeapon.HASTE.attribute),120),"stacked optional config attribute");
        var customMobArrow=ProjectileUtil.createArrowProjectile(skeleton,new ItemStack(Items.ARROW),1F);
        require(vanillaMobDamage>0 && close(customMobArrow.getDamage()/vanillaMobDamage,10D/6D),"2.3.3 bow-using mob damage scaling");

        double vm=ScalingUtil.arrowVelocityMultiplier(custom,0.2); require(vm>1,"velocity scaling");
        require(ScalingUtil.arrowDamageMultiplier(6,10,vm)>1,"damage scaling");
        require(((CustomRangedWeapon)(Object)custom).getTypeBaseline()==RangedConfig.BOW,"custom bow type baseline");
        skeleton.discard();
        System.out.println("[Ranged Weapon API CI] Runtime self-test passed: registries + item/equipment attributes + 2.3.3 mob bow damage + scaling");
    }
    private static boolean close(double a,double b) { return Math.abs(a-b)<0.0001; }
    private static void require(boolean ok,String label) { if (!ok) throw new IllegalStateException("[Ranged Weapon API CI] Failed: "+label); }
}
