package com.github.theredbrain.bundleapi.item;

import com.github.theredbrain.bundleapi.BundleAPI;
import com.github.theredbrain.bundleapi.component.type.CustomBundleContentsComponent;
import net.minecraft.entity.Entity;
import net.minecraft.entity.ItemEntity;
import net.minecraft.entity.player.PlayerEntity;
import net.minecraft.inventory.StackReference;
import net.minecraft.item.Item;
import net.minecraft.item.ItemStack;
import net.minecraft.item.ItemUsage;
import net.minecraft.registry.tag.TagKey;
import net.minecraft.screen.slot.Slot;
import net.minecraft.server.network.ServerPlayerEntity;
import net.minecraft.sound.SoundEvents;
import net.minecraft.stat.Stats;
import net.minecraft.text.Text;
import net.minecraft.util.ClickType;
import net.minecraft.util.Formatting;
import net.minecraft.util.Hand;
import net.minecraft.util.TypedActionResult;
import net.minecraft.util.math.MathHelper;
import net.minecraft.world.World;
import org.apache.commons.lang3.math.Fraction;
import org.jetbrains.annotations.Nullable;

import java.util.HashSet;
import java.util.List;

/** Bundle API 1.1.0 item semantics adapted from 1.21 data components to 1.20.1 NBT. */
public class CustomBundleItem extends Item {
    private static final int ITEM_BAR_COLOR = MathHelper.packRgb(0.4F, 0.4F, 1.0F);
    public static final HashSet<CustomBundleItem> instances = new HashSet<>();

    @Nullable
    private final TagKey<Item> tag;
    private final int defaultSizeMultiplier;

    public CustomBundleItem(@Nullable TagKey<Item> tag, Settings settings) {
        this(tag, 1, settings);
    }

    /** 1.20.1 extension used when a 1.21 Data Component supplied per-item capacity. */
    public CustomBundleItem(@Nullable TagKey<Item> tag, int sizeMultiplier, Settings settings) {
        super(settings);
        this.tag = tag;
        this.defaultSizeMultiplier = Math.max(1, sizeMultiplier);
        instances.add(this);
    }

    public CustomBundleItem(Settings settings) {
        this(null, 1, settings);
    }

    public int defaultSizeMultiplier() {
        return defaultSizeMultiplier;
    }

    private CustomBundleContentsComponent contents(ItemStack stack) {
        return BundleAPI.getContents(stack, defaultSizeMultiplier);
    }

    private void store(ItemStack stack, CustomBundleContentsComponent contents) {
        BundleAPI.setContents(stack, contents);
    }

    private boolean accepts(ItemStack stack) {
        return !stack.isEmpty() && stack.getItem().canBeNested() && (tag == null || stack.isIn(tag));
    }

    public static float getAmountFilled(ItemStack stack) {
        if (stack.getItem() instanceof CustomBundleItem item) {
            return item.contents(stack).getOccupancy().floatValue();
        }
        return BundleAPI.getContents(stack).getOccupancy().floatValue();
    }

    @Override
    public boolean onStackClicked(ItemStack stack, Slot slot, ClickType clickType, PlayerEntity player) {
        if (clickType != ClickType.RIGHT) return false;

        CustomBundleContentsComponent.Builder builder = new CustomBundleContentsComponent.Builder(contents(stack));
        ItemStack slotStack = slot.getStack();
        if (slotStack.isEmpty()) {
            ItemStack removed = builder.removeFirst();
            if (removed != null) {
                playRemoveOneSound(player);
                ItemStack remainder = slot.insertStack(removed);
                if (!remainder.isEmpty()) builder.add(remainder);
            }
        } else if (accepts(slotStack)) {
            int inserted = builder.add(slot, player);
            if (inserted > 0) playInsertSound(player);
        }

        store(stack, builder.build());
        return true;
    }

    @Override
    public boolean onClicked(ItemStack stack, ItemStack otherStack, Slot slot, ClickType clickType,
                             PlayerEntity player, StackReference cursorStackReference) {
        if (clickType != ClickType.RIGHT || !slot.canTakePartial(player)) return false;

        CustomBundleContentsComponent.Builder builder = new CustomBundleContentsComponent.Builder(contents(stack));
        if (otherStack.isEmpty()) {
            ItemStack removed = builder.removeFirst();
            if (removed != null) {
                playRemoveOneSound(player);
                cursorStackReference.set(removed);
            }
        } else if (accepts(otherStack)) {
            int inserted = builder.add(otherStack);
            if (inserted > 0) playInsertSound(player);
        }

        store(stack, builder.build());
        return true;
    }

    @Override
    public TypedActionResult<ItemStack> use(World world, PlayerEntity user, Hand hand) {
        ItemStack stack = user.getStackInHand(hand);
        if (dropAllBundledItems(stack, user)) {
            playDropContentsSound(user);
            user.incrementStat(Stats.USED.getOrCreateStat(this));
            return TypedActionResult.success(stack, world.isClient());
        }
        return TypedActionResult.fail(stack);
    }

    @Override
    public boolean isItemBarVisible(ItemStack stack) {
        return contents(stack).getOccupancy().compareTo(Fraction.ZERO) > 0;
    }

    @Override
    public int getItemBarStep(ItemStack stack) {
        Fraction occupancy = contents(stack).getOccupancy();
        return Math.min(1 + occupancy.multiplyBy(Fraction.getFraction(12, 1)).intValue(), 13);
    }

    @Override
    public int getItemBarColor(ItemStack stack) {
        return ITEM_BAR_COLOR;
    }

    @Override
    public void appendTooltip(ItemStack stack, @Nullable World world, List<Text> tooltip, TooltipContext context) {
        CustomBundleContentsComponent contents = contents(stack);
        int bundleMaxSize = contents.sizeMultiplier() * 64;
        int used = contents.getOccupancy().multiplyBy(Fraction.getFraction(bundleMaxSize, 1)).intValue();
        tooltip.add(Text.translatable("item.minecraft.bundle.fullness", used, bundleMaxSize).formatted(Formatting.GRAY));
    }

    private boolean dropAllBundledItems(ItemStack stack, PlayerEntity player) {
        CustomBundleContentsComponent contents = contents(stack);
        if (contents.isEmpty()) return false;

        store(stack, new CustomBundleContentsComponent.Builder(contents).clear().build());
        if (player instanceof ServerPlayerEntity) {
            contents.iterateCopy().forEach(item -> player.dropItem(item, true));
        }
        return true;
    }

    @Override
    public void onItemEntityDestroyed(ItemEntity entity) {
        CustomBundleContentsComponent contents = contents(entity.getStack());
        if (!contents.isEmpty()) {
            store(entity.getStack(), new CustomBundleContentsComponent.Builder(contents).clear().build());
            ItemUsage.spawnItemContents(entity, contents.iterateCopy());
        }
    }

    private void playRemoveOneSound(Entity entity) {
        entity.playSound(SoundEvents.ITEM_BUNDLE_REMOVE_ONE, 0.8F, 0.8F + entity.getWorld().getRandom().nextFloat() * 0.4F);
    }

    private void playInsertSound(Entity entity) {
        entity.playSound(SoundEvents.ITEM_BUNDLE_INSERT, 0.8F, 0.8F + entity.getWorld().getRandom().nextFloat() * 0.4F);
    }

    private void playDropContentsSound(Entity entity) {
        entity.playSound(SoundEvents.ITEM_BUNDLE_DROP_CONTENTS, 0.8F, 0.8F + entity.getWorld().getRandom().nextFloat() * 0.4F);
    }
}