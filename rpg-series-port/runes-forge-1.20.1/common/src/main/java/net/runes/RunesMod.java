package net.runes;

import net.minecraft.sound.SoundEvent;
import net.minecraft.util.Identifier;
import net.runes.crafting.RuneCraftingRecipe;

public final class RunesMod {
    public static final String ID = "runes";
    private RunesMod() { }

    public static final Identifier CRAFTING_ID = new Identifier(ID, RuneCraftingRecipe.NAME);
    public static final SoundEvent CRAFTING_SOUND = SoundEvent.of(CRAFTING_ID);

    /**
     * Common source intentionally owns no intrusive Forge registry bootstrap.
     * Platform entrypoints create registry-backed objects only while their registry is open.
     */
    public static void bootstrapCommon() { }
}
