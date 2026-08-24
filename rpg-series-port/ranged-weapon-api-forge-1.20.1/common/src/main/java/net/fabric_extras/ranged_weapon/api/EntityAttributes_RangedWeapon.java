package net.fabric_extras.ranged_weapon.api;

import net.minecraft.entity.attribute.ClampedEntityAttribute;
import net.minecraft.entity.attribute.EntityAttribute;
import net.minecraft.registry.Registries;
import net.minecraft.registry.entry.RegistryEntry;
import net.minecraft.util.Identifier;
import org.jetbrains.annotations.Nullable;
import java.util.ArrayList;

public final class EntityAttributes_RangedWeapon {
    public static final String NAMESPACE = "ranged_weapon";
    public static final ArrayList<Entry> all = new ArrayList<>();
    private static Entry entry(String name, double base, boolean tracked) { return entry(name, 0, base, tracked); }
    private static Entry entry(String name, double min, double base, boolean tracked) {
        var value = new Entry(name, min, base, tracked); all.add(value); return value;
    }
    public static final class Entry {
        public final Identifier id;
        public final String translationKey;
        public final EntityAttribute attribute;
        public final double baseValue;
        @Nullable public RegistryEntry<EntityAttribute> entry;
        @Nullable private Identifier baseModifierId;
        Entry(String name, double min, double base, boolean tracked) {
            id = new Identifier(NAMESPACE, name);
            translationKey = "attribute.name." + NAMESPACE + "." + name;
            attribute = new ClampedEntityAttribute(translationKey, base, min, 2048).setTracked(tracked);
            baseValue = base;
        }
        public double asMultiplier(double value) { return value / baseValue; }
        public Entry setBaseAttributeId(Identifier id) { this.baseModifierId = id; return this; }
        @Nullable public Identifier getBaseAttributeId() { return baseModifierId; }
        public void bindRegistryEntry() { this.entry = Registries.ATTRIBUTE.getEntry(id); }
    }
    public static final Entry DAMAGE = entry("damage", 0, true).setBaseAttributeId(AttributeModifierIDs.WEAPON_DAMAGE_ID);
    public static final Entry PULL_TIME = entry("pull_time", 0.1, 1.0, true).setBaseAttributeId(AttributeModifierIDs.WEAPON_PULL_TIME_ID);
    public static final Entry HASTE = entry("haste", 100, true);
    public static final Entry VELOCITY = entry("velocity", 0, false);
    private EntityAttributes_RangedWeapon() {}
}
