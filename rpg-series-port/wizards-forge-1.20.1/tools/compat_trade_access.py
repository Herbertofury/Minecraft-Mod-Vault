#!/usr/bin/env python3
from pathlib import Path
import re
import sys

if len(sys.argv) != 2:
    raise SystemExit('usage: compat_trade_access.py <generated-wizards-root>')

root = Path(sys.argv[1]).resolve()
rel = 'common/src/main/java/net/wizards/villager/WizardVillagers.java'
path = root / rel
if not path.is_file():
    raise SystemExit(f'Wizards trade compatibility source missing: {rel}')


def add_import(import_line: str) -> None:
    text = path.read_text()
    if import_line in text:
        return
    match = re.search(r'^package [^;]+;\n', text)
    if not match:
        raise SystemExit('Wizards trade compatibility could not locate package declaration')
    path.write_text(text[:match.end()] + '\n' + import_line + '\n' + text[match.end():])


def replace_exact(old: str, new: str, label: str) -> None:
    text = path.read_text()
    count = text.count(old)
    if count != 1:
        raise SystemExit(f'Wizards trade compatibility stale transform for {label}: expected 1 occurrence, found {count}')
    path.write_text(text.replace(old, new, 1))


add_import('import net.minecraft.enchantment.EnchantmentHelper;')
add_import('import net.minecraft.item.Item;')

# Forge 1.20.1 remaps Yarn TradeOffers.SellItemFactory to Mojmap
# VillagerTrades.ItemsForEmeralds, a package-private implementation class. The
# same is true for EnchantedItemForEmeralds. Referencing those concrete vanilla
# classes from Wizards therefore compiles in named mappings but throws
# IllegalAccessError in a real Forge packaged runtime. Re-express the exact
# vanilla factories through the public TradeOffers.Factory/TradeOffer contract.
ordinary = {
    'new TradeOffers.SellItemFactory(new ItemStack(RuneItems.get(RuneItems.RuneType.ARCANE)), 2, 8, 128, 3, 0.1f)':
        'sell(RuneItems.get(RuneItems.RuneType.ARCANE), 2, 8, 128, 3, 0.1F)',
    'new TradeOffers.SellItemFactory(new ItemStack(RuneItems.get(RuneItems.RuneType.FIRE)), 2, 8, 128, 3, 0.1f)':
        'sell(RuneItems.get(RuneItems.RuneType.FIRE), 2, 8, 128, 3, 0.1F)',
    'new TradeOffers.SellItemFactory(new ItemStack(RuneItems.get(RuneItems.RuneType.FROST)), 2, 8, 128, 3, 0.1f)':
        'sell(RuneItems.get(RuneItems.RuneType.FROST), 2, 8, 128, 3, 0.1F)',
    'new TradeOffers.SellItemFactory(WizardWeapons.wizardStaff.item(), 4, 1, 12, 18)':
        'sell(WizardWeapons.wizardStaff.item(), 4, 1, 12, 18, 0.05F)',
    'new TradeOffers.SellItemFactory(WizardWeapons.noviceWand.item(), 4, 1, 12, 18)':
        'sell(WizardWeapons.noviceWand.item(), 4, 1, 12, 18, 0.05F)',
    'new TradeOffers.SellItemFactory(WizardWeapons.arcaneWand.item(), 18, 1, 12, 18)':
        'sell(WizardWeapons.arcaneWand.item(), 18, 1, 12, 18, 0.05F)',
    'new TradeOffers.SellItemFactory(WizardWeapons.fireWand.item(), 18, 1, 12, 18)':
        'sell(WizardWeapons.fireWand.item(), 18, 1, 12, 18, 0.05F)',
    'new TradeOffers.SellItemFactory(WizardWeapons.frostWand.item(), 18, 1, 12, 18)':
        'sell(WizardWeapons.frostWand.item(), 18, 1, 12, 18, 0.05F)',
    'new TradeOffers.SellItemFactory(new ItemStack(WizardArmors.wizardRobeSet.head), 15, 1, 12, 16, 0.1F)':
        'sell(WizardArmors.wizardRobeSet.head, 15, 1, 12, 16, 0.1F)',
    'new TradeOffers.SellItemFactory(new ItemStack(WizardArmors.wizardRobeSet.feet), 15, 1, 12, 16, 0.1F)':
        'sell(WizardArmors.wizardRobeSet.feet, 15, 1, 12, 16, 0.1F)',
    'new TradeOffers.SellItemFactory(new ItemStack(WizardArmors.wizardRobeSet.chest), 20, 1, 12, 16, 0.1F)':
        'sell(WizardArmors.wizardRobeSet.chest, 20, 1, 12, 16, 0.1F)',
    'new TradeOffers.SellItemFactory(new ItemStack(WizardArmors.wizardRobeSet.legs), 20, 1, 12, 16, 0.1F)':
        'sell(WizardArmors.wizardRobeSet.legs, 20, 1, 12, 16, 0.1F)',
}
for old, new in ordinary.items():
    replace_exact(old, new, old)

for school in ('arcane', 'fire', 'frost'):
    old = f'''(entity, random) -> new TradeOffers.SellEnchantedToolFactory(\n                        WizardWeapons.{school}Staff.item(), 40, 3, 30, 0F).create(entity, random)'''
    new = f'sellEnchanted(WizardWeapons.{school}Staff.item(), 40, 3, 30, 0F)'
    replace_exact(old, new, f'{school} enchanted staff trade')

helpers = '''
    private static TradeOffers.Factory sell(Item item, int emeraldCost, int count,
                                             int maxUses, int villagerXp, float priceMultiplier) {
        return (entity, random) -> new TradeOffer(
                new ItemStack(Items.EMERALD, emeraldCost),
                new ItemStack(item, count),
                maxUses, villagerXp, priceMultiplier);
    }

    private static TradeOffers.Factory sellEnchanted(Item item, int baseEmeraldCost,
                                                      int maxUses, int villagerXp, float priceMultiplier) {
        return (entity, random) -> {
            int level = 5 + random.nextInt(15);
            ItemStack enchanted = EnchantmentHelper.enchant(random, new ItemStack(item), level, false);
            int emeraldCost = Math.min(baseEmeraldCost + level, 64);
            return new TradeOffer(
                    new ItemStack(Items.EMERALD, emeraldCost),
                    enchanted,
                    maxUses, villagerXp, priceMultiplier);
        };
    }
'''
text = path.read_text().rstrip()
if helpers.strip() not in text:
    if not text.endswith('}'):
        raise SystemExit('Wizards trade compatibility could not locate outer class close')
    text = text[:-1].rstrip() + '\n\n' + helpers.strip('\n') + '\n}\n'
    path.write_text(text)

final = path.read_text()
for forbidden in ('TradeOffers.SellItemFactory', 'TradeOffers.SellEnchantedToolFactory'):
    if forbidden in final:
        raise SystemExit(f'Wizards Forge runtime-inaccessible vanilla trade implementation survived: {forbidden}')
if final.count('sell(') != 13 or final.count('sellEnchanted(') != 4:
    raise SystemExit('Wizards trade compatibility helper/use inventory drifted')

print('Wizards villager trade compatibility applied: package-private vanilla trade factories replaced with public TradeOffer semantics')
