package net.runes.api;

import net.minecraft.item.Item;
import net.minecraft.util.Identifier;
import net.runes.RunesMod;

import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Locale;

public final class RuneItems {
    public enum RuneType { ARCANE, FIRE, FROST, HEALING, LIGHTNING, SOUL }
    public record Entry(Identifier id, RuneType type, Item item) { }

    public static final List<Entry> entries = new ArrayList<>();
    private static boolean bootstrapped;

    private RuneItems() { }

    /** Called only while Forge has the ITEM registry open. */
    public static void bootstrap() {
        if (bootstrapped) return;
        for (var type : RuneType.values()) {
            var id = new Identifier(RunesMod.ID, type.toString().toLowerCase(Locale.ENGLISH) + "_stone");
            entries.add(new Entry(id, type, new Item(new Item.Settings())));
        }
        bootstrapped = true;
    }

    public static Item get(RuneType type) {
        requireBootstrapped();
        return entries.stream().filter(e -> e.type() == type).findFirst().orElseThrow().item();
    }

    public static List<Entry> all() {
        requireBootstrapped();
        return Collections.unmodifiableList(entries);
    }

    private static void requireBootstrapped() {
        if (!bootstrapped) {
            throw new IllegalStateException("Rune items were accessed before Forge ITEM registration");
        }
    }
}
