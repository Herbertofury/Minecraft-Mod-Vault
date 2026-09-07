package net.runes.crafting;

import net.minecraft.block.*;
import net.minecraft.client.item.TooltipContext;
import net.minecraft.entity.player.PlayerEntity;
import net.minecraft.item.BlockItem;
import net.minecraft.item.Item;
import net.minecraft.item.ItemPlacementContext;
import net.minecraft.item.ItemStack;
import net.minecraft.screen.NamedScreenHandlerFactory;
import net.minecraft.screen.ScreenHandlerContext;
import net.minecraft.screen.SimpleNamedScreenHandlerFactory;
import net.minecraft.state.StateManager;
import net.minecraft.state.property.DirectionProperty;
import net.minecraft.state.property.Properties;
import net.minecraft.text.Text;
import net.minecraft.util.ActionResult;
import net.minecraft.util.Formatting;
import net.minecraft.util.Hand;
import net.minecraft.util.hit.BlockHitResult;
import net.minecraft.util.math.BlockPos;
import net.minecraft.util.shape.VoxelShape;
import net.minecraft.util.shape.VoxelShapes;
import net.minecraft.world.BlockView;
import net.minecraft.world.World;
import net.runes.RunesMod;
import org.jetbrains.annotations.Nullable;

import java.util.List;

public class RuneCraftingBlock extends CraftingTableBlock {
    public static final String NAME = "crafting_altar";
    public static RuneCraftingBlock INSTANCE;
    public static BlockItem ITEM;

    private static final Text SCREEN_TITLE = Text.translatable("gui.runes.rune_crafting");
    private static final VoxelShape SHAPE = VoxelShapes.union(
            Block.createCuboidShape(1, 12, 1, 15, 16, 15),
            Block.createCuboidShape(4, 3, 4, 12, 12, 12),
            Block.createCuboidShape(1, 0, 1, 15, 3, 15));
    private static final DirectionProperty FACING = Properties.HORIZONTAL_FACING;

    /** Called only while Forge has the BLOCK registry open. */
    public static void bootstrapBlock() {
        if (INSTANCE == null) {
            INSTANCE = new RuneCraftingBlock(AbstractBlock.Settings.create().strength(2.0F).nonOpaque());
        }
    }

    /** Called only while Forge has the ITEM registry open, after bootstrapBlock(). */
    public static void bootstrapItem() {
        if (ITEM != null) return;
        if (INSTANCE == null) {
            throw new IllegalStateException("Rune crafting altar item registered before its block");
        }
        ITEM = new BlockItem(INSTANCE, new Item.Settings());
    }

    public RuneCraftingBlock(Settings settings) {
        super(settings);
        setDefaultState(getStateManager().getDefaultState().with(FACING, net.minecraft.util.math.Direction.NORTH));
    }

    @Override
    public void appendTooltip(ItemStack stack, @Nullable BlockView world, List<Text> tooltip, TooltipContext options) {
        super.appendTooltip(stack, world, tooltip, options);
        tooltip.add(Text.translatable("block." + RunesMod.ID + "." + NAME + ".hint")
                .formatted(Formatting.GRAY, Formatting.ITALIC));
    }

    public NamedScreenHandlerFactory createScreenHandlerFactory(BlockState state, World world, BlockPos pos) {
        return new SimpleNamedScreenHandlerFactory(
                (syncId, inventory, player) -> new RuneCraftingScreenHandler(syncId, inventory, ScreenHandlerContext.create(world, pos)),
                SCREEN_TITLE);
    }

    public ActionResult onUse(BlockState state, World world, BlockPos pos, PlayerEntity player, Hand hand, BlockHitResult hit) {
        if (world.isClient) return ActionResult.SUCCESS;
        player.openHandledScreen(state.createScreenHandlerFactory(world, pos));
        return ActionResult.CONSUME;
    }

    @Override
    public VoxelShape getOutlineShape(BlockState state, BlockView world, BlockPos pos, ShapeContext context) { return SHAPE; }

    @Nullable
    @Override
    public BlockState getPlacementState(ItemPlacementContext ctx) {
        return getDefaultState().with(FACING, ctx.getHorizontalPlayerFacing().getOpposite());
    }

    @Override
    protected void appendProperties(StateManager.Builder<Block, BlockState> builder) { builder.add(FACING); }

    public boolean isTranslucent(BlockState state, BlockView world, BlockPos pos) { return true; }
}
