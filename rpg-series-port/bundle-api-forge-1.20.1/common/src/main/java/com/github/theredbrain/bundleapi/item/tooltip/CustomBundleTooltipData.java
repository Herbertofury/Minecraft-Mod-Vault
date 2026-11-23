package com.github.theredbrain.bundleapi.item.tooltip;

import com.github.theredbrain.bundleapi.component.type.CustomBundleContentsComponent;
import net.minecraft.client.item.TooltipData;

/** Client tooltip payload preserving Bundle API's current item-grid semantics on Minecraft 1.20.1. */
public record CustomBundleTooltipData(CustomBundleContentsComponent contents) implements TooltipData {
}
