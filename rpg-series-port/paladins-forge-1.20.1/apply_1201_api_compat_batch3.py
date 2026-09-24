#!/usr/bin/env python3
from __future__ import annotations

import pathlib
import re
import sys

root = pathlib.Path(sys.argv[1]).resolve()
if not root.is_dir():
    raise SystemExit(f"missing generated Paladins Java root: {root}")


def replace_exact(rel: str, old: str, new: str, label: str) -> None:
    path = root / rel
    text = path.read_text(encoding="utf-8")
    count = text.count(old)
    if count != 1:
        raise SystemExit(f"[{label}] expected one pinned shape in {rel}, found {count}")
    path.write_text(text.replace(old, new, 1), encoding="utf-8")
    print(f"[Paladins 1.20.1 API batch3] {label}: {rel}")


# ---- Shield API integration seam ---------------------------------------------------------------
# Target Spell Engine exposes holder-backed metadata; the separately graduated Shield API consumes
# raw SoundEvent/EntityAttribute values. Unwrap only at Paladins' factory seam.
replace_exact(
    "net/paladins/item/PaladinShields.java",
    "import net.minecraft.util.Identifier;",
    "import net.minecraft.util.Identifier;\nimport net.minecraft.util.Pair;",
    "shield Pair import",
)
replace_exact(
    "net/paladins/item/PaladinShields.java",
    "        Shield.register(configs, entries, Group.KEY, CustomShieldItem::new);",
    '''        Shield.register(configs, entries, Group.KEY, (equipSound, repairIngredient, modifiers, settings) ->
                new CustomShieldItem(
                        equipSound.value(),
                        repairIngredient,
                        modifiers.stream()
                                .map(pair -> new Pair<>(pair.getLeft().value(), pair.getRight()))
                                .toList(),
                        settings));''',
    "Spell Engine holder metadata -> graduated Shield raw values",
)

# ---- Armor material bridge --------------------------------------------------------------------
# 1.20.1 ArmorMaterial is an interface and there is no ARMOR_MATERIAL registry. Spell Engine 1.20.1
# still accepts RegistryEntry<ArmorMaterial>, so preserve current material semantics in Direct entries.
armors = root / "net/paladins/item/armor/Armors.java"
text = armors.read_text(encoding="utf-8")
method_pattern = re.compile(
    r'''    public static RegistryEntry<ArmorMaterial> material\(\n'''
    r'''            String name, int protectionHead, int protectionChest, int protectionLegs, int protectionFeet,\n'''
    r'''            int enchantability, RegistryEntry<SoundEvent> equipSound, Supplier<Ingredient> repairIngredient\) \{\n\n'''
    r'''        var material = new ArmorMaterial\(\n.*?'''
    r'''        return net\.spell_engine\.compat\.registry\.RegistrationBridge\.registerReference\(Registries\.ARMOR_MATERIAL, new Identifier\(PaladinsMod\.ID, name\), material\);\n'''
    r'''    \}\n''',
    re.DOTALL,
)
replacement = '''    private static final Map<ArmorItem.Type, Integer> BASE_DURABILITY = Map.of(
            ArmorItem.Type.HELMET, 11,
            ArmorItem.Type.CHESTPLATE, 16,
            ArmorItem.Type.LEGGINGS, 15,
            ArmorItem.Type.BOOTS, 13);

    private static final class PaladinArmorMaterial implements ArmorMaterial {
        private final String name;
        private final int durabilityMultiplier;
        private final Map<ArmorItem.Type, Integer> protection;
        private final int enchantability;
        private final SoundEvent equipSound;
        private final Supplier<Ingredient> repairIngredient;

        private PaladinArmorMaterial(String name, int durabilityMultiplier,
                                     int protectionHead, int protectionChest, int protectionLegs, int protectionFeet,
                                     int enchantability, SoundEvent equipSound, Supplier<Ingredient> repairIngredient) {
            this.name = PaladinsMod.ID + ":" + name;
            this.durabilityMultiplier = durabilityMultiplier;
            this.protection = Map.of(
                    ArmorItem.Type.HELMET, protectionHead,
                    ArmorItem.Type.CHESTPLATE, protectionChest,
                    ArmorItem.Type.LEGGINGS, protectionLegs,
                    ArmorItem.Type.BOOTS, protectionFeet);
            this.enchantability = enchantability;
            this.equipSound = equipSound;
            this.repairIngredient = repairIngredient;
        }

        @Override public int getDurability(ArmorItem.Type type) { return BASE_DURABILITY.get(type) * durabilityMultiplier; }
        @Override public int getProtection(ArmorItem.Type type) { return protection.get(type); }
        @Override public int getEnchantability() { return enchantability; }
        @Override public SoundEvent getEquipSound() { return equipSound; }
        @Override public Ingredient getRepairIngredient() { return repairIngredient.get(); }
        @Override public String getName() { return name; }
        @Override public float getToughness() { return 0F; }
        @Override public float getKnockbackResistance() { return 0F; }
    }

    public static RegistryEntry<ArmorMaterial> material(
            String name, int durabilityMultiplier,
            int protectionHead, int protectionChest, int protectionLegs, int protectionFeet,
            int enchantability, SoundEvent equipSound, Supplier<Ingredient> repairIngredient) {
        return new RegistryEntry.Direct<>(new PaladinArmorMaterial(
                name, durabilityMultiplier,
                protectionHead, protectionChest, protectionLegs, protectionFeet,
                enchantability, equipSound, repairIngredient));
    }
'''
text, count = method_pattern.subn(replacement, text, count=1)
if count != 1:
    raise SystemExit(f"[armor material bridge] expected one transformed current material method, found {count}")

# Current Armor.Entry durability multipliers are authoritative. Move each multiplier into the 1.20.1
# ArmorMaterial implementation so vanilla ArmorItem durability and Spell Engine's configured durability agree.
material_durability = {
    "paladin_armor": 15,
    "crusader_armor": 25,
    "netherite_crusader_armor": 37,
    "priest_robe": 10,
    "prior_robe": 20,
    "netherite_prior_robe": 30,
}
for name, durability in material_durability.items():
    old = f'            "{name}",\n'
    new = f'            "{name}", {durability},\n'
    if text.count(old) != 1:
        raise SystemExit(f"[armor durability] expected one material declaration for {name}, found {text.count(old)}")
    text = text.replace(old, new, 1)

# Avoid holder initialization coupling: the 1.20.1 interface wants raw SoundEvent and PaladinSounds already
# owns that exact raw value. There are exactly six armor-material sound arguments.
entry_sound_count = text.count("PaladinSounds.paladin_armor_equip.entry()") + text.count("PaladinSounds.priest_robe_equip.entry()")
if entry_sound_count != 6:
    raise SystemExit(f"[armor sound] expected six holder-backed armor sounds, found {entry_sound_count}")
text = text.replace("PaladinSounds.paladin_armor_equip.entry()", "PaladinSounds.paladin_armor_equip.soundEvent()")
text = text.replace("PaladinSounds.priest_robe_equip.entry()", "PaladinSounds.priest_robe_equip.soundEvent()")

for old, new in (
    ('Identifier.ofVanilla("generic.attack_damage")', 'new Identifier("minecraft", "generic.attack_damage")'),
    ('Identifier.ofVanilla("generic.armor_toughness")', 'new Identifier("minecraft", "generic.armor_toughness")'),
):
    if text.count(old) != 1:
        raise SystemExit(f"[ofVanilla] expected one {old}, found {text.count(old)}")
    text = text.replace(old, new, 1)
armors.write_text(text, encoding="utf-8")
print("[Paladins 1.20.1 API batch3] armor materials preserved as direct target-native interface entries")

# ---- Villager trade factories -----------------------------------------------------------------
# Historical Paladins 1.20.1 used public TradeOffer lambdas. Keep current 3.1.1 trade contents/numbers,
# but express them through public 1.20.1 constructors instead of later private/nested convenience types.
replace_exact(
    "net/paladins/village/PaladinVillagers.java",
    "import net.minecraft.item.Items;",
    "import net.minecraft.enchantment.EnchantmentHelper;\nimport net.minecraft.item.Item;\nimport net.minecraft.item.ItemStack;\nimport net.minecraft.item.Items;",
    "trade imports",
)
replace_exact(
    "net/paladins/village/PaladinVillagers.java",
    "import net.minecraft.village.TradeOffers;",
    "import net.minecraft.village.TradeOffer;\nimport net.minecraft.village.TradeOffers;",
    "TradeOffer import",
)
replace_exact(
    "net/paladins/village/PaladinVillagers.java",
    "    public static void registerVillagers() {",
    '''    private static TradeOffers.Factory sell(Item item, int emeraldPrice, int count, int maxUses, int experience) {
        return sell(item, emeraldPrice, count, maxUses, experience, 0.05F);
    }

    private static TradeOffers.Factory sell(Item item, int emeraldPrice, int count, int maxUses, int experience, float multiplier) {
        return (entity, random) -> new TradeOffer(
                new ItemStack(Items.EMERALD, emeraldPrice), new ItemStack(item, count),
                maxUses, experience, multiplier);
    }

    private static TradeOffers.Factory buy(Item item, int count, int maxUses, int experience, int emeraldPrice) {
        return (entity, random) -> new TradeOffer(
                new ItemStack(item, count), new ItemStack(Items.EMERALD, emeraldPrice),
                maxUses, experience, 0.05F);
    }

    private static TradeOffers.Factory sellEnchanted(Item item, int basePrice, int maxUses, int experience, float multiplier) {
        return (entity, random) -> {
            int enchantmentPower = 5 + random.nextInt(15);
            ItemStack enchanted = EnchantmentHelper.enchant(random, new ItemStack(item), enchantmentPower, false);
            int price = Math.min(basePrice + enchantmentPower, 64);
            return new TradeOffer(new ItemStack(Items.EMERALD, price), enchanted, maxUses, experience, multiplier);
        };
    }

    public static void registerVillagers() {''',
    "public 1.20.1 trade helper seam",
)

villagers = root / "net/paladins/village/PaladinVillagers.java"
text = villagers.read_text(encoding="utf-8")
replacements = {
    'new TradeOffers.SellItemFactory(RuneItems.get(RuneItems.RuneType.HEALING), 2, 8, 128, 1, 0.01f)': 'sell(RuneItems.get(RuneItems.RuneType.HEALING), 2, 8, 128, 1, 0.01f)',
    'new TradeOffers.SellItemFactory(PaladinWeapons.acolyte_wand.item(), 4, 1, 12, 5)': 'sell(PaladinWeapons.acolyte_wand.item(), 4, 1, 12, 5)',
    'new TradeOffers.SellItemFactory(PaladinWeapons.wooden_great_hammer.item(), 8, 1, 12, 8)': 'sell(PaladinWeapons.wooden_great_hammer.item(), 8, 1, 12, 8)',
    'new TradeOffers.BuyItemFactory(Items.WHITE_WOOL, 5, 12, 5, 8)': 'buy(Items.WHITE_WOOL, 5, 12, 5, 8)',
    'new TradeOffers.BuyItemFactory(Items.IRON_INGOT, 6, 12, 5, 8)': 'buy(Items.IRON_INGOT, 6, 12, 5, 8)',
    'new TradeOffers.BuyItemFactory(Items.CHAIN, 6, 12, 5, 8)': 'buy(Items.CHAIN, 6, 12, 5, 8)',
    'new TradeOffers.BuyItemFactory(Items.GOLD_INGOT, 6, 12, 5, 8)': 'buy(Items.GOLD_INGOT, 6, 12, 5, 8)',
    'new TradeOffers.SellItemFactory(Armors.paladinArmorSet_t1.head, 15, 1, 12, 13)': 'sell(Armors.paladinArmorSet_t1.head, 15, 1, 12, 13)',
    'new TradeOffers.SellItemFactory(Armors.paladinArmorSet_t1.feet, 15, 1, 12, 13)': 'sell(Armors.paladinArmorSet_t1.feet, 15, 1, 12, 13)',
    'new TradeOffers.SellItemFactory(Armors.priestArmorSet_t1.head, 15, 1, 12, 13)': 'sell(Armors.priestArmorSet_t1.head, 15, 1, 12, 13)',
    'new TradeOffers.SellItemFactory(Armors.priestArmorSet_t1.feet, 15, 1, 12, 13)': 'sell(Armors.priestArmorSet_t1.feet, 15, 1, 12, 13)',
    'new TradeOffers.SellItemFactory(Armors.paladinArmorSet_t1.chest, 20, 1, 12, 15)': 'sell(Armors.paladinArmorSet_t1.chest, 20, 1, 12, 15)',
    'new TradeOffers.SellItemFactory(Armors.paladinArmorSet_t1.legs, 20, 1, 12, 15)': 'sell(Armors.paladinArmorSet_t1.legs, 20, 1, 12, 15)',
    'new TradeOffers.SellItemFactory(Armors.priestArmorSet_t1.chest, 20, 1, 12, 15)': 'sell(Armors.priestArmorSet_t1.chest, 20, 1, 12, 15)',
    'new TradeOffers.SellItemFactory(Armors.priestArmorSet_t1.legs, 20, 1, 12, 15)': 'sell(Armors.priestArmorSet_t1.legs, 20, 1, 12, 15)',
}
for old, new in replacements.items():
    if text.count(old) != 1:
        raise SystemExit(f"[trade adapter] expected one current factory call, found {text.count(old)}: {old}")
    text = text.replace(old, new, 1)

for weapon in ("diamond_holy_staff", "diamond_claymore", "diamond_great_hammer"):
    old = f'''(entity, random) -> new TradeOffers.SellEnchantedToolFactory(
                        PaladinWeapons.{weapon}.item(), 40, 3, 30, 0F).create(entity, random)'''
    new = f'''sellEnchanted(PaladinWeapons.{weapon}.item(), 40, 3, 30, 0F)'''
    if text.count(old) != 1:
        raise SystemExit(f"[enchanted trade] expected one {weapon} current factory, found {text.count(old)}")
    text = text.replace(old, new, 1)
villagers.write_text(text, encoding="utf-8")
print("[Paladins 1.20.1 API batch3] current merchant contents expressed with public target factories")

# ---- Acceptance -------------------------------------------------------------------------------
forbidden = (
    "CustomShieldItem::new",
    "Registries.ARMOR_MATERIAL",
    "new ArmorMaterial(",
    "Identifier.ofVanilla(",
    "TradeOffers.SellItemFactory(",
    "TradeOffers.BuyItemFactory(",
    "TradeOffers.SellEnchantedToolFactory(",
)
survivors: list[str] = []
for path in sorted(root.rglob("*.java")):
    for line_no, line in enumerate(path.read_text(encoding="utf-8").splitlines(), 1):
        code = line.split("//", 1)[0]
        if any(token in code for token in forbidden):
            survivors.append(f"{path.relative_to(root)}:{line_no}:{line.strip()}")
if survivors:
    raise SystemExit("batch3-owned API survived compatibility pass:\n" + "\n".join(survivors))

print("[Paladins 1.20.1 API batch3] shield/armor/trade frontier translated fail-closed")
