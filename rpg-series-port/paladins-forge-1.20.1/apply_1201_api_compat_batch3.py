#!/usr/bin/env python3
from __future__ import annotations

import pathlib
import sys

root = pathlib.Path(sys.argv[1]).resolve()
if not root.is_dir():
    raise SystemExit(f"missing generated Paladins Java root: {root}")


def replace_exact(rel: str, old: str, new: str, label: str) -> None:
    path = root / rel
    text = path.read_text(encoding="utf-8")
    count = text.count(old)
    if count != 1:
        raise SystemExit(f"[{label}] expected exactly one pinned source shape in {rel}, found {count}")
    path.write_text(text.replace(old, new, 1), encoding="utf-8")
    print(f"[Paladins 1.20.1 API batch3] {label}: {rel}")


# Spell Engine 1.20.1 deliberately exposes holder-backed shield metadata while the graduated
# Shield API consumes raw values. Unwrap at the Paladins integration seam; do not modify either
# foundation and do not shade either one into the Paladins release.
replace_exact(
    "net/paladins/item/PaladinShields.java",
    "import net.minecraft.util.Identifier;",
    "import net.minecraft.util.Identifier;\nimport net.minecraft.util.Pair;",
    "Shield adapter import target Pair",
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
    "Spell Engine holder -> graduated Shield raw-value adapter",
)

# Minecraft 1.20.1 has no ARMOR_MATERIAL registry and ArmorMaterial is still an interface.
# Preserve the current 3.1.1 protection/enchantability/sound/repair/toughness/KB semantics, plus
# the current per-set durability multipliers, and supply Spell Engine a direct registry entry.
replace_exact(
    "net/paladins/item/armor/Armors.java",
    "import net.minecraft.registry.Registries;\nimport net.minecraft.registry.Registry;",
    "import net.minecraft.registry.Registries;\nimport net.minecraft.registry.Registry;",
    "Armor imports remain stable after registration bridge",
)
replace_exact(
    "net/paladins/item/armor/Armors.java",
    '''    public static RegistryEntry<ArmorMaterial> material(
            String name, int protectionHead, int protectionChest, int protectionLegs, int protectionFeet,
            int enchantability, RegistryEntry<SoundEvent> equipSound, Supplier<Ingredient> repairIngredient) {

        var material = new ArmorMaterial(
                Map.of(
                        ArmorItem.Type.HELMET, protectionHead,
                        ArmorItem.Type.CHESTPLATE, protectionChest,
                        ArmorItem.Type.LEGGINGS, protectionLegs,
                        ArmorItem.Type.BOOTS, protectionFeet),
                enchantability, equipSound, repairIngredient,
                List.of(new ArmorMaterial.Layer(new Identifier(PaladinsMod.ID, name))),
                0,0
        );
        return net.spell_engine.compat.registry.RegistrationBridge.registerReference(Registries.ARMOR_MATERIAL, new Identifier(PaladinsMod.ID, name), material);
    }
''',
    '''    private static final Map<ArmorItem.Type, Integer> BASE_DURABILITY = Map.of(
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
            int enchantability, RegistryEntry<SoundEvent> equipSound, Supplier<Ingredient> repairIngredient) {
        ArmorMaterial material = new PaladinArmorMaterial(
                name, durabilityMultiplier,
                protectionHead, protectionChest, protectionLegs, protectionFeet,
                enchantability, equipSound.value(), repairIngredient);
        return new RegistryEntry.Direct<>(material);
    }
''',
    "1.20.1 direct ArmorMaterial implementation",
)

# Add the current create(...) durability multiplier to the six material declarations. One material
# backs one current armor set, so this is a lossless translation of the later split material/item model.
armor = root / "net/paladins/item/armor/Armors.java"
text = armor.read_text(encoding="utf-8")
material_durabilities = {
    '            "paladin_armor",\n            2, 6, 5, 2,': '            "paladin_armor", 15,\n            2, 6, 5, 2,',
    '            "crusader_armor",\n            3, 8, 6, 3,': '            "crusader_armor", 25,\n            3, 8, 6, 3,',
    '            "netherite_crusader_armor",\n            3, 8, 6, 3,': '            "netherite_crusader_armor", 37,\n            3, 8, 6, 3,',
    '            "priest_robe",\n            1, 3, 2, 1,': '            "priest_robe", 10,\n            1, 3, 2, 1,',
    '            "prior_robe",\n            1, 3, 2, 1,': '            "prior_robe", 20,\n            1, 3, 2, 1,',
    '            "netherite_prior_robe",\n            1, 3, 2, 1,': '            "netherite_prior_robe", 30,\n            1, 3, 2, 1,',
}
for old, new in material_durabilities.items():
    count = text.count(old)
    if count != 1:
        raise SystemExit(f"[armor durability] expected one current material declaration for {old.split(chr(10))[0]}, found {count}")
    text = text.replace(old, new, 1)
text = text.replace('Identifier.ofVanilla("generic.attack_damage")', 'new Identifier("minecraft", "generic.attack_damage")')
text = text.replace('Identifier.ofVanilla("generic.armor_toughness")', 'new Identifier("minecraft", "generic.armor_toughness")')
armor.write_text(text, encoding="utf-8")
print("[Paladins 1.20.1 API batch3] six current armor durability multipliers preserved; ofVanilla translated")

# Yarn 1.20.1 keeps the vanilla convenience trade implementations private. Keep current 3.1.1
# quantities/prices/use caps/xp/multipliers through public TradeOffers.Factory lambdas instead.
replace_exact(
    "net/paladins/village/PaladinVillagers.java",
    "import net.minecraft.item.Items;",
    "import net.minecraft.enchantment.EnchantmentHelper;\nimport net.minecraft.item.Item;\nimport net.minecraft.item.ItemStack;\nimport net.minecraft.item.Items;",
    "Trade adapter imports",
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
                new ItemStack(Items.EMERALD, emeraldPrice),
                new ItemStack(item, count),
                maxUses, experience, multiplier);
    }

    private static TradeOffers.Factory buy(Item item, int count, int maxUses, int experience, int emeraldPrice) {
        return (entity, random) -> new TradeOffer(
                new ItemStack(item, count),
                new ItemStack(Items.EMERALD, emeraldPrice),
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
    "Public 1.20.1 trade factory adapters",
)

villagers = root / "net/paladins/village/PaladinVillagers.java"
text = villagers.read_text(encoding="utf-8")
trade_replacements = {
    'new TradeOffers.SellItemFactory(RuneItems.get(RuneItems.RuneType.HEALING), 2, 8, 128, 1, 0.01f)':
        'sell(RuneItems.get(RuneItems.RuneType.HEALING), 2, 8, 128, 1, 0.01f)',
    'new TradeOffers.SellItemFactory(PaladinWeapons.acolyte_wand.item(), 4, 1, 12, 5)':
        'sell(PaladinWeapons.acolyte_wand.item(), 4, 1, 12, 5)',
    'new TradeOffers.SellItemFactory(PaladinWeapons.wooden_great_hammer.item(), 8, 1, 12, 8)':
        'sell(PaladinWeapons.wooden_great_hammer.item(), 8, 1, 12, 8)',
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
for old, new in trade_replacements.items():
    count = text.count(old)
    if count != 1:
        raise SystemExit(f"[trade adapter] expected one pinned factory call, found {count}: {old}")
    text = text.replace(old, new, 1)
for weapon in ("diamond_holy_staff", "diamond_claymore", "diamond_great_hammer"):
    old = f'''(entity, random) -> new TradeOffers.SellEnchantedToolFactory(
                        PaladinWeapons.{weapon}.item(), 40, 3, 30, 0F).create(entity, random)'''
    new = f'''sellEnchanted(PaladinWeapons.{weapon}.item(), 40, 3, 30, 0F)'''
    count = text.count(old)
    if count != 1:
        raise SystemExit(f"[enchanted trade adapter] expected one {weapon} factory call, found {count}")
    text = text.replace(old, new, 1)
villagers.write_text(text, encoding="utf-8")
print("[Paladins 1.20.1 API batch3] current merchant quantities/prices/use caps/xp preserved with public factories")

# Fail closed on the entire boundary this pass owns.
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
        if any(token in line for token in forbidden):
            survivors.append(f"{path.relative_to(root)}:{line_no}:{line.strip()}")
if survivors:
    raise SystemExit("batch3-owned 1.21/private API survived compatibility pass:\n" + "\n".join(survivors))

print("[Paladins 1.20.1 API batch3] shield/armor/trade compatibility adapters translated fail-closed")
