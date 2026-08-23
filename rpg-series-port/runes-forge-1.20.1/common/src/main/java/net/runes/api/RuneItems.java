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

    public static void bootstrap() {
        if (bootstrapped) return;
        bootstrapped = true;
        for (var type : RuneType.values()) {
            var id = new Identifier(RunesMod.ID, type.toString().toLowerCase(Locale.ENGLISH) + "_stone");
            entries.add(new Entry(id, type, new Item(new Item.Settings())));
        }
    }

    public static Item get(RuneType type) {
        bootstrap();
        return entries.stream().filter(e -> e.type() == type).findFirst().orElseThrow().item();
    }

    public static List<Entry> all() { bootstrap(); return Collections.unmodifiableList(entries); }
}
