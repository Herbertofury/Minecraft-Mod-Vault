package com.github.theredbrain.bundleapi.forge;

import com.github.theredbrain.bundleapi.BundleAPI;
import com.github.theredbrain.bundleapi.component.type.CustomBundleContentsComponent;
import com.github.theredbrain.bundleapi.item.CustomBundleItem;
import net.minecraft.item.Item;
import net.minecraft.item.ItemStack;
import net.minecraft.item.Items;
import net.minecraft.nbt.NbtCompound;
import net.minecraft.nbt.NbtList;
import org.apache.commons.lang3.math.Fraction;

final class BundleAPISelfTest {
    private BundleAPISelfTest() { }

    static void run() {
        capacityAndStackSplitting();
        roundTripAndMigration();
        specialOccupancy();
        BundleAPI.LOGGER.info("BUNDLE_API_SELF_TEST_PASS");
    }

    private static void require(boolean condition, String message) {
        if (!condition) throw new IllegalStateException("BUNDLE_API_SELF_TEST_FAILED: " + message);
    }

    private static void capacityAndStackSplitting() {
        CustomBundleContentsComponent.Builder builder = CustomBundleContentsComponent.builder().size_multiplier(4);
        ItemStack arrows = new ItemStack(Items.ARROW);
        arrows.setCount(300);
        require(builder.add(arrows) == 256, "4x bundle must accept exactly 256 ordinary arrows");
        require(arrows.getCount() == 44, "accepted arrows must be decremented from source");
        CustomBundleContentsComponent contents = builder.build();
        require(contents.getOccupancy().equals(Fraction.ONE), "256 arrows at 4x capacity must be full");
        require(contents.size() == 4, "accepted 256 arrows must remain four legal 64-count stacks");
        for (ItemStack stack : contents.iterate()) require(stack.getCount() <= stack.getMaxCount(), "stored stack exceeded vanilla max count");
        require(builder.add(new ItemStack(Items.ARROW, 1)) == 0, "full bundle accepted another arrow");
        ItemStack removed = builder.removeFirst();
        require(removed != null && removed.getCount() == 64, "removeFirst must return newest/front stack");
    }

    private static void roundTripAndMigration() {
        ItemStack owner = new ItemStack(Items.BUNDLE);
        owner.getOrCreateNbt().putString("UnrelatedMarker", "preserve-me");
        CustomBundleContentsComponent.Builder builder = CustomBundleContentsComponent.builder().size_multiplier(8);
        builder.add(new ItemStack(Items.ARROW, 64));
        builder.add(new ItemStack(Items.SPECTRAL_ARROW, 32));
        BundleAPI.setContents(owner, builder.build());
        CustomBundleContentsComponent decoded = BundleAPI.getContents(owner, 1);
        require(decoded.sizeMultiplier() == 8, "canonical size multiplier did not round-trip");
        require(decoded.size() == 2, "canonical contents did not round-trip");
        require("preserve-me".equals(owner.getNbt().getString("UnrelatedMarker")), "unrelated root NBT was not preserved");

        ItemStack vanilla = new ItemStack(Items.BUNDLE);
        NbtList items = new NbtList();
        items.add(new ItemStack(Items.ARROW, 16).writeNbt(new NbtCompound()));
        vanilla.getOrCreateNbt().put("Items", items);
        vanilla.getOrCreateNbt().putInt("ForeignValue", 73);
        CustomBundleContentsComponent migrated = BundleAPI.getContents(vanilla, 12);
        require(migrated.sizeMultiplier() == 12 && migrated.size() == 1, "vanilla Items migration failed");
        BundleAPI.setContents(vanilla, migrated);
        require(BundleAPI.hasCustomData(vanilla), "mutation did not canonicalize custom BundleAPI data");
        require(vanilla.getNbt().getInt("ForeignValue") == 73, "migration damaged unrelated NBT");

        ItemStack malformed = new ItemStack(Items.BUNDLE);
        NbtList malformedItems = new NbtList();
        malformedItems.add(new NbtCompound());
        malformed.getOrCreateNbt().put("Items", malformedItems);
        require(BundleAPI.getContents(malformed, 4).isEmpty(), "malformed empty item entry was not skipped safely");
    }

    private static void specialOccupancy() {
        ItemStack hive = new ItemStack(Items.BEEHIVE);
        NbtCompound blockEntity = new NbtCompound();
        NbtList bees = new NbtList();
        bees.add(new NbtCompound());
        blockEntity.put("Bees", bees);
        hive.getOrCreateNbt().put("BlockEntityTag", blockEntity);
        require(CustomBundleContentsComponent.getOccupancy(hive, 12).equals(Fraction.ONE), "occupied beehive must consume full bundle occupancy");

        CustomBundleItem nestedItem = new CustomBundleItem(null, 1, new Item.Settings().maxCount(1));
        ItemStack nested = new ItemStack(nestedItem);
        CustomBundleContentsComponent.Builder nestedBuilder = CustomBundleContentsComponent.builder();
        nestedBuilder.add(new ItemStack(Items.ARROW, 1));
        BundleAPI.setContents(nested, nestedBuilder.build());
        Fraction expected = Fraction.getFraction(5, 64);
        require(CustomBundleContentsComponent.getOccupancy(nested, 1).equals(expected), "nested bundle occupancy must be 1/16 plus nested contents");
    }
}
