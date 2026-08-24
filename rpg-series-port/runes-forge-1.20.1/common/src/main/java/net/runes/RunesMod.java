package net.runes;

import net.minecraft.sound.SoundEvent;
import net.minecraft.util.Identifier;
import net.runes.api.RuneItems;
import net.runes.crafting.RuneCrafting;
import net.runes.crafting.RuneCraftingBlock;
import net.runes.crafting.RuneCraftingRecipe;
import net.runes.crafting.RunePouches;

public final class RunesMod {
    public static final String ID = "runes";
    private RunesMod() { }

    public static final Identifier CRAFTING_ID = new Identifier(ID, RuneCraftingRecipe.NAME);
    public static final SoundEvent CRAFTING_SOUND = SoundEvent.of(CRAFTING_ID);

    public static void bootstrapCommon() {
        // Deliberately loader-agnostic. Forge owns registration timing and loader-specific factories.
        RuneItems.bootstrap();
        RunePouches.bootstrap();
        RuneCraftingBlock.bootstrap();
        RuneCrafting.bootstrap();
    }
}
