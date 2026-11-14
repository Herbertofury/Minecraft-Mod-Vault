#!/usr/bin/env python3
from __future__ import annotations
import pathlib, sys

root = pathlib.Path(sys.argv[1]).resolve()
if not root.is_dir():
    raise SystemExit(f"missing generated Rogues Java root: {root}")

def replace_exact(rel: str, old: str, new: str, label: str) -> None:
    path = root / rel
    text = path.read_text(encoding="utf-8")
    count = text.count(old)
    if count != 1:
        raise SystemExit(f"[{label}] expected one pinned shape in {rel}, found {count}")
    path.write_text(text.replace(old, new, 1), encoding="utf-8")
    print(f"[Rogues 1.20.1 compat batch2] {label}: {rel}")

# Current Bear Trap stays intact; only adapt the target model render signature.
replace_exact(
    "net/rogues/client/entity/BearTrapEntityRenderer.java",
    "        model.render(matrices, vertices, light, OverlayTexture.DEFAULT_UV, -1);",
    "        model.render(matrices, vertices, light, OverlayTexture.DEFAULT_UV, 1F, 1F, 1F, 1F);",
    "Bear Trap model RGBA render signature",
)

# In 1.20.1 LivingEntity status APIs accept raw StatusEffect values, while current Spell Engine
# Entries expose both the current holder and the authoritative raw effect. Unwrap only at vanilla calls.
raw_effect_replacements = {
    "net/rogues/effect/RogueEffects.java": (
        ("attacker.hasStatusEffect(STEALTH.entry)", "attacker.hasStatusEffect(STEALTH.effect)"),
        ("attacker.removeStatusEffect(STEALTH.entry)", "attacker.removeStatusEffect(STEALTH.effect)"),
        ("caster.hasStatusEffect(STEALTH.entry)", "caster.hasStatusEffect(STEALTH.effect)"),
        ("caster.removeStatusEffect(STEALTH.entry)", "caster.removeStatusEffect(STEALTH.effect)"),
        ("user.hasStatusEffect(STEALTH.entry)", "user.hasStatusEffect(STEALTH.effect)"),
        ("user.removeStatusEffect(STEALTH.entry)", "user.removeStatusEffect(STEALTH.effect)"),
        ("context.entity().hasStatusEffect(STEALTH_SPEED.entry)", "context.entity().hasStatusEffect(STEALTH_SPEED.effect)"),
        ("context.entity().removeStatusEffect(STEALTH_SPEED.entry)", "context.entity().removeStatusEffect(STEALTH_SPEED.effect)"),
    ),
    "net/rogues/mixin/LivingEntityStealth.java": (
        ("instance.hasStatusEffect(RogueEffects.STEALTH.entry)", "instance.hasStatusEffect(RogueEffects.STEALTH.effect)"),
        ("thisEntity.hasStatusEffect(RogueEffects.STEALTH.entry)", "thisEntity.hasStatusEffect(RogueEffects.STEALTH.effect)"),
    ),
    "net/rogues/mixin/TrackTargetGoalStealth.java": (
        ("target.hasStatusEffect(RogueEffects.STEALTH.entry)", "target.hasStatusEffect(RogueEffects.STEALTH.effect)"),
        ("target.hasStatusEffect(RogueEffects.SHADOW_STEP.entry)", "target.hasStatusEffect(RogueEffects.SHADOW_STEP.effect)"),
    ),
}
for rel, replacements in raw_effect_replacements.items():
    path = root / rel
    text = path.read_text(encoding="utf-8")
    for old, new in replacements:
        if text.count(old) != 1:
            raise SystemExit(f"[raw status effect] expected one pinned call in {rel}: {old}")
        text = text.replace(old, new, 1)
    path.write_text(text, encoding="utf-8")
    print(f"[Rogues 1.20.1 compat batch2] raw target status-effect API: {rel}")

# Later Mojang mappings expose convenience trade factories that are private/absent in Yarn 1.20.1.
# Preserve every current Rogues trade item/count/price/use/xp value using public target TradeOffer lambdas.
replace_exact(
    "net/rogues/village/RogueVillagers.java",
    "import net.minecraft.item.Items;",
    "import net.minecraft.enchantment.EnchantmentHelper;\nimport net.minecraft.item.Item;\nimport net.minecraft.item.ItemStack;\nimport net.minecraft.item.Items;",
    "trade helper imports",
)
replace_exact(
    "net/rogues/village/RogueVillagers.java",
    "import net.minecraft.village.TradeOffers;",
    "import net.minecraft.village.TradeOffer;\nimport net.minecraft.village.TradeOffers;",
    "TradeOffer import",
)
replace_exact(
    "net/rogues/village/RogueVillagers.java",
    "    public static void registerVillagers() {",
    '''    private static TradeOffers.Factory sell(Item item, int emeraldPrice, int count, int maxUses, int experience) {\n        return sell(item, emeraldPrice, count, maxUses, experience, 0.05F);\n    }\n\n    private static TradeOffers.Factory sell(Item item, int emeraldPrice, int count, int maxUses, int experience, float multiplier) {\n        return (entity, random) -> new TradeOffer(\n                new ItemStack(Items.EMERALD, emeraldPrice), new ItemStack(item, count),\n                maxUses, experience, multiplier);\n    }\n\n    private static TradeOffers.Factory buy(Item item, int count, int maxUses, int experience, int emeraldPrice) {\n        return (entity, random) -> new TradeOffer(\n                new ItemStack(item, count), new ItemStack(Items.EMERALD, emeraldPrice),\n                maxUses, experience, 0.05F);\n    }\n\n    private static TradeOffers.Factory sellEnchanted(Item item, int basePrice, int maxUses, int experience, float multiplier) {\n        return (entity, random) -> {\n            int enchantmentPower = 5 + random.nextInt(15);\n            ItemStack enchanted = EnchantmentHelper.enchant(random, new ItemStack(item), enchantmentPower, false);\n            int price = Math.min(basePrice + enchantmentPower, 64);\n            return new TradeOffer(new ItemStack(Items.EMERALD, price), enchanted, maxUses, experience, multiplier);\n        };\n    }\n\n    public static void registerVillagers() {''',
    "public target trade helper seam",
)

villagers = root / "net/rogues/village/RogueVillagers.java"
text = villagers.read_text(encoding="utf-8")
replacements = {
    "new TradeOffers.BuyItemFactory(Items.LEATHER, 8, 12, 4, 5)": "buy(Items.LEATHER, 8, 12, 4, 5)",
    "new TradeOffers.SellItemFactory(RogueWeapons.flint_dagger.item(), 6, 1, 12, 3)": "sell(RogueWeapons.flint_dagger.item(), 6, 1, 12, 3)",
    "new TradeOffers.SellItemFactory(RogueWeapons.stone_double_axe.item(), 8, 1, 12, 4)": "sell(RogueWeapons.stone_double_axe.item(), 8, 1, 12, 4)",
    "new TradeOffers.BuyItemFactory(Items.IRON_INGOT, 12, 12, 5, 8)": "buy(Items.IRON_INGOT, 12, 12, 5, 8)",
    "new TradeOffers.SellItemFactory(RogueWeapons.iron_sickle.item(), 12, 1, 12, 10)": "sell(RogueWeapons.iron_sickle.item(), 12, 1, 12, 10)",
    "new TradeOffers.SellItemFactory(RogueWeapons.iron_glaive.item(), 18, 1, 12, 10)": "sell(RogueWeapons.iron_glaive.item(), 18, 1, 12, 10)",
    "new TradeOffers.SellItemFactory(RogueArmors.RogueArmorSet_t1.head, 15, 1, 12, 13)": "sell(RogueArmors.RogueArmorSet_t1.head, 15, 1, 12, 13)",
    "new TradeOffers.SellItemFactory(RogueArmors.WarriorArmorSet_t1.head, 15, 1, 12, 13)": "sell(RogueArmors.WarriorArmorSet_t1.head, 15, 1, 12, 13)",
    "new TradeOffers.SellItemFactory(RogueWeapons.iron_dagger.item(), 14, 1, 12, 15)": "sell(RogueWeapons.iron_dagger.item(), 14, 1, 12, 15)",
    "new TradeOffers.SellItemFactory(RogueWeapons.iron_double_axe.item(), 18, 1, 12, 15)": "sell(RogueWeapons.iron_double_axe.item(), 18, 1, 12, 15)",
    "new TradeOffers.SellItemFactory(RogueArmors.RogueArmorSet_t1.feet, 15, 1, 12, 15)": "sell(RogueArmors.RogueArmorSet_t1.feet, 15, 1, 12, 15)",
    "new TradeOffers.SellItemFactory(RogueArmors.WarriorArmorSet_t1.feet, 15, 1, 12, 15)": "sell(RogueArmors.WarriorArmorSet_t1.feet, 15, 1, 12, 15)",
    "new TradeOffers.SellItemFactory(RogueArmors.RogueArmorSet_t1.legs, 15, 1, 12, 15)": "sell(RogueArmors.RogueArmorSet_t1.legs, 15, 1, 12, 15)",
    "new TradeOffers.SellItemFactory(RogueArmors.WarriorArmorSet_t1.legs, 15, 1, 12, 15)": "sell(RogueArmors.WarriorArmorSet_t1.legs, 15, 1, 12, 15)",
    "new TradeOffers.SellItemFactory(RogueArmors.RogueArmorSet_t1.chest, 15, 1, 12, 15)": "sell(RogueArmors.RogueArmorSet_t1.chest, 15, 1, 12, 15)",
    "new TradeOffers.SellItemFactory(RogueArmors.WarriorArmorSet_t1.chest, 15, 1, 12, 15)": "sell(RogueArmors.WarriorArmorSet_t1.chest, 15, 1, 12, 15)",
    "new TradeOffers.SellItemFactory(Items.GOAT_HORN, 15, 1, 12, 5)": "sell(Items.GOAT_HORN, 15, 1, 12, 5)",
}
for old, new in replacements.items():
    if text.count(old) != 1:
        raise SystemExit(f"[trade adapter] expected one current call, found {text.count(old)}: {old}")
    text = text.replace(old, new, 1)
for weapon, price in (("diamond_dagger", 30), ("diamond_sickle", 30), ("diamond_double_axe", 40), ("diamond_glaive", 40)):
    old = f'''(entity, random) -> new TradeOffers.SellEnchantedToolFactory(\n                        RogueWeapons.{weapon}.item(), {price}, 3, 30, 0F).create(entity, random)'''
    new = f"sellEnchanted(RogueWeapons.{weapon}.item(), {price}, 3, 30, 0F)"
    if text.count(old) != 1:
        raise SystemExit(f"[enchanted trade] expected one current {weapon} call, found {text.count(old)}")
    text = text.replace(old, new, 1)
villagers.write_text(text, encoding="utf-8")
print("[Rogues 1.20.1 compat batch2] current arms-merchant contents expressed through public target factories")

forbidden = (
    "model.render(matrices, vertices, light, OverlayTexture.DEFAULT_UV, -1)",
    ".hasStatusEffect(STEALTH.entry)", ".removeStatusEffect(STEALTH.entry)",
    "RogueEffects.STEALTH.entry)", "RogueEffects.SHADOW_STEP.entry)",
    "TradeOffers.SellItemFactory(", "TradeOffers.BuyItemFactory(", "TradeOffers.SellEnchantedToolFactory(",
)
survivors = []
for path in sorted(root.rglob("*.java")):
    for line_no, line in enumerate(path.read_text(encoding="utf-8").splitlines(), 1):
        code = line.split("//", 1)[0]
        if any(token in code for token in forbidden):
            survivors.append(f"{path.relative_to(root)}:{line_no}:{line.strip()}")
if survivors:
    raise SystemExit("batch2-owned API survived compatibility pass:\n" + "\n".join(survivors))
print("[Rogues 1.20.1 compat batch2] renderer/status/trade frontier translated fail-closed")
