package net.fabric_extras.ranged_weapon.client;

import net.fabric_extras.ranged_weapon.api.CustomBow;
import net.fabric_extras.ranged_weapon.api.CustomCrossbow;
import net.fabric_extras.ranged_weapon.api.CustomRangedWeapon;
import net.minecraft.client.item.ModelPredicateProviderRegistry;
import net.minecraft.item.Items;
import net.minecraft.util.Identifier;

public final class ModelPredicateHelper {
    public static void registerBowModelPredicates(CustomBow bow) {
        ModelPredicateProviderRegistry.register(bow, new Identifier("pull"), (stack, world, entity, seed) -> {
            if (entity == null || entity.getActiveItem() != stack) return 0F;
            var seconds = ((CustomRangedWeapon)(Object)bow).getRangedWeaponConfig().pullTimeSeconds();
            return (float)(stack.getMaxUseTime() - entity.getItemUseTimeLeft()) / Math.max(1F, seconds * 20F);
        });
        ModelPredicateProviderRegistry.register(bow, new Identifier("pulling"), (stack, world, entity, seed) -> entity != null && entity.isUsingItem() && entity.getActiveItem() == stack ? 1F : 0F);
    }
    public static void registerCrossbowModelPredicates(CustomCrossbow crossbow) {
        for (var id : new Identifier[]{new Identifier("pull"),new Identifier("pulling"),new Identifier("charged"),new Identifier("firework")}) {
            var provider = ModelPredicateProviderRegistry.get(Items.CROSSBOW, id);
            if (provider != null) ModelPredicateProviderRegistry.register(crossbow, id, provider);
        }
    }
    private ModelPredicateHelper() {}
}
