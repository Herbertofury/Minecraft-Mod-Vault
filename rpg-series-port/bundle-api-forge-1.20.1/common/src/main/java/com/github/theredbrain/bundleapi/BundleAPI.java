package com.github.theredbrain.bundleapi;

import com.github.theredbrain.bundleapi.component.type.CustomBundleContentsComponent;
import net.minecraft.item.ItemStack;
import net.minecraft.nbt.NbtCompound;
import net.minecraft.nbt.NbtElement;
import net.minecraft.nbt.NbtList;
import net.minecraft.util.Identifier;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * Forge 1.20.1 compatibility facade for Bundle API 1.1.0.
 *
 * <p>Minecraft 1.20.1 has no 1.21 data-component system, so bundle contents are
 * persisted on the owning ItemStack. The public content object and builder stay
 * source-compatible for current RPG-Series consumers while storage is adapted
 * to the native 1.20.1 ItemStack NBT contract.</p>
 */
public final class BundleAPI {
    public static final String MOD_ID = "bundleapi";
    public static final Logger LOGGER = LoggerFactory.getLogger(MOD_ID);
    public static final String NBT_ROOT = "BundleAPI";
    public static final String NBT_ITEMS = "Items";
    public static final String NBT_SIZE_MULTIPLIER = "SizeMultiplier";

    private BundleAPI() { }

    public static void init() {
        LOGGER.info("Customized Bundles! Native Forge 1.20.1 compatibility active.");
    }

    public static Identifier identifier(String path) {
        return new Identifier(MOD_ID, path);
    }

    public static CustomBundleContentsComponent getContents(ItemStack owner) {
        NbtCompound root = owner.getNbt();
        if (root == null) {
            return CustomBundleContentsComponent.DEFAULT;
        }

        int multiplier = 1;
        NbtList encoded = null;
        if (root.contains(NBT_ROOT, NbtElement.COMPOUND_TYPE)) {
            NbtCompound api = root.getCompound(NBT_ROOT);
            if (api.contains(NBT_SIZE_MULTIPLIER, NbtElement.NUMBER_TYPE)) {
                multiplier = Math.max(1, api.getInt(NBT_SIZE_MULTIPLIER));
            }
            if (api.contains(NBT_ITEMS, NbtElement.LIST_TYPE)) {
                encoded = api.getList(NBT_ITEMS, NbtElement.COMPOUND_TYPE);
            }
        } else if (root.contains(NBT_ITEMS, NbtElement.LIST_TYPE)) {
            // Safe vanilla-1.20.1 bundle interoperability/migration path.
            encoded = root.getList(NBT_ITEMS, NbtElement.COMPOUND_TYPE);
        }

        CustomBundleContentsComponent.Builder builder = CustomBundleContentsComponent.builder().size_multiplier(multiplier);
        if (encoded != null) {
            // Serialized order is already newest/front first. Builder.add() also
            // inserts at the front, so replay in reverse to preserve ordering.
            for (int i = encoded.size() - 1; i >= 0; --i) {
                ItemStack decoded = ItemStack.fromNbt(encoded.getCompound(i));
                if (!decoded.isEmpty()) {
                    builder.addCopy(decoded);
                }
            }
        }
        return builder.build();
    }

    public static void setContents(ItemStack owner, CustomBundleContentsComponent contents) {
        NbtCompound root = owner.getOrCreateNbt();
        NbtCompound api = root.contains(NBT_ROOT, NbtElement.COMPOUND_TYPE)
                ? root.getCompound(NBT_ROOT).copy()
                : new NbtCompound();
        NbtList items = new NbtList();
        for (ItemStack stack : contents.iterate()) {
            if (!stack.isEmpty()) {
                items.add(stack.writeNbt(new NbtCompound()));
            }
        }
        api.put(NBT_ITEMS, items);
        api.putInt(NBT_SIZE_MULTIPLIER, Math.max(1, contents.sizeMultiplier()));
        root.put(NBT_ROOT, api);
    }

    public static void initializeContents(ItemStack owner, int sizeMultiplier) {
        CustomBundleContentsComponent existing = getContents(owner);
        setContents(owner, new CustomBundleContentsComponent.Builder(existing)
                .size_multiplier(Math.max(1, sizeMultiplier))
                .build());
    }
}
