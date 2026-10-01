package net.fabric_extras.shield_api.forge;

import com.mojang.authlib.GameProfile;
import net.fabric_extras.shield_api.ShieldAPI;
import net.fabric_extras.shield_api.item.CustomShieldItem;
import net.minecraft.entity.EquipmentSlot;
import net.minecraft.entity.attribute.EntityAttributeModifier;
import net.minecraft.entity.attribute.EntityAttributes;
import net.minecraft.entity.player.ItemCooldownManager;
import net.minecraft.item.Item;
import net.minecraft.item.ItemStack;
import net.minecraft.item.Items;
import net.minecraft.recipe.Ingredient;
import net.minecraft.server.MinecraftServer;
import net.minecraft.server.network.ServerPlayerEntity;
import net.minecraft.server.world.ServerWorld;
import net.minecraft.stat.Stats;
import net.minecraft.util.Hand;
import net.minecraft.util.Pair;
import net.minecraftforge.common.ToolActions;
import net.minecraftforge.event.server.ServerStartedEvent;
import net.minecraftforge.registries.DeferredRegister;
import net.minecraftforge.registries.ForgeRegistries;
import net.minecraftforge.registries.RegistryObject;

import java.util.IdentityHashMap;
import java.util.List;
import java.util.Map;
import java.util.UUID;

final class ShieldAPISelfTest {
    static final DeferredRegister<Item> TEST_ITEMS = DeferredRegister.create(ForgeRegistries.ITEMS, ShieldAPI.MOD_ID);
    static final RegistryObject<Item> TEST_SHIELD = TEST_ITEMS.register(
            "self_test_shield",
            () -> new CustomShieldItem(
                    null,
                    () -> Ingredient.ofItems(Items.IRON_INGOT),
                    List.of(new Pair<>(
                            EntityAttributes.GENERIC_ARMOR,
                            new EntityAttributeModifier("shield_api_self_test_armor", 4.0D, EntityAttributeModifier.Operation.ADDITION)
                    )),
                    new Item.Settings().maxDamage(8)
            )
    );

    private ShieldAPISelfTest() { }

    static boolean enabled() {
        return "1".equals(System.getenv("SHIELD_API_SELF_TEST"));
    }

    static void onServerStarted(ServerStartedEvent event) {
        if (!enabled()) return;
        try {
            run(event);
            ShieldAPI.LOGGER.info("SHIELD_API_SELF_TEST_PASS");
        } catch (Throwable t) {
            ShieldAPI.LOGGER.error("SHIELD_API_SELF_TEST_FAILED", t);
            throw t;
        }
    }

    private static void run(ServerStartedEvent event) {
        CustomShieldItem shield = (CustomShieldItem) TEST_SHIELD.get();
        require(CustomShieldItem.instances.contains(shield), "registered custom shield missing from instances");
        require(shield.canRepair(new ItemStack(shield), new ItemStack(Items.IRON_INGOT)), "repair ingredient contract failed");
        require(!shield.canRepair(new ItemStack(shield), new ItemStack(Items.DIRT)), "non-repair ingredient was accepted");
        require(shield.getEquipSound() != null, "null equip sound did not fall back to ShieldItem sound");
        require(new ItemStack(shield).canPerformAction(ToolActions.SHIELD_BLOCK),
                "CustomShieldItem did not inherit Forge SHIELD_BLOCK action from ShieldItem");

        var initial = shield.getAttributeModifiers(EquipmentSlot.OFFHAND).get(EntityAttributes.GENERIC_ARMOR);
        require(initial.size() == 1, "initial offhand armor modifier missing");
        require(Math.abs(initial.iterator().next().getValue() - 4.0D) < 0.000001D, "initial armor modifier value changed");
        require(shield.getAttributeModifiers(EquipmentSlot.MAINHAND).get(EntityAttributes.GENERIC_ARMOR).isEmpty(),
                "shield armor modifier leaked into main hand");

        shield.setAttributeModifiers(List.of(new Pair<>(
                EntityAttributes.GENERIC_ARMOR,
                new EntityAttributeModifier("shield_api_self_test_armor_mutated", 2.0D, EntityAttributeModifier.Operation.ADDITION)
        )));
        var mutated = shield.getAttributeModifiers(EquipmentSlot.OFFHAND).get(EntityAttributes.GENERIC_ARMOR);
        require(mutated.size() == 1, "mutated offhand armor modifier missing");
        require(Math.abs(mutated.iterator().next().getValue() - 2.0D) < 0.000001D, "setAttributeModifiers did not replace value");

        TestServerPlayer player = new TestServerPlayer(
                event.getServer(),
                event.getServer().getOverworld(),
                new GameProfile(UUID.fromString("00000000-0000-0000-0000-000000000211"), "ShieldApiSelfTest")
        );

        // Historical 1.20.1 Shield API could probabilistically skip this when non-sprinting.
        // Current 2.1.0 semantics are unconditional and exactly 100 ticks for every CustomShieldItem.
        // instances is a HashSet, so iteration order is intentionally not part of the contract.
        player.disableShield(false);
        require(player.trackingCooldown.calls > 0, "disableShield(false) did not touch custom shield cooldowns");
        require(Integer.valueOf(100).equals(player.trackingCooldown.durationByItem.get(shield)),
                "disableShield(false) did not apply exactly 100 ticks to the registered custom shield");

        var usedStat = Stats.USED.getOrCreateStat(shield);
        ItemStack damaged = new ItemStack(shield);
        player.setStackInHand(Hand.OFF_HAND, damaged);
        player.setCurrentHand(Hand.OFF_HAND);
        require(player.getActiveItem().getItem() == shield, "custom shield did not become active offhand item");

        int statsBefore = player.getStatHandler().getStat(usedStat);
        player.invokeDamageShield(2.0F);
        int statsAfter = player.getStatHandler().getStat(usedStat);
        require(statsAfter == statsBefore + 1,
                "damageShield <3 must increment USED exactly once; before=" + statsBefore + " after=" + statsAfter);
        require(damaged.getDamage() == 0, "damageShield <3 damaged the shield");

        statsBefore = statsAfter;
        player.invokeDamageShield(3.0F);
        statsAfter = player.getStatHandler().getStat(usedStat);
        require(statsAfter == statsBefore + 1,
                "damageShield >=3 must increment USED exactly once; before=" + statsBefore + " after=" + statsAfter);
        require(damaged.getDamage() == 4,
                "damageShield >=3 must apply exactly 1+floor(amount)=4 durability once; actual=" + damaged.getDamage());
        require(!player.getStackInHand(Hand.OFF_HAND).isEmpty(), "non-breaking shield damage unexpectedly cleared offhand");
        require(!player.getActiveItem().isEmpty(), "non-breaking shield damage unexpectedly cleared active item");

        // Prove the exact break/clear path separately without hiding double-damage behind an immediate break.
        player.clearActiveItem();
        ItemStack breaking = new ItemStack(shield);
        breaking.setDamage(4);
        player.setStackInHand(Hand.OFF_HAND, breaking);
        player.setCurrentHand(Hand.OFF_HAND);
        require(player.getActiveItem().getItem() == shield, "pre-damaged custom shield did not become active offhand item");
        statsBefore = statsAfter;
        player.invokeDamageShield(3.0F);
        statsAfter = player.getStatHandler().getStat(usedStat);
        require(statsAfter == statsBefore + 1,
                "breaking damageShield must increment USED exactly once; before=" + statsBefore + " after=" + statsAfter);
        require(player.getStackInHand(Hand.OFF_HAND).isEmpty(), "breaking custom shield did not clear offhand");
        require(player.getActiveItem().isEmpty(), "breaking custom shield did not clear active item");
    }

    private static void require(boolean condition, String message) {
        if (!condition) throw new IllegalStateException(message);
    }

    private static final class TrackingCooldownManager extends ItemCooldownManager {
        int calls;
        final Map<Item, Integer> durationByItem = new IdentityHashMap<>();

        @Override
        public void set(Item item, int duration) {
            calls++;
            durationByItem.put(item, duration);
            super.set(item, duration);
        }
    }

    private static final class TestServerPlayer extends ServerPlayerEntity {
        private TrackingCooldownManager trackingCooldown;

        TestServerPlayer(MinecraftServer server, ServerWorld world, GameProfile profile) {
            super(server, world, profile);
            require(this.trackingCooldown != null, "tracking cooldown manager was not created");
        }

        @Override
        protected ItemCooldownManager createCooldownManager() {
            this.trackingCooldown = new TrackingCooldownManager();
            return this.trackingCooldown;
        }

        void invokeDamageShield(float amount) {
            this.damageShield(amount);
        }
    }
}
