package net.runes.crafting;

import net.minecraft.item.Item;
import net.minecraft.util.Identifier;
import net.minecraft.util.Rarity;
import net.runes.RunesMod;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;

public final class RunePouches {
    public record Entry(Identifier id,int capacity,RunePouchItem item) { }
    public static final List<Entry> entries=new ArrayList<>();
    private static boolean bootstrapped;
    private RunePouches() { }
    public static void bootstrap(){
        if(bootstrapped) return; bootstrapped=true;
        entries.add(entry("small_rune_pouch",4,null));
        entries.add(entry("medium_rune_pouch",8,null));
        entries.add(entry("large_rune_pouch",12,Rarity.UNCOMMON));
    }
    private static Entry entry(String name,int capacity,Rarity rarity){
        Item.Settings settings=new Item.Settings().maxCount(1);
        if(rarity!=null) settings.rarity(rarity);
        return new Entry(new Identifier(RunesMod.ID,name),capacity,new RunePouchItem(capacity,settings));
    }
    public static List<Entry> all(){ bootstrap(); return Collections.unmodifiableList(entries); }
}
