package net.runes.forge;

import net.minecraft.item.ItemStack;
import net.minecraft.item.Items;
import net.minecraft.registry.Registries;
import net.minecraft.util.Identifier;
import net.minecraftforge.event.server.ServerStartedEvent;
import net.runes.RunesMod;
import net.runes.api.RuneItems;
import net.runes.crafting.RuneCrafting;
import net.runes.crafting.RunePouchItem;
import net.runes.crafting.RunePouches;

/** Runtime-only regression gate enabled by -Drunes.ci.selftest=true in the CI dev server. */
final class CiSelfTest {
    private CiSelfTest() { }

    static void onServerStarted(ServerStartedEvent event) {
        if (!Boolean.getBoolean("runes.ci.selftest")) return;

        require(new Identifier(RunesMod.ID, "small_rune_pouch")
                        .equals(Registries.ITEM.getId(RunePouches.all().get(0).item())),
                "small pouch registry id changed");
        require(new Identifier(RunesMod.ID, "arcane_stone")
                        .equals(Registries.ITEM.getId(RuneItems.get(RuneItems.RuneType.ARCANE))),
                "arcane rune registry id changed");

        int altarRecipes = event.getServer().getRecipeManager().listAllOfType(RuneCrafting.RECIPE_TYPE).size();
        require(altarRecipes >= 8, "expected at least 8 rune altar recipes, got " + altarRecipes);

        RunePouchItem small = RunePouches.all().get(0).item();
        ItemStack pouch = new ItemStack(small);
        for (int i = 0; i < 4; i++) {
            ItemStack runes = new ItemStack(RuneItems.get(RuneItems.RuneType.ARCANE), 64);
            require(small.insertRunes(pouch, runes) == 64, "failed to insert full rune stack " + i);
            require(runes.isEmpty(), "source rune stack was not consumed");
        }
        require(RunePouchItem.count(pouch) == 256, "small pouch capacity is not 4 stacks");

        ItemStack overflow = new ItemStack(RuneItems.get(RuneItems.RuneType.FIRE), 1);
        require(small.insertRunes(pouch, overflow) == 0, "small pouch accepted overflow rune");
        require(overflow.getCount() == 1, "overflow source was mutated");

        ItemStack invalid = new ItemStack(Items.DIAMOND, 1);
        require(small.insertRunes(pouch, invalid) == 0, "pouch accepted a non-rune item");
        require(invalid.getCount() == 1, "non-rune source was mutated");

        ItemStack extracted = small.extractLast(pouch, 64);
        require(extracted.getCount() == 64, "expected a full extracted rune stack");
        require(RunePouchItem.count(pouch) == 192, "pouch count did not persist after extraction");
        require(RunePouchItem.contents(pouch).stream().allMatch(RunePouchItem::isRune),
                "pouch persisted a non-rune entry");

        System.out.println("[Runes CI] Runtime self-test passed: registry + recipes + pouch NBT/capacity/filtering");
    }

    private static void require(boolean condition, String message) {
        if (!condition) throw new IllegalStateException("[Runes CI] " + message);
    }
}
