package net.archers.validation;

import com.github.theredbrain.bundleapi.BundleAPI;
import com.github.theredbrain.bundleapi.component.type.CustomBundleContentsComponent;
import com.github.theredbrain.bundleapi.item.CustomBundleItem;
import net.minecraft.item.Item;
import net.minecraft.item.ItemStack;
import net.minecraft.item.Items;
import net.minecraft.registry.tag.ItemTags;
import net.minecraft.util.Rarity;
import net.minecraftforge.common.MinecraftForge;
import net.minecraftforge.event.server.ServerStartedEvent;
import net.minecraftforge.eventbus.api.IEventBus;
import net.minecraftforge.fml.common.Mod;
import net.minecraftforge.fml.javafmlmod.FMLJavaModLoadingContext;
import net.minecraftforge.registries.DeferredRegister;
import net.minecraftforge.registries.ForgeRegistries;
import net.minecraftforge.registries.RegistryObject;
import org.apache.commons.lang3.math.Fraction;

import java.lang.reflect.Method;

/**
 * Exact Bundle API consumer boundary derived from Archers 3.1.1 Quivers.java.
 * The Bundle API implementation is supplied only as an external packaged JAR.
 */
@Mod(ArchersQuiverConsumerMod.MOD_ID)
public final class ArchersQuiverConsumerMod {
    public static final String MOD_ID = "archersquiverconsumer";
    private static final DeferredRegister<Item> ITEMS = DeferredRegister.create(ForgeRegistries.ITEMS, MOD_ID);

    private static final RegistryObject<Item> SMALL = ITEMS.register("small_quiver",
            () -> new CustomBundleItem(ItemTags.ARROWS, 4, new Item.Settings().maxCount(1)));
    private static final RegistryObject<Item> MEDIUM = ITEMS.register("medium_quiver",
            () -> new CustomBundleItem(ItemTags.ARROWS, 8, new Item.Settings().maxCount(1)));
    private static final RegistryObject<Item> LARGE = ITEMS.register("large_quiver",
            () -> new CustomBundleItem(ItemTags.ARROWS, 12, new Item.Settings().maxCount(1).rarity(Rarity.UNCOMMON)));

    public ArchersQuiverConsumerMod() {
        IEventBus modBus = FMLJavaModLoadingContext.get().getModEventBus();
        ITEMS.register(modBus);
        MinecraftForge.EVENT_BUS.addListener(this::onServerStarted);
    }

    private void onServerStarted(ServerStartedEvent event) {
        try {
            verifyRegisteredQuiver("small", SMALL.get(), 4, 256, false);
            verifyRegisteredQuiver("medium", MEDIUM.get(), 8, 512, false);
            verifyRegisteredQuiver("large", LARGE.get(), 12, 768, true);
            verifyLoadedArrowTagAndFilter((CustomBundleItem) SMALL.get());
            System.out.println("ARCHERS_QUIVER_CONSUMER_PASS");
        } catch (Throwable failure) {
            System.err.println("ARCHERS_QUIVER_CONSUMER_FAILED: " + failure);
            failure.printStackTrace();
            throw new RuntimeException("Archers quiver consumer acceptance failed", failure);
        }
    }

    private static void verifyRegisteredQuiver(String name, Item rawItem, int multiplier, int expectedCapacity,
                                                boolean expectUncommon) {
        require(rawItem instanceof CustomBundleItem, name + " is not a CustomBundleItem");
        CustomBundleItem item = (CustomBundleItem) rawItem;
        require(item.defaultSizeMultiplier() == multiplier,
                name + " multiplier expected " + multiplier + " but was " + item.defaultSizeMultiplier());
        require(rawItem.getMaxCount() == 1, name + " max stack count must be 1");

        ItemStack quiver = new ItemStack(rawItem);
        if (expectUncommon) {
            require(quiver.getRarity() == Rarity.UNCOMMON, "large quiver rarity must be UNCOMMON");
        }

        CustomBundleContentsComponent initial = BundleAPI.getContents(quiver, item.defaultSizeMultiplier());
        require(initial.sizeMultiplier() == multiplier, name + " NBT/default multiplier drifted");
        CustomBundleContentsComponent.Builder builder = new CustomBundleContentsComponent.Builder(initial);

        int insertedTotal = 0;
        while (insertedTotal < expectedCapacity) {
            ItemStack arrows = new ItemStack(Items.ARROW, Math.min(64, expectedCapacity - insertedTotal));
            int before = arrows.getCount();
            int inserted = builder.add(arrows);
            require(inserted == before, name + " rejected arrows before capacity at " + insertedTotal);
            require(arrows.isEmpty(), name + " did not consume accepted input stack");
            insertedTotal += inserted;
        }
        require(insertedTotal == expectedCapacity, name + " capacity mismatch: " + insertedTotal);
        require(builder.getOccupancy().equals(Fraction.ONE), name + " occupancy must be exactly 1 at capacity");

        ItemStack overflow = new ItemStack(Items.ARROW, 1);
        require(builder.add(overflow) == 0, name + " accepted an arrow past full capacity");
        require(overflow.getCount() == 1, name + " mutated rejected overflow input");

        CustomBundleContentsComponent full = builder.build();
        require(full.getOccupancy().equals(Fraction.ONE), name + " built occupancy must remain exactly 1");
        require(full.size() == expectedCapacity / 64,
                name + " must store legal 64-item chunks; expected " + (expectedCapacity / 64) + " stacks but got " + full.size());
        for (ItemStack stored : full.iterate()) {
            require(stored.getCount() > 0 && stored.getCount() <= stored.getMaxCount(),
                    name + " contains illegal stored stack count " + stored.getCount());
        }

        CustomBundleContentsComponent.Builder removal = new CustomBundleContentsComponent.Builder(full);
        ItemStack removed = removal.removeFirst();
        require(removed != null && removed.isOf(Items.ARROW) && removed.getCount() == 64,
                name + " removeFirst must return the newest/front legal arrow stack");
    }

    private static void verifyLoadedArrowTagAndFilter(CustomBundleItem quiver) throws Exception {
        ItemStack normal = new ItemStack(Items.ARROW);
        ItemStack spectral = new ItemStack(Items.SPECTRAL_ARROW);
        ItemStack nonArrow = new ItemStack(Items.STONE);
        require(normal.isIn(ItemTags.ARROWS), "loaded ItemTags.ARROWS is missing normal arrows");
        require(spectral.isIn(ItemTags.ARROWS), "loaded ItemTags.ARROWS is missing spectral arrows");
        require(!nonArrow.isIn(ItemTags.ARROWS), "loaded ItemTags.ARROWS unexpectedly contains stone");

        Method accepts = CustomBundleItem.class.getDeclaredMethod("accepts", ItemStack.class);
        accepts.setAccessible(true);
        require((boolean) accepts.invoke(quiver, normal), "quiver filter rejected normal arrow");
        require((boolean) accepts.invoke(quiver, spectral), "quiver filter rejected spectral arrow");
        require(!(boolean) accepts.invoke(quiver, nonArrow), "quiver filter accepted non-arrow item");
    }

    private static void require(boolean condition, String message) {
        if (!condition) throw new IllegalStateException(message);
    }
}
