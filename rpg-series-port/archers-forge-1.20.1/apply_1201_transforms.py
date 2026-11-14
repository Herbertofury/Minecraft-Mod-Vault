#!/usr/bin/env python3
from __future__ import annotations

import json
import pathlib
import sys

java_root = pathlib.Path(sys.argv[1]).resolve()
resources_root = pathlib.Path(sys.argv[2]).resolve()
if not java_root.is_dir():
    raise SystemExit(f"missing generated Java root: {java_root}")
if not resources_root.is_dir():
    raise SystemExit(f"missing generated resources root: {resources_root}")

identifier_rewrites = 0
for path in sorted(java_root.rglob("*.java")):
    text = path.read_text(encoding="utf-8")
    updated = text.replace("Identifier.of(", "new Identifier(")
    if updated != text:
        identifier_rewrites += text.count("Identifier.of(")
        path.write_text(updated, encoding="utf-8")

quivers = java_root / "net/archers/item/Quivers.java"
if not quivers.is_file():
    raise SystemExit(f"current Archers Quivers.java missing: {quivers}")
quivers.write_text('''package net.archers.item;

import com.github.theredbrain.bundleapi.item.CustomBundleItem;
import net.archers.ArchersMod;
import net.minecraft.client.item.TooltipContext;
import net.minecraft.item.Item;
import net.minecraft.item.ItemStack;
import net.minecraft.registry.Registries;
import net.minecraft.registry.Registry;
import net.minecraft.registry.tag.ItemTags;
import net.minecraft.registry.tag.TagKey;
import net.minecraft.text.Text;
import net.minecraft.util.Formatting;
import net.minecraft.util.Identifier;
import net.minecraft.util.Rarity;
import net.minecraft.world.World;
import org.jetbrains.annotations.Nullable;

import java.util.ArrayList;
import java.util.List;
import java.util.function.Function;

public class Quivers {
    public static final List<Entry> entries = new ArrayList<>();
    public record Entry(Identifier id, int capacity, Item item) { }
    public record Args(TagKey<Item> tag, int capacity, Item.Settings settings) { }

    public static class QuiverItem extends CustomBundleItem {
        public QuiverItem(@Nullable TagKey<Item> tag, int capacity, Settings settings) {
            super(tag, capacity, settings);
        }

        @Override
        public void appendTooltip(ItemStack stack, @Nullable World world, List<Text> tooltip, TooltipContext context) {
            tooltip.add(Text.translatable("item.archers.quiver.hint").formatted(Formatting.GRAY));
            super.appendTooltip(stack, world, tooltip, context);
        }
    }

    public static Function<Args, Item> factory = args -> new QuiverItem(args.tag(), args.capacity(), args.settings());

    public static Entry entry(String name, int capacity, @Nullable Rarity rarity) {
        var settings = new Item.Settings().maxCount(1);
        if (rarity != null) settings.rarity(rarity);
        var bundle = factory.apply(new Args(ItemTags.ARROWS, capacity, settings));
        var id = new Identifier(ArchersMod.ID, name);
        var entry = new Entry(id, capacity, bundle);
        entries.add(entry);
        return entry;
    }

    public static void register() {
        entry("small_quiver", 4, null);
        entry("medium_quiver", 8, null);
        entry("large_quiver", 12, Rarity.UNCOMMON);
        for (var entry : entries) Registry.register(Registries.ITEM, entry.id(), entry.item());
    }
}
''', encoding="utf-8")

# 1.21 ComponentType<Boolean> AUTO_FIRE -> 1.20.1 stack NBT boolean. The feature semantics are
# identical (persistent per-stack true/false state); only the storage API changes.
auto_fire = java_root / "net/archers/item/misc/AutoFireHook.java"
if not auto_fire.is_file():
    raise SystemExit(f"current AutoFireHook.java missing: {auto_fire}")
auto_fire.write_text('''package net.archers.item.misc;

import net.archers.ArchersMod;
import net.minecraft.item.CrossbowItem;
import net.minecraft.item.Item;
import net.minecraft.item.ItemStack;
import net.minecraft.registry.RegistryKeys;
import net.minecraft.registry.tag.TagKey;
import net.minecraft.util.Identifier;

public class AutoFireHook {
    private static final String NBT_KEY = "ArchersAutoFire";
    public static final Identifier id = new Identifier(ArchersMod.ID, "auto_fire_hook");
    public static final Item item = new AutoFireHookItem((new Item.Settings()).maxCount(1));
    public static final TagKey<Item> AFH_ATTACHABLE = TagKey.of(RegistryKeys.ITEM, new Identifier(ArchersMod.ID, "auto_fire_hook_attachables"));

    public static boolean isApplied(ItemStack itemStack) {
        return itemStack != null && !itemStack.isEmpty() && itemStack.hasNbt() && itemStack.getNbt().getBoolean(NBT_KEY);
    }

    public static void apply(ItemStack itemStack) {
        if (itemStack != null && !itemStack.isEmpty()) itemStack.getOrCreateNbt().putBoolean(NBT_KEY, true);
    }

    public static void remove(ItemStack itemStack) {
        if (itemStack == null || !itemStack.hasNbt()) return;
        itemStack.getNbt().remove(NBT_KEY);
        if (itemStack.getNbt().isEmpty()) itemStack.setNbt(null);
    }

    public static boolean isApplicable(ItemStack itemStack) {
        if (itemStack == null || itemStack.isEmpty()) return false;
        return (itemStack.getItem() instanceof CrossbowItem || itemStack.isIn(AFH_ATTACHABLE)) && !isApplied(itemStack);
    }
}
''', encoding="utf-8")

# 1.20.1 has ordinary SoundEvent registry values, not the 1.21 reference-registration helper.
# Preserve every current 3.1.1 sound id/variant and only translate the registration representation.
archer_sounds = java_root / "net/archers/content/ArcherSounds.java"
if not archer_sounds.is_file():
    raise SystemExit(f"current ArcherSounds.java missing: {archer_sounds}")
archer_sounds.write_text('''package net.archers.content;

import net.archers.ArchersMod;
import net.archers.block.ArcherWorkbenchBlock;
import net.minecraft.registry.Registries;
import net.minecraft.registry.Registry;
import net.minecraft.sound.SoundEvent;
import net.minecraft.util.Identifier;

import java.util.ArrayList;
import java.util.List;

public class ArcherSounds {
    public static class Entry {
        private final Identifier id;
        private final SoundEvent soundEvent;
        private int variants = 1;

        public Entry(Identifier id, SoundEvent soundEvent) {
            this.id = id;
            this.soundEvent = soundEvent;
        }

        public Entry(String name) {
            this(new Identifier(ArchersMod.ID, name));
        }

        public Entry(Identifier id) {
            this(id, SoundEvent.of(id));
        }

        public Entry travelDistance(float distance) {
            return new Entry(id, SoundEvent.of(id, distance));
        }

        public Entry variants(int variants) {
            this.variants = variants;
            return this;
        }

        public Identifier id() { return id; }
        public SoundEvent soundEvent() { return soundEvent; }
        public int variants() { return variants; }
    }

    public static final List<Entry> entries = new ArrayList<>();
    public static Entry add(Entry entry) {
        entries.add(entry);
        return entry;
    }

    public static final Entry MARKER_SHOT = add(new Entry("marker_shot"));
    public static final Entry ENTANGLING_ROOTS = add(new Entry("entangling_roots"));
    public static final Entry BOW_PULL = add(new Entry("bow_pull"));
    public static final Entry MAGIC_ARROW_IMPACT = add(new Entry("magic_arrow_impact"));
    public static final Entry MAGIC_ARROW_RELEASE = add(new Entry("magic_arrow_release"));
    public static final Entry MAGIC_ARROW_START = add(new Entry("magic_arrow_start"));
    public static final Entry WORKBENCH = add(new Entry(ArcherWorkbenchBlock.ID.getPath()));
    public static final Entry ARCHER_ARMOR_EQUIP = add(new Entry("archer_armor"));
    public static final Entry RAIN_OF_ARROWS_RELEASE = add(new Entry("rain_of_arrows_release"));
    public static final Entry RAIN_OF_ARROWS_IMPACT = add(new Entry("rain_of_arrows_impact"));
    public static final Entry SPIRIT_WOLF_SPAWN = add(new Entry("spirit_wolf_spawn"));
    public static final Entry SPIRIT_WOLF_SUMMON = add(new Entry("spirit_wolf_summon").variants(2));

    public static void register() {
        for (var entry : entries) {
            Registry.register(Registries.SOUND_EVENT, entry.id(), entry.soundEvent());
        }
    }
}
''', encoding="utf-8")

# 1.20.1 ArmorMaterial is a value interface and has no ARMOR_MATERIAL registry. Keep the current
# 3.1.1 armor protection/enchant/sound/repair/tier/config semantics, wrap a target-native material
# in RegistryEntry.of(...), and expose the namespaced material name used by the proven Spell Engine
# 1.20.1 compatibility ABI for armor texture identity.
archer_armors = java_root / "net/archers/item/ArcherArmors.java"
if not archer_armors.is_file():
    raise SystemExit(f"current ArcherArmors.java missing: {archer_armors}")
archer_armors.write_text('''package net.archers.item;

import net.archers.ArchersMod;
import net.archers.content.ArcherSounds;
import net.archers.item.armor.ArcherArmor;
import net.fabric_extras.ranged_weapon.api.EntityAttributes_RangedWeapon;
import net.minecraft.entity.attribute.EntityAttributeModifier;
import net.minecraft.item.ArmorItem;
import net.minecraft.item.ArmorMaterial;
import net.minecraft.item.Items;
import net.minecraft.recipe.Ingredient;
import net.minecraft.registry.entry.RegistryEntry;
import net.minecraft.sound.SoundEvent;
import net.minecraft.util.Identifier;
import net.minecraft.util.Lazy;
import net.spell_engine.rpg_series.config.ArmorSetConfig;
import net.spell_engine.rpg_series.config.AttributeModifier;
import net.spell_engine.rpg_series.item.Armor;
import net.spell_engine.rpg_series.item.Equipment;

import java.util.ArrayList;
import java.util.EnumMap;
import java.util.Map;
import java.util.function.Supplier;

public class ArcherArmors {
    public static final class CompatArmorMaterial implements ArmorMaterial {
        private static final EnumMap<ArmorItem.Type, Integer> BASE_DURABILITY = new EnumMap<>(Map.of(
                ArmorItem.Type.BOOTS, 13,
                ArmorItem.Type.LEGGINGS, 15,
                ArmorItem.Type.CHESTPLATE, 16,
                ArmorItem.Type.HELMET, 11));

        private final String name;
        private final int durabilityMultiplier;
        private final EnumMap<ArmorItem.Type, Integer> protection;
        private final int enchantability;
        private final SoundEvent equipSound;
        private final Lazy<Ingredient> repairIngredient;

        public CompatArmorMaterial(String name, int durabilityMultiplier,
                                   int protectionHead, int protectionChest, int protectionLegs, int protectionFeet,
                                   int enchantability, SoundEvent equipSound, Supplier<Ingredient> repairIngredient) {
            this.name = ArchersMod.ID + ":" + name;
            this.durabilityMultiplier = durabilityMultiplier;
            this.protection = new EnumMap<>(Map.of(
                    ArmorItem.Type.HELMET, protectionHead,
                    ArmorItem.Type.CHESTPLATE, protectionChest,
                    ArmorItem.Type.LEGGINGS, protectionLegs,
                    ArmorItem.Type.BOOTS, protectionFeet));
            this.enchantability = enchantability;
            this.equipSound = equipSound;
            this.repairIngredient = new Lazy<>(repairIngredient);
        }

        @Override public int getDurability(ArmorItem.Type type) { return BASE_DURABILITY.get(type) * durabilityMultiplier; }
        @Override public int getProtection(ArmorItem.Type type) { return protection.get(type); }
        @Override public int getEnchantability() { return enchantability; }
        @Override public SoundEvent getEquipSound() { return equipSound; }
        @Override public Ingredient getRepairIngredient() { return repairIngredient.get(); }
        @Override public String getName() { return name; }
        @Override public float getToughness() { return 0; }
        @Override public float getKnockbackResistance() { return 0; }
    }

    public static RegistryEntry<ArmorMaterial> material(String name, int durabilityMultiplier,
                                                        int protectionHead, int protectionChest, int protectionLegs, int protectionFeet,
                                                        int enchantability, SoundEvent equipSound, Supplier<Ingredient> repairIngredient) {
        return RegistryEntry.of(new CompatArmorMaterial(name, durabilityMultiplier,
                protectionHead, protectionChest, protectionLegs, protectionFeet,
                enchantability, equipSound, repairIngredient));
    }

    public static RegistryEntry<ArmorMaterial> material_t1 = material(
            "archer_armor", 15,
            2, 3, 3, 2,
            9,
            ArcherSounds.ARCHER_ARMOR_EQUIP.soundEvent(), () -> Ingredient.ofItems(Items.LEATHER));

    public static RegistryEntry<ArmorMaterial> material_t2 = material(
            "ranger_armor", 25,
            2, 3, 3, 2,
            10,
            ArcherSounds.ARCHER_ARMOR_EQUIP.soundEvent(), () -> Ingredient.ofItems(Items.RABBIT_HIDE));

    public static RegistryEntry<ArmorMaterial> material_t3 = material(
            "netherite_ranger_armor", 35,
            2, 3, 3, 2,
            15,
            ArcherSounds.ARCHER_ARMOR_EQUIP.soundEvent(), () -> Ingredient.ofItems(Items.NETHERITE_INGOT));

    public static final ArrayList<Armor.Entry> entries = new ArrayList<>();
    private static Armor.Entry create(RegistryEntry<ArmorMaterial> material, Identifier id, int durability,
                                      Armor.Set.ItemFactory factory, ArmorSetConfig defaults, int tier) {
        var entry = Armor.Entry.create(material, id, durability, factory, defaults, Equipment.LootProperties.of(tier));
        entries.add(entry);
        return entry;
    }

    private static AttributeModifier damageMultiplier(float value) {
        return new AttributeModifier(EntityAttributes_RangedWeapon.DAMAGE.id.toString(), value,
                EntityAttributeModifier.Operation.MULTIPLY_BASE);
    }

    private static AttributeModifier hasteMultiplier(float value) {
        return new AttributeModifier(EntityAttributes_RangedWeapon.HASTE.id.toString(), value,
                EntityAttributeModifier.Operation.MULTIPLY_BASE);
    }

    public static final float damage_T1 = 0.05F;
    public static final float haste_T2 = 0.03F;
    public static final float damage_T2 = 0.08F;
    public static final float haste_T3 = 0.04F;
    public static final float damage_T3 = 0.09F;

    public static final Armor.Set archerArmorSet_T1 = create(
            material_t1, new Identifier(ArchersMod.ID, "archer_armor"), 15, ArcherArmor::archer,
            ArmorSetConfig.with(
                    new ArmorSetConfig.Piece(2).add(damageMultiplier(damage_T1)),
                    new ArmorSetConfig.Piece(3).add(damageMultiplier(damage_T1)),
                    new ArmorSetConfig.Piece(3).add(damageMultiplier(damage_T1)),
                    new ArmorSetConfig.Piece(2).add(damageMultiplier(damage_T1))), 1)
            .translatedName("Archer Hood", "Archer Tunic", "Archer Leggings", "Archer Boots").armorSet();

    public static final Armor.Set archerArmorSet_T2 = create(
            material_t2, new Identifier(ArchersMod.ID, "ranger_armor"), 25, ArcherArmor::ranger,
            ArmorSetConfig.with(
                    new ArmorSetConfig.Piece(2).add(damageMultiplier(damage_T2)).add(hasteMultiplier(haste_T2)),
                    new ArmorSetConfig.Piece(3).add(damageMultiplier(damage_T2)).add(hasteMultiplier(haste_T2)),
                    new ArmorSetConfig.Piece(3).add(damageMultiplier(damage_T2)).add(hasteMultiplier(haste_T2)),
                    new ArmorSetConfig.Piece(2).add(damageMultiplier(damage_T2)).add(hasteMultiplier(haste_T2))), 2)
            .translatedName("Ranger Hood", "Ranger Tunic", "Ranger Leggings", "Ranger Boots").armorSet();

    public static final Armor.Set archerArmorSet_T3 = create(
            material_t3, new Identifier(ArchersMod.ID, "netherite_ranger_armor"), 35, ArcherArmor::ranger,
            ArmorSetConfig.with(
                    new ArmorSetConfig.Piece(2).add(damageMultiplier(damage_T3)).add(hasteMultiplier(haste_T3)),
                    new ArmorSetConfig.Piece(3).add(damageMultiplier(damage_T3)).add(hasteMultiplier(haste_T3)),
                    new ArmorSetConfig.Piece(3).add(damageMultiplier(damage_T3)).add(hasteMultiplier(haste_T3)),
                    new ArmorSetConfig.Piece(2).add(damageMultiplier(damage_T3)).add(hasteMultiplier(haste_T3))), 3)
            .translatedName("Netherite Ranger Hood", "Netherite Ranger Tunic", "Netherite Ranger Leggings", "Netherite Ranger Boots").armorSet();

    public static void register(Map<String, ArmorSetConfig> configs) {
        Armor.register(configs, entries, Group.KEY);
    }
}
''', encoding="utf-8")

for obsolete in [
    java_root / "net/archers/component/ArcherComponents.java",
    java_root / "net/archers/mixin/component/DataComponentTypesMixin.java",
]:
    if obsolete.exists():
        obsolete.unlink()

mixins = resources_root / "archers.mixins.json"
if not mixins.is_file():
    raise SystemExit(f"current archers.mixins.json missing: {mixins}")
data = json.loads(mixins.read_text(encoding="utf-8"))
data["mixins"] = [m for m in data.get("mixins", []) if m != "component.DataComponentTypesMixin"]
mixins.write_text(json.dumps(data, indent=2) + "\n", encoding="utf-8")

print(f"[Archers transforms] Identifier.of rewrites: {identifier_rewrites}")
print("[Archers transforms] Quivers: 1.21 components -> Bundle API 1.20.1 capacity + tooltip override")
print("[Archers transforms] Auto Fire: 1.21 boolean component -> persistent 1.20.1 ItemStack NBT; bootstrap mixin removed")
print("[Archers transforms] Sounds: current 3.1.1 entries -> native 1.20.1 SoundEvent registry values")
print("[Archers transforms] Armor: current 3.1.1 materials -> target-native 1.20.1 ArmorMaterial + RegistryEntry.of")
