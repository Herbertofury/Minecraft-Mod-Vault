#!/usr/bin/env python3
from pathlib import Path
import sys

if len(sys.argv) != 2:
    raise SystemExit("usage: compat_pass_1.py <generated-port-root>")

root = Path(sys.argv[1]).resolve()
common = root / "common/src/main/java"
forge = root / "forge/src/main/java"

# The 2.4.0 target is intentionally retained in full. This pass only translates API representation
# that changed between 1.21.1 and 1.20.1; catalog/config values and resources stay current.

items_path = common / "net/jewelry/items/JewelryItems.java"
items = items_path.read_text()
items = items.replace("import net.jewelry.JewelryMod;\nimport net.jewelry.config.ItemConfig;\nimport net.minecraft.component.type.AttributeModifierSlot;\nimport net.minecraft.component.type.AttributeModifiersComponent;\n",
                      "import net.jewelry.JewelryMod;\nimport net.jewelry.api.AttributeResolver;\nimport net.jewelry.config.ItemConfig;\n")
items = items.replace("public Item create(Item.Settings settings, AttributeModifiersComponent attributes)",
                      "public Item create(Item.Settings settings, java.util.List<JewelryAttributeModifier> attributes)")
start_marker = "            AttributeModifiersComponent.Builder attributes = AttributeModifiersComponent.builder();\n"
end_marker = "            var settings = new Item.Settings()\n"
if start_marker not in items or end_marker not in items:
    raise SystemExit("JewelryItems 2.4.0 attribute-component anchors missing")
start = items.index(start_marker)
end = items.index(end_marker, start)
bridge = '''            // 1.20.1 has no item attribute-component carrier. Preserve 2.4.0's selected config and\n            // per-item/per-attribute identity in a loader-neutral record consumed by Curios.\n            List<JewelryAttributeModifier> attributes = new ArrayList<>();\n            for (var modifier : itemConfig.selectedAttributes()) {\n                var id = new Identifier(modifier.id);\n                var attribute = AttributeResolver.get(id);\n                if (attribute != null) {\n                    var modifierName = JewelryMod.ID + ":" + entry.id().getPath()\n                            + "_" + id.getPath().replace('.', '_') + "_bonus";\n                    attributes.add(new JewelryAttributeModifier(\n                            attribute, modifierName, modifier.value, modifier.operation));\n                } else {\n                    System.err.println("Failed to resolve EntityAttribute with id: " + modifier.id);\n                }\n            }\n'''
items = items[:start] + bridge + items[end:]
items = items.replace("entry.create(settings.maxCount(1), attributes.build())",
                      "entry.create(settings.maxCount(1), attributes)")
items_path.write_text(items)

(common / "net/jewelry/api").mkdir(parents=True, exist_ok=True)
(common / "net/jewelry/api/AttributeResolver.java").write_text(r'''package net.jewelry.api;

import net.minecraft.entity.attribute.EntityAttribute;
import net.minecraft.registry.Registries;
import net.minecraft.util.Identifier;

import java.util.HashMap;

/** 1.20.1 attribute lookup bridge, retained from the verified Jewelry substrate. */
public final class AttributeResolver {
    private static final HashMap<Identifier, EntityAttribute> attributes = new HashMap<>();

    private AttributeResolver() { }

    public static void register(Identifier id, EntityAttribute attribute) {
        attributes.put(id, attribute);
    }

    public static EntityAttribute get(Identifier id) {
        var attribute = attributes.get(id);
        if (attribute == null) {
            attribute = Registries.ATTRIBUTE.get(id);
        }
        return attribute;
    }
}
''')

(common / "net/jewelry/items/JewelryAttributeModifier.java").write_text(r'''package net.jewelry.items;

import net.minecraft.entity.attribute.EntityAttribute;
import net.minecraft.entity.attribute.EntityAttributeModifier;

/** Loader-neutral 1.20.1 representation of Jewelry 2.4.0 configured item attributes. */
public record JewelryAttributeModifier(
        EntityAttribute attribute,
        String id,
        double value,
        EntityAttributeModifier.Operation operation) {
}
''')

(common / "net/jewelry/items/JewelryFactory.java").write_text(r'''package net.jewelry.items;

import net.minecraft.item.Item;
import org.jetbrains.annotations.Nullable;

import java.util.List;
import java.util.function.Function;

public class JewelryFactory {
    public record ItemArgs(Item.Settings settings, @Nullable List<JewelryAttributeModifier> attributes,
                           @Nullable String lore, @Nullable String slot) { }

    public static Function<ItemArgs, Item> factory = args ->
            new VanillaJewelryItem(args.settings(), args.lore());

    public static Function<ItemArgs, Item> getFactory() {
        return factory;
    }
}
''')

(common / "net/jewelry/items/VanillaJewelryItem.java").write_text(r'''package net.jewelry.items;

import net.minecraft.client.item.TooltipContext;
import net.minecraft.item.Item;
import net.minecraft.item.ItemStack;
import net.minecraft.text.Text;
import net.minecraft.util.Formatting;
import net.minecraft.world.World;
import org.jetbrains.annotations.Nullable;

import java.util.List;

public class VanillaJewelryItem extends Item {
    private final String lore;

    public VanillaJewelryItem(Settings settings, String lore) {
        super(settings);
        this.lore = lore;
    }

    @Override
    public void appendTooltip(ItemStack stack, @Nullable World world, List<Text> tooltip, TooltipContext context) {
        super.appendTooltip(stack, world, tooltip, context);
        if (lore != null && !lore.isEmpty()) {
            tooltip.add(Text.translatable(lore).formatted(Formatting.ITALIC, Formatting.GOLD));
        }
    }
}
''')

(common / "net/jewelry/blocks/JewelersKitBlock.java").write_text(r'''package net.jewelry.blocks;

import net.minecraft.block.AbstractBlock;
import net.minecraft.block.Block;
import net.minecraft.block.BlockState;
import net.minecraft.client.item.TooltipContext;
import net.minecraft.item.ItemPlacementContext;
import net.minecraft.item.ItemStack;
import net.minecraft.state.StateManager;
import net.minecraft.state.property.DirectionProperty;
import net.minecraft.state.property.Properties;
import net.minecraft.text.Text;
import net.minecraft.util.Formatting;
import net.minecraft.util.math.BlockPos;
import net.minecraft.world.BlockView;
import org.jetbrains.annotations.Nullable;

import java.util.List;

public class JewelersKitBlock extends Block {
    public JewelersKitBlock(AbstractBlock.Settings settings) {
        super(settings);
    }

    @Override
    public void appendTooltip(ItemStack stack, @Nullable BlockView world, List<Text> tooltip, TooltipContext options) {
        super.appendTooltip(stack, world, tooltip, options);
        tooltip.add(Text.translatable("block.jewelry.jewelers_kit.hint").formatted(Formatting.GRAY, Formatting.ITALIC));
    }

    private static DirectionProperty FACING = Properties.HORIZONTAL_FACING;

    @Nullable
    @Override
    public BlockState getPlacementState(ItemPlacementContext ctx) {
        return this.getDefaultState().with(FACING, ctx.getHorizontalPlayerFacing().getOpposite());
    }

    @Override
    protected void appendProperties(StateManager.Builder<Block, BlockState> builder) {
        FACING = Properties.HORIZONTAL_FACING;
        builder.add(FACING);
    }

    public boolean isTranslucent(BlockState state, BlockView world, BlockPos pos) {
        return true;
    }
}
''')

(common / "net/jewelry/blocks/JewelryBlocks.java").write_text(r'''package net.jewelry.blocks;

import net.jewelry.JewelryMod;
import net.minecraft.block.AbstractBlock;
import net.minecraft.block.Block;
import net.minecraft.block.ExperienceDroppingBlock;
import net.minecraft.block.MapColor;
import net.minecraft.block.enums.Instrument;
import net.minecraft.item.BlockItem;
import net.minecraft.item.Item;
import net.minecraft.registry.Registries;
import net.minecraft.registry.Registry;
import net.minecraft.sound.BlockSoundGroup;
import net.minecraft.util.Identifier;
import net.minecraft.util.math.intprovider.UniformIntProvider;

import java.util.ArrayList;

public class JewelryBlocks {
    public record Entry(String name, Block block, BlockItem item) {
        public Entry(String name, Block block) {
            this(name, block, new BlockItem(block, new Item.Settings()));
        }
    }

    public static final ArrayList<Entry> all = new ArrayList<>();

    private static Entry entry(String name, Block block) {
        var entry = new Entry(name, block);
        all.add(entry);
        return entry;
    }

    public static final Entry GEM_VEIN = entry("gem_vein", new ExperienceDroppingBlock(
            AbstractBlock.Settings.create()
                    .mapColor(MapColor.STONE_GRAY)
                    .instrument(Instrument.BASEDRUM)
                    .requiresTool()
                    .strength(3.0F, 3.0F),
            UniformIntProvider.create(3, 7)
    ));

    public static final Entry DEEPSLATE_GEM_VEIN = entry("deepslate_gem_vein", new ExperienceDroppingBlock(
            AbstractBlock.Settings.create()
                    .instrument(Instrument.BASEDRUM)
                    .requiresTool()
                    .mapColor(MapColor.DEEPSLATE_GRAY)
                    .sounds(BlockSoundGroup.DEEPSLATE)
                    .strength(4.5F, 3.0F),
            UniformIntProvider.create(3, 7)
    ));

    public static final Entry JEWELERS_KIT = entry("jewelers_kit", new JewelersKitBlock(
            AbstractBlock.Settings.create()
                    .mapColor(MapColor.OAK_TAN)
                    .instrument(Instrument.BASS)
                    .strength(2.5F)
                    .sounds(BlockSoundGroup.WOOD)
                    .nonOpaque()
    ));

    public static void register() {
        for (var entry : all) {
            Registry.register(Registries.BLOCK, new Identifier(JewelryMod.ID, entry.name), entry.block);
            Registry.register(Registries.ITEM, new Identifier(JewelryMod.ID, entry.name), entry.item());
        }
    }
}
''')

curio = forge / "net/jewelry/forge/compat/curios/JewelryCurioItem.java"
if not curio.exists():
    raise SystemExit("JewelryCurioItem Forge translation missing before compat pass 1")
curio.write_text(r'''package net.jewelry.forge.compat.curios;

import com.google.common.collect.LinkedHashMultimap;
import com.google.common.collect.Multimap;
import net.jewelry.items.JewelryAttributeModifier;
import net.jewelry.items.JewelryItem;
import net.jewelry.util.SoundHelper;
import net.minecraft.client.item.TooltipContext;
import net.minecraft.entity.attribute.EntityAttribute;
import net.minecraft.entity.attribute.EntityAttributeModifier;
import net.minecraft.item.Item;
import net.minecraft.item.ItemStack;
import net.minecraft.registry.Registries;
import net.minecraft.text.Text;
import net.minecraft.util.Formatting;
import net.minecraft.world.World;
import org.jetbrains.annotations.Nullable;
import top.theillusivec4.curios.api.SlotContext;
import top.theillusivec4.curios.api.type.capability.ICurio;
import top.theillusivec4.curios.api.type.capability.ICurioItem;

import java.nio.charset.StandardCharsets;
import java.util.List;
import java.util.UUID;

public class JewelryCurioItem extends Item implements ICurioItem, JewelryItem {
    private List<JewelryAttributeModifier> customAttributes = List.of();
    private final String lore;

    public JewelryCurioItem(Item.Settings settings, String lore) {
        super(settings);
        this.lore = lore;
    }

    @Override
    public void appendTooltip(ItemStack stack, @Nullable World world, List<Text> tooltip, TooltipContext context) {
        super.appendTooltip(stack, world, tooltip, context);
        if (lore != null && !lore.isEmpty()) {
            tooltip.add(Text.translatable(lore).formatted(Formatting.ITALIC, Formatting.GOLD));
        }
    }

    @Override
    public Multimap<EntityAttribute, EntityAttributeModifier> getAttributeModifiers(
            SlotContext slotContext, UUID slotUuid, ItemStack stack) {
        Multimap<EntityAttribute, EntityAttributeModifier> modifiers = LinkedHashMultimap.create();
        modifiers.putAll(ICurioItem.super.getAttributeModifiers(slotContext, slotUuid, stack));

        // Curios 1.20.x supplies a slot UUID. Mix the item id into it so quick replacement in the
        // same slot cannot reuse an existing modifier identity, while different slots still stack.
        var itemId = Registries.ITEM.getId(stack.getItem());
        UUID itemSlotUuid = UUID.nameUUIDFromBytes(
                (slotUuid + "/" + itemId).getBytes(StandardCharsets.UTF_8));
        for (var entry : customAttributes) {
            modifiers.put(entry.attribute(), new EntityAttributeModifier(
                    itemSlotUuid, entry.id(), entry.value(), entry.operation()));
        }
        return modifiers;
    }

    public void setConfigurableModifiers(List<JewelryAttributeModifier> modifiers) {
        this.customAttributes = modifiers;
    }

    @Override
    public ICurio.SoundInfo getEquipSound(SlotContext slotContext, ItemStack stack) {
        return new ICurio.SoundInfo(SoundHelper.JEWELRY_EQUIP, 1.0F, 1.0F);
    }

    @Override
    public void onEquip(SlotContext slotContext, ItemStack prevStack, ItemStack stack) {
        var entity = slotContext.entity();
        if (entity == null) {
            return;
        }
        var world = entity.getWorld();
        if (world.isClient()
                || entity.age <= 100
                || prevStack.isOf(stack.getItem())) {
            return;
        }
        world.playSound(null, entity.getBlockPos(), SoundHelper.JEWELRY_EQUIP,
                entity.getSoundCategory(), 1.0F, 1.0F);
    }

    @Override
    public void onEquipFromUse(SlotContext slotContext, ItemStack stack) {
        // Silent: onEquip already owns the Jewelry-specific equip sound for every equip path.
    }
}
''')

# Mechanical symbol renames shared by current 2.4.0 source. These are one-to-one Yarn API changes.
for java_root in (common, forge):
    for path in java_root.rglob("*.java"):
        text = path.read_text()
        text = text.replace("net.tiny_config.ConfigManager", "net.tinyconfig.ConfigManager")
        text = text.replace("EntityAttributeModifier.Operation.ADD_VALUE", "EntityAttributeModifier.Operation.ADDITION")
        text = text.replace("EntityAttributeModifier.Operation.ADD_MULTIPLIED_BASE", "EntityAttributeModifier.Operation.MULTIPLY_BASE")
        text = text.replace("EntityAttributeModifier.Operation.ADD_MULTIPLIED_TOTAL", "EntityAttributeModifier.Operation.MULTIPLY_TOTAL")
        text = text.replace("Identifier.of(", "new Identifier(")
        # Current Curios helper has a 1.21-type name only in an explanatory comment; keep the
        # assertion below focused on actual code/imports instead of documentation text.
        text = text.replace("because Curios ignores `AttributeModifiersComponent`",
                            "because Curios ignores vanilla item-setting attribute carriers")
        path.write_text(text)

# Fail early if any known 1.21-only representation survived this pass.
for forbidden in (
    "net.minecraft.component.type",
    "AttributeModifiersComponent",
    "AttributeModifierSlot",
    "net.minecraft.item.tooltip.TooltipType",
    "Item.TooltipContext",
    "NoteBlockInstrument",
    "net.tiny_config",
    "Operation.ADD_VALUE",
    "Operation.ADD_MULTIPLIED_BASE",
    "Operation.ADD_MULTIPLIED_TOTAL",
    "Identifier.of(",
):
    hits = []
    for java_root in (common, forge):
        for path in java_root.rglob("*.java"):
            if forbidden in path.read_text():
                hits.append(str(path.relative_to(root)))
    if hits:
        raise SystemExit(f"compat pass 1 left 1.21-only symbol {forbidden}: {hits[:8]}")

# Acceptance invariants for the modern Jewelry behavior carried through the bridge.
final_items = items_path.read_text()
for required in (
    "itemConfig.selectedAttributes()",
    "new JewelryAttributeModifier(",
    "entry.id().getPath()",
    "unique_crit_ring",
    "unique_dex_ring",
    "diamond_ring",
):
    if required not in final_items:
        raise SystemExit(f"compat pass 1 lost Jewelry 2.4.0 behavior/catalog anchor: {required}")
final_curio = curio.read_text()
for required in (
    "UUID.nameUUIDFromBytes",
    "slotUuid + \"/\" + itemId",
    "LinkedHashMultimap.create()",
    "setConfigurableModifiers(List<JewelryAttributeModifier>",
):
    if required not in final_curio:
        raise SystemExit(f"compat pass 1 lost Curios stacking invariant: {required}")

print("Jewelry compatibility pass 1 applied: 1.20.1 tooltips/blocks/config + component-to-Curios attribute bridge")
