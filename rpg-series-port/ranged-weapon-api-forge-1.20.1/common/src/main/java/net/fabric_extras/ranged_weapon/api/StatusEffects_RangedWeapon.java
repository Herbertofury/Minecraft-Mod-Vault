package net.fabric_extras.ranged_weapon.api;

import net.fabric_extras.ranged_weapon.internal.CustomStatusEffect;
import net.minecraft.entity.effect.StatusEffect;
import net.minecraft.entity.effect.StatusEffectCategory;
import net.minecraft.registry.Registries;
import net.minecraft.registry.entry.RegistryEntry;
import net.minecraft.util.Identifier;
import org.jetbrains.annotations.Nullable;
import java.util.ArrayList;

public final class StatusEffects_RangedWeapon {
    public static final String NAMESPACE = "ranged_weapon";
    public static final class Entry {
        public final Identifier id;
        public final StatusEffect effect;
        @Nullable public RegistryEntry<StatusEffect> entry;
        Entry(String name, int color) { id = new Identifier(NAMESPACE, name); effect = new CustomStatusEffect(StatusEffectCategory.BENEFICIAL, color); }
        public void bindRegistryEntry() { this.entry = Registries.STATUS_EFFECT.getEntry(effect); }
    }
    public static final ArrayList<Entry> all = new ArrayList<>();
    private static Entry entry(String name, int color) { var e = new Entry(name, color); all.add(e); return e; }
    public static final Entry DAMAGE = entry("damage", 0xAAFFDD);
    public static final Entry HASTE = entry("haste", 0xB30000);
    private StatusEffects_RangedWeapon() {}
}
