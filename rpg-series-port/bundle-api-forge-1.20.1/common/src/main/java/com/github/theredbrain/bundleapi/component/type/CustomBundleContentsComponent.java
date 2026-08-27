package com.github.theredbrain.bundleapi.component.type;

import com.github.theredbrain.bundleapi.BundleAPI;
import net.minecraft.block.entity.BeehiveBlockEntity;
import net.minecraft.entity.player.PlayerEntity;
import net.minecraft.item.ItemStack;
import net.minecraft.nbt.NbtElement;
import net.minecraft.screen.slot.Slot;
import org.apache.commons.lang3.math.Fraction;
import org.jetbrains.annotations.Nullable;

import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.stream.Stream;

/** Immutable current Bundle API contents model backed by native 1.20.1 NBT. */
public final class CustomBundleContentsComponent {
    private static final Fraction NESTED_BUNDLE_OCCUPANCY = Fraction.getFraction(1, 16);
    public static final CustomBundleContentsComponent DEFAULT = new CustomBundleContentsComponent(List.of(), 1);

    private final List<ItemStack> stacks;
    private final Fraction occupancy;
    private final int sizeMultiplier;

    public CustomBundleContentsComponent(int sizeMultiplier) {
        this(List.of(), sizeMultiplier);
    }

    private CustomBundleContentsComponent(List<ItemStack> stacks, int sizeMultiplier) {
        this.sizeMultiplier = Math.max(1, sizeMultiplier);
        List<ItemStack> copies = new ArrayList<>();
        for (ItemStack stack : stacks) copies.add(stack.copy());
        this.stacks = Collections.unmodifiableList(copies);
        this.occupancy = calculateOccupancy(copies, this.sizeMultiplier);
    }

    public static Builder builder() { return new Builder(DEFAULT); }
    public ItemStack get(int index) { return stacks.get(index); }
    public Stream<ItemStack> stream() { return stacks.stream().map(ItemStack::copy); }
    public Iterable<ItemStack> iterate() { return stacks; }
    public Iterable<ItemStack> iterateCopy() { return () -> stacks.stream().map(ItemStack::copy).iterator(); }
    public int size() { return stacks.size(); }
    public int sizeMultiplier() { return sizeMultiplier; }
    public Fraction getOccupancy() { return occupancy; }
    public boolean isEmpty() { return stacks.isEmpty(); }

    private static Fraction calculateOccupancy(List<ItemStack> stacks, int sizeMultiplier) {
        Fraction result = Fraction.ZERO;
        for (ItemStack stack : stacks) {
            result = result.add(getOccupancy(stack, sizeMultiplier).multiplyBy(Fraction.getFraction(stack.getCount(), 1)));
        }
        return result;
    }

    public static Fraction getOccupancy(ItemStack stack, int sizeMultiplier) {
        if (stack.getItem() instanceof com.github.theredbrain.bundleapi.item.CustomBundleItem) {
            CustomBundleContentsComponent nested = BundleAPI.getContents(stack);
            return NESTED_BUNDLE_OCCUPANCY.add(nested.getOccupancy());
        }
        if (stack.hasNbt()) {
            var blockEntity = stack.getSubNbt("BlockEntityTag");
            if (blockEntity != null && blockEntity.contains("Bees", NbtElement.LIST_TYPE)
                    && !blockEntity.getList("Bees", NbtElement.COMPOUND_TYPE).isEmpty()) {
                return Fraction.ONE;
            }
        }
        return Fraction.getFraction(1, Math.max(1, stack.getMaxCount() * Math.max(1, sizeMultiplier)));
    }

    public static final class Builder {
        private final List<ItemStack> stacks = new ArrayList<>();
        private int sizeMultiplier;

        public Builder(CustomBundleContentsComponent base) {
            base.iterateCopy().forEach(stacks::add);
            sizeMultiplier = base.sizeMultiplier;
        }

        public Builder clear() {
            stacks.clear();
            return this;
        }

        public Builder size_multiplier(int sizeMultiplier) {
            this.sizeMultiplier = Math.max(1, sizeMultiplier);
            return this;
        }

        private Fraction occupancy() { return calculateOccupancy(stacks, sizeMultiplier); }

        private int getMaxAllowed(ItemStack stack) {
            Fraction remaining = Fraction.ONE.subtract(occupancy());
            Fraction each = getOccupancy(stack, sizeMultiplier);
            if (each.compareTo(Fraction.ZERO) <= 0 || remaining.compareTo(Fraction.ZERO) <= 0) return 0;
            return Math.max(0, remaining.divideBy(each).intValue());
        }

        private int mergeIndex(ItemStack stack) {
            if (!stack.isStackable()) return -1;
            for (int i = 0; i < stacks.size(); i++) {
                ItemStack existing = stacks.get(i);
                if (ItemStack.canCombine(existing, stack) && existing.getCount() < existing.getMaxCount()) return i;
            }
            return -1;
        }

        public int add(ItemStack stack) { return addInternal(stack, true); }

        /** Migration/storage replay without consuming the supplied decoded stack. */
        public int addCopy(ItemStack stack) { return addInternal(stack.copy(), true); }

        private int addInternal(ItemStack stack, boolean consume) {
            if (stack.isEmpty() || !stack.getItem().canBeNested()) return 0;
            int accepted = Math.min(stack.getCount(), getMaxAllowed(stack));
            if (accepted <= 0) return 0;

            int merge = mergeIndex(stack);
            if (merge >= 0) {
                ItemStack existing = stacks.remove(merge);
                int room = existing.getMaxCount() - existing.getCount();
                int merged = Math.min(room, accepted);
                ItemStack joined = existing.copy();
                joined.increment(merged);
                stacks.add(0, joined);
                int remainder = accepted - merged;
                if (remainder > 0) {
                    ItemStack extra = stack.copy();
                    extra.setCount(remainder);
                    stacks.add(0, extra);
                }
            } else {
                ItemStack inserted = stack.copy();
                inserted.setCount(accepted);
                stacks.add(0, inserted);
            }
            if (consume) stack.decrement(accepted);
            return accepted;
        }

        public int add(Slot slot, PlayerEntity player) {
            ItemStack source = slot.getStack();
            int max = getMaxAllowed(source);
            if (max <= 0) return 0;
            ItemStack removed = slot.takeStackRange(source.getCount(), max, player);
            int inserted = add(removed);
            if (!removed.isEmpty()) slot.insertStack(removed);
            return inserted;
        }

        @Nullable
        public ItemStack removeFirst() {
            return stacks.isEmpty() ? null : stacks.remove(0).copy();
        }

        public Fraction getOccupancy() { return occupancy(); }
        public CustomBundleContentsComponent build() { return new CustomBundleContentsComponent(stacks, sizeMultiplier); }
    }
}
