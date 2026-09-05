package net.fabric_extras.ranged_weapon.forge;

import net.fabric_extras.ranged_weapon.RangedWeaponMod;
import net.fabric_extras.ranged_weapon.api.*;
import net.fabric_extras.ranged_weapon.internal.ArrowExtension;
import net.fabric_extras.ranged_weapon.internal.ScalingUtil;
import net.minecraft.entity.EntityType;
import net.minecraft.entity.EquipmentSlot;
import net.minecraft.entity.attribute.EntityAttributeModifier;
import net.minecraft.entity.projectile.ProjectileUtil;
import net.minecraft.item.ItemStack;
import net.minecraft.item.Items;
import net.minecraft.registry.Registries;
import net.minecraftforge.event.server.ServerStartedEvent;

public final class CiSelfTest {
    private CiSelfTest() {}

    public static void onServerStarted(ServerStartedEvent event) {
        var world = event.getServer().getOverworld();
        require(new net.minecraft.util.Identifier("ranged_weapon", "damage")
                .equals(Registries.ATTRIBUTE.getId(EntityAttributes_RangedWeapon.DAMAGE.attribute)), "damage registry");
        require(new net.minecraft.util.Identifier("ranged_weapon", "pull_time")
                .equals(Registries.ATTRIBUTE.getId(EntityAttributes_RangedWeapon.PULL_TIME.attribute)), "pull time registry");
        require(Registries.STATUS_EFFECT.getId(StatusEffects_RangedWeapon.HASTE.effect).getPath().equals("haste"),
                "haste status effect registry");
        require(Registries.POTION.containsId(RangedWeaponMod.potionIdFrom(StatusEffects_RangedWeapon.DAMAGE.id)),
                "potion helper registration");

        var bow = (CustomRangedWeapon)(Object)Items.BOW;
        var originalConfig = bow.getRangedWeaponConfig();
        var skeleton = EntityType.SKELETON.create(world);
        require(skeleton != null, "skeleton creation");
        require(skeleton.getAttributeInstance(EntityAttributes_RangedWeapon.DAMAGE.attribute) != null,
                "living damage attribute");
        require(skeleton.getAttributeInstance(EntityAttributes_RangedWeapon.PULL_TIME.attribute) != null,
                "living pull-time attribute");
        require(world.spawnEntity(skeleton), "skeleton spawn");

        try {
            require(Items.BOW.getAttributeModifiers(EquipmentSlot.MAINHAND)
                            .get(EntityAttributes_RangedWeapon.DAMAGE.attribute).stream()
                            .anyMatch(modifier -> close(modifier.getValue(), 6)),
                    "vanilla bow item damage modifier");

            skeleton.equipStack(EquipmentSlot.MAINHAND, new ItemStack(Items.BOW));
            skeleton.tick();
            require(close(skeleton.getAttributeValue(EntityAttributes_RangedWeapon.DAMAGE.attribute), 6),
                    "vanilla bow baseline damage");
            require(close(skeleton.getAttributeValue(EntityAttributes_RangedWeapon.PULL_TIME.attribute), 1),
                    "vanilla bow pull baseline");
            var vanillaMobArrow = ProjectileUtil.createArrowProjectile(skeleton, new ItemStack(Items.ARROW), 1F);
            require(((ArrowExtension)(Object)vanillaMobArrow).rwa_isModified(),
                    "2.3.3 vanilla mob projectile hook");

            skeleton.equipStack(EquipmentSlot.MAINHAND, ItemStack.EMPTY);
            skeleton.tick();

            var customConfig = new RangedConfig(10, 0.5F, 0.2F)
                    .withAttribute(EntityAttributes_RangedWeapon.HASTE.id,
                            EntityAttributeModifier.Operation.ADDITION, 20);
            bow.setRangedWeaponConfig(customConfig);

            require(Items.BOW.getAttributeModifiers(EquipmentSlot.MAINHAND)
                            .get(EntityAttributes_RangedWeapon.DAMAGE.attribute).stream()
                            .anyMatch(modifier -> close(modifier.getValue(), 10)),
                    "custom config item damage modifier");
            require(Items.BOW.getAttributeModifiers(EquipmentSlot.MAINHAND)
                            .get(EntityAttributes_RangedWeapon.PULL_TIME.attribute).stream()
                            .anyMatch(modifier -> close(modifier.getValue(), 0.5)),
                    "custom config item pull modifier");

            skeleton.equipStack(EquipmentSlot.MAINHAND, new ItemStack(Items.BOW));
            skeleton.tick();
            require(close(skeleton.getAttributeValue(EntityAttributes_RangedWeapon.DAMAGE.attribute), 10),
                    "custom config damage");
            require(close(skeleton.getAttributeValue(EntityAttributes_RangedWeapon.PULL_TIME.attribute), 1.5),
                    "custom config pull-time");
            require(close(skeleton.getAttributeValue(EntityAttributes_RangedWeapon.VELOCITY.attribute), 0.2),
                    "custom config velocity bonus");
            require(close(skeleton.getAttributeValue(EntityAttributes_RangedWeapon.HASTE.attribute), 120),
                    "stacked optional config attribute");

            var customMobArrow = ProjectileUtil.createArrowProjectile(skeleton, new ItemStack(Items.ARROW), 1F);
            require(((ArrowExtension)(Object)customMobArrow).rwa_isModified(),
                    "2.3.3 custom mob projectile hook");
            // Vanilla seeds each arrow with independent random damage before RWA's multiplier, so
            // comparing two arrows by exact final ratio is invalid. Prove the actual hook executed
            // above, and prove the shared deterministic multiplier used by that hook here.
            require(close(ScalingUtil.arrowDamageMultiplier(6, 10, 1D), 10D / 6D),
                    "2.3.3 mob damage multiplier math");

            double velocityMultiplier = ScalingUtil.arrowVelocityMultiplier(Items.BOW, 0.2);
            require(velocityMultiplier > 1, "velocity scaling");
            require(ScalingUtil.arrowDamageMultiplier(6, 10, velocityMultiplier) > 1, "damage scaling");
            require(bow.getTypeBaseline() == RangedConfig.BOW, "vanilla bow type baseline");

            System.out.println("[Ranged Weapon API CI] Runtime self-test passed: registries + item/equipment attributes + 2.3.3 mob hook + scaling");
        } finally {
            skeleton.equipStack(EquipmentSlot.MAINHAND, ItemStack.EMPTY);
            skeleton.tick();
            bow.setRangedWeaponConfig(originalConfig);
            skeleton.discard();
        }
    }

    private static boolean close(double a, double b) { return Math.abs(a - b) < 0.0001; }
    private static void require(boolean ok, String label) {
        if (!ok) throw new IllegalStateException("[Ranged Weapon API CI] Failed: " + label);
    }
}
