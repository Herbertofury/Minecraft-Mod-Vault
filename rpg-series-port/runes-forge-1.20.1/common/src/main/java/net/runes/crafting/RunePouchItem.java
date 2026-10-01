package net.runes.crafting;

import net.minecraft.client.item.TooltipContext;
import net.minecraft.entity.player.PlayerEntity;
import net.minecraft.inventory.StackReference;
import net.minecraft.item.Item;
import net.minecraft.item.ItemStack;
import net.minecraft.nbt.NbtCompound;
import net.minecraft.nbt.NbtElement;
import net.minecraft.nbt.NbtList;
import net.minecraft.registry.Registries;
import net.minecraft.registry.tag.TagKey;
import net.minecraft.screen.slot.Slot;
import net.minecraft.text.Text;
import net.minecraft.util.ClickType;
import net.minecraft.util.Formatting;
import net.minecraft.util.Hand;
import net.minecraft.util.Identifier;
import net.minecraft.util.TypedActionResult;
import net.minecraft.world.World;
import net.runes.RunesMod;
import org.jetbrains.annotations.Nullable;

import java.util.ArrayList;
import java.util.Collections;
import java.util.List;

/**
 * Native 1.20.1 implementation of the 1.3.x Rune Pouch behavior.
 *
 * The 1.21 build uses Bundle API data components, which do not exist on 1.20.1. This implementation
 * deliberately uses vanilla-style item NBT under {@code Items}; it keeps the same public item IDs,
 * rune-only filtering and 4/8/12-stack capacities while removing the Bundle API runtime dependency.
 */
public class RunePouchItem extends Item {
    public static final String ITEMS_KEY = "Items";
    private static final TagKey<Item> RUNES = TagKey.of(
            Registries.ITEM.getKey(), new Identifier(RunesMod.ID, "runes"));

    private final int capacityStacks;

    public RunePouchItem(int capacityStacks, Settings settings) {
        super(settings);
        this.capacityStacks = capacityStacks;
    }

    public int capacityStacks() { return capacityStacks; }
    public int capacityItems() { return capacityStacks * 64; }

    public static boolean isRune(ItemStack stack) {
        return !stack.isEmpty() && stack.isIn(RUNES) && !(stack.getItem() instanceof RunePouchItem);
    }

    public static boolean isPouch(ItemStack stack) {
        return !stack.isEmpty() && stack.getItem() instanceof RunePouchItem;
    }

    /** Immutable snapshot used by compatibility integrations such as the later Spell Engine port. */
    public static List<ItemStack> contents(ItemStack pouch) {
        if (!isPouch(pouch)) return List.of();
        NbtCompound nbt = pouch.getNbt();
        if (nbt == null || !nbt.contains(ITEMS_KEY, NbtElement.LIST_TYPE)) return List.of();

        List<ItemStack> out = new ArrayList<>();
        NbtList list = nbt.getList(ITEMS_KEY, NbtElement.COMPOUND_TYPE);
        for (int i = 0; i < list.size(); i++) {
            ItemStack stack = ItemStack.fromNbt(list.getCompound(i));
            if (isRune(stack)) out.add(stack);
        }
        return Collections.unmodifiableList(out);
    }

    private static List<ItemStack> mutableContents(ItemStack pouch) {
        List<ItemStack> copy = new ArrayList<>();
        for (ItemStack stack : contents(pouch)) copy.add(stack.copy());
        return copy;
    }

    private static void save(ItemStack pouch, List<ItemStack> stacks) {
        NbtList list = new NbtList();
        for (ItemStack stack : stacks) {
            if (!isRune(stack) || stack.isEmpty()) continue;
            list.add(stack.writeNbt(new NbtCompound()));
        }
        if (list.isEmpty()) {
            NbtCompound nbt = pouch.getNbt();
            if (nbt != null) {
                nbt.remove(ITEMS_KEY);
                if (nbt.isEmpty()) pouch.setNbt(null);
            }
        } else {
            pouch.getOrCreateNbt().put(ITEMS_KEY, list);
        }
    }

    public static int count(ItemStack pouch) {
        return contents(pouch).stream().mapToInt(ItemStack::getCount).sum();
    }

    public static boolean hasContents(ItemStack pouch) { return count(pouch) > 0; }

    /** Inserts as many runes as possible and mutates the supplied source stack like vanilla bundles do. */
    public int insertRunes(ItemStack pouch, ItemStack incoming) {
        if (!isRune(incoming)) return 0;
        List<ItemStack> list = mutableContents(pouch);
        int room = Math.max(0, capacityItems() - list.stream().mapToInt(ItemStack::getCount).sum());
        if (room == 0) return 0;

        int toMove = Math.min(room, incoming.getCount());
        int remaining = toMove;
        for (ItemStack existing : list) {
            if (ItemStack.canCombine(existing, incoming) && existing.getCount() < existing.getMaxCount()) {
                int move = Math.min(remaining, existing.getMaxCount() - existing.getCount());
                existing.increment(move);
                remaining -= move;
                if (remaining == 0) break;
            }
        }
        while (remaining > 0) {
            int move = Math.min(remaining, incoming.getMaxCount());
            ItemStack copy = incoming.copy();
            copy.setCount(move);
            list.add(copy);
            remaining -= move;
        }
        incoming.decrement(toMove);
        save(pouch, list);
        return toMove;
    }

    /** Extracts up to {@code maximum} items from the most recently added rune stack. */
    public ItemStack extractLast(ItemStack pouch, int maximum) {
        if (maximum <= 0) return ItemStack.EMPTY;
        List<ItemStack> list = mutableContents(pouch);
        if (list.isEmpty()) return ItemStack.EMPTY;

        ItemStack stored = list.get(list.size() - 1);
        int amount = Math.min(maximum, stored.getCount());
        ItemStack out = stored.copy();
        out.setCount(amount);
        stored.decrement(amount);
        if (stored.isEmpty()) list.remove(list.size() - 1);
        save(pouch, list);
        return out;
    }

    @Override
    public boolean onClicked(ItemStack pouch, ItemStack cursor, Slot slot, ClickType click,
                             PlayerEntity player, StackReference cursorRef) {
        if (click == ClickType.LEFT && isRune(cursor)) {
            if (insertRunes(pouch, cursor) > 0) {
                cursorRef.set(cursor);
                return true;
            }
        }
        if (click == ClickType.RIGHT && cursor.isEmpty()) {
            ItemStack removed = extractLast(pouch, 64);
            if (!removed.isEmpty()) {
                cursorRef.set(removed);
                return true;
            }
        }
        return false;
    }

    @Override
    public boolean onStackClicked(ItemStack pouch, Slot slot, ClickType click, PlayerEntity player) {
        if (click == ClickType.LEFT && slot.hasStack() && isRune(slot.getStack())) {
            if (insertRunes(pouch, slot.getStack()) > 0) {
                slot.markDirty();
                return true;
            }
        }
        if (click == ClickType.RIGHT && !slot.hasStack()) {
            ItemStack removed = extractLast(pouch, 64);
            if (!removed.isEmpty()) {
                slot.setStack(removed);
                return true;
            }
        }
        return false;
    }

    @Override
    public TypedActionResult<ItemStack> use(World world, PlayerEntity user, Hand hand) {
        ItemStack pouch = user.getStackInHand(hand);
        if (world.isClient) return TypedActionResult.success(pouch, true);
        ItemStack removed = extractLast(pouch, 64);
        if (removed.isEmpty()) return TypedActionResult.pass(pouch);
        if (!user.getInventory().insertStack(removed)) user.dropItem(removed, false);
        return TypedActionResult.success(pouch, false);
    }

    @Override public boolean isItemBarVisible(ItemStack stack) { return hasContents(stack); }

    @Override
    public int getItemBarStep(ItemStack stack) {
        int capacity = Math.max(1, capacityItems());
        return Math.min(13, Math.round(13.0F * count(stack) / capacity));
    }

    @Override public int getItemBarColor(ItemStack stack) { return 0x6d42d8; }

    @Override
    public void appendTooltip(ItemStack stack, @Nullable World world, List<Text> tooltip, TooltipContext context) {
        super.appendTooltip(stack, world, tooltip, context);
        tooltip.add(Text.translatable("item.runes.rune_pouch.hint").formatted(Formatting.GRAY));
        tooltip.add(Text.literal(count(stack) + " / " + capacityItems()).formatted(Formatting.DARK_GRAY));

        List<ItemStack> stored = contents(stack);
        int shown = 0;
        for (int i = stored.size() - 1; i >= 0 && shown < 4; i--, shown++) {
            ItemStack content = stored.get(i);
            tooltip.add(Text.literal("  ").append(content.getName()).append(" x" + content.getCount())
                    .formatted(Formatting.DARK_GRAY));
        }
        if (stored.size() > shown) {
            tooltip.add(Text.literal("  +" + (stored.size() - shown) + " more").formatted(Formatting.DARK_GRAY));
        }
    }
}
