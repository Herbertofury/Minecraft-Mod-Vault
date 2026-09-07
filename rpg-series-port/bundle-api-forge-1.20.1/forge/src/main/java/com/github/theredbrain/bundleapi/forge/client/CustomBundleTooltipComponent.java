package com.github.theredbrain.bundleapi.forge.client;

import com.github.theredbrain.bundleapi.component.type.CustomBundleContentsComponent;
import com.github.theredbrain.bundleapi.item.tooltip.CustomBundleTooltipData;
import net.minecraft.client.font.TextRenderer;
import net.minecraft.client.gui.DrawContext;
import net.minecraft.client.gui.screen.ingame.HandledScreen;
import net.minecraft.client.gui.tooltip.TooltipComponent;
import net.minecraft.item.ItemStack;
import net.minecraft.util.Identifier;
import org.apache.commons.lang3.math.Fraction;

/** 1.20.1 renderer for the current Bundle API grid tooltip, using the vanilla bundle texture atlas. */
public final class CustomBundleTooltipComponent implements TooltipComponent {
    private static final Identifier TEXTURE = new Identifier("textures/gui/container/bundle.png");
    private static final int WIDTH_PER_COLUMN = 18;
    private static final int HEIGHT_PER_ROW = 20;
    private static final int TEXTURE_SIZE = 128;

    private final CustomBundleContentsComponent contents;

    public CustomBundleTooltipComponent(CustomBundleTooltipData data) {
        this.contents = data.contents();
    }

    @Override
    public int getHeight() {
        return rows() * HEIGHT_PER_ROW + 2 + 4;
    }

    @Override
    public int getWidth(TextRenderer textRenderer) {
        return columns() * WIDTH_PER_COLUMN + 2;
    }

    @Override
    public void drawItems(TextRenderer textRenderer, int x, int y, DrawContext context) {
        int columns = columns();
        int rows = rows();
        boolean blocked = contents.getOccupancy().compareTo(Fraction.ONE) >= 0;
        int index = 0;
        for (int row = 0; row < rows; row++) {
            for (int column = 0; column < columns; column++) {
                drawSlot(x + column * WIDTH_PER_COLUMN + 1, y + row * HEIGHT_PER_ROW + 1,
                        index++, blocked, context, textRenderer);
            }
        }
        drawOutline(x, y, columns, rows, context);
    }

    private void drawSlot(int x, int y, int index, boolean blocked, DrawContext context, TextRenderer textRenderer) {
        if (index >= contents.size()) {
            draw(context, x, y, blocked ? Sprite.BLOCKED_SLOT : Sprite.SLOT);
            return;
        }

        ItemStack stack = contents.get(index);
        draw(context, x, y, Sprite.SLOT);
        context.drawItem(stack, x + 1, y + 1, index);
        context.drawItemInSlot(textRenderer, stack, x + 1, y + 1);
        if (index == 0) {
            HandledScreen.drawSlotHighlight(context, x + 1, y + 1, 0);
        }
    }

    private void drawOutline(int x, int y, int columns, int rows, DrawContext context) {
        draw(context, x, y, Sprite.BORDER_CORNER_TOP);
        draw(context, x + columns * WIDTH_PER_COLUMN + 1, y, Sprite.BORDER_CORNER_TOP);
        for (int column = 0; column < columns; column++) {
            draw(context, x + 1 + column * WIDTH_PER_COLUMN, y, Sprite.BORDER_HORIZONTAL_TOP);
            draw(context, x + 1 + column * WIDTH_PER_COLUMN, y + rows * HEIGHT_PER_ROW, Sprite.BORDER_HORIZONTAL_BOTTOM);
        }
        for (int row = 0; row < rows; row++) {
            draw(context, x, y + row * HEIGHT_PER_ROW + 1, Sprite.BORDER_VERTICAL);
            draw(context, x + columns * WIDTH_PER_COLUMN + 1, y + row * HEIGHT_PER_ROW + 1, Sprite.BORDER_VERTICAL);
        }
        draw(context, x, y + rows * HEIGHT_PER_ROW, Sprite.BORDER_CORNER_BOTTOM);
        draw(context, x + columns * WIDTH_PER_COLUMN + 1, y + rows * HEIGHT_PER_ROW, Sprite.BORDER_CORNER_BOTTOM);
    }

    private void draw(DrawContext context, int x, int y, Sprite sprite) {
        context.drawTexture(TEXTURE, x, y, 0, (float) sprite.u, (float) sprite.v,
                sprite.width, sprite.height, TEXTURE_SIZE, TEXTURE_SIZE);
    }

    private int columns() {
        return Math.max(2, (int) Math.ceil(Math.sqrt(contents.size() + 1.0)));
    }

    private int rows() {
        return (int) Math.ceil((contents.size() + 1.0) / columns());
    }

    private enum Sprite {
        SLOT(0, 0, 18, 20),
        BLOCKED_SLOT(0, 40, 18, 20),
        BORDER_VERTICAL(0, 18, 1, 20),
        BORDER_HORIZONTAL_TOP(0, 20, 18, 1),
        BORDER_HORIZONTAL_BOTTOM(0, 60, 18, 1),
        BORDER_CORNER_TOP(0, 20, 1, 1),
        BORDER_CORNER_BOTTOM(0, 60, 1, 1);

        private final int u;
        private final int v;
        private final int width;
        private final int height;

        Sprite(int u, int v, int width, int height) {
            this.u = u;
            this.v = v;
            this.width = width;
            this.height = height;
        }
    }
}
